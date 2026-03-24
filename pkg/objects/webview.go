// Package objects defines the object types for Xxlang.
// This file contains the WebView object type for GUI support.
package objects

import (
	"sync"
	"unsafe"
)

// WebViewType is the object type for WebView.
const WebViewType ObjectType = "WEBVIEW"

// TagWebView is the type tag for WebView objects.
const TagWebView TypeTag = 101 // Continue from existing tags

// WebView represents a WebView2 window in Xxlang.
type WebView struct {
	Handle    interface{} // *webview2.WebView2 on Windows
	callbacks map[string]*Function
	mu        sync.Mutex
	closed    bool
}

// NewWebView creates a new WebView object.
func NewWebView(handle interface{}) *WebView {
	return &WebView{
		Handle:    handle,
		callbacks: make(map[string]*Function),
	}
}

// Type returns the object type.
func (w *WebView) Type() ObjectType { return WebViewType }

// TypeTag returns the type tag for fast type checking.
func (w *WebView) TypeTag() TypeTag { return TagWebView }

// Inspect returns a string representation.
func (w *WebView) Inspect() string {
	if w.closed {
		return "<WEBVIEW closed>"
	}
	return "<WEBVIEW active>"
}

// ToBool returns true if the WebView is active.
func (w *WebView) ToBool() *Bool { return &Bool{Value: !w.closed} }

// HashKey returns a hash key for the WebView.
func (w *WebView) HashKey() HashKey {
	return HashKey{
		Type:  WebViewType,
		Value: uint64(uintptr(unsafe.Pointer(w))),
	}
}

// Equals checks if two WebViews are equal.
func (w *WebView) Equals(other Object) *Bool {
	if other == nil {
		return FALSE
	}
	if otherWv, ok := other.(*WebView); ok {
		return &Bool{Value: w.Handle == otherWv.Handle}
	}
	return FALSE
}

// IsClosed returns true if the WebView is closed.
func (w *WebView) IsClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

// SetClosed sets the closed state.
func (w *WebView) SetClosed(closed bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = closed
}

// AddCallback adds a callback function.
func (w *WebView) AddCallback(name string, fn *Function) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks[name] = fn
}

// GetCallback gets a callback function.
func (w *WebView) GetCallback(name string) (*Function, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fn, ok := w.callbacks[name]
	return fn, ok
}

// Close closes the WebView.
func (w *WebView) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		w.callbacks = nil
	}
}
