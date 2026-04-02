// pkg/objects/set_test.go
package objects

import (
	"testing"
)

func TestNewSet(t *testing.T) {
	s := NewSet()
	if s == nil {
		t.Fatal("expected set instance")
	}
	if s.Len() != 0 {
		t.Errorf("expected empty set, got %d elements", s.Len())
	}
}

func TestNewSetWithCapacity(t *testing.T) {
	s := NewSetWithCapacity(10)
	if s == nil {
		t.Fatal("expected set instance")
	}
}

func TestNewSetFrom(t *testing.T) {
	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	s := NewSetFrom(arr)
	if s == nil {
		t.Fatal("expected set instance")
	}
	if s.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", s.Len())
	}
}

func TestSetAdd(t *testing.T) {
	s := NewSet()
	added := s.Add(NewInt(1))
	if !added {
		t.Error("expected Add to return true for new element")
	}
	added = s.Add(NewInt(2))
	if !added {
		t.Error("expected Add to return true for new element")
	}
	added = s.Add(NewInt(1)) // duplicate
	if added {
		t.Error("expected Add to return false for duplicate")
	}

	if s.Len() != 2 {
		t.Errorf("expected 2 unique elements, got %d", s.Len())
	}
}

func TestSetContains(t *testing.T) {
	s := NewSet()
	s.Add(NewInt(1))

	if !s.Contains(NewInt(1)) {
		t.Error("expected to contain 1")
	}
	if s.Contains(NewInt(2)) {
		t.Error("expected not to contain 2")
	}
}

func TestSetRemove(t *testing.T) {
	s := NewSet()
	s.Add(NewInt(1))
	s.Add(NewInt(2))

	removed := s.Remove(NewInt(1))
	if !removed {
		t.Error("expected Remove to return true")
	}
	if s.Len() != 1 {
		t.Errorf("expected 1 element after remove, got %d", s.Len())
	}

	removed = s.Remove(NewInt(999))
	if removed {
		t.Error("expected Remove to return false for non-existent element")
	}
}

func TestSetEquals(t *testing.T) {
	s1 := NewSet()
	s1.Add(NewInt(1))
	s1.Add(NewInt(2))

	s2 := NewSet()
	s2.Add(NewInt(1))
	s2.Add(NewInt(2))

	if !s1.Equals(s2) {
		t.Error("expected sets to be equal")
	}

	s3 := NewSet()
	s3.Add(NewInt(1))
	if s1.Equals(s3) {
		t.Error("expected sets to be different")
	}
}

func TestSetIsEmpty(t *testing.T) {
	s := NewSet()
	if !s.IsEmpty() {
		t.Error("expected new set to be empty")
	}
	s.Add(NewInt(1))
	if s.IsEmpty() {
		t.Error("expected set with element to not be empty")
	}
}

func TestSetClear(t *testing.T) {
	s := NewSet()
	s.Add(NewInt(1))
	s.Add(NewInt(2))
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("expected empty set after Clear, got %d", s.Len())
	}
}

func TestSetToArray(t *testing.T) {
	s := NewSet()
	s.Add(NewInt(1))
	s.Add(NewInt(2))
	arr := s.ToArray()
	if arr == nil {
		t.Fatal("expected array")
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}
}

func TestSetClone(t *testing.T) {
	s := NewSet()
	s.Add(NewInt(1))
	cloned := s.Clone()
	if !s.Equals(cloned) {
		t.Error("expected cloned set to equal original")
	}
}

func TestSetUnion(t *testing.T) {
	s1 := NewSet()
	s1.Add(NewInt(1))
	s1.Add(NewInt(2))

	s2 := NewSet()
	s2.Add(NewInt(2))
	s2.Add(NewInt(3))

	union := s1.Union(s2)
	if union.Len() != 3 {
		t.Errorf("expected 3 elements in union, got %d", union.Len())
	}
}

func TestSetIntersect(t *testing.T) {
	s1 := NewSet()
	s1.Add(NewInt(1))
	s1.Add(NewInt(2))

	s2 := NewSet()
	s2.Add(NewInt(2))
	s2.Add(NewInt(3))

	intersect := s1.Intersect(s2)
	if intersect.Len() != 1 {
		t.Errorf("expected 1 element in intersection, got %d", intersect.Len())
	}
}

func TestSetDifference(t *testing.T) {
	s1 := NewSet()
	s1.Add(NewInt(1))
	s1.Add(NewInt(2))

	s2 := NewSet()
	s2.Add(NewInt(2))
	s2.Add(NewInt(3))

	diff := s1.Difference(s2)
	if diff.Len() != 1 {
		t.Errorf("expected 1 element in difference, got %d", diff.Len())
	}
}

func TestSetIsSubset(t *testing.T) {
	s1 := NewSet()
	s1.Add(NewInt(1))

	s2 := NewSet()
	s2.Add(NewInt(1))
	s2.Add(NewInt(2))

	if !s1.IsSubset(s2) {
		t.Error("expected s1 to be subset of s2")
	}
	if s2.IsSubset(s1) {
		t.Error("expected s2 not to be subset of s1")
	}
}
