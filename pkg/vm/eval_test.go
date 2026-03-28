// pkg/vm/eval_test.go
package vm

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

func TestEval_ContinuousExecution(t *testing.T) {
	// Create VM and compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// Compile and run first code
	code1 := `var x = 10`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Verify x is set
	x, ok := vm.GetGlobal("x")
	if !ok {
		t.Fatal("Expected x to be defined")
	}
	// Numbers can be Int or Float
	switch v := x.(type) {
	case *objects.Int:
		if v.Value != 10 {
			t.Errorf("Expected x=10, got %v", v.Value)
		}
	case *objects.Float:
		if v.Value != 10.0 {
			t.Errorf("Expected x=10, got %v", v.Value)
		}
	default:
		t.Errorf("Expected numeric value, got %T", x)
	}
}

func TestEval_VariableSharing(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// First execution: define a variable
	code1 := `var greeting = "Hello"`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Second execution: use the variable
	code2 := `greeting + " World"`
	result, err := vm.Eval(code2)
	if err != nil {
		t.Fatalf("Failed to eval code2: %v", err)
	}

	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if str.Value != "Hello World" {
		t.Errorf("Expected 'Hello World', got %v", str.Value)
	}
}

func TestEval_VariableModification(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// First execution: define counter
	code1 := `var counter = 0`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Second execution: increment
	code2 := `counter = counter + 1`
	_, err = vm.Eval(code2)
	if err != nil {
		t.Fatalf("Failed to eval code2: %v", err)
	}

	// Third execution: increment again
	code3 := `counter = counter + 1`
	_, err = vm.Eval(code3)
	if err != nil {
		t.Fatalf("Failed to eval code3: %v", err)
	}

	// Verify counter is 2
	counter, ok := vm.GetGlobal("counter")
	if !ok {
		t.Fatal("Expected counter to be defined")
	}
	assertNumericValue(t, counter, 2)
}

func TestEval_ComplexDataSharing(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// First execution: define an object
	code1 := `var user = {"name": "Alice", "age": 30}`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Second execution: access object properties
	code2 := `user.name + " is " + toStr(user.age) + " years old"`
	result, err := vm.Eval(code2)
	if err != nil {
		t.Fatalf("Failed to eval code2: %v", err)
	}

	expected := "Alice is 30 years old"
	if result.(*objects.String).Value != expected {
		t.Errorf("Expected '%s', got %v", expected, result)
	}
}

func TestEval_FunctionDefinition(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// First execution: define a function
	code1 := `func add(a, b) { return a + b }`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Second execution: call the function
	code2 := `add(10, 20)`
	result, err := vm.Eval(code2)
	if err != nil {
		t.Fatalf("Failed to eval code2: %v", err)
	}

	if result.(*objects.Int).Value != 30 {
		t.Errorf("Expected 30, got %v", result)
	}
}

func TestEval_MultipleExecutions(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// Create empty VM
	code1 := ``
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Execute multiple code snippets
	codes := []string{
		`var a = 1`,
		`var b = 2`,
		`var c = a + b`,
		`c * 2`,
	}

	for i, code := range codes {
		_, err := vm.Eval(code)
		if err != nil {
			t.Fatalf("Failed to eval code[%d]: %v", i, err)
		}
	}

	// Verify final result
	cVal, ok := vm.GetGlobal("c")
	if !ok {
		t.Fatal("Expected c to be defined")
	}
	if cVal.(*objects.Int).Value != 3 {
		t.Errorf("Expected c=3, got %v", cVal)
	}
}

func TestEval_SetAndGetGlobal(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	// Create VM
	code := `var x = 0`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Set global from Go
	vm.SetGlobal("y", objects.NewInt(42))

	// Use the global in Eval
	result, err := vm.Eval(`x + y`)
	if err != nil {
		t.Fatalf("Failed to eval: %v", err)
	}

	if result.(*objects.Int).Value != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Get the global
	y, ok := vm.GetGlobal("y")
	if !ok {
		t.Fatal("Expected y to be defined")
	}
	if y.(*objects.Int).Value != 42 {
		t.Errorf("Expected y=42, got %v", y)
	}
}

func TestEval_DefinedGlobals(t *testing.T) {
	// Create compiler with shared symbol table
	c := compiler.NewRegCompiler()

	code := `var a = 1; var b = 2; var c = 3`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	globals := vm.DefinedGlobals()
	if len(globals) < 3 {
		t.Errorf("Expected at least 3 globals, got %d: %v", len(globals), globals)
	}

	// Check that a, b, c are in the list
	found := make(map[string]bool)
	for _, name := range globals {
		found[name] = true
	}
	if !found["a"] || !found["b"] || !found["c"] {
		t.Errorf("Expected a, b, c in globals, got %v", globals)
	}
}

// Helper function to compile code with a given compiler
func compileWithCompiler(c *compiler.RegCompiler, code string) (*compiler.Bytecode, error) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	_, err := c.Compile(program)
	if err != nil {
		return nil, err
	}

	return c.Bytecode(), nil
}

// compileCode is kept for backward compatibility
func compileCode(code string) (*compiler.Bytecode, error) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		return nil, err
	}

	return c.Bytecode(), nil
}

// assertNumericValue checks if an object has the expected numeric value
func assertNumericValue(t *testing.T, obj objects.Object, expected int64) bool {
	switch v := obj.(type) {
	case *objects.Int:
		if v.Value != expected {
			t.Errorf("Expected %d, got %d", expected, v.Value)
			return false
		}
		return true
	case *objects.Float:
		if int64(v.Value) != expected {
			t.Errorf("Expected %d, got %v", expected, v.Value)
			return false
		}
		return true
	default:
		t.Errorf("Expected numeric value, got %T", obj)
		return false
	}
}

// TestEval_Closure tests closure support in Eval mode
func TestEval_Closure(t *testing.T) {
	c := compiler.NewRegCompiler()

	// Define a function that returns a closure
	code1 := `
var makeCounter = func() {
	var count = 0
	return func() {
		count = count + 1
		return count
	}
}
`
	bytecode1, err := compileWithCompiler(c, code1)
	if err != nil {
		t.Fatalf("Failed to compile code1: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode1, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run code1: %v", err)
	}

	// Create a counter
	_, err = vm.Eval(`var counter = makeCounter()`)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Call the counter multiple times
	result1, err := vm.Eval(`counter()`)
	if err != nil {
		t.Fatalf("Failed to call counter: %v", err)
	}
	assertNumericValue(t, result1, 1)

	result2, err := vm.Eval(`counter()`)
	if err != nil {
		t.Fatalf("Failed to call counter: %v", err)
	}
	assertNumericValue(t, result2, 2)

	result3, err := vm.Eval(`counter()`)
	if err != nil {
		t.Fatalf("Failed to call counter: %v", err)
	}
	assertNumericValue(t, result3, 3)
}

// TestEval_Conditionals tests if/else in Eval mode
func TestEval_Conditionals(t *testing.T) {
	c := compiler.NewRegCompiler()

	code := `var x = 10`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Test conditional
	result, err := vm.Eval(`if (x > 5) { "big" } else { "small" }`)
	if err != nil {
		t.Fatalf("Failed to eval conditional: %v", err)
	}

	str, ok := result.(*objects.String)
	if !ok || str.Value != "big" {
		t.Errorf("Expected 'big', got %v", result)
	}
}

// TestEval_Loops tests loops in Eval mode
func TestEval_Loops(t *testing.T) {
	c := compiler.NewRegCompiler()

	code := `var sum = 0`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Run a loop
	_, err = vm.Eval(`for (var i = 1; i <= 5; i = i + 1) { sum = sum + i }`)
	if err != nil {
		t.Fatalf("Failed to eval loop: %v", err)
	}

	// Check result
	sum, ok := vm.GetGlobal("sum")
	if !ok {
		t.Fatal("Expected sum to be defined")
	}
	assertNumericValue(t, sum, 15)
}

// TestEval_Arrays tests array operations in Eval mode
func TestEval_Arrays(t *testing.T) {
	c := compiler.NewRegCompiler()

	code := `var arr = [1, 2, 3, 4, 5]`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Access array element
	result, err := vm.Eval(`arr[2]`)
	if err != nil {
		t.Fatalf("Failed to access array: %v", err)
	}
	assertNumericValue(t, result, 3)

	// Modify array
	_, err = vm.Eval(`arr[2] = 10`)
	if err != nil {
		t.Fatalf("Failed to modify array: %v", err)
	}

	// Check modification
	result, err = vm.Eval(`arr[2]`)
	if err != nil {
		t.Fatalf("Failed to access modified array: %v", err)
	}
	assertNumericValue(t, result, 10)
}

// TestEval_ErrorHandling tests error handling in Eval mode
func TestEval_ErrorHandling(t *testing.T) {
	c := compiler.NewRegCompiler()

	code := `var result = "no error"`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Test try-catch
	_, err = vm.Eval(`
try {
	throw "test error"
} catch (e) {
	result = e
}
`)
	if err != nil {
		t.Fatalf("Failed to eval try-catch: %v", err)
	}

	// Check result
	result, ok := vm.GetGlobal("result")
	if !ok {
		t.Fatal("Expected result to be defined")
	}
	str, ok := result.(*objects.String)
	if !ok || str.Value != "test error" {
		t.Errorf("Expected 'test error', got %v", result)
	}
}

// TestEval_ClassDefinition tests class definition in Eval mode
func TestEval_ClassDefinition(t *testing.T) {
	c := compiler.NewRegCompiler()

	// Define a class with proper Xxlang syntax (func keyword for methods)
	code := `
class Point {
	func init(x, y) {
		this.x = x
		this.y = y
	}

	func add(other) {
		return Point(this.x + other.x, this.y + other.y)
	}
}
`
	bytecode, err := compileWithCompiler(c, code)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	// Create instances
	_, err = vm.Eval(`var p1 = Point(1, 2)`)
	if err != nil {
		t.Fatalf("Failed to create p1: %v", err)
	}

	_, err = vm.Eval(`var p2 = Point(3, 4)`)
	if err != nil {
		t.Fatalf("Failed to create p2: %v", err)
	}

	// Call method
	result, err := vm.Eval(`p1.add(p2)`)
	if err != nil {
		t.Fatalf("Failed to call add: %v", err)
	}

	// Result should be a Point instance
	instance, ok := result.(*objects.Instance)
	if !ok {
		t.Fatalf("Expected Instance, got %T", result)
	}

	// Check the result
	xVal := instance.Fields["x"]
	yVal := instance.Fields["y"]

	if xVal.(*objects.Int).Value != 4 {
		t.Errorf("Expected x=4, got %v", xVal)
	}
	if yVal.(*objects.Int).Value != 6 {
		t.Errorf("Expected y=6, got %v", yVal)
	}
}

// TestEval_ReturnValues tests that Eval properly returns expression values
func TestEval_ReturnValues(t *testing.T) {
	c := compiler.NewRegCompiler()

	// Create empty VM
	bytecode, err := compileWithCompiler(c, ``)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	vm := NewRegVMWithSymbolTable(bytecode, c.SymbolTable())
	if err := vm.Run(); err != nil {
		t.Fatalf("Failed to run: %v", err)
	}

	tests := []struct {
		code     string
		expected interface{}
	}{
		{`1 + 2`, int64(3)},
		{`"hello" + " world"`, "hello world"},
		{`true && false`, false},
		{`true || false`, true},
		{`[1, 2, 3][0]`, int64(1)},
		{`len([1, 2, 3])`, int64(3)},
	}

	for _, tt := range tests {
		result, err := vm.Eval(tt.code)
		if err != nil {
			t.Errorf("Failed to eval %q: %v", tt.code, err)
			continue
		}

		switch expected := tt.expected.(type) {
		case int64:
			assertNumericValue(t, result, expected)
		case string:
			str, ok := result.(*objects.String)
			if !ok || str.Value != expected {
				t.Errorf("Code %q: expected %q, got %v", tt.code, expected, result)
			}
		case bool:
			b, ok := result.(*objects.Bool)
			if !ok || b.Value != expected {
				t.Errorf("Code %q: expected %v, got %v", tt.code, expected, result)
			}
		}
	}
}
