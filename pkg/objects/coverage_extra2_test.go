// pkg/objects/coverage_extra2_test.go
// Additional tests to improve code coverage for the objects package
package objects

import (
	"testing"
	"time"
)

// TestEncodeToml tests TOML encoding
func TestEncodeToml(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := EncodeToml(nil)
		if result != "" {
			t.Errorf("expected empty string for nil, got '%s'", result)
		}
	})

	t.Run("simple table", func(t *testing.T) {
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"name":  {Type: TomlString, Value: "test"},
				"count": {Type: TomlInteger, Value: int64(42)},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with float", func(t *testing.T) {
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"pi": {Type: TomlFloat, Value: 3.14159},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with boolean", func(t *testing.T) {
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"enabled":  {Type: TomlBoolean, Value: true},
				"disabled": {Type: TomlBoolean, Value: false},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with datetime", func(t *testing.T) {
		now := time.Now()
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"created": {Type: TomlDatetime, Value: now},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with date", func(t *testing.T) {
		date := time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC)
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"birthday": {Type: TomlDate, Value: date},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with time", func(t *testing.T) {
		tm := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"alarm": {Type: TomlTime, Value: tm},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("table with array", func(t *testing.T) {
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"items": {
					Type: TomlArray,
					Value: []*TomlValue{
						{Type: TomlString, Value: "a"},
						{Type: TomlString, Value: "b"},
					},
				},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})

	t.Run("nested table", func(t *testing.T) {
		root := &TomlValue{
			Type: TomlTable,
			Value: map[string]*TomlValue{
				"server": {
					Type: TomlTable,
					Value: map[string]*TomlValue{
						"host": {Type: TomlString, Value: "localhost"},
						"port": {Type: TomlInteger, Value: int64(8080)},
					},
				},
			},
		}
		result := EncodeToml(root)
		if len(result) == 0 {
			t.Error("expected non-empty TOML output")
		}
	})
}

// TestSetDelegateImpl tests the SetDelegateImpl function
func TestSetDelegateImpl(t *testing.T) {
	// Save original
	original := delegateImpl

	// Set a new implementation
	called := false
	testImpl := func(source string) (Object, error) {
		called = true
		return NewString("delegate result"), nil
	}

	_ = SetDelegateImpl(testImpl)

	// Verify the new implementation is set
	if delegateImpl == nil {
		t.Error("delegateImpl should not be nil after setting")
	}

	// Call the implementation
	if delegateImpl != nil {
		result, err := delegateImpl("test source")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !called {
			t.Error("implementation should have been called")
		}
		if s, ok := result.(*String); !ok || s.Value != "delegate result" {
			t.Errorf("unexpected result: %v", result)
		}
	}

	// Restore original
	SetDelegateImpl(original)
}

// TestDeepCopyObjectExtra tests deepCopyObject function
func TestDeepCopyObjectExtra(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		original := NewInt(42)
		copy := deepCopyObject(original)
		if copy.(*Int).Value != 42 {
			t.Errorf("expected 42, got %d", copy.(*Int).Value)
		}
	})

	t.Run("Float", func(t *testing.T) {
		original := NewFloat(3.14)
		copy := deepCopyObject(original)
		if copy.(*Float).Value != 3.14 {
			t.Errorf("expected 3.14, got %f", copy.(*Float).Value)
		}
	})

	t.Run("String", func(t *testing.T) {
		original := NewString("hello")
		copy := deepCopyObject(original)
		if copy.(*String).Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", copy.(*String).Value)
		}
	})

	t.Run("Bool true", func(t *testing.T) {
		copy := deepCopyObject(TRUE)
		if copy != TRUE {
			t.Error("expected TRUE")
		}
	})

	t.Run("Bool false", func(t *testing.T) {
		copy := deepCopyObject(FALSE)
		if copy != FALSE {
			t.Error("expected FALSE")
		}
	})

	t.Run("Null", func(t *testing.T) {
		copy := deepCopyObject(NULL)
		if copy != NULL {
			t.Error("expected NULL")
		}
	})

	t.Run("Array", func(t *testing.T) {
		original := NewArray([]Object{NewInt(1), NewInt(2)})
		copy := deepCopyObject(original).(*Array)
		if len(copy.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(copy.Elements))
		}
		// Verify it's a deep copy
		copy.Elements[0] = NewInt(99)
		if original.Elements[0].(*Int).Value != 1 {
			t.Error("original should not be modified")
		}
	})

	t.Run("Map", func(t *testing.T) {
		pairs := make(map[HashKey]MapPair)
		key := NewString("a")
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewInt(1)}
		original := NewMap(pairs)
		copy := deepCopyObject(original).(*Map)
		if len(copy.Pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(copy.Pairs))
		}
	})

	t.Run("Other types", func(t *testing.T) {
		// Test that other types are returned as-is
		fn := &Builtin{Fn: func(args ...Object) Object { return NULL }}
		copy := deepCopyObject(fn)
		if copy != fn {
			t.Error("builtin should be returned as-is")
		}
	})
}

// TestShallowCopyObject tests shallowCopyObject function
func TestShallowCopyObject(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		original := NewInt(42)
		copy := shallowCopyObject(original)
		if copy.(*Int).Value != 42 {
			t.Errorf("expected 42, got %d", copy.(*Int).Value)
		}
	})

	t.Run("String", func(t *testing.T) {
		original := NewString("hello")
		copy := shallowCopyObject(original)
		if copy.(*String).Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", copy.(*String).Value)
		}
	})

	t.Run("Array", func(t *testing.T) {
		original := NewArray([]Object{NewInt(1), NewInt(2)})
		copy := shallowCopyObject(original).(*Array)
		if len(copy.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(copy.Elements))
		}
		// Verify it's a shallow copy - modifying the copy should affect original
		// since elements are shared
		if &original.Elements == &copy.Elements {
			t.Error("elements slice should be different")
		}
	})

	t.Run("Map", func(t *testing.T) {
		pairs := make(map[HashKey]MapPair)
		key := NewString("a")
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewInt(1)}
		original := NewMap(pairs)
		copy := shallowCopyObject(original).(*Map)
		if len(copy.Pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(copy.Pairs))
		}
	})
}

// TestFlattenArrayExtra tests flattenArray function
func TestFlattenArrayExtra(t *testing.T) {
	t.Run("flat array", func(t *testing.T) {
		elements := []Object{NewInt(1), NewInt(2), NewInt(3)}
		result := flattenArray(elements, 0)
		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
	})

	t.Run("nested array depth 1", func(t *testing.T) {
		inner := NewArray([]Object{NewInt(2), NewInt(3)})
		elements := []Object{NewInt(1), inner}
		result := flattenArray(elements, 1)
		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
	})

	t.Run("nested array depth -1 (infinite)", func(t *testing.T) {
		innermost := NewArray([]Object{NewInt(4), NewInt(5)})
		middle := NewArray([]Object{NewInt(3), innermost})
		elements := []Object{NewInt(1), NewInt(2), middle}
		result := flattenArray(elements, -1)
		if len(result) != 5 {
			t.Errorf("expected 5 elements, got %d", len(result))
		}
	})

	t.Run("depth 0 (no flattening)", func(t *testing.T) {
		inner := NewArray([]Object{NewInt(2), NewInt(3)})
		elements := []Object{NewInt(1), inner}
		result := flattenArray(elements, 0)
		if len(result) != 2 {
			t.Errorf("expected 2 elements, got %d", len(result))
		}
	})
}

// TestCompareObjectsExtra tests compareObjects function
func TestCompareObjectsExtra(t *testing.T) {
	t.Run("same type equal", func(t *testing.T) {
		a := NewInt(42)
		b := NewInt(42)
		if !compareObjects(a, b) {
			t.Error("expected true for equal ints")
		}
	})

	t.Run("same type not equal", func(t *testing.T) {
		a := NewInt(42)
		b := NewInt(43)
		if compareObjects(a, b) {
			t.Error("expected false for different ints")
		}
	})

	t.Run("different types", func(t *testing.T) {
		a := NewInt(42)
		b := NewString("42")
		if compareObjects(a, b) {
			t.Error("expected false for different types")
		}
	})

	t.Run("strings equal", func(t *testing.T) {
		a := NewString("hello")
		b := NewString("hello")
		if !compareObjects(a, b) {
			t.Error("expected true for equal strings")
		}
	})
}

// TestNewErrorExtra2 tests newError function
func TestNewErrorExtra2(t *testing.T) {
	err := newError("test error: %s", "value")
	if err.Message != "test error: value" {
		t.Errorf("expected 'test error: value', got '%s'", err.Message)
	}

	err2 := newError("simple error")
	if err2.Message != "simple error" {
		t.Errorf("expected 'simple error', got '%s'", err2.Message)
	}
}

// TestRandomFunctions tests random number functions
func TestRandomFunctions(t *testing.T) {
	t.Run("randInt63", func(t *testing.T) {
		n := randInt63()
		if n < 0 {
			t.Error("randInt63 should return non-negative value")
		}
	})

	t.Run("multiple calls return different values", func(t *testing.T) {
		n1 := randInt63()
		n2 := randInt63()
		// Very unlikely to get the same value twice
		// (This test could theoretically fail, but probability is extremely low)
		if n1 == n2 {
			t.Log("Warning: got same random value twice (extremely rare)")
		}
	})
}

// TestBigIntMethodsExtra tests BigInt methods
func TestBigIntMethodsExtra(t *testing.T) {
	t.Run("DivInt", func(t *testing.T) {
		bi := NewBigIntFromInt64(100)
		result := bi.DivInt(5)
		val, _ := result.ToInt64()
		if val != 20 {
			t.Errorf("expected 20, got %d", val)
		}
	})

	t.Run("ModInt", func(t *testing.T) {
		bi := NewBigIntFromInt64(17)
		result := bi.ModInt(5)
		val, _ := result.ToInt64()
		if val != 2 {
			t.Errorf("expected 2, got %d", val)
		}
	})
}

// TestBigFloatMethodsExtra tests BigFloat methods
func TestBigFloatMethodsExtra(t *testing.T) {
	t.Run("ToInt64", func(t *testing.T) {
		bf := NewBigFloatFromFloat64(3.9)
		result, _ := bf.ToInt64()
		if result != 3 {
			t.Errorf("expected 3, got %d", result)
		}
	})
}
