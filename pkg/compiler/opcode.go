// pkg/compiler/opcode.go
package compiler

import "fmt"

// Opcode represents a single bytecode instruction
type Opcode byte

const (
	// Stack operations
	OpConstant Opcode = iota // Load constant from constant pool
	OpPop                    // Pop value from stack
	OpDup                    // Duplicate top of stack

	// Arithmetic operations
	OpAdd  // Addition
	OpSub  // Subtraction
	OpMul  // Multiplication
	OpDiv  // Division
	OpMod  // Modulo
	OpNeg  // Unary minus (negation)

	// Comparison operations
	OpEqual       // Equality check
	OpNotEqual    // Inequality check
	OpLess        // Less than
	OpGreater     // Greater than
	OpLessEqual   // Less than or equal
	OpGreaterEqual // Greater than or equal

	// Logical operations
	OpAnd // Logical AND
	OpOr  // Logical OR
	OpNot // Logical NOT

	// Variable operations
	OpGetLocal    // Get local variable
	OpSetLocal    // Set local variable
	OpDefineLocal // Define local variable
	OpGetGlobal   // Get global variable
	OpSetGlobal   // Set global variable

	// Scope operations
	OpPushScope // Push new scope
	OpPopScope  // Pop scope

	// Control flow operations
	OpJump        // Unconditional jump
	OpJumpIfFalse // Jump if top of stack is false
	OpJumpIfTrue  // Jump if top of stack is true
	OpCall        // Call function
	OpReturn      // Return from function

	// Collection operations
	OpArray   // Create array from stack elements
	OpMap     // Create map from stack elements
	OpIndex   // Index into array/map
	OpSetIndex // Set index in array/map

	// Method/Field operations
	OpGetMethod  // Get method from object
	OpCallMethod // Call method on object

	// Built-in operations
	OpBuiltin // Call built-in function

	// Literal operations
	OpNull  // Push null onto stack
	OpTrue  // Push true onto stack
	OpFalse // Push false onto stack

	// Loop control
	OpBreak    // Break from loop
	OpContinue // Continue to next iteration
)

// Definition describes an opcode's format
type Definition struct {
	Name          string // Human-readable name
	OperandWidths []int  // Width in bytes of each operand
}

// definitions maps opcodes to their definitions
var definitions = map[Opcode]*Definition{
	// Stack operations
	OpConstant: {"OpConstant", []int{2}}, // 2-byte constant index
	OpPop:      {"OpPop", []int{}},
	OpDup:      {"OpDup", []int{}},

	// Arithmetic operations
	OpAdd:  {"OpAdd", []int{}},
	OpSub:  {"OpSub", []int{}},
	OpMul:  {"OpMul", []int{}},
	OpDiv:  {"OpDiv", []int{}},
	OpMod:  {"OpMod", []int{}},
	OpNeg:  {"OpNeg", []int{}},

	// Comparison operations
	OpEqual:        {"OpEqual", []int{}},
	OpNotEqual:     {"OpNotEqual", []int{}},
	OpLess:         {"OpLess", []int{}},
	OpGreater:      {"OpGreater", []int{}},
	OpLessEqual:    {"OpLessEqual", []int{}},
	OpGreaterEqual: {"OpGreaterEqual", []int{}},

	// Logical operations
	OpAnd: {"OpAnd", []int{}},
	OpOr:  {"OpOr", []int{}},
	OpNot: {"OpNot", []int{}},

	// Variable operations
	OpGetLocal:    {"OpGetLocal", []int{1}},    // 1-byte local index
	OpSetLocal:    {"OpSetLocal", []int{1}},    // 1-byte local index
	OpDefineLocal: {"OpDefineLocal", []int{1}}, // 1-byte local index
	OpGetGlobal:   {"OpGetGlobal", []int{2}},   // 2-byte global index
	OpSetGlobal:   {"OpSetGlobal", []int{2}},   // 2-byte global index

	// Scope operations
	OpPushScope: {"OpPushScope", []int{}},
	OpPopScope:  {"OpPopScope", []int{}},

	// Control flow operations
	OpJump:        {"OpJump", []int{2}},        // 2-byte jump offset
	OpJumpIfFalse: {"OpJumpIfFalse", []int{2}}, // 2-byte jump offset
	OpJumpIfTrue:  {"OpJumpIfTrue", []int{2}},  // 2-byte jump offset
	OpCall:        {"OpCall", []int{1}},        // 1-byte argument count
	OpReturn:      {"OpReturn", []int{}},

	// Collection operations
	OpArray:    {"OpArray", []int{2}},    // 2-byte element count
	OpMap:      {"OpMap", []int{2}},      // 2-byte pair count
	OpIndex:    {"OpIndex", []int{}},
	OpSetIndex: {"OpSetIndex", []int{}},

	// Method/Field operations
	OpGetMethod:  {"OpGetMethod", []int{2}},  // 2-byte constant index for name
	OpCallMethod: {"OpCallMethod", []int{1}}, // 1-byte argument count

	// Built-in operations
	OpBuiltin: {"OpBuiltin", []int{1}}, // 1-byte built-in index

	// Literal operations
	OpNull:  {"OpNull", []int{}},
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	// Loop control
	OpBreak:    {"OpBreak", []int{}},
	OpContinue: {"OpContinue", []int{}},
}

// Lookup finds an opcode's definition
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates a bytecode instruction with operands
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 1:
			instruction[offset] = byte(operand)
		case 2:
			instruction[offset] = byte(operand >> 8)
			instruction[offset+1] = byte(operand)
		}
		offset += width
	}

	return instruction
}

// ReadOperands reads operands from bytecode
func ReadOperands(def *Definition, ins []byte) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(ins[offset])
		case 2:
			operands[i] = int(ins[offset])<<8 | int(ins[offset+1])
		}
		offset += width
	}

	return operands, offset
}

// String prints bytecode as a formatted string
func String(ins []byte) string {
	var out string
	i := 0

	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			out += fmt.Sprintf("%04d ERROR: %s\n", i, err)
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		out += fmt.Sprintf("%04d %s", i, def.Name)

		for _, op := range operands {
			out += fmt.Sprintf(" %d", op)
		}
		out += "\n"

		i += 1 + read
	}

	return out
}
