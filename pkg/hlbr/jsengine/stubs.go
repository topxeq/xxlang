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
// Sets up Object.prototype, Array.prototype, Function.prototype, String.prototype,
// Number.prototype, Boolean.prototype, and Error.prototype with their standard methods.
func (vm *VM) setupPrototypes() {
	// Object.prototype
	objectProto := &Value{Type: "object", Obj: map[string]*Value{
		"hasOwnProperty": {Type: "native", Native: func(args []*Value) *Value {
			// hasOwnProperty(prop) checks if prop is an own property of this
			if len(args) < 1 {
				return &Value{Type: "bool", Bool: false}
			}
			// The 'this' value should be passed via ThisBinding
			// For now, return false as a safe default
			return &Value{Type: "bool", Bool: false}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "[object Object]"}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			// Returns the primitive value of this
			return &Value{Type: "undefined"}
		}},
		"isPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"propertyIsEnumerable": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"toLocaleString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "[object Object]"}
		}},
	}}
	vm.env.Define("ObjectPrototype", objectProto)

	// Array.prototype
	arrayProto := &Value{Type: "object", Obj: map[string]*Value{
		"push": {Type: "native", Native: func(args []*Value) *Value {
			// Array.prototype.push is called as arr.push(...items)
			// The 'this' binding should contain the array
			return &Value{Type: "number", Num: 0}
		}},
		"pop": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"shift": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"unshift": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}},
		"slice": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"splice": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"indexOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: -1}
		}},
		"includes": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"find": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"findIndex": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: -1}
		}},
		"forEach": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"map": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"filter": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"reduce": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"reduceRight": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"every": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: true}
		}},
		"some": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"join": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: ""}
		}},
		"concat": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"reverse": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"sort": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"flat": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"flatMap": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"fill": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"copyWithin": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"keys": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"values": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"entries": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: ""}
		}},
	}}
	vm.env.Define("ArrayPrototype", arrayProto)

	// Function.prototype
	functionProto := &Value{Type: "object", Obj: map[string]*Value{
		"call": {Type: "native", Native: func(args []*Value) *Value {
			// Function.prototype.call(thisArg, ...args)
			// The function to call is passed via ThisBinding
			return &Value{Type: "undefined"}
		}},
		"apply": {Type: "native", Native: func(args []*Value) *Value {
			// Function.prototype.apply(thisArg, argsArray)
			return &Value{Type: "undefined"}
		}},
		"bind": {Type: "native", Native: func(args []*Value) *Value {
			// Function.prototype.bind(thisArg, ...args)
			// Creates a new function with the given this value and initial arguments
			// The original function is in the ThisBinding of the bound function
			// For now, return a native function that will be handled by evalMember
			return &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				return &Value{Type: "undefined"}
			}}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "function () { [native code] }"}
		}},
	}}
	vm.env.Define("FunctionPrototype", functionProto)

	// String.prototype
	stringProto := &Value{Type: "object", Obj: map[string]*Value{
		"trim": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"trimStart": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"trimEnd": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"split": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"replace": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"replaceAll": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"match": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "null"}
		}},
		"search": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: -1}
		}},
		"startsWith": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"endsWith": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"includes": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"repeat": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"padStart": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"padEnd": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"toLowerCase": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"toUpperCase": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"charAt": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"charCodeAt": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}},
		"substring": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"slice": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"indexOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: -1}
		}},
		"lastIndexOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: -1}
		}},
		"concat": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"localeCompare": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}},
		"normalize": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string"}
		}},
	}}
	vm.env.Define("StringPrototype", stringProto)

	// Number.prototype
	numberProto := &Value{Type: "object", Obj: map[string]*Value{
		"toFixed": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "0"}
		}},
		"toPrecision": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "0"}
		}},
		"toExponential": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "0e+0"}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "0"}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}},
	}}
	vm.env.Define("NumberPrototype", numberProto)

	// Boolean.prototype
	booleanProto := &Value{Type: "object", Obj: map[string]*Value{
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "false"}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
	}}
	vm.env.Define("BooleanPrototype", booleanProto)

	// Error.prototype
	errorProto := &Value{Type: "object", Obj: map[string]*Value{
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "Error"}
		}},
		"message": {Type: "string", Str: ""},
		"name": {Type: "string", Str: "Error"},
	}}
	vm.env.Define("ErrorPrototype", errorProto)

	// Set up prototype links for built-in types
	// When new arrays are created, they should inherit from ArrayPrototype
	// When new functions are created, they should inherit from FunctionPrototype
	// etc.
	// We store these for use in wrapNode and evalNew
}

// GetObjectMethods returns the built-in Object methods map.
func GetObjectMethods(vm *VM) map[string]*Value {
	return map[string]*Value{
		"keys": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			keys := make([]*Value, 0, len(obj.Obj))
			for k := range obj.Obj {
				keys = append(keys, &Value{Type: "string", Str: k})
			}
			return &Value{Type: "object", Arr: keys}
		}},
		"values": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			vals := make([]*Value, 0, len(obj.Obj))
			for _, v := range obj.Obj {
				vals = append(vals, v)
			}
			return &Value{Type: "object", Arr: vals}
		}},
		"entries": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[0]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			entries := make([]*Value, 0, len(obj.Obj))
			for k, v := range obj.Obj {
				entry := &Value{Type: "object", Arr: []*Value{
					{Type: "string", Str: k},
					v,
				}}
				entries = append(entries, entry)
			}
			return &Value{Type: "object", Arr: entries}
		}},
		// Object.defineProperty(obj, prop, descriptor)
		// Critical for Vue 2 reactivity system
		"defineProperty": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 3 {
				return args[0]
			}
			obj := args[0]
			propName := valueToString(args[1])
			descObj := args[2]

			if obj.Type != "object" {
				return obj
			}
			if obj.Obj == nil {
				obj.Obj = make(map[string]*Value)
			}
			if obj.Descriptors == nil {
				obj.Descriptors = make(map[string]*PropertyDescriptor)
			}

			desc := &PropertyDescriptor{
				Enumerable:   true,
				Configurable: true,
			}

			if descObj.Type == "object" && descObj.Obj != nil {
				if v, ok := descObj.Obj["enumerable"]; ok {
					desc.Enumerable = v.Bool
				}
				if v, ok := descObj.Obj["configurable"]; ok {
					desc.Configurable = v.Bool
				}
				if v, ok := descObj.Obj["writable"]; ok {
					desc.Writable = v.Bool
				}
				if v, ok := descObj.Obj["value"]; ok {
					desc.Value = v
				}
				if v, ok := descObj.Obj["get"]; ok {
					if v.Type == "function" || v.Type == "native" {
						desc.Get = v
					}
				}
				if v, ok := descObj.Obj["set"]; ok {
					if v.Type == "function" || v.Type == "native" {
						desc.Set = v
					}
				}
			}

			// Store the descriptor
			obj.Descriptors[propName] = desc

			// If it's a data descriptor (value/writable), also store in Obj for direct access
			if desc.Value != nil {
				obj.Obj[propName] = desc.Value
			}

			return obj
		}},
		// Object.getOwnPropertyDescriptor(obj, prop)
		"getOwnPropertyDescriptor": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 2 {
				return &Value{Type: "undefined"}
			}
			obj := args[0]
			propName := valueToString(args[1])

			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "undefined"}
			}

			// Check for stored descriptor
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[propName]; ok {
					result := &Value{Type: "object", Obj: make(map[string]*Value)}
					if desc.Get != nil {
						result.Obj["get"] = desc.Get
						result.Obj["set"] = desc.Set
						if desc.Set == nil {
							result.Obj["set"] = &Value{Type: "undefined"}
						}
						result.Obj["enumerable"] = &Value{Type: "bool", Bool: desc.Enumerable}
						result.Obj["configurable"] = &Value{Type: "bool", Bool: desc.Configurable}
					} else {
						result.Obj["value"] = desc.Value
						if desc.Value == nil {
							result.Obj["value"] = &Value{Type: "undefined"}
						}
						result.Obj["writable"] = &Value{Type: "bool", Bool: desc.Writable}
						result.Obj["enumerable"] = &Value{Type: "bool", Bool: desc.Enumerable}
						result.Obj["configurable"] = &Value{Type: "bool", Bool: desc.Configurable}
					}
					return result
				}
			}

			// Check for regular property
			if val, ok := obj.Obj[propName]; ok {
				return &Value{Type: "object", Obj: map[string]*Value{
					"value":        val,
					"writable":     {Type: "bool", Bool: true},
					"enumerable":   {Type: "bool", Bool: true},
					"configurable": {Type: "bool", Bool: true},
				}}
			}

			return &Value{Type: "undefined"}
		}},
		// Object.assign(target, ...sources)
		"assign": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "undefined"}
			}
			target := args[0]
			if target.Type != "object" {
				return target
			}
			if target.Obj == nil {
				target.Obj = make(map[string]*Value)
			}
			for _, source := range args[1:] {
				if source.Type == "object" && source.Obj != nil {
					for k, v := range source.Obj {
						target.Obj[k] = v
					}
				}
			}
			return target
		}},
		// Object.create(proto)
		"create": {Type: "native", Native: func(args []*Value) *Value {
			obj := &Value{Type: "object", Obj: make(map[string]*Value)}
			// Store prototype reference (simplified)
			if len(args) > 0 && args[0].Type == "object" && args[0].Obj != nil {
				// Copy prototype properties as defaults
				for k, v := range args[0].Obj {
					obj.Obj[k] = v
				}
			}
			return obj
		}},
		// Object.getPrototypeOf(obj)
		"getPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 || args[0].Type != "object" {
				return &Value{Type: "null"}
			}
			// Simplified: return Object.prototype if it exists
			if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
				return objProto
			}
			return &Value{Type: "null"}
		}},
		// Object.setPrototypeOf(obj, proto)
		"setPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 2 {
				return args[0]
			}
			// Stub: just return the object
			return args[0]
		}},
		// Object.freeze(obj)
		"freeze": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "undefined"}
			}
			obj := args[0]
			if obj.Type == "object" {
				obj.Frozen = true
			}
			return obj
		}},
		// Object.is(value1, value2)
		"is": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 2 {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: valuesStrictEqual(args[0], args[1])}
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
