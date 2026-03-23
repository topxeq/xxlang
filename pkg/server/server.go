// pkg/server/server.go
// HTTP/HTTPS server implementation for Xxlang.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// EndResponseMarker is returned by scripts to signal that the response is complete
const EndResponseMarker = "TX_END_RESPONSE_XT"

// Config holds server configuration
type Config struct {
	HTTPPort  int    `json:"httpPort"`
	HTTPSPort int    `json:"httpsPort"`
	WebPath   string `json:"webPath"`
	MSPath    string `json:"msPath"`
	CertPath  string `json:"certPath"`
}

// Server represents the HTTP/HTTPS server
type Server struct {
	Config  Config
	Mux     *http.ServeMux
	Verbose bool
}

// NewServer creates a new server instance
func NewServer(cfg Config) *Server {
	return &Server{
		Config: cfg,
		Mux:    http.NewServeMux(),
	}
}

// Start starts the HTTP and HTTPS servers
func (s *Server) Start() error {
	// Register HTTP built-in functions
	objects.RegisterHttpBuiltins()

	// Setup routes
	s.setupRoutes()

	errChan := make(chan error, 2)

	// Start HTTP server
	if s.Config.HTTPPort > 0 {
		go func() {
			addr := fmt.Sprintf(":%d", s.Config.HTTPPort)
			fmt.Printf("[HTTP] Starting server on port %d\n", s.Config.HTTPPort)
			errChan <- http.ListenAndServe(addr, s.Mux)
		}()
	}

	// Start HTTPS server
	if s.Config.HTTPSPort > 0 {
		go func() {
			certFile := filepath.Join(s.Config.CertPath, "server.crt")
			keyFile := filepath.Join(s.Config.CertPath, "server.key")

			if _, err := os.Stat(certFile); os.IsNotExist(err) {
				fmt.Printf("[HTTPS] Certificate file not found: %s\n", certFile)
				errChan <- fmt.Errorf("certificate not found: %s", certFile)
				return
			}

			addr := fmt.Sprintf(":%d", s.Config.HTTPSPort)
			fmt.Printf("[HTTPS] Starting server on port %d\n", s.Config.HTTPSPort)
			errChan <- http.ListenAndServeTLS(addr, certFile, keyFile, s.Mux)
		}()
	}

	return <-errChan
}

// setupRoutes configures URL routing
func (s *Server) setupRoutes() {
	// Microservice routes (prefix: /ms/ or /api/)
	s.Mux.HandleFunc("/ms/", s.handleMicroservice)
	s.Mux.HandleFunc("/api/", s.handleMicroservice)

	// Web routes (everything else)
	s.Mux.HandleFunc("/", s.handleWebRequest)
}

// handleMicroservice handles microservice requests
func (s *Server) handleMicroservice(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Headers", "*")
	res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if req.Method == "OPTIONS" {
		res.WriteHeader(200)
		return
	}

	req.ParseForm()

	// Extract service name from path
	path := req.URL.Path
	if strings.HasPrefix(path, "/ms/") {
		path = path[4:]
	} else if strings.HasPrefix(path, "/api/") {
		path = path[5:]
	}

	// Find script file
	scriptPath := filepath.Join(s.Config.MSPath, path)
	if !strings.HasSuffix(scriptPath, ".xxl") {
		scriptPath += ".xxl"
	}

	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		res.WriteHeader(404)
		res.Write([]byte("Service not found: " + path))
		return
	}

	// Execute script
	s.executeScript(scriptPath, res, req)
}

// handleWebRequest handles web page requests
func (s *Server) handleWebRequest(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Default to index.xxl for root
	if path == "/" {
		path = "/index.xxl"
	}

	// Check for dynamic page (.xxl)
	scriptPath := filepath.Join(s.Config.WebPath, path)

	if strings.HasSuffix(scriptPath, ".xxl") {
		if _, err := os.Stat(scriptPath); err == nil {
			s.executeScript(scriptPath, res, req)
			return
		}
	}

	// Check for XHP dynamic page (.xhp)
	if strings.HasSuffix(scriptPath, ".xhp") {
		if _, err := os.Stat(scriptPath); err == nil {
			s.executeXHP(scriptPath, res, req)
			return
		}
	}

	// Serve static file
	staticPath := filepath.Join(s.Config.WebPath, path)
	http.ServeFile(res, req, staticPath)
}

// executeScript executes an Xxlang script for HTTP request
func (s *Server) executeScript(scriptPath string, res http.ResponseWriter, req *http.Request) {
	// Read script
	code, err := os.ReadFile(scriptPath)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte("Failed to read script: " + err.Error()))
		return
	}

	// Parse form data
	req.ParseForm()
	paraMap := make(map[string]string)
	for k, v := range req.Form {
		if len(v) > 0 {
			paraMap[k] = v[0]
		}
	}

	// Parse URL query parameters
	for k, v := range req.URL.Query() {
		if len(v) > 0 && paraMap[k] == "" {
			paraMap[k] = v[0]
		}
	}

	// Run script with globals
	result, err := RunScriptOnHttp(string(code), scriptPath, res, req, paraMap, map[string]interface{}{
		"webPathG":  s.Config.WebPath,
		"basePathG": s.Config.MSPath,
		"msPathG":   s.Config.MSPath,
		"runModeG":  "server",
	})

	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte("Script error: " + err.Error()))
		return
	}

	// Special return value means response already handled
	if result == EndResponseMarker {
		return
	}

	// Write result as response
	if result != "" {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.Write([]byte(result))
	}
}

// executeXHP executes an XHP dynamic page for HTTP request
// XHP files contain HTML with embedded <?xhp ... ?> code blocks
func (s *Server) executeXHP(scriptPath string, res http.ResponseWriter, req *http.Request) {
	// Read XHP file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte("Failed to read XHP file: " + err.Error()))
		return
	}

	// Parse form data
	req.ParseForm()
	paraMap := make(map[string]string)
	for k, v := range req.Form {
		if len(v) > 0 {
			paraMap[k] = v[0]
		}
	}

	// Parse URL query parameters
	for k, v := range req.URL.Query() {
		if len(v) > 0 && paraMap[k] == "" {
			paraMap[k] = v[0]
		}
	}

	// Process XHP content
	result, err := ProcessXHP(string(content), scriptPath, res, req, paraMap, map[string]interface{}{
		"webPathG":  s.Config.WebPath,
		"basePathG": s.Config.MSPath,
		"msPathG":   s.Config.MSPath,
		"runModeG":  "server",
	})

	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte("XHP error: " + err.Error()))
		return
	}

	// Write result as HTML response
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write([]byte(result))
}

// RunScriptOnHttp executes Xxlang code for HTTP request
func RunScriptOnHttp(code, scriptPath string, res http.ResponseWriter, req *http.Request,
	paraMap map[string]string, globals map[string]interface{}) (string, error) {

	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parser errors: %v", p.Errors())
	}

	// Compilation
	c := compiler.NewRegCompiler()
	c.SetSourceFile(scriptPath)

	// Define global variable symbols
	symbols := make(map[string]compiler.Symbol)
	for name := range globals {
		symbols[name] = c.SymbolTable().Define(name)
	}
	// Define HTTP-specific globals
	symbols["requestG"] = c.SymbolTable().Define("requestG")
	symbols["responseG"] = c.SymbolTable().Define("responseG")
	symbols["paraMapG"] = c.SymbolTable().Define("paraMapG")
	symbols["reqNameG"] = c.SymbolTable().Define("reqNameG")
	symbols["reqUriG"] = c.SymbolTable().Define("reqUriG")
	symbols["methodG"] = c.SymbolTable().Define("methodG")
	symbols["webPathG"] = c.SymbolTable().Define("webPathG")
	symbols["basePathG"] = c.SymbolTable().Define("basePathG")

	if _, err := c.Compile(program); err != nil {
		return "", fmt.Errorf("compiler error: %v", err)
	}

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    scriptPath,
		Exports: make(map[string]objects.Object),
	}

	// Prepare globals array with preset values
	vmGlobals := make([]vm.Value, compiler.GlobalsSize)

	// Set custom globals
	for name, value := range globals {
		if sym, ok := symbols[name]; ok {
			vmGlobals[sym.Index] = convertToVMValue(value)
		}
	}

	// Set HTTP request object
	if sym, ok := symbols["requestG"]; ok {
		vmGlobals[sym.Index] = vm.NewObject(objects.NewHttpReq(req))
	}

	// Set HTTP response object
	if sym, ok := symbols["responseG"]; ok {
		vmGlobals[sym.Index] = vm.NewObject(objects.NewHttpResp(res))
	}

	// Set paraMapG
	if sym, ok := symbols["paraMapG"]; ok {
		pairs := make(map[objects.HashKey]objects.MapPair)
		for k, v := range paraMap {
			key := objects.NewString(k)
			pairs[key.HashKey()] = objects.MapPair{
				Key:   key,
				Value: objects.NewString(v),
			}
		}
		vmGlobals[sym.Index] = vm.NewObject(objects.NewMap(pairs))
	}

	// Set HTTP-specific globals (only if request is provided)
	if req != nil {
		// Set reqNameG (last segment of path)
		reqName := req.URL.Path
		if idx := strings.LastIndex(reqName, "/"); idx >= 0 {
			reqName = reqName[idx+1:]
		}
		if sym, ok := symbols["reqNameG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(reqName))
		}

		// Set reqUriG
		if sym, ok := symbols["reqUriG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(req.RequestURI))
		}

		// Set methodG
		if sym, ok := symbols["methodG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(req.Method))
		}
	} else {
		// Set empty values when no request
		if sym, ok := symbols["reqNameG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(""))
		}
		if sym, ok := symbols["reqUriG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(""))
		}
		if sym, ok := symbols["methodG"]; ok {
			vmGlobals[sym.Index] = vm.NewObject(objects.NewString(""))
		}
	}

	// Set webPathG - web root directory
	webPath := ""
	if v, ok := globals["webPathG"]; ok {
		if s, ok := v.(string); ok {
			webPath = s
		}
	}
	if sym, ok := symbols["webPathG"]; ok {
		vmGlobals[sym.Index] = vm.NewObject(objects.NewString(webPath))
	}

	// Set basePathG - microservice root directory
	basePath := ""
	if v, ok := globals["basePathG"]; ok {
		if s, ok := v.(string); ok {
			basePath = s
		}
	}
	if sym, ok := symbols["basePathG"]; ok {
		vmGlobals[sym.Index] = vm.NewObject(objects.NewString(basePath))
	}

	// Execution
	bytecode := c.Bytecode()
	v := vm.NewRegVMWithGlobals(bytecode, vmGlobals)
	v.SetSourcePath(scriptPath)
	v.SetCurrentModule(mainModule)

	// Set up module loader
	v.SetLoader(module.NewLoader())

	if err := v.Run(); err != nil {
		return "", fmt.Errorf("runtime error: %v", err)
	}

	result := v.LastResult()
	if result == vm.ValueNull || result.ToObject() == objects.NULL {
		return "", nil
	}

	return result.ToObject().Inspect(), nil
}

// convertToVMValue converts a Go value to a VM Value
func convertToVMValue(value interface{}) vm.Value {
	switch v := value.(type) {
	case string:
		return vm.NewObject(objects.NewString(v))
	case int:
		return vm.NewObject(objects.NewInt(int64(v)))
	case int64:
		return vm.NewObject(objects.NewInt(v))
	case float64:
		return vm.NewObject(objects.NewFloat(v))
	case bool:
		if v {
			return vm.NewObject(objects.TRUE)
		}
		return vm.NewObject(objects.FALSE)
	case nil:
		return vm.NewObject(objects.NULL)
	case map[string]interface{}:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for key, val := range v {
			k := objects.NewString(key)
			pairs[k.HashKey()] = objects.MapPair{
				Key:   k,
				Value: convertToObject(val),
			}
		}
		return vm.NewObject(objects.NewMap(pairs))
	case []interface{}:
		elements := make([]objects.Object, len(v))
		for i, val := range v {
			elements[i] = convertToObject(val)
		}
		return vm.NewObject(objects.NewArray(elements))
	default:
		return vm.NewObject(objects.NewString(fmt.Sprintf("%v", v)))
	}
}

// convertToObject converts a Go value to an Object
func convertToObject(value interface{}) objects.Object {
	switch v := value.(type) {
	case string:
		return objects.NewString(v)
	case int:
		return objects.NewInt(int64(v))
	case int64:
		return objects.NewInt(v)
	case float64:
		return objects.NewFloat(v)
	case bool:
		if v {
			return objects.TRUE
		}
		return objects.FALSE
	case nil:
		return objects.NULL
	case map[string]interface{}:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for key, val := range v {
			k := objects.NewString(key)
			pairs[k.HashKey()] = objects.MapPair{
				Key:   k,
				Value: convertToObject(val),
			}
		}
		return objects.NewMap(pairs)
	case []interface{}:
		elements := make([]objects.Object, len(v))
		for i, val := range v {
			elements[i] = convertToObject(val)
		}
		return objects.NewArray(elements)
	default:
		return objects.NewString(fmt.Sprintf("%v", v))
	}
}

// LoadConfig loads server configuration from a JSON file
func LoadConfig(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %v", err)
	}

	return cfg, nil
}

// XHP tag markers
const (
	XHPStartTag = "<?xhp"
	XHPEndTag   = "?>"
)

// ProcessXHP processes an XHP file, executing embedded code blocks
// XHP files contain HTML with <?xhp ... ?> code blocks that are executed
// and their return values replace the code blocks in the output
// All code blocks in the same XHP file share the same context (variables)
func ProcessXHP(content, scriptPath string, res http.ResponseWriter, req *http.Request,
	paraMap map[string]string, globals map[string]interface{}) (string, error) {

	// Build combined script that:
	// 1. Uses a variable to accumulate output
	// 2. Outputs HTML parts directly
	// 3. Executes code blocks directly (sharing context)
	// 4. return statements in code blocks output their values
	var combinedScript strings.Builder
	pos := 0

	combinedScript.WriteString("var __xhp_result = \"\"\n")
	combinedScript.WriteString("func echo(s) { if (s != null) { __xhp_result = __xhp_result + toStr(s) } }\n")

	for {
		// Find next XHP start tag
		startIdx := strings.Index(content[pos:], XHPStartTag)
		if startIdx == -1 {
			// No more XHP blocks, output remaining HTML
			remaining := content[pos:]
			if remaining != "" {
				escaped := escapeForXHP(remaining)
				combinedScript.WriteString(fmt.Sprintf("echo(\"%s\")\n", escaped))
			}
			break
		}

		// Adjust index to absolute position
		startIdx += pos

		// Output HTML before this block
		if startIdx > pos {
			htmlPart := content[pos:startIdx]
			escaped := escapeForXHP(htmlPart)
			combinedScript.WriteString(fmt.Sprintf("echo(\"%s\")\n", escaped))
		}

		// Find the end of the start tag (after "<?xhp")
		codeStart := startIdx + len(XHPStartTag)

		// Find the closing "?>"
		endIdx := strings.Index(content[codeStart:], XHPEndTag)
		if endIdx == -1 {
			return "", fmt.Errorf("unclosed XHP block at position %d", startIdx)
		}

		// Calculate absolute end position
		endIdx += codeStart

		// Extract the code block
		code := strings.TrimSpace(content[codeStart:endIdx])

		// Transform the code: replace 'return X' with 'echo(X); return'
		// This allows return statements to output values while maintaining context
		if code != "" {
			transformedCode := transformXHPCode(code)
			combinedScript.WriteString(transformedCode)
			combinedScript.WriteString("\n")
		}

		// Move position past the end tag
		pos = endIdx + len(XHPEndTag)
	}

	// Return the accumulated result
	combinedScript.WriteString("return __xhp_result\n")

	// Execute the combined script
	result, err := RunScriptOnHttp(combinedScript.String(), scriptPath, res, req, paraMap, globals)
	if err != nil {
		return "", fmt.Errorf("XHP execution error: %v", err)
	}

	return result, nil
}

// transformXHPCode transforms XHP code to handle return statements
// 'return X' becomes 'echo(X); return'
func transformXHPCode(code string) string {
	// Simple transformation: find 'return' statements and convert them
	// This is a basic implementation - a full parser would be more robust

	lines := strings.Split(code, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if line starts with 'return'
		if strings.HasPrefix(trimmed, "return ") {
			// Extract the expression after 'return'
			expr := strings.TrimPrefix(trimmed, "return ")
			expr = strings.TrimSuffix(expr, "}")
			expr = strings.TrimSpace(expr)

			// Check for trailing brace (end of block)
			trailingBrace := ""
			if strings.HasSuffix(trimmed, "}") && !strings.HasPrefix(expr, "{") {
				trailingBrace = "}"
				expLen := len(expr)
				for expLen > 0 && expr[expLen-1] == '}' {
					expLen--
				}
				expr = expr[:expLen]
			}

			// Generate transformed code
			if expr != "" {
				result.WriteString("echo(" + expr + ")\n")
			}
			if trailingBrace != "" {
				result.WriteString(trailingBrace + "\n")
			}
		} else {
			result.WriteString(line + "\n")
		}
	}

	return result.String()
}

// escapeForXHP escapes a string for inclusion in XHP generated script
func escapeForXHP(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
