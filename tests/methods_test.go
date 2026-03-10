// tests/methods_test.go
// Tests for the unified object system - method calls on primitive types
package tests

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// ============================================
// Universal Methods Tests
// ============================================

func TestUniversalTypeOf(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Int
		{"42.typeOf()", "INT"},
		{"(-5).typeOf()", "INT"},
		// Float
		{"3.14.typeOf()", "FLOAT"},
		{"(-2.5).typeOf()", "FLOAT"},
		// String
		{`"hello".typeOf()`, "STRING"},
		{`"".typeOf()`, "STRING"},
		// Array
		{"[1, 2, 3].typeOf()", "ARRAY"},
		{"[].typeOf()", "ARRAY"},
		// Map
		{`{"a": 1}.typeOf()`, "MAP"},
		{"{}.typeOf()", "MAP"},
		// Bool
		{"true.typeOf()", "BOOL"},
		{"false.typeOf()", "BOOL"},
		// Null
		{"null.typeOf()", "NULL"},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertString(t, result, tt.expected)
	}
}

func TestUniversalToStr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Int
		{"42.toStr()", "42"},
		{"(-5).toStr()", "-5"},
		// Float
		{"3.14.toStr()", "3.14"},
		// String
		{`"hello".toStr()`, "hello"},
		// Bool
		{"true.toStr()", "true"},
		{"false.toStr()", "false"},
		// Null
		{"null.toStr()", "null"},
		// Array
		{"[1, 2, 3].toStr()", "[1, 2, 3]"},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertString(t, result, tt.expected)
	}
}

// ============================================
// Int Method Tests
// ============================================

func TestIntMethods(t *testing.T) {
	t.Run("toFloat", func(t *testing.T) {
		tests := []struct {
			input    string
			expected float64
		}{
			{"42.toFloat()", 42.0},
			{"(-5).toFloat()", -5.0},
			{"0.toFloat()", 0.0},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertFloat(t, result, tt.expected)
		}
	})

	t.Run("abs", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"42.abs()", 42},
			{"(-42).abs()", 42},
			{"0.abs()", 0},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})
}

// ============================================
// Float Method Tests
// ============================================

func TestFloatMethods(t *testing.T) {
	t.Run("toInt", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"3.14.toInt()", 3},
			{"(-2.7).toInt()", -2},
			{"42.0.toInt()", 42},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("abs", func(t *testing.T) {
		tests := []struct {
			input    string
			expected float64
		}{
			{"3.14.abs()", 3.14},
			{"(-3.14).abs()", 3.14},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertFloat(t, result, tt.expected)
		}
	})

	t.Run("floor", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"3.7.floor()", 3},
			{"(-3.7).floor()", -4},
			{"3.0.floor()", 3},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("ceil", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"3.2.ceil()", 4},
			{"(-3.2).ceil()", -3},
			{"3.0.ceil()", 3},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("round", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"3.4.round()", 3},
			{"3.6.round()", 4},
			{"3.5.round()", 4},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})
}

// ============================================
// String Method Tests
// ============================================

func TestStringMethods(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{`"hello".len()`, 5},
			{`"".len()`, 0},
			{`"a b c".len()`, 5},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("upper and lower", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`"hello".upper()`, "HELLO"},
			{`"HELLO".lower()`, "hello"},
			{`"HeLLo".upper()`, "HELLO"},
			{`"HeLLo".lower()`, "hello"},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertString(t, result, tt.expected)
		}
	})

	t.Run("trim", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`"  hello  ".trim()`, "hello"},
			{`"\t\nhello\n\t".trim()`, "hello"},
			{`"hello".trim()`, "hello"},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertString(t, result, tt.expected)
		}
	})

	t.Run("split", func(t *testing.T) {
		result := runCode(t, `"a,b,c".split(",")`)
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		assertString(t, arr.Elements[0], "a")
		assertString(t, arr.Elements[1], "b")
		assertString(t, arr.Elements[2], "c")
	})

	t.Run("contains", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{`"hello".contains("ell")`, true},
			{`"hello".contains("xyz")`, false},
			{`"hello".contains("hello")`, true},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertBool(t, result, tt.expected)
		}
	})

	t.Run("indexOf", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{`"hello".indexOf("e")`, 1},
			{`"hello".indexOf("l")`, 2},
			{`"hello".indexOf("xyz")`, -1},
			{`"hello".indexOf("hello")`, 0},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("startsWith and endsWith", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{`"hello".startsWith("he")`, true},
			{`"hello".startsWith("lo")`, false},
			{`"hello".endsWith("lo")`, true},
			{`"hello".endsWith("he")`, false},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertBool(t, result, tt.expected)
		}
	})

	t.Run("toInt and toFloat", func(t *testing.T) {
		t.Run("toInt", func(t *testing.T) {
			result := runCode(t, `"42".toInt()`)
			assertInt(t, result, 42)
		})
		t.Run("toFloat", func(t *testing.T) {
			result := runCode(t, `"3.14".toFloat()`)
			assertFloat(t, result, 3.14)
		})
	})
}

// ============================================
// Array Method Tests
// ============================================

func TestArrayMethods(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"[1, 2, 3].len()", 3},
			{"[].len()", 0},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("push", func(t *testing.T) {
		result := runCode(t, "[1, 2].push(3)")
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[2], 3)
	})

	t.Run("pop", func(t *testing.T) {
		result := runCode(t, "[1, 2, 3].pop()")
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
		}
		assertInt(t, arr.LastPopped, 3)
	})

	t.Run("first and last", func(t *testing.T) {
		t.Run("first", func(t *testing.T) {
			result := runCode(t, "[1, 2, 3].first()")
			assertInt(t, result, 1)
		})
		t.Run("last", func(t *testing.T) {
			result := runCode(t, "[1, 2, 3].last()")
			assertInt(t, result, 3)
		})
		t.Run("first empty", func(t *testing.T) {
			result := runCode(t, "[].first()")
			assertNull(t, result)
		})
		t.Run("last empty", func(t *testing.T) {
			result := runCode(t, "[].last()")
			assertNull(t, result)
		})
	})

	t.Run("indexOf", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"[1, 2, 3].indexOf(2)", 1},
			{"[1, 2, 3].indexOf(4)", -1},
			{`["a", "b", "c"].indexOf("b")`, 1},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("contains", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"[1, 2, 3].contains(2)", true},
			{"[1, 2, 3].contains(4)", false},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertBool(t, result, tt.expected)
		}
	})

	t.Run("reverse", func(t *testing.T) {
		result := runCode(t, "[1, 2, 3].reverse()")
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[0], 3)
		assertInt(t, arr.Elements[1], 2)
		assertInt(t, arr.Elements[2], 1)
	})

	t.Run("join", func(t *testing.T) {
		result := runCode(t, `["a", "b", "c"].join("-")`)
		assertString(t, result, "a-b-c")
	})
}

// ============================================
// Map Method Tests
// ============================================

func TestMapMethods(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{`{"a": 1, "b": 2}.len()`, 2},
			{"{}.len()", 0},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		}
	})

	t.Run("keys", func(t *testing.T) {
		result := runCode(t, `{"a": 1, "b": 2}.keys()`)
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
		}
		// Keys can be in any order
		keys := make(map[string]bool)
		for _, elem := range arr.Elements {
			if s, ok := elem.(*objects.String); ok {
				keys[s.Value] = true
			}
		}
		if !keys["a"] || !keys["b"] {
			t.Fatalf("expected keys 'a' and 'b', got %v", keys)
		}
	})

	t.Run("values", func(t *testing.T) {
		result := runCode(t, `{"a": 1, "b": 2}.values()`)
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
		}
	})

	t.Run("hasKey", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{`{"a": 1}.hasKey("a")`, true},
			{`{"a": 1}.hasKey("b")`, false},
		}
		for _, tt := range tests {
			result := runCode(t, tt.input)
			assertBool(t, result, tt.expected)
		}
	})

	t.Run("delete", func(t *testing.T) {
		result := runCode(t, `{"a": 1, "b": 2}.delete("a")`)
		m, ok := result.(*objects.Map)
		if !ok {
			t.Fatalf("expected Map, got %T", result)
		}
		if len(m.Pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(m.Pairs))
		}
	})
}

// ============================================
// Method Chaining Tests
// ============================================

func TestMethodChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"  hello  ".trim().upper()`, "HELLO"},
		{`"HELLO".lower().trim()`, "hello"},
		{`"a,b,c".split(",").join("-")`, "a-b-c"},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertString(t, result, tt.expected)
	}
}

func TestMethodChainingWithTypes(t *testing.T) {
	// Test chaining that changes types
	t.Run("string to array to string", func(t *testing.T) {
		result := runCode(t, `"a,b,c".split(",").join("-")`)
		assertString(t, result, "a-b-c")
	})

	t.Run("int to float to int", func(t *testing.T) {
		result := runCode(t, "42.toFloat().toInt()")
		assertInt(t, result, 42)
	})
}

// ============================================
// Method with Variable Tests
// ============================================

func TestMethodsWithVariables(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`var s = "hello"; s.len()`, int64(5)},
		{"var arr = [1, 2, 3]; arr.push(4).len()", int64(4)},
		{"var x = 42; x.typeOf()", "INT"},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			assertInt(t, result, expected)
		case string:
			assertString(t, result, expected)
		}
	}
}

// ============================================
// Error Cases
// ============================================

func TestMethodNotFound(t *testing.T) {
	// These should return errors from the VM
	// We test that they don't crash and return an error
	tests := []string{
		"42.unknownMethod()",
		`"hello".unknownMethod()`,
		"[1, 2, 3].unknownMethod()",
	}

	for _, input := range tests {
		// Create a helper that expects the VM to return an error
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			// Parser error is acceptable for unknown methods
			continue
		}

		c := compiler.New()
		if err := c.Compile(program); err != nil {
			// Compiler error is acceptable
			continue
		}

		bytecode := c.Bytecode()
		v := vm.New(bytecode)

		err := v.Run()
		// We expect an error for unknown methods
		if err == nil {
			t.Fatalf("expected error for unknown method in: %s", input)
		}
	}
}

// ============================================
// Interaction with Map Property Access
// ============================================

func TestMapPropertyAccessPrecedence(t *testing.T) {
	// Map property access should take precedence over methods
	// If a map has a key "len", accessing m.len should return that value, not the method
	t.Run("map key shadows method name", func(t *testing.T) {
		result := runCode(t, `{"len": 42}.len`)
		assertInt(t, result, 42)
	})

	t.Run("method still works when key doesn't exist", func(t *testing.T) {
		result := runCode(t, `{"a": 1}.len()`)
		assertInt(t, result, 1)
	})
}
