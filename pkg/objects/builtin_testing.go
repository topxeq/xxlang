// pkg/objects/builtin_testing.go
// Test assertion built-in functions for Xxlang
package objects

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
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
	// If all chars match up to minLen, difference is at minLen
	if len(a) != len(b) {
		return minLen
	}
	return -1 // strings are identical
}

func init() {
	// Test assertion functions
	Builtins["testByText"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByText. got=%d, want>=2", len(args))
			}

			// Get test name (optional 3rd argument)
			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			// Get actual and expected values as strings
			actualStr, ok1 := args[0].(*String)
			expectedStr, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			actual := actualStr.Value
			expected := expectedStr.Value

			if actual == expected {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			// Find position of first difference
			diffPos := findFirstDiffIndex(actual, expected)
			return newError("test %s%s failed at position %d:\n-----\n%s\n-----\n%s", testName, testNum, diffPos, actual, expected)
		},
	}

	Builtins["testByStartsWith"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByStartsWith. got=%d, want>=2", len(args))
			}

			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			str, ok1 := args[0].(*String)
			prefix, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			if strings.HasPrefix(str.Value, prefix.Value) {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			return newError("test %s%s failed: string does not start with prefix\n-----\n%s\n-----\n%s", testName, testNum, str.Value, prefix.Value)
		},
	}

	Builtins["testByEndsWith"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByEndsWith. got=%d, want>=2", len(args))
			}

			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			str, ok1 := args[0].(*String)
			suffix, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			if strings.HasSuffix(str.Value, suffix.Value) {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			return newError("test %s%s failed: string does not end with suffix\n-----\n%s\n-----\n%s", testName, testNum, str.Value, suffix.Value)
		},
	}

	Builtins["testByContains"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByContains. got=%d, want>=2", len(args))
			}

			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			str, ok1 := args[0].(*String)
			substr, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			if strings.Contains(str.Value, substr.Value) {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			return newError("test %s%s failed: string does not contain substring\n-----\n%s\n-----\n%s", testName, testNum, str.Value, substr.Value)
		},
	}

	Builtins["testByReg"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByReg. got=%d, want>=2", len(args))
			}

			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			str, ok1 := args[0].(*String)
			pattern, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			// Compile and match regex
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("test %s%s failed: invalid regex pattern: %v", testName, testNum, err)
			}

			if re.MatchString(str.Value) {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			return newError("test %s%s failed: string does not match regex pattern\n-----\n%s\n-----\n%s", testName, testNum, str.Value, pattern.Value)
		},
	}

	Builtins["testByRegContains"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for testByRegContains. got=%d, want>=2", len(args))
			}

			var testName string
			var testNum string
			if len(args) >= 3 {
				if s, ok := args[2].(*String); ok {
					testName = s.Value
				}
			}
			if len(args) >= 4 {
				if s, ok := args[3].(*String); ok {
					testNum = "(" + s.Value + ")"
				}
			}
			if testName == "" {
				testName = fmt.Sprintf("%d", getNextTestNum())
			}

			str, ok1 := args[0].(*String)
			pattern, ok2 := args[1].(*String)

			if !ok1 {
				return newError("test %s%s failed: first argument must be STRING, got %s", testName, testNum, args[0].Type())
			}
			if !ok2 {
				return newError("test %s%s failed: second argument must be STRING, got %s", testName, testNum, args[1].Type())
			}

			// Compile and find regex match
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("test %s%s failed: invalid regex pattern: %v", testName, testNum, err)
			}

			match := re.FindString(str.Value)
			if match != "" {
				fmt.Printf("test %s%s passed\n", testName, testNum)
				return NULL
			}

			return newError("test %s%s failed: string does not contain regex match\n-----\n%s\n-----\n%s", testName, testNum, str.Value, pattern.Value)
		},
	}

	Builtins["dumpVar"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for dumpVar. got=%d, want>=1", len(args))
			}

			// Dump the variable with detailed formatting
			fmt.Printf("Dump: %s\n", args[0].Inspect())
			fmt.Printf("Type: %s\n", args[0].Type())

			// For more detailed output, use JSON-like formatting
			switch v := args[0].(type) {
			case *Map:
				fmt.Println("Contents:")
				for _, pair := range v.Pairs {
					fmt.Printf("  %s: %s\n", pair.Key.Inspect(), pair.Value.Inspect())
				}
			case *Array:
				fmt.Println("Elements:")
				for i, elem := range v.Elements {
					fmt.Printf("  [%d]: %s\n", i, elem.Inspect())
				}
			case *String:
				fmt.Printf("Value: %q\n", v.Value)
				fmt.Printf("Length: %d\n", len(v.Value))
			case *Int:
				fmt.Printf("Value: %d\n", v.Value)
			case *Float:
				fmt.Printf("Value: %f\n", v.Value)
			case *Bool:
				fmt.Printf("Value: %v\n", v.Value)
			}

			return NULL
		},
	}

	// debugInfo requires VM state which is not available in pure builtin context
	// It will be implemented as a special builtin that gets VM state injected
	Builtins["debugInfo"] = &Builtin{
		Fn: func(args ...Object) Object {
			// This builtin needs VM context to work properly
			// In basic form, return what we can
			var sb strings.Builder
			sb.WriteString("=== Debug Info ===\n")
			sb.WriteString("Note: Full debug info requires VM context.\n")

			if len(args) > 0 {
				sb.WriteString("Arguments:\n")
				for i, arg := range args {
					sb.WriteString(fmt.Sprintf("  [%d]: %s (type: %s)\n", i, arg.Inspect(), arg.Type()))
				}
			}

			return &String{Value: sb.String()}
		},
	}
}