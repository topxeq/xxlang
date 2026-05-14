//go:build windows && amd64
// +build windows,amd64

package jit

// callNative calls native code with only the globals pointer.
func callNative(entry uintptr, globals *int64) int64

// callNativeWithArgs calls native code with globals and up to three arguments.
func callNativeWithArgs(entry uintptr, globals *int64, arg0, arg1, arg2 int64) int64

// callNativeWithArgs8 calls native code with globals and up to eight arguments.
func callNativeWithArgs8(entry uintptr, globals *int64, args *int64) int64

func callNativeWithGlobals(entry uintptr, globals []int64) int64 {
	if len(globals) == 0 {
		return callNative(entry, nil)
	}

	return callNative(entry, &globals[0])
}
