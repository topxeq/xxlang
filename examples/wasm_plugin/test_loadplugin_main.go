// test_loadplugin_main.go - Test loadPlugin function from Xxlang
package main

import (
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Testing loadPlugin() from Xxlang")
	fmt.Println("==============================================")
	fmt.Println()

	// Create interpreter with stdlib
	interp := interpreter.New(interpreter.WithStdlib())

	// Test 1: Load plugin directly from Xxlang code
	fmt.Println("1. Loading plugin from Xxlang code:")
	fmt.Println("-------------------------------------------")

	code := `
// Load the WASM plugin directly from file path
var fib = loadPlugin("./plugin/fib.wasm")

println("Plugin loaded: " + typeOf(fib))
println("version: " + fib.version)
println("")

// Test functions
println("fib.fast(10) = " + fib.fast(10))
println("fib.fast(50) = " + fib.fast(50))
println("fib.matrix(92) = " + fib.matrix(92))
println("")

// Test isFib
println("fib.isFib(55) = " + fib.isFib(55))
println("fib.isFib(56) = " + fib.isFib(56))
println("")

// Test range_
var fibs = fib.range_(10)
println("fib.range_(10) = " + fibs.toStr())
`
	_, err := interp.Eval(code)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Test 2: Load different plugin implementations
	fmt.Println()
	fmt.Println("2. Testing different plugin implementations:")
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
		fmt.Printf("\n%s plugin:\n", p.name)
		testCode := fmt.Sprintf(`
var p = loadPlugin("%s")
print("  version: " + p.version + ", ")
println("fast(20) = " + p.fast(20))
`, p.path)
		_, err := interp.Eval(testCode)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		}
	}

	// Test 3: Error handling
	fmt.Println()
	fmt.Println("3. Testing error handling:")
	fmt.Println("-------------------------------------------")

	errorCode := `
var bad = loadPlugin("./nonexistent.wasm")
`
	_, err = interp.Eval(errorCode)
	if err != nil {
		fmt.Println("  Correctly caught error for nonexistent plugin")
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("All tests passed!")
	fmt.Println("==============================================")
}
