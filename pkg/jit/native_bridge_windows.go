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

// Legacy callback declarations for assembly linkage in bridge_amd64.s.
// These are not used on Windows — the actual callbacks use syscall.NewCallback
// with the Windows x64 ABI. The assembly implementations exist for non-Windows
// platforms and must have matching Go declarations to compile.

func callBuiltinCallback(callback uintptr, builtinIdx, numArgs int, argsPtr *int64) int64

func callFunctionCallback(callback uintptr, funcReg, numArgs int, argsPtr *int64) int64

func callCollectionCallback(callback uintptr, opKind, numArgs int, argsPtr *int64) int64

func callObjectCallback(callback uintptr, opKind, numArgs int, argsPtr *int64, nameIdx int) int64

func builtinCallbackWrapper(builtinIdx, numArgs int64, argsPtr *int64) int64

func functionCallbackWrapper(funcReg, numArgs int64, argsPtr *int64) int64

func collectionCallbackWrapper(opKind, numArgs int64, argsPtr *int64) int64

func objectCallbackWrapper(opKind, numArgs int64, argsPtr *int64, nameIdx int64) int64
