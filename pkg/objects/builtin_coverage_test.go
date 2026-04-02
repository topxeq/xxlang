// pkg/objects/builtin_coverage_test.go
package objects

import (
	"testing"
)

// TestBuiltinLen tests len builtin
func TestBuiltinLenCoverage(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected int64
		hasError bool
	}{
		{"string", []Object{NewString("hello")}, 5, false},
		{"empty string", []Object{NewString("")}, 0, false},
		{"array", []Object{NewArray([]Object{NewInt(1), NewInt(2)})}, 2, false},
		{"empty array", []Object{NewArray([]Object{})}, 0, false},
		{"map", []Object{NewMapWithCapacity(2)}, 0, false},
		{"no args", []Object{}, 0, true},
		{"invalid type", []Object{NewInt(42)}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Builtins["len"]
			result := fn.Fn(tt.args...)

			if tt.hasError {
				if _, ok := result.(*Error); !ok {
					t.Error("expected error")
				}
				return
			}

			n, ok := result.(*Int)
			if !ok {
				t.Fatalf("expected Int, got %T", result)
			}
			if n.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, n.Value)
			}
		})
	}
}

// TestBuiltinTypeOf tests typeOf builtin
func TestBuiltinTypeOfCoverage(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected string
	}{
		{"int", []Object{NewInt(42)}, "INT"},
		{"string", []Object{NewString("hello")}, "STRING"},
		{"bool", []Object{TRUE}, "BOOL"},
		{"null", []Object{NULL}, "NULL"},
		{"array", []Object{NewArray([]Object{})}, "ARRAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := Builtins["typeOf"]
			result := fn.Fn(tt.args...)

			s, ok := result.(*String)
			if !ok {
				t.Fatalf("expected String, got %T", result)
			}
			if s.Value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, s.Value)
			}
		})
	}
}

// TestBuiltinToStr tests toStr builtin
func TestBuiltinToStrCoverage(t *testing.T) {
	fn := Builtins["toStr"]

	tests := []struct {
		name string
		arg  Object
	}{
		{"int", NewInt(42)},
		{"string", NewString("hello")},
		{"bool", TRUE},
		{"null", NULL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			if _, ok := result.(*String); !ok {
				t.Errorf("expected String, got %T", result)
			}
		})
	}

	t.Run("no args", func(t *testing.T) {
		result := fn.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("expected error for no args")
		}
	})
}

// TestBuiltinToChars tests toChars builtin
func TestBuiltinToCharsCoverage(t *testing.T) {
	fn := Builtins["toChars"]

	t.Run("valid string", func(t *testing.T) {
		result := fn.Fn(NewString("hello"))
		c, ok := result.(*Chars)
		if !ok {
			t.Fatalf("expected Chars, got %T", result)
		}
		if len(c.Value) != 5 {
			t.Errorf("expected 5 chars, got %d", len(c.Value))
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		result := fn.Fn(NewInt(42))
		if _, ok := result.(*Error); !ok {
			t.Error("expected error for invalid type")
		}
	})
}

// TestBuiltinCharLen tests charLen builtin
func TestBuiltinCharLenCoverage(t *testing.T) {
	fn := Builtins["charLen"]

	t.Run("ascii string", func(t *testing.T) {
		result := fn.Fn(NewString("hello"))
		n, ok := result.(*Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 5 {
			t.Errorf("expected 5, got %d", n.Value)
		}
	})

	t.Run("unicode string", func(t *testing.T) {
		result := fn.Fn(NewString("你好世界"))
		n, ok := result.(*Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 4 {
			t.Errorf("expected 4, got %d", n.Value)
		}
	})
}

// TestBuiltinSubstr tests substr builtin
func TestBuiltinSubstrCoverage(t *testing.T) {
	fn := Builtins["substr"]

	tests := []struct {
		name     string
		args     []Object
		expected string
		hasError bool
	}{
		{"slice", []Object{NewString("hello"), NewInt(1), NewInt(4)}, "ell", false},
		{"to end", []Object{NewString("hello"), NewInt(2)}, "llo", false},
		{"invalid start type", []Object{NewString("hello"), NewString("1")}, "", true},
		{"invalid string type", []Object{NewInt(42), NewInt(1)}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.args...)

			if tt.hasError {
				if _, ok := result.(*Error); !ok {
					t.Error("expected error")
				}
				return
			}

			s, ok := result.(*String)
			if !ok {
				t.Fatalf("expected String, got %T", result)
			}
			if s.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, s.Value)
			}
		})
	}
}

// TestBuiltinInt tests int builtin
func TestBuiltinIntCoverage(t *testing.T) {
	fn := Builtins["int"]

	tests := []struct {
		name     string
		arg      Object
		expected int64
	}{
		{"from float", NewFloat(42.9), 42},
		{"from string", NewString("42"), 42},
		{"from int", NewInt(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			n, ok := result.(*Int)
			if !ok {
				t.Fatalf("expected Int, got %T", result)
			}
			if n.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, n.Value)
			}
		})
	}
}

// TestBuiltinFloat tests float builtin
func TestBuiltinFloatCoverage(t *testing.T) {
	fn := Builtins["float"]

	tests := []struct {
		name     string
		arg      Object
		expected float64
	}{
		{"from int", NewInt(42), 42.0},
		{"from string", NewString("3.14"), 3.14},
		{"from float", NewFloat(3.14), 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			f, ok := result.(*Float)
			if !ok {
				t.Fatalf("expected Float, got %T", result)
			}
			if f.Value != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, f.Value)
			}
		})
	}
}

// TestBuiltinString tests string builtin
func TestBuiltinStringCoverage(t *testing.T) {
	fn := Builtins["string"]

	tests := []struct {
		name     string
		arg      Object
		expected string
	}{
		{"from int", NewInt(42), "42"},
		{"from float", NewFloat(3.14), "3.14"},
		{"from bool", TRUE, "true"},
		{"from string", NewString("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			s, ok := result.(*String)
			if !ok {
				t.Fatalf("expected String, got %T", result)
			}
			if s.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, s.Value)
			}
		})
	}
}

// TestBuiltinAbs tests abs builtin
func TestBuiltinAbsCoverage(t *testing.T) {
	fn := Builtins["abs"]

	tests := []struct {
		name     string
		arg      Object
		expected int64
	}{
		{"positive int", NewInt(42), 42},
		{"negative int", NewInt(-42), 42},
		{"zero", NewInt(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			n, ok := result.(*Int)
			if !ok {
				t.Fatalf("expected Int, got %T", result)
			}
			if n.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, n.Value)
			}
		})
	}
}

// TestBuiltinCeil tests ceil builtin
func TestBuiltinCeilCoverage(t *testing.T) {
	fn := Builtins["ceil"]
	if fn == nil {
		t.Skip("ceil builtin not found")
	}

	tests := []struct {
		name     string
		arg      Object
		expected int64
	}{
		{"float 3.2", NewFloat(3.2), 4},
		{"float -3.2", NewFloat(-3.2), -3},
		{"int", NewInt(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			n, ok := result.(*Int)
			if !ok {
				t.Fatalf("expected Int, got %T", result)
			}
			if n.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, n.Value)
			}
		})
	}
}

// TestBuiltinFloor tests floor builtin
func TestBuiltinFloorCoverage(t *testing.T) {
	fn := Builtins["floor"]
	if fn == nil {
		t.Skip("floor builtin not found")
	}

	tests := []struct {
		name     string
		arg      Object
		expected int64
	}{
		{"float 3.7", NewFloat(3.7), 3},
		{"float -3.7", NewFloat(-3.7), -4},
		{"int", NewInt(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			n, ok := result.(*Int)
			if !ok {
				t.Fatalf("expected Int, got %T", result)
			}
			if n.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, n.Value)
			}
		})
	}
}

// TestBuiltinSqrt tests sqrt builtin
func TestBuiltinSqrtCoverage(t *testing.T) {
	fn := Builtins["sqrt"]

	result := fn.Fn(NewInt(16))
	f, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value != 4.0 {
		t.Errorf("expected 4.0, got %f", f.Value)
	}
}

// TestBuiltinPow tests pow builtin
func TestBuiltinPowCoverage(t *testing.T) {
	fn := Builtins["pow"]

	result := fn.Fn(NewInt(2), NewInt(3))
	f, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if f.Value != 8.0 {
		t.Errorf("expected 8.0, got %f", f.Value)
	}
}

// TestBuiltinMin tests min builtin
func TestBuiltinMinCoverage(t *testing.T) {
	fn := Builtins["min"]
	if fn == nil {
		t.Skip("min builtin not found")
	}

	result := fn.Fn(NewInt(3), NewInt(1))
	n, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value != 1 {
		t.Errorf("expected 1, got %d", n.Value)
	}
}

// TestBuiltinMax tests max builtin
func TestBuiltinMaxCoverage(t *testing.T) {
	fn := Builtins["max"]
	if fn == nil {
		t.Skip("max builtin not found")
	}

	result := fn.Fn(NewInt(3), NewInt(1))
	n, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value != 3 {
		t.Errorf("expected 3, got %d", n.Value)
	}
}

// TestBuiltinIsEmpty tests isEmpty builtin
func TestBuiltinIsEmptyCoverage(t *testing.T) {
	fn := Builtins["isEmpty"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"empty string", NewString(""), true},
		{"non-empty string", NewString("hello"), false},
		{"empty array", NewArray([]Object{}), true},
		{"non-empty array", NewArray([]Object{NewInt(1)}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsString tests isString builtin
func TestBuiltinIsStringCoverage(t *testing.T) {
	fn := Builtins["isString"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"string", NewString("hello"), true},
		{"int", NewInt(42), false},
		{"array", NewArray([]Object{}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsNumber tests isNumber builtin
func TestBuiltinIsNumberCoverage(t *testing.T) {
	fn := Builtins["isNumber"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"int", NewInt(42), true},
		{"float", NewFloat(3.14), true},
		{"string", NewString("42"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsInt tests isInt builtin
func TestBuiltinIsIntCoverage(t *testing.T) {
	fn := Builtins["isInt"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"int", NewInt(42), true},
		{"float", NewFloat(3.14), false},
		{"string", NewString("42"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsFloat tests isFloat builtin
func TestBuiltinIsFloatCoverage(t *testing.T) {
	fn := Builtins["isFloat"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"float", NewFloat(3.14), true},
		{"int", NewInt(42), false},
		{"string", NewString("3.14"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsArray tests isArray builtin
func TestBuiltinIsArrayCoverage(t *testing.T) {
	fn := Builtins["isArray"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"array", NewArray([]Object{}), true},
		{"int", NewInt(42), false},
		{"map", NewMapWithCapacity(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsMap tests isMap builtin
func TestBuiltinIsMapCoverage(t *testing.T) {
	fn := Builtins["isMap"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"map", NewMapWithCapacity(0), true},
		{"int", NewInt(42), false},
		{"array", NewArray([]Object{}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsBool tests isBool builtin
func TestBuiltinIsBoolCoverage(t *testing.T) {
	fn := Builtins["isBool"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"bool true", TRUE, true},
		{"bool false", FALSE, true},
		{"int", NewInt(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsNull tests isNull builtin
func TestBuiltinIsNullCoverage(t *testing.T) {
	fn := Builtins["isNull"]

	tests := []struct {
		name     string
		arg      Object
		expected bool
	}{
		{"null", NULL, true},
		{"int", NewInt(0), false},
		{"empty string", NewString(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn.Fn(tt.arg)
			b, ok := result.(*Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", result)
			}
			if b.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Value)
			}
		})
	}
}

// TestBuiltinIsFunction tests isFunction builtin
func TestBuiltinIsFunctionCoverage(t *testing.T) {
	fn := Builtins["isFunction"]

	t.Run("function", func(t *testing.T) {
		result := fn.Fn(&Function{})
		b, ok := result.(*Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("expected true for Function")
		}
	})

	t.Run("int", func(t *testing.T) {
		result := fn.Fn(NewInt(42))
		b, ok := result.(*Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value {
			t.Error("expected false for Int")
		}
	})
}
