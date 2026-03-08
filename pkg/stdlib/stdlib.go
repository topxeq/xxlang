// pkg/stdlib/stdlib.go
// Standard library modules for Xxlang.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// Module represents a standard library module.
type Module struct {
	Name    string
	Exports map[string]objects.Object
}

// Registry holds all registered standard library modules.
var Registry = make(map[string]*Module)

// Register adds a module to the registry.
func Register(m *Module) {
	Registry[m.Name] = m
}

// Get retrieves a module from the registry.
func Get(name string) *Module {
	return Registry[name]
}

// Has checks if a module exists in the registry.
func Has(name string) bool {
	_, ok := Registry[name]
	return ok
}

// BuiltinFunc creates a builtin function object.
func BuiltinFunc(fn func(...objects.Object) objects.Object) *objects.Builtin {
	return &objects.Builtin{Fn: fn}
}

// Float creates a float object.
func Float(v float64) *objects.Float {
	return &objects.Float{Value: v}
}

// String creates a string object.
func String(v string) *objects.String {
	return &objects.String{Value: v}
}

// Int creates an integer object.
func Int(v int64) *objects.Int {
	return &objects.Int{Value: v}
}

// Bool creates a boolean object.
func Bool(v bool) *objects.Bool {
	if v {
		return objects.TRUE
	}
	return objects.FALSE
}

// Array creates an array object.
func Array(elements ...objects.Object) *objects.Array {
	return &objects.Array{Elements: elements}
}

// Null returns the null object.
func Null() *objects.Null {
	return objects.NULL
}

// Error creates an error object.
func Error(msg string) *objects.Error {
	return &objects.Error{Message: msg}
}
