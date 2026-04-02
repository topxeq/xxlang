// pkg/objects/builtin_encode_test.go
package objects

import (
	"testing"
)

func TestBuiltinUrlEncode(t *testing.T) {
	fn, ok := Builtins["urlEncode"]
	if !ok {
		t.Fatal("urlEncode builtin not found")
	}

	result := fn.Fn(NewString("hello world"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinUrlDecode(t *testing.T) {
	fn, ok := Builtins["urlDecode"]
	if !ok {
		t.Fatal("urlDecode builtin not found")
	}

	result := fn.Fn(NewString("hello%2Fworld"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello/world" {
		t.Errorf("expected 'hello/world', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinUrlEncodeComponent(t *testing.T) {
	fn, ok := Builtins["urlEncodeComponent"]
	if !ok {
		t.Fatal("urlEncodeComponent builtin not found")
	}

	result := fn.Fn(NewString("hello world"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinUrlDecodeComponent(t *testing.T) {
	fn, ok := Builtins["urlDecodeComponent"]
	if !ok {
		t.Fatal("urlDecodeComponent builtin not found")
	}

	result := fn.Fn(NewString("hello%20world"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinHtmlEncode(t *testing.T) {
	fn, ok := Builtins["htmlEncode"]
	if !ok {
		t.Fatal("htmlEncode builtin not found")
	}

	result := fn.Fn(NewString("<div>hello</div>"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinHtmlDecode(t *testing.T) {
	fn, ok := Builtins["htmlDecode"]
	if !ok {
		t.Fatal("htmlDecode builtin not found")
	}

	result := fn.Fn(NewString("&lt;div&gt;hello&lt;/div&gt;"))
	_, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSha1(t *testing.T) {
	fn, ok := Builtins["sha1"]
	if !ok {
		t.Fatal("sha1 builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSha512(t *testing.T) {
	fn, ok := Builtins["sha512"]
	if !ok {
		t.Fatal("sha512 builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinHashStr(t *testing.T) {
	fn, ok := Builtins["hashStr"]
	if !ok {
		t.Fatal("hashStr builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero hash")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinToHex(t *testing.T) {
	fn, ok := Builtins["toHex"]
	if !ok {
		t.Fatal("toHex builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinUnhex(t *testing.T) {
	fn, ok := Builtins["unhex"]
	if !ok {
		t.Fatal("unhex builtin not found")
	}

	result := fn.Fn(NewString("68656c6c6f"))
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

func TestBuiltinHexToStr(t *testing.T) {
	fn, ok := Builtins["hexToStr"]
	if !ok {
		t.Fatal("hexToStr builtin not found")
	}

	result := fn.Fn(NewString("68656c6c6f"))
	_, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSimpleEncode(t *testing.T) {
	fn, ok := Builtins["simpleEncode"]
	if !ok {
		t.Fatal("simpleEncode builtin not found")
	}

	result := fn.Fn(NewString("hello"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSimpleDecode(t *testing.T) {
	fn, ok := Builtins["simpleDecode"]
	if !ok {
		t.Fatal("simpleDecode builtin not found")
	}

	result := fn.Fn(NewString("aGVsbG8g")) // base64 encoded "hello "
	_, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestUnhexChar(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{'0', 0},
		{'9', 9},
		{'a', 10},
		{'f', 15},
		{'A', 10},
		{'F', 15},
		{'g', 0},
		{'z', 0},
	}

	for _, tt := range tests {
		result := unhexChar(tt.input)
		if result != tt.expected {
			t.Errorf("unhexChar(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestBuiltinSimpleEncodeDecode(t *testing.T) {
	encFn, ok := Builtins["simpleEncode"]
	if !ok {
		t.Fatal("simpleEncode builtin not found")
	}
	decFn, ok := Builtins["simpleDecode"]
	if !ok {
		t.Fatal("simpleDecode builtin not found")
	}

	tests := []string{
		"hello",
		"world",
		"test123",
		"abc xyz",
	}

	for _, tt := range tests {
		encoded := encFn.Fn(NewString(tt))
		encStr, ok := encoded.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", encoded)
		}

		decoded := decFn.Fn(encStr)
		decStr, ok := decoded.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", decoded)
		}

		if decStr.Value != tt {
			t.Errorf("encode/decode roundtrip failed: got %q, want %q", decStr.Value, tt)
		}
	}
}
