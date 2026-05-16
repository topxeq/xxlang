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
    PUSHQ DI       // RDI is callee-saved on Windows x64
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
    POPQ DI        // Restore RDI
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
    PUSHQ DI       // RDI is callee-saved on Windows x64
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
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func callNativeWithArgs8(entry uintptr, globals *int64, args *int64) int64
// Calls native code with 8 arguments passed via args pointer
// Stack layout:
//   entry at +0(FP), globals at +8(FP), args at +16(FP), result at +24(FP)
// Arguments are loaded from args array into VM registers:
//   args[0] -> RAX (VM reg 0)
//   args[1] -> RBX (VM reg 1)
//   args[2] -> RCX (VM reg 2)
//   args[3] -> RDX (VM reg 3)
//   args[4] -> R8  (VM reg 4)
//   args[5] -> R9  (VM reg 5)
//   args[6] -> R10 (VM reg 6)
//   args[7] -> R11 (VM reg 7)
TEXT ·callNativeWithArgs8(SB), 4, $0-32
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    // We need to save all registers we'll use
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read entry point and globals
    MOVQ entry+0(FP), R15   // Entry point (save in R15)
    MOVQ globals+8(FP), DI  // Globals pointer

    // Load args pointer
    MOVQ args+16(FP), R14   // args pointer

    // Load all 8 arguments into VM registers
    MOVQ 0(R14), AX         // args[0] -> RAX (VM reg 0)
    MOVQ 8(R14), BX         // args[1] -> RBX (VM reg 1)
    MOVQ 16(R14), CX        // args[2] -> RCX (VM reg 2)
    MOVQ 24(R14), DX        // args[3] -> RDX (VM reg 3)
    MOVQ 32(R14), R8        // args[4] -> R8  (VM reg 4)
    MOVQ 40(R14), R9        // args[5] -> R9  (VM reg 5)
    MOVQ 48(R14), R10       // args[6] -> R10 (VM reg 6)
    MOVQ 56(R14), R11       // args[7] -> R11 (VM reg 7)

    // Call the native code (entry in R15)
    CALL R15

    // Result is now in AX
    MOVQ AX, ret+24(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func callBuiltinCallback(callback uintptr, builtinIdx, numArgs int, argsPtr *int64) int64
// Calls a Go callback for builtin functions from native code
// Stack layout:
//   callback at +0(FP), builtinIdx at +8(FP), numArgs at +16(FP)
//   argsPtr at +24(FP), result at +32(FP)
TEXT ·callBuiltinCallback(SB), 4, $0-40
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read callback pointer
    MOVQ callback+0(FP), R15  // Callback function pointer

    // Set up arguments for Go callback (System V ABI):
    //   RDI = builtinIdx
    //   RSI = numArgs
    //   RDX = argsPtr
    MOVQ builtinIdx+8(FP), DI
    MOVQ numArgs+16(FP), SI
    MOVQ argsPtr+24(FP), DX

    // Call the Go callback
    CALL R15

    // Result is now in AX
    MOVQ AX, ret+32(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func callFunctionCallback(callback uintptr, funcReg, numArgs int, argsPtr *int64) int64
// Calls a Go callback for function dispatch from native code
// Stack layout:
//   callback at +0(FP), funcReg at +8(FP), numArgs at +16(FP)
//   argsPtr at +24(FP), result at +32(FP)
TEXT ·callFunctionCallback(SB), 4, $0-40
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read callback pointer
    MOVQ callback+0(FP), R15  // Callback function pointer

    // Set up arguments for Go callback (System V ABI):
    //   RDI = funcReg
    //   RSI = numArgs
    //   RDX = argsPtr
    MOVQ funcReg+8(FP), DI
    MOVQ numArgs+16(FP), SI
    MOVQ argsPtr+24(FP), DX

    // Call the Go callback
    CALL R15

    // Result is now in AX
    MOVQ AX, ret+32(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func callCollectionCallback(callback uintptr, opKind, numArgs int, argsPtr *int64) int64
// Calls a Go callback for collection operations from native code
// Stack layout:
//   callback at +0(FP), opKind at +8(FP), numArgs at +16(FP)
//   argsPtr at +24(FP), result at +32(FP)
TEXT ·callCollectionCallback(SB), 4, $0-40
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Read callback pointer
    MOVQ callback+0(FP), R15  // Callback function pointer

    // Set up arguments for Go callback (System V ABI):
    //   RDI = opKind
    //   RSI = numArgs
    //   RDX = argsPtr
    MOVQ opKind+8(FP), DI
    MOVQ numArgs+16(FP), SI
    MOVQ argsPtr+24(FP), DX

    // Call the Go callback
    CALL R15

    // Result is now in AX
    MOVQ AX, ret+32(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// Callback wrappers that convert from System V ABI to Go calling convention
// These are the actual callbacks that native code calls

// func builtinCallbackWrapper(builtinIdx int64, numArgs int64, argsPtr *int64) int64
// Called from native code with System V ABI args: DI=builtinIdx, SI=numArgs, DX=argsPtr
TEXT ·builtinCallbackWrapper(SB), 4, $0-32
    // Save callee-saved registers
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ SI       // RSI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Arguments are already in DI, SI, DX (System V ABI)
    // We need to convert to Go ABI (stack-based for Go 1.17+, but let's use the Go function)

    // Call the Go function CallBuiltinFromNative
    // Move args to appropriate locations for Go call
    MOVQ DI, AX        // builtinIdx
    MOVQ SI, BX        // numArgs
    MOVQ DX, CX        // argsPtr

    // Allocate stack space for Go call (Go passes args on stack in some cases)
    SUBQ $32, SP
    MOVQ AX, 0(SP)     // builtinIdx
    MOVQ BX, 8(SP)     // numArgs
    MOVQ CX, 16(SP)    // argsPtr

    CALL ·CallBuiltinFromNative(SB)

    // Result is in AX
    ADDQ $32, SP
    MOVQ AX, ret+24(FP)

    // Restore callee-saved registers
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ SI        // Restore RSI
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func functionCallbackWrapper(funcReg int64, numArgs int64, argsPtr *int64) int64
// Called from native code with System V ABI args: DI=funcReg, SI=numArgs, DX=argsPtr
TEXT ·functionCallbackWrapper(SB), 4, $0-32
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ SI       // RSI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    MOVQ DI, AX
    MOVQ SI, BX
    MOVQ DX, CX

    SUBQ $32, SP
    MOVQ AX, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ CX, 16(SP)

    CALL ·CallFunctionFromNative(SB)

    ADDQ $32, SP
    MOVQ AX, ret+24(FP)

    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ SI        // Restore RSI
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func collectionCallbackWrapper(opKind int64, numArgs int64, argsPtr *int64) int64
// Called from native code with System V ABI args: DI=opKind, SI=numArgs, DX=argsPtr
TEXT ·collectionCallbackWrapper(SB), 4, $0-32
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ SI       // RSI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    MOVQ DI, AX
    MOVQ SI, BX
    MOVQ DX, CX

    SUBQ $32, SP
    MOVQ AX, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ CX, 16(SP)

    CALL ·CallCollectionFromNative(SB)

    ADDQ $32, SP
    MOVQ AX, ret+24(FP)

    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ SI        // Restore RSI
    POPQ DI        // Restore RDI
    POPQ BP
    RET

// func objectCallbackWrapper(opKind int64, numArgs int64, argsPtr *int64, nameIdx int64) int64
// Called from native code with System V ABI args: DI=opKind, SI=numArgs, DX=argsPtr, CX=nameIdx
TEXT ·objectCallbackWrapper(SB), 4, $0-40
    PUSHQ BP
    MOVQ SP, BP
    PUSHQ DI       // RDI is callee-saved on Windows x64
    PUSHQ SI       // RSI is callee-saved on Windows x64
    PUSHQ BX
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Save nameIdx (in CX) before we use it
    MOVQ CX, R12        // nameIdx

    MOVQ DI, AX         // opKind
    MOVQ SI, BX         // numArgs
    // DX already has argsPtr

    SUBQ $40, SP
    MOVQ AX, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ DX, 16(SP)
    MOVQ R12, 24(SP)    // nameIdx

    CALL ·CallObjectFromNative(SB)

    ADDQ $40, SP
    MOVQ AX, ret+32(FP)

    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ BX
    POPQ SI        // Restore RSI
    POPQ DI        // Restore RDI
    POPQ BP
    RET
