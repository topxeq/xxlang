// pkg/objects/function_test.go
package objects

import "testing"

func TestFunctionType(t *testing.T) {
	f := &Function{Parameters: []*Identifier{}, Body: nil, Env: nil}
	if got := f.Type(); got != FunctionType {
		t.Errorf("Function.Type() = %s, want FUNCTION", got)
	}
}

func TestFunctionInspect(t *testing.T) {
	f := &Function{
		Parameters: []*Identifier{{Value: "x"}},
		Body:       nil,
		Env:        nil,
	}
	got := f.Inspect()
	if got != "func(x) { ... }" {
		t.Errorf("Function.Inspect() = %s, want 'func(x) { ... }'", got)
	}
}

func TestFunctionInspectMultipleParams(t *testing.T) {
	f := &Function{
		Parameters: []*Identifier{{Value: "x"}, {Value: "y"}, {Value: "z"}},
		Body:       nil,
		Env:        nil,
	}
	got := f.Inspect()
	if got != "func(x, y, z) { ... }" {
		t.Errorf("Function.Inspect() = %s, want 'func(x, y, z) { ... }'", got)
	}
}

func TestFunctionInspectNoParams(t *testing.T) {
	f := &Function{
		Parameters: []*Identifier{},
		Body:       nil,
		Env:        nil,
	}
	got := f.Inspect()
	if got != "func() { ... }" {
		t.Errorf("Function.Inspect() = %s, want 'func() { ... }'", got)
	}
}

func TestFunctionToBool(t *testing.T) {
	f := &Function{Parameters: []*Identifier{}, Body: nil, Env: nil}
	if f.ToBool() != TRUE {
		t.Error("Function.ToBool() should be TRUE")
	}
}

func TestFunctionHashKey(t *testing.T) {
	f1 := &Function{Parameters: []*Identifier{}, Body: nil, Env: nil}
	f2 := &Function{Parameters: []*Identifier{}, Body: nil, Env: nil}
	// Functions have the same hash key (type-based)
	if f1.HashKey() != f2.HashKey() {
		t.Error("Function hash keys should be equal")
	}
}

// ============================================================
// Environment Tests
// ============================================================

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("NewEnvironment() returned nil")
	}
	if env.Store == nil {
		t.Error("Environment.Store should not be nil")
	}
	if env.Outer != nil {
		t.Error("NewEnvironment() should have nil Outer")
	}
}

func TestNewEnclosedEnvironment(t *testing.T) {
	outer := NewEnvironment()
	env := NewEnclosedEnvironment(outer)
	if env == nil {
		t.Fatal("NewEnclosedEnvironment() returned nil")
	}
	if env.Outer != outer {
		t.Error("EnclosedEnvironment.Outer should be the outer environment")
	}
}

func TestEnvironmentSetGet(t *testing.T) {
	env := NewEnvironment()

	// Set and get in same environment
	val := &Int{Value: 42}
	env.Set("x", val)

	got, ok := env.Get("x")
	if !ok {
		t.Fatal("expected to find x in environment")
	}
	if got != val {
		t.Errorf("env.Get('x') = %v, want %v", got, val)
	}
}

func TestEnvironmentGetNotFound(t *testing.T) {
	env := NewEnvironment()

	_, ok := env.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent variable")
	}
}

func TestEnvironmentGetFromOuter(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Int{Value: 42})

	inner := NewEnclosedEnvironment(outer)

	got, ok := inner.Get("x")
	if !ok {
		t.Fatal("expected to find x from outer environment")
	}
	if got.Inspect() != "42" {
		t.Errorf("inner.Get('x') = %s, want 42", got.Inspect())
	}
}

func TestEnvironmentShadowing(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Int{Value: 42})

	inner := NewEnclosedEnvironment(outer)
	inner.Set("x", &Int{Value: 100})

	// Inner should shadow outer
	got, ok := inner.Get("x")
	if !ok {
		t.Fatal("expected to find x")
	}
	if got.Inspect() != "100" {
		t.Errorf("inner.Get('x') = %s, want 100", got.Inspect())
	}

	// Outer should still have original value
	got, ok = outer.Get("x")
	if !ok {
		t.Fatal("expected to find x in outer")
	}
	if got.Inspect() != "42" {
		t.Errorf("outer.Get('x') = %s, want 42", got.Inspect())
	}
}

// ============================================================
// Identifier Tests
// ============================================================

func TestIdentifierString(t *testing.T) {
	id := &Identifier{Value: "myVar"}
	if id.String() != "myVar" {
		t.Errorf("Identifier.String() = %s, want 'myVar'", id.String())
	}
}

// ============================================================
// CompiledFunction Tests
// ============================================================

func TestCompiledFunctionType(t *testing.T) {
	cf := &CompiledFunction{Instructions: []byte{}, Name: "test"}
	if got := cf.Type(); got != CompiledFunctionType {
		t.Errorf("CompiledFunction.Type() = %s, want COMPILED_FUNCTION", got)
	}
}

func TestCompiledFunctionInspect(t *testing.T) {
	cf := &CompiledFunction{Instructions: []byte{}, Name: "myFunc"}
	if got := cf.Inspect(); got != "compiled function: myFunc" {
		t.Errorf("CompiledFunction.Inspect() = %s, want 'compiled function: myFunc'", got)
	}
}

func TestCompiledFunctionToBool(t *testing.T) {
	cf := &CompiledFunction{Instructions: []byte{}, Name: "test"}
	if cf.ToBool() != TRUE {
		t.Error("CompiledFunction.ToBool() should be TRUE")
	}
}

func TestCompiledFunctionHashKey(t *testing.T) {
	cf := &CompiledFunction{Instructions: []byte{}, Name: "test"}
	if cf.HashKey() != (HashKey{Type: CompiledFunctionType, Value: 0}) {
		t.Error("CompiledFunction.HashKey() should return constant hash")
	}
}
