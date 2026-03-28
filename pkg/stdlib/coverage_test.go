// pkg/stdlib/coverage_test.go
// Tests to improve code coverage for the stdlib package
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Helper functions
// ============================================

func callStdlibFunc(module, name string, args ...objects.Object) objects.Object {
	mod := Get(module)
	if mod == nil {
		return nil
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return nil
	}
	return fn.Fn(args...)
}

// ============================================
// Tests for math module
// ============================================

func TestMathModuleAbs(t *testing.T) {
	result := callStdlibFunc("math", "abs", Int(-10))
	if result == nil {
		t.Fatal("abs returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 10 {
		t.Errorf("abs(-10) = %d, want 10", intResult.Value)
	}
}

func TestMathModuleMinMax(t *testing.T) {
	// Test max with multiple arguments
	result := callStdlibFunc("math", "max", Int(1), Int(5), Int(3))
	if result == nil {
		t.Fatal("max returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("max(1, 5, 3) = %d, want 5", intResult.Value)
	}

	// Test min with multiple arguments
	result = callStdlibFunc("math", "min", Int(1), Int(5), Int(3))
	if result == nil {
		t.Fatal("min returned nil")
	}
	intResult, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("min(1, 5, 3) = %d, want 1", intResult.Value)
	}
}

func TestMathModuleCeilFloor(t *testing.T) {
	// Test ceil
	result := callStdlibFunc("math", "ceil", Float(1.5))
	if result == nil {
		t.Fatal("ceil returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 2 {
		t.Errorf("ceil(1.5) = %d, want 2", intResult.Value)
	}

	// Test floor
	result = callStdlibFunc("math", "floor", Float(1.5))
	if result == nil {
		t.Fatal("floor returned nil")
	}
	intResult, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("floor(1.5) = %d, want 1", intResult.Value)
	}
}

func TestMathModuleSqrtPow(t *testing.T) {
	// Test sqrt
	result := callStdlibFunc("math", "sqrt", Int(16))
	if result == nil {
		t.Fatal("sqrt returned nil")
	}
	floatResult, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 4.0 {
		t.Errorf("sqrt(16) = %f, want 4.0", floatResult.Value)
	}

	// Test pow
	result = callStdlibFunc("math", "pow", Int(2), Int(3))
	if result == nil {
		t.Fatal("pow returned nil")
	}
	floatResult, ok = result.(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", result)
	}
	if floatResult.Value != 8.0 {
		t.Errorf("pow(2, 3) = %f, want 8.0", floatResult.Value)
	}
}

// ============================================
// Tests for strings module
// ============================================

func TestStringModuleToUpperLower(t *testing.T) {
	result := callStdlibFunc("strings", "toUpper", String("hello"))
	if result == nil {
		t.Fatal("toUpper returned nil")
	}
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "HELLO" {
		t.Errorf("toUpper('hello') = %q, want 'HELLO'", strResult.Value)
	}

	result = callStdlibFunc("strings", "toLower", String("HELLO"))
	if result == nil {
		t.Fatal("toLower returned nil")
	}
	strResult, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("toLower('HELLO') = %q, want 'hello'", strResult.Value)
	}
}

func TestStringModuleTrim(t *testing.T) {
	result := callStdlibFunc("strings", "trim", String("  hello  "))
	if result == nil {
		t.Fatal("trim returned nil")
	}
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "hello" {
		t.Errorf("trim('  hello  ') = %q, want 'hello'", strResult.Value)
	}
}

func TestStringModuleContains(t *testing.T) {
	result := callStdlibFunc("strings", "contains", String("hello world"), String("world"))
	if result == nil {
		t.Fatal("contains returned nil")
	}
	boolResult, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("contains('hello world', 'world') should be true")
	}
}

func TestStringModuleSplitJoin(t *testing.T) {
	// Test split
	result := callStdlibFunc("strings", "split", String("a,b,c"), String(","))
	if result == nil {
		t.Fatal("split returned nil")
	}
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("split('a,b,c', ',') should return 3 elements, got %d", len(arrResult.Elements))
	}

	// Test join
	result = callStdlibFunc("strings", "join", arrResult, String("-"))
	if result == nil {
		t.Fatal("join returned nil")
	}
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "a-b-c" {
		t.Errorf("join(['a','b','c'], '-') = %q, want 'a-b-c'", strResult.Value)
	}
}

func TestStringModuleHasPrefixHasSuffix(t *testing.T) {
	// Test hasPrefix
	result := callStdlibFunc("strings", "hasPrefix", String("hello"), String("he"))
	if result == nil {
		t.Fatal("hasPrefix returned nil")
	}
	boolResult, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("hasPrefix('hello', 'he') should be true")
	}

	// Test hasSuffix
	result = callStdlibFunc("strings", "hasSuffix", String("hello"), String("lo"))
	if result == nil {
		t.Fatal("hasSuffix returned nil")
	}
	boolResult, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("hasSuffix('hello', 'lo') should be true")
	}
}

// ============================================
// Tests for array module
// ============================================

func TestArrayModuleFirstLast(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
	}}

	// Test first
	result := callStdlibFunc("array", "first", arr)
	if result == nil {
		t.Fatal("first returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("first([1,2,3]) = %d, want 1", intResult.Value)
	}

	// Test last
	result = callStdlibFunc("array", "last", arr)
	if result == nil {
		t.Fatal("last returned nil")
	}
	intResult, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("last([1,2,3]) = %d, want 3", intResult.Value)
	}
}

func TestArrayModuleContains(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
	}}

	result := callStdlibFunc("array", "contains", arr, Int(2))
	if result == nil {
		t.Fatal("contains returned nil")
	}
	boolResult, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("contains([1,2,3], 2) should be true")
	}
}

func TestArrayModuleIndexOf(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
	}}

	result := callStdlibFunc("array", "indexOf", arr, Int(2))
	if result == nil {
		t.Fatal("indexOf returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("indexOf([1,2,3], 2) = %d, want 1", intResult.Value)
	}
}

func TestArrayModuleSlice(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
		&objects.Int{Value: 4},
		&objects.Int{Value: 5},
	}}

	result := callStdlibFunc("array", "slice", arr, Int(1), Int(4))
	if result == nil {
		t.Fatal("slice returned nil")
	}
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("slice([1,2,3,4,5], 1, 4) should have 3 elements, got %d", len(arrResult.Elements))
	}
}

func TestArrayModuleReverse(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
	}}

	result := callStdlibFunc("array", "reverse", arr)
	if result == nil {
		t.Fatal("reverse returned nil")
	}

	// Check that original array is reversed
	first := arr.Elements[0].(*objects.Int)
	if first.Value != 3 {
		t.Errorf("after reverse, first element should be 3, got %d", first.Value)
	}
}

// ============================================
// Tests for json module
// ============================================

func TestJsonModuleEncode(t *testing.T) {
	obj := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		(&objects.String{Value: "name"}).HashKey(): {
			Key:   &objects.String{Value: "name"},
			Value: &objects.String{Value: "test"},
		},
	}}

	result := callStdlibFunc("json", "encode", obj)
	if result == nil {
		t.Fatal("json.encode returned nil")
	}
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("json.encode should return non-empty string")
	}
}

func TestJsonModuleDecode(t *testing.T) {
	jsonStr := `{"name": "test", "value": 42}`

	result := callStdlibFunc("json", "decode", String(jsonStr))
	if result == nil {
		t.Fatal("json.decode returned nil")
	}
	mapResult, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}
	if len(mapResult.Pairs) != 2 {
		t.Errorf("json.decode should return map with 2 pairs, got %d", len(mapResult.Pairs))
	}
}

// ============================================
// Tests for io module
// ============================================

func TestIoModuleCwd(t *testing.T) {
	result := callStdlibFunc("io", "cwd")
	if result == nil {
		t.Fatal("cwd returned nil")
	}
	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("cwd should return non-empty string")
	}
}

func TestIoModuleExists(t *testing.T) {
	result := callStdlibFunc("io", "exists", String("."))
	if result == nil {
		t.Fatal("exists returned nil")
	}
	boolResult, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("exists('.') should be true")
	}
}

// ============================================
// Tests for fp module
// ============================================

func TestFpModuleIdentity(t *testing.T) {
	result := callStdlibFunc("fp", "identity", Int(42))
	if result == nil {
		t.Fatal("identity returned nil")
	}
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 42 {
		t.Errorf("identity(42) = %d, want 42", intResult.Value)
	}
}

func TestFpModuleConstant(t *testing.T) {
	result := callStdlibFunc("fp", "constant", Int(42))
	if result == nil {
		t.Fatal("constant returned nil")
	}
	fn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("constant should return function, got %T", result)
	}
	// Call the returned function
	result = fn.Fn(Int(999))
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("constant() result should return Int, got %T", result)
	}
	if intResult.Value != 42 {
		t.Errorf("constant(42)() = %d, want 42", intResult.Value)
	}
}

func TestFpModuleCompose(t *testing.T) {
	// Create two simple functions
	doubleFn := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if i, ok := args[0].(*objects.Int); ok {
				return Int(i.Value * 2)
			}
			return nil
		},
	}
	addOneFn := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if i, ok := args[0].(*objects.Int); ok {
				return Int(i.Value + 1)
			}
			return nil
		},
	}

	result := callStdlibFunc("fp", "compose", doubleFn, addOneFn)
	if result == nil {
		t.Fatal("compose returned nil")
	}

	// Call the composed function
	composedFn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("compose should return function, got %T", result)
	}

	result = composedFn.Fn(Int(5))
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	// (5 + 1) * 2 = 12
	if intResult.Value != 12 {
		t.Errorf("compose(double, addOne)(5) = %d, want 12", intResult.Value)
	}
}

func TestFpModulePipe(t *testing.T) {
	// Create two simple functions
	doubleFn := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if i, ok := args[0].(*objects.Int); ok {
				return Int(i.Value * 2)
			}
			return nil
		},
	}
	addOneFn := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if i, ok := args[0].(*objects.Int); ok {
				return Int(i.Value + 1)
			}
			return nil
		},
	}

	result := callStdlibFunc("fp", "pipe", doubleFn, addOneFn)
	if result == nil {
		t.Fatal("pipe returned nil")
	}

	// Call the piped function
	pipedFn, ok := result.(*objects.Builtin)
	if !ok {
		t.Fatalf("pipe should return function, got %T", result)
	}

	result = pipedFn.Fn(Int(5))
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	// (5 * 2) + 1 = 11
	if intResult.Value != 11 {
		t.Errorf("pipe(double, addOne)(5) = %d, want 11", intResult.Value)
	}
}

// ============================================
// Tests for regex module
// ============================================

func TestRegexModuleMatch(t *testing.T) {
	result := callStdlibFunc("regex", "match", String(`\d+`), String("abc123def"))
	if result == nil {
		t.Fatal("regex.match returned nil")
	}
	boolResult, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("regex.match('\\d+', 'abc123def') should be true")
	}
}
