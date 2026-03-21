// +build amd64,!windows

// pkg/jit/cache.go
// Function compilation cache for JIT
package jit

import (
	"hash/fnv"
	"sync"
	"sync/atomic"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// FunctionCache manages JIT-compiled function caching
type FunctionCache struct {
	// Compiled functions indexed by bytecode hash
	entries sync.Map // map[uint64]*CacheEntry

	// Execution counters for hot path detection
	execCounts sync.Map // map[uint64]*int64

	// Configuration
	config CacheConfig

	// Statistics
	stats CacheStats
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	// HotThreshold is the execution count before JIT compilation
	HotThreshold int64

	// MaxEntries limits the cache size
	MaxEntries int

	// MaxCodeSize limits bytecode size for compilation
	MaxCodeSize int
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		HotThreshold: 100,
		MaxEntries:   10000,
		MaxCodeSize:  4096,
	}
}

// CacheEntry represents a cached compiled function
type CacheEntry struct {
	// Bytecode hash
	Hash uint64

	// Compiled function
	Compiled *CompiledFunc

	// Original bytecode (for verification)
	Bytecode []byte

	// Function info
	NumParams int
	NumLocals int

	// Compilation state
	State CompilationState

	// Execution count at compilation time
	CompileCount int64
}

// CompilationState represents the state of compilation
type CompilationState int32

const (
	StateNotCompiled CompilationState = iota
	StateCompiling
	StateCompiled
	StateFailed
)

// CacheStats holds cache statistics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Compiles  int64
	Evictions int64
}

// NewFunctionCache creates a new function cache
func NewFunctionCache(config CacheConfig) *FunctionCache {
	return &FunctionCache{
		config: config,
	}
}

// RecordExecution records a function execution for hot path detection
// Returns true if the function should be JIT compiled
func (c *FunctionCache) RecordExecution(fn *compiler.CompiledFunction) bool {
	hash := hashBytecode(fn.Instructions)

	// Get or create counter
	counterIface, _ := c.execCounts.LoadOrStore(hash, new(int64))
	counter := counterIface.(*int64)

	// Increment and check threshold
	newCount := atomic.AddInt64(counter, 1)
	return newCount == c.config.HotThreshold
}

// ShouldCompile returns true if the function should be compiled
func (c *FunctionCache) ShouldCompile(fn *compiler.CompiledFunction) bool {
	// Check size limit
	if len(fn.Instructions) > c.config.MaxCodeSize {
		return false
	}

	// Check if already compiled
	hash := hashBytecode(fn.Instructions)
	if entry := c.Get(hash); entry != nil && entry.State == StateCompiled {
		return false
	}

	// Check execution count
	counterIface, ok := c.execCounts.Load(hash)
	if !ok {
		return false
	}

	counter := counterIface.(*int64)
	return atomic.LoadInt64(counter) >= c.config.HotThreshold
}

// Get retrieves a cached entry
func (c *FunctionCache) Get(hash uint64) *CacheEntry {
	if iface, ok := c.entries.Load(hash); ok {
		atomic.AddInt64(&c.stats.Hits, 1)
		return iface.(*CacheEntry)
	}
	atomic.AddInt64(&c.stats.Misses, 1)
	return nil
}

// GetByBytecode retrieves a cached entry by bytecode
func (c *FunctionCache) GetByBytecode(bytecode []byte) *CacheEntry {
	hash := hashBytecode(bytecode)
	return c.Get(hash)
}

// Put stores a compiled function in the cache
func (c *FunctionCache) Put(fn *compiler.CompiledFunction, compiled *CompiledFunc) *CacheEntry {
	hash := hashBytecode(fn.Instructions)

	entry := &CacheEntry{
		Hash:      hash,
		Compiled:  compiled,
		Bytecode:  fn.Instructions,
		NumParams: fn.NumParameters,
		NumLocals: fn.NumLocals,
		State:     StateCompiled,
	}

	c.entries.Store(hash, entry)
	atomic.AddInt64(&c.stats.Compiles, 1)

	return entry
}

// StartCompiling marks a function as being compiled
func (c *FunctionCache) StartCompiling(fn *compiler.CompiledFunction) *CacheEntry {
	hash := hashBytecode(fn.Instructions)

	// Check for existing entry
	if existing := c.Get(hash); existing != nil {
		return existing
	}

	// Create new entry in compiling state
	counterIface, _ := c.execCounts.Load(hash)
	var execCount int64
	if counterIface != nil {
		execCount = atomic.LoadInt64(counterIface.(*int64))
	}

	entry := &CacheEntry{
		Hash:         hash,
		Bytecode:     fn.Instructions,
		NumParams:    fn.NumParameters,
		NumLocals:    fn.NumLocals,
		State:        StateCompiling,
		CompileCount: execCount,
	}

	c.entries.Store(hash, entry)
	return entry
}

// MarkFailed marks a function compilation as failed
func (c *FunctionCache) MarkFailed(fn *compiler.CompiledFunction) {
	hash := hashBytecode(fn.Instructions)
	if iface, ok := c.entries.Load(hash); ok {
		entry := iface.(*CacheEntry)
		entry.State = StateFailed
	}
}

// GetStats returns cache statistics
func (c *FunctionCache) GetStats() CacheStats {
	return CacheStats{
		Hits:      atomic.LoadInt64(&c.stats.Hits),
		Misses:    atomic.LoadInt64(&c.stats.Misses),
		Compiles:  atomic.LoadInt64(&c.stats.Compiles),
		Evictions: atomic.LoadInt64(&c.stats.Evictions),
	}
}

// Clear clears the cache
func (c *FunctionCache) Clear() {
	c.entries = sync.Map{}
	c.execCounts = sync.Map{}
}

// GetCompiledFunction returns the compiled function if available
func (c *FunctionCache) GetCompiledFunction(fn *compiler.CompiledFunction) *CompiledFunc {
	hash := hashBytecode(fn.Instructions)
	if entry := c.Get(hash); entry != nil && entry.State == StateCompiled {
		return entry.Compiled
	}
	return nil
}

// IsCompiled returns true if the function is JIT compiled
func (c *FunctionCache) IsCompiled(fn *compiler.CompiledFunction) bool {
	hash := hashBytecode(fn.Instructions)
	if entry := c.Get(hash); entry != nil {
		return entry.State == StateCompiled
	}
	return false
}

// GetExecutionCount returns the execution count for a function
func (c *FunctionCache) GetExecutionCount(fn *compiler.CompiledFunction) int64 {
	hash := hashBytecode(fn.Instructions)
	if counterIface, ok := c.execCounts.Load(hash); ok {
		return atomic.LoadInt64(counterIface.(*int64))
	}
	return 0
}

// hashBytecode creates a hash of bytecode for cache lookup
func hashBytecode(bytecode []byte) uint64 {
	h := fnv.New64a()
	h.Write(bytecode)
	return h.Sum64()
}

// Global function cache
var globalCache = NewFunctionCache(DefaultCacheConfig())

// GetGlobalCache returns the global function cache
func GetGlobalCache() *FunctionCache {
	return globalCache
}

// WarmupFunction pre-warms the cache by recording executions
func WarmupFunction(fn *compiler.CompiledFunction, count int) {
	cache := GetGlobalCache()
	for i := 0; i < count; i++ {
		cache.RecordExecution(fn)
	}
}

// CompileAndCache compiles a function and caches the result
func CompileAndCache(jit *JITCompiler, fn *compiler.CompiledFunction, constants, globals []vm.Value) (*CompiledFunc, error) {
	cache := GetGlobalCache()

	// Check cache first
	if compiled := cache.GetCompiledFunction(fn); compiled != nil {
		return compiled, nil
	}

	// Mark as compiling
	cache.StartCompiling(fn)

	// Compile
	compiled, err := jit.Compile(fn, constants, globals)
	if err != nil {
		cache.MarkFailed(fn)
		return nil, err
	}

	// Cache the result
	cache.Put(fn, compiled)

	return compiled, nil
}
