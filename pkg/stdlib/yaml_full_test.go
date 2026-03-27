// pkg/stdlib/yaml_full_test.go
// Full YAML 1.2 specification compliance tests
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestYAMLFullSpecification tests all YAML 1.2 features
func TestYAMLFullSpecification(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		comment string
	}{
		// ==================== Basic Types ====================
		{
			name:    "string unquoted",
			yaml:    "key: hello world",
			comment: "Plain style string",
		},
		{
			name:    "string double quoted",
			yaml:    `key: "hello world"`,
			comment: "Double-quoted string",
		},
		{
			name:    "string single quoted",
			yaml:    `key: 'hello world'`,
			comment: "Single-quoted string",
		},
		{
			name:    "integer decimal",
			yaml:    "key: 12345",
			comment: "Decimal integer",
		},
		{
			name:    "integer negative",
			yaml:    "key: -12345",
			comment: "Negative integer",
		},
		{
			name:    "integer positive sign",
			yaml:    "key: +12345",
			comment: "Positive sign integer",
		},
		{
			name:    "integer hex",
			yaml:    "key: 0x1A2B",
			comment: "Hexadecimal integer",
		},
		{
			name:    "integer octal",
			yaml:    "key: 0o755",
			comment: "Octal integer",
		},
		{
			name:    "integer binary",
			yaml:    "key: 0b101010",
			comment: "Binary integer",
		},
		{
			name:    "integer with underscores",
			yaml:    "key: 1_000_000",
			comment: "Integer with underscore separators",
		},
		{
			name:    "float decimal",
			yaml:    "key: 3.14159",
			comment: "Decimal float",
		},
		{
			name:    "float exponent",
			yaml:    "key: 1.5e+10",
			comment: "Float with exponent",
		},
		{
			name:    "float infinity",
			yaml:    "key: .inf",
			comment: "Positive infinity",
		},
		{
			name:    "float negative infinity",
			yaml:    "key: -.inf",
			comment: "Negative infinity",
		},
		{
			name:    "float nan",
			yaml:    "key: .nan",
			comment: "Not a number",
		},
		{
			name:    "boolean true",
			yaml:    "key: true",
			comment: "Boolean true",
		},
		{
			name:    "boolean false",
			yaml:    "key: false",
			comment: "Boolean false",
		},
		{
			name:    "null explicit",
			yaml:    "key: null",
			comment: "Explicit null",
		},
		{
			name:    "null tilde",
			yaml:    "key: ~",
			comment: "Tilde null",
		},
		{
			name:    "null empty",
			yaml:    "key:",
			comment: "Empty value null",
		},

		// ==================== Collections ====================
		{
			name: "sequence block",
			yaml: `- one
- two
- three`,
			comment: "Block sequence",
		},
		{
			name:    "sequence flow",
			yaml:    "key: [one, two, three]",
			comment: "Flow sequence",
		},
		{
			name: "mapping block",
			yaml: `one: 1
two: 2
three: 3`,
			comment: "Block mapping",
		},
		{
			name:    "mapping flow",
			yaml:    "key: {one: 1, two: 2}",
			comment: "Flow mapping",
		},
		{
			name: "nested collections",
			yaml: `users:
  - name: alice
    roles:
      - admin
      - user
  - name: bob
    roles:
      - user`,
			comment: "Deeply nested collections",
		},

		// ==================== Block Scalars ====================
		{
			name: "literal block",
			yaml: `key: |
  line 1
  line 2
  line 3`,
			comment: "Literal block scalar",
		},
		{
			name: "literal block strip",
			yaml: `key: |-
  line 1
  line 2`,
			comment: "Literal block with strip chomping",
		},
		{
			name: "literal block keep",
			yaml: `key: |+
  line 1
  line 2

`,
			comment: "Literal block with keep chomping",
		},
		{
			name: "folded block",
			yaml: `key: >
  This is a long
  paragraph that should
  be folded.`,
			comment: "Folded block scalar",
		},
		{
			name: "folded block strip",
			yaml: `key: >-
  line 1
  line 2`,
			comment: "Folded block with strip chomping",
		},
		{
			name: "literal with indent indicator",
			yaml: `key: |2
    line 1
    line 2`,
			comment: "Literal block with explicit indent",
		},

		// ==================== Anchors and Aliases ====================
		{
			name: "simple anchor",
			yaml: `anchor: &anchor value
alias: *anchor`,
			comment: "Simple anchor and alias",
		},
		{
			name: "anchor on mapping",
			yaml: `defaults: &defaults
  adapter: postgres
  host: localhost
development:
  <<: *defaults
  database: dev`,
			comment: "Anchor on mapping with merge",
		},
		{
			name: "anchor on sequence",
			yaml: `- &item one
- two
- *item`,
			comment: "Anchor on sequence item",
		},
		{
			name: "merge key array",
			yaml: `a: &a
  x: 1
b: &b
  y: 2
result:
  <<: [*a, *b]`,
			comment: "Merge key with array of aliases",
		},

		// ==================== Tags ====================
		{
			name:    "tag str",
			yaml:    `key: !!str 123`,
			comment: "!!str tag",
		},
		{
			name:    "tag int",
			yaml:    `key: !!int "456"`,
			comment: "!!int tag",
		},
		{
			name:    "tag float",
			yaml:    `key: !!float 3`,
			comment: "!!float tag",
		},
		{
			name:    "tag bool",
			yaml:    `key: !!bool "yes"`,
			comment: "!!bool tag",
		},
		{
			name:    "tag null",
			yaml:    `key: !!null value`,
			comment: "!!null tag",
		},
		{
			name:    "tag timestamp",
			yaml:    `key: !!timestamp 2024-01-15`,
			comment: "!!timestamp tag",
		},
		{
			name:    "tag binary",
			yaml:    `key: !!binary SGVsbG8=`,
			comment: "!!binary tag",
		},
		{
			name: `tag set`,
			yaml: `!!set
? a
? b
? c`,
			comment: "!!set tag",
		},
		{
			name: `tag omap`,
			yaml: `!!omap
- key1: value1
- key2: value2`,
			comment: "!!omap tag",
		},
		{
			name:    "local tag",
			yaml:    `key: !local value`,
			comment: "Local tag",
		},

		// ==================== Complex Keys ====================
		{
			name: "explicit key",
			yaml: `? key
: value`,
			comment: "Explicit key notation",
		},
		{
			name: "sequence as key",
			yaml: `? - a
  - b
: value`,
			comment: "Sequence as mapping key",
		},
		{
			name: "mapping as key",
			yaml: `? key: value
: result`,
			comment: "Mapping as mapping key",
		},
		{
			name:    "empty key",
			yaml:    `"": value`,
			comment: "Empty string key",
		},

		// ==================== Multi-document ====================
		{
			name: "multiple documents",
			yaml: `---
doc: 1
---
doc: 2
---
doc: 3`,
			comment: "Multiple documents",
		},
		{
			name: "document with explicit end",
			yaml: `---
doc: 1
...
---
doc: 2`,
			comment: "Document with explicit end marker",
		},

		// ==================== Directives ====================
		{
			name:    "YAML directive",
			yaml:    "%YAML 1.2\n---\nkey: value",
			comment: "YAML directive",
		},
		{
			name:    "TAG directive",
			yaml:    "%TAG ! tag:example.com,2014:\n---\n!foo bar",
			comment: "TAG directive",
		},

		// ==================== Comments ====================
		{
			name:    "inline comment",
			yaml:    "key: value # comment",
			comment: "Inline comment",
		},
		{
			name: "comment between items",
			yaml: `- one
# comment
- two`,
			comment: "Comment between sequence items",
		},
		{
			name:    "comment only document",
			yaml:    "# just a comment",
			comment: "Comment only document",
		},

		// ==================== Escape Sequences ====================
		{
			name:    "escape newline",
			yaml:    `key: "line1\nline2"`,
			comment: "Escaped newline",
		},
		{
			name:    "escape tab",
			yaml:    `key: "col1\tcol2"`,
			comment: "Escaped tab",
		},
		{
			name:    "escape unicode",
			yaml:    `key: "\u0041"`,
			comment: "Unicode escape",
		},
		{
			name:    "escape hex",
			yaml:    `key: "\x41"`,
			comment: "Hex escape",
		},

		// ==================== Multiline Strings ====================
		{
			name: "multiline double quoted continuation",
			yaml: `key: "first \
  second"`,
			comment: "Multiline double-quoted with backslash continuation",
		},
		{
			name: `multiline single quoted`,
			yaml: `key: 'line1
line2'`,
			comment: "Multiline single-quoted",
		},
		{
			name:    `single quoted escaped quote`,
			yaml:    `key: 'it''s working'`,
			comment: "Single-quoted with escaped quote",
		},

		// ==================== Edge Cases ====================
		{
			name:    "empty document",
			yaml:    "",
			comment: "Empty document",
		},
		{
			name:    "empty sequence",
			yaml:    "key: []",
			comment: "Empty flow sequence",
		},
		{
			name:    "empty mapping",
			yaml:    "key: {}",
			comment: "Empty flow mapping",
		},
		{
			name:    "nested empty",
			yaml:    "key: [[]]",
			comment: "Nested empty sequence",
		},
		{
			name:    "key with special chars",
			yaml:    "my-key.with_special: value",
			comment: "Key with special characters",
		},
		{
			name:    "url as value",
			yaml:    "url: http://example.com:8080/path",
			comment: "URL with colons as value",
		},
		{
			name:    "unicode content",
			yaml:    "名字: 你好世界 🎉",
			comment: "Unicode content",
		},
		{
			name: "multiline flow sequence",
			yaml: `key: [
  1,
  2,
  3
]`,
			comment: "Multiline flow sequence",
		},
		{
			name: "multiline flow mapping",
			yaml: `key: {
  a: 1,
  b: 2
}`,
			comment: "Multiline flow mapping",
		},

		// ==================== Sexagesimal ====================
		{
			name:    "sexagesimal simple",
			yaml:    "key: 1:30",
			comment: "Sexagesimal number",
		},
		{
			name:    "sexagesimal complex",
			yaml:    "key: 1:30:45",
			comment: "Complex sexagesimal",
		},

		// ==================== Timestamps ====================
		{
			name:    "date only",
			yaml:    "key: 2024-01-15",
			comment: "Date only",
		},
		{
			name:    "datetime",
			yaml:    "key: 2024-01-15T10:30:00Z",
			comment: "DateTime with timezone Z",
		},
		{
			name:    "datetime with offset",
			yaml:    "key: 2024-01-15T10:30:00+08:00",
			comment: "DateTime with offset",
		},
	}

	passed := 0
	failed := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if tt.wantErr {
				if err == nil {
					failed++
					t.Errorf("Expected error, got: %s", obj.Inspect())
				} else {
					passed++
					t.Logf("OK: %s", tt.comment)
				}
				return
			}
			if err != nil {
				failed++
				t.Errorf("Unexpected error: %v", err)
				return
			}
			passed++
			t.Logf("OK: %s -> %s", tt.comment, obj.Inspect())
		})
	}

	t.Logf("\n=== Summary ===")
	t.Logf("Passed: %d, Failed: %d", passed, failed)
}

// TestYAMLSpecificationCoverage tests coverage of YAML spec
func TestYAMLSpecificationCoverage(t *testing.T) {
	t.Run("all boolean representations", func(t *testing.T) {
		trueValues := []string{"true", "True", "TRUE", "yes", "Yes", "YES", "on", "On", "ON"}
		falseValues := []string{"false", "False", "FALSE", "no", "No", "NO", "off", "Off", "OFF"}

		for _, v := range trueValues {
			obj, err := objects.ParseYAML("key: " + v)
			if err != nil {
				t.Errorf("Error parsing '%s': %v", v, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			if b, ok := val.(*objects.Bool); !ok || !b.Value {
				t.Errorf("'%s' should be true, got %s", v, val.Inspect())
			}
		}

		for _, v := range falseValues {
			obj, err := objects.ParseYAML("key: " + v)
			if err != nil {
				t.Errorf("Error parsing '%s': %v", v, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			if b, ok := val.(*objects.Bool); !ok || b.Value {
				t.Errorf("'%s' should be false, got %s", v, val.Inspect())
			}
		}
	})

	t.Run("all null representations", func(t *testing.T) {
		nullValues := []string{"null", "Null", "NULL", "~", ""}

		for _, v := range nullValues {
			var yaml string
			if v == "" {
				yaml = "key:"
			} else {
				yaml = "key: " + v
			}
			obj, err := objects.ParseYAML(yaml)
			if err != nil {
				t.Errorf("Error parsing '%s': %v", v, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			if val != objects.NULL {
				t.Errorf("'%s' should be null, got %s", v, val.Inspect())
			}
		}
	})

	t.Run("quoted values stay strings", func(t *testing.T) {
		tests := []string{`"null"`, `"true"`, `"false"`, `"123"`, `"1.5"`}
		for _, v := range tests {
			yaml := "key: " + v
			obj, err := objects.ParseYAML(yaml)
			if err != nil {
				t.Errorf("Error parsing %s: %v", yaml, err)
				continue
			}
			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("key").HashKey()].Value
			if _, ok := val.(*objects.String); !ok {
				t.Errorf("Quoted %s should be string, got %T", v, val)
			}
		}
	})
}
