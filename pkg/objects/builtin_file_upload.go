// pkg/objects/builtin_file_upload.go
// Built-in functions for file upload handling in server mode.
package objects

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// FileUploadBuiltins contains all file upload related built-in functions.
// These are only available in server mode.
var FileUploadBuiltins = map[string]*Builtin{
	// getFileUploads retrieves all uploaded files from an HTTP request.
	// Returns a map of field names to arrays of FileUpload objects.
	// Usage: files = getFileUploads(request)
	"getFileUploads": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getFileUploads. got=%d, want=1", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("argument to 'getFileUploads' must be HTTP_REQ, got %s", args[0].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			// Parse multipart form (max 32MB in memory)
			if err := req.Value.ParseMultipartForm(32 << 20); err != nil {
				// Not a multipart form, return empty map
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
				pairs[k.HashKey()] = MapPair{Key: k, Value: NewArray(uploads)}
			}

			return NewMap(pairs)
		},
	},

	// getFileUpload retrieves a specific uploaded file by field name.
	// Returns a FileUpload object or null if not found.
	// Usage: file = getFileUpload(request, "fieldName")
	"getFileUpload": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for getFileUpload. got=%d, want=2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'getFileUpload' must be HTTP_REQ, got %s", args[0].Type())
			}

			fieldName, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'getFileUpload' must be STRING, got %s", args[1].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			// Parse multipart form
			if err := req.Value.ParseMultipartForm(32 << 20); err != nil {
				return NULL
			}

			// Get the file
			file, header, err := req.Value.FormFile(fieldName.Value)
			if err != nil {
				return NULL
			}
			file.Close()

			return NewFileUpload(header)
		},
	},

	// saveFile saves an uploaded file to a specified path.
	// Returns a FileUploadResult object.
	// Usage: result = saveFile(fileUpload, path)
	"saveFile": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for saveFile. got=%d, want=2", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("first argument to 'saveFile' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			path, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'saveFile' must be STRING, got %s", args[1].Type())
			}

			savedPath, err := file.Save(path.Value)
			if err != nil {
				return NewFileUploadResult(false, err.Error(), "", file.Header.Filename, 0)
			}

			return NewFileUploadResult(true, "File saved successfully", savedPath, file.Header.Filename, file.Header.Size)
		},
	},

	// saveFileToDir saves an uploaded file to a directory.
	// Optionally auto-renames to avoid conflicts.
	// Returns a FileUploadResult object.
	// Usage: result = saveFileToDir(fileUpload, dirPath, autoRename)
	"saveFileToDir": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for saveFileToDir. got=%d, want=2 or 3", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("first argument to 'saveFileToDir' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			dirPath, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'saveFileToDir' must be STRING, got %s", args[1].Type())
			}

			autoRename := false
			if len(args) == 3 {
				ar, ok := args[2].(*Bool)
				if !ok {
					return newError("third argument to 'saveFileToDir' must be BOOL, got %s", args[2].Type())
				}
				autoRename = ar.Value
			}

			savedPath, err := file.SaveToDir(dirPath.Value, autoRename)
			if err != nil {
				return NewFileUploadResult(false, err.Error(), "", file.Header.Filename, 0)
			}

			return NewFileUploadResult(true, "File saved successfully", savedPath, file.Header.Filename, file.Header.Size)
		},
	},

	// readFile reads the content of an uploaded file as a string.
	// Usage: content = readFile(fileUpload)
	"readFile": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for readFile. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'readFile' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			content, err := file.ReadAsString()
			if err != nil {
				return newError("failed to read file: %v", err)
			}

			return NewString(content)
		},
	},

	// readFileBytes reads the content of an uploaded file as a Bytes object.
	// Usage: bytes = readFileBytes(fileUpload)
	"readFileBytes": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for readFileBytes. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'readFileBytes' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			data, err := file.ReadAll()
			if err != nil {
				return newError("failed to read file: %v", err)
			}

			return NewBytesBufferFromBytes(data)
		},
	},

	// fileHashSHA256 calculates the SHA256 hash of an uploaded file.
	// Usage: hash = fileHashSHA256(fileUpload)
	"fileHashSHA256": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for fileHashSHA256. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'fileHashSHA256' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			hash, err := file.HashSHA256()
			if err != nil {
				return newError("failed to calculate hash: %v", err)
			}

			return NewString(hash)
		},
	},

	// isFileUpload checks if a value is a FileUpload object.
	// Usage: isFileUpload(value)
	"isFileUpload": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isFileUpload. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*FileUpload)
			return &Bool{Value: ok}
		},
	},

	// isFileUploadResult checks if a value is a FileUploadResult object.
	// Usage: isFileUploadResult(value)
	"isFileUploadResult": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isFileUploadResult. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*FileUploadResult)
			return &Bool{Value: ok}
		},
	},

	// parseMultipartForm parses a multipart form from an HTTP request.
	// Returns a map with 'files' and 'values' keys.
	// Usage: result = parseMultipartForm(request, maxMemory)
	"parseMultipartForm": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for parseMultipartForm. got=%d, want=1 or 2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'parseMultipartForm' must be HTTP_REQ, got %s", args[0].Type())
			}

			maxMemory := int64(32 << 20) // 32MB default
			if len(args) == 2 {
				m, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'parseMultipartForm' must be INT, got %s", args[1].Type())
				}
				maxMemory = m.Value
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			formValues, files, err := ParseMultipartForm(req.Value, maxMemory)
			if err != nil {
				return newError("failed to parse multipart form: %v", err)
			}

			// Build result map
			resultPairs := make(map[HashKey]MapPair)

			// Add values
			valuesPairs := make(map[HashKey]MapPair)
			for key, vals := range formValues {
				k := NewString(key)
				var v Object
				if len(vals) == 1 {
					v = NewString(vals[0])
				} else {
					elements := make([]Object, len(vals))
					for i, val := range vals {
						elements[i] = NewString(val)
					}
					v = NewArray(elements)
				}
				valuesPairs[k.HashKey()] = MapPair{Key: k, Value: v}
			}
			resultPairs[NewString("values").HashKey()] = MapPair{
				Key:   NewString("values"),
				Value: NewMap(valuesPairs),
			}

			// Add files
			filesPairs := make(map[HashKey]MapPair)
			for key, uploads := range files {
				k := NewString(key)
				elements := make([]Object, len(uploads))
				for i, upload := range uploads {
					elements[i] = upload
				}
				filesPairs[k.HashKey()] = MapPair{Key: k, Value: NewArray(elements)}
			}
			resultPairs[NewString("files").HashKey()] = MapPair{
				Key:   NewString("files"),
				Value: NewMap(filesPairs),
			}

			return NewMap(resultPairs)
		},
	},

	// safePath validates and returns a safe file path.
	// Usage: path = safePath(baseDir, filename)
	"safePath": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for safePath. got=%d, want=2", len(args))
			}

			baseDir, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'safePath' must be STRING, got %s", args[0].Type())
			}

			filename, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'safePath' must be STRING, got %s", args[1].Type())
			}

			safe, err := SafePath(baseDir.Value, filename.Value)
			if err != nil {
				return newError("unsafe path: %v", err)
			}

			return NewString(safe)
		},
	},

	// validateFile validates an uploaded file against size and type constraints.
	// Returns true if valid, error message otherwise.
	// Usage: result = validateFile(fileUpload, maxSize, allowedExtensions...)
	"validateFile": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for validateFile. got=%d, want at least 2", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("first argument to 'validateFile' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			maxSize, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'validateFile' must be INT, got %s", args[1].Type())
			}

			// Check file size
			if maxSize.Value > 0 && file.Header.Size > maxSize.Value {
				return &Bool{Value: false}
			}

			// Check extensions if provided
			if len(args) > 2 {
				ext := strings.ToLower(filepath.Ext(file.Header.Filename))
				allowed := false
				for i := 2; i < len(args); i++ {
					allowedExt, ok := args[i].(*String)
					if !ok {
						continue
					}
					checkExt := strings.ToLower(allowedExt.Value)
					if !strings.HasPrefix(checkExt, ".") {
						checkExt = "." + checkExt
					}
					if checkExt == ext {
						allowed = true
						break
					}
				}
				if !allowed {
					return &Bool{Value: false}
				}
			}

			return &Bool{Value: true}
		},
	},

	// getFileExtension gets the extension of an uploaded file.
	// Usage: ext = getFileExtension(fileUpload)
	"getFileExtension": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getFileExtension. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'getFileExtension' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			return NewString(filepath.Ext(file.Header.Filename))
		},
	},

	// getFileName gets the filename of an uploaded file without extension.
	// Usage: name = getFileName(fileUpload)
	"getFileName": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getFileName. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'getFileName' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			name := strings.TrimSuffix(file.Header.Filename, filepath.Ext(file.Header.Filename))
			return NewString(name)
		},
	},

	// getFileSize gets the size of an uploaded file in bytes.
	// Usage: size = getFileSize(fileUpload)
	"getFileSize": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getFileSize. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'getFileSize' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			return NewInt(file.Header.Size)
		},
	},

	// getFileContentType gets the content type of an uploaded file.
	// Usage: contentType = getFileContentType(fileUpload)
	"getFileContentType": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getFileContentType. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'getFileContentType' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			return NewString(file.Header.Header.Get("Content-Type"))
		},
	},

	// openFileUpload opens an uploaded file and returns a reader.
	// The reader can be used to stream large files without loading into memory.
	// Returns a Bytes object (for compatibility) with the file content.
	// Usage: data = openFileUpload(fileUpload)
	"openFileUpload": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for openFileUpload. got=%d, want=1", len(args))
			}

			file, ok := args[0].(*FileUpload)
			if !ok {
				return newError("argument to 'openFileUpload' must be FILE_UPLOAD, got %s", args[0].Type())
			}

			f, err := file.Header.Open()
			if err != nil {
				return newError("failed to open file: %v", err)
			}
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				return newError("failed to read file: %v", err)
			}

			return NewBytesBufferFromBytes(data)
		},
	},

	// saveUploadedFile saves an uploaded file with validation.
	// This is a higher-level function that combines validation and saving.
	// Usage: result = saveUploadedFile(request, fieldName, dir, options)
	// options is a map with: maxSize, allowedExtensions, autoRename
	"saveUploadedFile": {
		Fn: func(args ...Object) Object {
			if len(args) < 3 || len(args) > 4 {
				return newError("wrong number of arguments for saveUploadedFile. got=%d, want=3 or 4", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'saveUploadedFile' must be HTTP_REQ, got %s", args[0].Type())
			}

			fieldName, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'saveUploadedFile' must be STRING, got %s", args[1].Type())
			}

			dirPath, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'saveUploadedFile' must be STRING, got %s", args[2].Type())
			}

			if req.Value == nil {
				return NewFileUploadResult(false, "http request is nil", "", "", 0)
			}

			// Parse options
			maxSize := int64(10 * 1024 * 1024) // 10MB default
			autoRename := true
			var allowedExtensions []string

			if len(args) == 4 {
				opts, ok := args[3].(*Map)
				if !ok {
					return newError("fourth argument to 'saveUploadedFile' must be MAP, got %s", args[3].Type())
				}

				for _, pair := range opts.Pairs {
					keyStr, ok := pair.Key.(*String)
					if !ok {
						continue
					}

					switch keyStr.Value {
					case "maxSize":
						if v, ok := pair.Value.(*Int); ok {
							maxSize = v.Value
						}
					case "autoRename":
						if v, ok := pair.Value.(*Bool); ok {
							autoRename = v.Value
						}
					case "allowedExtensions":
						if arr, ok := pair.Value.(*Array); ok {
							for _, elem := range arr.Elements {
								if s, ok := elem.(*String); ok {
									allowedExtensions = append(allowedExtensions, s.Value)
								}
							}
						}
					}
				}
			}

			// Parse multipart form
			if err := req.Value.ParseMultipartForm(maxSize); err != nil {
				return NewFileUploadResult(false, fmt.Sprintf("failed to parse form: %v", err), "", "", 0)
			}

			// Get file
			file, header, err := req.Value.FormFile(fieldName.Value)
			if err != nil {
				return NewFileUploadResult(false, fmt.Sprintf("file not found: %v", err), "", "", 0)
			}
			file.Close()

			upload := NewFileUpload(header)

			// Validate size
			if header.Size > maxSize {
				return NewFileUploadResult(false, fmt.Sprintf("file size %d exceeds maximum %d", header.Size, maxSize), "", header.Filename, header.Size)
			}

			// Validate extension
			if len(allowedExtensions) > 0 {
				ext := strings.ToLower(filepath.Ext(header.Filename))
				allowed := false
				for _, ae := range allowedExtensions {
					checkExt := strings.ToLower(ae)
					if !strings.HasPrefix(checkExt, ".") {
						checkExt = "." + checkExt
					}
					if checkExt == ext {
						allowed = true
						break
					}
				}
				if !allowed {
					return NewFileUploadResult(false, fmt.Sprintf("file extension %s is not allowed", ext), "", header.Filename, header.Size)
				}
			}

			// Save file
			savedPath, err := upload.SaveToDir(dirPath.Value, autoRename)
			if err != nil {
				return NewFileUploadResult(false, err.Error(), "", header.Filename, header.Size)
			}

			return NewFileUploadResult(true, "File saved successfully", savedPath, header.Filename, header.Size)
		},
	},
}

// RegisterFileUploadBuiltins registers file upload built-in functions into the main Builtins map.
func RegisterFileUploadBuiltins() {
	for name, builtin := range FileUploadBuiltins {
		Builtins[name] = builtin
	}
}

// Helper functions for multipart form handling

// parseMultipartFileHeader parses a multipart file header and returns file info.
func parseMultipartFileHeader(header *multipart.FileHeader) map[string]interface{} {
	return map[string]interface{}{
		"filename":     header.Filename,
		"size":         header.Size,
		"contentType":  header.Header.Get("Content-Type"),
		"extension":    filepath.Ext(header.Filename),
		"basename":     strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)),
	}
}

// saveMultipartFile saves a multipart file to disk.
func saveMultipartFile(header *multipart.FileHeader, destPath string) error {
	// Open source file
	src, err := header.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination directory if needed
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy content
	_, err = io.Copy(dst, src)
	return err
}
