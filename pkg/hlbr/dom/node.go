package dom

import "strings"

type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
	CommentNode
	DocumentNode
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

	if voidElements[n.Data] || len(n.Children) == 0 {
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
