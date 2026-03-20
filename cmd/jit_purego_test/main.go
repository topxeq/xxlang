// Pure Go JIT - without CGO
// Uses Go assembly bridge to call JIT code

package main

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

func main() {
	fmt.Println("=== Pure Go JIT Test (No CGO) ===\n")

	// Test: Can we use syscall to call JIT code?
	// On Linux, we can use syscall.Syscall to call function pointers

	// First, let's verify executable memory works
	fmt.Println("Test 1: Allocate and write to executable memory")

	// Allocate executable memory
	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	flags := syscall.MAP_ANONYMOUS | syscall.MAP_PRIVATE

	code := []byte{
		0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00, // mov rax, 42
		0xC3, // ret
	}

	mem, err := syscall.Mmap(-1, 0, 4096, prot, flags)
	if err != nil {
		fmt.Printf("Mmap failed: %v\n", err)
		return
	}
	defer syscall.Munmap(mem)

	copy(mem, code)
	fmt.Printf("Allocated executable memory at %p\n", &mem[0])
	fmt.Printf("Code: %s\n", hex.Dump(code))

	// Method 1: Try using syscall.Syscall (won't work for arbitrary functions)
	// syscall.Syscall is for actual syscalls, not arbitrary function calls

	// Method 2: Use a Go assembly bridge
	// We need a small assembly function that can call a function pointer

	// Method 3: The hacky way - use reflect.MakeFunc with unsafe pointer conversion
	// This might work but is dangerous

	fmt.Println("\n=== Using Go Assembly Bridge ===")

	// Call via assembly bridge (defined in jit_bridge_amd64.s)
	result := callJitCode0(&mem[0])
	fmt.Printf("JIT result (via asm bridge): %d (expected 42)\n", result)

	if result == 42 {
		fmt.Println("✓ Pure Go JIT works!")
	} else {
		fmt.Println("✗ Failed")
	}

	// Test Fibonacci
	fmt.Println("\n=== Fibonacci JIT ===")
	fibCode := buildFibCode()
	fmt.Printf("Fibonacci code (%d bytes):\n%s\n", len(fibCode), hex.Dump(fibCode))

	fibMem, err := syscall.Mmap(-1, 0, 4096, prot, flags)
	if err != nil {
		fmt.Printf("Mmap failed: %v\n", err)
		return
	}
	defer syscall.Munmap(fibMem)
	copy(fibMem, fibCode)

	// Test
	for n := int64(0); n <= 10; n++ {
		r := callJitCode1(&fibMem[0], n)
		expected := fibReference(n)
		status := "✓"
		if r != expected {
			status = "✗"
		}
		fmt.Printf("fib(%d) = %d (expected %d) %s\n", n, r, expected, status)
	}

	// Benchmark
	fmt.Println("\n=== Performance Benchmark ===")

	// JIT benchmark
	jitIterations := 10000000
	start := time.Now()
	for i := 0; i < jitIterations; i++ {
		callJitCode1(&fibMem[0], 30)
	}
	jitTime := time.Since(start) / time.Duration(jitIterations)

	// Native Go benchmark
	goIterations := 100000000
	start = time.Now()
	for i := 0; i < goIterations; i++ {
		fibReference(30)
	}
	goTime := time.Since(start) / time.Duration(goIterations)

	fmt.Printf("Native Go:  %v per iteration\n", goTime)
	fmt.Printf("JIT (asm):  %v per iteration\n", jitTime)
	fmt.Printf("Overhead:   %.1fx slower than native\n", float64(jitTime)/float64(goTime))

	// Compare with Xxlang interpreter
	fmt.Println("\n=== Full Comparison with Xxlang Interpreter ===")

	// Use the Xxlang interpreter for comparison
	interpreterCode := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
		func fib(n) {
			return fibHelper(n, 0, 1)
		}
		fib(30)
	`

	l := lexer.New(interpreterCode)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Interpreter benchmark
	interpreterIterations := 5000
	start = time.Now()
	for i := 0; i < interpreterIterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterTime := time.Since(start) / time.Duration(interpreterIterations)

	vmInst := vm.NewRegVM(bytecode)
	vmInst.Run()
	vmResult := vmInst.LastPoppedObject()
	fmt.Printf("\nInterpreter result: fib(30) = %s\n", vmResult.Inspect())
	fmt.Printf("Interpreter (TCO): %v per iteration\n", interpreterTime)

	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Pure Go JIT Performance Summary                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Printf("  Native Go:     %v\n", goTime)
	fmt.Printf("  JIT (asm):     %v\n", jitTime)
	fmt.Printf("  Interpreter:   %v\n", interpreterTime)
	fmt.Printf("\nSpeed improvements:\n")
	fmt.Printf("  JIT vs Interpreter: %.0fx faster\n", float64(interpreterTime)/float64(jitTime))
	fmt.Printf("  JIT vs Native Go:   %.1fx slower\n", float64(jitTime)/float64(goTime))
	fmt.Printf("  Interpreter vs Native: %.0fx slower\n", float64(interpreterTime)/float64(goTime))
}

// Reference Fibonacci implementation
func fibReference(n int64) int64 {
	if n <= 1 {
		return n
	}
	var a, b int64 = 0, 1
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// Build Fibonacci machine code
// System V AMD64 ABI: arg in rdi, result in rax
func buildFibCode() []byte {
	code := make([]byte, 0, 128)

	// Save callee-saved registers
	code = append(code, 0x53)                          // push rbx
	code = append(code, 0x41, 0x54)                    // push r12
	code = append(code, 0x41, 0x55)                    // push r13
	code = append(code, 0x41, 0x56)                    // push r14
	code = append(code, 0x41, 0x57)                    // push r15

	// Save n to r13
	code = append(code, 0x49, 0x89, 0xFD)              // mov r13, rdi

	// Base case check: if n <= 1, return n
	code = append(code, 0x48, 0x89, 0xF8)              // mov rax, rdi
	code = append(code, 0x48, 0x83, 0xFF, 0x01)        // cmp rdi, 1
	jgPos := len(code)
	code = append(code, 0x7F, 0x00)                    // jg (placeholder)

	jmpPos := len(code)
	code = append(code, 0xEB, 0x00)                    // jmp to epilogue (placeholder)

	// Fix jg to skip over jmp
	code[jgPos+1] = 0x02                               // jg +2

	// Initialize: a = 0, b = 1, i = 2
	code = append(code, 0x48, 0x31, 0xDB)              // xor rbx, rbx (a = 0)
	code = append(code, 0x49, 0xC7, 0xC4, 0x01, 0x00, 0x00, 0x00) // mov r12, 1 (b = 1)
	code = append(code, 0x49, 0xC7, 0xC6, 0x02, 0x00, 0x00, 0x00) // mov r14, 2 (i = 2)

	// Loop start
	loopStart := len(code)

	// temp = a + b => r15 = rbx + r12
	code = append(code, 0x4C, 0x89, 0xE0)              // mov rax, r12 (b)
	code = append(code, 0x48, 0x01, 0xD8)              // add rax, rbx (a)
	code = append(code, 0x49, 0x89, 0xC7)              // mov r15, rax (temp)

	// a = b => rbx = r12
	code = append(code, 0x4C, 0x89, 0xE3)              // mov rbx, r12

	// b = temp => r12 = r15
	code = append(code, 0x4D, 0x89, 0xFC)              // mov r12, r15

	// i++ => inc r14
	code = append(code, 0x49, 0xFF, 0xC6)              // inc r14

	// if i <= n, continue loop
	code = append(code, 0x4D, 0x39, 0xEE)              // cmp r14, r13
	jlePos := len(code)
	code = append(code, 0x7E, 0x00)                    // jle (placeholder)

	// Fix backward jump
	loopOffset := int8(loopStart - (jlePos + 2))
	code[jlePos+1] = byte(loopOffset)

	// Return b in rax
	code = append(code, 0x4C, 0x89, 0xE0)              // mov rax, r12

	// Epilogue
	epilogueStart := len(code)
	code[jmpPos+1] = byte(epilogueStart - (jmpPos + 2))

	// Restore callee-saved registers
	code = append(code, 0x41, 0x5F)                    // pop r15
	code = append(code, 0x41, 0x5E)                    // pop r14
	code = append(code, 0x41, 0x5D)                    // pop r13
	code = append(code, 0x41, 0x5C)                    // pop r12
	code = append(code, 0x5B)                          // pop rbx
	code = append(code, 0xC3)                          // ret

	return code
}
