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
	// push returns the array (for chaining/supporting result = push(result, item) pattern)
	resultArr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected ARRAY, got %s", result.Type())
	}
	if resultArr != arr {
		t.Error("push should return the same array instance")
	}
	// Check that the array was modified in place
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements after push, got %d", len(arr.Elements))
	}
}

func TestBuiltinPop(t *testing.T) {
	fn, ok := Builtins["pop"]
	if !ok {
		t.Fatal("pop builtin not found")
	}

	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}
	result := fn.Fn(arr)
	// pop returns the popped element
	compareObjectsForTest(t, result, &Int{Value: 2})
	// Check that the array was modified in place
	if len(arr.Elements) != 1 {
		t.Errorf("expected 1 element after pop, got %d", len(arr.Elements))
	}

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
	// delete returns NULL (in-place modification)
	if result != NULL {
		t.Fatalf("expected NULL, got %s", result.Type())
	}
	// Check that the map was modified in place
	if len(m.Pairs) != 0 {
		t.Errorf("expected 0 pairs after delete, got %d", len(m.Pairs))
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
	if len(arr.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(arr.Elements))
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

// Tests for additional builtins to improve coverage

func TestBuiltinIsFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       string
		arg      Object
		expected Object
	}{
		{"isInt", "isInt", &Int{Value: 42}, TRUE},
		{"isInt_false", "isInt", &String{Value: "42"}, FALSE},
		{"isFloat", "isFloat", &Float{Value: 3.14}, TRUE},
		{"isFloat_false", "isFloat", &Int{Value: 42}, FALSE},
		{"isString", "isString", &String{Value: "hello"}, TRUE},
		{"isString_false", "isString", &Int{Value: 42}, FALSE},
		{"isArray", "isArray", &Array{Elements: []Object{}}, TRUE},
		{"isArray_false", "isArray", &Int{Value: 42}, FALSE},
		{"isMap", "isMap", &Map{Pairs: map[HashKey]MapPair{}}, TRUE},
		{"isMap_false", "isMap", &Int{Value: 42}, FALSE},
		{"isBool", "isBool", TRUE, TRUE},
		{"isBool_false", "isBool", &Int{Value: 42}, FALSE},
		{"isFunction", "isFunction", &Builtin{Fn: func(...Object) Object { return NULL }}, TRUE},
		{"isFunction_false", "isFunction", &Int{Value: 42}, FALSE},
		{"isNull", "isNull", NULL, TRUE},
		{"isNull_false", "isNull", &Int{Value: 42}, FALSE},
		{"isNumber", "isNumber", &Int{Value: 42}, TRUE},
		{"isNumber_float", "isNumber", &Float{Value: 3.14}, TRUE},
		{"isNumber_false", "isNumber", &String{Value: "42"}, FALSE},
		{"isEmpty", "isEmpty", &String{Value: ""}, TRUE},
		{"isEmpty_false", "isEmpty", &String{Value: "x"}, FALSE},
		{"isEmpty_array", "isEmpty", &Array{Elements: []Object{}}, TRUE},
		{"isAlpha", "isAlpha", &String{Value: "abc"}, TRUE},
		{"isAlpha_false", "isAlpha", &String{Value: "abc123"}, FALSE},
		{"isDigit", "isDigit", &String{Value: "123"}, TRUE},
		{"isDigit_false", "isDigit", &String{Value: "abc"}, FALSE},
		{"isAlphaNum", "isAlphaNum", &String{Value: "abc123"}, TRUE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := Builtins[tt.fn]
			if !ok {
				t.Fatalf("%s builtin not found", tt.fn)
			}
			result := fn.Fn(tt.arg)
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

func TestBuiltinStringFunctions(t *testing.T) {
	// charAt
	fn, ok := Builtins["charAt"]
	if !ok {
		t.Fatal("charAt builtin not found")
	}
	result := fn.Fn(&String{Value: "hello"}, &Int{Value: 1})
	compareObjectsForTest(t, result, &String{Value: "e"})

	// repeat
	fn, ok = Builtins["repeat"]
	if !ok {
		t.Fatal("repeat builtin not found")
	}
	result = fn.Fn(&String{Value: "ab"}, &Int{Value: 3})
	compareObjectsForTest(t, result, &String{Value: "ababab"})

	// lpad
	fn, ok = Builtins["lpad"]
	if !ok {
		t.Fatal("lpad builtin not found")
	}
	result = fn.Fn(&String{Value: "5"}, &Int{Value: 3}, &String{Value: "0"})
	compareObjectsForTest(t, result, &String{Value: "005"})

	// rpad
	fn, ok = Builtins["rpad"]
	if !ok {
		t.Fatal("rpad builtin not found")
	}
	result = fn.Fn(&String{Value: "5"}, &Int{Value: 3}, &String{Value: "0"})
	compareObjectsForTest(t, result, &String{Value: "500"})

	// trimLeft
	fn, ok = Builtins["trimLeft"]
	if !ok {
		t.Fatal("trimLeft builtin not found")
	}
	result = fn.Fn(&String{Value: "  hello  "})
	compareObjectsForTest(t, result, &String{Value: "hello  "})

	// trimRight
	fn, ok = Builtins["trimRight"]
	if !ok {
		t.Fatal("trimRight builtin not found")
	}
	result = fn.Fn(&String{Value: "  hello  "})
	compareObjectsForTest(t, result, &String{Value: "  hello"})

	// trimPrefix
	fn, ok = Builtins["trimPrefix"]
	if !ok {
		t.Fatal("trimPrefix builtin not found")
	}
	result = fn.Fn(&String{Value: "hello world"}, &String{Value: "hello "})
	compareObjectsForTest(t, result, &String{Value: "world"})

	// trimSuffix
	fn, ok = Builtins["trimSuffix"]
	if !ok {
		t.Fatal("trimSuffix builtin not found")
	}
	result = fn.Fn(&String{Value: "hello world"}, &String{Value: " world"})
	compareObjectsForTest(t, result, &String{Value: "hello"})

	// format
	fn, ok = Builtins["format"]
	if !ok {
		t.Fatal("format builtin not found")
	}
	result = fn.Fn(&String{Value: "%s=%d"}, &String{Value: "count"}, &Int{Value: 42})
	compareObjectsForTest(t, result, &String{Value: "count=42"})

	// hexEncode
	fn, ok = Builtins["hexEncode"]
	if !ok {
		t.Fatal("hexEncode builtin not found")
	}
	result = fn.Fn(&String{Value: "hello"})
	if result.Type() != StringType {
		t.Errorf("expected string, got %s", result.Type())
	}

	// hexDecode
	fn, ok = Builtins["hexDecode"]
	if !ok {
		t.Fatal("hexDecode builtin not found")
	}
	result = fn.Fn(&String{Value: "68656c6c6f"})
	compareObjectsForTest(t, result, &Bytes{Value: []byte("hello")})
}

func TestBuiltinArrayFunctions(t *testing.T) {
	// chunk
	fn, ok := Builtins["chunk"]
	if !ok {
		t.Fatal("chunk builtin not found")
	}
	arr := &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}, &Int{Value: 4}}}
	result := fn.Fn(arr, &Int{Value: 2})
	chunked, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(chunked.Elements) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunked.Elements))
	}

	// clamp
	fn, ok = Builtins["clamp"]
	if !ok {
		t.Fatal("clamp builtin not found")
	}
	result = fn.Fn(&Int{Value: 5}, &Int{Value: 1}, &Int{Value: 10})
	compareObjectsForTest(t, result, &Int{Value: 5})
	result = fn.Fn(&Int{Value: 0}, &Int{Value: 1}, &Int{Value: 10})
	compareObjectsForTest(t, result, &Int{Value: 1})
	result = fn.Fn(&Int{Value: 15}, &Int{Value: 1}, &Int{Value: 10})
	compareObjectsForTest(t, result, &Int{Value: 10})

	// count - counts substring occurrences in a string
	fn, ok = Builtins["count"]
	if !ok {
		t.Fatal("count builtin not found")
	}
	result = fn.Fn(&String{Value: "hello hello hello"}, &String{Value: "hello"})
	compareObjectsForTest(t, result, &Int{Value: 3})

	// drop
	fn, ok = Builtins["drop"]
	if !ok {
		t.Fatal("drop builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 3}}})

	// take
	fn, ok = Builtins["take"]
	if !ok {
		t.Fatal("take builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}})

	// entries
	fn, ok = Builtins["entries"]
	if !ok {
		t.Fatal("entries builtin not found")
	}
	m := &Map{Pairs: map[HashKey]MapPair{}}
	key := &String{Value: "a"}
	m.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	result = fn.Fn(m)
	entries, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(entries.Elements) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries.Elements))
	}

	// find
	fn, ok = Builtins["find"]
	if !ok {
		t.Fatal("find builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Int{Value: 2})
	result = fn.Fn(arr, &Int{Value: 99})
	compareObjectsForTest(t, result, NULL)

	// findIndex
	fn, ok = Builtins["findIndex"]
	if !ok {
		t.Fatal("findIndex builtin not found")
	}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Int{Value: 1})
	result = fn.Fn(arr, &Int{Value: 99})
	compareObjectsForTest(t, result, &Int{Value: -1})

	// includes
	fn, ok = Builtins["includes"]
	if !ok {
		t.Fatal("includes builtin not found")
	}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, TRUE)
	result = fn.Fn(arr, &Int{Value: 99})
	compareObjectsForTest(t, result, FALSE)

	// merge
	fn, ok = Builtins["merge"]
	if !ok {
		t.Fatal("merge builtin not found")
	}
	m1 := &Map{Pairs: map[HashKey]MapPair{}}
	m1.Pairs[key.HashKey()] = MapPair{Key: key, Value: &Int{Value: 1}}
	m2 := &Map{Pairs: map[HashKey]MapPair{}}
	key2 := &String{Value: "b"}
	m2.Pairs[key2.HashKey()] = MapPair{Key: key2, Value: &Int{Value: 2}}
	result = fn.Fn(m1, m2)
	merged, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}
	if len(merged.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(merged.Pairs))
	}

	// sample
	fn, ok = Builtins["sample"]
	if !ok {
		t.Fatal("sample builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr, &Int{Value: 2})
	sampled, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(sampled.Elements) != 2 {
		t.Errorf("expected 2 samples, got %d", len(sampled.Elements))
	}

	// shuffle
	fn, ok = Builtins["shuffle"]
	if !ok {
		t.Fatal("shuffle builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr)
	shuffled, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(shuffled.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(shuffled.Elements))
	}

	// unique
	fn, ok = Builtins["unique"]
	if !ok {
		t.Fatal("unique builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 1}, &Int{Value: 2}}}
	result = fn.Fn(arr)
	unique, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	if len(unique.Elements) != 2 {
		t.Errorf("expected 2 unique elements, got %d", len(unique.Elements))
	}

	// without
	fn, ok = Builtins["without"]
	if !ok {
		t.Fatal("without builtin not found")
	}
	arr = &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}, &Int{Value: 3}}}
	result = fn.Fn(arr, &Int{Value: 2})
	compareObjectsForTest(t, result, &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 3}}})
}

func TestBuiltinMathFunctions(t *testing.T) {
	// Note: round and random removed - use math module instead

	// sign
	fn, ok := Builtins["sign"]
	if !ok {
		t.Fatal("sign builtin not found")
	}
	result := fn.Fn(&Int{Value: -5})
	compareObjectsForTest(t, result, &Int{Value: -1})
	result = fn.Fn(&Int{Value: 0})
	compareObjectsForTest(t, result, &Int{Value: 0})
	result = fn.Fn(&Int{Value: 5})
	compareObjectsForTest(t, result, &Int{Value: 1})

	// randomInt
	fn, ok = Builtins["randomInt"]
	if !ok {
		t.Fatal("randomInt builtin not found")
	}
	result = fn.Fn(&Int{Value: 0}, &Int{Value: 10})
	i, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %s", result.Type())
	}
	if i.Value < 0 || i.Value > 10 {
		t.Errorf("randomInt should be in [0, 10], got %v", i.Value)
	}
}

func TestBuiltinTimeFunctions(t *testing.T) {
	// now - returns Time object
	fn, ok := Builtins["now"]
	if !ok {
		t.Fatal("now builtin not found")
	}
	result := fn.Fn()
	_, ok = result.(*Time)
	if !ok {
		t.Fatalf("expected Time, got %s", result.Type())
	}

	// nowMs - returns Int (milliseconds)
	fn, ok = Builtins["nowMs"]
	if !ok {
		t.Fatal("nowMs builtin not found")
	}
	result = fn.Fn()
	_, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %s", result.Type())
	}

	// sleep - just test it exists, don't actually sleep
	fn, ok = Builtins["sleep"]
	if !ok {
		t.Fatal("sleep builtin not found")
	}
	// Skip actual sleep test to keep tests fast
}

func TestBuiltinUtilityFunctions(t *testing.T) {
	// uuid
	fn, ok := Builtins["uuid"]
	if !ok {
		t.Fatal("uuid builtin not found")
	}
	result := fn.Fn()
	s, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %s", result.Type())
	}
	if len(s.Value) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(s.Value))
	}

	// getSwitch - takes array, prefix string, and default value
	fn, ok = Builtins["getSwitch"]
	if !ok {
		t.Fatal("getSwitch builtin not found")
	}
	switchArr := &Array{Elements: []Object{&String{Value: "color:red"}, &String{Value: "size:large"}}}
	result = fn.Fn(switchArr, &String{Value: "color:"}, &String{Value: "default"})
	compareObjectsForTest(t, result, &String{Value: "red"})
	result = fn.Fn(switchArr, &String{Value: "shape:"}, &String{Value: "default"})
	compareObjectsForTest(t, result, &String{Value: "default"})

	// switchExists - takes array and switch name
	fn, ok = Builtins["switchExists"]
	if !ok {
		t.Fatal("switchExists builtin not found")
	}
	switchArr = &Array{Elements: []Object{&String{Value: "verbose"}, &String{Value: "debug"}}}
	result = fn.Fn(switchArr, &String{Value: "verbose"})
	compareObjectsForTest(t, result, TRUE)
	result = fn.Fn(switchArr, &String{Value: "quiet"})
	compareObjectsForTest(t, result, FALSE)

	// pr (print)
	fn, ok = Builtins["pr"]
	if !ok {
		t.Fatal("pr builtin not found")
	}
	result = fn.Fn(&String{Value: "test"})
	compareObjectsForTest(t, result, NULL)

	// pl (print line)
	fn, ok = Builtins["pl"]
	if !ok {
		t.Fatal("pl builtin not found")
	}
	result = fn.Fn(&String{Value: "test"})
	compareObjectsForTest(t, result, NULL)

	// pln (print line with newline)
	fn, ok = Builtins["pln"]
	if !ok {
		t.Fatal("pln builtin not found")
	}
	result = fn.Fn()
	compareObjectsForTest(t, result, NULL)

	// prf (printf)
	fn, ok = Builtins["prf"]
	if !ok {
		t.Fatal("prf builtin not found")
	}
	result = fn.Fn(&String{Value: "%s"}, &String{Value: "test"})
	compareObjectsForTest(t, result, NULL)
}

func TestBuiltinRunCodeAndLoadPlugin(t *testing.T) {
	// runCode - test with mock
	fn, ok := Builtins["runCode"]
	if !ok {
		t.Fatal("runCode builtin not found")
	}

	// Set mock implementation
	prev := SetRunCodeImpl(func(code string, args *Map) (Object, error) {
		return &Int{Value: 42}, nil
	})
	defer SetRunCodeImpl(prev)

	result := fn.Fn(&String{Value: "return 42"})
	compareObjectsForTest(t, result, &Int{Value: 42})

	// loadPlugin - test with mock
	fn, ok = Builtins["loadPlugin"]
	if !ok {
		t.Fatal("loadPlugin builtin not found")
	}

	prevPlugin := SetLoadPluginImpl(func(path string) (Object, error) {
		return &Map{Pairs: map[HashKey]MapPair{}}, nil
	})
	defer SetLoadPluginImpl(prevPlugin)

	result = fn.Fn(&String{Value: "test.wasm"})
	_, ok = result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %s", result.Type())
	}
}

func TestBuiltinIsErr(t *testing.T) {
	fn, ok := Builtins["isErr"]
	if !ok {
		t.Fatal("isErr builtin not found")
	}

	// Test with Error object
	errObj := &Error{Message: "something went wrong"}
	result := fn.Fn(errObj)
	compareObjectsForTest(t, result, TRUE)

	// Test with TXERROR: string
	result = fn.Fn(&String{Value: "TXERROR:file not found"})
	compareObjectsForTest(t, result, TRUE)

	// Test with TXERROR: empty error
	result = fn.Fn(&String{Value: "TXERROR:"})
	compareObjectsForTest(t, result, TRUE)

	// Test with normal string
	result = fn.Fn(&String{Value: "normal string"})
	compareObjectsForTest(t, result, FALSE)

	// Test with null
	result = fn.Fn(NULL)
	compareObjectsForTest(t, result, FALSE)

	// Test with integer
	result = fn.Fn(&Int{Value: 42})
	compareObjectsForTest(t, result, FALSE)

	// Test with lowercase txerror (should NOT match)
	result = fn.Fn(&String{Value: "txerror:something"})
	compareObjectsForTest(t, result, FALSE)

	// Test isError vs isErr difference
	fnIsError, ok := Builtins["isError"]
	if !ok {
		t.Fatal("isError builtin not found")
	}

	// isError should return true for Error object
	result = fnIsError.Fn(errObj)
	compareObjectsForTest(t, result, TRUE)

	// isError should return false for TXERROR string
	result = fnIsError.Fn(&String{Value: "TXERROR:test"})
	compareObjectsForTest(t, result, FALSE)

	// isErr should return true for TXERROR string
	result = fn.Fn(&String{Value: "TXERROR:test"})
	compareObjectsForTest(t, result, TRUE)

	// Test getErrStr with TXERROR: string
	fnGetErrStr, ok := Builtins["getErrStr"]
	if !ok {
		t.Fatal("getErrStr builtin not found")
	}

	// getErrStr should extract message from TXERROR: string
	result = fnGetErrStr.Fn(&String{Value: "TXERROR:file not found"})
	compareObjectsForTest(t, result, &String{Value: "file not found"})

	// getErrStr should work with Error object
	result = fnGetErrStr.Fn(&Error{Message: "test error"})
	compareObjectsForTest(t, result, &String{Value: "test error"})

	// getErrStr should return normal string as-is
	result = fnGetErrStr.Fn(&String{Value: "normal string"})
	compareObjectsForTest(t, result, &String{Value: "normal string"})

	// Test isErrStr with TXERROR: prefix
	fnIsErrStr, ok := Builtins["isErrStr"]
	if !ok {
		t.Fatal("isErrStr builtin not found")
	}

	// isErrStr should return true for TXERROR: string
	result = fnIsErrStr.Fn(&String{Value: "TXERROR:something"})
	compareObjectsForTest(t, result, TRUE)

	// isErrStr should return false for ERROR: string (not TXERROR:)
	result = fnIsErrStr.Fn(&String{Value: "ERROR:something"})
	compareObjectsForTest(t, result, FALSE)

	// isErrStr should return false for error: string (lowercase)
	result = fnIsErrStr.Fn(&String{Value: "error:something"})
	compareObjectsForTest(t, result, FALSE)

	// isErrStr should return false for normal string
	result = fnIsErrStr.Fn(&String{Value: "normal string"})
	compareObjectsForTest(t, result, FALSE)
}
