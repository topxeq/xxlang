// pkg/vm/escape_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestEscapeToHeap(t *testing.T) {
	obj := &objects.Int{Value: 42}
	result := escapeToHeap(obj)

	if result == nil {
		t.Fatal("escapeToHeap should return non-nil")
	}

	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if i.Value != 42 {
		t.Errorf("escapeToHeap: expected 42, got %d", i.Value)
	}
}

func TestStackBool(t *testing.T) {
	// Test true
	result := stackBool(true)
	if result != objects.TRUE {
		t.Error("stackBool(true) should return TRUE")
	}

	// Test false
	result = stackBool(false)
	if result != objects.FALSE {
		t.Error("stackBool(false) should return FALSE")
	}
}

func TestNullIfZero(t *testing.T) {
	// Test with nil
	result := nullIfZero(nil)
	if result != objects.NULL {
		t.Error("nullIfZero(nil) should return NULL")
	}

	// Test with non-nil object
	obj := &objects.Int{Value: 42}
	result = nullIfZero(obj)
	if result != obj {
		t.Error("nullIfZero with non-nil should return the same object")
	}
}

func TestIntResult(t *testing.T) {
	result := intResult(42)

	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if i.Value != 42 {
		t.Errorf("intResult: expected 42, got %d", i.Value)
	}
}

func TestFloatResult(t *testing.T) {
	result := floatResult(3.14)

	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}

	if f.Value != 3.14 {
		t.Errorf("floatResult: expected 3.14, got %f", f.Value)
	}
}
