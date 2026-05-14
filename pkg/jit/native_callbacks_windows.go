//go:build windows && amd64
// +build windows,amd64

package jit

// Windows currently uses the pure-native subset without VM callback trampolines.

func callBuiltinCallback(callback uintptr, builtinIdx, numArgs int, argsPtr *int64) int64

func callFunctionCallback(callback uintptr, funcReg, numArgs int, argsPtr *int64) int64

func callCollectionCallback(callback uintptr, opKind, numArgs int, argsPtr *int64) int64

func callObjectCallback(callback uintptr, opKind, numArgs int, argsPtr *int64, nameIdx int) int64

func builtinCallbackWrapper(builtinIdx, numArgs int64, argsPtr *int64) int64

func functionCallbackWrapper(funcReg, numArgs int64, argsPtr *int64) int64

func collectionCallbackWrapper(opKind, numArgs int64, argsPtr *int64) int64

func objectCallbackWrapper(opKind, numArgs int64, argsPtr *int64, nameIdx int64) int64

func GetBuiltinCallbackPtr() uintptr {
	return 0
}

func GetFunctionCallbackPtr() uintptr {
	return 0
}

func GetCollectionCallbackPtr() uintptr {
	return 0
}

func GetObjectCallbackPtr() uintptr {
	return 0
}
