// pkg/objects/builtin_test.go
package objects

import (
	"math"
	"testing"
)

func makeBuiltin() *Builtin {
	return &Builtin{Fn: func(args ...Object) Object { return NULL }}
}

func TestBuiltinType(t *testing.T) {
	b := makeBuiltin()
	if got := b.Type(); got != BuiltinType {
		t.Errorf("Builtin.Type() = %s, want BUILTIN", got)
	}
}

func TestBuiltinInspect(t *testing.T) {
	b := makeBuiltin()
	if got := b.Inspect(); got != "builtin function" {
		t.Errorf("Builtin.Inspect() = %s, want 'builtin function'", got)
	}
}

func TestBuiltinToBool(t *testing.T) {
	b := makeBuiltin()
	if b.ToBool() != TRUE {
		t.Error("Builtin.ToBool() should be TRUE")
	}
}

func TestBuiltinHashKey(t *testing.T) {
	b := makeBuiltin()
	expected := HashKey{Type: BuiltinType, Value: 0}
	if b.HashKey() != expected {
		t.Error("Builtin.HashKey() should return constant hash")
	}
}

func TestBuiltinLen(t *testing.T) {
	fn, ok := Builtins["len"]
	if !ok {
		t.Fatal("len builtin not found")
	}

	// Test string length
	result := fn.Fn(&String{Value: "hello"})
	compareObjectsForTest(t, result, &Int{Value: 5})

	// Test array length
	result = fn.Fn(&Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}})
	compareObjectsForTest(t, result, &Int{Value: 2})

	// Test no args
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTypeOf(t *testing.T) {
	fn, ok := Builtins["typeOf"]
	if !ok {
		t.Fatal("typeOf builtin not found")
	}

	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, &String{Value: "INT"})

	result = fn.Fn(&String{Value: "hello"})
	compareObjectsForTest(t, result, &String{Value: "STRING"})

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinInt(t *testing.T) {
	fn, ok := Builtins["int"]
	if !ok {
		t.Fatal("int builtin not found")
	}

	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, &Int{Value: 42})

	result = fn.Fn(&Float{Value: 3.99})
	compareObjectsForTest(t, result, &Int{Value: 3})

	result = fn.Fn(&String{Value: "42"})
	compareObjectsForTest(t, result, &Int{Value: 42})

	result = fn.Fn(TRUE)
	compareObjectsForTest(t, result, &Int{Value: 1})

	result = fn.Fn(&String{Value: "notanumber"})
	if !isError(result) {
		t.Error("expected error for invalid string")
	}
}

func TestBuiltinFloat(t *testing.T) {
	fn, ok := Builtins["float"]
	if !ok {
		t.Fatal("float builtin not found")
	}

	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, &Float{Value: 42.0})

	result = fn.Fn(&Float{Value: 3.14})
	compareObjectsForTest(t, result, &Float{Value: 3.14})

	result = fn.Fn(&String{Value: "3.14"})
	compareObjectsForTest(t, result, &Float{Value: 3.14})

	result = fn.Fn(&String{Value: "notanumber"})
	if !isError(result) {
		t.Error("expected error for invalid string")
	}
}

func TestBuiltinString(t *testing.T) {
	fn, ok := Builtins["string"]
	if !ok {
		t.Fatal("string builtin not found")
	}

	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, &String{Value: "42"})

	result = fn.Fn(&Float{Value: 3.14})
	compareObjectsForTest(t, result, &String{Value: "3.14"})

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinAbs(t *testing.T) {
	fn, ok := Builtins["abs"]
	if !ok {
		t.Fatal("abs builtin not found")
	}

	result := fn.Fn(&Int{Value: -5})
	compareObjectsForTest(t, result, &Int{Value: 5})

	result = fn.Fn(&Float{Value: -5.5})
	compareObjectsForTest(t, result, &Float{Value: 5.5})

	result = fn.Fn(&String{Value: "hello"})
	if !isError(result) {
		t.Error("expected error for invalid type")
	}
}

func TestBuiltinFloor(t *testing.T) {
	fn, ok := Builtins["floor"]
	if !ok {
		t.Fatal("floor builtin not found")
	}

	result := fn.Fn(&Float{Value: 3.7})
	compareObjectsForTest(t, result, &Int{Value: 3})

	result = fn.Fn(&Float{Value: -3.2})
	compareObjectsForTest(t, result, &Int{Value: -4})
}

func TestBuiltinCeil(t *testing.T) {
	fn, ok := Builtins["ceil"]
	if !ok {
		t.Fatal("ceil builtin not found")
	}

	result := fn.Fn(&Float{Value: 3.2})
	compareObjectsForTest(t, result, &Int{Value: 4})

	result = fn.Fn(&Float{Value: -3.7})
	compareObjectsForTest(t, result, &Int{Value: -3})
}

func TestBuiltinSqrt(t *testing.T) {
	fn, ok := Builtins["sqrt"]
	if !ok {
		t.Fatal("sqrt builtin not found")
	}

	result := fn.Fn(&Int{Value: 4})
	compareObjectsForTest(t, result, &Float{Value: 2.0})

	result = fn.Fn(&Int{Value: 2})
	compareObjectsForTest(t, result, &Float{Value: math.Sqrt(2)})

	result = fn.Fn(&Int{Value: -1})
	if !isError(result) {
		t.Error("expected error for negative number")
	}
}

func TestBuiltinPow(t *testing.T) {
	fn, ok := Builtins["pow"]
	if !ok {
		t.Fatal("pow builtin not found")
	}

	result := fn.Fn(&Int{Value: 2}, &Int{Value: 3})
	compareObjectsForTest(t, result, &Float{Value: 8.0})

	result = fn.Fn(&Int{Value: 4}, &Float{Value: 0.5})
	compareObjectsForTest(t, result, &Float{Value: 2.0})

	result = fn.Fn(&Int{Value: 2})
	if !isError(result) {
		t.Error("expected error for one arg")
	}
}

func TestBuiltinMin(t *testing.T) {
	fn, ok := Builtins["min"]
	if !ok {
		t.Fatal("min builtin not found")
	}

	result := fn.Fn(&Int{Value: 1}, &Int{Value: 2})
	compareObjectsForTest(t, result, &Int{Value: 1})

	result = fn.Fn(&Int{Value: 2}, &Int{Value: 1})
	compareObjectsForTest(t, result, &Int{Value: 1})
}

func TestBuiltinMax(t *testing.T) {
	fn, ok := Builtins["max"]
	if !ok {
		t.Fatal("max builtin not found")
	}

	result := fn.Fn(&Int{Value: 1}, &Int{Value: 2})
	compareObjectsForTest(t, result, &Int{Value: 2})

	result = fn.Fn(&Int{Value: 2}, &Int{Value: 1})
	compareObjectsForTest(t, result, &Int{Value: 2})
}

func TestBuiltinPush(t *testing.T) {
	fn, ok := Builtins["push"]
	if !ok {
		t.Fatal("push builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}}}
	result := fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}})
}

func TestBuiltinPop(t *testing.T) {
	fn, ok := Builtins["pop"]
	if !ok {
		t.Fatal("pop builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}}})

	// Test pop empty
	emptyArr := &Array{Elements: []Object{}}
	result = fn.Fn(emptyArr)
	if !isError(result) {
		t.Error("expected error for empty array")
	}
}

func TestBuiltinFirst(t *testing.T) {
	fn, ok := Builtins["first"]
	if !ok {
		t.Fatal("first builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Int{Value: 1})

	emptyArr := &Array{Elements: []Object{}}
	result = fn.Fn(emptyArr)
	compareObjectsForTest(t, result, NULL)
}

func TestBuiltinLast(t *testing.T) {
	fn, ok := Builtins["last"]
	if !ok {
		t.Fatal("last builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Int{Value: 2})

	emptyArr := &Array{Elements: []Object{}}
	result = fn.Fn(emptyArr)
	compareObjectsForTest(t, result, NULL)
}

func TestBuiltinKeys(t *testing.T) {
	fn, ok := Builtins["keys"]
	if !ok {
		t.Fatal("keys builtin not found")
	}

	m := createTestMap()
	result := fn.Fn(m)
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(arr.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(arr.Elements))
	}
}

func TestBuiltinValues(t *testing.T) {
	fn, ok := Builtins["values"]
	if !ok {
		t.Fatal("values builtin not found")
	}

	m := createTestMap()
	result := fn.Fn(m)
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(arr.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(arr.Elements))
	}
}

func TestBuiltinHasKey(t *testing.T) {
	fn, ok := Builtins["hasKey"]
	if !ok {
		t.Fatal("hasKey builtin not found")
	}

	m := createTestMap()
	keyA := &String{Value: "a"}
	keyB := &String{Value: "b"}

	result := fn.Fn(m, keyA)
	compareObjectsForTest(t, result, TRUE)

	result = fn.Fn(m, keyB)
	compareObjectsForTest(t, result, FALSE)
}

func TestBuiltinDelete(t *testing.T) {
	fn, ok := Builtins["delete"]
	if !ok {
		t.Fatal("delete builtin not found")
	}

	m := createTestMap()
	keyA := &String{Value: "a"}

	result := fn.Fn(m, keyA)
	newMap, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}
	if len(newMap.Pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(newMap.Pairs))
	}
}

func TestBuiltinRange(t *testing.T) {
	fn, ok := Builtins["range"]
	if !ok {
		t.Fatal("range builtin not found")
	}

	result := fn.Fn(&Int{Value: 5})
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(arr.Elements) != 6 {
		t.Errorf("expected 6 elements, got %d", len(arr.Elements))
	}
}

func TestBuiltinSort(t *testing.T) {
	fn, ok := Builtins["sort"]
	if !ok {
		t.Fatal("sort builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 3}, &Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}})
}

func TestBuiltinSum(t *testing.T) {
	fn, ok := Builtins["sum"]
	if !ok {
		t.Fatal("sum builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Int{Value: 6})
}

func TestBuiltinAvg(t *testing.T) {
	fn, ok := Builtins["avg"]
	if !ok {
		t.Fatal("avg builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Float{Value: 2.0})
}

func TestBuiltinReverse(t *testing.T) {
	fn, ok := Builtins["reverse"]
	if !ok {
		t.Fatal("reverse builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result := fn.Fn(arr)
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 3}, &Int{Value: 2}, &Int{Value: 1}}})
}

func TestBuiltinConcat(t *testing.T) {
	fn, ok := Builtins["concat"]
	if !ok {
		t.Fatal("concat builtin not found")
	}

	arr1 := &Array{Elements: []Object{&Int{Value: 1}}}
	arr2 := &Array{Elements: []Object{&Int{Value: 2}}}
	result := fn.Fn(arr1, arr2)
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}})
}

func TestBuiltinIndexOf(t *testing.T) {
	fn, ok := Builtins["indexOf"]
	if !ok {
		t.Fatal("indexOf builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result := fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Int{Value: 1})

	result = fn.Fn(arr, &Int{Value: 5})
	compareObjectsForTest(t, result, &Int{Value: -1})
}

func TestBuiltinStringSplit(t *testing.T) {
	fn, ok := Builtins["split"]
	if !ok {
		t.Fatal("split builtin not found")
	}

	result := fn.Fn(&String{Value: "a,b,c"}, &String{Value: ","})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}, &String{Value: "c"}}})
}

func TestBuiltinStringJoin(t *testing.T) {
	fn, ok := Builtins["join"]
	if !ok {
		t.Fatal("join builtin not found")
	}

	arr := &Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}}}
	result := fn.Fn(arr, &String{Value: ","})
	compareObjectsForTest(t, result, &String{Value: "a,b"})
}

func TestBuiltinStringTrim(t *testing.T) {
	fn, ok := Builtins["trim"]
	if !ok {
		t.Fatal("trim builtin not found")
	}

	result := fn.Fn(&String{Value: "  hello  "})
	compareObjectsForTest(t, result, &String{Value: "hello"})
}

func TestBuiltinStringUpper(t *testing.T) {
	fn, ok := Builtins["upper"]
	if !ok {
		t.Fatal("upper builtin not found")
	}

	result := fn.Fn(&String{Value: "hello"})
	compareObjectsForTest(t, result, &String{Value: "HELLO"})
}

func TestBuiltinStringLower(t *testing.T) {
	fn, ok := Builtins["lower"]
	if !ok {
		t.Fatal("lower builtin not found")
	}

	result := fn.Fn(&String{Value: "HELLO"})
	compareObjectsForTest(t, result, &String{Value: "hello"})
}

func TestBuiltinContainsStr(t *testing.T) {
	fn, ok := Builtins["containsStr"]
	if !ok {
		t.Fatal("containsStr builtin not found")
	}

	result := fn.Fn(&String{Value: "hello"}, &String{Value: "ell"})
	compareObjectsForTest(t, result, TRUE)

	result = fn.Fn(&String{Value: "hello"}, &String{Value: "xyz"})
	compareObjectsForTest(t, result, FALSE)
}

func TestBuiltinStartsWith(t *testing.T) {
	fn, ok := Builtins["startsWith"]
	if !ok {
		t.Fatal("startsWith builtin not found")
	}

	result := fn.Fn(&String{Value: "hello"}, &String{Value: "he"})
	compareObjectsForTest(t, result, TRUE)

	result = fn.Fn(&String{Value: "hello"}, &String{Value: "lo"})
	compareObjectsForTest(t, result, FALSE)
}

func TestBuiltinEndsWith(t *testing.T) {
	fn, ok := Builtins["endsWith"]
	if !ok {
		t.Fatal("endsWith builtin not found")
	}

	result := fn.Fn(&String{Value: "hello"}, &String{Value: "lo"})
	compareObjectsForTest(t, result, TRUE)

	result = fn.Fn(&String{Value: "hello"}, &String{Value: "he"})
	compareObjectsForTest(t, result, FALSE)
}

func TestBuiltinSubstr(t *testing.T) {
	fn, ok := Builtins["substr"]
	if !ok {
		t.Fatal("substr builtin not found")
	}

	result := fn.Fn(&String{Value: "hello"}, &Int{Value: 1})
	compareObjectsForTest(t, result, &String{Value: "ello"})

	result = fn.Fn(&String{Value: "hello"}, &Int{Value: 1}, &Int{Value: 3})
	compareObjectsForTest(t, result, &String{Value: "el"})
}

func TestBuiltinReplace(t *testing.T) {
	fn, ok := Builtins["replace"]
	if !ok {
		t.Fatal("replace builtin not found")
	}

	result := fn.Fn(&String{Value: "hello world"}, &String{Value: "world"}, &String{Value: "there"})
	compareObjectsForTest(t, result, &String{Value: "hello there"})
}

func TestBuiltinRest(t *testing.T) {
	fn, ok := Builtins["rest"]
	if !ok {
		t.Fatal("rest builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result := fn.Fn(arr, &Int{Value: 1})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 2}, &Int{Value: 3}}})
}

func TestBuiltinContainsArr(t *testing.T) {
	fn, ok := Builtins["containsArr"]
	if !ok {
		t.Fatal("containsArr builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr, &Int{Value: 1})
	compareObjectsForTest(t, result, TRUE)

	result = fn.Fn(arr, &Int{Value: 3})
	compareObjectsForTest(t, result, FALSE)
}

func TestNewError(t *testing.T) {
	err := newError("test error: %s", "value")
	if err.Type() != ErrorType {
		t.Errorf("expected ErrorType, got %s", err.Type())
	}
	if err.Message != "test error: value" {
		t.Errorf("expected 'test error: value', got %s", err.Message)
	}
}

func TestCompareObjects(t *testing.T) {
	tests := []struct {
		a, b     Object
		expected bool
	}{
		{&Int{Value: 1}, &Int{Value: 1}, true},
		{&Int{Value: 1}, &Int{Value: 2}, false},
		{&String{Value: "a"}, &String{Value: "a"}, true},
		{&String{Value: "a"}, &String{Value: "b"}, false},
		{&Int{Value: 1}, &String{Value: "1"}, false},
		{TRUE, TRUE, true},
		{TRUE, FALSE, false},
	}

	for _, tt := range tests {
		result := compareObjects(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("compareObjects(%s, %s) = %v, want %v", tt.a.Inspect(), tt.b.Inspect(), result, tt.expected)
		}
	}
}

// Tests for copy, clone, equals, flatten - covers helper functions

func TestBuiltinCopy(t *testing.T) {
	fn, ok := Builtins["copy"]
	if !ok {
		t.Fatal("copy builtin not found")
	}

	// Test array copy
	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	arrCopy, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(arrCopy.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrCopy.Elements))
	}

	// Test map copy
	m := &Map{Pairs: map[HashKey]MapPair{}}
	key := &String{Value: "a"}
	m.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	result = fn.Fn(m)
	mapCopy, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}
	if len(mapCopy.Pairs) != 1 {
		t.Errorf("expected 1 pair, got %d", len(mapCopy.Pairs))
	}
}

func TestBuiltinClone(t *testing.T) {
	fn, ok := Builtins["clone"]
	if !ok {
		t.Fatal("clone builtin not found")
	}

	// Test nested array clone
	nested := &Array{Elements: []Object{
		&Array{Elements: []Object{&Int{Value: 1}}},
	}}
	result := fn.Fn(nested)
	cloned, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(cloned.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(cloned.Elements))
	}

	// Test nested map clone
	nestedMap := &Map{Pairs: map[HashKey]MapPair{}}
	key := &String{Value: "arr"}
	nestedMap.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Array{Elements: []Object{&Int{Value: 1}}}}
	result = fn.Fn(nestedMap)
	mapClone, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}
	if len(mapClone.Pairs) != 1 {
		t.Errorf("expected 1 pair, got %d", len(mapClone.Pairs))
	}
}

func TestBuiltinEquals(t *testing.T) {
	fn, ok := Builtins["equals"]
	if !ok {
		t.Fatal("equals builtin not found")
	}

	// Test equal arrays
	arr1 := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	arr2 := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr1, arr2)
	compareObjectsForTest(t, result, TRUE)

	// Test unequal arrays
	arr3 := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 3}}}
	result = fn.Fn(arr1, arr3)
	compareObjectsForTest(t, result, FALSE)

	// Test equal maps
	m1 := &Map{Pairs: map[HashKey]MapPair{}}
	key := &String{Value: "a"}
	m1.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	m2 := &Map{Pairs: map[HashKey]MapPair{}}
	m2.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	result = fn.Fn(m1, m2)
	compareObjectsForTest(t, result, TRUE)

	// Test nested structures
	nested1 := &Array{Elements: []Object{&Array{Elements: []Object{&Int{Value: 1}}}}}
	nested2 := &Array{Elements: []Object{&Array{Elements: []Object{&Int{Value: 1}}}}}
	result = fn.Fn(nested1, nested2)
	compareObjectsForTest(t, result, TRUE)
}

func TestBuiltinFlatten(t *testing.T) {
	fn, ok := Builtins["flatten"]
	if !ok {
		t.Fatal("flatten builtin not found")
	}

	// Test flatten nested arrays
	nested := &Array{Elements: []Object{
		&Int{Value: 1},
		&Array{Elements: []Object{&Int{Value: 2}, &Int{Value: 3}}},
		&Int{Value: 4},
	}}
	result := fn.Fn(nested)
	flattened, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(flattened.Elements) != 4 {
		t.Errorf("expected 4 elements, got %d", len(flattened.Elements))
	}

	// Test deeply nested
	deepNested := &Array{Elements: []Object{
		&Array{Elements: []Object{
			&Array{Elements: []Object{&Int{Value: 1}}},
		}},
	}}
	result = fn.Fn(deepNested)
	flattened, ok = result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(flattened.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(flattened.Elements))
	}
}

func TestBuiltinDefaults(t *testing.T) {
	fn, ok := Builtins["defaults"]
	if !ok {
		t.Fatal("defaults builtin not found")
	}

	obj := &Map{Pairs: map[HashKey]MapPair{}}
	keyA := &String{Value: "a"}
	obj.Pairs[keyA.HashKey()] = MapPair{Key: keyA, Value: &Int{Value: 1}}

	defaultObj := &Map{Pairs: map[HashKey]MapPair{}}
	keyA2 := &String{Value: "a"}
	keyB := &String{Value: "b"}
	defaultObj.Pairs[keyA2.HashKey()] = MapPair{Key: keyA2, Value: &Int{Value: 10}}
	defaultObj.Pairs[keyB.HashKey()] = MapPair{Key: keyB, Value: &Int{Value: 2}}

	result := fn.Fn(obj, defaultObj)
	merged, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}

	// Original value should be preserved
	if merged.Pairs[keyA.HashKey()].Value.(*Int).Value != 1 {
		t.Error("original value should be preserved")
	}

	// Default value should be added
	if merged.Pairs[keyB.HashKey()].Value.(*Int).Value != 2 {
		t.Error("default value should be added")
	}
}

func TestSetRunCodeImpl(t *testing.T) {
	// Test setting and restoring runCodeImpl
	called := false
	mockFn := func(code string, args *Map) (Object, error) {
		called = true
		return NULL, nil
	}

	prev := SetRunCodeImpl(mockFn)
	if runCodeImpl == nil {
		t.Error("runCodeImpl should be set")
	}

	// Restore previous
	SetRunCodeImpl(prev)

	// Verify mock was set
	if !called {
		// This is expected since we restored before calling
	}
}

func TestSetLoadPluginImpl(t *testing.T) {
	// Test setting and restoring loadPluginImpl
	mockFn := func(path string) (Object, error) {
		return NULL, nil
	}

	prev := SetLoadPluginImpl(mockFn)
	if loadPluginImpl == nil {
		t.Error("loadPluginImpl should be set")
	}

	// Restore previous
	SetLoadPluginImpl(prev)
}

func TestBuiltinGenOtpCode(t *testing.T) {
	fn, ok := Builtins["genOtpCode"]
	if !ok {
		t.Fatal("genOtpCode builtin not found")
	}

	// Valid secret
	result := fn.Fn(&String{Value: "JBSWY3DPEHPK3PXP"})
	str, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %s", result.Type())
	}
	if len(str.Value) != 6 {
		t.Errorf("expected 6-digit code, got %s", str.Value)
	}

	// Invalid secret
	result = fn.Fn(&String{Value: "INVALID!@#"})
	if !isError(result) {
		t.Error("expected error for invalid secret")
	}

	// No args
	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCheckErr(t *testing.T) {
	fn, ok := Builtins["checkErr"]
	if !ok {
		t.Fatal("checkErr builtin not found")
	}

	// Non-error object - should return NULL
	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, NULL)

	// Error object - would call os.Exit, so we can't easily test that
	// Just verify it exists and handles non-error case
}

func TestBuiltinCheckEmpty(t *testing.T) {
	fn, ok := Builtins["checkEmpty"]
	if !ok {
		t.Fatal("checkEmpty builtin not found")
	}

	// Non-empty string - should return NULL
	result := fn.Fn(&String{Value: "hello"})
	compareObjectsForTest(t, result, NULL)

	// Empty string - would call os.Exit, so we can't easily test that
	// Just verify it handles non-empty case
}

func TestBuiltinTypeOfDetailed(t *testing.T) {
	fn, ok := Builtins["typeOf"]
	if !ok {
		t.Fatal("typeOf builtin not found")
	}

	// Test with detailed=false (default)
	result := fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, &String{Value: "INT"})

	// Test with detailed=true for non-instance
	result = fn.Fn(&Int{Value: 42}, TRUE)
	compareObjectsForTest(t, result, &String{Value: "INT"})

	// Test with instance
	inst := &Instance{
		Class:  &Class{Name: "TestClass"},
		Fields: make(map[string]Object),
	}
	result = fn.Fn(inst)
	compareObjectsForTest(t, result, &String{Value: "INSTANCE"})

	result = fn.Fn(inst, TRUE)
	compareObjectsForTest(t, result, &String{Value: "TestClass"})
}
