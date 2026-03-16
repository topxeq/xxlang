// pkg/stdlib/array.go
// Array utilities for the Xxlang standard library.
package stdlib

import (
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "array",
		Exports: map[string]objects.Object{
			"len": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("len() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("len() requires an array argument")
				}
				return Int(int64(len(arr.Elements)))
			}),

			"push": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("push() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("push() requires an array as first argument")
				}
				arr.Elements = append(arr.Elements, args[1:]...)
				return arr
			}),

			"pop": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("pop() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("pop() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				last := arr.Elements[len(arr.Elements)-1]
				arr.Elements = arr.Elements[:len(arr.Elements)-1]
				return last
			}),

			"shift": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("shift() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("shift() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				first := arr.Elements[0]
				arr.Elements = arr.Elements[1:]
				return first
			}),

			"unshift": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("unshift() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("unshift() requires an array as first argument")
				}
				arr.Elements = append(args[1:], arr.Elements...)
				return arr
			}),

			"first": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("first() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("first() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				return arr.Elements[0]
			}),

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

			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("get() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("get() requires an array as first argument")
				}
				var idx int
				switch v := args[1].(type) {
				case *objects.Int:
					idx = int(v.Value)
				case *objects.Float:
					idx = int(v.Value)
				default:
					return Error("get() requires an integer index")
				}
				if idx < 0 || idx >= len(arr.Elements) {
					return Null()
				}
				return arr.Elements[idx]
			}),

			"set": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("set() takes exactly 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("set() requires an array as first argument")
				}
				var idx int
				switch v := args[1].(type) {
				case *objects.Int:
					idx = int(v.Value)
				case *objects.Float:
					idx = int(v.Value)
				default:
					return Error("set() requires an integer index")
				}
				if idx < 0 || idx >= len(arr.Elements) {
					return Error("set() index out of range")
				}
				arr.Elements[idx] = args[2]
				return arr
			}),

			"contains": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("contains() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("contains() requires an array as first argument")
				}
				for _, elem := range arr.Elements {
					if elem.Inspect() == args[1].Inspect() {
						return Bool(true)
					}
				}
				return Bool(false)
			}),

			"indexOf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("indexOf() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("indexOf() requires an array as first argument")
				}
				for i, elem := range arr.Elements {
					if elem.Inspect() == args[1].Inspect() {
						return Int(int64(i))
					}
				}
				return Int(-1)
			}),

			"reverse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("reverse() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("reverse() requires an array argument")
				}
				for i, j := 0, len(arr.Elements)-1; i < j; i, j = i+1, j-1 {
					arr.Elements[i], arr.Elements[j] = arr.Elements[j], arr.Elements[i]
				}
				return arr
			}),

			"slice": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("slice() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("slice() requires an array as first argument")
				}
				var start int
				switch v := args[1].(type) {
				case *objects.Int:
					start = int(v.Value)
				case *objects.Float:
					start = int(v.Value)
				default:
					return Error("slice() requires integer indices")
				}
				end := len(arr.Elements)
				if len(args) > 2 {
					switch v := args[2].(type) {
					case *objects.Int:
						end = int(v.Value)
					case *objects.Float:
						end = int(v.Value)
					default:
						return Error("slice() requires integer indices")
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(arr.Elements) {
					end = len(arr.Elements)
				}
				if start > end {
					return Array()
				}
				return Array(arr.Elements[start:end]...)
			}),

			"concat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("concat() takes at least 2 arguments")
				}
				result := Array()
				for _, arg := range args {
					arr, ok := arg.(*objects.Array)
					if !ok {
						return Error("concat() requires array arguments")
					}
					result.Elements = append(result.Elements, arr.Elements...)
				}
				return result
			}),

			"join": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("join() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("join() requires an array as first argument")
				}
				sep, ok := args[1].(*objects.String)
				if !ok {
					return Error("join() requires a string separator")
				}
				strs := make([]string, len(arr.Elements))
				for i, elem := range arr.Elements {
					strs[i] = elem.Inspect()
				}
				return String(strings.Join(strs, sep.Value))
			}),

			"map": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("map() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("map() requires an array as first argument")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					// Also accept compiled functions
					if _, ok := args[1].(*objects.CompiledFunction); ok {
						return Error("map() with user functions not yet supported")
					}
					return Error("map() requires a function as second argument")
				}
				result := make([]objects.Object, len(arr.Elements))
				for i, elem := range arr.Elements {
					result[i] = fn.Fn(elem)
				}
				return Array(result...)
			}),

			"filter": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("filter() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("filter() requires an array as first argument")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					if _, ok := args[1].(*objects.CompiledFunction); ok {
						return Error("filter() with user functions not yet supported")
					}
					return Error("filter() requires a function as second argument")
				}
				result := []objects.Object{}
				for _, elem := range arr.Elements {
					res := fn.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),

			"reduce": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("reduce() takes at least 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("reduce() requires an array as first argument")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					if _, ok := args[1].(*objects.CompiledFunction); ok {
						return Error("reduce() with user functions not yet supported")
					}
					return Error("reduce() requires a function as second argument")
				}
				accumulator := args[2]
				for _, elem := range arr.Elements {
					accumulator = fn.Fn(accumulator, elem)
				}
				return accumulator
			}),

			"fill": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("fill() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("fill() requires an array as first argument")
				}
				value := args[1]
				start := 0
				end := len(arr.Elements)
				if len(args) > 2 {
					switch v := args[2].(type) {
					case *objects.Int:
						start = int(v.Value)
					case *objects.Float:
						start = int(v.Value)
					}
				}
				if len(args) > 3 {
					switch v := args[3].(type) {
					case *objects.Int:
						end = int(v.Value)
					case *objects.Float:
						end = int(v.Value)
					}
				}
				for i := start; i < end && i < len(arr.Elements); i++ {
					if i >= 0 {
						arr.Elements[i] = value
					}
				}
				return arr
			}),

			"sort": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("sort() takes at least 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("sort() requires an array argument")
				}
				// Simple numeric sort (ascending)
				// For now, just return the array as-is since proper sorting needs a comparator
				return arr
			}),

			"isEmpty": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isEmpty() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("isEmpty() requires an array argument")
				}
				return Bool(len(arr.Elements) == 0)
			}),

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

			"unique": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unique() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("unique() requires an array argument")
				}
				seen := make(map[string]bool)
				result := []objects.Object{}
				for _, elem := range arr.Elements {
					key := elem.Inspect()
					if !seen[key] {
						seen[key] = true
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),
		},
	})
}
