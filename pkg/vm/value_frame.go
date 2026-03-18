// pkg/vm/value_frame.go
// Value-based frame for zero-allocation local variable access
package vm

import (
	"sync"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ValueFrame represents a call frame optimized for Value-based execution
// Uses Value type for locals to avoid heap allocation for primitives
type ValueFrame struct {
	Fn          *compiler.CompiledFunction
	IP          int     // Instruction pointer
	BasePointer int     // Stack base pointer
	Locals      []Value // Value-based local variables
	FreeVars    []Value // Value-based free variables
	Constants   []Value // Value-based constants
	Globals     []objects.Object
	This        objects.Object
}

// ValueFrame pool for reducing allocations
var valueFramePool = sync.Pool{
	New: func() interface{} {
		return &ValueFrame{}
	},
}

// NewValueFrame creates a new Value-based call frame
func NewValueFrame(fn *compiler.CompiledFunction, basePointer int) *ValueFrame {
	f := valueFramePool.Get().(*ValueFrame)

	f.Fn = fn
	f.IP = -1
	f.BasePointer = basePointer

	// Allocate or reuse locals array
	if cap(f.Locals) >= fn.NumLocals {
		f.Locals = f.Locals[:fn.NumLocals]
	} else {
		f.Locals = make([]Value, fn.NumLocals)
	}
	// Initialize all locals to null
	for i := range f.Locals {
		f.Locals[i] = ValueNull
	}

	// Allocate free variables
	numFreeVars := len(fn.FreeVariables)
	if cap(f.FreeVars) >= numFreeVars {
		f.FreeVars = f.FreeVars[:numFreeVars]
	} else {
		f.FreeVars = make([]Value, numFreeVars)
	}

	f.Constants = nil
	f.Globals = nil
	f.This = nil

	return f
}

// Release returns the frame to the pool
func (f *ValueFrame) Release() {
	f.Fn = nil
	f.IP = -1
	f.BasePointer = 0
	f.Locals = f.Locals[:0]
	f.FreeVars = f.FreeVars[:0]
	f.Constants = nil
	f.Globals = nil
	f.This = nil
	valueFramePool.Put(f)
}

// Instructions returns the compiled function's instructions
func (f *ValueFrame) Instructions() []byte {
	return f.Fn.Instructions
}

// GetLocal gets a local variable as Value
func (f *ValueFrame) GetLocal(idx int) Value {
	if idx < 0 || idx >= len(f.Locals) {
		return ValueNull
	}
	return f.Locals[idx]
}

// SetLocal sets a local variable
func (f *ValueFrame) SetLocal(idx int, v Value) {
	if idx >= 0 && idx < len(f.Locals) {
		f.Locals[idx] = v
	}
}

// GetFreeVar gets a free variable as Value
func (f *ValueFrame) GetFreeVar(idx int) Value {
	if idx < 0 || idx >= len(f.FreeVars) {
		return ValueNull
	}
	return f.FreeVars[idx]
}

// SetFreeVar sets a free variable
func (f *ValueFrame) SetFreeVar(idx int, v Value) {
	if idx >= 0 && idx < len(f.FreeVars) {
		f.FreeVars[idx] = v
	}
}

// FrameFromValueFrame converts a ValueFrame to a regular Frame
// Used when falling back to Object-based execution
func FrameFromValueFrame(vf *ValueFrame) *Frame {
	f := framePool.Get().(*Frame)
	f.Fn = vf.Fn
	f.IP = vf.IP
	f.BasePointer = vf.BasePointer

	// Convert Value locals to Object locals
	if cap(f.Locals) >= len(vf.Locals) {
		f.Locals = f.Locals[:len(vf.Locals)]
	} else {
		f.Locals = make([]objects.Object, len(vf.Locals))
	}
	for i, v := range vf.Locals {
		f.Locals[i] = v.ToObject()
	}

	// Convert free variables
	if cap(f.FreeVars) >= len(vf.FreeVars) {
		f.FreeVars = f.FreeVars[:len(vf.FreeVars)]
	} else {
		f.FreeVars = make([]objects.Object, len(vf.FreeVars))
	}
	for i, v := range vf.FreeVars {
		f.FreeVars[i] = v.ToObject()
	}

	f.This = vf.This
	f.Globals = vf.Globals

	return f
}

// ValueFrameFromFrame converts a regular Frame to a ValueFrame
func ValueFrameFromFrame(of *Frame) *ValueFrame {
	vf := valueFramePool.Get().(*ValueFrame)
	vf.Fn = of.Fn
	vf.IP = of.IP
	vf.BasePointer = of.BasePointer

	// Convert Object locals to Value locals
	if cap(vf.Locals) >= len(of.Locals) {
		vf.Locals = vf.Locals[:len(of.Locals)]
	} else {
		vf.Locals = make([]Value, len(of.Locals))
	}
	for i, obj := range of.Locals {
		vf.Locals[i] = NewValue(obj)
	}

	// Convert free variables
	if cap(vf.FreeVars) >= len(of.FreeVars) {
		vf.FreeVars = vf.FreeVars[:len(of.FreeVars)]
	} else {
		vf.FreeVars = make([]Value, len(of.FreeVars))
	}
	for i, obj := range of.FreeVars {
		vf.FreeVars[i] = NewValue(obj)
	}

	vf.This = of.This
	vf.Globals = of.Globals

	return vf
}
