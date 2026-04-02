// pkg/objects/builtin_json_test.go
package objects

import (
	"testing"
)

func TestBuiltinFormatJson(t *testing.T) {
	fn, ok := Builtins["formatJson"]
	if !ok {
		t.Fatal("formatJson builtin not found")
	}

	result := fn.Fn(NewString(`{"name":"test"}`))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(NewString(`{"name":"test"}`), NewString("\t"))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(NewString(`invalid json`))
	if !isError(result) {
		t.Error("expected error for invalid JSON")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first argument")
	}

	result = fn.Fn(NewString(`{}`), NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string second argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinCompactJson(t *testing.T) {
	fn, ok := Builtins["compactJson"]
	if !ok {
		t.Fatal("compactJson builtin not found")
	}

	result := fn.Fn(NewString(`{  "name":  "test"  }`))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != `{"name":"test"}` {
		t.Errorf("expected compact JSON, got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`invalid json`))
	if !isError(result) {
		t.Error("expected error for invalid JSON")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetJsonNodeStr(t *testing.T) {
	fn, ok := Builtins["getJsonNodeStr"]
	if !ok {
		t.Fatal("getJsonNodeStr builtin not found")
	}

	result := fn.Fn(NewString(`{"name":"John","age":30}`), NewString("name"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "John" {
		t.Errorf("expected 'John', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"user":{"name":"John"}}`), NewString("user.name"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "John" {
		t.Errorf("expected 'John', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"items":["a","b"]}`), NewString("items.0"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "a" {
		t.Errorf("expected 'a', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"items":["a","b"]}`), NewString("items.#"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String (for array length), got %T", result)
	}
	if strResult.Value != "2" {
		t.Errorf("expected '2', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"value":123}`), NewString("value"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 123 {
		t.Errorf("expected 123, got %d", intResult.Value)
	}

	result = fn.Fn(NewString(`{"flag":true}`), NewString("flag"))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true, got %v", boolResult.Value)
	}

	result = fn.Fn(NewString(`{"data":null}`), NewString("data"))
	if result != NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	result = fn.Fn(NewString(`{"items":[{"name":"a"},{"name":"b"}]}`), NewString("items.0.name"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "a" {
		t.Errorf("expected 'a', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`invalid`))
	if !isError(result) {
		t.Error("expected error for invalid JSON")
	}

	result = fn.Fn(NewString(`{"name":"test"}`), NewString("nonexistent"))
	if result != NULL {
		t.Errorf("expected NULL for nonexistent path, got %T", result)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinGetJsonNodeStrs(t *testing.T) {
	fn, ok := Builtins["getJsonNodeStrs"]
	if !ok {
		t.Fatal("getJsonNodeStrs builtin not found")
	}

	result := fn.Fn(NewString(`{"items":["a","b","c"]}`), NewString("items"))
	arrResult, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arrResult.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrResult.Elements))
	}

	result = fn.Fn(NewString(`{"items":"not an array"}`), NewString("items"))
	if result != NULL {
		t.Errorf("expected NULL for non-array, got %T", result)
	}

	result = fn.Fn(NewString(`invalid`))
	if !isError(result) {
		t.Error("expected error for invalid JSON")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrsToJson(t *testing.T) {
	fn, ok := Builtins["strsToJson"]
	if !ok {
		t.Fatal("strsToJson builtin not found")
	}

	result := fn.Fn(NewArray([]Object{NewString("name"), NewString("value")}),
		NewArray([]Object{NewString("test"), NewInt(123)}))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(NewArray([]Object{NewString("key")}),
		NewArray([]Object{NewString("value")}))
	if _, ok := result.(*String); !ok {
		t.Fatalf("expected String, got %T", result)
	}

	result = fn.Fn(NewString("not an array"), NewArray([]Object{NewString("value")}))
	if !isError(result) {
		t.Error("expected error for non-array first argument")
	}

	result = fn.Fn(NewArray([]Object{NewString("key")}), NewString("not an array"))
	if !isError(result) {
		t.Error("expected error for non-array second argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinJsonValid(t *testing.T) {
	fn, ok := Builtins["jsonValid"]
	if !ok {
		t.Fatal("jsonValid builtin not found")
	}

	result := fn.Fn(NewString(`{"valid":true}`))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("expected true for valid JSON")
	}

	result = fn.Fn(NewString(`invalid`))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("expected false for invalid JSON")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinJsonType(t *testing.T) {
	fn, ok := Builtins["jsonType"]
	if !ok {
		t.Fatal("jsonType builtin not found")
	}

	result := fn.Fn(NewString(`"string value"`))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "string" {
		t.Errorf("expected 'string', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`123`))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "number" {
		t.Errorf("expected 'number', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`true`))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "boolean" {
		t.Errorf("expected 'boolean', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`[1,2,3]`))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "array" {
		t.Errorf("expected 'array', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"key":"value"}`))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "object" {
		t.Errorf("expected 'object', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`null`))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "null" {
		t.Errorf("expected 'null', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"user":{"name":"John"}}`), NewString("user.name"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "string" {
		t.Errorf("expected 'string', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`{"data":null}`), NewString("data"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "null" {
		t.Errorf("expected 'null', got '%s'", strResult.Value)
	}

	result = fn.Fn(NewString(`invalid`))
	if !isError(result) {
		t.Error("expected error for invalid JSON")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinJsonPath(t *testing.T) {
	fn, ok := Builtins["jsonPath"]
	if !ok {
		t.Fatal("jsonPath builtin not found")
	}

	result := fn.Fn(NewString(`{"name":"John"}`), NewString("name"))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value != "John" {
		t.Errorf("expected 'John', got '%s'", strResult.Value)
	}
}

func TestGoArrayToObject(t *testing.T) {
	arr := []interface{}{"hello", 123.0, true, nil}
	result := goArrayToObject(arr)

	if result == nil {
		t.Fatal("expected non-nil Array")
	}

	if len(result.Elements) != 4 {
		t.Errorf("expected 4 elements, got %d", len(result.Elements))
	}

	if _, ok := result.Elements[0].(*String); !ok {
		t.Error("expected first element to be String")
	}

	if _, ok := result.Elements[1].(*Int); !ok {
		t.Error("expected second element to be Int")
	}

	if _, ok := result.Elements[2].(*Bool); !ok {
		t.Error("expected third element to be Bool")
	}

	if result.Elements[3] != NULL {
		t.Error("expected fourth element to be NULL")
	}
}

func TestGoMapToObject(t *testing.T) {
	m := map[string]interface{}{
		"name":   "John",
		"age":    30.0,
		"active": true,
	}

	result := goMapToObject(m)

	if result == nil {
		t.Fatal("expected non-nil Map")
	}

	if len(result.Pairs) != 3 {
		t.Errorf("expected 3 pairs, got %d", len(result.Pairs))
	}
}

func TestGoValueToObject(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "NULL"},
		{true, "true"},
		{false, "false"},
		{123.0, "123"},
		{123.5, "123.5"},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		result := goValueToObject(tt.input)
		if result.Inspect() != tt.expected && !(tt.input == nil && result == NULL) {
			t.Errorf("goValueToObject(%v) = %s, want %s", tt.input, result.Inspect(), tt.expected)
		}
	}

	arr := []interface{}{1.0, 2.0}
	result := goValueToObject(arr)
	if _, ok := result.(*Array); !ok {
		t.Errorf("expected Array for slice input, got %T", result)
	}

	m := map[string]interface{}{"key": "value"}
	result = goValueToObject(m)
	if _, ok := result.(*Map); !ok {
		t.Errorf("expected Map for map input, got %T", result)
	}
}
