//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains COM callback implementations for WebView2.
package webview2

import (
	"sync"
	"syscall"
	"unsafe"
)

// Callback handler structures and VTables

// EnvironmentCompletedHandler handles environment creation callback.
type EnvironmentCompletedHandler struct {
	vtbl  *EnvironmentCompletedHandlerVTable
	ref   int32
	wv    *WebView2
}

type EnvironmentCompletedHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ControllerCompletedHandler handles controller creation callback.
type ControllerCompletedHandler struct {
	vtbl  *ControllerCompletedHandlerVTable
	ref   int32
	wv    *WebView2
}

type ControllerCompletedHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// WebMessageHandler handles web message callback.
type WebMessageHandler struct {
	vtbl  *WebMessageHandlerVTable
	ref   int32
	wv    *WebView2
}

type WebMessageHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ExecuteScriptHandler handles script execution callback.
type ExecuteScriptHandler struct {
	vtbl  *ExecuteScriptHandlerVTable
	ref   int32
	wv    *WebView2
	id    int64
	callback func(string, error)
}

type ExecuteScriptHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

var (
	handlerVTablesInit sync.Once

	envHandlerVTable    EnvironmentCompletedHandlerVTable
	ctrlHandlerVTable   ControllerCompletedHandlerVTable
	msgHandlerVTable    WebMessageHandlerVTable
	scriptHandlerVTable ExecuteScriptHandlerVTable
)

// initHandlerVTables initializes the VTables for callbacks.
func initHandlerVTables() {
	handlerVTablesInit.Do(func() {
		envHandlerVTable = EnvironmentCompletedHandlerVTable{
			QueryInterface: syscall.NewCallback(environmentQueryInterface),
			AddRef:         syscall.NewCallback(environmentAddRef),
			Release:        syscall.NewCallback(environmentRelease),
			Invoke:         syscall.NewCallback(environmentInvoke),
		}

		ctrlHandlerVTable = ControllerCompletedHandlerVTable{
			QueryInterface: syscall.NewCallback(controllerQueryInterface),
			AddRef:         syscall.NewCallback(controllerAddRef),
			Release:        syscall.NewCallback(controllerRelease),
			Invoke:         syscall.NewCallback(controllerInvoke),
		}

		msgHandlerVTable = WebMessageHandlerVTable{
			QueryInterface: syscall.NewCallback(messageQueryInterface),
			AddRef:         syscall.NewCallback(messageAddRef),
			Release:        syscall.NewCallback(messageRelease),
			Invoke:         syscall.NewCallback(messageInvoke),
		}

		scriptHandlerVTable = ExecuteScriptHandlerVTable{
			QueryInterface: syscall.NewCallback(scriptQueryInterface),
			AddRef:         syscall.NewCallback(scriptAddRef),
			Release:        syscall.NewCallback(scriptRelease),
			Invoke:         syscall.NewCallback(scriptInvoke),
		}
	})
}

// createEnvironmentCompletedHandler creates a new environment completed handler.
func (wv *WebView2) createEnvironmentCompletedHandler() *EnvironmentCompletedHandler {
	initHandlerVTables()
	return &EnvironmentCompletedHandler{
		vtbl: &envHandlerVTable,
		ref:  1,
		wv:   wv,
	}
}

// createControllerCompletedHandler creates a new controller completed handler.
func (wv *WebView2) createControllerCompletedHandler() *ControllerCompletedHandler {
	initHandlerVTables()
	return &ControllerCompletedHandler{
		vtbl: &ctrlHandlerVTable,
		ref:  1,
		wv:   wv,
	}
}

// createWebMessageHandler creates a new web message handler.
func (wv *WebView2) createWebMessageHandler() *WebMessageHandler {
	initHandlerVTables()
	return &WebMessageHandler{
		vtbl: &msgHandlerVTable,
		ref:  1,
		wv:   wv,
	}
}

// createExecuteScriptHandler creates a new execute script handler.
func (wv *WebView2) createExecuteScriptHandler(id int64, callback func(string, error)) *ExecuteScriptHandler {
	initHandlerVTables()
	return &ExecuteScriptHandler{
		vtbl:     &scriptHandlerVTable,
		ref:      1,
		wv:       wv,
		id:       id,
		callback: callback,
	}
}

// Environment handler callbacks

func environmentQueryInterface(this *EnvironmentCompletedHandler, riid *syscall.GUID, ppvObject *uintptr) uintptr {
	*ppvObject = uintptr(unsafe.Pointer(this))
	return S_OK
}

func environmentAddRef(this *EnvironmentCompletedHandler) uintptr {
	this.ref++
	return uintptr(this.ref)
}

func environmentRelease(this *EnvironmentCompletedHandler) uintptr {
	this.ref--
	return uintptr(this.ref)
}

func environmentInvoke(this *EnvironmentCompletedHandler, hr uintptr, env *ICoreWebView2Environment) uintptr {
	if hr == S_OK && env != nil {
		this.wv.env = env
		this.wv.envCreated <- nil
	} else {
		this.wv.envCreated <- syscall.Errno(hr)
	}
	return S_OK
}

// Controller handler callbacks

func controllerQueryInterface(this *ControllerCompletedHandler, riid *syscall.GUID, ppvObject *uintptr) uintptr {
	*ppvObject = uintptr(unsafe.Pointer(this))
	return S_OK
}

func controllerAddRef(this *ControllerCompletedHandler) uintptr {
	this.ref++
	return uintptr(this.ref)
}

func controllerRelease(this *ControllerCompletedHandler) uintptr {
	this.ref--
	return uintptr(this.ref)
}

func controllerInvoke(this *ControllerCompletedHandler, hr uintptr, controller *ICoreWebView2Controller) uintptr {
	if hr == S_OK && controller != nil {
		this.wv.controller = controller

		// Get ICoreWebView2 from controller
		var webviewPtr uintptr
		syscall.Syscall(controller.vtbl.GetCoreWebView2, 2,
			uintptr(unsafe.Pointer(controller)),
			uintptr(unsafe.Pointer(&webviewPtr)), 0)

		this.wv.webview = &ICoreWebView2{
			vtbl: (*ICoreWebView2VTable)(unsafe.Pointer(webviewPtr)),
		}

		this.wv.controllerCreated <- nil
	} else {
		this.wv.controllerCreated <- syscall.Errno(hr)
	}
	return S_OK
}

// Web message handler callbacks

func messageQueryInterface(this *WebMessageHandler, riid *syscall.GUID, ppvObject *uintptr) uintptr {
	*ppvObject = uintptr(unsafe.Pointer(this))
	return S_OK
}

func messageAddRef(this *WebMessageHandler) uintptr {
	this.ref++
	return uintptr(this.ref)
}

func messageRelease(this *WebMessageHandler) uintptr {
	this.ref--
	return uintptr(this.ref)
}

func messageInvoke(this *WebMessageHandler, sender uintptr, args *ICoreWebView2WebMessageReceivedEventArgs) uintptr {
	if this.wv == nil {
		return S_OK
	}

	// Get message as JSON
	argsVtbl := (*ICoreWebView2WebMessageReceivedEventArgsVTable)(unsafe.Pointer(args))
	var msgBSTR BSTR
	syscall.Syscall(argsVtbl.GetWebMessageAsJson, 2,
		uintptr(unsafe.Pointer(args)),
		uintptr(unsafe.Pointer(&msgBSTR)), 0)

	msg := BSTRToString(msgBSTR)
	FreeBSTR(msgBSTR)

	this.wv.handleMessage(msg)

	return S_OK
}

// ICoreWebView2WebMessageReceivedEventArgs (forward declaration)
type ICoreWebView2WebMessageReceivedEventArgs struct {
	vtbl *ICoreWebView2WebMessageReceivedEventArgsVTable
}

// Execute script handler callbacks

func scriptQueryInterface(this *ExecuteScriptHandler, riid *syscall.GUID, ppvObject *uintptr) uintptr {
	*ppvObject = uintptr(unsafe.Pointer(this))
	return S_OK
}

func scriptAddRef(this *ExecuteScriptHandler) uintptr {
	this.ref++
	return uintptr(this.ref)
}

func scriptRelease(this *ExecuteScriptHandler) uintptr {
	this.ref--
	return uintptr(this.ref)
}

func scriptInvoke(this *ExecuteScriptHandler, hr uintptr, resultObjectAsJson BSTR) uintptr {
	if this.callback != nil {
		if hr == S_OK {
			result := BSTRToString(resultObjectAsJson)
			this.callback(result, nil)
		} else {
			this.callback("", syscall.Errno(hr))
		}
	}
	FreeBSTR(resultObjectAsJson)
	return S_OK
}
