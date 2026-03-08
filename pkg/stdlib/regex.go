// pkg/stdlib/regex.go
// Regular expression utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"regexp"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/regex",
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

				re, err := regexp.Compile(pattern.Value)
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

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				return Bool(re.MatchString(s.Value))
			}),

			// find returns the first match of a regex pattern in a string
			"find": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("find() takes at least 2 arguments")
				}

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				// Check if we want all matches
				findAll := false
				if len(args) > 2 {
					if b, ok := args[2].(*objects.Bool); ok && b.Value {
						findAll = true
					}
				}

				if findAll {
					matches := re.FindAllString(s.Value, -1)
					result := make([]objects.Object, len(matches))
					for i, m := range matches {
						result[i] = String(m)
					}
					return Array(result...)
				}

				match := re.FindString(s.Value)
				return String(match)
			}),

			// findAll returns all matches of a regex pattern in a string
			"findAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("findAll() takes at least 2 arguments")
				}

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				matches := re.FindAllString(s.Value, limit)
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

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				matches := re.FindStringSubmatch(s.Value)
				if matches == nil {
					return Null()
				}

				// Return array of matched groups (including full match at index 0)
				result := make([]objects.Object, len(matches))
				for i, m := range matches {
					result[i] = String(m)
				}
				return Array(result...)
			}),

			// replace replaces matches of a regex pattern with a replacement string
			"replace": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("replace() takes exactly 3 arguments")
				}

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				result := re.ReplaceAllString(s.Value, repl.Value)
				return String(result)
			}),


			// split splits a string by a regex pattern
			"split": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("split() takes at least 2 arguments")
				}

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				parts := re.Split(s.Value, limit)
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
				return String(regexp.QuoteMeta(s.Value))
			}),

			// count returns the number of matches
			"count": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("count() takes exactly 2 arguments")
				}

				var re *regexp.Regexp
				var err error

				switch pattern := args[0].(type) {
				case *objects.String:
					re, err = regexp.Compile(pattern.Value)
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

				matches := re.FindAllString(s.Value, -1)
				return Int(int64(len(matches)))
			}),
		},
	})
}

// CompiledRegex represents a compiled regular expression
type CompiledRegex struct {
	Pattern string
	Re      *regexp.Regexp
}

func (cr *CompiledRegex) Type() objects.ObjectType { return objects.ObjectType("COMPILED_REGEX") }
func (cr *CompiledRegex) Inspect() string          { return fmt.Sprintf("[compiled regex: %s]", cr.Pattern) }
func (cr *CompiledRegex) ToBool() *objects.Bool    { return objects.TRUE }
func (cr *CompiledRegex) HashKey() objects.HashKey {
	return objects.HashKey{Type: objects.ObjectType("COMPILED_REGEX"), Value: 0}
}
