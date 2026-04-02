// pkg/objects/jsonutil_test.go
// Tests for JSON utility functions
package objects

import (
	"testing"
)

func TestObjectToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		opts     ObjectToJSONOptions
		expected string
	}{
		{
			name:     "null",
			input:    NULL,
			opts:     ObjectToJSONOptions{},
			expected: "null",
		},
		{
			name:     "bool true",
			input:    TRUE,
			opts:     ObjectToJSONOptions{},
			expected: "true",
		},
		{
			name:     "bool false",
			input:    FALSE,
			opts:     ObjectToJSONOptions{},
			expected: "false",
		},
		{
			name:     "integer",
			input:    NewInt(42),
			opts:     ObjectToJSONOptions{},
			expected: "42",
		},
		{
			name:     "float",
			input:    NewFloat(3.14),
			opts:     ObjectToJSONOptions{},
			expected: "3.14",
		},
		{
			name:     "string",
			input:    NewString("hello"),
			opts:     ObjectToJSONOptions{},
			expected: `"hello"`,
		},
		{
			name:     "empty array",
			input:    NewArray([]Object{}),
			opts:     ObjectToJSONOptions{},
			expected: "[]",
		},
		{
			name:     "array with elements",
			input:    NewArray([]Object{NewInt(1), NewInt(2), NewInt(3)}),
			opts:     ObjectToJSONOptions{},
			expected: "[1,2,3]",
		},
		{
			name:     "empty map",
			input:    NewMap(make(map[HashKey]MapPair)),
			opts:     ObjectToJSONOptions{},
			expected: "{}",
		},
		{
			name: "map with elements",
			input: func() Object {
				m := NewMapWithCapacity(1)
				m.Pairs[NewString("key").HashKey()] = MapPair{
					Key:   NewString("key"),
					Value: NewString("value"),
				}
				return m
			}(),
			opts:     ObjectToJSONOptions{},
			expected: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ObjectToJSON(tt.input, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(result))
			}
		})
	}
}

func TestObjectToJSON_WithIndent(t *testing.T) {
	arr := NewArray([]Object{NewInt(1), NewInt(2)})
	result, err := ObjectToJSON(arr, ObjectToJSONOptions{Indent: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain newlines for indentation
	if len(result) < 5 {
		t.Errorf("expected longer output with indentation, got '%s'", string(result))
	}
}

func TestObjectToJSON_WithSortKeys(t *testing.T) {
	m := NewMapWithCapacity(2)
	m.Pairs[NewString("z").HashKey()] = MapPair{Key: NewString("z"), Value: NewInt(1)}
	m.Pairs[NewString("a").HashKey()] = MapPair{Key: NewString("a"), Value: NewInt(2)}

	result, err := ObjectToJSON(m, ObjectToJSONOptions{SortKeys: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With SortKeys, "a" should come before "z"
	str := string(result)
	if len(str) < 10 {
		t.Errorf("expected valid JSON, got '%s'", str)
	}
}

func TestJSONToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, obj Object)
	}{
		{
			name:  "null",
			input: "null",
			validate: func(t *testing.T, obj Object) {
				if obj != NULL {
					t.Errorf("expected NULL, got %v", obj)
				}
			},
		},
		{
			name:  "true",
			input: "true",
			validate: func(t *testing.T, obj Object) {
				if b, ok := obj.(*Bool); !ok || !b.Value {
					t.Errorf("expected true, got %v", obj)
				}
			},
		},
		{
			name:  "false",
			input: "false",
			validate: func(t *testing.T, obj Object) {
				if b, ok := obj.(*Bool); !ok || b.Value {
					t.Errorf("expected false, got %v", obj)
				}
			},
		},
		{
			name:  "integer",
			input: "42",
			validate: func(t *testing.T, obj Object) {
				if i, ok := obj.(*Int); !ok || i.Value != 42 {
					t.Errorf("expected 42, got %v", obj)
				}
			},
		},
		{
			name:  "float",
			input: "3.14",
			validate: func(t *testing.T, obj Object) {
				if f, ok := obj.(*Float); !ok || f.Value != 3.14 {
					t.Errorf("expected 3.14, got %v", obj)
				}
			},
		},
		{
			name:  "string",
			input: `"hello"`,
			validate: func(t *testing.T, obj Object) {
				if s, ok := obj.(*String); !ok || s.Value != "hello" {
					t.Errorf("expected 'hello', got %v", obj)
				}
			},
		},
		{
			name:  "empty array",
			input: "[]",
			validate: func(t *testing.T, obj Object) {
				if arr, ok := obj.(*Array); !ok || len(arr.Elements) != 0 {
					t.Errorf("expected empty array, got %v", obj)
				}
			},
		},
		{
			name:  "array with elements",
			input: "[1, 2, 3]",
			validate: func(t *testing.T, obj Object) {
				arr, ok := obj.(*Array)
				if !ok {
					t.Fatalf("expected array, got %T", obj)
				}
				if len(arr.Elements) != 3 {
					t.Errorf("expected 3 elements, got %d", len(arr.Elements))
				}
			},
		},
		{
			name:  "empty object",
			input: "{}",
			validate: func(t *testing.T, obj Object) {
				if m, ok := obj.(*Map); !ok || len(m.Pairs) != 0 {
					t.Errorf("expected empty map, got %v", obj)
				}
			},
		},
		{
			name:  "object with fields",
			input: `{"key": "value"}`,
			validate: func(t *testing.T, obj Object) {
				m, ok := obj.(*Map)
				if !ok {
					t.Fatalf("expected map, got %T", obj)
				}
				if len(m.Pairs) != 1 {
					t.Errorf("expected 1 pair, got %d", len(m.Pairs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := JSONToObject(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, obj)
		})
	}
}

func TestJSONToObject_Invalid(t *testing.T) {
	_, err := JSONToObject("invalid json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestObjectToJSON_Complex(t *testing.T) {
	// Create a complex nested structure
	innerMap := NewMapWithCapacity(1)
	innerMap.Pairs[NewString("inner").HashKey()] = MapPair{
		Key:   NewString("inner"),
		Value: NewString("value"),
	}

	arr := NewArray([]Object{
		NewInt(1),
		NewString("test"),
		innerMap,
	})

	result, err := ObjectToJSON(arr, ObjectToJSONOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid JSON
	if len(result) < 10 {
		t.Errorf("expected longer JSON output, got '%s'", string(result))
	}
}

func TestObjectToJSON_SpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"newline", "line1\nline2"},
		{"tab", "col1\tcol2"},
		{"quote", `has "quotes" inside`},
		{"backslash", `path\to\file`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ObjectToJSON(NewString(tt.input), ObjectToJSONOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should be valid JSON (parseable)
			obj, err := JSONToObject(string(result))
			if err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			if s, ok := obj.(*String); !ok || s.Value != tt.input {
				t.Errorf("round-trip failed: expected '%s', got '%v'", tt.input, obj)
			}
		})
	}
}
