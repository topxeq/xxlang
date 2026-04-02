// pkg/stdlib/utils_extra_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestUtilsModule(t *testing.T) {
	mod := Get("utils")
	if mod == nil {
		t.Fatal("utils module not found")
	}
}

func TestBuiltinTypeOf(t *testing.T) {
	fn, ok := objects.Builtins["typeOf"]
	if !ok {
		t.Fatal("typeOf builtin not found")
	}

	result := fn.Fn(objects.NewInt(42))
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "INT" {
		t.Errorf("expected 'INT', got '%s'", strResult.Value)
	}

	result = fn.Fn(objects.NewString("hello"))
	strResult, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "STRING" {
		t.Errorf("expected 'STRING', got '%s'", strResult.Value)
	}

	result = fn.Fn(objects.TRUE)
	strResult, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "BOOL" {
		t.Errorf("expected 'BOOL', got '%s'", strResult.Value)
	}
}

func TestBuiltinLen(t *testing.T) {
	fn, ok := objects.Builtins["len"]
	if !ok {
		t.Fatal("len builtin not found")
	}

	result := fn.Fn(&objects.Array{Elements: []objects.Object{objects.NewInt(1), objects.NewInt(2)}})
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 2 {
		t.Errorf("expected 2, got %d", intResult.Value)
	}

	result = fn.Fn(objects.NewString("hello"))
	intResult, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}
}

func TestBuiltinPrint(t *testing.T) {
	fn, ok := objects.Builtins["print"]
	if !ok {
		t.Fatal("print builtin not found")
	}

	result := fn.Fn(objects.NewString("hello"))
	if result != objects.NULL {
		t.Errorf("expected NULL, got %v", result)
	}
}

func TestBuiltinPrintln(t *testing.T) {
	fn, ok := objects.Builtins["println"]
	if !ok {
		t.Fatal("println builtin not found")
	}

	result := fn.Fn(objects.NewString("hello"))
	if result != objects.NULL {
		t.Errorf("expected NULL, got %v", result)
	}
}

func TestBuiltinSprintf(t *testing.T) {
	fn, ok := objects.Builtins["sprintf"]
	if !ok {
		t.Fatal("sprintf builtin not found")
	}

	result := fn.Fn(objects.NewString("value: %d"), objects.NewInt(42))
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "value: 42" {
		t.Errorf("expected 'value: 42', got '%s'", strResult.Value)
	}
}

func TestBuiltinPrintf(t *testing.T) {
	fn, ok := objects.Builtins["printf"]
	if !ok {
		t.Fatal("printf builtin not found")
	}

	result := fn.Fn(objects.NewString("value: %d"), objects.NewInt(42))
	if result != objects.NULL {
		t.Errorf("expected NULL, got %v", result)
	}
}
