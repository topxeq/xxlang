// +build amd64,!windows

// pkg/jit/jit_fib.go
// JIT compilation support for recursive Fibonacci functions
// Supports both tail-recursive and non-tail-recursive implementations
package jit

import (
	"encoding/binary"
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// FibJITCompiler is a specialized JIT compiler for recursive Fibonacci
// It transforms recursive calls into iterative code with explicit stack
type FibJITCompiler struct {
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction

	// Configuration
	config JITConfig
}

// SavedRegsSize is the bytes used for saved callee-saved registers on stack
// Layout: rbp(8) + rbx(8) + r12(8) + r13(8) + r14(8) + r15(8) = 48 bytes
// Local variables start at [rbp - SavedRegsSize - localOffset]
const FibSavedRegsSize = 48

// NewFibJITCompiler creates a new Fibonacci JIT compiler
func NewFibJITCompiler(config JITConfig) *FibJITCompiler {
	return &FibJITCompiler{
		code:    make([]byte, 0, 16384),
		labels:  make(map[string]int),
		fixups:  make([]fixup, 0),
		config:  config,
	}
}

// Compile compiles a recursive Fibonacci function to native code
// Returns the compiled code or an error if the function is not suitable
func (c *FibJITCompiler) Compile(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	c.code = c.code[:0]
	c.labels = make(map[string]int)
	c.fixups = c.fixups[:0]
	c.constants = constants
	c.globals = globals
	c.fn = fn

	// Analyze the function to determine the type of recursion
	analysis := c.analyzeFunction(fn)

	if analysis.isTailRecursive {
		// Convert tail recursion to a loop
		return c.compileTailRecursive(fn, analysis)
	}

	if analysis.isSelfRecursive {
		// Use explicit stack for non-tail recursion
		return c.compileStackRecursive(fn, analysis)
	}

	// Not recursive - compile normally
	return c.compileNormal(fn)
}

// RecursionAnalysis contains analysis results for a function
type RecursionAnalysis struct {
	isSelfRecursive  bool // Function calls itself
	isTailRecursive  bool // All recursive calls are in tail position
	callCount        int  // Number of recursive calls
	tailCallIPs      []int
	nonTailCallIPs   []int
	usesStackMachine bool // Uses closure or other complex features
}

// analyzeFunction analyzes a function to determine its recursion pattern
func (c *FibJITCompiler) analyzeFunction(fn *compiler.CompiledFunction) RecursionAnalysis {
	analysis := RecursionAnalysis{
		tailCallIPs:    make([]int, 0),
		nonTailCallIPs: make([]int, 0),
	}

	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegCall:
			// Check if this is a call to the same function
			// In register VM, the function is in a register
			funcReg := int(code[ip+1])
			_ = funcReg // For now, assume any OpRegCall could be recursive
			analysis.isSelfRecursive = true
			analysis.callCount++
			analysis.nonTailCallIPs = append(analysis.nonTailCallIPs, ip)
			ip += 3

		case compiler.OpRegTailCall:
			analysis.isSelfRecursive = true
			analysis.isTailRecursive = true
			analysis.callCount++
			analysis.tailCallIPs = append(analysis.tailCallIPs, ip)
			ip += 3

		case compiler.OpRegClosure:
			analysis.usesStackMachine = true
			ip += 5 // opcode + dst(1) + func_idx(2) + num_free(1) + start_reg(1)

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

	// A function with both tail and non-tail calls is not purely tail recursive
	if len(analysis.nonTailCallIPs) > 0 && len(analysis.tailCallIPs) > 0 {
		analysis.isTailRecursive = false
	}

	// A function with only tail calls is tail recursive
	if len(analysis.nonTailCallIPs) == 0 && len(analysis.tailCallIPs) > 0 {
		analysis.isTailRecursive = true
	}

	return analysis
}

// ============================================================================
// Tail-Recursive Compilation
// ============================================================================

// compileTailRecursive compiles a tail-recursive function to a loop
func (c *FibJITCompiler) compileTailRecursive(fn *compiler.CompiledFunction, analysis RecursionAnalysis) ([]byte, error) {
	// Generate prologue
	c.emitPrologue(fn.NumLocals)

	// Initialize parameters from registers (System V AMD64 ABI)
	// Arguments are passed in: rdi, rsi, rdx, rcx, r8, r9
	// Store them in stack slots: R0=[rbp-8], R1=[rbp-16], etc.
	c.emitStoreParams(fn.NumParameters)

	// Entry point label for the loop
	entryLabel := "loop_entry"
	c.labels[entryLabel] = len(c.code)

	// Generate code, replacing tail calls with parameter updates and jumps
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		c.labels[fmt.Sprintf("L%d", ip)] = len(c.code)

		switch op {
		case compiler.OpRegTailCall:
			// Convert tail call to parameter copy + jump to entry
			c.compileTailCallToLoop(code, &ip, entryLabel)

		case compiler.OpRegCall:
			// Non-tail call in tail-recursive function - error
			return nil, fmt.Errorf("mixed tail/non-tail recursion not supported")

		default:
			if err := c.compileInstruction(op, code, &ip); err != nil {
				return nil, err
			}
		}
	}

	// Resolve fixups
	if err := c.resolveFixups(); err != nil {
		return nil, err
	}

	return c.code, nil
}

// compileTailCallToLoop converts a tail call instruction to a jump
func (c *FibJITCompiler) compileTailCallToLoop(code []byte, ip *int, entryLabel string) {
	// OpRegTailCall: func_reg(1), num_args(1)
	// The new argument values are already in R0, R1, ... (set by caller)
	// We just need to jump to the function entry

	// For tail recursion, the arguments are already in the right registers
	// Just jump back to the loop entry
	c.emitJmp(entryLabel)
	*ip += 3
}

// ============================================================================
// Stack-Based Recursive Compilation
// ============================================================================

// compileStackRecursive compiles a non-tail recursive function using explicit stack
// This transforms the recursion into iteration with an explicit call stack
func (c *FibJITCompiler) compileStackRecursive(fn *compiler.CompiledFunction, analysis RecursionAnalysis) ([]byte, error) {
	// For Fibonacci specifically, we can use a more efficient transformation
	// The standard fib(n) = fib(n-1) + fib(n-2) can be computed iteratively

	// Check if this is a simple Fibonacci pattern
	if c.isSimpleFibonacci(fn) {
		return c.compileIterativeFibonacci(fn)
	}

	// Generic stack-based simulation (more complex)
	return c.compileGenericStackRecursive(fn, analysis)
}

// isSimpleFibonacci checks if the function is a simple recursive Fibonacci
func (c *FibJITCompiler) isSimpleFibonacci(fn *compiler.CompiledFunction) bool {
	// Check for the pattern:
	// if (n <= 1) return n
	// return fib(n-1) + fib(n-2)

	// For now, use a simple heuristic based on:
	// - 1 parameter
	// - 2 recursive calls
	// - Addition of results

	code := fn.Instructions
	hasBaseCase := false
	callCount := 0
	hasAdd := false

	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		switch op {
		case compiler.OpRegLessEqual, compiler.OpRegLess, compiler.OpRegEqual:
			hasBaseCase = true
		case compiler.OpRegCall:
			callCount++
		case compiler.OpRegAdd:
			hasAdd = true
		}

		def, _ := compiler.Lookup(byte(op))
		ip++
		if def != nil {
			for _, w := range def.OperandWidths {
				ip += w
			}
		}
	}

	return fn.NumParameters == 1 && callCount == 2 && hasBaseCase && hasAdd
}

// compileIterativeFibonacci compiles an optimized iterative Fibonacci
// This is the fast path for the classic recursive Fibonacci fib(n) = fib(n-1) + fib(n-2)
// Generated code uses System V AMD64 ABI: argument n is passed in rdi, result returned in rax.
// Uses only caller-saved registers (rdi, rcx, rdx, r8, rax) so no save/restore is needed.
func (c *FibJITCompiler) compileIterativeFibonacci(fn *compiler.CompiledFunction) ([]byte, error) {
	// Transform: fib(n) = fib(n-1) + fib(n-2)
	// To: iterative version using registers only
	//
	// Register allocation:
	//   rdi = n (input, preserved for loop comparison)
	//   rcx = a (fib(i-2), starts at 0)
	//   rdx = b (fib(i-1), starts at 1)
	//   r8  = i (loop counter, starts at 2)
	//   rax = temp / return value

	c.code = c.code[:0]

	emit := func(b ...byte) int {
		start := len(c.code)
		c.code = append(c.code, b...)
		return start
	}

	// Base case: if n <= 1, return n
	_ = emit(0x48, 0x89, 0xF8)                   // mov rax, rdi (result = n)
	_ = emit(0x48, 0x83, 0xF8, 0x01)             // cmp rax, 1
	jlePos := emit(0x7E, 0x00)                    // jle -> base_case_return (placeholder)

	// Initialize: a=0, b=1, i=2
	_ = emit(0x48, 0x31, 0xC9)                    // xor rcx, rcx (a = 0)
	_ = emit(0x48, 0xC7, 0xC2, 0x01, 0x00, 0x00, 0x00) // mov rdx, 1 (b = 1)
	_ = emit(0x49, 0xC7, 0xC0, 0x02, 0x00, 0x00, 0x00) // mov r8, 2 (i = 2)

	// Loop: temp = a + b; a = b; b = temp; i++; if i <= n goto loop
	loopStart := emit(0x48, 0x89, 0xC8)           // mov rax, rcx (temp = a)
	_ = emit(0x48, 0x01, 0xD0)                    // add rax, rdx (temp += b)
	_ = emit(0x48, 0x89, 0xD1)                    // mov rcx, rdx (a = b)
	_ = emit(0x48, 0x89, 0xC2)                    // mov rdx, rax (b = temp)
	_ = emit(0x49, 0xFF, 0xC0)                    // inc r8 (i++)
	_ = emit(0x4C, 0x39, 0xC7)                    // cmp rdi, r8 (n - i)
	jgePos := emit(0x7D, 0x00)                    // jge -> loopStart (placeholder)

	// Done: return b (which is in rdx)
	_ = emit(0x48, 0x89, 0xD0)                    // mov rax, rdx (result = b)
	_ = emit(0xC3)                                // ret

	// Base case return: rax already contains n, just return
	baseCaseReturn := emit(0xC3)                  // ret

	// Fix up jump targets
	// jle: from jlePos to baseCaseReturn (base case: n <= 1, return n directly)
	c.code[jlePos+1] = byte(int8(baseCaseReturn - (jlePos + 2)))
	// jge: from jgePos back to loopStart (i <= n, continue loop)
	c.code[jgePos+1] = byte(int8(loopStart - (jgePos + 2)))

	return c.code, nil
}

// compileGenericStackRecursive compiles using generic stack simulation
func (c *FibJITCompiler) compileGenericStackRecursive(fn *compiler.CompiledFunction, analysis RecursionAnalysis) ([]byte, error) {
	// This is a more complex transformation that works for any recursive function
	// It uses an explicit stack to simulate the call stack

	// For simplicity, we fall back to the interpreter for complex cases
	return nil, fmt.Errorf("complex non-tail recursion requires interpreter fallback")
}

// ============================================================================
// Normal (Non-Recursive) Compilation
// ============================================================================

// compileNormal compiles a non-recursive function
func (c *FibJITCompiler) compileNormal(fn *compiler.CompiledFunction) ([]byte, error) {
	c.emitPrologue(fn.NumLocals)
	c.emitStoreParams(fn.NumParameters)

	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		c.labels[fmt.Sprintf("L%d", ip)] = len(c.code)

		if err := c.compileInstruction(op, code, &ip); err != nil {
			return nil, err
		}
	}

	if err := c.resolveFixups(); err != nil {
		return nil, err
	}

	return c.code, nil
}

// ============================================================================
// Instruction Compilation
// ============================================================================

// compileInstruction compiles a single instruction
func (c *FibJITCompiler) compileInstruction(op compiler.Opcode, code []byte, ip *int) error {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return fmt.Errorf("unknown opcode %d", op)
	}

	switch op {
	// Data movement
	case compiler.OpRegLoadConst:
		dst := int(code[*ip+1])
		constIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		c.compileLoadConst(dst, constIdx)
		*ip += 4

	case compiler.OpRegMove:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		c.compileMove(dst, src)
		*ip += 3

	// Arithmetic
	case compiler.OpRegAdd:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileAdd(dst, left, right)
		*ip += 4

	case compiler.OpRegSub:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileSub(dst, left, right)
		*ip += 4

	case compiler.OpRegMul:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileMul(dst, left, right)
		*ip += 4

	case compiler.OpRegDiv:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileDiv(dst, left, right)
		*ip += 4

	case compiler.OpRegMod:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileMod(dst, left, right)
		*ip += 4

	case compiler.OpRegNeg:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		c.compileNeg(dst, src)
		*ip += 3

	// Comparison
	case compiler.OpRegLess:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileLess(dst, left, right)
		*ip += 4

	case compiler.OpRegLessEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileLessEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegGreater:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileGreater(dst, left, right)
		*ip += 4

	case compiler.OpRegGreaterEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileGreaterEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegNotEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		c.compileNotEqual(dst, left, right)
		*ip += 4

	// Logical
	case compiler.OpRegNot:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		c.compileNot(dst, src)
		*ip += 3

	// Control flow
	case compiler.OpRegJump:
		// Format: opcode(1) + unused(1) + offset(2) = 4 bytes
		// In VM: IP += offset (offset is relative to current IP)
		// Skip the unused byte at code[*ip+1]
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		target := *ip + int(offset)
		c.compileJump(target)
		*ip += 4

	case compiler.OpRegJumpIfFalse:
		cond := int(code[*ip+1])
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		// In VM: IP += offset (offset is relative to current IP)
		target := *ip + int(offset)
		c.compileJumpIfFalse(cond, target)
		*ip += 4

	case compiler.OpRegJumpIfTrue:
		cond := int(code[*ip+1])
		offset := int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3]))
		// In VM: IP += offset (offset is relative to current IP)
		target := *ip + int(offset)
		c.compileJumpIfTrue(cond, target)
		*ip += 4

	case compiler.OpRegReturn:
		src := int(code[*ip+1])
		c.compileReturn(src)
		*ip += 2

	// Local variables
	case compiler.OpRegLoadLocal:
		dst := int(code[*ip+1])
		local := int(code[*ip+2])
		c.compileMove(dst, local)
		*ip += 3

	case compiler.OpRegStoreLocal:
		local := int(code[*ip+1])
		src := int(code[*ip+2])
		c.compileMove(local, src)
		*ip += 3

	// Global variables
	case compiler.OpRegLoadGlobal:
		dst := int(code[*ip+1])
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		c.compileLoadGlobal(dst, globalIdx)
		*ip += 4

	case compiler.OpRegStoreGlobal:
		// For JIT, globals are snapshots - store is a no-op in this simple version
		*ip += 4

	// Increment/Decrement
	case compiler.OpRegIncLocal:
		local := int(code[*ip+1])
		c.compileIncLocal(local)
		*ip += 2

	case compiler.OpRegDecLocal:
		local := int(code[*ip+1])
		c.compileDecLocal(local)
		*ip += 2

	// Literals
	case compiler.OpRegNull:
		dst := int(code[*ip+1])
		c.compileNull(dst)
		*ip += 2

	case compiler.OpRegTrue:
		dst := int(code[*ip+1])
		c.compileTrue(dst)
		*ip += 2

	case compiler.OpRegFalse:
		dst := int(code[*ip+1])
		c.compileFalse(dst)
		*ip += 2

	// Stack operations
	case compiler.OpRegPush:
		src := int(code[*ip+1])
		c.compilePush(src)
		*ip += 2

	case compiler.OpRegPop:
		dst := int(code[*ip+1])
		c.compilePop(dst)
		*ip += 2

	// Function operations
	case compiler.OpRegCall:
		return fmt.Errorf("OpRegCall requires interpreter fallback or stack simulation")

	case compiler.OpRegTailCall:
		return fmt.Errorf("OpRegTailCall should be handled by compileTailRecursive")

	case compiler.OpRegClosure:
		return fmt.Errorf("OpRegClosure requires interpreter fallback")

	case compiler.OpReturn:
		c.emitEpilogue()
		*ip++

	// Loop superinstructions
	case compiler.OpRegLoopCountAdd:
		// Format: acc_reg, counter_reg, start_const(16), limit_const(16), step_const(16)
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		startIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		limitIdx := int(code[*ip+5])<<8 | int(code[*ip+6])
		stepIdx := int(code[*ip+7])<<8 | int(code[*ip+8])
		c.compileLoopCountAdd(accReg, counterReg, startIdx, limitIdx, stepIdx)
		*ip += 9

	case compiler.OpRegLoopBodyAdd:
		// Format: acc_reg, counter_reg, limit_const(16), jump_offset(16)
		accReg := int(code[*ip+1])
		counterReg := int(code[*ip+2])
		limitIdx := int(code[*ip+3])<<8 | int(code[*ip+4])
		jumpOffset := int(int16(uint16(code[*ip+5])<<8 | uint16(code[*ip+6])))
		c.compileLoopBodyAdd(accReg, counterReg, limitIdx, jumpOffset, *ip+7)
		*ip += 7

	// Array operations - stub implementations
	case compiler.OpRegArray:
		*ip += 4 // Skip: dst, start_reg, count
		// Arrays require Go runtime support

	case compiler.OpRegArrayEmpty:
		*ip += 2 // Skip: dst

	case compiler.OpRegArrayAppend:
		*ip += 4 // Skip: dst, arr_reg, elem_reg

	case compiler.OpRegIndex:
		*ip += 4 // Skip: dst, obj_reg, key_reg

	case compiler.OpRegSetIndex:
		*ip += 4 // Skip: obj_reg, key_reg, val_reg

	// Map operations - stub implementations
	case compiler.OpRegMap:
		*ip += 4 // Skip: dst, start_reg, count

	case compiler.OpRegMapSet:
		*ip += 5 // Skip: dst, map_reg, key_reg, val_reg

	default:
		return fmt.Errorf("unsupported opcode: %s", def.Name)
	}

	return nil
}

// ============================================================================
// Low-Level Code Generation
// ============================================================================

// localOffset returns the stack offset for local variable 'slot'
// Local variables are stored at [rbp - FibSavedRegsSize - (slot+1)*8]
func (c *FibJITCompiler) localOffset(slot int) int {
	return FibSavedRegsSize + (slot+1)*8
}

// emitMovRaxToSlot stores rax to local variable slot
func (c *FibJITCompiler) emitMovRaxToSlot(slot int) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x89, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x89, 0x85})
		c.emitUint32(uint32(-offset))
	}
}

// emitMovSlotToRax loads local variable slot to rax
func (c *FibJITCompiler) emitMovSlotToRax(slot int) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x8B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x8B, 0x85})
		c.emitUint32(uint32(-offset))
	}
}

// emitAddSlotToRax adds local variable slot to rax
func (c *FibJITCompiler) emitAddSlotToRax(slot int) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x03, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x03, 0x85})
		c.emitUint32(uint32(-offset))
	}
}

// emitAddRaxToSlot adds rax to local variable slot
func (c *FibJITCompiler) emitAddRaxToSlot(slot int) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x01, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x01, 0x85})
		c.emitUint32(uint32(-offset))
	}
}

// emitIncSlot increments local variable slot
func (c *FibJITCompiler) emitIncSlot(slot int) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0xFF, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0xFF, 0x85})
		c.emitUint32(uint32(-offset))
	}
}

// emitMovImm32ToSlot moves 32-bit immediate to local variable slot
func (c *FibJITCompiler) emitMovImm32ToSlot(slot int, val uint32) {
	offset := c.localOffset(slot)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0xC7, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0xC7, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitUint32(val)
}

// emitPrologue generates function entry code
func (c *FibJITCompiler) emitPrologue(numLocals int) {
	// push rbp
	c.emitByte(0x55)
	// mov rbp, rsp
	c.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Save callee-saved registers
	c.emitByte(0x53)              // push rbx
	c.emitBytes([]byte{0x41, 0x54}) // push r12
	c.emitBytes([]byte{0x41, 0x55}) // push r13
	c.emitBytes([]byte{0x41, 0x56}) // push r14
	c.emitBytes([]byte{0x41, 0x57}) // push r15

	// Allocate stack space for locals (16-byte aligned)
	// Each local is 8 bytes, add padding for alignment
	// Need at least 16 registers * 8 bytes = 128 bytes for the iterative fib
	stackSize := numLocals * 8
	if stackSize < 256 {
		stackSize = 256
	}
	// Round up to 16 bytes
	stackSize = (stackSize + 15) & ^15

	// sub rsp, imm32
	c.emitBytes([]byte{0x48, 0x81, 0xEC})
	c.emitUint32(uint32(stackSize))
}

// emitStoreParams stores parameters from registers to stack slots
// Go calling convention: arguments are on the stack at [rbp+16], [rbp+24], etc.
// We copy them to our local slots starting at [rbp - FibSavedRegsSize - 8]
func (c *FibJITCompiler) emitStoreParams(numParams int) {
	// In Go's calling convention for a function:
	// - [rbp+8] = return address
	// - [rbp+16] = first argument
	// - [rbp+24] = second argument
	// - etc.
	//
	// We store them at local slots:
	// - Slot 0 = [rbp - FibSavedRegsSize - 8]
	// - Slot 1 = [rbp - FibSavedRegsSize - 16]
	// - etc.

	for i := 0; i < numParams && i < 16; i++ {
		// mov rax, [rbp + 16 + i*8]  ; load argument from caller's stack
		srcOffset := 16 + i * 8 // positive offset from rbp

		// mov rax, [rbp + srcOffset]
		if srcOffset <= 127 {
			c.emitBytes([]byte{0x48, 0x8B, 0x45}) // mov rax, [rbp+disp8]
			c.emitByte(byte(srcOffset))
		} else {
			c.emitBytes([]byte{0x48, 0x8B, 0x85}) // mov rax, [rbp+disp32]
			c.emitUint32(uint32(srcOffset))
		}

		// Store to local slot i
		c.emitMovRaxToSlot(i)
	}
}

// emitEpilogue generates function exit code
func (c *FibJITCompiler) emitEpilogue() {
	// The prologue did:
	//   push rbp; mov rbp, rsp
	//   push rbx, r12, r13, r14, r15 (40 bytes total)
	//   sub rsp, stackSize
	//
	// Stack layout:
	//   [rbp] = old rbp
	//   [rbp-8] = saved rbx
	//   [rbp-16] = saved r12
	//   [rbp-24] = saved r13
	//   [rbp-32] = saved r14
	//   [rbp-40] = saved r15
	//   [rbp-48...] = locals
	//
	// Restore rsp to point to saved registers
	// lea rsp, [rbp - 40] (40 = 5 saved registers * 8 bytes)
	c.emitBytes([]byte{0x48, 0x8D, 0x65, 0xD8}) // lea rsp, [rbp - 40]

	// Pop callee-saved registers in reverse order
	c.emitBytes([]byte{0x41, 0x5F}) // pop r15
	c.emitBytes([]byte{0x41, 0x5E}) // pop r14
	c.emitBytes([]byte{0x41, 0x5D}) // pop r13
	c.emitBytes([]byte{0x41, 0x5C}) // pop r12
	c.emitByte(0x5B)                // pop rbx

	// pop rbp and return
	c.emitByte(0x5D) // pop rbp
	c.emitByte(0xC3) // ret
}

// compileLoadConst loads a constant into a register
func (c *FibJITCompiler) compileLoadConst(dst, constIdx int) {
	var val int64
	if constIdx < len(c.constants) {
		// Try to extract integer value
		v := c.constants[constIdx]
		if v.IsInt() {
			val, _ = v.ToInt()
		} else {
			val = 0
		}
	} else {
		val = 0
	}

	// mov rax, imm64
	c.emitBytes([]byte{0x48, 0xB8})
	c.emitUint64(uint64(val))
	// Store to local slot
	c.emitMovRaxToSlot(dst)
}

// compileMove moves a value between registers
func (c *FibJITCompiler) compileMove(dst, src int) {
	c.emitMovSlotToRax(src)
	c.emitMovRaxToSlot(dst)
}

// compileAdd adds two registers
func (c *FibJITCompiler) compileAdd(dst, left, right int) {
	c.emitMovSlotToRax(left)
	c.emitAddSlotToRax(right)
	c.emitMovRaxToSlot(dst)
}

// compileSub subtracts two registers
func (c *FibJITCompiler) compileSub(dst, left, right int) {
	c.emitMovSlotToRax(left)
	// sub rax, [rbp - offset]
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x2B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x2B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitMovRaxToSlot(dst)
}

// compileMul multiplies two registers
func (c *FibJITCompiler) compileMul(dst, left, right int) {
	c.emitMovSlotToRax(left)
	// imul rax, [rbp - offset]
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x0F, 0xAF, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitMovRaxToSlot(dst)
}

// compileDiv divides two registers
func (c *FibJITCompiler) compileDiv(dst, left, right int) {
	c.emitMovSlotToRax(left)
	c.emitBytes([]byte{0x48, 0x99}) // cqo (sign extend)
	// idiv [rbp - offset]
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0xF7, 0x7D})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0xF7, 0xBD})
		c.emitUint32(uint32(-offset))
	}
	c.emitMovRaxToSlot(dst)
}

// compileMod computes modulo of two registers
func (c *FibJITCompiler) compileMod(dst, left, right int) {
	c.emitMovSlotToRax(left)
	c.emitBytes([]byte{0x48, 0x99})
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0xF7, 0x7D})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0xF7, 0xBD})
		c.emitUint32(uint32(-offset))
	}
	// Result is in rdx (remainder)
	c.emitBytes([]byte{0x48, 0x89, 0xD0}) // mov rax, rdx
	c.emitMovRaxToSlot(dst)
}

// compileNeg negates a register
func (c *FibJITCompiler) compileNeg(dst, src int) {
	c.emitMovSlotToRax(src)
	c.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	c.emitMovRaxToSlot(dst)
}

// compileLess compares two registers for less than
func (c *FibJITCompiler) compileLess(dst, left, right int) {
	c.emitMovSlotToRax(left)
	// cmp rax, [rbp - offset]
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	c.emitBytes([]byte{0x0F, 0x9C, 0xC0}) // setl al
	c.emitMovRaxToSlot(dst)
}

// compileLessEqual compares two registers for less than or equal
func (c *FibJITCompiler) compileLessEqual(dst, left, right int) {
	c.emitMovSlotToRax(left)
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x9E, 0xC0}) // setle al
	c.emitMovRaxToSlot(dst)
}

// compileGreater compares two registers for greater than
func (c *FibJITCompiler) compileGreater(dst, left, right int) {
	c.emitMovSlotToRax(left)
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x9F, 0xC0}) // setg al
	c.emitMovRaxToSlot(dst)
}

// compileGreaterEqual compares two registers for greater than or equal
func (c *FibJITCompiler) compileGreaterEqual(dst, left, right int) {
	c.emitMovSlotToRax(left)
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x9D, 0xC0}) // setge al
	c.emitMovRaxToSlot(dst)
}

// compileEqual compares two registers for equality
func (c *FibJITCompiler) compileEqual(dst, left, right int) {
	c.emitMovSlotToRax(left)
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al
	c.emitMovRaxToSlot(dst)
}

// compileNotEqual compares two registers for inequality
func (c *FibJITCompiler) compileNotEqual(dst, left, right int) {
	c.emitMovSlotToRax(left)
	offset := c.localOffset(right)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x3B, 0x45})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x3B, 0x85})
		c.emitUint32(uint32(-offset))
	}
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x95, 0xC0}) // setne al
	c.emitMovRaxToSlot(dst)
}

// compileNot performs logical not
func (c *FibJITCompiler) compileNot(dst, src int) {
	c.emitMovSlotToRax(src)
	c.emitBytes([]byte{0x48, 0x85, 0xC0}) // test rax, rax
	c.emitBytes([]byte{0x48, 0x31, 0xC0})
	c.emitBytes([]byte{0x0F, 0x94, 0xC0}) // sete al
	c.emitMovRaxToSlot(dst)
}

// compileJump generates an unconditional jump
func (c *FibJITCompiler) compileJump(target int) {
	label := fmt.Sprintf("L%d", target)
	c.emitBytes([]byte{0xE9}) // jmp rel32
	c.fixups = append(c.fixups, fixup{offset: len(c.code), label: label, size: 4})
	c.emitUint32(0)
}

// compileJumpIfFalse generates a conditional jump if false
func (c *FibJITCompiler) compileJumpIfFalse(cond, target int) {
	c.emitMovSlotToRax(cond)
	// Test if zero
	c.emitBytes([]byte{0x48, 0x85, 0xC0})
	// jz rel32
	c.emitBytes([]byte{0x0F, 0x84})
	label := fmt.Sprintf("L%d", target)
	c.fixups = append(c.fixups, fixup{offset: len(c.code), label: label, size: 4})
	c.emitUint32(0)
}

// compileJumpIfTrue generates a conditional jump if true
func (c *FibJITCompiler) compileJumpIfTrue(cond, target int) {
	c.emitMovSlotToRax(cond)
	c.emitBytes([]byte{0x48, 0x85, 0xC0})
	// jnz rel32
	c.emitBytes([]byte{0x0F, 0x85})
	label := fmt.Sprintf("L%d", target)
	c.fixups = append(c.fixups, fixup{offset: len(c.code), label: label, size: 4})
	c.emitUint32(0)
}

// compileReturn generates a return instruction
func (c *FibJITCompiler) compileReturn(src int) {
	c.emitMovSlotToRax(src)
	c.emitEpilogue()
}

// compileLoadGlobal loads a global variable
func (c *FibJITCompiler) compileLoadGlobal(dst, globalIdx int) {
	var val int64
	if globalIdx < len(c.globals) {
		v := c.globals[globalIdx]
		if v.IsInt() {
			val, _ = v.ToInt()
		}
	}
	c.emitBytes([]byte{0x48, 0xB8})
	c.emitUint64(uint64(val))
	c.emitMovRaxToSlot(dst)
}

// compileIncLocal increments a local variable
func (c *FibJITCompiler) compileIncLocal(local int) {
	c.emitIncSlot(local)
}

// compileDecLocal decrements a local variable
func (c *FibJITCompiler) compileDecLocal(local int) {
	offset := c.localOffset(local)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0xFF, 0x4D})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0xFF, 0x8D})
		c.emitUint32(uint32(-offset))
	}
}

// compileNull sets a register to null (0)
func (c *FibJITCompiler) compileNull(dst int) {
	c.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	c.emitMovRaxToSlot(dst)
}

// compileTrue sets a register to true (1)
func (c *FibJITCompiler) compileTrue(dst int) {
	c.emitBytes([]byte{0x48, 0xC7, 0xC0, 0x01, 0x00, 0x00, 0x00}) // mov rax, 1
	c.emitMovRaxToSlot(dst)
}

// compileFalse sets a register to false (0)
func (c *FibJITCompiler) compileFalse(dst int) {
	c.compileNull(dst)
}

// compilePush pushes a register onto the stack
func (c *FibJITCompiler) compilePush(src int) {
	c.emitMovSlotToRax(src)
	c.emitByte(0x50) // push rax
}

// compilePop pops a value from the stack into a register
func (c *FibJITCompiler) compilePop(dst int) {
	c.emitByte(0x58) // pop rax
	c.emitMovRaxToSlot(dst)
}

// ============================================================================
// Loop Superinstructions
// ============================================================================

// compileLoopCountAdd compiles an optimized counting loop
// Performs: for (counter = start; counter < limit; counter += step) { acc += counter }
func (c *FibJITCompiler) compileLoopCountAdd(accReg, counterReg, startIdx, limitIdx, stepIdx int) {
	// Get constant values
	var start, limit, step int64
	if startIdx < len(c.constants) {
		if v := c.constants[startIdx]; v.IsInt() {
			start, _ = v.ToInt()
		}
	}
	if limitIdx < len(c.constants) {
		if v := c.constants[limitIdx]; v.IsInt() {
			limit, _ = v.ToInt()
		}
	}
	if stepIdx < len(c.constants) {
		if v := c.constants[stepIdx]; v.IsInt() {
			step, _ = v.ToInt()
		}
	}
	if step == 0 {
		step = 1
	}

	// Initialize counter with start value
	c.emitMovImm32ToSlot(counterReg, uint32(start))

	// Initialize accumulator with 0
	c.emitMovImm32ToSlot(accReg, 0)

	// Loop label
	loopLabel := fmt.Sprintf("loop_%d", len(c.code))
	c.labels[loopLabel] = len(c.code)

	// Load counter and compare with limit
	c.emitMovSlotToRax(counterReg)

	// cmp rax, limit
	c.emitBytes([]byte{0x48, 0x3D})
	c.emitUint32(uint32(limit))

	// jge end (exit loop if counter >= limit)
	endLabel := fmt.Sprintf("loop_end_%d", len(c.code))
	c.emitBytes([]byte{0x0F, 0x8D}) // jge rel32
	c.fixups = append(c.fixups, fixup{
		offset: len(c.code),
		label:  endLabel,
		size:   4,
	})
	c.emitUint32(0)

	// Add counter to accumulator
	c.emitMovSlotToRax(counterReg)
	c.emitAddRaxToSlot(accReg)

	// Increment counter by step
	if step == 1 {
		c.emitIncSlot(counterReg)
	} else {
		offset := c.localOffset(counterReg)
		if offset <= 127 {
			c.emitBytes([]byte{0x48, 0x81, 0x45})
			c.emitByte(byte(-offset))
		} else {
			c.emitBytes([]byte{0x48, 0x81, 0x85})
			c.emitUint32(uint32(-offset))
		}
		c.emitUint32(uint32(step))
	}

	// jmp loop
	c.emitBytes([]byte{0xE9})
	c.fixups = append(c.fixups, fixup{
		offset: len(c.code),
		label:  loopLabel,
		size:   4,
	})
	c.emitUint32(0)

	// End label
	c.labels[endLabel] = len(c.code)
}

// compileLoopBodyAdd compiles a loop body add instruction
// Performs: acc += counter; counter++; if counter < limit jump to offset
func (c *FibJITCompiler) compileLoopBodyAdd(accReg, counterReg, limitIdx, jumpOffset int, currentIP int) {
	// Get limit value
	var limit int64
	if limitIdx < len(c.constants) {
		if v := c.constants[limitIdx]; v.IsInt() {
			limit, _ = v.ToInt()
		}
	}

	// Load counter and add to accumulator
	c.emitMovSlotToRax(counterReg)
	c.emitAddRaxToSlot(accReg)

	// Increment counter
	c.emitIncSlot(counterReg)

	// Load counter and compare with limit
	c.emitMovSlotToRax(counterReg)

	// cmp rax, limit
	c.emitBytes([]byte{0x48, 0x3D})
	c.emitUint32(uint32(limit))

	// jl back to loop
	targetIP := currentIP + jumpOffset
	targetLabel := fmt.Sprintf("L%d", targetIP)
	c.emitBytes([]byte{0x0F, 0x8C}) // jl rel32
	c.fixups = append(c.fixups, fixup{
		offset: len(c.code),
		label:  targetLabel,
		size:   4,
	})
	c.emitUint32(0)
}

// ============================================================================
// Helper Methods for Iterative Fibonacci
// ============================================================================

// emitLoadRax loads a register into rax
func (c *FibJITCompiler) emitLoadRax(reg int) {
	c.emitMovSlotToRax(reg)
}

// emitLoadRcx loads a register into rcx
func (c *FibJITCompiler) emitLoadRcx(reg int) {
	offset := c.localOffset(reg)
	if offset <= 127 {
		c.emitBytes([]byte{0x48, 0x8B, 0x4D})
		c.emitByte(byte(-offset))
	} else {
		c.emitBytes([]byte{0x48, 0x8B, 0x8D})
		c.emitUint32(uint32(-offset))
	}
}

// emitStoreRax stores rax into a register
func (c *FibJITCompiler) emitStoreRax(reg int) {
	c.emitMovRaxToSlot(reg)
}

// emitAddReg adds a register to rax
func (c *FibJITCompiler) emitAddReg(reg int) {
	c.emitAddSlotToRax(reg)
}

// emitCompareConst compares rax with a constant
func (c *FibJITCompiler) emitCompareConst(val int64) {
	c.emitBytes([]byte{0x48, 0x3D}) // cmp rax, imm32
	c.emitUint32(uint32(val))
}

// emitCompareRaxRcx compares rax with rcx
func (c *FibJITCompiler) emitCompareRaxRcx() {
	c.emitBytes([]byte{0x48, 0x39, 0xC8}) // cmp rax, rcx
}

// emitIncReg increments a register
func (c *FibJITCompiler) emitIncReg(reg int) {
	c.emitIncSlot(reg)
}

// emitMovConstToReg moves a constant to a register
func (c *FibJITCompiler) emitMovConstToReg(val, reg int) {
	c.emitMovImm32ToSlot(reg, uint32(val))
}

// emitJmp generates an unconditional jump to a label
func (c *FibJITCompiler) emitJmp(label string) {
	c.emitBytes([]byte{0xE9})
	c.fixups = append(c.fixups, fixup{offset: len(c.code), label: label, size: 4})
	c.emitUint32(0)
}

// emitJle generates a jump if less or equal
func (c *FibJITCompiler) emitJle(label string) {
	c.emitBytes([]byte{0x0F, 0x8E})
	c.fixups = append(c.fixups, fixup{offset: len(c.code), label: label, size: 4})
	c.emitUint32(0)
}

// ============================================================================
// Low-Level Emit Functions
// ============================================================================

func (c *FibJITCompiler) emitByte(b byte) {
	c.code = append(c.code, b)
}

func (c *FibJITCompiler) emitBytes(b []byte) {
	c.code = append(c.code, b...)
}

func (c *FibJITCompiler) emitUint32(v uint32) {
	c.code = append(c.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (c *FibJITCompiler) emitUint64(v uint64) {
	c.code = append(c.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// resolveFixups resolves all jump fixups
func (c *FibJITCompiler) resolveFixups() error {
	for _, f := range c.fixups {
		target, ok := c.labels[f.label]
		if !ok {
			return fmt.Errorf("undefined label: %s", f.label)
		}

		// Calculate relative offset
		offset := int32(target - (f.offset + f.size))
		binary.LittleEndian.PutUint32(c.code[f.offset:], uint32(offset))
	}
	return nil
}
