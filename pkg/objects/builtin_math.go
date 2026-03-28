// pkg/objects/builtin_math.go
// Math enhancement built-in functions for Xxlang
// Note: sin, cos, tan, asin, acos, atan, atan2, exp, log, log10, log2, pi, e, degToRad, radToDeg
// have been removed - use math module instead
package objects

import (
	"math"
)

func init() {
	// Float utilities (kept - not in math module)
	Builtins["adjustFloat"] = &Builtin{Fn: builtinAdjustFloat}
	Builtins["toKMG"] = &Builtin{Fn: builtinToKMG}

	// Additional math functions (kept - not in math module)
	Builtins["trunc"] = &Builtin{Fn: builtinTrunc}
	Builtins["isInf"] = &Builtin{Fn: builtinIsInf}
	Builtins["isNaN"] = &Builtin{Fn: builtinIsNaN}
	Builtins["isFinite"] = &Builtin{Fn: builtinIsFinite}
}

// getFloatValue extracts float value from Object
func getFloatValue(arg Object) (float64, bool) {
	switch v := arg.(type) {
	case *Int:
		return float64(v.Value), true
	case *Float:
		return v.Value, true
	default:
		return 0, false
	}
}

// builtinAdjustFloat - adjust float precision errors
// Usage: adjustFloat(x) -> float
//
//	adjustFloat(x, precision) -> float
func builtinAdjustFloat(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for adjustFloat. got=%d, want=1 or 2", len(args))
	}

	val, ok := getFloatValue(args[0])
	if !ok {
		return newError("first argument to 'adjustFloat' must be numeric, got %s", args[0].Type())
	}

	precision := 10
	if len(args) == 2 {
		p, ok := args[1].(*Int)
		if !ok {
			return newError("second argument to 'adjustFloat' must be INT, got %s", args[1].Type())
		}
		precision = int(p.Value)
	}

	multiplier := math.Pow10(precision)
	result := math.Round(val*multiplier) / multiplier
	return NewFloat(result)
}

// builtinToKMG - convert number to K/M/G/T format
// Usage: toKMG(x) -> string
func builtinToKMG(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toKMG. got=%d, want=1", len(args))
	}

	var val int64
	switch v := args[0].(type) {
	case *Int:
		val = v.Value
	case *Float:
		val = int64(v.Value)
	default:
		return newError("argument to 'toKMG' must be numeric, got %s", args[0].Type())
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case val >= TB:
		return NewString(formatFloat(float64(val)/float64(TB), "T"))
	case val >= GB:
		return NewString(formatFloat(float64(val)/float64(GB), "G"))
	case val >= MB:
		return NewString(formatFloat(float64(val)/float64(MB), "M"))
	case val >= KB:
		return NewString(formatFloat(float64(val)/float64(KB), "K"))
	default:
		return NewString(formatFloat(float64(val), ""))
	}
}

// builtinTrunc - truncate float to integer
// Usage: trunc(x) -> int
func builtinTrunc(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for trunc. got=%d, want=1", len(args))
	}

	val, ok := getFloatValue(args[0])
	if !ok {
		return newError("argument to 'trunc' must be numeric, got %s", args[0].Type())
	}

	return NewInt(int64(math.Trunc(val)))
}

// builtinIsInf - check if value is infinity
// Usage: isInf(x) -> bool
func builtinIsInf(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isInf. got=%d, want=1", len(args))
	}

	val, ok := getFloatValue(args[0])
	if !ok {
		return FALSE
	}

	return &Bool{Value: math.IsInf(val, 0)}
}

// builtinIsNaN - check if value is NaN
// Usage: isNaN(x) -> bool
func builtinIsNaN(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isNaN. got=%d, want=1", len(args))
	}

	val, ok := getFloatValue(args[0])
	if !ok {
		return FALSE
	}

	return &Bool{Value: math.IsNaN(val)}
}

// builtinIsFinite - check if value is finite
// Usage: isFinite(x) -> bool
func builtinIsFinite(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isFinite. got=%d, want=1", len(args))
	}

	val, ok := getFloatValue(args[0])
	if !ok {
		return FALSE
	}

	return &Bool{Value: !math.IsInf(val, 0) && !math.IsNaN(val)}
}

// Helper function
func formatFloat(val float64, suffix string) string {
	if val == float64(int64(val)) {
		if suffix == "" {
			return string(intToBytes(int64(val)))
		}
		return string(intToBytes(int64(val))) + suffix
	}
	if suffix == "" {
		return string(floatToBytes(val))
	}
	return string(floatToBytes(val)) + suffix
}

func intToBytes(n int64) []byte {
	if n == 0 {
		return []byte{'0'}
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return digits
}

func floatToBytes(f float64) []byte {
	s := formatFloatSimple(f)
	return []byte(s)
}

func formatFloatSimple(f float64) string {
	str := ""
	if f < 0 {
		str = "-"
		f = -f
	}

	intPart := int64(f)
	fracPart := f - float64(intPart)

	str += string(intToBytes(intPart))

	if fracPart > 0 {
		str += "."
		for i := 0; i < 6 && fracPart > 0; i++ {
			fracPart *= 10
			digit := int64(fracPart)
			str += string(intToBytes(digit))
			fracPart -= float64(digit)
		}
	}

	return str
}
