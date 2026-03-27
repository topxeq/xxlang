// pkg/stdlib/html.go
// HTML module for Xxlang - HTML document handling.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "html",
		Exports: map[string]objects.Object{
			// parse parses an HTML string and returns an HTMLDocument.
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				doc, err := objects.ParseHTML(str.Value)
				if err != nil {
					return Error("failed to parse HTML: " + err.Error())
				}
				return doc
			}),

			// parseFile parses an HTML file and returns an HTMLDocument.
			"parseFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseFile takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("path must be a string")
				}
				doc, err := objects.ParseHTMLFile(path.Value)
				if err != nil {
					return Error("failed to parse HTML file: " + err.Error())
				}
				return doc
			}),

			// parseFragment parses an HTML fragment and returns an array of elements.
			"parseFragment": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseFragment takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				elements, err := objects.ParseHTMLFragment(str.Value)
				if err != nil {
					return Error("failed to parse HTML fragment: " + err.Error())
				}
				result := make([]objects.Object, len(elements))
				for i, elem := range elements {
					result[i] = elem
				}
				return &objects.Array{Elements: result}
			}),

			// newDocument creates a new HTML document with html, head, and body elements.
			"newDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewHTMLDocument()
			}),

			// newDocumentWithTitle creates a new HTML document with a title.
			"newDocumentWithTitle": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newDocumentWithTitle takes exactly 1 argument")
				}
				title, ok := args[0].(*objects.String)
				if !ok {
					return Error("title must be a string")
				}
				return objects.NewHTMLDocumentWithTitle(title.Value)
			}),

			// newElement creates a new HTML element with the given tag name.
			"newElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newElement takes exactly 1 argument")
				}
				tagName, ok := args[0].(*objects.String)
				if !ok {
					return Error("tagName must be a string")
				}
				return objects.NewHTMLElement(tagName.Value)
			}),

			// newTextNode creates a new text node.
			"newTextNode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newTextNode takes exactly 1 argument")
				}
				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("text must be a string")
				}
				return objects.NewHTMLTextNode(text.Value)
			}),

			// newComment creates a new comment node.
			"newComment": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newComment takes exactly 1 argument")
				}
				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("text must be a string")
				}
				return objects.NewHTMLComment(text.Value)
			}),

			// escape escapes special characters for HTML content.
			"escape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escape takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				escaped := objects.EscapeHTML(str.Value)
				return String(escaped)
			}),

			// escapeAttr escapes special characters for HTML attributes.
			"escapeAttr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escapeAttr takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				escaped := objects.EscapeHTMLAttr(str.Value)
				return String(escaped)
			}),

			// unescape unescapes HTML entities.
			"unescape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unescape takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				unescaped := objects.UnescapeHTML(str.Value)
				return String(unescaped)
			}),

			// stripTags removes all HTML tags from a string.
			"stripTags": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stripTags takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				stripped := objects.StripTags(str.Value)
				return String(stripped)
			}),

			// sanitize removes potentially dangerous HTML content.
			"sanitize": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sanitize takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				sanitized := objects.SanitizeHTML(str.Value)
				return String(sanitized)
			}),

			// isHTMLDocument checks if an object is an HTMLDocument.
			"isHTMLDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isHTMLDocument takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.HTMLDocument)
				return Bool(ok)
			}),

			// isHTMLElement checks if an object is an HTMLElement.
			"isHTMLElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isHTMLElement takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.HTMLElement)
				return Bool(ok)
			}),

			// encode converts an Object to HTML string.
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encode takes at least 1 argument")
				}
				rootName := "div"
				if len(args) >= 2 {
					if rn, ok := args[1].(*objects.String); ok {
						rootName = rn.Value
					}
				}
				result, err := objects.EncodeToHTML(args[0], rootName)
				if err != nil {
					return Error("encode failed: " + err.Error())
				}
				return String(result)
			}),

			// createElement is an alias for newElement.
			"createElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("createElement takes exactly 1 argument")
				}
				tagName, ok := args[0].(*objects.String)
				if !ok {
					return Error("tagName must be a string")
				}
				return objects.NewHTMLElement(tagName.Value)
			}),

			// createTextNode is an alias for newTextNode.
			"createTextNode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("createTextNode takes exactly 1 argument")
				}
				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("text must be a string")
				}
				return objects.NewHTMLTextNode(text.Value)
			}),
		},
	})
}
