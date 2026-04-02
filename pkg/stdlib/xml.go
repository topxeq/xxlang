// pkg/stdlib/xml.go
// XML module for Xxlang - XML document handling.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "xml",
		Exports: map[string]objects.Object{
			// parse parses an XML string and returns an XMLDocument.
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				doc, err := objects.ParseXML(str.Value)
				if err != nil {
					return Error("failed to parse XML: " + err.Error())
				}
				return doc
			}),

			// parseFile parses an XML file and returns an XMLDocument.
			"parseFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseFile takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("path must be a string")
				}
				doc, err := objects.ParseXMLFile(path.Value)
				if err != nil {
					return Error("failed to parse XML file: " + err.Error())
				}
				return doc
			}),

			// create creates a new XML document with a root element.
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("create takes exactly 1 argument")
				}
				rootName, ok := args[0].(*objects.String)
				if !ok {
					return Error("root name must be a string")
				}
				return objects.NewXMLDocumentWithRoot(rootName.Value)
			}),

			// newDocument creates a new empty XML document.
			"newDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewXMLDocument()
			}),

			// newNode creates a new XML node.
			"newNode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newNode takes exactly 1 argument")
				}
				name, ok := args[0].(*objects.String)
				if !ok {
					return Error("name must be a string")
				}
				return objects.NewXMLNode(name.Value)
			}),

			// isXMLDocument checks if an object is an XMLDocument.
			"isXMLDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isXMLDocument takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.XMLDocument)
				return Bool(ok)
			}),

			// isXMLNode checks if an object is an XMLNode.
			"isXMLNode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isXMLNode takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.XMLNode)
				return Bool(ok)
			}),

			// encode converts an Object to XML string.
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encode takes at least 1 argument")
				}
				rootName := "root"
				if len(args) >= 2 {
					if rn, ok := args[1].(*objects.String); ok {
						rootName = rn.Value
					}
				}
				result, err := objects.EncodeToXML(args[0], rootName)
				if err != nil {
					return Error("encode failed: " + err.Error())
				}
				return String(result)
			}),

			// escape escapes special characters for XML.
			"escape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escape takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				escaped := objects.EscapeXMLText(str.Value)
				return String(escaped)
			}),

			// escapeAttr escapes special characters for XML attributes.
			"escapeAttr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escapeAttr takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				escaped := objects.EscapeXMLAttr(str.Value)
				return String(escaped)
			}),

			// setAttribute sets an attribute on an XML node.
			// Usage: xml.setAttribute(node, name, value)
			"setAttribute": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("setAttribute takes exactly 3 arguments")
				}
				node, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("first argument must be an XMLNode")
				}
				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (attribute name)")
				}
				value, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (attribute value)")
				}
				node.SetAttr(name.Value, value.Value)
				return Null()
			}),

			// getAttribute gets an attribute value from an XML node.
			// Usage: value = xml.getAttribute(node, name)
			"getAttribute": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("getAttribute takes exactly 2 arguments")
				}
				node, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("first argument must be an XMLNode")
				}
				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (attribute name)")
				}
				return String(node.Attr(name.Value))
			}),

			// addChild adds a child node to an XML node.
			// Usage: xml.addChild(parent, child)
			"addChild": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("addChild takes exactly 2 arguments")
				}
				parent, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("first argument must be an XMLNode")
				}
				child, ok := args[1].(*objects.XMLNode)
				if !ok {
					return Error("second argument must be an XMLNode")
				}
				parent.AddChild(child)
				return Null()
			}),

			// getChildren returns all child nodes of an XML node as an array.
			// Usage: children = xml.getChildren(node)
			"getChildren": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getChildren takes exactly 1 argument")
				}
				node, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("argument must be an XMLNode")
				}
				return node.Children()
			}),

			// setText sets the text content of an XML node.
			// Usage: xml.setText(node, text)
			"setText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setText takes exactly 2 arguments")
				}
				node, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("first argument must be an XMLNode")
				}
				text, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}
				node.SetText(text.Value)
				return Null()
			}),

			// getText gets the text content of an XML node.
			// Usage: text = xml.getText(node)
			"getText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getText takes exactly 1 argument")
				}
				node, ok := args[0].(*objects.XMLNode)
				if !ok {
					return Error("argument must be an XMLNode")
				}
				return String(node.Text())
			}),
		},
	})
}
