// +build !windows,amd64

package bridge

import (
	"syscall"
)

// AllocExecMem allocates executable memory.
// On Unix systems, uses mmap with PROT_READ|PROT_WRITE|PROT_EXEC.
func AllocExecMem(size int) ([]byte, error) {
	// Round up to page size
	pageSize := syscall.Getpagesize()
	allocSize := (size + pageSize - 1) & ^(pageSize - 1)

	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	flags := syscall.MAP_ANON | syscall.MAP_PRIVATE

	return syscall.Mmap(-1, 0, allocSize, prot, flags)
}

// FreeExecMem frees executable memory allocated by AllocExecMem.
func FreeExecMem(mem []byte) error {
	return syscall.Munmap(mem)
}
