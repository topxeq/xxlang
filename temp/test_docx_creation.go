// temp/test_docx_creation.go
// Test program to create a DOCX file and verify its structure.
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/topxeq/xxlang/pkg/objects"
)

func main() {
	// Create a new document
	doc := objects.NewDocxDocument()

	// Get body
	body := doc.GetBody()
	if body == nil {
		fmt.Println("ERROR: Body is nil")
		return
	}

	// Add a paragraph with text
	para := objects.NewXMLNode("w:p")
	run := objects.NewXMLNode("w:r")
	text := objects.NewXMLNode("w:t")
	text.SetText("Hello from Xxlang DOCX module!")
	text.SetAttr("xml:space", "preserve")
	run.AddChild(text)
	para.AddChild(run)
	body.AddChild(para)

	// Add a second paragraph
	para2 := objects.NewXMLNode("w:p")
	run2 := objects.NewXMLNode("w:r")
	text2 := objects.NewXMLNode("w:t")
	text2.SetText("This is a test document created by the Xxlang DOCX module.")
	text2.SetAttr("xml:space", "preserve")
	run2.AddChild(text2)
	para2.AddChild(run2)
	body.AddChild(para2)

	// Set document properties
	props := doc.GetProperties()
	if props != nil {
		props.Title = "Test Document"
		props.Author = "Xxlang"
	}

	// Save to temp directory
	outputPath := filepath.Join("temp", "test_docx_output.docx")
	if err := os.MkdirAll("temp", 0755); err != nil {
		fmt.Printf("ERROR: Failed to create temp directory: %v\n", err)
		return
	}

	if err := doc.Save(outputPath); err != nil {
		fmt.Printf("ERROR: Failed to save document: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Document saved to %s\n", outputPath)

	// Verify by reading back
	openedDoc, err := objects.OpenDocx(outputPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to open saved document: %v\n", err)
		return
	}
	defer openedDoc.Close()

	// Get text content
	textNodes := openedDoc.FindElements("//w:t")
	var content string
	for _, elem := range textNodes.Elements {
		if node, ok := elem.(*objects.XMLNode); ok {
			content += node.Text()
		}
	}

	fmt.Printf("Document content: %q\n", content)
	if content == "Hello from Xxlang DOCX module!This is a test document created by the Xxlang DOCX module." {
		fmt.Println("SUCCESS: Content verified correctly!")
	} else {
		fmt.Println("WARNING: Content mismatch!")
	}

	// Check file size
	info, err := os.Stat(outputPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to get file info: %v\n", err)
		return
	}
	fmt.Printf("File size: %d bytes\n", info.Size())

	// List ZIP contents
	fmt.Println("\nVerifying ZIP structure...")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to read file: %v\n", err)
		return
	}

	// Parse as ZIP
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		fmt.Printf("ERROR: Failed to parse ZIP: %v\n", err)
		return
	}

	fmt.Println("ZIP contents:")
	for _, file := range reader.File {
		fmt.Printf("  - %s\n", file.Name)
	}
}