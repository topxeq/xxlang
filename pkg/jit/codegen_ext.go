// pkg/jit/codegen_ext.go
// Extended x86-64 Code Generator with function call support
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// ExtendedCodeGenerator generates x86-64 machine code with function support
type ExtendedCodeGenerator struct {
	config JITConfig
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction
	numRegs   int

	// For function calls
	hasCalls    bool
	callTargets map[int]string // IP -> label for internal calls

	// Interpreter callback pointer
	interpCallPtr uintptr
}

// NewExtendedCodeGenerator creates a new extended code generator
func NewExtendedCodeGenerator(config JITConfig) *ExtendedCodeGenerator {
	return &ExtendedCodeGenerator{
		config:      config,
		code:        make([]byte, 0, 8192),
		labels:      make(map[string]int),
		fixups:      make([]fixup, 0),
		callTargets: make(map[int]string),
	}
}

// SetInterpreterCallback sets the interpreter callback function pointer
func (cg *ExtendedCodeGenerator) SetInterpreterCallback(ptr uintptr) {
	cg.interpCallPtr = ptr
}

// Generate generates x86-64 code from VM bytecode with extended support
func (cg *ExtendedCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn
	cg.hasCalls = false
	cg.callTargets = make(map[int]string)

	cg.numRegs = 64

	// First pass: identify function calls and labels
	cg.analyzeBytecode(fn.Instructions)

	// Generate function prologue
	cg.emitExtendedPrologue()

	// Main dispatch loop
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		cg.labels[fmt.Sprintf("ip_%d", ip)] = len(cg.code)

		handled, err := cg.compileOpcode(op, code, &ip)
		if !handled {
			if cg.config.Debug {
				fmt.Printf("[JIT] Unknown opcode %d at IP %d\n", op, ip)
			}
			return nil, fmt.Errorf("unsupported opcode %d (%s) at IP %d", op, cg.opcodeName(op), ip)
		}
		if err != nil {
			return nil, err
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

// analyzeBytecode does a first pass to identify all labels and calls
func (cg *ExtendedCodeGenerator) analyzeBytecode(code []byte) {
	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegJump:
			offset := int(int16(uint16(code[ip+1])<<8 | uint16(code[ip+2])))
			target := ip + 3 + offset
			cg.labels[fmt.Sprintf("ip_%d", target)] = 0
			ip += 3

		case compiler.OpRegJumpIfFalse, compiler.OpRegJumpIfTrue:
			offset := int(int16(uint16(code[ip+2])<<8 | uint16(code[ip+3])))
			target := ip + 4 + offset
			cg.labels[fmt.Sprintf("ip_%d", target)] = 0
			ip += 4

		case compiler.OpRegCall:
			cg.hasCalls = true
			ip += 3

		case compiler.OpRegTailCall:
			cg.hasCalls = true
			ip += 3

		default:
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
}

// compileOpcode compiles a single opcode
func (cg *ExtendedCodeGenerator) compileOpcode(op compiler.Opcode, code []byte, ip *int) (bool, error) {
	switch op {
	// Data movement
	case compiler.OpRegLoadConst:
		cg.compileOpRegLoadConst(code, ip)
		return true, nil
	case compiler.OpRegMove:
		cg.compileOpRegMove(code, ip)
		return true, nil

	// Arithmetic
	case compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul:
		cg.compileOpRegArithmetic(code, ip, op)
		return true, nil
	case compiler.OpRegDiv:
		cg.compileOpRegDiv(code, ip)
		return true, nil
	case compiler.OpRegMod:
		cg.compileOpRegMod(code, ip)
		return true, nil
	case compiler.OpRegNeg:
		cg.compileOpRegNeg(code, ip)
		return true, nil

	// Comparison
	case compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegEqual:
		cg.compileOpRegComparison(code, ip, op)
		return true, nil
	case compiler.OpRegLessEqual, compiler.OpRegGreaterEqual:
		cg.compileOpRegComparisonEqual(code, ip, op)
		return true, nil
	case compiler.OpRegNotEqual:
		cg.compileOpRegNotEqual(code, ip)
		return true, nil

	// Logical
	case compiler.OpRegNot:
		cg.compileOpRegNot(code, ip)
		return true, nil

	// Control flow
	case compiler.OpRegJump:
		cg.compileOpRegJump(code, ip)
		return true, nil
	case compiler.OpRegJumpIfFalse:
		cg.compileOpRegJumpIfFalse(code, ip)
		return true, nil
	case compiler.OpRegJumpIfTrue:
		cg.compileOpRegJumpIfTrue(code, ip)
		return true, nil
	case compiler.OpRegReturn:
		cg.compileOpRegReturn(code, ip)
		return true, nil

	// Local variables
	case compiler.OpRegLoadLocal:
		cg.compileOpRegLoadLocal(code, ip)
		return true, nil
	case compiler.OpRegStoreLocal:
		cg.compileOpRegStoreLocal(code, ip)
		return true, nil

	// Global variables
	case compiler.OpRegLoadGlobal:
		cg.compileOpRegLoadGlobal(code, ip)
		return true, nil
	case compiler.OpRegStoreGlobal:
		cg.compileOpRegStoreGlobal(code, ip)
		return true, nil

	// Increment/Decrement
	case compiler.OpRegIncLocal:
		cg.compileOpRegIncLocal(code, ip)
		return true, nil
	case compiler.OpRegDecLocal:
		cg.compileOpRegDecLocal(code, ip)
		return true, nil

	// Loop superinstructions
	case compiler.OpRegLoopBodyAdd:
		cg.compileOpRegLoopBodyAdd(code, ip)
		return true, nil
	case compiler.OpRegLoopCountAdd:
		cg.compileOpRegLoopCountAdd(code, ip)
		return true, nil

	// Null/True/False
	case compiler.OpRegNull:
		cg.compileOpRegNull(code, ip)
		return true, nil
	case compiler.OpRegTrue:
		cg.compileOpRegTrue(code, ip)
		return true, nil
	case compiler.OpRegFalse:
		cg.compileOpRegFalse(code, ip)
		return true, nil

	// Stack operations
	case compiler.OpRegPush:
		cg.compileOpRegPush(code, ip)
		return true, nil
	case compiler.OpRegPop:
		cg.compileOpRegPop(code, ip)
		return true, nil

	// Closure
	case compiler.OpRegClosure:
		cg.compileOpRegClosure(code, ip)
		return true, nil

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++
		return true, nil

	// Function calls - these are complex, we'll emit a fallback
	case compiler.OpRegCall:
		return false, fmt.Errorf("OpRegCall not yet supported in JIT - use interpreter")

	case compiler.OpRegTailCall:
		return false, fmt.Errorf("OpRegTailCall not yet supported in JIT - use interpreter")

	case compiler.OpRegCallMethod:
		return false, fmt.Errorf("OpRegCallMethod not yet supported in JIT - use interpreter")
	}

	return false, nil
}

// emitExtendedPrologue generates extended function entry code
func (cg *ExtendedCodeGenerator) emitExtendedPrologue() {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Save callee-saved registers
	cg.emitByte(0x53)                    // push rbx
	cg.emitBytes([]byte{0x41, 0x54})     // push r12
	cg.emitBytes([]byte{0x41, 0x55})     // push r13
	cg.emitBytes([]byte{0x41, 0x56})     // push r14
	cg.emitBytes([]byte{0x41, 0x57})     // push r15

	// Allocate stack space for registers (64 registers * 8 bytes = 512 bytes)
	// Plus extra for temp stack operations
	stackSize := uint32(1024)
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(stackSize)

	// Initialize all registers to null
	nullValue := uint64(TagNull) << 48
	for i := 0; i < 64; i++ {
		off := int32(8 * (i + 1))
		cg.emitBytes([]byte{0x48, 0xC7, 0x45})
		cg.emitByte(byte(-off))
		cg.emitUint32(uint32(nullValue))
	}
}

// emitEpilogue generates function exit code
func (cg *ExtendedCodeGenerator) emitEpilogue() {
	// add rsp, stack_size
	cg.emitBytes([]byte{0x48, 0x81, 0xC4})
	cg.emitUint32(1024)

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F})     // pop r15
	cg.emitBytes([]byte{0x41, 0x5E})     // pop r14
	cg.emitBytes([]byte{0x41, 0x5D})     // pop r13
	cg.emitBytes([]byte{0x41, 0x5C})     // pop r12
	cg.emitByte(0x5B)                    // pop rbx
	cg.emitByte(0x5D)                    // pop rbp
	cg.emitByte(0xC3)                    // ret
}

// Stack offset helpers
func regOffsetExt(reg int) int32 {
	return int32(8 * (reg + 1))
}

func (cg *ExtendedCodeGenerator) loadRegToRax(reg int) {
	off := regOffsetExt(reg)
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *ExtendedCodeGenerator) storeRaxToReg(reg int) {
	off := regOffsetExt(reg)
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *ExtendedCodeGenerator) loadRegToRcx(reg int) {
	off := regOffsetExt(reg)
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-off))
}

func (cg *ExtendedCodeGenerator) loadRegToRdx(reg int) {
	off := regOffsetExt(reg)
	cg.emitBytes([]byte{0x48, 0x8B, 0x55})
	cg.emitByte(byte(-off))
}

// Byte emission helpers
func (cg *ExtendedCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *ExtendedCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *ExtendedCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *ExtendedCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// Jump helpers
func (cg *ExtendedCodeGenerator) emitJmp(label string) {
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *ExtendedCodeGenerator) emitJcc(cc byte, label string) {
	cg.emitBytes([]byte{0x0F, 0x80 | cc})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// opcodeName returns the name of an opcode
func (cg *ExtendedCodeGenerator) opcodeName(op compiler.Opcode) string {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return "UNKNOWN"
	}
	return def.Name
}

// ============================================================================
// Opcode compilation functions
// ============================================================================

func (cg *ExtendedCodeGenerator) compileOpRegLoadConst(code []byte, ip *int) {
	dst := int(code[*ip+1])
	constIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	var val uint64
	if constIdx < len(cg.constants) {
		val = uint64(cg.constants[constIdx])
	} else {
		val = uint64(TagNull) << 48
	}

	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(val)
	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegMove(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRegToRax(src)
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegArithmetic(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	switch op {
	case compiler.OpRegAdd:
		cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	case compiler.OpRegSub:
		cg.emitBytes([]byte{0x48, 0x29, 0xC8})
	case compiler.OpRegMul:
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1})
	}

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegDiv(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegMod(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegNeg(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8})
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegComparison(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})

	switch op {
	case compiler.OpRegLess:
		cg.emitBytes([]byte{0x0F, 0x9C, 0xC0})
	case compiler.OpRegGreater:
		cg.emitBytes([]byte{0x0F, 0x9F, 0xC0})
	case compiler.OpRegEqual:
		cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	}

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegComparisonEqual(code []byte, ip *int, op compiler.Opcode) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})

	switch op {
	case compiler.OpRegLessEqual:
		cg.emitBytes([]byte{0x0F, 0x9E, 0xC0})
	case compiler.OpRegGreaterEqual:
		cg.emitBytes([]byte{0x0F, 0x9D, 0xC0})
	}

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegNotEqual(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRegToRax(left)
	cg.loadRegToRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x95, 0xC0})

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegNot(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegJump(code []byte, ip *int) {
	offset := int(int16(uint16(code[*ip+1])<<8 | uint16(code[*ip+2])))
	target := *ip + 3 + offset

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJmp(label)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegJumpIfFalse(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRegToRax(cond)

	// Check if value is falsy (null or false)
	// We check the upper 16 bits for the tag
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48)

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label) // JE - jump if equal (false)
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegJumpIfTrue(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRegToRax(cond)

	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48 | 1) // true value

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label) // JE
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegReturn(code []byte, ip *int) {
	src := int(code[*ip+1])
	// Move return value to R0 (return register convention)
	cg.loadRegToRax(src)
	cg.storeRaxToReg(compiler.ReturnRegister)
	cg.emitEpilogue()
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegLoadLocal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	localIdx := int(code[*ip+2])

	cg.loadRegToRax(localIdx)
	cg.storeRaxToReg(dst)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegStoreLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRegToRax(src)
	cg.storeRaxToReg(localIdx)
	*ip += 3
}

func (cg *ExtendedCodeGenerator) compileOpRegLoadGlobal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	if globalIdx < len(cg.globals) {
		val := cg.globals[globalIdx]
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(val))
		cg.storeRaxToReg(dst)
	} else {
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(TagNull) << 48)
		cg.storeRaxToReg(dst)
	}
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegStoreGlobal(code []byte, ip *int) {
	// For JIT, global stores are no-ops (globals are snapshots)
	_ = code[*ip+1]
	_ = code[*ip+2:]
	*ip += 4
}

func (cg *ExtendedCodeGenerator) compileOpRegIncLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	cg.loadRegToRax(localIdx)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01})
	cg.storeRaxToReg(localIdx)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegDecLocal(code []byte, ip *int) {
	localIdx := int(code[*ip+1])
	cg.loadRegToRax(localIdx)
	cg.emitBytes([]byte{0x48, 0x83, 0xE8, 0x01})
	cg.storeRaxToReg(localIdx)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegLoopBodyAdd(code []byte, ip *int) {
	accReg := int(code[*ip+1])
	counterReg := int(code[*ip+2])
	limit := int64(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
	jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))

	cg.loadRegToRax(counterReg)
	cg.loadRegToRcx(accReg)
	cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	cg.storeRaxToReg(accReg)

	cg.loadRegToRax(counterReg)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01})
	cg.storeRaxToReg(counterReg)

	limitTagged := uint64(limit) | (uint64(TagInt) << 48)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(limitTagged)

	loopLabel := fmt.Sprintf("ip_%d", *ip+6+jumpOffset)
	cg.emitJcc(0x8C, loopLabel)

	*ip += 7
}

func (cg *ExtendedCodeGenerator) compileOpRegLoopCountAdd(code []byte, ip *int) {
	accReg := int(code[*ip+1])
	counterReg := int(code[*ip+2])
	start := int64(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
	limit := int64(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
	step := int64(int16(uint16(code[*ip+7])<<8 | uint16(code[*ip+8])))

	counterTagged := uint64(start) | (uint64(TagInt) << 48)
	accTagged := uint64(TagInt) << 48

	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(counterTagged)
	cg.storeRaxToReg(counterReg)

	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(accTagged)
	cg.storeRaxToReg(accReg)

	loopLabel := fmt.Sprintf("loop_%d", *ip)
	cg.labels[loopLabel] = len(cg.code)

	cg.loadRegToRax(counterReg)
	limitTagged := uint64(limit) | (uint64(TagInt) << 48)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(limitTagged)

	endLabel := fmt.Sprintf("loop_end_%d", *ip)
	cg.emitJcc(0x8D, endLabel)

	cg.loadRegToRax(counterReg)
	cg.loadRegToRcx(accReg)
	cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	cg.storeRaxToReg(accReg)

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

func (cg *ExtendedCodeGenerator) compileOpRegNull(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegTrue(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool)<<48 | 1)
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegFalse(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool) << 48)
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegPush(code []byte, ip *int) {
	src := int(code[*ip+1])
	// Push to temp stack (using rsp-relative location)
	cg.loadRegToRax(src)
	// mov [rsp + temp_offset], rax
	cg.emitBytes([]byte{0x48, 0x89, 0x04, 0x24}) // mov [rsp], rax
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegPop(code []byte, ip *int) {
	dst := int(code[*ip+1])
	// Pop from temp stack
	cg.emitBytes([]byte{0x48, 0x8B, 0x04, 0x24}) // mov rax, [rsp]
	cg.storeRaxToReg(dst)
	*ip += 2
}

func (cg *ExtendedCodeGenerator) compileOpRegClosure(code []byte, ip *int) {
	dst := int(code[*ip+1])
	// funcIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
	// numFree := int(code[*ip+4])
	// startReg := int(code[*ip+5])

	// For now, set to null as closures need interpreter support
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRaxToReg(dst)
	*ip += 6
}
