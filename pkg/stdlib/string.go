// pkg/stdlib/string.go
// String utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/string",
		Exports: map[string]objects.Object{
			"len": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("len() takes exactly 1 argument")
				}
				switch s := args[0].(type) {
				case *objects.String:
					return Int(int64(len(s.Value)))
				default:
					return Error("len() requires a string argument")
				}
			}),

			"substr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("substr() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("substr() requires a string as first argument")
				}
				var start int
				switch v := args[1].(type) {
				case *objects.Int:
					start = int(v.Value)
				case *objects.Float:
					start = int(v.Value)
				default:
					return Error("start position must be an integer")
				}
				end := len(s.Value)
				if len(args) > 2 {
					switch v := args[2].(type) {
					case *objects.Int:
						end = int(v.Value)
					case *objects.Float:
						end = int(v.Value)
					default:
						return Error("end position must be an integer")
					}
				}
				if start < 0 || end > len(s.Value) || start > end {
					return Error("substr() index out of range")
				}
				return String(s.Value[start:end])
			}),

			"indexOf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("indexOf() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("indexOf() requires a string as first argument")
				}
				sub, ok := args[1].(*objects.String)
				if !ok {
					return Error("indexOf() requires a string as second argument")
				}
				return Int(int64(strings.Index(s.Value, sub.Value)))
			}),

			"contains": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("contains() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("contains() requires a string as first argument")
				}
				sub, ok := args[1].(*objects.String)
				if !ok {
					return Error("contains() requires a string as second argument")
				}
				return Bool(strings.Contains(s.Value, sub.Value))
			}),

			"hasPrefix": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("hasPrefix() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hasPrefix() requires a string as first argument")
				}
				prefix, ok := args[1].(*objects.String)
				if !ok {
					return Error("hasPrefix() requires a string as second argument")
				}
				return Bool(strings.HasPrefix(s.Value, prefix.Value))
			}),

			"hasSuffix": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("hasSuffix() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hasSuffix() requires a string as first argument")
				}
				suffix, ok := args[1].(*objects.String)
				if !ok {
					return Error("hasSuffix() requires a string as second argument")
				}
				return Bool(strings.HasSuffix(s.Value, suffix.Value))
			}),

			"toUpper": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toUpper() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("toUpper() requires a string argument")
				}
				return String(strings.ToUpper(s.Value))
			}),

			"toLower": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toLower() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("toLower() requires a string argument")
				}
				return String(strings.ToLower(s.Value))
			}),

			"trim": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("trim() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("trim() requires a string argument")
				}
				return String(strings.TrimSpace(s.Value))
			}),

			"trimSpace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("trimSpace() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("trimSpace() requires a string argument")
				}
				return String(strings.TrimSpace(s.Value))
			}),

			"split": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("split() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("split() requires a string as first argument")
				}
				sep, ok := args[1].(*objects.String)
				if !ok {
					return Error("split() requires a string as second argument")
				}
				parts := strings.Split(s.Value, sep.Value)
				result := make([]objects.Object, len(parts))
				for i, p := range parts {
					result[i] = String(p)
				}
				return Array(result...)
			}),

			"join": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("join() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("join() requires an array as first argument")
				}
				sep, ok := args[1].(*objects.String)
				if !ok {
					return Error("join() requires a string as second argument")
				}
				strs := make([]string, len(arr.Elements))
				for i, elem := range arr.Elements {
					str, ok := elem.(*objects.String)
					if !ok {
						return Error("join() requires all array elements to be strings")
					}
					strs[i] = str.Value
				}
				return String(strings.Join(strs, sep.Value))
			}),

			"repeat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("repeat() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("repeat() requires a string as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("repeat() requires an integer as second argument")
				}
				if n.Value < 0 {
					return Error("repeat() count must be non-negative")
				}
				return String(strings.Repeat(s.Value, int(n.Value)))
			}),

			"replace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("replace() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("replace() requires a string as first argument")
				}
				old, ok := args[1].(*objects.String)
				if !ok {
					return Error("replace() requires a string as second argument")
				}
				newStr, ok := args[2].(*objects.String)
				if !ok {
					return Error("replace() requires a string as third argument")
				}
				return String(strings.ReplaceAll(s.Value, old.Value, newStr.Value))
			}),

			"parseInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseInt() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseInt() requires a string argument")
				}
				i, err := strconv.ParseInt(s.Value, 0, 64)
				if err != nil {
					return Error(fmt.Sprintf("parseInt() failed: %s", err.Error()))
				}
				return Int(i)
			}),

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
					return Error(fmt.Sprintf("parseFloat() failed: %s", err.Error()))
				}
				return Float(f)
			}),

			"toString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toString() takes exactly 1 argument")
				}
				switch arg := args[0].(type) {
				case *objects.Int:
					return String(strconv.FormatInt(arg.Value, 10))
				case *objects.Float:
					return String(strconv.FormatFloat(arg.Value, 'f', -1, 64))
				case *objects.Bool:
					if arg.Value {
						return String("true")
					}
					return String("false")
				case *objects.String:
					return arg
				default:
					return Error("toString() requires a primitive argument")
				}
			}),

			"reverse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("reverse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("reverse() requires a string argument")
				}
				runes := []rune(s.Value)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return String(string(runes))
			}),

			"format": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("format() takes at least 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("format() requires a string as first argument")
				}
				// Simple concatenation for now
				var buf strings.Builder
				for i := 1; i < len(args); i++ {
					switch a := args[i].(type) {
					case *objects.String:
						buf.WriteString(a.Value)
					case *objects.Int:
						buf.WriteString(strconv.FormatInt(a.Value, 10))
					case *objects.Float:
						buf.WriteString(strconv.FormatFloat(a.Value, 'f', -1, 64))
					case *objects.Bool:
						if a.Value {
							buf.WriteString("true")
						} else {
							buf.WriteString("false")
						}
					default:
						buf.WriteString(fmt.Sprintf("%v", a))
					}
				}
				// If format string contains placeholders, use it; otherwise just concatenate
				result := buf.String()
				if strings.Contains(format.Value, "{}") || strings.Contains(format.Value, "%") {
					// For now, just return the format string
					return String(format.Value)
				}
				return String(result)
			}),
		},
	})
}
