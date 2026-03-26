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
		},
	})
}
