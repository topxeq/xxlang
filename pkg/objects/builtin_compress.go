// pkg/objects/builtin_compress.go
// Compression and archive built-in functions for Xxlang
package objects

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func init() {
	// Data compression
	Builtins["compressData"] = &Builtin{Fn: builtinCompressData}
	Builtins["uncompressData"] = &Builtin{Fn: builtinUncompressData}
	Builtins["compressStr"] = &Builtin{Fn: builtinCompressStr}
	Builtins["uncompressStr"] = &Builtin{Fn: builtinUncompressStr}

	// Zip archive functions
	Builtins["zipPath"] = &Builtin{Fn: builtinZipPath}
	Builtins["zipPaths"] = &Builtin{Fn: builtinZipPaths}
	Builtins["unzipToPath"] = &Builtin{Fn: builtinUnzipToPath}
	Builtins["getFileListInZip"] = &Builtin{Fn: builtinGetFileListInZip}
	Builtins["loadBytesInZip"] = &Builtin{Fn: builtinLoadBytesInZip}
	Builtins["addFileToZip"] = &Builtin{Fn: builtinAddFileToZip}
}

// builtinCompressData - compress byte data
// Usage: compressData(data) -> bytes (hex encoded string)
func builtinCompressData(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for compressData. got=%d, want=1", len(args))
	}

	var data []byte
	switch v := args[0].(type) {
	case *String:
		data = []byte(v.Value)
	case *Array:
		data = make([]byte, len(v.Elements))
		for i, elem := range v.Elements {
			if b, ok := elem.(*Int); ok {
				data[i] = byte(b.Value)
			} else {
				return newError("array elements must be integers 0-255")
			}
		}
	default:
		return newError("argument to 'compressData' must be STRING or ARRAY, got %s", args[0].Type())
	}

	var buf bytes.Buffer
	writer, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	writer.Write(data)
	writer.Close()

	// Return as hex string
	return NewString(encodeHex(buf.Bytes()))
}

// builtinUncompressData - uncompress byte data
// Usage: uncompressData(hexStr) -> string
func builtinUncompressData(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for uncompressData. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'uncompressData' must be STRING, got %s", args[0].Type())
	}

	data, err := decodeHex(str.Value)
	if err != nil {
		return newError("uncompressData hex decode error: %v", err)
	}

	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		return newError("uncompressData error: %v", err)
	}

	return NewString(string(result))
}

// builtinCompressStr - compress string
// Usage: compressStr(str) -> string (hex encoded)
func builtinCompressStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for compressStr. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'compressStr' must be STRING, got %s", args[0].Type())
	}

	var buf bytes.Buffer
	writer, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	writer.Write([]byte(str.Value))
	writer.Close()

	return NewString(encodeHex(buf.Bytes()))
}

// builtinUncompressStr - uncompress to string
// Usage: uncompressStr(hexStr) -> string
func builtinUncompressStr(args ...Object) Object {
	return builtinUncompressData(args...)
}

// builtinZipPath - compress a file or directory to zip
// Usage: zipPath(sourcePath, zipPath) -> null
func builtinZipPath(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for zipPath. got=%d, want=2 or 3", len(args))
	}

	sourcePath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'zipPath' must be STRING, got %s", args[0].Type())
	}

	zipPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'zipPath' must be STRING, got %s", args[1].Type())
	}

	// Check optional overwrite flag
	overwrite := false
	if len(args) == 3 {
		if opt, ok := args[2].(*String); ok {
			overwrite = opt.Value == "-overwrite"
		}
	}

	// Check if zip file exists
	if _, err := os.Stat(zipPath.Value); err == nil && !overwrite {
		return newError("zip file already exists: %s", zipPath.Value)
	}

	// Create zip file
	zipFile, err := os.Create(zipPath.Value)
	if err != nil {
		return newError("zipPath create error: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk source path
	err = filepath.Walk(sourcePath.Value, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(sourcePath.Value, path)
		if err != nil {
			return err
		}

		// Create zip entry
		if info.IsDir() {
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		// Add file to zip
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return newError("zipPath error: %v", err)
	}

	return NULL
}

// builtinZipPaths - compress multiple paths to zip
// Usage: zipPaths(paths, zipPath) -> null
//
//	zipPaths(paths, zipPath, options...) -> null
func builtinZipPaths(args ...Object) Object {
	if len(args) < 2 {
		return newError("wrong number of arguments for zipPaths. got=%d, want>=2", len(args))
	}

	paths, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'zipPaths' must be ARRAY, got %s", args[0].Type())
	}

	zipPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'zipPaths' must be STRING, got %s", args[1].Type())
	}

	// Create zip file
	zipFile, err := os.Create(zipPath.Value)
	if err != nil {
		return newError("zipPaths create error: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each path
	for _, elem := range paths.Elements {
		path, ok := elem.(*String)
		if !ok {
			continue
		}

		err = addPathToZip(zipWriter, path.Value)
		if err != nil {
			return newError("zipPaths error adding %s: %v", path.Value, err)
		}
	}

	return NULL
}

// builtinUnzipToPath - extract zip to directory
// Usage: unzipToPath(zipPath, destPath) -> null
func builtinUnzipToPath(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for unzipToPath. got=%d, want=2 or 3", len(args))
	}

	zipPath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'unzipToPath' must be STRING, got %s", args[0].Type())
	}

	destPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'unzipToPath' must be STRING, got %s", args[1].Type())
	}

	// Open zip file
	reader, err := zip.OpenReader(zipPath.Value)
	if err != nil {
		return newError("unzipToPath open error: %v", err)
	}
	defer reader.Close()

	// Extract each file
	for _, file := range reader.File {
		destFile := filepath.Join(destPath.Value, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(destFile, 0755)
			continue
		}

		// Ensure parent directory exists
		os.MkdirAll(filepath.Dir(destFile), 0755)

		// Extract file
		src, err := file.Open()
		if err != nil {
			src.Close()
			continue
		}

		dst, err := os.Create(destFile)
		if err != nil {
			src.Close()
			continue
		}

		io.Copy(dst, src)
		dst.Close()
		src.Close()
	}

	return NULL
}

// builtinGetFileListInZip - get list of files in zip
// Usage: getFileListInZip(zipPath) -> array
func builtinGetFileListInZip(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getFileListInZip. got=%d, want=1", len(args))
	}

	zipPath, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileListInZip' must be STRING, got %s", args[0].Type())
	}

	reader, err := zip.OpenReader(zipPath.Value)
	if err != nil {
		return newError("getFileListInZip open error: %v", err)
	}
	defer reader.Close()

	files := make([]Object, 0, len(reader.File))
	for _, file := range reader.File {
		files = append(files, NewString(file.Name))
	}

	return NewArray(files)
}

// builtinLoadBytesInZip - read file content from zip
// Usage: loadBytesInZip(zipPath, innerPath) -> string
func builtinLoadBytesInZip(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for loadBytesInZip. got=%d, want=2 or 3", len(args))
	}

	zipPath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'loadBytesInZip' must be STRING, got %s", args[0].Type())
	}

	innerPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'loadBytesInZip' must be STRING, got %s", args[1].Type())
	}

	reader, err := zip.OpenReader(zipPath.Value)
	if err != nil {
		return newError("loadBytesInZip open error: %v", err)
	}
	defer reader.Close()

	// Find file
	for _, file := range reader.File {
		if file.Name == innerPath.Value {
			src, err := file.Open()
			if err != nil {
				return newError("loadBytesInZip open inner error: %v", err)
			}
			defer src.Close()

			content, err := io.ReadAll(src)
			if err != nil {
				return newError("loadBytesInZip read error: %v", err)
			}

			return NewString(string(content))
		}
	}

	return newError("file not found in zip: %s", innerPath.Value)
}

// builtinAddFileToZip - add file to existing zip (creates new zip)
// Usage: addFileToZip(zipPath, filePath, innerPath) -> null
func builtinAddFileToZip(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments for addFileToZip. got=%d, want=3", len(args))
	}

	zipPath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'addFileToZip' must be STRING, got %s", args[0].Type())
	}

	filePath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'addFileToZip' must be STRING, got %s", args[1].Type())
	}

	innerPath, ok := args[2].(*String)
	if !ok {
		return newError("third argument to 'addFileToZip' must be STRING, got %s", args[2].Type())
	}

	// Read existing zip
	reader, err := zip.OpenReader(zipPath.Value)
	if err != nil {
		return newError("addFileToZip open error: %v", err)
	}

	// Create temp buffer
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Copy existing files
	for _, file := range reader.File {
		src, err := file.Open()
		if err != nil {
			src.Close()
			continue
		}

		writer, err := zipWriter.Create(file.Name)
		if err != nil {
			src.Close()
			continue
		}

		io.Copy(writer, src)
		src.Close()
	}
	reader.Close()

	// Add new file
	file, err := os.Open(filePath.Value)
	if err != nil {
		return newError("addFileToZip open file error: %v", err)
	}
	defer file.Close()

	writer, err := zipWriter.Create(innerPath.Value)
	if err != nil {
		return newError("addFileToZip create entry error: %v", err)
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return newError("addFileToZip copy error: %v", err)
	}

	zipWriter.Close()

	// Write back to zip file
	zipFile, err := os.Create(zipPath.Value)
	if err != nil {
		return newError("addFileToZip create zip error: %v", err)
	}
	defer zipFile.Close()

	zipFile.Write(buf.Bytes())

	return NULL
}

// Helper functions

func addPathToZip(zipWriter *zip.Writer, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(path, filePath)
			if err != nil {
				return err
			}

			if info.IsDir() {
				_, err = zipWriter.Create(relPath + "/")
				return err
			}

			writer, err := zipWriter.Create(relPath)
			if err != nil {
				return err
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
	}

	// Single file
	writer, err := zipWriter.Create(info.Name())
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

func encodeHex(data []byte) string {
	const hextable = "0123456789abcdef"
	dst := make([]byte, len(data)*2)
	for i, v := range data {
		dst[i*2] = hextable[v>>4]
		dst[i*2+1] = hextable[v&0x0f]
	}
	return string(dst)
}

func decodeHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string has odd length")
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		a := hexCharToByte(s[i])
		b := hexCharToByte(s[i+1])
		if a == 255 || b == 255 {
			return nil, errors.New("invalid hex character")
		}
		result[i/2] = (a << 4) | b
	}
	return result, nil
}

func hexCharToByte(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	default:
		return 255
	}
}
