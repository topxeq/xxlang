// pkg/stdlib/math.go
// Standard library math module.
package stdlib

import (
	"math"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "math",
		Exports: map[string]objects.Object{
			// Constants
			"PI": Float(math.Pi),
			"E":  Float(math.E),

			// Basic functions
			"abs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("abs() takes exactly 1 argument")
				}
				switch n := args[0].(type) {
				case *objects.Int:
					if n.Value < 0 {
						return Int(-n.Value)
					}
					return n
				case *objects.Float:
					return Float(math.Abs(n.Value))
				default:
					return Error("abs() requires a numeric argument")
				}
			}),

			"ceil": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("ceil() takes exactly 1 argument")
				}
				switch n := args[0].(type) {
				case *objects.Int:
					return n
				case *objects.Float:
					return Int(int64(math.Ceil(n.Value)))
				default:
					return Error("ceil() requires a numeric argument")
				}
			}),

			"floor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("floor() takes exactly 1 argument")
				}
				switch n := args[0].(type) {
				case *objects.Int:
					return n
				case *objects.Float:
					return Int(int64(math.Floor(n.Value)))
				default:
					return Error("floor() requires a numeric argument")
				}
			}),

			"round": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("round() takes 1 or 2 arguments")
				}
				var val float64
				switch n := args[0].(type) {
				case *objects.Int:
					if len(args) == 1 {
						return n
					}
					val = float64(n.Value)
				case *objects.Float:
					val = n.Value
				default:
					return Error("round() requires a numeric argument")
				}
				if len(args) == 1 {
					return Int(int64(math.Round(val)))
				}
				precision, ok := args[1].(*objects.Int)
				if !ok {
					return Error("precision must be INT")
				}
				p := int(precision.Value)
				multiplier := math.Pow(10, float64(p))
				result := math.Round(val*multiplier) / multiplier
				return Float(result)
			}),

			"sqrt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sqrt() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("sqrt() requires a numeric argument")
				}
				if v < 0 {
					return Error("sqrt() requires a non-negative argument")
				}
				return Float(math.Sqrt(v))
			}),

			"pow": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("pow() takes exactly 2 arguments")
				}
				var base, exp float64
				switch n := args[0].(type) {
				case *objects.Int:
					base = float64(n.Value)
				case *objects.Float:
					base = n.Value
				default:
					return Error("pow() requires numeric arguments")
				}
				switch n := args[1].(type) {
				case *objects.Int:
					exp = float64(n.Value)
				case *objects.Float:
					exp = n.Value
				default:
					return Error("pow() requires numeric arguments")
				}
				return Float(math.Pow(base, exp))
			}),

			"sin": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sin() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("sin() requires a numeric argument")
				}
				return Float(math.Sin(v))
			}),

			"cos": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("cos() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("cos() requires a numeric argument")
				}
				return Float(math.Cos(v))
			}),

			"tan": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("tan() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("tan() requires a numeric argument")
				}
				return Float(math.Tan(v))
			}),

			"asin": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("asin() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("asin() requires a numeric argument")
				}
				return Float(math.Asin(v))
			}),

			"acos": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("acos() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("acos() requires a numeric argument")
				}
				return Float(math.Acos(v))
			}),

			"atan": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("atan() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("atan() requires a numeric argument")
				}
				return Float(math.Atan(v))
			}),

			"atan2": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("atan2() takes exactly 2 arguments")
				}
				var y, x float64
				switch n := args[0].(type) {
				case *objects.Int:
					y = float64(n.Value)
				case *objects.Float:
					y = n.Value
				default:
					return Error("atan2() requires numeric arguments")
				}
				switch n := args[1].(type) {
				case *objects.Int:
					x = float64(n.Value)
				case *objects.Float:
					x = n.Value
				default:
					return Error("atan2() requires numeric arguments")
				}
				return Float(math.Atan2(y, x))
			}),

			"log": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("log() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("log() requires a numeric argument")
				}
				if v <= 0 {
					return Error("log() requires a positive argument")
				}
				return Float(math.Log(v))
			}),

			"log10": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("log10() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("log10() requires a numeric argument")
				}
				if v <= 0 {
					return Error("log10() requires a positive argument")
				}
				return Float(math.Log10(v))
			}),

			"log2": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("log2() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("log2() requires a numeric argument")
				}
				if v <= 0 {
					return Error("log2() requires a positive argument")
				}
				return Float(math.Log2(v))
			}),

			"exp": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("exp() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("exp() requires a numeric argument")
				}
				return Float(math.Exp(v))
			}),

			"min": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("min() requires at least 2 arguments")
				}
				minVal := args[0]
				for _, arg := range args[1:] {
					if compareNumeric(arg, minVal) < 0 {
						minVal = arg
					}
				}
				return minVal
			}),

			"max": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("max() requires at least 2 arguments")
				}
				maxVal := args[0]
				for _, arg := range args[1:] {
					if compareNumeric(arg, maxVal) > 0 {
						maxVal = arg
					}
				}
				return maxVal
			}),

			"random": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Float(float64(randomInt()) / float64(2147483647))
			}),

			"degToRad": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("degToRad() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("degToRad() requires a numeric argument")
				}
				return Float(v * math.Pi / 180)
			}),

			"radToDeg": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("radToDeg() takes exactly 1 argument")
				}
				var v float64
				switch n := args[0].(type) {
				case *objects.Int:
					v = float64(n.Value)
				case *objects.Float:
					v = n.Value
				default:
					return Error("radToDeg() requires a numeric argument")
				}
				return Float(v * 180 / math.Pi)
			}),
		},
	})
}

// compareNumeric compares two numeric objects.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareNumeric(a, b objects.Object) int {
	var av, bv float64
	switch n := a.(type) {
	case *objects.Int:
		av = float64(n.Value)
	case *objects.Float:
		av = n.Value
	}
	switch n := b.(type) {
	case *objects.Int:
		bv = float64(n.Value)
	case *objects.Float:
		bv = n.Value
	}
	if av < bv {
		return -1
	} else if av > bv {
		return 1
	}
	return 0
}

// randomInt returns a pseudo-random int64.
func randomInt() int {
	// Simple xorshift for deterministic "randomness"
	// In production, we'd use math/rand or crypto/rand
	return int(123456789 * 1103515245 % 2147483647)
}
