// pkg/vm/frame_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

func TestNewFrame(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01, 0x02, 0x03},
		NumLocals:     3,
		NumParameters: 2,
	}

	frame := NewFrame(fn, 5)

	if frame == nil {
		t.Fatal("NewFrame returned nil")
	}

	if frame.Fn != fn {
		t.Error("Frame.Fn not set correctly")
	}

	if frame.IP != -1 {
		t.Errorf("expected initial IP to be -1, got %d", frame.IP)
	}

	if frame.BasePointer != 5 {
		t.Errorf("expected BasePointer to be 5, got %d", frame.BasePointer)
	}

	if len(frame.Locals) != 3 {
		t.Errorf("expected Locals length to be 3, got %d", len(frame.Locals))
	}
}

func TestFrameInstructions(t *testing.T) {
	instructions := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	fn := &compiler.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     0,
		NumParameters: 0,
	}

	frame := NewFrame(fn, 0)

	// Get instructions
	ins := frame.Instructions()
	if len(ins) != len(instructions) {
		t.Errorf("expected instructions length %d, got %d", len(instructions), len(ins))
	}

	// Verify instructions are a copy or reference correctly
	for i, b := range ins {
		if b != instructions[i] {
			t.Errorf("instruction mismatch at index %d: expected %d, got %d", i, instructions[i], b)
		}
	}
}

func TestFrameIP(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		NumLocals:     0,
		NumParameters: 0,
	}

	frame := NewFrame(fn, 0)

	// Initial IP should be -1
	if frame.IP != -1 {
		t.Errorf("expected initial IP to be -1, got %d", frame.IP)
	}

	// Advance IP
	frame.IP = 0
	if frame.IP != 0 {
		t.Errorf("expected IP to be 0, got %d", frame.IP)
	}

	frame.IP = 2
	if frame.IP != 2 {
		t.Errorf("expected IP to be 2, got %d", frame.IP)
	}
}

func TestFrameLocals(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01},
		NumLocals:     5,
		NumParameters: 0,
	}

	frame := NewFrame(fn, 0)

	// Locals should be initialized to correct size
	if len(frame.Locals) != 5 {
		t.Errorf("expected Locals length to be 5, got %d", len(frame.Locals))
	}

	// Set and get local variables
	frame.Locals[0] = &objects.Int{Value: 10}
	frame.Locals[1] = &objects.String{Value: "hello"}
	frame.Locals[4] = &objects.Bool{Value: true}

	if frame.Locals[0].(*objects.Int).Value != 10 {
		t.Error("local variable 0 not set correctly")
	}

	if frame.Locals[1].(*objects.String).Value != "hello" {
		t.Error("local variable 1 not set correctly")
	}

	if frame.Locals[4].(*objects.Bool).Value != true {
		t.Error("local variable 4 not set correctly")
	}

	// Unset locals should be nil
	if frame.Locals[2] != nil {
		t.Errorf("expected unset local to be nil, got %v", frame.Locals[2])
	}
	if frame.Locals[3] != nil {
		t.Errorf("expected unset local to be nil, got %v", frame.Locals[3])
	}
}

func TestFrameWithZeroLocals(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01},
		NumLocals:     0,
		NumParameters: 0,
	}

	frame := NewFrame(fn, 0)

	if len(frame.Locals) != 0 {
		t.Errorf("expected Locals length to be 0, got %d", len(frame.Locals))
	}
}

func TestFrameBasePointer(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01},
		NumLocals:     1,
		NumParameters: 1,
	}

	// Test with different base pointers
	for _, bp := range []int{0, 5, 10, 100} {
		frame := NewFrame(fn, bp)
		if frame.BasePointer != bp {
			t.Errorf("expected BasePointer %d, got %d", bp, frame.BasePointer)
		}
	}
}
