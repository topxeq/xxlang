// +build !windows

// bridge_amd64.s
// Assembly bridge for calling JIT-compiled code from Go (pure Go, no CGO)
// Supports System V AMD64 ABI (Linux/macOS)
//
// NOSPLIT (value 4) is critical: prevents Go from inserting stack growth checks
// which would fail during JIT code execution

// func Call0(fn *byte) int64
TEXT ·Call0(SB), 4, $0-16
    MOVQ fn+0(FP), AX
    CALL AX
    MOVQ AX, ret+8(FP)
    RET

// func Call1(fn *byte, arg1 int64) int64
TEXT ·Call1(SB), 4, $0-24
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), DI
    CALL AX
    MOVQ AX, ret+16(FP)
    RET

// func Call2(fn *byte, arg1, arg2 int64) int64
TEXT ·Call2(SB), 4, $0-32
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), DI
    MOVQ arg2+16(FP), SI
    CALL AX
    MOVQ AX, ret+24(FP)
    RET

// func Call3(fn *byte, arg1, arg2, arg3 int64) int64
TEXT ·Call3(SB), 4, $0-40
    MOVQ fn+0(FP), AX
    MOVQ arg1+8(FP), DI
    MOVQ arg2+16(FP), SI
    MOVQ arg3+24(FP), DX
    CALL AX
    MOVQ AX, ret+32(FP)
    RET
