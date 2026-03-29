// pkg/objects/builtin_encode.go
// Encoding and encryption related built-in functions for Xxlang
package objects

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"html"
	"net/url"
	"strings"
)

func init() {
	// URL encoding functions
	Builtins["urlEncode"] = &Builtin{Fn: builtinUrlEncode}
	Builtins["urlDecode"] = &Builtin{Fn: builtinUrlDecode}
	Builtins["urlEncodeComponent"] = &Builtin{Fn: builtinUrlEncodeComponent}
	Builtins["urlDecodeComponent"] = &Builtin{Fn: builtinUrlDecodeComponent}

	// HTML encoding functions
	Builtins["htmlEncode"] = &Builtin{Fn: builtinHtmlEncode}
	Builtins["htmlDecode"] = &Builtin{Fn: builtinHtmlDecode}

	// Additional hash functions
	Builtins["sha1"] = &Builtin{Fn: builtinSha1}
	Builtins["sha512"] = &Builtin{Fn: builtinSha512}
	Builtins["hashStr"] = &Builtin{Fn: builtinHashStr}

	// Additional encoding functions
	Builtins["toHex"] = &Builtin{Fn: builtinHexEncode}
	Builtins["unhex"] = &Builtin{Fn: builtinHexDecode}
	Builtins["hexToStr"] = &Builtin{Fn: builtinHexToStr}

	// Simple obfuscation encoding
	Builtins["simpleEncode"] = &Builtin{Fn: builtinSimpleEncode}
	Builtins["simpleDecode"] = &Builtin{Fn: builtinSimpleDecode}
}

// urlEncode - encode string for URL (path encoding)
// Usage: urlEncode(str) -> string
func builtinUrlEncode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for urlEncode. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'urlEncode' must be STRING, got %s", args[0].Type())
	}

	return NewString(url.PathEscape(str.Value))
}

// urlDecode - decode URL encoded string (path decoding)
// Usage: urlDecode(str) -> string
func builtinUrlDecode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for urlDecode. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'urlDecode' must be STRING, got %s", args[0].Type())
	}

	decoded, err := url.PathUnescape(str.Value)
	if err != nil {
		return newError("urlDecode error: %v", err)
	}

	return NewString(decoded)
}

// urlEncodeComponent - encode string for URL component (query encoding)
// Usage: urlEncodeComponent(str) -> string
func builtinUrlEncodeComponent(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for urlEncodeComponent. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'urlEncodeComponent' must be STRING, got %s", args[0].Type())
	}

	return NewString(url.QueryEscape(str.Value))
}

// urlDecodeComponent - decode URL component encoded string
// Usage: urlDecodeComponent(str) -> string
func builtinUrlDecodeComponent(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for urlDecodeComponent. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'urlDecodeComponent' must be STRING, got %s", args[0].Type())
	}

	decoded, err := url.QueryUnescape(str.Value)
	if err != nil {
		return newError("urlDecodeComponent error: %v", err)
	}

	return NewString(decoded)
}

// htmlEncode - encode HTML special characters
// Usage: htmlEncode(str) -> string
func builtinHtmlEncode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for htmlEncode. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'htmlEncode' must be STRING, got %s", args[0].Type())
	}

	return NewString(html.EscapeString(str.Value))
}

// htmlDecode - decode HTML entities
// Usage: htmlDecode(str) -> string
func builtinHtmlDecode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for htmlDecode. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'htmlDecode' must be STRING, got %s", args[0].Type())
	}

	return NewString(html.UnescapeString(str.Value))
}

// sha1 - calculate SHA1 hash
// Usage: sha1(str) -> string (hex encoded)
func builtinSha1(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for sha1. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'sha1' must be STRING, got %s", args[0].Type())
	}

	hash := sha1.Sum([]byte(str.Value))
	return NewString(hex.EncodeToString(hash[:]))
}

// sha512 - calculate SHA512 hash
// Usage: sha512(str) -> string (hex encoded)
func builtinSha512(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for sha512. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'sha512' must be STRING, got %s", args[0].Type())
	}

	hash := sha512.Sum512([]byte(str.Value))
	return NewString(hex.EncodeToString(hash[:]))
}

// hashStr - simple string hash (djb2 algorithm)
// Usage: hashStr(str) -> int
func builtinHashStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for hashStr. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'hashStr' must be STRING, got %s", args[0].Type())
	}

	// DJB2 hash algorithm
	hash := uint64(5381)
	for _, c := range str.Value {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return NewInt(int64(hash))
}

// builtinHexEncode - alias for hexEncode (already defined in main builtin.go)
func builtinHexEncode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toHex. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'toHex' must be STRING, got %s", args[0].Type())
	}

	return NewString(hex.EncodeToString([]byte(str.Value)))
}

// builtinHexDecode - alias for hexDecode
func builtinHexDecode(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for unhex. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'unhex' must be STRING, got %s", args[0].Type())
	}

	decoded, err := hex.DecodeString(str.Value)
	if err != nil {
		return newError("unhex failed: %s", err.Error())
	}

	return NewString(string(decoded))
}

// builtinHexToStr - convert hex string to string
func builtinHexToStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for hexToStr. got=%d, want=1", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'hexToStr' must be STRING, got %s", args[0].Type())
	}

	decoded, err := hex.DecodeString(str.Value)
	if err != nil {
		return newError("hexToStr failed: %s", err.Error())
	}

	return NewString(string(decoded))
}

// hexChars for simple encoding
const hexChars = "0123456789ABCDEF"

// builtinSimpleEncode - simple encoding compatible with Charlang
// Keeps '0'-'9' and 'a'-'z' unchanged, encodes other chars as padding + hex
// Usage: simpleEncode(str) -> string (default padding '_')
//
//	simpleEncode(str, padding) -> string
func builtinSimpleEncode(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for simpleEncode. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'simpleEncode' must be STRING, got %s", args[0].Type())
	}

	padding := byte('_')
	if len(args) == 2 {
		p, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'simpleEncode' must be STRING, got %s", args[1].Type())
		}
		if len(p.Value) > 0 {
			padding = p.Value[0]
		}
	}

	var result strings.Builder
	for i := 0; i < len(str.Value); i++ {
		v := str.Value[i]
		// Keep '0'-'9' and 'a'-'z' unchanged
		if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'z') {
			result.WriteByte(v)
		} else {
			result.WriteByte(padding)
			result.WriteByte(hexChars[v>>4])
			result.WriteByte(hexChars[v&0x0F])
		}
	}

	return NewString(result.String())
}

// builtinSimpleDecode - simple decoding compatible with Charlang
// Usage: simpleDecode(str) -> string (default padding '_')
//
//	simpleDecode(str, padding) -> string
func builtinSimpleDecode(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for simpleDecode. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'simpleDecode' must be STRING, got %s", args[0].Type())
	}

	padding := byte('_')
	if len(args) == 2 {
		p, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'simpleDecode' must be STRING, got %s", args[1].Type())
		}
		if len(p.Value) > 0 {
			padding = p.Value[0]
		}
	}

	var result strings.Builder
	s := str.Value
	lenS := len(s)

	for i := 0; i < lenS; {
		if s[i] == padding {
			if i+2 >= lenS {
				// Invalid format, return as-is
				return str
			}
			// Decode hex value
			high := unhexChar(s[i+1])
			low := unhexChar(s[i+2])
			result.WriteByte(high<<4 | low)
			i += 3
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return NewString(result.String())
}

// unhexChar converts a hex character to its numeric value
func unhexChar(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
