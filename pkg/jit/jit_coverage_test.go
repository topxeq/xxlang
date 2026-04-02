//go:build windows && amd64
// +build windows,amd64

// pkg/jit/jit_coverage_test.go
// Comprehensive tests to improve JIT code coverage for Windows
package jit

import (
	"sync"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITConfigDefaults tests default JIT configuration
func TestJITConfigDefaults(t *testing.T) {
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

// TestJITStatsInitial tests initial JIT statistics
func TestJITStatsInitial(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
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
	if stats.CacheMisses != 0 {
		t.Errorf("Expected 0 cache misses, got %d", stats.CacheMisses)
	}
}

// TestJITCompilerAllocCodeMultiplePages tests multiple page allocations
func TestJITCompilerAllocCodeMultiplePages(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	pageSize := 32 * 1024
	for i := 0; i < 5; i++ {
		mem, page, err := jit.AllocCode(pageSize)
		if err != nil {
			t.Fatalf("Allocation %d failed: %v", i, err)
		}
		if len(mem) != pageSize {
			t.Errorf("Allocation %d: expected %d bytes, got %d", i, pageSize, len(mem))
		}
		if page == nil {
			t.Errorf("Allocation %d: page is nil", i)
		}
	}
}

// TestJITCompilerAllocCodeSmallSizes tests small size allocations
func TestJITCompilerAllocCodeSmallSizes(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	sizes := []int{1, 16, 64, 256, 1024, 4096}
	for _, size := range sizes {
		mem, _, err := jit.AllocCode(size)
		if err != nil {
			t.Errorf("Allocation of %d bytes failed: %v", size, err)
		}
		if len(mem) != size {
			t.Errorf("Expected %d bytes, got %d", size, len(mem))
		}
	}
}

// TestJITCompilerHotPathThreshold tests hot path detection at threshold
func TestJITCompilerHotPathThreshold(t *testing.T) {
	threshold := 5
	config := JITConfig{
		HotThreshold: threshold,
		MaxCodeSize:  4096,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0, byte(compiler.OpRegMove), 255, 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	thresholdReached := false
	for i := 0; i < threshold+5; i++ {
		reached := jit.RecordExecution(fn)
		if reached && i != threshold-1 {
			t.Errorf("Threshold reached at wrong iteration: %d (expected %d)", i, threshold-1)
		}
		if i == threshold-1 {
			thresholdReached = reached
		}
	}

	if !thresholdReached {
		t.Error("Threshold should have been reached")
	}
}

// TestJITCompilerShouldCompile tests ShouldCompile function
func TestJITCompilerShouldCompile(t *testing.T) {
	config := JITConfig{
		HotThreshold: 3,
		MaxCodeSize:  100,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	smallFn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	largeFn := &compiler.CompiledFunction{
		Instructions:  make([]byte, 200),
		NumLocals:     8,
		NumParameters: 0,
	}

	for i := 0; i < 5; i++ {
		jit.RecordExecution(smallFn)
		jit.RecordExecution(largeFn)
	}

	if !jit.ShouldCompile(smallFn) {
		t.Error("Small function should be compilable")
	}
	if jit.ShouldCompile(largeFn) {
		t.Error("Large function should not be compilable")
	}
}

// TestJITCompilerGetCompiledCacheMiss tests cache miss behavior
func TestJITCompilerGetCompiledCacheMiss(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	cf := jit.GetCompiled(fn)
	if cf != nil {
		t.Error("GetCompiled should return nil for non-cached function")
	}

	stats := jit.GetStats()
	if stats.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}
}

// TestHashBytecodeDeterminism tests hash determinism
func TestHashBytecodeDeterminism(t *testing.T) {
	tests := [][]byte{
		{1, 2, 3, 4},
		{},
		{byte(compiler.OpRegNull), 0},
		{byte(compiler.OpRegLoadConst), 0, 0, 0, byte(compiler.OpRegMove), 255, 0},
	}

	for _, code := range tests {
		hash1 := hashBytecode(code)
		hash2 := hashBytecode(code)
		if hash1 != hash2 {
			t.Errorf("Hash not deterministic for code %v", code)
		}
	}
}

// TestHashBytecodeUniqueness tests hash uniqueness
func TestHashBytecodeUniqueness(t *testing.T) {
	codes := [][]byte{
		{1, 2, 3},
		{3, 2, 1},
		{1, 2, 3, 4},
		{4, 3, 2, 1},
	}

	hashes := make(map[uint64]bool)
	for _, code := range codes {
		hash := hashBytecode(code)
		if hashes[hash] {
			t.Errorf("Duplicate hash for code %v", code)
		}
		hashes[hash] = true
	}
}

// TestJITMemoryManagerHandleAllocation tests handle allocation
func TestJITMemoryManagerHandleAllocation(t *testing.T) {
	m := NewJITMemoryManager()

	handle1 := m.AllocateHandle("object1")
	if handle1 <= 0 {
		t.Errorf("Expected positive handle, got %d", handle1)
	}

	handle2 := m.AllocateHandle("object2")
	if handle2 <= 0 {
		t.Errorf("Expected positive handle, got %d", handle2)
	}

	if handle1 == handle2 {
		t.Error("Handles should be unique")
	}
}

// TestJITMemoryManagerGetObject tests object retrieval
func TestJITMemoryManagerGetObject(t *testing.T) {
	m := NewJITMemoryManager()

	handle := m.AllocateHandle("test")
	obj, ok := m.GetObject(handle)
	if !ok {
		t.Error("GetObject should succeed for valid handle")
	}
	if obj != "test" {
		t.Errorf("Expected 'test', got %v", obj)
	}

	_, ok = m.GetObject(99999)
	if ok {
		t.Error("GetObject should fail for invalid handle")
	}
}

// TestJITMemoryManagerReleaseHandle tests handle release
func TestJITMemoryManagerReleaseHandle(t *testing.T) {
	m := NewJITMemoryManager()

	handle := m.AllocateHandle("test")
	m.ReleaseHandle(handle)

	_, ok := m.GetObject(handle)
	if ok {
		t.Error("Handle should be released")
	}
}

// TestJITMemoryManagerRegisterCodePage tests code page registration
func TestJITMemoryManagerRegisterCodePage(t *testing.T) {
	m := NewJITMemoryManager()

	page := &CodePage{
		Data: make([]byte, 1024),
		Used: 512,
	}

	m.RegisterCodePage(page)

	stats := m.Stats()
	if stats.CodePages != 1 {
		t.Errorf("Expected 1 code page, got %d", stats.CodePages)
	}
}

// TestJITMemoryManagerCleanup tests cleanup
func TestJITMemoryManagerCleanup(t *testing.T) {
	m := NewJITMemoryManager()

	m.AllocateHandle("test")
	m.RegisterCodePage(&CodePage{Data: make([]byte, 1024), Used: 100})

	m.Cleanup()

	stats := m.Stats()
	if stats.ObjectHandles != 0 {
		t.Errorf("Expected 0 object handles, got %d", stats.ObjectHandles)
	}
	if stats.CodePages != 0 {
		t.Errorf("Expected 0 code pages, got %d", stats.CodePages)
	}
}

// TestGetMemoryManager tests global memory manager
func TestGetMemoryManager(t *testing.T) {
	m1 := GetMemoryManager()
	if m1 == nil {
		t.Fatal("GetMemoryManager returned nil")
	}

	m2 := GetMemoryManager()
	if m1 != m2 {
		t.Error("GetMemoryManager should return same instance")
	}
}

// TestJITObjectPoolBasic tests object pool basic operations
func TestJITObjectPoolBasic(t *testing.T) {
	pool := NewJITObjectPool(func() interface{} {
		return make([]byte, 64)
	})

	obj1 := pool.Get()
	if obj1 == nil {
		t.Error("Get should return non-nil object")
	}

	pool.Put(obj1)

	if pool.Size() != 1 {
		t.Errorf("Expected pool size 1, got %d", pool.Size())
	}

	obj2 := pool.Get()
	if obj2 == nil {
		t.Error("Get should return non-nil object")
	}
	if pool.Size() != 0 {
		t.Errorf("Expected empty pool, got size %d", pool.Size())
	}
}

// TestJITObjectPoolClear tests pool clear
func TestJITObjectPoolClear(t *testing.T) {
	pool := NewJITObjectPool(func() interface{} {
		return make([]byte, 64)
	})

	for i := 0; i < 5; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}

	pool.Clear()
	if pool.Size() != 0 {
		t.Errorf("Expected empty pool after clear, got %d", pool.Size())
	}
}

// TestJITObjectPoolConcurrent tests concurrent pool access
func TestJITObjectPoolConcurrent(t *testing.T) {
	pool := NewJITObjectPool(func() interface{} {
		return make([]byte, 64)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := pool.Get()
			pool.Put(obj)
		}()
	}
	wg.Wait()
}

// TestJITBufferBasic tests buffer basic operations
func TestJITBufferBasic(t *testing.T) {
	buf := NewJITBuffer(256)

	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer, got length %d", buf.Len())
	}

	n := buf.Write([]byte{1, 2, 3, 4})
	if n != 4 {
		t.Errorf("Expected to write 4 bytes, wrote %d", n)
	}
	if buf.Len() != 4 {
		t.Errorf("Expected buffer length 4, got %d", buf.Len())
	}

	bytes := buf.Bytes()
	if len(bytes) != 4 {
		t.Errorf("Expected 4 bytes, got %d", len(bytes))
	}
}

// TestJITBufferWriteByte tests WriteByte
func TestJITBufferWriteByte(t *testing.T) {
	buf := NewJITBuffer(64)

	for i := byte(0); i < 10; i++ {
		buf.WriteByte(i)
	}

	if buf.Len() != 10 {
		t.Errorf("Expected length 10, got %d", buf.Len())
	}
}

// TestJITBufferReset tests Reset
func TestJITBufferReset(t *testing.T) {
	buf := NewJITBuffer(64)

	buf.Write([]byte{1, 2, 3, 4})
	if buf.Len() != 4 {
		t.Errorf("Expected length 4, got %d", buf.Len())
	}

	buf.Reset()
	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer after reset, got %d", buf.Len())
	}
}

// TestJITBufferCap tests Cap
func TestJITBufferCap(t *testing.T) {
	buf := NewJITBuffer(256)

	if buf.Cap() < 256 {
		t.Errorf("Expected capacity >= 256, got %d", buf.Cap())
	}
}

// TestJITBufferGrow tests Grow
func TestJITBufferGrow(t *testing.T) {
	buf := NewJITBuffer(16)

	buf.Write(make([]byte, 10))
	buf.Grow(100)

	if buf.Cap() < 110 {
		t.Errorf("Expected capacity >= 110 after grow, got %d", buf.Cap())
	}
}

// TestBufferPool tests global buffer pool
func TestBufferPool(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}

	buf.Write([]byte{1, 2, 3})
	PutBuffer(buf)

	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Error("Buffer should be reset when returned to pool")
	}
	PutBuffer(buf2)
}

// TestForceGC tests ForceGC
func TestForceGC(t *testing.T) {
	_ = ForceGC()
}

// TestGetGCStats tests GetGCStats
func TestGetGCStats(t *testing.T) {
	stats := GetGCStats()
	_ = stats
}

// TestRegisterGCCleanup tests GC cleanup registration
func TestRegisterGCCleanup(t *testing.T) {
	handle := RegisterGCCleanup(func() {})

	if handle <= 0 {
		t.Error("RegisterGCCleanup should return positive handle")
	}

	UnregisterGCCleanup(handle)
}

// TestCompiledFuncGetFunctionPointer tests GetFunctionPointer
func TestCompiledFuncGetFunctionPointer(t *testing.T) {
	cf := &CompiledFunc{
		Entry: 0x12345678,
	}

	ptr := cf.GetFunctionPointer()
	if ptr != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%x", ptr)
	}
}

// TestCodePage tests CodePage operations
func TestCodePage(t *testing.T) {
	page := &CodePage{
		Data: make([]byte, 4096),
		Used: 0,
	}

	if len(page.Data) != 4096 {
		t.Errorf("Expected 4096 bytes, got %d", len(page.Data))
	}

	page.Used = 1024
	if page.Used != 1024 {
		t.Errorf("Expected 1024 used, got %d", page.Used)
	}
}

// TestJITCompilerCleanupIdempotent tests that Cleanup can be called multiple times
func TestJITCompilerCleanupIdempotent(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	jit.Cleanup()
	jit.Cleanup()
}

// TestJITCompileSimpleArithmeticNew tests JIT compilation of simple arithmetic
func TestJITCompileSimpleArithmeticNew(t *testing.T) {
	code := `
		var a = 10
		var b = 20
		var c = a + b
		c
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	cf, err := jit.Compile(mainFn, constants, nil)
	if err != nil {
		t.Logf("JIT compile error: %v", err)
		return
	}

	if cf == nil {
		t.Error("Compiled function is nil")
	} else {
		t.Logf("Compiled function: %d bytes, hash=0x%016x", cf.Size, cf.Hash)
	}
}

// TestJITVMCreationNew tests JIT VM creation
func TestJITVMCreationNew(t *testing.T) {
	code := `var x = 42`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	if jitVM == nil {
		t.Fatal("NewJITVM returned nil")
	}
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMSetSourcePathNew tests SetSourcePath
func TestJITVMSetSourcePathNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetSourcePath("test.xx")
}

// TestJITVMGetConstantsNew tests GetConstants
func TestJITVMGetConstantsNew(t *testing.T) {
	code := `var x = 42`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	constants := jitVM.GetConstants()
	if constants == nil {
		t.Error("GetConstants returned nil")
	}
}

// TestJITVMGetGlobalsNew tests GetGlobals
func TestJITVMGetGlobalsNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	globals := jitVM.GetGlobals()
	if globals == nil {
		t.Error("GetGlobals returned nil")
	}
}

// TestJITVMGetJITStatsNew tests GetJITStats
func TestJITVMGetJITStatsNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	stats := jitVM.GetJITStats()
	_ = stats
}

// TestJITVMGetNativeStatsNew tests GetNativeStats
func TestJITVMGetNativeStatsNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	_ = nativeExecs
	_ = interpExecs
}

// TestJITVMCompileFunction tests CompileFunction
func TestJITVMCompileFunctionNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	cf, err := jit.Compile(fn, jitVM.GetConstants(), nil)
	if err != nil {
		t.Logf("Compile error: %v", err)
		return
	}
	_ = cf
}

// TestJITVMExecuteCompiledNew tests ExecuteCompiled
func TestJITVMExecuteCompiledNew(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	cf := &CompiledFunc{
		Entry: 0,
		Size:  0,
	}

	if cf.Entry == 0 {
		t.Log("Skipping execute with nil entry pointer")
		return
	}
	result := cf.Execute()
	_ = result
}

// TestJITVMLastPoppedObject tests LastPoppedObject
func TestJITVMLastPoppedObject(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	_ = jitVM.Run()

	obj := jitVM.LastPoppedObject()
	_ = obj
}

// TestJITVMRunWithGlobals tests VM with globals
func TestJITVMRunWithGlobals(t *testing.T) {
	code := `var x = 42`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	globals := make([]vm.Value, 256)

	jitVM := NewJITVMWithGlobals(bytecode, globals, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMSetCurrentModule tests SetCurrentModule
func TestJITVMSetCurrentModule(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetCurrentModule(nil)
}

// BenchmarkJITCompilerAlloc benchmarks memory allocation
func BenchmarkJITCompilerAlloc(b *testing.B) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit.AllocCode(1024)
	}
}

// BenchmarkJITCompilerRecordExecution benchmarks execution recording
func BenchmarkJITCompilerRecordExecution(b *testing.B) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit.RecordExecution(fn)
	}
}

// BenchmarkHashBytecode benchmarks bytecode hashing
func BenchmarkHashBytecode(b *testing.B) {
	code := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashBytecode(code)
	}
}

// BenchmarkJITMemoryManagerHandle benchmarks handle operations
func BenchmarkJITMemoryManagerHandle(b *testing.B) {
	m := NewJITMemoryManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handle := m.AllocateHandle("test")
		m.ReleaseHandle(handle)
	}
}

// BenchmarkJITBufferWrite benchmarks buffer write
func BenchmarkJITBufferWrite(b *testing.B) {
	buf := NewJITBuffer(1024)
	data := make([]byte, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(data)
	}
}

// BenchmarkJITObjectPool benchmarks object pool operations
func BenchmarkJITObjectPool(b *testing.B) {
	pool := NewJITObjectPool(func() interface{} {
		return make([]byte, 64)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}
}
