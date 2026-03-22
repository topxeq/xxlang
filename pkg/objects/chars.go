// pkg/objects/chars.go
package objects

import (
	"hash/fnv"
)

// Chars represents a sequence of Unicode code points ([]rune in Go).
// Unlike String which is UTF-8 encoded bytes, Chars operates on character level,
// making it suitable for proper Unicode text processing.
type Chars struct {
	Value []rune
}

// Pre-cached common chars values
var (
	CHARS_EMPTY = &Chars{Value: []rune{}}
)

// NewChars creates a new Chars object from a rune slice
func NewChars(val []rune) *Chars {
	if len(val) == 0 {
		return CHARS_EMPTY
	}
	return &Chars{Value: val}
}

// NewCharsFromString creates a new Chars object from a string
func NewCharsFromString(s string) *Chars {
	if len(s) == 0 {
		return CHARS_EMPTY
	}
	return &Chars{Value: []rune(s)}
}

// Type returns the object type
func (c *Chars) Type() ObjectType { return CharsType }

// TypeTag returns the type tag for fast type checking
func (c *Chars) TypeTag() TypeTag { return TagChars }

// Inspect returns the string representation
func (c *Chars) Inspect() string { return string(c.Value) }

// ToBool converts the chars to a boolean (true if non-empty)
func (c *Chars) ToBool() *Bool {
	if len(c.Value) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (c *Chars) HashKey() HashKey {
	h := fnv.New64a()
	// Hash each rune
	for _, r := range c.Value {
		h.Write([]byte{byte(r >> 24), byte(r >> 16), byte(r >> 8), byte(r)})
	}
	return HashKey{Type: CharsType, Value: h.Sum64()}
}

// String converts Chars back to a Go string
func (c *Chars) String() string {
	return string(c.Value)
}

// Len returns the number of characters (runes)
func (c *Chars) Len() int {
	return len(c.Value)
}

// At returns the character at the given index as a string
func (c *Chars) At(index int) (string, bool) {
	if index < 0 || index >= len(c.Value) {
		return "", false
	}
	return string(c.Value[index]), true
}

// Slice returns a new Chars from start to end index
func (c *Chars) Slice(start, end int) *Chars {
	if start < 0 {
		start = 0
	}
	if end > len(c.Value) {
		end = len(c.Value)
	}
	if start >= end {
		return CHARS_EMPTY
	}
	return NewChars(c.Value[start:end])
}
