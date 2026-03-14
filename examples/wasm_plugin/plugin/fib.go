// examples/wasm_plugin/plugin/fib.go
// A WebAssembly plugin for Fibonacci calculation.
//
// IMPORTANT: This requires TinyGo, not standard Go!
// Standard Go's wasip1 target does NOT support plugin-style WASM modules.
//
// Build with TinyGo:
//   tinygo build -o fib.wasm -target=wasi fib.go
//
// Why TinyGo is required:
// - Standard Go runs _start (main) and exits, closing the module
// - TinyGo's architecture allows exported functions to remain callable
package main

import (
	"unsafe"
)

// Memory allocation for WASM
var memory []byte

//export alloc
func alloc(size uint32) uint32 {
	offset := uint32(len(memory))
	// Ensure 8-byte alignment for int64
	if offset%8 != 0 {
		offset += 8 - offset%8
	}
	newLen := offset + size
	if uint32(cap(memory)) < newLen {
		newMem := make([]byte, newLen*2)
		copy(newMem, memory)
		memory = newMem[:newLen]
	} else {
		memory = memory[:newLen]
	}
	return offset
}

//export plugin_name
func pluginName() (ptr uint32, size uint32) {
	name := "fib"
	ptr = alloc(uint32(len(name)))
	copy(memory[ptr:], name)
	return ptr, uint32(len(name))
}

//export plugin_version
func pluginVersion() (ptr uint32, size uint32) {
	version := "1.0.0-wasm"
	ptr = alloc(uint32(len(version)))
	copy(memory[ptr:], version)
	return ptr, uint32(len(version))
}

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

//export call_range_
func fibRange(n int64) (ptr uint32, count uint32) {
	if n < 0 {
		return 0, 0
	}

	count = uint32(n + 1)
	size := count * 8
	ptr = alloc(size)

	// Fill array with Fibonacci numbers
	a, b := int64(0), int64(1)
	for i := int64(0); i <= n; i++ {
		offset := ptr + uint32(i*8)
		*(*int64)(unsafe.Pointer(uintptr(offset))) = a
		a, b = b, a+b
	}

	return ptr, count
}

func main() {
	// main is required but not used for TinyGo WASM plugins
}
