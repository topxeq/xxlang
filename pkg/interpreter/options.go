// pkg/interpreter/options.go
// Functional options for configuring the interpreter.
package interpreter

import (
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/stdlib"
)

// Option configures the interpreter.
type Option func(*Interpreter)

// WithGlobals sets initial global variables.
// The globals slice is copied, so modifications to the original
// slice after calling this will not affect the interpreter.
func WithGlobals(globals []objects.Object) Option {
	return func(i *Interpreter) {
		if len(globals) > 0 {
			copy(i.globals, globals)
		}
	}
}

// WithLoader sets a custom module loader.
// This is useful for sharing modules between multiple interpreters
// or for pre-loading modules.
func WithLoader(loader *module.Loader) Option {
	return func(i *Interpreter) {
		if loader != nil {
			i.loader = loader
		}
	}
}

// WithStdlib enables standard library modules.
// This registers all stdlib modules (math, string, etc.)
// so they can be imported with `import "math"`.
func WithStdlib() Option {
	return func(i *Interpreter) {
		i.stdlib = true
		// The stdlib is already registered globally in stdlib.Registry
		// The VM's module loading will check stdlib.Has() and stdlib.Get()
	}
}

// WithGlobal sets a single global variable.
// This is a convenience option for setting a few globals without
// creating a full slice.
func WithGlobal(name string, value objects.Object) Option {
	return func(i *Interpreter) {
		symbol := i.symbolTable.Define(name)
		if symbol.Index < len(i.globals) {
			i.globals[symbol.Index] = value
		}
	}
}

// WithGlobalGo sets a single global variable from a Go value.
// The value is converted using FromGo.
func WithGlobalGo(name string, value interface{}) Option {
	return func(i *Interpreter) {
		if obj, err := FromGo(value); err == nil {
			symbol := i.symbolTable.Define(name)
			if symbol.Index < len(i.globals) {
				i.globals[symbol.Index] = obj
			}
		}
	}
}

// Ensure stdlib is imported
var _ = stdlib.Registry
