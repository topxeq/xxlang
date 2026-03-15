// pkg/stdlib/env.go
// Environment variables and configuration utilities for the Xxlang standard library.
package stdlib

import (
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

// scriptArgs stores the script-specific arguments (after -- separator)
var scriptArgs []string

// SetScriptArgs sets the script-specific arguments for scripts to access
func SetScriptArgs(args []string) {
	scriptArgs = args
}

func init() {
	Register(&Module{
		Name: "std/env",
		Exports: map[string]objects.Object{
			// Get environment variable
			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("get() takes exactly 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("get() requires a string argument")
				}
				val := os.Getenv(key.Value)
				return String(val)
			}),

			// Get environment variable with default
			"getOr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("getOr() takes exactly 2 arguments")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("getOr() requires a string as first argument")
				}
				defaultVal := args[1]
				val := os.Getenv(key.Value)
				if val == "" {
					return defaultVal
				}
				return String(val)
			}),

			// Set environment variable
			"set": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("set() takes exactly 2 arguments")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("set() requires a string as first argument")
				}
				val, ok := args[1].(*objects.String)
				if !ok {
					return Error("set() requires a string as second argument")
				}
				os.Setenv(key.Value, val.Value)
				return Null()
			}),

			// Unset environment variable
			"unset": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unset() takes exactly 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("unset() requires a string argument")
				}
				os.Unsetenv(key.Value)
				return Null()
			}),

			// Check if environment variable exists
			"has": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("has() takes exactly 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("has() requires a string argument")
				}
				_, exists := os.LookupEnv(key.Value)
				return Bool(exists)
			}),

			// Get all environment variables
			"all": BuiltinFunc(func(args ...objects.Object) objects.Object {
				env := os.Environ()
				result := []objects.Object{}
				for _, e := range env {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						result = append(result, Array(String(parts[0]), String(parts[1])))
					}
				}
				return Array(result...)
			}),

			// Get all environment variables as map
			"map": BuiltinFunc(func(args ...objects.Object) objects.Object {
				pairs := make(map[objects.HashKey]objects.MapPair)
				for _, e := range os.Environ() {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						key := String(parts[0])
						pairs[key.HashKey()] = objects.MapPair{
							Key:   key,
							Value: String(parts[1]),
						}
					}
				}
				return &objects.Map{Pairs: pairs}
			}),

			// Get PATH as array
			"path": BuiltinFunc(func(args ...objects.Object) objects.Object {
				path := os.Getenv("PATH")
				if path == "" {
					return Array()
				}
				paths := strings.Split(path, string(os.PathListSeparator))
				result := make([]objects.Object, len(paths))
				for i, p := range paths {
					result[i] = String(p)
				}
				return Array(result...)
			}),

			// Expand environment variables in string
			"expand": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("expand() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("expand() requires a string argument")
				}
				expanded := os.ExpandEnv(s.Value)
				return String(expanded)
			}),

			// Get current working directory
			"cwd": BuiltinFunc(func(args ...objects.Object) objects.Object {
				dir, err := os.Getwd()
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			// Change working directory
			"cd": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("cd() takes exactly 1 argument")
				}
				dir, ok := args[0].(*objects.String)
				if !ok {
					return Error("cd() requires a string argument")
				}
				err := os.Chdir(dir.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			// Get process ID
			"pid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(os.Getpid()))
			}),

			// Get parent process ID
			"ppid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(os.Getppid()))
			}),

			// Exit program
			"exit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				code := 0
				if len(args) > 0 {
					n, ok := args[0].(*objects.Int)
					if ok {
						code = int(n.Value)
					}
				}
				os.Exit(code)
				return Null()
			}),

			// Get arguments (command line args)
			"args": BuiltinFunc(func(args ...objects.Object) objects.Object {
				cmdArgs := os.Args
				result := make([]objects.Object, len(cmdArgs))
				for i, arg := range cmdArgs {
					result[i] = String(arg)
				}
				return Array(result...)
			}),

			// Get script-specific arguments (after -- separator)
			"scriptArgs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				result := make([]objects.Object, len(scriptArgs))
				for i, arg := range scriptArgs {
					result[i] = String(arg)
				}
				return Array(result...)
			}),

			// Get mixed arguments: script args if -- exists, otherwise all args
			"mixArgs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				// If scriptArgs has content (-- was used), return those
				if len(scriptArgs) > 0 {
					result := make([]objects.Object, len(scriptArgs))
					for i, arg := range scriptArgs {
						result[i] = String(arg)
					}
					return Array(result...)
				}
				// Otherwise return all args
				cmdArgs := os.Args
				result := make([]objects.Object, len(cmdArgs))
				for i, arg := range cmdArgs {
					result[i] = String(arg)
				}
				return Array(result...)
			}),

			// Get executable path
			"exe": BuiltinFunc(func(args ...objects.Object) objects.Object {
				exe, err := os.Executable()
				if err != nil {
					return Error(err.Error())
				}
				return String(exe)
			}),

			// Get user cache directory
			"cacheDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				dir, err := os.UserCacheDir()
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			// Get user config directory
			"configDir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				dir, err := os.UserConfigDir()
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			// Lookup environment variable
			"lookup": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("lookup() takes exactly 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("lookup() requires a string argument")
				}
				val, exists := os.LookupEnv(key.Value)
				return Array(Bool(exists), String(val))
			}),

			// Get integer environment variable
			"getInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getInt() takes at least 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("getInt() requires a string as first argument")
				}
				val := os.Getenv(key.Value)
				if val == "" {
					if len(args) > 1 {
						return args[1]
					}
					return Int(0)
				}
				var result int64
				for i, c := range val {
					if c >= '0' && c <= '9' {
						result = result*10 + int64(c-'0')
					} else if c == '-' && i == 0 {
						continue
					} else {
						if len(args) > 1 {
							return args[1]
						}
						return Int(0)
					}
				}
				if len(val) > 0 && val[0] == '-' {
					result = -result
				}
				return Int(result)
			}),

			// Get boolean environment variable
			"getBool": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getBool() takes at least 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("getBool() requires a string as first argument")
				}
				val := os.Getenv(key.Value)
				if val == "" {
					if len(args) > 1 {
						return args[1]
					}
					return Bool(false)
				}
				val = strings.ToLower(val)
				return Bool(val == "true" || val == "1" || val == "yes" || val == "on")
			}),

			// Clear all environment variables
			"clear": BuiltinFunc(func(args ...objects.Object) objects.Object {
				os.Clearenv()
				return Null()
			}),

			// Get stdin/out/err info
			"streams": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Array(
					Bool(os.Stdin != nil),
					Bool(os.Stdout != nil),
					Bool(os.Stderr != nil),
				)
			}),
		},
	})
}
