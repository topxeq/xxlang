// pkg/objects/builtin_test.go
package objects

import "testing"

func TestBuiltinType(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if got := b.Type(); got != BuiltinType {
		t.Errorf("Builtin.Type() = %s, want BUILTIN", got)
	}
}

func TestBuiltinInspect(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if got := b.Inspect(); got != "builtin function" {
		t.Errorf("Builtin.Inspect() = %s, want 'builtin function'", got)
	}
}
