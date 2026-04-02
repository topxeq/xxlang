// pkg/objects/toml_obj_test.go
package objects

import (
	"testing"
)

func TestNewTomlDocument(t *testing.T) {
	doc := NewTomlDocument()
	if doc == nil {
		t.Fatal("expected document instance")
	}
}

func TestTomlDocumentType(t *testing.T) {
	doc := NewTomlDocument()
	if doc.Type() != TomlDocumentType {
		t.Errorf("expected TomlDocumentType, got %v", doc.Type())
	}
}

func TestTomlDocumentInspect(t *testing.T) {
	doc := NewTomlDocument()
	s := doc.Inspect()
	if s == "" {
		t.Error("expected non-empty inspect string")
	}
}

func TestTomlDocumentToBool(t *testing.T) {
	doc := NewTomlDocument()
	if doc.ToBool() != TRUE {
		t.Error("expected TRUE")
	}
}

func TestTomlValueTypeString(t *testing.T) {
	tests := []struct {
		vt       TomlValueType
		expected string
	}{
		{TomlString, "STRING"},
		{TomlInteger, "INTEGER"},
		{TomlFloat, "FLOAT"},
		{TomlBoolean, "BOOLEAN"},
		{TomlArray, "ARRAY"},
		{TomlTable, "TABLE"},
	}

	for _, tc := range tests {
		if tc.vt.String() != tc.expected {
			t.Errorf("expected '%s', got '%s'", tc.expected, tc.vt.String())
		}
	}
}
