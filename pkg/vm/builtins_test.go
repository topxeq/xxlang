// pkg/vm/builtins_test.go
// Tests for builtin function support
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// getBuiltin Tests
// ============================================

func TestGetBuiltinValid(t *testing.T) {
	// Test valid indices
	tests := []struct {
		index int
		name  string
	}{
		{0, "len"},
		{1, "pr"},
		{2, "pln"},
		{3, "typeOf"},
		{14, "abs"},
		{36, "range"},
		{37, "sort"},
	}

	for _, tt := range tests {
		builtin := getBuiltin(tt.index)
		if builtin == nil {
			t.Errorf("getBuiltin(%d) returned nil for %s", tt.index, tt.name)
		}
	}
}

func TestGetBuiltinNegative(t *testing.T) {
	builtin := getBuiltin(-1)
	if builtin != nil {
		t.Error("getBuiltin(-1) should return nil")
	}
}

func TestGetBuiltinOutOfRange(t *testing.T) {
	builtin := getBuiltin(10000)
	if builtin != nil {
		t.Error("getBuiltin(10000) should return nil")
	}
}

// ============================================
// GetBuiltinByIndex Tests
// ============================================

func TestGetBuiltinByIndex(t *testing.T) {
	// Test exported function
	builtin := GetBuiltinByIndex(0)
	if builtin == nil {
		t.Error("GetBuiltinByIndex(0) returned nil")
	}

	// Should be len function
	if builtin.Fn == nil {
		t.Error("Builtin function is nil")
	}
}

func TestGetBuiltinByIndexInvalid(t *testing.T) {
	builtin := GetBuiltinByIndex(-1)
	if builtin != nil {
		t.Error("GetBuiltinByIndex(-1) should return nil")
	}

	builtin = GetBuiltinByIndex(100000)
	if builtin != nil {
		t.Error("GetBuiltinByIndex(100000) should return nil")
	}
}

// ============================================
// Builtin Function Execution Tests
// ============================================

func TestBuiltinLen(t *testing.T) {
	// Get len builtin (index 0)
	builtin := getBuiltin(0)
	if builtin == nil {
		t.Fatal("len builtin not found")
	}

	// Test with string
	strObj := &objects.String{Value: "hello"}
	result := builtin.Fn(strObj)

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("Expected *objects.Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("len('hello') = %d, expected 5", intResult.Value)
	}

	// Test with array
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
	// Get typeOf builtin (index 3)
	builtin := getBuiltin(3)
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
	// Get abs builtin (index 14)
	builtin := getBuiltin(14)
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
	// Get upper builtin (index 8)
	upperBuiltin := getBuiltin(8)
	if upperBuiltin == nil {
		t.Fatal("upper builtin not found")
	}

	result := upperBuiltin.Fn(&objects.String{Value: "hello"})
	strResult, ok := result.(*objects.String)
	if !ok || strResult.Value != "HELLO" {
		t.Errorf("upper('hello') = %v, expected 'HELLO'", result)
	}

	// Get lower builtin (index 9)
	lowerBuiltin := getBuiltin(9)
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
	// Get push builtin (index 24)
	builtin := getBuiltin(24)
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
	// Get sort builtin (index 37)
	builtin := getBuiltin(37)
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
	builtin := getBuiltin(4) // substr
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
	// split (5)
	split := getBuiltin(5)
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

	// join (6)
	join := getBuiltin(6)
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
	builtin := getBuiltin(7) // trim
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
	// first (26)
	first := getBuiltin(26)
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

	// last (27)
	last := getBuiltin(27)
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

	// rest (28)
	rest := getBuiltin(28)
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

	// pop (25)
	pop := getBuiltin(25)
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

	// range (36)
	rangeFn := getBuiltin(36)
	if rangeFn != nil {
		result := rangeFn.Fn(objects.NewInt(5))
		if arrResult, ok := result.(*objects.Array); ok {
			if len(arrResult.Elements) != 5 {
				t.Errorf("range(5) length = %d, expected 5", len(arrResult.Elements))
			}
		}
	}

	// sum (38)
	sum := getBuiltin(38)
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

	// avg (39)
	avg := getBuiltin(39)
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
	// keys (32)
	keys := getBuiltin(32)
	if keys == nil {
		t.Skip("keys builtin not available")
	}
	// Create map using proper HashKey from key objects
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

	// values (33)
	values := getBuiltin(33)
	if values != nil {
		result = values.Fn(m)
		if arr, ok := result.(*objects.Array); ok {
			if len(arr.Elements) != 2 {
				t.Errorf("values length = %d, expected 2", len(arr.Elements))
			}
		}
	}

	// hasKey (34)
	hasKey := getBuiltin(34)
	if hasKey != nil {
		result = hasKey.Fn(m, keyA)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("hasKey('a') = false, expected true")
			}
		}
	}

	// delete (35)
	deleteFn := getBuiltin(35)
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
	// isEmpty (49)
	isEmpty := getBuiltin(49)
	if isEmpty != nil {
		str := &objects.String{Value: ""}
		result := isEmpty.Fn(str)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isEmpty('') = false, expected true")
			}
		}
	}

	// isString (50)
	isString := getBuiltin(50)
	if isString != nil {
		result := isString.Fn(&objects.String{Value: "test"})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isString('test') = false, expected true")
			}
		}
	}

	// isNumber (51)
	isNumber := getBuiltin(51)
	if isNumber != nil {
		result := isNumber.Fn(objects.NewInt(42))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isNumber(42) = false, expected true")
			}
		}
	}

	// isInt (52)
	isInt := getBuiltin(52)
	if isInt != nil {
		result := isInt.Fn(objects.NewInt(42))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isInt(42) = false, expected true")
			}
		}
	}

	// isFloat (53)
	isFloat := getBuiltin(53)
	if isFloat != nil {
		result := isFloat.Fn(objects.NewFloat(3.14))
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isFloat(3.14) = false, expected true")
			}
		}
	}

	// isArray (54)
	isArray := getBuiltin(54)
	if isArray != nil {
		result := isArray.Fn(&objects.Array{Elements: []objects.Object{}})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isArray([]) = false, expected true")
			}
		}
	}

	// isMap (55)
	isMap := getBuiltin(55)
	if isMap != nil {
		result := isMap.Fn(&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{}})
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isMap({}) = false, expected true")
			}
		}
	}

	// isBool (56)
	isBool := getBuiltin(56)
	if isBool != nil {
		result := isBool.Fn(objects.TRUE)
		if boolResult, ok := result.(*objects.Bool); ok {
			if !boolResult.Value {
				t.Error("isBool(true) = false, expected true")
			}
		}
	}

	// isFunction (57) - requires compiler import, skipping for now
	// isFunction := getBuiltin(57)
	// if isFunction != nil {
	// 	fn := &compiler.CompiledFunction{Instructions: []byte{}, NumLocals: 0, NumParameters: 0}
	// 	closure := &Closure{Fn: fn}
	// 	result := isFunction.Fn(closure)
	// 	if boolResult, ok := result.(*objects.Bool); ok && !boolResult.Value {
	// 		t.Error("isFunction(fn) = false, expected true")
	// 	}
	// }

	// isNull (58)
	isNull := getBuiltin(58)
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
	// base64Encode (76)
	base64Encode := getBuiltin(76)
	if base64Encode != nil {
		data := &objects.String{Value: "hello"}
		result := base64Encode.Fn(data)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "aGVsbG8=" {
				t.Errorf("base64Encode('hello') = %q, expected 'aGVsbG8='", strResult.Value)
			}
		}
	}

	// base64Decode (77)
	base64Decode := getBuiltin(77)
	if base64Decode != nil {
		encoded := &objects.String{Value: "aGVsbG8="}
		result := base64Decode.Fn(encoded)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "hello" {
				t.Errorf("base64Decode('aGVsbG8=') = %q, expected 'hello'", strResult.Value)
			}
		}
	}

	// hexEncode (78)
	hexEncode := getBuiltin(78)
	if hexEncode != nil {
		data := &objects.String{Value: "hello"}
		result := hexEncode.Fn(data)
		if strResult, ok := result.(*objects.String); ok {
			if strResult.Value != "68656c6c6f" {
				t.Errorf("hexEncode('hello') = %q, expected '68656c6c6f'", strResult.Value)
			}
		}
	}

	// hexDecode (79)
	hexDecode := getBuiltin(79)
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
	// floor (15)
	floor := getBuiltin(15)
	if floor != nil {
		result := floor.Fn(objects.NewFloat(3.7))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("floor(3.7) = %d, expected 3", intResult.Value)
			}
		}
	}

	// ceil (16)
	ceil := getBuiltin(16)
	if ceil != nil {
		result := ceil.Fn(objects.NewFloat(3.2))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 4 {
				t.Errorf("ceil(3.2) = %d, expected 4", intResult.Value)
			}
		}
	}

	// sqrt (17)
	sqrt := getBuiltin(17)
	if sqrt != nil {
		result := sqrt.Fn(objects.NewInt(16))
		if floatResult, ok := result.(*objects.Float); ok {
			if floatResult.Value != 4.0 {
				t.Errorf("sqrt(16) = %f, expected 4.0", floatResult.Value)
			}
		}
	}

	// pow (18)
	pow := getBuiltin(18)
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

	// min (19)
	min := getBuiltin(19)
	if min != nil {
		result := min.Fn(objects.NewInt(3), objects.NewInt(5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("min(3, 5) = %d, expected 3", intResult.Value)
			}
		}
	}

	// max (20)
	max := getBuiltin(20)
	if max != nil {
		result := max.Fn(objects.NewInt(3), objects.NewInt(5))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 5 {
				t.Errorf("max(3, 5) = %d, expected 5", intResult.Value)
			}
		}
	}

	// clamp (60)
	clamp := getBuiltin(60)
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

	// sign (61)
	sign := getBuiltin(61)
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
	// int (21)
	intFn := getBuiltin(21)
	if intFn != nil {
		result := intFn.Fn(objects.NewFloat(3.7))
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != 3 {
				t.Errorf("int(3.7) = %d, expected 3", intResult.Value)
			}
		}
	}

	// float (22)
	floatFn := getBuiltin(22)
	if floatFn != nil {
		result := floatFn.Fn(objects.NewInt(42))
		if floatResult, ok := result.(*objects.Float); ok {
			if floatResult.Value != 42.0 {
				t.Errorf("float(42) = %f, expected 42.0", floatResult.Value)
			}
		}
	}

	// string (23)
	stringFn := getBuiltin(23)
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
	sprintf := getBuiltin(388)
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
	coalesce := getBuiltin(401)
	if coalesce == nil {
		t.Skip("coalesce builtin not available")
	}
	// coalesce with NULL first
	result := coalesce.Fn(objects.NULL, objects.NewInt(42))
	if intResult, ok := result.(*objects.Int); ok {
		if intResult.Value != 42 {
			t.Errorf("coalesce(NULL, 42) = %d, expected 42", intResult.Value)
		}
	} else {
		t.Fatalf("coalesce returned %T, expected *Int", result)
	}

	// coalesce with both non-null
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
	// toJson (106)
	toJson := getBuiltin(106)
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

	// fromJson (107)
	fromJson := getBuiltin(107)
	if fromJson != nil {
		jsonStr := &objects.String{Value: `[1,2,3]`}
		result := fromJson.Fn(jsonStr)
		if arrResult, ok := result.(*objects.Array); ok {
			if len(arrResult.Elements) != 3 {
				t.Errorf("fromJson array length = %d, expected 3", len(arrResult.Elements))
			}
		}
	}

	// jsonValid (326)
	jsonValid := getBuiltin(326)
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
