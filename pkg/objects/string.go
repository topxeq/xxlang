// pkg/objects/string.go
package objects

import (
	"hash/fnv"
	"sync"
)

// String represents a string value
type String struct {
	Value string
}

// Type returns the object type
func (s *String) Type() ObjectType { return StringType }

// TypeTag returns the type tag for fast type checking
func (s *String) TypeTag() TypeTag { return TagString }

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

// StringIntern is a simple string interning pool
// Caches commonly used strings to reduce allocations
var (
	stringIntern sync.Map // map[string]*String
)

// Intern returns a cached *String for the given value
// This reduces allocations for frequently used strings
func InternString(val string) *String {
	// Check cache first
	if cached, ok := stringIntern.Load(val); ok {
		return cached.(*String)
	}

	// Create new string object
	s := &String{Value: val}

	// Store in cache (loadOrStore to handle race condition)
	if actual, loaded := stringIntern.LoadOrStore(val, s); loaded {
		return actual.(*String)
	}
	return s
}

// InternBatch pre-interns a batch of strings
func InternBatch(strings []string) []*String {
	result := make([]*String, len(strings))
	for i, s := range strings {
		result[i] = InternString(s)
	}
	return result
}
