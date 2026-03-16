// pkg/stdlib/utils.go
// Utility functions for the Xxlang standard library.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "utils",
		Exports: map[string]objects.Object{
			// deepCopy creates a deep copy of any object
			"deepCopy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("deepCopy() takes exactly 1 argument")
				}
				return deepCopyValue(args[0])
			}),

			// shallowCopy creates a shallow copy of arrays and maps
			"shallowCopy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("shallowCopy() takes exactly 1 argument")
				}
				return shallowCopyValue(args[0])
			}),

			// deepMerge deeply merges two maps
			"deepMerge": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("deepMerge() takes exactly 2 arguments")
				}
				target, ok := args[0].(*objects.Map)
				if !ok {
					return Error("deepMerge() requires a map as first argument")
				}
				source, ok := args[1].(*objects.Map)
				if !ok {
					return Error("deepMerge() requires a map as second argument")
				}
				return deepMergeMaps(target, source)
			}),

			// deepEquals compares two objects for deep equality
			"deepEquals": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("deepEquals() takes exactly 2 arguments")
				}
				return Bool(deepCompare(args[0], args[1]))
			}),

			// pick selects specified keys from a map
			"pick": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("pick() takes exactly 2 arguments")
				}
				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("pick() requires a map as first argument")
				}
				keys, ok := args[1].(*objects.Array)
				if !ok {
					return Error("pick() requires an array as second argument")
				}
				return pickKeys(m, keys)
			}),

			// omit excludes specified keys from a map
			"omit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("omit() takes exactly 2 arguments")
				}
				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("omit() requires a map as first argument")
				}
				keys, ok := args[1].(*objects.Array)
				if !ok {
					return Error("omit() requires an array as second argument")
				}
				return omitKeys(m, keys)
			}),

			// keys returns all keys of a map as an array
			"keys": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("keys() takes exactly 1 argument")
				}
				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("keys() requires a map argument")
				}
				result := make([]objects.Object, 0, len(m.Pairs))
				for _, pair := range m.Pairs {
					result = append(result, pair.Key)
				}
				return Array(result...)
			}),

			// values returns all values of a map as an array
			"values": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("values() takes exactly 1 argument")
				}
				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("values() requires a map argument")
				}
				result := make([]objects.Object, 0, len(m.Pairs))
				for _, pair := range m.Pairs {
					result = append(result, pair.Value)
				}
				return Array(result...)
			}),

			// entries returns key-value pairs of a map as an array of [key, value]
			"entries": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("entries() takes exactly 1 argument")
				}
				m, ok := args[0].(*objects.Map)
				if !ok {
					return Error("entries() requires a map argument")
				}
				result := make([]objects.Object, 0, len(m.Pairs))
				for _, pair := range m.Pairs {
					result = append(result, Array(pair.Key, pair.Value))
				}
				return Array(result...)
			}),

			// fromEntries converts an array of [key, value] pairs to a map
			"fromEntries": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromEntries() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("fromEntries() requires an array argument")
				}
				pairs := make(map[objects.HashKey]objects.MapPair)
				for _, elem := range arr.Elements {
					pair, ok := elem.(*objects.Array)
					if !ok || len(pair.Elements) < 2 {
						continue
					}
					key := pair.Elements[0]
					value := pair.Elements[1]
					pairs[key.HashKey()] = objects.MapPair{Key: key, Value: value}
				}
				return &objects.Map{Pairs: pairs}
			}),

			// type returns the type name of an object
			"type": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("type() takes exactly 1 argument")
				}
				return String(string(args[0].Type()))
			}),

			// isPrimitive checks if a value is a primitive type
			"isPrimitive": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isPrimitive() takes exactly 1 argument")
				}
				switch args[0].(type) {
				case *objects.Int, *objects.Float, *objects.String, *objects.Bool, *objects.Null:
					return Bool(true)
				default:
					return Bool(false)
				}
			}),

			// isEmpty checks if a value is empty
			"isEmpty": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isEmpty() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.String:
					return Bool(len(v.Value) == 0)
				case *objects.Array:
					return Bool(len(v.Elements) == 0)
				case *objects.Map:
					return Bool(len(v.Pairs) == 0)
				case *objects.Null:
					return Bool(true)
				default:
					return Bool(false)
				}
			}),

			// size returns the size of a collection
			"size": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("size() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.String:
					return Int(int64(len(v.Value)))
				case *objects.Array:
					return Int(int64(len(v.Elements)))
				case *objects.Map:
					return Int(int64(len(v.Pairs)))
				default:
					return Error("size() requires a string, array, or map argument")
				}
			}),
		},
	})
}

// deepCopyValue creates a deep copy of an object
func deepCopyValue(obj objects.Object) objects.Object {
	switch o := obj.(type) {
	case *objects.Int:
		return Int(o.Value)
	case *objects.Float:
		return Float(o.Value)
	case *objects.String:
		return String(o.Value)
	case *objects.Bool:
		if o.Value {
			return Bool(true)
		}
		return Bool(false)
	case *objects.Null:
		return Null()
	case *objects.Array:
		elements := make([]objects.Object, len(o.Elements))
		for i, elem := range o.Elements {
			elements[i] = deepCopyValue(elem)
		}
		return Array(elements...)
	case *objects.Map:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for k, v := range o.Pairs {
			pairs[k] = objects.MapPair{
				Key:   deepCopyValue(v.Key),
				Value: deepCopyValue(v.Value),
			}
		}
		return &objects.Map{Pairs: pairs}
	default:
		return obj
	}
}

// shallowCopyValue creates a shallow copy of an object
func shallowCopyValue(obj objects.Object) objects.Object {
	switch o := obj.(type) {
	case *objects.Array:
		elements := make([]objects.Object, len(o.Elements))
		copy(elements, o.Elements)
		return Array(elements...)
	case *objects.Map:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for k, v := range o.Pairs {
			pairs[k] = v
		}
		return &objects.Map{Pairs: pairs}
	default:
		return obj
	}
}

// deepMergeMaps deeply merges two maps
func deepMergeMaps(target, source *objects.Map) *objects.Map {
	result := make(map[objects.HashKey]objects.MapPair)

	// Copy target
	for k, v := range target.Pairs {
		result[k] = v
	}

	// Merge source
	for k, sourcePair := range source.Pairs {
		if targetPair, exists := result[k]; exists {
			// Both exist - try to merge if both are maps
			targetMap, targetIsMap := targetPair.Value.(*objects.Map)
			sourceMap, sourceIsMap := sourcePair.Value.(*objects.Map)
			if targetIsMap && sourceIsMap {
				merged := deepMergeMaps(targetMap, sourceMap)
				result[k] = objects.MapPair{Key: sourcePair.Key, Value: merged}
			} else {
				// Overwrite with source value
				result[k] = sourcePair
			}
		} else {
			result[k] = sourcePair
		}
	}

	return &objects.Map{Pairs: result}
}

// deepCompare compares two objects for deep equality
func deepCompare(a, b objects.Object) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type() != b.Type() {
		return false
	}

	switch aObj := a.(type) {
	case *objects.Int:
		bObj, ok := b.(*objects.Int)
		return ok && aObj.Value == bObj.Value
	case *objects.Float:
		bObj, ok := b.(*objects.Float)
		return ok && aObj.Value == bObj.Value
	case *objects.String:
		bObj, ok := b.(*objects.String)
		return ok && aObj.Value == bObj.Value
	case *objects.Bool:
		bObj, ok := b.(*objects.Bool)
		return ok && aObj.Value == bObj.Value
	case *objects.Null:
		return true
	case *objects.Array:
		bObj, ok := b.(*objects.Array)
		if !ok || len(aObj.Elements) != len(bObj.Elements) {
			return false
		}
		for i := range aObj.Elements {
			if !deepCompare(aObj.Elements[i], bObj.Elements[i]) {
				return false
			}
		}
		return true
	case *objects.Map:
		bObj, ok := b.(*objects.Map)
		if !ok || len(aObj.Pairs) != len(bObj.Pairs) {
			return false
		}
		for k, v := range aObj.Pairs {
			bVal, exists := bObj.Pairs[k]
			if !exists || !deepCompare(v.Value, bVal.Value) {
				return false
			}
		}
		return true
	default:
		return a.Inspect() == b.Inspect()
	}
}

// pickKeys selects specified keys from a map
func pickKeys(m *objects.Map, keys *objects.Array) *objects.Map {
	keySet := make(map[objects.HashKey]bool)
	for _, k := range keys.Elements {
		keySet[k.HashKey()] = true
	}

	pairs := make(map[objects.HashKey]objects.MapPair)
	for k, v := range m.Pairs {
		if keySet[k] {
			pairs[k] = v
		}
	}

	return &objects.Map{Pairs: pairs}
}

// omitKeys excludes specified keys from a map
func omitKeys(m *objects.Map, keys *objects.Array) *objects.Map {
	keySet := make(map[objects.HashKey]bool)
	for _, k := range keys.Elements {
		keySet[k.HashKey()] = true
	}

	pairs := make(map[objects.HashKey]objects.MapPair)
	for k, v := range m.Pairs {
		if !keySet[k] {
			pairs[k] = v
		}
	}

	return &objects.Map{Pairs: pairs}
}
