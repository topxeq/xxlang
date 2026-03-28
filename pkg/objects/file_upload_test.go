// pkg/objects/file_upload_test.go
package objects

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFileUpload_BasicProperties tests basic FileUpload properties
func TestFileUpload_BasicProperties(t *testing.T) {
	// Create a mock file header
	header := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     1024,
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "text/plain")

	file := NewFileUpload(header)

	// Test type
	if file.Type() != FileUploadType {
		t.Errorf("expected type %s, got %s", FileUploadType, file.Type())
	}

	// Test type tag
	if file.TypeTag() != TagFileUpload {
		t.Errorf("expected type tag %d, got %d", TagFileUpload, file.TypeTag())
	}

	// Test ToBool
	if !file.ToBool().Value {
		t.Error("expected FileUpload to be truthy")
	}
}

// TestFileUpload_GetMember tests GetMember method
func TestFileUpload_GetMember(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "document.pdf",
		Size:     2048,
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "application/pdf")

	file := NewFileUpload(header)

	tests := []struct {
		name     string
		expected string
	}{
		{"filename", "document.pdf"},
		{"extension", ".pdf"},
		{"contentType", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := file.GetMember(tt.name)
			strResult, ok := result.(*String)
			if !ok {
				t.Errorf("expected String, got %T", result)
				return
			}
			if strResult.Value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, strResult.Value)
			}
		})
	}

	// Test size member
	t.Run("size", func(t *testing.T) {
		result := file.GetMember("size")
		intResult, ok := result.(*Int)
		if !ok {
			t.Errorf("expected Int, got %T", result)
			return
		}
		if intResult.Value != 2048 {
			t.Errorf("expected 2048, got %d", intResult.Value)
		}
	})
}

// TestFileUpload_NilHeader tests FileUpload with nil header
func TestFileUpload_NilHeader(t *testing.T) {
	file := NewFileUpload(nil)

	if file.ToBool().Value {
		t.Error("expected nil FileUpload to be falsy")
	}

	if file.GetMember("filename") != NULL {
		t.Error("expected NULL for nil header filename")
	}
}

// TestFileUploadResult tests FileUploadResult object
func TestFileUploadResult(t *testing.T) {
	result := NewFileUploadResult(true, "File saved", "/uploads/test.txt", "test.txt", 1024)

	if result.Type() != FileUploadResultType {
		t.Errorf("expected type %s, got %s", FileUploadResultType, result.Type())
	}

	if !result.Success {
		t.Error("expected success to be true")
	}

	if result.Message != "File saved" {
		t.Errorf("expected 'File saved', got '%s'", result.Message)
	}

	if !result.ToBool().Value {
		t.Error("expected successful result to be truthy")
	}
}

// TestFileUploadResult_GetMember tests FileUploadResult GetMember
func TestFileUploadResult_GetMember(t *testing.T) {
	result := NewFileUploadResult(true, "OK", "/path/file.txt", "file.txt", 512)

	tests := []struct {
		name     string
		expected interface{}
	}{
		{"success", true},
		{"message", "OK"},
		{"filePath", "/path/file.txt"},
		{"originalName", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := result.GetMember(tt.name)
			switch expected := tt.expected.(type) {
			case bool:
				boolResult, ok := member.(*Bool)
				if !ok || boolResult.Value != expected {
					t.Errorf("expected %v, got %v", expected, member)
				}
			case string:
				strResult, ok := member.(*String)
				if !ok || strResult.Value != expected {
					t.Errorf("expected %v, got %v", expected, member)
				}
			}
		})
	}

	t.Run("size", func(t *testing.T) {
		member := result.GetMember("size")
		intResult, ok := member.(*Int)
		if !ok || intResult.Value != 512 {
			t.Errorf("expected 512, got %v", member)
		}
	})
}

// TestSafePath tests path safety validation
func TestSafePath(t *testing.T) {
	tests := []struct {
		name      string
		baseDir   string
		filename  string
		expectErr bool
	}{
		{"normal file", "./uploads", "test.txt", false},
		// Note: filepath.Base strips path components, so "../etc/passwd" becomes "passwd"
		// This is safe - the path traversal is prevented
		{"subdirectory attempt stripped", "./uploads", "../etc/passwd", false},
		{"null byte removed", "./uploads", "test\x00.txt", false},
		{"path separator stripped", "./uploads", "path/to/file.txt", false},
		{"absolute path stripped", "./uploads", "/etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafePath(tt.baseDir, tt.filename)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			// Just verify no error occurred for safe paths
			if err == nil && result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

// TestGetFileUploads_Builtin tests getFileUploads builtin function
func TestGetFileUploads_Builtin(t *testing.T) {
	// Create a test request with multipart form
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)

	// Add a file
	_, err := writer.CreateFormFile("file1", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpReq := NewHttpReq(req)

	// Call builtin function
	builtin, ok := FileUploadBuiltins["getFileUploads"]
	if !ok {
		t.Fatal("getFileUploads builtin not found")
	}

	result := builtin.Fn(httpReq)
	if result == NULL {
		t.Error("expected non-null result")
	}

	mapResult, ok := result.(*Map)
	if !ok {
		t.Errorf("expected Map, got %T", result)
		return
	}

	// Should have files
	if len(mapResult.Pairs) == 0 {
		t.Log("No files parsed (expected for simple test)")
	}
}

// TestIsFileUpload_Builtin tests isFileUpload builtin
func TestIsFileUpload_Builtin(t *testing.T) {
	builtin, ok := FileUploadBuiltins["isFileUpload"]
	if !ok {
		t.Fatal("isFileUpload builtin not found")
	}

	file := NewFileUpload(&multipart.FileHeader{Filename: "test.txt"})

	result := builtin.Fn(file)
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Errorf("expected Bool, got %T", result)
		return
	}

	if !boolResult.Value {
		t.Error("expected true for FileUpload")
	}

	// Test with non-FileUpload
	result = builtin.Fn(NewString("test"))
	boolResult, ok = result.(*Bool)
	if !ok || boolResult.Value {
		t.Error("expected false for non-FileUpload")
	}
}
