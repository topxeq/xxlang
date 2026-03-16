// pkg/stdlib/text_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestTextModuleExists(t *testing.T) {
	mod := Get("text")
	if mod == nil {
		t.Fatal("text module not found")
	}
}

func TestWordWrap(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["wordWrap"].(*objects.Builtin)

	result := fn.Fn(String("hello world test"), Int(10))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Should wrap at "hello world" (11 chars) to fit within 10
	// Result should have at least one newline
	if s.Value == "hello world test" {
		t.Error("wordWrap should have wrapped the text")
	}
}

func TestTruncate(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["truncate"].(*objects.Builtin)

	tests := []struct {
		args     []objects.Object
		expected string
	}{
		{[]objects.Object{String("hello world"), Int(8)}, "hello..."},
		{[]objects.Object{String("short"), Int(10)}, "short"},
		{[]objects.Object{String("hello world"), Int(10), String("!")}, "hello wor!"},
	}

	for _, tt := range tests {
		result := fn.Fn(tt.args...)
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("truncate(%v) = %q, want %q", tt.args, s.Value, tt.expected)
		}
	}
}

func TestWordCount(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["wordCount"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected int64
	}{
		{"hello world", 2},
		{"one two three four", 4},
		{"", 0},
		{"   ", 0},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("wordCount(%q) = %d, want %d", tt.input, i.Value, tt.expected)
		}
	}
}

func TestLineCount(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["lineCount"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"one line", 1},
		{"line1\nline2", 2},
		{"line1\nline2\nline3", 3},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("lineCount(%q) = %d, want %d", tt.input, i.Value, tt.expected)
		}
	}
}

func TestCharCount(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["charCount"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected int64
	}{
		{"hello", 5},
		{"", 0},
		{"日本語", 3}, // Unicode test
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		i, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if i.Value != tt.expected {
			t.Errorf("charCount(%q) = %d, want %d", tt.input, i.Value, tt.expected)
		}
	}
}

func TestByteCount(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["byteCount"].(*objects.Builtin)

	result := fn.Fn(String("hello"))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if i.Value != 5 {
		t.Errorf("byteCount(\"hello\") = %d, want 5", i.Value)
	}
}

func TestLines(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["lines"].(*objects.Builtin)

	result := fn.Fn(String("line1\nline2\nline3"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 3 {
		t.Errorf("lines() returned %d elements, want 3", len(arr.Elements))
	}
}

func TestJoinLines(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["joinLines"].(*objects.Builtin)

	lines := Array(String("line1"), String("line2"), String("line3"))
	result := fn.Fn(lines)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	expected := "line1\nline2\nline3"
	if s.Value != expected {
		t.Errorf("joinLines() = %q, want %q", s.Value, expected)
	}
}

func TestWords(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["words"].(*objects.Builtin)

	result := fn.Fn(String("hello world test"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 3 {
		t.Errorf("words() returned %d elements, want 3", len(arr.Elements))
	}
}

func TestChars(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["chars"].(*objects.Builtin)

	result := fn.Fn(String("abc"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 3 {
		t.Errorf("chars() returned %d elements, want 3", len(arr.Elements))
	}
}

func TestTitle(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["title"].(*objects.Builtin)

	result := fn.Fn(String("hello world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "Hello World" {
		t.Errorf("title() = %q, want %q", s.Value, "Hello World")
	}
}

func TestCapitalize(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["capitalize"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"", ""},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("capitalize(%q) = %q, want %q", tt.input, s.Value, tt.expected)
		}
	}
}

func TestSwapCase(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["swapCase"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "hELLO"},
		{"ABC", "abc"},
		{"abc", "ABC"},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != tt.expected {
			t.Errorf("swapCase(%q) = %q, want %q", tt.input, s.Value, tt.expected)
		}
	}
}

func TestIsAlphaNum(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["isAlphaNum"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"abc123", true},
		{"hello", true},
		{"hello world", false}, // has space
		{"", false},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isAlphaNum(%q) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestIsAlpha(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["isAlpha"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"HelloWorld", true},
		{"hello123", false},
		{"", false},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isAlpha(%q) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["isNumeric"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"42", true},
		{"12.3", false}, // dot is not a digit
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isNumeric(%q) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestIsSpace(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["isSpace"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"   ", true},
		{"\t\n", true},
		{"  x  ", false},
		{"", true}, // empty string has no non-space chars
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isSpace(%q) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestIsBlank(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["isBlank"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"  hello  ", false},
	}

	for _, tt := range tests {
		result := fn.Fn(String(tt.input))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isBlank(%q) = %v, want %v", tt.input, b.Value, tt.expected)
		}
	}
}

func TestRemoveSpaces(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["removeSpaces"].(*objects.Builtin)

	result := fn.Fn(String("hello world test"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "helloworldtest" {
		t.Errorf("removeSpaces() = %q, want %q", s.Value, "helloworldtest")
	}
}

func TestNormalizeSpace(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["normalizeSpace"].(*objects.Builtin)

	result := fn.Fn(String("hello    world\t\ttest"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "hello world test" {
		t.Errorf("normalizeSpace() = %q, want %q", s.Value, "hello world test")
	}
}

func TestPadLeft(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["padLeft"].(*objects.Builtin)

	result := fn.Fn(String("42"), Int(5), String("0"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "00042" {
		t.Errorf("padLeft() = %q, want %q", s.Value, "00042")
	}
}

func TestPadRight(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["padRight"].(*objects.Builtin)

	result := fn.Fn(String("42"), Int(5), String("."))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "42..." {
		t.Errorf("padRight() = %q, want %q", s.Value, "42...")
	}
}

func TestIndent(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["indent"].(*objects.Builtin)

	result := fn.Fn(String("line1\nline2"), String("  "))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	expected := "  line1\n  line2"
	if s.Value != expected {
		t.Errorf("indent() = %q, want %q", s.Value, expected)
	}
}

func TestDedent(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["dedent"].(*objects.Builtin)

	result := fn.Fn(String("  line1\n  line2\n    line3"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Should remove minimum indent (2 spaces)
	if s.Value == "" {
		t.Error("dedent returned empty string")
	}
}

func TestCenterText(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["centerText"].(*objects.Builtin)

	result := fn.Fn(String("hi"), Int(6))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Should center "hi" in 6 chars
	if len(s.Value) != 6 {
		t.Errorf("centerText() length = %d, want 6", len(s.Value))
	}
}

func TestRepeat(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["repeat"].(*objects.Builtin)

	result := fn.Fn(String("ab"), Int(3))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "ababab" {
		t.Errorf("repeat() = %q, want %q", s.Value, "ababab")
	}
}

func TestCharAt(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["charAt"].(*objects.Builtin)

	result := fn.Fn(String("hello"), Int(1))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "e" {
		t.Errorf("charAt() = %q, want %q", s.Value, "e")
	}
}

func TestCharCode(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["charCode"].(*objects.Builtin)

	result := fn.Fn(String("A"), Int(0))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if i.Value != 65 { // ASCII code for 'A'
		t.Errorf("charCode() = %d, want 65", i.Value)
	}
}

func TestFromCode(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["fromCode"].(*objects.Builtin)

	result := fn.Fn(Int(65))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "A" {
		t.Errorf("fromCode(65) = %q, want %q", s.Value, "A")
	}
}

func TestShellEscape(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["shellEscape"].(*objects.Builtin)

	result := fn.Fn(String("hello world"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Should be wrapped in single quotes
	if s.Value != `'hello world'` {
		t.Errorf("shellEscape() = %q, want %q", s.Value, `'hello world'`)
	}
}

func TestJsonEscape(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["jsonEscape"].(*objects.Builtin)

	result := fn.Fn(String(`hello "world"`))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Should escape quotes
	if s.Value == "" {
		t.Error("jsonEscape returned empty string")
	}
}

func TestJsonUnescape(t *testing.T) {
	mod := Get("text")
	fn := mod.Exports["jsonUnescape"].(*objects.Builtin)

	result := fn.Fn(String(`hello\nworld`))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	expected := "hello\nworld"
	if s.Value != expected {
		t.Errorf("jsonUnescape() = %q, want %q", s.Value, expected)
	}
}

func TestTextErrors(t *testing.T) {
	mod := Get("text")

	tests := []struct {
		name   string
		fnName string
		args   []objects.Object
	}{
		{"wordWrap no args", "wordWrap", []objects.Object{}},
		{"wordWrap wrong first type", "wordWrap", []objects.Object{Int(42), Int(10)}},
		{"wordWrap wrong second type", "wordWrap", []objects.Object{String("hello"), String("10")}},
		{"truncate no args", "truncate", []objects.Object{}},
		{"truncate wrong first type", "truncate", []objects.Object{Int(42), Int(10)}},
		{"truncate wrong second type", "truncate", []objects.Object{String("hello"), String("10")}},
		{"wordCount no args", "wordCount", []objects.Object{}},
		{"wordCount wrong type", "wordCount", []objects.Object{Int(42)}},
		{"wordCount wrong arg count", "wordCount", []objects.Object{String("hello"), String("world")}},
		{"lineCount no args", "lineCount", []objects.Object{}},
		{"lineCount wrong type", "lineCount", []objects.Object{Int(42)}},
		{"charCount no args", "charCount", []objects.Object{}},
		{"byteCount no args", "byteCount", []objects.Object{}},
		{"lines no args", "lines", []objects.Object{}},
		{"joinLines no args", "joinLines", []objects.Object{}},
		{"joinLines wrong type", "joinLines", []objects.Object{String("hello")}},
		{"words no args", "words", []objects.Object{}},
		{"chars no args", "chars", []objects.Object{}},
		{"title no args", "title", []objects.Object{}},
		{"capitalize no args", "capitalize", []objects.Object{}},
		{"swapCase no args", "swapCase", []objects.Object{}},
		{"isAlphaNum no args", "isAlphaNum", []objects.Object{}},
		{"isAlpha no args", "isAlpha", []objects.Object{}},
		{"isNumeric no args", "isNumeric", []objects.Object{}},
		{"isSpace no args", "isSpace", []objects.Object{}},
		{"isBlank no args", "isBlank", []objects.Object{}},
		{"removeSpaces no args", "removeSpaces", []objects.Object{}},
		{"normalizeSpace no args", "normalizeSpace", []objects.Object{}},
		{"padLeft no args", "padLeft", []objects.Object{}},
		{"padRight no args", "padRight", []objects.Object{}},
		{"indent no args", "indent", []objects.Object{}},
		{"dedent no args", "dedent", []objects.Object{}},
		{"centerText no args", "centerText", []objects.Object{}},
		{"repeat no args", "repeat", []objects.Object{}},
		{"repeat negative count", "repeat", []objects.Object{String("a"), Int(-1)}},
		{"charAt no args", "charAt", []objects.Object{}},
		{"charAt out of range", "charAt", []objects.Object{String("hi"), Int(10)}},
		{"charCode no args", "charCode", []objects.Object{}},
		{"fromCode no args", "fromCode", []objects.Object{}},
		{"shellEscape no args", "shellEscape", []objects.Object{}},
		{"jsonEscape no args", "jsonEscape", []objects.Object{}},
		{"jsonUnescape no args", "jsonUnescape", []objects.Object{}},
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
