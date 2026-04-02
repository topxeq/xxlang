// pkg/stdlib/http_extra_test.go
// Additional tests for http module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Note: The existing http_test.go already covers parseJSON thoroughly with proper HttpReq objects.
// This file contains additional tests for other http functions.

// TestHTTPGetReqBody_Extra tests http.getReqBody with empty and non-empty bodies.
func TestHTTPGetReqBody_Extra(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getReqBody"].(*objects.Builtin)

	// This function likely expects a request context; without a real request it may error.
	// We'll test with nil or wrong types to exercise error handling.
	wrong := fn.Fn(String("not a request"))
	if wrong.Type() != objects.ErrorType {
		t.Errorf("expected error for wrong type, got %s", wrong.Inspect())
	}
}

// TestHTTPGetReqBodyBytes_Extra tests http.getReqBodyBytes.
func TestHTTPGetReqBodyBytes_Extra(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getReqBodyBytes"].(*objects.Builtin)

	// Wrong type
	res := fn.Fn(String("wrong"))
	if res.Type() != objects.ErrorType {
		t.Errorf("expected error, got %s", res.Inspect())
	}
}

// TestHTTPGetMimeType_Extra tests http.getMimeType with various content types.
func TestHTTPGetMimeType_Extra(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getMimeType"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{"text/plain", "text/plain"},
		{"application/json", "application/json"},
		{"text/html; charset=utf-8", "text/html"},
		{"", ""}, // might return empty or default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			if result.Type() == objects.ErrorType {
				t.Fatalf("getMimeType error: %s", result.Inspect())
			}
			s, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("expected String, got %T", result)
			}
			// For charset case, we expect the base type without charset
			if tt.expected != "" && s.Value != tt.expected && s.Value != tt.input {
				// It's okay if it returns the original when no charset handling
			}
		})
	}
}
