// pkg/stdlib/sort_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callSortFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("sort")
	if mod == nil {
		panic("sort module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestSortNumbers(t *testing.T) {
	arr := Array(Int(3), Int(1), Int(2), Float(2.5))

	result := callSortFunc("numbers", arr)
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("numbers() should return Array, got %T", result)
	}
	if len(sorted.Elements) != 4 {
		t.Errorf("numbers() length = %d, want 4", len(sorted.Elements))
	}
	// First element should be 1
	if i, ok := sorted.Elements[0].(*objects.Int); !ok || i.Value != 1 {
		t.Errorf("numbers()[0] = %v, want 1", sorted.Elements[0])
	}
}

func TestSortNumbersDesc(t *testing.T) {
	arr := Array(Int(1), Int(3), Int(2))

	result := callSortFunc("numbersDesc", arr)
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("numbersDesc() should return Array, got %T", result)
	}
	// First element should be 3
	if i, ok := sorted.Elements[0].(*objects.Int); !ok || i.Value != 3 {
		t.Errorf("numbersDesc()[0] = %v, want 3", sorted.Elements[0])
	}
}

func TestSortStrings(t *testing.T) {
	arr := Array(String("c"), String("a"), String("b"))

	result := callSortFunc("strings", arr)
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("strings() should return Array, got %T", result)
	}
	// First element should be "a"
	if s, ok := sorted.Elements[0].(*objects.String); !ok || s.Value != "a" {
		t.Errorf("strings()[0] = %v, want 'a'", sorted.Elements[0])
	}
}

func TestSortStringsDesc(t *testing.T) {
	arr := Array(String("a"), String("c"), String("b"))

	result := callSortFunc("stringsDesc", arr)
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("stringsDesc() should return Array, got %T", result)
	}
	// First element should be "c"
	if s, ok := sorted.Elements[0].(*objects.String); !ok || s.Value != "c" {
		t.Errorf("stringsDesc()[0] = %v, want 'c'", sorted.Elements[0])
	}
}

func TestSortReverse(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))

	result := callSortFunc("reverse", arr)
	reversed, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("reverse() should return Array, got %T", result)
	}
	// First element should be 3
	if i, ok := reversed.Elements[0].(*objects.Int); !ok || i.Value != 3 {
		t.Errorf("reverse()[0] = %v, want 3", reversed.Elements[0])
	}
}

func TestSortIsSorted(t *testing.T) {
	// Sorted array
	sortedArr := Array(Int(1), Int(2), Int(3))
	result := callSortFunc("isSorted", sortedArr)
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isSorted() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isSorted([1,2,3]) should be true")
	}

	// Unsorted array
	unsortedArr := Array(Int(3), Int(1), Int(2))
	result = callSortFunc("isSorted", unsortedArr)
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isSorted() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isSorted([3,1,2]) should be false")
	}

	// Empty array
	emptyArr := Array()
	result = callSortFunc("isSorted", emptyArr)
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isSorted() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isSorted([]) should be true")
	}

	// Mixed types - tests string comparison fallback
	mixedArr := Array(String("a"), String("b"))
	result = callSortFunc("isSorted", mixedArr)
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isSorted() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isSorted(['a','b']) should be true")
	}
}

func TestSortMin(t *testing.T) {
	arr := Array(Int(3), Int(1), Int(2))

	result := callSortFunc("min", arr)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("min() should return Int, got %T", result)
	}
	if i.Value != 1 {
		t.Errorf("min([3,1,2]) = %d, want 1", i.Value)
	}

	// Empty array
	emptyArr := Array()
	result = callSortFunc("min", emptyArr)
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("min([]) should return Null, got %T", result)
	}

	// Float values
	floatArr := Array(Float(3.5), Float(1.2), Float(2.8))
	result = callSortFunc("min", floatArr)
	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("min() should return Float, got %T", result)
	}
	if f.Value != 1.2 {
		t.Errorf("min([3.5,1.2,2.8]) = %f, want 1.2", f.Value)
	}

	// Mixed numeric types
	mixedArr := Array(Int(3), Float(1.5), Int(2))
	result = callSortFunc("min", mixedArr)
	f, ok = result.(*objects.Float)
	if !ok {
		t.Fatalf("min() should return Float for mixed, got %T", result)
	}
	if f.Value != 1.5 {
		t.Errorf("min([3,1.5,2]) = %f, want 1.5", f.Value)
	}
}

func TestSortMax(t *testing.T) {
	arr := Array(Int(3), Int(1), Int(2))

	result := callSortFunc("max", arr)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("max() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("max([3,1,2]) = %d, want 3", i.Value)
	}

	// Empty array
	emptyArr := Array()
	result = callSortFunc("max", emptyArr)
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("max([]) should return Null, got %T", result)
	}

	// Float values
	floatArr := Array(Float(3.5), Float(1.2), Float(2.8))
	result = callSortFunc("max", floatArr)
	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("max() should return Float, got %T", result)
	}
	if f.Value != 3.5 {
		t.Errorf("max([3.5,1.2,2.8]) = %f, want 3.5", f.Value)
	}
}

func TestSortBy(t *testing.T) {
	arr := Array(
		Array(String("b"), Int(2)),
		Array(String("a"), Int(1)),
		Array(String("c"), Int(3)),
	)

	// Sort by second element
	keyFn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		if arr, ok := args[0].(*objects.Array); ok && len(arr.Elements) > 1 {
			return arr.Elements[1]
		}
		return Int(0)
	})

	result := callSortFunc("by", arr, keyFn)
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("by() should return Array, got %T", result)
	}
	// First element should have value 1
	if inner, ok := sorted.Elements[0].(*objects.Array); ok {
		if i, ok := inner.Elements[1].(*objects.Int); !ok || i.Value != 1 {
			t.Errorf("by()[0][1] = %v, want 1", inner.Elements[1])
		}
	} else {
		t.Errorf("by()[0] should be Array, got %T", sorted.Elements[0])
	}
}
