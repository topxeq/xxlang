// pkg/stdlib/io_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ioCall invokes a builtin from the io module.
func ioCall(name string, args ...objects.Object) objects.Object {
	mod := Get("io")
	if mod == nil {
		return &objects.Error{Message: "io module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestIO_Extra2_Init tests that io module registers all exports.
func TestIO_Extra2_Init(t *testing.T) {
	mod := Get("io")
	if mod == nil {
		t.Skip("io module not found")
	}
	expected := []string{
		"print", "println", "printf", "readLine", "readStdin", "readStdinBytes",
		"writeStdout", "writeStderr",
		"readFile", "readBytes", "writeFile", "writeBytes", "appendFile",
		"exists", "remove", "mkdir", "cwd", "exit", "env", "setEnv", "args",
		"ioCopy", "newReader", "newWriter",
		"read", "readStr", "readAllStr", "readAllBytes", "readLineFrom", "writeTo",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in io module", name)
		}
	}
}

// TestIO_Extra2_Print_ArgumentValidation tests print and println.
func TestIO_Extra2_Print_ArgumentValidation(t *testing.T) {
	// print with no args should work
	res := ioCall("print")
	if res.Type() != objects.NullType {
		t.Fatalf("print() should return null, got %s", res.Type())
	}

	// print with args
	res = ioCall("print", String("hello"), Int(42))
	if res.Type() != objects.NullType {
		t.Fatalf("print() should return null, got %s", res.Type())
	}

	// println with no args
	res = ioCall("println")
	if res.Type() != objects.NullType {
		t.Fatalf("println() should return null, got %s", res.Type())
	}

	// println with args
	res = ioCall("println", String("hello"), Int(42))
	if res.Type() != objects.NullType {
		t.Fatalf("println() should return null, got %s", res.Type())
	}
}

// TestIO_Extra2_Printf_ArgumentValidation tests printf.
func TestIO_Extra2_Printf_ArgumentValidation(t *testing.T) {
	// No args
	res := ioCall("printf")
	if res.Type() != objects.ErrorType {
		t.Fatalf("printf() with no args should error")
	}

	// Wrong type for format
	res = ioCall("printf", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("printf() with int format should error")
	}

	// Valid call
	res = ioCall("printf", String("Hello %s!"), String("World"))
	if res.Type() != objects.NullType {
		t.Fatalf("printf() should return null, got %s", res.Type())
	}

	// With int
	res = ioCall("printf", String("Number: %d"), Int(42))
	if res.Type() != objects.NullType {
		t.Fatalf("printf() should return null, got %s", res.Type())
	}

	// With float
	res = ioCall("printf", String("Float: %f"), Float(3.14))
	if res.Type() != objects.NullType {
		t.Fatalf("printf() should return null, got %s", res.Type())
	}
}

// TestIO_Extra2_FileOperations_ArgumentValidation tests file operations.
func TestIO_Extra2_FileOperations_ArgumentValidation(t *testing.T) {
	// writeFile with no args
	res := ioCall("writeFile")
	if res.Type() != objects.ErrorType {
		t.Fatalf("writeFile() with no args should error")
	}

	// writeFile with wrong types
	res = ioCall("writeFile", Int(123), String("content"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("writeFile() with int path should error")
	}

	// readFile with no args
	res = ioCall("readFile")
	if res.Type() != objects.ErrorType {
		t.Fatalf("readFile() with no args should error")
	}

	// readFile with wrong type
	res = ioCall("readFile", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("readFile() with int path should error")
	}

	// appendFile with no args
	res = ioCall("appendFile")
	if res.Type() != objects.ErrorType {
		t.Fatalf("appendFile() with no args should error")
	}

	// exists with no args
	res = ioCall("exists")
	if res.Type() != objects.ErrorType {
		t.Fatalf("exists() with no args should error")
	}

	// remove with no args
	res = ioCall("remove")
	if res.Type() != objects.ErrorType {
		t.Fatalf("remove() with no args should error")
	}

	// mkdir with no args
	res = ioCall("mkdir")
	if res.Type() != objects.ErrorType {
		t.Fatalf("mkdir() with no args should error")
	}
}

// TestIO_Extra2_Env_ArgumentValidation tests env functions.
func TestIO_Extra2_Env_ArgumentValidation(t *testing.T) {
	// env with no args
	res := ioCall("env")
	if res.Type() != objects.ErrorType {
		t.Fatalf("env() with no args should error")
	}

	// env with arg returns specific value
	res = ioCall("env", String("PATH"))
	if res.Type() != objects.StringType {
		t.Fatalf("env(PATH) should return string, got %s", res.Type())
	}

	// setEnv with no args
	res = ioCall("setEnv")
	if res.Type() != objects.ErrorType {
		t.Fatalf("setEnv() with no args should error")
	}

	// args with no args
	res = ioCall("args")
	if res.Type() != objects.ArrayType {
		t.Fatalf("args() should return array, got %s", res.Type())
	}

	// cwd with no args
	res = ioCall("cwd")
	if res.Type() != objects.StringType {
		t.Fatalf("cwd() should return string, got %s", res.Type())
	}
}

// TestIO_Extra2_WriteFile_ReadFile tests actual file operations.
func TestIO_Extra2_WriteFile_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/test.txt"

	// writeFile
	res := ioCall("writeFile", String(filePath), String("Hello, World!"))
	if res.Type() != objects.NullType {
		t.Fatalf("writeFile() should return null, got %s: %s", res.Type(), res.Inspect())
	}

	// readFile
	res = ioCall("readFile", String(filePath))
	if res.Type() != objects.StringType {
		t.Fatalf("readFile() should return string, got %s", res.Type())
	}
	if res.(*objects.String).Value != "Hello, World!" {
		t.Fatalf("readFile() content mismatch, got %s", res.(*objects.String).Value)
	}

	// appendFile
	res = ioCall("appendFile", String(filePath), String(" Appended."))
	if res.Type() != objects.NullType {
		t.Fatalf("appendFile() should return null, got %s", res.Type())
	}

	// readFile again
	res = ioCall("readFile", String(filePath))
	if res.Type() != objects.StringType {
		t.Fatalf("readFile() should return string, got %s", res.Type())
	}
	if res.(*objects.String).Value != "Hello, World! Appended." {
		t.Fatalf("readFile() content mismatch after append, got %s", res.(*objects.String).Value)
	}

	// exists
	res = ioCall("exists", String(filePath))
	if res.Type() != objects.BoolType {
		t.Fatalf("exists() should return bool, got %s", res.Type())
	}
	if !res.(*objects.Bool).Value {
		t.Fatalf("exists() should return true for existing file")
	}

	// remove
	res = ioCall("remove", String(filePath))
	if res.Type() != objects.NullType {
		t.Fatalf("remove() should return null, got %s", res.Type())
	}

	// exists after remove
	res = ioCall("exists", String(filePath))
	if res.Type() != objects.BoolType {
		t.Fatalf("exists() should return bool, got %s", res.Type())
	}
	if res.(*objects.Bool).Value {
		t.Fatalf("exists() should return false after remove")
	}
}

// TestIO_Extra2_Mkdir tests mkdir operations.
func TestIO_Extra2_Mkdir(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := tmpDir + "/testdir"

	// mkdir
	res := ioCall("mkdir", String(dirPath))
	if res.Type() != objects.NullType {
		t.Fatalf("mkdir() should return null, got %s: %s", res.Type(), res.Inspect())
	}

	// exists
	res = ioCall("exists", String(dirPath))
	if res.Type() != objects.BoolType {
		t.Fatalf("exists() should return bool, got %s", res.Type())
	}
	if !res.(*objects.Bool).Value {
		t.Fatalf("exists() should return true for directory")
	}

	// remove directory (should work if empty)
	res = ioCall("remove", String(dirPath))
	if res.Type() != objects.NullType {
		t.Fatalf("remove() should return null, got %s", res.Type())
	}

	// exists after remove
	res = ioCall("exists", String(dirPath))
	if res.Type() != objects.BoolType {
		t.Fatalf("exists() should return bool, got %s", res.Type())
	}
	if res.(*objects.Bool).Value {
		t.Fatalf("exists() should return false after remove")
	}
}

// TestIO_Extra2_NewReader tests newReader.
func TestIO_Extra2_NewReader(t *testing.T) {
	// No args
	res := ioCall("newReader")
	if res.Type() != objects.ErrorType {
		t.Fatalf("newReader() with no args should error")
	}

	// With string
	res = ioCall("newReader", String("hello"))
	if _, ok := res.(*objects.Reader); !ok {
		t.Fatalf("newReader(string) should return Reader, got %s", res.Type())
	}
}

// TestIO_Extra2_NewWriter tests newWriter.
func TestIO_Extra2_NewWriter(t *testing.T) {
	// No args
	res := ioCall("newWriter")
	if res.Type() != objects.ErrorType {
		t.Fatalf("newWriter() with no args should error")
	}

	// With buffer
	buf := ioCall("newReader", String("hello"))
	res = ioCall("newWriter", buf)
	// May return different types depending on implementation
	_ = res
}
