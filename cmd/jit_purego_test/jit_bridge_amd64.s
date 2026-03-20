// +build !windows

// jit_bridge_amd64.s
// Assembly bridge for calling JIT-compiled code from Go
// This is pure Go (no CGO) - just assembly

// Go assembly parameter layout (amd64):
// - Arguments are at positive offsets from FP
// - Return values follow arguments
// - Total frame size is specified in the TEXT directive

// func callJitCode0(fn *byte) int64
// Frame: fn(8) + ret(8) = 16 bytes
TEXT ·callJitCode0(SB), 0, $0-16
    // Load function pointer
    MOVQ fn+0(FP), AX

    // Call the JIT function
    CALL AX

    // Result is in AX, store to return value
    MOVQ AX, ret+8(FP)
    RET

// func callJitCode1(fn *byte, arg1 int64) int64
// Frame: fn(8) + arg1(8) + ret(8) = 24 bytes
TEXT ·callJitCode1(SB), 0, $0-24
    // Load function pointer
    MOVQ fn+0(FP), AX

    // Load argument into DI (System V ABI: first arg in rdi)
    MOVQ arg1+8(FP), DI

    // Call the JIT function
    CALL AX

    // Result is in AX, store to return value
    MOVQ AX, ret+16(FP)
    RET

// func callJitCode2(fn *byte, arg1, arg2 int64) int64
// Frame: fn(8) + arg1(8) + arg2(8) + ret(8) = 32 bytes
TEXT ·callJitCode2(SB), 0, $0-32
    // Load function pointer
    MOVQ fn+0(FP), AX

    // Load arguments (System V ABI: rdi, rsi)
    MOVQ arg1+8(FP), DI
    MOVQ arg2+16(FP), SI

    // Call the JIT function
    CALL AX

    // Result is in AX, store to return value
    MOVQ AX, ret+24(FP)
    RET

// func callJitCode3(fn *byte, arg1, arg2, arg3 int64) int64
// Frame: fn(8) + arg1(8) + arg2(8) + arg3(8) + ret(8) = 40 bytes
TEXT ·callJitCode3(SB), 0, $0-40
    // Load function pointer
    MOVQ fn+0(FP), AX

    // Load arguments (System V ABI: rdi, rsi, rdx)
    MOVQ arg1+8(FP), DI
    MOVQ arg2+16(FP), SI
    MOVQ arg3+24(FP), DX

    // Call the JIT function
    CALL AX

    // Result is in AX, store to return value
    MOVQ AX, ret+32(FP)
    RET
