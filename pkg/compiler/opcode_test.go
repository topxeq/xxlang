// pkg/compiler/opcode_test.go
package compiler

import (
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		operands []int
		expected []byte
	}{
		{
			name:     "OpConstant with operand",
			op:       OpConstant,
			operands: []int{65534},
			expected: []byte{byte(OpConstant), 255, 254},
		},
		{
			name:     "OpAdd with no operand",
			op:       OpAdd,
			operands: []int{},
			expected: []byte{byte(OpAdd)},
		},
		{
			name:     "OpJump with operand",
			op:       OpJump,
			operands: []int{999},
			expected: []byte{byte(OpJump), 3, 231},
		},
		{
			name:     "OpGetLocal with single byte operand",
			op:       OpGetLocal,
			operands: []int{255},
			expected: []byte{byte(OpGetLocal), 255},
		},
		{
			name:     "OpCall with single byte operand",
			op:       OpCall,
			operands: []int{5},
			expected: []byte{byte(OpCall), 5},
		},
		{
			name:     "OpArray with operand",
			op:       OpArray,
			operands: []int{10},
			expected: []byte{byte(OpArray), 0, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instruction := Make(tt.op, tt.operands...)

			if len(instruction) != len(tt.expected) {
				t.Errorf("instruction has wrong length. want=%d, got=%d",
					len(tt.expected), len(instruction))
			}

			for i, b := range tt.expected {
				if instruction[i] != b {
					t.Errorf("wrong byte at pos %d. want=%d, got=%d",
						i, b, instruction[i])
				}
			}
		})
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		name        string
		op          Opcode
		operands    []int
		bytesToRead int
	}{
		{
			name:        "OpConstant operands",
			op:          OpConstant,
			operands:    []int{65534},
			bytesToRead: 2,
		},
		{
			name:        "OpJump operands",
			op:          OpJump,
			operands:    []int{999},
			bytesToRead: 2,
		},
		{
			name:        "OpGetLocal operands",
			op:          OpGetLocal,
			operands:    []int{255},
			bytesToRead: 1,
		},
		{
			name:        "OpCall operands",
			op:          OpCall,
			operands:    []int{5},
			bytesToRead: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instruction := Make(tt.op, tt.operands...)

			def, err := Lookup(byte(tt.op))
			if err != nil {
				t.Fatalf("definition not found: %s\n", err)
			}

			operandsRead, n := ReadOperands(def, instruction[1:])
			if n != tt.bytesToRead {
				t.Errorf("n wrong. want=%d, got=%d", tt.bytesToRead, n)
			}

			for i, want := range tt.operands {
				if operandsRead[i] != want {
					t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
				}
			}
		})
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []byte{
		byte(OpAdd),
		byte(OpConstant), 0, 1,
		byte(OpConstant), 0, 2,
		byte(OpMul),
	}

	expected := `0000 OpAdd
0001 OpConstant 1
0004 OpConstant 2
0007 OpMul
`

	got := String(instructions)

	if got != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q", expected, got)
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		name      string
		opcode    byte
		expectErr bool
	}{
		{
			name:      "valid opcode OpConstant",
			opcode:    byte(OpConstant),
			expectErr: false,
		},
		{
			name:      "valid opcode OpAdd",
			opcode:    byte(OpAdd),
			expectErr: false,
		},
		{
			name:      "invalid opcode 255",
			opcode:    255,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := Lookup(tt.opcode)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error for opcode %d, got nil", tt.opcode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for opcode %d: %s", tt.opcode, err)
				}
				if def == nil {
					t.Errorf("expected definition for opcode %d, got nil", tt.opcode)
				}
			}
		})
	}
}

func TestOpcodeNames(t *testing.T) {
	// Test that all opcodes have proper names defined
	opcodes := []Opcode{
		OpConstant, OpPop, OpDup,
		OpAdd, OpSub, OpMul, OpDiv, OpMod, OpNeg,
		OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual,
		OpAnd, OpOr, OpNot,
		OpGetLocal, OpSetLocal, OpDefineLocal, OpGetGlobal, OpSetGlobal,
		OpPushScope, OpPopScope,
		OpJump, OpJumpIfFalse, OpJumpIfTrue, OpCall, OpReturn,
		OpArray, OpMap, OpIndex, OpSetIndex,
		OpGetMethod, OpCallMethod,
		OpBuiltin,
		OpNull, OpTrue, OpFalse,
		OpBreak, OpContinue,
	}

	for _, op := range opcodes {
		def, err := Lookup(byte(op))
		if err != nil {
			t.Errorf("opcode %d has no definition", op)
			continue
		}
		if def.Name == "" {
			t.Errorf("opcode %d has empty name", op)
		}
	}
}

func TestDefinitionOperandWidths(t *testing.T) {
	tests := []struct {
		name          string
		op            Opcode
		expectedWidth []int
	}{
		{
			name:          "OpConstant has 2-byte operand",
			op:            OpConstant,
			expectedWidth: []int{2},
		},
		{
			name:          "OpAdd has no operands",
			op:            OpAdd,
			expectedWidth: []int{},
		},
		{
			name:          "OpJump has 2-byte operand",
			op:            OpJump,
			expectedWidth: []int{2},
		},
		{
			name:          "OpGetLocal has 1-byte operand",
			op:            OpGetLocal,
			expectedWidth: []int{1},
		},
		{
			name:          "OpCall has 1-byte operand",
			op:            OpCall,
			expectedWidth: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := Lookup(byte(tt.op))
			if err != nil {
				t.Fatalf("definition not found for opcode %d: %s", tt.op, err)
			}

			if len(def.OperandWidths) != len(tt.expectedWidth) {
				t.Errorf("wrong number of operand widths for %s. want=%d, got=%d",
					def.Name, len(tt.expectedWidth), len(def.OperandWidths))
				return
			}

			for i, width := range tt.expectedWidth {
				if def.OperandWidths[i] != width {
					t.Errorf("wrong operand width at index %d for %s. want=%d, got=%d",
						i, def.Name, width, def.OperandWidths[i])
				}
			}
		})
	}
}
