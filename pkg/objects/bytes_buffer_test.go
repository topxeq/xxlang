// pkg/objects/bytes_buffer_test.go
package objects

import (
	"testing"
)

func TestBytesBufferBasic(t *testing.T) {
	bb := NewBytesBuffer()
	if bb == nil {
		t.Fatal("NewBytesBuffer returned nil")
	}

	if bb.Type() != BytesBufferType {
		t.Errorf("expected type %s, got %s", BytesBufferType, bb.Type())
	}

	if bb.TypeTag() != TagBytesBuffer {
		t.Errorf("expected type tag %d, got %d", TagBytesBuffer, bb.TypeTag())
	}

	if bb.Len() != 0 {
		t.Errorf("expected length 0, got %d", bb.Len())
	}

	if bb.ToBool() != TRUE {
		t.Error("BytesBuffer should be truthy")
	}
}

func TestBytesBufferWriteRead(t *testing.T) {
	bb := NewBytesBuffer()

	// Write string
	n := bb.WriteString("hello")
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}

	if bb.Len() != 5 {
		t.Errorf("expected length 5, got %d", bb.Len())
	}

	// Read string
	s := bb.String()
	if s != "hello" {
		t.Errorf("expected 'hello', got '%s'", s)
	}

	// Clear
	bb.Clear()
	if bb.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", bb.Len())
	}
}

func TestBytesBufferWriteByte(t *testing.T) {
	bb := NewBytesBuffer()

	bb.WriteByte(65) // 'A'
	bb.WriteByte(66) // 'B'
	bb.WriteByte(67) // 'C'

	if bb.Len() != 3 {
		t.Errorf("expected length 3, got %d", bb.Len())
	}

	if bb.String() != "ABC" {
		t.Errorf("expected 'ABC', got '%s'", bb.String())
	}
}

func TestBytesBufferWriteInt(t *testing.T) {
	bb := NewBytesBuffer()

	// Write int64
	err := bb.WriteInt64(123456789)
	if err != nil {
		t.Errorf("WriteInt64 error: %v", err)
	}

	if bb.Len() != 8 {
		t.Errorf("expected length 8, got %d", bb.Len())
	}

	// Read int64
	v, err := bb.ReadInt64()
	if err != nil {
		t.Errorf("ReadInt64 error: %v", err)
	}

	if v != 123456789 {
		t.Errorf("expected 123456789, got %d", v)
	}
}

func TestBytesBufferWriteFloat(t *testing.T) {
	bb := NewBytesBuffer()

	// Write float64
	err := bb.WriteFloat64(3.14159)
	if err != nil {
		t.Errorf("WriteFloat64 error: %v", err)
	}

	if bb.Len() != 8 {
		t.Errorf("expected length 8, got %d", bb.Len())
	}

	// Read float64
	v, err := bb.ReadFloat64()
	if err != nil {
		t.Errorf("ReadFloat64 error: %v", err)
	}

	// Check approximate equality
	if v < 3.14158 || v > 3.14160 {
		t.Errorf("expected ~3.14159, got %f", v)
	}
}

func TestBytesBufferFromBytes(t *testing.T) {
	data := []byte{72, 101, 108, 108, 111} // "Hello"
	bb := NewBytesBufferFromBytes(data)

	if bb.Len() != 5 {
		t.Errorf("expected length 5, got %d", bb.Len())
	}

	if bb.String() != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", bb.String())
	}
}

func TestBytesBufferBytes(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("test")

	data := bb.Bytes()
	if len(data) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(data))
	}

	// Verify it's a copy
	data[0] = 'T'
	if bb.String() != "test" {
		t.Error("Bytes() should return a copy")
	}
}

func TestBytesBufferPeek(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	data := bb.Peek(3)
	if len(data) != 3 {
		t.Errorf("expected 3 bytes, got %d", len(data))
	}

	if string(data) != "hel" {
		t.Errorf("expected 'hel', got '%s'", string(data))
	}

	// Verify buffer unchanged
	if bb.Len() != 5 {
		t.Errorf("Peek should not consume bytes, got length %d", bb.Len())
	}
}

func TestBytesBufferTruncate(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello world")

	bb.Truncate(5)
	if bb.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", bb.String())
	}
}

func TestBytesBufferGrow(t *testing.T) {
	bb := NewBytesBuffer()
	bb.Grow(100)

	// Capacity should be at least 100
	if bb.Cap() < 100 {
		t.Errorf("expected capacity >= 100, got %d", bb.Cap())
	}
}

func TestBytesBufferInspect(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("test")

	inspect := bb.Inspect()
	if inspect == "" {
		t.Error("Inspect should not be empty")
	}
}

func TestBytesBufferHashKey(t *testing.T) {
	bb1 := NewBytesBuffer()
	bb2 := NewBytesBuffer()

	// Different instances should have different hash keys
	if bb1.HashKey() == bb2.HashKey() {
		t.Error("Different BytesBuffer instances should have different hash keys")
	}
}

func TestBytesBufferWrite(t *testing.T) {
	bb := NewBytesBuffer()
	data := []byte{72, 101, 108, 108, 111}

	n := bb.Write(data)
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
}

func TestBytesBufferBytesRef(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	ref := bb.BytesRef()
	if len(ref) != 5 {
		t.Errorf("expected 5 bytes, got %d", len(ref))
	}
}

func TestBytesBufferReset(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	bb.Reset()
	if bb.Len() != 0 {
		t.Errorf("expected length 0 after reset, got %d", bb.Len())
	}
}

func TestBytesBufferWriteInt16(t *testing.T) {
	bb := NewBytesBuffer()

	err := bb.WriteInt16(12345)
	if err != nil {
		t.Errorf("WriteInt16 error: %v", err)
	}

	v, err := bb.ReadInt16()
	if err != nil {
		t.Errorf("ReadInt16 error: %v", err)
	}
	if v != 12345 {
		t.Errorf("expected 12345, got %d", v)
	}
}

func TestBytesBufferWriteInt32(t *testing.T) {
	bb := NewBytesBuffer()

	err := bb.WriteInt32(123456789)
	if err != nil {
		t.Errorf("WriteInt32 error: %v", err)
	}

	v, err := bb.ReadInt32()
	if err != nil {
		t.Errorf("ReadInt32 error: %v", err)
	}
	if v != 123456789 {
		t.Errorf("expected 123456789, got %d", v)
	}
}

func TestBytesBufferWriteFloat32(t *testing.T) {
	bb := NewBytesBuffer()

	err := bb.WriteFloat32(3.14)
	if err != nil {
		t.Errorf("WriteFloat32 error: %v", err)
	}

	v, err := bb.ReadFloat32()
	if err != nil {
		t.Errorf("ReadFloat32 error: %v", err)
	}
	if v < 3.13 || v > 3.15 {
		t.Errorf("expected ~3.14, got %f", v)
	}
}

func TestBytesBufferRead(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	buf := make([]byte, 3)
	n, err := bb.Read(buf)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 bytes read, got %d", n)
	}
	if string(buf) != "hel" {
		t.Errorf("expected 'hel', got '%s'", string(buf))
	}
}

func TestBytesBufferReadByte(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("ABC")

	b, err := bb.ReadByte()
	if err != nil {
		t.Errorf("ReadByte error: %v", err)
	}
	if b != 65 {
		t.Errorf("expected 65 ('A'), got %d", b)
	}
}

func TestBytesBufferReadBytes(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello,world")

	data, err := bb.ReadBytes(',')
	if err != nil {
		t.Errorf("ReadBytes error: %v", err)
	}
	if string(data) != "hello," {
		t.Errorf("expected 'hello,', got '%s'", string(data))
	}
}

func TestBytesBufferReadString(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello,world")

	s, err := bb.ReadString(',')
	if err != nil {
		t.Errorf("ReadString error: %v", err)
	}
	if s != "hello," {
		t.Errorf("expected 'hello,', got '%s'", s)
	}
}

func TestBytesBufferSeek(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	// Seek to beginning
	pos := bb.Seek(0, 0)
	if pos != 0 {
		t.Errorf("expected position 0, got %d", pos)
	}
}

func TestBytesBufferEquals(t *testing.T) {
	bb1 := NewBytesBuffer()
	bb1.WriteString("hello")

	bb2 := NewBytesBuffer()
	bb2.WriteString("hello")

	bb3 := NewBytesBuffer()
	bb3.WriteString("world")

	if !bb1.Equals(bb2) {
		t.Error("expected equal buffers to be equal")
	}
	if bb1.Equals(bb3) {
		t.Error("expected different buffers to not be equal")
	}
}

func TestBytesBufferGetIOReader(t *testing.T) {
	bb := NewBytesBuffer()
	bb.WriteString("hello")

	reader := bb.GetIOReader()
	if reader == nil {
		t.Error("expected non-nil reader")
	}
}

func TestBytesBufferGetIOWriter(t *testing.T) {
	bb := NewBytesBuffer()

	writer := bb.GetIOWriter()
	if writer == nil {
		t.Error("expected non-nil writer")
	}
}
