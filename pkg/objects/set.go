// pkg/objects/set.go
// Set type for Xxlang - unordered collection of unique elements.
package objects

import (
	"sort"
	"strconv"
	"unsafe"
)

// Set represents an unordered collection of unique elements (not thread-safe).
type Set struct {
	items map[string]Object // key is the Inspect() representation
}

// NewSet creates a new empty set.
func NewSet() *Set {
	return &Set{
		items: make(map[string]Object),
	}
}

// NewSetWithCapacity creates a new set with the specified initial capacity.
func NewSetWithCapacity(capacity int) *Set {
	return &Set{
		items: make(map[string]Object, capacity),
	}
}

// NewSetFrom creates a new set from an array.
func NewSetFrom(arr *Array) *Set {
	s := NewSetWithCapacity(len(arr.Elements))
	for _, elem := range arr.Elements {
		s.Add(elem)
	}
	return s
}

// Type returns the object type.
func (s *Set) Type() ObjectType { return SetType }

// TypeTag returns the fast type tag.
func (s *Set) TypeTag() TypeTag { return TagSet }

// Inspect returns a string representation.
func (s *Set) Inspect() string {
	return "Set(len=" + strconv.Itoa(len(s.items)) + ")"
}

// ToBool returns true (Set is always truthy).
func (s *Set) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the Set.
func (s *Set) HashKey() HashKey {
	return HashKey{
		Type:  SetType,
		Value: uint64(uintptr(unsafe.Pointer(s))),
	}
}

// Add adds an element to the set. Returns true if added, false if already present.
func (s *Set) Add(item Object) bool {
	key := item.Inspect()
	if _, exists := s.items[key]; exists {
		return false
	}
	s.items[key] = item
	return true
}

// Remove removes an element from the set. Returns true if removed, false if not found.
func (s *Set) Remove(item Object) bool {
	key := item.Inspect()
	if _, exists := s.items[key]; !exists {
		return false
	}
	delete(s.items, key)
	return true
}

// Contains returns true if the set contains the element.
func (s *Set) Contains(item Object) bool {
	_, exists := s.items[item.Inspect()]
	return exists
}

// Len returns the number of elements in the set.
func (s *Set) Len() int {
	return len(s.items)
}

// IsEmpty returns true if the set is empty.
func (s *Set) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear removes all elements from the set.
func (s *Set) Clear() {
	s.items = make(map[string]Object)
}

// ToArray returns all elements as an array.
func (s *Set) ToArray() *Array {
	elements := make([]Object, 0, len(s.items))
	for _, item := range s.items {
		elements = append(elements, item)
	}
	return &Array{Elements: elements}
}

// ToSortedArray returns all elements as an array, sorted by Inspect() representation.
func (s *Set) ToSortedArray() *Array {
	// Collect and sort keys for deterministic order
	keys := make([]string, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	elements := make([]Object, 0, len(s.items))
	for _, k := range keys {
		elements = append(elements, s.items[k])
	}
	return &Array{Elements: elements}
}

// Clone returns a shallow copy of the set.
func (s *Set) Clone() *Set {
	newSet := &Set{
		items: make(map[string]Object, len(s.items)),
	}
	for k, v := range s.items {
		newSet.items[k] = v
	}
	return newSet
}

// Union returns a new set containing all elements from both sets.
func (s *Set) Union(other *Set) *Set {
	result := NewSetWithCapacity(len(s.items) + len(other.items))
	for k, v := range s.items {
		result.items[k] = v
	}
	for k, v := range other.items {
		result.items[k] = v
	}
	return result
}

// Intersect returns a new set containing elements present in both sets.
func (s *Set) Intersect(other *Set) *Set {
	// Iterate over the smaller set
	if len(s.items) > len(other.items) {
		s, other = other, s
	}

	result := NewSetWithCapacity(len(s.items))
	for k, v := range s.items {
		if _, exists := other.items[k]; exists {
			result.items[k] = v
		}
	}
	return result
}

// Difference returns a new set containing elements in s but not in other.
func (s *Set) Difference(other *Set) *Set {
	result := NewSetWithCapacity(len(s.items))
	for k, v := range s.items {
		if _, exists := other.items[k]; !exists {
			result.items[k] = v
		}
	}
	return result
}

// SymmetricDifference returns a new set containing elements in either set but not both.
func (s *Set) SymmetricDifference(other *Set) *Set {
	result := NewSetWithCapacity(len(s.items) + len(other.items))
	for k, v := range s.items {
		if _, exists := other.items[k]; !exists {
			result.items[k] = v
		}
	}
	for k, v := range other.items {
		if _, exists := s.items[k]; !exists {
			result.items[k] = v
		}
	}
	return result
}

// IsSubset returns true if s is a subset of other.
func (s *Set) IsSubset(other *Set) bool {
	for k := range s.items {
		if _, exists := other.items[k]; !exists {
			return false
		}
	}
	return true
}

// IsSuperset returns true if s is a superset of other.
func (s *Set) IsSuperset(other *Set) bool {
	return other.IsSubset(s)
}

// Equals returns true if both sets contain the same elements.
func (s *Set) Equals(other *Set) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for k := range s.items {
		if _, exists := other.items[k]; !exists {
			return false
		}
	}
	return true
}