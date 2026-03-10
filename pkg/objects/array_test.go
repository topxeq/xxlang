// pkg/objects/array_test.go
package objects

import "testing"

func TestArrayInspect(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Int{Value: 1},
		&Int{Value: 2},
		&Int{Value: 3},
	}}
	if got := arr.Inspect(); got != "[1, 2, 3]" {
		t.Errorf("Array.Inspect() = %s, want [1, 2, 3]", got)
	}
}

func TestArrayType(t *testing.T) {
	arr := &Array{Elements: []Object{}}
	if got := arr.Type(); got != ArrayType {
		t.Errorf("Array.Type() = %s, want ARRAY", got)
	}
}

func TestArrayToBool(t *testing.T) {
	empty := &Array{Elements: []Object{}}
	if empty.ToBool() != FALSE {
		t.Error("Array([]).ToBool() should be FALSE")
	}

	nonempty := &Array{Elements: []Object{&Int{Value: 1}}}
	if nonempty.ToBool() != TRUE {
		t.Error("Array([1]).ToBool() should be TRUE")
	}
}

func TestArray_Methods(t *testing.T) {
	tests := []struct {
		name     string
		input    *Array
		method   string
		args     []Object
		expected Object
	}{
		{"len", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "len", nil, &Int{Value: 2}},
		{"first", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "first", nil, &Int{Value: 1}},
		{"last", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "last", nil, &Int{Value: 2}},
		{"first empty", &Array{Elements: []Object{}}, "first", nil, NULL},
		{"last empty", &Array{Elements: []Object{}}, "last", nil, NULL},
		{"indexOf found", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "indexOf", []Object{&Int{Value: 2}}, &Int{Value: 1}},
		{"indexOf not found", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "indexOf", []Object{&Int{Value: 3}}, &Int{Value: -1}},
		{"contains true", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "contains", []Object{&Int{Value: 1}}, TRUE},
		{"contains false", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "contains", []Object{&Int{Value: 3}}, FALSE},
		{"reverse", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}, "reverse", nil, &Array{Elements: []Object{&Int{Value: 3}, &Int{Value: 2}, &Int{Value: 1}}}},
		{"push", &Array{Elements: []Object{&Int{Value: 1}}}, "push", []Object{&Int{Value: 2}}, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}},
		{"join", &Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}}}, "join", []Object{&String{Value: ","}}, &String{Value: "a,b"}},
		{"typeOf", &Array{Elements: []Object{&Int{Value: 1}}}, "typeOf", nil, &String{Value: "ARRAY"}},
		{"toStr", &Array{Elements: []Object{&Int{Value: 1}}}, "toStr", nil, &String{Value: "[1]"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(ArrayType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for ArrayType", tt.method)
			}

			// Build args with receiver as first argument
			args := []Object{tt.input}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}
