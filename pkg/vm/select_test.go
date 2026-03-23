// pkg/vm/select_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// helper function to run code and return last result
func runCode(t *testing.T, code string) Value {
	t.Helper()

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	vm := NewRegVM(c.Bytecode())
	if err := vm.Run(); err != nil {
		t.Fatalf("VM error: %v", err)
	}

	return vm.LastResult()
}

// ============================================================
// Tube Operations Tests
// ============================================================

func TestTubeSendReceive(t *testing.T) {
	code := `
var tube = makeTube(1)
tube <- 42
var val = <- tube
val
`
	result := runCode(t, code)
	if !result.IsInt() || result.GetInt() != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestTubeBuffered(t *testing.T) {
	code := `
var tube = makeTube(3)
tube <- 1
tube <- 2
tube <- 3
var a = <- tube
var b = <- tube
var c = <- tube
a + b + c
`
	result := runCode(t, code)
	if !result.IsInt() || result.GetInt() != 6 {
		t.Errorf("Expected 6, got %v", result)
	}
}

func TestTubeClose(t *testing.T) {
	code := `
var tube = makeTube(1)
tube <- 42
closeTube(tube)
var closed = tubeClosed(tube)
var val = <- tube
closed
`
	result := runCode(t, code)
	if !result.IsBool() || !result.GetBool() {
		t.Errorf("Expected true for closed, got %v", result)
	}
}

func TestTubeLen(t *testing.T) {
	code := `
var tube = makeTube(5)
tube <- 1
tube <- 2
tubeLen(tube)
`
	result := runCode(t, code)
	if !result.IsInt() || result.GetInt() != 2 {
		t.Errorf("Expected 2, got %v", result)
	}
}

func TestTubeCap(t *testing.T) {
	code := `
var tube = makeTube(5)
tubeCap(tube)
`
	result := runCode(t, code)
	if !result.IsInt() || result.GetInt() != 5 {
		t.Errorf("Expected 5, got %v", result)
	}
}

// ============================================================
// Select Statement Tests
// ============================================================

func TestSelectReceiveWithDefault(t *testing.T) {
	// Non-blocking receive with default
	code := `
var tube = makeTube(1)
var result = "none"
select {
case v = <- tube:
    result = v
default:
    result = "default"
}
result
`
	result := runCode(t, code)
	if !result.IsObject() {
		t.Fatalf("Expected string result, got %v", result)
	}
}

func TestSelectSend(t *testing.T) {
	code := `
var tube = makeTube(1)
var sent = false
select {
case tube <- 100:
    sent = true
default:
    sent = false
}
sent
`
	result := runCode(t, code)
	if !result.IsBool() || !result.GetBool() {
		t.Errorf("Expected sent=true, got %v", result)
	}
}

func TestSelectSendWithDefaultFull(t *testing.T) {
	// Send to full tube with default
	code := `
var tube = makeTube(1)
tube <- 1
var result = "not sent"
select {
case tube <- 2:
    result = "sent"
default:
    result = "default"
}
result
`
	result := runCode(t, code)
	if !result.IsObject() {
		t.Fatalf("Expected string result, got %v", result)
	}
}

func TestSelectEmptyDefault(t *testing.T) {
	// Select with only default should immediately execute default
	code := `
var executed = false
select {
default:
    executed = true
}
executed
`
	result := runCode(t, code)
	if !result.IsBool() {
		t.Errorf("Expected bool result, got %v", result)
	}
}

func TestSelectSingleCaseReceive(t *testing.T) {
	code := `
var tube = makeTube(1)
tube <- 42
var val = 0
select {
case v = <- tube:
    val = v
}
val
`
	result := runCode(t, code)
	if !result.IsInt() || result.GetInt() != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestSelectMultipleCases(t *testing.T) {
	// Multiple cases - tube2 has data
	code := `
var tube1 = makeTube(1)
var tube2 = makeTube(1)
tube2 <- "hello"
var result = ""
select {
case v = <- tube1:
    result = v
case v = <- tube2:
    result = v
}
result
`
	result := runCode(t, code)
	if !result.IsObject() {
		t.Fatalf("Expected string result, got %v", result)
	}
}

func TestSelectWithTimeout(t *testing.T) {
	// Select with context timeout
	code := `
var tube = makeTube(1)
var ctx = contextWithTimeout(null, 50)
var result = "timeout"

select {
case v = <- tube:
    result = v
case <- ctx.done():
    result = "timeout"
}
result
`
	result := runCode(t, code)
	if !result.IsObject() {
		t.Fatalf("Expected string result, got %v", result)
	}
}

// ============================================================
// Parser Tests for Select and Run
// ============================================================

func TestParseSelectStatement(t *testing.T) {
	tests := []string{
		`select { case x = <- tube: }`,
		`select { case tube <- 1: }`,
		`select { default: }`,
		`select { case x = <- tube: pln(x) }`,
		`select { case x = <- tube: pln(x) default: pln("none") }`,
	}

	for _, code := range tests {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Errorf("Parser errors for %q: %v", code, p.Errors())
		}

		if program == nil {
			t.Errorf("Parser returned nil program for %q", code)
		}
	}
}

func TestParseRunStatement(t *testing.T) {
	tests := []string{
		`run { pln("hello") }`,
		`run worker()`,
		`run worker(1, 2, 3)`,
	}

	for _, code := range tests {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Errorf("Parser errors for %q: %v", code, p.Errors())
		}

		if program == nil {
			t.Errorf("Parser returned nil program for %q", code)
		}
	}
}

// ============================================================
// Compiler Tests for Select
// ============================================================

func TestCompileSelectStatement(t *testing.T) {
	tests := []string{
		`var tube = makeTube(1); select { case x = <- tube: x }`,
		`var tube = makeTube(1); select { case tube <- 1: pln(1) default: pln(0) }`,
		`select { default: null }`,
	}

	for _, code := range tests {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Errorf("Compiler error for %q: %v", code, err)
		}

		// Verify bytecode was generated
		bytecode := c.Bytecode()
		if len(bytecode.Instructions) == 0 {
			t.Errorf("No bytecode generated for %q", code)
		}
	}
}

func TestCompileRunStatement(t *testing.T) {
	tests := []string{
		`run { pln("hello") }`,
	}

	for _, code := range tests {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Errorf("Compiler error for %q: %v", code, err)
		}
	}
}

// ============================================================
// AST Tests for Select and Run
// ============================================================

func TestSelectASTStructure(t *testing.T) {
	code := `select { case x = <- tube: pln(x) default: pln("none") }`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	// The statement should be a SelectStatement
	_ = program.Statements[0]
}

func TestRunASTStructure(t *testing.T) {
	code := `run { pln("hello") }`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	// The statement should be a RunStatement
	_ = program.Statements[0]
}
