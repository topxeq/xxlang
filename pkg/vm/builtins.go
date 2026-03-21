// pkg/vm/builtins.go
// Builtin function support for the VM
package vm

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// getBuiltin returns a builtin function by index
func getBuiltin(index int) *objects.Builtin {
	builtins := []*objects.Builtin{
		objects.Builtins["len"],         // 0
		objects.Builtins["pr"],          // 1
		objects.Builtins["pln"],         // 2
		objects.Builtins["typeOf"],      // 3
		objects.Builtins["substr"],      // 4
		objects.Builtins["split"],       // 5
		objects.Builtins["join"],        // 6
		objects.Builtins["trim"],        // 7
		objects.Builtins["upper"],       // 8
		objects.Builtins["lower"],       // 9
		objects.Builtins["containsStr"], // 10
		objects.Builtins["replace"],     // 11
		objects.Builtins["startsWith"],  // 12
		objects.Builtins["endsWith"],    // 13
		objects.Builtins["abs"],         // 14
		objects.Builtins["floor"],       // 15
		objects.Builtins["ceil"],        // 16
		objects.Builtins["sqrt"],        // 17
		objects.Builtins["pow"],         // 18
		objects.Builtins["min"],         // 19
		objects.Builtins["max"],         // 20
		objects.Builtins["int"],         // 21
		objects.Builtins["float"],       // 22
		objects.Builtins["string"],      // 23
		objects.Builtins["push"],        // 24
		objects.Builtins["pop"],         // 25
		objects.Builtins["first"],       // 26
		objects.Builtins["last"],        // 27
		objects.Builtins["rest"],        // 28
		objects.Builtins["keys"],        // 29
		objects.Builtins["values"],      // 30
		objects.Builtins["hasKey"],      // 31
		objects.Builtins["delete"],      // 32
		objects.Builtins["size"],        // 33
		objects.Builtins["range"],       // 34
		objects.Builtins["runCode"],     // 35
		objects.Builtins["loadPlugin"],  // 36
		objects.Builtins["error"],       // 37
		objects.Builtins["assert"],      // 38
		objects.Builtins["exit"],        // 39
		objects.Builtins["print"],       // 40 - alias for pln
		objects.Builtins["println"],     // 41 - alias for pln
		objects.Builtins["sprintf"],     // 42
		objects.Builtins["printf"],      // 43
		objects.Builtins["time"],        // 44
		objects.Builtins["parseJSON"],   // 45
		objects.Builtins["toJSON"],      // 46
		objects.Builtins["sin"],         // 47
		objects.Builtins["cos"],         // 48
		objects.Builtins["tan"],         // 49
		objects.Builtins["log"],         // 50
		objects.Builtins["exp"],         // 51
		objects.Builtins["random"],      // 52
		objects.Builtins["seed"],        // 53
	}

	if index < 0 || index >= len(builtins) {
		return nil
	}
	return builtins[index]
}

// GetBuiltinByIndex returns a builtin function by index (exported for JIT)
func GetBuiltinByIndex(index int) *objects.Builtin {
	return getBuiltin(index)
}
