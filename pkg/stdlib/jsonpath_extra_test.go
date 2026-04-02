// pkg/stdlib/jsonpath_extra_test.go
// Additional tests for jsonpath module to cover previously untested functions
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJSONPathSlice tests slice syntax in JSONPath (parseSliceSegment, getSlice)
func TestJSONPathSlice(t *testing.T) {
	arr := objects.NewArray([]objects.Object{
		objects.NewString("a"),
		objects.NewString("b"),
		objects.NewString("c"),
		objects.NewString("d"),
		objects.NewString("e"),
	})

	tests := []struct {
		path string
		want []string
	}{
		{"$[0:2]", []string{"a", "b"}},
		{"$[1:4]", []string{"b", "c", "d"}},
		{"$[:3]", []string{"a", "b", "c"}},
		{"$[2:]", []string{"c", "d", "e"}},
		{"$[0:5:2]", []string{"a", "c", "e"}}, // step 2
		{"$[::2]", []string{"a", "c", "e"}},
		{"$[1::2]", []string{"b", "d"}},
		{"$[-3:-1]", []string{"c", "d"}}, // negative start/end
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			jp, err := ParseJSONPath(tt.path)
			if err != nil {
				t.Fatalf("ParseJSONPath error: %v", err)
			}
			results := jp.Get(arr)
			if len(results) != len(tt.want) {
				t.Fatalf("expected %d results, got %d", len(tt.want), len(results))
			}
			for i, r := range results {
				s, ok := r.(*objects.String)
				if !ok || s.Value != tt.want[i] {
					t.Fatalf("result %d: expected %q, got %q", i, tt.want[i], r)
				}
			}
		})
	}
}

// TestJSONPathIndex tests single index access (getIndex)
func TestJSONPathIndex(t *testing.T) {
	arr := objects.NewArray([]objects.Object{
		objects.NewString("a"),
		objects.NewString("b"),
		objects.NewString("c"),
	})

	tests := []struct {
		path string
		want string
	}{
		{"$[0]", "a"},
		{"$[1]", "b"},
		{"$[2]", "c"},
		{"$[-1]", "c"},
		{"$[-2]", "b"},
		{"$[-3]", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			jp, err := ParseJSONPath(tt.path)
			if err != nil {
				t.Fatalf("ParseJSONPath error: %v", err)
			}
			results := jp.Get(arr)
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			s, ok := results[0].(*objects.String)
			if !ok || s.Value != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, results[0])
			}
		})
	}
}

// TestJSONPathFilterOperators tests filter operators: in, contains, startsWith, endsWith, regex, between, isNull, isNotNull, isType
func TestJSONPathFilterOperators(t *testing.T) {
	objs := []objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("alice")},
			objects.NewString("age").HashKey():  {Key: objects.NewString("age"), Value: objects.NewInt(30)},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("go"), objects.NewString("test")})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("bob")},
			objects.NewString("age").HashKey():  {Key: objects.NewString("age"), Value: objects.NewInt(25)},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("java")})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("charlie")},
			objects.NewString("age").HashKey():  {Key: objects.NewString("age"), Value: objects.NewInt(35)},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("go"), objects.NewString("python")})},
		}),
	}
	arr := objects.NewArray(objs)

	tests := []struct {
		path    string
		wantIdx []int // indices of expected matches
	}{
		// in operator: @.name in ["alice","charlie"]
		{"$[?(@.name in [\"alice\",\"charlie\"])]", []int{0, 2}},
		// contains: @.tags contains "go"
		{"$[?(@.tags contains \"go\")]", []int{0, 2}},
		// startsWith: @.name startsWith "a"
		{"$[?(@.name startsWith \"a\")]", []int{0}},
		// endsWith: @.name endsWith "e"
		{"$[?(@.name endsWith \"e\")]", []int{0, 2}},
		// regex: @.name =~ "^a" (string pattern)
		{"$[?(@.name =~ \"^a\")]", []int{0}},
		// between: @.age between [20,30] (assume inclusive)
		{"$[?(@.age between [20,30])]", []int{0, 1}},
		// isNull: @.nickname isNull (field absent)
		{"$[?(@.nickname isNull)]", []int{0, 1, 2}},
		// isNotNull: @.age isNotNull
		{"$[?(@.age isNotNull)]", []int{0, 1, 2}},
		// isType: @.age isType "int"
		{"$[?(@.age isType \"int\")]", []int{0, 1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			jp, err := ParseJSONPath(tt.path)
			if err != nil {
				t.Fatalf("ParseJSONPath error: %v", err)
			}
			results := jp.Get(arr)
			if len(results) != len(tt.wantIdx) {
				t.Fatalf("expected %d results for path %s, got %d", len(tt.wantIdx), tt.path, len(results))
			}
			for i, idx := range tt.wantIdx {
				if results[i] != objs[idx] {
					t.Fatalf("result %d: expected object at index %d, got %T", i, idx, results[i])
				}
			}
		})
	}
}

// TestJSONPathFilterFunctions tests filter functions: empty() and length()
func TestJSONPathFilterFunctions(t *testing.T) {
	objs := []objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("alice")},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{})}, // empty array
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("bob")},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("go")})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("charlie")},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NULL}, // null
		}),
	}
	arr := objects.NewArray(objs)

	// empty(@.tags): true for empty array or null
	jp1, err := ParseJSONPath("$[?(empty(@.tags))]")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	res1 := jp1.Get(arr)
	if len(res1) != 2 {
		t.Fatalf("empty filter expected 2 matches, got %d", len(res1))
	}

	// length(@.tags) > 0: matches second
	jp2, err := ParseJSONPath("$[?(length(@.tags) > 0)]")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	res2 := jp2.Get(arr)
	if len(res2) != 1 {
		t.Fatalf("length filter expected 1 match, got %d", len(res2))
	}
}

// TestJSONPaths tests the exported Paths function (0% coverage)
func TestJSONPaths(t *testing.T) {
	// Build a nested object: {"a": {"b": 1, "c": 2}, "d": [3,4]}
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewInt(1)},
		objects.NewString("c").HashKey(): {Key: objects.NewString("c"), Value: objects.NewInt(2)},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: inner},
		objects.NewString("d").HashKey(): {Key: objects.NewString("d"), Value: objects.NewArray([]objects.Object{objects.NewInt(3), objects.NewInt(4)})},
	})
	paths := Paths(root)
	// Expected paths: "a", "a.b", "a.c", "d", "d[0]", "d[1]"
	expected := []string{"a", "a.b", "a.c", "d", "d[0]", "d[1]"}
	set := make(map[string]bool)
	for _, p := range paths {
		set[p] = true
	}
	for _, exp := range expected {
		if !set[exp] {
			t.Fatalf("missing expected path %s in %v", exp, paths)
		}
	}
}

// TestJSONPathDeleteWithPath tests Delete with nested path (deleteValue)
func TestJSONPathDeleteWithPath(t *testing.T) {
	// Structure: {"a": {"b": 1, "c": 2}}
	bVal := objects.NewInt(1)
	cVal := objects.NewInt(2)
	aMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: bVal},
		objects.NewString("c").HashKey(): {Key: objects.NewString("c"), Value: cVal},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: aMap},
	})

	// Delete $.a.b
	jp, _ := ParseJSONPath("$.a.b")
	newRoot, err := jp.Delete(root)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify b is gone
	jpCheckB, _ := ParseJSONPath("$.a.b")
	resultsB := jpCheckB.Get(newRoot)
	if len(resultsB) != 0 {
		t.Error("Field 'b' should be deleted")
	}

	// Verify c still exists
	jpCheckC, _ := ParseJSONPath("$.a.c")
	resultsC := jpCheckC.Get(newRoot)
	if len(resultsC) != 1 {
		t.Error("Field 'c' should still exist")
	}
}

// TestJSONPathSetWithArrayIndex tests Set using array index syntax
func TestJSONPathSetWithArrayIndex(t *testing.T) {
	arr := objects.NewArray([]objects.Object{
		objects.NewString("old1"),
		objects.NewString("old2"),
	})
	jp, _ := ParseJSONPath("$[1]")
	newArr, err := jp.Set(arr, objects.NewString("new2"))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	results := jp.Get(newArr)
	if len(results) != 1 {
		t.Fatalf("expected 1 result after set, got %d", len(results))
	}
	if s, ok := results[0].(*objects.String); !ok || s.Value != "new2" {
		t.Fatalf("expected value 'new2', got %v", results[0])
	}
}

// TestJSONPathArrayLiteralInFilter tests parseArrayLiteral via filter with array literal
func TestJSONPathArrayLiteralInFilter(t *testing.T) {
	objs := []objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("A")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("B")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("C")},
		}),
	}
	arr := objects.NewArray(objs)

	jp, _ := ParseJSONPath(`$[?(@.category in ["A","C"])]`)
	results := jp.Get(arr)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}
