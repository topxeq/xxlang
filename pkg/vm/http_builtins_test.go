// pkg/vm/http_builtins_test.go
// Tests for HTTP built-in function visibility to the compiler.
// Regression for the bug where HttpBuiltins (writeResp, setRespHeader, etc.)
// were implemented but never listed in BuiltinRegistry, so the compiler
// rejected them as "undefined variable".
package vm

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// TestHttpBuiltinsInRegistry verifies that all 16 HttpBuiltins names plus
// the helper builtins are listed in BuiltinRegistry, which is what the
// compiler's symbol table iterates over to register BuiltinScope symbols.
func TestHttpBuiltinsInRegistry(t *testing.T) {
	want := []string{
		"writeResp", "setRespHeader", "addRespHeader", "setContentType",
		"status", "getReqHeader", "getReqHeaders", "setCookie",
		"getCookie", "getCookies", "redirect", "serveFile",
		"queryParam", "queryParams", "formValue", "parseForm",
		// helpers that were already registered
		"httpStatusName", "isHttpReq", "isHttpResp",
	}

	registry := make(map[string]bool, len(objects.BuiltinRegistry))
	for _, name := range objects.BuiltinRegistry {
		registry[name] = true
	}

	for _, name := range want {
		if !registry[name] {
			t.Errorf("BuiltinRegistry missing %q", name)
		}
	}
}

// TestHttpBuiltinsResolvedByCompiler verifies that the compiler's symbol
// table resolves these names to BuiltinScope (not "undefined").
func TestHttpBuiltinsResolvedByCompiler(t *testing.T) {
	names := []string{
		"writeResp", "setRespHeader", "addRespHeader", "setContentType",
		"status", "getReqHeader", "getReqHeaders", "setCookie",
		"getCookie", "getCookies", "redirect", "serveFile",
		"queryParam", "queryParams", "formValue", "parseForm",
		"httpStatusName", "isHttpReq", "isHttpResp",
	}

	st := compiler.NewSymbolTable()
	for _, name := range names {
		sym, ok := st.Resolve(name)
		if !ok {
			t.Errorf("compiler did not resolve %q", name)
			continue
		}
		if sym.Scope != compiler.BuiltinScope {
			t.Errorf("%q resolved to scope %v, want BuiltinScope", name, sym.Scope)
		}
	}
}

// TestHttpBuiltinsCompileAndRun verifies the end-to-end path: the script
// compiles (no "undefined variable") and the builtins execute against a real
// httptest.ResponseRecorder / http.Request.
func TestHttpBuiltinsCompileAndRun(t *testing.T) {
	// Build request and response objects
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?foo=bar", nil)

	respObj := objects.NewHttpResp(rec)
	reqObj := objects.NewHttpReq(req)

	// Script uses setRespHeader + writeResp + queryParam + status, then
	// returns the end-response marker. We inject resp/req as globals.
	src := `setRespHeader(responseG, "Content-Type", "text/plain; charset=utf-8")
status(responseG, 201)
writeResp(responseG, "hello", requestG)
return queryParam(requestG, "foo")
`

	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}

	c := compiler.NewRegCompiler()
	// Predefine requestG / responseG as globals (server normally injects these).
	symReq := c.SymbolTable().Define("requestG")
	symResp := c.SymbolTable().Define("responseG")
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("compile error (this is the bug being fixed): %v", err)
	}

	// Build a globals array (same pattern as pkg/server uses) and inject
	// the request/response objects at the indices the compiler assigned.
	vmGlobals := make([]Value, compiler.GlobalsSize)
	vmGlobals[symReq.Index] = NewObject(reqObj)
	vmGlobals[symResp.Index] = NewObject(respObj)

	v := NewRegVMWithGlobals(c.Bytecode(), vmGlobals)

	if err := v.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}

	// Verify response side effects
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
	}
	if rec.Code != 201 {
		t.Errorf("status code = %d, want 201", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "hello")
	}

	// Verify return value (queryParam result)
	result := v.LastResult().ToObject()
	if s, ok := result.(*objects.String); !ok || s.Value != "bar" {
		t.Errorf("return value = %v, want \"bar\"", result.Inspect())
	}
}

// TestHttpBuiltinsNoLongerUndefined is the minimal reproduction of the
// reported bug: just compiling `setRespHeader(responseG, ...)` must not fail.
func TestHttpBuiltinsNoLongerUndefined(t *testing.T) {
	scripts := []string{
		`setRespHeader(responseG, "Content-Type", "text/plain")`,
		`writeResp(responseG, "hello", requestG)`,
		`setContentType(responseG, "text/plain")`,
		`addRespHeader(responseG, "X-Custom", "v")`,
		`status(responseG, 404)`,
		`getReqHeader(requestG, "Host")`,
		`getReqHeaders(requestG, "Accept")`,
		`setCookie(responseG, "k", "v")`,
		`getCookie(requestG, "k")`,
		`getCookies(requestG)`,
		`redirect(responseG, "/x", 302)`,
		`serveFile(responseG, requestG, "/x")`,
		`queryParam(requestG, "q")`,
		`queryParams(requestG)`,
		`formValue(requestG, "f")`,
		`parseForm(requestG)`,
		`httpStatusName(200)`,
		`isHttpReq(requestG)`,
		`isHttpResp(responseG)`,
	}

	for _, src := range scripts {
		l := lexer.New(src)
		p := parser.New(l)
		program := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			t.Errorf("parse error for %q: %v", src, errs)
			continue
		}
		c := compiler.NewRegCompiler()
		// Predefine the server-injected globals so compilation focuses on
		// whether the builtin names resolve, not on requestG/responseG.
		c.SymbolTable().Define("requestG")
		c.SymbolTable().Define("responseG")
		if _, err := c.Compile(program); err != nil {
			if strings.Contains(err.Error(), "undefined variable") {
				t.Errorf("compile %q: %v (the bug being fixed)", src, err)
			}
		}
	}
}
