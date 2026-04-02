// pkg/stdlib/xml_test.go
// Tests for xml module.
package stdlib

import (
	"os"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestXMLParse(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["parse"].(*objects.Builtin)

	xmlStr := `<root><item>value</item></root>`
	result := fn.Fn(objects.NewString(xmlStr))
	if result.Type() != objects.XMLDocumentType {
		t.Fatalf("expected XMLDocument, got %s", result.Type())
	}
}

func TestXMLParseFile(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["parseFile"].(*objects.Builtin)

	// Create temp XML file
	tmpFile, _ := os.CreateTemp("", "test_*.xml")
	tmpFile.WriteString(`<root><item>value</item></root>`)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	result := fn.Fn(objects.NewString(tmpFile.Name()))
	if result.Type() != objects.XMLDocumentType {
		t.Fatalf("expected XMLDocument, got %s", result.Type())
	}
}

func TestXMLCreate(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["create"].(*objects.Builtin)

	doc := fn.Fn(objects.NewString("root"))
	if doc.Type() != objects.XMLDocumentType {
		t.Fatalf("expected XMLDocument, got %s", doc.Type())
	}
}

func TestXMLNewDocument(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["newDocument"].(*objects.Builtin)

	doc := fn.Fn()
	if doc.Type() != objects.XMLDocumentType {
		t.Fatalf("expected XMLDocument, got %s", doc.Type())
	}
}

func TestXMLNewNode(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["newNode"].(*objects.Builtin)

	node := fn.Fn(objects.NewString("item"))
	if node.Type() != objects.XMLNodeType {
		t.Fatalf("expected XMLNode, got %s", node.Type())
	}
}

func TestXMLIsXMLDocument(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["isXMLDocument"].(*objects.Builtin)

	// Not a document
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	doc := objects.NewXMLDocument()
	res = fn.Fn(doc)
	if b, ok := res.(*objects.Bool); ok {
		if !b.Value {
			t.Error("expected true for XMLDocument")
		}
	}
}

func TestXMLIsXMLNode(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	fn := mod.Exports["isXMLNode"].(*objects.Builtin)

	// Not a node
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	node := objects.NewXMLNode("test")
	res = fn.Fn(node)
	if b, ok := res.(*objects.Bool); ok {
		if !b.Value {
			t.Error("expected true for XMLNode")
		}
	}
}

func TestXMLNodeOperations(t *testing.T) {
	mod := Get("xml")
	if mod == nil {
		t.Skip("xml module not found")
	}
	// Test setting and getting attributes, adding children, etc.
	newNode := mod.Exports["newNode"].(*objects.Builtin)
	setAttr := mod.Exports["setAttribute"].(*objects.Builtin)
	getAttr := mod.Exports["getAttribute"].(*objects.Builtin)
	addChild := mod.Exports["addChild"].(*objects.Builtin)
	getChildren := mod.Exports["getChildren"].(*objects.Builtin)
	setText := mod.Exports["setText"].(*objects.Builtin)
	getText := mod.Exports["getText"].(*objects.Builtin)

	// Create a node
	node := newNode.Fn(objects.NewString("item"))

	// Set attribute
	res := setAttr.Fn(node, objects.NewString("id"), objects.NewString("123"))
	if res.Type() != objects.NullType {
		t.Fatalf("expected Null from setAttribute, got %s", res.Type())
	}

	// Get attribute
	val := getAttr.Fn(node, objects.NewString("id"))
	if str, ok := val.(*objects.String); ok {
		if str.Value != "123" {
			t.Errorf("expected attribute id='123', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", val)
	}

	// Set text
	res = setText.Fn(node, objects.NewString("hello"))
	if res.Type() != objects.NullType {
		t.Fatalf("expected Null from setText, got %s", res.Type())
	}

	// Get text
	val = getText.Fn(node)
	if str, ok := val.(*objects.String); ok {
		if str.Value != "hello" {
			t.Errorf("expected text 'hello', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", val)
	}

	// Add child
	child := newNode.Fn(objects.NewString("child"))
	res = addChild.Fn(node, child)
	if res.Type() != objects.NullType {
		t.Fatalf("expected Null from addChild, got %s", res.Type())
	}

	// Get children
	children := getChildren.Fn(node)
	if arr, ok := children.(*objects.Array); ok {
		if len(arr.Elements) != 1 {
			t.Errorf("expected 1 child, got %d", len(arr.Elements))
		}
	} else {
		t.Fatalf("expected Array, got %T", children)
	}
}
