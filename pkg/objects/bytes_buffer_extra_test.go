// pkg/objects/bytes_buffer_extra_test.go
// Additional tests for BytesBuffer operations.
package objects

import (
	"testing"
)

func TestBytesBuffer_WriteRead(t *testing.T) {
	buf := NewBytesBuffer()

	n := buf.Write([]byte("hello"))
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got %s", buf.String())
	}

	p := make([]byte, 5)
	n, err := buf.Read(p)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes read, got %d", n)
	}
	if string(p) != "hello" {
		t.Errorf("expected 'hello', got %s", string(p))
	}
}

func TestBytesBuffer_WriteString(t *testing.T) {
	buf := NewBytesBuffer()
	n := buf.WriteString("world")
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
	if buf.String() != "world" {
		t.Errorf("expected 'world', got %s", buf.String())
	}
}

func TestBytesBuffer_WriteByte(t *testing.T) {
	buf := NewBytesBuffer()
	if err := buf.WriteByte('A'); err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}
	if buf.String() != "A" {
		t.Errorf("expected 'A', got %s", buf.String())
	}
}

func TestBytesBuffer_IntOps(t *testing.T) {
	buf := NewBytesBuffer()

	buf.WriteInt16(1000)
	buf.WriteInt32(100000)
	buf.WriteInt64(1000000000)

	buf.Seek(0, 0)

	v16, err := buf.ReadInt16()
	if err != nil || v16 != 1000 {
		t.Errorf("expected 1000, got %d, err=%v", v16, err)
	}

	v32, err := buf.ReadInt32()
	if err != nil || v32 != 100000 {
		t.Errorf("expected 100000, got %d, err=%v", v32, err)
	}

	v64, err := buf.ReadInt64()
	if err != nil || v64 != 1000000000 {
		t.Errorf("expected 1000000000, got %d, err=%v", v64, err)
	}
}

func TestBytesBuffer_FloatOps(t *testing.T) {
	buf := NewBytesBuffer()

	buf.WriteFloat32(3.14)
	buf.WriteFloat64(2.718281828)

	buf.Seek(0, 0)

	v32, err := buf.ReadFloat32()
	if err != nil {
		t.Errorf("ReadFloat32 failed: %v", err)
	}
	if v32 < 3.13 || v32 > 3.15 {
		t.Errorf("expected ~3.14, got %f", v32)
	}

	v64, err := buf.ReadFloat64()
	if err != nil {
		t.Errorf("ReadFloat64 failed: %v", err)
	}
	if v64 < 2.71828 || v64 > 2.71829 {
		t.Errorf("expected ~2.71828, got %f", v64)
	}
}

func TestBytesBuffer_WriteTo(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("output data"))

	other := NewBytesBuffer()
	n, err := buf.WriteTo(other)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != 11 {
		t.Errorf("expected 11 bytes, got %d", n)
	}
	if other.String() != "output data" {
		t.Errorf("expected 'output data', got %s", other.String())
	}
}

func TestBytesBuffer_Seek(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello world"))

	offset := buf.Seek(6, 0)
	if offset != 6 {
		t.Errorf("expected offset 6, got %d", offset)
	}
	// Seek position is tracked, verify it returns correct offset
	_ = offset
}

func TestBytesBuffer_Truncate(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello world"))

	buf.Truncate(5)
	if buf.Len() != 5 {
		t.Errorf("expected length 5, got %d", buf.Len())
	}
	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got %s", buf.String())
	}
}

func TestBytesBuffer_Reset(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello"))
	buf.Reset()
	if buf.Len() != 0 {
		t.Errorf("expected length 0 after reset, got %d", buf.Len())
	}
}

func TestBytesBuffer_Clear(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello"))
	buf.Clear()
	if buf.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", buf.Len())
	}
}

func TestBytesBuffer_Bytes(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("test"))

	b := buf.Bytes()
	if string(b) != "test" {
		t.Errorf("expected 'test', got %s", string(b))
	}
}

func TestBytesBuffer_Cap(t *testing.T) {
	buf := NewBytesBuffer()
	if buf.Cap() < 0 {
		t.Error("Cap should not be negative")
	}
	buf.Grow(100)
	if buf.Cap() < 100 {
		t.Errorf("expected cap >= 100, got %d", buf.Cap())
	}
}

func TestBytesBuffer_Peek(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello"))

	p := buf.Peek(3)
	if string(p) != "hel" {
		t.Errorf("expected 'hel', got %s", string(p))
	}
}

func TestBytesBuffer_ReadByte(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("AB"))

	b, err := buf.ReadByte()
	if err != nil || b != 'A' {
		t.Errorf("expected 'A', got %c, err=%v", b, err)
	}
}

func TestBytesBuffer_ReadBytes(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello,world"))

	data, err := buf.ReadBytes(',')
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if string(data) != "hello," {
		t.Errorf("expected 'hello,', got %s", string(data))
	}
}

func TestBytesBuffer_ReadString(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("hello\nworld"))

	s, err := buf.ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if s != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", s)
	}
}

func TestBytesBuffer_Equals(t *testing.T) {
	buf1 := NewBytesBuffer()
	buf1.Write([]byte("test"))

	buf2 := NewBytesBuffer()
	buf2.Write([]byte("test"))

	if !buf1.Equals(buf2) {
		t.Error("identical buffers should be equal")
	}

	buf3 := NewBytesBuffer()
	buf3.Write([]byte("different"))
	if buf1.Equals(buf3) {
		t.Error("different buffers should not be equal")
	}
}

func TestBytesBuffer_NewFromBytes(t *testing.T) {
	buf := NewBytesBufferFromBytes([]byte("initial"))
	if buf.String() != "initial" {
		t.Errorf("expected 'initial', got %s", buf.String())
	}
}

func TestBytesBuffer_GetIOReader(t *testing.T) {
	buf := NewBytesBuffer()
	buf.Write([]byte("test"))

	reader := buf.GetIOReader()
	p := make([]byte, 4)
	n, err := reader.Read(p)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 4 || string(p) != "test" {
		t.Errorf("expected 'test', got %s", string(p))
	}
}

func TestBytesBuffer_GetIOWriter(t *testing.T) {
	buf := NewBytesBuffer()
	writer := buf.GetIOWriter()
	n, err := writer.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
	if buf.String() != "test" {
		t.Errorf("expected 'test', got %s", buf.String())
	}
}
