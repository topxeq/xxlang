// pkg/vm/builtins.go
// Builtin function support for the VM.
// Indices are auto-assigned from objects.BuiltinRegistry at init time.
// Lookup is O(1) array access by auto-assigned index.
package vm

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// getBuiltin returns a builtin function by auto-assigned index.
func getBuiltin(index int) *objects.Builtin {
	return objects.GetBuiltinByIndex(index)
}

// getBuiltinByName returns a builtin function by name (for tests).
func getBuiltinByName(name string) *objects.Builtin {
	return objects.GetBuiltinByName(name)
}
