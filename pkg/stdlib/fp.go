// pkg/stdlib/fp.go
// Functional programming utilities for the Xxlang standard library.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/fp",
		Exports: map[string]objects.Object{
			// Compose functions (right to left)
			"compose": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("compose() takes at least 2 arguments")
				}
				// Return a composed function
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					result := innerArgs[0]
					// Apply functions from right to left
					for i := len(args) - 1; i >= 0; i-- {
						fn, ok := args[i].(*objects.Builtin)
						if !ok {
							return Error("compose() requires function arguments")
						}
						result = fn.Fn(result)
					}
					return result
				})
			}),

			// Pipe functions (left to right)
			"pipe": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("pipe() takes at least 2 arguments")
				}
				// Return a piped function
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					result := innerArgs[0]
					// Apply functions from left to right
					for _, arg := range args {
						fn, ok := arg.(*objects.Builtin)
						if !ok {
							return Error("pipe() requires function arguments")
						}
						result = fn.Fn(result)
					}
					return result
				})
			}),

			// Identity function
			"identity": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("identity() takes exactly 1 argument")
				}
				return args[0]
			}),

			// Constant function
			"constant": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("constant() takes exactly 1 argument")
				}
				val := args[0]
				return BuiltinFunc(func(...objects.Object) objects.Object {
					return val
				})
			}),

			// Always true
			"alwaysTrue": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Bool(true)
			}),

			// Always false
			"alwaysFalse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Bool(false)
			}),

			// Negate a predicate
			"not": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("not() takes exactly 1 argument")
				}
				fn, ok := args[0].(*objects.Builtin)
				if !ok {
					return Error("not() requires a function argument")
				}
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					result := fn.Fn(innerArgs...)
					if b, ok := result.(*objects.Bool); ok {
						return Bool(!b.Value)
					}
					return Error("not() predicate must return boolean")
				})
			}),

			// All predicates must be true
			"allPass": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("allPass() takes at least 2 arguments")
				}
				predicates := args
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					for _, pred := range predicates {
						fn, ok := pred.(*objects.Builtin)
						if !ok {
							return Error("allPass() requires function arguments")
						}
						result := fn.Fn(innerArgs...)
						if b, ok := result.(*objects.Bool); !ok || !b.Value {
							return Bool(false)
						}
					}
					return Bool(true)
				})
			}),

			// Any predicate must be true
			"anyPass": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("anyPass() takes at least 2 arguments")
				}
				predicates := args
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					for _, pred := range predicates {
						fn, ok := pred.(*objects.Builtin)
						if !ok {
							return Error("anyPass() requires function arguments")
						}
						result := fn.Fn(innerArgs...)
						if b, ok := result.(*objects.Bool); ok && b.Value {
							return Bool(true)
						}
					}
					return Bool(false)
				})
			}),

			// Tap (side effect without modifying value)
			"tap": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("tap() takes exactly 1 argument")
				}
				fn, ok := args[0].(*objects.Builtin)
				if !ok {
					return Error("tap() requires a function argument")
				}
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					fn.Fn(innerArgs...)
					return innerArgs[0]
				})
			}),

			// Default value if null
			"defaultTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("defaultTo() takes exactly 2 arguments")
				}
				defaultVal := args[0]
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) != 1 {
						return Error("defaultTo() function takes exactly 1 argument")
					}
					if _, ok := innerArgs[0].(*objects.Null); ok {
						return defaultVal
					}
					return innerArgs[0]
				})
			}),

			// Equals
			"equals": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("equals() takes exactly 1 argument")
				}
				target := args[0]
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) != 1 {
						return Error("equals() function takes exactly 1 argument")
					}
					return Bool(innerArgs[0].Inspect() == target.Inspect())
				})
			}),

			// Prop (get property from object/map)
			"prop": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("prop() takes exactly 1 argument")
				}
				key := args[0]
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) != 1 {
						return Error("prop() function takes exactly 1 argument")
					}
					m, ok := innerArgs[0].(*objects.Map)
					if !ok {
						return Null()
					}
					keyStr, ok := key.(*objects.String)
					if !ok {
						return Null()
					}
					for _, pair := range m.Pairs {
						if k, ok := pair.Key.(*objects.String); ok && k.Value == keyStr.Value {
							return pair.Value
						}
					}
					return Null()
				})
			}),

			// Pick properties from map
			"pick": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("pick() takes exactly 1 argument")
				}
				keys, ok := args[0].(*objects.Array)
				if !ok {
					return Error("pick() requires an array of keys")
				}
				keySet := make(map[string]bool)
				for _, k := range keys.Elements {
					if s, ok := k.(*objects.String); ok {
						keySet[s.Value] = true
					}
				}
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) != 1 {
						return Error("pick() function takes exactly 1 argument")
					}
					m, ok := innerArgs[0].(*objects.Map)
					if !ok {
						return Error("pick() function requires a map argument")
					}
					pairs := make(map[objects.HashKey]objects.MapPair)
					for _, pair := range m.Pairs {
						if k, ok := pair.Key.(*objects.String); ok && keySet[k.Value] {
							pairs[k.HashKey()] = pair
						}
					}
					return &objects.Map{Pairs: pairs}
				})
			}),

			// Omit properties from map
			"omit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("omit() takes exactly 1 argument")
				}
				keys, ok := args[0].(*objects.Array)
				if !ok {
					return Error("omit() requires an array of keys")
				}
				keySet := make(map[string]bool)
				for _, k := range keys.Elements {
					if s, ok := k.(*objects.String); ok {
						keySet[s.Value] = true
					}
				}
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) != 1 {
						return Error("omit() function takes exactly 1 argument")
					}
					m, ok := innerArgs[0].(*objects.Map)
					if !ok {
						return Error("omit() function requires a map argument")
					}
					pairs := make(map[objects.HashKey]objects.MapPair)
					for _, pair := range m.Pairs {
						if k, ok := pair.Key.(*objects.String); ok && !keySet[k.Value] {
							pairs[k.HashKey()] = pair
						}
					}
					return &objects.Map{Pairs: pairs}
				})
			}),

			// Merge two arrays
			"concat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("concat() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("concat() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("concat() requires array arguments")
				}
				result := make([]objects.Object, len(arr1.Elements)+len(arr2.Elements))
				copy(result, arr1.Elements)
				copy(result[len(arr1.Elements):], arr2.Elements)
				return Array(result...)
			}),

			// Flatten array (one level)
			"flatten": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("flatten() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("flatten() requires an array argument")
				}
				result := []objects.Object{}
				for _, elem := range arr.Elements {
					if nested, ok := elem.(*objects.Array); ok {
						result = append(result, nested.Elements...)
					} else {
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),

			// Head (first element)
			"head": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("head() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("head() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				return arr.Elements[0]
			}),

			// Tail (all but first)
			"tail": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("tail() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("tail() requires an array argument")
				}
				if len(arr.Elements) <= 1 {
					return Array()
				}
				return Array(arr.Elements[1:]...)
			}),

			// Init (all but last)
			"init": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("init() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("init() requires an array argument")
				}
				if len(arr.Elements) <= 1 {
					return Array()
				}
				return Array(arr.Elements[:len(arr.Elements)-1]...)
			}),

			// Last element
			"last": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("last() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("last() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				return arr.Elements[len(arr.Elements)-1]
			}),

			// Length
			"length": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("length() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Array:
					return Int(int64(len(v.Elements)))
				case *objects.String:
					return Int(int64(len(v.Value)))
				case *objects.Map:
					return Int(int64(len(v.Pairs)))
				default:
					return Error("length() requires array, string, or map argument")
				}
			}),

			// Is empty
			"isEmpty": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isEmpty() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Array:
					return Bool(len(v.Elements) == 0)
				case *objects.String:
					return Bool(len(v.Value) == 0)
				case *objects.Map:
					return Bool(len(v.Pairs) == 0)
				case *objects.Null:
					return Bool(true)
				default:
					return Bool(false)
				}
			}),

			// Range
			"range": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("range() takes at least 1 argument")
				}
				var start, end int64
				var step int64 = 1
				if len(args) == 1 {
					start = 0
					n, ok := args[0].(*objects.Int)
					if !ok {
						return Error("range() requires integer arguments")
					}
					end = n.Value
				} else {
					s, ok := args[0].(*objects.Int)
					if !ok {
						return Error("range() requires integer arguments")
					}
					e, ok := args[1].(*objects.Int)
					if !ok {
						return Error("range() requires integer arguments")
					}
					start = s.Value
					end = e.Value
					if len(args) > 2 {
						st, ok := args[2].(*objects.Int)
						if ok && st.Value != 0 {
							step = st.Value
						}
					}
				}
				result := []objects.Object{}
				if step > 0 {
					for i := start; i < end; i += step {
						result = append(result, Int(i))
					}
				} else if step < 0 {
					for i := start; i > end; i += step {
						result = append(result, Int(i))
					}
				}
				return Array(result...)
			}),

			// Times (repeat function n times)
			"times": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("times() takes exactly 2 arguments")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("times() requires an integer as first argument")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("times() requires a function as second argument")
				}
				result := []objects.Object{}
				for i := int64(0); i < n.Value; i++ {
					result = append(result, fn.Fn(Int(i)))
				}
				return Array(result...)
			}),

			// Memoize (cache function results)
			"memoize": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("memoize() takes exactly 1 argument")
				}
				fn, ok := args[0].(*objects.Builtin)
				if !ok {
					return Error("memoize() requires a function argument")
				}
				cache := make(map[string]objects.Object)
				return BuiltinFunc(func(innerArgs ...objects.Object) objects.Object {
					key := ""
					for _, arg := range innerArgs {
						key += arg.Inspect()
					}
					if cached, ok := cache[key]; ok {
						return cached
					}
					result := fn.Fn(innerArgs...)
					cache[key] = result
					return result
				})
			}),

			// Until (apply function until predicate is true)
			"until": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("until() takes exactly 3 arguments")
				}
				pred, ok := args[0].(*objects.Builtin)
				if !ok {
					return Error("until() requires a predicate function")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("until() requires a function")
				}
				val := args[2]
				maxIter := 1000
				for i := 0; i < maxIter; i++ {
					result := pred.Fn(val)
					if b, ok := result.(*objects.Bool); ok && b.Value {
						return val
					}
					val = fn.Fn(val)
				}
				return Error("until() exceeded maximum iterations")
			}),
		},
	})
}
