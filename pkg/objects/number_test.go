// pkg/objects/number_test.go
package objects

import "testing"

func TestIntInspect(t *testing.T) {
	i := &Int{Value: 42}
	if got := i.Inspect(); got != "42" {
		t.Errorf("Int.Inspect() = %s, want 42", got)
	}
}

func TestIntType(t *testing.T) {
	i := &Int{Value: 42}
	if got := i.Type(); got != IntType {
		t.Errorf("Int.Type() = %s, want INT", got)
	}
}

func TestIntToBool(t *testing.T) {
	zero := &Int{Value: 0}
	if zero.ToBool() != FALSE {
		t.Error("Int(0).ToBool() should be FALSE")
	}

	nonzero := &Int{Value: 42}
	if nonzero.ToBool() != TRUE {
		t.Error("Int(42).ToBool() should be TRUE")
	}
}

func TestIntHashKey(t *testing.T) {
	a := &Int{Value: 42}
	b := &Int{Value: 42}
	c := &Int{Value: 43}

	if a.HashKey() != b.HashKey() {
		t.Error("same int values should have same hash keys")
	}
	if a.HashKey() == c.HashKey() {
		t.Error("different int values should have different hash keys")
	}
}

func TestFloatInspect(t *testing.T) {
	f := &Float{Value: 3.14}
	if got := f.Inspect(); got != "3.14" {
		t.Errorf("Float.Inspect() = %s, want 3.14", got)
	}
}

func TestFloatType(t *testing.T) {
	f := &Float{Value: 3.14}
	if got := f.Type(); got != FloatType {
		t.Errorf("Float.Type() = %s, want FLOAT", got)
	}
}

func TestFloatToBool(t *testing.T) {
	zero := &Float{Value: 0.0}
	if zero.ToBool() != FALSE {
		t.Error("Float(0).ToBool() should be FALSE")
	}

	nonzero := &Float{Value: 3.14}
	if nonzero.ToBool() != TRUE {
		t.Error("Float(3.14).ToBool() should be TRUE")
	}
}
