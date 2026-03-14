// examples/wasm_plugin/main.go
// Demonstrates using WebAssembly plugins in Xxlang.
//
// Unlike native .so plugins, WASM plugins work on all platforms including Windows,
// and don't require CGO.
//
// Prerequisites:
//   1. Install TinyGo: https://tinygo.org/getting-started/
//   2. Build the plugin: cd plugin && tinygo build -o fib.wasm -target=wasi fib.go
//   3. Run this example: go run main.go
package main

import (
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/interpreter"
	"github.com/topxeq/xxlang/pkg/plugin"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Xxlang WASM Plugin Demo")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("This demo shows WASM plugins that work on all platforms")
	fmt.Println("including Windows, without CGO.")
	fmt.Println()

	// Create plugin loader with search path
	loader := plugin.NewLoader()
	loader.AddPath("./plugin")

	// Load the WASM plugin
	fmt.Println("Loading fib.wasm plugin...")
	p, err := loader.Load("fib")
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		fmt.Println()
		fmt.Println("To build the plugin, run:")
		fmt.Println("  cd plugin && tinygo build -o fib.wasm -target=wasi fib.go")
		return
	}
	fmt.Printf("Loaded plugin: %s\n", p.Name())
	fmt.Println()

	// Create interpreter
	interp := interpreter.New(interpreter.WithStdlib())

	// Register plugin as a module
	// In a real implementation, this would be done automatically by the import system
	plugin.Register(p)

	// Test the plugin
	fmt.Println("==============================================")
	fmt.Println("1. Testing WASM Plugin Functions")
	fmt.Println("==============================================")
	fmt.Println()

	testCode := `
import "plugin/fib"

println("Plugin version: " + fib.version)

println("")
println("=== fib.fast (O(n) algorithm) ===")
println("fib.fast(10) = " + fib.fast(10).toStr())
println("fib.fast(50) = " + fib.fast(50).toStr())

println("")
println("=== fib.matrix (O(log n) algorithm) ===")
println("fib.matrix(10) = " + fib.matrix(10).toStr())
println("fib.matrix(50) = " + fib.matrix(50).toStr())

println("")
println("=== fib.isFib (Fibonacci check) ===")
println("isFib(13) = " + fib.isFib(13).toStr())
println("isFib(14) = " + fib.isFib(14).toStr())
`

	result, err := interp.Eval(testCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	_ = result

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("2. Performance Comparison")
	fmt.Println("==============================================")
	fmt.Println()

	// Performance test
	perfTest := `
import "plugin/fib"

// Xxlang tail recursion
func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)
}
`
	interp.Eval(perfTest)

	// Test Xxlang tail recursion
	start := time.Now()
	interp.Eval("fibTail(35, 0, 1)")
	tailTime := time.Since(start)

	// Test WASM plugin fast
	start = time.Now()
	interp.Eval("fib.fast(35)")
	fastTime := time.Since(start)

	// Test WASM plugin matrix
	start = time.Now()
	interp.Eval("fib.matrix(35)")
	matrixTime := time.Since(start)

	fmt.Println("fib(35) performance:")
	fmt.Printf("  Xxlang tail recursion:  %v\n", tailTime)
	fmt.Printf("  WASM fib.fast:          %v (O(n))\n", fastTime)
	fmt.Printf("  WASM fib.matrix:        %v (O(log n))\n", matrixTime)

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("3. Boundary Test (int64 max Fibonacci)")
	fmt.Println("==============================================")
	fmt.Println()

	start = time.Now()
	result, _ = interp.Eval("fib.matrix(92)")
	fmt.Printf("fib.matrix(92) = %s\n", result.Inspect())
	fmt.Printf("Time: %v\n", time.Since(start))

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Summary")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("WASM Plugin Benefits:")
	fmt.Println("1. Works on Windows, Linux, macOS")
	fmt.Println("2. No CGO required")
	fmt.Println("3. Cross-platform plugin files")
	fmt.Println("4. Sandboxed execution")
	fmt.Println()
	fmt.Println("Building WASM Plugins:")
	fmt.Println("  tinygo build -o plugin.wasm -target=wasi plugin.go")
}
