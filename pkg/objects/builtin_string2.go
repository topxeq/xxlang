// pkg/objects/builtin_string2.go
// String processing enhancement built-in functions for Xxlang
package objects

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func init() {
	Builtins["strSplitLines"] = &Builtin{Fn: builtinStrSplitLines}
	Builtins["strContainsAny"] = &Builtin{Fn: builtinStrContainsAny}
	Builtins["strIndex"] = &Builtin{Fn: builtinStrIndex}
	Builtins["strLastIndex"] = &Builtin{Fn: builtinStrLastIndex}
	Builtins["strSplitN"] = &Builtin{Fn: builtinStrSplitN}
	Builtins["strPad"] = &Builtin{Fn: builtinStrPad}
	Builtins["strSub"] = &Builtin{Fn: builtinStrSub}
	Builtins["intToStr"] = &Builtin{Fn: builtinIntToStr}
	Builtins["floatToStr"] = &Builtin{Fn: builtinFloatToStr}
	Builtins["charCode"] = &Builtin{Fn: builtinCharCode}
	Builtins["charFromCode"] = &Builtin{Fn: builtinCharFromCode}
	Builtins["reverseMap"] = &Builtin{Fn: builtinReverseMap}
	Builtins["simpleStrToMap"] = &Builtin{Fn: builtinSimpleStrToMap}
	Builtins["mapToStr"] = &Builtin{Fn: builtinMapToStr}
	Builtins["bitNot"] = &Builtin{Fn: builtinBitNot}
	Builtins["bitAnd"] = &Builtin{Fn: builtinBitAnd}
	Builtins["bitOr"] = &Builtin{Fn: builtinBitOr}
	Builtins["bitXor"] = &Builtin{Fn: builtinBitXor}
	Builtins["bitShiftLeft"] = &Builtin{Fn: builtinBitShiftLeft}
	Builtins["bitShiftRight"] = &Builtin{Fn: builtinBitShiftRight}
}

// builtinStrSplitLines - split string by lines
// Usage: strSplitLines(str) -> array
func builtinStrSplitLines(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for strSplitLines. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'strSplitLines' must be STRING, got %s", args[0].Type())
	}

	lines := strings.Split(str.Value, "\n")
	elements := make([]Object, len(lines))
	for i, line := range lines {
		elements[i] = NewString(strings.TrimRight(line, "\r"))
	}

	return NewArray(elements)
}

// builtinStrContainsAny - check if string contains any of the characters
// Usage: strContainsAny(str, chars) -> bool
func builtinStrContainsAny(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strContainsAny. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strContainsAny' must be STRING, got %s", args[0].Type())
	}

	chars, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strContainsAny' must be STRING, got %s", args[1].Type())
	}

	return &Bool{Value: strings.ContainsAny(str.Value, chars.Value)}
}

// builtinStrIndex - find index of substring
// Usage: strIndex(str, substr) -> int
func builtinStrIndex(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strIndex. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strIndex' must be STRING, got %s", args[0].Type())
	}

	substr, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strIndex' must be STRING, got %s", args[1].Type())
	}

	return NewInt(int64(strings.Index(str.Value, substr.Value)))
}

// builtinStrLastIndex - find last index of substring
// Usage: strLastIndex(str, substr) -> int
func builtinStrLastIndex(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for strLastIndex. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strLastIndex' must be STRING, got %s", args[0].Type())
	}

	substr, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strLastIndex' must be STRING, got %s", args[1].Type())
	}

	return NewInt(int64(strings.LastIndex(str.Value, substr.Value)))
}

// builtinStrSplitN - split string with limit
// Usage: strSplitN(str, sep, n) -> array
func builtinStrSplitN(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments for strSplitN. got=%d, want=3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strSplitN' must be STRING, got %s", args[0].Type())
	}

	sep, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'strSplitN' must be STRING, got %s", args[1].Type())
	}

	n, ok := args[2].(*Int)
	if !ok {
		return newError("third argument to 'strSplitN' must be INT, got %s", args[2].Type())
	}

	parts := strings.SplitN(str.Value, sep.Value, int(n.Value))
	elements := make([]Object, len(parts))
	for i, part := range parts {
		elements[i] = NewString(part)
	}

	return NewArray(elements)
}

// builtinStrPad - pad string to specified length
// Usage: strPad(str, length) -> string
//
//	strPad(str, length, padStr) -> string
//
//	strPad(str, length, padStr, padRight) -> string
func builtinStrPad(args ...Object) Object {
	if len(args) < 2 || len(args) > 4 {
		return newError("wrong number of arguments for strPad. got=%d, want=2-4", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strPad' must be STRING, got %s", args[0].Type())
	}

	length, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'strPad' must be INT, got %s", args[1].Type())
	}

	padStr := " "
	if len(args) >= 3 {
		p, ok := args[2].(*String)
		if !ok {
			return newError("pad string must be STRING, got %s", args[2].Type())
		}
		if p.Value != "" {
			padStr = p.Value
		}
	}

	padRight := false
	if len(args) >= 4 {
		if b, ok := args[3].(*Bool); ok {
			padRight = b.Value
		}
	}

	targetLen := int(length.Value)
	strLen := utf8.RuneCountInString(str.Value)

	if strLen >= targetLen {
		return str
	}

	padLen := targetLen - strLen
	padRunes := []rune(padStr)

	var result []rune
	result = append(result, []rune(str.Value)...)

	for i := 0; i < padLen; i++ {
		if padRight {
			result = append(result, padRunes[i%len(padRunes)])
		} else {
			result = append([]rune{padRunes[(padLen-1-i)%len(padRunes)]}, result...)
		}
	}

	return NewString(string(result))
}

// builtinStrSub - get substring with start and end index
// Usage: strSub(str, start) -> string
//
//	strSub(str, start, end) -> string
func builtinStrSub(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for strSub. got=%d, want=2 or 3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strSub' must be STRING, got %s", args[0].Type())
	}

	start, ok := args[1].(*Int)
	if !ok {
		return newError("start index must be INT, got %s", args[1].Type())
	}

	runes := []rune(str.Value)
	startIdx := int(start.Value)

	if startIdx < 0 {
		startIdx = len(runes) + startIdx
	}

	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(runes) {
		return NewString("")
	}

	if len(args) == 2 {
		return NewString(string(runes[startIdx:]))
	}

	end, ok := args[2].(*Int)
	if !ok {
		return newError("end index must be INT, got %s", args[2].Type())
	}

	endIdx := int(end.Value)
	if endIdx < 0 {
		endIdx = len(runes) + endIdx
	}

	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > len(runes) {
		endIdx = len(runes)
	}

	if startIdx >= endIdx {
		return NewString("")
	}

	return NewString(string(runes[startIdx:endIdx]))
}

// builtinIntToStr - convert integer to string
// Usage: intToStr(n) -> string
//
//	intToStr(n, base) -> string
func builtinIntToStr(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for intToStr. got=%d, want=1 or 2", len(args))
	}

	n, ok := args[0].(*Int)
	if !ok {
		return newError("argument to 'intToStr' must be INT, got %s", args[0].Type())
	}

	base := 10
	if len(args) == 2 {
		b, ok := args[1].(*Int)
		if !ok {
			return newError("base must be INT, got %s", args[1].Type())
		}
		base = int(b.Value)
	}

	return NewString(strconv.FormatInt(n.Value, base))
}

// builtinFloatToStr - convert float to string
// Usage: floatToStr(f) -> string
//
//	floatToStr(f, prec) -> string
func builtinFloatToStr(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for floatToStr. got=%d, want=1 or 2", len(args))
	}

	f, ok := args[0].(*Float)
	if !ok {
		return newError("argument to 'floatToStr' must be FLOAT, got %s", args[0].Type())
	}

	prec := -1
	if len(args) == 2 {
		p, ok := args[1].(*Int)
		if !ok {
			return newError("precision must be INT, got %s", args[1].Type())
		}
		prec = int(p.Value)
	}

	if prec >= 0 {
		return NewString(strconv.FormatFloat(f.Value, 'f', prec, 64))
	}
	return NewString(strconv.FormatFloat(f.Value, 'g', -1, 64))
}

// builtinCharCode - get character code (Unicode code point)
// Usage: charCode(str) -> int
//
//	charCode(str, index) -> int
func builtinCharCode(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for charCode. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'charCode' must be STRING, got %s", args[0].Type())
	}

	if len(str.Value) == 0 {
		return NewInt(0)
	}

	index := 0
	if len(args) == 2 {
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT, got %s", args[1].Type())
		}
		index = int(idx.Value)
	}

	runes := []rune(str.Value)
	if index < 0 || index >= len(runes) {
		return newError("index out of bounds")
	}

	return NewInt(int64(runes[index]))
}

// builtinCharFromCode - create character from code point
// Usage: charFromCode(code) -> string
func builtinCharFromCode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for charFromCode. got=%d, want=1", len(args))
	}

	code, ok := args[0].(*Int)
	if !ok {
		return newError("argument to 'charFromCode' must be INT, got %s", args[0].Type())
	}

	return NewString(string(rune(code.Value)))
}

// builtinReverseMap - reverse map keys and values
// Usage: reverseMap(map) -> map
func builtinReverseMap(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for reverseMap. got=%d, want=1", len(args))
	}

	m, ok := args[0].(*Map)
	if !ok {
		return newError("argument to 'reverseMap' must be MAP, got %s", args[0].Type())
	}

	result := NewMapWithCapacity(len(m.Pairs))

	for _, pair := range m.Pairs {
		newKey := pair.Value
		newValue := pair.Key

		hashKey := newKey.HashKey()
		result.Pairs[hashKey] = MapPair{Key: newKey, Value: newValue}
	}

	return result
}

// builtinSimpleStrToMap - parse simple string to map
// Usage: simpleStrToMap(str) -> map
//
//	simpleStrToMap(str, sep1) -> map
//
//	simpleStrToMap(str, sep1, sep2) -> map
func builtinSimpleStrToMap(args ...Object) Object {
	if len(args) < 1 || len(args) > 3 {
		return newError("wrong number of arguments for simpleStrToMap. got=%d, want=1-3", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'simpleStrToMap' must be STRING, got %s", args[0].Type())
	}

	sep1 := ","
	sep2 := "="

	if len(args) >= 2 {
		s, ok := args[1].(*String)
		if !ok {
			return newError("separator must be STRING, got %s", args[1].Type())
		}
		sep1 = s.Value
	}

	if len(args) >= 3 {
		s, ok := args[2].(*String)
		if !ok {
			return newError("separator must be STRING, got %s", args[2].Type())
		}
		sep2 = s.Value
	}

	if str.Value == "" {
		return NewMapWithCapacity(0)
	}

	pairs := strings.Split(str.Value, sep1)
	result := NewMapWithCapacity(len(pairs))

	for _, pair := range pairs {
		parts := strings.SplitN(pair, sep2, 2)
		if len(parts) == 2 {
			key := NewString(strings.TrimSpace(parts[0]))
			value := NewString(strings.TrimSpace(parts[1]))
			hashKey := key.HashKey()
			result.Pairs[hashKey] = MapPair{Key: key, Value: value}
		}
	}

	return result
}

// builtinMapToStr - convert map to simple string
// Usage: mapToStr(map) -> string
//
//	mapToStr(map, sep1) -> string
//
//	mapToStr(map, sep1, sep2) -> string
func builtinMapToStr(args ...Object) Object {
	if len(args) < 1 || len(args) > 3 {
		return newError("wrong number of arguments for mapToStr. got=%d, want=1-3", len(args))
	}

	m, ok := args[0].(*Map)
	if !ok {
		return newError("argument to 'mapToStr' must be MAP, got %s", args[0].Type())
	}

	sep1 := ","
	sep2 := "="

	if len(args) >= 2 {
		s, ok := args[1].(*String)
		if !ok {
			return newError("separator must be STRING, got %s", args[1].Type())
		}
		sep1 = s.Value
	}

	if len(args) >= 3 {
		s, ok := args[2].(*String)
		if !ok {
			return newError("separator must be STRING, got %s", args[2].Type())
		}
		sep2 = s.Value
	}

	var parts []string
	for _, pair := range m.Pairs {
		parts = append(parts, fmt.Sprintf("%s%s%s", pair.Key.Inspect(), sep2, pair.Value.Inspect()))
	}

	return NewString(strings.Join(parts, sep1))
}

// builtinBitNot - bitwise NOT
// Usage: bitNot(n) -> int
func builtinBitNot(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for bitNot. got=%d, want=1", len(args))
	}

	n, ok := args[0].(*Int)
	if !ok {
		return newError("argument to 'bitNot' must be INT, got %s", args[0].Type())
	}

	return NewInt(^n.Value)
}

// builtinBitAnd - bitwise AND
// Usage: bitAnd(a, b) -> int
func builtinBitAnd(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bitAnd. got=%d, want=2", len(args))
	}

	a, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'bitAnd' must be INT, got %s", args[0].Type())
	}

	b, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'bitAnd' must be INT, got %s", args[1].Type())
	}

	return NewInt(a.Value & b.Value)
}

// builtinBitOr - bitwise OR
// Usage: bitOr(a, b) -> int
func builtinBitOr(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bitOr. got=%d, want=2", len(args))
	}

	a, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'bitOr' must be INT, got %s", args[0].Type())
	}

	b, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'bitOr' must be INT, got %s", args[1].Type())
	}

	return NewInt(a.Value | b.Value)
}

// builtinBitXor - bitwise XOR
// Usage: bitXor(a, b) -> int
func builtinBitXor(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bitXor. got=%d, want=2", len(args))
	}

	a, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'bitXor' must be INT, got %s", args[0].Type())
	}

	b, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'bitXor' must be INT, got %s", args[1].Type())
	}

	return NewInt(a.Value ^ b.Value)
}

// builtinBitShiftLeft - bitwise left shift
// Usage: bitShiftLeft(n, shift) -> int
func builtinBitShiftLeft(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bitShiftLeft. got=%d, want=2", len(args))
	}

	n, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'bitShiftLeft' must be INT, got %s", args[0].Type())
	}

	shift, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'bitShiftLeft' must be INT, got %s", args[1].Type())
	}

	return NewInt(n.Value << uint(shift.Value))
}

// builtinBitShiftRight - bitwise right shift
// Usage: bitShiftRight(n, shift) -> int
func builtinBitShiftRight(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for bitShiftRight. got=%d, want=2", len(args))
	}

	n, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'bitShiftRight' must be INT, got %s", args[0].Type())
	}

	shift, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'bitShiftRight' must be INT, got %s", args[1].Type())
	}

	return NewInt(n.Value >> uint(shift.Value))
}
