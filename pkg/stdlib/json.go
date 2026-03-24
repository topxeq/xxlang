// pkg/stdlib/json.go
// JSON utilities for the Xxlang standard library.
// This module provides comprehensive JSON parsing, serialization, and file operations.
// The builtin functions toJson and fromJson are aliases for json.stringify and json.parse respectively.
package stdlib

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "json",
		Exports: map[string]objects.Object{
			// ============================================================
			// Core JSON functions
			// ============================================================

			// parse parses a JSON string and returns an Xxlang object.
			// Usage: obj = json.parse(jsonStr)
			// This is also available as the builtin function fromJson.
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}

				obj, err := objects.JSONToObject(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("parse() failed: %s", err.Error()))
				}

				return obj
			}),

			// stringify converts an Xxlang object to a JSON string.
			// Usage: str = json.stringify(obj, [indent])
			// indent can be a string (like "  ") or an integer (number of spaces).
			// This is also available as the builtin function toJson.
			"stringify": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stringify() takes at least 1 argument")
				}

				indent := ""
				if len(args) > 1 {
					indentStr, ok := args[1].(*objects.String)
					if !ok {
						// Check if it's an integer for number of spaces
						if indentInt, ok := args[1].(*objects.Int); ok {
							for i := int64(0); i < indentInt.Value; i++ {
								indent += " "
							}
						} else {
							return Error("stringify() second argument must be a string or integer")
						}
					} else {
						indent = indentStr.Value
					}
				}

				goValue, err := objects.ObjectToGoValue(args[0])
				if err != nil {
					return Error(fmt.Sprintf("stringify() failed: %s", err.Error()))
				}

				var bytes []byte
				var marshalErr error
				if indent != "" {
					bytes, marshalErr = json.MarshalIndent(goValue, "", indent)
				} else {
					bytes, marshalErr = json.Marshal(goValue)
				}
				if marshalErr != nil {
					return Error(fmt.Sprintf("stringify() failed: %s", marshalErr.Error()))
				}

				return String(string(bytes))
			}),

			// encode converts an Xxlang object to a compact JSON string.
			// Alias for stringify without indent.
			// Usage: str = json.encode(obj)
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encode() takes at least 1 argument")
				}

				goValue, err := objects.ObjectToGoValue(args[0])
				if err != nil {
					return Error(fmt.Sprintf("encode() failed: %s", err.Error()))
				}

				bytes, marshalErr := json.Marshal(goValue)
				if marshalErr != nil {
					return Error(fmt.Sprintf("encode() failed: %s", marshalErr.Error()))
				}

				return String(string(bytes))
			}),

			// decode parses a JSON string and returns an Xxlang object.
			// Alias for parse.
			// Usage: obj = json.decode(jsonStr)
			"decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("decode() requires a string argument")
				}

				obj, err := objects.JSONToObject(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("decode() failed: %s", err.Error()))
				}

				return obj
			}),

			// ============================================================
			// Aliases for builtin functions
			// toJson is an alias for stringify, fromJson is an alias for parse
			// ============================================================

			// toJson converts an Xxlang object to a JSON string.
			// This is an alias for stringify and matches the builtin function toJson.
			// Usage: str = json.toJson(obj, [indent])
			"toJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("toJson() takes at least 1 argument")
				}

				// Check for options like "-indent" and "-sort"
				indent := false
				sortKeys := false
				indentStr := ""

				for i := 1; i < len(args); i++ {
					if str, ok := args[i].(*objects.String); ok {
						switch str.Value {
						case "-indent":
							indent = true
							indentStr = "  "
						case "-sort":
							sortKeys = true
						default:
							// Could be indent string
							if indent {
								indentStr = str.Value
							}
						}
					}
				}

				jsonBytes, err := objects.ObjectToJSON(args[0], objects.ObjectToJSONOptions{
					Indent:    indent,
					SortKeys:  sortKeys,
					IndentStr: indentStr,
				})
				if err != nil {
					return Error(fmt.Sprintf("toJson() failed: %s", err.Error()))
				}

				return String(string(jsonBytes))
			}),

			// fromJson parses a JSON string and returns an Xxlang object.
			// This is an alias for parse and matches the builtin function fromJson.
			// Usage: obj = json.fromJson(jsonStr)
			"fromJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromJson() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("fromJson() requires a string argument")
				}

				obj, err := objects.JSONToObject(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("fromJson() failed: %s", err.Error()))
				}

				return obj
			}),

			// ============================================================
			// File-based JSON operations
			// ============================================================

			// readFile reads a JSON file and parses it.
			// Usage: data = json.readFile(path)
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

				obj, err := objects.JSONToObject(string(content))
				if err != nil {
					return Error(fmt.Sprintf("readFile() parse failed: %s", err.Error()))
				}

				return obj
			}),

			// writeFile writes an object to a JSON file.
			// Usage: json.writeFile(path, data, [indent])
			"writeFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("writeFile() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFile() requires a string path")
				}

				goValue, err := objects.ObjectToGoValue(args[1])
				if err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				var bytes []byte
				var marshalErr error
				if len(args) == 3 {
					// Pretty print with indent
					indent := "  "
					if indentStr, ok := args[2].(*objects.String); ok {
						indent = indentStr.Value
					} else if indentInt, ok := args[2].(*objects.Int); ok {
						indent = ""
						for i := int64(0); i < indentInt.Value; i++ {
							indent += " "
						}
					}
					bytes, marshalErr = json.MarshalIndent(goValue, "", indent)
				} else {
					bytes, marshalErr = json.Marshal(goValue)
				}

				if marshalErr != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", marshalErr.Error()))
				}

				if err := os.WriteFile(path.Value, bytes, 0644); err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				return Null()
			}),

			// writeFilePretty writes an object to a JSON file with pretty formatting.
			// Usage: json.writeFilePretty(path, data, [indent])
			"writeFilePretty": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("writeFilePretty() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFilePretty() requires a string path")
				}

				goValue, err := objects.ObjectToGoValue(args[1])
				if err != nil {
					return Error(fmt.Sprintf("writeFilePretty() failed: %s", err.Error()))
				}

				indent := "  "
				if len(args) == 3 {
					if indentStr, ok := args[2].(*objects.String); ok {
						indent = indentStr.Value
					}
				}

				bytes, marshalErr := json.MarshalIndent(goValue, "", indent)
				if marshalErr != nil {
					return Error(fmt.Sprintf("writeFilePretty() failed: %s", marshalErr.Error()))
				}

				if err := os.WriteFile(path.Value, bytes, 0644); err != nil {
					return Error(fmt.Sprintf("writeFilePretty() failed: %s", err.Error()))
				}

				return Null()
			}),

			// updateFile reads a JSON file, updates it with new values, and writes it back.
			// Usage: json.updateFile(path, updates)
			"updateFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("updateFile() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("updateFile() requires a string path")
				}
				updates, ok := args[1].(*objects.Map)
				if !ok {
					return Error("updateFile() requires a map for updates")
				}

				// Read existing file
				content, err := os.ReadFile(path.Value)
				if err != nil {
					return Error(fmt.Sprintf("updateFile() read failed: %s", err.Error()))
				}

				// Parse existing content
				var existing map[string]interface{}
				if err := json.Unmarshal(content, &existing); err != nil {
					// If file is empty or invalid, start with empty map
					existing = make(map[string]interface{})
				}

				// Apply updates
				for _, pair := range updates.Pairs {
					key, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}
					goVal, err := objects.ObjectToGoValue(pair.Value)
					if err != nil {
						continue
					}
					existing[key.Value] = goVal
				}

				// Write back
				bytes, marshalErr := json.MarshalIndent(existing, "", "  ")
				if marshalErr != nil {
					return Error(fmt.Sprintf("updateFile() marshal failed: %s", marshalErr.Error()))
				}

				if err := os.WriteFile(path.Value, bytes, 0644); err != nil {
					return Error(fmt.Sprintf("updateFile() write failed: %s", err.Error()))
				}

				return Null()
			}),

			// appendToArrayFile appends an element to a JSON array file.
			// Usage: json.appendToArrayFile(path, element)
			"appendToArrayFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("appendToArrayFile() takes exactly 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("appendToArrayFile() requires a string path")
				}

				// Read existing file
				content, err := os.ReadFile(path.Value)
				var existing []interface{}
				if err != nil {
					// File doesn't exist, start with empty array
					existing = []interface{}{}
				} else {
					// Parse existing content
					if err := json.Unmarshal(content, &existing); err != nil {
						// If file doesn't contain an array, wrap it in an array
						var singleValue interface{}
						if unmarshalErr := json.Unmarshal(content, &singleValue); unmarshalErr == nil {
							existing = []interface{}{singleValue}
						} else {
							existing = []interface{}{}
						}
					}
				}

				// Convert new element
				newElem, convErr := objects.ObjectToGoValue(args[1])
				if convErr != nil {
					return Error(fmt.Sprintf("appendToArrayFile() failed: %s", convErr.Error()))
				}

				// Append
				existing = append(existing, newElem)

				// Write back
				bytes, marshalErr := json.MarshalIndent(existing, "", "  ")
				if marshalErr != nil {
					return Error(fmt.Sprintf("appendToArrayFile() marshal failed: %s", marshalErr.Error()))
				}

				if err := os.WriteFile(path.Value, bytes, 0644); err != nil {
					return Error(fmt.Sprintf("appendToArrayFile() write failed: %s", err.Error()))
				}

				return Null()
			}),

			// ============================================================
			// Additional utility functions
			// ============================================================

			// isValid checks if a string is valid JSON.
			// Usage: valid = json.isValid(jsonStr)
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isValid() requires a string argument")
				}

				var v interface{}
				err := json.Unmarshal([]byte(s.Value), &v)
				return Bool(err == nil)
			}),

			// getType returns the type of the JSON value at the root.
			// Returns: "object", "array", "string", "number", "boolean", "null", or "invalid".
			// Usage: type = json.getType(jsonStr)
			"getType": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getType() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("getType() requires a string argument")
				}

				var v interface{}
				if err := json.Unmarshal([]byte(s.Value), &v); err != nil {
					return String("invalid")
				}

				switch v.(type) {
				case map[string]interface{}:
					return String("object")
				case []interface{}:
					return String("array")
				case string:
					return String("string")
				case float64:
					return String("number")
				case bool:
					return String("boolean")
				case nil:
					return String("null")
				default:
					return String("unknown")
				}
			}),

			// ============================================================
			// JSONPath operations
			// ============================================================

			// get retrieves a value from an object using JSONPath.
			// Returns the first matching value, or null if not found.
			// Usage: value = json.get(path, obj)
			// Example: json.get("$.store.book[0].title", data)
			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("get() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("get() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("get() invalid path: %s", err.Error()))
				}

				results := path.Get(args[1])
				if len(results) == 0 {
					return Null()
				}

				return results[0]
			}),

			// getAll retrieves all values matching a JSONPath.
			// Returns an array of all matching values.
			// Usage: values = json.getAll(path, obj)
			// Example: json.getAll("$..author", data)
			"getAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("getAll() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("getAll() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("getAll() invalid path: %s", err.Error()))
				}

				results := path.Get(args[1])
				return Array(results...)
			}),

			// getWithPath retrieves all values matching a JSONPath with their paths.
			// Returns a map with path strings as keys and values.
			// Usage: result = json.getWithPath(path, obj)
			"getWithPath": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("getWithPath() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("getWithPath() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("getWithPath() invalid path: %s", err.Error()))
				}

				matches := path.GetWithPath(args[1])
				pairs := make(map[objects.HashKey]objects.MapPair)
				for _, m := range matches {
					key := String(m.Path)
					pairs[key.HashKey()] = objects.MapPair{
						Key:   key,
						Value: m.Value,
					}
				}
				return &objects.Map{Pairs: pairs}
			}),

			// set sets a value at the specified JSONPath.
			// Returns a new object with the value set (does not mutate original).
			// Usage: newObj = json.set(path, obj, value)
			// Example: json.set("$.store.book[0].price", data, 9.99)
			"set": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("set() takes at least 3 arguments: path, object, and value")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("set() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("set() invalid path: %s", err.Error()))
				}

				result, err := path.Set(args[1], args[2])
				if err != nil {
					return Error(fmt.Sprintf("set() failed: %s", err.Error()))
				}

				return result
			}),

			// delete removes values at the specified JSONPath.
			// Returns a new object with the value removed (does not mutate original).
			// Usage: newObj = json.delete(path, obj)
			// Example: json.delete("$.store.book[0]", data)
			"delete": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("delete() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("delete() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("delete() invalid path: %s", err.Error()))
				}

				result, err := path.Delete(args[1])
				if err != nil {
					return Error(fmt.Sprintf("delete() failed: %s", err.Error()))
				}

				return result
			}),

			// paths returns all JSONPath strings that exist in an object.
			// Useful for debugging and exploring JSON structure.
			// Usage: pathList = json.paths(obj)
			"paths": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("paths() takes at least 1 argument")
				}

				paths := Paths(args[0])
				elements := make([]objects.Object, len(paths))
				for i, p := range paths {
					elements[i] = String(p)
				}
				return Array(elements...)
			}),

			// has checks if a JSONPath exists in an object.
			// Returns true if at least one value matches the path.
			// Usage: exists = json.has(path, obj)
			"has": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("has() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("has() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Bool(false)
				}

				results := path.Get(args[1])
				return Bool(len(results) > 0)
			}),

			// count returns the number of values matching a JSONPath.
			// Usage: n = json.count(path, obj)
			"count": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("count() takes at least 2 arguments: path and object")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("count() first argument must be a string path")
				}

				path, err := ParseJSONPath(pathStr.Value)
				if err != nil {
					return Int(0)
				}

				results := path.Get(args[1])
				return Int(int64(len(results)))
			}),

			// query parses a JSON string and queries it with JSONPath.
			// Combines parse and get in one operation.
			// Usage: value = json.query(path, jsonStr)
			"query": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("query() takes at least 2 arguments: path and JSON string")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("query() first argument must be a string path")
				}

				jsonStr, ok := args[1].(*objects.String)
				if !ok {
					return Error("query() second argument must be a JSON string")
				}

				// Parse JSON
				obj, err := objects.JSONToObject(jsonStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("query() parse failed: %s", err.Error()))
				}

				// Query with path
				path, pathErr := ParseJSONPath(pathStr.Value)
				if pathErr != nil {
					return Error(fmt.Sprintf("query() invalid path: %s", pathErr.Error()))
				}

				results := path.Get(obj)
				if len(results) == 0 {
					return Null()
				}

				return results[0]
			}),

			// queryAll parses a JSON string and queries all matches with JSONPath.
			// Combines parse and getAll in one operation.
			// Usage: values = json.queryAll(path, jsonStr)
			"queryAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("queryAll() takes at least 2 arguments: path and JSON string")
				}

				pathStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("queryAll() first argument must be a string path")
				}

				jsonStr, ok := args[1].(*objects.String)
				if !ok {
					return Error("queryAll() second argument must be a JSON string")
				}

				// Parse JSON
				obj, err := objects.JSONToObject(jsonStr.Value)
				if err != nil {
					return Error(fmt.Sprintf("queryAll() parse failed: %s", err.Error()))
				}

				// Query with path
				path, pathErr := ParseJSONPath(pathStr.Value)
				if pathErr != nil {
					return Error(fmt.Sprintf("queryAll() invalid path: %s", pathErr.Error()))
				}

				results := path.Get(obj)
				return Array(results...)
			}),
		},
	})
}
