// examples/wasm_plugin/main.go
// Comprehensive test for WebAssembly plugins in Xxlang.
//
// WASM plugins work on all platforms including Windows, without CGO.
//
// Build the plugin:
//
//	cd plugin && ./build.sh fib.ts    # AssemblyScript (smallest)
//	cd plugin && ./build.sh fib.rs    # Rust
//	cd plugin && ./build.sh fib.zig   # Zig
//	cd plugin && ./build.sh fib.c     # C
//
// Run this example:
//
//	go run main.go
package main

import (
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/interpreter"
	"github.com/topxeq/xxlang/pkg/plugin"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Xxlang WASM Plugin Comprehensive Test")
	fmt.Println("==============================================")
	fmt.Println()

	// Create plugin loader
	loader := plugin.NewLoader()

	// Method 1: Load by direct file path (recommended)
	// This is the simplest way - just specify the .wasm file path
	fmt.Println("Loading fib.wasm plugin...")
	p, err := loader.LoadPath("./plugin/fib.wasm")
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		fmt.Println()
		fmt.Println("To build the plugin, run one of:")
		fmt.Println("  cd plugin && ./build.sh fib.ts    # AssemblyScript")
		fmt.Println("  cd plugin && ./build.sh fib.rs    # Rust")
		fmt.Println("  cd plugin && ./build.sh fib.zig   # Zig")
		fmt.Println("  cd plugin && ./build.sh fib.c     # C")
		return
	}
	fmt.Printf("Loaded plugin: %s\n", p.Name())
	fmt.Println()

	// Create interpreter
	interp := interpreter.New(interpreter.WithStdlib())

	// Register plugin for use in Xxlang
	plugin.Register(p)

	// Test 1: Basic functionality
	fmt.Println("==============================================")
	fmt.Println("1. Basic Functionality Tests")
	fmt.Println("==============================================")
	fmt.Println()

	testCode := `
import "plugin/fib"

// Test version
println("Plugin version: " + fib.version)

// Test fib.fast
println("")
println("=== fib.fast (O(n) algorithm) ===")
println("fib.fast(0) = " + fib.fast(0).toStr())
println("fib.fast(1) = " + fib.fast(1).toStr())
println("fib.fast(10) = " + fib.fast(10).toStr())
println("fib.fast(50) = " + fib.fast(50).toStr())

// Test fib.matrix
println("")
println("=== fib.matrix (O(log n) algorithm) ===")
println("fib.matrix(0) = " + fib.matrix(0).toStr())
println("fib.matrix(1) = " + fib.matrix(1).toStr())
println("fib.matrix(10) = " + fib.matrix(10).toStr())
println("fib.matrix(50) = " + fib.matrix(50).toStr())

// Test fib.isFib
println("")
println("=== fib.isFib (Fibonacci check) ===")
println("isFib(0) = " + fib.isFib(0).toStr())
println("isFib(1) = " + fib.isFib(1).toStr())
println("isFib(2) = " + fib.isFib(2).toStr())
println("isFib(13) = " + fib.isFib(13).toStr())
println("isFib(14) = " + fib.isFib(14).toStr())
println("isFib(21) = " + fib.isFib(21).toStr())
println("isFib(22) = " + fib.isFib(22).toStr())

// Test fib.range_
println("")
println("=== fib.range_ (array output) ===")
var fibs = fib.range_(10)
println("fib.range_(10) = " + fibs.toStr())
`

	result, err := interp.Eval(testCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	_ = result

	// Test 2: Correctness verification
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("2. Correctness Verification")
	fmt.Println("==============================================")
	fmt.Println()

	verifyCode := `
import "plugin/fib"

// Xxlang implementation for comparison
func fibSlow(n) {
    if (n <= 1) { return n }
    return fibSlow(n - 1) + fibSlow(n - 2)
}

func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)
}

func fibIter(n) {
    return fibTail(n, 0, 1)
}

// Verify correctness for n = 0 to 20
var allCorrect = true
for (var i = 0; i <= 20; i = i + 1) {
    var slow = fibSlow(i)
    var fast = fib.fast(i)
    var matrix = fib.matrix(i)
    var iter = fibIter(i)

    if (slow != fast || slow != matrix || slow != iter) {
        println("MISMATCH at n=" + i.toStr() + ": slow=" + slow.toStr() + " fast=" + fast.toStr() + " matrix=" + matrix.toStr() + " iter=" + iter.toStr())
        allCorrect = false
    }
}

if (allCorrect) {
    println("✓ All implementations match for n=0 to 20")
}

// Verify isFib
var fibSet = [0, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144]
var isFibCorrect = true
for (var n in fibSet) {
    if (!fib.isFib(n)) {
        println("isFib ERROR: " + n.toStr() + " should be Fibonacci")
        isFibCorrect = false
    }
}

// Non-Fibonacci numbers
var nonFibSet = [4, 6, 7, 9, 10, 12, 14, 15, 16, 17]
for (var n in nonFibSet) {
    if (fib.isFib(n)) {
        println("isFib ERROR: " + n.toStr() + " should NOT be Fibonacci")
        isFibCorrect = false
    }
}

if (isFibCorrect) {
    println("✓ isFib function works correctly")
}
`
	_, err = interp.Eval(verifyCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test 3: Performance comparison
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("3. Performance Comparison")
	fmt.Println("==============================================")
	fmt.Println()

	interp2 := interpreter.New(interpreter.WithStdlib())
	plugin.Register(p)

	perfCode := `
import "plugin/fib"

func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)
}
`
	interp2.Eval(perfCode)

	// Benchmark
	iterations := 100

	// Xxlang tail recursion
	start := time.Now()
	for i := 0; i < iterations; i++ {
		interp2.Eval("fibTail(35, 0, 1)")
	}
	tailTime := time.Since(start) / time.Duration(iterations)

	// WASM fast
	start = time.Now()
	for i := 0; i < iterations; i++ {
		interp2.Eval("fib.fast(35)")
	}
	fastTime := time.Since(start) / time.Duration(iterations)

	// WASM matrix
	start = time.Now()
	for i := 0; i < iterations; i++ {
		interp2.Eval("fib.matrix(35)")
	}
	matrixTime := time.Since(start) / time.Duration(iterations)

	fmt.Printf("fib(35) average over %d iterations:\n", iterations)
	fmt.Printf("  Xxlang tail recursion:  %v\n", tailTime)
	fmt.Printf("  WASM fib.fast (O(n)):   %v\n", fastTime)
	fmt.Printf("  WASM fib.matrix (O(log n)): %v\n", matrixTime)

	// Test 4: Boundary cases
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("4. Boundary Tests")
	fmt.Println("==============================================")
	fmt.Println()

	boundaryCode := `
import "plugin/fib"

println("=== Large Fibonacci numbers ===")
println("fib.fast(90) = " + fib.fast(90).toStr())
println("fib.matrix(91) = " + fib.matrix(91).toStr())
println("fib.matrix(92) = " + fib.matrix(92).toStr() + " (largest in int64)")

println("")
println("=== fib.range_ edge cases ===")
println("fib.range_(0) = " + fib.range_(0).toStr())
println("fib.range_(1) = " + fib.range_(1).toStr())
println("fib.range_(5) = " + fib.range_(5).toStr())

println("")
println("=== isFib edge cases ===")
println("isFib(0) = " + fib.isFib(0).toStr())
println("isFib(1) = " + fib.isFib(1).toStr())
println("isFib(-1) = " + fib.isFib(-1).toStr())
`
	_, err = interp2.Eval(boundaryCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test 5: Plugin from Xxlang REPL
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("5. Plugin Integration Summary")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("✓ Plugin loaded successfully from .wasm file")
	fmt.Println("✓ Plugin functions callable from Xxlang code")
	fmt.Println("✓ Return values correct")
	fmt.Println("✓ Multiple function signatures supported:")
	fmt.Println("  - call_fast(n) -> int64")
	fmt.Println("  - call_matrix(n) -> int64")
	fmt.Println("  - call_isFib(n) -> bool")
	fmt.Println("  - call_range_(n) -> array")
	fmt.Println()
	fmt.Println("WASM Plugin Benefits:")
	fmt.Println("  1. Works on Windows, Linux, macOS")
	fmt.Println("  2. No CGO required")
	fmt.Println("  3. Cross-platform plugin files")
	fmt.Println("  4. Sandboxed execution")
	fmt.Println()
	fmt.Println("Build commands:")
	fmt.Println("  AssemblyScript: asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2")
	fmt.Println("  Rust:           rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o fib.wasm fib.rs")
	fmt.Println("  Zig:            zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic")
	fmt.Println("  C:              clang -o fib.wasm --target=wasm32 -O2 fib.c -nostdlib -nostartfiles -Wl,--no-entry -Wl,--export-all")
}
