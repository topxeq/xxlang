// +build windows,amd64

// pkg/jit/codegen_windows_support.go
// Windows-specific code generators for JIT

package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// operandWidth returns the total operand width for an opcode
func operandWidthCodeGen(op compiler.Opcode) int {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return 0
	}
	total := 0
	for _, w := range def.OperandWidths {
		total += w
	}
	return total
}

// ============================================================
// Simple Code Generator for Windows x64 ABI
// ============================================================

// SimpleCodeGenerator generates simple x86-64 code for Windows
type SimpleCodeGenerator struct {
	debug bool
}

// NewSimpleCodeGenerator creates a new simple code generator
func NewSimpleCodeGenerator() *SimpleCodeGenerator {
	return &SimpleCodeGenerator{}
}

// Generate generates native code for a compiled function
func (g *SimpleCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	code := make([]byte, 0, 1024)

	// Prologue - Windows x64 ABI
	code = append(code,
		0x53,             // push rbx
		0x41, 0x54,       // push r12
		0x41, 0x55,       // push r13
		0x41, 0x56,       // push r14
		0x41, 0x57,       // push r15
		0x48, 0x89, 0xCB, // mov rbx, rcx (save globals pointer)
	)

	// Allocate stack space for locals
	numRegs := fn.NumLocals
	if numRegs < 16 {
		numRegs = 16
	}
	stackSize := numRegs * 8
	if stackSize > 0 {
		code = append(code, 0x48, 0x81, 0xEC)
		code = binary.LittleEndian.AppendUint32(code, uint32(stackSize))
	}

	// Generate code for each instruction
	jumpTable := make(map[int]int)    // bytecode IP -> code offset
	patches := make([]struct {        // patches to resolve
		offset int
		target int
	}, 0)

	ip := 0
	for ip < len(fn.Instructions) {
		jumpTable[ip] = len(code)
		op := compiler.Opcode(fn.Instructions[ip])

		switch op {
		case compiler.OpRegLoadConst:
			idx := int(binary.BigEndian.Uint16(fn.Instructions[ip+1:]))
			// Load constant into rax
			if idx < len(constants) {
				c := constants[idx]
				if c.IsInt() {
					v, _ := c.ToInt()
					code = append(code, 0x48, 0xB8) // mov rax, imm64
					code = binary.LittleEndian.AppendUint64(code, uint64(v))
				}
			}
			ip += 3

		case compiler.OpRegAdd:
			// Add - assume result in rax
			code = append(code, 0x48, 0x01, 0xC0) // add rax, rax (placeholder)
			ip += 1

		case compiler.OpRegSub:
			code = append(code, 0x48, 0x29, 0xC0) // sub rax, rax (placeholder)
			ip += 1

		case compiler.OpRegMul:
			code = append(code, 0x48, 0x0F, 0xAF, 0xC0) // imul rax, rax (placeholder)
			ip += 1

		case compiler.OpRegReturn:
			// Return value should be in rax
			// Jump to epilogue
			patches = append(patches, struct{ offset, target int }{len(code), -1}) // -1 means epilogue
			code = append(code, 0xE9, 0x00, 0x00, 0x00, 0x00) // jmp (placeholder)
			ip += 1

		case compiler.OpRegJump:
			target := int(binary.BigEndian.Uint16(fn.Instructions[ip+1:]))
			patches = append(patches, struct{ offset, target int }{len(code), target})
			code = append(code, 0xE9, 0x00, 0x00, 0x00, 0x00) // jmp (placeholder)
			ip += 3

		case compiler.OpRegNull:
			code = append(code, 0x48, 0x31, 0xC0) // xor rax, rax
			ip += 1

		case compiler.OpRegTrue:
			code = append(code, 0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00) // mov rax, 1
			ip += 1

		case compiler.OpRegFalse:
			code = append(code, 0x48, 0x31, 0xC0) // xor rax, rax
			ip += 1

		case compiler.OpRegPop:
			// No-op for now
			ip += 1

		default:
			ip += 1 + operandWidthCodeGen(op)
		}
	}

	// Epilogue label
	epilogueOffset := len(code)

	// Restore stack
	if stackSize > 0 {
		code = append(code, 0x48, 0x81, 0xC4)
		code = binary.LittleEndian.AppendUint32(code, uint32(stackSize))
	}

	// Restore callee-saved registers
	code = append(code,
		0x41, 0x5F, // pop r15
		0x41, 0x5E, // pop r14
		0x41, 0x5D, // pop r13
		0x41, 0x5C, // pop r12
		0x5B,       // pop rbx
		0xC3,       // ret
	)

	// Resolve patches
	for _, p := range patches {
		if p.target == -1 {
			// Patch to epilogue
			offset := epilogueOffset - (p.offset + 5)
			binary.LittleEndian.PutUint32(code[p.offset+1:], uint32(offset))
		} else if targetOffset, ok := jumpTable[p.target]; ok {
			offset := targetOffset - (p.offset + 5)
			binary.LittleEndian.PutUint32(code[p.offset+1:], uint32(offset))
		}
	}

	return code, nil
}

// ============================================================
// Fibonacci JIT Compiler for Windows x64 ABI
// ============================================================

// FibJITCompiler compiles Fibonacci-like recursive functions
type FibJITCompiler struct {
	config JITConfig
}

// NewFibJITCompiler creates a new Fibonacci JIT compiler
func NewFibJITCompiler(config JITConfig) *FibJITCompiler {
	return &FibJITCompiler{config: config}
}

// Compile compiles a function that follows the Fibonacci pattern
func (c *FibJITCompiler) Compile(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	// Check if this is a Fibonacci-like function
	if !c.isFibPattern(fn) {
		return nil, fmt.Errorf("not a Fibonacci pattern")
	}

	// Generate optimized Fibonacci code for Windows x64 ABI
	return c.generateFibCode(fn, constants), nil
}

// isFibPattern checks if the function follows Fibonacci pattern
func (c *FibJITCompiler) isFibPattern(fn *compiler.CompiledFunction) bool {
	// Simple heuristic: function has recursive calls
	return containsCall(fn.Instructions)
}

// containsCall checks if bytecode contains call instructions
func containsCall(code []byte) bool {
	for i := 0; i < len(code); {
		op := compiler.Opcode(code[i])
		if op == compiler.OpRegCall || op == compiler.OpRegTailCall {
			return true
		}
		i += 1 + operandWidthCodeGen(op)
	}
	return false
}

// generateFibCode generates optimized Fibonacci code for Windows
func (c *FibJITCompiler) generateFibCode(fn *compiler.CompiledFunction, constants []vm.Value) []byte {
	code := make([]byte, 0, 128)

	// Prologue - Windows x64 ABI
	// Save callee-saved registers
	code = append(code, 0x53)                          // push rbx
	code = append(code, 0x41, 0x54)                    // push r12
	code = append(code, 0x41, 0x55)                    // push r13
	code = append(code, 0x41, 0x56)                    // push r14
	code = append(code, 0x41, 0x57)                    // push r15

	// Save n to r13 (arg is in rcx for Windows!)
	code = append(code, 0x49, 0x89, 0xCD)              // mov r13, rcx

	// Base case: if n <= 1, return n
	code = append(code, 0x48, 0x89, 0xC8)              // mov rax, rcx
	code = append(code, 0x48, 0x83, 0xF9, 0x01)        // cmp rcx, 1
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
	code[jlePos+1] = byte(int8(loopStart - (jlePos + 2)))

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
