// pkg/stdlib/uuid_test.go
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callUUIDFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("uuid")
	if mod == nil {
		panic("uuid module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestUUIDV4(t *testing.T) {
	result := callUUIDFunc("v4")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("v4() should return String, got %T", result)
	}
	if len(s.Value) != 36 {
		t.Errorf("v4() length = %d, want 36", len(s.Value))
	}
	// Check dashes at correct positions
	if s.Value[8] != '-' || s.Value[13] != '-' || s.Value[18] != '-' || s.Value[23] != '-' {
		t.Error("v4() UUID should have dashes at positions 8, 13, 18, 23")
	}
	// Check version bit (position 14 should be '4')
	if s.Value[14] != '4' {
		t.Errorf("v4() UUID version bit should be '4', got '%c'", s.Value[14])
	}
}

func TestUUIDV4Short(t *testing.T) {
	result := callUUIDFunc("v4Short")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("v4Short() should return String, got %T", result)
	}
	if len(s.Value) != 32 {
		t.Errorf("v4Short() length = %d, want 32", len(s.Value))
	}
	if strings.Contains(s.Value, "-") {
		t.Error("v4Short() should not contain dashes")
	}
}

func TestUUIDSimple(t *testing.T) {
	result := callUUIDFunc("simple")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("simple() should return String, got %T", result)
	}
	if len(s.Value) != 16 {
		t.Errorf("simple() length = %d, want 16", len(s.Value))
	}
}

func TestUUIDTimeID(t *testing.T) {
	result := callUUIDFunc("timeID")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("timeID() should return String, got %T", result)
	}
	// timeID: 16 hex chars for timestamp + 8 hex chars for random = 24
	if len(s.Value) != 24 {
		t.Errorf("timeID() length = %d, want 24", len(s.Value))
	}
}

func TestUUIDRandom(t *testing.T) {
	result := callUUIDFunc("random")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("random() should return String, got %T", result)
	}
	if len(s.Value) != 16 {
		t.Errorf("random() default length = %d, want 16", len(s.Value))
	}
}

func TestUUIDRandomWithLength(t *testing.T) {
	result := callUUIDFunc("random", Int(8))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("random() should return String, got %T", result)
	}
	if len(s.Value) != 8 {
		t.Errorf("random(8) length = %d, want 8", len(s.Value))
	}
}

func TestUUIDHex(t *testing.T) {
	result := callUUIDFunc("hex")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hex() should return String, got %T", result)
	}
	if len(s.Value) != 32 { // 16 bytes -> 32 hex chars
		t.Errorf("hex() length = %d, want 32", len(s.Value))
	}
}

func TestUUIDHexWithLength(t *testing.T) {
	result := callUUIDFunc("hex", Int(8))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hex() should return String, got %T", result)
	}
	if len(s.Value) != 16 { // 8 bytes -> 16 hex chars
		t.Errorf("hex(8) length = %d, want 16", len(s.Value))
	}
}

func TestUUIDIsValid(t *testing.T) {
	// Valid UUID
	result := callUUIDFunc("isValid", String("550e8400-e29b-41d4-a716-446655440000"))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isValid() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isValid() should return true for valid UUID")
	}

	// Invalid - wrong length
	result = callUUIDFunc("isValid", String("not-a-uuid"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isValid() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isValid() should return false for invalid UUID")
	}

	// Invalid - wrong dash positions
	result = callUUIDFunc("isValid", String("550e8400e29b41d4a716446655440000"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isValid() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isValid() should return false for UUID without dashes")
	}
}

func TestUUIDIsValidErrors(t *testing.T) {
	result := callUUIDFunc("isValid")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isValid() with no args should return Error")
	}

	result = callUUIDFunc("isValid", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isValid() with non-string should return Error")
	}
}

func TestUUIDParse(t *testing.T) {
	result := callUUIDFunc("parse", String("550e8400-e29b-41d4-a716-446655440000"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("parse() should return String, got %T", result)
	}
	if len(s.Value) != 16 {
		t.Errorf("parse() length = %d, want 16", len(s.Value))
	}
}

func TestUUIDParseErrors(t *testing.T) {
	result := callUUIDFunc("parse")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parse() with no args should return Error")
	}

	result = callUUIDFunc("parse", String("invalid"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parse() with invalid UUID should return Error")
	}
}

func TestUUIDFormat(t *testing.T) {
	result := callUUIDFunc("format", String("550e8400e29b41d4a716446655440000"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("format() should return String, got %T", result)
	}
	expected := "550e8400-e29b-41d4-a716-446655440000"
	if s.Value != expected {
		t.Errorf("format() = %s, want %s", s.Value, expected)
	}
}

func TestUUIDFormatErrors(t *testing.T) {
	result := callUUIDFunc("format")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("format() with no args should return Error")
	}

	result = callUUIDFunc("format", String("tooshort"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("format() with short string should return Error")
	}
}
