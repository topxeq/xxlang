// pkg/vm/vm_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// ============================================
// Helper Functions
// ============================================

func parse(input string) *parser.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testCompile(input string) (*compiler.Bytecode, error) {
	program := parse(input)
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		return nil, err
	}
	return c.Bytecode(), nil
}

func runVM(t *testing.T, input string) *VM {
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	return vm
}

func testIntegerObject(t *testing.T, expected int64, actual objects.Object) {
	result, ok := actual.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", actual)
	}
	if result.Value != expected {
		t.Fatalf("expected %d, got %d", expected, result.Value)
	}
}

func testFloatObject(t *testing.T, expected float64, actual objects.Object) {
	result, ok := actual.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", actual)
	}
	if result.Value != expected {
		t.Fatalf("expected %f, got %f", expected, result.Value)
	}
}

func testStringObject(t *testing.T, expected string, actual objects.Object) {
	result, ok := actual.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", actual)
	}
	if result.Value != expected {
		t.Fatalf("expected %q, got %q", expected, result.Value)
	}
}

func testBooleanObject(t *testing.T, expected bool, actual objects.Object) {
	result, ok := actual.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", actual)
	}
	if result.Value != expected {
		t.Fatalf("expected %t, got %t", expected, result.Value)
	}
}

func testNullObject(t *testing.T, actual objects.Object) {
	if actual != objects.NULL {
		t.Fatalf("expected NULL, got %T (%+v)", actual, actual)
	}
}

func testArrayObject(t *testing.T, expected []interface{}, actual objects.Object) {
	arr, ok := actual.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", actual)
	}
	if len(arr.Elements) != len(expected) {
		t.Fatalf("expected array of length %d, got %d", len(expected), len(arr.Elements))
	}
	for i, exp := range expected {
		switch e := exp.(type) {
		case int:
			testIntegerObject(t, int64(e), arr.Elements[i])
		case int64:
			testIntegerObject(t, e, arr.Elements[i])
		case string:
			testStringObject(t, e, arr.Elements[i])
		case bool:
			testBooleanObject(t, e, arr.Elements[i])
		default:
			t.Fatalf("unsupported expected type %T", exp)
		}
	}
}

// ============================================
// VM Tests
// ============================================

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1;", 1},
		{"2;", 2},
		{"1 + 2;", 3},
		{"1 - 2;", -1},
		{"1 * 2;", 2},
		{"4 / 2;", 2},
		{"10 % 3;", 1},
		{"1 + 2 * 3;", 7},
		{"(1 + 2) * 3;", 9},
		{"-5;", -5},
		{"50 / 2 * 2 + 10 - 5;", 55},
		{"5 + 5 + 5 + 5 - 10;", 10},
		{"2 * 2 * 2 * 2 * 2;", 32},
		{"5 * 2 + 10;", 20},
		{"5 + 2 * 10;", 25},
		{"5 * (2 + 10);", 60},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.5;", 1.5},
		{"1.5 + 2.5;", 4.0},
		{"5.0 - 2.5;", 2.5},
		{"2.5 * 2.0;", 5.0},
		{"7.5 / 2.5;", 3.0},
		{"-3.5;", -3.5},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testFloatObject(t, tt.expected, vm.LastPopped())
	}
}

func TestStringOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello";`, "hello"},
		{`"hello" + " " + "world";`, "hello world"},
		{`"";`, ""},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testStringObject(t, tt.expected, vm.LastPopped())
	}
}

func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testBooleanObject(t, tt.expected, vm.LastPopped())
	}
}

func TestNullLiteral(t *testing.T) {
	vm := runVM(t, "null;")
	testNullObject(t, vm.LastPopped())
}

func TestComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2;", true},
		{"1 > 2;", false},
		{"1 < 1;", false},
		{"1 > 1;", false},
		{"1 == 1;", true},
		{"1 != 1;", false},
		{"1 == 2;", false},
		{"1 != 2;", true},
		{"1 <= 2;", true},
		{"1 >= 2;", false},
		{"1 <= 1;", true},
		{"1 >= 1;", true},
		{"true == true;", true},
		{"true == false;", false},
		{"false != true;", true},
		{"null == null;", true},
		{"null != null;", false},
		{`"a" == "a";`, true},
		{`"a" == "b";`, false},
		{`"a" != "b";`, true},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testBooleanObject(t, tt.expected, vm.LastPopped())
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true;", true},
		{"true && false;", false},
		{"false && true;", false},
		{"false && false;", false},
		{"true || true;", true},
		{"true || false;", true},
		{"false || true;", true},
		{"false || false;", false},
		{"!true;", false},
		{"!false;", true},
		{"!null;", true},
		{"!0;", true},
		{"!1;", false},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testBooleanObject(t, tt.expected, vm.LastPopped())
	}
}

func TestVariableBinding(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x = 5; x;", 5},
		{"var x = 5; var y = 10; x;", 5},
		{"var x = 5; var y = 10; y;", 10},
		{"var x = 5; x = 10; x;", 10},
		{"var a = 1; var b = 2; var c = 3; a + b + c;", 6},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestIfStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10; }", 10},
		{"if (false) { 10; }", nil},
		{"if (1) { 10; }", 10},
		{"if (1 < 2) { 10; }", 10},
		{"if (1 > 2) { 10; }", nil},
		{"if (1 > 2) { 10; } else { 20; }", 20},
		{"if (1 < 2) { 10; } else { 20; }", 10},
		{"if (true) { var x = 10; x; } else { var y = 20; y; }", 10},
		{"if (false) { var x = 10; x; } else { var y = 20; y; }", 20},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

func TestWhileLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var i = 0; while (i < 5) { i = i + 1; } i;", 5},
		{"var sum = 0; var i = 1; while (i <= 5) { sum = sum + i; i = i + 1; } sum;", 15},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestForLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var sum = 0; for (var i = 1; i <= 5; i = i + 1) { sum = sum + i; } sum;", 15},
		{"var i = 0; for (; i < 5; ) { i = i + 1; } i;", 5},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestFunctionDefinition(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"func f() { return 5; }; f();", 5},
		{"func f() { return 10; }; f();", 10},
		{"func f() { return 5; }; func g() { return f(); }; g();", 5},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestFunctionWithArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"func f(x) { return x; }; f(5);", 5},
		{"func f(x) { return x + 1; }; f(5);", 6},
		{"func f(x, y) { return x + y; }; f(3, 4);", 7},
		{"func f(x, y) { return x * y; }; f(3, 4);", 12},
		{"func f(a, b, c) { return a + b + c; }; f(1, 2, 3);", 6},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestClosures(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"func makeCounter() { var count = 0; func counter() { count = count + 1; return count; }; return counter; }; var c = makeCounter(); c(); c();",
			2,
		},
		{
			"func makeAdder(x) { func adder(y) { return x + y; }; return adder; }; var add5 = makeAdder(5); add5(3);",
			8,
		},
		{
			"func makeAdder(x) { func adder(y) { return x + y; }; return adder; }; var add10 = makeAdder(10); add10(5);",
			15,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestRecursion(t *testing.T) {
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
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestArrayLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{"[];", []interface{}{}},
		{"[1];", []interface{}{1}},
		{"[1, 2, 3];", []interface{}{1, 2, 3}},
		{"[1, 2 + 3, 4 * 5];", []interface{}{1, 5, 20}},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testArrayObject(t, tt.expected, vm.LastPopped())
	}
}

func TestArrayIndexing(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][0];", 1},
		{"[1, 2, 3][1];", 2},
		{"[1, 2, 3][2];", 3},
		{"var arr = [1, 2, 3]; arr[1];", 2},
		{"var arr = [1, 2, 3]; arr[0] + arr[1] + arr[2];", 6},
		{"var arr = [1, 2, 3]; var i = 0; arr[i];", 1},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		}
	}
}

func TestMapLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`var m = {"a": 1}; m["a"];`, 1},
		{`var m = {"a": 1, "b": 2}; m["a"] + m["b"];`, 3},
		{`var m = {}; m["c"] = 3; m["c"];`, 3},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("hello");`, 5},
		{`len("");`, 0},
		{`len([1, 2, 3]);`, 3},
		{`len([]);`, 0},
		{`typeOf(1);`, "INT"},
		{`typeOf("hello");`, "STRING"},
		{`typeOf(true);`, "BOOL"},
		{`typeOf(null);`, "NULL"},
		{`typeOf([1, 2]);`, "ARRAY"},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

func TestDivisionByZero(t *testing.T) {
	input := "1 / 0;"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for division by zero")
	}
}

func TestBuiltinPrint(t *testing.T) {
	// Just ensure print and println don't error
	tests := []string{
		`print("hello");`,
		`println("hello");`,
		`print(1, 2, 3);`,
	}

	for _, input := range tests {
		vm := runVM(t, input)
		testNullObject(t, vm.LastPopped())
	}
}

func TestMixedIntFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1 + 2.0;", 3.0},
		{"1.0 + 2;", 3.0},
		{"5.0 - 2;", 3.0},
		{"5 - 2.0;", 3.0},
		{"2 * 3.0;", 6.0},
		{"2.0 * 3;", 6.0},
		{"6.0 / 2;", 3.0},
		{"6 / 2.0;", 3.0},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testFloatObject(t, tt.expected, vm.LastPopped())
	}
}

func TestAssignmentOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x = 5; x += 3; x;", 8},
		{"var x = 5; x -= 3; x;", 2},
		{"var x = 5; x *= 3; x;", 15},
		{"var x = 6; x /= 3; x;", 2},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestPostfixOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x = 5; x++; x;", 6},
		{"var x = 5; x--; x;", 4},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestNestedFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"func outer(x) { func inner(y) { return x + y; }; return inner; }; var f = outer(10); f(5);",
			15,
		},
		{
			"func outer(x) { func middle(y) { func inner(z) { return x + y + z; }; return inner; }; return middle; }; var m = outer(1)(2); m(3);",
			6,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestArrayMutation(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var arr = [1, 2, 3]; arr[0] = 10; arr[0];", 10},
		{"var arr = [1, 2, 3]; arr[1] = 20; arr[1];", 20},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestVMNewWithGlobals(t *testing.T) {
	input := "var x = 5; x;"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	globals := make([]objects.Object, compiler.GlobalsSize)
	vm := NewWithGlobalsStore(bytecode, globals)

	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	testIntegerObject(t, 5, vm.LastPopped())
	testIntegerObject(t, 5, globals[0])
}

func TestVMGlobals(t *testing.T) {
	input := "var x = 42;"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	globals := vm.Globals()
	if len(globals) == 0 {
		t.Fatal("expected globals to have at least one element")
	}

	testIntegerObject(t, 42, globals[0])
}
