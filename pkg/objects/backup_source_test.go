// pkg/objects/backup_source_test.go
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

// MockSSHClient for testing RemoteSource
type MockSSHClient struct {
	files    map[string][]byte
	fileInfo map[string]map[string]interface{}
}

func (m *MockSSHClient) Exec(cmd string) (string, error) {
	return "", nil
}

func (m *MockSSHClient) ReadFile(path string) (string, error) {
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	if content, ok := m.files[path]; ok {
		return string(content), nil
	}
	return "", os.ErrNotExist
}

func (m *MockSSHClient) WriteFile(path, content string) error {
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	m.files[path] = []byte(content)
	return nil
}

func (m *MockSSHClient) ListDir(path string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *MockSSHClient) Stat(path string) (map[string]interface{}, error) {
	if m.fileInfo == nil {
		m.fileInfo = make(map[string]map[string]interface{})
	}
	if info, ok := m.fileInfo[path]; ok {
		return info, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockSSHClient) MkdirAll(path string) error {
	return nil
}

func (m *MockSSHClient) Remove(path string) error {
	return nil
}

func (m *MockSSHClient) Exists(path string) bool {
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	_, ok := m.files[path]
	return ok
}

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

// ============================================================
// RemoteSource Tests
// ============================================================

func TestRemoteSourceNew(t *testing.T) {
	mockClient := &MockSSHClient{}
	src := NewRemoteSource(mockClient, "/remote/path")
	if src == nil {
		t.Fatal("expected RemoteSource instance")
	}
	if src.BasePath != "/remote/path" {
		t.Errorf("expected BasePath='/remote/path', got '%s'", src.BasePath)
	}
}

func TestRemoteSourceGetBasePath(t *testing.T) {
	mockClient := &MockSSHClient{}
	src := NewRemoteSource(mockClient, "/remote/backup")
	if src.GetBasePath() != "/remote/backup" {
		t.Errorf("expected GetBasePath='/remote/backup', got '%s'", src.GetBasePath())
	}
}

func TestRemoteSourceReadFile(t *testing.T) {
	mockClient := &MockSSHClient{}
	mockClient.files = map[string][]byte{
		"/remote/path/test.txt": []byte("remote content"),
	}
	src := NewRemoteSource(mockClient, "/remote/path")

	content, err := src.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "remote content" {
		t.Errorf("expected 'remote content', got '%s'", content)
	}
}

func TestRemoteSourceWriteFile(t *testing.T) {
	mockClient := &MockSSHClient{}
	src := NewRemoteSource(mockClient, "/remote/path")

	err := src.WriteFile("new.txt", []byte("new remote content"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file was written in mock
	content, ok := mockClient.files["/remote/path/new.txt"]
	if !ok {
		t.Fatal("file not written to mock")
	}
	if string(content) != "new remote content" {
		t.Errorf("expected 'new remote content', got '%s'", content)
	}
}

func TestRemoteSourceExists(t *testing.T) {
	mockClient := &MockSSHClient{}
	mockClient.files = map[string][]byte{
		"/remote/path/exists.txt": []byte("content"),
	}
	src := NewRemoteSource(mockClient, "/remote/path")

	if !src.Exists("exists.txt") {
		t.Error("expected Exists=true for existing file")
	}
	if src.Exists("notexists.txt") {
		t.Error("expected Exists=false for non-existing file")
	}
}

func TestRemoteSourceMkdirAll(t *testing.T) {
	mockClient := &MockSSHClient{}
	src := NewRemoteSource(mockClient, "/remote/path")

	err := src.MkdirAll("subdir/nested")
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
}

func TestRemoteSourceDeleteFile(t *testing.T) {
	mockClient := &MockSSHClient{}
	src := NewRemoteSource(mockClient, "/remote/path")

	err := src.DeleteFile("delete.txt")
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
}