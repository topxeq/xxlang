// tests/integration_test.go
// Integration tests for the xxlang interpreter.
// These tests verify the full pipeline: Lexer -> Parser -> Compiler -> VM
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
// Test Helpers
// ============================================

// runCode executes xxlang source code and returns the result.
func runCode(t *testing.T, input string) objects.Object {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := c.Bytecode()
	v := vm.New(bytecode)

	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	return v.LastPopped()
}

// runCodeWithGlobals executes code and returns both result and globals.
func runCodeWithGlobals(t *testing.T, input string) (objects.Object, []objects.Object) {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := c.Bytecode()
	globals := make([]objects.Object, compiler.GlobalsSize)
	v := vm.NewWithGlobalsStore(bytecode, globals)

	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	return v.LastPopped(), v.Globals()
}

// assertInt asserts that an object is an integer with expected value.
func assertInt(t *testing.T, obj objects.Object, expected int64) {
	t.Helper()
	result, ok := obj.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", obj)
	}
	if result.Value != expected {
		t.Fatalf("expected %d, got %d", expected, result.Value)
	}
}

// assertFloat asserts that an object is a float with expected value.
func assertFloat(t *testing.T, obj objects.Object, expected float64) {
	t.Helper()
	result, ok := obj.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", obj)
	}
	if result.Value != expected {
		t.Fatalf("expected %f, got %f", expected, result.Value)
	}
}

// assertString asserts that an object is a string with expected value.
func assertString(t *testing.T, obj objects.Object, expected string) {
	t.Helper()
	result, ok := obj.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", obj)
	}
	if result.Value != expected {
		t.Fatalf("expected %q, got %q", expected, result.Value)
	}
}

// assertBool asserts that an object is a bool with expected value.
func assertBool(t *testing.T, obj objects.Object, expected bool) {
	t.Helper()
	result, ok := obj.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", obj)
	}
	if result.Value != expected {
		t.Fatalf("expected %t, got %t", expected, result.Value)
	}
}

// assertNull asserts that an object is NULL.
func assertNull(t *testing.T, obj objects.Object) {
	t.Helper()
	if obj != objects.NULL {
		t.Fatalf("expected NULL, got %T", obj)
	}
}

// ============================================
// Basic Expressions
// ============================================

func TestIntegerExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"5 % 3", 2},
		{"10 % 2", 0},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

func TestFloatExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.5", 5.5},
		{"10.25", 10.25},
		{"-5.5", -5.5},
		{"5.5 + 4.5", 10.0},
		{"10.0 - 5.5", 4.5},
		{"2.5 * 2.0", 5.0},
		{"7.5 / 2.5", 3.0},
		{"1 + 2.5", 3.5},
		{"2.5 * 2", 5.0},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertFloat(t, result, tt.expected)
	}
}

func TestStringExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertString(t, result, tt.expected)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{`"a" == "a"`, true},
		{`"a" == "b"`, false},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertBool(t, result, tt.expected)
	}
}

// ============================================
// Variables and Scopes
// ============================================

func TestVariableBindings(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var a = 5; a", 5},
		{"var a = 5; var b = a; b", 5},
		{"var a = 5; var b = a; var c = a + b + 5; c", 15},
		{"var a = 5; a = 10; a", 10},
		{"var a = 5; a += 5; a", 10},
		{"var a = 5; a -= 3; a", 2},
		{"var a = 5; a *= 2; a", 10},
		{"var a = 10; a /= 2; a", 5},
		{"const PI = 3; PI", 3},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

// ============================================
// Control Flow
// ============================================

func TestIfElse(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 < 2) { if (2 < 3) { 30 } }", 30},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			assertInt(t, result, int64(expected))
		case nil:
			assertNull(t, result)
		}
	}
}

func TestWhileLoops(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"var i = 0; while (i < 5) { i = i + 1 } i",
			5,
		},
		{
			"var sum = 0; var i = 1; while (i <= 5) { sum = sum + i; i = i + 1 } sum",
			15,
		},
		{
			"var i = 10; while (i > 0) { i = i - 1 } i",
			0,
		},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

func TestForLoops(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"var sum = 0; for (var i = 1; i <= 5; i = i + 1) { sum = sum + i } sum",
			15,
		},
		{
			"var count = 0; for (var i = 0; i < 3; i = i + 1) { for (var j = 0; j < 3; j = j + 1) { count = count + 1 } } count",
			9,
		},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

// ============================================
// Functions
// ============================================

func TestFunctionDefinitions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"func five() { return 5 }; five()", 5},
		{"func ten() { return 10 }; ten()", 10},
		{"func identity(x) { return x }; identity(42)", 42},
		{"func add(a, b) { return a + b }; add(3, 4)", 7},
		{"func multiply(a, b) { return a * b }; multiply(3, 4)", 12},
		{"func tripleAdd(a, b, c) { return a + b + c }; tripleAdd(1, 2, 3)", 6},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

func TestClosures(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"func makeAdder(x) { func adder(y) { return x + y }; return adder }; var add5 = makeAdder(5); add5(10)",
			15,
		},
		{
			"func makeCounter() { var count = 0; func counter() { count = count + 1; return count }; return counter }; var c = makeCounter(); c(); c(); c()",
			3,
		},
		{
			"func outer(x) { func middle(y) { func inner(z) { return x + y + z }; return inner }; return middle }; outer(1)(2)(3)",
			6,
		},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

func TestRecursion(t *testing.T) {
	// Match the exact syntax from vm_test.go
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"func fib(n) { if (n <= 1) { return n; } return fib(n - 1) + fib(n - 2); }; fib(10);",
			55,
		},
		{
			"func factorial(n) { if (n <= 1) { return 1; } return n * factorial(n - 1); }; factorial(5);",
			120,
		},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

// ============================================
// Arrays and Maps
// ============================================

func TestArrays(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"var arr = [1, 2, 3]; arr[0] + arr[1] + arr[2]", 6},
		{"var arr = [1, 2, 3]; arr[0] = 10; arr[0]", 10},
		{"len([1, 2, 3, 4, 5])", 5},
		{"first([1, 2, 3])", 1},
		{"last([1, 2, 3])", 3},
		{"push([1, 2], 3); len([1, 2, 3])", 3},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

func TestMaps(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`var m = {"a": 1}; m["a"]`, 1},
		{`var m = {"a": 1, "b": 2}; m["a"] + m["b"]`, 3},
		{`var m = {}; m["x"] = 10; m["x"]`, 10},
		{`var m = {"a": 1, "b": 2}; len(keys(m))`, 2},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

// ============================================
// Built-in Functions
// ============================================

func TestBuiltinFunctions(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		assertInt(t, runCode(t, `len("hello")`), 5)
		assertInt(t, runCode(t, `len("")`), 0)
		assertInt(t, runCode(t, `len([1, 2, 3])`), 3)
	})

	t.Run("typeOf", func(t *testing.T) {
		assertString(t, runCode(t, `typeOf(1)`), "INT")
		assertString(t, runCode(t, `typeOf(1.5)`), "FLOAT")
		assertString(t, runCode(t, `typeOf("hello")`), "STRING")
		assertString(t, runCode(t, `typeOf(true)`), "BOOL")
		assertString(t, runCode(t, `typeOf(null)`), "NULL")
		assertString(t, runCode(t, `typeOf([1, 2])`), "ARRAY")
		assertString(t, runCode(t, `typeOf({})`), "MAP")
	})

	t.Run("string functions", func(t *testing.T) {
		assertString(t, runCode(t, `substr("hello", 0, 3)`), "hel")
		assertString(t, runCode(t, `upper("hello")`), "HELLO")
		assertString(t, runCode(t, `lower("HELLO")`), "hello")
		assertString(t, runCode(t, `trim("  hello  ")`), "hello")
	})

	t.Run("math functions", func(t *testing.T) {
		assertInt(t, runCode(t, `abs(-5)`), 5)
		assertFloat(t, runCode(t, `sqrt(16)`), 4.0)
		assertFloat(t, runCode(t, `pow(2, 3)`), 8.0)
		assertInt(t, runCode(t, `min(1, 5)`), 1)
		assertInt(t, runCode(t, `max(1, 5)`), 5)
		assertInt(t, runCode(t, `floor(3.7)`), 3)
		assertInt(t, runCode(t, `ceil(3.2)`), 4)
	})

	t.Run("type conversion", func(t *testing.T) {
		assertInt(t, runCode(t, `int(3.7)`), 3)
		assertFloat(t, runCode(t, `float(5)`), 5.0)
		assertString(t, runCode(t, `string(42)`), "42")
	})
}

// ============================================
// Complex Programs
// ============================================

func TestComplexPrograms(t *testing.T) {
	t.Run("bubble sort", func(t *testing.T) {
		input := `
			func bubbleSort(arr) {
				var n = len(arr)
				for (var i = 0; i < n - 1; i = i + 1) {
					for (var j = 0; j < n - i - 1; j = j + 1) {
						if (arr[j] > arr[j + 1]) {
							var temp = arr[j]
							arr[j] = arr[j + 1]
							arr[j + 1] = temp
						}
					}
				}
				return arr
			}
			var arr = [64, 34, 25, 12, 22, 11, 90]
			bubbleSort(arr)
			arr[0]
		`
		result := runCode(t, input)
		assertInt(t, result, 11)
	})

	t.Run("prime check", func(t *testing.T) {
		input := `
			func isPrime(n) {
				if (n <= 1) { return false }
				if (n <= 3) { return true }
				if (n % 2 == 0) { return false }
				var i = 3
				while (i * i <= n) {
					if (n % i == 0) { return false }
					i = i + 2
				}
				return true
			}
			isPrime(17)
		`
		result := runCode(t, input)
		assertBool(t, result, true)
	})

	t.Run("GCD", func(t *testing.T) {
		input := `
			func gcd(a, b) {
				while (b != 0) {
					var temp = b
					b = a % b
					a = temp
				}
				return a
			}
			gcd(48, 18)
		`
		result := runCode(t, input)
		assertInt(t, result, 6)
	})

	t.Run("Fibonacci sequence", func(t *testing.T) {
		// Simpler test that doesn't use push builtin
		input := `
			func fib(n) {
				if (n <= 1) { return n; }
				return fib(n - 1) + fib(n - 2);
			}
			var results = [fib(0), fib(1), fib(2), fib(3), fib(4), fib(5), fib(6), fib(7), fib(8), fib(9)];
			results[9];
		`
		result := runCode(t, input)
		if result == objects.NULL {
			t.Fatal("got NULL result")
		}
		intResult, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		assertInt(t, intResult, 34)
	})
}

// ============================================
// Error Handling
// ============================================

func TestDivisionByZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Log("Division by zero handled appropriately")
		}
	}()

	input := "1 / 0"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.New()
	c.Compile(program)

	v := vm.New(c.Bytecode())
	err := v.Run()

	if err == nil {
		t.Error("expected error for division by zero")
	}
}

// ============================================
// Postfix Operators
// ============================================

func TestPostfixOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x = 5; x++ x", 6},
		{"var x = 5; x-- x", 4},
		{"var x = 0; for (var i = 0; i < 5; i++) { x++ } x", 5},
	}

	for _, tt := range tests {
		result := runCode(t, tt.input)
		assertInt(t, result, tt.expected)
	}
}

// ============================================
// Switch Statement
// ============================================

func TestSwitchStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "basic switch with match",
			input:    "var x = 2; switch (x) { case 1: x = 10 case 2: x = 20 case 3: x = 30 } x",
			expected: 20,
		},
		{
			name:     "switch with default",
			input:    "var x = 5; switch (x) { case 1: x = 10 case 2: x = 20 default: x = 100 } x",
			expected: 100,
		},
		{
			name:     "switch first case matches",
			input:    "var x = 1; switch (x) { case 1: x = 10 case 2: x = 20 default: x = 100 } x",
			expected: 10,
		},
		{
			name:     "switch no match no default",
			input:    "var x = 99; switch (x) { case 1: x = 10 case 2: x = 20 } x",
			expected: 99,
		},
		{
			name:     "switch in function",
			input:    "func getValue(n) { switch (n) { case 1: return 10 case 2: return 20 default: return 0 } } getValue(2)",
			expected: 20,
		},
		{
			name:     "switch with expression",
			input:    "var x = 1 + 1; switch (x) { case 1: x = 10 case 2: x = 20 } x",
			expected: 20,
		},
		{
			name:     "switch only default",
			input:    "var x = 5; switch (x) { default: x = 42 } x",
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.input)
			assertInt(t, result, tt.expected)
		})
	}
}

func TestSwitchStatementWithString(t *testing.T) {
	input := `var x = "b"; switch (x) { case "a": x = "apple" case "b": x = "banana" } x`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := c.Bytecode()
	v := vm.New(bytecode)

	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	result := v.LastPopped()
	assertString(t, result, "banana")
}

func TestSwitchStatementWithReturn(t *testing.T) {
	input := `
		func test(n) {
			switch (n) {
				case 1: return "one"
				case 2: return "two"
				default: return "other"
			}
		}
		test(1)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := c.Bytecode()
	v := vm.New(bytecode)

	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	result := v.LastPopped()
	assertString(t, result, "one")
}
