// pkg/objects/bytes_test.go
package objects

import (
	"testing"
)

func TestBytesBasic(t *testing.T) {
	// Test creation from byte slice
	b := NewBytes([]byte{72, 101, 108, 108, 111})
	if b.Len() != 5 {
		t.Errorf("Expected len 5, got %d", b.Len())
	}
	if b.String() != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", b.String())
	}
}

func TestBytesFromArray(t *testing.T) {
	elements := []Object{
		NewInt(72), NewInt(101), NewInt(108), NewInt(108), NewInt(111),
	}
	b := NewBytesFromArray(elements)
	if b == nil {
		t.Fatal("NewBytesFromArray returned nil")
	}
	if b.Len() != 5 {
		t.Errorf("Expected len 5, got %d", b.Len())
	}

	// Test with invalid element
	invalidElements := []Object{NewInt(72), NewString("invalid")}
	b2 := NewBytesFromArray(invalidElements)
	if b2 != nil {
		t.Error("Expected nil for invalid elements")
	}

	// Test with out of range value
	outOfRange := []Object{NewInt(256)}
	b3 := NewBytesFromArray(outOfRange)
	if b3 != nil {
		t.Error("Expected nil for out of range value")
	}
}

func TestBytesAt(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	val, ok := b.At(0)
	if !ok || val != 72 {
		t.Errorf("Expected 72, got %d, ok=%v", val, ok)
	}

	_, ok = b.At(10)
	if ok {
		t.Error("Expected false for out of range index")
	}
}

func TestBytesSlice(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	sliced := b.Slice(0, 5)
	if sliced.Len() != 5 {
		t.Errorf("Expected len 5, got %d", sliced.Len())
	}

	sliced2 := b.Slice(1, 4)
	if sliced2.Len() != 3 {
		t.Errorf("Expected len 3, got %d", sliced2.Len())
	}
	if sliced2.String() != "ell" {
		t.Errorf("Expected 'ell', got '%s'", sliced2.String())
	}
}

func TestBytesToArray(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108})
	arr := b.ToArray()
	if len(arr.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr.Elements))
	}

	for i, expected := range []int64{72, 101, 108} {
		val := arr.Elements[i].(*Int)
		if val.Value != expected {
			t.Errorf("Expected %d at index %d, got %d", expected, i, val.Value)
		}
	}
}

func TestBytesContains(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	sub := NewBytes([]byte{108, 108})
	if !b.Contains(sub) {
		t.Error("Expected to contain 'll'")
	}

	notSub := NewBytes([]byte{119, 111})
	if b.Contains(notSub) {
		t.Error("Expected not to contain 'wo'")
	}
}

func TestBytesHasPrefixSuffix(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	prefix := NewBytes([]byte{72, 101})
	if !b.HasPrefix(prefix) {
		t.Error("Expected to have prefix 'He'")
	}

	suffix := NewBytes([]byte{108, 111})
	if !b.HasSuffix(suffix) {
		t.Error("Expected to have suffix 'lo'")
	}
}

func TestBytesRepeat(t *testing.T) {
	b := NewBytes([]byte{65, 66})
	repeated := b.Repeat(3)
	if repeated.Len() != 6 {
		t.Errorf("Expected len 6, got %d", repeated.Len())
	}
	if repeated.String() != "ABABAB" {
		t.Errorf("Expected 'ABABAB', got '%s'", repeated.String())
	}
}

func TestBytesConcat(t *testing.T) {
	b1 := NewBytes([]byte{72, 101})
	b2 := NewBytes([]byte{108, 108, 111})

	concatenated := b1.Concat(b2)
	if concatenated.Len() != 5 {
		t.Errorf("Expected len 5, got %d", concatenated.Len())
	}
	if concatenated.String() != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", concatenated.String())
	}
}

func TestBytesGetIOReader(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})
	reader := b.GetIOReader()
	if reader == nil {
		t.Error("Expected non-nil reader")
	}

	// Read all from reader
	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}
}

func TestBytesEmpty(t *testing.T) {
	empty := BYTES_EMPTY
	if empty.Len() != 0 {
		t.Errorf("Expected empty bytes to have len 0, got %d", empty.Len())
	}
	if empty.ToBool() != FALSE {
		t.Error("Expected empty bytes to be falsy")
	}

	nonEmpty := NewBytes([]byte{65})
	if nonEmpty.ToBool() != TRUE {
		t.Error("Expected non-empty bytes to be truthy")
	}
}

func TestBytesType(t *testing.T) {
	b := NewBytes([]byte{65})
	if b.Type() != BytesType {
		t.Errorf("Expected BytesType, got %v", b.Type())
	}
}

func TestBytesTypeTag(t *testing.T) {
	b := NewBytes([]byte{65})
	if b.TypeTag() != TagBytes {
		t.Errorf("Expected TagBytes, got %v", b.TypeTag())
	}
}

func TestBytesInspect(t *testing.T) {
	b := NewBytes([]byte{65, 66, 67})
	result := b.Inspect()
	if result == "" {
		t.Error("Expected non-empty inspect string")
	}
}

func TestBytesHashKey(t *testing.T) {
	b1 := NewBytes([]byte{65, 66, 67})
	b2 := NewBytes([]byte{65, 66, 67})
	b3 := NewBytes([]byte{68, 69, 70})

	h1 := b1.HashKey()
	h2 := b2.HashKey()
	h3 := b3.HashKey()

	if h1 != h2 {
		t.Error("Expected same hash for same bytes")
	}
	if h1 == h3 {
		t.Error("Expected different hash for different bytes")
	}
}

func TestBytesEqual(t *testing.T) {
	b1 := NewBytes([]byte{65, 66, 67})
	b2 := NewBytes([]byte{65, 66, 67})
	b3 := NewBytes([]byte{68, 69, 70})

	if !b1.Equal(b2) {
		t.Error("Expected equal bytes to be equal")
	}
	if b1.Equal(b3) {
		t.Error("Expected different bytes to not be equal")
	}
}

func TestBytesIndex(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	sub := NewBytes([]byte{108})
	idx := b.Index(sub)
	if idx != 2 {
		t.Errorf("Expected index 2, got %d", idx)
	}

	notSub := NewBytes([]byte{119})
	idx = b.Index(notSub)
	if idx != -1 {
		t.Errorf("Expected -1 for not found, got %d", idx)
	}
}

func TestBytesLastIndex(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	sub := NewBytes([]byte{108})
	idx := b.LastIndex(sub)
	if idx != 3 {
		t.Errorf("Expected last index 3, got %d", idx)
	}
}

func TestBytesCount(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	sub := NewBytes([]byte{108})
	count := b.Count(sub)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestBytesGetMember(t *testing.T) {
	b := NewBytes([]byte{72, 101, 108, 108, 111})

	result := b.GetMember("len")
	if result == nil {
		t.Error("Expected non-nil result for 'len'")
	}

	result = b.GetMember("toString")
	if result == nil {
		t.Error("Expected non-nil result for 'toString'")
	}

	result = b.GetMember("at")
	if result == nil {
		t.Error("Expected non-nil result for 'at'")
	}
}
