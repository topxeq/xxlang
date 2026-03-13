// pkg/objects/float.go
package objects

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

// Float represents a floating-point value
type Float struct {
	Value float64
}

// Type returns the object type
func (f *Float) Type() ObjectType { return FloatType }

// TypeTag returns the type tag for fast type checking
func (f *Float) TypeTag() TypeTag { return TagFloat }

// Inspect returns the string representation
func (f *Float) Inspect() string { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

// ToBool converts the float to a boolean
func (f *Float) ToBool() *Bool {
	if f.Value == 0.0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%f", f.Value)))
	return HashKey{Type: FloatType, Value: h.Sum64()}
}
