package jsengine

import (
	"math"
	"regexp"
	"sort"
	"strconv"
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

// getNativeThis extracts the 'this' object from native method call arguments.
// Convention: when evalCall invokes a native method with a ThisBinding,
// it prepends the 'this' object as args[0] with _isThisArg=true.
func getNativeThis(args []*Value) *Value {
	if len(args) > 0 && args[0]._isThisArg {
		return args[0]
	}
	return nil
}

// nativeThisOffset returns 1 if args[0] is a prepended 'this' argument,
// 0 otherwise. Used by Object static methods to skip the 'this' arg.
func NativeThisOffset(args []*Value) int {
	if len(args) > 0 && args[0]._isThisArg {
		return 1
	}
	return 0
}

// nativeThisOffset is the unexported alias for internal use.
func nativeThisOffset(args []*Value) int {
	return NativeThisOffset(args)
}

// Sets up Object.prototype, Array.prototype, Function.prototype, String.prototype,
// Number.prototype, Boolean.prototype, and Error.prototype with their standard methods.
func (vm *VM) setupPrototypes() {
	// Object.prototype
	objectProto := &Value{Type: "object", Obj: map[string]*Value{
		"hasOwnProperty": {Type: "native", Native: func(args []*Value) *Value {
			// Convention: args[0] is 'this' (passed by evalCall), args[1] is the property name
			thisObj := getNativeThis(args)
			propIdx := 0
			if thisObj != nil {
				propIdx = 1
			}
			if len(args) <= propIdx {
				return &Value{Type: "bool", Bool: false}
			}
			propName := valueToString(args[propIdx])
			if thisObj == nil || thisObj.Type != "object" {
				return &Value{Type: "bool", Bool: false}
			}
			// Check own Obj properties
			if thisObj.Obj != nil {
				if _, ok := thisObj.Obj[propName]; ok {
					return &Value{Type: "bool", Bool: true}
				}
			}
			// Check own Descriptors (Object.defineProperty sets these)
			if thisObj.Descriptors != nil {
				if _, ok := thisObj.Descriptors[propName]; ok {
					return &Value{Type: "bool", Bool: true}
				}
			}
			// Check Arr length for arrays
			if thisObj.Arr != nil && propName == "length" {
				return &Value{Type: "bool", Bool: true}
			}
			return &Value{Type: "bool", Bool: false}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil {
				switch thisObj.Type {
				case "undefined":
					return &Value{Type: "string", Str: "[object Undefined]"}
				case "null":
					return &Value{Type: "string", Str: "[object Null]"}
				case "string":
					return &Value{Type: "string", Str: "[object String]"}
				case "number":
					return &Value{Type: "string", Str: "[object Number]"}
				case "bool":
					return &Value{Type: "string", Str: "[object Boolean]"}
				case "function", "native":
					return &Value{Type: "string", Str: "[object Function]"}
				case "object":
					if thisObj.Arr != nil {
						return &Value{Type: "string", Str: "[object Array]"}
					}
					if thisObj.Obj != nil {
						return &Value{Type: "string", Str: "[object Object]"}
					}
					if thisObj.MapData != nil {
						return &Value{Type: "string", Str: "[object Map]"}
					}
					if thisObj.SetData != nil {
						return &Value{Type: "string", Str: "[object Set]"}
					}
					return &Value{Type: "string", Str: "[object Object]"}
				}
			}
			return &Value{Type: "string", Str: "[object Undefined]"}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil {
				return thisObj
			}
			return &Value{Type: "undefined"}
		}},
		"isPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "bool", Bool: false}
		}},
		"propertyIsEnumerable": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			propIdx := 0
			if thisObj != nil {
				propIdx = 1
			}
			if len(args) <= propIdx {
				return &Value{Type: "bool", Bool: false}
			}
			propName := valueToString(args[propIdx])
			if thisObj == nil || thisObj.Type != "object" {
				return &Value{Type: "bool", Bool: false}
			}
			if thisObj.Descriptors != nil {
				if desc, ok := thisObj.Descriptors[propName]; ok {
					return &Value{Type: "bool", Bool: desc.Enumerable}
				}
			}
			if thisObj.Obj != nil {
				if _, ok := thisObj.Obj[propName]; ok {
					return &Value{Type: "bool", Bool: true}
				}
			}
			return &Value{Type: "bool", Bool: false}
		}},
		"toLocaleString": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: "[object Object]"}
		}},
	}}
	vm.env.Define("ObjectPrototype", objectProto)

	// Array.prototype - delegate to callArrayMethod which has real implementations
	arrayProto := &Value{Type: "object", Obj: map[string]*Value{
		"push": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("push", thisObj, args[1:], vm)
			}
			return &Value{Type: "number", Num: 0}
		}},
		"pop": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("pop", thisObj, args[1:], vm)
			}
			return &Value{Type: "undefined"}
		}},
		"shift": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("shift", thisObj, args[1:], vm)
			}
			return &Value{Type: "undefined"}
		}},
		"unshift": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("unshift", thisObj, args[1:], vm)
			}
			return &Value{Type: "number", Num: 0}
		}},
		"slice": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("slice", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"splice": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("splice", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"indexOf": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("indexOf", thisObj, args[1:], vm)
			}
			return &Value{Type: "number", Num: -1}
		}},
		"includes": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("includes", thisObj, args[1:], vm)
			}
			return &Value{Type: "bool", Bool: false}
		}},
		"find": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("find", thisObj, args[1:], vm)
			}
			return &Value{Type: "undefined"}
		}},
		"findIndex": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("findIndex", thisObj, args[1:], vm)
			}
			return &Value{Type: "number", Num: -1}
		}},
		"forEach": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("forEach", thisObj, args[1:], vm)
			}
			return &Value{Type: "undefined"}
		}},
		"map": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("map", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"filter": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("filter", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"reduce": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				// Simple reduce implementation
				arr := thisObj.Arr
				if len(arr) == 0 {
					return &Value{Type: "undefined"}
				}
				if len(args) < 2 || args[1].Type != "function" {
					return arr[0]
				}
				fn := args[1].Func
				acc := arr[0]
				start := 1
				if len(args) >= 3 {
					acc = args[2]
					start = 0
				}
				for i := start; i < len(arr); i++ {
					acc = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{acc, arr[i], &Value{Type: "number", Num: float64(i)}, thisObj})
				}
				return acc
			}
			return &Value{Type: "undefined"}
		}},
		"reduceRight": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				arr := thisObj.Arr
				if len(arr) == 0 {
					return &Value{Type: "undefined"}
				}
				if len(args) < 2 || args[1].Type != "function" {
					return arr[len(arr)-1]
				}
				fn := args[1].Func
				acc := arr[len(arr)-1]
				start := len(arr) - 2
				if len(args) >= 3 {
					acc = args[2]
					start = len(arr) - 1
				}
				for i := start; i >= 0; i-- {
					acc = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{acc, arr[i], &Value{Type: "number", Num: float64(i)}, thisObj})
				}
				return acc
			}
			return &Value{Type: "undefined"}
		}},
		"every": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("every", thisObj, args[1:], vm)
			}
			return &Value{Type: "bool", Bool: true}
		}},
		"some": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("some", thisObj, args[1:], vm)
			}
			return &Value{Type: "bool", Bool: false}
		}},
		"join": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("join", thisObj, args[1:], vm)
			}
			return &Value{Type: "string", Str: ""}
		}},
		"concat": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("concat", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"reverse": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("reverse", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"sort": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("sort", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"flat": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				depth := 1
				if len(args) > 1 && args[1].Type == "number" {
					depth = int(args[1].Num)
				}
				var flatten func([]*Value, int) []*Value
				flatten = func(arr []*Value, d int) []*Value {
					result := make([]*Value, 0)
					for _, v := range arr {
						if v.Type == "object" && v.Arr != nil && d > 0 {
							result = append(result, flatten(v.Arr, d-1)...)
						} else {
							result = append(result, v)
						}
					}
					return result
				}
				return &Value{Type: "object", Arr: flatten(thisObj.Arr, depth)}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"flatMap": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				mapped := callArrayMethod("map", thisObj, args[1:], vm)
				if mapped.Type == "object" && mapped.Arr != nil {
					return &Value{Type: "object", Arr: mapped.Arr}
				}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"fill": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				arr := thisObj.Arr
				if len(args) < 2 {
					return thisObj
				}
				val := args[1]
				start, end := 0, len(arr)
				if len(args) >= 3 && args[2].Type == "number" {
					start = int(args[2].Num)
					if start < 0 {
						start = len(arr) + start
					}
					if start < 0 {
						start = 0
					}
				}
				if len(args) >= 4 && args[3].Type == "number" {
					end = int(args[3].Num)
					if end < 0 {
						end = len(arr) + end
					}
					if end > len(arr) {
						end = len(arr)
					}
				}
				if start > len(arr) {
					start = len(arr)
				}
				if end > len(arr) {
					end = len(arr)
				}
				for i := start; i < end; i++ {
					arr[i] = val
				}
				return thisObj
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"copyWithin": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				arr := thisObj.Arr
				if len(args) < 2 {
					return thisObj
				}
				target := 0
				if args[1].Type == "number" {
					target = int(args[1].Num)
					if target < 0 {
						target = len(arr) + target
					}
					if target < 0 {
						target = 0
					}
				}
				start := 0
				if len(args) >= 3 && args[2].Type == "number" {
					start = int(args[2].Num)
					if start < 0 {
						start = len(arr) + start
					}
					if start < 0 {
						start = 0
					}
				}
				end := len(arr)
				if len(args) >= 4 && args[3].Type == "number" {
					end = int(args[3].Num)
					if end < 0 {
						end = len(arr) + end
					}
					if end > len(arr) {
						end = len(arr)
					}
				}
				if target >= len(arr) || start >= end {
					return thisObj
				}
				if end > len(arr) {
					end = len(arr)
				}
				// Copy to temporary slice to handle overlapping
				toCopy := make([]*Value, end-start)
				copy(toCopy, arr[start:end])
				for i, v := range toCopy {
					idx := target + i
					if idx < len(arr) {
						arr[idx] = v
					}
				}
				return thisObj
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"keys": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				arr := thisObj.Arr
				keys := make([]*Value, len(arr))
				for i := range arr {
					keys[i] = &Value{Type: "number", Num: float64(i)}
				}
				return &Value{Type: "object", Arr: keys}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"values": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return &Value{Type: "object", Arr: append([]*Value{}, thisObj.Arr...)}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"entries": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				arr := thisObj.Arr
				entries := make([]*Value, len(arr))
				for i, v := range arr {
					entries[i] = &Value{Type: "object", Arr: []*Value{
						{Type: "number", Num: float64(i)},
						v,
					}}
				}
				return &Value{Type: "object", Arr: entries}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				parts := make([]string, len(thisObj.Arr))
				for i, v := range thisObj.Arr {
					parts[i] = valueToString(v)
				}
				return &Value{Type: "string", Str: strings.Join(parts, ",")}
			}
			return &Value{Type: "string", Str: ""}
		}},
		"length": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return &Value{Type: "number", Num: float64(len(thisObj.Arr))}
			}
			return &Value{Type: "number", Num: 0}
		}},
	}}
	vm.env.Define("ArrayPrototype", arrayProto)

	// Function.prototype
	functionProto := &Value{Type: "object", Obj: map[string]*Value{
		"call": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj == nil {
				return &Value{Type: "undefined"}
			}
			var callArgs []*Value
			var thisArg *Value
			if len(args) > 1 {
				thisArg = args[1]
				if len(args) > 2 {
					callArgs = args[2:]
				}
			}
			if thisArg == nil {
				thisArg = &Value{Type: "undefined"}
			}
			thisArg._isThisArg = true
			callArgs = append([]*Value{thisArg}, callArgs...)
			if thisObj.Type == "function" && thisObj.Func != nil {
				return vm.callFunction(thisObj, callArgs)
			} else if thisObj.Type == "native" && thisObj.Native != nil {
				return thisObj.Native(callArgs)
			}
			return &Value{Type: "undefined"}
		}},
		"apply": {Type: "native", Native: func(args []*Value) *Value {
			// Function.prototype.apply(thisArg, argsArray)
			thisObj := getNativeThis(args)
			if thisObj == nil {
				return &Value{Type: "undefined"}
			}
			var callArgs []*Value
			var thisArg *Value
			if len(args) > 1 {
				thisArg = args[1]
			}
			if len(args) > 2 && args[2].Type == "object" && args[2].Arr != nil {
				callArgs = args[2].Arr
			}
			if thisArg == nil {
				thisArg = &Value{Type: "undefined"}
			}
			thisArg._isThisArg = true
			callArgs = append([]*Value{thisArg}, callArgs...)
			if thisObj.Type == "function" && thisObj.Func != nil {
				return vm.callFunction(thisObj, callArgs)
			} else if thisObj.Type == "native" && thisObj.Native != nil {
				return thisObj.Native(callArgs)
			}
			return &Value{Type: "undefined"}
		}},
		"bind": {Type: "native", Native: func(args []*Value) *Value {
			// Function.prototype.bind(thisArg, ...args)
			thisObj := getNativeThis(args)
			if thisObj == nil {
				return &Value{Type: "undefined"}
			}
			var boundThis *Value
			var boundArgs []*Value
			if len(args) > 1 {
				boundThis = args[1]
				if len(args) > 2 {
					boundArgs = args[2:]
				}
			}
			if boundThis == nil {
				boundThis = &Value{Type: "undefined"}
			}
			// Return a new native function that calls the original with bound this and args
			return &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				// Combine bound args with new args
				combinedArgs := append([]*Value{}, boundArgs...)
				combinedArgs = append(combinedArgs, innerArgs...)
				// Prepend bound this
				boundThisCopy := &Value{}
				*boundThisCopy = *boundThis
				boundThisCopy._isThisArg = true
				combinedArgs = append([]*Value{boundThisCopy}, combinedArgs...)
				if thisObj.Type == "function" && thisObj.Func != nil {
					return vm.callFunction(thisObj, combinedArgs)
				} else if (thisObj.Type == "native" || thisObj.Type == "arrayMethod") && thisObj.Native != nil {
					return thisObj.Native(combinedArgs)
				}
				return &Value{Type: "undefined"}
			}}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := args[0]
			if thisObj != nil && thisObj.Type == "function" && thisObj.Func != nil {
				params := strings.Join(thisObj.Func.Params, ", ")
				return &Value{Type: "string", Str: "function (" + params + ") { [code] }"}
			}
			return &Value{Type: "string", Str: "function () { [native code] }"}
		}},
	}}
	vm.env.Define("FunctionPrototype", functionProto)

	// String.prototype - methods delegate to callStringMethod via thisBinding
	stringProto := &Value{Type: "object", Obj: map[string]*Value{
		"constructor": {Type: "native", BuiltInConstructor: "String", Native: func(args []*Value) *Value {
			s := ""
			if len(args) > 0 {
				s = valueToString(args[0])
			}
			return &Value{Type: "string", Str: s}
		}},
		"toString": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			return &Value{Type: "string", Str: s}
		}},
		"valueOf": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			return &Value{Type: "string", Str: s}
		}},
		"charAt": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("charAt", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"charCodeAt": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("charCodeAt", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"concat": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("concat", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"indexOf": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("indexOf", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"lastIndexOf": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("lastIndexOf", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"includes": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("includes", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"startsWith": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("startsWith", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"endsWith": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("endsWith", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"slice": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("slice", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"substring": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("substring", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"substr": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("substr", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"toLowerCase": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("toLowerCase", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"toUpperCase": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("toUpperCase", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"trim": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("trim", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"split": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("split", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"replace": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("replace", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"match": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("match", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"search": {Type: "native", Native: func(args []*Value) *Value {
			return callStringMethod("search", stringThisValue(args), nativeMethodArgs(args), vm)
		}},
		"trimStart": {Type: "native", Native: func(args []*Value) *Value {
			s := strings.TrimSpace(stringThisValue(args))
			return &Value{Type: "string", Str: s}
		}},
		"trimEnd": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			return &Value{Type: "string", Str: strings.TrimRight(s, " \t\n\r")}
		}},
		"repeat": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			count := 0
			if len(args) > 1 && args[1].Type == "number" {
				count = int(args[1].Num)
			}
			return &Value{Type: "string", Str: strings.Repeat(s, count)}
		}},
		"padStart": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			targetLen := 0
			if len(args) > 1 && args[1].Type == "number" {
				targetLen = int(args[1].Num)
			}
			padStr := " "
			if len(args) > 2 {
				padStr = valueToString(args[2])
			}
			for len(s) < targetLen {
				s = padStr + s
			}
			if len(s) > targetLen {
				s = s[len(s)-targetLen:]
			}
			return &Value{Type: "string", Str: s}
		}},
		"padEnd": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			targetLen := 0
			if len(args) > 1 && args[1].Type == "number" {
				targetLen = int(args[1].Num)
			}
			padStr := " "
			if len(args) > 2 {
				padStr = valueToString(args[2])
			}
			for len(s) < targetLen {
				s = s + padStr
			}
			if len(s) > targetLen {
				s = s[:targetLen]
			}
			return &Value{Type: "string", Str: s}
		}},
		"localeCompare": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}},
		"normalize": {Type: "native", Native: func(args []*Value) *Value {
			s := stringThisValue(args)
			return &Value{Type: "string", Str: s}
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
			thisObj := getNativeThis(args)
			offset := 0
			if thisObj != nil {
				offset = 1
			}
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[offset]
			if obj.Type != "object" {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			seen := make(map[string]bool)
			keys := make([]*Value, 0)
			// Include own enumerable keys from Obj map (not prototype-inherited)
			if obj.Obj != nil {
				for k := range obj.Obj {
					if !seen[k] {
						// Skip non-enumerable descriptor properties
						if obj.Descriptors != nil {
							if desc, ok := obj.Descriptors[k]; ok && !desc.Enumerable {
								continue
							}
						}
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			// Include enumerable keys from descriptors (Object.defineProperty)
			if obj.Descriptors != nil {
				for k, desc := range obj.Descriptors {
					if !seen[k] && desc.Enumerable {
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			// Include array indices
			if obj.Arr != nil {
				for i := range obj.Arr {
					k := strconv.Itoa(i)
					if !seen[k] {
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			return &Value{Type: "object", Arr: keys}
		}},
		"values": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			offset := 0
			if thisObj != nil {
				offset = 1
			}
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[offset]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			vals := make([]*Value, 0, len(obj.Obj))
			seen := make(map[string]bool)
			for k, v := range obj.Obj {
				if !seen[k] {
					vals = append(vals, v)
					seen[k] = true
				}
			}
			if obj.Descriptors != nil {
				for k, desc := range obj.Descriptors {
					if !seen[k] && desc.Value != nil {
						vals = append(vals, desc.Value)
						seen[k] = true
					}
				}
			}
			return &Value{Type: "object", Arr: vals}
		}},
		"entries": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			offset := 0
			if thisObj != nil {
				offset = 1
			}
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[offset]
			if obj.Type != "object" || obj.Obj == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			seen := make(map[string]bool)
			entries := make([]*Value, 0, len(obj.Obj))
			for k, v := range obj.Obj {
				if !seen[k] {
					entry := &Value{Type: "object", Arr: []*Value{
						{Type: "string", Str: k},
						v,
					}}
					entries = append(entries, entry)
					seen[k] = true
				}
			}
			if obj.Descriptors != nil {
				for k, desc := range obj.Descriptors {
					if !seen[k] && desc.Value != nil {
						entry := &Value{Type: "object", Arr: []*Value{
							{Type: "string", Str: k},
							desc.Value,
						}}
						entries = append(entries, entry)
						seen[k] = true
					}
				}
			}
			return &Value{Type: "object", Arr: entries}
		}},
		// Object.defineProperty(obj, prop, descriptor)
		// Critical for Vue 2 reactivity system
		"defineProperty": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			offset := 0
			if thisObj != nil {
				offset = 1
			}
			if len(args) < offset+3 {
				if offset == 0 && len(args) >= 3 {
					offset = 0
				} else {
					return args[offset]
				}
			}
			obj := args[offset]
			propName := valueToString(args[offset+1])
			descObj := args[offset+2]

			if obj.Type != "object" && obj.Type != "function" && obj.Type != "native" {
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

			if desc.Get == nil && desc.Set == nil {
				if desc.Value != nil {
					obj.Obj[propName] = desc.Value
				}
			} else {
				delete(obj.Obj, propName)
			}

			return obj
		}},
		// Object.getOwnPropertyDescriptor(obj, prop)
		"getOwnPropertyDescriptor": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) < offset+2 {
				return &Value{Type: "undefined"}
			}
			obj := args[offset]
			propName := valueToString(args[offset+1])

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
			offset := nativeThisOffset(args)
			if len(args) < offset+1 {
				return &Value{Type: "undefined"}
			}
			target := args[offset]
			if target.Type != "object" {
				return target
			}
			if target.Obj == nil {
				target.Obj = make(map[string]*Value)
			}
			for _, source := range args[offset+1:] {
				if source.Type == "object" && source.Obj != nil {
					for k, v := range source.Obj {
						target.Obj[k] = v
					}
					// Also copy descriptor properties
					if source.Descriptors != nil {
						if target.Descriptors == nil {
							target.Descriptors = make(map[string]*PropertyDescriptor)
						}
						for k, desc := range source.Descriptors {
							if desc.Value != nil {
								target.Obj[k] = desc.Value
							}
							descCopy := *desc
							target.Descriptors[k] = &descCopy
						}
					}
				}
			}
			return target
		}},
		// Object.create(proto)
		"create": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			obj := &Value{Type: "object", Obj: make(map[string]*Value)}
			// Set prototype reference via Proto field
			if len(args) > offset && args[offset].Type == "object" {
				obj.Proto = args[offset]
			}
			return obj
		}},
		// Object.getPrototypeOf(obj)
		"getPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || args[offset].Type != "object" {
				return &Value{Type: "null"}
			}
			obj := args[offset]
			if obj.Proto != nil {
				return obj.Proto
			}
			if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
				return objProto
			}
			return &Value{Type: "null"}
		}},
		// Object.setPrototypeOf(obj, proto)
		"setPrototypeOf": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return args[offset]
			}
			obj := args[offset]
			if obj.Type == "object" {
				obj.Proto = args[offset+1]
			}
			return obj
		}},
		// Object.freeze(obj)
		"freeze": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			obj := args[offset]
			if obj.Type == "object" {
				obj.Frozen = true
			}
			return obj
		}},
		// Object.is(value1, value2)
		"is": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: valuesStrictEqual(args[offset], args[offset+1])}
		}},
		// Object.getOwnPropertyNames(obj) — returns all own property names
		// (enumerable and non-enumerable), but not Symbol-keyed properties.
		"getOwnPropertyNames": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			obj := args[offset]
			if obj.Type != "object" && obj.Type != "function" {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			seen := make(map[string]bool)
			keys := make([]*Value, 0)
			// Obj map properties
			if obj.Obj != nil {
				for k := range obj.Obj {
					if !seen[k] {
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			// Descriptor properties (even non-enumerable ones)
			if obj.Descriptors != nil {
				for k := range obj.Descriptors {
					if !seen[k] {
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			// Array indices
			if obj.Arr != nil {
				for i := range obj.Arr {
					k := strconv.Itoa(i)
					if !seen[k] {
						keys = append(keys, &Value{Type: "string", Str: k})
						seen[k] = true
					}
				}
			}
			// For functions, include "prototype" if it has one
			if (obj.Type == "function" || obj.Type == "native") && !seen["prototype"] {
				keys = append(keys, &Value{Type: "string", Str: "prototype"})
				seen["prototype"] = true
			}
			return &Value{Type: "object", Arr: keys}
		}},
	}
}

// GetArrayPrototypeMethods returns the built-in Array.prototype methods that
// core-js polyfills may overwrite with broken JS implementations. These are
// restored after script execution to ensure correct behavior in our VM.
func GetArrayPrototypeMethods(vm *VM) map[string]*Value {
	return map[string]*Value{
		"concat": {Type: "native", Native: func(args []*Value) *Value {
			thisObj := getNativeThis(args)
			if thisObj != nil && thisObj.Type == "object" && thisObj.Arr != nil {
				return callArrayMethod("concat", thisObj, args[1:], vm)
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
	}
}

// stripLookahead removes JS lookahead (?=...) and (?!...) and lookbehind
// (?<=...) and (?<!...) patterns from a regex so it can compile with Go's
// RE2 engine. The content inside the lookahead is preserved as a
// non-capturing group to keep group numbering stable.
func stripLookahead(pattern string) string {
	result := make([]byte, 0, len(pattern))
	i := 0
	for i < len(pattern) {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			result = append(result, pattern[i], pattern[i+1])
			i += 2
			continue
		}
		if pattern[i] == '(' && i+2 < len(pattern) && pattern[i+1] == '?' {
			if pattern[i+2] == '=' || pattern[i+2] == '!' {
				// Lookahead: (?=...) or (?!...) -> replace with (?:...)
				result = append(result, "(?:"...)
				i += 3
				depth := 1
				for i < len(pattern) && depth > 0 {
					if pattern[i] == '\\' && i+1 < len(pattern) {
						result = append(result, pattern[i], pattern[i+1])
						i += 2
						continue
					}
					if pattern[i] == '(' {
						depth++
					} else if pattern[i] == ')' {
						depth--
						if depth == 0 {
							result = append(result, ')')
							i++
							break
						}
					}
					result = append(result, pattern[i])
					i++
				}
				continue
			}
			if i+3 < len(pattern) && pattern[i+2] == '<' && (pattern[i+3] == '=' || pattern[i+3] == '!') {
				// Lookbehind: (?<=...) or (?<!...) -> replace with (?:...)
				result = append(result, "(?:"...)
				i += 4
				depth := 1
				for i < len(pattern) && depth > 0 {
					if pattern[i] == '\\' && i+1 < len(pattern) {
						result = append(result, pattern[i], pattern[i+1])
						i += 2
						continue
					}
					if pattern[i] == '(' {
						depth++
					} else if pattern[i] == ')' {
						depth--
						if depth == 0 {
							result = append(result, ')')
							i++
							break
						}
					}
					result = append(result, pattern[i])
					i++
				}
				continue
			}
		}
		result = append(result, pattern[i])
		i++
	}
	return string(result)
}

// newRegExp creates a new RegExp object.
func newRegExp(pattern, flags string) *Value {
	goPattern := pattern
	// Convert JS regex flags to Go regex flags where possible
	goFlags := ""
	if strings.Contains(flags, "i") {
		goFlags += "(?i)"
	}
	if strings.Contains(flags, "m") {
		goFlags += "(?m)"
	}
	if strings.Contains(flags, "s") {
		goFlags += "(?s)"
	}

	// Go's regexp (RE2) does not support lookahead (?=...) and (?!...)
	// or lookbehind (?<=...) and (?<!...). Strip these patterns so
	// that common JS regex patterns (e.g., path-to-regexp output)
	// can compile. This is a heuristic; it changes semantics slightly
	// but is necessary for VueRouter compatibility.
	goPattern = stripLookahead(goPattern)

	compiled, err := regexp.Compile(goFlags + goPattern)
	if err != nil {
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

	regexpObj := &Value{Type: "regexp", Obj: make(map[string]*Value)}
	regexpObj.Obj["pattern"] = &Value{Type: "string", Str: pattern}
	regexpObj.Obj["flags"] = &Value{Type: "string", Str: flags}
	regexpObj.Obj["lastIndex"] = &Value{Type: "number", Num: 0}
	regexpObj.Obj["_goSource"] = &Value{Type: "string", Str: goPattern}
	regexpObj.Obj["test"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		testStr := ""
		if len(args) >= 1 {
			if args[0]._isThisArg {
				if len(args) >= 2 {
					testStr = valueToString(args[1])
				}
			} else {
				testStr = valueToString(args[0])
			}
		}
		return &Value{Type: "bool", Bool: compiled.MatchString(testStr)}
	}}
	regexpObj.Obj["exec"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		testStr := ""
		if len(args) >= 1 {
			if args[0]._isThisArg {
				if len(args) >= 2 {
					testStr = valueToString(args[1])
				}
			} else {
				testStr = valueToString(args[0])
			}
		}
		// Read lastIndex for global flag support
		lastIdx := 0
		if regexpObj.Obj["lastIndex"] != nil && regexpObj.Obj["lastIndex"].Type == "number" {
			lastIdx = int(regexpObj.Obj["lastIndex"].Num)
		}
		// If lastIndex is beyond the string, reset and return null
		if lastIdx > len(testStr) {
			regexpObj.Obj["lastIndex"] = &Value{Type: "number", Num: 0}
			return &Value{Type: "null"}
		}
		searchStr := testStr[lastIdx:]
		matches := compiled.FindStringSubmatch(searchStr)
		if matches == nil {
			// With g flag, reset lastIndex on failure
			if strings.Contains(flags, "g") {
				regexpObj.Obj["lastIndex"] = &Value{Type: "number", Num: 0}
			}
			return &Value{Type: "null"}
		}
		// Update lastIndex: position after the match in the original string
		matchPos := strings.Index(searchStr, matches[0])
		newLastIndex := lastIdx + matchPos + len(matches[0])
		regexpObj.Obj["lastIndex"] = &Value{Type: "number", Num: float64(newLastIndex)}
		// Build result array with all capture groups
		arr := make([]*Value, len(matches))
		for i, m := range matches {
			if m != "" {
				arr[i] = &Value{Type: "string", Str: m}
			} else {
				arr[i] = &Value{Type: "undefined"}
			}
		}
		obj := map[string]*Value{
			"index":  {Type: "number", Num: float64(lastIdx + matchPos)},
			"input":  {Type: "string", Str: testStr},
			"length": {Type: "number", Num: float64(len(matches))},
		}
		for i, m := range matches {
			if m != "" {
				obj[strconv.Itoa(i)] = &Value{Type: "string", Str: m}
			} else {
				obj[strconv.Itoa(i)] = &Value{Type: "undefined"}
			}
		}
		return &Value{Type: "object", Obj: obj, Arr: arr}
	}}
	regexpObj.Obj["source"] = &Value{Type: "string", Str: pattern}
	regexpObj.Obj["global"] = &Value{Type: "bool", Bool: strings.Contains(flags, "g")}
	regexpObj.Obj["ignoreCase"] = &Value{Type: "bool", Bool: strings.Contains(flags, "i")}
	regexpObj.Obj["multiline"] = &Value{Type: "bool", Bool: strings.Contains(flags, "m")}
	return regexpObj
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
// It checks whether obj appears in the prototype chain of constructor.prototype.
func (vm *VM) instanceof(obj, constructor *Value) bool {
	// constructor must be a function (native or user-defined) or a constructor object
	if constructor.Type != "function" && constructor.Type != "native" && constructor.Type != "object" {
		return false
	}

	// Determine the built-in constructor name from the constructor value
	constructorBIC := constructor.BuiltInConstructor
	if constructorBIC == "" && constructor.Obj != nil {
		if bic, ok := constructor.Obj["_builtInConstructor"]; ok && bic.Type == "string" {
			constructorBIC = bic.Str
		}
	}

	// Functions are always instances of Function
	if (obj.Type == "function" || obj.Type == "native") && constructorBIC == "Function" {
		return true
	}

	// Primitives (non-object, non-function) are never instances
	// Exception: regexp type values are objects and can be instanceof RegExp
	if obj.Type != "object" && obj.Type != "regexp" {
		return false
	}

	if constructorBIC != "" {
		// Direct match on BuiltInConstructor
		if obj.BuiltInConstructor == constructorBIC {
			return true
		}
		// RegExp type is always instanceof RegExp
		if obj.Type == "regexp" && constructorBIC == "RegExp" {
			return true
		}
		// Also check Proto chain for BuiltInConstructor
		proto := obj.Proto
		for proto != nil {
			if proto.BuiltInConstructor == constructorBIC {
				return true
			}
			proto = proto.Proto
		}
		// Structural checks when obj.BuiltInConstructor is not set
		// Arrays are instances of Array
		if constructorBIC == "Array" && obj.Arr != nil {
			return true
		}
		// Everything that is an object is an instance of Object
		if constructorBIC == "Object" {
			return true
		}
		// Functions are instances of Function
		if constructorBIC == "Function" && (obj.Type == "function" || obj.Type == "native") {
			return true
		}
		return false
	}

	// User-defined function: check prototype chain
	// constructor.PrototypeObj is the function's .prototype property.
	// obj must have constructor.PrototypeObj in its Proto chain.
	protoObj := constructor.PrototypeObj
	if protoObj == nil {
		// Try to get .prototype from constructor's Obj map
		if constructor.Obj != nil {
			if p, ok := constructor.Obj["prototype"]; ok {
				protoObj = p
			}
		}
	}
	if protoObj == nil {
		return false
	}

	// Walk the prototype chain of obj
	proto := obj.Proto
	depth := 0
	for proto != nil {
		if proto == protoObj {
			return true
		}
		proto = proto.Proto
		depth++
		if depth > 100 {
			break
		}
	}

	return false
}

// stringThisValue extracts the string value from the 'this' binding in args.
// When a String.prototype method is called, args[0] is the this value.
func stringThisValue(args []*Value) string {
	if len(args) > 0 && args[0] != nil {
		this := args[0]
		if this.Type == "string" {
			return this.Str
		}
		if this.Type == "object" && this.Obj != nil {
			if s, ok := this.Obj["__string_value__"]; ok && s != nil && s.Type == "string" {
				return s.Str
			}
		}
	}
	return ""
}

// nativeMethodArgs returns the method arguments, stripping the 'this' binding.
// When native methods are called with a this-binding, args[0] is 'this'.
func nativeMethodArgs(args []*Value) []*Value {
	offset := nativeThisOffset(args)
	if offset >= len(args) {
		return nil
	}
	return args[offset:]
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
		if len(args) >= 1 && args[0].Type == "regexp" && args[0].Obj != nil {
			if source, ok := args[0].Obj["source"]; ok {
				re, err := regexp.Compile(source.Str)
				if err == nil {
					var result []*Value
					lastIdx := 0
					matches := re.FindAllStringSubmatchIndex(str, -1)
					for _, match := range matches {
						if match[0] > lastIdx {
							result = append(result, &Value{Type: "string", Str: str[lastIdx:match[0]]})
						} else if match[0] == lastIdx && lastIdx == 0 {
							// First match at start
						} else if match[0] == lastIdx {
							result = append(result, &Value{Type: "string", Str: ""})
						}
						// Add capturing groups
						for i := 2; i < len(match); i += 2 {
							if match[i] >= 0 {
								result = append(result, &Value{Type: "string", Str: str[match[i]:match[i+1]]})
							}
						}
						// Add the full match if no capturing groups
						if len(match) <= 2 {
							result = append(result, &Value{Type: "string", Str: str[match[0]:match[1]]})
						}
						lastIdx = match[1]
					}
					if lastIdx <= len(str) {
						result = append(result, &Value{Type: "string", Str: str[lastIdx:]})
					}
					return &Value{Type: "object", Arr: result}
				}
			}
		}
		sep := ""
		if len(args) >= 1 {
			sep = valueToString(args[0])
		}
		if sep == "" {
			arr := make([]*Value, len(str))
			for i, ch := range str {
				arr[i] = &Value{Type: "string", Str: string(ch)}
			}
			return &Value{Type: "object", Arr: arr}
		}
		parts := strings.Split(str, sep)
		arr := make([]*Value, len(parts))
		for i, p := range parts {
			arr[i] = &Value{Type: "string", Str: p}
		}
		return &Value{Type: "object", Arr: arr}
	case "replace":
		if len(args) < 2 {
			return &Value{Type: "string", Str: str}
		}
		if args[0].Type == "regexp" && args[0].Obj != nil {
			sourceKey := "_goSource"
			if _, ok := args[0].Obj[sourceKey]; !ok {
				sourceKey = "source"
			}
			if source, ok := args[0].Obj[sourceKey]; ok {
				goFlags := ""
				if flags, ok := args[0].Obj["flags"]; ok && strings.Contains(flags.Str, "i") {
					goFlags = "(?i)"
				}
				if flags, ok := args[0].Obj["flags"]; ok && strings.Contains(flags.Str, "m") {
					goFlags += "(?m)"
				}
				re, err := regexp.Compile(goFlags + source.Str)
				if err == nil {
					if args[1].Type == "function" || args[1].Type == "native" {
						global := false
						if gFlag, ok := args[0].Obj["global"]; ok && gFlag.Type == "bool" && gFlag.Bool {
							global = true
						}
						result := strReplaceWithCallback(re, str, args[1], vm, global)
						return &Value{Type: "string", Str: result}
				}
					replacement := valueToString(args[1])
					if global, ok := args[0].Obj["global"]; ok && global.Type == "bool" && global.Bool {
						return &Value{Type: "string", Str: jsRegexReplaceAll(re, str, replacement)}
					}
					loc := re.FindStringSubmatchIndex(str)
					if loc != nil && len(loc) >= 2 {
						result := str[:loc[0]] + jsExpandReplacement(replacement, re, str, loc) + str[loc[1]:]
						return &Value{Type: "string", Str: result}
					}
					return &Value{Type: "string", Str: str}
				}
			}
		}
		if args[1].Type == "function" || args[1].Type == "native" {
			searchStr := valueToString(args[0])
			idx := strings.Index(str, searchStr)
			if idx >= 0 {
				match := str[idx : idx+len(searchStr)]
				cbArgs := []*Value{{Type: "string", Str: match}}
				retVal := vm.callFunction(args[1], cbArgs)
				return &Value{Type: "string", Str: str[:idx] + valueToString(retVal) + str[idx+len(searchStr):]}
			}
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
	case "charCodeAt":
		idx := 0
		if len(args) >= 1 && args[0].Type == "number" {
			idx = int(args[0].Num)
		}
		runes := []rune(str)
		if idx < 0 || idx >= len(runes) {
			return &Value{Type: "number", Num: math.NaN()}
		}
		return &Value{Type: "number", Num: float64(runes[idx])}
	case "lastIndexOf":
		if len(args) < 1 {
			return &Value{Type: "number", Num: -1}
		}
		searchStr := valueToString(args[0])
		idx := strings.LastIndex(str, searchStr)
		if len(args) >= 2 && args[1].Type == "number" {
			pos := int(args[1].Num)
			if pos >= 0 && pos < idx {
				lastPart := str[:pos+1]
				idx = strings.LastIndex(lastPart, searchStr)
			}
		}
		return &Value{Type: "number", Num: float64(idx)}
	case "match":
		if len(args) < 1 {
			return &Value{Type: "null"}
		}
		var re *regexp.Regexp
		regexArg := args[0]
		if regexArg.Type == "regexp" && regexArg.Obj != nil {
			// Use _goSource (lookahead-stripped pattern) if available,
			// otherwise fall back to source.
			sourceKey := "_goSource"
			if _, ok := regexArg.Obj[sourceKey]; !ok {
				sourceKey = "source"
			}
			if source, ok := regexArg.Obj[sourceKey]; ok {
				goFlags := ""
				if flags, ok := regexArg.Obj["flags"]; ok && strings.Contains(flags.Str, "i") {
					goFlags = "(?i)"
				}
				if flags, ok := regexArg.Obj["flags"]; ok && strings.Contains(flags.Str, "m") {
					goFlags += "(?m)"
				}
				re, _ = regexp.Compile(goFlags + source.Str)
			}
		} else {
			pattern := valueToString(regexArg)
			re, _ = regexp.Compile(pattern)
		}
		if re == nil {
			return &Value{Type: "null"}
		}
		matches := re.FindStringSubmatch(str)
		if matches == nil {
			return &Value{Type: "null"}
		}
		arr := make([]*Value, len(matches))
		for i, m := range matches {
			arr[i] = &Value{Type: "string", Str: m}
		}
		result := &Value{Type: "object", Arr: arr, Obj: map[string]*Value{
			"index":  {Type: "number", Num: 0},
			"input":  {Type: "string", Str: str},
			"length": {Type: "number", Num: float64(len(matches))},
		}}
		for i, m := range matches {
			result.Obj[strconv.Itoa(i)] = &Value{Type: "string", Str: m}
		}
		return result
	case "search":
		if len(args) < 1 {
			return &Value{Type: "number", Num: -1}
		}
		var re *regexp.Regexp
		if args[0].Type == "regexp" && args[0].Obj != nil {
			sourceKey := "_goSource"
			if _, ok := args[0].Obj[sourceKey]; !ok {
				sourceKey = "source"
			}
			if source, ok := args[0].Obj[sourceKey]; ok {
				re, _ = regexp.Compile(source.Str)
			}
		} else {
			pattern := valueToString(args[0])
			re, _ = regexp.Compile(pattern)
		}
		if re == nil {
			return &Value{Type: "number", Num: -1}
		}
		loc := re.FindStringIndex(str)
		if loc == nil {
			return &Value{Type: "number", Num: -1}
		}
		return &Value{Type: "number", Num: float64(loc[0])}
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
		return &Value{Type: "object", Arr: append([]*Value{}, arr[start:end]...)}
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
		return &Value{Type: "object", Arr: deleted}
	case "concat":
		result := append([]*Value{}, arr...)
		for _, arg := range args {
			if arg.Type == "object" && arg.Arr != nil {
				result = append(result, arg.Arr...)
			} else {
				result = append(result, arg)
			}
		}
		return &Value{Type: "object", Arr: result}
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
			return &Value{Type: "object", Arr: []*Value{}}
		}
		fn := args[0].Func
		result := make([]*Value, len(arr))
		for i, v := range arr {
			result[i] = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
		}
		return &Value{Type: "object", Arr: result}
	case "filter":
		if len(args) < 1 || args[0].Type != "function" {
			return &Value{Type: "object", Arr: append([]*Value{}, arr...)}
		}
		fn := args[0].Func
		result := make([]*Value, 0)
		for i, v := range arr {
			ret := vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{v, &Value{Type: "number", Num: float64(i)}, obj})
			if isTruthy(ret) {
				result = append(result, v)
			}
		}
		return &Value{Type: "object", Arr: result}
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
	case "reduce":
		if len(args) < 1 || args[0].Type != "function" {
			if len(arr) == 0 {
				return &Value{Type: "undefined"}
			}
			return arr[0]
		}
		fn := args[0].Func
		if len(arr) == 0 {
			return &Value{Type: "undefined"}
		}
		acc := arr[0]
		start := 1
		if len(args) >= 2 {
			acc = args[1]
			start = 0
		}
		for i := start; i < len(arr); i++ {
			acc = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{acc, arr[i], &Value{Type: "number", Num: float64(i)}, obj})
		}
		return acc
	case "reduceRight":
		if len(args) < 1 || args[0].Type != "function" {
			if len(arr) == 0 {
				return &Value{Type: "undefined"}
			}
			return arr[len(arr)-1]
		}
		fn := args[0].Func
		if len(arr) == 0 {
			return &Value{Type: "undefined"}
		}
		acc := arr[len(arr)-1]
		start := len(arr) - 2
		if len(args) >= 2 {
			acc = args[1]
			start = len(arr) - 1
		}
		for i := start; i >= 0; i-- {
			acc = vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{acc, arr[i], &Value{Type: "number", Num: float64(i)}, obj})
		}
		return acc
	case "length":
		return &Value{Type: "number", Num: float64(len(arr))}
	default:
		return &Value{Type: "undefined"}
	}
}

// strReplaceWithCallback performs str.replace(regex, callback) with proper
// capture group arguments passed to the callback, matching JS spec behavior.
func strReplaceWithCallback(re *regexp.Regexp, str string, callback *Value, vm *VM, global bool) string {
	if !global {
		loc := re.FindStringSubmatchIndex(str)
		if loc == nil {
			return str
		}
		cbArgs := buildReplaceCallbackArgs(str, loc)
		retVal := vm.callFunction(callback, cbArgs)
		replacement := valueToString(retVal)
		return str[:loc[0]] + replacement + str[loc[1]:]
	}

	var result strings.Builder
	lastEnd := 0
	matches := re.FindAllStringSubmatchIndex(str, -1)
	for _, loc := range matches {
		result.WriteString(str[lastEnd:loc[0]])
		cbArgs := buildReplaceCallbackArgs(str, loc)
		retVal := vm.callFunction(callback, cbArgs)
		result.WriteString(valueToString(retVal))
		lastEnd = loc[1]
	}
	result.WriteString(str[lastEnd:])
	return result.String()
}

// buildReplaceCallbackArgs creates the callback arguments for String.prototype.replace
// per the ECMAScript spec: (match, p1, p2, ..., offset, string)
func buildReplaceCallbackArgs(str string, loc []int) []*Value {
	match := str[loc[0]:loc[1]]
	cbArgs := []*Value{{Type: "string", Str: match}}

	// Capture groups: loc[2], loc[3] is group 1, loc[4], loc[5] is group 2, etc.
	for i := 2; i+1 < len(loc); i += 2 {
		if loc[i] == -1 {
			cbArgs = append(cbArgs, &Value{Type: "undefined"})
		} else {
			cbArgs = append(cbArgs, &Value{Type: "string", Str: str[loc[i]:loc[i+1]]})
		}
	}

	// Offset argument
	cbArgs = append(cbArgs, &Value{Type: "number", Num: float64(loc[0])})

	// Original string argument
	cbArgs = append(cbArgs, &Value{Type: "string", Str: str})

	return cbArgs
}

// jsExpandReplacement expands JavaScript-style replacement patterns in a
// replacement string using the match results from a regex substitution.
// JS replacement patterns: $1, $2, ..., $& (match), $$ (literal $).
// Unlike Go's ReplaceAllString, backslashes have no special meaning in JS
// replacement strings — only $ patterns are special.
func jsExpandReplacement(replacement string, re *regexp.Regexp, str string, matchIndexes []int) string {
	var result strings.Builder
	i := 0
	for i < len(replacement) {
		if replacement[i] == '$' && i+1 < len(replacement) {
			next := replacement[i+1]
			if next == '$' {
				result.WriteByte('$')
				i += 2
				continue
			}
			if next == '&' {
				// $& = matched substring
				if len(matchIndexes) >= 2 {
					result.WriteString(str[matchIndexes[0]:matchIndexes[1]])
				}
				i += 2
				continue
			}
			if next == '`' {
				// $` = portion before match
				if len(matchIndexes) >= 2 {
					result.WriteString(str[:matchIndexes[0]])
				}
				i += 2
				continue
			}
			if next == '\'' {
				// $' = portion after match
				if len(matchIndexes) >= 2 {
					result.WriteString(str[matchIndexes[1]:])
				}
				i += 2
				continue
			}
			if next >= '0' && next <= '9' {
				// $n = capture group reference
				groupNum := int(next - '0')
				// Check for two-digit group numbers
				if i+2 < len(replacement) && replacement[i+2] >= '0' && replacement[i+2] <= '9' {
					twoDigit := groupNum*10 + int(replacement[i+2]-'0')
					if twoDigit*2+2 < len(matchIndexes) && matchIndexes[twoDigit*2] >= 0 {
						groupNum = twoDigit
						i += 3
					} else {
						i += 2
					}
				} else {
					i += 2
				}
				idx := groupNum * 2
				if idx+1 < len(matchIndexes) && matchIndexes[idx] >= 0 {
					result.WriteString(str[matchIndexes[idx]:matchIndexes[idx+1]])
				}
				continue
			}
		}
		result.WriteByte(replacement[i])
		i++
	}
	return result.String()
}

// jsRegexReplaceAll performs a global regex replace using JS replacement patterns.
func jsRegexReplaceAll(re *regexp.Regexp, str string, replacement string) string {
	var result strings.Builder
	lastEnd := 0
	matches := re.FindAllStringSubmatchIndex(str, -1)
	if matches == nil {
		return str
	}
	for _, match := range matches {
		result.WriteString(str[lastEnd:match[0]])
		result.WriteString(jsExpandReplacement(replacement, re, str, match))
		lastEnd = match[1]
	}
	result.WriteString(str[lastEnd:])
	return result.String()
}
