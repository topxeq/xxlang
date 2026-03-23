//go:build windows
// +build windows

// pkg/objects/file_lock_windows.go
// File locking implementation for Windows systems.
package objects

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	lockFileExProc   = kernel32.NewProc("LockFileEx")
	unlockFileExProc = kernel32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
)

// lockFile implements file locking using LockFileEx on Windows.
func (f *File) lockFile(lockType FileLockType, blocking bool) error {
	var flags uint32
	var overlapped syscall.Overlapped

	if lockType == LockExclusive {
		flags |= LOCKFILE_EXCLUSIVE_LOCK
	}

	if !blocking {
		flags |= LOCKFILE_FAIL_IMMEDIATELY
	}

	handle := syscall.Handle(f.Handle.Fd())

	ret, _, err := lockFileExProc.Call(
		uintptr(handle),
		uintptr(flags),
		0,
		0xFFFFFFFF, // Lock entire file (low 32 bits)
		0xFFFFFFFF, // Lock entire file (high 32 bits)
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if ret == 0 {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return nil
}

// unlockFile releases the file lock on Windows.
func (f *File) unlockFile() error {
	var overlapped syscall.Overlapped
	handle := syscall.Handle(f.Handle.Fd())

	ret, _, err := unlockFileExProc.Call(
		uintptr(handle),
		0,
		0xFFFFFFFF, // Unlock entire file (low 32 bits)
		0xFFFFFFFF, // Unlock entire file (high 32 bits)
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if ret == 0 {
		return fmt.Errorf("failed to release file lock: %w", err)
	}

	return nil
}
