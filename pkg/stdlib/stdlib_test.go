// pkg/stdlib/stdlib_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestRegistry(t *testing.T) {
	modules := []string{"std/math", "std/string", "std/array", "std/io"}

	for _, name := range modules {
		if !Has(name) {
			t.Errorf("module %s not registered", name)
		}

		mod := Get(name)
		if mod == nil {
			t.Errorf("module %s is nil", name)
			continue
		}

		if mod.Name != name {
				t.Errorf("module name mismatch: got %s, want %s", mod.Name, name)
		}

		if len(mod.Exports) == 0 {
			t.Errorf("module %s has no exports", name)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	f := Float(3.14)
	if f.Value != 3.14 {
		t.Errorf("Float() = %v, want 3.14", f.Value)
	}

	s := String("hello")
	if s.Value != "hello" {
		t.Errorf("String() = %v, want hello", s.Value)
	}

	i := Int(42)
	if i.Value != 42 {
		t.Errorf("Int() = %v, want 42", i.Value)
	}

	if !Bool(true).Value {
		t.Error("Bool(true) should be true")
	}
	if Bool(false).Value {
		t.Error("Bool(false) should be false")
	}

	arr := Array(Int(1), Int(2), Int(3))
	if len(arr.Elements) != 3 {
		t.Errorf("Array() length = %d, want 3", len(arr.Elements))
	}

	if Null() != objects.NULL {
		t.Error("Null() should return objects.NULL")
	}

	err := Error("test error")
	if err.Message != "test error" {
		t.Errorf("Error() = %v, want 'test error'", err.Message)
	}
}

func TestBuiltinFunc(t *testing.T) {
	called := false
	fn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		called = true
		return Int(int64(len(args)))
	})

	if fn.Fn == nil {
		t.Fatal("BuiltinFunc created nil function")
	}

	result := fn.Fn(Int(1), Int(2), Int(3))
	if !called {
		t.Error("function was not called")
	}

	if result.(*objects.Int).Value != 3 {
		t.Errorf("function returned %v, want 3", result)
	}
}
