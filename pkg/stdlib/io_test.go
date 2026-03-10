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

func TestIOPrintf(t *testing.T) {
	// Test printf with format strings
	tests := []struct {
		name   string
		format string
		args   []objects.Object
	}{
		{"string format", "Hello %s", []objects.Object{String("World")}},
		{"int format", "Number: %d", []objects.Object{Int(42)}},
		{"float format", "Float: %f", []objects.Object{Float(3.14)}},
		{"bool format", "Bool: %v", []objects.Object{Bool(true)}},
		{"multiple args", "%s %d %f", []objects.Object{String("test"), Int(10), Float(2.5)}},
		{"no placeholders", "no placeholders", []objects.Object{}},
		{"empty format", "", []objects.Object{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// printf returns null
			result := callIOFunc("printf", append([]objects.Object{String(tt.format)}, tt.args...)...)
			if _, ok := result.(*objects.Null); !ok {
				t.Errorf("printf should return null, got %T", result)
			}
		})
	}
}

func TestIOPrintfErrors(t *testing.T) {
	// printf with no args
	result := callIOFunc("printf")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("printf with 0 args should return error, got %T", result)
	}

	// printf with non-string format
	result = callIOFunc("printf", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("printf with non-string format should return error, got %T", result)
	}
}

func TestIOReadFileEdgeCases(t *testing.T) {
	// Test readFile with empty path
	result := callIOFunc("readFile", String(""))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("readFile with empty path should return error, got %T", result)
	}

	// Test readFile with non-string argument
	result = callIOFunc("readFile", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("readFile with non-string should return error, got %T", result)
	}

	// Test readFile with directory path
	tmpDir := os.TempDir()
	result = callIOFunc("readFile", String(tmpDir))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("readFile with directory path should return error, got %T", result)
	}
}

func TestIOWriteFileEdgeCases(t *testing.T) {
	// Test writeFile with non-string filename
	result := callIOFunc("writeFile", Int(42), String("content"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("writeFile with non-string filename should return error, got %T", result)
	}

	// Test writeFile with non-string content
	result = callIOFunc("writeFile", String("/tmp/test.txt"), Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("writeFile with non-string content should return error, got %T", result)
	}

	// Test writeFile with empty content
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "xxlang_test_empty.txt")
	defer os.Remove(testFile)

	result = callIOFunc("writeFile", String(testFile), String(""))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("writeFile with empty content should return null, got %T", result)
	}

	// Verify file was created and is empty
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read file: %v", err)
	}
	if string(content) != "" {
		t.Errorf("file content = %q, want empty", content)
	}

	// Test writeFile to invalid path (should fail)
	result = callIOFunc("writeFile", String("/nonexistent/directory/file.txt"), String("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("writeFile to invalid path should return error, got %T", result)
	}
}

func TestIOPrintPrintln(t *testing.T) {
	// print with multiple args
	result := callIOFunc("print", String("hello"), Int(42))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("print should return null, got %T", result)
	}

	// println with multiple args
	result = callIOFunc("println", String("hello"), Int(42), Bool(true))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("println should return null, got %T", result)
	}

	// print with no args
	result = callIOFunc("print")
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("print with no args should return null, got %T", result)
	}

	// println with no args
	result = callIOFunc("println")
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("println with no args should return null, got %T", result)
	}
}
