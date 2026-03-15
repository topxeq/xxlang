// pkg/stdlib/strconv.go
// String conversion utilities for the Xxlang standard library.
package stdlib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/strconv",
		Exports: map[string]objects.Object{
			// Integer conversions
			"parseInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("parseInt() takes at least 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseInt() requires a string argument")
				}
				base := int64(10)
				if len(args) > 1 {
					b, ok := args[1].(*objects.Int)
					if ok {
						base = b.Value
					}
				}
				i, err := strconv.ParseInt(s.Value, int(base), 64)
				if err != nil {
					return Error(err.Error())
				}
				return Int(i)
			}),

			"formatInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("formatInt() takes at least 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("formatInt() requires an integer argument")
				}
				base := int64(10)
				if len(args) > 1 {
					b, ok := args[1].(*objects.Int)
					if ok {
						base = b.Value
					}
				}
				return String(strconv.FormatInt(n.Value, int(base)))
			}),

			// Float conversions
			"parseFloat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseFloat() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseFloat() requires a string argument")
				}
				f, err := strconv.ParseFloat(s.Value, 64)
				if err != nil {
					return Error(err.Error())
				}
				return Float(f)
			}),

			"formatFloat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("formatFloat() takes at least 1 argument")
				}
				f, ok := args[0].(*objects.Float)
				if !ok {
					return Error("formatFloat() requires a float argument")
				}
				prec := -1
				if len(args) > 1 {
					p, ok := args[1].(*objects.Int)
					if ok {
						prec = int(p.Value)
					}
				}
				return String(strconv.FormatFloat(f.Value, 'f', prec, 64))
			}),

			// Boolean conversions
			"parseBool": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseBool() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseBool() requires a string argument")
				}
				b, err := strconv.ParseBool(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Bool(b)
			}),

			"formatBool": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("formatBool() takes exactly 1 argument")
				}
				b, ok := args[0].(*objects.Bool)
				if !ok {
					return Error("formatBool() requires a boolean argument")
				}
				return String(strconv.FormatBool(b.Value))
			}),

			// Quote/Unquote
			"quote": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("quote() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("quote() requires a string argument")
				}
				return String(strconv.Quote(s.Value))
			}),

			"unquote": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unquote() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("unquote() requires a string argument")
				}
				u, err := strconv.Unquote(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(u)
			}),

			// Type conversion
			"toString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toString() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Int:
					return String(strconv.FormatInt(v.Value, 10))
				case *objects.Float:
					return String(strconv.FormatFloat(v.Value, 'f', -1, 64))
				case *objects.Bool:
					return String(strconv.FormatBool(v.Value))
				case *objects.String:
					return v
				case *objects.Array:
					parts := make([]string, len(v.Elements))
					for i, e := range v.Elements {
						parts[i] = e.Inspect()
					}
					return String("[" + strings.Join(parts, ", ") + "]")
				case *objects.Map:
					return String(v.Inspect())
				default:
					return String(args[0].Inspect())
				}
			}),

			"toInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toInt() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Int:
					return v
				case *objects.Float:
					return Int(int64(v.Value))
				case *objects.String:
					i, err := strconv.ParseInt(v.Value, 10, 64)
					if err != nil {
						return Error(err.Error())
					}
					return Int(i)
				case *objects.Bool:
					if v.Value {
						return Int(1)
					}
					return Int(0)
				default:
					return Error("cannot convert to int")
				}
			}),

			"toFloat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toFloat() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Int:
					return Float(float64(v.Value))
				case *objects.Float:
					return v
				case *objects.String:
					f, err := strconv.ParseFloat(v.Value, 64)
					if err != nil {
						return Error(err.Error())
					}
					return Float(f)
				case *objects.Bool:
					if v.Value {
						return Float(1.0)
					}
					return Float(0.0)
				default:
					return Error("cannot convert to float")
				}
			}),

			"toBool": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toBool() takes exactly 1 argument")
				}
				switch v := args[0].(type) {
				case *objects.Bool:
					return v
				case *objects.Int:
					return Bool(v.Value != 0)
				case *objects.Float:
					return Bool(v.Value != 0.0)
				case *objects.String:
					b, err := strconv.ParseBool(v.Value)
					if err != nil {
						return Bool(len(v.Value) > 0)
					}
					return Bool(b)
				default:
					return Bool(true)
				}
			}),

			// JSON helpers
			"toJSON": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toJSON() takes exactly 1 argument")
				}
				data, err := json.Marshal(objectToGoValue(args[0]))
				if err != nil {
					return Error(err.Error())
				}
				return String(string(data))
			}),

			"toJSONPretty": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toJSONPretty() takes exactly 1 argument")
				}
				data, err := json.MarshalIndent(objectToGoValue(args[0]), "", "  ")
				if err != nil {
					return Error(err.Error())
				}
				return String(string(data))
			}),

			"formatNumber": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("formatNumber() takes at least 2 arguments")
				}
				var num float64
				switch v := args[0].(type) {
				case *objects.Int:
					num = float64(v.Value)
				case *objects.Float:
					num = v.Value
				default:
					return Error("formatNumber() requires a numeric first argument")
				}
				prec, ok := args[1].(*objects.Int)
				if !ok {
					return Error("formatNumber() requires an integer precision")
				}
				format := fmt.Sprintf("%%.%df", prec.Value)
				return String(fmt.Sprintf(format, num))
			}),

			"formatBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("formatBytes() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("formatBytes() requires an integer argument")
				}
				bytes := float64(n.Value)
				units := []string{"B", "KB", "MB", "GB", "TB"}
				unit := 0
				for bytes >= 1024 && unit < len(units)-1 {
					bytes /= 1024
					unit++
				}
				return String(fmt.Sprintf("%.2f %s", bytes, units[unit]))
			}),

			"formatDuration": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("formatDuration() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("formatDuration() requires an integer argument (milliseconds)")
				}
				ms := n.Value
				if ms < 1000 {
					return String(fmt.Sprintf("%dms", ms))
				}
				sec := ms / 1000
				ms = ms % 1000
				if sec < 60 {
					return String(fmt.Sprintf("%ds %dms", sec, ms))
				}
				min := sec / 60
				sec = sec % 60
				if min < 60 {
					return String(fmt.Sprintf("%dm %ds", min, sec))
				}
				hour := min / 60
				min = min % 60
				return String(fmt.Sprintf("%dh %dm %ds", hour, min, sec))
			}),
		},
	})
}

// objectToGoValue converts an Xxlang object to a Go value for JSON marshaling
func objectToGoValue(obj objects.Object) interface{} {
	switch v := obj.(type) {
	case *objects.Int:
		return v.Value
	case *objects.Float:
		return v.Value
	case *objects.String:
		return v.Value
	case *objects.Bool:
		return v.Value
	case *objects.Null:
		return nil
	case *objects.Array:
		result := make([]interface{}, len(v.Elements))
		for i, e := range v.Elements {
			result[i] = objectToGoValue(e)
		}
		return result
	case *objects.Map:
		result := make(map[string]interface{})
		for _, pair := range v.Pairs {
			key, ok := pair.Key.(*objects.String)
			if ok {
				result[key.Value] = objectToGoValue(pair.Value)
			}
		}
		return result
	default:
		return v.Inspect()
	}
}
