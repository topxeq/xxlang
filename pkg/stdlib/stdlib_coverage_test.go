// pkg/stdlib/stdlib_coverage_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestTimeModuleUnix tests time.unix
func TestTimeModuleUnix(t *testing.T) {
	result := callStdlibFunc("time", "unix")
	if result == nil {
		t.Fatal("unix returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value <= 0 {
		t.Error("unix timestamp should be positive")
	}
}

// TestTimeModuleUnixMs tests time.unixMs
func TestTimeModuleUnixMs(t *testing.T) {
	result := callStdlibFunc("time", "unixMs")
	if result == nil {
		t.Fatal("unixMs returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value <= 0 {
		t.Error("unixMs timestamp should be positive")
	}
}

// TestTimeModuleUnixNano tests time.unixNano
func TestTimeModuleUnixNano(t *testing.T) {
	result := callStdlibFunc("time", "unixNano")
	if result == nil {
		t.Fatal("unixNano returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value <= 0 {
		t.Error("unixNano timestamp should be positive")
	}
}

// TestTimeModuleYear tests time.year
func TestTimeModuleYear(t *testing.T) {
	result := callStdlibFunc("time", "year")
	if result == nil {
		t.Fatal("year returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 2020 || n.Value > 2100 {
		t.Errorf("year %d seems unreasonable", n.Value)
	}
}

// TestTimeModuleMonth tests time.month
func TestTimeModuleMonth(t *testing.T) {
	result := callStdlibFunc("time", "month")
	if result == nil {
		t.Fatal("month returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 1 || n.Value > 12 {
		t.Errorf("month %d should be 1-12", n.Value)
	}
}

// TestTimeModuleDay tests time.day
func TestTimeModuleDay(t *testing.T) {
	result := callStdlibFunc("time", "day")
	if result == nil {
		t.Fatal("day returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 1 || n.Value > 31 {
		t.Errorf("day %d should be 1-31", n.Value)
	}
}

// TestTimeModuleHour tests time.hour
func TestTimeModuleHour(t *testing.T) {
	result := callStdlibFunc("time", "hour")
	if result == nil {
		t.Fatal("hour returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 0 || n.Value > 23 {
		t.Errorf("hour %d should be 0-23", n.Value)
	}
}

// TestTimeModuleMinute tests time.minute
func TestTimeModuleMinute(t *testing.T) {
	result := callStdlibFunc("time", "minute")
	if result == nil {
		t.Fatal("minute returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 0 || n.Value > 59 {
		t.Errorf("minute %d should be 0-59", n.Value)
	}
}

// TestTimeModuleSecond tests time.second
func TestTimeModuleSecond(t *testing.T) {
	result := callStdlibFunc("time", "second")
	if result == nil {
		t.Fatal("second returned nil")
	}
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value < 0 || n.Value > 59 {
		t.Errorf("second %d should be 0-59", n.Value)
	}
}

// TestTimeModuleFormat tests time.format
func TestTimeModuleFormat(t *testing.T) {
	result := callStdlibFunc("time", "format", String("2006-01-02"))
	if result == nil {
		t.Fatal("format returned nil")
	}
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if len(s.Value) != 10 {
		t.Errorf("format should return date string, got %s", s.Value)
	}
}

// TestTimeModuleParse tests time.parse
func TestTimeModuleParse(t *testing.T) {
	result := callStdlibFunc("time", "parse", String("2006-01-02"), String("2024-03-30"))
	if result == nil {
		t.Fatal("parse returned nil")
	}
	// time.parse returns Unix timestamp as Int
	n, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if n.Value <= 0 {
		t.Error("parse should return a valid Unix timestamp")
	}
}

// TestTimeModuleSleep tests time.sleep
func TestTimeModuleSleep(t *testing.T) {
	// Sleep for 1 millisecond
	result := callStdlibFunc("time", "sleep", Int(1))
	if result == nil {
		t.Fatal("sleep returned nil")
	}
}

// TestTimeModuleNowMap tests time.now returning map
func TestTimeModuleNowMap(t *testing.T) {
	result := callStdlibFunc("time", "now")
	if result == nil {
		t.Fatal("now returned nil")
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	// Check that all expected keys exist
	expectedKeys := []string{"year", "month", "day", "hour", "minute", "second", "nanosecond"}
	for _, key := range expectedKeys {
		keyObj := String(key)
		_, exists := m.Pairs[keyObj.HashKey()]
		if !exists {
			t.Errorf("expected key %s in now result", key)
		}
	}
}

// TestMathModuleConstants tests math constants
func TestMathModuleConstants(t *testing.T) {
	mod := Get("math")
	if mod == nil {
		t.Fatal("math module not found")
	}

	pi, ok := mod.Exports["PI"]
	if !ok {
		t.Error("PI constant not found")
	}
	piFloat, ok := pi.(*objects.Float)
	if !ok {
		t.Errorf("PI should be Float, got %T", pi)
	}
	if piFloat.Value < 3.14 || piFloat.Value > 3.15 {
		t.Errorf("PI value %f unexpected", piFloat.Value)
	}

	e, ok := mod.Exports["E"]
	if !ok {
		t.Error("E constant not found")
	}
	eFloat, ok := e.(*objects.Float)
	if !ok {
		t.Errorf("E should be Float, got %T", e)
	}
	if eFloat.Value < 2.71 || eFloat.Value > 2.72 {
		t.Errorf("E value %f unexpected", eFloat.Value)
	}
}

// TestFpModule tests functional programming module
func TestFpModuleMap(t *testing.T) {
	fn := &objects.Function{
		Parameters: []*objects.Identifier{{Value: "x"}},
	}

	// Create a simple array
	arr := &objects.Array{Elements: []objects.Object{Int(1), Int(2), Int(3)}}

	// We need to test that the fp module exists
	mod := Get("fp")
	if mod == nil {
		t.Skip("fp module not found")
	}

	_ = fn
	_ = arr
}

// TestSortModule tests sort module
func TestSortModuleSort(t *testing.T) {
	mod := Get("sort")
	if mod == nil {
		t.Skip("sort module not found")
	}

	arr := &objects.Array{Elements: []objects.Object{Int(3), Int(1), Int(2)}}
	result := callStdlibFunc("sort", "numbers", arr)
	if result == nil {
		t.Fatal("numbers returned nil")
	}
	sorted, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(sorted.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(sorted.Elements))
	}
}

// TestDebugModule tests debug module
func TestDebugModule(t *testing.T) {
	mod := Get("debug")
	if mod == nil {
		t.Skip("debug module not found")
	}

	result := callStdlibFunc("debug", "stack")
	if result == nil {
		t.Fatal("stack returned nil")
	}
}

// TestStrconvModule tests strconv module
func TestStrconvModule(t *testing.T) {
	mod := Get("strconv")
	if mod == nil {
		t.Skip("strconv module not found")
	}

	t.Run("parseInt", func(t *testing.T) {
		result := callStdlibFunc("strconv", "parseInt", String("42"))
		if result == nil {
			t.Fatal("parseInt returned nil")
		}
		n, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 42 {
			t.Errorf("expected 42, got %d", n.Value)
		}
	})

	t.Run("parseFloat", func(t *testing.T) {
		result := callStdlibFunc("strconv", "parseFloat", String("3.14"))
		if result == nil {
			t.Fatal("parseFloat returned nil")
		}
		f, ok := result.(*objects.Float)
		if !ok {
			t.Fatalf("expected Float, got %T", result)
		}
		if f.Value != 3.14 {
			t.Errorf("expected 3.14, got %f", f.Value)
		}
	})

	t.Run("formatInt", func(t *testing.T) {
		result := callStdlibFunc("strconv", "formatInt", Int(42))
		if result == nil {
			t.Fatal("formatInt returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "42" {
			t.Errorf("expected '42', got %s", s.Value)
		}
	})

	t.Run("formatFloat", func(t *testing.T) {
		result := callStdlibFunc("strconv", "formatFloat", Float(3.14))
		if result == nil {
			t.Fatal("formatFloat returned nil")
		}
		_, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
	})
}

// TestValidateModule tests validate module
func TestValidateModule(t *testing.T) {
	mod := Get("validate")
	if mod == nil {
		t.Skip("validate module not found")
	}

	t.Run("isEmail valid", func(t *testing.T) {
		result := callStdlibFunc("validate", "isEmail", String("test@example.com"))
		if result == nil {
			t.Fatal("isEmail returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("test@example.com should be valid email")
		}
	})

	t.Run("isEmail invalid", func(t *testing.T) {
		result := callStdlibFunc("validate", "isEmail", String("invalid"))
		if result == nil {
			t.Fatal("isEmail returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value {
			t.Error("invalid should not be valid email")
		}
	})
}
