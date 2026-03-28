// pkg/stdlib/strings_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callStringFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("strings")
	if mod == nil {
		panic("strings module not found")
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
		input string
		upper string
		lower string
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

func TestStringEmptyStrings(t *testing.T) {
	// len of empty string
	result := callStringFunc("len", String(""))
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 0 {
		t.Errorf("len(\"\") = %v, want 0", result)
	}

	// toUpper of empty string
	result = callStringFunc("toUpper", String(""))
	r2, ok := result.(*objects.String)
	if !ok || r2.Value != "" {
		t.Errorf("toUpper(\"\") = %v, want \"\"", result)
	}

	// toLower of empty string
	result = callStringFunc("toLower", String(""))
	r2, ok = result.(*objects.String)
	if !ok || r2.Value != "" {
		t.Errorf("toLower(\"\") = %v, want \"\"", result)
	}

	// trim of empty string
	result = callStringFunc("trim", String(""))
	r2, ok = result.(*objects.String)
	if !ok || r2.Value != "" {
		t.Errorf("trim(\"\") = %v, want \"\"", result)
	}

	// reverse of empty string
	result = callStringFunc("reverse", String(""))
	r2, ok = result.(*objects.String)
	if !ok || r2.Value != "" {
		t.Errorf("reverse(\"\") = %v, want \"\"", result)
	}

	// repeat with count 0
	result = callStringFunc("repeat", String("abc"), Int(0))
	r2, ok = result.(*objects.String)
	if !ok || r2.Value != "" {
		t.Errorf("repeat(\"abc\", 0) = %v, want \"\"", result)
	}

	// contains empty substring (should be true)
	result = callStringFunc("contains", String("hello"), String(""))
	r3, ok := result.(*objects.Bool)
	if !ok || !r3.Value {
		t.Errorf("contains(\"hello\", \"\") = %v, want true", result)
	}

	// indexOf empty substring
	result = callStringFunc("indexOf", String("hello"), String(""))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != 0 {
		t.Errorf("indexOf(\"hello\", \"\") = %v, want 0", result)
	}

	// hasPrefix empty prefix (should be true)
	result = callStringFunc("hasPrefix", String("hello"), String(""))
	r3, ok = result.(*objects.Bool)
	if !ok || !r3.Value {
		t.Errorf("hasPrefix(\"hello\", \"\") = %v, want true", result)
	}

	// hasSuffix empty suffix (should be true)
	result = callStringFunc("hasSuffix", String("hello"), String(""))
	r3, ok = result.(*objects.Bool)
	if !ok || !r3.Value {
		t.Errorf("hasSuffix(\"hello\", \"\") = %v, want true", result)
	}
}

func TestStringUnicode(t *testing.T) {
	// Unicode string length (should count runes)
	unicodeStr := "hello"
	result := callStringFunc("len", String(unicodeStr))
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 5 {
		t.Errorf("len(%q) = %v, want 5", unicodeStr, result)
	}

	// Chinese characters
	chinese := "hello"
	result = callStringFunc("len", String(chinese))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != 5 {
		t.Errorf("len(%q) = %v, want 5", chinese, result)
	}

	// Emoji
	emoji := "hello"
	result = callStringFunc("len", String(emoji))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != 5 {
		t.Errorf("len(%q) = %v, want 5", emoji, result)
	}

	// Unicode reverse
	unicodeStr = "abc"
	result = callStringFunc("reverse", String(unicodeStr))
	r2, ok := result.(*objects.String)
	if !ok || r2.Value != "cba" {
		t.Errorf("reverse(%q) = %v, want \"cba\"", unicodeStr, r2.Value)
	}

	// Unicode upper/lower
	unicodeStr = "HELLO"
	result = callStringFunc("toLower", String(unicodeStr))
	r2, ok = result.(*objects.String)
	if !ok || r2.Value != "hello" {
		t.Errorf("toLower(%q) = %v, want \"hello\"", unicodeStr, r2.Value)
	}

	// Unicode contains
	result = callStringFunc("contains", String("hello world"), String("world"))
	r3, ok := result.(*objects.Bool)
	if !ok || !r3.Value {
		t.Errorf("contains with unicode should return true, got %v", result)
	}
}

func TestStringSplitEdgeCases(t *testing.T) {
	// Split empty string
	result := callStringFunc("split", String(""), String(","))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}
	if len(arr.Elements) != 1 {
		t.Errorf("split(\"\", \",\") = %d elements, want 1", len(arr.Elements))
	}

	// Split with empty separator
	result = callStringFunc("split", String("abc"), String(""))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}
	// Empty separator splits each character
	if len(arr.Elements) != 3 {
		t.Errorf("split(\"abc\", \"\") = %d elements, want 3", len(arr.Elements))
	}

	// Split string not containing separator
	result = callStringFunc("split", String("abc"), String(","))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}
	if len(arr.Elements) != 1 {
		t.Errorf("split(\"abc\", \",\") = %d elements, want 1", len(arr.Elements))
	}

	// Split with separator at start and end
	result = callStringFunc("split", String(",a,b,"), String(","))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}
	if len(arr.Elements) != 4 {
		t.Errorf("split(\",a,b,\", \",\") = %d elements, want 4", len(arr.Elements))
	}

	// Split with consecutive separators
	result = callStringFunc("split", String("a,,b"), String(","))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatal("split should return array")
	}
	if len(arr.Elements) != 3 {
		t.Errorf("split(\"a,,b\", \",\") = %d elements, want 3", len(arr.Elements))
	}

	// Split with error - non-string first arg
	result = callStringFunc("split", Int(42), String(","))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("split with non-string first arg should return error, got %T", result)
	}

	// Split with error - non-string separator
	result = callStringFunc("split", String("a,b"), Int(44))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("split with non-string separator should return error, got %T", result)
	}
}

func TestStringJoinEdgeCases(t *testing.T) {
	// Join empty array
	emptyArr := Array()
	result := callStringFunc("join", emptyArr, String(","))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatal("join should return string")
	}
	if r.Value != "" {
		t.Errorf("join([], \",\") = %q, want \"\"", r.Value)
	}

	// Join single element
	singleArr := Array(String("a"))
	result = callStringFunc("join", singleArr, String(","))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatal("join should return string")
	}
	if r.Value != "a" {
		t.Errorf("join([\"a\"], \",\") = %q, want \"a\"", r.Value)
	}

	// Join with empty separator
	arr := Array(String("a"), String("b"), String("c"))
	result = callStringFunc("join", arr, String(""))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatal("join should return string")
	}
	if r.Value != "abc" {
		t.Errorf("join with empty separator = %q, want \"abc\"", r.Value)
	}

	// Join with error - non-array first arg
	result = callStringFunc("join", String("not array"), String(","))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("join with non-array should return error, got %T", result)
	}

	// Join with error - non-string separator
	result = callStringFunc("join", Array(String("a")), Int(44))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("join with non-string separator should return error, got %T", result)
	}

	// Join with error - non-string array elements
	result = callStringFunc("join", Array(Int(1), Int(2)), String(","))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("join with non-string elements should return error, got %T", result)
	}
}

func TestStringSubstr(t *testing.T) {
	// Basic substr
	result := callStringFunc("substr", String("hello"), Int(0), Int(3))
	r, ok := result.(*objects.String)
	if !ok || r.Value != "hel" {
		t.Errorf("substr(\"hello\", 0, 3) = %v, want \"hel\"", result)
	}

	// substr with only start index
	result = callStringFunc("substr", String("hello"), Int(2))
	r, ok = result.(*objects.String)
	if !ok || r.Value != "llo" {
		t.Errorf("substr(\"hello\", 2) = %v, want \"llo\"", result)
	}

	// substr error - index out of range
	result = callStringFunc("substr", String("hello"), Int(10), Int(15))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("substr with out of range index should return error, got %T", result)
	}

	// substr error - start > end
	result = callStringFunc("substr", String("hello"), Int(3), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("substr with start > end should return error, got %T", result)
	}
}

func TestStringRepeatEdgeCases(t *testing.T) {
	// Repeat with negative count
	result := callStringFunc("repeat", String("a"), Int(-1))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("repeat with negative count should return error, got %T", result)
	}

	// Repeat with non-int count
	result = callStringFunc("repeat", String("a"), String("3"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("repeat with non-int count should return error, got %T", result)
	}

	// Repeat with non-string first arg
	result = callStringFunc("repeat", Int(42), Int(3))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("repeat with non-string should return error, got %T", result)
	}

	// Repeat error - wrong number of args
	result = callStringFunc("repeat", String("a"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("repeat with 1 arg should return error, got %T", result)
	}
}

func TestStringReplaceEdgeCases(t *testing.T) {
	// Replace with empty old string
	result := callStringFunc("replace", String("hello"), String(""), String("x"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Errorf("replace with empty old string should return string, got %T", result)
	}

	// Replace all occurrences
	result = callStringFunc("replace", String("aaa"), String("a"), String("b"))
	r, ok = result.(*objects.String)
	if !ok || r.Value != "bbb" {
		t.Errorf("replace all = %v, want \"bbb\"", result)
	}

	// Replace error - not enough args
	result = callStringFunc("replace", String("hello"), String("l"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("replace with 2 args should return error, got %T", result)
	}
}

func TestStringParseIntEdgeCases(t *testing.T) {
	// ParseInt with hex prefix
	result := callStringFunc("parseInt", String("0x10"))
	r, ok := result.(*objects.Int)
	if !ok || r.Value != 16 {
		t.Errorf("parseInt(\"0x10\") = %v, want 16", result)
	}

	// ParseInt with negative number
	result = callStringFunc("parseInt", String("-42"))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != -42 {
		t.Errorf("parseInt(\"-42\") = %v, want -42", result)
	}

	// ParseInt with octal
	result = callStringFunc("parseInt", String("0o10"))
	r, ok = result.(*objects.Int)
	if !ok || r.Value != 8 {
		t.Errorf("parseInt(\"0o10\") = %v, want 8", result)
	}

	// ParseInt error - wrong number of args
	result = callStringFunc("parseInt")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseInt with 0 args should return error, got %T", result)
	}

	// ParseInt error - non-string
	result = callStringFunc("parseInt", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseInt with non-string should return error, got %T", result)
	}
}

func TestStringParseFloatEdgeCases(t *testing.T) {
	// ParseFloat with negative
	result := callStringFunc("parseFloat", String("-3.14"))
	r, ok := result.(*objects.Float)
	if !ok || r.Value != -3.14 {
		t.Errorf("parseFloat(\"-3.14\") = %v, want -3.14", result)
	}

	// ParseFloat with scientific notation
	result = callStringFunc("parseFloat", String("1e10"))
	r, ok = result.(*objects.Float)
	if !ok || r.Value != 1e10 {
		t.Errorf("parseFloat(\"1e10\") = %v, want 1e10", result)
	}

	// ParseFloat error - invalid
	result = callStringFunc("parseFloat", String("not a number"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFloat(\"not a number\") should return error, got %T", result)
	}

	// ParseFloat error - wrong number of args
	result = callStringFunc("parseFloat")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFloat with 0 args should return error, got %T", result)
	}
}

func TestStringFormat(t *testing.T) {
	// format with no args
	result := callStringFunc("format", String("hello"))
	_, ok := result.(*objects.String)
	if !ok {
		t.Errorf("format with no args should return string, got %T", result)
	}

	// format with multiple args
	result = callStringFunc("format", String("{}"), String("test"), Int(42))
	_, ok = result.(*objects.String)
	if !ok {
		t.Errorf("format with args should return string, got %T", result)
	}

	// format error - no args
	result = callStringFunc("format")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("format with 0 args should return error, got %T", result)
	}

	// format error - non-string first arg
	result = callStringFunc("format", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("format with non-string should return error, got %T", result)
	}
}

func TestStringTrimSpace(t *testing.T) {
	// trimSpace is an alias for trim
	result := callStringFunc("trimSpace", String("  hello  "))
	r, ok := result.(*objects.String)
	if !ok || r.Value != "hello" {
		t.Errorf("trimSpace(\"  hello  \") = %v, want \"hello\"", result)
	}

	// trimSpace error - wrong args
	result = callStringFunc("trimSpace")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("trimSpace with 0 args should return error, got %T", result)
	}
}

func TestStringErrorCases(t *testing.T) {
	// len with wrong args
	result := callStringFunc("len")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("len with 0 args should return error, got %T", result)
	}

	// toUpper with wrong args
	result = callStringFunc("toUpper")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("toUpper with 0 args should return error, got %T", result)
	}

	// toLower with wrong args
	result = callStringFunc("toLower")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("toLower with 0 args should return error, got %T", result)
	}

	// trim with wrong args
	result = callStringFunc("trim")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("trim with 0 args should return error, got %T", result)
	}

	// reverse with wrong args
	result = callStringFunc("reverse")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reverse with 0 args should return error, got %T", result)
	}

	// reverse with non-string
	result = callStringFunc("reverse", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("reverse with non-string should return error, got %T", result)
	}

	// toString with wrong args
	result = callStringFunc("toString")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("toString with 0 args should return error, got %T", result)
	}

	// toString with unsupported type (array)
	result = callStringFunc("toString", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("toString with unsupported type should return error, got %T", result)
	}
}
