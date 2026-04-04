// pkg/objects/chars.go
package objects

import (
	"hash/fnv"
	"strconv"
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
func (c *Chars) Inspect() string { return "Chars(len=" + strconv.Itoa(len(c.Value)) + ")" }

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

// GetMember returns a member by name for script access
func (c *Chars) GetMember(name string) Object {
	switch name {
	case "len":
		return NewInt(int64(len(c.Value)))
	case "toString":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewString(c.String())
		}}
	case "toStr":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewString(c.String())
		}}
	case "toArray":
		return &Builtin{Fn: func(args ...Object) Object {
			elements := make([]Object, len(c.Value))
			for i, r := range c.Value {
				elements[i] = NewString(string(r))
			}
			return NewArray(elements)
		}}
	case "at":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("at() takes exactly 1 argument")
			}
			idx, ok := args[0].(*Int)
			if !ok {
				return newError("at() requires an integer argument")
			}
			val, ok := c.At(int(idx.Value))
			if !ok {
				return newError("index out of range")
			}
			return NewString(val)
		}}
	case "slice":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("slice() takes 1 or 2 arguments")
			}
			start, ok := args[0].(*Int)
			if !ok {
				return newError("slice() requires integer arguments")
			}
			end := len(c.Value)
			if len(args) == 2 {
				endVal, ok := args[1].(*Int)
				if !ok {
					return newError("slice() requires integer arguments")
				}
				end = int(endVal.Value)
			}
			return c.Slice(int(start.Value), end)
		}}
	}
	return NULL
}
