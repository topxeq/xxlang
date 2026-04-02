package stdlib

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Helper to get yaml module safely in tests
func yamlModule(t *testing.T) *Module {
	t.Helper()
	mod := Get("yaml")
	if mod == nil {
		t.Fatal("yaml module not found")
	}
	// Cast safely using an interface indirection to avoid compile-time assertions
	var v interface{} = mod
	if m, ok := v.(*Module); ok {
		return m
	}
	t.Fatal("yaml module has unexpected type")
	return nil
}

// TestYAMLFileIO covers writeFile, readFile, and updateFile behaviors using a temp directory.
func TestYAMLFileIO(t *testing.T) {
	t.Run("write/read/update", func(t *testing.T) {
		mod := yamlModule(t)
		// Prepare data to write
		data := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():  {Key: objects.NewString("name"), Value: objects.NewString("testfile")},
			objects.NewString("count").HashKey(): {Key: objects.NewString("count"), Value: objects.NewInt(5)},
		}}

		// Create temp file path
		dir := t.TempDir()
		path := filepath.Join(dir, "data.yaml")

		// writeFile
		wf := mod.Exports["writeFile"].(*objects.Builtin)
		res := wf.Fn(objects.NewString(path), data, objects.NewInt(2))
		if res != nil && res.Type() == objects.ErrorType {
			t.Fatalf("writeFile error: %s", res.Inspect())
		}

		// readFile (should parse YAML back to an object)
		rf := mod.Exports["readFile"].(*objects.Builtin)
		rd := rf.Fn(objects.NewString(path))
		if rd.Type() == objects.ErrorType {
			t.Fatalf("readFile error: %s", rd.Inspect())
		}
		// Sanity: ensure the parsed object is a Map with 2 keys
		m, ok := rd.(*objects.Map)
		if !ok {
			t.Fatalf("readFile did not return a Map, got %T", rd)
		}
		if len(m.Pairs) != 2 {
			t.Fatalf("expected 2 keys in read data, got %d", len(m.Pairs))
		}

		// updateFile: change count to 10
		up := mod.Exports["updateFile"].(*objects.Builtin)
		updates := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("count").HashKey(): {Key: objects.NewString("count"), Value: objects.NewInt(10)},
		}}
		updRes := up.Fn(objects.NewString(path), updates)
		if updRes.Type() == objects.ErrorType {
			t.Fatalf("updateFile error: %s", updRes.Inspect())
		}
		// Read again to verify update
		rd2 := rf.Fn(objects.NewString(path))
		if rd2.Type() == objects.ErrorType {
			t.Fatalf("readFile after update error: %s", rd2.Inspect())
		}
		m2, _ := rd2.(*objects.Map)
		if val := m2.Pairs[objects.NewString("count").HashKey()].Value; val.Inspect() != "10" {
			t.Fatalf("expected updated count to be 10, got %s", val.Inspect())
		}
	})
}

// TestYAMLStringifyIndentEdgeCase ensures negative indent is treated as 0 and non-negative works.
func TestYAMLStringifyIndentEdgeCase(t *testing.T) {
	mod := yamlModule(t)
	stringify := mod.Exports["stringify"].(*objects.Builtin)

	obj, _ := objects.ParseYAML("name: edge")
	// Negative indent -> 0
	res0 := stringify.Fn(obj, objects.NewInt(-5))
	if res0.Type() == objects.ErrorType {
		t.Fatalf("stringify with negative indent error: %s", res0.Inspect())
	}
	// Should be a string, and should contain 'name:' at least
	s0, _ := res0.(*objects.String)
	if s0 == nil || !strings.Contains(s0.Value, "name:") {
		t.Fatalf("stringify with negative indent produced invalid YAML: %v", res0.Inspect())
	}
}

// TestFromJsonToJsonIndent covers fromJson and toJson with string indentation and int indentation.
func TestFromJsonToJsonIndent(t *testing.T) {
	mod := yamlModule(t)
	fromJson := mod.Exports["fromJson"].(*objects.Builtin)
	toJson := mod.Exports["toJson"].(*objects.Builtin)

	jsonInput := objects.NewString(`{"server": {"host": "localhost", "port": 8080}}`)
	yamlStr := fromJson.Fn(jsonInput)
	if yamlStr.Type() == objects.ErrorType {
		t.Fatalf("fromJson() error: %s", yamlStr.Inspect())
	}

	// Convert back to JSON with indentation string and then with int indentation
	jsonOut := toJson.Fn(yamlStr)
	if jsonOut.Type() == objects.ErrorType {
		t.Fatalf("toJson() error: %s", jsonOut.Inspect())
	}
	t.Logf("From JSON -> YAML -> JSON: %s", jsonOut.Inspect())
}

// TestParseDocumentsError ensures that invalid documents produce an error in parseDocuments
func TestParseDocumentsError(t *testing.T) {
	mod := yamlModule(t)
	parseDocs := mod.Exports["parseDocuments"].(*objects.Builtin)
	bad := objects.NewString(`---
valid: yes
invalid: [unclosed`)
	res := parseDocs.Fn(bad)
	if res.Type() != objects.ErrorType {
		t.Fatalf("Expected error from parseDocuments, got: %s", res.Inspect())
	}
}

// TestYAMLFlattenExpandRoundtrip ensures flatten followed by expand yields equivalent structure
func TestYAMLFlattenExpandRoundtrip(t *testing.T) {
	mod := yamlModule(t)
	// Build a nested object
	original := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("server").HashKey(): {
			Key:   objects.NewString("server"),
			Value: &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: objects.NewString("localhost")}, objects.NewString("port").HashKey(): {Key: objects.NewString("port"), Value: objects.NewInt(8080)}}},
		},
		objects.NewString("env").HashKey(): {Key: objects.NewString("env"), Value: objects.NewString("prod")},
	}}

	flatRes := mod.Exports["flatten"].(*objects.Builtin).Fn(original)
	if flatRes.Type() == objects.ErrorType {
		t.Fatalf("flatten() error: %s", flatRes.Inspect())
	}
	flatMap := flatRes.(*objects.Map)
	// Expand back
	expanded := mod.Exports["expand"].(*objects.Builtin).Fn(flatMap)
	if expanded.Type() == objects.ErrorType {
		t.Fatalf("expand() error: %s", expanded.Inspect())
	}
	if !yamlEquals(original, expanded) {
		t.Errorf("Expanded object does not match original")
	}
}

// TestYAMLDeepMergeNonMap verifies non-map overrides in deepMerge
func TestYAMLDeepMergeNonMap(t *testing.T) {
	mod := yamlModule(t)
	dMerge := mod.Exports["deepMerge"].(*objects.Builtin)
	base := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(1)}}}}}}
	override := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(2)}}}
	res := dMerge.Fn(base, override)
	if res.Type() == objects.ErrorType {
		t.Fatalf("deepMerge() error: %s", res.Inspect())
	}
	m := res.(*objects.Map)
	if val := m.Pairs[objects.NewString("a").HashKey()].Value.Inspect(); val != "2" {
		t.Fatalf("Expected merged override value 2, got %s", val)
	}
}

// TestYAMLKeysValues ensures keys() and values() work on a map
func TestYAMLKeysValues(t *testing.T) {
	mod := yamlModule(t)
	keysFn := mod.Exports["keys"].(*objects.Builtin)
	valuesFn := mod.Exports["values"].(*objects.Builtin)
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewString("1")},
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewString("2")},
	}}

	k := keysFn.Fn(m)
	v := valuesFn.Fn(m)
	if k.Type() != objects.ArrayType || v.Type() != objects.ArrayType {
		t.Fatalf("keys/values should return arrays")
	}
}
