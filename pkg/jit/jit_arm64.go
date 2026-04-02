//go:build arm64
// +build arm64

// pkg/jit/jit_arm64.go
// JIT Compiler for Xxlang VM on ARM64 platforms
// Generates ARM64 machine code at runtime
package jit

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
	"syscall"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITCompiler handles JIT compilation of VM bytecode on ARM64
type JITCompiler struct {
	config JITConfig

	// Executable memory pools
	codePages []*CodePage
	pageLock  sync.Mutex

	// Function cache: bytecode hash -> compiled function
	funcCache sync.Map // map[uint64]*CompiledFunc

	// Hot path counters
	hotPaths sync.Map // map[uint64]int

	// Statistics
	stats JITStats
}

// CompiledFunc represents a JIT-compiled function
type CompiledFunc struct {
	Entry     uintptr
	Page      *CodePage
	Size      int
	Hash      uint64
	NumRegs   int
	NumParams int
}

// CodePage represents a page of executable memory
type CodePage struct {
	Data []byte
	Used int
}

// NewJITCompiler creates a new JIT compiler for ARM64
func NewJITCompiler(config JITConfig) *JITCompiler {
	return &JITCompiler{
		config: config,
	}
}

// allocCodePageLocked allocates a new executable code page
// Must be called with pageLock held
func (j *JITCompiler) allocCodePageLocked() (*CodePage, error) {
	size := 64 * 1024 // 64KB pages

	// Allocate executable memory using mmap
	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	flags := syscall.MAP_ANON | syscall.MAP_PRIVATE

	data, err := syscall.Mmap(-1, 0, size, prot, flags)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate code page: %w", err)
	}

	page := &CodePage{
		Data: data,
		Used: 0,
	}

	j.codePages = append(j.codePages, page)
	return page, nil
}

// AllocCode allocates code in an executable page
func (j *JITCompiler) AllocCode(size int) ([]byte, *CodePage, error) {
	j.pageLock.Lock()
	defer j.pageLock.Unlock()

	// Find a page with enough space
	for _, page := range j.codePages {
		if len(page.Data)-page.Used >= size {
			start := page.Used
			page.Used += size
			return page.Data[start : start+size], page, nil
		}
	}

	// Need a new page
	page, err := j.allocCodePageLocked()
	if err != nil {
		return nil, nil, err
	}

	start := 0
	page.Used = size
	return page.Data[start:size], page, nil
}

// hashBytecode computes a hash of bytecode for caching
func hashBytecode(code []byte) uint64 {
	h := fnv.New64a()
	h.Write(code)
	return h.Sum64()
}

// RecordExecution records an execution for hot path detection
func (j *JITCompiler) RecordExecution(fn *compiler.CompiledFunction) bool {
	hash := hashBytecode(fn.Instructions)

	count, _ := j.hotPaths.LoadOrStore(hash, 0)
	newCount := count.(int) + 1
	j.hotPaths.Store(hash, newCount)

	return newCount == j.config.HotThreshold
}

// ShouldCompile returns true if the function should be JIT compiled
func (j *JITCompiler) ShouldCompile(fn *compiler.CompiledFunction) bool {
	if len(fn.Instructions) > j.config.MaxCodeSize {
		return false
	}

	hash := hashBytecode(fn.Instructions)
	count, ok := j.hotPaths.Load(hash)
	if !ok {
		return false
	}

	return count.(int) >= j.config.HotThreshold
}

// GetCompiled returns a cached compiled function
func (j *JITCompiler) GetCompiled(fn *compiler.CompiledFunction) *CompiledFunc {
	hash := hashBytecode(fn.Instructions)
	if cf, ok := j.funcCache.Load(hash); ok {
		j.stats.CacheHits++
		return cf.(*CompiledFunc)
	}
	j.stats.CacheMisses++
	return nil
}

// Compile compiles a function to ARM64 native code
func (j *JITCompiler) Compile(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) (*CompiledFunc, error) {
	// Check cache first
	if cf := j.GetCompiled(fn); cf != nil {
		return cf, nil
	}

	// Use ARM64 code generator
	cg := NewARM64CodeGenerator()
	code, err := cg.Generate(fn, constants, globals)
	if err != nil {
		if j.config.Debug {
			fmt.Printf("[JIT-ARM64] Code generation failed: %v\n", err)
		}
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	// Allocate executable memory
	mem, page, err := j.AllocCode(len(code))
	if err != nil {
		return nil, fmt.Errorf("memory allocation failed: %w", err)
	}

	// Copy code to executable memory
	copy(mem, code)

	// Create compiled function
	cf := &CompiledFunc{
		Entry:     uintptr(unsafe.Pointer(&mem[0])),
		Page:      page,
		Size:      len(code),
		Hash:      hashBytecode(fn.Instructions),
		NumRegs:   fn.NumLocals,
		NumParams: fn.NumParameters,
	}

	// Cache the compiled function
	j.funcCache.Store(cf.Hash, cf)
	j.stats.CompiledFunctions++
	j.stats.TotalCodeSize += int64(len(code))

	if j.config.Debug {
		fmt.Printf("[JIT-ARM64] Compiled function: hash=%016x, size=%d bytes\n", cf.Hash, cf.Size)
	}

	return cf, nil
}

// Execute runs a compiled function
func (cf *CompiledFunc) Execute() int64 {
	// Call the compiled code
	// The entry point is a function that takes no args and returns int64
	fn := *(*func() int64)(unsafe.Pointer(&cf.Entry))
	return fn()
}

// Cleanup releases all JIT resources
func (j *JITCompiler) Cleanup() {
	j.pageLock.Lock()
	defer j.pageLock.Unlock()

	for _, page := range j.codePages {
		syscall.Munmap(page.Data)
	}
	j.codePages = nil
	j.funcCache = sync.Map{}
	j.hotPaths = sync.Map{}
}

// GetStats returns current JIT statistics
func (j *JITCompiler) GetStats() JITStats {
	return j.stats
}

// GetFunctionPointer returns a callable pointer for the compiled function
func (cf *CompiledFunc) GetFunctionPointer() uintptr {
	return cf.Entry
}

// ============================================================================
// ARM64 Code Generator
// ============================================================================

// ARM64CodeGenerator generates ARM64 machine code
type ARM64CodeGenerator struct {
	code      []byte
	labels    map[string]int
	fixups    []arm64Fixup
	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction
}

// arm64Fixup represents a location that needs patching
type arm64Fixup struct {
	offset int
	label  string
	size   int
}

// NewARM64CodeGenerator creates a new ARM64 code generator
func NewARM64CodeGenerator() *ARM64CodeGenerator {
	return &ARM64CodeGenerator{
		code:   make([]byte, 0, 4096),
		labels: make(map[string]int),
		fixups: make([]arm64Fixup, 0),
	}
}

// Generate generates ARM64 code from VM bytecode
func (cg *ARM64CodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
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
func (cg *ARM64CodeGenerator) identifyLabels(code []byte) {
	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegJump:
			offset := int(int16(uint16(code[ip+1])<<8 | uint16(code[ip+2])))
			target := ip + 3 + offset
			cg.labels[fmt.Sprintf("L%d", target)] = -1

		case compiler.OpRegJumpIfFalse, compiler.OpRegJumpIfTrue:
			offset := int(int16(uint16(code[ip+2])<<8 | uint16(code[ip+3])))
			target := ip + 4 + offset
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

// emitPrologue generates function entry code for ARM64
func (cg *ARM64CodeGenerator) emitPrologue() {
	// ARM64 calling convention:
	// x0-x7: arguments
	// x8: indirect result location
	// x9-x15: temporary registers
	// x19-x28: callee-saved registers
	// x29: frame pointer (FP)
	// x30: link register (LR)
	// SP: stack pointer

	// Save frame pointer and link register
	// stp x29, x30, [sp, #-16]!
	cg.emitUint32(0xA9BF7BFD)

	// Set up frame pointer
	// mov x29, sp
	cg.emitUint32(0x910003FD)

	// Save callee-saved registers (x19-x28)
	// stp x19, x20, [sp, #-16]!
	cg.emitUint32(0xA9BF6FF3)
	// stp x21, x22, [sp, #-16]!
	cg.emitUint32(0xA9BF67F5)
	// stp x23, x24, [sp, #-16]!
	cg.emitUint32(0xA9BF5FF7)
	// stp x25, x26, [sp, #-16]!
	cg.emitUint32(0xA9BF57F9)
	// stp x27, x28, [sp, #-16]!
	cg.emitUint32(0xA9BF4FFB)

	// Allocate stack space for local variables (512 bytes)
	// sub sp, sp, #512
	cg.emitUint32(0xD10083FF)
}

// emitEpilogue generates function exit code for ARM64
func (cg *ARM64CodeGenerator) emitEpilogue() {
	// Deallocate stack space
	// add sp, sp, #512
	cg.emitUint32(0x910083FF)

	// Restore callee-saved registers
	// ldp x27, x28, [sp], #16
	cg.emitUint32(0xA8C14FFB)
	// ldp x25, x26, [sp], #16
	cg.emitUint32(0xA8C157F9)
	// ldp x23, x24, [sp], #16
	cg.emitUint32(0xA8C15FF7)
	// ldp x21, x22, [sp], #16
	cg.emitUint32(0xA8C167F5)
	// ldp x19, x20, [sp], #16
	cg.emitUint32(0xA8C16FF3)

	// Restore frame pointer and return
	// ldp x29, x30, [sp], #16
	cg.emitUint32(0xA8C17BFD)
	// ret
	cg.emitUint32(0xD65F03C0)
}

// compileInstruction compiles a single instruction
func (cg *ARM64CodeGenerator) compileInstruction(op compiler.Opcode, code []byte, ip *int) error {
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
		cg.compileLess(dst, left, right)
		*ip += 4

	case compiler.OpRegGreater:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileGreater(dst, left, right)
		*ip += 4

	case compiler.OpRegEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegNotEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileNotEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegLessEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileLessEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegGreaterEqual:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileGreaterEqual(dst, left, right)
		*ip += 4

	case compiler.OpRegJump:
		offset := int(int16(uint16(code[*ip+1])<<8 | uint16(code[*ip+2])))
		target := *ip + 3 + offset
		cg.compileJump(target)
		*ip += 3

	case compiler.OpRegJumpIfFalse:
		cond := int(code[*ip+1])
		offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
		target := *ip + 4 + offset
		cg.compileJumpIfFalse(cond, target)
		*ip += 4

	case compiler.OpRegJumpIfTrue:
		cond := int(code[*ip+1])
		offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
		target := *ip + 4 + offset
		cg.compileJumpIfTrue(cond, target)
		*ip += 4

	case compiler.OpRegReturn:
		src := int(code[*ip+1])
		cg.compileReturn(src)
		*ip += 2

	case compiler.OpRegNull:
		dst := int(code[*ip+1])
		cg.compileNull(dst)
		*ip += 2

	case compiler.OpRegTrue:
		dst := int(code[*ip+1])
		cg.compileTrue(dst)
		*ip += 2

	case compiler.OpRegFalse:
		dst := int(code[*ip+1])
		cg.compileFalse(dst)
		*ip += 2

	case compiler.OpRegIncLocal:
		local := int(code[*ip+1])
		cg.compileIncLocal(local)
		*ip += 2

	case compiler.OpRegDecLocal:
		local := int(code[*ip+1])
		cg.compileDecLocal(local)
		*ip += 2

	case compiler.OpRegNot:
		dst := int(code[*ip+1])
		src := int(code[*ip+2])
		cg.compileNot(dst, src)
		*ip += 3

	case compiler.OpRegLoadGlobal:
		dst := int(code[*ip+1])
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileLoadGlobal(dst, globalIdx)
		*ip += 4

	case compiler.OpRegStoreGlobal:
		src := int(code[*ip+1])
		globalIdx := int(code[*ip+2])<<8 | int(code[*ip+3])
		cg.compileStoreGlobal(src, globalIdx)
		*ip += 4

	case compiler.OpRegAnd:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileAnd(dst, left, right)
		*ip += 4

	case compiler.OpRegOr:
		dst := int(code[*ip+1])
		left := int(code[*ip+2])
		right := int(code[*ip+3])
		cg.compileOr(dst, left, right)
		*ip += 4

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++

	default:
		// For unsupported opcodes, skip and continue
		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		*ip += width
	}

	return nil
}

// ============================================================================
// ARM64 Instruction Implementations
// ============================================================================

// Stack offset for local variables (after saved registers)
const arm64LocalOffset = 96

func (cg *ARM64CodeGenerator) localOffset(slot int) int32 {
	return int32(arm64LocalOffset + slot*8)
}

func (cg *ARM64CodeGenerator) compileLoadConst(dst, constIdx int) {
	var val int64
	if constIdx < len(cg.constants) {
		if cg.constants[constIdx].IsInt() {
			val = cg.constants[constIdx].GetInt()
		}
	}

	// Load immediate into x0, then store to slot
	// movz x0, #imm16, lsl #0
	// movk x0, #imm16, lsl #16
	// movk x0, #imm16, lsl #32
	// movk x0, #imm16, lsl #48
	cg.emitMovImm64(0, val)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileMove(dst, src int) {
	cg.loadFromSlot(0, src)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileAdd(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// add x0, x0, x1
	cg.emitUint32(0x8B010000)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileSub(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// sub x0, x0, x1
	cg.emitUint32(0xCB010000)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileMul(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// mul x0, x0, x1
	cg.emitUint32(0x9B017C00)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileDiv(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// sdiv x0, x0, x1
	cg.emitUint32(0x9AC10C00)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileMod(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// sdiv x2, x0, x1
	cg.emitUint32(0x9AC10C02)
	// msub x0, x2, x1, x0
	cg.emitUint32(0x9B018040)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileNeg(dst, src int) {
	cg.loadFromSlot(0, src)
	// neg x0, x0
	cg.emitUint32(0xCB0003E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileLess(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, lt
	cg.emitUint32(0x9A9F27E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileGreater(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, gt
	cg.emitUint32(0x9A9F37E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileEqual(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, eq
	cg.emitUint32(0x9A9F07E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileNotEqual(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, ne
	cg.emitUint32(0x9A9F17E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileLessEqual(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, le
	cg.emitUint32(0x9A9FD7E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileGreaterEqual(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// cmp x0, x1
	cg.emitUint32(0xEB01001F)
	// cset x0, ge
	cg.emitUint32(0x9A9FA7E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileJump(target int) {
	label := fmt.Sprintf("L%d", target)
	// b label (will be patched)
	cg.emitUint32(0x14000000)
	cg.fixups = append(cg.fixups, arm64Fixup{
		offset: len(cg.code) - 4,
		label:  label,
		size:   4,
	})
}

func (cg *ARM64CodeGenerator) compileJumpIfFalse(cond, target int) {
	cg.loadFromSlot(0, cond)
	// cbz x0, label
	cg.emitUint32(0xB4000000)
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, arm64Fixup{
		offset: len(cg.code) - 4,
		label:  label,
		size:   4,
	})
}

func (cg *ARM64CodeGenerator) compileJumpIfTrue(cond, target int) {
	cg.loadFromSlot(0, cond)
	// cbnz x0, label
	cg.emitUint32(0xB5000000)
	label := fmt.Sprintf("L%d", target)
	cg.fixups = append(cg.fixups, arm64Fixup{
		offset: len(cg.code) - 4,
		label:  label,
		size:   4,
	})
}

func (cg *ARM64CodeGenerator) compileReturn(src int) {
	cg.loadFromSlot(0, src)
	cg.emitEpilogue()
}

func (cg *ARM64CodeGenerator) compileNull(dst int) {
	// mov x0, #0
	cg.emitUint32(0xD2800000)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileTrue(dst int) {
	// mov x0, #1
	cg.emitUint32(0xD2800020)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileFalse(dst int) {
	// mov x0, #0
	cg.emitUint32(0xD2800000)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileIncLocal(local int) {
	cg.loadFromSlot(0, local)
	// add x0, x0, #1
	cg.emitUint32(0x91000400)
	cg.storeToSlot(0, local)
}

func (cg *ARM64CodeGenerator) compileDecLocal(local int) {
	cg.loadFromSlot(0, local)
	// sub x0, x0, #1
	cg.emitUint32(0xD1000400)
	cg.storeToSlot(0, local)
}

func (cg *ARM64CodeGenerator) compileNot(dst, src int) {
	cg.loadFromSlot(0, src)
	// cmp x0, #0
	cg.emitUint32(0xF100001F)
	// cset x0, eq
	cg.emitUint32(0x9A9F07E0)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileLoadGlobal(dst, globalIdx int) {
	// Load from globals[x0 + globalIdx*8]
	// x0 contains globals pointer (first argument)
	// ldr x1, [x0, #offset]
	offset := globalIdx * 8
	if offset < 32760 {
		cg.emitUint32(uint32(0xF9400011 | (offset/8)<<10))
	} else {
		// Large offset - use add then ldr
		cg.emitMovImm64(1, int64(offset))
		// add x1, x0, x1
		cg.emitUint32(0x8B010001)
		// ldr x1, [x1]
		cg.emitUint32(0xF9400231)
	}
	cg.storeToSlot(1, dst)
}

func (cg *ARM64CodeGenerator) compileStoreGlobal(src, globalIdx int) {
	cg.loadFromSlot(1, src)
	// Store to globals[x0 + globalIdx*8]
	offset := globalIdx * 8
	if offset < 32760 {
		cg.emitUint32(uint32(0xF9000011 | (offset/8)<<10))
	} else {
		// Large offset
		cg.emitMovImm64(2, int64(offset))
		// add x2, x0, x2
		cg.emitUint32(0x8B020002)
		// str x1, [x2]
		cg.emitUint32(0xF9000051)
	}
}

func (cg *ARM64CodeGenerator) compileAnd(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// and x0, x0, x1
	cg.emitUint32(0x8A010000)
	cg.storeToSlot(0, dst)
}

func (cg *ARM64CodeGenerator) compileOr(dst, left, right int) {
	cg.loadFromSlot(0, left)
	cg.loadFromSlot(1, right)
	// orr x0, x0, x1
	cg.emitUint32(0xAA010000)
	cg.storeToSlot(0, dst)
}

// ============================================================================
// ARM64 Helper Functions
// ============================================================================

func (cg *ARM64CodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *ARM64CodeGenerator) emitMovImm64(reg int, val int64) {
	// MOV (bitmask immediate) - may not work for all values
	// For simplicity, use MOVZ/MOVK sequence
	uval := uint64(val)
	// movz xN, #imm16, lsl #0
	cg.emitUint32(0xD2800000 | uint32(reg&31) | uint32((uval&0xFFFF)<<5))

	// movk xN, #imm16, lsl #16
	cg.emitUint32(0xF2A00000 | uint32(reg&31) | uint32(((uval>>16)&0xFFFF)<<5))

	// movk xN, #imm16, lsl #32
	cg.emitUint32(0xF2C00000 | uint32(reg&31) | uint32(((uval>>32)&0xFFFF)<<5))

	// movk xN, #imm16, lsl #48
	cg.emitUint32(0xF2E00000 | uint32(reg&31) | uint32(((uval>>48)&0xFFFF)<<5))
}

func (cg *ARM64CodeGenerator) loadFromSlot(reg, slot int) {
	offset := cg.localOffset(slot)
	// ldr xN, [x29, #offset]
	if offset >= 0 && offset < 32760 {
		cg.emitUint32(0xF94003E0 | uint32(reg&31) | uint32(offset/8)<<10)
	} else {
		cg.emitMovImm64(2, int64(offset))
		// add x2, x29, x2
		cg.emitUint32(0x8B0203E2)
		// ldr xN, [x2]
		cg.emitUint32(0xF9400040 | uint32(reg&31))
	}
}

func (cg *ARM64CodeGenerator) storeToSlot(reg, slot int) {
	offset := cg.localOffset(slot)
	// str xN, [x29, #offset]
	if offset >= 0 && offset < 32760 {
		cg.emitUint32(0xF90003E0 | uint32(reg&31) | uint32(offset/8)<<10)
	} else {
		cg.emitMovImm64(2, int64(offset))
		// add x2, x29, x2
		cg.emitUint32(0x8B0203E2)
		// str xN, [x2]
		cg.emitUint32(0xF9000040 | uint32(reg&31))
	}
}

func (cg *ARM64CodeGenerator) resolveFixups() error {
	for _, f := range cg.fixups {
		target, ok := cg.labels[f.label]
		if !ok {
			return fmt.Errorf("undefined label: %s", f.label)
		}

		// Calculate relative offset (in instructions, not bytes)
		// ARM64 branch offset is counted in 4-byte instructions
		offset := (target - (f.offset + 4)) / 4

		// Patch the instruction
		// For unconditional branch (b): offset is bits 0-25
		// For conditional branch (cbz/cbnz): offset is bits 5-23
		insn := binary.LittleEndian.Uint32(cg.code[f.offset:])
		if (insn & 0xFC000000) == 0x14000000 {
			// Unconditional branch
			insn = (insn & 0xFC000000) | (uint32(offset) & 0x03FFFFFF)
		} else {
			// Conditional branch (cbz/cbnz)
			insn = (insn & 0xFF00001F) | ((uint32(offset) & 0x0007FFFF) << 5)
		}
		binary.LittleEndian.PutUint32(cg.code[f.offset:], insn)
	}
	return nil
}

// ============================================================================
// Native Executor for ARM64
// ============================================================================

// NativeExecutor executes JIT-compiled native code on ARM64
type NativeExecutor struct {
	compiler *JITCompiler
	config   JITConfig
}

// NewNativeExecutor creates a new native JIT executor
func NewNativeExecutor(config JITConfig) *NativeExecutor {
	return &NativeExecutor{
		compiler: NewJITCompiler(config),
		config:   config,
	}
}

// CanExecuteNatively checks if a function can be executed natively
func CanExecuteNatively(fn *compiler.CompiledFunction) bool {
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegLoadConst, compiler.OpRegMove,
			compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
			compiler.OpRegNeg, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegEqual,
			compiler.OpRegNotEqual, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegReturn, compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegIncLocal, compiler.OpRegDecLocal,
			compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal:
			// Supported

		default:
			return false
		}

		// Skip operands
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return false
		}
		operandWidth := 0
		for _, w := range def.OperandWidths {
			operandWidth += w
		}
		ip += operandWidth + 1
	}

	return true
}

// ExecuteFunction compiles and executes a function natively
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	if !CanExecuteNatively(fn) {
		return 0, fmt.Errorf("function cannot be executed natively")
	}

	cf, err := ne.compiler.Compile(fn, constants, nil)
	if err != nil {
		return 0, err
	}

	return cf.Execute(), nil
}

// Cleanup releases JIT resources
func (ne *NativeExecutor) Cleanup() {
	ne.compiler.Cleanup()
}

// ============================================================================
// Additional Types and Functions
// ============================================================================

// NativeFunction represents a natively compiled function
type NativeFunction struct {
	Code      []byte
	NumParams int
}

// Execute returns 0 for stub
func (nf *NativeFunction) Execute(globals []int64, args ...int64) int64 {
	return 0
}

// NativeFunctionRegistry manages compiled native functions
type NativeFunctionRegistry struct {
	compiler *JITCompiler
	config   JITConfig
}

// NewNativeFunctionRegistry creates a new registry
func NewNativeFunctionRegistry(config JITConfig) *NativeFunctionRegistry {
	return &NativeFunctionRegistry{
		compiler: NewJITCompiler(config),
		config:   config,
	}
}

// Get returns nil
func (r *NativeFunctionRegistry) Get(idx int) *NativeFunction {
	return nil
}

// CompileFunction is a stub
func (r *NativeFunctionRegistry) CompileFunction(fn *compiler.CompiledFunction, idx int, constants []int64) error {
	return nil
}

// CompileRecursiveFunction is a stub
func (r *NativeFunctionRegistry) CompileRecursiveFunction(fn *compiler.CompiledFunction, idx int, constants []vm.Value) error {
	return nil
}

// Cleanup releases resources
func (r *NativeFunctionRegistry) Cleanup() {
	r.compiler.Cleanup()
}

// AnalyzeNativeSupport returns the support level
func AnalyzeNativeSupport(fn *compiler.CompiledFunction) int {
	if CanExecuteNatively(fn) {
		return 1 // SupportPureArithmetic
	}
	return 0 // SupportNone
}

// Support levels
const (
	SupportNone           = 0
	SupportPureArithmetic = 1
	SupportWithBuiltins   = 2
	SupportWithCalls      = 3
	SupportWithArrays     = 4
	SupportWithObjects    = 5
)

// CallNativeCode stub
func CallNativeCode(codePtr uintptr, arg int64) int64 {
	return 0
}

// SimpleNativeCall stub
func SimpleNativeCall(code []byte, arg int64) int64 {
	return 0
}

// GetCollectionCallbackPtr returns 0 for ARM64 stub
func GetCollectionCallbackPtr() uintptr {
	return 0
}

// GetJITCollectionCallbackPtr returns 0 for ARM64 stub
func GetJITCollectionCallbackPtr() uintptr {
	return 0
}

// GetGlobalJITCollectionContext returns nil for ARM64 stub
func GetGlobalJITCollectionContext() *JITCollectionContext {
	return nil
}

// ResetGlobalJITCollectionContext is a no-op for ARM64
func ResetGlobalJITCollectionContext() {}

// RequiresInterpreterFallback checks if a function requires interpreter
func RequiresInterpreterFallback(fn *compiler.CompiledFunction) bool {
	return !CanExecuteNatively(fn)
}

// JITSupportLevel represents the level of JIT support
type JITSupportLevel int

const (
	JITSupportNone JITSupportLevel = iota
	JITSupportPartial
	JITSupportFull
)

// GetJITSupportLevel returns the support level for a function
func GetJITSupportLevel(fn *compiler.CompiledFunction) JITSupportLevel {
	if CanExecuteNatively(fn) {
		return JITSupportFull
	}
	return JITSupportNone
}

// JITCollectionContext is a stub for ARM64
type JITCollectionContext struct{}

// NewJITCollectionContext creates a new context
func NewJITCollectionContext() *JITCollectionContext {
	return &JITCollectionContext{}
}
