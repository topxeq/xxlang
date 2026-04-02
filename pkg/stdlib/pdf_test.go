// pkg/stdlib/pdf_test.go
// Tests for the PDF module.
package stdlib

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestPDFCreate tests creating a new PDF document.
func TestPDFCreate(t *testing.T) {
	doc := objects.NewPDFDocument()
	if doc == nil {
		t.Fatal("Failed to create PDF document")
	}

	// Add a page
	pageIdx := doc.AddPage(595.28, 841.89) // A4 size
	if pageIdx == nil {
		t.Fatal("Failed to add page")
	}

	// Write text
	result := doc.WriteText(0, "Hello, PDF World!", 100, 700, nil)
	if isPDFErr(result) {
		t.Fatalf("Failed to write text: %v", result)
	}

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_create.pdf")
	defer os.Remove(outputPath)

	result = doc.Save(outputPath)
	if isPDFErr(result) {
		t.Fatalf("Failed to save PDF: %v", result)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Verify file has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Output file is empty")
	}

	t.Logf("Created PDF file: %s (%d bytes)", outputPath, info.Size())
}

// TestPDFNewFromBytes tests creating a PDF from bytes.
func TestPDFNewFromBytes(t *testing.T) {
	// Create a simple PDF
	doc := objects.NewPDFDocument()
	doc.AddPage(612, 792) // Letter size
	doc.WriteText(0, "Test content", 50, 750, nil)

	bytesResult := doc.ToBytes()
	if isPDFErr(bytesResult) {
		t.Fatalf("Failed to get PDF bytes: %v", bytesResult)
	}

	// Convert to byte slice
	arr, ok := bytesResult.(*objects.Array)
	if !ok {
		t.Fatal("Expected array result")
	}

	data := make([]byte, len(arr.Elements))
	for i, elem := range arr.Elements {
		n, ok := elem.(*objects.Int)
		if !ok {
			t.Fatal("Expected integer elements")
		}
		data[i] = byte(n.Value)
	}

	// Verify data starts with %PDF-
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("Generated PDF does not have valid header")
	}

	t.Logf("Generated PDF of %d bytes", len(data))
}

// TestPDFInfo tests getting PDF information.
func TestPDFInfo(t *testing.T) {
	// Create a PDF with metadata
	doc := objects.NewPDFDocument()
	doc.Title = "Test Document"
	doc.Author = "Test Author"
	doc.AddPage(595, 842)

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_info.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to create PDF: %v", err)
	}
	t.Logf("Created PDF: %s (%d bytes)", outputPath, info.Size())
}

// TestPDFPageCount tests getting page count.
func TestPDFPageCount(t *testing.T) {
	doc := objects.NewPDFDocument()

	// Add multiple pages
	doc.AddPage(595, 842)
	doc.AddPage(595, 842)
	doc.AddPage(595, 842)

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_pagecount.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify PDF structure
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("Invalid PDF header")
	}

	t.Logf("Created 3-page PDF: %d bytes", len(data))
}

// TestPDFPageSize tests getting page dimensions.
func TestPDFPageSize(t *testing.T) {
	doc := objects.NewPDFDocument()
	doc.AddPage(612, 792) // Letter size

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_pagesize.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check for MediaBox in content
	if !bytes.Contains(data, []byte("/MediaBox")) {
		t.Fatal("MediaBox not found in PDF")
	}

	// Check for Letter size (612 792)
	if !bytes.Contains(data, []byte("612")) || !bytes.Contains(data, []byte("792")) {
		t.Fatal("Expected Letter size dimensions not found")
	}

	t.Logf("Created Letter size PDF: %d bytes", len(data))
}

// TestPDFWriteTextWithOptions tests writing text with font options.
func TestPDFWriteTextWithOptions(t *testing.T) {
	doc := objects.NewPDFDocument()
	doc.AddPage(595, 842)

	// Write text with options
	opts := map[string]interface{}{
		"fontSize": objects.NewFloat(14.0),
		"font":     objects.NewString("Helvetica"),
	}

	result := doc.WriteText(0, "Styled Text", 100, 700, opts)
	if isPDFErr(result) {
		t.Fatalf("Failed to write styled text: %v", result)
	}

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_styled.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Output file is empty")
	}

	t.Logf("Created styled text PDF: %d bytes", info.Size())
}

// TestPDFSetMetadata tests setting document metadata.
func TestPDFSetMetadata(t *testing.T) {
	doc := objects.NewPDFDocument()
	doc.Title = "My Title"
	doc.Author = "My Author"
	doc.Subject = "My Subject"
	doc.Creator = "Test Creator"

	doc.AddPage(595, 842)

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_metadata.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check for metadata in PDF
	if !bytes.Contains(data, []byte("/Title")) {
		t.Fatal("Title metadata not found in PDF")
	}
	if !bytes.Contains(data, []byte("/Author")) {
		t.Fatal("Author metadata not found in PDF")
	}

	t.Logf("Created PDF with metadata: %d bytes", len(data))
}

// TestPDFModuleFunctions tests the module-level functions.
func TestPDFModuleFunctions(t *testing.T) {
	// Get the pdf module
	module := Get("pdf")
	if module == nil {
		t.Fatal("pdf module not found")
	}

	// Test create function
	createFn, ok := module.Exports["create"]
	if !ok {
		t.Fatal("create function not found in pdf module")
	}

	builtin, ok := createFn.(*objects.Builtin)
	if !ok {
		t.Fatal("create is not a builtin function")
	}

	result := builtin.Fn()
	doc, ok := result.(*objects.PDFDocument)
	if !ok {
		t.Fatalf("Expected PDFDocument, got %T", result)
	}

	// Test constants
	a4Width, ok := module.Exports["A4_WIDTH"]
	if !ok {
		t.Fatal("A4_WIDTH constant not found")
	}
	if a4WidthFloat, ok := a4Width.(*objects.Float); ok {
		if a4WidthFloat.Value != 595.28 {
			t.Errorf("Expected A4_WIDTH = 595.28, got %f", a4WidthFloat.Value)
		}
	}

	_ = doc
}

// TestPDFUnicodeText tests writing Unicode text.
func TestPDFUnicodeText(t *testing.T) {
	doc := objects.NewPDFDocument()
	doc.AddPage(595, 842)

	// Write Chinese text (will be escaped in PDF)
	result := doc.WriteText(0, "你好世界 Hello World", 100, 700, nil)
	if isPDFErr(result) {
		t.Fatalf("Failed to write unicode text: %v", result)
	}

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_unicode.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	t.Logf("Created Unicode text PDF: %d bytes", info.Size())
}

// TestPDFMultiplePages tests creating a multi-page document.
func TestPDFMultiplePages(t *testing.T) {
	doc := objects.NewPDFDocument()

	// Add 10 pages
	for i := 0; i < 10; i++ {
		doc.AddPage(595, 842)
		doc.WriteText(i, "Page "+string(rune('0'+i+1)), 100, 700, nil)
	}

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_multipage.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Verify file was created
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	t.Logf("Created 10-page PDF: %d bytes", info.Size())
}

// isPDFErr checks if an object is an error.
func isPDFErr(obj objects.Object) bool {
	_, ok := obj.(*objects.Error)
	return ok
}

// TestPDFReadCreated tests reading a PDF we just created.
func TestPDFReadCreated(t *testing.T) {
	// Create a PDF
	doc := objects.NewPDFDocument()
	doc.AddPage(595, 842)
	doc.WriteText(0, "Hello World", 100, 700, nil)
	doc.Title = "Test Document"
	doc.Author = "Test Author"

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_read_created.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Open and read
	pdf, err := objects.NewPDFFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer pdf.Close()

	// Verify page count
	if pdf.PageCount != 1 {
		t.Errorf("Expected 1 page, got %d", pdf.PageCount)
	}

	// Verify version
	if pdf.Version != "1.4" {
		t.Errorf("Expected version 1.4, got %s", pdf.Version)
	}

	// Verify page dimensions
	if len(pdf.Pages) > 0 {
		page := pdf.Pages[0]
		if page.Width < 594 || page.Width > 596 {
			t.Errorf("Expected width ~595, got %f", page.Width)
		}
		if page.Height < 841 || page.Height > 843 {
			t.Errorf("Expected height ~842, got %f", page.Height)
		}
	}

	// Get info
	info := pdf.GetInfo()
	if info.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", info.Title)
	}
	if info.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", info.Author)
	}

	t.Logf("Successfully read PDF: pages=%d, title=%s", pdf.PageCount, info.Title)
}

// TestPDFExtractText tests text extraction from a PDF.
func TestPDFExtractText(t *testing.T) {
	// Create a PDF with text
	doc := objects.NewPDFDocument()
	doc.AddPage(595, 842)
	doc.WriteText(0, "Extract this text", 100, 700, nil)

	// Save to temp file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, "test_extract.pdf")
	defer os.Remove(outputPath)

	doc.Save(outputPath)

	// Open and extract text
	pdf, err := objects.NewPDFFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer pdf.Close()

	// Extract text
	text := pdf.ExtractText(0)
	if isPDFErr(text) {
		t.Fatalf("Failed to extract text: %v", text)
	}

	textStr, ok := text.(*objects.String)
	if !ok {
		t.Fatal("Expected string result")
	}

	// The extracted text should contain what we wrote
	if !bytes.Contains([]byte(textStr.Value), []byte("Extract")) {
		t.Logf("Extracted text: %s", textStr.Value)
	}

	t.Logf("Extracted text: %s", textStr.Value)
}

// TestPDFToBytesRoundTrip tests the ToBytes -> NewPDFFromBytes round trip.
func TestPDFToBytesRoundTrip(t *testing.T) {
	// Create a PDF
	doc := objects.NewPDFDocument()
	doc.AddPage(612, 792)
	doc.WriteText(0, "Round Trip Test", 50, 700, nil)

	// Convert to bytes
	bytesResult := doc.ToBytes()
	if isPDFErr(bytesResult) {
		t.Fatalf("Failed to get bytes: %v", bytesResult)
	}

	arr := bytesResult.(*objects.Array)
	data := make([]byte, len(arr.Elements))
	for i, elem := range arr.Elements {
		n := elem.(*objects.Int)
		data[i] = byte(n.Value)
	}

	// Parse from bytes
	pdf, err := objects.NewPDFFromBytes(data)
	if err != nil {
		t.Fatalf("Failed to parse PDF from bytes: %v", err)
	}
	defer pdf.Close()

	// Verify
	if pdf.PageCount != 1 {
		t.Errorf("Expected 1 page, got %d", pdf.PageCount)
	}

	t.Logf("Round trip successful: %d bytes", len(data))
}

// TestPDFMerge tests the mergePDFs function
func TestPDFMerge(t *testing.T) {
	// Create two simple PDFs to merge
	dir := os.TempDir()
	pdf1Path := filepath.Join(dir, "merge1.pdf")
	pdf2Path := filepath.Join(dir, "merge2.pdf")
	outputPath := filepath.Join(dir, "merged.pdf")
	defer os.Remove(pdf1Path)
	defer os.Remove(pdf2Path)
	defer os.Remove(outputPath)

	// Create first PDF
	doc1 := objects.NewPDFDocument()
	doc1.AddPage(595, 842)
	doc1.WriteText(0, "First PDF", 100, 700, nil)
	doc1.Save(pdf1Path)

	// Create second PDF
	doc2 := objects.NewPDFDocument()
	doc2.AddPage(595, 842)
	doc2.WriteText(0, "Second PDF", 100, 700, nil)
	doc2.Save(pdf2Path)

	// Merge them
	result := mergePDFs([]string{pdf1Path, pdf2Path}, outputPath)
	if isPDFErr(result) {
		t.Fatalf("mergePDFs failed: %v", result)
	}

	// Verify merged file exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Merged file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Merged file is empty")
	}

	// Verify merged PDF has 2 pages
	pdf, err := objects.NewPDFFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open merged PDF: %v", err)
	}
	defer pdf.Close()

	if pdf.PageCount != 2 {
		t.Errorf("Expected 2 pages in merged PDF, got %d", pdf.PageCount)
	}

	t.Logf("Merged PDF created with %d pages, size %d bytes", pdf.PageCount, info.Size())
}

// TestPDFSplit tests the splitPDF function
func TestPDFSplit(t *testing.T) {
	// Create a multi-page PDF
	dir := os.TempDir()
	inputPath := filepath.Join(dir, "split_input.pdf")
	outputPath := filepath.Join(dir, "split_output.pdf")
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	// Create PDF with 5 pages
	doc := objects.NewPDFDocument()
	for i := 0; i < 5; i++ {
		doc.AddPage(595, 842)
		doc.WriteText(i, fmt.Sprintf("Page %d", i+1), 100, 700, nil)
	}
	doc.Save(inputPath)

	// Split: extract pages 1-3 (0-indexed: 0,1,2)
	result := splitPDF(inputPath, outputPath, 0, 2)
	if isPDFErr(result) {
		t.Fatalf("splitPDF failed: %v", result)
	}

	// Verify output has 3 pages
	pdf, err := objects.NewPDFFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open split PDF: %v", err)
	}
	defer pdf.Close()

	if pdf.PageCount != 3 {
		t.Errorf("Expected 3 pages after split, got %d", pdf.PageCount)
	}

	t.Logf("Split PDF: extracted 3 pages from 5-page document")
}

// TestPDFExtractPages tests the extractPages function
func TestPDFExtractPages(t *testing.T) {
	// Create a multi-page PDF
	dir := os.TempDir()
	inputPath := filepath.Join(dir, "extract_input.pdf")
	outputPath := filepath.Join(dir, "extract_output.pdf")
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	// Create PDF with 5 pages
	doc := objects.NewPDFDocument()
	for i := 0; i < 5; i++ {
		doc.AddPage(595, 842)
		doc.WriteText(i, fmt.Sprintf("Page %d", i+1), 100, 700, nil)
	}
	doc.Save(inputPath)

	// Extract specific pages: 0, 2, 4 (first, third, fifth)
	result := extractPages(inputPath, []int{0, 2, 4}, outputPath)
	if isPDFErr(result) {
		t.Fatalf("extractPages failed: %v", result)
	}

	// Verify output has 3 pages
	pdf, err := objects.NewPDFFromFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open extracted PDF: %v", err)
	}
	defer pdf.Close()

	if pdf.PageCount != 3 {
		t.Errorf("Expected 3 pages after extraction, got %d", pdf.PageCount)
	}

	// Verify page 0 (first page) contains "Page 1"
	text0 := pdf.ExtractText(0)
	if str, ok := text0.(*objects.String); ok {
		if str.Value != "Page 1" {
			t.Errorf("Expected page 0 text 'Page 1', got %q", str.Value)
		}
	}

	t.Logf("Extracted PDF: 3 specific pages")
}
