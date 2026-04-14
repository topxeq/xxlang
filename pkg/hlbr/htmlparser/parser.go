package htmlparser

import (
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

func Parse(html string) *dom.Document {
	lexer := NewLexer(html)
	doc := dom.NewDocument()
	root := &dom.Node{Type: dom.ElementNode, Data: "html"}
	doc.AppendChild(root)

	var stack []*dom.Node
	stack = append(stack, root)

	for {
		token := lexer.NextToken()
		if token.Type == TextToken && token.Data == "" {
			break
		}

		current := stack[len(stack)-1]

		switch token.Type {
		case TextToken:
			text := strings.TrimSpace(token.Data)
			if text != "" {
				current.AppendChild(&dom.Node{
					Type: dom.TextNode,
					Data: token.Data,
				})
			}

		case StartTagToken:
			node := &dom.Node{
				Type: dom.ElementNode,
				Data: token.Data,
			}
			for _, attr := range token.Attr {
				node.SetAttribute(attr.Key, attr.Value)
			}
			current.AppendChild(node)

			if !token.SelfClosed {
				stack = append(stack, node)
			}

			if token.Data == "style" || token.Data == "script" {
				rawText := lexer.ReadRawText(token.Data)
				if rawText != "" {
					node.AppendChild(&dom.Node{
						Type: dom.TextNode,
						Data: rawText,
					})
				}
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
			}

		case EndTagToken:
			for i := len(stack) - 1; i > 0; i-- {
				if stack[i].Data == token.Data {
					stack = stack[:i]
					break
				}
			}

		case CommentToken, DoctypeToken:
			// skip
		}
	}

	return doc
}
