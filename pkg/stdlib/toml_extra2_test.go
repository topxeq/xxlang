// pkg/stdlib/toml_extra2_test.go
// Additional tests to further increase coverage for the toml stdlib module.
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestTomlParseFile_Extra exercises toml.parseFile with a real TOML file.
func TestTomlParseFile_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	parseFile := mod.Exports["parseFile"].(*objects.Builtin)

	// Create a temporary TOML file
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "sample.toml")
	if err := os.WriteFile(tmpPath, []byte("title = \"Hello\""), 0644); err != nil {
		t.Fatalf("failed to write temp toml: %v", err)
	}

	res := parseFile.Fn(String(tmpPath))
	if res.Type() == objects.ErrorType {
		t.Fatalf("parseFile returned error: %s", res.Inspect())
	}
	if _, ok := res.(*objects.TomlDocument); !ok {
		t.Fatalf("expected TomlDocument, got %T", res)
	}
}

// TestTomlParseFile_Missing_Extra ensures parseFile reports an error for missing files.
func TestTomlParseFile_Missing_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	parseFile := mod.Exports["parseFile"].(*objects.Builtin)

	// Use a path that does not exist
	missing := filepath.Join(t.TempDir(), "does_not_exist.toml")
	res := parseFile.Fn(String(missing))
	if res.Type() != objects.ErrorType {
		t.Fatalf("expected error for missing file, got %T", res)
	}
}

// TestTomlStringify_WrongArg_Extra validates stringify errors on non-document input.
func TestTomlStringify_WrongArg_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	stringify := mod.Exports["stringify"].(*objects.Builtin)
	// Pass a non-document object
	res := stringify.Fn(String("not a document"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("expected error for non-document input, got %T", res)
	}
}

// TestTomlToJson_FromString_Extra verifies toJson can consume TOML provided as a string.
func TestTomlToJson_FromString_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	toJson := mod.Exports["toJson"].(*objects.Builtin)
	// Provide a small TOML snippet as string
	res := toJson.Fn(String("title = \"World\""))
	if res.Type() == objects.ErrorType {
		t.Fatalf("toJson from string error: %s", res.Inspect())
	}
	s, ok := res.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", res)
	}
	if s.Value == "" {
		t.Error("toJson from string returned empty string")
	}
}

// TestTomlToJson_FromString_InvalidToml_Extra ensures error on invalid TOML string.
func TestTomlToJson_FromString_InvalidToml_Extra(t *testing.T) {
	mod := Get("toml")
	if mod == nil {
		t.Skip("toml module not found")
	}
	toJson := mod.Exports["toJson"].(*objects.Builtin)
	res := toJson.Fn(String("title = World")) // invalid TOML (missing quotes)
	if res.Type() != objects.ErrorType {
		t.Fatalf("expected error for invalid TOML, got %T", res)
	}
}
