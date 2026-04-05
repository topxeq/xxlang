//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains BSTR string handling utilities.
package webview2

import (
	"syscall"
	"unsafe"
)

// maxValidPointer is the maximum valid user-space pointer address.
// On 64-bit Windows, user-space addresses are typically below 0x7fff00000000.
// On 32-bit Windows, the maximum is 0x7fffffff (2GB user space).
const maxValidPointer = uintptr(1<<(unsafe.Sizeof(uintptr(0))*8-1) - 1)

var (
	oleaut32DLL      = syscall.NewLazyDLL("oleaut32.dll")
	sysAllocString   = oleaut32DLL.NewProc("SysAllocString")
	sysAllocStringByteLen = oleaut32DLL.NewProc("SysAllocStringByteLen")
	sysFreeString    = oleaut32DLL.NewProc("SysFreeString")
	sysStringLen     = oleaut32DLL.NewProc("SysStringLen")
	sysStringByteLen = oleaut32DLL.NewProc("SysStringByteLen")
)

// BSTR represents a Windows BSTR string.
// BSTR is a length-prefixed UTF-16 string used in COM.
type BSTR uintptr

// BSTRToString converts a Windows BSTR to a Go string.
// Returns empty string if bstr is nil or invalid.
func BSTRToString(bstr BSTR) string {
	if bstr == 0 {
		return ""
	}

	// Validate BSTR pointer - should be 4-byte aligned and not too low
	if bstr < 0x10000 {
		return ""
	}

	// Validate pointer is in reasonable range (not too high)
	// Typical Windows user-space addresses are below maxValidPointer
	if uintptr(bstr) > maxValidPointer {
		return ""
	}

	// Try to get string length - this will fail if BSTR is invalid
	lenRet, _, err := sysStringLen.Call(uintptr(bstr))
	if lenRet == 0 {
		return ""
	}

	// Check for syscall error (indicates invalid pointer)
	// Note: err might be nil on success, not Errno(0)
	if err != nil && err != syscall.Errno(0) {
		return ""
	}

	// Sanity check: BSTR length should be reasonable
	// Max 1MB of UTF-16 characters
	const maxLen = 512 * 1024
	if lenRet > maxLen {
		return ""
	}

	length := int(lenRet)

	// Try a simpler approach using UTF16PtrToString with error handling
	// BSTR data starts at the pointer location
	utf16Ptr := (*uint16)(unsafe.Pointer(bstr))

	// Use a safer slice-based approach with bounds checking
	utf16Slice := make([]uint16, 0, length+1)
	for i := 0; i < length; i++ {
		// Access each character with explicit pointer arithmetic
		charPtr := (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(utf16Ptr)) + uintptr(i)*2))
		// Check if pointer is still valid (basic sanity)
		if uintptr(unsafe.Pointer(charPtr)) > maxValidPointer || uintptr(unsafe.Pointer(charPtr)) < 0x10000 {
			// Invalid pointer encountered, return what we have
			break
		}
		utf16Slice = append(utf16Slice, *charPtr)
	}
	utf16Slice = append(utf16Slice, 0) // null terminate

	return syscall.UTF16ToString(utf16Slice)
}

// BSTRToBytes converts a Windows BSTR to a byte slice.
func BSTRToBytes(bstr BSTR) []byte {
	if bstr == 0 {
		return nil
	}

	// Get byte length
	lenRet, _, _ := sysStringByteLen.Call(uintptr(bstr))
	if lenRet == 0 {
		return nil
	}

	ptr := (*[0xffff]byte)(unsafe.Pointer(bstr))[:lenRet:lenRet]
	return append([]byte(nil), ptr...)
}

// StringToBSTR converts a Go string to a Windows BSTR.
// Caller must free the BSTR using FreeBSTR when done.
func StringToBSTR(s string) BSTR {
	if s == "" {
		return 0
	}

	utf16, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return 0
	}

	ret, _, _ := sysAllocString.Call(uintptr(unsafe.Pointer(utf16)))
	return BSTR(ret)
}

// BytesToBSTR converts a byte slice to a Windows BSTR.
func BytesToBSTR(b []byte) BSTR {
	if len(b) == 0 {
		return 0
	}

	ret, _, _ := sysAllocStringByteLen.Call(
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(len(b)),
	)
	return BSTR(ret)
}

// FreeBSTR frees a BSTR allocated by StringToBSTR or BytesToBSTR.
func FreeBSTR(bstr BSTR) {
	if bstr != 0 {
		sysFreeString.Call(uintptr(bstr))
	}
}

// StringToLPCWSTR converts a Go string to LPCWSTR (UTF-16 pointer).
// Returns a pointer that must be kept alive for the duration of use.
func StringToLPCWSTR(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

// LPCWSTRToString converts a LPCWSTR to a Go string.
func LPCWSTRToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}

	// Find length
	var length int
	for p := ptr; *p != 0; p = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + 2)) {
		length++
	}

	if length == 0 {
		return ""
	}

	// Convert to slice and then to string
	sl := (*[0xffff]uint16)(unsafe.Pointer(ptr))[:length:length]
	return syscall.UTF16ToString(sl)
}

// StringToLPWSTR converts a Go string to a mutable UTF-16 buffer.
// Caller must free the returned buffer using CoTaskMemFree.
func StringToLPWSTR(s string) *uint16 {
	utf16, _ := syscall.UTF16FromString(s)
	buf := CoTaskMemAlloc(uintptr(len(utf16) * 2))
	if buf == 0 {
		return nil
	}
	dst := (*[0xffff]uint16)(unsafe.Pointer(buf))[:len(utf16):len(utf16)]
	copy(dst, utf16)
	return (*uint16)(unsafe.Pointer(buf))
}
