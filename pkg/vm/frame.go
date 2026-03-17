// pkg/vm/frame.go
package vm

import (
	"sync"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Frame represents a call frame for function execution
type Frame struct {
	Fn          *compiler.CompiledFunction
	IP          int              // Instruction pointer (index into Instructions)
	BasePointer int              // Stack base pointer for this frame
	Locals      []objects.Object // Local variables
	FreeVars    []objects.Object // Free variables (captured from closure)
	Constants   []objects.Object // Constants for this frame (from closure or VM)
	Globals     []objects.Object // Globals for this frame (from closure's module or VM)
	This        objects.Object   // 'this' for method calls
}

// Frame pool for reducing allocations
var framePool = sync.Pool{
	New: func() interface{} {
		return &Frame{}
	},
}

// NewFrame creates a new call frame
func NewFrame(fn *compiler.CompiledFunction, basePointer int) *Frame {
	// Get frame from pool
	f := framePool.Get().(*Frame)

	// Initialize frame
	f.Fn = fn
	f.IP = -1 // Start at -1 so first increment goes to 0
	f.BasePointer = basePointer

	// Allocate or reuse locals array
	if cap(f.Locals) >= fn.NumLocals {
		f.Locals = f.Locals[:fn.NumLocals]
	} else {
		f.Locals = make([]objects.Object, fn.NumLocals)
	}

	// Allocate or reuse free variables array
	numFreeVars := len(fn.FreeVariables)
	if cap(f.FreeVars) >= numFreeVars {
		f.FreeVars = f.FreeVars[:numFreeVars]
	} else {
		f.FreeVars = make([]objects.Object, numFreeVars)
	}

	f.Constants = nil
	f.Globals = nil
	f.This = nil

	return f
}

// Release returns the frame to the pool for reuse
func (f *Frame) Release() {
	// Clear references to allow GC
	f.Fn = nil
	f.IP = -1
	f.BasePointer = 0
	f.Locals = f.Locals[:0]
	f.FreeVars = f.FreeVars[:0]
	f.Constants = nil
	f.Globals = nil
	f.This = nil

	// Return to pool
	framePool.Put(f)
}

// Instructions returns the compiled function's instructions
func (f *Frame) Instructions() []byte {
	return f.Fn.Instructions
}
