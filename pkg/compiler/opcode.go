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
	OpAdd // Addition
	OpSub // Subtraction
	OpMul // Multiplication
	OpDiv // Division
	OpMod // Modulo
	OpNeg // Unary minus (negation)

	// Comparison operations
	OpEqual        // Equality check
	OpNotEqual     // Inequality check
	OpLess         // Less than
	OpGreater      // Greater than
	OpLessEqual    // Less than or equal
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
	OpGetLocalAdd     // Get two locals and add
	OpGetLocalSub     // Get two locals and subtract
	OpGetLocalMul     // Get two locals and multiply
	OpConstantAdd     // Load two constants and add
	OpConstantSub     // Load two constants and subtract
	OpConstantMul     // Load two constants and multiply
	OpGetLocalLess    // Get two locals and compare less
	OpGetLocalGreater // Get two locals and compare greater
	OpGetLocalEqual   // Get two locals and compare equal
	OpGetLocalNotEqual // Get two locals and compare not equal

	// Type-specialized instructions (for hot path optimization)
	OpIncLocal      // Increment local variable by 1 (optimized for loop counters)
	OpDecLocal      // Decrement local variable by 1
	OpAddLocalConst // Add constant to local variable
	OpSubLocalConst // Subtract constant from local variable
	OpMulLocalConst // Multiply local variable by constant

	// Control flow operations
	OpJump        // Unconditional jump
	OpJumpIfFalse // Jump if top of stack is false
	OpJumpIfTrue  // Jump if top of stack is true
	OpCall        // Call function
	OpTailCall    // Tail call (reuse current frame)
	OpReturn      // Return from function
	OpClosure     // Create closure with captured variables

	// Collection operations
	OpArray     // Create array from stack elements
	OpMap       // Create map from stack elements
	OpIndex     // Index into array/map
	OpSetIndex  // Set index in array/map
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

	// Exception handling
	OpThrow       // Throw exception (value on stack)
	OpPushHandler // Push exception handler (catchAddr, finallyAddr)
	OpPopHandler  // Pop exception handler

	// Value-based operations (NaN boxing optimized - zero allocation)
	// These are the fast path for numeric operations
	OpValueAdd // Add two Values (no type check, direct numeric)
	OpValueSub // Subtract two Values
	OpValueMul // Multiply two Values
	OpValueDiv // Divide two Values
	OpValueMod // Modulo two Values
	OpValueNeg // Negate a Value

	// Value-based comparisons
	OpValueLess        // Compare two Values for less than
	OpValueGreater     // Compare two Values for greater than
	OpValueEqual       // Compare two Values for equality
	OpValueNotEqual    // Compare two Values for inequality
	OpValueLessEqual   // Compare two Values for less or equal
	OpValueGreaterEqual // Compare two Values for greater or equal

	// Value-based local operations (combined for zero allocation)
	OpValueGetLocal      // Get local as Value (push to value stack)
	OpValueSetLocal      // Set local from Value (pop from value stack)
	OpValueIncLocal      // Increment local by 1
	OpValueDecLocal      // Decrement local by 1
	OpValueAddLocalConst // Add constant to local
	OpValueSubLocalConst // Subtract constant from local
	OpValueMulLocalConst // Multiply local by constant

	// Value-based superinstructions (maximum optimization)
	OpValueGetLocalAdd     // Get two locals and add (Value path)
	OpValueGetLocalSub     // Get two locals and subtract (Value path)
	OpValueGetLocalMul     // Get two locals and multiply (Value path)
	OpValueGetLocalLess    // Get two locals and compare less (Value path)
	OpValueGetLocalGreater // Get two locals and compare greater (Value path)
	OpValueGetLocalEqual   // Get two locals and compare equal (Value path)

	// ============================================================================
	// REGISTER-BASED OPERATIONS (New VM architecture)
	// ============================================================================
	// Fixed 4-byte format: opcode(8) | dst(8) | src1(8) | src2(8)
	// Extended format: opcode(8) | reg(8) | const_idx(16)

	// Register arithmetic: R[dst] = R[src1] op R[src2]
	OpRegAdd    // R[dst] = R[src1] + R[src2]
	OpRegSub    // R[dst] = R[src1] - R[src2]
	OpRegMul    // R[dst] = R[src1] * R[src2]
	OpRegDiv    // R[dst] = R[src1] / R[src2]
	OpRegMod    // R[dst] = R[src1] % R[src2]
	OpRegNeg    // R[dst] = -R[src1] (src2 unused)
	OpRegAnd    // R[dst] = R[src1] && R[src2]
	OpRegOr     // R[dst] = R[src1] || R[src2]
	OpRegNot    // R[dst] = !R[src1] (src2 unused)

	// Register comparison: R[dst] = R[src1] op R[src2] (result is boolean)
	OpRegLess        // R[dst] = R[src1] < R[src2]
	OpRegLessEqual   // R[dst] = R[src1] <= R[src2]
	OpRegGreater     // R[dst] = R[src1] > R[src2]
	OpRegGreaterEqual // R[dst] = R[src1] >= R[src2]
	OpRegEqual       // R[dst] = R[src1] == R[src2]
	OpRegNotEqual    // R[dst] = R[src1] != R[src2]

	// Register data movement (extended format: opcode | reg(8) | idx(16))
	OpRegLoadConst  // R[dst] = Constants[idx]
	OpRegLoadGlobal // R[dst] = Globals[idx]
	OpRegStoreGlobal // Globals[idx] = R[src]
	OpRegMove       // R[dst] = R[src] (src2 unused, but format consistent)

	// Register local variable operations
	OpRegLoadLocal  // R[dst] = Locals[idx]
	OpRegStoreLocal // Locals[idx] = R[src]
	OpRegLoadFree   // R[dst] = FreeVars[idx]
	OpRegStoreFree  // FreeVars[idx] = R[src]

	// Register control flow (extended format for jump offsets)
	OpRegJump        // IP += offset (signed 16-bit)
	OpRegJumpIfTrue  // if R[cond] then IP += offset
	OpRegJumpIfFalse // if !R[cond] then IP += offset

	// Register function call convention
	// R0-R7: argument registers
	// RRet (255): return value register
	OpRegCall   // Call function, args in R0-R7, result in RRet
	OpRegReturn // Return value in R[reg]

	// Register closure and function operations
	OpRegClosure   // Create closure from constant function
	OpRegLoadFunc  // Load function from constant pool into register

	// Register collection operations
	OpRegArray     // R[dst] = Array from R[src1..src1+src2-1]
	OpRegMap       // R[dst] = Map from pairs starting at R[src1]
	OpRegIndex     // R[dst] = R[obj][R[key]]
	OpRegSetIndex  // R[obj][R[key]] = R[val]
	OpRegPush      // Push R[src] to temp stack (for complex exprs)
	OpRegPop       // Pop to R[dst] from temp stack

	// Register method/field operations (extended format for name index)
	OpRegGetMethod  // R[dst] = R[obj].method(name_idx)
	OpRegCallMethod // Call method, args in R1-R7, result in RRet
	OpRegGetField   // R[dst] = R[obj].field(name_idx)
	OpRegSetField   // R[obj].field(name_idx) = R[src]

	// Register class operations
	OpRegClass // Create class
	OpRegNew   // Create instance, args in R0-R7, result in RRet

	// Register built-in call
	OpRegBuiltin // Call builtin, args in R0-R7, result in RRet

	// Register null/true/false literals
	OpRegNull  // R[dst] = null
	OpRegTrue  // R[dst] = true
	OpRegFalse // R[dst] = false

	// Register exception handling
	OpRegThrow       // throw R[src]
	OpRegPushHandler // push exception handler
	OpRegPopHandler  // pop exception handler

	// Register superinstructions for common patterns
	OpRegAddConst    // R[dst] = R[src1] + Constants[idx]
	OpRegSubConst    // R[dst] = R[src1] - Constants[idx]
	OpRegMulConst    // R[dst] = R[src1] * Constants[idx]
	OpRegIncLocal    // Locals[idx]++
	OpRegDecLocal    // Locals[idx]--

	// Register module operations
	OpRegLoadModule // R[dst] = LoadModule(Constants[idx])
	OpRegGetExport  // R[dst] = R[module].Exports[name_idx]
	OpRegSetExport  // CurrentModule.Exports[name_idx] = R[src]
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
	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},
	OpMod: {"OpMod", []int{}},
	OpNeg: {"OpNeg", []int{}},

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
	OpClosure:     {"OpClosure", []int{2, 1}}, // 2-byte constant index, 1-byte num free

	// Collection operations
	OpArray:     {"OpArray", []int{2}}, // 2-byte element count
	OpMap:       {"OpMap", []int{2}},   // 2-byte pair count
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

	// Exception handling
	OpThrow:       {"OpThrow", []int{}},           // No operands, value on stack
	OpPushHandler: {"OpPushHandler", []int{2, 2}}, // 2-byte catchAddr, 2-byte finallyAddr
	OpPopHandler:  {"OpPopHandler", []int{}},      // No operands

	// Superinstructions
	OpGetLocalAdd: {"OpGetLocalAdd", []int{1, 1}}, // 2x 1-byte local indices
	OpGetLocalSub: {"OpGetLocalSub", []int{1, 1}},
	OpGetLocalMul: {"OpGetLocalMul", []int{1, 1}},
	OpConstantAdd: {"OpConstantAdd", []int{2, 2}}, // 2x 2-byte constant indices
	OpConstantSub: {"OpConstantSub", []int{2, 2}},
	OpConstantMul: {"OpConstantMul", []int{2, 2}},

	// Comparison superinstructions
	OpGetLocalLess:    {"OpGetLocalLess", []int{1, 1}},    // 2x 1-byte local indices
	OpGetLocalGreater: {"OpGetLocalGreater", []int{1, 1}}, // 2x 1-byte local indices
	OpGetLocalEqual:   {"OpGetLocalEqual", []int{1, 1}},   // 2x 1-byte local indices
	OpGetLocalNotEqual: {"OpGetLocalNotEqual", []int{1, 1}}, // 2x 1-byte local indices

	// Type-specialized instructions
	OpIncLocal:      {"OpIncLocal", []int{1}},         // 1-byte: local index
	OpDecLocal:      {"OpDecLocal", []int{1}},         // 1-byte: local index
	OpAddLocalConst: {"OpAddLocalConst", []int{1, 2}}, // 1-byte local index, 2-byte constant index
	OpSubLocalConst: {"OpSubLocalConst", []int{1, 2}}, // 1-byte local index, 2-byte constant index
	OpMulLocalConst: {"OpMulLocalConst", []int{1, 2}}, // 1-byte local index, 2-byte constant index

	// Value-based arithmetic operations (zero allocation)
	OpValueAdd: {"OpValueAdd", []int{}}, // Pop 2, push result
	OpValueSub: {"OpValueSub", []int{}},
	OpValueMul: {"OpValueMul", []int{}},
	OpValueDiv: {"OpValueDiv", []int{}},
	OpValueMod: {"OpValueMod", []int{}},
	OpValueNeg: {"OpValueNeg", []int{}},

	// Value-based comparisons
	OpValueLess:        {"OpValueLess", []int{}},
	OpValueGreater:     {"OpValueGreater", []int{}},
	OpValueEqual:       {"OpValueEqual", []int{}},
	OpValueNotEqual:    {"OpValueNotEqual", []int{}},
	OpValueLessEqual:   {"OpValueLessEqual", []int{}},
	OpValueGreaterEqual: {"OpValueGreaterEqual", []int{}},

	// Value-based local operations
	OpValueGetLocal:      {"OpValueGetLocal", []int{1}},         // 1-byte: local index
	OpValueSetLocal:      {"OpValueSetLocal", []int{1}},         // 1-byte: local index
	OpValueIncLocal:      {"OpValueIncLocal", []int{1}},         // 1-byte: local index
	OpValueDecLocal:      {"OpValueDecLocal", []int{1}},         // 1-byte: local index
	OpValueAddLocalConst: {"OpValueAddLocalConst", []int{1, 2}}, // 1-byte local, 2-byte const
	OpValueSubLocalConst: {"OpValueSubLocalConst", []int{1, 2}},
	OpValueMulLocalConst: {"OpValueMulLocalConst", []int{1, 2}},

	// Value-based superinstructions
	OpValueGetLocalAdd:     {"OpValueGetLocalAdd", []int{1, 1}},     // 2x 1-byte local indices
	OpValueGetLocalSub:     {"OpValueGetLocalSub", []int{1, 1}},
	OpValueGetLocalMul:     {"OpValueGetLocalMul", []int{1, 1}},
	OpValueGetLocalLess:    {"OpValueGetLocalLess", []int{1, 1}},
	OpValueGetLocalGreater: {"OpValueGetLocalGreater", []int{1, 1}},
	OpValueGetLocalEqual:   {"OpValueGetLocalEqual", []int{1, 1}},

	// ============================================================================
	// REGISTER-BASED OPERATIONS
	// ============================================================================

	// Register arithmetic: 4-byte format - dst, src1, src2
	OpRegAdd:    {"OpRegAdd", []int{1, 1, 1}},    // dst, src1, src2
	OpRegSub:    {"OpRegSub", []int{1, 1, 1}},    // dst, src1, src2
	OpRegMul:    {"OpRegMul", []int{1, 1, 1}},    // dst, src1, src2
	OpRegDiv:    {"OpRegDiv", []int{1, 1, 1}},    // dst, src1, src2
	OpRegMod:    {"OpRegMod", []int{1, 1, 1}},    // dst, src1, src2
	OpRegNeg:    {"OpRegNeg", []int{1, 1}},       // dst, src
	OpRegAnd:    {"OpRegAnd", []int{1, 1, 1}},    // dst, src1, src2
	OpRegOr:     {"OpRegOr", []int{1, 1, 1}},     // dst, src1, src2
	OpRegNot:    {"OpRegNot", []int{1, 1}},       // dst, src

	// Register comparison: 4-byte format
	OpRegLess:         {"OpRegLess", []int{1, 1, 1}},        // dst, src1, src2
	OpRegLessEqual:    {"OpRegLessEqual", []int{1, 1, 1}},   // dst, src1, src2
	OpRegGreater:      {"OpRegGreater", []int{1, 1, 1}},     // dst, src1, src2
	OpRegGreaterEqual: {"OpRegGreaterEqual", []int{1, 1, 1}}, // dst, src1, src2
	OpRegEqual:        {"OpRegEqual", []int{1, 1, 1}},       // dst, src1, src2
	OpRegNotEqual:     {"OpRegNotEqual", []int{1, 1, 1}},    // dst, src1, src2

	// Register data movement (extended format: reg(8), idx(16))
	OpRegLoadConst:   {"OpRegLoadConst", []int{1, 2}},   // dst, const_idx
	OpRegLoadGlobal:  {"OpRegLoadGlobal", []int{1, 2}},  // dst, global_idx
	OpRegStoreGlobal: {"OpRegStoreGlobal", []int{1, 2}}, // src, global_idx
	OpRegMove:        {"OpRegMove", []int{1, 1}},        // dst, src

	// Register local/free operations
	OpRegLoadLocal:  {"OpRegLoadLocal", []int{1, 1}},  // dst, local_idx
	OpRegStoreLocal: {"OpRegStoreLocal", []int{1, 1}}, // src, local_idx
	OpRegLoadFree:   {"OpRegLoadFree", []int{1, 1}},   // dst, free_idx
	OpRegStoreFree:  {"OpRegStoreFree", []int{1, 1}},  // src, free_idx

	// Register control flow (extended format for 16-bit offset)
	OpRegJump:        {"OpRegJump", []int{1, 2}},       // unused_byte, offset (16-bit signed) - 4 bytes total
	OpRegJumpIfTrue:  {"OpRegJumpIfTrue", []int{1, 2}},  // cond, offset
	OpRegJumpIfFalse: {"OpRegJumpIfFalse", []int{1, 2}}, // cond, offset

	// Register function operations
	OpRegCall:   {"OpRegCall", []int{1, 1}},   // func_reg, num_args
	OpRegReturn: {"OpRegReturn", []int{1}},    // return_reg

	// Register closure and function operations
	OpRegClosure:  {"OpRegClosure", []int{1, 2, 1, 1}}, // dst, func_idx(16-bit), num_free, start_reg
	OpRegLoadFunc: {"OpRegLoadFunc", []int{1, 2}},      // dst, func_idx

	// Register collection operations
	OpRegArray:    {"OpRegArray", []int{1, 1, 1}},    // dst, start_reg, count
	OpRegMap:      {"OpRegMap", []int{1, 1, 1}},      // dst, start_reg, count
	OpRegIndex:    {"OpRegIndex", []int{1, 1, 1}},    // dst, obj, key
	OpRegSetIndex: {"OpRegSetIndex", []int{1, 1, 1}}, // obj, key, val
	OpRegPush:     {"OpRegPush", []int{1}},           // src
	OpRegPop:      {"OpRegPop", []int{1}},            // dst

	// Register method/field operations
	OpRegGetMethod:  {"OpRegGetMethod", []int{1, 1, 2}},  // dst, obj, name_idx
	OpRegCallMethod: {"OpRegCallMethod", []int{1, 2, 1}}, // obj, name_idx, num_args
	OpRegGetField:   {"OpRegGetField", []int{1, 1, 2}},   // dst, obj, name_idx
	OpRegSetField:   {"OpRegSetField", []int{1, 1, 2}},   // obj, val, name_idx

	// Register class operations
	OpRegClass: {"OpRegClass", []int{1, 2}},     // dst, name_idx
	OpRegNew:   {"OpRegNew", []int{1, 1, 1}},    // dst, class_reg, num_args

	// Register built-in call
	OpRegBuiltin: {"OpRegBuiltin", []int{1, 1}}, // builtin_idx, num_args

	// Register literals
	OpRegNull:  {"OpRegNull", []int{1}},  // dst
	OpRegTrue:  {"OpRegTrue", []int{1}},  // dst
	OpRegFalse: {"OpRegFalse", []int{1}}, // dst

	// Register exception handling
	OpRegThrow:       {"OpRegThrow", []int{1}},           // src
	OpRegPushHandler: {"OpRegPushHandler", []int{2, 2}},  // catch_addr, finally_addr
	OpRegPopHandler:  {"OpRegPopHandler", []int{}},       // no operands

	// Register superinstructions
	OpRegAddConst: {"OpRegAddConst", []int{1, 1, 2}}, // dst, src, const_idx
	OpRegSubConst: {"OpRegSubConst", []int{1, 1, 2}}, // dst, src, const_idx
	OpRegMulConst: {"OpRegMulConst", []int{1, 1, 2}}, // dst, src, const_idx
	OpRegIncLocal: {"OpRegIncLocal", []int{1}},       // local_idx
	OpRegDecLocal: {"OpRegDecLocal", []int{1}},       // local_idx

	// Register module operations
	OpRegLoadModule: {"OpRegLoadModule", []int{1, 2}}, // dst, const_idx
	OpRegGetExport:  {"OpRegGetExport", []int{1, 1, 2}}, // dst, module_reg, name_idx
	OpRegSetExport:  {"OpRegSetExport", []int{1, 2}}, // src, name_idx
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

// ============================================================================
// Register VM Constants
// ============================================================================

const (
	// NumRegisters is the number of registers per frame
	NumRegisters = 256

	// NumArgRegisters is the number of argument registers (R0-R7)
	NumArgRegisters = 8

	// ReturnRegister is the register index for return values
	ReturnRegister = 255

	// FirstLocalRegister is the first register available for local variables
	// R0-R7 are reserved for arguments
	FirstLocalRegister = 8
)

// IsRegisterOpcode returns true if the opcode is a register-based operation
func IsRegisterOpcode(op Opcode) bool {
	return op >= OpRegAdd && op <= OpRegDecLocal
}

// MakeRegInstruction creates a fixed 4-byte register instruction
// Format: opcode(1) | dst(1) | src1(1) | src2(1)
func MakeRegInstruction(op Opcode, dst, src1, src2 int) []byte {
	return []byte{byte(op), byte(dst), byte(src1), byte(src2)}
}

// MakeRegInstruction1 creates a register instruction with 1 operand
func MakeRegInstruction1(op Opcode, reg int) []byte {
	return []byte{byte(op), byte(reg)}
}

// MakeRegInstruction2 creates a register instruction with 2 operands
func MakeRegInstruction2(op Opcode, reg1, reg2 int) []byte {
	return []byte{byte(op), byte(reg1), byte(reg2)}
}

// MakeRegInstructionConst creates a register instruction with a 16-bit constant index
// Format: opcode(1) | reg(1) | const_idx(2)
func MakeRegInstructionConst(op Opcode, reg, constIdx int) []byte {
	return []byte{
		byte(op),
		byte(reg),
		byte(constIdx >> 8),
		byte(constIdx),
	}
}

// MakeRegJump creates a jump instruction with 16-bit offset
// Format: opcode(1) | unused(1) | offset(2)
func MakeRegJump(op Opcode, offset int) []byte {
	return []byte{
		byte(op),
		0, // unused byte for alignment
		byte(offset >> 8),
		byte(offset),
	}
}

// MakeRegJumpCond creates a conditional jump instruction
// Format: opcode(1) | cond_reg(1) | offset(2)
func MakeRegJumpCond(op Opcode, condReg, offset int) []byte {
	return []byte{
		byte(op),
		byte(condReg),
		byte(offset >> 8),
		byte(offset),
	}
}

// DecodeRegInstruction decodes a fixed 4-byte register instruction
func DecodeRegInstruction(ins []byte) (op Opcode, dst, src1, src2 byte) {
	return Opcode(ins[0]), ins[1], ins[2], ins[3]
}

// DecodeRegInstructionConst decodes a register instruction with 16-bit constant
func DecodeRegInstructionConst(ins []byte) (op Opcode, reg byte, constIdx int) {
	return Opcode(ins[0]), ins[1], int(ins[2])<<8 | int(ins[3])
}

// DecodeRegJump decodes a jump instruction
func DecodeRegJump(ins []byte) (op Opcode, offset int) {
	return Opcode(ins[0]), int(ins[2])<<8 | int(ins[3])
}

// DecodeRegJumpCond decodes a conditional jump instruction
func DecodeRegJumpCond(ins []byte) (op Opcode, condReg byte, offset int) {
	return Opcode(ins[0]), ins[1], int(ins[2])<<8 | int(ins[3])
}

// GetDefinitions returns the opcode definitions map
func GetDefinitions() map[Opcode]*Definition {
	return definitions
}
