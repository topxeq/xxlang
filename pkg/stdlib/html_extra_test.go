// pkg/stdlib/html_extra_test.go
// Additional tests for html module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestHTMLParse_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["parse"].(*objects.Builtin)
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"", false},
		{"<div>Hello</div>", false},
		{"<p>Paragraph</p>", false},
		{"<html><body>Content</body></html>", false},
		{"just text", false},
		{"<br>", false},
		{"<a href='url'>Link</a>", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			if tt.wantErr {
				if result.Type() != objects.ErrorType {
					t.Errorf("parse(%q) expected error, got %s", tt.input, result.Inspect())
				}
			} else {
				if result.Type() == objects.ErrorType {
					t.Errorf("parse(%q) unexpected error: %s", tt.input, result.Inspect())
				}
			}
		})
	}
}

func TestHTMLEscape_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	escapeFn := mod.Exports["escape"].(*objects.Builtin)
	res1 := escapeFn.Fn(String("<>&\""))
	s1, ok := res1.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", res1)
	}
	expected1 := "&lt;&gt;&amp;\""
	if s1.Value != expected1 {
		t.Errorf("escape = %q, want %q", s1.Value, expected1)
	}
}

func TestHTMLCreate_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["newElement"].(*objects.Builtin)
	result := fn.Fn(String("div"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("newElement(div) error: %s", result.Inspect())
	}
	if result == objects.NULL {
		t.Error("newElement returned NULL")
	}
}

func TestHTMLNewDocument_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["newDocument"].(*objects.Builtin)
	doc := fn.Fn()
	if doc.Type() == objects.ErrorType {
		t.Fatalf("newDocument error: %s", doc.Inspect())
	}
	if doc == objects.NULL {
		t.Error("newDocument returned NULL")
	}
}

func TestHTMLIsHTMLDocument_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["isHTMLDocument"].(*objects.Builtin)
	createFn := mod.Exports["newElement"].(*objects.Builtin)
	newDocFn := mod.Exports["newDocument"].(*objects.Builtin)
	doc := newDocFn.Fn()
	resDoc := fn.Fn(doc)
	if resDoc != objects.TRUE {
		t.Errorf("isHTMLDocument(newDocument()) = %v, want TRUE", resDoc)
	}
	node := createFn.Fn(String("span"))
	resNode := fn.Fn(node)
	if resNode != objects.FALSE {
		t.Errorf("isHTMLDocument(newElement(span)) = %v, want FALSE", resNode)
	}
}

func TestHTMLIsHTMLElement_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["isHTMLElement"].(*objects.Builtin)
	createFn := mod.Exports["newElement"].(*objects.Builtin)
	newDocFn := mod.Exports["newDocument"].(*objects.Builtin)
	node := createFn.Fn(String("p"))
	resNode := fn.Fn(node)
	if resNode != objects.TRUE {
		t.Errorf("isHTMLElement(newElement(p)) = %v, want TRUE", resNode)
	}
	doc := newDocFn.Fn()
	resDoc := fn.Fn(doc)
	if resDoc != objects.FALSE {
		t.Errorf("isHTMLElement(newDocument()) = %v, want FALSE", resDoc)
	}
}

func TestHTMLStripTags_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["stripTags"].(*objects.Builtin)
	input := "<p>Hello <b>world</b>!</p>"
	result := fn.Fn(String(input))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "Hello world!"
	if s.Value != expected {
		t.Errorf("stripTags(%q) = %q, want %q", input, s.Value, expected)
	}
}

func TestHTMLSanitize_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["sanitize"].(*objects.Builtin)
	input := "<script>alert('xss')</script><p>Safe</p>"
	result := fn.Fn(String(input))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "<p>Safe</p>" {
		t.Errorf("sanitize(%q) = %q, want %q", input, s.Value, "<p>Safe</p>")
	}
}

func TestHTMLEscapeAttr_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["escapeAttr"].(*objects.Builtin)
	input := `"quotes' & < >`
	result := fn.Fn(String(input))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `&quot;quotes&#39; &amp; &lt; &gt;`
	if s.Value != expected {
		t.Errorf("escapeAttr(%q) = %q, want %q", input, s.Value, expected)
	}
}

func TestHTMLUnescape_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["unescape"].(*objects.Builtin)
	input := "&lt;div&gt;Hello&amp;"
	result := fn.Fn(String(input))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "<div>Hello&"
	if s.Value != expected {
		t.Errorf("unescape(%q) = %q, want %q", input, s.Value, expected)
	}
}

func TestHTMLParseFragment_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["parseFragment"].(*objects.Builtin)
	input := "<p>One</p><p>Two</p>"
	result := fn.Fn(String(input))
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFragment error: %s", result.Inspect())
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) < 2 {
		t.Errorf("expected at least 2 elements, got %d", len(arr.Elements))
	}
}

func TestHTMLNewTextNode_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["newTextNode"].(*objects.Builtin)
	result := fn.Fn(String("Hello text"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("newTextNode error: %s", result.Inspect())
	}
	if result == objects.NULL {
		t.Error("newTextNode returned NULL")
	}
}

func TestHTMLNewComment_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["newComment"].(*objects.Builtin)
	result := fn.Fn(String("A comment"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("newComment error: %s", result.Inspect())
	}
	if result == objects.NULL {
		t.Error("newComment returned NULL")
	}
}

func TestHTMLEncode_Extra(t *testing.T) {
	mod := Get("html")
	if mod == nil {
		t.Skip("html module not found")
	}
	fn := mod.Exports["encode"].(*objects.Builtin)
	input := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("key").HashKey(): {Key: objects.NewString("key"), Value: objects.NewString("value")},
	}}
	result := fn.Fn(input)
	if result.Type() == objects.ErrorType {
		t.Fatalf("encode error: %s", result.Inspect())
	}
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value == "" {
		t.Error("encode returned empty string")
	}
}
