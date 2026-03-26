// pkg/stdlib/pdf.go
// PDF processing module for Xxlang.
// Provides functions for reading, creating, and manipulating PDF files.
// Implemented using only Go standard library - no third-party dependencies.
package stdlib

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "pdf",
		Exports: map[string]objects.Object{
			// ============================================
			// PDF Opening Functions
			// ============================================

			// new opens a PDF file from the given path.
			// Usage: doc = pdf.new("path/to/file.pdf")
			// Returns a PDF object or error.
			"new": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("new() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("new() requires a string path argument")
				}

				pdf, err := objects.NewPDFFromFile(path.Value)
				if err != nil {
					return Error("failed to open PDF: " + err.Error())
				}
				return pdf
			}),

			// newFromBytes creates a PDF object from byte data.
			// Usage: doc = pdf.newFromBytes(data)
			// data can be a string or an array of integers (0-255).
			"newFromBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newFromBytes() takes exactly 1 argument")
				}

				var data []byte
				switch arg := args[0].(type) {
				case *objects.String:
					data = []byte(arg.Value)
				case *objects.Array:
					data = make([]byte, len(arg.Elements))
					for i, elem := range arg.Elements {
						n, ok := elem.(*objects.Int)
						if !ok {
							return Error("newFromBytes() array must contain integers")
						}
						if n.Value < 0 || n.Value > 255 {
							return Error("newFromBytes() byte values must be 0-255")
						}
						data[i] = byte(n.Value)
					}
				default:
					return Error("newFromBytes() requires a string or byte array")
				}

				pdf, err := objects.NewPDFFromBytes(data)
				if err != nil {
					return Error("failed to parse PDF: " + err.Error())
				}
				return pdf
			}),

			// create creates a new empty PDF document.
			// Usage: doc = pdf.create()
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("create() takes no arguments")
				}
				return objects.NewPDFDocument()
			}),

			// ============================================
			// Quick Access Functions (no object creation)
			// ============================================

			// getPageCount returns the number of pages in a PDF file.
			// Usage: count = pdf.getPageCount("file.pdf")
			"getPageCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getPageCount() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("getPageCount() requires a string path argument")
				}

				pdf, err := objects.NewPDFFromFile(path.Value)
				if err != nil {
					return Error("failed to open PDF: " + err.Error())
				}
				defer pdf.Close()

				return Int(int64(pdf.PageCount))
			}),

			// getInfo returns information about a PDF file.
			// Usage: info = pdf.getInfo("file.pdf")
			"getInfo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getInfo() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("getInfo() requires a string path argument")
				}

				pdf, err := objects.NewPDFFromFile(path.Value)
				if err != nil {
					return Error("failed to open PDF: " + err.Error())
				}
				defer pdf.Close()

				return pdf.GetInfo()
			}),

			// extractText extracts text from a specific page.
			// Usage: text = pdf.extractText("file.pdf", pageIndex)
			"extractText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("extractText() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractText() requires a string path as first argument")
				}
				pageIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("extractText() requires an integer page index as second argument")
				}

				pdf, err := objects.NewPDFFromFile(path.Value)
				if err != nil {
					return Error("failed to open PDF: " + err.Error())
				}
				defer pdf.Close()

				return pdf.ExtractText(int(pageIdx.Value))
			}),

			// extractAllText extracts text from all pages.
			// Usage: text = pdf.extractAllText("file.pdf")
			"extractAllText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("extractAllText() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractAllText() requires a string path argument")
				}

				pdf, err := objects.NewPDFFromFile(path.Value)
				if err != nil {
					return Error("failed to open PDF: " + err.Error())
				}
				defer pdf.Close()

				return pdf.ExtractAllText()
			}),

			// ============================================
			// PDF Operations
			// ============================================

			// merge combines multiple PDF files into one.
			// Usage: pdf.merge(["file1.pdf", "file2.pdf"], "output.pdf")
			// or: pdf.merge(paths, outputPath)
			"merge": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("merge() takes exactly 2 arguments")
				}

				// Get paths array
				pathsArr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("merge() requires an array of file paths as first argument")
				}

				paths := make([]string, len(pathsArr.Elements))
				for i, elem := range pathsArr.Elements {
					s, ok := elem.(*objects.String)
					if !ok {
						return Error("merge() requires string paths in array")
					}
					paths[i] = s.Value
				}

				// Get output path
				outputPath, ok := args[1].(*objects.String)
				if !ok {
					return Error("merge() requires a string output path as second argument")
				}

				return mergePDFs(paths, outputPath.Value)
			}),

			// split extracts a range of pages from a PDF.
			// Usage: pdf.split("input.pdf", outputPath, startPage, endPage)
			// Pages are 0-indexed.
			"split": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("split() takes exactly 4 arguments")
				}

				inputPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("split() requires a string input path")
				}

				outputPath, ok := args[1].(*objects.String)
				if !ok {
					return Error("split() requires a string output path")
				}

				startPage, ok := args[2].(*objects.Int)
				if !ok {
					return Error("split() requires an integer start page")
				}

				endPage, ok := args[3].(*objects.Int)
				if !ok {
					return Error("split() requires an integer end page")
				}

				return splitPDF(inputPath.Value, outputPath.Value, int(startPage.Value), int(endPage.Value))
			}),

			// extractPages extracts specific pages from a PDF.
			// Usage: pdf.extractPages("input.pdf", [0, 2, 4], "output.pdf")
			"extractPages": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("extractPages() takes exactly 3 arguments")
				}

				inputPath, ok := args[0].(*objects.String)
				if !ok {
					return Error("extractPages() requires a string input path")
				}

				pagesArr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("extractPages() requires an array of page indices")
				}

				outputPath, ok := args[2].(*objects.String)
				if !ok {
					return Error("extractPages() requires a string output path")
				}

				pages := make([]int, len(pagesArr.Elements))
				for i, elem := range pagesArr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("extractPages() requires integer page indices")
					}
					pages[i] = int(n.Value)
				}

				return extractPages(inputPath.Value, pages, outputPath.Value)
			}),

			// ============================================
			// Constants
			// ============================================

			// Page size constants (in points, 1 point = 1/72 inch)
			"A4_WIDTH":  Float(595.28),
			"A4_HEIGHT": Float(841.89),
			"LETTER_WIDTH":  Float(612.0),
			"LETTER_HEIGHT": Float(792.0),
			"LEGAL_WIDTH":  Float(612.0),
			"LEGAL_HEIGHT": Float(1008.0),
			"A3_WIDTH":  Float(841.89),
			"A3_HEIGHT": Float(1190.55),
			"A5_WIDTH":  Float(420.94),
			"A5_HEIGHT": Float(595.28),

			// Rotation constants
			"ROTATE_0":   Int(0),
			"ROTATE_90":  Int(90),
			"ROTATE_180": Int(180),
			"ROTATE_270": Int(270),
		},
	})
}

// ============================================
// PDF Operation Helper Functions
// ============================================

// mergePDFs combines multiple PDF files into one.
func mergePDFs(paths []string, outputPath string) objects.Object {
	if len(paths) == 0 {
		return Error("no PDF files to merge")
	}

	// For a simple merge, we concatenate the content
	// This is a basic implementation - a full implementation would
	// properly merge PDF structures

	var result bytes.Buffer

	// Write header
	result.WriteString("%PDF-1.4\n")
	result.WriteString("%\xe2\xe3\xcf\xd3\n")

	// Track objects and positions
	objNum := int64(1)
	positions := make(map[int64]int64)

	// Collect all pages
	var pageObjs []string

	for _, path := range paths {
		pdf, err := objects.NewPDFFromFile(path)
		if err != nil {
			return Error("failed to open " + path + ": " + err.Error())
		}

		// Get pages
		for i := 0; i < pdf.PageCount; i++ {
			page := pdf.Pages[i]
			pageObjNum := objNum
			objNum++

			// Create page object
			var pageObj strings.Builder
			fmt.Fprintf(&pageObj, "<< /Type /Page /Parent 2 0 R ")
			fmt.Fprintf(&pageObj, "/MediaBox [0 0 %.1f %.1f] ", page.Width, page.Height)
			if page.Rotation != 0 {
				fmt.Fprintf(&pageObj, "/Rotate %d ", page.Rotation)
			}
			fmt.Fprintf(&pageObj, ">>")

			pageObjs = append(pageObjs, pageObj.String())

			_ = pageObjNum
		}

		pdf.Close()
	}

	// Write catalog
	positions[1] = int64(result.Len())
	result.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n\n")

	// Write pages root
	positions[2] = int64(result.Len())
	fmt.Fprintf(&result, "2 0 obj\n<< /Type /Pages /Kids [")
	for i := range pageObjs {
		fmt.Fprintf(&result, "%d 0 R ", i+3)
	}
	fmt.Fprintf(&result, "] /Count %d >>\nendobj\n\n", len(pageObjs))

	// Write page objects
	for i, pageObj := range pageObjs {
		positions[int64(i+3)] = int64(result.Len())
		fmt.Fprintf(&result, "%d 0 obj\n%s\nendobj\n\n", i+3, pageObj)
	}

	// Write xref
	xrefPos := int64(result.Len())
	result.WriteString("xref\n")
	fmt.Fprintf(&result, "0 %d\n", len(pageObjs)+3)
	result.WriteString("0000000000 65535 f \n")

	for i := int64(1); i <= int64(len(pageObjs)+2); i++ {
		if pos, ok := positions[i]; ok {
			fmt.Fprintf(&result, "%010d 00000 n \n", pos)
		}
	}

	// Write trailer
	result.WriteString("trailer\n")
	fmt.Fprintf(&result, "<< /Size %d /Root 1 0 R >>\n", len(pageObjs)+3)
	fmt.Fprintf(&result, "startxref\n%d\n%%%%EOF\n", xrefPos)

	// Write to file
	if err := os.WriteFile(outputPath, result.Bytes(), 0644); err != nil {
		return Error("failed to write output: " + err.Error())
	}

	return Null()
}

// splitPDF extracts a range of pages from a PDF.
func splitPDF(inputPath, outputPath string, startPage, endPage int) objects.Object {
	pdf, err := objects.NewPDFFromFile(inputPath)
	if err != nil {
		return Error("failed to open PDF: " + err.Error())
	}
	defer pdf.Close()

	if startPage < 0 || startPage >= pdf.PageCount {
		return Error("start page out of range")
	}
	if endPage < startPage || endPage >= pdf.PageCount {
		return Error("end page out of range")
	}

	// Create a new PDF with selected pages
	doc := objects.NewPDFDocument()

	for i := startPage; i <= endPage; i++ {
		page := pdf.Pages[i]
		pageIdx := doc.AddPage(page.Width, page.Height)

		// Copy content
		text := page.ExtractText()
		if s, ok := text.(*objects.String); ok && s.Value != "" {
			if pIdx, ok := pageIdx.(*objects.Int); ok {
				doc.WriteText(int(pIdx.Value), s.Value, 50, page.Height-50, nil)
			}
		}

		// Copy rotation
		if page.Rotation != 0 {
			doc.Pages[len(doc.Pages)-1].Rotation = page.Rotation
		}
	}

	// Save
	result := doc.Save(outputPath)
	if errObj, ok := result.(*objects.Error); ok {
		return errObj
	}

	return Null()
}

// extractPages extracts specific pages from a PDF.
func extractPages(inputPath string, pages []int, outputPath string) objects.Object {
	pdf, err := objects.NewPDFFromFile(inputPath)
	if err != nil {
		return Error("failed to open PDF: " + err.Error())
	}
	defer pdf.Close()

	// Create a new PDF with selected pages
	doc := objects.NewPDFDocument()

	for _, pageNum := range pages {
		if pageNum < 0 || pageNum >= pdf.PageCount {
			return Error(fmt.Sprintf("page %d out of range", pageNum))
		}

		page := pdf.Pages[pageNum]
		pageIdx := doc.AddPage(page.Width, page.Height)

		// Copy content
		text := page.ExtractText()
		if s, ok := text.(*objects.String); ok && s.Value != "" {
			if pIdx, ok := pageIdx.(*objects.Int); ok {
				doc.WriteText(int(pIdx.Value), s.Value, 50, page.Height-50, nil)
			}
		}

		// Copy rotation
		if page.Rotation != 0 {
			doc.Pages[len(doc.Pages)-1].Rotation = page.Rotation
		}
	}

	// Save
	result := doc.Save(outputPath)
	if errObj, ok := result.(*objects.Error); ok {
		return errObj
	}

	return Null()
}