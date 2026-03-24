// pkg/stdlib/jsonpath_test.go
// Tests for JSONPath functionality
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJSONPathParse tests JSONPath parsing
func TestJSONPathParse(t *testing.T) {
	testCases := []struct {
		path     string
		hasError bool
	}{
		{"$", false},
		{"$.field", false},
		{"$.field.nested", false},
		{"$[0]", false},
		{"$.field[0]", false},
		{"$[*]", false},
		{"$..field", false},
		{"$.field[*]", false},
		{"$.field[0:5]", false},
		{"$.field[0,1,2]", false},
		{".field", false}, // Relative path
		{"", false},
	}

	for _, tc := range testCases {
		_, err := ParseJSONPath(tc.path)
		if tc.hasError && err == nil {
			t.Errorf("Expected error for path: %s", tc.path)
		}
		if !tc.hasError && err != nil {
			t.Errorf("Unexpected error for path %s: %v", tc.path, err)
		}
	}
}

// TestJSONPathGet tests basic JSONPath get operations
func TestJSONPathGet(t *testing.T) {
	// Create test object: {"store": {"book": [{"title": "Book1"}, {"title": "Book2"}], "name": "MyStore"}}
	book1 := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("title").HashKey(): {Key: objects.NewString("title"), Value: objects.NewString("Book1")},
	})
	book2 := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("title").HashKey(): {Key: objects.NewString("title"), Value: objects.NewString("Book2")},
	})
	books := objects.NewArray([]objects.Object{book1, book2})
	store := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("book").HashKey(): {Key: objects.NewString("book"), Value: books},
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("MyStore")},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("store").HashKey(): {Key: objects.NewString("store"), Value: store},
	})

	testCases := []struct {
		path     string
		expected string
	}{
		{"$.store.name", "MyStore"},
		{"$.store.book[0].title", "Book1"},
		{"$.store.book[1].title", "Book2"},
		{"$.store.book[-1].title", "Book2"}, // Negative index
	}

	for _, tc := range testCases {
		jp, err := ParseJSONPath(tc.path)
		if err != nil {
			t.Errorf("Failed to parse path %s: %v", tc.path, err)
			continue
		}

		results := jp.Get(root)
		if len(results) == 0 {
			t.Errorf("No results for path: %s", tc.path)
			continue
		}

		str, ok := results[0].(*objects.String)
		if !ok {
			t.Errorf("Expected string for path %s, got %T", tc.path, results[0])
			continue
		}

		if str.Value != tc.expected {
			t.Errorf("Path %s: expected %s, got %s", tc.path, tc.expected, str.Value)
		}
	}
}

// TestJSONPathWildcard tests wildcard operations
func TestJSONPathWildcard(t *testing.T) {
	// Create array: [1, 2, 3]
	arr := objects.NewArray([]objects.Object{
		objects.NewInt(1),
		objects.NewInt(2),
		objects.NewInt(3),
	})

	jp, _ := ParseJSONPath("$[*]")
	results := jp.Get(arr)

	if len(results) != 3 {
		t.Errorf("Expected 3 results from wildcard, got %d", len(results))
	}
}

// TestJSONPathRecursive tests recursive descent
func TestJSONPathRecursive(t *testing.T) {
	// Create nested structure
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("inner")},
	})
	outer := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("outer")},
		objects.NewString("child").HashKey(): {Key: objects.NewString("child"), Value: inner},
	})

	jp, _ := ParseJSONPath("$..name")
	results := jp.Get(outer)

	if len(results) != 2 {
		t.Errorf("Expected 2 results from recursive search, got %d", len(results))
	}
}

// TestJSONPathSet tests setting values
func TestJSONPathSet(t *testing.T) {
	// Create object: {"a": 1}
	obj := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(1)},
	})

	jp, _ := ParseJSONPath("$.a")
	newObj, err := jp.Set(obj, objects.NewInt(2))
	if err != nil {
		t.Errorf("Set failed: %v", err)
		return
	}

	// Verify the value was set
	results := jp.Get(newObj)
	if len(results) == 0 {
		t.Error("No result after set")
		return
	}

	if i, ok := results[0].(*objects.Int); !ok || i.Value != 2 {
		t.Errorf("Expected 2 after set, got %v", results[0])
	}

	// Verify original is unchanged
	origResults := jp.Get(obj)
	if len(origResults) == 0 {
		t.Error("No result from original")
		return
	}

	if i, ok := origResults[0].(*objects.Int); !ok || i.Value != 1 {
		t.Errorf("Original should still be 1, got %v", origResults[0])
	}
}

// TestJSONPathDelete tests deleting values
func TestJSONPathDelete(t *testing.T) {
	// Create object: {"a": 1, "b": 2}
	obj := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(1)},
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewInt(2)},
	})

	jp, _ := ParseJSONPath("$.a")
	newObj, err := jp.Delete(obj)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
		return
	}

	// Verify 'a' was deleted
	jpCheck, _ := ParseJSONPath("$.a")
	results := jpCheck.Get(newObj)
	if len(results) != 0 {
		t.Error("Field 'a' should be deleted")
	}

	// Verify 'b' still exists
	jpCheckB, _ := ParseJSONPath("$.b")
	resultsB := jpCheckB.Get(newObj)
	if len(resultsB) == 0 {
		t.Error("Field 'b' should still exist")
	}
}

// TestJSONPathSlice tests array slicing
func TestJSONPathSlice(t *testing.T) {
	// Create array: [0, 1, 2, 3, 4]
	arr := objects.NewArray([]objects.Object{
		objects.NewInt(0),
		objects.NewInt(1),
		objects.NewInt(2),
		objects.NewInt(3),
		objects.NewInt(4),
	})

	testCases := []struct {
		path       string
		expected   []int64
	}{
		{"$[1:3]", []int64{1, 2}},
		{"$[::2]", []int64{0, 2, 4}}, // Every other element
		{"$[:3]", []int64{0, 1, 2}},  // First 3
		{"$[2:]", []int64{2, 3, 4}},  // From index 2 to end
	}

	for _, tc := range testCases {
		jp, err := ParseJSONPath(tc.path)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", tc.path, err)
			continue
		}

		results := jp.Get(arr)
		if len(results) != len(tc.expected) {
			t.Errorf("Path %s: expected %d results, got %d", tc.path, len(tc.expected), len(results))
			continue
		}

		for i, exp := range tc.expected {
			if val, ok := results[i].(*objects.Int); !ok || val.Value != exp {
				t.Errorf("Path %s[%d]: expected %d, got %v", tc.path, i, exp, results[i])
			}
		}
	}
}

// TestJSONPathFilter tests filter expressions
func TestJSONPathFilter(t *testing.T) {
	// Create array of objects: [{"price": 5}, {"price": 10}, {"price": 15}]
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(5)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(10)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(15)},
		}),
	})

	jp, err := ParseJSONPath("$[?(@.price > 8)]")
	if err != nil {
		t.Errorf("Failed to parse filter path: %v", err)
		return
	}

	results := jp.Get(items)
	if len(results) != 2 {
		t.Errorf("Expected 2 items with price > 8, got %d", len(results))
	}
}

// TestJSONPaths tests the Paths function
func TestJSONPaths(t *testing.T) {
	// Create simple object: {"a": {"b": 1}}
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewInt(1)},
	})
	obj := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: inner},
	})

	paths := Paths(obj)

	expected := []string{"$.a", "$.a.b"}
	for _, exp := range expected {
		found := false
		for _, p := range paths {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path %s not found in %v", exp, paths)
		}
	}
}

// TestJSONPathModuleFunctions tests the module functions
func TestJSONPathModuleFunctions(t *testing.T) {
	// Get the json module
	jsonModule := Get("json")
	if jsonModule == nil {
		t.Fatal("json module not found")
	}

	// Check that JSONPath functions are exported
	expectedFuncs := []string{
		"get", "getAll", "getWithPath",
		"set", "delete",
		"paths", "has", "count",
		"query", "queryAll",
	}

	for _, fn := range expectedFuncs {
		if _, ok := jsonModule.Exports[fn]; !ok {
			t.Errorf("json module missing function: %s", fn)
		}
	}
}
