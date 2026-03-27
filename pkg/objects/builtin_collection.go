// pkg/objects/builtin_collection.go
// Array and collection enhancement built-in functions for Xxlang
package objects

import (
	"math"
)

func init() {
	Builtins["mapArray"] = &Builtin{Fn: builtinMapArray}
	Builtins["filterArray"] = &Builtin{Fn: builtinFilterArray}
	Builtins["reduceArray"] = &Builtin{Fn: builtinReduceArray}
	Builtins["forEach"] = &Builtin{Fn: builtinForEach}
	Builtins["flatMap"] = &Builtin{Fn: builtinFlatMap}
	Builtins["every"] = &Builtin{Fn: builtinEvery}
	Builtins["some"] = &Builtin{Fn: builtinSome}
	Builtins["groupBy"] = &Builtin{Fn: builtinGroupBy}
	Builtins["partition"] = &Builtin{Fn: builtinPartition}
	Builtins["zip"] = &Builtin{Fn: builtinZip}
	Builtins["unzip"] = &Builtin{Fn: builtinUnzip}
	Builtins["fill"] = &Builtin{Fn: builtinFill}
	Builtins["rangeNum"] = &Builtin{Fn: builtinRangeNum}
	Builtins["intersection"] = &Builtin{Fn: builtinIntersection}
	Builtins["difference"] = &Builtin{Fn: builtinDifference}
	Builtins["union"] = &Builtin{Fn: builtinUnion}
	Builtins["countBy"] = &Builtin{Fn: builtinCountBy}
	Builtins["sortBy"] = &Builtin{Fn: builtinSortBy}
}

// builtinMapArray - apply function to each element and return new array
// Usage: mapArray(arr, fn) -> array
func builtinMapArray(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for mapArray. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'mapArray' must be ARRAY, got %s", args[0].Type())
	}

	result := make([]Object, len(arr.Elements))
	for i, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("mapArray callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}
		result[i] = ret
	}

	return NewArray(result)
}

// builtinFilterArray - filter array elements that satisfy condition
// Usage: filterArray(arr, fn) -> array
func builtinFilterArray(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for filterArray. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'filterArray' must be ARRAY, got %s", args[0].Type())
	}

	var result []Object
	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("filterArray callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		if IsTruthy(ret) {
			result = append(result, elem)
		}
	}

	return NewArray(result)
}

// builtinReduceArray - reduce array to single value
// Usage: reduceArray(arr, fn) -> value
//
//	reduceArray(arr, fn, initial) -> value
func builtinReduceArray(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for reduceArray. got=%d, want=2 or 3", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'reduceArray' must be ARRAY, got %s", args[0].Type())
	}

	if len(arr.Elements) == 0 {
		if len(args) == 3 {
			return args[2]
		}
		return NULL
	}

	var accumulator Object
	startIndex := 0

	if len(args) == 3 {
		accumulator = args[2]
	} else {
		accumulator = arr.Elements[0]
		startIndex = 1
	}

	for i := startIndex; i < len(arr.Elements); i++ {
		ret, err := CallUserFunc(args[1], accumulator, arr.Elements[i])
		if err != nil {
			return newError("reduceArray callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}
		accumulator = ret
	}

	return accumulator
}

// builtinForEach - iterate array and execute callback
// Usage: forEach(arr, fn) -> null
func builtinForEach(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for forEach. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'forEach' must be ARRAY, got %s", args[0].Type())
	}

	for i, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem, NewInt(int64(i)))
		if err != nil {
			return newError("forEach callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}
	}

	return NULL
}

// builtinFlatMap - map then flatten
// Usage: flatMap(arr, fn) -> array
func builtinFlatMap(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for flatMap. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'flatMap' must be ARRAY, got %s", args[0].Type())
	}

	var result []Object
	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("flatMap callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		if subArr, ok := ret.(*Array); ok {
			result = append(result, subArr.Elements...)
		} else {
			result = append(result, ret)
		}
	}

	return NewArray(result)
}

// builtinEvery - check if all elements satisfy condition
// Usage: every(arr, fn) -> bool
func builtinEvery(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for every. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'every' must be ARRAY, got %s", args[0].Type())
	}

	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("every callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		if !IsTruthy(ret) {
			return FALSE
		}
	}

	return TRUE
}

// builtinSome - check if any element satisfies condition
// Usage: some(arr, fn) -> bool
func builtinSome(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for some. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'some' must be ARRAY, got %s", args[0].Type())
	}

	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("some callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		if IsTruthy(ret) {
			return TRUE
		}
	}

	return FALSE
}

// builtinGroupBy - group array elements by key function
// Usage: groupBy(arr, fn) -> map
func builtinGroupBy(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for groupBy. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'groupBy' must be ARRAY, got %s", args[0].Type())
	}

	groups := make(map[string][]Object)

	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("groupBy callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		key := ret.Inspect()
		groups[key] = append(groups[key], elem)
	}

	result := NewMapWithCapacity(len(groups))
	for k, v := range groups {
		keyObj := NewString(k)
		hashKey := keyObj.HashKey()
		result.Pairs[hashKey] = MapPair{Key: keyObj, Value: NewArray(v)}
	}

	return result
}

// builtinPartition - split array into two groups by condition
// Usage: partition(arr, fn) -> [trueGroup, falseGroup]
func builtinPartition(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for partition. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'partition' must be ARRAY, got %s", args[0].Type())
	}

	var trueGroup, falseGroup []Object

	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("partition callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		if IsTruthy(ret) {
			trueGroup = append(trueGroup, elem)
		} else {
			falseGroup = append(falseGroup, elem)
		}
	}

	return NewArray([]Object{NewArray(trueGroup), NewArray(falseGroup)})
}

// builtinZip - combine multiple arrays into array of pairs
// Usage: zip(arr1, arr2) -> array
//
//	zip(arr1, arr2, arr3, ...) -> array
func builtinZip(args ...Object) Object {
	if len(args) < 2 {
		return newError("wrong number of arguments for zip. got=%d, want>=2", len(args))
	}

	arrays := make([]*Array, len(args))
	minLen := math.MaxInt32

	for i, arg := range args {
		arr, ok := arg.(*Array)
		if !ok {
			return newError("argument %d to 'zip' must be ARRAY, got %s", i+1, arg.Type())
		}
		arrays[i] = arr
		if len(arr.Elements) < minLen {
			minLen = len(arr.Elements)
		}
	}

	if minLen == 0 {
		return NewArray([]Object{})
	}

	result := make([]Object, minLen)
	for i := 0; i < minLen; i++ {
		pair := make([]Object, len(arrays))
		for j, arr := range arrays {
			pair[j] = arr.Elements[i]
		}
		result[i] = NewArray(pair)
	}

	return NewArray(result)
}

// builtinUnzip - split array of pairs into separate arrays
// Usage: unzip(arr) -> array of arrays
func builtinUnzip(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for unzip. got=%d, want=1", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("argument to 'unzip' must be ARRAY, got %s", args[0].Type())
	}

	if len(arr.Elements) == 0 {
		return NewArray([]Object{})
	}

	firstPair, ok := arr.Elements[0].(*Array)
	if !ok {
		return newError("elements of array must be ARRAY for unzip")
	}

	numArrays := len(firstPair.Elements)
	resultArrays := make([][]Object, numArrays)
	for i := 0; i < numArrays; i++ {
		resultArrays[i] = []Object{}
	}

	for _, elem := range arr.Elements {
		pair, ok := elem.(*Array)
		if !ok {
			return newError("elements of array must be ARRAY for unzip")
		}

		for i := 0; i < numArrays && i < len(pair.Elements); i++ {
			resultArrays[i] = append(resultArrays[i], pair.Elements[i])
		}
	}

	result := make([]Object, numArrays)
	for i, arr := range resultArrays {
		result[i] = NewArray(arr)
	}

	return NewArray(result)
}

// builtinFill - fill array range with value
// Usage: fill(arr, value) -> array
//
//	fill(arr, value, start) -> array
//
//	fill(arr, value, start, end) -> array
func builtinFill(args ...Object) Object {
	if len(args) < 2 || len(args) > 4 {
		return newError("wrong number of arguments for fill. got=%d, want=2-4", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'fill' must be ARRAY, got %s", args[0].Type())
	}

	value := args[1]
	start := 0
	end := len(arr.Elements)

	if len(args) >= 3 {
		s, ok := args[2].(*Int)
		if !ok {
			return newError("start index must be INT, got %s", args[2].Type())
		}
		start = int(s.Value)
		if start < 0 {
			start = len(arr.Elements) + start
		}
	}

	if len(args) >= 4 {
		e, ok := args[3].(*Int)
		if !ok {
			return newError("end index must be INT, got %s", args[3].Type())
		}
		end = int(e.Value)
		if end < 0 {
			end = len(arr.Elements) + end
		}
	}

	if start < 0 {
		start = 0
	}
	if end > len(arr.Elements) {
		end = len(arr.Elements)
	}

	result := make([]Object, len(arr.Elements))
	copy(result, arr.Elements)

	for i := start; i < end && i < len(result); i++ {
		result[i] = value
	}

	return NewArray(result)
}

// builtinRangeNum - generate range of numbers
// Usage: rangeNum(end) -> array
//
//	rangeNum(start, end) -> array
//
//	rangeNum(start, end, step) -> array
func builtinRangeNum(args ...Object) Object {
	if len(args) < 1 || len(args) > 3 {
		return newError("wrong number of arguments for rangeNum. got=%d, want=1-3", len(args))
	}

	var start, end, step int64 = 0, 0, 1

	if len(args) == 1 {
		e, ok := args[0].(*Int)
		if !ok {
			return newError("argument must be INT, got %s", args[0].Type())
		}
		end = e.Value
	} else if len(args) >= 2 {
		s, ok := args[0].(*Int)
		if !ok {
			return newError("start must be INT, got %s", args[0].Type())
		}
		e, ok := args[1].(*Int)
		if !ok {
			return newError("end must be INT, got %s", args[1].Type())
		}
		start = s.Value
		end = e.Value
	}

	if len(args) == 3 {
		st, ok := args[2].(*Int)
		if !ok {
			return newError("step must be INT, got %s", args[2].Type())
		}
		step = st.Value
		if step == 0 {
			return newError("step cannot be zero")
		}
	}

	var result []Object

	if step > 0 {
		for i := start; i < end; i += step {
			result = append(result, NewInt(i))
		}
	} else {
		for i := start; i > end; i += step {
			result = append(result, NewInt(i))
		}
	}

	return NewArray(result)
}

// builtinIntersection - find intersection of two arrays
// Usage: intersection(arr1, arr2) -> array
func builtinIntersection(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for intersection. got=%d, want=2", len(args))
	}

	arr1, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'intersection' must be ARRAY, got %s", args[0].Type())
	}

	arr2, ok := args[1].(*Array)
	if !ok {
		return newError("second argument to 'intersection' must be ARRAY, got %s", args[1].Type())
	}

	set2 := make(map[string]bool)
	for _, elem := range arr2.Elements {
		set2[elem.Inspect()] = true
	}

	seen := make(map[string]bool)
	var result []Object

	for _, elem := range arr1.Elements {
		key := elem.Inspect()
		if set2[key] && !seen[key] {
			result = append(result, elem)
			seen[key] = true
		}
	}

	return NewArray(result)
}

// builtinDifference - find elements in arr1 but not in arr2
// Usage: difference(arr1, arr2) -> array
func builtinDifference(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for difference. got=%d, want=2", len(args))
	}

	arr1, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'difference' must be ARRAY, got %s", args[0].Type())
	}

	arr2, ok := args[1].(*Array)
	if !ok {
		return newError("second argument to 'difference' must be ARRAY, got %s", args[1].Type())
	}

	set2 := make(map[string]bool)
	for _, elem := range arr2.Elements {
		set2[elem.Inspect()] = true
	}

	var result []Object
	for _, elem := range arr1.Elements {
		if !set2[elem.Inspect()] {
			result = append(result, elem)
		}
	}

	return NewArray(result)
}

// builtinUnion - union of two arrays (unique elements)
// Usage: union(arr1, arr2) -> array
func builtinUnion(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for union. got=%d, want=2", len(args))
	}

	arr1, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'union' must be ARRAY, got %s", args[0].Type())
	}

	arr2, ok := args[1].(*Array)
	if !ok {
		return newError("second argument to 'union' must be ARRAY, got %s", args[1].Type())
	}

	seen := make(map[string]bool)
	var result []Object

	for _, elem := range arr1.Elements {
		key := elem.Inspect()
		if !seen[key] {
			result = append(result, elem)
			seen[key] = true
		}
	}

	for _, elem := range arr2.Elements {
		key := elem.Inspect()
		if !seen[key] {
			result = append(result, elem)
			seen[key] = true
		}
	}

	return NewArray(result)
}

// builtinCountBy - count elements by key function
// Usage: countBy(arr, fn) -> map
func builtinCountBy(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for countBy. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'countBy' must be ARRAY, got %s", args[0].Type())
	}

	counts := make(map[string]int64)

	for _, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("countBy callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}

		key := ret.Inspect()
		counts[key]++
	}

	result := NewMapWithCapacity(len(counts))
	for k, v := range counts {
		keyObj := NewString(k)
		hashKey := keyObj.HashKey()
		result.Pairs[hashKey] = MapPair{Key: keyObj, Value: NewInt(v)}
	}

	return result
}

// builtinSortBy - sort array by key function
// Usage: sortBy(arr, fn) -> array
func builtinSortBy(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for sortBy. got=%d, want=2", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'sortBy' must be ARRAY, got %s", args[0].Type())
	}

	type indexed struct {
		original Object
		key      Object
	}

	indexedElems := make([]indexed, len(arr.Elements))
	for i, elem := range arr.Elements {
		ret, err := CallUserFunc(args[1], elem)
		if err != nil {
			return newError("sortBy callback error: %v", err)
		}
		if e, ok := ret.(*Error); ok {
			return e
		}
		indexedElems[i] = indexed{original: elem, key: ret}
	}

	for i := 0; i < len(indexedElems)-1; i++ {
		for j := i + 1; j < len(indexedElems); j++ {
			if compareForSort(indexedElems[i].key, indexedElems[j].key) > 0 {
				indexedElems[i], indexedElems[j] = indexedElems[j], indexedElems[i]
			}
		}
	}

	result := make([]Object, len(indexedElems))
	for i, ie := range indexedElems {
		result[i] = ie.original
	}

	return NewArray(result)
}

// compareForSort compares two objects for sorting purposes
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareForSort(a, b Object) int {
	switch aObj := a.(type) {
	case *Int:
		bObj, ok := b.(*Int)
		if ok {
			switch {
			case aObj.Value < bObj.Value:
				return -1
			case aObj.Value > bObj.Value:
				return 1
			default:
				return 0
			}
		}
		bFloat, ok := b.(*Float)
		if ok {
			switch {
			case float64(aObj.Value) < bFloat.Value:
				return -1
			case float64(aObj.Value) > bFloat.Value:
				return 1
			default:
				return 0
			}
		}
	case *Float:
		bObj, ok := b.(*Float)
		if ok {
			switch {
			case aObj.Value < bObj.Value:
				return -1
			case aObj.Value > bObj.Value:
				return 1
			default:
				return 0
			}
		}
		bInt, ok := b.(*Int)
		if ok {
			switch {
			case aObj.Value < float64(bInt.Value):
				return -1
			case aObj.Value > float64(bInt.Value):
				return 1
			default:
				return 0
			}
		}
	case *String:
		bObj, ok := b.(*String)
		if ok {
			if aObj.Value < bObj.Value {
				return -1
			} else if aObj.Value > bObj.Value {
				return 1
			}
			return 0
		}
	case *Bool:
		bObj, ok := b.(*Bool)
		if ok {
			aVal := 0
			bVal := 0
			if aObj.Value {
				aVal = 1
			}
			if bObj.Value {
				bVal = 1
			}
			return aVal - bVal
		}
	}

	aStr := a.Inspect()
	bStr := b.Inspect()
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}
