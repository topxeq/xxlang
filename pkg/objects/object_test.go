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
