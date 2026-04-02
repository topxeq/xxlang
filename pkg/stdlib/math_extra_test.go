package stdlib

import (
	"math"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// mathCall invokes a builtin function from the math module, mirroring the
// pattern used in json_extra_test.go.
func mathCall(name string, args ...objects.Object) objects.Object {
	mod := Get("math")
	if mod == nil {
		return &objects.Error{Message: "math module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

func approx(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestMathInit(t *testing.T) {
	t.Run("PI", func(t *testing.T) {
		// PI is a constant in the math module, not a builtin function.
		mod := Get("math")
		if mod == nil {
			t.Fatalf("math module not found")
		}
		p, ok := mod.Exports["PI"].(*objects.Float)
		if !ok {
			t.Fatalf("PI should be Float, got %T", mod.Exports["PI"])
		}
		if !approx(p.Value, math.Pi, 1e-12) {
			t.Fatalf("PI value mismatch: got %v want %v", p.Value, math.Pi)
		}
	})

	t.Run("E", func(t *testing.T) {
		mod := Get("math")
		if mod == nil {
			t.Fatalf("math module not found")
		}
		e, ok := mod.Exports["E"].(*objects.Float)
		if !ok {
			t.Fatalf("E should be Float, got %T", mod.Exports["E"])
		}
		if !approx(e.Value, math.E, 1e-12) {
			t.Fatalf("E value mismatch: got %v want %v", e.Value, math.E)
		}
	})

	t.Run("Abs_Int_and_Float", func(t *testing.T) {
		r := mathCall("abs", Int(-7))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 7 {
			t.Fatalf("abs(-7) expected 7 Int, got %T %v", r, r)
		}
		r = mathCall("abs", Float(-3.5))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 3.5, 1e-12) {
			t.Fatalf("abs(-3.5) expected 3.5 Float, got %T %v", r, r)
		}
	})

	t.Run("Ceil and Floor", func(t *testing.T) {
		r := mathCall("ceil", Float(2.3))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 3 {
			t.Fatalf("ceil(2.3) expected 3 Int, got %T %v", r, r)
		}
		r = mathCall("floor", Float(2.9))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 2 {
			t.Fatalf("floor(2.9) expected 2 Int, got %T %v", r, r)
		}
	})

	t.Run("Round with and without precision", func(t *testing.T) {
		r := mathCall("round", Float(2.5))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 3 {
			t.Fatalf("round(2.5) expected 3 Int, got %T %v", r, r)
		}
		r = mathCall("round", Float(2.345), Int(2))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 2.35, 1e-12) {
			t.Fatalf("round(2.345, 2) expected 2.35 Float, got %T %v", r, r)
		}
	})

	t.Run("Sqrt and Pow", func(t *testing.T) {
		r := mathCall("sqrt", Int(9))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 3.0, 1e-12) {
			t.Fatalf("sqrt(9) should be 3.0 Float, got %T %v", r, r)
		}
		r = mathCall("pow", Int(2), Int(3))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 8.0, 1e-12) {
			t.Fatalf("pow(2,3) should be 8.0, got %T %v", r, r)
		}
	})

	t.Run("Trig Functions", func(t *testing.T) {
		r := mathCall("sin", Float(math.Pi/6))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 0.5, 1e-12) {
			t.Fatalf("sin(pi/6) expected 0.5, got %T %v", r, r)
		}
		r = mathCall("cos", Float(0))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 1.0, 1e-12) {
			t.Fatalf("cos(0) expected 1, got %T %v", r, r)
		}
		r = mathCall("tan", Float(0))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 0.0, 1e-12) {
			t.Fatalf("tan(0) expected 0, got %T %v", r, r)
		}
	})

	t.Run("Asin/Acos/Atan/Atan2", func(t *testing.T) {
		r := mathCall("asin", Float(0.5))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.Asin(0.5), 1e-12) {
			t.Fatalf("asin(0.5) mismatch: %T %v", r, r)
		}
		r = mathCall("acos", Float(0))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.Acos(0), 1e-12) {
			t.Fatalf("acos(0) mismatch: %T %v", r, r)
		}
		r = mathCall("atan", Float(1))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.Atan(1), 1e-12) {
			t.Fatalf("atan(1) mismatch: %T %v", r, r)
		}
		r = mathCall("atan2", Int(1), Int(1))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.Atan2(1, 1), 1e-12) {
			t.Fatalf("atan2(1,1) mismatch: %T %v", r, r)
		}
	})

	t.Run("Log/Log10/Log2/Exp", func(t *testing.T) {
		r := mathCall("log", Float(math.E))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 1.0, 1e-12) {
			t.Fatalf("log(e) mismatch: %T %v", r, r)
		}
		r = mathCall("log10", Float(1000))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 3.0, 1e-12) {
			t.Fatalf("log10(1000) mismatch: %T %v", r, r)
		}
		r = mathCall("log2", Float(8))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 3.0, 1e-12) {
			t.Fatalf("log2(8) mismatch: %T %v", r, r)
		}
		r = mathCall("exp", Float(1))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.E, 1e-12) {
			t.Fatalf("exp(1) mismatch: %T %v", r, r)
		}
	})

	t.Run("Min/Max", func(t *testing.T) {
		r := mathCall("min", Int(3), Int(5), Int(2))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 2 {
			t.Fatalf("min(3,5,2) expected 2, got %T %v", r, r)
		}
		r = mathCall("max", Int(3), Int(5), Int(2))
		if ii, ok := r.(*objects.Int); !ok || ii.Value != 5 {
			t.Fatalf("max(3,5,2) expected 5, got %T %v", r, r)
		}
	})

	t.Run("Random", func(t *testing.T) {
		r := mathCall("random")
		f, ok := r.(*objects.Float)
		if !ok {
			t.Fatalf("random should return Float, got %T", r)
		}
		if f.Value < 0 || f.Value >= 1 {
			t.Fatalf("random out of range: %v", f.Value)
		}
	})

	t.Run("DegToRad/RadToDeg", func(t *testing.T) {
		r := mathCall("degToRad", Int(180))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, math.Pi, 1e-12) {
			t.Fatalf("degToRad(180) mismatch: %T %v", r, r)
		}
		r = mathCall("radToDeg", Float(math.Pi))
		if f, ok := r.(*objects.Float); !ok || !approx(f.Value, 180, 1e-12) {
			t.Fatalf("radToDeg(pi) mismatch: %T %v", r, r)
		}
	})
}
