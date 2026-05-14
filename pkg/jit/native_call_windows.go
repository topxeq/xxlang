//go:build windows && amd64
// +build windows,amd64

package jit

import (
	"fmt"
	"sync"
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

// nativeCallFromNative is the Go callback invoked from native code for OpRegCall.
// Parameters follow Windows x64 ABI: rcx=constIdx, rdx=numArgs, r8=argsPtr.
// Returns the int64 result of the function call.
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
			return 0
		}
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

	// Create a Windows-callable function pointer from the Go function
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
