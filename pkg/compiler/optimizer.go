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
	flags    OptimizationFlags
}

// NewOptimizer creates a new optimizer for given bytecode
func NewOptimizer(bytecode *Bytecode) *Optimizer {
	return &Optimizer{
		bytecode: bytecode,
		flags:    DefaultOptimizations(),
	}
}

// NewOptimizerWithFlags creates an optimizer with custom flags
func NewOptimizerWithFlags(bytecode *Bytecode, flags OptimizationFlags) *Optimizer {
	return &Optimizer{
		bytecode: bytecode,
		flags:    flags,
	}
}

// Optimize runs all optimization passes
func (o *Optimizer) Optimize() *Bytecode {
	// FoldConstants is part of BytecodeOptimizer
	if o.flags.BytecodeOptimizer {
		o.bytecode = o.FoldConstants()
	}
	if o.flags.InlineFunctions {
		o.bytecode = o.InlineFunctions()
	}
	if o.flags.Superinstructions {
		o.bytecode = o.GenerateSuperinstructions()
	}
	if o.flags.TypeSpecialization {
		o.bytecode = o.GenerateTypeSpecializations()
	}
	return o.bytecode
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

// GenerateTypeSpecializations creates type-specialized instructions
// for common patterns that can be optimized
func (o *Optimizer) GenerateTypeSpecializations() *Bytecode {
	instructions := o.bytecode.Instructions

	if len(instructions) < 5 {
		return o.bytecode
	}

	newInstructions := make([]byte, 0, len(instructions))
	constants := o.bytecode.Constants

	i := 0
	for i < len(instructions) {
		op := Opcode(instructions[i])

		// Pattern: OpGetLocal + OpConstant(1) + OpAdd + OpSetLocal
		// This is "local = local + constant" pattern (like loop counters)
		// Can be optimized to OpAddLocalConst
		if op == OpGetLocal && i+10 < len(instructions) {
			localIdx := int(instructions[i+1])

			// Check for OpConstant
			if Opcode(instructions[i+2]) == OpConstant {
				constIdx := int(instructions[i+3])<<8 | int(instructions[i+4])

				// Check if constant is integer
				if _, ok := constants[constIdx].(*objects.Int); ok {
					// Check for OpAdd
					if Opcode(instructions[i+5]) == OpAdd {
						// Check for OpSetLocal with same index
						if Opcode(instructions[i+6]) == OpSetLocal && int(instructions[i+7]) == localIdx {
							// Found pattern! Use OpAddLocalConst
							newInstructions = append(newInstructions, byte(OpAddLocalConst), byte(localIdx))
							newInstructions = append(newInstructions, byte(constIdx>>8), byte(constIdx))
							i += 8 // Skip the whole sequence
							continue
						}
					}
				}
			}
		}

		// Pattern: OpGetLocal + OpConstant(1) + OpAdd -> OpIncLocal (if constant is 1)
		// This is for expressions like "i + 1" without assignment
		if op == OpGetLocal && i+5 < len(instructions) {
			localIdx := int(instructions[i+1])

			if Opcode(instructions[i+2]) == OpConstant {
				constIdx := int(instructions[i+3])<<8 | int(instructions[i+4])

				if c, ok := constants[constIdx].(*objects.Int); ok && c.Value == 1 {
					if Opcode(instructions[i+5]) == OpAdd {
						// OpIncLocal increments and pushes the result
						newInstructions = append(newInstructions, byte(OpIncLocal), byte(localIdx))
						i += 6
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

// InlineFunctions inlines small functions at call sites to reduce call overhead
// Pattern: OpGetGlobal/OpConstant (callee), args..., OpCall -> inlined body
func (o *Optimizer) InlineFunctions() *Bytecode {
	instructions := o.bytecode.Instructions
	constants := o.bytecode.Constants

	if len(instructions) < 5 {
		return o.bytecode
	}

	// Build a map of inlineable functions from constant pool
	// Key: constant index, Value: function info
	inlineableConsts := make(map[int]*inlineableFunc)
	for i, c := range constants {
		if fn, ok := c.(*CompiledFunction); ok && fn.IsInlineable && len(fn.FreeVariables) == 0 {
			inlineableConsts[i] = &inlineableFunc{
				numParams: fn.NumParameters,
				body:      fn.InlineBody,
			}
		}
	}

	// Build a map of inlineable functions from globals
	// Key: global index, Value: function info
	inlineableGlobals := o.bytecode.InlineableGlobals
	if inlineableGlobals == nil {
		inlineableGlobals = make(map[int]*InlineableFuncInfo)
	}

	if len(inlineableConsts) == 0 && len(inlineableGlobals) == 0 {
		return o.bytecode
	}

	newInstructions := make([]byte, 0, len(instructions))

	i := 0
	for i < len(instructions) {
		op := Opcode(instructions[i])

		// Look for call pattern: OpGetGlobal/OpConstant (callee), followed by args, then OpCall
		var fn *inlineableFunc
		var calleePos int

		if op == OpConstant && i+3 < len(instructions) {
			// Anonymous function stored in constant pool
			constIdx := int(instructions[i+1])<<8 | int(instructions[i+2])
			if f, ok := inlineableConsts[constIdx]; ok {
				fn = f
				calleePos = i
			}
		} else if op == OpGetGlobal && i+3 < len(instructions) {
			// Named function stored in global
			globalIdx := int(instructions[i+1])<<8 | int(instructions[i+2])
			if info, ok := inlineableGlobals[globalIdx]; ok {
				fn = &inlineableFunc{
					numParams: info.NumParams,
					body:      info.Body,
				}
				calleePos = i
			}
		}

		if fn != nil && calleePos >= 0 {
			// Found potential inlineable call
			numArgs := fn.numParams

			// Find OpCall instruction
			callPos, found := o.findCallInstruction(instructions, calleePos, numArgs)
			if found {
				// Inline the function
				inlined := o.inlineCall(instructions, calleePos, callPos, fn)
				if inlined != nil {
					newInstructions = append(newInstructions, inlined...)
					i = callPos + 2 // Skip past OpCall (1 byte opcode + 1 byte numArgs)
					continue
				}
				// If inlining failed, fall through to copy instruction as-is
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

type inlineableFunc struct {
	numParams int
	body      []byte
}

// findCallInstruction finds the OpCall instruction after loading callee and args
// Uses stack effect tracking to properly handle complex argument expressions
func (o *Optimizer) findCallInstruction(instructions []byte, calleePos int, numArgs int) (int, bool) {
	// Skip callee load instruction
	def, err := Lookup(instructions[calleePos])
	if err != nil {
		return -1, false
	}
	pos := calleePos + 1
	for _, w := range def.OperandWidths {
		pos += w
	}

	// Track stack depth relative to callee position
	// We need to find when numArgs values are on stack, then OpCall
	// Note: callee is already on stack, we're tracking arguments pushed after it
	stackDepth := 0

	for pos < len(instructions) {
		op := Opcode(instructions[pos])
		def, err := Lookup(byte(op))
		if err != nil {
			return -1, false
		}

		// Calculate instruction length
		instrLen := 1
		for _, w := range def.OperandWidths {
			instrLen += w
		}

		// Check for OpCall BEFORE applying its stack effect
		// At this point, stackDepth should equal numArgs
		if op == OpCall {
			callNumArgs := int(instructions[pos+1])
			if callNumArgs == numArgs && stackDepth == numArgs {
				return pos, true
			}
		}

		// Track stack effect
		stackEffect := stackEffectOf(op, instructions, pos)
		stackDepth += stackEffect

		// If stack goes negative or too deep, something's wrong
		if stackDepth < 0 || stackDepth > numArgs+20 {
			return -1, false
		}

		pos += instrLen
	}

	return -1, false
}

// stackEffectOf returns the net stack effect of an instruction (positive = pushes, negative = pops)
func stackEffectOf(op Opcode, instructions []byte, pos int) int {
	switch op {
	// Instructions that push 1 value
	case OpConstant, OpGetLocal, OpGetGlobal, OpGetFree, OpTrue, OpFalse, OpNull,
		OpBuiltin, OpGetField, OpIndex, OpIndexSafe, OpDup:
		return 1

	// Superinstructions that push 1 value (pop 2, push 1 = net -1 each, but these compute and push)
	case OpConstantAdd, OpConstantSub, OpConstantMul,
		OpGetLocalAdd, OpGetLocalSub, OpGetLocalMul:
		return -1 // pop 2, push 1 = net -1

	// Array creates array from stack elements
	case OpArray:
		if pos+2 < len(instructions) {
			count := int(instructions[pos+1])<<8 | int(instructions[pos+2])
			return 1 - count // pops count elements, pushes array
		}
		return 0

	// Map creates map from key-value pairs
	case OpMap:
		if pos+2 < len(instructions) {
			count := int(instructions[pos+1])<<8 | int(instructions[pos+2]) // number of pairs
			return 1 - count*2 // pops count*2 elements, pushes map
		}
		return 0

	// Call pops callee + numArgs, pushes 1 result
	case OpCall:
		if pos+1 < len(instructions) {
			numArgs := int(instructions[pos+1])
			return -numArgs // pops args and callee (numArgs+1), pushes result (1): net = -numArgs
		}
		return 0

	// Method call pops receiver + args, pushes result
	case OpCallMethod:
		if pos+1 < len(instructions) {
			numArgs := int(instructions[pos+1])
			return -numArgs // pops args + receiver, pushes result
		}
		return 0

	// Tail call
	case OpTailCall:
		if pos+1 < len(instructions) {
			numArgs := int(instructions[pos+1])
			return -numArgs
		}
		return 0

	// Instructions that pop 1 value
	case OpPop, OpSetGlobal, OpSetLocal, OpSetFree, OpSetIndex, OpSetField:
		return -1

	// Binary ops pop 2, push 1 = net -1
	case OpAdd, OpSub, OpMul, OpDiv, OpMod:
		return -1

	// Comparison ops pop 2, push 1 = net -1
	case OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual:
		return -1

	// Logical ops
	case OpAnd, OpOr:
		return -1
	case OpNot:
		return 0 // pop 1, push 1

	// Unary minus
	case OpNeg:
		return 0 // pop 1, push 1

	// Jump instructions - no net effect (we're looking for call sites, not tracking across jumps)
	case OpJump, OpJumpIfFalse, OpJumpIfTrue, OpBreak, OpContinue, OpReturn:
		return 0

	// Closure pushes 1
	case OpClosure:
		return 1

	// New creates object (pops args based on count, pushes object)
	case OpNew:
		if pos+1 < len(instructions) {
			numArgs := int(instructions[pos+1])
			return 1 - numArgs // pops numArgs, pushes object
		}
		return 0

	// Increment/decrement local
	case OpIncLocal, OpDecLocal, OpAddLocalConst:
		return 0 // modify local, no stack effect

	// Push/pop scope - no stack effect
	case OpPushScope, OpPopScope:
		return 0

	// GetMethod pushes method
	case OpGetMethod:
		return 1

	// Class pushes class object
	case OpClass:
		return 1

	// Module operations
	case OpLoadModule, OpModule:
		return 1
	case OpGetExport, OpSetExport:
		return 0

	// DefineLocal stores TOS
	case OpDefineLocal:
		return -1

	default:
		// Conservative: assume no stack effect
		return 0
	}
}

// inlineCall generates inlined bytecode for a function call
// Returns nil if inlining is not possible
func (o *Optimizer) inlineCall(instructions []byte, calleePos int, callPos int, fn *inlineableFunc) []byte {
	// Transform the function body
	// For simple bodies, we need to remap OpGetLocal N to access the Nth argument
	transformedBody := o.transformInlineBody(fn.body, fn.numParams)
	if transformedBody == nil {
		// Inlining not possible for this body pattern
		return nil
	}

	// Extract argument instructions (between callee and OpCall)
	argStart := calleePos
	def, _ := Lookup(instructions[calleePos])
	argStart += 1
	for _, w := range def.OperandWidths {
		argStart += w
	}

	// Copy argument instructions
	result := make([]byte, 0)
	pos := argStart
	for pos < callPos {
		op := Opcode(instructions[pos])
		def, err := Lookup(byte(op))
		if err != nil {
			break
		}

		instrLen := 1
		for _, w := range def.OperandWidths {
			instrLen += w
		}

		result = append(result, instructions[pos:pos+instrLen]...)
		pos += instrLen
	}

	// Append the transformed body
	result = append(result, transformedBody...)

	return result
}

// transformInlineBody transforms function body to work with args on stack
func (o *Optimizer) transformInlineBody(body []byte, numParams int) []byte {
	if len(body) == 0 {
		return body
	}

	// Pattern 1: Single parameter returned directly
	// Body: OpGetLocal 0 -> args are already on stack, nothing needed
	if numParams == 1 && len(body) == 2 {
		if Opcode(body[0]) == OpGetLocal && body[1] == 0 {
			// The argument is already on stack, nothing to do
			return []byte{}
		}
	}

	// Pattern 2: Two parameters with binary operation
	// Body: OpGetLocal 0, OpGetLocal 1, OpBinaryOp
	if o.isSimpleBinaryBody(body, numParams) {
		// The body is: OpGetLocal 0, OpGetLocal 1, ..., OpBinaryOp
		// Args are on stack, so we just return the binary op
		lastOp := body[len(body)-1]
		return []byte{lastOp}
	}

	// Pattern 3: Single parameter with unary operation
	// Body: OpGetLocal 0, OpNeg
	if numParams == 1 && len(body) == 3 {
		if Opcode(body[0]) == OpGetLocal && body[1] == 0 {
			// OpGetLocal 0, <unary op>
			unaryOp := Opcode(body[2])
			switch unaryOp {
			case OpNeg, OpNot:
				// Arg is on stack, just emit the unary op
				return []byte{byte(unaryOp)}
			}
		}
	}

	// Pattern 3.5: Single parameter used twice with binary operation (e.g., x * x, x + x)
	// Body: OpGetLocal 0, OpGetLocal 0, OpBinaryOp
	// Transformation: OpDup, OpBinaryOp (duplicate the arg, then operate)
	if numParams == 1 && len(body) == 5 {
		if Opcode(body[0]) == OpGetLocal && body[1] == 0 &&
			Opcode(body[2]) == OpGetLocal && body[3] == 0 {
			binaryOp := Opcode(body[4])
			switch binaryOp {
			case OpAdd, OpSub, OpMul, OpDiv, OpMod,
				OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual:
				// Emit: OpDup, OpBinaryOp
				return []byte{byte(OpDup), byte(binaryOp)}
			}
		}
	}

	// Pattern 4: Two parameters with comparison (same as binary)
	// Already handled by isSimpleBinaryBody

	// Pattern 5: Three parameters with chained binary operations
	// Body: OpGetLocal 0, OpGetLocal 1, OpAdd, OpGetLocal 2, OpAdd
	if transformed := o.transformThreeParamBinary(body, numParams); transformed != nil {
		return transformed
	}

	// Pattern 6: Parameter with constant operation
	// Body: OpGetLocal 0, OpConstant N, OpBinaryOp
	if transformed := o.transformParamConstBinary(body, numParams); transformed != nil {
		return transformed
	}

	// Pattern 7: General transformation for simple expressions
	// Transform OpGetLocal N to a stack-relative access
	if transformed := o.transformGeneralBody(body, numParams); transformed != nil {
		return transformed
	}

	// Return nil to signal that inlining is not possible.
	return nil
}

// transformThreeParamBinary handles patterns like: return a + b + c
// Body: OpGetLocal 0, OpGetLocal 1, OpAdd, OpGetLocal 2, OpAdd
func (o *Optimizer) transformThreeParamBinary(body []byte, numParams int) []byte {
	if numParams != 3 || len(body) != 9 {
		return nil
	}

	// Expected pattern: OpGetLocal 0, OpGetLocal 1, OpAdd, OpGetLocal 2, OpAdd
	if Opcode(body[0]) != OpGetLocal || body[1] != 0 {
		return nil
	}
	if Opcode(body[2]) != OpGetLocal || body[3] != 1 {
		return nil
	}
	if Opcode(body[4]) != OpAdd {
		return nil
	}
	if Opcode(body[5]) != OpGetLocal || body[6] != 2 {
		return nil
	}
	if Opcode(body[7]) != OpAdd {
		return nil
	}

	// Args are on stack: arg0, arg1, arg2 (arg2 on top)
	// We need: arg0 + arg1 + arg2
	// Stack has arg0, arg1, arg2 - we can just do OpAdd twice
	return []byte{byte(OpAdd), byte(OpAdd)}
}

// transformParamConstBinary handles patterns like: return a + 1 or return a * 2
// Body: OpGetLocal 0, OpConstant N, OpBinaryOp
func (o *Optimizer) transformParamConstBinary(body []byte, numParams int) []byte {
	if numParams != 1 || len(body) != 6 {
		return nil
	}

	// Pattern: OpGetLocal 0, OpConstant idx, OpBinaryOp
	if Opcode(body[0]) != OpGetLocal || body[1] != 0 {
		return nil
	}
	if Opcode(body[2]) != OpConstant {
		return nil
	}

	binaryOp := Opcode(body[5])
	switch binaryOp {
	case OpAdd, OpSub, OpMul, OpDiv, OpMod,
		OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual:
		// Arg is on stack, just emit: OpConstant idx, OpBinaryOp
		return body[2:]
	default:
		return nil
	}
}

// transformGeneralBody attempts to transform any simple expression body
// by replacing OpGetLocal N with the assumption that args are on stack
func (o *Optimizer) transformGeneralBody(body []byte, numParams int) []byte {
	// For a general body, we need to ensure that the parameter accesses
	// are in the right order for the stack.

	// Analyze the body: collect all OpGetLocal instructions and their indices
	// Check if the pattern is: all params accessed in order, then operations
	getLocalPattern := o.analyzeGetLocalPattern(body, numParams)
	if getLocalPattern == nil {
		return nil
	}

	// If params are accessed in sequential order (0, 1, 2, ..., n-1),
	// we can strip the OpGetLocal instructions because args are already on stack
	if getLocalPattern.isSequential {
		// Strip OpGetLocal instructions, keep operations
		return o.stripGetLocals(body, numParams)
	}

	return nil
}

type getLocalPattern struct {
	isSequential bool   // Params accessed in order: 0, 1, 2, ...
	indices      []int  // Order of parameter access
}

// analyzeGetLocalPattern analyzes the pattern of OpGetLocal in a body
func (o *Optimizer) analyzeGetLocalPattern(body []byte, numParams int) *getLocalPattern {
	pattern := &getLocalPattern{
		indices: make([]int, 0),
	}

	seen := make(map[int]bool)
	i := 0
	for i < len(body) {
		op := Opcode(body[i])

		if op == OpGetLocal {
			if i+1 >= len(body) {
				return nil
			}
			idx := int(body[i+1])
			if idx >= numParams {
				// Accessing a non-parameter local, can't inline simply
				return nil
			}
			if seen[idx] {
				// Parameter accessed multiple times, need more complex handling
				return nil
			}
			seen[idx] = true
			pattern.indices = append(pattern.indices, idx)
			i += 2
		} else {
			// Skip operands for other instructions
			def, err := Lookup(byte(op))
			if err != nil {
				i++
				continue
			}
			i++
			for _, w := range def.OperandWidths {
				i += w
			}
		}
	}

	// Check if sequential
	pattern.isSequential = true
	for i, idx := range pattern.indices {
		if idx != i {
			pattern.isSequential = false
			break
		}
	}

	return pattern
}

// stripGetLocals removes OpGetLocal instructions from body
// Assumes params are accessed in order 0, 1, 2, ...
func (o *Optimizer) stripGetLocals(body []byte, numParams int) []byte {
	result := make([]byte, 0, len(body)-2*numParams)
	i := 0
	for i < len(body) {
		op := Opcode(body[i])

		if op == OpGetLocal && i+1 < len(body) {
			// Skip OpGetLocal and its operand
			i += 2
			continue
		}

		// Copy other instructions
		def, err := Lookup(byte(op))
		if err != nil {
			result = append(result, body[i])
			i++
			continue
		}

		instrLen := 1
		for _, w := range def.OperandWidths {
			instrLen += w
		}

		result = append(result, body[i:i+instrLen]...)
		i += instrLen
	}

	return result
}
func (o *Optimizer) isSimpleBinaryBody(body []byte, numParams int) bool {
	if numParams != 2 {
		return false
	}

	// Expected: OpGetLocal 0, OpGetLocal 1, OpBinaryOp (5 bytes)
	if len(body) != 5 {
		return false
	}

	// Check OpGetLocal 0
	if Opcode(body[0]) != OpGetLocal || body[1] != 0 {
		return false
	}

	// Check OpGetLocal 1
	if Opcode(body[2]) != OpGetLocal || body[3] != 1 {
		return false
	}

	// Check binary op
	op := Opcode(body[4])
	switch op {
	case OpAdd, OpSub, OpMul, OpDiv, OpMod,
		OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual:
		return true
	default:
		return false
	}
}
