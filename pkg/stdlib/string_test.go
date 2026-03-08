// pkg/stdlib/string_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callStringFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/string")
	if mod == nil {
		panic("std/string module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestStringLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"hello", 5},
		{"", 0},
		{"hello world", 11},
	}

	for _, tt := range tests {
		result := callStringFunc("len", String(tt.input))
		r, ok := result.(*objects.Int)
		if !ok || r.Value != tt.expected {
			t.Errorf("len(%q) = %v, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestStringToUpperLower(t *testing.T) {
	tests := []struct {
		input  string
		upper  string
		lower  string
	}{
		{"hello", "HELLO", "hello"},
		{"HELLO", "HELLO", "hello"},
		{"Hello World", "HELLO WORLD", "hello world"},
	}

	for _, tt := range tests {
		result := callStringFunc("toUpper", String(tt.input))
		r, ok := result.(*objects.String)
		if !ok || r.Value != tt.upper {
			t.Errorf("toUpper(%q) = %v, want %v", tt.input, result, tt.upper)
		}

		result = callStringFunc("toLower", String(tt.input))
		r, ok = result.(*objects.String)
		if !ok || r.Value != tt.lower {
			t.Errorf("toLower(%q) = %v, want %v", tt.input, result, tt.lower)
		}
	}
}

func TestStringTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello  ", "hello"},
		{"\t\n", ""},
	}

	for _, tt := range tests {
		result := callStringFunc("trim", String(tt.input))
		r, ok := result.(*objects.String)
		if !ok || r.Value != tt.expected {
			t.Errorf("trim(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestStringSplit(t *testing.T) {
	result := callStringFunc("split", String("a,b,c"), String(","))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}

	if len(arr.Elements) != 3 {
		t.Errorf("split(\"a,b,c\", \",\") = %d elements, want 3", len(arr.Elements))
	}

	expected := []string{"a", "b", "c"}
	for i := 0; i < len(expected); i++ {
		s, ok := arr.Elements[i].(*objects.String)
		if !ok || s.Value != expected[i] {
			t.Errorf("split result[%d] = %v, want %v", i, arr.Elements[i], expected[i])
		}
	}
}

func TestStringJoin(t *testing.T) {
	arr := Array(String("a"), String("b"), String("c"))
	result := callStringFunc("join", arr, String(","))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatal("join should return string")
	}
	if r.Value != "a,b,c" {
		t.Errorf("join([\"a\",\"b\",\"c\"], \",\") = %v, want \"a,b,c\"", r.Value)
	}
}

func TestStringContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"hello world", "ello", true},
	}

	for _, tt := range tests {
		result := callStringFunc("contains", String(tt.s), String(tt.substr))
		r, ok := result.(*objects.Bool)
		if !ok {
			t.Fatal("contains should return bool")
		}
		if r.Value != tt.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, r.Value, tt.want)
		}
	}
}

func TestStringHasPrefix(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		want   bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"hello world", "he", true},
	}

	for _, tt := range tests {
		result := callStringFunc("hasPrefix", String(tt.s), String(tt.prefix))
		r, ok := result.(*objects.Bool)
		if !ok {
			t.Fatal("hasPrefix should return bool")
		}
		if r.Value != tt.want {
			t.Errorf("hasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, r.Value, tt.want)
		}
	}
}

func TestStringHasSuffix(t *testing.T) {
	tests := []struct {
		s      string
		suffix string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", false},
		{"hello world", "d", true},
	}

	for _, tt := range tests {
		result := callStringFunc("hasSuffix", String(tt.s), String(tt.suffix))
		r, ok := result.(*objects.Bool)
		if !ok {
			t.Fatal("hasSuffix should return bool")
		}
		if r.Value != tt.want {
			t.Errorf("hasSuffix(%q, %q) = %v, want %v", tt.s, tt.suffix, r.Value, tt.want)
		}
	}
}

func TestStringIndexOf(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   int64
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "foo", -1},
		{"hello world", "o", 4},
	}

	for _, tt := range tests {
		result := callStringFunc("indexOf", String(tt.s), String(tt.substr))
		r, ok := result.(*objects.Int)
		if !ok {
			t.Fatal("indexOf should return int")
		}
		if r.Value != tt.want {
			t.Errorf("indexOf(%q, %q) = %d, want %d", tt.s, tt.substr, r.Value, tt.want)
		}
	}
}

func TestStringReplace(t *testing.T) {
	result := callStringFunc("replace", String("hello world"), String("world"), String("universe"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatal("replace should return string")
	}
	if r.Value != "hello universe" {
		t.Errorf("replace(\"hello world\", \"world\", \"universe\") = %v, want \"hello universe\"", r.Value)
	}
}

func TestStringRepeat(t *testing.T) {
	result := callStringFunc("repeat", String("ab"), Int(3))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatal("repeat should return string")
	}
	if r.Value != "ababab" {
		t.Errorf("repeat(\"ab\", 3) = %v, want \"ababab\"", r.Value)
	}
}

func TestStringReverse(t *testing.T) {
	result := callStringFunc("reverse", String("hello"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatal("reverse should return string")
	}
	if r.Value != "olleh" {
		t.Errorf("reverse(\"hello\") = %v, want \"olleh\"", r.Value)
	}
}

func TestStringParseInt(t *testing.T) {
	result := callStringFunc("parseInt", String("42"))
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("parseInt should return int")
	}
	if r.Value != 42 {
		t.Errorf("parseInt(\"42\") = %d, want 42", r.Value)
	}

	// Test error case
	err := callStringFunc("parseInt", String("not a number"))
	if _, ok := err.(*objects.Error); !ok {
		t.Errorf("parseInt(\"not a number\") should return error, got %v", err)
	}
}

func TestStringParseFloat(t *testing.T) {
	result := callStringFunc("parseFloat", String("3.14"))
	r, ok := result.(*objects.Float)
	if !ok {
		t.Fatal("parseFloat should return float")
	}
	if r.Value != 3.14 {
		t.Errorf("parseFloat(\"3.14\") = %v, want 3.14", r.Value)
	}
}

func TestStringToString(t *testing.T) {
	tests := []struct {
		input objects.Object
		want  string
	}{
		{Int(42), "42"},
		{Float(3.14), "3.14"},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{String("hello"), "hello"},
	}

	for _, tt := range tests {
		result := callStringFunc("toString", tt.input)
		r, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("toString should return string, got %T", result)
		}
		if r.Value != tt.want {
			t.Errorf("toString(%v) = %v, want %v", tt.input, r.Value, tt.want)
		}
	}
}
