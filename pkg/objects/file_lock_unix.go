//go:build !windows
// +build !windows

// pkg/objects/file_lock_unix.go
// File locking implementation for Unix-like systems (Linux, macOS, BSD, etc.)
package objects

import (
	"fmt"
	"syscall"
)

// lockFile implements file locking using flock on Unix systems.
func (f *File) lockFile(lockType FileLockType, blocking bool) error {
	var how int
	if blocking {
		how = syscall.LOCK_NB
	} else {
		how = 0
	}

	switch lockType {
	case LockShared:
		how |= syscall.LOCK_SH
	case LockExclusive:
		how |= syscall.LOCK_EX
	default:
		return fmt.Errorf("invalid lock type: %d", lockType)
	}

	err := syscall.Flock(int(f.Handle.Fd()), how)
	if err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return nil
}

// unlockFile releases the file lock on Unix systems.
func (f *File) unlockFile() error {
	err := syscall.Flock(int(f.Handle.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("failed to release file lock: %w", err)
	}
	return nil
}
