// pkg/stdlib/toml_test.go
// Tests for toml module.
package stdlib

import (
	"os"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestTomlParse(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["parse"].(*objects.Builtin)
	// Valid TOML
	result := fn.Fn(objects.NewString(`a = 1`))
	if result.Type() != objects.TomlDocumentType {
		t.Fatalf("expected TomlDocument, got %s", result.Type())
	}
	doc := result.(*objects.TomlDocument)
	keys := doc.Keys()
	if len(keys) < 1 || keys[0] != "a" {
		t.Errorf("expected key 'a', got %v", keys)
	}
	// Invalid TOML
	result = fn.Fn(objects.NewString(`[invalid`))
	if result.Type() != objects.ErrorType {
		t.Fatalf("expected Error for invalid TOML, got %s", result.Type())
	}
}

func TestTomlParseFile(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["parseFile"].(*objects.Builtin)

	// Create temporary TOML file
	tmpFile, err := os.CreateTemp("", "test_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(`app = "MyApp"
version = 1.0
`)
	tmpFile.Close()

	// Parse the file
	result := fn.Fn(objects.NewString(tmpFile.Name()))
	if result.Type() != objects.TomlDocumentType {
		t.Fatalf("expected TomlDocument, got %s", result.Type())
	}
	doc := result.(*objects.TomlDocument)
	keys := doc.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	if !contains(keys, "app") || !contains(keys, "version") {
		t.Errorf("expected keys 'app' and 'version', got %v", keys)
	}
}

func TestTomlStringify(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["stringify"].(*objects.Builtin)

	// Parse a TOML string first to get a document
	parseFn := mod.Exports["parse"].(*objects.Builtin)
	doc := parseFn.Fn(objects.NewString(`name = "Bob"
age = 25
`))
	if doc.Type() != objects.TomlDocumentType {
		t.Fatalf("expected TomlDocument from parse, got %s", doc.Type())
	}

	// Stringify the document
	result := fn.Fn(doc)
	if str, ok := result.(*objects.String); ok {
		tomlStr := str.Value
		if !strings.Contains(tomlStr, "name") || !strings.Contains(tomlStr, "Bob") {
			t.Errorf("expected TOML string to contain name='Bob', got %s", tomlStr)
		}
		if !strings.Contains(tomlStr, "age") || !strings.Contains(tomlStr, "25") {
			t.Errorf("expected TOML string to contain age=25, got %s", tomlStr)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}
}

func TestTomlEncode(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["encode"].(*objects.Builtin)
	// Encode a simple string
	result := fn.Fn(objects.NewString("hello"))
	if result.Type() != objects.StringType {
		t.Fatalf("expected String, got %s", result.Type())
	}
}

func TestTomlCreate(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["create"].(*objects.Builtin)

	doc := fn.Fn()
	if doc.Type() != objects.TomlDocumentType {
		t.Fatalf("expected TomlDocument, got %s", doc.Type())
	}
}

func TestTomlIsValid(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["isValid"].(*objects.Builtin)

	// Valid TOML
	result := fn.Fn(objects.NewString(`a = 1`))
	if b, ok := result.(*objects.Bool); ok {
		if !b.Value {
			t.Errorf("expected true for valid TOML")
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}

	// Invalid TOML
	result = fn.Fn(objects.NewString(`[`))
	if b, ok := result.(*objects.Bool); ok {
		if b.Value {
			t.Errorf("expected false for invalid TOML")
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}
}

func TestTomlIsTomlDocument(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["isTomlDocument"].(*objects.Builtin)

	// Not a TomlDocument
	result := fn.Fn(objects.NULL)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != false {
			t.Errorf("expected false for NULL, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}

	// TomlDocument
	doc := objects.NewTomlDocument()
	result = fn.Fn(doc)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != true {
			t.Errorf("expected true for TomlDocument, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}
}

func TestTomlFromJson(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["fromJson"].(*objects.Builtin)

	jsonStr := `{"name":"Alice","age":30}`
	result := fn.Fn(objects.NewString(jsonStr))
	if str, ok := result.(*objects.String); ok {
		tomlStr := str.Value
		// Should produce TOML representation
		if !strings.Contains(tomlStr, "name") || !strings.Contains(tomlStr, "Alice") {
			t.Errorf("expected TOML to contain name='Alice', got %s", tomlStr)
		}
		if !strings.Contains(tomlStr, "age") || !strings.Contains(tomlStr, "30") {
			t.Errorf("expected TOML to contain age=30, got %s", tomlStr)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}
}

func TestTomlToJson(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	fn := mod.Exports["toJson"].(*objects.Builtin)

	// Pass a TomlDocument
	parseFn := mod.Exports["parse"].(*objects.Builtin)
	doc := parseFn.Fn(objects.NewString(`name = "Bob"
age = 25
`))
	if doc.Type() != objects.TomlDocumentType {
		t.Fatalf("expected TomlDocument, got %s", doc.Type())
	}

	result := fn.Fn(doc)
	if str, ok := result.(*objects.String); ok {
		jsonStr := str.Value
		// Should contain JSON fields
		if !strings.Contains(jsonStr, "name") || !strings.Contains(jsonStr, "Bob") {
			t.Errorf("expected JSON to contain name='Bob', got %s", jsonStr)
		}
		if !strings.Contains(jsonStr, "age") || !strings.Contains(jsonStr, "25") {
			t.Errorf("expected JSON to contain age=25, got %s", jsonStr)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
