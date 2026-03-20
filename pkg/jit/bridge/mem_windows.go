// +build windows

package bridge

import (
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	virtualAllocProc = kernel32.NewProc("VirtualAlloc")
	virtualFreeProc  = kernel32.NewProc("VirtualFree")
)

const (
	MEM_COMMIT     = 0x1000
	MEM_RESERVE    = 0x2000
	MEM_RELEASE    = 0x8000
	PAGE_READWRITE = 0x04
	PAGE_EXECUTE   = 0x10
	PAGE_EXECUTE_READWRITE = 0x40
)

// AllocExecMem allocates executable memory on Windows.
// Uses VirtualAlloc with PAGE_EXECUTE_READWRITE protection.
func AllocExecMem(size int) ([]byte, error) {
	// VirtualAlloc returns a pointer to the allocated memory
	ret, _, err := virtualAllocProc.Call(
		0,                             // lpAddress (NULL = system decides)
		uintptr(size),                 // dwSize
		uintptr(MEM_COMMIT|MEM_RESERVE), // flAllocationType
		uintptr(PAGE_EXECUTE_READWRITE), // flProtect
	)

	if ret == 0 {
		return nil, err
	}

	// Convert to slice
	var mem []byte
	header := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(&mem))
	header.Data = ret
	header.Len = size
	header.Cap = size

	return mem, nil
}

// FreeExecMem frees executable memory allocated by AllocExecMem.
func FreeExecMem(mem []byte) error {
	if len(mem) == 0 {
		return nil
	}

	header := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(&mem))

	ret, _, err := virtualFreeProc.Call(
		header.Data,
		0,
		uintptr(MEM_RELEASE),
	)

	if ret == 0 {
		return err
	}

	return nil
}
