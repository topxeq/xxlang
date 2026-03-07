// pkg/vm/frame.go
package vm

import (
	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Frame represents a call frame for function execution
type Frame struct {
	Fn          *compiler.CompiledFunction
	IP          int               // Instruction pointer (index into Instructions)
	BasePointer int               // Stack base pointer for this frame
	Locals      []objects.Object  // Local variables
	FreeVars    []objects.Object  // Free variables (captured from closure)
}

// NewFrame creates a new call frame
func NewFrame(fn *compiler.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		Fn:          fn,
		IP:          -1, // Start at -1 so first increment goes to 0
		BasePointer: basePointer,
		Locals:      make([]objects.Object, fn.NumLocals),
		FreeVars:    make([]objects.Object, len(fn.FreeVariables)),
	}
}

// Instructions returns the compiled function's instructions
func (f *Frame) Instructions() []byte {
	return f.Fn.Instructions
}
