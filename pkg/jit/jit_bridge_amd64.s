// +build !windows

// jit_bridge_amd64.s
// Assembly bridge for calling JIT code with minimal stack overhead
//
// NOSPLIT (value 4) is critical: prevents Go from inserting stack growth checks
// which would fail during JIT code execution

// func callNativeCodeImpl(codePtr uintptr, arg int64) int64
TEXT ·callNativeCodeImpl(SB), 4, $0-24
    // Move argument to DI (System V AMD64 ABI first argument)
    MOVQ arg+8(FP), DI

    // Move code pointer to AX
    MOVQ codePtr+0(FP), AX

    // Call the native code
    CALL AX

    // Result is in AX, which is also Go's return register
    MOVQ AX, ret+16(FP)
    RET
