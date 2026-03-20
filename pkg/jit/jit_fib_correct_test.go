// Correctly calculated JIT Fibonacci test
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/jit/bridge"
)

// TestIterativeFibCorrect tests with correctly calculated offsets
func TestIterativeFibCorrect(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Build code with correct offset tracking
	var code []byte
	
	emit := func(b ...byte) int {
		start := len(code)
		code = append(code, b...)
		return start
	}

	// Check base case: if n <= 1, return n
	_ = emit(0x48, 0x89, 0xF8)                   // mov rax, rdi
	_ = emit(0x48, 0x83, 0xF8, 0x01)             // cmp rax, 1
	jlePos := emit(0x7E, 0x00)                   // jle (placeholder)

	// Initialize: a=0, b=1, i=2
	_ = emit(0x48, 0x31, 0xC9)                   // xor rcx, rcx (a = 0)
	_ = emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00) // mov rdx, 1 (b = 1)
	_ = emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00) // mov r8, 2 (i = 2)

	// Loop start
	loopStart := emit(0x48, 0x89, 0xC8)          // mov rax, rcx (temp = a)
	_ = emit(0x48, 0x01, 0xD0)                   // add rax, rdx (temp += b)
	_ = emit(0x48, 0x89, 0xD1)                   // mov rcx, rdx (a = b)
	_ = emit(0x48, 0x89, 0xC2)                   // mov rdx, rax (b = temp)
	_ = emit(0x49, 0xFF, 0xC0)                   // inc r8 (i++)
	_ = emit(0x4C, 0x39, 0xC7)                   // cmp rdi, r8 (n - i)
	jgePos := emit(0x7D, 0x00)                   // jge (placeholder)

	// Done: return b
	donePos := emit(0x48, 0x89, 0xD0)            // mov rax, rdx
	_ = emit(0xC3)                               // ret

	// Now fix up jumps with correct calculations
	// jle: jumps from jlePos to donePos
	// jle is 2 bytes, so relative offset = donePos - (jlePos + 2)
	jleRel := donePos - (jlePos + 2)
	t.Logf("jle: pos=%d, target=%d, rel=%d", jlePos, donePos, jleRel)
	code[jlePos+1] = byte(jleRel)

	// jge: jumps from jgePos back to loopStart
	// jge is 2 bytes, so relative offset = loopStart - (jgePos + 2)
	jgeRel := loopStart - (jgePos + 2)
	t.Logf("jge: pos=%d, target=%d, rel=%d", jgePos, loopStart, jgeRel)
	code[jgePos+1] = byte(jgeRel)

	t.Logf("Code length: %d bytes", len(code))
	t.Logf("Code: %x", code)

	// Disassemble for verification
	t.Logf("\nDisassembly:")
	ip := 0
	for ip < len(code) {
		oldIP := ip
		b := code[ip]
		switch b {
		case 0x48:
			if ip+1 < len(code) {
				switch code[ip+1] {
				case 0x89:
					t.Logf("%2d: mov r64, r64", oldIP)
					ip += 3
				case 0x83:
					t.Logf("%2d: cmp r64, imm8", oldIP)
					ip += 4
				case 0x31:
					t.Logf("%2d: xor r64, r64", oldIP)
					ip += 3
				case 0xC7:
					t.Logf("%2d: mov r64, imm32", oldIP)
					ip += 7
				case 0x01:
					t.Logf("%2d: add r64, r64", oldIP)
					ip += 3
				case 0xFF:
					t.Logf("%2d: inc r64", oldIP)
					ip += 3
				default:
					t.Logf("%2d: unknown 48 xx", oldIP)
					ip += 2
				}
			} else {
				ip++
			}
		case 0x49:
			if ip+1 < len(code) {
				switch code[ip+1] {
				case 0xC7:
					t.Logf("%2d: mov r8, imm32", oldIP)
					ip += 7
				case 0xFF:
					t.Logf("%2d: inc r8", oldIP)
					ip += 3
				default:
					t.Logf("%2d: unknown 49 xx", oldIP)
					ip += 2
				}
			} else {
				ip++
			}
		case 0x4C:
			t.Logf("%2d: cmp rdi, r8", oldIP)
			ip += 3
		case 0x7E:
			rel := int(int8(code[ip+1]))
			target := ip + 2 + rel
			t.Logf("%2d: jle %d (rel=%d)", oldIP, target, rel)
			ip += 2
		case 0x7D:
			rel := int(int8(code[ip+1]))
			target := ip + 2 + rel
			t.Logf("%2d: jge %d (rel=%d)", oldIP, target, rel)
			ip += 2
		case 0xC3:
			t.Logf("%2d: ret", oldIP)
			ip++
		default:
			t.Logf("%2d: unknown byte %02x", oldIP, b)
			ip++
		}
	}

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

// TestFibJITPerformance_Native tests native JIT performance
func TestFibJITPerformance_Native(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	// Build Fibonacci code (same as above)
	var code []byte
	emit := func(b ...byte) int {
		start := len(code)
		code = append(code, b...)
		return start
	}

	_ = emit(0x48, 0x89, 0xF8)
	_ = emit(0x48, 0x83, 0xF8, 0x01)
	jlePos := emit(0x7E, 0x00)
	_ = emit(0x48, 0x31, 0xC9)
	_ = emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00)
	_ = emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00)
	loopStart := emit(0x48, 0x89, 0xC8)
	_ = emit(0x48, 0x01, 0xD0)
	_ = emit(0x48, 0x89, 0xD1)
	_ = emit(0x48, 0x89, 0xC2)
	_ = emit(0x49, 0xFF, 0xC0)
	_ = emit(0x4C, 0x39, 0xC7)
	jgePos := emit(0x7D, 0x00)
	donePos := emit(0x48, 0x89, 0xD0)
	_ = emit(0xC3)

	code[jlePos+1] = byte(donePos - (jlePos + 2))
	code[jgePos+1] = byte(loopStart - (jgePos + 2))

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, _ := jit.AllocCode(len(code))
	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Warm up
	for i := 0; i < 1000; i++ {
		bridge.Call1(fnPtr, 35)
	}

	// Benchmark JIT
	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 35)
	}
	jitTime := time.Since(start)

	t.Logf("JIT fib(35) x %d: %v (%.0f ops/sec)", 
		iterations, jitTime, float64(iterations)/jitTime.Seconds())
}
