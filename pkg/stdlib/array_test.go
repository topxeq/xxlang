// pkg/stdlib/array_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callArrayFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/array")
	if mod == nil {
		panic("std/array module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestArrayLen(t *testing.T) {
	tests := []struct {
		size int
		want int64
	}{
		{0, 0},
		{3, 3},
		{5, 5},
	}

	for _, tt := range tests {
		elems := make([]objects.Object, tt.size)
		for i := 0; i < tt.size; i++ {
			elems[i] = Int(int64(i))
		}
		arr := Array(elems...)
		result := callArrayFunc("len", arr)
		r, ok := result.(*objects.Int)
		if !ok {
			t.Errorf("len(%v) = %v, want Int", arr, result)
			continue
		}
		if r.Value != tt.want {
			t.Errorf("len(%v) = %d, want %d", arr, r.Value, tt.want)
		}
	}
}

func TestArrayPush(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("push", arr, Int(4))
	r, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("push should return array")
	}
	if len(r.Elements) != 4 {
		t.Errorf("push should add 4 elements, got %d", len(r.Elements))
	}
}

func TestArrayPop(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("pop", arr)
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("pop should return int")
	}
	if r.Value != 3 {
		t.Errorf("pop should return 3, got %d", r.Value)
	}
}

func TestArrayFirst(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("first", arr)
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("first should return int")
	}
	if r.Value != 1 {
		t.Errorf("first should return 1, got %d", r.Value)
	}
}

func TestArrayLast(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("last", arr)
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("last should return int")
	}
	if r.Value != 3 {
		t.Errorf("last should return 3, got %d", r.Value)
	}
}

func TestArrayContains(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("contains", arr, Int(2))
	r, ok := result.(*objects.Bool)
	if !ok {
		t.Fatal("contains should return bool")
	}
	if !r.Value {
		t.Errorf("contains([1,2,3], 2) should be true, got %v", r.Value)
	}

	result = callArrayFunc("contains", arr, Int(4))
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatal("contains should return bool")
	}
	if r.Value {
		t.Errorf("contains([1,2,3], 4) should be false, got %v", r.Value)
	}
}

func TestArrayIndexOf(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("indexOf", arr, Int(2))
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("indexOf should return int")
	}
	if r.Value != 1 {
		t.Errorf("indexOf([1,2,3], 2) should return 1, got %d", r.Value)
	}

	result = callArrayFunc("indexOf", arr, Int(4))
	r, ok = result.(*objects.Int)
	if !ok {
		t.Fatal("indexOf should return int")
	}
	if r.Value != -1 {
		t.Errorf("indexOf([1,2,3], 4) should return -1, got %d", r.Value)
	}
}

func TestArrayReverse(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("reverse", arr)
	r, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("reverse should return array")
	}
	if len(r.Elements) != 3 {
		t.Errorf("reverse should return 3 elements, got %d", len(r.Elements))
	}
	if r.Elements[0].(*objects.Int).Value != 3 || r.Elements[1].(*objects.Int).Value != 2 || r.Elements[2].(*objects.Int).Value != 1 {
		t.Errorf("reverse([1,2,3]) = %v, want [3,2,1]", r.Elements)
	}
}

func TestArraySlice(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))
	result := callArrayFunc("slice", arr, Int(1), Int(4))
	r, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("slice should return array")
	}
	if len(r.Elements) != 3 {
		t.Errorf("slice([1,2,3,4,5], 1, 4) should return 3 elements, got %d", len(r.Elements))
	}
}

func TestArrayIsEmpty(t *testing.T) {
	emptyArr := Array()
	result := callArrayFunc("isEmpty", emptyArr)
	r, ok := result.(*objects.Bool)
	if !ok {
		t.Fatal("isEmpty should return bool")
	}
	if !r.Value {
		t.Errorf("isEmpty([]) should return true, got %v", r.Value)
	}

	nonEmptyArr := Array(Int(1))
	result = callArrayFunc("isEmpty", nonEmptyArr)
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatal("isEmpty should return bool")
	}
	if r.Value {
		t.Errorf("isEmpty([1]) should return false, got %v", r.Value)
	}
}

func TestArrayEmptyArray(t *testing.T) {
	emptyArr := Array()

	// len of empty array
	result := callArrayFunc("len", emptyArr)
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 0 {
		t.Errorf("len([]) = %v, want 0", result)
	}

	// pop on empty array returns null
	result = callArrayFunc("pop", Array())
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("pop([]) should return null, got %T", result)
	}

	// shift on empty array returns null
	result = callArrayFunc("shift", Array())
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("shift([]) should return null, got %T", result)
	}

	// first on empty array returns null
	result = callArrayFunc("first", Array())
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("first([]) should return null, got %T", result)
	}

	// last on empty array returns null
	result = callArrayFunc("last", Array())
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("last([]) should return null, got %T", result)
	}

	// reverse empty array returns empty array
	result = callArrayFunc("reverse", Array())
	arr, ok := result.(*objects.Array)
	if !ok || len(arr.Elements) != 0 {
		t.Errorf("reverse([]) should return empty array, got %v", result)
	}

	// contains on empty array returns false
	result = callArrayFunc("contains", Array(), Int(1))
	r2, ok := result.(*objects.Bool)
	if !ok || r2.Value {
		t.Errorf("contains([], 1) should return false, got %v", result)
	}

	// indexOf on empty array returns -1
	result = callArrayFunc("indexOf", Array(), Int(1))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != -1 {
		t.Errorf("indexOf([], 1) = %v, want -1", result)
	}
}

func TestArrayReverseEdgeCases(t *testing.T) {
	// Reverse single element array
	arr := Array(Int(1))
	result := callArrayFunc("reverse", arr)
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 1 || r.Elements[0].(*objects.Int).Value != 1 {
		t.Errorf("reverse([1]) = %v, want [1]", result)
	}

	// Reverse two element array
	arr = Array(Int(1), Int(2))
	result = callArrayFunc("reverse", arr)
	r, ok = result.(*objects.Array)
	if !ok || r.Elements[0].(*objects.Int).Value != 2 || r.Elements[1].(*objects.Int).Value != 1 {
		t.Errorf("reverse([1, 2]) = %v, want [2, 1]", r.Elements)
	}

	// Reverse error - non-array
	result = callArrayFunc("reverse", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reverse with non-array should return error, got %T", result)
	}

	// Reverse error - wrong args
	result = callArrayFunc("reverse")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reverse with 0 args should return error, got %T", result)
	}
}

func TestArraySort(t *testing.T) {
	// Sort empty array
	result := callArrayFunc("sort", Array())
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("sort([]) should return empty array, got %v", result)
	}

	// Sort single element
	result = callArrayFunc("sort", Array(Int(1)))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 1 {
		t.Errorf("sort([1]) should return array with 1 element, got %v", result)
	}

	// Sort error - non-array
	result = callArrayFunc("sort", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("sort with non-array should return error, got %T", result)
	}

	// Sort error - no args
	result = callArrayFunc("sort")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("sort with 0 args should return error, got %T", result)
	}
}

func TestArrayPushEdgeCases(t *testing.T) {
	// Push multiple elements
	arr := Array(Int(1))
	result := callArrayFunc("push", arr, Int(2), Int(3), Int(4))
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 4 {
		t.Errorf("push should add 3 elements to array, got %d", len(r.Elements))
	}

	// Push to empty array
	emptyArr := Array()
	result = callArrayFunc("push", emptyArr, Int(1))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 1 {
		t.Errorf("push to empty array should result in 1 element, got %d", len(r.Elements))
	}

	// Push error - no args
	result = callArrayFunc("push")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("push with 0 args should return error, got %T", result)
	}

	// Push error - non-array first arg
	result = callArrayFunc("push", String("hello"), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("push with non-array should return error, got %T", result)
	}
}

func TestArrayPopEdgeCases(t *testing.T) {
	// Pop error - non-array
	result := callArrayFunc("pop", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("pop with non-array should return error, got %T", result)
	}

	// Pop error - wrong args
	result = callArrayFunc("pop")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("pop with 0 args should return error, got %T", result)
	}
}

func TestArrayShift(t *testing.T) {
	// Basic shift
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("shift", arr)
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 1 {
		t.Errorf("shift should return 1, got %v", result)
	}

	// Verify array was modified
	if len(arr.Elements) != 2 {
		t.Errorf("shift should modify array, got %d elements", len(arr.Elements))
	}

	// Shift error - non-array
	result = callArrayFunc("shift", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("shift with non-array should return error, got %T", result)
	}

	// Shift error - wrong args
	result = callArrayFunc("shift")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("shift with 0 args should return error, got %T", result)
	}
}

func TestArrayUnshift(t *testing.T) {
	// Basic unshift
	arr := Array(Int(2), Int(3))
	result := callArrayFunc("unshift", arr, Int(1))
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("unshift should add element, got %d", len(r.Elements))
	}

	// Verify element was added at front
	if r.Elements[0].(*objects.Int).Value != 1 {
		t.Errorf("unshift should add at front, got %v", r.Elements[0])
	}

	// Unshift multiple elements
	arr2 := Array(Int(3))
	result = callArrayFunc("unshift", arr2, Int(1), Int(2))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("unshift multiple should add 2 elements, got %d", len(r.Elements))
	}

	// Unshift error - no args
	result = callArrayFunc("unshift")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unshift with 0 args should return error, got %T", result)
	}

	// Unshift error - non-array
	result = callArrayFunc("unshift", String("hello"), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unshift with non-array should return error, got %T", result)
	}
}

func TestArrayGet(t *testing.T) {
	// Basic get
	arr := Array(Int(10), Int(20), Int(30))
	result := callArrayFunc("get", arr, Int(1))
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 20 {
		t.Errorf("get([10,20,30], 1) = %v, want 20", result)
	}

	// Get with negative index returns null
	result = callArrayFunc("get", arr, Int(-1))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("get with negative index should return null, got %T", result)
	}

	// Get with out of range index returns null
	result = callArrayFunc("get", arr, Int(10))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("get with out of range index should return null, got %T", result)
	}

	// Get error - wrong args
	result = callArrayFunc("get", arr)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("get with 1 arg should return error, got %T", result)
	}

	// Get error - non-array
	result = callArrayFunc("get", String("hello"), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("get with non-array should return error, got %T", result)
	}

	// Get error - non-int index
	result = callArrayFunc("get", arr, String("0"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("get with non-int index should return error, got %T", result)
	}
}

func TestArraySet(t *testing.T) {
	// Basic set
	arr := Array(Int(10), Int(20), Int(30))
	result := callArrayFunc("set", arr, Int(1), Int(25))
	r, ok := result.(*objects.Array)
	if !ok {
		t.Errorf("set should return array, got %T", result)
	}
	if r.Elements[1].(*objects.Int).Value != 25 {
		t.Errorf("set should modify array, got %v", r.Elements[1])
	}

	// Set error - out of range
	result = callArrayFunc("set", arr, Int(10), Int(100))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("set with out of range index should return error, got %T", result)
	}

	// Set error - negative index
	result = callArrayFunc("set", arr, Int(-1), Int(100))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("set with negative index should return error, got %T", result)
	}

	// Set error - wrong args
	result = callArrayFunc("set", arr, Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("set with 2 args should return error, got %T", result)
	}

	// Set error - non-array
	result = callArrayFunc("set", String("hello"), Int(0), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("set with non-array should return error, got %T", result)
	}
}

func TestArraySliceEdgeCases(t *testing.T) {
	// Slice empty array
	result := callArrayFunc("slice", Array(), Int(0))
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("slice([], 0) should return empty array, got %v", result)
	}

	// Slice with start > end returns empty array
	result = callArrayFunc("slice", Array(Int(1), Int(2), Int(3)), Int(2), Int(1))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("slice with start > end should return empty array, got %v", result)
	}

	// Slice with negative start (clamped to 0)
	result = callArrayFunc("slice", Array(Int(1), Int(2), Int(3)), Int(-5), Int(2))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 2 {
		t.Errorf("slice with negative start should be clamped, got %d elements", len(r.Elements))
	}

	// Slice with end > length (clamped)
	result = callArrayFunc("slice", Array(Int(1), Int(2), Int(3)), Int(0), Int(100))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("slice with end > length should be clamped, got %d elements", len(r.Elements))
	}

	// Slice error - no args
	result = callArrayFunc("slice")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("slice with 0 args should return error, got %T", result)
	}

	// Slice error - non-array
	result = callArrayFunc("slice", String("hello"), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("slice with non-array should return error, got %T", result)
	}

	// Slice error - non-int index
	result = callArrayFunc("slice", Array(Int(1)), String("0"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("slice with non-int index should return error, got %T", result)
	}
}

func TestArrayConcat(t *testing.T) {
	// Basic concat
	arr1 := Array(Int(1), Int(2))
	arr2 := Array(Int(3), Int(4))
	result := callArrayFunc("concat", arr1, arr2)
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 4 {
		t.Errorf("concat should return 4 elements, got %d", len(r.Elements))
	}

	// Concat empty arrays
	result = callArrayFunc("concat", Array(), Array())
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("concat empty arrays should return empty, got %d", len(r.Elements))
	}

	// Concat error - non-array
	result = callArrayFunc("concat", Array(), String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("concat with non-array should return error, got %T", result)
	}

	// Concat error - not enough args
	result = callArrayFunc("concat", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("concat with 1 arg should return error, got %T", result)
	}
}

func TestArrayJoin(t *testing.T) {
	// Basic join
	arr := Array(Int(1), Int(2), Int(3))
	result := callArrayFunc("join", arr, String("-"))
	r, ok := result.(*objects.String)
	if !ok || r.Value != "1-2-3" {
		t.Errorf("join([1,2,3], \"-\") = %v, want \"1-2-3\"", result)
	}

	// Join empty array
	result = callArrayFunc("join", Array(), String(","))
	r, ok = result.(*objects.String)
	if !ok || r.Value != "" {
		t.Errorf("join([], \",\") = %v, want \"\"", result)
	}

	// Join with string elements
	arr2 := Array(String("a"), String("b"))
	result = callArrayFunc("join", arr2, String(","))
	r, ok = result.(*objects.String)
	if !ok || r.Value != "a,b" {
		t.Errorf("join with strings = %v, want \"a,b\"", result)
	}

	// Join error - non-array
	result = callArrayFunc("join", String("hello"), String(","))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("join with non-array should return error, got %T", result)
	}

	// Join error - non-string separator
	result = callArrayFunc("join", Array(), Int(44))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("join with non-string separator should return error, got %T", result)
	}
}

func TestArrayFlatten(t *testing.T) {
	// Basic flatten
	arr := Array(Array(Int(1), Int(2)), Int(3))
	result := callArrayFunc("flatten", arr)
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("flatten should return 3 elements, got %d", len(r.Elements))
	}

	// Flatten empty array
	result = callArrayFunc("flatten", Array())
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("flatten([]) should return empty, got %d", len(r.Elements))
	}

	// Flatten array with no nested arrays
	result = callArrayFunc("flatten", Array(Int(1), Int(2)))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 2 {
		t.Errorf("flatten with no nesting = %d elements, want 2", len(r.Elements))
	}

	// Flatten error - non-array
	result = callArrayFunc("flatten", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("flatten with non-array should return error, got %T", result)
	}

	// Flatten error - wrong args
	result = callArrayFunc("flatten")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("flatten with 0 args should return error, got %T", result)
	}
}

func TestArrayUnique(t *testing.T) {
	// Basic unique
	arr := Array(Int(1), Int(2), Int(1), Int(3), Int(2))
	result := callArrayFunc("unique", arr)
	r, ok := result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("unique should return 3 elements, got %d", len(r.Elements))
	}

	// Unique empty array
	result = callArrayFunc("unique", Array())
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 0 {
		t.Errorf("unique([]) should return empty, got %d", len(r.Elements))
	}

	// Unique with all different
	result = callArrayFunc("unique", Array(Int(1), Int(2), Int(3)))
	r, ok = result.(*objects.Array)
	if !ok || len(r.Elements) != 3 {
		t.Errorf("unique with all different should return 3, got %d", len(r.Elements))
	}

	// Unique error - non-array
	result = callArrayFunc("unique", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unique with non-array should return error, got %T", result)
	}

	// Unique error - wrong args
	result = callArrayFunc("unique")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unique with 0 args should return error, got %T", result)
	}
}

func TestArrayFill(t *testing.T) {
	// Basic fill
	arr := Array(Int(1), Int(2), Int(3), Int(4))
	result := callArrayFunc("fill", arr, Int(0), Int(1), Int(3))
	r, ok := result.(*objects.Array)
	if !ok {
		t.Errorf("fill should return array, got %T", result)
	}
	if r.Elements[0].(*objects.Int).Value != 1 || r.Elements[1].(*objects.Int).Value != 0 || r.Elements[2].(*objects.Int).Value != 0 {
		t.Errorf("fill should modify specified range, got %v", r.Elements)
	}

	// Fill entire array
	arr2 := Array(Int(1), Int(2), Int(3))
	result = callArrayFunc("fill", arr2, Int(9))
	r, ok = result.(*objects.Array)
	if !ok {
		t.Errorf("fill should return array, got %T", result)
	}
	for i, e := range r.Elements {
		if e.(*objects.Int).Value != 9 {
			t.Errorf("fill entire array should set all to 9, element %d = %v", i, e)
		}
	}

	// Fill error - non-array
	result = callArrayFunc("fill", String("hello"), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("fill with non-array should return error, got %T", result)
	}

	// Fill error - not enough args
	result = callArrayFunc("fill", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("fill with 1 arg should return error, got %T", result)
	}
}

func TestArrayMap(t *testing.T) {
	// Map error - non-array
	result := callArrayFunc("map", String("hello"), nil)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("map with non-array should return error, got %T", result)
	}

	// Map error - wrong args
	result = callArrayFunc("map", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("map with 1 arg should return error, got %T", result)
	}
}

func TestArrayFilter(t *testing.T) {
	// Filter error - non-array
	result := callArrayFunc("filter", String("hello"), nil)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("filter with non-array should return error, got %T", result)
	}

	// Filter error - wrong args
	result = callArrayFunc("filter", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("filter with 1 arg should return error, got %T", result)
	}
}

func TestArrayReduce(t *testing.T) {
	// Reduce error - non-array
	result := callArrayFunc("reduce", String("hello"), nil, nil)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reduce with non-array should return error, got %T", result)
	}

	// Reduce error - not enough args
	result = callArrayFunc("reduce", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reduce with 1 arg should return error, got %T", result)
	}

	result = callArrayFunc("reduce", Array(), nil)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reduce with 2 args should return error, got %T", result)
	}
}

func TestArrayFirstLastEdgeCases(t *testing.T) {
	// first error - non-array
	result := callArrayFunc("first", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("first with non-array should return error, got %T", result)
	}

	// first error - wrong args
	result = callArrayFunc("first")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("first with 0 args should return error, got %T", result)
	}

	// last error - non-array
	result = callArrayFunc("last", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("last with non-array should return error, got %T", result)
	}

	// last error - wrong args
	result = callArrayFunc("last")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("last with 0 args should return error, got %T", result)
	}
}

func TestArrayContainsIndexOfEdgeCases(t *testing.T) {
	// contains error - non-array
	result := callArrayFunc("contains", String("hello"), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("contains with non-array should return error, got %T", result)
	}

	// contains error - wrong args
	result = callArrayFunc("contains", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("contains with 1 arg should return error, got %T", result)
	}

	// indexOf error - non-array
	result = callArrayFunc("indexOf", String("hello"), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("indexOf with non-array should return error, got %T", result)
	}

	// indexOf error - wrong args
	result = callArrayFunc("indexOf", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("indexOf with 1 arg should return error, got %T", result)
	}
}

func TestArrayIsEmptyEdgeCases(t *testing.T) {
	// isEmpty error - non-array
	result := callArrayFunc("isEmpty", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("isEmpty with non-array should return error, got %T", result)
	}

	// isEmpty error - wrong args
	result = callArrayFunc("isEmpty")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("isEmpty with 0 args should return error, got %T", result)
	}
}

func TestArrayLenEdgeCases(t *testing.T) {
	// len error - non-array
	result := callArrayFunc("len", String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("len with non-array should return error, got %T", result)
	}

	// len error - wrong args
	result = callArrayFunc("len")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("len with 0 args should return error, got %T", result)
	}
}
