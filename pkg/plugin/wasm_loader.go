// pkg/plugin/plugin_wasm.go
// WebAssembly plugin loader using wazero runtime.
// Works on all platforms including Windows without CGO.
package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
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

	// Create wazero runtime with default config
	rt := wazero.NewRuntime(ctx)

	// Instantiate the WASM module
	module, err := rt.Instantiate(ctx, wasmBytes)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate wasm module: %v", err)
	}

	// Get plugin name from exported function if available
	name := "wasm_plugin"
	if nameFn := module.ExportedFunction("plugin_name"); nameFn != nil {
		if results, err := nameFn.Call(ctx); err == nil && len(results) > 0 {
			name = readStringFromMemory(module, results[0])
		}
	}

	return NewWasmPlugin(name, module, rt), nil
}
