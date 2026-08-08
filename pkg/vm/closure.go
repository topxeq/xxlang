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
	FreeVars  []objects.Object // Free variables for stack VM
	Constants []objects.Object // Constants from the creating VM
	Globals   []objects.Object // Globals from the creating module (for exported functions)

	// FreeVarsValues is used by register VM for shared mutable free variables
	// When set, register VM uses this instead of FreeVars
	FreeVarsValues []Value

	// GlobalsValues is used by register VM for module globals
	// When set, register VM uses this instead of Globals
	GlobalsValues []Value

	// userFuncCtx is bound by the VM that created this closure. It lets
	// higher-order builtins (map/filter/reduce/foreach/...) call this closure
	// back into the correct VM instance under concurrent executions.
	userFuncCtx func(args ...objects.Object) (objects.Object, error)
}

// CallUserFuncInContext implements objects.UserFuncCaller.
// It invokes this closure inside the VM context that created it.
func (c *Closure) CallUserFuncInContext(args ...objects.Object) (objects.Object, error) {
	if c.userFuncCtx == nil {
		return nil, fmt.Errorf("closure has no VM context bound")
	}
	return c.userFuncCtx(args...)
}

// bindVMContext attaches this closure's user-function callback to the given
// VM. Must be called by the VM that creates (or re-wraps) the closure so that
// higher-order builtins (map/filter/reduce/...) dispatch back to the correct
// VM instance — required for safe concurrent execution of multiple VMs.
func (c *Closure) bindVMContext(vm *RegVM) *Closure {
	c.userFuncCtx = func(args ...objects.Object) (objects.Object, error) {
		return CallUserFuncInRegVM(c, args, vm)
	}
	return c
}

// Type returns the object type
func (c *Closure) Type() objects.ObjectType { return objects.ClosureType }

// TypeTag returns the type tag for fast type checking
func (c *Closure) TypeTag() objects.TypeTag { return objects.TagClosure }

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
