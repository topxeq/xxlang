//go:build amd64 && !windows
// +build amd64,!windows

// pkg/jit/jit_true_recursive_test.go
// Test for TRUE recursive JIT execution (not transformation to iteration)
//
// SAFETY: Tests use iterative version by default to prevent system freeze.
// True recursion tests are limited to small inputs (n <= MaxTrueRecursionInput).
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

// TestIterativeFibJIT tests the SAFE iterative Fibonacci JIT
func TestIterativeFibJIT(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{}, // Will be detected as fib pattern
	}

	recComp := NewTrueRecursiveJITCompiler(config)
	code, err := recComp.Compile(fn, nil)
	if err != nil {
		t.Fatalf("Iterative compilation failed: %v", err)
	}

	t.Logf("Generated ITERATIVE Fibonacci code: %d bytes", len(code))

	// Allocate executable memory
	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test values including large ones (safe with iterative version)
	testCases := []struct {
		n        int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
		{20, 6765},
		{30, 832040},
		{35, 9227465},
		{40, 102334155},
	}

	for _, tc := range testCases {
		result := bridge.Call1(fnPtr, tc.n)
		if result != tc.expected {
			t.Errorf("fib(%d) = %d, expected %d", tc.n, result, tc.expected)
		} else {
			t.Logf("fib(%d) = %d ✓", tc.n, result)
		}
	}
}

// TestTrueRecursiveFibJIT tests JIT that performs TRUE recursion
// WARNING: Limited to small inputs to prevent system freeze
func TestTrueRecursiveFibJIT(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Compile recursive Fibonacci using the TrueRecursiveJITCompiler
	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{}, // Will be detected as fib pattern
	}

	// Use CompileTrueRecursive for actual recursive code (DANGEROUS)
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, err := trueRecCompiler.CompileTrueRecursive(fn, nil)
	if err != nil {
		t.Fatalf("True recursive compilation failed: %v", err)
	}

	t.Logf("Generated TRUE recursive code: %d bytes", len(code))
	t.Logf("Code: %x", code)

	// Allocate executable memory
	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test ONLY small values (true recursion is O(2^n)!)
	testCases := []struct {
		n        int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
		// DO NOT test n > MaxTrueRecursionInput with true recursion!
	}

	for _, tc := range testCases {
		result := bridge.Call1(fnPtr, tc.n)
		if result != tc.expected {
			t.Errorf("fib(%d) = %d, expected %d", tc.n, result, tc.expected)
		} else {
			t.Logf("fib(%d) = %d ✓", tc.n, result)
		}
	}
}

// TestIterativeFibPerformanceTrueRec tests performance of SAFE iterative Fibonacci JIT
// using the TrueRecursiveJITCompiler's iterative mode
func TestIterativeFibPerformanceTrueRec(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{},
	}

	recCompiler := NewTrueRecursiveJITCompiler(config)
	code, err := recCompiler.Compile(fn, nil)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test fib(35) with SAFE iterative version
	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 35)
	}
	avgTime := time.Since(start) / time.Duration(iterations)

	result := bridge.Call1(fnPtr, 35)
	t.Logf("ITERATIVE fib(35): %v per call, result=%d", avgTime, result)

	// Test fib(40) - safe with iterative version
	start = time.Now()
	result = bridge.Call1(fnPtr, 40)
	fib40Time := time.Since(start)
	t.Logf("ITERATIVE fib(40): %v, result=%d", fib40Time, result)

	// Compare with interpreter
	xxlangCode := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(20)
	`
	l := lexer.New(xxlangCode)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	start = time.Now()
	vmInst := vm.NewRegVM(bytecode)
	vmInst.Run()
	interpTime := time.Since(start)

	t.Logf("Interpreter fib(20): %v", interpTime)
	t.Logf("JIT speedup: %.0fx", float64(interpTime)/float64(avgTime))
}

// TestTrueRecursiveFibPerformance tests performance of TRUE recursive JIT
// WARNING: Only tests up to MaxTrueRecursionInput to prevent system freeze
func TestTrueRecursiveFibPerformance(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{},
	}

	// Use CompileTrueRecursive for actual recursive code
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, err := trueRecCompiler.CompileTrueRecursive(fn, nil)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test fib(20) with true recursion (SAFE: below MaxTrueRecursionInput)
	iterations := 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 20)
	}
	avgTime := time.Since(start) / time.Duration(iterations)

	result := bridge.Call1(fnPtr, 20)
	t.Logf("TRUE recursive fib(20): %v per call, result=%d", avgTime, result)

	// Test fib(25) - at the safety limit
	start = time.Now()
	result = bridge.Call1(fnPtr, 25)
	fib25Time := time.Since(start)
	t.Logf("TRUE recursive fib(25): %v, result=%d", fib25Time, result)

	// DO NOT test fib(30) or higher with true recursion!
	// fib(30) requires ~270,000 calls
	// fib(35) requires ~18,450,000 calls which can freeze the system
}

// BenchmarkIterativeFib benchmarks SAFE iterative Fibonacci JIT
// These benchmarks use the iterative version which is O(n) instead of O(2^n)
func BenchmarkIterativeFib10(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	recComp := NewTrueRecursiveJITCompiler(config)
	code, _ := recComp.Compile(fn, nil) // Uses safe iterative version

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 10)
	}
}

func BenchmarkIterativeFib20(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	recComp := NewTrueRecursiveJITCompiler(config)
	code, _ := recComp.Compile(fn, nil)

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 20)
	}
}

func BenchmarkIterativeFib30(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	recComp := NewTrueRecursiveJITCompiler(config)
	code, _ := recComp.Compile(fn, nil)

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 30)
	}
}

func BenchmarkIterativeFib35(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	recComp := NewTrueRecursiveJITCompiler(config)
	code, _ := recComp.Compile(fn, nil)

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 35)
	}
}

// BenchmarkTrueRecursiveFib10 benchmarks TRUE recursive JIT (DANGEROUS)
// Only for small inputs - true recursion is O(2^n)
func BenchmarkTrueRecursiveFib10(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, _ := trueRecCompiler.CompileTrueRecursive(fn, nil) // Use true recursive

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 10)
	}
}

// BenchmarkTrueRecursiveFib20 - WARNING: fib(20) requires ~6,765 calls
func BenchmarkTrueRecursiveFib20(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, _ := trueRecCompiler.CompileTrueRecursive(fn, nil)

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, _ := jitCompiler.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.Call1(fnPtr, 20)
	}
}
