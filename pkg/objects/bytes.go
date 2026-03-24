// pkg/objects/bytes.go
// Bytes is an immutable byte sequence object.
// It represents a read-only []byte and is useful for binary data handling.
package objects

import (
	"bytes"
	"hash/fnv"
	"io"
	"strconv"
)

// Bytes represents an immutable sequence of bytes.
// Unlike BytesBuffer which is mutable, Bytes is read-only after creation.
type Bytes struct {
	Value []byte
}

// Pre-cached common Bytes values
var (
	BYTES_EMPTY = &Bytes{Value: []byte{}}
)

// NewBytes creates a new Bytes object from a byte slice.
// The byte slice is copied to ensure immutability.
func NewBytes(data []byte) *Bytes {
	if len(data) == 0 {
		return BYTES_EMPTY
	}
	// Copy to ensure immutability
	copied := make([]byte, len(data))
	copy(copied, data)
	return &Bytes{Value: copied}
}

// NewBytesFromArray creates a new Bytes object from an integer array.
// Each integer must be in range 0-255.
func NewBytesFromArray(elements []Object) *Bytes {
	if len(elements) == 0 {
		return BYTES_EMPTY
	}
	data := make([]byte, len(elements))
	for i, elem := range elements {
		n, ok := elem.(*Int)
		if !ok {
			return nil
		}
		if n.Value < 0 || n.Value > 255 {
			return nil
		}
		data[i] = byte(n.Value)
	}
	return &Bytes{Value: data}
}

// Type returns the object type.
func (b *Bytes) Type() ObjectType { return BytesType }

// TypeTag returns the type tag for fast type checking.
func (b *Bytes) TypeTag() TypeTag { return TagBytes }

// Inspect returns a string representation of the Bytes.
func (b *Bytes) Inspect() string {
	return "Bytes(len=" + strconv.Itoa(len(b.Value)) + ")"
}

// ToBool converts the Bytes to a boolean (true if non-empty).
func (b *Bytes) ToBool() *Bool {
	if len(b.Value) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns a hash key for the Bytes.
func (b *Bytes) HashKey() HashKey {
	h := fnv.New64a()
	h.Write(b.Value)
	return HashKey{Type: BytesType, Value: h.Sum64()}
}

// Len returns the number of bytes.
func (b *Bytes) Len() int {
	return len(b.Value)
}

// At returns the byte value at the given index.
func (b *Bytes) At(index int) (int64, bool) {
	if index < 0 || index >= len(b.Value) {
		return 0, false
	}
	return int64(b.Value[index]), true
}

// Slice returns a new Bytes from start to end index.
func (b *Bytes) Slice(start, end int) *Bytes {
	if start < 0 {
		start = 0
	}
	if end > len(b.Value) {
		end = len(b.Value)
	}
	if start >= end {
		return BYTES_EMPTY
	}
	return NewBytes(b.Value[start:end])
}

// String returns the bytes as a string (UTF-8 interpretation).
func (b *Bytes) String() string {
	return string(b.Value)
}

// ToArray converts the bytes to an array of integers.
func (b *Bytes) ToArray() *Array {
	elements := make([]Object, len(b.Value))
	for i, v := range b.Value {
		elements[i] = NewInt(int64(v))
	}
	return NewArray(elements)
}

// GetIOReader returns an io.Reader for the bytes.
// This allows Bytes to be used with Reader objects.
func (b *Bytes) GetIOReader() io.Reader {
	return bytes.NewReader(b.Value)
}

// Equal checks if two Bytes objects are equal.
func (b *Bytes) Equal(other *Bytes) bool {
	return bytes.Equal(b.Value, other.Value)
}

// HasPrefix checks if the bytes start with the given prefix.
func (b *Bytes) HasPrefix(prefix *Bytes) bool {
	return bytes.HasPrefix(b.Value, prefix.Value)
}

// HasSuffix checks if the bytes end with the given suffix.
func (b *Bytes) HasSuffix(suffix *Bytes) bool {
	return bytes.HasSuffix(b.Value, suffix.Value)
}

// Contains checks if the bytes contain the given subsequence.
func (b *Bytes) Contains(sub *Bytes) bool {
	return bytes.Contains(b.Value, sub.Value)
}

// Index returns the index of the first occurrence of sub.
// Returns -1 if not found.
func (b *Bytes) Index(sub *Bytes) int {
	return bytes.Index(b.Value, sub.Value)
}

// LastIndex returns the index of the last occurrence of sub.
// Returns -1 if not found.
func (b *Bytes) LastIndex(sub *Bytes) int {
	return bytes.LastIndex(b.Value, sub.Value)
}

// Count counts the number of non-overlapping instances of sub.
func (b *Bytes) Count(sub *Bytes) int {
	return bytes.Count(b.Value, sub.Value)
}

// Repeat returns a new Bytes with the content repeated n times.
func (b *Bytes) Repeat(n int) *Bytes {
	if n <= 0 {
		return BYTES_EMPTY
	}
	return NewBytes(bytes.Repeat(b.Value, n))
}

// Concat concatenates multiple Bytes objects.
func (b *Bytes) Concat(others ...*Bytes) *Bytes {
	totalLen := len(b.Value)
	for _, other := range others {
		totalLen += len(other.Value)
	}
	result := make([]byte, totalLen)
	copy(result, b.Value)
	offset := len(b.Value)
	for _, other := range others {
		copy(result[offset:], other.Value)
		offset += len(other.Value)
	}
	return NewBytes(result)
}

// GetMember returns a member by name for script access.
func (b *Bytes) GetMember(name string) Object {
	switch name {
	case "len":
		return NewInt(int64(len(b.Value)))
	case "toArray":
		return &Builtin{Fn: func(args ...Object) Object {
			return b.ToArray()
		}}
	case "toString":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewString(b.String())
		}}
	case "toStr":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewString(b.String())
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
			end := len(b.Value)
			if len(args) == 2 {
				endVal, ok := args[1].(*Int)
				if !ok {
					return newError("slice() requires integer arguments")
				}
				end = int(endVal.Value)
			}
			return b.Slice(int(start.Value), end)
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
			val, ok := b.At(int(idx.Value))
			if !ok {
				return newError("index out of range")
			}
			return NewInt(val)
		}}
	case "getReader":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewReader(b.GetIOReader())
		}}
	case "hasPrefix":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("hasPrefix() takes exactly 1 argument")
			}
			other, ok := args[0].(*Bytes)
			if !ok {
				return newError("hasPrefix() requires a BYTES argument")
			}
			return &Bool{Value: b.HasPrefix(other)}
		}}
	case "hasSuffix":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("hasSuffix() takes exactly 1 argument")
			}
			other, ok := args[0].(*Bytes)
			if !ok {
				return newError("hasSuffix() requires a BYTES argument")
			}
			return &Bool{Value: b.HasSuffix(other)}
		}}
	case "contains":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("contains() takes exactly 1 argument")
			}
			other, ok := args[0].(*Bytes)
			if !ok {
				return newError("contains() requires a BYTES argument")
			}
			return &Bool{Value: b.Contains(other)}
		}}
	case "index":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("index() takes exactly 1 argument")
			}
			other, ok := args[0].(*Bytes)
			if !ok {
				return newError("index() requires a BYTES argument")
			}
			return NewInt(int64(b.Index(other)))
		}}
	case "count":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("count() takes exactly 1 argument")
			}
			other, ok := args[0].(*Bytes)
			if !ok {
				return newError("count() requires a BYTES argument")
			}
			return NewInt(int64(b.Count(other)))
		}}
	case "repeat":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("repeat() takes exactly 1 argument")
			}
			n, ok := args[0].(*Int)
			if !ok {
				return newError("repeat() requires an integer argument")
			}
			return b.Repeat(int(n.Value))
		}}
	case "concat":
		return &Builtin{Fn: func(args ...Object) Object {
			others := make([]*Bytes, len(args))
			for i, arg := range args {
				other, ok := arg.(*Bytes)
				if !ok {
					return newError("concat() requires BYTES arguments")
				}
				others[i] = other
			}
			return b.Concat(others...)
		}}
	}
	return NULL
}
