// pkg/stdlib/bytes_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callBytesFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("bytes")
	if mod == nil {
		panic("bytes module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestBytesFromString(t *testing.T) {
	result := callBytesFunc("fromString", String("hello"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("fromString() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("fromString('hello') length = %d, want 5", len(arr.Elements))
	}

	// Check individual bytes
	expected := []int64{104, 101, 108, 108, 111}
	for i, e := range expected {
		if arr.Elements[i].(*objects.Int).Value != e {
			t.Errorf("fromString()[%d] = %d, want %d", i, arr.Elements[i].(*objects.Int).Value, e)
		}
	}
}

func TestBytesFromStringErrors(t *testing.T) {
	// Wrong number of args
	result := callBytesFunc("fromString")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("fromString() with no args should return Error")
	}

	// Wrong type
	result = callBytesFunc("fromString", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("fromString() with non-string should return Error")
	}
}

func TestBytesToString(t *testing.T) {
	arr := Array(Int(104), Int(101), Int(108), Int(108), Int(111))
	result := callBytesFunc("toString", arr)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("toString() should return String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("toString() = %s, want 'hello'", s.Value)
	}
}

func TestBytesToStringErrors(t *testing.T) {
	// Wrong number of args
	result := callBytesFunc("toString")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("toString() with no args should return Error")
	}

	// Wrong type
	result = callBytesFunc("toString", String("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("toString() with non-array should return Error")
	}

	// Non-integer elements
	result = callBytesFunc("toString", Array(String("a")))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("toString() with non-int elements should return Error")
	}

	// Out of range
	result = callBytesFunc("toString", Array(Int(300)))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("toString() with out-of-range value should return Error")
	}
}

func TestBytesGet(t *testing.T) {
	arr := Array(Int(10), Int(20), Int(30))
	result := callBytesFunc("get", arr, Int(1))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("get() should return Int, got %T", result)
	}
	if i.Value != 20 {
		t.Errorf("get(arr, 1) = %d, want 20", i.Value)
	}
}

func TestBytesGetErrors(t *testing.T) {
	// Wrong number of args
	result := callBytesFunc("get", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("get() with wrong args should return Error")
	}

	// Out of range
	result = callBytesFunc("get", Array(Int(1)), Int(10))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("get() with out-of-range index should return Error")
	}
}

func TestBytesSet(t *testing.T) {
	arr := Array(Int(10), Int(20), Int(30))
	result := callBytesFunc("set", arr, Int(1), Int(99))
	resultArr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("set() should return Array, got %T", result)
	}
	if resultArr.Elements[1].(*objects.Int).Value != 99 {
		t.Errorf("set() did not update element, got %d", resultArr.Elements[1].(*objects.Int).Value)
	}
}

func TestBytesEncodeInt64BE(t *testing.T) {
	result := callBytesFunc("encodeInt64BE", Int(0x0102030405060708))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("encodeInt64BE() should return Array, got %T", result)
	}
	if len(arr.Elements) != 8 {
		t.Errorf("encodeInt64BE() length = %d, want 8", len(arr.Elements))
	}
	// Big endian: first byte should be 1
	if arr.Elements[0].(*objects.Int).Value != 1 {
		t.Errorf("encodeInt64BE()[0] = %d, want 1", arr.Elements[0].(*objects.Int).Value)
	}
}

func TestBytesDecodeInt64BE(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5), Int(6), Int(7), Int(8))
	result := callBytesFunc("decodeInt64BE", arr)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("decodeInt64BE() should return Int, got %T", result)
	}
	expected := int64(0x0102030405060708)
	if i.Value != expected {
		t.Errorf("decodeInt64BE() = %d, want %d", i.Value, expected)
	}
}

func TestBytesDecodeInt64BEErrors(t *testing.T) {
	// Wrong array size
	result := callBytesFunc("decodeInt64BE", Array(Int(1), Int(2)))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("decodeInt64BE() with wrong size should return Error")
	}
}

func TestBytesEncodeInt64LE(t *testing.T) {
	result := callBytesFunc("encodeInt64LE", Int(0x0102030405060708))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("encodeInt64LE() should return Array, got %T", result)
	}
	// Little endian: last byte should be 1
	if arr.Elements[7].(*objects.Int).Value != 1 {
		t.Errorf("encodeInt64LE()[7] = %d, want 1", arr.Elements[7].(*objects.Int).Value)
	}
}

func TestBytesDecodeInt64LE(t *testing.T) {
	arr := Array(Int(8), Int(7), Int(6), Int(5), Int(4), Int(3), Int(2), Int(1))
	result := callBytesFunc("decodeInt64LE", arr)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("decodeInt64LE() should return Int, got %T", result)
	}
	expected := int64(0x0102030405060708)
	if i.Value != expected {
		t.Errorf("decodeInt64LE() = %d, want %d", i.Value, expected)
	}
}

func TestBytesConcat(t *testing.T) {
	arr1 := Array(Int(1), Int(2))
	arr2 := Array(Int(3), Int(4))
	result := callBytesFunc("concat", arr1, arr2)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("concat() should return Array, got %T", result)
	}
	if len(arr.Elements) != 4 {
		t.Errorf("concat() length = %d, want 4", len(arr.Elements))
	}
}

func TestBytesSlice(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3), Int(4), Int(5))
	result := callBytesFunc("slice", arr, Int(1), Int(4))
	slice, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("slice() should return Array, got %T", result)
	}
	if len(slice.Elements) != 3 {
		t.Errorf("slice() length = %d, want 3", len(slice.Elements))
	}
}

func TestBytesCompare(t *testing.T) {
	arr1 := Array(Int(1), Int(2))
	arr2 := Array(Int(1), Int(3))
	result := callBytesFunc("compare", arr1, arr2)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("compare() should return Int, got %T", result)
	}
	if i.Value >= 0 {
		t.Errorf("compare([1,2], [1,3]) = %d, should be < 0", i.Value)
	}
}

func TestBytesEqual(t *testing.T) {
	arr1 := Array(Int(1), Int(2))
	arr2 := Array(Int(1), Int(2))
	result := callBytesFunc("equal", arr1, arr2)
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("equal() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("equal([1,2], [1,2]) should be true")
	}
}

func TestBytesCount(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(1), Int(1))
	result := callBytesFunc("count", arr, Int(1))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("count() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("count([1,2,1,1], 1) = %d, want 3", i.Value)
	}
}

func TestBytesIndexOf(t *testing.T) {
	arr := Array(Int(10), Int(20), Int(30))
	result := callBytesFunc("indexOf", arr, Int(20))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("indexOf() should return Int, got %T", result)
	}
	if i.Value != 1 {
		t.Errorf("indexOf([10,20,30], 20) = %d, want 1", i.Value)
	}

	// Not found
	result = callBytesFunc("indexOf", arr, Int(99))
	i, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("indexOf() should return Int, got %T", result)
	}
	if i.Value != -1 {
		t.Errorf("indexOf() for not found should return -1, got %d", i.Value)
	}
}
