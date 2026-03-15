// pkg/stdlib/text.go
// Text processing utilities for the Xxlang standard library.
package stdlib

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/text",
		Exports: map[string]objects.Object{
			// Word wrap text
			"wordWrap": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("wordWrap() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("wordWrap() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("wordWrap() requires an integer width")
				}
				words := strings.Fields(s.Value)
				var result strings.Builder
				lineLen := 0
				for i, word := range words {
					wordLen := utf8.RuneCountInString(word)
					if i == 0 {
						result.WriteString(word)
						lineLen = wordLen
					} else if lineLen+1+wordLen > int(width.Value) {
						result.WriteString("\n")
						result.WriteString(word)
						lineLen = wordLen
					} else {
						result.WriteString(" ")
						result.WriteString(word)
						lineLen += 1 + wordLen
					}
				}
				return String(result.String())
			}),

			// Truncate text
			"truncate": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("truncate() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("truncate() requires a string as first argument")
				}
				maxLen, ok := args[1].(*objects.Int)
				if !ok {
					return Error("truncate() requires an integer length")
				}
				suffix := "..."
				if len(args) > 2 {
					suf, ok := args[2].(*objects.String)
					if ok {
						suffix = suf.Value
					}
				}
				runes := []rune(s.Value)
				if len(runes) <= int(maxLen.Value) {
					return s
				}
				return String(string(runes[:int(maxLen.Value)-len(suffix)]) + suffix)
			}),

			// Count words
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

			// Count lines
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
				count := strings.Count(s.Value, "\n") + 1
				return Int(int64(count))
			}),

			// Count characters (runes)
			"charCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("charCount() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("charCount() requires a string argument")
				}
				return Int(int64(utf8.RuneCountInString(s.Value)))
			}),

			// Count bytes
			"byteCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("byteCount() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("byteCount() requires a string argument")
				}
				return Int(int64(len(s.Value)))
			}),

			// Get lines
			"lines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("lines() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("lines() requires a string argument")
				}
				lines := strings.Split(s.Value, "\n")
				result := make([]objects.Object, len(lines))
				for i, line := range lines {
					result[i] = String(line)
				}
				return Array(result...)
			}),

			// Join lines
			"joinLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("joinLines() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("joinLines() requires an array argument")
				}
				lines := make([]string, len(arr.Elements))
				for i, elem := range arr.Elements {
					s, ok := elem.(*objects.String)
					if !ok {
						return Error("joinLines() requires string array elements")
					}
					lines[i] = s.Value
				}
				return String(strings.Join(lines, "\n"))
			}),

			// Get words
			"words": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("words() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("words() requires a string argument")
				}
				words := strings.Fields(s.Value)
				result := make([]objects.Object, len(words))
				for i, word := range words {
					result[i] = String(word)
				}
				return Array(result...)
			}),

			// Get characters
			"chars": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("chars() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("chars() requires a string argument")
				}
				runes := []rune(s.Value)
				result := make([]objects.Object, len(runes))
				for i, r := range runes {
					result[i] = String(string(r))
				}
				return Array(result...)
			}),

			// Title case
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

			// Capitalize first letter
			"capitalize": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("capitalize() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("capitalize() requires a string argument")
				}
				if s.Value == "" {
					return s
				}
				runes := []rune(s.Value)
				runes[0] = unicode.ToUpper(runes[0])
				return String(string(runes))
			}),

			// Swap case
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
					if unicode.IsUpper(r) {
						runes[i] = unicode.ToLower(r)
					} else if unicode.IsLower(r) {
						runes[i] = unicode.ToUpper(r)
					}
				}
				return String(string(runes))
			}),

			// Is alphanumeric
			"isAlphaNum": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isAlphaNum() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isAlphaNum() requires a string argument")
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return Bool(false)
					}
				}
				return Bool(len(s.Value) > 0)
			}),

			// Is alphabetic
			"isAlpha": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isAlpha() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isAlpha() requires a string argument")
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) {
						return Bool(false)
					}
				}
				return Bool(len(s.Value) > 0)
			}),

			// Is numeric
			"isNumeric": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isNumeric() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isNumeric() requires a string argument")
				}
				for _, r := range s.Value {
					if !unicode.IsDigit(r) {
						return Bool(false)
					}
				}
				return Bool(len(s.Value) > 0)
			}),

			// Is whitespace
			"isSpace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSpace() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isSpace() requires a string argument")
				}
				for _, r := range s.Value {
					if !unicode.IsSpace(r) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Is empty or whitespace
			"isBlank": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isBlank() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isBlank() requires a string argument")
				}
				return Bool(strings.TrimSpace(s.Value) == "")
			}),

			// Remove all whitespace
			"removeSpaces": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("removeSpaces() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("removeSpaces() requires a string argument")
				}
				var result strings.Builder
				for _, r := range s.Value {
					if !unicode.IsSpace(r) {
						result.WriteRune(r)
					}
				}
				return String(result.String())
			}),

			// Normalize whitespace
			"normalizeSpace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("normalizeSpace() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("normalizeSpace() requires a string argument")
				}
				// Split and rejoin to normalize all whitespace
				words := strings.Fields(s.Value)
				return String(strings.Join(words, " "))
			}),

			// Pad left
			"padLeft": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("padLeft() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("padLeft() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("padLeft() requires an integer width")
				}
				pad, ok := args[2].(*objects.String)
				if !ok {
					return Error("padLeft() requires a string pad")
				}
				runes := []rune(s.Value)
				padRunes := []rune(pad.Value)
				if len(padRunes) == 0 {
					return s
				}
				for len(runes) < int(width.Value) {
					runes = append(padRunes, runes...)
				}
				if len(runes) > int(width.Value) {
					runes = runes[len(runes)-int(width.Value):]
				}
				return String(string(runes))
			}),

			// Pad right
			"padRight": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("padRight() takes at least 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("padRight() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("padRight() requires an integer width")
				}
				pad, ok := args[2].(*objects.String)
				if !ok {
					return Error("padRight() requires a string pad")
				}
				runes := []rune(s.Value)
				padRunes := []rune(pad.Value)
				if len(padRunes) == 0 {
					return s
				}
				for len(runes) < int(width.Value) {
					runes = append(runes, padRunes...)
				}
				if len(runes) > int(width.Value) {
					runes = runes[:int(width.Value)]
				}
				return String(string(runes))
			}),

			// Indent lines
			"indent": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("indent() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("indent() requires a string as first argument")
				}
				indent, ok := args[1].(*objects.String)
				if !ok {
					return Error("indent() requires a string indent")
				}
				lines := strings.Split(s.Value, "\n")
				for i, line := range lines {
					if line != "" {
						lines[i] = indent.Value + line
					}
				}
				return String(strings.Join(lines, "\n"))
			}),

			// Dedent lines
			"dedent": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("dedent() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("dedent() requires a string argument")
				}
				lines := strings.Split(s.Value, "\n")
				minIndent := -1
				for _, line := range lines {
					if line == "" {
						continue
					}
					space := 0
					for _, r := range line {
						if r == ' ' || r == '\t' {
							space++
						} else {
							break
						}
					}
					if minIndent == -1 || space < minIndent {
						minIndent = space
					}
				}
				if minIndent <= 0 {
					return s
				}
				for i, line := range lines {
					if len(line) >= minIndent {
						lines[i] = line[minIndent:]
					}
				}
				return String(strings.Join(lines, "\n"))
			}),

			// Center text in width
			"centerText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("centerText() takes at least 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("centerText() requires a string as first argument")
				}
				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("centerText() requires an integer width")
				}
				runes := []rune(s.Value)
				sLen := len(runes)
				w := int(width.Value)
				if sLen >= w {
					return s
				}
				left := (w - sLen) / 2
				right := w - sLen - left
				return String(strings.Repeat(" ", left) + s.Value + strings.Repeat(" ", right))
			}),

			// Repeat string
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
					return Error("repeat() requires an integer count")
				}
				if n.Value < 0 {
					return Error("repeat() count must be non-negative")
				}
				return String(strings.Repeat(s.Value, int(n.Value)))
			}),

			// Get character at index
			"charAt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("charAt() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("charAt() requires a string as first argument")
				}
				idx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("charAt() requires an integer index")
				}
				runes := []rune(s.Value)
				i := int(idx.Value)
				if i < 0 || i >= len(runes) {
					return Error("charAt() index out of range")
				}
				return String(string(runes[i]))
			}),

			// Get character code
			"charCode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("charCode() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("charCode() requires a string as first argument")
				}
				idx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("charCode() requires an integer index")
				}
				runes := []rune(s.Value)
				i := int(idx.Value)
				if i < 0 || i >= len(runes) {
					return Error("charCode() index out of range")
				}
				return Int(int64(runes[i]))
			}),

			// Character from code
			"fromCode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromCode() takes exactly 1 argument")
				}
				code, ok := args[0].(*objects.Int)
				if !ok {
					return Error("fromCode() requires an integer code")
				}
				return String(string(rune(code.Value)))
			}),

			// Escape for shell
			"shellEscape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("shellEscape() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("shellEscape() requires a string argument")
				}
				// Simple shell escaping - wrap in single quotes and escape single quotes
				escaped := strings.ReplaceAll(s.Value, "'", "'\"'\"'")
				return String("'" + escaped + "'")
			}),

			// Escape for JSON string
			"jsonEscape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("jsonEscape() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("jsonEscape() requires a string argument")
				}
				// Use strconv.Quote for proper JSON escaping
				quoted := strings.Trim(strconv.Quote(s.Value), `"`)
				return String(quoted)
			}),

			// Unescape JSON string
			"jsonUnescape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("jsonUnescape() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("jsonUnescape() requires a string argument")
				}
				unquoted, err := strconv.Unquote(`"` + s.Value + `"`)
				if err != nil {
					return Error(err.Error())
				}
				return String(unquoted)
			}),
		},
	})
}
