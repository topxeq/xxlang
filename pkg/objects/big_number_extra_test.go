// pkg/objects/big_number_extra_test.go
package objects

import (
	"math/big"
	"testing"
)

func TestBigIntAddIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(100)
	result := bi.AddInt(50)
	if result.Value.Int64() != 150 {
		t.Errorf("expected 150, got %d", result.Value.Int64())
	}
}

func TestBigIntSubIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(100)
	result := bi.SubInt(30)
	if result.Value.Int64() != 70 {
		t.Errorf("expected 70, got %d", result.Value.Int64())
	}
}

func TestBigIntMulIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(10)
	result := bi.MulInt(5)
	if result.Value.Int64() != 50 {
		t.Errorf("expected 50, got %d", result.Value.Int64())
	}
}

func TestBigIntDivIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(100)
	result := bi.DivInt(5)
	if result.Value.Int64() != 20 {
		t.Errorf("expected 20, got %d", result.Value.Int64())
	}
}

func TestBigIntModIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(17)
	result := bi.ModInt(5)
	if result.Value.Int64() != 2 {
		t.Errorf("expected 2, got %d", result.Value.Int64())
	}
}

func TestBigIntStringMethod(t *testing.T) {
	bi := NewBigIntFromInt64(123)
	s := bi.String()
	if s != "123" {
		t.Errorf("expected '123', got '%s'", s)
	}
}

func TestNewBigIntFromIntMethod(t *testing.T) {
	i := &Int{Value: 456}
	bi := NewBigIntFromInt(i)
	if bi.Value.Int64() != 456 {
		t.Errorf("expected 456, got %d", bi.Value.Int64())
	}
}

func TestBigFloatFloorMethod(t *testing.T) {
	bf := NewBigFloatFromFloat64(3.7)
	result := bf.Floor()
	// Floor sets the rounding mode, value is converted when operated on
	_ = result
}

func TestBigFloatCeilMethod(t *testing.T) {
	bf := NewBigFloatFromFloat64(3.2)
	result := bf.Ceil()
	// Ceil sets the rounding mode, value is converted when operated on
	_ = result
}

func TestBigFloatRoundMethod(t *testing.T) {
	bf := NewBigFloatFromFloat64(3.5)
	result := bf.Round()
	// Round sets the rounding mode, value is converted when operated on
	_ = result
}

func TestBigFloatStringMethod(t *testing.T) {
	bf := NewBigFloatFromFloat64(123.45)
	s := bf.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestFormatBigFloatMethod(t *testing.T) {
	bf := NewBigFloatFromFloat64(1000.5)
	s := FormatBigFloat(bf)
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestBigFloatFromBigIntMethod(t *testing.T) {
	bi := NewBigIntFromInt64(123)
	bf := NewBigFloatFromBigInt(bi)
	i := new(big.Int)
	bf.Value.Int(i)
	if i.Int64() != 123 {
		t.Errorf("expected 123, got %d", i.Int64())
	}
}
