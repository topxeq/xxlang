// pkg/objects/int.go
package objects

import (
	"strconv"
)

// Int represents an integer value
type Int struct {
	Value int64
}

// Type returns the object type
func (i *Int) Type() ObjectType { return IntType }

// Inspect returns the string representation
func (i *Int) Inspect() string { return strconv.FormatInt(i.Value, 10) }

// ToBool converts the integer to a boolean
func (i *Int) ToBool() *Bool {
	if i.Value == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (i *Int) HashKey() HashKey {
	return HashKey{Type: IntType, Value: uint64(i.Value)}
}
