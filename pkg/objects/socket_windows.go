//go:build windows
// +build windows

// pkg/objects/socket_windows.go
// Socket options implementation for Windows systems.
package objects

import (
	"syscall"
)

// syscallConn is an interface for getting the underlying file descriptor.
type syscallConn interface {
	SyscallConn() (syscall.RawConn, error)
}

// setReuseAddr sets SO_REUSEADDR on a socket connection.
// On Windows, SO_REUSEADDR has different semantics than Unix,
// but we enable it for consistency with the API.
func setReuseAddr(c syscallConn) error {
	rawConn, err := c.SyscallConn()
	if err != nil {
		return err
	}

	var sockErr error
	err = rawConn.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return sockErr
}