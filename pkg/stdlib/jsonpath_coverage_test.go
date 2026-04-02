// pkg/stdlib/jsonpath_coverage_test.go
// Additional tests to improve coverage of jsonpath module
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJSONPathGetMultiIndex tests getMultiIndex function with various index combinations
func TestJSONPathGetMultiIndex(t *testing.T) {
	// Create array: ["a", "b", "c", "d", "e"]
	arr := objects.NewArray([]objects.Object{
		objects.NewString("a"),
		objects.NewString("b"),
		objects.NewString("c"),
		objects.NewString("d"),
		objects.NewString("e"),
	})

	// Test positive indices
	jp, _ := ParseJSONPath("$[0,2,4]")
	results := jp.Get(arr)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	if s, ok := results[0].(*objects.String); !ok || s.Value != "a" {
		t.Error("First result should be 'a'")
	}
	if s, ok := results[1].(*objects.String); !ok || s.Value != "c" {
		t.Error("Second result should be 'c'")
	}
	if s, ok := results[2].(*objects.String); !ok || s.Value != "e" {
		t.Error("Third result should be 'e'")
	}

	// Test negative indices
	jp2, _ := ParseJSONPath("$[-1,-3]")
	results2 := jp2.Get(arr)
	if len(results2) != 2 {
		t.Errorf("Expected 2 results with negative indices, got %d", len(results2))
	}
	if s, ok := results2[0].(*objects.String); !ok || s.Value != "e" {
		t.Error("Negative index -1 should give 'e'")
	}
	if s, ok := results2[1].(*objects.String); !ok || s.Value != "c" {
		t.Error("Negative index -3 should give 'c'")
	}

	// Test out-of-bounds indices (should be ignored)
	jp3, _ := ParseJSONPath("$[0,10,-10]")
	results3 := jp3.Get(arr)
	if len(results3) != 1 {
		t.Errorf("Expected 1 result (only index 0 valid), got %d", len(results3))
	}
}

// TestJSONPathGetWildcardOnMap tests getWildcard function on a map (array case already tested)
func TestJSONPathGetWildcardOnMap(t *testing.T) {
	// Create map: {"a": 1, "b": 2, "c": 3}
	m := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(1)},
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewInt(2)},
		objects.NewString("c").HashKey(): {Key: objects.NewString("c"), Value: objects.NewInt(3)},
	})

	jp, _ := ParseJSONPath("$.*")
	results := jp.Get(m)

	// Should return all values (1,2,3) - order not guaranteed but length 3
	if len(results) != 3 {
		t.Errorf("Expected 3 results from map wildcard, got %d", len(results))
	}

	// Check that we have the correct values (set for quick lookup)
	vals := make(map[int]bool)
	for _, v := range results {
		if i, ok := v.(*objects.Int); ok {
			vals[int(i.Value)] = true
		}
	}
	expected := map[int]bool{1: true, 2: true, 3: true}
	for k := range expected {
		if !vals[k] {
			t.Errorf("Missing value %d in wildcard results", k)
		}
	}
}

// TestJSONPathGetWithPathRecursiveField tests GetWithPath with recursive field descent (..field)
func TestJSONPathGetWithPathRecursiveField(t *testing.T) {
	// Nested structure: {"a": {"x": 1}, {"b": {"x": 2}} with array
	inner1 := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(1)},
	})
	inner2 := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(2)},
	})
	arr := objects.NewArray([]objects.Object{
		inner1,
		inner2,
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey():    {Key: objects.NewString("a"), Value: inner1},
		objects.NewString("list").HashKey(): {Key: objects.NewString("list"), Value: arr},
	})

	jp, _ := ParseJSONPath("$..x")
	matches := jp.GetWithPath(root)

	// Should find three matches: a.x, list[0].x, list[1].x
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches for recursive field 'x', got %d", len(matches))
	}

	// Verify paths and values
	paths := make(map[string]int)
	for _, m := range matches {
		paths[m.Path] = 1
	}
	expectedPaths := []string{"$.a.x", "$.list[0].x", "$.list[1].x"}
	for _, p := range expectedPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("Missing expected path %s in GetWithPath results", p)
		}
	}
}

// TestJSONPathGetWithPathRecursiveWildcard tests GetWithPath with recursive wildcard (..*)
func TestJSONPathGetWithPathRecursiveWildcard(t *testing.T) {
	// Simple nested structure
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("val").HashKey(): {Key: objects.NewString("val"), Value: objects.NewInt(10)},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("inner").HashKey(): {Key: objects.NewString("inner"), Value: inner},
		objects.NewString("top").HashKey():   {Key: objects.NewString("top"), Value: objects.NewString("topval")},
	})

	jp, _ := ParseJSONPath("$..*")
	matches := jp.GetWithPath(root)

	// Should include root itself? Actually GetWithPath starts with root as initial match, then applies segments. For "..*", it will recursively collect all values. Should include inner map, inner.val, top string, etc. At least a few.
	if len(matches) < 3 {
		t.Errorf("Expected at least 3 matches for recursive wildcard, got %d", len(matches))
	}
}

// TestJSONPathSetMultiLevelCreate tests Set with multi-level path creating intermediate maps
func TestJSONPathSetMultiLevelCreate(t *testing.T) {
	// Start with empty map
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{})

	// Set a deep path that doesn't exist: $.a.b.c = 42
	jp, _ := ParseJSONPath("$.a.b.c")
	newRoot, err := jp.Set(root, objects.NewInt(42))
	if err != nil {
		t.Fatalf("Set on deep path failed: %v", err)
	}

	// Verify the value was set and intermediate maps were created
	jpCheck, _ := ParseJSONPath("$.a.b.c")
	results := jpCheck.Get(newRoot)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result after set, got %d", len(results))
	}
	if i, ok := results[0].(*objects.Int); !ok || i.Value != 42 {
		t.Fatalf("Expected value 42, got %v", results[0])
	}

	// Also verify intermediate map exists
	jpA, _ := ParseJSONPath("$.a")
	aVal := jpA.Get(newRoot)
	if len(aVal) != 1 || aVal[0].(*objects.Map) == nil {
		t.Error("Intermediate map 'a' should exist")
	}
}

// TestJSONPathSetMultiLevelExisting tests Set on existing multi-level path
func TestJSONPathSetMultiLevelExisting(t *testing.T) {
	// Existing structure: {"a": {"b": {"x": 1}}}
	bMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: objects.NewInt(1)},
	})
	aMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: bMap},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: aMap},
	})

	// Change $.a.b.x to 100
	jp, _ := ParseJSONPath("$.a.b.x")
	newRoot, err := jp.Set(root, objects.NewInt(100))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify
	jpCheck, _ := ParseJSONPath("$.a.b.x")
	results := jpCheck.Get(newRoot)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if i, ok := results[0].(*objects.Int); !ok || i.Value != 100 {
		t.Fatalf("Expected 100, got %v", results[0])
	}
}

// TestJSONPathDeleteMultiLevel tests Delete on a nested path
func TestJSONPathDeleteMultiLevel(t *testing.T) {
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

// TestJSONPathFilterFalsy tests filter expressions using falsy checks via !@.field
func TestJSONPathFilterFalsy(t *testing.T) {
	// Create array of objects with various falsy/truthy fields
	objs := []objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("flag").HashKey(): {Key: objects.NewString("flag"), Value: Bool(false)},
			objects.NewString("num").HashKey():  {Key: objects.NewString("num"), Value: objects.NewInt(0)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("flag").HashKey(): {Key: objects.NewString("flag"), Value: Bool(true)},
			objects.NewString("num").HashKey():  {Key: objects.NewString("num"), Value: objects.NewInt(1)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("text").HashKey(): {Key: objects.NewString("text"), Value: objects.NewString("")}, // empty string falsy
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("text").HashKey(): {Key: objects.NewString("text"), Value: objects.NewString("hello")},
		}),
	}
	arr := objects.NewArray(objs)

	// Filter: ?(!@.flag) should select objects where flag is falsy (false, 0, empty, nil)
	jp, _ := ParseJSONPath("$[?(!@.flag)]")
	results := jp.Get(arr)

	// Should match first object (flag=false) and maybe third? third has no flag -> absent? Actually !@.flag on missing field: evalAbsentOperator? No, !@.flag is handled by the NOT branch: if inner starts with @ and no spaces, it evaluates inner and checks isFalsy. If field missing, evalFilterExpr will return something? Let's see: evalFilterExpr for a field reference likely returns the value or Null? If field doesn't exist, it might return Null. Null is falsy. So objects without the field should also match. That would include third object (no flag) and maybe second? second has flag true -> truthy -> !truthy = false. So first and third match? Actually third has no flag, so eval returns Null (falsy) -> !falsy = true? Wait: logic: !@.flag: first evaluate inner: @.flag returns value or Null if absent. Then isFalsy(val) returns true if falsy. Then NOT isFalsy? Actually the code: if strings.HasPrefix(expr, "!") && ... then val := jp.evalFilterExpr(obj, inner); return isFalsy(val). That means ! operator returns isFalsy(val). So !@.flag returns true if the field is falsy (including absent). So objects with flag false or absent will match. That includes first (flag=false) and third (no flag). Second (flag=true) does not match. Fourth has no flag? It has text, not flag, so also absent -> matches as well. That would be 3 matches. But we need to be careful: our array has four objects: 1: flag=false, num=0; 2: flag=true, num=1; 3: text=""; 4: text="hello". So objects 1,3,4 should match (no flag or falsy). That's 3. Let's test.

	expectedCount := 3
	if len(results) != expectedCount {
		t.Errorf("Filter !@.flag expected %d results, got %d", expectedCount, len(results))
	}
}

// TestJSONPathFilterAbsentOperator tests the "absent" operator: @.field absent
func TestJSONPathFilterAbsentOperator(t *testing.T) {
	// Objects with and without "exists" field
	objs := []objects.Object{
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("exists").HashKey(): {Key: objects.NewString("exists"), Value: Bool(true)},
		}),
		objects.NewMap(map[objects.HashKey]objects.MapPair{
			// no "exists" field
		}),
	}
	arr := objects.NewArray(objs)

	// Filter: $[?(@.exists absent)] should select objects where "exists" field does NOT exist
	jp, _ := ParseJSONPath("$[?(@.exists absent)]")
	results := jp.Get(arr)

	// Should match only second object
	if len(results) != 1 {
		t.Errorf("Expected 1 result for absent operator, got %d", len(results))
	}
	// Verify it's the second object (no field)
	if _, ok := results[0].(*objects.Map); ok {
		// Check that the map doesn't have "exists"
		m := results[0].(*objects.Map)
		key := objects.NewString("exists").HashKey()
		if _, exists := m.Pairs[key]; exists {
			t.Error("Result should not have 'exists' field")
		}
	} else {
		t.Error("Result should be a map")
	}
}

// TestJSONPathGetWithPathOnComplexStructure tests GetWithPath with a mix of segments including wildcard and recursive
func TestJSONPathGetWithPathComplex(t *testing.T) {
	// Build a moderately complex nested structure
	leaf1 := objects.NewString("leaf1")
	leaf2 := objects.NewString("leaf2")
	innerMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("x").HashKey(): {Key: objects.NewString("x"), Value: leaf1},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey():   {Key: objects.NewString("a"), Value: innerMap},
		objects.NewString("arr").HashKey(): {Key: objects.NewString("arr"), Value: objects.NewArray([]objects.Object{leaf2})},
	})

	// GetWithPath with "$.a.x" should yield one match with path "$.a.x"
	jp1, _ := ParseJSONPath("$.a.x")
	matches1 := jp1.GetWithPath(root)
	if len(matches1) != 1 {
		t.Errorf("Expected 1 match for $.a.x, got %d", len(matches1))
	}
	if matches1[0].Path != "$.a.x" || matches1[0].Value != leaf1 {
		t.Errorf("Unexpected match: %s, %v", matches1[0].Path, matches1[0].Value)
	}

	// GetWithPath with "$.arr[*]" should yield one match for the array element
	jp2, _ := ParseJSONPath("$.arr[*]")
	matches2 := jp2.GetWithPath(root)
	if len(matches2) != 1 {
		t.Errorf("Expected 1 match for $.arr[*], got %d", len(matches2))
	}
	if matches2[0].Value != leaf2 {
		t.Errorf("Expected leaf2, got %v", matches2[0].Value)
	}
}

// TestJSONPathRecursiveWildcard tests getRecursiveWildcard via recursive wildcard segment
func TestJSONPathRecursiveWildcard(t *testing.T) {
	// Build nested structure: {"a": {"b": {"c": 1}}}
	c := objects.NewInt(1)
	bMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("c").HashKey(): {Key: objects.NewString("c"), Value: c},
	})
	aMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: bMap},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: aMap},
	})

	// Use recursive wildcard: $..* should return all values recursively
	jp, _ := ParseJSONPath("$..*")
	results := jp.Get(root)

	// Should include aMap, bMap, c, and maybe root? Actually Get starts with root and then applies segments. First segment is recursive_wildcard? The path "$..*" has segments: root, recursive_wildcard. It will start with [root] then apply recursive_wildcard to root, which collects all nested values. So results should include aMap, bMap, c. Possibly also root? No, root is the starting point but after applying segment we get values from root's fields. So we should get aMap and bMap and c? Let's see: recursive_wildcard on root: switch type root is Map -> iterate pairs, append values (aMap). Then recurse into aMap: that'll add bMap, then recurse into bMap: add c. So results: aMap, bMap, c. That's 3.
	if len(results) != 3 {
		t.Errorf("Expected 3 results from recursive wildcard, got %d", len(results))
	}
	// Check that c is present
	foundC := false
	for _, r := range results {
		if r == c {
			foundC = true
			break
		}
	}
	if !foundC {
		t.Error("Recursive wildcard should include the deepest integer value")
	}
}

// TestJSONPathCollectRecursiveFieldWithPath tests collectRecursiveFieldWithPath via GetWithPath with recursive field
func TestJSONPathCollectRecursiveFieldWithPath(t *testing.T) {
	// Similar to recursive field test but ensures the path tracking works
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("inner")},
	})
	outer := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey():  {Key: objects.NewString("name"), Value: objects.NewString("outer")},
		objects.NewString("child").HashKey(): {Key: objects.NewString("child"), Value: inner},
	})

	jp, _ := ParseJSONPath("$..name")
	matches := jp.GetWithPath(outer)

	// Should find both "outer" and "inner" with correct paths
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
	pathVals := make(map[string]string)
	for _, m := range matches {
		if s, ok := m.Value.(*objects.String); ok {
			pathVals[m.Path] = s.Value
		}
	}
	if val, ok := pathVals["$.name"]; !ok || val != "outer" {
		t.Error("Missing or incorrect $.name")
	}
	if val, ok := pathVals["$.child.name"]; !ok || val != "inner" {
		t.Error("Missing or incorrect $.child.name")
	}
}

// TestJSONPathCollectRecursiveWildcardWithPath tests collectRecursiveWildcardWithPath via GetWithPath with recursive wildcard
func TestJSONPathCollectRecursiveWildcardWithPath(t *testing.T) {
	// Simple structure: {"a": {"b": 1}}
	bInt := objects.NewInt(1)
	bMap := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: bInt},
	})
	root := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: bMap},
	})

	jp, _ := ParseJSONPath("$..*")
	matches := jp.GetWithPath(root)

	// Should include bMap (value of a) and bInt. That's 2.
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches from recursive wildcard, got %d", len(matches))
	}
	// Check that bInt is present with path "$.a.b"
	found := false
	for _, m := range matches {
		if m.Value == bInt {
			if m.Path == "$.a.b" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected match with path $.a.b and value 1")
	}
	// Also check that bMap is present with path "$.a"
	foundMap := false
	for _, m := range matches {
		if m.Value == bMap {
			if m.Path == "$.a" {
				foundMap = true
				break
			}
		}
	}
	if !foundMap {
		t.Error("Expected match with path $.a and the intermediate map")
	}
}
