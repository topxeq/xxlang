// pkg/stdlib/csv_extra2_test.go
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// csvCall invokes a builtin from the csv module.
func csvCall(name string, args ...objects.Object) objects.Object {
	mod := Get("csv")
	if mod == nil {
		return &objects.Error{Message: "csv module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestCSVExtra2_Parse tests csv.parse.
func TestCSVExtra2_Parse(t *testing.T) {
	// No args
	if res := csvCall("parse"); res.Type() != objects.ErrorType {
		t.Fatalf("parse() with no args should error")
	}
	// Non-string arg
	if res := csvCall("parse", objects.NewInt(123)); res.Type() != objects.ErrorType {
		t.Fatalf("parse() with int arg should error")
	}
	// Valid simple CSV
	csvData := "a,b,c\n1,2,3"
	if res := csvCall("parse", objects.NewString(csvData)); res.Type() != objects.ArrayType {
		t.Fatalf("parse() should return Array of rows")
	}
	// With custom delimiter (tab)
	tabCSV := "a\tb\n1\t2"
	if res := csvCall("parse", objects.NewString(tabCSV), objects.NewString("\t")); res.Type() != objects.ArrayType {
		t.Fatalf("parse() with tab delimiter should return Array")
	}
}

// TestCSVExtra2_ParseWithHeader tests csv.parseWithHeader.
func TestCSVExtra2_ParseWithHeader(t *testing.T) {
	// No args
	if res := csvCall("parseWithHeader"); res.Type() != objects.ErrorType {
		t.Fatalf("parseWithHeader() with no args should error")
	}
	// Non-string arg
	if res := csvCall("parseWithHeader", objects.NewInt(1)); res.Type() != objects.ErrorType {
		t.Fatalf("parseWithHeader() with int should error")
	}
	// Valid CSV with header
	csvData := "name,age\nAlice,30\nBob,25"
	res := csvCall("parseWithHeader", objects.NewString(csvData))
	if res.Type() != objects.ArrayType {
		t.Fatalf("parseWithHeader() should return Array of Maps")
	}
	// Verify that each element is a Map
	arr := res.(*objects.Array)
	if len(arr.Elements) > 0 {
		if arr.Elements[0].Type() != objects.MapType {
			t.Fatalf("parseWithHeader() row should be Map")
		}
	}
}

// TestCSVExtra2_Stringify tests csv.stringify.
func TestCSVExtra2_Stringify(t *testing.T) {
	// No args
	if res := csvCall("stringify"); res.Type() != objects.ErrorType {
		t.Fatalf("stringify() with no args should error")
	}
	// Non-array arg
	if res := csvCall("stringify", objects.NewString("not array")); res.Type() != objects.ErrorType {
		t.Fatalf("stringify() with string should error")
	}
	// Valid array of arrays
	rows := &objects.Array{
		Elements: []objects.Object{
			&objects.Array{Elements: []objects.Object{objects.NewString("a"), objects.NewString("b")}},
			&objects.Array{Elements: []objects.Object{objects.NewString("c"), objects.NewString("d")}},
		},
	}
	res := csvCall("stringify", rows)
	if res.Type() != objects.StringType {
		t.Fatalf("stringify() should return String")
	}
	s := res.(*objects.String).Value
	if !strings.Contains(s, "a,b") {
		t.Fatalf("stringify() output should contain 'a,b', got %s", s)
	}
	// With custom delimiter (tab)
	if res := csvCall("stringify", rows, objects.NewString("\t")); res.Type() != objects.StringType {
		t.Fatalf("stringify() with tab delimiter should return String")
	}
}

// TestCSVExtra2_StringifyMaps tests csv.stringifyMaps.
func TestCSVExtra2_StringifyMaps(t *testing.T) {
	// No args
	if res := csvCall("stringifyMaps"); res.Type() != objects.ErrorType {
		t.Fatalf("stringifyMaps() with no args should error")
	}
	// Only one arg
	if res := csvCall("stringifyMaps", &objects.Array{}); res.Type() != objects.ErrorType {
		t.Fatalf("stringifyMaps() with 1 arg should error")
	}
	// First arg non-array
	if res := csvCall("stringifyMaps", objects.NewString("arr"), &objects.Array{Elements: []objects.Object{objects.NewString("h")}}); res.Type() != objects.ErrorType {
		t.Fatalf("stringifyMaps() with non-array first arg should error")
	}
	// Second arg non-array
	if res := csvCall("stringifyMaps", &objects.Array{Elements: []objects.Object{&objects.Map{}}}, objects.NewString("h")); res.Type() != objects.ErrorType {
		t.Fatalf("stringifyMaps() with non-array second arg should error")
	}
	// Valid: array of maps and headers
	m1 := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Alice")},
		objects.NewString("age").HashKey():  {Key: objects.NewString("age"), Value: objects.NewInt(30)},
	}}
	m2 := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("name").HashKey(): {Key: objects.NewString("name"), Value: objects.NewString("Bob")},
		objects.NewString("age").HashKey():  {Key: objects.NewString("age"), Value: objects.NewInt(25)},
	}}
	arr := &objects.Array{Elements: []objects.Object{m1, m2}}
	headers := &objects.Array{Elements: []objects.Object{objects.NewString("name"), objects.NewString("age")}}
	res := csvCall("stringifyMaps", arr, headers)
	if res.Type() != objects.StringType {
		t.Fatalf("stringifyMaps() should return String")
	}
	s := res.(*objects.String).Value
	if !strings.Contains(s, "Alice") || !strings.Contains(s, "30") {
		t.Fatalf("stringifyMaps() output missing expected values: %s", s)
	}
}
