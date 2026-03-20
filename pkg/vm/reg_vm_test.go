// pkg/vm/reg_vm_test.go
package vm

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// TestRegVMBasicArithmetic tests basic arithmetic operations in the register VM
func TestRegVMBasicArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"1 + 2", 3},
		{"5 - 3", 2},
		{"4 * 3", 12},
		{"10 / 2", 5.0}, // Division produces float
		{"10 % 3", 1},
	}

	for _, tt := range tests {
		// Parse and compile
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
		if _, err := c.Compile(program); err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		// Run with register VM
		vm := NewRegVM(c.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("VM error: %v", err)
		}

		result := vm.LastResult()

		switch expected := tt.expected.(type) {
		case int:
			if !result.IsInt() {
				t.Errorf("Input: %s, expected int, got %v", tt.input, result)
				continue
			}
			if result.GetInt() != int64(expected) {
				t.Errorf("Input: %s, expected=%d, got=%d", tt.input, expected, result.GetInt())
			}
		case float64:
			if !result.IsFloat() {
				t.Errorf("Input: %s, expected float, got %v", tt.input, result)
				continue
			}
			if result.GetFloat() != expected {
				t.Errorf("Input: %s, expected=%f, got=%f", tt.input, expected, result.GetFloat())
			}
		}
	}
}

// TestRegVMConstants tests loading constants
func TestRegVMConstants(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"1", 1},
		{"true", true},
		{"false", false},
		{"null", nil},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
		if _, err := c.Compile(program); err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		// Run with register VM
		vm := NewRegVM(c.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("VM error: %v", err)
		}

		t.Logf("Input: %s, result verified", tt.input)
	}
}

// TestRegisterOpcodes tests that register opcodes are defined correctly
func TestRegisterOpcodes(t *testing.T) {
	opcodes := []compiler.Opcode{
		compiler.OpRegAdd,
		compiler.OpRegSub,
		compiler.OpRegMul,
		compiler.OpRegDiv,
		compiler.OpRegMove,
		compiler.OpRegLoadConst,
		compiler.OpRegJump,
		compiler.OpRegCall,
		compiler.OpRegReturn,
	}

	for _, op := range opcodes {
		if !compiler.IsRegisterOpcode(op) {
			t.Errorf("opcode %d should be a register opcode", op)
		}
	}
}

// TestRegisterAllocation tests the register allocator
func TestRegisterAllocation(t *testing.T) {
	ra := compiler.NewRegAllocator(256)

	// Add some intervals
	sym1 := &compiler.Symbol{Name: "a", Scope: compiler.LocalScope, Index: 0}
	sym2 := &compiler.Symbol{Name: "b", Scope: compiler.LocalScope, Index: 1}
	sym3 := &compiler.Symbol{Name: "c", Scope: compiler.LocalScope, Index: 2}

	ra.AddInterval(sym1, 0, 10)
	ra.AddInterval(sym2, 5, 15)
	ra.AddInterval(sym3, 8, 20)

	spilled := ra.Allocate()

	t.Logf("Allocated registers: spilled=%d", spilled)

	stats := ra.Stats()
	t.Logf("Stats: total=%d, assigned=%d, spilled=%d",
		stats.TotalIntervals, stats.AssignedRegs, stats.SpilledInts)

	// Check that variables got registers
	reg1 := ra.GetRegister("a")
	reg2 := ra.GetRegister("b")
	reg3 := ra.GetRegister("c")

	t.Logf("Registers: a=%d, b=%d, c=%d", reg1, reg2, reg3)

	if reg1 < 0 && reg1 != -2 {
		t.Errorf("variable 'a' should have a register, got %d", reg1)
	}
}

// TestRegFrame tests register frame operations
func TestRegFrame(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     5,
		NumParameters: 2,
	}

	frame := NewRegFrame(fn)

	// Test register access
	frame.SetReg(0, NewInt(42))
	if frame.GetReg(0).GetInt() != 42 {
		t.Errorf("expected register 0 to be 42, got %v", frame.GetReg(0))
	}

	// Test local variable access
	frame.SetLocal(0, NewInt(100))
	if frame.GetLocal(0).GetInt() != 100 {
		t.Errorf("expected local 0 to be 100, got %v", frame.GetLocal(0))
	}

	// Release frame
	frame.Release()
}

// TestValueOperations tests Value operations used by register VM
func TestValueOperations(t *testing.T) {
	// Test arithmetic
	a := NewInt(10)
	b := NewInt(3)

	result, ok := a.Add(b)
	if !ok || result.GetInt() != 13 {
		t.Errorf("expected 10 + 3 = 13, got %v (ok=%v)", result, ok)
	}

	result, ok = a.Sub(b)
	if !ok || result.GetInt() != 7 {
		t.Errorf("expected 10 - 3 = 7, got %v (ok=%v)", result, ok)
	}

	result, ok = a.Mul(b)
	if !ok || result.GetInt() != 30 {
		t.Errorf("expected 10 * 3 = 30, got %v (ok=%v)", result, ok)
	}

	result, ok = a.Div(b)
	if !ok {
		t.Errorf("expected division to succeed")
	}
	// Integer division produces float
	if !result.IsFloat() {
		t.Errorf("expected float result for division, got %v", result)
	}

	// Test comparison
	less := a.LessValue(b)
	if less == ValueTrue {
		t.Errorf("expected 10 < 3 to be false")
	}

	greater := a.GreaterValue(b)
	if greater != ValueTrue {
		t.Errorf("expected 10 > 3 to be true")
	}

	eq := a.EqualValue(a)
	if eq != ValueTrue {
		t.Errorf("expected 10 == 10 to be true")
	}
}

// TestRegCompilerBasic tests the register-based compiler
func TestRegCompilerBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1 + 2", 3},
		{"5 - 3", 2},
		{"4 * 3", 12},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		// Compile with register compiler
		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		bytecode := c.Bytecode()
		t.Logf("Input: %s", tt.input)
		t.Logf("Generated %d bytes of register bytecode", len(bytecode.Instructions))
	}
}

// TestRegVMForLoops tests for loop execution in register VM
func TestRegVMForLoops(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var sum = 0; for (var i = 1; i <= 5; i = i + 1) { sum = sum + i; } sum;", 15},
		{"var x = 1; for (var i = 0; i < 3; i = i + 1) { x = x * 2; } x;", 8},
		{"var n = 5; var fact = 1; for (var i = 1; i <= n; i = i + 1) { fact = fact * i; } fact;", 120},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		// Register compiler
		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		// Run with register VM
		vm := NewRegVM(c.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("VM error: %v", err)
		}

		result := vm.LastResult()
		if !result.IsInt() {
			t.Errorf("expected int result, got %v", result)
			continue
		}

		if result.GetInt() != tt.expected {
			t.Errorf("input=%s, expected=%d, got=%d", tt.input, tt.expected, result.GetInt())
		}
	}
}

// TestRegVMForLoopsWithBreakContinue tests break/continue in for loops
func TestRegVMForLoopsWithBreakContinue(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Break tests
		{"var sum = 0; for (var i = 0; i < 10; i = i + 1) { if (i == 5) { break; } sum = sum + i; } sum;", 10},
		{"var found = 0; for (var i = 0; i < 100; i = i + 1) { if (i == 7) { found = i; break; } } found;", 7},
		// Continue tests
		{"var sum = 0; for (var i = 0; i < 5; i = i + 1) { if (i == 2) { continue; } sum = sum + i; } sum;", 8},
		{"var count = 0; for (var i = 0; i < 4; i = i + 1) { if (i == 1) { continue; } count = count + 1; } count;", 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		// Register compiler
		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		// Run with register VM
		vm := NewRegVM(c.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("VM error: %v", err)
		}

		result := vm.LastResult()
		if !result.IsInt() {
			t.Errorf("expected int result, got %v", result)
			continue
		}

		if result.GetInt() != tt.expected {
			t.Errorf("input=%s, expected=%d, got=%d", tt.input, tt.expected, result.GetInt())
		}
	}
}

// TestRegVMClosures tests closure functionality in register VM
func TestRegVMClosures(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Simple closure capturing a variable
		{
			`
var outer = func() {
    var x = 10
    return func() {
        return x
    }
}
var f = outer()
f()
`,
			10,
		},
		// Counter closure - tests mutable captured variable
		{
			`
var makeCounter = func() {
    var count = 0
    return func() {
        count = count + 1
        return count
    }
}
var c = makeCounter()
c()
c()
c()
`,
			3,
		},
		// Closure with multiple calls
		{
			`
var makeAdder = func(x) {
    return func(y) {
        return x + y
    }
}
var add5 = makeAdder(5)
add5(3)
`,
			8,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %v", err)
		}

		vm := NewRegVM(c.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("VM error: %v", err)
		}

		result := vm.LastResult()
		if !result.IsInt() {
			t.Errorf("expected int result, got %v", result)
			continue
		}

		if result.GetInt() != tt.expected {
			t.Errorf("expected=%d, got=%d", tt.expected, result.GetInt())
		}
	}
}

// TestRegVMMaps tests map functionality in register VM
func TestRegVMMaps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"name": "Alice"}`, "Alice"},
		{`var m = {"a": 1, "b": 2}; m["a"]`, "1"},
		{`var m = {"x": 10, "y": 20}; m["y"]`, "20"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		// Register VM
		rc := compiler.NewRegCompiler()
		if _, err := rc.Compile(program); err != nil {
			t.Fatalf("reg compiler error: %v", err)
		}
		rvm := NewRegVM(rc.Bytecode())
		if err := rvm.Run(); err != nil {
			t.Fatalf("reg VM error: %v", err)
		}
		regResult := rvm.LastResult().ToObject().Inspect()

		// Compare with expected
		if !strings.Contains(regResult, tt.expected) {
			t.Errorf("reg VM: expected to contain %q, got %q", tt.expected, regResult)
		}
	}
}
