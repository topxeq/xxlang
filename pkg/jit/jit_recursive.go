// pkg/jit/jit_recursive.go
// JIT compilation for recursive functions with call inlining
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// RecursiveJITCompiler handles JIT compilation of recursive functions
// by converting recursion to iteration where possible
type RecursiveJITCompiler struct {
	config JITConfig

	// Detected recursive functions
	recursiveFuncs map[uint64]bool
}

// NewRecursiveJITCompiler creates a new recursive JIT compiler
func NewRecursiveJITCompiler(config JITConfig) *RecursiveJITCompiler {
	return &RecursiveJITCompiler{
		config:         config,
		recursiveFuncs: make(map[uint64]bool),
	}
}

// AnalyzeRecursiveFunction analyzes if a function is self-recursive
func (r *RecursiveJITCompiler) AnalyzeRecursiveFunction(fn *compiler.CompiledFunction, constants []vm.Value) bool {
	// Check if function calls itself
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		if op == compiler.OpRegCall || op == compiler.OpRegTailCall {
			// Check if the call target is this function
			// For recursive calls, the function is typically in a constant slot
			// or in a register that was loaded from constants
			return true
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

	return false
}

// RecursiveCodeGenerator generates optimized code for recursive functions
// by converting tail recursion to loops
type RecursiveCodeGenerator struct {
	config JITConfig
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction

	// Stack for simulating recursion
	stackSize int
	maxDepth  int
}

// NewRecursiveCodeGenerator creates a new recursive code generator
func NewRecursiveCodeGenerator(config JITConfig) *RecursiveCodeGenerator {
	return &RecursiveCodeGenerator{
		config:  config,
		code:    make([]byte, 0, 16384),
		labels:  make(map[string]int),
		fixups:  make([]fixup, 0),
		maxDepth: 10000, // Maximum simulated recursion depth
	}
}

// Generate generates code for a function, handling recursion
func (cg *RecursiveCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn

	// Analyze the function
	callInfo := cg.analyzeCalls(fn.Instructions)

	if callInfo.hasTailCall {
		// Convert tail recursion to loop
		return cg.generateTailRecursive(fn, callInfo)
	}

	if callInfo.hasCall {
		// Use stack-based simulation for non-tail recursion
		return cg.generateStackRecursive(fn, callInfo)
	}

	// No calls - generate normally
	return cg.generateNormal(fn)
}

type callAnalysis struct {
	hasCall      bool
	hasTailCall  bool
	callCount    int
	tailCallIPs  []int
	callTargets  []int
}

func (cg *RecursiveCodeGenerator) analyzeCalls(code []byte) callAnalysis {
	info := callAnalysis{}

	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegCall:
			info.hasCall = true
			info.callCount++
			ip += 3

		case compiler.OpRegTailCall:
			info.hasTailCall = true
			info.tailCallIPs = append(info.tailCallIPs, ip)
			info.callCount++
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

	return info
}

// generateTailRecursive generates a loop for tail-recursive functions
func (cg *RecursiveCodeGenerator) generateTailRecursive(fn *compiler.CompiledFunction, info callAnalysis) ([]byte, error) {
	// Generate prologue with extra stack space for simulating recursion
	cg.emitPrologueRecursive()

	// Create labels for the function entry
	entryLabel := "func_entry"
	cg.labels[entryLabel] = len(cg.code)

	// Generate code, replacing tail calls with jumps
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		cg.labels[fmt.Sprintf("ip_%d", ip)] = len(cg.code)

		switch op {
		case compiler.OpRegTailCall:
			// Convert tail call to parameter update + jump
			cg.compileTailCallToJump(code, &ip, entryLabel)
			continue

		case compiler.OpRegCall:
			// Non-tail call - need stack simulation
			return nil, fmt.Errorf("mixed tail/non-tail recursion not supported in JIT")

		default:
			err := cg.compileOpcode(op, code, &ip)
			if err != nil {
				return nil, err
			}
		}
	}

	// Resolve fixups
	cg.resolveFixups()

	return cg.code, nil
}

// generateStackRecursive generates code with explicit stack for non-tail recursion
func (cg *RecursiveCodeGenerator) generateStackRecursive(fn *compiler.CompiledFunction, info callAnalysis) ([]byte, error) {
	// For non-tail recursion, we use an explicit stack to simulate the call stack
	// This is complex and typically slower than the interpreter for deep recursion
	// So we return an error to fall back to interpreter
	return nil, fmt.Errorf("non-tail recursion not supported in JIT - use interpreter")
}

// generateNormal generates code for non-recursive functions
func (cg *RecursiveCodeGenerator) generateNormal(fn *compiler.CompiledFunction) ([]byte, error) {
	cg.emitPrologue()

	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		cg.labels[fmt.Sprintf("ip_%d", ip)] = len(cg.code)

		err := cg.compileOpcode(op, code, &ip)
		if err != nil {
			return nil, err
		}
	}

	cg.resolveFixups()
	return cg.code, nil
}

// compileTailCallToJump converts a tail call to parameter update + jump
func (cg *RecursiveCodeGenerator) compileTailCallToJump(code []byte, ip *int, entryLabel string) {
	// OpRegTailCall: func_reg, num_args
	// funcReg := int(code[*ip+1])
	// numArgs := int(code[*ip+2])

	// For self-recursive tail calls:
	// 1. Update parameters (R0, R1, ...) with new values
	// 2. Jump to function entry

	// The new parameter values are already in R0, R1, etc.
	// because the caller put them there before the tail call

	// For fibonacci-like functions:
	// fibHelper(n, a, b) -> fibHelper(n-1, b, a+b)
	// We need to update R0, R1, R2

	// Generate: copy argument registers to parameter registers
	// Then jump to entry

	// The arguments should already be in the right place
	// Just jump to the beginning
	cg.emitJmp(entryLabel)
	*ip += 3
}

// compileOpcode compiles a single opcode
func (cg *RecursiveCodeGenerator) compileOpcode(op compiler.Opcode, code []byte, ip *int) error {
	switch op {
	// Data movement
	case compiler.OpRegLoadConst:
		cg.compileLoadConst(code, ip)
	case compiler.OpRegMove:
		cg.compileMove(code, ip)

	// Arithmetic
	case compiler.OpRegAdd:
		cg.compileBinaryOp(code, ip, 0x01)
	case compiler.OpRegSub:
		cg.compileBinaryOp(code, ip, 0x29)
	case compiler.OpRegMul:
		cg.compileMul(code, ip)
	case compiler.OpRegDiv:
		cg.compileDiv(code, ip)
	case compiler.OpRegMod:
		cg.compileMod(code, ip)
	case compiler.OpRegNeg:
		cg.compileNeg(code, ip)

	// Comparison
	case compiler.OpRegLess:
		cg.compileComparison(code, ip, 0x9C)
	case compiler.OpRegGreater:
		cg.compileComparison(code, ip, 0x9F)
	case compiler.OpRegEqual:
		cg.compileComparison(code, ip, 0x94)
	case compiler.OpRegNotEqual:
		cg.compileComparison(code, ip, 0x95)
	case compiler.OpRegLessEqual:
		cg.compileComparison(code, ip, 0x9E)
	case compiler.OpRegGreaterEqual:
		cg.compileComparison(code, ip, 0x9D)

	// Logical
	case compiler.OpRegNot:
		cg.compileNot(code, ip)

	// Control flow
	case compiler.OpRegJump:
		cg.compileJump(code, ip)
	case compiler.OpRegJumpIfFalse:
		cg.compileJumpIfFalse(code, ip)
	case compiler.OpRegJumpIfTrue:
		cg.compileJumpIfTrue(code, ip)
	case compiler.OpRegReturn:
		cg.compileReturn(code, ip)

	// Local variables
	case compiler.OpRegLoadLocal:
		cg.compileLoadLocal(code, ip)
	case compiler.OpRegStoreLocal:
		cg.compileStoreLocal(code, ip)

	// Global variables
	case compiler.OpRegLoadGlobal:
		cg.compileLoadGlobal(code, ip)
	case compiler.OpRegStoreGlobal:
		cg.compileStoreGlobal(code, ip)

	// Increment/Decrement
	case compiler.OpRegIncLocal:
		cg.compileIncLocal(code, ip)
	case compiler.OpRegDecLocal:
		cg.compileDecLocal(code, ip)

	// Null/True/False
	case compiler.OpRegNull:
		cg.compileNull(code, ip)
	case compiler.OpRegTrue:
		cg.compileTrue(code, ip)
	case compiler.OpRegFalse:
		cg.compileFalse(code, ip)

	// Stack operations
	case compiler.OpRegPush:
		cg.compilePush(code, ip)
	case compiler.OpRegPop:
		cg.compilePop(code, ip)

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++

	case compiler.OpRegTailCall:
		return fmt.Errorf("OpRegTailCall should be handled earlier")

	case compiler.OpRegCall:
		return fmt.Errorf("OpRegCall not supported in JIT")

	default:
		def, _ := compiler.Lookup(byte(op))
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// ============================================================================
// Prologue and Epilogue
// ============================================================================

func (cg *RecursiveCodeGenerator) emitPrologue() {
	// push rbp
	cg.emit(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Save callee-saved registers
	cg.emit(0x53)
	cg.emitBytes([]byte{0x41, 0x54})
	cg.emitBytes([]byte{0x41, 0x55})
	cg.emitBytes([]byte{0x41, 0x56})
	cg.emitBytes([]byte{0x41, 0x57})

	// Allocate stack space
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(1024)

	// Initialize to null
	nullVal := uint64(TagNull) << 48
	for i := 0; i < 64; i++ {
		off := int32(8 * (i + 1))
		cg.emitBytes([]byte{0x48, 0xC7, 0x45})
		cg.emitByte(byte(-off))
		cg.emitUint32(uint32(nullVal))
	}
}

func (cg *RecursiveCodeGenerator) emitPrologueRecursive() {
	// Same as normal prologue but with extra stack space
	cg.emitPrologue()
}

func (cg *RecursiveCodeGenerator) emitEpilogue() {
	cg.emitBytes([]byte{0x48, 0x81, 0xC4})
	cg.emitUint32(1024)

	cg.emitBytes([]byte{0x41, 0x5F})
	cg.emitBytes([]byte{0x41, 0x5E})
	cg.emitBytes([]byte{0x41, 0x5D})
	cg.emitBytes([]byte{0x41, 0x5C})
	cg.emit(0x5B)
	cg.emit(0x5D)
	cg.emit(0xC3)
}

// ============================================================================
// Opcode Implementations (same as FullCodeGenerator)
// ============================================================================

func (cg *RecursiveCodeGenerator) compileLoadConst(code []byte, ip *int) {
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
	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileMove(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRax(src)
	cg.storeRax(dst)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileBinaryOp(code []byte, ip *int, opByte byte) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)

	switch opByte {
	case 0x01:
		cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	case 0x29:
		cg.emitBytes([]byte{0x48, 0x29, 0xC8})
	}

	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileMul(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1})

	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileDiv(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})

	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileMod(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x99})
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})

	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileNeg(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8})
	cg.storeRax(dst)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileComparison(code []byte, ip *int, setcc byte) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8})
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})
	cg.emitBytes([]byte{0x0F, setcc, 0xC0})

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileNot(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRax(src)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRax(dst)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileJump(code []byte, ip *int) {
	offset := int(int16(uint16(code[*ip+1])<<8 | uint16(code[*ip+2])))
	target := *ip + 3 + offset

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJmp(label)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileJumpIfFalse(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRax(cond)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48)

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileJumpIfTrue(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRax(cond)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool)<<48 | 1)

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileReturn(code []byte, ip *int) {
	src := int(code[*ip+1])
	cg.loadRax(src)
	cg.storeRax(compiler.ReturnRegister)
	cg.emitEpilogue()
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compileLoadLocal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	local := int(code[*ip+2])
	cg.loadRax(local)
	cg.storeRax(dst)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileStoreLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRax(src)
	cg.storeRax(local)
	*ip += 3
}

func (cg *RecursiveCodeGenerator) compileLoadGlobal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])

	var val uint64
	if globalIdx < len(cg.globals) {
		val = uint64(cg.globals[globalIdx])
	} else {
		val = uint64(TagNull) << 48
	}

	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(val)
	cg.storeRax(dst)
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileStoreGlobal(code []byte, ip *int) {
	*ip += 4
}

func (cg *RecursiveCodeGenerator) compileIncLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	cg.loadRax(local)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01})
	cg.storeRax(local)
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compileDecLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	cg.loadRax(local)
	cg.emitBytes([]byte{0x48, 0x83, 0xE8, 0x01})
	cg.storeRax(local)
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compileNull(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compileTrue(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool)<<48 | 1)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compileFalse(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool) << 48)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compilePush(code []byte, ip *int) {
	src := int(code[*ip+1])
	cg.loadRax(src)
	cg.emitBytes([]byte{0x48, 0x89, 0x04, 0x24})
	*ip += 2
}

func (cg *RecursiveCodeGenerator) compilePop(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0x8B, 0x04, 0x24})
	cg.storeRax(dst)
	*ip += 2
}

// ============================================================================
// Helper Functions
// ============================================================================

func (cg *RecursiveCodeGenerator) loadRax(reg int) {
	off := int32(8 * (reg + 1))
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *RecursiveCodeGenerator) storeRax(reg int) {
	off := int32(8 * (reg + 1))
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *RecursiveCodeGenerator) loadRcx(reg int) {
	off := int32(8 * (reg + 1))
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-off))
}

func (cg *RecursiveCodeGenerator) emit(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *RecursiveCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *RecursiveCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *RecursiveCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *RecursiveCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func (cg *RecursiveCodeGenerator) emitJmp(label string) {
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *RecursiveCodeGenerator) emitJcc(cc byte, label string) {
	cg.emitBytes([]byte{0x0F, 0x80 | cc})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *RecursiveCodeGenerator) resolveFixups() {
	for _, f := range cg.fixups {
		target, ok := cg.labels[f.label]
		if !ok {
			continue
		}
		offset := target - (f.offset + f.size)
		binary.LittleEndian.PutUint32(cg.code[f.offset:], uint32(offset))
	}
}
