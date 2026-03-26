// pkg/objects/xml.go
// XML types for Xxlang - XML document and node handling.
package objects

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// XMLDocument represents an XML document.
type XMLDocument struct {
	root      *XMLNode
	version   string
	encoding  string
	standalone string
}

// XMLNode represents an XML element node.
type XMLNode struct {
	name       string
	text       string
	attributes map[string]string
	children   []*XMLNode
	parent     *XMLNode
}

// Type returns the object type.
func (d *XMLDocument) Type() ObjectType { return XMLDocumentType }

// TypeTag returns the fast type tag.
func (d *XMLDocument) TypeTag() TypeTag { return TagXMLDocument }

// Inspect returns a string representation.
func (d *XMLDocument) Inspect() string {
	if d.root == nil {
		return "XMLDocument(empty)"
	}
	return fmt.Sprintf("XMLDocument(root=%s)", d.root.name)
}

// ToBool returns true (XMLDocument is always truthy).
func (d *XMLDocument) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the XMLDocument.
func (d *XMLDocument) HashKey() HashKey {
	return HashKey{
		Type:  XMLDocumentType,
		Value: uint64(uintptr(unsafe.Pointer(d))),
	}
}

// Root returns the root node.
func (d *XMLDocument) Root() *XMLNode {
	return d.root
}

// SetRoot sets the root node.
func (d *XMLDocument) SetRoot(node *XMLNode) {
	d.root = node
	node.parent = nil
}

// Version returns the XML version.
func (d *XMLDocument) Version() string {
	return d.version
}

// Encoding returns the XML encoding.
func (d *XMLDocument) Encoding() string {
	return d.encoding
}

// ToString converts the document to XML string.
func (d *XMLDocument) ToString() string {
	if d.root == nil {
		return ""
	}
	return d.root.toStringInternal(0, false)
}

// ToIndented converts the document to indented XML string.
func (d *XMLDocument) ToIndented() string {
	if d.root == nil {
		return ""
	}
	return d.root.toStringInternal(0, true)
}

// Save saves the document to a file.
func (d *XMLDocument) Save(path string) error {
	content := d.ToString()
	return os.WriteFile(path, []byte(content), 0644)
}

// Find finds nodes by path expression.
func (d *XMLDocument) Find(path string) *Array {
	if d.root == nil {
		return &Array{}
	}
	return d.root.Find(path)
}

// FindFirst finds the first matching node by path.
func (d *XMLDocument) FindFirst(path string) *XMLNode {
	if d.root == nil {
		return nil
	}
	return d.root.FindFirst(path)
}

// FindElement is an alias for FindFirst.
func (d *XMLDocument) FindElement(path string) *XMLNode {
	return d.FindFirst(path)
}

// ToMap converts the document to a Map.
func (d *XMLDocument) ToMap() *Map {
	if d.root == nil {
		return &Map{Pairs: make(map[HashKey]MapPair)}
	}
	return d.root.ToMap()
}

// Type returns the object type.
func (n *XMLNode) Type() ObjectType { return XMLNodeType }

// TypeTag returns the fast type tag.
func (n *XMLNode) TypeTag() TypeTag { return TagXMLNode }

// Inspect returns a string representation.
func (n *XMLNode) Inspect() string {
	return fmt.Sprintf("XMLNode(%s, attrs=%d, children=%d)", n.name, len(n.attributes), len(n.children))
}

// ToBool returns true if the node has content.
func (n *XMLNode) ToBool() *Bool { return &Bool{Value: n.name != ""} }

// HashKey returns a hash key for the XMLNode.
func (n *XMLNode) HashKey() HashKey {
	return HashKey{
		Type:  XMLNodeType,
		Value: uint64(uintptr(unsafe.Pointer(n))),
	}
}

// Name returns the node name.
func (n *XMLNode) Name() string {
	return n.name
}

// SetName sets the node name.
func (n *XMLNode) SetName(name string) {
	n.name = name
}

// Text returns the text content.
func (n *XMLNode) Text() string {
	return n.text
}

// SetText sets the text content.
func (n *XMLNode) SetText(text string) {
	n.text = text
}

// Attr returns the attribute value by name.
func (n *XMLNode) Attr(name string) string {
	return n.attributes[name]
}

// SetAttr sets an attribute.
func (n *XMLNode) SetAttr(name, value string) {
	if n.attributes == nil {
		n.attributes = make(map[string]string)
	}
	n.attributes[name] = value
}

// DelAttr deletes an attribute.
func (n *XMLNode) DelAttr(name string) {
	delete(n.attributes, name)
}

// Attrs returns all attributes as a Map.
func (n *XMLNode) Attrs() *Map {
	pairs := make(map[HashKey]MapPair)
	for k, v := range n.attributes {
		key := NewString(k)
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewString(v)}
	}
	return &Map{Pairs: pairs}
}

// Children returns all child nodes.
func (n *XMLNode) Children() *Array {
	elements := make([]Object, len(n.children))
	for i, child := range n.children {
		elements[i] = child
	}
	return &Array{Elements: elements}
}

// ChildCount returns the number of children.
func (n *XMLNode) ChildCount() int {
	return len(n.children)
}

// Parent returns the parent node.
func (n *XMLNode) Parent() *XMLNode {
	return n.parent
}

// AddChild adds a child node.
func (n *XMLNode) AddChild(child *XMLNode) {
	child.parent = n
	n.children = append(n.children, child)
}

// RemoveChild removes a child node by index.
func (n *XMLNode) RemoveChild(index int) bool {
	if index < 0 || index >= len(n.children) {
		return false
	}
	n.children = append(n.children[:index], n.children[index+1:]...)
	return true
}

// Clear removes all children.
func (n *XMLNode) Clear() {
	n.children = nil
}

// NewXMLDocument creates a new XML document.
func NewXMLDocument() *XMLDocument {
	return &XMLDocument{
		version:  "1.0",
		encoding: "UTF-8",
	}
}

// NewXMLDocumentWithRoot creates a new XML document with a root element.
func NewXMLDocumentWithRoot(rootName string) *XMLDocument {
	doc := NewXMLDocument()
	doc.root = NewXMLNode(rootName)
	return doc
}

// NewXMLNode creates a new XML node.
func NewXMLNode(name string) *XMLNode {
	return &XMLNode{
		name:       name,
		attributes: make(map[string]string),
		children:   make([]*XMLNode, 0),
	}
}

// ParseXML parses an XML string and returns an XMLDocument.
func ParseXML(xmlStr string) (*XMLDocument, error) {
	doc := NewXMLDocument()

	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	decoder.Strict = false

	var stack []*XMLNode
	var current *XMLNode

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.ProcInst:
			if t.Target == "xml" {
				// Parse XML declaration
				parts := strings.Split(string(t.Inst), " ")
				for _, part := range parts {
					if strings.HasPrefix(part, "version=") {
						doc.version = strings.Trim(strings.TrimPrefix(part, "version="), `"`)
					} else if strings.HasPrefix(part, "encoding=") {
						doc.encoding = strings.Trim(strings.TrimPrefix(part, "encoding="), `"`)
					} else if strings.HasPrefix(part, "standalone=") {
						doc.standalone = strings.Trim(strings.TrimPrefix(part, "standalone="), `"`)
					}
				}
			}
		case xml.StartElement:
			node := NewXMLNode(t.Name.Local)
			for _, attr := range t.Attr {
				node.attributes[attr.Name.Local] = attr.Value
			}
			if current != nil {
				current.AddChild(node)
			}
			stack = append(stack, current)
			current = node
			if doc.root == nil {
				doc.root = node
			}
		case xml.EndElement:
			if len(stack) > 0 {
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if current != nil {
				text := strings.TrimSpace(string(t))
				if text != "" {
					current.text += text
				}
			}
		case xml.Comment:
			// Ignore comments for now
		case xml.Directive:
			// Ignore directives
		}
	}

	return doc, nil
}

// ParseXMLFile parses an XML file and returns an XMLDocument.
func ParseXMLFile(path string) (*XMLDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseXML(string(data))
}

// toStringInternal converts the node to XML string.
func (n *XMLNode) toStringInternal(level int, indent bool) string {
	var buf bytes.Buffer
	indentStr := "  "

	if indent {
		for i := 0; i < level; i++ {
			buf.WriteString(indentStr)
		}
	}

	buf.WriteString("<")
	buf.WriteString(n.name)

	// Write attributes
	// Sort for consistent output
	var attrKeys []string
	for k := range n.attributes {
		attrKeys = append(attrKeys, k)
	}
	sort.Strings(attrKeys)
	for _, k := range attrKeys {
		buf.WriteString(" ")
		buf.WriteString(k)
		buf.WriteString("=\"")
		buf.WriteString(escapeXMLAttr(n.attributes[k]))
		buf.WriteString("\"")
	}

	// Check if empty element
	if len(n.children) == 0 && n.text == "" {
		buf.WriteString("/>")
		return buf.String()
	}

	buf.WriteString(">")

	// Write text content
	if n.text != "" {
		buf.WriteString(escapeXMLText(n.text))
	}

	// Write children
	for _, child := range n.children {
		if indent {
			buf.WriteString("\n")
		}
		buf.WriteString(child.toStringInternal(level+1, indent))
	}

	if indent && len(n.children) > 0 {
		buf.WriteString("\n")
		for i := 0; i < level; i++ {
			buf.WriteString(indentStr)
		}
	}

	buf.WriteString("</")
	buf.WriteString(n.name)
	buf.WriteString(">")

	return buf.String()
}

// ToString converts the node to XML string.
func (n *XMLNode) ToString() string {
	return n.toStringInternal(0, false)
}

// ToIndented converts the node to indented XML string.
func (n *XMLNode) ToIndented() string {
	return n.toStringInternal(0, true)
}

// Find finds nodes by path expression.
func (n *XMLNode) Find(path string) *Array {
	nodes := n.findNodes(path)
	elements := make([]Object, len(nodes))
	for i, node := range nodes {
		elements[i] = node
	}
	return &Array{Elements: elements}
}

// FindFirst finds the first matching node by path.
func (n *XMLNode) FindFirst(path string) *XMLNode {
	nodes := n.findNodes(path)
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// findNodes implements the path search algorithm.
func (n *XMLNode) findNodes(path string) []*XMLNode {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	// Handle attribute filter at the end: [@attr='value']
	attrFilter := ""
	if idx := strings.LastIndex(path, "[@"); idx != -1 && strings.HasSuffix(path, "]") {
		attrFilter = path[idx+1 : len(path)-1] // @attr='value'
		path = path[:idx]
	}

	segments := splitXPath(path)
	if len(segments) == 0 {
		return nil
	}

	var results []*XMLNode

	// Parse path type
	anyDepth := false
	isAbsolute := false
	startIdx := 0

	// Check first segment for path type
	if len(segments) > 0 {
		if segments[0] == "/" {
			// Absolute path
			isAbsolute = true
			startIdx = 1
		} else if segments[0] == "//" {
			// Any depth search
			anyDepth = true
			startIdx = 1
		}
	}

	if startIdx >= len(segments) {
		return nil
	}

	// Collect starting nodes
	var currentNodes []*XMLNode

	if anyDepth {
		// Any depth search - collect all matching descendants
		currentNodes = n.collectAllDescendants(segments[startIdx])
		startIdx++ // Move past the search term
	} else if isAbsolute {
		// For absolute path, match against self first
		if segments[startIdx] == "*" || n.name == segments[startIdx] {
			currentNodes = []*XMLNode{n}
			startIdx++ // Move past the root match
		} else {
			return nil // Root name doesn't match
		}
	} else {
		// Relative path - search in children
		currentNodes = n.matchChildren(segments[startIdx])
		startIdx++
	}

	// Process remaining segments
	for i := startIdx; i < len(segments); i++ {
		seg := segments[i]
		if seg == "//" {
			// Switch to any depth search
			if i+1 < len(segments) {
				var nextNodes []*XMLNode
				for _, node := range currentNodes {
					nextNodes = append(nextNodes, node.collectAllDescendants(segments[i+1])...)
				}
				currentNodes = nextNodes
				i++ // Skip the next segment (already processed)
			}
			continue
		}
		if len(currentNodes) == 0 {
			break
		}
		var nextNodes []*XMLNode
		for _, node := range currentNodes {
			nextNodes = append(nextNodes, node.matchChildren(seg)...)
		}
		currentNodes = nextNodes
	}

	// Apply attribute filter
	if attrFilter != "" {
		for _, node := range currentNodes {
			if node.matchesAttrFilter(attrFilter) {
				results = append(results, node)
			}
		}
	} else {
		results = currentNodes
	}

	return results
}

// matchChildren matches children by segment.
func (n *XMLNode) matchChildren(segment string) []*XMLNode {
	var results []*XMLNode

	// Check for index filter: name[0]
	name, index := parseIndexFilter(segment)

	for _, child := range n.children {
		if name == "*" || child.name == name {
			results = append(results, child)
		}
	}

	// Apply index if specified
	if index >= 0 {
		if index < len(results) {
			return []*XMLNode{results[index]}
		}
		return nil
	}

	return results
}

// collectAllDescendants collects all descendants matching the segment.
func (n *XMLNode) collectAllDescendants(segment string) []*XMLNode {
	var results []*XMLNode

	name, index := parseIndexFilter(segment)

	var collect func(node *XMLNode)
	collect = func(node *XMLNode) {
		if name == "*" || node.name == name {
			results = append(results, node)
		}
		for _, child := range node.children {
			collect(child)
		}
	}

	for _, child := range n.children {
		collect(child)
	}

	// Apply index if specified
	if index >= 0 {
		if index < len(results) {
			return []*XMLNode{results[index]}
		}
		return nil
	}

	return results
}

// matchesAttrFilter checks if the node matches the attribute filter.
func (n *XMLNode) matchesAttrFilter(filter string) bool {
	// Parse @attr='value' or @attr="value"
	re := regexp.MustCompile(`^@(\w+)\s*=\s*['"](.+)['"]$`)
	matches := re.FindStringSubmatch(filter)
	if len(matches) == 3 {
		attrName := matches[1]
		attrValue := matches[2]
		return n.attributes[attrName] == attrValue
	}
	// Parse @attr (just check existence)
	if strings.HasPrefix(filter, "@") {
		attrName := strings.TrimPrefix(filter, "@")
		_, exists := n.attributes[attrName]
		return exists
	}
	return false
}

// parseIndexFilter parses "name[0]" into (name, 0) or "name" into (name, -1).
func parseIndexFilter(segment string) (string, int) {
	re := regexp.MustCompile(`^(\w+|\*)\[(\d+)\]$`)
	matches := re.FindStringSubmatch(segment)
	if len(matches) == 3 {
		name := matches[1]
		index, _ := strconv.Atoi(matches[2])
		return name, index
	}
	return segment, -1
}

// splitXPath splits an XPath-like expression into segments.
// Handles: /a/b (absolute), //a (any depth), a/b (relative)
func splitXPath(path string) []string {
	var segments []string
	var current strings.Builder
	inBracket := false
	i := 0

	for i < len(path) {
		ch := rune(path[i])
		switch ch {
		case '[':
			inBracket = true
			current.WriteRune(ch)
		case ']':
			inBracket = false
			current.WriteRune(ch)
		case '/':
			if !inBracket {
				// Check for double slash
				if i+1 < len(path) && path[i+1] == '/' {
					// Double slash - mark as any depth search
					if current.Len() > 0 {
						segments = append(segments, current.String())
						current.Reset()
					}
					segments = append(segments, "//")
					i += 2
					continue
				}
				if current.Len() > 0 {
					segments = append(segments, current.String())
					current.Reset()
				} else if len(segments) == 0 {
					// Leading slash marks absolute path
					segments = append(segments, "/")
				}
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
		i++
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// ToMap converts the node to a Map for easy JSON conversion.
func (n *XMLNode) ToMap() *Map {
	pairs := make(map[HashKey]MapPair)

	// Add attributes with @ prefix
	for k, v := range n.attributes {
		key := NewString("@" + k)
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewString(v)}
	}

	// Add text content
	if n.text != "" {
		key := NewString("#text")
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewString(n.text)}
	}

	// Add children
	childMap := make(map[string][]*XMLNode)
	for _, child := range n.children {
		childMap[child.name] = append(childMap[child.name], child)
	}

	for name, children := range childMap {
		key := NewString(name)
		if len(children) == 1 {
			pairs[key.HashKey()] = MapPair{Key: key, Value: children[0].ToMap()}
		} else {
			elements := make([]Object, len(children))
			for i, child := range children {
				elements[i] = child.ToMap()
			}
			pairs[key.HashKey()] = MapPair{Key: key, Value: NewArray(elements)}
		}
	}

	return &Map{Pairs: pairs}
}

// EscapeXMLAttr escapes special characters for XML attributes.
func EscapeXMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// EscapeXMLText escapes special characters for XML text.
func EscapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// Internal escape functions
func escapeXMLAttr(s string) string { return EscapeXMLAttr(s) }
func escapeXMLText(s string) string { return EscapeXMLText(s) }

// EncodeToXML converts an Object to XML string.
func EncodeToXML(obj Object, rootName string) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("cannot encode nil to XML")
	}

	switch o := obj.(type) {
	case *Map:
		node := mapToXMLNode(rootName, o)
		if node == nil {
			return "", fmt.Errorf("failed to convert map to XML")
		}
		return node.ToString(), nil
	case *XMLNode:
		return o.ToString(), nil
	case *XMLDocument:
		return o.ToString(), nil
	default:
		node := NewXMLNode(rootName)
		node.SetText(o.Inspect())
		return node.ToString(), nil
	}
}

// mapToXMLNode converts a Map to an XMLNode.
func mapToXMLNode(name string, m *Map) *XMLNode {
	node := NewXMLNode(name)

	for _, pair := range m.Pairs {
		keyStr, ok := pair.Key.(*String)
		if !ok {
			continue
		}
		key := keyStr.Value

		if strings.HasPrefix(key, "@") {
			// Attribute
			attrName := strings.TrimPrefix(key, "@")
			if val, ok := pair.Value.(*String); ok {
				node.SetAttr(attrName, val.Value)
			}
		} else if key == "#text" {
			// Text content
			if val, ok := pair.Value.(*String); ok {
				node.SetText(val.Value)
			}
		} else {
			// Child element
			switch v := pair.Value.(type) {
			case *Map:
				child := mapToXMLNode(key, v)
				if child != nil {
					node.AddChild(child)
				}
			case *Array:
				for _, elem := range v.Elements {
					if elemMap, ok := elem.(*Map); ok {
						child := mapToXMLNode(key, elemMap)
						if child != nil {
							node.AddChild(child)
						}
					}
				}
			case *String:
				child := NewXMLNode(key)
				child.SetText(v.Value)
				node.AddChild(child)
			}
		}
	}

	return node
}