// pkg/stdlib/env.go
// Environment variables and configuration utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"os"
	"strconv"
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
		Name: "env",
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

			// parseFlags parses command-line arguments according to a flag specification.
			// This is a convenience function that replaces the common pattern of manually
			// iterating over argsG to extract --key=value and --key value pairs.
			//
			// Usage:
			//   var opts = env.parseFlags([
			//       {"name": "url", "short": "u", "default": "http://localhost", "desc": "Target URL"},
			//       {"name": "user", "short": "U", "default": "admin", "desc": "Username"},
			//       {"name": "verbose", "short": "v", "type": "bool", "desc": "Verbose output"},
			//       {"name": "timeout", "short": "t", "type": "int", "default": 10000, "desc": "Timeout ms"},
			//       {"name": "help", "short": "h", "type": "bool", "desc": "Show help"},
			//   ])
			//
			// The returned map contains:
			//   - Each flag name as a key with its parsed value
			//   - "_args": array of remaining positional arguments (not consumed by any flag)
			//   - "_help": pre-formatted help text string (if any flag has a "desc" field)
			//
			// Supported flag types: "string" (default), "int", "float", "bool".
			// Boolean flags are set to true when present (no value needed).
			// If no args parameter is provided, uses env.scriptArgs().
			"parseFlags": BuiltinFunc(func(fnArgs ...objects.Object) objects.Object {
				if len(fnArgs) < 1 {
					return Error("parseFlags() takes at least 1 argument: specs array")
				}

				// Extract the specs array
				specsArr, ok := fnArgs[0].(*objects.Array)
				if !ok {
					return Error("parseFlags() first argument must be an array of flag specs")
				}

				// Determine the args to parse:
				// - If a second argument is provided (string array), use that
				// - Otherwise, use scriptArgs (the args after -- separator)
				var argStrings []string
				if len(fnArgs) >= 2 {
					argsArr, ok := fnArgs[1].(*objects.Array)
					if !ok {
						return Error("parseFlags() second argument must be a string array")
					}
					for _, elem := range argsArr.Elements {
						if s, ok := elem.(*objects.String); ok {
							argStrings = append(argStrings, s.Value)
						}
					}
				} else {
					// Default: use scriptArgs
					argStrings = scriptArgs
				}

				// Parse the specs array into flag definitions
				type flagSpec struct {
					name       string
					short      string
					flagType   string // "string", "int", "float", "bool"
					defaultVal objects.Object
					desc       string
				}

				specs := make([]flagSpec, 0, len(specsArr.Elements))
				helpLines := make([]string, 0)

				for i, elem := range specsArr.Elements {
					m, ok := elem.(*objects.Map)
					if !ok {
						return Error(fmt.Sprintf("parseFlags() spec[%d] must be a map", i))
					}

					spec := flagSpec{flagType: "string"}

					// name (required)
					nameKey := objects.NewString("name")
					if pair, found := m.Pairs[nameKey.HashKey()]; found {
						if s, ok := pair.Value.(*objects.String); ok {
							spec.name = s.Value
						}
					}
					if spec.name == "" {
						return Error(fmt.Sprintf("parseFlags() spec[%d] missing 'name' field", i))
					}

					// short (optional)
					shortKey := objects.NewString("short")
					if pair, found := m.Pairs[shortKey.HashKey()]; found {
						if s, ok := pair.Value.(*objects.String); ok {
							spec.short = s.Value
						}
					}

					// type (optional, default "string")
					typeKey := objects.NewString("type")
					if pair, found := m.Pairs[typeKey.HashKey()]; found {
						if s, ok := pair.Value.(*objects.String); ok {
							spec.flagType = s.Value
						}
					}

					// default (optional)
					defaultKey := objects.NewString("default")
					if pair, found := m.Pairs[defaultKey.HashKey()]; found {
						spec.defaultVal = pair.Value
					} else {
						// Set type-appropriate zero value
						switch spec.flagType {
						case "bool":
							spec.defaultVal = Bool(false)
						case "int":
							spec.defaultVal = Int(0)
						case "float":
							spec.defaultVal = Float(0)
						default:
							spec.defaultVal = String("")
						}
					}

					// desc (optional)
					descKey := objects.NewString("desc")
					if pair, found := m.Pairs[descKey.HashKey()]; found {
						if s, ok := pair.Value.(*objects.String); ok {
							spec.desc = s.Value
						}
					}

					specs = append(specs, spec)

					// Build help line
					if spec.desc != "" {
						shortPart := ""
						if spec.short != "" {
							shortPart = fmt.Sprintf("-%s, ", spec.short)
						}
						typePart := ""
						if spec.flagType != "bool" {
							typePart = "=<value>"
						}
						defaultPart := ""
						if spec.defaultVal != nil {
							switch val := spec.defaultVal.(type) {
							case *objects.String:
								if val.Value != "" {
									defaultPart = fmt.Sprintf(" (default: %s)", val.Value)
								}
							case *objects.Int:
								defaultPart = fmt.Sprintf(" (default: %d)", val.Value)
							case *objects.Float:
								defaultPart = fmt.Sprintf(" (default: %v)", val.Value)
							case *objects.Bool:
								if val.Value {
									defaultPart = " (default: true)"
								}
							}
						}
						helpLines = append(helpLines, fmt.Sprintf("  %s--%s%s  %s%s", shortPart, spec.name, typePart, spec.desc, defaultPart))
					}
				}

				// Build lookup maps: long name -> spec index, short name -> spec index
				longMap := make(map[string]int)
				shortMap := make(map[string]int)
				for i, spec := range specs {
					longMap[spec.name] = i
					if spec.short != "" {
						shortMap[spec.short] = i
					}
				}

				// Initialize result map with default values
				resultPairs := make(map[objects.HashKey]objects.MapPair, len(specs)+2)
				for _, spec := range specs {
					key := objects.NewString(spec.name)
					resultPairs[key.HashKey()] = objects.MapPair{Key: key, Value: spec.defaultVal}
				}

				// Parse the argument strings
				var posArgs []objects.Object
				helpRequested := false

				i := 0
				for i < len(argStrings) {
					arg := argStrings[i]

					if arg == "--" {
						// Everything after bare "--" is positional
						i++
						for i < len(argStrings) {
							posArgs = append(posArgs, String(argStrings[i]))
							i++
						}
						break
					}

					if strings.HasPrefix(arg, "--") {
						// Long flag: --key=value or --key value
						afterDash := arg[2:]
						var flagName, flagValue string
						hasExplicitValue := false

						if eqIdx := strings.IndexByte(afterDash, '='); eqIdx >= 0 {
							flagName = afterDash[:eqIdx]
							flagValue = afterDash[eqIdx+1:]
							hasExplicitValue = true
						} else {
							flagName = afterDash
						}

						specIdx, found := longMap[flagName]
						if !found {
							// Unknown flag: treat as positional argument
							posArgs = append(posArgs, String(arg))
							i++
							continue
						}

						spec := specs[specIdx]

						// Handle --help
						if flagName == "help" && spec.flagType == "bool" {
							helpRequested = true
						}

						if spec.flagType == "bool" {
							// Boolean flag: --flag means true
							if hasExplicitValue {
								resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
									Key:   objects.NewString(spec.name),
									Value: Bool(strings.ToLower(flagValue) == "true" || flagValue == "1"),
								}
							} else {
								resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
									Key:   objects.NewString(spec.name),
									Value: Bool(true),
								}
							}
						} else {
							// Value flag: --flag value or --flag=value
							if !hasExplicitValue {
								i++
								if i < len(argStrings) {
									flagValue = argStrings[i]
								} else {
									return Error(fmt.Sprintf("parseFlags() --%s requires a value", flagName))
								}
							}
							val, err := convertFlagValue(flagValue, spec.flagType)
							if err != nil {
								return Error(fmt.Sprintf("parseFlags() --%s: %s", flagName, err.Error()))
							}
							resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
								Key:   objects.NewString(spec.name),
								Value: val,
							}
						}

					} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
						// Short flag: -k value or -k=value or -k (bool)
						afterDash := arg[1:]
						var flagKey, flagValue string
						hasExplicitValue := false

						if eqIdx := strings.IndexByte(afterDash, '='); eqIdx >= 0 {
							flagKey = afterDash[:eqIdx]
							flagValue = afterDash[eqIdx+1:]
							hasExplicitValue = true
						} else {
							flagKey = afterDash
						}

						specIdx, found := shortMap[flagKey]
						if !found {
							// Unknown short flag: treat as positional
							posArgs = append(posArgs, String(arg))
							i++
							continue
						}

						spec := specs[specIdx]

						// Handle -h (help)
						if spec.name == "help" && spec.flagType == "bool" {
							helpRequested = true
						}

						if spec.flagType == "bool" {
							if hasExplicitValue {
								resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
									Key:   objects.NewString(spec.name),
									Value: Bool(strings.ToLower(flagValue) == "true" || flagValue == "1"),
								}
							} else {
								resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
									Key:   objects.NewString(spec.name),
									Value: Bool(true),
								}
							}
						} else {
							if !hasExplicitValue {
								i++
								if i < len(argStrings) {
									flagValue = argStrings[i]
								} else {
									return Error(fmt.Sprintf("parseFlags() -%s requires a value", flagKey))
								}
							}
							val, err := convertFlagValue(flagValue, spec.flagType)
							if err != nil {
								return Error(fmt.Sprintf("parseFlags() -%s: %s", flagKey, err.Error()))
							}
							resultPairs[objects.NewString(spec.name).HashKey()] = objects.MapPair{
								Key:   objects.NewString(spec.name),
								Value: val,
							}
						}

					} else {
						// Positional argument
						posArgs = append(posArgs, String(arg))
					}

					i++
				}

				// Add _args (positional arguments)
				if posArgs == nil {
					posArgs = []objects.Object{}
				}
				argsKey := objects.NewString("_args")
				resultPairs[argsKey.HashKey()] = objects.MapPair{Key: argsKey, Value: &objects.Array{Elements: posArgs}}

				// Add _help (pre-formatted help text)
				helpText := strings.Join(helpLines, "\n")
				if helpRequested {
					helpText = "Options:\n" + helpText
				}
				helpKey := objects.NewString("_help")
				resultPairs[helpKey.HashKey()] = objects.MapPair{Key: helpKey, Value: String(helpText)}

				return &objects.Map{Pairs: resultPairs}
			}),
		},
	})
}

// convertFlagValue converts a string value to the appropriate Xxlang object type.
func convertFlagValue(s string, flagType string) (objects.Object, error) {
	switch flagType {
	case "int":
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value: %s", s)
		}
		return Int(n), nil
	case "float":
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", s)
		}
		return Float(f), nil
	case "bool":
		return Bool(strings.ToLower(s) == "true" || s == "1"), nil
	default:
		return String(s), nil
	}
}
