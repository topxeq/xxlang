// pkg/objects/chars_test.go
package objects

import (
	"testing"
)

func TestCharsBasic(t *testing.T) {
	// Test empty chars
	empty := NewChars([]rune{})
	if empty.Type() != CharsType {
		t.Errorf("expected CharsType, got %s", empty.Type())
	}
	if empty.Len() != 0 {
		t.Errorf("expected length 0, got %d", empty.Len())
	}

	// Test chars from string
	ch := NewCharsFromString("hello")
	if ch.Type() != CharsType {
		t.Errorf("expected CharsType, got %s", ch.Type())
	}
	if ch.Len() != 5 {
		t.Errorf("expected length 5, got %d", ch.Len())
	}
	if ch.Inspect() != "hello" {
		t.Errorf("expected 'hello', got '%s'", ch.Inspect())
	}
}

func TestCharsUnicode(t *testing.T) {
	// Test Unicode handling
	ch := NewCharsFromString("中文测试")
	if ch.Len() != 4 {
		t.Errorf("expected length 4 (4 Chinese characters), got %d", ch.Len())
	}

	// Test At method
	s, ok := ch.At(0)
	if !ok {
		t.Error("expected ok to be true")
	}
	if s != "中" {
		t.Errorf("expected '中', got '%s'", s)
	}

	// Test Slice method
	slice := ch.Slice(1, 3)
	if slice.Len() != 2 {
		t.Errorf("expected length 2, got %d", slice.Len())
	}
	if slice.Inspect() != "文测" {
		t.Errorf("expected '文测', got '%s'", slice.Inspect())
	}
}

func TestCharsToBool(t *testing.T) {
	// Empty chars should be falsy
	empty := NewChars([]rune{})
	if empty.ToBool() == TRUE {
		t.Error("empty chars should be falsy")
	}

	// Non-empty chars should be truthy
	ch := NewCharsFromString("test")
	if ch.ToBool() == FALSE {
		t.Error("non-empty chars should be truthy")
	}
}

func TestCharsString(t *testing.T) {
	ch := NewCharsFromString("test")
	if ch.String() != "test" {
		t.Errorf("expected 'test', got '%s'", ch.String())
	}
}

func TestCharsTypeTag(t *testing.T) {
	ch := NewCharsFromString("test")
	if ch.TypeTag() != TagChars {
		t.Errorf("expected TagChars, got %d", ch.TypeTag())
	}
}

func TestCharsHashKey(t *testing.T) {
	ch1 := NewCharsFromString("hello")
	ch2 := NewCharsFromString("hello")
	ch3 := NewCharsFromString("world")

	h1 := ch1.HashKey()
	h2 := ch2.HashKey()
	h3 := ch3.HashKey()

	if h1 != h2 {
		t.Error("expected same hash for same chars")
	}

	if h1 == h3 {
		t.Error("expected different hash for different chars")
	}

	empty := NewChars([]rune{})
	emptyHash := empty.HashKey()
	if emptyHash.Type != CharsType {
		t.Error("expected CharsType for empty chars hash")
	}
}

func TestCharsAt(t *testing.T) {
	ch := NewCharsFromString("hello")

	s, ok := ch.At(0)
	if !ok || s != "h" {
		t.Errorf("expected 'h', got '%s', ok=%v", s, ok)
	}

	s, ok = ch.At(4)
	if !ok || s != "o" {
		t.Errorf("expected 'o', got '%s', ok=%v", s, ok)
	}

	_, ok = ch.At(-1)
	if ok {
		t.Error("expected false for negative index")
	}

	_, ok = ch.At(10)
	if ok {
		t.Error("expected false for out of bounds index")
	}
}

func TestCharsSlice(t *testing.T) {
	ch := NewCharsFromString("hello world")

	slice := ch.Slice(0, 5)
	if slice.Inspect() != "hello" {
		t.Errorf("expected 'hello', got '%s'", slice.Inspect())
	}

	slice = ch.Slice(6, 11)
	if slice.Inspect() != "world" {
		t.Errorf("expected 'world', got '%s'", slice.Inspect())
	}

	slice = ch.Slice(0, 100)
	if slice.Inspect() != "hello world" {
		t.Errorf("expected 'hello world' for clamped slice, got '%s'", slice.Inspect())
	}

	slice = ch.Slice(-5, 5)
	if slice.Inspect() != "hello" {
		t.Errorf("expected 'hello' for negative start, got '%s'", slice.Inspect())
	}
}
