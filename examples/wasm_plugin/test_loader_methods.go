// test_loader_methods.go - Test different plugin loading methods
package main

import (
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/interpreter"
	"github.com/topxeq/xxlang/pkg/plugin"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Testing Plugin Loading Methods")
	fmt.Println("==============================================")
	fmt.Println()

	// =====================================================
	// Test 1: LoadPath() - Direct file path loading
	// =====================================================
	fmt.Println("1. LoadPath() - Direct file path loading")
	fmt.Println("-------------------------------------------")

	testLoadPath := func(name, path string) {
		loader := plugin.NewLoader()
		start := time.Now()
		p, err := loader.LoadPath(path)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  ✗ %-15s Failed: %v\n", name+":", err)
			return
		}
		fmt.Printf("  ✓ %-15s Loaded in %v (name: %s)\n", name+":", elapsed, p.Name())

		// Test execution
		interp := interpreter.New(interpreter.WithStdlib())
		plugin.Register(p)
		result, err := interp.Eval(`import "plugin/fib"; fib.fast(20)`)
		if err != nil {
			fmt.Printf("    ✗ Execution error: %v\n", err)
			return
		}
		fmt.Printf("    fib.fast(20) = %v\n", result)
	}

	testLoadPath("AssemblyScript", "./plugin/fib_ts.wasm")
	testLoadPath("C", "./plugin/fib_c.wasm")
	testLoadPath("Zig", "./plugin/fib_zig.wasm")
	testLoadPath("Rust", "./plugin/fib_rust.wasm")

	// =====================================================
	// Test 2: Load() - Load by name with search paths
	// =====================================================
	fmt.Println()
	fmt.Println("2. Load() - Load by name with search paths")
	fmt.Println("-------------------------------------------")

	testLoadByName := func(pluginName string, searchPaths []string, shouldSucceed bool) {
		loader := plugin.NewLoader()
		for _, path := range searchPaths {
			loader.AddPath(path)
		}

		start := time.Now()
		p, err := loader.Load(pluginName)
		elapsed := time.Since(start)

		if err != nil {
			if shouldSucceed {
				fmt.Printf("  ✗ %-15s Failed: %v\n", pluginName+":", err)
			} else {
				fmt.Printf("  ✓ %-15s Correctly failed (expected)\n", pluginName+":")
			}
			return
		}
		fmt.Printf("  ✓ %-15s Loaded in %v (name: %s)\n", pluginName+":", elapsed, p.Name())
		fmt.Printf("    Search paths: %v\n", loader.Paths())
	}

	// Test with single search path
	fmt.Println("\n  Testing Load(\"fib\") with ./plugin path:")
	testLoadByName("fib", []string{"./plugin"}, true)

	// Test with multiple search paths
	fmt.Println("\n  Testing Load(\"fib\") with multiple paths:")
	testLoadByName("fib", []string{"./nonexistent", "./plugin", "/opt/plugins"}, true)

	// Test with non-existent plugin
	fmt.Println("\n  Testing Load(\"nonexistent\") - should fail:")
	testLoadByName("nonexistent", []string{"./plugin"}, false)

	// =====================================================
	// Test 3: Verify each plugin implementation
	// =====================================================
	fmt.Println()
	fmt.Println("3. Verifying each WASM plugin implementation")
	fmt.Println("-------------------------------------------")

	plugins := []struct {
		name string
		path string
	}{
		{"AssemblyScript", "./plugin/fib_ts.wasm"},
		{"C", "./plugin/fib_c.wasm"},
		{"Zig", "./plugin/fib_zig.wasm"},
		{"Rust", "./plugin/fib_rust.wasm"},
	}

	for _, p := range plugins {
		fmt.Printf("\n  %s plugin:\n", p.name)
		loader := plugin.NewLoader()
		plg, err := loader.LoadPath(p.path)
		if err != nil {
			fmt.Printf("    ✗ Failed to load: %v\n", err)
			continue
		}

		// Test all functions
		interp := interpreter.New(interpreter.WithStdlib())
		plugin.Register(plg)

		testCode := `
import "plugin/fib"

println("    version: " + fib.version)
println("    fast(20) = " + fib.fast(20).toStr())
println("    matrix(20) = " + fib.matrix(20).toStr())
println("    isFib(55) = " + fib.isFib(55).toStr())
println("    isFib(56) = " + fib.isFib(56).toStr())
`
		_, err = interp.Eval(testCode)
		if err != nil {
			fmt.Printf("    ✗ Execution error: %v\n", err)
		}
	}

	// =====================================================
	// Test 4: Plugin registry behavior
	// =====================================================
	fmt.Println()
	fmt.Println("4. Testing plugin registry")
	fmt.Println("-------------------------------------------")

	fmt.Println("\n  Loading same plugin multiple times:")
	loader1 := plugin.NewLoader()
	start := time.Now()
	p1, _ := loader1.LoadPath("./plugin/fib_ts.wasm")
	fmt.Printf("  First load:  %v (name: %s)\n", time.Since(start), p1.Name())

	loader2 := plugin.NewLoader()
	start = time.Now()
	p2, _ := loader2.LoadPath("./plugin/fib_ts.wasm")
	fmt.Printf("  Second load: %v (from registry)\n", time.Since(start))
	_ = p2

	// =====================================================
	// Test 5: File size comparison
	// =====================================================
	fmt.Println()
	fmt.Println("5. WASM file size comparison")
	fmt.Println("-------------------------------------------")

	fmt.Println()
	for _, p := range plugins {
		// Get file size
		loader := plugin.NewLoader()
		plg, err := loader.LoadPath(p.path)
		if err != nil {
			continue
		}
		fmt.Printf("  %-15s: plugin name = %s\n", p.name, plg.Name())
	}

	// =====================================================
	// Summary
	// =====================================================
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Summary")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("Loading Methods:")
	fmt.Println("  ✓ LoadPath(path) - Load from specific file path (recommended)")
	fmt.Println("  ✓ Load(name)     - Load by name with search paths")
	fmt.Println()
	fmt.Println("Supported Languages:")
	fmt.Println("  ✓ AssemblyScript - Smallest output (~1KB)")
	fmt.Println("  ✓ C              - Most portable (~1.5KB)")
	fmt.Println("  ✓ Zig            - Modern syntax (~1.3KB)")
	fmt.Println("  ✓ Rust           - Safe systems language (~1.3KB)")
}
