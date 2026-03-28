// pkg/stdlib/docx_test.go
// Tests for the DOCX module.
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestDocxCreate tests creating a new document
func TestDocxCreate(t *testing.T) {
	// Get create function
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn, ok := mod.Exports["create"].(*objects.Builtin)
	if !ok {
		t.Fatal("create function not found")
	}

	// Call create
	result := createFn.Fn()
	doc, ok := result.(*objects.DocxDocument)
	if !ok {
		t.Fatalf("expected DocxDocument, got %T", result)
	}

	// Verify document has body
	body := doc.GetBody()
	if body == nil {
		t.Fatal("document body is nil")
	}

	// Clean up
	doc.Close()
}

// TestDocxAddParagraph tests adding a paragraph
func TestDocxAddParagraph(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	addParaFn := mod.Exports["addParagraph"].(*objects.Builtin)

	// Create document
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)
	defer doc.Close()

	// Add paragraph with text
	paraResult := addParaFn.Fn(doc, &objects.String{Value: "Hello, World!"})
	para, ok := paraResult.(*objects.DocxParagraph)
	if !ok {
		t.Fatalf("expected DocxParagraph, got %T", paraResult)
	}

	if para.XmlNode == nil {
		t.Fatal("paragraph XML node is nil")
	}
}

// TestDocxAddTable tests adding a table
func TestDocxAddTable(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	addTableFn := mod.Exports["addTable"].(*objects.Builtin)

	// Create document
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)
	defer doc.Close()

	// Add table
	tableResult := addTableFn.Fn(doc, &objects.Int{Value: 3}, &objects.Int{Value: 4})
	table, ok := tableResult.(*objects.DocxTable)
	if !ok {
		t.Fatalf("expected DocxTable, got %T", tableResult)
	}

	if table.Rows != 3 {
		t.Errorf("expected 3 rows, got %d", table.Rows)
	}
	if table.Cols != 4 {
		t.Errorf("expected 4 cols, got %d", table.Cols)
	}
}

// TestDocxSaveAndOpen tests saving and opening a document
func TestDocxSaveAndOpen(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	addParaFn := mod.Exports["addParagraph"].(*objects.Builtin)
	saveFn := mod.Exports["save"].(*objects.Builtin)
	openFn := mod.Exports["open"].(*objects.Builtin)
	getTextFn := mod.Exports["getText"].(*objects.Builtin)
	closeFn := mod.Exports["close"].(*objects.Builtin)

	// Create temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_docx.docx")
	defer os.Remove(tmpFile)

	// Create document with text
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)

	testText := "Test paragraph content"
	addParaFn.Fn(doc, &objects.String{Value: testText})

	// Save
	saveResult := saveFn.Fn(doc, &objects.String{Value: tmpFile})
	if _, ok := saveResult.(*objects.Error); ok {
		t.Fatalf("save failed: %v", saveResult)
	}
	doc.Close()

	// Open the saved file
	openResult := openFn.Fn(&objects.String{Value: tmpFile})
	openedDoc, ok := openResult.(*objects.DocxDocument)
	if !ok {
		t.Fatalf("expected DocxDocument, got %T", openResult)
	}
	defer closeFn.Fn(openedDoc)

	// Get text
	textResult := getTextFn.Fn(openedDoc)
	text, ok := textResult.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", textResult)
	}

	if text.Value != testText {
		t.Errorf("expected text %q, got %q", testText, text.Value)
	}
}

// TestDocxSetCellText tests setting table cell text
func TestDocxSetCellText(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	addTableFn := mod.Exports["addTable"].(*objects.Builtin)
	setCellTextFn := mod.Exports["setCellText"].(*objects.Builtin)
	getTextFn := mod.Exports["getText"].(*objects.Builtin)

	// Create document with table
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)
	defer doc.Close()

	tableResult := addTableFn.Fn(doc, &objects.Int{Value: 2}, &objects.Int{Value: 2})
	table := tableResult.(*objects.DocxTable)

	// Set cell text
	setCellTextFn.Fn(table, &objects.Int{Value: 0}, &objects.Int{Value: 0}, &objects.String{Value: "Cell 0,0"})
	setCellTextFn.Fn(table, &objects.Int{Value: 0}, &objects.Int{Value: 1}, &objects.String{Value: "Cell 0,1"})
	setCellTextFn.Fn(table, &objects.Int{Value: 1}, &objects.Int{Value: 0}, &objects.String{Value: "Cell 1,0"})
	setCellTextFn.Fn(table, &objects.Int{Value: 1}, &objects.Int{Value: 1}, &objects.String{Value: "Cell 1,1"})

	// Get text from document
	textResult := getTextFn.Fn(doc)
	text := textResult.(*objects.String)

	expectedText := "Cell 0,0Cell 0,1Cell 1,0Cell 1,1"
	if text.Value != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, text.Value)
	}
}

// TestDocxProperties tests document properties
func TestDocxProperties(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	setTitleFn := mod.Exports["setTitle"].(*objects.Builtin)
	getTitleFn := mod.Exports["getTitle"].(*objects.Builtin)
	setAuthorFn := mod.Exports["setAuthor"].(*objects.Builtin)
	getAuthorFn := mod.Exports["getAuthor"].(*objects.Builtin)

	// Create document
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)
	defer doc.Close()

	// Set and get title
	testTitle := "Test Document Title"
	setTitleFn.Fn(doc, &objects.String{Value: testTitle})
	titleResult := getTitleFn.Fn(doc)
	title := titleResult.(*objects.String)
	if title.Value != testTitle {
		t.Errorf("expected title %q, got %q", testTitle, title.Value)
	}

	// Set and get author
	testAuthor := "Test Author"
	setAuthorFn.Fn(doc, &objects.String{Value: testAuthor})
	authorResult := getAuthorFn.Fn(doc)
	author := authorResult.(*objects.String)
	if author.Value != testAuthor {
		t.Errorf("expected author %q, got %q", testAuthor, author.Value)
	}
}

// TestDocxPageBreak tests adding page breaks
func TestDocxPageBreak(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	addParaFn := mod.Exports["addParagraph"].(*objects.Builtin)
	addPageBreakFn := mod.Exports["addPageBreak"].(*objects.Builtin)

	// Create document
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)
	defer doc.Close()

	// Add paragraph
	addParaFn.Fn(doc, &objects.String{Value: "Page 1"})

	// Add page break
	breakResult := addPageBreakFn.Fn(doc)
	if _, ok := breakResult.(*objects.Error); ok {
		t.Fatalf("addPageBreak failed: %v", breakResult)
	}

	// Add another paragraph
	addParaFn.Fn(doc, &objects.String{Value: "Page 2"})

	// Verify the document can be saved
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_pagebreak.docx")
	defer os.Remove(tmpFile)

	saveFn := mod.Exports["save"].(*objects.Builtin)
	saveResult := saveFn.Fn(doc, &objects.String{Value: tmpFile})
	if _, ok := saveResult.(*objects.Error); ok {
		t.Fatalf("save failed: %v", saveResult)
	}
}

// TestDocxTypeChecks tests type check functions
func TestDocxTypeChecks(t *testing.T) {
	mod := Get("docx")
	if mod == nil {
		t.Fatal("docx module not found")
	}

	createFn := mod.Exports["create"].(*objects.Builtin)
	isDocxDocFn := mod.Exports["isDocxDocument"].(*objects.Builtin)
	isDocxParaFn := mod.Exports["isDocxParagraph"].(*objects.Builtin)

	// Test isDocxDocument
	result := createFn.Fn()
	doc := result.(*objects.DocxDocument)

	checkResult := isDocxDocFn.Fn(doc)
	check := checkResult.(*objects.Bool)
	if !check.Value {
		t.Error("isDocxDocument should return true for DocxDocument")
	}

	// Test with wrong type
	checkResult = isDocxDocFn.Fn(&objects.String{Value: "test"})
	check = checkResult.(*objects.Bool)
	if check.Value {
		t.Error("isDocxDocument should return false for String")
	}

	// Test isDocxParagraph
	addParaFn := mod.Exports["addParagraph"].(*objects.Builtin)
	paraResult := addParaFn.Fn(doc, &objects.String{Value: "test"})
	para := paraResult.(*objects.DocxParagraph)

	checkResult = isDocxParaFn.Fn(para)
	check = checkResult.(*objects.Bool)
	if !check.Value {
		t.Error("isDocxParagraph should return true for DocxParagraph")
	}
}