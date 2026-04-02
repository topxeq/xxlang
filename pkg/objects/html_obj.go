// pkg/objects/html_obj.go
// HTML types for Xxlang - HTML document and element handling.
package objects

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unsafe"
)

// HTMLDocument represents an HTML document.
type HTMLDocument struct {
	docType string
	root    *HTMLElement
	head    *HTMLElement
	body    *HTMLElement
	title   string
}

// HTMLElement represents an HTML element node.
type HTMLElement struct {
	tagName     string
	attributes  map[string]string
	children    []*HTMLElement
	textContent string
	parentNode  *HTMLElement
	selfClosing bool
	nodeType    HTMLNodeType
}

// HTMLNodeType represents the type of HTML node.
type HTMLNodeType int

const (
	HTMLNodeElement HTMLNodeType = iota
	HTMLNodeText
	HTMLNodeComment
	HTMLNodeDoctype
)

// Type returns the object type.
func (d *HTMLDocument) Type() ObjectType { return HTMLDocumentType }

// TypeTag returns the fast type tag.
func (d *HTMLDocument) TypeTag() TypeTag { return TagHTMLDocument }

// Inspect returns a string representation.
func (d *HTMLDocument) Inspect() string {
	if d.root == nil {
		return "HTMLDocument(empty)"
	}
	return fmt.Sprintf("HTMLDocument(root=%s)", d.root.tagName)
}

// ToBool returns true (HTMLDocument is always truthy).
func (d *HTMLDocument) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the HTMLDocument.
func (d *HTMLDocument) HashKey() HashKey {
	return HashKey{
		Type:  HTMLDocumentType,
		Value: uint64(uintptr(unsafe.Pointer(d))),
	}
}

// Type returns the object type.
func (e *HTMLElement) Type() ObjectType { return HTMLElementType }

// TypeTag returns the fast type tag.
func (e *HTMLElement) TypeTag() TypeTag { return TagHTMLElement }

// Inspect returns a string representation.
func (e *HTMLElement) Inspect() string {
	if e.nodeType == HTMLNodeText {
		return fmt.Sprintf("HTMLText(%q)", e.textContent)
	}
	return fmt.Sprintf("HTMLElement(%s, attrs=%d, children=%d)", e.tagName, len(e.attributes), len(e.children))
}

// ToBool returns true if the element has content.
func (e *HTMLElement) ToBool() *Bool {
	if e.nodeType == HTMLNodeText {
		return &Bool{Value: e.textContent != ""}
	}
	return &Bool{Value: e.tagName != ""}
}

// HashKey returns a hash key for the HTMLElement.
func (e *HTMLElement) HashKey() HashKey {
	return HashKey{
		Type:  HTMLElementType,
		Value: uint64(uintptr(unsafe.Pointer(e))),
	}
}

// ============================================================
// HTMLDocument Methods
// ============================================================

// DocType returns the document type declaration.
func (d *HTMLDocument) DocType() string {
	return d.docType
}

// Root returns the root element.
func (d *HTMLDocument) Root() *HTMLElement {
	return d.root
}

// SetRoot sets the root element.
func (d *HTMLDocument) SetRoot(root *HTMLElement) {
	d.root = root
	if root != nil {
		root.parentNode = nil
	}
}

// Head returns the head element.
func (d *HTMLDocument) Head() *HTMLElement {
	return d.head
}

// Body returns the body element.
func (d *HTMLDocument) Body() *HTMLElement {
	return d.body
}

// Title returns the document title.
func (d *HTMLDocument) Title() string {
	return d.title
}

// SetTitle sets the document title.
func (d *HTMLDocument) SetTitle(title string) {
	d.title = title
	if d.head != nil {
		titleElem := d.head.findChildByTag("title")
		if titleElem == nil {
			titleElem = NewHTMLElement("title")
			d.head.AppendChild(titleElem)
		}
		titleElem.SetTextContent(title)
	}
}

// GetElementById returns the element with the specified ID.
func (d *HTMLDocument) GetElementById(id string) *HTMLElement {
	if d.root == nil {
		return nil
	}
	return d.root.findElementByID(id)
}

// GetElementsByTagName returns all elements with the specified tag name.
func (d *HTMLDocument) GetElementsByTagName(tag string) *Array {
	if d.root == nil {
		return &Array{}
	}
	elements := d.root.findElementsByTag(tag)
	result := make([]Object, len(elements))
	for i, elem := range elements {
		result[i] = elem
	}
	return &Array{Elements: result}
}

// GetElementsByClassName returns all elements with the specified class name.
func (d *HTMLDocument) GetElementsByClassName(className string) *Array {
	if d.root == nil {
		return &Array{}
	}
	elements := d.root.findElementsByClass(className)
	result := make([]Object, len(elements))
	for i, elem := range elements {
		result[i] = elem
	}
	return &Array{Elements: result}
}

// QuerySelector returns the first element matching the CSS selector.
func (d *HTMLDocument) QuerySelector(selector string) *HTMLElement {
	if d.root == nil {
		return nil
	}
	return d.root.querySelector(selector)
}

// QuerySelectorAll returns all elements matching the CSS selector.
func (d *HTMLDocument) QuerySelectorAll(selector string) *Array {
	if d.root == nil {
		return &Array{}
	}
	elements := d.root.querySelectorAll(selector)
	result := make([]Object, len(elements))
	for i, elem := range elements {
		result[i] = elem
	}
	return &Array{Elements: result}
}

// Find is an alias for QuerySelectorAll.
func (d *HTMLDocument) Find(selector string) *Array {
	return d.QuerySelectorAll(selector)
}

// FindFirst is an alias for QuerySelector.
func (d *HTMLDocument) FindFirst(selector string) *HTMLElement {
	return d.QuerySelector(selector)
}

// ToString converts the document to HTML string.
func (d *HTMLDocument) ToString() string {
	var buf bytes.Buffer
	if d.docType != "" {
		buf.WriteString(d.docType)
		buf.WriteString("\n")
	}
	if d.root != nil {
		buf.WriteString(d.root.toStringInternal(0, false))
	}
	return buf.String()
}

// ToIndented converts the document to indented HTML string.
func (d *HTMLDocument) ToIndented() string {
	var buf bytes.Buffer
	if d.docType != "" {
		buf.WriteString(d.docType)
		buf.WriteString("\n")
	}
	if d.root != nil {
		buf.WriteString(d.root.toStringInternal(0, true))
	}
	return buf.String()
}

// Save saves the document to a file.
func (d *HTMLDocument) Save(path string) error {
	content := d.ToString()
	return os.WriteFile(path, []byte(content), 0644)
}

// SetMeta sets a meta tag in the head.
func (d *HTMLDocument) SetMeta(name, content string) {
	if d.head == nil {
		return
	}
	meta := d.head.findChildByTagAndAttr("meta", "name", name)
	if meta == nil {
		meta = NewHTMLElement("meta")
		meta.selfClosing = true
		d.head.AppendChild(meta)
	}
	meta.SetAttribute("name", name)
	meta.SetAttribute("content", content)
}

// AddStyle adds a style tag to the head.
func (d *HTMLDocument) AddStyle(css string) {
	if d.head == nil {
		return
	}
	style := NewHTMLElement("style")
	style.SetTextContent(css)
	d.head.AppendChild(style)
}

// AddScript adds a script tag to the body.
func (d *HTMLDocument) AddScript(js string, src string) {
	if d.body == nil {
		return
	}
	script := NewHTMLElement("script")
	if src != "" {
		script.SetAttribute("src", src)
	} else {
		script.SetTextContent(js)
	}
	d.body.AppendChild(script)
}

// ============================================================
// HTMLElement Methods
// ============================================================

// TagName returns the tag name.
func (e *HTMLElement) TagName() string {
	return e.tagName
}

// SetTagName sets the tag name.
func (e *HTMLElement) SetTagName(name string) {
	e.tagName = strings.ToLower(name)
}

// NodeType returns the node type.
func (e *HTMLElement) NodeType() HTMLNodeType {
	return e.nodeType
}

// TextContent returns the text content.
func (e *HTMLElement) TextContent() string {
	if e.nodeType == HTMLNodeText || e.nodeType == HTMLNodeComment {
		return e.textContent
	}
	var buf bytes.Buffer
	e.collectText(&buf)
	return buf.String()
}

// SetTextContent sets the text content.
func (e *HTMLElement) SetTextContent(text string) {
	e.textContent = text
	if e.nodeType == HTMLNodeElement && !e.selfClosing {
		e.children = nil
		textNode := NewHTMLTextNode(text)
		textNode.parentNode = e
		e.children = append(e.children, textNode)
	}
}

// InnerHTML returns the inner HTML.
func (e *HTMLElement) InnerHTML() string {
	if e.nodeType == HTMLNodeText {
		return EscapeHTML(e.textContent)
	}
	var buf bytes.Buffer
	for _, child := range e.children {
		buf.WriteString(child.toStringInternal(0, false))
	}
	return buf.String()
}

// SetInnerHTML sets the inner HTML by parsing it.
func (e *HTMLElement) SetInnerHTML(html string) error {
	e.children = nil
	if html == "" {
		return nil
	}
	nodes, err := ParseHTMLFragment(html)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		node.parentNode = e
		e.children = append(e.children, node)
	}
	return nil
}

// OuterHTML returns the outer HTML.
func (e *HTMLElement) OuterHTML() string {
	return e.toStringInternal(0, false)
}

// Attribute returns the attribute value by name.
func (e *HTMLElement) Attribute(name string) string {
	return e.attributes[name]
}

// SetAttribute sets an attribute.
func (e *HTMLElement) SetAttribute(name, value string) {
	if e.attributes == nil {
		e.attributes = make(map[string]string)
	}
	e.attributes[strings.ToLower(name)] = value
}

// HasAttribute checks if an attribute exists.
func (e *HTMLElement) HasAttribute(name string) bool {
	_, ok := e.attributes[strings.ToLower(name)]
	return ok
}

// RemoveAttribute removes an attribute.
func (e *HTMLElement) RemoveAttribute(name string) {
	delete(e.attributes, strings.ToLower(name))
}

// Attributes returns all attributes as a Map.
func (e *HTMLElement) Attributes() *Map {
	pairs := make(map[HashKey]MapPair)
	for k, v := range e.attributes {
		key := NewString(k)
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewString(v)}
	}
	return &Map{Pairs: pairs}
}

// Class returns the class attribute.
func (e *HTMLElement) Class() string {
	return e.attributes["class"]
}

// SetClass sets the class attribute.
func (e *HTMLElement) SetClass(class string) {
	e.SetAttribute("class", class)
}

// AddClass adds a class to the element.
func (e *HTMLElement) AddClass(className string) {
	classes := strings.Fields(e.attributes["class"])
	for _, c := range classes {
		if c == className {
			return
		}
	}
	classes = append(classes, className)
	e.SetAttribute("class", strings.Join(classes, " "))
}

// RemoveClass removes a class from the element.
func (e *HTMLElement) RemoveClass(className string) {
	classes := strings.Fields(e.attributes["class"])
	var newClasses []string
	for _, c := range classes {
		if c != className {
			newClasses = append(newClasses, c)
		}
	}
	if len(newClasses) > 0 {
		e.SetAttribute("class", strings.Join(newClasses, " "))
	} else {
		e.RemoveAttribute("class")
	}
}

// HasClass checks if the element has a class.
func (e *HTMLElement) HasClass(className string) bool {
	classes := strings.Fields(e.attributes["class"])
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

// ToggleClass toggles a class on the element.
func (e *HTMLElement) ToggleClass(className string) {
	if e.HasClass(className) {
		e.RemoveClass(className)
	} else {
		e.AddClass(className)
	}
}

// ID returns the id attribute.
func (e *HTMLElement) ID() string {
	return e.attributes["id"]
}

// SetID sets the id attribute.
func (e *HTMLElement) SetID(id string) {
	e.SetAttribute("id", id)
}

// Children returns all child elements.
func (e *HTMLElement) Children() *Array {
	elements := make([]Object, len(e.children))
	for i, child := range e.children {
		elements[i] = child
	}
	return &Array{Elements: elements}
}

// ChildCount returns the number of children.
func (e *HTMLElement) ChildCount() int {
	return len(e.children)
}

// FirstChild returns the first child element.
func (e *HTMLElement) FirstChild() *HTMLElement {
	if len(e.children) == 0 {
		return nil
	}
	return e.children[0]
}

// LastChild returns the last child element.
func (e *HTMLElement) LastChild() *HTMLElement {
	if len(e.children) == 0 {
		return nil
	}
	return e.children[len(e.children)-1]
}

// Parent returns the parent element.
func (e *HTMLElement) Parent() *HTMLElement {
	return e.parentNode
}

// AppendChild appends a child element.
func (e *HTMLElement) AppendChild(child *HTMLElement) {
	child.parentNode = e
	e.children = append(e.children, child)
}

// RemoveChild removes a child element by index.
func (e *HTMLElement) RemoveChild(index int) bool {
	if index < 0 || index >= len(e.children) {
		return false
	}
	e.children[index].parentNode = nil
	e.children = append(e.children[:index], e.children[index+1:]...)
	return true
}

// InsertBefore inserts a new element before a reference element.
func (e *HTMLElement) InsertBefore(newElem, refElem *HTMLElement) bool {
	for i, child := range e.children {
		if child == refElem {
			newElem.parentNode = e
			e.children = append(e.children[:i], append([]*HTMLElement{newElem}, e.children[i:]...)...)
			return true
		}
	}
	return false
}

// InsertAfter inserts a new element after a reference element.
func (e *HTMLElement) InsertAfter(newElem, refElem *HTMLElement) bool {
	for i, child := range e.children {
		if child == refElem {
			newElem.parentNode = e
			if i+1 >= len(e.children) {
				e.children = append(e.children, newElem)
			} else {
				e.children = append(e.children[:i+1], append([]*HTMLElement{newElem}, e.children[i+1:]...)...)
			}
			return true
		}
	}
	return false
}

// ReplaceChild replaces a child element with a new element.
func (e *HTMLElement) ReplaceChild(newElem, oldElem *HTMLElement) bool {
	for i, child := range e.children {
		if child == oldElem {
			oldElem.parentNode = nil
			newElem.parentNode = e
			e.children[i] = newElem
			return true
		}
	}
	return false
}

// Clear removes all children.
func (e *HTMLElement) Clear() {
	for _, child := range e.children {
		child.parentNode = nil
	}
	e.children = nil
}

// Remove removes the element from its parent.
func (e *HTMLElement) Remove() {
	if e.parentNode == nil {
		return
	}
	for i, child := range e.parentNode.children {
		if child == e {
			e.parentNode.children = append(e.parentNode.children[:i], e.parentNode.children[i+1:]...)
			e.parentNode = nil
			return
		}
	}
}

// Clone returns a deep clone of the element.
func (e *HTMLElement) Clone() *HTMLElement {
	clone := NewHTMLElement(e.tagName)
	clone.selfClosing = e.selfClosing
	clone.nodeType = e.nodeType
	clone.textContent = e.textContent
	for k, v := range e.attributes {
		clone.attributes[k] = v
	}
	for _, child := range e.children {
		childClone := child.Clone()
		childClone.parentNode = clone
		clone.children = append(clone.children, childClone)
	}
	return clone
}

// QuerySelector returns the first element matching the CSS selector.
func (e *HTMLElement) QuerySelector(selector string) *HTMLElement {
	return e.querySelector(selector)
}

// QuerySelectorAll returns all elements matching the CSS selector.
func (e *HTMLElement) QuerySelectorAll(selector string) *Array {
	elements := e.querySelectorAll(selector)
	result := make([]Object, len(elements))
	for i, elem := range elements {
		result[i] = elem
	}
	return &Array{Elements: result}
}

// Find is an alias for QuerySelectorAll.
func (e *HTMLElement) Find(selector string) *Array {
	return e.QuerySelectorAll(selector)
}

// FindFirst is an alias for QuerySelector.
func (e *HTMLElement) FindFirst(selector string) *HTMLElement {
	return e.QuerySelector(selector)
}

// ToString converts the element to HTML string.
func (e *HTMLElement) ToString() string {
	return e.toStringInternal(0, false)
}

// ToIndented converts the element to indented HTML string.
func (e *HTMLElement) ToIndented() string {
	return e.toStringInternal(0, true)
}

// ============================================================
// Constructor Functions
// ============================================================

// NewHTMLDocument creates a new HTML document.
func NewHTMLDocument() *HTMLDocument {
	doc := &HTMLDocument{
		docType: "<!DOCTYPE html>",
	}
	html := NewHTMLElement("html")
	doc.root = html

	head := NewHTMLElement("head")
	html.AppendChild(head)
	doc.head = head

	body := NewHTMLElement("body")
	html.AppendChild(body)
	doc.body = body

	return doc
}

// NewHTMLDocumentWithTitle creates a new HTML document with a title.
func NewHTMLDocumentWithTitle(title string) *HTMLDocument {
	doc := NewHTMLDocument()
	doc.SetTitle(title)
	return doc
}

// NewHTMLElement creates a new HTML element.
func NewHTMLElement(tagName string) *HTMLElement {
	return &HTMLElement{
		tagName:     strings.ToLower(tagName),
		attributes:  make(map[string]string),
		children:    make([]*HTMLElement, 0),
		nodeType:    HTMLNodeElement,
		selfClosing: isSelfClosingTag(strings.ToLower(tagName)),
	}
}

// NewHTMLTextNode creates a new text node.
func NewHTMLTextNode(text string) *HTMLElement {
	return &HTMLElement{
		nodeType:    HTMLNodeText,
		textContent: text,
	}
}

// NewHTMLComment creates a new comment node.
func NewHTMLComment(text string) *HTMLElement {
	return &HTMLElement{
		nodeType:    HTMLNodeComment,
		textContent: text,
	}
}

// isSelfClosingTag returns true if the tag is self-closing.
func isSelfClosingTag(tag string) bool {
	selfClosingTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return selfClosingTags[tag]
}

// ============================================================
// HTML Parsing
// ============================================================

// ParseHTML parses an HTML string and returns an HTMLDocument.
func ParseHTML(htmlStr string) (*HTMLDocument, error) {
	p := newHTMLParser(htmlStr)
	return p.parseDocument()
}

// ParseHTMLFile parses an HTML file and returns an HTMLDocument.
func ParseHTMLFile(path string) (*HTMLDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHTML(string(data))
}

// ParseHTMLFragment parses an HTML fragment and returns elements.
func ParseHTMLFragment(htmlStr string) ([]*HTMLElement, error) {
	p := newHTMLParser(htmlStr)
	return p.parseFragment()
}

// htmlParser is a simple HTML parser.
type htmlParser struct {
	input   string
	pos     int
	len     int
	inTag   bool
	tagName string
}

func newHTMLParser(input string) *htmlParser {
	return &htmlParser{
		input: input,
		len:   len(input),
	}
}

func (p *htmlParser) parseDocument() (*HTMLDocument, error) {
	doc := &HTMLDocument{}

	p.skipWhitespace()

	if strings.HasPrefix(p.input[p.pos:], "<!DOCTYPE") || strings.HasPrefix(p.input[p.pos:], "<!doctype") {
		doc.docType = p.parseDoctype()
	}

	p.skipWhitespace()

	// Parse root element
	for p.pos < p.len {
		p.skipWhitespace()
		if p.pos >= p.len {
			break
		}

		if p.peek() == '<' {
			elem, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			if elem != nil && elem.nodeType == HTMLNodeElement {
				doc.root = elem
				break
			}
		} else {
			p.pos++
		}
	}

	// Find head and body
	if doc.root != nil {
		for _, child := range doc.root.children {
			if child.tagName == "head" {
				doc.head = child
			} else if child.tagName == "body" {
				doc.body = child
			}
		}
	}

	// Extract title
	if doc.head != nil {
		titleElem := doc.head.findChildByTag("title")
		if titleElem != nil {
			doc.title = titleElem.TextContent()
		}
	}

	return doc, nil
}

func (p *htmlParser) parseFragment() ([]*HTMLElement, error) {
	var elements []*HTMLElement

	for p.pos < p.len {
		p.skipWhitespace()
		if p.pos >= p.len {
			break
		}

		if p.peek() == '<' {
			elem, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			if elem != nil {
				elements = append(elements, elem)
			}
		} else {
			text := p.parseText()
			if text != "" {
				elements = append(elements, NewHTMLTextNode(text))
			}
		}
	}

	return elements, nil
}

func (p *htmlParser) parseDoctype() string {
	start := p.pos
	p.pos += 9 // skip "<!DOCTYPE" or "<!doctype"
	for p.pos < p.len && p.input[p.pos] != '>' {
		p.pos++
	}
	if p.pos < p.len {
		p.pos++ // skip '>'
	}
	return p.input[start:p.pos]
}

func (p *htmlParser) parseElement() (*HTMLElement, error) {
	if p.peek() != '<' {
		return nil, fmt.Errorf("expected '<'")
	}
	p.pos++ // skip '<'

	// Check for comment
	if strings.HasPrefix(p.input[p.pos:], "!--") {
		return p.parseComment(), nil
	}

	// Parse tag name
	tagName := p.parseTagName()
	if tagName == "" {
		return nil, fmt.Errorf("empty tag name")
	}

	// Check for closing tag
	if strings.HasPrefix(tagName, "/") {
		p.skipToTagEnd()
		return nil, nil
	}

	elem := NewHTMLElement(tagName)

	// Parse attributes
	p.skipWhitespace()
	for p.pos < p.len && p.peek() != '>' && p.peek() != '/' {
		attrName, attrValue := p.parseAttribute()
		if attrName != "" {
			elem.SetAttribute(attrName, attrValue)
		}
		p.skipWhitespace()
	}

	// Check for self-closing
	if p.peek() == '/' {
		p.pos++
		elem.selfClosing = true
	}

	if p.peek() == '>' {
		p.pos++
	}

	// Parse children if not self-closing
	if !elem.selfClosing && !isSelfClosingTag(tagName) {
		for p.pos < p.len {
			// Check for closing tag
			if strings.HasPrefix(p.input[p.pos:], "</") {
				p.pos += 2
				closeTag := p.parseTagName()
				if strings.EqualFold(closeTag, tagName) {
					p.skipToTagEnd()
					break
				}
				// Mismatched closing tag, continue
			}

			if p.peek() == '<' {
				child, err := p.parseElement()
				if err != nil {
					// Skip error and continue
					continue
				}
				if child != nil {
					child.parentNode = elem
					elem.children = append(elem.children, child)
				}
			} else {
				text := p.parseText()
				if text != "" {
					textNode := NewHTMLTextNode(text)
					textNode.parentNode = elem
					elem.children = append(elem.children, textNode)
				}
			}
		}
	}

	return elem, nil
}

func (p *htmlParser) parseComment() *HTMLElement {
	p.pos += 3 // skip "!--"
	start := p.pos
	for p.pos < p.len-2 {
		if p.input[p.pos] == '-' && p.input[p.pos+1] == '-' && p.input[p.pos+2] == '>' {
			break
		}
		p.pos++
	}
	text := p.input[start:p.pos]
	if p.pos < p.len-2 {
		p.pos += 3 // skip "-->"
	}
	return NewHTMLComment(text)
}

func (p *htmlParser) parseTagName() string {
	start := p.pos
	for p.pos < p.len {
		c := p.input[p.pos]
		if isTagNameChar(c) {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *htmlParser) parseAttribute() (string, string) {
	// Parse attribute name
	start := p.pos
	for p.pos < p.len {
		c := p.input[p.pos]
		if isAttrNameChar(c) {
			p.pos++
		} else {
			break
		}
	}
	name := p.input[start:p.pos]

	if name == "" {
		return "", ""
	}

	p.skipWhitespace()

	// Check for =
	if p.peek() != '=' {
		// Boolean attribute
		return name, ""
	}
	p.pos++ // skip '='
	p.skipWhitespace()

	// Parse value
	var value string
	if p.peek() == '"' || p.peek() == '\'' {
		quote := p.input[p.pos]
		p.pos++
		start := p.pos
		for p.pos < p.len && p.input[p.pos] != quote {
			p.pos++
		}
		value = p.input[start:p.pos]
		if p.pos < p.len {
			p.pos++ // skip closing quote
		}
	} else {
		start := p.pos
		for p.pos < p.len {
			c := p.input[p.pos]
			if c == '>' || c == '/' || unicode.IsSpace(rune(c)) {
				break
			}
			p.pos++
		}
		value = p.input[start:p.pos]
	}

	return name, value
}

func (p *htmlParser) parseText() string {
	start := p.pos
	for p.pos < p.len && p.peek() != '<' {
		p.pos++
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *htmlParser) skipWhitespace() {
	for p.pos < p.len && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *htmlParser) skipToTagEnd() {
	for p.pos < p.len && p.input[p.pos] != '>' {
		p.pos++
	}
	if p.pos < p.len {
		p.pos++
	}
}

func (p *htmlParser) peek() byte {
	if p.pos >= p.len {
		return 0
	}
	return p.input[p.pos]
}

func isTagNameChar(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '-' || c == '_' || c == ':'
}

func isAttrNameChar(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '-' || c == '_' || c == ':'
}

// ============================================================
// Internal Helper Methods
// ============================================================

func (e *HTMLElement) toStringInternal(level int, indent bool) string {
	var buf bytes.Buffer
	indentStr := "  "

	switch e.nodeType {
	case HTMLNodeText:
		buf.WriteString(EscapeHTML(e.textContent))
		return buf.String()
	case HTMLNodeComment:
		buf.WriteString("<!--")
		buf.WriteString(e.textContent)
		buf.WriteString("-->")
		return buf.String()
	}

	if indent {
		for i := 0; i < level; i++ {
			buf.WriteString(indentStr)
		}
	}

	buf.WriteString("<")
	buf.WriteString(e.tagName)

	// Write attributes
	var attrKeys []string
	for k := range e.attributes {
		attrKeys = append(attrKeys, k)
	}
	sort.Strings(attrKeys)
	for _, k := range attrKeys {
		buf.WriteString(" ")
		buf.WriteString(k)
		buf.WriteString("=\"")
		buf.WriteString(EscapeHTMLAttr(e.attributes[k]))
		buf.WriteString("\"")
	}

	if e.selfClosing || isSelfClosingTag(e.tagName) {
		buf.WriteString("/>")
		return buf.String()
	}

	buf.WriteString(">")

	// Write children
	if len(e.children) > 0 {
		// Check if only text children
		onlyText := true
		for _, child := range e.children {
			if child.nodeType == HTMLNodeElement {
				onlyText = false
				break
			}
		}

		for _, child := range e.children {
			if indent && !onlyText && child.nodeType == HTMLNodeElement {
				buf.WriteString("\n")
			}
			buf.WriteString(child.toStringInternal(level+1, indent))
		}

		if indent && !onlyText && len(e.children) > 0 {
			buf.WriteString("\n")
			for i := 0; i < level; i++ {
				buf.WriteString(indentStr)
			}
		}
	}

	buf.WriteString("</")
	buf.WriteString(e.tagName)
	buf.WriteString(">")

	return buf.String()
}

func (e *HTMLElement) collectText(buf *bytes.Buffer) {
	if e.nodeType == HTMLNodeText {
		buf.WriteString(e.textContent)
		return
	}
	for _, child := range e.children {
		child.collectText(buf)
	}
}

func (e *HTMLElement) findChildByTag(tag string) *HTMLElement {
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement && child.tagName == tag {
			return child
		}
	}
	return nil
}

func (e *HTMLElement) findChildByTagAndAttr(tag, attrName, attrValue string) *HTMLElement {
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement && child.tagName == tag {
			if child.Attribute(attrName) == attrValue {
				return child
			}
		}
	}
	return nil
}

func (e *HTMLElement) findElementByID(id string) *HTMLElement {
	if e.attributes["id"] == id {
		return e
	}
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement {
			if found := child.findElementByID(id); found != nil {
				return found
			}
		}
	}
	return nil
}

func (e *HTMLElement) findElementsByTag(tag string) []*HTMLElement {
	var result []*HTMLElement
	tag = strings.ToLower(tag)
	if e.tagName == tag {
		result = append(result, e)
	}
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement {
			result = append(result, child.findElementsByTag(tag)...)
		}
	}
	return result
}

func (e *HTMLElement) findElementsByClass(className string) []*HTMLElement {
	var result []*HTMLElement
	if e.HasClass(className) {
		result = append(result, e)
	}
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement {
			result = append(result, child.findElementsByClass(className)...)
		}
	}
	return result
}

// ============================================================
// CSS Selector Implementation
// ============================================================

func (e *HTMLElement) querySelector(selector string) *HTMLElement {
	elements := e.querySelectorAll(selector)
	if len(elements) > 0 {
		return elements[0]
	}
	return nil
}

func (e *HTMLElement) querySelectorAll(selector string) []*HTMLElement {
	selectors := parseCSSSelector(selector)
	var result []*HTMLElement
	e.collectBySelector(selectors, &result)
	return result
}

type cssSelector struct {
	tag      string
	id       string
	classes  []string
	attrs    map[string]string
	children []*cssSelector // nested selectors
}

func parseCSSSelector(selector string) []*cssSelector {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil
	}

	// Handle comma-separated selectors
	if strings.Contains(selector, ",") {
		parts := strings.Split(selector, ",")
		var selectors []*cssSelector
		for _, part := range parts {
			subs := parseCSSSelector(strings.TrimSpace(part))
			selectors = append(selectors, subs...)
		}
		return selectors
	}

	// Parse descendant selectors
	var selectors []*cssSelector
	var current *cssSelector
	var inAttr bool
	var attrQuote byte

	start := 0
	for i := 0; i < len(selector); i++ {
		c := selector[i]

		if inAttr {
			if attrQuote != 0 {
				if c == attrQuote {
					attrValue := selector[start:i]
					inAttr = false
					attrQuote = 0
					start = i + 2 // skip ] or similar
					if current != nil && current.attrs != nil {
						// attrName was set when entering attribute mode
					}
					_ = attrValue // placeholder for future attribute handling
				}
			} else if c == '"' || c == '\'' {
				attrQuote = c
				start = i + 1
			} else if c == ']' {
				attrValue := selector[start:i]
				inAttr = false
				start = i + 1
				_ = attrValue // placeholder for future attribute handling
			}
			continue
		}

		switch c {
		case '#':
			if current == nil {
				current = &cssSelector{tag: "*"}
			}
			if i > start && current.id == "" && len(current.classes) == 0 && current.tag == "*" {
				current.tag = selector[start:i]
			}
			start = i + 1
			for j := start; j < len(selector); j++ {
				if isSelectorEnd(selector[j]) {
					current.id = selector[start:j]
					start = j
					i = j - 1
					break
				}
			}
			// If we reached the end without finding a selector end, take the rest
			if current.id == "" && start < len(selector) {
				current.id = selector[start:]
				start = len(selector)
			}
		case '.':
			if current == nil {
				current = &cssSelector{tag: "*"}
			}
			if i > start && current.id == "" && len(current.classes) == 0 && current.tag == "*" {
				current.tag = selector[start:i]
			}
			start = i + 1
			for j := start; j < len(selector); j++ {
				if isSelectorEnd(selector[j]) {
					current.classes = append(current.classes, selector[start:j])
					start = j
					i = j - 1
					break
				}
			}
			// If we reached the end without finding a selector end, take the rest
			if len(current.classes) == 0 && start < len(selector) {
				current.classes = append(current.classes, selector[start:])
				start = len(selector)
			}
		case '[':
			if current == nil {
				current = &cssSelector{tag: "*"}
			}
			if i > start && current.tag == "*" {
				current.tag = selector[start:i]
			}
			start = i + 1
			inAttr = true
			// Find attribute name
			var attrName string
			for j := start; j < len(selector); j++ {
				if selector[j] == '=' || selector[j] == ']' {
					attrName = selector[start:j]
					if selector[j] == '=' {
						start = j + 1
					} else {
						inAttr = false
						start = j + 1
					}
					break
				}
			}
			if current.attrs == nil {
				current.attrs = make(map[string]string)
			}
			if attrName != "" {
				// Mark that we're expecting an attribute with this name
				// Value will be set when the attribute is fully parsed
				_ = attrName
			}
		case ' ', '>', '+', '~':
			if current == nil {
				current = &cssSelector{tag: selector[start:i]}
			} else if i > start && current.tag == "*" {
				current.tag = selector[start:i]
			}
			selectors = append(selectors, current)
			current = nil
			start = i + 1
		}
	}

	if current == nil && start < len(selector) {
		current = &cssSelector{tag: selector[start:]}
	} else if current != nil && current.tag == "*" && start < len(selector) {
		current.tag = selector[start:]
	}
	if current != nil {
		selectors = append(selectors, current)
	}

	return selectors
}

func isSelectorEnd(c byte) bool {
	return c == '#' || c == '.' || c == '[' || c == ' ' || c == '>' || c == '+' || c == '~' || c == ','
}

func (e *HTMLElement) collectBySelector(selectors []*cssSelector, result *[]*HTMLElement) {
	if len(selectors) == 0 {
		return
	}

	sel := selectors[0]
	if e.matchesSelector(sel) {
		if len(selectors) == 1 {
			*result = append(*result, e)
		} else {
			// Continue with child selectors
			for _, child := range e.children {
				if child.nodeType == HTMLNodeElement {
					child.collectBySelector(selectors[1:], result)
				}
			}
		}
	}

	// Always check children for first selector
	for _, child := range e.children {
		if child.nodeType == HTMLNodeElement {
			child.collectBySelector(selectors, result)
		}
	}
}

func (e *HTMLElement) matchesSelector(sel *cssSelector) bool {
	// Check tag
	if sel.tag != "" && sel.tag != "*" && e.tagName != sel.tag {
		return false
	}

	// Check ID
	if sel.id != "" && e.attributes["id"] != sel.id {
		return false
	}

	// Check classes
	for _, class := range sel.classes {
		if !e.HasClass(class) {
			return false
		}
	}

	// Check attributes
	for name, value := range sel.attrs {
		if value == "" {
			if !e.HasAttribute(name) {
				return false
			}
		} else {
			if e.Attribute(name) != value {
				return false
			}
		}
	}

	return true
}

// ============================================================
// HTML Utility Functions
// ============================================================

// EscapeHTML escapes special characters for HTML content.
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// EscapeHTMLAttr escapes special characters for HTML attributes.
func EscapeHTMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// UnescapeHTML unescapes HTML entities.
func UnescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}

// StripTags removes all HTML tags from a string.
func StripTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// SanitizeHTML removes potentially dangerous HTML content.
func SanitizeHTML(html string) string {
	// Remove script tags and content
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = scriptRe.ReplaceAllString(html, "")

	// Remove style tags and content
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = styleRe.ReplaceAllString(html, "")

	// Remove iframe tags
	iframeRe := regexp.MustCompile(`(?is)<iframe[^>]*>.*?</iframe>`)
	html = iframeRe.ReplaceAllString(html, "")

	// Remove on* event handlers
	eventRe := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	html = eventRe.ReplaceAllString(html, "")

	// Remove javascript: URLs
	jsRe := regexp.MustCompile(`(?i)javascript\s*:`)
	html = jsRe.ReplaceAllString(html, "")

	return html
}

// EncodeToHTML converts an Object to HTML string.
func EncodeToHTML(obj Object, rootName string) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("cannot encode nil to HTML")
	}

	switch o := obj.(type) {
	case *Map:
		elem := mapToHTMLElement(rootName, o)
		if elem == nil {
			return "", fmt.Errorf("failed to convert map to HTML")
		}
		return elem.ToString(), nil
	case *HTMLElement:
		return o.ToString(), nil
	case *HTMLDocument:
		return o.ToString(), nil
	case *String:
		elem := NewHTMLElement(rootName)
		elem.SetTextContent(o.Value)
		return elem.ToString(), nil
	case *Array:
		var buf bytes.Buffer
		for _, item := range o.Elements {
			if elem, ok := item.(*HTMLElement); ok {
				buf.WriteString(elem.ToString())
			}
		}
		return buf.String(), nil
	default:
		elem := NewHTMLElement(rootName)
		elem.SetTextContent(o.Inspect())
		return elem.ToString(), nil
	}
}

// mapToHTMLElement converts a Map to an HTMLElement.
func mapToHTMLElement(tagName string, m *Map) *HTMLElement {
	elem := NewHTMLElement(tagName)

	for _, pair := range m.Pairs {
		keyStr, ok := pair.Key.(*String)
		if !ok {
			continue
		}
		key := keyStr.Value

		switch {
		case key == "tagName" || key == "tag":
			if val, ok := pair.Value.(*String); ok {
				elem.tagName = strings.ToLower(val.Value)
			}
		case key == "text" || key == "textContent":
			if val, ok := pair.Value.(*String); ok {
				elem.SetTextContent(val.Value)
			}
		case key == "id":
			if val, ok := pair.Value.(*String); ok {
				elem.SetID(val.Value)
			}
		case key == "class":
			if val, ok := pair.Value.(*String); ok {
				elem.SetClass(val.Value)
			}
		case key == "attrs" || key == "attributes":
			if attrsMap, ok := pair.Value.(*Map); ok {
				for _, attrPair := range attrsMap.Pairs {
					if attrKey, ok := attrPair.Key.(*String); ok {
						if attrVal, ok := attrPair.Value.(*String); ok {
							elem.SetAttribute(attrKey.Value, attrVal.Value)
						}
					}
				}
			}
		case key == "children":
			if childrenArr, ok := pair.Value.(*Array); ok {
				for _, child := range childrenArr.Elements {
					switch c := child.(type) {
					case *HTMLElement:
						childClone := c.Clone()
						elem.AppendChild(childClone)
					case *Map:
						childElem := mapToHTMLElement("div", c)
						if childElem != nil {
							elem.AppendChild(childElem)
						}
					case *String:
						textNode := NewHTMLTextNode(c.Value)
						elem.AppendChild(textNode)
					}
				}
			}
		default:
			if strings.HasPrefix(key, "@") {
				attrName := strings.TrimPrefix(key, "@")
				if val, ok := pair.Value.(*String); ok {
					elem.SetAttribute(attrName, val.Value)
				}
			} else if pair.Value != nil {
				if val, ok := pair.Value.(*String); ok {
					child := NewHTMLElement(key)
					child.SetTextContent(val.Value)
					elem.AppendChild(child)
				} else if childMap, ok := pair.Value.(*Map); ok {
					child := mapToHTMLElement(key, childMap)
					if child != nil {
						elem.AppendChild(child)
					}
				}
			}
		}
	}

	return elem
}

// ============================================================
// HTML to Map conversion
// ============================================================

// ToMap converts the element to a Map.
func (e *HTMLElement) ToMap() *Map {
	pairs := make(map[HashKey]MapPair)

	pairs[NewString("tagName").HashKey()] = MapPair{Key: NewString("tagName"), Value: NewString(e.tagName)}
	pairs[NewString("textContent").HashKey()] = MapPair{Key: NewString("textContent"), Value: NewString(e.textContent)}

	// Add attributes
	attrsPairs := make(map[HashKey]MapPair)
	for k, v := range e.attributes {
		key := NewString(k)
		attrsPairs[key.HashKey()] = MapPair{Key: key, Value: NewString(v)}
	}
	pairs[NewString("attributes").HashKey()] = MapPair{Key: NewString("attributes"), Value: &Map{Pairs: attrsPairs}}

	// Add children
	var childElements []Object
	for _, child := range e.children {
		childElements = append(childElements, child.ToMap())
	}
	pairs[NewString("children").HashKey()] = MapPair{Key: NewString("children"), Value: NewArray(childElements)}

	return &Map{Pairs: pairs}
}

// ToMap converts the document to a Map.
func (d *HTMLDocument) ToMap() *Map {
	pairs := make(map[HashKey]MapPair)

	pairs[NewString("docType").HashKey()] = MapPair{Key: NewString("docType"), Value: NewString(d.docType)}
	pairs[NewString("title").HashKey()] = MapPair{Key: NewString("title"), Value: NewString(d.title)}

	if d.root != nil {
		pairs[NewString("root").HashKey()] = MapPair{Key: NewString("root"), Value: d.root.ToMap()}
	}

	return &Map{Pairs: pairs}
}

// ============================================================
// Reader interface for parsing
// ============================================================

// ParseHTMLFromReader parses HTML from an io.Reader.
func ParseHTMLFromReader(r io.Reader) (*HTMLDocument, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseHTML(string(data))
}
