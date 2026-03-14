// pkg/objects/int.go
package objects

import (
	"strconv"
)

// Int represents an integer value
type Int struct {
	Value int64
}

// IntCacheMin is the minimum cached integer value
const IntCacheMin = -100

// IntCacheMax is the maximum cached integer value
// Extended to cover common loop indices and calculation results
const IntCacheMax = 10000

// intCache holds pre-allocated integers for common values
var intCache [IntCacheMax - IntCacheMin + 1]*Int

// initIntCache pre-allocates common integer values
func initIntCache() {
	for i := IntCacheMin; i <= IntCacheMax; i++ {
		intCache[i-IntCacheMin] = &Int{Value: int64(i)}
	}
}

func init() {
	initIntCache()
}

// NewInt creates a new Int object, using cache for small values
func NewInt(val int64) *Int {
	if val >= IntCacheMin && val <= IntCacheMax {
		return intCache[val-IntCacheMin]
	}
	return &Int{Value: val}
}

// Type returns the object type
func (i *Int) Type() ObjectType { return IntType }

// TypeTag returns the type tag for fast type checking
func (i *Int) TypeTag() TypeTag { return TagInt }

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
