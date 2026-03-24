// pkg/stdlib/xml.go
// XML utilities for the Xxlang standard library.
// This module provides comprehensive XML parsing, serialization, and manipulation.
package stdlib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

// xmlElementToMap converts an xml.StartElement and its content to an Xxlang map
func xmlElementToMap(decoder *xml.Decoder, start xml.StartElement) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Add attributes with @ prefix
	attrs := make(map[string]interface{})
	for _, attr := range start.Attr {
		attrs[attr.Name.Local] = attr.Value
	}
	if len(attrs) > 0 {
		result["@attributes"] = attrs
	}

	// Collect child elements and text
	var textContent strings.Builder
	children := make(map[string][]interface{})

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Recursively parse child element
			child, err := xmlElementToMap(decoder, t)
			if err != nil {
				return nil, err
			}
			// Add to children map (support multiple elements with same name)
			name := t.Name.Local
			children[name] = append(children[name], child)

		case xml.EndElement:
			// End of current element
			if t.Name.Local == start.Name.Local {
				goto done
			}

		case xml.CharData:
			// Text content
			text := strings.TrimSpace(string(t))
			if text != "" {
				textContent.WriteString(text)
			}

		case xml.Comment:
			// Ignore comments

		case xml.ProcInst:
			// Processing instruction (like <?xml version="1.0"?>)
			if t.Target == "xml" {
				result["@declaration"] = string(t.Inst)
			}

		case xml.Directive:
			// Ignore directives (like <!DOCTYPE>)
		}
	}

done:
	// Add text content
	text := strings.TrimSpace(textContent.String())
	if text != "" {
		result["@text"] = text
	}

	// Add children
	for name, elements := range children {
		if len(elements) == 1 {
			result[name] = elements[0]
		} else {
			result[name] = elements
		}
	}

	return result, nil
}

// mapToXML converts an Xxlang map to XML bytes
func mapToXML(name string, data map[string]interface{}, indent string, level int) ([]byte, error) {
	var buf bytes.Buffer
	prefix := strings.Repeat(indent, level)

	// Start element
	buf.WriteString(prefix)
	buf.WriteString("<")
	buf.WriteString(name)

	// Add attributes
	if attrs, ok := data["@attributes"]; ok {
		if attrMap, ok := attrs.(map[string]interface{}); ok {
			for attrName, attrValue := range attrMap {
				buf.WriteString(fmt.Sprintf(` %s="%v"`, attrName, attrValue))
			}
		}
	}

	// Check if self-closing (no content)
	_, hasText := data["@text"]
	hasChildren := false
	children := make(map[string][]interface{})
	for k, v := range data {
		if strings.HasPrefix(k, "@") {
			continue
		}
		if arr, ok := v.([]interface{}); ok {
			children[k] = arr
			hasChildren = true
		} else if m, ok := v.(map[string]interface{}); ok {
			children[k] = []interface{}{m}
			hasChildren = true
		} else {
			children[k] = []interface{}{v}
			hasChildren = true
		}
	}

	if !hasText && !hasChildren {
		buf.WriteString("/>")
		if indent != "" {
			buf.WriteString("\n")
		}
		return buf.Bytes(), nil
	}

	buf.WriteString(">")

	// Add text content
	if hasText {
		if text, ok := data["@text"].(string); ok {
			buf.WriteString(text)
		}
	}

	// Add children
	if hasChildren {
		if indent != "" {
			buf.WriteString("\n")
		}
		for childName, childElements := range children {
			for _, child := range childElements {
				switch c := child.(type) {
				case map[string]interface{}:
					childBytes, err := mapToXML(childName, c, indent, level+1)
					if err != nil {
						return nil, err
					}
					buf.Write(childBytes)
				case string:
					buf.WriteString(strings.Repeat(indent, level+1))
					buf.WriteString(fmt.Sprintf("<%s>%s</%s>", childName, c, childName))
					if indent != "" {
						buf.WriteString("\n")
					}
				case int, int64, float64, bool:
					buf.WriteString(strings.Repeat(indent, level+1))
					buf.WriteString(fmt.Sprintf("<%s>%v</%s>", childName, c, childName))
					if indent != "" {
						buf.WriteString("\n")
					}
				}
			}
		}
	}

	// End element
	if hasChildren && indent != "" {
		buf.WriteString(prefix)
	}
	buf.WriteString(fmt.Sprintf("</%s>", name))
	if indent != "" {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// objectToXMLMap converts an Xxlang object to a Go map for XML conversion
func objectToXMLMap(obj objects.Object) (map[string]interface{}, error) {
	switch o := obj.(type) {
	case *objects.Map:
		result := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, _ := pair.Key.(*objects.String)
			if key == nil {
				continue
			}
			val, err := objects.ObjectToGoValue(pair.Value)
			if err != nil {
				return nil, err
			}
			result[key.Value] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected map, got %s", obj.Type())
	}
}

func init() {
	Register(&Module{
		Name: "xml",
		Exports: map[string]objects.Object{
			// ============================================================
			// Core XML functions
			// ============================================================

			// parse parses an XML string and returns an Xxlang object.
			// Usage: obj = xml.parse(xmlStr)
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}

				decoder := xml.NewDecoder(strings.NewReader(s.Value))
				var result map[string]interface{}

				for {
					token, err := decoder.Token()
					if err != nil {
						if err == io.EOF {
							break
						}
						return Error(fmt.Sprintf("parse() failed: %s", err.Error()))
					}

					switch t := token.(type) {
					case xml.StartElement:
						// Parse root element
						element, err := xmlElementToMap(decoder, t)
						if err != nil {
							return Error(fmt.Sprintf("parse() failed: %s", err.Error()))
						}
						// Wrap in map with element name as key
						result = map[string]interface{}{
							t.Name.Local: element,
						}
					case xml.ProcInst:
						// Handle XML declaration
						if t.Target == "xml" {
							if result == nil {
								result = make(map[string]interface{})
							}
							result["@declaration"] = string(t.Inst)
						}
					}
				}

				if result == nil {
					return Error("parse() failed: no valid XML content")
				}

				return objects.GoValueToObject(result)
			}),

			// stringify converts an Xxlang object to an XML string.
			// Usage: str = xml.stringify(rootName, obj, [indent])
			"stringify": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("stringify() takes at least 2 arguments: rootName, obj, [indent]")
				}

				rootName, ok := args[0].(*objects.String)
				if !ok {
					return Error("stringify() first argument must be a string (root element name)")
				}

				data, err := objectToXMLMap(args[1])
				if err != nil {
					return Error(fmt.Sprintf("stringify() failed: %s", err.Error()))
				}

				indent := "  "
				if len(args) > 2 {
					if indentStr, ok := args[2].(*objects.String); ok {
						indent = indentStr.Value
					} else if indentInt, ok := args[2].(*objects.Int); ok {
						indent = strings.Repeat(" ", int(indentInt.Value))
					}
				}

				xmlBytes, err := mapToXML(rootName.Value, data, indent, 0)
				if err != nil {
					return Error(fmt.Sprintf("stringify() failed: %s", err.Error()))
				}

				return String(string(xmlBytes))
			}),

			// encode converts an Xxlang object to a compact XML string.
			// Usage: str = xml.encode(rootName, obj)
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("encode() takes exactly 2 arguments: rootName, obj")
				}

				rootName, ok := args[0].(*objects.String)
				if !ok {
					return Error("encode() first argument must be a string (root element name)")
				}

				data, err := objectToXMLMap(args[1])
				if err != nil {
					return Error(fmt.Sprintf("encode() failed: %s", err.Error()))
				}

				xmlBytes, err := mapToXML(rootName.Value, data, "", 0)
				if err != nil {
					return Error(fmt.Sprintf("encode() failed: %s", err.Error()))
				}

				return String(string(xmlBytes))
			}),

			// decode parses an XML string and returns an Xxlang object.
			// Alias for parse.
			"decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("decode() requires a string argument")
				}

				decoder := xml.NewDecoder(strings.NewReader(s.Value))
				var result map[string]interface{}

				for {
					token, err := decoder.Token()
					if err != nil {
						if err == io.EOF {
							break
						}
						return Error(fmt.Sprintf("decode() failed: %s", err.Error()))
					}

					switch t := token.(type) {
					case xml.StartElement:
						element, err := xmlElementToMap(decoder, t)
						if err != nil {
							return Error(fmt.Sprintf("decode() failed: %s", err.Error()))
						}
						result = map[string]interface{}{
							t.Name.Local: element,
						}
					case xml.ProcInst:
						if t.Target == "xml" {
							if result == nil {
								result = make(map[string]interface{})
							}
							result["@declaration"] = string(t.Inst)
						}
					}
				}

				if result == nil {
					return Error("decode() failed: no valid XML content")
				}

				return objects.GoValueToObject(result)
			}),

			// ============================================================
			// File operations
			// ============================================================

			// readFile reads an XML file and parses it.
			// Usage: obj = xml.readFile(path)
			"readFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readFile() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("readFile() requires a string path")
				}

				content, err := os.ReadFile(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("readFile() failed: %s", err.Error()))
				}

				decoder := xml.NewDecoder(bytes.NewReader(content))
				var result map[string]interface{}

				for {
					token, err := decoder.Token()
					if err != nil {
						if err == io.EOF {
							break
						}
						return Error(fmt.Sprintf("readFile() parse failed: %s", err.Error()))
					}

					switch t := token.(type) {
					case xml.StartElement:
						element, err := xmlElementToMap(decoder, t)
						if err != nil {
							return Error(fmt.Sprintf("readFile() parse failed: %s", err.Error()))
						}
						result = map[string]interface{}{
							t.Name.Local: element,
						}
					case xml.ProcInst:
						if t.Target == "xml" {
							if result == nil {
								result = make(map[string]interface{})
							}
							result["@declaration"] = string(t.Inst)
						}
					}
				}

				if result == nil {
					return Error("readFile() failed: no valid XML content")
				}

				return objects.GoValueToObject(result)
			}),

			// writeFile writes an Xxlang object to an XML file.
			// Usage: xml.writeFile(path, rootName, obj, [indent])
			"writeFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("writeFile() takes at least 3 arguments: path, rootName, obj, [indent]")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFile() first argument must be a string path")
				}

				rootName, ok := args[1].(*objects.String)
				if !ok {
					return Error("writeFile() second argument must be a string (root element name)")
				}

				data, err := objectToXMLMap(args[2])
				if err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				indent := "  "
				if len(args) > 3 {
					if indentStr, ok := args[3].(*objects.String); ok {
						indent = indentStr.Value
					} else if indentInt, ok := args[3].(*objects.Int); ok {
						indent = strings.Repeat(" ", int(indentInt.Value))
					}
				}

				xmlBytes, err := mapToXML(rootName.Value, data, indent, 0)
				if err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				// Add XML declaration
				header := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
				fullContent := header + string(xmlBytes)

				err = os.WriteFile(path.Value, []byte(fullContent), 0644)
				if err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				return Null()
			}),

			// ============================================================
			// Element extraction
			// ============================================================

			// getAttr gets an attribute value from an XML element map.
			// Usage: value = xml.getAttr(element, attrName)
			"getAttr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("getAttr() takes exactly 2 arguments")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("getAttr() first argument must be a map (XML element)")
				}

				attrName, ok := args[1].(*objects.String)
				if !ok {
					return Error("getAttr() second argument must be a string (attribute name)")
				}

				// Get @attributes map
				attrsKey := objects.NewString("@attributes")
				attrsPair, found := element.Pairs[attrsKey.HashKey()]
				if !found {
					return Null()
				}

				attrs, ok := attrsPair.Value.(*objects.Map)
				if !ok {
					return Null()
				}

				// Get the attribute
				attrKey := objects.NewString(attrName.Value)
				attrPair, found := attrs.Pairs[attrKey.HashKey()]
				if !found {
					return Null()
				}

				return attrPair.Value
			}),

			// getText gets the text content from an XML element map.
			// Usage: text = xml.getText(element)
			"getText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getText() takes exactly 1 argument")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("getText() argument must be a map (XML element)")
				}

				textKey := objects.NewString("@text")
				textPair, found := element.Pairs[textKey.HashKey()]
				if !found {
					return String("")
				}

				if text, ok := textPair.Value.(*objects.String); ok {
					return text
				}

				return String("")
			}),

			// getElement gets a child element from an XML element map.
			// Usage: child = xml.getElement(element, name)
			"getElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("getElement() takes exactly 2 arguments")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("getElement() first argument must be a map (XML element)")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("getElement() second argument must be a string (element name)")
				}

				key := objects.NewString(name.Value)
				pair, found := element.Pairs[key.HashKey()]
				if !found {
					return Null()
				}
				return pair.Value
			}),

			// getElements gets all child elements with a given name (for repeated elements).
			// Usage: children = xml.getElements(element, name)
			"getElements": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("getElements() takes exactly 2 arguments")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("getElements() first argument must be a map (XML element)")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("getElements() second argument must be a string (element name)")
				}

				key := objects.NewString(name.Value)
				pair, found := element.Pairs[key.HashKey()]
				if !found {
					return Array()
				}

				obj := pair.Value

				// If already an array, return it
				if arr, ok := obj.(*objects.Array); ok {
					return arr
				}

				// If a single element, wrap in array
				return Array(obj)
			}),

			// ============================================================
			// XML validation and utilities
			// ============================================================

			// isValid checks if a string is valid XML.
			// Usage: valid = xml.isValid(xmlStr)
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument")
				}

				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isValid() requires a string argument")
				}

				decoder := xml.NewDecoder(strings.NewReader(s.Value))
				for {
					_, err := decoder.Token()
					if err != nil {
						if err == io.EOF {
							return Bool(true)
						}
						return Bool(false)
					}
				}
			}),

			// escape escapes special XML characters.
			// Usage: escaped = xml.escape(str)
			"escape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escape() takes exactly 1 argument")
				}

				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("escape() requires a string argument")
				}

				var buf bytes.Buffer
				xml.Escape(&buf, []byte(s.Value))
				return String(buf.String())
			}),

			// unescape unescapes XML entities.
			// Usage: unescaped = xml.unescape(str)
			"unescape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unescape() takes exactly 1 argument")
				}

				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("unescape() requires a string argument")
				}

				var buf bytes.Buffer
				xml.EscapeText(&buf, []byte(s.Value))
				return String(buf.String())
			}),

			// setAttr sets an attribute on an XML element map.
			// Usage: element = xml.setAttr(element, name, value)
			"setAttr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("setAttr() takes exactly 3 arguments")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("setAttr() first argument must be a map (XML element)")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("setAttr() second argument must be a string (attribute name)")
				}

				// Get or create @attributes map
				attrsKey := objects.NewString("@attributes")
				attrsPair, found := element.Pairs[attrsKey.HashKey()]
				var attrs *objects.Map
				if !found {
					attrs = objects.NewMap(make(map[objects.HashKey]objects.MapPair))
				} else {
					attrs, ok = attrsPair.Value.(*objects.Map)
					if !ok {
						attrs = objects.NewMap(make(map[objects.HashKey]objects.MapPair))
					}
				}

				// Set the attribute (create a new map with the attribute)
				// Since maps are immutable, we need to create a copy
				newPairs := make(map[objects.HashKey]objects.MapPair)
				for _, pair := range element.Pairs {
					newPairs[pair.Key.HashKey()] = pair
				}

				// Update @attributes
				newAttrsPairs := make(map[objects.HashKey]objects.MapPair)
				for _, pair := range attrs.Pairs {
					newAttrsPairs[pair.Key.HashKey()] = pair
				}
				newAttrsPairs[objects.NewString(name.Value).HashKey()] = objects.MapPair{
					Key:   objects.NewString(name.Value),
					Value: args[2],
				}

				newPairs[objects.NewString("@attributes").HashKey()] = objects.MapPair{
					Key:   objects.NewString("@attributes"),
					Value: objects.NewMap(newAttrsPairs),
				}

				return objects.NewMap(newPairs)
			}),

			// setText sets the text content of an XML element map.
			// Usage: element = xml.setText(element, text)
			"setText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setText() takes exactly 2 arguments")
				}

				element, ok := args[0].(*objects.Map)
				if !ok {
					return Error("setText() first argument must be a map (XML element)")
				}

				text, ok := args[1].(*objects.String)
				if !ok {
					return Error("setText() second argument must be a string")
				}

				// Create a copy with updated @text
				newPairs := make(map[objects.HashKey]objects.MapPair)
				for _, pair := range element.Pairs {
					newPairs[pair.Key.HashKey()] = pair
				}

				newPairs[objects.NewString("@text").HashKey()] = objects.MapPair{
					Key:   objects.NewString("@text"),
					Value: text,
				}

				return objects.NewMap(newPairs)
			}),

			// addElement adds a child element to an XML element map.
			// Usage: element = xml.addElement(parent, name, child)
			"addElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("addElement() takes exactly 3 arguments")
				}

				parent, ok := args[0].(*objects.Map)
				if !ok {
					return Error("addElement() first argument must be a map (XML element)")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("addElement() second argument must be a string (element name)")
				}

				// Create a copy with the new child
				newPairs := make(map[objects.HashKey]objects.MapPair)
				for _, pair := range parent.Pairs {
					newPairs[pair.Key.HashKey()] = pair
				}

				newPairs[objects.NewString(name.Value).HashKey()] = objects.MapPair{
					Key:   objects.NewString(name.Value),
					Value: args[2],
				}

				return objects.NewMap(newPairs)
			}),

			// newElement creates a new XML element map.
			// Usage: element = xml.newElement(name, [text], [attributes])
			"newElement": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("newElement() takes at least 1 argument")
				}

				// Just return an empty map with optional text and attributes
				result := make(map[objects.HashKey]objects.MapPair)

				if len(args) > 1 {
					if text, ok := args[1].(*objects.String); ok {
						result[objects.NewString("@text").HashKey()] = objects.MapPair{
							Key:   objects.NewString("@text"),
							Value: text,
						}
					}
				}

				if len(args) > 2 {
					if attrs, ok := args[2].(*objects.Map); ok {
						result[objects.NewString("@attributes").HashKey()] = objects.MapPair{
							Key:   objects.NewString("@attributes"),
							Value: attrs,
						}
					}
				}

				return objects.NewMap(result)
			}),
		},
	})
}
