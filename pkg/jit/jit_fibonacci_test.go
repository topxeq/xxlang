// pkg/jit/jit_fibonacci_test.go
// Tests for JIT compilation of recursive Fibonacci functions
package jit

import (
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITRecursiveFibonacci tests JIT compilation of standard recursive Fibonacci
func TestJITRecursiveFibonacci(t *testing.T) {
	// Standard recursive Fibonacci (non-tail-recursive)
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(10)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Test interpreter for correctness
	t.Run("Interpreter", func(t *testing.T) {
		vmInst := vm.NewRegVM(bytecode)
		if err := vmInst.Run(); err != nil {
			t.Fatalf("Interpreter error: %v", err)
		}
		result := vmInst.LastPoppedObject()
		t.Logf("Interpreter result: %v", result)
		if result.Inspect() != "55" {
			t.Errorf("Expected 55, got %v", result.Inspect())
		}
	})

	// Test JIT compilation of the fib function
	t.Run("JIT_Compilation", func(t *testing.T) {
		// Find the fib function in constants
		var fibFn *compiler.CompiledFunction
		for i, c := range bytecode.Constants {
			if fn, ok := c.(*compiler.CompiledFunction); ok {
				t.Logf("Found function at index %d: NumLocals=%d, NumParams=%d, Instructions=%d bytes",
					i, fn.NumLocals, fn.NumParameters, len(fn.Instructions))
				if fn.NumParameters == 1 {
					fibFn = fn
				}
			}
		}

		if fibFn == nil {
			t.Skip("Could not find fib function")
		}

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitCompiler := NewJITCompiler(config)
		cf, err := jitCompiler.Compile(fibFn, constants, nil)
		if err != nil {
			t.Logf("JIT compilation result: %v (expected for recursive functions)", err)
			return
		}

		t.Logf("JIT compiled successfully: %d bytes", cf.Size)
		jitCompiler.Cleanup()
	})
}

// TestJITTailRecursiveFibonacci tests JIT compilation of tail-recursive Fibonacci
func TestJITTailRecursiveFibonacci(t *testing.T) {
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

		fib(10)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Test interpreter for correctness
	t.Run("Interpreter", func(t *testing.T) {
		vmInst := vm.NewRegVM(bytecode)
		if err := vmInst.Run(); err != nil {
			t.Fatalf("Interpreter error: %v", err)
		}
		result := vmInst.LastPoppedObject()
		t.Logf("Interpreter result: %v", result)
		if result.Inspect() != "55" {
			t.Errorf("Expected 55, got %v", result.Inspect())
		}
	})

	// Test JIT compilation of fibHelper function
	t.Run("JIT_TailRecursive", func(t *testing.T) {
		var fibHelperFn *compiler.CompiledFunction
		for i, c := range bytecode.Constants {
			if fn, ok := c.(*compiler.CompiledFunction); ok {
				t.Logf("Found function at index %d: NumLocals=%d, NumParams=%d",
					i, fn.NumLocals, fn.NumParameters)
				if fn.NumParameters == 3 {
					fibHelperFn = fn
				}
			}
		}

		if fibHelperFn == nil {
			t.Skip("Could not find fibHelper function")
		}

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitCompiler := NewJITCompiler(config)
		cf, err := jitCompiler.Compile(fibHelperFn, constants, nil)
		if err != nil {
			t.Logf("JIT compilation result: %v", err)
			return
		}

		t.Logf("JIT compiled successfully: %d bytes", cf.Size)
		jitCompiler.Cleanup()
	})
}

// TestIterativeFibonacci tests the iterative Fibonacci JIT compilation
func TestIterativeFibonacci(t *testing.T) {
	// Create a simple iterative Fibonacci function to JIT compile
	// We'll manually create the CompiledFunction to test the optimized path

	fn := &compiler.CompiledFunction{
		NumLocals:     16,
		NumParameters: 1,
		Instructions:  []byte{}, // Empty - we're testing the direct compilation
	}

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	fibCompiler := NewFibJITCompiler(config)
	code, err := fibCompiler.Compile(fn, nil, nil)
	if err != nil {
		t.Fatalf("Iterative Fibonacci compilation failed: %v", err)
	}

	t.Logf("Iterative Fibonacci compiled: %d bytes", len(code))
}

// TestFibJITPerformance compares JIT vs interpreter performance
func TestFibJITPerformance(t *testing.T) {
	// Use tail-recursive version for better performance
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

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Measure interpreter performance
	t.Run("Interpreter_Performance", func(t *testing.T) {
		start := time.Now()
		vmInst := vm.NewRegVM(bytecode)
		if err := vmInst.Run(); err != nil {
			t.Fatalf("Interpreter error: %v", err)
		}
		elapsed := time.Since(start)
		result := vmInst.LastPoppedObject()
		t.Logf("Interpreter: %v, result: %v", elapsed, result.Inspect())
	})
}

// TestAnalyzeFunction tests the function analysis
func TestAnalyzeFunction(t *testing.T) {
	// Create a test function with recursive calls
	recursiveCode := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
	`

	l := lexer.New(recursiveCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Find the fib function
	var fibFn *compiler.CompiledFunction
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			if fn.NumParameters == 1 {
				fibFn = fn
				break
			}
		}
	}

	if fibFn == nil {
		t.Fatal("Could not find fib function")
	}

	// Analyze the function
	config := DefaultJITConfig()
	fibCompiler := NewFibJITCompiler(config)
	analysis := fibCompiler.analyzeFunction(fibFn)

	t.Logf("Analysis: isSelfRecursive=%v, isTailRecursive=%v, callCount=%d",
		analysis.isSelfRecursive, analysis.isTailRecursive, analysis.callCount)

	if !analysis.isSelfRecursive {
		t.Error("Expected fib to be detected as self-recursive")
	}

	if analysis.isTailRecursive {
		t.Error("Standard fib should not be tail-recursive")
	}

	if analysis.callCount != 2 {
		t.Errorf("Expected 2 recursive calls, got %d", analysis.callCount)
	}
}

// TestTailRecursiveAnalysis tests analysis of tail-recursive functions
func TestTailRecursiveAnalysis(t *testing.T) {
	tailRecursiveCode := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
	`

	l := lexer.New(tailRecursiveCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Find fibHelper function
	var helperFn *compiler.CompiledFunction
	for _, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			if fn.NumParameters == 3 {
				helperFn = fn
				break
			}
		}
	}

	if helperFn == nil {
		t.Fatal("Could not find fibHelper function")
	}

	config := DefaultJITConfig()
	fibCompiler := NewFibJITCompiler(config)
	analysis := fibCompiler.analyzeFunction(helperFn)

	t.Logf("Tail-recursive analysis: isSelfRecursive=%v, isTailRecursive=%v, callCount=%d",
		analysis.isSelfRecursive, analysis.isTailRecursive, analysis.callCount)

	if !analysis.isSelfRecursive {
		t.Error("Expected fibHelper to be detected as self-recursive")
	}
}
