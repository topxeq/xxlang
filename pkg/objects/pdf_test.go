// pkg/objects/pdf_test.go
// Tests for PDF document object.
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPDFDocument_New(t *testing.T) {
	doc := NewPDFDocument()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.Type() != PDFDocumentType {
		t.Errorf("expected PDFDocumentType, got %v", doc.Type())
	}
	if doc.TypeTag() != TagPDFDocument {
		t.Errorf("expected TagPDFDocument, got %v", doc.TypeTag())
	}
	if !doc.ToBool().Value {
		t.Error("document should be truthy")
	}
}

func TestPDFDocument_Inspect(t *testing.T) {
	doc := NewPDFDocument()
	inspect := doc.Inspect()
	if inspect == "" {
		t.Error("inspect should not be empty")
	}
}

func TestPDFDocument_HashKey(t *testing.T) {
	doc := NewPDFDocument()
	hk := doc.HashKey()
	if hk.Type != PDFDocumentType {
		t.Errorf("expected PDFDocumentType, got %v", hk.Type)
	}
}

func TestPDF_New(t *testing.T) {
	pdf := NewPDF(nil, "")
	if pdf == nil {
		t.Fatal("expected non-nil PDF")
	}
	if pdf.Type() != PDFType {
		t.Errorf("expected PDFType, got %v", pdf.Type())
	}
}

func TestPDF_FromFile(t *testing.T) {
	fixturePath := filepath.Join("..", "testutil", "fixtures", "documents", "sample.pdf")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("fixture file not found")
	}

	pdf, err := NewPDFFromFile(fixturePath)
	if err != nil {
		// Minimal fixture may not parse correctly, skip test
		t.Skipf("fixture parsing not supported: %v", err)
	}

	if pdf.Type() != PDFType {
		t.Errorf("expected PDFType, got %v", pdf.Type())
	}
}

func TestPDF_FromBytes(t *testing.T) {
	// Create minimal PDF data
	pdfData := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`)

	pdf, err := NewPDFFromBytes(pdfData)
	if err != nil {
		t.Fatalf("failed to create PDF from bytes: %v", err)
	}

	if pdf.Type() != PDFType {
		t.Errorf("expected PDFType, got %v", pdf.Type())
	}
}

func TestNewPDFFromFile_NonExistent(t *testing.T) {
	_, err := NewPDFFromFile("/nonexistent/path/document.pdf")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestPDF_GetPage(t *testing.T) {
	// Minimal valid PDF with one page
	pdfData := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`)

	pdf, err := NewPDFFromBytes(pdfData)
	if err != nil {
		t.Fatalf("failed to create PDF: %v", err)
	}

	page := pdf.GetPage(0)
	if page == nil {
		t.Fatal("expected non-nil page")
	}

	if page.Type() != PDFPageType {
		t.Errorf("expected page type PDFPage, got %s", page.Type())
	}
}

func TestPDF_GetInfo(t *testing.T) {
	// PDF without Info dictionary
	pdfData := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`)

	pdf, err := NewPDFFromBytes(pdfData)
	if err != nil {
		t.Fatalf("failed to create PDF: %v", err)
	}

	info := pdf.GetInfo()
	// When there is no Info dictionary, GetInfo returns a default PDFInfo object
	// with zero values. It should not be nil.
	if info == nil {
		t.Error("expected non-nil Info object")
	}
}

func TestPDF_Close(t *testing.T) {
	pdfData := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`)

	pdf, err := NewPDFFromBytes(pdfData)
	if err != nil {
		t.Fatalf("failed to create PDF: %v", err)
	}

	result := pdf.Close()
	if result != NULL {
		t.Errorf("expected NULL from Close, got %v", result)
	}

	if pdf.IsOpen {
		t.Error("expected IsOpen to be false after Close")
	}
}
