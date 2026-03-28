// pkg/objects/builtin_unicode.go
// Unicode and text processing built-in functions for Xxlang
// Note: toPinYin, kanaToRomaji, kanjiToKana, kanjiToRomaji have been moved to locale module
package objects

import (
	"strings"
)

func init() {
	// Unicode builtins removed - use locale module instead
	// Builtins["toPinYin"] = &Builtin{Fn: builtinToPinYin}
	// Builtins["kanaToRomaji"] = &Builtin{Fn: builtinKanaToRomaji}
	// Builtins["kanjiToKana"] = &Builtin{Fn: builtinKanjiToKana}
	// Builtins["kanjiToRomaji"] = &Builtin{Fn: builtinKanjiToRomaji}
}

// Exported functions for locale module

// ToPinYin converts Chinese characters to Pinyin
func ToPinYin(text string, toneStyle ToneStyle, initialsOnly bool, separator string) string {
	return ConvertToPinyinWithSeparatorEx(text, toneStyle, initialsOnly, separator)
}

// KanaToRomaji converts Hiragana/Katakana to Romaji
func KanaToRomaji(input string) string {
	result := strings.Builder{}
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		str := string(r)

		if i+1 < len(runes) {
			combined := string(runes[i]) + string(runes[i+1])
			if romaji, ok := hiraganaToRomajiMap[combined]; ok {
				result.WriteString(romaji)
				i++
				continue
			}
			if romaji, ok := katakanaToRomajiMap[combined]; ok {
				result.WriteString(romaji)
				i++
				continue
			}
		}

		if romaji, ok := hiraganaToRomajiMap[str]; ok {
			result.WriteString(romaji)
			continue
		}
		if romaji, ok := katakanaToRomajiMap[str]; ok {
			result.WriteString(romaji)
			continue
		}

		result.WriteString(str)
	}

	return result.String()
}

// KanjiToKana converts Kanji to Kana (Hiragana)
func KanjiToKana(input string) string {
	kanjiKeys := make([]string, 0, len(kanjiToKanaMap))
	for kanji := range kanjiToKanaMap {
		kanjiKeys = append(kanjiKeys, kanji)
	}

	for i := 0; i < len(kanjiKeys)-1; i++ {
		for j := i + 1; j < len(kanjiKeys); j++ {
			if len(kanjiKeys[j]) > len(kanjiKeys[i]) {
				kanjiKeys[i], kanjiKeys[j] = kanjiKeys[j], kanjiKeys[i]
			}
		}
	}

	result := input
	for _, kanji := range kanjiKeys {
		kana := kanjiToKanaMap[kanji]
		result = strings.ReplaceAll(result, kanji, kana)
	}

	return result
}

// KanjiToRomaji converts Kanji directly to Romaji
func KanjiToRomaji(input string) string {
	kana := KanjiToKana(input)
	return KanaToRomaji(kana)
}

// builtinToPinYin - convert Chinese characters to Pinyin (kept for reference)
// Usage: toPinYin(text) -> string
//
//	toPinYin(text, "-tone") -> string (with tone marks: zhōng)
//	toPinYin(text, "-toneNum") -> string (with tone numbers: zhong1)
//	toPinYin(text, "-initials") -> string (only initials)
//	toPinYin(text, "-sep=-") -> string (with separator)
//
// Options can be combined: toPinYin(text, "-toneNum-sep=-")
func builtinToPinYin(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for toPinYin. got=%d, want=1 or 2", len(args))
	}

	text, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'toPinYin' must be STRING, got %s", args[0].Type())
	}

	// Parse options
	toneStyle := ToneNone
	initialsOnly := false
	separator := ""

	if len(args) == 2 {
		opts, ok := args[1].(*String)
		if !ok {
			return newError("options must be STRING, got %s", args[1].Type())
		}
		optStr := opts.Value
		if strings.Contains(optStr, "-toneNum") || strings.Contains(optStr, "-tonenum") {
			toneStyle = ToneNumber
		} else if strings.Contains(optStr, "-tone") {
			toneStyle = ToneMark
		}
		if strings.Contains(optStr, "-initials") {
			initialsOnly = true
		}
		// Extract separator from options like "-sep=-"
		if idx := strings.Index(optStr, "-sep="); idx != -1 {
			rest := optStr[idx+5:]
			if len(rest) > 0 {
				separator = string(rest[0])
			}
		}
	}

	// Convert to pinyin using custom implementation
	result := ConvertToPinyinWithSeparatorEx(text.Value, toneStyle, initialsOnly, separator)

	return NewString(result)
}

// hiraganaToRomajiMap maps Hiragana to Romaji
var hiraganaToRomajiMap = map[string]string{
	// Basic vowels
	"あ": "a", "い": "i", "う": "u", "え": "e", "お": "o",
	// K-row
	"か": "ka", "き": "ki", "く": "ku", "け": "ke", "こ": "ko",
	// S-row
	"さ": "sa", "し": "shi", "す": "su", "せ": "se", "そ": "so",
	// T-row
	"た": "ta", "ち": "chi", "つ": "tsu", "て": "te", "と": "to",
	// N-row
	"な": "na", "に": "ni", "ぬ": "nu", "ね": "ne", "の": "no",
	// H-row
	"は": "ha", "ひ": "hi", "ふ": "fu", "へ": "he", "ほ": "ho",
	// M-row
	"ま": "ma", "み": "mi", "む": "mu", "め": "me", "も": "mo",
	// Y-row
	"や": "ya", "ゆ": "yu", "よ": "yo",
	// R-row
	"ら": "ra", "り": "ri", "る": "ru", "れ": "re", "ろ": "ro",
	// W-row
	"わ": "wa", "を": "wo",
	// N
	"ん": "n",
	// G-row (voiced)
	"が": "ga", "ぎ": "gi", "ぐ": "gu", "げ": "ge", "ご": "go",
	// Z-row (voiced)
	"ざ": "za", "じ": "ji", "ず": "zu", "ぜ": "ze", "ぞ": "zo",
	// D-row (voiced)
	"だ": "da", "ぢ": "ji", "づ": "zu", "で": "de", "ど": "do",
	// B-row (voiced)
	"ば": "ba", "び": "bi", "ぶ": "bu", "べ": "be", "ぼ": "bo",
	// P-row (semi-voiced)
	"ぱ": "pa", "ぴ": "pi", "ぷ": "pu", "ぺ": "pe", "ぽ": "po",
	// K-row combined (ya/yu/yo)
	"きゃ": "kya", "きゅ": "kyu", "きょ": "kyo",
	// S-row combined
	"しゃ": "sha", "しゅ": "shu", "しょ": "sho",
	// T-row combined
	"ちゃ": "cha", "ちゅ": "chu", "ちょ": "cho",
	// N-row combined
	"にゃ": "nya", "にゅ": "nyu", "にょ": "nyo",
	// H-row combined
	"ひゃ": "hya", "ひゅ": "hyu", "ひょ": "hyo",
	// M-row combined
	"みゃ": "mya", "みゅ": "myu", "みょ": "myo",
	// R-row combined
	"りゃ": "rya", "りゅ": "ryu", "りょ": "ryo",
	// G-row combined
	"ぎゃ": "gya", "ぎゅ": "gyu", "ぎょ": "gyo",
	// Z-row combined
	"じゃ": "ja", "じゅ": "ju", "じょ": "jo",
	// B-row combined
	"びゃ": "bya", "びゅ": "byu", "びょ": "byo",
	// P-row combined
	"ぴゃ": "pya", "ぴゅ": "pyu", "ぴょ": "pyo",
	// Small tsu (sokuon) - handled specially
	"っ": "",
	// Small ya/yu/yo
	"ゃ": "ya", "ゅ": "yu", "ょ": "yo",
	// Small tsu variants for double consonants
	"っか": "kka", "っき": "kki", "っく": "kku", "っけ": "kke", "っこ": "kko",
	"っさ": "ssa", "っし": "sshi", "っす": "ssu", "っせ": "sse", "っそ": "sso",
	"った": "tta", "っち": "tchi", "っつ": "ttsu", "って": "tte", "っと": "tto",
	"っな": "nna", "っに": "nni", "っぬ": "nnu", "っね": "nne", "っの": "nno",
	"っは": "hha", "っひ": "hhi", "っふ": "ffu", "っへ": "hhe", "っほ": "hho",
	"っま": "mma", "っみ": "mmi", "っむ": "mmu", "っめ": "mme", "っも": "mmo",
	"っや": "yya", "っゆ": "yyu", "っよ": "yyo",
	"っら": "rra", "っり": "rri", "っる": "rru", "っれ": "rre", "っろ": "rro",
	"っわ": "wwa", "っを": "wwo",
	"っが": "gga", "っぎ": "ggi", "っぐ": "ggu", "っげ": "gge", "っご": "ggo",
	"っざ": "zza", "っじ": "jji", "っず": "zzu", "っぜ": "zze", "っぞ": "zzo",
	"っだ": "dda", "っぢ": "jji", "っづ": "zzu", "っで": "dde", "っど": "ddo",
	"っば": "bba", "っび": "bbi", "っぶ": "bbu", "っべ": "bbe", "っぼ": "bbo",
	"っぱ": "ppa", "っぴ": "ppi", "っぷ": "ppu", "っぺ": "ppe", "っぽ": "ppo",
	// Long vowel marker
	"ー": "-",
}

// katakanaToRomajiMap maps Katakana to Romaji
var katakanaToRomajiMap = map[string]string{
	// Basic vowels
	"ア": "a", "イ": "i", "ウ": "u", "エ": "e", "オ": "o",
	// K-row
	"カ": "ka", "キ": "ki", "ク": "ku", "ケ": "ke", "コ": "ko",
	// S-row
	"サ": "sa", "シ": "shi", "ス": "su", "セ": "se", "ソ": "so",
	// T-row
	"タ": "ta", "チ": "chi", "ツ": "tsu", "テ": "te", "ト": "to",
	// N-row
	"ナ": "na", "ニ": "ni", "ヌ": "nu", "ネ": "ne", "ノ": "no",
	// H-row
	"ハ": "ha", "ヒ": "hi", "フ": "fu", "ヘ": "he", "ホ": "ho",
	// M-row
	"マ": "ma", "ミ": "mi", "ム": "mu", "メ": "me", "モ": "mo",
	// Y-row
	"ヤ": "ya", "ユ": "yu", "ヨ": "yo",
	// R-row
	"ラ": "ra", "リ": "ri", "ル": "ru", "レ": "re", "ロ": "ro",
	// W-row
	"ワ": "wa", "ヲ": "wo",
	// N
	"ン": "n",
	// G-row (voiced)
	"ガ": "ga", "ギ": "gi", "グ": "gu", "ゲ": "ge", "ゴ": "go",
	// Z-row (voiced)
	"ザ": "za", "ジ": "ji", "ズ": "zu", "ゼ": "ze", "ゾ": "zo",
	// D-row (voiced)
	"ダ": "da", "ヂ": "ji", "ヅ": "zu", "デ": "de", "ド": "do",
	// B-row (voiced)
	"バ": "ba", "ビ": "bi", "ブ": "bu", "ベ": "be", "ボ": "bo",
	// P-row (semi-voiced)
	"パ": "pa", "ピ": "pi", "プ": "pu", "ペ": "pe", "ポ": "po",
	// Combined forms
	"キャ": "kya", "キュ": "kyu", "キョ": "kyo",
	"シャ": "sha", "シュ": "shu", "ショ": "sho",
	"チャ": "cha", "チュ": "chu", "チョ": "cho",
	"ニャ": "nya", "ニュ": "nyu", "ニョ": "nyo",
	"ヒャ": "hya", "ヒュ": "hyu", "ヒョ": "hyo",
	"ミャ": "mya", "ミュ": "myu", "ミョ": "myo",
	"リャ": "rya", "リュ": "ryu", "リョ": "ryo",
	"ギャ": "gya", "ギュ": "gyu", "ギョ": "gyo",
	"ジャ": "ja", "ジュ": "ju", "ジョ": "jo",
	"ビャ": "bya", "ビュ": "byu", "ビョ": "byo",
	"ピャ": "pya", "ピュ": "pyu", "ピョ": "pyo",
	// Small tsu (sokuon)
	"ッ": "",
	// Small ya/yu/yo
	"ャ": "ya", "ュ": "yu", "ョ": "yo",
	// Small vowels (for foreign words)
	"ァ": "a", "ィ": "i", "ゥ": "u", "ェ": "e", "ォ": "o",
	// Long vowel marker
	"ー": "-",
	// Additional Katakana for foreign words
	"ヴ": "vu", "ヴァ": "va", "ヴィ": "vi", "ヴェ": "ve", "ヴォ": "vo",
	"ファ": "fa", "フィ": "fi", "フェ": "fe", "フォ": "fo",
	"ティ": "ti", "ディ": "di", "トゥ": "tu", "ドゥ": "du",
	"ウィ": "wi", "ウェ": "we", "ウォ": "wo",
	"クァ": "kwa", "クィ": "kwi", "クェ": "kwe", "クォ": "kwo",
	"ツァ": "tsa", "ツィ": "tsi", "ツェ": "tse", "ツォ": "tso",
}

// builtinKanaToRomaji - convert Hiragana/Katakana to Romaji
// Usage: kanaToRomaji(text) -> string
func builtinKanaToRomaji(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for kanaToRomaji. got=%d, want=1", len(args))
	}

	text, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'kanaToRomaji' must be STRING, got %s", args[0].Type())
	}

	input := text.Value
	result := strings.Builder{}
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		str := string(r)

		// Check for combined characters (like きゃ, しゃ, etc.)
		if i+1 < len(runes) {
			combined := string(runes[i]) + string(runes[i+1])
			if romaji, ok := hiraganaToRomajiMap[combined]; ok {
				result.WriteString(romaji)
				i++
				continue
			}
			if romaji, ok := katakanaToRomajiMap[combined]; ok {
				result.WriteString(romaji)
				i++
				continue
			}
		}

		// Check single character
		if romaji, ok := hiraganaToRomajiMap[str]; ok {
			result.WriteString(romaji)
			continue
		}
		if romaji, ok := katakanaToRomajiMap[str]; ok {
			result.WriteString(romaji)
			continue
		}

		// Keep non-kana characters as is
		result.WriteString(str)
	}

	return NewString(result.String())
}

// kanjiToKanaMap provides a basic Kanji to Kana mapping
// Note: This is a simplified mapping. For full conversion, consider using a library.
var kanjiToKanaMap = map[string]string{
	// Common kanji numbers
	"一": "いち", "二": "に", "三": "さん", "四": "よん", "五": "ご",
	"六": "ろく", "七": "なな", "八": "はち", "九": "きゅう", "十": "じゅう",
	"百": "ひゃく", "千": "せん", "万": "まん",

	// Common words
	"日": "にち", "月": "つき", "火": "ひ", "水": "みず", "木": "き", "金": "きん", "土": "つち",
	"人": "じん", "大": "だい", "小": "しょう", "中": "ちゅう", "上": "うえ", "下": "した",
	"前": "まえ", "後": "あと", "左": "ひだり", "右": "みぎ",
	"川": "かわ", "海": "うみ",
	"男": "おとこ", "女": "おんな", "子": "こ", "父": "ちち", "母": "はは",
	"村": "むら",

	// Time related
	"年": "ねん", "時": "じ", "分": "ふん", "間": "かん", "今日": "きょう",
	"明日": "あした", "昨日": "きのう", "朝": "あさ", "昼": "ひる", "夜": "よる",

	// Common verbs (dictionary form)
	"行": "い", "来": "き", "見": "み", "聞": "き", "食": "た", "飲": "の",
	"書": "か", "読": "よ", "話": "はな", "思": "おも", "知": "し",
	"買": "か", "売": "う", "住": "す", "働": "はたら",

	// Common adjectives
	"良": "よ", "悪": "わる", "高": "たか", "安": "やす", "新": "あたら",
	"古": "ふる", "長": "なが", "短": "みじか",

	// Japan related
	"日本": "にほん", "東京": "とうきょう", "京都": "きょうと", "大阪": "おおさか",

	// Common expressions
	"你好": "こんにちは", "谢谢": "ありがとう", "再见": "さようなら",
	"对不起": "すみません", "没关系": "だいじょうぶ",
}

// builtinKanjiToKana - convert Kanji to Kana (Hiragana)
// Usage: kanjiToKana(text) -> string
//
// Note: This uses a basic mapping. For comprehensive conversion,
// use a dedicated library or the kanjiToKana function which may
// call external libraries for better accuracy.
func builtinKanjiToKana(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for kanjiToKana. got=%d, want=1", len(args))
	}

	text, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'kanjiToKana' must be STRING, got %s", args[0].Type())
	}

	input := text.Value

	// Sort kanji keys by length (descending) to match longer compounds first
	kanjiKeys := make([]string, 0, len(kanjiToKanaMap))
	for kanji := range kanjiToKanaMap {
		kanjiKeys = append(kanjiKeys, kanji)
	}

	// Sort by length descending
	for i := 0; i < len(kanjiKeys)-1; i++ {
		for j := i + 1; j < len(kanjiKeys); j++ {
			if len(kanjiKeys[j]) > len(kanjiKeys[i]) {
				kanjiKeys[i], kanjiKeys[j] = kanjiKeys[j], kanjiKeys[i]
			}
		}
	}

	// Replace from longest to shortest
	result := input
	for _, kanji := range kanjiKeys {
		kana := kanjiToKanaMap[kanji]
		result = strings.ReplaceAll(result, kanji, kana)
	}

	return NewString(result)
}

// builtinKanjiToRomaji - convert Kanji directly to Romaji
// Usage: kanjiToRomaji(text) -> string
func builtinKanjiToRomaji(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for kanjiToRomaji. got=%d, want=1", len(args))
	}

	// First convert kanji to kana, then kana to romaji
	kanaResult := builtinKanjiToKana(args...)
	if err, ok := kanaResult.(*Error); ok {
		return err
	}

	kanaStr, ok := kanaResult.(*String)
	if !ok {
		return newError("unexpected result from kanjiToKana")
	}

	// Convert the kana result to romaji
	return builtinKanaToRomaji(kanaStr)
}
