//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains COM callback implementations for WebView2.
package webview2

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// ComProc stores a COM procedure.
type ComProc uintptr

// NewComProc creates a new COM proc from a Go function.
func NewComProc(fn interface{}) ComProc {
	return ComProc(syscall.NewCallback(fn))
}

//go:uintptrescapes
// Call calls a COM procedure.
// The go:uintptrescapes directive prevents the Go compiler from moving pointers
// during the call, which is critical for COM interop.
func (p ComProc) Call(a ...uintptr) (r1, r2 uintptr, lastErr error) {
	return syscall.SyscallN(uintptr(p), a...)
}

// Initialize COM on the main thread
func init() {
	runtime.LockOSThread()
}

// ============================================================
// Environment Completed Handler
// ============================================================

// EnvironmentHandlerVTable is the vtable for environment creation callback.
type EnvironmentHandlerVTable struct {
	QueryInterface ComProc
	AddRef         ComProc
	Release        ComProc
	Invoke         ComProc
}

// EnvironmentHandler handles environment creation callback.
// IMPORTANT: vtbl must be the first field for COM compatibility.
type EnvironmentHandler struct {
	vtbl *EnvironmentHandlerVTable
	wv   *WebView2
	ref  int32
}

// Global vtable for environment handlers (shared by all instances)
var environmentHandlerVTable = EnvironmentHandlerVTable{
	QueryInterface: NewComProc(environmentQueryInterface),
	AddRef:         NewComProc(environmentAddRef),
	Release:        NewComProc(environmentRelease),
	Invoke:         NewComProc(environmentInvoke),
}

func newEnvironmentHandler(wv *WebView2) *EnvironmentHandler {
	return &EnvironmentHandler{
		vtbl: &environmentHandlerVTable,
		wv:   wv,
		ref:  1,
	}
}

func environmentQueryInterface(this *EnvironmentHandler, riid, ppvObject uintptr) uintptr {
	if ppvObject == 0 {
		return uintptr(E_POINTER)
	}
	// Return self for IUnknown and our interface
	*(*uintptr)(unsafe.Pointer(ppvObject)) = uintptr(unsafe.Pointer(this))
	return S_OK
}

func environmentAddRef(this *EnvironmentHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, 1))
}

func environmentRelease(this *EnvironmentHandler) uintptr {
	newRef := atomic.AddInt32(&this.ref, -1)
	return uintptr(newRef)
}

func environmentInvoke(this *EnvironmentHandler, hr uintptr, env *ICoreWebView2Environment) uintptr {
	if hr != S_OK || env == nil {
		// Signal failure - the main loop will detect controller is nil
		return S_OK
	}

	// AddRef the environment
	syscall.SyscallN(env.vtbl.AddRef, uintptr(unsafe.Pointer(env)))
	this.wv.env = env

	// Create controller from within the callback
	ret, _, _ := syscall.SyscallN(
		env.vtbl.CreateCoreWebView2Controller,
		uintptr(unsafe.Pointer(env)),
		this.wv.hwnd,
		uintptr(unsafe.Pointer(this.wv.controllerHandler)),
	)

	// Return value is checked in controller callback
	_ = ret
	return S_OK
}

// ============================================================
// Controller Completed Handler
// ============================================================

// ControllerHandlerVTable is the vtable for controller creation callback.
type ControllerHandlerVTable struct {
	QueryInterface ComProc
	AddRef         ComProc
	Release        ComProc
	Invoke         ComProc
}

// ControllerHandler handles controller creation callback.
// IMPORTANT: vtbl must be the first field for COM compatibility.
type ControllerHandler struct {
	vtbl *ControllerHandlerVTable
	wv   *WebView2
	ref  int32
}

// Global vtable for controller handlers (shared by all instances)
var controllerHandlerVTable = ControllerHandlerVTable{
	QueryInterface: NewComProc(controllerQueryInterface),
	AddRef:         NewComProc(controllerAddRef),
	Release:        NewComProc(controllerRelease),
	Invoke:         NewComProc(controllerInvoke),
}

func newControllerHandler(wv *WebView2) *ControllerHandler {
	return &ControllerHandler{
		vtbl: &controllerHandlerVTable,
		wv:   wv,
		ref:  1,
	}
}

func controllerQueryInterface(this *ControllerHandler, riid, ppvObject uintptr) uintptr {
	if ppvObject == 0 {
		return uintptr(E_POINTER)
	}
	*(*uintptr)(unsafe.Pointer(ppvObject)) = uintptr(unsafe.Pointer(this))
	return S_OK
}

func controllerAddRef(this *ControllerHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, 1))
}

func controllerRelease(this *ControllerHandler) uintptr {
	newRef := atomic.AddInt32(&this.ref, -1)
	return uintptr(newRef)
}

func controllerInvoke(this *ControllerHandler, hr uintptr, controller *ICoreWebView2Controller) uintptr {
	if hr != S_OK || controller == nil {
		// Signal that initialization is done (but failed)
		atomic.StoreUintptr(&this.wv.inited, 1)
		return S_OK
	}

	// AddRef the controller
	syscall.SyscallN(controller.vtbl.AddRef, uintptr(unsafe.Pointer(controller)))
	this.wv.controller = controller

	// Get ICoreWebView2 from controller
	// GetCoreWebView2 returns an ICoreWebView2* (interface pointer)
	// Pass pointer to the webview field directly
	syscall.SyscallN(
		controller.vtbl.GetCoreWebView2,
		uintptr(unsafe.Pointer(controller)),
		uintptr(unsafe.Pointer(&this.wv.webview)),
	)

	// AddRef the webview
	if this.wv.webview != nil {
		syscall.SyscallN(this.wv.webview.vtbl.AddRef, uintptr(unsafe.Pointer(this.wv.webview)))
	}

	// Signal that initialization is complete
	atomic.StoreUintptr(&this.wv.inited, 1)
	return S_OK
}

// ============================================================
// Web Message Handler
// ============================================================

// MessageHandlerVTable is the vtable for web message callback.
type MessageHandlerVTable struct {
	QueryInterface ComProc
	AddRef         ComProc
	Release        ComProc
	Invoke         ComProc
}

// MessageHandler handles web message callback.
type MessageHandler struct {
	vtbl *MessageHandlerVTable
	wv   *WebView2
	ref  int32
}

var messageHandlerVTable = MessageHandlerVTable{
	QueryInterface: NewComProc(messageQueryInterface),
	AddRef:         NewComProc(messageAddRef),
	Release:        NewComProc(messageRelease),
	Invoke:         NewComProc(messageInvoke),
}

func newMessageHandler(wv *WebView2) *MessageHandler {
	return &MessageHandler{
		vtbl: &messageHandlerVTable,
		wv:   wv,
		ref:  1,
	}
}

func messageQueryInterface(this *MessageHandler, riid, ppvObject uintptr) uintptr {
	if ppvObject == 0 {
		return uintptr(E_POINTER)
	}
	*(*uintptr)(unsafe.Pointer(ppvObject)) = uintptr(unsafe.Pointer(this))
	return S_OK
}

func messageAddRef(this *MessageHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, 1))
}

func messageRelease(this *MessageHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, -1))
}

func messageInvoke(this *MessageHandler, sender uintptr, args uintptr) uintptr {
	// Use defer/recover to catch any panic from invalid pointer access
	defer func() {
		if r := recover(); r != nil {
			// Silently ignore conversion failures
		}
	}()

	if this == nil || this.wv == nil {
		return S_OK
	}

	// Validate args pointer
	if args == 0 {
		return S_OK
	}

	// Get message as JSON - use the vtable from interfaces.go
	argsVtbl := *(**ICoreWebView2WebMessageReceivedEventArgsVTable)(unsafe.Pointer(args))
	if argsVtbl == nil || argsVtbl.GetWebMessageAsJson == 0 {
		return S_OK
	}

	// GetWebMessageAsJson returns LPWSTR (not BSTR!), which is a null-terminated UTF-16 string
	// The memory is owned by the args object and should NOT be freed by us
	var msgPtr *uint16
	hr, _, _ := syscall.SyscallN(uintptr(argsVtbl.GetWebMessageAsJson),
		args, uintptr(unsafe.Pointer(&msgPtr)))

	// Check HRESULT
	if hr != 0 || msgPtr == nil {
		return S_OK
	}

	// Convert LPWSTR to Go string
	msg := LPCWSTRToString(msgPtr)

	// Only process if we got a valid message
	if msg != "" {
		this.wv.handleMessage(msg)
	}

	return S_OK
}

// ============================================================
// Execute Script Handler
// ============================================================

// ScriptHandlerVTable is the vtable for script execution callback.
type ScriptHandlerVTable struct {
	QueryInterface ComProc
	AddRef         ComProc
	Release        ComProc
	Invoke         ComProc
}

// ScriptHandler handles script execution callback.
type ScriptHandler struct {
	vtbl     *ScriptHandlerVTable
	wv       *WebView2
	ref      int32
	id       int64
	callback func(string, error)
}

var scriptHandlerVTable = ScriptHandlerVTable{
	QueryInterface: NewComProc(scriptQueryInterface),
	AddRef:         NewComProc(scriptAddRef),
	Release:        NewComProc(scriptRelease),
	Invoke:         NewComProc(scriptInvoke),
}

func newScriptHandler(wv *WebView2, id int64, callback func(string, error)) *ScriptHandler {
	return &ScriptHandler{
		vtbl:     &scriptHandlerVTable,
		wv:       wv,
		ref:      1,
		id:       id,
		callback: callback,
	}
}

func scriptQueryInterface(this *ScriptHandler, riid, ppvObject uintptr) uintptr {
	if ppvObject == 0 {
		return uintptr(E_POINTER)
	}
	*(*uintptr)(unsafe.Pointer(ppvObject)) = uintptr(unsafe.Pointer(this))
	return S_OK
}

func scriptAddRef(this *ScriptHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, 1))
}

func scriptRelease(this *ScriptHandler) uintptr {
	return uintptr(atomic.AddInt32(&this.ref, -1))
}

func scriptInvoke(this *ScriptHandler, hr uintptr, resultObjectAsJson uintptr) uintptr {
	if this == nil {
		return S_OK
	}

	// For fire-and-forget (nil callback), skip BSTR processing entirely
	// This avoids potential crashes from invalid BSTR pointers
	if this.callback == nil {
		// Just return - no need to process result
		return S_OK
	}

	// Use defer/recover to catch any panic from invalid BSTR access
	defer func() {
		if r := recover(); r != nil {
			this.callback("", fmt.Errorf("BSTR conversion panic: %v", r))
		}
	}()

	if hr == S_OK && resultObjectAsJson != 0 {
		// resultObjectAsJson is a BSTR (which is uintptr)
		// Add comprehensive safety check for BSTR
		// Valid user-space addresses in Windows are typically 0x10000 to 0x7FFEFFFF
		if resultObjectAsJson > 0x10000 && resultObjectAsJson < 0x7FFEFFFF {
			// Try to validate the BSTR by checking if we can read the length prefix
			// BSTR has a 4-byte length prefix at address -4 from the data pointer
			lenPtr := resultObjectAsJson - 4
			if lenPtr > 0x10000 && lenPtr < 0x7FFEFFFF {
				result := BSTRToString(BSTR(resultObjectAsJson))
				this.callback(result, nil)
			} else {
				this.callback("", nil)
			}
		} else {
			this.callback("", nil)
		}
	} else if hr != S_OK {
		this.callback("", syscall.Errno(hr))
	}
	return S_OK
}

// ============================================================
// Handler creation methods for WebView2
// ============================================================

var handlerInitOnce sync.Once

func (wv *WebView2) createEnvironmentCompletedHandler() *EnvironmentHandler {
	handlerInitOnce.Do(func() {
		// Ensure vtables are initialized
	})
	return newEnvironmentHandler(wv)
}

func (wv *WebView2) createControllerCompletedHandler() *ControllerHandler {
	return newControllerHandler(wv)
}

func (wv *WebView2) createWebMessageHandler() *MessageHandler {
	return newMessageHandler(wv)
}

func (wv *WebView2) createExecuteScriptHandler(id int64, callback func(string, error)) *ScriptHandler {
	return newScriptHandler(wv, id, callback)
}