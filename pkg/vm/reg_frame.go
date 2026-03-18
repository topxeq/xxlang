// pkg/vm/reg_frame.go
package vm

import (
	"sync"

	"github.com/topxeq/xxlang/pkg/compiler"
)

// RegFrame represents a call frame for the register-based VM
// Each frame has 256 fixed registers (R0-R255)
type RegFrame struct {
	Fn        *compiler.CompiledFunction
	IP        int            // Instruction pointer
	Registers [compiler.NumRegisters]Value // Fixed-size register array
	FreeVars  []Value // Closure free variables
	Constants []Value // Constants for this frame
	Globals   []Value // Global variables reference
	This      Value   // 'this' for method calls
	Locals    []Value // Local variables (for spilled values)
}

// Register frame pool for reducing allocations
var regFramePool = sync.Pool{
	New: func() interface{} {
		return &RegFrame{}
	},
}

// NewRegFrame creates a new register-based call frame
func NewRegFrame(fn *compiler.CompiledFunction) *RegFrame {
	f := regFramePool.Get().(*RegFrame)

	f.Fn = fn
	f.IP = 0 // Start at instruction 0

	// Reset all registers to null
	for i := range f.Registers {
		f.Registers[i] = ValueNull
	}

	// Allocate or reuse free variables array
	numFreeVars := len(fn.FreeVariables)
	if cap(f.FreeVars) >= numFreeVars {
		f.FreeVars = f.FreeVars[:numFreeVars]
		for i := range f.FreeVars {
			f.FreeVars[i] = ValueNull
		}
	} else {
		f.FreeVars = make([]Value, numFreeVars)
	}

	// Allocate locals for spilled values
	if cap(f.Locals) >= fn.NumLocals {
		f.Locals = f.Locals[:fn.NumLocals]
		for i := range f.Locals {
			f.Locals[i] = ValueNull
		}
	} else {
		f.Locals = make([]Value, fn.NumLocals)
	}

	f.Constants = nil
	f.Globals = nil
	f.This = ValueNull

	return f
}

// Release returns the frame to the pool for reuse
func (f *RegFrame) Release() {
	// Clear references
	f.Fn = nil
	f.IP = 0

	// Clear registers
	for i := range f.Registers {
		f.Registers[i] = ValueNull
	}

	f.FreeVars = f.FreeVars[:0]
	f.Constants = nil
	f.Globals = nil
	f.This = ValueNull
	f.Locals = f.Locals[:0]

	regFramePool.Put(f)
}

// Instructions returns the compiled function's instructions
func (f *RegFrame) Instructions() []byte {
	return f.Fn.Instructions
}

// GetReg returns the value in a register
func (f *RegFrame) GetReg(reg int) Value {
	return f.Registers[reg]
}

// SetReg sets the value in a register
func (f *RegFrame) SetReg(reg int, val Value) {
	f.Registers[reg] = val
}

// GetLocal returns a local variable value
func (f *RegFrame) GetLocal(idx int) Value {
	if idx < len(f.Locals) {
		return f.Locals[idx]
	}
	return ValueNull
}

// SetLocal sets a local variable value
func (f *RegFrame) SetLocal(idx int, val Value) {
	if idx < len(f.Locals) {
		f.Locals[idx] = val
	}
}

// GetFree returns a free variable value
func (f *RegFrame) GetFree(idx int) Value {
	if idx < len(f.FreeVars) {
		return f.FreeVars[idx]
	}
	return ValueNull
}

// SetFree sets a free variable value
func (f *RegFrame) SetFree(idx int, val Value) {
	if idx < len(f.FreeVars) {
		f.FreeVars[idx] = val
	}
}

// GetConstant returns a constant value
func (f *RegFrame) GetConstant(idx int) Value {
	if idx < len(f.Constants) {
		return f.Constants[idx]
	}
	return ValueNull
}

// GetGlobal returns a global variable value
func (f *RegFrame) GetGlobal(idx int) Value {
	if idx < len(f.Globals) {
		return f.Globals[idx]
	}
	return ValueNull
}

// SetGlobal sets a global variable value
func (f *RegFrame) SetGlobal(idx int, val Value) {
	if idx < len(f.Globals) {
		f.Globals[idx] = val
	}
}

// CopyArgRegisters copies argument values from source frame's return registers
// to this frame's argument registers (R0-R7)
func (f *RegFrame) CopyArgRegisters(src *RegFrame, numArgs int) {
	for i := 0; i < numArgs && i < compiler.NumArgRegisters; i++ {
		f.Registers[i] = src.Registers[i]
	}
}

// ClearArgRegisters clears the argument registers
func (f *RegFrame) ClearArgRegisters() {
	for i := 0; i < compiler.NumArgRegisters; i++ {
		f.Registers[i] = ValueNull
	}
}
