// pkg/objects/builtin_compress_test.go
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinCompressData(t *testing.T) {
	fn, ok := Builtins["compressData"]
	if !ok {
		t.Fatal("compressData builtin not found")
	}

	result := fn.Fn(NewString("hello world"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) == 0 {
		t.Error("expected non-empty compressed data")
	}

	arr := &Array{Elements: []Object{NewInt(72), NewInt(101), NewInt(108), NewInt(108), NewInt(111)}}
	result = fn.Fn(arr)
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) == 0 {
		t.Error("expected non-empty compressed data from array")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string/array arg")
	}
}

func TestBuiltinUncompressData(t *testing.T) {
	fn, ok := Builtins["uncompressData"]
	if !ok {
		t.Fatal("uncompressData builtin not found")
	}

	compressFn, _ := Builtins["compressData"]
	compressed := compressFn.Fn(NewString("hello world"))
	compressedStr, ok := compressed.(*String)
	if !ok {
		t.Fatal("failed to compress test data")
	}

	result := fn.Fn(compressedStr)
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}

	result = fn.Fn(NewString("invalid hex!"))
	if !isError(result) {
		t.Error("expected error for invalid hex string")
	}
}

func TestBuiltinCompressStr(t *testing.T) {
	fn, ok := Builtins["compressStr"]
	if !ok {
		t.Fatal("compressStr builtin not found")
	}

	result := fn.Fn(NewString("test string for compression"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(strResult.Value) == 0 {
		t.Error("expected non-empty compressed string")
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

func TestBuiltinUncompressStr(t *testing.T) {
	fn, ok := Builtins["uncompressStr"]
	if !ok {
		t.Fatal("uncompressStr builtin not found")
	}

	compressFn, _ := Builtins["compressStr"]
	compressed := compressFn.Fn(NewString("test data"))
	compressedStr, ok := compressed.(*String)
	if !ok {
		t.Fatal("failed to compress test data")
	}

	result := fn.Fn(compressedStr)
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "test data" {
		t.Errorf("expected 'test data', got '%s'", strResult.Value)
	}
}

func TestBuiltinZipPath(t *testing.T) {
	fn, ok := Builtins["zipPath"]
	if !ok {
		t.Fatal("zipPath builtin not found")
	}

	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "test_zip_source.txt")
	zipFile := filepath.Join(tempDir, "test_output.zip")

	os.WriteFile(testFile, []byte("test content"), 0644)
	defer os.Remove(testFile)
	defer os.Remove(zipFile)

	result := fn.Fn(NewString(testFile), NewString(zipFile))
	if isError(result) {
		t.Errorf("unexpected error: %v", result)
	}

	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
		t.Error("zip file was not created")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetFileListInZip(t *testing.T) {
	fn, ok := Builtins["getFileListInZip"]
	if !ok {
		t.Fatal("getFileListInZip builtin not found")
	}

	zipFn, _ := Builtins["zipPath"]
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "test_list_source.txt")
	zipFile := filepath.Join(tempDir, "test_list.zip")

	os.WriteFile(testFile, []byte("test content"), 0644)
	defer os.Remove(testFile)
	defer os.Remove(zipFile)

	zipFn.Fn(NewString(testFile), NewString(zipFile))

	result := fn.Fn(NewString(zipFile))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) == 0 {
		t.Error("expected at least one file in zip")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = fn.Fn(NewString("nonexistent.zip"))
	if !isError(result) {
		t.Error("expected error for nonexistent zip file")
	}
}

func TestBuiltinLoadBytesInZip(t *testing.T) {
	fn, ok := Builtins["loadBytesInZip"]
	if !ok {
		t.Fatal("loadBytesInZip builtin not found")
	}

	zipFn, _ := Builtins["zipPath"]
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "test_load_source.txt")
	zipFile := filepath.Join(tempDir, "test_load.zip")

	os.WriteFile(testFile, []byte("test content"), 0644)
	defer os.Remove(testFile)
	defer os.Remove(zipFile)

	result := zipFn.Fn(NewString(testFile), NewString(zipFile))
	if isError(result) {
		t.Fatalf("zipPath failed: %v", result)
	}

	// Get the file list to see what the internal name is
	listFn, _ := Builtins["getFileListInZip"]
	listResult := listFn.Fn(NewString(zipFile))
	if listArr, ok := listResult.(*Array); ok && len(listArr.Elements) > 0 {
		internalName := listArr.Elements[0].(*String).Value
		t.Logf("Internal file name: %s", internalName)

		result = fn.Fn(NewString(zipFile), NewString(internalName))
		strResult, ok := result.(*String)
		if !ok {
			t.Fatalf("expected String, got %T: %v", result, result)
		}
		if strResult.Value != "test content" {
			t.Errorf("expected 'test content', got '%s'", strResult.Value)
		}
	} else {
		t.Skip("Could not get file list from zip")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinAddFileToZip(t *testing.T) {
	fn, ok := Builtins["addFileToZip"]
	if !ok {
		t.Fatal("addFileToZip builtin not found")
	}

	zipFn, _ := Builtins["zipPath"]
	tempDir := os.TempDir()
	testFile1 := filepath.Join(tempDir, "test_add1.txt")
	testFile2 := filepath.Join(tempDir, "test_add2.txt")
	zipFile := filepath.Join(tempDir, "test_add.zip")

	os.WriteFile(testFile1, []byte("content 1"), 0644)
	os.WriteFile(testFile2, []byte("content 2"), 0644)
	defer os.Remove(testFile1)
	defer os.Remove(testFile2)
	defer os.Remove(zipFile)

	zipFn.Fn(NewString(testFile1), NewString(zipFile))

	// addFileToZip needs 3 args: zipPath, filePath, innerPath
	result := fn.Fn(NewString(zipFile), NewString(testFile2), NewString("test_add2.txt"))
	if isError(result) {
		t.Errorf("unexpected error: %v", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinUnzipToPath(t *testing.T) {
	fn, ok := Builtins["unzipToPath"]
	if !ok {
		t.Fatal("unzipToPath builtin not found")
	}

	zipFn, _ := Builtins["zipPath"]
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "test_unzip_source.txt")
	zipFile := filepath.Join(tempDir, "test_unzip.zip")
	extractDir := filepath.Join(tempDir, "extracted")

	os.WriteFile(testFile, []byte("test content"), 0644)
	defer os.Remove(testFile)
	defer os.Remove(zipFile)
	defer os.RemoveAll(extractDir)

	zipFn.Fn(NewString(testFile), NewString(zipFile))

	result := fn.Fn(NewString(zipFile), NewString(extractDir))
	if isError(result) {
		t.Errorf("unexpected error: %v", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinZipPaths(t *testing.T) {
	fn, ok := Builtins["zipPaths"]
	if !ok {
		t.Fatal("zipPaths builtin not found")
	}

	tempDir := os.TempDir()
	testFile1 := filepath.Join(tempDir, "test_paths1.txt")
	testFile2 := filepath.Join(tempDir, "test_paths2.txt")
	zipFile := filepath.Join(tempDir, "test_paths.zip")

	os.WriteFile(testFile1, []byte("content 1"), 0644)
	os.WriteFile(testFile2, []byte("content 2"), 0644)
	defer os.Remove(testFile1)
	defer os.Remove(testFile2)
	defer os.Remove(zipFile)

	paths := &Array{Elements: []Object{NewString(testFile1), NewString(testFile2)}}
	result := fn.Fn(paths, NewString(zipFile))
	if isError(result) {
		t.Errorf("unexpected error: %v", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestEncodeDecodeHex(t *testing.T) {
	testCases := []string{
		"hello world",
		"",
		"test123",
		"\x00\xff\x10",
	}

	for _, tc := range testCases {
		encoded := encodeHex([]byte(tc))
		decoded, err := decodeHex(encoded)
		if err != nil {
			t.Errorf("decodeHex error for '%s': %v", tc, err)
			continue
		}
		if string(decoded) != tc {
			t.Errorf("expected '%s', got '%s'", tc, string(decoded))
		}
	}
}

func TestHexCharToByte(t *testing.T) {
	testCases := []struct {
		char     byte
		expected byte
	}{
		{'0', 0},
		{'9', 9},
		{'a', 10},
		{'f', 15},
		{'A', 10},
		{'F', 15},
	}

	for _, tc := range testCases {
		result := hexCharToByte(tc.char)
		if result != tc.expected {
			t.Errorf("hexCharToByte('%c') = %d, expected %d", tc.char, result, tc.expected)
		}
	}
}
