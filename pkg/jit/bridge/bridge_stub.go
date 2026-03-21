// +build !amd64,!windows

package bridge

import "unsafe"

// AllocExecMem stub for non-amd64
func AllocExecMem(size int) ([]byte, error) {
	return nil, nil
}

// FreeExecMem stub for non-amd64
func FreeExecMem(mem []byte) error {
	return nil
}

// BuildFibCode stub for non-amd64
func BuildFibCode() []byte {
	return nil
}

// Call0 stub for non-amd64
func Call0(fn *byte) int64 {
	_ = unsafe.Pointer(fn)
	return 0
}

// Call1 stub for non-amd64
func Call1(fn *byte, arg1 int64) int64 {
	_ = unsafe.Pointer(fn)
	_ = arg1
	return 0
}

// Call2 stub for non-amd64
func Call2(fn *byte, arg1, arg2 int64) int64 {
	_ = unsafe.Pointer(fn)
	_ = arg1
	_ = arg2
	return 0
}

// Call3 stub for non-amd64
func Call3(fn *byte, arg1, arg2, arg3 int64) int64 {
	_ = unsafe.Pointer(fn)
	_ = arg1
	_ = arg2
	_ = arg3
	return 0
}
