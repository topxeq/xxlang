//go:build !js

// pkg/objects/builtin_env.go
// Environment and OS information related built-in functions for Xxlang
package objects

import (
	"fmt"
	"os"
	"runtime"
)

func init() {
	// Environment functions
	Builtins["getEnv"] = &Builtin{Fn: builtinGetEnv}
	Builtins["setEnv"] = &Builtin{Fn: builtinSetEnv}
	Builtins["getOSName"] = &Builtin{Fn: builtinGetOSName}
	Builtins["getOSArch"] = &Builtin{Fn: builtinGetOSArch}
	Builtins["getOSArgs"] = &Builtin{Fn: builtinGetOSArgs}
	Builtins["getAppPath"] = &Builtin{Fn: builtinGetAppPath}
	Builtins["getAppDir"] = &Builtin{Fn: builtinGetAppDir}
	Builtins["exit"] = &Builtin{Fn: builtinExit}
	Builtins["getSysInfo"] = &Builtin{Fn: builtinGetSysInfo}
	Builtins["getPid"] = &Builtin{Fn: builtinGetPid}
	Builtins["getPPid"] = &Builtin{Fn: builtinGetPPid}
	Builtins["hostname"] = &Builtin{Fn: builtinHostname}
	Builtins["getXxlVersion"] = &Builtin{Fn: builtinGetXxlVersion}
	Builtins["getXxlBuildNumber"] = &Builtin{Fn: builtinGetXxlBuildNumber}
}

// XxlVersion holds the Xxlang version string. It defaults to "dev" and is
// overridden by the binary entry point (cmd/xxl) at startup with the value
// injected via -ldflags (-X main.Version=...). Exposing it as a package-level
// variable here lets the builtin getXxlVersion() read it without import cycles.
var XxlVersion = "dev"

// XxlBuildNumber holds the Xxlang build number. It defaults to "0" and is
// overridden by the binary entry point (cmd/xxl) at startup. The build number
// is hard-coded in cmd/xxl/main.go (not ldflags-injected), so it is propagated
// to objects the same way XxlVersion is.
var XxlBuildNumber = "0"

// getEnv - get environment variable
// Usage: getEnv(key) -> string or null
//
//	getEnv(key, defaultValue) -> string
func builtinGetEnv(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for getEnv. got=%d, want=1 or 2", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getEnv' must be STRING, got %s", args[0].Type())
	}

	value := os.Getenv(key.Value)
	if value == "" {
		if len(args) == 2 {
			return args[1]
		}
		return NULL
	}

	return NewString(value)
}

// setEnv - set environment variable
// Usage: setEnv(key, value) -> null
func builtinSetEnv(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for setEnv. got=%d, want=2", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'setEnv' must be STRING, got %s", args[0].Type())
	}

	value, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'setEnv' must be STRING, got %s", args[1].Type())
	}

	err := os.Setenv(key.Value, value.Value)
	if err != nil {
		return newError("setEnv error: %v", err)
	}
	return NULL
}

// getOSName - get operating system name
// Usage: getOSName() -> string
func builtinGetOSName(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getOSName. got=%d, want=0", len(args))
	}
	return NewString(runtime.GOOS)
}

// getOSArch - get system architecture
// Usage: getOSArch() -> string
func builtinGetOSArch(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getOSArch. got=%d, want=0", len(args))
	}
	return NewString(runtime.GOARCH)
}

// getOSArgs - get command line arguments
// Usage: getOSArgs() -> array
func builtinGetOSArgs(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getOSArgs. got=%d, want=0", len(args))
	}

	argList := os.Args
	elements := make([]Object, len(argList))
	for i, arg := range argList {
		elements[i] = NewString(arg)
	}
	return NewArray(elements)
}

// getAppPath - get executable path
// Usage: getAppPath() -> string
func builtinGetAppPath(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getAppPath. got=%d, want=0", len(args))
	}

	path, err := os.Executable()
	if err != nil {
		return newError("getAppPath error: %v", err)
	}
	return NewString(path)
}

// getAppDir - get executable directory
// Usage: getAppDir() -> string
func builtinGetAppDir(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getAppDir. got=%d, want=0", len(args))
	}

	path, err := os.Executable()
	if err != nil {
		return newError("getAppDir error: %v", err)
	}
	dir := getDirFromPath(path)
	return NewString(dir)
}

// exit - exit program
// Usage: exit() -> exits with code 0
//
//	exit(code) -> exits with specified code
func builtinExit(args ...Object) Object {
	code := 0
	if len(args) > 1 {
		return newError("wrong number of arguments for exit. got=%d, want=0 or 1", len(args))
	}

	if len(args) == 1 {
		c, ok := args[0].(*Int)
		if !ok {
			return newError("argument to 'exit' must be INT, got %s", args[0].Type())
		}
		code = int(c.Value)
	}

	os.Exit(code)
	return NULL
}

// getSysInfo - get system information
// Usage: getSysInfo() -> map
func builtinGetSysInfo(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getSysInfo. got=%d, want=0", len(args))
	}

	pairs := make(map[HashKey]MapPair)

	pairs[NewString("os").HashKey()] = MapPair{
		Key:   NewString("os"),
		Value: NewString(runtime.GOOS),
	}
	pairs[NewString("arch").HashKey()] = MapPair{
		Key:   NewString("arch"),
		Value: NewString(runtime.GOARCH),
	}
	pairs[NewString("cpus").HashKey()] = MapPair{
		Key:   NewString("cpus"),
		Value: NewInt(int64(runtime.NumCPU())),
	}
	pairs[NewString("goroutines").HashKey()] = MapPair{
		Key:   NewString("goroutines"),
		Value: NewInt(int64(runtime.NumGoroutine())),
	}
	pairs[NewString("version").HashKey()] = MapPair{
		Key:   NewString("version"),
		Value: NewString(runtime.Version()),
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pairs[NewString("memAlloc").HashKey()] = MapPair{
		Key:   NewString("memAlloc"),
		Value: NewInt(int64(memStats.Alloc)),
	}
	pairs[NewString("memTotal").HashKey()] = MapPair{
		Key:   NewString("memTotal"),
		Value: NewInt(int64(memStats.TotalAlloc)),
	}
	pairs[NewString("memSys").HashKey()] = MapPair{
		Key:   NewString("memSys"),
		Value: NewInt(int64(memStats.Sys)),
	}

	return NewMap(pairs)
}

// getXxlVersion - get the Xxlang version string.
// Usage: getXxlVersion() -> string
// The version is injected at build time via -ldflags (-X main.Version=...)
// and defaults to "dev" for local builds.
func builtinGetXxlVersion(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getXxlVersion. got=%d, want=0", len(args))
	}
	return NewString(XxlVersion)
}

// getXxlBuildNumber - get the Xxlang build number string.
// Usage: getXxlBuildNumber() -> string
// The build number is hard-coded in cmd/xxl/main.go and propagated to this
// package at startup. It defaults to "0" for local builds.
func builtinGetXxlBuildNumber(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getXxlBuildNumber. got=%d, want=0", len(args))
	}
	return NewString(XxlBuildNumber)
}

// getPid - get process ID
// Usage: getPid() -> int
func builtinGetPid(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getPid. got=%d, want=0", len(args))
	}
	return NewInt(int64(os.Getpid()))
}

// getPPid - get parent process ID
// Usage: getPPid() -> int
func builtinGetPPid(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getPPid. got=%d, want=0", len(args))
	}
	return NewInt(int64(os.Getppid()))
}

// hostname - get system hostname
// Usage: hostname() -> string
func builtinHostname(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for hostname. got=%d, want=0", len(args))
	}

	name, err := os.Hostname()
	if err != nil {
		return newError("hostname error: %v", err)
	}
	return NewString(name)
}

// Helper function to get directory from path
func getDirFromPath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

// Helper function for formatting bytes
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
