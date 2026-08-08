// pkg/vm/builtins.go
// Builtin function support for the VM.
// Indices are auto-assigned from objects.BuiltinRegistry at init time.
// Lookup is O(1) array access by auto-assigned index.
package vm

import (
	"fmt"

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

// buildVMBuiltinFuncs creates this VM's private copy of the builtin function
// table. The global table is immutable/shared; the VM needs its own copy so
// that context-bound builtins (runCode / loadPlugin / delegate) can capture
// THIS VM instance. With a shared global table, concurrent VM executions
// would overwrite each other's callbacks (data race + wrong-context calls).
//
// Must be called during VM construction, after the VM struct is allocated
// (the closures capture vm).
func buildVMBuiltinFuncs(vm *RegVM) []*objects.Builtin {
	// Trigger lazy initialization of the global index tables.
	objects.GetBuiltinByIndex(0)
	funcs := make([]*objects.Builtin, len(objects.BuiltinFuncArray))
	copy(funcs, objects.BuiltinFuncArray)

	// Replace context-bound builtins with VM-bound closures.
	for name, idx := range objects.BuiltinIndexMap {
		if idx < 0 || idx >= len(funcs) || funcs[idx] == nil {
			continue
		}
		switch name {
		case "runCode":
			b := *funcs[idx] // shallow copy; do not mutate the global entry
			b.Fn = func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return errorObj("wrong number of arguments for runCode. got=%d, want=1 or 2", len(args))
				}
				code, ok := args[0].(*objects.String)
				if !ok {
					return errorObj("first argument to 'runCode' must be STRING, got %s", args[0].Type())
				}
				var argMap *objects.Map
				if len(args) == 2 {
					argMap, ok = args[1].(*objects.Map)
					if !ok {
						return errorObj("second argument to 'runCode' must be MAP, got %s", args[1].Type())
					}
				}
				result, err := RunCodeInRegVM(code.Value, argMap, vm)
				if err != nil {
					return errorObj("runCode error: %v", err)
				}
				if result == nil {
					return objects.NULL
				}
				return result
			}
			funcs[idx] = &b

		case "delegate":
			b := *funcs[idx]
			b.Fn = func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return errorObj("wrong number of arguments for delegate. got=%d, want=1", len(args))
				}
				source, ok := args[0].(*objects.String)
				if !ok {
					return errorObj("argument to 'delegate' must be STRING, got %s", args[0].Type())
				}
				result, err := CreateDelegateInRegVM(source.Value, vm)
				if err != nil {
					return errorObj("delegate error: %v", err)
				}
				if result == nil {
					return objects.NULL
				}
				return result
			}
			funcs[idx] = &b

		case "loadPlugin":
			b := *funcs[idx]
			b.Fn = func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return errorObj("wrong number of arguments for loadPlugin. got=%d, want=1", len(args))
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return errorObj("argument to 'loadPlugin' must be STRING (file path), got %s", args[0].Type())
				}
				result, err := vm.loadPluginByPath(path.Value)
				if err != nil {
					return errorObj("loadPlugin error: %v", err)
				}
				if result == nil {
					return objects.NULL
				}
				return result
			}
			funcs[idx] = &b
		}
	}

	return funcs
}

// vmBuiltin returns this VM's bound builtin by index.
func (vm *RegVM) vmBuiltin(index int) (*objects.Builtin, error) {
	if vm.builtinFuncs == nil {
		return nil, fmt.Errorf("vm builtin table not initialized")
	}
	if index < 0 || index >= len(vm.builtinFuncs) {
		return nil, fmt.Errorf("invalid builtin index: %d", index)
	}
	return vm.builtinFuncs[index], nil
}

// errorObj creates an Error object (objects.newError is unexported).
func errorObj(format string, a ...interface{}) *objects.Error {
	return &objects.Error{Message: fmt.Sprintf(format, a...)}
}
