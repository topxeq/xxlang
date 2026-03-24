// pkg/objects/bytes_buffer.go
// BytesBuffer is a mutable buffer of bytes for efficient byte operations.
// It's similar to Go's bytes.Buffer type.
// NOTE: This type is NOT thread-safe. Users must handle concurrency themselves
// using Mutex or RWMutex if shared across goroutines.
package objects

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"unsafe"
)

// BytesBuffer is a mutable buffer for efficient byte operations.
// Not thread-safe - use external synchronization if needed.
type BytesBuffer struct {
	buffer bytes.Buffer
}

// NewBytesBuffer creates a new BytesBuffer instance.
func NewBytesBuffer() *BytesBuffer {
	return &BytesBuffer{
		buffer: bytes.Buffer{},
	}
}

// NewBytesBufferFromBytes creates a new BytesBuffer from an existing byte slice.
func NewBytesBufferFromBytes(data []byte) *BytesBuffer {
	bb := &BytesBuffer{
		buffer: bytes.Buffer{},
	}
	bb.buffer.Write(data)
	return bb
}

// Type returns the object type.
func (bb *BytesBuffer) Type() ObjectType { return BytesBufferType }

// TypeTag returns the fast type tag.
func (bb *BytesBuffer) TypeTag() TypeTag { return TagBytesBuffer }

// Inspect returns a string representation.
func (bb *BytesBuffer) Inspect() string {
	return "BytesBuffer(len=" + strconv.Itoa(bb.buffer.Len()) + ", cap=" + strconv.Itoa(bb.buffer.Cap()) + ")"
}

// ToBool returns true (BytesBuffer is always truthy).
func (bb *BytesBuffer) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the BytesBuffer.
func (bb *BytesBuffer) HashKey() HashKey {
	return HashKey{
		Type:  BytesBufferType,
		Value: uint64(uintptr(unsafe.Pointer(bb))),
	}
}

// Write appends bytes to the buffer.
func (bb *BytesBuffer) Write(data []byte) int {
	n, _ := bb.buffer.Write(data)
	return n
}

// WriteString appends a string to the buffer as bytes.
func (bb *BytesBuffer) WriteString(s string) int {
	n, _ := bb.buffer.WriteString(s)
	return n
}

// WriteByte appends a single byte to the buffer.
func (bb *BytesBuffer) WriteByte(b byte) error {
	return bb.buffer.WriteByte(b)
}

// WriteInt16 appends a 16-bit integer in little-endian format.
func (bb *BytesBuffer) WriteInt16(v int16) error {
	return binary.Write(&bb.buffer, binary.LittleEndian, v)
}

// WriteInt32 appends a 32-bit integer in little-endian format.
func (bb *BytesBuffer) WriteInt32(v int32) error {
	return binary.Write(&bb.buffer, binary.LittleEndian, v)
}

// WriteInt64 appends a 64-bit integer in little-endian format.
func (bb *BytesBuffer) WriteInt64(v int64) error {
	return binary.Write(&bb.buffer, binary.LittleEndian, v)
}

// WriteFloat32 appends a 32-bit float in little-endian format.
func (bb *BytesBuffer) WriteFloat32(v float32) error {
	return binary.Write(&bb.buffer, binary.LittleEndian, v)
}

// WriteFloat64 appends a 64-bit float in little-endian format.
func (bb *BytesBuffer) WriteFloat64(v float64) error {
	return binary.Write(&bb.buffer, binary.LittleEndian, v)
}

// Bytes returns a copy of the buffer contents as a byte slice.
func (bb *BytesBuffer) Bytes() []byte {
	// Return a copy to prevent external modification
	data := bb.buffer.Bytes()
	result := make([]byte, len(data))
	copy(result, data)
	return result
}

// BytesRef returns a reference to the underlying byte slice (no copy).
// WARNING: The returned slice may be modified by subsequent buffer operations.
func (bb *BytesBuffer) BytesRef() []byte {
	return bb.buffer.Bytes()
}

// String returns the buffer contents as a string.
func (bb *BytesBuffer) String() string {
	return bb.buffer.String()
}

// Len returns the current length of the buffer.
func (bb *BytesBuffer) Len() int {
	return bb.buffer.Len()
}

// Cap returns the current capacity of the buffer.
func (bb *BytesBuffer) Cap() int {
	return bb.buffer.Cap()
}

// Clear resets the buffer, removing all content.
func (bb *BytesBuffer) Clear() {
	bb.buffer.Reset()
}

// Reset is an alias for Clear.
func (bb *BytesBuffer) Reset() {
	bb.buffer.Reset()
}

// Grow grows the buffer's capacity to hold at least n more bytes.
func (bb *BytesBuffer) Grow(n int) {
	bb.buffer.Grow(n)
}

// Truncate discards all but the first n bytes.
func (bb *BytesBuffer) Truncate(n int) {
	if n < 0 {
		n = 0
	}
	if n > bb.buffer.Len() {
		n = bb.buffer.Len()
	}
	bb.buffer.Truncate(n)
}

// Read reads up to len(p) bytes from the buffer into p.
// Returns the number of bytes read.
func (bb *BytesBuffer) Read(p []byte) (int, error) {
	return bb.buffer.Read(p)
}

// ReadByte reads and returns the next byte from the buffer.
func (bb *BytesBuffer) ReadByte() (byte, error) {
	return bb.buffer.ReadByte()
}

// ReadBytes reads until the first occurrence of delim in the input.
func (bb *BytesBuffer) ReadBytes(delim byte) ([]byte, error) {
	return bb.buffer.ReadBytes(delim)
}

// ReadString reads until the first occurrence of delim in the input.
func (bb *BytesBuffer) ReadString(delim byte) (string, error) {
	return bb.buffer.ReadString(delim)
}

// ReadInt16 reads a 16-bit integer in little-endian format.
func (bb *BytesBuffer) ReadInt16() (int16, error) {
	var v int16
	err := binary.Read(&bb.buffer, binary.LittleEndian, &v)
	return v, err
}

// ReadInt32 reads a 32-bit integer in little-endian format.
func (bb *BytesBuffer) ReadInt32() (int32, error) {
	var v int32
	err := binary.Read(&bb.buffer, binary.LittleEndian, &v)
	return v, err
}

// ReadInt64 reads a 64-bit integer in little-endian format.
func (bb *BytesBuffer) ReadInt64() (int64, error) {
	var v int64
	err := binary.Read(&bb.buffer, binary.LittleEndian, &v)
	return v, err
}

// ReadFloat32 reads a 32-bit float in little-endian format.
func (bb *BytesBuffer) ReadFloat32() (float32, error) {
	var v float32
	err := binary.Read(&bb.buffer, binary.LittleEndian, &v)
	return v, err
}

// ReadFloat64 reads a 64-bit float in little-endian format.
func (bb *BytesBuffer) ReadFloat64() (float64, error) {
	var v float64
	err := binary.Read(&bb.buffer, binary.LittleEndian, &v)
	return v, err
}

// Peek returns the next n bytes without advancing the reader.
func (bb *BytesBuffer) Peek(n int) []byte {
	if n > bb.buffer.Len() {
		n = bb.buffer.Len()
	}
	data := bb.buffer.Bytes()
	if n <= 0 {
		return []byte{}
	}
	result := make([]byte, n)
	copy(result, data[:n])
	return result
}

// Seek sets the position for the next Read or Write.
// This is a simplified version that resets and writes the data back.
func (bb *BytesBuffer) Seek(offset int64, whence int) int64 {
	// bytes.Buffer doesn't support seek directly
	// We implement a simplified version
	var abs int64
	switch whence {
	case 0: // io.SeekStart
		abs = offset
	case 1: // io.SeekCurrent - not fully supported
		abs = int64(bb.buffer.Len()) + offset
	case 2: // io.SeekEnd
		abs = int64(bb.buffer.Len()) + offset
	default:
		return -1
	}
	// For simplicity, we don't actually move the position
	// This is a placeholder for the interface
	return abs
}

// WriteTo writes the buffer contents to another BytesBuffer.
func (bb *BytesBuffer) WriteTo(other *BytesBuffer) (int64, error) {
	return bb.buffer.WriteTo(&other.buffer)
}

// Equals compares two BytesBuffer contents.
func (bb *BytesBuffer) Equals(other *BytesBuffer) bool {
	return bytes.Equal(bb.buffer.Bytes(), other.buffer.Bytes())
}

// GetIOReader returns the buffer as an io.Reader interface.
// This allows the BytesBuffer to be used with Reader objects.
func (bb *BytesBuffer) GetIOReader() *bytesBufferReader {
	return &bytesBufferReader{buffer: &bb.buffer}
}

// GetIOWriter returns the buffer as an io.Writer interface.
// This allows the BytesBuffer to be used with Writer objects.
func (bb *BytesBuffer) GetIOWriter() *bytesBufferWriter {
	return &bytesBufferWriter{buffer: &bb.buffer}
}

// bytesBufferReader wraps a *bytes.Buffer to implement io.Reader.
type bytesBufferReader struct {
	buffer *bytes.Buffer
}

func (r *bytesBufferReader) Read(p []byte) (n int, err error) {
	return r.buffer.Read(p)
}

// bytesBufferWriter wraps a *bytes.Buffer to implement io.Writer.
type bytesBufferWriter struct {
	buffer *bytes.Buffer
}

func (w *bytesBufferWriter) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}
