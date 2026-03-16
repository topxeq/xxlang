// pkg/stdlib/collections_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callCollectionsFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("collections")
	if mod == nil {
		panic("collections module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestCollectionsUnion(t *testing.T) {
	arr1 := Array(Int(1), Int(2), Int(3))
	arr2 := Array(Int(3), Int(4), Int(5))

	result := callCollectionsFunc("union", arr1, arr2)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("union() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("union() length = %d, want 5", len(arr.Elements))
	}
}

func TestCollectionsIntersection(t *testing.T) {
	arr1 := Array(Int(1), Int(2), Int(3))
	arr2 := Array(Int(2), Int(3), Int(4))

	result := callCollectionsFunc("intersection", arr1, arr2)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("intersection() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("intersection() length = %d, want 2", len(arr.Elements))
	}
}

func TestCollectionsDifference(t *testing.T) {
	arr1 := Array(Int(1), Int(2), Int(3))
	arr2 := Array(Int(2), Int(3), Int(4))

	result := callCollectionsFunc("difference", arr1, arr2)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("difference() should return Array, got %T", result)
	}
	if len(arr.Elements) != 1 {
		t.Errorf("difference() length = %d, want 1", len(arr.Elements))
	}
}

func TestCollectionsSample(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	// Test single sample
	result := callCollectionsFunc("sample", arr)
	_, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("sample() should return Int, got %T", result)
	}

	// Test multiple samples
	result = callCollectionsFunc("sample", arr, Int(3))
	sampleArr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("sample() with count should return Array, got %T", result)
	}
	if len(sampleArr.Elements) != 3 {
		t.Errorf("sample() with count=3 length = %d, want 3", len(sampleArr.Elements))
	}

	// Test empty array
	emptyArr := Array()
	result = callCollectionsFunc("sample", emptyArr)
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("sample() on empty array should return Null, got %T", result)
	}
}

func TestCollectionsShuffle(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	result := callCollectionsFunc("shuffle", arr)
	shuffled, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("shuffle() should return Array, got %T", result)
	}
	if len(shuffled.Elements) != 5 {
		t.Errorf("shuffle() length = %d, want 5", len(shuffled.Elements))
	}
}

func TestCollectionsChunk(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	result := callCollectionsFunc("chunk", arr, Int(2))
	chunks, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("chunk() should return Array, got %T", result)
	}
	if len(chunks.Elements) != 3 {
		t.Errorf("chunk() length = %d, want 3", len(chunks.Elements))
	}
}

func TestCollectionsFlattenDeep(t *testing.T) {
	arr := Array(
		Int(1),
		Array(Int(2), Int(3)),
		Int(4),
	)

	result := callCollectionsFunc("flattenDeep", arr)
	flat, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("flattenDeep() should return Array, got %T", result)
	}
	if len(flat.Elements) != 4 {
		t.Errorf("flattenDeep() length = %d, want 4", len(flat.Elements))
	}
}

func TestCollectionsGroupBy(t *testing.T) {
	arr := Array(
		Array(String("a"), Int(1)),
		Array(String("b"), Int(2)),
		Array(String("a"), Int(3)),
	)

	// Group by first element
	keyFn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		if innerArr, ok := args[0].(*objects.Array); ok && len(innerArr.Elements) > 0 {
			return innerArr.Elements[0]
		}
		return String("")
	})

	result := callCollectionsFunc("groupBy", arr, keyFn)
	groups, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("groupBy() should return Array, got %T", result)
	}
	if len(groups.Elements) < 1 {
		t.Error("groupBy() should return at least one group")
	}
}

func TestCollectionsCountBy(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(2), Int(3), Int(3), Int(3))

	result := callCollectionsFunc("countBy", arr)
	counts, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("countBy() should return Array, got %T", result)
	}
	if len(counts.Elements) < 1 {
		t.Error("countBy() should return at least one count pair")
	}
}

func TestCollectionsPartition(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	// Partition by even/odd
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		if i, ok := args[0].(*objects.Int); ok {
			return Bool(i.Value%2 == 0)
		}
		return Bool(false)
	})

	result := callCollectionsFunc("partition", arr, pred)
	parts, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("partition() should return Array, got %T", result)
	}
	if len(parts.Elements) != 2 {
		t.Errorf("partition() should return 2 arrays, got %d", len(parts.Elements))
	}
}

// ============================================
// Error handling tests for collections
// ============================================

func TestCollectionsUnionErrors(t *testing.T) {
	// Wrong number of args
	result := callCollectionsFunc("union", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("union() with 1 arg should return Error")
	}

	// Wrong type
	result = callCollectionsFunc("union", Int(1), Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("union() with non-array should return Error")
	}

	result = callCollectionsFunc("union", Array(), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("union() with non-array second arg should return Error")
	}
}

func TestCollectionsIntersectionErrors(t *testing.T) {
	result := callCollectionsFunc("intersection", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("intersection() with 1 arg should return Error")
	}

	result = callCollectionsFunc("intersection", Int(1), Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("intersection() with non-array should return Error")
	}
}

func TestCollectionsDifferenceErrors(t *testing.T) {
	result := callCollectionsFunc("difference", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("difference() with 1 arg should return Error")
	}

	result = callCollectionsFunc("difference", Int(1), Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("difference() with non-array should return Error")
	}
}

func TestCollectionsChunkErrors(t *testing.T) {
	// Wrong number of args
	result := callCollectionsFunc("chunk", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("chunk() with 1 arg should return Error")
	}

	// Wrong type for array
	result = callCollectionsFunc("chunk", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("chunk() with non-array should return Error")
	}

	// Wrong type for size
	result = callCollectionsFunc("chunk", Array(), String("x"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("chunk() with non-int size should return Error")
	}

	// Zero or negative size
	result = callCollectionsFunc("chunk", Array(), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("chunk() with size 0 should return Error")
	}

	result = callCollectionsFunc("chunk", Array(), Int(-1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("chunk() with negative size should return Error")
	}
}

func TestCollectionsZip(t *testing.T) {
	arr1 := Array(Int(1), Int(2), Int(3))
	arr2 := Array(String("a"), String("b"), String("c"))

	result := callCollectionsFunc("zip", arr1, arr2)
	zip, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("zip() should return Array, got %T", result)
	}
	if len(zip.Elements) != 3 {
		t.Errorf("zip() length = %d, want 3", len(zip.Elements))
	}
}

func TestCollectionsZipErrors(t *testing.T) {
	// Too few args
	result := callCollectionsFunc("zip", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("zip() with 1 arg should return Error")
	}

	// Wrong type
	result = callCollectionsFunc("zip", Int(1), Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("zip() with non-array should return Error")
	}
}

func TestCollectionsFlattenDeepErrors(t *testing.T) {
	result := callCollectionsFunc("flattenDeep")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("flattenDeep() with no args should return Error")
	}

	result = callCollectionsFunc("flattenDeep", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("flattenDeep() with non-array should return Error")
	}
}

func TestCollectionsCountByErrors(t *testing.T) {
	result := callCollectionsFunc("countBy")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("countBy() with no args should return Error")
	}

	result = callCollectionsFunc("countBy", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("countBy() with non-array should return Error")
	}
}

func TestCollectionsGroupByErrors(t *testing.T) {
	// Wrong number of args
	result := callCollectionsFunc("groupBy", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("groupBy() with 1 arg should return Error")
	}

	// Wrong type for array
	result = callCollectionsFunc("groupBy", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("groupBy() with non-array should return Error")
	}

	// Wrong type for function
	result = callCollectionsFunc("groupBy", Array(), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("groupBy() with non-function should return Error")
	}
}

func TestCollectionsPartitionErrors(t *testing.T) {
	result := callCollectionsFunc("partition", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("partition() with 1 arg should return Error")
	}

	result = callCollectionsFunc("partition", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("partition() with non-array should return Error")
	}

	result = callCollectionsFunc("partition", Array(), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("partition() with non-function should return Error")
	}
}

// ============================================
// Tests for take, takeWhile, drop
// ============================================

func TestCollectionsTake(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	result := callCollectionsFunc("take", arr, Int(3))
	taken, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("take() should return Array, got %T", result)
	}
	if len(taken.Elements) != 3 {
		t.Errorf("take() length = %d, want 3", len(taken.Elements))
	}

	// Take more than available
	result = callCollectionsFunc("take", Array(Int(1), Int(2)), Int(10))
	taken, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("take() should return Array, got %T", result)
	}
	if len(taken.Elements) != 2 {
		t.Errorf("take() length = %d, want 2", len(taken.Elements))
	}
}

func TestCollectionsTakeErrors(t *testing.T) {
	result := callCollectionsFunc("take", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("take() with 1 arg should return Error")
	}

	result = callCollectionsFunc("take", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("take() with non-array should return Error")
	}

	result = callCollectionsFunc("take", Array(), String("x"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("take() with non-int should return Error")
	}
}

func TestCollectionsTakeWhile(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value < 4)
	})

	result := callCollectionsFunc("takeWhile", Array(Int(1), Int(2), Int(3), Int(5), Int(6)), pred)
	taken, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("takeWhile() should return Array, got %T", result)
	}
	if len(taken.Elements) != 3 {
		t.Errorf("takeWhile() length = %d, want 3", len(taken.Elements))
	}
}

func TestCollectionsTakeWhileErrors(t *testing.T) {
	result := callCollectionsFunc("takeWhile", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("takeWhile() with 1 arg should return Error")
	}

	result = callCollectionsFunc("takeWhile", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("takeWhile() with non-array should return Error")
	}

	result = callCollectionsFunc("takeWhile", Array(), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("takeWhile() with non-function should return Error")
	}
}

func TestCollectionsDrop(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))

	result := callCollectionsFunc("drop", arr, Int(2))
	dropped, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("drop() should return Array, got %T", result)
	}
	if len(dropped.Elements) != 3 {
		t.Errorf("drop() length = %d, want 3", len(dropped.Elements))
	}

	// Drop more than available
	result = callCollectionsFunc("drop", Array(Int(1), Int(2)), Int(10))
	dropped, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("drop() should return Array, got %T", result)
	}
	if len(dropped.Elements) != 0 {
		t.Errorf("drop() length = %d, want 0", len(dropped.Elements))
	}

	// Drop negative (should return all)
	result = callCollectionsFunc("drop", arr, Int(-1))
	dropped, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("drop() should return Array, got %T", result)
	}
	if len(dropped.Elements) != 5 {
		t.Errorf("drop() with negative should return all elements")
	}
}

func TestCollectionsDropErrors(t *testing.T) {
	result := callCollectionsFunc("drop", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("drop() with 1 arg should return Error")
	}

	result = callCollectionsFunc("drop", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("drop() with non-array should return Error")
	}

	result = callCollectionsFunc("drop", Array(), String("x"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("drop() with non-int should return Error")
	}
}

// ============================================
// Tests for find, findIndex
// ============================================

func TestCollectionsFind(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 3)
	})

	result := callCollectionsFunc("find", Array(Int(1), Int(2), Int(4), Int(5)), pred)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("find() should return Int, got %T", result)
	}
	if i.Value != 4 {
		t.Errorf("find() = %d, want 4", i.Value)
	}

	// Not found
	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 100)
	})
	result = callCollectionsFunc("find", Array(Int(1), Int(2), Int(3)), pred2)
	if result != objects.NULL {
		t.Errorf("find() not found should return null, got %T", result)
	}
}

func TestCollectionsFindErrors(t *testing.T) {
	result := callCollectionsFunc("find", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("find() with 1 arg should return Error")
	}

	result = callCollectionsFunc("find", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("find() with non-array should return Error")
	}

	result = callCollectionsFunc("find", Array(), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("find() with non-function should return Error")
	}
}

func TestCollectionsFindIndex(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 3)
	})

	result := callCollectionsFunc("findIndex", Array(Int(1), Int(2), Int(4), Int(5)), pred)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("findIndex() should return Int, got %T", result)
	}
	if i.Value != 2 {
		t.Errorf("findIndex() = %d, want 2", i.Value)
	}

	// Not found
	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 100)
	})
	result = callCollectionsFunc("findIndex", Array(Int(1), Int(2), Int(3)), pred2)
	i, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("findIndex() should return Int, got %T", result)
	}
	if i.Value != -1 {
		t.Errorf("findIndex() not found = %d, want -1", i.Value)
	}
}

func TestCollectionsFindIndexErrors(t *testing.T) {
	result := callCollectionsFunc("findIndex", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("findIndex() with 1 arg should return Error")
	}

	result = callCollectionsFunc("findIndex", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("findIndex() with non-array should return Error")
	}
}

// ============================================
// Tests for every, some
// ============================================

func TestCollectionsEvery(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 0)
	})

	result := callCollectionsFunc("every", Array(Int(1), Int(2), Int(3)), pred)
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("every() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("every() should return true when all match")
	}

	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 5)
	})
	result = callCollectionsFunc("every", Array(Int(1), Int(2), Int(3)), pred2)
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("every() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("every() should return false when not all match")
	}
}

func TestCollectionsEveryErrors(t *testing.T) {
	result := callCollectionsFunc("every", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("every() with 1 arg should return Error")
	}

	result = callCollectionsFunc("every", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("every() with non-array should return Error")
	}

	result = callCollectionsFunc("every", Array(), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("every() with non-function should return Error")
	}
}

func TestCollectionsSome(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 5)
	})

	result := callCollectionsFunc("some", Array(Int(1), Int(2), Int(10)), pred)
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("some() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("some() should return true when any match")
	}

	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 100)
	})
	result = callCollectionsFunc("some", Array(Int(1), Int(2), Int(3)), pred2)
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("some() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("some() should return false when none match")
	}
}

func TestCollectionsSomeErrors(t *testing.T) {
	result := callCollectionsFunc("some", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("some() with 1 arg should return Error")
	}

	result = callCollectionsFunc("some", Int(1), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("some() with non-array should return Error")
	}
}

// ============================================
// Tests for rangeStep, repeat
// ============================================

func TestCollectionsRangeStep(t *testing.T) {
	// Positive step
	result := callCollectionsFunc("rangeStep", Int(0), Int(10), Int(2))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("rangeStep() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("rangeStep(0, 10, 2) length = %d, want 5", len(arr.Elements))
	}

	// Negative step
	result = callCollectionsFunc("rangeStep", Int(10), Int(0), Int(-2))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("rangeStep() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("rangeStep(10, 0, -2) length = %d, want 5", len(arr.Elements))
	}
}

func TestCollectionsRangeStepErrors(t *testing.T) {
	// Too few args
	result := callCollectionsFunc("rangeStep", Int(0), Int(10))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("rangeStep() with 2 args should return Error")
	}

	// Wrong types
	result = callCollectionsFunc("rangeStep", String("a"), Int(10), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("rangeStep() with non-int should return Error")
	}

	// Zero step
	result = callCollectionsFunc("rangeStep", Int(0), Int(10), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("rangeStep() with step 0 should return Error")
	}

	result = callCollectionsFunc("rangeStep", Int(0), Int(10), String("x"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("rangeStep() with non-int step should return Error")
	}
}

func TestCollectionsRepeat(t *testing.T) {
	result := callCollectionsFunc("repeat", Int(42), Int(5))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("repeat() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("repeat() length = %d, want 5", len(arr.Elements))
	}

	// Zero count
	result = callCollectionsFunc("repeat", Int(42), Int(0))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("repeat() should return Array, got %T", result)
	}
	if len(arr.Elements) != 0 {
		t.Errorf("repeat() with 0 count should be empty")
	}
}

func TestCollectionsRepeatErrors(t *testing.T) {
	result := callCollectionsFunc("repeat", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("repeat() with 1 arg should return Error")
	}

	result = callCollectionsFunc("repeat", Int(42), Int(-1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("repeat() with negative count should return Error")
	}

	result = callCollectionsFunc("repeat", Int(42), String("x"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("repeat() with non-int count should return Error")
	}
}

func TestCollectionsShuffleErrors(t *testing.T) {
	result := callCollectionsFunc("shuffle")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("shuffle() with no args should return Error")
	}

	result = callCollectionsFunc("shuffle", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("shuffle() with non-array should return Error")
	}
}

func TestCollectionsSampleErrors(t *testing.T) {
	result := callCollectionsFunc("sample")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sample() with no args should return Error")
	}

	result = callCollectionsFunc("sample", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sample() with non-array should return Error")
	}
}
