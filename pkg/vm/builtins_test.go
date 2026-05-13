// pkg/vm/builtins_test.go
// Tests for builtin function support
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// getBuiltinByName Tests
// ============================================

func TestGetBuiltinByNameValid(t *testing.T) {
	names := []string{
		"len", "pr", "pln", "typeOf", "abs",
		"range", "sort", "substr", "split", "join",
		"trim", "upper", "lower", "floor", "ceil",
		"sqrt", "pow", "min", "max",
	}

	for _, name := range names {
		builtin := getBuiltinByName(name)
		if builtin == nil {
			t.Errorf("getBuiltinByName(%q) returned nil", name)
		}
	}
}

func TestGetBuiltinByNameInvalid(t *testing.T) {
	builtin := getBuiltinByName("")
	if builtin != nil {
		t.Error("getBuiltinByName('') should return nil")
	}

	builtin = getBuiltinByName("nonExistentBuiltin")
	if builtin != nil {
		t.Error("getBuiltinByName('nonExistentBuiltin') should return nil")
	}
}

// ============================================
// Name-based Builtin Lookup Tests
// ============================================

func TestGetBuiltinByNameLen(t *testing.T) {
	builtin := getBuiltinByName("len")
	if builtin == nil {
		t.Error("getBuiltinByName('len') returned nil")
	}

	if builtin.Fn == nil {
		t.Error("Builtin function is nil")
	}
}

// ============================================
// Builtin Function Execution Tests
// ============================================

func TestBuiltinLen(t *testing.T) {
	builtin := getBuiltinByName("len")
	if builtin == nil {
		t.Fatal("len builtin not found")
	}

	strObj := &objects.String{Value: "hello"}
	result := builtin.Fn(strObj)

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("Expected *objects.Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("len('hello') = %d, expected 5", intResult.Value)
	}

	arrObj := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
		&objects.Int{Value: 3},
	}}
	result = builtin.Fn(arrObj)

	intResult, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("Expected *objects.Int, got %T", result)
	}
	if intResult.Value != 3 {
		t.Errorf("len([1,2,3]) = %d, expected 3", intResult.Value)
	}
}

func TestBuiltinTypeOf(t *testing.T) {
	builtin := getBuiltinByName("typeOf")
	if builtin == nil {
		t.Fatal("typeOf builtin not found")
	}

	tests := []struct {
		obj      objects.Object
		expected string
	}{
		{&objects.Int{Value: 42}, "INT"},
		{&objects.Float{Value: 3.14}, "FLOAT"},
		{&objects.String{Value: "test"}, "STRING"},
		{&objects.Bool{Value: true}, "BOOL"},
		{&objects.Array{Elements: []objects.Object{}}, "ARRAY"},
		{&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{}}, "MAP"},
		{objects.NULL, "NULL"},
	}

	for _, tt := range tests {
		result := builtin.Fn(tt.obj)

		strResult, ok := result.(*objects.String)
		if !ok {
			t.Errorf("typeOf returned %T, expected *objects.String", result)
			continue
		}
		if strResult.Value != tt.expected {
			t.Errorf("typeOf(%v) = %s, expected %s", tt.obj, strResult.Value, tt.expected)
		}
	}
}

func TestBuiltinAbs(t *testing.T) {
	builtin := getBuiltinByName("abs")
	if builtin == nil {
		t.Fatal("abs builtin not found")
	}

	tests := []struct {
		input    int64
		expected int64
	}{
		{42, 42},
		{-42, 42},
		{0, 0},
		{-1, 1},
	}

	for _, tt := range tests {
		result := builtin.Fn(&objects.Int{Value: tt.input})

		intResult, ok := result.(*objects.Int)
		if !ok {
			t.Errorf("abs returned %T", result)
			continue
		}
		if intResult.Value != tt.expected {
			t.Errorf("abs(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

func TestBuiltinUpperLower(t *testing.T) {
	upperBuiltin := getBuiltinByName("upper")
	if upperBuiltin == nil {
		t.Fatal("upper builtin not found")
	}

	result := upperBuiltin.Fn(&objects.String{Value: "hello"})
	strResult, ok := result.(*objects.String)
	if !ok || strResult.Value != "HELLO" {
		t.Errorf("upper('hello') = %v, expected 'HELLO'", result)
	}

	lowerBuiltin := getBuiltinByName("lower")
	if lowerBuiltin == nil {
		t.Fatal("lower builtin not found")
	}

	result = lowerBuiltin.Fn(&objects.String{Value: "HELLO"})
	strResult, ok = result.(*objects.String)
	if !ok || strResult.Value != "hello" {
		t.Errorf("lower('HELLO') = %v, expected 'hello'", result)
	}
}

func TestBuiltinPush(t *testing.T) {
	builtin := getBuiltinByName("push")
	if builtin == nil {
		t.Fatal("push builtin not found")
	}

	arr := &objects.Array{Elements: []objects.Object{&objects.Int{Value: 1}}}
	result := builtin.Fn(arr, &objects.Int{Value: 2})

	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("push returned %T", result)
	}
	if len(arrResult.Elements) != 2 {
		t.Errorf("push result length = %d, expected 2", len(arrResult.Elements))
	}
}

func TestBuiltinSort(t *testing.T) {
	builtin := getBuiltinByName("sort")
	if builtin == nil {
		t.Fatal("sort builtin not found")
	}

	arr := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: 3},
		&objects.Int{Value: 1},
		&objects.Int{Value: 2},
	}}
	result := builtin.Fn(arr)

	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("sort returned %T", result)
	}

	expected := []int64{1, 2, 3}
	for i, elem := range arrResult.Elements {
		intElem, ok := elem.(*objects.Int)
		if !ok || intElem.Value != expected[i] {
			t.Errorf("sort result[%d] = %v, expected %d", i, elem, expected[i])
		}
	}
}

// ============================================
// Additional Builtin Tests for Coverage
// ============================================

func TestAdditionalBuiltinSubstr(t *testing.T) {
	builtin := getBuiltinByName("substr")
	if builtin == nil {
		t.Skip("substr builtin not available")
	}

	str := &objects.String{Value: "hello world"}
	result := builtin.Fn(str, objects.NewInt(0), objects.NewInt(5))
	if strResult, ok := result.(*objects.String); ok {
		if strResult.Value != "hello" {
			t.Errorf("substr('hello world', 0, 5) = %q, expected 'hello'", strResult.Value)
		}
	} else {
		t.Fatalf("substr returned %T, expected *String", result)
	}
}

func TestAdditionalBuiltinSplitJoin(t *testing.T) {
	split := getBuiltinByName("split")
	if split != nil {
		str := &objects.String{Value: "a,b,c"}
		sep := &objects.String{Value: ","}
		result := split.Fn(str, sep)
		if arr, ok := result.(*objects.Array); ok {
			if len(arr.Elements) != 3 {
				t.Errorf("split('a,b,c', ',') returned %d elements, expected 3", len(arr.Elements))
			}
		}
	}

	join := getBuiltinByName("join")
	if join != nil {
		arr := &objects.Array{Elements: []objects.Object{
			&objects.String{Value: "a"},
			&objects.String{Value: "b"},
			&objects.String{Value: "c"},
		}}
		sep := &objects.String{Value: "-"}
		result := join.Fn(arr, sep)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "a-b-c" {
				t.Errorf("join(['a','b','c'], '-') = %q, expected 'a-b-c'", strResult.Value)
			}
		}
	}
}

func TestAdditionalBuiltinTrim(t *testing.T) {
	builtin := getBuiltinByName("trim")
	if builtin == nil {
		t.Skip("trim builtin not available")
	}

	str := &objects.String{Value: "  hello  "}
	result := builtin.Fn(str)
	if strResult, ok := result.(*objects.String); ok {
		if strResult.Value != "hello" {
			t.Errorf("trim('  hello  ') = %q, expected 'hello'", strResult.Value)
		}
	} else {
		t.Fatalf("trim returned %T, expected *String", result)
	}
}

func TestAdditionalBuiltinArrayUtilities(t *testing.T) {
	first := getBuiltinByName("first")
	if first != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := first.Fn(arr)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 1 {
				t.Errorf("first([1,2,3]) = %d, expected 1", intResult.Value)
			}
		}
	}

	last := getBuiltinByName("last")
	if last != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := last.Fn(arr)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("last([1,2,3]) = %d, expected 3", intResult.Value)
			}
		}
	}

	rest := getBuiltinByName("rest")
	if rest != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := rest.Fn(arr)
		if arrResult, ok := result.(*objects.Array); ok {
			if len(arrResult.Elements) != 2 {
				t.Errorf("rest([1,2,3]) length = %d, expected 2", len(arrResult.Elements))
			}
		}
	}

	pop := getBuiltinByName("pop")
	if pop != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := pop.Fn(arr)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("pop([1,2,3]) = %d, expected 3", intResult.Value)
			}
		}
	}

	rangeFn := getBuiltinByName("range")
	if rangeFn != nil {
		result := rangeFn.Fn(objects.NewInt(5))
		if arrResult, ok := result.(*objects.Array); ok {
			if len(arrResult.Elements) != 5 {
				t.Errorf("range(5) length = %d, expected 5", len(arrResult.Elements))
			}
		}
	}

	sum := getBuiltinByName("sum")
	if sum != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := sum.Fn(arr)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 6 {
				t.Errorf("sum([1,2,3]) = %d, expected 6", intResult.Value)
			}
		}
	}

	avg := getBuiltinByName("avg")
	if avg != nil {
		arr := &objects.Array{Elements: []objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		}}
		result := avg.Fn(arr)
		if floatResult, ok := result.(*objects.Float); ok {
			if floatResult.Value != 2.0 {
				t.Errorf("avg([1,2,3]) = %f, expected 2.0", floatResult.Value)
			}
		}
	}
}

func TestAdditionalBuiltinMapUtilities(t *testing.T) {
	keys := getBuiltinByName("keys")
	if keys == nil {
		t.Skip("keys builtin not available")
	}
	keyA := &objects.String{Value: "a"}
	keyB := &objects.String{Value: "b"}
	m := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		keyA.HashKey(): {Key: keyA, Value: objects.NewInt(1)},
		keyB.HashKey(): {Key: keyB, Value: objects.NewInt(2)},
	}}
	result := keys.Fn(m)
	if arr, ok := result.(*objects.Array); ok {
		if len(arr.Elements) != 2 {
			t.Errorf("keys length = %d, expected 2", len(arr.Elements))
		}
	}

	values := getBuiltinByName("values")
	if values != nil {
		result = values.Fn(m)
		if arr, ok := result.(*objects.Array); ok {
			if len(arr.Elements) != 2 {
				t.Errorf("values length = %d, expected 2", len(arr.Elements))
			}
		}
	}

	hasKey := getBuiltinByName("hasKey")
	if hasKey != nil {
		result = hasKey.Fn(m, keyA)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("hasKey('a') = false, expected true")
			}
		}
	}

	deleteFn := getBuiltinByName("delete")
	if deleteFn != nil {
		m2 := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			keyA.HashKey(): {Key: keyA, Value: objects.NewInt(1)},
		}}
		result := deleteFn.Fn(m2, keyA)
		if mResult, ok := result.(*objects.Map); ok {
			if _, exists := mResult.Pairs[keyA.HashKey()]; exists {
				t.Error("delete did not remove key")
			}
		}
	}
}

func TestAdditionalBuiltinTypeChecking(t *testing.T) {
	isEmpty := getBuiltinByName("isEmpty")
	if isEmpty != nil {
		str := &objects.String{Value: ""}
		result := isEmpty.Fn(str)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isEmpty('') = false, expected true")
			}
		}
	}

	isString := getBuiltinByName("isString")
	if isString != nil {
		result := isString.Fn(&objects.String{Value: "test"})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isString('test') = false, expected true")
			}
		}
	}

	isNumber := getBuiltinByName("isNumber")
	if isNumber != nil {
		result := isNumber.Fn(objects.NewInt(42))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isNumber(42) = false, expected true")
			}
		}
	}

	isInt := getBuiltinByName("isInt")
	if isInt != nil {
		result := isInt.Fn(objects.NewInt(42))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isInt(42) = false, expected true")
			}
		}
	}

	isFloat := getBuiltinByName("isFloat")
	if isFloat != nil {
		result := isFloat.Fn(objects.NewFloat(3.14))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isFloat(3.14) = false, expected true")
			}
		}
	}

	isArray := getBuiltinByName("isArray")
	if isArray != nil {
		result := isArray.Fn(&objects.Array{Elements: []objects.Object{}})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isArray([]) = false, expected true")
			}
		}
	}

	isMap := getBuiltinByName("isMap")
	if isMap != nil {
		result := isMap.Fn(&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{}})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isMap({}) = false, expected true")
			}
		}
	}

	isBool := getBuiltinByName("isBool")
	if isBool != nil {
		result := isBool.Fn(objects.TRUE)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isBool(true) = false, expected true")
			}
		}
	}

	// isFunction requires compiler import, skipping for now

	isNull := getBuiltinByName("isNull")
	if isNull != nil {
		result := isNull.Fn(objects.NULL)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isNull(NULL) = false, expected true")
			}
		}
	}
}

func TestAdditionalBuiltinEncoding(t *testing.T) {
	base64Encode := getBuiltinByName("base64Encode")
	if base64Encode != nil {
		data := &objects.String{Value: "hello"}
		result := base64Encode.Fn(data)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "aGVsbG8=" {
				t.Errorf("base64Encode('hello') = %q, expected 'aGVsbG8='", strResult.Value)
			}
		}
	}

	base64Decode := getBuiltinByName("base64Decode")
	if base64Decode != nil {
		encoded := &objects.String{Value: "aGVsbG8="}
		result := base64Decode.Fn(encoded)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "hello" {
				t.Errorf("base64Decode('aGVsbG8=') = %q, expected 'hello'", strResult.Value)
			}
		}
	}

	hexEncode := getBuiltinByName("hexEncode")
	if hexEncode != nil {
		data := &objects.String{Value: "hello"}
		result := hexEncode.Fn(data)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "68656c6c6f" {
				t.Errorf("hexEncode('hello') = %q, expected '68656c6c6f'", strResult.Value)
			}
		}
	}

	hexDecode := getBuiltinByName("hexDecode")
	if hexDecode != nil {
		hex := &objects.String{Value: "68656c6c6f"}
		result := hexDecode.Fn(hex)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "hello" {
				t.Errorf("hexDecode('68656c6c6f') = %q, expected 'hello'", strResult.Value)
			}
		}
	}
}

func TestAdditionalBuiltinMath(t *testing.T) {
	floor := getBuiltinByName("floor")
	if floor != nil {
		result := floor.Fn(objects.NewFloat(3.7))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("floor(3.7) = %d, expected 3", intResult.Value)
			}
		}
	}

	ceil := getBuiltinByName("ceil")
	if ceil != nil {
		result := ceil.Fn(objects.NewFloat(3.2))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 4 {
				t.Errorf("ceil(3.2) = %d, expected 4", intResult.Value)
			}
		}
	}

	sqrt := getBuiltinByName("sqrt")
	if sqrt != nil {
		result := sqrt.Fn(objects.NewInt(16))
		if floatResult, ok := result.(*objects.Float); ok {
			if floatResult.Value != 4.0 {
				t.Errorf("sqrt(16) = %f, expected 4.0", floatResult.Value)
			}
		}
	}

	pow := getBuiltinByName("pow")
	if pow != nil {
		base := objects.NewInt(2)
		exp := objects.NewInt(3)
		result := pow.Fn(base, exp)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 8 {
				t.Errorf("pow(2, 3) = %d, expected 8", intResult.Value)
			}
		}
	}

	min := getBuiltinByName("min")
	if min != nil {
		result := min.Fn(objects.NewInt(3), objects.NewInt(5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("min(3, 5) = %d, expected 3", intResult.Value)
			}
		}
	}

	max := getBuiltinByName("max")
	if max != nil {
		result := max.Fn(objects.NewInt(3), objects.NewInt(5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 5 {
				t.Errorf("max(3, 5) = %d, expected 5", intResult.Value)
			}
		}
	}

	clamp := getBuiltinByName("clamp")
	if clamp != nil {
		val := objects.NewInt(15)
		minVal := objects.NewInt(0)
		maxVal := objects.NewInt(10)
		result := clamp.Fn(val, minVal, maxVal)
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 10 {
				t.Errorf("clamp(15, 0, 10) = %d, expected 10", intResult.Value)
			}
		}
	}

	sign := getBuiltinByName("sign")
	if sign != nil {
		result := sign.Fn(objects.NewInt(-5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != -1 {
				t.Errorf("sign(-5) = %d, expected -1", intResult.Value)
			}
		}
		result = sign.Fn(objects.NewInt(0))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 0 {
				t.Errorf("sign(0) = %d, expected 0", intResult.Value)
			}
		}
		result = sign.Fn(objects.NewInt(5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 1 {
				t.Errorf("sign(5) = %d, expected 1", intResult.Value)
			}
		}
	}
}

func TestAdditionalBuiltinTypeConversion(t *testing.T) {
	intFn := getBuiltinByName("int")
	if intFn != nil {
		result := intFn.Fn(objects.NewFloat(3.7))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("int(3.7) = %d, expected 3", intResult.Value)
			}
		}
	}

	floatFn := getBuiltinByName("float")
	if floatFn != nil {
		result := floatFn.Fn(objects.NewInt(42))
		if floatResult, ok := result.(*objects.Float); ok {
			if floatResult.Value != 42.0 {
				t.Errorf("float(42) = %f, expected 42.0", floatResult.Value)
			}
		}
	}

	stringFn := getBuiltinByName("string")
	if stringFn != nil {
		result := stringFn.Fn(objects.NewInt(42))
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "42" {
				t.Errorf("string(42) = %q, expected '42'", strResult.Value)
			}
		}
	}
}

func TestAdditionalBuiltinSprintf(t *testing.T) {
	sprintf := getBuiltinByName("sprintf")
	if sprintf == nil {
		t.Skip("sprintf builtin not available")
	}
	format := &objects.String{Value: "Hello %s, you have %d messages"}
	name := &objects.String{Value: "World"}
	count := objects.NewInt(5)
	result := sprintf.Fn(format, name, count)
	if strResult, ok := result.(*objects.String); ok {
		expected := "Hello World, you have 5 messages"
		if strResult.Value != expected {
			t.Errorf("sprintf got %q, expected %q", strResult.Value, expected)
		}
	}
}

func TestAdditionalBuiltinCoalesce(t *testing.T) {
	coalesce := getBuiltinByName("coalesce")
	if coalesce == nil {
		t.Skip("coalesce builtin not available")
	}
	result := coalesce.Fn(objects.NULL, objects.NewInt(42))
	if intResult, ok := result.(*objects.Int); ok {
		if intResult.Value != 42 {
			t.Errorf("coalesce(NULL, 42) = %d, expected 42", intResult.Value)
		}
	} else {
		t.Fatalf("coalesce returned %T: %v, expected *Int", result, result)
	}

	result = coalesce.Fn(objects.NewInt(10), objects.NewInt(20))
	if intResult, ok := result.(*objects.Int); ok {
		if intResult.Value != 10 {
			t.Errorf("coalesce(10, 20) = %d, expected 10", intResult.Value)
		}
	} else {
		t.Fatalf("coalesce returned %T, expected *Int", result)
	}
}

func TestAdditionalBuiltinJson(t *testing.T) {
	toJson := getBuiltinByName("toJson")
	if toJson == nil {
		t.Skip("toJson builtin not available")
	}
	arr := &objects.Array{Elements: []objects.Object{
		objects.NewInt(1),
		objects.NewString("two"),
	}}
	result := toJson.Fn(arr)
	if strResult, ok := result.(*objects.String); ok {
		if strResult.Value == "" {
			t.Error("toJson produced empty string")
		}
	}

	fromJson := getBuiltinByName("fromJson")
	if fromJson != nil {
		jsonStr := &objects.String{Value: `[1,2,3]`}
		result := fromJson.Fn(jsonStr)
		if arrResult, ok := result.(*objects.Array); ok {
			if len(arrResult.Elements) != 3 {
				t.Errorf("fromJson array length = %d, expected 3", len(arrResult.Elements))
			}
		}
	}

	jsonValid := getBuiltinByName("jsonValid")
	if jsonValid != nil {
		validJson := &objects.String{Value: `{"a":1}`}
		result := jsonValid.Fn(validJson)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("jsonValid(valid) = false, expected true")
			}
		}
	}
}
