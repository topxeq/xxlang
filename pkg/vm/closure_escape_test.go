// pkg/vm/closure_escape_test.go
// Tests for closure and escape analysis helpers
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Closure Tests
// ============================================

func TestClosureType(t *testing.T) {
	c := &Closure{
		Fn: &compiler.CompiledFunction{},
	}

	if c.Type() != objects.ClosureType {
		t.Errorf("Closure.Type() = %s, expected %s", c.Type(), objects.ClosureType)
	}
}

func TestClosureTypeTag(t *testing.T) {
	c := &Closure{}

	if c.TypeTag() != objects.TagClosure {
		t.Errorf("Closure.TypeTag() = %d, expected %d", c.TypeTag(), objects.TagClosure)
	}
}

func TestClosureInspect(t *testing.T) {
	tests := []struct {
		freeVars int
		expected string
	}{
		{0, "closure[0 freeVars]"},
		{1, "closure[1 freeVars]"},
		{3, "closure[3 freeVars]"},
	}

	for _, tt := range tests {
		c := &Closure{
			Fn:       &compiler.CompiledFunction{},
			FreeVars: make([]objects.Object, tt.freeVars),
		}

		if c.Inspect() != tt.expected {
			t.Errorf("Closure.Inspect() = %s, expected %s", c.Inspect(), tt.expected)
		}
	}
}

func TestClosureToBool(t *testing.T) {
	c := &Closure{}

	if c.ToBool() != objects.TRUE {
		t.Error("Closure.ToBool() should return TRUE")
	}
}

func TestClosureHashKey(t *testing.T) {
	c := &Closure{}

	hk := c.HashKey()
	if hk.Type != objects.ClosureType {
		t.Errorf("HashKey.Type = %s, expected %s", hk.Type, objects.ClosureType)
	}
}

func TestClosureWithFreeVars(t *testing.T) {
	freeVars := []objects.Object{
		&objects.Int{Value: 42},
		&objects.String{Value: "hello"},
	}

	c := &Closure{
		Fn:        &compiler.CompiledFunction{},
		FreeVars:  freeVars,
		Constants: []objects.Object{},
		Globals:   []objects.Object{},
	}

	if len(c.FreeVars) != 2 {
		t.Errorf("Expected 2 free vars, got %d", len(c.FreeVars))
	}
}

func TestClosureWithFreeVarsValues(t *testing.T) {
	freeVarsValues := []Value{
		NewInt(42),
		NewFloat(3.14),
	}

	c := &Closure{
		Fn:              &compiler.CompiledFunction{},
		FreeVarsValues:  freeVarsValues,
	}

	if len(c.FreeVarsValues) != 2 {
		t.Errorf("Expected 2 free var values, got %d", len(c.FreeVarsValues))
	}
}

// ============================================
// Escape Analysis Helper Tests
// ============================================

func TestEscapeToHeap(t *testing.T) {
	obj := &objects.Int{Value: 42}
	result := escapeToHeap(obj)

	if result != obj {
		t.Error("escapeToHeap should return the same object")
	}
}

func TestStackBool(t *testing.T) {
	// Test true
	result := stackBool(true)
	if result != objects.TRUE {
		t.Error("stackBool(true) should return TRUE")
	}

	// Test false
	result = stackBool(false)
	if result != objects.FALSE {
		t.Error("stackBool(false) should return FALSE")
	}
}

func TestNullIfZero(t *testing.T) {
	// Test with nil
	result := nullIfZero(nil)
	if result != objects.NULL {
		t.Error("nullIfZero(nil) should return NULL")
	}

	// Test with non-nil object
	obj := &objects.Int{Value: 42}
	result = nullIfZero(obj)
	if result != obj {
		t.Error("nullIfZero with non-nil should return the object")
	}
}

func TestIntResult(t *testing.T) {
	result := intResult(42)

	intObj, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("Expected *objects.Int, got %T", result)
	}
	if intObj.Value != 42 {
		t.Errorf("intResult(42) = %d, expected 42", intObj.Value)
	}
}

func TestFloatResult(t *testing.T) {
	result := floatResult(3.14)

	floatObj, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("Expected *objects.Float, got %T", result)
	}
	if floatObj.Value != 3.14 {
		t.Errorf("floatResult(3.14) = %f, expected 3.14", floatObj.Value)
	}
}

// ============================================
// RegFrame Tests
// ============================================

func TestNewRegFrame(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{0x01, 0x02, 0x03},
		NumLocals:     2,
		NumRegs:       10,
		FreeVariables: []compiler.Symbol{},
	}

	frame := NewRegFrame(fn)
	if frame == nil {
		t.Fatal("NewRegFrame returned nil")
	}

	if frame.Fn != fn {
		t.Error("Frame function not set correctly")
	}

	if frame.IP != 0 {
		t.Errorf("IP = %d, expected 0", frame.IP)
	}

	// Release the frame
	frame.Release()
}

func TestRegFrameGetSetReg(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}
	frame := NewRegFrame(fn)
	defer frame.Release()

	// Set a register
	frame.SetReg(5, NewInt(42))

	// Get the register
	val := frame.GetReg(5)
	if !val.IsInt() {
		t.Fatal("Expected int value")
	}

	got, _ := val.ToInt()
	if got != 42 {
		t.Errorf("GetReg(5) = %d, expected 42", got)
	}
}

func TestRegFrameGetSetLocal(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 5, NumRegs: 10}
	frame := NewRegFrame(fn)
	defer frame.Release()

	// Set a local
	frame.SetLocal(2, NewInt(100))

	// Get the local
	val := frame.GetLocal(2)
	if !val.IsInt() {
		t.Fatal("Expected int value")
	}

	got, _ := val.ToInt()
	if got != 100 {
		t.Errorf("GetLocal(2) = %d, expected 100", got)
	}
}

func TestRegFrameGetLocalOutOfRange(t *testing.T) {
	fn := &compiler.CompiledFunction{NumLocals: 2, NumRegs: 10}
	frame := NewRegFrame(fn)
	defer frame.Release()

	// Get out of range local
	val := frame.GetLocal(10)
	if val != ValueNull {
		t.Error("Out of range local should return ValueNull")
	}
}

func TestRegFrameGetSetFree(t *testing.T) {
	fn := &compiler.CompiledFunction{
		NumRegs:       10,
		FreeVariables: []compiler.Symbol{{}, {}}, // 2 free vars
	}
	frame := NewRegFrame(fn)
	defer frame.Release()

	// Set a free variable
	frame.SetFree(0, NewInt(42))

	// Get the free variable
	val := frame.GetFree(0)
	if !val.IsInt() {
		t.Fatal("Expected int value")
	}

	got, _ := val.ToInt()
	if got != 42 {
		t.Errorf("GetFree(0) = %d, expected 42", got)
	}
}

func TestRegFrameGetFreeOutOfRange(t *testing.T) {
	fn := &compiler.CompiledFunction{
		NumRegs:       10,
		FreeVariables: []compiler.Symbol{},
	}
	frame := NewRegFrame(fn)
	defer frame.Release()

	val := frame.GetFree(0)
	if val != ValueNull {
		t.Error("Out of range free var should return ValueNull")
	}
}

func TestRegFrameGetConstant(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}
	frame := NewRegFrame(fn)
	frame.Constants = []Value{NewInt(1), NewInt(2), NewInt(3)}
	defer frame.Release()

	val := frame.GetConstant(1)
	if !val.IsInt() {
		t.Fatal("Expected int value")
	}

	got, _ := val.ToInt()
	if got != 2 {
		t.Errorf("GetConstant(1) = %d, expected 2", got)
	}
}

func TestRegFrameGetConstantOutOfRange(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}
	frame := NewRegFrame(fn)
	frame.Constants = []Value{}
	defer frame.Release()

	val := frame.GetConstant(0)
	if val != ValueNull {
		t.Error("Out of range constant should return ValueNull")
	}
}

func TestRegFrameGetSetGlobal(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}
	frame := NewRegFrame(fn)
	frame.Globals = make([]Value, 10)
	defer frame.Release()

	// Set a global
	frame.SetGlobal(5, NewInt(42))

	// Get the global
	val := frame.GetGlobal(5)
	if !val.IsInt() {
		t.Fatal("Expected int value")
	}

	got, _ := val.ToInt()
	if got != 42 {
		t.Errorf("GetGlobal(5) = %d, expected 42", got)
	}
}

func TestRegFrameGetGlobalOutOfRange(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}
	frame := NewRegFrame(fn)
	frame.Globals = []Value{}
	defer frame.Release()

	val := frame.GetGlobal(0)
	if val != ValueNull {
		t.Error("Out of range global should return ValueNull")
	}
}

func TestRegFrameInstructions(t *testing.T) {
	instructions := []byte{0x01, 0x02, 0x03}
	fn := &compiler.CompiledFunction{
		Instructions: instructions,
		NumRegs:      10,
	}
	frame := NewRegFrame(fn)
	defer frame.Release()

	got := frame.Instructions()
	if len(got) != len(instructions) {
		t.Errorf("Instructions length = %d, expected %d", len(got), len(instructions))
	}
}

func TestRegFrameCopyArgRegisters(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}

	src := NewRegFrame(fn)
	src.Registers[0] = NewInt(1)
	src.Registers[1] = NewInt(2)
	src.Registers[2] = NewInt(3)

	dst := NewRegFrame(fn)
	dst.CopyArgRegisters(src, 3)

	// Check arguments were copied
	if dst.Registers[0] != NewInt(1) {
		t.Error("Arg 0 not copied correctly")
	}
	if dst.Registers[1] != NewInt(2) {
		t.Error("Arg 1 not copied correctly")
	}
	if dst.Registers[2] != NewInt(3) {
		t.Error("Arg 2 not copied correctly")
	}

	src.Release()
	dst.Release()
}

func TestRegFrameClearArgRegisters(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}

	frame := NewRegFrame(fn)
	frame.Registers[0] = NewInt(1)
	frame.Registers[1] = NewInt(2)

	frame.ClearArgRegisters()

	// Check arguments were cleared
	for i := 0; i < 8; i++ {
		if frame.Registers[i] != ValueNull {
			t.Errorf("Register %d not cleared", i)
		}
	}

	frame.Release()
}

func TestRegFramePool(t *testing.T) {
	fn := &compiler.CompiledFunction{NumRegs: 10}

	// Create and release multiple frames
	for i := 0; i < 10; i++ {
		frame := NewRegFrame(fn)
		frame.SetReg(0, NewInt(int64(i)))
		frame.Release()
	}

	// Just verify no panic occurred
}
