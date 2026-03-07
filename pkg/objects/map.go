// pkg/objects/map.go
package objects

import (
	"bytes"
)

// MapPair represents a key-value pair in a map
type MapPair struct {
	Key   Object
	Value Object
}

// Map represents a map value
type Map struct {
	Pairs map[HashKey]MapPair
}

// Type returns the object type
func (m *Map) Type() ObjectType { return MapType }

// Inspect returns the string representation
func (m *Map) Inspect() string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for _, pair := range m.Pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(pair.Key.Inspect())
		out.WriteString(": ")
		out.WriteString(pair.Value.Inspect())
	}
	out.WriteString("}")
	return out.String()
}

// ToBool converts the map to a boolean
func (m *Map) ToBool() *Bool {
	if len(m.Pairs) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (m *Map) HashKey() HashKey {
	// Maps are not hashable
	return HashKey{Type: MapType, Value: 0}
}
