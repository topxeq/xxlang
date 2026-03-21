// pkg/vm/value_ops_test.go
// Tests for value operations
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// executeBinaryOp Tests
// ============================================

func TestExecuteBinaryOpAddInt(t *testing.T) {
	left := NewInt(10)
	right := NewInt(5)

	result, err := executeBinaryOp(compiler.OpRegAdd, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Add) failed: %v", err)
	}

	if !result.IsInt() {
		t.Fatal("Expected int result")
	}

	got, _ := result.ToInt()
	if got != 15 {
		t.Errorf("10 + 5 = %d, expected 15", got)
	}
}

func TestExecuteBinaryOpSub(t *testing.T) {
	left := NewInt(10)
	right := NewInt(3)

	result, err := executeBinaryOp(compiler.OpRegSub, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Sub) failed: %v", err)
	}

	got, _ := result.ToInt()
	if got != 7 {
		t.Errorf("10 - 3 = %d, expected 7", got)
	}
}

func TestExecuteBinaryOpMul(t *testing.T) {
	left := NewInt(6)
	right := NewInt(7)

	result, err := executeBinaryOp(compiler.OpRegMul, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Mul) failed: %v", err)
	}

	got, _ := result.ToInt()
	if got != 42 {
		t.Errorf("6 * 7 = %d, expected 42", got)
	}
}

func TestExecuteBinaryOpDiv(t *testing.T) {
	left := NewFloat(10.0)
	right := NewFloat(4.0)

	result, err := executeBinaryOp(compiler.OpRegDiv, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Div) failed: %v", err)
	}

	if !result.IsFloat() {
		t.Fatal("Expected float result")
	}

	got := result.GetFloat()
	if got != 2.5 {
		t.Errorf("10 / 4 = %f, expected 2.5", got)
	}
}

func TestExecuteBinaryOpDivByZero(t *testing.T) {
	left := NewInt(10)
	right := NewInt(0)

	_, err := executeBinaryOp(compiler.OpRegDiv, left, right)
	if err == nil {
		t.Error("Division by zero should return error")
	}
}

func TestExecuteBinaryOpMod(t *testing.T) {
	left := NewInt(10)
	right := NewInt(3)

	result, err := executeBinaryOp(compiler.OpRegMod, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Mod) failed: %v", err)
	}

	got, _ := result.ToInt()
	if got != 1 {
		t.Errorf("10 %% 3 = %d, expected 1", got)
	}
}

func TestExecuteBinaryOpLess(t *testing.T) {
	left := NewInt(5)
	right := NewInt(10)

	result, err := executeBinaryOp(compiler.OpRegLess, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Less) failed: %v", err)
	}

	if result != ValueTrue {
		t.Errorf("5 < 10 = %v, expected true", result)
	}

	// Reverse
	result, err = executeBinaryOp(compiler.OpRegLess, right, left)
	if err != nil {
		t.Fatalf("executeBinaryOp(Less) failed: %v", err)
	}

	if result != ValueFalse {
		t.Errorf("10 < 5 = %v, expected false", result)
	}
}

func TestExecuteBinaryOpGreater(t *testing.T) {
	left := NewInt(10)
	right := NewInt(5)

	result, err := executeBinaryOp(compiler.OpRegGreater, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Greater) failed: %v", err)
	}

	if result != ValueTrue {
		t.Errorf("10 > 5 = %v, expected true", result)
	}
}

func TestExecuteBinaryOpEqual(t *testing.T) {
	left := NewInt(42)
	right := NewInt(42)

	result, err := executeBinaryOp(compiler.OpRegEqual, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Equal) failed: %v", err)
	}

	if result != ValueTrue {
		t.Errorf("42 == 42 = %v, expected true", result)
	}

	// Different values
	right = NewInt(43)
	result, err = executeBinaryOp(compiler.OpRegEqual, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(Equal) failed: %v", err)
	}

	if result != ValueFalse {
		t.Errorf("42 == 43 = %v, expected false", result)
	}
}

func TestExecuteBinaryOpNotEqual(t *testing.T) {
	left := NewInt(42)
	right := NewInt(43)

	result, err := executeBinaryOp(compiler.OpRegNotEqual, left, right)
	if err != nil {
		t.Fatalf("executeBinaryOp(NotEqual) failed: %v", err)
	}

	if result != ValueTrue {
		t.Errorf("42 != 43 = %v, expected true", result)
	}
}

func TestExecuteBinaryOpLessEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{5, 10, true},
		{10, 10, true},
		{15, 10, false},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)

		result, err := executeBinaryOp(compiler.OpRegLessEqual, left, right)
		if err != nil {
			t.Fatalf("executeBinaryOp(LessEqual) failed: %v", err)
		}

		expected := ValueFalse
		if tt.expected {
			expected = ValueTrue
		}

		if result != expected {
			t.Errorf("%d <= %d = %v, expected %v", tt.left, tt.right, result, expected)
		}
	}
}

func TestExecuteBinaryOpGreaterEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{10, 5, true},
		{10, 10, true},
		{5, 10, false},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)

		result, err := executeBinaryOp(compiler.OpRegGreaterEqual, left, right)
		if err != nil {
			t.Fatalf("executeBinaryOp(GreaterEqual) failed: %v", err)
		}

		expected := ValueFalse
		if tt.expected {
			expected = ValueTrue
		}

		if result != expected {
			t.Errorf("%d >= %d = %v, expected %v", tt.left, tt.right, result, expected)
		}
	}
}

func TestExecuteBinaryOpUnknown(t *testing.T) {
	left := NewInt(1)
	right := NewInt(2)

	_, err := executeBinaryOp(compiler.Opcode(0xFF), left, right)
	if err == nil {
		t.Error("Unknown opcode should return error")
	}
}

// ============================================
// binaryAddObjects Tests
// ============================================

func TestBinaryAddObjectsString(t *testing.T) {
	left := &objects.String{Value: "hello"}
	right := &objects.String{Value: " world"}

	result, err := binaryAddObjects(left, right)
	if err != nil {
		t.Fatalf("binaryAddObjects failed: %v", err)
	}

	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected *objects.String, got %T", result)
	}

	if strResult.Value != "hello world" {
		t.Errorf("String concatenation = %q, expected %q", strResult.Value, "hello world")
	}
}

func TestBinaryAddObjectsStringMismatch(t *testing.T) {
	left := &objects.String{Value: "hello"}
	right := &objects.Int{Value: 42}

	_, err := binaryAddObjects(left, right)
	if err == nil {
		t.Error("Adding int to string should fail")
	}
}

func TestBinaryAddObjectsArray(t *testing.T) {
	left := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
	}}
	right := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 3},
		&objects.Int{Value: 4},
	}}

	result, err := binaryAddObjects(left, right)
	if err != nil {
		t.Fatalf("binaryAddObjects failed: %v", err)
	}

	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("Expected *objects.Array, got %T", result)
	}

	if len(arrResult.Elements) != 4 {
		t.Errorf("Array concatenation length = %d, expected 4", len(arrResult.Elements))
	}
}

func TestBinaryAddObjectsArrayMismatch(t *testing.T) {
	left := &objects.Array{Elements: []objects.Object{}}
	right := &objects.String{Value: "test"}

	_, err := binaryAddObjects(left, right)
	if err == nil {
		t.Error("Adding string to array should fail")
	}
}

func TestBinaryAddObjectsUnsupported(t *testing.T) {
	left := &objects.Int{Value: 1}
	right := &objects.Int{Value: 2}

	_, err := binaryAddObjects(left, right)
	if err == nil {
		t.Error("binaryAddObjects with int should fail (int is handled by Value.Add)")
	}
}

// ============================================
// executeUnaryOp Tests
// ============================================

func TestExecuteUnaryOpNeg(t *testing.T) {
	val := NewInt(42)

	result, err := executeUnaryOp(compiler.OpRegNeg, val)
	if err != nil {
		t.Fatalf("executeUnaryOp(Neg) failed: %v", err)
	}

	got, _ := result.ToInt()
	if got != -42 {
		t.Errorf("-42 = %d, expected -42", got)
	}
}

func TestExecuteUnaryOpNegFloat(t *testing.T) {
	val := NewFloat(3.14)

	result, err := executeUnaryOp(compiler.OpRegNeg, val)
	if err != nil {
		t.Fatalf("executeUnaryOp(Neg) failed: %v", err)
	}

	got := result.GetFloat()
	if got != -3.14 {
		t.Errorf("-3.14 = %f, expected -3.14", got)
	}
}

func TestExecuteUnaryOpNot(t *testing.T) {
	tests := []struct {
		val      Value
		expected Value
	}{
		{ValueTrue, ValueFalse},
		{ValueFalse, ValueTrue},
		{NewInt(0), ValueTrue},
		{NewInt(1), ValueFalse},
		{ValueNull, ValueTrue},
	}

	for _, tt := range tests {
		result, err := executeUnaryOp(compiler.OpRegNot, tt.val)
		if err != nil {
			t.Fatalf("executeUnaryOp(Not) failed: %v", err)
		}

		if result != tt.expected {
			t.Errorf("!%v = %v, expected %v", tt.val, result, tt.expected)
		}
	}
}

func TestExecuteUnaryOpNegInvalid(t *testing.T) {
	// Try to negate a string (should fail)
	str := NewObject(&objects.String{Value: "hello"})

	_, err := executeUnaryOp(compiler.OpRegNeg, str)
	if err == nil {
		t.Error("Negating string should fail")
	}
}

func TestExecuteUnaryOpUnknown(t *testing.T) {
	val := NewInt(42)

	_, err := executeUnaryOp(compiler.Opcode(0xFF), val)
	if err == nil {
		t.Error("Unknown opcode should return error")
	}
}
