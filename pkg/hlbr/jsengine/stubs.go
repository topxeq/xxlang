package jsengine

import (
	"sort"
	"strings"
)

// PropertyDescriptor represents a JavaScript property descriptor.
type PropertyDescriptor struct {
	Value          *Value
	Writable       bool
	Enumerable     bool
	Configurable   bool
	Get            *Value
	Set            *Value
}

// Bytecode represents compiled JavaScript bytecode for faster execution.
type Bytecode struct {
	// Stub: bytecode compilation not yet implemented
	Instructions []byte
}

// setupPrototypes initializes JavaScript prototype chains.
func (vm *VM) setupPrototypes() {
	// Stub: prototype chain setup not yet fully implemented
	vm.debugLog("setupPrototypes (stub)")
}

// GetObjectMethods returns the built-in Object methods map.
func GetObjectMethods(vm *VM) map[string]*Value {
	return map[string]*Value{
		"keys": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "undefined"}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "array", Arr: []*Value{}}
			}
			keys := make([]*Value, 0, len(obj.Obj))
			for k := range obj.Obj {
				keys = append(keys, &Value{Type: "string", Str: k})
			}
			return &Value{Type: "array", Arr: keys}
		}},
		"values": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "undefined"}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "array", Arr: []*Value{}}
			}
			vals := make([]*Value, 0, len(obj.Obj))
			for _, v := range obj.Obj {
				vals = append(vals, v)
			}
			return &Value{Type: "array", Arr: vals}
		}},
		"entries": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "undefined"}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "array", Arr: []*Value{}}
			}
			entries := make([]*Value, 0, len(obj.Obj))
			for k, v := range obj.Obj {
				entry := &Value{Type: "array", Arr: []*Value{
					{Type: "string", Str: k},
					v,
				}}
				entries = append(entries, entry)
			}
			return &Value{Type: "array", Arr: entries}
		}},
	}
}

// newRegExp creates a new RegExp object.
func newRegExp(pattern, flags string) *Value {
	return &Value{Type: "object", Obj: map[string]*Value{
		"pattern": {Type: "string", Str: pattern},
		"flags":   {Type: "string", Str: flags},
		"test": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"exec": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "null"}
		}},
	}}
}

// NewError creates a new Error object.
func NewError(name, message string) *Value {
	return &Value{Type: "object", Obj: map[string]*Value{
		"name":    {Type: "string", Str: name},
		"message": {Type: "string", Str: message},
	}}
}

// NewTypeError creates a new TypeError object.
func NewTypeError(message string) *Value {
	return NewError("TypeError", message)
}

// NewReferenceError creates a new ReferenceError object.
func NewReferenceError(message string) *Value {
	return NewError("ReferenceError", message)
}

// NewSyntaxError creates a new SyntaxError object.
func NewSyntaxError(message string) *Value {
	return NewError("SyntaxError", message)
}

// NewRangeError creates a new RangeError object.
func NewRangeError(message string) *Value {
	return NewError("RangeError", message)
}

// JSException represents a JavaScript exception thrown during execution.
type JSException struct {
	Value *Value
}

func (e *JSException) Error() string {
	if e.Value != nil {
		return valueToString(e.Value)
	}
	return "JSException"
}

// ToJSValue converts a Go value to a jsengine Value.
func ToJSValue(v interface{}) *Value {
	switch val := v.(type) {
	case *Value:
		return val
	case string:
		return &Value{Type: "string", Str: val}
	case float64:
		return &Value{Type: "number", Num: val}
	case int:
		return &Value{Type: "number", Num: float64(val)}
	case bool:
		return &Value{Type: "bool", Bool: val}
	case nil:
		return &Value{Type: "null"}
	default:
		return &Value{Type: "string", Str: "[object Object]"}
	}
}

// ThrowJS throws a JavaScript exception by panicking with a JSException.
func ThrowJS(val *Value) {
	panic(&JSException{Value: val})
}

// getVar retrieves a variable from the VM environment.
func (vm *VM) getVar(name string) *Value {
	return vm.env.Get(name)
}

// setVar sets a variable in the VM environment.
func (vm *VM) setVar(name string, val *Value) {
	vm.env.Set(name, val)
}

// evalUpdate evaluates an update expression (++i, i++, --i, i--).
func (vm *VM) evalUpdate(e *UpdateExpr) *Value {
	ident, ok := e.Operand.(*Ident)
	if !ok {
		return &Value{Type: "undefined"}
	}
	current := vm.getVar(ident.Name)
	if current == nil {
		current = &Value{Type: "number", Num: 0}
	}
	var newVal float64
	if current.Type == "number" {
		newVal = current.Num
	} else {
		newVal = 0
	}
	if e.Operator == "++" {
		newVal++
	} else {
		newVal--
	}
	result := &Value{Type: "number", Num: newVal}
	vm.setVar(ident.Name, result)
	if e.Prefix {
		return result
	}
	return current
}

// instanceof implements the JavaScript instanceof operator.
func (vm *VM) instanceof(obj, constructor *Value) bool {
	if constructor.Type != "function" || constructor.Func == nil {
		return false
	}
	if obj.Type != "object" {
		return false
	}
	// Stub: proper prototype chain walking not yet implemented
	return false
}

// isStringMethod checks if the given name is a string method.
func isStringMethod(name string) bool {
	switch name {
	case "charAt", "charCodeAt", "concat", "includes", "endsWith",
		"indexOf", "lastIndexOf", "match", "replace", "search",
		"slice", "split", "startsWith", "substr", "substring",
		"toLowerCase", "toUpperCase", "trim", "length":
		return true
	}
	return false
}

// callStringMethod calls a string method by name.
func callStringMethod(name string, str string, args []*Value, vm *VM) *Value {
	switch name {
	case "length":
		return &Value{Type: "number", Num: float64(len(str))}
	case "toLowerCase":
		return &Value{Type: "string", Str: strings.ToLower(str)}
	case "toUpperCase":
		return &Value{Type: "string", Str: strings.ToUpper(str)}
	case "trim":
		return &Value{Type: "string", Str: strings.TrimSpace(str)}
	case "indexOf":
		if len(args) < 1 {
			return &Value{Type: "number", Num: -1}
		}
		idx := strings.Index(str, valueToString(args[0]))
		return &Value{Type: "number", Num: float64(idx)}
	case "includes":
		if len(args) < 1 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: strings.Contains(str, valueToString(args[0]))}
	case "startsWith":
		if len(args) < 1 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: strings.HasPrefix(str, valueToString(args[0]))}
	case "endsWith":
		if len(args) < 1 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: strings.HasSuffix(str, valueToString(args[0]))}
	case "slice":
		start, end := 0, len(str)
		if len(args) >= 1 && args[0].Type == "number" {
			start = int(args[0].Num)
			if start < 0 {
				start = len(str) + start
			}
			if start < 0 {
				start = 0
			}
		}
		if len(args) >= 2 && args[1].Type == "number" {
			end = int(args[1].Num)
			if end < 0 {
				end = len(str) + end
			}
		}
		if start > len(str) {
			start = len(str)
		}
		if end > len(str) {
			end = len(str)
		}
		if start > end {
			start = end
		}
		return &Value{Type: "string", Str: str[start:end]}
	case "substr":
		start, length := 0, len(str)
		if len(args) >= 1 && args[0].Type == "number" {
			start = int(args[0].Num)
			if start < 0 {
				start = len(str) + start
			}
			if start < 0 {
				start = 0
			}
		}
		if len(args) >= 2 && args[1].Type == "number" {
			length = int(args[1].Num)
		}
		if start > len(str) {
			return &Value{Type: "string", Str: ""}
		}
		end := start + length
		if end > len(str) {
			end = len(str)
		}
		return &Value{Type: "string", Str: str[start:end]}
	case "substring":
		start, end := 0, len(str)
		if len(args) >= 1 && args[0].Type == "number" {
			start = int(args[0].Num)
			if start < 0 {
				start = 0
			}
		}
		if len(args) >= 2 && args[1].Type == "number" {
			end = int(args[1].Num)
			if end < 0 {
				end = 0
			}
		}
		if start > end {
			start, end = end, start
		}
		if end > len(str) {
			end = len(str)
		}
		return &Value{Type: "string", Str: str[start:end]}
	case "split":
		sep := ""
		if len(args) >= 1 {
			sep = valueToString(args[0])
		}
		parts := strings.Split(str, sep)
		arr := make([]*Value, len(parts))
		for i, p := range parts {
			arr[i] = &Value{Type: "string", Str: p}
		}
		return &Value{Type: "array", Arr: arr}
	case "replace":
		if len(args) < 2 {
			return &Value{Type: "string", Str: str}
		}
		return &Value{Type: "string", Str: strings.Replace(str, valueToString(args[0]), valueToString(args[1]), 1)}
	case "concat":
		var b strings.Builder
		b.WriteString(str)
		for _, arg := range args {
			b.WriteString(valueToString(arg))
		}
		return &Value{Type: "string", Str: b.String()}
	case "charAt":
		idx := 0
		if len(args) >= 1 && args[0].Type == "number" {
			idx = int(args[0].Num)
		}
		if idx < 0 || idx >= len(str) {
			return &Value{Type: "string", Str: ""}
		}
		return &Value{Type: "string", Str: string(str[idx])}
	default:
		return &Value{Type: "undefined"}
	}
}

// callArrayMethod calls an array method by name.
func callArrayMethod(name string, obj *Value, args []*Value, vm *VM) *Value {
	arr := obj.Arr
	switch name {
	case "push":
		for _, arg := range args {
			arr = append(arr, arg)
		}
		obj.Arr = arr
		return &Value{Type: "number", Num: float64(len(arr))}
	case "pop":
		if len(arr) == 0 {
			return &Value{Type: "undefined"}
		}
		val := arr[len(arr)-1]
		obj.Arr = arr[:len(arr)-1]
		return val
	case "shift":
		if len(arr) == 0 {
			return &Value{Type: "undefined"}
		}
		val := arr[0]
		obj.Arr = arr[1:]
		return val
	case "unshift":
		newArr := make([]*Value, 0, len(arr)+len(args))
		newArr = append(newArr, args...)
		newArr = append(newArr, arr...)
		obj.Arr = newArr
		return &Value{Type: "number", Num: float64(len(newArr))}
	case "slice":
		start, end := 0, len(arr)
		if len(args) >= 1 && args[0].Type == "number" {
			start = int(args[0].Num)
			if start < 0 {
				start = len(arr) + start
			}
			if start < 0 {
				start = 0
			}
		}
		if len(args) >= 2 && args[1].Type == "number" {
			end = int(args[1].Num)
			if end < 0 {
				end = len(arr) + end
			}
		}
		if start > len(arr) {
			start = len(arr)
		}
		if end > len(arr) {
			end = len(arr)
		}
		if start > end {
			start = end
		}
		return &Value{Type: "array", Arr: append([]*Value{}, arr[start:end]...)}
	case "splice":
		start, deleteCount := 0, 0
		if len(args) >= 1 && args[0].Type == "number" {
			start = int(args[0].Num)
			if start < 0 {
				start = len(arr) + start
			}
			if start < 0 {
				start = 0
			}
		}
		if start > len(arr) {
			start = len(arr)
		}
		if len(args) >= 2 && args[1].Type == "number" {
			deleteCount = int(args[1].Num)
		} else {
			deleteCount = len(arr) - start
		}
		if deleteCount < 0 {
			deleteCount = 0
		}
		if start+deleteCount > len(arr) {
			deleteCount = len(arr) - start
		}
		deleted := append([]*Value{}, arr[start:start+deleteCount]...)
		newArr := make([]*Value, 0, len(arr)-deleteCount+len(args)-2)
		newArr = append(newArr, arr[:start]...)
		if len(args) > 2 {
			newArr = append(newArr, args[2:]...)
		}
		newArr = append(newArr, arr[start+deleteCount:]...)
		obj.Arr = newArr
		return &Value{Type: "array", Arr: deleted}
	case "concat":
		result := append([]*Value{}, arr...)
		for _, arg := range args {
			if arg.Type == "array" {
				result = append(result, arg.Arr...)
			} else {
				result = append(result, arg)
			}
		}
		return &Value{Type: "array", Arr: result}
	case "join":
		sep := ","
		if len(args) >= 1 {
			sep = valueToString(args[0])
		}
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = valueToString(v)
		}
		return &Value{Type: "string", Str: strings.Join(parts, sep)}
	case "indexOf":
		if len(args) < 1 {
			return &Value{Type: "number", Num: -1}
		}
		for i, v := range arr {
			if valuesEqual(v, args[0]) {
				return &Value{Type: "number", Num: float64(i)}
			}
		}
		return &Value{Type: "number", Num: -1}
	case "includes":
		if len(args) < 1 {
			return &Value{Type: "bool", Bool: false}
		}
		for _, v := range arr {
			if valuesEqual(v, args[0]) {
				return &Value{Type: "bool", Bool: true}
			}
		}
		return &Value{Type: "bool", Bool: false}
	case "forEach":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "undefined"}
		}
		fn := args[0].Func
		for i, v := range arr {
			vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
		}
		return &Value{Type: "undefined"}
	case "map":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "array", Arr: []*Value{}}
		}
		fn := args[0].Func
		result := make([]*Value, len(arr))
		for i, v := range arr {
			result[i] = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
		}
		return &Value{Type: "array", Arr: result}
	case "filter":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "array", Arr: append([]*Value{}, arr...)}
		}
		fn := args[0].Func
		result := make([]*Value, 0)
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if isTruthy(ret) {
				result = append(result, v)
			}
		}
		return &Value{Type: "array", Arr: result}
	case "find":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "undefined"}
		}
		fn := args[0].Func
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if isTruthy(ret) {
				return v
			}
		}
		return &Value{Type: "undefined"}
	case "findIndex":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "number", Num: -1}
		}
		fn := args[0].Func
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if isTruthy(ret) {
				return &Value{Type: "number", Num: float64(i)}
			}
		}
		return &Value{Type: "number", Num: -1}
	case "every":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "bool", Bool: true}
		}
		fn := args[0].Func
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if !isTruthy(ret) {
				return &Value{Type: "bool", Bool: false}
			}
		}
		return &Value{Type: "bool", Bool: true}
	case "some":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "bool", Bool: false}
		}
		fn := args[0].Func
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if isTruthy(ret) {
				return &Value{Type: "bool", Bool: true}
			}
		}
		return &Value{Type: "bool", Bool: false}
	case "reverse":
		for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
			arr[i], arr[j] = arr[j], arr[i]
		}
		return obj
	case "sort":
		// Simple sort: convert to strings and sort lexicographically
		sort.Slice(arr, func(i, j int) bool {
			return valueToString(arr[i]) < valueToString(arr[j])
		})
		return obj
	case "length":
		return &Value{Type: "number", Num: float64(len(arr))}
	default:
		return &Value{Type: "undefined"}
	}
}
