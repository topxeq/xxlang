//go:build !windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains COM stub implementations for non-Windows platforms.
package webview2

import "errors"

// COMInitialize returns an error on non-Windows platforms.
func COMInitialize() error {
	return errors.New("COM is only supported on Windows")
}

// COMUninitialize does nothing on non-Windows platforms.
func COMUninitialize() {
	// No-op on non-Windows platforms
}
