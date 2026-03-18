// pkg/vm/instructions.go
// Instruction decoder for register-based VM
package vm

import (
	"github.com/topxeq/xxlang/pkg/compiler"
)

// InstructionInfo holds decoded instruction information
type InstructionInfo struct {
	Opcode   compiler.Opcode
	Dst      byte
	Src1     byte
	Src2     byte
	ConstIdx int
	Offset   int
	NumArgs  byte
}

// DecodeInstruction decodes a single instruction from bytecode
// Returns the instruction info and the number of bytes consumed
func DecodeInstruction(code []byte, ip int) (InstructionInfo, int) {
	if ip >= len(code) {
		return InstructionInfo{}, 0
	}

	op := compiler.Opcode(code[ip])
	info := InstructionInfo{Opcode: op}

	// Get the opcode definition
	defs := compiler.GetDefinitions()
	def := defs[op]
	if def == nil {
		// Unknown opcode, skip 1 byte
		return info, 1
	}

	// Calculate total instruction length
	totalLen := 1
	for _, w := range def.OperandWidths {
		totalLen += w
	}

	// Check if we have enough bytes
	if ip+totalLen > len(code) {
		return info, 1
	}

	// Decode operands based on widths
	offset := ip + 1
	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			val := code[offset]
			switch i {
			case 0:
				info.Dst = val
				info.Src1 = val // Also store as Src1 for convenience
			case 1:
				info.Src1 = val
				info.Src2 = val // Also store as Src2 for convenience
			case 2:
				info.Src2 = val
			}
		case 2:
			val := int(code[offset])<<8 | int(code[offset+1])
			switch i {
			case 0:
				info.ConstIdx = val
				info.Offset = val // Also store as Offset for jumps
			case 1:
				info.ConstIdx = val
				info.Offset = val
			case 2:
				info.ConstIdx = val
			}
		}
		offset += width
	}

	return info, totalLen
}

// DecodeReg3 decodes a 3-operand register instruction: dst, src1, src2
func DecodeReg3(code []byte, ip int) (dst, src1, src2 byte) {
	return code[ip+1], code[ip+2], code[ip+3]
}

// DecodeReg2 decodes a 2-operand register instruction: dst, src
func DecodeReg2(code []byte, ip int) (dst, src byte) {
	return code[ip+1], code[ip+2]
}

// DecodeReg1 decodes a 1-operand register instruction: reg
func DecodeReg1(code []byte, ip int) (reg byte) {
	return code[ip+1]
}

// DecodeConst decodes an instruction with register and constant index
func DecodeConst(code []byte, ip int) (reg byte, constIdx int) {
	return code[ip+1], int(code[ip+2])<<8 | int(code[ip+3])
}

// DecodeJump decodes a jump instruction with 16-bit offset
func DecodeJump(code []byte, ip int) (offset int) {
	return int(code[ip+2])<<8 | int(code[ip+3])
}

// DecodeJumpCond decodes a conditional jump instruction
func DecodeJumpCond(code []byte, ip int) (condReg byte, offset int) {
	return code[ip+1], int(code[ip+2])<<8 | int(code[ip+3])
}

// DecodeCall decodes a call instruction
func DecodeCall(code []byte, ip int) (funcReg byte, numArgs byte) {
	return code[ip+1], code[ip+2]
}

// DecodeClosure decodes a closure creation instruction
func DecodeClosure(code []byte, ip int) (dstReg byte, funcIdx int, numFree byte) {
	return code[ip+1], int(code[ip+2])<<8 | int(code[ip+3]), code[ip+4]
}

// DecodeMethodCall decodes a method call instruction
func DecodeMethodCall(code []byte, ip int) (objReg byte, nameIdx int, numArgs byte) {
	return code[ip+1], int(code[ip+2])<<8 | int(code[ip+3]), code[ip+4]
}

// DecodeField decodes a field access instruction
func DecodeField(code []byte, ip int) (reg byte, objReg byte, nameIdx int) {
	return code[ip+1], code[ip+2], int(code[ip+3])<<8 | int(code[ip+4])
}

// DecodeBuiltin decodes a builtin call instruction
func DecodeBuiltin(code []byte, ip int) (builtinIdx byte, numArgs byte) {
	return code[ip+1], code[ip+2]
}

// InstructionLen returns the length of an instruction in bytes
func InstructionLen(op compiler.Opcode) int {
	def := compiler.GetDefinitions()[op]
	if def == nil {
		return 1
	}

	length := 1
	for _, w := range def.OperandWidths {
		length += w
	}
	return length
}

// IsRegisterOpcode checks if an opcode is a register-based operation
func IsRegisterOpcode(op compiler.Opcode) bool {
	return compiler.IsRegisterOpcode(op)
}
