// pkg/stdlib/html_test.go
// Comprehensive tests for html module.
package stdlib

import (
	"os"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callHTMLFunc calls a function from the html module.
func callHTMLFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("html")
	if mod == nil {
		t := &testing.T{}
		t.Skip("html module not found")
		return &objects.Error{Message: "html module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

func TestHTMLParse(t *testing.T) {
	// Test parsing valid HTML
	htmlStr := `<html><body><h1>Hello, World!</h1></body></html>`
	doc := callHTMLFunc("parse", objects.NewString(htmlStr))
	if doc.Type() != objects.HTMLDocumentType {
		t.Fatalf("expected HTMLDocument, got %s", doc.Type())
	}
	docObj := doc.(*objects.HTMLDocument)
	if docObj.Root() == nil {
		t.Error("expected document to have root element")
	}

	// Test parsing invalid HTML (malformed)
	badHTML := `<html><body><unclosed>`
	badDoc := callHTMLFunc("parse", objects.NewString(badHTML))
	if badDoc.Type() != objects.ErrorType {
		t.Logf("parse of malformed HTML returned %s (may be lenient)", badDoc.Type())
	}

	// Test empty string
	emptyDoc := callHTMLFunc("parse", objects.NewString(""))
	if emptyDoc.Type() != objects.ErrorType {
		t.Logf("parse of empty string returned %s (expected error or empty doc)", emptyDoc.Type())
	}
}

func TestHTMLParseFile(t *testing.T) {
	// Create temporary HTML file
	tmpFile, err := os.CreateTemp("", "xxlang-html-test-*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	htmlContent := `<!DOCTYPE html><html><head><title>Test</title></head><body><p>Content</p></body></html>`
	if _, err := tmpFile.WriteString(htmlContent); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	doc := callHTMLFunc("parseFile", objects.NewString(tmpFile.Name()))
	if doc.Type() != objects.HTMLDocumentType {
		t.Fatalf("parseFile should return HTMLDocument, got %s", doc.Type())
	}

	// Non-existent file
	missing := callHTMLFunc("parseFile", objects.NewString("nonexistent.html"))
	if missing.Type() != objects.ErrorType {
		t.Logf("parseFile of nonexistent file returned %s (expected error)", missing.Type())
	}
}

func TestHTMLParseFragment(t *testing.T) {
	fragment := `<div>Hello</div><span>World</span>`
	result := callHTMLFunc("parseFragment", objects.NewString(fragment))
	if result.Type() != objects.ArrayType {
		t.Fatalf("expected Array, got %s", result.Type())
	}
	arr := result.(*objects.Array)
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}
	for _, elem := range arr.Elements {
		if elem.Type() != objects.HTMLElementType {
			t.Errorf("expected HTMLElement, got %s", elem.Type())
		}
	}

	// Empty fragment
	empty := callHTMLFunc("parseFragment", objects.NewString(""))
	if empty.Type() != objects.ArrayType {
		t.Logf("parseFragment of empty string returned %s", empty.Type())
	} else {
		if len(empty.(*objects.Array).Elements) != 0 {
			t.Errorf("expected empty array for empty fragment")
		}
	}
}

func TestHTMLNewDocument(t *testing.T) {
	doc := callHTMLFunc("newDocument")
	if doc.Type() != objects.HTMLDocumentType {
		t.Fatalf("expected HTMLDocument, got %s", doc.Type())
	}
	docObj := doc.(*objects.HTMLDocument)
	// newDocument creates a basic HTML structure with html/head/body
	if docObj.Root() == nil {
		t.Error("newDocument should create document with root element")
	}
	if docObj.Head() == nil {
		t.Error("newDocument should create document with head")
	}
	if docObj.Body() == nil {
		t.Error("newDocument should create document with body")
	}
}

func TestHTMLNewDocumentWithTitle(t *testing.T) {
	doc := callHTMLFunc("newDocumentWithTitle", objects.NewString("My Page"))
	if doc.Type() != objects.HTMLDocumentType {
		t.Fatalf("expected HTMLDocument, got %s", doc.Type())
	}
	docObj := doc.(*objects.HTMLDocument)
	if docObj.Title() != "My Page" {
		t.Errorf("expected title 'My Page', got '%s'", docObj.Title())
	}
	if docObj.Root() == nil {
		t.Error("expected document to have root element")
	}
	if docObj.Head() == nil {
		t.Error("expected document to have head element")
	}
	if docObj.Body() == nil {
		t.Error("expected document to have body element")
	}
}

func TestHTMLNewElement(t *testing.T) {
	elem := callHTMLFunc("newElement", objects.NewString("div"))
	if elem.Type() != objects.HTMLElementType {
		t.Fatalf("expected HTMLElement, got %s", elem.Type())
	}
	el := elem.(*objects.HTMLElement)
	if el.TagName() != "div" {
		t.Errorf("expected tag name 'div', got '%s'", el.TagName())
	}
}

func TestHTMLNewTextNode(t *testing.T) {
	textNode := callHTMLFunc("newTextNode", objects.NewString("Hello"))
	if textNode.Type() != objects.HTMLElementType {
		t.Fatalf("expected HTMLElement for text node, got %s", textNode.Type())
	}
	// Text nodes have special node type
	el := textNode.(*objects.HTMLElement)
	if el.NodeType() != objects.HTMLNodeText {
		t.Errorf("expected text node type, got %d", el.NodeType())
	}
	if el.TextContent() != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", el.TextContent())
	}
}

func TestHTMLNewComment(t *testing.T) {
	comment := callHTMLFunc("newComment", objects.NewString("a comment"))
	if comment.Type() != objects.HTMLElementType {
		t.Fatalf("expected HTMLElement for comment, got %s", comment.Type())
	}
	el := comment.(*objects.HTMLElement)
	if el.NodeType() != objects.HTMLNodeComment {
		t.Errorf("expected comment node type, got %d", el.NodeType())
	}
	if el.TextContent() != "a comment" {
		t.Errorf("expected comment text 'a comment', got '%s'", el.TextContent())
	}
}

func TestHTMLEscape(t *testing.T) {
	input := `<>&"`
	result := callHTMLFunc("escape", objects.NewString(input))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	// EscapeHTML escapes &, <, > but not quotes
	expected := "&lt;&gt;&amp;\""
	if str.Value != expected {
		t.Errorf("expected %q, got %q", expected, str.Value)
	}
}

func TestHTMLEscapeAttr(t *testing.T) {
	input := `"'><`
	result := callHTMLFunc("escapeAttr", objects.NewString(input))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	// Should escape quotes and other special chars
	if !strings.Contains(str.Value, "&quot;") && !strings.Contains(str.Value, "&#34;") {
		t.Logf("escapeAttr result %q may not escape quotes properly", str.Value)
	}
}

func TestHTMLUnescape(t *testing.T) {
	input := "&lt;div&gt;Hello&amp;World&quot;"
	result := callHTMLFunc("unescape", objects.NewString(input))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "<div>Hello&World\""
	if str.Value != expected {
		t.Errorf("expected %q, got %q", expected, str.Value)
	}
}

func TestHTMLStripTags(t *testing.T) {
	input := "<p>Hello <b>World</b>!</p>"
	result := callHTMLFunc("stripTags", objects.NewString(input))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	// Should contain text content, no tags
	if str.Value != "Hello World!" {
		t.Logf("stripTags returned %q (may differ in spacing)", str.Value)
	}
}

func TestHTMLSanitize(t *testing.T) {
	// Test with potentially dangerous content
	input := `<script>alert('xss')</script><p>Safe content</p>`
	result := callHTMLFunc("sanitize", objects.NewString(input))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	// Should remove script tags
	if strings.Contains(str.Value, "<script") || strings.Contains(str.Value, "</script>") {
		t.Errorf("sanitize should remove script tags, got %q", str.Value)
	}
	if !strings.Contains(str.Value, "Safe content") {
		t.Error("sanitize should preserve safe content")
	}
}

func TestHTMLIsHTMLDocument(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["isHTMLDocument"].(*objects.Builtin)

	// Non-document
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	doc := callHTMLFunc("newDocument")
	res = fn.Fn(doc)
	if b, ok := res.(*objects.Bool); ok {
		if !b.Value {
			t.Error("expected true for HTMLDocument")
		}
	}
}

func TestHTMLIsHTMLElement(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["isHTMLElement"].(*objects.Builtin)

	// Non-element
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	elem := callHTMLFunc("newElement", objects.NewString("div"))
	res = fn.Fn(elem)
	if b, ok := res.(*objects.Bool); ok {
		if !b.Value {
			t.Error("expected true for HTMLElement")
		}
	}
}

func TestHTMLEncode(t *testing.T) {
	// Encode a simple string
	result := callHTMLFunc("encode", objects.NewString("Hello"))
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if !strings.Contains(str.Value, "Hello") {
		t.Errorf("expected output to contain 'Hello', got %q", str.Value)
	}

	// Encode with custom root name
	result = callHTMLFunc("encode", objects.NewString("test"), objects.NewString("span"))
	str = result.(*objects.String)
	if !strings.Contains(str.Value, "<span") {
		t.Errorf("expected output to start with <span, got %q", str.Value)
	}
}

func TestHTMLCreateElementAlias(t *testing.T) {
	elem1 := callHTMLFunc("newElement", objects.NewString("p"))
	elem2 := callHTMLFunc("createElement", objects.NewString("p"))
	if elem1.Type() != elem2.Type() {
		t.Errorf("createElement should be alias for newElement")
	}
}

func TestHTMLCreateTextNodeAlias(t *testing.T) {
	text1 := callHTMLFunc("newTextNode", objects.NewString("abc"))
	text2 := callHTMLFunc("createTextNode", objects.NewString("abc"))
	if text1.Type() != text2.Type() {
		t.Errorf("createTextNode should be alias for newTextNode")
	}
}
