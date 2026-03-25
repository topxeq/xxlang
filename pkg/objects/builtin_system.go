//go:build !js

// pkg/objects/builtin_system.go
// System command builtin functions for Xxlang.
// Provides systemCmd, systemCmdDetached, and systemStart functions.
package objects

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

// ============================================================
// System Command Builtin Functions
// ============================================================

// BuiltinSystemCmd executes a system command synchronously and returns the result.
// Usage: systemCmd(cmd) or systemCmd(cmd, args...)
// Returns a map with:
//   - "success": bool - whether the command succeeded
//   - "exitCode": int - the exit code
//   - "output": string - combined stdout and stderr
//   - "error": string - error message if any
var BuiltinSystemCmd = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for systemCmd. got=%d, want>=1", len(args))
		}

		cmdStr, ok := args[0].(*String)
		if !ok {
			return newError("first argument to 'systemCmd' must be STRING, got %s", args[0].Type())
		}

		// Build command arguments
		var cmdArgs []string
		for i := 1; i < len(args); i++ {
			if arg, ok := args[i].(*String); ok {
				cmdArgs = append(cmdArgs, arg.Value)
			} else {
				return newError("argument %d to 'systemCmd' must be STRING, got %s", i+1, args[i].Type())
			}
		}

		// Execute command
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			// On Windows, use cmd.exe to run the command
			if len(cmdArgs) > 0 {
				cmd = exec.Command(cmdStr.Value, cmdArgs...)
			} else {
				cmd = exec.Command("cmd", "/c", cmdStr.Value)
			}
		} else {
			// On Unix-like systems, use sh to run the command
			if len(cmdArgs) > 0 {
				cmd = exec.Command(cmdStr.Value, cmdArgs...)
			} else {
				cmd = exec.Command("sh", "-c", cmdStr.Value)
			}
		}

		// Capture output
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Stdin = nil

		err := cmd.Run()
		output := stdout.String() + stderr.String()

		// Build result map
		result := NewOrderedMap()
		result.Set(NewString("output"), NewString(output))

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.Set(NewString("success"), FALSE)
				result.Set(NewString("exitCode"), NewInt(int64(exitErr.ExitCode())))
				result.Set(NewString("error"), NewString(err.Error()))
			} else {
				result.Set(NewString("success"), FALSE)
				result.Set(NewString("exitCode"), NewInt(-1))
				result.Set(NewString("error"), NewString(err.Error()))
			}
		} else {
			result.Set(NewString("success"), TRUE)
			result.Set(NewString("exitCode"), NewInt(0))
			result.Set(NewString("error"), NewString(""))
		}

		return result
	},
}

// BuiltinSystemCmdDetached executes a system command in detached mode (asynchronously).
// The command runs in the background without waiting for completion.
// Usage: systemCmdDetached(cmd) or systemCmdDetached(cmd, args...)
// Returns a map with:
//   - "success": bool - whether the command was started successfully
//   - "pid": int - process ID (0 if not available)
//   - "error": string - error message if any
var BuiltinSystemCmdDetached = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for systemCmdDetached. got=%d, want>=1", len(args))
		}

		cmdStr, ok := args[0].(*String)
		if !ok {
			return newError("first argument to 'systemCmdDetached' must be STRING, got %s", args[0].Type())
		}

		// Build command arguments
		var cmdArgs []string
		for i := 1; i < len(args); i++ {
			if arg, ok := args[i].(*String); ok {
				cmdArgs = append(cmdArgs, arg.Value)
			} else {
				return newError("argument %d to 'systemCmdDetached' must be STRING, got %s", i+1, args[i].Type())
			}
		}

		// Create command
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			if len(cmdArgs) > 0 {
				cmd = exec.Command(cmdStr.Value, cmdArgs...)
			} else {
				cmd = exec.Command("cmd", "/c", cmdStr.Value)
			}
		} else {
			if len(cmdArgs) > 0 {
				cmd = exec.Command(cmdStr.Value, cmdArgs...)
			} else {
				cmd = exec.Command("sh", "-c", cmdStr.Value)
			}
		}

		// Detach from parent process
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		// Set platform-specific attributes for detached execution
		setDetachedAttr(cmd)

		err := cmd.Start()

		result := NewOrderedMap()
		if err != nil {
			result.Set(NewString("success"), FALSE)
			result.Set(NewString("pid"), NewInt(0))
			result.Set(NewString("error"), NewString(err.Error()))
			return result
		}

		// Get PID if available
		pid := 0
		if cmd.Process != nil {
			pid = cmd.Process.Pid
		}

		// Release the process so it continues running after we return
		cmd.Process.Release()

		result.Set(NewString("success"), TRUE)
		result.Set(NewString("pid"), NewInt(int64(pid)))
		result.Set(NewString("error"), NewString(""))

		return result
	},
}

// BuiltinSystemStart starts a program or opens a file/URL with the default application.
// This is similar to the "start" command on Windows or "open" on macOS.
// Usage: systemStart(path) or systemStart(path, workingDir)
// Returns a map with:
//   - "success": bool - whether the operation succeeded
//   - "error": string - error message if any
var BuiltinSystemStart = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for systemStart. got=%d, want>=1", len(args))
		}

		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("first argument to 'systemStart' must be STRING, got %s", args[0].Type())
		}

		path := pathStr.Value

		// Optional working directory
		var workingDir string
		if len(args) >= 2 {
			if wd, ok := args[1].(*String); ok {
				workingDir = wd.Value
			}
		}

		var err error

		switch runtime.GOOS {
		case "windows":
			// Use cmd.exe start command on Windows
			// The empty string after "start" is the window title
			cmd := exec.Command("cmd", "/c", "start", "", path)
			if workingDir != "" {
				cmd.Dir = workingDir
			}
			cmd.Stdout = nil
			cmd.Stderr = nil
			err = cmd.Run()

		case "darwin":
			// Use 'open' command on macOS
			cmd := exec.Command("open", path)
			if workingDir != "" {
				cmd.Dir = workingDir
			}
			cmd.Stdout = nil
			cmd.Stderr = nil
			err = cmd.Run()

		case "linux":
			// Try xdg-open on Linux
			cmd := exec.Command("xdg-open", path)
			if workingDir != "" {
				cmd.Dir = workingDir
			}
			cmd.Stdout = nil
			cmd.Stderr = nil
			err = cmd.Run()

		default:
			err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
		}

		result := NewOrderedMap()
		if err != nil {
			result.Set(NewString("success"), FALSE)
			result.Set(NewString("error"), NewString(err.Error()))
		} else {
			result.Set(NewString("success"), TRUE)
			result.Set(NewString("error"), NewString(""))
		}

		return result
	},
}