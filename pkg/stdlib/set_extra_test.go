// pkg/stdlib/set_extra_test.go
// Additional argument validation tests for set module.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callSetFunc calls a function from the set module.
func callSetFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("set")
	if mod == nil {
		t := &testing.T{}
		t.Skip("set module not found")
		return &objects.Error{Message: "set module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestSetCreate_ArgumentValidation tests create argument validation.
func TestSetCreate_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, false},                  // no args ok
		{[]objects.Object{Int(10)}, false},           // valid capacity
		{[]objects.Object{Int(2)}, false},            // capacity < 4 creates empty set
		{[]objects.Object{String("not int")}, false}, // non-int treated as element (will be added)
		{[]objects.Object{Int(10), Int(20)}, false},  // multiple elements: creates set with those elements
	}
	for _, tt := range tests {
		result := callSetFunc("create", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("create(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("create(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSetFromArray_ArgumentValidation tests fromArray argument validation.
func TestSetFromArray_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("not array")}, true},
		{[]objects.Object{Array(Int(1), Int(2))}, false},
	}
	for _, tt := range tests {
		result := callSetFunc("fromArray", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("fromArray(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("fromArray(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSetIsSet_TypeCheck tests isSet with various types.
func TestSetIsSet_TypeCheck(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn, ok := mod.Exports["isSet"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSet not found or not builtin")
	}

	// Create a Set object
	s := objects.NewSet()
	res := fn.Fn(s)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSet should return true for Set, got %T %v", res, res)
	}

	// Test with non-Set objects
	nonSetTypes := []objects.Object{
		String("not a set"),
		Int(123),
		Bool(false),
		Null(),
		Array(Int(1)),
	}
	for _, obj := range nonSetTypes {
		res2 := fn.Fn(obj)
		if b, ok := res2.(*objects.Bool); !ok || b.Value {
			t.Fatalf("isSet should return false for %T, got %T %v", obj, res2, res2)
		}
	}
}

// TestSetUnion_ArgumentValidation tests union argument validation.
func TestSetUnion_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("set1")}, true},
		{[]objects.Object{String("set1"), String("set2"), String("extra")}, true},
		{[]objects.Object{Array(Int(1)), Array(Int(2))}, true}, // not Set types
	}
	// First create two actual sets
	set1 := objects.NewSet()
	set1.Add(Int(1))
	set2 := objects.NewSet()
	set2.Add(Int(2))
	tests = append(tests, []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{set1, set2}, false},
	}...)
	for _, tt := range tests {
		result := callSetFunc("union", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("union(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("union(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSetIntersect_ArgumentValidation tests intersect argument validation.
func TestSetIntersect_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("set1")}, true},
		{[]objects.Object{String("set1"), String("set2"), String("extra")}, true},
	}
	set1 := objects.NewSet()
	set1.Add(Int(1))
	set2 := objects.NewSet()
	set2.Add(Int(2))
	tests = append(tests, []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{set1, set2}, false},
	}...)
	for _, tt := range tests {
		result := callSetFunc("intersect", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("intersect(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("intersect(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSetDifference_ArgumentValidation tests difference argument validation.
func TestSetDifference_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("set1")}, true},
		{[]objects.Object{String("set1"), String("set2"), String("extra")}, true},
	}
	set1 := objects.NewSet()
	set1.Add(Int(1))
	set2 := objects.NewSet()
	set2.Add(Int(2))
	tests = append(tests, []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{set1, set2}, false},
	}...)
	for _, tt := range tests {
		result := callSetFunc("difference", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("difference(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("difference(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSetSymmetricDiff_ArgumentValidation tests symmetricDiff argument validation.
func TestSetSymmetricDiff_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("set1")}, true},
		{[]objects.Object{String("set1"), String("set2"), String("extra")}, true},
	}
	set1 := objects.NewSet()
	set1.Add(Int(1))
	set2 := objects.NewSet()
	set2.Add(Int(2))
	tests = append(tests, []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{set1, set2}, false},
	}...)
	for _, tt := range tests {
		result := callSetFunc("symmetricDiff", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("symmetricDiff(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("symmetricDiff(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}
