// pkg/objects/module_test.go
package objects

import "testing"

func TestModuleType(t *testing.T) {
	m := &Module{Name: "test", Exports: map[string]Object{}}
	if got := m.Type(); got != ModuleType {
		t.Errorf("Module.Type() = %s, want MODULE", got)
	}
}

func TestModuleInspect(t *testing.T) {
	m := &Module{Name: "test", Exports: map[string]Object{}}
	if got := m.Inspect(); got != "[module test]" {
		t.Errorf("Module.Inspect() = %s, want '[module test]'", got)
	}
}

func TestModuleToBool(t *testing.T) {
	m := &Module{Name: "test", Exports: map[string]Object{}}
	if m.ToBool() != TRUE {
		t.Error("Module.ToBool() should be TRUE")
	}
}

func TestModuleHashKey(t *testing.T) {
	m := &Module{Name: "test", Exports: map[string]Object{}}
	if m.HashKey() != (HashKey{Type: ModuleType, Value: 0}) {
		t.Error("Module.HashKey() should return constant hash")
	}
}

func TestModuleWithExports(t *testing.T) {
	exports := map[string]Object{
		"foo": &Int{Value: 42},
		"bar": &String{Value: "hello"},
	}
	m := &Module{
		Name:    "mymodule",
		Exports: exports,
		Globals: []Object{&Int{Value: 1}},
	}

	if m.Type() != ModuleType {
		t.Errorf("expected ModuleType, got %s", m.Type())
	}

	if m.Inspect() != "[module mymodule]" {
		t.Errorf("expected '[module mymodule]', got %s", m.Inspect())
	}

	if len(m.Exports) != 2 {
		t.Errorf("expected 2 exports, got %d", len(m.Exports))
	}

	if len(m.Globals) != 1 {
		t.Errorf("expected 1 global, got %d", len(m.Globals))
	}
}
