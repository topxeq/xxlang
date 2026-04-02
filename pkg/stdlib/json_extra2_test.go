// pkg/stdlib/json_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJSONExtra2_Parse tests json.parse.
func TestJSONExtra2_Parse(t *testing.T) {
	// No args
	if res := jsonCall("parse"); res.Type() != objects.ErrorType {
		t.Fatalf("parse() with no args should error, got %s", res.Type())
	}
	// Non-string arg
	if res := jsonCall("parse", objects.NewInt(123)); res.Type() != objects.ErrorType {
		t.Fatalf("parse() with int arg should error, got %s", res.Type())
	}
	// Invalid JSON
	if res := jsonCall("parse", objects.NewString("invalid")); res.Type() != objects.ErrorType {
		t.Fatalf("parse() with invalid JSON should error, got %s", res.Type())
	}
	// Valid JSON object
	if res := jsonCall("parse", objects.NewString(`{"a":1}`)); res.Type() != objects.MapType {
		t.Fatalf("parse() with valid JSON should return Map, got %s", res.Type())
	}
	// Valid JSON array
	if res := jsonCall("parse", objects.NewString(`[1,2,3]`)); res.Type() != objects.ArrayType {
		t.Fatalf("parse() with valid JSON array should return Array, got %s", res.Type())
	}
}

// TestJSONExtra2_Stringify tests json.stringify.
func TestJSONExtra2_Stringify(t *testing.T) {
	// No args
	if res := jsonCall("stringify"); res.Type() != objects.ErrorType {
		t.Fatalf("stringify() with no args should error")
	}
	// String arg is valid; should produce JSON string
	if res := jsonCall("stringify", objects.NewString("str")); res.Type() != objects.StringType {
		t.Fatalf("stringify() with string arg should return String, got %s", res.Type())
	}
	// Valid: a map
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(1)},
	}}
	if res := jsonCall("stringify", m); res.Type() != objects.StringType {
		t.Fatalf("stringify() with map should return String, got %s", res.Type())
	}
	// With indent (string)
	if res := jsonCall("stringify", m, objects.NewString("  ")); res.Type() != objects.StringType {
		t.Fatalf("stringify() with indent should return String")
	}
	// With indent (int)
	if res := jsonCall("stringify", m, objects.NewInt(2)); res.Type() != objects.StringType {
		t.Fatalf("stringify() with indent int should return String")
	}
}

// TestJSONExtra2_Encode tests json.encode.
func TestJSONExtra2_Encode(t *testing.T) {
	// No args
	if res := jsonCall("encode"); res.Type() != objects.ErrorType {
		t.Fatalf("encode() with no args should error")
	}
	// Valid map
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(42)},
	}}
	if res := jsonCall("encode", m); res.Type() != objects.StringType {
		t.Fatalf("encode() should return String, got %s", res.Type())
	}
}

// TestJSONExtra2_Decode tests json.decode (alias of parse).
func TestJSONExtra2_Decode(t *testing.T) {
	// No args
	if res := jsonCall("decode"); res.Type() != objects.ErrorType {
		t.Fatalf("decode() with no args should error")
	}
	// Non-string
	if res := jsonCall("decode", objects.NewInt(1)); res.Type() != objects.ErrorType {
		t.Fatalf("decode() with int should error")
	}
	// Invalid JSON
	if res := jsonCall("decode", objects.NewString("bad")); res.Type() != objects.ErrorType {
		t.Fatalf("decode() with invalid JSON should error")
	}
	// Valid
	if res := jsonCall("decode", objects.NewString(`["a","b"]`)); res.Type() != objects.ArrayType {
		t.Fatalf("decode() should return Array")
	}
}

// TestJSONExtra2_ToJson tests json.toJson with options.
func TestJSONExtra2_ToJson(t *testing.T) {
	// No args
	if res := jsonCall("toJson"); res.Type() != objects.ErrorType {
		t.Fatalf("toJson() with no args should error")
	}
	// Valid map
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("k").HashKey(): {Key: objects.NewString("k"), Value: objects.NewString("v")},
	}}
	// Basic
	if res := jsonCall("toJson", m); res.Type() != objects.StringType {
		t.Fatalf("toJson() should return String")
	}
	// With -indent flag
	if res := jsonCall("toJson", m, objects.NewString("-indent")); res.Type() != objects.StringType {
		t.Fatalf("toJson() with -indent should return String")
	}
	// With -sort flag
	if res := jsonCall("toJson", m, objects.NewString("-sort")); res.Type() != objects.StringType {
		t.Fatalf("toJson() with -sort should return String")
	}
}

// TestJSONExtra2_FromJson tests json.fromJson (alias of parse).
func TestJSONExtra2_FromJson(t *testing.T) {
	// Same as parse but using fromJson name
	if res := jsonCall("fromJson"); res.Type() != objects.ErrorType {
		t.Fatalf("fromJson() with no args should error")
	}
	if res := jsonCall("fromJson", objects.NewString(`null`)); res.Type() != objects.NullType {
		t.Fatalf("fromJson() null should return Null")
	}
}

// TestJSONExtra2_ReadFile tests json.readFile.
func TestJSONExtra2_ReadFile(t *testing.T) {
	// No args
	if res := jsonCall("readFile"); res.Type() != objects.ErrorType {
		t.Fatalf("readFile() with no args should error")
	}
	// Non-string arg
	if res := jsonCall("readFile", objects.NewInt(1)); res.Type() != objects.ErrorType {
		t.Fatalf("readFile() with int should error")
	}
	// Non-existent file should error
	if res := jsonCall("readFile", objects.NewString("nonexistent.json")); res.Type() != objects.ErrorType {
		t.Fatalf("readFile() with missing file should error")
	}
}

// TestJSONExtra2_WriteFile tests json.writeFile.
func TestJSONExtra2_WriteFile(t *testing.T) {
	// No args
	if res := jsonCall("writeFile"); res.Type() != objects.ErrorType {
		t.Fatalf("writeFile() with no args should error")
	}
	// Only one arg
	if res := jsonCall("writeFile", objects.NewString("path")); res.Type() != objects.ErrorType {
		t.Fatalf("writeFile() with 1 arg should error")
	}
	// First arg non-string
	if res := jsonCall("writeFile", objects.NewInt(1), objects.NewString("{}")); res.Type() != objects.ErrorType {
		t.Fatalf("writeFile() with int path should error")
	}
	// Second arg can be any object; Int is valid
	if res := jsonCall("writeFile", objects.NewString("test.json"), objects.NewInt(2)); res.Type() != objects.NullType {
		t.Fatalf("writeFile() with int data should return Null, got %s", res.Type())
	}
	// Valid write (to temp file) with map
	tmp := t.TempDir() + "/test.json"
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(1)},
	}}
	if res := jsonCall("writeFile", objects.NewString(tmp), m); res.Type() != objects.NullType {
		t.Fatalf("writeFile() should return Null, got %s", res.Type())
	}
	// With indent (string)
	if res := jsonCall("writeFile", objects.NewString(tmp), m, objects.NewString("  ")); res.Type() != objects.NullType {
		t.Fatalf("writeFile() with indent should return Null")
	}
	// With indent (int)
	if res := jsonCall("writeFile", objects.NewString(tmp), m, objects.NewInt(2)); res.Type() != objects.NullType {
		t.Fatalf("writeFile() with indent int should return Null")
	}
}

// TestJSONExtra2_WriteFilePretty tests json.writeFilePretty.
func TestJSONExtra2_WriteFilePretty(t *testing.T) {
	// Argument validation similar to writeFile
	if res := jsonCall("writeFilePretty"); res.Type() != objects.ErrorType {
		t.Fatalf("writeFilePretty() with no args should error")
	}
	if res := jsonCall("writeFilePretty", objects.NewString("path")); res.Type() != objects.ErrorType {
		t.Fatalf("writeFilePretty() with 1 arg should error")
	}
	// Valid write
	tmp := t.TempDir() + "/pretty.json"
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewString("b")},
	}}
	if res := jsonCall("writeFilePretty", objects.NewString(tmp), m); res.Type() != objects.NullType {
		t.Fatalf("writeFilePretty() should return Null")
	}
	// With custom indent string
	if res := jsonCall("writeFilePretty", objects.NewString(tmp), m, objects.NewString("    ")); res.Type() != objects.NullType {
		t.Fatalf("writeFilePretty() with indent should return Null")
	}
}

// TestJSONExtra2_UpdateFile tests json.updateFile.
func TestJSONExtra2_UpdateFile(t *testing.T) {
	// No args
	if res := jsonCall("updateFile"); res.Type() != objects.ErrorType {
		t.Fatalf("updateFile() with no args should error")
	}
	// Only one arg
	if res := jsonCall("updateFile", objects.NewString("path")); res.Type() != objects.ErrorType {
		t.Fatalf("updateFile() with 1 arg should error")
	}
	// First arg non-string
	if res := jsonCall("updateFile", objects.NewInt(1), &objects.Map{}); res.Type() != objects.ErrorType {
		t.Fatalf("updateFile() with int path should error")
	}
	// Second arg non-map
	if res := jsonCall("updateFile", objects.NewString("path"), objects.NewString("not a map")); res.Type() != objects.ErrorType {
		t.Fatalf("updateFile() with non-map updates should error")
	}
	// Valid update on existing file: create a file first
	tmp := t.TempDir() + "/update.json"
	// Write initial JSON using jsonCall("writeFile")
	m0 := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("old").HashKey(): {Key: objects.NewString("old"), Value: objects.NewInt(1)},
	}}
	if res := jsonCall("writeFile", objects.NewString(tmp), m0); res.Type() != objects.NullType {
		t.Fatalf("setup: writeFile failed: %s", res.Inspect())
	}
	// Now update: provide a map with new value
	updates := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("new").HashKey(): {Key: objects.NewString("new"), Value: objects.NewString("added")},
	}}
	if res := jsonCall("updateFile", objects.NewString(tmp), updates); res.Type() != objects.NullType {
		t.Fatalf("updateFile() should return Null, got %s", res.Type())
	}
	// Verify result by reading file via json.readFile
	result := jsonCall("readFile", objects.NewString(tmp))
	if result.Type() != objects.MapType {
		t.Fatalf("updateFile result should be Map, got %s", result.Type())
	}
	// Check that both old and new are present (update merges)
	resultMap := result.(*objects.Map)
	if _, hasOld := resultMap.Pairs[objects.NewString("old").HashKey()]; !hasOld {
		t.Error("updateFile: old key missing")
	}
	if _, hasNew := resultMap.Pairs[objects.NewString("new").HashKey()]; !hasNew {
		t.Error("updateFile: new key missing")
	}
}
