// pkg/objects/builtin_misc.go
// Miscellaneous built-in functions for Xxlang (random, temp, url, security)
package objects

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	Builtins["getRandomInt"] = &Builtin{Fn: builtinGetRandomInt}
	Builtins["getRandomFloat"] = &Builtin{Fn: builtinGetRandomFloat}
	Builtins["getRandomStr"] = &Builtin{Fn: builtinGetRandomStr}
	Builtins["createTempDir"] = &Builtin{Fn: builtinCreateTempDir}
	Builtins["createTempFile"] = &Builtin{Fn: builtinCreateTempFile}
	Builtins["changeDir"] = &Builtin{Fn: builtinChangeDir}
	Builtins["lookPath"] = &Builtin{Fn: builtinLookPath}
	Builtins["joinUrlPath"] = &Builtin{Fn: builtinJoinUrlPath}
	Builtins["parseUrl"] = &Builtin{Fn: builtinParseUrl}
	Builtins["parseQuery"] = &Builtin{Fn: builtinParseQuery}
	Builtins["isHttps"] = &Builtin{Fn: builtinIsHttps}
	Builtins["genToken"] = &Builtin{Fn: builtinGenToken}
	Builtins["genOtpCode"] = &Builtin{Fn: builtinGenOtpCode}
	Builtins["checkOtpCode"] = &Builtin{Fn: builtinCheckOtpCode}
	rand.Seed(time.Now().UnixNano())
}

// builtinGetRandomInt - get random integer in range [min, max]
// Usage: getRandomInt(max) -> int
//
//	getRandomInt(min, max) -> int
func builtinGetRandomInt(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for getRandomInt. got=%d, want=1 or 2", len(args))
	}

	var min, max int64

	if len(args) == 1 {
		m, ok := args[0].(*Int)
		if !ok {
			return newError("argument must be INT, got %s", args[0].Type())
		}
		max = m.Value
		min = 0
	} else {
		minVal, ok := args[0].(*Int)
		if !ok {
			return newError("min must be INT, got %s", args[0].Type())
		}
		maxVal, ok := args[1].(*Int)
		if !ok {
			return newError("max must be INT, got %s", args[1].Type())
		}
		min = minVal.Value
		max = maxVal.Value
	}

	if min > max {
		min, max = max, min
	}

	r := rand.Int63n(max - min + 1)
	return NewInt(min + r)
}

// builtinGetRandomFloat - get random float in range [0, 1)
// Usage: getRandomFloat() -> float
func builtinGetRandomFloat(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getRandomFloat. got=%d, want=0", len(args))
	}

	return NewFloat(rand.Float64())
}

// builtinGetRandomStr - generate random string
// Usage: getRandomStr(length) -> string
//
//	getRandomStr(length, charset) -> string
func builtinGetRandomStr(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for getRandomStr. got=%d, want=1 or 2", len(args))
	}

	length, ok := args[0].(*Int)
	if !ok {
		return newError("length must be INT, got %s", args[0].Type())
	}

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(args) == 2 {
		cs, ok := args[1].(*String)
		if !ok {
			return newError("charset must be STRING, got %s", args[1].Type())
		}
		if cs.Value != "" {
			charset = cs.Value
		}
	}

	result := make([]byte, length.Value)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return NewString(string(result))
}

// builtinCreateTempDir - create temporary directory
// Usage: createTempDir() -> string
//
//	createTempDir(dir) -> string
//
//	createTempDir(dir, pattern) -> string
func builtinCreateTempDir(args ...Object) Object {
	if len(args) > 2 {
		return newError("wrong number of arguments for createTempDir. got=%d, want=0-2", len(args))
	}

	dir := ""
	pattern := "xxlang_*"

	if len(args) >= 1 {
		d, ok := args[0].(*String)
		if !ok {
			return newError("dir must be STRING, got %s", args[0].Type())
		}
		dir = d.Value
	}

	if len(args) >= 2 {
		p, ok := args[1].(*String)
		if !ok {
			return newError("pattern must be STRING, got %s", args[1].Type())
		}
		pattern = p.Value
	}

	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	return NewString(tempDir)
}

// builtinCreateTempFile - create temporary file
// Usage: createTempFile() -> string
//
//	createTempFile(dir) -> string
//
//	createTempFile(dir, pattern) -> string
func builtinCreateTempFile(args ...Object) Object {
	if len(args) > 2 {
		return newError("wrong number of arguments for createTempFile. got=%d, want=0-2", len(args))
	}

	dir := ""
	pattern := "xxlang_*"

	if len(args) >= 1 {
		d, ok := args[0].(*String)
		if !ok {
			return newError("dir must be STRING, got %s", args[0].Type())
		}
		dir = d.Value
	}

	if len(args) >= 2 {
		p, ok := args[1].(*String)
		if !ok {
			return newError("pattern must be STRING, got %s", args[1].Type())
		}
		pattern = p.Value
	}

	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return &Error{Message: err.Error()}
	}
	tempFile.Close()

	return NewString(tempFile.Name())
}

// builtinChangeDir - change current working directory
// Usage: changeDir(path) -> null
func builtinChangeDir(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for changeDir. got=%d, want=1", len(args))
	}

	path, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'changeDir' must be STRING, got %s", args[0].Type())
	}

	err := os.Chdir(path.Value)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	return NULL
}

// builtinLookPath - find executable file in PATH
// Usage: lookPath(name) -> string or null
func builtinLookPath(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for lookPath. got=%d, want=1", len(args))
	}

	name, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'lookPath' must be STRING, got %s", args[0].Type())
	}

	path, err := exec.LookPath(name.Value)
	if err != nil {
		return NULL
	}

	return NewString(path)
}

// builtinJoinUrlPath - join URL path components
// Usage: joinUrlPath(base, elem...) -> string
func builtinJoinUrlPath(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for joinUrlPath. got=%d, want>=1", len(args))
	}

	base, ok := args[0].(*String)
	if !ok {
		return newError("first argument must be STRING, got %s", args[0].Type())
	}

	u, err := url.Parse(base.Value)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	if len(args) > 1 {
		elements := make([]string, len(args)-1)
		for i, arg := range args[1:] {
			s, ok := arg.(*String)
			if !ok {
				return newError("argument %d must be STRING, got %s", i+2, arg.Type())
			}
			elements[i] = s.Value
		}
		u.Path = joinPathElements(elements)
	}

	return NewString(u.String())
}

// joinPathElements joins path elements
func joinPathElements(elements []string) string {
	if len(elements) == 0 {
		return ""
	}
	result := elements[0]
	for _, e := range elements[1:] {
		if !strings.HasSuffix(result, "/") && !strings.HasPrefix(e, "/") {
			result += "/"
		}
		result += e
	}
	return result
}

// builtinParseUrl - parse URL and return map
// Usage: parseUrl(urlStr) -> map
func builtinParseUrl(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for parseUrl. got=%d, want=1", len(args))
	}

	urlStr, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'parseUrl' must be STRING, got %s", args[0].Type())
	}

	u, err := url.Parse(urlStr.Value)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	result := NewMapWithCapacity(10)

	addToMap(result, "scheme", u.Scheme)
	addToMap(result, "host", u.Host)
	addToMap(result, "path", u.Path)
	addToMap(result, "rawQuery", u.RawQuery)
	addToMap(result, "fragment", u.Fragment)
	addToMap(result, "rawPath", u.RawPath)
	addToMap(result, "rawFragment", u.RawFragment)
	addToMap(result, "opaque", u.Opaque)
	addToMap(result, "user", u.User.Username())
	if password, ok := u.User.Password(); ok {
		addToMap(result, "password", password)
	}

	return result
}

// builtinParseQuery - parse URL query string
// Usage: parseQuery(queryStr) -> map
func builtinParseQuery(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for parseQuery. got=%d, want=1", len(args))
	}

	queryStr, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'parseQuery' must be STRING, got %s", args[0].Type())
	}

	values, err := url.ParseQuery(queryStr.Value)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	result := NewMapWithCapacity(len(values))

	for key, vals := range values {
		keyObj := NewString(key)
		hashKey := keyObj.HashKey()

		if len(vals) == 1 {
			result.Pairs[hashKey] = MapPair{Key: keyObj, Value: NewString(vals[0])}
		} else {
			elements := make([]Object, len(vals))
			for i, v := range vals {
				elements[i] = NewString(v)
			}
			result.Pairs[hashKey] = MapPair{Key: keyObj, Value: NewArray(elements)}
		}
	}

	return result
}

// builtinIsHttps - check if URL is HTTPS
// Usage: isHttps(urlStr) -> bool
func builtinIsHttps(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isHttps. got=%d, want=1", len(args))
	}

	urlStr, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'isHttps' must be STRING, got %s", args[0].Type())
	}

	u, err := url.Parse(urlStr.Value)
	if err != nil {
		return FALSE
	}

	return &Bool{Value: strings.EqualFold(u.Scheme, "https")}
}

// builtinGenToken - generate random token
// Usage: genToken() -> string
//
//	genToken(length) -> string
func builtinGenToken(args ...Object) Object {
	if len(args) > 1 {
		return newError("wrong number of arguments for genToken. got=%d, want=0 or 1", len(args))
	}

	length := 32
	if len(args) == 1 {
		l, ok := args[0].(*Int)
		if !ok {
			return newError("length must be INT, got %s", args[0].Type())
		}
		length = int(l.Value)
	}

	bytes := make([]byte, length)
	if _, err := cryptoRand.Read(bytes); err != nil {
		return &Error{Message: err.Error()}
	}

	return NewString(base64.URLEncoding.EncodeToString(bytes))
}

// builtinGenOtpCode - generate simple OTP code
// Usage: genOtpCode(secret) -> string
//
//	genOtpCode(secret, digits) -> string
func builtinGenOtpCode(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for genOtpCode. got=%d, want=1 or 2", len(args))
	}

	secret, ok := args[0].(*String)
	if !ok {
		return newError("secret must be STRING, got %s", args[0].Type())
	}

	digits := 6
	if len(args) == 2 {
		d, ok := args[1].(*Int)
		if !ok {
			return newError("digits must be INT, got %s", args[1].Type())
		}
		digits = int(d.Value)
	}

	secretStr := strings.ToUpper(strings.TrimSpace(secret.Value))
	if secretStr == "" {
		return newError("secret cannot be empty")
	}

	for _, c := range secretStr {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
			return newError("invalid character in secret: must be base32 (A-Z, 2-7)")
		}
	}

	counter := time.Now().Unix() / 30
	code := generateOTP(secretStr, counter, digits)
	return NewString(code)
}

// builtinCheckOtpCode - validate OTP code
// Usage: checkOtpCode(secret, code) -> bool
func builtinCheckOtpCode(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for checkOtpCode. got=%d, want=2 or 3", len(args))
	}

	secret, ok := args[0].(*String)
	if !ok {
		return newError("secret must be STRING, got %s", args[0].Type())
	}

	code, ok := args[1].(*String)
	if !ok {
		return newError("code must be STRING, got %s", args[1].Type())
	}

	digits := 6
	if len(args) == 3 {
		d, ok := args[2].(*Int)
		if !ok {
			return newError("digits must be INT, got %s", args[2].Type())
		}
		digits = int(d.Value)
	}

	// Check current and adjacent time windows
	counter := time.Now().Unix() / 30
	for i := int64(-1); i <= 1; i++ {
		expected := generateOTP(secret.Value, counter+i, digits)
		if expected == code.Value {
			return TRUE
		}
	}

	return FALSE
}

// generateOTP generates a simple OTP code
func generateOTP(secret string, counter int64, digits int) string {
	// Simple hash-based OTP generation
	data := fmt.Sprintf("%s%d", secret, counter)
	hash := 0
	for _, c := range data {
		hash = hash*31 + int(c)
	}

	if hash < 0 {
		hash = -hash
	}

	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, hash%int(pow10(digits)))
}

// pow10 returns 10^n
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// addToMap helper
func addToMap(result *Map, key, value string) {
	if value == "" {
		return
	}
	keyObj := NewString(key)
	hashKey := keyObj.HashKey()
	result.Pairs[hashKey] = MapPair{Key: keyObj, Value: NewString(value)}
}

// strToInt converts string to int
func strToInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// cleanPath cleans path
func cleanPath(p string) string {
	return filepath.Clean(p)
}

// isAbsPath checks if path is absolute (supports both Unix and Windows style)
func isAbsPath(p string) bool {
	if filepath.IsAbs(p) {
		return true
	}
	// On Windows, Unix-style absolute paths (/foo) are not recognized by filepath.IsAbs
	if len(p) > 0 && p[0] == '/' {
		return true
	}
	return false
}
