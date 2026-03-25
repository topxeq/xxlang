//go:build windows && !js

// pkg/objects/builtin_system_windows.go
// Windows-specific system command utilities.
package objects

import (
	"os/exec"
	"syscall"
)

// setDetachedAttr sets the process attributes for detached execution on Windows.
func setDetachedAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}