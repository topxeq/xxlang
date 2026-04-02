// pkg/objects/html_test.go
package objects

import (
	"os"
	"strings"
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

// Additional tests to improve coverage for html_obj.go

func TestHTMLDocument_BasicObjectMethods(t *testing.T) {
	doc := NewHTMLDocument()

	if doc.Type() != HTMLDocumentType {
		t.Errorf("Type() = %v, want %v", doc.Type(), HTMLDocumentType)
	}
	if doc.TypeTag() != TagHTMLDocument {
		t.Errorf("TypeTag() = %v, want %v", doc.TypeTag(), TagHTMLDocument)
	}
	if !doc.ToBool().Value {
		t.Error("ToBool() should return true")
	}
	hk := doc.HashKey()
	if hk.Type != HTMLDocumentType {
		t.Errorf("HashKey().Type = %v, want %v", hk.Type, HTMLDocumentType)
	}
	inspect := doc.Inspect()
	if inspect == "" {
		t.Error("Inspect() should not be empty")
	}
}

func TestHTMLElement_BasicObjectMethods(t *testing.T) {
	elem := NewHTMLElement("div")

	if elem.Type() != HTMLElementType {
		t.Errorf("Type() = %v, want %v", elem.Type(), HTMLElementType)
	}
	if elem.TypeTag() != TagHTMLElement {
		t.Errorf("TypeTag() = %v, want %v", elem.TypeTag(), TagHTMLElement)
	}
	if !elem.ToBool().Value {
		t.Error("ToBool() should return true")
	}
	hk := elem.HashKey()
	if hk.Type != HTMLElementType {
		t.Errorf("HashKey().Type = %v, want %v", hk.Type, HTMLElementType)
	}
	inspect := elem.Inspect()
	if inspect == "" {
		t.Error("Inspect() should not be empty")
	}
}

func TestHTMLDocument_Setters(t *testing.T) {
	doc := NewHTMLDocument()

	// SetRoot
	newRoot := NewHTMLElement("section")
	doc.SetRoot(newRoot)
	if doc.Root() != newRoot {
		t.Error("SetRoot did not set root correctly")
	}
	if newRoot.parentNode != nil {
		t.Error("Root's parent should be nil after setting as root")
	}

	// SetTitle
	doc.SetTitle("New Title")
	if doc.Title() != "New Title" {
		t.Errorf("SetTitle failed, got %s", doc.Title())
	}

	// SetMeta
	doc.SetMeta("description", "Test description")

	// AddStyle
	doc.AddStyle("body { color: red; }")

	// AddScript (two parameters: js, src)
	doc.AddScript("console.log('test');", "")
}

func TestHTMLDocument_FindMethods(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	div := NewHTMLElement("div")
	div.SetAttribute("id", "test")
	body.AppendChild(div)

	// Find (alias for QuerySelectorAll)
	results := doc.Find("#test")
	if results == nil {
		t.Fatal("Find returned nil")
	}
	if len(results.Elements) != 1 {
		t.Errorf("Find expected 1 element, got %d", len(results.Elements))
	}

	// FindFirst (alias for QuerySelector)
	elem := doc.FindFirst("#test")
	if elem == nil {
		t.Fatal("FindFirst returned nil")
	}
	if elem.TagName() != "div" {
		t.Errorf("FindFirst expected div, got %s", elem.TagName())
	}
}

func TestHTMLDocument_ToIndented(t *testing.T) {
	doc := NewHTMLDocumentWithTitle("Test")
	result := doc.ToIndented()
	if result == "" {
		t.Error("ToIndented returned empty")
	}
	if !strings.Contains(result, "\n") {
		t.Error("ToIndented should contain newlines for indentation")
	}
}

func TestHTMLElement_NodeType(t *testing.T) {
	elem := NewHTMLElement("p")
	if elem.NodeType() != HTMLNodeElement {
		t.Errorf("Expected HTMLNodeElement, got %v", elem.NodeType())
	}

	textNode := NewHTMLTextNode("hello")
	if textNode.NodeType() != HTMLNodeText {
		t.Errorf("Expected HTMLNodeText, got %v", textNode.NodeType())
	}
}

func TestHTMLElement_InnerOuterHTML(t *testing.T) {
	elem := NewHTMLElement("div")
	elem.SetAttribute("class", "container")
	elem.SetTextContent("Hello")

	// InnerHTML
	inner := elem.InnerHTML()
	if inner != "Hello" {
		t.Errorf("InnerHTML = %q, want %q", inner, "Hello")
	}

	// OuterHTML
	outer := elem.OuterHTML()
	if outer == "" {
		t.Error("OuterHTML is empty")
	}
	if !strings.Contains(outer, "<div") {
		t.Error("OuterHTML missing opening tag")
	}
	if !strings.Contains(outer, "Hello") {
		t.Error("OuterHTML missing text content")
	}

	// SetInnerHTML
	err := elem.SetInnerHTML("<p>World</p>")
	if err != nil {
		t.Fatalf("SetInnerHTML failed: %v", err)
	}
	newInner := elem.InnerHTML()
	if !strings.Contains(newInner, "<p>") {
		t.Error("SetInnerHTML did not update children")
	}
}

func TestHTMLElement_ClassMethods(t *testing.T) {
	elem := NewHTMLElement("div")

	// Class (getter) - initially empty
	if elem.Class() != "" {
		t.Errorf("Empty class expected, got %s", elem.Class())
	}

	// SetClass
	elem.SetClass("foo bar")
	if elem.Class() != "foo bar" {
		t.Errorf("Class() = %s, want %s", elem.Class(), "foo bar")
	}
}

func TestHTMLElement_ChildrenManipulation(t *testing.T) {
	parent := NewHTMLElement("div")
	child1 := NewHTMLElement("p")
	child2 := NewHTMLElement("span")

	// AppendChild
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	// Children
	children := parent.Children()
	if children == nil || len(children.Elements) != 2 {
		t.Errorf("Expected 2 children, got %v", children)
	}

	// InsertBefore
	newChild := NewHTMLElement("section")
	parent.InsertBefore(newChild, child1)
	children = parent.Children()
	if len(children.Elements) != 3 {
		t.Errorf("Expected 3 children after InsertBefore, got %d", len(children.Elements))
	}
	if children.Elements[0] != newChild {
		t.Error("InsertBefore did not insert at correct position")
	}

	// InsertAfter
	another := NewHTMLElement("article")
	parent.InsertAfter(another, child2)
	children = parent.Children()
	if len(children.Elements) != 4 {
		t.Errorf("Expected 4 children after InsertAfter, got %d", len(children.Elements))
	}

	// ReplaceChild
	replacement := NewHTMLElement("header")
	parent.ReplaceChild(replacement, child1)
	children = parent.Children()
	if len(children.Elements) != 4 {
		t.Errorf("Expected 4 children after ReplaceChild, got %d", len(children.Elements))
	}
	// Verify child1 is removed and replacement is in its place
	foundChild1 := false
	foundReplacement := false
	for _, c := range children.Elements {
		if c == child1 {
			foundChild1 = true
		}
		if c == replacement {
			foundReplacement = true
		}
	}
	if foundChild1 {
		t.Error("child1 should be removed after ReplaceChild")
	}
	if !foundReplacement {
		t.Error("replacement should be in children after ReplaceChild")
	}

	// Clear
	parent.Clear()
	if parent.ChildCount() != 0 {
		t.Errorf("Clear should remove all children, got %d", parent.ChildCount())
	}

	// Remove
	parent2 := NewHTMLElement("div")
	childA := NewHTMLElement("p")
	parent2.AppendChild(childA)
	childA.Remove()
	if parent2.ChildCount() != 0 {
		t.Errorf("After Remove, parent should have 0 children, got %d", parent2.ChildCount())
	}
}

func TestHTMLElement_ToIndented(t *testing.T) {
	div := NewHTMLElement("div")
	p := NewHTMLElement("p")
	p.SetTextContent("Hello")
	div.AppendChild(p)

	result := div.ToIndented()
	if result == "" {
		t.Error("ToIndented returned empty")
	}
	if !strings.Contains(result, "\n") {
		t.Error("ToIndented should contain newlines for nested structure")
	}
	if !strings.Contains(result, "<div") {
		t.Error("ToIndented should contain opening div")
	}
	if !strings.Contains(result, "<p>") {
		t.Error("ToIndented should contain opening p")
	}
	if !strings.Contains(result, "</p>") {
		t.Error("ToIndented should contain closing p")
	}
	if !strings.Contains(result, "</div>") {
		t.Error("ToIndented should contain closing div")
	}
}

func TestHTMLElement_Attributes(t *testing.T) {
	elem := NewHTMLElement("a")
	elem.SetAttribute("href", "https://example.com")
	elem.SetAttribute("title", "Example")

	attrs := elem.Attributes()
	if attrs == nil {
		t.Fatal("Attributes() returned nil")
	}

	// Check href
	hrefKey := NewString("href")
	hrefPair, hasHref := attrs.Pairs[hrefKey.HashKey()]
	if !hasHref {
		t.Error("Expected href attribute")
	}
	hrefStr, ok := hrefPair.Value.(*String)
	if !ok || hrefStr.Value != "https://example.com" {
		t.Errorf("href attribute = %v, want %s", hrefPair.Value, "https://example.com")
	}

	// Check title
	titleKey := NewString("title")
	titlePair, hasTitle := attrs.Pairs[titleKey.HashKey()]
	if !hasTitle {
		t.Error("Expected title attribute")
	}
	titleStr, ok := titlePair.Value.(*String)
	if !ok || titleStr.Value != "Example" {
		t.Errorf("title attribute = %v, want %s", titlePair.Value, "Example")
	}
}

func TestNewHTMLComment(t *testing.T) {
	comment := NewHTMLComment("This is a comment")
	if comment == nil {
		t.Fatal("NewHTMLComment returned nil")
	}
	if comment.NodeType() != HTMLNodeComment {
		t.Errorf("Expected HTMLNodeComment, got %v", comment.NodeType())
	}
	if comment.TextContent() != "This is a comment" {
		t.Errorf("Comment text = %s, want %s", comment.TextContent(), "This is a comment")
	}
	// Comments are HTMLElement objects
	if comment.Type() != HTMLElementType {
		t.Errorf("Type() = %v, want %v", comment.Type(), HTMLElementType)
	}
}

func TestHTMLDocument_GetElementsByTagName(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	p1 := NewHTMLElement("p")
	p2 := NewHTMLElement("p")
	div := NewHTMLElement("div")
	body.AppendChild(p1)
	body.AppendChild(p2)
	body.AppendChild(div)

	result := doc.GetElementsByTagName("p")
	if result == nil {
		t.Fatal("GetElementsByTagName returned nil")
	}
	if len(result.Elements) != 2 {
		t.Errorf("Expected 2 p elements, got %d", len(result.Elements))
	}

	// Test with tag that doesn't exist
	result2 := doc.GetElementsByTagName("span")
	if result2 == nil || len(result2.Elements) != 0 {
		t.Errorf("Expected empty array for non-existent tag")
	}
}

func TestHTMLDocument_GetElementsByClassName(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	d1 := NewHTMLElement("div")
	d1.AddClass("item")
	d1.AddClass("active")
	d2 := NewHTMLElement("div")
	d2.AddClass("item")
	d3 := NewHTMLElement("span")
	d3.AddClass("item")

	body.AppendChild(d1)
	body.AppendChild(d2)
	body.AppendChild(d3)

	result := doc.GetElementsByClassName("item")
	if result == nil {
		t.Fatal("GetElementsByClassName returned nil")
	}
	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 elements with class 'item', got %d", len(result.Elements))
	}

	// Test non-existent class
	result2 := doc.GetElementsByClassName("none")
	if result2 == nil || len(result2.Elements) != 0 {
		t.Errorf("Expected empty array for non-existent class")
	}
}

func TestHTMLDocument_QuerySelector(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	div := NewHTMLElement("div")
	div.SetAttribute("id", "main")
	p := NewHTMLElement("p")
	p.AddClass("text")
	div.AppendChild(p)
	body.AppendChild(div)

	// By ID
	elem := doc.QuerySelector("#main")
	if elem == nil {
		t.Fatal("QuerySelector('#main') returned nil")
	}
	if elem.TagName() != "div" {
		t.Errorf("Expected div, got %s", elem.TagName())
	}

	// By class
	elem2 := doc.QuerySelector(".text")
	if elem2 == nil {
		t.Fatal("QuerySelector('.text') returned nil")
	}
	if elem2.TagName() != "p" {
		t.Errorf("Expected p, got %s", elem2.TagName())
	}

	// Non-existent
	elem3 := doc.QuerySelector("#missing")
	if elem3 != nil {
		t.Error("QuerySelector('#missing') should return nil")
	}
}

func TestHTMLDocument_QuerySelectorAll(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	for i := 0; i < 3; i++ {
		p := NewHTMLElement("p")
		p.AddClass("item")
		body.AppendChild(p)
	}
	div := NewHTMLElement("div")
	div.AddClass("container")
	body.AppendChild(div)

	result := doc.QuerySelectorAll(".item")
	if result == nil {
		t.Fatal("QuerySelectorAll returned nil")
	}
	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 .item elements, got %d", len(result.Elements))
	}
}

func TestHTMLElement_FindMethods(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	div := NewHTMLElement("div")
	p := NewHTMLElement("p")
	p.SetAttribute("id", "target")
	div.AppendChild(p)
	body.AppendChild(div)

	// Find (alias for QuerySelectorAll)
	results := div.Find("p")
	if results == nil {
		t.Fatal("Find returned nil")
	}
	if len(results.Elements) != 1 {
		t.Errorf("Find expected 1 p, got %d", len(results.Elements))
	}

	// FindFirst (alias for QuerySelector)
	elem := div.FindFirst("#target")
	if elem == nil {
		t.Fatal("FindFirst returned nil")
	}
	if elem.TagName() != "p" {
		t.Errorf("FindFirst expected p, got %s", elem.TagName())
	}
}

func TestHTMLElement_Children(t *testing.T) {
	elem := NewHTMLElement("div")
	children := elem.Children()
	// For a new element with no children, Children() returns an empty Array (non-nil)
	if children != nil && len(children.Elements) != 0 {
		t.Errorf("Expected 0 children for new element, got %d", len(children.Elements))
	}
	// After appending, Children should return slice with 1 element
	child := NewHTMLElement("span")
	elem.AppendChild(child)
	children = elem.Children()
	if children == nil || len(children.Elements) != 1 {
		t.Errorf("Expected 1 child after append, got %v", children)
	}
}

// Additional tests for coverage gaps

func TestEscapeHTMLAttr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quotes"`, "&quot;quotes&quot;"},
		{"'apostrophe'", "&#39;apostrophe&#39;"},
		{`a"b'c`, `a&quot;b&#39;c`},
		{"plain", "plain"},
		{"&amp;", "&amp;amp;"},
	}
	for _, tt := range tests {
		result := EscapeHTMLAttr(tt.input)
		if result != tt.expected {
			t.Errorf("EscapeHTMLAttr(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseHTMLFromReader(t *testing.T) {
	html := `<!DOCTYPE html><html><body><div>From Reader</div></body></html>`
	reader := strings.NewReader(html)

	doc, err := ParseHTMLFromReader(reader)
	if err != nil {
		t.Fatalf("ParseHTMLFromReader failed: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	div := doc.FindFirst("div")
	if div == nil {
		t.Fatal("expected div element")
	}
	if div.TextContent() != "From Reader" {
		t.Errorf("expected text 'From Reader', got '%s'", div.TextContent())
	}
}

func TestHTMLElement_SetTagName(t *testing.T) {
	elem := NewHTMLElement("DIV")
	// NewHTMLElement lowercases tag name
	if elem.TagName() != "div" {
		t.Errorf("initial tag should be lowercase 'div', got '%s'", elem.TagName())
	}
	// SetTagName should lowercase
	elem.SetTagName("SPAN")
	if elem.TagName() != "span" {
		t.Errorf("after SetTagName, expected 'span', got '%s'", elem.TagName())
	}
}

func TestHTMLElement_Parent(t *testing.T) {
	parent := NewHTMLElement("div")
	child := NewHTMLElement("span")

	// Initially no parent
	if child.Parent() != nil {
		t.Error("child should have no parent initially")
	}

	// After appending
	parent.AppendChild(child)
	if child.Parent() != parent {
		t.Error("child should reference parent after appending")
	}

	// After removing
	child.Remove()
	if child.Parent() != nil {
		t.Error("child should have no parent after removal")
	}
}

func TestHTMLElement_FirstLastChild(t *testing.T) {
	parent := NewHTMLElement("div")

	// No children
	if parent.FirstChild() != nil {
		t.Error("FirstChild should be nil for empty parent")
	}
	if parent.LastChild() != nil {
		t.Error("LastChild should be nil for empty parent")
	}

	// Single child
	child1 := NewHTMLElement("p")
	parent.AppendChild(child1)
	if parent.FirstChild() != child1 {
		t.Error("FirstChild should return the only child")
	}
	if parent.LastChild() != child1 {
		t.Error("LastChild should return the only child")
	}

	// Multiple children
	child2 := NewHTMLElement("span")
	parent.AppendChild(child2)
	if parent.FirstChild() != child1 {
		t.Error("FirstChild should return first child")
	}
	if parent.LastChild() != child2 {
		t.Error("LastChild should return last child")
	}
}

func TestHTMLElement_ID(t *testing.T) {
	elem := NewHTMLElement("div")

	// Initially no ID
	if elem.ID() != "" {
		t.Errorf("Empty ID expected, got %s", elem.ID())
	}

	// SetID
	elem.SetID("main")
	if elem.ID() != "main" {
		t.Errorf("ID() = %s, want %s", elem.ID(), "main")
	}

	// Attribute should also reflect ID
	if elem.Attribute("id") != "main" {
		t.Errorf("Attribute('id') = %s, want %s", elem.Attribute("id"), "main")
	}

	// SetID should update attribute
	elem.SetID("new-id")
	if elem.ID() != "new-id" {
		t.Errorf("After SetID, ID() = %s, want %s", elem.ID(), "new-id")
	}
	if elem.Attribute("id") != "new-id" {
		t.Errorf("After SetID, Attribute('id') = %s, want %s", elem.Attribute("id"), "new-id")
	}
}

func TestHTMLDocument_GetElementById(t *testing.T) {
	doc := NewHTMLDocument()
	body := doc.Body()

	div1 := NewHTMLElement("div")
	div1.SetID("first")
	div2 := NewHTMLElement("div")
	div2.SetID("second")
	body.AppendChild(div1)
	body.AppendChild(div2)

	found := doc.GetElementById("first")
	if found == nil {
		t.Fatal("GetElementById('first') returned nil")
	}
	if found != div1 {
		t.Error("GetElementById returned wrong element")
	}

	// Check second element
	found2 := doc.GetElementById("second")
	if found2 != div2 {
		t.Error("GetElementById returned wrong element for 'second'")
	}

	// Non-existent ID
	notFound := doc.GetElementById("nonexistent")
	if notFound != nil {
		t.Error("GetElementById should return nil for non-existent ID")
	}

	// Nested elements
	child := NewHTMLElement("span")
	child.SetID("nested")
	div1.AppendChild(child)
	foundNested := doc.GetElementById("nested")
	if foundNested != child {
		t.Error("GetElementById should find nested elements")
	}
}

func TestParseHTMLFile(t *testing.T) {
	// Create temporary HTML file
	tmpDir := t.TempDir()
	tmpPath := tmpDir + "/test.html"
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>File Test</title></head>
<body><div id="content">Hello</div></body>
</html>`

	err := os.WriteFile(tmpPath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	doc, err := ParseHTMLFile(tmpPath)
	if err != nil {
		t.Fatalf("ParseHTMLFile failed: %v", err)
	}

	if doc.Title() != "File Test" {
		t.Errorf("Expected title 'File Test', got '%s'", doc.Title())
	}

	div := doc.GetElementById("content")
	if div == nil {
		t.Fatal("Element with id 'content' not found")
	}
	if div.TextContent() != "Hello" {
		t.Errorf("Expected text 'Hello', got '%s'", div.TextContent())
	}
}

func TestEncodeToHTML_Array(t *testing.T) {
	// Create array with HTMLElements
	elem1 := NewHTMLElement("p")
	elem1.SetTextContent("First")
	elem2 := NewHTMLElement("div")
	elem2.SetTextContent("Second")
	arr := NewArray([]Object{elem1, elem2})

	result, err := EncodeToHTML(arr, "root")
	if err != nil {
		t.Fatalf("EncodeToHTML failed: %v", err)
	}

	if result == "" {
		t.Error("EncodeToHTML returned empty string")
	}
	if !strings.Contains(result, "<p>") {
		t.Error("Result missing <p> tag")
	}
	if !strings.Contains(result, "<div") {
		t.Error("Result missing <div> tag")
	}
	if !strings.Contains(result, "First") {
		t.Error("Result missing first element content")
	}
	if !strings.Contains(result, "Second") {
		t.Error("Result missing second element content")
	}
}

func TestEncodeToHTML_String(t *testing.T) {
	str := NewString("Hello World")
	result, err := EncodeToHTML(str, "div")
	if err != nil {
		t.Fatalf("EncodeToHTML failed: %v", err)
	}

	expected := "<div>Hello World</div>"
	if result != expected {
		t.Errorf("EncodeToHTML(String) = %q, want %q", result, expected)
	}
}

func TestHTMLElement_InsertBeforeEdgeCases(t *testing.T) {
	parent := NewHTMLElement("div")
	child1 := NewHTMLElement("p")
	child2 := NewHTMLElement("span")
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	// InsertBefore with nil reference should fail
	newElem := NewHTMLElement("section")
	result := parent.InsertBefore(newElem, nil)
	if result {
		t.Error("InsertBefore with nil ref should return false")
	}

	// InsertBefore with element not in children should fail
	notChild := NewHTMLElement("aside")
	result = parent.InsertBefore(newElem, notChild)
	if result {
		t.Error("InsertBefore with non-child ref should return false")
	}

	// InsertBefore at beginning
	newElem2 := NewHTMLElement("header")
	result = parent.InsertBefore(newElem2, child1)
	if !result {
		t.Error("InsertBefore should return true on success")
	}
	children := parent.Children()
	if children.Elements[0] != newElem2 {
		t.Error("InsertBefore did not insert at correct position")
	}
	if newElem2.Parent() != parent {
		t.Error("InsertBefore did not set parent correctly")
	}
}

func TestHTMLElement_InsertAfterEdgeCases(t *testing.T) {
	parent := NewHTMLElement("div")
	child1 := NewHTMLElement("p")
	child2 := NewHTMLElement("span")
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	// InsertAfter with nil reference should fail
	newElem := NewHTMLElement("section")
	result := parent.InsertAfter(newElem, nil)
	if result {
		t.Error("InsertAfter with nil ref should return false")
	}

	// InsertAfter with element not in children should fail
	notChild := NewHTMLElement("aside")
	result = parent.InsertAfter(newElem, notChild)
	if result {
		t.Error("InsertAfter with non-child ref should return false")
	}

	// InsertAfter at end
	newElem2 := NewHTMLElement("footer")
	result = parent.InsertAfter(newElem2, child2)
	if !result {
		t.Error("InsertAfter should return true on success")
	}
	children := parent.Children()
	if children.Elements[len(children.Elements)-1] != newElem2 {
		t.Error("InsertAfter did not insert at correct position")
	}
	if newElem2.Parent() != parent {
		t.Error("InsertAfter did not set parent correctly")
	}
}

func TestHTMLElement_ReplaceChildEdgeCases(t *testing.T) {
	parent := NewHTMLElement("div")
	child1 := NewHTMLElement("p")
	child2 := NewHTMLElement("span")
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	// ReplaceChild with nil old element should fail
	newElem := NewHTMLElement("section")
	result := parent.ReplaceChild(newElem, nil)
	if result {
		t.Error("ReplaceChild with nil old element should return false")
	}

	// ReplaceChild with element not in children should fail
	notChild := NewHTMLElement("aside")
	result = parent.ReplaceChild(newElem, notChild)
	if result {
		t.Error("ReplaceChild with non-child should return false")
	}

	// Successful replacement
	replacement := NewHTMLElement("header")
	result = parent.ReplaceChild(replacement, child1)
	if !result {
		t.Error("ReplaceChild should return true on success")
	}
	children := parent.Children()
	if len(children.Elements) != 2 {
		t.Errorf("Expected 2 children after replacement, got %d", len(children.Elements))
	}
	// child1 should be removed
	for _, c := range children.Elements {
		if c == child1 {
			t.Error("child1 should be removed after ReplaceChild")
		}
	}
	// replacement should be present
	if replacement.Parent() != parent {
		t.Error("Replacement's parent not set correctly")
	}
}

func TestHTMLElement_CloneDeep(t *testing.T) {
	parent := NewHTMLElement("div")
	parent.SetAttribute("id", "parent")
	child := NewHTMLElement("p")
	child.SetTextContent("Hello")
	parent.AppendChild(child)

	clone := parent.Clone()

	// Check that clone is a different object
	if clone == parent {
		t.Error("Clone should be a different object")
	}

	// Check attributes are copied
	if clone.Attribute("id") != "parent" {
		t.Errorf("Clone attribute id = %s, want %s", clone.Attribute("id"), "parent")
	}

	// Check children are cloned (not same objects)
	if clone.ChildCount() != 1 {
		t.Errorf("Clone should have 1 child, got %d", clone.ChildCount())
	}
	clonedChild := clone.children[0]
	if clonedChild == child {
		t.Error("Child should be cloned, not the same object")
	}
	if clonedChild.TextContent() != "Hello" {
		t.Errorf("Cloned child text = %s, want %s", clonedChild.TextContent(), "Hello")
	}

	// Check parent references
	if clone.Parent() != nil {
		t.Error("Clone's parent should be nil")
	}
	if clonedChild.Parent() != clone {
		t.Error("Cloned child's parent should be the clone")
	}

	// Modifying clone should not affect original
	clone.SetAttribute("class", "clone")
	if parent.HasAttribute("class") {
		t.Error("Modifying clone should not affect original")
	}
	clonedChild.SetTextContent("Modified")
	if child.TextContent() == "Modified" {
		t.Error("Modifying cloned child should not affect original")
	}
}

func TestSanitizeHTML_AllDangerousPatterns(t *testing.T) {
	html := `<script>alert('xss')</script>
<style>body { color: red; }</style>
<iframe src="evil.com"></iframe>
<a href="javascript:alert('xss')">click</a>
<div onclick="evil()">danger</div>`

	result := SanitizeHTML(html)

	// Check script removed
	if containsScript(result) {
		t.Errorf("SanitizeHTML did not remove script: %s", result)
	}
	// Check style removed
	if strings.Contains(result, "<style") {
		t.Errorf("SanitizeHTML did not remove style: %s", result)
	}
	// Check iframe removed
	if strings.Contains(result, "<iframe") {
		t.Errorf("SanitizeHTML did not remove iframe: %s", result)
	}
	// Check javascript: removed
	if strings.Contains(result, "javascript:") {
		t.Errorf("SanitizeHTML did not remove javascript: %s", result)
	}
	// Check event handlers removed
	if strings.Contains(result, "onclick") {
		t.Errorf("SanitizeHTML did not remove onclick: %s", result)
	}
}
