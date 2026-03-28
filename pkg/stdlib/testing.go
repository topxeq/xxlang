// pkg/stdlib/testing.go
// Testing utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/topxeq/xxlang/pkg/objects"
)

// testCounter is a global counter for test numbering
var testCounter int64

// getNextTestNum returns the next test number in sequence
func getNextTestNum() int64 {
	return atomic.AddInt64(&testCounter, 1)
}

// findFirstDiffIndex returns the index of the first differing character between two strings
func findFirstDiffIndex(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return minLen
	}
	return -1
}

func init() {
	Register(&Module{
		Name: "testing",
		Exports: map[string]objects.Object{
			// byText compares actual and expected strings.
			// Usage: testing.byText(actual, expected, testName?, testGroup?)
			"byText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byText() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				actualStr, ok1 := args[0].(*objects.String)
				expectedStr, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				actual := actualStr.Value
				expected := expectedStr.Value

				if actual == expected {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				diffPos := findFirstDiffIndex(actual, expected)
				return Error(fmt.Sprintf("test %s%s failed at position %d:\n-----\n%s\n-----\n%s", testName, testNum, diffPos, actual, expected))
			}),

			// byStartsWith checks if a string starts with a prefix.
			// Usage: testing.byStartsWith(str, prefix, testName?, testGroup?)
			"byStartsWith": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byStartsWith() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				str, ok1 := args[0].(*objects.String)
				prefix, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				if strings.HasPrefix(str.Value, prefix.Value) {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				return Error(fmt.Sprintf("test %s%s failed: string does not start with prefix\n-----\n%s\n-----\n%s", testName, testNum, str.Value, prefix.Value))
			}),

			// byEndsWith checks if a string ends with a suffix.
			// Usage: testing.byEndsWith(str, suffix, testName?, testGroup?)
			"byEndsWith": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byEndsWith() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				str, ok1 := args[0].(*objects.String)
				suffix, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				if strings.HasSuffix(str.Value, suffix.Value) {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				return Error(fmt.Sprintf("test %s%s failed: string does not end with suffix\n-----\n%s\n-----\n%s", testName, testNum, str.Value, suffix.Value))
			}),

			// byContains checks if a string contains a substring.
			// Usage: testing.byContains(str, substr, testName?, testGroup?)
			"byContains": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byContains() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				str, ok1 := args[0].(*objects.String)
				substr, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				if strings.Contains(str.Value, substr.Value) {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				return Error(fmt.Sprintf("test %s%s failed: string does not contain substring\n-----\n%s\n-----\n%s", testName, testNum, str.Value, substr.Value))
			}),

			// byReg checks if a string matches a regex pattern.
			// Usage: testing.byReg(str, pattern, testName?, testGroup?)
			"byReg": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byReg() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				str, ok1 := args[0].(*objects.String)
				pattern, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return Error(fmt.Sprintf("test %s%s failed: invalid regex pattern: %v", testName, testNum, err))
				}

				if re.MatchString(str.Value) {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				return Error(fmt.Sprintf("test %s%s failed: string does not match regex pattern\n-----\n%s\n-----\n%s", testName, testNum, str.Value, pattern.Value))
			}),

			// byRegContains checks if a string contains a regex match.
			// Usage: testing.byRegContains(str, pattern, testName?, testGroup?)
			"byRegContains": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("byRegContains() requires at least 2 arguments")
				}

				var testName string
				var testNum string
				if len(args) >= 3 {
					if s, ok := args[2].(*objects.String); ok {
						testName = s.Value
					}
				}
				if len(args) >= 4 {
					if s, ok := args[3].(*objects.String); ok {
						testNum = "(" + s.Value + ")"
					}
				}
				if testName == "" {
					testName = fmt.Sprintf("%d", getNextTestNum())
				}

				str, ok1 := args[0].(*objects.String)
				pattern, ok2 := args[1].(*objects.String)

				if !ok1 {
					return Error(fmt.Sprintf("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type()))
				}
				if !ok2 {
					return Error(fmt.Sprintf("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type()))
				}

				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return Error(fmt.Sprintf("test %s%s failed: invalid regex pattern: %v", testName, testNum, err))
				}

				match := re.FindString(str.Value)
				if match != "" {
					fmt.Printf("test %s%s passed\n", testName, testNum)
					return Null()
				}

				return Error(fmt.Sprintf("test %s%s failed: string does not contain regex match\n-----\n%s\n-----\n%s", testName, testNum, str.Value, pattern.Value))
			}),
		},
	})
}
