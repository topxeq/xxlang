// pkg/jit/native_codegen.go
// Pure native x86-64 code generator for JIT
// Generates self-contained native code without VM callbacks
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
)

// NativeCodeGenerator generates pure native x86-64 code
// No VM context needed - all values are in registers or stack
type NativeCodeGenerator struct {
	code     []byte
	labels   map[string]int
	fixups   []fixup
	constants []int64

	// Register allocation
	// We use: rax(0), rbx(1), rcx(2), rdx(3), r8(4), r9(5), r10(6), r11(7)
	// r12-r15 are callee-saved, used for locals 8-11
	// Stack is used for locals 12+
}

// NewNativeCodeGenerator creates a new native code generator
func NewNativeCodeGenerator() *NativeCodeGenerator {
	return &NativeCodeGenerator{
		code:     make([]byte, 0, 4096),
		labels:   make(map[string]int),
		fixups:   make([]fixup, 0),
		constants: nil,
	}
}

// Generate generates native x86-64 code
func (cg *NativeCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []int64) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants

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
func (cg *NativeCodeGenerator) emitPrologue() {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Save callee-saved registers (for locals 8-11)
	// push rbx
	cg.emitByte(0x53)
	// push r12
	cg.emitBytes([]byte{0x41, 0x54})
	// push r13
	cg.emitBytes([]byte{0x41, 0x55})
	// push r14
	cg.emitBytes([]byte{0x41, 0x56})
	// push r15
	cg.emitBytes([]byte{0x41, 0x57})

	// Allocate stack space for spilled registers (256 bytes)
	cg.emitBytes([]byte{0x48, 0x81, 0xEC, 0x00, 0x01, 0x00, 0x00})
}

// emitEpilogue generates function exit code
func (cg *NativeCodeGenerator) emitEpilogue() {
	// add rsp, 0x100
	cg.emitBytes([]byte{0x48, 0x81, 0xC4, 0x00, 0x01, 0x00, 0x00})
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
	// pop rbp
	cg.emitByte(0x5D)
	// ret
	cg.emitByte(0xC3)
}

// compileInstruction compiles a single instruction
func (cg *NativeCodeGenerator) compileInstruction(op compiler.Opcode, code []byte, ip *int) error {
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
		cg.compileMove(dst, src)
		*ip += 3

	case compiler.OpRegAdd:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileAdd(dst, left, right)
		*ip += 4

	case compiler.OpRegSub:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileSub(dst, left, right)
		*ip += 4

	case compiler.OpRegMul:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileMul(dst, left, right)
		*ip += 4

	case compiler.OpRegDiv:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileDiv(dst, left, right)
		*ip += 4

	case compiler.OpRegMod:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileMod(dst, left, right)
		*ip += 4

	case compiler.OpRegNeg:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		cg.compileNeg(dst, src)
		*ip += 3

	case compiler.OpRegLess:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "less")
		*ip += 4

	case compiler.OpRegGreater:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "greater")
		*ip += 4

	case compiler.OpRegEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "equal")
		*ip += 4

	case compiler.OpRegNotEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "not_equal")
		*ip += 4

	case compiler.OpRegLessEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "less_equal")
		*ip += 4

	case compiler.OpRegGreaterEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileCompare(dst, left, right, "greater_equal")
		*ip += 4

	case compiler.OpRegJump:
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		cg.compileJump(target)
		*ip += 4

	case compiler.OpRegJumpIfTrue:
		cond := int(code[*ip+1])
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		cg.compileJumpIfTrue(cond, target)
		*ip += 4

	case compiler.OpRegJumpIfFalse:
		cond := int(code[*ip+1])
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		cg.compileJumpIfFalse(cond, target)
		*ip += 4

	case compiler.OpRegReturn:
		src := int(code[*ip+1])
		cg.compileReturn(src)
		*ip += 2

	case compiler.OpRegNull:
		dst := int(code[*ip+1])
		cg.compileLoadImm(dst, 0)
		*ip += 2

	case compiler.OpRegTrue:
		dst := int(code[*ip+1])
		cg.compileLoadImm(dst, 1)
		*ip += 2

	case compiler.OpRegFalse:
		dst := int(code[*ip+1])
		cg.compileLoadImm(dst, 0)
		*ip += 2

	case compiler.OpRegIncLocal:
		reg := int(code[*ip+1])
		cg.compileInc(reg)
		*ip += 2

	case compiler.OpRegDecLocal:
		reg := int(code[*ip+1])
		cg.compileDec(reg)
		*ip += 2

	case compiler.OpRegLoopCountAdd:
		// Optimized loop: acc += counter; counter++
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		startIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		limitIdx := int(code[*ip+5])<<8 | int(code[*ip+6])
		stepIdx := int(code[*ip+7])<<8 | int(code[*ip+8])
		cg.compileLoopCountAdd(accReg, counterReg, startIdx, limitIdx, stepIdx)
		*ip += 9

	default:
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// ============================================================================
// Register operations
// ============================================================================

// isRegCached returns true if register is in hardware register
func (cg *NativeCodeGenerator) isRegCached(r int) bool {
	return r < 12
}

// loadRegToRax loads a register to rax
func (cg *NativeCodeGenerator) loadRegToRax(r int) {
	if r < 8 {
		// Standard registers
		switch r {
		case 0: // already rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x89, 0xD8}) // mov rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x89, 0xC8}) // mov rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x89, 0xD0}) // mov rax, rdx
		case 4:
			cg.emitBytes([]byte{0x4C, 0x89, 0xC0}) // mov rax, r8
		case 5:
			cg.emitBytes([]byte{0x4C, 0x89, 0xC8}) // mov rax, r9
		case 6:
			cg.emitBytes([]byte{0x4C, 0x89, 0xD0}) // mov rax, r10
		case 7:
			cg.emitBytes([]byte{0x4C, 0x89, 0xD8}) // mov rax, r11
		}
	} else {
		// Spilled to stack
		offset := (r - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x8B, 0x45}) // mov rax, [rbp - offset]
		cg.emitByte(byte(offset))
	}
}

// storeRaxToReg stores rax to a register
func (cg *NativeCodeGenerator) storeRaxToReg(r int) {
	if r < 8 {
		switch r {
		case 0: // already rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x89, 0xC3}) // mov rbx, rax
		case 2:
			cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
		case 3:
			cg.emitBytes([]byte{0x48, 0x89, 0xC2}) // mov rdx, rax
		case 4:
			cg.emitBytes([]byte{0x49, 0x89, 0xC0}) // mov r8, rax
		case 5:
			cg.emitBytes([]byte{0x49, 0x89, 0xC1}) // mov r9, rax
		case 6:
			cg.emitBytes([]byte{0x49, 0x89, 0xC2}) // mov r10, rax
		case 7:
			cg.emitBytes([]byte{0x49, 0x89, 0xC3}) // mov r11, rax
		}
	} else {
		offset := (r - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x89, 0x45}) // mov [rbp - offset], rax
		cg.emitByte(byte(offset))
	}
}

// ============================================================================
// Instruction compilation
// ============================================================================

func (cg *NativeCodeGenerator) compileLoadConst(dst, constIdx int) {
	if constIdx < len(cg.constants) {
		cg.compileLoadImm(dst, cg.constants[constIdx])
	} else {
		cg.compileLoadImm(dst, 0)
	}
}

func (cg *NativeCodeGenerator) compileLoadImm(dst int, val int64) {
	// Load immediate to rax first
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(val))

	// Store to destination
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileMove(dst, src int) {
	if dst == src {
		return
	}

	cg.loadRegToRax(src)
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileAdd(dst, left, right int) {
	// Load left to rax
	cg.loadRegToRax(left)

	// Add right
	if right < 8 {
		switch right {
		case 0:
			cg.emitBytes([]byte{0x48, 0x01, 0xC0}) // add rax, rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x01, 0xD8}) // add rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x01, 0xC8}) // add rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x01, 0xD0}) // add rax, rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0x01, 0xC0}) // add rax, r8
		case 5:
			cg.emitBytes([]byte{0x49, 0x01, 0xC8}) // add rax, r9
		case 6:
			cg.emitBytes([]byte{0x49, 0x01, 0xD0}) // add rax, r10
		case 7:
			cg.emitBytes([]byte{0x49, 0x01, 0xD8}) // add rax, r11
		}
	} else {
		// Add from stack
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x03, 0x45}) // add rax, [rbp - offset]
		cg.emitByte(byte(offset))
	}

	// Store result
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileSub(dst, left, right int) {
	cg.loadRegToRax(left)

	if right < 8 {
		switch right {
		case 0:
			cg.emitBytes([]byte{0x48, 0x29, 0xC0}) // sub rax, rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x29, 0xD8}) // sub rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x29, 0xC8}) // sub rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x29, 0xD0}) // sub rax, rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0x29, 0xC0}) // sub rax, r8
		case 5:
			cg.emitBytes([]byte{0x49, 0x29, 0xC8}) // sub rax, r9
		case 6:
			cg.emitBytes([]byte{0x49, 0x29, 0xD0}) // sub rax, r10
		case 7:
			cg.emitBytes([]byte{0x49, 0x29, 0xD8}) // sub rax, r11
		}
	} else {
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x2B, 0x45})
		cg.emitByte(byte(offset))
	}

	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileMul(dst, left, right int) {
	cg.loadRegToRax(left)

	if right < 8 {
		switch right {
		case 0:
			cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC0}) // imul rax, rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC3}) // imul rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1}) // imul rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC2}) // imul rax, rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC0}) // imul rax, r8
		case 5:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC1}) // imul rax, r9
		case 6:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC2}) // imul rax, r10
		case 7:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC3}) // imul rax, r11
		}
	} else {
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x45})
		cg.emitByte(byte(offset))
	}

	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileDiv(dst, left, right int) {
	// Division in x86-64: idiv uses rdx:rax / src
	// Result: quotient in rax, remainder in rdx

	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x99}) // cqo: sign-extend rax to rdx:rax

	if right < 8 {
		// Need to move right to a register and use idiv
		// Save right value in rcx if it's not already there
		if right == 2 {
			cg.emitBytes([]byte{0x48, 0xF7, 0xF9}) // idiv rcx
		} else {
			// Move right to rcx first
			cg.loadRegToRax(right)
			cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
			// Restore left
			cg.loadRegToRax(left)
			cg.emitBytes([]byte{0x48, 0x99}) // cqo
			cg.emitBytes([]byte{0x48, 0xF7, 0xF9}) // idiv rcx
		}
	} else {
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0xF7, 0x7D}) // idiv [rbp - offset]
		cg.emitByte(byte(offset))
	}

	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileMod(dst, left, right int) {
	// Mod is the remainder after division
	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x99}) // cqo

	if right < 8 {
		if right == 3 {
			cg.emitBytes([]byte{0x48, 0xF7, 0xFA}) // idiv rdx
		} else {
			cg.loadRegToRax(right)
			cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
			cg.loadRegToRax(left)
			cg.emitBytes([]byte{0x48, 0x99})
			cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
		}
	} else {
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0xF7, 0x7D})
		cg.emitByte(byte(offset))
	}

	// Result is in rdx (remainder)
	cg.emitBytes([]byte{0x48, 0x89, 0xD0}) // mov rax, rdx
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileNeg(dst, src int) {
	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileCompare(dst, left, right int, op string) {
	// Compare left and right
	cg.loadRegToRax(left)

	if right < 8 {
		switch right {
		case 0:
			cg.emitBytes([]byte{0x48, 0x39, 0xC0}) // cmp rax, rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x39, 0xD8}) // cmp rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x39, 0xD0}) // cmp rax, rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0x39, 0xC0}) // cmp rax, r8
		case 5:
			cg.emitBytes([]byte{0x49, 0x39, 0xC8}) // cmp rax, r9
		case 6:
			cg.emitBytes([]byte{0x49, 0x39, 0xD0}) // cmp rax, r10
		case 7:
			cg.emitBytes([]byte{0x49, 0x39, 0xD8}) // cmp rax, r11
		}
	} else {
		offset := (right - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0x3B, 0x45})
		cg.emitByte(byte(offset))
	}

	// Set result based on comparison
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax

	switch op {
	case "less":
		cg.emitBytes([]byte{0x0F, 0x9C, 0xC0}) // setl al
	case "greater":
		cg.emitBytes([]byte{0x0F, 0x9F, 0xC0}) // setg al
	case "equal":
		cg.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al
	case "not_equal":
		cg.emitBytes([]byte{0x0F, 0x95, 0xC0}) // setne al
	case "less_equal":
		cg.emitBytes([]byte{0x0F, 0x9E, 0xC0}) // setle al
	case "greater_equal":
		cg.emitBytes([]byte{0x0F, 0x9D, 0xC0}) // setge al
	}

	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileJump(target int) {
	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0xE9}) // jmp rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *NativeCodeGenerator) compileJumpIfTrue(cond, target int) {
	cg.loadRegToRax(cond)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	cg.emitBytes([]byte{0x0F, 0x85})        // jnz rel32
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *NativeCodeGenerator) compileJumpIfFalse(cond, target int) {
	cg.loadRegToRax(cond)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	cg.emitBytes([]byte{0x0F, 0x84})        // jz rel32
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *NativeCodeGenerator) compileReturn(src int) {
	cg.loadRegToRax(src)
	cg.emitEpilogue()
}

func (cg *NativeCodeGenerator) compileInc(reg int) {
	if reg < 8 {
		switch reg {
		case 0:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC0}) // inc rax
		case 1:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC3}) // inc rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC1}) // inc rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC2}) // inc rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC0}) // inc r8
		case 5:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC1}) // inc r9
		case 6:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC2}) // inc r10
		case 7:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC3}) // inc r11
		}
	} else {
		offset := (reg - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0xFF, 0x45}) // inc qword [rbp - offset]
		cg.emitByte(byte(offset))
	}
}

func (cg *NativeCodeGenerator) compileDec(reg int) {
	if reg < 8 {
		switch reg {
		case 0:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC8}) // dec rax
		case 1:
			cg.emitBytes([]byte{0x48, 0xFF, 0xCB}) // dec rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0xFF, 0xC9}) // dec rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0xFF, 0xCA}) // dec rdx
		case 4:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC8}) // dec r8
		case 5:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC9}) // dec r9
		case 6:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCA}) // dec r10
		case 7:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCB}) // dec r11
		}
	} else {
		offset := (reg - 8 + 1) * 8
		cg.emitBytes([]byte{0x48, 0xFF, 0x4D}) // dec qword [rbp - offset]
		cg.emitByte(byte(offset))
	}
}

func (cg *NativeCodeGenerator) compileLoopCountAdd(accReg, counterReg, startIdx, limitIdx, stepIdx int) {
	// Get constant values
	var limit, step int64 = 0, 1
	if limitIdx < len(cg.constants) {
		limit = cg.constants[limitIdx]
	}
	if stepIdx < len(cg.constants) {
		step = cg.constants[stepIdx]
	}
	if step == 0 {
		step = 1
	}

	// Initialize counter with start value
	cg.compileLoadConst(counterReg, startIdx)

	// Initialize accumulator to 0
	cg.compileLoadImm(accReg, 0)

	// Loop label
	loopLabel := fmt.Sprintf("loop_%d", len(cg.code))
	cg.labels[loopLabel] = len(cg.code)

	// Compare counter with limit
	cg.loadRegToRax(counterReg)
	cg.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm32
	cg.emitUint32(uint32(limit))

	// jge end
	endLabel := fmt.Sprintf("loop_end_%d", len(cg.code))
	cg.emitBytes([]byte{0x0F, 0x8D}) // jge rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: endLabel, size: 4})
	cg.emitUint32(0)

	// Add counter to accumulator
	cg.compileAdd(accReg, accReg, counterReg)

	// Increment counter
	cg.compileInc(counterReg)

	// jmp loop
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: loopLabel, size: 4})
	cg.emitUint32(0)

	// End label
	cg.labels[endLabel] = len(cg.code)
}

// ============================================================================
// Low-level emit functions
// ============================================================================

func (cg *NativeCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *NativeCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *NativeCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *NativeCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func (cg *NativeCodeGenerator) resolveFixups() error {
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
