// pkg/vm/value_test.go
// Tests for NaN-boxed Value type
package vm

import (
	"math"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Integer Tests
// ============================================

func TestValueInt(t *testing.T) {
	tests := []int64{
		0, 1, -1, 42, -42,
		100, -100, 1000, -1000,
		1 << 20, -(1 << 20),
		1 << 30, -(1 << 30),
		// 47-bit range (with sign extension)
		// Note: we use 47 bits for the magnitude, so max is 2^47 - 1
		1<<46, -(1 << 46),
	}

	for _, n := range tests {
		v := NewInt(n)
		if !v.IsInt() {
			t.Errorf("NewInt(%d).IsInt() = false", n)
		}
		if got := v.GetInt(); got != n {
			t.Errorf("NewInt(%d).GetInt() = %d", n, got)
		}
	}
}

func TestValueIntRange(t *testing.T) {
	// Test the boundaries of 48-bit signed integer storage
	// Max positive: 2^47 - 1
	// Min negative: -2^47
	maxVal := int64(1<<47 - 1)
	minVal := int64(-(1 << 47))

	v := NewInt(maxVal)
	if v.GetInt() != maxVal {
		t.Errorf("max int: got %d, want %d", v.GetInt(), maxVal)
	}

	v = NewInt(minVal)
	if v.GetInt() != minVal {
		t.Errorf("min int: got %d, want %d", v.GetInt(), minVal)
	}
}

// ============================================
// Float Tests
// ============================================

func TestValueFloat(t *testing.T) {
	tests := []float64{
		0.0, 1.0, -1.0,
		3.14159, -3.14159,
		1e10, 1e-10,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
		math.Inf(1), math.Inf(-1),
	}

	for _, f := range tests {
		v := NewFloat(f)
		if !v.IsFloat() {
			t.Errorf("NewFloat(%v).IsFloat() = false", f)
		}
		if got := v.GetFloat(); got != f {
			t.Errorf("NewFloat(%v).GetFloat() = %v", f, got)
		}
	}
}

// ============================================
// Boolean Tests
// ============================================

func TestValueBool(t *testing.T) {
	vt := NewBool(true)
	vf := NewBool(false)

	if !vt.IsBool() {
		t.Error("NewBool(true).IsBool() = false")
	}
	if !vf.IsBool() {
		t.Error("NewBool(false).IsBool() = false")
	}
	if !vt.GetBool() {
		t.Error("NewBool(true).GetBool() = false")
	}
	if vf.GetBool() {
		t.Error("NewBool(false).GetBool() = true")
	}
	if ValueTrue != vt {
		t.Error("ValueTrue != NewBool(true)")
	}
	if ValueFalse != vf {
		t.Error("ValueFalse != NewBool(false)")
	}
}

// ============================================
// Null Tests
// ============================================

func TestValueNull(t *testing.T) {
	v := ValueNull
	if !v.IsNull() {
		t.Error("ValueNull.IsNull() = false")
	}
	if v.IsTruthy() {
		t.Error("ValueNull.IsTruthy() = true")
	}
}

// ============================================
// Type Detection Tests
// ============================================

func TestValueTypeDetection(t *testing.T) {
	tests := []struct {
		v      Value
		isInt  bool
		isFloat bool
		isBool bool
		isNull bool
	}{
		{NewInt(42), true, false, false, false},
		{NewFloat(3.14), false, true, false, false},
		{NewBool(true), false, false, true, false},
		{ValueNull, false, false, false, true},
	}

	for _, tt := range tests {
		if tt.v.IsInt() != tt.isInt {
			t.Errorf("%v.IsInt() = %v, want %v", tt.v, tt.v.IsInt(), tt.isInt)
		}
		if tt.v.IsFloat() != tt.isFloat {
			t.Errorf("%v.IsFloat() = %v, want %v", tt.v, tt.v.IsFloat(), tt.isFloat)
		}
		if tt.v.IsBool() != tt.isBool {
			t.Errorf("%v.IsBool() = %v, want %v", tt.v, tt.v.IsBool(), tt.isBool)
		}
		if tt.v.IsNull() != tt.isNull {
			t.Errorf("%v.IsNull() = %v, want %v", tt.v, tt.v.IsNull(), tt.isNull)
		}
	}
}

// ============================================
// Arithmetic Tests
// ============================================

func TestValueAdd(t *testing.T) {
	tests := []struct {
		a, b Value
		want int64
	}{
		{NewInt(1), NewInt(2), 3},
		{NewInt(10), NewInt(5), 15},
		{NewInt(-5), NewInt(5), 0},
		{NewInt(100), NewInt(-50), 50},
	}

	for _, tt := range tests {
		result, ok := tt.a.Add(tt.b)
		if !ok {
			t.Errorf("%v.Add(%v) failed", tt.a, tt.b)
			continue
		}
		if !result.IsInt() {
			t.Errorf("%v.Add(%v) = %v, not int", tt.a, tt.b, result)
			continue
		}
		if got := result.GetInt(); got != tt.want {
			t.Errorf("%v.Add(%v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueSub(t *testing.T) {
	tests := []struct {
		a, b Value
		want int64
	}{
		{NewInt(5), NewInt(3), 2},
		{NewInt(10), NewInt(20), -10},
		{NewInt(-5), NewInt(-5), 0},
	}

	for _, tt := range tests {
		result, ok := tt.a.Sub(tt.b)
		if !ok {
			t.Errorf("%v.Sub(%v) failed", tt.a, tt.b)
			continue
		}
		if got := result.GetInt(); got != tt.want {
			t.Errorf("%v.Sub(%v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueMul(t *testing.T) {
	tests := []struct {
		a, b Value
		want int64
	}{
		{NewInt(3), NewInt(4), 12},
		{NewInt(-2), NewInt(5), -10},
		{NewInt(0), NewInt(100), 0},
	}

	for _, tt := range tests {
		result, ok := tt.a.Mul(tt.b)
		if !ok {
			t.Errorf("%v.Mul(%v) failed", tt.a, tt.b)
			continue
		}
		if got := result.GetInt(); got != tt.want {
			t.Errorf("%v.Mul(%v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueDiv(t *testing.T) {
	tests := []struct {
		a, b Value
		want float64
	}{
		{NewInt(10), NewInt(2), 5.0},
		{NewInt(7), NewInt(2), 3.5},
		{NewFloat(10.0), NewFloat(4.0), 2.5},
	}

	for _, tt := range tests {
		result, ok := tt.a.Div(tt.b)
		if !ok {
			t.Errorf("%v.Div(%v) failed", tt.a, tt.b)
			continue
		}
		if got := result.GetFloat(); got != tt.want {
			t.Errorf("%v.Div(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueDivByZero(t *testing.T) {
	_, ok := NewInt(10).Div(NewInt(0))
	if ok {
		t.Error("division by zero should fail")
	}
}

func TestValueMod(t *testing.T) {
	tests := []struct {
		a, b Value
		want int64
	}{
		{NewInt(10), NewInt(3), 1},
		{NewInt(15), NewInt(5), 0},
		{NewInt(17), NewInt(4), 1},
	}

	for _, tt := range tests {
		result, ok := tt.a.Mod(tt.b)
		if !ok {
			t.Errorf("%v.Mod(%v) failed", tt.a, tt.b)
			continue
		}
		if got := result.GetInt(); got != tt.want {
			t.Errorf("%v.Mod(%v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueNeg(t *testing.T) {
	tests := []struct {
		v    Value
		want int64
	}{
		{NewInt(5), -5},
		{NewInt(-10), 10},
		{NewInt(0), 0},
	}

	for _, tt := range tests {
		result, ok := tt.v.Neg()
		if !ok {
			t.Errorf("%v.Neg() failed", tt.v)
			continue
		}
		if got := result.GetInt(); got != tt.want {
			t.Errorf("%v.Neg() = %d, want %d", tt.v, got, tt.want)
		}
	}
}

// ============================================
// Comparison Tests
// ============================================

func TestValueLess(t *testing.T) {
	tests := []struct {
		a, b Value
		want bool
	}{
		{NewInt(1), NewInt(2), true},
		{NewInt(2), NewInt(1), false},
		{NewInt(5), NewInt(5), false},
		{NewInt(-5), NewInt(5), true},
	}

	for _, tt := range tests {
		got, ok := tt.a.Less(tt.b)
		if !ok {
			t.Errorf("%v.Less(%v) failed", tt.a, tt.b)
			continue
		}
		if got != tt.want {
			t.Errorf("%v.Less(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueGreater(t *testing.T) {
	tests := []struct {
		a, b Value
		want bool
	}{
		{NewInt(2), NewInt(1), true},
		{NewInt(1), NewInt(2), false},
		{NewInt(5), NewInt(5), false},
	}

	for _, tt := range tests {
		got, ok := tt.a.Greater(tt.b)
		if !ok {
			t.Errorf("%v.Greater(%v) failed", tt.a, tt.b)
			continue
		}
		if got != tt.want {
			t.Errorf("%v.Greater(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValueEqual(t *testing.T) {
	tests := []struct {
		a, b Value
		want bool
	}{
		{NewInt(5), NewInt(5), true},
		{NewInt(5), NewInt(6), false},
		{ValueTrue, ValueTrue, true},
		{ValueFalse, ValueFalse, true},
		{ValueTrue, ValueFalse, false},
		{ValueNull, ValueNull, true},
	}

	for _, tt := range tests {
		got, ok := tt.a.Equal(tt.b)
		if !ok {
			t.Errorf("%v.Equal(%v) failed", tt.a, tt.b)
			continue
		}
		if got != tt.want {
			t.Errorf("%v.Equal(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

// ============================================
// ToObject Tests
// ============================================

func TestValueToObject(t *testing.T) {
	// Int
	v := NewInt(42)
	obj := v.ToObject()
	if obj.Type() != objects.IntType {
		t.Errorf("ToInt: wrong type %v", obj.Type())
	}
	if obj.(*objects.Int).Value != 42 {
		t.Errorf("ToInt: wrong value %v", obj.(*objects.Int).Value)
	}

	// Float
	v = NewFloat(3.14)
	obj = v.ToObject()
	if obj.Type() != objects.FloatType {
		t.Errorf("ToFloat: wrong type %v", obj.Type())
	}

	// Bool
	v = NewBool(true)
	obj = v.ToObject()
	if obj.Type() != objects.BoolType {
		t.Errorf("ToBool: wrong type %v", obj.Type())
	}

	// Null
	v = ValueNull
	obj = v.ToObject()
	if obj.Type() != objects.NullType {
		t.Errorf("ToNull: wrong type %v", obj.Type())
	}
}

func TestNewValueFromObject(t *testing.T) {
	// Int object
	intObj := objects.NewInt(100)
	v := NewValue(intObj)
	if !v.IsInt() {
		t.Error("NewValue(Int).IsInt() = false")
	}
	if v.GetInt() != 100 {
		t.Errorf("NewValue(Int).GetInt() = %d, want 100", v.GetInt())
	}

	// Float object
	floatObj := &objects.Float{Value: 2.5}
	v = NewValue(floatObj)
	if !v.IsFloat() {
		t.Error("NewValue(Float).IsFloat() = false")
	}

	// Bool object
	v = NewValue(objects.TRUE)
	if !v.IsBool() {
		t.Error("NewValue(TRUE).IsBool() = false")
	}
	if !v.GetBool() {
		t.Error("NewValue(TRUE).GetBool() = false")
	}

	// Null
	v = NewValue(objects.NULL)
	if !v.IsNull() {
		t.Error("NewValue(NULL).IsNull() = false")
	}
}

// ============================================
// IsTruthy Tests
// ============================================

func TestValueIsTruthy(t *testing.T) {
	tests := []struct {
		v    Value
		want bool
	}{
		{NewInt(0), false},
		{NewInt(1), true},
		{NewInt(-1), true},
		{ValueFalse, false},
		{ValueTrue, true},
		{ValueNull, false},
		{NewFloat(0.0), false},
		{NewFloat(0.1), true},
	}

	for _, tt := range tests {
		if got := tt.v.IsTruthy(); got != tt.want {
			t.Errorf("%v.IsTruthy() = %v, want %v", tt.v, got, tt.want)
		}
	}
}

// ============================================
// ValueStack Tests
// ============================================

func TestValueStackPushPop(t *testing.T) {
	s := NewValueStack()

	// Push some values
	s.MustPush(NewInt(1))
	s.MustPush(NewInt(2))
	s.MustPush(NewInt(3))

	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}

	// Pop in reverse order
	if v := s.Pop(); v.GetInt() != 3 {
		t.Errorf("Pop() = %d, want 3", v.GetInt())
	}
	if v := s.Pop(); v.GetInt() != 2 {
		t.Errorf("Pop() = %d, want 2", v.GetInt())
	}
	if v := s.Pop(); v.GetInt() != 1 {
		t.Errorf("Pop() = %d, want 1", v.GetInt())
	}

	if s.Len() != 0 {
		t.Errorf("Len() after pops = %d, want 0", s.Len())
	}
}

func TestValueStackTop(t *testing.T) {
	s := NewValueStack()
	s.MustPush(NewInt(1))
	s.MustPush(NewInt(2))

	if v := s.Top(); v.GetInt() != 2 {
		t.Errorf("Top() = %d, want 2", v.GetInt())
	}

	s.Pop()
	if v := s.Top(); v.GetInt() != 1 {
		t.Errorf("Top() after pop = %d, want 1", v.GetInt())
	}
}

func TestValueStackDup(t *testing.T) {
	s := NewValueStack()
	s.MustPush(NewInt(42))
	s.Dup()

	if s.Len() != 2 {
		t.Errorf("Len() after Dup = %d, want 2", s.Len())
	}

	if v := s.Top(); v.GetInt() != 42 {
		t.Errorf("Top() after Dup = %d, want 42", v.GetInt())
	}
}

func TestValueStackSwap(t *testing.T) {
	s := NewValueStack()
	s.MustPush(NewInt(1))
	s.MustPush(NewInt(2))
	s.Swap()

	if v := s.Pop(); v.GetInt() != 1 {
		t.Errorf("After Swap, first pop = %d, want 1", v.GetInt())
	}
	if v := s.Pop(); v.GetInt() != 2 {
		t.Errorf("After Swap, second pop = %d, want 2", v.GetInt())
	}
}

// ============================================
// Benchmarks
// ============================================

func BenchmarkValueNewInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewInt(int64(i))
	}
}

func BenchmarkValueAdd(b *testing.B) {
	v1 := NewInt(100)
	v2 := NewInt(200)
	for i := 0; i < b.N; i++ {
		_, _ = v1.Add(v2)
	}
}

func BenchmarkValueMul(b *testing.B) {
	v1 := NewInt(100)
	v2 := NewInt(200)
	for i := 0; i < b.N; i++ {
		_, _ = v1.Mul(v2)
	}
}

func BenchmarkValueLess(b *testing.B) {
	v1 := NewInt(100)
	v2 := NewInt(200)
	for i := 0; i < b.N; i++ {
		_, _ = v1.Less(v2)
	}
}

func BenchmarkValueStackPushPop(b *testing.B) {
	s := NewValueStack()
	v := NewInt(42)
	for i := 0; i < b.N; i++ {
		s.MustPush(v)
		s.Pop()
	}
}

// Compare with object-based operations
func BenchmarkObjectAdd(b *testing.B) {
	o1 := objects.NewInt(100)
	o2 := objects.NewInt(200)
	for i := 0; i < b.N; i++ {
		_ = objects.NewInt(o1.Value + o2.Value)
	}
}

func BenchmarkObjectNewInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = objects.NewInt(int64(i))
	}
}

// More realistic benchmarks: operations on non-cached integers
func BenchmarkValueAddNonCached(b *testing.B) {
	// Integers > 10000 are not cached
	v1 := NewInt(100001)
	v2 := NewInt(100002)
	for i := 0; i < b.N; i++ {
		_, _ = v1.Add(v2)
	}
}

func BenchmarkObjectAddNonCached(b *testing.B) {
	// Create non-cached integers
	o1 := &objects.Int{Value: 100001}
	o2 := &objects.Int{Value: 100002}
	for i := 0; i < b.N; i++ {
		_ = &objects.Int{Value: o1.Value + o2.Value}
	}
}

// Stack operations comparison
func BenchmarkValueStackOps(b *testing.B) {
	s := NewValueStack()
	v := NewInt(42)
	for i := 0; i < b.N; i++ {
		s.MustPush(v)
		s.MustPush(v)
		_, _ = v.Add(v)
		s.Pop()
		s.Pop()
	}
}

func BenchmarkObjectStackOps(b *testing.B) {
	s := NewStack()
	var o objects.Object = objects.NewInt(42)
	for i := 0; i < b.N; i++ {
		s.Push(o)
		s.Push(o)
		// Type assertion and operation
		oi := o.(*objects.Int)
		_ = objects.NewInt(oi.Value + oi.Value)
		s.Pop()
		s.Pop()
	}
}

// Mixed operations (realistic workload)
func BenchmarkValueMixedOps(b *testing.B) {
	s := NewValueStack()
	for i := 0; i < b.N; i++ {
		a := NewInt(int64(i % 1000))
		bv := NewInt(int64((i + 1) % 1000))
		s.MustPush(a)
		s.MustPush(bv)
		sum, _ := a.Add(bv)
		s.MustPush(sum)
		s.Pop()
		s.Pop()
		s.Pop()
	}
}

func BenchmarkObjectMixedOps(b *testing.B) {
	s := NewStack()
	for i := 0; i < b.N; i++ {
		var a objects.Object = objects.NewInt(int64(i % 1000))
		var bv objects.Object = objects.NewInt(int64((i + 1) % 1000))
		s.Push(a)
		s.Push(bv)
		ai := a.(*objects.Int)
		bi := bv.(*objects.Int)
		sum := objects.NewInt(ai.Value + bi.Value)
		s.Push(sum)
		s.Pop()
		s.Pop()
		s.Pop()
	}
}
