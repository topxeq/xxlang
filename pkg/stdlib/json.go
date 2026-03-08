// pkg/stdlib/json.go
// JSON utilities for the Xxlang standard library.
package stdlib

import (
	"encoding/json"
	"fmt"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/json",
		Exports: map[string]objects.Object{
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}

				var value interface{}
				if err := json.Unmarshal([]byte(s.Value), &value); err != nil {
					return Error(fmt.Sprintf("parse() failed: %s", err.Error()))
				}

				return jsonToXxlang(value)
			}),

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

				goValue, err := xxlangToJSON(args[0])
				if err != nil {
					return err
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

			"encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				// Alias for stringify
				if len(args) < 1 {
					return Error("encode() takes at least 1 argument")
				}

				goValue, err := xxlangToJSON(args[0])
				if err != nil {
					return err
				}

				bytes, marshalErr := json.Marshal(goValue)
				if marshalErr != nil {
					return Error(fmt.Sprintf("encode() failed: %s", marshalErr.Error()))
				}

				return String(string(bytes))
			}),

			"decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				// Alias for parse
				if len(args) != 1 {
					return Error("decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("decode() requires a string argument")
				}

				var value interface{}
				if err := json.Unmarshal([]byte(s.Value), &value); err != nil {
					return Error(fmt.Sprintf("decode() failed: %s", err.Error()))
				}

				return jsonToXxlang(value)
			}),
		},
	})
}

// jsonToXxlang converts a Go value (from JSON unmarshaling) to an Xxlang object
func jsonToXxlang(value interface{}) objects.Object {
	if value == nil {
		return Null()
	}

	switch v := value.(type) {
	case bool:
		return Bool(v)
	case float64:
		// JSON numbers are always float64, but try to represent as int if possible
		if v == float64(int64(v)) {
			return Int(int64(v))
		}
		return Float(v)
	case string:
		return String(v)
	case []interface{}:
		elements := make([]objects.Object, len(v))
		for i, elem := range v {
			elements[i] = jsonToXxlang(elem)
		}
		return Array(elements...)
	case map[string]interface{}:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for key, val := range v {
			keyObj := String(key)
			valObj := jsonToXxlang(val)
			pairs[keyObj.HashKey()] = objects.MapPair{
				Key:   keyObj,
				Value: valObj,
			}
		}
		return &objects.Map{Pairs: pairs}
	default:
		return Error(fmt.Sprintf("unsupported JSON type: %T", value))
	}
}

// xxlangToJSON converts an Xxlang object to a Go value suitable for JSON marshaling
// Returns an error object if the value cannot be serialized
func xxlangToJSON(obj objects.Object) (interface{}, *objects.Error) {
	if obj == nil {
		return nil, nil
	}

	switch o := obj.(type) {
	case *objects.Null:
		return nil, nil
	case *objects.Bool:
		return o.Value, nil
	case *objects.Int:
		return o.Value, nil
	case *objects.Float:
		return o.Value, nil
	case *objects.String:
		return o.Value, nil
	case *objects.Array:
		result := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			val, err := xxlangToJSON(elem)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case *objects.Map:
		result := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, ok := pair.Key.(*objects.String)
			if !ok {
				// Skip non-string keys
				continue
			}
			val, err := xxlangToJSON(pair.Value)
			if err != nil {
				return nil, err
			}
			result[key.Value] = val
		}
		return result, nil
	default:
		return nil, Error(fmt.Sprintf("stringify() cannot serialize type: %s", obj.Type()))
	}
}
