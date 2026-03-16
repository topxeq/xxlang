// pkg/stdlib/crypto_test.go
package stdlib

import (
	"encoding/hex"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callCryptoFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("crypto")
	if mod == nil {
		panic("crypto module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestCryptoMD5(t *testing.T) {
	result := callCryptoFunc("md5", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("md5() should return String, got %T", result)
	}
	// MD5 of "hello" is 5d41402abc4b2a76b9719d911017c592
	expected := "5d41402abc4b2a76b9719d911017c592"
	if r.Value != expected {
		t.Errorf("md5('hello') = %s, want %s", r.Value, expected)
	}
}

func TestCryptoSHA1(t *testing.T) {
	result := callCryptoFunc("sha1", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("sha1() should return String, got %T", result)
	}
	// SHA1 of "hello" is aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if r.Value != expected {
		t.Errorf("sha1('hello') = %s, want %s", r.Value, expected)
	}
}

func TestCryptoSHA256(t *testing.T) {
	result := callCryptoFunc("sha256", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("sha256() should return String, got %T", result)
	}
	// SHA256 of "hello" is 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if r.Value != expected {
		t.Errorf("sha256('hello') = %s, want %s", r.Value, expected)
	}
}

func TestCryptoSHA512(t *testing.T) {
	result := callCryptoFunc("sha512", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("sha512() should return String, got %T", result)
	}
	// SHA512 of "hello" (first 16 chars): 9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caad
	if len(r.Value) != 128 {
		t.Errorf("sha512() length = %d, want 128", len(r.Value))
	}
}

func TestCryptoHMAC(t *testing.T) {
	// HMAC-SHA256("key", "message")
	result := callCryptoFunc("hmacSha256", String("key"), String("message"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hmacSha256() should return String, got %T", result)
	}
	// Known value: HMAC-SHA256 of "message" with key "key"
	expected := "6e9ef29b75fffc5b7abae527d58fdadb2fe42e7219011976917343065f58ed4a"
	if r.Value != expected {
		t.Errorf("hmacSha256('key', 'message') = %s, want %s", r.Value, expected)
	}
}

func TestCryptoBase64(t *testing.T) {
	// Encode
	result := callCryptoFunc("base64Encode", String("hello world"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64Encode() should return String, got %T", result)
	}
	expected := "aGVsbG8gd29ybGQ="
	if r.Value != expected {
		t.Errorf("base64Encode('hello world') = %s, want %s", r.Value, expected)
	}

	// Decode
	result = callCryptoFunc("base64Decode", String("aGVsbG8gd29ybGQ="))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("base64Decode() should return String, got %T", result)
	}
	if r.Value != "hello world" {
		t.Errorf("base64Decode() = %s, want 'hello world'", r.Value)
	}

	// Invalid base64
	result = callCryptoFunc("base64Decode", String("invalid!@#$"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("base64Decode() with invalid input should return Error, got %T", result)
	}
}

func TestCryptoBase64URL(t *testing.T) {
	// URL-safe encoding
	result := callCryptoFunc("base64URLEncode", String("hello world"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base64URLEncode() should return String, got %T", result)
	}
	// Should not contain + or / (URL-safe)
	result2 := callCryptoFunc("base64URLDecode", r)
	r2, ok := result2.(*objects.String)
	if !ok {
		t.Fatalf("base64URLDecode() should return String, got %T", result2)
	}
	if r2.Value != "hello world" {
		t.Errorf("base64URLDecode() = %s, want 'hello world'", r2.Value)
	}
}

func TestCryptoHex(t *testing.T) {
	// Encode
	result := callCryptoFunc("hexEncode", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hexEncode() should return String, got %T", result)
	}
	expected := "68656c6c6f"
	if r.Value != expected {
		t.Errorf("hexEncode('hello') = %s, want %s", r.Value, expected)
	}

	// Decode
	result = callCryptoFunc("hexDecode", String("68656c6c6f"))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("hexDecode() should return String, got %T", result)
	}
	if r.Value != "hello" {
		t.Errorf("hexDecode() = %s, want 'hello'", r.Value)
	}

	// Invalid hex
	result = callCryptoFunc("hexDecode", String("invalid hex"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("hexDecode() with invalid input should return Error, got %T", result)
	}
}

func TestCryptoRandomBytes(t *testing.T) {
	result := callCryptoFunc("randomBytes", Int(16))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("randomBytes() should return String, got %T", result)
	}
	// Should be 32 hex characters (16 bytes = 32 hex chars)
	if len(r.Value) != 32 {
		t.Errorf("randomBytes(16) length = %d, want 32", len(r.Value))
	}

	// Test it's valid hex
	_, err := hex.DecodeString(r.Value)
	if err != nil {
		t.Errorf("randomBytes() result is not valid hex: %v", err)
	}
}

func TestCryptoRandomHex(t *testing.T) {
	result := callCryptoFunc("randomHex", Int(16))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("randomHex() should return String, got %T", result)
	}
	// Should be 32 hex characters (16 bytes = 32 hex chars)
	if len(r.Value) != 32 {
		t.Errorf("randomHex(16) length = %d, want 32", len(r.Value))
	}
}

func TestCryptoRandomBase64(t *testing.T) {
	result := callCryptoFunc("randomBase64", Int(16))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("randomBase64() should return String, got %T", result)
	}
	// Base64 of 16 bytes should be 24 chars (with padding)
	if len(r.Value) < 22 {
		t.Errorf("randomBase64(16) length = %d, too short", len(r.Value))
	}
}

func TestCryptoUUID(t *testing.T) {
	result := callCryptoFunc("uuid")
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("uuid() should return String, got %T", result)
	}

	// UUID should be 36 chars (32 hex + 4 dashes)
	if len(r.Value) != 36 {
		t.Errorf("uuid() length = %d, want 36", len(r.Value))
	}

	// Check format: 8-4-4-4-12
	expectedPattern := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	if len(r.Value) != len(expectedPattern) {
		t.Errorf("uuid() format incorrect")
	}

	// Generate multiple UUIDs and check they're unique
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result = callCryptoFunc("uuid")
		r = result.(*objects.String)
		if uuids[r.Value] {
			t.Errorf("uuid() generated duplicate: %s", r.Value)
		}
		uuids[r.Value] = true
	}
}

func TestCryptoErrorCases(t *testing.T) {
	// md5 with wrong args
	result := callCryptoFunc("md5")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("md5() with no args should return Error, got %T", result)
	}

	// sha256 with wrong type
	result = callCryptoFunc("sha256", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("sha256(int) should return Error, got %T", result)
	}

	// hmacSha256 with wrong args
	result = callCryptoFunc("hmacSha256", String("key"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("hmacSha256() with 1 arg should return Error, got %T", result)
	}

	// randomBytes with negative
	result = callCryptoFunc("randomBytes", Int(-1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("randomBytes(-1) should return Error, got %T", result)
	}

	// randomBytes with too large
	result = callCryptoFunc("randomBytes", Int(2*1024*1024))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("randomBytes(2MB) should return Error, got %T", result)
	}
}
