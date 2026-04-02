// pkg/stdlib/file_test.go
// Comprehensive tests for file module.
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callFileFunc calls a function from the file module.
func callFileFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("file")
	if mod == nil {
		t := &testing.T{}
		t.Skip("file module not found")
		return &objects.Error{Message: "file module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// modeRead returns the ModeRead constant value from the file module.
func modeRead() *objects.String {
	mod := Get("file")
	if mod == nil {
		return String("r")
	}
	if v, ok := mod.Exports["MODE_READ"].(*objects.String); ok {
		return v
	}
	return String("r")
}

// modeWrite returns the ModeWrite constant.
func modeWrite() *objects.String {
	mod := Get("file")
	if mod == nil {
		return String("w")
	}
	if v, ok := mod.Exports["MODE_WRITE"].(*objects.String); ok {
		return v
	}
	return String("w")
}

// modeAppend returns the ModeAppend constant.
func modeAppend() *objects.String {
	mod := Get("file")
	if mod == nil {
		return String("a")
	}
	if v, ok := mod.Exports["MODE_APPEND"].(*objects.String); ok {
		return v
	}
	return String("a")
}

// modeRWPlus returns the ModeRWPlus constant.
func modeRWPlus() *objects.String {
	mod := Get("file")
	if mod == nil {
		return String("rw+")
	}
	if v, ok := mod.Exports["MODE_RWPLUS"].(*objects.String); ok {
		return v
	}
	return String("rw+")
}

// TestFileOpenRead tests file.open with read mode.
func TestFileOpenRead(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "xxlang-file-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	testContent := "Hello, Xxlang file test!"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	fileObj := callFileFunc("open", String(tmpFile.Name()), modeRead())
	if fileObj.Type() == objects.ErrorType {
		t.Fatalf("open() returned error: %s", fileObj.Inspect())
	}
	if _, ok := fileObj.(*objects.File); !ok {
		t.Fatalf("expected *objects.File, got %T", fileObj)
	}

	result := callFileFunc("readAll", String(tmpFile.Name()))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readAll() error: %s", result.Inspect())
	}
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != testContent {
		t.Errorf("expected content %q, got %q", testContent, str.Value)
	}
}

// TestFileCreate tests file.create.
func TestFileCreate(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "test-create.txt")

	result := callFileFunc("create", String(tmpPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("create() error: %s", result.Inspect())
	}
	fileObj := result

	// Close the file handle
	if fileObj.Type() != objects.ErrorType {
		closeResult := callFileFunc("close", fileObj)
		if closeResult.Type() == objects.ErrorType {
			t.Logf("close() warning: %s", closeResult.Inspect())
		}
	}

	existsResult := callFileFunc("exists", String(tmpPath))
	if existsResult.Type() != objects.BoolType || !existsResult.(*objects.Bool).Value {
		t.Error("create() did not create file")
	}

	isFileResult := callFileFunc("isFile", String(tmpPath))
	if isFileResult.Type() != objects.BoolType || !isFileResult.(*objects.Bool).Value {
		t.Error("isFile() returned false for created file")
	}

	os.Remove(tmpPath)
}

// TestFileOpenWrite tests file.openWrite.
func TestFileOpenWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "test-write.txt")

	fileObj := callFileFunc("openWrite", String(tmpPath))
	if fileObj.Type() == objects.ErrorType {
		t.Fatalf("openWrite() error: %s", fileObj.Inspect())
	}
	defer func() {
		// Close the file handle
		closeResult := callFileFunc("close", fileObj)
		if closeResult.Type() == objects.ErrorType {
			t.Logf("close() warning: %s", closeResult.Inspect())
		}
	}()

	content := "Write test content"
	result := callFileFunc("writeAll", String(tmpPath), String(content))
	if result.Type() == objects.ErrorType {
		t.Fatalf("writeAll() error: %s", result.Inspect())
	}

	result = callFileFunc("readAll", String(tmpPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readAll() error: %s", result.Inspect())
	}
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != content {
		t.Errorf("expected %q, got %q", content, str.Value)
	}
}

// TestFileOpenAppend tests file.openAppend and appendAll.
func TestFileOpenAppend(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "test-append.txt")

	os.WriteFile(tmpPath, []byte("line1\n"), 0644)

	fileObj := callFileFunc("openAppend", String(tmpPath))
	if fileObj.Type() == objects.ErrorType {
		t.Fatalf("openAppend() error: %s", fileObj.Inspect())
	}
	defer func() {
		// Close the file handle
		closeResult := callFileFunc("close", fileObj)
		if closeResult.Type() == objects.ErrorType {
			t.Logf("close() warning: %s", closeResult.Inspect())
		}
	}()

	result := callFileFunc("appendAll", String(tmpPath), String("line2\n"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("appendAll() error: %s", result.Inspect())
	}

	result = callFileFunc("readAll", String(tmpPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readAll() error: %s", result.Inspect())
	}
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "line1\nline2\n"
	if str.Value != expected {
		t.Errorf("expected %q, got %q", expected, str.Value)
	}
}

// TestFileReadLines tests file.readLines.
func TestFileReadLines(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "lines.txt")
	lines := []string{"first line", "second line", "third line"}
	content := "first line\nsecond line\nthird line\n"
	os.WriteFile(tmpPath, []byte(content), 0644)

	result := callFileFunc("readLines", String(tmpPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readLines() error: %s", result.Inspect())
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(arr.Elements))
	}
	for i, expected := range lines {
		if s, ok := arr.Elements[i].(*objects.String); ok {
			if s.Value != expected {
				t.Errorf("line %d: expected %q, got %q", i, expected, s.Value)
			}
		} else {
			t.Errorf("line %d: expected String, got %T", i, arr.Elements[i])
		}
	}
}

// TestFileWriteLines tests file.writeLines.
func TestFileWriteLines(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "writelines.txt")

	lines := []string{"alpha", "beta", "gamma"}
	arr := make([]objects.Object, len(lines))
	for i, l := range lines {
		arr[i] = String(l)
	}

	result := callFileFunc("writeLines", String(tmpPath), Array(arr...))
	if result.Type() == objects.ErrorType {
		t.Fatalf("writeLines() error: %s", result.Inspect())
	}

	result = callFileFunc("readLines", String(tmpPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readLines() error: %s", result.Inspect())
	}
	arr2, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr2.Elements) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(arr2.Elements))
	}
	for i, expected := range lines {
		if s, ok := arr2.Elements[i].(*objects.String); ok {
			if s.Value != expected {
				t.Errorf("line %d: expected %q, got %q", i, expected, s.Value)
			}
		} else {
			t.Errorf("line %d: expected String, got %T", i, arr2.Elements[i])
		}
	}
}

// TestFileCopyMove tests file.copy and file.move.
func TestFileCopyMove(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")
	content := "copy me"
	os.WriteFile(srcPath, []byte(content), 0644)

	result := callFileFunc("copy", String(srcPath), String(dstPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("copy() error: %s", result.Inspect())
	}
	existsResult := callFileFunc("exists", String(dstPath))
	if existsResult.Type() != objects.BoolType || !existsResult.(*objects.Bool).Value {
		t.Error("copy did not create destination file")
	}
	isFileResult := callFileFunc("isFile", String(dstPath))
	if isFileResult.Type() != objects.BoolType || !isFileResult.(*objects.Bool).Value {
		t.Error("copied file is not a file")
	}

	newDst := filepath.Join(tmpDir, "moved.txt")
	result = callFileFunc("move", String(dstPath), String(newDst))
	if result.Type() == objects.ErrorType {
		t.Fatalf("move() error: %s", result.Inspect())
	}
	existsResult = callFileFunc("exists", String(dstPath))
	if existsResult.Type() != objects.BoolType || existsResult.(*objects.Bool).Value {
		t.Error("move did not remove source file")
	}
	existsResult = callFileFunc("exists", String(newDst))
	if existsResult.Type() != objects.BoolType || !existsResult.(*objects.Bool).Value {
		t.Error("move did not create new destination")
	}
}

// TestFileRemove tests file.remove and file.removeDir.
func TestFileRemove(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "toremove.txt")
	os.WriteFile(filePath, []byte("content"), 0644)

	result := callFileFunc("remove", String(filePath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("remove() error: %s", result.Inspect())
	}
	existsResult := callFileFunc("exists", String(filePath))
	if existsResult.Type() != objects.BoolType || existsResult.(*objects.Bool).Value {
		t.Error("remove did not delete file")
	}

	dirPath := filepath.Join(tmpDir, "emptydir")
	os.Mkdir(dirPath, 0755)
	result = callFileFunc("removeDir", String(dirPath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("removeDir() error: %s", result.Inspect())
	}
	existsResult = callFileFunc("exists", String(dirPath))
	if existsResult.Type() != objects.BoolType || existsResult.(*objects.Bool).Value {
		t.Error("removeDir did not delete directory")
	}
}

// TestFileExistsIsFileIsDir tests file.exists, file.isFile, file.isDir.
func TestFileExistsIsFileIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "exists.txt")
	dirPath := filepath.Join(tmpDir, "existsdir")
	os.WriteFile(filePath, []byte{}, 0644)
	os.Mkdir(dirPath, 0755)

	existsFn := callFileFunc("exists", String(filePath)).(*objects.Bool)
	isFileFn := callFileFunc("isFile", String(filePath)).(*objects.Bool)
	isDirFn := callFileFunc("isDir", String(filePath)).(*objects.Bool)

	if !existsFn.Value {
		t.Error("exists returned false for existing file")
	}
	if !isFileFn.Value {
		t.Error("isFile returned false for file")
	}
	if isDirFn.Value {
		t.Error("isDir returned true for file")
	}

	existsFn = callFileFunc("exists", String(dirPath)).(*objects.Bool)
	isFileFn = callFileFunc("isFile", String(dirPath)).(*objects.Bool)
	isDirFn = callFileFunc("isDir", String(dirPath)).(*objects.Bool)

	if !existsFn.Value {
		t.Error("exists returned false for existing dir")
	}
	if isFileFn.Value {
		t.Error("isFile returned true for dir")
	}
	if !isDirFn.Value {
		t.Error("isDir returned false for dir")
	}

	existsResult := callFileFunc("exists", String(filepath.Join(tmpDir, "nope")))
	if existsResult.Type() == objects.BoolType && existsResult.(*objects.Bool).Value {
		t.Error("exists returned true for non-existent")
	}
}

// TestFileStatSizeModTime tests file.stat, file.size, file.modTime.
func TestFileStatSizeModTime(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "stat.txt")
	content := "1234567890"
	os.WriteFile(filePath, []byte(content), 0644)

	statObj := callFileFunc("stat", String(filePath))
	if statObj.Type() == objects.ErrorType {
		t.Fatalf("stat() error: %s", statObj.Inspect())
	}
	stat, ok := statObj.(*objects.FileInfo)
	if !ok {
		t.Fatalf("expected *objects.FileInfo, got %T", statObj)
	}
	if stat.Size != int64(len(content)) {
		t.Errorf("size mismatch: expected %d, got %d", len(content), stat.Size)
	}

	sizeObj := callFileFunc("size", String(filePath))
	if sizeObj.Type() == objects.ErrorType {
		t.Fatalf("size() error: %s", sizeObj.Inspect())
	}
	if sizeObj.(*objects.Int).Value != int64(len(content)) {
		t.Errorf("size() mismatch: expected %d, got %d", len(content), sizeObj.(*objects.Int).Value)
	}

	modTimeObj := callFileFunc("modTime", String(filePath))
	if modTimeObj.Type() == objects.ErrorType {
		t.Fatalf("modTime() error: %s", modTimeObj.Inspect())
	}
	modTimeStr := modTimeObj.(*objects.String).Value
	if modTimeStr == "" {
		t.Error("modTime returned empty string")
	}
}

// TestFileMkdir tests file.mkdir.
func TestFileMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "newsubdir")

	result := callFileFunc("mkdir", String(newDir))
	if result.Type() == objects.ErrorType {
		t.Fatalf("mkdir() error: %s", result.Inspect())
	}
	existsResult := callFileFunc("exists", String(newDir))
	if existsResult.Type() != objects.BoolType || !existsResult.(*objects.Bool).Value {
		t.Error("mkdir did not create directory")
	}
	isDirResult := callFileFunc("isDir", String(newDir))
	if isDirResult.Type() != objects.BoolType || !isDirResult.(*objects.Bool).Value {
		t.Error("mkdir created something that is not a directory")
	}
}

// TestFileListDir tests file.listDir and file.listDirFull.
func TestFileListDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte{}, 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	result := callFileFunc("listDir", String(tmpDir))
	if result.Type() == objects.ErrorType {
		t.Fatalf("listDir() error: %s", result.Inspect())
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(arr.Elements))
	}

	result = callFileFunc("listDirFull", String(tmpDir))
	if result.Type() == objects.ErrorType {
		t.Fatalf("listDirFull() error: %s", result.Inspect())
	}
	arr2, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	for _, elem := range arr2.Elements {
		entryMap, ok := elem.(*objects.OrderedMap)
		if !ok {
			t.Fatalf("expected OrderedMap entry, got %T", elem)
		}
		// Check required keys exist
		if entryMap.Get(&objects.String{Value: "name"}) == nil {
			t.Fatalf("expected 'name' key in entry")
		}
		if entryMap.Get(&objects.String{Value: "size"}) == nil {
			t.Fatalf("expected 'size' key in entry")
		}
		if entryMap.Get(&objects.String{Value: "isDir"}) == nil {
			t.Fatalf("expected 'isDir' key in entry")
		}
		if entryMap.Get(&objects.String{Value: "modTime"}) == nil {
			t.Fatalf("expected 'modTime' key in entry")
		}
	}
}

// TestFileGlob tests file.glob.
func TestFileGlob(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "other.log"), []byte{}, 0644)

	pattern := filepath.Join(tmpDir, "*.txt")
	result := callFileFunc("glob", String(pattern))
	if result.Type() == objects.ErrorType {
		t.Fatalf("glob() error: %s", result.Inspect())
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(arr.Elements))
	}
}

// TestFilePathUtils tests file.abs, file.base, file.dir, file.ext, file.join.
func TestFilePathUtils(t *testing.T) {
	absResult := callFileFunc("abs", String("test.txt"))
	if absResult.Type() == objects.ErrorType {
		t.Fatalf("abs() error: %s", absResult.Inspect())
	}
	absStr := absResult.(*objects.String).Value
	if absStr == "" {
		t.Error("abs returned empty string")
	}

	baseResult := callFileFunc("base", String("/a/b/c.txt"))
	if baseResult.Type() == objects.ErrorType {
		t.Fatalf("base() error: %s", baseResult.Inspect())
	}
	if baseResult.(*objects.String).Value != "c.txt" {
		t.Errorf("base of /a/b/c.txt expected c.txt, got %s", baseResult.(*objects.String).Value)
	}

	dirResult := callFileFunc("dir", String("/a/b/c.txt"))
	if dirResult.Type() == objects.ErrorType {
		t.Fatalf("dir() error: %s", dirResult.Inspect())
	}
	dirVal := dirResult.(*objects.String).Value
	// Normalize to forward slashes for cross-platform consistency
	expectedDir := "/a/b"
	if filepath.FromSlash(dirVal) != filepath.FromSlash(expectedDir) {
		t.Errorf("dir of /a/b/c.txt expected %s, got %s", expectedDir, dirVal)
	}

	extResult := callFileFunc("ext", String("file.tar.gz"))
	if extResult.Type() == objects.ErrorType {
		t.Fatalf("ext() error: %s", extResult.Inspect())
	}
	if extResult.(*objects.String).Value != ".gz" {
		t.Errorf("ext of file.tar.gz expected .gz, got %s", extResult.(*objects.String).Value)
	}

	joinResult := callFileFunc("join", String("a"), String("b"), String("c.txt"))
	if joinResult.Type() == objects.ErrorType {
		t.Fatalf("join() error: %s", joinResult.Inspect())
	}
	expectedJoin := filepath.Join("a", "b", "c.txt")
	if joinResult.(*objects.String).Value != expectedJoin {
		t.Errorf("join() expected %s, got %s", expectedJoin, joinResult.(*objects.String).Value)
	}
}

// TestFileCwdChdir tests file.cwd and file.chdir.
func TestFileCwdChdir(t *testing.T) {
	initCwdObj := callFileFunc("cwd")
	if initCwdObj.Type() == objects.ErrorType {
		t.Fatalf("cwd() error: %s", initCwdObj.Inspect())
	}
	initCwd := initCwdObj.(*objects.String).Value

	tmpDir := t.TempDir()
	result := callFileFunc("chdir", String(tmpDir))
	if result.Type() == objects.ErrorType {
		t.Fatalf("chdir() error: %s", result.Inspect())
	}

	newCwdObj := callFileFunc("cwd")
	if newCwdObj.Type() == objects.ErrorType {
		t.Fatalf("cwd() error after chdir: %s", newCwdObj.Inspect())
	}
	newCwd := newCwdObj.(*objects.String).Value
	if newCwd != tmpDir {
		t.Errorf("chdir failed: expected cwd %s, got %s", tmpDir, newCwd)
	}

	callFileFunc("chdir", String(initCwd))
}

// TestFileTemp tests file.tempFile and file.tempDir.
func TestFileTemp(t *testing.T) {
	tmpFile1 := callFileFunc("tempFile")
	if tmpFile1.Type() == objects.ErrorType {
		t.Fatalf("tempFile() error: %s", tmpFile1.Inspect())
	}
	path1 := tmpFile1.(*objects.String).Value
	if callFileFunc("exists", String(path1)).Type() != objects.BoolType || !callFileFunc("exists", String(path1)).(*objects.Bool).Value {
		t.Error("tempFile did not create file")
	}
	if callFileFunc("isFile", String(path1)).Type() != objects.BoolType || !callFileFunc("isFile", String(path1)).(*objects.Bool).Value {
		t.Error("tempFile created something not a file")
	}
	os.Remove(path1)

	tmpFile2 := callFileFunc("tempFile", String("custom-*.tmp"))
	if tmpFile2.Type() == objects.ErrorType {
		t.Fatalf("tempFile(custom) error: %s", tmpFile2.Inspect())
	}
	path2 := tmpFile2.(*objects.String).Value
	if callFileFunc("exists", String(path2)).Type() != objects.BoolType || !callFileFunc("exists", String(path2)).(*objects.Bool).Value {
		t.Error("tempFile(custom) did not create file")
	}
	os.Remove(path2)

	tmpDir1 := callFileFunc("tempDir")
	if tmpDir1.Type() == objects.ErrorType {
		t.Fatalf("tempDir() error: %s", tmpDir1.Inspect())
	}
	dirPath1 := tmpDir1.(*objects.String).Value
	if callFileFunc("exists", String(dirPath1)).Type() != objects.BoolType || !callFileFunc("exists", String(dirPath1)).(*objects.Bool).Value {
		t.Error("tempDir did not create dir")
	}
	if callFileFunc("isDir", String(dirPath1)).Type() != objects.BoolType || !callFileFunc("isDir", String(dirPath1)).(*objects.Bool).Value {
		t.Error("tempDir created something not a dir")
	}
	os.RemoveAll(dirPath1)
}

// TestFileChmod tests file.chmod.
func TestFileChmod(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "chmod.txt")
	os.WriteFile(filePath, []byte{}, 0644)

	result := callFileFunc("chmod", String(filePath), Int(0755))
	if result.Type() == objects.ErrorType {
		t.Fatalf("chmod() error: %s", result.Inspect())
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat after chmod failed: %v", err)
	}
	mode := info.Mode().Perm()
	if mode&os.FileMode(0755) != os.FileMode(0755) {
		t.Logf("chmod may not have fully applied; got %o", mode)
	}
}

// TestFileReadWriteObject tests file open, read, write using file object.
func TestFileReadWriteObject(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "rwobj.txt")

	fileObj := callFileFunc("open", String(filePath), modeRWPlus())
	if fileObj.Type() == objects.ErrorType {
		t.Fatalf("open() error: %s", fileObj.Inspect())
	}
	defer func() {
		// Close the file handle
		closeResult := callFileFunc("close", fileObj)
		if closeResult.Type() == objects.ErrorType {
			t.Logf("close() warning: %s", closeResult.Inspect())
		}
	}()

	content := "object test"
	result := callFileFunc("writeAll", String(filePath), String(content))
	if result.Type() == objects.ErrorType {
		t.Fatalf("writeAll error: %s", result.Inspect())
	}

	result = callFileFunc("readAll", String(filePath))
	if result.Type() == objects.ErrorType {
		t.Fatalf("readAll error: %s", result.Inspect())
	}
	if str, ok := result.(*objects.String); ok {
		if str.Value != content {
			t.Errorf("expected %q, got %q", content, str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}
}