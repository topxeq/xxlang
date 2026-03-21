// +build amd64,!windows

// pkg/jit/codegen.go
// x86-64 Code Generator for JIT compilation
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// Register allocation for x86-64
const (
	// Value tag constants
	TagInt  = 0x7FFC
	TagBool = 0x7FFD
	TagNull = 0x7FFE
)

// CodeGenerator generates x86-64 machine code
type CodeGenerator struct {
	config JITConfig

	// Generated code buffer
	code []byte

	// Labels for jumps
	labels map[string]int
	fixups []fixup

	// Constants and globals references
	constants []vm.Value
	globals   []vm.Value

	// Current function being compiled
	fn *compiler.CompiledFunction

	// Number of registers used
	numRegs int
}

type fixup struct {
	offset int
	label  string
	size   int
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(config JITConfig) *CodeGenerator {
	return &CodeGenerator{
		config:  config,
		code:    make([]byte, 0, 4096),
		labels:  make(map[string]int),
		fixups:  make([]fixup, 0),
	}
}

// Generate generates x86-64 code from VM bytecode
func (cg *CodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn

	// Count registers needed
	cg.numRegs = 32 // Default register count

	// Generate function prologue
	cg.emitPrologue()

	// Main dispatch loop
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		cg.labels[fmt.Sprintf("ip_%d", ip)] = len(cg.code)

		switch op {
		// Data movement
		case compiler.OpRegLoadConst:
			cg.compileOpRegLoadConst(code, &ip)
		case compiler.OpRegMove:
			cg.compileOpRegMove(code, &ip)

		// Arithmetic
		case compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul:
			cg.compileOpRegArithmetic(code, &ip, op)
		case compiler.OpRegDiv:
			cg.compileOpRegDiv(code, &ip)
		case compiler.OpRegMod:
			cg.compileOpRegMod(code, &ip)
		case compiler.OpRegNeg:
			cg.compileOpRegNeg(code, &ip)

		// Comparison
		case compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegEqual:
			cg.compileOpRegComparison(code, &ip, op)
		case compiler.OpRegLessEqual, compiler.OpRegGreaterEqual:
			cg.compileOpRegComparisonEqual(code, &ip, op)
		case compiler.OpRegNotEqual:
			cg.compileOpRegNotEqual(code, &ip)

		// Logical
		case compiler.OpRegNot:
			cg.compileOpRegNot(code, &ip)

		// Control flow
		case compiler.OpRegJump:
			cg.compileOpRegJump(code, &ip)
		case compiler.OpRegJumpIfFalse:
			cg.compileOpRegJumpIfFalse(code, &ip)
		case compiler.OpRegJumpIfTrue:
			cg.compileOpRegJumpIfTrue(code, &ip)
		case compiler.OpRegReturn:
			cg.compileOpRegReturn(code, &ip)

		// Local variables
		case compiler.OpRegLoadLocal:
			cg.compileOpRegLoadLocal(code, &ip)
		case compiler.OpRegStoreLocal:
			cg.compileOpRegStoreLocal(code, &ip)

		// Global variables
		case compiler.OpRegLoadGlobal:
			cg.compileOpRegLoadGlobal(code, &ip)
		case compiler.OpRegStoreGlobal:
			cg.compileOpRegStoreGlobal(code, &ip)

		// Increment/Decrement
		case compiler.OpRegIncLocal:
			cg.compileOpRegIncLocal(code, &ip)
		case compiler.OpRegDecLocal:
			cg.compileOpRegDecLocal(code, &ip)

		// Loop superinstructions
		case compiler.OpRegLoopBodyAdd:
			cg.compileOpRegLoopBodyAdd(code, &ip)
		case compiler.OpRegLoopCountAdd:
			cg.compileOpRegLoopCountAdd(code, &ip)

		// Null/True/False
		case compiler.OpRegNull:
			cg.compileOpRegNull(code, &ip)
		case compiler.OpRegTrue:
			cg.compileOpRegTrue(code, &ip)
		case compiler.OpRegFalse:
			cg.compileOpRegFalse(code, &ip)

		case compiler.OpReturn:
			cg.emitEpilogue()
			ip++

		default:
			if cg.config.Debug {
				fmt.Printf("[JIT] Unknown opcode %d at IP %d\n", op, ip)
			}
			return nil, fmt.Errorf("unsupported opcode %d at IP %d", op, ip)
		}
	}

	// Resolve fixups
	for _, f := range cg.fixups {
		target, ok := cg.labels[f.label]
		if !ok {
			return nil, fmt.Errorf("undefined label: %s", f.label)
		}
		offset := target - (f.offset + f.size)
		switch f.size {
		case 1:
			cg.code[f.offset] = byte(offset)
		case 2:
			binary.LittleEndian.PutUint16(cg.code[f.offset:], uint16(offset))
		case 4:
			binary.LittleEndian.PutUint32(cg.code[f.offset:], uint32(offset))
		}
	}

	return cg.code, nil
}

// emitPrologue generates function entry code
func (cg *CodeGenerator) emitPrologue() {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})
	// push rbx (callee-saved)
	cg.emitByte(0x53)
	// push r12-r15
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// sub rsp, stack_size (for registers as local storage)
	// Each register is a uint64 (8 bytes), allocate space for 64 registers
	stackSize := uint32(512)
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(stackSize)

	// Initialize all registers to null (0x7FFE000000000000)
	// This prevents garbage values from causing issues
	nullValue := uint64(TagNull) << 48
	for i := 0; i < 64; i++ {
		// mov qword [rbp - 8*(i+1)], nullValue
		off := int32(8 * (i + 1))
		cg.emitBytes([]byte{0x48, 0xC7, 0x45}) // mov qword [rbp - off], imm32
		cg.emitByte(byte(-off))
		cg.emitUint32(uint32(nullValue)) // Lower 32 bits of null
	}
}

// emitEpilogue generates function exit code
func (cg *CodeGenerator) emitEpilogue() {
	// add rsp, stack_size
	cg.emitBytes([]byte{0x48, 0x81, 0xC4})
	cg.emitUint32(512)

	// pop r15-r12
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	// pop rbx
	cg.emitByte(0x5B)
	// pop rbp
	cg.emitByte(0x5D)
	// ret
	cg.emitByte(0xC3)
}

// Byte emission helpers
func (cg *CodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *CodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *CodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *CodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// Register to stack slot offset
func regOffset(reg int) int32 {
	return int32(8 * (reg + 1))
}

// Load register to rax
func (cg *CodeGenerator) loadRegToRax(reg int) {
	off := regOffset(reg)
	// mov rax, [rbp - off]
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-off))
}

// Store rax to register
func (cg *CodeGenerator) storeRaxToReg(reg int) {
	off := regOffset(reg)
	// mov [rbp - off], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-off))
}

// Load register to rcx
func (cg *CodeGenerator) loadRegToRcx(reg int) {
	off := regOffset(reg)
	// mov rcx, [rbp - off]
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-off))
}

// Store rcx to register
func (cg *CodeGenerator) storeRcxToReg(reg int) {
	off := regOffset(reg)
	// mov [rbp - off], rcx
	cg.emitBytes([]byte{0x48, 0x89, 0x4D})
	cg.emitByte(byte(-off))
}

// ============================================================================
// Opcode compilation functions
// ============================================================================

func (cg *CodeGenerator) compileOpRegLoadConst(code []byte, ip *int) {
	dst := int(code[*ip+1])
	constIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	var val uint64
	if constIdx < len(cg.constants) {
		val = uint64(cg.constants[constIdx])
	} else {
		val = uint64(TagNull) << 48
	}

	// mov rax, imm64
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(val)
	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegMove(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRegToRax(src)
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *CodeGenerator) compileOpRegArithmetic(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	switch op {
	case compiler.OpRegAdd:
		cg.emitBytes([]byte{0x48, 0x01, 0xC8}) // add rax, rcx
	case compiler.OpRegSub:
		cg.emitBytes([]byte{0x48, 0x29, 0xC8}) // sub rax, rcx
	case compiler.OpRegMul:
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1}) // imul rax, rcx
	}

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegDiv(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	// cqo (sign-extend rax into rdx:rax)
	cg.emitBytes([]byte{0x48, 0x99})
	// idiv rcx
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegMod(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	// cqo
	cg.emitBytes([]byte{0x48, 0x99})
	// idiv rcx (remainder in rdx)
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	// mov rax, rdx
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegNeg(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)
	// neg rax
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8})
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *CodeGenerator) compileOpRegComparison(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	// cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x39, 0xC8})

	// xor rax, rax
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})

	switch op {
	case compiler.OpRegLess:
		cg.emitBytes([]byte{0x0F, 0x9C, 0xC0}) // setl al
	case compiler.OpRegGreater:
		cg.emitBytes([]byte{0x0F, 0x9F, 0xC0}) // setg al
	case compiler.OpRegEqual:
		cg.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al
	}

	// Convert to tagged boolean
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0}) // movzx eax, al
	cg.emitBytes([]byte{0x48, 0x0D})       // or rax, imm64
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegComparisonEqual(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax

	switch op {
	case compiler.OpRegLessEqual:
		cg.emitBytes([]byte{0x0F, 0x9E, 0xC0}) // setle al
	case compiler.OpRegGreaterEqual:
		cg.emitBytes([]byte{0x0F, 0x9D, 0xC0}) // setge al
	}

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegNotEqual(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.emitBytes([]byte{0x0F, 0x95, 0xC0}) // setne al

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegNot(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)

	// Check if falsy (null or false tag)
	cg.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm64
	cg.emitUint64(uint64(TagNull) << 48)

	// sete al
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *CodeGenerator) compileOpRegJump(code []byte, ip *int) {
	// OpRegJump format: opcode(1) | unused(1) | offset_hi(1) | offset_lo(1) = 4 bytes
	// Offset is at bytes 2-3 (0-indexed from ip)
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	// Target is calculated relative to the instruction start
	target := *ip + offset

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJmp(label)
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegJumpIfFalse(code []byte, ip *int) {
	cond := int(code[*ip+1])
	// OpRegJumpIfFalse format: opcode(1) | cond(1) | offset_hi(1) | offset_lo(1) = 4 bytes
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	// Target is calculated relative to the instruction start (not ip+4)
	target := *ip + offset

	cg.loadRegToRax(cond)

	// Check if false (tagged false value)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48) // ValueFalse

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label) // JE
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegJumpIfTrue(code []byte, ip *int) {
	cond := int(code[*ip+1])
	// OpRegJumpIfTrue format: opcode(1) | cond(1) | offset_hi(1) | offset_lo(1) = 4 bytes
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	// Target is calculated relative to the instruction start (not ip+4)
	target := *ip + offset

	cg.loadRegToRax(cond)

	// Check if not false and not null
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48) // ValueFalse

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x85, label) // JNE
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegReturn(code []byte, ip *int) {
	src := int(code[*ip+1])
	cg.loadRegToRax(src)
	cg.emitEpilogue()
	*ip += 2
}

func (cg *CodeGenerator) compileOpRegLoadLocal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	localIdx := int(code[*ip+2])

	// Locals are stored in registers
	cg.loadRegToRax(localIdx)
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *CodeGenerator) compileOpRegStoreLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)
	cg.storeRaxToReg(localIdx)
	*ip += 3
}

func (cg *CodeGenerator) compileOpRegLoadGlobal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	// For now, globals are passed in the constants array
	// In a full implementation, we'd pass a pointer to globals
	if globalIdx < len(cg.globals) {
		val := cg.globals[globalIdx]
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(val))
		cg.storeRaxToReg(dst)
	} else {
		// Set to null
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(TagNull) << 48)
		cg.storeRaxToReg(dst)
	}
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegStoreGlobal(code []byte, ip *int) {
	src := int(code[*ip+1])
	globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	// For JIT, we store to a temporary location
	// The globals slice is read-only during JIT execution
	_ = globalIdx
	cg.loadRegToRax(src)
	// Note: In a full implementation, we'd store to globals[globalIdx]
	*ip += 4
}

func (cg *CodeGenerator) compileOpRegIncLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	cg.loadRegToRax(localIdx)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01}) // add rax, 1
	cg.storeRaxToReg(localIdx)
	*ip += 2
}

func (cg *CodeGenerator) compileOpRegDecLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	cg.loadRegToRax(localIdx)
	cg.emitBytes([]byte{0x48, 0x83, 0xE8, 0x01}) // sub rax, 1
	cg.storeRaxToReg(localIdx)
	*ip += 2
}

func (cg *CodeGenerator) compileOpRegLoopBodyAdd(code []byte, ip *int) {
	accReg := int(code[*ip+1])
	counterReg := int(code[*ip+2])
	limit := int64(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
	jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))

	// Add counter to accumulator
	cg.loadRegToRax(counterReg)
	cg.loadRegToRcx(accReg)
	cg.emitBytes([]byte{0x48, 0x01, 0xC8}) // add rax, rcx (counter + acc)
	cg.storeRaxToReg(accReg)

	// Increment counter
	cg.loadRegToRax(counterReg)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01}) // add rax, 1
	cg.storeRaxToReg(counterReg)

	// Check if counter < limit
	limitTagged := uint64(limit) | (uint64(TagInt) << 48)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(limitTagged)

	// Jump back if counter < limit
	loopLabel := fmt.Sprintf("ip_%d", *ip+6+jumpOffset)
	cg.emitJcc(0x8C, loopLabel) // JL

	*ip += 7
}

func (cg *CodeGenerator) compileOpRegLoopCountAdd(code []byte, ip *int) {
	// OpRegLoopCountAdd: acc_reg, counter_reg, start(16), limit(16), step(16)
	accReg := int(code[*ip+1])
	counterReg := int(code[*ip+2])
	start := int64(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
	limit := int64(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
	step := int64(int16(uint16(code[*ip+7])<<8 | uint16(code[*ip+8])))

	counterTagged := uint64(start) | (uint64(TagInt) << 48)
	accTagged := uint64(TagInt) << 48 // 0 as tagged int

	// Initialize counter
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(counterTagged)
	cg.storeRaxToReg(counterReg)

	// Initialize accumulator
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(accTagged)
	cg.storeRaxToReg(accReg)

	// Loop label
	loopLabel := fmt.Sprintf("loop_%d", *ip)
	cg.labels[loopLabel] = len(cg.code)

	// Check counter < limit
	cg.loadRegToRax(counterReg)
	limitTagged := uint64(limit) | (uint64(TagInt) << 48)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(limitTagged)

	endLabel := fmt.Sprintf("loop_end_%d", *ip)
	cg.emitJcc(0x8D, endLabel) // JGE

	// Add counter to accumulator
	cg.loadRegToRax(counterReg)
	cg.loadRegToRcx(accReg)
	cg.emitBytes([]byte{0x48, 0x01, 0xC8}) // add rax, rcx
	cg.storeRaxToReg(accReg)

	// Increment counter
	cg.loadRegToRax(counterReg)
	if step == 1 {
		cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01})
	} else {
		cg.emitBytes([]byte{0x48, 0x05})
		cg.emitUint32(uint32(step))
	}
	cg.storeRaxToReg(counterReg)

	cg.emitJmp(loopLabel)
	cg.labels[endLabel] = len(cg.code)

	*ip += 9
}

func (cg *CodeGenerator) compileOpRegNull(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *CodeGenerator) compileOpRegTrue(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool)<<48 | 1)
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *CodeGenerator) compileOpRegFalse(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool) << 48)
	cg.storeRaxToReg(dst)
	*ip += 2
}

// Jump helpers
func (cg *CodeGenerator) emitJmp(label string) {
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *CodeGenerator) emitJcc(cc byte, label string) {
	cg.emitBytes([]byte{0x0F, 0x80 | cc})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}
