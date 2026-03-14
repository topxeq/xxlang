// examples/wasm_plugin/plugin/fib.go
// A WebAssembly plugin for Fibonacci calculation.
//
// Build with TinyGo:
//   tinygo build -o fib.wasm -target=wasi fib.go
//
// This creates a cross-platform plugin that works on Windows, Linux, macOS
// without CGO.
package main

import (
	"unsafe"
)

// Plugin name - exported for the host
//export plugin_name
func pluginName() (ptr uint32, size uint32) {
	name := "fib"
	return stringToPtr(name)
}

// Plugin version - exported for the host
//export plugin_version
func pluginVersion() (ptr uint32, size uint32) {
	version := "1.0.0-wasm"
	return stringToPtr(version)
}

// Fast Fibonacci - O(n) time complexity
//export call_fast
func fibFast(n int64) int64 {
	if n <= 1 {
		return n
	}
	a, b := int64(0), int64(1)
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// Matrix Fibonacci - O(log n) time complexity
//export call_matrix
func fibMatrix(n int64) int64 {
	if n <= 1 {
		return n
	}

	// Matrix multiplication
	mul := func(a, b [2][2]int64) [2][2]int64 {
		return [2][2]int64{
			{a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
			{a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
		}
	}

	// Fast power
	result := [2][2]int64{{1, 0}, {0, 1}}
	base := [2][2]int64{{1, 1}, {1, 0}}

	for n > 0 {
		if n&1 == 1 {
			result = mul(result, base)
		}
		base = mul(base, base)
		n >>= 1
	}

	return result[0][1]
}

// Check if a number is a Fibonacci number
//export call_isFib
func isFib(n int64) int32 {
	if n < 0 {
		return 0
	}

	// Check if 5*n^2+4 or 5*n^2-4 is a perfect square
	check := func(x int64) bool {
		sqrt := int64(0)
		for sqrt*sqrt < x {
			sqrt++
		}
		return sqrt*sqrt == x
	}

	n2 := n * n
	if check(5*n2+4) || check(5*n2-4) {
		return 1
	}
	return 0
}

// Batch Fibonacci - returns fib(0) to fib(n)
// Returns pointer to array and count
//export call_range_
func fibRange(n int64) (ptr uint32, count uint32) {
	if n < 0 {
		return 0, 0
	}

	results := make([]int64, n+1)
	a, b := int64(0), int64(1)
	for i := int64(0); i <= n; i++ {
		results[i] = a
		a, b = b, a+b
	}

	// Allocate memory and copy results
	size := uint32((n + 1) * 8)
	buf := alloc(size)

	// Copy int64 values as bytes
	for i, val := range results {
		offset := buf + uint32(i*8)
		*(*int64)(unsafe.Pointer(uintptr(offset))) = val
	}

	return buf, uint32(n + 1)
}

// Memory allocation for WASM
var memory []byte

//export alloc
func alloc(size uint32) uint32 {
	offset := uint32(len(memory))
	memory = append(memory, make([]byte, size)...)
	return offset
}

// Helper: convert string to pointer and size
func stringToPtr(s string) (uint32, uint32) {
	ptr := alloc(uint32(len(s)))
	copy(memory[ptr:], s)
	return ptr, uint32(len(s))
}

func main() {}
