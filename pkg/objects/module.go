// pkg/objects/module.go
// Module object type for runtime module representation.
package objects

import "fmt"

// Module represents a loaded module with its exported symbols.
// Modules are created when a source file is imported and compiled,
// and they hold all exported values accessible to importers.
type Module struct {
	// Name is the module's identifier (typically the file path)
	Name string

	// Exports maps exported symbol names to their values
	Exports map[string]Object

	// Globals holds the module's global variables state.
	// This is needed so exported functions can access module-level variables.
	Globals []Object
}

// Type returns the object type.
func (m *Module) Type() ObjectType { return ModuleType }

// Inspect returns a string representation of the module.
func (m *Module) Inspect() string {
	return fmt.Sprintf("[module %s]", m.Name)
}

// ToBool converts the module to a boolean (always true).
func (m *Module) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the module.
// Modules are not hashable in a meaningful way, so we return a constant.
func (m *Module) HashKey() HashKey {
	return HashKey{Type: ModuleType, Value: 0}
}
