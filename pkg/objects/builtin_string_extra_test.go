// pkg/objects/builtin_string_extra_test.go
package objects

import (
	"testing"
)

func TestBuiltinStrContainsIn(t *testing.T) {
	fn, ok := Builtins["strContainsIn"]
	if !ok {
		t.Fatal("strContainsIn builtin not found")
	}

	result := fn.Fn(NewString("hello world"), NewString("world"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for containing substring")
	}

	result = fn.Fn(NewString("hello world"), NewString("foo"), NewString("world"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for containing one of substrings")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrRuneLen(t *testing.T) {
	fn, ok := Builtins["strRuneLen"]
	if !ok {
		t.Fatal("strRuneLen builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("你好世界"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 4 {
		t.Errorf("expected 4, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrIn(t *testing.T) {
	fn, ok := Builtins["strIn"]
	if !ok {
		t.Fatal("strIn builtin not found")
	}

	arr := &Array{Elements: []Object{NewString("a"), NewString("b"), NewString("c")}}
	result := fn.Fn(NewString("b"), arr)
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for string in array")
	}

	result = fn.Fn(NewString("d"), arr)
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for string not in array")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinLimitStr(t *testing.T) {
	fn, ok := Builtins["limitStr"]
	if !ok {
		t.Fatal("limitStr builtin not found")
	}

	result := fn.Fn(NewString("hello world"), NewInt(5))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "he..." {
		t.Errorf("expected 'he...', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrQuote(t *testing.T) {
	fn, ok := Builtins["strQuote"]
	if !ok {
		t.Fatal("strQuote builtin not found")
	}

	result := fn.Fn(NewString(`hello "world"`))
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

func TestBuiltinStrUnquote(t *testing.T) {
	fn, ok := Builtins["strUnquote"]
	if !ok {
		t.Fatal("strUnquote builtin not found")
	}

	result := fn.Fn(NewString(`"hello"`))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrToInt(t *testing.T) {
	fn, ok := Builtins["strToInt"]
	if !ok {
		t.Fatal("strToInt builtin not found")
	}

	result := fn.Fn(NewString("123"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 123 {
		t.Errorf("expected 123, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrReverse(t *testing.T) {
	fn, ok := Builtins["strReverse"]
	if !ok {
		t.Fatal("strReverse builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "olleh" {
		t.Errorf("expected 'olleh', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrGetLastComponent(t *testing.T) {
	fn, ok := Builtins["strGetLastComponent"]
	if !ok {
		t.Fatal("strGetLastComponent builtin not found")
	}

	result := fn.Fn(NewString("/path/to/file.txt"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "file.txt" {
		t.Errorf("expected 'file.txt', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("a,b,c"), NewString(","))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "c" {
		t.Errorf("expected 'c', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrFindDiffPos(t *testing.T) {
	fn, ok := Builtins["strFindDiffPos"]
	if !ok {
		t.Fatal("strFindDiffPos builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewString("hallo"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("expected 1, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("hello"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -1 {
		t.Errorf("expected -1 for identical strings, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("hello world"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5 for different length, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrDiff(t *testing.T) {
	fn, ok := Builtins["strDiff"]
	if !ok {
		t.Fatal("strDiff builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewString("hallo"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 1 {
		t.Errorf("expected 1 diff, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("hello"), NewString("hello"))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected 0 diffs for identical strings, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrFindAllSub(t *testing.T) {
	fn, ok := Builtins["strFindAllSub"]
	if !ok {
		t.Fatal("strFindAllSub builtin not found")
	}

	result := fn.Fn(NewString("hello hello hello"), NewString("hello"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 occurrences, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("hello world"), NewString("xyz"))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected 0 occurrences for not found, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString("hello"), NewString(""))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected 0 occurrences for empty substring, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetTextSimilarity(t *testing.T) {
	fn, ok := Builtins["getTextSimilarity"]
	if !ok {
		t.Fatal("getTextSimilarity builtin not found")
	}

	result := fn.Fn(NewString("hello"), NewString("hello"))
	floatResult, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value < 0.99 || floatResult.Value > 1.01 {
		t.Errorf("expected ~1.0 for identical strings, got %f", floatResult.Value)
	}

	result = fn.Fn(NewString("hello"), NewString("world"))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value < 0 || floatResult.Value > 1 {
		t.Errorf("expected similarity between 0 and 1, got %f", floatResult.Value)
	}

	result = fn.Fn(NewString(""), NewString("hello"))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 0 {
		t.Errorf("expected 0 for empty string, got %f", floatResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinFuzzyFind(t *testing.T) {
	fn, ok := Builtins["fuzzyFind"]
	if !ok {
		t.Fatal("fuzzyFind builtin not found")
	}

	arr := NewArray([]Object{
		NewString("hello"),
		NewString("help"),
		NewString("world"),
	})

	result := fn.Fn(arr, NewString("hel"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) < 1 {
		t.Error("expected at least 1 match")
	}

	result = fn.Fn(arr, NewString("xyz"))
	arrResult, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("expected 0 matches for not found, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}
