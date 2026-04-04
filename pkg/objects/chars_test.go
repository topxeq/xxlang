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
	// Inspect returns representation format
	if ch.Inspect() != "Chars(len=5)" {
		t.Errorf("expected 'Chars(len=5)', got '%s'", ch.Inspect())
	}
	// String() returns the actual content
	if ch.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", ch.String())
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
	if slice.String() != "文测" {
		t.Errorf("expected '文测', got '%s'", slice.String())
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
	if slice.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", slice.String())
	}

	slice = ch.Slice(6, 11)
	if slice.String() != "world" {
		t.Errorf("expected 'world', got '%s'", slice.String())
	}

	slice = ch.Slice(0, 100)
	if slice.String() != "hello world" {
		t.Errorf("expected 'hello world' for clamped slice, got '%s'", slice.String())
	}

	slice = ch.Slice(-5, 5)
	if slice.String() != "hello" {
		t.Errorf("expected 'hello' for negative start, got '%s'", slice.String())
	}
}

// TestBuiltinCharsFromString tests the chars() built-in with string argument
func TestBuiltinCharsFromString(t *testing.T) {
	// Get the chars built-in
	builtin, ok := Builtins["chars"]
	if !ok {
		t.Fatal("chars built-in not found")
	}

	// Test converting string to chars
	str := NewString("hello")
	result := builtin.Fn(str)
	charsObj, ok := result.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result)
	}
	if charsObj.String() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", charsObj.String())
	}
	if charsObj.Len() != 5 {
		t.Errorf("Expected len 5, got %d", charsObj.Len())
	}

	// Test with string containing Unicode characters
	str2 := NewString("中文测试")
	result2 := builtin.Fn(str2)
	charsObj2, ok := result2.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result2)
	}
	if charsObj2.Len() != 4 {
		t.Errorf("Expected len 4 for '中文测试', got %d", charsObj2.Len())
	}

	// Test empty string
	emptyStr := NewString("")
	result3 := builtin.Fn(emptyStr)
	emptyChars, ok := result3.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result3)
	}
	if emptyChars.Len() != 0 {
		t.Errorf("Expected len 0 for empty string, got %d", emptyChars.Len())
	}
}

// TestBuiltinCharsFromCodePoints tests the chars() built-in with Unicode code points
func TestBuiltinCharsFromCodePoints(t *testing.T) {
	builtin, ok := Builtins["chars"]
	if !ok {
		t.Fatal("chars built-in not found")
	}

	// Test creating chars from Unicode code points
	args := []Object{NewInt(65), NewInt(66), NewInt(67), NewInt(20013)}
	result := builtin.Fn(args...)
	charsObj, ok := result.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result)
	}
	if charsObj.String() != "ABC中" {
		t.Errorf("Expected 'ABC中', got '%s'", charsObj.String())
	}
	if charsObj.Len() != 4 {
		t.Errorf("Expected len 4, got %d", charsObj.Len())
	}

	// Test with single code point
	singleArg := []Object{NewInt(65)}
	result2 := builtin.Fn(singleArg...)
	charsObj2, ok := result2.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result2)
	}
	if charsObj2.String() != "A" {
		t.Errorf("Expected 'A', got '%s'", charsObj2.String())
	}
}

// TestBuiltinCharsFromChars tests the chars() built-in with Chars argument (identity)
func TestBuiltinCharsFromChars(t *testing.T) {
	builtin, ok := Builtins["chars"]
	if !ok {
		t.Fatal("chars built-in not found")
	}

	// Test passing Chars to chars() returns the same object
	original := NewCharsFromString("test")
	result := builtin.Fn(original)
	charsObj, ok := result.(*Chars)
	if !ok {
		t.Fatalf("Expected Chars, got %T", result)
	}
	if charsObj.String() != "test" {
		t.Errorf("Expected 'test', got '%s'", charsObj.String())
	}
}

// TestCharsToStringMethod tests the toString method on Chars
func TestCharsToStringMethod(t *testing.T) {
	// Test via GetMember
	ch := NewCharsFromString("hello")
	method := ch.GetMember("toString")
	if method == NULL {
		t.Error("Expected toString method to exist")
	}

	// Test via TypeMethods
	builtin, ok := GetMethod(CharsType, "toString")
	if !ok {
		t.Fatal("toString method not found in TypeMethods")
	}
	result := builtin.Fn(ch)
	str, ok := result.(*String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if str.Value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", str.Value)
	}
}
