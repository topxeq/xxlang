// pkg/vm/value_ops.go
// Value-based execution methods for hot path optimization
// These methods work directly with Value type, avoiding heap allocation
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ExecuteValueOp executes a Value-based opcode
// Returns true if the opcode was handled, false to fall back to Object path
func (vm *VM) ExecuteValueOp(op compiler.Opcode) (bool, error) {
	switch op {
	// Value-based arithmetic
	case compiler.OpValueAdd:
		return true, vm.execValueAdd()
	case compiler.OpValueSub:
		return true, vm.execValueSub()
	case compiler.OpValueMul:
		return true, vm.execValueMul()
	case compiler.OpValueDiv:
		return true, vm.execValueDiv()
	case compiler.OpValueMod:
		return true, vm.execValueMod()
	case compiler.OpValueNeg:
		return true, vm.execValueNeg()

	// Value-based comparisons
	case compiler.OpValueLess:
		return true, vm.execValueLess()
	case compiler.OpValueGreater:
		return true, vm.execValueGreater()
	case compiler.OpValueEqual:
		return true, vm.execValueEqual()
	case compiler.OpValueNotEqual:
		return true, vm.execValueNotEqual()
	case compiler.OpValueLessEqual:
		return true, vm.execValueLessEqual()
	case compiler.OpValueGreaterEqual:
		return true, vm.execValueGreaterEqual()

	// Value-based local operations
	case compiler.OpValueGetLocal:
		return true, vm.execValueGetLocal()
	case compiler.OpValueSetLocal:
		return true, vm.execValueSetLocal()
	case compiler.OpValueIncLocal:
		return true, vm.execValueIncLocal()
	case compiler.OpValueDecLocal:
		return true, vm.execValueDecLocal()
	case compiler.OpValueAddLocalConst:
		return true, vm.execValueAddLocalConst()
	case compiler.OpValueSubLocalConst:
		return true, vm.execValueSubLocalConst()
	case compiler.OpValueMulLocalConst:
		return true, vm.execValueMulLocalConst()

	// Value-based superinstructions
	case compiler.OpValueGetLocalAdd:
		return true, vm.execValueGetLocalAdd()
	case compiler.OpValueGetLocalSub:
		return true, vm.execValueGetLocalSub()
	case compiler.OpValueGetLocalMul:
		return true, vm.execValueGetLocalMul()
	case compiler.OpValueGetLocalLess:
		return true, vm.execValueGetLocalLess()
	case compiler.OpValueGetLocalGreater:
		return true, vm.execValueGetLocalGreater()
	case compiler.OpValueGetLocalEqual:
		return true, vm.execValueGetLocalEqual()
	}

	return false, nil
}

// ============================================
// Value-based arithmetic operations
// These convert objects to Values, perform the operation, and convert back
// ============================================

func (vm *VM) execValueAdd() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	// Convert to Values
	rv := NewValue(right)
	lv := NewValue(left)

	// Perform operation
	result, ok := lv.Add(rv)
	if !ok {
		// Fallback to Object path for non-numeric types (strings, etc.)
		resultObj, err := vm.binaryAdd(left, right)
		if err != nil {
			return err
		}
		vm.stack.Push(resultObj)
		return nil
	}

	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueSub() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Sub(rv)
	if !ok {
		return fmt.Errorf("cannot subtract values")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueMul() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Mul(rv)
	if !ok {
		return fmt.Errorf("cannot multiply values")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueDiv() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Div(rv)
	if !ok {
		f, isNum := rv.ToFloat()
		if isNum && f == 0 {
			return fmt.Errorf("division by zero")
		}
		return fmt.Errorf("cannot divide values")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueMod() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Mod(rv)
	if !ok {
		f, isNum := rv.ToFloat()
		if isNum && f == 0 {
			return fmt.Errorf("modulo by zero")
		}
		return fmt.Errorf("cannot compute modulo")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueNeg() error {
	operand := vm.stack.Pop()

	v := NewValue(operand)
	result, ok := v.Neg()
	if !ok {
		return fmt.Errorf("cannot negate value")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

// ============================================
// Value-based comparison operations
// ============================================

func (vm *VM) execValueLess() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Less(rv)
	if !ok {
		return fmt.Errorf("cannot compare values")
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueGreater() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Greater(rv)
	if !ok {
		return fmt.Errorf("cannot compare values")
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueEqual() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.Equal(rv)
	if !ok {
		// Fall back to object comparison
		return vm.executeComparison(compiler.OpEqual)
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueNotEqual() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	result, ok := lv.NotEqual(rv)
	if !ok {
		return vm.executeComparison(compiler.OpNotEqual)
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueLessEqual() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	less, ok1 := lv.Less(rv)
	equal, ok2 := lv.Equal(rv)
	if !ok1 || !ok2 {
		return fmt.Errorf("cannot compare values")
	}
	if less || equal {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueGreaterEqual() error {
	right := vm.stack.Pop()
	left := vm.stack.Pop()

	rv := NewValue(right)
	lv := NewValue(left)

	greater, ok1 := lv.Greater(rv)
	equal, ok2 := lv.Equal(rv)
	if !ok1 || !ok2 {
		return fmt.Errorf("cannot compare values")
	}
	if greater || equal {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

// ============================================
// Value-based local operations
// These operate directly on frame locals using Value type
// ============================================

func (vm *VM) execValueGetLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	var obj objects.Object
	if frame.This != nil && localIndex == 0 {
		obj = frame.This
	} else if frame.This != nil && localIndex > 0 {
		obj = frame.Locals[localIndex-1]
	} else {
		obj = frame.Locals[localIndex]
	}

	vm.stack.Push(obj)
	return nil
}

func (vm *VM) execValueSetLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	obj := vm.stack.Pop()

	if frame.This != nil && localIndex > 0 {
		frame.Locals[localIndex-1] = obj
	} else if frame.This == nil {
		frame.Locals[localIndex] = obj
	}

	vm.stack.Push(obj)
	return nil
}

func (vm *VM) execValueIncLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	var idx int
	if frame.This != nil && localIndex > 0 {
		idx = localIndex - 1
	} else if frame.This == nil {
		idx = localIndex
	} else {
		return fmt.Errorf("cannot increment 'this'")
	}

	obj := frame.Locals[idx]
	v := NewValue(obj)
	if v.IsInt() {
		result := NewInt(v.GetInt() + 1)
		frame.Locals[idx] = result.ToObject()
		vm.stack.Push(frame.Locals[idx])
	} else if v.IsFloat() {
		result := NewFloat(v.GetFloat() + 1)
		frame.Locals[idx] = result.ToObject()
		vm.stack.Push(frame.Locals[idx])
	} else {
		return fmt.Errorf("cannot increment non-numeric value")
	}
	return nil
}

func (vm *VM) execValueDecLocal() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	frame.IP++

	var idx int
	if frame.This != nil && localIndex > 0 {
		idx = localIndex - 1
	} else if frame.This == nil {
		idx = localIndex
	} else {
		return fmt.Errorf("cannot decrement 'this'")
	}

	obj := frame.Locals[idx]
	v := NewValue(obj)
	if v.IsInt() {
		result := NewInt(v.GetInt() - 1)
		frame.Locals[idx] = result.ToObject()
		vm.stack.Push(frame.Locals[idx])
	} else if v.IsFloat() {
		result := NewFloat(v.GetFloat() - 1)
		frame.Locals[idx] = result.ToObject()
		vm.stack.Push(frame.Locals[idx])
	} else {
		return fmt.Errorf("cannot decrement non-numeric value")
	}
	return nil
}

func (vm *VM) execValueAddLocalConst() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	constIndex := int(frame.Instructions()[frame.IP+2])<<8 | int(frame.Instructions()[frame.IP+3])
	frame.IP += 3

	var idx int
	if frame.This != nil && localIndex > 0 {
		idx = localIndex - 1
	} else if frame.This == nil {
		idx = localIndex
	} else {
		return fmt.Errorf("cannot add to 'this'")
	}

	// Get constant
	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}
	constObj := constants[constIndex]

	localObj := frame.Locals[idx]
	lv := NewValue(localObj)
	cv := NewValue(constObj)

	result, ok := lv.Add(cv)
	if !ok {
		return fmt.Errorf("cannot add constant to local")
	}
	frame.Locals[idx] = result.ToObject()
	vm.stack.Push(frame.Locals[idx])
	return nil
}

func (vm *VM) execValueSubLocalConst() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	constIndex := int(frame.Instructions()[frame.IP+2])<<8 | int(frame.Instructions()[frame.IP+3])
	frame.IP += 3

	var idx int
	if frame.This != nil && localIndex > 0 {
		idx = localIndex - 1
	} else if frame.This == nil {
		idx = localIndex
	} else {
		return fmt.Errorf("cannot subtract from 'this'")
	}

	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}
	constObj := constants[constIndex]

	localObj := frame.Locals[idx]
	lv := NewValue(localObj)
	cv := NewValue(constObj)

	result, ok := lv.Sub(cv)
	if !ok {
		return fmt.Errorf("cannot subtract constant from local")
	}
	frame.Locals[idx] = result.ToObject()
	vm.stack.Push(frame.Locals[idx])
	return nil
}

func (vm *VM) execValueMulLocalConst() error {
	frame := vm.currentFrame()
	localIndex := int(frame.Instructions()[frame.IP+1])
	constIndex := int(frame.Instructions()[frame.IP+2])<<8 | int(frame.Instructions()[frame.IP+3])
	frame.IP += 3

	var idx int
	if frame.This != nil && localIndex > 0 {
		idx = localIndex - 1
	} else if frame.This == nil {
		idx = localIndex
	} else {
		return fmt.Errorf("cannot multiply 'this'")
	}

	constants := frame.Constants
	if constants == nil {
		constants = vm.constants
	}
	constObj := constants[constIndex]

	localObj := frame.Locals[idx]
	lv := NewValue(localObj)
	cv := NewValue(constObj)

	result, ok := lv.Mul(cv)
	if !ok {
		return fmt.Errorf("cannot multiply local by constant")
	}
	frame.Locals[idx] = result.ToObject()
	vm.stack.Push(frame.Locals[idx])
	return nil
}

// ============================================
// Value-based superinstructions
// These combine multiple operations for zero-allocation hot paths
// ============================================

func (vm *VM) execValueGetLocalAdd() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Add(rv)
	if !ok {
		// Fallback to object path
		resultObj, err := vm.binaryAdd(v1, v2)
		if err != nil {
			return err
		}
		vm.stack.Push(resultObj)
		return nil
	}

	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueGetLocalSub() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Sub(rv)
	if !ok {
		return fmt.Errorf("cannot subtract locals")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueGetLocalMul() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Mul(rv)
	if !ok {
		return fmt.Errorf("cannot multiply locals")
	}
	vm.stack.Push(result.ToObject())
	return nil
}

func (vm *VM) execValueGetLocalLess() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Less(rv)
	if !ok {
		return fmt.Errorf("cannot compare locals")
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueGetLocalGreater() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Greater(rv)
	if !ok {
		return fmt.Errorf("cannot compare locals")
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

func (vm *VM) execValueGetLocalEqual() error {
	frame := vm.currentFrame()
	idx1 := int(frame.Instructions()[frame.IP+1])
	idx2 := int(frame.Instructions()[frame.IP+2])
	frame.IP += 2

	var v1, v2 objects.Object
	if frame.This != nil {
		if idx1 == 0 {
			v1 = frame.This
		} else {
			v1 = frame.Locals[idx1-1]
		}
		if idx2 == 0 {
			v2 = frame.This
		} else {
			v2 = frame.Locals[idx2-1]
		}
	} else {
		v1 = frame.Locals[idx1]
		v2 = frame.Locals[idx2]
	}

	lv := NewValue(v1)
	rv := NewValue(v2)

	result, ok := lv.Equal(rv)
	if !ok {
		// Fall back to object equality
		if v1 == v2 {
			vm.stack.Push(objects.TRUE)
		} else {
			vm.stack.Push(objects.FALSE)
		}
		return nil
	}
	if result {
		vm.stack.Push(objects.TRUE)
	} else {
		vm.stack.Push(objects.FALSE)
	}
	return nil
}

// binaryAdd handles the fallback case for addition (strings, etc.)
func (vm *VM) binaryAdd(left, right objects.Object) (objects.Object, error) {
	leftIsStr := isString(left)
	rightIsStr := isString(right)

	if leftIsStr || rightIsStr {
		leftStr := objectToString(left)
		rightStr := objectToString(right)
		return &objects.String{Value: leftStr + rightStr}, nil
	}

	return nil, fmt.Errorf("cannot add values")
}

// Helper: convert object to string
func objectToString(obj objects.Object) string {
	if obj == nil {
		return "null"
	}
	switch o := obj.(type) {
	case *objects.String:
		return o.Value
	case *objects.Int:
		return o.Inspect()
	case *objects.Float:
		return o.Inspect()
	case *objects.Bool:
		return o.Inspect()
	default:
		return obj.Inspect()
	}
}
