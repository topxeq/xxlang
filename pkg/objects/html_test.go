// pkg/objects/html_test.go
package objects

import (
	"testing"
)

func TestNewHTMLDocument(t *testing.T) {
	doc := NewHTMLDocument()
	if doc == nil {
		t.Fatal("NewHTMLDocument returned nil")
	}
	if doc.Root() == nil {
		t.Error("Document root should not be nil")
	}
	if doc.Head() == nil {
		t.Error("Document head should not be nil")
	}
	if doc.Body() == nil {
		t.Error("Document body should not be nil")
	}
	if doc.DocType() != "<!DOCTYPE html>" {
		t.Errorf("Expected doctype '<!DOCTYPE html>', got '%s'", doc.DocType())
	}
}

func TestNewHTMLDocumentWithTitle(t *testing.T) {
	doc := NewHTMLDocumentWithTitle("Test Page")
	if doc.Title() != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", doc.Title())
	}
}

func TestNewHTMLElement(t *testing.T) {
	elem := NewHTMLElement("div")
	if elem == nil {
		t.Fatal("NewHTMLElement returned nil")
	}
	if elem.TagName() != "div" {
		t.Errorf("Expected tag name 'div', got '%s'", elem.TagName())
	}
	if elem.ChildCount() != 0 {
		t.Errorf("Expected 0 children, got %d", elem.ChildCount())
	}
}

func TestHTMLElementAttributes(t *testing.T) {
	elem := NewHTMLElement("div")
	elem.SetAttribute("id", "main")
	elem.SetAttribute("class", "container")

	if elem.Attribute("id") != "main" {
		t.Errorf("Expected id 'main', got '%s'", elem.Attribute("id"))
	}
	if elem.Attribute("class") != "container" {
		t.Errorf("Expected class 'container', got '%s'", elem.Attribute("class"))
	}
	if !elem.HasAttribute("id") {
		t.Error("Element should have 'id' attribute")
	}
	elem.RemoveAttribute("id")
	if elem.HasAttribute("id") {
		t.Error("Element should not have 'id' attribute after removal")
	}
}

func TestHTMLElementClasses(t *testing.T) {
	elem := NewHTMLElement("div")

	elem.AddClass("container")
	if !elem.HasClass("container") {
		t.Error("Element should have 'container' class")
	}

	elem.AddClass("main")
	if !elem.HasClass("main") {
		t.Error("Element should have 'main' class")
	}

	elem.RemoveClass("container")
	if elem.HasClass("container") {
		t.Error("Element should not have 'container' class after removal")
	}

	elem.ToggleClass("main")
	if elem.HasClass("main") {
		t.Error("Element should not have 'main' class after toggle")
	}
}

func TestHTMLElementChildren(t *testing.T) {
	parent := NewHTMLElement("div")
	child1 := NewHTMLElement("p")
	child2 := NewHTMLElement("span")

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	if parent.ChildCount() != 2 {
		t.Errorf("Expected 2 children, got %d", parent.ChildCount())
	}

	if parent.FirstChild() != child1 {
		t.Error("FirstChild should return child1")
	}

	if parent.LastChild() != child2 {
		t.Error("LastChild should return child2")
	}

	parent.RemoveChild(0)
	if parent.ChildCount() != 1 {
		t.Errorf("Expected 1 child after removal, got %d", parent.ChildCount())
	}
}

func TestHTMLElementTextContent(t *testing.T) {
	elem := NewHTMLElement("p")
	elem.SetTextContent("Hello World")

	if elem.TextContent() != "Hello World" {
		t.Errorf("Expected text content 'Hello World', got '%s'", elem.TextContent())
	}
}

func TestHTMLElementClone(t *testing.T) {
	original := NewHTMLElement("div")
	original.SetAttribute("id", "original")
	child := NewHTMLElement("p")
	original.AppendChild(child)

	clone := original.Clone()

	if clone == original {
		t.Error("Clone should be a different object")
	}
	if clone.Attribute("id") != "original" {
		t.Error("Clone should have same attributes")
	}
	if clone.ChildCount() != 1 {
		t.Error("Clone should have same number of children")
	}
}

func TestParseHTML(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<div id="main" class="container">
		<p>Hello World</p>
	</div>
</body>
</html>`

	doc, err := ParseHTML(html)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title() != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", doc.Title())
	}

	div := doc.GetElementById("main")
	if div == nil {
		t.Fatal("Element with id 'main' not found")
	}
	if div.Attribute("class") != "container" {
		t.Errorf("Expected class 'container', got '%s'", div.Attribute("class"))
	}
}

func TestParseHTMLFragment(t *testing.T) {
	html := `<p>First</p><p>Second</p>`

	elements, err := ParseHTMLFragment(html)
	if err != nil {
		t.Fatalf("ParseHTMLFragment failed: %v", err)
	}

	if len(elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(elements))
	}
}

func TestHTMLElementQuerySelector(t *testing.T) {
	html := `<div>
		<p class="first">One</p>
		<p class="second">Two</p>
		<span id="test">Three</span>
	</div>`

	elements, err := ParseHTMLFragment(html)
	if err != nil {
		t.Fatalf("ParseHTMLFragment failed: %v", err)
	}

	if len(elements) == 0 {
		t.Fatal("No elements parsed")
	}

	div := elements[0]

	elem := div.QuerySelector("#test")
	if elem == nil {
		t.Fatal("Element with id 'test' not found")
	}
	if elem.TagName() != "span" {
		t.Errorf("Expected tag 'span', got '%s'", elem.TagName())
	}

	elems := div.QuerySelectorAll("p")
	if len(elems.Elements) != 2 {
		t.Errorf("Expected 2 <p> elements, got %d", len(elems.Elements))
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{"Hello", "Hello"},
	}

	for _, test := range tests {
		result := EscapeHTML(test.input)
		if result != test.expected {
			t.Errorf("EscapeHTML(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestUnescapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"&lt;script&gt;", "<script>"},
		{"a &amp; b", "a & b"},
		{"Hello", "Hello"},
	}

	for _, test := range tests {
		result := UnescapeHTML(test.input)
		if result != test.expected {
			t.Errorf("UnescapeHTML(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestStripTags(t *testing.T) {
	html := "<p>Hello <b>World</b></p>"
	result := StripTags(html)
	expected := "Hello World"
	if result != expected {
		t.Errorf("StripTags(%q) = %q, want %q", html, result, expected)
	}
}

func TestSanitizeHTML(t *testing.T) {
	html := `<script>alert('xss')</script><p>Safe</p>`
	result := SanitizeHTML(html)
	if containsScript(result) {
		t.Errorf("SanitizeHTML did not remove script tag: %s", result)
	}
}

func containsScript(s string) bool {
	return len(s) > 0 && (s[0:7] == "<script" || s[len(s)-9:] == "</script>")
}

func TestHTMLDocumentToMap(t *testing.T) {
	doc := NewHTMLDocumentWithTitle("Test")
	m := doc.ToMap()

	if m == nil {
		t.Fatal("ToMap returned nil")
	}

	titleKey := NewString("title")
	pair, ok := m.Pairs[titleKey.HashKey()]
	if !ok {
		t.Fatal("Map should have 'title' key")
	}
	titleStr, ok := pair.Value.(*String)
	if !ok {
		t.Fatal("title should be a String")
	}
	if titleStr.Value != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", titleStr.Value)
	}
}

func TestHTMLElementToMap(t *testing.T) {
	elem := NewHTMLElement("div")
	elem.SetAttribute("id", "test")
	elem.SetTextContent("Hello")

	m := elem.ToMap()

	if m == nil {
		t.Fatal("ToMap returned nil")
	}

	tagKey := NewString("tagName")
	pair, ok := m.Pairs[tagKey.HashKey()]
	if !ok {
		t.Fatal("Map should have 'tagName' key")
	}
	tagStr, ok := pair.Value.(*String)
	if !ok {
		t.Fatal("tagName should be a String")
	}
	if tagStr.Value != "div" {
		t.Errorf("Expected tagName 'div', got '%s'", tagStr.Value)
	}
}

func TestHTMLElementToString(t *testing.T) {
	elem := NewHTMLElement("div")
	elem.SetAttribute("id", "test")
	elem.SetTextContent("Hello")

	result := elem.ToString()
	expected := `<div id="test">Hello</div>`
	if result != expected {
		t.Errorf("ToString() = %q, want %q", result, expected)
	}
}

func TestHTMLDocumentSaveAndLoad(t *testing.T) {
	doc := NewHTMLDocumentWithTitle("Test Save")
	div := NewHTMLElement("div")
	div.SetAttribute("id", "content")
	div.SetTextContent("Test content")
	doc.Body().AppendChild(div)

	tmpPath := t.TempDir() + "/test.html"
	err := doc.Save(tmpPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loadedDoc, err := ParseHTMLFile(tmpPath)
	if err != nil {
		t.Fatalf("ParseHTMLFile failed: %v", err)
	}

	if loadedDoc.Title() != "Test Save" {
		t.Errorf("Expected title 'Test Save', got '%s'", loadedDoc.Title())
	}
}

func TestEncodeToHTML(t *testing.T) {
	pairs := make(map[HashKey]MapPair)
	pairs[NewString("tagName").HashKey()] = MapPair{Key: NewString("tagName"), Value: NewString("div")}
	pairs[NewString("id").HashKey()] = MapPair{Key: NewString("id"), Value: NewString("test")}
	pairs[NewString("text").HashKey()] = MapPair{Key: NewString("text"), Value: NewString("Hello")}
	m := &Map{Pairs: pairs}

	result, err := EncodeToHTML(m, "div")
	if err != nil {
		t.Fatalf("EncodeToHTML failed: %v", err)
	}

	if result == "" {
		t.Error("EncodeToHTML returned empty string")
	}
}
