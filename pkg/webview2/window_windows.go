//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains Win32 window management for WebView2.
package webview2

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32DLL       = syscall.NewLazyDLL("user32.dll")
	registerClassEx = user32DLL.NewProc("RegisterClassExW")
	createWindowEx  = user32DLL.NewProc("CreateWindowExW")
	defWindowProc   = user32DLL.NewProc("DefWindowProcW")
	getMessage      = user32DLL.NewProc("GetMessageW")
	translateMessage = user32DLL.NewProc("TranslateMessage")
	dispatchMessage  = user32DLL.NewProc("DispatchMessageW")
	postQuitMessage  = user32DLL.NewProc("PostQuitMessage")
	getClientRect    = user32DLL.NewProc("GetClientRect")
	setWindowPos     = user32DLL.NewProc("SetWindowPos")
	destroyWindow    = user32DLL.NewProc("DestroyWindow")
	postMessage      = user32DLL.NewProc("PostMessageW")

	// Window messages
	WM_DESTROY   uint32 = 0x0002
	WM_SIZE      uint32 = 0x0005
	WM_CLOSE     uint32 = 0x0010
	WM_USER      uint32 = 0x0400

	// Window styles
	WS_OVERLAPPED       = 0x00000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_VISIBLE          = 0x10000000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX

	// Show window commands
	SW_SHOW uintptr = 5
)

// WndClassEx represents the WNDCLASSEX structure.
type WndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// Msg represents the MSG structure.
type Msg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       [2]int32
}

var (
	windowInstances     = make(map[uintptr]*WebView2)
	windowInstancesLock sync.RWMutex
	className           = syscall.StringToUTF16Ptr("XxlangWebView2Class")
	classRegistered     bool
)

// createWindow creates a Win32 window for WebView2.
func (wv *WebView2) createWindow(title string, width, height int) (uintptr, error) {
	// Register window class if not already done
	if !classRegistered {
		if err := registerWindowClass(); err != nil {
			return 0, err
		}
		classRegistered = true
	}

	// Get module handle
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getModuleHandle := kernel32.NewProc("GetModuleHandleW")
	hInstance, _, _ := getModuleHandle.Call(0)

	// Create window
	titlePtr := syscall.StringToUTF16Ptr(title)

	// Use WS_CHILD style to create a child window for WebView2
	// Or use WS_OVERLAPPEDWINDOW for a top-level window
	hwnd, _, err := createWindowEx.Call(
		0, // dwExStyle
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(WS_OVERLAPPEDWINDOW|WS_VISIBLE),
		0x80000000, // CW_USEDEFAULT
		0x80000000, // CW_USEDEFAULT
		uintptr(width),
		uintptr(height),
		0, // hWndParent
		0, // hMenu
		hInstance,
		0, // lpParam
	)

	if hwnd == 0 {
		return 0, err
	}

	// Store window instance
	windowInstancesLock.Lock()
	windowInstances[hwnd] = wv
	windowInstancesLock.Unlock()

	// Update window to ensure it's ready
	user32DLL.NewProc("UpdateWindow").Call(hwnd)

	return hwnd, nil
}

// registerWindowClass registers the window class.
func registerWindowClass() error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getModuleHandle := kernel32.NewProc("GetModuleHandleW")
	hInstance, _, _ := getModuleHandle.Call(0)

	// Load cursor
	cursor, _, _ := user32DLL.NewProc("LoadCursorW").Call(0, 0x7F00) // IDC_ARROW

	wc := WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(WndClassEx{})),
		Style:         0x0003, // CS_HREDRAW | CS_VREDRAW
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       cursor,
		HbrBackground: 6, // COLOR_WINDOW+1
		LpszClassName: className,
	}

	ret, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return err
	}

	return nil
}

// wndProc is the window procedure.
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	windowInstancesLock.RLock()
	wv, exists := windowInstances[hwnd]
	windowInstancesLock.RUnlock()

	switch msg {
	case WM_SIZE:
		if exists && wv.controller != nil {
			// Resize WebView2
			var rect Rect
			getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
			syscall.Syscall(wv.controller.vtbl.PutBounds, 2,
				uintptr(unsafe.Pointer(wv.controller)),
				uintptr(unsafe.Pointer(&rect)), 0)
		}

	case WM_CLOSE:
		if exists {
			wv.Close()
		}
		postQuitMessage.Call(0)
		return 0

	case WM_DESTROY:
		windowInstancesLock.Lock()
		delete(windowInstances, hwnd)
		windowInstancesLock.Unlock()
		postQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// destroyWindow destroys the window.
func (wv *WebView2) destroyWindow() {
	if wv.hwnd != 0 {
		destroyWindow.Call(wv.hwnd)
		wv.hwnd = 0
	}
}

// runMessageLoop runs the Windows message loop.
func runMessageLoop(hwnd uintptr) {
	var msg Msg
	for {
		ret, _, _ := getMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)

		if ret == 0 {
			break
		}

		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// processSingleMessage processes a single Windows message without blocking.
// Returns true if a message was processed, false if no message was available.
func processSingleMessage() bool {
	// Use PeekMessageW to check for messages without blocking
	peekMessage := user32DLL.NewProc("PeekMessageW")

	var msg Msg
	ret, _, _ := peekMessage.Call(
		uintptr(unsafe.Pointer(&msg)),
		0, 0, 0,
		1, // PM_REMOVE
	)

	if ret != 0 {
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		return true
	}

	return false
}
