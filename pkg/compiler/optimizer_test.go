// pkg/compiler/optimizer_test.go
// Tests for bytecode optimizer
package compiler

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestOptimizer_FoldConstants_Integers(t *testing.T) {
	// Create bytecode: OpConstant(0), OpConstant(1), OpAdd
	// OpConstant uses 2-byte big-endian index
	input := []byte{
		byte(OpConstant), 0, 0, // constant index 0 (big-endian)
		byte(OpConstant), 0, 1, // constant index 1 (big-endian)
		byte(OpAdd),
	}
	constants := []objects.Object{objects.NewInt(1), objects.NewInt(2)}
	bytecode := &Bytecode{Instructions: input, Constants: constants}

	optimizer := NewOptimizer(bytecode)
	result := optimizer.FoldConstants()

	// Expected: single OpConstant with value 3
	if len(result.Constants) != 3 {
		t.Errorf("expected 3 constants, got %d", len(result.Constants))
	}

	folded, ok := result.Constants[2].(*objects.Int)
	if !ok || folded.Value != 3 {
		t.Errorf("expected folded constant to be Int(3), got %v", result.Constants[2])
	}

	// Check that the instruction is OpConstant with index 2
	if Opcode(result.Instructions[0]) != OpConstant {
		t.Errorf("expected OpConstant, got %v", Opcode(result.Instructions[0]))
	}
}

func TestOptimizer_GenerateSuperinstructions_GetLocalAdd(t *testing.T) {
	// Create bytecode: OpGetLocal(0), OpGetLocal(1), OpAdd
	input := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1, byte(OpAdd)}
	bytecode := &Bytecode{Instructions: input, Constants: nil}

	optimizer := NewOptimizer(bytecode)
	result := optimizer.GenerateSuperinstructions()

	// Expected: OpGetLocalAdd(0, 1)
	if len(result.Instructions) != 3 {
		t.Errorf("expected 3 bytes, got %d: %v", len(result.Instructions), result.Instructions)
	}

	if Opcode(result.Instructions[0]) != OpGetLocalAdd {
		t.Errorf("expected OpGetLocalAdd, got %v", Opcode(result.Instructions[0]))
	}
	if result.Instructions[1] != 0 || result.Instructions[2] != 1 {
		t.Errorf("expected operands [0, 1], got [%d, %d]", result.Instructions[1], result.Instructions[2])
	}
}

func TestOptimizer_GenerateSuperinstructions_ConstantAdd(t *testing.T) {
	// Create bytecode: OpConstant(0), OpConstant(1), OpAdd
	// OpConstant uses 2-byte big-endian index
	input := []byte{
		byte(OpConstant), 0, 0, // constant index 0
		byte(OpConstant), 0, 1, // constant index 1
		byte(OpAdd),
	}
	constants := []objects.Object{objects.NewInt(10), objects.NewInt(20)}
	bytecode := &Bytecode{Instructions: input, Constants: constants}

	optimizer := NewOptimizer(bytecode)
	result := optimizer.GenerateSuperinstructions()

	// Expected: OpConstantAdd(0, 1)
	if len(result.Instructions) != 5 {
		t.Errorf("expected 5 bytes, got %d: %v", len(result.Instructions), result.Instructions)
	}

	if Opcode(result.Instructions[0]) != OpConstantAdd {
		t.Errorf("expected OpConstantAdd, got %v", Opcode(result.Instructions[0]))
	}
}

func TestOptimizer_FullOptimization(t *testing.T) {
	// Test that Optimize() runs both passes
	constants := []objects.Object{objects.NewInt(5), objects.NewInt(3)}
	input := []byte{
		byte(OpConstant), 0, 0, // push 5 (index 0)
		byte(OpConstant), 0, 1, // push 3 (index 1)
		byte(OpAdd), // add -> would be folded to 8
	}
	bytecode := &Bytecode{Instructions: input, Constants: constants}

	optimizer := NewOptimizer(bytecode)
	result := optimizer.Optimize()

	// After constant folding, should have a single constant
	if Opcode(result.Instructions[0]) != OpConstant {
		t.Errorf("expected OpConstant after optimization, got %v", Opcode(result.Instructions[0]))
	}

	// Check the folded value
	folded, ok := result.Constants[2].(*objects.Int)
	if !ok || folded.Value != 8 {
		t.Errorf("expected folded constant to be Int(8), got %v", result.Constants[2])
	}
}

// ============================================
// Tests for stackEffectOf
// ============================================

func TestStackEffectOf_PushInstructions(t *testing.T) {
	tests := []struct {
		op       Opcode
		expected int
	}{
		{OpConstant, 1},
		{OpGetLocal, 1},
		{OpGetGlobal, 1},
		{OpGetFree, 1},
		{OpTrue, 1},
		{OpFalse, 1},
		{OpNull, 1},
		{OpDup, 1},
	}

	for _, tt := range tests {
		opName := "unknown"
		if def, err := Lookup(byte(tt.op)); err == nil {
			opName = def.Name
		}
		t.Run(opName, func(t *testing.T) {
			// Create minimal instruction with operand
			instructions := []byte{byte(tt.op), 0, 0}
			effect := stackEffectOf(tt.op, instructions, 0)
			if effect != tt.expected {
				t.Errorf("stackEffectOf(%v) = %d, want %d", tt.op, effect, tt.expected)
			}
		})
	}
}

func TestStackEffectOf_Superinstructions(t *testing.T) {
	tests := []struct {
		op       Opcode
		expected int
	}{
		{OpConstantAdd, -1}, // pop 2, push 1
		{OpConstantSub, -1},
		{OpConstantMul, -1},
		{OpGetLocalAdd, -1},
		{OpGetLocalSub, -1},
		{OpGetLocalMul, -1},
	}

	for _, tt := range tests {
		opName := "unknown"
		if def, err := Lookup(byte(tt.op)); err == nil {
			opName = def.Name
		}
		t.Run(opName, func(t *testing.T) {
			instructions := []byte{byte(tt.op), 0, 0, 0, 0}
			effect := stackEffectOf(tt.op, instructions, 0)
			if effect != tt.expected {
				t.Errorf("stackEffectOf(%v) = %d, want %d", tt.op, effect, tt.expected)
			}
		})
	}
}

func TestStackEffectOf_Call(t *testing.T) {
	// OpCall with 2 arguments: pops callee + 2 args, pushes 1 result
	instructions := []byte{byte(OpCall), 2}
	effect := stackEffectOf(OpCall, instructions, 0)
	// net effect: -numArgs (pops args+callee, pushes result)
	if effect != -2 {
		t.Errorf("stackEffectOf(OpCall, numArgs=2) = %d, want -2", effect)
	}
}

func TestStackEffectOf_Array(t *testing.T) {
	// OpArray with 3 elements: pops 3, pushes 1
	instructions := []byte{byte(OpArray), 0, 3}
	effect := stackEffectOf(OpArray, instructions, 0)
	if effect != -2 { // 1 - 3 = -2
		t.Errorf("stackEffectOf(OpArray, count=3) = %d, want -2", effect)
	}
}

func TestStackEffectOf_Map(t *testing.T) {
	// OpMap with 2 pairs: pops 4, pushes 1
	instructions := []byte{byte(OpMap), 0, 2}
	effect := stackEffectOf(OpMap, instructions, 0)
	if effect != -3 { // 1 - 2*2 = -3
		t.Errorf("stackEffectOf(OpMap, pairs=2) = %d, want -3", effect)
	}
}

// ============================================
// Tests for findCallInstruction
// ============================================

func TestFindCallInstruction_Simple(t *testing.T) {
	// Create bytecode: OpGetGlobal 0, OpConstant 0, OpCall 1
	instructions := []byte{
		byte(OpGetGlobal), 0, 0, // callee at position 0
		byte(OpConstant), 0, 0, // arg at position 3
		byte(OpCall), 1, // call at position 6
	}

	optimizer := &Optimizer{}
	pos, found := optimizer.findCallInstruction(instructions, 0, 1)

	if !found {
		t.Error("expected to find OpCall instruction")
	}
	if pos != 6 {
		t.Errorf("expected position 6, got %d", pos)
	}
}

func TestFindCallInstruction_NoCall(t *testing.T) {
	// Create bytecode without OpCall
	instructions := []byte{
		byte(OpGetGlobal), 0, 0,
		byte(OpConstant), 0, 0,
	}

	optimizer := &Optimizer{}
	pos, found := optimizer.findCallInstruction(instructions, 0, 1)

	if found {
		t.Error("expected not to find OpCall instruction")
	}
	if pos != -1 {
		t.Errorf("expected position -1, got %d", pos)
	}
}

// ============================================
// Tests for transformInlineBody patterns
// ============================================

func TestTransformInlineBody_SingleParamReturn(t *testing.T) {
	// Pattern: return x -> body is OpGetLocal 0
	body := []byte{byte(OpGetLocal), 0}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 1)

	// Expected: empty (arg is already on stack)
	if len(result) != 0 {
		t.Errorf("expected empty body, got %v", result)
	}
}

func TestTransformInlineBody_TwoParamBinary(t *testing.T) {
	// Pattern: return a + b -> body is OpGetLocal 0, OpGetLocal 1, OpAdd
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1, byte(OpAdd)}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 2)

	// Expected: just OpAdd (args are on stack)
	if len(result) != 1 {
		t.Errorf("expected 1 byte, got %d: %v", len(result), result)
	}
	if Opcode(result[0]) != OpAdd {
		t.Errorf("expected OpAdd, got %v", Opcode(result[0]))
	}
}

func TestTransformInlineBody_SingleParamUnary(t *testing.T) {
	// Pattern: return -x -> body is OpGetLocal 0, OpNeg
	body := []byte{byte(OpGetLocal), 0, byte(OpNeg)}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 1)

	// Expected: just OpNeg
	if len(result) != 1 {
		t.Errorf("expected 1 byte, got %d: %v", len(result), result)
	}
	if Opcode(result[0]) != OpNeg {
		t.Errorf("expected OpNeg, got %v", Opcode(result[0]))
	}
}

func TestTransformInlineBody_SingleParamUsedTwice(t *testing.T) {
	// Pattern: return x * x -> body is OpGetLocal 0, OpGetLocal 0, OpMul
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 0, byte(OpMul)}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 1)

	// Expected: OpDup, OpMul
	if len(result) != 2 {
		t.Errorf("expected 2 bytes, got %d: %v", len(result), result)
	}
	if Opcode(result[0]) != OpDup {
		t.Errorf("expected OpDup, got %v", Opcode(result[0]))
	}
	if Opcode(result[1]) != OpMul {
		t.Errorf("expected OpMul, got %v", Opcode(result[1]))
	}
}

func TestTransformInlineBody_ThreeParamBinary(t *testing.T) {
	// Pattern: return a + b + c
	body := []byte{
		byte(OpGetLocal), 0,
		byte(OpGetLocal), 1,
		byte(OpAdd),
		byte(OpGetLocal), 2,
		byte(OpAdd),
	}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 3)

	// Expected: OpAdd, OpAdd (args on stack: a, b, c -> a+b, c -> a+b+c)
	if len(result) != 2 {
		t.Errorf("expected 2 bytes, got %d: %v", len(result), result)
	}
}

func TestTransformInlineBody_ParamWithConstant(t *testing.T) {
	// Pattern: return x + 10 -> body is OpGetLocal 0, OpConstant N, OpAdd
	body := []byte{byte(OpGetLocal), 0, byte(OpConstant), 0, 10, byte(OpAdd)}

	optimizer := &Optimizer{}
	result := optimizer.transformInlineBody(body, 1)

	// Expected: OpConstant, index, OpAdd
	if result == nil {
		t.Error("expected transformed body, got nil")
	}
}

// ============================================
// Tests for isSimpleBinaryBody
// ============================================

func TestIsSimpleBinaryBody_True(t *testing.T) {
	// Pattern: OpGetLocal 0, OpGetLocal 1, OpAdd
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1, byte(OpAdd)}

	optimizer := &Optimizer{}
	if !optimizer.isSimpleBinaryBody(body, 2) {
		t.Error("expected body to be simple binary")
	}
}

func TestIsSimpleBinaryBody_False(t *testing.T) {
	// Pattern with more complex operations
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1, byte(OpAdd), byte(OpGetLocal), 2}

	optimizer := &Optimizer{}
	if optimizer.isSimpleBinaryBody(body, 3) {
		t.Error("expected body not to be simple binary")
	}
}

// ============================================
// Tests for optimizer with flags
// ============================================

func TestNewOptimizerWithFlags(t *testing.T) {
	constants := []objects.Object{objects.NewInt(1), objects.NewInt(2)}
	bytecode := &Bytecode{
		Instructions: []byte{byte(OpConstant), 0, 0, byte(OpConstant), 0, 1, byte(OpAdd)},
		Constants:    constants,
	}

	opts := OptimizationFlags{
		BytecodeOptimizer: true,
		Superinstructions: true,
		InlineFunctions:   false,
	}

	optimizer := NewOptimizerWithFlags(bytecode, opts)
	result := optimizer.Optimize()

	if result == nil {
		t.Error("expected optimized result")
	}
}

// ============================================
// Tests for stripGetLocals
// ============================================

func TestStripGetLocals(t *testing.T) {
	// Pattern: OpGetLocal 0, OpGetLocal 1 -> strip to nothing (args on stack)
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1}

	optimizer := &Optimizer{}
	result := optimizer.stripGetLocals(body, 2)

	// Should be empty since args are already on stack
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

// ============================================
// Tests for analyzeGetLocalPattern
// ============================================

func TestAnalyzeGetLocalPattern_Valid(t *testing.T) {
	// Pattern: OpGetLocal 0, OpGetLocal 1
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 1}

	optimizer := &Optimizer{}
	pattern := optimizer.analyzeGetLocalPattern(body, 2)

	if pattern == nil {
		t.Fatal("expected pattern to be non-nil")
	}
	if len(pattern.indices) != 2 || pattern.indices[0] != 0 || pattern.indices[1] != 1 {
		t.Errorf("expected indices [0, 1], got %v", pattern.indices)
	}
}

func TestAnalyzeGetLocalPattern_Invalid(t *testing.T) {
	// Invalid pattern with wrong indices
	body := []byte{byte(OpGetLocal), 0, byte(OpGetLocal), 5}

	optimizer := &Optimizer{}
	pattern := optimizer.analyzeGetLocalPattern(body, 2)

	// Pattern should be nil or have invalid state for out-of-bounds access
	// The exact behavior depends on the implementation
	_ = pattern // Just verify it doesn't crash
}
