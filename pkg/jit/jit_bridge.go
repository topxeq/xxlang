// +build amd64,!windows

// pkg/jit/jit_bridge.go
// JIT code execution bridge that avoids Go stack management issues
package jit

import (
	"unsafe"
)

// CallNativeCode calls JIT-compiled native code
// This is a wrapper that avoids Go's stack management issues
func CallNativeCode(codePtr uintptr, arg int64) int64 {
	// Use inline assembly to call the code without Go stack management
	// The key is to avoid any Go function calls between us and the native code
	return callNativeCodeImpl(codePtr, arg)
}

//go:noescape
//go:nosplit
func callNativeCodeImpl(codePtr uintptr, arg int64) int64

// SimpleNativeCall provides a simple way to call JIT code
// It uses a minimal wrapper to avoid stack issues
func SimpleNativeCall(code []byte, arg int64) int64 {
	if len(code) == 0 {
		return 0
	}

	// Get pointer to code
	codePtr := uintptr(unsafe.Pointer(&code[0]))

	// Call the native code
	return callNativeCodeImpl(codePtr, arg)
}
