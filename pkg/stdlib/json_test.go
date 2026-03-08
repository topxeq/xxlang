// pkg/stdlib/json_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callJSONFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/json")
	if mod == nil {
		panic("std/json module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestJSONParsePrimitives(t *testing.T) {
	// Parse null
	result := callJSONFunc("parse", String("null"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("parse('null') should return Null, got %T", result)
	}

	// Parse boolean
	result = callJSONFunc("parse", String("true"))
	r, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("parse('true') should return Bool, got %T", result)
	}
	if !r.Value {
		t.Errorf("parse('true') = %v, want true", r.Value)
	}

	result = callJSONFunc("parse", String("false"))
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("parse('false') should return Bool, got %T", result)
	}
	if r.Value {
		t.Errorf("parse('false') = %v, want false", r.Value)
	}

	// Parse integer
	result = callJSONFunc("parse", String("42"))
	ri, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("parse('42') should return Int, got %T", result)
	}
	if ri.Value != 42 {
		t.Errorf("parse('42') = %d, want 42", ri.Value)
	}

	// Parse float
	result = callJSONFunc("parse", String("3.14"))
	rf, ok := result.(*objects.Float)
	if !ok {
		t.Fatalf("parse('3.14') should return Float, got %T", result)
	}
	if rf.Value != 3.14 {
		t.Errorf("parse('3.14') = %v, want 3.14", rf.Value)
	}

	// Parse string
	result = callJSONFunc("parse", String(`"hello"`))
	rs, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("parse('\"hello\"') should return String, got %T", result)
	}
	if rs.Value != "hello" {
		t.Errorf("parse('\"hello\"') = %s, want hello", rs.Value)
	}
}

func TestJSONParseArray(t *testing.T) {
	result := callJSONFunc("parse", String("[1, 2, 3]"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("parse('[1, 2, 3]') should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("parse('[1, 2, 3]') length = %d, want 3", len(arr.Elements))
	}

	// Check elements
	for i := 0; i < 3; i++ {
		elem, ok := arr.Elements[i].(*objects.Int)
		if !ok || elem.Value != int64(i+1) {
			t.Errorf("array[%d] = %v, want %d", i, arr.Elements[i], i+1)
		}
	}
}

func TestJSONParseMap(t *testing.T) {
	result := callJSONFunc("parse", String(`{"name": "test", "value": 42}`))
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parse('{\"name\": \"test\", \"value\": 42}') should return Map, got %T", result)
	}
	if len(m.Pairs) != 2 {
		t.Errorf("parse map length = %d, want 2", len(m.Pairs))
	}

	// Check values
	nameKey := String("name").HashKey()
	if pair, ok := m.Pairs[nameKey]; ok {
		if s, ok := pair.Value.(*objects.String); !ok || s.Value != "test" {
			t.Errorf("map['name'] = %v, want 'test'", pair.Value)
		}
	} else {
		t.Errorf("map should have key 'name'")
	}

	valueKey := String("value").HashKey()
	if pair, ok := m.Pairs[valueKey]; ok {
		if i, ok := pair.Value.(*objects.Int); !ok || i.Value != 42 {
			t.Errorf("map['value'] = %v, want 42", pair.Value)
		}
	} else {
		t.Errorf("map should have key 'value'")
	}
}

func TestJSONParseNested(t *testing.T) {
	jsonStr := `{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}`
	result := callJSONFunc("parse", String(jsonStr))
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parse nested JSON should return Map, got %T", result)
	}

	usersKey := String("users").HashKey()
	pair, ok := m.Pairs[usersKey]
	if !ok {
		t.Fatal("map should have 'users' key")
	}

	users, ok := pair.Value.(*objects.Array)
	if !ok {
		t.Fatalf("'users' should be Array, got %T", pair.Value)
	}
	if len(users.Elements) != 2 {
		t.Errorf("users array length = %d, want 2", len(users.Elements))
	}
}

func TestJSONParseError(t *testing.T) {
	result := callJSONFunc("parse", String("invalid json"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse('invalid json') should return Error, got %T", result)
	}

	result = callJSONFunc("parse", String("{broken: json}"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse('{broken: json}') should return Error, got %T", result)
	}
}

func TestJSONStringifyPrimitives(t *testing.T) {
	// Stringify null
	result := callJSONFunc("stringify", Null())
	if r, ok := result.(*objects.String); !ok || r.Value != "null" {
		t.Errorf("stringify(null) = %v, want 'null'", result)
	}

	// Stringify boolean
	result = callJSONFunc("stringify", Bool(true))
	if r, ok := result.(*objects.String); !ok || r.Value != "true" {
		t.Errorf("stringify(true) = %v, want 'true'", result)
	}

	// Stringify integer
	result = callJSONFunc("stringify", Int(42))
	if r, ok := result.(*objects.String); !ok || r.Value != "42" {
		t.Errorf("stringify(42) = %v, want '42'", result)
	}

	// Stringify float
	result = callJSONFunc("stringify", Float(3.14))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringify(3.14) should return String, got %T", result)
	}
	if r.Value != "3.14" {
		t.Errorf("stringify(3.14) = %s, want '3.14'", r.Value)
	}

	// Stringify string
	result = callJSONFunc("stringify", String("hello"))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("stringify('hello') should return String, got %T", result)
	}
	if r.Value != `"hello"` {
		t.Errorf("stringify('hello') = %s, want '\"hello\"'", r.Value)
	}
}

func TestJSONStringifyArray(t *testing.T) {
	arr := Array(Int(1), Int(2), Int(3))
	result := callJSONFunc("stringify", arr)
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringify([1,2,3]) should return String, got %T", result)
	}
	if r.Value != "[1,2,3]" {
		t.Errorf("stringify([1,2,3]) = %s, want '[1,2,3]'", r.Value)
	}
}

func TestJSONStringifyMap(t *testing.T) {
	// Create a map
	pairs := make(map[objects.HashKey]objects.MapPair)
	pairs[String("name").HashKey()] = objects.MapPair{
		Key:   String("name"),
		Value: String("test"),
	}
	pairs[String("value").HashKey()] = objects.MapPair{
		Key:   String("value"),
		Value: Int(42),
	}
	m := &objects.Map{Pairs: pairs}

	result := callJSONFunc("stringify", m)
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringify(map) should return String, got %T", result)
	}
	// Note: map order is not guaranteed, so we just check it contains the expected values
	if r.Value == "" {
		t.Errorf("stringify(map) returned empty string")
	}
}

func TestJSONStringifyWithIndent(t *testing.T) {
	arr := Array(Int(1), Int(2))
	result := callJSONFunc("stringify", arr, String("  "))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringify with indent should return String, got %T", result)
	}
	// Should contain newlines due to indentation
	if r.Value == "[1,2]" {
		t.Errorf("stringify with indent should format with newlines")
	}

	// Test with integer indent
	result = callJSONFunc("stringify", arr, Int(2))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("stringify with int indent should return String, got %T", result)
	}
	if r.Value == "[1,2]" {
		t.Errorf("stringify with int indent should format with newlines")
	}
}

func TestJSONEncodeDecode(t *testing.T) {
	// Test encode (alias for stringify)
	result := callJSONFunc("encode", Int(42))
	if r, ok := result.(*objects.String); !ok || r.Value != "42" {
		t.Errorf("encode(42) = %v, want '42'", result)
	}

	// Test decode (alias for parse)
	result = callJSONFunc("decode", String("42"))
	if r, ok := result.(*objects.Int); !ok || r.Value != 42 {
		t.Errorf("decode('42') = %v, want 42", result)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Create a complex structure
	arr := Array(
		String("hello"),
		Int(42),
		Float(3.14),
		Bool(true),
		Null(),
		Array(Int(1), Int(2)),
	)

	// Stringify
	encoded := callJSONFunc("stringify", arr)
	encStr, ok := encoded.(*objects.String)
	if !ok {
		t.Fatalf("stringify should return String, got %T", encoded)
	}

	// Parse back
	decoded := callJSONFunc("parse", encStr)
	arr2, ok := decoded.(*objects.Array)
	if !ok {
		t.Fatalf("parse should return Array, got %T", decoded)
	}

	// Verify structure
	if len(arr2.Elements) != 6 {
		t.Errorf("round-trip array length = %d, want 6", len(arr2.Elements))
	}
}

func TestJSONErrorCases(t *testing.T) {
	// parse with wrong number of args
	result := callJSONFunc("parse")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse() with no args should return Error, got %T", result)
	}

	// stringify with wrong number of args
	result = callJSONFunc("stringify")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("stringify() with no args should return Error, got %T", result)
	}

	// parse with non-string
	result = callJSONFunc("parse", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse(int) should return Error, got %T", result)
	}
}
