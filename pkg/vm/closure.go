// pkg/vm/closure.go
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Closure represents a function with captured variables
type Closure struct {
	Fn        *compiler.CompiledFunction
	FreeVars  []objects.Object
	Constants []objects.Object // Constants from the creating VM
	Globals   []objects.Object // Globals from the creating module (for exported functions)
}

// Type returns the object type
func (c *Closure) Type() objects.ObjectType { return objects.ClosureType }

// Inspect returns a string representation
func (c *Closure) Inspect() string {
	return fmt.Sprintf("closure[%d freeVars]", len(c.FreeVars))
}

// ToBool returns the boolean value
func (c *Closure) ToBool() *objects.Bool {
	return objects.TRUE
}

// HashKey returns the hash key
func (c *Closure) HashKey() objects.HashKey {
	return objects.HashKey{Type: objects.ClosureType, Value: 0}
}

// executeClosure creates a new closure with captured free variables
func (vm *VM) executeClosure() error {
	// Read the function index (2 bytes)
	fnIndex := int(vm.readUint16())
	vm.currentFrame().IP += 2

	// Read number of free variables (1 byte)
	numFree := int(vm.readUint8())
	vm.currentFrame().IP++

	// Get the compiled function from constants
	fn, ok := vm.constants[fnIndex].(*compiler.CompiledFunction)
	if !ok {
		return fmt.Errorf("expected CompiledFunction at index %d, got %T", fnIndex, vm.constants[fnIndex])
	}

	// Get globals from module context if available, otherwise use VM globals
	globals := vm.globals
	if vm.currentModule != nil && vm.currentModule.Globals != nil {
		globals = vm.currentModule.Globals
	}

	// Create closure with captured free variables, constants, and globals
	closure := &Closure{
		Fn:        fn,
		FreeVars:  make([]objects.Object, numFree),
		Constants: vm.constants, // Store the constants from this VM
		Globals:   globals,       // Store globals for module-level variable access
	}

	// Pop the free variables from stack and store them in the closure
	for i := numFree - 1; i >= 0; i-- {
		closure.FreeVars[i] = vm.stack.Pop()
	}

	// Push closure onto stack
	vm.stack.Push(closure)

	return nil
}

// executeGetFree gets a free variable from the current closure
func (vm *VM) executeGetFree() error {
	freeIndex := int(vm.readUint8())
	vm.currentFrame().IP++

	frame := vm.currentFrame()

	// Get the free variable from the closure's free vars
	if freeIndex >= len(frame.FreeVars) {
		return fmt.Errorf("free variable index %d out of bounds (len=%d)", freeIndex, len(frame.FreeVars))
	}

	value := frame.FreeVars[freeIndex]
	vm.stack.Push(value)

	return nil
}

// executeSetFree sets a free variable in the current closure
func (vm *VM) executeSetFree() error {
	freeIndex := int(vm.readUint8())
	vm.currentFrame().IP++

	value := vm.stack.Pop()

	frame := vm.currentFrame()

	// Set the free variable in the closure's free vars
	if freeIndex >= len(frame.FreeVars) {
		return fmt.Errorf("free variable index %d out of bounds (len=%d)", freeIndex, len(frame.FreeVars))
	}

	frame.FreeVars[freeIndex] = value
	vm.stack.Push(value) // Push the value back for chaining

	return nil
}
