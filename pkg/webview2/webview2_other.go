//go:build !windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains stub implementations for non-Windows platforms.
package webview2

import "errors"

// ErrNotSupported is returned when WebView2 is not supported on the current platform.
var ErrNotSupported = errors.New("WebView2 is only supported on Windows")

// WebView2 represents a WebView2 instance (stub).
type WebView2 struct{}

// WebView2Config holds configuration for WebView2 (stub).
type WebView2Config struct {
	Title          string
	Width          int
	Height         int
	UserDataFolder string
	Debug          bool
}

// GetInstalledVersion returns an error on non-Windows platforms.
func GetInstalledVersion() (string, error) {
	return "", ErrNotSupported
}

// NewWebView2 returns an error on non-Windows platforms.
func NewWebView2(config WebView2Config) (*WebView2, error) {
	return nil, ErrNotSupported
}

// Navigate returns an error on non-Windows platforms.
func (wv *WebView2) Navigate(url string) error {
	return ErrNotSupported
}

// NavigateToString returns an error on non-Windows platforms.
func (wv *WebView2) NavigateToString(html string) error {
	return ErrNotSupported
}

// ExecuteScript returns an error on non-Windows platforms.
func (wv *WebView2) ExecuteScript(script string, callback func(result string, err error)) error {
	return ErrNotSupported
}

// PostWebMessageAsJson returns an error on non-Windows platforms.
func (wv *WebView2) PostWebMessageAsJson(json string) error {
	return ErrNotSupported
}

// BindFunction does nothing on non-Windows platforms.
func (wv *WebView2) BindFunction(name string, fn func(args map[string]interface{})) {
	// No-op on non-Windows platforms
}

// Run does nothing on non-Windows platforms.
func (wv *WebView2) Run() {
	// No-op on non-Windows platforms
}

// Close does nothing on non-Windows platforms.
func (wv *WebView2) Close() {
	// No-op on non-Windows platforms
}
