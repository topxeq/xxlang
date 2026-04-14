package renderer

import (
	"fmt"
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

type TextRenderer struct {
	width    int
	sb       strings.Builder
	col      int
	row      int
	maxWidth int
}

func NewTextRenderer(width int) *TextRenderer {
	return &TextRenderer{
		width:    width,
		maxWidth: 0,
	}
}

func (r *TextRenderer) Render(n *dom.Node) string {
	r.renderNode(n, 0)
	return r.sb.String()
}

func (r *TextRenderer) renderNode(n *dom.Node, depth int) {
	if n == nil {
		return
	}

	if n.Type == dom.ElementNode {
		tag := strings.ToLower(n.Data)
		if tag == "style" || tag == "script" || tag == "noscript" || tag == "link" || tag == "meta" || tag == "head" {
			return
		}
	}

	switch n.Type {
	case dom.TextNode:
		r.renderText(n.Data, depth)
	case dom.ElementNode:
		r.renderElement(n, depth)
	}

	for _, child := range n.Children {
		r.renderNode(child, depth)
	}
}

func (r *TextRenderer) renderText(text string, depth int) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	words := strings.Fields(text)
	for _, word := range words {
		if r.col+len(word) > r.width {
			r.sb.WriteString("\n")
			r.row++
			r.col = 0
		}
		if r.col > 0 {
			r.sb.WriteString(" ")
			r.col++
		}
		r.sb.WriteString(word)
		r.col += len(word)
		if r.col > r.maxWidth {
			r.maxWidth = r.col
		}
	}
}

func (r *TextRenderer) renderElement(n *dom.Node, depth int) {
	tag := strings.ToLower(n.Data)

	switch tag {
	case "br":
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "hr":
		r.ensureNewLine()
		for i := 0; i < r.width; i++ {
			r.sb.WriteString("-")
		}
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "p":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "h1", "h2", "h3", "h4", "h5", "h6":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
		for _, child := range n.Children {
			if child.Type == dom.TextNode {
				text := strings.TrimSpace(child.Data)
				if text != "" {
					r.sb.WriteString(strings.ToUpper(text))
					r.col += len(text)
				}
			}
		}
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "a":
		href := n.GetAttribute("href")
		text := n.TextContent()
		text = strings.TrimSpace(text)
		if text == "" {
			text = href
		}
		if r.col > 0 {
			r.sb.WriteString(" ")
			r.col++
		}
		r.sb.WriteString("[")
		r.sb.WriteString(text)
		if href != "" && href != "#" {
			r.sb.WriteString("](")
			r.sb.WriteString(href)
			r.sb.WriteString(")")
		}
		r.sb.WriteString("]")
		r.col += len(text) + 4
	case "img":
		alt := n.GetAttribute("alt")
		if alt == "" {
			alt = "[图片]"
		}
		if r.col > 0 {
			r.sb.WriteString(" ")
			r.col++
		}
		r.sb.WriteString(fmt.Sprintf("[图片: %s]", alt))
		r.col += len(alt) + 6
	case "li":
		r.ensureNewLine()
		r.sb.WriteString("  • ")
		r.col = 4
	case "ul", "ol":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "table":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "tr":
		r.ensureNewLine()
		r.col = 0
	case "td", "th":
		if r.col > 0 {
			r.sb.WriteString(" | ")
			r.col += 3
		}
	case "blockquote":
		r.ensureNewLine()
		r.sb.WriteString("> ")
		r.col = 2
	case "pre":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
		r.renderText(n.TextContent(), 0)
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "div":
		r.ensureNewLine()
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	case "span":
		// inline, no special handling
	case "strong", "b":
		if r.col > 0 {
			r.sb.WriteString(" ")
			r.col++
		}
		r.sb.WriteString("**")
		r.col += 2
	case "em", "i":
		if r.col > 0 {
			r.sb.WriteString(" ")
			r.col++
		}
		r.sb.WriteString("*")
		r.col++
	case "title":
		text := strings.TrimSpace(n.TextContent())
		if text != "" {
			r.ensureNewLine()
			r.sb.WriteString(strings.Repeat("=", r.width))
			r.sb.WriteString("\n")
			r.sb.WriteString(text)
			r.sb.WriteString("\n")
			r.sb.WriteString(strings.Repeat("=", r.width))
			r.sb.WriteString("\n\n")
			r.row += 4
			r.col = 0
		}
	case "style", "script", "noscript", "link", "meta", "head":
		return
	}

	if tag == "strong" || tag == "b" {
		defer func() {
			r.sb.WriteString("**")
			r.col += 2
		}()
	}
	if tag == "em" || tag == "i" {
		defer func() {
			r.sb.WriteString("*")
			r.col++
		}()
	}
}

func (r *TextRenderer) ensureNewLine() {
	if r.col > 0 {
		r.sb.WriteString("\n")
		r.row++
		r.col = 0
	}
}

func (r *TextRenderer) Width() int {
	return r.width
}

func (r *TextRenderer) Height() int {
	return r.row
}
