// pkg/stdlib/toml_extra_test.go
// Additional tests for toml module to increase coverage.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestTomlIsValid_Extra tests toml.isValid with various inputs.
func TestTomlIsValid_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["isValid"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"", true}, // empty is valid TOML
		{"key = \"value\"", true},
		{"[section]\nkey = 123", true},
		{"invalid toml here", false},
		{"key = 123", true},
		{"key = true", true},
		{"key = [1,2,3]", true},
		{"key = {a = 1}", true},
		{"# comment only", true}, // comment-only is valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			// isValid returns a boolean object
			b, ok := result.(*objects.Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("isValid(%q) = %v, want %v", tt.input, b.Value, tt.expected)
			}
		})
	}
}

// TestTomlIsTomlDocument_Extra tests toml.isTomlDocument.
func TestTomlIsTomlDocument_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["isTomlDocument"].(*objects.Builtin)
	createFn := mod.Exports["create"].(*objects.Builtin)

	// Create a toml document
	doc := createFn.Fn()
	resDoc := fn.Fn(doc)
	if resDoc != objects.TRUE {
		t.Errorf("isTomlDocument(create()) = %v, want TRUE", resDoc)
	}

	// Non-document (e.g., a string) should be false
	resStr := fn.Fn(String("not a document"))
	if resStr != objects.FALSE {
		t.Errorf("isTomlDocument(string) = %v, want FALSE", resStr)
	}
}

// TestTomlFromJson_Extra tests toml.fromJson conversion.
func TestTomlFromJson_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["fromJson"].(*objects.Builtin)

	// Simple JSON string
	jsonStr := `{"name":"test","value":42}`
	result := fn.Fn(String(jsonStr))
	if result.Type() == objects.ErrorType {
		t.Fatalf("fromJson returned error: %s", result.Inspect())
	}
	// Should return a TOML string
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String result, got %T", result)
	}
	if s.Value == "" {
		t.Error("fromJson returned empty string")
	}
}

// TestTomlToJson_Extra tests toml.toJson conversion.
func TestTomlToJson_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["toJson"].(*objects.Builtin)

	// We can construct a map that represents TOML data directly.
	tomlMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("title").HashKey(): {Key: objects.NewString("title"), Value: objects.NewString("Hello")},
	}}

	result := fn.Fn(tomlMap)
	if result.Type() == objects.ErrorType {
		t.Fatalf("toJson returned error: %s", result.Inspect())
	}
	// Expect a JSON string
	jsonStr, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if jsonStr.Value == "" {
		t.Error("toJson returned empty string")
	}
}

// TestTomlParseEdgeCases tests toml.parse with edge cases.
func TestTomlParseEdgeCases(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["parse"].(*objects.Builtin)

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"", false}, // empty TOML is valid
		{"key = \"value\"", false},
		{"[section]\nkey = 123", false},
		{"invalid", true},
		{"key = 123\nkey2 = \"abc\"", false},
		{"key = [1,2,3]", false},
		{"key = {a = 1, b = 2}", false},
		{"# comment only", false}, // comment-only is valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			if tt.wantErr {
				if result.Type() != objects.ErrorType {
					t.Errorf("parse(%q) expected error, got %s", tt.input, result.Inspect())
				}
			} else {
				if result.Type() == objects.ErrorType {
					t.Errorf("parse(%q) unexpected error: %s", tt.input, result.Inspect())
				}
			}
		})
	}
}

// TestTomlStringifyEdgeCases tests toml.stringify with TomlDocument objects.
func TestTomlStringifyEdgeCases(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	stringifyFn := mod.Exports["stringify"].(*objects.Builtin)
	parseFn := mod.Exports["parse"].(*objects.Builtin)

	// Parse a simple TOML to obtain a TomlDocument
	docResult := parseFn.Fn(String("a = 1"))
	if docResult.Type() == objects.ErrorType {
		t.Fatalf("parse error: %s", docResult.Inspect())
	}
	doc, ok := docResult.(*objects.TomlDocument)
	if !ok {
		t.Fatalf("expected TomlDocument, got %T", docResult)
	}

	// Stringify the document
	result := stringifyFn.Fn(doc)
	if result.Type() == objects.ErrorType {
		t.Fatalf("stringify(doc) error: %s", result.Inspect())
	}
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value == "" {
		t.Error("stringify returned empty string")
	}
	// Should contain the key "a"
	if !strings.Contains(s.Value, "a") {
		t.Errorf("stringify output doesn't contain 'a': %s", s.Value)
	}
}

// TestTomlEncode_Extra tests toml.encode (serialize to TOML string).
func TestTomlEncode_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["encode"].(*objects.Builtin)

	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("key").HashKey(): {Key: objects.NewString("key"), Value: objects.NewString("value")},
	}}

	result := fn.Fn(m)
	if result.Type() == objects.ErrorType {
		t.Fatalf("encode error: %s", result.Inspect())
	}
	// Expect a TOML string
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value == "" {
		t.Error("encode returned empty string")
	}
}

// TestTomlFileOperations tests readFile/writeFile if available.
// Note: toml.readFile and toml.writeFile may exist; check existence.
func TestTomlFileOperations(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	// These functions might be named differently or not present; check existence.
	// If not present, skip.
	readFn, readOk := mod.Exports["readFile"].(*objects.Builtin)
	writeFn, writeOk := mod.Exports["writeFile"].(*objects.Builtin)
	if !readOk || !writeOk {
		t.Skip("toml file functions not available")
	}

	// Create a temp TOML file
	tmpPath := t.TempDir() + "/test.toml"
	content := "key = \"value\"\ncount = 42"
	// Write file
	writeRes := writeFn.Fn(String(tmpPath), String(content), objects.NewInt(2))
	if writeRes.Type() == objects.ErrorType {
		t.Fatalf("writeFile error: %s", writeRes.Inspect())
	}
	// Read file
	readRes := readFn.Fn(String(tmpPath))
	if readRes.Type() == objects.ErrorType {
		t.Fatalf("readFile error: %s", readRes.Inspect())
	}
	// Should return a map
	if readRes.Type() != objects.MapType && readRes.Type() != objects.NullType {
		t.Fatalf("expected Map or NULL, got %s", readRes.Type())
	}
}
