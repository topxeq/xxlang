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
		if results, err := nameFn.Call(ctx); err == nil && len(results) >= 2 {
			// Go returns (ptr uint32, size uint32) as two separate values
			name = readStringFromMemory2(module, uint32(results[0]), uint32(results[1]))
		}
	}

	return NewWasmPlugin(name, module, rt), nil
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
