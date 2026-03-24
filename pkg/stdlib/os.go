// pkg/stdlib/os.go
// OS utilities for the Xxlang standard library.
package stdlib

import (
	"encoding/json"
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
		Name: "os",
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

			// glob returns files matching a pattern.
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

			// split splits a path into directory and file components.
			"split": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("split() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("split() requires a string argument")
				}
				dir, file := filepath.Split(s.Value)
				return Array(String(dir), String(file))
			}),

			// relative returns a relative path from base to target.
			"relative": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("relative() takes exactly 2 arguments")
				}
				base, ok := args[0].(*objects.String)
				if !ok {
					return Error("relative() requires string arguments")
				}
				target, ok := args[1].(*objects.String)
				if !ok {
					return Error("relative() requires string arguments")
				}

				rel, err := filepath.Rel(base.Value, target.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(rel)
			}),

			// volumeName returns the leading volume name (Windows only).
			"volumeName": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("volumeName() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("volumeName() requires a string argument")
				}
				return String(filepath.VolumeName(s.Value))
			}),

			// walkDir walks a directory tree and returns all file paths.
			"walkDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("walkDir() takes exactly 1 argument")
				}
				root, ok := args[0].(*objects.String)
				if !ok {
					return Error("walkDir() requires a string argument")
				}

				var files []objects.Object
				err := filepath.WalkDir(root.Value, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					files = append(files, String(path))
					return nil
				})
				if err != nil {
					return Error(err.Error())
				}
				return Array(files...)
			}),

			// symlink creates a symbolic link.
			"symlink": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("symlink() takes exactly 2 arguments")
				}
				oldname, ok := args[0].(*objects.String)
				if !ok {
					return Error("symlink() requires string arguments")
				}
				newname, ok := args[1].(*objects.String)
				if !ok {
					return Error("symlink() requires string arguments")
				}

				err := os.Symlink(oldname.Value, newname.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			// readlink reads the target of a symbolic link.
			"readlink": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readlink() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("readlink() requires a string argument")
				}

				target, err := os.Readlink(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(target)
			}),

			// isLink checks if path is a symbolic link.
			"isLink": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isLink() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("isLink() requires a string argument")
				}

				info, err := os.Lstat(path.Value)
				if err != nil {
					return Bool(false)
				}
				return Bool(info.Mode()&os.ModeSymlink != 0)
			}),

			// lstat returns file info without following symlinks.
			"lstat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("lstat() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("lstat() requires a string argument")
				}
				info, err := os.Lstat(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Array(
					String(info.Name()),
					Int(info.Size()),
					Bool(info.IsDir()),
					String(info.ModTime().Format("2006-01-02 15:04:05")),
					Bool(info.Mode()&os.ModeSymlink != 0),
				)
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

			// Config object
			"getConfigObj": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return getConfigObjImpl()
			}),

			// Config string operations
			"getConfigStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getConfigStr() takes exactly 1 argument")
				}
				name, ok := args[0].(*objects.String)
				if !ok {
					return Error("getConfigStr() requires a string argument")
				}
				return getConfigStrImpl(name.Value)
			}),

			"setConfigStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setConfigStr() takes exactly 2 arguments")
				}
				name, ok := args[0].(*objects.String)
				if !ok {
					return Error("setConfigStr() requires string arguments")
				}
				value, ok := args[1].(*objects.String)
				if !ok {
					return Error("setConfigStr() requires string arguments")
				}
				return setConfigStrImpl(name.Value, value.Value)
			}),
		},
	})
}

// getConfigObjImpl reads the Xxlang configuration from a JSON file.
// Search path priority:
// 1. ~/.xxl/settings.json (user home directory)
// 2. /.xxl/settings.json (Linux/Unix systems)
// 3. C:\.xxl\settings.json (Windows systems)
// Returns an empty map if no config file is found.
func getConfigObjImpl() objects.Object {
	// Try user home directory first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".xxl", "settings.json")
		if cfg := tryReadConfigFile(configPath); cfg != nil {
			return cfg
		}
	}

	// Try system-wide config based on OS
	if runtime.GOOS == "windows" {
		configPath := filepath.Join("C:", ".xxl", "settings.json")
		if cfg := tryReadConfigFile(configPath); cfg != nil {
			return cfg
		}
	} else {
		// Linux/Unix systems
		configPath := filepath.Join("/", ".xxl", "settings.json")
		if cfg := tryReadConfigFile(configPath); cfg != nil {
			return cfg
		}
	}

	// No config file found, return empty map
	return &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
}

// tryReadConfigFile attempts to read and parse a config file at the given path.
// Returns nil if the file doesn't exist or cannot be parsed.
func tryReadConfigFile(path string) objects.Object {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	// Convert to Xxlang map
	pairs := make(map[objects.HashKey]objects.MapPair)
	for key, val := range cfg {
		keyObj := String(key)
		valObj := objects.GoValueToObject(val)
		if errObj, ok := valObj.(*objects.Error); ok {
			// Skip values that can't be converted
			_ = errObj
			continue
		}
		pairs[keyObj.HashKey()] = objects.MapPair{
			Key:   keyObj,
			Value: valObj,
		}
	}

	return &objects.Map{Pairs: pairs}
}

// GetConfigMap reads the Xxlang configuration and returns it as a Go map.
// This is useful for accessing config values from Go code before the VM starts.
// Search path priority:
// 1. ~/.xxl/settings.json (user home directory)
// 2. /.xxl/settings.json (Linux/Unix systems)
// 3. C:\.xxl\settings.json (Windows systems)
// Returns an empty map if no config file is found.
func GetConfigMap() map[string]interface{} {
	// Try user home directory first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".xxl", "settings.json")
		if cfg := tryReadConfigMap(configPath); cfg != nil {
			return cfg
		}
	}

	// Try system-wide config based on OS
	if runtime.GOOS == "windows" {
		configPath := filepath.Join("C:", ".xxl", "settings.json")
		if cfg := tryReadConfigMap(configPath); cfg != nil {
			return cfg
		}
	} else {
		// Linux/Unix systems
		configPath := filepath.Join("/", ".xxl", "settings.json")
		if cfg := tryReadConfigMap(configPath); cfg != nil {
			return cfg
		}
	}

	// No config file found, return empty map
	return make(map[string]interface{})
}

// tryReadConfigMap attempts to read and parse a config file at the given path.
// Returns nil if the file doesn't exist or cannot be parsed.
func tryReadConfigMap(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	return cfg
}

// getConfigStrImpl reads a config string from a .cfg file.
// Search path priority:
// 1. ~/.xxl/<name>.cfg (user home directory)
// 2. /.xxl/<name>.cfg (Linux/Unix systems)
// 3. C:\.xxl\<name>.cfg (Windows systems)
// Returns null if the file is not found.
func getConfigStrImpl(name string) objects.Object {
	// Try user home directory first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".xxl", name+".cfg")
		if data, err := os.ReadFile(configPath); err == nil {
			return String(string(data))
		}
	}

	// Try system-wide config based on OS
	if runtime.GOOS == "windows" {
		configPath := filepath.Join("C:", ".xxl", name+".cfg")
		if data, err := os.ReadFile(configPath); err == nil {
			return String(string(data))
		}
	} else {
		// Linux/Unix systems
		configPath := filepath.Join("/", ".xxl", name+".cfg")
		if data, err := os.ReadFile(configPath); err == nil {
			return String(string(data))
		}
	}

	// No config file found
	return Null()
}

// setConfigStrImpl writes a config string to a .cfg file.
// Creates the file in the user's home directory ~/.xxl/<name>.cfg
// Creates the .xxl directory if it doesn't exist.
func setConfigStrImpl(name, value string) objects.Object {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Error("cannot get user home directory: " + err.Error())
	}

	// Create .xxl directory if it doesn't exist
	xxlDir := filepath.Join(homeDir, ".xxl")
	if err := os.MkdirAll(xxlDir, 0755); err != nil {
		return Error("cannot create config directory: " + err.Error())
	}

	// Write the config file
	configPath := filepath.Join(xxlDir, name+".cfg")
	if err := os.WriteFile(configPath, []byte(value), 0644); err != nil {
		return Error("cannot write config file: " + err.Error())
	}

	return Null()
}

