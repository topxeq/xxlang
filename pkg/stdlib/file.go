// pkg/stdlib/file.go
// File module for streaming file operations in Xxlang standard library.
package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "file",
		Exports: map[string]objects.Object{
			// Open modes
			"MODE_READ":   String("r"),
			"MODE_WRITE":  String("w"),
			"MODE_APPEND": String("a"),
			"MODE_RW":     String("rw"),
			"MODE_RWPLUS": String("rw+"),

			// Seek constants
			"SEEK_START":   Int(0),
			"SEEK_CURRENT": Int(1),
			"SEEK_END":     Int(2),

			// Lock types
			"LOCK_SHARED":    Int(1),
			"LOCK_EXCLUSIVE": Int(2),

			// open opens a file for streaming operations.
			// Usage: f = file.open(path, mode)
			// mode: "r" = read, "w" = write (truncate), "a" = append, "rw" = read/write, "rw+" = read/write/create
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("open() takes 1 or 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("open() requires a string path")
				}

				mode := objects.ModeRead // default: read-only
				if len(args) == 2 {
					m, ok := args[1].(*objects.String)
					if !ok {
						return Error("open() mode must be a string")
					}
					mode = objects.FileMode(m.Value)
				}

				var flag int
				var perm os.FileMode = 0644

				switch mode {
				case objects.ModeRead:
					flag = os.O_RDONLY
				case objects.ModeWrite:
					flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
				case objects.ModeAppend:
					flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
				case objects.ModeRW:
					flag = os.O_RDWR
				case objects.ModeRWPlus:
					flag = os.O_RDWR | os.O_CREATE
				default:
					return Error(fmt.Sprintf("invalid file mode: %s", mode))
				}

				handle, err := os.OpenFile(path.Value, flag, perm)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFile(handle, path.Value, mode)
			}),

			// create creates a new file for writing (truncates if exists).
			// Usage: f = file.create(path)
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("create() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("create() requires a string path")
				}

				handle, err := os.Create(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFile(handle, path.Value, objects.ModeWrite)
			}),

			// openRead opens a file for reading.
			// Usage: f = file.openRead(path)
			"openRead": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("openRead() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("openRead() requires a string path")
				}

				handle, err := os.Open(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFile(handle, path.Value, objects.ModeRead)
			}),

			// openWrite opens a file for writing (creates if not exists, truncates if exists).
			// Usage: f = file.openWrite(path)
			"openWrite": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("openWrite() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("openWrite() requires a string path")
				}

				handle, err := os.Create(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFile(handle, path.Value, objects.ModeWrite)
			}),

			// openAppend opens a file for appending.
			// Usage: f = file.openAppend(path)
			"openAppend": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("openAppend() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("openAppend() requires a string path")
				}

				handle, err := os.OpenFile(path.Value, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFile(handle, path.Value, objects.ModeAppend)
			}),

			// readLines reads all lines from a file and returns as array.
			// Usage: lines = file.readLines(path)
			"readLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readLines() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("readLines() requires a string path")
				}

				file, err := os.Open(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				var lines []objects.Object
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					lines = append(lines, String(scanner.Text()))
				}

				if err := scanner.Err(); err != nil {
					return Error(err.Error())
				}

				return Array(lines...)
			}),

			// writeLines writes an array of strings to a file, one per line.
			// Usage: file.writeLines(path, lines)
			"writeLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeLines() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeLines() requires a string path")
				}
				lines, ok := args[1].(*objects.Array)
				if !ok {
					return Error("writeLines() requires an array of strings")
				}

				file, err := os.Create(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				writer := bufio.NewWriter(file)
				for _, line := range lines.Elements {
					s, ok := line.(*objects.String)
					if !ok {
						return Error("writeLines() requires an array of strings")
					}
					writer.WriteString(s.Value)
					writer.WriteString("\n")
				}

				if err := writer.Flush(); err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// readAll reads entire file content as string.
			// This is a convenience function that wraps io.readFile.
			// Usage: content = file.readAll(path)
			"readAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readAll() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("readAll() requires a string path")
				}

				content, err := os.ReadFile(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return String(string(content))
			}),

			// writeAll writes a string to a file (creates or truncates).
			// This is a convenience function that wraps io.writeFile.
			// Usage: file.writeAll(path, content)
			"writeAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeAll() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeAll() requires a string path")
				}
				content, ok := args[1].(*objects.String)
				if !ok {
					return Error("writeAll() requires a string content")
				}

				err := os.WriteFile(path.Value, []byte(content.Value), 0644)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// appendAll appends a string to a file (creates if not exists).
			// This is a convenience function that wraps io.appendFile.
			// Usage: file.appendAll(path, content)
			"appendAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("appendAll() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("appendAll() requires a string path")
				}
				content, ok := args[1].(*objects.String)
				if !ok {
					return Error("appendAll() requires a string content")
				}

				file, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				_, err = file.WriteString(content.Value)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// copy copies a file from src to dst.
			// Usage: file.copy(src, dst)
			"copy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("copy() takes exactly 2 arguments")
				}
				src, ok := args[0].(*objects.String)
				if !ok {
					return Error("copy() requires string arguments")
				}
				dst, ok := args[1].(*objects.String)
				if !ok {
					return Error("copy() requires string arguments")
				}

				// Open source file
				srcFile, err := os.Open(src.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer srcFile.Close()

				// Create destination file
				dstFile, err := os.Create(dst.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer dstFile.Close()

				// Copy content
				_, err = io.Copy(dstFile, srcFile)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// move moves (renames) a file from src to dst.
			// Usage: file.move(src, dst)
			"move": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("move() takes exactly 2 arguments")
				}
				src, ok := args[0].(*objects.String)
				if !ok {
					return Error("move() requires string arguments")
				}
				dst, ok := args[1].(*objects.String)
				if !ok {
					return Error("move() requires string arguments")
				}

				err := os.Rename(src.Value, dst.Value)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// remove removes a file.
			// Usage: file.remove(path)
			"remove": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("remove() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("remove() requires a string path")
				}

				err := os.Remove(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// removeDir removes a directory and all its contents.
			// Usage: file.removeDir(path)
			"removeDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("removeDir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("removeDir() requires a string path")
				}

				err := os.RemoveAll(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// exists checks if a file or directory exists.
			// Usage: file.exists(path)
			"exists": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("exists() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("exists() requires a string path")
				}

				_, err := os.Stat(path.Value)
				return Bool(err == nil)
			}),

			// isFile checks if path is a regular file.
			// Usage: file.isFile(path)
			"isFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isFile() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("isFile() requires a string path")
				}

				info, err := os.Stat(path.Value)
				if err != nil {
					return Bool(false)
				}

				return Bool(!info.IsDir())
			}),

			// isDir checks if path is a directory.
			// Usage: file.isDir(path)
			"isDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("isDir() requires a string path")
				}

				info, err := os.Stat(path.Value)
				if err != nil {
					return Bool(false)
				}

				return Bool(info.IsDir())
			}),

			// stat returns file information as a FileInfo object.
			// Usage: info = file.stat(path)
			"stat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stat() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("stat() requires a string path")
				}

				info, err := os.Stat(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return objects.NewFileInfo(info, path.Value)
			}),

			// size returns the file size in bytes.
			// Usage: size = file.size(path)
			"size": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("size() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("size() requires a string path")
				}

				info, err := os.Stat(path.Value)
				if err != nil {
					return Int(-1)
				}

				return Int(info.Size())
			}),

			// modTime returns the file modification time as a formatted string.
			// Usage: time = file.modTime(path)
			"modTime": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("modTime() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("modTime() requires a string path")
				}

				info, err := os.Stat(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return String(info.ModTime().Format("2006-01-02 15:04:05"))
			}),

			// mkdir creates a directory (and parent directories if needed).
			// Usage: file.mkdir(path)
			"mkdir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("mkdir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("mkdir() requires a string path")
				}

				err := os.MkdirAll(path.Value, 0755)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// listDir lists directory entries.
			// Usage: entries = file.listDir(path)
			"listDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("listDir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("listDir() requires a string path")
				}

				entries, err := os.ReadDir(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				result := make([]objects.Object, len(entries))
				for i, entry := range entries {
					result[i] = String(entry.Name())
				}

				return Array(result...)
			}),

			// listDirFull lists directory entries with full information.
			// Usage: entries = file.listDirFull(path)
			// Returns array of [name, size, isDir, modTime]
			"listDirFull": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("listDirFull() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("listDirFull() requires a string path")
				}

				entries, err := os.ReadDir(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				result := make([]objects.Object, len(entries))
				for i, entry := range entries {
					info, _ := entry.Info()
					result[i] = Array(
						String(entry.Name()),
						Int(info.Size()),
						Bool(entry.IsDir()),
						String(info.ModTime().Format("2006-01-02 15:04:05")),
					)
				}

				return Array(result...)
			}),

			// glob returns files matching a pattern.
			// Usage: matches = file.glob(pattern)
			"glob": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("glob() takes exactly 1 argument")
				}
				pattern, ok := args[0].(*objects.String)
				if !ok {
					return Error("glob() requires a string pattern")
				}

				matches, err := filepath.Glob(pattern.Value)
				if err != nil {
					return Error(err.Error())
				}

				result := make([]objects.Object, len(matches))
				for i, m := range matches {
					result[i] = String(m)
				}

				return Array(result...)
			}),

			// abs returns the absolute path.
			// Usage: absPath = file.abs(path)
			"abs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("abs() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("abs() requires a string path")
				}

				abs, err := filepath.Abs(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return String(abs)
			}),

			// base returns the base name of a file path.
			// Usage: name = file.base(path)
			"base": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("base() requires a string path")
				}

				return String(filepath.Base(path.Value))
			}),

			// dir returns the directory part of a file path.
			// Usage: dirPath = file.dir(path)
			"dir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("dir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("dir() requires a string path")
				}

				return String(filepath.Dir(path.Value))
			}),

			// ext returns the file extension.
			// Usage: extension = file.ext(path)
			"ext": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("ext() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("ext() requires a string path")
				}

				return String(filepath.Ext(path.Value))
			}),

			// join joins path elements.
			// Usage: path = file.join(elem1, elem2, ...)
			"join": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("join() takes at least 2 arguments")
				}

				parts := make([]string, len(args))
				for i, arg := range args {
					s, ok := arg.(*objects.String)
					if !ok {
						return Error("join() requires string arguments")
					}
					parts[i] = s.Value
				}

				return String(filepath.Join(parts...))
			}),

			// cwd returns the current working directory.
			// Usage: dir = file.cwd()
			"cwd": BuiltinFunc(func(args ...objects.Object) objects.Object {
				dir, err := os.Getwd()
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			// chdir changes the current working directory.
			// Usage: file.chdir(path)
			"chdir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("chdir() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("chdir() requires a string path")
				}

				err := os.Chdir(path.Value)
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// tempFile creates a temporary file and returns its path.
			// Usage: path = file.tempFile([pattern])
			"tempFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				pattern := "xxlang-*.tmp"
				if len(args) > 0 {
					s, ok := args[0].(*objects.String)
					if ok {
						pattern = s.Value
					}
				}

				file, err := os.CreateTemp("", pattern)
				if err != nil {
					return Error(err.Error())
				}
				name := file.Name()
				file.Close()
				return String(name)
			}),

			// tempDir creates a temporary directory and returns its path.
			// Usage: path = file.tempDir([pattern])
			"tempDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				pattern := "xxlang-*"
				if len(args) > 0 {
					s, ok := args[0].(*objects.String)
					if ok {
						pattern = s.Value
					}
				}

				dir, err := os.MkdirTemp("", pattern)
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			// chmod changes file permissions.
			// Usage: file.chmod(path, mode)
			"chmod": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("chmod() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("chmod() requires a string path")
				}
				mode, ok := args[1].(*objects.Int)
				if !ok {
					return Error("chmod() requires an integer mode")
				}

				err := os.Chmod(path.Value, os.FileMode(mode.Value))
				if err != nil {
					return Error(err.Error())
				}

				return Null()
			}),
		},
	})
}
