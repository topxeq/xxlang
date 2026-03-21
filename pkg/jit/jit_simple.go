// +build amd64,!windows

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

// SavedRegsSize is the bytes used for saved callee-saved registers on stack
// Layout: rbp(8) + rbx(8) + r12(8) + r13(8) + r14(8) + r15(8) = 48 bytes
// Local variables start at [rbp - SavedRegsSize - localOffset]
const SavedRegsSize = 48

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
	// push callee-saved registers (must be pushed before allocating local space)
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
	// sub rsp, 0x200 (512 bytes for locals)
	// Stack layout after this:
	// [rbp]     = old rbp
	// [rbp-8]   = saved rbx
	// [rbp-16]  = saved r12
	// [rbp-24]  = saved r13
	// [rbp-32]  = saved r14
	// [rbp-40]  = saved r15
	// [rbp-48] to [rbp-560] = local variables
	cg.emitBytes([]byte{0x48, 0x81, 0xEC, 0x00, 0x02, 0x00, 0x00})
}

// emitEpilogue generates function exit code
func (cg *SimpleCodeGenerator) emitEpilogue() {
	// add rsp, 0x200 (deallocate local space FIRST)
	cg.emitBytes([]byte{0x48, 0x81, 0xC4, 0x00, 0x02, 0x00, 0x00})
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

	case compiler.OpRegLoopCountAdd:
		// Optimized counting loop: acc_reg, counter_reg, start(16), limit(16), step(16)
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		start := int(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
		limit := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
		step := int(int16(uint16(code[*ip+7])<<8 | uint16(code[*ip+8])))
		cg.compileLoopCountAdd(accReg, counterReg, start, limit, step)
		*ip += 9

	case compiler.OpRegLoopBodyAdd:
		// Loop body add: acc_reg, counter_reg, limit(16), jump_offset(16)
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		limit := int(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
		jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
		cg.compileLoopBodyAdd(accReg, counterReg, limit, jumpOffset, *ip+7)
		*ip += 7

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++

	case compiler.OpRegCall:
		// Function call: compile as inline interpreter call
		funcReg := int(code[*ip+1])
		numArgs := int(code[*ip+2])
		cg.compileCall(funcReg, numArgs)
		*ip += 3

	case compiler.OpRegTailCall:
		// Tail call: compile as jump back to function start
		funcReg := int(code[*ip+1])
		numArgs := int(code[*ip+2])
		cg.compileTailCall(funcReg, numArgs)
		*ip += 3

	case compiler.OpRegClosure:
		dst := int(code[*ip+1])
		cg.compileClosure(dst)
		*ip += 6

	// Array operations - use interpreter fallback via placeholder
	case compiler.OpRegArray:
		// R[dst] = Array from R[start..start+count-1]
		dst := int(code[*ip+1])
		startReg := int(code[*ip+2])
		count := int(code[*ip+3])
		cg.compileArrayCreate(dst, startReg, count)
		*ip += 4

	case compiler.OpRegArrayEmpty:
		// R[dst] = empty array
		dst := int(code[*ip+1])
		cg.compileArrayEmpty(dst)
		*ip += 2

	case compiler.OpRegArrayAppend:
		// R[dst] = append(R[arr], R[elem])
		dst := int(code[*ip+1])
		arrReg := int(code[*ip+2])
		elemReg := int(code[*ip+3])
		cg.compileArrayAppend(dst, arrReg, elemReg)
		*ip += 4

	case compiler.OpRegIndex:
		// R[dst] = R[obj][R[key]]
		dst := int(code[*ip+1])
		objReg := int(code[*ip+2])
		keyReg := int(code[*ip+3])
		cg.compileIndex(dst, objReg, keyReg)
		*ip += 4

	case compiler.OpRegSetIndex:
		// R[obj][R[key]] = R[val]
		objReg := int(code[*ip+1])
		keyReg := int(code[*ip+2])
		valReg := int(code[*ip+3])
		cg.compileSetIndex(objReg, keyReg, valReg)
		*ip += 4

	// Map operations
	case compiler.OpRegMap:
		// R[dst] = Map from key-value pairs
		dst := int(code[*ip+1])
		_ = dst // Map creation requires interpreter support
		cg.compileNull(dst)
		*ip += 4

	case compiler.OpRegMapSet:
		// R[dst] = R[map] with R[key] = R[val]
		_ = code[*ip+1] // dst
		_ = code[*ip+2] // mapReg
		_ = code[*ip+3] // keyReg
		_ = code[*ip+4] // valReg
		// Map modification requires interpreter support
		*ip += 5

	default:
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// ============================================================================
// Instruction Implementations
// ============================================================================

func (cg *SimpleCodeGenerator) compileLoadConst(dst, constIdx int) {
	var val int64
	if constIdx < len(cg.constants) {
		// Properly extract integer from NaN-boxed Value
		if cg.constants[constIdx].IsInt() {
			val = cg.constants[constIdx].GetInt()
		}
		// For floats, we could convert but for JIT we focus on integers
	}

	// mov rax, imm64
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(val))
	// mov [rbp - offset], rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileMove(dst, src int) {
	// mov rax, [rbp - offset]
	cg.emitMovSlotToRax(src)
	// mov [rbp - offset], rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileAdd(dst, left, right int) {
	// mov rax, [rbp - offset]
	cg.emitMovSlotToRax(left)
	// mov rcx, [rbp - offset]
	cg.emitMovSlotToRcx(right)
	// add rax, rcx
	cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	// mov [rbp - offset], rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileSub(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x29, 0xC8})
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileMul(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1})
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileLess(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	// cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	// xor rax, rax
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	// setl al
	cg.emitBytes([]byte{0x0F, 0x9C, 0xC0})
	// Store result
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileEqual(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	cg.emitMovRaxToSlot(dst)
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
	// mov rax, [rbp - offset]
	cg.emitMovSlotToRax(cond)
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
	cg.emitMovSlotToRax(cond)
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
	// mov rax, [rbp - offset]
	cg.emitMovSlotToRax(src)
	cg.emitEpilogue()
}

func (cg *SimpleCodeGenerator) compileNull(dst int) {
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileTrue(dst int) {
	cg.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00}) // mov rax, 1
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileFalse(dst int) {
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileLoadGlobal(dst, globalIdx int) {
	var val int64
	if globalIdx < len(cg.globals) {
		// Properly extract integer from NaN-boxed Value
		if cg.globals[globalIdx].IsInt() {
			val = cg.globals[globalIdx].GetInt()
		}
	}
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(val))
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileIncLocal(local int) {
	cg.emitMovSlotToRax(local)
	cg.emitBytes([]byte{0x48, 0xFF, 0xC0}) // inc rax
	cg.emitMovRaxToSlot(local)
}

func (cg *SimpleCodeGenerator) compileDecLocal(local int) {
	cg.emitMovSlotToRax(local)
	cg.emitBytes([]byte{0x48, 0xFF, 0xC8}) // dec rax
	cg.emitMovRaxToSlot(local)
}

func (cg *SimpleCodeGenerator) compileNot(dst, src int) {
	cg.emitMovSlotToRax(src)
	// test rax, rax
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	// setz al
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	// movzx eax, al
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileDiv(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	// cqo (sign extend)
	cg.emitBytes([]byte{0x48, 0x99})
	// idiv rcx
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileMod(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	// mov rax, rdx (remainder)
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileNeg(dst, src int) {
	cg.emitMovSlotToRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileNotEqual(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x95, 0xC0}) // setne al
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileGreater(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9F, 0xC0}) // setg al
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileLessEqual(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9E, 0xC0}) // setle al
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compileGreaterEqual(dst, left, right int) {
	cg.emitMovSlotToRax(left)
	cg.emitMovSlotToRcx(right)
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x9D, 0xC0}) // setge al
	cg.emitMovRaxToSlot(dst)
}

func (cg *SimpleCodeGenerator) compilePush(src int) {
	cg.emitMovSlotToRax(src)
	cg.emitByte(0x50) // push rax
}

func (cg *SimpleCodeGenerator) compilePop(dst int) {
	cg.emitByte(0x58) // pop rax
	cg.emitMovRaxToSlot(dst)
}

// ============================================================================
// Helper Functions
// ============================================================================

// localOffset returns the stack offset for local variable 'slot'
// Local variables are stored at [rbp - SavedRegsSize - (slot+1)*8]
func (cg *SimpleCodeGenerator) localOffset(slot int) int {
	return SavedRegsSize + (slot+1)*8
}

// emitMovRaxToSlot stores rax to local variable slot
// Uses proper addressing mode based on offset size
func (cg *SimpleCodeGenerator) emitMovRaxToSlot(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		// Short form: mov [rbp - disp8], rax
		cg.emitBytes([]byte{0x48, 0x89, 0x45})
		cg.emitByte(byte(-offset))
	} else {
		// Long form: mov [rbp - disp32], rax
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32(-offset))
	}
}

// emitMovSlotToRax loads local variable slot to rax
func (cg *SimpleCodeGenerator) emitMovSlotToRax(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		// Short form: mov rax, [rbp - disp8]
		cg.emitBytes([]byte{0x48, 0x8B, 0x45})
		cg.emitByte(byte(-offset))
	} else {
		// Long form: mov rax, [rbp - disp32]
		cg.emitBytes([]byte{0x48, 0x8B, 0x85})
		cg.emitUint32(uint32(-offset))
	}
}

// emitMovRcxToSlot stores rcx to local variable slot
func (cg *SimpleCodeGenerator) emitMovRcxToSlot(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0x89, 0x4D})
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0x89, 0x8D})
		cg.emitUint32(uint32(-offset))
	}
}

// emitMovSlotToRcx loads local variable slot to rcx
func (cg *SimpleCodeGenerator) emitMovSlotToRcx(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0x8B, 0x8D})
		cg.emitUint32(uint32(-offset))
	}
}

// emitAddRaxToSlot adds rax to local variable slot
func (cg *SimpleCodeGenerator) emitAddRaxToSlot(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0x01, 0x45})
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0x01, 0x85})
		cg.emitUint32(uint32(-offset))
	}
}

// emitIncSlot increments local variable slot
func (cg *SimpleCodeGenerator) emitIncSlot(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0xFF, 0x45})
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0xFF, 0x85})
		cg.emitUint32(uint32(-offset))
	}
}

// emitDecSlot decrements local variable slot
func (cg *SimpleCodeGenerator) emitDecSlot(slot int) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0xFF, 0x4D})
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0xFF, 0x8D})
		cg.emitUint32(uint32(-offset))
	}
}

// emitMovImm32ToSlot moves 32-bit immediate to local variable slot
func (cg *SimpleCodeGenerator) emitMovImm32ToSlot(slot int, val uint32) {
	offset := cg.localOffset(slot)
	if offset <= 127 {
		cg.emitBytes([]byte{0x48, 0xC7, 0x45}) // mov qword [rbp - disp8], imm32
		cg.emitByte(byte(-offset))
	} else {
		cg.emitBytes([]byte{0x48, 0xC7, 0x85}) // mov qword [rbp - disp32], imm32
		cg.emitUint32(uint32(-offset))
	}
	cg.emitUint32(val)
}

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

// ============================================================================
// Function Call Support
// ============================================================================

// compileCall compiles a function call instruction
// For JIT, we use a hybrid approach:
// 1. For self-recursive calls, we could inline them as loops
// 2. For other calls, we use a trampoline to the interpreter
func (cg *SimpleCodeGenerator) compileCall(funcReg, numArgs int) {
	// For now, we implement a simple approach:
	// Save current state, set up arguments, and use a call gate
	//
	// The call gate is a placeholder that will be patched at runtime
	// to call the interpreter for the actual function execution.

	// Move arguments to the "argument area" (registers 16-23)
	// This preserves R0-R7 for the callee
	for i := 0; i < numArgs && i < 8; i++ {
		cg.emitMovSlotToRax(i)
		cg.emitMovRaxToSlot(16 + i)
	}

	// For the function pointer, store it in a known location
	cg.emitMovSlotToRax(funcReg)

	// Store function pointer for later use (using register slot 30)
	// Use mov with 32-bit displacement for larger offsets
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
	cg.emitUint32(uint32(-cg.localOffset(30)))

	// Push argument count
	// mov rcx, numArgs
	cg.emitBytes([]byte{0x48, 0xB9})
	cg.emitUint64(uint64(numArgs))
	// push rcx
	cg.emitByte(0x51)

	// Call the interpreter trampoline
	// For now, we use a simple approach: call a Go function
	// that will interpret the callee and return the result.

	// mov rax, imm64 (placeholder for trampoline address)
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(0) // Will be patched at runtime
	// call rax
	cg.emitBytes([]byte{0xFF, 0xD0})

	// Restore stack (pop argument count)
	cg.emitByte(0x59) // pop rcx

	// Store result in return register (R0)
	cg.emitMovRaxToSlot(0)
}

// compileTailCall compiles a tail call instruction
// For self-recursive tail calls, we can optimize to a loop
func (cg *SimpleCodeGenerator) compileTailCall(funcReg, numArgs int) {
	// For tail calls, we want to:
	// 1. Move new arguments to R0-R7
	// 2. Jump back to function entry

	// Move arguments to R0-R7
	for i := 0; i < numArgs && i < 8; i++ {
		cg.emitMovSlotToRax(16 + i)
		cg.emitMovRaxToSlot(i)
	}

	// Jump to function entry (label L0)
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  "L0",
		size:   4,
	})
	cg.emitUint32(0)

	_ = funcReg // Function register not needed for self-recursive tail call
}

// compileClosure compiles a closure creation instruction
func (cg *SimpleCodeGenerator) compileClosure(dst int) {
	// For JIT, closures need interpreter support
	// We create a placeholder value

	// Set the destination to null for now
	// Real implementation would create a closure object
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitMovRaxToSlot(dst)
}

// SetTrampolineAddress patches the trampoline address in generated code
// This should be called after code generation to set up the actual call target
func (cg *SimpleCodeGenerator) SetTrampolineAddress(code []byte, offset int, addr uintptr) {
	if offset+8 <= len(code) {
		binary.LittleEndian.PutUint64(code[offset:], uint64(addr))
	}
}

// compileLoopCountAdd compiles an optimized counting loop
// Performs: for (counter = start; counter < limit; counter += step) { acc += counter }
func (cg *SimpleCodeGenerator) compileLoopCountAdd(accReg, counterReg, start, limit, step int) {
	// Initialize counter with start value
	cg.emitMovImm32ToSlot(counterReg, uint32(start))

	// Initialize accumulator with 0
	cg.emitMovImm32ToSlot(accReg, 0)

	// Loop label
	loopLabel := fmt.Sprintf("loop_%d", len(cg.code))
	cg.labels[loopLabel] = len(cg.code)

	// Load counter and compare with limit
	cg.emitMovSlotToRax(counterReg)

	// cmp rax, limit
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint32(uint32(limit))

	// jge end (exit loop if counter >= limit)
	endLabel := fmt.Sprintf("loop_end_%d", len(cg.code))
	cg.emitBytes([]byte{0x0F, 0x8D}) // jge rel32
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  endLabel,
		size:   4,
	})
	cg.emitUint32(0)

	// Add counter to accumulator
	cg.emitMovSlotToRax(counterReg)
	cg.emitAddRaxToSlot(accReg)

	// Increment counter by step
	if step == 1 {
		cg.emitIncSlot(counterReg)
	} else {
		// add qword [rbp - offset], step
		offset := cg.localOffset(counterReg)
		if offset <= 127 {
			cg.emitBytes([]byte{0x48, 0x81, 0x45})
			cg.emitByte(byte(-offset))
		} else {
			cg.emitBytes([]byte{0x48, 0x81, 0x85})
			cg.emitUint32(uint32(-offset))
		}
		cg.emitUint32(uint32(step))
	}

	// jmp loop
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  loopLabel,
		size:   4,
	})
	cg.emitUint32(0)

	// End label
	cg.labels[endLabel] = len(cg.code)
}

// compileLoopBodyAdd compiles a loop body add instruction
// Performs: acc += counter; counter++; if counter < limit jump to offset
func (cg *SimpleCodeGenerator) compileLoopBodyAdd(accReg, counterReg, limit, jumpOffset int, currentIP int) {
	// Load counter and add to accumulator
	cg.emitMovSlotToRax(counterReg)
	cg.emitAddRaxToSlot(accReg)

	// Increment counter
	cg.emitIncSlot(counterReg)

	// Load counter and compare with limit
	cg.emitMovSlotToRax(counterReg)

	// cmp rax, limit
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint32(uint32(limit))

	// jl back to loop
	targetIP := currentIP + jumpOffset
	targetLabel := fmt.Sprintf("L%d", targetIP)
	cg.emitBytes([]byte{0x0F, 0x8C}) // jl rel32
	cg.fixups = append(cg.fixups, fixup{
		offset: len(cg.code),
		label:  targetLabel,
		size:   4,
	})
	cg.emitUint32(0)
}

// ============================================================================
// Array Operations
// ============================================================================

// compileArrayCreate creates an array from registers
// Note: This is a stub implementation. Real array creation requires Go runtime.
func (cg *SimpleCodeGenerator) compileArrayCreate(dst, startReg, count int) {
	// For JIT without Go object support, we create a placeholder
	// Real implementation would call a Go helper function
	_ = startReg
	_ = count
	cg.compileNull(dst)
}

// compileArrayEmpty creates an empty array
func (cg *SimpleCodeGenerator) compileArrayEmpty(dst int) {
	// Placeholder: set to null (real implementation would create empty array)
	cg.compileNull(dst)
}

// compileArrayAppend appends an element to an array
func (cg *SimpleCodeGenerator) compileArrayAppend(dst, arrReg, elemReg int) {
	// Placeholder: copy source array (real implementation would create new array)
	_ = elemReg
	cg.compileMove(dst, arrReg)
}

// compileIndex gets an element by index
func (cg *SimpleCodeGenerator) compileIndex(dst, objReg, keyReg int) {
	// Placeholder: return null (real implementation would access array/map)
	_ = objReg
	_ = keyReg
	cg.compileNull(dst)
}

// compileSetIndex sets an element by index
func (cg *SimpleCodeGenerator) compileSetIndex(objReg, keyReg, valReg int) {
	// Placeholder: no-op (real implementation would modify array/map)
	_ = objReg
	_ = keyReg
	_ = valReg
	// No-op: the array modification would require creating a new array
}
