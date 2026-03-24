// pkg/stdlib/zip.go
// ZIP file utilities for the Xxlang standard library.
// Provides comprehensive ZIP creation, extraction, and manipulation with full UTF-8 support.
//
// UTF-8/Unicode Filename Support:
// - Modern ZIP tools correctly set the UTF-8 flag for non-ASCII filenames
// - Go's archive/zip package automatically handles UTF-8 flagged entries
// - For legacy ZIP files with GBK/CP437 encoded names, use the gbkToUtf8 option
// - This module includes special handling for Chinese and other Unicode filenames
package stdlib

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/topxeq/xxlang/pkg/objects"
)

// gbkToUTF8 converts GBK encoded bytes to UTF-8 string.
// This is useful for legacy ZIP files created on Windows with Chinese filenames.
// Returns the original string if conversion fails.
func gbkToUTF8(s string) string {
	// Check if already valid UTF-8
	if strings.ToValidUTF8(s, "") == s {
		return s
	}

	// Try GBK to UTF-8 conversion
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return s // Return original on error
	}
	return string(result)
}

// zipWriter tracks open zip writers for multi-operation support
type zipWriter struct {
	writer   *zip.Writer
	file     *os.File
	path     string
	entries  map[string]bool
	modified bool
}

// zipWriterPool tracks open zip writers
var zipWriterPool = make(map[string]*zipWriter)

func init() {
	Register(&Module{
		Name: "zip",
		Exports: map[string]objects.Object{
			// ============================================================
			// High-level convenience functions
			// ============================================================

			// list lists all entries in a ZIP file.
			// Usage: entries = zip.list(path)
			// Returns an array of maps with name, size, compressedSize, modTime, isDir
			"list": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("list() takes exactly 1 argument: path")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("list() requires a string path")
				}

				reader, err := zip.OpenReader(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("list() failed: %s", err.Error()))
				}
				defer reader.Close()

				var entries []objects.Object
				for _, f := range reader.File {
					entry := make(map[objects.HashKey]objects.MapPair)

					entry[objects.NewString("name").HashKey()] = objects.MapPair{
						Key:   objects.NewString("name"),
						Value: objects.NewString(f.Name),
					}

					entry[objects.NewString("size").HashKey()] = objects.MapPair{
						Key:   objects.NewString("size"),
						Value: objects.NewInt(int64(f.UncompressedSize64)),
					}

					entry[objects.NewString("compressedSize").HashKey()] = objects.MapPair{
						Key:   objects.NewString("compressedSize"),
						Value: objects.NewInt(int64(f.CompressedSize64)),
					}

					entry[objects.NewString("modTime").HashKey()] = objects.MapPair{
						Key:   objects.NewString("modTime"),
						Value: objects.NewString(f.Modified.Format("2006-01-02 15:04:05")),
					}

					entry[objects.NewString("isDir").HashKey()] = objects.MapPair{
						Key:   objects.NewString("isDir"),
						Value: Bool(f.FileInfo().IsDir()),
					}

					entry[objects.NewString("method").HashKey()] = objects.MapPair{
						Key:   objects.NewString("method"),
						Value: objects.NewString(map[uint16]string{
							zip.Store:   "store",
							zip.Deflate: "deflate",
						}[f.Method]),
					}

					entries = append(entries, objects.NewMap(entry))
				}

				return Array(entries...)
			}),

			// extract extracts all files from a ZIP file to a directory.
			// Usage: zip.extract(zipPath, destDir)
			// Returns null on success, error on failure
			"extract": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("extract() takes exactly 2 arguments: zipPath, destDir")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("extract() first argument must be a string (zip path)")
				}

				destDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("extract() second argument must be a string (destination directory)")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("extract() failed to open: %s", err.Error()))
				}
				defer reader.Close()

				// Create destination directory
				if err := os.MkdirAll(destDir.Value, 0755); err != nil {
					return Error(fmt.Sprintf("extract() failed to create dir: %s", err.Error()))
				}

				for _, f := range reader.File {
					// Handle UTF-8 encoded names properly
					name := f.Name
					if f.NonUTF8 {
						// Try to decode as UTF-8 anyway (many tools set this incorrectly)
						name = strings.ToValidUTF8(name, "?")
					}

					// Sanitize path to prevent zip slip
					targetPath := filepath.Join(destDir.Value, name)
					if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir.Value)) {
						return Error(fmt.Sprintf("extract() unsafe path: %s", name))
					}

					if f.FileInfo().IsDir() {
						os.MkdirAll(targetPath, 0755)
						continue
					}

					// Create parent directories
					if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
						return Error(fmt.Sprintf("extract() failed to create parent dir: %s", err.Error()))
					}

					// Extract file
					src, err := f.Open()
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extract() failed to open entry: %s", err.Error()))
					}

					dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extract() failed to create file: %s", err.Error()))
					}

					_, err = io.Copy(dst, src)
					src.Close()
					dst.Close()

					if err != nil {
						return Error(fmt.Sprintf("extract() failed to write: %s", err.Error()))
					}

					// Set modification time
					os.Chtimes(targetPath, f.Modified, f.Modified)
				}

				return Null()
			}),

			// extractFile extracts a single file from a ZIP archive.
			// Usage: zip.extractFile(zipPath, entryName, destPath)
			"extractFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("extractFile() takes exactly 3 arguments: zipPath, entryName, destPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractFile() first argument must be a string")
				}

				entryName, ok := args[1].(*objects.String)
				if !ok {
					return Error("extractFile() second argument must be a string")
				}

				destPath, ok := args[2].(*objects.String)
				if !ok {
					return Error("extractFile() third argument must be a string")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("extractFile() failed: %s", err.Error()))
				}
				defer reader.Close()

				for _, f := range reader.File {
					if f.Name == entryName.Value {
						src, err := f.Open()
						if err != nil {
							src.Close()
							return Error(fmt.Sprintf("extractFile() failed to open entry: %s", err.Error()))
						}

						// Create parent directories
						if err := os.MkdirAll(filepath.Dir(destPath.Value), 0755); err != nil {
							src.Close()
							return Error(fmt.Sprintf("extractFile() failed to create dir: %s", err.Error()))
						}

						dst, err := os.OpenFile(destPath.Value, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
						if err != nil {
							src.Close()
							return Error(fmt.Sprintf("extractFile() failed to create file: %s", err.Error()))
						}

						_, err = io.Copy(dst, src)
						src.Close()
						dst.Close()

						if err != nil {
							return Error(fmt.Sprintf("extractFile() failed to write: %s", err.Error()))
						}

						return Null()
					}
				}

				return Error(fmt.Sprintf("extractFile() entry not found: %s", entryName.Value))
			}),

			// readEntry reads a file from a ZIP archive and returns its content as string.
			// Usage: content = zip.readEntry(zipPath, entryName)
			"readEntry": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("readEntry() takes exactly 2 arguments: zipPath, entryName")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("readEntry() first argument must be a string")
				}

				entryName, ok := args[1].(*objects.String)
				if !ok {
					return Error("readEntry() second argument must be a string")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("readEntry() failed: %s", err.Error()))
				}
				defer reader.Close()

				for _, f := range reader.File {
					if f.Name == entryName.Value {
						src, err := f.Open()
						if err != nil {
							src.Close()
							return Error(fmt.Sprintf("readEntry() failed to open entry: %s", err.Error()))
						}

						content, err := io.ReadAll(src)
						src.Close()

						if err != nil {
							return Error(fmt.Sprintf("readEntry() failed to read: %s", err.Error()))
						}

						return String(string(content))
					}
				}

				return Error(fmt.Sprintf("readEntry() entry not found: %s", entryName.Value))
			}),

			// readEntryBytes reads a file from a ZIP archive and returns its content as byte array.
			// Usage: bytes = zip.readEntryBytes(zipPath, entryName)
			"readEntryBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("readEntryBytes() takes exactly 2 arguments: zipPath, entryName")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("readEntryBytes() first argument must be a string")
				}

				entryName, ok := args[1].(*objects.String)
				if !ok {
					return Error("readEntryBytes() second argument must be a string")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("readEntryBytes() failed: %s", err.Error()))
				}
				defer reader.Close()

				for _, f := range reader.File {
					if f.Name == entryName.Value {
						src, err := f.Open()
						if err != nil {
							src.Close()
							return Error(fmt.Sprintf("readEntryBytes() failed to open entry: %s", err.Error()))
						}

						content, err := io.ReadAll(src)
						src.Close()

						if err != nil {
							return Error(fmt.Sprintf("readEntryBytes() failed to read: %s", err.Error()))
						}

						result := make([]objects.Object, len(content))
						for i, b := range content {
							result[i] = Int(int64(b))
						}
						return Array(result...)
					}
				}

				return Error(fmt.Sprintf("readEntryBytes() entry not found: %s", entryName.Value))
			}),

			// ============================================================
			// ZIP creation functions
			// ============================================================

			// create creates a new ZIP file.
			// Usage: zip.create(path)
			// Returns a zip handle (string) for use with other functions
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("create() takes exactly 1 argument: path")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("create() requires a string path")
				}

				file, err := os.Create(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("create() failed: %s", err.Error()))
				}

				writer := zip.NewWriter(file)

				handle := fmt.Sprintf("zip:%s", path.Value)
				zipWriterPool[handle] = &zipWriter{
					writer:  writer,
					file:    file,
					path:    path.Value,
					entries: make(map[string]bool),
				}

				return String(handle)
			}),

			// open opens an existing ZIP file for modification.
			// Usage: handle = zip.open(path)
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("open() takes exactly 1 argument: path")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("open() requires a string path")
				}

				// Read existing content
				content, err := os.ReadFile(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("open() failed: %s", err.Error()))
				}

				// Open file for writing
				file, err := os.Create(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("open() failed: %s", err.Error()))
				}

				writer := zip.NewWriter(file)

				handle := fmt.Sprintf("zip:%s", path.Value)
				zw := &zipWriter{
					writer:  writer,
					file:    file,
					path:    path.Value,
					entries: make(map[string]bool),
				}

				// Copy existing entries if any
				if len(content) > 0 {
					reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
					if err == nil {
						for _, f := range reader.File {
							zw.entries[f.Name] = true
						}
					}
				}

				zipWriterPool[handle] = zw
				return String(handle)
			}),

			// close closes a ZIP handle and writes the file.
			// Usage: zip.close(handle)
			"close": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("close() takes exactly 1 argument: handle")
				}

				handle, ok := args[0].(*objects.String)
				if !ok {
					return Error("close() requires a string handle")
				}

				zw, exists := zipWriterPool[handle.Value]
				if !exists {
					return Error("close() invalid handle")
				}

				err := zw.writer.Close()
				zw.file.Close()
				delete(zipWriterPool, handle.Value)

				if err != nil {
					return Error(fmt.Sprintf("close() failed: %s", err.Error()))
				}

				return Null()
			}),

			// addString adds a string as a file to the ZIP.
			// Usage: zip.addString(handle, name, content)
			// Supports UTF-8 filenames including Chinese characters
			"addString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("addString() takes exactly 3 arguments: handle, name, content")
				}

				handle, ok := args[0].(*objects.String)
				if !ok {
					return Error("addString() first argument must be a string handle")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("addString() second argument must be a string name")
				}

				content, ok := args[2].(*objects.String)
				if !ok {
					return Error("addString() third argument must be a string content")
				}

				zw, exists := zipWriterPool[handle.Value]
				if !exists {
					return Error("addString() invalid handle")
				}

				// Create header with UTF-8 support
				header := &zip.FileHeader{
					Name:     name.Value,
					Method:   zip.Deflate,
					Modified: time.Now(),
				}
				header.SetMode(0644)

				// The Go zip package automatically sets the UTF-8 flag for non-ASCII names
				writer, err := zw.writer.CreateHeader(header)
				if err != nil {
					return Error(fmt.Sprintf("addString() failed: %s", err.Error()))
				}

				_, err = writer.Write([]byte(content.Value))
				if err != nil {
					return Error(fmt.Sprintf("addString() failed to write: %s", err.Error()))
				}

				zw.entries[name.Value] = true
				return Null()
			}),

			// addBytes adds bytes as a file to the ZIP.
			// Usage: zip.addBytes(handle, name, byteArray)
			"addBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("addBytes() takes exactly 3 arguments: handle, name, byteArray")
				}

				handle, ok := args[0].(*objects.String)
				if !ok {
					return Error("addBytes() first argument must be a string handle")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("addBytes() second argument must be a string name")
				}

				arr, ok := args[2].(*objects.Array)
				if !ok {
					return Error("addBytes() third argument must be an array of bytes")
				}

				zw, exists := zipWriterPool[handle.Value]
				if !exists {
					return Error("addBytes() invalid handle")
				}

				// Convert array to bytes
				data := make([]byte, len(arr.Elements))
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("addBytes() array must contain integers 0-255")
					}
					if n.Value < 0 || n.Value > 255 {
						return Error("addBytes() byte values must be 0-255")
					}
					data[i] = byte(n.Value)
				}

				header := &zip.FileHeader{
					Name:     name.Value,
					Method:   zip.Deflate,
					Modified: time.Now(),
				}
				header.SetMode(0644)

				writer, err := zw.writer.CreateHeader(header)
				if err != nil {
					return Error(fmt.Sprintf("addBytes() failed: %s", err.Error()))
				}

				_, err = writer.Write(data)
				if err != nil {
					return Error(fmt.Sprintf("addBytes() failed to write: %s", err.Error()))
				}

				zw.entries[name.Value] = true
				return Null()
			}),

			// addFile adds a file from filesystem to the ZIP.
			// Usage: zip.addFile(handle, filePath, [nameInZip])
			// nameInZip defaults to the base filename if not provided
			"addFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("addFile() takes at least 2 arguments: handle, filePath, [nameInZip]")
				}

				handle, ok := args[0].(*objects.String)
				if !ok {
					return Error("addFile() first argument must be a string handle")
				}

				filePath, ok := args[1].(*objects.String)
				if !ok {
					return Error("addFile() second argument must be a string filePath")
				}

				zw, exists := zipWriterPool[handle.Value]
				if !exists {
					return Error("addFile() invalid handle")
				}

				// Read file info
				info, err := os.Stat(filePath.Value)
				if err != nil {
					return Error(fmt.Sprintf("addFile() failed to stat: %s", err.Error()))
				}

				if info.IsDir() {
					return Error("addFile() cannot add a directory, use addDir() instead")
				}

				// Determine name in zip
				nameInZip := filepath.Base(filePath.Value)
				if len(args) > 2 {
					if n, ok := args[2].(*objects.String); ok {
						nameInZip = n.Value
					}
				}

				// Read file content
				content, err := os.ReadFile(filePath.Value)
				if err != nil {
					return Error(fmt.Sprintf("addFile() failed to read: %s", err.Error()))
				}

				header := &zip.FileHeader{
					Name:     nameInZip,
					Method:   zip.Deflate,
					Modified: info.ModTime(),
				}
				header.SetMode(info.Mode())

				writer, err := zw.writer.CreateHeader(header)
				if err != nil {
					return Error(fmt.Sprintf("addFile() failed: %s", err.Error()))
				}

				_, err = writer.Write(content)
				if err != nil {
					return Error(fmt.Sprintf("addFile() failed to write: %s", err.Error()))
				}

				zw.entries[nameInZip] = true
				return Null()
			}),

			// addDir adds a directory (recursively) to the ZIP.
			// Usage: zip.addDir(handle, dirPath, [prefixInZip])
			"addDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("addDir() takes at least 2 arguments: handle, dirPath, [prefixInZip]")
				}

				handle, ok := args[0].(*objects.String)
				if !ok {
					return Error("addDir() first argument must be a string handle")
				}

				dirPath, ok := args[1].(*objects.String)
				if !ok {
					return Error("addDir() second argument must be a string dirPath")
				}

				zw, exists := zipWriterPool[handle.Value]
				if !exists {
					return Error("addDir() invalid handle")
				}

				prefixInZip := ""
				if len(args) > 2 {
					if p, ok := args[2].(*objects.String); ok {
						prefixInZip = p.Value
					}
				}

				err := filepath.Walk(dirPath.Value, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// Calculate relative path
					relPath, err := filepath.Rel(dirPath.Value, path)
					if err != nil {
						return err
					}

					// Create name in zip
					nameInZip := filepath.Join(prefixInZip, relPath)
					// Convert to forward slashes for zip
					nameInZip = filepath.ToSlash(nameInZip)

					if info.IsDir() {
						// Add directory entry
						nameInZip += "/"
						header := &zip.FileHeader{
							Name:     nameInZip,
							Modified: info.ModTime(),
						}
						header.SetMode(info.Mode() | os.ModeDir)
						_, err := zw.writer.CreateHeader(header)
						return err
					}

					// Add file
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					header := &zip.FileHeader{
						Name:     nameInZip,
						Method:   zip.Deflate,
						Modified: info.ModTime(),
					}
					header.SetMode(info.Mode())

					writer, err := zw.writer.CreateHeader(header)
					if err != nil {
						return err
					}

					_, err = writer.Write(content)
					if err != nil {
						return err
					}

					zw.entries[nameInZip] = true
					return nil
				})

				if err != nil {
					return Error(fmt.Sprintf("addDir() failed: %s", err.Error()))
				}

				return Null()
			}),

			// ============================================================
			// Convenience functions (single operation)
			// ============================================================

			// writeFile creates a ZIP file with the given entries.
			// Usage: zip.writeFile(zipPath, entries)
			// entries is a map of {name: content} or {name: {content: "...", mode: 0644}}
			"writeFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeFile() takes exactly 2 arguments: zipPath, entries")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFile() first argument must be a string path")
				}

				entries, ok := args[1].(*objects.Map)
				if !ok {
					return Error("writeFile() second argument must be a map of entries")
				}

				file, err := os.Create(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}
				defer file.Close()

				writer := zip.NewWriter(file)
				defer writer.Close()

				for _, pair := range entries.Pairs {
					name, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}

					var content []byte
					var modTime time.Time = time.Now()
					var mode fs.FileMode = 0644

					switch v := pair.Value.(type) {
					case *objects.String:
						content = []byte(v.Value)
					case *objects.Map:
						// Check for content field
						if cPair, found := v.Pairs[objects.NewString("content").HashKey()]; found {
							if c, ok := cPair.Value.(*objects.String); ok {
								content = []byte(c.Value)
							} else if arr, ok := cPair.Value.(*objects.Array); ok {
								content = make([]byte, len(arr.Elements))
								for i, elem := range arr.Elements {
									if n, ok := elem.(*objects.Int); ok {
										content[i] = byte(n.Value)
									}
								}
							}
						}
						// Check for mode field
						if mPair, found := v.Pairs[objects.NewString("mode").HashKey()]; found {
							if m, ok := mPair.Value.(*objects.Int); ok {
								mode = fs.FileMode(m.Value)
							}
						}
					case *objects.Array:
						content = make([]byte, len(v.Elements))
						for i, elem := range v.Elements {
							if n, ok := elem.(*objects.Int); ok {
								content[i] = byte(n.Value)
							}
						}
					default:
						continue
					}

					header := &zip.FileHeader{
						Name:     name.Value,
						Method:   zip.Deflate,
						Modified: modTime,
					}
					header.SetMode(mode)

					w, err := writer.CreateHeader(header)
					if err != nil {
						return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
					}

					_, err = w.Write(content)
					if err != nil {
						return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
					}
				}

				return Null()
			}),

			// ============================================================
			// Utility functions
			// ============================================================

			// isValid checks if a file is a valid ZIP file.
			// Usage: valid = zip.isValid(path)
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument: path")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("isValid() requires a string path")
				}

				reader, err := zip.OpenReader(path.Value)
				if err != nil {
					return Bool(false)
				}
				reader.Close()
				return Bool(true)
			}),

			// hasEntry checks if a ZIP file contains a specific entry.
			// Usage: exists = zip.hasEntry(zipPath, entryName)
			"hasEntry": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("hasEntry() takes exactly 2 arguments: zipPath, entryName")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("hasEntry() first argument must be a string")
				}

				entryName, ok := args[1].(*objects.String)
				if !ok {
					return Error("hasEntry() second argument must be a string")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Bool(false)
				}
				defer reader.Close()

				for _, f := range reader.File {
					if f.Name == entryName.Value {
						return Bool(true)
					}
				}

				return Bool(false)
			}),

			// count returns the number of entries in a ZIP file.
			// Usage: n = zip.count(zipPath)
			"count": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("count() takes exactly 1 argument: zipPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("count() requires a string path")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("count() failed: %s", err.Error()))
				}
				defer reader.Close()

				return Int(int64(len(reader.File)))
			}),

			// ============================================================
			// Enhanced UTF-8/Chinese filename support
			// ============================================================

			// extractWithEncoding extracts all files from a ZIP with specified filename encoding.
			// This is useful for legacy ZIP files with non-UTF-8 encoded filenames.
			// Usage: zip.extractWithEncoding(zipPath, destDir, "gbk")
			// Supported encodings: "utf8" (default), "gbk"
			// Returns null on success, error on failure
			"extractWithEncoding": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("extractWithEncoding() takes at least 2 arguments: zipPath, destDir, [encoding]")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractWithEncoding() first argument must be a string (zip path)")
				}

				destDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("extractWithEncoding() second argument must be a string (destination directory)")
				}

				encoding := "utf8"
				if len(args) > 2 {
					enc, ok := args[2].(*objects.String)
					if ok {
						encoding = strings.ToLower(enc.Value)
					}
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("extractWithEncoding() failed to open: %s", err.Error()))
				}
				defer reader.Close()

				// Create destination directory
				if err := os.MkdirAll(destDir.Value, 0755); err != nil {
					return Error(fmt.Sprintf("extractWithEncoding() failed to create dir: %s", err.Error()))
				}

				for _, f := range reader.File {
					// Handle filename encoding
					name := f.Name
					if encoding == "gbk" || (f.NonUTF8 && encoding != "utf8") {
						// Try to decode as GBK for Chinese filenames
						name = decodeFilename(f.Name, encoding)
					} else if f.NonUTF8 {
						// Try to interpret as UTF-8 anyway
						name = strings.ToValidUTF8(name, "?")
					}

					// Sanitize path to prevent zip slip
					targetPath := filepath.Join(destDir.Value, name)
					if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir.Value)) {
						return Error(fmt.Sprintf("extractWithEncoding() unsafe path: %s", name))
					}

					if f.FileInfo().IsDir() {
						os.MkdirAll(targetPath, 0755)
						continue
					}

					// Create parent directories
					if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
						return Error(fmt.Sprintf("extractWithEncoding() failed to create parent dir: %s", err.Error()))
					}

					// Extract file
					src, err := f.Open()
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extractWithEncoding() failed to open entry: %s", err.Error()))
					}

					dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extractWithEncoding() failed to create file: %s", err.Error()))
					}

					_, err = io.Copy(dst, src)
					src.Close()
					dst.Close()

					if err != nil {
						return Error(fmt.Sprintf("extractWithEncoding() failed to write: %s", err.Error()))
					}

					// Set modification time
					os.Chtimes(targetPath, f.Modified, f.Modified)
				}

				return Null()
			}),

			// listNames returns a list of entry names in a ZIP file.
			// Usage: names = zip.listNames(zipPath)
			// Returns an array of strings
			"listNames": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("listNames() takes exactly 1 argument: zipPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("listNames() requires a string path")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("listNames() failed: %s", err.Error()))
				}
				defer reader.Close()

				var names []objects.Object
				for _, f := range reader.File {
					names = append(names, String(f.Name))
				}

				return Array(names...)
			}),

			// listNamesWithEncoding returns entry names with specified encoding.
			// Usage: names = zip.listNamesWithEncoding(zipPath, "gbk")
			"listNamesWithEncoding": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("listNamesWithEncoding() takes at least 1 argument: zipPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("listNamesWithEncoding() requires a string path")
				}

				encoding := "utf8"
				if len(args) > 1 {
					enc, ok := args[1].(*objects.String)
					if ok {
						encoding = strings.ToLower(enc.Value)
					}
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("listNamesWithEncoding() failed: %s", err.Error()))
				}
				defer reader.Close()

				var names []objects.Object
				for _, f := range reader.File {
					name := f.Name
					if encoding == "gbk" || (f.NonUTF8 && encoding != "utf8") {
						name = decodeFilename(f.Name, encoding)
					}
					names = append(names, String(name))
				}

				return Array(names...)
			}),

			// getInfo returns detailed information about a ZIP file.
			// Usage: info = zip.getInfo(zipPath)
			// Returns: map with fileCount, totalSize, compressedSize, compressionRatio
			"getInfo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getInfo() takes exactly 1 argument: zipPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("getInfo() requires a string path")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("getInfo() failed: %s", err.Error()))
				}
				defer reader.Close()

				var totalSize, compressedSize int64
				fileCount := len(reader.File)

				for _, f := range reader.File {
					totalSize += int64(f.UncompressedSize64)
					compressedSize += int64(f.CompressedSize64)
				}

				ratio := float64(0)
				if totalSize > 0 {
					ratio = float64(compressedSize) / float64(totalSize) * 100
				}

				return &objects.Map{
					Pairs: map[objects.HashKey]objects.MapPair{
						objects.NewString("fileCount").HashKey(): {
							Key:   objects.NewString("fileCount"),
							Value: objects.NewInt(int64(fileCount)),
						},
						objects.NewString("totalSize").HashKey(): {
							Key:   objects.NewString("totalSize"),
							Value: objects.NewInt(totalSize),
						},
						objects.NewString("compressedSize").HashKey(): {
							Key:   objects.NewString("compressedSize"),
							Value: objects.NewInt(compressedSize),
						},
						objects.NewString("compressionRatio").HashKey(): {
							Key:   objects.NewString("compressionRatio"),
							Value: objects.NewFloat(ratio),
						},
					},
				}
			}),

			// removeEntry removes an entry from a ZIP file.
			// Usage: zip.removeEntry(zipPath, entryName)
			// Note: This rewrites the entire ZIP file without the specified entry
			"removeEntry": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("removeEntry() takes exactly 2 arguments: zipPath, entryName")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("removeEntry() first argument must be a string")
				}

				entryName, ok := args[1].(*objects.String)
				if !ok {
					return Error("removeEntry() second argument must be a string")
				}

				// Read existing content
				content, err := os.ReadFile(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("removeEntry() failed to read: %s", err.Error()))
				}

				reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
				if err != nil {
					return Error(fmt.Sprintf("removeEntry() failed to parse: %s", err.Error()))
				}

				// Create a buffer for the new zip
				var buf bytes.Buffer
				writer := zip.NewWriter(&buf)

				found := false
				for _, f := range reader.File {
					if f.Name == entryName.Value {
						found = true
						continue // Skip this entry
					}

					// Copy the entry
					src, err := f.Open()
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("removeEntry() failed to open entry: %s", err.Error()))
					}

					dst, err := writer.CreateHeader(&f.FileHeader)
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("removeEntry() failed to create entry: %s", err.Error()))
					}

					_, err = io.Copy(dst, src)
					src.Close()
					if err != nil {
						return Error(fmt.Sprintf("removeEntry() failed to copy: %s", err.Error()))
					}
				}

				writer.Close()

				if !found {
					return Error(fmt.Sprintf("removeEntry() entry not found: %s", entryName.Value))
				}

				// Write the new zip content
				err = os.WriteFile(zipPath.Value, buf.Bytes(), 0644)
				if err != nil {
					return Error(fmt.Sprintf("removeEntry() failed to write: %s", err.Error()))
				}

				return Null()
			}),

			// renameEntry renames an entry in a ZIP file.
			// Usage: zip.renameEntry(zipPath, oldName, newName)
			// Note: This rewrites the entire ZIP file with the renamed entry
			"renameEntry": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("renameEntry() takes exactly 3 arguments: zipPath, oldName, newName")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("renameEntry() first argument must be a string")
				}

				oldName, ok := args[1].(*objects.String)
				if !ok {
					return Error("renameEntry() second argument must be a string")
				}

				newName, ok := args[2].(*objects.String)
				if !ok {
					return Error("renameEntry() third argument must be a string")
				}

				// Read existing content
				content, err := os.ReadFile(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("renameEntry() failed to read: %s", err.Error()))
				}

				reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
				if err != nil {
					return Error(fmt.Sprintf("renameEntry() failed to parse: %s", err.Error()))
				}

				// Create a buffer for the new zip
				var buf bytes.Buffer
				writer := zip.NewWriter(&buf)

				found := false
				for _, f := range reader.File {
					src, err := f.Open()
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("renameEntry() failed to open entry: %s", err.Error()))
					}

					// Copy header and potentially rename
					header := f.FileHeader
					if f.Name == oldName.Value {
						header.Name = newName.Value
						found = true
					}

					dst, err := writer.CreateHeader(&header)
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("renameEntry() failed to create entry: %s", err.Error()))
					}

					_, err = io.Copy(dst, src)
					src.Close()
					if err != nil {
						return Error(fmt.Sprintf("renameEntry() failed to copy: %s", err.Error()))
					}
				}

				writer.Close()

				if !found {
					return Error(fmt.Sprintf("renameEntry() entry not found: %s", oldName.Value))
				}

				// Write the new zip content
				err = os.WriteFile(zipPath.Value, buf.Bytes(), 0644)
				if err != nil {
					return Error(fmt.Sprintf("renameEntry() failed to write: %s", err.Error()))
				}

				return Null()
			}),

			// extractByPattern extracts files matching a glob pattern.
			// Usage: count = zip.extractByPattern(zipPath, destDir, "*.txt")
			// Returns the number of files extracted
			"extractByPattern": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("extractByPattern() takes exactly 3 arguments: zipPath, destDir, pattern")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractByPattern() first argument must be a string")
				}

				destDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("extractByPattern() second argument must be a string")
				}

				pattern, ok := args[2].(*objects.String)
				if !ok {
					return Error("extractByPattern() third argument must be a string")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("extractByPattern() failed: %s", err.Error()))
				}
				defer reader.Close()

				// Create destination directory
				if err := os.MkdirAll(destDir.Value, 0755); err != nil {
					return Error(fmt.Sprintf("extractByPattern() failed to create dir: %s", err.Error()))
				}

				count := 0
				for _, f := range reader.File {
					// Match pattern against the base filename
					matched, err := filepath.Match(pattern.Value, filepath.Base(f.Name))
					if err != nil {
						return Error(fmt.Sprintf("extractByPattern() invalid pattern: %s", err.Error()))
					}

					if !matched {
						continue
					}

					// Sanitize path
					targetPath := filepath.Join(destDir.Value, f.Name)
					if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir.Value)) {
						continue // Skip unsafe paths
					}

					if f.FileInfo().IsDir() {
						os.MkdirAll(targetPath, 0755)
						continue
					}

					// Create parent directories
					if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
						return Error(fmt.Sprintf("extractByPattern() failed: %s", err.Error()))
					}

					// Extract file
					src, err := f.Open()
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extractByPattern() failed: open: %s", err.Error()))
					}

					dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
					if err != nil {
						src.Close()
						return Error(fmt.Sprintf("extractByPattern() failed: create: %s", err.Error()))
					}

					_, err = io.Copy(dst, src)
					src.Close()
					dst.Close()

					if err != nil {
						return Error(fmt.Sprintf("extractByPattern() failed to write: %s", err.Error()))
					}

					count++
				}

				return Int(int64(count))
			}),

			// ============================================================
			// In-memory ZIP operations
			// ============================================================

			// createFromMap creates a ZIP file from a map of filename -> content.
			// This is useful for creating ZIP files in memory without filesystem operations.
			// Usage: zip.createFromMap(zipPath, {"文件名.txt": "内容", "data.bin": [1, 2, 3]})
			// Supports both string content and byte array content
			"createFromMap": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("createFromMap() takes exactly 2 arguments: zipPath, entries")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("createFromMap() first argument must be a string path")
				}

				entries, ok := args[1].(*objects.Map)
				if !ok {
					return Error("createFromMap() second argument must be a map of entries")
				}

				file, err := os.Create(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("createFromMap() failed: %s", err.Error()))
				}
				defer file.Close()

				writer := zip.NewWriter(file)
				defer writer.Close()

				for _, pair := range entries.Pairs {
					name, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}

					var content []byte
					switch v := pair.Value.(type) {
					case *objects.String:
						content = []byte(v.Value)
					case *objects.Array:
						content = make([]byte, len(v.Elements))
						for i, elem := range v.Elements {
							if n, ok := elem.(*objects.Int); ok && n.Value >= 0 && n.Value <= 255 {
								content[i] = byte(n.Value)
							}
						}
					case *objects.Bytes:
						content = v.Value
					default:
						continue
					}

					// Create header with UTF-8 support for non-ASCII names
					header := &zip.FileHeader{
						Name:     name.Value,
						Method:   zip.Deflate,
						Modified: time.Now(),
					}
					header.SetMode(0644)

					w, err := writer.CreateHeader(header)
					if err != nil {
						return Error(fmt.Sprintf("createFromMap() failed: %s", err.Error()))
					}

					_, err = w.Write(content)
					if err != nil {
						return Error(fmt.Sprintf("createFromMap() failed to write: %s", err.Error()))
					}
				}

				return Null()
			}),

			// readAll reads all entries from a ZIP file and returns as a map.
			// Usage: entries = zip.readAll(zipPath)
			// Returns: map of {filename: content}
			"readAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readAll() takes exactly 1 argument: zipPath")
				}

				zipPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("readAll() requires a string path")
				}

				reader, err := zip.OpenReader(zipPath.Value)
				if err != nil {
					return Error(fmt.Sprintf("readAll() failed: %s", err.Error()))
				}
				defer reader.Close()

				pairs := make(map[objects.HashKey]objects.MapPair)

				for _, f := range reader.File {
					if f.FileInfo().IsDir() {
						continue
					}

					src, err := f.Open()
					if err != nil {
						src.Close()
						continue
					}

					content, err := io.ReadAll(src)
					src.Close()
					if err != nil {
						continue
					}

					key := objects.NewString(f.Name)
					pairs[key.HashKey()] = objects.MapPair{
						Key:   key,
						Value: String(string(content)),
					}
				}

				return objects.NewMap(pairs)
			}),

			// setPassword sets the password for encrypted ZIP entries.
			// Note: Standard library doesn't support encrypted ZIPs, this is a placeholder.
			// Usage: zip.setPassword(handle, "password")
			"setPassword": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Error("setPassword() is not supported - encrypted ZIPs require external libraries")
			}),
		},
	})
}

// decodeFilename attempts to decode a filename using the specified encoding.
// This is particularly useful for ZIP files created on Windows with Chinese filenames.
// Supported encodings: "utf8" (default), "gbk", "gb18030"
func decodeFilename(name string, encoding string) string {
	// Check if already valid UTF-8
	if strings.ToValidUTF8(name, "") == name {
		return name
	}

	encoding = strings.ToLower(encoding)

	switch encoding {
	case "gbk", "gb18030":
		return gbkToUTF8(name)
	default:
		// Try GBK as fallback for non-UTF8 names
		return gbkToUTF8(name)
	}
}
