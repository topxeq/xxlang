// pkg/plugin/wasm_loader.go
// WebAssembly plugin loader using wazero runtime.
// Works on all platforms including Windows without CGO.
package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// loadPluginWASM loads a .wasm file using wazero runtime.
// This works on all platforms including Windows without CGO.
//
// Note: Requires TinyGo-compiled WASM. Standard Go's wasip1 target
// runs _start (main) and exits, closing the module.
func loadPluginWASM(path string) (Plugin, error) {
	// Read the WASM file
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %v", err)
	}

	// Create context
	ctx := context.Background()

	// Create wazero runtime
	rt := wazero.NewRuntime(ctx)

	// Instantiate WASI, required for Go-compiled WASM (wasip1)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	// Instantiate the WASM module
	module, err := rt.Instantiate(ctx, wasmBytes)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate wasm module: %v", err)
	}

	// Get plugin name from exported function if available
	name := "wasm_plugin"
	if nameFn := module.ExportedFunction("plugin_name"); nameFn != nil {
		// Allocate memory for result
		allocFn := module.ExportedFunction("alloc")
		if allocFn != nil {
			if results, err := allocFn.Call(ctx, 8); err == nil && len(results) > 0 {
				resultPtr := uint32(results[0])
				if _, err := nameFn.Call(ctx, uint64(resultPtr)); err == nil {
					name = readStringFromResultPtr(module, resultPtr)
				}
			}
		}
	}

	return NewWasmPlugin(name, module, rt), nil
}

// readStringFromResultPtr reads a string from WASM memory.
// The resultPtr points to two uint32s: (string_ptr, string_size)
func readStringFromResultPtr(module api.Module, resultPtr uint32) string {
	mem := module.Memory()
	if mem == nil {
		return ""
	}
	// Read ptr and size from resultPtr
	ptr, ok1 := mem.ReadUint32Le(resultPtr)
	size, ok2 := mem.ReadUint32Le(resultPtr + 4)
	if !ok1 || !ok2 || size == 0 {
		return ""
	}
	buf, ok := mem.Read(ptr, size)
	if !ok {
		return ""
	}
	return string(buf)
}

// readStringFromMemory2 reads a string from WASM memory using separate ptr and size.
func readStringFromMemory2(module api.Module, ptr, size uint32) string {
	if size == 0 {
		return ""
	}
	mem := module.Memory()
	if mem == nil {
		return ""
	}
	buf, ok := mem.Read(ptr, size)
	if !ok {
		return ""
	}
	return string(buf)
}
