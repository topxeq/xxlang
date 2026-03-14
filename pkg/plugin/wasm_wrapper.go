// pkg/plugin/wasm_plugin.go
// WebAssembly plugin wrapper that implements the Plugin interface.
package plugin

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/topxeq/xxlang/pkg/objects"
)

// WasmPlugin wraps a WASM module to implement the Plugin interface.
type WasmPlugin struct {
	name   string
	module api.Module
	rt     interface{ Close(context.Context) error }
}

// NewWasmPlugin creates a new WasmPlugin from a loaded module.
func NewWasmPlugin(name string, module api.Module, rt interface{ Close(context.Context) error }) *WasmPlugin {
	return &WasmPlugin{
		name:   name,
		module: module,
		rt:     rt,
	}
}

// Name returns the plugin name.
func (p *WasmPlugin) Name() string {
	return p.name
}

// Exports returns the plugin's exported functions as Xxlang objects.
func (p *WasmPlugin) Exports() map[string]objects.Object {
	result := make(map[string]objects.Object)
	ctx := context.Background()

	// Get version if available
	if versionFn := p.module.ExportedFunction("plugin_version"); versionFn != nil {
		allocFn := p.module.ExportedFunction("alloc")
		if allocFn != nil {
			if allocResults, err := allocFn.Call(ctx, 8); err == nil && len(allocResults) > 0 {
				resultPtr := uint32(allocResults[0])
				if _, err := versionFn.Call(ctx, uint64(resultPtr)); err == nil {
					version := readStringFromResultPtr(p.module, resultPtr)
					result["version"] = &objects.String{Value: version}
				}
			}
		}
	}

	// Get all exported function definitions
	funcDefs := p.module.ExportedFunctionDefinitions()

	// Export all call_* functions
	for name := range funcDefs {
		// Skip internal functions
		if name == "plugin_name" || name == "plugin_version" || name == "alloc" {
			continue
		}

		// Get the function
		fn := p.module.ExportedFunction(name)
		if fn == nil {
			continue
		}

		// Export call_* functions (e.g., call_fast -> fast)
		exportName := name
		if len(name) > 5 && name[:5] == "call_" {
			exportName = name[5:]
		}

		// Create a wrapped builtin function
		result[exportName] = p.wrapFunction(fn, name)
	}

	return result
}

// wrapFunction wraps a WASM function as an Xxlang builtin.
func (p *WasmPlugin) wrapFunction(fn api.Function, name string) *objects.Builtin {
	return &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			ctx := context.Background()

			// Handle different function signatures
			switch name {
			case "call_fast", "call_matrix":
				// Single int64 argument, single int64 result
				if len(args) != 1 {
					return &objects.Error{Message: fmt.Sprintf("%s requires 1 argument", name)}
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				results, err := fn.Call(ctx, uint64(n.Value))
				if err != nil {
					return &objects.Error{Message: err.Error()}
				}
				if len(results) == 0 {
					return &objects.Error{Message: "function returned no result"}
				}
				return &objects.Int{Value: int64(results[0])}

			case "call_isFib":
				// Single int64 argument, returns int32 (0 or 1)
				if len(args) != 1 {
					return &objects.Error{Message: "isFib requires 1 argument"}
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				results, err := fn.Call(ctx, uint64(n.Value))
				if err != nil {
					return &objects.Error{Message: err.Error()}
				}
				if results[0] == 1 {
					return objects.TRUE
				}
				return objects.FALSE

			case "call_range_":
				// int64 argument + resultPtr, writes to memory
				if len(args) != 1 {
					return &objects.Error{Message: "range_ requires 1 argument"}
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				// Allocate memory for result pointer
				allocFn := p.module.ExportedFunction("alloc")
				if allocFn == nil {
					return &objects.Error{Message: "alloc function not found"}
				}

				allocResults, err := allocFn.Call(ctx, 8) // 8 bytes for ptr + count
				if err != nil {
					return &objects.Error{Message: err.Error()}
				}
				resultPtr := uint32(allocResults[0])

				// Call the function with n and resultPtr
				_, err = fn.Call(ctx, uint64(n.Value), uint64(resultPtr))
				if err != nil {
					return &objects.Error{Message: err.Error()}
				}

				// Read ptr and count from resultPtr
				mem := p.module.Memory()
				if mem == nil {
					return &objects.Error{Message: "memory not available"}
				}
				ptr, ok1 := mem.ReadUint32Le(resultPtr)
				count, ok2 := mem.ReadUint32Le(resultPtr + 4)
				if !ok1 || !ok2 {
					return &objects.Error{Message: "failed to read result"}
				}

				// Read array from memory
				arr := readInt64ArrayFromMemory(p.module, ptr, count)

				// Convert to Xxlang array
				elements := make([]objects.Object, len(arr))
				for i, v := range arr {
					elements[i] = &objects.Int{Value: v}
				}
				return &objects.Array{Elements: elements}

			default:
				// Generic handling: pass int64 arguments
				intArgs := make([]uint64, len(args))
				for i, arg := range args {
					if n, ok := arg.(*objects.Int); ok {
						intArgs[i] = uint64(n.Value)
					} else {
						return &objects.Error{Message: fmt.Sprintf("argument %d must be integer", i)}
					}
				}

				results, err := fn.Call(ctx, intArgs...)
				if err != nil {
					return &objects.Error{Message: err.Error()}
				}
				if len(results) == 0 {
					return objects.NULL
				}
				return &objects.Int{Value: int64(results[0])}
			}
		},
	}
}

// Close releases the WASM module resources.
func (p *WasmPlugin) Close(ctx context.Context) error {
	if p.module != nil {
		p.module.Close(ctx)
	}
	if p.rt != nil {
		return p.rt.Close(ctx)
	}
	return nil
}

// readStringFromMemory reads a string from WASM memory.
// ptrSize is a packed value: high 32 bits = offset, low 32 bits = size.
func readStringFromMemory(module api.Module, ptrSize uint64) string {
	offset := uint32(ptrSize >> 32)
	size := uint32(ptrSize & 0xFFFFFFFF)

	mem := module.Memory()
	if mem == nil || size == 0 {
		return ""
	}

	buf, ok := mem.Read(offset, size)
	if !ok {
		return ""
	}
	return string(buf)
}

// readInt64ArrayFromMemory reads an int64 array from WASM memory.
func readInt64ArrayFromMemory(module api.Module, offset uint32, count uint32) []int64 {
	if count == 0 {
		return nil
	}

	mem := module.Memory()
	if mem == nil {
		return nil
	}

	result := make([]int64, count)
	for i := uint32(0); i < count; i++ {
		val, ok := mem.ReadUint64Le(offset + i*8)
		if !ok {
			return nil
		}
		result[i] = int64(val)
	}
	return result
}
