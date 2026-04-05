//go:build windows && amd64
// +build windows,amd64

// pkg/jit/jit_windows_test.go
// JIT tests specific to Windows AMD64 platform
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
)

// TestWindowsJITCompilerCreation tests JIT compiler creation on Windows
func TestWindowsJITCompilerCreation(t *testing.T) {
	config := DefaultJITConfig()
	config.Debug = true

	jit := NewJITCompiler(config)
	if jit == nil {
		t.Fatal("Failed to create JIT compiler")
	}

	stats := jit.GetStats()
	if stats.CompiledFunctions != 0 {
		t.Errorf("Expected 0 compiled functions, got %d", stats.CompiledFunctions)
	}

	jit.Cleanup()
}

// TestWindowsJITConfig tests JIT configuration
func TestWindowsJITConfig(t *testing.T) {
	config := DefaultJITConfig()

	if config.HotThreshold <= 0 {
		t.Errorf("HotThreshold should be positive, got %d", config.HotThreshold)
	}

	if config.MaxCodeSize <= 0 {
		t.Errorf("MaxCodeSize should be positive, got %d", config.MaxCodeSize)
	}

	if config.Debug {
		t.Error("Debug should be false by default")
	}
}

// TestWindowsJITStats tests JIT statistics
func TestWindowsJITStats(t *testing.T) {
	config := DefaultJITConfig()
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	stats := jit.GetStats()
	if stats.CompiledFunctions != 0 {
		t.Errorf("Expected 0 compiled functions, got %d", stats.CompiledFunctions)
	}
	if stats.TotalCodeSize != 0 {
		t.Errorf("Expected 0 total code size, got %d", stats.TotalCodeSize)
	}
	if stats.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits, got %d", stats.CacheHits)
	}
}

// TestWindowsCanExecuteNatively tests native execution capability detection
func TestWindowsCanExecuteNatively(t *testing.T) {
	tests := []struct {
		name      string
		code      []byte
		canNative bool
	}{
		{
			name: "PureArithmetic",
			code: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegLoadConst), 1, 0, 1,
				byte(compiler.OpRegAdd), 2, 0, 1,
				byte(compiler.OpRegReturn), 2,
			},
			canNative: true,
		},
		{
			name: "WithCall",
			code: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegCall), 1, 0,
				byte(compiler.OpRegReturn), 0,
			},
			canNative: false,
		},
		{
			name: "WithClosure",
			code: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
				byte(compiler.OpRegReturn), 0,
			},
			canNative: false,
		},
		{
			name: "WithArray",
			code: []byte{
				byte(compiler.OpRegArray), 0, 1, 2,
				byte(compiler.OpRegReturn), 0,
			},
			canNative: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := CanExecuteNatively(fn)
			if result != tt.canNative {
				t.Errorf("Expected %v, got %v", tt.canNative, result)
			}
		})
	}
}

// TestWindowsRequiresInterpreterFallback tests fallback detection
func TestWindowsRequiresInterpreterFallback(t *testing.T) {
	// Pure arithmetic - should not require fallback
	code := []byte{
		byte(compiler.OpRegLoadConst), 0, 0, 0,
		byte(compiler.OpRegLoadConst), 1, 0, 1,
		byte(compiler.OpRegAdd), 2, 0, 1,
		byte(compiler.OpRegReturn), 2,
	}
	fn := &compiler.CompiledFunction{
		Instructions:  code,
		NumLocals:     8,
		NumParameters: 0,
	}
	if RequiresInterpreterFallback(fn) {
		t.Error("Pure arithmetic should not require fallback")
	}

	// Closure - should require fallback
	code2 := []byte{
		byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
		byte(compiler.OpRegReturn), 0,
	}
	fn2 := &compiler.CompiledFunction{
		Instructions:  code2,
		NumLocals:     8,
		NumParameters: 0,
	}
	if !RequiresInterpreterFallback(fn2) {
		t.Error("Closure should require fallback")
	}
}

// TestWindowsGetJITSupportLevel tests support level detection
func TestWindowsGetJITSupportLevel(t *testing.T) {
	// Pure arithmetic - should have full support
	code := []byte{
		byte(compiler.OpRegLoadConst), 0, 0, 0,
		byte(compiler.OpRegLoadConst), 1, 0, 1,
		byte(compiler.OpRegAdd), 2, 0, 1,
		byte(compiler.OpRegReturn), 2,
	}
	fn := &compiler.CompiledFunction{
		Instructions:  code,
		NumLocals:     8,
		NumParameters: 0,
	}
	level := GetJITSupportLevel(fn)
	if level != JITSupportFull {
		t.Errorf("Expected JITSupportFull, got %v", level)
	}

	// Closure - should have no support
	code2 := []byte{
		byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
		byte(compiler.OpRegReturn), 0,
	}
	fn2 := &compiler.CompiledFunction{
		Instructions:  code2,
		NumLocals:     8,
		NumParameters: 0,
	}
	level2 := GetJITSupportLevel(fn2)
	if level2 != JITSupportNone {
		t.Errorf("Expected JITSupportNone, got %v", level2)
	}
}

// TestWindowsMemoryManager tests memory manager
func TestWindowsMemoryManager(t *testing.T) {
	m := NewJITMemoryManager()
	if m == nil {
		t.Fatal("Failed to create memory manager")
	}

	// Test handle allocation
	handle := m.AllocateHandle("test")
	if handle <= 0 {
		t.Errorf("Expected positive handle, got %d", handle)
	}

	// Test object retrieval
	obj, ok := m.GetObject(handle)
	if !ok {
		t.Error("Failed to retrieve object")
	}
	if obj != "test" {
		t.Errorf("Expected 'test', got %v", obj)
	}

	// Test release
	m.ReleaseHandle(handle)
	_, ok = m.GetObject(handle)
	if ok {
		t.Error("Handle should be released")
	}
}

// TestWindowsMemoryStats tests memory statistics
func TestWindowsMemoryStats(t *testing.T) {
	m := NewJITMemoryManager()
	stats := m.Stats()

	if stats.CodePages != 0 {
		t.Errorf("Expected 0 code pages, got %d", stats.CodePages)
	}
	if stats.ObjectHandles != 0 {
		t.Errorf("Expected 0 object handles, got %d", stats.ObjectHandles)
	}
}

// TestWindowsBuffer tests buffer
func TestWindowsBuffer(t *testing.T) {
	buf := NewJITBuffer(256)

	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer, got length %d", buf.Len())
	}

	// Write some bytes
	n := buf.Write([]byte{1, 2, 3, 4})
	if n != 4 {
		t.Errorf("Expected to write 4 bytes, wrote %d", n)
	}
	if buf.Len() != 4 {
		t.Errorf("Expected buffer length 4, got %d", buf.Len())
	}

	// Reset
	buf.Reset()
	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer after reset, got %d", buf.Len())
	}
}

// TestWindowsObjectPool tests object pool
func TestWindowsObjectPool(t *testing.T) {
	pool := NewJITObjectPool(func() interface{} {
		return make([]byte, 64)
	})

	// Get from empty pool
	obj := pool.Get()
	if obj == nil {
		t.Error("Expected non-nil object")
	}

	// Put back
	pool.Put(obj)
	if pool.Size() != 1 {
		t.Errorf("Expected pool size 1, got %d", pool.Size())
	}

	// Get again
	obj2 := pool.Get()
	if obj2 == nil {
		t.Error("Expected non-nil object")
	}
	if pool.Size() != 0 {
		t.Errorf("Expected empty pool, got size %d", pool.Size())
	}

	pool.Clear()
	if pool.Size() != 0 {
		t.Errorf("Expected empty pool after clear, got %d", pool.Size())
	}
}

// BenchmarkWindowsJITCompilerCreation benchmarks JIT compiler creation
func BenchmarkWindowsJITCompilerCreation(b *testing.B) {
	config := DefaultJITConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit := NewJITCompiler(config)
		jit.Cleanup()
	}
}

// BenchmarkWindowsMemoryManager benchmarks memory manager
func BenchmarkWindowsMemoryManager(b *testing.B) {
	m := NewJITMemoryManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handle := m.AllocateHandle("test")
		m.ReleaseHandle(handle)
	}
}

// BenchmarkWindowsBuffer benchmarks buffer
func BenchmarkWindowsBuffer(b *testing.B) {
	buf := NewJITBuffer(1024)
	data := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(data)
	}
}

// TestWindowsIterativeFibJIT tests the SAFE iterative Fibonacci JIT on Windows
// This test verifies that the JIT generates correct iterative code that won't freeze the system
func TestWindowsIterativeFibJIT(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{}, // Will be detected as fib pattern
	}

	// Use FibJITCompiler which now defaults to iterative
	fibCompiler := NewFibJITCompiler(config)
	code, err := fibCompiler.Compile(fn, nil, nil)
	if err != nil {
		t.Fatalf("Iterative compilation failed: %v", err)
	}

	t.Logf("Generated ITERATIVE Fibonacci code: %d bytes", len(code))

	// Allocate executable memory
	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test values including large ones (safe with iterative version)
	testCases := []struct {
		n        int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
		{20, 6765},
		{30, 832040},
		{35, 9227465},
	}

	for _, tc := range testCases {
		result := bridge.Call1(fnPtr, tc.n)
		if result != tc.expected {
			t.Errorf("fib(%d) = %d, expected %d", tc.n, result, tc.expected)
		} else {
			t.Logf("fib(%d) = %d ✓", tc.n, result)
		}
	}
}

// TestWindowsIterativeFibPerformance tests performance of SAFE iterative Fibonacci JIT
func TestWindowsIterativeFibPerformance(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	fn := &compiler.CompiledFunction{
		NumLocals:     1,
		NumParameters: 1,
		Instructions:  []byte{},
	}

	fibCompiler := NewFibJITCompiler(config)
	code, err := fibCompiler.Compile(fn, nil, nil)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	jitCompiler := NewJITCompiler(config)
	defer jitCompiler.Cleanup()

	mem, _, err := jitCompiler.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	// Test fib(35) with SAFE iterative version
	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		bridge.Call1(fnPtr, 35)
	}
	avgTime := time.Since(start) / time.Duration(iterations)

	result := bridge.Call1(fnPtr, 35)
	t.Logf("ITERATIVE fib(35): %v per call, result=%d", avgTime, result)

	// Test fib(40) - safe with iterative version
	start = time.Now()
	result = bridge.Call1(fnPtr, 40)
	fib40Time := time.Since(start)
	t.Logf("ITERATIVE fib(40): %v, result=%d", fib40Time, result)
}
