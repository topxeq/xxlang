// pkg/objects/xml_obj_test.go
package objects

import (
	"testing"
)

func TestNewXMLDocument(t *testing.T) {
	doc := NewXMLDocument()
	if doc == nil {
		t.Fatal("expected document instance")
	}
}

func TestNewXMLDocumentWithRoot(t *testing.T) {
	doc := NewXMLDocumentWithRoot("root")
	if doc == nil {
		t.Fatal("expected document instance")
	}
	if doc.Root() == nil {
		t.Error("expected root node")
	}
}

func TestXMLDocumentType(t *testing.T) {
	doc := NewXMLDocument()
	if doc.Type() != XMLDocumentType {
		t.Errorf("expected XMLDocumentType, got %v", doc.Type())
	}
}

func TestXMLDocumentInspect(t *testing.T) {
	doc := NewXMLDocument()
	s := doc.Inspect()
	if s == "" {
		t.Error("expected non-empty inspect string")
	}
}

func TestXMLDocumentToBool(t *testing.T) {
	doc := NewXMLDocument()
	if doc.ToBool() != TRUE {
		t.Error("expected TRUE")
	}
}

func TestParseXML(t *testing.T) {
	xml := `<root><child>value</child></root>`
	doc, err := ParseXML(xml)
	if err != nil {
		t.Fatalf("ParseXML error: %v", err)
	}
	if doc == nil {
		t.Error("expected non-nil document")
	}
}

func TestXMLDocumentRoot(t *testing.T) {
	doc := NewXMLDocumentWithRoot("root")
	root := doc.Root()
	if root == nil {
		t.Error("expected non-nil root")
	}
}

func TestNewXMLNode(t *testing.T) {
	node := NewXMLNode("test")
	if node == nil {
		t.Fatal("expected node instance")
	}
	if node.Name() != "test" {
		t.Errorf("expected 'test', got '%s'", node.Name())
	}
}

func TestXMLNodeSetText(t *testing.T) {
	node := NewXMLNode("test")
	node.SetText("value")
	if node.Text() != "value" {
		t.Errorf("expected 'value', got '%s'", node.Text())
	}
}

func TestXMLNodeSetAttr(t *testing.T) {
	node := NewXMLNode("test")
	node.SetAttr("key", "value")
	if node.Attr("key") != "value" {
		t.Errorf("expected 'value', got '%s'", node.Attr("key"))
	}
}

func TestXMLNodeChildren(t *testing.T) {
	node := NewXMLNode("parent")
	child := NewXMLNode("child")
	node.AddChild(child)

	children := node.Children()
	if len(children.Elements) != 1 {
		t.Errorf("expected 1 child, got %d", len(children.Elements))
	}
}
