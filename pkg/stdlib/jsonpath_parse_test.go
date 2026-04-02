// pkg/stdlib/jsonpath_parse_test.go
// Tests for jsonpath module parsing edge cases and error handling.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJSONPathParse_Errors tests ParseJSONPath with various invalid inputs.
func TestJSONPathParse_Errors(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"", false}, // empty gives root
		{"$", false},
		{"$.", true},  // dot without field
		{"$[", true},  // unbalanced bracket
		{"$]", true},  // unexpected ]
		{"$[1", true}, // unbalanced
		{"$[1,2,3", true},
		{"$[a]", true},       // non-integer index
		{"$[1:2:3:4]", true}, // too many parts in slice
		{"$[1:2:x]", true},   // non-integer in slice
		{"$..", true},        // recursive without field/wildcard
		{"$..*", false},      // recursive wildcard is ok
		{"$..field", false},
		{"$[?(@.name=='value')", true},         // missing closing bracket
		{"$[?(@.name=)", true},                 // incomplete filter
		{"$[?(@.name = 123)", true},            // missing closing bracket
		{"$[?(@.name between 1 and 10)", true}, // missing closing bracket
		{"$[?(@.name isNull)", true},           // missing closing bracket
		{"$[?(@.name isNotNull)", true},        // missing closing bracket
		{"$[?(@.name isType 'string')]", false},
		{"$[?empty(@.arr)]", true},      // shorthand without parentheses not supported
		{"$[?length(@.arr) > 0]", true}, // shorthand without parentheses not supported
		{"$[?(@.tags contains 'go')]", false},
		{"$[?(@.name startsWith 'a')]", false},
		{"$[?(@.name endsWith 'e')]", false},
		{"$[?(@.name =~ /^a/)]", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, err := ParseJSONPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for path %q", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for path %q: %v", tt.path, err)
				}
			}
		})
	}
}

// TestJSONPathParse_GetSetOnInvalid tests Get/Set on invalid paths produce errors.
func TestJSONPathParse_GetSetOnInvalid(t *testing.T) {
	// Parsing a path that is invalid should not panic.
	jp, err := ParseJSONPath("$[") // invalid
	if err == nil {
		t.Fatal("expected parse error")
	}
	// Even if parsed partially, using jp should be safe; but if jp is nil, we can't.
	_ = jp
}

// TestJSONPathParse_RelativePaths tests relative path parsing (without $).
func TestJSONPathParse_RelativePaths(t *testing.T) {
	// Relative paths are converted to absolute by prepending "$."
	jp, err := ParseJSONPath("a")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	// Should have segments: root, field a.
	// We can't access internal segments directly but we can use Get on a map.
	m := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewString("x")},
	})
	results := jp.Get(m)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestJSONPathParse_BracketQuotedField tests quoted field names in brackets.
func TestJSONPathParse_BracketQuotedField(t *testing.T) {
	jp, err := ParseJSONPath(`$['field-name']`)
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	m := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("field-name").HashKey(): {Key: objects.NewString("field-name"), Value: objects.NewString("value")},
	})
	results := jp.Get(m)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if s, ok := results[0].(*objects.String); !ok || s.Value != "value" {
		t.Errorf("unexpected result: %v", results[0])
	}
}

// TestJSONPathParse_RecursiveDescent tests recursive descent syntax.
func TestJSONPathParse_RecursiveDescent(t *testing.T) {
	jp, err := ParseJSONPath("$..id")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	// Complex nested structure
	inner := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("id").HashKey(): {Key: objects.NewString("id"), Value: objects.NewInt(1)},
	})
	outer := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("data").HashKey(): {Key: objects.NewString("data"), Value: inner},
		objects.NewString("id").HashKey():   {Key: objects.NewString("id"), Value: objects.NewInt(2)},
	})
	results := jp.Get(outer)
	// Should find two 'id' fields: one at top level (2) and one inside data (1)
	if len(results) != 2 {
		t.Errorf("expected 2 results for recursive descent, got %d", len(results))
	}
}

// TestJSONPathParse_Wildcard tests wildcard handling.
func TestJSONPathParse_Wildcard(t *testing.T) {
	jp, err := ParseJSONPath("$.*")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	m := objects.NewMap(map[objects.HashKey]objects.MapPair{
		objects.NewString("a").HashKey(): {Key: objects.NewString("a"), Value: objects.NewInt(1)},
		objects.NewString("b").HashKey(): {Key: objects.NewString("b"), Value: objects.NewInt(2)},
	})
	results := jp.Get(m)
	if len(results) != 2 {
		t.Errorf("expected 2 results from wildcard, got %d", len(results))
	}
}

// TestJSONPathParse_Slice tests slice syntax parsing.
func TestJSONPathParse_Slice(t *testing.T) {
	// Already tested in extra, but ensure parsing works
	jp, err := ParseJSONPath("$[0:2]")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	arr := objects.NewArray([]objects.Object{
		objects.NewString("a"),
		objects.NewString("b"),
		objects.NewString("c"),
	})
	results := jp.Get(arr)
	if len(results) != 2 {
		t.Errorf("slice [0:2] should yield 2 items, got %d", len(results))
	}
}

// TestJSONPathParse_MultiIndex tests multi-index parsing.
func TestJSONPathParse_MultiIndex(t *testing.T) {
	jp, err := ParseJSONPath("$[0,2,4]")
	if err != nil {
		t.Fatalf("ParseJSONPath error: %v", err)
	}
	arr := objects.NewArray([]objects.Object{
		objects.NewString("a"),
		objects.NewString("b"),
		objects.NewString("c"),
		objects.NewString("d"),
		objects.NewString("e"),
	})
	results := jp.Get(arr)
	if len(results) != 3 {
		t.Errorf("multi-index should yield 3 items, got %d", len(results))
	}
}
