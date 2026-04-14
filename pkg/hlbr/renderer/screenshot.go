package renderer

import (
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

func ScreenshotText(doc *dom.Document, width int) string {
	if doc == nil || doc.Root == nil {
		return ""
	}

	r := NewTextRenderer(width)
	return r.Render(doc.Root)
}

func ScreenshotTextToFile(doc *dom.Document, path string, width int) error {
	text := ScreenshotText(doc, width)
	if text == "" {
		return fmt.Errorf("empty document")
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func ScreenshotTextToLines(doc *dom.Document, width int) []string {
	text := ScreenshotText(doc, width)
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func PrintText(doc *dom.Document, width int) {
	text := ScreenshotText(doc, width)
	fmt.Print(text)
}
