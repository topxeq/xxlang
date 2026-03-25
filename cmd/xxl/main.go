package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Database drivers - imported for their side effects (driver registration)
	// These drivers are only available when using the main xxl executable
	// For embedded use, import these drivers in your application
	_ "github.com/glebarez/go-sqlite"        // SQLite3 driver (pure Go)
	_ "github.com/go-sql-driver/mysql"       // MySQL driver (pure Go)
	_ "github.com/jackc/pgx/v5/stdlib"       // PostgreSQL driver (pure Go)
	_ "github.com/microsoft/go-mssqldb"      // MSSQL Server driver (pure Go)
	_ "github.com/sijms/go-ora/v2"           // Oracle driver (pure Go)

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/server"
	"github.com/topxeq/xxlang/pkg/stdlib"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITEnabled controls whether JIT compilation is enabled
// Default is false (interpreter mode)
var useJIT = false

// JITConfig holds JIT compiler settings
var jitHotThreshold = 100
var jitMaxCodeSize = 4096
var jitDebug = false

// DebugMode controls whether comprehensive debug output is enabled
var debugMode = false

// ViewMode controls whether to only display script content without executing
var viewMode = false

// PipeMode controls whether to read code from stdin
var pipeMode = false

const (
	PROMPT          = ">> "
	CONTINUE_PROMPT = ".. "
)

// Version is set via -ldflags at build time. Default is "dev" for local builds.
var Version = "dev"

// BuildNumber is a hardcoded build number for development builds.
// Format: YYYYMMDDNN (year month day + daily sequence number, e.g., 2026032401)
// This should be updated manually for each significant build.
var BuildNumber = "2026032402"

// REPL represents an interactive REPL session
type REPL struct {
	symbolTable *compiler.SymbolTable
	constants   []objects.Object
	globals     []objects.Object
	history     []string
}

// NewREPL creates a new REPL session
func NewREPL() *REPL {
	st := compiler.NewSymbolTable()

	// Define preset global variables
	argsGSymbol := st.Define("argsG")
	scriptPathGSymbol := st.Define("scriptPathG")

	// Initialize globals array
	globals := make([]objects.Object, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[argsGSymbol.Index] = &objects.Array{Elements: argsElements}

	// Set scriptPathG - empty for REPL mode
	globals[scriptPathGSymbol.Index] = &objects.String{Value: ""}

	return &REPL{
		symbolTable: st,
		constants:   make([]objects.Object, 0),
		globals:     globals,
		history:     make([]string, 0),
	}
}

// splitArgs splits command line arguments at the -- separator.
// Returns (interpreterArgs, scriptArgs).
// Everything after -- is passed to the script.
func splitArgs(args []string) (interpreterArgs []string, scriptArgs []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func main() {
	// No arguments - start REPL
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	// Split arguments at -- separator
	interpreterArgs, scriptArgs := splitArgs(os.Args[1:])

	// Parse JIT flags
	interpreterArgs = parseFlags(interpreterArgs)

	// Set script args for scripts to access via env::scriptArgs()
	stdlib.SetScriptArgs(scriptArgs)

	// Check for -pipe flag to read code from stdin
	if pipeMode {
		code, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		executeCode(string(code), "<stdin>")
		return
	}

	// Check for help or version flags (only in interpreter args)
	for _, arg := range interpreterArgs {
		if arg == "--help" || arg == "-h" || arg == "help" {
			printUsage()
			return
		}
		if arg == "--version" || arg == "-v" || arg == "version" {
			printVersion()
			return
		}
	}

	// Check for -cloud flag
	if len(interpreterArgs) >= 2 && interpreterArgs[0] == "-cloud" {
		runFromCloud(interpreterArgs[1])
		return
	}

	// Check for -e/--eval flag to execute code directly
	if len(interpreterArgs) >= 2 && (interpreterArgs[0] == "-e" || interpreterArgs[0] == "--eval") {
		executeCode(interpreterArgs[1], "<eval>")
		return
	}

	switch interpreterArgs[0] {
	case "compile":
		if err := compileCmd(interpreterArgs[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if len(interpreterArgs) < 2 {
			fmt.Println("Usage: xxl run <file|url> [-- script args...]")
			os.Exit(1)
		}
		runFileOrURL(interpreterArgs[1])
	case "serve":
		if err := serveCmd(interpreterArgs[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "update":
		if err := updateCmd(interpreterArgs[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		printVersion()
	case "help":
		printUsage()
	default:
		// Check if it's a file or URL that we can run directly
		arg := interpreterArgs[0]
		if isURL(arg) || strings.HasSuffix(arg, ".xxl") || strings.HasSuffix(arg, ".xxb") {
			runFileOrURL(arg)
		} else {
			// Unknown subcommand - show error
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", arg)
			printUsage()
			os.Exit(1)
		}
	}
}

// parseFlags extracts and processes JIT-related and debug flags
// Returns the remaining arguments
func parseFlags(args []string) []string {
	var result []string
	for _, arg := range args {
		if arg == "--jit" {
			useJIT = true
		} else if arg == "--jit-debug" {
			useJIT = true
			jitDebug = true
		} else if strings.HasPrefix(arg, "--jit-threshold=") {
			threshold := strings.TrimPrefix(arg, "--jit-threshold=")
			fmt.Sscanf(threshold, "%d", &jitHotThreshold)
			useJIT = true
		} else if arg == "--no-jit" {
			useJIT = false
		} else if arg == "--debug" {
			debugMode = true
		} else if arg == "-view" || arg == "--view" {
			viewMode = true
		} else if arg == "-pipe" || arg == "--pipe" {
			pipeMode = true
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func printUsage() {
	if Version == "dev" {
		fmt.Printf("Xxlang v%s.%s\n", Version, BuildNumber)
	} else {
		fmt.Printf("Xxlang v%s\n", Version)
	}
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  xxl                         Start interactive REPL")
	fmt.Println("  xxl <file|url> [-- args...] Execute a .xxl file or script from URL")
	fmt.Println("  xxl -e <code>               Execute code directly from command line")
	fmt.Println("  echo 'code' | xxl -pipe     Execute code from stdin")
	fmt.Println("  xxl run <file|url> [-- args...]")
	fmt.Println("                              Execute a .xxl file or script from URL")
	fmt.Println("  xxl serve [options]         Start HTTP/HTTPS server")
	fmt.Println("  xxl -cloud <script>         Execute script from cloud URL base")
	fmt.Println("  xxl compile <file>          Compile file to bytecode")
	fmt.Println("  xxl update                  Self-update to latest version from GitHub")
	fmt.Println("  xxl version                 Print version information")
	fmt.Println("  xxl help                    Print this help message")
	fmt.Println()
	fmt.Println("Serve Options:")
	fmt.Println("  -web=<path>       Web root path (default: .)")
	fmt.Println("  -ms=<path>        Microservice root path (default: .)")
	fmt.Println("  -cert=<path>      Certificate path (default: .)")
	fmt.Println("  -http=<port>      HTTP port (default: 80, 0 to disable)")
	fmt.Println("  -https=<port>     HTTPS port (default: 443, 0 to disable)")
	fmt.Println("  -config=<file>    Configuration file path (JSON format)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o, --output path     Output path for compiled file")
	fmt.Println("      --target os/arch  Cross-compile for target OS/architecture")
	fmt.Println("      --bytecode        Output as bytecode (.xxb) instead of executable")
	fmt.Println("  -cloud <script>       Execute script from configured cloudUrlBase")
	fmt.Println("  -e, --eval <code>     Execute code directly from command line")
	fmt.Println("  -pipe, --pipe         Read and execute code from stdin")
	fmt.Println("  -view, --view         View script content without executing")
	fmt.Println("      --jit             Enable JIT compilation for hot paths (experimental)")
	fmt.Println("      --jit-threshold=N Set JIT hot path threshold (default: 100)")
	fmt.Println("      --no-jit          Disable JIT compilation (default)")
	fmt.Println("      --debug           Show debug info (bytecode count, runtime, JIT usage)")
	fmt.Println()
	fmt.Println("Script Arguments:")
	fmt.Println("  Use '--' to separate interpreter arguments from script arguments.")
	fmt.Println("  Script arguments are accessible via env::scriptArgs().")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  xxl")
	fmt.Println("  xxl script.xxl")
	fmt.Println("  xxl -e 'pln(\"Hello, World!\")'    Execute code directly")
	fmt.Println("  xxl -e '1 + 2 + 3'                Execute expression")
	fmt.Println("  echo 'pln(1+2+3)' | xxl -pipe     Execute code from stdin")
	fmt.Println("  cat script.xxl | xxl -pipe        Execute file from stdin")
	fmt.Println("  echo '{\"f1\":\"v1\"}' | xxl fmt.xxl  Pipe data to script (use scan/read)")
	fmt.Println("  xxl -view script.xxl              View script content without executing")
	fmt.Println("  xxl -view https://example.com/script.xxl")
	fmt.Println("  xxl script.xxl -- arg1 arg2 --help")
	fmt.Println("  xxl run script.xxl -- --verbose -f file.txt")
	fmt.Println("  xxl https://raw.githubusercontent.com/user/repo/main/script.xxl")
	fmt.Println("  xxl compile -o program script.xxl")
	fmt.Println("  xxl compile -o program.exe --target windows/amd64 script.xxl")
	fmt.Println("  xxl compile --bytecode script.xxl")
	fmt.Println("  xxl -cloud basic.xxl")
	fmt.Println("  xxl update")
}

func printVersion() {
	if Version == "dev" {
		fmt.Printf("Xxlang v%s.%s\n", Version, BuildNumber)
	} else {
		fmt.Printf("Xxlang v%s\n", Version)
	}
}

func startREPL() {
	repl := NewREPL()
	scanner := bufio.NewScanner(os.Stdin)

	if Version == "dev" {
		fmt.Println("Xxlang REPL v" + Version + "." + BuildNumber)
	} else {
		fmt.Println("Xxlang REPL v" + Version)
	}
	fmt.Println("Type 'exit' or 'quit' to exit, 'help' for help, 'history' for command history")
	fmt.Println("Multi-line: end line with '{' to continue")
	fmt.Print(PROMPT)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			fmt.Print(PROMPT)
			continue
		}

		// Handle special commands
		switch line {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return
		case "help":
			printHelp()
			fmt.Print(PROMPT)
			continue
		case "history":
			printHistory(repl.history)
			fmt.Print(PROMPT)
			continue
		case "clear":
			repl = NewREPL()
			fmt.Println("Cleared all variables and functions")
			fmt.Print(PROMPT)
			continue
		}

		// Check for multi-line input
		input := line
		for countOpenBraces(input) > 0 {
			fmt.Print(CONTINUE_PROMPT)
			if !scanner.Scan() {
				break
			}
			nextLine := scanner.Text()
			input += "\n" + nextLine
		}

		// Add to history
		repl.history = append(repl.history, input)

		// Parse and execute the input
		result, err := repl.Execute(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if result != nil && result != objects.NULL {
			fmt.Println(result.Inspect())
		}

		fmt.Print(PROMPT)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// Execute compiles and runs code in the REPL context
func (r *REPL) Execute(input string) (objects.Object, error) {
	// Lexical analysis
	l := lexer.New(input)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, formatParserErrors(p.Errors())
	}

	return r.executeRegister(program)
}

// executeRegister uses the register-based VM for REPL execution
func (r *REPL) executeRegister(program *parser.Program) (objects.Object, error) {
	// Compilation with register compiler
	c := compiler.NewRegCompiler()
	c.SetSymbolTable(r.symbolTable)
	c.SetConstants(r.constants)
	if _, err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error: %v", err)
	}

	// Update constants
	r.constants = c.Bytecode().Constants

	// Execution with persistent globals
	bytecode := c.Bytecode()
	v := vm.NewRegVMWithObjectGlobals(bytecode, r.globals)

	if err := v.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	// Update globals for next execution
	r.globals = v.GlobalsAsObjects()

	return v.LastPoppedObject(), nil
}

// isURL checks if the given string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// runFromCloud fetches and executes a script from the cloud URL base configured in settings
func runFromCloud(scriptName string) {
	// Get config map
	cfg := stdlib.GetConfigMap()

	// Get cloudUrlBase from config
	cloudUrlBase, ok := cfg["cloudUrlBase"].(string)
	if !ok || cloudUrlBase == "" {
		fmt.Fprintln(os.Stderr, "Error: cloudUrlBase not configured in settings.json")
		fmt.Fprintln(os.Stderr, "Please add a 'cloudUrlBase' field to your ~/.xxl/settings.json file")
		os.Exit(1)
	}

	// Construct full URL
	// Ensure the base URL ends with / and the script name is appended
	baseURL := strings.TrimSuffix(cloudUrlBase, "/")
	fullURL := baseURL + "/" + scriptName

	// Run from the constructed URL
	runFromURL(fullURL)
}

// runFileOrURL runs an xxlang source file or script from URL
func runFileOrURL(source string) {
	if isURL(source) {
		runFromURL(source)
	} else {
		runFile(source)
	}
}

// runFromURL fetches and executes an xxlang script from a URL
func runFromURL(url string) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the script
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching URL '%s': %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP status %d (%s)\n", resp.StatusCode, resp.Status)
		os.Exit(1)
	}

	// Read the response body
	code, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// If view mode, just print the content and return
	if viewMode {
		fmt.Printf("// Source: %s\n", url)
		fmt.Print(string(code))
		return
	}

	// Execute the code
	executeCode(string(code), url)
}

// runFile runs an xxlang source file
func runFile(filename string) {
	// Get absolute path for module resolution
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error resolving path '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// Check if it's a bytecode file
	if strings.HasSuffix(absPath, ".xxb") {
		if viewMode {
			// Bytecode files are binary, just show info
			fmt.Printf("// Binary bytecode file: %s\n", absPath)
			info, err := os.Stat(absPath)
			if err == nil {
				fmt.Printf("// Size: %d bytes\n", info.Size())
			}
			return
		}
		runBytecodeFile(absPath)
		return
	}

	// Read the source file
	code, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Printf("Error reading file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// If view mode, just print the content and return
	if viewMode {
		fmt.Printf("// Source: %s\n", absPath)
		fmt.Print(string(code))
		return
	}

	// Execute the code
	executeCode(string(code), absPath)
}

// executeCode compiles and executes xxlang source code
func executeCode(code, sourcePath string) {
	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Printf("Error: %v\n", formatParserErrors(p.Errors()))
		os.Exit(1)
	}

	executeCodeRegister(program, sourcePath, code)
}

// executeCodeRegister uses the register-based VM
func executeCodeRegister(program *parser.Program, sourcePath, code string) {
	// Compilation
	c := compiler.NewRegCompiler()
	// Define preset global variables before compilation
	argsGSymbol := c.SymbolTable().Define("argsG")
	scriptPathGSymbol := c.SymbolTable().Define("scriptPathG")
	c.SetSourceFile(sourcePath)

	compileStart := time.Now()
	if _, err := c.Compile(program); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	compileTime := time.Since(compileStart)

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    sourcePath,
		Exports: make(map[string]objects.Object),
	}

	// Prepare globals array with preset values
	globals := make([]vm.Value, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[argsGSymbol.Index] = vm.NewObject(&objects.Array{Elements: argsElements})

	// Set scriptPathG - script path
	globals[scriptPathGSymbol.Index] = vm.NewObject(&objects.String{Value: sourcePath})

	// Execution
	bytecode := c.Bytecode()

	// Print debug info before execution
	if debugMode {
		fmt.Fprintf(os.Stderr, "[Debug] Source: %s\n", sourcePath)
		fmt.Fprintf(os.Stderr, "[Debug] Source size: %d bytes\n", len(code))
		fmt.Fprintf(os.Stderr, "[Debug] Bytecode instructions: %d\n", len(bytecode.Instructions))
		fmt.Fprintf(os.Stderr, "[Debug] Constants: %d\n", len(bytecode.Constants))
		fmt.Fprintf(os.Stderr, "[Debug] Compile time: %v\n", compileTime)
		fmt.Fprintf(os.Stderr, "[Debug] JIT enabled: %v\n", useJIT)
	}

	execStart := time.Now()

	// Use JIT VM if enabled
	if useJIT {
		jitConfig := jit.JITConfig{
			HotThreshold: jitHotThreshold,
			MaxCodeSize:  jitMaxCodeSize,
			Debug:        jitDebug,
		}
		jitVM := jit.NewJITVMWithGlobals(bytecode, globals, jitConfig)
		jitVM.SetSourcePath(sourcePath)
		jitVM.SetCurrentModule(mainModule)

		if err := jitVM.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n%s", jitVM.GetCallStack())
			os.Exit(1)
		}
		execTime := time.Since(execStart)

		// Print JIT stats
		stats := jitVM.GetJITStats()
		nativeExecs, interpExecs := jitVM.GetNativeStats()

		if debugMode {
			fmt.Fprintf(os.Stderr, "[Debug] VM mode: JIT (hybrid)\n")
			fmt.Fprintf(os.Stderr, "[Debug] Execution time: %v\n", execTime)
			fmt.Fprintf(os.Stderr, "[Debug] JIT compiled functions: %d\n", stats.CompiledFunctions)
			fmt.Fprintf(os.Stderr, "[Debug] JIT total code size: %d bytes\n", stats.TotalCodeSize)
			fmt.Fprintf(os.Stderr, "[Debug] Native executions: %d\n", nativeExecs)
			fmt.Fprintf(os.Stderr, "[Debug] Interpreter executions: %d\n", interpExecs)
			fmt.Fprintf(os.Stderr, "[Debug] Total time: %v\n", compileTime+execTime)
		} else if stats.CompiledFunctions > 0 {
			fmt.Fprintf(os.Stderr, "[JIT] Compiled %d functions, %d bytes\n", stats.CompiledFunctions, stats.TotalCodeSize)
		}

		jitVM.Cleanup()
		return
	}

	// Standard register VM execution
	v := vm.NewRegVMWithGlobals(bytecode, globals)
	v.SetSourcePath(sourcePath)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\n%s", v.GetCallStack())
		os.Exit(1)
	}
	execTime := time.Since(execStart)

	if debugMode {
		fmt.Fprintf(os.Stderr, "[Debug] VM mode: Interpreter\n")
		fmt.Fprintf(os.Stderr, "[Debug] Execution time: %v\n", execTime)
		fmt.Fprintf(os.Stderr, "[Debug] Total time: %v\n", compileTime+execTime)
	}
}

// runBytecodeFile runs a compiled bytecode file
func runBytecodeFile(filename string) {
	// Read the bytecode file
	bytecodeData, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading bytecode file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// Deserialize the bytecode
	bytecode, err := compiler.Deserialize(bytecodeData)
	if err != nil {
		fmt.Printf("Error loading bytecode: %v\n", err)
		os.Exit(1)
	}

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    filename,
		Exports: make(map[string]objects.Object),
	}

	// Prepare globals array with preset values
	globals := make([]vm.Value, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array (index 0)
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[0] = vm.NewObject(&objects.Array{Elements: argsElements})

	// Set scriptPathG - script path (index 1)
	globals[1] = vm.NewObject(&objects.String{Value: filename})

	// Execution
	v := vm.NewRegVMWithGlobals(bytecode, globals)
	v.SetSourcePath(filename)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print result if it's meaningful
	result := v.LastPoppedObject()
	if result != nil && result != objects.NULL {
		fmt.Println(result.Inspect())
	}
}

func formatParserErrors(errors []string) error {
	var sb strings.Builder
	sb.WriteString("parser errors:\n")
	for _, err := range errors {
		sb.WriteString("  ")
		sb.WriteString(err)
		sb.WriteString("\n")
	}
	return fmt.Errorf("%s", sb.String())
}

// countOpenBraces counts unclosed braces in the input
func countOpenBraces(input string) int {
	count := 0
	inString := false
	escape := false

	for _, ch := range input {
		if escape {
			escape = false
			continue
		}
		switch ch {
		case '\\':
			escape = true
		case '"':
			inString = !inString
		case '{':
			if !inString {
				count++
			}
		case '}':
			if !inString {
				count--
			}
		}
	}

	return count
}

func printHelp() {
	fmt.Print(`Xxlang REPL Commands:
  exit, quit  - Exit the REPL
  help        - Show this help message
  history     - Show command history
  clear       - Clear all variables and functions

Language Features:
  Variables:    var x = 10
  Constants:   const PI = 3.14
  Functions:   func add(a, b) { return a + b }
  Closures:    func makeCounter() { var c = 0; func() { c++ }; }
  Arrays:      [1, 2, 3]
  Maps:        {"key": "value"}
  Loops:       for (var i = 0; i < 10; i++) { }
  Conditionals: if (x > 0) { } else { }

Examples:
  >> var x = 10
  10
  >> x + 5
  15
  >> func fib(n) { if (n <= 1) { return n; } return fib(n-1) + fib(n-2); }
  >> fib(10)
  55

`)
}

func printHistory(history []string) {
	if len(history) == 0 {
		fmt.Println("No commands in history")
		return
	}
	for i, cmd := range history {
		// Truncate long commands for display
		display := cmd
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		// Replace newlines with escaped version
		display = strings.ReplaceAll(display, "\n", "\\n")
		fmt.Printf("%3d: %s\n", i+1, display)
	}
}

// serveCmd handles the serve subcommand
func serveCmd(args []string) error {
	// Default configuration
	cfg := server.Config{
		HTTPPort:  80,
		HTTPSPort: 443,
	}

	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "-web="):
			cfg.WebPath = strings.TrimPrefix(args[i], "-web=")
		case strings.HasPrefix(args[i], "-ms="):
			cfg.MSPath = strings.TrimPrefix(args[i], "-ms=")
		case strings.HasPrefix(args[i], "-cert="):
			cfg.CertPath = strings.TrimPrefix(args[i], "-cert=")
		case strings.HasPrefix(args[i], "-http="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "-http="), "%d", &cfg.HTTPPort)
		case strings.HasPrefix(args[i], "-https="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "-https="), "%d", &cfg.HTTPSPort)
		case strings.HasPrefix(args[i], "-config="):
			// Load from config file
			configFile := strings.TrimPrefix(args[i], "-config=")
			loadedCfg, err := server.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %v", err)
			}
			// Merge with any command line overrides
			if cfg.WebPath == "" {
				cfg.WebPath = loadedCfg.WebPath
			}
			if cfg.MSPath == "" {
				cfg.MSPath = loadedCfg.MSPath
			}
			if cfg.CertPath == "" {
				cfg.CertPath = loadedCfg.CertPath
			}
			if cfg.HTTPPort == 80 && loadedCfg.HTTPPort != 0 {
				cfg.HTTPPort = loadedCfg.HTTPPort
			}
			if cfg.HTTPSPort == 443 && loadedCfg.HTTPSPort != 0 {
				cfg.HTTPSPort = loadedCfg.HTTPSPort
			}
		}
	}

	// Set defaults - both web and ms paths default to current directory
	if cfg.WebPath == "" {
		cfg.WebPath = "."
	}
	if cfg.MSPath == "" {
		cfg.MSPath = "."
	}
	if cfg.CertPath == "" {
		cfg.CertPath = "."
	}

	// Validate that at least one port is enabled
	if cfg.HTTPPort == 0 && cfg.HTTPSPort == 0 {
		return fmt.Errorf("at least one of -http or -https port must be non-zero")
	}

	// Create and start server
	srv := server.NewServer(cfg)
	if Version == "dev" {
		fmt.Printf("Xxlang Server v%s.%s\n", Version, BuildNumber)
	} else {
		fmt.Printf("Xxlang Server v%s\n", Version)
	}
	fmt.Printf("Web path: %s\n", cfg.WebPath)
	fmt.Printf("Microservice path: %s\n", cfg.MSPath)
	if cfg.HTTPSPort > 0 {
		fmt.Printf("Certificate path: %s\n", cfg.CertPath)
	}
	fmt.Println()

	return srv.Start()
}
