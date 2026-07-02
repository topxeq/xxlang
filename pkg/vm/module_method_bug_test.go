// pkg/vm/module_method_bug_test.go
// Tests for the "module method call vs return" bug class.
// Covers:
//   - Bug 1: OpRegTailCallMethod missing Module branch (Cases 2/3/7)
//   - Bug 2: vm.objConstants should be frame.Constants (Case 5 panic)
//   - Bug 3: TCO functions not updating frame.Constants (cross-module TCO)
package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// runRegScriptReturn compiles and runs src, returns the inspect string of
// the last popped value (i.e. the top-level return value).
func runRegScriptReturn(t *testing.T, src string) (string, error) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		return "", err
	}
	v := NewRegVM(c.Bytecode())
	if err := v.Run(); err != nil {
		return "", err
	}
	r := v.LastResult()
	if r.IsNull() {
		return "null", nil
	}
	return r.ToObject().Inspect(), nil
}

// Case 1 PASS — var then return inside function (baseline)
func TestModuleMethod_Case1_VarThenReturn(t *testing.T) {
	src := `import crypto from "crypto"
func f() {
    var s = crypto.sha256("test")
    return s
}
return f()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q, got %q", want, out)
	}
}

// Case 2 — direct return module method call inside function
// Was: "cannot call method 'sha256' on type MODULE"
func TestModuleMethod_Case2_DirectReturnInFunc(t *testing.T) {
	src := `import crypto from "crypto"
func f() {
    return crypto.sha256("test")
}
return f()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q, got %q", want, out)
	}
}

// Case 3 — top-level direct return module method call
// Was: "cannot call method 'sha256' on type MODULE"
func TestModuleMethod_Case3_TopLevelDirectReturn(t *testing.T) {
	src := `import crypto from "crypto"
return crypto.sha256("test")
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q, got %q", want, out)
	}
}

// Case 6 PASS — var then return with regex (baseline)
func TestModuleMethod_Case6_VarThenReturnRegex(t *testing.T) {
	src := `import regex from "regex"
func f() {
    var s = regex.find("a", "abc")
    return s
}
return f()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "a") {
		t.Errorf("expected output to contain 'a', got %q", out)
	}
}

// Case 7 — direct return regex.find in function
// Was: "cannot call method 'find' on type MODULE"
func TestModuleMethod_Case7_DirectReturnRegex(t *testing.T) {
	src := `import regex from "regex"
func f() {
    return regex.find("a", "abc")
}
return f()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "a") {
		t.Errorf("expected output to contain 'a', got %q", out)
	}
}

// Case 5 — imported export function calling module method
// Was: panic: runtime error: index out of range
func TestModuleMethod_Case5_ImportedExportFunc(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "m.xxl")
	modSrc := `import crypto from "crypto"
export func f() {
    return crypto.sha256("test")
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	callerSrc := `import { f } from "` + modPath + `"
return f()
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error (was panic before fix): %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q, got %q", want, out)
	}
}

// Extra: imported export function calling module method with field access
// Exercises OpRegGetField inside an imported function (would have read
// the wrong constants pool before the fix).
func TestModuleMethod_ImportedFuncFieldAccess(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "m2.xxl")
	modSrc := `import strings from "strings"
export func upper(s) {
    return strings.toUpper(s)
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	callerSrc := `import { upper } from "` + modPath + `"
return upper("hello")
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "HELLO") {
		t.Errorf("expected 'HELLO', got %q", out)
	}
}

// Extra: cross-module tail call of an imported function
// Exercises Bug 3 fix (handleRegTailCall updating frame.Constants).
func TestModuleMethod_CrossModuleTailCall(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "mtc.xxl")
	modSrc := `import crypto from "crypto"
export func hashIt(s) {
    return crypto.sha256(s)
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	callerSrc := `import { hashIt } from "` + modPath + `"
func caller(s) {
    return hashIt(s)
}
return caller("test")
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q, got %q", want, out)
	}
}

// Extra: top-level return of strings.toUpper (another module method)
// to ensure the fix generalizes across stdlib modules.
func TestModuleMethod_TopLevelReturnStringsUpper(t *testing.T) {
	src := `import strings from "strings"
return strings.toUpper("abc")
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ABC") {
		t.Errorf("expected 'ABC', got %q", out)
	}
}

// Extra: direct return json.stringify in function
// (the original exp1.md reported `return json.stringify(obj)` failing)
func TestModuleMethod_DirectReturnJsonStringify(t *testing.T) {
	src := `import json from "json"
func f() {
    return json.stringify({"a": 1})
}
return f()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// xxlang's JSON formatter emits unquoted keys for simplicity
	if !strings.Contains(out, "a") || !strings.Contains(out, "1") {
		t.Errorf("expected JSON containing 'a' and '1', got %q", out)
	}
}

// Extra: verify caller's globals are NOT corrupted after a cross-module TCO call.
// After calling an imported function via TCO, the caller should still be able
// to access its own globals correctly.
func TestModuleMethod_CrossModuleTCOPreservesCallerGlobals(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "mpg.xxl")
	modSrc := `import crypto from "crypto"
export func hashIt(s) {
    return crypto.sha256(s)
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	// The caller also imports crypto, so caller's globals[0] = crypto Module.
	// After TCO of hashIt (which switches globals to module's), the caller
	// must still be able to use its own crypto import.
	callerSrc := `import crypto from "crypto"
import { hashIt } from "` + modPath + `"
func caller(s) {
    var h = hashIt(s)
    var d = crypto.sha256(s)
    return h + "|" + d
}
return caller("test")
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !strings.Contains(out, want) {
		t.Errorf("expected hash %q in output, got %q", want, out)
	}
}

// Extra: multiple sequential cross-module TCO calls.
// Ensures that frame reuse doesn't leak state between calls.
func TestModuleMethod_MultipleCrossModuleTCO(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "mm.xxl")
	modSrc := `import crypto from "crypto"
export func hashIt(s) {
    return crypto.sha256(s)
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	callerSrc := `import { hashIt } from "` + modPath + `"
func caller(s) {
    return hashIt(s)
}
var r1 = caller("test")
var r2 = caller("hello")
return r1 + "|" + r2
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want1 := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	want2 := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if !strings.Contains(out, want1) || !strings.Contains(out, want2) {
		t.Errorf("expected both hashes in output, got %q", out)
	}
}

// Extra: cross-module function that calls another module method (not just crypto).
// Tests that the module's globals are correctly set up for multiple imports.
func TestModuleMethod_CrossModuleMultipleImports(t *testing.T) {
	dir := t.TempDir()
	modFile := filepath.Join(dir, "mmi.xxl")
	modSrc := `import crypto from "crypto"
import strings from "strings"
export func process(s) {
    var h = crypto.sha256(s)
    return strings.toUpper(h)
}
`
	if err := os.WriteFile(modFile, []byte(modSrc), 0644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	modPath := filepath.ToSlash(modFile)

	callerSrc := `import { process } from "` + modPath + `"
return process("test")
`
	out, err := runRegScriptReturn(t, callerSrc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sha256("test") is all lowercase hex, toUpper should make it uppercase
	want := "9F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A08"
	if !strings.Contains(out, want) {
		t.Errorf("expected uppercased hash, got %q", out)
	}
}

// Extra: indirect builtin tail call.
// `var f = len; return f("hello")` — f is a GlobalScope variable holding a
// Builtin. compileTailCall emits OpRegTailCall (not the BuiltinScope fast-path),
// so handleRegTailCall must correctly return the result.
func TestModuleMethod_IndirectBuiltinTailCall(t *testing.T) {
	src := `var f = len
func g() {
    return f("hello")
}
return g()
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected len('hello')=5, got %q", out)
	}
}

// Extra: top-level indirect builtin tail call.
func TestModuleMethod_TopLevelIndirectBuiltinTailCall(t *testing.T) {
	src := `var f = len
return f("hello")
`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected len('hello')=5, got %q", out)
	}
}
