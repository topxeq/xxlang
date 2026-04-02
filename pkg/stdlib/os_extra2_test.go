// pkg/stdlib/os_extra2_test.go
package stdlib

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// osCall invokes a builtin from the os module.
func osCall(name string, args ...objects.Object) objects.Object {
	mod := Get("os")
	if mod == nil {
		return &objects.Error{Message: "os module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestOs_Extra2_Init tests that os module registers all exports.
func TestOs_Extra2_Init(t *testing.T) {
	mod := Get("os")
	if mod == nil {
		t.Skip("os module not found")
	}
	expected := []string{
		"join", "base", "dir", "ext", "abs", "clean", "isAbs",
		"glob", "split", "relative", "volumeName",
		"walkDir", "walk",
		"symlink", "readlink", "isLink", "lstat", "stat",
		"size", "isDir", "isFile",
		"listDir", "listDirFull",
		"exec", "shell",
		"hostname", "platform", "arch", "cpus", "home", "temp",
		"chmod", "rename", "copy",
		"tempFile", "tempDir",
		"userInfo",
		"getConfigObj", "getConfigStr", "setConfigStr",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in os module", name)
		}
	}
}

// TestOs_Extra2_PathFunctions_ArgumentValidation tests path functions.
func TestOs_Extra2_PathFunctions_ArgumentValidation(t *testing.T) {
	// join: needs at least 2 args
	res := osCall("join")
	if res.Type() != objects.ErrorType {
		t.Fatalf("join() with no args should error")
	}
	res = osCall("join", objects.NewString("a"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("join() with 1 arg should error")
	}
	// Valid join
	res = osCall("join", objects.NewString("a"), objects.NewString("b"), objects.NewString("c"))
	if res.Type() != objects.StringType {
		t.Fatalf("join() should return string")
	}

	// base: exactly 1 arg
	res = osCall("base")
	if res.Type() != objects.ErrorType {
		t.Fatalf("base() with no args should error")
	}
	res = osCall("base", objects.NewString("a"), objects.NewString("b"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("base() with too many args should error")
	}
	res = osCall("base", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("base() with int should error")
	}
	res = osCall("base", objects.NewString("/a/b/c"))
	if res.Type() != objects.StringType {
		t.Fatalf("base() should return string")
	}

	// dir: similar validation
	res = osCall("dir", objects.NewString("/a/b/c"))
	if res.Type() != objects.StringType {
		t.Fatalf("dir() should return string")
	}

	// ext
	res = osCall("ext", objects.NewString("file.txt"))
	if res.Type() != objects.StringType {
		t.Fatalf("ext() should return string")
	}

	// abs
	res = osCall("abs", objects.NewString("/a/b/c"))
	if res.Type() != objects.StringType {
		t.Fatalf("abs() should return string, got %s", res.Type())
	}
}

// TestOs_Extra2_PathFunctions_DirectoryFunctions tests directory related helpers.
func TestOs_Extra2_PathFunctions_DirectoryFunctions(t *testing.T) {
	curDir := objects.NewString(".")
	// listDirFull
	res := osCall("listDirFull", curDir)
	if res.Type() != objects.ArrayType {
		t.Fatalf("listDirFull should return array, got %s", res.Type())
	}
	// walkDir
	res = osCall("walkDir", curDir)
	if res.Type() != objects.ArrayType {
		t.Fatalf("walkDir should return array, got %s", res.Type())
	}
	// walk
	res = osCall("walk", curDir)
	if res.Type() != objects.ArrayType {
		t.Fatalf("walk should return array, got %s", res.Type())
	}
}

// TestOs_Extra2_FileOperations tests rename, copy (with temp files).
func TestOs_Extra2_FileOperations(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	// Write initial content using Go directly to create a file
	os.WriteFile(file1, []byte("test content"), 0644)

	// Test copy (using string paths)
	res := osCall("copy", objects.NewString(file1), objects.NewString(file2))
	if res.Type() != objects.NullType {
		t.Fatalf("copy failed: %s", res.Inspect())
	}
	// Verify copy worked
	if _, err := os.Stat(file2); err != nil {
		t.Fatalf("copy did not create destination file")
	}
	// rename
	res = osCall("rename", objects.NewString(file1), objects.NewString(file2+".renamed"))
	if res.Type() != objects.NullType {
		t.Fatalf("rename failed: %s", res.Inspect())
	}
}

// TestOs_Extra2_TempFiles tests tempFile and tempDir.
func TestOs_Extra2_TempFiles(t *testing.T) {
	// tempFile
	res := osCall("tempFile")
	if res.Type() != objects.StringType {
		t.Fatalf("tempFile should return string")
	}
	// tempFile with pattern
	res = osCall("tempFile", objects.NewString("xxlang-test-*.tmp"))
	if res.Type() != objects.StringType {
		t.Fatalf("tempFile with pattern should return string")
	}
	// tempDir
	res = osCall("tempDir")
	if res.Type() != objects.StringType {
		t.Fatalf("tempDir should return string")
	}
	// tempDir with pattern
	res = osCall("tempDir", objects.NewString("xxlang-dir-*"))
	if res.Type() != objects.StringType {
		t.Fatalf("tempDir with pattern should return string")
	}
}

// TestOs_Extra2_SystemInfo tests hostname, platform, arch, cpus, home, temp.
func TestOs_Extra2_SystemInfo(t *testing.T) {
	// hostname
	res := osCall("hostname")
	if res.Type() != objects.StringType {
		t.Fatalf("hostname should return string, got %s", res.Type())
	}
	// platform
	res = osCall("platform")
	if res.Type() != objects.StringType {
		t.Fatalf("platform should return string, got %s", res.Type())
	}
	// arch
	res = osCall("arch")
	if res.Type() != objects.StringType {
		t.Fatalf("arch should return string, got %s", res.Type())
	}
	// cpus
	res = osCall("cpus")
	if res.Type() != objects.IntType {
		t.Fatalf("cpus should return int, got %s", res.Type())
	}
	// home
	res = osCall("home")
	if res.Type() != objects.StringType {
		t.Fatalf("home should return string, got %s", res.Type())
	}
	// temp
	res = osCall("temp")
	if res.Type() != objects.StringType {
		t.Fatalf("temp should return string, got %s", res.Type())
	}
}

// TestOs_Extra2_Exec_Shell tests exec and shell command execution.
func TestOs_Extra2_Exec_Shell(t *testing.T) {
	// exec: simple echo (platform-specific)
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	} else {
		cmd = "echo hello"
	}
	res := osCall("exec", objects.NewString(cmd))
	if res.Type() != objects.ArrayType {
		t.Fatalf("exec should return array, got %s", res.Type())
	}
	arr := res.(*objects.Array)
	if len(arr.Elements) != 3 {
		t.Fatalf("exec should return [output, code, error], got %d elements", len(arr.Elements))
	}
	// shell should also return array
	res = osCall("shell", objects.NewString(cmd))
	if res.Type() != objects.ArrayType {
		t.Fatalf("shell should return array, got %s", res.Type())
	}
}

// TestOs_Extra2_ConfigFunctions tests getConfigObj, getConfigStr, setConfigStr.
func TestOs_Extra2_ConfigFunctions(t *testing.T) {
	// getConfigObj
	res := osCall("getConfigObj")
	if res.Type() != objects.MapType {
		t.Fatalf("getConfigObj should return map, got %s", res.Type())
	}
	// getConfigStr: non-existent config should return null
	res = osCall("getConfigStr", objects.NewString("nonexistent"))
	if res.Type() != objects.NullType {
		t.Fatalf("getConfigStr nonexistent should return null, got %s", res.Type())
	}
	// setConfigStr: writes to user's .xxl dir; just check it doesn't error outright
	res = osCall("setConfigStr", objects.NewString("testkey"), objects.NewString("testvalue"))
	if res.Type() != objects.NullType {
		t.Fatalf("setConfigStr should return null, got %s (message: %s)", res.Type(), res.Inspect())
	}
}

// TestOs_Extra2_SymlinkFunctions tests symlink, readlink, isLink (may require permissions).
func TestOs_Extra2_SymlinkFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")
	// Create target file
	os.WriteFile(target, []byte("test"), 0644)
	// symlink
	res := osCall("symlink", objects.NewString(target), objects.NewString(link))
	if res.Type() != objects.NullType {
		t.Fatalf("symlink failed: %s", res.Inspect())
	}
	// readlink
	res = osCall("readlink", objects.NewString(link))
	if res.Type() != objects.StringType {
		t.Fatalf("readlink should return string, got %s", res.Type())
	}
	// isLink
	res = osCall("isLink", objects.NewString(link))
	if res.Type() != objects.BoolType {
		t.Fatalf("isLink should return bool, got %s", res.Type())
	}
}
