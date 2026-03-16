// pkg/stdlib/sort.go
// Sorting utilities for the Xxlang standard library.
package stdlib

import (
	"sort"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "sort",
		Exports: map[string]objects.Object{
			// Sort numbers ascending
			"numbers": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("numbers() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("numbers() requires an array argument")
				}
				// Extract numbers
				nums := make([]float64, 0, len(arr.Elements))
				for _, elem := range arr.Elements {
					switch v := elem.(type) {
					case *objects.Int:
						nums = append(nums, float64(v.Value))
					case *objects.Float:
						nums = append(nums, v.Value)
					default:
						return Error("numbers() requires numeric array elements")
					}
				}
				sort.Float64s(nums)
				result := make([]objects.Object, len(nums))
				for i, n := range nums {
					if n == float64(int64(n)) {
						result[i] = Int(int64(n))
					} else {
						result[i] = Float(n)
					}
				}
				return Array(result...)
			}),

			// Sort numbers descending
			"numbersDesc": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("numbersDesc() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("numbersDesc() requires an array argument")
				}
				nums := make([]float64, 0, len(arr.Elements))
				for _, elem := range arr.Elements {
					switch v := elem.(type) {
					case *objects.Int:
						nums = append(nums, float64(v.Value))
					case *objects.Float:
						nums = append(nums, v.Value)
					}
				}
				sort.Sort(sort.Reverse(sort.Float64Slice(nums)))
				result := make([]objects.Object, len(nums))
				for i, n := range nums {
					if n == float64(int64(n)) {
						result[i] = Int(int64(n))
					} else {
						result[i] = Float(n)
					}
				}
				return Array(result...)
			}),

			// Sort strings ascending
			"strings": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("strings() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("strings() requires an array argument")
				}
				strs := make([]string, 0, len(arr.Elements))
				for _, elem := range arr.Elements {
					s, ok := elem.(*objects.String)
					if !ok {
						return Error("strings() requires string array elements")
					}
					strs = append(strs, s.Value)
				}
				sort.Strings(strs)
				result := make([]objects.Object, len(strs))
				for i, s := range strs {
					result[i] = String(s)
				}
				return Array(result...)
			}),

			// Sort strings descending
			"stringsDesc": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stringsDesc() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("stringsDesc() requires an array argument")
				}
				strs := make([]string, 0, len(arr.Elements))
				for _, elem := range arr.Elements {
					s, ok := elem.(*objects.String)
					if !ok {
						return Error("stringsDesc() requires string array elements")
					}
					strs = append(strs, s.Value)
				}
				sort.Sort(sort.Reverse(sort.StringSlice(strs)))
				result := make([]objects.Object, len(strs))
				for i, s := range strs {
					result[i] = String(s)
				}
				return Array(result...)
			}),

			// Sort by key function
			"by": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("by() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("by() requires an array as first argument")
				}
				keyFn, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("by() requires a function as second argument")
				}
				// Create index-value pairs with keys
				type kv struct {
					idx int
					key float64
				}
				pairs := make([]kv, len(arr.Elements))
				for i, elem := range arr.Elements {
					res := keyFn.Fn(elem)
					switch v := res.(type) {
					case *objects.Int:
						pairs[i] = kv{idx: i, key: float64(v.Value)}
					case *objects.Float:
						pairs[i] = kv{idx: i, key: v.Value}
					default:
						return Error("key function must return numeric value")
					}
				}
				sort.Slice(pairs, func(i, j int) bool {
					return pairs[i].key < pairs[j].key
				})
				result := make([]objects.Object, len(arr.Elements))
				for i, p := range pairs {
					result[i] = arr.Elements[p.idx]
				}
				return Array(result...)
			}),

			// Reverse array
			"reverse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("reverse() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("reverse() requires an array argument")
				}
				n := len(arr.Elements)
				result := make([]objects.Object, n)
				for i, elem := range arr.Elements {
					result[n-1-i] = elem
				}
				return Array(result...)
			}),

			// Check if sorted
			"isSorted": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSorted() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("isSorted() requires an array argument")
				}
				if len(arr.Elements) <= 1 {
					return Bool(true)
				}
				for i := 1; i < len(arr.Elements); i++ {
					// Compare adjacent elements
					prev := arr.Elements[i-1]
					curr := arr.Elements[i]
					if compareObjects(prev, curr) > 0 {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Min element
			"min": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("min() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("min() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				minIdx := 0
				for i := 1; i < len(arr.Elements); i++ {
					if compareObjects(arr.Elements[i], arr.Elements[minIdx]) < 0 {
						minIdx = i
					}
				}
				return arr.Elements[minIdx]
			}),

			// Max element
			"max": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("max() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("max() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Null()
				}
				maxIdx := 0
				for i := 1; i < len(arr.Elements); i++ {
					if compareObjects(arr.Elements[i], arr.Elements[maxIdx]) > 0 {
						maxIdx = i
					}
				}
				return arr.Elements[maxIdx]
			}),

			// Min index
			"minIndex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("minIndex() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("minIndex() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Int(-1)
				}
				minIdx := 0
				for i := 1; i < len(arr.Elements); i++ {
					if compareObjects(arr.Elements[i], arr.Elements[minIdx]) < 0 {
						minIdx = i
					}
				}
				return Int(int64(minIdx))
			}),

			// Max index
			"maxIndex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("maxIndex() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("maxIndex() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Int(-1)
				}
				maxIdx := 0
				for i := 1; i < len(arr.Elements); i++ {
					if compareObjects(arr.Elements[i], arr.Elements[maxIdx]) > 0 {
						maxIdx = i
					}
				}
				return Int(int64(maxIdx))
			}),
		},
	})
}

// compareObjects returns -1, 0, or 1
func compareObjects(a, b objects.Object) int {
	// Try numeric comparison first
	var av, bv float64
	var aIsNum, bIsNum bool
	switch v := a.(type) {
	case *objects.Int:
		av = float64(v.Value)
		aIsNum = true
	case *objects.Float:
		av = v.Value
		aIsNum = true
	}
	switch v := b.(type) {
	case *objects.Int:
		bv = float64(v.Value)
		bIsNum = true
	case *objects.Float:
		bv = v.Value
		bIsNum = true
	}
	if aIsNum && bIsNum {
		if av < bv {
			return -1
		} else if av > bv {
			return 1
		}
		return 0
	}
	// Fall back to string comparison
	as := a.Inspect()
	bs := b.Inspect()
	if as < bs {
		return -1
	} else if as > bs {
		return 1
	}
	return 0
}
