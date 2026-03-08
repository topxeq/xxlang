// pkg/stdlib/math_test.go
package stdlib

import (
	"math"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callMathFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/math")
	if mod == nil {
		panic("std/math module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestMathConstants(t *testing.T) {
	mod := Get("std/math")
	if mod == nil {
		t.Fatal("std/math module not found")
	}

	pi := mod.Exports["PI"].(*objects.Float)
	if !floatEquals(pi.Value, math.Pi) {
		t.Errorf("PI = %v, want %v", pi.Value, math.Pi)
	}

	e := mod.Exports["E"].(*objects.Float)
	if !floatEquals(e.Value, math.E) {
		t.Errorf("E = %v, want %v", e.Value, math.E)
	}
}

func TestAbs(t *testing.T) {
	// Test with integers
	result := callMathFunc("abs", Int(-5))
	if r, ok := result.(*objects.Int); !ok || r.Value != 5 {
		t.Errorf("abs(-5) = %v, want 5", result)
	}

	result = callMathFunc("abs", Int(5))
	if r, ok := result.(*objects.Int); !ok || r.Value != 5 {
		t.Errorf("abs(5) = %v, want 5", result)
	}

	// Test with floats
	result = callMathFunc("abs", Float(-3.14))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 3.14) {
		t.Errorf("abs(-3.14) = %v, want 3.14", result)
	}

	// Test error cases
	err := callMathFunc("abs")
	if _, ok := err.(*objects.Error); !ok {
		t.Errorf("abs() should return error, got %v", err)
	}

	err = callMathFunc("abs", String("not a number"))
	if _, ok := err.(*objects.Error); !ok {
		t.Errorf("abs(string) should return error, got %v", err)
	}
}

func TestCeilFloorRound(t *testing.T) {
	// Test ceil
	result := callMathFunc("ceil", Float(3.2))
	if r, ok := result.(*objects.Int); !ok || r.Value != 4 {
		t.Errorf("ceil(3.2) = %v, want 4", result)
	}

	result = callMathFunc("ceil", Float(-3.2))
	if r, ok := result.(*objects.Int); !ok || r.Value != -3 {
		t.Errorf("ceil(-3.2) = %v, want -3", result)
	}

	// Test floor
	result = callMathFunc("floor", Float(3.8))
	if r, ok := result.(*objects.Int); !ok || r.Value != 3 {
		t.Errorf("floor(3.8) = %v, want 3", result)
	}

	// Test round
	result = callMathFunc("round", Float(3.5))
	if r, ok := result.(*objects.Int); !ok || r.Value != 4 {
		t.Errorf("round(3.5) = %v, want 4", result)
	}

	result = callMathFunc("round", Float(3.2))
	if r, ok := result.(*objects.Int); !ok || r.Value != 3 {
		t.Errorf("round(3.2) = %v, want 3", result)
	}
}

func TestSqrt(t *testing.T) {
	result := callMathFunc("sqrt", Int(4))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 2) {
		t.Errorf("sqrt(4) = %v, want 2", result)
	}

	result = callMathFunc("sqrt", Float(16))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 4) {
		t.Errorf("sqrt(16) = %v, want 4", result)
	}

	// Test negative number
	err := callMathFunc("sqrt", Int(-1))
	if _, ok := err.(*objects.Error); !ok {
		t.Errorf("sqrt(-1) should return error, got %v", err)
	}
}

func TestPow(t *testing.T) {
	result := callMathFunc("pow", Int(2), Int(3))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 8) {
		t.Errorf("pow(2, 3) = %v, want 8", result)
	}

	result = callMathFunc("pow", Float(2), Float(0.5))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, math.Sqrt(2)) {
		t.Errorf("pow(2, 0.5) = %v, want sqrt(2)", result)
	}
}

func TestSinCosTan(t *testing.T) {
	result := callMathFunc("sin", Int(0))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 0) {
		t.Errorf("sin(0) = %v, want 0", result)
	}

	result = callMathFunc("cos", Int(0))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 1) {
		t.Errorf("cos(0) = %v, want 1", result)
	}

	result = callMathFunc("tan", Int(0))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 0) {
		t.Errorf("tan(0) = %v, want 0", result)
	}
}

func TestLogExp(t *testing.T) {
	result := callMathFunc("log", Int(1))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 0) {
		t.Errorf("log(1) = %v, want 0", result)
	}

	result = callMathFunc("log10", Int(10))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 1) {
		t.Errorf("log10(10) = %v, want 1", result)
	}

	result = callMathFunc("exp", Int(0))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 1) {
		t.Errorf("exp(0) = %v, want 1", result)
	}
}

func TestMinMax(t *testing.T) {
	result := callMathFunc("min", Int(3), Int(1), Int(2))
	if r, ok := result.(*objects.Int); !ok || r.Value != 1 {
		t.Errorf("min(3, 1, 2) = %v, want 1", result)
	}

	result = callMathFunc("max", Int(3), Int(1), Int(2))
	if r, ok := result.(*objects.Int); !ok || r.Value != 3 {
		t.Errorf("max(3, 1, 2) = %v, want 3", result)
	}

	// Test with floats
	result = callMathFunc("min", Float(1.5), Float(2.5))
	if r, ok := result.(*objects.Float); !ok || !floatEquals(r.Value, 1.5) {
		t.Errorf("min(1.5, 2.5) = %v, want 1.5", result)
	}

	// Test error for insufficient arguments
	err := callMathFunc("min", Int(1))
	if _, ok := err.(*objects.Error); !ok {
		t.Errorf("min(1) should return error, got %v", err)
	}
}

func TestRandom(t *testing.T) {
	result := callMathFunc("random")
	r, ok := result.(*objects.Float)
	if !ok {
		t.Errorf("random() should return Float, got %v", result)
	}
	if r.Value < 0 || r.Value > 1 {
		t.Errorf("random() = %v, should be in [0, 1]", r.Value)
	}
}
