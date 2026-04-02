// pkg/objects/xlsx_test.go
// Tests for XLSX document object.
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXLSX_New(t *testing.T) {
	doc := NewXLSX()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.Type() != XLSXType {
		t.Errorf("expected XLSXType, got %v", doc.Type())
	}
	if doc.TypeTag() != TagXLSX {
		t.Errorf("expected TagXLSX, got %v", doc.TypeTag())
	}
	if !doc.ToBool().Value {
		t.Error("document should be truthy")
	}
}

func TestXLSX_Inspect(t *testing.T) {
	doc := NewXLSX()
	inspect := doc.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
}

func TestXLSX_HashKey(t *testing.T) {
	doc := NewXLSX()
	hk := doc.HashKey()
	if hk.Type != XLSXType {
		t.Errorf("expected XLSXType, got %v", hk.Type)
	}
}

func TestXLSX_OpenFixture(t *testing.T) {
	fixturePath := filepath.Join("..", "testutil", "fixtures", "documents", "sample.xlsx")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("fixture file not found")
	}

	doc, err := OpenXLSX(fixturePath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer doc.Close()

	if doc.Type() != XLSXType {
		t.Errorf("expected XLSXType, got %v", doc.Type())
	}

	// Test sheet operations
	sheets := doc.GetSheetList()
	if len(sheets) == 0 {
		t.Error("expected at least one sheet")
	}
}

func TestXLSX_GetSheetCount(t *testing.T) {
	doc := NewXLSX()
	count := doc.GetSheetCount()
	// New document should have default sheet
	_ = count
}

func TestOpenXLSX_NonExistent(t *testing.T) {
	_, err := OpenXLSX("/nonexistent/path/document.xlsx")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
