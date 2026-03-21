// +build amd64,!windows

// pkg/jit/codegen_optimized.go
// Optimized x86-64 code generator that uses hardware registers for VM registers
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// OptimizedCodeGenerator generates more efficient x86-64 code by using
// hardware registers (rax, rbx, rcx, rdx, r8-r15) to cache VM registers
type OptimizedCodeGenerator struct {
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction

	// Register mapping: VM register -> x86 register (0-7)
	// We use: rax(0), rbx(1), rcx(2), rdx(3), r8(4), r9(5), r10(6), r11(7)
	// VM registers 0-7 are cached in hardware registers
	// VM registers 8+ are spilled to stack
	registerMap map[int]int

	// Track which hardware registers are dirty (need to be spilled)
	dirtyRegs uint8

	// Stack offset for spilled registers
	stackOffset int
}

// x86 register encodings for REX prefix
var x86Regs = []byte{
	0,  // rax
	3,  // rbx
	1,  // rcx
	2,  // rdx
	8,  // r8
	9,  // r9
	10, // r10
	11, // r11
}

// NewOptimizedCodeGenerator creates a new optimized code generator
func NewOptimizedCodeGenerator() *OptimizedCodeGenerator {
	return &OptimizedCodeGenerator{
		code:        make([]byte, 0, 4096),
		labels:      make(map[string]int),
		fixups:      make([]fixup, 0),
		registerMap: make(map[int]int),
	}
}

// Generate generates optimized machine code
func (cg *OptimizedCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn

	// Setup register mapping: first 8 VM registers map to hardware registers
	for i := 0; i < 8; i++ {
		cg.registerMap[i] = i
	}

	// Generate prologue
	cg.emitPrologue()

	// Compile instructions
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		cg.labels[fmt.Sprintf("L%d", ip)] = len(cg.code)

		if err := cg.compileInstruction(op, code, &ip); err != nil {
			return nil, err
		}
	}

	// Resolve fixups
	if err := cg.resolveFixups(); err != nil {
		return nil, err
	}

	return cg.code, nil
}

// emitPrologue generates function entry code
func (cg *OptimizedCodeGenerator) emitPrologue() {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// We'll use callee-saved registers: rbx, r12-r15
	// push rbx (VM reg 1)
	cg.emitByte(0x53)
	// push r12
	cg.emitBytes([]byte{0x41, 0x54})
	// push r13
	cg.emitBytes([]byte{0x41, 0x55})
	// push r14
	cg.emitBytes([]byte{0x41, 0x56})
	// push r15
	cg.emitBytes([]byte{0x41, 0x57})

	// Allocate stack space for spilled registers (VM regs 8+)
	// We allocate 512 bytes for safety
	cg.emitBytes([]byte{0x48, 0x81, 0xEC, 0x00, 0x02, 0x00, 0x00})
}

// emitEpilogue generates function exit code
func (cg *OptimizedCodeGenerator) emitEpilogue() {
	// pop r15
	cg.emitBytes([]byte{0x41, 0x5F})
	// pop r14
	cg.emitBytes([]byte{0x41, 0x5E})
	// pop r13
	cg.emitBytes([]byte{0x41, 0x5D})
	// pop r12
	cg.emitBytes([]byte{0x41, 0x5C})
	// pop rbx
	cg.emitByte(0x5B)
	// add rsp, 0x200
	cg.emitBytes([]byte{0x48, 0x81, 0xC4, 0x00, 0x02, 0x00, 0x00})
	// pop rbp
	cg.emitByte(0x5D)
	// ret
	cg.emitByte(0xC3)
}

// compileInstruction compiles a single instruction with optimization
func (cg *OptimizedCodeGenerator) compileInstruction(op compiler.Opcode, code []byte, ip *int) error {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return fmt.Errorf("unknown opcode %d", op)
	}

	switch op {
	case compiler.OpRegLoadConst:
		dst := int(code[*ip+1])
		constIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileLoadConst(dst, constIdx)
		*ip += 4

	case compiler.OpRegMove:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		cg.compileMoveOptimized(dst, src)
		*ip += 3

	case compiler.OpRegAdd:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileAddOptimized(dst, left, right)
		*ip += 4

	case compiler.OpRegSub:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileSubOptimized(dst, left, right)
		*ip += 4

	case compiler.OpRegMul:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileMulOptimized(dst, left, right)
		*ip += 4

	case compiler.OpRegLess:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileLessOptimized(dst, left, right)
		*ip += 4

	case compiler.OpRegEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileEqualOptimized(dst, left, right)
		*ip += 4

	case compiler.OpRegJump:
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		cg.compileJump(target)
		*ip += 4

	case compiler.OpRegJumpIfFalse:
		cond := int(code[*ip+1])
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		cg.compileJumpIfFalseOptimized(cond, target)
		*ip += 4

	case compiler.OpRegReturn:
		src := int(code[*ip+1])
		cg.compileReturnOptimized(src)
		*ip += 2

	case compiler.OpRegNull:
		dst := int(code[*ip+1])
		cg.compileNullOptimized(dst)
		*ip += 2

	case compiler.OpRegTrue:
		dst := int(code[*ip+1])
		cg.compileTrueOptimized(dst)
		*ip += 2

	case compiler.OpRegFalse:
		dst := int(code[*ip+1])
		cg.compileFalseOptimized(dst)
		*ip += 2

	case compiler.OpRegLoopCountAdd:
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		startIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		limitIdx := int(code[*ip+5])<<8 | int(code[*ip+6])
		stepIdx := int(code[*ip+7])<<8 | int(code[*ip+8])
		cg.compileLoopCountAddOptimized(accReg, counterReg, startIdx, limitIdx, stepIdx)
		*ip += 9

	default:
		// Fallback to simple code generator for unsupported opcodes
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// isRegisterCached returns true if VM register is cached in hardware register
func (cg *OptimizedCodeGenerator) isRegisterCached(vmReg int) bool {
	return vmReg < 8
}

// compileLoadConst loads a constant with optimization
func (cg *OptimizedCodeGenerator) compileLoadConst(dst, constIdx int) {
	var val int64
	if constIdx < len(cg.constants) {
		v := cg.constants[constIdx]
		if v.IsInt() {
			val, _ = v.ToInt()
		}
	}

	if cg.isRegisterCached(dst) {
		// Load directly into hardware register
		// mov reg, imm64
		switch dst {
		case 0: // rax
			cg.emitBytes([]byte{0x48, 0xB8})
		case 1: // rbx
			cg.emitBytes([]byte{0x48, 0xBB})
		case 2: // rcx
			cg.emitBytes([]byte{0x48, 0xB9})
		case 3: // rdx
			cg.emitBytes([]byte{0x48, 0xBA})
		default: // r8-r11
			cg.emitBytes([]byte{0x49, 0xB8 + byte(dst-4)})
		}
		cg.emitUint64(uint64(val))
	} else {
		// Spill to stack
		// mov rax, imm64
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(val))
		// mov [rbp - (dst+1)*8], rax
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileMoveOptimized moves a value with register caching
func (cg *OptimizedCodeGenerator) compileMoveOptimized(dst, src int) {
	if cg.isRegisterCached(dst) && cg.isRegisterCached(src) {
		// Both cached: mov dst_reg, src_reg
		// This is a simple register-to-register move
		cg.emitMovRegReg(dst, src)
	} else if cg.isRegisterCached(dst) {
		// Load from stack into hardware register
		cg.emitLoadFromStack(dst, src)
	} else if cg.isRegisterCached(src) {
		// Store hardware register to stack
		cg.emitStoreToStack(dst, src)
	} else {
		// Both spilled: load/store via rax
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((src + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileAddOptimized adds with register caching
func (cg *OptimizedCodeGenerator) compileAddOptimized(dst, left, right int) {
	// Move left to dst
	cg.compileMoveOptimized(dst, left)

	// Add right to dst
	if cg.isRegisterCached(dst) && cg.isRegisterCached(right) {
		// add dst_reg, right_reg
		cg.emitAddRegReg(dst, right)
	} else if cg.isRegisterCached(dst) {
		// add dst_reg, [stack]
		cg.emitAddRegStack(dst, right)
	} else {
		// Both spilled: load right, add, store
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x01, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileSubOptimized subtracts with register caching
func (cg *OptimizedCodeGenerator) compileSubOptimized(dst, left, right int) {
	cg.compileMoveOptimized(dst, left)
	if cg.isRegisterCached(dst) && cg.isRegisterCached(right) {
		cg.emitSubRegReg(dst, right)
	} else if cg.isRegisterCached(dst) {
		cg.emitBytes([]byte{0x48, 0x2B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x29, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileMulOptimized multiplies with register caching
func (cg *OptimizedCodeGenerator) compileMulOptimized(dst, left, right int) {
	cg.compileMoveOptimized(dst, left)
	if cg.isRegisterCached(dst) && cg.isRegisterCached(right) {
		cg.emitMulRegReg(dst, right)
	} else if cg.isRegisterCached(dst) {
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileLessOptimized compares for less than
func (cg *OptimizedCodeGenerator) compileLessOptimized(dst, left, right int) {
	// Compare left and right
	if cg.isRegisterCached(left) && cg.isRegisterCached(right) {
		cg.emitCmpRegReg(left, right)
	} else if cg.isRegisterCached(left) {
		cg.emitBytes([]byte{0x48, 0x3B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	} else if cg.isRegisterCached(right) {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((left + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x3B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((left + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x3B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	}

	// Set result
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})   // xor rax, rax
	cg.emitBytes([]byte{0x0F, 0x9C, 0xC0})   // setl al

	// Store result
	if cg.isRegisterCached(dst) {
		if dst != 0 { // rax
			cg.emitMovRaxToReg(dst)
		}
	} else {
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileEqualOptimized compares for equality
func (cg *OptimizedCodeGenerator) compileEqualOptimized(dst, left, right int) {
	if cg.isRegisterCached(left) && cg.isRegisterCached(right) {
		cg.emitCmpRegReg(left, right)
	} else if cg.isRegisterCached(left) {
		cg.emitBytes([]byte{0x48, 0x3B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((left + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x3B, 0x85})
		cg.emitUint32(uint32((right + 1) * 8))
	}

	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al

	if cg.isRegisterCached(dst) {
		if dst != 0 {
			cg.emitMovRaxToReg(dst)
		}
	} else {
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileJumpIfFalseOptimized jumps if condition is false
func (cg *OptimizedCodeGenerator) compileJumpIfFalseOptimized(cond, target int) {
	if cg.isRegisterCached(cond) {
		// test reg, reg
		switch cond {
		case 0:
			cg.emitBytes([]byte{0x48, 0x85, 0xC0})
		case 1:
			cg.emitBytes([]byte{0x48, 0x85, 0xDB})
		case 2:
			cg.emitBytes([]byte{0x48, 0x85, 0xC9})
		case 3:
			cg.emitBytes([]byte{0x48, 0x85, 0xD2})
		default:
			cg.emitBytes([]byte{0x4D, 0x85, 0xC0 + byte(cond-4)})
		}
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((cond + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	}

	// jz rel32
	cg.emitBytes([]byte{0x0F, 0x84})
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// compileReturnOptimized returns a value
func (cg *OptimizedCodeGenerator) compileReturnOptimized(src int) {
	if cg.isRegisterCached(src) {
		if src != 0 { // rax
			cg.emitMovRegToRax(src)
		}
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((src + 1) * 8))
	}
	cg.emitEpilogue()
}

// compileNullOptimized sets a register to null (0)
func (cg *OptimizedCodeGenerator) compileNullOptimized(dst int) {
	if cg.isRegisterCached(dst) {
		// xor reg, reg
		switch dst {
		case 0:
			cg.emitBytes([]byte{0x48, 0x31, 0xC0})
		case 1:
			cg.emitBytes([]byte{0x48, 0x31, 0xDB})
		case 2:
			cg.emitBytes([]byte{0x48, 0x31, 0xC9})
		case 3:
			cg.emitBytes([]byte{0x48, 0x31, 0xD2})
		default:
			cg.emitBytes([]byte{0x4D, 0x31, 0xC0 + byte(dst-4)})
		}
	} else {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0})
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
	}
}

// compileTrueOptimized sets a register to true (1)
func (cg *OptimizedCodeGenerator) compileTrueOptimized(dst int) {
	if cg.isRegisterCached(dst) {
		// mov reg, 1
		switch dst {
		case 0:
			cg.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00})
		case 1:
			cg.emitBytes([]byte{0x48, 0xC7, 0xC3, 0x01, 0x00, 0x00, 0x00})
		case 2:
			cg.emitBytes([]byte{0x48, 0xC7, 0xC1, 0x01, 0x00, 0x00, 0x00})
		case 3:
			cg.emitBytes([]byte{0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00})
		default:
			cg.emitBytes([]byte{0x49, 0xC7, 0xC0 + byte(dst-4), 0x01, 0x00, 0x00, 0x00})
		}
	} else {
		cg.emitBytes([]byte{0x48, 0xC7, 0x85})
		cg.emitUint32(uint32((dst + 1) * 8))
		cg.emitUint32(1)
	}
}

// compileFalseOptimized sets a register to false (0)
func (cg *OptimizedCodeGenerator) compileFalseOptimized(dst int) {
	cg.compileNullOptimized(dst)
}

// compileLoopCountAddOptimized compiles an optimized counting loop
func (cg *OptimizedCodeGenerator) compileLoopCountAddOptimized(accReg, counterReg, startIdx, limitIdx, stepIdx int) {
	// Get constant values
	var limit, step int64
	if limitIdx < len(cg.constants) {
		if v := cg.constants[limitIdx]; v.IsInt() {
			limit, _ = v.ToInt()
		}
	}
	if stepIdx < len(cg.constants) {
		if v := cg.constants[stepIdx]; v.IsInt() {
			step, _ = v.ToInt()
		}
	}
	if step == 0 {
		step = 1
	}

	// Initialize counter
	cg.compileLoadConst(counterReg, startIdx)

	// Initialize accumulator to 0
	cg.compileNullOptimized(accReg)

	// Loop label
	loopLabel := fmt.Sprintf("loop_%d", len(cg.code))
	cg.labels[loopLabel] = len(cg.code)

	// Compare counter with limit
	if cg.isRegisterCached(counterReg) {
		// cmp reg, limit
		switch counterReg {
		case 0:
			cg.emitBytes([]byte{0x48, 0x3D})
		case 1:
			cg.emitBytes([]byte{0x48, 0x81, 0xFB})
		case 2:
			cg.emitBytes([]byte{0x48, 0x81, 0xF9})
		case 3:
			cg.emitBytes([]byte{0x48, 0x81, 0xFA})
		default:
			cg.emitBytes([]byte{0x49, 0x81, 0xF8 + byte(counterReg-4)})
		}
		cg.emitUint32(uint32(limit))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32((counterReg + 1) * 8))
		cg.emitBytes([]byte{0x48, 0x3D})
		cg.emitUint32(uint32(limit))
	}

	// jge end
	endLabel := fmt.Sprintf("loop_end_%d", len(cg.code))
	cg.emitBytes([]byte{0x0F, 0x8D})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: endLabel, size: 4})
	cg.emitUint32(0)

	// Add counter to accumulator
	cg.compileAddOptimized(accReg, accReg, counterReg)

	// Increment counter
	if cg.isRegisterCached(counterReg) {
		// inc reg
		switch counterReg {
		case 0:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC0})
		case 1:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC3})
		case 2:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC1})
		case 3:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC2})
		default:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC0 + byte(counterReg-4)})
		}
	} else {
		cg.emitBytes([]byte{0x48, 0xFF, 0x85})
		cg.emitUint32(uint32((counterReg + 1) * 8))
	}

	// jmp loop
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: loopLabel, size: 4})
	cg.emitUint32(0)

	// End label
	cg.labels[endLabel] = len(cg.code)
}

// compileJump generates an unconditional jump
func (cg *OptimizedCodeGenerator) compileJump(target int) {
	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// ============================================================================
// Low-level emit functions
// ============================================================================

func (cg *OptimizedCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *OptimizedCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *OptimizedCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *OptimizedCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func (cg *OptimizedCodeGenerator) emitMovRegReg(dst, src int) {
	// mov dst, src using REX prefix
	if dst < 4 && src < 4 {
		// Simple case: mov r64, r64
		// opcode: 0x89 (mov r/m64, r64)
		cg.emitBytes([]byte{0x48, 0x89, 0xC0 + byte(src<<3) + byte(dst)})
	} else if dst < 4 {
		// dst is low reg, src is high reg
		cg.emitBytes([]byte{0x49, 0x89, 0xC0 + byte(src-4)<<3 + byte(dst)})
	} else if src < 4 {
		// dst is high reg, src is low reg
		cg.emitBytes([]byte{0x4C, 0x89, 0xC0 + byte(src)<<3 + byte(dst-4)})
	} else {
		// Both high regs
		cg.emitBytes([]byte{0x4D, 0x89, 0xC0 + byte(src-4)<<3 + byte(dst-4)})
	}
}

func (cg *OptimizedCodeGenerator) emitAddRegReg(dst, src int) {
	// add dst, src
	if dst < 4 && src < 4 {
		cg.emitBytes([]byte{0x48, 0x01, 0xC0 + byte(src<<3) + byte(dst)})
	} else if dst < 4 {
		cg.emitBytes([]byte{0x49, 0x01, 0xC0 + byte(src-4)<<3 + byte(dst)})
	} else if src < 4 {
		cg.emitBytes([]byte{0x4C, 0x01, 0xC0 + byte(src)<<3 + byte(dst-4)})
	} else {
		cg.emitBytes([]byte{0x4D, 0x01, 0xC0 + byte(src-4)<<3 + byte(dst-4)})
	}
}

func (cg *OptimizedCodeGenerator) emitSubRegReg(dst, src int) {
	if dst < 4 && src < 4 {
		cg.emitBytes([]byte{0x48, 0x29, 0xC0 + byte(src<<3) + byte(dst)})
	} else if dst < 4 {
		cg.emitBytes([]byte{0x49, 0x29, 0xC0 + byte(src-4)<<3 + byte(dst)})
	} else if src < 4 {
		cg.emitBytes([]byte{0x4C, 0x29, 0xC0 + byte(src)<<3 + byte(dst-4)})
	} else {
		cg.emitBytes([]byte{0x4D, 0x29, 0xC0 + byte(src-4)<<3 + byte(dst-4)})
	}
}

func (cg *OptimizedCodeGenerator) emitMulRegReg(dst, src int) {
	// imul dst, src
	if dst < 4 && src < 4 {
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC0 + byte(dst)<<3 + byte(src)})
	} else if dst < 4 {
		cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC0 + byte(dst)<<3 + byte(src-4)})
	} else if src < 4 {
		cg.emitBytes([]byte{0x4C, 0x0F, 0xAF, 0xC0 + byte(dst-4)<<3 + byte(src)})
	} else {
		cg.emitBytes([]byte{0x4D, 0x0F, 0xAF, 0xC0 + byte(dst-4)<<3 + byte(src-4)})
	}
}

func (cg *OptimizedCodeGenerator) emitCmpRegReg(left, right int) {
	// cmp left, right
	if left < 4 && right < 4 {
		cg.emitBytes([]byte{0x48, 0x39, 0xC0 + byte(right<<3) + byte(left)})
	} else if left < 4 {
		cg.emitBytes([]byte{0x49, 0x39, 0xC0 + byte(right-4)<<3 + byte(left)})
	} else if right < 4 {
		cg.emitBytes([]byte{0x4C, 0x39, 0xC0 + byte(right)<<3 + byte(left-4)})
	} else {
		cg.emitBytes([]byte{0x4D, 0x39, 0xC0 + byte(right-4)<<3 + byte(left-4)})
	}
}

func (cg *OptimizedCodeGenerator) emitMovRegToRax(src int) {
	switch src {
	case 1:
		cg.emitBytes([]byte{0x48, 0x89, 0xD8}) // mov rax, rbx
	case 2:
		cg.emitBytes([]byte{0x48, 0x89, 0xC8}) // mov rax, rcx
	case 3:
		cg.emitBytes([]byte{0x48, 0x89, 0xD0}) // mov rax, rdx
	default:
		cg.emitBytes([]byte{0x4C, 0x89, 0xC0 + byte(src-4)<<3}) // mov rax, rX
	}
}

func (cg *OptimizedCodeGenerator) emitMovRaxToReg(dst int) {
	switch dst {
	case 1:
		cg.emitBytes([]byte{0x48, 0x89, 0xC3}) // mov rbx, rax
	case 2:
		cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
	case 3:
		cg.emitBytes([]byte{0x48, 0x89, 0xC2}) // mov rdx, rax
	default:
		cg.emitBytes([]byte{0x49, 0x89, 0xC0 + byte(dst-4)}) // mov rX, rax
	}
}

func (cg *OptimizedCodeGenerator) emitLoadFromStack(dst, src int) {
	// mov dst, [rbp + offset]
	cg.emitBytes([]byte{0x48, 0x8B, 0x85})
	cg.emitUint32(uint32((src + 1) * 8))
	// Now move rax to dst
	if dst != 0 {
		cg.emitMovRaxToReg(dst)
	}
}

func (cg *OptimizedCodeGenerator) emitStoreToStack(dst, src int) {
	// Move src to rax first
	if src != 0 {
		cg.emitMovRegToRax(src)
	}
	// mov [rbp + offset], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x85})
	cg.emitUint32(uint32((dst + 1) * 8))
}

func (cg *OptimizedCodeGenerator) emitAddRegStack(dst, src int) {
	// add dst_reg, [rbp + offset]
	switch dst {
	case 0:
		cg.emitBytes([]byte{0x48, 0x03, 0x85})
	case 1:
		cg.emitBytes([]byte{0x48, 0x03, 0x9D})
	case 2:
		cg.emitBytes([]byte{0x48, 0x03, 0x8D})
	case 3:
		cg.emitBytes([]byte{0x48, 0x03, 0x95})
	default:
		cg.emitBytes([]byte{0x49, 0x03, 0x85})
	}
	cg.emitUint32(uint32((src + 1) * 8))
}

func (cg *OptimizedCodeGenerator) resolveFixups() error {
	for _, f := range cg.fixups {
		target, ok := cg.labels[f.label]
		if !ok {
			return fmt.Errorf("undefined label: %s", f.label)
		}
		offset := int32(target - (f.offset + f.size))
		binary.LittleEndian.PutUint32(cg.code[f.offset:], uint32(offset))
	}
	return nil
}
