// pkg/stdlib/log.go
// Logging utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

// logLevel represents log levels
type logLevel int

const (
	levelDebug logLevel = iota
	levelInfo
	levelWarn
	levelError
)

var currentLevel = levelInfo

func levelToString(l logLevel) string {
	switch l {
	case levelDebug:
		return "DEBUG"
	case levelInfo:
		return "INFO"
	case levelWarn:
		return "WARN"
	case levelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func logMessage(level logLevel, args []objects.Object) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	parts := make([]string, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case *objects.String:
			parts[i] = v.Value
		default:
			parts[i] = v.Inspect()
		}
	}
	return fmt.Sprintf("[%s] %s: %s", timestamp, levelToString(level), strings.Join(parts, " "))
}

func init() {
	Register(&Module{
		Name: "std/log",
		Exports: map[string]objects.Object{
			// Debug level log
			"debug": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if currentLevel > levelDebug {
					return Null()
				}
				msg := logMessage(levelDebug, args)
				fmt.Fprintln(os.Stderr, msg)
				return String(msg)
			}),

			// Info level log
			"info": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if currentLevel > levelInfo {
					return Null()
				}
				msg := logMessage(levelInfo, args)
				fmt.Fprintln(os.Stdout, msg)
				return String(msg)
			}),

			// Warning level log
			"warn": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if currentLevel > levelWarn {
					return Null()
				}
				msg := logMessage(levelWarn, args)
				fmt.Fprintln(os.Stderr, msg)
				return String(msg)
			}),

			// Error level log
			"error": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if currentLevel > levelError {
					return Null()
				}
				msg := logMessage(levelError, args)
				fmt.Fprintln(os.Stderr, msg)
				return String(msg)
			}),

			// Fatal error (logs and exits)
			"fatal": BuiltinFunc(func(args ...objects.Object) objects.Object {
				msg := logMessage(levelError, args)
				fmt.Fprintln(os.Stderr, msg)
				os.Exit(1)
				return String(msg)
			}),

			// Set log level
			"setLevel": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("setLevel() takes exactly 1 argument")
				}
				level, ok := args[0].(*objects.String)
				if !ok {
					return Error("setLevel() requires a string argument")
				}
				switch strings.ToLower(level.Value) {
				case "debug":
					currentLevel = levelDebug
				case "info":
					currentLevel = levelInfo
				case "warn", "warning":
					currentLevel = levelWarn
				case "error":
					currentLevel = levelError
				default:
					return Error("setLevel() invalid level: " + level.Value)
				}
				return Null()
			}),

			// Get current log level
			"getLevel": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return String(levelToString(currentLevel))
			}),

			// Format log message without printing
			"format": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("format() takes at least 2 arguments")
				}
				levelStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("format() requires a string level as first argument")
				}
				var level logLevel
				switch strings.ToLower(levelStr.Value) {
				case "debug":
					level = levelDebug
				case "info":
					level = levelInfo
				case "warn", "warning":
					level = levelWarn
				case "error":
					level = levelError
				default:
					level = levelInfo
				}
				return String(logMessage(level, args[1:]))
			}),

			// Log to file (append mode)
			"toFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("toFile() takes at least 3 arguments")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("toFile() requires a string filename")
				}
				levelStr, ok := args[1].(*objects.String)
				if !ok {
					return Error("toFile() requires a string level")
				}
				var level logLevel
				switch strings.ToLower(levelStr.Value) {
				case "debug":
					level = levelDebug
				case "info":
					level = levelInfo
				case "warn", "warning":
					level = levelWarn
				case "error":
					level = levelError
				default:
					level = levelInfo
				}
				msg := logMessage(level, args[2:])
				file, err := os.OpenFile(filename.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()
				fmt.Fprintln(file, msg)
				return String(msg)
			}),

			// Simple print (no level)
			"print": BuiltinFunc(func(args ...objects.Object) objects.Object {
				parts := make([]string, len(args))
				for i, arg := range args {
					switch v := arg.(type) {
					case *objects.String:
						parts[i] = v.Value
					default:
						parts[i] = v.Inspect()
					}
				}
				msg := strings.Join(parts, " ")
				fmt.Println(msg)
				return String(msg)
			}),

			// Simple print without newline
			"printNoNL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				parts := make([]string, len(args))
				for i, arg := range args {
					switch v := arg.(type) {
					case *objects.String:
						parts[i] = v.Value
					default:
						parts[i] = v.Inspect()
					}
				}
				msg := strings.Join(parts, " ")
				fmt.Print(msg)
				return String(msg)
			}),

			// Printf style
			"printf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("printf() takes at least 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("printf() requires a string format")
				}
				goArgs := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					switch v := arg.(type) {
					case *objects.Int:
						goArgs[i] = v.Value
					case *objects.Float:
						goArgs[i] = v.Value
					case *objects.String:
						goArgs[i] = v.Value
					case *objects.Bool:
						goArgs[i] = v.Value
					default:
						goArgs[i] = v.Inspect()
					}
				}
				msg := fmt.Sprintf(format.Value, goArgs...)
				fmt.Print(msg)
				return String(msg)
			}),

			// Log with custom prefix
			"withPrefix": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("withPrefix() takes at least 2 arguments")
				}
				prefix, ok := args[0].(*objects.String)
				if !ok {
					return Error("withPrefix() requires a string prefix")
				}
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				parts := make([]string, len(args)-1)
				for i, arg := range args[1:] {
					switch v := arg.(type) {
					case *objects.String:
						parts[i] = v.Value
					default:
						parts[i] = v.Inspect()
					}
				}
				msg := fmt.Sprintf("[%s] %s: %s", timestamp, prefix.Value, strings.Join(parts, " "))
				fmt.Println(msg)
				return String(msg)
			}),

			// Log with JSON format
			"json": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("json() takes at least 2 arguments")
				}
				levelStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("json() requires a string level")
				}
				parts := make([]string, len(args)-1)
				for i, arg := range args[1:] {
					switch v := arg.(type) {
					case *objects.String:
						parts[i] = v.Value
					default:
						parts[i] = v.Inspect()
					}
				}
				msg := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"}`,
					time.Now().Format(time.RFC3339),
					strings.ToLower(levelStr.Value),
					strings.Join(parts, " "))
				fmt.Println(msg)
				return String(msg)
			}),

			// Log stack trace
			"stack": BuiltinFunc(func(args ...objects.Object) objects.Object {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				buf := make([]byte, 4096)
				n := runtimeStack(buf, false)
				msg := fmt.Sprintf("[%s] STACK:\n%s", timestamp, string(buf[:n]))
				fmt.Fprintln(os.Stderr, msg)
				return String(msg)
			}),

			// Is level enabled
			"isLevel": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isLevel() takes exactly 1 argument")
				}
				level, ok := args[0].(*objects.String)
				if !ok {
					return Error("isLevel() requires a string argument")
				}
				var checkLevel logLevel
				switch strings.ToLower(level.Value) {
				case "debug":
					checkLevel = levelDebug
				case "info":
					checkLevel = levelInfo
				case "warn", "warning":
					checkLevel = levelWarn
				case "error":
					checkLevel = levelError
				default:
					return Bool(false)
				}
				return Bool(checkLevel >= currentLevel)
			}),
		},
	})
}

// runtimeStack wrapper
func runtimeStack(buf []byte, all bool) int {
	return runtime.Stack(buf, all)
}
