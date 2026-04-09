// pkg/objects/backup_source_test.go
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSourceNew(t *testing.T) {
	src := NewLocalSource("/tmp/test")
	if src == nil {
		t.Fatal("expected LocalSource instance")
	}
	if src.BasePath != "/tmp/test" {
		t.Errorf("expected BasePath='/tmp/test', got '%s'", src.BasePath)
	}
}

func TestLocalSourceListFiles(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("content3"), 0644)

	src := NewLocalSource(tmpDir)
	files, err := src.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	// 2 files + 1 subdir + 1 nested file = 4 items
	if len(files) != 4 {
		t.Errorf("expected 4 items (2 files + 1 dir + 1 nested file), got %d", len(files))
	}
}

func TestLocalSourceGetFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	src := NewLocalSource(tmpDir)
	info, err := src.GetFileInfo("test.txt")
	if err != nil {
		t.Fatalf("GetFileInfo failed: %v", err)
	}
	if info.Size != 5 {
		t.Errorf("expected Size=5, got %d", info.Size)
	}
	if info.IsDir {
		t.Error("expected IsDir=false")
	}
}

func TestLocalSourceReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	src := NewLocalSource(tmpDir)
	content, err := src.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", content)
	}
}

func TestLocalSourceWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := NewLocalSource(tmpDir)

	err := src.WriteFile("new.txt", []byte("new content"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file was written
	content, _ := os.ReadFile(filepath.Join(tmpDir, "new.txt"))
	if string(content) != "new content" {
		t.Errorf("expected 'new content', got '%s'", content)
	}
}