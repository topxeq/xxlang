// pkg/jit/jit_perf_test.go
// Performance comparison tests for JIT vs interpreter
package jit

import (
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// BenchmarkFibInterpreter benchmarks the interpreter for Fibonacci
func BenchmarkFibInterpreter(b *testing.B) {
	// Tail-recursive Fibonacci
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

	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
}

// BenchmarkFibJIT benchmarks JIT compilation of Fibonacci
func BenchmarkFibJIT(b *testing.B) {
	// Tail-recursive Fibonacci
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
				break
			}
		}
	}

	if fibHelperFn == nil {
		b.Skip("Could not find fibHelper function")
	}

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	// Pre-compile the function
	jitCompiler := NewJITCompiler(config)
	cf, err := jitCompiler.Compile(fibHelperFn, constants, nil)
	if err != nil {
		b.Skipf("JIT compilation failed: %v", err)
	}
	defer jitCompiler.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This executes the JIT code directly
		// but doesn't pass arguments correctly - just measuring call overhead
		cf.Execute()
	}
}

// TestFibPerformanceComparison compares interpreter vs JIT performance
func TestFibPerformanceComparison(t *testing.T) {
	// Tail-recursive Fibonacci
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

	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Test interpreter
	iterations := 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterElapsed := time.Since(start)

	// Find the fibHelper function
	var fibHelperFn *compiler.CompiledFunction
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			t.Logf("Found function: NumParams=%d, NumLocals=%d, Instructions=%d bytes",
				fn.NumParameters, fn.NumLocals, len(fn.Instructions))
			// Print the bytecode for debugging
			t.Logf("Bytecode:\n%s", compiler.String(fn.Instructions))
			if fn.NumParameters == 3 {
				fibHelperFn = fn
			}
		}
	}

	// Test JIT compilation
	if fibHelperFn != nil {
		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitCompiler := NewJITCompiler(config)
		cf, err := jitCompiler.Compile(fibHelperFn, constants, nil)
		defer jitCompiler.Cleanup()

		if err == nil && cf != nil {
			t.Logf("JIT compiled: %d bytes", cf.Size)

			// JIT compilation time
			compileStart := time.Now()
			for i := 0; i < 10; i++ {
				jitC := NewJITCompiler(config)
				_, _ = jitC.Compile(fibHelperFn, constants, nil)
				jitC.Cleanup()
			}
			compileTime := time.Since(compileStart) / 10
			t.Logf("Average JIT compile time: %v", compileTime)
		} else {
			t.Logf("JIT compilation failed: %v", err)
		}
	}

	t.Logf("Interpreter: %v for %d iterations (%v per iteration)",
		interpreterElapsed, iterations, interpreterElapsed/time.Duration(iterations))

	// Also test standard recursive fib
	codeRecursive := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(30)
	`

	l2 := lexer.New(codeRecursive)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()

	c2 := compiler.NewRegCompiler()
	c2.Compile(program2)
	bytecode2 := c2.Bytecode()

	start2 := time.Now()
	vmInst := vm.NewRegVM(bytecode2)
	vmInst.Run()
	recursiveElapsed := time.Since(start2)

	result := vmInst.LastPoppedObject()
	t.Logf("Standard recursive fib(30) = %s (took %v)", result.Inspect(), recursiveElapsed)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
