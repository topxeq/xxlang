// pkg/interpreter/convert.go
// Type conversion helpers between Go and xxlang objects.
package interpreter

import (
	"fmt"
	"reflect"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ToGo converts an xxlang object to a Go value.
// Returns a Go native type that best represents the object:
//   - Int -> int64
//   - Float -> float64
//   - String -> string
//   - Bool -> bool
//   - Array -> []interface{} (elements converted recursively)
//   - Map -> map[string]interface{} (values converted recursively)
//   - Null -> nil
//   - Other -> objects.Object (returned as-is)
func ToGo(obj objects.Object) interface{} {
	if obj == nil || obj == objects.NULL {
		return nil
	}

	switch o := obj.(type) {
	case *objects.Int:
		return o.Value
	case *objects.Float:
		return o.Value
	case *objects.String:
		return o.Value
	case *objects.Bool:
		return o.Value
	case *objects.Array:
		result := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			result[i] = ToGo(elem)
		}
		return result
	case *objects.Map:
		result := make(map[string]interface{}, len(o.Pairs))
		for _, pair := range o.Pairs {
			// Use key's string representation
			result[pair.Key.Inspect()] = ToGo(pair.Value)
		}
		return result
	case *objects.Null:
		return nil
	default:
		// Return the object itself for types we can't convert
		return obj
	}
}

// FromGo converts a Go value to an xxlang object.
// Supported Go types:
//   - int, int8, int16, int32, int64 -> Int
//   - uint, uint8, uint16, uint32, uint64 -> Int
//   - float32, float64 -> Float
//   - string -> String
//   - bool -> Bool
//   - []T -> Array (elements converted recursively)
//   - map[string]T -> Map (values converted recursively)
//   - nil -> Null
//   - objects.Object -> returned as-is
func FromGo(value interface{}) (objects.Object, error) {
	if value == nil {
		return objects.NULL, nil
	}

	// If it's already an Object, return it directly
	if obj, ok := value.(objects.Object); ok {
		return obj, nil
	}

	// Use reflection for other types
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &objects.Int{Value: v.Int()}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &objects.Int{Value: int64(v.Uint())}, nil

	case reflect.Float32, reflect.Float64:
		return &objects.Float{Value: v.Float()}, nil

	case reflect.String:
		return &objects.String{Value: v.String()}, nil

	case reflect.Bool:
		if v.Bool() {
			return objects.TRUE, nil
		}
		return objects.FALSE, nil

	case reflect.Slice, reflect.Array:
		length := v.Len()
		elements := make([]objects.Object, length)
		for i := 0; i < length; i++ {
			elem, err := FromGo(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("error converting slice element %d: %v", i, err)
			}
			elements[i] = elem
		}
		return &objects.Array{Elements: elements}, nil

	case reflect.Map:
		// Only support string keys for now
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map keys must be strings, got %v", v.Type().Key())
		}

		pairs := make(map[objects.HashKey]objects.MapPair)
		iter := v.MapRange()
		for iter.Next() {
			keyStr := iter.Key().String()
			key := &objects.String{Value: keyStr}
			value, err := FromGo(iter.Value().Interface())
			if err != nil {
				return nil, fmt.Errorf("error converting map value for key %s: %v", keyStr, err)
			}
			pairs[key.HashKey()] = objects.MapPair{
				Key:   key,
				Value: value,
			}
		}
		return &objects.Map{Pairs: pairs}, nil

	case reflect.Ptr:
		if v.IsNil() {
			return objects.NULL, nil
		}
		return FromGo(v.Elem().Interface())

	case reflect.Interface:
		if v.IsNil() {
			return objects.NULL, nil
		}
		return FromGo(v.Elem().Interface())

	default:
		return nil, fmt.Errorf("cannot convert Go type %T to xxlang object", value)
	}
}

// ToInt converts an object to an int64 if possible.
// Returns an error if the object is not an Int.
func ToInt(obj objects.Object) (int64, error) {
	if i, ok := obj.(*objects.Int); ok {
		return i.Value, nil
	}
	return 0, fmt.Errorf("expected Int, got %s", obj.Type())
}

// ToFloat converts an object to a float64 if possible.
// Returns an error if the object is not a Float.
func ToFloat(obj objects.Object) (float64, error) {
	if f, ok := obj.(*objects.Float); ok {
		return f.Value, nil
	}
	return 0, fmt.Errorf("expected Float, got %s", obj.Type())
}

// ToString converts an object to a string if possible.
// Returns an error if the object is not a String.
func ToString(obj objects.Object) (string, error) {
	if s, ok := obj.(*objects.String); ok {
		return s.Value, nil
	}
	return "", fmt.Errorf("expected String, got %s", obj.Type())
}

// ToBool converts an object to a bool if possible.
// Returns an error if the object is not a Bool.
func ToBool(obj objects.Object) (bool, error) {
	if b, ok := obj.(*objects.Bool); ok {
		return b.Value, nil
	}
	return false, fmt.Errorf("expected Bool, got %s", obj.Type())
}

// ToArray converts an object to a slice of objects if possible.
// Returns an error if the object is not an Array.
func ToArray(obj objects.Object) ([]objects.Object, error) {
	if a, ok := obj.(*objects.Array); ok {
		return a.Elements, nil
	}
	return nil, fmt.Errorf("expected Array, got %s", obj.Type())
}

// ToMap converts an object to a map of objects if possible.
// Returns an error if the object is not a Map.
func ToMap(obj objects.Object) (map[string]objects.Object, error) {
	if m, ok := obj.(*objects.Map); ok {
		result := make(map[string]objects.Object, len(m.Pairs))
		for _, pair := range m.Pairs {
			result[pair.Key.Inspect()] = pair.Value
		}
		return result, nil
	}
	return nil, fmt.Errorf("expected Map, got %s", obj.Type())
}
