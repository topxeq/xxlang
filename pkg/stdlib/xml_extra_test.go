// pkg/stdlib/xml_extra_test.go
// Additional tests for xml module to increase coverage.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestXMLEscape tests xml.escape function with various inputs.
func TestXMLEscape(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["escape"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "abc"},
		{"<tag>", "&lt;tag&gt;"},
		{"\"quotes\"", "\"quotes\""},               // quotes not escaped in text
		{"'single'", "'single'"},                   // apostrophe not escaped in text
		{"&ampersand", "&amp;ampersand"},           // & becomes &amp;
		{"a<b>c&d\"e'f", "a&lt;b&gt;c&amp;d\"e'f"}, // only &, <, > escaped
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			s, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}
			if s.Value != tt.expected {
				t.Errorf("escape(%q) = %q, want %q", tt.input, s.Value, tt.expected)
			}
		})
	}
}

// TestXMLEscapeAttr tests xml.escapeAttr with attribute values.
func TestXMLEscapeAttr(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["escapeAttr"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "abc"},
		{"\"quotes\"", "&quot;quotes&quot;"},
		{"'single'", "&apos;single&apos;"},
		{"a&b", "a&amp;b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fn.Fn(String(tt.input))
			s, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}
			if s.Value != tt.expected {
				t.Errorf("escapeAttr(%q) = %q, want %q", tt.input, s.Value, tt.expected)
			}
		})
	}
}

// TestXMLSetGetAttribute tests setAttribute and getAttribute on XML nodes.
func TestXMLSetGetAttribute(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	setAttrFn := mod.Exports["setAttribute"].(*objects.Builtin)
	getAttrFn := mod.Exports["getAttribute"].(*objects.Builtin)
	newNodeFn := mod.Exports["newNode"].(*objects.Builtin)

	// Create a node
	node, _ := newNodeFn.Fn(String("element")).(*objects.XMLNode)

	// Set an attribute
	setResult := setAttrFn.Fn(node, String("key"), String("value"))
	if setResult.Type() != objects.NullType {
		t.Errorf("setAttribute returned non-null: %s", setResult.Inspect())
	}

	// Get the attribute
	getResult := getAttrFn.Fn(node, String("key"))
	s, ok := getResult.(*objects.String)
	if !ok {
		t.Fatalf("expected string, got %T", getResult)
	}
	if s.Value != "value" {
		t.Errorf("getAttribute('key') = %q, want %q", s.Value, "value")
	}

	// Get non-existent attribute (should return empty string)
	missing := getAttrFn.Fn(node, String("nonexistent"))
	if missing == objects.NULL {
		t.Errorf("expected empty string for missing attribute, got NULL")
	}
	if s, ok := missing.(*objects.String); ok {
		if s.Value != "" {
			t.Errorf("expected empty string for missing attribute, got %q", s.Value)
		}
	} else {
		t.Errorf("expected String, got %T", missing)
	}
}

// TestXMLAddGetChildren tests addChild and getChildren.
func TestXMLAddGetChildren(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	addChildFn := mod.Exports["addChild"].(*objects.Builtin)
	getChildrenFn := mod.Exports["getChildren"].(*objects.Builtin)
	newNodeFn := mod.Exports["newNode"].(*objects.Builtin)

	parent, _ := newNodeFn.Fn(String("parent")).(*objects.XMLNode)
	child1, _ := newNodeFn.Fn(String("child1")).(*objects.XMLNode)
	child2, _ := newNodeFn.Fn(String("child2")).(*objects.XMLNode)

	// Add children
	addChildFn.Fn(parent, child1)
	addChildFn.Fn(parent, child2)

	// Get children
	result := getChildrenFn.Fn(parent)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 children, got %d", len(arr.Elements))
	}
}

// TestXMLSetGetText tests setText and getText.
func TestXMLSetGetText(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	setTextFn := mod.Exports["setText"].(*objects.Builtin)
	getTextFn := mod.Exports["getText"].(*objects.Builtin)
	newNodeFn := mod.Exports["newNode"].(*objects.Builtin)

	node, _ := newNodeFn.Fn(String("element")).(*objects.XMLNode)

	// Set text
	setResult := setTextFn.Fn(node, String("Hello World"))
	if setResult.Type() != objects.NullType {
		t.Errorf("setText returned non-null: %s", setResult.Inspect())
	}

	// Get text
	getResult := getTextFn.Fn(node)
	s, ok := getResult.(*objects.String)
	if !ok {
		t.Fatalf("expected string, got %T", getResult)
	}
	if s.Value != "Hello World" {
		t.Errorf("getText() = %q, want %q", s.Value, "Hello World")
	}
}

// TestXMLIsXMLDocument_Extra tests isXMLDocument.
func TestXMLIsXMLDocument_Extra(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["isXMLDocument"].(*objects.Builtin)
	newDocFn := mod.Exports["newDocument"].(*objects.Builtin)
	newNodeFn := mod.Exports["newNode"].(*objects.Builtin)

	doc, _ := newDocFn.Fn().(*objects.XMLDocument)
	node, _ := newNodeFn.Fn(String("node")).(*objects.XMLNode)

	// Document should be true
	resDoc := fn.Fn(doc)
	if resDoc != objects.TRUE {
		t.Errorf("isXMLDocument(document) = %v, want true", resDoc)
	}

	// Node should be false
	resNode := fn.Fn(node)
	if resNode != objects.FALSE {
		t.Errorf("isXMLDocument(node) = %v, want false", resNode)
	}
}

// TestXMLIsXMLNode_Extra tests isXMLNode.
func TestXMLIsXMLNode_Extra(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["isXMLNode"].(*objects.Builtin)
	newDocFn := mod.Exports["newDocument"].(*objects.Builtin)
	newNodeFn := mod.Exports["newNode"].(*objects.Builtin)

	doc, _ := newDocFn.Fn().(*objects.XMLDocument)
	node, _ := newNodeFn.Fn(String("node")).(*objects.XMLNode)

	// Node should be true
	resNode := fn.Fn(node)
	if resNode != objects.TRUE {
		t.Errorf("isXMLNode(node) = %v, want true", resNode)
	}

	// Document should be false (document is not a node)
	resDoc := fn.Fn(doc)
	if resDoc != objects.FALSE {
		t.Errorf("isXMLNode(document) = %v, want false", resDoc)
	}
}
