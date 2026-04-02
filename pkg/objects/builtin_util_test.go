// pkg/objects/builtin_util_test.go
package objects

import (
	"testing"
)

func TestBuiltinSprintf(t *testing.T) {
	fn, ok := Builtins["sprintf"]
	if !ok {
		t.Fatal("sprintf builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("hello %s"), NewString("world"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("%d + %d = %d"), NewInt(1), NewInt(2), NewInt(3))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "1 + 2 = 3" {
		t.Errorf("expected '1 + 2 = 3', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string format")
	}
}

func TestBuiltinToBoolUtil(t *testing.T) {
	fn, ok := Builtins["toBool"]
	if !ok {
		t.Fatal("toBool builtin not found")
	}

	testCases := []struct {
		input    Object
		expected bool
	}{
		{&Bool{Value: true}, true},
		{&Bool{Value: false}, false},
		{NewInt(1), true},
		{NewInt(0), false},
		{NewFloat(1.0), true},
		{NewFloat(0.0), false},
		{NewString("true"), true},
		{NewString(""), false},
		{NULL, false},
	}

	for _, tc := range testCases {
		result := fn.Fn(tc.input)
		boolResult, ok := result.(*Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T for input type %T", result, tc.input)
		}
		if boolResult.Value != tc.expected {
			t.Errorf("expected %v for %T, got %v", tc.expected, tc.input, boolResult.Value)
		}
	}

	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinToInt(t *testing.T) {
	fn, ok := Builtins["toInt"]
	if !ok {
		t.Fatal("toInt builtin not found")
	}

	result := fn.Fn(NewInt(42))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 42 {
		t.Errorf("expected 42, got %d", intResult.Value)
	}

	result = fn.Fn(NewFloat(3.7))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("expected 3, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("123"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 123 {
		t.Errorf("expected 123, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("1111"), NewInt(2))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 15 {
		t.Errorf("expected 15 (binary 1111), got %d", intResult.Value)
	}

	result = fn.Fn(&Bool{Value: true})
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("expected 1, got %d", intResult.Value)
	}

	result = fn.Fn(&Bool{Value: false})
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 0 {
		t.Errorf("expected 0, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for invalid string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinToFloat(t *testing.T) {
	fn, ok := Builtins["toFloat"]
	if !ok {
		t.Fatal("toFloat builtin not found")
	}

	result := fn.Fn(NewFloat(3.14))
	floatResult, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", floatResult.Value)
	}

	result = fn.Fn(NewInt(5))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 5.0 {
		t.Errorf("expected 5.0, got %f", floatResult.Value)
	}

	result = fn.Fn(NewString("2.5"))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 2.5 {
		t.Errorf("expected 2.5, got %f", floatResult.Value)
	}

	result = fn.Fn(&Bool{Value: true})
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 1.0 {
		t.Errorf("expected 1.0, got %f", floatResult.Value)
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for invalid string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsUndefined(t *testing.T) {
	fn, ok := Builtins["isUndefined"]
	if !ok {
		t.Fatal("isUndefined builtin not found")
	}

	result := fn.Fn(NULL)
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for NULL")
	}

	result = fn.Fn(NewString("hello"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for String")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsCallable(t *testing.T) {
	fn, ok := Builtins["isCallable"]
	if !ok {
		t.Fatal("isCallable builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for String")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsIterable(t *testing.T) {
	fn, ok := Builtins["isIterable"]
	if !ok {
		t.Fatal("isIterable builtin not found")
	}

	testCases := []struct {
		input    Object
		expected bool
	}{
		{NewArray([]Object{}), true},
		{NewString("hello"), true},
		{NewMap(make(map[HashKey]MapPair)), true},
		{NewOrderedMap(), true},
		{NewChars([]rune("hello")), true},
		{NewBytes([]byte("hello")), true},
		{NewInt(42), false},
		{NewFloat(3.14), false},
		{&Bool{Value: true}, false},
	}

	for _, tc := range testCases {
		result := fn.Fn(tc.input)
		boolResult, ok := result.(*Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T for input type %T", result, tc.input)
		}
		if boolResult.Value != tc.expected {
			t.Errorf("expected %v for %T, got %v", tc.expected, tc.input, boolResult.Value)
		}
	}

	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsError(t *testing.T) {
	fn, ok := Builtins["isError"]
	if !ok {
		t.Fatal("isError builtin not found")
	}

	result := fn.Fn(&Error{Message: "test error"})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for Error")
	}

	result = fn.Fn(NewString("hello"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for String")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsErrUtil(t *testing.T) {
	fn, ok := Builtins["isErr"]
	if !ok {
		t.Fatal("isErr builtin not found")
	}

	result := fn.Fn(&Error{Message: "test error"})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for Error")
	}

	result = fn.Fn(NewString("TXERROR: some error"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for TXERROR: string")
	}

	result = fn.Fn(NewString("normal string"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for normal string")
	}

	result = fn.Fn(NewInt(42))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinError(t *testing.T) {
	fn, ok := Builtins["error"]
	if !ok {
		t.Fatal("error builtin not found")
	}

	result := fn.Fn(NewString("something went wrong"))
	errResult, ok := result.(*Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if errResult.Message != "something went wrong" {
		t.Errorf("expected 'something went wrong', got '%s'", errResult.Message)
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetErrStr(t *testing.T) {
	fn, ok := Builtins["getErrStr"]
	if !ok {
		t.Fatal("getErrStr builtin not found")
	}

	result := fn.Fn(&Error{Message: "test error"})
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "test error" {
		t.Errorf("expected 'test error', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("TXERROR: extracted error"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != " extracted error" {
		t.Errorf("expected ' extracted error', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("normal string"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "normal string" {
		t.Errorf("expected 'normal string', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewInt(42))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "" {
		t.Errorf("expected empty string for Int, got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsErrStr(t *testing.T) {
	fn, ok := Builtins["isErrStr"]
	if !ok {
		t.Fatal("isErrStr builtin not found")
	}

	result := fn.Fn(NewString("TXERROR: error message"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for TXERROR: string")
	}

	result = fn.Fn(NewString("normal string"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for normal string")
	}

	result = fn.Fn(NewInt(42))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTypeCode(t *testing.T) {
	fn, ok := Builtins["typeCode"]
	if !ok {
		t.Fatal("typeCode builtin not found")
	}

	result := fn.Fn(NULL)
	if _, ok := result.(*Int); !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn(&Bool{Value: true})
	if _, ok := result.(*Int); !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn(NewInt(42))
	if _, ok := result.(*Int); !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn(NewFloat(3.14))
	if _, ok := result.(*Int); !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn(NewString("hello"))
	if _, ok := result.(*Int); !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCloneUtil(t *testing.T) {
	fn, ok := Builtins["clone"]
	if !ok {
		t.Fatal("clone builtin not found")
	}

	result := fn.Fn(NewInt(42))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 42 {
		t.Errorf("expected 42, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strResult.Value)
	}

	arr := NewArray([]Object{NewInt(1), NewInt(2)})
	result = fn.Fn(arr)
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSwap(t *testing.T) {
	fn, ok := Builtins["swap"]
	if !ok {
		t.Fatal("swap builtin not found")
	}

	arr := NewArray([]Object{NewInt(1), NewInt(2), NewInt(3)})
	result := fn.Fn(arr, NewInt(0), NewInt(2))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if arrResult.Elements[0].(*Int).Value != 3 {
		t.Errorf("expected first element to be 3, got %d", arrResult.Elements[0].(*Int).Value)
	}
	if arrResult.Elements[2].(*Int).Value != 1 {
		t.Errorf("expected third element to be 1, got %d", arrResult.Elements[2].(*Int).Value)
	}

	result = fn.Fn(NewString("hello"), NewInt(0), NewInt(1))
	if !isError(result) {
		t.Error("expected error for non-array first argument")
	}

	result = fn.Fn(arr, NewString("a"), NewInt(1))
	if !isError(result) {
		t.Error("expected error for non-int second argument")
	}

	result = fn.Fn(arr, NewInt(0), NewString("b"))
	if !isError(result) {
		t.Error("expected error for non-int third argument")
	}

	result = fn.Fn(arr, NewInt(-1), NewInt(0))
	if !isError(result) {
		t.Error("expected error for out of bounds index")
	}

	result = fn.Fn(arr, NewInt(0), NewInt(100))
	if !isError(result) {
		t.Error("expected error for out of bounds index")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCoalesce(t *testing.T) {
	fn, ok := Builtins["coalesce"]
	if !ok {
		t.Fatal("coalesce builtin not found")
	}

	result := fn.Fn(NewString("first"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "first" {
		t.Errorf("expected 'first', got '%s'", strResult.Value)
	}

	result = fn.Fn(NULL, NewString("second"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "second" {
		t.Errorf("expected 'second', got '%s'", strResult.Value)
	}

	result = fn.Fn(&Error{Message: "error"}, NewString("fallback"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", strResult.Value)
	}

	result = fn.Fn(NULL, &Error{Message: "error"}, NewInt(42))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 42 {
		t.Errorf("expected 42, got %d", intResult.Value)
	}

	result = fn.Fn(NULL, NULL, NULL)
	_, isNull := result.(*Null)
	if !isNull {
		t.Errorf("expected NULL when all args are null/error")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinDefaultVal(t *testing.T) {
	fn, ok := Builtins["defaultVal"]
	if !ok {
		t.Fatal("defaultVal builtin not found")
	}

	result := fn.Fn(NewString("value"), NewString("default"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "value" {
		t.Errorf("expected 'value', got '%s'", strResult.Value)
	}

	result = fn.Fn(NULL, NewString("default"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "default" {
		t.Errorf("expected 'default', got '%s'", strResult.Value)
	}

	result = fn.Fn(&Error{Message: "error"}, NewInt(100))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 100 {
		t.Errorf("expected 100, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("only one"))
	if !isError(result) {
		t.Error("expected error for single arg")
	}
}

func TestObjectToString(t *testing.T) {
	tests := []struct {
		obj      Object
		expected string
	}{
		{NewString("hello"), "hello"},
		{NewInt(123), "123"},
		{NewFloat(3.14), "3.14"},
		{TRUE, "true"},
		{FALSE, "false"},
	}

	for _, tt := range tests {
		result := objectToString(tt.obj)
		if result != tt.expected {
			t.Errorf("objectToString(%T) = %q, want %q", tt.obj, result, tt.expected)
		}
	}
}

func TestGetTypeCode(t *testing.T) {
	tests := []struct {
		obj      Object
		expected int
	}{
		{NULL, int(TypeCodeNull)},
		{TRUE, int(TypeCodeBool)},
		{NewInt(42), int(TypeCodeInt)},
		{NewFloat(3.14), int(TypeCodeFloat)},
		{NewString("hello"), int(TypeCodeString)},
		{NewArray([]Object{}), int(TypeCodeArray)},
		{NewMapWithCapacity(0), int(TypeCodeMap)},
		{NewBytes([]byte{}), int(TypeCodeBytes)},
		{NewChars([]rune{}), int(TypeCodeChars)},
	}

	for _, tt := range tests {
		result := getTypeCode(tt.obj)
		if result != tt.expected {
			t.Errorf("getTypeCode(%T) = %d, want %d", tt.obj, result, tt.expected)
		}
	}
}

func TestIsNil(t *testing.T) {
	if !isNil(nil) {
		t.Error("expected nil to be nil")
	}

	if isNil(NewInt(42)) {
		t.Error("expected Int to not be nil")
	}

	if isNil(NewString("hello")) {
		t.Error("expected String to not be nil")
	}
}
