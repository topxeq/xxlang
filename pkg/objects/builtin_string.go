// pkg/objects/builtin_string.go
// String enhancement built-in functions for Xxlang
package objects

import (
	"strings"
	"unicode/utf8"
)

func init() {
	// String enhancement functions
	Builtins["strContainsIn"] = &Builtin{Fn: builtinStrContainsIn}
	Builtins["strRuneLen"] = &Builtin{Fn: builtinStrRuneLen}
	Builtins["strIn"] = &Builtin{Fn: builtinStrIn}
	Builtins["strGetLastComponent"] = &Builtin{Fn: builtinStrGetLastComponent}
	Builtins["strFindDiffPos"] = &Builtin{Fn: builtinStrFindDiffPos}
	Builtins["strDiff"] = &Builtin{Fn: builtinStrDiff}
	Builtins["strFindAllSub"] = &Builtin{Fn: builtinStrFindAllSub}
	Builtins["limitStr"] = &Builtin{Fn: builtinLimitStr}
	Builtins["strQuote"] = &Builtin{Fn: builtinStrQuote}
	Builtins["strUnquote"] = &Builtin{Fn: builtinStrUnquote}
	Builtins["strToInt"] = &Builtin{Fn: builtinStrToInt}
	Builtins["getTextSimilarity"] = &Builtin{Fn: builtinGetTextSimilarity}
	Builtins["fuzzyFind"] = &Builtin{Fn: builtinFuzzyFind}
	// strRemoveBom moved to strings module: strings.removeBom(), strings.addBom(), strings.bom()
	Builtins["strReverse"] = &Builtin{Fn: builtinReverseStr}
}

// builtinStrContainsIn - check if string contains any of the substrings
// Usage: strContainsIn(str, substrs...) -> bool
func builtinStrContainsIn(args ...Object) Object {
	if len(args) < 2 {
		return newError("wrong number of arguments for strContainsIn. got=%d, want>=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strContainsIn' must be STRING, got %s", args[0].Type())
	}

	for i := 1; i < len(args); i++ {
		substr, ok := args[i].(*String)
		if !ok {
			continue
		}
		if strings.Contains(str.Value, substr.Value) {
			return TRUE
		}
	}

	return FALSE
}

// builtinStrRuneLen - get Unicode character count
// Usage: strRuneLen(str) -> int
func builtinStrRuneLen(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for strRuneLen. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'strRuneLen' must be STRING, got %s", args[0].Type())
	}

	return NewInt(int64(utf8.RuneCountInString(str.Value)))
}

// builtinStrIn - check if string is in array
// Usage: strIn(str, arr) -> bool
func builtinStrIn(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strIn. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strIn' must be STRING, got %s", args[0].Type())
	}

	arr, ok := args[1].(*Array)
	if !ok {
		return newError("second argument to 'strIn' must be ARRAY, got %s", args[1].Type())
	}

	for _, elem := range arr.Elements {
		if s, ok := elem.(*String); ok {
			if s.Value == str.Value {
				return TRUE
			}
		}
	}

	return FALSE
}

// builtinStrGetLastComponent - get last component of path
// Usage: strGetLastComponent(path) -> string
//
//	strGetLastComponent(path, separator) -> string
func builtinStrGetLastComponent(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for strGetLastComponent. got=%d, want=1 or 2", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strGetLastComponent' must be STRING, got %s", args[0].Type())
	}

	sep := "/"
	if len(args) == 2 {
		s, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'strGetLastComponent' must be STRING, got %s", args[1].Type())
		}
		sep = s.Value
	}

	parts := strings.Split(path.Value, sep)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return NewString(parts[i])
		}
	}

	return NewString("")
}

// builtinStrFindDiffPos - find first different position
// Usage: strFindDiffPos(str1, str2) -> int
func builtinStrFindDiffPos(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strFindDiffPos. got=%d, want=2", len(args))
	}

	str1, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strFindDiffPos' must be STRING, got %s", args[0].Type())
	}

	str2, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strFindDiffPos' must be STRING, got %s", args[1].Type())
	}

	minLen := len(str1.Value)
	if len(str2.Value) < minLen {
		minLen = len(str2.Value)
	}

	for i := 0; i < minLen; i++ {
		if str1.Value[i] != str2.Value[i] {
			return NewInt(int64(i))
		}
	}

	if len(str1.Value) != len(str2.Value) {
		return NewInt(int64(minLen))
	}

	return NewInt(-1)
}

// builtinStrDiff - get string differences
// Usage: strDiff(str1, str2) -> array of [pos, char1, char2]
func builtinStrDiff(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for strDiff. got=%d, want=2 or 3", len(args))
	}

	str1, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strDiff' must be STRING, got %s", args[0].Type())
	}

	str2, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strDiff' must be STRING, got %s", args[1].Type())
	}

	limit := 100
	if len(args) == 3 {
		l, ok := args[2].(*Int)
		if !ok {
			return newError("third argument to 'strDiff' must be INT, got %s", args[2].Type())
		}
		limit = int(l.Value)
	}

	var diffs []Object
	maxLen := len(str1.Value)
	if len(str2.Value) > maxLen {
		maxLen = len(str2.Value)
	}

	for i := 0; i < maxLen && len(diffs) < limit; i++ {
		var c1, c2 string
		if i < len(str1.Value) {
			c1 = string(str1.Value[i])
		}
		if i < len(str2.Value) {
			c2 = string(str2.Value[i])
		}

		if c1 != c2 {
			diffs = append(diffs, NewArray([]Object{
				NewInt(int64(i)),
				NewString(c1),
				NewString(c2),
			}))
		}
	}

	return NewArray(diffs)
}

// builtinStrFindAllSub - find all occurrences of substring
// Usage: strFindAllSub(str, substr) -> array of [start, end]
func builtinStrFindAllSub(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strFindAllSub. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strFindAllSub' must be STRING, got %s", args[0].Type())
	}

	substr, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strFindAllSub' must be STRING, got %s", args[1].Type())
	}

	if substr.Value == "" {
		return NewArray([]Object{})
	}

	var results []Object
	start := 0
	for {
		idx := strings.Index(str.Value[start:], substr.Value)
		if idx == -1 {
			break
		}
		actualIdx := start + idx
		results = append(results, NewArray([]Object{
			NewInt(int64(actualIdx)),
			NewInt(int64(actualIdx + len(substr.Value))),
		}))
		start = actualIdx + len(substr.Value)
	}

	return NewArray(results)
}

// builtinLimitStr - limit string length (by Unicode characters)
// Usage: limitStr(str, maxLen, suffix?) -> string
func builtinLimitStr(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for limitStr. got=%d, want=2 or 3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'limitStr' must be STRING, got %s", args[0].Type())
	}

	maxLen, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'limitStr' must be INT, got %s", args[1].Type())
	}

	suffix := "..."
	if len(args) == 3 {
		s, ok := args[2].(*String)
		if !ok {
			return newError("third argument to 'limitStr' must be STRING, got %s", args[2].Type())
		}
		suffix = s.Value
	}

	runes := []rune(str.Value)
	runeLen := len(runes)

	if runeLen <= int(maxLen.Value) {
		return str
	}

	suffixRunes := []rune(suffix)
	suffixLen := len(suffixRunes)

	if int(maxLen.Value) <= suffixLen {
		return NewString(string(runes[:maxLen.Value]))
	}

	result := string(runes[:int(maxLen.Value)-suffixLen]) + suffix
	return NewString(result)
}

// builtinStrQuote - quote string
// Usage: strQuote(str) -> string
func builtinStrQuote(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for strQuote. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'strQuote' must be STRING, got %s", args[0].Type())
	}

	return NewString("\"" + strings.ReplaceAll(strings.ReplaceAll(str.Value, "\\", "\\\\"), "\"", "\\\"") + "\"")
}

// builtinStrUnquote - unquote string
// Usage: strUnquote(str) -> string
func builtinStrUnquote(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for strUnquote. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'strUnquote' must be STRING, got %s", args[0].Type())
	}

	s := str.Value
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, "\\\"", "\"")
		s = strings.ReplaceAll(s, "\\\\", "\\")
	}

	return NewString(s)
}

// builtinStrToInt - convert string to int with error
// Usage: strToInt(str) -> int or error
//
//	strToInt(str, default) -> int
func builtinStrToInt(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for strToInt. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strToInt' must be STRING, got %s", args[0].Type())
	}

	var result int64
	valid := true

	for i, c := range str.Value {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else if i == 0 && c == '-' {
			continue
		} else if i == 0 && c == '+' {
			continue
		} else {
			valid = false
			break
		}
	}

	if len(str.Value) > 0 && str.Value[0] == '-' {
		result = -result
	}

	if !valid {
		if len(args) == 2 {
			if d, ok := args[1].(*Int); ok {
				return d
			}
		}
		return newError("invalid integer string: %s", str.Value)
	}

	return NewInt(result)
}

// builtinGetTextSimilarity - calculate text similarity (cosine similarity)
// Usage: getTextSimilarity(str1, str2) -> float
func builtinGetTextSimilarity(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for getTextSimilarity. got=%d, want=2", len(args))
	}

	str1, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getTextSimilarity' must be STRING, got %s", args[0].Type())
	}

	str2, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'getTextSimilarity' must be STRING, got %s", args[1].Type())
	}

	// Simple character frequency similarity
	freq1 := make(map[rune]int)
	freq2 := make(map[rune]int)

	for _, r := range str1.Value {
		freq1[r]++
	}
	for _, r := range str2.Value {
		freq2[r]++
	}

	// Calculate cosine similarity
	var dot, mag1, mag2 float64
	for r, c := range freq1 {
		dot += float64(c) * float64(freq2[r])
		mag1 += float64(c * c)
	}
	for _, c := range freq2 {
		mag2 += float64(c * c)
	}

	if mag1 == 0 || mag2 == 0 {
		return NewFloat(0)
	}

	return NewFloat(dot / (sqrtFloat(mag1) * sqrtFloat(mag2)))
}

// builtinFuzzyFind - fuzzy find strings
// Usage: fuzzyFind(arr, query) -> array of matches
//
//	fuzzyFind(arr, query, "-sort") -> array of matches sorted by score
func builtinFuzzyFind(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for fuzzyFind. got=%d, want=2 or 3", len(args))
	}

	arr, ok := args[0].(*Array)
	if !ok {
		return newError("first argument to 'fuzzyFind' must be ARRAY, got %s", args[0].Type())
	}

	query, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'fuzzyFind' must be STRING, got %s", args[1].Type())
	}

	doSort := false
	if len(args) == 3 {
		if opt, ok := args[2].(*String); ok {
			doSort = opt.Value == "-sort"
		}
	}

	queryLower := strings.ToLower(query.Value)
	var matches []Object

	for _, elem := range arr.Elements {
		if s, ok := elem.(*String); ok {
			score := fuzzyScore(queryLower, strings.ToLower(s.Value))
			if score > 0 {
				matches = append(matches, NewArray([]Object{
					s,
					NewInt(int64(score)),
				}))
			}
		}
	}

	if doSort && len(matches) > 1 {
		// Sort by score descending
		for i := 0; i < len(matches)-1; i++ {
			for j := i + 1; j < len(matches); j++ {
				arrI := matches[i].(*Array)
				arrJ := matches[j].(*Array)
				if len(arrI.Elements) > 1 && len(arrJ.Elements) > 1 {
					scoreI := arrI.Elements[1].(*Int).Value
					scoreJ := arrJ.Elements[1].(*Int).Value
					if scoreJ > scoreI {
						matches[i], matches[j] = matches[j], matches[i]
					}
				}
			}
		}
	}

	return NewArray(matches)
}

// builtinReverseStr - reverse string (used by strReverse builtin)
// Usage: reverseStr(str) -> string
func builtinReverseStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for reverseStr. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'reverseStr' must be STRING, got %s", args[0].Type())
	}

	runes := []rune(str.Value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return NewString(string(runes))
}

// Helper functions

func fuzzyScore(query, target string) int {
	if len(query) == 0 {
		return 0
	}

	queryIdx := 0
	score := 0

	for _, c := range target {
		if queryIdx < len(query) && rune(query[queryIdx]) == c {
			score++
			queryIdx++
		}
	}

	if queryIdx == len(query) {
		return score
	}

	return 0
}

func sqrtFloat(x float64) float64 {
	if x <= 0 {
		return 0
	}

	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}

	return z
}
