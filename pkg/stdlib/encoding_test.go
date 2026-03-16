// pkg/stdlib/encoding_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callEncodingFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("encoding")
	if mod == nil {
		panic("encoding module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestBase64Encode(t *testing.T) {
	result := callEncodingFunc("base64Encode", String("hello"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64Encode() should return String, got %T", result)
	}
	if s.Value != "aGVsbG8=" {
		t.Errorf("base64Encode('hello') = %s, want 'aGVsbG8='", s.Value)
	}
}

func TestBase64Decode(t *testing.T) {
	result := callEncodingFunc("base64Decode", String("aGVsbG8="))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64Decode() should return String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("base64Decode() = %s, want 'hello'", s.Value)
	}
}

func TestBase64DecodeError(t *testing.T) {
	result := callEncodingFunc("base64Decode", String("invalid!!!"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("base64Decode() with invalid input should return Error")
	}
}

func TestBase64URLEncode(t *testing.T) {
	result := callEncodingFunc("base64URLEncode", String("hello"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64URLEncode() should return String, got %T", result)
	}
	if s.Value != "aGVsbG8=" {
		t.Errorf("base64URLEncode('hello') = %s, want 'aGVsbG8='", s.Value)
	}
}

func TestBase64URLDecode(t *testing.T) {
	result := callEncodingFunc("base64URLDecode", String("aGVsbG8="))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64URLDecode() should return String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("base64URLDecode() = %s, want 'hello'", s.Value)
	}
}

func TestHexEncode(t *testing.T) {
	result := callEncodingFunc("hexEncode", String("hello"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hexEncode() should return String, got %T", result)
	}
	if s.Value != "68656c6c6f" {
		t.Errorf("hexEncode('hello') = %s, want '68656c6c6f'", s.Value)
	}
}

func TestHexDecode(t *testing.T) {
	result := callEncodingFunc("hexDecode", String("68656c6c6f"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hexDecode() should return String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("hexDecode() = %s, want 'hello'", s.Value)
	}
}

func TestHexDecodeError(t *testing.T) {
	result := callEncodingFunc("hexDecode", String("invalid"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("hexDecode() with invalid input should return Error")
	}
}

func TestURLEncode(t *testing.T) {
	result := callEncodingFunc("urlEncode", String("hello world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("urlEncode() should return String, got %T", result)
	}
	if s.Value != "hello+world" {
		t.Errorf("urlEncode('hello world') = %s, want 'hello+world'", s.Value)
	}
}

func TestURLDecode(t *testing.T) {
	result := callEncodingFunc("urlDecode", String("hello+world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("urlDecode() should return String, got %T", result)
	}
	if s.Value != "hello world" {
		t.Errorf("urlDecode() = %s, want 'hello world'", s.Value)
	}
}

func TestParseURL(t *testing.T) {
	result := callEncodingFunc("parseURL", String("https://example.com/path?query=1#frag"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("parseURL() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("parseURL() length = %d, want 5", len(arr.Elements))
	}
	if arr.Elements[0].(*objects.String).Value != "https" {
		t.Errorf("parseURL() scheme = %s, want 'https'", arr.Elements[0].(*objects.String).Value)
	}
}

func TestBuildURL(t *testing.T) {
	result := callEncodingFunc("buildURL", String("https"), String("example.com"), String("/path"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("buildURL() should return String, got %T", result)
	}
	if s.Value != "https://example.com/path" {
		t.Errorf("buildURL() = %s, want 'https://example.com/path'", s.Value)
	}
}

func TestSetQuery(t *testing.T) {
	result := callEncodingFunc("setQuery", String("https://example.com"), String("key"), String("value"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("setQuery() should return String, got %T", result)
	}
	if s.Value != "https://example.com?key=value" {
		t.Errorf("setQuery() = %s, want 'https://example.com?key=value'", s.Value)
	}
}

func TestEscapeHTML(t *testing.T) {
	result := callEncodingFunc("escapeHTML", String("hello world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("escapeHTML() should return String, got %T", result)
	}
	if s.Value != "hello%20world" {
		t.Errorf("escapeHTML() = %s, want 'hello%%20world'", s.Value)
	}
}

func TestUnescapeHTML(t *testing.T) {
	result := callEncodingFunc("unescapeHTML", String("hello%20world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("unescapeHTML() should return String, got %T", result)
	}
	if s.Value != "hello world" {
		t.Errorf("unescapeHTML() = %s, want 'hello world'", s.Value)
	}
}

func TestEncodingErrors(t *testing.T) {
	// Test wrong number of args
	result := callEncodingFunc("base64Encode")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("base64Encode() with no args should return Error")
	}

	// Test wrong type
	result = callEncodingFunc("base64Encode", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("base64Encode() with non-string should return Error")
	}
}
