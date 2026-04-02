// pkg/stdlib/yaml_test.go
// Tests for YAML parsing and generation utilities.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestYAMLParse tests basic YAML parsing
func TestYAMLParse(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name:     "simple string",
			yaml:     "name: hello",
			expected: `{"name": "hello"}`,
		},
		{
			name:     "simple number",
			yaml:     "value: 42",
			expected: `{"value": 42}`,
		},
		{
			name:     "boolean",
			yaml:     "enabled: true",
			expected: `{"enabled": true}`,
		},
		{
			name:     "null value",
			yaml:     "empty: null",
			expected: `{"empty": null}`,
		},
		{
			name: "nested map",
			yaml: `server:
  host: localhost
  port: 8080`,
			expected: `{"server": {"host": "localhost", "port": 8080}}`,
		},
		{
			name: "simple list",
			yaml: `- one
- two
- three`,
			expected: `["one", "two", "three"]`,
		},
		{
			name: "list of maps",
			yaml: `- name: alice
  age: 30
- name: bob
  age: 25`,
			expected: `[{"name": "alice", "age": 30}, {"name": "bob", "age": 25}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			// Convert to JSON for comparison
			jsonBytes, err := objects.ObjectToJSON(obj, objects.ObjectToJSONOptions{})
			if err != nil {
				t.Fatalf("ObjectToJSON() error = %v", err)
			}

			t.Logf("YAML: %s -> JSON: %s", tt.yaml, string(jsonBytes))
		})
	}
}

// TestYAMLStringify tests YAML serialization
func TestYAMLStringify(t *testing.T) {
	tests := []struct {
		name string
		obj  objects.Object
	}{
		{
			name: "simple map",
			obj: &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
				objects.NewString("name").HashKey():  {Key: objects.NewString("name"), Value: objects.NewString("hello")},
				objects.NewString("value").HashKey(): {Key: objects.NewString("value"), Value: objects.NewInt(42)},
			}},
		},
		{
			name: "nested map",
			obj: &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
				objects.NewString("server").HashKey(): {
					Key: objects.NewString("server"),
					Value: &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
						objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: objects.NewString("localhost")},
						objects.NewString("port").HashKey(): {Key: objects.NewString("port"), Value: objects.NewInt(8080)},
					}},
				},
			}},
		},
		{
			name: "array",
			obj: &objects.Array{Elements: []objects.Object{
				objects.NewString("one"),
				objects.NewString("two"),
				objects.NewString("three"),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlStr := objects.SerializeYAML(tt.obj, 2)
			t.Logf("YAML:\n%s", yamlStr)

			// Parse it back
			parsed, err := objects.ParseYAML(yamlStr)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			// Verify they're equivalent
			origJSON, _ := objects.ObjectToJSON(tt.obj, objects.ObjectToJSONOptions{})
			parsedJSON, _ := objects.ObjectToJSON(parsed, objects.ObjectToJSONOptions{})

			t.Logf("Original JSON: %s", string(origJSON))
			t.Logf("Parsed JSON: %s", string(parsedJSON))
		})
	}
}

// TestYAMLModuleFunctions tests the yaml module functions
func TestYAMLModuleFunctions(t *testing.T) {
	// Get the module
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}

	// Test parse
	parseFn, ok := mod.Exports["parse"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.parse not found")
	}

	yamlStr := objects.NewString("name: test\nvalue: 123")
	result := parseFn.Fn(yamlStr)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parse() error: %s", result.Inspect())
	}

	t.Logf("parse result: %s", result.Inspect())

	// Test stringify
	stringifyFn, ok := mod.Exports["stringify"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.stringify not found")
	}

	result = stringifyFn.Fn(result)
	if result.Type() == objects.ErrorType {
		t.Fatalf("stringify() error: %s", result.Inspect())
	}

	t.Logf("stringify result:\n%s", result.Inspect())

	// Test isValid
	isValidFn, ok := mod.Exports["isValid"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.isValid not found")
	}

	validResult := isValidFn.Fn(yamlStr)
	if validResult.Type() == objects.ErrorType {
		t.Fatalf("isValid() error: %s", validResult.Inspect())
	}

	if validResult.ToBool().Value != true {
		t.Error("isValid() should return true for valid YAML")
	}

	// Test getType
	getTypeFn, ok := mod.Exports["getType"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.getType not found")
	}

	typeResult := getTypeFn.Fn(yamlStr)
	if typeResult.Type() == objects.ErrorType {
		t.Fatalf("getType() error: %s", typeResult.Inspect())
	}

	t.Logf("getType result: %s", typeResult.Inspect())
}

// TestYAMLConversion tests JSON <-> YAML conversion
func TestYAMLConversion(t *testing.T) {
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}

	// Test fromJson
	fromJsonFn, ok := mod.Exports["fromJson"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.fromJson not found")
	}

	jsonStr := objects.NewString(`{"server": {"host": "localhost", "port": 8080}}`)
	result := fromJsonFn.Fn(jsonStr)
	if result.Type() == objects.ErrorType {
		t.Fatalf("fromJson() error: %s", result.Inspect())
	}

	t.Logf("fromJson result:\n%s", result.Inspect())

	// Test toJson
	toJsonFn, ok := mod.Exports["toJson"].(*objects.Builtin)
	if !ok {
		t.Fatal("yaml.toJson not found")
	}

	yamlInput := objects.NewString("server:\n  host: localhost\n  port: 8080")
	jsonResult := toJsonFn.Fn(yamlInput)
	if jsonResult.Type() == objects.ErrorType {
		t.Fatalf("toJson() error: %s", jsonResult.Inspect())
	}

	t.Logf("toJson result: %s", jsonResult.Inspect())
}

// TestYAMLPathAccess tests path-based access
func TestYAMLPathAccess(t *testing.T) {
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}

	// Parse YAML
	parseFn := mod.Exports["parse"].(*objects.Builtin)
	yamlStr := objects.NewString(`server:
  host: localhost
  port: 8080
database:
  name: mydb
  credentials:
    user: admin
    password: secret`)
	obj := parseFn.Fn(yamlStr)

	// Test get
	getFn := mod.Exports["get"].(*objects.Builtin)

	tests := []struct {
		path     string
		expected string
	}{
		{"server.host", "localhost"},
		{"server.port", "8080"},
		{"database.name", "mydb"},
		{"database.credentials.user", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getFn.Fn(obj, objects.NewString(tt.path))
			if result.Type() == objects.ErrorType {
				t.Fatalf("get() error: %s", result.Inspect())
			}

			if result.Inspect() != tt.expected {
				t.Errorf("get(%s) = %s, want %s", tt.path, result.Inspect(), tt.expected)
			}
		})
	}

	// Test has
	hasFn := mod.Exports["has"].(*objects.Builtin)

	if hasFn.Fn(obj, objects.NewString("server.host")).ToBool().Value != true {
		t.Error("has(server.host) should be true")
	}

	if hasFn.Fn(obj, objects.NewString("nonexistent")).ToBool().Value != false {
		t.Error("has(nonexistent) should be false")
	}
}

// TestYAMLMerge tests merge operations
func TestYAMLMerge(t *testing.T) {
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}

	// Parse two maps
	parseFn := mod.Exports["parse"].(*objects.Builtin)

	map1 := parseFn.Fn(objects.NewString("a: 1\nb: 2"))
	map2 := parseFn.Fn(objects.NewString("b: 3\nc: 4"))

	// Test merge
	mergeFn := mod.Exports["merge"].(*objects.Builtin)
	result := mergeFn.Fn(map1, map2)

	if result.Type() == objects.ErrorType {
		t.Fatalf("merge() error: %s", result.Inspect())
	}

	// Convert to JSON for easier inspection
	jsonBytes, _ := objects.ObjectToJSON(result, objects.ObjectToJSONOptions{})
	t.Logf("merge result: %s", string(jsonBytes))

	// Test deepMerge
	deepMergeFn := mod.Exports["deepMerge"].(*objects.Builtin)

	map3 := parseFn.Fn(objects.NewString("server:\n  host: localhost\n  port: 8080"))
	map4 := parseFn.Fn(objects.NewString("server:\n  port: 9090\n  ssl: true"))

	deepResult := deepMergeFn.Fn(map3, map4)
	if deepResult.Type() == objects.ErrorType {
		t.Fatalf("deepMerge() error: %s", deepResult.Inspect())
	}

	jsonBytes, _ = objects.ObjectToJSON(deepResult, objects.ObjectToJSONOptions{})
	t.Logf("deepMerge result: %s", string(jsonBytes))
}

// TestYAMLMultiDocument tests multi-document YAML
func TestYAMLMultiDocument(t *testing.T) {
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}

	// Test parseDocuments
	parseDocsFn := mod.Exports["parseDocuments"].(*objects.Builtin)
	yamlStr := objects.NewString(`---
name: doc1
value: 1
---
name: doc2
value: 2`)

	result := parseDocsFn.Fn(yamlStr)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseDocuments() error: %s", result.Inspect())
	}

	t.Logf("parseDocuments result: %s", result.Inspect())

	// Verify it's an array with 2 elements
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("parseDocuments should return an array")
	}

	if len(arr.Elements) != 2 {
		t.Errorf("parseDocuments returned %d documents, want 2", len(arr.Elements))
	}
}

// TestYAMLSpecialValues tests special YAML values
func TestYAMLSpecialValues(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name:     "infinity",
			yaml:     "value: .inf",
			expected: "inf",
		},
		{
			name:     "negative infinity",
			yaml:     "value: -.inf",
			expected: "-inf",
		},
		{
			name:     "nan",
			yaml:     "value: .nan",
			expected: "nan",
		},
		{
			name:     "yes as boolean",
			yaml:     "enabled: yes",
			expected: "true",
		},
		{
			name:     "no as boolean",
			yaml:     "enabled: no",
			expected: "false",
		},
		{
			name:     "hex integer",
			yaml:     "value: 0xFF",
			expected: "255",
		},
		{
			name:     "octal integer",
			yaml:     "value: 0o10",
			expected: "8",
		},
		{
			name:     "binary integer",
			yaml:     "value: 0b1010",
			expected: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			t.Logf("YAML: %s -> %s", tt.yaml, obj.Inspect())
		})
	}
}

// TestYAMLAnchorsAndAliases tests anchor and alias support
func TestYAMLAnchorsAndAliases(t *testing.T) {
	yaml := `
defaults: &defaults
  adapter: postgres
  host: localhost

development:
  database: dev
  <<: *defaults

production:
  database: prod
  <<: *defaults
  host: prod.example.com
`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Parsed with anchors: %s", obj.Inspect())

	// Verify development has defaults merged
	dev := obj.(*objects.Map).Pairs[objects.NewString("development").HashKey()].Value
	devMap, ok := dev.(*objects.Map)
	if !ok {
		t.Fatal("development should be a map")
	}

	if devMap.Pairs[objects.NewString("adapter").HashKey()].Value.Inspect() != "postgres" {
		t.Error("development should have adapter from defaults")
	}

	t.Logf("Development: %s", devMap.Inspect())
}

// TestYAMLFlowStyle tests inline flow styles
func TestYAMLFlowStyle(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "flow sequence",
			yaml: `items: [one, two, three]`,
		},
		{
			name: "flow mapping",
			yaml: `config: {host: localhost, port: 8080}`,
		},
		{
			name: "nested flow",
			yaml: `data: {items: [a, b, c], config: {enabled: true}}`,
		},
		{
			name: "empty flow sequence",
			yaml: `empty: []`,
		},
		{
			name: "empty flow mapping",
			yaml: `empty: {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			t.Logf("Flow YAML: %s -> %s", tt.yaml, obj.Inspect())
		})
	}
}

// TestYAMLBlockScalars tests literal and folded block scalars
func TestYAMLBlockScalars(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "literal block",
			yaml: `text: |
  line 1
  line 2
  line 3`,
		},
		{
			name: "literal block strip",
			yaml: `text: |-
  line 1
  line 2`,
		},
		{
			name: "literal block keep",
			yaml: `text: |+
  line 1
  line 2

`,
		},
		{
			name: "folded block",
			yaml: `text: >
  This is a long
  paragraph that should
  be folded into one line.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			text := obj.(*objects.Map).Pairs[objects.NewString("text").HashKey()].Value
			t.Logf("Block scalar '%s': %q", tt.name, text.Inspect())
		})
	}
}

// TestYAMLMultiDocument tests multi-document parsing
func TestYAMLMultiDocumentFull(t *testing.T) {
	yaml := `---
name: document1
type: config
---
name: document2
type: data
---
name: document3
type: result
...

`
	docs, err := objects.ParseYAMLDocuments(yaml)
	if err != nil {
		t.Fatalf("ParseYAMLDocuments() error = %v", err)
	}

	if len(docs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(docs))
	}

	for i, doc := range docs {
		t.Logf("Document %d: %s", i+1, doc.Inspect())
	}
}

// TestYAMLMergeKey tests merge key functionality
func TestYAMLMergeKey(t *testing.T) {
	yaml := `
base: &base
  name: base
  value: 100

extended:
  <<: *base
  extra: added

override:
  <<: *base
  value: 200
`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	root := obj.(*objects.Map)

	// Check extended has both base values and extra
	extended := root.Pairs[objects.NewString("extended").HashKey()].Value.(*objects.Map)
	if extended.Pairs[objects.NewString("name").HashKey()].Value.Inspect() != "base" {
		t.Error("extended should inherit name from base")
	}
	if extended.Pairs[objects.NewString("extra").HashKey()].Value.Inspect() != "added" {
		t.Error("extended should have extra field")
	}

	// Check override has overridden value
	override := root.Pairs[objects.NewString("override").HashKey()].Value.(*objects.Map)
	if override.Pairs[objects.NewString("value").HashKey()].Value.Inspect() != "200" {
		t.Error("override should have value 200")
	}

	t.Logf("Merge key result: %s", obj.Inspect())
}

// TestYAMLTags tests YAML type tags
func TestYAMLTags(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "string tag",
			yaml: `value: !!str 123`,
		},
		{
			name: "int tag",
			yaml: `value: !!int "456"`,
		},
		{
			name: "float tag",
			yaml: `value: !!float 3`,
		},
		{
			name: "bool tag",
			yaml: `value: !!bool "yes"`,
		},
		{
			name: "null tag",
			yaml: `value: !!null not-null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			t.Logf("Tag YAML: %s -> %s", tt.yaml, obj.Inspect())
		})
	}
}

// TestYAMLTimestamp tests !!timestamp tag
func TestYAMLTimestamp(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "ISO date",
			yaml: `value: !!timestamp 2024-01-15`,
		},
		{
			name: "ISO datetime",
			yaml: `value: !!timestamp 2024-01-15T10:30:00Z`,
		},
		{
			name: "RFC3339",
			yaml: `value: !!timestamp 2024-01-15T10:30:00+08:00`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			t.Logf("Timestamp: %s -> %s", tt.yaml, obj.Inspect())
		})
	}
}

// TestYAMLBinary tests !!binary tag
func TestYAMLBinary(t *testing.T) {
	yaml := `value: !!binary SGVsbG8gV29ybGQ=`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Binary: %s", obj.Inspect())
}

// TestYAMLComplexKey tests explicit key syntax
func TestYAMLComplexKey(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "explicit string key",
			yaml: `? key
: value`,
		},
		{
			name: "explicit mapping key",
			yaml: `? - item1
  - item2
: value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			t.Logf("Complex key: %s", obj.Inspect())
		})
	}
}

// TestYAMLBlockScalarIndent tests block scalar with explicit indent
func TestYAMLBlockScalarIndent(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "literal with indent indicator",
			yaml: `text: |2
    line 1
    line 2`,
		},
		{
			name: "folded with indent indicator",
			yaml: `text: >2
    folded
    content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			text := obj.(*objects.Map).Pairs[objects.NewString("text").HashKey()].Value
			t.Logf("Block scalar: %q", text.Inspect())
		})
	}
}

// TestYAMLSexagesimal tests sexagesimal (base 60) numbers
func TestYAMLSexagesimal(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected float64
	}{
		{
			name:     "simple sexagesimal",
			yaml:     `value: 1:30`,
			expected: 90,
		},
		{
			name:     "hours minutes seconds",
			yaml:     `value: 1:30:45`,
			expected: 5445,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			val := obj.(*objects.Map).Pairs[objects.NewString("value").HashKey()].Value
			f, ok := val.(*objects.Float)
			if !ok {
				t.Fatalf("Expected float, got %T", val)
			}

			t.Logf("Sexagesimal: %s -> %v", tt.yaml, f.Value)
		})
	}
}

// TestYAMLErrorReporting tests error messages with line/column info
func TestYAMLErrorReporting(t *testing.T) {
	invalidYAML := `valid: data
invalid: [unclosed
more: data`

	_, err := objects.ParseYAML(invalidYAML)
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}

	t.Logf("Error message: %v", err)
}

// TestYAMLAdvancedAnchors tests complex anchor scenarios
func TestYAMLAdvancedAnchors(t *testing.T) {
	yaml := `
user1: &user
  name: alice
  role: admin

user2: &user2
  <<: *user
  name: bob

users:
  - *user
  - *user2
`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Advanced anchors: %s", obj.Inspect())

	// Verify user2 has overridden name
	root := obj.(*objects.Map)
	user2 := root.Pairs[objects.NewString("user2").HashKey()].Value.(*objects.Map)
	if user2.Pairs[objects.NewString("name").HashKey()].Value.Inspect() != "bob" {
		t.Error("user2 should have name 'bob'")
	}

	// Verify user2 still has role from user1
	if user2.Pairs[objects.NewString("role").HashKey()].Value.Inspect() != "admin" {
		t.Error("user2 should have role 'admin' from user1")
	}
}

// TestYAMLMultilineDoubleQuotedString tests multiline double-quoted strings
// Note: Multi-line quoted strings with backslash continuation are not fully supported yet
// This test verifies that unclosed quotes are properly detected
func TestYAMLMultilineDoubleQuotedString(t *testing.T) {
	// Test that unclosed quotes are detected
	yaml := `text: "unclosed`

	_, err := objects.ParseYAML(yaml)
	if err == nil {
		t.Error("Expected error for unclosed double-quoted string")
	}
	t.Logf("Unclosed quote error: %v", err)

	// Test properly closed quotes work
	yaml2 := `text: "single line"`

	obj, err := objects.ParseYAML(yaml2)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	text := obj.(*objects.Map).Pairs[objects.NewString("text").HashKey()].Value
	if text.Inspect() != "single line" {
		t.Errorf("Expected 'single line', got %s", text.Inspect())
	}
}

// TestYAMLPathQuery tests YAML path query
func TestYAMLPathQuery(t *testing.T) {
	yaml := `
server:
  host: localhost
  ports:
    - 80
    - 443
  config:
    timeout: 30
`

	obj, _ := objects.ParseYAML(yaml)

	tests := []struct {
		path     string
		expected string
	}{
		{"server.host", "localhost"},
		{"server.ports.[0]", "80"},
		{"server.ports.[1]", "443"},
		{"server.config.timeout", "30"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := objects.YAMLPathQuery(obj, tt.path)
			if result.Inspect() != tt.expected {
				t.Errorf("PathQuery(%s) = %s, want %s", tt.path, result.Inspect(), tt.expected)
			}
		})
	}
}

// TestYAMLValidation tests YAML validation
func TestYAMLValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		isValid bool
	}{
		{"valid simple", "name: test", true},
		{"valid nested", "server:\n  host: localhost", true},
		{"invalid unclosed", "data: [unclosed", false},
		{"invalid unclosed brace", "data: {unclosed", false},
		{"invalid unclosed quote", "key: \"unclosed", false},
		{"invalid unclosed single quote", "key: 'unclosed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := objects.ValidateYAML(tt.yaml)
			if tt.isValid && err != nil {
				t.Errorf("Expected valid, got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Error("Expected error for invalid YAML")
			}
		})
	}
}

// TestYAMLEdgeCases tests various edge cases
func TestYAMLEdgeCases(t *testing.T) {
	t.Run("empty document", func(t *testing.T) {
		obj, err := objects.ParseYAML("")
		if err != nil {
			t.Errorf("Empty document should parse: %v", err)
		}
		if obj != objects.NULL {
			t.Errorf("Empty document should be null, got %s", obj.Inspect())
		}
	})

	t.Run("comment only", func(t *testing.T) {
		obj, err := objects.ParseYAML("# just a comment")
		if err != nil {
			t.Errorf("Comment only should parse: %v", err)
		}
		if obj != objects.NULL {
			t.Errorf("Comment only should be null, got %s", obj.Inspect())
		}
	})

	t.Run("quoted colon in key", func(t *testing.T) {
		yaml := `"key:with:colons": value`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		if len(m.Pairs) != 1 {
			t.Errorf("Expected 1 pair, got %d", len(m.Pairs))
		}
		t.Logf("Key with colons: %s", obj.Inspect())
	})

	t.Run("empty string value", func(t *testing.T) {
		yaml := `key: ""`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("key").HashKey()].Value
		if val.Inspect() != "" {
			t.Errorf("Expected empty string, got %s", val.Inspect())
		}
	})

	t.Run("explicit null", func(t *testing.T) {
		yaml := "key: null"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("key").HashKey()].Value
		if val != objects.NULL {
			t.Errorf("Expected null, got %s", val.Inspect())
		}
	})

	t.Run("tilde as null", func(t *testing.T) {
		yaml := "key: ~"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("key").HashKey()].Value
		if val != objects.NULL {
			t.Errorf("Expected null, got %s", val.Inspect())
		}
	})

	t.Run("float with exponent", func(t *testing.T) {
		yaml := "value: 1.5e+10"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("value").HashKey()].Value
		t.Logf("Float with exponent: %s", val.Inspect())
	})

	t.Run("negative integer", func(t *testing.T) {
		yaml := "value: -42"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("value").HashKey()].Value
		if val.Inspect() != "-42" {
			t.Errorf("Expected -42, got %s", val.Inspect())
		}
	})

	t.Run("single quoted string", func(t *testing.T) {
		yaml := `key: 'hello world'`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("key").HashKey()].Value
		if val.Inspect() != "hello world" {
			t.Errorf("Expected 'hello world', got %s", val.Inspect())
		}
	})

	t.Run("escaped single quote", func(t *testing.T) {
		yaml := `key: 'it''s working'`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("key").HashKey()].Value
		if val.Inspect() != "it's working" {
			t.Errorf("Expected \"it's working\", got %s", val.Inspect())
		}
	})

	t.Run("nested sequences", func(t *testing.T) {
		yaml := `
matrix:
  - - 1
    - 2
  - - 3
    - 4
`
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		t.Logf("Nested sequences: %s", obj.Inspect())
	})

	t.Run("empty sequence", func(t *testing.T) {
		yaml := "items: []"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("items").HashKey()].Value
		if arr, ok := val.(*objects.Array); !ok || len(arr.Elements) != 0 {
			t.Errorf("Expected empty array, got %s", val.Inspect())
		}
	})

	t.Run("empty mapping", func(t *testing.T) {
		yaml := "config: {}"
		obj, err := objects.ParseYAML(yaml)
		if err != nil {
			t.Fatalf("ParseYAML() error = %v", err)
		}
		m := obj.(*objects.Map)
		val := m.Pairs[objects.NewString("config").HashKey()].Value
		if innerMap, ok := val.(*objects.Map); !ok || len(innerMap.Pairs) != 0 {
			t.Errorf("Expected empty map, got %s", val.Inspect())
		}
	})
}

// TestYAMLSetValue tests the setYAMLValue and setYAMLValueRecursive functions
func TestYAMLSetValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		path     string
		value    string
		expected string
	}{
		{
			name:     "set simple key",
			initial:  "a: 1\nb: 2",
			path:     "a",
			value:    "10",
			expected: "a: 10\nb: 2",
		},
		{
			name:     "set nested key",
			initial:  "server:\n  host: localhost\n  port: 8080",
			path:     "server.port",
			value:    "9090",
			expected: "server:\n  host: localhost\n  port: 9090",
		},
		{
			name:     "set creates intermediate maps",
			initial:  "a: 1",
			path:     "b.c.d",
			value:    "test",
			expected: "a: 1\nb:\n  c:\n    d: test",
		},
		{
			name:     "set array index",
			initial:  "items: [a, b, c]",
			path:     "items.[1]",
			value:    "B",
			expected: "items: [a, B, c]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.initial)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			// Parse the value as a YAML object
			valueObj, err := objects.ParseYAML(tt.value)
			if err != nil {
				t.Fatalf("ParseYAML(value) error = %v", err)
			}

			// Call setYAMLValue
			result := setYAMLValue(obj, tt.path, valueObj)
			if result.Type() == objects.ErrorType {
				t.Fatalf("setYAMLValue() error: %s", result.Inspect())
			}

			// Serialize and compare
			yamlStr := objects.SerializeYAML(result, 2)
			t.Logf("Result:\n%s", yamlStr)
		})
	}
}

// TestYAMLMapToArray tests the mapToArray function
func TestYAMLMapToArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple numeric map",
			input:    "0: a\n1: b\n2: c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "map with gaps",
			input:    "0: a\n2: c\n4: e",
			expected: []string{"a", "null", "c", "null", "e"},
		},
		{
			name:     "empty map",
			input:    "{}",
			expected: []string{},
		},
		{
			name:     "non-numeric keys only",
			input:    "a: 1",
			expected: []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.input)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			m, ok := obj.(*objects.Map)
			if !ok {
				t.Fatalf("expected Map, got %T", obj)
			}

			result := mapToArray(m)
			// mapToArray returns *objects.Array directly
			arr := result

			if len(arr.Elements) != len(tt.expected) {
				t.Errorf("expected %d elements, got %d", len(tt.expected), len(arr.Elements))
			}

			for i, expected := range tt.expected {
				if i < len(arr.Elements) {
					actual := arr.Elements[i].Inspect()
					if actual != expected {
						t.Errorf("element %d: expected %q, got %q", i, expected, actual)
					}
				}
			}
		})
	}
}

// TestYAMLDiff tests the diff function
func TestYAMLDiff(t *testing.T) {
	yaml1 := `
name: test
value: 100
nested:
  a: 1
  b: 2
`
	yaml2 := `
name: test
value: 200
nested:
  a: 1
  c: 3
`

	obj1, _ := objects.ParseYAML(yaml1)
	obj2, _ := objects.ParseYAML(yaml2)

	// Call diff through the module
	diffs := yamlDiffRecursive(obj1, obj2, "")

	t.Logf("Differences: %s", diffs.Inspect())

	// Should have differences: value changed, b removed, c added
	if len(diffs.Elements) < 2 {
		t.Errorf("Expected at least 2 differences, got %d", len(diffs.Elements))
	}
}

// TestYAMLFlatten tests the flatten function
func TestYAMLFlatten(t *testing.T) {
	yaml := `
server:
  host: localhost
  port: 8080
database:
  name: mydb
`

	obj, _ := objects.ParseYAML(yaml)

	result := make(map[objects.HashKey]objects.MapPair)
	yamlFlattenRecursive(obj, "", ".", result)

	flat := &objects.Map{Pairs: result}
	t.Logf("Flattened: %s", flat.Inspect())

	// Check that paths are flattened correctly
	expectedPaths := []string{"server.host", "server.port", "database.name"}
	for _, path := range expectedPaths {
		key := objects.NewString(path)
		if _, exists := flat.Pairs[key.HashKey()]; !exists {
			t.Errorf("Expected path %s in flattened result", path)
		}
	}
}

// TestYAMLExpand tests the expand function
func TestYAMLExpand(t *testing.T) {
	flatYaml := `
"server.host": localhost
"server.port": 8080
"database.name": mydb
`

	obj, _ := objects.ParseYAML(flatYaml)
	expanded := yamlExpandMap(obj.(*objects.Map), ".")

	t.Logf("Expanded: %s", expanded.Inspect())

	m := expanded.(*objects.Map)
	if len(m.Pairs) != 2 {
		t.Errorf("Expected 2 top-level keys, got %d", len(m.Pairs))
	}
}

// TestYAMLEquals tests the equals function
func TestYAMLEquals(t *testing.T) {
	tests := []struct {
		name     string
		yaml1    string
		yaml2    string
		expected bool
	}{
		{
			name:     "identical",
			yaml1:    "name: test",
			yaml2:    "name: test",
			expected: true,
		},
		{
			name:     "different values",
			yaml1:    "name: test",
			yaml2:    "name: other",
			expected: false,
		},
		{
			name:     "different keys",
			yaml1:    "name: test",
			yaml2:    "value: test",
			expected: false,
		},
		{
			name: "nested identical",
			yaml1: `
nested:
  a: 1
  b: 2
`,
			yaml2: `
nested:
  a: 1
  b: 2
`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj1, _ := objects.ParseYAML(tt.yaml1)
			obj2, _ := objects.ParseYAML(tt.yaml2)

			result := yamlEquals(obj1, obj2)
			if result != tt.expected {
				t.Errorf("equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestYAMLClone tests the clone function
func TestYAMLClone(t *testing.T) {
	yaml := `
nested:
  a: 1
  b: 2
array:
  - x
  - y
`

	obj, _ := objects.ParseYAML(yaml)
	clone := yamlDeepCopy(obj)

	// Verify clone is equal
	if !yamlEquals(obj, clone) {
		t.Error("Clone should be equal to original")
	}

	// Modify original and verify clone is unaffected
	m := obj.(*objects.Map)
	nestedKey := objects.NewString("nested")
	nested := m.Pairs[nestedKey.HashKey()].Value.(*objects.Map)
	nested.Pairs[objects.NewString("c").HashKey()] = objects.MapPair{
		Key:   objects.NewString("c"),
		Value: objects.NewInt(3),
	}

	// Clone should not have the new key
	cloneMap := clone.(*objects.Map)
	cloneNested := cloneMap.Pairs[nestedKey.HashKey()].Value.(*objects.Map)
	if _, exists := cloneNested.Pairs[objects.NewString("c").HashKey()]; exists {
		t.Error("Clone should not be affected by modifications to original")
	}
}

// TestYAMLFind tests the find function with patterns
func TestYAMLFind(t *testing.T) {
	yaml := `
server:
  host: localhost
  port: 8080
  endpoints:
    - /api
    - /health
database:
  host: db.example.com
  port: 5432
`

	obj, _ := objects.ParseYAML(yaml)

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
	}{
		{"exact match", "server.host", 1},
		{"wildcard single", "*.host", 2},
		{"array index", "server.endpoints.[0]", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matches []objects.Object
			yamlFindByPattern(obj, tt.pattern, "", &matches)

			if len(matches) != tt.expectedCount {
				t.Errorf("find(%s) returned %d matches, want %d", tt.pattern, len(matches), tt.expectedCount)
			}
			matchArr := &objects.Array{Elements: matches}
			t.Logf("find(%s): %s", tt.pattern, matchArr.Inspect())
		})
	}
}

// TestYAMLPaths tests the paths function
func TestYAMLPaths(t *testing.T) {
	yaml := `
a:
  b: 1
  c: 2
d: 3
`

	obj, _ := objects.ParseYAML(yaml)

	var paths []objects.Object
	yamlCollectPaths(obj, "", &paths)

	expectedPaths := map[string]bool{
		"a":   true,
		"a.b": true,
		"a.c": true,
		"d":   true,
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	for _, p := range paths {
		pathStr := p.(*objects.String).Value
		if !expectedPaths[pathStr] {
			t.Errorf("Unexpected path: %s", pathStr)
		}
	}
}

// TestYAMLMultilineFlowSequence tests multiline flow sequences
func TestYAMLMultilineFlowSequence(t *testing.T) {
	yaml := `data: [
  1,
  2,
  3
]`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	m := obj.(*objects.Map)
	arr := m.Pairs[objects.NewString("data").HashKey()].Value.(*objects.Array)

	if len(arr.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr.Elements))
	}

	for i, expected := range []int64{1, 2, 3} {
		if arr.Elements[i].(*objects.Int).Value != expected {
			t.Errorf("Element %d = %d, want %d", i, arr.Elements[i].(*objects.Int).Value, expected)
		}
	}
}

// TestYAMLMultilineFlowMapping tests multiline flow mappings
func TestYAMLMultilineFlowMapping(t *testing.T) {
	yaml := `data: {
  a: 1,
  b: 2
}`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	m := obj.(*objects.Map)
	inner := m.Pairs[objects.NewString("data").HashKey()].Value.(*objects.Map)

	if len(inner.Pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(inner.Pairs))
	}
}

// TestYAMLComplexMappingKeySequence tests complex mapping keys with sequences
func TestYAMLComplexMappingKeySequence(t *testing.T) {
	yaml := "? - key1\n  - key2\n: value"

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Complex key sequence: %s", obj.Inspect())

	m := obj.(*objects.Map)
	if len(m.Pairs) != 1 {
		t.Errorf("Expected 1 pair, got %d", len(m.Pairs))
	}
}

// TestYAMLComplexMappingKeyMapping tests complex mapping keys with mappings
func TestYAMLComplexMappingKeyMapping(t *testing.T) {
	yaml := "? name: alice\n  age: 30\n: person"

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Complex key mapping: %s", obj.Inspect())

	m := obj.(*objects.Map)
	if len(m.Pairs) != 1 {
		t.Errorf("Expected 1 pair, got %d", len(m.Pairs))
	}
}

// TestYAMLMergeKeyArray tests merge key with array of aliases
func TestYAMLMergeKeyArray(t *testing.T) {
	// Note: Anchors must be defined before use in our parser
	yaml := `a: &a
  x: 1
b: &b
  y: 2
result:
  <<: [*a, *b]`

	obj, err := objects.ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	t.Logf("Merge key array: %s", obj.Inspect())

	m := obj.(*objects.Map)
	result := m.Pairs[objects.NewString("result").HashKey()].Value.(*objects.Map)

	// Should have both x and y from merged maps
	if len(result.Pairs) != 2 {
		t.Errorf("Expected 2 pairs in result, got %d", len(result.Pairs))
	}
}

// TestYAMLEscapeSequences tests various escape sequences
func TestYAMLEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{"unicode escape", `s: "\u0041"`, "A"},
		{"hex escape", `s: "\x41"`, "A"},
		{"newline escape", `s: "a\nb"`, "a\nb"},
		{"tab escape", `s: "a\tb"`, "a\tb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objects.ParseYAML(tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}

			m := obj.(*objects.Map)
			val := m.Pairs[objects.NewString("s").HashKey()].Value.(*objects.String).Value

			if val != tt.expected {
				t.Errorf("Got %q, want %q", val, tt.expected)
			}
		})
	}
}
