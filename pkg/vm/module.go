// pkg/vm/module.go
// Module loading functionality for the VM.
package vm

import (
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/plugin"
	"github.com/topxeq/xxlang/pkg/stdlib"
)

// loadModuleFile loads, compiles, and executes a module file.
// It handles caching and circular dependency detection.
// Supports module types:
//   - std/* : Standard library modules
//   - plugin/* : WebAssembly plugins by name
//   - *.wasm : WebAssembly plugins by file path
//   - *.xxl : xxlang source files
func (vm *VM) loadModuleFile(resolvedPath string) (*objects.Module, error) {
	// Check if it's a WASM plugin by name (plugin/xxx)
	if strings.HasPrefix(resolvedPath, "plugin/") {
		return vm.loadWasmPlugin(resolvedPath)
	}

	// Check if it's a WASM plugin by file path (*.wasm)
	if strings.HasSuffix(resolvedPath, ".wasm") {
		return vm.loadWasmPluginByPath(resolvedPath)
	}

	// Check if it's a standard library module
	if strings.HasPrefix(resolvedPath, "std/") {
		if stdlib.Has(resolvedPath) {
			stdMod := stdlib.Get(resolvedPath)
			if stdMod == nil {
				return nil, fmt.Errorf("stdlib module not found: %s", resolvedPath)
			}
			// Convert stdlib.Module to objects.Module
			return &objects.Module{
				Name:    stdMod.Name,
				Exports: stdMod.Exports,
				Globals: nil, // stdlib modules don't have isolated globals
			}, nil
		}
	}

    // Check cache first
    if vm.loader.HasModule(resolvedPath) {
        cachedMod, err := vm.loader.Get(resolvedPath)
        if err != nil {
            return nil, err
        }
        // Convert module.Module to objects.Module
        return &objects.Module{
            Name:    cachedMod.Name,
            Exports: cachedMod.Exports,
            Globals: cachedMod.Globals,
        }, nil
    }

    // Check for circular dependency
    if vm.loader.IsLoading(resolvedPath) {
        return nil, fmt.Errorf("circular import: %s", resolvedPath)
    }

    // Read the module file
    code, err := os.ReadFile(resolvedPath)
    if err != nil {
        return nil, fmt.Errorf("module not found: %s", resolvedPath)
    }
    // Parse the module
    l := lexer.New(string(code))
    p := parser.New(l)
    program := p.ParseProgram()

    if len(p.Errors()) > 0 {
        return nil, fmt.Errorf("parse errors in module %s: %v", resolvedPath, p.Errors())
    }
    // Compile the module
    c := compiler.New()
    if err := c.Compile(program); err != nil {
        return nil, fmt.Errorf("compile error in module %s: %v", resolvedPath, err)
    }
    // Create the module object (before execution for circular ref support)
    mod := &objects.Module{
        Name:    resolvedPath,
        Exports: make(map[string]objects.Object),
    }
    // Mark as loading (for cycle detection)
    vm.loader.MarkLoading(resolvedPath)
    // Execute the module in an isolated VM context
    moduleVM := NewWithGlobalsStore(c.Bytecode(), make([]objects.Object, GlobalsSize))
    moduleVM.SetLoader(vm.loader)         // Share loader for nested imports
    moduleVM.SetSourcePath(resolvedPath)   // For nested imports
    moduleVM.SetCurrentModule(mod)         // For OpSetExport
    if err := moduleVM.Run(); err != nil {
        vm.loader.MarkDone(resolvedPath) // Clean up loading state on error
        return nil, fmt.Errorf("runtime error in module %s: %v", resolvedPath, err)
    }
    // Mark as done loading
    vm.loader.MarkDone(resolvedPath)
    // Store the module's globals for exported functions to access
    mod.Globals = moduleVM.Globals()
    // Cache the module using the module package's Module type
    vm.loader.Set(resolvedPath, &module.Module{
        Name:    mod.Name,
        Exports: mod.Exports,
    })
    return mod, nil
}

// loadWasmPlugin loads a WebAssembly plugin (.wasm file).
// Plugin path format: plugin/name (e.g., plugin/fib)
//
// WASM plugins work on all platforms (Windows, Linux, macOS) without CGO.
// Plugins can be written in TinyGo, Rust, C/C++, Zig, AssemblyScript, etc.
func (vm *VM) loadWasmPlugin(path string) (*objects.Module, error) {
	// Extract plugin name from path (plugin/name -> name)
	pluginName := strings.TrimPrefix(path, "plugin/")
	if pluginName == "" {
		return nil, fmt.Errorf("invalid plugin path: %s", path)
	}

	// Check if already loaded
	if vm.loader.HasModule(path) {
		cachedMod, err := vm.loader.Get(path)
		if err != nil {
			return nil, err
		}
		return &objects.Module{
			Name:    cachedMod.Name,
			Exports: cachedMod.Exports,
			Globals: cachedMod.Globals,
		}, nil
	}

	// Create a plugin loader
	loader := plugin.NewLoader()

	// Add plugin search paths from environment or defaults
	if pluginPath := os.Getenv("XXLANG_PLUGIN_PATH"); pluginPath != "" {
		for _, p := range strings.Split(pluginPath, ":") {
			loader.AddPath(p)
		}
	}

	// Also search relative to source path if available
	if vm.sourcePath != "" {
		loader.AddPath(vm.sourcePath + "/plugins")
		loader.AddPath(vm.sourcePath + "/plugin")
	}

	// Load the plugin
	p, err := loader.Load(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM plugin %s: %v", pluginName, err)
	}

	// Convert plugin to module
	mod := plugin.ToModule(p)

	// Cache the module
	vm.loader.Set(path, &module.Module{
		Name:    mod.Name,
		Exports: mod.Exports,
	})

	return mod, nil
}

// loadPluginByPath loads a WASM plugin from a specific file path.
// This is called by the loadPlugin() built-in function from Xxlang code.
// Returns the plugin as a Module object that can be used in Xxlang.
func (vm *VM) loadPluginByPath(wasmPath string) (objects.Object, error) {
	// Check if already loaded (use absolute path as cache key)
	absPath := wasmPath
	if vm.sourcePath != "" && !strings.HasPrefix(wasmPath, "/") {
		absPath = vm.sourcePath + "/" + wasmPath
	}

	cacheKey := "wasm:" + absPath
	if vm.loader.HasModule(cacheKey) {
		cachedMod, err := vm.loader.Get(cacheKey)
		if err != nil {
			return nil, err
		}
		return &objects.Module{
			Name:    cachedMod.Name,
			Exports: cachedMod.Exports,
			Globals: cachedMod.Globals,
		}, nil
	}

	// Create a plugin loader and load from the specified path
	loader := plugin.NewLoader()
	p, err := loader.LoadPath(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM plugin '%s': %v", wasmPath, err)
	}

	// Convert plugin to module
	mod := plugin.ToModule(p)

	// Cache the module
	vm.loader.Set(cacheKey, &module.Module{
		Name:    mod.Name,
		Exports: mod.Exports,
	})

	return mod, nil
}

// loadWasmPluginByPath loads a WASM plugin from a specific file path.
// This is called when importing a .wasm file via import statement.
func (vm *VM) loadWasmPluginByPath(wasmPath string) (*objects.Module, error) {
	// Check if already loaded
	cacheKey := "wasm:" + wasmPath
	if vm.loader.HasModule(cacheKey) {
		cachedMod, err := vm.loader.Get(cacheKey)
		if err != nil {
			return nil, err
		}
		return &objects.Module{
			Name:    cachedMod.Name,
			Exports: cachedMod.Exports,
			Globals: cachedMod.Globals,
		}, nil
	}

	// Create a plugin loader and load from the specified path
	loader := plugin.NewLoader()
	p, err := loader.LoadPath(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM plugin '%s': %v", wasmPath, err)
	}

	// Convert plugin to module
	mod := plugin.ToModule(p)

	// Cache the module
	vm.loader.Set(cacheKey, &module.Module{
		Name:    mod.Name,
		Exports: mod.Exports,
	})

	return mod, nil
}
