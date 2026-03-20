// pkg/jit/jit_simple.go
// Simplified JIT implementation with correct code generation
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// SimpleCodeGenerator generates x86-64 code with careful attention to correctness
type SimpleCodeGenerator struct {
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction

	// Track stack usage
	stackOffset int
}

// NewSimpleCodeGenerator creates a new simple code generator
func NewSimpleCodeGenerator() *SimpleCodeGenerator {
	return &SimpleCodeGenerator{
		code:   make([]byte, 0, 4096),
		labels: make(map[string]int),
		fixups: make([]fixup, 0),
	}
}

// Generate generates machine code
func (cg *SimpleCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn

	// First pass: identify all labels
	cg.identifyLabels(fn.Instructions)

	// Generate prologue
	cg.emitPrologue()

	// Main compilation loop
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		// Record label position
		cg.labels[fmt.Sprintf("L%d", ip)] = len(cg.code)

		if err := cg.compileInstruction(op, code, &ip); err != nil {
			return nil, err
		}
	}

	// Resolve all fixups
	if err := cg.resolveFixups(); err != nil {
		return nil, err
	}

	return cg.code, nil
}

// identifyLabels does a first pass to identify all jump targets
func (cg *SimpleCodeGenerator) identifyLabels(code []byte) {
	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegJump:
			offset := int16(uint16(code[ip+1])<<8 | uint16(code[ip+2]))
			target := ip + 3 + int(offset)
			cg.labels[fmt.Sprintf("L%d", target)] = -1 // Mark as needed

		case compiler.OpRegJumpIfFalse, compiler.OpRegJumpIfTrue:
			offset := int16(uint16(code[ip+2])<<8 | uint16(code[ip+3]))
			target := ip + 4 + int(offset)
			cg.labels[fmt.Sprintf("L%d", target)] = -1
		}

		// Advance IP
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			ip++
			continue
		}
		ip++
		for _, w := range def.OperandWidths {
			ip += w
		}
	}
}

// emitPrologue generates function entry code
func (cg *SimpleCodeGenerator) emitPrologue() {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})
	// sub rsp, 0x200 (512 bytes for locals)
	cg.emitBytes([]byte{0x48, 0x81, 0xEC, 0x00, 0x02, 0x00, 0x00})
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
}

// emitEpilogue generates function exit code
func (cg *SimpleCodeGenerator) emitEpilogue() {
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

// compileInstruction compiles a single instruction
func (cg *SimpleCodeGenerator) compileInstruction(op compiler.Opcode, code []byte, ip *int) error {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return fmt.Errorf("unknown opcode %d", op)
	}

	switch op {
	case compiler.OpRegLoadConst:
		dst := code[*ip+1]
		constIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileLoadConst(int(dst), constIdx)
		*ip += 4

	case compiler.OpRegMove:
		dst := code[*ip+1]
		src := code[*ip+2]
		cg.compileMove(int(dst), int(src))
		*ip += 3

	case compiler.OpRegAdd:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileAdd(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegSub:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileSub(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegMul:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileMul(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegLess:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileLess(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegEqual:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileEqual(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegJump:
		offset := int16(uint16(code[*ip+1])<<8 | uint16(code[*ip+2]))
		target := *ip + 3 + int(offset)
		cg.compileJump(target)
		*ip += 3

	case compiler.OpRegJumpIfFalse:
		cond := code[*ip+1]
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + 4 + int(offset)
		cg.compileJumpIfFalse(int(cond), target)
		*ip += 4

	case compiler.OpRegJumpIfTrue:
		cond := code[*ip+1]
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + 4 + int(offset)
		cg.compileJumpIfTrue(int(cond), target)
		*ip += 4

	case compiler.OpRegReturn:
		src := code[*ip+1]
		cg.compileReturn(int(src))
		*ip += 2

	case compiler.OpRegNull:
		dst := code[*ip+1]
		cg.compileNull(int(dst))
		*ip += 2

	case compiler.OpRegTrue:
		dst := code[*ip+1]
		cg.compileTrue(int(dst))
		*ip += 2

	case compiler.OpRegFalse:
		dst := code[*ip+1]
		cg.compileFalse(int(dst))
		*ip += 2

	case compiler.OpRegLoadLocal:
		dst := code[*ip+1]
		local := code[*ip+2]
		cg.compileMove(int(dst), int(local))
		*ip += 3

	case compiler.OpRegStoreLocal:
		local := code[*ip+1]
		src := code[*ip+2]
		cg.compileMove(int(local), int(src))
		*ip += 3

	case compiler.OpRegLoadGlobal:
		dst := code[*ip+1]
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileLoadGlobal(int(dst), globalIdx)
		*ip += 4

	case compiler.OpRegStoreGlobal:
		// For JIT, globals are snapshots - store is a no-op
		*ip += 4

	case compiler.OpRegIncLocal:
		local := code[*ip+1]
		cg.compileIncLocal(int(local))
		*ip += 2

	case compiler.OpRegDecLocal:
		local := code[*ip+1]
		cg.compileDecLocal(int(local))
		*ip += 2

	case compiler.OpRegNot:
		dst := code[*ip+1]
		src := code[*ip+2]
		cg.compileNot(int(dst), int(src))
		*ip += 3

	case compiler.OpRegDiv:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileDiv(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegMod:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileMod(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegNeg:
		dst := code[*ip+1]
		src := code[*ip+2]
		cg.compileNeg(int(dst), int(src))
		*ip += 3

	case compiler.OpRegNotEqual:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileNotEqual(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegGreater:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileGreater(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegLessEqual:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileLessEqual(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegGreaterEqual:
		dst := code[*ip+1]
		left := code[*ip+2]
		right := code[*ip+3]
		cg.compileGreaterEqual(int(dst), int(left), int(right))
		*ip += 4

	case compiler.OpRegPush:
		src := code[*ip+1]
		cg.compilePush(int(src))
		*ip += 2

	case compiler.OpRegPop:
		dst := code[*ip+1]
		cg.compilePop(int(dst))
		*ip += 2

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++

	case compiler.OpRegCall, compiler.OpRegTailCall, compiler.OpRegClosure:
		return fmt.Errorf("opcode %s requires interpreter", def.Name)

	default:
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// ============================================================================
// Instruction Implementations
// ============================================================================

func (cg *SimpleCodeGenerator) compileLoadConst(dst, constIdx int) {
	var val uint64
	if constIdx < len(cg.constants) {
		val = uint64(cg.constants[constIdx])
	} else {
		val = 0
	}

	// mov rax, imm64
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(val)
	// mov [rbp - dst*8 - 8], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileMove(dst, src int) {
	// mov rax, [rbp - src*8 - 8]
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(src + 1) * 8))
	// mov [rbp - dst*8 - 8], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileAdd(dst, left, right int) {
	// mov rax, [rbp - left*8 - 8]
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	// mov rcx, [rbp - right*8 - 8]
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	// add rax, rcx
	cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	// mov [rbp - dst*8 - 8], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileSub(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x29, 0xC8})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileMul(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileLess(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	// cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	// xor rax, rax
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	// setl al
	cg.emitBytes([]byte{0x0F, 0x9C, 0xC0})
	// Store result
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileEqual(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileJump(target int) {
	label := fmt.Sprintf("L%d", target)
	// jmp rel32
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  label,
		size:   4,
	})
	cg.emitUint32(0)
}

func (cg *SimpleCodeGenerator) compileJumpIfFalse(cond, target int) {
	// mov rax, [rbp - cond*8 - 8]
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(cond + 1) * 8))
	// test rax, rax
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	// jz rel32
	cg.emitBytes([]byte{0x0F, 0x84})
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  label,
		size:   4,
	})
	cg.emitUint32(0)
}

func (cg *SimpleCodeGenerator) compileJumpIfTrue(cond, target int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(cond + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	// jnz rel32
	cg.emitBytes([]byte{0x0F, 0x85})
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  label,
		size:   4,
	})
	cg.emitUint32(0)
}

func (cg *SimpleCodeGenerator) compileReturn(src int) {
	// mov rax, [rbp - src*8 - 8]
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(src + 1) * 8))
	cg.emitEpilogue()
}

func (cg *SimpleCodeGenerator) compileNull(dst int) {
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileTrue(dst int) {
	cg.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00}) // mov rax, 1
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileFalse(dst int) {
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileLoadGlobal(dst, globalIdx int) {
	var val uint64
	if globalIdx < len(cg.globals) {
		val = uint64(cg.globals[globalIdx])
	} else {
		val = 0
	}
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(val)
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileIncLocal(local int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(local + 1) * 8))
	cg.emitBytes([]byte{0x48, 0xFF, 0xC0}) // inc rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(local + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileDecLocal(local int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(local + 1) * 8))
	cg.emitBytes([]byte{0x48, 0xFF, 0xC8}) // dec rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(local + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileNot(dst, src int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(src + 1) * 8))
	// test rax, rax
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	// setz al
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	// movzx eax, al
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileDiv(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	// cqo (sign extend)
	cg.emitBytes([]byte{0x48, 0x99})
	// idiv rcx
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileMod(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	// mov rax, rdx (remainder)
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileNeg(dst, src int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(src + 1) * 8))
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileNotEqual(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x95, 0xC0}) // setne al
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileGreater(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9F, 0xC0}) // setg al
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileLessEqual(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9E, 0xC0}) // setle al
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compileGreaterEqual(dst, left, right int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(left + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-(right + 1) * 8))
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9D, 0xC0}) // setge al
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

func (cg *SimpleCodeGenerator) compilePush(src int) {
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-(src + 1) * 8))
	cg.emitByte(0x50) // push rax
}

func (cg *SimpleCodeGenerator) compilePop(dst int) {
	cg.emitByte(0x58) // pop rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-(dst + 1) * 8))
}

// ============================================================================
// Helper Functions
// ============================================================================

func (cg *SimpleCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *SimpleCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *SimpleCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *SimpleCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func (cg *SimpleCodeGenerator) resolveFixups() error {
	for _, f := range cg.fixups {
		target, ok := cg.labels[f.label]
		if !ok {
			return fmt.Errorf("undefined label: %s", f.label)
		}

		// Calculate relative offset
		// offset = target - (current_position + 4)
		offset := int32(target - (f.offset + f.size))

		binary.LittleEndian.PutUint32(cg.code[f.offset:], uint32(offset))
	}
	return nil
}
