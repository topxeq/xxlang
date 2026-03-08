// pkg/module/module.go
// Package module provides module loading and management for Xxlang.
package module

import "github.com/topxeq/xxlang/pkg/objects"

// Module represents a compiled module with exported symbols.
// A module is created when a source file is loaded and compiled,
// and its exported symbols are available for import by other modules.
type Module struct {
	// Name is the module's identifier (e.g., "./math", "stdlib/json")
	Name string

	// Exports maps exported symbol names to their values
	Exports map[string]objects.Object

	// Globals stores the module's global variables for exported functions to access
	Globals []objects.Object
}

// NewModule creates a new module with the given name.
// The exports map is initialized to an empty map.
func NewModule(name string) *Module {
	return &Module{
		Name:    name,
		Exports: make(map[string]objects.Object),
	}
}

// Export adds an export to the module.
// If an export with the same name already exists, it is overwritten.
func (m *Module) Export(name string, value objects.Object) {
	m.Exports[name] = value
}

// HasExport checks if an export with the given name exists.
func (m *Module) HasExport(name string) bool {
	_, ok := m.Exports[name]
	return ok
}

// GetExport retrieves an export by name.
// Returns the exported object and true if found, or nil and false otherwise.
func (m *Module) GetExport(name string) (objects.Object, bool) {
	val, ok := m.Exports[name]
	return val, ok
}
