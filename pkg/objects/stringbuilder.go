// pkg/objects/stringbuilder.go
package objects

import (
	"strings"
	"sync"
	"unsafe"
)

// StringBuilder is a mutable string builder for efficient string concatenation.
// Unlike regular string concatenation which creates a new string each time,
// StringBuilder uses an internal buffer to accumulate strings efficiently.
type StringBuilder struct {
	mu     sync.Mutex
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
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return "StringBuilder(len=" + intToStr(len(sb.builder.String())) + ")"
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
	sb.mu.Lock()
	defer sb.mu.Unlock()
	n, _ := sb.builder.WriteString(s)
	return n
}

// WriteLine appends a string followed by a newline to the builder.
func (sb *StringBuilder) WriteLine(s string) int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	n, _ := sb.builder.WriteString(s)
	sb.builder.WriteByte('\n')
	return n + 1
}

// String returns the accumulated string.
func (sb *StringBuilder) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.builder.String()
}

// Len returns the current length of the accumulated string.
func (sb *StringBuilder) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.builder.Len()
}

// Clear resets the builder, removing all content.
func (sb *StringBuilder) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.builder.Reset()
}

// Reset is an alias for Clear.
func (sb *StringBuilder) Reset() {
	sb.Clear()
}

// Grow grows the builder's capacity to hold at least n more bytes.
func (sb *StringBuilder) Grow(n int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.builder.Grow(n)
}

// Cap returns the current capacity of the builder's internal buffer.
func (sb *StringBuilder) Cap() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	// strings.Builder doesn't expose Cap directly, but we can check Len
	// after growing. For now, return Len as approximation.
	return sb.builder.Len()
}

// helper for int to string without fmt
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var negative bool
	if n < 0 {
		negative = true
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if negative {
		digits = append(digits, '-')
	}
	// reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
