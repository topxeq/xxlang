// pkg/objects/function.go
package objects

import (
	"bytes"
)

// Identifier represents an identifier (used in function parameters)
type Identifier struct {
	Value string
}

func (i *Identifier) String() string {
	return i.Value
}

// Environment represents a variable scope
type Environment struct {
	Store map[string]Object
	Outer *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{Store: make(map[string]Object), Outer: nil}
}

// NewEnclosedEnvironment creates a new environment with an outer scope
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{Store: make(map[string]Object), Outer: outer}
}

// Get retrieves a variable from the environment
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.Store[name]
	if !ok && e.Outer != nil {
		obj, ok = e.Outer.Get(name)
	}
	return obj, ok
}

// Set sets a variable in the environment
func (e *Environment) Set(name string, val Object) Object {
	e.Store[name] = val
	return val
}

// Function represents a user-defined function
type Function struct {
	Parameters []*Identifier
	Body       interface{} // Will be *ast.BlockStatement, using interface{} to avoid import cycle
	Env        *Environment
	Name       string // Optional: for named functions
}

func (f *Function) Type() ObjectType { return FunctionType }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	out.WriteString("func(")
	for i, p := range f.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") { ... }")
	return out.String()
}
func (f *Function) ToBool() *Bool    { return TRUE }
func (f *Function) HashKey() HashKey { return HashKey{Type: FunctionType, Value: 0} }
