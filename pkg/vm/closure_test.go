// pkg/vm/closure_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

func TestClosureType(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 0}
	c := &Closure{Fn: fn}

	if c.Type() != objects.ClosureType {
		t.Errorf("Closure.Type() = %s, want ClosureType", c.Type())
	}
}

func TestClosureTypeTag(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 0}
	c := &Closure{Fn: fn}

	if c.TypeTag() != objects.TagClosure {
		t.Errorf("Closure.TypeTag() = %d, want TagClosure", c.TypeTag())
	}
}

func TestClosureInspect(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 0}

	// Closure with no free variables
	c := &Closure{Fn: fn, FreeVars: []objects.Object{}}
	if c.Inspect() != "closure[0 freeVars]" {
		t.Errorf("Closure.Inspect() = %s, want 'closure[0 freeVars]'", c.Inspect())
	}

	// Closure with free variables
	freeVars := []objects.Object{&objects.Int{Value: 42}, &objects.String{Value: "hello"}}
	c = &Closure{Fn: fn, FreeVars: freeVars}
	if c.Inspect() != "closure[2 freeVars]" {
		t.Errorf("Closure.Inspect() = %s, want 'closure[2 freeVars]'", c.Inspect())
	}
}

func TestClosureToBool(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 0}
	c := &Closure{Fn: fn}

	if c.ToBool() != objects.TRUE {
		t.Error("Closure.ToBool() should return TRUE")
	}
}

func TestClosureHashKey(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 0}
	c := &Closure{Fn: fn}

	hk := c.HashKey()
	if hk.Type != objects.ClosureType {
		t.Errorf("Closure.HashKey().Type = %s, want ClosureType", hk.Type)
	}
	if hk.Value != 0 {
		t.Errorf("Closure.HashKey().Value = %d, want 0", hk.Value)
	}
}

func TestClosuresInVM(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Basic closure
		{"func makeCounter() { var count = 0; func counter() { count = count + 1; return count; }; return counter; }; var c = makeCounter(); c();", 1},
		{"func makeCounter() { var count = 0; func counter() { count = count + 1; return count; }; return counter; }; var c = makeCounter(); c(); c();", 2},
		// Closure capturing multiple variables
		{"func makeAdder(x) { func adder(y) { return x + y; }; return adder; }; var add5 = makeAdder(5); add5(3);", 8},
		// Nested closures
		{"func outer(x) { func middle(y) { func inner(z) { return x + y + z; }; return inner; }; return middle; }; outer(1)(2)(3);", 6},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}
