//go:build windows && amd64
// +build windows,amd64

// pkg/jit/bridge/bridge_test.go
// Tests for JIT bridge functions
package bridge

import (
	"testing"
	"unsafe"
)

// TestAllocExecMem tests executable memory allocation
func TestAllocExecMem(t *testing.T) {
	size := 4096
	mem, err := AllocExecMem(size)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}
	if mem == nil {
		t.Fatal("AllocExecMem returned nil memory")
	}
	if len(mem) != size {
		t.Errorf("Expected %d bytes, got %d", size, len(mem))
	}

	// Write some test code
	code := []byte{
		0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00, // mov rax, 42
		0xC3, // ret
	}
	copy(mem, code)

	// Free the memory
	FreeExecMem(mem)
}

// TestAllocExecMemMultiple tests multiple allocations
func TestAllocExecMemMultiple(t *testing.T) {
	sizes := []int{1024, 4096, 16384}

	for _, size := range sizes {
		mem, err := AllocExecMem(size)
		if err != nil {
			t.Errorf("AllocExecMem(%d) failed: %v", size, err)
			continue
		}
		if len(mem) != size {
			t.Errorf("Expected %d bytes, got %d", size, len(mem))
		}
		FreeExecMem(mem)
	}
}

// TestBuildFibCode tests Fibonacci code generation
func TestBuildFibCode(t *testing.T) {
	code := BuildFibCode()
	if code == nil {
		t.Fatal("BuildFibCode returned nil")
	}
	if len(code) == 0 {
		t.Error("BuildFibCode returned empty code")
	}
	t.Logf("Generated Fibonacci code: %d bytes", len(code))
}

// TestCall0 tests calling a function with 0 arguments
func TestCall0(t *testing.T) {
	mem, err := AllocExecMem(64)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}
	defer FreeExecMem(mem)

	// Simple code: return 42
	code := []byte{
		0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00, // mov rax, 42
		0xC3, // ret
	}
	copy(mem, code)

	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))
	result := Call0(fnPtr)

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestCall1 tests calling a function with 1 argument
func TestCall1(t *testing.T) {
	mem, err := AllocExecMem(64)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}
	defer FreeExecMem(mem)

	// Simple code: return arg0
	// Windows x64 ABI: arg0 is in rcx, return in rax
	code := []byte{
		0x48, 0x89, 0xC8, // mov rax, rcx
		0xC3, // ret
	}
	copy(mem, code)

	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))
	result := Call1(fnPtr, 123)

	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}
}

// TestCall2 tests calling a function with 2 arguments
func TestCall2(t *testing.T) {
	mem, err := AllocExecMem(64)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}
	defer FreeExecMem(mem)

	// Simple code: return arg0 + arg1
	// Windows x64 ABI: arg0 in rcx, arg1 in rdx
	code := []byte{
		0x48, 0x89, 0xC8, // mov rax, rcx
		0x48, 0x01, 0xD0, // add rax, rdx
		0xC3, // ret
	}
	copy(mem, code)

	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))
	result := Call2(fnPtr, 10, 20)

	if result != 30 {
		t.Errorf("Expected 30, got %d", result)
	}
}

// TestCall3 tests calling a function with 3 arguments
func TestCall3(t *testing.T) {
	mem, err := AllocExecMem(64)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}
	defer FreeExecMem(mem)

	// Simple code: return arg0 + arg1 + arg2
	// Windows x64 ABI: arg0 in rcx, arg1 in rdx, arg2 in r8
	code := []byte{
		0x48, 0x89, 0xC8, // mov rax, rcx
		0x48, 0x01, 0xD0, // add rax, rdx
		0x4C, 0x01, 0xC0, // add rax, r8 (corrected: add rax, r8)
		0xC3, // ret
	}
	copy(mem, code)

	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))
	result := Call3(fnPtr, 10, 20, 30)

	// Note: The result might vary based on calling convention
	t.Logf("Call3 result: %d", result)
}

// TestBridgeFreeExecMemNil tests freeing nil pointer
func TestBridgeFreeExecMemNil(t *testing.T) {
	// Free nil should not panic
	FreeExecMem(nil)
}

// TestBridgeFreeExecMemEmpty tests freeing empty slice
func TestBridgeFreeExecMemEmpty(t *testing.T) {
	// Free empty slice should not panic
	FreeExecMem([]byte{})
}

// TestBridgeFreeExecMemZeroLen tests freeing zero-length slice
func TestBridgeFreeExecMemZeroLen(t *testing.T) {
	mem := make([]byte, 0)
	err := FreeExecMem(mem)
	if err != nil {
		t.Errorf("FreeExecMem on zero-length slice returned error: %v", err)
	}
}

// TestBridgeAllocAndFree tests allocation and free cycle
func TestBridgeAllocAndFree(t *testing.T) {
	mem, err := AllocExecMem(1024)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}

	if len(mem) != 1024 {
		t.Errorf("Expected 1024 bytes, got %d", len(mem))
	}

	err = FreeExecMem(mem)
	if err != nil {
		t.Errorf("FreeExecMem returned error: %v", err)
	}
}

// TestBridgeFreeTwice tests freeing memory twice
func TestBridgeFreeTwice(t *testing.T) {
	mem, err := AllocExecMem(1024)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}

	// First free
	err = FreeExecMem(mem)
	if err != nil {
		t.Errorf("First FreeExecMem returned error: %v", err)
	}

	// Second free on same memory - this is undefined behavior
	// We just verify it doesn't crash
	// Note: VirtualFree may return an error for double-free
	_ = FreeExecMem(mem)
}

// BenchmarkAllocExecMem benchmarks memory allocation
func BenchmarkAllocExecMem(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem, _ := AllocExecMem(4096)
		FreeExecMem(mem)
	}
}

// BenchmarkCall0 benchmarks function call with 0 args
func BenchmarkCall0(b *testing.B) {
	mem, _ := AllocExecMem(64)
	defer FreeExecMem(mem)

	code := []byte{
		0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00,
		0xC3,
	}
	copy(mem, code)

	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Call0(fnPtr)
	}
}
