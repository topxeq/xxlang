// pkg/objects/docx_test.go
// Tests for DOCX document object.
package objects

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocxDocument_New(t *testing.T) {
	doc := NewDocxDocument()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.Type() != DocxDocumentType {
		t.Errorf("expected DocxDocumentType, got %v", doc.Type())
	}
	if doc.TypeTag() != TagDocxDocument {
		t.Errorf("expected TagDocxDocument, got %v", doc.TypeTag())
	}
	if !doc.ToBool().Value {
		t.Error("document should be truthy")
	}
}

func TestDocxDocument_Inspect(t *testing.T) {
	doc := NewDocxDocument()
	inspect := doc.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
}

func TestDocxDocument_HashKey(t *testing.T) {
	doc := NewDocxDocument()
	hk := doc.HashKey()
	if hk.Type != DocxDocumentType {
		t.Errorf("expected DocxDocumentType, got %v", hk.Type)
	}
}

func TestDocxDocument_Modified(t *testing.T) {
	doc := NewDocxDocument()
	if doc.IsModified() {
		t.Error("new document should not be modified")
	}
	doc.SetModified(true)
	if !doc.IsModified() {
		t.Error("document should be modified after setting")
	}
}

func TestDocxDocument_Properties(t *testing.T) {
	doc := NewDocxDocument()
	props := doc.GetProperties()
	if props == nil {
		t.Fatal("expected non-nil properties")
	}
	// New document has empty properties
	if props.Title != "" {
		t.Errorf("expected empty title, got %s", props.Title)
	}
}

func TestDocxDocument_GetXMLDoc(t *testing.T) {
	doc := NewDocxDocument()
	xmlDoc := doc.GetXMLDoc()
	if xmlDoc == nil {
		t.Error("expected non-nil XML document")
	}
}

func TestDocxDocument_GetBody(t *testing.T) {
	doc := NewDocxDocument()
	body := doc.GetBody()
	// New document should have a body
	if body == nil {
		t.Error("expected non-nil body")
	}
}

func TestDocxDocument_SaveAndOpen(t *testing.T) {
	doc := NewDocxDocument()

	// Create temp file
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, "test_docx_save.docx")
	defer os.Remove(tmpPath)

	// Save document
	err := doc.Save(tmpPath)
	if err != nil {
		t.Fatalf("failed to save document: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		t.Fatal("saved file does not exist")
	}

	// Open the saved document
	openedDoc, err := OpenDocx(tmpPath)
	if err != nil {
		t.Fatalf("failed to open saved document: %v", err)
	}
	defer openedDoc.Close()

	if openedDoc.Type() != DocxDocumentType {
		t.Errorf("expected DocxDocumentType, got %v", openedDoc.Type())
	}
}

func TestDocxDocument_ToBytes(t *testing.T) {
	doc := NewDocxDocument()

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("failed to get document bytes: %v", err)
	}
	if len(data) == 0 {
		t.Error("document bytes should not be empty")
	}
}

func TestDocxDocument_OpenFromBytes(t *testing.T) {
	// Create a document and get its bytes
	doc := NewDocxDocument()
	data, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("failed to get document bytes: %v", err)
	}

	// Open from bytes
	openedDoc, err := OpenDocxFromBytes(data)
	if err != nil {
		t.Fatalf("failed to open document from bytes: %v", err)
	}

	if openedDoc.Type() != DocxDocumentType {
		t.Errorf("expected DocxDocumentType, got %v", openedDoc.Type())
	}
}

func TestDocxDocument_OpenFixture(t *testing.T) {
	// Get fixture path
	fixturePath := filepath.Join("..", "testutil", "fixtures", "documents", "sample.docx")

	// Check if fixture exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("fixture file not found")
	}

	// Open fixture
	doc, err := OpenDocx(fixturePath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer doc.Close()

	// Verify document is valid
	if doc.Type() != DocxDocumentType {
		t.Errorf("expected DocxDocumentType, got %v", doc.Type())
	}

	// Get body
	body := doc.GetBody()
	if body == nil {
		t.Error("expected non-nil body")
	}
}

func TestDocxDocument_AddRelationship(t *testing.T) {
	doc := NewDocxDocument()
	relID := doc.AddRelationship("http://schemas.openxmlformats.org/officeDocument/2006/relationships/image", "media/image1.png")
	if relID == "" {
		t.Error("expected non-empty relationship ID")
	}
}

func TestDocxDocument_FindElements(t *testing.T) {
	doc := NewDocxDocument()
	// Try to find elements - may return empty for new document
	elements := doc.FindElements("//w:p")
	// Should not error, returns array
	if elements == nil {
		t.Error("FindElements should return an array")
	}
}

func TestDocxDocument_FindFirstElement(t *testing.T) {
	doc := NewDocxDocument()
	// Try to find first element - may return nil for new document
	node := doc.FindFirstElement("//w:p")
	// Should not error
	_ = node
}

func TestDocxDocument_Close(t *testing.T) {
	doc := NewDocxDocument()
	err := doc.Close()
	if err != nil {
		t.Errorf("close should not error: %v", err)
	}
}

func TestOpenDocx_NonExistent(t *testing.T) {
	_, err := OpenDocx("/nonexistent/path/document.docx")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestOpenDocxFromBytes_InvalidData(t *testing.T) {
	// Try to open invalid data
	_, err := OpenDocxFromBytes([]byte("not a valid docx"))
	if err == nil {
		t.Error("expected error for invalid data")
	}
}

// Tests for uncovered docx object types (coverage gaps)

func TestDocxTOC(t *testing.T) {
	doc := NewDocxDocument()
	toc := &DocxTOC{document: doc, xmlNode: NewXMLNode("w:toc")}

	if toc.Type() != DocxTOCType {
		t.Errorf("expected DocxTOCType, got %v", toc.Type())
	}
	if toc.TypeTag() != TagDocxTOC {
		t.Errorf("expected TagDocxTOC, got %v", toc.TypeTag())
	}
	if !toc.ToBool().Value {
		t.Error("TOC should be truthy")
	}
	hk := toc.HashKey()
	if hk.Type != DocxTOCType {
		t.Errorf("expected HashKey.Type DocxTOCType, got %v", hk.Type)
	}
	inspect := toc.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
}

func TestDocxTextBox(t *testing.T) {
	doc := NewDocxDocument()
	box := &DocxTextBox{document: doc, xmlNode: NewXMLNode("w:textbox")}

	if box.Type() != DocxTextBoxType {
		t.Errorf("expected DocxTextBoxType, got %v", box.Type())
	}
	if box.TypeTag() != TagDocxTextBox {
		t.Errorf("expected TagDocxTextBox, got %v", box.TypeTag())
	}
	if !box.ToBool().Value {
		t.Error("TextBox should be truthy")
	}
	hk := box.HashKey()
	if hk.Type != DocxTextBoxType {
		t.Errorf("expected HashKey.Type DocxTextBoxType, got %v", hk.Type)
	}
	if box.Inspect() == "" {
		t.Error("inspect should not be empty")
	}
}

func TestDocxShape(t *testing.T) {
	doc := NewDocxDocument()
	shape := &DocxShape{document: doc, xmlNode: NewXMLNode("w:shape")}

	if shape.Type() != DocxShapeType {
		t.Errorf("expected DocxShapeType, got %v", shape.Type())
	}
	if shape.TypeTag() != TagDocxShape {
		t.Errorf("expected TagDocxShape, got %v", shape.TypeTag())
	}
	if !shape.ToBool().Value {
		t.Error("Shape should be truthy")
	}
	hk := shape.HashKey()
	if hk.Type != DocxShapeType {
		t.Errorf("expected HashKey.Type DocxShapeType, got %v", hk.Type)
	}
	if shape.Inspect() == "" {
		t.Error("inspect should not be empty")
	}
}

func TestDocxChart(t *testing.T) {
	doc := NewDocxDocument()
	chart := &DocxChart{document: doc, xmlNode: NewXMLNode("w:chart"), relationID: "rId1"}

	if chart.Type() != DocxChartType {
		t.Errorf("expected DocxChartType, got %v", chart.Type())
	}
	if chart.TypeTag() != TagDocxChart {
		t.Errorf("expected TagDocxChart, got %v", chart.TypeTag())
	}
	if !chart.ToBool().Value {
		t.Error("Chart should be truthy")
	}
	hk := chart.HashKey()
	if hk.Type != DocxChartType {
		t.Errorf("expected HashKey.Type DocxChartType, got %v", hk.Type)
	}
	if chart.Inspect() == "" {
		t.Error("inspect should not be empty")
	}
}

func TestDocxComment(t *testing.T) {
	doc := NewDocxDocument()
	comment := &DocxComment{document: doc, id: 1, author: "tester", date: "2024-01-01", content: nil}

	if comment.Type() != DocxCommentType {
		t.Errorf("expected DocxCommentType, got %v", comment.Type())
	}
	if comment.TypeTag() != TagDocxComment {
		t.Errorf("expected TagDocxComment, got %v", comment.TypeTag())
	}
	if !comment.ToBool().Value {
		t.Error("Comment should be truthy")
	}
	hk := comment.HashKey()
	if hk.Type != DocxCommentType {
		t.Errorf("expected HashKey.Type DocxCommentType, got %v", hk.Type)
	}
	inspect := comment.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
	if !strings.Contains(inspect, "id=1") || !strings.Contains(inspect, "author=tester") {
		t.Errorf("inspect should contain id and author: %s", inspect)
	}
}

func TestDocxRevision(t *testing.T) {
	doc := NewDocxDocument()
	rev := &DocxRevision{document: doc, id: 1, revType: "insert", author: "tester", date: "2024-01-01", content: "test"}

	if rev.Type() != DocxRevisionType {
		t.Errorf("expected DocxRevisionType, got %v", rev.Type())
	}
	if rev.TypeTag() != TagDocxRevision {
		t.Errorf("expected TagDocxRevision, got %v", rev.TypeTag())
	}
	if !rev.ToBool().Value {
		t.Error("Revision should be truthy")
	}
	hk := rev.HashKey()
	if hk.Type != DocxRevisionType {
		t.Errorf("expected HashKey.Type DocxRevisionType, got %v", hk.Type)
	}
	inspect := rev.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
	if !strings.Contains(inspect, "id=1") || !strings.Contains(inspect, "type=insert") {
		t.Errorf("inspect should contain id and type: %s", inspect)
	}
}

func TestDocxFootnote(t *testing.T) {
	doc := NewDocxDocument()
	fn := &DocxFootnote{document: doc, id: 1}

	if fn.Type() != DocxFootnoteType {
		t.Errorf("expected DocxFootnoteType, got %v", fn.Type())
	}
	if fn.TypeTag() != TagDocxFootnote {
		t.Errorf("expected TagDocxFootnote, got %v", fn.TypeTag())
	}
	if !fn.ToBool().Value {
		t.Error("Footnote should be truthy")
	}
	hk := fn.HashKey()
	if hk.Type != DocxFootnoteType {
		t.Errorf("expected HashKey.Type DocxFootnoteType, got %v", hk.Type)
	}
	inspect := fn.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
	if !strings.Contains(inspect, "id=1") {
		t.Errorf("inspect should contain id: %s", inspect)
	}
}

func TestDocxEndnote(t *testing.T) {
	doc := NewDocxDocument()
	en := &DocxEndnote{document: doc, id: 1}

	if en.Type() != DocxEndnoteType {
		t.Errorf("expected DocxEndnoteType, got %v", en.Type())
	}
	if en.TypeTag() != TagDocxEndnote {
		t.Errorf("expected TagDocxEndnote, got %v", en.TypeTag())
	}
	if !en.ToBool().Value {
		t.Error("Endnote should be truthy")
	}
	hk := en.HashKey()
	if hk.Type != DocxEndnoteType {
		t.Errorf("expected HashKey.Type DocxEndnoteType, got %v", hk.Type)
	}
	inspect := en.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
	if !strings.Contains(inspect, "id=1") {
		t.Errorf("inspect should contain id: %s", inspect)
	}
}

// Additional tests for other uncovered methods in docx.go
// These cover methods like parseZipContents, parseRelationships, etc.
// Those are exercised indirectly by OpenDocx and Save tests above,
// but we can add more specific edge case tests if needed.
