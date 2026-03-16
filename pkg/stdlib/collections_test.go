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
