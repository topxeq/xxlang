// pkg/objects/builtin_file_test.go
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinFileExists(t *testing.T) {
	fn, ok := Builtins["fileExists"]
	if !ok {
		t.Fatal("fileExists builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_file_exists.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for existing file")
	}

	result = fn.Fn(NewString("/nonexistent/path/file.txt"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for nonexistent file")
	}

	result = fn.Fn(NewString(os.TempDir()))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for directory")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}
}

func TestBuiltinDirExists(t *testing.T) {
	fn, ok := Builtins["dirExists"]
	if !ok {
		t.Fatal("dirExists builtin not found")
	}

	result := fn.Fn(NewString(os.TempDir()))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for existing directory")
	}

	result = fn.Fn(NewString("/nonexistent/path/dir"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for nonexistent directory")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	result = fn.Fn(NewString(tmpFile))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for file")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}
}

func TestBuiltinPathExists(t *testing.T) {
	fn, ok := Builtins["pathExists"]
	if !ok {
		t.Fatal("pathExists builtin not found")
	}

	result := fn.Fn(NewString(os.TempDir()))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for existing directory")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	result = fn.Fn(NewString(tmpFile))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for existing file")
	}

	result = fn.Fn(NewString("/nonexistent/path"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for nonexistent path")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsDir(t *testing.T) {
	fn, ok := Builtins["isDir"]
	if !ok {
		t.Fatal("isDir builtin not found")
	}

	result := fn.Fn(NewString(os.TempDir()))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for directory")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	result = fn.Fn(NewString(tmpFile))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for file")
	}

	result = fn.Fn(NewString("/nonexistent/path"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for nonexistent path")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsFile(t *testing.T) {
	fn, ok := Builtins["isFile"]
	if !ok {
		t.Fatal("isFile builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for file")
	}

	result = fn.Fn(NewString(os.TempDir()))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for directory")
	}

	result = fn.Fn(NewString("/nonexistent/path"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for nonexistent path")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinLoadText(t *testing.T) {
	fn, ok := Builtins["loadText"]
	if !ok {
		t.Fatal("loadText builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_loadtext.txt")
	os.WriteFile(tmpFile, []byte("hello world"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("/nonexistent/file.txt"))
	if !isError(result) {
		t.Error("expected error for nonexistent file")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}
}

func TestBuiltinSaveText(t *testing.T) {
	fn, ok := Builtins["saveText"]
	if !ok {
		t.Fatal("saveText builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_savetext.txt")
	result := fn.Fn(NewString(tmpFile), NewString("test content"))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
	os.Remove(tmpFile)

	result = fn.Fn(NewInt(123), NewString("content"))
	if !isError(result) {
		t.Error("expected error for non-string path")
	}

	result = fn.Fn(NewString(tmpFile), NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string content")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString(tmpFile))
	if !isError(result) {
		t.Error("expected error for single arg")
	}
}

func TestBuiltinJoinPath(t *testing.T) {
	fn, ok := Builtins["joinPath"]
	if !ok {
		t.Fatal("joinPath builtin not found")
	}

	result := fn.Fn(NewString("dir1"), NewString("dir2"), NewString("file.txt"))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(NewString("dir1"))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSaveBytes(t *testing.T) {
	fn, ok := Builtins["saveBytes"]
	if !ok {
		t.Fatal("saveBytes builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_savebytes.bin")
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile), NewBytes([]byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}))
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	// Verify file was written
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", string(content))
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("path"))
	if !isError(result) {
		t.Error("expected error for single arg")
	}

	result = fn.Fn(NewInt(123), NewBytes([]byte{1}))
	if !isError(result) {
		t.Error("expected error for non-string path")
	}

	result = fn.Fn(NewString(tmpFile), NewString("not bytes"))
	if !isError(result) {
		t.Error("expected error for non-bytes content")
	}
}

func TestBuiltinLoadBytes(t *testing.T) {
	fn, ok := Builtins["loadBytes"]
	if !ok {
		t.Fatal("loadBytes builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_loadbytes.bin")
	os.WriteFile(tmpFile, []byte("Hello World"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	b, ok := result.(*Bytes)
	if !ok {
		t.Fatalf("expected Bytes, got %T", result)
	}
	if string(b.Value) != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", string(b.Value))
	}

	// Test with limit
	result = fn.Fn(NewString(tmpFile), NewInt(5))
	b, ok = result.(*Bytes)
	if !ok {
		t.Fatalf("expected Bytes, got %T", result)
	}
	if len(b.Value) != 5 {
		t.Errorf("expected 5 bytes, got %d", len(b.Value))
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string path")
	}

	result = fn.Fn(NewString(tmpFile), NewString("not int"))
	if !isError(result) {
		t.Error("expected error for non-int limit")
	}
}

func TestBuiltinAppendText(t *testing.T) {
	fn, ok := Builtins["appendText"]
	if !ok {
		t.Fatal("appendText builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_appendtext.txt")
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	// First append
	result := fn.Fn(NewString(tmpFile), NewString("Hello"))
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	// Second append
	result = fn.Fn(NewString(tmpFile), NewString(" World"))
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	// Verify content
	content, _ := os.ReadFile(tmpFile)
	if string(content) != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", string(content))
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString(tmpFile))
	if !isError(result) {
		t.Error("expected error for single arg")
	}
}

func TestBuiltinRenameFile(t *testing.T) {
	fn, ok := Builtins["renameFile"]
	if !ok {
		t.Fatal("renameFile builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_rename.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	newFile := filepath.Join(os.TempDir(), "xxlang_test_renamed.txt")
	defer os.Remove(newFile)

	result := fn.Fn(NewString(tmpFile), NewString(newFile))
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	// Verify old file is gone
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("expected old file to be renamed")
	}

	// Verify new file exists
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("expected new file to exist")
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetFileList(t *testing.T) {
	fn, ok := Builtins["getFileList"]
	if !ok {
		t.Fatal("getFileList builtin not found")
	}

	// Test with temp directory
	result := fn.Fn(NewString(os.TempDir()))
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	// Temp dir should have at least some files
	_ = arr

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

func TestBuiltinGetCurDir(t *testing.T) {
	fn, ok := Builtins["getCurDir"]
	if !ok {
		t.Fatal("getCurDir builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty current directory")
	}

	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetHomeDir(t *testing.T) {
	fn, ok := Builtins["getHomeDir"]
	if !ok {
		t.Fatal("getHomeDir builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty home directory")
	}

	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetTempDir(t *testing.T) {
	fn, ok := Builtins["getTempDir"]
	if !ok {
		t.Fatal("getTempDir builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty temp directory")
	}

	result = fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetFileExt(t *testing.T) {
	fn, ok := Builtins["getFileExt"]
	if !ok {
		t.Fatal("getFileExt builtin not found")
	}

	result := fn.Fn(NewString("file.txt"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != ".txt" {
		t.Errorf("expected '.txt', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("file.tar.gz"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != ".gz" {
		t.Errorf("expected '.gz', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString("file"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "" {
		t.Errorf("expected '', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinExtractFileDir(t *testing.T) {
	fn, ok := Builtins["extractFileDir"]
	if !ok {
		t.Fatal("extractFileDir builtin not found")
	}

	result := fn.Fn(NewString("/home/user/file.txt"))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinExtractFileName(t *testing.T) {
	fn, ok := Builtins["extractFileName"]
	if !ok {
		t.Fatal("extractFileName builtin not found")
	}

	result := fn.Fn(NewString("/home/user/file.txt"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "file.txt" {
		t.Errorf("expected 'file.txt', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetFileAbs(t *testing.T) {
	fn, ok := Builtins["getFileAbs"]
	if !ok {
		t.Fatal("getFileAbs builtin not found")
	}

	result := fn.Fn(NewString("."))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRemoveFile(t *testing.T) {
	fn, ok := Builtins["removeFile"]
	if !ok {
		t.Fatal("removeFile builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_remove.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	result := fn.Fn(NewString(tmpFile))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	result = fn.Fn(NewString("/nonexistent/file.txt"))
	if !isError(result) {
		t.Error("expected error for nonexistent file")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinRemoveDir(t *testing.T) {
	fn, ok := Builtins["removeDir"]
	if !ok {
		t.Fatal("removeDir builtin not found")
	}

	tmpDir := filepath.Join(os.TempDir(), "xxlang_test_removedir")
	os.Mkdir(tmpDir, 0755)

	result := fn.Fn(NewString(tmpDir))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	result = fn.Fn(NewString("/nonexistent/dir"))
	if result != NULL {
		t.Errorf("expected NULL for nonexistent dir, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCopyFile(t *testing.T) {
	fn, ok := Builtins["copyFile"]
	if !ok {
		t.Fatal("copyFile builtin not found")
	}

	srcFile := filepath.Join(os.TempDir(), "xxlang_test_src.txt")
	dstFile := filepath.Join(os.TempDir(), "xxlang_test_dst.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)
	defer os.Remove(srcFile)
	defer os.Remove(dstFile)

	result := fn.Fn(NewString(srcFile), NewString(dstFile))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	content, _ := os.ReadFile(dstFile)
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}

	result = fn.Fn(NewInt(123), NewString(dstFile))
	if !isError(result) {
		t.Error("expected error for non-string src")
	}

	result = fn.Fn(NewString(srcFile), NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string dst")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinEnsureMakeDirs(t *testing.T) {
	fn, ok := Builtins["ensureMakeDirs"]
	if !ok {
		t.Fatal("ensureMakeDirs builtin not found")
	}

	tmpDir := filepath.Join(os.TempDir(), "xxlang_test_mkdir", "subdir")
	result := fn.Fn(NewString(tmpDir))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	os.RemoveAll(filepath.Join(os.TempDir(), "xxlang_test_mkdir"))

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetFileSize(t *testing.T) {
	fn, ok := Builtins["getFileSize"]
	if !ok {
		t.Fatal("getFileSize builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_size.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("/nonexistent/file.txt"))
	if !isError(result) {
		t.Error("expected error for nonexistent file")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetFileRel(t *testing.T) {
	fn, ok := Builtins["getFileRel"]
	if !ok {
		t.Fatal("getFileRel builtin not found")
	}

	result := fn.Fn(NewString("/a/b"), NewString("/a/b/c.txt"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "c.txt" {
		t.Errorf("expected 'c.txt', got '%s'", strResult.Value)
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("/a"))
	if !isError(result) {
		t.Error("expected error for single arg")
	}

	result = fn.Fn(NewInt(123), NewString("/a"))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestBuiltinGetFileInfo(t *testing.T) {
	fn, ok := Builtins["getFileInfo"]
	if !ok {
		t.Fatal("getFileInfo builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_info.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	// Check for expected keys
	expectedKeys := []string{"name", "size", "isDir", "modTime", "mode"}
	for _, key := range expectedKeys {
		k := NewString(key).HashKey()
		if _, exists := mapResult.Pairs[k]; !exists {
			t.Errorf("expected key '%s' in result", key)
		}
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

	result = fn.Fn(NewString("/nonexistent/file.txt"))
	if !isError(result) {
		t.Error("expected error for nonexistent file")
	}
}

func TestBuiltinLoadLines(t *testing.T) {
	fn, ok := Builtins["loadLines"]
	if !ok {
		t.Fatal("loadLines builtin not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "xxlang_test_lines.txt")
	os.WriteFile(tmpFile, []byte("line1\nline2\nline3"), 0644)
	defer os.Remove(tmpFile)

	result := fn.Fn(NewString(tmpFile))
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 lines, got %d", len(arr.Elements))
	}

	// Test with limit
	result = fn.Fn(NewString(tmpFile), NewInt(2))
	arr, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 lines with limit, got %d", len(arr.Elements))
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

func TestBuiltinCopyPath(t *testing.T) {
	fn, ok := Builtins["copyPath"]
	if !ok {
		t.Fatal("copyPath builtin not found")
	}

	// Create source file
	srcFile := filepath.Join(os.TempDir(), "xxlang_test_copy_src.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)
	defer os.Remove(srcFile)

	// Create destination path
	dstFile := filepath.Join(os.TempDir(), "xxlang_test_copy_dst.txt")
	defer os.Remove(dstFile)

	result := fn.Fn(NewString(srcFile), NewString(dstFile))
	if result != NULL {
		t.Errorf("expected NULL, got %v", result)
	}

	// Verify copy
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}

	// Test error cases
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString(srcFile))
	if !isError(result) {
		t.Error("expected error for single arg")
	}
}
