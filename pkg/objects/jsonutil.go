// pkg/objects/jsonutil.go
// JSON utility functions for Xxlang objects.
// These functions are exported for use by both the builtin functions and the stdlib json module.
package objects

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ObjectToJSONOptions contains options for JSON serialization
type ObjectToJSONOptions struct {
	Indent     bool
	SortKeys   bool
	IndentStr  string
	StartLevel int
}

// ObjectToJSON converts an Object to JSON bytes with the given options.
// This is the primary function for JSON serialization in Xxlang.
func ObjectToJSON(obj Object, opts ObjectToJSONOptions) ([]byte, error) {
	if opts.IndentStr == "" {
		opts.IndentStr = "  " // default 2 spaces
	}
	return objectToJSONRecursive(obj, opts.Indent, opts.SortKeys, opts.IndentStr, opts.StartLevel)
}

// objectToJSONRecursive recursively converts an Object to JSON bytes
func objectToJSONRecursive(obj Object, indent, sortKeys bool, indentStr string, level int) ([]byte, error) {
	if obj == nil {
		return []byte("null"), nil
	}

	switch o := obj.(type) {
	case *Null:
		return []byte("null"), nil
	case *Bool:
		if o.Value {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case *Int:
		return []byte(strconv.FormatInt(o.Value, 10)), nil
	case *Float:
		return []byte(strconv.FormatFloat(o.Value, 'f', -1, 64)), nil
	case *String:
		data, err := json.Marshal(o.Value)
		if err != nil {
			return nil, err
		}
		return data, nil
	case *Array:
		var buf strings.Builder
		buf.WriteString("[")
		if indent {
			buf.WriteString("\n")
		}
		for i, elem := range o.Elements {
			if indent {
				for j := 0; j <= level; j++ {
					buf.WriteString(indentStr)
				}
			}
			data, err := objectToJSONRecursive(elem, indent, sortKeys, indentStr, level+1)
			if err != nil {
				return nil, err
			}
			buf.Write(data)
			if i < len(o.Elements)-1 {
				buf.WriteString(",")
			}
			if indent {
				buf.WriteString("\n")
			}
		}
		if indent {
			for j := 0; j < level; j++ {
				buf.WriteString(indentStr)
			}
		}
		buf.WriteString("]")
		return []byte(buf.String()), nil
	case *Map:
		var buf strings.Builder
		buf.WriteString("{")
		if indent {
			buf.WriteString("\n")
		}

		// Get keys
		keys := make([]string, 0, len(o.Pairs))
		for _, pair := range o.Pairs {
			if keyStr, ok := pair.Key.(*String); ok {
				keys = append(keys, keyStr.Value)
			}
		}

		// Sort keys if requested
		if sortKeys {
			sort.Strings(keys)
		}

		for i, key := range keys {
			if indent {
				for j := 0; j <= level; j++ {
					buf.WriteString(indentStr)
				}
			}
			// Write key
			keyData, _ := json.Marshal(key)
			buf.Write(keyData)
			buf.WriteString(":")
			if indent {
				buf.WriteString(" ")
			}
			// Find value for this key
			for _, pair := range o.Pairs {
				if keyStr, ok := pair.Key.(*String); ok && keyStr.Value == key {
					data, err := objectToJSONRecursive(pair.Value, indent, sortKeys, indentStr, level+1)
					if err != nil {
						return nil, err
					}
					buf.Write(data)
					break
				}
			}
			if i < len(keys)-1 {
				buf.WriteString(",")
			}
			if indent {
				buf.WriteString("\n")
			}
		}
		if indent {
			for j := 0; j < level; j++ {
				buf.WriteString(indentStr)
			}
		}
		buf.WriteString("}")
		return []byte(buf.String()), nil
	case *OrderedMap:
		var buf strings.Builder
		buf.WriteString("{")
		if indent {
			buf.WriteString("\n")
		}

		// Iterate in insertion order (ignore sortKeys for OrderedMap)
		for i, pair := range o.orderSlice {
			if indent {
				for j := 0; j <= level; j++ {
					buf.WriteString(indentStr)
				}
			}
			// Write key (must be string for JSON)
			if keyStr, ok := pair.Key.(*String); ok {
				keyData, _ := json.Marshal(keyStr.Value)
				buf.Write(keyData)
			} else {
				// Non-string keys: use Inspect()
				keyData, _ := json.Marshal(pair.Key.Inspect())
				buf.Write(keyData)
			}
			buf.WriteString(":")
			if indent {
				buf.WriteString(" ")
			}
			// Write value recursively
			data, err := objectToJSONRecursive(pair.Value, indent, sortKeys, indentStr, level+1)
			if err != nil {
				return nil, err
			}
			buf.Write(data)
			if i < len(o.orderSlice)-1 {
				buf.WriteString(",")
			}
			if indent {
				buf.WriteString("\n")
			}
		}
		if indent {
			for j := 0; j < level; j++ {
				buf.WriteString(indentStr)
			}
		}
		buf.WriteString("}")
		return []byte(buf.String()), nil
	default:
		// For other types, use Inspect as string
		data, err := json.Marshal(o.Inspect())
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

// JSONToObject parses a JSON string and returns an Xxlang Object.
// This is the primary function for JSON deserialization in Xxlang.
func JSONToObject(s string) (Object, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return GoValueToObject(v), nil
}

// GoValueToObject converts a Go value (from JSON unmarshaling) to an Xxlang Object.
// This function is exported for use by other packages that need to convert Go values to Xxlang objects.
func GoValueToObject(v interface{}) Object {
	if v == nil {
		return NULL
	}

	switch val := v.(type) {
	case bool:
		if val {
			return TRUE
		}
		return FALSE
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return NewInt(int64(val))
		}
		return NewFloat(val)
	case string:
		return NewString(val)
	case []interface{}:
		elements := make([]Object, len(val))
		for i, elem := range val {
			elements[i] = GoValueToObject(elem)
		}
		return NewArray(elements)
	case map[string]interface{}:
		pairs := make(map[HashKey]MapPair)
		for k, v := range val {
			key := NewString(k)
			hashKey := key.HashKey()
			pairs[hashKey] = MapPair{
				Key:   key,
				Value: GoValueToObject(v),
			}
		}
		return NewMap(pairs)
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

// ObjectToGoValue converts an Xxlang Object to a Go value suitable for JSON marshaling.
// Returns nil for null, and returns an error for types that cannot be serialized.
func ObjectToGoValue(obj Object) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	switch o := obj.(type) {
	case *Null:
		return nil, nil
	case *Bool:
		return o.Value, nil
	case *Int:
		return o.Value, nil
	case *Float:
		return o.Value, nil
	case *String:
		return o.Value, nil
	case *Array:
		result := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			val, err := ObjectToGoValue(elem)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case *Map:
		result := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, ok := pair.Key.(*String)
			if !ok {
				// Skip non-string keys
				continue
			}
			val, err := ObjectToGoValue(pair.Value)
			if err != nil {
				return nil, err
			}
			result[key.Value] = val
		}
		return result, nil
	case *OrderedMap:
		// Note: Go's map doesn't preserve order, but json.Marshal in Go 1.14+ preserves insertion order
		result := make(map[string]interface{})
		for _, pair := range o.orderSlice {
			key, ok := pair.Key.(*String)
			if !ok {
				// Skip non-string keys
				continue
			}
			val, err := ObjectToGoValue(pair.Value)
			if err != nil {
				return nil, err
			}
			result[key.Value] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot serialize type %s to JSON", obj.Type())
	}
}
