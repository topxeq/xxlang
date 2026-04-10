// pkg/vm/builtins.go
// Builtin function support for the VM.
// The builtin array is auto-built from objects.BuiltinRegistry at init time,
// so this file never needs manual index management.
package vm

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// builtinArray is the index-to-function lookup table, built once at init.
var builtinArray []*objects.Builtin

func init() {
	registry := objects.BuiltinRegistry
	builtinArray = make([]*objects.Builtin, len(registry))
	for i, name := range registry {
		if name != "" {
			builtinArray[i] = objects.Builtins[name]
		}
	}
}

// getBuiltin returns a builtin function by index.
func getBuiltin(index int) *objects.Builtin {
	if index < 0 || index >= len(builtinArray) {
		return nil
	}
	return builtinArray[index]
}

// GetBuiltinByIndex returns a builtin function by index (exported for JIT).
func GetBuiltinByIndex(index int) *objects.Builtin {
	return getBuiltin(index)
}
