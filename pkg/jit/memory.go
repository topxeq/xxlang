//go:build amd64
// +build amd64

// pkg/jit/memory.go
// Memory management for JIT-compiled code and objects
// Provides GC integration for proper cleanup
package jit

import (
	"runtime"
	"sync"
)

// JITMemoryManager manages memory allocation for JIT
type JITMemoryManager struct {
	mu sync.RWMutex

	// Code pages allocated
	codePages []*CodePage

	// Object handles for GC visibility
	objectHandles map[int64]interface{}

	// Free handle pool
	freeHandles []int64

	// Next handle index
	nextHandle int64

	// Finalizer callback
	finalizerFunc func()
}

// globalMemoryManager is the global memory manager instance
var globalMemoryManager *JITMemoryManager
var globalMemoryOnce sync.Once

// GetMemoryManager returns the global memory manager
func GetMemoryManager() *JITMemoryManager {
	globalMemoryOnce.Do(func() {
		globalMemoryManager = NewJITMemoryManager()
	})
	return globalMemoryManager
}

// NewJITMemoryManager creates a new memory manager
func NewJITMemoryManager() *JITMemoryManager {
	m := &JITMemoryManager{
		codePages:     make([]*CodePage, 0),
		objectHandles: make(map[int64]interface{}),
		freeHandles:   make([]int64, 0),
		nextHandle:    1,
	}

	// Set up finalizer for cleanup when the manager is garbage collected
	runtime.SetFinalizer(m, func(m *JITMemoryManager) {
		m.Cleanup()
	})

	return m
}

// AllocateHandle allocates a handle for an object
// The object will be visible to GC and cleaned up when no longer needed
func (m *JITMemoryManager) AllocateHandle(obj interface{}) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try to reuse a free handle
	if len(m.freeHandles) > 0 {
		idx := len(m.freeHandles) - 1
		handle := m.freeHandles[idx]
		m.freeHandles = m.freeHandles[:idx]
		m.objectHandles[handle] = obj
		return handle
	}

	// Allocate new handle
	handle := m.nextHandle
	m.nextHandle++
	m.objectHandles[handle] = obj
	return handle
}

// GetObject retrieves an object by handle
func (m *JITMemoryManager) GetObject(handle int64) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	obj, ok := m.objectHandles[handle]
	return obj, ok
}

// ReleaseHandle releases a handle for reuse
func (m *JITMemoryManager) ReleaseHandle(handle int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.objectHandles[handle]; ok {
		delete(m.objectHandles, handle)
		m.freeHandles = append(m.freeHandles, handle)
	}
}

// RegisterCodePage registers a code page for tracking
func (m *JITMemoryManager) RegisterCodePage(page *CodePage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codePages = append(m.codePages, page)
}

// Cleanup releases all allocated resources
func (m *JITMemoryManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear object handles
	m.objectHandles = make(map[int64]interface{})
	m.freeHandles = make([]int64, 0)
	m.nextHandle = 1

	// Code pages are cleaned up by JITCompiler.Cleanup()
	m.codePages = nil
}

// Stats returns memory statistics
func (m *JITMemoryManager) Stats() JITMemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCodeSize := 0
	for _, page := range m.codePages {
		if page != nil {
			totalCodeSize += page.Used
		}
	}

	return JITMemoryStats{
		CodePages:     len(m.codePages),
		TotalCodeSize: totalCodeSize,
		ObjectHandles: len(m.objectHandles),
		FreeHandles:   len(m.freeHandles),
		NextHandle:    m.nextHandle,
	}
}

// JITMemoryStats contains memory statistics
type JITMemoryStats struct {
	CodePages     int
	TotalCodeSize int
	ObjectHandles int
	FreeHandles   int
	NextHandle    int64
}

// ============================================================================
// GC Integration
// ============================================================================

// JITGCStats tracks GC-related statistics
type JITGCStats struct {
	// Number of GC cycles that triggered JIT cleanup
	GCCycles int

	// Number of objects freed by GC
	ObjectsFreed int64

	// Number of code pages freed by GC
	PagesFreed int64
}

// globalGCStats tracks global GC statistics
var globalGCStats JITGCStats
var gcStatsMu sync.Mutex

// RegisterGCCleanup registers a cleanup function to be called on GC
// Returns a handle that can be used to unregister
func RegisterGCCleanup(cleanupFunc func()) int64 {
	handle := GetMemoryManager().AllocateHandle(cleanupFunc)
	return handle
}

// UnregisterGCCleanup removes a registered cleanup function
func UnregisterGCCleanup(handle int64) {
	GetMemoryManager().ReleaseHandle(handle)
}

// ForceGC forces a garbage collection and returns statistics
func ForceGC() JITGCStats {
	runtime.GC()
	gcStatsMu.Lock()
	defer gcStatsMu.Unlock()
	return globalGCStats
}

// GetGCStats returns current GC statistics
func GetGCStats() JITGCStats {
	gcStatsMu.Lock()
	defer gcStatsMu.Unlock()
	return globalGCStats
}

// ============================================================================
// Object Pool for JIT
// ============================================================================

// JITObjectPool is a pool of reusable objects
type JITObjectPool struct {
	mu     sync.Mutex
	pool   []interface{}
	create func() interface{}
}

// NewJITObjectPool creates a new object pool
func NewJITObjectPool(create func() interface{}) *JITObjectPool {
	return &JITObjectPool{
		pool:   make([]interface{}, 0),
		create: create,
	}
}

// Get retrieves an object from the pool or creates a new one
func (p *JITObjectPool) Get() interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pool) > 0 {
		obj := p.pool[len(p.pool)-1]
		p.pool = p.pool[:len(p.pool)-1]
		return obj
	}

	return p.create()
}

// Put returns an object to the pool
func (p *JITObjectPool) Put(obj interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pool = append(p.pool, obj)
}

// Clear empties the pool
func (p *JITObjectPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pool = make([]interface{}, 0)
}

// Size returns the number of objects in the pool
func (p *JITObjectPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pool)
}

// ============================================================================
// Memory Buffer for JIT
// ============================================================================

// JITBuffer is a reusable byte buffer
type JITBuffer struct {
	data []byte
}

// NewJITBuffer creates a new buffer with initial capacity
func NewJITBuffer(capacity int) *JITBuffer {
	return &JITBuffer{
		data: make([]byte, 0, capacity),
	}
}

// Reset resets the buffer
func (b *JITBuffer) Reset() {
	b.data = b.data[:0]
}

// Bytes returns the buffer contents
func (b *JITBuffer) Bytes() []byte {
	return b.data
}

// Write appends bytes to the buffer
func (b *JITBuffer) Write(p []byte) int {
	b.data = append(b.data, p...)
	return len(p)
}

// WriteByte appends a single byte to the buffer
func (b *JITBuffer) WriteByte(c byte) error {
	b.data = append(b.data, c)
	return nil
}

// Len returns the current length
func (b *JITBuffer) Len() int {
	return len(b.data)
}

// Cap returns the capacity
func (b *JITBuffer) Cap() int {
	return cap(b.data)
}

// Grow increases the buffer capacity
func (b *JITBuffer) Grow(n int) {
	if cap(b.data)-len(b.data) < n {
		newData := make([]byte, len(b.data), cap(b.data)+n)
		copy(newData, b.data)
		b.data = newData
	}
}

// Global buffer pool
var bufferPool = NewJITObjectPool(func() interface{} {
	return NewJITBuffer(4096)
})

// GetBuffer gets a buffer from the pool
func GetBuffer() *JITBuffer {
	return bufferPool.Get().(*JITBuffer)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(b *JITBuffer) {
	b.Reset()
	bufferPool.Put(b)
}
