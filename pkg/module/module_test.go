// pkg/module/module_test.go
package module

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestModuleExports(t *testing.T) {
	m := NewModule("./math")
	m.Export("add", &objects.Int{Value: 1})
	m.Export("sub", &objects.Int{Value: 2})

	if !m.HasExport("add") {
		t.Error("expected HasExport('add') to be true")
	}
	if !m.HasExport("sub") {
		t.Error("expected HasExport('sub') to be true")
	}
	if m.HasExport("mul") {
		t.Error("expected HasExport('mul') to be false")
	}
}

func TestModuleGetExport(t *testing.T) {
	m := NewModule("./math")
	m.Export("add", &objects.Int{Value: 1})

	val, ok := m.GetExport("add")
	if !ok {
		t.Error("expected GetExport('add') to return ok=true")
	}
	if val.Type() != objects.IntType {
		t.Errorf("expected Int type, got %s", val.Type())
	}

	_, ok = m.GetExport("nonexistent")
	if ok {
		t.Error("expected GetExport('nonexistent') to return ok=false")
	}
}

func TestModuleName(t *testing.T) {
	m := NewModule("./math")
	if m.Name != "./math" {
		t.Errorf("expected Name to be './math', got '%s'", m.Name)
	}
}
