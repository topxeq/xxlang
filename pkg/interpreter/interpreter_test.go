// pkg/interpreter/interpreter_test.go
package interpreter

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestInterpreter_New(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		interp := New()
		if interp == nil {
			t.Fatal("expected interpreter, got nil")
		}
	})

	t.Run("with stdlib", func(t *testing.T) {
		interp := New(WithStdlib())
		result, err := interp.Eval(`len("hello")`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
	})
}

func TestInterpreter_Eval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "42", int64(42)},
		{"float", "3.14", 3.14},
		{"string", `"hello"`, "hello"},
		{"boolean", "true", true},
		{"arithmetic", "2 + 3 * 4", int64(14)},
		{"comparison", "1 < 2", true},
		{"array", "[1, 2, 3][0]", int64(1)},
		{"map", `{"a": 1}["a"]`, int64(1)},
	}

	interp := New(WithStdlib())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := interp.Eval(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			verifyResult(t, result, tt.expected)
		})
	}
}

func TestInterpreter_SetGetGlobal(t *testing.T) {
	interp := New(WithStdlib())

	t.Run("set and get integer", func(t *testing.T) {
		err := interp.SetGlobal("x", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		val, ok := interp.GetGlobal("x")
		if !ok {
			t.Fatal("expected to find global x")
		}
		if intVal, ok := val.(*objects.Int); !ok {
			t.Fatalf("expected Int, got %T", val)
		} else if intVal.Value != 42 {
			t.Fatalf("expected 42, got %d", intVal.Value)
		}
	})

	t.Run("use in code", func(t *testing.T) {
		err := interp.SetGlobal("y", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result, err := interp.Eval("y * 2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		verifyResult(t, result, int64(20))
	})
}

func TestInterpreter_Errors(t *testing.T) {
	interp := New(WithStdlib())

	t.Run("compile error", func(t *testing.T) {
		_, err := interp.Eval("var x = ")
		if err == nil {
			t.Fatal("expected error for incomplete expression")
		}
	})

	t.Run("runtime error", func(t *testing.T) {
		_, err := interp.Eval("1 / 0")
		if err == nil {
			t.Fatal("expected error for division by zero")
		}
	})
}

func verifyResult(t *testing.T, obj objects.Object, expected interface{}) {
	t.Helper()
	switch exp := expected.(type) {
	case int64:
		if intVal, ok := obj.(*objects.Int); !ok {
			t.Fatalf("expected Int, got %T", obj)
		} else if intVal.Value != exp {
			t.Fatalf("expected %d, got %d", exp, intVal.Value)
		}
	case float64:
		if floatVal, ok := obj.(*objects.Float); !ok {
			t.Fatalf("expected Float, got %T", obj)
		} else if floatVal.Value != exp {
			t.Fatalf("expected %f, got %f", exp, floatVal.Value)
		}
	case string:
		if strVal, ok := obj.(*objects.String); !ok {
			t.Fatalf("expected String, got %T", obj)
		} else if strVal.Value != exp {
			t.Fatalf("expected %q, got %q", exp, strVal.Value)
		}
	case bool:
		if boolVal, ok := obj.(*objects.Bool); !ok {
			t.Fatalf("expected Bool, got %T", obj)
		} else if boolVal.Value != exp {
			t.Fatalf("expected %t, got %t", exp, boolVal.Value)
		}
	}
}
