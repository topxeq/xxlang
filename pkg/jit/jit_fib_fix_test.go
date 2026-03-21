// Fixed JIT Fibonacci test
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/jit/bridge"
)

// TestIterativeFibJITFixed tests iterative Fibonacci with correct jump offsets
func TestIterativeFibJITFixed(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Iterative Fibonacci using only registers
	// Input: rdi = n
	// Output: rax = fib(n)
	//
	// Register allocation:
	//   rdi = n (input)
	//   rcx = a (fib(i-2))
	//   rdx = b (fib(i-1))
	//   r8  = i (counter, starts at 2)
	//   rax = temp/result
	//
	// Algorithm:
	//   if n <= 1: return n
	//   a = 0, b = 1, i = 2
	//   while i <= n:
	//     temp = a + b
	//     a = b
	//     b = temp
	//     i++
	//   return b

	// Build code with correct byte counting
	var code []byte

	emit := func(b ...byte) {
		code = append(code, b...)
	}

	// Base case check
	emit(0x48, 0x89, 0xF8)                   // mov rax, rdi
	emit(0x48, 0x83, 0xF8, 0x01)             // cmp rax, 1
	jleOffset := len(code)                   // Record jle position
	emit(0x7E, 0x00)                         // jle (placeholder)

	// Initialize
	emit(0x48, 0x31, 0xC9)                   // xor rcx, rcx (a = 0)
	emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00) // mov rdx, 1 (b = 1)
	emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00) // mov r8, 2 (i = 2)

	// Loop start
	loopStart := len(code)                   // Record loop start position
	emit(0x48, 0x89, 0xC8)                   // mov rax, rcx (temp = a)
	emit(0x48, 0x01, 0xD0)                   // add rax, rdx (temp += b)
	emit(0x48, 0x89, 0xD1)                   // mov rcx, rdx (a = b)
	emit(0x48, 0x89, 0xC2)                   // mov rdx, rax (b = temp)
	emit(0x49, 0xFF, 0xC0)                   // inc r8 (i++)
	emit(0x4C, 0x39, 0xC7)                   // cmp rdi, r8 (compare n with i)
	jgeOffset := len(code)                   // Record jge position
	emit(0x7D, 0x00)                         // jge (placeholder)

	// Done - return b (for n > 1)
	emit(0x48, 0x89, 0xD0)                   // mov rax, rdx
	emit(0xC3)                               // ret

	// Base case - return n (already in rax) for n <= 1
	baseCaseOffset := len(code)
	emit(0xC3)                               // ret

	// Fix up jumps using actual calculated positions
	// jle: jump to base case (n <= 1)
	jleRel := int8(baseCaseOffset - (jleOffset + 2))
	t.Logf("jle at %d: target=%d, from=%d, rel=%d", jleOffset, baseCaseOffset, jleOffset+2, jleRel)
	code[jleOffset+1] = byte(jleRel)

	// jge: jump back to loop start (continue loop)
	jgeRel := int8(loopStart - (jgeOffset + 2))
	t.Logf("jge at %d: target=%d, from=%d, rel=%d", jgeOffset, loopStart, jgeOffset+2, jgeRel)
	code[jgeOffset+1] = byte(jgeRel)

	t.Logf("Code length: %d bytes", len(code))
	t.Logf("Code: %x", code)

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test cases
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

// TestSimpleLoop tests a simple counting loop
func TestSimpleLoop(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Simple loop: sum 1 to n
	// Input: rdi = n
	// Output: rax = sum
	//
	// rax = 0
	// rcx = 1
	// while rcx <= rdi:
	//   rax += rcx
	//   rcx++
	// ret

	var code []byte
	
	emit := func(b ...byte) {
		code = append(code, b...)
	}

	// rax = 0
	emit(0x48, 0x31, 0xC0)                   // xor rax, rax
	// rcx = 1
	emit(0x48, 0xC7, 0xC1, 0x01, 0x00, 0x00, 0x00) // mov rcx, 1

	// Loop start
	loopStart := len(code)
	// cmp rcx, rdi
	emit(0x48, 0x39, 0xF9)                   // cmp rcx, rdi
	// jg done
	jgOffset := len(code)
	emit(0x7F, 0x00)                         // jg (placeholder)
	// rax += rcx
	emit(0x48, 0x01, 0xC8)                   // add rax, rcx
	// rcx++
	emit(0x48, 0xFF, 0xC1)                   // inc rcx
	// jmp loop
	jmpOffset := len(code)
	emit(0xEB, 0x00)                         // jmp (placeholder)
	// Done
	doneOffset := len(code)
	emit(0xC3)                               // ret

	// Fix up jumps
	// jg: jump to doneOffset
	code[jgOffset+1] = byte(doneOffset - (jgOffset + 2))
	// jmp: jump back to loopStart
	code[jmpOffset+1] = byte(loopStart - (jmpOffset + 2))

	t.Logf("Code: %x", code)

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// sum(1..10) = 55
	result := bridge.Call1(fnPtr, 10)
	t.Logf("sum(1..10) = %d (expected 55)", result)
	if result != 55 {
		t.Errorf("Expected 55, got %d", result)
	}

	// sum(1..100) = 5050
	result = bridge.Call1(fnPtr, 100)
	t.Logf("sum(1..100) = %d (expected 5050)", result)
	if result != 5050 {
		t.Errorf("Expected 5050, got %d", result)
	}
}

// TestJITPerformanceSimple tests simple JIT performance
func TestJITPerformanceSimple(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	// Simple loop: sum 1 to n
	var code []byte
	emit := func(b ...byte) { code = append(code, b...) }

	emit(0x48, 0x31, 0xC0)                   // xor rax, rax
	emit(0x48, 0xC7, 0xC1, 0x01, 0x00, 0x00, 0x00) // mov rcx, 1
	loopStart := len(code)
	emit(0x48, 0x39, 0xF9)                   // cmp rcx, rdi
	jgOffset := len(code)
	emit(0x7F, 0x00)                         // jg
	emit(0x48, 0x01, 0xC8)                   // add rax, rcx
	emit(0x48, 0xFF, 0xC1)                   // inc rcx
	jmpOffset := len(code)
	emit(0xEB, 0x00)                         // jmp
	doneOffset := len(code)
	emit(0xC3)                               // ret

	code[jgOffset+1] = byte(doneOffset - (jgOffset + 2))
	code[jmpOffset+1] = byte(loopStart - (jmpOffset + 2))

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, _ := jit.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Warm up
	for i := 0; i < 1000; i++ {
		bridge.Call1(fnPtr, 100)
	}

	// Benchmark
	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 100)
	}
	elapsed := time.Since(start)

	avgNs := elapsed.Nanoseconds() / int64(iterations)
	opsPerSec := float64(iterations) / elapsed.Seconds()

	t.Logf("JIT sum(1..100): %d iterations in %v", iterations, elapsed)
	t.Logf("Average: %d ns/op, %.0f ops/sec", avgNs, opsPerSec)
}
