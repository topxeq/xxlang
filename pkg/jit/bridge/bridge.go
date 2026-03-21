// +build amd64

// Package bridge provides assembly bridges for calling JIT-compiled code from Go.
// This is a pure Go solution (no CGO required).
//
// Platform Support:
//   - Linux/macOS (AMD64): Uses System V AMD64 ABI (args in rdi, rsi, rdx)
//   - Windows (AMD64): Uses Microsoft x64 ABI (args in rcx, rdx, r8, r9)
//   - ARM64: Not supported (stub implementations in bridge_arm64.go)
//
// The JIT code generator (BuildFibCode) automatically generates
// platform-specific code for the target platform.
package bridge

// Call0 calls a JIT function with no arguments.
// The JIT function must follow System V AMD64 ABI.
// Returns the result from rax.
func Call0(fn *byte) int64

// Call1 calls a JIT function with 1 argument.
// Arguments are passed in: rdi (System V AMD64 ABI)
// Returns the result from rax.
func Call1(fn *byte, arg1 int64) int64

// Call2 calls a JIT function with 2 arguments.
// Arguments are passed in: rdi, rsi
// Returns the result from rax.
func Call2(fn *byte, arg1, arg2 int64) int64

// Call3 calls a JIT function with 3 arguments.
// Arguments are passed in: rdi, rsi, rdx
// Returns the result from rax.
func Call3(fn *byte, arg1, arg2, arg3 int64) int64
