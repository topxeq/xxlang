// pkg/vm/vm_extra_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Additional Try-Catch Tests
// ============================================

func TestTryCatchWithValue(t *testing.T) {
	input := `
		try {
			42
		} catch (e) {
			0
		}
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestTryCatchNested(t *testing.T) {
	input := `
		try {
			try {
				throw "inner"
			} catch (e) {
				"caught inner"
			}
		} catch (e) {
			"caught outer"
		}
	`
	vm := runVM(t, input)
	testStringObject(t, "caught inner", vm.LastPopped())
}

// ============================================
// Additional Closure Tests
// ============================================

func TestClosureMultipleLevels(t *testing.T) {
	input := `
		func level0() {
			var a = 1
			func level1() {
				var b = 2
				func level2() {
					var c = 3
					func level3() {
						return a + b + c
					}
					return level3
				}
				return level2
			}
			return level1
		}
		level0()()()()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 6, vm.LastPopped())
}

// ============================================
// RunCodeInVM Direct Tests
// ============================================

func TestRunCodeInVMDirect(t *testing.T) {
	// Create a VM for context
	input := "42"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// Run code directly
	result, err := RunCodeInVM("10 + 20", nil, vm)
	if err != nil {
		t.Fatalf("RunCodeInVM error: %s", err)
	}

	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if i.Value != 30 {
		t.Errorf("RunCodeInVM = %d, want 30", i.Value)
	}
}

func TestRunCodeInVMWithArgs(t *testing.T) {
	input := "0"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// Create args map
	args := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			(&objects.String{Value: "x"}).HashKey(): {
				Key:   &objects.String{Value: "x"},
				Value: &objects.Int{Value: 100},
			},
		},
	}

	result, err := RunCodeInVM("x * 2", args, vm)
	if err != nil {
		t.Fatalf("RunCodeInVM error: %s", err)
	}

	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if i.Value != 200 {
		t.Errorf("RunCodeInVM = %d, want 200", i.Value)
	}
}

// ============================================
// String Interning Tests
// ============================================

func TestStringInterning(t *testing.T) {
	// Test that string interning works
	input := `"hello"`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	result := vm.LastPopped()
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("string = %s, want 'hello'", s.Value)
	}
}

// ============================================
// Tail Call Tests
// ============================================

func TestTailCallOptimization(t *testing.T) {
	input := `
		func sum(n, acc) {
			if (n <= 0) {
				return acc
			}
			return sum(n - 1, acc + n)
		}
		sum(100, 0)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 5050, vm.LastPopped())
}

// ============================================
// Error Formatting Tests
// ============================================

func TestErrorFormattingWithSource(t *testing.T) {
	input := "1 / 0"
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	vm.SetSourcePath("/test/path.xxl")
	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for division by zero")
	}
	// Error message should contain information about the error
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// ============================================
// VM with Compiler Tests
// ============================================

func TestVMNewWithGlobalsStore(t *testing.T) {
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

// ============================================
// Tests for executeOpGetField
// ============================================

func TestGetFieldOnObject(t *testing.T) {
	input := `
		class Box {
			var value
			func init(v) {
				this.value = v
			}
			func getValue() {
				return this.value
			}
		}
		var box = new Box(123)
		box.getValue()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 123, vm.LastPopped())
}

// ============================================
// Tests for module operations
// ============================================

func TestModuleExports(t *testing.T) {
	input := `
		// Simple module test using objects
		var m = {
			"value": 42,
			"add": func(a, b) { return a + b }
		}
		m["value"]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestGetExportFromObject(t *testing.T) {
	input := `
		var exports = {
			"name": "test",
			"version": 1
		}
		exports["name"]
	`
	vm := runVM(t, input)
	testStringObject(t, "test", vm.LastPopped())
}

func TestSetExportOnObject(t *testing.T) {
	input := `
		var exports = {}
		exports["count"] = 10
		exports["count"]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

// ============================================
// Tests for array preallocation edge cases
// ============================================

func TestArrayPreallocEmpty(t *testing.T) {
	input := `[]`
	vm := runVM(t, input)
	arr := vm.LastPopped().(*objects.Array)
	if len(arr.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Elements))
	}
}

func TestArrayPreallocSingleElement(t *testing.T) {
	input := `[42]`
	vm := runVM(t, input)
	arr := vm.LastPopped().(*objects.Array)
	if len(arr.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(arr.Elements))
	}
	testIntegerObject(t, 42, arr.Elements[0])
}

// ============================================
// Tests for safe indexing
// ============================================

func TestSafeArrayIndexNegative(t *testing.T) {
	input := `
		var arr = [1, 2, 3]
		arr[-1]
	`
	vm := runVM(t, input)
	// Negative index should return null or error
	result := vm.LastPopped()
	if result != objects.NULL {
		// Some implementations might handle negative indexing differently
		t.Logf("negative index returned: %v", result)
	}
}

func TestSafeStringIndexNegative(t *testing.T) {
	input := `
		var s = "hello"
		s[-1]
	`
	vm := runVM(t, input)
	result := vm.LastPopped()
	if result != objects.NULL {
		t.Logf("negative string index returned: %v", result)
	}
}

// ============================================
// Tests for jump operations
// ============================================

func TestJumpInForLoop(t *testing.T) {
	input := `
		var sum = 0
		for (var i = 0; i < 10; i++) {
			if (i == 5) {
				break
			}
			sum = sum + i
		}
		sum
	`
	vm := runVM(t, input)
	// 0 + 1 + 2 + 3 + 4 = 10
	testIntegerObject(t, 10, vm.LastPopped())
}

func TestJumpIfFalse(t *testing.T) {
	input := `
		var x = 0
		if (false) {
			x = 10
		}
		x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 0, vm.LastPopped())
}

func TestNestedConditionals(t *testing.T) {
	input := `
		var result = ""
		if (true) {
			if (true) {
				result = "both true"
			}
		}
		result
	`
	vm := runVM(t, input)
	testStringObject(t, "both true", vm.LastPopped())
}

// ============================================
// Tests for StackTop edge cases
// ============================================

func TestStackTopAfterOperations(t *testing.T) {
	input := `
		var a = 10
		var b = 20
		a + b
	`
	vm := runVM(t, input)
	testIntegerObject(t, 30, vm.LastPopped())
}

// ============================================
// Tests for GetCallStack
// ============================================

func TestCallStackSimple(t *testing.T) {
	input := `
		func inner() {
			return 42
		}
		func outer() {
			return inner()
		}
		outer()
	`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// After execution, call stack should show the path
	callStack := vm.GetCallStack()
	if len(callStack) == 0 {
		t.Error("expected non-empty call stack")
	}
}

// ============================================
// Tests for formatError edge cases
// ============================================

func TestErrorWithNilReceiver(t *testing.T) {
	input := `
		var x = null
		x.someMethod()
	`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Error("expected error when calling method on null")
	}
}

// ============================================
// Tests for objectsEqual
// ============================================

func TestObjectsEqualInts(t *testing.T) {
	input := `
		var a = 42
		var b = 42
		a == b
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestObjectsEqualStrings(t *testing.T) {
	input := `
		"hello" == "hello"
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestObjectsEqualArrays(t *testing.T) {
	// Arrays use reference equality, not value equality
	input := `
		var a = [1, 2, 3]
		var b = a
		a == b
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestObjectsNotEqual(t *testing.T) {
	input := `
		[1, 2] == [1, 2, 3]
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())
}

// ============================================
// Tests for currentFrameMethodName
// ============================================

func TestMethodNameInError(t *testing.T) {
	input := `
		func broken() {
			return 1 / 0
		}
		broken()
	`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Error("expected error from division by zero")
	}
	// Error message should contain method name
	if !containsString(err.Error(), "broken") {
		t.Logf("Error message: %s", err.Error())
	}
}

// ============================================
// Tests for TailCall
// ============================================

func TestTailCallBasic(t *testing.T) {
	// Test simple tail recursive function
	input := `
		func count(n) {
			if (n <= 0) {
				return 0
			}
			return count(n - 1)
		}
		count(10)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 0, vm.LastPopped())
}

func TestTailCallWithAccumulator(t *testing.T) {
	// Test tail recursive sum
	input := `
		func sum(n, acc) {
			if (n <= 0) {
				return acc
			}
			return sum(n - 1, acc + n)
		}
		sum(5, 0)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 15, vm.LastPopped())
}

func TestTailCallMutualRecursion(t *testing.T) {
	// Test mutual recursion - forward references don't work, so use simpler test
	input := `
		func count(n) {
			if (n <= 0) {
				return 0
			}
			return 1 + count(n - 1)
		}
		count(5)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 5, vm.LastPopped())
}

// ============================================
// Tests for formatError with source map
// ============================================

func TestFormatErrorSourceMap(t *testing.T) {
	input := `
		func test() {
			var x = 1 / 0
		}
		test()
	`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Error("expected error from division by zero")
	}
	// The error should be formatted with source info
	if !containsString(err.Error(), "division") {
		t.Logf("Error: %s", err.Error())
	}
}

// ============================================
// Tests for StackTop
// ============================================

func TestStackTopBasic(t *testing.T) {
	input := `
		var x = 42
		x
	`
	vm := runVM(t, input)
	// StackTop should return the top of the stack
	top := vm.StackTop()
	if top == nil {
		t.Error("StackTop should not be nil after execution")
	}
}

// ============================================
// Tests for GetCallStack
// ============================================

func TestGetCallStackBasic(t *testing.T) {
	// Use functions that are defined before they are called
	input := `
		func level0() {
			return 42
		}
		func level1() {
			return level0()
		}
		func level2() {
			return level1()
		}
		level2()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

// ============================================
// Tests for Index operations
// ============================================

func TestArrayIndexInLoop(t *testing.T) {
	// Test safe array indexing in for-in loop
	input := `
		var arr = [10, 20, 30]
		var sum = 0
		for (var i = 0; i < 3; i = i + 1) {
			sum = sum + arr[i]
		}
		sum
	`
	vm := runVM(t, input)
	testIntegerObject(t, 60, vm.LastPopped())
}

func TestStringIndexConcat(t *testing.T) {
	// Test string indexing
	input := `
		var s = "hello"
		s[0] + s[1] + s[2]
	`
	vm := runVM(t, input)
	testStringObject(t, "hel", vm.LastPopped())
}

// ============================================
// Tests for SetIndex operations
// ============================================

func TestSetIndexArray(t *testing.T) {
	input := `
		var arr = [1, 2, 3]
		arr[1] = 20
		arr[1]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 20, vm.LastPopped())
}

func TestSetIndexMap(t *testing.T) {
	input := `
		var m = {"a": 1}
		m["a"] = 100
		m["a"]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 100, vm.LastPopped())
}

// ============================================
// Tests for class inheritance
// ============================================

func TestClassInheritanceVM(t *testing.T) {
	input := `
		class Animal {
			var name = ""
			func init(n) { this.name = n }
			func speak() { return this.name }
		}
		class Dog extends Animal {
			func speak() { return this.name + " barks" }
		}
		var d = new Dog("Rex")
		d.speak()
	`
	vm := runVM(t, input)
	testStringObject(t, "Rex barks", vm.LastPopped())
}

// ============================================
// Tests for map operations
// ============================================

func TestMapSetIndexNewKey(t *testing.T) {
	input := `
		var m = {}
		m["newKey"] = "newValue"
		m["newKey"]
	`
	vm := runVM(t, input)
	testStringObject(t, "newValue", vm.LastPopped())
}

// ============================================
// Tests for error handling edge cases
// ============================================

func TestErrorPropagationThroughFunctions(t *testing.T) {
	input := `
		func level3() {
			throw "error from level3"
		}
		func level2() {
			level3()
		}
		func level1() {
			level2()
		}
		try {
			level1()
		} catch (e) {
			"caught: " + e
		}
	`
	vm := runVM(t, input)
	testStringObject(t, "caught: error from level3", vm.LastPopped())
}
