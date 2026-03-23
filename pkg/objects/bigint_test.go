// pkg/objects/bigint_test.go
package objects

import (
	"math/big"
	"testing"
)

func TestBigIntType(t *testing.T) {
	bi := NewBigIntFromInt64(42)
	if bi.Type() != BigIntType {
		t.Errorf("Expected BigIntType, got %s", bi.Type())
	}
	if bi.TypeTag() != TagBigInt {
		t.Errorf("Expected TagBigInt, got %d", bi.TypeTag())
	}
}

func TestBigIntInspect(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{-17, "-17"},
	}

	for _, tt := range tests {
		bi := NewBigIntFromInt64(tt.input)
		if bi.Inspect() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, bi.Inspect())
		}
	}
}

func TestBigIntToBool(t *testing.T) {
	// Zero should be false
	zero := NewBigIntFromInt64(0)
	if zero.ToBool() != FALSE {
		t.Error("Zero BigInt should be FALSE")
	}

	// Positive should be true
	pos := NewBigIntFromInt64(42)
	if pos.ToBool() != TRUE {
		t.Error("Positive BigInt should be TRUE")
	}

	// Negative should be true
	neg := NewBigIntFromInt64(-17)
	if neg.ToBool() != TRUE {
		t.Error("Negative BigInt should be TRUE")
	}
}

func TestBigIntFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"123", "123", false},
		{"-456", "-456", false},
		{"0x10", "16", false},   // hex
		{"0b101", "5", false},   // binary
		{"0o77", "63", false},   // octal
		{"invalid", "", true},   // error case
	}

	for _, tt := range tests {
		bi, err := NewBigIntFromString(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
			}
			if bi.Inspect() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, bi.Inspect())
			}
		}
	}
}

func TestBigIntArithmetic(t *testing.T) {
	a := NewBigIntFromInt64(10)
	b := NewBigIntFromInt64(3)

	// Add
	sum := a.Add(b)
	if sum.Inspect() != "13" {
		t.Errorf("Expected 13, got %s", sum.Inspect())
	}

	// Sub
	diff := a.Sub(b)
	if diff.Inspect() != "7" {
		t.Errorf("Expected 7, got %s", diff.Inspect())
	}

	// Mul
	prod := a.Mul(b)
	if prod.Inspect() != "30" {
		t.Errorf("Expected 30, got %s", prod.Inspect())
	}

	// Div
	quot := a.Div(b)
	if quot.Inspect() != "3" {
		t.Errorf("Expected 3, got %s", quot.Inspect())
	}

	// Mod
	rem := a.Mod(b)
	if rem.Inspect() != "1" {
		t.Errorf("Expected 1, got %s", rem.Inspect())
	}
}

func TestBigIntDivByZero(t *testing.T) {
	a := NewBigIntFromInt64(10)
	zero := NewBigIntFromInt64(0)

	result := a.Div(zero)
	if result != nil {
		t.Error("Division by zero should return nil")
	}

	result = a.Mod(zero)
	if result != nil {
		t.Error("Mod by zero should return nil")
	}
}

func TestBigIntNegAndAbs(t *testing.T) {
	a := NewBigIntFromInt64(42)

	// Neg
	neg := a.Neg()
	if neg.Inspect() != "-42" {
		t.Errorf("Expected -42, got %s", neg.Inspect())
	}

	// Abs of negative
	b := NewBigIntFromInt64(-17)
	abs := b.Abs()
	if abs.Inspect() != "17" {
		t.Errorf("Expected 17, got %s", abs.Inspect())
	}
}

func TestBigIntCmp(t *testing.T) {
	a := NewBigIntFromInt64(10)
	b := NewBigIntFromInt64(20)
	c := NewBigIntFromInt64(10)

	if a.Cmp(b) >= 0 {
		t.Error("10 should be less than 20")
	}
	if b.Cmp(a) <= 0 {
		t.Error("20 should be greater than 10")
	}
	if a.Cmp(c) != 0 {
		t.Error("10 should equal 10")
	}
}

func TestBigIntToInt64(t *testing.T) {
	small := NewBigIntFromInt64(42)
	val, ok := small.ToInt64()
	if !ok || val != 42 {
		t.Errorf("Expected 42, got %d, ok=%v", val, ok)
	}

	// Large value that overflows int64
	large, _ := NewBigIntFromString("9223372036854775808") // int64 max + 1
	_, ok = large.ToInt64()
	if ok {
		t.Error("Large value should not convert to int64")
	}
}

func TestBigIntToFloat64(t *testing.T) {
	bi := NewBigIntFromInt64(42)
	f := bi.ToFloat64()
	if f != 42.0 {
		t.Errorf("Expected 42.0, got %f", f)
	}
}

func TestBigIntToBigFloat(t *testing.T) {
	bi := NewBigIntFromInt64(42)
	bf := bi.ToBigFloat()
	if bf.Type() != BigFloatType {
		t.Errorf("Expected BigFloatType, got %s", bf.Type())
	}
}

func TestBigIntBitLen(t *testing.T) {
	tests := []struct {
		input    int64
		expected int
	}{
		{0, 0},
		{1, 1},
		{7, 3},
		{8, 4},
		{255, 8},
	}

	for _, tt := range tests {
		bi := NewBigIntFromInt64(tt.input)
		if bi.BitLen() != tt.expected {
			t.Errorf("Expected bit length %d, got %d for %d", tt.expected, bi.BitLen(), tt.input)
		}
	}
}

func TestBigIntSign(t *testing.T) {
	pos := NewBigIntFromInt64(42)
	if pos.Sign() != 1 {
		t.Error("Positive BigInt should have sign 1")
	}
	if !pos.IsPositive() {
		t.Error("IsPositive should be true")
	}

	zero := NewBigIntFromInt64(0)
	if zero.Sign() != 0 {
		t.Error("Zero BigInt should have sign 0")
	}
	if !zero.IsZero() {
		t.Error("IsZero should be true")
	}

	neg := NewBigIntFromInt64(-17)
	if neg.Sign() != -1 {
		t.Error("Negative BigInt should have sign -1")
	}
	if !neg.IsNegative() {
		t.Error("IsNegative should be true")
	}
}

func TestBigIntClone(t *testing.T) {
	a := NewBigIntFromInt64(42)
	b := a.Clone()

	if a.Inspect() != b.Inspect() {
		t.Error("Clone should have same value")
	}

	// Modify b, a should be unchanged
	b = b.Add(NewBigIntFromInt64(1))
	if a.Inspect() != "42" {
		t.Error("Original should be unchanged")
	}
}

func TestBigIntHashKey(t *testing.T) {
	a := NewBigIntFromInt64(42)
	b := NewBigIntFromInt64(42)
	c := NewBigIntFromInt64(43)

	// Same value should have same hash key
	if a.HashKey() != b.HashKey() {
		t.Error("Same BigInt values should have same hash key")
	}

	// Different values should have different hash keys
	if a.HashKey() == c.HashKey() {
		t.Error("Different BigInt values should have different hash keys")
	}
}

func TestBigIntLargeValues(t *testing.T) {
	// Test with a very large number
	large, _ := NewBigIntFromString("1234567890123456789012345678901234567890")
	if large.Inspect() != "1234567890123456789012345678901234567890" {
		t.Errorf("Large value not preserved: %s", large.Inspect())
	}

	// Arithmetic with large values
	a, _ := NewBigIntFromString("999999999999999999999999999999")
	b, _ := NewBigIntFromString("1")
	sum := a.Add(b)
	if sum.Inspect() != "1000000000000000000000000000000" {
		t.Errorf("Large addition failed: %s", sum.Inspect())
	}
}

func TestNewBigIntFromBig(t *testing.T) {
	val := big.NewInt(12345)
	bi := NewBigInt(val)
	if bi.Inspect() != "12345" {
		t.Errorf("Expected 12345, got %s", bi.Inspect())
	}

	// Nil should return zero
	bi = NewBigInt(nil)
	if bi.Inspect() != "0" {
		t.Errorf("Expected 0 for nil, got %s", bi.Inspect())
	}
}
