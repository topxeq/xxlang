// pkg/objects/builtin_misc_test.go
package objects

import (
	"os"
	"testing"
)

func TestBuiltinGetRandomInt(t *testing.T) {
	fn, ok := Builtins["getRandomInt"]
	if !ok {
		t.Fatal("getRandomInt builtin not found")
	}

	result := fn.Fn(NewInt(10))
	_, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for non-int arg")
	}
}

func TestBuiltinGetRandomFloat(t *testing.T) {
	fn, ok := Builtins["getRandomFloat"]
	if !ok {
		t.Fatal("getRandomFloat builtin not found")
	}

	result := fn.Fn()
	floatResult, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value < 0 || floatResult.Value >= 1 {
		t.Errorf("expected value in [0, 1), got %f", floatResult.Value)
	}
}

func TestBuiltinGetRandomStr(t *testing.T) {
	fn, ok := Builtins["getRandomStr"]
	if !ok {
		t.Fatal("getRandomStr builtin not found")
	}

	result := fn.Fn(NewInt(10))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) != 10 {
		t.Errorf("expected length 10, got %d", len(strResult.Value))
	}

	result = fn.Fn(NewInt(5), NewString("abc"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) != 5 {
		t.Errorf("expected length 5, got %d", len(strResult.Value))
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinJoinUrlPath(t *testing.T) {
	fn, ok := Builtins["joinUrlPath"]
	if !ok {
		t.Fatal("joinUrlPath builtin not found")
	}

	result := fn.Fn(NewString("https://example.com"), NewString("/path/to/resource"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "https://example.com/path/to/resource" {
		t.Errorf("unexpected result: %s", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinParseUrl(t *testing.T) {
	fn, ok := Builtins["parseUrl"]
	if !ok {
		t.Fatal("parseUrl builtin not found")
	}

	result := fn.Fn(NewString("https://example.com:8080/path?query=value#anchor"))
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) == 0 {
		t.Error("expected non-empty map")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestBuiltinParseQuery(t *testing.T) {
	fn, ok := Builtins["parseQuery"]
	if !ok {
		t.Fatal("parseQuery builtin not found")
	}

	result := fn.Fn(NewString("name=John&age=30"))
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) == 0 {
		t.Error("expected non-empty map")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsHttps(t *testing.T) {
	fn, ok := Builtins["isHttps"]
	if !ok {
		t.Fatal("isHttps builtin not found")
	}

	result := fn.Fn(NewString("https://example.com"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected true for https URL")
	}

	result = fn.Fn(NewString("http://example.com"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected false for http URL")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGenToken(t *testing.T) {
	fn, ok := Builtins["genToken"]
	if !ok {
		t.Fatal("genToken builtin not found")
	}

	result := fn.Fn(NewInt(32))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) == 0 {
		t.Error("expected non-empty token")
	}

	result = fn.Fn()
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
}

func TestBuiltinCheckOtpCode(t *testing.T) {
	fn, ok := Builtins["checkOtpCode"]
	if !ok {
		t.Fatal("checkOtpCode builtin not found")
	}

	result := fn.Fn(NewString("JBSWY3DPEHPK3PXP"), NewString("123456"))
	_, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCreateTempDir(t *testing.T) {
	fn, ok := Builtins["createTempDir"]
	if !ok {
		t.Fatal("createTempDir builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty temp dir path")
	}
	defer os.RemoveAll(strResult.Value)

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first arg")
	}

	result = fn.Fn(NewString(""), NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string second arg")
	}
}

func TestBuiltinCreateTempFile(t *testing.T) {
	fn, ok := Builtins["createTempFile"]
	if !ok {
		t.Fatal("createTempFile builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty temp file path")
	}
	defer os.Remove(strResult.Value)

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first arg")
	}
}

func TestBuiltinLookPath(t *testing.T) {
	fn, ok := Builtins["lookPath"]
	if !ok {
		t.Fatal("lookPath builtin not found")
	}

	result := fn.Fn(NewString("go"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty path for 'go'")
	}

	result = fn.Fn(NewString("nonexistent_binary_12345"))
	if result != NULL {
		t.Errorf("expected NULL for nonexistent binary, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestBuiltinChangeDir(t *testing.T) {
	fn, ok := Builtins["changeDir"]
	if !ok {
		t.Fatal("changeDir builtin not found")
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Skip("could not get current directory")
	}
	defer os.Chdir(originalDir)

	tempDir := os.TempDir()
	result := fn.Fn(NewString(tempDir))
	if isError(result) {
		t.Errorf("unexpected error: %v", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestStrToInt(t *testing.T) {
	tests := []struct {
		s        string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-456", -456},
		{"", 0},
		{"abc", 0},
	}

	for _, tt := range tests {
		result := strToInt(tt.s)
		if result != tt.expected {
			t.Errorf("strToInt(%q) = %d, want %d", tt.s, result, tt.expected)
		}
	}
}

func TestCleanPath(t *testing.T) {
	result := cleanPath("/path/to/../file.txt")
	if result == "" {
		t.Error("expected non-empty cleaned path")
	}

	result = cleanPath("/path/./file.txt")
	if result == "" {
		t.Error("expected non-empty cleaned path")
	}
}

func TestIsAbsPath(t *testing.T) {
	// Test Unix-style paths (works on all platforms)
	result := isAbsPath("/path/to/file.txt")
	if !result {
		t.Error("expected true for absolute Unix path")
	}

	result = isAbsPath("path/to/file.txt")
	if result {
		t.Error("expected false for relative path")
	}

	// Note: Windows-style paths only work as absolute on Windows
	// On Linux, filepath.IsAbs returns false for "C:\path"
}
