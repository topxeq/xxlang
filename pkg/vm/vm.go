// pkg/vm/vm.go
package vm

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Constants
const (
	GlobalsSize = 65536
	MaxFrames   = 1024
)

// InlineCacheEntry represents a cached method lookup
type InlineCacheEntry struct {
	Class  *objects.Class
	Method objects.Object
}

// InlineCacheSize is the size of the method cache
const InlineCacheSize = 256

// ExceptionHandler represents a try-catch-finally handler
type ExceptionHandler struct {
	catchAddr   int // Address to jump to for catch (0 if no catch)
	finallyAddr int // Address to jump to for finally (0 if no finally)
	frameIndex  int // Frame index when handler was pushed
	stackSize   int // Stack size when handler was pushed
}

// VM is the virtual machine that executes bytecode
type VM struct {
	constants        []objects.Object
	stack            *Stack
	frames           []*Frame
	frameIndex       int
	globals          []objects.Object
	loader           *module.Loader   // Module cache and cycle detection
	currentModule    *objects.Module  // Current module context for exports
	sourcePath       string           // Current source file path for imports
	currentInstance  *objects.Instance // Current instance for this binding
	pendingInstance  *objects.Instance // Instance to push after init returns
	initFrame        *Frame           // The init frame that should push pendingInstance
	sourceMap        *compiler.SourceMap // Source map for error reporting

	// Exception handling
	handlers []ExceptionHandler

	// Inline cache for method lookups
	methodCache [InlineCacheSize]InlineCacheEntry
	cacheHits   int
	cacheMisses int
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
		loader:     module.NewLoader(),
		sourceMap:  bytecode.SourceMap,
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
		loader:     module.NewLoader(),
		sourceMap:  bytecode.SourceMap,
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

// SetSourcePath sets the source file path for module resolution
func (vm *VM) SetSourcePath(path string) {
	vm.sourcePath = path
}

// SetLoader sets the module loader (for sharing between VMs)
func (vm *VM) SetLoader(loader *module.Loader) {
	vm.loader = loader
}

// SetCurrentModule sets the current module context
func (vm *VM) SetCurrentModule(mod *objects.Module) {
	vm.currentModule = mod
}

// formatError formats an error message with source location
func (vm *VM) formatError(message string) string {
	if vm.sourceMap == nil {
		return message
	}

	// Get current instruction position
	ip := vm.currentFrame().IP
	return vm.sourceMap.FormatError(ip, message)
}

// GetCallStack returns the current call stack as a formatted string
func (vm *VM) GetCallStack() string {
	var sb strings.Builder
	sb.WriteString("Call Stack:\n")

	// Walk through frames in reverse order (most recent first)
	for i := vm.frameIndex - 1; i >= 0; i-- {
		frame := vm.frames[i]
		ip := frame.IP

		var location string
		if vm.sourceMap != nil {
			if loc, ok := vm.sourceMap.Get(ip); ok {
				if vm.sourceMap.SourceFile != "" {
					location = fmt.Sprintf("%s:%d:%d", vm.sourceMap.SourceFile, loc.Line, loc.Column)
				} else {
					location = fmt.Sprintf("line %d:%d", loc.Line, loc.Column)
				}
			}
		}

		if location == "" {
			location = fmt.Sprintf("instruction %d", ip)
		}

		if i == 0 {
			sb.WriteString(fmt.Sprintf("  at <main> (%s)\n", location))
		} else {
			sb.WriteString(fmt.Sprintf("  at function#%d (%s)\n", i, location))
		}
	}

	return sb.String()
}

// Run executes the bytecode
// Optimized: caches frame pointer in local variable to reduce method calls
func (vm *VM) Run() error {
	// Register runCode callback for dynamic code execution
	// Save the previous callback to restore after execution (for nested calls)
	prevCallback := objects.SetRunCodeImpl(func(code string, args *objects.Map) (objects.Object, error) {
		return RunCodeInVM(code, args, vm)
	})
	defer objects.SetRunCodeImpl(prevCallback)

	frame := vm.frames[vm.frameIndex-1]
	frameIns := frame.Instructions()

	for frame.IP < len(frameIns)-1 {
		frame.IP++

		op := compiler.Opcode(frameIns[frame.IP])

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

		case compiler.OpGetLocalAdd:
			if err := vm.executeGetLocalAdd(); err != nil {
				return err
			}

		case compiler.OpGetLocalSub:
			if err := vm.executeGetLocalSub(); err != nil {
				return err
			}

		case compiler.OpGetLocalMul:
			if err := vm.executeGetLocalMul(); err != nil {
				return err
			}

		case compiler.OpConstantAdd:
			if err := vm.executeConstantAdd(); err != nil {
				return err
			}

		case compiler.OpConstantSub:
			if err := vm.executeConstantSub(); err != nil {
				return err
			}

		case compiler.OpConstantMul:
			if err := vm.executeConstantMul(); err != nil {
				return err
			}

		// Type-specialized instructions
		case compiler.OpIncLocal:
			if err := vm.executeIncLocal(); err != nil {
				return err
			}

		case compiler.OpDecLocal:
			if err := vm.executeDecLocal(); err != nil {
				return err
			}

		case compiler.OpAddLocalConst:
			if err := vm.executeAddLocalConst(); err != nil {
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
			pos := int(vm.readUint16())
			frame.IP = pos - 1 // -1 because loop will increment

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
			// If we returned from the main frame, exit the loop
			if vm.frameIndex == 0 {
				return nil
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

		case compiler.OpIndexSafe:
			if err := vm.executeIndexSafe(); err != nil {
				return err
			}

		case compiler.OpSetIndex:
			if err := vm.executeSetIndex(); err != nil {
				return err
			}

		case compiler.OpGetMethod:
			if err := vm.executeGetMethod(); err != nil {
				return err
			}

		case compiler.OpCallMethod:
			if err := vm.executeCallMethod(); err != nil {
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

		case compiler.OpLoadModule:
			if err := vm.executeLoadModule(); err != nil {
				return err
			}

		case compiler.OpGetExport:
			if err := vm.executeGetExport(); err != nil {
				return err
			}

		case compiler.OpSetExport:
			if err := vm.executeSetExport(); err != nil {
				return err
			}

		case compiler.OpModule:
			if err := vm.executeModule(); err != nil {
				return err
			}

		case compiler.OpClass:
			if err := vm.executeOpClass(); err != nil {
				return err
			}

		case compiler.OpNew:
			if err := vm.executeOpNew(); err != nil {
				return err
			}

		case compiler.OpGetField:
			if err := vm.executeOpGetField(); err != nil {
				return err
			}

		case compiler.OpSetField:
			if err := vm.executeOpSetField(); err != nil {
				return err
			}

		case compiler.OpSuper:
			if err := vm.executeOpSuper(); err != nil {
				return err
			}

		case compiler.OpPushHandler:
			vm.executeOpPushHandler()

		case compiler.OpPopHandler:
			vm.executeOpPopHandler()

		case compiler.OpThrow:
			if err := vm.executeOpThrow(); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}

		// Re-check frame in case it changed (function calls, returns)
		if vm.frameIndex > 0 {
			frame = vm.frames[vm.frameIndex-1]
			frameIns = frame.Instructions()
		}
	}

	return nil
}

// Helper methods

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) currentFrameMethodName() string {
	frame := vm.currentFrame()
	if frame.This != nil {
		if inst, ok := frame.This.(*objects.Instance); ok {
			return fmt.Sprintf("method of %s", inst.Class.Name)
		}
	}
	return "main"
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

	// Use frame's constants if available (for closures from modules), otherwise use VM's constants
	frame := vm.currentFrame()
	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}

	if int(constIndex) >= len(constants) {
		return fmt.Errorf("constant index out of range: %d", constIndex)
	}

	vm.stack.Push(constants[constIndex])
	return nil
}

func (vm *VM) executeBinaryOp(op compiler.Opcode) error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	// Check for mixed int/float operations
	leftIsInt, rightIsInt := isInt(left), isInt(right)
	leftIsFloat, rightIsFloat := isFloat(left), isFloat(right)

	// String concatenation (with auto-conversion)
	if op == compiler.OpAdd {
		leftIsStr := isString(left)
		rightIsStr := isString(right)

		// If either operand is a string, convert both to strings and concatenate
		if leftIsStr || rightIsStr {
			var leftStr, rightStr string
			if leftIsStr {
				leftStr = left.(*objects.String).Value
			} else {
				leftStr = left.Inspect()
			}
			if rightIsStr {
				rightStr = right.(*objects.String).Value
			} else {
				rightStr = right.Inspect()
			}
			concatenated := leftStr + rightStr
			vm.stack.Push(objects.InternString(concatenated))
			return nil
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
			return vm.divByZeroError()
		}
		result = leftVal / rightVal
	case compiler.OpMod:
		if rightVal == 0 {
			return vm.modByZeroError()
		}
		result = leftVal % rightVal
	default:
		return vm.unknownIntOpError(op)
	}

	vm.stack.Push(objects.NewInt(result))
	return nil
}

// Cold path error handlers (never inlined, kept out of hot path)

//go:noinline
func (vm *VM) divByZeroError() error {
	return fmt.Errorf("division by zero")
}

//go:noinline
func (vm *VM) modByZeroError() error {
	return fmt.Errorf("modulo by zero")
}

//go:noinline
func (vm *VM) unknownIntOpError(op compiler.Opcode) error {
	return fmt.Errorf("unknown integer operator: %d", op)
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
		vm.stack.Push(objects.NewInt(-val))
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

	// If this is a method call and localIndex is 0, return 'this'
	if frame.This != nil {
		if localIndex == 0 {
			vm.stack.Push(frame.This)
			return nil
		}
		// Adjust index if 'this' is present (shift locals by 1)
		adjustedIndex := int(localIndex) - 1
		if adjustedIndex >= len(frame.Locals) {
			return fmt.Errorf("local variable index %d out of bounds", localIndex)
		}
		vm.stack.Push(frame.Locals[adjustedIndex])
		return nil
	}
	vm.stack.Push(frame.Locals[localIndex])
	return nil
}

// Superinstruction implementations

func (vm *VM) executeGetLocalAdd() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var val1, val2 int64

	// Get first local
	if frame.This != nil {
		if idx1 == 0 {
			val1 = 0 // 'this' is not an int, this is an error case
		} else {
			val1 = frame.Locals[idx1-1].(*objects.Int).Value
		}
	} else {
		val1 = frame.Locals[idx1].(*objects.Int).Value
	}

	// Get second local
	if frame.This != nil {
		if idx2 == 0 {
			val2 = 0
		} else {
			val2 = frame.Locals[idx2-1].(*objects.Int).Value
		}
	} else {
		val2 = frame.Locals[idx2].(*objects.Int).Value
	}

	vm.stack.Push(objects.NewInt(val1 + val2))
	return nil
}

func (vm *VM) executeGetLocalSub() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var val1, val2 int64

	if frame.This != nil {
		if idx1 == 0 {
			val1 = 0
		} else {
			val1 = frame.Locals[idx1-1].(*objects.Int).Value
		}
		if idx2 == 0 {
			val2 = 0
		} else {
			val2 = frame.Locals[idx2-1].(*objects.Int).Value
		}
	} else {
		val1 = frame.Locals[idx1].(*objects.Int).Value
		val2 = frame.Locals[idx2].(*objects.Int).Value
	}

	vm.stack.Push(objects.NewInt(val1 - val2))
	return nil
}

func (vm *VM) executeGetLocalMul() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var val1, val2 int64

	if frame.This != nil {
		if idx1 == 0 {
			val1 = 0
		} else {
			val1 = frame.Locals[idx1-1].(*objects.Int).Value
		}
		if idx2 == 0 {
			val2 = 0
		} else {
			val2 = frame.Locals[idx2-1].(*objects.Int).Value
		}
	} else {
		val1 = frame.Locals[idx1].(*objects.Int).Value
		val2 = frame.Locals[idx2].(*objects.Int).Value
	}

	vm.stack.Push(objects.NewInt(val1 * val2))
	return nil
}

func (vm *VM) executeConstantAdd() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])<<8 | int(frame.Instructions()[frame.IP+2])
	idx2 := int(frame.Instructions()[frame.IP+3])<<8 | int(frame.Instructions()[frame.IP+4])
	frame.IP += 4

	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}

	val1 := constants[idx1].(*objects.Int).Value
	val2 := constants[idx2].(*objects.Int).Value

	vm.stack.Push(objects.NewInt(val1 + val2))
	return nil
}

func (vm *VM) executeConstantSub() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])<<8 | int(frame.Instructions()[frame.IP+2])
	idx2 := int(frame.Instructions()[frame.IP+3])<<8 | int(frame.Instructions()[frame.IP+4])
	frame.IP += 4

	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}

	val1 := constants[idx1].(*objects.Int).Value
	val2 := constants[idx2].(*objects.Int).Value

	vm.stack.Push(objects.NewInt(val1 - val2))
	return nil
}

func (vm *VM) executeConstantMul() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])<<8 | int(frame.Instructions()[frame.IP+2])
	idx2 := int(frame.Instructions()[frame.IP+3])<<8 | int(frame.Instructions()[frame.IP+4])
	frame.IP += 4

	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}

	val1 := constants[idx1].(*objects.Int).Value
	val2 := constants[idx2].(*objects.Int).Value

	vm.stack.Push(objects.NewInt(val1 * val2))
	return nil
}

// Type-specialized instruction implementations

// executeIncLocal increments a local variable by 1
// Optimized for loop counters - avoids push/pop overhead
func (vm *VM) executeIncLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	// Adjust index if 'this' is present
	var val *objects.Int
	if frame.This != nil && localIndex > 0 {
		val = frame.Locals[localIndex-1].(*objects.Int)
	} else if frame.This == nil {
		val = frame.Locals[localIndex].(*objects.Int)
	} else {
		// localIndex == 0 with 'this' - this is an error
		return fmt.Errorf("cannot increment 'this'")
	}

	val.Value++
	vm.stack.Push(val)
	return nil
}

// executeDecLocal decrements a local variable by 1
func (vm *VM) executeDecLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	var val *objects.Int
	if frame.This != nil && localIndex > 0 {
		val = frame.Locals[localIndex-1].(*objects.Int)
	} else if frame.This == nil {
		val = frame.Locals[localIndex].(*objects.Int)
	} else {
		return fmt.Errorf("cannot decrement 'this'")
	}

	val.Value--
	vm.stack.Push(val)
	return nil
}

// executeAddLocalConst adds a constant to a local variable
func (vm *VM) executeAddLocalConst() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	constIndex := int(frame.Instructions()[frame.IP+2])<<8 | int(frame.Instructions()[frame.IP+3])
	frame.IP += 3

	// Get constants
	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}

	// Get local value
	var localVal *objects.Int
	if frame.This != nil && localIndex > 0 {
		localVal = frame.Locals[localIndex-1].(*objects.Int)
	} else if frame.This == nil {
		localVal = frame.Locals[localIndex].(*objects.Int)
	} else {
		return fmt.Errorf("cannot add to 'this'")
	}

	// Get constant value
	constVal := constants[constIndex].(*objects.Int).Value

	localVal.Value += constVal
	vm.stack.Push(localVal)
	return nil
}

func (vm *VM) executeSetLocal() error {
	localIndex := vm.readUint8()
	vm.currentFrame().IP++

	frame := vm.currentFrame()
	value := vm.stack.Pop()
	// Adjust index if 'this' is present (shift locals by 1)
	if frame.This != nil && localIndex > 0 {
		frame.Locals[localIndex-1] = value
	} else if frame.This == nil {
		frame.Locals[localIndex] = value
	}
	// If localIndex is 0 and This is set, ignore (can't reassign this)
	vm.stack.Push(value)
	return nil
}

func (vm *VM) executeGetGlobal() error {
	globalIndex := vm.readUint16()
	vm.currentFrame().IP += 2

	frame := vm.currentFrame()
	// Use frame's globals if available (for module functions), otherwise use VM globals
	if frame.Globals != nil {
		vm.stack.Push(frame.Globals[globalIndex])
	} else {
		vm.stack.Push(vm.globals[globalIndex])
	}
	return nil
}

func (vm *VM) executeSetGlobal() error {
	globalIndex := vm.readUint16()
	vm.currentFrame().IP += 2

	value := vm.stack.Pop()
	frame := vm.currentFrame()
	// Use frame's globals if available (for module functions), otherwise use VM globals
	if frame.Globals != nil {
		frame.Globals[globalIndex] = value
	} else {
		vm.globals[globalIndex] = value
	}
	// Push the value back for assignment chaining (a = b = c)
	vm.stack.Push(value)
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
		return vm.callFunction(fn, numArgs, nil, nil, nil)
	case *Closure:
		return vm.callFunction(fn.Fn, numArgs, fn.FreeVars, fn.Constants, fn.Globals)
	case *objects.Builtin:
		return vm.callBuiltin(fn, numArgs)
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
}

func (vm *VM) callFunction(fn *compiler.CompiledFunction, numArgs int, freeVars []objects.Object, constants []objects.Object, globals []objects.Object) error {
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

	// Set constants for this frame (from closure or nil to use VM's constants)
	frame.Constants = constants

	// Set globals for this frame (from closure's module or nil to use VM's globals)
	frame.Globals = globals

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
		return vm.tailCallFunction(fn, numArgs, nil, nil, nil)
	case *Closure:
		return vm.tailCallFunction(fn.Fn, numArgs, fn.FreeVars, fn.Constants, fn.Globals)
	case *objects.Builtin:
		// Builtins don't benefit from TCO, just use regular call
		return vm.callBuiltin(fn, numArgs)
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
}

// tailCallFunction reuses the current frame for a tail call
func (vm *VM) tailCallFunction(fn *compiler.CompiledFunction, numArgs int, freeVars []objects.Object, constants []objects.Object, globals []objects.Object) error {
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

	// Resize locals array if the new function needs more locals
	// This is critical: the new function may have more local variables
	// than the current frame was originally allocated for
	if cap(frame.Locals) < fn.NumLocals {
		// Need to allocate a larger slice
		newLocals := make([]objects.Object, fn.NumLocals)
		// Copy existing arguments to new slice
		for i := 0; i < numArgs; i++ {
			newLocals[i] = args[i]
		}
		frame.Locals = newLocals
	} else {
		// Existing slice has enough capacity
		frame.Locals = frame.Locals[:fn.NumLocals]
		// Update locals with new arguments
		for i := 0; i < numArgs; i++ {
			frame.Locals[i] = args[i]
		}
		// Initialize any additional locals beyond parameters to nil
		for i := numArgs; i < fn.NumLocals; i++ {
			frame.Locals[i] = nil
		}
	}

	// Reset frame to new function
	frame.Fn = fn
	frame.IP = -1 // Will be incremented to 0 in the main loop
	frame.FreeVars = freeVars
	frame.Constants = constants
	frame.Globals = globals

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

	// Check if we're returning from the init frame that should push the instance
	if vm.pendingInstance != nil && frame == vm.initFrame {
		vm.stack.Push(vm.pendingInstance)
		vm.pendingInstance = nil
		vm.initFrame = nil
		// Don't release init frame yet, may need it for instance
		return nil
	}

	// Push the return value
	vm.stack.Push(result)

	// Release the frame back to the pool for reuse
	frame.Release()

	return nil
}

// isMainFrame returns true if we're at the main (bottom) frame
func (vm *VM) isMainFrame() bool {
	return vm.frameIndex <= 1
}

func (vm *VM) executeArray() error {
	numElements := int(vm.readUint16())
	vm.currentFrame().IP += 2

	// Pre-allocate with capacity for better performance
	elements := make([]objects.Object, numElements)
	for i := numElements - 1; i >= 0; i-- {
		elements[i] = vm.stack.Pop()
	}

	vm.stack.Push(&objects.Array{Elements: elements})
	return nil
}

// arrayCacheSize is the threshold for using pooled arrays
const arrayCacheSize = 16

// arrayPool holds reusable array slices for small arrays
var arrayPool = make([][]objects.Object, arrayCacheSize)

func (vm *VM) executeArrayPrealloc() error {
	numElements := int(vm.readUint16())
	vm.currentFrame().IP += 2

	var elements []objects.Object

	// Use pooled array for small sizes
	if numElements > 0 && numElements <= arrayCacheSize {
		if pooled := arrayPool[numElements-1]; pooled != nil {
			elements = pooled[:numElements]
			arrayPool[numElements-1] = nil // Clear from pool
		}
	}

	if elements == nil {
		elements = make([]objects.Object, numElements)
	}

	for i := numElements - 1; i >= 0; i-- {
		elements[i] = vm.stack.Pop()
	}

	vm.stack.Push(&objects.Array{Elements: elements})
	return nil
}

func (vm *VM) executeMap() error {
	numPairs := int(vm.readUint16())
	vm.currentFrame().IP += 2

	// Pre-allocate map with capacity hint for better performance
	// Go maps can be initialized with a capacity hint
	pairs := make(map[objects.HashKey]objects.MapPair, numPairs)

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

// executeIndexSafe performs array indexing without bounds checking
// Used in contexts where index is known to be valid (e.g., loop iteration)
func (vm *VM) executeIndexSafe() error {
	index := vm.stack.Pop()
	left := vm.stack.Pop()

	switch obj := left.(type) {
	case *objects.Array:
		return vm.executeArrayIndexSafe(obj, index)
	case *objects.String:
		return vm.executeStringIndexSafe(obj, index)
	default:
		// Fall back to regular index for maps and other types
		return vm.executeIndex()
	}
}

// executeArrayIndexSafe performs array indexing without bounds checking
func (vm *VM) executeArrayIndexSafe(arr *objects.Array, index objects.Object) error {
	intIndex, ok := index.(*objects.Int)
	if !ok {
		return fmt.Errorf("array index must be integer, got: %s", index.Type())
	}

	// Skip bounds check - assume index is valid
	idx := int(intIndex.Value)
	vm.stack.Push(arr.Elements[idx])
	return nil
}

// executeStringIndexSafe performs string indexing without bounds checking
func (vm *VM) executeStringIndexSafe(str *objects.String, index objects.Object) error {
	intIndex, ok := index.(*objects.Int)
	if !ok {
		return fmt.Errorf("string index must be integer, got: %s", index.Type())
	}

	// Skip bounds check - assume index is valid
	idx := int(intIndex.Value)
	vm.stack.Push(&objects.String{Value: string(str.Value[idx])})
	return nil
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

// isInt checks if an object is an integer using fast type tag
func isInt(obj objects.Object) bool {
	return obj.TypeTag() == objects.TagInt
}

// isFloat checks if an object is a float using fast type tag
func isFloat(obj objects.Object) bool {
	return obj.TypeTag() == objects.TagFloat
}

// isString checks if an object is a string using fast type tag
func isString(obj objects.Object) bool {
	return obj.TypeTag() == objects.TagString
}

func objectsEqual(a, b objects.Object) bool {
	// Handle nil cases
	if a == nil || b == nil {
		return a == b
	}

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
		objects.Builtins["runCode"],       // 41
	}

	if index < 0 || index >= len(builtins) {
		return nil
	}

	return builtins[index]
}

// Module opcode implementations

// executeLoadModule loads a module and pushes it onto the stack.
func (vm *VM) executeLoadModule() error {
	pathIdx := vm.readUint16()
	vm.currentFrame().IP += 2

	importPath := vm.constants[pathIdx].(*objects.String).Value

	// Resolve path relative to current source
	resolvedPath, err := module.Resolve(vm.sourcePath, importPath)
	if err != nil {
		return fmt.Errorf("failed to resolve import path '%s': %v", importPath, err)
	}

	// Check cache
	if vm.loader.HasModule(resolvedPath) {
		cachedMod, err := vm.loader.Get(resolvedPath)
		if err != nil {
			return err
		}
		// Convert to objects.Module and push onto stack
		mod := &objects.Module{
			Name:    cachedMod.Name,
			Exports: cachedMod.Exports,
		}
		vm.stack.Push(mod)
		return nil
	}

	// Load, compile, and execute the module
	mod, err := vm.loadModuleFile(resolvedPath)
	if err != nil {
		return err
	}

	// Push module onto stack
	vm.stack.Push(mod)
	return nil
}

// executeGetExport gets an export from a module on the stack.
func (vm *VM) executeGetExport() error {
	nameIdx := vm.readUint16()
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	// Pop the module from the stack
	modObj := vm.stack.Pop()
	mod, ok := modObj.(*objects.Module)
	if !ok {
		return fmt.Errorf("cannot get export from non-module type: %s", modObj.Type())
	}

	// Get the export
	val, ok := mod.Exports[name]
	if !ok {
		return fmt.Errorf("export '%s' not found in module %s", name, mod.Name)
	}

	// Push the export value onto the stack
	vm.stack.Push(val)
	return nil
}

// executeSetExport sets an export in the current module.
func (vm *VM) executeSetExport() error {
	nameIdx := vm.readUint16()
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	// Pop the value from the stack
	value := vm.stack.Pop()

	// Check if we have a current module context
	if vm.currentModule == nil {
		return fmt.Errorf("export statement outside of module context")
	}

	// If the exported value is a CompiledFunction, wrap it in a Closure
	// so it has access to the module's constants and globals when called
	if fn, ok := value.(*compiler.CompiledFunction); ok {
		value = &Closure{
			Fn:        fn,
			FreeVars:  nil,
			Constants: vm.constants,
			Globals:   vm.globals,
		}
	}

	// Set the export
	vm.currentModule.Exports[name] = value

	// Push the value back (export statements may be used in expressions)
	vm.stack.Push(value)
	return nil
}

// executeModule creates a module object from exports on the stack.
func (vm *VM) executeModule() error {
	numExports := int(vm.readUint16())
	vm.currentFrame().IP += 2

	// Create a new module
	mod := &objects.Module{
		Name:    vm.sourcePath,
		Exports: make(map[string]objects.Object),
	}

	// Pop name-value pairs from the stack
	for i := 0; i < numExports; i++ {
		value := vm.stack.Pop()
		nameObj := vm.stack.Pop()
		name := nameObj.(*objects.String).Value
		mod.Exports[name] = value
	}

	// Push the module onto the stack
	vm.stack.Push(mod)
	return nil
}

// executeGetMethod gets a property or method from an object.
// For instances getting a method, the instance is left on stack below the method
// so that OpCallMethod can use it as 'this'.
func (vm *VM) executeGetMethod() error {
	nameIdx := vm.readUint16()
	vm.currentFrame().IP += 2

	frame := vm.currentFrame()
	var name string
	if frame.Constants != nil {
		name = frame.Constants[nameIdx].(*objects.String).Value
	} else {
		name = vm.constants[nameIdx].(*objects.String).Value
	}

	// Peek at the object on the stack (don't pop yet)
	obj := vm.stack.Top()

	// Handle Instance objects (for method/field access)
	if instance, ok := obj.(*objects.Instance); ok {
		// Check inline cache first - use pointer identity for class comparison
		cacheKey := uint32(uintptr(unsafe.Pointer(instance.Class))) ^ uint32(len(name))
		cacheIdx := cacheKey % InlineCacheSize
		cached := &vm.methodCache[cacheIdx]

		if cached.Class == instance.Class {
			vm.cacheHits++
			if cached.Method != nil {
				// Cache hit - method found
				vm.stack.Push(cached.Method)
				return nil
			}
			// Cache hit - method not found (negative cache)
			vm.stack.Pop() // pop instance
			vm.stack.Push(objects.NULL)
			return nil
		}

		// Cache miss - perform full lookup
		vm.cacheMisses++

		// First check for method - leave instance on stack and push method
		method := vm.findMethod(instance.Class, name)
		if method != nil {
			// Update cache
			cached.Class = instance.Class
			cached.Method = method

			// Stack: [... instance] -> [... instance, method]
			// OpCallMethod will pop method, use instance as 'this'
			vm.stack.Push(method)
			return nil
		}

		// Check if it's a field - pop instance, push field value
		if val, ok := instance.Fields[name]; ok {
			vm.stack.Pop() // pop instance
			vm.stack.Push(val)
			return nil
		}

		// Cache negative result (method not found)
		cached.Class = instance.Class
		cached.Method = nil

		vm.stack.Pop() // pop instance
		vm.stack.Push(objects.NULL)
		return nil
	}

	// Handle Map objects - property access takes precedence over methods
	if m, ok := obj.(*objects.Map); ok {
		key := objects.InternString(name)
		pair, exists := m.Pairs[key.HashKey()]
		if exists {
			// Property exists - pop map and push value
			vm.stack.Pop()
			vm.stack.Push(pair.Value)
			return nil
		}
		// Property doesn't exist - check for map method
		if method, ok := objects.GetMethod(objects.MapType, name); ok {
			// Stack: [... map] -> [... map, method]
			vm.stack.Push(method)
			return nil
		}
		// No property and no method - pop map and push NULL
		vm.stack.Pop()
		vm.stack.Push(objects.NULL)
		return nil
	}

	// Check for primitive type methods (Int, Float, String, Array, Bool, Null)
	if method, ok := objects.GetMethod(obj.Type(), name); ok {
		// Stack: [... obj] -> [... obj, method]
		// OpCallMethod will pop method, use obj as receiver
		vm.stack.Push(method)
		return nil
	}

	// For non-instance objects without methods, pop and handle property access
	obj = vm.stack.Pop()

	// Handle Module objects (for namespace imports)
	// Check if this is a method call by scanning ahead for OpCallMethod
	if mod, ok := obj.(*objects.Module); ok {
		val, ok := mod.Exports[name]
		if !ok {
			return fmt.Errorf("export '%s' not found in module %s", name, mod.Name)
		}

		// Scan ahead for OpCallMethod (skip up to 20 instructions to find it)
		frameIns := frame.Instructions()
		isMethodCall := false
		scanIP := frame.IP + 1
		maxScan := 20
		for i := 0; i < maxScan && scanIP < len(frameIns); i++ {
			op := compiler.Opcode(frameIns[scanIP])
			if op == compiler.OpCallMethod {
				isMethodCall = true
				break
			}
			// Stop scanning if we hit certain opcodes that indicate this is not a method call
			if op == compiler.OpPop || op == compiler.OpCall || op == compiler.OpSetGlobal ||
				op == compiler.OpSetLocal || op == compiler.OpJump || op == compiler.OpJumpIfFalse {
				break
			}
			// Skip operands based on opcode
			switch op {
			case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal,
				compiler.OpGetLocal, compiler.OpSetLocal, compiler.OpGetMethod,
				compiler.OpGetExport, compiler.OpSetExport:
				scanIP += 3 // 1 byte opcode + 2 bytes operand
			case compiler.OpCall, compiler.OpCallMethod:
				scanIP += 2 // 1 byte opcode + 1 byte operand
			default:
				scanIP += 1
			}
		}

		if isMethodCall {
			// Method call: keep module on stack so executeCallMethod can detect it
			// Stack: [... module] -> [... module, export]
			vm.stack.Push(obj) // push module back
			vm.stack.Push(val)
		} else {
			// Property access: just push the value
			vm.stack.Push(val)
		}
		return nil
	}

	return fmt.Errorf("cannot access property '%s' on type %s", name, obj.Type())
}

// executeCallMethod calls a method on an instance with 'this' binding
// For non-instance calls (like module functions), it falls back to regular call
func (vm *VM) executeCallMethod() error {
	numArgs := int(vm.readUint8())
	vm.currentFrame().IP++

	// Stack: [... instance, method, arg1, arg2, ...]
	// Peek(numArgs) gets the method
	// Peek(numArgs+1) gets the instance (this)

	method := vm.stack.Peek(numArgs)
	instance := vm.stack.Peek(numArgs + 1)

	// Check if this is an Instance object for method binding
	inst, ok := instance.(*objects.Instance)
	if !ok {
		// Check if this is a Module function call (Module instance)
		// Module functions should NOT receive the module as a receiver
		if _, isModule := instance.(*objects.Module); isModule {
			// Module function call - don't pass module as receiver
			// Stack: [... module, method, arg1, arg2, ...]
			// Pop arguments first
			args := make([]objects.Object, numArgs)
			for i := numArgs - 1; i >= 0; i-- {
				args[i] = vm.stack.Pop()
			}

			// Pop the function
			vm.stack.Pop()

			// Pop the module (not used as argument)
			vm.stack.Pop()

			// Call as regular function (without receiver)
			switch f := method.(type) {
			case *compiler.CompiledFunction:
				return vm.callFunction(f, numArgs, nil, nil, nil)
			case *Closure:
				return vm.callFunction(f.Fn, numArgs, f.FreeVars, f.Constants, f.Globals)
			case *objects.Builtin:
				// Call builtin directly (we already popped all the stack items)
				result := f.Fn(args...)
				vm.stack.Push(result)
				return nil
			default:
				return fmt.Errorf("cannot call non-function: %T", method)
			}
		}

		// Check if method is a Builtin (for primitive type methods)
		// This handles calls like 42.typeOf(), null.typeOf(), etc.
		if builtin, isBuiltin := method.(*objects.Builtin); isBuiltin {
			// Stack: [... receiver, builtin_method, arg1, arg2, ...]
			// Pop arguments first
			args := make([]objects.Object, numArgs)
			for i := numArgs - 1; i >= 0; i-- {
				args[i] = vm.stack.Pop()
			}

			// Pop the builtin method
			vm.stack.Pop()

			// Pop the receiver
			receiver := vm.stack.Pop()

			// Call builtin with receiver as first argument
			callArgs := make([]objects.Object, numArgs+1)
			callArgs[0] = receiver
			for i := 0; i < numArgs; i++ {
				callArgs[i+1] = args[i]
			}

			result := builtin.Fn(callArgs...)
			vm.stack.Push(result)
			return nil
		}

		// Not an instance, module, or builtin method - treat as regular function call
		// Stack is: [method, arg1, arg2, ...]
		// Pop arguments first
		args := make([]objects.Object, numArgs)
		for i := numArgs - 1; i >= 0; i-- {
			args[i] = vm.stack.Pop()
		}

		// Pop the function
		fn := vm.stack.Pop()

		// Call as regular function
		switch f := fn.(type) {
		case *compiler.CompiledFunction:
			return vm.callFunction(f, numArgs, nil, nil, nil)
		case *Closure:
			return vm.callFunction(f.Fn, numArgs, f.FreeVars, f.Constants, f.Globals)
		case *objects.Builtin:
			return vm.callBuiltin(f, numArgs)
		default:
			return fmt.Errorf("cannot call non-function: %T", fn)
		}
	}

	// Set current instance for this/super binding
	vm.currentInstance = inst

	// Get the compiled function
	var fn *compiler.CompiledFunction
	var freeVars []objects.Object
	var constants []objects.Object
	var globals []objects.Object

	switch m := method.(type) {
	case *compiler.CompiledFunction:
		fn = m
	case *Closure:
		fn = m.Fn
		freeVars = m.FreeVars
		constants = m.Constants
		globals = m.Globals
	default:
		return fmt.Errorf("method is not a function: %T", method)
	}

	// Check argument count
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	// Stack order: [... instance, method, arg1, arg2, ...]
	// Pop arguments first (they are on top of the method)
	args := make([]objects.Object, numArgs)
	for i := numArgs - 1; i >= 0; i-- {
		args[i] = vm.stack.Pop()
	}

	// Pop the method from stack
	vm.stack.Pop()

	// Pop instance from stack
	vm.stack.Pop()

	// Create new frame with the instance as 'this'
	frame := NewFrame(fn, vm.stack.Len())
	frame.This = inst

	// Store arguments in locals (starting at index 0, shifted by executeGetLocal)
	for i := 0; i < numArgs; i++ {
		frame.Locals[i] = args[i]
	}

	// Set free variables
	if freeVars != nil {
		frame.FreeVars = freeVars
	}

	// Set constants and globals
	frame.Constants = constants
	frame.Globals = globals

	vm.pushFrame(frame)


	return nil
}

// executeOpClass creates a class object
func (vm *VM) executeOpClass() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	// Pop methods map
	methodsObj := vm.stack.Pop()
	methods, ok := methodsObj.(*objects.Map)
	if !ok {
		return fmt.Errorf("methods must be a map")
	}

	// Pop fields map
	fieldsObj := vm.stack.Pop()
	fields, ok := fieldsObj.(*objects.Map)
	if !ok {
		return fmt.Errorf("fields must be a map")
	}

	// Pop superclass (or null)
	superObj := vm.stack.Pop()
	var superClass *objects.Class
	if superObj != objects.NULL {
		superClass, ok = superObj.(*objects.Class)
		if !ok {
			return fmt.Errorf("superclass must be a class")
		}
	}

	// Build methods map
	classMethods := make(map[string]objects.Object)
	var initMethod objects.Object
	for _, pair := range methods.Pairs {
		methodName := pair.Key.(*objects.String).Value
		classMethods[methodName] = pair.Value
		if methodName == "init" {
			initMethod = pair.Value
		}
	}

	// If no init method in this class, look in superclass chain
	if initMethod == nil && superClass != nil {
		initMethod = vm.findInitMethod(superClass)
	}

	// Build fields map
	classFields := make(map[string]objects.Object)
	for _, pair := range fields.Pairs {
		fieldName := pair.Key.(*objects.String).Value
		classFields[fieldName] = pair.Value
	}

	// Create class
	class := &objects.Class{
		Name:       name,
		SuperClass: superClass,
		Methods:    classMethods,
		InitMethod: initMethod,
		Fields:     classFields,
	}

	vm.stack.Push(class)
	return nil
}

// executeOpNew creates a new instance
func (vm *VM) executeOpNew() error {
	argCount := int(vm.readUint8())
	vm.currentFrame().IP++

	// Stack has: [class, arg1, arg2, ...]
	// Pop arguments first (in reverse order)
	args := make([]objects.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.stack.Pop()
	}

	// Now pop the class
	classObj := vm.stack.Pop()
	class, ok := classObj.(*objects.Class)
	if !ok {
		return fmt.Errorf("cannot use 'new' on non-class type")
	}

	// Collect fields from class hierarchy
	fields := make(map[string]objects.Object)
	vm.collectFields(class, fields)

	// Create instance
	instance := &objects.Instance{
		Class:  class,
		Fields: fields,
	}

	// Call init if exists
	if class.InitMethod != nil {
		// The init method can be either a Closure or CompiledFunction
		var initFn *compiler.CompiledFunction
		var freeVars []objects.Object
		var constants []objects.Object
		var globals []objects.Object

		switch m := class.InitMethod.(type) {
		case *Closure:
			initFn = m.Fn
			freeVars = m.FreeVars
			constants = m.Constants
			globals = m.Globals
		case *compiler.CompiledFunction:
			initFn = m
		default:
			return fmt.Errorf("init method must be a function, got %T", class.InitMethod)
		}

		// Check argument count
		if argCount != initFn.NumParameters {
			return fmt.Errorf("wrong number of arguments for init: want=%d, got=%d", initFn.NumParameters, argCount)
		}

		// Set current instance for this/super binding
		vm.currentInstance = instance

		// Create new frame with the instance as 'this'
		frame := NewFrame(initFn, vm.stack.Len())
		frame.This = instance

		// Store arguments in locals
		for i := 0; i < argCount; i++ {
			frame.Locals[i] = args[i]
		}

		// Set free variables
		if freeVars != nil {
			frame.FreeVars = freeVars
		}

		// Set constants and globals
		frame.Constants = constants
		frame.Globals = globals

		// Store the instance so we can push it after init returns
		vm.pendingInstance = instance
		vm.initFrame = frame

		vm.pushFrame(frame)
		return nil
	}

	// Push instance onto stack
	vm.stack.Push(instance)
	return nil
}

// executeOpGetField gets a field from an instance
func (vm *VM) executeOpGetField() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	frame := vm.currentFrame()
	var name string
	if frame.Constants != nil && int(nameIdx) < len(frame.Constants) {
		name = frame.Constants[nameIdx].(*objects.String).Value
	} else {
		name = vm.constants[nameIdx].(*objects.String).Value
	}

	obj := vm.stack.Pop()
	instance, ok := obj.(*objects.Instance)
	if !ok {
		return fmt.Errorf("cannot access field '%s' on %s", name, obj.Type())
	}

	value, ok := instance.Fields[name]
	if !ok {
		vm.stack.Push(objects.NULL)
	} else {
		vm.stack.Push(value)
	}
	return nil
}

// executeOpSetField sets a field on an instance
func (vm *VM) executeOpSetField() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	frame := vm.currentFrame()
	var name string
	if frame.Constants != nil && int(nameIdx) < len(frame.Constants) {
		name = frame.Constants[nameIdx].(*objects.String).Value
	} else {
		name = vm.constants[nameIdx].(*objects.String).Value
	}

	// Stack order from compiler: [value, object]
	// value is pushed first, then object
	// So we pop: object (top), then value (below)
	obj := vm.stack.Pop()
	value := vm.stack.Pop()

	instance, ok := obj.(*objects.Instance)
	if !ok {
		return fmt.Errorf("cannot set field '%s' on %s", name, obj.Type())
	}

	instance.Fields[name] = value
	vm.stack.Push(value)
	return nil
}

// executeOpSuper gets a method from the superclass
func (vm *VM) executeOpSuper() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	frame := vm.currentFrame()
	var name string
	if frame.Constants != nil && int(nameIdx) < len(frame.Constants) {
		name = frame.Constants[nameIdx].(*objects.String).Value
	} else {
		name = vm.constants[nameIdx].(*objects.String).Value
	}

	if vm.currentInstance == nil {
		return fmt.Errorf("cannot use 'super' outside of method context")
	}

	class := vm.currentInstance.Class
	if class.SuperClass == nil {
		return fmt.Errorf("class '%s' has no superclass", class.Name)
	}

	// Find method in superclass chain
	method := vm.findMethod(class.SuperClass, name)
	if method == nil {
		return fmt.Errorf("method '%s' not found in superclass", name)
	}

	vm.stack.Push(method)
	return nil
}

// collectFields collects fields from class hierarchy
func (vm *VM) collectFields(class *objects.Class, fields map[string]objects.Object) {
	// Collect from parent first
	if class.SuperClass != nil {
		vm.collectFields(class.SuperClass, fields)
	}
	// Collect from this class (overrides parent)
	for name, value := range class.Fields {
		fields[name] = value
	}
}

// findMethod finds a method in class hierarchy
func (vm *VM) findMethod(class *objects.Class, name string) objects.Object {
	for c := class; c != nil; c = c.SuperClass {
		if method, ok := c.Methods[name]; ok {
			return method
		}
	}
	return nil
}

// findInitMethod finds the init method in class hierarchy
func (vm *VM) findInitMethod(class *objects.Class) objects.Object {
	for c := class; c != nil; c = c.SuperClass {
		if method, ok := c.Methods["init"]; ok {
			return method
		}
	}
	return nil
}

// executeOpPushHandler pushes an exception handler onto the handler stack
func (vm *VM) executeOpPushHandler() {
	catchAddr := int(vm.readUint16())
	vm.currentFrame().IP += 2
	finallyAddr := int(vm.readUint16())
	vm.currentFrame().IP += 2

	handler := ExceptionHandler{
		catchAddr:   catchAddr,
		finallyAddr: finallyAddr,
		frameIndex:  vm.frameIndex,
		stackSize:   vm.stack.Len(),
	}
	vm.handlers = append(vm.handlers, handler)
}

// executeOpPopHandler pops an exception handler from the handler stack
func (vm *VM) executeOpPopHandler() {
	if len(vm.handlers) > 0 {
		vm.handlers = vm.handlers[:len(vm.handlers)-1]
	}
}

// executeOpThrow handles throwing an exception
func (vm *VM) executeOpThrow() error {
	// Get the exception value from the stack
	err := vm.stack.Pop()

	// Find a handler
	for len(vm.handlers) > 0 {
		handler := vm.handlers[len(vm.handlers)-1]

		// Pop this handler since we're using it
		vm.handlers = vm.handlers[:len(vm.handlers)-1]

		// If we need to unwind frames
		for vm.frameIndex > handler.frameIndex {
			vm.popFrame()
		}

		// Restore stack size
		for vm.stack.Len() > handler.stackSize {
			vm.stack.Pop()
		}

		// Push the error value for catch block
		vm.stack.Push(err)

		// Jump to catch if available, otherwise finally
		if handler.catchAddr > 0 {
			vm.currentFrame().IP = handler.catchAddr - 1 // -1 because main loop will increment
			return nil
		} else if handler.finallyAddr > 0 {
			vm.currentFrame().IP = handler.finallyAddr - 1
			return nil
		}
	}

	// No handler found, return as error
	return fmt.Errorf("unhandled exception: %s", err.Inspect())
}
