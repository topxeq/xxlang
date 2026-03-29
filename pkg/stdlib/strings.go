// pkg/stdlib/strings.go
// String utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "strings",
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
				runes := []rune(s.Value)
				runeLen := len(runes)
				end := runeLen
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
				if start < 0 || end > runeLen || start > end {
					return Error("substr() index out of range")
				}
				return String(string(runes[start:end]))
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
				byteIdx := strings.Index(s.Value, sub.Value)
				if byteIdx < 0 {
					return Int(-1)
				}
				charIdx := utf8.RuneCountInString(s.Value[:byteIdx])
				return Int(int64(charIdx))
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

			// wordCount counts words in a string.
			// Usage: wordCount(str) -> int
			"wordCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("wordCount() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("wordCount() requires a string argument")
				}
				words := strings.Fields(s.Value)
				return Int(int64(len(words)))
			}),

			// lineCount counts lines in a string.
			// Usage: lineCount(str) -> int
			"lineCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("lineCount() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("lineCount() requires a string argument")
				}
				if s.Value == "" {
					return Int(0)
				}
				count := 1
				for _, c := range s.Value {
					if c == '\n' {
						count++
					}
				}
				return Int(int64(count))
			}),

			// capitalize capitalizes the first letter of a string.
			// Usage: capitalize(str) -> string
			"capitalize": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("capitalize() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("capitalize() requires a string argument")
				}
				if len(s.Value) == 0 {
					return args[0]
				}
				runes := []rune(s.Value)
				runes[0] = unicode.ToUpper(runes[0])
				return String(string(runes))
			}),

			// title converts a string to title case.
			// Usage: title(str) -> string
			"title": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("title() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("title() requires a string argument")
				}
				return String(strings.Title(s.Value))
			}),

			// swapCase swaps the case of each character in a string.
			// Usage: swapCase(str) -> string
			"swapCase": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("swapCase() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("swapCase() requires a string argument")
				}
				runes := []rune(s.Value)
				for i, r := range runes {
					if unicode.IsLower(r) {
						runes[i] = unicode.ToUpper(r)
					} else if unicode.IsUpper(r) {
						runes[i] = unicode.ToLower(r)
					}
				}
				return String(string(runes))
			}),

			// center centers a string with padding.
			// Usage: center(str, width) -> string
			//        center(str, width, padChar) -> string
			"center": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("center() takes 2 or 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("center() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("center() requires an integer as second argument")
				}
				padChar := " "
				if len(args) == 3 {
					p, ok := args[2].(*objects.String)
					if !ok {
						return Error("center() requires a string as third argument")
					}
					if len(p.Value) > 0 {
						padChar = p.Value
					}
				}
				strLen := len(s.Value)
				if strLen >= int(width.Value) {
					return args[0]
				}
				totalPad := int(width.Value) - strLen
				leftPad := totalPad / 2
				rightPad := totalPad - leftPad
				result := strings.Repeat(padChar, leftPad) + s.Value + strings.Repeat(padChar, rightPad)
				return String(result)
			}),

			// zfill pads a string with zeros on the left.
			// Usage: zfill(str, width) -> string
			"zfill": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("zfill() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("zfill() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("zfill() requires an integer as second argument")
				}
				if len(s.Value) >= int(width.Value) {
					return args[0]
				}
				sign := ""
				str := s.Value
				if len(str) > 0 && (str[0] == '+' || str[0] == '-') {
					sign = string(str[0])
					str = str[1:]
				}
				padLen := int(width.Value) - len(s.Value)
				return String(sign + strings.Repeat("0", padLen) + str)
			}),

			// isSpace checks if a string consists only of whitespace.
			// Usage: isSpace(str) -> bool
			"isSpace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSpace() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isSpace() requires a string argument")
				}
				if len(s.Value) == 0 {
					return Bool(false)
				}
				for _, c := range s.Value {
					if !unicode.IsSpace(c) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// bom returns the UTF-8 BOM (Byte Order Mark) string.
			// Usage: bom() -> string
			"bom": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return String("\ufeff")
			}),

			// removeBom removes UTF-8 BOM from the beginning of a string.
			// Usage: removeBom(str) -> string
			"removeBom": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("removeBom() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("removeBom() requires a string argument")
				}
				str := s.Value
				if len(str) >= 3 && str[0] == 0xEF && str[1] == 0xBB && str[2] == 0xBF {
					str = str[3:]
				}
				return String(str)
			}),

			// addBom adds UTF-8 BOM to the beginning of a string.
			// Usage: addBom(str) -> string
			"addBom": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("addBom() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("addBom() requires a string argument")
				}
				str := s.Value
				// Check if BOM already exists
				if len(str) >= 3 && str[0] == 0xEF && str[1] == 0xBB && str[2] == 0xBF {
					return args[0]
				}
				return String("\ufeff" + str)
			}),

			// hasBom checks if a string starts with UTF-8 BOM.
			// Usage: hasBom(str) -> bool
			"hasBom": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("hasBom() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hasBom() requires a string argument")
				}
				str := s.Value
				return Bool(len(str) >= 3 && str[0] == 0xEF && str[1] == 0xBB && str[2] == 0xBF)
			}),
		},
	})
}
