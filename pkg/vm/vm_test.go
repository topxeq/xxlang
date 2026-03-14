// pkg/vm/vm_test.go
package vm

import (
	"strings"
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
		{"var total = 0; var n = 1; while (n <= 5) { total = total + n; n = n + 1; } total;", 15},
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
		{"var total = 0; for (var i = 1; i <= 5; i = i + 1) { total = total + i; } total;", 15},
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

func TestClassCreation(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				var name = ""
			}
			var p = new Person()
			p.name
			`,
			"",
		},
		{
			`
			class Counter {
				var count = 0
				func inc() {
					this.count = this.count + 1
				}
			}
			var c = new Counter()
			c.inc()
			c.count
			`,
			1,
		},
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
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

// ============================================
// Error Condition Tests
// ============================================

// runVMExpectError compiles and runs code expecting an error
func runVMExpectError(t *testing.T, input string) error {
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
	return err
}

func TestDivisionByZeroErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"1 / 0;"},
		{"10 / 0;"},
		{"-5 / 0;"},
		{"1.0 / 0.0;"},
		{"5 % 0;"},
		{"10.0 % 0;"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestArrayIndexOutOfBoundsErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		// Setting index out of bounds should error
		{"var arr = [1, 2, 3]; arr[10] = 5;"},
		{"var arr = [1, 2, 3]; arr[-1] = 5;"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestArrayIndexOutOfBoundsReturnsNull(t *testing.T) {
	// Reading out of bounds returns null instead of error
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][10];", nil},
		{"[1, 2, 3][-1];", nil},
		{"var arr = [1]; arr[5];", nil},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testNullObject(t, vm.LastPopped())
	}
}

func TestWrongNumberOfArgumentsErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"func f(x) { return x; }; f();"},
		{"func f(x) { return x; }; f(1, 2);"},
		{"func f(x, y) { return x + y; }; f(1);"},
		{"func f() { return 1; }; f(1);"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestCallNonFunctionErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"var x = 5; x();"},
		{"5();"},
		{`"hello"();`},
		{"[1, 2, 3]();"},
		{"true();"},
		{"null();"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestIndexNonIndexableErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"5[0];"},
		{"true[0];"},
		{"null[0];"},
		{"var f = func() { return 1; }; f[0];"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestArrayIndexMustBeInteger(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`[1, 2, 3]["a"];`},
		{`var arr = [1, 2, 3]; arr["index"];`},
		{"[1, 2, 3][true];"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestSetIndexNotSupported(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"var x = 5; x[0] = 1;"},
		{"true[0] = 1;"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestTypeMismatchBinaryOp(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`5 - "hello";`},
		{`"hello" * 5;`},
		{"true + false;"},
		{"[1, 2] - [3, 4];"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestNegationNotSupported(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`-"hello";`},
		{"-true;"},
		{"-null;"},
		{"-[1, 2, 3];"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

func TestComparisonNotSupported(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`"hello" < 5;`},
		{"[1, 2] > 3;"},
		{"true < 10;"},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

// ============================================
// Stack Operation Tests
// ============================================

func TestStackOperations(t *testing.T) {
	// Test that stack operations work correctly through actual code
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Nested expressions
		{"((1 + 2) * (3 + 4));", 21},
		{"(10 - 5) * (8 / 2);", 20},
		// Complex nesting
		{"1 + 2 + 3 + 4 + 5;", 15},
		{"(((1)));", 1},
		// Mixed types
		{`"a" + "b" + "c";`, "abc"},
		{"1.5 + 2.5 + 3.0;", 7.0},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case float64:
			testFloatObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

func TestDeeplyNestedCalls(t *testing.T) {
	input := `
		func a(x) { return x + 1; }
		func b(x) { return a(x) + 1; }
		func c(x) { return b(x) + 1; }
		func d(x) { return c(x) + 1; }
		func e(x) { return d(x) + 1; }
		e(0);
	`
	vm := runVM(t, input)
	testIntegerObject(t, 5, vm.LastPopped())
}

func TestStringIndexing(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`"hello"[0];`, "h"},
		{`"hello"[4];`, "o"},
		{`"hello"[10];`, nil},  // out of bounds returns null
		{`"hello"[-1];`, nil},  // negative returns null
		{`var s = "abc"; s[1];`, "b"},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case string:
			testStringObject(t, expected, vm.LastPopped())
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

func TestStringIndexMustBeInteger(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`"hello"["a"];`},
		{`"hello"[true];`},
	}

	for _, tt := range tests {
		runVMExpectError(t, tt.input)
	}
}

// ============================================
// Additional Coverage Tests
// ============================================

func TestMoreClassFeatures(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Test class with multiple fields
		{
			`
			class Point {
				var x = 0
				var y = 0
			}
			var p = new Point()
			p.x = 10
			p.y = 20
			p.x + p.y
			`,
			30,
		},
		// Test class with method that returns value
		{
			`
			class Adder {
				func add(a, b) {
					return a + b
				}
			}
			var a = new Adder()
			a.add(3, 4)
			`,
			7,
		},
		// Test class with this
		{
			`
			class Counter {
				var count = 0
				func increment() {
					this.count = this.count + 1
				 return this.count
                }
            }
            var c = new Counter()
            c.increment()
            c.increment()
            `,
			2,
		},
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
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

func TestMapOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Map with various key types
		{`var m = {"a": 1, "b": 2}; m["a"];`, 1},
		{`var m = {1: "one", 2: "two"}; m[1];`, "one"},
		{`var m = {"x": 1}; m["y"];`, nil}, // non-existent key
		// Map mutation
		{`var m = {}; m["key"] = 42; m["key"];`, 42},
		{`var m = {"a": 1}; m["a"] = 99; m["a"];`, 99},
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
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

func TestBuiltinLenOnNonSequence(t *testing.T) {
	// len on non-sequence types returns an error object
	tests := []struct {
		input string
	}{
		{`len(1);`},
		{`len(true);`},
		{`len(null);`},
	}

	for _, tt := range tests {
		bytecode, err := testCompile(tt.input)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(bytecode)
		err = vm.Run()
		// The VM may return an error or push an Error object onto the stack
		// Either way, the operation should fail in some way
		if err == nil {
			// Check if the result is an Error object
			result := vm.LastPopped()
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("expected error for len on non-sequence, got %T", result)
			}
		}
	}
}

func TestBuiltinTypeOfWithNil(t *testing.T) {
	// Test that typeOf works with nil
	tests := []struct {
		input    string
		expected string
	}{
		{`typeOf(null);`, "NULL"},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testStringObject(t, tt.expected, vm.LastPopped())
	}
}

func TestArrayConcatenation(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Arrays created and accessed
		{"var a = [1]; var b = [2, 3]; a[0] + b[0];", 3},
		{"var a = [1, 2, 3]; var elem = a[0]; elem;", 1},
		{"[[1, 2], [3, 4]][0][0];", 1}, // nested arrays
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestEarlyReturn(t *testing.T) {
	// Test early return from function
	tests := []struct {
		input    string
		expected int64
	}{
		{"func f() { return 1; return 2; } f();", 1},
		{"func f(x) { if (x > 5) { return x; } return 0; } f(10);", 10},
		{"func f(x) { if (x > 5) { return x; } return 0; } f(3);", 0},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testIntegerObject(t, tt.expected, vm.LastPopped())
	}
}

func TestDefaultValues(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Variables with explicit null initialization
		{"var x = null; x;", nil},
		{"var y = null; y = null; y;", nil},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		if tt.expected == nil {
			testNullObject(t, vm.LastPopped())
		}
	}
}

func TestComplexConditionals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (1 && 1) { 10; }", 10},
		{"if (1 && 0) { 10; } else { 20; }", 20},
		{"if (0 || 1) { 10; }", 10},
		{"if (0 || 0) { 10; } else { 20; }", 20},
		// !0 is true (since 0 is falsy), so the if branch should execute
		{"if (!0) { 10; } else { 20; }", 10},
		// !1 is false (since 1 is truthy), so the else branch should execute
		{"if (!1) { 10; } else { 20; }", 20},
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

func TestCallStackFormatting(t *testing.T) {
	// Test that GetCallStack returns something without panicking
	// Define g first, then f which calls g
	input := "func g() { 1 / 0; } func f() { g(); } f();"

	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for division by zero")
	}

	// Just verify GetCallStack doesn't panic and returns a non-empty string
	callStack := vm.GetCallStack()
	if callStack == "" {
		t.Error("expected non-empty call stack")
	}
}

func TestFormatError(t *testing.T) {
	// Test that formatError works with and without source map
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

	// Error message should contain "division by zero"
	if !containsString(err.Error(), "division by zero") {
		t.Errorf("expected error to contain 'division by zero', got: %s", err.Error())
	}
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================
// Closure Tests
// ============================================

func TestClosureMethods(t *testing.T) {
	// Test Closure Inspect
	c := &Closure{
		Fn:       &compiler.CompiledFunction{Instructions: []byte{}},
		FreeVars: []objects.Object{&objects.Int{Value: 1}},
	}
	if c.Inspect() == "" {
		t.Error("Closure.Inspect() should return non-empty string")
	}
	if !strings.Contains(c.Inspect(), "closure") {
		t.Errorf("Closure.Inspect() should contain 'closure', got %s", c.Inspect())
	}

	// Test Closure ToBool
	if c.ToBool() != objects.TRUE {
		t.Error("Closure.ToBool() should return TRUE")
	}

	// Test Closure HashKey
	hk := c.HashKey()
	if hk.Type != objects.ClosureType {
		t.Errorf("Closure.HashKey().Type should be ClosureType, got %s", hk.Type)
	}
}

// ============================================
// String Comparison Tests
// ============================================

func TestStringComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"a" < "b"`, true},
		{`"b" < "a"`, false},
		{`"a" < "a"`, false},
		{`"a" <= "a"`, true},
		{`"a" <= "b"`, true},
		{`"b" <= "a"`, false},
		{`"b" > "a"`, true},
		{`"a" > "b"`, false},
		{`"a" >= "a"`, true},
		{`"b" >= "a"`, true},
		{`"a" >= "b"`, false},
		{`"abc" < "abd"`, true},
		{`"abc" > "abb"`, true},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testBooleanObject(t, tt.expected, vm.LastPopped())
	}
}

// ============================================
// JumpIfTrue Tests
// ============================================

func TestJumpIfTrue(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Test || short-circuit with true
		{`true || false`, true},
		{`true || true`, true},
		{`false || true`, true},
		{`false || false`, false},
		// Test complex || expressions
		{`1 == 1 || 2 == 3`, true},
		{`1 == 2 || 2 == 2`, true},
		{`1 == 2 || 2 == 3`, false},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case bool:
			testBooleanObject(t, expected, vm.LastPopped())
		}
	}
}

// ============================================
// Class with Init Method Tests
// ============================================

func TestClassWithInit(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Point {
				func init(x, y) {
					this.x = x
					this.y = y
				}
			}
			var p = new Point(3, 4)
			p.x
			`,
			3,
		},
		{
			`
			class Point {
				func init(x, y) {
					this.x = x
					this.y = y
				}
			}
			var p = new Point(3, 4)
			p.y
			`,
			4,
		},
		{
			`
			class Counter {
				func init() {
					this.count = 0
				}
				func inc() {
					this.count = this.count + 1
				}
			}
			var c = new Counter()
			c.inc()
			c.inc()
			c.count
			`,
			2,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		}
	}
}

// ============================================
// Field Access Tests
// ============================================

func TestFieldAccess(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				var name = "John"
				var age = 30
			}
			var p = new Person()
			p.name
			`,
			"John",
		},
		{
			`
			class Person {
				var name = "John"
				var age = 30
			}
			var p = new Person()
			p.age
			`,
			30,
		},
		{
			`
			class Point {
				var x = 0
				var y = 0
			}
			var p = new Point()
			p.x = 10
			p.y = 20
			p.x + p.y
			`,
			30,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

// ============================================
// VM Utility Method Tests
// ============================================

func TestVMStackTop(t *testing.T) {
	input := "var x = 42; x"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	top := vm.StackTop()
	if top == nil {
		t.Fatal("StackTop() should not return nil after run")
	}
	testIntegerObject(t, 42, top)
}

func TestVMSetSourcePath(t *testing.T) {
	input := "42"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	vm.SetSourcePath("/test/path.xxl")
	// Should not panic
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
}

func TestVMGlobalsMethod(t *testing.T) {
	input := "var x = 10; x"
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
	if globals == nil {
		t.Fatal("Globals() should not return nil")
	}
}

// ============================================
// Return Statement Tests
// ============================================

func TestReturnValues(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`func f() { return 42; } f()`, 42},
		{`func f() { var x = 1; return x; } f()`, 1},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		case nil:
			testNullObject(t, vm.LastPopped())
		}
	}
}

// ============================================
// JumpIfTrue Extended Tests
// ============================================

func TestJumpIfTrueExtended(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Short-circuit || with true on left
		{`true || false`, true},
		{`true || true`, true},
		{`false || true`, true},
		{`false || false`, false},
		// With expressions
		{`1 == 1 || 2 == 3`, true},
		{`1 == 2 || 2 == 2`, true},
		{`1 == 2 || 2 == 3`, false},
		// Nested
		{`(true || false) || false`, true},
		{`false || (true || false)`, true},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case bool:
			testBooleanObject(t, expected, vm.LastPopped())
		}
	}
}

// ============================================
// Class Init Method Tests
// ============================================

func TestClassInitMethod(t *testing.T) {
	input := `
		class Point {
			func init(x, y) {
				this.x = x
				this.y = y
			}
		}
		var p = new Point(3, 4)
		p.x + p.y
	`
	vm := runVM(t, input)
	testIntegerObject(t, 7, vm.LastPopped())
}

// ============================================
// Class Field Mutation Tests
// ============================================

func TestClassFieldMutation(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Box {
				var value = 0
			}
			var b = new Box()
			b.value = 10
			b.value
			`,
			10,
		},
		{
			`
			class Box {
				var value = 0
			}
			var b = new Box()
			b.value = 5
			b.value = b.value + 5
			b.value
			`,
			10,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		}
	}
}

// ============================================
// Multiple Class Instance Tests
// ============================================

func TestMultipleClassInstances(t *testing.T) {
	input := `
		class Counter {
			var count = 0
			func inc() {
				this.count = this.count + 1
			}
		}
		var c1 = new Counter()
		var c2 = new Counter()
		c1.inc()
		c1.inc()
		c2.inc()
		c1.count + c2.count
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}

// ============================================
// Compound Assignment Tests
// ============================================

func TestCompoundAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`var x = 5; x += 3; x`, 8},
		{`var x = 5; x -= 3; x`, 2},
		{`var x = 5; x *= 3; x`, 15},
		{`var x = 15; x /= 3; x`, 5},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), vm.LastPopped())
		}
	}
}

// ============================================
// Float Arithmetic Extended Tests
// ============================================

func TestFloatArithmeticExtended(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`1.5 + 2.5`, 4.0},
		{`5.5 - 2.5`, 3.0},
		{`2.5 * 2.0`, 5.0},
		{`10.0 / 4.0`, 2.5},
		{`3.5 + 2`, 5.5},
		{`2 + 3.5`, 5.5},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testFloatObject(t, tt.expected, vm.LastPopped())
	}
}

// ============================================
// Boolean Operations Tests
// ============================================

func TestBooleanOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`!true`, false},
		{`!false`, true},
		{`!!true`, true},
		{`!!false`, false},
		{`true && true`, true},
		{`true && false`, false},
		{`false && true`, false},
		{`false && false`, false},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		testBooleanObject(t, tt.expected, vm.LastPopped())
	}
}

// ============================================
// VM Helper Methods Tests
// ============================================

func TestSetCurrentModule(t *testing.T) {
	input := `5`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	mod := &objects.Module{
		Name:    "test",
		Exports: make(map[string]objects.Object),
	}
	vm.SetCurrentModule(mod)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
}

func TestGetCallStack(t *testing.T) {
	input := `func outer() { func inner() { }; inner() }; outer()`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// GetCallStack should return call stack
	stack := vm.GetCallStack()
	if len(stack) == 0 {
		t.Error("expected non-empty call stack")
	}
}

// ============================================
// Tail Call Optimization Tests
// ============================================

// TestTailCallLocalsResizing tests that tail calls properly resize the locals
// array when the called function has more local variables than the caller.
// This is a regression test for a bug where tail-calling a function with
// more locals would cause an index out of bounds panic.
func TestTailCallLocalsResizing(t *testing.T) {
	// This test case has a wrapper function with only 3 locals (parameters)
	// that tail-calls a helper function with 6 locals (3 params + 3 vars)
	input := `
		func helper(a, b, c) {
			var x = a + 1
			var y = b + 2
			var z = c + 3
			return x + y + z
		}

		func simpleWrapper(x, y, z) {
			return helper(x, y, z)
		}

		simpleWrapper(1, 2, 3)
	`

	vm := runVM(t, input)
	// Expected: (1+1) + (2+2) + (3+3) = 2 + 4 + 6 = 12
	testIntegerObject(t, 12, vm.LastPopped())
}

// TestTailCallRecursiveWithLocals tests recursive tail calls with local variables
func TestTailCallRecursiveWithLocals(t *testing.T) {
	input := `
		func tailFact(n, acc) {
			if (n <= 1) {
				return acc
			}
			return tailFact(n - 1, n * acc)
		}

		tailFact(5, 1)
	`

	vm := runVM(t, input)
	testIntegerObject(t, 120, vm.LastPopped())
}

// ============================================
// Try-Catch-Finally Tests
// ============================================

func TestTryCatchBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`try { 42 } catch (e) { 0 }`,
			int64(42),
		},
		{
			`try { throw "error"; 0 } catch (e) { 42 }`,
			int64(42),
		},
		{
			`try { throw 100 } catch (e) { e }`,
			int64(100),
		},
		{
			`try { throw "test" } catch (e) { e }`,
			"test",
		},
		{
			`try { throw true } catch (e) { e }`,
			true,
		},
		{
			`try { throw } catch (e) { e }`,
			objects.NULL,
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		case bool:
			testBooleanObject(t, expected, vm.LastPopped())
		default:
			if expected == objects.NULL {
				testNullObject(t, vm.LastPopped())
			}
		}
	}
}

func TestTryFinally(t *testing.T) {
	// Test that finally block executes
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Finally block's value becomes the result (simplified behavior)
		{
			`try { 42 } finally { 100 }`,
			int64(100),
		},
		{
			// Note: can't use semicolon after block, use newline
			`var x = 0
			try { x = 1 } finally { x = 2 }
			x`,
			int64(2),
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		}
	}
}

func TestTryCatchFinally(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Finally block's value becomes the result
		{
			`try { throw "err" } catch (e) { 42 } finally { 100 }`,
			int64(100),
		},
		{
			`try { 1 } catch (e) { 0 } finally { 42 }`,
			int64(42),
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		}
	}
}

func TestNestedTryCatch(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`try {
				try {
					throw "inner"
				} catch (e) {
					e
				}
			} catch (e) {
				"outer"
			}`,
			"inner",
		},
		{
			`try {
				try {
					throw "inner"
				} catch (e) {
					throw "re-" + e
				}
			} catch (e) {
				e
			}`,
			"re-inner",
		},
		{
			`try {
				try {
					throw 1
				} finally {
					throw 2
				}
			} catch (e) {
				e
			}`,
			int64(2),
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

func TestThrowInFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`func mightThrow(should) {
				if (should) { throw "error" }
				return "ok"
			}
			try { mightThrow(true) } catch (e) { e }`,
			"error",
		},
		{
			`func inner() { throw 42 }
			func outer() {
				var result = 0
				try { inner() } catch (e) { result = e + 1 }
				return result
			}
			outer()`,
			int64(43),
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

func TestUnhandledException(t *testing.T) {
	tests := []string{
		`throw "unhandled"`,
		`func f() { throw "error" }; f()`,
	}

	for _, input := range tests {
		bytecode, err := testCompile(input)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(bytecode)
		err = vm.Run()
		if err == nil {
			t.Errorf("expected error for input: %s", input)
		}
	}
}

func TestTryCatchInLoop(t *testing.T) {
	// Test try-catch inside a while loop
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`var count = 0
			var i = 0
			while (i < 3) {
				try {
					throw i
				} catch (e) {
					count = count + 1
				}
				i = i + 1
			}
			count`,
			int64(3),
		},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		}
	}
}

func TestExceptionUnwinding(t *testing.T) {
	input := `
		var reached = ""
		func level3() {
			reached = reached + "3"
			throw "error"
			reached = reached + "X"
		}
		func level2() {
			reached = reached + "2"
			level3()
			reached = reached + "Y"
		}
		func level1() {
			reached = reached + "1"
			try {
				level2()
			} catch (e) {
				reached = reached + "C"
			}
			reached = reached + "E"
		}
		level1()
		reached
	`

	vm := runVM(t, input)
	testStringObject(t, "123CE", vm.LastPopped())
}

func TestFinallyRunsOnException(t *testing.T) {
	// Note: Current implementation doesn't re-throw after finally
	// This tests that finally runs and the value is from finally
	input := `
		try {
			throw "error"
		} finally {
			42
		}
	`

	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

// ============================================
// runCode Tests
// ============================================

func TestRunCodeBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`runCode("1 + 2")`, int64(3)},
		{`runCode("10 * 5")`, int64(50)},
		{`runCode("\"hello\" + \" world\"")`, "hello world"},
		{`runCode("true && false")`, false},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		case bool:
			testBooleanObject(t, expected, vm.LastPopped())
		}
	}
}

func TestRunCodeWithArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`runCode("x + y", {"x": 10, "y": 20})`, int64(30)},
		{`runCode("x * 2", {"x": 5})`, int64(10)},
		{`runCode("a + b + c", {"a": 1, "b": 2, "c": 3})`, int64(6)},
		{`runCode("name", {"name": "test"})`, "test"},
	}

	for _, tt := range tests {
		vm := runVM(t, tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, expected, vm.LastPopped())
		case string:
			testStringObject(t, expected, vm.LastPopped())
		}
	}
}

func TestRunCodeFunctionDefinition(t *testing.T) {
	// Each runCode call is an independent execution context
	// Functions must be defined and called within the same runCode call
	input := `runCode("func add(a, b) { return a + b }; add(10, 20)")`
	vm := runVM(t, input)
	testIntegerObject(t, 30, vm.LastPopped())
}

func TestRunCodeNested(t *testing.T) {
	input := `runCode("runCode(\"5 * 3\")")`
	vm := runVM(t, input)
	testIntegerObject(t, 15, vm.LastPopped())
}

func TestRunCodeReturnArray(t *testing.T) {
	input := `runCode("var arr = [1, 2, 3]; arr")`
	vm := runVM(t, input)
	arr, ok := vm.LastPopped().(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", vm.LastPopped())
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestRunCodeModifyArgument(t *testing.T) {
	input := `runCode("x = x * 2; x", {"x": 5})`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}
