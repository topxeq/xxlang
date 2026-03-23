// pkg/objects/stringbuilder.go
// StringBuilder is a mutable string builder for efficient string concatenation.
// Unlike regular string concatenation which creates a new string each time,
// StringBuilder uses an internal buffer to accumulate strings efficiently.
// NOTE: This type is NOT thread-safe. Users must handle concurrency themselves
// using Mutex or RWMutex if shared across goroutines.
package objects

import (
	"strconv"
	"strings"
	"unsafe"
)

// StringBuilder is a mutable string builder for efficient string concatenation.
// Not thread-safe - use external synchronization if needed.
type StringBuilder struct {
	builder strings.Builder
}

// NewStringBuilder creates a new StringBuilder instance.
func NewStringBuilder() *StringBuilder {
	return &StringBuilder{
		builder: strings.Builder{},
	}
}

// Type returns the object type.
func (sb *StringBuilder) Type() ObjectType { return StringBuilderType }

// TypeTag returns the fast type tag.
func (sb *StringBuilder) TypeTag() TypeTag { return TagStringBuilder }

// Inspect returns a string representation.
func (sb *StringBuilder) Inspect() string {
	return "StringBuilder(len=" + strconv.Itoa(sb.builder.Len()) + ")"
}

// ToBool returns true (StringBuilder is always truthy).
func (sb *StringBuilder) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the StringBuilder.
func (sb *StringBuilder) HashKey() HashKey {
	return HashKey{
		Type:  StringBuilderType,
		Value: uint64(uintptr(unsafe.Pointer(sb))),
	}
}

// Write appends a string to the builder.
func (sb *StringBuilder) Write(s string) int {
	n, _ := sb.builder.WriteString(s)
	return n
}

// WriteLine appends a string followed by a newline to the builder.
func (sb *StringBuilder) WriteLine(s string) int {
	n, _ := sb.builder.WriteString(s)
	sb.builder.WriteByte('\n')
	return n + 1
}

// String returns the accumulated string.
func (sb *StringBuilder) String() string {
	return sb.builder.String()
}

// Len returns the current length of the accumulated string.
func (sb *StringBuilder) Len() int {
	return sb.builder.Len()
}

// Cap returns the current capacity of the builder's internal buffer.
func (sb *StringBuilder) Cap() int {
	return sb.builder.Cap()
}

// Clear resets the builder, removing all content.
func (sb *StringBuilder) Clear() {
	sb.builder.Reset()
}

// Reset is an alias for Clear.
func (sb *StringBuilder) Reset() {
	sb.builder.Reset()
}

// Grow grows the builder's capacity to hold at least n more bytes.
func (sb *StringBuilder) Grow(n int) {
	sb.builder.Grow(n)
}
