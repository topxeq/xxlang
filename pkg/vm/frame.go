// pkg/vm/frame.go
package vm

import "github.com/topxeq/xxlang/pkg/objects"

// Frame represents a call frame for function execution
type Frame struct {
	Fn          *objects.CompiledFunction
	IP          int // Instruction pointer (index into Instructions)
	BasePointer int // Stack base pointer for this frame
	Locals      []objects.Object // Local variables
}

// NewFrame creates a new call frame
func NewFrame(fn *objects.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		Fn:          fn,
		IP:          -1, // Start at -1 so first increment goes to 0
		BasePointer: basePointer,
		Locals:      make([]objects.Object, fn.NumLocals),
	}
}

// Instructions returns the compiled function's instructions
func (f *Frame) Instructions() []byte {
	return f.Fn.Instructions
}
