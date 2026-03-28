// pkg/objects/builtin_jwt_test.go
// Note: JWT builtins have been moved to crypto module
// These tests verify the exported functions still work correctly
package objects

import (
	"testing"
)

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
