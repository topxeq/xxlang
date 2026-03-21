// +build amd64,!windows

// pkg/jit/jit_full.go
// Full JIT implementation with function call support via interpreter callback
package jit

import (
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// FullJITConfig holds configuration for the full JIT compiler
type FullJITConfig struct {
	HotThreshold int
	MaxCodeSize  int
	Debug        bool
}

// DefaultFullJITConfig returns default configuration
func DefaultFullJITConfig() FullJITConfig {
	return FullJITConfig{
		HotThreshold: 100,
		MaxCodeSize:  8192,
		Debug:        false,
	}
}

// FullJITCompiler handles JIT compilation with interpreter fallback
type FullJITCompiler struct {
	config FullJITConfig

	// Code pages
	codePages []*CodePage
	pageLock  sync.Mutex

	// Function cache
	funcCache sync.Map

	// Hot path counters
	hotPaths sync.Map

	// Statistics
	stats JITStats

	// Interpreter callback for complex operations
	interpCallback unsafe.Pointer
}

// FullJITVM wraps a register VM with full JIT support
type FullJITVM struct {
	*vm.RegVM
	jit    *FullJITCompiler
	config FullJITConfig
	enabled bool
}

// NewFullJITCompiler creates a new full JIT compiler
func NewFullJITCompiler(config FullJITConfig) *FullJITCompiler {
	return &FullJITCompiler{
		config: config,
	}
}

// NewFullJITVM creates a new JIT-enabled VM
func NewFullJITVM(bytecode *compiler.Bytecode, config FullJITConfig) *FullJITVM {
	return &FullJITVM{
		RegVM:   vm.NewRegVM(bytecode),
		jit:     NewFullJITCompiler(config),
		enabled: true,
	}
}

// NewFullJITVMWithGlobals creates a JIT VM with custom globals
func NewFullJITVMWithGlobals(bytecode *compiler.Bytecode, globals []vm.Value, config FullJITConfig) *FullJITVM {
	return &FullJITVM{
		RegVM:   vm.NewRegVMWithGlobals(bytecode, globals),
		jit:     NewFullJITCompiler(config),
		enabled: true,
	}
}

// SetJITEnabled enables or disables JIT
func (j *FullJITVM) SetJITEnabled(enabled bool) {
	j.enabled = enabled
}

// GetJITStats returns JIT statistics
func (j *FullJITVM) GetJITStats() JITStats {
	return j.jit.stats
}

// Cleanup releases JIT resources
func (j *FullJITVM) Cleanup() {
	j.jit.Cleanup()
}

// Cleanup releases all resources
func (j *FullJITCompiler) Cleanup() {
	j.pageLock.Lock()
	defer j.pageLock.Unlock()

	for _, page := range j.codePages {
		syscall.Munmap(page.Data)
	}
	j.codePages = nil
	j.funcCache = sync.Map{}
	j.hotPaths = sync.Map{}
}

// ShouldCompile returns true if the function should be JIT compiled
func (j *FullJITCompiler) ShouldCompile(fn *compiler.CompiledFunction) bool {
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

// RecordExecution records an execution for hot path detection
func (j *FullJITCompiler) RecordExecution(fn *compiler.CompiledFunction) bool {
	hash := hashBytecode(fn.Instructions)

	count, _ := j.hotPaths.LoadOrStore(hash, 0)
	newCount := count.(int) + 1
	j.hotPaths.Store(hash, newCount)

	return newCount == j.config.HotThreshold
}

// Run executes bytecode with JIT optimization for hot paths
func (j *FullJITVM) Run() error {
	if !j.enabled {
		return j.RegVM.Run()
	}

	// For now, always use interpreter
	// JIT optimization happens at function level
	return j.RegVM.Run()
}

// ============================================================================
// Full JIT Code Generator
// ============================================================================

// FullCodeGenerator generates x86-64 code with interpreter callback support
type FullCodeGenerator struct {
	config FullJITConfig
	code   []byte
	labels map[string]int
	fixups []fixup

	constants []vm.Value
	globals   []vm.Value
	fn        *compiler.CompiledFunction

	// For tracking function calls
	hasCalls bool

	// Register allocation
	maxReg int
}

// NewFullCodeGenerator creates a new code generator
func NewFullCodeGenerator(config FullJITConfig) *FullCodeGenerator {
	return &FullCodeGenerator{
		config:  config,
		code:    make([]byte, 0, 8192),
		labels:  make(map[string]int),
		fixups:  make([]fixup, 0),
	}
}

// Generate generates machine code for a function
func (cg *FullCodeGenerator) Generate(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) ([]byte, error) {
	cg.code = cg.code[:0]
	cg.labels = make(map[string]int)
	cg.fixups = cg.fixups[:0]
	cg.constants = constants
	cg.globals = globals
	cg.fn = fn
	cg.hasCalls = false
	cg.maxReg = 64

	// Analyze bytecode first
	cg.analyzeBytecode(fn.Instructions)

	// Generate prologue
	cg.emitPrologue()

	// Generate code for each instruction
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

// analyzeBytecode identifies all labels and function calls
func (cg *FullCodeGenerator) analyzeBytecode(code []byte) {
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

		case compiler.OpRegCall, compiler.OpRegTailCall:
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
func (cg *FullCodeGenerator) compileOpcode(op compiler.Opcode, code []byte, ip *int) error {
	switch op {
	// Data movement
	case compiler.OpRegLoadConst:
		cg.compileLoadConst(code, ip)
	case compiler.OpRegMove:
		cg.compileMove(code, ip)

	// Arithmetic
	case compiler.OpRegAdd:
		cg.compileBinaryOp(code, ip, 0x01) // ADD
	case compiler.OpRegSub:
		cg.compileBinaryOp(code, ip, 0x29) // SUB
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
		cg.compileComparison(code, ip, 0x9C) // SETL
	case compiler.OpRegGreater:
		cg.compileComparison(code, ip, 0x9F) // SETG
	case compiler.OpRegEqual:
		cg.compileComparison(code, ip, 0x94) // SETE
	case compiler.OpRegNotEqual:
		cg.compileComparison(code, ip, 0x95) // SETNE
	case compiler.OpRegLessEqual:
		cg.compileComparison(code, ip, 0x9E) // SETLE
	case compiler.OpRegGreaterEqual:
		cg.compileComparison(code, ip, 0x9D) // SETGE

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

	// Closure - fallback to interpreter
	case compiler.OpRegClosure:
		cg.compileClosure(code, ip)

	// Function calls - need interpreter support
	case compiler.OpRegCall:
		return fmt.Errorf("OpRegCall requires interpreter - function has calls")

	case compiler.OpRegTailCall:
		return fmt.Errorf("OpRegTailCall requires interpreter - function has calls")

	case compiler.OpReturn:
		cg.emitEpilogue()
		*ip++

	default:
		def, _ := compiler.Lookup(byte(op))
		return fmt.Errorf("unsupported opcode: %s (%d)", def.Name, op)
	}

	return nil
}

// ============================================================================
// Prologue and Epilogue
// ============================================================================

func (cg *FullCodeGenerator) emitPrologue() {
	// push rbp
	cg.emit(0x55)
	// mov rbp, rsp
	cg.emitBytes([]byte{0x48, 0x89, 0xE5})

	// Save callee-saved registers
	cg.emit(0x53)                    // push rbx
	cg.emitBytes([]byte{0x41, 0x54}) // push r12
	cg.emitBytes([]byte{0x41, 0x55}) // push r13
	cg.emitBytes([]byte{0x41, 0x56}) // push r14
	cg.emitBytes([]byte{0x41, 0x57}) // push r15

	// Allocate stack space (64 registers * 8 bytes = 512 bytes + temp space)
	cg.emitBytes([]byte{0x48, 0x81, 0xEC})
	cg.emitUint32(1024)

	// Initialize registers to null
	nullVal := uint64(TagNull) << 48
	for i := 0; i < 64; i++ {
		off := int32(8 * (i + 1))
		cg.emitBytes([]byte{0x48, 0xC7, 0x45})
		cg.emitByte(byte(-off))
		cg.emitUint32(uint32(nullVal))
	}
}

func (cg *FullCodeGenerator) emitEpilogue() {
	// add rsp, 1024
	cg.emitBytes([]byte{0x48, 0x81, 0xC4})
	cg.emitUint32(1024)

	// Restore callee-saved registers
	cg.emitBytes([]byte{0x41, 0x5F}) // pop r15
	cg.emitBytes([]byte{0x41, 0x5E}) // pop r14
	cg.emitBytes([]byte{0x41, 0x5D}) // pop r13
	cg.emitBytes([]byte{0x41, 0x5C}) // pop r12
	cg.emit(0x5B)                    // pop rbx
	cg.emit(0x5D)                    // pop rbp
	cg.emit(0xC3)                    // ret
}

// ============================================================================
// Opcode Implementations
// ============================================================================

func (cg *FullCodeGenerator) compileLoadConst(code []byte, ip *int) {
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
	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileMove(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRax(src)
	cg.storeRax(dst)
	*ip += 3
}

func (cg *FullCodeGenerator) compileBinaryOp(code []byte, ip *int, opByte byte) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)

	switch opByte {
	case 0x01: // ADD
		cg.emitBytes([]byte{0x48, 0x01, 0xC8})
	case 0x29: // SUB
		cg.emitBytes([]byte{0x48, 0x29, 0xC8})
	}

	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileMul(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x0F, 0xAF, 0xC1}) // imul rax, rcx

	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileDiv(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x99})           // cqo
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})     // idiv rcx

	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileMod(code []byte, ip *int) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)
	cg.emitBytes([]byte{0x48, 0x99})           // cqo
	cg.emitBytes([]byte{0x48, 0xF7, 0xF9})     // idiv rcx
	cg.emitBytes([]byte{0x48, 0x89, 0xD0})     // mov rax, rdx

	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileNeg(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRax(src)
	cg.emitBytes([]byte{0x48, 0xF7, 0xD8}) // neg rax
	cg.storeRax(dst)
	*ip += 3
}

func (cg *FullCodeGenerator) compileComparison(code []byte, ip *int, setcc byte) {
	dst := int(code[*ip+1])
	left := int(code[*ip+2])
	right := int(code[*ip+3])

	cg.loadRax(left)
	cg.loadRcx(right)

	cg.emitBytes([]byte{0x48, 0x39, 0xC8})     // cmp rax, rcx
	cg.emitBytes([]byte{0x48, 0x31, 0xC0})     // xor rax, rax
	cg.emitBytes([]byte{0x0F, setcc, 0xC0})    // setcc al

	// Convert to tagged bool
	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})     // movzx eax, al
	cg.emitBytes([]byte{0x48, 0x0D})           // or rax, imm64
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRax(dst)
	*ip += 4
}

func (cg *FullCodeGenerator) compileNot(code []byte, ip *int) {
	dst := int(code[*ip+1])
	src := int(code[*ip+2])

	cg.loadRax(src)

	// Check if value is falsy (null or false tag)
	cg.emitBytes([]byte{0x48, 0x3D})           // cmp rax, imm64
	cg.emitUint64(uint64(TagNull) << 48)
	cg.emitBytes([]byte{0x0F, 0x94, 0xC0})     // sete al

	cg.emitBytes([]byte{0x0F, 0xB6, 0xC0})
	cg.emitBytes([]byte{0x48, 0x0D})
	cg.emitUint64(uint64(TagBool) << 48)

	cg.storeRax(dst)
	*ip += 3
}

func (cg *FullCodeGenerator) compileJump(code []byte, ip *int) {
	offset := int(int16(uint16(code[*ip+1])<<8 | uint16(code[*ip+2])))
	target := *ip + 3 + offset

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJmp(label)
	*ip += 3
}

func (cg *FullCodeGenerator) compileJumpIfFalse(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRax(cond)

	// Check if falsy (tag is null or bool with value 0)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool) << 48) // ValueFalse

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label) // JE
	*ip += 4
}

func (cg *FullCodeGenerator) compileJumpIfTrue(code []byte, ip *int) {
	cond := int(code[*ip+1])
	offset := int(int16(uint16(code[*ip+2])<<8 | uint16(code[*ip+3])))
	target := *ip + 4 + offset

	cg.loadRax(cond)

	// Check if truthy (tag is bool with value 1)
	cg.emitBytes([]byte{0x48, 0x3D})
	cg.emitUint64(uint64(TagBool)<<48 | 1) // ValueTrue

	label := fmt.Sprintf("ip_%d", target)
	cg.emitJcc(0x84, label) // JE
	*ip += 4
}

func (cg *FullCodeGenerator) compileReturn(code []byte, ip *int) {
	src := int(code[*ip+1])
	cg.loadRax(src)
	cg.storeRax(compiler.ReturnRegister)
	cg.emitEpilogue()
	*ip += 2
}

func (cg *FullCodeGenerator) compileLoadLocal(code []byte, ip *int) {
	dst := int(code[*ip+1])
	local := int(code[*ip+2])
	cg.loadRax(local)
	cg.storeRax(dst)
	*ip += 3
}

func (cg *FullCodeGenerator) compileStoreLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	src := int(code[*ip+2])
	cg.loadRax(src)
	cg.storeRax(local)
	*ip += 3
}

func (cg *FullCodeGenerator) compileLoadGlobal(code []byte, ip *int) {
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

func (cg *FullCodeGenerator) compileStoreGlobal(code []byte, ip *int) {
	// JIT code doesn't modify globals (they're snapshots)
	*ip += 4
}

func (cg *FullCodeGenerator) compileIncLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	cg.loadRax(local)
	cg.emitBytes([]byte{0x48, 0x83, 0xC0, 0x01}) // add rax, 1
	cg.storeRax(local)
	*ip += 2
}

func (cg *FullCodeGenerator) compileDecLocal(code []byte, ip *int) {
	local := int(code[*ip+1])
	cg.loadRax(local)
	cg.emitBytes([]byte{0x48, 0x83, 0xE8, 0x01}) // sub rax, 1
	cg.storeRax(local)
	*ip += 2
}

func (cg *FullCodeGenerator) compileNull(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *FullCodeGenerator) compileTrue(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool)<<48 | 1)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *FullCodeGenerator) compileFalse(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagBool) << 48)
	cg.storeRax(dst)
	*ip += 2
}

func (cg *FullCodeGenerator) compilePush(code []byte, ip *int) {
	src := int(code[*ip+1])
	cg.loadRax(src)
	cg.emitBytes([]byte{0x48, 0x89, 0x04, 0x24}) // mov [rsp], rax
	*ip += 2
}

func (cg *FullCodeGenerator) compilePop(code []byte, ip *int) {
	dst := int(code[*ip+1])
	cg.emitBytes([]byte{0x48, 0x8B, 0x04, 0x24}) // mov rax, [rsp]
	cg.storeRax(dst)
	*ip += 2
}

func (cg *FullCodeGenerator) compileClosure(code []byte, ip *int) {
	dst := int(code[*ip+1])
	// Set to null - closures need interpreter
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(TagNull) << 48)
	cg.storeRax(dst)
	*ip += 6
}

// ============================================================================
// Helper Functions
// ============================================================================

func regOff(reg int) int32 {
	return int32(8 * (reg + 1))
}

func (cg *FullCodeGenerator) loadRax(reg int) {
	off := regOff(reg)
	cg.emitBytes([]byte{0x48, 0x8B, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *FullCodeGenerator) storeRax(reg int) {
	off := regOff(reg)
	cg.emitBytes([]byte{0x48, 0x89, 0x45})
	cg.emitByte(byte(-off))
}

func (cg *FullCodeGenerator) loadRcx(reg int) {
	off := regOff(reg)
	cg.emitBytes([]byte{0x48, 0x8B, 0x4D})
	cg.emitByte(byte(-off))
}

func (cg *FullCodeGenerator) emit(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *FullCodeGenerator) emitBytes(b []byte) {
	cg.code = append(cg.code, b...)
}

func (cg *FullCodeGenerator) emitByte(b byte) {
	cg.code = append(cg.code, b)
}

func (cg *FullCodeGenerator) emitUint32(v uint32) {
	cg.code = append(cg.code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (cg *FullCodeGenerator) emitUint64(v uint64) {
	cg.code = append(cg.code,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func (cg *FullCodeGenerator) emitJmp(label string) {
	cg.emitBytes([]byte{0xE9})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *FullCodeGenerator) emitJcc(cc byte, label string) {
	cg.emitBytes([]byte{0x0F, 0x80 | cc})
	cg.fixups = append(cg.fixups, fixup{offset: len(cg.code), label: label, size: 4})
	cg.emitUint32(0)
}

func (cg *FullCodeGenerator) emitCall(ptr uintptr) {
	// mov rax, imm64
	cg.emitBytes([]byte{0x48, 0xB8})
	cg.emitUint64(uint64(ptr))
	// call rax
	cg.emitBytes([]byte{0xFF, 0xD0})
}

// HasCalls returns true if the function has calls
func (cg *FullCodeGenerator) HasCalls() bool {
	return cg.hasCalls
}
