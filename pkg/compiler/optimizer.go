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
	// More passes can be added here
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
