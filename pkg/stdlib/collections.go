// pkg/stdlib/collections.go
// Collection utilities for the Xxlang standard library.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "collections",
		Exports: map[string]objects.Object{
			// Set operations (using arrays)
			"union": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("union() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("union() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("union() requires array arguments")
				}
				seen := make(map[string]bool)
				result := []objects.Object{}
				for _, elem := range arr1.Elements {
					key := elem.Inspect()
					if !seen[key] {
						seen[key] = true
						result = append(result, elem)
					}
				}
				for _, elem := range arr2.Elements {
					key := elem.Inspect()
					if !seen[key] {
						seen[key] = true
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),

			"intersection": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("intersection() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("intersection() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("intersection() requires array arguments")
				}
				seen := make(map[string]bool)
				for _, elem := range arr1.Elements {
					seen[elem.Inspect()] = true
				}
				result := []objects.Object{}
				added := make(map[string]bool)
				for _, elem := range arr2.Elements {
					key := elem.Inspect()
					if seen[key] && !added[key] {
						added[key] = true
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),

			"difference": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("difference() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("difference() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("difference() requires array arguments")
				}
				exclude := make(map[string]bool)
				for _, elem := range arr2.Elements {
					exclude[elem.Inspect()] = true
				}
				result := []objects.Object{}
				for _, elem := range arr1.Elements {
					if !exclude[elem.Inspect()] {
						result = append(result, elem)
					}
				}
				return Array(result...)
			}),

			// Chunk - split array into chunks
			"chunk": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("chunk() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("chunk() requires an array as first argument")
				}
				size, ok := args[1].(*objects.Int)
				if !ok || size.Value <= 0 {
					return Error("chunk() requires a positive integer size")
				}
				s := int(size.Value)
				result := []objects.Object{}
				for i := 0; i < len(arr.Elements); i += s {
					end := i + s
					if end > len(arr.Elements) {
						end = len(arr.Elements)
					}
					result = append(result, Array(arr.Elements[i:end]...))
				}
				return Array(result...)
			}),

			// Zip - combine arrays
			"zip": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("zip() takes at least 2 arguments")
				}
				arrays := make([]*objects.Array, len(args))
				minLen := -1
				for i, arg := range args {
					arr, ok := arg.(*objects.Array)
					if !ok {
						return Error("zip() requires array arguments")
					}
					arrays[i] = arr
					if minLen == -1 || len(arr.Elements) < minLen {
						minLen = len(arr.Elements)
					}
				}
				result := []objects.Object{}
				for i := 0; i < minLen; i++ {
					pair := []objects.Object{}
					for _, arr := range arrays {
						pair = append(pair, arr.Elements[i])
					}
					result = append(result, Array(pair...))
				}
				return Array(result...)
			}),

			// Flatten deep
			"flattenDeep": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("flattenDeep() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("flattenDeep() requires an array argument")
				}
				var flatten func(objs []objects.Object) []objects.Object
				flatten = func(objs []objects.Object) []objects.Object {
					result := []objects.Object{}
					for _, elem := range objs {
						if nested, ok := elem.(*objects.Array); ok {
							result = append(result, flatten(nested.Elements)...)
						} else {
							result = append(result, elem)
						}
					}
					return result
				}
				return Array(flatten(arr.Elements)...)
			}),

			// Count occurrences
			"countBy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("countBy() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("countBy() requires an array argument")
				}
				counts := make(map[string]int64)
				for _, elem := range arr.Elements {
					key := elem.Inspect()
					counts[key]++
				}
				result := make(map[string]objects.Object)
				for k, v := range counts {
					result[k] = Int(v)
				}
				// Convert to array of pairs for easier use
				pairs := []objects.Object{}
				for k, v := range counts {
					pairs = append(pairs, Array(String(k), Int(v)))
				}
				return Array(pairs...)
			}),

			// Group by
			"groupBy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("groupBy() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("groupBy() requires an array as first argument")
				}
				keyFn, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("groupBy() requires a function as second argument")
				}
				groups := make(map[string][]objects.Object)
				for _, elem := range arr.Elements {
					key := keyFn.Fn(elem).Inspect()
					groups[key] = append(groups[key], elem)
				}
				result := []objects.Object{}
				for k, v := range groups {
					result = append(result, Array(String(k), Array(v...)))
				}
				return Array(result...)
			}),

			// Partition
			"partition": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("partition() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("partition() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("partition() requires a function as second argument")
				}
				pass := []objects.Object{}
				fail := []objects.Object{}
				for _, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						pass = append(pass, elem)
					} else {
						fail = append(fail, elem)
					}
				}
				return Array(Array(pass...), Array(fail...))
			}),

			// Take first N
			"take": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("take() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("take() requires an array as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("take() requires an integer as second argument")
				}
				count := int(n.Value)
				if count > len(arr.Elements) {
					count = len(arr.Elements)
				}
				return Array(arr.Elements[:count]...)
			}),

			// Take while
			"takeWhile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("takeWhile() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("takeWhile() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("takeWhile() requires a function as second argument")
				}
				result := []objects.Object{}
				for _, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						result = append(result, elem)
					} else {
						break
					}
				}
				return Array(result...)
			}),

			// Drop first N
			"drop": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("drop() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("drop() requires an array as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("drop() requires an integer as second argument")
				}
				start := int(n.Value)
				if start < 0 {
					start = 0
				}
				if start > len(arr.Elements) {
					return Array()
				}
				return Array(arr.Elements[start:]...)
			}),

			// Find first
			"find": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("find() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("find() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("find() requires a function as second argument")
				}
				for _, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						return elem
					}
				}
				return Null()
			}),

			// Find index
			"findIndex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("findIndex() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("findIndex() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("findIndex() requires a function as second argument")
				}
				for i, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						return Int(int64(i))
					}
				}
				return Int(-1)
			}),

			// Every (all)
			"every": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("every() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("every() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("every() requires a function as second argument")
				}
				for _, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && !b.Value {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Some (any)
			"some": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("some() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("some() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("some() requires a function as second argument")
				}
				for _, elem := range arr.Elements {
					res := pred.Fn(elem)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						return Bool(true)
					}
				}
				return Bool(false)
			}),

			// Range with step
			"rangeStep": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("rangeStep() takes at least 3 arguments")
				}
				start, ok := args[0].(*objects.Int)
				if !ok {
					return Error("rangeStep() requires integer arguments")
				}
				end, ok := args[1].(*objects.Int)
				if !ok {
					return Error("rangeStep() requires integer arguments")
				}
				step, ok := args[2].(*objects.Int)
				if !ok || step.Value == 0 {
					return Error("rangeStep() requires non-zero integer step")
				}
				result := []objects.Object{}
				if step.Value > 0 {
					for i := start.Value; i < end.Value; i += step.Value {
						result = append(result, Int(i))
					}
				} else {
					for i := start.Value; i > end.Value; i += step.Value {
						result = append(result, Int(i))
					}
				}
				return Array(result...)
			}),

			// Repeat element
			"repeat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("repeat() takes exactly 2 arguments")
				}
				n, ok := args[1].(*objects.Int)
				if !ok || n.Value < 0 {
					return Error("repeat() requires non-negative integer count")
				}
				result := make([]objects.Object, n.Value)
				for i := range result {
					result[i] = args[0]
				}
				return Array(result...)
			}),

			// Shuffle
			"shuffle": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("shuffle() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("shuffle() requires an array argument")
				}
				// Fisher-Yates shuffle with simple PRNG
				n := len(arr.Elements)
				result := make([]objects.Object, n)
				copy(result, arr.Elements)
				for i := n - 1; i > 0; i-- {
					j := int(simpleRand() % uint64(i+1))
					result[i], result[j] = result[j], result[i]
				}
				return Array(result...)
			}),

			// Sample
			"sample": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("sample() takes at least 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("sample() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				count := int64(1)
				if len(args) > 1 {
					n, ok := args[1].(*objects.Int)
					if ok && n.Value > 0 {
						count = n.Value
					}
				}
				if count == 1 {
					idx := int(simpleRand() % uint64(len(arr.Elements)))
					return arr.Elements[idx]
				}
				// Return multiple samples
				result := []objects.Object{}
				for i := int64(0); i < count && i < int64(len(arr.Elements)); i++ {
					idx := int(simpleRand() % uint64(len(arr.Elements)))
					result = append(result, arr.Elements[idx])
				}
				return Array(result...)
			}),
		},
	})
}

// Simple PRNG
var randState uint64 = 12345

func simpleRand() uint64 {
	randState = randState*1103515245 + 12345
	return randState
}
