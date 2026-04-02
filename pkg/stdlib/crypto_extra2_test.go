// pkg/stdlib/crypto_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// cryptoCall invokes a builtin from the crypto module.
func cryptoCall(name string, args ...objects.Object) objects.Object {
	mod := Get("crypto")
	if mod == nil {
		return &objects.Error{Message: "crypto module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestCrypto_Extra2_Init tests that crypto module registers all exports.
func TestCrypto_Extra2_Init(t *testing.T) {
	mod := Get("crypto")
	if mod == nil {
		t.Skip("crypto module not found")
	}
	expected := []string{
		"hexEncode", "hexDecode",
		"randomBytes", "randomHex", "uuid",
		"encryptTextByTXTE", "decryptTextByTXTE",
		"encryptDataByTXDEE", "decryptDataByTXDEE",
		"encryptTextByTXDEE", "decryptTextByTXDEE",
		"encryptData", "decryptData",
		"encryptBytes", "decryptBytes",
		"encryptText", "decryptText",
		"encryptStr", "decryptStr",
		"encryptStream", "decryptStream",
		"aesEncrypt", "aesDecrypt",
		"genJwtToken", "parseJwtToken",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in crypto module", name)
		}
	}
}

// TestCrypto_Extra2_Hex_ArgumentValidation tests hex functions.
func TestCrypto_Extra2_Hex_ArgumentValidation(t *testing.T) {
	// hexEncode with no args
	res := cryptoCall("hexEncode")
	if res.Type() != objects.ErrorType {
		t.Fatalf("hexEncode() with no args should error")
	}

	// hexEncode with wrong type
	res = cryptoCall("hexEncode", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("hexEncode() with int should error")
	}

	// hexEncode with string
	res = cryptoCall("hexEncode", String("hello"))
	if res.Type() != objects.StringType {
		t.Fatalf("hexEncode(string) should return string, got %s", res.Type())
	}

	// hexDecode with no args
	res = cryptoCall("hexDecode")
	if res.Type() != objects.ErrorType {
		t.Fatalf("hexDecode() with no args should error")
	}

	// hexDecode with wrong type
	res = cryptoCall("hexDecode", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("hexDecode() with int should error")
	}

	// hexDecode with valid hex
	res = cryptoCall("hexDecode", String("68656c6c6f"))
	if res.Type() != objects.StringType {
		t.Fatalf("hexDecode(valid) should return string, got %s", res.Type())
	}
}

// TestCrypto_Extra2_Random_ArgumentValidation tests random functions.
func TestCrypto_Extra2_Random_ArgumentValidation(t *testing.T) {
	// randomBytes with no args
	res := cryptoCall("randomBytes")
	if res.Type() != objects.ErrorType {
		t.Fatalf("randomBytes() with no args should error")
	}

	// randomBytes with wrong type
	res = cryptoCall("randomBytes", String("16"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("randomBytes() with string should error")
	}

	// randomBytes with valid int
	res = cryptoCall("randomBytes", Int(16))
	if res.Type() != objects.StringType {
		t.Fatalf("randomBytes(16) should return string, got %s", res.Type())
	}

	// randomHex with no args
	res = cryptoCall("randomHex")
	if res.Type() != objects.ErrorType {
		t.Fatalf("randomHex() with no args should error")
	}

	// randomHex with valid int
	res = cryptoCall("randomHex", Int(32))
	if res.Type() != objects.StringType {
		t.Fatalf("randomHex(32) should return string, got %s", res.Type())
	}

	// uuid with no args
	res = cryptoCall("uuid")
	if res.Type() != objects.StringType {
		t.Fatalf("uuid() should return string, got %s", res.Type())
	}
}

// TestCrypto_Extra2_Encrypt_ArgumentValidation tests encrypt functions.
func TestCrypto_Extra2_Encrypt_ArgumentValidation(t *testing.T) {
	// encryptTextByTXTE with no args
	res := cryptoCall("encryptTextByTXTE")
	if res.Type() != objects.ErrorType {
		t.Fatalf("encryptTextByTXTE() with no args should error")
	}

	// decryptTextByTXTE with no args
	res = cryptoCall("decryptTextByTXTE")
	if res.Type() != objects.ErrorType {
		t.Fatalf("decryptTextByTXTE() with no args should error")
	}

	// encryptData with no args
	res = cryptoCall("encryptData")
	if res.Type() != objects.ErrorType {
		t.Fatalf("encryptData() with no args should error")
	}

	// decryptData with no args
	res = cryptoCall("decryptData")
	if res.Type() != objects.ErrorType {
		t.Fatalf("decryptData() with no args should error")
	}

	// aesEncrypt with no args
	res = cryptoCall("aesEncrypt")
	if res.Type() != objects.ErrorType {
		t.Fatalf("aesEncrypt() with no args should error")
	}

	// aesDecrypt with no args
	res = cryptoCall("aesDecrypt")
	if res.Type() != objects.ErrorType {
		t.Fatalf("aesDecrypt() with no args should error")
	}
}

// TestCrypto_Extra2_Jwt_ArgumentValidation tests JWT functions.
func TestCrypto_Extra2_Jwt_ArgumentValidation(t *testing.T) {
	// genJwtToken with no args
	res := cryptoCall("genJwtToken")
	if res.Type() != objects.ErrorType {
		t.Fatalf("genJwtToken() with no args should error")
	}

	// parseJwtToken with no args
	res = cryptoCall("parseJwtToken")
	if res.Type() != objects.ErrorType {
		t.Fatalf("parseJwtToken() with no args should error")
	}

	// parseJwtToken with wrong type
	res = cryptoCall("parseJwtToken", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("parseJwtToken() with int should error")
	}
}
