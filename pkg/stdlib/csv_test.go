// pkg/stdlib/csv_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callCSVFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("csv")
	if mod == nil {
		panic("csv module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestCSVParse(t *testing.T) {
	result := callCSVFunc("parse", String("a,b,c\n1,2,3"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("parse() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("parse() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVParseWithComma(t *testing.T) {
	result := callCSVFunc("parse", String("a;b;c\n1;2;3"), String(";"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("parse() should return Array, got %T", result)
	}
	row := arr.Elements[0].(*objects.Array)
	if row.Elements[0].(*objects.String).Value != "a" {
		t.Errorf("parse() with semicolon delimiter failed")
	}
}

func TestCSVParseErrors(t *testing.T) {
	result := callCSVFunc("parse")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parse() with no args should return Error")
	}

	result = callCSVFunc("parse", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parse() with non-string should return Error")
	}
}

func TestCSVParseWithHeader(t *testing.T) {
	result := callCSVFunc("parseWithHeader", String("name,age\nJohn,30\nJane,25"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("parseWithHeader() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("parseWithHeader() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVStringify(t *testing.T) {
	data := Array(
		Array(String("a"), String("b")),
		Array(String("1"), String("2")),
	)
	result := callCSVFunc("stringify", data)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringify() should return String, got %T", result)
	}
	expected := "a,b\n1,2\n"
	if s.Value != expected {
		t.Errorf("stringify() = %q, want %q", s.Value, expected)
	}
}

func TestCSVStringifyMaps(t *testing.T) {
	data := Array(
		&objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			String("name").HashKey(): {Key: String("name"), Value: String("John")},
		}},
	)
	headers := Array(String("name"))
	result := callCSVFunc("stringifyMaps", data, headers)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stringifyMaps() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("stringifyMaps() should return non-empty string")
	}
}

func TestCSVColumn(t *testing.T) {
	data := Array(
		Array(String("a"), String("b"), String("c")),
		Array(String("1"), String("2"), String("3")),
	)
	result := callCSVFunc("column", data, Int(0))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("column() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("column() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVRow(t *testing.T) {
	data := Array(
		Array(String("a"), String("b")),
		Array(String("1"), String("2")),
	)
	result := callCSVFunc("row", data, Int(1))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("row() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("row() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVTranspose(t *testing.T) {
	data := Array(
		Array(String("a"), String("b")),
		Array(String("1"), String("2")),
	)
	result := callCSVFunc("transpose", data)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("transpose() should return Array, got %T", result)
	}
	// Transposed should have 2 columns -> 2 rows
	if len(arr.Elements) != 2 {
		t.Errorf("transpose() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVRowCount(t *testing.T) {
	data := Array(
		Array(String("a")),
		Array(String("b")),
		Array(String("c")),
	)
	result := callCSVFunc("rowCount", data)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("rowCount() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("rowCount() = %d, want 3", i.Value)
	}
}

func TestCSVColCount(t *testing.T) {
	data := Array(
		Array(String("a"), String("b"), String("c")),
	)
	result := callCSVFunc("colCount", data)
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("colCount() should return Int, got %T", result)
	}
	if i.Value != 3 {
		t.Errorf("colCount() = %d, want 3", i.Value)
	}
}

func TestCSVSkip(t *testing.T) {
	data := Array(
		Array(String("a")),
		Array(String("b")),
		Array(String("c")),
	)
	result := callCSVFunc("skip", data, Int(1))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("skip() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("skip() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVTake(t *testing.T) {
	data := Array(
		Array(String("a")),
		Array(String("b")),
		Array(String("c")),
	)
	result := callCSVFunc("take", data, Int(2))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("take() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("take() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVAppendRow(t *testing.T) {
	data := Array(Array(String("a")))
	newRow := Array(String("b"))
	result := callCSVFunc("appendRow", data, newRow)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("appendRow() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("appendRow() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVPrependRow(t *testing.T) {
	data := Array(Array(String("a")))
	newRow := Array(String("b"))
	result := callCSVFunc("prependRow", data, newRow)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("prependRow() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("prependRow() length = %d, want 2", len(arr.Elements))
	}
	// First element should be "b"
	if arr.Elements[0].(*objects.Array).Elements[0].(*objects.String).Value != "b" {
		t.Error("prependRow() should put new row first")
	}
}

func TestCSVFilterRows(t *testing.T) {
	data := Array(
		Array(String("a"), Int(1)),
		Array(String("b"), Int(2)),
		Array(String("c"), Int(3)),
	)
	// Filter rows where second element > 1
	pred := BuiltinFunc(func(args ...objects.Object) objects.Object {
		arr := args[0].(*objects.Array)
		return Bool(arr.Elements[1].(*objects.Int).Value > 1)
	})
	result := callCSVFunc("filterRows", data, pred)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("filterRows() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("filterRows() length = %d, want 2", len(arr.Elements))
	}
}

func TestCSVMapRows(t *testing.T) {
	data := Array(
		Array(String("a")),
		Array(String("b")),
	)
	// Map rows to their first element
	fn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		return args[0].(*objects.Array).Elements[0]
	})
	result := callCSVFunc("mapRows", data, fn)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("mapRows() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("mapRows() length = %d, want 2", len(arr.Elements))
	}
}
