// pkg/objects/file_upload.go
// File upload object types for server mode.
// Provides file upload handling capabilities for Xxlang server mode.

package objects

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileUpload represents an uploaded file from an HTTP request.
// It wraps the multipart.FileHeader and provides methods for accessing
// file information and saving the file to disk.
type FileUpload struct {
	// Header is the underlying multipart.FileHeader
	Header *multipart.FileHeader
	// Members holds dynamically accessed members for script use
	Members map[string]Object
	// tempFile stores the temporary file path if the file has been saved temporarily
	tempFile string
	// size stores the file size
	size int64
}

// Type returns the object type.
func (f *FileUpload) Type() ObjectType { return FileUploadType }

// TypeTag returns the type tag for fast type checking.
func (f *FileUpload) TypeTag() TypeTag { return TagFileUpload }

// Inspect returns a string representation of the file upload.
func (f *FileUpload) Inspect() string {
	if f.Header == nil {
		return "[file_upload nil]"
	}
	return fmt.Sprintf("[file_upload %s size=%d]", f.Header.Filename, f.Header.Size)
}

// ToBool converts the file upload to a boolean (true if file exists).
func (f *FileUpload) ToBool() *Bool {
	return &Bool{Value: f.Header != nil}
}

// HashKey returns a hash key for the file upload.
func (f *FileUpload) HashKey() HashKey {
	return HashKey{Type: FileUploadType, Value: uint64(f.size)}
}

// GetMember returns a member by name for script access.
func (f *FileUpload) GetMember(name string) Object {
	if f.Members != nil {
		if obj, ok := f.Members[name]; ok {
			return obj
		}
	}

	if f.Header == nil {
		return NULL
	}

	switch name {
	case "filename":
		return NewString(f.Header.Filename)
	case "size":
		return NewInt(f.Header.Size)
	case "contentType", "type":
		return NewString(f.Header.Header.Get("Content-Type"))
	case "header":
		return f.getHeaders()
	case "extension", "ext":
		return NewString(filepath.Ext(f.Header.Filename))
	case "basename":
		return NewString(strings.TrimSuffix(f.Header.Filename, filepath.Ext(f.Header.Filename)))
	}

	return NULL
}

// getHeaders converts file headers to a Map object.
func (f *FileUpload) getHeaders() Object {
	pairs := make(map[HashKey]MapPair)
	for key, values := range f.Header.Header {
		k := NewString(key)
		var v Object
		if len(values) == 1 {
			v = NewString(values[0])
		} else {
			elements := make([]Object, len(values))
			for i, val := range values {
				elements[i] = NewString(val)
			}
			v = NewArray(elements)
		}
		pairs[k.HashKey()] = MapPair{Key: k, Value: v}
	}
	return NewMap(pairs)
}

// Open opens the uploaded file for reading.
// Returns an io.ReadCloser or an error.
func (f *FileUpload) Open() (multipart.File, error) {
	if f.Header == nil {
		return nil, fmt.Errorf("file header is nil")
	}
	return f.Header.Open()
}

// Save saves the uploaded file to the specified path.
// Returns the absolute path of the saved file or an error.
func (f *FileUpload) Save(destPath string) (string, error) {
	if f.Header == nil {
		return "", fmt.Errorf("file header is nil")
	}

	// Open the uploaded file
	src, err := f.Header.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy the file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	absPath, _ := filepath.Abs(destPath)
	return absPath, nil
}

// SaveToDir saves the uploaded file to the specified directory with the original filename.
// If autoRename is true, adds timestamp prefix to avoid name conflicts.
// Returns the full path of the saved file or an error.
func (f *FileUpload) SaveToDir(dir string, autoRename bool) (string, error) {
	if f.Header == nil {
		return "", fmt.Errorf("file header is nil")
	}

	// Create directory if not exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	filename := f.Header.Filename
	if autoRename {
		// Add timestamp prefix
		ext := filepath.Ext(filename)
		basename := strings.TrimSuffix(filename, ext)
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s%s", basename, timestamp, ext)
	}

	destPath := filepath.Join(dir, filename)

	// Check if file exists
	if _, err := os.Stat(destPath); err == nil {
		if !autoRename {
			return "", fmt.Errorf("file already exists: %s", destPath)
		}
		// Add random suffix
		ext := filepath.Ext(filename)
		basename := strings.TrimSuffix(filename, ext)
		filename = fmt.Sprintf("%s_%d%s", basename, time.Now().UnixNano(), ext)
		destPath = filepath.Join(dir, filename)
	}

	return f.Save(destPath)
}

// ReadAll reads the entire content of the uploaded file.
// Returns the file content as a byte slice or an error.
func (f *FileUpload) ReadAll() ([]byte, error) {
	if f.Header == nil {
		return nil, fmt.Errorf("file header is nil")
	}

	src, err := f.Header.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	return io.ReadAll(src)
}

// ReadAsString reads the entire content of the uploaded file as a string.
func (f *FileUpload) ReadAsString() (string, error) {
	data, err := f.ReadAll()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HashSHA256 calculates the SHA256 hash of the uploaded file.
func (f *FileUpload) HashSHA256() (string, error) {
	data, err := f.ReadAll()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// NewFileUpload creates a new FileUpload object from a multipart.FileHeader.
func NewFileUpload(header *multipart.FileHeader) *FileUpload {
	size := int64(0)
	if header != nil {
		size = header.Size
	}
	return &FileUpload{
		Header:  header,
		Members: make(map[string]Object),
		size:    size,
	}
}

// FileUploadResult represents the result of a file upload operation.
type FileUploadResult struct {
	// Success indicates whether the operation was successful
	Success bool
	// Message contains the result message or error description
	Message string
	// FilePath is the path where the file was saved
	FilePath string
	// OriginalName is the original filename
	OriginalName string
	// Size is the file size in bytes
	Size int64
}

// Type returns the object type.
func (r *FileUploadResult) Type() ObjectType { return FileUploadResultType }

// TypeTag returns the type tag for fast type checking.
func (r *FileUploadResult) TypeTag() TypeTag { return TagFileUploadResult }

// Inspect returns a string representation of the upload result.
func (r *FileUploadResult) Inspect() string {
	if r.Success {
		return fmt.Sprintf("[upload_result success=%v path=%s]", r.Success, r.FilePath)
	}
	return fmt.Sprintf("[upload_result success=%v message=%s]", r.Success, r.Message)
}

// ToBool converts the upload result to a boolean.
func (r *FileUploadResult) ToBool() *Bool { return &Bool{Value: r.Success} }

// HashKey returns a hash key for the upload result.
func (r *FileUploadResult) HashKey() HashKey {
	return HashKey{Type: FileUploadResultType, Value: 0}
}

// GetMember returns a member by name for script access.
func (r *FileUploadResult) GetMember(name string) Object {
	switch name {
	case "success":
		return &Bool{Value: r.Success}
	case "message":
		return NewString(r.Message)
	case "filePath", "path":
		return NewString(r.FilePath)
	case "originalName", "name":
		return NewString(r.OriginalName)
	case "size":
		return NewInt(r.Size)
	}
	return NULL
}

// NewFileUploadResult creates a new FileUploadResult object.
func NewFileUploadResult(success bool, message, filePath, originalName string, size int64) *FileUploadResult {
	return &FileUploadResult{
		Success:      success,
		Message:      message,
		FilePath:     filePath,
		OriginalName: originalName,
		Size:         size,
	}
}

// FileUploadConfig holds configuration for file upload handling.
type FileUploadConfig struct {
	// MaxSize is the maximum file size in bytes (0 = unlimited)
	MaxSize int64
	// AllowedExtensions is a list of allowed file extensions (empty = all allowed)
	AllowedExtensions []string
	// UploadDir is the default directory for file uploads
	UploadDir string
	// AutoRename indicates whether to automatically rename files on conflict
	AutoRename bool
	// AllowedMimeTypes is a list of allowed MIME types (empty = all allowed)
	AllowedMimeTypes []string
}

// DefaultFileUploadConfig returns the default file upload configuration.
func DefaultFileUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		MaxSize:           10 * 1024 * 1024, // 10MB default
		AllowedExtensions: []string{},
		UploadDir:         "./uploads",
		AutoRename:        true,
		AllowedMimeTypes:  []string{},
	}
}

// Validate validates a file upload against the configuration.
func (c *FileUploadConfig) Validate(file *FileUpload) error {
	if file.Header == nil {
		return fmt.Errorf("file is nil")
	}

	// Check file size
	if c.MaxSize > 0 && file.Header.Size > c.MaxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", file.Header.Size, c.MaxSize)
	}

	// Check file extension
	if len(c.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(file.Header.Filename))
		allowed := false
		for _, ae := range c.AllowedExtensions {
			if strings.ToLower(ae) == ext || strings.ToLower("."+ae) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension %s is not allowed", ext)
		}
	}

	// Check MIME type
	if len(c.AllowedMimeTypes) > 0 {
		mimeType := file.Header.Header.Get("Content-Type")
		allowed := false
		for _, am := range c.AllowedMimeTypes {
			if strings.ToLower(am) == strings.ToLower(mimeType) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("MIME type %s is not allowed", mimeType)
		}
	}

	return nil
}

// SafePath validates and returns a safe file path.
// It prevents directory traversal attacks by checking for ".." and ensuring
// the path is within the allowed directory.
func SafePath(baseDir, filename string) (string, error) {
	// Clean the filename to remove any path separators
	filename = filepath.Base(filename)

	// Remove any null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Build the full path
	fullPath := filepath.Join(baseDir, filename)

	// Get absolute paths
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for base directory: %v", err)
	}

	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for file: %v", err)
	}

	// Ensure the file path is within the base directory
	if !strings.HasPrefix(absFull, absBase) {
		return "", fmt.Errorf("file path is outside the allowed directory")
	}

	return absFull, nil
}

// ParseMultipartForm parses a multipart form from an HTTP request.
// It returns a map of form values and a map of file uploads.
func ParseMultipartForm(req *http.Request, maxMemory int64) (map[string][]string, map[string][]*FileUpload, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request is nil")
	}

	// Parse multipart form
	if err := req.ParseMultipartForm(maxMemory); err != nil {
		return nil, nil, fmt.Errorf("failed to parse multipart form: %v", err)
	}

	// Extract form values
	formValues := make(map[string][]string)
	for key, values := range req.MultipartForm.Value {
		formValues[key] = values
	}

	// Extract files
	files := make(map[string][]*FileUpload)
	for key, fileHeaders := range req.MultipartForm.File {
		var uploads []*FileUpload
		for _, fh := range fileHeaders {
			uploads = append(uploads, NewFileUpload(fh))
		}
		files[key] = uploads
	}

	return formValues, files, nil
}
