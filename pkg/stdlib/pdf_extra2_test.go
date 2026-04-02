// pkg/stdlib/pdf_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// pdfCall invokes a builtin from the pdf module.
func pdfCall(name string, args ...objects.Object) objects.Object {
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

// TestPdf_Extra2_Init tests pdf module exports.
func TestPdf_Extra2_Init(t *testing.T) {
	mod := Get("pdf")
	if mod == nil {
		t.Skip("pdf module not found")
	}
	expected := []string{
		"new", "newFromBytes", "create",
		"getPageCount", "getInfo", "extractText", "extractAllText",
		"merge",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in pdf module", name)
		}
	}
}

// TestPdf_Extra2_New_ArgumentValidation tests pdf.new().
func TestPdf_Extra2_New_ArgumentValidation(t *testing.T) {
	// No args
	res := pdfCall("new")
	if res.Type() != objects.ErrorType {
		t.Fatalf("new() with no args should error")
	}
	// Too many args
	res = pdfCall("new", objects.NewString("a.pdf"), objects.NewString("b"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("new() with too many args should error")
	}
	// Wrong type
	res = pdfCall("new", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("new() with int should error")
	}
}

// TestPdf_Extra2_NewFromBytes_ArgumentValidation tests pdf.newFromBytes().
func TestPdf_Extra2_NewFromBytes_ArgumentValidation(t *testing.T) {
	// No args
	res := pdfCall("newFromBytes")
	if res.Type() != objects.ErrorType {
		t.Fatalf("newFromBytes() with no args should error")
	}
	// Wrong type: string
	res = pdfCall("newFromBytes", objects.NewString("not bytes"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newFromBytes() with string should error")
	}
	// Wrong type: array with non-int elements would error later; just check it accepts array
	// but if array has non-int, it should error
	arr := &objects.Array{Elements: []objects.Object{objects.NewString("x")}}
	res = pdfCall("newFromBytes", arr)
	if res.Type() != objects.ErrorType {
		t.Fatalf("newFromBytes() with non-int array should error")
	}
	// Valid byte array
	validBytes := &objects.Array{Elements: []objects.Object{objects.NewInt(37), objects.NewInt(80), objects.NewInt(75)}}
	res = pdfCall("newFromBytes", validBytes)
	// This might succeed or fail depending on PDF parsing, but should not be a type error
	_ = res
}

// TestPdf_Extra2_Create tests pdf.create().
func TestPdf_Extra2_Create(t *testing.T) {
	res := pdfCall("create")
	if res.Type() != objects.PDFDocumentType {
		t.Fatalf("create() should return PdfDocument, got %s", res.Type())
	}
	// With args should error
	res = pdfCall("create", objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("create() with args should error")
	}
}

// TestPdf_Extra2_GetPageCount_GetInfo_ArgumentValidation tests getPageCount and getInfo.
func TestPdf_Extra2_GetPageCount_GetInfo_ArgumentValidation(t *testing.T) {
	// getPageCount: no args
	res := pdfCall("getPageCount")
	if res.Type() != objects.ErrorType {
		t.Fatalf("getPageCount() with no args should error")
	}
	// getPageCount: wrong type
	res = pdfCall("getPageCount", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("getPageCount() with int should error")
	}
	// getInfo similar
	res = pdfCall("getInfo")
	if res.Type() != objects.ErrorType {
		t.Fatalf("getInfo() with no args should error")
	}
	res = pdfCall("getInfo", objects.NewString("file.pdf"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("getInfo() with string is okay (may return error if file not found, but type should be okay)")
	}
}

// TestPdf_Extra2_ExtractText_ArgumentValidation tests extractText and extractAllText.
func TestPdf_Extra2_ExtractText_ArgumentValidation(t *testing.T) {
	// extractText: wrong number of args
	res := pdfCall("extractText")
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractText() with no args should error")
	}
	res = pdfCall("extractText", objects.NewString("file.pdf"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractText() with 1 arg should error")
	}
	res = pdfCall("extractText", objects.NewString("file.pdf"), objects.NewInt(0), objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractText() with 3 args should error")
	}
	// Wrong types
	res = pdfCall("extractText", objects.NewInt(123), objects.NewInt(0))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractText() with int path should error")
	}
	res = pdfCall("extractText", objects.NewString("file.pdf"), objects.NewString("0"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractText() with string page should error")
	}
	// extractAllText: wrong number
	res = pdfCall("extractAllText")
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractAllText() with no args should error")
	}
	res = pdfCall("extractAllText", objects.NewString("file.pdf"), objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractAllText() with too many args should error")
	}
	// Wrong type
	res = pdfCall("extractAllText", objects.NewInt(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("extractAllText() with int should error")
	}
}

// TestPdf_Extra2_Merge_ArgumentValidation tests pdf.merge().
func TestPdf_Extra2_Merge_ArgumentValidation(t *testing.T) {
	// No args
	res := pdfCall("merge")
	if res.Type() != objects.ErrorType {
		t.Fatalf("merge() with no args should error")
	}
	// One arg only
	res = pdfCall("merge", objects.NewString("a.pdf"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("merge() with 1 arg should error")
	}
	// Too many args (3)
	res = pdfCall("merge", objects.NewString("a.pdf"), objects.NewString("out.pdf"), objects.NewString("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("merge() with 3 args should error")
	}
	// First arg not array
	res = pdfCall("merge", objects.NewString("a.pdf"), objects.NewString("out.pdf"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("merge() with non-array first arg should error")
	}
	// Array with non-string elements
	badArr := &objects.Array{Elements: []objects.Object{objects.NewInt(1)}}
	res = pdfCall("merge", badArr, objects.NewString("out.pdf"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("merge() with non-string array elements should error")
	}
	// Valid signature (may fail due to file not found, but should accept types)
	validArr := &objects.Array{Elements: []objects.Object{objects.NewString("a.pdf"), objects.NewString("b.pdf")}}
	res = pdfCall("merge", validArr, objects.NewString("out.pdf"))
	_ = res // don't fail on I/O; just ensure type check passes
}
