// pkg/objects/builtin_env_test.go
package objects

import (
	"os"
	"testing"
)

func TestBuiltinGetEnv(t *testing.T) {
	fn, ok := Builtins["getEnv"]
	if !ok {
		t.Fatal("getEnv builtin not found")
	}

	// Test getting existing env var
	os.Setenv("XXLANG_TEST_GET", "test_value")
	result := fn.Fn(NewString("XXLANG_TEST_GET"))
	if s, ok := result.(*String); !ok || s.Value != "test_value" {
		t.Errorf("expected 'test_value', got %v", result)
	}

	// Test getting non-existent env var
	result = fn.Fn(NewString("NON_EXISTENT_VAR_12345"))
	if result != NULL {
		t.Errorf("expected NULL for non-existent var, got %v", result)
	}

	// Test with default value
	result = fn.Fn(NewString("NON_EXISTENT_VAR_12345"), NewString("default"))
	if s, ok := result.(*String); !ok || s.Value != "default" {
		t.Errorf("expected 'default', got %v", result)
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestBuiltinSetEnv(t *testing.T) {
	fn, ok := Builtins["setEnv"]
	if !ok {
		t.Fatal("setEnv builtin not found")
	}

	result := fn.Fn(NewString("XXLANG_TEST_SET"), NewString("test_value"))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	// Verify it was set
	if os.Getenv("XXLANG_TEST_SET") != "test_value" {
		t.Error("env var was not set correctly")
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("key"))
	if !isError(result) {
		t.Error("expected error for missing value")
	}

	result = fn.Fn(NewInt(123), NewString("value"))
	if !isError(result) {
		t.Error("expected error for non-string key")
	}

	result = fn.Fn(NewString("key"), NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string value")
	}
}

func TestBuiltinGetOSName(t *testing.T) {
	fn, ok := Builtins["getOSName"]
	if !ok {
		t.Fatal("getOSName builtin not found")
	}

	result := fn.Fn()
	if s, ok := result.(*String); !ok || s.Value == "" {
		t.Errorf("expected non-empty string, got %v", result)
	}

	// Test error case
	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetOSArch(t *testing.T) {
	fn, ok := Builtins["getOSArch"]
	if !ok {
		t.Fatal("getOSArch builtin not found")
	}

	result := fn.Fn()
	if s, ok := result.(*String); !ok || s.Value == "" {
		t.Errorf("expected non-empty string, got %v", result)
	}

	// Test error case
	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetOSArgs(t *testing.T) {
	fn, ok := Builtins["getOSArgs"]
	if !ok {
		t.Fatal("getOSArgs builtin not found")
	}

	result := fn.Fn()
	if _, ok := result.(*Array); !ok {
		t.Errorf("expected Array, got %T", result)
	}

	// Test error case
	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetPid(t *testing.T) {
	fn, ok := Builtins["getPid"]
	if !ok {
		t.Fatal("getPid builtin not found")
	}

	result := fn.Fn()
	if _, ok := result.(*Int); !ok {
		t.Errorf("expected Int, got %T", result)
	}
}

func TestBuiltinGetPPid(t *testing.T) {
	fn, ok := Builtins["getPPid"]
	if !ok {
		t.Fatal("getPPid builtin not found")
	}

	result := fn.Fn()
	if _, ok := result.(*Int); !ok {
		t.Errorf("expected Int, got %T", result)
	}
}

func TestBuiltinHostname(t *testing.T) {
	fn, ok := Builtins["hostname"]
	if !ok {
		t.Fatal("hostname builtin not found")
	}

	result := fn.Fn()
	if s, ok := result.(*String); !ok || s.Value == "" {
		t.Errorf("expected non-empty string, got %v", result)
	}
}

func TestBuiltinGetSysInfo(t *testing.T) {
	fn, ok := Builtins["getSysInfo"]
	if !ok {
		t.Fatal("getSysInfo builtin not found")
	}

	result := fn.Fn()
	if m, ok := result.(*Map); !ok {
		t.Errorf("expected Map, got %T", result)
	} else if len(m.Pairs) == 0 {
		t.Error("expected non-empty map for sys info")
	}
}

func TestBuiltinExit(t *testing.T) {
	fn, ok := Builtins["exit"]
	if !ok {
		t.Fatal("exit builtin not found")
	}

	// Can't really test exit as it terminates the process
	_ = fn
}

func TestBuiltinGetAppPath(t *testing.T) {
	fn, ok := Builtins["getAppPath"]
	if !ok {
		t.Fatal("getAppPath builtin not found")
	}

	result := fn.Fn()
	if s, ok := result.(*String); !ok || s.Value == "" {
		t.Errorf("expected non-empty string, got %v", result)
	}

	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetAppDir(t *testing.T) {
	fn, ok := Builtins["getAppDir"]
	if !ok {
		t.Fatal("getAppDir builtin not found")
	}

	result := fn.Fn()
	if s, ok := result.(*String); !ok || s.Value == "" {
		t.Errorf("expected non-empty string, got %v", result)
	}

	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestGetDirFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/file.txt", "/path/to"},
		{"C:\\path\\to\\file.txt", "C:\\path\\to"},
		{"file.txt", "."},
		{"/file.txt", ""},
	}

	for _, tt := range tests {
		result := getDirFromPath(tt.path)
		if result != tt.expected {
			t.Errorf("getDirFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
		}
	}
}
