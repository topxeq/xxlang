// pkg/vm/vm.go
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Constants
const (
	GlobalsSize = 65536
	MaxFrames   = 1024
)

// VM is the virtual machine that executes bytecode
type VM struct {
	constants   []objects.Object
	stack       *Stack
	frames      []*Frame
	frameIndex  int
	globals     []objects.Object
}

// New creates a new VM with the given bytecode
func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:  bytecode.Constants,
		stack:      NewStack(),
		frames:     frames,
		frameIndex: 1,
		globals:    make([]objects.Object, GlobalsSize),
	}
}

// NewWithGlobalsStore creates a new VM with a custom globals store
func NewWithGlobalsStore(bytecode *compiler.Bytecode, globals []objects.Object) *VM {
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:  bytecode.Constants,
		stack:      NewStack(),
		frames:     frames,
		frameIndex: 1,
		globals:    globals,
	}
}

// StackTop returns the top of the stack
func (vm *VM) StackTop() objects.Object {
	if vm.stack.Len() == 0 {
		return nil
	}
	return vm.stack.Top()
}

// LastPopped returns the last popped element from the stack
func (vm *VM) LastPopped() objects.Object {
	return vm.stack.LastPopped()
}

// Globals returns the globals array
func (vm *VM) Globals() []objects.Object {
	return vm.globals
}

// Run executes the bytecode
func (vm *VM) Run() error {
	for vm.currentFrame().IP < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().IP++

		op := compiler.Opcode(vm.currentFrame().Instructions()[vm.currentFrame().IP])

		switch op {
		case compiler.OpConstant:
			if err := vm.executeConstant(); err != nil {
				return err
			}

		case compiler.OpPop:
			vm.stack.Pop()

		case compiler.OpDup:
			val := vm.stack.Pop()
			vm.stack.Push(val)
			vm.stack.Push(val)

		case compiler.OpAdd:
			if err := vm.executeBinaryOp(op); err != nil {
				return err
			}

		case compiler.OpSub:
			if err := vm.executeBinaryOp(op); err != nil {
				return err
			}

		case compiler.OpMul:
			if err := vm.executeBinaryOp(op); err != nil {
				return err
			}

		case compiler.OpDiv:
			if err := vm.executeBinaryOp(op); err != nil {
				return err
			}

		case compiler.OpMod:
			if err := vm.executeBinaryOp(op); err != nil {
				return err
			}

		case compiler.OpNeg:
			if err := vm.executeNeg(); err != nil {
				return err
			}

		case compiler.OpEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpNotEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpLess:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpGreater:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpLessEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpGreaterEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case compiler.OpAnd:
			if err := vm.executeLogicalAnd(); err != nil {
				return err
			}

		case compiler.OpOr:
			if err := vm.executeLogicalOr(); err != nil {
				return err
			}

		case compiler.OpNot:
			if err := vm.executeNot(); err != nil {
				return err
			}

		case compiler.OpGetLocal:
			if err := vm.executeGetLocal(); err != nil {
				return err
			}

		case compiler.OpSetLocal:
			if err := vm.executeSetLocal(); err != nil {
				return err
			}

		case compiler.OpGetGlobal:
			if err := vm.executeGetGlobal(); err != nil {
				return err
			}

		case compiler.OpSetGlobal:
			if err := vm.executeSetGlobal(); err != nil {
				return err
			}

		case compiler.OpJump:
			if err := vm.executeJump(); err != nil {
				return err
			}

		case compiler.OpJumpIfFalse:
			if err := vm.executeJumpIfFalse(); err != nil {
				return err
			}

		case compiler.OpJumpIfTrue:
			if err := vm.executeJumpIfTrue(); err != nil {
				return err
			}

		case compiler.OpCall:
			if err := vm.executeCall(); err != nil {
				return err
			}

		case compiler.OpTailCall:
			if err := vm.executeTailCall(); err != nil {
				return err
			}

		case compiler.OpReturn:
			if err := vm.executeReturn(); err != nil {
				return err
			}

		case compiler.OpClosure:
			if err := vm.executeClosure(); err != nil {
				return err
			}

		case compiler.OpGetFree:
			if err := vm.executeGetFree(); err != nil {
				return err
			}

		case compiler.OpSetFree:
			if err := vm.executeSetFree(); err != nil {
				return err
			}

		case compiler.OpArray:
			if err := vm.executeArray(); err != nil {
				return err
			}

		case compiler.OpMap:
			if err := vm.executeMap(); err != nil {
				return err
			}

		case compiler.OpIndex:
			if err := vm.executeIndex(); err != nil {
				return err
			}

		case compiler.OpSetIndex:
			if err := vm.executeSetIndex(); err != nil {
				return err
			}

		case compiler.OpBuiltin:
			if err := vm.executeBuiltin(); err != nil {
				return err
			}

		case compiler.OpNull:
			vm.stack.Push(objects.NULL)

		case compiler.OpTrue:
			vm.stack.Push(objects.TRUE)

		case compiler.OpFalse:
			vm.stack.Push(objects.FALSE)

		case compiler.OpBreak, compiler.OpContinue:
			// These are handled during compilation, ignore at runtime

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}

	return nil
}

// Helper methods

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	if vm.frameIndex >= MaxFrames {
		panic("frame overflow")
	}
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) readUint16() uint16 {
	ins := vm.currentFrame().Instructions()
	ip := vm.currentFrame().IP
	return uint16(ins[ip+1])<<8 | uint16(ins[ip+2])
}

func (vm *VM) readUint8() uint8 {
	ins := vm.currentFrame().Instructions()
	ip := vm.currentFrame().IP
	return ins[ip+1]
}

// Opcode implementations

func (vm *VM) executeConstant() error {
	constIndex := vm.readUint16()
	vm.currentFrame().IP += 2

	if int(constIndex) >= len(vm.constants) {
		return fmt.Errorf("constant index out of range: %d", constIndex)
	}

	vm.stack.Push(vm.constants[constIndex])
	return nil
}

func (vm *VM) executeBinaryOp(op compiler.Opcode) error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	// Check for mixed int/float operations
	leftIsInt, rightIsInt := isInt(left), isInt(right)
	leftIsFloat, rightIsFloat := isFloat(left), isFloat(right)

	// String concatenation
	if op == compiler.OpAdd {
		if leftStr, ok := left.(*objects.String); ok {
			if rightStr, ok := right.(*objects.String); ok {
				vm.stack.Push(&objects.String{Value: leftStr.Value + rightStr.Value})
				return nil
			}
		}
	}

	// Mixed int/float - promote to float
	if (leftIsInt && rightIsFloat) || (leftIsFloat && rightIsInt) {
		return vm.executeFloatBinaryOp(left, right, op)
	}

	// Integer arithmetic
	if leftIsInt && rightIsInt {
		return vm.executeIntBinaryOp(left, right, op)
	}

	// Float arithmetic
	if leftIsFloat && rightIsFloat {
		return vm.executeFloatBinaryOp(left, right, op)
	}

	return fmt.Errorf("type mismatch in binary operation: %s %d %s", left.Type(), op, right.Type())
}

func (vm *VM) executeIntBinaryOp(left, right objects.Object, op compiler.Opcode) error {
	leftVal := left.(*objects.Int).Value
	rightVal := right.(*objects.Int).Value

	var result int64
	switch op {
	case compiler.OpAdd:
		result = leftVal + rightVal
	case compiler.OpSub:
		result = leftVal - rightVal
	case compiler.OpMul:
		result = leftVal * rightVal
	case compiler.OpDiv:
		if rightVal == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftVal / rightVal
	case compiler.OpMod:
		if rightVal == 0 {
			return fmt.Errorf("modulo by zero")
		}
		result = leftVal % rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	vm.stack.Push(&objects.Int{Value: result})
	return nil
}

func (vm *VM) executeFloatBinaryOp(left, right objects.Object, op compiler.Opcode) error {
	var leftVal, rightVal float64

	if leftInt, ok := left.(*objects.Int); ok {
		leftVal = float64(leftInt.Value)
	} else {
		leftVal = left.(*objects.Float).Value
	}

	if rightInt, ok := right.(*objects.Int); ok {
		rightVal = float64(rightInt.Value)
	} else {
		rightVal = right.(*objects.Float).Value
	}

	var result float64
	switch op {
	case compiler.OpAdd:
		result = leftVal + rightVal
	case compiler.OpSub:
		result = leftVal - rightVal
	case compiler.OpMul:
		result = leftVal * rightVal
	case compiler.OpDiv:
		if rightVal == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftVal / rightVal
	case compiler.OpMod:
		if rightVal == 0 {
			return fmt.Errorf("modulo by zero")
		}
		result = float64(int64(leftVal) % int64(rightVal))
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	vm.stack.Push(&objects.Float{Value: result})
	return nil
}

func (vm *VM) executeNeg() error {
	operand := vm.stack.Pop()

	if isInt(operand) {
		val := operand.(*objects.Int).Value
		vm.stack.Push(&objects.Int{Value: -val})
		return nil
	}

	if isFloat(operand) {
		val := operand.(*objects.Float).Value
		vm.stack.Push(&objects.Float{Value: -val})
		return nil
	}

	return fmt.Errorf("negation not supported for type: %s", operand.Type())
}

func (vm *VM) executeComparison(op compiler.Opcode) error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	// Handle equality for all types
	if op == compiler.OpEqual || op == compiler.OpNotEqual {
		return vm.executeEquality(left, right, op)
	}

	// Numeric comparisons
	if (isInt(left) || isFloat(left)) && (isInt(right) || isFloat(right)) {
		return vm.executeNumericComparison(left, right, op)
	}

	// String comparisons
	if isString(left) && isString(right) {
		return vm.executeStringComparison(left, right, op)
	}

	return fmt.Errorf("comparison not supported for types: %s and %s", left.Type(), right.Type())
}

func (vm *VM) executeEquality(left, right objects.Object, op compiler.Opcode) error {
	var result bool

	switch op {
	case compiler.OpEqual:
		result = objectsEqual(left, right)
	case compiler.OpNotEqual:
		result = !objectsEqual(left, right)
	}

	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) executeNumericComparison(left, right objects.Object, op compiler.Opcode) error {
	var leftVal, rightVal float64

	if isInt(left) {
		leftVal = float64(left.(*objects.Int).Value)
	} else {
		leftVal = left.(*objects.Float).Value
	}

	if isInt(right) {
		rightVal = float64(right.(*objects.Int).Value)
	} else {
		rightVal = right.(*objects.Float).Value
	}

	var result bool
	switch op {
	case compiler.OpLess:
		result = leftVal < rightVal
	case compiler.OpGreater:
		result = leftVal > rightVal
	case compiler.OpLessEqual:
		result = leftVal <= rightVal
	case compiler.OpGreaterEqual:
		result = leftVal >= rightVal
	}

	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) executeStringComparison(left, right objects.Object, op compiler.Opcode) error {
	leftVal := left.(*objects.String).Value
	rightVal := right.(*objects.String).Value

	var result bool
	switch op {
	case compiler.OpLess:
		result = leftVal < rightVal
	case compiler.OpGreater:
		result = leftVal > rightVal
	case compiler.OpLessEqual:
		result = leftVal <= rightVal
	case compiler.OpGreaterEqual:
		result = leftVal >= rightVal
	}

	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) executeLogicalAnd() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	if !objects.IsTruthy(left) {
		vm.stack.Push(objects.FALSE)
	} else {
		if objects.IsTruthy(right) {
			vm.stack.Push(objects.TRUE)
		} else {
			vm.stack.Push(objects.FALSE)
		}
	}
	return nil
}

func (vm *VM) executeLogicalOr() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	if objects.IsTruthy(left) {
		vm.stack.Push(objects.TRUE)
	} else {
		if objects.IsTruthy(right) {
			vm.stack.Push(objects.TRUE)
		} else {
			vm.stack.Push(objects.FALSE)
		}
	}
	return nil
}

func (vm *VM) executeNot() error {
	operand := vm.stack.Pop()

	if objects.IsTruthy(operand) {
		vm.stack.Push(objects.FALSE)
	} else {
		vm.stack.Push(objects.TRUE)
	}
	return nil
}

func (vm *VM) executeGetLocal() error {
	localIndex := vm.readUint8()
	vm.currentFrame().IP++

	frame := vm.currentFrame()
	vm.stack.Push(frame.Locals[localIndex])
	return nil
}

func (vm *VM) executeSetLocal() error {
	localIndex := vm.readUint8()
	vm.currentFrame().IP++

	frame := vm.currentFrame()
	value := vm.stack.Pop()
	frame.Locals[localIndex] = value
	return nil
}

func (vm *VM) executeGetGlobal() error {
	globalIndex := vm.readUint16()
	vm.currentFrame().IP += 2

	vm.stack.Push(vm.globals[globalIndex])
	return nil
}

func (vm *VM) executeSetGlobal() error {
	globalIndex := vm.readUint16()
	vm.currentFrame().IP += 2

	value := vm.stack.Pop()
	vm.globals[globalIndex] = value
	return nil
}

func (vm *VM) executeJump() error {
	pos := vm.readUint16()
	vm.currentFrame().IP = int(pos) - 1 // -1 because the loop will increment
	return nil
}

func (vm *VM) executeJumpIfFalse() error {
	pos := vm.readUint16()
	vm.currentFrame().IP += 2

	condition := vm.stack.Pop()

	if !objects.IsTruthy(condition) {
		vm.currentFrame().IP = int(pos) - 1 // -1 because the loop will increment
	}
	return nil
}

func (vm *VM) executeJumpIfTrue() error {
	pos := vm.readUint16()
	vm.currentFrame().IP += 2

	condition := vm.stack.Pop()

	if objects.IsTruthy(condition) {
		vm.currentFrame().IP = int(pos) - 1 // -1 because the loop will increment
	}
	return nil
}

func (vm *VM) executeCall() error {
	numArgs := int(vm.readUint8())
	vm.currentFrame().IP++

	// Callee is below the arguments on the stack
	// Stack: [callee, arg1, arg2, ...] (callee at bottom, args on top)
	// Peek(numArgs) gets the callee which is numArgs positions below the top
	callee := vm.stack.Peek(numArgs)

	switch fn := callee.(type) {
	case *compiler.CompiledFunction:
		return vm.callFunction(fn, numArgs, nil)
	case *Closure:
		return vm.callFunction(fn.Fn, numArgs, fn.FreeVars)
	case *objects.Builtin:
		return vm.callBuiltin(fn, numArgs)
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
}

func (vm *VM) callFunction(fn *compiler.CompiledFunction, numArgs int, freeVars []objects.Object) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	// Create new frame
	frame := NewFrame(fn, vm.stack.Len()-numArgs-1)

	// Pop arguments in reverse order and store in locals
	// Stack before: [callee, arg1, arg2, ...] (callee at bottom, last arg at top)
	// Arguments go to locals[0], locals[1], ...
	for i := numArgs - 1; i >= 0; i-- {
		frame.Locals[i] = vm.stack.Pop()
	}

	// Pop callee from stack (we already have fn from executeCall)
	vm.stack.Pop()

	// Store reference to closure's free variables (not a copy!)
	// This allows modifications to persist across calls
	if freeVars != nil {
		frame.FreeVars = freeVars
	}

	vm.pushFrame(frame)
	return nil
}

func (vm *VM) callBuiltin(fn *objects.Builtin, numArgs int) error {
	args := make([]objects.Object, numArgs)
	for i := numArgs - 1; i >= 0; i-- {
		args[i] = vm.stack.Pop()
	}

	// Pop the callee from the stack
	vm.stack.Pop()

	result := fn.Fn(args...)
	vm.stack.Push(result)
	return nil
}

// executeTailCall implements tail call optimization
// Instead of creating a new frame, it reuses the current frame
func (vm *VM) executeTailCall() error {
	numArgs := int(vm.readUint8())
	vm.currentFrame().IP++

	// Callee is below the arguments on the stack
	callee := vm.stack.Peek(numArgs)

	switch fn := callee.(type) {
	case *compiler.CompiledFunction:
		return vm.tailCallFunction(fn, numArgs, nil)
	case *Closure:
		return vm.tailCallFunction(fn.Fn, numArgs, fn.FreeVars)
	case *objects.Builtin:
		// Builtins don't benefit from TCO, just use regular call
		return vm.callBuiltin(fn, numArgs)
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
}

// tailCallFunction reuses the current frame for a tail call
func (vm *VM) tailCallFunction(fn *compiler.CompiledFunction, numArgs int, freeVars []objects.Object) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	// Get current frame
	frame := vm.currentFrame()

	// Pop arguments in reverse order
	args := make([]objects.Object, numArgs)
	for i := numArgs - 1; i >= 0; i-- {
		args[i] = vm.stack.Pop()
	}

	// Pop callee from stack
	vm.stack.Pop()

	// Update locals with new arguments
	for i := 0; i < numArgs; i++ {
		frame.Locals[i] = args[i]
	}

	// Reset frame to new function
	frame.Fn = fn
	frame.IP = -1 // Will be incremented to 0 in the main loop
	frame.FreeVars = freeVars

	return nil
}

func (vm *VM) executeReturn() error {
	// Pop the return value
	result := vm.stack.Pop()

	// Pop the current frame
	frame := vm.popFrame()

	// Restore stack pointer (remove locals from stack)
	if frame.BasePointer > 0 {
		for vm.stack.Len() > frame.BasePointer {
			vm.stack.Pop()
		}
	}

	// Push the return value
	vm.stack.Push(result)
	return nil
}

func (vm *VM) executeArray() error {
	numElements := int(vm.readUint16())
	vm.currentFrame().IP += 2

	elements := make([]objects.Object, numElements)
	for i := numElements - 1; i >= 0; i-- {
		elements[i] = vm.stack.Pop()
	}

	vm.stack.Push(&objects.Array{Elements: elements})
	return nil
}

func (vm *VM) executeMap() error {
	numPairs := int(vm.readUint16())
	vm.currentFrame().IP += 2

	pairs := make(map[objects.HashKey]objects.MapPair)

	for i := 0; i < numPairs; i++ {
		value := vm.stack.Pop()
		key := vm.stack.Pop()

		hashKey := key.HashKey()
		pairs[hashKey] = objects.MapPair{Key: key, Value: value}
	}

	vm.stack.Push(&objects.Map{Pairs: pairs})
	return nil
}

func (vm *VM) executeIndex() error {
	index := vm.stack.Pop()
	left := vm.stack.Pop()

	switch obj := left.(type) {
	case *objects.Array:
		return vm.executeArrayIndex(obj, index)
	case *objects.Map:
		return vm.executeMapIndex(obj, index)
	case *objects.String:
		return vm.executeStringIndex(obj, index)
	default:
		return fmt.Errorf("index operator not supported for type: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(arr *objects.Array, index objects.Object) error {
	intIndex, ok := index.(*objects.Int)
	if !ok {
		return fmt.Errorf("array index must be integer, got: %s", index.Type())
	}

	idx := int(intIndex.Value)
	if idx < 0 || idx >= len(arr.Elements) {
		vm.stack.Push(objects.NULL)
		return nil
	}

	vm.stack.Push(arr.Elements[idx])
	return nil
}

func (vm *VM) executeMapIndex(m *objects.Map, index objects.Object) error {
	pair, ok := m.Pairs[index.HashKey()]
	if !ok {
		vm.stack.Push(objects.NULL)
		return nil
	}

	vm.stack.Push(pair.Value)
	return nil
}

func (vm *VM) executeStringIndex(str *objects.String, index objects.Object) error {
	intIndex, ok := index.(*objects.Int)
	if !ok {
		return fmt.Errorf("string index must be integer, got: %s", index.Type())
	}

	idx := int(intIndex.Value)
	if idx < 0 || idx >= len(str.Value) {
		vm.stack.Push(objects.NULL)
		return nil
	}

	vm.stack.Push(&objects.String{Value: string(str.Value[idx])})
	return nil
}

func (vm *VM) executeSetIndex() error {
	// Stack order (from compilation):
	// 1. value is pushed first (bottom)
	// 2. arr is pushed second
	// 3. index is pushed last (top)
	// So we pop: index, then arr, then value
	index := vm.stack.Pop()
	left := vm.stack.Pop()
	value := vm.stack.Pop()

	switch obj := left.(type) {
	case *objects.Array:
		return vm.executeArraySetIndex(obj, index, value)
	case *objects.Map:
		return vm.executeMapSetIndex(obj, index, value)
	default:
		return fmt.Errorf("set index operator not supported for type: %s", left.Type())
	}
}

func (vm *VM) executeArraySetIndex(arr *objects.Array, index, value objects.Object) error {
	intIndex, ok := index.(*objects.Int)
	if !ok {
		return fmt.Errorf("array index must be integer, got: %s", index.Type())
	}

	idx := int(intIndex.Value)
	if idx < 0 || idx >= len(arr.Elements) {
		return fmt.Errorf("array index out of bounds: %d", idx)
	}

	arr.Elements[idx] = value
	vm.stack.Push(value)
	return nil
}

func (vm *VM) executeMapSetIndex(m *objects.Map, index, value objects.Object) error {
	hashKey := index.HashKey()
	m.Pairs[hashKey] = objects.MapPair{Key: index, Value: value}
	vm.stack.Push(value)
	return nil
}

func (vm *VM) executeBuiltin() error {
	builtinIndex := vm.readUint8()
	vm.currentFrame().IP++

	// Get the builtin function
	builtin := getBuiltin(int(builtinIndex))
	if builtin == nil {
		return fmt.Errorf("builtin function not found: %d", builtinIndex)
	}

	// Push the builtin onto the stack - it will be called by OpCall
	vm.stack.Push(builtin)
	return nil
}

// Helper functions

func isInt(obj objects.Object) bool {
	_, ok := obj.(*objects.Int)
	return ok
}

func isFloat(obj objects.Object) bool {
	_, ok := obj.(*objects.Float)
	return ok
}

func isString(obj objects.Object) bool {
	_, ok := obj.(*objects.String)
	return ok
}

func objectsEqual(a, b objects.Object) bool {
	// Direct equality check (for NULL, TRUE, FALSE singletons)
	if a == b {
		return true
	}

	// Type check
	if a.Type() != b.Type() {
		return false
	}

	// Value comparison
	switch a := a.(type) {
	case *objects.Int:
		return a.Value == b.(*objects.Int).Value
	case *objects.Float:
		return a.Value == b.(*objects.Float).Value
	case *objects.String:
		return a.Value == b.(*objects.String).Value
	case *objects.Bool:
		return a.Value == b.(*objects.Bool).Value
	default:
		return false
	}
}

// getBuiltin returns the builtin function by index
func getBuiltin(index int) *objects.Builtin {
	builtins := []*objects.Builtin{
		objects.Builtins["len"],           // 0
		objects.Builtins["print"],         // 1
		objects.Builtins["println"],       // 2
		objects.Builtins["typeOf"],        // 3
		objects.Builtins["substr"],        // 4
		objects.Builtins["split"],         // 5
		objects.Builtins["join"],          // 6
		objects.Builtins["trim"],          // 7
		objects.Builtins["upper"],         // 8
		objects.Builtins["lower"],         // 9
		objects.Builtins["containsStr"],   // 10
		objects.Builtins["replace"],       // 11
		objects.Builtins["startsWith"],    // 12
		objects.Builtins["endsWith"],      // 13
		objects.Builtins["abs"],           // 14
		objects.Builtins["floor"],         // 15
		objects.Builtins["ceil"],          // 16
		objects.Builtins["sqrt"],          // 17
		objects.Builtins["pow"],           // 18
		objects.Builtins["min"],           // 19
		objects.Builtins["max"],           // 20
		objects.Builtins["int"],           // 21
		objects.Builtins["float"],         // 22
		objects.Builtins["string"],        // 23
		objects.Builtins["push"],          // 24
		objects.Builtins["pop"],           // 25
		objects.Builtins["first"],         // 26
		objects.Builtins["last"],          // 27
		objects.Builtins["rest"],          // 28
		objects.Builtins["concat"],        // 29
		objects.Builtins["indexOf"],       // 30
		objects.Builtins["containsArr"],   // 31
		objects.Builtins["keys"],          // 32
		objects.Builtins["values"],        // 33
		objects.Builtins["hasKey"],        // 34
		objects.Builtins["delete"],        // 35
		objects.Builtins["range"],         // 36
		objects.Builtins["sort"],          // 37
		objects.Builtins["sum"],           // 38
		objects.Builtins["avg"],           // 39
		objects.Builtins["reverse"],       // 40
	}

	if index < 0 || index >= len(builtins) {
		return nil
	}

	return builtins[index]
}
