// pkg/objects/builtin_file_upload_test.go
// Tests for file upload built-in functions
package objects

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestGetFileUploads(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func() *HttpReq
		wantError   bool
		checkResult func(t *testing.T, result Object)
	}{
		{
			name: "nil request",
			setupReq: func() *HttpReq {
				return &HttpReq{Value: nil}
			},
			wantError: true,
		},
		{
			name: "non-multipart request",
			setupReq: func() *HttpReq {
				req, _ := http.NewRequest("POST", "/test", strings.NewReader("key=value"))
				return &HttpReq{Value: req}
			},
			wantError: false,
		},
		{
			name: "empty multipart form",
			setupReq: func() *HttpReq {
				req, _ := http.NewRequest("POST", "/test", strings.NewReader("--boundary--"))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
				return &HttpReq{Value: req}
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := builtinGetFileUploads(req)

			if tt.wantError {
				if _, ok := result.(*Error); !ok {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if _, ok := result.(*Map); !ok {
				t.Errorf("expected Map, got %s", result.Type())
			}
		})
	}
}

// builtinGetFileUploads retrieves file uploads from HTTP request (exposed for testing)
func builtinGetFileUploads(req *HttpReq) Object {
	if req.Value == nil {
		return newError("http request is nil")
	}

	// Parse multipart form
	if err := req.Value.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(err.Error(), "multipart") {
			return NewMap(make(map[HashKey]MapPair))
		}
		return newError("failed to parse multipart form: %v", err)
	}

	// Build result map
	pairs := make(map[HashKey]MapPair)
	for key, fileHeaders := range req.Value.MultipartForm.File {
		k := NewString(key)
		uploads := make([]Object, len(fileHeaders))
		for i, fh := range fileHeaders {
			uploads[i] = NewFileUpload(fh)
		}
		hashKey := k.HashKey()
		pairs[hashKey] = MapPair{
			Key:   k,
			Value: NewArray(uploads),
		}
	}

	return NewMap(pairs)
}

func TestFileUploadObject(t *testing.T) {
	// Create a mock file header
	header := &multipart.FileHeader{
		Filename: "test.txt",
		Header:   map[string][]string{"Content-Type": {"text/plain"}},
	}

	upload := NewFileUpload(header)

	if upload.Type() != "FILE_UPLOAD" {
		t.Errorf("expected type FILE_UPLOAD, got %s", upload.Type())
	}

	if upload.Inspect() == "" {
		t.Errorf("expected non-empty inspect string")
	}

	if !upload.ToBool().Value {
		t.Errorf("FileUpload should convert to true")
	}
}

// Additional tests for multipart form processing
func TestGetFileUploadsWithFiles(t *testing.T) {
	// Create a multipart form with file upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("upload", "test.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpReq := &HttpReq{Value: req}
	result := builtinGetFileUploads(httpReq)

	mapObj, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	// Check that upload field exists
	key := NewString("upload").HashKey()
	if _, exists := mapObj.Pairs[key]; !exists {
		t.Errorf("expected 'upload' key in result map")
	}
}
