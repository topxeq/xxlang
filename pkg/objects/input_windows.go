//go:build windows
// +build windows

// pkg/objects/input_windows.go
// Windows-specific non-blocking keyboard input implementation

package objects

import (
	"syscall"
	"unsafe"
)

var (
	kernel32DLL                  = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle             = kernel32DLL.NewProc("GetStdHandle")
	procCreateFile               = kernel32DLL.NewProc("CreateFileW")
	procReadConsoleInput         = kernel32DLL.NewProc("ReadConsoleInputW")
	procPeekConsoleInput         = kernel32DLL.NewProc("PeekConsoleInputW")
	procGetNumberOfConsoleInputEvents = kernel32DLL.NewProc("GetNumberOfConsoleInputEvents")
	procFlushConsoleInputBuffer  = kernel32DLL.NewProc("FlushConsoleInputBuffer")
	procGetConsoleMode           = kernel32DLL.NewProc("GetConsoleMode")
	procSetConsoleMode           = kernel32DLL.NewProc("SetConsoleMode")
	procReadFile                 = kernel32DLL.NewProc("ReadFile")
)

const (
	stdInputHandle = uintptr(0xFFFFFFF6) // STD_INPUT_HANDLE = -10 (as uint32)

	// Console mode flags
	enable_processed_input = 0x0001
	enable_line_input      = 0x0002
	enable_echo_input      = 0x0004
	enable_window_input    = 0x0008
	enable_mouse_input     = 0x0010
	enable_fast_io_input   = 0x0020
	enable_virtual_terminal_input = 0x0200

	// Input event types
	key_event = 1

	// Key modifiers
	right_alt_pressed  = 1
	left_alt_pressed   = 2
	right_ctrl_pressed = 4
	left_ctrl_pressed  = 8
	shift_pressed      = 16
	capslock_on        = 128
	numlock_on         = 32
	scrolllock_on      = 64
	enhanced_key       = 256

	// Virtual key codes
	vk_shift   = 0x10
	vk_control = 0x11
	vk_menu    = 0x12 // Alt
	vk_capital = 0x14 // CapsLock
	vk_escape  = 0x1B
	vk_space   = 0x20
	vk_return  = 0x0D // Enter
	vk_back    = 0x08 // Backspace
	vk_tab     = 0x09 // Tab
	vk_left    = 0x25
	vk_up      = 0x26
	vk_right   = 0x27
	vk_down    = 0x28
	vk_f1      = 0x70
	vk_f2      = 0x71
	vk_f3      = 0x72
	vk_f4      = 0x73
	vk_f5      = 0x74
	vk_f6      = 0x75
	vk_f7      = 0x76
	vk_f8      = 0x77
	vk_f9      = 0x78
	vk_f10     = 0x79
	vk_f11     = 0x7A
	vk_f12     = 0x7B
	// VK_A - VK_Z (0x41 - 0x5A = 65 - 90)
	vk_a = 0x41
	vk_z = 0x5A
)

// INPUT_RECORD represents a console input record
type inputRecord struct {
	eventType uint16
	padding   uint16
	event     [16]byte
}

// KEY_EVENT_RECORD represents a key event
// Windows structure (20 bytes total, but packed into 16 bytes in INPUT_RECORD union):
//   DWORD bKeyDown (4 bytes)
//   WORD wRepeatCount (2 bytes)
//   WORD wReserved (2 bytes) - padding
//   WORD wVirtualKeyCode (2 bytes) - VK code (0x41='A', 0x1E='A' scan code)
//   WORD wVirtualScanCode (2 bytes) - Hardware scan code
//   WORD uChar (2 bytes) - Unicode character or control key state low word
//   DWORD dwControlKeyState (4 bytes) - High word may overlap with uChar
//
// In Go, we need to match the actual memory layout used by ReadConsoleInput
type keyEventRecord struct {
	keyDown         uint32  // bKeyDown (4 bytes)
	repeatCount     uint16  // wRepeatCount (2 bytes)
	wReserved       uint16  // wReserved (2 bytes)
	virtualScanCode uint16  // wVirtualScanCode FIRST (2 bytes) - matches observed data
	virtualKeyCode  uint16  // wVirtualKeyCode SECOND (2 bytes)
	char            uint16  // uChar (2 bytes)
	keyState        uint32  // dwControlKeyState (4 bytes)
}

var stdinHandle syscall.Handle
var stdinInitialized bool
var origMode uint32
var modeSet bool
var inputBuffer []byte  // Buffer for raw input
var bufferPos int

// initStdin initializes Windows console input and sets raw mode
func initStdin() error {
	if stdinInitialized {
		return nil
	}

	// Open console input device using CreateFile with read-write access
	// CONIN$ is the console input device in Windows
	conin, err := syscall.UTF16PtrFromString("CONIN$")
	if err != nil {
		return err
	}

	// GENERIC_READ | GENERIC_WRITE = 0x80000000 | 0x40000000 = 0xC0000000
	// This is needed because SetConsoleMode requires write access
	handle, _, errno := procCreateFile.Call(
		uintptr(unsafe.Pointer(conin)),
		0xC0000000, // GENERIC_READ | GENERIC_WRITE
		1,          // FILE_SHARE_READ
		0,          // default security
		3,          // OPEN_EXISTING
		0,          // FILE_ATTRIBUTE_NORMAL
		0,          // no template file
	)

	if handle == 0 || handle == ^uintptr(0) {
		return errno
	}

	stdinHandle = syscall.Handle(handle)

	// Get current console mode
	var mode uint32
	ret, _, errno := procGetConsoleMode.Call(uintptr(stdinHandle), uintptr(unsafe.Pointer(&mode)))

	if ret != 0 {
		// Successfully got mode
		origMode = mode

		// Clear line input and echo flags
		newMode := mode &^ (enable_line_input | enable_echo_input)

		ret2, _, _ := procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(newMode))
		if ret2 != 0 {
			modeSet = true
		}
	}

	// Flush any pending input events
	procFlushConsoleInputBuffer.Call(uintptr(stdinHandle))

	stdinInitialized = true
	return nil
}

// readKey reads a single key press (blocking with timeout)
// Returns the character or special key code string
func readKey() (string, error) {
	if err := initStdin(); err != nil {
		return "", err
	}

	var rec inputRecord
	var numRead uint32

	for {
		ret, _, errno := procReadConsoleInput.Call(
			uintptr(stdinHandle),
			uintptr(unsafe.Pointer(&rec)),
			1,
			uintptr(unsafe.Pointer(&numRead)),
		)

		if ret == 0 {
			return "", errno
		}

		if numRead == 0 {
			continue
		}

		// Check if it's a key event
		if rec.eventType != key_event {
			continue
		}

		// Parse key event
		var keyRec keyEventRecord
		copy((*[unsafe.Sizeof(keyRec)]byte)(unsafe.Pointer(&keyRec))[:], rec.event[:])

		// Skip key-up events (only process key-down)
		if keyRec.keyDown == 0 {
			continue
		}

		// Skip modifier keys (Shift, Ctrl, Alt, CapsLock)
		switch keyRec.virtualKeyCode {
		case vk_shift, vk_control, vk_menu, vk_capital:
			continue
		}

		ch := keyRec.char

		// Handle Ctrl+C (ASCII 3)
		if keyRec.keyState&(left_ctrl_pressed|right_ctrl_pressed) != 0 {
			if ch == 3 || keyRec.virtualKeyCode == vk_escape {
				return "\x03", nil
			}
			// Ctrl+Letter: convert to lowercase
			if keyRec.virtualKeyCode >= 'A' && keyRec.virtualKeyCode <= 'Z' {
				return string(rune(keyRec.virtualKeyCode + 32)), nil
			}
		}

		// Handle character keys (when char is provided)
		if ch != 0 {
			return string(rune(ch)), nil
		}

		// Handle keys where virtualKeyCode contains ASCII value
		if (keyRec.virtualKeyCode >= 0x61 && keyRec.virtualKeyCode <= 0x7A) || // a-z
		   (keyRec.virtualKeyCode >= 0x41 && keyRec.virtualKeyCode <= 0x5A) {  // A-Z
			// Flush buffer to remove key-up events
			procFlushConsoleInputBuffer.Call(uintptr(stdinHandle))
			return string(rune(keyRec.virtualKeyCode)), nil
		}

		// Handle keys by scan code (when virtualKeyCode doesn't have ASCII)
		// Scan codes for letters: A=0x1E, B=0x30, C=0x2E, D=0x20, E=0x12, etc.
		// This is a fallback for when other methods fail
		scanCode := keyRec.virtualScanCode
		if scanCode >= 0x1E && scanCode <= 0x39 {
			// Map scan code to letter (simplified - only handles A-Z row)
			// A=0x1E, B=0x30, C=0x2E, D=0x20, E=0x12, F=0x21, G=0x22, H=0x23, I=0x17
			// J=0x24, K=0x25, L=0x26, M=0x32, N=0x31, O=0x18, P=0x19, Q=0x10, R=0x13
			// S=0x1F, T=0x14, U=0x16, V=0x2F, W=0x11, X=0x2D, Y=0x15, Z=0x2C
			scanToChar := map[uint16]rune{
				0x1E: 'a', 0x30: 'b', 0x2E: 'c', 0x20: 'd', 0x12: 'e',
				0x21: 'f', 0x22: 'g', 0x23: 'h', 0x17: 'i', 0x24: 'j',
				0x25: 'k', 0x26: 'l', 0x32: 'm', 0x31: 'n', 0x18: 'o',
				0x19: 'p', 0x10: 'q', 0x13: 'r', 0x1F: 's', 0x14: 't',
				0x16: 'u', 0x2F: 'v', 0x11: 'w', 0x2D: 'x', 0x15: 'y',
				0x2C: 'z',
			}
			if char, ok := scanToChar[scanCode]; ok {
				// Check shift/caps state
				shiftPressed := (keyRec.keyState & shift_pressed) != 0
				capsOn := (keyRec.keyState & capslock_on) != 0
				if shiftPressed != capsOn {
					// Uppercase
					return string(char - 32), nil
				}
				return string(char), nil
			}
		}

		// Handle special keys
		switch keyRec.virtualKeyCode {
		case vk_left:
			return "LEFT", nil
		case vk_right:
			return "RIGHT", nil
		case vk_up:
			return "UP", nil
		case vk_down:
			return "DOWN", nil
		case vk_escape:
			return "ESCAPE", nil
		case vk_space:
			return " ", nil
		case vk_return:
			return "ENTER", nil
		case vk_back:
			return "BACKSPACE", nil
		case vk_tab:
			return "TAB", nil
		case vk_f1:
			return "F1", nil
		case vk_f2:
			return "F2", nil
		case vk_f3:
			return "F3", nil
		case vk_f4:
			return "F4", nil
		case vk_f5:
			return "F5", nil
		case vk_f6:
			return "F6", nil
		case vk_f7:
			return "F7", nil
		case vk_f8:
			return "F8", nil
		case vk_f9:
			return "F9", nil
		case vk_f10:
			return "F10", nil
		case vk_f11:
			return "F11", nil
		case 28: // Scan code for Enter
			return "ENTER", nil
		case vk_f12:
			return "F12", nil
		}
	}
}

// hasKeyAvailable checks if there's a key available without blocking
func hasKeyAvailable() bool {
	if err := initStdin(); err != nil {
		return false
	}

	var numEvents uint32
	ret, _, _ := procGetNumberOfConsoleInputEvents.Call(
		uintptr(stdinHandle),
		uintptr(unsafe.Pointer(&numEvents)),
	)

	if ret == 0 || numEvents == 0 {
		return false
	}

	// Peek at pending events
	var rec inputRecord
	procPeekConsoleInput.Call(
		uintptr(stdinHandle),
		uintptr(unsafe.Pointer(&rec)),
		1,
		uintptr(unsafe.Pointer(&numEvents)),
	)

	if rec.eventType != key_event {
		return false
	}

	var keyRec keyEventRecord
	copy((*[unsafe.Sizeof(keyRec)]byte)(unsafe.Pointer(&keyRec))[:], rec.event[:])

	// Skip key-up events (only process key-down)
	if keyRec.keyDown == 0 {
		// Read and discard the key-up event to clear it from buffer
		var discardRec inputRecord
		procReadConsoleInput.Call(uintptr(stdinHandle), uintptr(unsafe.Pointer(&discardRec)), 1, uintptr(unsafe.Pointer(&numEvents)))
		return false
	}

	// Skip modifier keys (Shift, Ctrl, Alt, CapsLock)
	switch keyRec.virtualKeyCode {
	case vk_shift, vk_control, vk_menu, vk_capital:
		return false
	}

	// Check if this is a valid key (has a character or is a special key)
	if keyRec.char != 0 {
		return true
	}

	// Check if virtualKeyCode contains ASCII value (a-z, A-Z, 0-9, etc.)
	if (keyRec.virtualKeyCode >= 0x61 && keyRec.virtualKeyCode <= 0x7A) || // a-z
	   (keyRec.virtualKeyCode >= 0x41 && keyRec.virtualKeyCode <= 0x5A) || // A-Z
	   (keyRec.virtualKeyCode >= 0x30 && keyRec.virtualKeyCode <= 0x39) {  // 0-9
		return true
	}

	// Check for letter keys by scan code (0x1E-0x39 for A-Z)
	if keyRec.virtualScanCode >= 0x1E && keyRec.virtualScanCode <= 0x39 {
		return true
	}

	// Check for special keys
	switch keyRec.virtualKeyCode {
	case vk_left, vk_right, vk_up, vk_down, vk_escape, vk_space, vk_f1, vk_f2, vk_f3, vk_f4, vk_f5, vk_f6, vk_f7, vk_f8, vk_f9, vk_f10, vk_f11, vk_f12, vk_return, vk_back, vk_tab:
		return true
	}

	return false
}

// flushInputBuffer clears the input buffer
func flushInputBuffer() {
	if stdinInitialized {
		procFlushConsoleInputBuffer.Call(uintptr(stdinHandle))
	}
}

// restoreMode restores the original console mode
func restoreMode() {
	if modeSet && stdinInitialized {
		procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(origMode))
		modeSet = false
	}
}
