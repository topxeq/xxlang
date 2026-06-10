//go:build !windows

package jit

// getWindowsCallbackPtr returns 0 on non-Windows platforms.
// Windows x64 ABI callbacks are only available on Windows.
func getWindowsCallbackPtr() uintptr {
	return 0
}
