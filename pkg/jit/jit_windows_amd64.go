//go:build windows && amd64
// +build windows,amd64

// pkg/jit/jit_windows_amd64.go
// JIT Compiler for Xxlang VM on Windows
// Uses VirtualAlloc for executable memory allocation
package jit

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITCompiler handles JIT compilation of VM bytecode
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
	// Entry point to the compiled code
	Entry uintptr
	// Code page containing the compiled code
	Page *CodePage
	// Size of compiled code
	Size int
	// Bytecode hash for cache invalidation
	Hash uint64
	// Number of locals/registers needed
	NumRegs int
	// Number of parameters
	NumParams int
}

// CodePage represents a page of executable memory
type CodePage struct {
	Data []byte
	Used int
}

// NewJITCompiler creates a new JIT compiler
func NewJITCompiler(config JITConfig) *JITCompiler {
	return &JITCompiler{
		config: config,
	}
}

// allocCodePageLocked allocates a new executable code page
// Must be called with pageLock held
func (j *JITCompiler) allocCodePageLocked() (*CodePage, error) {
	size := 64 * 1024 // 64KB pages

	// Use Windows VirtualAlloc for executable memory
	data, err := bridge.AllocExecMem(size)
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

// allocCode allocates code in an executable page
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

// Compile compiles a function to native code
func (j *JITCompiler) Compile(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) (*CompiledFunc, error) {
	// Check cache first
	if cf := j.GetCompiled(fn); cf != nil {
		return cf, nil
	}

	var code []byte
	var err error

	// Try Fibonacci JIT compiler first (handles recursive functions)
	fibCompiler := NewFibJITCompiler(j.config)
	code, err = fibCompiler.Compile(fn, constants, globals)
	if err == nil {
		if j.config.Debug {
			fmt.Printf("[JIT] Fibonacci compiler succeeded: %d bytes\n", len(code))
		}
		// Continue to allocate and cache
	} else {
		if j.config.Debug {
			fmt.Printf("[JIT] Fibonacci compiler failed: %v, trying simple generator\n", err)
		}

		// Try simple code generator (most reliable for non-recursive functions)
		scg := NewSimpleCodeGenerator()
		code, err = scg.Generate(fn, constants, globals)
		if err != nil {
			if j.config.Debug {
				fmt.Printf("[JIT] Simple generator failed: %v\n", err)
			}
			return nil, fmt.Errorf("code generation failed: %w", err)
		}

		if j.config.Debug {
			fmt.Printf("[JIT] Simple generator succeeded: %d bytes\n", len(code))
		}
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
		fmt.Printf("[JIT] Compiled function: hash=%016x, size=%d bytes\n", cf.Hash, cf.Size)
	}

	return cf, nil
}

// Execute runs a compiled function
func (cf *CompiledFunc) Execute() int64 {
	// Call the compiled code using Windows x64 ABI
	// The entry point is a function that takes no args and returns int64
	fn := *(*func() int64)(unsafe.Pointer(&cf.Entry))
	return fn()
}

// Cleanup releases all JIT resources
func (j *JITCompiler) Cleanup() {
	j.pageLock.Lock()
	defer j.pageLock.Unlock()

	for _, page := range j.codePages {
		bridge.FreeExecMem(page.Data)
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
// JIT Support Analysis Functions
// ============================================================================

// JITSupportLevel represents the level of JIT support for a function
type JITSupportLevel int

const (
	JITSupportNone    JITSupportLevel = iota // Cannot be JIT compiled
	JITSupportPartial                        // Can be JIT compiled with limitations
	JITSupportFull                           // Full JIT support
)

// RequiresInterpreterFallback checks if a function requires interpreter fallback
func RequiresInterpreterFallback(fn *compiler.CompiledFunction) bool {
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		// Check for operations that require interpreter
		switch op {
		case compiler.OpRegClosure,
			compiler.OpRegLoadFree,
			compiler.OpRegStoreFree,
			compiler.OpRegBuiltin,
			compiler.OpRegGetMethod,
			compiler.OpRegCallMethod,
			compiler.OpRegClass,
			compiler.OpRegNew,
			compiler.OpRegThrow,
			compiler.OpRegPushHandler,
			compiler.OpRegPopHandler,
			compiler.OpRegLoadModule,
			compiler.OpRegGetExport,
			compiler.OpRegSetExport,
			compiler.OpRegIterKey,
			compiler.OpRegIterValue:
			return true

		// Concurrency operations
		case compiler.OpRegRunStart,
			compiler.OpRegRunWait,
			compiler.OpRegMakeTube,
			compiler.OpRegTubeSend,
			compiler.OpRegTubeRecv,
			compiler.OpRegTubeClose,
			compiler.OpRegSelectStart,
			compiler.OpRegSelectCase,
			compiler.OpRegSelectEnd,
			compiler.OpRegMutexLock,
			compiler.OpRegMutexUnlock,
			compiler.OpRegWGAdd,
			compiler.OpRegWGWait,
			compiler.OpRegWGDone,
			compiler.OpRegAtomicAdd,
			compiler.OpRegAtomicLoad,
			compiler.OpRegAtomicSwap:
			return true
		}

		// Skip operands
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return true // Unknown opcode - fallback to interpreter
		}
		ip++
		for _, w := range def.OperandWidths {
			ip += w
		}
	}

	return false
}

// GetJITSupportLevel analyzes a function and returns its JIT support level
func GetJITSupportLevel(fn *compiler.CompiledFunction) JITSupportLevel {
	if RequiresInterpreterFallback(fn) {
		return JITSupportNone
	}

	// Check for operations that limit JIT support
	code := fn.Instructions
	ip := 0
	hasArrayOps := false
	hasMapOps := false
	hasFieldOps := false

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegArray, compiler.OpRegArrayEmpty, compiler.OpRegArrayAppend,
			compiler.OpRegIndex, compiler.OpRegSetIndex:
			hasArrayOps = true

		case compiler.OpRegMap, compiler.OpRegMapEmpty, compiler.OpRegMapSet:
			hasMapOps = true

		case compiler.OpRegGetField, compiler.OpRegSetField:
			hasFieldOps = true
		}

		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return JITSupportNone
		}
		ip++
		for _, w := range def.OperandWidths {
			ip += w
		}
	}

	// If there are array/map/field operations, JIT support is partial
	if hasArrayOps || hasMapOps || hasFieldOps {
		return JITSupportPartial
	}

	return JITSupportFull
}
