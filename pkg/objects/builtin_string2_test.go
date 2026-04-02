// pkg/objects/builtin_string2_test.go
package objects

import (
	"testing"
)

func TestBuiltinStrSplitLines(t *testing.T) {
	fn, ok := Builtins["strSplitLines"]
	if !ok {
		t.Fatal("strSplitLines builtin not found")
	}

	result := fn.Fn(NewString("hello\nworld"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("line1\nline2\nline3"))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("hello\r\nworld"))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 elements for CRLF, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrContainsAny(t *testing.T) {
	fn, ok := Builtins["strContainsAny"]
	if !ok {
		t.Fatal("strContainsAny builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewString("aeiou"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected TRUE for vowels present")
	}

	result = fn.Fn(NewString("xyz"), NewString("aeiou"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected FALSE for no vowels")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrIndex(t *testing.T) {
	fn, ok := Builtins["strIndex"]
	if !ok {
		t.Fatal("strIndex builtin not found")
	}

	result := fn.Fn(NewString("hello world"), NewString("world"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 6 {
		t.Errorf("expected 6, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("xyz"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -1 {
		t.Errorf("expected -1 for not found, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrLastIndex(t *testing.T) {
	fn, ok := Builtins["strLastIndex"]
	if !ok {
		t.Fatal("strLastIndex builtin not found")
	}

	result := fn.Fn(NewString("hello world world"), NewString("world"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 12 {
		t.Errorf("expected 12 (last occurrence), got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("l"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("expected 3, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrSplitN(t *testing.T) {
	fn, ok := Builtins["strSplitN"]
	if !ok {
		t.Fatal("strSplitN builtin not found")
	}

	result := fn.Fn(NewString("a,b,c,d"), NewString(","), NewInt(3))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrPad(t *testing.T) {
	fn, ok := Builtins["strPad"]
	if !ok {
		t.Fatal("strPad builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewInt(10))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) != 10 {
		t.Errorf("expected length 10, got %d", len(strResult.Value))
	}

	result = fn.Fn(NewString("hello"), NewInt(10), NewString("*"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "*****hello" {
		t.Errorf("expected '*****hello', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrSub(t *testing.T) {
	fn, ok := Builtins["strSub"]
	if !ok {
		t.Fatal("strSub builtin not found")
	}

	result := fn.Fn(NewString("hello world"), NewInt(0), NewInt(5))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello world"), NewInt(6))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "world" {
		t.Errorf("expected 'world', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewInt(0), NewInt(10))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello' for clamped end index, got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIntToStr(t *testing.T) {
	fn, ok := Builtins["intToStr"]
	if !ok {
		t.Fatal("intToStr builtin not found")
	}

	result := fn.Fn(NewInt(123))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "123" {
		t.Errorf("expected '123', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewInt(-456))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "-456" {
		t.Errorf("expected '-456', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinFloatToStr(t *testing.T) {
	fn, ok := Builtins["floatToStr"]
	if !ok {
		t.Fatal("floatToStr builtin not found")
	}

	result := fn.Fn(NewFloat(123.456))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCharCode(t *testing.T) {
	fn, ok := Builtins["charCode"]
	if !ok {
		t.Fatal("charCode builtin not found")
	}

	result := fn.Fn(NewString("A"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 65 {
		t.Errorf("expected 65 (ASCII 'A'), got %d", intResult.Value)
	}

	result = fn.Fn(NewString("a"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 97 {
		t.Errorf("expected 97 (ASCII 'a'), got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCharFromCode(t *testing.T) {
	fn, ok := Builtins["charFromCode"]
	if !ok {
		t.Fatal("charFromCode builtin not found")
	}

	result := fn.Fn(NewInt(65))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "A" {
		t.Errorf("expected 'A', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewInt(97))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "a" {
		t.Errorf("expected 'a', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitNot(t *testing.T) {
	fn, ok := Builtins["bitNot"]
	if !ok {
		t.Fatal("bitNot builtin not found")
	}

	result := fn.Fn(NewInt(0))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -1 {
		t.Errorf("expected -1, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(1))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -2 {
		t.Errorf("expected -2, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitAnd(t *testing.T) {
	fn, ok := Builtins["bitAnd"]
	if !ok {
		t.Fatal("bitAnd builtin not found")
	}

	result := fn.Fn(NewInt(15), NewInt(7))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 7 {
		t.Errorf("expected 7, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(10), NewInt(6))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 2 {
		t.Errorf("expected 2, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitOr(t *testing.T) {
	fn, ok := Builtins["bitOr"]
	if !ok {
		t.Fatal("bitOr builtin not found")
	}

	result := fn.Fn(NewInt(10), NewInt(6))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 14 {
		t.Errorf("expected 14, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(15), NewInt(7))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 15 {
		t.Errorf("expected 15, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitXor(t *testing.T) {
	fn, ok := Builtins["bitXor"]
	if !ok {
		t.Fatal("bitXor builtin not found")
	}

	result := fn.Fn(NewInt(10), NewInt(6))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 12 {
		t.Errorf("expected 12, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(15), NewInt(15))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 0 {
		t.Errorf("expected 0, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitShiftLeft(t *testing.T) {
	fn, ok := Builtins["bitShiftLeft"]
	if !ok {
		t.Fatal("bitShiftLeft builtin not found")
	}

	result := fn.Fn(NewInt(1), NewInt(3))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 8 {
		t.Errorf("expected 8, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(5), NewInt(2))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 20 {
		t.Errorf("expected 20, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBitShiftRight(t *testing.T) {
	fn, ok := Builtins["bitShiftRight"]
	if !ok {
		t.Fatal("bitShiftRight builtin not found")
	}

	result := fn.Fn(NewInt(8), NewInt(3))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("expected 1, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(20), NewInt(2))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinReverseMap(t *testing.T) {
	fn, ok := Builtins["reverseMap"]
	if !ok {
		t.Fatal("reverseMap builtin not found")
	}

	m := NewMapWithCapacity(2)
	key1 := NewString("a")
	val1 := NewString("1")
	key2 := NewString("b")
	val2 := NewString("2")
	m.Pairs[key1.HashKey()] = MapPair{Key: key1, Value: val1}
	m.Pairs[key2.HashKey()] = MapPair{Key: key2, Value: val2}

	result := fn.Fn(m)
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(mapResult.Pairs))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinSimpleStrToMap(t *testing.T) {
	fn, ok := Builtins["simpleStrToMap"]
	if !ok {
		t.Fatal("simpleStrToMap builtin not found")
	}

	result := fn.Fn(NewString("a=1,b=2"))
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(mapResult.Pairs))
	}

	result = fn.Fn(NewString("a:1;b:2"), NewString(";"), NewString(":"))
	_, ok = result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	result = fn.Fn(NewString(""))
	mapResult, ok = result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 0 {
		t.Errorf("expected 0 pairs for empty string, got %d", len(mapResult.Pairs))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinMapToStr(t *testing.T) {
	fn, ok := Builtins["mapToStr"]
	if !ok {
		t.Fatal("mapToStr builtin not found")
	}

	m := NewMapWithCapacity(2)
	key1 := NewString("a")
	val1 := NewString("1")
	key2 := NewString("b")
	val2 := NewString("2")
	m.Pairs[key1.HashKey()] = MapPair{Key: key1, Value: val1}
	m.Pairs[key2.HashKey()] = MapPair{Key: key2, Value: val2}

	result := fn.Fn(m)
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty string")
	}

	result = fn.Fn(m, NewString(";"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(m, NewString(";"), NewString(":"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinStrPadWithPadRight(t *testing.T) {
	fn, ok := Builtins["strPad"]
	if !ok {
		t.Fatal("strPad builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewInt(10), NewString("*"), TRUE)
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello*****" {
		t.Errorf("expected 'hello*****', got '%s'", strResult.Value)
	}
}

func TestBuiltinStrSubWithNegativeIndex(t *testing.T) {
	fn, ok := Builtins["strSub"]
	if !ok {
		t.Fatal("strSub builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewInt(-2))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "lo" {
		t.Errorf("expected 'lo', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewInt(0), NewInt(-1))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hell" {
		t.Errorf("expected 'hell', got '%s'", strResult.Value)
	}
}

func TestBuiltinIntToStrWithBase(t *testing.T) {
	fn, ok := Builtins["intToStr"]
	if !ok {
		t.Fatal("intToStr builtin not found")
	}

	result := fn.Fn(NewInt(255), NewInt(16))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "ff" {
		t.Errorf("expected 'ff', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewInt(8), NewInt(2))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "1000" {
		t.Errorf("expected '1000', got '%s'", strResult.Value)
	}
}

func TestBuiltinFloatToStrWithPrecision(t *testing.T) {
	fn, ok := Builtins["floatToStr"]
	if !ok {
		t.Fatal("floatToStr builtin not found")
	}

	result := fn.Fn(NewFloat(3.14159), NewInt(2))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "3.14" {
		t.Errorf("expected '3.14', got '%s'", strResult.Value)
	}
}

func TestBuiltinCharCodeWithIndex(t *testing.T) {
	fn, ok := Builtins["charCode"]
	if !ok {
		t.Fatal("charCode builtin not found")
	}

	result := fn.Fn(NewString("ABC"), NewInt(1))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 66 {
		t.Errorf("expected 66 (ASCII 'B'), got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewInt(100))
	if !isError(result) {
		t.Error("expected error for index out of bounds")
	}
}

func TestBuiltinStrSplitLinesEmpty(t *testing.T) {
	fn, ok := Builtins["strSplitLines"]
	if !ok {
		t.Fatal("strSplitLines builtin not found")
	}

	result := fn.Fn(NewString(""))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 1 {
		t.Errorf("expected 1 element for empty string, got %d", len(arrResult.Elements))
	}
}

func TestBuiltinStrContainsAnyError(t *testing.T) {
	fn, ok := Builtins["strContainsAny"]
	if !ok {
		t.Fatal("strContainsAny builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	if !isError(result) {
		t.Error("expected error for 1 arg")
	}

	result = fn.Fn(NewString("hello"), NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinStrIndexError(t *testing.T) {
	fn, ok := Builtins["strIndex"]
	if !ok {
		t.Fatal("strIndex builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	if !isError(result) {
		t.Error("expected error for 1 arg")
	}

	result = fn.Fn(NewString("hello"), NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}
