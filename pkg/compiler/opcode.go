// pkg/compiler/opcode.go
package compiler

import "fmt"

// GlobalsSize is the maximum number of global variables
const GlobalsSize = 65536

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
	OpGetFree     // Get free variable (from closure)
	OpSetFree     // Set free variable (in closure)

	// Scope operations
	OpPushScope // Push new scope
	OpPopScope  // Pop scope

	// Superinstructions (combined operations for performance)
	OpGetLocalAdd    // Get two locals and add
	OpGetLocalSub    // Get two locals and subtract
	OpGetLocalMul    // Get two locals and multiply
	OpConstantAdd    // Load two constants and add
	OpConstantSub    // Load two constants and subtract
	OpConstantMul    // Load two constants and multiply

	// Control flow operations
	OpJump        // Unconditional jump
	OpJumpIfFalse // Jump if top of stack is false
	OpJumpIfTrue  // Jump if top of stack is true
	OpCall        // Call function
	OpTailCall    // Tail call (reuse current frame)
	OpReturn      // Return from function
	OpClosure     // Create closure with captured variables

	// Collection operations
	OpArray   // Create array from stack elements
	OpMap     // Create map from stack elements
	OpIndex   // Index into array/map
	OpSetIndex // Set index in array/map
	OpIndexSafe // Index into array without bounds check (safe context)

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

	// Module operations
	OpLoadModule // Load module and push onto stack
	OpGetExport  // Get export from module
	OpModule     // Create module from exports
	OpSetExport  // Set export in current module

	// Class operations
	OpClass    // Create class object
	OpNew      // Create instance
	OpGetField // Get instance field
	OpSetField // Set instance field
	OpSuper    // Get superclass method
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
	OpGetFree:     {"OpGetFree", []int{1}},     // 1-byte free variable index
	OpSetFree:     {"OpSetFree", []int{1}},     // 1-byte free variable index

	// Scope operations
	OpPushScope: {"OpPushScope", []int{}},
	OpPopScope:  {"OpPopScope", []int{}},

	// Control flow operations
	OpJump:        {"OpJump", []int{2}},        // 2-byte jump offset
	OpJumpIfFalse: {"OpJumpIfFalse", []int{2}}, // 2-byte jump offset
	OpJumpIfTrue:  {"OpJumpIfTrue", []int{2}},  // 2-byte jump offset
	OpCall:        {"OpCall", []int{1}},        // 1-byte argument count
	OpTailCall:    {"OpTailCall", []int{1}},    // 1-byte argument count
	OpReturn:      {"OpReturn", []int{}},
	OpClosure:     {"OpClosure", []int{2, 1}},  // 2-byte constant index, 1-byte num free

	// Collection operations
	OpArray:     {"OpArray", []int{2}},    // 2-byte element count
	OpMap:       {"OpMap", []int{2}},      // 2-byte pair count
	OpIndex:     {"OpIndex", []int{}},
	OpSetIndex:  {"OpSetIndex", []int{}},
	OpIndexSafe: {"OpIndexSafe", []int{}}, // No bounds check

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

	// Module operations
	OpLoadModule: {"OpLoadModule", []int{2}}, // 2-byte constant index for path
	OpGetExport:  {"OpGetExport", []int{2}},  // 2-byte constant index for name
	OpModule:     {"OpModule", []int{2}},     // 2-byte export count
	OpSetExport:  {"OpSetExport", []int{2}},  // 2-byte constant index for name

	// Class operations
	OpClass:    {"OpClass", []int{2}},    // 2-byte: class name constant index
	OpNew:      {"OpNew", []int{1}},      // 1-byte: argument count
	OpGetField: {"OpGetField", []int{2}}, // 2-byte: field name constant index
	OpSetField: {"OpSetField", []int{2}}, // 2-byte: field name constant index
	OpSuper:    {"OpSuper", []int{2}},    // 2-byte: method name constant index

	// Superinstructions
	OpGetLocalAdd: {"OpGetLocalAdd", []int{1, 1}}, // 2x 1-byte local indices
	OpGetLocalSub: {"OpGetLocalSub", []int{1, 1}},
	OpGetLocalMul: {"OpGetLocalMul", []int{1, 1}},
	OpConstantAdd: {"OpConstantAdd", []int{2, 2}}, // 2x 2-byte constant indices
	OpConstantSub: {"OpConstantSub", []int{2, 2}},
	OpConstantMul: {"OpConstantMul", []int{2, 2}},
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
