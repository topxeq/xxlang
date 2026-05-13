// pkg/compiler/compiler_test.go
package compiler

import (
	"bytes"
	"io"
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
		input        string
		expectedOps  []Opcode
		numConstants int
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
		// Disable optimizations to test raw bytecode generation
		compiler := NewWithOptions(OptimizationFlags{BytecodeOptimizer: false})
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
		input      string
		expectedOp Opcode
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
		input      string
		expectedOp Opcode
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
		input   string
		varName string
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
		input     string
		numParams int
		hasReturn bool
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
	// Use a function with side effects (pr) so it won't be inlined
	input := "func f(x) { pr(x); return x; } f(42);"

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
		input   string
		builtin string
	}{
		{`len("hello");`, "len"},
		{`pr("hello");`, "pr"},
		{`pln("hello");`, "pln"},
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

		symbol := global.DefineBuiltin("customBuiltin")
		if symbol.Name != "customBuiltin" {
			t.Errorf("expected name 'customBuiltin', got=%s", symbol.Name)
		}
		if symbol.Scope != BuiltinScope {
			t.Errorf("expected scope BuiltinScope, got=%s", symbol.Scope)
		}
		if symbol.BuiltinName != "customBuiltin" {
			t.Errorf("expected BuiltinName 'customBuiltin', got=%s", symbol.BuiltinName)
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

// ============================================
// Import Statement Tests
// ============================================

func TestImportStatement_Simple(t *testing.T) {
	tests := []struct {
		input       string
		expectedOps []Opcode
	}{
		{
			`import "./math";`,
			[]Opcode{OpLoadModule, OpPop},
		},
		{
			`import "../utils";`,
			[]Opcode{OpLoadModule, OpPop},
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		for _, op := range tt.expectedOps {
			if !containsOpcode(bytecode.Instructions, op) {
				t.Errorf("expected opcode %v not found in instructions for %q", op, tt.input)
			}
		}
	}
}

func TestImportStatement_Default(t *testing.T) {
	input := `import math from "./math";`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpLoadModule and OpSetGlobal
	if !containsOpcode(bytecode.Instructions, OpLoadModule) {
		t.Error("expected OpLoadModule in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Check that 'math' is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("math")
	if !ok {
		t.Error("variable 'math' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'math', got %v", symbol.Scope)
	}
}

func TestImportStatement_Namespace(t *testing.T) {
	input := `import * as math from "./math";`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpLoadModule and OpSetGlobal for the namespace alias
	if !containsOpcode(bytecode.Instructions, OpLoadModule) {
		t.Error("expected OpLoadModule in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Check that 'math' namespace alias is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("math")
	if !ok {
		t.Error("variable 'math' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'math', got %v", symbol.Scope)
	}
}

func TestImportStatement_Destructuring(t *testing.T) {
	input := `import { add, sub } from "./math";`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpLoadModule, OpGetExport for each name, and OpSetGlobal for each
	if !containsOpcode(bytecode.Instructions, OpLoadModule) {
		t.Error("expected OpLoadModule in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpGetExport) {
		t.Error("expected OpGetExport in instructions")
	}

	// Should have at least 2 OpSetGlobal for 'add' and 'sub'
	count := countOpcode(bytecode.Instructions, OpSetGlobal)
	if count < 2 {
		t.Errorf("expected at least 2 OpSetGlobal, got %d", count)
	}

	// Check that 'add' and 'sub' are defined in symbol table
	for _, name := range []string{"add", "sub"} {
		symbol, ok := compiler.symbolTable.Resolve(name)
		if !ok {
			t.Errorf("variable %q not defined in symbol table", name)
		}
		if symbol.Scope != GlobalScope {
			t.Errorf("expected GlobalScope for %q, got %v", name, symbol.Scope)
		}
	}
}

func TestImportStatement_ModulePath(t *testing.T) {
	tests := []struct {
		input string
		path  string
	}{
		{`import "./math";`, "./math"},
		{`import "../utils";`, "../utils"},
		{`import "module";`, "module"},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		// Find the module path in constants
		found := false
		for _, c := range bytecode.Constants {
			if str, ok := c.(*objects.String); ok {
				if str.Value == tt.path {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("expected path %q in constants for %q", tt.path, tt.input)
		}
	}
}

// ============================================
// Export Statement Tests
// ============================================

func TestExportStatement_Var(t *testing.T) {
	input := `export var x = 10;`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpSetGlobal (for defining the variable)
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Should have OpSetExport (for exporting the variable)
	if !containsOpcode(bytecode.Instructions, OpSetExport) {
		t.Error("expected OpSetExport in instructions")
	}

	// Check that 'x' is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("x")
	if !ok {
		t.Error("variable 'x' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'x', got %v", symbol.Scope)
	}
}

func TestExportStatement_Const(t *testing.T) {
	input := `export const PI = 3.14;`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpSetGlobal (for defining the constant)
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Should have OpSetExport (for exporting the constant)
	if !containsOpcode(bytecode.Instructions, OpSetExport) {
		t.Error("expected OpSetExport in instructions")
	}

	// Check that 'PI' is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("PI")
	if !ok {
		t.Error("variable 'PI' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'PI', got %v", symbol.Scope)
	}
}

func TestExportStatement_Func(t *testing.T) {
	input := `export func add(a, b) { return a + b; }`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have OpSetGlobal (for defining the function)
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Should have OpSetExport (for exporting the function)
	if !containsOpcode(bytecode.Instructions, OpSetExport) {
		t.Error("expected OpSetExport in instructions")
	}

	// Should have OpClosure for the function
	if !containsOpcode(bytecode.Instructions, OpClosure) {
		t.Error("expected OpClosure in instructions")
	}

	// Check that 'add' is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("add")
	if !ok {
		t.Error("variable 'add' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'add', got %v", symbol.Scope)
	}
}

func TestExportStatement_Multiple(t *testing.T) {
	input := `
export var x = 10;
export const PI = 3.14;
export func add(a, b) { return a + b; }
`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := compiler.Bytecode()

	// Should have 3 OpSetExport for the three exports
	count := countOpcode(bytecode.Instructions, OpSetExport)
	if count != 3 {
		t.Errorf("expected 3 OpSetExport, got %d", count)
	}

	// Check that all variables are defined in symbol table
	for _, name := range []string{"x", "PI", "add"} {
		symbol, ok := compiler.symbolTable.Resolve(name)
		if !ok {
			t.Errorf("variable %q not defined in symbol table", name)
		}
		if symbol.Scope != GlobalScope {
			t.Errorf("expected GlobalScope for %q, got %v", name, symbol.Scope)
		}
	}
}

// ============================================
// Class Compilation Tests
// ============================================

func TestClassStatementCompilation(t *testing.T) {
	input := `
		class Person {
			var name = ""
			func init(name) { this.name = name }
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()
	if len(bytecode.Instructions) == 0 {
		t.Error("expected instructions to be generated")
	}

	// Should have OpClass opcode
	if !containsOpcode(bytecode.Instructions, OpClass) {
		t.Error("expected OpClass in instructions")
	}

	// Should have OpSetGlobal for storing the class
	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	// Check that 'Person' is defined in symbol table
	symbol, ok := compiler.symbolTable.Resolve("Person")
	if !ok {
		t.Error("class 'Person' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'Person', got %v", symbol.Scope)
	}
}

func TestNewExpressionCompilation(t *testing.T) {
	input := `
		class Person {}
		var p = new Person();
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()
	if !containsOpcode(bytecode.Instructions, OpNew) {
		t.Error("expected OpNew in instructions")
	}
}

func TestThisExpressionCompilation(t *testing.T) {
	input := `
		class Counter {
			var count = 0
			func increment() { this.count = this.count + 1 }
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Check that OpGetLocal and OpSetField are in method instructions
	// Methods are compiled functions stored in constants
	foundGetLocal := false
	foundSetField := false
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpGetLocal) {
				foundGetLocal = true
			}
			if containsOpcode(fn.Instructions, OpSetField) {
				foundSetField = true
			}
		}
	}

	if !foundGetLocal {
		t.Error("expected OpGetLocal in method instructions for 'this'")
	}
	if !foundSetField {
		t.Error("expected OpSetField in method instructions for field assignment")
	}
}

func TestSuperCallExpressionCompilation(t *testing.T) {
	input := `
		class Animal {
			func init(name) { }
		}
		class Dog extends Animal {
			func init(name) { super.init(name) }
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Check that OpSuper is in method instructions
	foundSuper := false
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpSuper) {
				foundSuper = true
				break
			}
		}
	}

	if !foundSuper {
		t.Error("expected OpSuper in method instructions")
	}
}

func TestClassWithSuperclassCompilation(t *testing.T) {
	input := `
		class Animal { var name = "" }
		class Dog extends Animal { var breed = "" }
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Should have 2 OpClass (one for each class)
	count := countOpcode(bytecode.Instructions, OpClass)
	if count != 2 {
		t.Errorf("expected 2 OpClass, got %d", count)
	}

	// Check that both classes are in symbol table
	for _, name := range []string{"Animal", "Dog"} {
		symbol, ok := compiler.symbolTable.Resolve(name)
		if !ok {
			t.Errorf("class %q not defined in symbol table", name)
		}
		if symbol.Scope != GlobalScope {
			t.Errorf("expected GlobalScope for %q, got %v", name, symbol.Scope)
		}
	}
}

// ============================================
// Error Handling Tests
// ============================================

func TestCompiler_UndefinedVariable(t *testing.T) {
	input := `x`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Fatal("expected error for undefined variable, got nil")
	}
}

func TestCompiler_OpcodeVerification(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectedOp Opcode
	}{
		{"constant load", "42;", OpConstant},
		{"addition", "1 + 2;", OpAdd},
		{"subtraction", "2 - 1;", OpSub},
		{"multiplication", "2 * 3;", OpMul},
		{"division", "6 / 2;", OpDiv},
		{"modulo", "5 % 2;", OpMod},
		{"less than", "1 < 2;", OpLess},
		{"greater than", "2 > 1;", OpGreater},
		{"equal", "1 == 1;", OpEqual},
		{"not equal", "1 != 2;", OpNotEqual},
		{"logical and", "true && false;", OpAnd},
		{"logical or", "true || false;", OpOr},
		{"logical not", "!true;", OpNot},
		{"array creation", "[1, 2, 3];", OpArray},
		{"map creation", `{"a": 1};`, OpMap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)
			// Disable optimizations to test raw bytecode generation
			compiler := NewWithOptions(OptimizationFlags{BytecodeOptimizer: false})
			err := compiler.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %v", err)
			}

			bytecode := compiler.Bytecode()
			if !containsOpcode(bytecode.Instructions, tt.expectedOp) {
				t.Errorf("expected opcode %v in bytecode", tt.expectedOp)
			}
		})
	}
}

// ============================================
// Error Case Tests for Export Statement
// ============================================

func TestExportStatement_UnnamedFunctionError(t *testing.T) {
	// Test: export func without name should return error
	input := `export func(a, b) { return a + b; }`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Error("expected error for unnamed exported function, got nil")
	}
}

// ============================================
// Method with No Return Tests (for lastInstructionIs coverage)
// ============================================

func TestMethodWithoutExplicitReturn(t *testing.T) {
	// Test: method without explicit return should still end with OpReturn
	// This exercises the lastInstructionIs check in compileMethod
	input := `
		class Counter {
			var count = 0
			func increment() {
				this.count = this.count + 1
			}
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Check that OpReturn is in method instructions
	foundReturn := false
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpReturn) {
				foundReturn = true
				break
			}
		}
	}

	if !foundReturn {
		t.Error("expected OpReturn in method instructions")
	}
}

func TestMethodWithParameters(t *testing.T) {
	// Test: method with multiple parameters
	// This exercises parameter compilation in compileMethod
	input := `
		class Calculator {
			func add(a, b, c) {
				return a + b + c
			}
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Check that method is compiled with 3 parameters
	foundMethod := false
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if fn.NumParameters == 3 {
				foundMethod = true
				break
			}
		}
	}

	if !foundMethod {
		t.Error("expected method with 3 parameters")
	}
}

// ============================================
// Function Return Statement Tests (for lastInstructionIs coverage)
// ============================================

func TestFunctionWithoutExplicitReturn(t *testing.T) {
	// Test: function without explicit return should still end with OpReturn
	// This exercises the lastInstructionIs check in compileFunction
	input := `func f() { }`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Check that OpReturn is in function instructions
	foundReturn := false
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpReturn) {
				foundReturn = true
				break
			}
		}
	}

	if !foundReturn {
		t.Error("expected OpReturn in a function instructions")
	}
}

// ============================================
// Serialization Tests
// ============================================

func TestSerializeRoundTrip(t *testing.T) {
	// Create a simple bytecode
	compiler := New()
	program := parse(`var x = 42; func add(a, b) { return a + b; }`)
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	original := compiler.Bytecode()

	// Serialize
	serialized, err := original.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// Deserialize
	deserialized, err := Deserialize(serialized)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	// Compare instructions
	if !bytes.Equal(original.Instructions, deserialized.Instructions) {
		t.Error("instructions don't match after round-trip")
	}

	// Compare constant counts
	if len(original.Constants) != len(deserialized.Constants) {
		t.Error("constant counts don't match")
	}
}

func TestSerializeToFile(t *testing.T) {
	compiler := New()
	program := parse(`var x = 42;`)
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Use t.TempDir() for a unique temp directory
	tmpFile := t.TempDir() + "/test_bytecode.xxl"

	// Serialize to file
	err = bytecode.SerializeToFile(tmpFile)
	if err != nil {
		t.Fatalf("serialize to file error: %v", err)
	}

	// Deserialize from file
	deserialized, err := DeserializeFromFile(tmpFile)
	if err != nil {
		t.Fatalf("deserialize from file error: %v", err)
	}

	// Verify constants match
	if len(bytecode.Constants) != len(deserialized.Constants) {
		t.Error("constants don't match after file round-trip")
	}
}

func TestSerializeWithEmptyBytecode(t *testing.T) {
	// Test serialization of empty bytecode
	empty := &Bytecode{}

	serialized, err := empty.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	deserialized, err := Deserialize(serialized)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	if len(deserialized.Instructions) != 0 {
		t.Error("expected empty instructions")
	}
}

func TestSerializeWithCompiledFunction(t *testing.T) {
	// Create a CompiledFunction with some instructions
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpConstant), 0, byte(OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
		FreeVariables: []Symbol{},
	}

	// Create bytecode with compiled function in constants
	bytecode := &Bytecode{
		Instructions: []byte{byte(OpConstant), 0, byte(OpCall), 0, byte(OpPop)},
		Constants:    []objects.Object{fn},
	}

	serialized, err := bytecode.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	deserialized, err := Deserialize(serialized)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	// Verify we got a CompiledFunction back
	if len(deserialized.Constants) != 1 {
		t.Fatal("expected 1 constant")
	}

	deserFn, ok := deserialized.Constants[0].(*CompiledFunction)
	if !ok {
		t.Fatal("expected CompiledFunction")
	}

	if !bytes.Equal(fn.Instructions, deserFn.Instructions) {
		t.Error("instructions don't match")
	}
}

func TestSerializeWithSourceMap(t *testing.T) {
	compiler := New()
	program := parse(`var x = 42;`)
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Create a source map
	sourceMap := NewSourceMap()
	sourceMap.Add(0, SourceLocation{Line: 1, Column: 1})
	sourceMap.SourceFile = "test.xxl"

	// Serialize bytecode
	serialized, err := bytecode.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// SourceMap should serialize independently
	var smBuf bytes.Buffer
	err = sourceMap.Serialize(&smBuf)
	if err != nil {
		t.Fatalf("source map serialize error: %v", err)
	}

	smDeser, err := DeserializeSourceMap(bytes.NewReader(smBuf.Bytes()))
	if err != nil {
		t.Fatalf("source map deserialize error: %v", err)
	}

	if smDeser.SourceFile != "test.xxl" {
		t.Error("file name doesn't match")
	}

	// Clean up
	_, _ = Deserialize(serialized)
	_, _ = DeserializeSourceMap(bytes.NewReader(smBuf.Bytes()))
	_ = io.Reader(nil) // Use io to avoid import warning
}
