// pkg/stdlib/coverage_stdlib_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestMathModuleRound tests round function
func TestMathModuleRound(t *testing.T) {
	tests := []struct {
		name     string
		args     []objects.Object
		expected interface{}
	}{
		{"round int", []objects.Object{Int(42)}, int64(42)},
		{"round float", []objects.Object{Float(3.7)}, int64(4)},
		{"round with precision", []objects.Object{Float(3.14159), Int(2)}, 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callStdlibFunc("math", "round", tt.args...)
			if result == nil {
				t.Fatal("round returned nil")
			}
			switch e := tt.expected.(type) {
			case int64:
				n, ok := result.(*objects.Int)
				if !ok {
					t.Fatalf("expected Int, got %T", result)
				}
				if n.Value != e {
					t.Errorf("round() = %d, want %d", n.Value, e)
				}
			case float64:
				f, ok := result.(*objects.Float)
				if !ok {
					t.Fatalf("expected Float, got %T", result)
				}
				if f.Value != e {
					t.Errorf("round() = %f, want %f", f.Value, e)
				}
			}
		})
	}
}

// TestMathModuleSinCos tests trigonometric functions
func TestMathModuleSinCos(t *testing.T) {
	result := callStdlibFunc("math", "sin", Float(0))
	if result == nil {
		t.Fatal("sin returned nil")
	}
	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value != 0.0 {
		t.Errorf("sin(0) = %f, want 0.0", f.Value)
	}

	result = callStdlibFunc("math", "cos", Float(0))
	if result == nil {
		t.Fatal("cos returned nil")
	}
	f, ok = result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value != 1.0 {
		t.Errorf("cos(0) = %f, want 1.0", f.Value)
	}
}

// TestMathModuleLog tests logarithmic functions
func TestMathModuleLog(t *testing.T) {
	result := callStdlibFunc("math", "log", Float(2.718281828))
	if result == nil {
		t.Fatal("log returned nil")
	}
	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value < 0.9 || f.Value > 1.1 {
		t.Errorf("log(e) = %f, want ~1.0", f.Value)
	}
}

// TestMathModuleLog10 tests log10 function
func TestMathModuleLog10(t *testing.T) {
	result := callStdlibFunc("math", "log10", Float(100))
	if result == nil {
		t.Fatal("log10 returned nil")
	}
	f, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value < 1.9 || f.Value > 2.1 {
		t.Errorf("log10(100) = %f, want ~2.0", f.Value)
	}
}

// TestStringsModuleReplace tests replace function
func TestStringsModuleReplace(t *testing.T) {
	result := callStdlibFunc("strings", "replace", String("hello world"), String("world"), String("universe"), Int(-1))
	if result == nil {
		t.Fatal("replace returned nil")
	}
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "hello universe" {
		t.Errorf("replace() = %q, want 'hello universe'", s.Value)
	}
}

// TestStringsModuleRepeat tests repeat function
func TestStringsModuleRepeat(t *testing.T) {
	result := callStdlibFunc("strings", "repeat", String("ab"), Int(3))
	if result == nil {
		t.Fatal("repeat returned nil")
	}
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "ababab" {
		t.Errorf("repeat('ab', 3) = %q, want 'ababab'", s.Value)
	}
}

// TestArrayModuleReverseExtra tests reverse function
func TestArrayModuleReverseExtra(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{Int(1), Int(2), Int(3)}}
	result := callStdlibFunc("array", "reverse", arr)
	if result == nil {
		t.Fatal("reverse returned nil")
	}
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arrResult.Elements))
	}
	if arrResult.Elements[0].(*objects.Int).Value != 3 {
		t.Error("reverse first element should be 3")
	}
}

// TestArrayModuleContainsExtra tests contains function
func TestArrayModuleContainsExtra(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{Int(1), Int(2), Int(3)}}
	result := callStdlibFunc("array", "contains", arr, Int(2))
	if result == nil {
		t.Fatal("contains returned nil")
	}
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !b.Value {
		t.Error("contains([1,2,3], 2) should be true")
	}

	result = callStdlibFunc("array", "contains", arr, Int(99))
	if result == nil {
		t.Fatal("contains returned nil")
	}
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if b.Value {
		t.Error("contains([1,2,3], 99) should be false")
	}
}

// TestArrayModuleIndexOfExtra tests indexOf function
func TestArrayModuleIndexOfExtra(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{Int(1), Int(2), Int(3)}}
	result := callStdlibFunc("array", "indexOf", arr, Int(2))
	if result == nil {
		t.Fatal("indexOf returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value != 1 {
		t.Errorf("indexOf([1,2,3], 2) = %d, want 1", n.Value)
	}
}

// TestArrayModuleUnique tests unique function
func TestArrayModuleUnique(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{Int(1), Int(2), Int(1), Int(3), Int(2)}}
	result := callStdlibFunc("array", "unique", arr)
	if result == nil {
		t.Fatal("unique returned nil")
	}
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("unique([1,2,1,3,2]) should have 3 elements, got %d", len(arrResult.Elements))
	}
}

// TestEncodingModuleBase64 tests base64 encoding/decoding
func TestEncodingModuleBase64(t *testing.T) {
	original := String("hello world")
	encoded := callStdlibFunc("encoding", "base64Encode", original)
	if encoded == nil {
		t.Fatal("base64Encode returned nil")
	}

	decoded := callStdlibFunc("encoding", "base64Decode", encoded)
	if decoded == nil {
		t.Fatal("base64Decode returned nil")
	}
	s, ok := decoded.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", decoded)
	}
	if s.Value != "hello world" {
		t.Errorf("base64Decode() = %q, want 'hello world'", s.Value)
	}
}

// TestEncodingModuleHex tests hex encoding/decoding
func TestEncodingModuleHex(t *testing.T) {
	original := String("hello")
	encoded := callStdlibFunc("encoding", "hexEncode", original)
	if encoded == nil {
		t.Fatal("hexEncode returned nil")
	}

	decoded := callStdlibFunc("encoding", "hexDecode", encoded)
	if decoded == nil {
		t.Fatal("hexDecode returned nil")
	}
	s, ok := decoded.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", decoded)
	}
	if s.Value != "hello" {
		t.Errorf("hexDecode() = %q, want 'hello'", s.Value)
	}
}

// TestCryptoModuleMd5Sha256 tests hash functions
func TestCryptoModuleMd5Sha256(t *testing.T) {
	data := String("hello")

	md5Result := callStdlibFunc("crypto", "md5", data)
	if md5Result == nil {
		t.Fatal("md5 returned nil")
	}
	md5Str, ok := md5Result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", md5Result)
	}
	if len(md5Str.Value) != 32 {
		t.Errorf("md5 hash length should be 32, got %d", len(md5Str.Value))
	}

	shaResult := callStdlibFunc("crypto", "sha256", data)
	if shaResult == nil {
		t.Fatal("sha256 returned nil")
	}
	shaStr, ok := shaResult.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", shaResult)
	}
	if len(shaStr.Value) != 64 {
		t.Errorf("sha256 hash length should be 64, got %d", len(shaStr.Value))
	}
}

// TestJsonModuleEncodeDecode tests JSON encoding/decoding
func TestJsonModuleEncodeDecode(t *testing.T) {
	obj := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("name").HashKey():  {Key: objects.NewString("name"), Value: objects.NewString("test")},
			objects.NewString("value").HashKey(): {Key: objects.NewString("value"), Value: objects.NewInt(42)},
		},
	}

	encoded := callStdlibFunc("json", "encode", obj)
	if encoded == nil {
		t.Fatal("json encode returned nil")
	}
	_, ok := encoded.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", encoded)
	}
}
