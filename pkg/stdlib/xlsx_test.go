// pkg/stdlib/xlsx_test.go
// Tests for xlsx module.
package stdlib

import (
	"os"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callXLSXFunc calls a function from the xlsx module.
func callXLSXFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("xlsx")
	if mod == nil {
		t := &testing.T{}
		t.Skip("xlsx module not found")
		return &objects.Error{Message: "xlsx module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

func TestXLSXCreate(t *testing.T) {
	wb := callXLSXFunc("create")
	if wb.Type() != objects.XLSXType {
		t.Fatalf("expected XLSX, got %s", wb.Type())
	}
}

func TestXLSXIsXLSX(t *testing.T) {
	mod := Get("xlsx")
	if mod == nil {
		t.Skip("xlsx module not found")
	}
	fn := mod.Exports["isXLSX"].(*objects.Builtin)

	// Not an XLSX
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	// XLSX object
	wb := objects.NewXLSX()
	res = fn.Fn(wb)
	if b, ok := res.(*objects.Bool); ok {
		if !b.Value {
			t.Error("expected true for XLSX")
		}
	}
}

func TestXLSXColToIndex(t *testing.T) {
	tests := []struct {
		col   string
		index int64
	}{
		{"A", 1},
		{"Z", 26},
		{"AA", 27},
		{"AB", 28},
		{"AZ", 52},
		{"BA", 53},
		{"ZZ", 702},
		{"AAA", 703},
	}

	for _, tt := range tests {
		result := callXLSXFunc("colToIndex", objects.NewString(tt.col))
		if num, ok := result.(*objects.Int); ok {
			if num.Value != tt.index {
				t.Errorf("colToIndex(%s) expected %d, got %d", tt.col, tt.index, num.Value)
			}
		} else {
			t.Fatalf("colToIndex(%s) returned non-Int: %T", tt.col, result)
		}
	}

	// Invalid column (empty)
	bad := callXLSXFunc("colToIndex", objects.NewString(""))
	if bad.Type() != objects.ErrorType {
		t.Logf("colToIndex('') returned %s (expected error or default)", bad.Type())
	}
}

func TestXLSXIndexToCol(t *testing.T) {
	tests := []struct {
		index int64
		col   string
	}{
		{1, "A"},
		{26, "Z"},
		{27, "AA"},
		{28, "AB"},
		{52, "AZ"},
		{53, "BA"},
		{702, "ZZ"},
		{703, "AAA"},
	}

	for _, tt := range tests {
		result := callXLSXFunc("indexToCol", objects.NewInt(tt.index))
		if str, ok := result.(*objects.String); ok {
			if str.Value != tt.col {
				t.Errorf("indexToCol(%d) expected %s, got %s", tt.index, tt.col, str.Value)
			}
		} else {
			t.Fatalf("indexToCol(%d) returned non-String: %T", tt.index, result)
		}
	}

	// Zero or negative index
	zero := callXLSXFunc("indexToCol", objects.NewInt(0))
	if zero.Type() != objects.ErrorType {
		t.Logf("indexToCol(0) returned %s (expected error)", zero.Type())
	}
}

func TestXLSXParseCellRef(t *testing.T) {
	tests := []struct {
		ref string
		col string
		row int64
	}{
		{"A1", "A", 1},
		{"Z100", "Z", 100},
		{"AA500", "AA", 500},
		{"XFD1048576", "XFD", 1048576}, // Excel max row
	}

	for _, tt := range tests {
		result := callXLSXFunc("parseCellRef", objects.NewString(tt.ref))
		if arr, ok := result.(*objects.Array); ok {
			if len(arr.Elements) != 2 {
				t.Errorf("parseCellRef(%s) expected array of 2, got %d", tt.ref, len(arr.Elements))
				continue
			}
			colObj, ok1 := arr.Elements[0].(*objects.String)
			rowObj, ok2 := arr.Elements[1].(*objects.Int)
			if !ok1 || !ok2 {
				t.Fatalf("parseCellRef(%s) returned wrong element types", tt.ref)
			}
			if colObj.Value != tt.col {
				t.Errorf("parseCellRef(%s) col expected %s, got %s", tt.ref, tt.col, colObj.Value)
			}
			if rowObj.Value != tt.row {
				t.Errorf("parseCellRef(%s) row expected %d, got %d", tt.ref, tt.row, rowObj.Value)
			}
		} else {
			t.Fatalf("parseCellRef(%s) returned non-Array: %T", tt.ref, result)
		}
	}

	// Invalid reference
	bad := callXLSXFunc("parseCellRef", objects.NewString("A"))
	if bad.Type() != objects.ErrorType {
		t.Logf("parseCellRef('A') returned %s (expected error)", bad.Type())
	}
}

func TestXLSXOpen(t *testing.T) {
	// Test opening nonexistent file
	missing := callXLSXFunc("open", objects.NewString("nonexistent.xlsx"))
	if missing.Type() != objects.ErrorType {
		t.Errorf("expected error for nonexistent file, got %s", missing.Type())
	}

	// Test opening a real file: create one then open it
	tmpFile := t.TempDir() + "/test.xlsx"
	wb := objects.NewXLSX()
	// Add a simple sheet with some data? The XLSX object might need to be populated.
	// For basic open test, we can just save an empty workbook.
	if err := wb.Save(tmpFile); err != nil {
		t.Fatalf("failed to save test xlsx: %v", err)
	}
	defer os.Remove(tmpFile)

	opened := callXLSXFunc("open", objects.NewString(tmpFile))
	if opened.Type() != objects.XLSXType {
		t.Fatalf("expected XLSX from open, got %s", opened.Type())
	}
	// Close the opened workbook to release file handle
	if xlsx, ok := opened.(*objects.XLSX); ok {
		xlsx.Close()
	}
}
