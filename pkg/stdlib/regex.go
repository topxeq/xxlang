// pkg/stdlib/regex.go
// Regular expression utilities for the Xxlang standard library.
// Uses regexp2 for full PCRE support including lookahead/lookbehind.
package stdlib

import (
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "regex",
		Exports: map[string]objects.Object{
			// compile compiles a regex pattern and returns a compiled regex object
			"compile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("compile() takes exactly 1 argument")
				}
				pattern, ok := args[0].(*objects.String)
				if !ok {
					return Error("compile() requires a string pattern")
				}

				// Use ECMAScript mode for broader compatibility
				re, err := regexp2.Compile(pattern.Value, regexp2.ECMAScript)
				if err != nil {
					return Error(fmt.Sprintf("compile() failed: %s", err.Error()))
				}

				return &CompiledRegex{Pattern: pattern.Value, Re: re}
			}),

			// match checks if a string matches a regex pattern
			"match": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("match() takes exactly 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("match() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("match() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("match() requires a string to match")
				}

				matched, err := re.MatchString(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("match() error: %s", err.Error()))
				}
				return Bool(matched)
			}),

			// find returns the first match of a regex pattern in a string
			"find": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("find() takes at least 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("find() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("find() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("find() requires a string to search")
				}

				match, err := re.FindStringMatch(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("find() error: %s", err.Error()))
				}
				if match == nil {
					return String("")
				}
				return String(match.String())
			}),

			// findAll returns all matches of a regex pattern in a string
			"findAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("findAll() takes at least 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("findAll() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("findAll() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("findAll() requires a string to search")
				}

				// Optional limit argument
				limit := -1
				if len(args) > 2 {
					if n, ok := args[2].(*objects.Int); ok {
						limit = int(n.Value)
					}
				}

				var matches []string
				match, err := re.FindStringMatch(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("findAll() error: %s", err.Error()))
				}

				for match != nil && (limit < 0 || len(matches) < limit) {
					matches = append(matches, match.String())
					match, err = re.FindNextMatch(match)
					if err != nil {
						return Error(fmt.Sprintf("findAll() error: %s", err.Error()))
					}
				}

				result := make([]objects.Object, len(matches))
				for i, m := range matches {
					result[i] = String(m)
				}
				return Array(result...)
			}),

			// findGroups returns captured groups from a regex match
			"findGroups": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("findGroups() takes at least 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("findGroups() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("findGroups() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("findGroups() requires a string to search")
				}

				match, err := re.FindStringMatch(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("findGroups() error: %s", err.Error()))
				}
				if match == nil {
					return Null()
				}

				groups := match.Groups()
				result := make([]objects.Object, len(groups))
				for i, g := range groups {
					result[i] = String(g.String())
				}
				return Array(result...)
			}),

			// replace replaces matches of a regex pattern with a replacement string
			"replace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("replace() takes exactly 3 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("replace() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("replace() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("replace() requires a string to search")
				}

				repl, ok := args[2].(*objects.String)
				if !ok {
					return Error("replace() requires a replacement string")
				}

				result, err := re.Replace(s.Value, repl.Value, 0, -1)
				if err != nil {
					return Error(fmt.Sprintf("replace() error: %s", err.Error()))
				}
				return String(result)
			}),

			// split splits a string by a regex pattern
			"split": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("split() takes at least 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("split() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("split() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("split() requires a string to split")
				}

				// Optional limit argument
				limit := -1
				if len(args) > 2 {
					if n, ok := args[2].(*objects.Int); ok {
						limit = int(n.Value)
					}
				}

				// regexp2 doesn't have Split, implement manually
				parts := regexSplit(re, s.Value, limit)

				result := make([]objects.Object, len(parts))
				for i, p := range parts {
					result[i] = String(p)
				}
				return Array(result...)
			}),

			// quote escapes special regex characters in a string
			"quote": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("quote() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("quote() requires a string argument")
				}
				return String(regexp2.Escape(s.Value))
			}),

			// count returns the number of matches
			"count": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("count() takes exactly 2 arguments")
				}

				var re *regexp2.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp2.Compile(pattern.Value, regexp2.ECMAScript)
					if err != nil {
						return Error(fmt.Sprintf("count() invalid pattern: %s", err.Error()))
					}
				case *CompiledRegex:
					re = pattern.Re
				default:
					return Error("count() requires a string pattern or compiled regex")
				}

				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("count() requires a string to search")
				}

				count := 0
				match, err := re.FindStringMatch(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("count() error: %s", err.Error()))
				}

				for match != nil {
					count++
					match, err = re.FindNextMatch(match)
					if err != nil {
						return Error(fmt.Sprintf("count() error: %s", err.Error()))
					}
				}

				return Int(int64(count))
			}),

			// test checks if a pattern is valid
			"test": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("test() takes exactly 1 argument")
				}
				pattern, ok := args[0].(*objects.String)
				if !ok {
					return Error("test() requires a string pattern")
				}

				_, err := regexp2.Compile(pattern.Value, regexp2.ECMAScript)
				return Bool(err == nil)
			}),
		},
	})
}

// CompiledRegex represents a compiled regular expression
type CompiledRegex struct {
	Pattern string
	Re      *regexp2.Regexp
}

func (cr *CompiledRegex) Type() objects.ObjectType { return objects.ObjectType("COMPILED_REGEX") }
func (cr *CompiledRegex) TypeTag() objects.TypeTag { return objects.TagUnknown } // Custom type, use Unknown
func (cr *CompiledRegex) Inspect() string          { return fmt.Sprintf("[compiled regex: %s]", cr.Pattern) }
func (cr *CompiledRegex) ToBool() *objects.Bool    { return objects.TRUE }
func (cr *CompiledRegex) HashKey() objects.HashKey {
	return objects.HashKey{Type: objects.ObjectType("COMPILED_REGEX"), Value: 0}
}

// regexSplit splits a string by a regex pattern (regexp2 doesn't have Split method)
func regexSplit(re *regexp2.Regexp, s string, limit int) []string {
	if limit == 0 {
		return []string{s}
	}

	var parts []string
	lastIndex := 0

	match, err := re.FindStringMatch(s)
	if err != nil {
		return []string{s}
	}

	count := 0
	for match != nil {
		if limit > 0 && count >= limit-1 {
			break
		}

		// Get the matched position
		start, _ := match.Groups()[0].Captures[0].Index, match.Groups()[0].Captures[0].Length
		end := start + match.Groups()[0].Captures[0].Length

		// Add the part before the match
		if lastIndex < int(start) {
			parts = append(parts, s[lastIndex:start])
		} else if lastIndex == int(start) && len(parts) > 0 {
			// Empty match, skip
		} else {
			parts = append(parts, "")
		}

		lastIndex = int(end)
		count++

		match, err = re.FindNextMatch(match)
		if err != nil {
			break
		}
	}

	// Add the remaining part
	if lastIndex <= len(s) {
		parts = append(parts, s[lastIndex:])
	}

	if limit > 0 && len(parts) > limit {
		parts = parts[:limit]
	}

	return parts
}
