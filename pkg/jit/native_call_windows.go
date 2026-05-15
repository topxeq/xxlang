//go:build windows && amd64
// +build windows,amd64

package jit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// windowsCallCtx holds the VM context needed for native→Go call callbacks on Windows.
type windowsCallCtx struct {
	mu       sync.Mutex
	vm       *JITVM
	callback uintptr // syscall.NewCallback result
}

// globalCallCtx is the shared Windows callback context.
var globalCallCtx windowsCallCtx

// inSyscallN is atomically set to 1 while we are inside syscall.SyscallN.
// If a callback fires while inSyscallN==1, we know we are in a nested
// callback and must NOT call syscall.SyscallN again (deadlock risk).
var inSyscallN int64

// nativeCallFromNative is the Go callback invoked from native code for OpRegCall.
// Parameters follow Windows x64 ABI: rcx=constIdx, rdx=numArgs, r8=argsPtr.
// Returns the int64 result of the function call.
//
// IMPORTANT: This function is called from within syscall.SyscallN (goroutine
// is in _Gsyscall state). The functions it dispatches to must use bridge ABI
// or interpreter — they must NOT use syscall.SyscallN (nested syscall deadlock).
// This is guaranteed because compileNativeFunctions only compiles leaf functions
// (CanExecuteNatively rejects OpRegCall), so UseSyscallABI is always false.
func nativeCallFromNative(constIdx int, numArgs int, argsPtr *int64) int64 {
	globalCallCtx.mu.Lock()
	jitVM := globalCallCtx.vm
	globalCallCtx.mu.Unlock()

	if jitVM == nil {
		return 0
	}

	args := make([]vm.Value, numArgs)
	if argsPtr != nil && numArgs > 0 {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = vm.NewInt(argsSlice[i])
		}
	}

	if jitVM.bytecode == nil || constIdx < 0 || constIdx >= len(jitVM.bytecode.Constants) {
		return 0
	}

	fnObj := jitVM.bytecode.Constants[constIdx]
	fn, ok := fnObj.(*compiler.CompiledFunction)
	if !ok {
		return 0
	}

	hook := jitVM.RegVM.GetNativeCallHook()
	if hook != nil {
		result, handled := hook(fn, args, nil)
		if handled {
			if result.IsInt() {
				val, _ := result.ToInt()
				return val
			}
			if result.IsBool() {
				if result.GetBool() {
					return 1
				}
				return 0
			}
			if result.IsNull() {
				return 0
			}
			return 0
		}
	}

	// Fallback: no native hook handled this function.
	// Execute the function using the VM's interpreter (runFunctionSync).
	// This handles cases where the function can't be compiled natively
	// (e.g., higher-order functions, closures, builtins).
	callerFrame := jitVM.RegVM.CurrentFrame()
	if callerFrame == nil {
		return 0
	}

	// Copy args into callerFrame registers so runFunctionSync can pick them up
	for i := 0; i < numArgs && i < compiler.NumArgRegisters; i++ {
		callerFrame.Registers[i] = args[i]
	}

	err := jitVM.RegVM.RunFunctionSync(fn, numArgs, callerFrame)
	if err != nil {
		return 0
	}

	result := callerFrame.Registers[compiler.ReturnRegister]
	if result.IsInt() {
		val, _ := result.ToInt()
		return val
	}
	if result.IsBool() {
		if result.GetBool() {
			return 1
		}
		return 0
	}
	return 0
}

// initWindowsCallback creates the Windows x64 ABI callback function pointer.
func (j *JITVM) initWindowsCallback() error {
	globalCallCtx.mu.Lock()
	defer globalCallCtx.mu.Unlock()

	if globalCallCtx.callback != 0 {
		globalCallCtx.vm = j
		return nil
	}

	cb := syscall.NewCallback(nativeCallFromNative)
	if cb == 0 {
		return fmt.Errorf("failed to create Windows callback")
	}

	globalCallCtx.callback = cb
	globalCallCtx.vm = j
	return nil
}

// getWindowsCallbackPtr returns the function pointer for native→Go callbacks.
func getWindowsCallbackPtr() uintptr {
	globalCallCtx.mu.Lock()
	defer globalCallCtx.mu.Unlock()
	return globalCallCtx.callback
}

// getNativeCallHook returns the native call hook from the JITVM's embedded RegVM.
func (j *JITVM) getNativeCallHook() func(fn *compiler.CompiledFunction, args []vm.Value, frame *vm.RegFrame) (vm.Value, bool) {
	return j.RegVM.GetNativeCallHook()
}

// SetInSyscallN marks that we are entering/exiting a syscall.SyscallN call.
// This is used to guard against nested syscall.SyscallN deadlocks.
func setInSyscallN(v bool) {
	if v {
		atomic.StoreInt64(&inSyscallN, 1)
	} else {
		atomic.StoreInt64(&inSyscallN, 0)
	}
}

// IsInSyscallN returns true if we are currently inside a syscall.SyscallN call.
func isInSyscallN() bool {
	return atomic.LoadInt64(&inSyscallN) != 0
}
