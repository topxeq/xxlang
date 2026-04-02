// pkg/objects/builtin_regex_test.go
package objects

import (
	"testing"
)

func TestBuiltinRegMatch(t *testing.T) {
	fn, ok := Builtins["regMatch"]
	if !ok {
		t.Fatal("regMatch builtin not found")
	}

	result := fn.Fn(NewString("hello123"), NewString("^[a-z]+[0-9]+$"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for matching pattern")
	}

	result = fn.Fn(NewString("hello"), NewString("^[0-9]+$"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for non-matching pattern")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1), NewString("pattern"))
	if !isError(result) {
		t.Error("expected error for non-string first arg")
	}
}

func TestBuiltinRegContains(t *testing.T) {
	fn, ok := Builtins["regContains"]
	if !ok {
		t.Fatal("regContains builtin not found")
	}

	result := fn.Fn(NewString("hello123world"), NewString("[0-9]+"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for containing pattern")
	}

	result = fn.Fn(NewString("hello"), NewString("[0-9]+"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for not containing pattern")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindFirst(t *testing.T) {
	fn, ok := Builtins["regFindFirst"]
	if !ok {
		t.Fatal("regFindFirst builtin not found")
	}

	result := fn.Fn(NewString("hello123world456"), NewString("[0-9]+"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "123" {
		t.Errorf("expected '123', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("[0-9]+"))
	if result != NULL {
		t.Errorf("expected NULL for no match, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindAll(t *testing.T) {
	fn, ok := Builtins["regFindAll"]
	if !ok {
		t.Fatal("regFindAll builtin not found")
	}

	result := fn.Fn(NewString("hello123world456"), NewString("[0-9]+"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 matches, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegReplace(t *testing.T) {
	fn, ok := Builtins["regReplace"]
	if !ok {
		t.Fatal("regReplace builtin not found")
	}

	result := fn.Fn(NewString("hello123world456"), NewString("[0-9]+"), NewString("X"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "helloXworldX" {
		t.Errorf("expected 'helloXworldX', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegSplit(t *testing.T) {
	fn, ok := Builtins["regSplit"]
	if !ok {
		t.Fatal("regSplit builtin not found")
	}

	result := fn.Fn(NewString("a1b2c3d"), NewString("[0-9]"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 4 {
		t.Errorf("expected 4 parts, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegCount(t *testing.T) {
	fn, ok := Builtins["regCount"]
	if !ok {
		t.Fatal("regCount builtin not found")
	}

	result := fn.Fn(NewString("a1b2c3d"), NewString("[0-9]"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("expected 3 matches, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegQuote(t *testing.T) {
	fn, ok := Builtins["regQuote"]
	if !ok {
		t.Fatal("regQuote builtin not found")
	}

	result := fn.Fn(NewString("a.b*c+d"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty quoted string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindFirstGroups(t *testing.T) {
	fn, ok := Builtins["regFindFirstGroups"]
	if !ok {
		t.Fatal("regFindFirstGroups builtin not found")
	}

	result := fn.Fn(NewString("hello123world"), NewString("([a-z]+)([0-9]+)"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 groups (full match + 2), got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("hello"), NewString("([0-9]+)"))
	if result != NULL {
		t.Errorf("expected NULL for no match, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindAllGroups(t *testing.T) {
	fn, ok := Builtins["regFindAllGroups"]
	if !ok {
		t.Fatal("regFindAllGroups builtin not found")
	}

	result := fn.Fn(NewString("hello123world456"), NewString("([a-z]+)([0-9]+)"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 match groups, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindAllIndex(t *testing.T) {
	fn, ok := Builtins["regFindAllIndex"]
	if !ok {
		t.Fatal("regFindAllIndex builtin not found")
	}

	result := fn.Fn(NewString("hello123world456"), NewString("[0-9]+"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 index pairs, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRegFindFirstWithGroupIndex(t *testing.T) {
	fn, ok := Builtins["regFindFirst"]
	if !ok {
		t.Fatal("regFindFirst builtin not found")
	}

	result := fn.Fn(NewString("hello123world"), NewString("([a-z]+)([0-9]+)"), NewInt(1))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello123world"), NewString("([a-z]+)([0-9]+)"), NewInt(2))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "123" {
		t.Errorf("expected '123', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello123"), NewString("([a-z]+)([0-9]+)"), NewInt(5))
	if !isError(result) {
		t.Error("expected error for out of range group index")
	}
}

func TestBuiltinRegFindAllWithGroupIndex(t *testing.T) {
	fn, ok := Builtins["regFindAll"]
	if !ok {
		t.Fatal("regFindAll builtin not found")
	}

	result := fn.Fn(NewString("a1b2a3b4"), NewString("([ab])([0-9])"), NewInt(1))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 4 {
		t.Errorf("expected 4 matches (2 chars + 2 nums), got %d", len(arrResult.Elements))
	}
}

func TestBuiltinRegSplitWithLimit(t *testing.T) {
	fn, ok := Builtins["regSplit"]
	if !ok {
		t.Fatal("regSplit builtin not found")
	}

	result := fn.Fn(NewString("a1b2c3d"), NewString("[0-9]"), NewInt(2))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 parts with limit, got %d", len(arrResult.Elements))
	}
}
