// jit_bridge.go
// Go declarations for assembly bridge functions

package main

// callJitCode0 calls a JIT function with no arguments
// The JIT function must follow System V AMD64 ABI
func callJitCode0(fn *byte) int64

// callJitCode1 calls a JIT function with 1 argument
// Arguments are passed in: rdi
func callJitCode1(fn *byte, arg1 int64) int64

// callJitCode2 calls a JIT function with 2 arguments
// Arguments are passed in: rdi, rsi
func callJitCode2(fn *byte, arg1, arg2 int64) int64

// callJitCode3 calls a JIT function with 3 arguments
// Arguments are passed in: rdi, rsi, rdx
func callJitCode3(fn *byte, arg1, arg2, arg3 int64) int64
