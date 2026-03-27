// pkg/objects/builtin_check.go
// Check/validate and bytes built-in functions for Xxlang
package objects

import (
	"bytes"
	"fmt"
	"strings"
)

func init() {
	Builtins["isNil"] = &Builtin{Fn: builtinIsNil}
	Builtins["isNull"] = &Builtin{Fn: builtinIsNil}
	Builtins["isNilOrEmpty"] = &Builtin{Fn: builtinIsNilOrEmpty}
	Builtins["isNilOrErr"] = &Builtin{Fn: builtinIsNilOrErr}
	Builtins["isBytes"] = &Builtin{Fn: builtinIsBytes}
	Builtins["isChars"] = &Builtin{Fn: builtinIsChars}
	Builtins["pass"] = &Builtin{Fn: builtinPass}
	Builtins["errStrf"] = &Builtin{Fn: builtinErrStrf}
	Builtins["errf"] = &Builtin{Fn: builtinErrf}
	Builtins["errToEmpty"] = &Builtin{Fn: builtinErrToEmpty}
	Builtins["sscanf"] = &Builtin{Fn: builtinSscanf}
	Builtins["bytesStartsWith"] = &Builtin{Fn: builtinBytesStartsWith}
	Builtins["bytesEndsWith"] = &Builtin{Fn: builtinBytesEndsWith}
	Builtins["bytesContains"] = &Builtin{Fn: builtinBytesContains}
	Builtins["bytesIndex"] = &Builtin{Fn: builtinBytesIndex}
	Builtins["compareBytes"] = &Builtin{Fn: builtinCompareBytes}
	Builtins["compareText"] = &Builtin{Fn: builtinCompareText}
}

// builtinIsNil - check if value is nil/null
// Usage: isNil(value) -> bool
func builtinIsNil(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isNil. got=%d, want=1", len(args))
	}

	_, isNull := args[0].(*Null)
	return &Bool{Value: isNull}
}

// builtinIsNilOrEmpty - check if value is nil or empty
// Usage: isNilOrEmpty(value) -> bool
func builtinIsNilOrEmpty(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isNilOrEmpty. got=%d, want=1", len(args))
	}

	if _, isNull := args[0].(*Null); isNull {
		return TRUE
	}

	switch v := args[0].(type) {
	case *String:
		return &Bool{Value: len(v.Value) == 0}
	case *Array:
		return &Bool{Value: len(v.Elements) == 0}
	case *Map:
		return &Bool{Value: len(v.Pairs) == 0}
	case *Bytes:
		return &Bool{Value: len(v.Value) == 0}
	case *Chars:
		return &Bool{Value: len(v.Value) == 0}
	default:
		return FALSE
	}
}

// builtinIsNilOrErr - check if value is nil or error
// Usage: isNilOrErr(value) -> bool
func builtinIsNilOrErr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isNilOrErr. got=%d, want=1", len(args))
	}

	if _, isNull := args[0].(*Null); isNull {
		return TRUE
	}

	if _, isError := args[0].(*Error); isError {
		return TRUE
	}

	return FALSE
}

// builtinIsBytes - check if value is bytes
// Usage: isBytes(value) -> bool
func builtinIsBytes(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isBytes. got=%d, want=1", len(args))
	}

	_, isBytes := args[0].(*Bytes)
	return &Bool{Value: isBytes}
}

// builtinIsChars - check if value is chars
// Usage: isChars(value) -> bool
func builtinIsChars(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isChars. got=%d, want=1", len(args))
	}

	_, isChars := args[0].(*Chars)
	return &Bool{Value: isChars}
}

// builtinPass - do nothing and return null
// Usage: pass() -> null
func builtinPass(args ...Object) Object {
	return NULL
}

// builtinErrStrf - format error string
// Usage: errStrf(format, args...) -> string
func builtinErrStrf(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for errStrf. got=%d, want>=1", len(args))
	}

	format, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'errStrf' must be STRING, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return NewString("ERROR: " + format.Value)
	}

	formatArgs := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		formatArgs[i] = objectToGoValue(arg)
	}

	return NewString("ERROR: " + fmt.Sprintf(format.Value, formatArgs...))
}

// builtinErrf - create formatted error
// Usage: errf(format, args...) -> error
func builtinErrf(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for errf. got=%d, want>=1", len(args))
	}

	format, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'errf' must be STRING, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return &Error{Message: format.Value}
	}

	formatArgs := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		formatArgs[i] = objectToGoValue(arg)
	}

	return &Error{Message: fmt.Sprintf(format.Value, formatArgs...)}
}

// builtinErrToEmpty - convert error to empty string
// Usage: errToEmpty(value) -> value or empty string
func builtinErrToEmpty(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for errToEmpty. got=%d, want=1", len(args))
	}

	if _, isError := args[0].(*Error); isError {
		return NewString("")
	}

	if str, ok := args[0].(*String); ok {
		if strings.HasPrefix(str.Value, "ERROR:") || strings.HasPrefix(str.Value, "error:") {
			return NewString("")
		}
	}

	return args[0]
}

// builtinSscanf - parse string according to format
// Usage: sscanf(str, format) -> array
func builtinSscanf(args ...Object) Object {
	if len(args) < 2 {
		return newError("wrong number of arguments for sscanf. got=%d, want>=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'sscanf' must be STRING, got %s", args[0].Type())
	}

	format, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'sscanf' must be STRING, got %s", args[1].Type())
	}

	var results []interface{}
	_, err := fmt.Sscanf(str.Value, format.Value, results...)
	if err != nil {
		return NewArray([]Object{})
	}

	elements := make([]Object, len(results))
	for i, r := range results {
		switch v := r.(type) {
		case int:
			elements[i] = NewInt(int64(v))
		case int64:
			elements[i] = NewInt(v)
		case float64:
			elements[i] = NewFloat(v)
		case string:
			elements[i] = NewString(v)
		case bool:
			elements[i] = &Bool{Value: v}
		default:
			elements[i] = NewString(fmt.Sprintf("%v", v))
		}
	}

	return NewArray(elements)
}

// builtinBytesStartsWith - check if bytes starts with prefix
// Usage: bytesStartsWith(data, prefix) -> bool
func builtinBytesStartsWith(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bytesStartsWith. got=%d, want=2", len(args))
	}

	data, ok := args[0].(*Bytes)
	if !ok {
		return newError("first argument to 'bytesStartsWith' must be BYTES, got %s", args[0].Type())
	}

	var prefix []byte
	switch p := args[1].(type) {
	case *Bytes:
		prefix = p.Value
	case *String:
		prefix = []byte(p.Value)
	case *Array:
		prefix = make([]byte, len(p.Elements))
		for i, elem := range p.Elements {
			if n, ok := elem.(*Int); ok {
				prefix[i] = byte(n.Value)
			}
		}
	default:
		return newError("second argument to 'bytesStartsWith' must be BYTES or STRING, got %s", args[1].Type())
	}

	return &Bool{Value: bytes.HasPrefix(data.Value, prefix)}
}

// builtinBytesEndsWith - check if bytes ends with suffix
// Usage: bytesEndsWith(data, suffix) -> bool
func builtinBytesEndsWith(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bytesEndsWith. got=%d, want=2", len(args))
	}

	data, ok := args[0].(*Bytes)
	if !ok {
		return newError("first argument to 'bytesEndsWith' must be BYTES, got %s", args[0].Type())
	}

	var suffix []byte
	switch s := args[1].(type) {
	case *Bytes:
		suffix = s.Value
	case *String:
		suffix = []byte(s.Value)
	case *Array:
		suffix = make([]byte, len(s.Elements))
		for i, elem := range s.Elements {
			if n, ok := elem.(*Int); ok {
				suffix[i] = byte(n.Value)
			}
		}
	default:
		return newError("second argument to 'bytesEndsWith' must be BYTES or STRING, got %s", args[1].Type())
	}

	return &Bool{Value: bytes.HasSuffix(data.Value, suffix)}
}

// builtinBytesContains - check if bytes contains substring
// Usage: bytesContains(data, sub) -> bool
func builtinBytesContains(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bytesContains. got=%d, want=2", len(args))
	}

	data, ok := args[0].(*Bytes)
	if !ok {
		return newError("first argument to 'bytesContains' must be BYTES, got %s", args[0].Type())
	}

	var sub []byte
	switch s := args[1].(type) {
	case *Bytes:
		sub = s.Value
	case *String:
		sub = []byte(s.Value)
	case *Array:
		sub = make([]byte, len(s.Elements))
		for i, elem := range s.Elements {
			if n, ok := elem.(*Int); ok {
				sub[i] = byte(n.Value)
			}
		}
	default:
		return newError("second argument to 'bytesContains' must be BYTES or STRING, got %s", args[1].Type())
	}

	return &Bool{Value: bytes.Contains(data.Value, sub)}
}

// builtinBytesIndex - find index of bytes in bytes
// Usage: bytesIndex(data, sub) -> int
func builtinBytesIndex(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bytesIndex. got=%d, want=2", len(args))
	}

	data, ok := args[0].(*Bytes)
	if !ok {
		return newError("first argument to 'bytesIndex' must be BYTES, got %s", args[0].Type())
	}

	var sub []byte
	switch s := args[1].(type) {
	case *Bytes:
		sub = s.Value
	case *String:
		sub = []byte(s.Value)
	case *Array:
		sub = make([]byte, len(s.Elements))
		for i, elem := range s.Elements {
			if n, ok := elem.(*Int); ok {
				sub[i] = byte(n.Value)
			}
		}
	default:
		return newError("second argument to 'bytesIndex' must be BYTES or STRING, got %s", args[1].Type())
	}

	return NewInt(int64(bytes.Index(data.Value, sub)))
}

// builtinCompareBytes - compare two byte arrays
// Usage: compareBytes(a, b) -> int (-1, 0, 1)
func builtinCompareBytes(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for compareBytes. got=%d, want=2", len(args))
	}

	var a, b []byte

	switch v := args[0].(type) {
	case *Bytes:
		a = v.Value
	case *String:
		a = []byte(v.Value)
	default:
		return newError("first argument to 'compareBytes' must be BYTES or STRING, got %s", args[0].Type())
	}

	switch v := args[1].(type) {
	case *Bytes:
		b = v.Value
	case *String:
		b = []byte(v.Value)
	default:
		return newError("second argument to 'compareBytes' must be BYTES or STRING, got %s", args[1].Type())
	}

	return NewInt(int64(bytes.Compare(a, b)))
}

// builtinCompareText - compare two text values
// Usage: compareText(a, b) -> int (-1, 0, 1)
func builtinCompareText(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for compareText. got=%d, want=2", len(args))
	}

	aStr := args[0].Inspect()
	bStr := args[1].Inspect()

	if aStr < bStr {
		return NewInt(-1)
	} else if aStr > bStr {
		return NewInt(1)
	}
	return NewInt(0)
}
