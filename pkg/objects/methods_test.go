// pkg/objects/methods_test.go
package objects

import (
	"testing"
)

// ============================================================
// Integer Method Tests
// ============================================================

func TestInt_Methods(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		{"typeOf", &Int{Value: 42}, "typeOf", nil, &String{Value: "INT"}, false},
		{"toStr", &Int{Value: 42}, "toStr", nil, &String{Value: "42"}, false},
		{"toFloat", &Int{Value: 42}, "toFloat", nil, &Float{Value: 42.0}, false},
		{"toFloat negative", &Int{Value: -10}, "toFloat", nil, &Float{Value: -10.0}, false},
		{"abs positive", &Int{Value: 5}, "abs", nil, &Int{Value: 5}, false},
		{"abs negative", &Int{Value: -5}, "abs", nil, &Int{Value: 5}, false},
		{"abs zero", &Int{Value: 0}, "abs", nil, &Int{Value: 0}, false},
		{"toFloat wrong args", &Int{Value: 42}, "toFloat", []Object{&Int{Value: 1}}, nil, true},
		{"abs wrong args", &Int{Value: 42}, "abs", []Object{&Int{Value: 1}}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(IntType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for IntType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// ============================================================
// Float Method Tests
// ============================================================

func TestFloat_Methods(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		{"typeOf", &Float{Value: 3.14}, "typeOf", nil, &String{Value: "FLOAT"}, false},
		{"toStr", &Float{Value: 3.14}, "toStr", nil, &String{Value: "3.14"}, false},
		{"toInt", &Float{Value: 3.14}, "toInt", nil, &Int{Value: 3}, false},
		{"toInt truncates", &Float{Value: 3.99}, "toInt", nil, &Int{Value: 3}, false},
		{"toInt negative", &Float{Value: -3.14}, "toInt", nil, &Int{Value: -3}, false},
		{"abs positive", &Float{Value: 5.5}, "abs", nil, &Float{Value: 5.5}, false},
		{"abs negative", &Float{Value: -5.5}, "abs", nil, &Float{Value: 5.5}, false},
		{"abs zero", &Float{Value: 0.0}, "abs", nil, &Float{Value: 0.0}, false},
		{"floor", &Float{Value: 3.7}, "floor", nil, &Int{Value: 3}, false},
		{"floor negative", &Float{Value: -3.2}, "floor", nil, &Int{Value: -4}, false},
		{"ceil", &Float{Value: 3.2}, "ceil", nil, &Int{Value: 4}, false},
		{"ceil negative", &Float{Value: -3.7}, "ceil", nil, &Int{Value: -3}, false},
		{"round up", &Float{Value: 3.7}, "round", nil, &Int{Value: 4}, false},
		{"round down", &Float{Value: 3.2}, "round", nil, &Int{Value: 3}, false},
		{"round .5", &Float{Value: 3.5}, "round", nil, &Int{Value: 4}, false},
		{"toInt wrong args", &Float{Value: 3.14}, "toInt", []Object{&Int{Value: 1}}, nil, true},
		{"floor wrong args", &Float{Value: 3.14}, "floor", []Object{&Int{Value: 1}}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(FloatType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for FloatType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// ============================================================
// Bool Method Tests
// ============================================================

func TestBool_Methods(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		{"typeOf TRUE", TRUE, "typeOf", nil, &String{Value: "BOOL"}, false},
		{"typeOf FALSE", FALSE, "typeOf", nil, &String{Value: "BOOL"}, false},
		{"toStr TRUE", TRUE, "toStr", nil, &String{Value: "true"}, false},
		{"toStr FALSE", FALSE, "toStr", nil, &String{Value: "false"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(BoolType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for BoolType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// ============================================================
// Null Method Tests
// ============================================================

func TestNull_Methods(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		{"typeOf", NULL, "typeOf", nil, &String{Value: "NULL"}, false},
		{"toStr", NULL, "toStr", nil, &String{Value: "null"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(NullType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for NullType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// ============================================================
// String Method Error Tests
// ============================================================

func TestString_MethodErrors(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
	}{
		{"len wrong args", &String{Value: "hello"}, "len", []Object{&Int{Value: 1}}},
		{"upper wrong args", &String{Value: "hello"}, "upper", []Object{&Int{Value: 1}}},
		{"lower wrong args", &String{Value: "hello"}, "lower", []Object{&Int{Value: 1}}},
		{"trim wrong args", &String{Value: "hello"}, "trim", []Object{&Int{Value: 1}}},
		{"split wrong args count", &String{Value: "a,b"}, "split", nil},
		{"split wrong arg type", &String{Value: "a,b"}, "split", []Object{&Int{Value: 1}}},
		{"contains wrong args count", &String{Value: "hello"}, "contains", nil},
		{"contains wrong arg type", &String{Value: "hello"}, "contains", []Object{&Int{Value: 1}}},
		{"indexOf wrong args count", &String{Value: "hello"}, "indexOf", nil},
		{"indexOf wrong arg type", &String{Value: "hello"}, "indexOf", []Object{&Int{Value: 1}}},
		{"startsWith wrong args count", &String{Value: "hello"}, "startsWith", nil},
		{"startsWith wrong arg type", &String{Value: "hello"}, "startsWith", []Object{&Int{Value: 1}}},
		{"endsWith wrong args count", &String{Value: "hello"}, "endsWith", nil},
		{"endsWith wrong arg type", &String{Value: "hello"}, "endsWith", []Object{&Int{Value: 1}}},
		{"toInt wrong args", &String{Value: "42"}, "toInt", []Object{&Int{Value: 1}}},
		{"toInt invalid", &String{Value: "notanumber"}, "toInt", nil},
		{"toFloat wrong args", &String{Value: "3.14"}, "toFloat", []Object{&Int{Value: 1}}},
		{"toFloat invalid", &String{Value: "notanumber"}, "toFloat", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(StringType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for StringType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if !isError(result) {
				t.Errorf("expected error, got %s", result.Inspect())
			}
		})
	}
}

// ============================================================
// Array Method Extended Tests
// ============================================================

func TestArray_MethodsExtended(t *testing.T) {
	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		// Pop tests
		{"pop", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "pop", nil, &Array{Elements: []Object{&Int{Value: 1}}}, false},
		{"pop empty", &Array{Elements: []Object{}}, "pop", nil, nil, true},
		{"pop wrong args", &Array{Elements: []Object{&Int{Value: 1}}}, "pop", []Object{&Int{Value: 1}}, nil, true},
		// Push tests
		{"push wrong args count", &Array{Elements: []Object{&Int{Value: 1}}}, "push", nil, nil, true},
		// Len tests
		{"len wrong args", &Array{Elements: []Object{&Int{Value: 1}}}, "len", []Object{&Int{Value: 1}}, nil, true},
		// First tests
		{"first wrong args", &Array{Elements: []Object{&Int{Value: 1}}}, "first", []Object{&Int{Value: 1}}, nil, true},
		// Last tests
		{"last wrong args", &Array{Elements: []Object{&Int{Value: 1}}}, "last", []Object{&Int{Value: 1}}, nil, true},
		// IndexOf tests
		{"indexOf wrong args count", &Array{Elements: []Object{&Int{Value: 1}}}, "indexOf", nil, nil, true},
		// Contains tests
		{"contains wrong args count", &Array{Elements: []Object{&Int{Value: 1}}}, "contains", nil, nil, true},
		// Reverse tests
		{"reverse empty", &Array{Elements: []Object{}}, "reverse", nil, &Array{Elements: []Object{}}, false},
		{"reverse wrong args", &Array{Elements: []Object{&Int{Value: 1}}}, "reverse", []Object{&Int{Value: 1}}, nil, true},
		// Join tests
		{"join wrong args count", &Array{Elements: []Object{&String{Value: "a"}}}, "join", nil, nil, true},
		{"join wrong arg type", &Array{Elements: []Object{&String{Value: "a"}}}, "join", []Object{&Int{Value: 1}}, nil, true},
		{"join with non-strings", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "join", []Object{&String{Value: ","}}, &String{Value: "1,2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(ArrayType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for ArrayType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// ============================================================
// Map Method Extended Tests
// ============================================================

func TestMap_MethodsExtended(t *testing.T) {
	pairs := make(map[HashKey]MapPair)
	keyA := &String{Value: "a"}
	keyB := &String{Value: "b"}
	pairs[keyA.HashKey()] = MapPair{Key: keyA, Value: &Int{Value: 1}}
	pairs[keyB.HashKey()] = MapPair{Key: keyB, Value: &Int{Value: 2}}
	m := &Map{Pairs: pairs}

	tests := []struct {
		name     string
		receiver Object
		method   string
		args     []Object
		expected Object
		wantErr  bool
	}{
		// Keys tests
		{"keys", m, "keys", nil, nil, false}, // Just check it returns an array
		// Values tests
		{"values", m, "values", nil, nil, false}, // Just check it returns an array
		// Delete tests
		{"delete", m, "delete", []Object{keyA}, nil, false},
		// Error cases
		{"len wrong args", m, "len", []Object{&Int{Value: 1}}, nil, true},
		{"keys wrong args", m, "keys", []Object{&Int{Value: 1}}, nil, true},
		{"values wrong args", m, "values", []Object{&Int{Value: 1}}, nil, true},
		{"hasKey wrong args count", m, "hasKey", nil, nil, true},
		{"delete wrong args count", m, "delete", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(MapType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for MapType", tt.method)
			}

			args := []Object{tt.receiver}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			if tt.expected != nil {
				compareObjectsForTest(t, result, tt.expected)
			} else {
				// For keys/values, just verify the type
				if tt.method == "keys" || tt.method == "values" {
					arr, ok := result.(*Array)
					if !ok {
						t.Errorf("expected Array, got %s", result.Type())
					}
					if len(arr.Elements) != 2 {
						t.Errorf("expected 2 elements, got %d", len(arr.Elements))
					}
				}
				if tt.method == "delete" {
					// Verify the key was deleted
					newMap, ok := result.(*Map)
					if !ok {
						t.Errorf("expected Map, got %s", result.Type())
					}
					if len(newMap.Pairs) != 1 {
						t.Errorf("expected 1 pair after delete, got %d", len(newMap.Pairs))
					}
				}
			}
		})
	}
}

// ============================================================
// GetMethod Tests
// ============================================================

func TestGetMethod(t *testing.T) {
	// Test existing method
	method, ok := GetMethod(IntType, "typeOf")
	if !ok {
		t.Error("expected to find typeOf method for IntType")
	}
	if method == nil {
		t.Error("method should not be nil")
	}

	// Test non-existing method
	method, ok = GetMethod(IntType, "nonExistent")
	if ok {
		t.Error("expected not to find nonExistent method")
	}
	if method != nil {
		t.Error("non-existing method should be nil")
	}

	// Test non-existing type
	method, ok = GetMethod(ObjectType("NONEXISTENT"), "typeOf")
	if ok {
		t.Error("expected not to find method for non-existing type")
	}
}

// ============================================================
// Universal Method Tests
// ============================================================

func TestUniversalTypeOf(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		expected Object
		wantErr  bool
	}{
		{"int", &Int{Value: 42}, &String{Value: "INT"}, false},
		{"float", &Float{Value: 3.14}, &String{Value: "FLOAT"}, false},
		{"string", &String{Value: "hello"}, &String{Value: "STRING"}, false},
		{"bool", TRUE, &String{Value: "BOOL"}, false},
		{"null", NULL, &String{Value: "NULL"}, false},
		{"array", &Array{Elements: []Object{}}, &String{Value: "ARRAY"}, false},
		{"map", &Map{Pairs: map[HashKey]MapPair{}}, &String{Value: "MAP"}, false},
		{"error", &Error{Message: "test"}, &String{Value: "ERROR"}, false},
		{"no args", nil, nil, true},
		{"too many args", nil, nil, true},
	}

	// Test with various types
	for i, tt := range tests {
		if tt.name == "no args" {
			result := universalTypeOf()
			if !isError(result) {
				t.Errorf("test %d: expected error for no args", i)
			}
			continue
		}
		if tt.name == "too many args" {
			result := universalTypeOf(&Int{Value: 1}, &Int{Value: 2})
			if !isError(result) {
				t.Errorf("test %d: expected error for too many args", i)
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			result := universalTypeOf(tt.input)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

func TestUniversalToStr(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		expected Object
		wantErr  bool
	}{
		{"int", &Int{Value: 42}, &String{Value: "42"}, false},
		{"float", &Float{Value: 3.14}, &String{Value: "3.14"}, false},
		{"string", &String{Value: "hello"}, &String{Value: "hello"}, false},
		{"bool true", TRUE, &String{Value: "true"}, false},
		{"bool false", FALSE, &String{Value: "false"}, false},
		{"null", NULL, &String{Value: "null"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := universalToStr(tt.input)
			if tt.wantErr {
				if !isError(result) {
					t.Errorf("expected error, got %s", result.Inspect())
				}
				return
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}

	// Test error cases
	result := universalToStr()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = universalToStr(&Int{Value: 1}, &Int{Value: 2})
	if !isError(result) {
		t.Error("expected error for too many args")
	}
}
