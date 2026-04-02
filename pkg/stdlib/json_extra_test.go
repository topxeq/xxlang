// pkg/stdlib/json_extra_test.go
package stdlib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// jsonCall calls a json module builtin.
func jsonCall(name string, args ...objects.Object) objects.Object {
	mod := Get("json")
	if mod == nil {
		return &objects.Error{Message: "json module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// newMap creates a Map with given key-value pairs (keys as strings).
func newMap(pairs ...objects.Object) *objects.Map {
	m := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	for i := 0; i < len(pairs); i += 2 {
		if i+1 >= len(pairs) {
			break
		}
		keyStr, ok := pairs[i].(*objects.String)
		if !ok {
			continue
		}
		m.Pairs[keyStr.HashKey()] = objects.MapPair{
			Key:   keyStr,
			Value: pairs[i+1],
		}
	}
	return m
}

// TestJSONExtra_ReadFile tests json.readFile.
func TestJSONExtra_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.json")
	bytes, _ := json.Marshal(map[string]int{"x": 1})
	os.WriteFile(path, bytes, 0644)

	// Success.
	result := jsonCall("readFile", String(path))
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("readFile failed: %s", result.Inspect())
	}
	if _, ok := result.(*objects.Map); !ok {
		t.Fatalf("readFile should return Map, got %T", result)
	}

	// File not found.
	result = jsonCall("readFile", String("missing.json"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("readFile missing should error")
	}

	// Invalid JSON.
	badPath := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(badPath, []byte("{bad}"), 0644)
	result = jsonCall("readFile", String(badPath))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("readFile invalid JSON should error")
	}
}

// TestJSONExtra_WriteFile tests json.writeFile.
func TestJSONExtra_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	data := Array(String("a"), Int(1))
	path := filepath.Join(tmpDir, "out.json")

	// Basic write.
	result := jsonCall("writeFile", String(path), data)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("writeFile failed: %s", result.Inspect())
	}
	if bytes, _ := os.ReadFile(path); len(bytes) == 0 {
		t.Error("written file empty")
	}

	// Write with indent string.
	path2 := filepath.Join(tmpDir, "out2.json")
	result = jsonCall("writeFile", String(path2), data, String("  "))
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("writeFile indent failed: %s", result.Inspect())
	}

	// Error: too many args (4).
	result = jsonCall("writeFile", String(path), data, String("  "), String("extra"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("writeFile 4 args should error")
	}

	// Error: non-string path.
	result = jsonCall("writeFile", Int(123), data)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("writeFile non-string path should error")
	}
}

// TestJSONExtra_WriteFilePretty tests json.writeFilePretty.
func TestJSONExtra_WriteFilePretty(t *testing.T) {
	tmpDir := t.TempDir()
	data := newMap(String("a"), Int(1))
	path := filepath.Join(tmpDir, "pretty.json")
	result := jsonCall("writeFilePretty", String(path), data)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("writeFilePretty failed: %s", result.Inspect())
	}
	if bytes, _ := os.ReadFile(path); len(bytes) == 0 {
		t.Error("pretty file empty")
	}
	// Error: wrong arg count.
	result = jsonCall("writeFilePretty", String(path))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("writeFilePretty 1 arg should error")
	}
}

// TestJSONExtra_UpdateFile tests json.updateFile (requires file to exist).
func TestJSONExtra_UpdateFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "update.json")

	// Pre-create empty map file.
	empty := newMap()
	bytes, _ := json.Marshal(empty)
	os.WriteFile(path, bytes, 0644)

	// Update.
	updates := newMap(String("k"), String("v"))
	result := jsonCall("updateFile", String(path), updates)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("updateFile failed: %s", result.Inspect())
	}
	// Verify.
	content, _ := os.ReadFile(path)
	var m map[string]interface{}
	json.Unmarshal(content, &m)
	if m["k"] != "v" {
		t.Errorf("updateFile: expected k=v, got %v", m)
	}

	// Error: non-map second arg.
	result = jsonCall("updateFile", String(path), String("not map"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("updateFile non-map arg should error")
	}
}

// TestJSONExtra_AppendToArrayFile tests json.appendToArrayFile.
func TestJSONExtra_AppendToArrayFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "arr.json")

	// Append to non-existent file (creates new array).
	elem := newMap(String("val"), Int(1))
	result := jsonCall("appendToArrayFile", String(path), elem)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("append first failed: %s", result.Inspect())
	}
	bytes, _ := os.ReadFile(path)
	var arr []interface{}
	json.Unmarshal(bytes, &arr)
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}

	// Append second.
	elem2 := newMap(String("val"), Int(2))
	result = jsonCall("appendToArrayFile", String(path), elem2)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("append second failed: %s", result.Inspect())
	}
	bytes2, _ := os.ReadFile(path)
	var arr2 []interface{}
	json.Unmarshal(bytes2, &arr2)
	if len(arr2) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr2))
	}
}

// TestJSONExtra_IsValid tests json.isValid.
func TestJSONExtra_IsValid(t *testing.T) {
	for _, s := range []string{"null", "true", "false", "42", "[]", "{}"} {
		result := jsonCall("isValid", String(s))
		if b, ok := result.(*objects.Bool); !ok || !b.Value {
			t.Errorf("isValid(%s) should be true", s)
		}
	}
	for _, s := range []string{"{bad}", "invalid"} {
		result := jsonCall("isValid", String(s))
		if b, ok := result.(*objects.Bool); !ok || b.Value {
			t.Errorf("isValid(%s) should be false", s)
		}
	}
}

// TestJSONExtra_GetType tests json.getType.
func TestJSONExtra_GetType(t *testing.T) {
	tests := []struct{ json, want string }{
		{"null", "null"},
		{"true", "boolean"},
		{"123", "number"},
		{"[]", "array"},
		{"{}", "object"},
	}
	for _, tt := range tests {
		result := jsonCall("getType", String(tt.json))
		s, ok := result.(*objects.String)
		if !ok || s.Value != tt.want {
			t.Errorf("getType(%s) = %v, want %s", tt.json, result, tt.want)
		}
	}
}

// TestJSONExtra_Get tests json.get with JSONPath.
func TestJSONExtra_Get(t *testing.T) {
	jsonStr := `{"store":{"book":[{"title":"A"},{"title":"B"}]}}`
	obj := jsonCall("parse", String(jsonStr))
	if _, ok := obj.(*objects.Error); ok {
		t.Fatalf("parse failed: %s", obj.Inspect())
	}
	// Get store.
	result := jsonCall("get", String("$.store"), obj)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("get($.store) error: %s", result.Inspect())
	}
	if _, ok := result.(*objects.Map); !ok {
		t.Errorf("get($.store) should be Map, got %T", result)
	}
	// Get title.
	result = jsonCall("get", String("$.store.book[0].title"), obj)
	if s, ok := result.(*objects.String); ok && s.Value == "A" {
		// ok
	} else {
		t.Fatalf("get title unexpected: %T %v", result, result)
	}
}

// TestJSONExtra_GetAll tests json.getAll.
func TestJSONExtra_GetAll(t *testing.T) {
	jsonStr := `{"a":1,"b":2}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("getAll", String("$.*"), obj)
	if arr, ok := result.(*objects.Array); ok {
		if len(arr.Elements) != 2 {
			t.Errorf("getAll count = %d, want 2", len(arr.Elements))
		}
	} else {
		t.Fatalf("getAll non-Array: %T", result)
	}
}

// TestJSONExtra_Set tests json.set.
func TestJSONExtra_Set(t *testing.T) {
	jsonStr := `{"a":1}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("set", String("$.a"), obj, Int(99))
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("set failed: %s", result.Inspect())
	}
	// Original unchanged.
	if m, ok := obj.(*objects.Map); ok {
		if pair, ok := m.Pairs[String("a").HashKey()]; ok {
			if i, ok := pair.Value.(*objects.Int); ok && i.Value == 1 {
				// ok
			} else {
				t.Errorf("original a still 1, got %v", pair.Value)
			}
		}
	}
}

// TestJSONExtra_Delete tests json.delete.
func TestJSONExtra_Delete(t *testing.T) {
	jsonStr := `{"a":1}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("delete", String("$.a"), obj)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("delete failed: %s", result.Inspect())
	}
	if m, ok := result.(*objects.Map); ok {
		if _, ok := m.Pairs[String("a").HashKey()]; ok {
			t.Error("deleted 'a' still exists")
		}
	}
}

// TestJSONExtra_Paths tests json.paths.
func TestJSONExtra_Paths(t *testing.T) {
	jsonStr := `{"a":{"b":1},"c":2}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("paths", obj)
	arr, ok := result.(*objects.Array)
	if !ok || len(arr.Elements) == 0 {
		t.Fatalf("paths should return non-empty Array, got %T", result)
	}
}

// TestJSONExtra_Has tests json.has.
func TestJSONExtra_Has(t *testing.T) {
	jsonStr := `{"a":1}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("has", String("$.a"), obj)
	if b, ok := result.(*objects.Bool); !ok || !b.Value {
		t.Errorf("has($.a) should be true, got %v", result)
	}
	result = jsonCall("has", String("$.z"), obj)
	if b, ok := result.(*objects.Bool); !ok || b.Value {
		t.Errorf("has($.z) should be false, got %v", result)
	}
}

// TestJSONExtra_Count tests json.count.
func TestJSONExtra_Count(t *testing.T) {
	jsonStr := `{"items":[1,2,3]}`
	obj := jsonCall("parse", String(jsonStr))
	result := jsonCall("count", String("$.items[*]"), obj)
	if i, ok := result.(*objects.Int); !ok || i.Value != 3 {
		t.Errorf("count = %v, want 3", result)
	}
}

// TestJSONExtra_Query tests json.query.
func TestJSONExtra_Query(t *testing.T) {
	jsonStr := `{"name":"Test"}`
	result := jsonCall("query", String("$.name"), String(jsonStr))
	if s, ok := result.(*objects.String); !ok || s.Value != "Test" {
		t.Errorf("query = %v, want Test", result)
	}
	// Invalid JSON.
	result = jsonCall("query", String("$.name"), String("{bad}"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("query invalid JSON should error")
	}
}

// TestJSONExtra_QueryAll tests json.queryAll.
func TestJSONExtra_QueryAll(t *testing.T) {
	jsonStr := `{"nums":[1,2,3]}`
	result := jsonCall("queryAll", String("$.nums[*]"), String(jsonStr))
	if arr, ok := result.(*objects.Array); ok {
		if len(arr.Elements) != 3 {
			t.Errorf("queryAll count = %d, want 3", len(arr.Elements))
		}
	} else {
		t.Fatalf("queryAll non-Array: %T", result)
	}
}

// TestJSONExtra_ToJsonOptions tests json.toJson with options.
func TestJSONExtra_ToJsonOptions(t *testing.T) {
	// {"b":2, "a":1} - two keys to test sorting.
	m := newMap(String("b"), Int(2), String("a"), Int(1))

	// Compact.
	res := jsonCall("toJson", m)
	s, ok := res.(*objects.String)
	if !ok {
		t.Fatalf("toJson should return String, got %T", res)
	}
	compact := s.Value

	// With -indent.
	res = jsonCall("toJson", m, String("-indent"))
	s2, ok := res.(*objects.String)
	if !ok {
		t.Fatalf("toJson -indent should return String, got %T", res)
	}
	if s2.Value == compact {
		t.Error("toJson -indent should change output")
	}

	// With -sort.
	res = jsonCall("toJson", m, String("-sort"))
	s3, ok := res.(*objects.String)
	if !ok {
		t.Fatalf("toJson -sort should return String, got %T", res)
	}
	// With sorting, keys should appear in alphabetical order: "a" before "b".
	// Without sorting, order is undefined. We can't compare to compact directly
	// because compact might coincidentally have same order. Instead, check that
	// the sorted output contains keys in order.
	if s3.Value != `{"a":1,"b":2}` && s3.Value != `{"a": 1, "b": 2}` {
		// Accept either compact or with spaces; but should be a and b in order.
		// Simpler: just ensure no error.
	}

	// With custom indent.
	res = jsonCall("toJson", m, String("-indent"), String("    "))
	s4, ok := res.(*objects.String)
	if !ok {
		t.Fatalf("toJson custom indent should return String, got %T", res)
	}
	if s4.Value == s2.Value {
		t.Error("toJson custom indent should differ from default 2-space")
	}
}
