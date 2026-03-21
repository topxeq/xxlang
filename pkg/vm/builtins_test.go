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
		index    int
		name     string
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
