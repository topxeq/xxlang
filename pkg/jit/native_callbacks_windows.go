//go:build windows && amd64
// +build windows,amd64

package jit

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/objects"
)

// builtinCallbackCtx holds context for the builtin callback.
type builtinCallbackCtx struct {
	mu       sync.Mutex
	vm       *JITVM
	callback uintptr
}

var globalBuiltinCtx builtinCallbackCtx

// collectionCallbackCtx holds context for the collection callback.
type collectionCallbackCtx struct {
	mu       sync.Mutex
	vm       *JITVM
	callback uintptr
}

var globalCollectionCtx collectionCallbackCtx

// objectCallbackCtx holds context for the object callback.
type objectCallbackCtx struct {
	mu       sync.Mutex
	vm       *JITVM
	callback uintptr
}

var globalObjectCtx objectCallbackCtx

// builtinCallbackFromNative is the Go callback invoked from native code for OpRegBuiltin.
// Parameters follow Windows x64 ABI: rcx=builtinIdx, rdx=numArgs, r8=argsPtr.
// Returns the int64 result of the builtin function call.
func builtinCallbackFromNative(builtinIdx int, numArgs int, argsPtr *int64) int64 {
	globalBuiltinCtx.mu.Lock()
	jitVM := globalBuiltinCtx.vm
	globalBuiltinCtx.mu.Unlock()

	if jitVM == nil {
		return 0
	}

	// Get the builtin function by auto-assigned index
	builtin := objects.GetBuiltinByIndex(builtinIdx)
	if builtin == nil {
		return 0
	}

	// Convert int64 args to objects.Object args
	args := make([]objects.Object, numArgs)
	if argsPtr != nil && numArgs > 0 {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = &objects.Int{Value: argsSlice[i]}
		}
	}

	// Call the builtin function
	result := builtin.Fn(args...)

	// Convert result back to int64
	return objectToNativeInt(result)
}

// collectionCallbackFromNative is the Go callback invoked from native code for collection ops.
// Parameters follow Windows x64 ABI: rcx=opKind, rdx=numArgs, r8=argsPtr.
// Returns the int64 result (handle or element value).
func collectionCallbackFromNative(opKind int, numArgs int, argsPtr *int64) int64 {
	globalCollectionCtx.mu.Lock()
	jitVM := globalCollectionCtx.vm
	globalCollectionCtx.mu.Unlock()

	if jitVM == nil {
		return 0
	}

	// Convert int64 args to objects.Object args
	args := make([]objects.Object, numArgs)
	if argsPtr != nil && numArgs > 0 {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = &objects.Int{Value: argsSlice[i]}
		}
	}

	var result objects.Object

	switch CollectionOpKind(opKind) {
	case OpArrayEmpty:
		result = &objects.Array{Elements: []objects.Object{}}
	case OpArrayCreate:
		elems := make([]objects.Object, numArgs)
		copy(elems, args)
		result = &objects.Array{Elements: elems}
	case OpArrayAppend:
		if numArgs >= 2 {
			arr, ok := args[0].(*objects.Array)
			if ok {
				newElems := make([]objects.Object, len(arr.Elements), len(arr.Elements)+1)
				copy(newElems, arr.Elements)
				newElems = append(newElems, args[1])
				result = &objects.Array{Elements: newElems}
			}
		}
	case OpArrayGet:
		if numArgs >= 2 {
			arr, ok := args[0].(*objects.Array)
			if ok {
				idx, idxOk := args[1].(*objects.Int)
				if idxOk && int(idx.Value) < len(arr.Elements) && idx.Value >= 0 {
					result = arr.Elements[idx.Value]
				}
			}
		}
	case OpArraySet:
		if numArgs >= 3 {
			arr, ok := args[0].(*objects.Array)
			if ok {
				idx, idxOk := args[1].(*objects.Int)
				if idxOk && int(idx.Value) < len(arr.Elements) && idx.Value >= 0 {
					newElems := make([]objects.Object, len(arr.Elements))
					copy(newElems, arr.Elements)
					newElems[idx.Value] = args[2]
					result = &objects.Array{Elements: newElems}
				}
			}
		}
	case OpMapEmpty:
		result = &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	case OpMapCreate:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for i := 0; i+1 < numArgs; i += 2 {
			keyObj := args[i]
			valObj := args[i+1]
			hk := keyObj.HashKey()
			pairs[hk] = objects.MapPair{Key: keyObj, Value: valObj}
		}
		result = &objects.Map{Pairs: pairs}
	case OpMapSet:
		if numArgs >= 3 {
			m, ok := args[0].(*objects.Map)
			if ok {
				newPairs := make(map[objects.HashKey]objects.MapPair, len(m.Pairs)+1)
				for k, v := range m.Pairs {
					newPairs[k] = v
				}
				keyObj := args[1]
				hk := keyObj.HashKey()
				newPairs[hk] = objects.MapPair{Key: keyObj, Value: args[2]}
				result = &objects.Map{Pairs: newPairs}
			}
		}
	case OpMapGet:
		if numArgs >= 2 {
			m, ok := args[0].(*objects.Map)
			if ok {
				keyObj := args[1]
				hk := keyObj.HashKey()
				if pair, found := m.Pairs[hk]; found {
					result = pair.Value
				}
			}
		}
	default:
		return 0
	}

	if result == nil {
		return 0
	}

	return objectToNativeInt(result)
}

// objectCallbackFromNative is the Go callback invoked from native code for object ops.
// Parameters follow Windows x64 ABI: rcx=opKind, rdx=numArgs, r8=argsPtr, r9=nameIdx.
func objectCallbackFromNative(opKind int, numArgs int, argsPtr *int64, nameIdx int) int64 {
	globalObjectCtx.mu.Lock()
	jitVM := globalObjectCtx.vm
	globalObjectCtx.mu.Unlock()

	if jitVM == nil || jitVM.bytecode == nil {
		return 0
	}

	// Look up the name from the constant pool
	if nameIdx < 0 || nameIdx >= len(jitVM.bytecode.Constants) {
		return 0
	}
	nameObj := jitVM.bytecode.Constants[nameIdx]
	nameStr, ok := nameObj.(*objects.String)
	if !ok {
		return 0
	}
	name := nameStr.Value

	// Convert int64 args to objects.Object args
	args := make([]objects.Object, numArgs)
	if argsPtr != nil && numArgs > 0 {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = &objects.Int{Value: argsSlice[i]}
		}
	}

	var result objects.Object

	switch ObjectOpKind(opKind) {
	case OpGetField:
		if numArgs >= 1 {
			// Try to get a field/property from an object
			if instance, ok := args[0].(*objects.Instance); ok {
				if val, found := instance.Fields[name]; found {
					result = val
				}
			}
		}
	case OpSetField:
		if numArgs >= 2 {
			if instance, ok := args[0].(*objects.Instance); ok {
				instance.Fields[name] = args[1]
				result = args[1]
			}
		}
	case OpGetMethod:
		if numArgs >= 1 {
			if instance, ok := args[0].(*objects.Instance); ok {
				if method, found := instance.Class.Methods[name]; found {
					result = method
				}
			}
		}
	default:
		return 0
	}

	if result == nil {
		return 0
	}

	return objectToNativeInt(result)
}

// objectToNativeInt converts an objects.Object to an int64 for native code.
func objectToNativeInt(obj objects.Object) int64 {
	if obj == nil {
		return 0
	}
	switch v := obj.(type) {
	case *objects.Int:
		return v.Value
	case *objects.Bool:
		if v.Value {
			return 1
		}
		return 0
	case *objects.Null:
		return 0
	default:
		// For non-primitive types (Array, Map, etc.), we can't represent
		// them as int64 natively. Return 0 as a placeholder.
		// These objects should be handled via the hybrid interpreter path.
		return 0
	}
}

// InitBuiltinCallback initializes the builtin callback and associates it with the JITVM.
func (j *JITVM) InitBuiltinCallback() error {
	globalBuiltinCtx.mu.Lock()
	defer globalBuiltinCtx.mu.Unlock()

	globalBuiltinCtx.vm = j

	if globalBuiltinCtx.callback == 0 {
		cb := syscall.NewCallback(builtinCallbackFromNative)
		if cb == 0 {
			return fmt.Errorf("failed to create builtin callback")
		}
		globalBuiltinCtx.callback = cb
	}
	return nil
}

// InitCollectionCallback initializes the collection callback and associates it with the JITVM.
func (j *JITVM) InitCollectionCallback() error {
	globalCollectionCtx.mu.Lock()
	defer globalCollectionCtx.mu.Unlock()

	globalCollectionCtx.vm = j

	if globalCollectionCtx.callback == 0 {
		cb := syscall.NewCallback(collectionCallbackFromNative)
		if cb == 0 {
			return fmt.Errorf("failed to create collection callback")
		}
		globalCollectionCtx.callback = cb
	}
	return nil
}

// InitObjectCallback initializes the object callback and associates it with the JITVM.
func (j *JITVM) InitObjectCallback() error {
	globalObjectCtx.mu.Lock()
	defer globalObjectCtx.mu.Unlock()

	globalObjectCtx.vm = j

	if globalObjectCtx.callback == 0 {
		cb := syscall.NewCallback(objectCallbackFromNative)
		if cb == 0 {
			return fmt.Errorf("failed to create object callback")
		}
		globalObjectCtx.callback = cb
	}
	return nil
}

// GetBuiltinCallbackPtr returns the function pointer for builtin callbacks.
func GetBuiltinCallbackPtr() uintptr {
	globalBuiltinCtx.mu.Lock()
	defer globalBuiltinCtx.mu.Unlock()
	return globalBuiltinCtx.callback
}

// GetFunctionCallbackPtr returns the function pointer for function callbacks.
func GetFunctionCallbackPtr() uintptr {
	return getWindowsCallbackPtr()
}

// GetCollectionCallbackPtr returns the function pointer for collection callbacks.
func GetCollectionCallbackPtr() uintptr {
	globalCollectionCtx.mu.Lock()
	defer globalCollectionCtx.mu.Unlock()
	return globalCollectionCtx.callback
}

// GetObjectCallbackPtr returns the function pointer for object callbacks.
func GetObjectCallbackPtr() uintptr {
	globalObjectCtx.mu.Lock()
	defer globalObjectCtx.mu.Unlock()
	return globalObjectCtx.callback
}
