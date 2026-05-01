package dom

import "strings"

type NodeType int

const (
	ElementNode  NodeType = 1
	TextNode     NodeType = 3
	CommentNode  NodeType = 8
	DocumentNode NodeType = 9
)

type Node struct {
	Type        NodeType
	Data        string
	Attr        []Attribute
	Children    []*Node
	Parent      *Node
	PrevSibling *Node
	NextSibling *Node
}

type Attribute struct {
	Key   string
	Value string
}

type Document struct {
	Root *Node
}

func NewDocument() *Document {
	return &Document{
		Root: &Node{Type: DocumentNode, Data: "#document"},
	}
}

func (d *Document) AppendChild(n *Node) {
	if d.Root.Children == nil {
		d.Root.Children = make([]*Node, 0)
	}
	n.Parent = d.Root
	d.Root.Children = append(d.Root.Children, n)
}

func (n *Node) AppendChild(child *Node) {
	if n.Children == nil {
		n.Children = make([]*Node, 0)
	}
	child.Parent = n
	if len(n.Children) > 0 {
		last := n.Children[len(n.Children)-1]
		last.NextSibling = child
		child.PrevSibling = last
	}
	n.Children = append(n.Children, child)
}

func (n *Node) SetAttribute(key, value string) {
	for i, attr := range n.Attr {
		if attr.Key == key {
			n.Attr[i].Value = value
			return
		}
	}
	n.Attr = append(n.Attr, Attribute{Key: key, Value: value})
}

func (n *Node) GetAttribute(key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

func (n *Node) HasAttribute(key string) bool {
	return n.GetAttribute(key) != ""
}

func (n *Node) RemoveAttribute(key string) {
	for i, attr := range n.Attr {
		if attr.Key == key {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return
		}
	}
}

func (n *Node) TextContent() string {
	if n.Type == TextNode {
		return n.Data
	}
	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(child.TextContent())
	}
	return sb.String()
}

// VisibleText returns text content excluding style and script tags.
// This mimics browser behavior where CSS and JS content is not visible.
func (n *Node) VisibleText() string {
	return n.visibleText(false, false)
}

// visibleText recursively builds visible text, skipping style/script tags.
// isBlock indicates if we're at a block-level element (should add newlines).
func (n *Node) visibleText(inHiddenTag bool, isBlock bool) string {
	if n.Type == TextNode {
		if inHiddenTag {
			return ""
		}
		return n.Data
	}

	// Check if this is a tag whose content should be hidden
	tagName := strings.ToLower(n.Data)
	isHidden := inHiddenTag || tagName == "style" || tagName == "script" || tagName == "noscript" || tagName == "head" || tagName == "title"

	// Block-level elements that should have spacing
	blockElements := map[string]bool{
		"p": true, "div": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true, "br": true, "li": true,
		"tr": true, "td": true, "th": true, "section": true, "article": true,
		"header": true, "footer": true, "main": true, "nav": true,
	}

	var sb strings.Builder
	for i, child := range n.Children {
		childText := child.visibleText(isHidden, blockElements[tagName])
		sb.WriteString(childText)
		// Add newline after block elements
		if !isHidden && blockElements[child.Data] && i < len(n.Children)-1 {
			// Only add newline if not already ending with newline
			text := sb.String()
			if len(text) > 0 && text[len(text)-1] != '\n' {
				sb.WriteString("\n")
			}
		}
	}

	// Add newline at end of block element
	result := sb.String()
	if !isHidden && blockElements[tagName] && len(result) > 0 {
		if result[len(result)-1] != '\n' {
			result += "\n"
		}
	}

	return result
}

func (n *Node) InnerHTML() string {
	var sb strings.Builder
	for _, child := range n.Children {
		writeNodeHTML(&sb, child)
	}
	return sb.String()
}

func (n *Node) OuterHTML() string {
	var sb strings.Builder
	writeNodeHTML(&sb, n)
	return sb.String()
}

func writeNodeHTML(sb *strings.Builder, n *Node) {
	if n.Type == TextNode {
		sb.WriteString(n.Data)
		return
	}
	if n.Type == CommentNode {
		sb.WriteString("<!--")
		sb.WriteString(n.Data)
		sb.WriteString("-->")
		return
	}

	sb.WriteString("<")
	sb.WriteString(n.Data)
	for _, attr := range n.Attr {
		sb.WriteString(" ")
		sb.WriteString(attr.Key)
		if attr.Value != "" {
			sb.WriteString("=\"")
			sb.WriteString(attr.Value)
			sb.WriteString("\"")
		}
	}

	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}

	if voidElements[n.Data] {
		sb.WriteString(">")
		return
	}

	sb.WriteString(">")
	for _, child := range n.Children {
		writeNodeHTML(sb, child)
	}
	sb.WriteString("</")
	sb.WriteString(n.Data)
	sb.WriteString(">")
}

func (n *Node) TagName() string {
	return strings.ToUpper(n.Data)
}

func (n *Node) ID() string {
	return n.GetAttribute("id")
}

func (n *Node) ClassName() string {
	return n.GetAttribute("class")
}

func (d *Document) Body() *Node {
	return QuerySelector(d.Root, "body")
}

func (d *Document) Head() *Node {
	return QuerySelector(d.Root, "head")
}

func (d *Document) Title() string {
	titleNode := QuerySelector(d.Root, "title")
	if titleNode != nil {
		return titleNode.TextContent()
	}
	return ""
}
