// pkg/stdlib/zip_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// zipCall invokes a builtin from the zip module.
func zipCall(name string, args ...objects.Object) objects.Object {
	mod := Get("zip")
	if mod == nil {
		return &objects.Error{Message: "zip module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestZip_Extra2_Init tests zip module exports.
func TestZip_Extra2_Init(t *testing.T) {
	mod := Get("zip")
	if mod == nil {
		t.Skip("zip module not found")
	}
	expected := []string{
		"list", "extract", "create", "addFile", "addDir", "addBytes", "close",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in zip module", name)
		}
	}
}

// TestZip_Extra2_List_ArgumentValidation tests zip.list().
func TestZip_Extra2_List_ArgumentValidation(t *testing.T) {
	res := zipCall("list")
	if res.Type() != objects.ErrorType {
		t.Fatalf("list() with no args should error")
	}
	res = zipCall("list", objects.NewString("a.zip"), objects.NewString("b"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("list() with too many args should error")
	}
	res = zipCall("list", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("list() with int should error")
	}
}

// TestZip_Extra2_Extract_ArgumentValidation tests zip.extract().
func TestZip_Extra2_Extract_ArgumentValidation(t *testing.T) {
	res := zipCall("extract")
	if res.Type() != objects.ErrorType {
		t.Fatalf("extract() with no args should error")
	}
	res = zipCall("extract", objects.NewString("a.zip"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extract() with 1 arg should error")
	}
	res = zipCall("extract", objects.NewString("a.zip"), objects.NewString("out"), objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extract() with 3 args should error")
	}
	res = zipCall("extract", objects.NewInt(123), objects.NewString("out"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extract() with int zip should error")
	}
	res = zipCall("extract", objects.NewString("a.zip"), objects.NewInt(456))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extract() with int dest should error")
	}
}

// TestZip_Extra2_WriterOperations tests create, addFile, addDir, addBytes, close.
func TestZip_Extra2_WriterOperations(t *testing.T) {
	// create requires a path argument
	tmpDir := t.TempDir()
	zipPath := tmpDir + "/test.zip"
	writer := zipCall("create", objects.NewString(zipPath))
	if writer.Type() == objects.ErrorType {
		t.Fatalf("create() returned error: %s", writer.Inspect())
	}
	// With no arguments should error
	res := zipCall("create")
	if res.Type() != objects.ErrorType {
		t.Fatalf("create() with no args should error")
	}
	// With wrong type should error
	res = zipCall("create", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("create() with int should error")
	}

	// addFile validation
	res = zipCall("addFile")
	if res.Type() != objects.ErrorType {
		t.Fatalf("addFile() with no args should error")
	}
	res = zipCall("addFile", writer)
	if res.Type() != objects.ErrorType {
		t.Fatalf("addFile() with 1 arg should error")
	}
	res = zipCall("addFile", writer, objects.NewString("src"), objects.NewString("dest"), objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("addFile() with 4 args should error")
	}
	res = zipCall("addFile", writer, objects.NewInt(123), objects.NewString("dest"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("addFile() with int src should error")
	}
	res = zipCall("addFile", writer, objects.NewString("src"), objects.NewInt(456))
	if res.Type() != objects.ErrorType {
		t.Fatalf("addFile() with int dest should error")
	}

	// addDir validation
	res = zipCall("addDir")
	if res.Type() != objects.ErrorType {
		t.Fatalf("addDir() with no args should error")
	}
	res = zipCall("addDir", writer)
	if res.Type() != objects.ErrorType {
		t.Fatalf("addDir() with 1 arg should error")
	}
	// addBytes validation
	res = zipCall("addBytes")
	if res.Type() != objects.ErrorType {
		t.Fatalf("addBytes() with no args should error")
	}
	res = zipCall("addBytes", writer)
	if res.Type() != objects.ErrorType {
		t.Fatalf("addBytes() with 1 arg should error")
	}
	res = zipCall("addBytes", writer, objects.NewString("name"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("addBytes() with 2 args should error")
	}
	res = zipCall("addBytes", writer, objects.NewString("name"), objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("addBytes() with int data should error")
	}

	// close validation
	res = zipCall("close")
	if res.Type() != objects.ErrorType {
		t.Fatalf("close() with no args should error")
	}
	res = zipCall("close", writer)
	if res.Type() != objects.NullType {
		t.Fatalf("close() should return null, got %s", res.Type())
	}
	res = zipCall("close", writer, objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("close() with extra args should error")
	}
}

// TestZip_Extra2_DecodeFilename_IfExists tests decodeFilename if present.
func TestZip_Extra2_DecodeFilename_IfExists(t *testing.T) {
	mod := Get("zip")
	if mod == nil {
		t.Skip("zip module not found")
	}
	if _, exists := mod.Exports["decodeFilename"]; !exists {
		t.Skip("decodeFilename not present")
	}
	res := zipCall("decodeFilename")
	if res.Type() != objects.ErrorType {
		t.Fatalf("decodeFilename() with no args should error")
	}
	res = zipCall("decodeFilename", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("decodeFilename() with int should error")
	}
	res = zipCall("decodeFilename", objects.NewString("test.txt"))
	_ = res
}
