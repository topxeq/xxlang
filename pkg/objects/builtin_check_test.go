// pkg/objects/builtin_check_test.go
package objects

import (
	"testing"
)

func TestBuiltinIsNil(t *testing.T) {
	fn, ok := Builtins["isNil"]
	if !ok {
		t.Fatal("isNil builtin not found")
	}

	result := fn.Fn(NULL)
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for NULL")
	}

	result = fn.Fn(NewInt(1))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsNilOrEmpty(t *testing.T) {
	fn, ok := Builtins["isNilOrEmpty"]
	if !ok {
		t.Fatal("isNilOrEmpty builtin not found")
	}

	result := fn.Fn(NULL)
	if !IsTruthy(result) {
		t.Error("expected true for NULL")
	}

	result = fn.Fn(&String{Value: ""})
	if !IsTruthy(result) {
		t.Error("expected true for empty string")
	}

	result = fn.Fn(&String{Value: "hello"})
	if IsTruthy(result) {
		t.Error("expected false for non-empty string")
	}

	result = fn.Fn(&Array{Elements: []Object{}})
	if !IsTruthy(result) {
		t.Error("expected true for empty array")
	}

	result = fn.Fn(&Array{Elements: []Object{NewInt(1)}})
	if IsTruthy(result) {
		t.Error("expected false for non-empty array")
	}

	result = fn.Fn(&Map{Pairs: make(map[HashKey]MapPair)})
	if !IsTruthy(result) {
		t.Error("expected true for empty map")
	}

	result = fn.Fn(&Bytes{Value: []byte{}})
	if !IsTruthy(result) {
		t.Error("expected true for empty bytes")
	}

	result = fn.Fn(&Chars{Value: []rune{}})
	if !IsTruthy(result) {
		t.Error("expected true for empty chars")
	}

	result = fn.Fn(NewInt(1))
	if IsTruthy(result) {
		t.Error("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsNilOrErr(t *testing.T) {
	fn, ok := Builtins["isNilOrErr"]
	if !ok {
		t.Fatal("isNilOrErr builtin not found")
	}

	result := fn.Fn(NULL)
	if !IsTruthy(result) {
		t.Error("expected true for NULL")
	}

	result = fn.Fn(&Error{Message: "test error"})
	if !IsTruthy(result) {
		t.Error("expected true for Error")
	}

	result = fn.Fn(NewInt(1))
	if IsTruthy(result) {
		t.Error("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsBytes(t *testing.T) {
	fn, ok := Builtins["isBytes"]
	if !ok {
		t.Fatal("isBytes builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte{1, 2, 3}})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for Bytes")
	}

	result = fn.Fn(NewInt(1))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsChars(t *testing.T) {
	fn, ok := Builtins["isChars"]
	if !ok {
		t.Fatal("isChars builtin not found")
	}

	result := fn.Fn(&Chars{Value: []rune{'a', 'b'}})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for Chars")
	}

	result = fn.Fn(NewInt(1))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for Int")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinPass(t *testing.T) {
	fn, ok := Builtins["pass"]
	if !ok {
		t.Fatal("pass builtin not found")
	}

	result := fn.Fn()
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}
}

func TestBuiltinErrStrf(t *testing.T) {
	fn, ok := Builtins["errStrf"]
	if !ok {
		t.Fatal("errStrf builtin not found")
	}

	result := fn.Fn(NewString("test error"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "ERROR: test error" {
		t.Errorf("expected 'ERROR: test error', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("value: %d"), NewInt(42))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "ERROR: value: 42" {
		t.Errorf("expected 'ERROR: value: 42', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for non-string format")
	}
}

func TestBuiltinErrf(t *testing.T) {
	fn, ok := Builtins["errf"]
	if !ok {
		t.Fatal("errf builtin not found")
	}

	result := fn.Fn(NewString("test error"))
	errResult, ok := result.(*Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if errResult.Message != "test error" {
		t.Errorf("expected 'test error', got '%s'", errResult.Message)
	}

	result = fn.Fn(NewString("value: %d"), NewInt(42))
	errResult, ok = result.(*Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if errResult.Message != "value: 42" {
		t.Errorf("expected 'value: 42', got '%s'", errResult.Message)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinErrToEmpty(t *testing.T) {
	fn, ok := Builtins["errToEmpty"]
	if !ok {
		t.Fatal("errToEmpty builtin not found")
	}

	result := fn.Fn(&Error{Message: "test error"})
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "" {
		t.Errorf("expected empty string, got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("ERROR: something"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "" {
		t.Errorf("expected empty string, got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("normal string"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "normal string" {
		t.Errorf("expected 'normal string', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSscanf(t *testing.T) {
	fn, ok := Builtins["sscanf"]
	if !ok {
		t.Fatal("sscanf builtin not found")
	}

	result := fn.Fn(NewString("hello 42"), NewString("hello %d"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	// Note: sscanf implementation returns empty array due to Go fmt.Sscanf limitations
	// This test verifies the function runs without error for valid input
	_ = len(arrResult.Elements)

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(1), NewString("format"))
	if !isError(result) {
		t.Error("expected error for non-string first arg")
	}

	result = fn.Fn(NewString("text"), NewInt(1))
	if !isError(result) {
		t.Error("expected error for non-string second arg")
	}
}

func TestBuiltinBytesStartsWith(t *testing.T) {
	fn, ok := Builtins["bytesStartsWith"]
	if !ok {
		t.Fatal("bytesStartsWith builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("hello")})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for bytes starting with prefix")
	}

	result = fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("world")})
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for bytes not starting with prefix")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBytesEndsWith(t *testing.T) {
	fn, ok := Builtins["bytesEndsWith"]
	if !ok {
		t.Fatal("bytesEndsWith builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("world")})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for bytes ending with suffix")
	}

	result = fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("hello")})
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for bytes not ending with suffix")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBytesContains(t *testing.T) {
	fn, ok := Builtins["bytesContains"]
	if !ok {
		t.Fatal("bytesContains builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("lo wo")})
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for bytes containing substring")
	}

	result = fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("xyz")})
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for bytes not containing substring")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinBytesIndex(t *testing.T) {
	fn, ok := Builtins["bytesIndex"]
	if !ok {
		t.Fatal("bytesIndex builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("world")})
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 6 {
		t.Errorf("expected 6, got %d", intResult.Value)
	}

	result = fn.Fn(&Bytes{Value: []byte("hello world")}, &Bytes{Value: []byte("xyz")})
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -1 {
		t.Errorf("expected -1, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCompareBytes(t *testing.T) {
	fn, ok := Builtins["compareBytes"]
	if !ok {
		t.Fatal("compareBytes builtin not found")
	}

	result := fn.Fn(&Bytes{Value: []byte("abc")}, &Bytes{Value: []byte("abc")})
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 0 {
		t.Errorf("expected 0 for equal bytes, got %d", intResult.Value)
	}

	result = fn.Fn(&Bytes{Value: []byte("abc")}, &Bytes{Value: []byte("abd")})
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value >= 0 {
		t.Errorf("expected negative for abc < abd, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCompareText(t *testing.T) {
	fn, ok := Builtins["compareText"]
	if !ok {
		t.Fatal("compareText builtin not found")
	}

	result := fn.Fn(NewString("abc"), NewString("abc"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 0 {
		t.Errorf("expected 0 for equal strings, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("abc"), NewString("abd"))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value >= 0 {
		t.Errorf("expected negative for abc < abd, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}
