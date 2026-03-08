// pkg/stdlib/io_test.go
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callIOFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/io")
	if mod == nil {
		panic("std/io module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestIOWriteReadFile(t *testing.T) {
	// Create temp file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "xxlang_test_io.txt")
	defer os.Remove(testFile)

	// Test writeFile
	content := String("Hello, Xxlang!")
	result := callIOFunc("writeFile", String(testFile), content)
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("writeFile should return null, got %T", result)
	}

	// Test readFile
	result = callIOFunc("readFile", String(testFile))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("readFile should return string, got %T", result)
	}
	if r.Value != "Hello, Xxlang!" {
		t.Errorf("readFile content = %q, want %q", r.Value, "Hello, Xxlang!")
	}

	// Test readFile with non-existent file
	result = callIOFunc("readFile", String("/nonexistent/file/path.txt"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("readFile of non-existent file should return error, got %T", result)
	}
}

func TestIOAppendFile(t *testing.T) {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "xxlang_test_append.txt")
	defer os.Remove(testFile)

	// Write initial content
	callIOFunc("writeFile", String(testFile), String("Hello"))

	// Append more content
	result := callIOFunc("appendFile", String(testFile), String(" World"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("appendFile should return null, got %T", result)
	}

	// Verify content
	result = callIOFunc("readFile", String(testFile))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("readFile should return string, got %T", result)
	}
	if r.Value != "Hello World" {
		t.Errorf("appended content = %q, want %q", r.Value, "Hello World")
	}
}

func TestIOExists(t *testing.T) {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "xxlang_test_exists.txt")

	// File should not exist
	result := callIOFunc("exists", String(testFile))
	r, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("exists should return bool, got %T", result)
	}
	if r.Value {
		t.Errorf("exists(%q) should be false for non-existent file", testFile)
	}

	// Create file
	os.WriteFile(testFile, []byte("test"), 0644)
	defer os.Remove(testFile)

	// File should exist now
	result = callIOFunc("exists", String(testFile))
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("exists should return bool, got %T", result)
	}
	if !r.Value {
		t.Errorf("exists(%q) should be true for existing file", testFile)
	}
}

func TestIORemove(t *testing.T) {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "xxlang_test_remove.txt")

	// Create file
	os.WriteFile(testFile, []byte("test"), 0644)

	// Remove file
	result := callIOFunc("remove", String(testFile))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("remove should return null, got %T", result)
	}

	// Verify file is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("file should be removed")
	}

	// Remove non-existent file should return error
	result = callIOFunc("remove", String("/nonexistent/file.txt"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("remove of non-existent file should return error, got %T", result)
	}
}

func TestIOMkdir(t *testing.T) {
	tmpDir := os.TempDir()
	testDir := filepath.Join(tmpDir, "xxlang_test_dir", "nested", "path")
	defer os.RemoveAll(filepath.Join(tmpDir, "xxlang_test_dir"))

	// Create nested directories
	result := callIOFunc("mkdir", String(testDir))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("mkdir should return null, got %T", result)
	}

	// Verify directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("directory should be created")
	}
}

func TestIOCwd(t *testing.T) {
	result := callIOFunc("cwd")
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("cwd should return string, got %T", result)
	}
	if r.Value == "" {
		t.Errorf("cwd should return non-empty string")
	}

	// Verify it's the actual working directory
	actual, _ := os.Getwd()
	if r.Value != actual {
		t.Errorf("cwd = %q, want %q", r.Value, actual)
	}
}

func TestIOEnv(t *testing.T) {
	// Set a test environment variable
	testKey := "XXLANG_TEST_VAR"
	testValue := "test_value_123"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	// Test env() returns the value
	result := callIOFunc("env", String(testKey))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("env should return string, got %T", result)
	}
	if r.Value != testValue {
		t.Errorf("env(%q) = %q, want %q", testKey, r.Value, testValue)
	}

	// Test env() returns null for non-existent variable
	result = callIOFunc("env", String("NON_EXISTENT_VAR_12345"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("env of non-existent var should return null, got %T", result)
	}
}

func TestIOSetEnv(t *testing.T) {
	testKey := "XXLANG_TEST_SET_VAR"
	defer os.Unsetenv(testKey)

	// Set environment variable
	result := callIOFunc("setEnv", String(testKey), String("new_value"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("setEnv should return null, got %T", result)
	}

	// Verify it was set
	value := os.Getenv(testKey)
	if value != "new_value" {
		t.Errorf("setEnv failed: got %q, want %q", value, "new_value")
	}
}

func TestIOArgs(t *testing.T) {
	result := callIOFunc("args")
	r, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("args should return array, got %T", result)
	}

	// Args should at least contain the program name
	if len(r.Elements) < 1 {
		t.Errorf("args should return at least one element")
	}

	// First arg should be a string
	if _, ok := r.Elements[0].(*objects.String); !ok {
		t.Errorf("args[0] should be string, got %T", r.Elements[0])
	}
}

func TestIOErrorCases(t *testing.T) {
	// writeFile with wrong number of args
	result := callIOFunc("writeFile", String("test.txt"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("writeFile with 1 arg should return error, got %T", result)
	}

	// readFile with wrong number of args
	result = callIOFunc("readFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("readFile with 0 args should return error, got %T", result)
	}

	// exists with wrong number of args
	result = callIOFunc("exists")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("exists with 0 args should return error, got %T", result)
	}

	// env with wrong number of args
	result = callIOFunc("env")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("env with 0 args should return error, got %T", result)
	}

	// setEnv with wrong number of args
	result = callIOFunc("setEnv", String("KEY"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("setEnv with 1 arg should return error, got %T", result)
	}
}
