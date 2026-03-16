// pkg/stdlib/utils_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestDeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		input    objects.Object
		expected objects.Object
	}{
		{
			name:     "int",
			input:    Int(42),
			expected: Int(42),
		},
		{
			name:     "float",
			input:    Float(3.14),
			expected: Float(3.14),
		},
		{
			name:     "string",
			input:    String("hello"),
			expected: String("hello"),
		},
		{
			name:     "bool true",
			input:    Bool(true),
			expected: Bool(true),
		},
		{
			name:     "bool false",
			input:    Bool(false),
			expected: Bool(false),
		},
		{
			name:     "null",
			input:    Null(),
			expected: objects.NULL,
		},
		{
			name:     "array",
			input:    Array(Int(1), Int(2), Int(3)),
			expected: Array(Int(1), Int(2), Int(3)),
		},
		{
			name:     "nested array",
			input:    Array(Array(Int(1), Int(2)), Int(3)),
			expected: Array(Array(Int(1), Int(2)), Int(3)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := Get("utils")
			if mod == nil {
				t.Fatal("utils module not found")
			}
			fn, ok := mod.Exports["deepCopy"]
			if !ok {
				t.Fatal("deepCopy not found in utils")
			}
			builtin := fn.(*objects.Builtin)
			result := builtin.Fn(tt.input)

			if !deepCompare(result, tt.expected) {
				t.Errorf("deepCopy() = %v, want %v", result, tt.expected)
			}

			// Verify it's a copy (not same reference for arrays/maps)
			if arr, ok := tt.input.(*objects.Array); ok {
				if result == arr {
					t.Error("deepCopy should return a new array, not same reference")
				}
			}
		})
	}
}

func TestDeepCopyMap(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["deepCopy"].(*objects.Builtin)

	// Create a map
	mapObj := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		},
	}

	result := fn.Fn(mapObj)
	resultMap, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	if len(resultMap.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(resultMap.Pairs))
	}

	// Verify it's a different object
	if resultMap == mapObj {
		t.Error("deepCopy should return a new map, not same reference")
	}
}

func TestShallowCopy(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["shallowCopy"].(*objects.Builtin)

	tests := []struct {
		name  string
		input objects.Object
	}{
		{"array", Array(Int(1), Int(2), Int(3))},
		{"map", &objects.Map{
			Pairs: map[objects.HashKey]objects.MapPair{
				String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			},
		}},
		{"int (returns same)", Int(42)},
		{"string (returns same)", String("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.input)

			switch v := tt.input.(type) {
			case *objects.Array:
				arr, ok := result.(*objects.Array)
				if !ok {
					t.Fatalf("expected Array, got %T", result)
				}
				if arr == v {
					t.Error("shallowCopy should return a new array")
				}
				if len(arr.Elements) != len(v.Elements) {
					t.Errorf("array length mismatch")
				}
			case *objects.Map:
				m, ok := result.(*objects.Map)
				if !ok {
					t.Fatalf("expected Map, got %T", result)
				}
				if m == v {
					t.Error("shallowCopy should return a new map")
				}
			default:
				if result != tt.input {
					t.Error("shallowCopy should return same object for non-array/map types")
				}
			}
		})
	}
}

func TestDeepMerge(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["deepMerge"].(*objects.Builtin)

	target := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		},
	}

	source := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("b").HashKey(): {Key: String("b"), Value: Int(3)},
			String("c").HashKey(): {Key: String("c"), Value: Int(4)},
		},
	}

	result := fn.Fn(target, source)
	merged, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	// Check that 'a' is preserved
	aVal := merged.Pairs[String("a").HashKey()]
	if aVal.Value.(*objects.Int).Value != 1 {
		t.Errorf("expected a=1, got %v", aVal.Value)
	}

	// Check that 'b' is overwritten
	bVal := merged.Pairs[String("b").HashKey()]
	if bVal.Value.(*objects.Int).Value != 3 {
		t.Errorf("expected b=3 (overwritten), got %v", bVal.Value)
	}

	// Check that 'c' is added
	cVal := merged.Pairs[String("c").HashKey()]
	if cVal.Value.(*objects.Int).Value != 4 {
		t.Errorf("expected c=4, got %v", cVal.Value)
	}
}

func TestDeepMergeNested(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["deepMerge"].(*objects.Builtin)

	target := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("nested").HashKey(): {
				Key: String("nested"),
				Value: &objects.Map{
					Pairs: map[objects.HashKey]objects.MapPair{
						String("a").HashKey(): {Key: String("a"), Value: Int(1)},
						String("b").HashKey(): {Key: String("b"), Value: Int(2)},
					},
				},
			},
		},
	}

	source := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("nested").HashKey(): {
				Key: String("nested"),
				Value: &objects.Map{
					Pairs: map[objects.HashKey]objects.MapPair{
						String("b").HashKey(): {Key: String("b"), Value: Int(3)},
						String("c").HashKey(): {Key: String("c"), Value: Int(4)},
					},
				},
			},
		},
	}

	result := fn.Fn(target, source)
	merged := result.(*objects.Map)

	nested := merged.Pairs[String("nested").HashKey()].Value.(*objects.Map)
	if len(nested.Pairs) != 3 {
		t.Errorf("expected 3 keys in nested map, got %d", len(nested.Pairs))
	}
}

func TestDeepEquals(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["deepEquals"].(*objects.Builtin)

	tests := []struct {
		name     string
		a        objects.Object
		b        objects.Object
		expected bool
	}{
		{"int equal", Int(42), Int(42), true},
		{"int not equal", Int(42), Int(43), false},
		{"float equal", Float(3.14), Float(3.14), true},
		{"string equal", String("hello"), String("hello"), true},
		{"bool equal", Bool(true), Bool(true), true},
		{"null equal", Null(), Null(), true},
		{"array equal", Array(Int(1), Int(2)), Array(Int(1), Int(2)), true},
		{"array different length", Array(Int(1)), Array(Int(1), Int(2)), false},
		{"array different values", Array(Int(1), Int(2)), Array(Int(1), Int(3)), false},
		{"map equal",
			&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
				String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			}},
			&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
				String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			}}, true},
		{"different types", Int(42), String("42"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.a, tt.b)
			b, ok := result.(*objects.Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("deepEquals(%v, %v) = %v, want %v", tt.a, tt.b, b.Value, tt.expected)
			}
		})
	}
}

func TestPick(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["pick"].(*objects.Builtin)

	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
			String("c").HashKey(): {Key: String("c"), Value: Int(3)},
		},
	}

	keys := Array(String("a"), String("c"))
	result := fn.Fn(m, keys)
	picked, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	if len(picked.Pairs) != 2 {
		t.Errorf("expected 2 keys, got %d", len(picked.Pairs))
	}

	if _, exists := picked.Pairs[String("b").HashKey()]; exists {
		t.Error("key 'b' should not be in result")
	}
}

func TestOmit(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["omit"].(*objects.Builtin)

	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
			String("c").HashKey(): {Key: String("c"), Value: Int(3)},
		},
	}

	keys := Array(String("b"))
	result := fn.Fn(m, keys)
	omitted, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	if len(omitted.Pairs) != 2 {
		t.Errorf("expected 2 keys, got %d", len(omitted.Pairs))
	}

	if _, exists := omitted.Pairs[String("b").HashKey()]; exists {
		t.Error("key 'b' should not be in result")
	}
}

func TestKeys(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["keys"].(*objects.Builtin)

	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		},
	}

	result := fn.Fn(m)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 keys, got %d", len(arr.Elements))
	}
}

func TestValues(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["values"].(*objects.Builtin)

	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		},
	}

	result := fn.Fn(m)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 values, got %d", len(arr.Elements))
	}
}

func TestEntries(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["entries"].(*objects.Builtin)

	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		},
	}

	result := fn.Fn(m)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(arr.Elements))
	}

	entry, ok := arr.Elements[0].(*objects.Array)
	if !ok {
		t.Fatalf("expected entry to be Array, got %T", arr.Elements[0])
	}

	if len(entry.Elements) != 2 {
		t.Errorf("expected 2 elements in entry, got %d", len(entry.Elements))
	}
}

func TestFromEntries(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["fromEntries"].(*objects.Builtin)

	entries := Array(
		Array(String("a"), Int(1)),
		Array(String("b"), Int(2)),
	)

	result := fn.Fn(entries)
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	if len(m.Pairs) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m.Pairs))
	}

	aVal := m.Pairs[String("a").HashKey()].Value.(*objects.Int)
	if aVal.Value != 1 {
		t.Errorf("expected a=1, got %d", aVal.Value)
	}
}

func TestType(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["type"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected string
	}{
		{Int(42), "INT"},
		{Float(3.14), "FLOAT"},
		{String("hello"), "STRING"},
		{Bool(true), "BOOL"},
		{Null(), "NULL"},
		{Array(Int(1)), "ARRAY"},
		{&objects.Map{}, "MAP"},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("type(%v) = %s, want %s", tt.input, s.Value, tt.expected)
		}
	}
}

func TestIsPrimitive(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["isPrimitive"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected bool
	}{
		{Int(42), true},
		{Float(3.14), true},
		{String("hello"), true},
		{Bool(true), true},
		{Null(), true},
		{Array(Int(1)), false},
		{&objects.Map{}, false},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isPrimitive(%v) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["isEmpty"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected bool
	}{
		{String(""), true},
		{String("hello"), false},
		{Array(), true},
		{Array(Int(1)), false},
		{&objects.Map{}, true},
		{&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		}}, false},
		{Null(), true},
		{Int(0), false},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isEmpty(%v) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestSize(t *testing.T) {
	mod := Get("utils")
	fn := mod.Exports["size"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected int64
	}{
		{String("hello"), 5},
		{Array(Int(1), Int(2), Int(3)), 3},
		{&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		}}, 2},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("size(%v) = %d, want %d", tt.input, i.Value, tt.expected)
		}
	}
}

func TestDeepCopyValueInternal(t *testing.T) {
	// Test the internal deepCopyValue function directly
	tests := []struct {
		name  string
		input objects.Object
	}{
		{"int", Int(42)},
		{"float", Float(3.14)},
		{"string", String("hello")},
		{"bool true", Bool(true)},
		{"bool false", Bool(false)},
		{"null", Null()},
		{"array", Array(Int(1), Int(2))},
		{"nested array", Array(Array(Int(1)), Int(2))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepCopyValue(tt.input)
			if !deepCompare(result, tt.input) {
				t.Errorf("deepCopyValue mismatch")
			}
		})
	}
}

func TestShallowCopyValueInternal(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := shallowCopyValue(arr)

	if result == arr {
		t.Error("should return a new array")
	}

	resultArr := result.(*objects.Array)
	if len(resultArr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(resultArr.Elements))
	}

	// Non-array/map should return same object
	intObj := Int(42)
	result = shallowCopyValue(intObj)
	if result != intObj {
		t.Error("shallowCopy of int should return same object")
	}
}

func TestDeepMergeMapsInternal(t *testing.T) {
	target := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		},
	}

	source := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		},
	}

	result := deepMergeMaps(target, source)
	if len(result.Pairs) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Pairs))
	}
}

func TestDeepCompareInternal(t *testing.T) {
	tests := []struct {
		a, b     objects.Object
		expected bool
	}{
		{nil, nil, true},
		{nil, Int(1), false},
		{Int(1), nil, false},
		{Int(42), Int(42), true},
		{Int(42), Int(43), false},
		{Int(42), String("42"), false},
	}

	for _, tt := range tests {
		result := deepCompare(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("deepCompare(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestPickKeysInternal(t *testing.T) {
	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
			String("c").HashKey(): {Key: String("c"), Value: Int(3)},
		},
	}

	keys := Array(String("a"), String("c"))
	result := pickKeys(m, keys)

	if len(result.Pairs) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Pairs))
	}
}

func TestOmitKeysInternal(t *testing.T) {
	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("a").HashKey(): {Key: String("a"), Value: Int(1)},
			String("b").HashKey(): {Key: String("b"), Value: Int(2)},
			String("c").HashKey(): {Key: String("c"), Value: Int(3)},
		},
	}

	keys := Array(String("b"))
	result := omitKeys(m, keys)

	if len(result.Pairs) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Pairs))
	}

	if _, exists := result.Pairs[String("b").HashKey()]; exists {
		t.Error("key 'b' should be omitted")
	}
}

func TestUtilsModuleErrors(t *testing.T) {
	mod := Get("utils")

	// Test argument errors
	tests := []struct {
		name   string
		fnName string
		args   []objects.Object
	}{
		{"deepCopy no args", "deepCopy", []objects.Object{}},
		{"shallowCopy no args", "shallowCopy", []objects.Object{}},
		{"deepMerge no args", "deepMerge", []objects.Object{}},
		{"deepMerge one arg", "deepMerge", []objects.Object{&objects.Map{}}},
		{"deepMerge wrong first arg", "deepMerge", []objects.Object{Int(1), &objects.Map{}}},
		{"deepMerge wrong second arg", "deepMerge", []objects.Object{&objects.Map{}, Int(1)}},
		{"deepEquals no args", "deepEquals", []objects.Object{}},
		{"deepEquals one arg", "deepEquals", []objects.Object{Int(1)}},
		{"pick no args", "pick", []objects.Object{}},
		{"pick wrong first arg", "pick", []objects.Object{Int(1), Array()}},
		{"pick wrong second arg", "pick", []objects.Object{&objects.Map{}, Int(1)}},
		{"omit no args", "omit", []objects.Object{}},
		{"omit wrong first arg", "omit", []objects.Object{Int(1), Array()}},
		{"keys no args", "keys", []objects.Object{}},
		{"keys wrong arg", "keys", []objects.Object{Int(1)}},
		{"values no args", "values", []objects.Object{}},
		{"values wrong arg", "values", []objects.Object{Int(1)}},
		{"entries no args", "entries", []objects.Object{}},
		{"entries wrong arg", "entries", []objects.Object{Int(1)}},
		{"fromEntries no args", "fromEntries", []objects.Object{}},
		{"fromEntries wrong arg", "fromEntries", []objects.Object{Int(1)}},
		{"type no args", "type", []objects.Object{}},
		{"isPrimitive no args", "isPrimitive", []objects.Object{}},
		{"isEmpty no args", "isEmpty", []objects.Object{}},
		{"size no args", "size", []objects.Object{}},
		{"size wrong arg", "size", []objects.Object{Int(1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mod.Exports[tt.fnName].(*objects.Builtin)
			result := fn.Fn(tt.args...)
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("expected Error for %s, got %T", tt.name, result)
			}
		})
	}
}
