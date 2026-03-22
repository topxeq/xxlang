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
