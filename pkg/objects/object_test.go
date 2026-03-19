// pkg/objects/object_test.go
package objects

import "testing"

func TestObjectType(t *testing.T) {
	tests := []struct {
		obj      Object
		expected ObjectType
	}{
		{NULL, NullType},
		{TRUE, BoolType},
		{FALSE, BoolType},
	}

	for _, tt := range tests {
		if got := tt.obj.Type(); got != tt.expected {
			t.Errorf("object.Type() = %s, want %s", got, tt.expected)
		}
	}
}

func TestNullInspect(t *testing.T) {
	if got := NULL.Inspect(); got != "null" {
		t.Errorf("NULL.Inspect() = %s, want null", got)
	}
}

func TestBoolInspect(t *testing.T) {
	if got := TRUE.Inspect(); got != "true" {
		t.Errorf("TRUE.Inspect() = %s, want true", got)
	}
	if got := FALSE.Inspect(); got != "false" {
		t.Errorf("FALSE.Inspect() = %s, want false", got)
	}
}

func TestBoolToBool(t *testing.T) {
	if TRUE.ToBool() != TRUE {
		t.Error("TRUE.ToBool() should return TRUE")
	}
	if FALSE.ToBool() != FALSE {
		t.Error("FALSE.ToBool() should return FALSE")
	}
}

func TestBoolHashKey(t *testing.T) {
	// Same values should have same hash keys
	if TRUE.HashKey() != TRUE.HashKey() {
		t.Error("TRUE hash keys should be equal")
	}
	if FALSE.HashKey() != FALSE.HashKey() {
		t.Error("FALSE hash keys should be equal")
	}
	// Different values should have different hash keys
	if TRUE.HashKey() == FALSE.HashKey() {
		t.Error("TRUE and FALSE hash keys should be different")
	}
}

func TestNullHashKey(t *testing.T) {
	if NULL.HashKey() != NULL.HashKey() {
		t.Error("NULL hash keys should be equal")
	}
}

// ============================================================
// Error Type Tests
// ============================================================

func TestErrorType(t *testing.T) {
	e := &Error{Message: "test error"}
	if got := e.Type(); got != ErrorType {
		t.Errorf("Error.Type() = %s, want ERROR", got)
	}
}

func TestErrorInspect(t *testing.T) {
	e := &Error{Message: "test error"}
	if got := e.Inspect(); got != "ERROR: test error" {
		t.Errorf("Error.Inspect() = %s, want 'ERROR: test error'", got)
	}
}

func TestErrorToBool(t *testing.T) {
	e := &Error{Message: "test error"}
	if e.ToBool() != FALSE {
		t.Error("Error.ToBool() should be FALSE")
	}
}

func TestErrorHashKey(t *testing.T) {
	e1 := &Error{Message: "error 1"}
	e2 := &Error{Message: "error 2"}
	// All errors have the same hash key (type-based, not value-based)
	if e1.HashKey() != e2.HashKey() {
		t.Error("Error hash keys should be equal")
	}
}

// ============================================================
// Return Type Tests
// ============================================================

func TestReturnType(t *testing.T) {
	r := &Return{Value: &Int{Value: 42}}
	if got := r.Type(); got != ReturnType {
		t.Errorf("Return.Type() = %s, want RETURN", got)
	}
}

func TestReturnInspect(t *testing.T) {
	r := &Return{Value: &Int{Value: 42}}
	if got := r.Inspect(); got != "42" {
		t.Errorf("Return.Inspect() = %s, want '42'", got)
	}
}

func TestReturnToBool(t *testing.T) {
	r := &Return{Value: &Int{Value: 42}}
	if r.ToBool() != TRUE {
		t.Error("Return(42).ToBool() should be TRUE")
	}

	rZero := &Return{Value: &Int{Value: 0}}
	if rZero.ToBool() != FALSE {
		t.Error("Return(0).ToBool() should be FALSE")
	}
}

func TestReturnHashKey(t *testing.T) {
	r := &Return{Value: &Int{Value: 42}}
	if r.HashKey() != (HashKey{Type: ReturnType, Value: 0}) {
		t.Error("Return.HashKey() should return constant hash")
	}
}

// ============================================================
// IsTruthy Tests
// ============================================================

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		expected bool
	}{
		{"null is falsy", NULL, false},
		{"zero is falsy", &Int{Value: 0}, false},
		{"non-zero is truthy", &Int{Value: 1}, true},
		{"empty string is falsy", &String{Value: ""}, false},
		{"non-empty string is truthy", &String{Value: "hello"}, true},
		{"empty array is falsy", &Array{Elements: []Object{}}, false},
		{"non-empty array is truthy", &Array{Elements: []Object{&Int{Value: 1}}}, true},
		{"empty map is falsy", &Map{Pairs: map[HashKey]MapPair{}}, false},
		{"non-empty map is truthy", createTestMap(), true},
		{"false is falsy", FALSE, false},
		{"true is truthy", TRUE, true},
		{"error is falsy", &Error{Message: "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTruthy(tt.input); got != tt.expected {
				t.Errorf("IsTruthy(%s) = %v, want %v", tt.input.Inspect(), got, tt.expected)
			}
		})
	}
}

func createTestMap() *Map {
	pairs := make(map[HashKey]MapPair)
	key := &String{Value: "a"}
	pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	return &Map{Pairs: pairs}
}

// ============================================================
// HashKey Tests
// ============================================================

func TestArrayHashKey(t *testing.T) {
	arr := &Array{Elements: []Object{}}
	key := arr.HashKey()
	// Array should be hashable for use as map keys
	_ = key
}

func TestFloatHashKey(t *testing.T) {
	f := &Float{Value: 3.14}
	key := f.HashKey()
	// Float should be hashable for use as map keys
	_ = key
}

func TestClassHashKey(t *testing.T) {
	c := &Class{Name: "MyClass", Methods: make(map[string]Object)}
	key := c.HashKey()
	// Class should be hashable for use as map keys
	_ = key
}

func TestClassToBool(t *testing.T) {
	c := &Class{Name: "MyClass", Methods: make(map[string]Object)}
	result := c.ToBool()
	if result != TRUE {
		t.Error("Class.ToBool() should return TRUE")
	}
}

// ============================================================
// TypeTag Tests
// ============================================================

func TestTypeTags(t *testing.T) {
	tests := []struct {
		name     string
		obj      Object
		expected TypeTag
	}{
		{"null", NULL, TagNull},
		{"true", TRUE, TagBool},
		{"false", FALSE, TagBool},
		{"int", &Int{Value: 42}, TagInt},
		{"float", &Float{Value: 3.14}, TagFloat},
		{"string", &String{Value: "hello"}, TagString},
		{"array", &Array{Elements: []Object{}}, TagArray},
		{"map", &Map{Pairs: map[HashKey]MapPair{}}, TagMap},
		{"error", &Error{Message: "test"}, TagError},
		{"return", &Return{Value: NULL}, TagReturn},
		{"class", &Class{Name: "MyClass"}, TagClass},
		{"instance", &Instance{Class: &Class{Name: "MyClass"}}, TagInstance},
		{"module", &Module{Name: "test"}, TagModule},
		{"builtin", &Builtin{Fn: func(...Object) Object { return NULL }}, TagBuiltin},
		{"compiled function", &CompiledFunction{NumLocals: 0}, TagCompiledFunction},
		{"function", &Function{Parameters: []*Identifier{}}, TagFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.obj.TypeTag(); got != tt.expected {
				t.Errorf("%s.TypeTag() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

// ============================================================
// Instance Tests
// ============================================================

func TestInstanceType(t *testing.T) {
	class := &Class{Name: "MyClass"}
	inst := &Instance{Class: class}
	if got := inst.Type(); got != InstanceType {
		t.Errorf("Instance.Type() = %s, want INSTANCE", got)
	}
}

func TestInstanceInspect(t *testing.T) {
	class := &Class{Name: "MyClass"}
	inst := &Instance{Class: class}
	if got := inst.Inspect(); got != "MyClass instance" {
		t.Errorf("Instance.Inspect() = %s, want 'MyClass instance'", got)
	}
}

func TestInstanceToBool(t *testing.T) {
	class := &Class{Name: "MyClass"}
	inst := &Instance{Class: class}
	if inst.ToBool() != TRUE {
		t.Error("Instance.ToBool() should return TRUE")
	}
}

func TestInstanceHashKey(t *testing.T) {
	class := &Class{Name: "MyClass"}
	inst := &Instance{Class: class}
	key := inst.HashKey()
	_ = key // Just verify it doesn't panic
}

// ============================================================
// Map HashKey Tests
// ============================================================

func TestMapHashKey(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{}}
	key := m.HashKey()
	_ = key // Just verify it doesn't panic
}

// ============================================================
// NewInt Tests
// ============================================================

func TestNewInt(t *testing.T) {
	// Test small values (should use cache)
	for i := int64(-10); i <= 100; i++ {
		n := NewInt(i)
		if n.Value != i {
			t.Errorf("NewInt(%d).Value = %d, want %d", i, n.Value, i)
		}
	}

	// Test large value (should not use cache)
	n := NewInt(10000)
	if n.Value != 10000 {
		t.Errorf("NewInt(10000).Value = %d, want 10000", n.Value)
	}
}

// ============================================================
// Null ToBool Tests
// ============================================================

func TestNullToBool(t *testing.T) {
	// NULL.ToBool() should return FALSE
	if NULL.ToBool() != FALSE {
		t.Error("NULL.ToBool() should return FALSE")
	}
}

// ============================================================
// Random Functions Tests
// ============================================================

func TestRandInt63(t *testing.T) {
	// Call randInt63 multiple times to test the random functions
	for i := 0; i < 10; i++ {
		n := randInt63()
		if n < 0 {
			t.Errorf("randInt63() = %d, should be non-negative", n)
		}
	}
}

// ============================================================
// NewInt Edge Cases Tests
// ============================================================

func TestNewIntCached(t *testing.T) {
	// Test values within cache range
	n1 := NewInt(0)
	n2 := NewInt(0)
	if n1 != n2 {
		t.Error("NewInt(0) should return cached value")
	}

	n1 = NewInt(100)
	n2 = NewInt(100)
	if n1 != n2 {
		t.Error("NewInt(100) should return cached value")
	}

	n1 = NewInt(-100)
	n2 = NewInt(-100)
	if n1 != n2 {
		t.Error("NewInt(-100) should return cached value")
	}
}

func TestNewIntNotCached(t *testing.T) {
	// Test values outside cache range (cache is -1000 to 100000)
	n1 := NewInt(100001)
	n2 := NewInt(100001)
	// These should be different objects (not cached)
	if n1 == n2 {
		t.Error("NewInt(100001) should not be cached")
	}
	if n1.Value != 100001 || n2.Value != 100001 {
		t.Error("NewInt values incorrect")
	}

	// Also test negative value outside cache
	n3 := NewInt(-1001)
	n4 := NewInt(-1001)
	if n3 == n4 {
		t.Error("NewInt(-1001) should not be cached")
	}
}

// ============================================================
// deepCopyObject Tests
// ============================================================

func TestDeepCopyInt(t *testing.T) {
	n := NewInt(42)
	copied := deepCopyObject(n)
	copiedInt, ok := copied.(*Int)
	if !ok {
		t.Fatalf("deepCopyObject(Int) should return Int, got %T", copied)
	}
	if copiedInt.Value != 42 {
		t.Errorf("deepCopyObject(Int).Value = %d, want 42", copiedInt.Value)
	}
	// Should be a different object for large values
	large := &Int{Value: 100000}
	copiedLarge := deepCopyObject(large)
	if large == copiedLarge {
		t.Error("deepCopyObject should create new object for large Int")
	}
}

func TestDeepCopyFloat(t *testing.T) {
	f := &Float{Value: 3.14}
	copied := deepCopyObject(f)
	copiedFloat, ok := copied.(*Float)
	if !ok {
		t.Fatalf("deepCopyObject(Float) should return Float, got %T", copied)
	}
	if copiedFloat.Value != 3.14 {
		t.Errorf("deepCopyObject(Float).Value = %f, want 3.14", copiedFloat.Value)
	}
}

func TestDeepCopyString(t *testing.T) {
	s := &String{Value: "hello"}
	copied := deepCopyObject(s)
	copiedStr, ok := copied.(*String)
	if !ok {
		t.Fatalf("deepCopyObject(String) should return String, got %T", copied)
	}
	if copiedStr.Value != "hello" {
		t.Errorf("deepCopyObject(String).Value = %s, want hello", copiedStr.Value)
	}
}

func TestDeepCopyBool(t *testing.T) {
	copiedTrue := deepCopyObject(TRUE)
	if copiedTrue != TRUE {
		t.Error("deepCopyObject(TRUE) should return TRUE")
	}

	copiedFalse := deepCopyObject(FALSE)
	if copiedFalse != FALSE {
		t.Error("deepCopyObject(FALSE) should return FALSE")
	}
}

func TestDeepCopyNull(t *testing.T) {
	copied := deepCopyObject(NULL)
	if copied != NULL {
		t.Error("deepCopyObject(NULL) should return NULL")
	}
}

func TestDeepCopyArray(t *testing.T) {
	arr := &Array{Elements: []Object{
		NewInt(1),
		NewInt(2),
		&String{Value: "test"},
	}}
	copied := deepCopyObject(arr)
	copiedArr, ok := copied.(*Array)
	if !ok {
		t.Fatalf("deepCopyObject(Array) should return Array, got %T", copied)
	}
	if len(copiedArr.Elements) != 3 {
		t.Errorf("deepCopyObject(Array) length = %d, want 3", len(copiedArr.Elements))
	}
	// Verify it's a deep copy
	if arr == copiedArr {
		t.Error("deepCopyObject should create new Array object")
	}
}

func TestDeepCopyMap(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: NewInt(1)},
		(&String{Value: "b"}).HashKey(): {Key: &String{Value: "b"}, Value: NewInt(2)},
	}}
	copied := deepCopyObject(m)
	copiedMap, ok := copied.(*Map)
	if !ok {
		t.Fatalf("deepCopyObject(Map) should return Map, got %T", copied)
	}
	if len(copiedMap.Pairs) != 2 {
		t.Errorf("deepCopyObject(Map) length = %d, want 2", len(copiedMap.Pairs))
	}
}

func TestDeepCopyBuiltin(t *testing.T) {
	// Builtins and functions should be returned as-is
	fn := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	copied := deepCopyObject(fn)
	if copied != fn {
		t.Error("deepCopyObject(Builtin) should return same object")
	}
}

// ============================================================
// shallowCopyObject Tests
// ============================================================

func TestShallowCopyArray(t *testing.T) {
	arr := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	copied := shallowCopyObject(arr)
	copiedArr, ok := copied.(*Array)
	if !ok {
		t.Fatalf("shallowCopyObject(Array) should return Array, got %T", copied)
	}
	if len(copiedArr.Elements) != 2 {
		t.Errorf("shallowCopyObject(Array) length = %d, want 2", len(copiedArr.Elements))
	}
	// Elements should be same references (shallow copy)
	if copiedArr.Elements[0] != arr.Elements[0] {
		t.Error("shallowCopyObject should not copy elements")
	}
}

func TestShallowCopyMap(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: NewInt(1)},
	}}
	copied := shallowCopyObject(m)
	copiedMap, ok := copied.(*Map)
	if !ok {
		t.Fatalf("shallowCopyObject(Map) should return Map, got %T", copied)
	}
	if len(copiedMap.Pairs) != 1 {
		t.Errorf("shallowCopyObject(Map) length = %d, want 1", len(copiedMap.Pairs))
	}
}

func TestShallowCopyPrimitive(t *testing.T) {
	// Primitives should be returned as-is
	n := NewInt(42)
	copied := shallowCopyObject(n)
	if copied != n {
		t.Error("shallowCopyObject(Int) should return same object")
	}
}

// ============================================================
// deepEquals Tests
// ============================================================

func TestDeepEqualsInts(t *testing.T) {
	a := NewInt(42)
	b := NewInt(42)
	if !deepEquals(a, b) {
		t.Error("deepEquals(42, 42) should be true")
	}

	c := NewInt(100)
	if deepEquals(a, c) {
		t.Error("deepEquals(42, 100) should be false")
	}
}

func TestDeepEqualsFloats(t *testing.T) {
	a := &Float{Value: 3.14}
	b := &Float{Value: 3.14}
	if !deepEquals(a, b) {
		t.Error("deepEquals(3.14, 3.14) should be true")
	}

	c := &Float{Value: 2.71}
	if deepEquals(a, c) {
		t.Error("deepEquals(3.14, 2.71) should be false")
	}
}

func TestDeepEqualsStrings(t *testing.T) {
	a := &String{Value: "hello"}
	b := &String{Value: "hello"}
	if !deepEquals(a, b) {
		t.Error("deepEquals('hello', 'hello') should be true")
	}

	c := &String{Value: "world"}
	if deepEquals(a, c) {
		t.Error("deepEquals('hello', 'world') should be false")
	}
}

func TestDeepEqualsBools(t *testing.T) {
	if !deepEquals(TRUE, TRUE) {
		t.Error("deepEquals(TRUE, TRUE) should be true")
	}
	if deepEquals(TRUE, FALSE) {
		t.Error("deepEquals(TRUE, FALSE) should be false")
	}
}

func TestDeepEqualsNulls(t *testing.T) {
	if !deepEquals(NULL, NULL) {
		t.Error("deepEquals(NULL, NULL) should be true")
	}
}

func TestDeepEqualsArrays(t *testing.T) {
	a := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	b := &Array{Elements: []Object{NewInt(1), NewInt(2)}}
	if !deepEquals(a, b) {
		t.Error("deepEquals([1,2], [1,2]) should be true")
	}

	c := &Array{Elements: []Object{NewInt(1), NewInt(3)}}
	if deepEquals(a, c) {
		t.Error("deepEquals([1,2], [1,3]) should be false")
	}

	d := &Array{Elements: []Object{NewInt(1)}}
	if deepEquals(a, d) {
		t.Error("deepEquals([1,2], [1]) should be false")
	}
}

func TestDeepEqualsMaps(t *testing.T) {
	a := &Map{Pairs: map[HashKey]MapPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: NewInt(1)},
	}}
	b := &Map{Pairs: map[HashKey]MapPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: NewInt(1)},
	}}
	if !deepEquals(a, b) {
		t.Error("deepEquals({a:1}, {a:1}) should be true")
	}

	c := &Map{Pairs: map[HashKey]MapPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: NewInt(2)},
	}}
	if deepEquals(a, c) {
		t.Error("deepEquals({a:1}, {a:2}) should be false")
	}
}

func TestDeepEqualsDifferentTypes(t *testing.T) {
	a := NewInt(42)
	b := &Float{Value: 42.0}
	if deepEquals(a, b) {
		t.Error("deepEquals(int, float) should be false")
	}
}

// ============================================================
// flattenArray Tests
// ============================================================

func TestFlattenArray(t *testing.T) {
	// Test flattening one level
	elements := []Object{
		NewInt(1),
		&Array{Elements: []Object{NewInt(2), NewInt(3)}},
		NewInt(4),
	}
	result := flattenArray(elements, 1)
	if len(result) != 4 {
		t.Errorf("flattenArray(depth=1) length = %d, want 4", len(result))
	}
}

func TestFlattenArrayDeep(t *testing.T) {
	// Test deep flattening
	elements := []Object{
		&Array{Elements: []Object{
			&Array{Elements: []Object{NewInt(1), NewInt(2)}},
			NewInt(3),
		}},
		NewInt(4),
	}
	result := flattenArray(elements, -1) // -1 means unlimited depth
	if len(result) != 4 {
		t.Errorf("flattenArray(depth=-1) length = %d, want 4", len(result))
	}
}

func TestFlattenArrayZeroDepth(t *testing.T) {
	// Test with depth 0 (no flattening)
	elements := []Object{
		NewInt(1),
		&Array{Elements: []Object{NewInt(2), NewInt(3)}},
		NewInt(4),
	}
	result := flattenArray(elements, 0)
	if len(result) != 3 {
		t.Errorf("flattenArray(depth=0) length = %d, want 3", len(result))
	}
}
