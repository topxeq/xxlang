//go:build amd64
// +build amd64

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
	numRegs   int // maximum VM register index + 1

	// Function entry point (after prologue) for tail calls
	funcEntry int

	// Compile-time tracking: maps VM register to the constant pool index
	// that was last loaded into it via OpRegLoadConst. Used by OpRegCall
	// to resolve the function's constant index at compile time.
	regConstMap map[int]int

	// Compile-time tracking: maps global index to the constant pool index
	// that was last stored into it. Propagated through StoreGlobal/LoadGlobal.
	globalConstMap map[int]int

	// Callback pointer for builtin/function dispatch
	builtinCallbackPtr    uintptr
	functionCallbackPtr   uintptr
	collectionCallbackPtr uintptr
	objectCallbackPtr     uintptr

	// syscallABI selects the calling convention for the JIT function entry.
	// When true: Windows x64 ABI (rcx=globals, rdx=arg0, r8=arg1, r9=arg2),
	//   invoked via syscall.Syscall6 so the goroutine enters _Gsyscall state.
	//   This enables syscall.NewCallback to work for OpRegCall callbacks.
	// When false: Bridge ABI (rdi=globals, rax=arg0), invoked via NOSPLIT assembly.
	syscallABI bool

	// Register allocation
	// We use: rax(0), rbx(1), rcx(2), rdx(3), r8(4), r9(5), r10(6), r11(7)
	// r12-r15 are callee-saved, used for locals 8-11
	// Stack is used for locals 12+
	// rdi holds globals pointer (first argument)
}

// spilledRegsCount returns the number of VM registers that must be spilled to the stack.
// VM registers R12 and above (indices 12..numRegs-1) are stored on the stack.
func (cg *NativeCodeGenerator) spilledRegsCount() int {
	if cg.numRegs > 12 {
		return cg.numRegs - 12
	}
	return 0
}

// spillAreaSize returns the byte size of the spilled register area on the stack.
func (cg *NativeCodeGenerator) spillAreaSize() int {
	return cg.spilledRegsCount() * 8
}

// localsBaseOffset returns the rbp-relative byte offset where local variables begin.
// Stack layout: [rbp-8]=R0, [rbp-16..rbp-40]=reserved, [rbp-48..]=spilled regs, then locals.
// So the first local is at [rbp - (48 + spillAreaSize + 8)].
func (cg *NativeCodeGenerator) localsBaseOffset() int {
	return 48 + cg.spillAreaSize()
}

// NewNativeCodeGenerator creates a new native code generator (bridge ABI)
func NewNativeCodeGenerator() *NativeCodeGenerator {
	return &NativeCodeGenerator{
		code:      make([]byte, 0, 4096),
		labels:    make(map[string]int),
		fixups:    make([]fixup, 0),
		constants: nil,
	}
}

// NewNativeCodeGeneratorSyscall creates a native code generator with Windows x64 syscall ABI.
// This enables OpRegCall callbacks via syscall.NewCallback because the goroutine
// enters _Gsyscall state when invoked via syscall.Syscall6.
func NewNativeCodeGeneratorSyscall() *NativeCodeGenerator {
	cg := NewNativeCodeGenerator()
	cg.syscallABI = true
	cg.builtinCallbackPtr = GetBuiltinCallbackPtr()
	cg.functionCallbackPtr = GetFunctionCallbackPtr()
	cg.collectionCallbackPtr = GetCollectionCallbackPtr()
	cg.objectCallbackPtr = GetObjectCallbackPtr()
	return cg
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
	cg.numRegs = fn.NumRegs
	cg.regConstMap = make(map[int]int)
	cg.globalConstMap = make(map[int]int)

	// Generate prologue
	cg.emitPrologue(fn.NumLocals, fn.NumRegs)

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

// emitPrologue generates function entry code.
//
// When syscallABI is false (bridge ABI, legacy):
//   - Entry via NOSPLIT bridge assembly: rdi=globals, rax=arg0(R0)
//   - Prologue stores rax to [rbp-8] for R0
//
// When syscallABI is true (Windows x64 ABI for syscall.Syscall6):
//   - Entry via syscall.Syscall6: rcx=globals, rdx=arg0, r8=arg1, r9=arg2
//   - Prologue adapts: rcx→rdi (globals), rdx→[rbp-8] (R0), r8→rbx (R1), r9→rcx (R2)
//   - This enables syscall.NewCallback to work because the goroutine is in _Gsyscall state
//
// Stack layout:
//
//	[rbp]                        = old rbp
//	[rbp-8]                      = R0 (VM reg 0, stored on stack)
//	[rbp-16..rbp-40]             = reserved / alignment
//	[rbp-48]                     = R12 (first spilled VM register)
//	[rbp-48+(numSpilled-1)*8]   = last spilled VM register
//	[rbp-48+numSpilled*8]       = local[0]
//	[rbp-48+numSpilled*8+numLocals*8] = end of locals
//	[rbp-48+numSpilled*8+numLocals*8+8] = builtin args spill area (up to 8*8=64 bytes)
//
// numSpilled = max(0, numRegs-12) for VM registers R12+
// Total stack = 48 (header+R0+reserved) + numSpilled*8 + numLocals*8 + 8 (buffer) + 64 (builtin args)
func (cg *NativeCodeGenerator) emitPrologue(numLocals int, numRegs int) {
	// push rbp
	cg.emitByte(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Calculate spilled register count: VM regs R12 and above go on stack
	numSpilled := 0
	if numRegs > 12 {
		numSpilled = numRegs - 12
	}

	// Allocate stack: 48 bytes (rbp+R0+reserved) + spilled regs + locals + builtin args spill (72 bytes)
	stackSize := 48 + numSpilled*8 + numLocals*8 + 72
	// Round up to 16 bytes for alignment
	if stackSize%16 != 0 {
		stackSize += 16 - (stackSize % 16)
	}
	// sub rsp, stackSize
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(uint32(stackSize))

	if cg.syscallABI {
		// Windows x64 ABI: rcx=globals, rdx=arg0, r8=arg1, r9=arg2
		// Adapt to VM convention: rdi=globals, [rbp-8]=R0(arg0), rbx=R1(arg1), rcx=R2(arg2)
		cg.emitBytes([]byte{0x48, 0x89, 0xCF}) // mov rdi, rcx (globals: rcx → rdi)
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp + disp32], rdx (arg0 → R0)
		cg.emitUint32(r0StackDisp)
		cg.emitBytes([]byte{0x49, 0x89, 0xC3}) // mov rbx, r8 (arg1 → R1)
		cg.emitBytes([]byte{0x49, 0x89, 0xC1}) // mov rcx, r9 (arg2 → R2, rcx freed by globals move)
	} else {
		// Bridge ABI (legacy): rdi=globals, rax=arg0(R0)
		// Save R0's initial value from rax to its stack slot at [rbp-8].
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(r0StackDisp)
	}
}

// emitEpilogue generates function exit code
func (cg *NativeCodeGenerator) emitEpilogue() {
	numSpilled := 0
	if cg.numRegs > 12 {
		numSpilled = cg.numRegs - 12
	}
	stackSize := 48 + numSpilled*8 + cg.numLocals*8 + 72
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

	case compiler.OpRegAddConst:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		constIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileAddConst(dst, src, constIdx)
		*ip += 5

	case compiler.OpRegSubConst:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		constIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileSubConst(dst, src, constIdx)
		*ip += 5

	case compiler.OpRegMulConst:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		constIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		cg.compileMulConst(dst, src, constIdx)
		*ip += 5

	case compiler.OpRegAnd:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileLogicalAnd(dst, left, right)
		*ip += 4

	case compiler.OpRegOr:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileLogicalOr(dst, left, right)
		*ip += 4

	case compiler.OpRegNot:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		cg.compileLogicalNot(dst, src)
		*ip += 3

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

	case compiler.OpRegPop:
		// Register VM pop only affects stack bookkeeping on the interpreter path.
		// The native register model does not need to materialize it.
		*ip += 1

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

	case compiler.OpRegAddLocalCheck:
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		limitIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
		target := *ip + jumpOffset
		cg.compileAddLocalCheck(accReg, counterReg, limitIdx, target)
		*ip += 7

	case compiler.OpRegLoopIncCheck:
		// Format: counter_reg, limit_const(16), jump_offset(16)
		// Increment counter; if counter < limit, jump to target
		counterReg := int(code[*ip+1])
		limitIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		jumpOffset := int(int16(uint16(code[*ip+4])<<8 | uint16(code[*ip+5])))
		target := *ip + jumpOffset
		cg.compileLoopIncCheck(counterReg, limitIdx, target)
		*ip += 6

	case compiler.OpRegLoopBodyAdd:
		// Format: acc_reg, counter_reg, limit_const(16), jump_offset(16)
		// acc += counter; counter++; if counter < limit, jump to target
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		limitIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
		target := *ip + jumpOffset
		cg.compileLoopBodyAdd(accReg, counterReg, limitIdx, target)
		*ip += 7

	case compiler.OpRegLoopMulCheck:
		// Format: i_reg, n_reg, jump_out_offset(16)
		// if i*i > n, jump out of loop
		iReg := int(code[*ip+1])
		nReg := int(code[*ip+2])
		jumpOutOffset := int(int16(uint16(code[*ip+3])<<8 | uint16(code[*ip+4])))
		target := *ip + jumpOutOffset
		cg.compileLoopMulCheck(iReg, nReg, target)
		*ip += 5

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

// r0StackDisp is the stack displacement for VM register 0 (R0).
// R0 lives at [rbp-8] so that rax can be used purely as a scratch register.
const r0StackDisp uint32 = 0xFFFFFFF8 // -8 as signed int32, stored as uint32 for x86 disp32

// loadRegToRax loads a register to rax
// R0 (VM reg 0) lives on the stack at [rbp-8] so that rax can be used
// purely as a scratch/temp register without clobbering R0 between operations.
func (cg *NativeCodeGenerator) loadRegToRax(r int) {
	if r == 255 {
		// R255 is the ReturnRegister — value is already in rax (set by OpRegMove dst=255
		// or by OpRegCall storing result to R255 via storeRaxToReg(255) which leaves it in rax)
		// No instruction needed.
		return
	}
	if r == 0 {
		// R0 is stored on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if r < 8 {
		// Standard registers: rbx(1), rcx(2), rdx(3), r8(4)-r11(7)
		switch r {
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
		// [rbp-8] = R0, [rbp-16..rbp-40] = reserved
		// [rbp-48] = R12 (first spilled register)
		// [rbp-48+(numSpilled-1)*8] = last spilled register
		// Note: displacement is signed, so we emit -offset as uint32
		offset := 48 + (r-12)*8
		cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp + disp32]
		cg.emitUint32(uint32(-offset))
	}
}

// storeRaxToReg stores rax to a register
// R0 (VM reg 0) lives on the stack at [rbp-8].
func (cg *NativeCodeGenerator) storeRaxToReg(r int) {
	if r == 255 {
		// R255 is the ReturnRegister — leave value in rax.
		// OpRegReturn 255 will load R255 to rax (which is a no-op since it's already there),
		// then call emitEpilogue which returns rax.
		// Do NOT store to stack — R255 is not allocated in the stack frame.
		return
	}
	if r == 0 {
		// R0 is stored on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(r0StackDisp)
	} else if r < 8 {
		switch r {
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
	cg.regConstMap[dst] = constIdx
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
	// Track constant propagation from globals
	if constIdx, ok := cg.globalConstMap[globalIdx]; ok {
		cg.regConstMap[dst] = constIdx
	}

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
	// Track constant propagation through globals
	if constIdx, ok := cg.regConstMap[src]; ok {
		cg.globalConstMap[globalIdx] = constIdx
	}

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
// Local variables are stored on stack: [rbp - (localsBaseOffset + (localIdx+1)*8)]
// localsBaseOffset = 48 + numSpilled*8, which accounts for R0 slot, reserved space,
// and any spilled VM registers (R12+).
func (cg *NativeCodeGenerator) compileLoadLocal(dst, localIdx int) {
	// Local offset: after header (48) + spilled register area
	offset := cg.localsBaseOffset() + (localIdx+1)*8

	// mov rax, [rbp - offset]
	// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp + disp32]
	cg.emitUint32(uint32(-offset))

	// Store to destination register
	cg.storeRaxToReg(dst)
}

// compileStoreLocal stores a register value to a local variable
// Local variables are stored on stack: [rbp - (localsBaseOffset + (localIdx+1)*8)]
func (cg *NativeCodeGenerator) compileStoreLocal(src, localIdx int) {
	// Load source to rax
	cg.loadRegToRax(src)

	// Local offset: after header (48) + spilled register area
	offset := cg.localsBaseOffset() + (localIdx+1)*8

	// mov [rbp - offset], rax
	// Note: displacement is signed, so we emit -offset as uint32 (two's complement)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
	cg.emitUint32(uint32(-offset))
}

// compileTailCall handles tail calls via the Windows x64 ABI callback.
// On Windows x64 ABI: rcx = constIdx, rdx = numArgs, r8 = argsPtr.
// After the callback returns, the result is in rax and we emit an epilogue+return.
func (cg *NativeCodeGenerator) compileTailCall(funcReg, numArgs int) {
	callbackPtr := getWindowsCallbackPtr()
	constIdx, known := cg.regConstMap[funcReg]

	if callbackPtr == 0 || !known {
		// Fallback: for self-recursive calls, jump to function entry
		if cg.funcEntry > 0 {
			// jmp to function entry
			offset := int32(cg.funcEntry - (len(cg.code) + 5))
			cg.emitBytes([]byte{0xE9}) // jmp rel32
			cg.emitUint32(uint32(offset))
		} else {
			cg.emitEpilogue()
		}
		return
	}

	// Save callee-saved registers
	cg.emitBytes([]byte{0x57})       // push rdi (globals pointer)
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// Spill args to stack below locals
	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	// Spill R0 ([rbp-8]) to args array
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp-8] (R0)
	cg.emitUint32(r0StackDisp)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp-baseOffset], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp-baseOffset-8], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp-baseOffset-16], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp-baseOffset-24], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp-baseOffset-32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp-baseOffset-40], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp-baseOffset-48], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp-baseOffset-56], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Allocate shadow space + stack alignment
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	// Set up Windows x64 ABI callback arguments:
	//   rcx = constIdx (function index in constant pool)
	//   rdx = numArgs
	//   r8  = argsPtr (lea r8, [rbp-baseOffset])
	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, imm32
	cg.emitUint32(uint32(constIdx))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, imm32
	cg.emitUint32(uint32(numArgs))

	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback function pointer
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	// Result is in rax, store to R255 then emit epilogue+return
	cg.storeRaxToReg(255)

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi

	// Emit epilogue and return with the callback result
	cg.emitEpilogue()
}

// compileCall handles function calls by calling back to Go via the Windows
// x64 ABI callback. The callback resolves the function from the constant
// pool using constIdx and executes it (natively if compiled, or via interpreter).
//
// On Windows x64 ABI: rcx = constIdx, rdx = numArgs, r8 = argsPtr.
// The callback pointer is obtained from getWindowsCallbackPtr() (syscall.NewCallback).
func (cg *NativeCodeGenerator) compileCall(funcReg, numArgs int) {
	callbackPtr := getWindowsCallbackPtr()

	constIdx, known := cg.regConstMap[funcReg]
	if !known || callbackPtr == 0 {
		// Cannot resolve function — emit return-0 stub
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		cg.storeRaxToReg(255)
		return
	}

	// Save callee-saved registers that will be clobbered by the callback
	// Windows x64 ABI requires: push rbp, then sub rsp for shadow space + alignment
	cg.emitBytes([]byte{0x57})       // push rdi (globals pointer)
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// Spill argument registers (R0-R7) to the stack area below locals.
	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	// Spill R0 ([rbp-8]) to args array
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp-8] (R0)
	cg.emitUint32(r0StackDisp)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp-baseOffset], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp-baseOffset-8], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp-baseOffset-16], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp-baseOffset-24], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp-baseOffset-32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp-baseOffset-40], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp-baseOffset-48], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp-baseOffset-56], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Allocate shadow space (32 bytes) + ensure 16-byte stack alignment
	// After 4 pushes (32 bytes), RSP is misaligned. sub rsp, 40 fixes alignment + shadow.
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	// Set up Windows x64 ABI callback arguments:
	//   rcx = constIdx (function index in constant pool)
	//   rdx = numArgs
	//   r8  = argsPtr (lea r8, [rbp-baseOffset])
	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, imm32
	cg.emitUint32(uint32(constIdx))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, imm32
	cg.emitUint32(uint32(numArgs))

	// lea r8, [rbp-baseOffset]
	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback function pointer
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	// Result is in rax, store to R255
	cg.storeRaxToReg(255)

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi
}

// compileBuiltin generates code to call a builtin function via callback
// The callback signature: callback(builtinIdx, numArgs, argsPtr) int64
// builtinIdx is the auto-assigned index from objects.BuiltinIndexMap.
// Arguments are in VM registers R0-R7 (which map to RAX, RBX, RCX, RDX, R8-R11).
//
// On Windows x64 ABI: rcx = builtinIdx, rdx = numArgs, r8 = argsPtr.
// The callback pointer is obtained from GetBuiltinCallbackPtr() (syscall.NewCallback).
func (cg *NativeCodeGenerator) compileBuiltin(builtinIdx, numArgs int) {
	callbackPtr := GetBuiltinCallbackPtr()
	if callbackPtr == 0 {
		// No callback available — emit return-0 stub
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		cg.storeRaxToReg(255)
		return
	}

	// Save callee-saved registers that will be clobbered by the callback
	cg.emitBytes([]byte{0x57})       // push rdi (globals pointer)
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// Spill argument registers (R0-R7) to the stack area below locals.
	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	// Spill R0 ([rbp-8]) to args array
	cg.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp-8] (R0)
	cg.emitUint32(r0StackDisp)
	cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp-baseOffset], rax
	cg.emitUint32(uint32(-baseOffset))

	if numArgs >= 2 {
		cg.emitBytes([]byte{0x48, 0x89, 0x9D}) // mov [rbp-baseOffset-8], rbx
		cg.emitUint32(uint32(-(baseOffset + 8)))
	}
	if numArgs >= 3 {
		cg.emitBytes([]byte{0x48, 0x89, 0x8D}) // mov [rbp-baseOffset-16], rcx
		cg.emitUint32(uint32(-(baseOffset + 16)))
	}
	if numArgs >= 4 {
		cg.emitBytes([]byte{0x48, 0x89, 0x95}) // mov [rbp-baseOffset-24], rdx
		cg.emitUint32(uint32(-(baseOffset + 24)))
	}
	if numArgs >= 5 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x85}) // mov [rbp-baseOffset-32], r8
		cg.emitUint32(uint32(-(baseOffset + 32)))
	}
	if numArgs >= 6 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x8D}) // mov [rbp-baseOffset-40], r9
		cg.emitUint32(uint32(-(baseOffset + 40)))
	}
	if numArgs >= 7 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x95}) // mov [rbp-baseOffset-48], r10
		cg.emitUint32(uint32(-(baseOffset + 48)))
	}
	if numArgs >= 8 {
		cg.emitBytes([]byte{0x4C, 0x89, 0x9D}) // mov [rbp-baseOffset-56], r11
		cg.emitUint32(uint32(-(baseOffset + 56)))
	}

	// Allocate shadow space (32 bytes) + ensure 16-byte stack alignment
	// After 4 pushes (32 bytes), RSP is misaligned. sub rsp, 40 fixes alignment + shadow.
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	// Set up Windows x64 ABI callback arguments:
	//   rcx = builtinIdx (auto-assigned index)
	//   rdx = numArgs
	//   r8  = argsPtr (lea r8, [rbp-baseOffset])
	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, imm32
	cg.emitUint32(uint32(builtinIdx))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, imm32
	cg.emitUint32(uint32(numArgs))

	// lea r8, [rbp-baseOffset]
	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback function pointer
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	// Result is in rax, store to R255
	cg.storeRaxToReg(255)

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi
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

// compileCollectionOp generates code to perform a collection operation via callback.
// On Windows x64 ABI: rcx = opKind, rdx = numArgs, r8 = argsPtr.
func (cg *NativeCodeGenerator) compileCollectionOp(opKind CollectionOpKind, dstReg int, numArgs int, argRegs []int) {
	callbackPtr := GetCollectionCallbackPtr()
	if callbackPtr == 0 {
		// No callback — emit return-0 stub
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		if dstReg > 0 && dstReg != 255 {
			cg.storeRaxToReg(dstReg)
		}
		return
	}

	// Save callee-saved registers
	cg.emitBytes([]byte{0x57})       // push rdi (globals pointer)
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// Spill args to stack below locals
	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	for i, reg := range argRegs {
		if i >= 8 {
			break
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85}) // mov [rbp + disp32], rax
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	// Allocate shadow space + stack alignment
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	// Set up Windows x64 ABI callback arguments:
	//   rcx = opKind
	//   rdx = numArgs
	//   r8  = argsPtr
	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, imm32
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, imm32
	cg.emitUint32(uint32(numArgs))

	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	// Store result to destination register
	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi
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

// compileCollectionOpDirect generates code for collection operations with direct arg passing.
// On Windows x64 ABI: rcx = opKind, rdx = numArgs, r8 = argsPtr.
func (cg *NativeCodeGenerator) compileCollectionOpDirect(opKind CollectionOpKind, dstReg int, numArgs int, argRegs []int) {
	callbackPtr := GetCollectionCallbackPtr()
	if callbackPtr == 0 {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		if dstReg > 0 && dstReg != 255 {
			cg.storeRaxToReg(dstReg)
		}
		return
	}

	// Save callee-saved registers
	cg.emitBytes([]byte{0x57})       // push rdi
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	for i, reg := range argRegs {
		if i >= 8 {
			break
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	// Allocate shadow space + stack alignment
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, opKind
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, numArgs
	cg.emitUint32(uint32(numArgs))

	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	// Call the callback
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi
}

// compileObjectOp generates code to perform object field operations via callback.
// On Windows x64 ABI: rcx = opKind, rdx = numArgs, r8 = argsPtr, r9 = nameIdx.
func (cg *NativeCodeGenerator) compileObjectOp(opKind ObjectOpKind, dstReg int, argRegs []int, nameIdx int) {
	callbackPtr := GetObjectCallbackPtr()
	if callbackPtr == 0 {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
		if dstReg > 0 && dstReg != 255 {
			cg.storeRaxToReg(dstReg)
		}
		return
	}

	// Save callee-saved registers
	cg.emitBytes([]byte{0x57})       // push rdi
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	baseOffset := int32(cg.localsBaseOffset() + cg.numLocals*8 + 8)

	for i, reg := range argRegs {
		if i >= 8 {
			break
		}
		cg.loadRegToRax(reg)
		cg.emitBytes([]byte{0x48, 0x89, 0x85})
		cg.emitUint32(uint32(-(baseOffset + int32(i*8))))
	}

	// Allocate shadow space + stack alignment
	cg.emitBytes([]byte{0x48, 0x83, 0xEC, 0x28}) // sub rsp, 40

	// Set up Windows x64 ABI callback arguments:
	//   rcx = opKind
	//   rdx = numArgs
	//   r8  = argsPtr
	//   r9  = nameIdx
	cg.emitBytes([]byte{0x48, 0xC7, 0xC1}) // mov rcx, opKind
	cg.emitUint32(uint32(opKind))

	cg.emitBytes([]byte{0x48, 0xC7, 0xC2}) // mov rdx, numArgs
	cg.emitUint32(uint32(len(argRegs)))

	cg.emitBytes([]byte{0x4C, 0x8D, 0x85}) // lea r8, [rbp + disp32]
	cg.emitUint32(uint32(-baseOffset))

	cg.emitBytes([]byte{0x49, 0xC7, 0xC1}) // mov r9, nameIdx
	cg.emitUint32(uint32(nameIdx))

	// Call the callback
	cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
	cg.emitUint64(uint64(callbackPtr))
	cg.emitBytes([]byte{0xFF, 0xD0}) // call rax

	// Restore shadow space
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x28}) // add rsp, 40

	if dstReg > 0 && dstReg != 255 {
		cg.storeRaxToReg(dstReg)
	}

	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emitBytes([]byte{0x5F})       // pop rdi
}

func (cg *NativeCodeGenerator) compileMove(dst, src int) {
	if dst == src {
		return
	}

	// Propagate constant tracking from src to dst
	if constIdx, ok := cg.regConstMap[src]; ok {
		cg.regConstMap[dst] = constIdx
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
	if right == 0 {
		// R0 is on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x03, 0x85}) // add rax, [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if right < 8 {
		switch right {
		case 1:
			cg.emitBytes([]byte{0x48, 0x01, 0xD8}) // add rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x01, 0xC8}) // add rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x01, 0xD0}) // add rax, rdx
		case 4:
			cg.emitBytes([]byte{0x4C, 0x01, 0xC0}) // add rax, r8
		case 5:
			cg.emitBytes([]byte{0x4C, 0x01, 0xC8}) // add rax, r9
		case 6:
			cg.emitBytes([]byte{0x4C, 0x01, 0xD0}) // add rax, r10
		case 7:
			cg.emitBytes([]byte{0x4C, 0x01, 0xD8}) // add rax, r11
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

	if right == 0 {
		// R0 is on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x2B, 0x85}) // sub rax, [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if right < 8 {
		switch right {
		case 1:
			cg.emitBytes([]byte{0x48, 0x29, 0xD8}) // sub rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x29, 0xC8}) // sub rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x29, 0xD0}) // sub rax, rdx
		case 4:
			cg.emitBytes([]byte{0x4C, 0x29, 0xC0}) // sub rax, r8
		case 5:
			cg.emitBytes([]byte{0x4C, 0x29, 0xC8}) // sub rax, r9
		case 6:
			cg.emitBytes([]byte{0x4C, 0x29, 0xD0}) // sub rax, r10
		case 7:
			cg.emitBytes([]byte{0x4C, 0x29, 0xD8}) // sub rax, r11
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

	if right == 0 {
		// R0 is on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x85}) // imul rax, [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if right < 8 {
		switch right {
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
		// For mov rcx, r12: RCX is r/m=1, R12 is reg=4 with R
		// REX = 0x4C (W=1, R=1, B=0)
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE1}) // mov rcx, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE9}) // mov rcx, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF1}) // mov rcx, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF9}) // mov rcx, r15
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

	// idiv rcx (ModRM=F9: mod=11, reg=7 for idiv, r/m=1 for RCX)
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})

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
		// For mov rcx, r12: RCX is r/m=1, R12 is reg=4 with R
		// REX = 0x4C (W=1, R=1, B=0)
		switch right {
		case 8:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE1}) // mov rcx, r12
		case 9:
			cg.emitBytes([]byte{0x4C, 0x89, 0xE9}) // mov rcx, r13
		case 10:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF1}) // mov rcx, r14
		case 11:
			cg.emitBytes([]byte{0x4C, 0x89, 0xF9}) // mov rcx, r15
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

	// idiv rcx (ModRM=F9: mod=11, reg=7 for idiv, r/m=1 for RCX)
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})

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

func (cg *NativeCodeGenerator) compileAddConst(dst, src, constIdx int) {
	cg.loadRegToRax(src)
	if constIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x05}) // add rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[constIdx])))
	}
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileSubConst(dst, src, constIdx int) {
	cg.loadRegToRax(src)
	if constIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x2D}) // sub rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[constIdx])))
	}
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileMulConst(dst, src, constIdx int) {
	cg.loadRegToRax(src)
	if constIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x69, 0xC0}) // imul rax, rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[constIdx])))
	}
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileAddLocalCheck(accReg, counterReg, limitIdx, target int) {
	cg.compileAdd(accReg, accReg, counterReg)
	cg.compileInc(counterReg)

	cg.loadRegToRax(counterReg)
	if limitIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[limitIdx])))
	} else {
		cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	}

	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0x0F, 0x8C}) // jl rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// compileLoopIncCheck implements OpRegLoopIncCheck:
// Increment counter; if counter < limit, jump to target
func (cg *NativeCodeGenerator) compileLoopIncCheck(counterReg, limitIdx, target int) {
	cg.compileInc(counterReg)

	cg.loadRegToRax(counterReg)
	if limitIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[limitIdx])))
	} else {
		cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	}

	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0x0F, 0x8C}) // jl rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// compileLoopBodyAdd implements OpRegLoopBodyAdd:
// acc += counter; counter++; if counter < limit, jump to target
func (cg *NativeCodeGenerator) compileLoopBodyAdd(accReg, counterReg, limitIdx, target int) {
	cg.compileAdd(accReg, accReg, counterReg)
	cg.compileInc(counterReg)

	cg.loadRegToRax(counterReg)
	if limitIdx < len(cg.constants) {
		cg.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm32
		cg.emitUint32(uint32(int32(cg.constants[limitIdx])))
	} else {
		cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	}

	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0x0F, 0x8C}) // jl rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

// compileLoopMulCheck implements OpRegLoopMulCheck:
// if i*i > n, jump to target (exit loop — used for prime checking)
func (cg *NativeCodeGenerator) compileLoopMulCheck(iReg, nReg, target int) {
	// Load i into rax, compute i*i
	cg.loadRegToRax(iReg)
	// imul rax, rax → rax = i * i
	cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC0}) // imul rax, rax

	// Compare i*i (in rax) with n (in register nReg)
	cg.compareRaxWithReg(nReg)

	// If i*i > n (i.e., rax > nReg), jump to target
	label := fmt.Sprintf("L%d", target)
	cg.emitBytes([]byte{0x0F, 0x8F}) // jg rel32
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *NativeCodeGenerator) compileLogicalAnd(dst, left, right int) {
	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	falseLabel := fmt.Sprintf("L_logic_false_%d", len(cg.fixups))
	endLabel := fmt.Sprintf("L_logic_end_%d", len(cg.fixups))
	cg.emitBytes([]byte{0x0F, 0x84})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: falseLabel, size: 4})
	cg.emitUint32(0)

	cg.loadRegToRax(right)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x84})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: falseLabel, size: 4})
	cg.emitUint32(0)

	cg.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00})
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: endLabel, size: 4})
	cg.emitUint32(0)

	cg.labels[falseLabel] = len(cg.code)
	cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	cg.labels[endLabel] = len(cg.code)
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileLogicalOr(dst, left, right int) {
	cg.loadRegToRax(left)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	trueLabel := fmt.Sprintf("L_logic_true_%d", len(cg.fixups))
	endLabel := fmt.Sprintf("L_logic_end_%d", len(cg.fixups))
	cg.emitBytes([]byte{0x0F, 0x85})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: trueLabel, size: 4})
	cg.emitUint32(0)

	cg.loadRegToRax(right)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0})
	cg.emitBytes([]byte{0x0F, 0x85})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: trueLabel, size: 4})
	cg.emitUint32(0)

	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: endLabel, size: 4})
	cg.emitUint32(0)

	cg.labels[trueLabel] = len(cg.code)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00})
	cg.labels[endLabel] = len(cg.code)
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileLogicalNot(dst, src int) {
	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al
	cg.emitBytes([]byte{0x48, 0x0F, 0xB6, 0xC0})
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileNeg(dst, src int) {
	cg.loadRegToRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.storeRaxToReg(dst)
}

func (cg *NativeCodeGenerator) compileCompare(dst, left, right int, op string) {
	// Load left to rax, then compare with right.
	// R0 is on the stack, so loadRegToRax/compareRaxWithReg handle it uniformly.
	cg.loadRegToRax(left)
	cg.compareRaxWithReg(right)

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
	if r == 0 {
		// R0 lives on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0x3B, 0x85}) // cmp rax, [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if r < 8 {
		switch r {
		case 1:
			cg.emitBytes([]byte{0x48, 0x39, 0xD8}) // cmp rax, rbx
		case 2:
			cg.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
		case 3:
			cg.emitBytes([]byte{0x48, 0x39, 0xD0}) // cmp rax, rdx
		case 4:
			cg.emitBytes([]byte{0x4C, 0x39, 0xC0}) // cmp rax, r8
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
	if reg == 0 {
		// R0 lives on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0xFF, 0x85}) // inc qword [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if reg < 8 {
		switch reg {
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
	if reg == 0 {
		// R0 lives on the stack at [rbp-8]
		cg.emitBytes([]byte{0x48, 0xFF, 0x8D}) // dec qword [rbp + disp32]
		cg.emitUint32(r0StackDisp)
	} else if reg < 8 {
		switch reg {
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
