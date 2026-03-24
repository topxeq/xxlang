// pkg/stdlib/zip_test.go
// Tests for ZIP file utilities with UTF-8/Chinese filename support.
package stdlib

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/topxeq/xxlang/pkg/objects"
)

// callZipFunc calls a function from the zip module
func callZipFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("zip")
	if mod == nil {
		panic("zip module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// TestZipUTF8Filenames tests ZIP operations with UTF-8 encoded filenames
func TestZipUTF8Filenames(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "test_utf8.zip")
	destDir := filepath.Join(tempDir, "extracted")

	// Test filenames with Chinese characters
	testEntries := map[string]string{
		"中文文件.txt":           "Chinese filename content",
		"日本語ファイル.txt":        "Japanese filename content",
		"한글파일.txt":            "Korean filename content",
		"emoji_😀_file.txt":    "Emoji filename content",
		"regular_file.txt":    "Regular filename content",
		"子目录/nested_中文.txt": "Nested Chinese filename",
	}

	// Create ZIP file using createFromMap
	t.Run("createFromMap with UTF-8 filenames", func(t *testing.T) {
		// Create entries map
		entries := &objects.Map{
			Pairs: make(map[objects.HashKey]objects.MapPair),
		}
		for name, content := range testEntries {
			key := objects.NewString(name)
			entries.Pairs[key.HashKey()] = objects.MapPair{
				Key:   key,
				Value: String(content),
			}
		}

		result := callZipFunc("createFromMap", String(zipPath), entries)
		_, isError := result.(*objects.Error)
		assert.False(t, isError, "createFromMap should not return error")

		// Verify file was created
		_, err := os.Stat(zipPath)
		assert.NoError(t, err, "ZIP file should be created")
	})

	// Test list function
	t.Run("list UTF-8 filenames", func(t *testing.T) {
		result := callZipFunc("list", String(zipPath))
		arr, ok := result.(*objects.Array)
		require.True(t, ok, "list should return array")

		// Check that we got all entries
		assert.GreaterOrEqual(t, len(arr.Elements), len(testEntries))

		// Verify Chinese filenames are preserved
		foundNames := make(map[string]bool)
		for _, elem := range arr.Elements {
			entryMap, ok := elem.(*objects.Map)
			if !ok {
				continue
			}
			if namePair, ok := entryMap.Pairs[objects.NewString("name").HashKey()]; ok {
				if name, ok := namePair.Value.(*objects.String); ok {
					foundNames[name.Value] = true
				}
			}
		}

		for name := range testEntries {
			assert.True(t, foundNames[name], "Entry '%s' should be found", name)
		}
	})

	// Test listNames function
	t.Run("listNames UTF-8 filenames", func(t *testing.T) {
		result := callZipFunc("listNames", String(zipPath))
		arr, ok := result.(*objects.Array)
		require.True(t, ok, "listNames should return array")

		foundNames := make(map[string]bool)
		for _, elem := range arr.Elements {
			if name, ok := elem.(*objects.String); ok {
				foundNames[name.Value] = true
			}
		}

		for name := range testEntries {
			assert.True(t, foundNames[name], "Entry '%s' should be in list", name)
		}
	})

	// Test extract function
	t.Run("extract UTF-8 filenames", func(t *testing.T) {
		result := callZipFunc("extract", String(zipPath), String(destDir))
		_, isError := result.(*objects.Error)
		assert.False(t, isError, "extract should not return error")

		// Verify files were extracted with correct names
		for name, expectedContent := range testEntries {
			extractedPath := filepath.Join(destDir, name)
			content, err := os.ReadFile(extractedPath)
			if err != nil {
				t.Errorf("Failed to read extracted file '%s': %v", name, err)
				continue
			}
			assert.Equal(t, expectedContent, string(content), "Content of '%s' should match", name)
		}
	})

	// Test readEntry function
	t.Run("readEntry UTF-8 filenames", func(t *testing.T) {
		for name, expectedContent := range testEntries {
			result := callZipFunc("readEntry", String(zipPath), String(name))
			content, ok := result.(*objects.String)
			require.True(t, ok, "readEntry should return string for '%s'", name)
			assert.Equal(t, expectedContent, content.Value, "Content should match for '%s'", name)
		}
	})

	// Test hasEntry function
	t.Run("hasEntry UTF-8 filenames", func(t *testing.T) {
		for name := range testEntries {
			result := callZipFunc("hasEntry", String(zipPath), String(name))
			b, ok := result.(*objects.Bool)
			require.True(t, ok, "hasEntry should return bool")
			assert.True(t, bool(b.Value), "Entry '%s' should exist", name)
		}

		// Test non-existent entry
		result := callZipFunc("hasEntry", String(zipPath), String("不存在的文件.txt"))
		b, ok := result.(*objects.Bool)
		require.True(t, ok, "hasEntry should return bool")
		assert.False(t, bool(b.Value), "Non-existent entry should return false")
	})
}

// TestZipHandleOperations tests the handle-based ZIP operations
func TestZipHandleOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "handle_test.zip")

	t.Run("create and close handle", func(t *testing.T) {
		result := callZipFunc("create", String(zipPath))
		handle, ok := result.(*objects.String)
		require.True(t, ok, "create should return handle string")
		assert.Contains(t, handle.Value, "zip:")

		// Close the handle
		result = callZipFunc("close", handle)
		_, isError := result.(*objects.Error)
		assert.False(t, isError, "close should not return error")
	})

	t.Run("addString with Chinese content", func(t *testing.T) {
		// Create new handle
		result := callZipFunc("create", String(zipPath))
		handle, ok := result.(*objects.String)
		require.True(t, ok)

		// Add files with Chinese names and content
		entries := []struct {
			name    string
			content string
		}{
			{"测试文件.txt", "这是测试内容，包含中文字符。"},
			{"数据/报告.csv", "姓名,年龄\n张三,25\n李四,30"},
			{"文档/说明.txt", "这是一个说明文档。"},
		}

		for _, entry := range entries {
			result := callZipFunc("addString", handle, String(entry.name), String(entry.content))
			_, isError := result.(*objects.Error)
			assert.False(t, isError, "addString should not return error for '%s'", entry.name)
		}

		// Close handle
		callZipFunc("close", handle)

		// Verify content
		for _, entry := range entries {
			result := callZipFunc("readEntry", String(zipPath), String(entry.name))
			content, ok := result.(*objects.String)
			require.True(t, ok, "readEntry should return string")
			assert.Equal(t, entry.content, content.Value)
		}
	})
}

// TestZipGBKDecoding tests GBK filename decoding
func TestZipGBKDecoding(t *testing.T) {
	// Test the gbkToUTF8 function directly
	t.Run("gbkToUTF8 conversion", func(t *testing.T) {
		// Test that valid UTF-8 is returned as-is
		utf8Str := "这是UTF-8编码的中文"
		result := gbkToUTF8(utf8Str)
		assert.Equal(t, utf8Str, result)

		// Test the decodeFilename function
		assert.Equal(t, "测试", decodeFilename("测试", "utf8"))
		assert.Equal(t, "测试", decodeFilename("测试", "gbk"))
	})
}

// TestZipInfoFunctions tests information retrieval functions
func TestZipInfoFunctions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "info_test.zip")

	// Create a test ZIP
	entries := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("file1.txt").HashKey(): {
				Key:   objects.NewString("file1.txt"),
				Value: String("content1"),
			},
			objects.NewString("file2.txt").HashKey(): {
				Key:   objects.NewString("file2.txt"),
				Value: String("content2"),
			},
			objects.NewString("中文文件.txt").HashKey(): {
				Key:   objects.NewString("中文文件.txt"),
				Value: String("中文内容"),
			},
		},
	}

	callZipFunc("createFromMap", String(zipPath), entries)

	t.Run("count", func(t *testing.T) {
		result := callZipFunc("count", String(zipPath))
		count, ok := result.(*objects.Int)
		require.True(t, ok, "count should return int")
		assert.Equal(t, int64(3), count.Value)
	})

	t.Run("getInfo", func(t *testing.T) {
		result := callZipFunc("getInfo", String(zipPath))
		info, ok := result.(*objects.Map)
		require.True(t, ok, "getInfo should return map")

		// Check fileCount
		if fcPair, ok := info.Pairs[objects.NewString("fileCount").HashKey()]; ok {
			fc, ok := fcPair.Value.(*objects.Int)
			require.True(t, ok)
			assert.Equal(t, int64(3), fc.Value)
		}
	})

	t.Run("isValid", func(t *testing.T) {
		result := callZipFunc("isValid", String(zipPath))
		valid, ok := result.(*objects.Bool)
		require.True(t, ok, "isValid should return bool")
		assert.True(t, bool(valid.Value))

		// Test with non-existent file
		result = callZipFunc("isValid", String(filepath.Join(tempDir, "nonexistent.zip")))
		valid, ok = result.(*objects.Bool)
		require.True(t, ok)
		assert.False(t, bool(valid.Value))
	})
}

// TestZipRemoveRename tests entry removal and renaming
func TestZipRemoveRename(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "modify_test.zip")

	// Create initial ZIP
	entries := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("file1.txt").HashKey(): {
				Key:   objects.NewString("file1.txt"),
				Value: String("content1"),
			},
			objects.NewString("中文文件.txt").HashKey(): {
				Key:   objects.NewString("中文文件.txt"),
				Value: String("中文内容"),
			},
		},
	}
	callZipFunc("createFromMap", String(zipPath), entries)

	t.Run("removeEntry", func(t *testing.T) {
		result := callZipFunc("removeEntry", String(zipPath), String("file1.txt"))
		_, isError := result.(*objects.Error)
		assert.False(t, isError, "removeEntry should not return error")

		// Verify entry was removed
		result = callZipFunc("hasEntry", String(zipPath), String("file1.txt"))
		hasEntry, ok := result.(*objects.Bool)
		require.True(t, ok)
		assert.False(t, bool(hasEntry.Value), "Entry should be removed")

		// Chinese entry should still exist
		result = callZipFunc("hasEntry", String(zipPath), String("中文文件.txt"))
		hasEntry, ok = result.(*objects.Bool)
		require.True(t, ok)
		assert.True(t, bool(hasEntry.Value), "Chinese entry should still exist")
	})

	t.Run("renameEntry", func(t *testing.T) {
		result := callZipFunc("renameEntry", String(zipPath), String("中文文件.txt"), String("重命名文件.txt"))
		_, isError := result.(*objects.Error)
		assert.False(t, isError, "renameEntry should not return error")

		// Verify old name doesn't exist
		result = callZipFunc("hasEntry", String(zipPath), String("中文文件.txt"))
		hasEntry, ok := result.(*objects.Bool)
		require.True(t, ok)
		assert.False(t, bool(hasEntry.Value), "Old entry should not exist")

		// Verify new name exists
		result = callZipFunc("hasEntry", String(zipPath), String("重命名文件.txt"))
		hasEntry, ok = result.(*objects.Bool)
		require.True(t, ok)
		assert.True(t, bool(hasEntry.Value), "New entry should exist")
	})
}

// TestZipExtractByPattern tests pattern-based extraction
func TestZipExtractByPattern(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "pattern_test.zip")
	destDir := filepath.Join(tempDir, "pattern_extracted")

	// Create test ZIP
	entries := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("file1.txt").HashKey(): {
				Key:   objects.NewString("file1.txt"),
				Value: String("content1"),
			},
			objects.NewString("file2.log").HashKey(): {
				Key:   objects.NewString("file2.log"),
				Value: String("log content"),
			},
			objects.NewString("数据.csv").HashKey(): {
				Key:   objects.NewString("数据.csv"),
				Value: String("csv,data"),
			},
			objects.NewString("文档.txt").HashKey(): {
				Key:   objects.NewString("文档.txt"),
				Value: String("文档内容"),
			},
		},
	}
	callZipFunc("createFromMap", String(zipPath), entries)

	t.Run("extract *.txt files", func(t *testing.T) {
		result := callZipFunc("extractByPattern", String(zipPath), String(destDir), String("*.txt"))
		count, ok := result.(*objects.Int)
		require.True(t, ok, "extractByPattern should return int count")
		assert.Equal(t, int64(2), count.Value, "Should extract 2 .txt files")

		// Verify extracted files
		_, err := os.Stat(filepath.Join(destDir, "file1.txt"))
		assert.NoError(t, err, "file1.txt should be extracted")
		_, err = os.Stat(filepath.Join(destDir, "文档.txt"))
		assert.NoError(t, err, "文档.txt should be extracted")
	})
}

// TestZipReadAll tests reading all entries from a ZIP
func TestZipReadAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "readall_test.zip")

	// Create test ZIP
	entries := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("test1.txt").HashKey(): {
				Key:   objects.NewString("test1.txt"),
				Value: String("content1"),
			},
			objects.NewString("测试.txt").HashKey(): {
				Key:   objects.NewString("测试.txt"),
				Value: String("测试内容"),
			},
		},
	}
	callZipFunc("createFromMap", String(zipPath), entries)

	t.Run("readAll", func(t *testing.T) {
		result := callZipFunc("readAll", String(zipPath))
		contentMap, ok := result.(*objects.Map)
		require.True(t, ok, "readAll should return map")

		// Check content
		if pair, ok := contentMap.Pairs[objects.NewString("test1.txt").HashKey()]; ok {
			content, ok := pair.Value.(*objects.String)
			require.True(t, ok)
			assert.Equal(t, "content1", content.Value)
		}

		if pair, ok := contentMap.Pairs[objects.NewString("测试.txt").HashKey()]; ok {
			content, ok := pair.Value.(*objects.String)
			require.True(t, ok)
			assert.Equal(t, "测试内容", content.Value)
		}
	})
}

// TestZipNativeGoIntegration tests that ZIP files created by Go stdlib work with our functions
func TestZipNativeGoIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xxlang_zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "native_test.zip")

	// Create ZIP using Go stdlib
	t.Run("create with Go stdlib", func(t *testing.T) {
		file, err := os.Create(zipPath)
		require.NoError(t, err)
		defer file.Close()

		writer := zip.NewWriter(file)
		defer writer.Close()

		// Add files with UTF-8 names
		files := map[string]string{
			"english.txt":      "English content",
			"中文.txt":          "中文内容",
			"日本語.txt":        "日本語内容",
			"子目录/nested.txt": "Nested content",
		}

		for name, content := range files {
			w, err := writer.Create(name)
			require.NoError(t, err)
			_, err = w.Write([]byte(content))
			require.NoError(t, err)
		}
	})

	// Test that our functions can read them
	t.Run("read with Xxlang functions", func(t *testing.T) {
		// Test listNames
		result := callZipFunc("listNames", String(zipPath))
		names, ok := result.(*objects.Array)
		require.True(t, ok)
		assert.Equal(t, 4, len(names.Elements))

		// Test readEntry for each file
		expectedFiles := map[string]string{
			"english.txt":      "English content",
			"中文.txt":          "中文内容",
			"日本語.txt":        "日本語内容",
			"子目录/nested.txt": "Nested content",
		}

		for name, expectedContent := range expectedFiles {
			result := callZipFunc("readEntry", String(zipPath), String(name))
			content, ok := result.(*objects.String)
			require.True(t, ok, "readEntry should return string for '%s'", name)
			assert.Equal(t, expectedContent, content.Value, "Content should match for '%s'", name)
		}
	})
}
