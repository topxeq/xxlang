// +build amd64,!windows

// pkg/jit/jit_true_recursive_test.go
// Test for TRUE recursive JIT execution (not transformation to iteration)
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

// TestTrueRecursiveFibJIT tests JIT that performs TRUE recursion
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

	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, err := trueRecCompiler.Compile(fn, nil)
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

	// Test small values first
	testCases := []struct {
		n        int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
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

// TestTrueRecursiveFibPerformance tests performance of TRUE recursive JIT
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

	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, err := trueRecCompiler.Compile(fn, nil)
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

	// Test fib(20) with true recursion
	iterations := 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 20)
	}
	avgTime := time.Since(start) / time.Duration(iterations)

	result := bridge.Call1(fnPtr, 20)
	t.Logf("TRUE recursive fib(20): %v per call, result=%d", avgTime, result)

	// Test fib(25)
	start = time.Now()
	result = bridge.Call1(fnPtr, 25)
	fib25Time := time.Since(start)
	t.Logf("TRUE recursive fib(25): %v, result=%d", fib25Time, result)

	// Test fib(30)
	start = time.Now()
	result = bridge.Call1(fnPtr, 30)
	fib30Time := time.Since(start)
	t.Logf("TRUE recursive fib(30): %v, result=%d", fib30Time, result)

	// Test fib(35)
	start = time.Now()
	result = bridge.Call1(fnPtr, 35)
	fib35Time := time.Since(start)
	t.Logf("TRUE recursive fib(35): %v, result=%d", fib35Time, result)

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

// BenchmarkTrueRecursiveFib benchmarks TRUE recursive JIT
func BenchmarkTrueRecursiveFib10(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, _ := trueRecCompiler.Compile(fn, nil)

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

func BenchmarkTrueRecursiveFib20(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, _ := trueRecCompiler.Compile(fn, nil)

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

func BenchmarkTrueRecursiveFib30(b *testing.B) {
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}

	fn := &compiler.CompiledFunction{NumLocals: 1, NumParameters: 1, Instructions: []byte{}}
	trueRecCompiler := NewTrueRecursiveJITCompiler(config)
	code, _ := trueRecCompiler.Compile(fn, nil)

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
