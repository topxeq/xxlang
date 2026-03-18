// pkg/vm/value_vm.go
// Value-optimized VM execution for hot paths
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ValueVM is a lightweight VM optimized for numeric operations using Value type
// It's designed for hot loops and arithmetic-heavy code
type ValueVM struct {
	constants  []Value
	stack      *ValueStack
	locals     []Value
	globals    []Value
	ip         int
	code       []byte
	err        error
}

// NewValueVM creates a Value-optimized VM from compiled bytecode
func NewValueVM(bytecode *compiler.Bytecode) *ValueVM {
	// Convert constants to Values
	constants := make([]Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = NewValue(c)
	}

	return &ValueVM{
		constants: constants,
		stack:     NewValueStack(),
		locals:    make([]Value, 256),
		globals:   make([]Value, GlobalsSize),
		code:      bytecode.Instructions,
		ip:        0,
	}
}

// Run executes the bytecode
func (vm *ValueVM) Run() error {
	for vm.ip < len(vm.code) {
		op := compiler.Opcode(vm.code[vm.ip])
		vm.ip++

		if err := vm.executeOp(op); err != nil {
			return err
		}
	}
	return nil
}

// Result returns the last value on the stack
func (vm *ValueVM) Result() Value {
	return vm.stack.LastPopped()
}

// ResultObject returns the result as an Object
func (vm *ValueVM) ResultObject() objects.Object {
	return vm.stack.LastPopped().ToObject()
}

func (vm *ValueVM) executeOp(op compiler.Opcode) error {
	switch op {
	// Stack operations
	case compiler.OpConstant:
		idx := int(vm.code[vm.ip])<<8 | int(vm.code[vm.ip+1])
		vm.ip += 2
		vm.stack.MustPush(vm.constants[idx])

	case compiler.OpPop:
		vm.stack.Pop()

	case compiler.OpNull:
		vm.stack.MustPush(ValueNull)

	case compiler.OpTrue:
		vm.stack.MustPush(ValueTrue)

	case compiler.OpFalse:
		vm.stack.MustPush(ValueFalse)

	// Arithmetic - Value-based
	case compiler.OpAdd:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Add(right)
		if !ok {
			return fmt.Errorf("cannot add")
		}
		vm.stack.MustPush(result)

	case compiler.OpSub:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Sub(right)
		if !ok {
			return fmt.Errorf("cannot subtract")
		}
		vm.stack.MustPush(result)

	case compiler.OpMul:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Mul(right)
		if !ok {
			return fmt.Errorf("cannot multiply")
		}
		vm.stack.MustPush(result)

	case compiler.OpDiv:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Div(right)
		if !ok {
			return fmt.Errorf("cannot divide")
		}
		vm.stack.MustPush(result)

	case compiler.OpMod:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Mod(right)
		if !ok {
			return fmt.Errorf("cannot modulo")
		}
		vm.stack.MustPush(result)

	case compiler.OpNeg:
		val := vm.stack.Pop()
		result, ok := val.Neg()
		if !ok {
			return fmt.Errorf("cannot negate")
		}
		vm.stack.MustPush(result)

	// Comparisons - Value-based
	case compiler.OpLess:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Less(right)
		if !ok {
			return fmt.Errorf("cannot compare")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpGreater:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Greater(right)
		if !ok {
			return fmt.Errorf("cannot compare")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpEqual:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.Equal(right)
		if !ok {
			return fmt.Errorf("cannot compare")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpNotEqual:
		right := vm.stack.Pop()
		left := vm.stack.Pop()
		result, ok := left.NotEqual(right)
		if !ok {
			return fmt.Errorf("cannot compare")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	// Local variables
	case compiler.OpGetLocal:
		idx := int(vm.code[vm.ip])
		vm.ip++
		vm.stack.MustPush(vm.locals[idx])

	case compiler.OpSetLocal:
		idx := int(vm.code[vm.ip])
		vm.ip++
		val := vm.stack.Pop()
		vm.locals[idx] = val
		vm.stack.MustPush(val)

	case compiler.OpIncLocal:
		idx := int(vm.code[vm.ip])
		vm.ip++
		val := vm.locals[idx]
		if val.IsInt() {
			vm.locals[idx] = NewInt(val.GetInt() + 1)
			vm.stack.MustPush(vm.locals[idx])
		} else {
			return fmt.Errorf("cannot increment non-int")
		}

	case compiler.OpDecLocal:
		idx := int(vm.code[vm.ip])
		vm.ip++
		val := vm.locals[idx]
		if val.IsInt() {
			vm.locals[idx] = NewInt(val.GetInt() - 1)
			vm.stack.MustPush(vm.locals[idx])
		} else {
			return fmt.Errorf("cannot decrement non-int")
		}

	// Jump instructions
	case compiler.OpJump:
		pos := int(vm.code[vm.ip])<<8 | int(vm.code[vm.ip+1])
		vm.ip = pos

	case compiler.OpJumpIfFalse:
		pos := int(vm.code[vm.ip])<<8 | int(vm.code[vm.ip+1])
		vm.ip += 2
		condition := vm.stack.Pop()
		if condition == ValueFalse || condition.IsNull() {
			vm.ip = pos
		}

	case compiler.OpJumpIfTrue:
		pos := int(vm.code[vm.ip])<<8 | int(vm.code[vm.ip+1])
		vm.ip += 2
		condition := vm.stack.Pop()
		if condition.IsTruthy() {
			vm.ip = pos
		}

	// Superinstructions
	case compiler.OpGetLocalAdd:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Add(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot add locals")
		}
		vm.stack.MustPush(result)

	case compiler.OpGetLocalSub:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Sub(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot subtract locals")
		}
		vm.stack.MustPush(result)

	case compiler.OpGetLocalMul:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Mul(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot multiply locals")
		}
		vm.stack.MustPush(result)

	case compiler.OpGetLocalLess:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Less(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot compare locals")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpGetLocalGreater:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Greater(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot compare locals")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpGetLocalEqual:
		idx1 := int(vm.code[vm.ip])
		idx2 := int(vm.code[vm.ip+1])
		vm.ip += 2
		result, ok := vm.locals[idx1].Equal(vm.locals[idx2])
		if !ok {
			return fmt.Errorf("cannot compare locals")
		}
		if result {
			vm.stack.MustPush(ValueTrue)
		} else {
			vm.stack.MustPush(ValueFalse)
		}

	case compiler.OpAddLocalConst:
		localIdx := int(vm.code[vm.ip])
		constIdx := int(vm.code[vm.ip+1])<<8 | int(vm.code[vm.ip+2])
		vm.ip += 3
		result, ok := vm.locals[localIdx].Add(vm.constants[constIdx])
		if !ok {
			return fmt.Errorf("cannot add constant")
		}
		vm.locals[localIdx] = result
		vm.stack.MustPush(result)

	case compiler.OpSubLocalConst:
		localIdx := int(vm.code[vm.ip])
		constIdx := int(vm.code[vm.ip+1])<<8 | int(vm.code[vm.ip+2])
		vm.ip += 3
		result, ok := vm.locals[localIdx].Sub(vm.constants[constIdx])
		if !ok {
			return fmt.Errorf("cannot subtract constant")
		}
		vm.locals[localIdx] = result
		vm.stack.MustPush(result)

	case compiler.OpMulLocalConst:
		localIdx := int(vm.code[vm.ip])
		constIdx := int(vm.code[vm.ip+1])<<8 | int(vm.code[vm.ip+2])
		vm.ip += 3
		result, ok := vm.locals[localIdx].Mul(vm.constants[constIdx])
		if !ok {
			return fmt.Errorf("cannot multiply by constant")
		}
		vm.locals[localIdx] = result
		vm.stack.MustPush(result)

	default:
		return fmt.Errorf("unsupported opcode: %d", op)
	}

	return nil
}
