//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains COM infrastructure for Windows.
package webview2

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	ole32DLL       = syscall.NewLazyDLL("ole32.dll")
	coInitialize   = ole32DLL.NewProc("CoInitializeEx")
	coUninitialize = ole32DLL.NewProc("CoUninitialize")
	coTaskMemAlloc = ole32DLL.NewProc("CoTaskMemAlloc")
	coTaskMemFree  = ole32DLL.NewProc("CoTaskMemFree")

	// WinRT initialization (combase.dll)
	combaseDLL     = syscall.NewLazyDLL("combase.dll")
	roInitialize   = combaseDLL.NewProc("RoInitialize")
	roUninitialize = combaseDLL.NewProc("RoUninitialize")

	// COM initialization flags
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0

	// WinRT initialization flags
	RO_INIT_SINGLETHREADED = 0x1
	RO_INIT_MULTITHREADED  = 0x0

	// Common HRESULT values
	S_OK          uintptr = 0x00000000
	S_FALSE       uintptr = 0x00000001
	E_FAIL        uintptr = 0x80004005
	E_NOINTERFACE uintptr = 0x80004002
	E_POINTER     uintptr = 0x80004003
)

var (
	comInitialized bool
	roInitialized  bool
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

	// Initialize COM in STA mode (WebView2 requires STA)
	hr, _, _ := coInitialize.Call(0, uintptr(COINIT_APARTMENTTHREADED))
	// RPC_E_CHANGED_MODE (0x80010106) means COM was already initialized in a different mode
	// We need to accept this and continue
	if hr != S_OK && hr != 0x80010106 {
		return syscall.Errno(hr)
	}

	comInitialized = true

	// Initialize WinRT (Windows Runtime) for WebView2
	// WebView2 uses WinRT components internally
	if !roInitialized {
		hr, _, _ = roInitialize.Call(uintptr(RO_INIT_SINGLETHREADED))
		// S_OK, S_FALSE (already initialized), or RPC_E_CHANGED_MODE are all acceptable
		if hr == S_OK || hr == S_FALSE || hr == 0x80010106 {
			roInitialized = true
		}
	}

	return nil
}

// COMUninitialize uninitializes COM library.
func COMUninitialize() {
	comMutex.Lock()
	defer comMutex.Unlock()

	if roInitialized {
		roUninitialize.Call()
		roInitialized = false
	}

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
// This creates a proper GUID string format like "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
func GUIDToString(g *syscall.GUID) string {
	// Use fmt.Sprintf for proper hex formatting
	// This matches the standard GUID string format
	return fmt.Sprintf(
		"%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X",
		g.Data1,
		g.Data2,
		g.Data3,
		g.Data4[0], g.Data4[1], g.Data4[2], g.Data4[3],
		g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7],
	)
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
