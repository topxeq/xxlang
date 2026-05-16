// +build windows

// bridge_windows_amd64.s
// Assembly bridge for calling JIT-compiled code from Go on Windows
// Uses Microsoft x64 calling convention (not System V)

// Microsoft x64 calling convention:
// - First 4 integer args: rcx, rdx, r8, r9
// - Return value: rax
// - Callee-saved: rbx, rbp, rdi, rsi, r12-r15, xmm6-xmm15

// func Call0(fn *byte) int64
TEXT ·Call0(SB), 4, $0-16
    MOVQ fn+0(FP), AX
    CALL AX
    MOVQ AX, ret+8(FP)
    RET

// func Call1(fn *byte, arg1 int64) int64
TEXT ·Call1(SB), 4, $0-24
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), CX    // Windows: first arg in rcx
    CALL AX
    MOVQ AX, ret+16(FP)
    RET

// func Call2(fn *byte, arg1, arg2 int64) int64
TEXT ·Call2(SB), 4, $0-32
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), CX    // Windows: first arg in rcx
    MOVQ arg2+16(FP), DX   // Windows: second arg in rdx
    CALL AX
    MOVQ AX, ret+24(FP)
    RET

// func Call3(fn *byte, arg1, arg2, arg3 int64) int64
TEXT ·Call3(SB), 4, $0-40
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), CX    // Windows: first arg in rcx
    MOVQ arg2+16(FP), DX   // Windows: second arg in rdx
    MOVQ arg3+24(FP), R8   // Windows: third arg in r8
    CALL AX
    MOVQ AX, ret+32(FP)
    RET
