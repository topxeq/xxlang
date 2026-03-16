// pkg/stdlib/fp_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callFPFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("fp")
	if mod == nil {
		panic("fp module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestFPIdentity(t *testing.T) {
	result := callFPFunc("identity", Int(42))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("identity() should return Int, got %T", result)
	}
	if i.Value != 42 {
		t.Errorf("identity(42) = %d, want 42", i.Value)
	}
}

func TestFPConstant(t *testing.T) {
	result := callFPFunc("constant", Int(42))
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("constant() should return function, got %T", result)
	}
	// Call the returned function
	result = fn.Fn(Int(999))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("constant() result should return Int, got %T", result)
	}
	if i.Value != 42 {
		t.Errorf("constant(42)() = %d, want 42", i.Value)
	}
}

func TestFPAlwaysTrue(t *testing.T) {
	result := callFPFunc("alwaysTrue")
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("alwaysTrue() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("alwaysTrue() should return true")
	}
}

func TestFPAlwaysFalse(t *testing.T) {
	result := callFPFunc("alwaysFalse")
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("alwaysFalse() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("alwaysFalse() should return false")
	}
}

func TestFPNot(t *testing.T) {
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(true)
	})
	result := callFPFunc("not", pred)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("not() should return function, got %T", result)
	}
	result = fn.Fn()
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("not() result should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("not(true) should return false")
	}
}

func TestFPCompose(t *testing.T) {
	double := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value * 2)
	})
	addOne := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value + 1)
	})
	result := callFPFunc("compose", double, addOne)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("compose() should return function, got %T", result)
	}
	// compose(double, addOne)(5) = double(addOne(5)) = double(6) = 12
	result = fn.Fn(Int(5))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("compose result should return Int, got %T", result)
	}
	if i.Value != 12 {
		t.Errorf("compose() = %d, want 12", i.Value)
	}
}

func TestFPPipe(t *testing.T) {
	double := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value * 2)
	})
	addOne := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value + 1)
	})
	result := callFPFunc("pipe", double, addOne)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("pipe() should return function, got %T", result)
	}
	// pipe(double, addOne)(5) = addOne(double(5)) = addOne(10) = 11
	result = fn.Fn(Int(5))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("pipe result should return Int, got %T", result)
	}
	if i.Value != 11 {
		t.Errorf("pipe() = %d, want 11", i.Value)
	}
}

func TestFPEquals(t *testing.T) {
	result := callFPFunc("equals", Int(42))
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("equals() should return function, got %T", result)
	}
	result = fn.Fn(Int(42))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("equals() result should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("equals(42)(42) should return true")
	}
}

func TestFPProp(t *testing.T) {
	result := callFPFunc("prop", String("name"))
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("prop() should return function, got %T", result)
	}
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("name").HashKey(): {Key: String("name"), Value: String("John")},
	}}
	result = fn.Fn(m)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("prop() result should return String, got %T", result)
	}
	if s.Value != "John" {
		t.Errorf("prop('name') = %s, want 'John'", s.Value)
	}
}

func TestFPHead(t *testing.T) {
	result := callFPFunc("head", Array(Int(1), Int(2), Int(3)))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("head() should return Int, got %T", result)
	}
	if i.Value != 1 {
		t.Errorf("head() = %d, want 1", i.Value)
	}
}

func TestFPTail(t *testing.T) {
	result := callFPFunc("tail", Array(Int(1), Int(2), Int(3)))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("tail() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("tail() length = %d, want 2", len(arr.Elements))
	}
}

func TestFPLast(t *testing.T) {
	result := callFPFunc("last", Array(Int(1), Int(2), Int(3)))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("last() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("last() = %d, want 3", i.Value)
	}
}

func TestFPInit(t *testing.T) {
	result := callFPFunc("init", Array(Int(1), Int(2), Int(3)))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("init() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("init() length = %d, want 2", len(arr.Elements))
	}
}

func TestFPLength(t *testing.T) {
	result := callFPFunc("length", Array(Int(1), Int(2), Int(3)))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("length() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("length() = %d, want 3", i.Value)
	}
}

func TestFPIsEmpty(t *testing.T) {
	result := callFPFunc("isEmpty", Array())
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isEmpty() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isEmpty([]) should return true")
	}
}

func TestFPRange(t *testing.T) {
	result := callFPFunc("range", Int(5))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("range() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("range(5) length = %d, want 5", len(arr.Elements))
	}
}

func TestFPRangeWithStartEnd(t *testing.T) {
	result := callFPFunc("range", Int(2), Int(5))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("range() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("range(2, 5) length = %d, want 3", len(arr.Elements))
	}
}

func TestFPConcat(t *testing.T) {
	result := callFPFunc("concat", Array(Int(1)), Array(Int(2)))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("concat() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("concat() length = %d, want 2", len(arr.Elements))
	}
}

func TestFPFlatten(t *testing.T) {
	result := callFPFunc("flatten", Array(Array(Int(1), Int(2)), Array(Int(3))))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("flatten() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("flatten() length = %d, want 3", len(arr.Elements))
	}
}

func TestFPTimes(t *testing.T) {
	fn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return args[0] // Return index
	})
	result := callFPFunc("times", Int(3), fn)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("times() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("times() length = %d, want 3", len(arr.Elements))
	}
}

func TestFPMemoize(t *testing.T) {
	callCount := 0
	fn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		callCount++
		return Int(args[0].(*objects.Int).Value * 2)
	})
	result := callFPFunc("memoize", fn)
	memoFn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("memoize() should return function, got %T", result)
	}
	// Call twice with same arg
	memoFn.Fn(Int(5))
	memoFn.Fn(Int(5))
	if callCount != 1 {
		t.Errorf("memoize() should only call function once for same arg, called %d times", callCount)
	}
}

func TestFPUntil(t *testing.T) {
	// Predicate: value >= 10
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value >= 10)
	})
	// Function: increment
	inc := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value + 1)
	})
	result := callFPFunc("until", pred, inc, Int(0))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("until() should return Int, got %T", result)
	}
	if i.Value != 10 {
		t.Errorf("until() = %d, want 10", i.Value)
	}
}

// ============================================
// Tests for allPass
// ============================================

func TestFPAllPass(t *testing.T) {
	pred1 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 0)
	})
	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value < 100)
	})
	result := callFPFunc("allPass", pred1, pred2)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("allPass() should return function, got %T", result)
	}
	// Test value that passes both
	result = fn.Fn(Int(50))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("allPass() result should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("allPass() should return true when all predicates pass")
	}
	// Test value that fails one
	result = fn.Fn(Int(200))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("allPass() result should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("allPass() should return false when one predicate fails")
	}
}

func TestFPAllPassErrors(t *testing.T) {
	result := callFPFunc("allPass")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("allPass() with no args should return Error")
	}

	result = callFPFunc("allPass", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("allPass() with 1 arg should return Error")
	}

	pred := BuiltinFunc(func(...objects.Object) objects.Object { return Bool(true) })
	fn := callFPFunc("allPass", pred, Int(1)).(*objects.Builtin)
	result = fn.Fn(Int(5))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("allPass() with non-function predicate should return Error")
	}
}

// ============================================
// Tests for anyPass
// ============================================

func TestFPAnyPass(t *testing.T) {
	pred1 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value < 0)
	})
	pred2 := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Bool(args[0].(*objects.Int).Value > 100)
	})
	result := callFPFunc("anyPass", pred1, pred2)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("anyPass() should return function, got %T", result)
	}
	// Test value that fails both
	result = fn.Fn(Int(50))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("anyPass() result should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("anyPass() should return false when all predicates fail")
	}
	// Test value that passes one
	result = fn.Fn(Int(200))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("anyPass() result should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("anyPass() should return true when any predicate passes")
	}
}

func TestFPAnyPassErrors(t *testing.T) {
	result := callFPFunc("anyPass")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("anyPass() with no args should return Error")
	}

	result = callFPFunc("anyPass", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("anyPass() with 1 arg should return Error")
	}
}

// ============================================
// Tests for tap
// ============================================

func TestFPTap(t *testing.T) {
	called := false
	sideEffect := BuiltinFunc(func(args ...objects.Object) objects.Object {
		called = true
		return Null()
	})
	result := callFPFunc("tap", sideEffect)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("tap() should return function, got %T", result)
	}
	result = fn.Fn(Int(42))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("tap() result should return Int, got %T", result)
	}
	if i.Value != 42 {
		t.Errorf("tap() = %d, want 42", i.Value)
	}
	if !called {
		t.Error("tap() should have called side effect function")
	}
}

func TestFPTapErrors(t *testing.T) {
	result := callFPFunc("tap")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("tap() with no args should return Error")
	}

	result = callFPFunc("tap", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("tap() with non-function should return Error")
	}
}

// ============================================
// Tests for defaultTo
// ============================================

func TestFPDefaultTo(t *testing.T) {
	// defaultTo takes 2 args: default value and a second arg (design quirk)
	result := callFPFunc("defaultTo", Int(0), Int(0))
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("defaultTo() should return function, got %T", result)
	}
	// Test with non-null value
	result = fn.Fn(Int(42))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("defaultTo() result should return Int, got %T", result)
	}
	if i.Value != 42 {
		t.Errorf("defaultTo() = %d, want 42", i.Value)
	}
	// Test with null value
	result = fn.Fn(Null())
	i, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("defaultTo() result should return Int, got %T", result)
	}
	if i.Value != 0 {
		t.Errorf("defaultTo() with null = %d, want 0", i.Value)
	}
}

func TestFPDefaultToErrors(t *testing.T) {
	result := callFPFunc("defaultTo")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("defaultTo() with no args should return Error")
	}

	result = callFPFunc("defaultTo", Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("defaultTo() with 1 arg should return Error")
	}
}

// ============================================
// Tests for pick
// ============================================

func TestFPPick(t *testing.T) {
	keys := Array(String("a"), String("b"))
	result := callFPFunc("pick", keys)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("pick() should return function, got %T", result)
	}
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		String("c").HashKey(): {Key: String("c"), Value: Int(3)},
	}}
	result = fn.Fn(m)
	picked, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("pick() result should return Map, got %T", result)
	}
	if len(picked.Pairs) != 2 {
		t.Errorf("pick() should have 2 keys, got %d", len(picked.Pairs))
	}
}

func TestFPPickErrors(t *testing.T) {
	result := callFPFunc("pick")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pick() with no args should return Error")
	}

	result = callFPFunc("pick", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pick() with non-array should return Error")
	}
}

// ============================================
// Tests for omit
// ============================================

func TestFPOmit(t *testing.T) {
	keys := Array(String("c"))
	result := callFPFunc("omit", keys)
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("omit() should return function, got %T", result)
	}
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		String("b").HashKey(): {Key: String("b"), Value: Int(2)},
		String("c").HashKey(): {Key: String("c"), Value: Int(3)},
	}}
	result = fn.Fn(m)
	omitted, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("omit() result should return Map, got %T", result)
	}
	if len(omitted.Pairs) != 2 {
		t.Errorf("omit() should have 2 keys, got %d", len(omitted.Pairs))
	}
}

func TestFPOmitErrors(t *testing.T) {
	result := callFPFunc("omit")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("omit() with no args should return Error")
	}

	result = callFPFunc("omit", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("omit() with non-array should return Error")
	}
}

// ============================================
// Tests for additional error cases
// ============================================

func TestFPIdentityErrors(t *testing.T) {
	result := callFPFunc("identity")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("identity() with no args should return Error")
	}

	result = callFPFunc("identity", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("identity() with 2 args should return Error")
	}
}

func TestFPConstantErrors(t *testing.T) {
	result := callFPFunc("constant")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("constant() with no args should return Error")
	}

	result = callFPFunc("constant", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("constant() with 2 args should return Error")
	}
}

func TestFPNotErrors(t *testing.T) {
	result := callFPFunc("not")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("not() with no args should return Error")
	}

	result = callFPFunc("not", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("not() with non-function should return Error")
	}

	// Test not with function that doesn't return bool
	fn := BuiltinFunc(func(...objects.Object) objects.Object { return Int(42) })
	result = callFPFunc("not", fn)
	notFn := result.(*objects.Builtin)
	result = notFn.Fn(Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("not() with non-bool-returning predicate should return Error")
	}
}

func TestFPComposeErrors(t *testing.T) {
	result := callFPFunc("compose")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("compose() with no args should return Error")
	}

	result = callFPFunc("compose", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("compose() with 1 arg should return Error")
	}

	// Test compose with non-function
	fn := BuiltinFunc(func(...objects.Object) objects.Object { return Int(42) })
	composed := callFPFunc("compose", fn, Int(1)).(*objects.Builtin)
	result = composed.Fn(Int(5))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("compose() with non-function arg should return Error")
	}
}

func TestFPPipeErrors(t *testing.T) {
	result := callFPFunc("pipe")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pipe() with no args should return Error")
	}

	result = callFPFunc("pipe", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pipe() with 1 arg should return Error")
	}
}

func TestFPEqualsErrors(t *testing.T) {
	result := callFPFunc("equals")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("equals() with no args should return Error")
	}

	result = callFPFunc("equals", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("equals() with 2 args should return Error")
	}
}

func TestFPPropErrors(t *testing.T) {
	result := callFPFunc("prop")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("prop() with no args should return Error")
	}

	result = callFPFunc("prop", Int(1), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("prop() with 2 args should return Error")
	}

	// Test prop with non-map
	propFn := callFPFunc("prop", String("name")).(*objects.Builtin)
	result = propFn.Fn(Int(42))
	if result != objects.NULL {
		t.Errorf("prop() on non-map should return null, got %T", result)
	}
}

func TestFPHeadEmpty(t *testing.T) {
	result := callFPFunc("head", Array())
	if result != objects.NULL {
		t.Errorf("head() on empty array should return null, got %T", result)
	}
}

func TestFPTailSingle(t *testing.T) {
	result := callFPFunc("tail", Array(Int(1)))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("tail() should return Array, got %T", result)
	}
	if len(arr.Elements) != 0 {
		t.Errorf("tail() on single element array should be empty")
	}
}

func TestFPInitSingle(t *testing.T) {
	result := callFPFunc("init", Array(Int(1)))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("init() should return Array, got %T", result)
	}
	if len(arr.Elements) != 0 {
		t.Errorf("init() on single element array should be empty")
	}
}

func TestFPLastEmpty(t *testing.T) {
	result := callFPFunc("last", Array())
	if result != objects.NULL {
		t.Errorf("last() on empty array should return null, got %T", result)
	}
}

func TestFPLengthTypes(t *testing.T) {
	// String length
	result := callFPFunc("length", String("hello"))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("length() should return Int, got %T", result)
	}
	if i.Value != 5 {
		t.Errorf("length('hello') = %d, want 5", i.Value)
	}

	// Map length
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("a").HashKey(): {Key: String("a"), Value: Int(1)},
	}}
	result = callFPFunc("length", m)
	i, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("length() should return Int, got %T", result)
	}
	if i.Value != 1 {
		t.Errorf("length(map) = %d, want 1", i.Value)
	}
}

func TestFPLengthErrors(t *testing.T) {
	result := callFPFunc("length")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("length() with no args should return Error")
	}

	result = callFPFunc("length", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("length() with invalid type should return Error")
	}
}

func TestFPIsEmptyTypes(t *testing.T) {
	// String
	result := callFPFunc("isEmpty", String(""))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isEmpty() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isEmpty('') should return true")
	}

	// Map
	result = callFPFunc("isEmpty", &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{}})
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isEmpty() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isEmpty(empty map) should return true")
	}

	// Null
	result = callFPFunc("isEmpty", Null())
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isEmpty() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isEmpty(null) should return true")
	}

	// Other types return false
	result = callFPFunc("isEmpty", Int(42))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isEmpty() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isEmpty(int) should return false")
	}
}

func TestFPIsEmptyErrors(t *testing.T) {
	result := callFPFunc("isEmpty")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isEmpty() with no args should return Error")
	}
}

func TestFPRangeWithStep(t *testing.T) {
	result := callFPFunc("range", Int(0), Int(10), Int(2))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("range() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("range(0, 10, 2) length = %d, want 5", len(arr.Elements))
	}

	// Negative step
	result = callFPFunc("range", Int(10), Int(0), Int(-2))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("range() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("range(10, 0, -2) length = %d, want 5", len(arr.Elements))
	}
}

func TestFPRangeErrors(t *testing.T) {
	result := callFPFunc("range")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("range() with no args should return Error")
	}

	result = callFPFunc("range", String("a"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("range() with non-int should return Error")
	}

	result = callFPFunc("range", Int(0), String("a"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("range() with non-int should return Error")
	}
}

func TestFPTimesErrors(t *testing.T) {
	result := callFPFunc("times")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("times() with no args should return Error")
	}

	result = callFPFunc("times", Int(3))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("times() with 1 arg should return Error")
	}

	result = callFPFunc("times", String("a"), BuiltinFunc(func(...objects.Object) objects.Object { return Null() }))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("times() with non-int count should return Error")
	}

	result = callFPFunc("times", Int(3), Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("times() with non-function should return Error")
	}
}

func TestFPMemoizeErrors(t *testing.T) {
	result := callFPFunc("memoize")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("memoize() with no args should return Error")
	}

	result = callFPFunc("memoize", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("memoize() with non-function should return Error")
	}
}

func TestFPUntilErrors(t *testing.T) {
	result := callFPFunc("until")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("until() with no args should return Error")
	}

	pred := BuiltinFunc(func(...objects.Object) objects.Object { return Bool(false) })
	inc := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return Int(args[0].(*objects.Int).Value + 1)
	})

	result = callFPFunc("until", pred, inc)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("until() with 2 args should return Error")
	}

	result = callFPFunc("until", Int(1), inc, Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("until() with non-function predicate should return Error")
	}

	result = callFPFunc("until", pred, Int(1), Int(0))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("until() with non-function should return Error")
	}
}

func TestFPFlattenErrors(t *testing.T) {
	result := callFPFunc("flatten")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("flatten() with no args should return Error")
	}

	result = callFPFunc("flatten", Int(1))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("flatten() with non-array should return Error")
	}
}

func TestFPConcatErrors(t *testing.T) {
	result := callFPFunc("concat")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("concat() with no args should return Error")
	}

	result = callFPFunc("concat", Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("concat() with 1 arg should return Error")
	}

	result = callFPFunc("concat", Int(1), Array())
	if _, ok := result.(*objects.Error); !ok {
		t.Error("concat() with non-array should return Error")
	}
}
