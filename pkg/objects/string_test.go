// pkg/objects/string_test.go
package objects

import "testing"

func TestStringInspect(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Inspect(); got != "hello" {
		t.Errorf("String.Inspect() = %s, want hello", got)
	}
}

func TestStringType(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Type(); got != StringType {
		t.Errorf("String.Type() = %s, want STRING", got)
	}
}

func TestStringToBool(t *testing.T) {
	empty := &String{Value: ""}
	if empty.ToBool() != FALSE {
		t.Error("String(\"\").ToBool() should be FALSE")
	}

	nonempty := &String{Value: "hello"}
	if nonempty.ToBool() != TRUE {
		t.Error("String(\"hello\").ToBool() should be TRUE")
	}
}

func TestStringHashKey(t *testing.T) {
	a := &String{Value: "hello"}
	b := &String{Value: "hello"}
	c := &String{Value: "world"}

	if a.HashKey() != b.HashKey() {
		t.Error("same string values should have same hash keys")
	}
	if a.HashKey() == c.HashKey() {
		t.Error("different string values should have different hash keys")
	}
}

func TestString_Methods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		method   string
		args     []Object
		expected Object
	}{
		{"len", "hello", "len", nil, &Int{Value: 5}},
		{"upper", "hello", "upper", nil, &String{Value: "HELLO"}},
		{"lower", "HELLO", "lower", nil, &String{Value: "hello"}},
		{"trim", "  hi  ", "trim", nil, &String{Value: "hi"}},
		{"contains true", "hello", "contains", []Object{&String{Value: "ell"}}, TRUE},
		{"contains false", "hello", "contains", []Object{&String{Value: "xyz"}}, FALSE},
		{"indexOf found", "hello", "indexOf", []Object{&String{Value: "l"}}, &Int{Value: 2}},
		{"indexOf not found", "hello", "indexOf", []Object{&String{Value: "z"}}, &Int{Value: -1}},
		{"startsWith true", "hello", "startsWith", []Object{&String{Value: "he"}}, TRUE},
		{"startsWith false", "hello", "startsWith", []Object{&String{Value: "lo"}}, FALSE},
		{"endsWith true", "hello", "endsWith", []Object{&String{Value: "lo"}}, TRUE},
		{"endsWith false", "hello", "endsWith", []Object{&String{Value: "he"}}, FALSE},
		{"typeOf", "hello", "typeOf", nil, &String{Value: "STRING"}},
		{"toStr", "hello", "toStr", nil, &String{Value: "hello"}},
		{"split", "a,b,c", "split", []Object{&String{Value: ","}}, &Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}, &String{Value: "c"}}}},
		{"toInt", "42", "toInt", nil, &Int{Value: 42}},
		{"toFloat", "3.14", "toFloat", nil, &Float{Value: 3.14}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{Value: tt.input}
			method, ok := GetMethod(StringType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for StringType", tt.method)
			}

			// Build args with receiver as first argument
			args := []Object{s}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

func isError(obj Object) bool {
	_, ok := obj.(*Error)
	return ok
}

func compareObjectsForTest(t *testing.T, got, expected Object) {
	switch g := got.(type) {
	case *Int:
		e, ok := expected.(*Int)
		if !ok {
			t.Errorf("expected %s, got INT", expected.Type())
			return
		}
		if g.Value != e.Value {
			t.Errorf("INT value mismatch: got %d, want %d", g.Value, e.Value)
		}
	case *Float:
		e, ok := expected.(*Float)
		if !ok {
			t.Errorf("expected %s, got FLOAT", expected.Type())
			return
		}
		if g.Value != e.Value {
			t.Errorf("FLOAT value mismatch: got %f, want %f", g.Value, e.Value)
		}
	case *String:
		e, ok := expected.(*String)
		if !ok {
			t.Errorf("expected %s, got STRING", expected.Type())
			return
		}
		if g.Value != e.Value {
			t.Errorf("STRING value mismatch: got %q, want %q", g.Value, e.Value)
		}
	case *Bool:
		e, ok := expected.(*Bool)
		if !ok {
			t.Errorf("expected %s, got BOOL", expected.Type())
			return
		}
		if g.Value != e.Value {
			t.Errorf("BOOL value mismatch: got %v, want %v", g.Value, e.Value)
		}
	case *Array:
		e, ok := expected.(*Array)
		if !ok {
			t.Errorf("expected %s, got ARRAY", expected.Type())
			return
		}
		if len(g.Elements) != len(e.Elements) {
			t.Errorf("ARRAY length mismatch: got %d, want %d", len(g.Elements), len(e.Elements))
			return
		}
		for i := range g.Elements {
			compareObjectsForTest(t, g.Elements[i], e.Elements[i])
		}
	default:
		if got.Inspect() != expected.Inspect() {
			t.Errorf("object mismatch: got %s, want %s", got.Inspect(), expected.Inspect())
		}
	}
}
