// pkg/stdlib/os.go
// OS utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/os",
		Exports: map[string]objects.Object{
			// Path operations
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

			"base": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base() requires a string argument")
				}
				return String(filepath.Base(s.Value))
			}),

			"dir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("dir() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("dir() requires a string argument")
				}
				return String(filepath.Dir(s.Value))
			}),

			"ext": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("ext() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("ext() requires a string argument")
				}
				return String(filepath.Ext(s.Value))
			}),

			"abs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("abs() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("abs() requires a string argument")
				}
				abs, err := filepath.Abs(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(abs)
			}),

			"clean": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("clean() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("clean() requires a string argument")
				}
				return String(filepath.Clean(s.Value))
			}),

			"isAbs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isAbs() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isAbs() requires a string argument")
				}
				return Bool(filepath.IsAbs(s.Value))
			}),

			// File info
			"stat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stat() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("stat() requires a string argument")
				}
				info, err := os.Stat(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Array(
					String(info.Name()),
					Int(info.Size()),
					Bool(info.IsDir()),
					String(info.ModTime().Format("2006-01-02 15:04:05")),
				)
			}),

			"size": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("size() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("size() requires a string argument")
				}
				info, err := os.Stat(s.Value)
				if err != nil {
					return Int(-1)
				}
				return Int(info.Size())
			}),

			"isDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDir() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isDir() requires a string argument")
				}
				info, err := os.Stat(s.Value)
				if err != nil {
					return Bool(false)
				}
				return Bool(info.IsDir())
			}),

			"isFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isFile() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isFile() requires a string argument")
				}
				info, err := os.Stat(s.Value)
				if err != nil {
					return Bool(false)
				}
				return Bool(!info.IsDir())
			}),

			// Directory operations
			"listDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("listDir() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("listDir() requires a string argument")
				}
				entries, err := os.ReadDir(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				result := make([]objects.Object, len(entries))
				for i, entry := range entries {
					result[i] = String(entry.Name())
				}
				return Array(result...)
			}),

			"listDirFull": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("listDirFull() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("listDirFull() requires a string argument")
				}
				entries, err := os.ReadDir(s.Value)
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
					)
				}
				return Array(result...)
			}),

			"walk": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("walk() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("walk() requires a string argument")
				}
				var files []objects.Object
				filepath.Walk(s.Value, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					files = append(files, String(path))
					return nil
				})
				return Array(files...)
			}),

			// Process execution
			"exec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("exec() takes at least 1 argument")
				}
				cmdStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("exec() requires a string command")
				}
				// Split command into name and args
				parts := strings.Fields(cmdStr.Value)
				if len(parts) == 0 {
					return Error("exec() requires a non-empty command")
				}
				cmd := exec.Command(parts[0], parts[1:]...)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return Array(String(string(output)), Int(1), String(err.Error()))
				}
				return Array(String(string(output)), Int(0), String(""))
			}),

			"shell": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("shell() takes exactly 1 argument")
				}
				cmdStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("shell() requires a string command")
				}
				var cmd *exec.Cmd
				if runtime.GOOS == "windows" {
					cmd = exec.Command("cmd", "/c", cmdStr.Value)
				} else {
					cmd = exec.Command("sh", "-c", cmdStr.Value)
				}
				output, err := cmd.CombinedOutput()
				if err != nil {
					return Array(String(string(output)), Int(1), String(err.Error()))
				}
				return Array(String(string(output)), Int(0), String(""))
			}),

			// System info
			"hostname": BuiltinFunc(func(args ...objects.Object) objects.Object {
				name, err := os.Hostname()
				if err != nil {
					return Error(err.Error())
				}
				return String(name)
			}),

			"platform": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return String(runtime.GOOS)
			}),

			"arch": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return String(runtime.GOARCH)
			}),

			"cpus": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(runtime.NumCPU()))
			}),

			"home": BuiltinFunc(func(args ...objects.Object) objects.Object {
				home, err := os.UserHomeDir()
				if err != nil {
					return Error(err.Error())
				}
				return String(home)
			}),

			"temp": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return String(os.TempDir())
			}),

			// File permissions
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

			"rename": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("rename() takes exactly 2 arguments")
				}
				old, ok := args[0].(*objects.String)
				if !ok {
					return Error("rename() requires string arguments")
				}
				newPath, ok := args[1].(*objects.String)
				if !ok {
					return Error("rename() requires string arguments")
				}
				err := os.Rename(old.Value, newPath.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

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
				data, err := os.ReadFile(src.Value)
				if err != nil {
					return Error(err.Error())
				}
				err = os.WriteFile(dst.Value, data, 0644)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			// Temp files
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

			// User info
			"userInfo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				user, err := os.UserHomeDir()
				if err != nil {
					return Error(err.Error())
				}
				uid := fmt.Sprintf("%d", os.Getuid())
				gid := fmt.Sprintf("%d", os.Getgid())
				pid := fmt.Sprintf("%d", os.Getpid())
				return Array(
					String(user),
					String(uid),
					String(gid),
					String(pid),
				)
			}),
		},
	})
}
