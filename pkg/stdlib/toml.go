// pkg/stdlib/toml.go
// TOML module for Xxlang - TOML document handling.
package stdlib

import (
	"os"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "toml",
		Exports: map[string]objects.Object{
			// parse parses a TOML string and returns a TomlDocument.
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				doc, err := objects.ParseToml(s.Value)
				if err != nil {
					return Error("failed to parse TOML: " + err.Error())
				}
				return doc
			}),

			// parseFile parses a TOML file and returns a TomlDocument.
			"parseFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseFile() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("path must be a string")
				}
				content, err := os.ReadFile(path.Value)
				if err != nil {
					return Error("failed to read file: " + err.Error())
				}
				doc, err := objects.ParseToml(string(content))
				if err != nil {
					return Error("failed to parse TOML: " + err.Error())
				}
				return doc
			}),

			// stringify converts a TomlDocument to a TOML string.
			"stringify": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stringify() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.TomlDocument)
				if !ok {
					return Error("argument must be a TomlDocument")
				}
				return String(doc.ToString())
			}),

			// encode converts an object to a TOML string.
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("encode() takes exactly 1 argument")
				}
				tomlValue := objects.FromXxlangObject(args[0])
				result := objects.EncodeToml(tomlValue)
				return String(result)
			}),

			// create creates a new empty TOML document.
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewTomlDocument()
			}),

			// isValid checks if a string is valid TOML.
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				valid, _ := objects.ValidateToml(s.Value)
				return Bool(valid)
			}),

			// isTomlDocument checks if an object is a TomlDocument.
			"isTomlDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isTomlDocument() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.TomlDocument)
				return Bool(ok)
			}),

			// fromJson converts a JSON string to TOML.
			"fromJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromJson() takes exactly 1 argument")
				}
				jsonStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				obj, err := objects.JSONToObject(jsonStr.Value)
				if err != nil {
					return Error("failed to parse JSON: " + err.Error())
				}
				tomlValue := objects.FromXxlangObject(obj)
				result := objects.EncodeToml(tomlValue)
				return String(result)
			}),

			// toJson converts a TomlDocument to JSON.
			"toJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("toJson() takes at least 1 argument")
				}

				var obj objects.Object
				switch v := args[0].(type) {
				case *objects.TomlDocument:
					obj = v.ToMap()
				case *objects.String:
					doc, err := objects.ParseToml(v.Value)
					if err != nil {
						return Error("failed to parse TOML: " + err.Error())
					}
					obj = doc.ToMap()
				default:
					obj = args[0]
				}

				jsonBytes, err := objects.ObjectToJSON(obj, objects.ObjectToJSONOptions{})
				if err != nil {
					return Error("failed to convert to JSON: " + err.Error())
				}
				return String(string(jsonBytes))
			}),
		},
	})
}
