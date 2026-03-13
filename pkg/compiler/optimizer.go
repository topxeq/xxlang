// pkg/compiler/optimizer.go
// Bytecode optimizer for compile-time transformations
package compiler

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// Optimizer transforms bytecode to reduce instruction count
// and improve execution patterns
type Optimizer struct {
	bytecode *Bytecode
}

// NewOptimizer creates a new optimizer for given bytecode
func NewOptimizer(bytecode *Bytecode) *Optimizer {
	return &Optimizer{bytecode: bytecode}
}

// Optimize runs all optimization passes
func (o *Optimizer) Optimize() *Bytecode {
	result := o.FoldConstants()
	result = o.GenerateSuperinstructions()
	return result
}

// FoldConstants evaluates constant expressions at compile time
// Pattern: Constant; Constant; BinaryOp → replace with single Constant
func (o *Optimizer) FoldConstants() *Bytecode {
	instructions := o.bytecode.Instructions
	constants := o.bytecode.Constants

	if len(instructions) < 3 {
		return o.bytecode
	}

	// Scan for constant folding opportunities
	// Pattern: OpConstant, OpConstant, OpAdd/OpSub/OpMul/OpDiv
	newInstructions := make([]byte, 0, len(instructions))

	// Start with all original constants to preserve indices
	newConstants := make([]objects.Object, len(constants))
	copy(newConstants, constants)

	i := 0
	for i < len(instructions) {
		op := Opcode(instructions[i])

		// Check for constant folding pattern
		if op == OpConstant && i+6 < len(instructions) {
			// First constant index
			idx1 := int(instructions[i+1])<<8 | int(instructions[i+2])

			// Check if next instruction is also OpConstant
			if Opcode(instructions[i+3]) == OpConstant {
				idx2 := int(instructions[i+4])<<8 | int(instructions[i+5])

				// Check if followed by binary operation
				binOp := Opcode(instructions[i+6])

				// Try to fold
				if result, ok := o.foldBinaryOp(constants[idx1], constants[idx2], binOp); ok {
					// Add result to new constants
					newIdx := len(newConstants)
					newConstants = append(newConstants, result)

					// Emit single OpConstant with folded result
					instr := Make(OpConstant, newIdx)
					newInstructions = append(newInstructions, instr...)

					// Skip the three instructions we just folded
					i += 7
					continue
				}
			}
		}

		// Copy instruction as-is
		def, err := Lookup(byte(op))
		if err != nil {
			i++
			continue
		}

		instrLen := 1
		for _, w := range def.OperandWidths {
			instrLen += w
		}

		newInstructions = append(newInstructions, instructions[i:i+instrLen]...)
		i += instrLen
	}

	// Update bytecode
	o.bytecode.Instructions = newInstructions
	o.bytecode.Constants = newConstants

	return o.bytecode
}

// foldBinaryOp evaluates a binary operation on constant operands
func (o *Optimizer) foldBinaryOp(left, right objects.Object, op Opcode) (objects.Object, bool) {
	leftInt, leftIsInt := left.(*objects.Int)
	rightInt, rightIsInt := right.(*objects.Int)

	if leftIsInt && rightIsInt {
		leftVal := leftInt.Value
		rightVal := rightInt.Value

		switch op {
		case OpAdd:
			return objects.NewInt(leftVal + rightVal), true
		case OpSub:
			return objects.NewInt(leftVal - rightVal), true
		case OpMul:
			return objects.NewInt(leftVal * rightVal), true
		case OpDiv:
			if rightVal != 0 {
				return objects.NewInt(leftVal / rightVal), true
			}
		case OpMod:
			if rightVal != 0 {
				return objects.NewInt(leftVal % rightVal), true
			}
		}
	}

	// Float operations
	leftFloat, leftIsFloat := left.(*objects.Float)
	rightFloat, rightIsFloat := right.(*objects.Float)

	if leftIsFloat && rightIsFloat {
		leftVal := leftFloat.Value
		rightVal := rightFloat.Value

		switch op {
		case OpAdd:
			return &objects.Float{Value: leftVal + rightVal}, true
		case OpSub:
			return &objects.Float{Value: leftVal - rightVal}, true
		case OpMul:
			return &objects.Float{Value: leftVal * rightVal}, true
		case OpDiv:
			if rightVal != 0 {
				return &objects.Float{Value: leftVal / rightVal}, true
			}
		}
	}

	return nil, false
}

// GenerateSuperinstructions combines common instruction sequences into single instructions
func (o *Optimizer) GenerateSuperinstructions() *Bytecode {
	instructions := o.bytecode.Instructions

	if len(instructions) < 3 {
		return o.bytecode
	}

	newInstructions := make([]byte, 0, len(instructions))

	i := 0
	for i < len(instructions) {
		op := Opcode(instructions[i])

		// Pattern: OpGetLocal + OpGetLocal + OpAdd/Sub/Mul
		if op == OpGetLocal && i+4 < len(instructions) {
			idx1 := int(instructions[i+1])

			if Opcode(instructions[i+2]) == OpGetLocal {
				idx2 := int(instructions[i+3])
				binOp := Opcode(instructions[i+4])

				// Check if both indices fit in a byte (0-255)
				if idx1 <= 255 && idx2 <= 255 {
					var superOp Opcode
					switch binOp {
					case OpAdd:
						superOp = OpGetLocalAdd
					case OpSub:
						superOp = OpGetLocalSub
					case OpMul:
						superOp = OpGetLocalMul
					}

					if superOp != 0 {
						// Emit superinstruction
						newInstructions = append(newInstructions, byte(superOp), byte(idx1), byte(idx2))
						i += 5 // Skip OpGetLocal(2) + OpGetLocal(2) + BinOp(1)
						continue
					}
				}
			}
		}

		// Pattern: OpConstant + OpConstant + OpAdd/Sub/Mul
		// OpConstant = 3 bytes (1 opcode + 2 byte index), so 2*3+1 = 7 bytes total
		if op == OpConstant && i+6 < len(instructions) {
			idx1 := int(instructions[i+1])<<8 | int(instructions[i+2])

			if Opcode(instructions[i+3]) == OpConstant {
				idx2 := int(instructions[i+4])<<8 | int(instructions[i+5])
				binOp := Opcode(instructions[i+6])

				// Only fold if both constants are integers (not functions, strings, etc.)
				constants := o.bytecode.Constants
				_, ok1 := constants[idx1].(*objects.Int)
				_, ok2 := constants[idx2].(*objects.Int)

				if ok1 && ok2 {
					var superOp Opcode
					switch binOp {
					case OpAdd:
						superOp = OpConstantAdd
					case OpSub:
						superOp = OpConstantSub
					case OpMul:
						superOp = OpConstantMul
					}

					if superOp != 0 {
						// Emit superinstruction
						newInstructions = append(newInstructions, byte(superOp))
						newInstructions = append(newInstructions, byte(idx1>>8), byte(idx1))
						newInstructions = append(newInstructions, byte(idx2>>8), byte(idx2))
						i += 7 // Skip OpConstant(3) + OpConstant(3) + BinOp(1)
						continue
					}
				}
			}
		}

		// Copy instruction as-is
		def, err := Lookup(byte(op))
		if err != nil {
			i++
			continue
		}

		instrLen := 1
		for _, w := range def.OperandWidths {
			instrLen += w
		}

		newInstructions = append(newInstructions, instructions[i:i+instrLen]...)
		i += instrLen
	}

	o.bytecode.Instructions = newInstructions
	return o.bytecode
}
