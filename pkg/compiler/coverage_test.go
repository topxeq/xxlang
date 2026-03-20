// pkg/compiler/coverage_test.go
// Additional tests to improve coverage
package compiler

import (
	"bytes"
	"os"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// CompiledFunction Method Tests
// ============================================

func TestCompiledFunction_Type(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
	}

	if fn.Type() != objects.FunctionType {
		t.Errorf("expected FunctionType, got %v", fn.Type())
	}
}

func TestCompiledFunction_Inspect(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
	}

	inspect := fn.Inspect()
	if inspect == "" {
		t.Error("expected non-empty inspect string")
	}
}

func TestCompiledFunction_ToBool(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
	}

	if fn.ToBool() != objects.TRUE {
		t.Error("expected CompiledFunction to be truthy")
	}
}

func TestCompiledFunction_HashKey(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
	}

	hk := fn.HashKey()
	if hk.Type != objects.FunctionType {
		t.Errorf("expected FunctionType in HashKey, got %v", hk.Type)
	}
}

// ============================================
// Closure Compilation Tests
// ============================================

func TestClosureCompilation(t *testing.T) {
	input := `
		func outer() {
			var x = 10
			func inner() {
				return x
			}
			return inner
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpClosure) {
		t.Error("expected OpClosure in instructions")
	}

	funcCount := 0
	for _, c := range bytecode.Constants {
		if _, ok := c.(*CompiledFunction); ok {
			funcCount++
		}
	}
	if funcCount < 2 {
		t.Errorf("expected at least 2 compiled functions, got %d", funcCount)
	}
}

func TestClosureWithFreeVariables(t *testing.T) {
	input := `
		func makeCounter() {
			var count = 0
			return func() {
				count = count + 1
				return count
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

	if !containsOpcode(bytecode.Instructions, OpClosure) {
		t.Error("expected OpClosure in instructions")
	}

	var innerFunc *CompiledFunction
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if len(fn.FreeVariables) > 0 {
				innerFunc = fn
				break
			}
		}
	}

	if innerFunc == nil {
		t.Error("expected to find a function with free variables")
	}
}

// ============================================
// Method Call Tests
// ============================================

func TestMethodCallCompilation(t *testing.T) {
	input := `
		class Person {
			func getName() { return "John" }
		}
		var p = new Person();
		p.getName();
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpCallMethod) {
		t.Error("expected OpCallMethod in instructions")
	}
}

// ============================================
// Ternary Expression Tests
// ============================================

// Note: Ternary expressions require parser support.
// The compiler has TernaryExpression handling but parser support may vary.

// ============================================
// Compound Assignment Tests
// ============================================

func TestCompoundAssignmentCompilation(t *testing.T) {
	tests := []struct {
		input      string
		expectedOp Opcode
		altOps     []Opcode // Alternative opcodes from optimization
	}{
		{"var x = 5; x += 1;", OpAdd, []Opcode{OpAddLocalConst, OpGetGlobalConstAdd}},
		{"var x = 5; x -= 1;", OpSub, []Opcode{OpSubLocalConst, OpGetGlobalConstSub}},
		{"var x = 5; x *= 2;", OpMul, []Opcode{OpMulLocalConst, OpGetGlobalConstMul}},
		{"var x = 10; x /= 2;", OpDiv, nil},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %v", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		// Check for expected opcode or any of its optimized alternatives
		found := containsOpcode(bytecode.Instructions, tt.expectedOp)
		if !found && tt.altOps != nil {
			for _, altOp := range tt.altOps {
				if containsOpcode(bytecode.Instructions, altOp) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("expected %v (or optimized version) in instructions for %q", tt.expectedOp, tt.input)
		}
	}
}

// ============================================
// Postfix Expression Tests
// ============================================

func TestPostfixExpressionCompilation(t *testing.T) {
	tests := []struct {
		input      string
		expectedOp Opcode
		altOps     []Opcode // Alternative opcodes from optimization
	}{
		{"var x = 5; x++;", OpAdd, []Opcode{OpIncLocal, OpGetGlobalConstAdd}},
		{"var x = 5; x--;", OpSub, []Opcode{OpDecLocal, OpGetGlobalConstSub}},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %v", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		// Check for expected opcode or any of its optimized alternatives
		found := containsOpcode(bytecode.Instructions, tt.expectedOp)
		if !found && tt.altOps != nil {
			for _, altOp := range tt.altOps {
				if containsOpcode(bytecode.Instructions, altOp) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("expected %v (or optimized version) in instructions for %q", tt.expectedOp, tt.input)
		}
		// Note: OpDup may or may not be present depending on optimization
	}
}

// ============================================
// For Loop Tests
// ============================================

func TestForLoopCompilation(t *testing.T) {
	input := "for (var i = 0; i < 10; i = i + 1) { i; }"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse in instructions")
	}
	if !containsOpcode(bytecode.Instructions, OpJump) {
		t.Error("expected OpJump in instructions")
	}
}

// ============================================
// Break and Continue Tests
// ============================================

func TestBreakStatement(t *testing.T) {
	input := "while (true) { break; }"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Break is now implemented as OpJump
	if !containsOpcode(bytecode.Instructions, OpJump) {
		t.Error("expected OpJump in instructions (break is now a jump)")
	}
}

func TestContinueStatement(t *testing.T) {
	input := "while (true) { continue; }"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	// Continue is now implemented as OpJump
	if !containsOpcode(bytecode.Instructions, OpJump) {
		t.Error("expected OpJump in instructions (continue is now a jump)")
	}
}

// ============================================
// Tail Call Tests
// ============================================

func TestTailCallCompilation(t *testing.T) {
	input := `
		func factorial(n, acc) {
			if (n <= 1) {
				return acc
			}
			return factorial(n - 1, n * acc)
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	var foundTailCall bool
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpTailCall) {
				foundTailCall = true
				break
			}
		}
	}

	if !foundTailCall {
		t.Error("expected OpTailCall in function instructions")
	}
}

// ============================================
// Const Declaration Tests
// ============================================

func TestConstDeclaration(t *testing.T) {
	input := "const PI = 3.14;"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}

	symbol, ok := compiler.symbolTable.Resolve("PI")
	if !ok {
		t.Error("variable 'PI' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'PI', got %v", symbol.Scope)
	}
}

// ============================================
// Local Variable Tests
// ============================================

func TestLocalVariablesInFunction(t *testing.T) {
	input := `
		func test() {
			var a = 1
			var b = 2
			return a + b
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	var fn *CompiledFunction
	for _, c := range bytecode.Constants {
		if f, ok := c.(*CompiledFunction); ok {
			fn = f
			break
		}
	}

	if fn == nil {
		t.Fatal("expected to find a compiled function")
	}

	if !containsOpcode(fn.Instructions, OpSetLocal) {
		t.Error("expected OpSetLocal in function instructions")
	}

	// After optimization, OpGetLocal may be combined into OpGetLocalAdd superinstruction
	// So check for either OpGetLocal or OpGetLocalAdd
	hasLocalAccess := containsOpcode(fn.Instructions, OpGetLocal) || containsOpcode(fn.Instructions, OpGetLocalAdd)
	if !hasLocalAccess {
		t.Error("expected OpGetLocal or OpGetLocalAdd in function instructions")
	}
}

// ============================================
// Named Function Tests
// ============================================

func TestNamedFunctionRecursion(t *testing.T) {
	input := `
		func fib(n) {
			if (n <= 1) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	symbol, ok := compiler.symbolTable.Resolve("fib")
	if !ok {
		t.Error("function 'fib' not defined in symbol table")
	}
	if symbol.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for 'fib', got %v", symbol.Scope)
	}

	var foundRecursiveCall bool
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if containsOpcode(fn.Instructions, OpGetGlobal) {
				foundRecursiveCall = true
				break
			}
		}
	}

	if !foundRecursiveCall {
		t.Error("expected OpGetGlobal in function for recursive call")
	}
}

// ============================================
// Map Literal Tests
// ============================================

func TestMapLiteral(t *testing.T) {
	tests := []struct {
		input    string
		numPairs int
	}{
		{`{};`, 0},
		{`{"a": 1};`, 1},
		{`{"a": 1, "b": 2, "c": 3};`, 3},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %q: %v", tt.input, err)
		}

		bytecode := compiler.Bytecode()

		if !containsOpcode(bytecode.Instructions, OpMap) {
			t.Errorf("expected OpMap in instructions for %q", tt.input)
		}
	}
}

// ============================================
// Index Assignment Tests
// ============================================

func TestIndexAssignment(t *testing.T) {
	input := "var arr = [1, 2, 3]; arr[0] = 10;"

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()

	if !containsOpcode(bytecode.Instructions, OpSetIndex) {
		t.Error("expected OpSetIndex in instructions")
	}
}

// ============================================
// Error Cases Tests
// ============================================

func TestCompilerError_UndefinedInAssignment(t *testing.T) {
	input := "x = 10;"
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Error("expected error for assignment to undefined variable")
	}
}

func TestCompilerError_UnknownPrefixOperator(t *testing.T) {
	input := "~5;"
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Error("expected error for unknown prefix operator")
	}
}

func TestCompilerError_UndefinedSuperclass(t *testing.T) {
	input := `class Dog extends UndefinedClass {}`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Error("expected error for undefined superclass")
	}
}

func TestCompilerError_UndefinedClassInNew(t *testing.T) {
	input := `var p = new UndefinedClass();`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err == nil {
		t.Error("expected error for undefined class in new expression")
	}
}

// ============================================
// Source Map Tests
// ============================================

func TestSourceMap_Basic(t *testing.T) {
	sm := NewSourceMap()

	loc := SourceLocation{Line: 5, Column: 10}
	sm.Add(100, loc)

	retrieved, ok := sm.Get(100)
	if !ok {
		t.Fatal("expected to retrieve source location")
	}
	if retrieved.Line != 5 || retrieved.Column != 10 {
		t.Errorf("expected line 5, column 10, got line %d, column %d", retrieved.Line, retrieved.Column)
	}
}

func TestSourceMap_ClosestMatch(t *testing.T) {
	sm := NewSourceMap()
	sm.Add(0, SourceLocation{Line: 1, Column: 1})
	sm.Add(10, SourceLocation{Line: 2, Column: 5})
	sm.Add(20, SourceLocation{Line: 3, Column: 1})

	loc, ok := sm.Get(15)
	if !ok {
		t.Fatal("expected to retrieve source location")
	}
	if loc.Line != 2 || loc.Column != 5 {
		t.Errorf("expected line 2, column 5, got line %d, column %d", loc.Line, loc.Column)
	}
}

func TestSourceMap_SetSourceFile(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("test.xxl", "line1\nline2\nline3")

	if sm.SourceFile != "test.xxl" {
		t.Errorf("expected source file 'test.xxl', got %q", sm.SourceFile)
	}
	if len(sm.SourceLines) != 3 {
		t.Errorf("expected 3 source lines, got %d", len(sm.SourceLines))
	}
}

func TestSourceMap_GetLine(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("test.xxl", "first line\nsecond line\nthird line")

	tests := []struct {
		lineNum  int
		expected string
	}{
		{1, "first line"},
		{2, "second line"},
		{3, "third line"},
		{0, ""},
		{4, ""},
	}

	for _, tt := range tests {
		result := sm.GetLine(tt.lineNum)
		if result != tt.expected {
			t.Errorf("GetLine(%d) = %q, want %q", tt.lineNum, result, tt.expected)
		}
	}
}

func TestSourceMap_FormatError(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("test.xxl", "var x = 1")
	sm.Add(0, SourceLocation{Line: 1, Column: 5})

	result := sm.FormatError(0, "test error")

	if result == "" {
		t.Error("expected non-empty formatted error")
	}
	if !containsStr(result, "test.xxl") {
		t.Error("expected source file in formatted error")
	}
	if !containsStr(result, "test error") {
		t.Error("expected error message in formatted error")
	}
}

// ============================================
// Bytecode Serialization Tests
// ============================================

func TestBytecodeSerialization(t *testing.T) {
	input := `
		var x = 42
		var y = "hello"
		var arr = [1, 2, 3]
		var m = {"key": "value"}
	`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	original := compiler.Bytecode()

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}

	if len(data) < 8 {
		t.Fatalf("serialized data too short: %d bytes", len(data))
	}
	if string(data[:4]) != MagicHeader {
		t.Errorf("wrong magic header: %q", string(data[:4]))
	}

	restored, err := Deserialize(data)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}

	if !bytes.Equal(original.Instructions, restored.Instructions) {
		t.Error("instructions mismatch after serialization/deserialization")
	}

	if len(original.Constants) != len(restored.Constants) {
		t.Errorf("constants count mismatch: original %d, restored %d", len(original.Constants), len(restored.Constants))
	}
}

func TestBytecodeSerialization_Functions(t *testing.T) {
	input := `
		func add(a, b) {
			return a + b
		}
	`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	original := compiler.Bytecode()

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}

	restored, err := Deserialize(data)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}

	var foundFunc bool
	for _, c := range restored.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			if fn.NumParameters == 2 {
				foundFunc = true
				break
			}
		}
	}

	if !foundFunc {
		t.Error("expected to find compiled function with 2 parameters")
	}
}

func TestBytecodeSerialization_InvalidMagic(t *testing.T) {
	data := []byte("XXXX1")
	_, err := Deserialize(data)
	if err == nil {
		t.Error("expected error for invalid magic header")
	}
}

func TestBytecodeSerialization_InvalidVersion(t *testing.T) {
	data := []byte(MagicHeader)
	data = append(data, 0xFF, 0xFF, 0xFF, 0xFF)
	_, err := Deserialize(data)
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

// ============================================
// Bytecode File Serialization Tests
// ============================================

func TestBytecodeSerializeToFile(t *testing.T) {
	input := `var x = 42;`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	original := compiler.Bytecode()

	tmpFile := "/tmp/test_bytecode.xxbc"
	err = original.SerializeToFile(tmpFile)
	if err != nil {
		t.Fatalf("serialization to file error: %v", err)
	}

	restored, err := DeserializeFromFile(tmpFile)
	if err != nil {
		t.Fatalf("deserialization from file error: %v", err)
	}

	if !bytes.Equal(original.Instructions, restored.Instructions) {
		t.Error("instructions mismatch after file serialization")
	}

	os.Remove(tmpFile)
}

// ============================================
// Source Map Serialization Tests
// ============================================

func TestSourceMapSerialization(t *testing.T) {
	original := NewSourceMap()
	original.SetSourceFile("test.xxl", "line1\nline2\nline3")
	original.Add(0, SourceLocation{Line: 1, Column: 1})
	original.Add(10, SourceLocation{Line: 2, Column: 5})

	var buf bytes.Buffer
	err := original.Serialize(&buf)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}

	restored, err := DeserializeSourceMap(&buf)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}

	if restored.SourceFile != original.SourceFile {
		t.Errorf("source file mismatch: want %q, got %q", original.SourceFile, restored.SourceFile)
	}

	if len(restored.SourceLines) != len(original.SourceLines) {
		t.Errorf("source lines count mismatch: want %d, got %d", len(original.SourceLines), len(restored.SourceLines))
	}

	if len(restored.Locations) != len(original.Locations) {
		t.Errorf("locations count mismatch: want %d, got %d", len(original.Locations), len(restored.Locations))
	}
}

// ============================================
// Free Variable Scope Tests
// ============================================

func TestFreeVariableScopes(t *testing.T) {
	global := NewSymbolTable()
	global.Define("globalVar")

	local1 := NewEnclosedSymbolTable(global)
	local1.Define("localVar1")

	local2 := NewEnclosedSymbolTable(local1)
	local2.Define("localVar2")

	g, ok := local2.Resolve("globalVar")
	if !ok {
		t.Fatal("expected to resolve globalVar")
	}
	if g.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for globalVar, got %v", g.Scope)
	}

	l1, ok := local2.Resolve("localVar1")
	if !ok {
		t.Fatal("expected to resolve localVar1")
	}
	if l1.Scope != FreeScope {
		t.Errorf("expected FreeScope for localVar1, got %v", l1.Scope)
	}

	l2, ok := local2.Resolve("localVar2")
	if !ok {
		t.Fatal("expected to resolve localVar2")
	}
	if l2.Scope != LocalScope {
		t.Errorf("expected LocalScope for localVar2, got %v", l2.Scope)
	}
}

// ============================================
// SetSource Tests
// ============================================

func TestCompiler_SetSource(t *testing.T) {
	compiler := New()
	compiler.SetSource("test.xxl", "var x = 1")

	if compiler.sourceFile != "test.xxl" {
		t.Errorf("expected source file 'test.xxl', got %q", compiler.sourceFile)
	}
	if compiler.sourceCode != "var x = 1" {
		t.Errorf("expected source code 'var x = 1', got %q", compiler.sourceCode)
	}
}

// ============================================
// Object Serialization Tests
// ============================================

func TestObjectSerialization_Int(t *testing.T) {
	original := &objects.Int{Value: 42}
	serial, err := objectToSerializable(original)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	restored, err := serializableToObject(serial)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if restored.(*objects.Int).Value != original.Value {
		t.Errorf("value mismatch: want %d, got %d", original.Value, restored.(*objects.Int).Value)
	}
}

func TestObjectSerialization_Float(t *testing.T) {
	original := &objects.Float{Value: 3.14}
	serial, err := objectToSerializable(original)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	restored, err := serializableToObject(serial)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if restored.(*objects.Float).Value != original.Value {
		t.Errorf("value mismatch: want %f, got %f", original.Value, restored.(*objects.Float).Value)
	}
}

func TestObjectSerialization_String(t *testing.T) {
	original := &objects.String{Value: "hello"}
	serial, err := objectToSerializable(original)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	restored, err := serializableToObject(serial)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if restored.(*objects.String).Value != original.Value {
		t.Errorf("value mismatch: want %q, got %q", original.Value, restored.(*objects.String).Value)
	}
}

func TestObjectSerialization_Array(t *testing.T) {
	original := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
	}}
	serial, err := objectToSerializable(original)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	restored, err := serializableToObject(serial)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if len(restored.(*objects.Array).Elements) != len(original.Elements) {
		t.Errorf("elements count mismatch: want %d, got %d", len(original.Elements), len(restored.(*objects.Array).Elements))
	}
}

func TestObjectSerialization_Map(t *testing.T) {
	original := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		(&objects.String{Value: "key"}).HashKey(): {Key: &objects.String{Value: "key"}, Value: &objects.Int{Value: 1}},
	}}
	serial, err := objectToSerializable(original)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	restored, err := serializableToObject(serial)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if len(restored.(*objects.Map).Pairs) != len(original.Pairs) {
		t.Errorf("pairs count mismatch: want %d, got %d", len(original.Pairs), len(restored.(*objects.Map).Pairs))
	}
}

// ============================================
// Helper Functions
// ============================================

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
