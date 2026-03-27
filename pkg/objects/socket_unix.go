//go:build !windows
// +build !windows

// pkg/objects/socket_unix.go
// Socket options implementation for Unix-like systems (Linux, macOS, BSD, etc.)
package objects

import (
	"syscall"
)

// syscallConn is an interface for getting the underlying file descriptor.
type syscallConn interface {
	SyscallConn() (syscall.RawConn, error)
}

// setReuseAddr sets SO_REUSEADDR on a socket connection.
// This allows binding to an address that is already in use.
func setReuseAddr(c syscallConn) error {
	rawConn, err := c.SyscallConn()
	if err != nil {
		return err
	}

	var sockErr error
	err = rawConn.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return sockErr
}