// pkg/stdlib/set_test.go
// Tests for set module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestSetCreate(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}

	fn := mod.Exports["create"].(*objects.Builtin)

	// Create empty set
	s := fn.Fn()
	if s.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", s.Type())
	}
	set := s.(*objects.Set)
	if set.Len() != 0 {
		t.Errorf("expected empty set size 0, got %d", set.Len())
	}

	// Create set with capacity
	s = fn.Fn(objects.NewInt(10))
	if s.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", s.Type())
	}
	set = s.(*objects.Set)
	// Capacity is internal; just check it's still empty
	if set.Len() != 0 {
		t.Errorf("expected empty set with capacity, got size %d", set.Len())
	}

	// Create set with elements
	s = fn.Fn(objects.NewString("a"), objects.NewString("b"), objects.NewInt(1))
	set = s.(*objects.Set)
	if set.Len() != 3 {
		t.Errorf("expected set size 3, got %d", set.Len())
	}
	// Check elements
	if !set.Contains(objects.NewString("a")) {
		t.Error("set missing element 'a'")
	}
	if !set.Contains(objects.NewString("b")) {
		t.Error("set missing element 'b'")
	}
	if !set.Contains(objects.NewInt(1)) {
		t.Error("set missing element 1")
	}
}

func TestSetFromArray(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["fromArray"].(*objects.Builtin)

	arr := objects.NewArray([]objects.Object{objects.NewString("x"), objects.NewString("y"), objects.NewString("x")}) // duplicate x
	s := fn.Fn(arr)
	if s.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", s.Type())
	}
	set := s.(*objects.Set)
	// Should have unique elements: x, y
	if set.Len() != 2 {
		t.Errorf("expected set size 2, got %d", set.Len())
	}
	if !set.Contains(objects.NewString("x")) {
		t.Error("set missing element 'x'")
	}
	if !set.Contains(objects.NewString("y")) {
		t.Error("set missing element 'y'")
	}
}

func TestSetIsSet(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["isSet"].(*objects.Builtin)

	// Non-set object
	result := fn.Fn(objects.NULL)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != false {
			t.Errorf("expected false for NULL, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}

	// Set object
	s := objects.NewSet()
	result = fn.Fn(s)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != true {
			t.Errorf("expected true for Set, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}
}

func TestSetUnion(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["union"].(*objects.Builtin)

	s1 := objects.NewSetWithCapacity(3)
	s1.Add(objects.NewString("a"))
	s1.Add(objects.NewString("b"))

	s2 := objects.NewSetWithCapacity(3)
	s2.Add(objects.NewString("b"))
	s2.Add(objects.NewString("c"))

	result := fn.Fn(s1, s2)
	if result.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", result.Type())
	}
	union := result.(*objects.Set)
	if union.Len() != 3 {
		t.Errorf("expected union size 3, got %d", union.Len())
	}
	if !union.Contains(objects.NewString("a")) {
		t.Error("union missing 'a'")
	}
	if !union.Contains(objects.NewString("b")) {
		t.Error("union missing 'b'")
	}
	if !union.Contains(objects.NewString("c")) {
		t.Error("union missing 'c'")
	}
}

func TestSetIntersect(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["intersect"].(*objects.Builtin)

	s1 := objects.NewSetWithCapacity(3)
	s1.Add(objects.NewString("a"))
	s1.Add(objects.NewString("b"))
	s1.Add(objects.NewString("c"))

	s2 := objects.NewSetWithCapacity(3)
	s2.Add(objects.NewString("b"))
	s2.Add(objects.NewString("c"))
	s2.Add(objects.NewString("d"))

	result := fn.Fn(s1, s2)
	if result.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", result.Type())
	}
	intersect := result.(*objects.Set)
	if intersect.Len() != 2 {
		t.Errorf("expected intersect size 2, got %d", intersect.Len())
	}
	if !intersect.Contains(objects.NewString("b")) {
		t.Error("intersect missing 'b'")
	}
	if !intersect.Contains(objects.NewString("c")) {
		t.Error("intersect missing 'c'")
	}
	if intersect.Contains(objects.NewString("a")) {
		t.Error("intersect should not contain 'a'")
	}
}

func TestSetDifference(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["difference"].(*objects.Builtin)

	s1 := objects.NewSetWithCapacity(3)
	s1.Add(objects.NewString("a"))
	s1.Add(objects.NewString("b"))
	s1.Add(objects.NewString("c"))

	s2 := objects.NewSetWithCapacity(2)
	s2.Add(objects.NewString("b"))
	s2.Add(objects.NewString("c"))

	result := fn.Fn(s1, s2)
	if result.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", result.Type())
	}
	diff := result.(*objects.Set)
	if diff.Len() != 1 {
		t.Errorf("expected difference size 1, got %d", diff.Len())
	}
	if !diff.Contains(objects.NewString("a")) {
		t.Error("difference missing 'a'")
	}
	if diff.Contains(objects.NewString("b")) {
		t.Error("difference should not contain 'b'")
	}
}

func TestSetSymmetricDiff(t *testing.T) {
	mod := Get("set")
	if mod == nil {
		t.Skip("set module not found")
	}
	fn := mod.Exports["symmetricDiff"].(*objects.Builtin)

	s1 := objects.NewSetWithCapacity(3)
	s1.Add(objects.NewString("a"))
	s1.Add(objects.NewString("b"))

	s2 := objects.NewSetWithCapacity(3)
	s2.Add(objects.NewString("b"))
	s2.Add(objects.NewString("c"))

	result := fn.Fn(s1, s2)
	if result.Type() != objects.SetType {
		t.Fatalf("expected Set, got %s", result.Type())
	}
	symDiff := result.(*objects.Set)
	if symDiff.Len() != 2 {
		t.Errorf("expected symmetric difference size 2, got %d", symDiff.Len())
	}
	if !symDiff.Contains(objects.NewString("a")) {
		t.Error("symmetric diff missing 'a'")
	}
	if !symDiff.Contains(objects.NewString("c")) {
		t.Error("symmetric diff missing 'c'")
	}
	if symDiff.Contains(objects.NewString("b")) {
		t.Error("symmetric diff should not contain 'b'")
	}
}
