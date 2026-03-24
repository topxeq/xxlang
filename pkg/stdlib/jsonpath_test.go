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

// TestJSONPathFilterAdvanced tests advanced filter operators
func TestJSONPathFilterAdvanced(t *testing.T) {
	// Create test data with various fields
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():    {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("age").HashKey():     {Key: objects.NewString("age"), Value: objects.NewInt(25)},
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("fiction")},
			objects.NewString("active").HashKey():  {Key: objects.NewString("active"), Value: objects.TRUE},
			objects.NewString("tags").HashKey():    {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("tag1"), objects.NewString("tag2")})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():     {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("age").HashKey():      {Key: objects.NewString("age"), Value: objects.NewInt(30)},
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("drama")},
			objects.NewString("active").HashKey():   {Key: objects.NewString("active"), Value: objects.FALSE},
			objects.NewString("tags").HashKey():     {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("tag2"), objects.NewString("tag3")})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():     {Key: objects.NewString("name"), Value: objects.NewString("Charlie")},
			objects.NewString("age").HashKey():      {Key: objects.NewString("age"), Value: objects.NewInt(35)},
			objects.NewString("category").HashKey(): {Key: objects.NewString("category"), Value: objects.NewString("fiction")},
			objects.NewString("active").HashKey():   {Key: objects.NewString("active"), Value: objects.TRUE},
			objects.NewString("tags").HashKey():     {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewString("tag1")})},
		}),
	})

	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		// Logical AND
		{"$[?(@.age > 20 && @.age < 30)]", 1, "AND: age between 20 and 30"},
		{"$[?(@.category == \"fiction\" && @.active == true)]", 2, "AND: fiction AND active"},

		// Logical OR
		{"$[?(@.age < 26 || @.age > 31)]", 2, "OR: age < 26 or age > 31"},

		// Logical NOT
		{"$[?(!@.active)]", 1, "NOT: not active"},

		// Combined AND/OR
		{"$[?(@.category == \"fiction\" && (@.age < 30 || @.active == true))]", 2, "Combined: fiction AND (age<30 OR active)"},

		// In operator
		{"$[?(@.category in [\"fiction\", \"drama\"])]", 3, "IN: category in fiction or drama"},
		{"$[?(@.age in [25, 35])]", 2, "IN: age is 25 or 35"},

		// Not in operator
		{"$[?(@.category nin [\"fiction\"])]", 1, "NIN: category not fiction"},

		// Contains for string
		{"$[?(@.name contains \"a\")]", 1, "Contains: name contains 'a' (only Charlie, case-sensitive)"},
		{"$[?(@.name contains \"li\")]", 2, "Contains: name contains 'li'"},

		// Contains for array
		{"$[?(@.tags contains \"tag1\")]", 2, "Contains: tags contains tag1"},

		// StartsWith
		{"$[?(@.name startsWith \"A\")]", 1, "StartsWith: name starts with 'A'"},
		{"$[?(@.name startsWith \"B\")]", 1, "StartsWith: name starts with 'B'"},

		// EndsWith
		{"$[?(@.name endsWith \"e\")]", 2, "EndsWith: name ends with 'e'"},

		// Regex match
		{"$[?(@.name =~ \"^[AB].*\")]", 2, "Regex: name starts with A or B"},
		{"$[?(@.name =~ \".*ie.*\")]", 1, "Regex: name contains 'ie'"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jp, err := ParseJSONPath(tc.path)
			if err != nil {
				t.Errorf("Failed to parse path %s: %v", tc.path, err)
				return
			}

			results := jp.Get(items)
			if len(results) != tc.expected {
				t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
			}
		})
	}
}

// TestJSONPathFilterEmpty tests the empty() function
func TestJSONPathFilterEmpty(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("items").HashKey(): {Key: objects.NewString("items"), Value: objects.NewArray([]objects.Object{})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("items").HashKey(): {Key: objects.NewString("items"), Value: objects.NewArray([]objects.Object{objects.NewInt(1)})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("")},
		}),
	})

	testCases := []struct {
		path     string
		expected int
	}{
		{"$[?(empty(@.items))]", 2},  // Alice has empty items, third item has no items field (null is empty)
		{"$[?(empty(@.name))]", 1},   // Empty name
		{"$[?(!empty(@.items))]", 1}, // Bob has non-empty items
	}

	for _, tc := range testCases {
		jp, err := ParseJSONPath(tc.path)
		if err != nil {
			t.Errorf("Failed to parse path %s: %v", tc.path, err)
			continue
		}

		results := jp.Get(items)
		if len(results) != tc.expected {
			t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
		}
	}
}

// TestJSONPathFilterLength tests the length() function
func TestJSONPathFilterLength(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewInt(1), objects.NewInt(2)})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("tags").HashKey(): {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{objects.NewInt(1)})},
		}),
	})

	testCases := []struct {
		path     string
		expected int
	}{
		{"$[?(length(@.name) > 3)]", 1},     // Alice has 5 chars
		{"$[?(length(@.tags) == 2)]", 1},    // Alice has 2 tags
		{"$[?(@.tags.length() == 1)]", 1},   // Bob has 1 tag (alternative syntax)
	}

	for _, tc := range testCases {
		jp, err := ParseJSONPath(tc.path)
		if err != nil {
			t.Errorf("Failed to parse path %s: %v", tc.path, err)
			continue
		}

		results := jp.Get(items)
		if len(results) != tc.expected {
			t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
		}
	}
}

// TestJSONPathFilterExistence tests existence checks
func TestJSONPathFilterExistence(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NewString("alice@example.com")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Charlie")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NULL},
		}),
	})

	testCases := []struct {
		path     string
		expected int
	}{
		{"$[?(@.email)]", 1},              // Only Alice has non-null email
		{"$[?(@.name)]", 3},               // All have name
		{"$[?(@.missing == null)]", 3},    // All have missing field which is null
	}

	for _, tc := range testCases {
		jp, err := ParseJSONPath(tc.path)
		if err != nil {
			t.Errorf("Failed to parse path %s: %v", tc.path, err)
			continue
		}

		results := jp.Get(items)
		if len(results) != tc.expected {
			t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
		}
	}
}

// TestJSONPathFilterBetween tests the "between" operator
func TestJSONPathFilterBetween(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(10)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(50)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Charlie")},
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(100)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("David")},
			objects.NewString("price").HashKey(): {Key: objects.NewString("price"), Value: objects.NewInt(150)},
		}),
	})

	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		{"$[?(@.price between [20, 100])]", 2, "price between 20 and 100 (Bob=50, Charlie=100)"},
		{"$[?(@.price between [10, 50])]", 2, "price between 10 and 50 (Alice=10, Bob=50)"},
		{"$[?(@.price between [0, 200])]", 4, "price between 0 and 200 (all items)"},
		{"$[?(@.price between [60, 90])]", 0, "price between 60 and 90 (none)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jp, err := ParseJSONPath(tc.path)
			if err != nil {
				t.Errorf("Failed to parse path %s: %v", tc.path, err)
				return
			}

			results := jp.Get(items)
			if len(results) != tc.expected {
				t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
			}
		})
	}
}

// TestJSONPathFilterIsNull tests the "isNull" and "isNotNull" operators
func TestJSONPathFilterIsNull(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NewString("alice@example.com")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NULL},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Charlie")},
		}),
	})

	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		{"$[?(@.email isNull)]", 2, "email is null (Bob has null, Charlie has missing)"},
		{"$[?(@.email isNotNull)]", 1, "email is not null (only Alice)"},
		{"$[?(@.name isNotNull)]", 3, "name is not null (all have name)"},
		{"$[?(@.missing isNull)]", 3, "missing field is null (all)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jp, err := ParseJSONPath(tc.path)
			if err != nil {
				t.Errorf("Failed to parse path %s: %v", tc.path, err)
				return
			}

			results := jp.Get(items)
			if len(results) != tc.expected {
				t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
			}
		})
	}
}

// TestJSONPathFilterIsType tests the "isType" operator
func TestJSONPathFilterIsType(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():   {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("age").HashKey():    {Key: objects.NewString("age"), Value: objects.NewInt(25)},
			objects.NewString("score").HashKey():  {Key: objects.NewString("score"), Value: objects.NewFloat(95.5)},
			objects.NewString("active").HashKey(): {Key: objects.NewString("active"), Value: objects.TRUE},
			objects.NewString("tags").HashKey():   {Key: objects.NewString("tags"), Value: objects.NewArray([]objects.Object{})},
			objects.NewString("meta").HashKey():   {Key: objects.NewString("meta"), Value: objects.NewMap(map[objects.HashKey]objects.MapPair{})},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():   {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("age").HashKey():    {Key: objects.NewString("age"), Value: objects.NewFloat(30.5)},
			objects.NewString("active").HashKey(): {Key: objects.NewString("active"), Value: objects.FALSE},
		}),
	})

	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		{"$[?(@.age isType \"int\")]", 1, "age is int (Alice)"},
		{"$[?(@.age isType \"float\")]", 1, "age is float (Bob)"},
		{"$[?(@.age isType \"number\")]", 2, "age is number (both)"},
		{"$[?(@.name isType \"string\")]", 2, "name is string (both)"},
		{"$[?(@.active isType \"boolean\")]", 2, "active is boolean (both)"},
		{"$[?(@.tags isType \"array\")]", 1, "tags is array (Alice)"},
		{"$[?(@.meta isType \"object\")]", 1, "meta is object (Alice)"},
		{"$[?(@.missing isType \"null\")]", 2, "missing is null (both)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jp, err := ParseJSONPath(tc.path)
			if err != nil {
				t.Errorf("Failed to parse path %s: %v", tc.path, err)
				return
			}

			results := jp.Get(items)
			if len(results) != tc.expected {
				t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
			}
		})
	}
}

// TestJSONPathFilterAbsent tests the "absent" operator
func TestJSONPathFilterAbsent(t *testing.T) {
	items := objects.NewArray([]objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NewString("alice@example.com")},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
			objects.NewString("email").HashKey(): {Key: objects.NewString("email"), Value: objects.NULL},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Charlie")},
		}),
	})

	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		{"$[?(@.email absent)]", 1, "email absent (only Charlie)"},
		{"$[?(@.name absent)]", 0, "name absent (none)"},
		{"$[?(@.phone absent)]", 3, "phone absent (all)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jp, err := ParseJSONPath(tc.path)
			if err != nil {
				t.Errorf("Failed to parse path %s: %v", tc.path, err)
				return
			}

			results := jp.Get(items)
			if len(results) != tc.expected {
				t.Errorf("Path %s: expected %d results, got %d", tc.path, tc.expected, len(results))
			}
		})
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
