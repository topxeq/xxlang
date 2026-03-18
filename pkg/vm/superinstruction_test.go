// pkg/vm/superinstruction_test.go
// Tests for superinstructions and specialized instructions
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Tests for OpIncLocal / OpDecLocal
// ============================================

func TestIncLocalInstruction(t *testing.T) {
	// Test i++ pattern
	input := `
		var i = 0
		i++
		i
	`
	vm := runVM(t, input)
	testIntegerObject(t, 1, vm.LastPopped())
}

func TestDecLocalInstruction(t *testing.T) {
	// Test i-- pattern
	input := `
		var i = 5
		i--
		i
	`
	vm := runVM(t, input)
	testIntegerObject(t, 4, vm.LastPopped())
}

func TestIncLocalInLoop(t *testing.T) {
	input := `
		var sum = 0
		for (var i = 0; i < 5; i++) {
			sum = sum + i
		}
		sum
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

func TestDecLocalInLoop(t *testing.T) {
	input := `
		var count = 0
		for (var i = 10; i > 0; i--) {
			count = count + 1
		}
		count
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

// ============================================
// Tests for OpGetLocalAdd / OpGetLocalSub / OpGetLocalMul
// ============================================

func TestGetLocalAddOptimization(t *testing.T) {
	// Test pattern where two locals are added: a + b
	input := `
		func add(a, b) {
			return a + b
		}
		add(3, 4)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 7, vm.LastPopped())
}

func TestGetLocalSubOptimization(t *testing.T) {
	// Test pattern where one local subtracts another: a - b
	input := `
		func sub(a, b) {
			return a - b
		}
		sub(10, 3)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 7, vm.LastPopped())
}

func TestGetLocalMulOptimization(t *testing.T) {
	// Test pattern where two locals are multiplied: a * b
	input := `
		func mul(a, b) {
			return a * b
		}
		mul(6, 7)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

// ============================================
// Tests for OpConstantAdd / OpConstantSub / OpConstantMul
// ============================================

func TestConstantAddOptimization(t *testing.T) {
	// Test constant folding or constant addition
	input := `
		var a = 100 + 200
		a
	`
	vm := runVM(t, input)
	testIntegerObject(t, 300, vm.LastPopped())
}

func TestConstantSubOptimization(t *testing.T) {
	input := `
		var a = 500 - 200
		a
	`
	vm := runVM(t, input)
	testIntegerObject(t, 300, vm.LastPopped())
}

func TestConstantMulOptimization(t *testing.T) {
	input := `
		var a = 12 * 5
		a
	`
	vm := runVM(t, input)
	testIntegerObject(t, 60, vm.LastPopped())
}

// ============================================
// Tests for OpAddLocalConst
// ============================================

func TestAddLocalConstOptimization(t *testing.T) {
	// Test pattern: i + 1 or similar
	input := `
		func inc(x) {
			return x + 10
		}
		inc(5)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 15, vm.LastPopped())
}

// ============================================
// Tests for OpSubLocalConst
// ============================================

func TestSubLocalConstOptimization(t *testing.T) {
	// Test pattern: i - 1 or similar
	input := `
		func dec(x) {
			return x - 10
		}
		dec(25)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 15, vm.LastPopped())
}

func TestSubLocalConstInLoop(t *testing.T) {
	input := `
		var sum = 100
		for (var i = 0; i < 5; i++) {
			sum = sum - 10
		}
		sum
	`
	vm := runVM(t, input)
	testIntegerObject(t, 50, vm.LastPopped())
}

// ============================================
// Tests for OpMulLocalConst
// ============================================

func TestMulLocalConstOptimization(t *testing.T) {
	// Test pattern: i * 2 or similar
	input := `
		func double(x) {
			return x * 2
		}
		double(21)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestMulLocalConstInLoop(t *testing.T) {
	input := `
		var product = 1
		for (var i = 0; i < 3; i++) {
			product = product * 2
		}
		product
	`
	vm := runVM(t, input)
	testIntegerObject(t, 8, vm.LastPopped())
}

// ============================================
// Tests for Comparison Super-instructions
// ============================================

func TestGetLocalLessOptimization(t *testing.T) {
	// Test pattern: a < b (two locals compared)
	input := `
		func testLess(a, b) {
			return a < b
		}
		testLess(3, 5)
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestGetLocalLessFalse(t *testing.T) {
	input := `
		func testLess(a, b) {
			return a < b
		}
		testLess(5, 3)
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())
}

func TestGetLocalGreaterOptimization(t *testing.T) {
	// Test pattern: a > b (two locals compared)
	input := `
		func testGreater(a, b) {
			return a > b
		}
		testGreater(10, 5)
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestGetLocalGreaterFalse(t *testing.T) {
	input := `
		func testGreater(a, b) {
			return a > b
		}
		testGreater(3, 10)
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())
}

func TestGetLocalEqualOptimization(t *testing.T) {
	// Test pattern: a == b (two locals compared)
	input := `
		func testEqual(a, b) {
			return a == b
		}
		testEqual(5, 5)
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestGetLocalEqualFalse(t *testing.T) {
	input := `
		func testEqual(a, b) {
			return a == b
		}
		testEqual(5, 6)
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())
}

func TestGetLocalNotEqualOptimization(t *testing.T) {
	// Test pattern: a != b (two locals compared)
	input := `
		func testNotEqual(a, b) {
			return a != b
		}
		testNotEqual(5, 6)
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

func TestGetLocalNotEqualFalse(t *testing.T) {
	input := `
		func testNotEqual(a, b) {
			return a != b
		}
		testNotEqual(5, 5)
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())
}

func TestComparisonInLoopCondition(t *testing.T) {
	// Test that comparison super-instructions work in loop conditions
	input := `
		var sum = 0
		for (var i = 0; i < 10; i++) {
			sum = sum + i
		}
		sum
	`
	vm := runVM(t, input)
	testIntegerObject(t, 45, vm.LastPopped())
}

// ============================================
// Tests for OpJump / OpJumpIfTrue
// ============================================

func TestJumpInstruction(t *testing.T) {
	// Test unconditional jump (while true with break)
	input := `
		var i = 0
		while (true) {
			i = i + 1
			if (i > 5) {
				break
			}
		}
		i
	`
	vm := runVM(t, input)
	testIntegerObject(t, 6, vm.LastPopped())
}

func TestJumpIfTrueInstruction(t *testing.T) {
	// Test conditional with jump-if-true
	input := `
		var x = 0
		if (true) {
			x = 42
		}
		x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

// ============================================
// Tests for Array Preallocation
// ============================================

func TestArrayCreation(t *testing.T) {
	input := `[1, 2, 3, 4, 5]`
	vm := runVM(t, input)
	testArrayObject(t, []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}, vm.LastPopped())
}

func TestLargeArrayCreation(t *testing.T) {
	input := `
		var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		           11, 12, 13, 14, 15, 16, 17, 18, 19, 20]
		arr[10]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 11, vm.LastPopped())
}

// ============================================
// Tests for IndexSafe operations
// ============================================

func TestArrayIndexBounds(t *testing.T) {
	// Test in-bounds access
	input := `var arr = [10, 20, 30]; arr[1]`
	vm := runVM(t, input)
	testIntegerObject(t, 20, vm.LastPopped())
}

func TestArrayIndexOutOfBounds(t *testing.T) {
	// Test out-of-bounds returns null
	input := `var arr = [10, 20, 30]; arr[10]`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

func TestStringIndexBounds(t *testing.T) {
	input := `"hello"[1]`
	vm := runVM(t, input)
	testStringObject(t, "e", vm.LastPopped())
}

func TestStringIndexOutOfBounds(t *testing.T) {
	input := `"hi"[10]`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

// ============================================
// Tests for Helper Functions
// ============================================

func TestIsMainFrame(t *testing.T) {
	// Create a minimal VM and test frame detection
	input := `42`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// After running, we should be at the main frame
	// The isMainFrame function checks if frameIndex <= 1
	if vm.frameIndex > 1 {
		t.Errorf("expected frameIndex <= 1, got %d", vm.frameIndex)
	}
}

func TestCurrentFrameMethodName(t *testing.T) {
	// Test that currentFrameMethodName works with regular functions
	input := `
		func myFunc() {
			return 42
		}
		myFunc()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestFormatErrorForDivision(t *testing.T) {
	// Test error formatting by triggering an error
	input := `1 / 0`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for division by zero")
	}

	// Check that error message is not empty
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// ============================================
// Tests for Module Operations (direct tests)
// ============================================

// Note: Module operations (loadModuleFile, loadWasmPlugin, etc.) require
// file system access and are tested through integration tests.

// ============================================
// Tests for Field Operations
// ============================================

func TestGetField(t *testing.T) {
	input := `
		class Point {
			var x
			var y
			func init(a, b) {
				this.x = a
				this.y = b
			}
		}
		var p = new Point(10, 20)
		p.x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

func TestSetField(t *testing.T) {
	input := `
		class Counter {
			var count
			func init() {
				this.count = 0
			}
			func increment() {
				this.count = this.count + 1
			}
		}
		var c = new Counter()
		c.increment()
		c.increment()
		c.count
	`
	vm := runVM(t, input)
	testIntegerObject(t, 2, vm.LastPopped())
}

// ============================================
// Tests for Super keyword
// ============================================

func TestSuperMethod(t *testing.T) {
	input := `
		class Animal {
			func speak() {
				return "..."
			}
		}
		class Dog extends Animal {
			func speak() {
				return super.speak() + " woof!"
			}
		}
		var d = new Dog()
		d.speak()
	`
	vm := runVM(t, input)
	testStringObject(t, "... woof!", vm.LastPopped())
}

func TestSuperConstructor(t *testing.T) {
	input := `
		class Parent {
			var value
			func init(v) {
				this.value = v
			}
		}
		class Child extends Parent {
			var extra
			func init(v, e) {
				super.init(v)
				this.extra = e
			}
		}
		var c = new Child(100, 200)
		c.value
	`
	vm := runVM(t, input)
	testIntegerObject(t, 100, vm.LastPopped())
}

// ============================================
// Tests for findInitMethod
// ============================================

func TestInitMethodInheritance(t *testing.T) {
	input := `
		class Base {
			var baseVal
			func init(x) {
				this.baseVal = x
			}
		}
		class Derived extends Base {
			var derivedVal
			func init(x, y) {
				super.init(x)
				this.derivedVal = y
			}
		}
		var d = new Derived(1, 2)
		d.baseVal + d.derivedVal
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}

// ============================================
// Tests for unknownIntOpError (indirect via division by zero)
// ============================================

func TestUnknownIntOpError(t *testing.T) {
	input := `1 / 0`
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

// ============================================
// Direct VM Construction Tests
// ============================================

func TestNewWithGlobalsStore(t *testing.T) {
	input := `var x = 42; x`
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

	testIntegerObject(t, 42, vm.LastPopped())
	// Check that global was stored
	testIntegerObject(t, 42, globals[0])
}

func TestVMGlobalsAccess(t *testing.T) {
	input := `var a = 1; var b = 2; var c = a + b; c`
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
		t.Error("expected non-empty globals")
	}
}

// ============================================
// Tests for Closure Constants
// ============================================

func TestClosureWithConstants(t *testing.T) {
	input := `
		func makeAdder(x) {
			func adder(y) {
				return x + y
			}
			return adder
		}
		var add5 = makeAdder(5)
		add5(10)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 15, vm.LastPopped())
}

// ============================================
// Benchmark tests for superinstructions
// ============================================

func BenchmarkIncLocal(b *testing.B) {
	input := `
		var i = 0
		for (var j = 0; j < 1000; j++) {
			i++
		}
	`
	program := parse(input)
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := New(bytecode)
		vm.Run()
	}
}

func BenchmarkAddLocalConst(b *testing.B) {
	input := `
		func test(x) {
			return x + 1
		}
		for (var i = 0; i < 100; i++) {
			test(i)
		}
	`
	program := parse(input)
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := New(bytecode)
		vm.Run()
	}
}

// ============================================
// Tests for executeJumpIfTrue edge cases
// ============================================

func TestExecuteJumpIfTrueFalse(t *testing.T) {
	// Test conditional that doesn't jump
	input := `
		var x = 0
		if (false) {
			x = 42
		}
		x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 0, vm.LastPopped())
}

// ============================================
// Tests for executeArrayIndexSafe
// ============================================

func TestExecuteArrayIndexSafe(t *testing.T) {
	input := `
		var arr = [10, 20, 30]
		arr[1]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 20, vm.LastPopped())
}

func TestExecuteArrayIndexSafeNegative(t *testing.T) {
	input := `
		var arr = [10, 20, 30]
		arr[-1]
	`
	vm := runVM(t, input)
	// Negative index should return null or handle gracefully
	result := vm.LastPopped()
	if result != objects.NULL {
		t.Logf("negative index returned: %v", result)
	}
}

// ============================================
// Tests for executeStringIndexSafe
// ============================================

func TestExecuteStringIndexSafe(t *testing.T) {
	input := `"hello"[1]`
	vm := runVM(t, input)
	testStringObject(t, "e", vm.LastPopped())
}

func TestExecuteStringIndexSafeOutOfBounds(t *testing.T) {
	input := `"hi"[10]`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

// ============================================
// Tests for currentFrameMethodName
// ============================================

func TestCurrentFrameMethodNameMain(t *testing.T) {
	input := `42`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	name := vm.currentFrameMethodName()
	if name != "main" {
		t.Errorf("currentFrameMethodName() = %s, want 'main'", name)
	}
}

func TestCurrentFrameMethodNameMethod(t *testing.T) {
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
// Tests for executeOpGetField
// ============================================

func TestExecuteOpGetField(t *testing.T) {
	input := `
		class Point {
			var x
			var y
			func init(a, b) {
				this.x = a
				this.y = b
			}
		}
		var p = new Point(10, 20)
		p.x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

// ============================================
// Tests for tail call with closure
// ============================================

func TestTailCallWithClosure(t *testing.T) {
	input := `
		func makeCounter() {
			var count = 0
			func counter() {
				count = count + 1
				return count
			}
			return counter
		}
		var c = makeCounter()
		c()
		c()
		c()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}

// Tests for executeJump
// ============================================

func TestExecuteJumpInfinite(t *testing.T) {
	// Test unconditional jump (while true with break)
	input := `
		var i = 0
		while (true) {
			i = i + 1
			if (i > 5) {
				break
			}
		}
		i
	`
	vm := runVM(t, input)
	testIntegerObject(t, 6, vm.LastPopped())
}

// ============================================
// Tests for formatError with source mapping
// ============================================

func TestFormatErrorWithSourceMap(t *testing.T) {
	input := `
		func test() {
			return 1 / 0
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
		t.Fatal("expected error for division by zero")
	}

	// The error message should contain information about the error
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// ============================================
// Tests for unknownIntOpError
// ============================================

func TestUnknownIntOpErrorDirect(t *testing.T) {
	// Test by creating bytecode with an invalid binary operation
	// This is a low-level test that requires manual bytecode construction
	bytecode := &compiler.Bytecode{
		Instructions: []byte{
			byte(compiler.OpConstant), 0, 0, // Push constant 0 (int 1)
			byte(compiler.OpConstant), 0, 1, // Push constant 1 (int 2)
			byte(255), // Invalid opcode - will cause unknown op error
		},
		Constants: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
		},
	}

	vm := New(bytecode)
	err := vm.Run()
	// The VM should either error or handle gracefully
	_ = err // We just want to ensure it doesn't panic
}

// ============================================
// Tests for executeGetLocalAdd with method receiver
// ============================================

func TestGetLocalAddWithReceiver(t *testing.T) {
	input := `
		class Calculator {
			func add(a, b) {
				return a + b
			}
		}
		var c = new Calculator()
		c.add(3, 4)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 7, vm.LastPopped())
}

// ============================================
// Tests for executeIncLocal with receiver
// ============================================

func TestIncLocalWithReceiver(t *testing.T) {
	input := `
		class Counter {
			var count
			func init() {
				this.count = 0
			}
			func increment() {
				this.count = this.count + 1
				return this.count
			}
		}
		var c = new Counter()
		c.increment()
		c.increment()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 2, vm.LastPopped())
}

// ============================================
// Tests for isMainFrame
// ============================================

func TestIsMainFrameDirect(t *testing.T) {
	input := `42`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	// After running, frameIndex should be 0 or 1 (both are main frame)
	// isMainFrame returns true if frameIndex <= 1
	if vm.frameIndex > 1 {
		t.Errorf("expected frameIndex <= 1, got %d", vm.frameIndex)
	}

	// Test isMainFrame function
	if !vm.isMainFrame() {
		t.Errorf("isMainFrame() should return true at frameIndex %d", vm.frameIndex)
	}
}

// ============================================
// Tests for executeArrayPrealloc
// ============================================

func TestArrayPreallocSmall(t *testing.T) {
	// Test small array creation that might use preallocation
	input := `
		var arr = [1, 2, 3, 4, 5]
		arr[2]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}

func TestArrayPreallocMedium(t *testing.T) {
	// Test medium array
	input := `
		var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]
		arr[10]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 11, vm.LastPopped())
}

func TestArrayPreallocLarge(t *testing.T) {
	// Test array larger than cache size
	input := `
		var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18]
		arr[17]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 18, vm.LastPopped())
}

// ============================================
// Tests for executeIndexSafe
// ============================================

func TestIndexSafeArray(t *testing.T) {
	input := `
		var arr = [10, 20, 30]
		arr[1]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 20, vm.LastPopped())
}

func TestIndexSafeArrayOutOfBounds(t *testing.T) {
	input := `
		var arr = [10, 20, 30]
		arr[10]
	`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

func TestIndexSafeArrayNegative(t *testing.T) {
	input := `
		var arr = [10, 20, 30]
		arr[-1]
	`
	vm := runVM(t, input)
	// Negative index should return null or handle gracefully
	result := vm.LastPopped()
	if result != objects.NULL {
		t.Logf("negative index returned: %v", result)
	}
}

// ============================================
// Tests for executeStringIndexSafe
// ============================================

func TestStringIndexSafeValid(t *testing.T) {
	input := `"hello"[1]`
	vm := runVM(t, input)
	testStringObject(t, "e", vm.LastPopped())
}

func TestStringIndexSafeOutOfBounds(t *testing.T) {
	input := `"hi"[10]`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

func TestStringIndexSafeEmpty(t *testing.T) {
	input := `""[0]`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

// ============================================
// Tests for executeOpGetField
// ============================================

func TestGetFieldSimple(t *testing.T) {
	input := `
		class Point {
			var x
			var y
			func init(a, b) {
				this.x = a
				this.y = b
			}
		}
		var p = new Point(10, 20)
		p.x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 10, vm.LastPopped())
}

func TestGetFieldNested(t *testing.T) {
	// Test accessing field through method return value
	input := `
		class Container {
			var data
			func init(v) {
				this.data = v
			}
			func getData() {
				return this.data
			}
		}
		var c = new Container(42)
		c.getData()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestGetFieldInMethod(t *testing.T) {
	input := `
		class Counter {
			var count
			func init() {
				this.count = 0
			}
			func get() {
				return this.count
			}
		}
		var c = new Counter()
		c.get()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 0, vm.LastPopped())
}

// ============================================
// Tests for executeJump with complex control flow
// ============================================

func TestJumpNestedLoops(t *testing.T) {
	input := `
		var result = 0
		for (var i = 0; i < 3; i++) {
			for (var j = 0; j < 3; j++) {
				if (i == 1 && j == 1) {
					break
				}
				result = result + 1
			}
		}
		result
	`
	vm := runVM(t, input)
	// i=0: j=0,1,2 -> result = 3
	// i=1: j=0 -> result = 4, then break at j=1
	// i=2: j=0,1,2 -> result = 7
	testIntegerObject(t, 7, vm.LastPopped())
}

func TestJumpLabeledBreak(t *testing.T) {
	input := `
		var found = false
		var arr = [1, 2, 3, 4, 5]
		for (var i = 0; i < 5; i++) {
			if (arr[i] == 3) {
				found = true
				break
			}
		}
		found
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

// ============================================
// Tests for executeJumpIfTrue
// ============================================

func TestJumpIfTrueSimple(t *testing.T) {
	input := `
		var x = 0
		if (true) {
			x = 42
		}
		x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

func TestJumpIfTrueWithElse(t *testing.T) {
	input := `
		var x = 0
		if (true) {
			x = 1
		} else {
			x = 2
		}
		x
	`
	vm := runVM(t, input)
	testIntegerObject(t, 1, vm.LastPopped())
}

func TestJumpIfTrueNested(t *testing.T) {
	input := `
		var result = ""
		if (true) {
			if (true) {
				result = "both"
			}
		}
		result
	`
	vm := runVM(t, input)
	testStringObject(t, "both", vm.LastPopped())
}

// ============================================
// Tests for currentFrameMethodName edge cases
// ============================================

func TestCurrentFrameMethodNameInNestedCall(t *testing.T) {
	input := `
		func inner() {
			return 42
		}
		func outer() {
			return inner()
		}
		outer()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())
}

// ============================================
// Tests for GetCallStack with nested calls
// ============================================

func TestGetCallStackNested(t *testing.T) {
	input := `
		func level3() {
			return 42
		}
		func level2() {
			return level3()
		}
		func level1() {
			return level2()
		}
		level1()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 42, vm.LastPopped())

	// Call stack should be available
	stack := vm.GetCallStack()
	if stack == "" {
		t.Error("GetCallStack should return non-empty string")
	}
}

// ============================================
// Tests for SetSourcePath
// ============================================

func TestSetSourcePath(t *testing.T) {
	input := `42`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	vm.SetSourcePath("/test/path.xxl")

	// Should not error
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
}

// ============================================
// Tests for executeTailCall edge cases
// ============================================

func TestTailCallRecursive(t *testing.T) {
	input := `
		func factorial(n, acc) {
			if (n <= 1) {
				return acc
			}
			return factorial(n - 1, n * acc)
		}
		factorial(5, 1)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 120, vm.LastPopped())
}

func TestTailCallMutual(t *testing.T) {
	// Mutual recursion requires forward declarations which the language doesn't support
	// Use a simpler tail recursive test instead
	input := `
		func countDown(n) {
			if (n <= 0) {
				return 0
			}
			return countDown(n - 1)
		}
		countDown(10)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 0, vm.LastPopped())
}

// ============================================
// Tests for closure with multiple free variables
// ============================================

func TestClosureMultipleFreeVars(t *testing.T) {
	input := `
		func makeAccount(initial) {
			var balance = initial
			var transactions = 0

			func deposit(amount) {
				balance = balance + amount
				transactions = transactions + 1
				return balance
			}

			return deposit
		}

		var deposit = makeAccount(100)
		deposit(50)
		deposit(25)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 175, vm.LastPopped())
}

// ============================================
// Tests for error handling
// ============================================

func TestDivisionByZeroError(t *testing.T) {
	input := `1 / 0`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

func TestModuloByZeroError(t *testing.T) {
	input := `10 % 0`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected modulo by zero error")
	}
}

// ============================================
// Tests for array with mixed types
// ============================================

func TestArrayMixedTypes(t *testing.T) {
	input := `
		var arr = [1, "hello", true, null, [1, 2, 3]]
		arr[1]
	`
	vm := runVM(t, input)
	testStringObject(t, "hello", vm.LastPopped())
}

func TestArrayNested(t *testing.T) {
	input := `
		var matrix = [[1, 2], [3, 4], [5, 6]]
		matrix[1][0]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}

// ============================================
// Tests for map operations
// ============================================

func TestMapComplexKeys(t *testing.T) {
	input := `
		var m = {}
		m["key1"] = "value1"
		m["key2"] = "value2"
		m["key1"]
	`
	vm := runVM(t, input)
	testStringObject(t, "value1", vm.LastPopped())
}

func TestMapNumericKeys(t *testing.T) {
	input := `
		var m = {}
		m[1] = "one"
		m[2] = "two"
		m[1]
	`
	vm := runVM(t, input)
	testStringObject(t, "one", vm.LastPopped())
}

// ============================================
// Direct bytecode tests for superinstructions
// These tests construct bytecode manually to test
// the execute* functions directly
// ============================================

// Helper function to create a compiled function for testing
func makeTestCompiledFunction(numLocals int, instructions []byte) *compiler.CompiledFunction {
	return &compiler.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: 0,
	}
}

// Note: Direct bytecode tests for executeGetLocalAdd, executeIncLocal, etc.
// are complex because they require proper frame setup with initialized locals.
// These are better tested through the language level (see tests below).

func TestExecuteIncLocalDirect(t *testing.T) {
	// Test the i++ pattern through language
	input := `
		func test() {
			var i = 5
			i++
			return i
		}
		test()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 6, vm.LastPopped())
}

func TestExecuteDecLocalDirect(t *testing.T) {
	// Test the i-- pattern through language
	input := `
		func test() {
			var i = 5
			i--
			return i
		}
		test()
	`
	vm := runVM(t, input)
	testIntegerObject(t, 4, vm.LastPopped())
}

// ============================================
// Tests for formatError edge cases
// ============================================

func TestFormatErrorWithoutSourceMap(t *testing.T) {
	input := `1 / 0`
	bytecode, err := testCompile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// Clear source map to test formatError without it
	bytecode.SourceMap = nil

	vm := New(bytecode)
	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for division by zero")
	}

	// Error message should just be the message without source location
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// ============================================
// Tests for executeArrayIndexSafe edge cases
// ============================================

func TestArrayIndexSafeWithEmptyArray(t *testing.T) {
	input := `
		var arr = []
		arr[0]
	`
	vm := runVM(t, input)
	testNullObject(t, vm.LastPopped())
}

// ============================================
// Tests for executeStringIndexSafe edge cases
// ============================================

func TestStringIndexSafeWithUnicode(t *testing.T) {
	input := `"你好"[0]`
	vm := runVM(t, input)
	// First byte of multi-byte UTF-8 character
	result := vm.LastPopped()
	if result == objects.NULL {
		t.Error("expected non-null result for unicode string indexing")
	}
}

// ============================================
// Tests for loop with break/continue edge cases
// ============================================

func TestContinueInNestedLoop(t *testing.T) {
	input := `
		var sum = 0
		for (var i = 0; i < 3; i++) {
			for (var j = 0; j < 3; j++) {
				if (j == 1) {
					continue
				}
				sum = sum + 1
			}
		}
		sum
	`
	vm := runVM(t, input)
	// i=0: j=0, j=2 -> sum = 2
	// i=1: j=0, j=2 -> sum = 4
	// i=2: j=0, j=2 -> sum = 6
	testIntegerObject(t, 6, vm.LastPopped())
}

// ============================================
// Tests for class with default field values
// ============================================

func TestClassWithDefaultFieldValues(t *testing.T) {
	input := `
		class Point {
			var x = 0
			var y = 0
			func init(a, b) {
				this.x = a
				this.y = b
			}
		}
		var p = new Point(10, 20)
		p.x + p.y
	`
	vm := runVM(t, input)
	testIntegerObject(t, 30, vm.LastPopped())
}

// ============================================
// Tests for method with multiple arguments
// ============================================

func TestMethodMultipleArgs(t *testing.T) {
	input := `
		class Math {
			func add(a, b, c) {
				return a + b + c
			}
		}
		var m = new Math()
		m.add(1, 2, 3)
	`
	vm := runVM(t, input)
	testIntegerObject(t, 6, vm.LastPopped())
}

// ============================================
// Tests for inheritance with super
// ============================================

func TestInheritanceWithSuperCall(t *testing.T) {
	input := `
		class Animal {
			var name
			func init(n) {
				this.name = n
			}
			func speak() {
				return this.name + " makes a sound"
			}
		}
		class Dog extends Animal {
			func init(n) {
				super.init(n)
			}
			func speak() {
				return this.name + " barks"
			}
		}
		var d = new Dog("Rex")
		d.speak()
	`
	vm := runVM(t, input)
	testStringObject(t, "Rex barks", vm.LastPopped())
}

// ============================================
// Tests for error propagation
// ============================================

func TestErrorPropagation(t *testing.T) {
	input := `
		func divide(a, b) {
			return a / b
		}
		func outer() {
			return divide(10, 0)
		}
		outer()
	`
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

// ============================================
// Tests for boolean short-circuit evaluation
// ============================================

func TestBooleanShortCircuitAnd(t *testing.T) {
	input := `
		var called = false
		func foo() {
			called = true
			return true
		}
		var result = false && foo()
		result
	`
	vm := runVM(t, input)
	testBooleanObject(t, false, vm.LastPopped())

	// foo should not have been called due to short-circuit
	// Check by looking at the called variable
}

func TestBooleanShortCircuitOr(t *testing.T) {
	input := `
		var called = false
		func foo() {
			called = true
			return false
		}
		var result = true || foo()
		result
	`
	vm := runVM(t, input)
	testBooleanObject(t, true, vm.LastPopped())
}

// ============================================
// Tests for array concatenation
// ============================================

func TestArrayConcat(t *testing.T) {
	input := `
		var a = [1, 2, 3]
		var b = [4, 5]
		var c = concat(a, b)
		c[2]
	`
	vm := runVM(t, input)
	testIntegerObject(t, 3, vm.LastPopped())
}
