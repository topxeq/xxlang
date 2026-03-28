// pkg/objects/builtin_regex.go
// Regular expression related built-in functions for Xxlang
package objects

import (
	"regexp"
)

func init() {
	// Regular expression functions
	Builtins["regMatch"] = &Builtin{Fn: builtinRegMatch}
	Builtins["regContains"] = &Builtin{Fn: builtinRegContains}
	Builtins["regFindFirst"] = &Builtin{Fn: builtinRegFindFirst}
	Builtins["regFindAll"] = &Builtin{Fn: builtinRegFindAll}
	Builtins["regFindFirstGroups"] = &Builtin{Fn: builtinRegFindFirstGroups}
	Builtins["regFindAllGroups"] = &Builtin{Fn: builtinRegFindAllGroups}
	Builtins["regReplace"] = &Builtin{Fn: builtinRegReplace}
	Builtins["regSplit"] = &Builtin{Fn: builtinRegSplit}
	Builtins["regCount"] = &Builtin{Fn: builtinRegCount}
	Builtins["regQuote"] = &Builtin{Fn: builtinRegQuote}
	Builtins["regFindAllIndex"] = &Builtin{Fn: builtinRegFindAllIndex}
}

// regMatch - check if string fully matches regex pattern
// Usage: regMatch(str, pattern) -> bool
func builtinRegMatch(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regMatch. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regMatch' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regMatch' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regMatch compile error: %v", err)
	}

	return &Bool{Value: re.MatchString(str.Value)}
}

// regContains - check if string contains regex match
// Usage: regContains(str, pattern) -> bool
func builtinRegContains(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regContains. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regContains' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regContains' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regContains compile error: %v", err)
	}

	return &Bool{Value: re.MatchString(str.Value)}
}

// regFindFirst - find first match of regex pattern
// Usage: regFindFirst(str, pattern) -> string or null
//
//	regFindFirst(str, pattern, groupIndex) -> string
func builtinRegFindFirst(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for regFindFirst. got=%d, want=2 or 3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regFindFirst' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regFindFirst' must be STRING, got %s", args[1].Type())
	}

	groupIndex := 0
	if len(args) == 3 {
		g, ok := args[2].(*Int)
		if !ok {
			return newError("third argument to 'regFindFirst' must be INT, got %s", args[2].Type())
		}
		groupIndex = int(g.Value)
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regFindFirst compile error: %v", err)
	}

	matches := re.FindStringSubmatch(str.Value)
	if matches == nil {
		return NULL
	}

	if groupIndex < 0 || groupIndex >= len(matches) {
		return newError("regFindFirst: group index %d out of range (0-%d)", groupIndex, len(matches)-1)
	}

	return NewString(matches[groupIndex])
}

// regFindAll - find all matches of regex pattern
// Usage: regFindAll(str, pattern) -> array
//
//	regFindAll(str, pattern, groupIndex) -> array
func builtinRegFindAll(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for regFindAll. got=%d, want=2 or 3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regFindAll' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regFindAll' must be STRING, got %s", args[1].Type())
	}

	groupIndex := 0
	if len(args) == 3 {
		g, ok := args[2].(*Int)
		if !ok {
			return newError("third argument to 'regFindAll' must be INT, got %s", args[2].Type())
		}
		groupIndex = int(g.Value)
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regFindAll compile error: %v", err)
	}

	matches := re.FindAllStringSubmatch(str.Value, -1)
	if matches == nil {
		return NewArray([]Object{})
	}

	results := make([]Object, 0, len(matches))
	for _, match := range matches {
		if groupIndex < len(match) {
			results = append(results, NewString(match[groupIndex]))
		}
	}

	return NewArray(results)
}

// regFindFirstGroups - find first match and return all groups
// Usage: regFindFirstGroups(str, pattern) -> array
func builtinRegFindFirstGroups(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regFindFirstGroups. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regFindFirstGroups' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regFindFirstGroups' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regFindFirstGroups compile error: %v", err)
	}

	matches := re.FindStringSubmatch(str.Value)
	if matches == nil {
		return NULL
	}

	results := make([]Object, len(matches))
	for i, match := range matches {
		results[i] = NewString(match)
	}

	return NewArray(results)
}

// regFindAllGroups - find all matches and return all groups
// Usage: regFindAllGroups(str, pattern) -> array of arrays
func builtinRegFindAllGroups(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regFindAllGroups. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regFindAllGroups' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regFindAllGroups' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regFindAllGroups compile error: %v", err)
	}

	matches := re.FindAllStringSubmatch(str.Value, -1)
	if matches == nil {
		return NewArray([]Object{})
	}

	results := make([]Object, len(matches))
	for i, match := range matches {
		groups := make([]Object, len(match))
		for j, g := range match {
			groups[j] = NewString(g)
		}
		results[i] = NewArray(groups)
	}

	return NewArray(results)
}

// regReplace - replace regex matches
// Usage: regReplace(str, pattern, replacement) -> string
func builtinRegReplace(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments for regReplace. got=%d, want=3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regReplace' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regReplace' must be STRING, got %s", args[1].Type())
	}

	replacement, ok := args[2].(*String)
	if !ok {
		return newError("third argument to 'regReplace' must be STRING, got %s", args[2].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regReplace compile error: %v", err)
	}

	result := re.ReplaceAllString(str.Value, replacement.Value)
	return NewString(result)
}

// regSplit - split string by regex pattern
// Usage: regSplit(str, pattern) -> array
//
//	regSplit(str, pattern, limit) -> array
func builtinRegSplit(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for regSplit. got=%d, want=2 or 3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regSplit' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regSplit' must be STRING, got %s", args[1].Type())
	}

	limit := -1
	if len(args) == 3 {
		l, ok := args[2].(*Int)
		if !ok {
			return newError("third argument to 'regSplit' must be INT, got %s", args[2].Type())
		}
		limit = int(l.Value)
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regSplit compile error: %v", err)
	}

	parts := re.Split(str.Value, limit)
	results := make([]Object, len(parts))
	for i, part := range parts {
		results[i] = NewString(part)
	}

	return NewArray(results)
}

// regCount - count regex matches
// Usage: regCount(str, pattern) -> int
func builtinRegCount(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regCount. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regCount' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regCount' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regCount compile error: %v", err)
	}

	matches := re.FindAllString(str.Value, -1)
	if matches == nil {
		return NewInt(0)
	}

	return NewInt(int64(len(matches)))
}

// regQuote - quote regex special characters in string
// Usage: regQuote(str) -> string
func builtinRegQuote(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for regQuote. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'regQuote' must be STRING, got %s", args[0].Type())
	}

	return NewString(regexp.QuoteMeta(str.Value))
}

// regFindAllIndex - find all match positions
// Usage: regFindAllIndex(str, pattern) -> array of [start, end] pairs
func builtinRegFindAllIndex(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for regFindAllIndex. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'regFindAllIndex' must be STRING, got %s", args[0].Type())
	}

	pattern, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'regFindAllIndex' must be STRING, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regFindAllIndex compile error: %v", err)
	}

	matches := re.FindAllStringIndex(str.Value, -1)
	if matches == nil {
		return NewArray([]Object{})
	}

	results := make([]Object, len(matches))
	for i, match := range matches {
		pair := make([]Object, 2)
		pair[0] = NewInt(int64(match[0]))
		pair[1] = NewInt(int64(match[1]))
		results[i] = NewArray(pair)
	}

	return NewArray(results)
}
