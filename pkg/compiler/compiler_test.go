// pkg/compiler/compiler_test.go
package compiler

import (
	"bytes"
	"testing"

	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// ============================================
// Helper Functions
// ============================================

func parse(input string) *parser.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

// ============================================
// Basic Compiler Tests
// ============================================

func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1;", 1},
		{"42;", 42},
		{"999;", 999},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if len(bytecode.Constants) != 1 {
			t.Fatalf("expected 1 constant, got %d", len(bytecode.Constants))
		}

		intObj, ok := bytecode.Constants[0].(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", bytecode.Constants[0])
		}

		if intObj.Value != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, intObj.Value)
		}

		// Check that OpConstant and OpPop are in instructions
		if !containsOpcode(bytecode.Instructions, OpConstant) {
			t.Error("expected OpConstant in instructions")
		}
		if !containsOpcode(bytecode.Instructions, OpPop) {
			t.Error("expected OpPop in instructions")
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14;", 3.14},
		{"2.5;", 2.5},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if len(bytecode.Constants) != 1 {
			t.Fatalf("expected 1 constant, got %d", len(bytecode.Constants))
		}

		floatObj, ok := bytecode.Constants[0].(*objects.Float)
		if !ok {
			t.Fatalf("expected Float, got %T", bytecode.Constants[0])
		}

		if floatObj.Value != tt.expected {
			t.Errorf("expected %f, got %f", tt.expected, floatObj.Value)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello";`, "hello"},
		{`"";`, ""},
		{`"test string";`, "test string"},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if len(bytecode.Constants) != 1 {
			t.Fatalf("expected 1 constant, got %d", len(bytecode.Constants))
		}

		strObj, ok := bytecode.Constants[0].(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", bytecode.Constants[0])
		}

		if strObj.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, strObj.Value)
		}
	}
}

func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		// true and false should use OpTrue/OpFalse, not constants
		expectedOp := OpTrue
		if !tt.expected {
			expectedOp = OpFalse
		}

		if !containsOpcode(bytecode.Instructions, expectedOp) {
			t.Errorf("expected %v in instructions", expectedOp)
		}
	}
}

func TestNullLiteral(t *testing.T) {
	program := parse("null;")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpNull) {
		t.Error("expected OpNull in instructions")
	}
}

func TestArithmeticExpressions(t *testing.T) {
	tests := []struct {
		input          string
		expectedOps    []Opcode
		numConstants   int
	}{
		{"1 + 2;", []Opcode{OpConstant, OpConstant, OpAdd, OpPop}, 2},
		{"5 - 3;", []Opcode{OpConstant, OpConstant, OpSub, OpPop}, 2},
		{"4 * 2;", []Opcode{OpConstant, OpConstant, OpMul, OpPop}, 2},
		{"10 / 2;", []Opcode{OpConstant, OpConstant, OpDiv, OpPop}, 2},
		{"10 % 3;", []Opcode{OpConstant, OpConstant, OpMod, OpPop}, 2},
		{"-5;", []Opcode{OpConstant, OpNeg, OpPop}, 1},
		{"1 + 2 * 3;", []Opcode{OpConstant, OpConstant, OpConstant, OpMul, OpAdd, OpPop}, 3},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if len(bytecode.Constants) != tt.numConstants {
			t.Errorf("wrong number of constants for %q: want %d, got %d", tt.input, tt.numConstants, len(bytecode.Constants))
		}

		for _, op := range tt.expectedOps {
			if !containsOpcode(bytecode.Instructions, op) {
				t.Errorf("expected opcode %v not found in instructions for %q", op, tt.input)
			}
		}
	}
}

func TestComparisonExpressions(t *testing.T) {
	tests := []struct {
		input       string
		expectedOp  Opcode
	}{
		{"1 < 2;", OpLess},
		{"2 > 1;", OpGreater},
		{"1 <= 2;", OpLessEqual},
		{"2 >= 1;", OpGreaterEqual},
		{"1 == 2;", OpEqual},
		{"1 != 2;", OpNotEqual},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, tt.expectedOp) {
			t.Errorf("expected %v in instructions for %q", tt.expectedOp, tt.input)
		}
	}
}

func TestLogicalExpressions(t *testing.T) {
	tests := []struct {
		input       string
		expectedOp  Opcode
	}{
		{"true && false;", OpAnd},
		{"true || false;", OpOr},
		{"!true;", OpNot},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, tt.expectedOp) {
			t.Errorf("expected %v in instructions for %q", tt.expectedOp, tt.input)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	tests := []struct {
		input       string
		numElements int
	}{
		{"[];", 0},
		{"[1];", 1},
		{"[1, 2, 3];", 3},
		{"[[1, 2], [3]];", 3}, // nested arrays
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpArray) {
			t.Errorf("expected OpArray in instructions for %q", tt.input)
		}
	}
}

func TestVariableDeclarations(t *testing.T) {
	tests := []struct {
		input      string
		varName    string
	}{
		{"var x = 5;", "x"},
		{"var y = 1 + 2;", "y"},
		{"var abc = true;", "abc"},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
			t.Errorf("expected OpSetGlobal in instructions for %q", tt.input)
		}

		// Check that variable is defined in symbol table
		symbol, ok := compiler.symbolTable.Resolve(tt.varName)
		if !ok {
			t.Errorf("variable %q not defined in symbol table", tt.varName)
		}
		if symbol.Scope != GlobalScope {
			t.Errorf("expected GlobalScope for %q, got %v", tt.varName, symbol.Scope)
		}
	}
}

func TestVariableAccess(t *testing.T) {
	input := "var x = 5; x;"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpSetGlobal and OpGetGlobal
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpGetGlobal) {
		t.Error("expected OpGetGlobal in instructions")
	}
}

func TestIfStatement(t *testing.T) {
	tests := []string{
		"if (true) { 1; }",
		"if (true) { 1; } else { 2; }",
		"if (1 < 2) { 10; }",
	}

	for _, input := range tests {
		program := parse(input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpJumpIfFalse) {
			t.Errorf("expected OpJumpIfFalse in instructions for %q", input)
		}
	}
}

func TestWhileLoop(t *testing.T) {
	input := "while (true) { 1; }"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpJump) {
		t.Error("expected OpJump in instructions (for loop back)")
	}
}

func TestForInLoop(t *testing.T) {
	input := "for (x in [1, 2]) { x; }"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpJump) {
		t.Error("expected OpJump in instructions (for loop back)")
	}
}

func TestFunctionDefinition(t *testing.T) {
	tests := []struct {
		input         string
		numParams     int
		hasReturn     bool
	}{
		{"func f() { }", 0, false},
		{"func f(x) { return x; }", 1, true},
		{"func add(a, b) { return a + b; }", 2, true},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		// Should have a CompiledFunction in constants
		var foundFunc *CompiledFunction
		for _, c := range bytecode.Constants {
			if fn, ok := c.(*CompiledFunction); ok {
				foundFunc = fn
				break
			}
		}
		if foundFunc == nil {
			t.Errorf("expected CompiledFunction in constants for %q", tt.input)
			continue
		}

		// Check that function has correct number of parameters
		if foundFunc.NumParameters != tt.numParams {
			t.Errorf("expected %d parameters, got %d for %q", tt.numParams, foundFunc.NumParameters, tt.input)
		}

		// Check that function with return has OpReturn in its instructions
		if tt.hasReturn && !containsOpcode(foundFunc.Instructions, OpReturn) {
			t.Errorf("expected OpReturn in function instructions for %q", tt.input)
		}
	}
}

func TestFunctionCall(t *testing.T) {
	input := "func f() { return 1; } f();"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpCall) {
		t.Error("expected OpCall in instructions")
	}
}

func TestIndexExpression(t *testing.T) {
	tests := []string{
		"[1, 2, 3][1];",
		"var arr = [1, 2]; arr[0];",
	}

	for _, input := range tests {
		program := parse(input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpIndex) {
			t.Errorf("expected OpIndex in instructions for %q", input)
		}
	}
}

func TestAssignmentExpression(t *testing.T) {
	tests := []string{
		"var x = 5; x = 10;",
		"var arr = [1, 2]; arr[0] = 10;",
	}

	for _, input := range tests {
		program := parse(input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", input, err)
		}

		bytecode := compiler.Bytecode()

		// Should have at least 2 OpSetGlobal (one for declaration, one for assignment)
		count := countOpcode(bytecode.Instructions, OpSetGlobal)
		if count < 1 {
			t.Errorf("expected at least 1 OpSetGlobal in instructions for %q, got %d", input, count)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		builtin  string
	}{
		{`len("hello");`, "len"},
		{`print("hello");`, "print"},
		{`println("hello");`, "println"},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpBuiltin) {
			t.Errorf("expected OpBuiltin in instructions for %q", tt.input)
		}

		// Check builtin is defined
		symbol, ok := compiler.symbolTable.Resolve(tt.builtin)
		if !ok {
			t.Errorf("builtin %q not defined in symbol table", tt.builtin)
		}
		if symbol.Scope != BuiltinScope {
			t.Errorf("expected BuiltinScope for %q, got %v", tt.builtin, symbol.Scope)
		}
	}
}

// ============================================
// Symbol Table Tests
// ============================================

func TestSymbolTable(t *testing.T) {
	t.Run("define and resolve global", func(t *testing.T) {
		global := NewSymbolTable()

		symbol := global.Define("x")
		if symbol.Name != "x" {
			t.Errorf("expected name 'x', got=%s", symbol.Name)
		}
		if symbol.Scope != GlobalScope {
			t.Errorf("expected scope GlobalScope, got=%s", symbol.Scope)
		}
		if symbol.Index != 0 {
			t.Errorf("expected index 0, got=%d", symbol.Index)
		}

		resolved, ok := global.Resolve("x")
		if !ok {
			t.Fatalf("name 'x' not resolvable")
		}
		if resolved != symbol {
			t.Errorf("expected resolved symbol to be the same")
		}
	})

	t.Run("define and resolve local", func(t *testing.T) {
		global := NewSymbolTable()
		global.Define("x")

		local := NewEnclosedSymbolTable(global)
		symbol := local.Define("y")

		if symbol.Scope != LocalScope {
			t.Errorf("expected scope LocalScope, got=%s", symbol.Scope)
		}

		resolved, ok := local.Resolve("y")
		if !ok {
			t.Fatalf("name 'y' not resolvable")
		}
		if resolved != symbol {
			t.Errorf("expected resolved symbol to be the same")
		}

		// Should still resolve x from outer scope
		resolvedX, okX := local.Resolve("x")
		if !okX {
			t.Fatalf("name 'x' not resolvable from outer scope")
		}
		if resolvedX.Scope != GlobalScope {
			t.Errorf("expected x to have GlobalScope, got=%s", resolvedX.Scope)
		}
	})

	t.Run("define builtin", func(t *testing.T) {
		global := NewSymbolTable()

		symbol := global.DefineBuiltin(0, "customBuiltin")
		if symbol.Name != "customBuiltin" {
			t.Errorf("expected name 'customBuiltin', got=%s", symbol.Name)
		}
		if symbol.Scope != BuiltinScope {
			t.Errorf("expected scope BuiltinScope, got=%s", symbol.Scope)
		}
		if symbol.Index != 0 {
			t.Errorf("expected index 0, got=%d", symbol.Index)
		}
	})

	t.Run("resolve undefined", func(t *testing.T) {
		global := NewSymbolTable()

		_, ok := global.Resolve("undefined")
		if ok {
			t.Errorf("expected 'undefined' to not be resolvable")
		}
	})

	t.Run("free symbols", func(t *testing.T) {
		global := NewSymbolTable()
		global.Define("x")
		global.Define("y")

		local1 := NewEnclosedSymbolTable(global)
		local1.Define("a")
		local1.Define("b")

		local2 := NewEnclosedSymbolTable(local1)
		local2.Define("c")

		// Resolve x (global) from local2
		x, ok := local2.Resolve("x")
		if !ok {
			t.Fatalf("expected to resolve 'x'")
		}
		if x.Scope != GlobalScope {
			t.Errorf("expected 'x' to be GlobalScope, got=%s", x.Scope)
		}

		// Resolve a (local1's local) from local2 - should become free
		a, ok := local2.Resolve("a")
		if !ok {
			t.Fatalf("expected to resolve 'a'")
		}
		if a.Scope != FreeScope {
			t.Errorf("expected 'a' to be FreeScope, got=%s", a.Scope)
		}
	})
}

func TestNewWithState(t *testing.T) {
	symbolTable := NewSymbolTable()
	constants := []objects.Object{&objects.Int{Value: 1}}

	compiler := NewWithState(symbolTable, constants)

	if compiler.symbolTable != symbolTable {
		t.Errorf("expected symbolTable to be set")
	}
	if len(compiler.constants) != 1 {
		t.Errorf("expected 1 constant, got=%d", len(compiler.constants))
	}
}

// ============================================
// Helper Functions
// ============================================

func containsOpcode(instructions []byte, op Opcode) bool {
	for i := 0; i < len(instructions); i++ {
		if Opcode(instructions[i]) == op {
			return true
		}
	}
	return false
}

func countOpcode(instructions []byte, op Opcode) int {
	count := 0
	for i := 0; i < len(instructions); i++ {
		if Opcode(instructions[i]) == op {
			count++
		}
	}
	return count
}

// CompiledFunction equality check for testing
func compiledFunctionEqual(a, b *CompiledFunction) bool {
	return bytes.Equal(a.Instructions, b.Instructions) &&
		a.NumLocals == b.NumLocals &&
		a.NumParameters == b.NumParameters
}
