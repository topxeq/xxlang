// pkg/objects/pptx_test.go
// Tests for PPTX document object.
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPPTXDocument_New(t *testing.T) {
	doc := NewPPTX()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.Type() != PPTXDocumentType {
		t.Errorf("expected PPTXDocumentType, got %v", doc.Type())
	}
	if doc.TypeTag() != TagPPTXDocument {
		t.Errorf("expected TagPPTXDocument, got %v", doc.TypeTag())
	}
	if !doc.ToBool().Value {
		t.Error("document should be truthy")
	}
}

func TestPPTXDocument_Inspect(t *testing.T) {
	doc := NewPPTX()
	inspect := doc.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
}

func TestPPTXDocument_HashKey(t *testing.T) {
	doc := NewPPTX()
	hk := doc.HashKey()
	if hk.Type != PPTXDocumentType {
		t.Errorf("expected PPTXDocumentType, got %v", hk.Type)
	}
}

func TestPPTXDocument_OpenFixture(t *testing.T) {
	fixturePath := filepath.Join("..", "testutil", "fixtures", "documents", "sample.pptx")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("fixture file not found")
	}

	doc, err := OpenPPTX(fixturePath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer doc.Close()

	if doc.Type() != PPTXDocumentType {
		t.Errorf("expected PPTXDocumentType, got %v", doc.Type())
	}
}

func TestOpenPPTX_NonExistent(t *testing.T) {
	_, err := OpenPPTX("/nonexistent/path/document.pptx")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
