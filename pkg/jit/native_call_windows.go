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

	// Collect int64 arguments and convert to vm.Value
	args := make([]vm.Value, numArgs)
	if argsPtr != nil && numArgs > 0 {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = vm.NewInt(argsSlice[i])
		}
	}

	// Look up the CompiledFunction from the JITVM's bytecode constant pool
	if jitVM.bytecode == nil || constIdx < 0 || constIdx >= len(jitVM.bytecode.Constants) {
		return 0
	}

	fnObj := jitVM.bytecode.Constants[constIdx]
	fn, ok := fnObj.(*compiler.CompiledFunction)
	if !ok {
		return 0
	}

	// Use the nativeCallHook from RegVM to execute through the hybrid path
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

	// Fallback: interpreter couldn't handle it
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

// compileCallWindows generates code for OpRegCall on Windows using callback.
// It uses syscall.NewCallback to create a Windows x64 ABI function pointer
// that native code can call directly.
func (cg *NativeCodeGenerator) compileCallWindows(constIdx, numArgs int) {
	// Get the callback pointer
	callbackPtr := getWindowsCallbackPtr()
	if callbackPtr == 0 {
		// No callback available, emit return-0 stub
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		cg.storeRaxToReg(255)
		return
	}

	// Save globals pointer (rdi) on stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Spill R0-R7 args to the stack area below locals
	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	// Spill R0 ([rbp-8]) to args array
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp-8] (R0)
	cg.emitUint32(r0StackDisp)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp-baseOffset], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp-baseOffset-8], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp-baseOffset-16], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp-baseOffset-24], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp-baseOffset-32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp-baseOffset-40], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp-baseOffset-48], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp-baseOffset-56], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Set up Windows x64 ABI callback args:
	// rcx = constIdx (function index in constant pool)
	// rdx = numArgs
	// r8 = argsPtr (lea r8, [rbp-baseOffset])

	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, imm32
	cg.emitUint32(uint32(constIdx))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, imm32
	cg.emitUint32(uint32(numArgs))

	// lea r8, [rbp-baseOffset]
	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback function pointer
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Result is in rax, store to R255
	cg.storeRaxToReg(255)

	// Restore globals pointer
	cg.emitBytes([]byte{0x5F}) // pop rdi
}
