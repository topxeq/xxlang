// pkg/objects/array.go
package objects

import (
	"bytes"
)

// Array represents an array value
type Array struct {
	Elements   []Object
	LastPopped Object // Used by pop() to return the popped value
}

// Type returns the object type
func (a *Array) Type() ObjectType { return ArrayType }

// TypeTag returns the type tag for fast type checking
func (a *Array) TypeTag() TypeTag { return TagArray }

// Inspect returns the string representation
func (a *Array) Inspect() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, e := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(e.Inspect())
	}
	out.WriteString("]")
	return out.String()
}

// ToBool converts the array to a boolean
func (a *Array) ToBool() *Bool {
	if len(a.Elements) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (a *Array) HashKey() HashKey {
	// Arrays are not hashable
	return HashKey{Type: ArrayType, Value: 0}
}
