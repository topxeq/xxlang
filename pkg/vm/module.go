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
	"github.com/topxeq/xxlang/pkg/stdlib"
)

// loadModuleFile loads, compiles, and executes a module file.
// It handles caching and circular dependency detection.
func (vm *VM) loadModuleFile(resolvedPath string) (*objects.Module, error) {
	// Check if it's a standard library module first
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
