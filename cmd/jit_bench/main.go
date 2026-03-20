// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     Xxlang JIT vs Interpreter Performance Comparison           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Test 1: Tail-recursive Fibonacci
	fmt.Println("━━━ Test 1: Tail-Recursive Fibonacci (TCO) ━━━")
	runTailRecursiveFibTest()

	fmt.Println()

	// Test 2: Standard recursive Fibonacci
	fmt.Println("━━━ Test 2: Standard Recursive Fibonacci ━━━")
	runStandardRecursiveFibTest()

	fmt.Println()

	// Test 3: JIT compilation stats
	fmt.Println("━━━ Test 3: JIT Compilation Stats ━━━")
	runJITStatsTest()

	fmt.Println()

	// Test 4: Comparison with native Go
	fmt.Println("━━━ Test 4: Native Go Comparison ━━━")
	runNativeGoComparison()
}

func runTailRecursiveFibTest() {
	// Tail-recursive Fibonacci with TCO
	code := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
		func fib(n) {
			return fibHelper(n, 0, 1)
		}
		fib(35)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Interpreter benchmark
	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterTime := time.Since(start) / time.Duration(iterations)

	// Get result
	vmInst := vm.NewRegVM(bytecode)
	vmInst.Run()
	result := vmInst.LastPoppedObject()

	fmt.Printf("Result: fib(35) = %s\n", result.Inspect())
	fmt.Printf("Interpreter (TCO): %v per iteration\n", interpreterTime)
}

func runStandardRecursiveFibTest() {
	// Standard recursive Fibonacci (O(2^n))
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(30)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Interpreter benchmark
	start := time.Now()
	vmInst := vm.NewRegVM(bytecode)
	vmInst.Run()
	interpreterTime := time.Since(start)

	result := vmInst.LastPoppedObject()

	fmt.Printf("Result: fib(30) = %s\n", result.Inspect())
	fmt.Printf("Interpreter (naive): %v\n", interpreterTime)
	fmt.Printf("Warning: O(2^n) complexity - use TCO version for large n\n")
}

func runJITStatsTest() {
	code := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
		fibHelper
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Find the fibHelper function
	var fibHelperFn *compiler.CompiledFunction
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			if fn.NumParameters == 3 {
				fibHelperFn = fn
			}
		}
	}

	if fibHelperFn == nil {
		fmt.Println("Could not find fibHelper function")
		return
	}

	config := jit.JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	// Measure compilation time
	compileTimes := make([]time.Duration, 10)
	var totalSize int
	for i := 0; i < 10; i++ {
		jitCompiler := jit.NewJITCompiler(config)
		start := time.Now()
		cf, err := jitCompiler.Compile(fibHelperFn, constants, nil)
		compileTimes[i] = time.Since(start)

		if err != nil {
			fmt.Printf("Compilation %d failed: %v\n", i, err)
			jitCompiler.Cleanup()
			continue
		}
		totalSize = cf.Size
		jitCompiler.Cleanup()
	}

	// Calculate average
	var total time.Duration
	for _, t := range compileTimes {
		total += t
	}
	avgCompileTime := total / time.Duration(len(compileTimes))

	fmt.Printf("JIT compilation time: %v (avg over 10 runs)\n", avgCompileTime)
	fmt.Printf("Generated code size: %d bytes (from %d bytes bytecode)\n", totalSize, len(fibHelperFn.Instructions))
	fmt.Printf("Code expansion ratio: %.1fx\n", float64(totalSize)/float64(len(fibHelperFn.Instructions)))
}

func runNativeGoComparison() {
	// Native Go iterative Fibonacci
	goFib := func(n int64) int64 {
		if n <= 1 {
			return n
		}
		var a, b int64 = 0, 1
		for i := int64(2); i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}

	// Warm up
	goFib(35)

	// Benchmark
	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		goFib(35)
	}
	goTime := time.Since(start) / time.Duration(iterations)

	result := goFib(35)

	fmt.Printf("Result: fib(35) = %d\n", result)
	fmt.Printf("Native Go: %v per iteration\n", goTime)
	fmt.Printf("\nSpeed comparison (for same fib(35) calculation):\n")

	// Compare with interpreter TCO version
	code := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
		func fib(n) {
			return fibHelper(n, 0, 1)
		}
		fib(35)
	`
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	iterIterations := 1000
	start = time.Now()
	for i := 0; i < iterIterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterTime := time.Since(start) / time.Duration(iterIterations)

	fmt.Printf("  Native Go:       %v\n", goTime)
	fmt.Printf("  Xxlang (TCO):    %v\n", interpreterTime)
	fmt.Printf("  Interpreter overhead: %.1fx slower than native\n", float64(interpreterTime)/float64(goTime))
}
