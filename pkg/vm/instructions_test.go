// pkg/vm/instructions_test.go
// Tests for instruction decoding
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
)

// ============================================
// DecodeReg3 Tests
// ============================================

func TestDecodeReg3(t *testing.T) {
	// Create a 4-byte instruction: opcode, dst, src1, src2
	code := []byte{byte(compiler.OpRegAdd), 1, 2, 3}

	dst, src1, src2 := DecodeReg3(code, 0)
	if dst != 1 || src1 != 2 || src2 != 3 {
		t.Errorf("DecodeReg3() = (%d, %d, %d), expected (1, 2, 3)", dst, src1, src2)
	}
}

func TestDecodeReg3AtOffset(t *testing.T) {
	// Create code with prefix, then instruction at offset
	code := []byte{0xFF, 0xFF, byte(compiler.OpRegAdd), 5, 6, 7}

	dst, src1, src2 := DecodeReg3(code, 2)
	if dst != 5 || src1 != 6 || src2 != 7 {
		t.Errorf("DecodeReg3() = (%d, %d, %d), expected (5, 6, 7)", dst, src1, src2)
	}
}

// ============================================
// DecodeReg2 Tests
// ============================================

func TestDecodeReg2(t *testing.T) {
	code := []byte{byte(compiler.OpRegMove), 10, 20}

	dst, src := DecodeReg2(code, 0)
	if dst != 10 || src != 20 {
		t.Errorf("DecodeReg2() = (%d, %d), expected (10, 20)", dst, src)
	}
}

// ============================================
// DecodeReg1 Tests
// ============================================

func TestDecodeReg1(t *testing.T) {
	code := []byte{byte(compiler.OpRegReturn), 42}

	reg := DecodeReg1(code, 0)
	if reg != 42 {
		t.Errorf("DecodeReg1() = %d, expected 42", reg)
	}
}

// ============================================
// DecodeConst Tests
// ============================================

func TestDecodeConst(t *testing.T) {
	// Constant index is 16-bit, big-endian
	code := []byte{byte(compiler.OpRegLoadConst), 5, 0x01, 0x00} // reg=5, constIdx=256

	reg, constIdx := DecodeConst(code, 0)
	if reg != 5 {
		t.Errorf("DecodeConst() reg = %d, expected 5", reg)
	}
	if constIdx != 256 {
		t.Errorf("DecodeConst() constIdx = %d, expected 256", constIdx)
	}
}

func TestDecodeConstMaxIndex(t *testing.T) {
	// Test max 16-bit value
	code := []byte{byte(compiler.OpRegLoadConst), 0, 0xFF, 0xFF} // constIdx=65535

	_, constIdx := DecodeConst(code, 0)
	if constIdx != 65535 {
		t.Errorf("DecodeConst() constIdx = %d, expected 65535", constIdx)
	}
}

// ============================================
// DecodeJump Tests
// ============================================

func TestDecodeJump(t *testing.T) {
	// Positive offset
	code := []byte{byte(compiler.OpRegJump), 0, 0x00, 100} // offset=100

	offset := DecodeJump(code, 0)
	if offset != 100 {
		t.Errorf("DecodeJump() = %d, expected 100", offset)
	}
}

func TestDecodeJumpNegative(t *testing.T) {
	// Negative offset (signed 16-bit)
	// -1 in two's complement 16-bit is 0xFFFF = 65535 unsigned
	code := []byte{byte(compiler.OpRegJump), 0, 0xFF, 0xFF}

	offset := DecodeJump(code, 0)
	if offset != -1 {
		t.Errorf("DecodeJump() = %d, expected -1", offset)
	}
}

func TestDecodeJumpNegativeLarge(t *testing.T) {
	// -100 in two's complement 16-bit
	// 65536 - 100 = 65436 = 0xFF9C
	code := []byte{byte(compiler.OpRegJump), 0, 0xFF, 0x9C}

	offset := DecodeJump(code, 0)
	if offset != -100 {
		t.Errorf("DecodeJump() = %d, expected -100", offset)
	}
}

// ============================================
// DecodeJumpCond Tests
// ============================================

func TestDecodeJumpCond(t *testing.T) {
	code := []byte{byte(compiler.OpRegJumpIfFalse), 3, 0x01, 0x00} // condReg=3, offset=256

	condReg, offset := DecodeJumpCond(code, 0)
	if condReg != 3 {
		t.Errorf("DecodeJumpCond() condReg = %d, expected 3", condReg)
	}
	if offset != 256 {
		t.Errorf("DecodeJumpCond() offset = %d, expected 256", offset)
	}
}

func TestDecodeJumpCondNegative(t *testing.T) {
	// -10 in two's complement 16-bit
	code := []byte{byte(compiler.OpRegJumpIfFalse), 5, 0xFF, 0xF6}

	condReg, offset := DecodeJumpCond(code, 0)
	if condReg != 5 {
		t.Errorf("DecodeJumpCond() condReg = %d, expected 5", condReg)
	}
	if offset != -10 {
		t.Errorf("DecodeJumpCond() offset = %d, expected -10", offset)
	}
}

// ============================================
// DecodeCall Tests
// ============================================

func TestDecodeCall(t *testing.T) {
	code := []byte{byte(compiler.OpRegCall), 10, 3} // funcReg=10, numArgs=3

	funcReg, numArgs := DecodeCall(code, 0)
	if funcReg != 10 {
		t.Errorf("DecodeCall() funcReg = %d, expected 10", funcReg)
	}
	if numArgs != 3 {
		t.Errorf("DecodeCall() numArgs = %d, expected 3", numArgs)
	}
}

// ============================================
// DecodeClosure Tests
// ============================================

func TestDecodeClosure(t *testing.T) {
	// dstReg=1, funcIdx=256, numFree=2
	code := []byte{byte(compiler.OpClosure), 1, 0x01, 0x00, 2}

	dstReg, funcIdx, numFree := DecodeClosure(code, 0)
	if dstReg != 1 {
		t.Errorf("DecodeClosure() dstReg = %d, expected 1", dstReg)
	}
	if funcIdx != 256 {
		t.Errorf("DecodeClosure() funcIdx = %d, expected 256", funcIdx)
	}
	if numFree != 2 {
		t.Errorf("DecodeClosure() numFree = %d, expected 2", numFree)
	}
}

// ============================================
// DecodeMethodCall Tests
// ============================================

func TestDecodeMethodCall(t *testing.T) {
	// objReg=5, nameIdx=100, numArgs=2
	code := []byte{byte(compiler.OpRegCallMethod), 5, 0x00, 100, 2}

	objReg, nameIdx, numArgs := DecodeMethodCall(code, 0)
	if objReg != 5 {
		t.Errorf("DecodeMethodCall() objReg = %d, expected 5", objReg)
	}
	if nameIdx != 100 {
		t.Errorf("DecodeMethodCall() nameIdx = %d, expected 100", nameIdx)
	}
	if numArgs != 2 {
		t.Errorf("DecodeMethodCall() numArgs = %d, expected 2", numArgs)
	}
}

// ============================================
// DecodeField Tests
// ============================================

func TestDecodeField(t *testing.T) {
	// reg=1, objReg=2, nameIdx=300
	code := []byte{byte(compiler.OpGetField), 1, 2, 0x01, 0x2C}

	reg, objReg, nameIdx := DecodeField(code, 0)
	if reg != 1 {
		t.Errorf("DecodeField() reg = %d, expected 1", reg)
	}
	if objReg != 2 {
		t.Errorf("DecodeField() objReg = %d, expected 2", objReg)
	}
	if nameIdx != 300 {
		t.Errorf("DecodeField() nameIdx = %d, expected 300", nameIdx)
	}
}

// ============================================
// DecodeBuiltin Tests
// ============================================

func TestDecodeBuiltin(t *testing.T) {
	code := []byte{byte(compiler.OpRegBuiltin), 10, 3} // builtinIdx=10, numArgs=3

	builtinIdx, numArgs := DecodeBuiltin(code, 0)
	if builtinIdx != 10 {
		t.Errorf("DecodeBuiltin() builtinIdx = %d, expected 10", builtinIdx)
	}
	if numArgs != 3 {
		t.Errorf("DecodeBuiltin() numArgs = %d, expected 3", numArgs)
	}
}

// ============================================
// DecodeInstruction Tests
// ============================================

func TestDecodeInstruction(t *testing.T) {
	// Test decoding a simple instruction
	code := []byte{byte(compiler.OpRegAdd), 1, 2, 3}

	info, consumed := DecodeInstruction(code, 0)
	if consumed != 4 {
		t.Errorf("DecodeInstruction() consumed = %d, expected 4", consumed)
	}
	if info.Opcode != compiler.OpRegAdd {
		t.Errorf("DecodeInstruction() opcode = %v, expected OpRegAdd", info.Opcode)
	}
	if info.Dst != 1 || info.Src1 != 2 || info.Src2 != 3 {
		t.Errorf("DecodeInstruction() operands = (%d, %d, %d), expected (1, 2, 3)", info.Dst, info.Src1, info.Src2)
	}
}

func TestDecodeInstructionAtEnd(t *testing.T) {
	// Try to decode beyond the code
	code := []byte{byte(compiler.OpRegAdd)}

	info, consumed := DecodeInstruction(code, 0)
	if consumed != 1 {
		t.Errorf("DecodeInstruction() consumed = %d, expected 1", consumed)
	}
	_ = info
}

func TestDecodeInstructionEmpty(t *testing.T) {
	code := []byte{}

	info, consumed := DecodeInstruction(code, 0)
	if consumed != 0 {
		t.Errorf("DecodeInstruction() consumed = %d, expected 0", consumed)
	}
	_ = info
}

// ============================================
// InstructionLen Tests
// ============================================

func TestInstructionLen(t *testing.T) {
	tests := []struct {
		op       compiler.Opcode
		expected int
	}{
		{compiler.OpRegAdd, 4},        // opcode + 3 registers
		{compiler.OpRegMove, 3},       // opcode + 2 registers
		{compiler.OpRegReturn, 2},     // opcode + 1 register
		{compiler.OpRegLoadConst, 4},  // opcode + 1 reg + 2 byte const
		{compiler.OpRegJump, 4},       // opcode + 1 unused + 2 byte offset
		{compiler.OpRegJumpIfFalse, 4}, // opcode + 1 reg + 2 byte offset
		{compiler.OpRegCall, 3},       // opcode + funcReg + numArgs
	}

	for _, tt := range tests {
		got := InstructionLen(tt.op)
		if got != tt.expected {
			t.Errorf("InstructionLen(%v) = %d, expected %d", tt.op, got, tt.expected)
		}
	}
}

// ============================================
// IsRegisterOpcode Tests
// ============================================

func TestIsRegisterOpcode(t *testing.T) {
	registerOps := []compiler.Opcode{
		compiler.OpRegAdd,
		compiler.OpRegSub,
		compiler.OpRegMul,
		compiler.OpRegDiv,
		compiler.OpRegMove,
		compiler.OpRegLoadConst,
		compiler.OpRegJump,
		compiler.OpRegCall,
	}

	for _, op := range registerOps {
		if !IsRegisterOpcode(op) {
			t.Errorf("IsRegisterOpcode(%v) = false, expected true", op)
		}
	}

	nonRegisterOps := []compiler.Opcode{
		compiler.OpConstant,
		compiler.OpAdd,
		compiler.OpPop,
		compiler.OpTrue,
	}

	for _, op := range nonRegisterOps {
		if IsRegisterOpcode(op) {
			t.Errorf("IsRegisterOpcode(%v) = true, expected false", op)
		}
	}
}
