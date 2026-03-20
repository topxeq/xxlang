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

// func callNativeWithArgs(entry uintptr, globals *int64, arg0, arg1, arg2 int64) int64
// Calls native code with initial register values set
// Stack layout:
//   entry at +0(FP), globals at +8(FP)
//   arg0 at +16(FP), arg1 at +24(FP), arg2 at +32(FP)
//   result at +40(FP)
TEXT ·callNativeWithArgs(SB), 4, $0-48
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read entry point and globals
    MOVQ entry+0(FP), R15   // Entry point (save in R15, a callee-saved reg)
    MOVQ globals+8(FP), DI  // Globals pointer

    // Load initial arguments into VM registers
    // Native code generator maps:
    //   VM reg 0 = RAX, VM reg 1 = RBX, VM reg 2 = RCX
    // We need to be careful: RAX is used for return value
    // Load args to callee-saved regs first, then move to VM regs
    MOVQ arg0+16(FP), R12   // arg0
    MOVQ arg1+24(FP), R13   // arg1
    MOVQ arg2+32(FP), R14   // arg2

    // Move to VM registers
    MOVQ R12, AX            // VM reg 0 = RAX = arg0 (n)
    MOVQ R13, BX            // VM reg 1 = RBX = arg1 (a)
    MOVQ R14, CX            // VM reg 2 = RCX = arg2 (b)

    // Call the native code (entry in R15)
    CALL R15

    // Result is now in AX
    MOVQ AX, ret+40(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ BP
    RET
