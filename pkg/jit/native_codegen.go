//go:build amd64 && !windows
// +build amd64,!windows

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
// Globals are passed as first argument (in rdi per x86-64 calling convention)
type NativeCodeGenerator struct {
	code      []byte
	labels    map[string]int
	fixups    []fixup
	constants []int64
	globals   []int64 // Pre-extracted global values for initialization

	// Function info for local variable layout
	numParams int
	numLocals int

	// Function entry point (after prologue) for tail calls
	funcEntry int

	// Callback pointer for builtin/function dispatch
	builtinCallbackPtr    uintptr
	functionCallbackPtr   uintptr
	collectionCallbackPtr uintptr
	objectCallbackPtr     uintptr

	// Register allocation
	// We use: rax(0), rbx(1), rcx(2), rdx(3), r8(4), r9(5), r10(6), r11(7)
	// r12-r15 are callee-saved, used for locals 8-11
	// Stack is used for locals 12+
	// rdi holds globals pointer (first argument)
}

// NewNativeCodeGenerator creates a new native code generator
func NewNativeCodeGenerator() *NativeCodeGenerator {
	return &NativeCodeGenerator{
		code:      make([]byte, 0, 4096),
		labels:    make(map[string]int),
		fixups:    make([]fixup, 0),
		constants: nil,
	}
}

// NewNativeCodeGeneratorWithCallbacks creates a code generator with callback pointers set up
func NewNativeCodeGeneratorWithCallbacks() *NativeCodeGenerator {
	cg := NewNativeCodeGenerator()
	cg.builtinCallbackPtr = GetBuiltinCallbackPtr()
	cg.functionCallbackPtr = GetFunctionCallbackPtr()
	cg.collectionCallbackPtr = GetCollectionCallbackPtr()
	cg.objectCallbackPtr = GetObjectCallbackPtr()
	return cg
}

// SetBuiltinCallback sets the callback pointer for builtin calls
func (cg *NativeCodeGenerator) SetBuiltinCallback(ptr uintptr) {
	cg.builtinCallbackPtr = ptr
}

// SetFunctionCallback sets the callback pointer for function calls
func (cg *NativeCodeGenerator) SetFunctionCallback(ptr uintptr) {
	cg.functionCallbackPtr = ptr
}

// SetCollectionCallback sets the callback pointer for collection operations
func (cg *NativeCodeGenerator) SetCollectionCallback(ptr uintptr) {
	cg.collectionCallbackPtr = ptr
}

// SetObjectCallback sets the callback pointer for object operations
func (cg *NativeCodeGenerator) SetObjectCallback(ptr uintptr) {
	cg.objectCallbackPtr = ptr
}

// Generate generates native x86-64 code
func (cg *NativeCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []int64) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.numParams = fn.NumParameters
	cg.numLocals = fn.NumLocals

	// Generate prologue
	cg.emitPrologue(fn.NumLocals)

	// Store function entry point (after prologue) for tail calls
	cg.funcEntry = len(cg.code)
	cg.labels["__func_entry__"] = cg.funcEntry

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

	// Always add epilogue at the end to ensure proper return
	// The value to return should already be in rax (set by OpRegMove dst=255)
	cg.emitEpilogue()

	return cg.code, nil
}

// emitPrologue generates function entry code
// Note: callee-saved registers are already saved by the bridge function (callNative)
// So we only need to set up the stack frame
// Stack layout:
//
//	[rbp] = old rbp
//	[rbp-8..rbp-48] = spilled registers space
//	[rbp-48..rbp-48-numLocals*8] = local variables
//
// Parameters are passed in: RAX (arg0), RBX (arg1), RCX (arg2), RDX (arg3), R8-R9 (arg4-5)
// The bytecode compiler will emit OpRegStoreLocal to copy parameters to Locals
func (cg *NativeCodeGenerator) emitPrologue(numLocals int) {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Allocate stack space for spilled registers (256 bytes) + locals
	stackSize := 256 + numLocals*8
	// Round up to 16 bytes for alignment
	if stackSize%16 != 0 {
		stackSize += 16 - (stackSize % 16)
	}
	// sub rsp, stackSize
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(uint32(stackSize))

	// Parameters are already in RAX, RBX, RCX, etc.
	// The bytecode will emit OpRegStoreLocal to copy them to Locals
	// We don't need to do anything here - the bytecode will handle it
}

// emitEpilogue generates function exit code
// Note: callee-saved registers are restored by the bridge function
func (cg *NativeCodeGenerator) emitEpilogue() {
	// Calculate stack size (must match prologue)
	stackSize := 256 + cg.numLocals*8
	if stackSize%16 != 0 {
		stackSize += 16 - (stackSize % 16)
	}
	// add rsp, stackSize
	cg.emitBytes([]byte{0x48, 0x81, 0xC4})
	cg.emitUint32(uint32(stackSize))
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
		constIdx := int(code[*ip+2])<<8 | int(code[*ip+3]) // big-endian
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

	case compiler.OpRegLoadGlobal:
		// R[dst] = Globals[idx]
		// Globals pointer is in rdi (first argument)
		dst := int(code[*ip+1])
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileLoadGlobal(dst, globalIdx)
		*ip += 4

	case compiler.OpRegStoreGlobal:
		// Globals[idx] = R[src]
		src := int(code[*ip+1])
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileStoreGlobal(src, globalIdx)
		*ip += 4

	case compiler.OpRegLoadLocal:
		// R[dst] = Locals[idx]
		dst := int(code[*ip+1])
		localIdx := int(code[*ip+2])
		cg.compileLoadLocal(dst, localIdx)
		*ip += 3

	case compiler.OpRegStoreLocal:
		// Locals[idx] = R[src]
		src := int(code[*ip+1])
		localIdx := int(code[*ip+2])
		cg.compileStoreLocal(src, localIdx)
		*ip += 3

	case compiler.OpRegTailCall:
		// Tail call: jump to function entry after moving args
		funcReg := int(code[*ip+1])
		numArgs := int(code[*ip+2])
		cg.compileTailCall(funcReg, numArgs)
		*ip += 3

	case compiler.OpRegCall:
		// Function call: call function in register with args
		funcReg := int(code[*ip+1])
		numArgs := int(code[*ip+2])
		cg.compileCall(funcReg, numArgs)
		*ip += 3

	case compiler.OpRegBuiltin:
		// Builtin call: call builtin by auto-assigned index with args in R0-R7
		builtinIdx := int(code[*ip+1])<<8 | int(code[*ip+2])
		numArgs := int(code[*ip+3])
		cg.compileBuiltin(builtinIdx, numArgs)
		*ip += 4

	case compiler.OpRegArray:
		// Create array: R[dst] = Array from R[start..start+count-1]
		dst := int(code[*ip+1])
		startReg := int(code[*ip+2])
		count := int(code[*ip+3])
		cg.compileArrayCreate(dst, startReg, count)
		*ip += 4

	case compiler.OpRegArrayEmpty:
		// Create empty array: R[dst] = []
		dst := int(code[*ip+1])
		cg.compileCollectionOp(OpArrayEmpty, dst, 0, nil)
		*ip += 2

	case compiler.OpRegArrayAppend:
		// Append to array: R[dst] = append(R[arr], R[elem])
		dst := int(code[*ip+1])
		arrReg := int(code[*ip+2])
		elemReg := int(code[*ip+3])
		cg.compileCollectionOp(OpArrayAppend, dst, 2, []int{arrReg, elemReg})
		*ip += 4

	case compiler.OpRegIndex:
		// Get element: R[dst] = R[obj][R[key]]
		dst := int(code[*ip+1])
		objReg := int(code[*ip+2])
		keyReg := int(code[*ip+3])
		cg.compileCollectionOp(OpArrayGet, dst, 2, []int{objReg, keyReg})
		*ip += 4

	case compiler.OpRegSetIndex:
		// Set element: R[obj][R[key]] = R[val]
		objReg := int(code[*ip+1])
		keyReg := int(code[*ip+2])
		valReg := int(code[*ip+3])
		cg.compileCollectionOp(OpArraySet, 0, 3, []int{objReg, keyReg, valReg})
		*ip += 4

	case compiler.OpRegMap:
		// Create map: R[dst] = Map from pairs starting at R[start]
		dst := int(code[*ip+1])
		startReg := int(code[*ip+2])
		count := int(code[*ip+3])
		cg.compileMapCreate(dst, startReg, count)
		*ip += 4

	case compiler.OpRegMapEmpty:
		// Create empty map: R[dst] = {}
		dst := int(code[*ip+1])
		cg.compileCollectionOp(OpMapEmpty, dst, 0, nil)
		*ip += 2

	case compiler.OpRegMapSet:
		// Set key-value: R[dst] = R[map] with R[key] = R[val]
		dst := int(code[*ip+1])
		mapReg := int(code[*ip+2])
		keyReg := int(code[*ip+3])
		valReg := int(code[*ip+4])
		cg.compileCollectionOp(OpMapSet, dst, 3, []int{mapReg, keyReg, valReg})
		*ip += 5

	case compiler.OpRegGetField:
		// Get field: R[dst] = R[obj].field(name_idx)
		dst := int(code[*ip+1])
		objReg := int(code[*ip+2])
		nameIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileObjectOp(OpGetField, dst, []int{objReg}, nameIdx)
		*ip += 5

	case compiler.OpRegSetField:
		// Set field: R[obj].field(name_idx) = R[val]
		objReg := int(code[*ip+1])
		valReg := int(code[*ip+2])
		nameIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileObjectOp(OpSetField, 0, []int{objReg, valReg}, nameIdx)
		*ip += 5

	case compiler.OpRegGetMethod:
		// Get method: R[dst] = R[obj].method(name_idx)
		dst := int(code[*ip+1])
		objReg := int(code[*ip+2])
		nameIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileObjectOp(OpGetMethod, dst, []int{objReg}, nameIdx)
		*ip += 5

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
		// Standard registers: rax, rbx, rcx, rdx, r8-r11
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
	} else if r < 12 {
		// Callee-saved registers: r12-r15 for VM regs 8-11
		switch r {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE0}) // mov rax, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE8}) // mov rax, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF0}) // mov rax, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF8}) // mov rax, r15
		}
	} else {
		// Spilled to stack (reg 12+)
		// Stack layout after prologue:
		// [rbp] = old rbp
		// [rbp-8] to [rbp-256] = spilled registers (regs 12-43)
		// [rbp-256-...] = local variables
		// Note: displacement is signed, so we emit -offset as uint32
		offset := 48 + (r-12)*8
		cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
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
	} else if r < 12 {
		// Callee-saved registers: r12-r15 for VM regs 8-11
		switch r {
		case 8:
			cg.emitBytes([]byte{0x49, 0x89, 0xC4}) // mov r12, rax
		case 9:
			cg.emitBytes([]byte{0x49, 0x89, 0xC5}) // mov r13, rax
		case 10:
			cg.emitBytes([]byte{0x49, 0x89, 0xC6}) // mov r14, rax
		case 11:
			cg.emitBytes([]byte{0x49, 0x89, 0xC7}) // mov r15, rax
		}
	} else {
		// Spilled to stack (reg 12+)
		// Note: displacement is signed, so we emit -offset as uint32
		offset := 48 + (r-12)*8
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(uint32(-offset))
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

// compileLoadGlobal loads a global variable
// Globals pointer is in rdi (first argument per x86-64 calling convention)
// Global variables are int64, so offset = globalIdx * 8
func (cg *NativeCodeGenerator) compileLoadGlobal(dst, globalIdx int) {
	// mov rax, [rdi + globalIdx*8]
	// Using SIB addressing: [rdi + disp32] or [rdi + idx*8]
	offset := globalIdx * 8

	if offset < 128 {
		// Short form: mov rax, [rdi + disp8]
		cg.emitBytes([]byte{0x48, 0x8B, 0x47}) // mov rax, [rdi + disp8]
		cg.emitByte(byte(offset))
	} else {
		// Long form: mov rax, [rdi + disp32]
		cg.emitBytes([]byte{0x48, 0x8B, 0x87}) // mov rax, [rdi + disp32]
		cg.emitUint32(uint32(offset))
	}

	// Store to destination register
	cg.storeRaxToReg(dst)
}

// compileStoreGlobal stores a value to a global variable
// Globals pointer is in rdi
func (cg *NativeCodeGenerator) compileStoreGlobal(src, globalIdx int) {
	// Load source to rax
	cg.loadRegToRax(src)

	// mov [rdi + offset], rax
	offset := globalIdx * 8

	if offset < 128 {
		// Short form: mov [rdi + disp8], rax
		cg.emitBytes([]byte{0x48, 0x89, 0x47}) // mov [rdi + disp8], rax
		cg.emitByte(byte(offset))
	} else {
		// Long form: mov [rdi + disp32], rax
		cg.emitBytes([]byte{0x48, 0x89, 0x87}) // mov [rdi + disp32], rax
		cg.emitUint32(uint32(offset))
	}
}

// compileLoadLocal loads a local variable to a register
// Local variables are stored on stack: [rbp - 256 - (localIdx+1)*8]
// The 256 bytes is for spilled registers
func (cg *NativeCodeGenerator) compileLoadLocal(dst, localIdx int) {
	// Local offset: after spilled register space (256 bytes)
	offset := 256 + (localIdx+1)*8

	// mov rax, [rbp - offset]
	// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp + disp32]
	cg.emitUint32(uint32(-offset))

	// Store to destination register
	cg.storeRaxToReg(dst)
}

// compileStoreLocal stores a register value to a local variable
func (cg *NativeCodeGenerator) compileStoreLocal(src, localIdx int) {
	// Load source to rax
	cg.loadRegToRax(src)

	// Local offset: after spilled register space (256 bytes)
	offset := 256 + (localIdx+1)*8

	// mov [rbp - offset], rax
	// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
	cg.emitUint32(uint32(-offset))
}

// compileTailCall handles tail calls
// For self-recursive tail calls, we jump back to function entry
// For cross-function tail calls, we use the callback mechanism
func (cg *NativeCodeGenerator) compileTailCall(funcReg, numArgs int) {
	// Tail call optimization: instead of calling and returning,
	// we jump directly to the target function's entry point.
	// This saves stack space and enables efficient recursion.

	// For self-recursive tail calls (funcReg typically < 8 and points to self):
	// Jump back to function entry point after moving arguments.

	// For cross-function tail calls:
	// We need to dispatch via callback since we don't know the target at compile time.

	// Check if we have a function callback pointer
	if cg.functionCallbackPtr != 0 {
		// Use callback for tail call dispatch
		// This is safer and handles both self-recursive and cross-function calls

		// Push globals pointer (rdi) to stack
		cg.emitBytes([]byte{0x57}) // push rdi

		// Spill args to stack for callback
		baseOffset := int32(640) // Use different offset than compileCall

		// Spill R0 (RAX) to [rbp - baseOffset]
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(uint32(-baseOffset))

		if numArgs >= 2 {
			cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp + disp32], rbx
			cg.emitUint32(uint32(-(baseOffset + 8)))
		}
		if numArgs >= 3 {
			cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp + disp32], rcx
			cg.emitUint32(uint32(-(baseOffset + 16)))
		}
		if numArgs >= 4 {
			cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp + disp32], rdx
			cg.emitUint32(uint32(-(baseOffset + 24)))
		}
		if numArgs >= 5 {
			cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp + disp32], r8
			cg.emitUint32(uint32(-(baseOffset + 32)))
		}
		if numArgs >= 6 {
			cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp + disp32], r9
			cg.emitUint32(uint32(-(baseOffset + 40)))
		}
		if numArgs >= 7 {
			cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp + disp32], r10
			cg.emitUint32(uint32(-(baseOffset + 48)))
		}
		if numArgs >= 8 {
			cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp + disp32], r11
			cg.emitUint32(uint32(-(baseOffset + 56)))
		}

		// Set up callback arguments (System V AMD64 ABI):
		//   rdi = funcReg
		//   rsi = numArgs
		//   rdx = args pointer
		cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
		cg.emitUint32(uint32(funcReg))

		cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
		cg.emitUint32(uint32(numArgs))

		cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp + disp32]
		cg.emitUint32(uint32(-baseOffset))

		// Call the function callback
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.functionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

		// Result is in RAX - this is our return value
		// Restore globals pointer and return
		cg.emitBytes([]byte{0x5F}) // pop rdi

		// Emit epilogue and return
		cg.emitEpilogue()
	} else {
		// Fallback: for self-recursive calls, jump to function entry
		// This assumes the function is self-recursive, which is common for tail calls
		// WARNING: This will infinite loop for cross-function tail calls!

		// Move arguments to proper registers (they should already be in place)
		// Then jump to function entry

		// Calculate relative offset to function entry
		if cg.funcEntry > 0 {
			// jmp to function entry
			offset := int32(cg.funcEntry - (len(cg.code) + 5))
			cg.emitBytes([]byte{0xE9}) // jmp rel32
			cg.emitUint32(uint32(offset))
		} else {
			// No entry point saved, just return
			cg.emitEpilogue()
		}
	}
}

// compileCall handles function calls by calling back to Go
// This allows native code to call functions (including natively-compiled ones)
// The function callback signature: callback(funcReg, numArgs, argsPtr) int64
// Arguments are in VM registers R0-R7 (which map to RAX, RBX, RCX, RDX, R8-R11)
func (cg *NativeCodeGenerator) compileCall(funcReg, numArgs int) {
	// For the callback, we use System V AMD64 ABI:
	//   rdi = funcReg
	//   rsi = numArgs
	//   rdx = pointer to args array (on stack)

	// But rdi currently holds globals pointer! We need to save it.
	// Push globals pointer (rdi) to stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Store args to stack for args pointer
	// We'll spill R0-R7 to [rbp - 320 - argNum*8] (below the existing spilled regs space)
	// This creates a contiguous array of int64 values
	baseOffset := int32(400) // Start after builtin callback space

	// Spill R0 (RAX) to [rbp - baseOffset]
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		// Spill R1 (RBX)
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp + disp32], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		// Spill R2 (RCX)
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp + disp32], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		// Spill R3 (RDX)
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp + disp32], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		// Spill R4 (R8)
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp + disp32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		// Spill R5 (R9)
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp + disp32], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		// Spill R6 (R10)
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp + disp32], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		// Spill R7 (R11)
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp + disp32], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Set up callback arguments (System V AMD64 ABI):
	//   rdi = funcReg
	//   rsi = numArgs
	//   rdx = args pointer (address of [rbp - baseOffset])

	// mov rdi, funcReg
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(funcReg))

	// mov rsi, numArgs
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
	cg.emitUint32(uint32(numArgs))

	// lea rdx, [rbp - baseOffset] (args pointer)
	cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the function callback
	if cg.functionCallbackPtr != 0 {
		// Direct call to callback pointer
		// mov rax, callback_ptr
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.functionCallbackPtr))
		// call rax
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		// Fallback: use indirect call through global
		// For now, emit a call with placeholder that will be patched
		cg.emitBytes([]byte{0xFF, 0x15, 0x00, 0x00, 0x00, 0x00}) // call [rip+0]
	}

	// Result is in RAX
	// Store to return register (R255)
	cg.storeRaxToReg(255)

	// Restore globals pointer
	cg.emitBytes([]byte{0x5F}) // pop rdi
}

// compileBuiltin generates code to call a builtin function via callback
// The callback signature: callback(builtinIdx, numArgs, argsPtr) int64
// builtinIdx is the auto-assigned index from objects.BuiltinIndexMap.
// Arguments are in VM registers R0-R7 (which map to RAX, RBX, RCX, RDX, R8-R11)
func (cg *NativeCodeGenerator) compileBuiltin(builtinIdx, numArgs int) {
	// Save callee-saved registers we need to preserve
	// We need to save RAX, RBX, RCX, RDX, R8, R9, R10, R11 (args) temporarily

	// For the callback, we use System V AMD64 ABI:
	//   rdi = nameConstIdx (constant pool index for builtin name)
	//   rsi = numArgs
	//   rdx = pointer to args array (on stack)

	// But rdi currently holds globals pointer! We need to save it.
	// Push globals pointer (rdi) to stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Store args to stack for args pointer
	// We'll spill R0-R7 to [rbp - 320 - argNum*8] (below the existing spilled regs space)
	// This creates a contiguous array of int64 values

	// First, spill the arguments to the stack
	// Note: we use negative displacement (as uint32 of signed value)
	// [rbp - 320] = R0, [rbp - 328] = R1, etc.
	baseOffset := int32(320) // Start after the 256-byte spilled regs + some buffer

	// Spill R0 (RAX) to [rbp - baseOffset]
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		// Spill R1 (RBX)
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp + disp32], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		// Spill R2 (RCX)
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp + disp32], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		// Spill R3 (RDX)
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp + disp32], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		// Spill R4 (R8)
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp + disp32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		// Spill R5 (R9)
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp + disp32], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		// Spill R6 (R10)
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp + disp32], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		// Spill R7 (R11)
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp + disp32], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Set up callback arguments (System V AMD64 ABI):
	//   rdi = builtinIdx (auto-assigned index)
	//   rsi = numArgs
	//   rdx = args pointer (address of [rbp - baseOffset])

	// mov rdi, builtinIdx
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(builtinIdx))

	// mov rsi, numArgs
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
	cg.emitUint32(uint32(numArgs))

	// lea rdx, [rbp - baseOffset] (args pointer)
	cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the builtin callback
	if cg.builtinCallbackPtr != 0 {
		// Direct call to callback pointer
		// mov rax, callback_ptr
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.builtinCallbackPtr))
		// call rax
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		// Fallback: use indirect call through global
		// For now, emit a call with placeholder that will be patched
		cg.emitBytes([]byte{0xFF, 0x15, 0x00, 0x00, 0x00, 0x00}) // call [rip+0]
	}

	// Result is in RAX
	// Store to return register (R255)
	cg.storeRaxToReg(255)

	// Restore globals pointer
	cg.emitBytes([]byte{0x5F}) // pop rdi
}

// CollectionOpKind indicates what collection operation to perform (must match native_executor.go)
type CollectionOpKind int

const (
	OpArrayCreate CollectionOpKind = iota
	OpArrayEmpty
	OpArrayAppend
	OpArrayGet
	OpArraySet
	OpMapCreate
	OpMapEmpty
	OpMapSet
	OpMapGet
)

// ObjectOpKind indicates what object operation to perform
type ObjectOpKind int

const (
	OpGetField ObjectOpKind = iota
	OpSetField
	OpGetMethod
)

// compileCollectionOp generates code to perform a collection operation via callback
func (cg *NativeCodeGenerator) compileCollectionOp(opKind CollectionOpKind, dstReg int, numArgs int, argRegs []int) {
	// Push globals pointer (rdi) to stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Spill args to stack
	baseOffset := int32(480) // After function callback space

	// Spill argument registers to stack
	for i, reg := range argRegs {
		if i >= 8 {
			break // Max 8 args
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	// Set up callback arguments:
	//   rdi = opKind
	//   rsi = numArgs
	//   rdx = args pointer

	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
	cg.emitUint32(uint32(numArgs))

	cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the collection callback
	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitBytes([]byte{0xFF, 0x15, 0x00, 0x00, 0x00, 0x00})
	}

	// Store result to destination register
	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	// Restore globals pointer
	cg.emitBytes([]byte{0x5F}) // pop rdi
}

// compileArrayCreate generates code to create an array from registers
func (cg *NativeCodeGenerator) compileArrayCreate(dst, startReg, count int) {
	// Build argRegs slice
	argRegs := make([]int, count)
	for i := 0; i < count; i++ {
		argRegs[i] = startReg + i
	}

	// Call collection op with OpArrayCreate
	// Result (object handle) will be stored in dst
	cg.compileCollectionOpDirect(OpArrayCreate, dst, count, argRegs)
}

// compileMapCreate generates code to create a map from key-value pairs
func (cg *NativeCodeGenerator) compileMapCreate(dst, startReg, count int) {
	// count is number of pairs, so total args = count * 2
	numArgs := count * 2
	argRegs := make([]int, numArgs)
	for i := 0; i < numArgs; i++ {
		argRegs[i] = startReg + i
	}

	cg.compileCollectionOpDirect(OpMapCreate, dst, numArgs, argRegs)
}

// compileCollectionOpDirect generates code for collection operations with direct arg passing
func (cg *NativeCodeGenerator) compileCollectionOpDirect(opKind CollectionOpKind, dstReg int, numArgs int, argRegs []int) {
	// Push globals pointer (rdi) to stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Spill args to stack
	baseOffset := int32(480)

	for i, reg := range argRegs {
		if i >= 8 {
			break
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, opKind
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, numArgs
	cg.emitUint32(uint32(numArgs))

	cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp - baseOffset]
	cg.emitUint32(uint32(-baseOffset))

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0})
	} else {
		cg.emitBytes([]byte{0xFF, 0x15, 0x00, 0x00, 0x00, 0x00})
	}

	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	cg.emitBytes([]byte{0x5F}) // pop rdi
}

// compileObjectOp generates code to perform object field operations via callback
func (cg *NativeCodeGenerator) compileObjectOp(opKind ObjectOpKind, dstReg int, argRegs []int, nameIdx int) {
	// Push globals pointer (rdi) to stack
	cg.emitBytes([]byte{0x57}) // push rdi

	// Spill args to stack
	baseOffset := int32(560) // After collection callback space

	for i, reg := range argRegs {
		if i >= 8 {
			break
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	// Set up callback arguments:
	//   rdi = opKind
	//   rsi = numArgs
	//   rdx = args pointer
	//   rcx = nameIdx

	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, opKind
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, numArgs
	cg.emitUint32(uint32(len(argRegs)))

	cg.emitBytes([]byte{0x48, 0x8D, 0x95}) // lea rdx, [rbp - baseOffset]
	cg.emitUint32(uint32(-baseOffset))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, nameIdx
	cg.emitUint32(uint32(nameIdx))

	// Call the object callback
	if cg.objectCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8})
		cg.emitUint64(uint64(cg.objectCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0})
	} else {
		cg.emitBytes([]byte{0xFF, 0x15, 0x00, 0x00, 0x00, 0x00})
	}

	// Store result to destination register
	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	cg.emitBytes([]byte{0x5F}) // pop rdi
}

func (cg *NativeCodeGenerator) compileMove(dst, src int) {
	if dst == src {
		return
	}

	// Special case: dst = 255 (ReturnRegister) means put value in rax for return
	if dst == 255 {
		cg.loadRegToRax(src)
		return
	}

	// Special case: src = 255 (ReturnRegister) means get value from rax
	if src == 255 {
		cg.storeRaxToReg(dst)
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
	} else if right < 12 {
		// VM regs 8-11 are in R12-R15 (callee-saved)
		// ADD r/m64, r64: REX.W 01 /r
		// REX prefix: W=1, R=1 (for R12-R15), X=0, B=0
		// REX = 0x4C (0100 1100)
		// ModR/M: mod=11 (register), reg=right (R12=4, R13=5, R14=6, R15=7), r/m=0 (rax)
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x01, 0xE0}) // add rax, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x01, 0xE8}) // add rax, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x01, 0xF0}) // add rax, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x01, 0xF8}) // add rax, r15
		}
	} else {
		// Add from stack (VM regs 12+)
		// Stack layout: [rbp - 48 - (right-12)*8]
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (right-12)*8
		cg.emitBytes([]byte{0x48, 0x03, 0x85}) // add rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
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
	} else if right < 12 {
		// VM regs 8-11 are in R12-R15
		// SUB r/m64, r64: REX.W 29 /r
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x29, 0xE0}) // sub rax, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x29, 0xE8}) // sub rax, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x29, 0xF0}) // sub rax, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x29, 0xF8}) // sub rax, r15
		}
	} else {
		// VM regs 12+ on stack
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (right-12)*8
		cg.emitBytes([]byte{0x48, 0x2B, 0x85}) // sub rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
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
	} else if right < 12 {
		// VM regs 8-11 are in R12-R15
		// IMUL r64, r/m64: REX.W 0F AF /r
		// NOTE: IMUL has reversed operand encoding compared to ADD/SUB!
		// IMUL format: r (dest in reg field), r/m (source in r/m field)
		// For imul rax, r12: RAX is dest (not extended), R12 is source (extended in r/m)
		// REX = 0x49 (W=1, B=1, R=0) - only B bit for source register extension
		switch right {
		case 8:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC4}) // imul rax, r12
		case 9:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC5}) // imul rax, r13
		case 10:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC6}) // imul rax, r14
		case 11:
			cg.emitBytes([]byte{0x49, 0x0F, 0xAF, 0xC7}) // imul rax, r15
		}
	} else {
		// VM regs 12+ on stack
		// Stack layout: [rbp - 48 - (right-12)*8]
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (right-12)*8
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x85}) // imul rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
	}

	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileDiv(dst, left, right int) {
	// Division in x86-64: idiv uses rdx:rax / src
	// Result: quotient in rax, remainder in rdx
	// SAFETY: Check for division by zero to avoid hardware exception

	// First, load divisor to rcx for zero check
	if right < 8 {
		if right == 2 {
			// Already in rcx, just test it
		} else {
			cg.loadRegToRax(right)
			cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
		}
	} else if right < 12 {
		// VM regs 8-11 are in R12-R15
		// MOV r/m, r: dest in r/m field, source in reg field
		// For mov rcx, r12: RCX is r/m=2 (no B), R12 is reg=4 with R
		// REX = 0x4C (W=1, R=1, B=0)
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE2}) // mov rcx, r12: REX=4C, modrm=E2
		case 9:
			cg.emitBytes([]byte{0x4C, 0x89, 0xEA}) // mov rcx, r13: REX=4C, modrm=EA
		case 10:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF2}) // mov rcx, r14: REX=4C, modrm=F2
		case 11:
			cg.emitBytes([]byte{0x4C, 0x89, 0xFA}) // mov rcx, r15: REX=4C, modrm=FA
		}
	} else {
		// Load from stack to rcx
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (right-12)*8
		cg.emitBytes([]byte{0x48, 0x8B, 0x8D}) // mov rcx, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
	}

	// Test if divisor is zero
	cg.emitBytes([]byte{0x48, 0x85, 0xC9}) // test rcx, rcx

	// jz to return_zero (placeholder)
	jzPos := len(cg.code)
	cg.emitBytes([]byte{0x74, 0x00}) // jz rel8

	// Load dividend
	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x99}) // cqo: sign-extend rax to rdx:rax

	// idiv rcx (ModRM=FA: mod=11, reg=7 for idiv, r/m=2 for RCX)
	cg.emitBytes([]byte{0x48, 0xF7, 0xFA})

	// Store result and jump over zero case
	cg.storeRaxToReg(dst)
	jmpPos := len(cg.code)
	cg.emitBytes([]byte{0xEB, 0x00}) // jmp rel8 (placeholder)

	// Return zero case
	returnZeroPos := len(cg.code)
	// Use safe jump offset with validation
	divOffset1 := returnZeroPos - (jzPos + 2)
	if !CanUseShortJump(divOffset1) {
		fmt.Printf("[JIT WARNING] Jump offset %d exceeds rel8 range in div/mod zero check\n", divOffset1)
	}
	cg.code[jzPos+1] = byte(int8(divOffset1))

	// Set result to 0
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.storeRaxToReg(dst)

	// Fix up jump over zero case
	endPos := len(cg.code)
	divOffset2 := endPos - (jmpPos + 2)
	if !CanUseShortJump(divOffset2) {
		fmt.Printf("[JIT WARNING] Jump offset %d exceeds rel8 range in div/mod skip\n", divOffset2)
	}
	cg.code[jmpPos+1] = byte(int8(divOffset2))
}

func (cg *NativeCodeGenerator) compileMod(dst, left, right int) {
	// Mod is the remainder after division
	// SAFETY: Check for division by zero to avoid hardware exception

	// First, load divisor to rcx for zero check
	if right < 8 {
		if right == 3 {
			// Already in rdx, move to rcx
			cg.emitBytes([]byte{0x48, 0x89, 0xD1}) // mov rcx, rdx
		} else {
			cg.loadRegToRax(right)
			cg.emitBytes([]byte{0x48, 0x89, 0xC1}) // mov rcx, rax
		}
	} else if right < 12 {
		// VM regs 8-11 are in R12-R15
		// MOV r/m, r: dest in r/m field, source in reg field
		// For mov rcx, r12: RCX is r/m=2 (no B), R12 is reg=4 with R
		// REX = 0x4C (W=1, R=1, B=0)
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE2}) // mov rcx, r12: REX=4C, modrm=E2
		case 9:
			cg.emitBytes([]byte{0x4C, 0x89, 0xEA}) // mov rcx, r13: REX=4C, modrm=EA
		case 10:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF2}) // mov rcx, r14: REX=4C, modrm=F2
		case 11:
			cg.emitBytes([]byte{0x4C, 0x89, 0xFA}) // mov rcx, r15: REX=4C, modrm=FA
		}
	} else {
		// Load from stack to rcx
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (right-12)*8
		cg.emitBytes([]byte{0x48, 0x8B, 0x8D}) // mov rcx, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
	}

	// Test if divisor is zero
	cg.emitBytes([]byte{0x48, 0x85, 0xC9}) // test rcx, rcx

	// jz to return_zero (placeholder)
	jzPos := len(cg.code)
	cg.emitBytes([]byte{0x74, 0x00}) // jz rel8

	// Load dividend
	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x99}) // cqo

	// idiv rcx (ModRM=FA: mod=11, reg=7 for idiv, r/m=2 for RCX)
	cg.emitBytes([]byte{0x48, 0xF7, 0xFA})

	// Result is in rdx (remainder)
	cg.emitBytes([]byte{0x48, 0x89, 0xD0}) // mov rax, rdx

	// Store result and jump over zero case
	cg.storeRaxToReg(dst)
	jmpPos := len(cg.code)
	cg.emitBytes([]byte{0xEB, 0x00}) // jmp rel8 (placeholder)

	// Return zero case
	returnZeroPos := len(cg.code)
	// Use safe jump offset with validation
	divOffset1 := returnZeroPos - (jzPos + 2)
	if !CanUseShortJump(divOffset1) {
		fmt.Printf("[JIT WARNING] Jump offset %d exceeds rel8 range in div/mod zero check\n", divOffset1)
	}
	cg.code[jzPos+1] = byte(int8(divOffset1))

	// Set result to 0
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.storeRaxToReg(dst)

	// Fix up jump over zero case
	endPos := len(cg.code)
	divOffset2 := endPos - (jmpPos + 2)
	if !CanUseShortJump(divOffset2) {
		fmt.Printf("[JIT WARNING] Jump offset %d exceeds rel8 range in div/mod skip\n", divOffset2)
	}
	cg.code[jmpPos+1] = byte(int8(divOffset2))
}

func (cg *NativeCodeGenerator) compileNeg(dst, src int) {
	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileCompare(dst, left, right int, op string) {
	// Special case: if right is in RAX (VM reg 0), we need to save left first
	// because loadRegToRax(left) will overwrite RAX
	if right == 0 && left != 0 {
		// Load left to RAX
		cg.loadRegToRax(left)
		// Move left to a temp register (R8)
		cg.emitBytes([]byte{0x49, 0x89, 0xC0}) // mov r8, rax
		// Load right (which is in VM reg 0 = RAX) to RAX
		// Actually right is in RAX already since VM reg 0 = RAX
		// Compare: left (now in R8) vs right (in RAX)
		cg.emitBytes([]byte{0x49, 0x39, 0xC0}) // cmp r8, rax
	} else if left == 0 && right != 0 {
		// Left is in RAX, right is somewhere else
		// RAX already contains left (VM reg 0)
		cg.compareRaxWithReg(right)
	} else if left == 0 && right == 0 {
		// Both in RAX - always equal
		cg.emitBytes([]byte{0x48, 0x39, 0xC0}) // cmp rax, rax
	} else {
		// Normal case: neither is in RAX, or left/right are different
		cg.loadRegToRax(left)
		cg.compareRaxWithReg(right)
	}

	// Set result based on comparison
	// IMPORTANT: setcc must come BEFORE xor rax,rax because xor affects flags!
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

	// Now zero out the upper bits of rax (al already has the result)
	cg.emitBytes([]byte{0x48, 0x0F, 0xB6, 0xC0}) // movzx rax, al

	cg.storeRaxToReg(dst)
}

// compareRaxWithReg compares RAX with a VM register
func (cg *NativeCodeGenerator) compareRaxWithReg(r int) {
	if r < 8 {
		switch r {
		case 0:
			cg.emitBytes([]byte{0x48, 0x39, 0xC0}) // cmp rax, rax
		case 1:
			cg.emitBytes([]byte{0x48, 0x39, 0xD8}) // cmp rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x39, 0xD0}) // cmp rax, rdx
		case 4:
			cg.emitBytes([]byte{0x4C, 0x39, 0xC0}) // cmp rax, r8 (REX.R for r8)
		case 5:
			cg.emitBytes([]byte{0x4C, 0x39, 0xC8}) // cmp rax, r9
		case 6:
			cg.emitBytes([]byte{0x4C, 0x39, 0xD0}) // cmp rax, r10
		case 7:
			cg.emitBytes([]byte{0x4C, 0x39, 0xD8}) // cmp rax, r11
		}
	} else if r < 12 {
		// VM regs 8-11 are in R12-R15
		// REX.R is needed for r12-r15, REX.B is NOT needed for rax
		// REX = 0x48 | 0x04 = 0x4C (W=1, R=1)
		switch r {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x39, 0xE0}) // cmp rax, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x39, 0xE8}) // cmp rax, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x39, 0xF0}) // cmp rax, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x39, 0xF8}) // cmp rax, r15
		}
	} else {
		// Spilled to stack
		// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
		offset := 48 + (r-12)*8
		cg.emitBytes([]byte{0x48, 0x3B, 0x85}) // cmp rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
	}
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
	cg.emitBytes([]byte{0x0F, 0x85})       // jnz rel32
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *NativeCodeGenerator) compileJumpIfFalse(cond, target int) {
	cg.loadRegToRax(cond)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	cg.emitBytes([]byte{0x0F, 0x84})       // jz rel32
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
	} else if reg < 12 {
		// VM regs 8-11 are in R12-R15 (callee-saved)
		switch reg {
		case 8:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC4}) // inc r12
		case 9:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC5}) // inc r13
		case 10:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC6}) // inc r14
		case 11:
			cg.emitBytes([]byte{0x49, 0xFF, 0xC7}) // inc r15
		}
	} else {
		// Spilled to stack (reg 12+)
		offset := 48 + (reg-12)*8
		cg.emitBytes([]byte{0x48, 0xFF, 0x85}) // inc qword [rbp + disp32]
		cg.emitUint32(uint32(-offset))
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
	} else if reg < 12 {
		// VM regs 8-11 are in R12-R15 (callee-saved)
		switch reg {
		case 8:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCC}) // dec r12
		case 9:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCD}) // dec r13
		case 10:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCE}) // dec r14
		case 11:
			cg.emitBytes([]byte{0x49, 0xFF, 0xCF}) // dec r15
		}
	} else {
		// Spilled to stack (reg 12+)
		offset := 48 + (reg-12)*8
		cg.emitBytes([]byte{0x48, 0xFF, 0x8D}) // dec qword [rbp + disp32]
		cg.emitUint32(uint32(-offset))
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
