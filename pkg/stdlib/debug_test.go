// pkg/stdlib/debug_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callDebugFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("debug")
	if mod == nil {
		panic("debug module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestDebugType(t *testing.T) {
	result := callDebugFunc("type", Int(42))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("type() should return String, got %T", result)
	}
	if s.Value != "INT" {
		t.Errorf("type() = %s, want 'INT'", s.Value)
	}
}

func TestDebugTypeTag(t *testing.T) {
	result := callDebugFunc("typeTag", Int(42))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("typeTag() should return Int, got %T", result)
	}
	if i.Value <= 0 {
		t.Errorf("typeTag() = %d, should be positive", i.Value)
	}
}

func TestDebugInspect(t *testing.T) {
	result := callDebugFunc("inspect", Int(42))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("inspect() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("inspect() should return non-empty string")
	}
}

func TestDebugMemStats(t *testing.T) {
	result := callDebugFunc("memStats")
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("memStats() should return Array, got %T", result)
	}
	if len(arr.Elements) != 4 {
		t.Errorf("memStats() should return 4 elements, got %d", len(arr.Elements))
	}
}

func TestDebugGC(t *testing.T) {
	result := callDebugFunc("gc")
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("gc() should return Null, got %T", result)
	}
}

func TestDebugGoroutines(t *testing.T) {
	result := callDebugFunc("goroutines")
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("goroutines() should return Int, got %T", result)
	}
	if i.Value <= 0 {
		t.Errorf("goroutines() = %d, should be positive", i.Value)
	}
}

func TestDebugStack(t *testing.T) {
	result := callDebugFunc("stack")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stack() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("stack() should return non-empty string")
	}
}

func TestDebugCallers(t *testing.T) {
	result := callDebugFunc("callers")
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("callers() should return Array, got %T", result)
	}
	// May have 0 or more callers depending on the context
	_ = arr
}

func TestDebugAssert(t *testing.T) {
	// Assert true - should return Null
	result := callDebugFunc("assert", Bool(true))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("assert(true) should return Null, got %T", result)
	}

	// Assert false - should return Error
	result = callDebugFunc("assert", Bool(false))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("assert(false) should return Error, got %T", result)
	}

	// Assert with custom message
	result = callDebugFunc("assert", Bool(false), String("custom error"))
	err, ok := result.(*objects.Error)
	if !ok {
		t.Fatalf("assert(false, msg) should return Error, got %T", result)
	}
	if err.Message != "custom error" {
		t.Errorf("assert() error = %s, want 'custom error'", err.Message)
	}
}

func TestDebugAssertEquals(t *testing.T) {
	// Equal values
	result := callDebugFunc("assertEquals", Int(42), Int(42))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("assertEquals(42, 42) should return Null, got %T", result)
	}

	// Unequal values
	result = callDebugFunc("assertEquals", Int(42), Int(43))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("assertEquals(42, 43) should return Error, got %T", result)
	}
}

func TestDebugDump(t *testing.T) {
	result := callDebugFunc("dump", Int(42))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("dump() should return Array, got %T", result)
	}
	if len(arr.Elements) < 2 {
		t.Errorf("dump() should return at least 2 elements, got %d", len(arr.Elements))
	}
}

func TestDebugIsNull(t *testing.T) {
	// Test with null
	result := callDebugFunc("isNull", Null())
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isNull() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isNull(null) should be true")
	}

	// Test with non-null
	result = callDebugFunc("isNull", Int(42))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isNull() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isNull(42) should be false")
	}
}

func TestDebugIsTruthy(t *testing.T) {
	// Test with truthy value
	result := callDebugFunc("isTruthy", Int(42))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isTruthy() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isTruthy(42) should be true")
	}

	// Test with falsy value
	result = callDebugFunc("isTruthy", Int(0))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isTruthy() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isTruthy(0) should be false")
	}
}

func TestDebugReflect(t *testing.T) {
	result := callDebugFunc("reflect", Int(42))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("reflect() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("reflect() should return 2 elements, got %d", len(arr.Elements))
	}
}

func TestDebugNanoTime(t *testing.T) {
	result := callDebugFunc("nanoTime")
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("nanoTime() should return Int, got %T", result)
	}
	if i.Value <= 0 {
		t.Errorf("nanoTime() = %d, should be positive", i.Value)
	}
}

func TestDebugElapsed(t *testing.T) {
	start := callDebugFunc("nanoTime")
	startInt, ok := start.(*objects.Int)
	if !ok {
		t.Fatalf("nanoTime() should return Int, got %T", start)
	}

	result := callDebugFunc("elapsed", startInt)
	elapsed, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("elapsed() should return Int, got %T", result)
	}
	if elapsed.Value < 0 {
		t.Errorf("elapsed() = %d, should be non-negative", elapsed.Value)
	}
}
