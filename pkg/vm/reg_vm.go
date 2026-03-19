// pkg/vm/reg_vm.go
// Register-based Virtual Machine implementation
package vm

import (
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/stdlib"
)

// RegVM is a register-based virtual machine
type RegVM struct {
	constants     []Value          // Constant pool (as Values)
	frames        []*RegFrame      // Call frames
	frameIndex    int              // Current frame index
	globals       []Value          // Global variables
	loader        *module.Loader   // Module loader
	currentModule *objects.Module  // Current module context
	sourcePath    string           // Source file path
	sourceMap     *compiler.SourceMap

	// Object constants (for non-Value types like strings, arrays, maps)
	objConstants []objects.Object

	// Exception handling
	handlers []ExceptionHandler

	// Inline cache for method lookups
	methodCache [InlineCacheSize]InlineCacheEntry

	// Temp stack for complex expressions that don't fit in registers
	tempStack *ValueStack
}

// NewRegVM creates a new register-based VM
func NewRegVM(bytecode *compiler.Bytecode) *RegVM {
	// Convert constants to Values
	constants := make([]Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = NewObject(c)
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
	mainFrame := NewRegFrame(mainFn)
	mainFrame.Constants = constants
	mainFrame.Globals = make([]Value, GlobalsSize)

	frames := make([]*RegFrame, MaxFrames)
	frames[0] = mainFrame

	return &RegVM{
		constants:   constants,
		objConstants: bytecode.Constants,
		frames:      frames,
		frameIndex:  1,
		globals:     make([]Value, GlobalsSize),
		loader:      module.NewLoader(),
		sourceMap:   bytecode.SourceMap,
		tempStack:   NewValueStack(),
	}
}

// NewRegVMWithGlobals creates a register VM with custom globals
func NewRegVMWithGlobals(bytecode *compiler.Bytecode, globals []Value) *RegVM {
	constants := make([]Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = NewObject(c)
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
	mainFrame := NewRegFrame(mainFn)
	mainFrame.Constants = constants
	mainFrame.Globals = globals

	frames := make([]*RegFrame, MaxFrames)
	frames[0] = mainFrame

	return &RegVM{
		constants:   constants,
		objConstants: bytecode.Constants,
		frames:      frames,
		frameIndex:  1,
		globals:     globals,
		loader:      module.NewLoader(),
		sourceMap:   bytecode.SourceMap,
		tempStack:   NewValueStack(),
	}
}

// currentFrame returns the current frame
func (vm *RegVM) currentFrame() *RegFrame {
	return vm.frames[vm.frameIndex-1]
}

// pushFrame pushes a new frame onto the call stack
func (vm *RegVM) pushFrame(frame *RegFrame) {
	if vm.frameIndex >= MaxFrames {
		panic("stack overflow")
	}
	vm.frames[vm.frameIndex] = frame
	vm.frameIndex++
}

// popFrame pops a frame from the call stack
func (vm *RegVM) popFrame() *RegFrame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

// LastPopped returns the last popped value from temp stack
// Note: For register VM, prefer LastResult() which returns from ReturnRegister
func (vm *RegVM) LastPopped() Value {
	return vm.tempStack.LastPopped()
}

// LastResult returns the value in the ReturnRegister
// This is the preferred method for getting results from the register VM
func (vm *RegVM) LastResult() Value {
	frame := vm.currentFrame()
	if frame == nil {
		return ValueNull
	}
	return frame.Registers[compiler.ReturnRegister]
}

// Globals returns the globals array
func (vm *RegVM) Globals() []Value {
	return vm.globals
}

// GlobalsAsObjects returns the globals as objects.Object slice
func (vm *RegVM) GlobalsAsObjects() []objects.Object {
	result := make([]objects.Object, len(vm.globals))
	for i, v := range vm.globals {
		result[i] = v.ToObject()
	}
	return result
}

// LastPoppedObject returns the result as objects.Object
// For register VM, this returns the value from ReturnRegister
func (vm *RegVM) LastPoppedObject() objects.Object {
	frame := vm.currentFrame()
	if frame == nil {
		return objects.NULL
	}

	// Return value from ReturnRegister (set by function/builtin calls or program end)
	return frame.Registers[compiler.ReturnRegister].ToObject()
}

// NewRegVMWithObjectGlobals creates a register VM with globals as objects.Object
func NewRegVMWithObjectGlobals(bytecode *compiler.Bytecode, globals []objects.Object) *RegVM {
	// Convert objects.Object to Value
	valueGlobals := make([]Value, len(globals))
	for i, obj := range globals {
		valueGlobals[i] = NewObject(obj)
	}
	return NewRegVMWithGlobals(bytecode, valueGlobals)
}

// SetSourcePath sets the source file path
func (vm *RegVM) SetSourcePath(path string) {
	vm.sourcePath = path
}

// SetLoader sets the module loader
func (vm *RegVM) SetLoader(loader *module.Loader) {
	vm.loader = loader
}

// SetCurrentModule sets the current module context
func (vm *RegVM) SetCurrentModule(mod *objects.Module) {
	vm.currentModule = mod
}

// Run executes the bytecode in the register VM
func (vm *RegVM) Run() error {
	// Register callbacks for dynamic code execution
	prevCallback := objects.SetRunCodeImpl(func(code string, args *objects.Map) (objects.Object, error) {
		return RunCodeInRegVM(code, args, vm)
	})
	defer objects.SetRunCodeImpl(prevCallback)

	prevLoadPlugin := objects.SetLoadPluginImpl(func(path string) (objects.Object, error) {
		return vm.loadPluginByPath(path)
	})
	defer objects.SetLoadPluginImpl(prevLoadPlugin)

	frame := vm.currentFrame()
	code := frame.Instructions()

	for frame.IP < len(code) {
		op := compiler.Opcode(code[frame.IP])

		// Check if this is a register-based opcode
		if compiler.IsRegisterOpcode(op) {
			if err := vm.executeRegInstruction(op, frame, code); err != nil {
				return err
			}
			frame = vm.currentFrame()
			code = frame.Instructions()
			continue
		}

		// Fall back to stack-based instruction (for mixed bytecode)
		frame.IP++
		switch op {
		case compiler.OpConstant:
			if err := vm.executeConstant(frame, code); err != nil {
				return err
			}
		case compiler.OpPop:
			vm.tempStack.Pop()
		case compiler.OpAdd:
			if err := vm.executeValueBinaryOp(func(a, b Value) (Value, error) {
				result, ok := a.Add(b)
				if !ok {
					return ValueNull, fmt.Errorf("type error in addition")
				}
				return result, nil
			}); err != nil {
				return err
			}
		case compiler.OpSub:
			if err := vm.executeValueBinaryOp(func(a, b Value) (Value, error) {
				result, ok := a.Sub(b)
				if !ok {
					return ValueNull, fmt.Errorf("type error in subtraction")
				}
				return result, nil
			}); err != nil {
				return err
			}
		case compiler.OpMul:
			if err := vm.executeValueBinaryOp(func(a, b Value) (Value, error) {
				result, ok := a.Mul(b)
				if !ok {
					return ValueNull, fmt.Errorf("type error in multiplication")
				}
				return result, nil
			}); err != nil {
				return err
			}
		case compiler.OpDiv:
			if err := vm.executeValueBinaryOp(func(a, b Value) (Value, error) {
				if b.IsInt() && b.GetInt() == 0 {
					return ValueNull, fmt.Errorf("division by zero")
				}
				result, ok := a.Div(b)
				if !ok {
					return ValueNull, fmt.Errorf("type error in division")
				}
				return result, nil
			}); err != nil {
				return err
			}
		case compiler.OpReturn:
			return vm.handleRegReturn(frame)
		default:
			return fmt.Errorf("unhandled stack opcode %d in register VM", op)
		}
		frame = vm.currentFrame()
		code = frame.Instructions()
	}

	return nil
}

// executeRegInstruction executes a single register-based instruction
func (vm *RegVM) executeRegInstruction(op compiler.Opcode, frame *RegFrame, code []byte) error {
	regs := &frame.Registers

	switch op {
	// Arithmetic operations
	case compiler.OpRegAdd:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		result, ok := regs[src1].Add(regs[src2])
		if !ok {
			return fmt.Errorf("type error in addition")
		}
		regs[dst] = result
		frame.IP += 4

	case compiler.OpRegSub:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		result, ok := regs[src1].Sub(regs[src2])
		if !ok {
			return fmt.Errorf("type error in subtraction")
		}
		regs[dst] = result
		frame.IP += 4

	case compiler.OpRegMul:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		result, ok := regs[src1].Mul(regs[src2])
		if !ok {
			return fmt.Errorf("type error in multiplication")
		}
		regs[dst] = result
		frame.IP += 4

	case compiler.OpRegDiv:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		if regs[src2].IsInt() && regs[src2].GetInt() == 0 {
			return fmt.Errorf("division by zero")
		}
		result, ok := regs[src1].Div(regs[src2])
		if !ok {
			return fmt.Errorf("type error in division")
		}
		regs[dst] = result
		frame.IP += 4

	case compiler.OpRegMod:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		result, ok := regs[src1].Mod(regs[src2])
		if !ok {
			return fmt.Errorf("type error in modulo")
		}
		regs[dst] = result
		frame.IP += 4

	case compiler.OpRegNeg:
		dst, src := DecodeReg2(code, frame.IP)
		result, ok := regs[src].Neg()
		if !ok {
			return fmt.Errorf("type error in negation")
		}
		regs[dst] = result
		frame.IP += 3

	case compiler.OpRegAnd:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = ValueBool(regs[src1].IsTruthy() && regs[src2].IsTruthy())
		frame.IP += 4

	case compiler.OpRegOr:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = ValueBool(regs[src1].IsTruthy() || regs[src2].IsTruthy())
		frame.IP += 4

	case compiler.OpRegNot:
		dst, src := DecodeReg2(code, frame.IP)
		regs[dst] = ValueBool(!regs[src].IsTruthy())
		frame.IP += 3

	// Comparison operations
	case compiler.OpRegLess:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].LessValue(regs[src2])
		frame.IP += 4

	case compiler.OpRegLessEqual:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].LessEqual(regs[src2])
		frame.IP += 4

	case compiler.OpRegGreater:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].GreaterValue(regs[src2])
		frame.IP += 4

	case compiler.OpRegGreaterEqual:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].GreaterEqual(regs[src2])
		frame.IP += 4

	case compiler.OpRegEqual:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].EqualValue(regs[src2])
		frame.IP += 4

	case compiler.OpRegNotEqual:
		dst, src1, src2 := DecodeReg3(code, frame.IP)
		regs[dst] = regs[src1].NotEqualValue(regs[src2])
		frame.IP += 4

	// Data movement
	case compiler.OpRegMove:
		dst, src := DecodeReg2(code, frame.IP)
		regs[dst] = regs[src]
		frame.IP += 3

	case compiler.OpRegLoadConst:
		dst, constIdx := DecodeConst(code, frame.IP)
		regs[dst] = frame.Constants[constIdx]
		frame.IP += 4

	case compiler.OpRegLoadGlobal:
		dst, globalIdx := DecodeConst(code, frame.IP)
		regs[dst] = frame.Globals[globalIdx]
		frame.IP += 4

	case compiler.OpRegStoreGlobal:
		src, globalIdx := DecodeConst(code, frame.IP)
		frame.Globals[globalIdx] = regs[src]
		frame.IP += 4

	// Local variable operations
	case compiler.OpRegLoadLocal:
		dst, localIdx := DecodeReg2(code, frame.IP)
		regs[dst] = frame.Locals[localIdx]
		frame.IP += 3

	case compiler.OpRegStoreLocal:
		src, localIdx := DecodeReg2(code, frame.IP)
		frame.Locals[localIdx] = regs[src]
		frame.IP += 3

	case compiler.OpRegLoadFree:
		dst, freeIdx := DecodeReg2(code, frame.IP)
		regs[dst] = frame.FreeVars[freeIdx]
		frame.IP += 3

	case compiler.OpRegStoreFree:
		src, freeIdx := DecodeReg2(code, frame.IP)
		frame.FreeVars[freeIdx] = regs[src]
		frame.IP += 3

	// Control flow
	case compiler.OpRegJump:
		offset := DecodeJump(code, frame.IP)
		frame.IP += offset

	case compiler.OpRegJumpIfTrue:
		condReg, offset := DecodeJumpCond(code, frame.IP)
		if regs[condReg].IsTruthy() {
			frame.IP += offset
		} else {
			frame.IP += 4
		}

	case compiler.OpRegJumpIfFalse:
		condReg, offset := DecodeJumpCond(code, frame.IP)
		if !regs[condReg].IsTruthy() {
			frame.IP += offset
		} else {
			frame.IP += 4
		}

	// Function operations
	case compiler.OpRegCall:
		return vm.handleRegCall(frame, code)

	case compiler.OpRegTailCall:
		return vm.handleRegTailCall(frame, code)

	case compiler.OpRegBuiltin:
		builtinIdx, numArgs := DecodeReg2(code, frame.IP)
		frame.IP += 3
		return vm.handleRegBuiltin(int(builtinIdx), int(numArgs), frame)

	case compiler.OpRegReturn:
		// Decode the register containing the return value
		retReg := DecodeReg1(code, frame.IP)
		frame.IP += 2
		// Copy return value to ReturnRegister
		frame.Registers[compiler.ReturnRegister] = regs[retReg]
		return vm.handleRegReturn(frame)

	case compiler.OpRegClosure:
		// Format: OpRegClosure dst func_idx(16-bit) num_free start_reg
		dst := code[frame.IP+1]
		fnIndex := int(code[frame.IP+2])<<8 | int(code[frame.IP+3])
		numFree := int(code[frame.IP+4])
		startReg := int(code[frame.IP+5])
		frame.IP += 6

		// Get the compiled function from constants
		fn, ok := frame.Constants[fnIndex].ToObject().(*compiler.CompiledFunction)
		if !ok {
			return fmt.Errorf("expected CompiledFunction at index %d, got %T", fnIndex, frame.Constants[fnIndex].ToObject())
		}

		// Create closure with captured free variables
		// Note: For register VM, Constants and Globals are taken from callerFrame
		closure := &Closure{
			Fn:             fn,
			FreeVars:       nil, // Not used in register VM
			Constants:      nil, // Not used in register VM
			Globals:        nil, // Not used in register VM
			FreeVarsValues: make([]Value, numFree),
		}

		// Copy free variables from registers (these are the captured values)
		for i := 0; i < numFree; i++ {
			closure.FreeVarsValues[i] = regs[startReg+i]
		}

		regs[dst] = NewObject(closure)

	// Collection operations
	case compiler.OpRegArray:
		dst, startReg, count := DecodeReg3(code, frame.IP)
		elements := make([]objects.Object, count)
		for i := 0; i < int(count); i++ {
			elements[i] = regs[int(startReg)+i].ToObject()
		}
		regs[dst] = NewObject(&objects.Array{Elements: elements})
		frame.IP += 4

	case compiler.OpRegMap:
		dst, startReg, count := DecodeReg3(code, frame.IP)
		pairs := make(map[objects.HashKey]objects.MapPair)
		for i := 0; i < int(count); i++ {
			keyObj := regs[int(startReg)+i*2].ToObject()
			valObj := regs[int(startReg)+i*2+1].ToObject()
			pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: valObj}
		}
		regs[dst] = NewObject(&objects.Map{Pairs: pairs})
		frame.IP += 4

	case compiler.OpRegArrayEmpty:
		dst := code[frame.IP+1]
		regs[dst] = NewObject(&objects.Array{Elements: []objects.Object{}})
		frame.IP += 2

	case compiler.OpRegArrayAppend:
		dst := code[frame.IP+1]
		arrReg := code[frame.IP+2]
		elemReg := code[frame.IP+3]

		arr := regs[arrReg].ToObject()
		arrObj, ok := arr.(*objects.Array)
		if !ok {
			return fmt.Errorf("OpRegArrayAppend: expected array, got %T", arr)
		}

		// Create new array with appended element
		newElements := make([]objects.Object, len(arrObj.Elements)+1)
		copy(newElements, arrObj.Elements)
		newElements[len(arrObj.Elements)] = regs[elemReg].ToObject()

		regs[dst] = NewObject(&objects.Array{Elements: newElements})
		frame.IP += 4

	case compiler.OpRegMapEmpty:
		dst := code[frame.IP+1]
		regs[dst] = NewObject(&objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)})
		frame.IP += 2

	case compiler.OpRegMapSet:
		dst := code[frame.IP+1]
		mapReg := code[frame.IP+2]
		keyReg := code[frame.IP+3]
		valReg := code[frame.IP+4]

		mapObj := regs[mapReg].ToObject()
		m, ok := mapObj.(*objects.Map)
		if !ok {
			return fmt.Errorf("OpRegMapSet: expected map, got %T", mapObj)
		}

		// Create new map with added key-value pair
		newPairs := make(map[objects.HashKey]objects.MapPair, len(m.Pairs)+1)
		for k, v := range m.Pairs {
			newPairs[k] = v
		}
		keyObj := regs[keyReg].ToObject()
		valObj := regs[valReg].ToObject()
		newPairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: valObj}

		regs[dst] = NewObject(&objects.Map{Pairs: newPairs})
		frame.IP += 5

	case compiler.OpRegPush:
		srcReg := code[frame.IP+1]
		vm.tempStack.Push(regs[srcReg])
		frame.IP += 2

	case compiler.OpRegPop:
		dstReg := code[frame.IP+1]
		val := vm.tempStack.Pop()
		regs[dstReg] = val
		frame.IP += 2

	case compiler.OpRegIndex:
		dst, objReg, keyReg := DecodeReg3(code, frame.IP)
		obj := regs[objReg].ToObject()
		key := regs[keyReg].ToObject()

		var result objects.Object
		switch o := obj.(type) {
		case *objects.Array:
			idx, ok := key.(*objects.Int)
			if !ok {
				return fmt.Errorf("array index must be integer")
			}
			if idx.Value < 0 || idx.Value >= int64(len(o.Elements)) {
				return fmt.Errorf("array index out of bounds: %d", idx.Value)
			}
			result = o.Elements[idx.Value]
		case *objects.Map:
			hashKey := key.HashKey()
			pair, ok := o.Pairs[hashKey]
			if !ok {
				result = objects.NULL
			} else {
				result = pair.Value
			}
		case *objects.String:
			idx, ok := key.(*objects.Int)
			if !ok {
				return fmt.Errorf("string index must be integer")
			}
			if idx.Value < 0 || idx.Value >= int64(len(o.Value)) {
				return fmt.Errorf("string index out of bounds: %d", idx.Value)
			}
			result = &objects.String{Value: string(o.Value[idx.Value])}
		default:
			return fmt.Errorf("cannot index type %s", obj.Type())
		}
		regs[dst] = NewObject(result)
		frame.IP += 4

	case compiler.OpRegSetIndex:
		objReg, keyReg, valReg := DecodeReg3(code, frame.IP)
		obj := regs[objReg].ToObject()
		key := regs[keyReg].ToObject()
		val := regs[valReg].ToObject()

		switch o := obj.(type) {
		case *objects.Array:
			idx, ok := key.(*objects.Int)
			if !ok {
				return fmt.Errorf("array index must be integer")
			}
			if idx.Value < 0 || idx.Value >= int64(len(o.Elements)) {
				return fmt.Errorf("array index out of bounds: %d", idx.Value)
			}
			o.Elements[idx.Value] = val
		case *objects.Map:
			hashKey := key.HashKey()
			o.Pairs[hashKey] = objects.MapPair{Key: key, Value: val}
		default:
			return fmt.Errorf("cannot set index on type %s", obj.Type())
		}
		frame.IP += 4

	case compiler.OpRegIterKey:
		// Get key at current index for iteration
		// For arrays: returns the index itself
		// For maps: returns the key at that index
		dst, iterReg, indexReg := DecodeReg3(code, frame.IP)
		obj := regs[iterReg].ToObject()
		idx := regs[indexReg].ToObject()

		idxInt, ok := idx.(*objects.Int)
		if !ok {
			return fmt.Errorf("iterator index must be integer")
		}

		var result objects.Object
		switch o := obj.(type) {
		case *objects.Array:
			// For arrays, the key is the index itself
			result = idxInt
		case *objects.Map:
			// For maps, get the key at this index
			// Build a sorted list of keys for consistent ordering
			keys := make([]objects.Object, 0, len(o.Pairs))
			for _, pair := range o.Pairs {
				keys = append(keys, pair.Key)
			}
			// Sort keys by their string representation for deterministic order
			for i := 0; i < len(keys); i++ {
				for j := i + 1; j < len(keys); j++ {
					if keys[i].Inspect() > keys[j].Inspect() {
						keys[i], keys[j] = keys[j], keys[i]
					}
				}
			}
			if int(idxInt.Value) < len(keys) {
				result = keys[idxInt.Value]
			} else {
				result = objects.NULL
			}
		default:
			return fmt.Errorf("cannot iterate over type %s", obj.Type())
		}
		regs[dst] = NewObject(result)
		frame.IP += 4

	case compiler.OpRegIterValue:
		// Get value at current index for iteration
		// For arrays: returns arr[index]
		// For maps: returns map[keys[index]]
		dst, iterReg, indexReg := DecodeReg3(code, frame.IP)
		obj := regs[iterReg].ToObject()
		idx := regs[indexReg].ToObject()

		idxInt, ok := idx.(*objects.Int)
		if !ok {
			return fmt.Errorf("iterator index must be integer")
		}

		var result objects.Object
		switch o := obj.(type) {
		case *objects.Array:
			if idxInt.Value < 0 || idxInt.Value >= int64(len(o.Elements)) {
				return fmt.Errorf("array index out of bounds: %d", idxInt.Value)
			}
			result = o.Elements[idxInt.Value]
		case *objects.Map:
			// For maps, get the key at this index, then get its value
			// Use same sorted order as OpRegIterKey
			keys := make([]objects.Object, 0, len(o.Pairs))
			for _, pair := range o.Pairs {
				keys = append(keys, pair.Key)
			}
			// Sort keys by their string representation for deterministic order
			for i := 0; i < len(keys); i++ {
				for j := i + 1; j < len(keys); j++ {
					if keys[i].Inspect() > keys[j].Inspect() {
						keys[i], keys[j] = keys[j], keys[i]
					}
				}
			}
			if int(idxInt.Value) < len(keys) {
				key := keys[idxInt.Value]
				pair, exists := o.Pairs[key.HashKey()]
				if exists {
					result = pair.Value
				} else {
					result = objects.NULL
				}
			} else {
				result = objects.NULL
			}
		default:
			return fmt.Errorf("cannot iterate over type %s", obj.Type())
		}
		regs[dst] = NewObject(result)
		frame.IP += 4

	case compiler.OpRegGetField:
		// Format: OpRegGetField dst obj name_idx_hi name_idx_lo
		dst := int(code[frame.IP+1])
		objReg := int(code[frame.IP+2])
		nameIdx := int(code[frame.IP+3])<<8 | int(code[frame.IP+4])

		obj := regs[objReg].ToObject()

		// Get the field name from object constants
		nameObj := vm.objConstants[nameIdx]
		name, ok := nameObj.(*objects.String)
		if !ok {
			return fmt.Errorf("field name is not a string")
		}

		var result objects.Object

		// Handle Instance objects
		if instance, ok := obj.(*objects.Instance); ok {
			// First check for field
			if val, ok := instance.Fields[name.Value]; ok {
				result = val
			} else {
				// Check for method
				method := vm.findMethod(instance.Class, name.Value)
				if method != nil {
					result = method
				} else {
					result = objects.NULL
				}
			}
		} else if m, ok := obj.(*objects.Map); ok {
			// Handle Map objects
			key := objects.InternString(name.Value)
			if pair, exists := m.Pairs[key.HashKey()]; exists {
				result = pair.Value
			} else {
				// Check for map method
				if method, ok := objects.GetMethod(objects.MapType, name.Value); ok {
					result = method
				} else {
					result = objects.NULL
				}
			}
		} else if method, ok := objects.GetMethod(obj.Type(), name.Value); ok {
			// Handle primitive type methods (Int, Float, String, Array, Bool, Null)
			result = method
		} else if mod, ok := obj.(*objects.Module); ok {
			// Handle Module objects
			if val, ok := mod.Exports[name.Value]; ok {
				result = val
			} else {
				return fmt.Errorf("export '%s' not found in module %s", name.Value, mod.Name)
			}
		} else {
			return fmt.Errorf("cannot access property '%s' on type %s", name.Value, obj.Type())
		}

		regs[dst] = NewObject(result)
		frame.IP += 5

	case compiler.OpRegSetField:
		// Format: OpRegSetField obj val name_idx_hi name_idx_lo
		objReg := int(code[frame.IP+1])
		valReg := int(code[frame.IP+2])
		nameIdx := int(code[frame.IP+3])<<8 | int(code[frame.IP+4])

		obj := regs[objReg].ToObject()
		val := regs[valReg].ToObject()

		// Get the field name from object constants
		nameObj := vm.objConstants[nameIdx]
		name, ok := nameObj.(*objects.String)
		if !ok {
			return fmt.Errorf("field name is not a string")
		}

		// Handle Instance objects
		if instance, ok := obj.(*objects.Instance); ok {
			instance.Fields[name.Value] = val
		} else if m, ok := obj.(*objects.Map); ok {
			// Handle Map objects
			key := objects.InternString(name.Value)
			m.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: val}
		} else {
			return fmt.Errorf("cannot set field '%s' on type %s", name.Value, obj.Type())
		}

		frame.IP += 5

	case compiler.OpRegCallMethod:
		// Format: OpRegCallMethod obj name_idx_hi name_idx_lo num_args
		objReg := int(code[frame.IP+1])
		nameIdx := int(code[frame.IP+2])<<8 | int(code[frame.IP+3])
		numArgs := int(code[frame.IP+4])

		obj := regs[objReg].ToObject()

		// Get the method name from object constants
		nameObj := vm.objConstants[nameIdx]
		name, ok := nameObj.(*objects.String)
		if !ok {
			return fmt.Errorf("method name is not a string")
		}

		var method objects.Object

		// Handle Instance objects
		// Track if this is a map function value (not a built-in method)
		isMapFunctionValue := false

		if instance, ok := obj.(*objects.Instance); ok {
			// Find method
			method = vm.findMethod(instance.Class, name.Value)
			if method == nil {
				return fmt.Errorf("method '%s' not found in class", name.Value)
			}
		} else if mapObj, ok := obj.(*objects.Map); ok {
			// First check if the map has a key with this name (for callable map values)
			key := &objects.String{Value: name.Value}
			if pair, found := mapObj.Pairs[key.HashKey()]; found {
				method = pair.Value
				isMapFunctionValue = true // This is a function stored in the map, not a built-in method
			} else {
				// Otherwise check for built-in map method
				var methodFound bool
				method, methodFound = objects.GetMethod(objects.MapType, name.Value)
				if !methodFound {
					return fmt.Errorf("method '%s' not found on Map", name.Value)
				}
			}
			_ = mapObj // avoid unused variable error
		} else if m, ok := objects.GetMethod(obj.Type(), name.Value); ok {
			// Handle primitive type methods
			method = m
		} else {
			return fmt.Errorf("cannot call method '%s' on type %s", name.Value, obj.Type())
		}

		// Call the method
		// For map function values, don't add receiver as first argument
		if isMapFunctionValue {
			// Call without receiver - just use provided arguments
			switch fn := method.(type) {
			case *objects.Builtin:
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = regs[i].ToObject()
				}
				result := fn.Fn(args...)
				regs[compiler.ReturnRegister] = NewObject(result)
			case *compiler.CompiledFunction:
				if err := vm.callCompiledFunction(fn, numArgs, frame); err != nil {
					return err
				}
			case *Closure:
				if err := vm.callClosure(fn, numArgs, frame); err != nil {
					return err
				}
			default:
				return fmt.Errorf("method '%s' is not callable", name.Value)
			}
		} else {
			// Shift arguments: R0->R1, R1->R2, etc. to make room for receiver in R0
			for i := numArgs; i > 0; i-- {
				regs[i] = regs[i-1]
			}
			// Put receiver in R0
			regs[0] = NewObject(obj)

			switch fn := method.(type) {
			case *objects.Builtin:
				// Collect args including receiver
				args := make([]objects.Object, numArgs+1)
				for i := 0; i <= numArgs; i++ {
					args[i] = regs[i].ToObject()
				}
				result := fn.Fn(args...)
				regs[compiler.ReturnRegister] = NewObject(result)
			case *compiler.CompiledFunction:
				// Call with numArgs+1 (including receiver)
				if err := vm.callCompiledFunction(fn, numArgs+1, frame); err != nil {
					return err
				}
			case *Closure:
				// Call with numArgs+1 (including receiver)
				if err := vm.callClosure(fn, numArgs+1, frame); err != nil {
					return err
				}
			default:
				return fmt.Errorf("method '%s' is not callable", name.Value)
			}
		}

		frame.IP += 5

	// Literals
	case compiler.OpRegNull:
		dst := DecodeReg1(code, frame.IP)
		regs[dst] = ValueNull
		frame.IP += 2

	case compiler.OpRegTrue:
		dst := DecodeReg1(code, frame.IP)
		regs[dst] = ValueTrue
		frame.IP += 2

	case compiler.OpRegFalse:
		dst := DecodeReg1(code, frame.IP)
		regs[dst] = ValueFalse
		frame.IP += 2

	// Superinstructions
	case compiler.OpRegAddConst:
		dst, src, constIdx := decodeRegConst(code, frame.IP)
		result, ok := regs[src].Add(frame.Constants[constIdx])
		if !ok {
			return fmt.Errorf("type error in addition with constant")
		}
		regs[dst] = result
		frame.IP += 5

	case compiler.OpRegSubConst:
		dst, src, constIdx := decodeRegConst(code, frame.IP)
		result, ok := regs[src].Sub(frame.Constants[constIdx])
		if !ok {
			return fmt.Errorf("type error in subtraction with constant")
		}
		regs[dst] = result
		frame.IP += 5

	case compiler.OpRegMulConst:
		dst, src, constIdx := decodeRegConst(code, frame.IP)
		result, ok := regs[src].Mul(frame.Constants[constIdx])
		if !ok {
			return fmt.Errorf("type error in multiplication with constant")
		}
		regs[dst] = result
		frame.IP += 5

	case compiler.OpRegIncLocal:
		localIdx := DecodeReg1(code, frame.IP)
		result, ok := frame.Locals[localIdx].Add(NewInt(1))
		if ok {
			frame.Locals[localIdx] = result
		}
		frame.IP += 2

	case compiler.OpRegDecLocal:
		localIdx := DecodeReg1(code, frame.IP)
		result, ok := frame.Locals[localIdx].Sub(NewInt(1))
		if ok {
			frame.Locals[localIdx] = result
		}
		frame.IP += 2

	case compiler.OpRegLoadModule:
		_, dst, constIdx := compiler.DecodeRegInstructionConst(code[frame.IP:])
		frame.IP += 4
		// Load the module path from constants
		pathObj := vm.objConstants[constIdx]
		path, ok := pathObj.(*objects.String)
		if !ok {
			return fmt.Errorf("module path is not a string")
		}
		mod, err := vm.loadModule(path.Value, frame)
		if err != nil {
			return err
		}
		frame.Registers[dst] = NewObject(mod)

	case compiler.OpRegGetExport:
		dst, moduleReg, nameIdx := DecodeReg3(code, frame.IP)
		frame.IP += 4
		// Get the module
		moduleObj := frame.Registers[moduleReg].ToObject()
		mod, ok := moduleObj.(*objects.Module)
		if !ok {
			return fmt.Errorf("cannot get export from non-module")
		}
		// Get the export name from object constants
		nameObj := vm.objConstants[nameIdx]
		name, ok := nameObj.(*objects.String)
		if !ok {
			return fmt.Errorf("export name is not a string")
		}
		// Get the export
		export, ok := mod.Exports[name.Value]
		if !ok {
			return fmt.Errorf("export '%s' not found in module", name.Value)
		}
		frame.Registers[dst] = NewObject(export)

	case compiler.OpRegSetExport:
		_, srcReg, nameIdx := compiler.DecodeRegInstructionConst(code[frame.IP:])
		frame.IP += 4
		// Get the current module
		if vm.currentModule == nil {
			return fmt.Errorf("no current module for export")
		}
		// Get the export name from object constants
		nameObj := vm.objConstants[nameIdx]
		name, ok := nameObj.(*objects.String)
		if !ok {
			return fmt.Errorf("export name is not a string")
		}
		// Set the export
		vm.currentModule.Exports[name.Value] = frame.Registers[srcReg].ToObject()

	default:
		return fmt.Errorf("unknown register opcode: %d", op)
	}

	return nil
}

// handleRegCall handles function calls
func (vm *RegVM) handleRegCall(frame *RegFrame, code []byte) error {
	funcReg, numArgs := DecodeCall(code, frame.IP)
	frame.IP += 3

	fn := frame.Registers[funcReg]
	if fn.IsNull() {
		return fmt.Errorf("cannot call null")
	}

	// Fast path: check specific types without full ToObject call
	if closure := fn.GetClosure(); closure != nil {
		return vm.callClosure(closure, int(numArgs), frame)
	}
	if compiledFn := fn.GetCompiledFunction(); compiledFn != nil {
		return vm.callCompiledFunction(compiledFn, int(numArgs), frame)
	}

	// Slow path: generic object
	obj := fn.ToObject()
	switch fnObj := obj.(type) {
	case *Closure:
		return vm.callClosure(fnObj, int(numArgs), frame)
	case *compiler.CompiledFunction:
		return vm.callCompiledFunction(fnObj, int(numArgs), frame)
	case *objects.Builtin:
		return vm.callBuiltin(fnObj, int(numArgs), frame)
	default:
		return fmt.Errorf("cannot call %s", obj.Type())
	}
}

// callClosure calls a closure function
func (vm *RegVM) callClosure(closure *Closure, numArgs int, callerFrame *RegFrame) error {
	fn := closure.Fn
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	// Create new frame
	newFrame := NewRegFrame(fn)
	newFrame.Constants = callerFrame.Constants
	newFrame.Globals = callerFrame.Globals

	// Copy arguments from caller's R0-R7 to callee's R0-R7
	for i := 0; i < numArgs && i < compiler.NumArgRegisters; i++ {
		newFrame.Registers[i] = callerFrame.Registers[i]
	}

	// Set up free variables - directly reference the closure's FreeVarsValues
	// This allows modifications to persist across calls
	if closure.FreeVarsValues != nil {
		newFrame.FreeVars = closure.FreeVarsValues
	} else if closure.FreeVars != nil {
		// Fallback for closures created by stack VM
		for i, free := range closure.FreeVars {
			newFrame.FreeVars[i] = NewObject(free)
		}
	}

	vm.pushFrame(newFrame)
	return nil
}

// callCompiledFunction calls a compiled function
func (vm *RegVM) callCompiledFunction(fn *compiler.CompiledFunction, numArgs int, callerFrame *RegFrame) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	// Create new frame
	newFrame := NewRegFrame(fn)
	newFrame.Constants = callerFrame.Constants
	newFrame.Globals = callerFrame.Globals

	// Copy arguments
	for i := 0; i < numArgs && i < compiler.NumArgRegisters; i++ {
		newFrame.Registers[i] = callerFrame.Registers[i]
	}

	vm.pushFrame(newFrame)
	return nil
}

// callBuiltin calls a built-in function
func (vm *RegVM) callBuiltin(builtin *objects.Builtin, numArgs int, frame *RegFrame) error {
	// Collect arguments from R0-R7
	args := make([]objects.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = frame.Registers[i].ToObject()
	}

	// Call the builtin
	result := builtin.Fn(args...)

	// Store result in return register
	frame.Registers[compiler.ReturnRegister] = NewObject(result)
	return nil
}

// handleRegBuiltin handles OpRegBuiltin - direct builtin call
func (vm *RegVM) handleRegBuiltin(builtinIdx, numArgs int, frame *RegFrame) error {
	// Get the builtin function
	builtin := getBuiltin(builtinIdx)
	if builtin == nil {
		return fmt.Errorf("builtin function not found: %d", builtinIdx)
	}

	// Collect arguments from R0-R7
	args := make([]objects.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = frame.Registers[i].ToObject()
	}

	// Call the builtin
	result := builtin.Fn(args...)

	// Store result in return register
	frame.Registers[compiler.ReturnRegister] = NewObject(result)
	return nil
}

// findMethod finds a method in class hierarchy
func (vm *RegVM) findMethod(class *objects.Class, name string) objects.Object {
	for c := class; c != nil; c = c.SuperClass {
		if method, ok := c.Methods[name]; ok {
			return method
		}
	}
	return nil
}

// handleRegReturn handles return statements
func (vm *RegVM) handleRegReturn(frame *RegFrame) error {
	// Get return value from return register
	returnValue := frame.Registers[compiler.ReturnRegister]

	// Release the frame back to pool before popping
	frame.Release()

	// Pop the frame
	vm.popFrame()

	if vm.frameIndex == 0 {
		// Main function returning
		return nil
	}

	// Store return value in caller's return register
	callerFrame := vm.currentFrame()
	callerFrame.Registers[compiler.ReturnRegister] = returnValue

	return nil
}

// executeConstant handles OpConstant for mixed mode
func (vm *RegVM) executeConstant(frame *RegFrame, code []byte) error {
	constIdx := int(code[frame.IP])<<8 | int(code[frame.IP+1])
	frame.IP += 2
	vm.tempStack.Push(frame.Constants[constIdx])
	return nil
}

// executeValueBinaryOp executes a binary operation using the temp stack
func (vm *RegVM) executeValueBinaryOp(op func(a, b Value) (Value, error)) error {
	b := vm.tempStack.Pop()
	a := vm.tempStack.Pop()
	result, err := op(a, b)
	if err != nil {
		return err
	}
	vm.tempStack.Push(result)
	return nil
}

// decodeRegConst decodes dst, src, constIdx format
func decodeRegConst(code []byte, ip int) (dst, src byte, constIdx int) {
	return code[ip+1], code[ip+2], int(code[ip+3])<<8 | int(code[ip+4])
}

// loadPluginByPath loads a WASM plugin
func (vm *RegVM) loadPluginByPath(wasmPath string) (objects.Object, error) {
	// Use the existing VM's loadPluginByPath implementation
	// Create a temporary VM to use its plugin loading
	tempVM := &VM{
		loader:     vm.loader,
		sourcePath: vm.sourcePath,
	}
	return tempVM.loadPluginByPath(wasmPath)
}

// RunCodeInRegVM executes code in the register VM context
func RunCodeInRegVM(code string, args *objects.Map, regVM *RegVM) (objects.Object, error) {
	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	// Create a new compiler
	c := compiler.New()

	// If args are provided, define them as globals before compilation
	if args != nil && len(args.Pairs) > 0 {
		for _, pair := range args.Pairs {
			key, ok := pair.Key.(*objects.String)
			if !ok {
				return nil, fmt.Errorf("argument keys must be strings")
			}
			c.DefineGlobal(key.Value)
		}
	}

	if err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Create a new globals array for this execution
	newGlobals := make([]Value, compiler.GlobalsSize)

	// Set argument values in globals
	if args != nil {
		for _, pair := range args.Pairs {
			key := pair.Key.(*objects.String)
			// Find the symbol to get its index
			symbol, ok := c.ResolveSymbol(key.Value)
			if ok && symbol.Scope == compiler.GlobalScope {
				newGlobals[symbol.Index] = NewObject(pair.Value)
			}
		}
	}

	// Create a new register VM with the prepared globals
	newVM := NewRegVMWithGlobals(bytecode, newGlobals)

	// Share the loader if available
	if regVM.loader != nil {
		newVM.SetLoader(regVM.loader)
	}

	if err := newVM.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	return newVM.LastPopped().ToObject(), nil
}

// GetCallStack returns the current call stack
func (vm *RegVM) GetCallStack() string {
	var sb strings.Builder
	sb.WriteString("Call Stack:\n")

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

// StackTop returns the top of the temp stack
func (vm *RegVM) StackTop() Value {
	return vm.tempStack.Top()
}

// loadModule loads a module for the register VM
func (vm *RegVM) loadModule(importPath string, frame *RegFrame) (*objects.Module, error) {
	// Check if it's a standard library module
	if stdlib.Has(importPath) {
		stdMod := stdlib.Get(importPath)
		if stdMod == nil {
			return nil, fmt.Errorf("stdlib module not found: %s", importPath)
		}
		return &objects.Module{
			Name:    stdMod.Name,
			Exports: stdMod.Exports,
			Globals: nil,
		}, nil
	}

	// Resolve path relative to current source
	resolvedPath, err := module.Resolve(vm.sourcePath, importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve import path '%s': %v", importPath, err)
	}

	// Check cache first
	if vm.loader.HasModule(resolvedPath) {
		cachedMod, err := vm.loader.Get(resolvedPath)
		if err != nil {
			return nil, err
		}
		return &objects.Module{
			Name:    cachedMod.Name,
			Exports: cachedMod.Exports,
			Globals: cachedMod.Globals,
		}, nil
	}

	// Check for circular dependency
	if vm.loader.IsLoading(resolvedPath) {
		return nil, fmt.Errorf("circular import: %s", resolvedPath)
	}

	// Read the module file
	code, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("module not found: %s", resolvedPath)
	}

	// Parse the module
	l := lexer.New(string(code))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in module %s: %v", resolvedPath, p.Errors())
	}

	// Compile the module with register compiler
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compile error in module %s: %v", resolvedPath, err)
	}

	// Create the module object
	mod := &objects.Module{
		Name:    resolvedPath,
		Exports: make(map[string]objects.Object),
	}

	// Mark as loading
	vm.loader.MarkLoading(resolvedPath)

	// Execute the module in an isolated VM context
	moduleVM := NewRegVMWithGlobals(c.Bytecode(), make([]Value, compiler.GlobalsSize))
	moduleVM.SetLoader(vm.loader)
	moduleVM.SetSourcePath(resolvedPath)
	moduleVM.SetCurrentModule(mod)

	if err := moduleVM.Run(); err != nil {
		vm.loader.MarkDone(resolvedPath)
		return nil, fmt.Errorf("runtime error in module %s: %v", resolvedPath, err)
	}

	// Mark as done loading
	vm.loader.MarkDone(resolvedPath)

	// Cache the module
	vm.loader.Set(resolvedPath, &module.Module{
		Name:    mod.Name,
		Exports: mod.Exports,
	})

	return mod, nil
}

// handleRegTailCall handles tail call optimization
// Instead of creating a new frame, it reuses the current frame
func (vm *RegVM) handleRegTailCall(frame *RegFrame, code []byte) error {
	funcReg, numArgs := DecodeCall(code, frame.IP)
	// Don't advance IP here - we'll reset it to the new function's start

	fn := frame.Registers[funcReg]
	if fn.IsNull() {
		return fmt.Errorf("cannot call null")
	}

	// Get the function to call
	var targetFn *compiler.CompiledFunction
	var freeVars []Value

	// Fast path: check specific types
	if closure := fn.GetClosure(); closure != nil {
		targetFn = closure.Fn
		// Use FreeVarsValues for register VM, convert FreeVars if needed
		if closure.FreeVarsValues != nil {
			freeVars = closure.FreeVarsValues
		} else if len(closure.FreeVars) > 0 {
			freeVars = make([]Value, len(closure.FreeVars))
			for i, obj := range closure.FreeVars {
				freeVars[i] = NewObject(obj)
			}
		}
	} else if compiledFn := fn.GetCompiledFunction(); compiledFn != nil {
		targetFn = compiledFn
	} else {
		// Slow path: generic object
		obj := fn.ToObject()
		switch fnObj := obj.(type) {
		case *Closure:
			targetFn = fnObj.Fn
			// Use FreeVarsValues for register VM, convert FreeVars if needed
			if fnObj.FreeVarsValues != nil {
				freeVars = fnObj.FreeVarsValues
			} else if len(fnObj.FreeVars) > 0 {
				freeVars = make([]Value, len(fnObj.FreeVars))
				for i, obj := range fnObj.FreeVars {
					freeVars[i] = NewObject(obj)
				}
			}
		case *compiler.CompiledFunction:
			targetFn = fnObj
		case *objects.Builtin:
			// Builtins don't benefit from TCO, fall back to normal call
			frame.IP += 3
			return vm.callBuiltin(fnObj, int(numArgs), frame)
		default:
			return fmt.Errorf("cannot tail call %s", obj.Type())
		}
	}

	if int(numArgs) != targetFn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", targetFn.NumParameters, numArgs)
	}

	// Save arguments from R0-R7 before resetting registers
	args := make([]Value, numArgs)
	for i := 0; i < int(numArgs); i++ {
		args[i] = frame.Registers[i]
	}

	// Reset the frame for the new function
	// Clear only the registers that will be used
	numRegs := targetFn.NumRegs
	if numRegs <= 0 || numRegs > compiler.NumRegisters {
		numRegs = compiler.NumRegisters
	}
	for i := 0; i < numRegs; i++ {
		frame.Registers[i] = ValueNull
	}

	// Restore arguments to R0-R7
	for i := 0; i < int(numArgs); i++ {
		frame.Registers[i] = args[i]
	}

	// Update frame to point to the new function
	frame.Fn = targetFn
	frame.IP = 0
	frame.FreeVars = freeVars

	// Allocate locals for the new function
	if cap(frame.Locals) >= targetFn.NumLocals {
		frame.Locals = frame.Locals[:targetFn.NumLocals]
		// Clear locals
		for i := range frame.Locals {
			frame.Locals[i] = ValueNull
		}
	} else {
		frame.Locals = make([]Value, targetFn.NumLocals)
	}

	return nil
}
