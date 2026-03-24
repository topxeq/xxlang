//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains COM infrastructure for Windows.
package webview2

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	ole32DLL       = syscall.NewLazyDLL("ole32.dll")
	coInitialize   = ole32DLL.NewProc("CoInitializeEx")
	coUninitialize = ole32DLL.NewProc("CoUninitialize")
	coTaskMemAlloc  = ole32DLL.NewProc("CoTaskMemAlloc")
	coTaskMemFree   = ole32DLL.NewProc("CoTaskMemFree")

	// COM initialization flags
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0

	// Common HRESULT values
	S_OK          uintptr = 0x00000000
	S_FALSE       uintptr = 0x00000001
	E_FAIL        uintptr = 0x80004005
	E_NOINTERFACE uintptr = 0x80004002
	E_POINTER     uintptr = 0x80004003
)

var (
	comInitialized bool
	comMutex       sync.Mutex
)

// COMInitialize initializes COM library.
// Should be called before using any COM objects.
func COMInitialize() error {
	comMutex.Lock()
	defer comMutex.Unlock()

	if comInitialized {
		return nil
	}

	hr, _, _ := coInitialize.Call(0, uintptr(COINIT_APARTMENTTHREADED))
	if hr != S_OK && hr != 0x80010106 { // S_OK or RPC_E_CHANGED_MODE
		return syscall.Errno(hr)
	}

	comInitialized = true
	return nil
}

// COMUninitialize uninitializes COM library.
func COMUninitialize() {
	comMutex.Lock()
	defer comMutex.Unlock()

	if comInitialized {
		coUninitialize.Call()
		comInitialized = false
	}
}

// IUnknownVTable represents the IUnknown vtable.
type IUnknownVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

// IUnknown represents the base COM interface.
type IUnknown struct {
	vtbl *IUnknownVTable
}

// QueryInterface queries for a specific interface.
func (u *IUnknown) QueryInterface(riid *syscall.GUID) (uintptr, error) {
	var result uintptr
	hr, _, _ := syscall.Syscall(
		u.vtbl.QueryInterface,
		3,
		uintptr(unsafe.Pointer(u)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(&result)),
	)
	if hr != S_OK {
		return 0, syscall.Errno(hr)
	}
	return result, nil
}

// AddRef increments the reference count.
func (u *IUnknown) AddRef() uint32 {
	ret, _, _ := syscall.Syscall(
		u.vtbl.AddRef,
		1,
		uintptr(unsafe.Pointer(u)),
		0, 0,
	)
	return uint32(ret)
}

// Release decrements the reference count.
func (u *IUnknown) Release() uint32 {
	ret, _, _ := syscall.Syscall(
		u.vtbl.Release,
		1,
		uintptr(unsafe.Pointer(u)),
		0, 0,
	)
	return uint32(ret)
}

// GUIDToString converts a GUID to its string representation.
func GUIDToString(g *syscall.GUID) string {
	return string(g.Data1>>24&0xFF) + string(g.Data1>>16&0xFF) +
		string(g.Data1>>8&0xFF) + string(g.Data1&0xFF)
}

// CoTaskMemAlloc allocates memory from COM task allocator.
func CoTaskMemAlloc(size uintptr) uintptr {
	ret, _, _ := coTaskMemAlloc.Call(size)
	return ret
}

// CoTaskMemFree frees memory from COM task allocator.
func CoTaskMemFree(ptr uintptr) {
	coTaskMemFree.Call(ptr)
}
