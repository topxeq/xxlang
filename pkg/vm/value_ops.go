// pkg/vm/value_ops.go
// Value operations for NaN-boxed values
// These methods work directly with Value type, avoiding heap allocation
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// executeBinaryOp executes a binary operation on two values
// Used by the register VM for complex operations
func executeBinaryOp(op compiler.Opcode, left, right Value) (Value, error) {
	switch op {
	case compiler.OpRegAdd:
		result, ok := left.Add(right)
		if !ok {
			// Fallback to object path for non-numeric types
			leftObj := left.ToObject()
			rightObj := right.ToObject()
			resultObj, err := binaryAddObjects(leftObj, rightObj)
			if err != nil {
				return ValueNull, err
			}
			return NewObject(resultObj), nil
		}
		return result, nil

	case compiler.OpRegSub:
		result, ok := left.Sub(right)
		if !ok {
			return ValueNull, fmt.Errorf("cannot subtract values")
		}
		return result, nil

	case compiler.OpRegMul:
		result, ok := left.Mul(right)
		if !ok {
			return ValueNull, fmt.Errorf("cannot multiply values")
		}
		return result, nil

	case compiler.OpRegDiv:
		f, isNum := right.ToFloat()
		if isNum && f == 0 {
			return ValueNull, fmt.Errorf("division by zero")
		}
		result, ok := left.Div(right)
		if !ok {
			return ValueNull, fmt.Errorf("cannot divide values")
		}
		return result, nil

	case compiler.OpRegMod:
		result, ok := left.Mod(right)
		if !ok {
			return ValueNull, fmt.Errorf("cannot mod values")
		}
		return result, nil

	case compiler.OpRegLess:
		result, ok := left.Less(right)
		if !ok {
			return ValueFalse, fmt.Errorf("cannot compare values")
		}
		return ValueBool(result), nil

	case compiler.OpRegGreater:
		result, ok := left.Greater(right)
		if !ok {
			return ValueFalse, fmt.Errorf("cannot compare values")
		}
		return ValueBool(result), nil

	case compiler.OpRegEqual:
		result, ok := left.Equal(right)
		if !ok {
			return ValueFalse, nil
		}
		return ValueBool(result), nil

	case compiler.OpRegNotEqual:
		result, ok := left.NotEqual(right)
		if !ok {
			return ValueFalse, nil
		}
		return ValueBool(result), nil

	case compiler.OpRegLessEqual:
		return left.LessEqual(right), nil

	case compiler.OpRegGreaterEqual:
		return left.GreaterEqual(right), nil

	default:
		return ValueNull, fmt.Errorf("unknown binary operation: %d", op)
	}
}

// binaryAddObjects handles addition for objects (strings, arrays)
func binaryAddObjects(left, right objects.Object) (objects.Object, error) {
	switch left := left.(type) {
	case *objects.String:
		if right, ok := right.(*objects.String); ok {
			return &objects.String{Value: left.Value + right.Value}, nil
		}
		return nil, fmt.Errorf("type mismatch: cannot add %s to string", right.Type())

	case *objects.Array:
		if right, ok := right.(*objects.Array); ok {
			result := make([]objects.Object, len(left.Elements)+len(right.Elements))
			copy(result, left.Elements)
			copy(result[len(left.Elements):], right.Elements)
			return &objects.Array{Elements: result}, nil
		}
		return nil, fmt.Errorf("type mismatch: cannot concatenate %s to array", right.Type())

	default:
		return nil, fmt.Errorf("type %s does not support addition", left.Type())
	}
}

// executeUnaryOp executes a unary operation on a value
func executeUnaryOp(op compiler.Opcode, val Value) (Value, error) {
	switch op {
	case compiler.OpRegNeg:
		if val.IsInt() {
			i, _ := val.ToInt()
			return NewInt(-i), nil
		}
		if val.IsFloat() {
			f, _ := val.ToFloat()
			return NewFloat(-f), nil
		}
		return ValueNull, fmt.Errorf("cannot negate value")

	case compiler.OpRegNot:
		if !val.IsTruthy() {
			return ValueTrue, nil
		}
		return ValueFalse, nil

	default:
		return ValueNull, fmt.Errorf("unknown unary operation: %d", op)
	}
}
