// pkg/stdlib/yaml.go
// YAML parsing and generation utilities for the Xxlang standard library.
// This module provides comprehensive YAML parsing, serialization, and file operations
// without using third-party libraries or CGO.
package stdlib

import (
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "yaml",
		Exports: map[string]objects.Object{
			// ============================================================
			// Core YAML functions
			// ============================================================

			// parse parses a YAML string and returns an Xxlang object.
			// Usage: obj = yaml.parse(yamlStr)
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}

				obj, err := objects.ParseYAML(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("parse() failed: %s", err.Error()))
				}

				return obj
			}),

			// stringify converts an Xxlang object to a YAML string.
			// Usage: str = yaml.stringify(obj, [indent])
			// indent defaults to 2 spaces.
			"stringify": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stringify() takes at least 1 argument")
				}

				indent := 2
				if len(args) > 1 {
					if indentInt, ok := args[1].(*objects.Int); ok {
						indent = int(indentInt.Value)
						if indent < 0 {
							indent = 0
						}
					}
				}

				return String(objects.SerializeYAML(args[0], indent))
			}),

			// encode converts an Xxlang object to a YAML string.
			// Alias for stringify with default indent.
			// Usage: str = yaml.encode(obj)
			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encode() takes at least 1 argument")
				}

				return String(objects.SerializeYAML(args[0], 2))
			}),

			// decode parses a YAML string and returns an Xxlang object.
			// Alias for parse.
			// Usage: obj = yaml.decode(yamlStr)
			"decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("decode() requires a string argument")
				}

				obj, err := objects.ParseYAML(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("decode() failed: %s", err.Error()))
				}

				return obj
			}),

			// ============================================================
			// File-based YAML operations
			// ============================================================

			// readFile reads a YAML file and parses it.
			// Usage: data = yaml.readFile(path)
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

				obj, err := objects.ParseYAML(string(content))
				if err != nil {
					return Error(fmt.Sprintf("readFile() parse failed: %s", err.Error()))
				}

				return obj
			}),

			// writeFile writes an object to a YAML file.
			// Usage: yaml.writeFile(path, data, [indent])
			"writeFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("writeFile() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFile() requires a string path")
				}

				indent := 2
				if len(args) == 3 {
					if indentInt, ok := args[2].(*objects.Int); ok {
						indent = int(indentInt.Value)
						if indent < 0 {
							indent = 0
						}
					}
				}

				yamlStr := objects.SerializeYAML(args[1], indent)

				if err := os.WriteFile(path.Value, []byte(yamlStr), 0644); err != nil {
					return Error(fmt.Sprintf("writeFile() failed: %s", err.Error()))
				}

				return Null()
			}),

			// appendFile appends YAML content to a file.
			// Usage: yaml.appendFile(path, data, [indent])
			"appendFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("appendFile() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("appendFile() requires a string path")
				}

				indent := 2
				if len(args) == 3 {
					if indentInt, ok := args[2].(*objects.Int); ok {
						indent = int(indentInt.Value)
						if indent < 0 {
							indent = 0
						}
					}
				}

				yamlStr := objects.SerializeYAML(args[1], indent)

				file, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return Error(fmt.Sprintf("appendFile() failed: %s", err.Error()))
				}
				defer file.Close()

				// Add document separator if file is not empty
				info, err := file.Stat()
				if err == nil && info.Size() > 0 {
					if _, err := file.WriteString("\n---\n"); err != nil {
						return Error(fmt.Sprintf("appendFile() failed: %s", err.Error()))
					}
				}

				if _, err := file.WriteString(yamlStr); err != nil {
					return Error(fmt.Sprintf("appendFile() failed: %s", err.Error()))
				}

				return Null()
			}),

			// updateFile reads a YAML file, updates it with new values, and writes it back.
			// Usage: yaml.updateFile(path, updates)
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
				existing, err := objects.ParseYAML(string(content))
				if err != nil {
					return Error(fmt.Sprintf("updateFile() parse failed: %s", err.Error()))
				}

				existingMap, ok := existing.(*objects.Map)
				if !ok {
					return Error("updateFile() root must be a map")
				}

				// Create new map with updates
				newPairs := make(map[objects.HashKey]objects.MapPair)
				for k, v := range existingMap.Pairs {
					newPairs[k] = v
				}

				// Apply updates
				for _, pair := range updates.Pairs {
					key, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}
					newPairs[key.HashKey()] = objects.MapPair{
						Key:   key,
						Value: pair.Value,
					}
				}

				// Write back
				result := &objects.Map{Pairs: newPairs}
				yamlStr := objects.SerializeYAML(result, 2)

				if err := os.WriteFile(path.Value, []byte(yamlStr), 0644); err != nil {
					return Error(fmt.Sprintf("updateFile() write failed: %s", err.Error()))
				}

				return Null()
			}),

			// ============================================================
			// Validation functions
			// ============================================================

			// isValid checks if a string is valid YAML.
			// Usage: valid = yaml.isValid(yamlStr)
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isValid() requires a string argument")
				}

				_, err := objects.ParseYAML(s.Value)
				return Bool(err == nil)
			}),

			// getType returns the type of the YAML value at the root.
			// Returns: "object", "array", "string", "number", "boolean", "null", or "invalid".
			// Usage: type = yaml.getType(yamlStr)
			"getType": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getType() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("getType() requires a string argument")
				}

				obj, err := objects.ParseYAML(s.Value)
				if err != nil {
					return String("invalid")
				}

				switch obj.(type) {
				case *objects.Map:
					return String("object")
				case *objects.Array:
					return String("array")
				case *objects.String:
					return String("string")
				case *objects.Int, *objects.Float:
					return String("number")
				case *objects.Bool:
					return String("boolean")
				case *objects.Null:
					return String("null")
				default:
					return String("unknown")
				}
			}),

			// ============================================================
			// Conversion utilities
			// ============================================================

			// fromJson converts a JSON string to YAML.
			// Usage: yamlStr = yaml.fromJson(jsonStr, [indent])
			"fromJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("fromJson() takes at least 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("fromJson() requires a string argument")
				}

				obj, err := objects.JSONToObject(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("fromJson() parse failed: %s", err.Error()))
				}

				indent := 2
				if len(args) > 1 {
					if indentInt, ok := args[1].(*objects.Int); ok {
						indent = int(indentInt.Value)
					}
				}

				return String(objects.SerializeYAML(obj, indent))
			}),

			// toJson converts a YAML string or object to JSON.
			// Usage: jsonStr = yaml.toJson(yamlStrOrObj, [indent])
			"toJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("toJson() takes at least 1 argument")
				}

				var obj objects.Object
				var err error

				// Check if argument is a string (YAML to parse) or an object (already parsed)
				if s, ok := args[0].(*objects.String); ok {
					obj, err = objects.ParseYAML(s.Value)
					if err != nil {
						return Error(fmt.Sprintf("toJson() parse failed: %s", err.Error()))
					}
				} else {
					// Assume it's already a parsed object
					obj = args[0]
				}

				indent := ""
				if len(args) > 1 {
					if indentInt, ok := args[1].(*objects.Int); ok {
						for i := int64(0); i < indentInt.Value; i++ {
							indent += " "
						}
					} else if indentStr, ok := args[1].(*objects.String); ok {
						indent = indentStr.Value
					}
				}

				jsonBytes, err := objects.ObjectToJSON(obj, objects.ObjectToJSONOptions{
					Indent:    indent != "",
					IndentStr: indent,
				})
				if err != nil {
					return Error(fmt.Sprintf("toJson() failed: %s", err.Error()))
				}

				return String(string(jsonBytes))
			}),

			// ============================================================
			// Merge utilities
			// ============================================================

			// merge merges multiple YAML maps into one.
			// Later maps override earlier ones for duplicate keys.
			// Usage: result = yaml.merge(map1, map2, ...)
			"merge": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("merge() takes at least 2 arguments")
				}

				result := make(map[objects.HashKey]objects.MapPair)

				for _, arg := range args {
					m, ok := arg.(*objects.Map)
					if !ok {
						return Error("merge() requires all arguments to be maps")
					}

					for _, pair := range m.Pairs {
						result[pair.Key.HashKey()] = pair
					}
				}

				return &objects.Map{Pairs: result}
			}),

			// deepMerge performs a deep merge of multiple YAML maps.
			// Nested maps are merged recursively.
			// Usage: result = yaml.deepMerge(map1, map2, ...)
			"deepMerge": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("deepMerge() takes at least 2 arguments")
				}

				result := make(map[objects.HashKey]objects.MapPair)

				// Start with the first map
				first, ok := args[0].(*objects.Map)
				if !ok {
					return Error("deepMerge() requires all arguments to be maps")
				}
				for k, v := range first.Pairs {
					result[k] = v
				}

				// Merge remaining maps
				for i := 1; i < len(args); i++ {
					next, ok := args[i].(*objects.Map)
					if !ok {
						return Error("deepMerge() requires all arguments to be maps")
					}

					for _, pair := range next.Pairs {
						existing, exists := result[pair.Key.HashKey()]
						if exists {
							// Check if both values are maps
							existingMap, existingIsMap := existing.Value.(*objects.Map)
							newMap, newIsMap := pair.Value.(*objects.Map)

							if existingIsMap && newIsMap {
								// Recursively merge
								merged := yamlDeepMergeMaps(existingMap, newMap)
								result[pair.Key.HashKey()] = objects.MapPair{
									Key:   pair.Key,
									Value: merged,
								}
								continue
							}
						}
						// Override with new value
						result[pair.Key.HashKey()] = pair
					}
				}

				return &objects.Map{Pairs: result}
			}),

			// ============================================================
			// Path-based access
			// ============================================================

			// get retrieves a value from an object using dot notation path.
			// Usage: value = yaml.get(obj, "path.to.key")
			// Example: yaml.get(data, "server.port")
			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("get() takes exactly 2 arguments")
				}

				path, ok := args[1].(*objects.String)
				if !ok {
					return Error("get() requires a string path")
				}

				return getYAMLValue(args[0], path.Value)
			}),

			// set sets a value in an object using dot notation path.
			// Returns a new object with the value set (does not mutate original).
			// Usage: newObj = yaml.set(obj, "path.to.key", value)
			"set": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("set() takes exactly 3 arguments")
				}

				path, ok := args[1].(*objects.String)
				if !ok {
					return Error("set() requires a string path")
				}

				return setYAMLValue(args[0], path.Value, args[2])
			}),

			// has checks if a path exists in an object.
			// Usage: exists = yaml.has(obj, "path.to.key")
			"has": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("has() takes exactly 2 arguments")
				}

				path, ok := args[1].(*objects.String)
				if !ok {
					return Error("has() requires a string path")
				}

				return Bool(yamlPathExists(args[0], path.Value))
			}),

			// keys returns all keys at the root level of an object.
			// Usage: keyList = yaml.keys(obj)
			"keys": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("keys() takes exactly 1 argument")
				}

				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("keys() requires a map argument")
				}

				keys := make([]objects.Object, 0, len(m.Pairs))
				for _, pair := range m.Pairs {
					keys = append(keys, pair.Key)
				}

				return Array(keys...)
			}),

			// values returns all values at the root level of an object.
			// Usage: valueList = yaml.values(obj)
			"values": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("values() takes exactly 1 argument")
				}

				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("values() requires a map argument")
				}

				values := make([]objects.Object, 0, len(m.Pairs))
				for _, pair := range m.Pairs {
					values = append(values, pair.Value)
				}

				return Array(values...)
			}),

			// ============================================================
			// Document operations
			// ============================================================

			// parseDocuments parses a YAML string with multiple documents (separated by ---).
			// Returns an array of objects.
			// Usage: docs = yaml.parseDocuments(yamlStr)
			"parseDocuments": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseDocuments() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseDocuments() requires a string argument")
				}

				docs := splitYAMLDocuments(s.Value)
				result := make([]objects.Object, 0, len(docs))

				for _, doc := range docs {
					if strings.TrimSpace(doc) == "" {
						continue
					}
					obj, err := objects.ParseYAML(doc)
					if err != nil {
						return Error(fmt.Sprintf("parseDocuments() failed: %s", err.Error()))
					}
					result = append(result, obj)
				}

				return Array(result...)
			}),

			// stringifyDocuments converts multiple objects to a multi-document YAML string.
			// Usage: yamlStr = yaml.stringifyDocuments([obj1, obj2, ...], [indent])
			"stringifyDocuments": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stringifyDocuments() takes at least 1 argument")
				}

				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("stringifyDocuments() requires an array argument")
				}

				indent := 2
				if len(args) > 1 {
					if indentInt, ok := args[1].(*objects.Int); ok {
						indent = int(indentInt.Value)
					}
				}

				var result string
				for i, elem := range arr.Elements {
					if i > 0 {
						result += "\n---\n"
					}
					result += objects.SerializeYAML(elem, indent)
				}

				return String(result)
			}),

			// ============================================================
			// Additional utility functions
			// ============================================================

			// diff compares two YAML documents and returns the differences.
			// Usage: differences = yaml.diff(doc1, doc2)
			// Returns an array of differences, each with path, old, new fields.
			"diff": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("diff() takes exactly 2 arguments")
				}

				differences := yamlDiffRecursive(args[0], args[1], "")
				return differences
			}),

			// flatten flattens a nested YAML document to dot-notation paths.
			// Usage: flat = yaml.flatten(obj, [separator])
			// separator defaults to "."
			"flatten": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("flatten() takes at least 1 argument")
				}

				separator := "."
				if len(args) > 1 {
					if sep, ok := args[1].(*objects.String); ok {
						separator = sep.Value
					}
				}

				result := make(map[objects.HashKey]objects.MapPair)
				yamlFlattenRecursive(args[0], "", separator, result)
				return &objects.Map{Pairs: result}
			}),

			// expand expands a flattened document back to nested structure.
			// Usage: nested = yaml.expand(flat, [separator])
			// separator defaults to "."
			"expand": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("expand() takes at least 1 argument")
				}

				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("expand() requires a map argument")
				}

				separator := "."
				if len(args) > 1 {
					if sep, ok := args[1].(*objects.String); ok {
						separator = sep.Value
					}
				}

				return yamlExpandMap(m, separator)
			}),

			// clone creates a deep copy of a YAML object.
			// Usage: copy = yaml.clone(obj)
			"clone": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("clone() takes exactly 1 argument")
				}
				return yamlDeepCopy(args[0])
			}),

			// equals compares two YAML objects for equality.
			// Usage: result = yaml.equals(obj1, obj2)
			"equals": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("equals() takes exactly 2 arguments")
				}
				return Bool(yamlEquals(args[0], args[1]))
			}),

			// paths returns all paths in a YAML document.
			// Usage: pathList = yaml.paths(obj)
			"paths": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("paths() takes exactly 1 argument")
				}

				var paths []objects.Object
				yamlCollectPaths(args[0], "", &paths)
				return &objects.Array{Elements: paths}
			}),

			// find finds all paths matching a pattern.
			// Usage: matches = yaml.find(obj, pattern)
			// Pattern supports wildcards: * for single segment, ** for multiple segments
			"find": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("find() takes exactly 2 arguments")
				}

				pattern, ok := args[1].(*objects.String)
				if !ok {
					return Error("find() requires a string pattern as second argument")
				}

				var matches []objects.Object
				yamlFindByPattern(args[0], pattern.Value, "", &matches)
				return &objects.Array{Elements: matches}
			}),
		},
	})
}

// Helper functions

// yamlDeepMergeMaps recursively merges two maps
func yamlDeepMergeMaps(base, override *objects.Map) *objects.Map {
	result := make(map[objects.HashKey]objects.MapPair)

	// Copy base
	for k, v := range base.Pairs {
		result[k] = v
	}

	// Merge override
	for _, pair := range override.Pairs {
		existing, exists := result[pair.Key.HashKey()]
		if exists {
			existingMap, existingIsMap := existing.Value.(*objects.Map)
			newMap, newIsMap := pair.Value.(*objects.Map)

			if existingIsMap && newIsMap {
				merged := yamlDeepMergeMaps(existingMap, newMap)
				result[pair.Key.HashKey()] = objects.MapPair{
					Key:   pair.Key,
					Value: merged,
				}
				continue
			}
		}
		result[pair.Key.HashKey()] = pair
	}

	return &objects.Map{Pairs: result}
}

// getYAMLValue retrieves a value using dot notation path
func getYAMLValue(obj objects.Object, path string) objects.Object {
	if path == "" {
		return obj
	}

	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		// Handle array index
		if idx, err := parseArrayIndex(part); err == nil {
			arr, ok := current.(*objects.Array)
			if !ok {
				return objects.NULL
			}
			if idx < 0 || idx >= len(arr.Elements) {
				return objects.NULL
			}
			current = arr.Elements[idx]
			continue
		}

		// Handle map key
		m, ok := current.(*objects.Map)
		if !ok {
			return objects.NULL
		}

		key := objects.NewString(part)
		pair, exists := m.Pairs[key.HashKey()]
		if !exists {
			return objects.NULL
		}
		current = pair.Value
	}

	return current
}

// yamlPathExists checks if a path exists in an object
func yamlPathExists(obj objects.Object, path string) bool {
	if path == "" {
		return true
	}

	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		// Handle array index
		if idx, err := parseArrayIndex(part); err == nil {
			arr, ok := current.(*objects.Array)
			if !ok {
				return false
			}
			if idx < 0 || idx >= len(arr.Elements) {
				return false
			}
			current = arr.Elements[idx]
			continue
		}

		// Handle map key
		m, ok := current.(*objects.Map)
		if !ok {
			return false
		}

		key := objects.NewString(part)
		pair, exists := m.Pairs[key.HashKey()]
		if !exists {
			return false
		}
		current = pair.Value
	}

	return true
}

// setYAMLValue sets a value using dot notation path
func setYAMLValue(obj objects.Object, path string, value objects.Object) objects.Object {
	if path == "" {
		return value
	}

	parts := strings.Split(path, ".")
	return setYAMLValueRecursive(obj, parts, 0, value)
}

// setYAMLValueRecursive recursively sets a value
func setYAMLValueRecursive(obj objects.Object, parts []string, index int, value objects.Object) objects.Object {
	if index >= len(parts) {
		return value
	}

	part := parts[index]

	// Handle array index
	if idx, err := parseArrayIndex(part); err == nil {
		arr, ok := obj.(*objects.Array)
		if !ok {
			return obj
		}

		newElements := make([]objects.Object, len(arr.Elements))
		copy(newElements, arr.Elements)

		if idx >= 0 && idx < len(newElements) {
			newElements[idx] = setYAMLValueRecursive(newElements[idx], parts, index+1, value)
		}

		return &objects.Array{Elements: newElements}
	}

	// Handle map key
	m, ok := obj.(*objects.Map)
	if !ok {
		// Create new map if needed
		m = &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	}

	newPairs := make(map[objects.HashKey]objects.MapPair)
	for k, v := range m.Pairs {
		newPairs[k] = v
	}

	key := objects.NewString(part)
	if index == len(parts)-1 {
		// Set the final value
		newPairs[key.HashKey()] = objects.MapPair{
			Key:   key,
			Value: value,
		}
	} else {
		// Recurse
		var existingValue objects.Object = objects.NULL
		if pair, exists := newPairs[key.HashKey()]; exists {
			existingValue = pair.Value
		}
		newPairs[key.HashKey()] = objects.MapPair{
			Key:   key,
			Value: setYAMLValueRecursive(existingValue, parts, index+1, value),
		}
	}

	return &objects.Map{Pairs: newPairs}
}

// parseArrayIndex parses a string as an array index (e.g., "[0]")
func parseArrayIndex(s string) (int, error) {
	if len(s) < 3 || s[0] != '[' || s[len(s)-1] != ']' {
		return -1, fmt.Errorf("not an array index")
	}

	var idx int
	_, err := fmt.Sscanf(s, "[%d]", &idx)
	return idx, err
}

// splitYAMLDocuments splits a YAML string by document separators
func splitYAMLDocuments(s string) []string {
	var docs []string
	var current strings.Builder
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if current.Len() > 0 {
				docs = append(docs, current.String())
				current.Reset()
			}
		} else if trimmed == "..." {
			if current.Len() > 0 {
				docs = append(docs, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				current.WriteByte('\n')
			}
			current.WriteString(line)
		}
	}

	if current.Len() > 0 {
		docs = append(docs, current.String())
	}

	return docs
}

// yamlDiffRecursive compares two objects recursively and returns differences
func yamlDiffRecursive(obj1, obj2 objects.Object, path string) *objects.Array {
	var differences []objects.Object

	// Check if types are different
	if obj1.Type() != obj2.Type() {
		differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(path)},
			objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("type_change")},
			objects.NewString("oldType").HashKey():  {Key: objects.NewString("oldType"), Value: objects.NewString(string(obj1.Type()))},
			objects.NewString("newType").HashKey():  {Key: objects.NewString("newType"), Value: objects.NewString(string(obj2.Type()))},
			objects.NewString("oldValue").HashKey(): {Key: objects.NewString("oldValue"), Value: obj1},
			objects.NewString("newValue").HashKey(): {Key: objects.NewString("newValue"), Value: obj2},
		}})
		return &objects.Array{Elements: differences}
	}

	switch o1 := obj1.(type) {
	case *objects.Map:
		o2 := obj2.(*objects.Map)

		// Check for keys in obj1 but not in obj2
		for _, pair := range o1.Pairs {
			keyStr := pair.Key.(*objects.String).Value
			newPath := path
			if path != "" {
				newPath = path + "." + keyStr
			} else {
				newPath = keyStr
			}

			if _, exists := o2.Pairs[pair.Key.HashKey()]; !exists {
				differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("removed")},
					objects.NewString("oldValue").HashKey(): {Key: objects.NewString("oldValue"), Value: pair.Value},
				}})
			}
		}

		// Check for keys in obj2 but not in obj1
		for _, pair := range o2.Pairs {
			keyStr := pair.Key.(*objects.String).Value
			newPath := path
			if path != "" {
				newPath = path + "." + keyStr
			} else {
				newPath = keyStr
			}

			if _, exists := o1.Pairs[pair.Key.HashKey()]; !exists {
				differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("added")},
					objects.NewString("newValue").HashKey(): {Key: objects.NewString("newValue"), Value: pair.Value},
				}})
			}
		}

		// Check for changed values
		for _, pair := range o1.Pairs {
			if pair2, exists := o2.Pairs[pair.Key.HashKey()]; exists {
				keyStr := pair.Key.(*objects.String).Value
				newPath := path
				if path != "" {
					newPath = path + "." + keyStr
				} else {
					newPath = keyStr
				}
				subDiffs := yamlDiffRecursive(pair.Value, pair2.Value, newPath)
				differences = append(differences, subDiffs.Elements...)
			}
		}

	case *objects.Array:
		o2 := obj2.(*objects.Array)

		maxLen := len(o1.Elements)
		if len(o2.Elements) > maxLen {
			maxLen = len(o2.Elements)
		}

		for i := 0; i < maxLen; i++ {
			newPath := fmt.Sprintf("%s.[%d]", path, i)
			if i >= len(o1.Elements) {
				differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("added")},
					objects.NewString("newValue").HashKey(): {Key: objects.NewString("newValue"), Value: o2.Elements[i]},
				}})
			} else if i >= len(o2.Elements) {
				differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("removed")},
					objects.NewString("oldValue").HashKey(): {Key: objects.NewString("oldValue"), Value: o1.Elements[i]},
				}})
			} else {
				subDiffs := yamlDiffRecursive(o1.Elements[i], o2.Elements[i], newPath)
				differences = append(differences, subDiffs.Elements...)
			}
		}

	default:
		// Scalar values - compare directly
		if !yamlEquals(obj1, obj2) {
			differences = append(differences, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
				objects.NewString("path").HashKey():     {Key: objects.NewString("path"), Value: objects.NewString(path)},
				objects.NewString("type").HashKey():     {Key: objects.NewString("type"), Value: objects.NewString("changed")},
				objects.NewString("oldValue").HashKey(): {Key: objects.NewString("oldValue"), Value: obj1},
				objects.NewString("newValue").HashKey(): {Key: objects.NewString("newValue"), Value: obj2},
			}})
		}
	}

	return &objects.Array{Elements: differences}
}

// yamlFlattenRecursive flattens a nested object to dot-notation paths
func yamlFlattenRecursive(obj objects.Object, prefix, separator string, result map[objects.HashKey]objects.MapPair) {
	switch o := obj.(type) {
	case *objects.Map:
		if len(o.Pairs) == 0 {
			// Empty map
			key := objects.NewString(prefix)
			result[key.HashKey()] = objects.MapPair{Key: key, Value: obj}
			return
		}
		for _, pair := range o.Pairs {
			keyStr := pair.Key.(*objects.String).Value
			newPath := prefix
			if prefix != "" {
				newPath = prefix + separator + keyStr
			} else {
				newPath = keyStr
			}
			yamlFlattenRecursive(pair.Value, newPath, separator, result)
		}

	case *objects.Array:
		if len(o.Elements) == 0 {
			// Empty array
			key := objects.NewString(prefix)
			result[key.HashKey()] = objects.MapPair{Key: key, Value: obj}
			return
		}
		for i, elem := range o.Elements {
			newPath := fmt.Sprintf("%s%s[%d]", prefix, separator, i)
			if prefix == "" {
				newPath = fmt.Sprintf("[%d]", i)
			}
			yamlFlattenRecursive(elem, newPath, separator, result)
		}

	default:
		key := objects.NewString(prefix)
		result[key.HashKey()] = objects.MapPair{Key: key, Value: obj}
	}
}

// yamlExpandMap expands a flattened map back to nested structure
func yamlExpandMap(flat *objects.Map, separator string) objects.Object {
	result := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}

	for _, pair := range flat.Pairs {
		keyStr := pair.Key.(*objects.String).Value
		parts := splitPath(keyStr, separator)
		setNestedValue(result, parts, 0, pair.Value)
	}

	if len(result.Pairs) == 0 {
		return result
	}

	// If result has only array indices as top-level keys, return array
	if isIndexPath(result) {
		return mapToArray(result)
	}

	return result
}

// splitPath splits a path into parts, handling array indices
func splitPath(path, separator string) []string {
	var parts []string
	current := ""
	inBracket := false

	for _, ch := range path {
		switch ch {
		case '[':
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			inBracket = true
		case ']':
			inBracket = false
		default:
			if !inBracket && string(ch) == separator {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(ch)
			}
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// setNestedValue sets a value at a nested path
func setNestedValue(m *objects.Map, parts []string, index int, value objects.Object) {
	if index >= len(parts) {
		return
	}

	part := parts[index]
	key := objects.NewString(part)

	if index == len(parts)-1 {
		m.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: value}
		return
	}

	// Check if next part is array index
	if index+1 < len(parts) && isNumericIndex(parts[index+1]) {
		// Create array at this position
		if _, exists := m.Pairs[key.HashKey()]; !exists {
			m.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}}
		}
		if inner, ok := m.Pairs[key.HashKey()].Value.(*objects.Map); ok {
			setNestedValue(inner, parts, index+1, value)
		}
	} else {
		// Create map at this position
		if _, exists := m.Pairs[key.HashKey()]; !exists {
			m.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}}
		}
		if inner, ok := m.Pairs[key.HashKey()].Value.(*objects.Map); ok {
			setNestedValue(inner, parts, index+1, value)
		}
	}
}

// isNumericIndex checks if a string is a numeric array index
func isNumericIndex(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return len(s) > 0
}

// isIndexPath checks if a map only has numeric keys (array indices)
func isIndexPath(m *objects.Map) bool {
	if len(m.Pairs) == 0 {
		return false
	}
	for _, pair := range m.Pairs {
		if !isNumericIndex(pair.Key.(*objects.String).Value) {
			return false
		}
	}
	return true
}

// mapToArray converts a map with numeric keys to an array
func mapToArray(m *objects.Map) *objects.Array {
	// Find max index
	maxIdx := -1
	for _, pair := range m.Pairs {
		idx := 0
		fmt.Sscanf(pair.Key.(*objects.String).Value, "%d", &idx)
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	if maxIdx < 0 {
		return &objects.Array{}
	}

	elements := make([]objects.Object, maxIdx+1)
	for i := range elements {
		elements[i] = objects.NULL
	}

	for _, pair := range m.Pairs {
		idx := 0
		fmt.Sscanf(pair.Key.(*objects.String).Value, "%d", &idx)
		if idx >= 0 && idx < len(elements) {
			// Check if value is also an index path (nested array)
			if inner, ok := pair.Value.(*objects.Map); ok && isIndexPath(inner) {
				elements[idx] = mapToArray(inner)
			} else {
				elements[idx] = pair.Value
			}
		}
	}

	return &objects.Array{Elements: elements}
}

// yamlDeepCopy creates a deep copy of an object
func yamlDeepCopy(obj objects.Object) objects.Object {
	switch o := obj.(type) {
	case *objects.Map:
		newPairs := make(map[objects.HashKey]objects.MapPair)
		for k, pair := range o.Pairs {
			newPairs[k] = objects.MapPair{
				Key:   pair.Key,
				Value: yamlDeepCopy(pair.Value),
			}
		}
		return &objects.Map{Pairs: newPairs}

	case *objects.Array:
		newElements := make([]objects.Object, len(o.Elements))
		for i, elem := range o.Elements {
			newElements[i] = yamlDeepCopy(elem)
		}
		return &objects.Array{Elements: newElements}

	default:
		return obj
	}
}

// yamlEquals compares two objects for equality
func yamlEquals(obj1, obj2 objects.Object) bool {
	if obj1.Type() != obj2.Type() {
		return false
	}

	switch o1 := obj1.(type) {
	case *objects.Map:
		o2 := obj2.(*objects.Map)
		if len(o1.Pairs) != len(o2.Pairs) {
			return false
		}
		for k, pair := range o1.Pairs {
			pair2, exists := o2.Pairs[k]
			if !exists || !yamlEquals(pair.Value, pair2.Value) {
				return false
			}
		}
		return true

	case *objects.Array:
		o2 := obj2.(*objects.Array)
		if len(o1.Elements) != len(o2.Elements) {
			return false
		}
		for i := range o1.Elements {
			if !yamlEquals(o1.Elements[i], o2.Elements[i]) {
				return false
			}
		}
		return true

	case *objects.String:
		return o1.Value == obj2.(*objects.String).Value

	case *objects.Int:
		return o1.Value == obj2.(*objects.Int).Value

	case *objects.Float:
		return o1.Value == obj2.(*objects.Float).Value

	case *objects.Bool:
		return o1.Value == obj2.(*objects.Bool).Value

	case *objects.Null:
		return true

	default:
		return obj1.Inspect() == obj2.Inspect()
	}
}

// yamlCollectPaths collects all paths in a YAML document
func yamlCollectPaths(obj objects.Object, prefix string, paths *[]objects.Object) {
	switch o := obj.(type) {
	case *objects.Map:
		for _, pair := range o.Pairs {
			keyStr := pair.Key.(*objects.String).Value
			newPath := prefix
			if prefix != "" {
				newPath = prefix + "." + keyStr
			} else {
				newPath = keyStr
			}
			*paths = append(*paths, objects.NewString(newPath))
			yamlCollectPaths(pair.Value, newPath, paths)
		}

	case *objects.Array:
		for i := range o.Elements {
			newPath := fmt.Sprintf("%s.[%d]", prefix, i)
			*paths = append(*paths, objects.NewString(newPath))
			yamlCollectPaths(o.Elements[i], newPath, paths)
		}
	}
}

// yamlFindByPattern finds all paths matching a pattern
func yamlFindByPattern(obj objects.Object, pattern, currentPath string, matches *[]objects.Object) {
	switch o := obj.(type) {
	case *objects.Map:
		for _, pair := range o.Pairs {
			keyStr := pair.Key.(*objects.String).Value
			newPath := currentPath
			if currentPath != "" {
				newPath = currentPath + "." + keyStr
			} else {
				newPath = keyStr
			}
			if pathMatchesPattern(newPath, pattern) {
				*matches = append(*matches, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():  {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("value").HashKey(): {Key: objects.NewString("value"), Value: pair.Value},
				}})
			}
			yamlFindByPattern(pair.Value, pattern, newPath, matches)
		}

	case *objects.Array:
		for i := range o.Elements {
			newPath := fmt.Sprintf("%s.[%d]", currentPath, i)
			if pathMatchesPattern(newPath, pattern) {
				*matches = append(*matches, &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("path").HashKey():  {Key: objects.NewString("path"), Value: objects.NewString(newPath)},
					objects.NewString("value").HashKey(): {Key: objects.NewString("value"), Value: o.Elements[i]},
				}})
			}
			yamlFindByPattern(o.Elements[i], pattern, newPath, matches)
		}
	}
}

// pathMatchesPattern checks if a path matches a pattern
// Supports * as wildcard for single path segment
// Supports ** as wildcard for multiple path segments
func pathMatchesPattern(path, pattern string) bool {
	if pattern == "" || pattern == "*" || pattern == "**" {
		return true
	}
	if path == pattern {
		return true
	}

	// Simple wildcard matching
	pathParts := strings.Split(path, ".")
	patternParts := strings.Split(pattern, ".")

	return matchPathParts(pathParts, patternParts, 0, 0)
}

// matchPathParts recursively matches path parts against pattern parts
func matchPathParts(pathParts, patternParts []string, pathIdx, patternIdx int) bool {
	if patternIdx >= len(patternParts) {
		return pathIdx >= len(pathParts)
	}

	if pathIdx >= len(pathParts) {
		// Check if remaining pattern parts are wildcards
		for i := patternIdx; i < len(patternParts); i++ {
			if patternParts[i] != "*" && patternParts[i] != "**" {
				return false
			}
		}
		return true
	}

	pattern := patternParts[patternIdx]
	path := pathParts[pathIdx]

	switch pattern {
	case "**":
		// ** matches zero or more path segments
		if matchPathParts(pathParts, patternParts, pathIdx, patternIdx+1) {
			return true
		}
		return matchPathParts(pathParts, patternParts, pathIdx+1, patternIdx)

	case "*":
		// * matches exactly one path segment
		return matchPathParts(pathParts, patternParts, pathIdx+1, patternIdx+1)

	default:
		if pattern == path {
			return matchPathParts(pathParts, patternParts, pathIdx+1, patternIdx+1)
		}
		return false
	}
}
