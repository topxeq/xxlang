// pkg/jit/bridge_amd64.s
// Assembly bridge for calling native JIT code with globals pointer
// This bridges Go's calling convention to the standard x86-64 C calling convention

// In Go 1.18+, arguments are passed in registers but also stored on stack
// We read from stack to be compatible with both old and new ABI
// In C calling convention, first arg is in DI

// NOSPLIT (value 4) is critical: prevents Go from inserting stack growth checks
// which would fail during JIT code execution

// func callNative(entry uintptr, globals *int64) int64
// Stack layout: entry at +0(FP), globals at +8(FP), result at +16(FP)
// Total: 24 bytes (8+8+8)
TEXT ·callNative(SB), 4, $0-24
    // Save callee-saved registers that Go expects us to preserve
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read arguments from stack (works with both register and stack ABI)
    MOVQ entry+0(FP), AX    // Entry point
    MOVQ globals+8(FP), DI  // Globals pointer (put directly in DI for C calling convention)

    // Call the native code
    CALL AX

    // Result is now in AX (which is also Go's return register)
    // Write result to stack
    MOVQ AX, ret+16(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ BP
    RET
