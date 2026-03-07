// pkg/objects/array_test.go
package objects

import "testing"

func TestArrayInspect(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Int{Value: 1},
		&Int{Value: 2},
		&Int{Value: 3},
	}}
	if got := arr.Inspect(); got != "[1, 2, 3]" {
		t.Errorf("Array.Inspect() = %s, want [1, 2, 3]", got)
	}
}

func TestArrayType(t *testing.T) {
	arr := &Array{Elements: []Object{}}
	if got := arr.Type(); got != ArrayType {
		t.Errorf("Array.Type() = %s, want ARRAY", got)
	}
}

func TestArrayToBool(t *testing.T) {
	empty := &Array{Elements: []Object{}}
	if empty.ToBool() != FALSE {
		t.Error("Array([]).ToBool() should be FALSE")
	}

	nonempty := &Array{Elements: []Object{&Int{Value: 1}}}
	if nonempty.ToBool() != TRUE {
		t.Error("Array([1]).ToBool() should be TRUE")
	}
}
