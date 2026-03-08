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
