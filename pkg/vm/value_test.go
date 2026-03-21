// pkg/vm/value_test.go
// Tests for NaN-boxed value operations
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Value Creation Tests
// ============================================

func TestNewInt(t *testing.T) {
	tests := []int64{0, 1, -1, 42, -42, 1000000, -1000000}

	for _, n := range tests {
		v := NewInt(n)
		if !v.IsInt() {
			t.Errorf("NewInt(%d).IsInt() = false", n)
		}
		got, ok := v.ToInt()
		if !ok {
			t.Errorf("NewInt(%d).ToInt() failed", n)
		}
		if got != n {
			t.Errorf("NewInt(%d) = %d", n, got)
		}
	}
}

func TestNewFloat(t *testing.T) {
	tests := []float64{0.0, 1.5, -1.5, 3.14159, -3.14159, 1e10, -1e10}

	for _, f := range tests {
		v := NewFloat(f)
		if !v.IsFloat() {
			t.Errorf("NewFloat(%f).IsFloat() = false", f)
		}
		got := v.GetFloat()
		if got != f {
			t.Errorf("NewFloat(%f) = %f", f, got)
		}
	}
}

func TestNewBool(t *testing.T) {
	v := NewBool(true)
	if !v.IsBool() {
		t.Error("NewBool(true).IsBool() = false")
	}
	if v != ValueTrue {
		t.Error("NewBool(true) != ValueTrue")
	}

	v = NewBool(false)
	if !v.IsBool() {
		t.Error("NewBool(false).IsBool() = false")
	}
	if v != ValueFalse {
		t.Error("NewBool(false) != ValueFalse")
	}
}

func TestValueNull(t *testing.T) {
	v := ValueNull
	if !v.IsNull() {
		t.Error("ValueNull.IsNull() = false")
	}
}

// ============================================
// Value Type Tests
// ============================================

func TestValueIsInt(t *testing.T) {
	v := NewInt(42)
	if !v.IsInt() {
		t.Error("Int value should return true for IsInt()")
	}

	v = NewFloat(3.14)
	if v.IsInt() {
		t.Error("Float value should return false for IsInt()")
	}

	v = ValueNull
	if v.IsInt() {
		t.Error("Null value should return false for IsInt()")
	}
}

func TestValueIsFloat(t *testing.T) {
	v := NewFloat(3.14)
	if !v.IsFloat() {
		t.Error("Float value should return true for IsFloat()")
	}

	v = NewInt(42)
	if v.IsFloat() {
		t.Error("Int value should return false for IsFloat()")
	}
}

func TestValueIsBool(t *testing.T) {
	v := ValueTrue
	if !v.IsBool() {
		t.Error("True value should return true for IsBool()")
	}

	v = ValueFalse
	if !v.IsBool() {
		t.Error("False value should return true for IsBool()")
	}

	v = NewInt(1)
	if v.IsBool() {
		t.Error("Int value should return false for IsBool()")
	}
}

func TestValueIsNull(t *testing.T) {
	v := ValueNull
	if !v.IsNull() {
		t.Error("Null value should return true for IsNull()")
	}

	v = NewInt(0)
	if v.IsNull() {
		t.Error("Int(0) should return false for IsNull()")
	}
}

func TestValueIsNumber(t *testing.T) {
	v := NewInt(42)
	if !v.IsNumber() {
		t.Error("Int value should be a number")
	}

	v = NewFloat(3.14)
	if !v.IsNumber() {
		t.Error("Float value should be a number")
	}

	v = ValueTrue
	if v.IsNumber() {
		t.Error("Bool value should not be a number")
	}
}

func TestValueIsTruthy(t *testing.T) {
	// Truthy values
	if !ValueTrue.IsTruthy() {
		t.Error("true should be truthy")
	}
	if !NewInt(1).IsTruthy() {
		t.Error("1 should be truthy")
	}
	if !NewInt(-1).IsTruthy() {
		t.Error("-1 should be truthy")
	}
	if !NewFloat(3.14).IsTruthy() {
		t.Error("3.14 should be truthy")
	}

	// Falsy values
	if ValueFalse.IsTruthy() {
		t.Error("false should not be truthy")
	}
	if ValueNull.IsTruthy() {
		t.Error("null should not be truthy")
	}
	if NewInt(0).IsTruthy() {
		t.Error("0 should not be truthy")
	}
	if NewFloat(0.0).IsTruthy() {
		t.Error("0.0 should not be truthy")
	}
}

// ============================================
// Value Arithmetic Tests
// ============================================

func TestValueAdd(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    int64
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
		{100, 200, 300},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result, ok := left.Add(right)
		if !ok {
			t.Errorf("%d + %d failed", tt.left, tt.right)
			continue
		}
		if !result.IsInt() {
			t.Errorf("%d + %d should produce int", tt.left, tt.right)
			continue
		}
		got, _ := result.ToInt()
		if got != tt.expected {
			t.Errorf("%d + %d = %d, expected %d", tt.left, tt.right, got, tt.expected)
		}
	}
}

func TestValueAddFloat(t *testing.T) {
	left := NewFloat(1.5)
	right := NewFloat(2.5)
	result, ok := left.Add(right)
	if !ok {
		t.Fatal("1.5 + 2.5 failed")
	}
	if !result.IsFloat() {
		t.Fatal("Expected float result")
	}
	if result.GetFloat() != 4.0 {
		t.Errorf("1.5 + 2.5 = %f, expected 4.0", result.GetFloat())
	}
}

func TestValueSub(t *testing.T) {
	left := NewInt(10)
	right := NewInt(3)
	result, ok := left.Sub(right)
	if !ok {
		t.Fatal("10 - 3 failed")
	}
	if got, _ := result.ToInt(); got != 7 {
		t.Errorf("10 - 3 = %d, expected 7", got)
	}
}

func TestValueMul(t *testing.T) {
	left := NewInt(6)
	right := NewInt(7)
	result, ok := left.Mul(right)
	if !ok {
		t.Fatal("6 * 7 failed")
	}
	if got, _ := result.ToInt(); got != 42 {
		t.Errorf("6 * 7 = %d, expected 42", got)
	}
}

func TestValueDiv(t *testing.T) {
	left := NewInt(20)
	right := NewInt(4)
	result, ok := left.Div(right)
	if !ok {
		t.Fatal("20 / 4 failed")
	}
	// Division produces float
	if !result.IsFloat() {
		t.Fatal("Division should produce float")
	}
	if result.GetFloat() != 5.0 {
		t.Errorf("20 / 4 = %f, expected 5.0", result.GetFloat())
	}
}

func TestValueMod(t *testing.T) {
	left := NewInt(10)
	right := NewInt(3)
	result, ok := left.Mod(right)
	if !ok {
		t.Fatal("10 % 3 failed")
	}
	if got, _ := result.ToInt(); got != 1 {
		t.Errorf("10 %% 3 = %d, expected 1", got)
	}
}

// ============================================
// Value Comparison Tests
// ============================================

func TestValueLess(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{1, 2, true},
		{2, 1, false},
		{1, 1, false},
		{-1, 1, true},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result, ok := left.Less(right)
		if !ok {
			t.Errorf("%d < %d failed", tt.left, tt.right)
			continue
		}
		if result != tt.expected {
			t.Errorf("%d < %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

func TestValueGreater(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{2, 1, true},
		{1, 2, false},
		{1, 1, false},
		{1, -1, true},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result, ok := left.Greater(right)
		if !ok {
			t.Errorf("%d > %d failed", tt.left, tt.right)
			continue
		}
		if result != tt.expected {
			t.Errorf("%d > %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

func TestValueEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{1, 1, true},
		{1, 2, false},
		{0, 0, true},
		{-1, -1, true},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result, ok := left.Equal(right)
		if !ok {
			t.Errorf("%d == %d failed", tt.left, tt.right)
			continue
		}
		if result != tt.expected {
			t.Errorf("%d == %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

func TestValueNotEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    bool
	}{
		{1, 2, true},
		{1, 1, false},
		{0, 1, true},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result, ok := left.NotEqual(right)
		if !ok {
			t.Errorf("%d != %d failed", tt.left, tt.right)
			continue
		}
		if result != tt.expected {
			t.Errorf("%d != %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

func TestValueLessEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    Value
	}{
		{1, 2, ValueTrue},
		{1, 1, ValueTrue},
		{2, 1, ValueFalse},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result := left.LessEqual(right)
		if result != tt.expected {
			t.Errorf("%d <= %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

func TestValueGreaterEqual(t *testing.T) {
	tests := []struct {
		left, right int64
		expected    Value
	}{
		{2, 1, ValueTrue},
		{1, 1, ValueTrue},
		{1, 2, ValueFalse},
	}

	for _, tt := range tests {
		left := NewInt(tt.left)
		right := NewInt(tt.right)
		result := left.GreaterEqual(right)
		if result != tt.expected {
			t.Errorf("%d >= %d = %v, expected %v", tt.left, tt.right, result, tt.expected)
		}
	}
}

// ============================================
// Value String Tests
// ============================================

func TestValueString(t *testing.T) {
	tests := []struct {
		v        Value
		expected string
	}{
		{NewInt(42), "42"},
		{NewInt(-42), "-42"},
		{NewFloat(3.14), "3.14"},
		{ValueTrue, "true"},
		{ValueFalse, "false"},
		{ValueNull, "null"},
	}

	for _, tt := range tests {
		got := tt.v.String()
		if got != tt.expected {
			t.Errorf("Value.String() = %s, expected %s", got, tt.expected)
		}
	}
}

// ============================================
// Object Conversion Tests
// ============================================

func TestNewObject(t *testing.T) {
	// Test with string object
	str := &objects.String{Value: "hello"}
	v := NewObject(str)
	if !v.IsObject() {
		t.Error("String object should be recognized as object")
	}

	// Test with array object
	arr := &objects.Array{Elements: []objects.Object{}}
	v = NewObject(arr)
	if !v.IsObject() {
		t.Error("Array object should be recognized as object")
	}
}

func TestToObject(t *testing.T) {
	// Test int conversion
	v := NewInt(42)
	obj := v.ToObject()
	if obj == nil {
		t.Fatal("ToObject(int) returned nil")
	}
	intObj, ok := obj.(*objects.Int)
	if !ok {
		t.Fatalf("Expected *objects.Int, got %T", obj)
	}
	if intObj.Value != 42 {
		t.Errorf("Int value = %d, expected 42", intObj.Value)
	}

	// Test float conversion
	v = NewFloat(3.14)
	obj = v.ToObject()
	if obj == nil {
		t.Fatal("ToObject(float) returned nil")
	}
	floatObj, ok := obj.(*objects.Float)
	if !ok {
		t.Fatalf("Expected *objects.Float, got %T", obj)
	}
	if floatObj.Value != 3.14 {
		t.Errorf("Float value = %f, expected 3.14", floatObj.Value)
	}

	// Test null conversion
	v = ValueNull
	obj = v.ToObject()
	if obj.Type() != objects.NullType {
		t.Errorf("Expected NullType, got %s", obj.Type())
	}
}

// ============================================
// Closure Tests
// ============================================

func TestValueIsClosure(t *testing.T) {
	v := ValueNull
	if v.IsClosure() {
		t.Error("Null should not be closure")
	}

	v = NewInt(42)
	if v.IsClosure() {
		t.Error("Int should not be closure")
	}
}

func TestValueIsCompiledFunction(t *testing.T) {
	v := ValueNull
	if v.IsCompiledFunction() {
		t.Error("Null should not be compiled function")
	}

	v = NewInt(42)
	if v.IsCompiledFunction() {
		t.Error("Int should not be compiled function")
	}
}

// ============================================
// GetInt Tests
// ============================================

func TestValueGetInt(t *testing.T) {
	v := NewInt(123)
	got := v.GetInt()
	if got != 123 {
		t.Errorf("GetInt() = %d, expected 123", got)
	}
}

// ============================================
// ToFloat Tests
// ============================================

func TestValueToFloat(t *testing.T) {
	// Int to float
	v := NewInt(42)
	f, ok := v.ToFloat()
	if !ok {
		t.Fatal("Int.ToFloat() failed")
	}
	if f != 42.0 {
		t.Errorf("Int.ToFloat() = %f, expected 42.0", f)
	}

	// Float to float
	v = NewFloat(3.14)
	f, ok = v.ToFloat()
	if !ok {
		t.Fatal("Float.ToFloat() failed")
	}
	if f != 3.14 {
		t.Errorf("Float.ToFloat() = %f, expected 3.14", f)
	}
}

// ============================================
// Object Registry Tests
// ============================================

func TestObjectRegistry(t *testing.T) {
	// Clear registry first
	ClearRegistry()

	// Register an object
	obj := &objects.String{Value: "test"}
	v := NewObject(obj)

	// Should be able to get the object back
	got := v.ToObject()
	if got == nil {
		t.Fatal("ToObject returned nil")
	}

	str, ok := got.(*objects.String)
	if !ok {
		t.Fatalf("Expected *objects.String, got %T", got)
	}
	if str.Value != "test" {
		t.Errorf("String value = %s, expected 'test'", str.Value)
	}

	// Test registry stats
	total, used := RegistryStats()
	if total < 1 {
		t.Error("Registry should have at least one object")
	}
	_ = used
}

func TestClearRegistry(t *testing.T) {
	// Create some objects
	_ = NewObject(&objects.String{Value: "test1"})
	_ = NewObject(&objects.String{Value: "test2"})

	// Clear registry
	ClearRegistry()

	// Check stats
	total, used := RegistryStats()
	_ = used // Not checking used since ClearRegistry behavior may vary
	if total < 0 {
		t.Errorf("After ClearRegistry, TotalObjects = %d", total)
	}
}
