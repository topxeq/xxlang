// pkg/objects/builtin_json.go
// JSON enhancement built-in functions for Xxlang
package objects

import (
	"bytes"
	"encoding/json"
	"strings"
)

func init() {
	// JSON functions
	Builtins["formatJson"] = &Builtin{Fn: builtinFormatJson}
	Builtins["compactJson"] = &Builtin{Fn: builtinCompactJson}
	Builtins["getJsonNodeStr"] = &Builtin{Fn: builtinGetJsonNodeStr}
	Builtins["getJsonNodeStrs"] = &Builtin{Fn: builtinGetJsonNodeStrs}
	Builtins["strsToJson"] = &Builtin{Fn: builtinStrsToJson}
	Builtins["jsonPath"] = &Builtin{Fn: builtinGetJsonNodeStr}
	Builtins["jsonValid"] = &Builtin{Fn: builtinJsonValid}
	Builtins["jsonType"] = &Builtin{Fn: builtinJsonType}
}

// builtinFormatJson - format/beautify JSON string
// Usage: formatJson(jsonStr) -> string
//
//	formatJson(jsonStr, indent) -> string
func builtinFormatJson(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for formatJson. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'formatJson' must be STRING, got %s", args[0].Type())
	}

	indent := "  "
	if len(args) == 2 {
		ind, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'formatJson' must be STRING, got %s", args[1].Type())
		}
		indent = ind.Value
	}

	var buf bytes.Buffer
	err := json.Indent(&buf, []byte(str.Value), "", indent)
	if err != nil {
		return newError("formatJson error: %v", err)
	}

	return NewString(buf.String())
}

// builtinCompactJson - compact JSON string (remove whitespace)
// Usage: compactJson(jsonStr) -> string
func builtinCompactJson(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for compactJson. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'compactJson' must be STRING, got %s", args[0].Type())
	}

	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(str.Value))
	if err != nil {
		return newError("compactJson error: %v", err)
	}

	return NewString(buf.String())
}

// builtinGetJsonNodeStr - get JSON node value by path
// Usage: getJsonNodeStr(jsonStr, path) -> string
// Path examples: "name", "user.name", "items.0", "items.#", "users.#.name"
func builtinGetJsonNodeStr(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for getJsonNodeStr. got=%d, want=2", len(args))
	}

	jsonStr, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getJsonNodeStr' must be STRING, got %s", args[0].Type())
	}

	path, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'getJsonNodeStr' must be STRING, got %s", args[1].Type())
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr.Value), &data)
	if err != nil {
		return newError("getJsonNodeStr parse error: %v", err)
	}

	result := getJSONValue(data, path.Value)
	if result == nil {
		return NULL
	}

	switch v := result.(type) {
	case string:
		return NewString(v)
	case float64:
		if v == float64(int64(v)) {
			return NewInt(int64(v))
		}
		return NewFloat(v)
	case bool:
		return &Bool{Value: v}
	case nil:
		return NULL
	case []interface{}:
		return goArrayToObject(v)
	case map[string]interface{}:
		return goMapToObject(v)
	default:
		return NewString(jsonValueToString(v))
	}
}

// builtinGetJsonNodeStrs - get JSON node array values by path
// Usage: getJsonNodeStrs(jsonStr, path) -> array
func builtinGetJsonNodeStrs(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for getJsonNodeStrs. got=%d, want=2", len(args))
	}

	jsonStr, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getJsonNodeStrs' must be STRING, got %s", args[0].Type())
	}

	path, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'getJsonNodeStrs' must be STRING, got %s", args[1].Type())
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr.Value), &data)
	if err != nil {
		return newError("getJsonNodeStrs parse error: %v", err)
	}

	result := getJSONValue(data, path.Value)
	if result == nil {
		return NULL
	}

	switch v := result.(type) {
	case []interface{}:
		elements := make([]Object, len(v))
		for i, item := range v {
			elements[i] = goValueToObject(item)
		}
		return NewArray(elements)
	default:
		return NULL
	}
}

// builtinStrsToJson - convert string array to JSON object
// Usage: strsToJson(keys, values) -> string
func builtinStrsToJson(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strsToJson. got=%d, want=2", len(args))
	}

	keys, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'strsToJson' must be ARRAY, got %s", args[0].Type())
	}

	values, ok := args[1].(*Array)
	if !ok {
		return newError("second argument to 'strsToJson' must be ARRAY, got %s", args[1].Type())
	}

	obj := make(map[string]interface{})
	minLen := len(keys.Elements)
	if len(values.Elements) < minLen {
		minLen = len(values.Elements)
	}

	for i := 0; i < minLen; i++ {
		key, ok := keys.Elements[i].(*String)
		if !ok {
			continue
		}
		obj[key.Value] = jsonObjectToGoValue(values.Elements[i])
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return newError("strsToJson error: %v", err)
	}

	return NewString(string(jsonBytes))
}

// builtinJsonValid - check if string is valid JSON
// Usage: jsonValid(jsonStr) -> bool
func builtinJsonValid(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for jsonValid. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'jsonValid' must be STRING, got %s", args[0].Type())
	}

	return &Bool{Value: json.Valid([]byte(str.Value))}
}

// builtinJsonType - get the type of JSON value at path
// Usage: jsonType(jsonStr) -> string
//
//	jsonType(jsonStr, path) -> string
func builtinJsonType(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for jsonType. got=%d, want=1 or 2", len(args))
	}

	jsonStr, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'jsonType' must be STRING, got %s", args[0].Type())
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr.Value), &data)
	if err != nil {
		return newError("jsonType parse error: %v", err)
	}

	if len(args) == 2 {
		path, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'jsonType' must be STRING, got %s", args[1].Type())
		}
		data = getJSONValue(data, path.Value)
		if data == nil {
			return NewString("null")
		}
	}

	return NewString(getJSONTypeName(data))
}

// Helper functions for JSON navigation

func getJSONValue(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		// Check for array index
		if part == "#" {
			if arr, ok := current.([]interface{}); ok {
				return len(arr)
			}
			return nil
		}

		// Try as map key
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
			continue
		}

		// Try as array index
		if arr, ok := current.([]interface{}); ok {
			idx := parseArrayIndex(part)
			if idx >= 0 && idx < len(arr) {
				current = arr[idx]
				continue
			}
			return nil
		}

		return nil
	}

	return current
}

func parseArrayIndex(s string) int {
	if len(s) == 0 {
		return -1
	}

	// Handle negative indices
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}

	idx := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		idx = idx*10 + int(c-'0')
	}

	if neg {
		return -idx
	}
	return idx
}

func getJSONTypeName(data interface{}) string {
	switch data.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func jsonValueToString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func goArrayToObject(arr []interface{}) *Array {
	elements := make([]Object, len(arr))
	for i, item := range arr {
		elements[i] = goValueToObject(item)
	}
	return NewArray(elements)
}

func goMapToObject(m map[string]interface{}) *Map {
	pairs := make(map[HashKey]MapPair)
	for k, v := range m {
		key := NewString(k)
		pairs[key.HashKey()] = MapPair{
			Key:   key,
			Value: goValueToObject(v),
		}
	}
	return NewMap(pairs)
}

func goValueToObject(v interface{}) Object {
	switch val := v.(type) {
	case nil:
		return NULL
	case bool:
		return &Bool{Value: val}
	case float64:
		if val == float64(int64(val)) {
			return NewInt(int64(val))
		}
		return NewFloat(val)
	case string:
		return NewString(val)
	case []interface{}:
		return goArrayToObject(val)
	case map[string]interface{}:
		return goMapToObject(val)
	default:
		return NewString(jsonValueToString(v))
	}
}

func jsonObjectToGoValue(obj Object) interface{} {
	switch v := obj.(type) {
	case *Null:
		return nil
	case *Bool:
		return v.Value
	case *Int:
		return v.Value
	case *Float:
		return v.Value
	case *String:
		return v.Value
	case *Array:
		arr := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			arr[i] = jsonObjectToGoValue(elem)
		}
		return arr
	case *Map:
		m := make(map[string]interface{})
		for _, pair := range v.Pairs {
			if key, ok := pair.Key.(*String); ok {
				m[key.Value] = jsonObjectToGoValue(pair.Value)
			}
		}
		return m
	default:
		return v.Inspect()
	}
}
