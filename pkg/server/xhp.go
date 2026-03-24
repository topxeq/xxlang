// pkg/server/xhp.go
// XHP (Xxlang HTML Processor) implementation using Eval mode for segment-by-segment execution
package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/interpreter"
	"github.com/topxeq/xxlang/pkg/objects"
)

// XHP tag markers
const (
	XHPStartTag = "<?xhp"
	XHPEndTag   = "?>"
)

// XHPProcessor handles XHP template processing with Eval mode execution
type XHPProcessor struct {
	interp     *interpreter.Interpreter
	scriptPath string
	res        http.ResponseWriter
	req        *http.Request
}

// NewXHPProcessor creates a new XHP processor
func NewXHPProcessor(scriptPath string, res http.ResponseWriter, req *http.Request,
	paraMap map[string]string, globals map[string]interface{}) *XHPProcessor {

	// Create interpreter with stdlib
	interp := interpreter.New(interpreter.WithStdlib())

	// Set up global variables
	if globals != nil {
		for name, value := range globals {
			interp.SetGlobal(name, value)
		}
	}

	// Set paraMapG as a global
	if paraMap != nil {
		interp.SetGlobal("paraMapG", paraMap)
	}

	// Set HTTP-specific globals
	if req != nil {
		reqName := req.URL.Path
		if idx := strings.LastIndex(reqName, "/"); idx >= 0 {
			reqName = reqName[idx+1:]
		}
		interp.SetGlobal("reqNameG", reqName)
		interp.SetGlobal("reqUriG", req.RequestURI)
		interp.SetGlobal("methodG", req.Method)
		interp.SetGlobal("requestG", objects.NewHttpReq(req))
	}

	if res != nil {
		interp.SetGlobal("responseG", objects.NewHttpResp(res))
	}

	return &XHPProcessor{
		interp:     interp,
		scriptPath: scriptPath,
		res:        res,
		req:        req,
	}
}

// Process executes all segments and returns the final output
func (p *XHPProcessor) Process(content string) (string, error) {
	// Initialize output buffer and echo function
	// All output goes through the __xhp_output buffer
	// Both echo() and __xhp_echo() are available
	initCode := `
		var __xhp_output = ""
		func __xhp_echo(s) {
			if (s != null) {
				__xhp_output = __xhp_output + toStr(s)
			}
		}
		func echo(s) {
			__xhp_echo(s)
		}
	`
	_, err := p.interp.Eval(initCode)
	if err != nil {
		return "", fmt.Errorf("failed to initialize: %v", err)
	}

	pos := 0

	for {
		// Find next XHP start tag
		startIdx := strings.Index(content[pos:], XHPStartTag)
		if startIdx == -1 {
			// No more XHP blocks, output remaining HTML
			remaining := content[pos:]
			if remaining != "" {
				// Output HTML directly to the buffer
				escaped := escapeForXHP(remaining)
				_, err := p.interp.Eval(fmt.Sprintf(`__xhp_output = __xhp_output + "%s"`, escaped))
				if err != nil {
					return "", err
				}
			}
			break
		}

		startIdx += pos

		// Output HTML before this block
		if startIdx > pos {
			htmlPart := content[pos:startIdx]
			escaped := escapeForXHP(htmlPart)
			_, err := p.interp.Eval(fmt.Sprintf(`__xhp_output = __xhp_output + "%s"`, escaped))
			if err != nil {
				return "", err
			}
		}

		// Find code block
		codeStart := startIdx + len(XHPStartTag)
		endIdx := strings.Index(content[codeStart:], XHPEndTag)
		if endIdx == -1 {
			return "", fmt.Errorf("unclosed XHP block at position %d in %s", startIdx, p.scriptPath)
		}
		endIdx += codeStart

		code := strings.TrimSpace(content[codeStart:endIdx])

		if code != "" {
			// Transform return statements to echo calls and execute
			transformed := transformXHPCode(code)
			_, err := p.interp.Eval(transformed)
			if err != nil {
				return "", fmt.Errorf("code error at position %d: %v", startIdx, err)
			}
		}

		pos = endIdx + len(XHPEndTag)
	}

	// Get the final output
	result, err := p.interp.Eval(`__xhp_output`)
	if err != nil {
		return "", err
	}

	if result == nil || result == objects.NULL {
		return "", nil
	}

	return result.Inspect(), nil
}

// GetInterpreter returns the interpreter for direct access
func (p *XHPProcessor) GetInterpreter() *interpreter.Interpreter {
	return p.interp
}

// ProcessXHP processes an XHP file using Eval mode for segment-by-segment execution
// XHP files contain HTML with <?xhp ... ?> code blocks that are executed
// All code blocks in the same XHP file share the same context (variables)
// return statements output their values, echo() function also outputs
func ProcessXHP(content, scriptPath string, res http.ResponseWriter, req *http.Request,
	paraMap map[string]string, globals map[string]interface{}) (string, error) {

	processor := NewXHPProcessor(scriptPath, res, req, paraMap, globals)
	return processor.Process(content)
}

// transformXHPCode transforms XHP code to handle return statements
// 'return X' becomes '__xhp_echo(X)'
func transformXHPCode(code string) string {
	// Use regex to handle return statements
	re := regexp.MustCompile(`(?m)^\s*return\s+(.+?)(\s*;?\s*)?$`)

	var result strings.Builder
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		matched := re.FindStringSubmatch(line)
		if matched != nil {
			expr := strings.TrimSpace(matched[1])
			expr = strings.TrimSuffix(expr, ";")
			expr = strings.TrimSpace(expr)
			if expr != "" {
				result.WriteString("__xhp_echo(" + expr + ")\n")
			}
		} else {
			result.WriteString(line)
			result.WriteString("\n")
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

// XHPSegment represents a parsed segment from XHP content
type XHPSegment struct {
	Type     XHPSegmentType // HTML or Code
	Content  string         // Raw content
	Position int            // Position in source
}

type XHPSegmentType int

const (
	XHPSegmentHTML XHPSegmentType = iota
	XHPSegmentCode
)

// ParseXHP parses XHP content into segments
func ParseXHP(content string) ([]XHPSegment, error) {
	var segments []XHPSegment
	pos := 0

	for {
		startIdx := strings.Index(content[pos:], XHPStartTag)
		if startIdx == -1 {
			if pos < len(content) {
				segments = append(segments, XHPSegment{
					Type:     XHPSegmentHTML,
					Content:  content[pos:],
					Position: pos,
				})
			}
			break
		}

		startIdx += pos

		if startIdx > pos {
			segments = append(segments, XHPSegment{
				Type:     XHPSegmentHTML,
				Content:  content[pos:startIdx],
				Position: pos,
			})
		}

		codeStart := startIdx + len(XHPStartTag)
		endIdx := strings.Index(content[codeStart:], XHPEndTag)
		if endIdx == -1 {
			return nil, fmt.Errorf("unclosed XHP block at position %d", startIdx)
		}
		endIdx += codeStart

		segments = append(segments, XHPSegment{
			Type:     XHPSegmentCode,
			Content:  strings.TrimSpace(content[codeStart:endIdx]),
			Position: startIdx,
		})

		pos = endIdx + len(XHPEndTag)
	}

	return segments, nil
}

// ExecuteXHPSegments executes parsed XHP segments using Eval mode
func ExecuteXHPSegments(segments []XHPSegment, interp *interpreter.Interpreter) (string, error) {
	// Initialize output buffer and echo function
	_, err := interp.Eval(`
		var __xhp_output = ""
		func __xhp_echo(s) {
			if (s != null) {
				__xhp_output = __xhp_output + toStr(s)
			}
		}
		func echo(s) {
			__xhp_echo(s)
		}
	`)
	if err != nil {
		return "", err
	}

	for i, seg := range segments {
		switch seg.Type {
		case XHPSegmentHTML:
			if seg.Content != "" {
				escaped := escapeForXHP(seg.Content)
				_, err := interp.Eval(fmt.Sprintf(`__xhp_output = __xhp_output + "%s"`, escaped))
				if err != nil {
					return "", fmt.Errorf("segment %d error: %v", i, err)
				}
			}

		case XHPSegmentCode:
			if seg.Content == "" {
				continue
			}
			transformed := transformXHPCode(seg.Content)
			_, err := interp.Eval(transformed)
			if err != nil {
				return "", fmt.Errorf("segment %d error: %v", i, err)
			}
		}
	}

	result, err := interp.Eval(`__xhp_output`)
	if err != nil {
		return "", err
	}

	if result == nil || result == objects.NULL {
		return "", nil
	}

	return result.Inspect(), nil
}

// CreateXHPInterpreter creates an interpreter configured for XHP execution
func CreateXHPInterpreter(scriptPath string, res http.ResponseWriter, req *http.Request,
	paraMap map[string]string, globals map[string]interface{}) (*interpreter.Interpreter, error) {

	interp := interpreter.New(interpreter.WithStdlib())

	if globals != nil {
		for name, value := range globals {
			if err := interp.SetGlobal(name, value); err != nil {
				return nil, fmt.Errorf("failed to set global %s: %v", name, err)
			}
		}
	}

	if paraMap != nil {
		pMap := make(map[string]interface{})
		for k, v := range paraMap {
			pMap[k] = v
		}
		interp.SetGlobal("paraMapG", pMap)
	}

	if req != nil {
		reqName := req.URL.Path
		if idx := strings.LastIndex(reqName, "/"); idx >= 0 {
			reqName = reqName[idx+1:]
		}
		interp.SetGlobal("reqNameG", reqName)
		interp.SetGlobal("reqUriG", req.RequestURI)
		interp.SetGlobal("methodG", req.Method)
		interp.SetGlobal("requestG", objects.NewHttpReq(req))
	}

	if res != nil {
		interp.SetGlobal("responseG", objects.NewHttpResp(res))
	}

	return interp, nil
}

// GetSymbolTable returns the symbol table for XHP processing
func (p *XHPProcessor) GetSymbolTable() *compiler.SymbolTable {
	return nil
}
