// pkg/objects/builtin_jwt_test.go
// Note: JWT builtins have been moved to crypto module
// These tests verify the exported functions still work correctly
package objects

import (
	"testing"
)

func TestBuiltinGenJwtToken(t *testing.T) {
	payload := NewMapWithCapacity(2)
	payload.Pairs[NewString("sub").HashKey()] = MapPair{
		Key:   NewString("sub"),
		Value: NewString("user1"),
	}
	payload.Pairs[NewString("userId").HashKey()] = MapPair{
		Key:   NewString("userId"),
		Value: NewInt(116),
	}

	tests := []struct {
		name        string
		args        []Object
		wantError   bool
		description string
	}{
		{
			name:        "basic token",
			args:        []Object{payload, NewString("my_secret"), NewString("-type")},
			wantError:   false,
			description: "Generate basic JWT token with -type option",
		},
		{
			name:        "token with expiration",
			args:        []Object{payload, NewString("secret"), NewString("-expire=3600")},
			wantError:   false,
			description: "Generate token with 1 hour expiration",
		},
		{
			name:        "token with base64 secret",
			args:        []Object{payload, NewString("bXlfc2VjcmV0"), NewString("-base64")},
			wantError:   false,
			description: "Generate token with base64-encoded secret",
		},
		{
			name:        "wrong number of arguments",
			args:        []Object{payload},
			wantError:   true,
			description: "Should return error for wrong number of arguments",
		},
		{
			name:        "first arg not map",
			args:        []Object{NewString("not a map"), NewString("secret")},
			wantError:   true,
			description: "Should return error when first arg is not a map",
		},
		{
			name:        "second arg not string",
			args:        []Object{payload, NewInt(123)},
			wantError:   true,
			description: "Should return error when second arg is not a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinGenJwtToken(tt.args...)

			if tt.wantError {
				if _, ok := result.(*Error); !ok {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}

			if str, ok := result.(*String); !ok {
				t.Errorf("expected String, got %s", result.Type())
				return
			} else {
				parts := splitString(str.Value, ".")
				if len(parts) != 3 {
					t.Errorf("JWT token should have 3 parts, got %d", len(parts))
				}
			}
		})
	}
}

func TestBuiltinParseJwtToken(t *testing.T) {
	payload := NewMapWithCapacity(2)
	payload.Pairs[NewString("sub").HashKey()] = MapPair{
		Key:   NewString("sub"),
		Value: NewString("user1"),
	}
	payload.Pairs[NewString("userId").HashKey()] = MapPair{
		Key:   NewString("userId"),
		Value: NewInt(116),
	}

	secret := []byte("test_secret")

	tokenStr, err := GenJwtToken(payload, secret, true, false, 0)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		args        []Object
		wantError   bool
		checkClaims bool
		description string
	}{
		{
			name:        "parse valid token",
			args:        []Object{NewString(tokenStr), NewString("test_secret")},
			wantError:   false,
			checkClaims: true,
			description: "Parse valid token with correct secret",
		},
		{
			name:        "parse with wrong secret",
			args:        []Object{NewString(tokenStr), NewString("wrong_secret")},
			wantError:   true,
			description: "Should fail with wrong secret",
		},
		{
			name:        "parse without validation",
			args:        []Object{NewString(tokenStr), NewString("any_secret"), NewString("-noValidate")},
			wantError:   false,
			description: "Parse without validation",
		},
		{
			name:        "parse with base64 secret",
			args:        []Object{NewString(tokenStr), NewString("dGVzdF9zZWNyZXQ="), NewString("-base64")},
			wantError:   false,
			description: "Parse with base64-encoded secret",
		},
		{
			name:        "wrong number of arguments",
			args:        []Object{NewString(tokenStr)},
			wantError:   true,
			description: "Should return error for wrong number of arguments",
		},
		{
			name:        "first arg not string",
			args:        []Object{NewInt(123), NewString("secret")},
			wantError:   true,
			description: "Should return error when first arg is not a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinParseJwtToken(tt.args...)

			if tt.wantError {
				if _, ok := result.(*Error); !ok {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}

			if tt.checkClaims {
				if _, ok := result.(*Map); !ok {
					t.Errorf("expected Map, got %s", result.Type())
				}
			}
		})
	}
}

func TestInterfaceToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected Object
	}{
		{"nil", nil, NULL},
		{"bool true", true, &Bool{Value: true}},
		{"bool false", false, &Bool{Value: false}},
		{"int", int(42), NewInt(42)},
		{"int64", int64(42), NewInt(42)},
		{"float64", float64(3.14), NewFloat(3.14)},
		{"string", "hello", NewString("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterfaceToObject(tt.input)
			if tt.name == "nil" {
				if result != NULL {
					t.Errorf("expected NULL, got %v", result)
				}
				return
			}
			if result.Type() != tt.expected.Type() {
				t.Errorf("expected type %s, got %s", tt.expected.Type(), result.Type())
			}
		})
	}
}

func TestGenJwtToken(t *testing.T) {
	// Create payload map
	payload := NewMapWithCapacity(2)
	payload.Pairs[NewString("sub").HashKey()] = MapPair{
		Key:   NewString("sub"),
		Value: NewString("user1"),
	}
	payload.Pairs[NewString("userId").HashKey()] = MapPair{
		Key:   NewString("userId"),
		Value: NewInt(116),
	}

	tests := []struct {
		name          string
		payload       *Map
		secret        []byte
		withType      bool
		base64Secret  bool
		expireSeconds int64
		wantError     bool
		description   string
	}{
		{
			name:          "basic token",
			payload:       payload,
			secret:        []byte("my_secret"),
			withType:      true,
			base64Secret:  false,
			expireSeconds: 0,
			wantError:     false,
			description:   "Generate basic JWT token",
		},
		{
			name:          "token with expiration",
			payload:       payload,
			secret:        []byte("secret"),
			withType:      true,
			base64Secret:  false,
			expireSeconds: 3600,
			wantError:     false,
			description:   "Generate token with 1 hour expiration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenJwtToken(tt.payload, tt.secret, tt.withType, tt.base64Secret, tt.expireSeconds)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got token: %s", result)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify token has three parts
			parts := splitString(result, ".")
			if len(parts) != 3 {
				t.Errorf("JWT token should have 3 parts, got %d", len(parts))
			}
		})
	}
}

func TestParseJwtToken(t *testing.T) {
	// First generate a token
	payload := NewMapWithCapacity(2)
	payload.Pairs[NewString("sub").HashKey()] = MapPair{
		Key:   NewString("sub"),
		Value: NewString("user1"),
	}
	payload.Pairs[NewString("userId").HashKey()] = MapPair{
		Key:   NewString("userId"),
		Value: NewInt(116),
	}

	secret := []byte("test_secret")

	// Generate token
	tokenStr, err := GenJwtToken(payload, secret, true, false, 0)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Now parse the token
	t.Run("parse valid token", func(t *testing.T) {
		claims, err := ParseJwtToken(tokenStr, secret, false, false)
		if err != nil {
			t.Errorf("failed to parse token: %v", err)
			return
		}

		// Check sub claim
		if sub, ok := claims["sub"]; ok {
			if str, ok := sub.(string); ok {
				if str != "user1" {
					t.Errorf("expected sub=user1, got %s", str)
				}
			}
		} else {
			t.Error("sub claim not found")
		}
	})

	t.Run("parse with wrong secret", func(t *testing.T) {
		_, err := ParseJwtToken(tokenStr, []byte("wrong_secret"), false, false)
		if err == nil {
			t.Error("expected error with wrong secret")
		}
	})

	t.Run("parse without validation", func(t *testing.T) {
		claims, err := ParseJwtToken(tokenStr, []byte("any_secret"), false, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if claims == nil {
			t.Error("expected claims map")
		}
	})
}

// Helper function to split string
func splitString(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
