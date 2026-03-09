// pkg/objects/class.go
package objects

import (
	"unsafe"
)

// Class represents a class definition
type Class struct {
	Name       string
	SuperClass *Class
	Methods    map[string]Object // methods are CompiledFunction objects
	InitMethod Object            // constructor method
	Fields     map[string]Object // default field values
}

func (c *Class) Type() ObjectType { return ClassType }
func (c *Class) Inspect() string  { return "class " + c.Name }
func (c *Class) ToBool() *Bool    { return TRUE }
func (c *Class) HashKey() HashKey {
	return HashKey{Type: ClassType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}

// Instance represents an instance of a class
type Instance struct {
	Class  *Class
	Fields map[string]Object
}

func (i *Instance) Type() ObjectType { return InstanceType }
func (i *Instance) Inspect() string  { return i.Class.Name + " instance" }
func (i *Instance) ToBool() *Bool    { return TRUE }
func (i *Instance) HashKey() HashKey {
	return HashKey{Type: InstanceType, Value: uint64(uintptr(unsafe.Pointer(i)))}
}
