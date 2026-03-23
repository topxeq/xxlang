// pkg/stdlib/bytesbuffer_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestBytesBuffer_Module(t *testing.T) {
	module := Get("bytesbuffer")
	if module == nil {
		t.Fatal("bytesbuffer module not found")
	}
	if module.Name != "bytesbuffer" {
		t.Errorf("expected module name 'bytesbuffer', got %q", module.Name)
	}
}

func TestBytesBuffer_Create(t *testing.T) {
	module := Get("bytesbuffer")
	createFn, ok := module.Exports["create"]
	if !ok {
		t.Fatal("create function not found")
	}

	builtin, ok := createFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", createFn)
	}

	t.Run("create empty", func(t *testing.T) {
		result := builtin.Fn()
		bb, ok := result.(*objects.BytesBuffer)
		if !ok {
			t.Fatalf("expected BytesBuffer, got %T", result)
		}
		if bb.Len() != 0 {
			t.Errorf("expected empty BytesBuffer, got len %d", bb.Len())
		}
	})

	t.Run("create with capacity", func(t *testing.T) {
		result := builtin.Fn(objects.NewInt(100))
		bb, ok := result.(*objects.BytesBuffer)
		if !ok {
			t.Fatalf("expected BytesBuffer, got %T", result)
		}
		if bb.Cap() < 100 {
			t.Errorf("expected capacity >= 100, got %d", bb.Cap())
		}
	})

	t.Run("create with invalid capacity type", func(t *testing.T) {
		result := builtin.Fn(objects.NewString("100"))
		if result.Type() != objects.ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})
}

func TestBytesBuffer_FromString(t *testing.T) {
	module := Get("bytesbuffer")
	fromStringFn, ok := module.Exports["fromString"]
	if !ok {
		t.Fatal("fromString function not found")
	}

	builtin, ok := fromStringFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fromStringFn)
	}

	result := builtin.Fn(objects.NewString("hello"))
	bb, ok := result.(*objects.BytesBuffer)
	if !ok {
		t.Fatalf("expected BytesBuffer, got %T", result)
	}
	if bb.String() != "hello" {
		t.Errorf("expected 'hello', got %q", bb.String())
	}
}

func TestBytesBuffer_FromBytes(t *testing.T) {
	module := Get("bytesbuffer")
	fromBytesFn, ok := module.Exports["fromBytes"]
	if !ok {
		t.Fatal("fromBytes function not found")
	}

	builtin, ok := fromBytesFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fromBytesFn)
	}

	arr := &objects.Array{
		Elements: []objects.Object{
			objects.NewInt(65),
			objects.NewInt(66),
			objects.NewInt(67),
		},
	}
	result := builtin.Fn(arr)
	bb, ok := result.(*objects.BytesBuffer)
	if !ok {
		t.Fatalf("expected BytesBuffer, got %T", result)
	}
	if bb.String() != "ABC" {
		t.Errorf("expected 'ABC', got %q", bb.String())
	}
}

func TestBytesBuffer_IsBytesBuffer(t *testing.T) {
	module := Get("bytesbuffer")
	isBB, ok := module.Exports["isBytesBuffer"]
	if !ok {
		t.Fatal("isBytesBuffer function not found")
	}

	builtin, ok := isBB.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", isBB)
	}

	t.Run("is BytesBuffer", func(t *testing.T) {
		bb := objects.NewBytesBuffer()
		result := builtin.Fn(bb)
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != true {
			t.Error("expected true for BytesBuffer")
		}
	})

	t.Run("is not BytesBuffer - String", func(t *testing.T) {
		result := builtin.Fn(objects.NewString("test"))
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != false {
			t.Error("expected false for String")
		}
	})
}
