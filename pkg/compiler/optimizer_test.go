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
		byte(OpConstant), 0, 0,  // constant index 0 (big-endian)
		byte(OpConstant), 0, 1,  // constant index 1 (big-endian)
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
		byte(OpConstant), 0, 0,  // constant index 0
		byte(OpConstant), 0, 1,  // constant index 1
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
		byte(OpConstant), 0, 0,  // push 5 (index 0)
		byte(OpConstant), 0, 1,  // push 3 (index 1)
		byte(OpAdd),             // add -> would be folded to 8
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
