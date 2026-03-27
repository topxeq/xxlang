// pkg/objects/builtin_util.go
// Utility built-in functions for Xxlang
package objects

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func init() {
	Builtins["sprintf"] = &Builtin{Fn: builtinSprintf}
	Builtins["toBool"] = &Builtin{Fn: builtinToBool}
	Builtins["toInt"] = &Builtin{Fn: builtinToInt}
	Builtins["toFloat"] = &Builtin{Fn: builtinToFloat}
	Builtins["isUndefined"] = &Builtin{Fn: builtinIsUndefined}
	Builtins["isCallable"] = &Builtin{Fn: builtinIsCallable}
	Builtins["isIterable"] = &Builtin{Fn: builtinIsIterable}
	Builtins["isError"] = &Builtin{Fn: builtinIsError}
	Builtins["error"] = &Builtin{Fn: builtinBuiltinError}
	Builtins["getErrStr"] = &Builtin{Fn: builtinGetErrStr}
	Builtins["isErrStr"] = &Builtin{Fn: builtinIsErrStr}
	Builtins["typeCode"] = &Builtin{Fn: builtinTypeCode}
	Builtins["clone"] = &Builtin{Fn: builtinClone}
	Builtins["swap"] = &Builtin{Fn: builtinSwap}
	Builtins["coalesce"] = &Builtin{Fn: builtinCoalesce}
	Builtins["defaultVal"] = &Builtin{Fn: builtinDefaultVal}
}

// builtinSprintf - format string and return result
// Usage: sprintf(format, args...) -> string
func builtinSprintf(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for sprintf. got=%d, want>=1", len(args))
	}

	format, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'sprintf' must be STRING, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return format
	}

	formatArgs := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		formatArgs[i] = objectToInterface(arg)
	}

	return NewString(fmt.Sprintf(format.Value, formatArgs...))
}

// builtinToBool - convert to boolean
// Usage: toBool(value) -> bool
func builtinToBool(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toBool. got=%d, want=1", len(args))
	}

	return args[0].ToBool()
}

// builtinToInt - convert to integer
// Usage: toInt(value) -> int
//
//	toInt(value, base) -> int
func builtinToInt(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for toInt. got=%d, want=1 or 2", len(args))
	}

	base := 10
	if len(args) == 2 {
		b, ok := args[1].(*Int)
		if !ok {
			return newError("base must be INT, got %s", args[1].Type())
		}
		base = int(b.Value)
	}

	switch v := args[0].(type) {
	case *Int:
		return v
	case *Float:
		return NewInt(int64(v.Value))
	case *String:
		val, err := strconv.ParseInt(v.Value, base, 64)
		if err != nil {
			return newError("cannot convert '%s' to int: %v", v.Value, err)
		}
		return NewInt(val)
	case *Bool:
		if v.Value {
			return NewInt(1)
		}
		return NewInt(0)
	case *BigInt:
		if v.Value.IsInt64() {
			return NewInt(v.Value.Int64())
		}
		return newError("bigint value too large for int")
	default:
		return newError("cannot convert %s to int", args[0].Type())
	}
}

// builtinToFloat - convert to float
// Usage: toFloat(value) -> float
func builtinToFloat(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
	}

	switch v := args[0].(type) {
	case *Float:
		return v
	case *Int:
		return NewFloat(float64(v.Value))
	case *String:
		val, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return newError("cannot convert '%s' to float: %v", v.Value, err)
		}
		return NewFloat(val)
	case *Bool:
		if v.Value {
			return NewFloat(1.0)
		}
		return NewFloat(0.0)
	case *BigInt:
		f, _ := v.Value.Float64()
		return NewFloat(f)
	case *BigFloat:
		f, _ := v.Value.Float64()
		return NewFloat(f)
	default:
		return newError("cannot convert %s to float", args[0].Type())
	}
}

// builtinIsUndefined - check if value is undefined/null
// Usage: isUndefined(value) -> bool
func builtinIsUndefined(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isUndefined. got=%d, want=1", len(args))
	}

	_, isNull := args[0].(*Null)
	return &Bool{Value: isNull}
}

// builtinIsCallable - check if value is callable
// Usage: isCallable(value) -> bool
func builtinIsCallable(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isCallable. got=%d, want=1", len(args))
	}

	tag := args[0].TypeTag()
	switch tag {
	case TagFunction, TagBuiltin, TagClosure:
		return TRUE
	default:
		return FALSE
	}
}

// builtinIsIterable - check if value is iterable
// Usage: isIterable(value) -> bool
func builtinIsIterable(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isIterable. got=%d, want=1", len(args))
	}

	switch args[0].(type) {
	case *Array, *String, *Map, *OrderedMap, *Chars, *Bytes:
		return TRUE
	default:
		return FALSE
	}
}

// builtinIsError - check if value is an error
// Usage: isError(value) -> bool
func builtinIsError(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isError. got=%d, want=1", len(args))
	}

	_, isError := args[0].(*Error)
	return &Bool{Value: isError}
}

// builtinBuiltinError - create an error object
// Usage: error(message) -> error
func builtinBuiltinError(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for error. got=%d, want=1", len(args))
	}

	msg, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'error' must be STRING, got %s", args[0].Type())
	}

	return &Error{Message: msg.Value}
}

// builtinGetErrStr - get error string from error object
// Usage: getErrStr(value) -> string
func builtinGetErrStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getErrStr. got=%d, want=1", len(args))
	}

	switch v := args[0].(type) {
	case *Error:
		return NewString(v.Message)
	case *String:
		return v
	default:
		return NewString("")
	}
}

// builtinIsErrStr - check if string is an error message
// Usage: isErrStr(value) -> bool
func builtinIsErrStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isErrStr. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return FALSE
	}

	return &Bool{Value: strings.HasPrefix(str.Value, "ERROR:") || strings.HasPrefix(str.Value, "error:")}
}

// builtinTypeCode - get type code of an object
// Usage: typeCode(value) -> int
func builtinTypeCode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for typeCode. got=%d, want=1", len(args))
	}

	return NewInt(int64(args[0].TypeTag()))
}

// builtinClone - deep clone an object
// Usage: clone(value) -> value
func builtinClone(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
	}

	return deepClone(args[0])
}

// deepClone performs a deep clone of an object
func deepClone(obj Object) Object {
	switch v := obj.(type) {
	case *Null:
		return NULL
	case *Bool:
		return &Bool{Value: v.Value}
	case *Int:
		return NewInt(v.Value)
	case *Float:
		return NewFloat(v.Value)
	case *String:
		return NewString(v.Value)
	case *Array:
		elements := make([]Object, len(v.Elements))
		for i, elem := range v.Elements {
			elements[i] = deepClone(elem)
		}
		return NewArray(elements)
	case *Map:
		pairs := make(map[HashKey]MapPair, len(v.Pairs))
		for key, pair := range v.Pairs {
			pairs[key] = MapPair{
				Key:   deepClone(pair.Key),
				Value: deepClone(pair.Value),
			}
		}
		return NewMap(pairs)
	case *Builtin:
		return v
	case *Function:
		return v
	case *CompiledFunction:
		return v
	default:
		return obj
	}
}

// builtinSwap - swap two values in an array
// Usage: swap(arr, i, j) -> array
func builtinSwap(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments for swap. got=%d, want=3", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'swap' must be ARRAY, got %s", args[0].Type())
	}

	i, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'swap' must be INT, got %s", args[1].Type())
	}

	j, ok := args[2].(*Int)
	if !ok {
		return newError("third argument to 'swap' must be INT, got %s", args[2].Type())
	}

	idxI, idxJ := int(i.Value), int(j.Value)

	if idxI < 0 || idxI >= len(arr.Elements) {
		return newError("index %d out of bounds", idxI)
	}
	if idxJ < 0 || idxJ >= len(arr.Elements) {
		return newError("index %d out of bounds", idxJ)
	}

	result := make([]Object, len(arr.Elements))
	copy(result, arr.Elements)
	result[idxI], result[idxJ] = result[idxJ], result[idxI]

	return NewArray(result)
}

// builtinCoalesce - return first non-null/non-error value
// Usage: coalesce(val1, val2, ...) -> value
func builtinCoalesce(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for coalesce. got=%d, want>=1", len(args))
	}

	for _, arg := range args {
		if _, isNull := arg.(*Null); isNull {
			continue
		}
		if _, isError := arg.(*Error); isError {
			continue
		}
		return arg
	}

	return NULL
}

// builtinDefaultVal - return default value if null or error
// Usage: defaultVal(value, defaultValue) -> value
func builtinDefaultVal(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for defaultVal. got=%d, want=2", len(args))
	}

	if _, isNull := args[0].(*Null); isNull {
		return args[1]
	}
	if _, isError := args[0].(*Error); isError {
		return args[1]
	}

	return args[0]
}

// Helper functions

func objectToInterface(obj Object) interface{} {
	switch v := obj.(type) {
	case *Int:
		return v.Value
	case *Float:
		return v.Value
	case *String:
		return v.Value
	case *Bool:
		return v.Value
	case *Null:
		return nil
	case *Array:
		arr := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			arr[i] = objectToInterface(elem)
		}
		return arr
	case *Map:
		m := make(map[string]interface{})
		for _, pair := range v.Pairs {
			key := pair.Key.Inspect()
			m[key] = objectToInterface(pair.Value)
		}
		return m
	default:
		return v.Inspect()
	}
}

// objectToString converts object to string representation
func objectToString(obj Object) string {
	switch v := obj.(type) {
	case *String:
		return v.Value
	case *Int:
		return strconv.FormatInt(v.Value, 10)
	case *Float:
		return strconv.FormatFloat(v.Value, 'f', -1, 64)
	case *Bool:
		return strconv.FormatBool(v.Value)
	default:
		return v.Inspect()
	}
}

// Type code constants for typeCode function
const (
	TypeCodeNull TypeTag = iota
	TypeCodeBool
	TypeCodeInt
	TypeCodeFloat
	TypeCodeString
	TypeCodeArray
	TypeCodeMap
	TypeCodeFunction
	TypeCodeBuiltin
	TypeCodeError
	TypeCodeBytes
	TypeCodeChars
	TypeCodeBigInt
	TypeCodeBigFloat
	TypeCodeCompiledFunction
	TypeCodeOrderedMap
)

// getTypeCode returns the type code for an object
func getTypeCode(obj Object) int {
	switch obj.(type) {
	case *Null:
		return int(TypeCodeNull)
	case *Bool:
		return int(TypeCodeBool)
	case *Int:
		return int(TypeCodeInt)
	case *Float:
		return int(TypeCodeFloat)
	case *String:
		return int(TypeCodeString)
	case *Array:
		return int(TypeCodeArray)
	case *Map:
		return int(TypeCodeMap)
	case *Function:
		return int(TypeCodeFunction)
	case *Builtin:
		return int(TypeCodeBuiltin)
	case *Error:
		return int(TypeCodeError)
	case *Bytes:
		return int(TypeCodeBytes)
	case *Chars:
		return int(TypeCodeChars)
	case *BigInt:
		return int(TypeCodeBigInt)
	case *BigFloat:
		return int(TypeCodeBigFloat)
	case *CompiledFunction:
		return int(TypeCodeCompiledFunction)
	case *OrderedMap:
		return int(TypeCodeOrderedMap)
	default:
		return 0
	}
}

// isNil checks if an object is nil (for reflection)
func isNil(obj Object) bool {
	if obj == nil {
		return true
	}

	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return val.IsNil()
	}
	return false
}
