// pkg/objects/map.go
package objects

import (
	"bytes"
	"sort"
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

// TypeTag returns the type tag for fast type checking
func (m *Map) TypeTag() TypeTag { return TagMap }

// Inspect returns the string representation
func (m *Map) Inspect() string {
	var out bytes.Buffer
	out.WriteString("{")

	// Collect and sort keys for deterministic output
	keys := make([]string, 0, len(m.Pairs))
	pairByKey := make(map[string]MapPair, len(m.Pairs))
	for _, pair := range m.Pairs {
		keyStr := pair.Key.Inspect()
		keys = append(keys, keyStr)
		pairByKey[keyStr] = pair
	}
	sort.Strings(keys)

	first := true
	for _, keyStr := range keys {
		if !first {
			out.WriteString(", ")
		}
		first = false
		pair := pairByKey[keyStr]
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
