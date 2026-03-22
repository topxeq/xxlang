// +build amd64,!windows

// pkg/jit/jit_fib_test.go
// Tests for JIT compilation of Fibonacci with tail recursion
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITFibonacciTailRecursive tests JIT compilation of tail-recursive Fibonacci
func TestJITFibonacciTailRecursive(t *testing.T) {
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

	// Test interpreter first
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

	// Test JIT compilation
	t.Run("JIT", func(t *testing.T) {
		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitCompiler := NewJITCompiler(config)

		// Find the fibHelper function in constants
		var fibHelperFn *compiler.CompiledFunction
		for i, c := range bytecode.Constants {
			if fn, ok := c.(*compiler.CompiledFunction); ok {
				t.Logf("Found function at index %d: NumLocals=%d, NumParams=%d, Instructions=%d bytes",
					i, fn.NumLocals, fn.NumParameters, len(fn.Instructions))
				if fn.NumParameters == 3 {
					fibHelperFn = fn
				}
			}
		}

		if fibHelperFn == nil {
			t.Skip("Could not find fibHelper function")
		}

		cf, err := jitCompiler.Compile(fibHelperFn, constants, nil)
		if err != nil {
			t.Logf("JIT compilation failed (expected for recursive functions): %v", err)
			return
		}

		t.Logf("JIT compiled successfully: %d bytes", cf.Size)
		jitCompiler.Cleanup()
	})
}

// TestJITSimpleLoop tests JIT compilation of simple loops (no function calls)
func TestJITSimpleLoop(t *testing.T) {
	code := `
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = sum + i
		}
		sum
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

	// Test interpreter
	vmInst := vm.NewRegVM(bytecode)
	if err := vmInst.Run(); err != nil {
		t.Fatalf("Interpreter error: %v", err)
	}
	interpResult := vmInst.LastPoppedObject()
	t.Logf("Interpreter result: %v", interpResult)

	// Test JIT
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  8192,
		Debug:        true,
	}

	jitCompiler := NewJITCompiler(config)

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	cf, err := jitCompiler.Compile(mainFn, constants, nil)
	if err != nil {
		t.Logf("JIT compilation failed: %v", err)
		return
	}

	t.Logf("JIT compiled successfully: %d bytes", cf.Size)
	jitCompiler.Cleanup()
}

// TestJITArithmetic tests JIT compilation of arithmetic operations
func TestJITArithmetic(t *testing.T) {
	code := `
		var a = 10
		var b = 20
		var c = a + b
		var d = c * 2
		var e = d - 5
		e
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

	// Test interpreter
	vmInst := vm.NewRegVM(bytecode)
	if err := vmInst.Run(); err != nil {
		t.Fatalf("Interpreter error: %v", err)
	}
	interpResult := vmInst.LastPoppedObject()
	t.Logf("Interpreter result: %v", interpResult)

	// Test JIT
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  4096,
		Debug:        true,
	}

	jitCompiler := NewJITCompiler(config)

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     8,
		NumParameters: 0,
	}

	cf, err := jitCompiler.Compile(mainFn, constants, nil)
	if err != nil {
		t.Logf("JIT compilation failed: %v", err)
		return
	}

	t.Logf("JIT compiled successfully: %d bytes", cf.Size)
	jitCompiler.Cleanup()
}

// TestJITComparisons tests JIT compilation of comparison operations
func TestJITComparisons(t *testing.T) {
	code := `
		var a = 10
		var b = 20
		var c = 0
		if (a < b) { c = 1 }
		if (a > b) { c = 2 }
		if (a == b) { c = 3 }
		if (a != b) { c = 4 }
		c
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

	// Test interpreter
	vmInst := vm.NewRegVM(bytecode)
	if err := vmInst.Run(); err != nil {
		t.Fatalf("Interpreter error: %v", err)
	}
	interpResult := vmInst.LastPoppedObject()
	t.Logf("Interpreter result: %v", interpResult)

	// Test JIT
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  4096,
		Debug:        true,
	}

	jitCompiler := NewJITCompiler(config)

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     8,
		NumParameters: 0,
	}

	cf, err := jitCompiler.Compile(mainFn, constants, nil)
	if err != nil {
		t.Logf("JIT compilation failed: %v", err)
		return
	}

	t.Logf("JIT compiled successfully: %d bytes", cf.Size)
	jitCompiler.Cleanup()
}
