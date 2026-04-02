// pkg/objects/builtin_collection_test.go
package objects

import (
	"fmt"
	"testing"
)

func TestBuiltinMapArray(t *testing.T) {
	fn, ok := Builtins["mapArray"]
	if !ok {
		t.Fatal("mapArray builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		if i, ok := args[0].(*Int); ok {
			return NewInt(i.Value * 2), nil
		}
		return args[0], nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(&String{Value: "not array"})
	if !isError(result) {
		t.Error("expected error for non-array arg")
	}
}

func TestBuiltinFilterArray(t *testing.T) {
	fn, ok := Builtins["filterArray"]
	if !ok {
		t.Fatal("filterArray builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		if i, ok := args[0].(*Int); ok {
			return &Bool{Value: i.Value > 1}, nil
		}
		return FALSE, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinReduceArray(t *testing.T) {
	fn, ok := Builtins["reduceArray"]
	if !ok {
		t.Fatal("reduceArray builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		a, _ := args[0].(*Int)
		b, _ := args[1].(*Int)
		return NewInt(a.Value + b.Value), nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 6 {
		t.Errorf("expected 6, got %d", intResult.Value)
	}

	result = fn.Fn(arr, &Builtin{}, NewInt(10))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 16 {
		t.Errorf("expected 16, got %d", intResult.Value)
	}

	emptyArr := &Array{Elements: []Object{}}
	result = fn.Fn(emptyArr, &Builtin{}, NewInt(5))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinForEach(t *testing.T) {
	fn, ok := Builtins["forEach"]
	if !ok {
		t.Fatal("forEach builtin not found")
	}

	callCount := 0
	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		callCount++
		return NULL, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}

	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinFlatMap(t *testing.T) {
	fn, ok := Builtins["flatMap"]
	if !ok {
		t.Fatal("flatMap builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Array{Elements: []Object{NewInt(i.Value), NewInt(i.Value * 2)}}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 4 {
		t.Errorf("expected 4 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinEvery(t *testing.T) {
	fn, ok := Builtins["every"]
	if !ok {
		t.Fatal("every builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Bool{Value: i.Value > 0}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSome(t *testing.T) {
	fn, ok := Builtins["some"]
	if !ok {
		t.Fatal("some builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Bool{Value: i.Value > 5}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(10)}}
	result := fn.Fn(arr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGroupBy(t *testing.T) {
	fn, ok := Builtins["groupBy"]
	if !ok {
		t.Fatal("groupBy builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		if i.Value%2 == 0 {
			return NewString("even"), nil
		}
		return NewString("odd"), nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3), NewInt(4)}}
	result := fn.Fn(arr, &Builtin{})

	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 2 {
		t.Errorf("expected 2 groups, got %d", len(mapResult.Pairs))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinPartition(t *testing.T) {
	fn, ok := Builtins["partition"]
	if !ok {
		t.Fatal("partition builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Bool{Value: i.Value > 2}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3), NewInt(4)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 partitions, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinZip(t *testing.T) {
	fn, ok := Builtins["zip"]
	if !ok {
		t.Fatal("zip builtin not found")
	}

	arr1 := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	arr2 := &Array{Elements: []Object{NewString("a"), NewString("b")}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinUnzip(t *testing.T) {
	fn, ok := Builtins["unzip"]
	if !ok {
		t.Fatal("unzip builtin not found")
	}

	arr := &Array{Elements: []Object{
		&Array{Elements: []Object{NewInt(1), NewString("a")}},
		&Array{Elements: []Object{NewInt(2), NewString("b")}},
	}}
	result := fn.Fn(arr)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 arrays, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinFill(t *testing.T) {
	fn, ok := Builtins["fill"]
	if !ok {
		t.Fatal("fill builtin not found")
	}

	arr := &Array{Elements: []Object{NewInt(0), NewInt(0), NewInt(0)}}
	result := fn.Fn(arr, NewInt(5))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(5), NewInt(3))
	if !isError(result) {
		t.Error("expected error for non-array first arg")
	}
}

func TestBuiltinRangeNum(t *testing.T) {
	fn, ok := Builtins["rangeNum"]
	if !ok {
		t.Fatal("rangeNum builtin not found")
	}

	result := fn.Fn(NewInt(0), NewInt(5))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIntersection(t *testing.T) {
	fn, ok := Builtins["intersection"]
	if !ok {
		t.Fatal("intersection builtin not found")
	}

	arr1 := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	arr2 := &Array{Elements: []Object{NewInt(2), NewInt(3), NewInt(4)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 common elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinDifference(t *testing.T) {
	fn, ok := Builtins["difference"]
	if !ok {
		t.Fatal("difference builtin not found")
	}

	arr1 := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	arr2 := &Array{Elements: []Object{NewInt(2), NewInt(3), NewInt(4)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 1 {
		t.Errorf("expected 1 unique element, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinUnion(t *testing.T) {
	fn, ok := Builtins["union"]
	if !ok {
		t.Fatal("union builtin not found")
	}

	arr1 := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	arr2 := &Array{Elements: []Object{NewInt(2), NewInt(3)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCountBy(t *testing.T) {
	fn, ok := Builtins["countBy"]
	if !ok {
		t.Fatal("countBy builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		if i.Value%2 == 0 {
			return NewString("even"), nil
		}
		return NewString("odd"), nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3), NewInt(4)}}
	result := fn.Fn(arr, &Builtin{})

	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 2 {
		t.Errorf("expected 2 categories, got %d", len(mapResult.Pairs))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSortBy(t *testing.T) {
	fn, ok := Builtins["sortBy"]
	if !ok {
		t.Fatal("sortBy builtin not found")
	}

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return i, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(3), NewInt(1), NewInt(2)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

// ========== Additional edge case and error path tests ==========

func TestBuiltinMapArray_EmptyArray(t *testing.T) {
	fn, _ := Builtins["mapArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return NewInt(i.Value * 2), nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinMapArray_CallbackError(t *testing.T) {
	fn, _ := Builtins["mapArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinMapArray_CallbackReturnsError(t *testing.T) {
	fn, _ := Builtins["mapArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return newError("callback error"), nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinFilterArray_EmptyArray(t *testing.T) {
	fn, _ := Builtins["filterArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return FALSE, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinFilterArray_CallbackError(t *testing.T) {
	fn, _ := Builtins["filterArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinFilterArray_AllMatch(t *testing.T) {
	fn, _ := Builtins["filterArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return TRUE, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinFilterArray_NoneMatch(t *testing.T) {
	fn, _ := Builtins["filterArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return FALSE, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinReduceArray_EmptyArray(t *testing.T) {
	fn, _ := Builtins["reduceArray"]

	arr := &Array{Elements: []Object{}}
	result := fn.Fn(arr, &Builtin{})

	// Empty array without initial returns NULL
	if result != NULL {
		t.Errorf("expected NULL for empty array without initial, got %v", result)
	}
}

func TestBuiltinReduceArray_SingleElement(t *testing.T) {
	fn, _ := Builtins["reduceArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		a, _ := args[0].(*Int)
		b, _ := args[1].(*Int)
		return NewInt(a.Value + b.Value), nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(5)}}
	result := fn.Fn(arr, &Builtin{})

	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}
}

func TestBuiltinReduceArray_CallbackError(t *testing.T) {
	fn, _ := Builtins["reduceArray"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinForEach_EmptyArray(t *testing.T) {
	fn, _ := Builtins["forEach"]

	callCount := 0
	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		callCount++
		return NULL, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}

	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}
}

func TestBuiltinForEach_CallbackError(t *testing.T) {
	fn, _ := Builtins["forEach"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinFlatMap_EmptyArray(t *testing.T) {
	fn, _ := Builtins["flatMap"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return &Array{Elements: []Object{}}, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinFlatMap_CallbackError(t *testing.T) {
	fn, _ := Builtins["flatMap"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinEvery_EmptyArray(t *testing.T) {
	fn, _ := Builtins["every"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return TRUE, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	// Vacuous truth: all zero elements satisfy the predicate
	if !boolResult.Value {
		t.Error("expected true for empty array")
	}
}

func TestBuiltinEvery_FalseCase(t *testing.T) {
	fn, _ := Builtins["every"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Bool{Value: i.Value > 10}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false")
	}
}

func TestBuiltinEvery_CallbackError(t *testing.T) {
	fn, _ := Builtins["every"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinSome_EmptyArray(t *testing.T) {
	fn, _ := Builtins["some"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return TRUE, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	// No elements satisfy the predicate
	if boolResult.Value {
		t.Error("expected false for empty array")
	}
}

func TestBuiltinSome_FalseCase(t *testing.T) {
	fn, _ := Builtins["some"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return &Bool{Value: i.Value > 10}, nil
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	result := fn.Fn(arr, &Builtin{})

	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false")
	}
}

func TestBuiltinSome_CallbackError(t *testing.T) {
	fn, _ := Builtins["some"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinGroupBy_EmptyArray(t *testing.T) {
	fn, _ := Builtins["groupBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NewString("key"), nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 0 {
		t.Errorf("expected empty map, got %d pairs", len(mapResult.Pairs))
	}
}

func TestBuiltinGroupBy_CallbackError(t *testing.T) {
	fn, _ := Builtins["groupBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinPartition_EmptyArray(t *testing.T) {
	fn, _ := Builtins["partition"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return TRUE, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 partitions, got %d", len(arrResult.Elements))
	}
	// Both partitions should be empty
	if len(arrResult.Elements) == 2 {
		if first := arrResult.Elements[0].(*Array); len(first.Elements) != 0 {
			t.Errorf("expected empty true partition, got %d", len(first.Elements))
		}
		if second := arrResult.Elements[1].(*Array); len(second.Elements) != 0 {
			t.Errorf("expected empty false partition, got %d", len(second.Elements))
		}
	}
}

func TestBuiltinPartition_CallbackError(t *testing.T) {
	fn, _ := Builtins["partition"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinZip_EmptyArrays(t *testing.T) {
	fn, _ := Builtins["zip"]

	arr1 := &Array{Elements: []Object{}}
	arr2 := &Array{Elements: []Object{}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinZip_MismatchedLengths(t *testing.T) {
	fn, _ := Builtins["zip"]

	arr1 := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	arr2 := &Array{Elements: []Object{NewString("a"), NewString("b")}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	// Should truncate to shorter length
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinZip_NonArrayArg(t *testing.T) {
	fn, _ := Builtins["zip"]

	result := fn.Fn(&String{Value: "not array"}, &Array{Elements: []Object{}})

	if !isError(result) {
		t.Errorf("expected error for non-array arg, got %v", result)
	}
}

func TestBuiltinUnzip_EmptyArray(t *testing.T) {
	fn, _ := Builtins["unzip"]

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinUnzip_NonArrayElement(t *testing.T) {
	fn, _ := Builtins["unzip"]

	arr := &Array{Elements: []Object{&String{Value: "not array"}}}
	result := fn.Fn(arr)

	if !isError(result) {
		t.Errorf("expected error for non-array element, got %v", result)
	}
}

func TestBuiltinUnzip_SinglePair(t *testing.T) {
	fn, _ := Builtins["unzip"]

	arr := &Array{Elements: []Object{
		&Array{Elements: []Object{NewInt(1), NewString("a")}},
	}}
	result := fn.Fn(arr)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 arrays, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinFill_EmptyArray(t *testing.T) {
	fn, _ := Builtins["fill"]

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, NewInt(5))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinFill_NonArrayArg(t *testing.T) {
	fn, _ := Builtins["fill"]

	result := fn.Fn(NewInt(5), NewInt(3))

	if !isError(result) {
		t.Errorf("expected error for non-array arg, got %v", result)
	}
}

func TestBuiltinRangeNum_ZeroRange(t *testing.T) {
	fn, _ := Builtins["rangeNum"]

	result := fn.Fn(NewInt(5), NewInt(5))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array for zero range, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinRangeNum_Negative(t *testing.T) {
	fn, _ := Builtins["rangeNum"]

	result := fn.Fn(NewInt(-3), NewInt(2))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinRangeNum_StartGreaterThanStop(t *testing.T) {
	fn, _ := Builtins["rangeNum"]

	result := fn.Fn(NewInt(5), NewInt(2))

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array when start > stop, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinIntersection_EmptyArrays(t *testing.T) {
	fn, _ := Builtins["intersection"]

	arr1 := &Array{Elements: []Object{}}
	arr2 := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinIntersection_NonArrayArg(t *testing.T) {
	fn, _ := Builtins["intersection"]

	result := fn.Fn(&String{Value: "not array"}, &Array{Elements: []Object{}})

	if !isError(result) {
		t.Errorf("expected error for non-array arg, got %v", result)
	}
}

func TestBuiltinDifference_EmptyArrays(t *testing.T) {
	fn, _ := Builtins["difference"]

	arr1 := &Array{Elements: []Object{}}
	arr2 := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinUnion_EmptyArrays(t *testing.T) {
	fn, _ := Builtins["union"]

	arr1 := &Array{Elements: []Object{}}
	arr2 := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr1, arr2)

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinCountBy_EmptyArray(t *testing.T) {
	fn, _ := Builtins["countBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NewString("key"), nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 0 {
		t.Errorf("expected empty map, got %d pairs", len(mapResult.Pairs))
	}
}

func TestBuiltinCountBy_CallbackError(t *testing.T) {
	fn, _ := Builtins["countBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}

func TestBuiltinSortBy_EmptyArray(t *testing.T) {
	fn, _ := Builtins["sortBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		i, _ := args[0].(*Int)
		return i, nil
	})
	defer SetCallUserFuncImpl(nil)

	emptyArr := &Array{Elements: []Object{}}
	result := fn.Fn(emptyArr, &Builtin{})

	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arrResult.Elements))
	}
}

func TestBuiltinSortBy_NonArrayArg(t *testing.T) {
	fn, _ := Builtins["sortBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return TRUE, nil
	})
	defer SetCallUserFuncImpl(nil)

	result := fn.Fn(&String{Value: "not array"}, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error for non-array arg, got %v", result)
	}
}

func TestBuiltinSortBy_CallbackError(t *testing.T) {
	fn, _ := Builtins["sortBy"]

	SetCallUserFuncImpl(func(fn Object, args ...Object) (Object, error) {
		return NULL, fmt.Errorf("callback failed")
	})
	defer SetCallUserFuncImpl(nil)

	arr := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	result := fn.Fn(arr, &Builtin{})

	if !isError(result) {
		t.Errorf("expected error, got %v", result)
	}
}
