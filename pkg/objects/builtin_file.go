// pkg/objects/builtin_file.go
// File system related built-in functions for Xxlang
package objects

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	// File system functions
	Builtins["fileExists"] = &Builtin{Fn: builtinFileExists}
	Builtins["dirExists"] = &Builtin{Fn: builtinDirExists}
	Builtins["pathExists"] = &Builtin{Fn: builtinPathExists}
	Builtins["isDir"] = &Builtin{Fn: builtinIsDir}
	Builtins["loadText"] = &Builtin{Fn: builtinLoadText}
	Builtins["saveText"] = &Builtin{Fn: builtinSaveText}
	Builtins["saveBytes"] = &Builtin{Fn: builtinSaveBytes}
	Builtins["loadBytes"] = &Builtin{Fn: builtinLoadBytes}
	Builtins["appendText"] = &Builtin{Fn: builtinAppendText}
	Builtins["copyFile"] = &Builtin{Fn: builtinCopyFile}
	Builtins["copyPath"] = &Builtin{Fn: builtinCopyPath}
	Builtins["renameFile"] = &Builtin{Fn: builtinRenameFile}
	Builtins["moveFile"] = &Builtin{Fn: builtinRenameFile} // alias for renameFile
	Builtins["removeFile"] = &Builtin{Fn: builtinRemoveFile}
	Builtins["removeDir"] = &Builtin{Fn: builtinRemoveDir}
	Builtins["getFileList"] = &Builtin{Fn: builtinGetFileList}
	Builtins["joinPath"] = &Builtin{Fn: builtinJoinPath}
	Builtins["getCurDir"] = &Builtin{Fn: builtinGetCurDir}
	Builtins["getHomeDir"] = &Builtin{Fn: builtinGetHomeDir}
	Builtins["getTempDir"] = &Builtin{Fn: builtinGetTempDir}
	Builtins["ensureMakeDirs"] = &Builtin{Fn: builtinEnsureMakeDirs}
	Builtins["getFileExt"] = &Builtin{Fn: builtinGetFileExt}
	Builtins["extractFileDir"] = &Builtin{Fn: builtinExtractFileDir}
	Builtins["extractFileName"] = &Builtin{Fn: builtinExtractFileName}
	Builtins["getFileInfo"] = &Builtin{Fn: builtinGetFileInfo}
	Builtins["getFileSize"] = &Builtin{Fn: builtinGetFileSize}
	Builtins["loadLines"] = &Builtin{Fn: builtinLoadLines}
	Builtins["getFileAbs"] = &Builtin{Fn: builtinGetFileAbs}
	Builtins["getFileRel"] = &Builtin{Fn: builtinGetFileRel}
	Builtins["isFile"] = &Builtin{Fn: builtinIsFile}
}

// fileExists - check if file exists
// Usage: fileExists(path) -> bool
func builtinFileExists(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for fileExists. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'fileExists' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if os.IsNotExist(err) {
		return FALSE
	}
	if err != nil {
		return newError("fileExists error: %v", err)
	}
	return &Bool{Value: !info.IsDir()}
}

// dirExists - check if directory exists
// Usage: dirExists(path) -> bool
func builtinDirExists(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for dirExists. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'dirExists' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if os.IsNotExist(err) {
		return FALSE
	}
	if err != nil {
		return FALSE
	}
	return &Bool{Value: info.IsDir()}
}

// pathExists - check if path exists (file or directory)
// Usage: pathExists(path) -> bool
func builtinPathExists(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for pathExists. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'pathExists' must be STRING, got %s", args[0].Type())
	}

	_, err := os.Stat(path.Value)
	if os.IsNotExist(err) {
		return FALSE
	}
	if err != nil {
		return FALSE
	}
	return TRUE
}

// isDir - check if path is a directory
// Usage: isDir(path) -> bool
func builtinIsDir(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isDir. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'isDir' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if err != nil {
		return FALSE
	}
	return &Bool{Value: info.IsDir()}
}

// isFile - check if path is a regular file
// Usage: isFile(path) -> bool
func builtinIsFile(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isFile. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'isFile' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if err != nil {
		return FALSE
	}
	return &Bool{Value: !info.IsDir()}
}

// loadText - load text content from file
// Usage: loadText(path) -> string
func builtinLoadText(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for loadText. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'loadText' must be STRING, got %s", args[0].Type())
	}

	content, err := os.ReadFile(path.Value)
	if err != nil {
		return newError("loadText error: %v", err)
	}
	return NewString(string(content))
}

// saveText - save text content to file
// Usage: saveText(path, content) -> null
func builtinSaveText(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for saveText. got=%d, want=2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'saveText' must be STRING, got %s", args[0].Type())
	}

	content, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'saveText' must be STRING, got %s", args[1].Type())
	}

	err := os.WriteFile(path.Value, []byte(content.Value), 0644)
	if err != nil {
		return newError("saveText error: %v", err)
	}
	return NULL
}

// saveBytes - save bytes to file
// Usage: saveBytes(path, bytes) -> null
func builtinSaveBytes(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for saveBytes. got=%d, want=2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'saveBytes' must be STRING, got %s", args[0].Type())
	}

	content, ok := args[1].(*Bytes)
	if !ok {
		return newError("second argument to 'saveBytes' must be BYTES, got %s", args[1].Type())
	}

	err := os.WriteFile(path.Value, content.Value, 0644)
	if err != nil {
		return newError("saveBytes error: %v", err)
	}
	return NULL
}

// loadBytes - load bytes from file
// Usage: loadBytes(path) -> bytes
func builtinLoadBytes(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for loadBytes. got=%d, want=1 or 2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'loadBytes' must be STRING, got %s", args[0].Type())
	}

	limit := int64(0)
	if len(args) == 2 {
		l, ok := args[1].(*Int)
		if !ok {
			return newError("second argument to 'loadBytes' must be INT, got %s", args[1].Type())
		}
		limit = l.Value
	}

	var content []byte
	var err error
	if limit > 0 {
		f, err := os.Open(path.Value)
		if err != nil {
			return newError("loadBytes error: %v", err)
		}
		defer f.Close()
		buf := make([]byte, limit)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return newError("loadBytes read error: %v", err)
		}
		content = buf[:n]
	} else {
		content, err = os.ReadFile(path.Value)
		if err != nil {
			return newError("loadBytes error: %v", err)
		}
	}

	return &Bytes{Value: content}
}

// appendText - append text content to file
// Usage: appendText(path, content) -> null
func builtinAppendText(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for appendText. got=%d, want=2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'appendText' must be STRING, got %s", args[0].Type())
	}

	content, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'appendText' must be STRING, got %s", args[1].Type())
	}

	f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return newError("appendText error: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(content.Value)
	if err != nil {
		return newError("appendText write error: %v", err)
	}
	return NULL
}

// copyFile - copy file from src to dst
// Usage: copyFile(src, dst) -> null
func builtinCopyFile(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for copyFile. got=%d, want=2", len(args))
	}

	src, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'copyFile' must be STRING, got %s", args[0].Type())
	}

	dst, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'copyFile' must be STRING, got %s", args[1].Type())
	}

	srcFile, err := os.Open(src.Value)
	if err != nil {
		return newError("copyFile open source error: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst.Value)
	if err != nil {
		return newError("copyFile create dest error: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return newError("copyFile copy error: %v", err)
	}
	return NULL
}

// renameFile - rename or move file
// Usage: renameFile(oldPath, newPath) -> null
func builtinRenameFile(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for renameFile. got=%d, want=2", len(args))
	}

	oldPath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'renameFile' must be STRING, got %s", args[0].Type())
	}

	newPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'renameFile' must be STRING, got %s", args[1].Type())
	}

	err := os.Rename(oldPath.Value, newPath.Value)
	if err != nil {
		return newError("renameFile error: %v", err)
	}
	return NULL
}

// removeFile - remove a file
// Usage: removeFile(path) -> null
func builtinRemoveFile(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for removeFile. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'removeFile' must be STRING, got %s", args[0].Type())
	}

	err := os.Remove(path.Value)
	if err != nil {
		return newError("removeFile error: %v", err)
	}
	return NULL
}

// removeDir - remove a directory
// Usage: removeDir(path) -> null
func builtinRemoveDir(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for removeDir. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'removeDir' must be STRING, got %s", args[0].Type())
	}

	err := os.RemoveAll(path.Value)
	if err != nil {
		return newError("removeDir error: %v", err)
	}
	return NULL
}

// getFileList - get list of files in directory
// Usage: getFileList(path) -> array
//
//	getFileList(path, pattern) -> array (glob pattern)
//	getFileList(path, "-recursive" or "-r") -> array (recursive)
func builtinGetFileList(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for getFileList. got=%d, want=1 or 2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileList' must be STRING, got %s", args[0].Type())
	}

	pattern := ""
	recursive := false
	if len(args) == 2 {
		if opt, ok := args[1].(*String); ok {
			if opt.Value == "-recursive" || opt.Value == "-r" {
				recursive = true
			} else {
				// Treat as glob pattern
				pattern = opt.Value
			}
		}
	}

	var files []Object

	// If pattern contains path separator, use filepath.Glob directly
	if pattern != "" && strings.Contains(pattern, string(filepath.Separator)) {
		// Full path glob
		matches, err := filepath.Glob(filepath.Join(path.Value, pattern))
		if err != nil {
			return newError("getFileList glob error: %v", err)
		}
		for _, m := range matches {
			files = append(files, NewString(m))
		}
		return NewArray(files)
	}

	// If pattern is provided without path separator, filter the directory
	if pattern != "" {
		entries, err := os.ReadDir(path.Value)
		if err != nil {
			return newError("getFileList error: %v", err)
		}
		for _, entry := range entries {
			matched, err := filepath.Match(pattern, entry.Name())
			if err != nil {
				return newError("getFileList pattern error: %v", err)
			}
			if matched {
				files = append(files, NewString(entry.Name()))
			}
		}
		return NewArray(files)
	}

	// No pattern, just list directory
	if recursive {
		err := filepath.Walk(path.Value, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			files = append(files, NewString(filePath))
			return nil
		})
		if err != nil {
			return newError("getFileList error: %v", err)
		}
	} else {
		entries, err := os.ReadDir(path.Value)
		if err != nil {
			return newError("getFileList error: %v", err)
		}
		for _, entry := range entries {
			files = append(files, NewString(entry.Name()))
		}
	}

	return NewArray(files)
}

// joinPath - join path components
// Usage: joinPath(path1, path2, ...) -> string
func builtinJoinPath(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for joinPath. got=%d, want>=1", len(args))
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		s, ok := arg.(*String)
		if !ok {
			return newError("argument %d to 'joinPath' must be STRING, got %s", i, arg.Type())
		}
		parts[i] = s.Value
	}

	return NewString(filepath.Join(parts...))
}

// getCurDir - get current working directory
// Usage: getCurDir() -> string
func builtinGetCurDir(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getCurDir. got=%d, want=0", len(args))
	}

	dir, err := os.Getwd()
	if err != nil {
		return newError("getCurDir error: %v", err)
	}
	return NewString(dir)
}

// getHomeDir - get user's home directory
// Usage: getHomeDir() -> string
func builtinGetHomeDir(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getHomeDir. got=%d, want=0", len(args))
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return newError("getHomeDir error: %v", err)
	}
	return NewString(dir)
}

// getTempDir - get temporary directory
// Usage: getTempDir() -> string
func builtinGetTempDir(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getTempDir. got=%d, want=0", len(args))
	}

	return NewString(os.TempDir())
}

// ensureMakeDirs - ensure directory exists (mkdir -p)
// Usage: ensureMakeDirs(path) -> null
func builtinEnsureMakeDirs(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for ensureMakeDirs. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'ensureMakeDirs' must be STRING, got %s", args[0].Type())
	}

	err := os.MkdirAll(path.Value, 0755)
	if err != nil {
		return newError("ensureMakeDirs error: %v", err)
	}
	return NULL
}

// getFileExt - get file extension
// Usage: getFileExt(path) -> string
func builtinGetFileExt(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getFileExt. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileExt' must be STRING, got %s", args[0].Type())
	}

	return NewString(filepath.Ext(path.Value))
}

// extractFileDir - extract directory from file path
// Usage: extractFileDir(path) -> string
func builtinExtractFileDir(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for extractFileDir. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'extractFileDir' must be STRING, got %s", args[0].Type())
	}

	return NewString(filepath.Dir(path.Value))
}

// extractFileName - extract file name from path
// Usage: extractFileName(path) -> string
func builtinExtractFileName(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for extractFileName. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'extractFileName' must be STRING, got %s", args[0].Type())
	}

	return NewString(filepath.Base(path.Value))
}

// getFileAbs - get absolute path
// Usage: getFileAbs(path) -> string
func builtinGetFileAbs(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getFileAbs. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileAbs' must be STRING, got %s", args[0].Type())
	}

	abs, err := filepath.Abs(path.Value)
	if err != nil {
		return newError("getFileAbs error: %v", err)
	}
	return NewString(abs)
}

// getFileRel - get relative path
// Usage: getFileRel(basePath, targetPath) -> string
func builtinGetFileRel(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for getFileRel. got=%d, want=2", len(args))
	}

	basePath, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getFileRel' must be STRING, got %s", args[0].Type())
	}

	targetPath, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'getFileRel' must be STRING, got %s", args[1].Type())
	}

	rel, err := filepath.Rel(basePath.Value, targetPath.Value)
	if err != nil {
		return newError("getFileRel error: %v", err)
	}
	return NewString(rel)
}

// getFileInfo - get file information
// Usage: getFileInfo(path) -> map
func builtinGetFileInfo(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getFileInfo. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileInfo' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if err != nil {
		return newError("getFileInfo error: %v", err)
	}

	pairs := make(map[HashKey]MapPair)

	pairs[NewString("name").HashKey()] = MapPair{
		Key:   NewString("name"),
		Value: NewString(info.Name()),
	}

	pairs[NewString("size").HashKey()] = MapPair{
		Key:   NewString("size"),
		Value: NewInt(info.Size()),
	}

	pairs[NewString("isDir").HashKey()] = MapPair{
		Key:   NewString("isDir"),
		Value: &Bool{Value: info.IsDir()},
	}

	pairs[NewString("modTime").HashKey()] = MapPair{
		Key:   NewString("modTime"),
		Value: NewInt(info.ModTime().Unix()),
	}

	pairs[NewString("mode").HashKey()] = MapPair{
		Key:   NewString("mode"),
		Value: NewString(info.Mode().String()),
	}

	return NewMap(pairs)
}

// loadLines - load lines from file
// Usage: loadLines(path) -> array or loadLines(path, limit) -> array
func builtinLoadLines(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for loadLines. got=%d, want=1 or 2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'loadLines' must be STRING, got %s", args[0].Type())
	}

	limit := int64(0)
	if len(args) == 2 {
		l, ok := args[1].(*Int)
		if !ok {
			return newError("second argument to 'loadLines' must be INT, got %s", args[1].Type())
		}
		limit = l.Value
	}

	content, err := os.ReadFile(path.Value)
	if err != nil {
		return newError("loadLines error: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Remove trailing empty line if file ends with newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if limit > 0 && int64(len(lines)) > limit {
		lines = lines[:limit]
	}

	elements := make([]Object, len(lines))
	for i, line := range lines {
		elements[i] = NewString(line)
	}

	return NewArray(elements)
}

// getFileSize - get file size in bytes
// Usage: getFileSize(path) -> int
func builtinGetFileSize(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for getFileSize. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'getFileSize' must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if err != nil {
		return newError("getFileSize error: %v", err)
	}

	return NewInt(info.Size())
}

// copyPath - copy directory or file recursively
// Usage: copyPath(src, dst) -> null
func builtinCopyPath(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for copyPath. got=%d, want=2", len(args))
	}

	src, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'copyPath' must be STRING, got %s", args[0].Type())
	}

	dst, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'copyPath' must be STRING, got %s", args[1].Type())
	}

	// Get source info
	srcInfo, err := os.Stat(src.Value)
	if err != nil {
		return newError("copyPath error: cannot stat source: %v", err)
	}

	if srcInfo.IsDir() {
		// Copy directory recursively
		err = copyDir(src.Value, dst.Value)
		if err != nil {
			return newError("copyPath error: %v", err)
		}
	} else {
		// Copy single file
		err = copySingleFile(src.Value, dst.Value)
		if err != nil {
			return newError("copyPath error: %v", err)
		}
	}

	return NULL
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	// Create destination directory
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copySingleFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copySingleFile copies a single file
func copySingleFile(src, dst string) error {
	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
