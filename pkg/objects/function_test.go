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
