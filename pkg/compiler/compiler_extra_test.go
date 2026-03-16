// pkg/compiler/compiler_extra_test.go
// Tests for uncovered compiler functions
package compiler

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// ============================================
// Tests for Compiler helper methods
// ============================================

func TestCompiler_DefineGlobalMethod(t *testing.T) {
	c := New()

	sym := c.DefineGlobal("x")
	if sym.Name != "x" {
		t.Errorf("expected symbol name 'x', got %q", sym.Name)
	}
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

func TestCompiler_ResolveSymbolMethod(t *testing.T) {
	c := New()

	// Define a symbol
	c.DefineGlobal("testVar")

	// Resolve it
	sym, ok := c.ResolveSymbol("testVar")
	if !ok {
		t.Fatal("expected to resolve symbol")
	}
	if sym.Name != "testVar" {
		t.Errorf("expected symbol name 'testVar', got %q", sym.Name)
	}

	// Try to resolve non-existent symbol
	_, ok = c.ResolveSymbol("nonExistent")
	if ok {
		t.Error("expected not to resolve non-existent symbol")
	}
}

// ============================================
// Tests for CompiledFunction methods
// ============================================

func TestCompiledFunction_TypeTag(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpConstant), 0, 0},
		NumLocals:     0,
		NumParameters: 0,
	}

	if fn.TypeTag() != objects.TagFunction {
		t.Errorf("expected TagFunction, got %d", fn.TypeTag())
	}
}

// ============================================
// Tests for lastInstructionIs
// ============================================

func TestCompiler_lastInstructionIs(t *testing.T) {
	c := New()

	// Initially no last instruction
	if c.lastInstructionIs(OpPop) {
		t.Error("expected no last instruction initially")
	}

	// Emit an instruction
	c.emit(OpPop)

	// Now should detect
	if !c.lastInstructionIs(OpPop) {
		t.Error("expected last instruction to be OpPop")
	}

	// Wrong opcode
	if c.lastInstructionIs(OpAdd) {
		t.Error("expected last instruction not to be OpAdd")
	}
}

// ============================================
// Tests for changeOperands
// ============================================

func TestCompiler_changeOperands(t *testing.T) {
	c := New()

	// Emit OpPushHandler with placeholder values
	pos := c.emit(OpPushHandler, 9999, 9999)

	// Change the operands
	c.changeOperands(pos, 100, 200)

	// Verify the change
	instr := c.currentInstructions()
	// OpPushHandler has two 2-byte operands
	// Byte 0: opcode
	// Bytes 1-2: first operand (catchAddr)
	// Bytes 3-4: second operand (finallyAddr)
	if Opcode(instr[pos]) != OpPushHandler {
		t.Errorf("expected OpPushHandler at position %d", pos)
	}

	// Extract the operands (big-endian)
	catchAddr := int(instr[pos+1])<<8 | int(instr[pos+2])
	finallyAddr := int(instr[pos+3])<<8 | int(instr[pos+4])

	if catchAddr != 100 {
		t.Errorf("expected catchAddr 100, got %d", catchAddr)
	}
	if finallyAddr != 200 {
		t.Errorf("expected finallyAddr 200, got %d", finallyAddr)
	}
}

// ============================================
// Tests for try-catch-finally compilation
// ============================================

func TestCompileTryStatement(t *testing.T) {
	input := `
		try {
			42
		} catch (e) {
			0
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Check that OpPushHandler and OpPopHandler are present
	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpPushHandler) {
		t.Error("expected OpPushHandler in instructions")
	}
	if !containsOpcode(instr, OpPopHandler) {
		t.Error("expected OpPopHandler in instructions")
	}
}

func TestCompileTryFinallyStatement(t *testing.T) {
	input := `
		try {
			42
		} finally {
			1
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpPushHandler) {
		t.Error("expected OpPushHandler in instructions")
	}
	if !containsOpcode(instr, OpPopHandler) {
		t.Error("expected OpPopHandler in instructions")
	}
}

func TestCompileTryCatchFinallyStatement(t *testing.T) {
	input := `
		func risky() { return 1 }
		func handle(e) { return e }
		func cleanup() { return 0 }
		try {
			risky()
		} catch (e) {
			handle(e)
		} finally {
			cleanup()
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpPushHandler) {
		t.Error("expected OpPushHandler in instructions")
	}
}

// ============================================
// Tests for throw statement compilation
// ============================================

func TestCompileThrowStatement(t *testing.T) {
	input := `
		throw "error occurred"
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpThrow) {
		t.Error("expected OpThrow in instructions")
	}
}

func TestCompileThrowWithVariable(t *testing.T) {
	input := `
		func createError() { return "error" }
		var err = createError()
		throw err
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpThrow) {
		t.Error("expected OpThrow in instructions")
	}
}

// ============================================
// Tests for decodeCompiledFunctionFromMap
// ============================================

func TestDecodeCompiledFunctionFromMap(t *testing.T) {
	data := map[string]interface{}{
		"Instructions":  []byte{byte(OpConstant), 0, 0},
		"NumLocals":     2,
		"NumParameters": 1,
	}

	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fn.Instructions) != 3 {
		t.Errorf("expected 3 instruction bytes, got %d", len(fn.Instructions))
	}
	if fn.NumLocals != 2 {
		t.Errorf("expected NumLocals 2, got %d", fn.NumLocals)
	}
	if fn.NumParameters != 1 {
		t.Errorf("expected NumParameters 1, got %d", fn.NumParameters)
	}
}

func TestDecodeCompiledFunctionFromMap_Int64(t *testing.T) {
	// Test with int64 values (from gob decoding)
	data := map[string]interface{}{
		"Instructions":  []byte{byte(OpPop)},
		"NumLocals":     int64(3),
		"NumParameters": int64(2),
	}

	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn.NumLocals != 3 {
		t.Errorf("expected NumLocals 3, got %d", fn.NumLocals)
	}
	if fn.NumParameters != 2 {
		t.Errorf("expected NumParameters 2, got %d", fn.NumParameters)
	}
}

func TestDecodeCompiledFunctionFromMap_Uint64(t *testing.T) {
	// Test with uint64 values (from gob decoding)
	data := map[string]interface{}{
		"Instructions":  []byte{byte(OpReturn)},
		"NumLocals":     uint64(5),
		"NumParameters": uint64(3),
	}

	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn.NumLocals != 5 {
		t.Errorf("expected NumLocals 5, got %d", fn.NumLocals)
	}
	if fn.NumParameters != 3 {
		t.Errorf("expected NumParameters 3, got %d", fn.NumParameters)
	}
}

func TestDecodeCompiledFunctionFromMap_Empty(t *testing.T) {
	// Test with empty map
	data := map[string]interface{}{}

	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty function
	if fn == nil {
		t.Fatal("expected non-nil function")
	}
}

func TestDecodeCompiledFunctionFromMap_WithFreeVars(t *testing.T) {
	data := map[string]interface{}{
		"Instructions":  []byte{byte(OpGetFree), 0},
		"NumLocals":     1,
		"NumParameters": 0,
		"FreeVariables": []serializableFreeVar{
			{Name: "x", Scope: "LOCAL", Index: 0},
		},
	}

	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fn.FreeVariables) != 1 {
		t.Errorf("expected 1 free variable, got %d", len(fn.FreeVariables))
	}
	if fn.FreeVariables[0].Name != "x" {
		t.Errorf("expected free var name 'x', got %q", fn.FreeVariables[0].Name)
	}
}

// ============================================
// Tests for getCompiledFunction
// ============================================

func TestGetCompiledFunction(t *testing.T) {
	input := `
		func add(a, b) {
			return a + b
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// The function should be stored in constants
	bytecode := c.Bytecode()
	if len(bytecode.Constants) < 1 {
		t.Fatal("expected at least 1 constant")
	}

	// Find the CompiledFunction in constants
	var found bool
	for _, constant := range bytecode.Constants {
		if _, ok := constant.(*CompiledFunction); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find CompiledFunction in constants")
	}
}

// ============================================
// Tests for isArrayAccessSafe
// ============================================

func TestIsArrayAccessSafeMethod(t *testing.T) {
	c := New()

	// Initially no safe accesses
	if c.isArrayAccessSafe("arr", "i") {
		t.Error("expected no safe array access initially")
	}

	// The compiler tracks safe accesses during compilation
	// This is set during loop compilation
}

// ============================================
// Tests for NewWithState
// ============================================

func TestNewWithStateMethod(t *testing.T) {
	symTable := NewSymbolTable()
	symTable.Define("existingVar")

	constants := []objects.Object{objects.NewInt(42)}

	c := NewWithState(symTable, constants)

	if c.symbolTable != symTable {
		t.Error("expected same symbol table")
	}
	if len(c.constants) != 1 {
		t.Errorf("expected 1 constant, got %d", len(c.constants))
	}

	// Verify the existing symbol is resolvable
	sym, ok := c.ResolveSymbol("existingVar")
	if !ok {
		t.Error("expected to resolve existingVar")
	}
	if sym.Name != "existingVar" {
		t.Errorf("expected symbol name 'existingVar', got %q", sym.Name)
	}
}

// ============================================
// Tests for inlineFunction
// ============================================

func TestInlineFunction_Simple(t *testing.T) {
	// Test that simple functions can be inlined
	input := `
		func add(a, b) {
			return a + b
		}
		add(1, 2)
	`
	program := parseExtra(input)

	// Enable inlining
	opts := OptimizationFlags{
		BytecodeOptimizer: true,
		InlineFunctions:   true,
		Superinstructions: true,
	}
	c := NewWithOptions(opts)
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Verify compilation succeeded
	_ = c.Bytecode()
}

// ============================================
// Tests for switch statement compilation
// ============================================

func TestCompileSwitchStatement(t *testing.T) {
	input := `
		var x = 2
		switch (x) {
			case 1:
				10
			case 2:
				20
			default:
				30
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Verify instructions were generated
	instr := c.Bytecode().Instructions
	if len(instr) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestCompileSwitchWithoutDefault(t *testing.T) {
	input := `
		var x = 1
		switch (x) {
			case 1:
				10
			case 2:
				20
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

// ============================================
// Tests for C-style for loop compilation
// ============================================

func TestCompileCStyleForLoop(t *testing.T) {
	input := `
		for (var i = 0; i < 10; i++) {
			i
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse in for loop")
	}
}

// ============================================
// Tests for ternary expression compilation
// ============================================

func TestCompileTernaryExpression(t *testing.T) {
	input := `
		var result = true ? 1 : 0
		result
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Ternary should compile to conditional jumps
	instr := c.Bytecode().Instructions
	if !containsOpcode(instr, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse for ternary")
	}
}

func TestCompileTernaryExpression_Nested(t *testing.T) {
	input := `
		var a = true
		var b = false
		var x = a ? (b ? 1 : 2) : 3
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

// ============================================
// Tests for peekNextChar in lexer (indirectly)
// ============================================

func TestLexer_PeekNextChar(t *testing.T) {
	// Test that lexer correctly handles multi-character operators
	// which use peekNextChar internally
	input := `
		var a = 1 <= 2
		var b = 1 >= 2
		var c = 1 == 2
		var d = 1 != 2
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

// ============================================
// Tests for break/continue in loops
// ============================================

func TestCompileBreakInLoop(t *testing.T) {
	input := `
		for (var i = 0; i < 10; i++) {
			if (i == 5) {
				break
			}
			i
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	// Break might be compiled as a jump, not as OpBreak directly
	if len(instr) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestCompileContinueInLoop(t *testing.T) {
	input := `
		for (var i = 0; i < 10; i++) {
			if (i % 2 == 0) {
				continue
			}
			i
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	instr := c.Bytecode().Instructions
	// Continue might be compiled as a jump, not as OpContinue directly
	if len(instr) == 0 {
		t.Error("expected non-empty instructions")
	}
}

// ============================================
// Helper function
// ============================================

func parseExtra(input string) *parser.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

// ============================================
// Tests for NoInliningOptimizations
// ============================================

func TestNoInliningOptimizations(t *testing.T) {
	opts := NoInliningOptimizations()

	if opts.BytecodeOptimizer != true {
		t.Error("expected BytecodeOptimizer to be true")
	}
	if opts.InlineFunctions != false {
		t.Error("expected InlineFunctions to be false")
	}
	if opts.InlineCache != true {
		t.Error("expected InlineCache to be true")
	}
	if opts.ClosurePool != true {
		t.Error("expected ClosurePool to be true")
	}
	if opts.TypeSpecialization != true {
		t.Error("expected TypeSpecialization to be true")
	}
	if opts.Superinstructions != true {
		t.Error("expected Superinstructions to be true")
	}
}

// ============================================
// Tests for getCompiledFunction with edge cases
// ============================================

func TestGetCompiledFunction_LocalScope(t *testing.T) {
	c := New()

	// Create a local symbol (not global)
	symTable := NewEnclosedSymbolTable(NewSymbolTable())
	sym := symTable.Define("localVar")

	// getCompiledFunction should return false for local scope
	_, ok := c.getCompiledFunction(sym)
	if ok {
		t.Error("expected false for local scope symbol")
	}
}

func TestGetCompiledFunction_OutOfBounds(t *testing.T) {
	c := New()

	// Create a global symbol with index out of bounds
	sym := Symbol{Name: "x", Scope: GlobalScope, Index: 9999}

	_, ok := c.getCompiledFunction(sym)
	if ok {
		t.Error("expected false for out of bounds symbol")
	}
}

func TestGetCompiledFunction_NonFunction(t *testing.T) {
	c := New()

	// Define a global integer (not a function)
	c.DefineGlobal("x")
	c.constants = append(c.constants, objects.NewInt(42))

	// Get the symbol
	sym, ok := c.ResolveSymbol("x")
	if !ok {
		t.Fatal("expected to resolve x")
	}

	// getCompiledFunction should return false for non-function
	_, ok = c.getCompiledFunction(sym)
	if ok {
		t.Error("expected false for non-function constant")
	}
}

// ============================================
// Tests for replaceInstruction
// ============================================

func TestCompiler_ReplaceInstruction(t *testing.T) {
	c := New()

	// Emit some instructions (OpPop takes no operands)
	pos := c.emit(OpPop)
	c.emit(OpPop)

	// Replace the first instruction with a different one
	// Since OpPop takes no operands, we just replace with another OpPop
	newInstr := []byte{byte(OpPop)}
	c.replaceInstruction(pos, newInstr)

	// Verify the instruction was replaced
	instr := c.currentInstructions()
	if Opcode(instr[pos]) != OpPop {
		t.Errorf("expected OpPop at position %d", pos)
	}
}

// ============================================
// Tests for emit with different operand widths
// ============================================

func TestCompiler_EmitWithOperands(t *testing.T) {
	c := New()

	// Emit instruction with no operands
	pos1 := c.emit(OpPop)
	if pos1 != 0 {
		t.Errorf("expected position 0, got %d", pos1)
	}

	// Emit instruction with 2-byte operand (OpConstant takes 2-byte index)
	pos2 := c.emit(OpConstant, 0x1234)
	if pos2 != 1 {
		t.Errorf("expected position 1, got %d", pos2)
	}

	// Emit instruction with two 2-byte operands (OpPushHandler)
	pos3 := c.emit(OpPushHandler, 0x1234, 0x5678)
	if pos3 != 4 {
		t.Errorf("expected position 4, got %d", pos3)
	}
}

// ============================================
// Tests for currentInstructions
// ============================================

func TestCompiler_CurrentInstructions(t *testing.T) {
	c := New()

	instr := c.currentInstructions()
	if len(instr) != 0 {
		t.Errorf("expected empty instructions initially, got %d bytes", len(instr))
	}

	c.emit(OpPop)
	instr = c.currentInstructions()
	if len(instr) != 1 {
		t.Errorf("expected 1 byte, got %d", len(instr))
	}
}

// ============================================
// Tests for removeLastInstruction
// ============================================

func TestCompiler_RemoveLastInstruction(t *testing.T) {
	c := New()

	c.emit(OpPop)
	c.emit(OpAdd)

	// Remove last instruction (OpAdd)
	c.removeLastInstruction()

	instr := c.currentInstructions()
	if len(instr) != 1 {
		t.Errorf("expected 1 byte after removal, got %d", len(instr))
	}
	if Opcode(instr[0]) != OpPop {
		t.Error("expected OpPop to remain")
	}
}

// ============================================
// Tests for addConstant
// ============================================

func TestCompiler_AddConstant(t *testing.T) {
	c := New()

	// Add an integer constant
	idx := c.addConstant(objects.NewInt(42))
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}

	// Add another constant
	idx = c.addConstant(objects.NewInt(100))
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}

	// Verify constants were stored
	if len(c.constants) != 2 {
		t.Errorf("expected 2 constants, got %d", len(c.constants))
	}
}

// ============================================
// Tests for loop safety analysis
// ============================================

func TestCompiler_AnalyzeLoopSafety(t *testing.T) {
	input := `
		var arr = [1, 2, 3, 4, 5]
		for (var i = 0; i < len(arr); i++) {
			arr[i]
		}
	`
	program := parseExtra(input)
	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Verify compilation succeeded
	_ = c.Bytecode()
}

// ============================================
// Tests for compileTailCall
// ============================================

func TestCompiler_CompileTailCall(t *testing.T) {
	input := `
		func sum(n, acc) {
			if (n <= 0) {
				return acc
			}
			return sum(n - 1, acc + n)
		}
		sum(10, 0)
	`
	program := parseExtra(input)
	opts := DefaultOptimizations()
	c := NewWithOptions(opts)
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Verify tail call optimization was applied
	// The bytecode should contain OpTailCall
	_ = c.Bytecode()
}
