// pkg/stdlib/yaml_comprehensive_test.go
package stdlib

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestYAML12Compliance tests YAML 1.2 specific features
func TestYAML12Compliance(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		checkFn func(t *testing.T, obj objects.Object)
	}{
		// YAML 1.2 Directives
		{
			name:    "YAML directive",
			yaml:    "%YAML 1.2\n---\nkey: value",
			wantErr: false,
		},
		{
			name:    "TAG directive",
			yaml:    "%TAG ! tag:example.com,2014:\n---\n!foo bar",
			wantErr: false,
		},

		// Local tags
		{
			name:    "local tag",
			yaml:    "value: !local myvalue",
			wantErr: false,
		},
		{
			name:    "local tag with complex value",
			yaml:    "!config\nkey: value\nnested:\n  a: 1",
			wantErr: false,
		},

		// Multiline double-quoted strings
		{
			name: "multiline double-quoted with continuation",
			yaml: `text: "first \
  second line"`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("text").HashKey()].Value
				expected := "first second line"
				if val.Inspect() != expected {
					t.Errorf("Expected %q, got %q", expected, val.Inspect())
				}
			},
		},

		// Set type
		{
			name: "set notation",
			yaml: `!!set
? a
? b
? c`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				// Set should have all keys with null values
				m, ok := obj.(*objects.Map)
				if !ok {
					t.Error("Set should parse as Map")
					return
				}
				if len(m.Pairs) != 3 {
					t.Errorf("Expected 3 items in set, got %d", len(m.Pairs))
				}
			},
		},

		// Ordered map
		{
			name: "omap notation",
			yaml: `!!omap
- key1: value1
- key2: value2
- key3: value3`,
			wantErr: false,
		},

		// Complex merge scenarios
		{
			name: "nested merge",
			yaml: `defaults: &defaults
  adapter: postgres
  pool: 5
development:
  <<: *defaults
  database: dev
  pool: 10`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				devKey := objects.NewString("development")
				dev := m.Pairs[devKey.HashKey()].Value.(*objects.Map)
				poolKey := objects.NewString("pool")
				pool := dev.Pairs[poolKey.HashKey()].Value
				if pool.Inspect() != "10" {
					t.Errorf("Expected pool to be 10 (overridden), got %s", pool.Inspect())
				}
			},
		},

		// Empty sequences in various contexts
		{
			name:    "empty sequence as value",
			yaml:    "items: []",
			wantErr: false,
		},
		{
			name:    "nested empty sequences",
			yaml:    "matrix: [[]]",
			wantErr: false,
		},

		// Empty mappings in various contexts
		{
			name:    "empty mapping as value",
			yaml:    "config: {}",
			wantErr: false,
		},
		{
			name:    "nested empty mappings",
			yaml:    "nested: {inner: {}}",
			wantErr: false,
		},

		// Explicit document end
		{
			name: "document with explicit end",
			yaml: `---
key: value
...
---
another: doc`,
			wantErr: false,
		},

		// Unicode in keys and values
		{
			name:    "unicode key and value",
			yaml:    "名字: 你好世界",
			wantErr: false,
		},
		{
			name:    "emoji in yaml",
			yaml:    "emoji: 🎉",
			wantErr: false,
		},

		// Special numeric forms
		{
			name:    "float with underscores (YAML 1.2 JSON schema)",
			yaml:    "value: 1_000_000",
			wantErr: false,
		},

		// Quoted special values
		{
			name:    "quoted null should be string",
			yaml:    `value: "null"`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("value").HashKey()].Value
				if _, ok := val.(*objects.String); !ok {
					t.Errorf("Quoted null should be string, got %T", val)
				}
			},
		},
		{
			name:    "quoted true should be string",
			yaml:    `value: "true"`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("value").HashKey()].Value
				if _, ok := val.(*objects.String); !ok {
					t.Errorf("Quoted true should be string, got %T", val)
				}
			},
		},
		{
			name:    "quoted number should be string",
			yaml:    `value: "123"`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("value").HashKey()].Value
				if _, ok := val.(*objects.String); !ok {
					t.Errorf("Quoted number should be string, got %T", val)
				}
			},
		},

		// Colon in values
		{
			name:    "colon in unquoted value (URL)",
			yaml:    "url: http://example.com",
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("url").HashKey()].Value
				if val.Inspect() != "http://example.com" {
					t.Errorf("Expected URL, got %s", val.Inspect())
				}
			},
		},

		// Block scalar edge cases
		{
			name: "literal block with all empty lines",
			yaml: `text: |



end`,
			wantErr: false,
		},
		{
			name: "folded block with indentation",
			yaml: `text: >
  This is line one.
    This is indented line two.
  This is line three.`,
			wantErr: false,
		},

		// Comments in various positions
		{
			name:    "inline comment after value",
			yaml:    "key: value # this is a comment",
			wantErr: false,
		},
		{
			name:    "comment between items",
			yaml:    "- item1\n# comment\n- item2",
			wantErr: false,
		},
		{
			name:    "comment after flow mapping",
			yaml:    "data: {a: 1} # comment",
			wantErr: false,
		},

		// Complex key with anchor
		{
			name: "complex key with anchor",
			yaml: `? &key
  - item1
  - item2
: value`,
			wantErr: false,
		},

		// Multiple anchors on same structure
		{
			name: "multiple anchors",
			yaml: `first: &a
  x: 1
second: &b
  x: 2
combined:
  <<: [*a, *b]`,
			wantErr: false,
		},

		// Empty key
		{
			name:    "empty string key",
			yaml:    `"": value`,
			wantErr: false,
		},

		// Value on next line after key
		{
			name: "value on next line",
			yaml: `key:
  value`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("key").HashKey()].Value
				if val.Inspect() != "value" {
					t.Errorf("Expected 'value', got %s", val.Inspect())
				}
			},
		},

		// Single quoted multiline
		{
			name: "single quoted multiline",
			yaml: `text: 'line1
line2'`,
			wantErr: false,
		},

		// Dash in key
		{
			name:    "key with dash",
			yaml:    "my-key: value",
			wantErr: false,
		},

		// Dot in key
		{
			name:    "key with dot",
			yaml:    "my.key: value",
			wantErr: false,
		},

		// Special characters in double quoted strings
		{
			name:    "escaped backslash",
			yaml:    `path: "C:\\Users\\name"`,
			wantErr: false,
			checkFn: func(t *testing.T, obj objects.Object) {
				m := obj.(*objects.Map)
				val := m.Pairs[objects.NewString("path").HashKey()].Value
				expected := "C:\\Users\\name"
				if val.Inspect() != expected {
					t.Errorf("Expected %q, got %q", expected, val.Inspect())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, obj)
			}
			t.Logf("Parsed: %s", obj.Inspect())
		})
	}
}

// TestYAMLEdgeCasesMore tests more edge cases
func TestYAMLEdgeCasesMore(t *testing.T) {
	t.Run("deeply nested structure", func(t *testing.T) {
		yaml := `
a:
  b:
    c:
      d:
        e:
          f: deep`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		result := objects.YAMLPathQuery(obj, "a.b.c.d.e.f")
		if result.Inspect() != "deep" {
			t.Errorf("Expected 'deep', got %s", result.Inspect())
		}
	})

	t.Run("mixed sequence and mapping", func(t *testing.T) {
		yaml := `
- name: item1
  value: 1
- - nested1
  - nested2
- key: item2
  other: value`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Mixed: %s", obj.Inspect())
	})

	t.Run("sequence in mapping key position", func(t *testing.T) {
		yaml := `
? - a
  - b
: sequence-as-key`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Sequence key: %s", obj.Inspect())
	})

	t.Run("mapping in mapping key position", func(t *testing.T) {
		yaml := `
? key: value
: mapping-as-key`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Mapping key: %s", obj.Inspect())
	})
}

// TestYAMLRoundTrip tests that parse -> serialize -> parse produces same result
func TestYAMLRoundTrip(t *testing.T) {
	tests := []string{
		"simple: value",
		"number: 42",
		"float: 3.14",
		"bool: true",
		"null: null",
		"list: [1, 2, 3]",
		"map: {a: 1, b: 2}",
		`
nested:
  level1:
    level2: value
`,
		`
- item1
- item2
- item3
`,
	}

	for i, yaml := range tests {
		t.Run("roundtrip_"+string(rune('0'+i)), func(t *testing.T) {
			// First parse
			obj1, err := objects.ParseYAML(yaml)
			if err != nil {
				t.Fatalf("First parse error: %v", err)
			}

			// Serialize
			serialized := objects.SerializeYAML(obj1, 2)

			// Second parse
			obj2, err := objects.ParseYAML(serialized)
			if err != nil {
				t.Fatalf("Second parse error: %v\nSerialized:\n%s", err, serialized)
			}

			// Compare by unmarshaling JSON to handle map key order
			json1, _ := objects.ObjectToJSON(obj1, objects.ObjectToJSONOptions{})
			json2, _ := objects.ObjectToJSON(obj2, objects.ObjectToJSONOptions{})

			var m1, m2 interface{}
			json.Unmarshal(json1, &m1)
			json.Unmarshal(json2, &m2)

			// Use reflect.DeepEqual for comparison
			if !reflect.DeepEqual(m1, m2) {
				t.Errorf("Round-trip mismatch:\nOriginal JSON: %s\nAfter round-trip: %s", json1, json2)
			}
		})
	}
}

// TestYAMLSerializationFeatures tests serialization options
func TestYAMLSerializationFeatures(t *testing.T) {
	t.Run("serialize with different indents", func(t *testing.T) {
		obj, _ := objects.ParseYAML("key: value")

		indent2 := objects.SerializeYAML(obj, 2)
		indent4 := objects.SerializeYAML(obj, 4)

		t.Logf("Indent 2:\n%s", indent2)
		t.Logf("Indent 4:\n%s", indent4)
	})

	t.Run("serialize special floats", func(t *testing.T) {
		tests := []struct {
			name  string
			value float64
		}{
			{"inf", math.Inf(1)},
			{"neg_inf", math.Inf(-1)},
			{"nan", math.NaN()},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				obj := objects.NewFloat(tt.value)
				yaml := objects.SerializeYAML(obj, 2)
				t.Logf("%s -> %s", tt.name, yaml)

				// Parse back
				parsed, err := objects.ParseYAML(yaml)
				if err != nil {
					t.Errorf("Failed to parse %s: %v", yaml, err)
				}
				t.Logf("Parsed back: %s", parsed.Inspect())
			})
		}
	})
}

// TestYAML12AdditionalFeatures tests additional YAML 1.2 features
func TestYAML12AdditionalFeatures(t *testing.T) {
	t.Run("flow style in block context", func(t *testing.T) {
		yaml := `
data:
  items: [a, b, c]
  config: {x: 1, y: 2}
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Result: %s", obj.Inspect())
	})

	t.Run("nested anchors", func(t *testing.T) {
		yaml := `
outer: &outer
  inner: &inner
    value: 42
copy: *inner
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Result: %s", obj.Inspect())
	})

	t.Run("anchor on sequence item", func(t *testing.T) {
		yaml := `
- &item one
- two
- *item
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		arr := obj.(*objects.Array)
		if arr.Elements[0].Inspect() != arr.Elements[2].Inspect() {
			t.Error("Anchored sequence item should be copied correctly")
		}
	})

	t.Run("anchor on mapping value", func(t *testing.T) {
		yaml := `
key: &value
  a: 1
  b: 2
copy: *value
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Result: %s", obj.Inspect())
	})

	t.Run("multiple documents with anchors", func(t *testing.T) {
		yaml := `---
data: &data
  value: 1
---
copy: *data
`
		docs, err := objects.ParseYAMLDocuments(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(docs) != 2 {
			t.Errorf("Expected 2 documents, got %d", len(docs))
		}
	})

	t.Run("explicit null value", func(t *testing.T) {
		tests := []string{
			"key: null",
			"key: Null",
			"key: NULL",
			"key: ~",
			"key:",
		}
		for _, yaml := range tests {
			obj, err := objects.ParseYAML(yaml)
			if err != nil {
				t.Errorf("Error parsing %q: %v", yaml, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			if val != objects.NULL {
				t.Errorf("Expected null for %q, got %s", yaml, val.Inspect())
			}
		}
	})

	t.Run("boolean variants", func(t *testing.T) {
		tests := []struct {
			yaml     string
			expected bool
		}{
			{"key: true", true},
			{"key: True", true},
			{"key: TRUE", true},
			{"key: false", false},
			{"key: False", false},
			{"key: FALSE", false},
			{"key: yes", true},
			{"key: Yes", true},
			{"key: YES", true},
			{"key: no", false},
			{"key: No", false},
			{"key: NO", false},
			{"key: on", true},
			{"key: On", true},
			{"key: ON", true},
			{"key: off", false},
			{"key: Off", false},
			{"key: OFF", false},
		}
		for _, tt := range tests {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Errorf("Error parsing %q: %v", tt.yaml, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			b, ok := val.(*objects.Bool)
			if !ok {
				t.Errorf("Expected bool for %q, got %T", tt.yaml, val)
				continue
			}
			if b.Value != tt.expected {
				t.Errorf("Expected %v for %q, got %v", tt.expected, tt.yaml, b.Value)
			}
		}
	})

	t.Run("integer formats", func(t *testing.T) {
		tests := []struct {
			yaml     string
			expected int64
		}{
			{"key: 42", 42},
			{"key: +42", 42},
			{"key: -42", -42},
			{"key: 0x2A", 42},
			{"key: 0x2a", 42},
			{"key: 0o52", 42},
			{"key: 0b101010", 42},
			{"key: 1_000", 1000},
		}
		for _, tt := range tests {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Errorf("Error parsing %q: %v", tt.yaml, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			i, ok := val.(*objects.Int)
			if !ok {
				t.Errorf("Expected int for %q, got %T (%s)", tt.yaml, val, val.Inspect())
				continue
			}
			if i.Value != tt.expected {
				t.Errorf("Expected %d for %q, got %d", tt.expected, tt.yaml, i.Value)
			}
		}
	})

	t.Run("float formats", func(t *testing.T) {
		tests := []struct {
			yaml    string
			checkFn func(float64) bool
		}{
			{"key: 3.14", func(f float64) bool { return f == 3.14 }},
			{"key: -3.14", func(f float64) bool { return f == -3.14 }},
			{"key: 3.0", func(f float64) bool { return f == 3.0 }},
			{"key: 1e10", func(f float64) bool { return f == 1e10 }},
			{"key: 1E10", func(f float64) bool { return f == 1e10 }},
			{"key: 1.5e+10", func(f float64) bool { return f == 1.5e10 }},
			{"key: 1.5e-10", func(f float64) bool { return f == 1.5e-10 }},
			{"key: .inf", func(f float64) bool { return math.IsInf(f, 1) }},
			{"key: -.inf", func(f float64) bool { return math.IsInf(f, -1) }},
			{"key: .nan", func(f float64) bool { return math.IsNaN(f) }},
		}
		for _, tt := range tests {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Errorf("Error parsing %q: %v", tt.yaml, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			f, ok := val.(*objects.Float)
			if !ok {
				t.Errorf("Expected float for %q, got %T (%s)", tt.yaml, val, val.Inspect())
				continue
			}
			if !tt.checkFn(f.Value) {
				t.Errorf("Check failed for %q, got %v", tt.yaml, f.Value)
			}
		}
	})

	t.Run("block scalar with content", func(t *testing.T) {
		yaml := `
literal: |
  line1
  line2
folded: >
  word1
  word2
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		m := obj.(*objects.Map)

		lit := m.Pairs[objects.NewString("literal").HashKey()].Value.(*objects.String).Value
		if !strings.Contains(lit, "line1\nline2") {
			t.Errorf("Literal block should contain 'line1\\nline2', got %q", lit)
		}

		fold := m.Pairs[objects.NewString("folded").HashKey()].Value.(*objects.String).Value
		if !strings.Contains(fold, "word1 word2") {
			t.Errorf("Folded block should join lines, got %q", fold)
		}
	})

	t.Run("complex nested structure", func(t *testing.T) {
		yaml := `
users:
  - name: alice
    roles:
      - admin
      - user
    settings:
      theme: dark
      notifications: true
  - name: bob
    roles:
      - user
    settings:
      theme: light
      notifications: false
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		t.Logf("Complex: %s", obj.Inspect())

		users := obj.(*objects.Map).Pairs[objects.NewString("users").HashKey()].Value.(*objects.Array)
		if len(users.Elements) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users.Elements))
		}
	})
}

// TestYAMLErrorCases tests various error conditions
func TestYAMLErrorCases(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{"unclosed bracket", "data: [1, 2"},
		{"unclosed brace", "data: {a: 1"},
		{"unclosed double quote", `key: "unclosed`},
		{"unclosed single quote", `key: 'unclosed`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := objects.ParseYAML(tt.yaml)
			if err == nil {
				t.Error("Expected error, got nil")
			} else {
				t.Logf("Error: %v", err)
			}
		})
	}
}
