// pkg/stdlib/locale.go
// Locale module for Xxlang - provides language/region specific text processing functions
package stdlib

import (
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "locale",
		Exports: map[string]objects.Object{
			// toPinYin - convert Chinese characters to Pinyin
			// Usage: locale.toPinYin(text) -> string
			//         locale.toPinYin(text, "-tone") -> string (with tone marks)
			//         locale.toPinYin(text, "-toneNum") -> string (with tone numbers)
			//         locale.toPinYin(text, "-initials") -> string (only initials)
			//         locale.toPinYin(text, "-sep=-") -> string (with separator)
			"toPinYin": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("toPinYin() takes 1 or 2 arguments")
				}

				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("toPinYin() requires a string as first argument")
				}

				toneStyle := objects.ToneNone
				initialsOnly := false
				separator := ""

				if len(args) == 2 {
					opts, ok := args[1].(*objects.String)
					if !ok {
						return Error("options must be STRING")
					}
					optStr := opts.Value
					if strings.Contains(optStr, "-toneNum") || strings.Contains(optStr, "-tonenum") {
						toneStyle = objects.ToneNumber
					} else if strings.Contains(optStr, "-tone") {
						toneStyle = objects.ToneMark
					}
					if strings.Contains(optStr, "-initials") {
						initialsOnly = true
					}
					if idx := strings.Index(optStr, "-sep="); idx != -1 {
						rest := optStr[idx+5:]
						if len(rest) > 0 {
							separator = string(rest[0])
						}
					}
				}

				result := objects.ToPinYin(text.Value, toneStyle, initialsOnly, separator)
				return String(result)
			}),

			// kanaToRomaji - convert Hiragana/Katakana to Romaji
			// Usage: locale.kanaToRomaji(text) -> string
			"kanaToRomaji": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("kanaToRomaji() takes exactly 1 argument")
				}

				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("kanaToRomaji() requires a string argument")
				}

				result := objects.KanaToRomaji(text.Value)
				return String(result)
			}),

			// kanjiToKana - convert Kanji to Kana (Hiragana)
			// Usage: locale.kanjiToKana(text) -> string
			"kanjiToKana": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("kanjiToKana() takes exactly 1 argument")
				}

				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("kanjiToKana() requires a string argument")
				}

				result := objects.KanjiToKana(text.Value)
				return String(result)
			}),

			// kanjiToRomaji - convert Kanji directly to Romaji
			// Usage: locale.kanjiToRomaji(text) -> string
			"kanjiToRomaji": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("kanjiToRomaji() takes exactly 1 argument")
				}

				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("kanjiToRomaji() requires a string argument")
				}

				result := objects.KanjiToRomaji(text.Value)
				return String(result)
			}),
		},
	})
}
