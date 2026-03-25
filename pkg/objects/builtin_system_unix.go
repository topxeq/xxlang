//go:build !windows && !js

// pkg/objects/builtin_system_unix.go
// Unix-specific system command utilities.
package objects

import (
	"os/exec"
	"syscall"
)

// setDetachedAttr sets the process attributes for detached execution on Unix.
func setDetachedAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}