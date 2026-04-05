//go:build linux || darwin
// +build linux darwin

// pkg/objects/input_unix.go
// Linux/MacOS-specific keyboard input implementation

package objects

import (
	"syscall"
	"unsafe"
)

const (
	tcgets = 0x5401
	tcsets = 0x5402

	// Terminal flags
	opost  = 0x1
	onlcr  = 0x2
	icanon = 0x2
	echok  = 0x4
	echo   = 0x8
	echoe  = 0x10
	isig   = 0x1
	iexten = 0x1
)

type termios struct {
	iflag  uint32
	oflag  uint32
	cflag  uint32
	lflag  uint32
	line   uint8
	cc     [32]uint8
	ispeed uint32
	ospeed uint32
}

var (
	origTermios termios
	termiosSet  bool
	stdinFd     = int(0) // stdin file descriptor
)

// initTermios initializes terminal in raw mode
func initTermios() error {
	if termiosSet {
		return nil
	}

	// Get current terminal settings
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(stdinFd), tcgets, uintptr(unsafe.Pointer(&origTermios)))
	if errno != 0 {
		return errno
	}

	// Save original settings
	termiosSet = true

	return nil
}

// setRawMode sets terminal to raw mode
func setRawMode() error {
	if err := initTermios(); err != nil {
		return err
	}

	// Copy current settings
	newTermios := origTermios

	// Disable canonical mode, echo, and signals
	newTermios.lflag &^= icanon | echo | echoe | echok | isig | iexten
	newTermios.oflag &^= opost | onlcr

	// Set minimum characters and timeout
	newTermios.cc[6] = 1  // VMIN
	newTermios.cc[7] = 0  // VTIME

	// Apply new settings
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(stdinFd), tcsets, uintptr(unsafe.Pointer(&newTermios)))
	if errno != 0 {
		return errno
	}

	return nil
}

// restoreTermios restores original terminal settings
func restoreTermios() {
	if termiosSet {
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(stdinFd), tcsets, uintptr(unsafe.Pointer(&origTermios)))
		termiosSet = false
	}
}

// readKey reads a single key press (blocking)
// Returns the character or special key code string
func readKey() (string, error) {
	if err := setRawMode(); err != nil {
		return "", err
	}
	defer restoreTermios()

	var buf [1]byte
	n, err := syscall.Read(stdinFd, buf[:])
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", nil
	}

	// Handle escape sequences
	if buf[0] == 27 { // ESC
		// Try to read more bytes for escape sequence
		var seq [3]byte
		syscall.SetNonblock(stdinFd, true)
		n2, _ := syscall.Read(stdinFd, seq[:])
		syscall.SetNonblock(stdinFd, false)

		if n2 >= 2 {
			if seq[0] == '[' {
				switch seq[1] {
				case 'A':
					return "UP", nil
				case 'B':
					return "DOWN", nil
				case 'C':
					return "RIGHT", nil
				case 'D':
					return "LEFT", nil
				}
			}
		}
		return "ESCAPE", nil
	}

	// Handle special characters
	switch buf[0] {
	case 3: // Ctrl+C
		return "\x03", nil
	case 13, 10: // Enter
		return "ENTER", nil
	case 8, 127: // Backspace
		return "BACKSPACE", nil
	case 9: // Tab
		return "TAB", nil
	}

	return string(buf[0]), nil
}

// hasKeyAvailable checks if there's a key available without blocking
func hasKeyAvailable() bool {
	if err := initTermios(); err != nil {
		return false
	}

	// Set non-blocking mode
	syscall.SetNonblock(stdinFd, true)
	defer syscall.SetNonblock(stdinFd, false)

	var buf [1]byte
	n, _ := syscall.Read(stdinFd, buf[:])

	return n > 0
}
