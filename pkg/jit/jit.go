// +build amd64,!windows

// pkg/jit/jit.go
// JIT Compiler for Xxlang VM - Pure Go Implementation
// Generates x86-64 machine code at runtime
package jit

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
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

	// Allocate executable memory
	// Use MAP_ANON for compatibility with both Linux and Darwin
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

// hashBytecode is defined in cache.go

// GetFunctionPointer returns a callable pointer for the compiled function
func (cf *CompiledFunc) GetFunctionPointer() uintptr {
	return cf.Entry
}
