// pkg/objects/builtin_math_test.go
package objects

import (
	"math"
	"testing"
)

func TestBuiltinAdjustFloat(t *testing.T) {
	fn, ok := Builtins["adjustFloat"]
	if !ok {
		t.Fatal("adjustFloat builtin not found")
	}

	result := fn.Fn(NewFloat(3.14159265359))
	floatResult, ok := result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 3.1415926536 {
		t.Errorf("expected 3.1415926536, got %f", floatResult.Value)
	}

	result = fn.Fn(NewFloat(3.14159265359), NewInt(4))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 3.1416 {
		t.Errorf("expected 3.1416, got %f", floatResult.Value)
	}

	result = fn.Fn(NewInt(5))
	floatResult, ok = result.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 5 {
		t.Errorf("expected 5, got %f", floatResult.Value)
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for non-numeric argument")
	}

	result = fn.Fn(NewFloat(1.5), NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for non-int precision")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinToKMG(t *testing.T) {
	fn, ok := Builtins["toKMG"]
	if !ok {
		t.Fatal("toKMG builtin not found")
	}

	testCases := []struct {
		input    Object
		expected string
	}{
		{NewInt(500), "500"},
		{NewInt(1024), "1K"},
		{NewInt(1536), "1.5K"},
		{NewInt(1048576), "1M"},
		{NewInt(1073741824), "1G"},
		{NewInt(1099511627776), "1T"},
		{NewFloat(500.0), "500"},
	}

	for _, tc := range testCases {
		result := fn.Fn(tc.input)
		strResult, ok := result.(*String)
		if !ok {
			t.Fatalf("expected String for %T, got %T", tc.input, result)
		}
		if strResult.Value != tc.expected {
			t.Errorf("expected %s for %v, got %s", tc.expected, tc.input, strResult.Value)
		}
	}

	result := fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for non-numeric argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTrunc(t *testing.T) {
	fn, ok := Builtins["trunc"]
	if !ok {
		t.Fatal("trunc builtin not found")
	}

	result := fn.Fn(NewFloat(3.7))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("expected 3, got %d", intResult.Value)
	}

	result = fn.Fn(NewFloat(-3.7))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -3 {
		t.Errorf("expected -3, got %d", intResult.Value)
	}

	result = fn.Fn(NewFloat(3.0))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("expected 3, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(5))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for non-numeric argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsInf(t *testing.T) {
	fn, ok := Builtins["isInf"]
	if !ok {
		t.Fatal("isInf builtin not found")
	}

	result := fn.Fn(NewFloat(math.Inf(1)))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for +Inf")
	}

	result = fn.Fn(NewFloat(math.Inf(-1)))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for -Inf")
	}

	result = fn.Fn(NewFloat(3.14))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for finite value")
	}

	result = fn.Fn(NewString("invalid"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for non-numeric")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsNaN(t *testing.T) {
	fn, ok := Builtins["isNaN"]
	if !ok {
		t.Fatal("isNaN builtin not found")
	}

	result := fn.Fn(NewFloat(math.NaN()))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for NaN")
	}

	result = fn.Fn(NewFloat(3.14))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for non-NaN")
	}

	result = fn.Fn(NewString("invalid"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for non-numeric")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsFinite(t *testing.T) {
	fn, ok := Builtins["isFinite"]
	if !ok {
		t.Fatal("isFinite builtin not found")
	}

	result := fn.Fn(NewFloat(3.14))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for finite value")
	}

	result = fn.Fn(NewFloat(math.Inf(1)))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for +Inf")
	}

	result = fn.Fn(NewFloat(math.NaN()))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for NaN")
	}

	result = fn.Fn(NewString("invalid"))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for non-numeric")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}
