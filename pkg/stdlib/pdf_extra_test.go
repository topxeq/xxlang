// pkg/stdlib/pdf_extra_test.go
// Additional tests for pdf module to cover builtin functions: merge, split, extractPages
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callPdfFunc calls a function from the pdf module.
func callPdfFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("pdf")
	if mod == nil {
		return &objects.Error{Message: "pdf module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// createSimplePDF creates a one-page PDF with given text and returns the file path.
func createSimplePDF(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "simple.pdf")
	doc := objects.NewPDFDocument()
	doc.AddPage(595, 842)
	doc.WriteText(0, content, 100, 700, nil)
	res := doc.Save(path)
	if isPdfErr(res) {
		t.Fatalf("Failed to create PDF: %v", res)
	}
	return path
}

// createMultiPagePDF creates a multi-page PDF with page-specific text.
func createMultiPagePDF(t *testing.T, pageTexts []string) string {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.pdf")
	doc := objects.NewPDFDocument()
	for i, text := range pageTexts {
		doc.AddPage(595, 842)
		doc.WriteText(i, text, 100, 700, nil)
	}
	res := doc.Save(path)
	if isPdfErr(res) {
		t.Fatalf("Failed to create multi-page PDF: %v", res)
	}
	return path
}

// TestPdfMergeBuiltin tests the builtin "merge" function (wrapper around mergePDFs).
func TestPdfMergeBuiltin(t *testing.T) {
	pdf1 := createSimplePDF(t, "First PDF")
	pdf2 := createSimplePDF(t, "Second PDF")

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "merged.pdf")

	result := callPdfFunc("merge", Array(String(pdf1), String(pdf2)), String(outputPath))
	if isPdfErr(result) {
		t.Fatalf("merge() failed: %v", result.Inspect())
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("merged PDF was not created")
	}

	countResult := callPdfFunc("getPageCount", String(outputPath))
	if isPdfErr(countResult) {
		t.Fatalf("getPageCount failed: %v", countResult.Inspect())
	}
	if count, ok := countResult.(*objects.Int); ok {
		if count.Value != 2 {
			t.Errorf("expected 2 pages in merged PDF, got %d", count.Value)
		}
	} else {
		t.Fatalf("expected Int for page count, got %T", countResult)
	}
}

// TestPdfMergeBuiltin_ErrorTests tests merge error handling via builtin.
func TestPdfMergeBuiltin_ErrorTests(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "merged.pdf")

	result := callPdfFunc("merge", String("a.pdf"))
	if !isPdfErr(result) {
		t.Error("merge with 1 arg should error")
	}

	result = callPdfFunc("merge", String("not an array"), String(outputPath))
	if !isPdfErr(result) {
		t.Error("merge with non-array first arg should error")
	}

	result = callPdfFunc("merge", Array(Int(123)), String(outputPath))
	if !isPdfErr(result) {
		t.Error("merge with non-string array element should error")
	}

	result = callPdfFunc("merge", Array(String("a.pdf")), Int(456))
	if !isPdfErr(result) {
		t.Error("merge with non-string second arg should error")
	}
}

// TestPdfSplitBuiltin tests the builtin "split" function (wrapper around splitPDF).
func TestPdfSplitBuiltin(t *testing.T) {
	pdfPath := createMultiPagePDF(t, []string{"P0", "P1", "P2", "P3"})

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "split.pdf")

	result := callPdfFunc("split", String(pdfPath), String(outputPath), Int(0), Int(1))
	if isPdfErr(result) {
		t.Fatalf("split() failed: %v", result.Inspect())
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("split PDF was not created")
	}

	countResult := callPdfFunc("getPageCount", String(outputPath))
	if isPdfErr(countResult) {
		t.Fatalf("getPageCount failed: %v", countResult.Inspect())
	}
	if count, ok := countResult.(*objects.Int); ok {
		if count.Value != 2 {
			t.Errorf("expected 2 pages in split PDF, got %d", count.Value)
		}
	} else {
		t.Fatalf("expected Int for page count, got %T", countResult)
	}
}

// TestPdfSplitBuiltin_ErrorTests tests split error handling via builtin.
func TestPdfSplitBuiltin_ErrorTests(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "dummy.pdf")
	outputPath := filepath.Join(tmpDir, "out.pdf")
	os.WriteFile(inputPath, []byte("dummy"), 0644)

	result := callPdfFunc("split", String(inputPath), String(outputPath), Int(0))
	if !isPdfErr(result) {
		t.Error("split with 3 args should error")
	}

	result = callPdfFunc("split", Int(123), String(outputPath), Int(0), Int(1))
	if !isPdfErr(result) {
		t.Error("split with non-string input should error")
	}

	result = callPdfFunc("split", String(inputPath), Int(456), Int(0), Int(1))
	if !isPdfErr(result) {
		t.Error("split with non-string output should error")
	}

	result = callPdfFunc("split", String(inputPath), String(outputPath), String("zero"), Int(1))
	if !isPdfErr(result) {
		t.Error("split with non-int start page should error")
	}

	result = callPdfFunc("split", String(inputPath), String(outputPath), Int(0), String("one"))
	if !isPdfErr(result) {
		t.Error("split with non-int end page should error")
	}
}

// TestPdfExtractPagesBuiltin tests the builtin "extractPages" function (wrapper around extractPages).
func TestPdfExtractPagesBuiltin(t *testing.T) {
	pdfPath := createMultiPagePDF(t, []string{"P0", "P1", "P2", "P3", "P4"})

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "extracted.pdf")

	result := callPdfFunc("extractPages", String(pdfPath), Array(Int(0), Int(2), Int(4)), String(outputPath))
	if isPdfErr(result) {
		t.Fatalf("extractPages() failed: %v", result.Inspect())
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("extracted PDF was not created")
	}

	countResult := callPdfFunc("getPageCount", String(outputPath))
	if isPdfErr(countResult) {
		t.Fatalf("getPageCount failed: %v", countResult.Inspect())
	}
	if count, ok := countResult.(*objects.Int); ok {
		if count.Value != 3 {
			t.Errorf("expected 3 pages in extracted PDF, got %d", count.Value)
		}
	} else {
		t.Fatalf("expected Int for page count, got %T", countResult)
	}
}

// TestPdfExtractPagesBuiltin_ErrorTests tests extractPages error handling via builtin.
func TestPdfExtractPagesBuiltin_ErrorTests(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "dummy.pdf")
	outputPath := filepath.Join(tmpDir, "out.pdf")
	os.WriteFile(inputPath, []byte("dummy"), 0644)

	result := callPdfFunc("extractPages", String(inputPath), Array(Int(0)))
	if !isPdfErr(result) {
		t.Error("extractPages with 2 args should error")
	}

	result = callPdfFunc("extractPages", Int(123), Array(Int(0)), String(outputPath))
	if !isPdfErr(result) {
		t.Error("extractPages with non-string input should error")
	}

	result = callPdfFunc("extractPages", String(inputPath), String("not array"), String(outputPath))
	if !isPdfErr(result) {
		t.Error("extractPages with non-array pages should error")
	}

	result = callPdfFunc("extractPages", String(inputPath), Array(String("zero")), String(outputPath))
	if !isPdfErr(result) {
		t.Error("extractPages with non-int page index should error")
	}

	result = callPdfFunc("extractPages", String(inputPath), Array(Int(0)), Int(456))
	if !isPdfErr(result) {
		t.Error("extractPages with non-string output should error")
	}
}

// isPdfErr checks if an object is an error.
func isPdfErr(obj objects.Object) bool {
	_, ok := obj.(*objects.Error)
	return ok
}
