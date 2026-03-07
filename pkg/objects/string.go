// pkg/objects/string.go
package objects

import (
	"hash/fnv"
)

// String represents a string value
type String struct {
	Value string
}

// Type returns the object type
func (s *String) Type() ObjectType { return StringType }

// Inspect returns the string representation
func (s *String) Inspect() string { return s.Value }

// ToBool converts the string to a boolean
func (s *String) ToBool() *Bool {
	if len(s.Value) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: StringType, Value: h.Sum64()}
}
