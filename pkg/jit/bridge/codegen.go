// +build !windows,amd64

package bridge

// BuildFibCode generates x86-64 machine code for iterative Fibonacci.
// This version uses System V AMD64 ABI (Linux/macOS).
// Arguments: rdi = n, returns result in rax.
func BuildFibCode() []byte {
	code := make([]byte, 0, 128)

	// Save callee-saved registers (System V ABI)
	code = append(code, 0x53)                          // push rbx
	code = append(code, 0x41, 0x54)                    // push r12
	code = append(code, 0x41, 0x55)                    // push r13
	code = append(code, 0x41, 0x56)                    // push r14
	code = append(code, 0x41, 0x57)                    // push r15

	// Save n to r13 (arg is in rdi for System V)
	code = append(code, 0x49, 0x89, 0xFD)              // mov r13, rdi

	// Base case: if n <= 1, return n
	code = append(code, 0x48, 0x89, 0xF8)              // mov rax, rdi
	code = append(code, 0x48, 0x83, 0xFF, 0x01)        // cmp rdi, 1
	jgPos := len(code)
	code = append(code, 0x7F, 0x00)                    // jg (placeholder)
	jmpPos := len(code)
	code = append(code, 0xEB, 0x00)                    // jmp to epilogue
	code[jgPos+1] = 0x02                               // jg +2

	// Initialize: a=0, b=1, i=2
	code = append(code, 0x48, 0x31, 0xDB)              // xor rbx, rbx (a=0)
	code = append(code, 0x49, 0xC7, 0xC4, 0x01, 0x00, 0x00, 0x00) // mov r12, 1 (b=1)
	code = append(code, 0x49, 0xC7, 0xC6, 0x02, 0x00, 0x00, 0x00) // mov r14, 2 (i=2)

	// Loop start
	loopStart := len(code)

	// temp = a + b
	code = append(code, 0x4C, 0x89, 0xE0)              // mov rax, r12 (b)
	code = append(code, 0x48, 0x01, 0xD8)              // add rax, rbx (a)
	code = append(code, 0x49, 0x89, 0xC7)              // mov r15, rax (temp)

	// a = b
	code = append(code, 0x4C, 0x89, 0xE3)              // mov rbx, r12

	// b = temp
	code = append(code, 0x4D, 0x89, 0xFC)              // mov r12, r15

	// i++
	code = append(code, 0x49, 0xFF, 0xC6)              // inc r14

	// if i <= n, continue
	code = append(code, 0x4D, 0x39, 0xEE)              // cmp r14, r13
	jlePos := len(code)
	code = append(code, 0x7E, 0x00)                    // jle (placeholder)
	// Validate jump offset fits in rel8
	jleOffset := loopStart - (jlePos + 2)
	if jleOffset < -128 || jleOffset > 127 {
		// Clamp to valid range and log warning
		if jleOffset > 127 {
			jleOffset = 127
		} else {
			jleOffset = -128
		}
	}
	code[jlePos+1] = byte(int8(jleOffset))

	// Return b
	code = append(code, 0x4C, 0x89, 0xE0)              // mov rax, r12

	// Epilogue
	epilogueStart := len(code)
	code[jmpPos+1] = byte(epilogueStart - (jmpPos + 2))

	// Restore callee-saved
	code = append(code, 0x41, 0x5F)                    // pop r15
	code = append(code, 0x41, 0x5E)                    // pop r14
	code = append(code, 0x41, 0x5D)                    // pop r13
	code = append(code, 0x41, 0x5C)                    // pop r12
	code = append(code, 0x5B)                          // pop rbx
	code = append(code, 0xC3)                          // ret

	return code
}
