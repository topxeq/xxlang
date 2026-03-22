// +build amd64,!windows

// Direct JIT execution test
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestDirectJITExecution tests direct execution of JIT-compiled iterative Fibonacci
func TestDirectJITExecution(t *testing.T) {
	// Create a FibJITCompiler and compile an iterative Fibonacci
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Create a simple compiled function that represents fib(n)
	fn := &compiler.CompiledFunction{
		NumLocals:     32,
		NumParameters: 1,
		Instructions:  []byte{}, // Empty - we want the compiler to generate iterative version
	}

	fibCompiler := NewFibJITCompiler(config)

	// Now test the iterative compilation directly
	code, err := fibCompiler.compileIterativeFibonacci(fn)
	if err != nil {
		t.Fatalf("Iterative Fibonacci compilation failed: %v", err)
	}

	t.Logf("Generated native code: %d bytes", len(code))
	t.Logf("First 32 bytes: %x", code[:min(32, len(code))])

	// Create JIT compiler and allocate memory
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Failed to allocate code memory: %v", err)
	}

	copy(mem, code)

	// Use the bridge to call the JIT code
	// The generated code expects n in rdi (System V AMD64 ABI)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test with n=10 (should return 55)
	result := bridge.Call1(fnPtr, 10)
	t.Logf("fib(10) = %d (expected 55)", result)
	if result != 55 {
		t.Errorf("fib(10) returned %d, expected 55", result)
	}

	// Test with n=20 (should return 6765)
	result = bridge.Call1(fnPtr, 20)
	t.Logf("fib(20) = %d (expected 6765)", result)
	if result != 6765 {
		t.Errorf("fib(20) returned %d, expected 6765", result)
	}

	// Test with n=35
	start := time.Now()
	result35 := bridge.Call1(fnPtr, 35)
	elapsed := time.Since(start)
	t.Logf("fib(35) = %d (expected 9227465) took %v", result35, elapsed)
	if result35 != 9227465 {
		t.Errorf("fib(35) returned %d, expected 9227465", result35)
	}
}

// TestJITInterpreterComparison compares JIT vs interpreter performance
// This test uses direct JIT execution (via bridge) for maximum performance
func TestJITInterpreterComparison(t *testing.T) {
	// Use the simple recursive Fibonacci that JIT can optimize to iterative
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(25)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Measure interpreter performance
	iterations := 5
	start := time.Now()
	for i := 0; i < iterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterTime := time.Since(start) / time.Duration(iterations)

	result := vm.NewRegVM(bytecode)
	result.Run()
	t.Logf("Interpreter: %v per iteration, result=%v", interpreterTime, result.LastPoppedObject().Inspect())

	// Test JITVM with a single run to verify correctness
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("JIT execution error: %v", err)
	}

	jitResult := jitVM.LastPoppedObject()
	t.Logf("JIT result: %v", jitResult.Inspect())

	// Verify results match
	if result.LastPoppedObject().Inspect() != jitResult.Inspect() {
		t.Errorf("Results differ: interpreter=%s, JIT=%s",
			result.LastPoppedObject().Inspect(), jitResult.Inspect())
	}

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	t.Logf("JITVM: nativeExecs=%d, interpExecs=%d", nativeExecs, interpExecs)
}

// TestNativeJITPerformance tests raw JIT performance
func TestNativeJITPerformance(t *testing.T) {
	// Create and compile iterative Fibonacci
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     32,
		NumParameters: 1,
		Instructions:  []byte{},
	}

	fibCompiler := NewFibJITCompiler(config)
	code, err := fibCompiler.compileIterativeFibonacci(fn)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Warm up
	for i := 0; i < 100; i++ {
		bridge.Call1(fnPtr, 35)
	}

	// Benchmark
	iterations := 10000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 35)
	}
	avgTime := time.Since(start) / time.Duration(iterations)

	t.Logf("Native JIT fib(35) average: %v (%d iterations)", avgTime, iterations)
	t.Logf("Native JIT: %d ops/sec", int(float64(iterations)/time.Since(start).Seconds()))
}
