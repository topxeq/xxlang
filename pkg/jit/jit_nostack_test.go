// No-stack JIT test - pure register-based code
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/jit/bridge"
)

// TestPureRegisterJIT tests JIT code that uses only registers (no stack)
func TestPureRegisterJIT(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Pure register code: return arg * 2
	code := []byte{
		0x48, 0x89, 0xF8, // mov rax, rdi
		0x48, 0x01, 0xC0, // add rax, rax
		0xC3, // ret
	}

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	result := bridge.Call1(fnPtr, 21)
	t.Logf("Call1: 21 * 2 = %d (expected 42)", result)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestIterativeFibNoStack tests iterative Fibonacci using only registers
func TestIterativeFibNoStack(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Build code dynamically with correct jump offsets
	var code []byte
	emit := func(b ...byte) { code = append(code, b...) }

	// Check base case: n <= 1
	emit(0x48, 0x89, 0xF8)             // mov rax, rdi
	emit(0x48, 0x83, 0xF8, 0x01)       // cmp rax, 1
	jleOffset := len(code)
	emit(0x7E, 0x00)                   // jle (placeholder)

	// Initialize: a=0, b=1, i=2
	emit(0x48, 0x31, 0xC9)             // xor rcx, rcx
	emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00) // mov rdx, 1
	emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00) // mov r8, 2

	// Loop start
	loopStart := len(code)
	emit(0x48, 0x89, 0xC8)             // mov rax, rcx
	emit(0x48, 0x01, 0xD0)             // add rax, rdx
	emit(0x48, 0x89, 0xD1)             // mov rcx, rdx
	emit(0x48, 0x89, 0xC2)             // mov rdx, rax
	emit(0x49, 0xFF, 0xC0)             // inc r8
	emit(0x4C, 0x39, 0xC7)             // cmp rdi, r8
	jgeOffset := len(code)
	emit(0x7D, 0x00)                   // jge (placeholder)

	// Done: return b
	doneOffset := len(code)
	emit(0x48, 0x89, 0xD0)             // mov rax, rdx
	emit(0xC3)                         // ret

	// Fix up jumps
	jleRel := int8(doneOffset - (jleOffset + 2))
	code[jleOffset+1] = byte(jleRel)

	jgeRel := int8(loopStart - (jgeOffset + 2))
	code[jgeOffset+1] = byte(jgeRel)

	t.Logf("Code length: %d bytes", len(code))
	t.Logf("Code: %x", code)
	t.Logf("jle at %d: target=%d, rel=%d", jleOffset, doneOffset, jleRel)
	t.Logf("jge at %d: target=%d, rel=%d", jgeOffset, loopStart, jgeRel)

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

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
		{35, 9227465},
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

// TestIterativeFibPerformance tests JIT performance
func TestIterativeFibPerformance(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	// Build code dynamically with correct jump offsets
	var code []byte
	emit := func(b ...byte) { code = append(code, b...) }

	emit(0x48, 0x89, 0xF8)
	emit(0x48, 0x83, 0xF8, 0x01)
	jleOffset := len(code)
	emit(0x7E, 0x00)
	emit(0x48, 0x31, 0xC9)
	emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00)
	emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00)
	loopStart := len(code)
	emit(0x48, 0x89, 0xC8)
	emit(0x48, 0x01, 0xD0)
	emit(0x48, 0x89, 0xD1)
	emit(0x48, 0x89, 0xC2)
	emit(0x49, 0xFF, 0xC0)
	emit(0x4C, 0x39, 0xC7)
	jgeOffset := len(code)
	emit(0x7D, 0x00)
	doneOffset := len(code)
	emit(0x48, 0x89, 0xD0)
	emit(0xC3)

	code[jleOffset+1] = byte(int8(doneOffset - (jleOffset + 2)))
	code[jgeOffset+1] = byte(int8(loopStart - (jgeOffset + 2)))

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Warm up
	for i := 0; i < 1000; i++ {
		bridge.Call1(fnPtr, 35)
	}

	// Benchmark
	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 35)
	}
	elapsed := time.Since(start)

	avgNs := elapsed.Nanoseconds() / int64(iterations)
	opsPerSec := float64(iterations) / elapsed.Seconds()

	t.Logf("JIT fib(35): %d iterations in %v", iterations, elapsed)
	t.Logf("Average: %d ns/op", avgNs)
	t.Logf("Throughput: %.0f ops/sec", opsPerSec)
}
