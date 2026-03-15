// pkg/stdlib/fmt.go
// Advanced formatting utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/fmt",
		Exports: map[string]objects.Object{
			// Sprintf-style formatting
			"sprintf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("sprintf() takes at least 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("sprintf() requires a string format")
				}
				// Convert arguments
				goArgs := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					switch v := arg.(type) {
					case *objects.Int:
						goArgs[i] = v.Value
					case *objects.Float:
						goArgs[i] = v.Value
					case *objects.String:
						goArgs[i] = v.Value
					case *objects.Bool:
						goArgs[i] = v.Value
					default:
						goArgs[i] = v.Inspect()
					}
				}
				result := fmt.Sprintf(format.Value, goArgs...)
				return String(result)
			}),

			// Printf-style formatting (returns string for further use)
			"printf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("printf() takes at least 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("printf() requires a string format")
				}
				goArgs := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					switch v := arg.(type) {
					case *objects.Int:
						goArgs[i] = v.Value
					case *objects.Float:
						goArgs[i] = v.Value
					case *objects.String:
						goArgs[i] = v.Value
					case *objects.Bool:
						goArgs[i] = v.Value
					default:
						goArgs[i] = v.Inspect()
					}
				}
				result := fmt.Sprintf(format.Value, goArgs...)
				fmt.Print(result)
				return String(result)
			}),

			// Println style
			"println": BuiltinFunc(func(args ...objects.Object) objects.Object {
				parts := make([]string, len(args))
				for i, arg := range args {
					switch v := arg.(type) {
					case *objects.String:
						parts[i] = v.Value
					case *objects.Int:
						parts[i] = strconv.FormatInt(v.Value, 10)
					case *objects.Float:
						parts[i] = strconv.FormatFloat(v.Value, 'f', -1, 64)
					case *objects.Bool:
						parts[i] = strconv.FormatBool(v.Value)
					default:
						parts[i] = v.Inspect()
					}
				}
				result := strings.Join(parts, " ")
				fmt.Println(result)
				return String(result)
			}),

			// Print style
			"print": BuiltinFunc(func(args ...objects.Object) objects.Object {
				var result strings.Builder
				for _, arg := range args {
					switch v := arg.(type) {
					case *objects.String:
						result.WriteString(v.Value)
					case *objects.Int:
						result.WriteString(strconv.FormatInt(v.Value, 10))
					case *objects.Float:
						result.WriteString(strconv.FormatFloat(v.Value, 'f', -1, 64))
					case *objects.Bool:
						result.WriteString(strconv.FormatBool(v.Value))
					default:
						result.WriteString(v.Inspect())
					}
				}
				fmt.Print(result.String())
				return String(result.String())
			}),

			// Format number with padding
			"padNum": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("padNum() takes at least 3 arguments")
				}
				var num int64
				switch v := args[0].(type) {
				case *objects.Int:
					num = v.Value
				case *objects.Float:
					num = int64(v.Value)
				default:
					return Error("padNum() requires a number as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("padNum() requires an integer width")
				}
				padChar, ok := args[2].(*objects.String)
				if !ok {
					return Error("padNum() requires a string pad character")
				}
				// Build format string like %05d
				format := fmt.Sprintf("%%%s%dd", padChar.Value, width.Value)
				result := fmt.Sprintf(format, num)
				if len(result) > int(width.Value) {
					result = result[len(result)-int(width.Value):]
				}
				return String(result)
			}),

			// Format with left padding
			"lpad": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("lpad() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					s = String(args[0].Inspect())
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("lpad() requires an integer width")
				}
				padChar, ok := args[2].(*objects.String)
				if !ok {
					return Error("lpad() requires a string pad character")
				}
				if len(padChar.Value) == 0 {
					return s
				}
				padLen := int(width.Value) - len(s.Value)
				if padLen <= 0 {
					return s
				}
				padding := strings.Repeat(padChar.Value, (padLen+len(padChar.Value)-1)/len(padChar.Value))
				return String(padding[:padLen] + s.Value)
			}),

			// Format with right padding
			"rpad": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("rpad() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					s = String(args[0].Inspect())
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("rpad() requires an integer width")
				}
				padChar, ok := args[2].(*objects.String)
				if !ok {
					return Error("rpad() requires a string pad character")
				}
				if len(padChar.Value) == 0 {
					return s
				}
				padLen := int(width.Value) - len(s.Value)
				if padLen <= 0 {
					return s
				}
				padding := strings.Repeat(padChar.Value, (padLen+len(padChar.Value)-1)/len(padChar.Value))
				return String(s.Value + padding[:padLen])
			}),

			// Center string
			"center": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("center() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					s = String(args[0].Inspect())
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("center() requires an integer width")
				}
				padChar, ok := args[2].(*objects.String)
				if !ok {
					return Error("center() requires a string pad character")
				}
				if len(padChar.Value) == 0 {
					return s
				}
				padLen := int(width.Value) - len(s.Value)
				if padLen <= 0 {
					return s
				}
				leftPad := padLen / 2
				rightPad := padLen - leftPad
				padding := strings.Repeat(padChar.Value, (int(width.Value)+len(padChar.Value)-1)/len(padChar.Value))
				return String(padding[:leftPad] + s.Value + padding[:rightPad])
			}),

			// Format table
			"table": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("table() takes at least 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("table() requires an array of arrays")
				}
				var result strings.Builder
				for _, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						continue
					}
					parts := make([]string, len(rowArr.Elements))
					for i, cell := range rowArr.Elements {
						parts[i] = cell.Inspect()
					}
					result.WriteString(strings.Join(parts, "\t"))
					result.WriteString("\n")
				}
				return String(result.String())
			}),

			// Format as currency
			"currency": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("currency() takes at least 1 argument")
				}
				var amount float64
				switch v := args[0].(type) {
				case *objects.Int:
					amount = float64(v.Value)
				case *objects.Float:
					amount = v.Value
				default:
					return Error("currency() requires a number")
				}
				symbol := "$"
				if len(args) > 1 {
					s, ok := args[1].(*objects.String)
					if ok {
						symbol = s.Value
					}
				}
				decimals := 2
				if len(args) > 2 {
					d, ok := args[2].(*objects.Int)
					if ok {
						decimals = int(d.Value)
					}
				}
				format := fmt.Sprintf("%%s%%.%df", decimals)
				return String(fmt.Sprintf(format, symbol, amount))
			}),

			// Format percentage
			"percent": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("percent() takes at least 1 argument")
				}
				var val float64
				switch v := args[0].(type) {
				case *objects.Int:
					val = float64(v.Value)
				case *objects.Float:
					val = v.Value
				default:
					return Error("percent() requires a number")
				}
				decimals := 0
				if len(args) > 1 {
					d, ok := args[1].(*objects.Int)
					if ok {
						decimals = int(d.Value)
					}
				}
				format := fmt.Sprintf("%%.%df%%", decimals)
				return String(fmt.Sprintf(format, val*100))
			}),

			// Format hex
			"hex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("hex() takes at least 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("hex() requires an integer")
				}
				prefix := true
				if len(args) > 1 {
					p, ok := args[1].(*objects.Bool)
					if ok {
						prefix = p.Value
					}
				}
				if prefix {
					return String(fmt.Sprintf("0x%x", n.Value))
				}
				return String(fmt.Sprintf("%x", n.Value))
			}),

			// Format binary
			"binary": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("binary() takes at least 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("binary() requires an integer")
				}
				prefix := true
				if len(args) > 1 {
					p, ok := args[1].(*objects.Bool)
					if ok {
						prefix = p.Value
					}
				}
				if prefix {
					return String(fmt.Sprintf("0b%b", n.Value))
				}
				return String(fmt.Sprintf("%b", n.Value))
			}),

			// Format octal
			"octal": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("octal() takes at least 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("octal() requires an integer")
				}
				prefix := true
				if len(args) > 1 {
					p, ok := args[1].(*objects.Bool)
					if ok {
						prefix = p.Value
					}
				}
				if prefix {
					return String(fmt.Sprintf("0o%o", n.Value))
				}
				return String(fmt.Sprintf("%o", n.Value))
			}),

			// Format scientific notation
			"scientific": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("scientific() takes at least 1 argument")
				}
				var val float64
				switch v := args[0].(type) {
				case *objects.Int:
					val = float64(v.Value)
				case *objects.Float:
					val = v.Value
				default:
					return Error("scientific() requires a number")
				}
				prec := 6
				if len(args) > 1 {
					p, ok := args[1].(*objects.Int)
					if ok {
						prec = int(p.Value)
					}
				}
				return String(fmt.Sprintf("%.*e", prec, val))
			}),

			// Format with commas (thousands separator)
			"commas": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("commas() takes at least 1 argument")
				}
				var num int64
				switch v := args[0].(type) {
				case *objects.Int:
					num = v.Value
				case *objects.Float:
					num = int64(v.Value)
				default:
					return Error("commas() requires a number")
				}
				sep := ","
				if len(args) > 1 {
					s, ok := args[1].(*objects.String)
					if ok {
						sep = s.Value
					}
				}
				str := strconv.FormatInt(num, 10)
				if num < 0 {
					str = str[1:]
				}
				var result strings.Builder
				for i, c := range str {
					if i > 0 && (len(str)-i)%3 == 0 {
						result.WriteString(sep)
					}
					result.WriteRune(c)
				}
				if num < 0 {
					return String("-" + result.String())
				}
				return String(result.String())
			}),

			// Template formatting with named placeholders
			"template": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("template() takes at least 2 arguments")
				}
				tmpl, ok := args[0].(*objects.String)
				if !ok {
					return Error("template() requires a string template")
				}
				data, ok := args[1].(*objects.Map)
				if !ok {
					return Error("template() requires a map as second argument")
				}
				result := tmpl.Value
				for _, pair := range data.Pairs {
					key, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}
					placeholder := fmt.Sprintf("{{%s}}", key.Value)
					var val string
					switch v := pair.Value.(type) {
					case *objects.String:
						val = v.Value
					case *objects.Int:
						val = strconv.FormatInt(v.Value, 10)
					case *objects.Float:
						val = strconv.FormatFloat(v.Value, 'f', -1, 64)
					case *objects.Bool:
						val = strconv.FormatBool(v.Value)
					default:
						val = v.Inspect()
					}
					result = strings.ReplaceAll(result, placeholder, val)
				}
				return String(result)
			}),

			// Format key-value pairs
			"kv": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("kv() takes at least 1 argument")
				}
				data, ok := args[0].(*objects.Map)
				if !ok {
					return Error("kv() requires a map")
				}
				sep := ": "
				if len(args) > 1 {
					s, ok := args[1].(*objects.String)
					if ok {
						sep = s.Value
					}
				}
				lineSep := "\n"
				if len(args) > 2 {
					s, ok := args[2].(*objects.String)
					if ok {
						lineSep = s.Value
					}
				}
				var lines []string
				for _, pair := range data.Pairs {
					key, ok := pair.Key.(*objects.String)
					if !ok {
						continue
					}
					var val string
					switch v := pair.Value.(type) {
					case *objects.String:
						val = v.Value
					case *objects.Int:
						val = strconv.FormatInt(v.Value, 10)
					case *objects.Float:
						val = strconv.FormatFloat(v.Value, 'f', -1, 64)
					case *objects.Bool:
						val = strconv.FormatBool(v.Value)
					default:
						val = v.Inspect()
					}
					lines = append(lines, key.Value+sep+val)
				}
				return String(strings.Join(lines, lineSep))
			}),
		},
	})
}
