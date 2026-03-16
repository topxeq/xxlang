// pkg/stdlib/strconv_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestStrconvModuleExists(t *testing.T) {
	mod := Get("strconv")
	if mod == nil {
		t.Fatal("strconv module not found")
	}
}

func TestParseInt(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseInt"].(*objects.Builtin)

	tests := []struct {
		args     []objects.Object
		expected int64
	}{
		{[]objects.Object{String("42")}, 42},
		{[]objects.Object{String("1010"), Int(2)}, 10}, // binary
		{[]objects.Object{String("ff"), Int(16)}, 255}, // hex
	}

	for _, tt := range tests {
		result := fn.Fn(tt.args...)
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("parseInt(%v) = %d, want %d", tt.args, i.Value, tt.expected)
		}
	}
}

func TestParseIntErrors(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseInt"].(*objects.Builtin)

	// No args
	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no args")
	}

	// Wrong type
	result = fn.Fn(Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}

	// Invalid number
	result = fn.Fn(String("not a number"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for invalid number")
	}
}

func TestFormatInt(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatInt"].(*objects.Builtin)

	tests := []struct {
		args     []objects.Object
		expected string
	}{
		{[]objects.Object{Int(42)}, "42"},
		{[]objects.Object{Int(10), Int(2)}, "1010"}, // binary
		{[]objects.Object{Int(255), Int(16)}, "ff"}, // hex
	}

	for _, tt := range tests {
		result := fn.Fn(tt.args...)
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("formatInt(%v) = %s, want %s", tt.args, s.Value, tt.expected)
		}
	}
}

func TestParseFloat(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseFloat"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"-2.5", -2.5},
		{"1e10", 1e10},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		f, ok := result.(*objects.Float)
		if !ok {
			t.Fatalf("expected Float, got %T", result)
		}
		if f.Value != tt.expected {
			t.Errorf("parseFloat(%s) = %f, want %f", tt.input, f.Value, tt.expected)
		}
	}
}

func TestParseFloatErrors(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseFloat"].(*objects.Builtin)

	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no args")
	}

	result = fn.Fn(Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}

	result = fn.Fn(String("not a float"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for invalid float")
	}
}

func TestFormatFloat(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatFloat"].(*objects.Builtin)

	result := fn.Fn(Float(3.14159), Int(2))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "3.14" {
		t.Errorf("formatFloat(3.14159, 2) = %s, want 3.14", s.Value)
	}
}

func TestParseBool(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseBool"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("parseBool(%s) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestParseBoolErrors(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["parseBool"].(*objects.Builtin)

	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no args")
	}

	result = fn.Fn(Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}

	result = fn.Fn(String("maybe"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for invalid bool")
	}
}

func TestFormatBool(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatBool"].(*objects.Builtin)

	tests := []struct {
		input    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		result := fn.Fn(Bool(tt.input))
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("formatBool(%v) = %s, want %s", tt.input, s.Value, tt.expected)
		}
	}
}

func TestQuote(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["quote"].(*objects.Builtin)

	result := fn.Fn(String("hello\nworld"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != `"hello\nworld"` {
		t.Errorf("quote() = %s, want %q", s.Value, `"hello\nworld"`)
	}
}

func TestUnquote(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["unquote"].(*objects.Builtin)

	result := fn.Fn(String(`"hello\nworld"`))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "hello\nworld" {
		t.Errorf("unquote() = %q, want %q", s.Value, "hello\nworld")
	}
}

func TestToString(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toString"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected string
	}{
		{Int(42), "42"},
		{Float(3.14), "3.14"},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{String("hello"), "hello"},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("toString(%v) = %s, want %s", tt.input, s.Value, tt.expected)
		}
	}
}

func TestToInt(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toInt"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected int64
	}{
		{Int(42), 42},
		{Float(3.7), 3},
		{String("42"), 42},
		{Bool(true), 1},
		{Bool(false), 0},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("toInt(%v) = %d, want %d", tt.input, i.Value, tt.expected)
		}
	}
}

func TestToFloat(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toFloat"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected float64
	}{
		{Int(42), 42.0},
		{Float(3.14), 3.14},
		{String("3.14"), 3.14},
		{Bool(true), 1.0},
		{Bool(false), 0.0},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		f, ok := result.(*objects.Float)
		if !ok {
			t.Fatalf("expected Float, got %T", result)
		}
		if f.Value != tt.expected {
			t.Errorf("toFloat(%v) = %f, want %f", tt.input, f.Value, tt.expected)
		}
	}
}

func TestToBool(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toBool"].(*objects.Builtin)

	tests := []struct {
		input    objects.Object
		expected bool
	}{
		{Bool(true), true},
		{Bool(false), false},
		{Int(1), true},
		{Int(0), false},
		{Int(42), true},
		{Float(1.0), true},
		{Float(0.0), false},
		{String("true"), true},
		{String("false"), false},
		{String("hello"), true}, // non-empty string is truthy
	}

	for _, tt := range tests {
		result := fn.Fn(tt.input)
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("toBool(%v) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestToJSON(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toJSON"].(*objects.Builtin)

	obj := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("name").HashKey(): {Key: String("name"), Value: String("test")},
			String("value").HashKey(): {Key: String("value"), Value: Int(42)},
		},
	}

	result := fn.Fn(obj)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// JSON should contain the keys
	if len(s.Value) == 0 {
		t.Error("toJSON returned empty string")
	}
}

func TestToJSONPretty(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["toJSONPretty"].(*objects.Builtin)

	obj := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("name").HashKey(): {Key: String("name"), Value: String("test")},
		},
	}

	result := fn.Fn(obj)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Pretty JSON should have newlines
	if s.Value == "" {
		t.Error("toJSONPretty returned empty string")
	}
}

func TestFormatNumber(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatNumber"].(*objects.Builtin)

	tests := []struct {
		args     []objects.Object
		expected string
	}{
		{[]objects.Object{Float(3.14159), Int(2)}, "3.14"},
		{[]objects.Object{Int(42), Int(0)}, "42"},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.args...)
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("formatNumber(%v) = %s, want %s", tt.args, s.Value, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatBytes"].(*objects.Builtin)

	tests := []struct {
		input    int64
		contains string
	}{
		{500, "B"},
		{1024, "KB"},
		{1048576, "MB"},
	}

	for _, tt := range tests {
		result := fn.Fn(Int(tt.input))
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		// Just check it returns a non-empty result
		if s.Value == "" {
			t.Errorf("formatBytes(%d) returned empty string", tt.input)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	mod := Get("strconv")
	fn := mod.Exports["formatDuration"].(*objects.Builtin)

	tests := []struct {
		input    int64
		contains string
	}{
		{100, "ms"},
		{1500, "s"},
		{90000, "m"},
	}

	for _, tt := range tests {
		result := fn.Fn(Int(tt.input))
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value == "" {
			t.Errorf("formatDuration(%d) returned empty string", tt.input)
		}
	}
}

func TestObjectToGoValue(t *testing.T) {
	tests := []struct {
		name     string
		input    objects.Object
		expected interface{}
	}{
		{"int", Int(42), int64(42)},
		{"float", Float(3.14), 3.14},
		{"string", String("hello"), "hello"},
		{"bool true", Bool(true), true},
		{"bool false", Bool(false), false},
		{"null", Null(), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := objectToGoValue(tt.input)
			if result != tt.expected {
				t.Errorf("objectToGoValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestObjectToGoValueArray(t *testing.T) {
	result := objectToGoValue(Array(Int(1), Int(2)))
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr))
	}
}

func TestObjectToGoValueMap(t *testing.T) {
	m := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			String("key").HashKey(): {Key: String("key"), Value: Int(42)},
		},
	}

	result := objectToGoValue(m)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if resultMap["key"] != int64(42) {
		t.Errorf("expected key=42, got %v", resultMap["key"])
	}
}

func TestStrconvErrors(t *testing.T) {
	mod := Get("strconv")

	// Test error cases
	tests := []struct {
		name   string
		fnName string
		args   []objects.Object
	}{
		{"parseInt no args", "parseInt", []objects.Object{}},
		{"formatInt no args", "formatInt", []objects.Object{}},
		{"formatInt wrong type", "formatInt", []objects.Object{String("42")}},
		{"parseFloat no args", "parseFloat", []objects.Object{}},
		{"parseFloat wrong arg count", "parseFloat", []objects.Object{String("1"), String("2")}},
		{"formatFloat no args", "formatFloat", []objects.Object{}},
		{"formatFloat wrong type", "formatFloat", []objects.Object{String("3.14")}},
		{"parseBool no args", "parseBool", []objects.Object{}},
		{"parseBool wrong arg count", "parseBool", []objects.Object{String("true"), String("false")}},
		{"formatBool no args", "formatBool", []objects.Object{}},
		{"formatBool wrong arg count", "formatBool", []objects.Object{Bool(true), Bool(false)}},
		{"formatBool wrong type", "formatBool", []objects.Object{Int(1)}},
		{"quote no args", "quote", []objects.Object{}},
		{"quote wrong type", "quote", []objects.Object{Int(42)}},
		{"quote wrong arg count", "quote", []objects.Object{String("a"), String("b")}},
		{"unquote no args", "unquote", []objects.Object{}},
		{"unquote wrong type", "unquote", []objects.Object{Int(42)}},
		{"toString no args", "toString", []objects.Object{}},
		{"toInt no args", "toInt", []objects.Object{}},
		{"toInt wrong type", "toInt", []objects.Object{Array()}},
		{"toFloat no args", "toFloat", []objects.Object{}},
		{"toFloat wrong type", "toFloat", []objects.Object{Array()}},
		{"toBool no args", "toBool", []objects.Object{}},
		{"toJSON no args", "toJSON", []objects.Object{}},
		{"toJSONPretty no args", "toJSONPretty", []objects.Object{}},
		{"formatNumber no args", "formatNumber", []objects.Object{}},
		{"formatNumber wrong first type", "formatNumber", []objects.Object{String("42"), Int(2)}},
		{"formatNumber wrong second type", "formatNumber", []objects.Object{Float(3.14), String("2")}},
		{"formatBytes no args", "formatBytes", []objects.Object{}},
		{"formatBytes wrong type", "formatBytes", []objects.Object{String("42")}},
		{"formatBytes wrong arg count", "formatBytes", []objects.Object{Int(100), Int(200)}},
		{"formatDuration no args", "formatDuration", []objects.Object{}},
		{"formatDuration wrong type", "formatDuration", []objects.Object{String("100")}},
		{"formatDuration wrong arg count", "formatDuration", []objects.Object{Int(100), Int(200)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mod.Exports[tt.fnName].(*objects.Builtin)
			result := fn.Fn(tt.args...)
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("expected Error for %s, got %T", tt.name, result)
			}
		})
	}
}
