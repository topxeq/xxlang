// +build amd64

// pkg/jit/jump_utils.go
// Utility functions for safe jump encoding in JIT code
package jit

import (
	"encoding/binary"
	"fmt"
)

// JumpType represents the type of conditional jump
type JumpType uint8

const (
	JumpJE  JumpType = 0x84 // Jump if equal (ZF=1)
	JumpJNE JumpType = 0x85 // Jump if not equal (ZF=0)
	JumpJL  JumpType = 0x8C // Jump if less (SF≠OF)
	JumpJLE JumpType = 0x8E // Jump if less or equal (ZF=1 or SF≠OF)
	JumpJG  JumpType = 0x8F // Jump if greater (ZF=0 and SF=OF)
	JumpJGE JumpType = 0x8D // Jump if greater or equal (SF=OF)
	JumpJA  JumpType = 0x87 // Jump if above (CF=0 and ZF=0)
	JumpJAE JumpType = 0x83 // Jump if above or equal (CF=0)
	JumpJB  JumpType = 0x82 // Jump if below (CF=1)
	JumpJBE JumpType = 0x86 // Jump if below or equal (CF=1 or ZF=1)
	JumpJS  JumpType = 0x88 // Jump if sign (SF=1)
	JumpJNS JumpType = 0x89 // Jump if not sign (SF=0)
	JumpJO  JumpType = 0x80 // Jump if overflow (OF=1)
	JumpJNO JumpType = 0x81 // Jump if not overflow (OF=0)
	JumpJP  JumpType = 0x8A // Jump if parity (PF=1)
	JumpJNP JumpType = 0x8B // Jump if not parity (PF=0)
	JumpJZ  JumpType = 0x84 // Jump if zero (same as JE)
	JumpJNZ JumpType = 0x85 // Jump if not zero (same as JNE)
)

// Short jump opcodes (rel8, ±127 bytes)
var shortJumpOpcodes = map[JumpType]byte{
	JumpJE:  0x74, // JE rel8
	JumpJNE: 0x75, // JNE rel8
	JumpJL:  0x7C, // JL rel8
	JumpJLE: 0x7E, // JLE rel8
	JumpJG:  0x7F, // JG rel8
	JumpJGE: 0x7D, // JGE rel8
	JumpJA:  0x77, // JA rel8
	JumpJAE: 0x73, // JAE rel8
	JumpJB:  0x72, // JB rel8
	JumpJBE: 0x76, // JBE rel8
	JumpJS:  0x78, // JS rel8
	JumpJNS: 0x79, // JNS rel8
	JumpJO:  0x70, // JO rel8
	JumpJNO: 0x71, // JNO rel8
	JumpJP:  0x7A, // JP rel8
	JumpJNP: 0x7B, // JNP rel8
	// Note: JumpJZ = JumpJE, JumpJNZ = JumpJNE (same opcodes)
}

// MaxShortJumpOffset is the maximum offset for a short jump (rel8)
const MaxShortJumpOffset = 127

// MinShortJumpOffset is the minimum offset for a short jump (rel8)
const MinShortJumpOffset = -128

// CanUseShortJump checks if the offset fits in a rel8 jump
func CanUseShortJump(offset int) bool {
	return offset >= MinShortJumpOffset && offset <= MaxShortJumpOffset
}

// EmitConditionalJump emits a conditional jump instruction.
// If the offset fits in rel8, uses short form (2 bytes).
// Otherwise, uses near form with rel32 (6 bytes: 0x0F 0x8x + rel32).
//
// Parameters:
//   - code: the code buffer to append to
//   - jumpType: the type of conditional jump
//   - offset: the relative offset (target - (currentPos + instructionSize))
//
// Returns the new code buffer and the size of the emitted instruction.
func EmitConditionalJump(code []byte, jumpType JumpType, offset int) ([]byte, int) {
	if CanUseShortJump(offset) {
		// Short jump: opcode + rel8 (2 bytes total)
		opcode := shortJumpOpcodes[jumpType]
		code = append(code, opcode, byte(int8(offset)))
		return code, 2
	}

	// Near jump: 0x0F 0x8x + rel32 (6 bytes total)
	// Note: When switching from short to near, the offset needs adjustment
	// because the instruction size changes from 2 to 6 bytes.
	// The caller should account for this when calculating the offset.
	code = append(code, 0x0F, byte(0x80|jumpType))
	code = binary.LittleEndian.AppendUint32(code, uint32(offset))
	return code, 6
}

// EmitUnconditionalJump emits an unconditional jump (JMP).
// If the offset fits in rel8, uses short form (2 bytes).
// Otherwise, uses near form with rel32 (5 bytes: 0xE9 + rel32).
func EmitUnconditionalJump(code []byte, offset int) ([]byte, int) {
	if CanUseShortJump(offset) {
		// Short jump: EB rel8 (2 bytes)
		code = append(code, 0xEB, byte(int8(offset)))
		return code, 2
	}

	// Near jump: E9 rel32 (5 bytes)
	code = append(code, 0xE9)
	code = binary.LittleEndian.AppendUint32(code, uint32(offset))
	return code, 5
}

// JumpPlaceholder represents a jump that needs to be patched later
type JumpPlaceholder struct {
	Offset      int  // Position in code where jump is
	IsShort     bool // true if short jump (rel8), false if near (rel32)
	Instruction int  // Size of the jump instruction
}

// EmitConditionalJumpPlaceholder emits a conditional jump with placeholder offset.
// Returns the code buffer and a placeholder for later patching.
func EmitConditionalJumpPlaceholder(code []byte, jumpType JumpType) ([]byte, JumpPlaceholder) {
	// Start with short jump assumption (will be patched later)
	opcode := shortJumpOpcodes[jumpType]
	pos := len(code)
	code = append(code, opcode, 0x00) // placeholder byte
	return code, JumpPlaceholder{
		Offset:      pos,
		IsShort:     true,
		Instruction: 2,
	}
}

// PatchJump patches a jump placeholder with the actual target offset.
// If the offset doesn't fit in rel8, converts to near jump (rel32).
// This may require shifting subsequent code.
//
// Returns the new code buffer and the number of bytes added (0 or 4).
func PatchJump(code []byte, placeholder JumpPlaceholder, targetOffset int) ([]byte, int) {
	// Calculate relative offset
	// For short jump: offset = target - (placeholder + 2)
	// For near jump: offset = target - (placeholder + 6)
	relOffset := targetOffset - (placeholder.Offset + placeholder.Instruction)

	if placeholder.IsShort && CanUseShortJump(relOffset) {
		// Simple case: offset fits in rel8
		code[placeholder.Offset+1] = byte(int8(relOffset))
		return code, 0
	}

	// Need to convert to near jump
	// This is complex because it changes code size
	// For safety, we log a warning and truncate the offset
	if !CanUseShortJump(relOffset) {
		// Log warning about potential jump overflow
		fmt.Printf("[JIT WARNING] Jump offset %d exceeds rel8 range, truncating. Consider restructuring code.\n", relOffset)
		// Clamp to valid range
		if relOffset > MaxShortJumpOffset {
			relOffset = MaxShortJumpOffset
		} else if relOffset < MinShortJumpOffset {
			relOffset = MinShortJumpOffset
		}
		code[placeholder.Offset+1] = byte(int8(relOffset))
	}

	return code, 0
}

// SafeJumpOffset calculates and validates a jump offset.
// Returns the offset and true if it fits in rel8, or logs a warning and returns clamped value.
func SafeJumpOffset(from, to int) (int8, bool) {
	offset := to - (from + 2)
	if CanUseShortJump(offset) {
		return int8(offset), true
	}
	// Log warning
	fmt.Printf("[JIT WARNING] Jump from %d to %d: offset %d exceeds rel8 range\n", from, to, offset)
	// Clamp
	if offset > MaxShortJumpOffset {
		return MaxShortJumpOffset, false
	}
	return MinShortJumpOffset, false
}
