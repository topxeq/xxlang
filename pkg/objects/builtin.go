// pkg/objects/builtin.go
package objects

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BuiltinFunction is the type for built-in functions
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinType }
func (b *Builtin) TypeTag() TypeTag { return TagBuiltin }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) ToBool() *Bool    { return TRUE }
func (b *Builtin) HashKey() HashKey { return HashKey{Type: BuiltinType, Value: 0} }

// Builtins contains all built-in functions
var Builtins = map[string]*Builtin{
	// ============================================================
	// Basic Functions
	// ============================================================
	"len": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for len. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return NewInt(int64(len(arg.Value)))
			case *Chars:
				return NewInt(int64(len(arg.Value)))
			case *Array:
				return NewInt(int64(len(arg.Elements)))
			case *Map:
				return NewInt(int64(len(arg.Pairs)))
			default:
				return newError("argument to 'len' not supported, got %s", args[0].Type())
			}
		},
	},
	"pr": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},
	"pln": {
		Fn: func(args ...Object) Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},
	"pl": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for pl. got=%d, want>=1", len(args))
			}

			format, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'pl' must be STRING, got %s", args[0].Type())
			}

			formatArgs := make([]interface{}, len(args)-1)
			for i, arg := range args[1:] {
				switch v := arg.(type) {
				case *Int:
					formatArgs[i] = v.Value
				case *Float:
					formatArgs[i] = v.Value
				case *String:
					formatArgs[i] = v.Value
				case *Bool:
					formatArgs[i] = v.Value
				default:
					formatArgs[i] = v.Inspect()
				}
			}

			fmt.Printf(format.Value+"\n", formatArgs...)
			return NULL
		},
	},
	"prf": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for prf. got=%d, want>=1", len(args))
			}

			format, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'prf' must be STRING, got %s", args[0].Type())
			}

			formatArgs := make([]interface{}, len(args)-1)
			for i, arg := range args[1:] {
				switch v := arg.(type) {
				case *Int:
					formatArgs[i] = v.Value
				case *Float:
					formatArgs[i] = v.Value
				case *String:
					formatArgs[i] = v.Value
				case *Bool:
					formatArgs[i] = v.Value
				default:
					formatArgs[i] = v.Inspect()
				}
			}

			fmt.Printf(format.Value, formatArgs...)
			return NULL
		},
	},
	"checkErr": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for checkErr. got=%d, want>=1", len(args))
			}

			// Check if argument is an error
			if errObj, ok := args[0].(*Error); ok {
				if len(args) > 1 {
					// Optional message with format support
					if msg, ok := args[1].(*String); ok {
						// Check if message contains format verbs
						if strings.Contains(msg.Value, "%") {
							// Use error message as argument for format verbs
							fmt.Fprintf(os.Stderr, msg.Value+"\n", errObj.Message)
						} else {
							// Plain message, just print it
							fmt.Fprintln(os.Stderr, msg.Value)
						}
					} else {
						fmt.Fprintln(os.Stderr, args[1].Inspect())
					}
				} else {
					fmt.Fprintln(os.Stderr, errObj.Inspect())
				}
				os.Exit(1)
			}

			return NULL
		},
	},
	"checkEmpty": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for checkEmpty. got=%d, want>=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'checkEmpty' must be STRING, got %s", args[0].Type())
			}

			if str.Value == "" {
				if len(args) > 1 {
					// Optional error message
					if msg, ok := args[1].(*String); ok {
						fmt.Fprintln(os.Stderr, msg.Value)
					}
				}
				os.Exit(1)
			}

			return NULL
		},
	},
	"genOtpCode": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for genOtpCode. got=%d, want>=1", len(args))
			}

			secret, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'genOtpCode' must be STRING, got %s", args[0].Type())
			}

			// Decode base32 secret (uppercase, no spaces)
			encoded := strings.ToUpper(strings.ReplaceAll(secret.Value, " ", ""))
			decoder := base32.StdEncoding.WithPadding(base32.NoPadding)
			key, err := decoder.DecodeString(encoded)
			if err != nil {
				return &Error{Message: fmt.Sprintf("base32 decode failed: %v", err)}
			}

			// Calculate time step (30 second intervals)
			timestamp := time.Now().Unix() / 30

			// Convert timestamp to 8-byte big-endian
			counter := make([]byte, 8)
			binary.BigEndian.PutUint64(counter, uint64(timestamp))

			// HMAC-SHA1
			mac := hmac.New(sha1.New, key)
			mac.Write(counter)
			hash := mac.Sum(nil)

			// Dynamic truncation (RFC 4226)
			offset := hash[len(hash)-1] & 0x0f
			code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

			// Get 6-digit code
			otp := code % 1000000

			return NewString(fmt.Sprintf("%06d", otp))
		},
	},
	"typeOf": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for typeOf. got=%d, want>=1", len(args))
			}

			obj := args[0]

			// Check if detailed mode requested
			detailed := false
			if len(args) > 1 {
				if b, ok := args[1].(*Bool); ok {
					detailed = b.Value
				}
			}

			// For instances with detailed mode, return class name
			if detailed {
				if inst, ok := obj.(*Instance); ok {
					return NewString(inst.Class.Name)
				}
			}

			return NewString(string(obj.Type()))
		},
	},
	"toStr": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for toStr. got=%d, want=1", len(args))
			}
			return NewString(args[0].Inspect())
		},
	},
	"toChars": {
		Fn: func(args ...Object) Object {
			// toChars converts a string to chars ([]rune) for character-based operations
			if len(args) != 1 {
				return newError("wrong number of arguments for toChars. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'toChars' must be STRING, got %s", args[0].Type())
			}
			return NewCharsFromString(str.Value)
		},
	},
	"charLen": {
		Fn: func(args ...Object) Object {
			// charLen returns the number of Unicode characters (runes) in a string
			if len(args) != 1 {
				return newError("wrong number of arguments for charLen. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'charLen' must be STRING, got %s", args[0].Type())
			}
			return NewInt(int64(len([]rune(str.Value))))
		},
	},

	// ============================================================
	// String Functions
	// ============================================================
	"substr": {
		Fn: func(args ...Object) Object {
			// substr uses BYTE indices (Go-compatible behavior)
			// For character-based slicing, use toChars(s).subStr(start, end) instead
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for substr. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'substr' must be STRING, got %s", args[0].Type())
			}

			start, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'substr' must be INT, got %s", args[1].Type())
			}

			// Use byte indices (Go-compatible)
			strLen := int64(len(str.Value))
			end := strLen
			if len(args) == 3 {
				e, ok := args[2].(*Int)
				if !ok {
					return newError("third argument to 'substr' must be INT, got %s", args[2].Type())
				}
				end = e.Value
			}

			if start.Value < 0 || start.Value > strLen || end < start.Value || end > strLen {
				return newError("substring indices out of range")
			}

			return NewString(str.Value[start.Value:end])
		},
	},
	"split": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for split. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'split' must be STRING, got %s", args[0].Type())
			}

			sep, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'split' must be STRING, got %s", args[1].Type())
			}

			parts := strings.Split(str.Value, sep.Value)
			elements := make([]Object, len(parts))
			for i, part := range parts {
				elements[i] = NewString(part)
			}

			return NewArray(elements)
		},
	},
	"join": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for join. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'join' must be ARRAY, got %s", args[0].Type())
			}

			sep, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'join' must be STRING, got %s", args[1].Type())
			}

			parts := make([]string, len(arr.Elements))
			for i, elem := range arr.Elements {
				if s, ok := elem.(*String); ok {
					parts[i] = s.Value
				} else {
					parts[i] = elem.Inspect()
				}
			}

			return NewString(strings.Join(parts, sep.Value))
		},
	},
	"trim": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for trim. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'trim' must be STRING, got %s", args[0].Type())
			}

			return NewString(strings.TrimSpace(str.Value))
		},
	},
	"padLeft": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for padLeft. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'padLeft' must be STRING, got %s", args[0].Type())
			}

			width, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'padLeft' must be INT, got %s", args[1].Type())
			}

			padChar := " "
			if len(args) == 3 {
				pc, ok := args[2].(*String)
				if !ok {
					return newError("third argument to 'padLeft' must be STRING, got %s", args[2].Type())
				}
				if len(pc.Value) > 0 {
					padChar = pc.Value
				}
			}

			result := str.Value
			for int64(len(result)) < width.Value {
				result = padChar + result
			}
			return NewString(result)
		},
	},
	"padRight": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for padRight. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'padRight' must be STRING, got %s", args[0].Type())
			}

			width, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'padRight' must be INT, got %s", args[1].Type())
			}

			padChar := " "
			if len(args) == 3 {
				pc, ok := args[2].(*String)
				if !ok {
					return newError("third argument to 'padRight' must be STRING, got %s", args[2].Type())
				}
				if len(pc.Value) > 0 {
					padChar = pc.Value
				}
			}

			result := str.Value
			for int64(len(result)) < width.Value {
				result = result + padChar
			}
			return NewString(result)
		},
	},
	"upper": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for upper. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'upper' must be STRING, got %s", args[0].Type())
			}

			return NewString(strings.ToUpper(str.Value))
		},
	},
	"lower": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for lower. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'lower' must be STRING, got %s", args[0].Type())
			}

			return NewString(strings.ToLower(str.Value))
		},
	},
	"containsStr": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'contains' must be STRING, got %s", args[0].Type())
			}

			substr, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'contains' must be STRING, got %s", args[1].Type())
			}

			return &Bool{Value: strings.Contains(str.Value, substr.Value)}
		},
	},
	"replace": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for replace. got=%d, want=3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'replace' must be STRING, got %s", args[0].Type())
			}

			old, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'replace' must be STRING, got %s", args[1].Type())
			}

			newStr, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'replace' must be STRING, got %s", args[2].Type())
			}

			return NewString(strings.ReplaceAll(str.Value, old.Value, newStr.Value))
		},
	},
	"startsWith": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for startsWith. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'startsWith' must be STRING, got %s", args[0].Type())
			}

			prefix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'startsWith' must be STRING, got %s", args[1].Type())
			}

			return &Bool{Value: strings.HasPrefix(str.Value, prefix.Value)}
		},
	},
	"endsWith": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for endsWith. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'endsWith' must be STRING, got %s", args[0].Type())
			}

			suffix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'endsWith' must be STRING, got %s", args[1].Type())
			}

			return &Bool{Value: strings.HasSuffix(str.Value, suffix.Value)}
		},
	},

	// ============================================================
	// Math Functions
	// ============================================================
	"abs": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Int:
				if arg.Value < 0 {
					return NewInt(-arg.Value)
				}
				return arg
			case *Float:
				if arg.Value < 0 {
					return NewFloat(-arg.Value)
				}
				return arg
			default:
				return newError("argument to 'abs' must be INT or FLOAT, got %s", args[0].Type())
			}
		},
	},
	"floor": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for floor. got=%d, want=1", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Int:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("argument to 'floor' must be INT or FLOAT, got %s", args[0].Type())
			}

			return NewInt(int64(math.Floor(val)))
		},
	},
	"ceil": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for ceil. got=%d, want=1", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Int:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("argument to 'ceil' must be INT or FLOAT, got %s", args[0].Type())
			}

			return NewInt(int64(math.Ceil(val)))
		},
	},
	"sqrt": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sqrt. got=%d, want=1", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Int:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("argument to 'sqrt' must be INT or FLOAT, got %s", args[0].Type())
			}
			if val < 0 {
				return newError("cannot calculate square root of negative number")
			}

			return NewFloat(math.Sqrt(val))
		},
	},
	"pow": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for pow. got=%d, want=2", len(args))
			}

			var base, exp float64
			switch a := args[0].(type) {
			case *Int:
				base = float64(a.Value)
			case *Float:
				base = a.Value
			default:
				return newError("first argument to 'pow' must be INT or FLOAT, got %s", args[0].Type())
			}

			switch a := args[1].(type) {
			case *Int:
				exp = float64(a.Value)
			case *Float:
				exp = a.Value
			default:
				return newError("second argument to 'pow' must be INT or FLOAT, got %s", args[1].Type())
			}

			return NewFloat(math.Pow(base, exp))
		},
	},
	"min": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for min. got=%d, want=2", len(args))
			}

			var a, b float64
			switch arg := args[0].(type) {
			case *Int:
				a = float64(arg.Value)
			case *Float:
				a = arg.Value
			default:
				return newError("first argument to 'min' must be INT or FLOAT, got %s", args[0].Type())
			}

			switch arg := args[1].(type) {
			case *Int:
				b = float64(arg.Value)
			case *Float:
				b = arg.Value
			default:
				return newError("second argument to 'min' must be INT or FLOAT, got %s", args[1].Type())
			}

			if a <= b {
				return args[0]
			}
			return args[1]
		},
	},
	"max": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for max. got=%d, want=2", len(args))
			}

			var a, b float64
			switch arg := args[0].(type) {
			case *Int:
				a = float64(arg.Value)
			case *Float:
				a = arg.Value
			default:
				return newError("first argument to 'max' must be INT or FLOAT, got %s", args[0].Type())
			}

			switch arg := args[1].(type) {
			case *Int:
				b = float64(arg.Value)
			case *Float:
				b = arg.Value
			default:
				return newError("second argument to 'max' must be INT or FLOAT, got %s", args[1].Type())
			}

			if a >= b {
				return args[0]
			}
			return args[1]
		},
	},

	// ============================================================
	// Type Conversion Functions
	// ============================================================
	"int": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for int. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Int:
				return arg
			case *Float:
				return NewInt(int64(arg.Value))
			case *String:
				val, err := strconv.ParseInt(arg.Value, 10, 64)
				if err != nil {
					return newError("could not convert string to int: %s", arg.Value)
				}
				return NewInt(val)
			case *Bool:
				if arg.Value {
					return NewInt(1)
				}
				return NewInt(0)
			default:
				return newError("cannot convert %s to INT", args[0].Type())
			}
		},
	},
	"float": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for float. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Int:
				return NewFloat(float64(arg.Value))
			case *Float:
				return arg
			case *String:
				val, err := strconv.ParseFloat(arg.Value, 64)
				if err != nil {
					return newError("could not convert string to float: %s", arg.Value)
				}
				return NewFloat(val)
			case *Bool:
				if arg.Value {
					return NewFloat(1.0)
				}
				return NewFloat(0.0)
			default:
				return newError("cannot convert %s to FLOAT", args[0].Type())
			}
		},
	},
	"string": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for string. got=%d, want=1", len(args))
			}
			return NewString(args[0].Inspect())
		},
	},

	// ============================================================
	// Array Functions
	// ============================================================
	"push": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for push. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'push' must be ARRAY, got %s", args[0].Type())
			}
			// Modify array in place and return it for chaining
			arr.Elements = append(arr.Elements, args[1])
			return arr
		},
	},
	"pop": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for pop. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'pop' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return newError("cannot pop from empty array")
			}
			// Get last element
			lastElem := arr.Elements[len(arr.Elements)-1]
			// Modify array in place
			arr.Elements = arr.Elements[:len(arr.Elements)-1]
			// Return the popped element
			return lastElem
		},
	},
	"first": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for first. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'first' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			return arr.Elements[0]
		},
	},
	"last": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for last. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'last' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			return arr.Elements[len(arr.Elements)-1]
		},
	},
	"rest": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for rest. got=%d, want=2 or 3", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'rest' must be ARRAY, got %s", args[0].Type())
			}

			start, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'rest' must be INT, got %s", args[1].Type())
			}
			arrLen := int64(len(arr.Elements))
			end := arrLen
			if len(args) == 3 {
				e, ok := args[2].(*Int)
				if !ok {
					return newError("third argument to 'rest' must be INT, got %s", args[2].Type())
				}
				end = e.Value
			}

			if start.Value < 0 || start.Value > arrLen || end < start.Value || end > arrLen {
				return newError("slice indices out of range")
			}
			return NewArray(arr.Elements[start.Value:end])
		},
	},
	"concat": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for concat. got=%d, want=2", len(args))
			}

			arr1, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'concat' must be ARRAY, got %s", args[0].Type())
			}
			arr2, ok := args[1].(*Array)
			if !ok {
				return newError("second argument to 'concat' must be ARRAY, got %s", args[1].Type())
			}
			newElements := make([]Object, len(arr1.Elements)+len(arr2.Elements))
			copy(newElements, arr1.Elements)
			copy(newElements[len(arr1.Elements):], arr2.Elements)
			return NewArray(newElements)
		},
	},
	"indexOf": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
			}

			// Support both array and string
			switch obj := args[0].(type) {
			case *Array:
				for i, elem := range obj.Elements {
					if compareObjects(elem, args[1]) {
						return NewInt(int64(i))
					}
				}
				return NewInt(-1)
			case *String:
				substr, ok := args[1].(*String)
				if !ok {
					return newError("second argument to 'indexOf' for string must be STRING, got %s", args[1].Type())
				}
				idx := strings.Index(obj.Value, substr.Value)
				return NewInt(int64(idx))
			default:
				return newError("first argument to 'indexOf' must be ARRAY or STRING, got %s", args[0].Type())
			}
		},
	},
	"containsArr": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for containsArr. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'containsArr' must be ARRAY, got %s", args[0].Type())
			}

			for _, elem := range arr.Elements {
				if compareObjects(elem, args[1]) {
					return TRUE
				}
			}
			return FALSE
		},
	},

	// ============================================================
	// Map Functions
	// ============================================================
	"keys": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for keys. got=%d, want=1", len(args))
			}

			m, ok := args[0].(*Map)
			if !ok {
				return newError("argument to 'keys' must be MAP, got %s", args[0].Type())
			}
			keys := make([]Object, len(m.Pairs))
			i := 0
			for _, pair := range m.Pairs {
				keys[i] = pair.Key
				i++
			}
			// Sort keys for deterministic output
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].Inspect() < keys[j].Inspect()
			})
			return NewArray(keys)
		},
	},
	"values": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for values. got=%d, want=1", len(args))
			}

			m, ok := args[0].(*Map)
			if !ok {
				return newError("argument to 'values' must be MAP, got %s", args[0].Type())
			}
			// Get keys and sort them for deterministic order
			keys := make([]Object, len(m.Pairs))
			i := 0
			for _, pair := range m.Pairs {
				keys[i] = pair.Key
				i++
			}
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].Inspect() < keys[j].Inspect()
			})
			// Get values in the same order as sorted keys
			vals := make([]Object, len(keys))
			for i, key := range keys {
				vals[i] = m.Pairs[key.HashKey()].Value
			}
			return NewArray(vals)
		},
	},
	"hasKey": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for hasKey. got=%d, want=2", len(args))
			}

			m, ok := args[0].(*Map)
			if !ok {
				return newError("first argument to 'hasKey' must be MAP, got %s", args[0].Type())
			}
			_, exists := m.Pairs[args[1].HashKey()]
			return &Bool{Value: exists}
		},
	},
	"delete": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
			}

			m, ok := args[0].(*Map)
			if !ok {
				return newError("first argument to 'delete' must be MAP, got %s", args[0].Type())
			}
			// Delete in place (imperative style)
			delete(m.Pairs, args[1].HashKey())
			return NULL
		},
	},

	// ============================================================
	// Utility Functions
	// ============================================================
	"range": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 3 {
				return newError("wrong number of arguments for range. got=%d, want=1, 2, or 3", len(args))
			}
			var start, end, step int64 = 0, 0, 1
			switch len(args) {
			case 1:
				e, ok := args[0].(*Int)
				if !ok {
					return newError("argument to 'range' must be INT, got %s", args[0].Type())
				}
				start = 0
				end = e.Value
			case 2:
				s, ok := args[0].(*Int)
				if !ok {
					return newError("first argument to 'range' must be INT, got %s", args[0].Type())
				}
				e, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'range' must be INT, got %s", args[1].Type())
				}
				start = s.Value
				end = e.Value
			case 3:
				s, ok := args[0].(*Int)
				if !ok {
					return newError("first argument to 'range' must be INT, got %s", args[0].Type())
				}
				e, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'range' must be INT, got %s", args[1].Type())
				}
				st, ok := args[2].(*Int)
				if !ok {
					return newError("third argument to 'range' must be INT, got %s", args[2].Type())
				}
				start = s.Value
				end = e.Value
				step = st.Value
				if step == 0 {
					return newError("step cannot be zero")
				}
			}
			elements := make([]Object, 0)
			if step == 1 {
				// Default behavior: inclusive range
				if start <= end {
					elements = make([]Object, end-start+1)
					for i := start; i <= end; i++ {
						elements[i-start] = NewInt(i)
					}
				} else {
					elements = make([]Object, start-end+1)
					for i := start; i >= end; i-- {
						elements[start-i] = NewInt(i)
					}
				}
			} else {
				// With custom step
				if step > 0 {
					if start < end {
						for i := start; i < end; i += step {
							elements = append(elements, NewInt(i))
						}
					}
				} else {
					if start > end {
						for i := start; i > end; i += step {
							elements = append(elements, NewInt(i))
						}
					}
				}
			}
			return NewArray(elements)
		},
	},
	"sort": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sort. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'sort' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return arr
			}
			// Copy and sort using Inspect for comparison
			sorted := make([]Object, len(arr.Elements))
			copy(sorted, arr.Elements)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].Inspect() > sorted[j].Inspect() {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			return NewArray(sorted)
		},
	},
	"sum": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sum. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'sum' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return NewInt(0)
			}
			var total int64
			var hasFloat bool
			for _, elem := range arr.Elements {
				switch e := elem.(type) {
				case *Int:
					total += e.Value
				case *Float:
					total += int64(e.Value)
					hasFloat = true
				default:
					return newError("array elements must be INT or FLOAT for sum, got %s", elem.Type())
				}
			}
			if hasFloat {
				return NewFloat(float64(total))
			}
			return NewInt(total)
		},
	},
	"avg": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for avg. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'avg' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return NewFloat(0)
			}
			var total float64
			for _, elem := range arr.Elements {
				switch e := elem.(type) {
				case *Int:
					total += float64(e.Value)
				case *Float:
					total += e.Value
				default:
					return newError("array elements must be INT or FLOAT for avg, got %s", elem.Type())
				}
			}
			return NewFloat(total / float64(len(arr.Elements)))
		},
	},
	"reverse": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'reverse' must be ARRAY, got %s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return arr
			}
			reversed := make([]Object, len(arr.Elements))
			for i := 0; i < len(arr.Elements); i++ {
				reversed[i] = arr.Elements[len(arr.Elements)-1-i]
			}
			return NewArray(reversed)
		},
	},

	// ============================================================
	// Dynamic Code Execution
	// ============================================================
	"runCode": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for runCode. got=%d, want=1 or 2", len(args))
			}

			code, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'runCode' must be STRING, got %s", args[0].Type())
			}

			// Optional second argument: a map of variables to pass to the code
			var argMap *Map
			if len(args) == 2 {
				argMap, ok = args[1].(*Map)
				if !ok {
					return newError("second argument to 'runCode' must be MAP, got %s", args[1].Type())
				}
			}

			// Use the registered callback if available
			if runCodeImpl != nil {
				result, err := runCodeImpl(code.Value, argMap)
				if err != nil {
					return newError("runCode error: %v", err)
				}
				if result == nil {
					return NULL
				}
				return result
			}

			return newError("runCode not available in this context")
		},
	},

	// ============================================================
	// Delegate - Dynamic Function Creation
	// ============================================================
	"delegate": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for delegate. got=%d, want=1", len(args))
			}

			source, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'delegate' must be STRING, got %s", args[0].Type())
			}

			// Use the registered callback if available
			if delegateImpl != nil {
				result, err := delegateImpl(source.Value)
				if err != nil {
					return newError("delegate error: %v", err)
				}
				if result == nil {
					return NULL
				}
				return result
			}

			return newError("delegate not available in this context")
		},
	},

	// ============================================================
	// Plugin Loading
	// ============================================================
	"loadPlugin": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for loadPlugin. got=%d, want=1", len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'loadPlugin' must be STRING (file path), got %s", args[0].Type())
			}

			// Use the registered callback if available
			if loadPluginImpl != nil {
				result, err := loadPluginImpl(path.Value)
				if err != nil {
					return newError("loadPlugin error: %v", err)
				}
				if result == nil {
					return NULL
				}
				return result
			}

			return newError("loadPlugin not available in this context")
		},
	},

	// ============================================================
	// String Utility Functions
	// ============================================================
	"repeat": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for repeat. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'repeat' must be STRING, got %s", args[0].Type())
			}

			count, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'repeat' must be INT, got %s", args[1].Type())
			}

			if count.Value < 0 {
				return newError("repeat count cannot be negative")
			}

			return NewString(strings.Repeat(str.Value, int(count.Value)))
		},
	},
	"lpad": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for lpad. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'lpad' must be STRING, got %s", args[0].Type())
			}

			length, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'lpad' must be INT, got %s", args[1].Type())
			}

			padChar := " "
			if len(args) == 3 {
				pad, ok := args[2].(*String)
				if !ok {
					return newError("third argument to 'lpad' must be STRING, got %s", args[2].Type())
				}
				if len(pad.Value) > 0 {
					padChar = pad.Value
				}
			}

			strLen := len(str.Value)
			targetLen := int(length.Value)
			if strLen >= targetLen {
				return str
			}

			padLen := targetLen - strLen
			padding := strings.Repeat(padChar, (padLen+len(padChar)-1)/len(padChar))
			return NewString(padding[:padLen] + str.Value)
		},
	},
	"rpad": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for rpad. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'rpad' must be STRING, got %s", args[0].Type())
			}

			length, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'rpad' must be INT, got %s", args[1].Type())
			}

			padChar := " "
			if len(args) == 3 {
				pad, ok := args[2].(*String)
				if !ok {
					return newError("third argument to 'rpad' must be STRING, got %s", args[2].Type())
				}
				if len(pad.Value) > 0 {
					padChar = pad.Value
				}
			}

			strLen := len(str.Value)
			targetLen := int(length.Value)
			if strLen >= targetLen {
				return str
			}

			padLen := targetLen - strLen
			padding := strings.Repeat(padChar, (padLen+len(padChar)-1)/len(padChar))
			return NewString(str.Value + padding[:padLen])
		},
	},
	"charAt": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for charAt. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'charAt' must be STRING, got %s", args[0].Type())
			}

			index, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'charAt' must be INT, got %s", args[1].Type())
			}

			idx := int(index.Value)
			if idx < 0 || idx >= len(str.Value) {
				return NULL
			}

			return NewString(string(str.Value[idx]))
		},
	},
	"trimLeft": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for trimLeft. got=%d, want=1 or 2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'trimLeft' must be STRING, got %s", args[0].Type())
			}

			cutset := " \t\n\r"
			if len(args) == 2 {
				cs, ok := args[1].(*String)
				if !ok {
					return newError("second argument to 'trimLeft' must be STRING, got %s", args[1].Type())
				}
				cutset = cs.Value
			}

			return NewString(strings.TrimLeft(str.Value, cutset))
		},
	},
	"trimRight": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for trimRight. got=%d, want=1 or 2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'trimRight' must be STRING, got %s", args[0].Type())
			}

			cutset := " \t\n\r"
			if len(args) == 2 {
				cs, ok := args[1].(*String)
				if !ok {
					return newError("second argument to 'trimRight' must be STRING, got %s", args[1].Type())
				}
				cutset = cs.Value
			}

			return NewString(strings.TrimRight(str.Value, cutset))
		},
	},

	// ============================================================
	// Type Checking Functions
	// ============================================================
	"isEmpty": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return &Bool{Value: len(arg.Value) == 0}
			case *Array:
				return &Bool{Value: len(arg.Elements) == 0}
			case *Map:
				return &Bool{Value: len(arg.Pairs) == 0}
			case *Null:
				return TRUE
			default:
				return FALSE
			}
		},
	},
	"isString": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isString. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*String)
			return &Bool{Value: ok}
		},
	},
	"isNumber": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isNumber. got=%d, want=1", len(args))
			}
			_, isInt := args[0].(*Int)
			_, isFloat := args[0].(*Float)
			return &Bool{Value: isInt || isFloat}
		},
	},
	"isInt": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isInt. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Int)
			return &Bool{Value: ok}
		},
	},
	"isFloat": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isFloat. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Float)
			return &Bool{Value: ok}
		},
	},
	"isArray": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isArray. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Array)
			return &Bool{Value: ok}
		},
	},
	"isMap": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isMap. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Map)
			return &Bool{Value: ok}
		},
	},
	"isBool": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isBool. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Bool)
			return &Bool{Value: ok}
		},
	},
	"isFunction": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isFunction. got=%d, want=1", len(args))
			}
			// Check for known function types
			switch args[0].(type) {
			case *Builtin, *CompiledFunction:
				return TRUE
			default:
				// Also check by type name for Closure (from vm package)
				t := string(args[0].Type())
				if t == "CLOSURE" || t == "FUNCTION" {
					return TRUE
				}
				return FALSE
			}
		},
	},
	"isNull": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isNull. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Null)
			return &Bool{Value: ok}
		},
	},

	// ============================================================
	// Math Functions
	// ============================================================
	"round": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for round. got=%d, want=1 or 2", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Int:
				return arg
			case *Float:
				val = arg.Value
			default:
				return newError("argument to 'round' must be INT or FLOAT, got %s", args[0].Type())
			}

			precision := 0
			if len(args) == 2 {
				p, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'round' must be INT, got %s", args[1].Type())
				}
				precision = int(p.Value)
			}

			if precision == 0 {
				return NewInt(int64(math.Round(val)))
			}

			multiplier := math.Pow(10, float64(precision))
			result := math.Round(val*multiplier) / multiplier
			return NewFloat(result)
		},
	},
	"clamp": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for clamp. got=%d, want=3", len(args))
			}

			var val, minVal, maxVal float64

			switch arg := args[0].(type) {
			case *Int:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("first argument to 'clamp' must be numeric, got %s", args[0].Type())
			}

			switch arg := args[1].(type) {
			case *Int:
				minVal = float64(arg.Value)
			case *Float:
				minVal = arg.Value
			default:
				return newError("second argument to 'clamp' must be numeric, got %s", args[1].Type())
			}

			switch arg := args[2].(type) {
			case *Int:
				maxVal = float64(arg.Value)
			case *Float:
				maxVal = arg.Value
			default:
				return newError("third argument to 'clamp' must be numeric, got %s", args[2].Type())
			}

			result := val
			if result < minVal {
				result = minVal
			}
			if result > maxVal {
				result = maxVal
			}

			// Return same type as input
			if _, isInt := args[0].(*Int); isInt {
				return NewInt(int64(result))
			}
			return NewFloat(result)
		},
	},
	"sign": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sign. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Int:
				if arg.Value > 0 {
					return NewInt(1)
				} else if arg.Value < 0 {
					return NewInt(-1)
				}
				return NewInt(0)
			case *Float:
				if arg.Value > 0 {
					return NewInt(1)
				} else if arg.Value < 0 {
					return NewInt(-1)
				}
				return NewInt(0)
			default:
				return newError("argument to 'sign' must be numeric, got %s", args[0].Type())
			}
		},
	},
	"random": {
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments for random. got=%d, want=0", len(args))
			}
			return NewFloat(float64(randInt63()) / float64(1<<63))
		},
	},
	"randomInt": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for randomInt. got=%d, want=2", len(args))
			}

			min, ok := args[0].(*Int)
			if !ok {
				return newError("first argument to 'randomInt' must be INT, got %s", args[0].Type())
			}

			max, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'randomInt' must be INT, got %s", args[1].Type())
			}

			if min.Value > max.Value {
				return newError("min cannot be greater than max")
			}

			range_ := max.Value - min.Value + 1
			result := min.Value + (randInt63() % range_)
			return NewInt(result)
		},
	},

	// ============================================================
	// Array Utility Functions
	// ============================================================
	"unique": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for unique. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'unique' must be ARRAY, got %s", args[0].Type())
			}

			seen := make(map[string]bool)
			result := make([]Object, 0)

			for _, elem := range arr.Elements {
				key := elem.Inspect()
				if !seen[key] {
					seen[key] = true
					result = append(result, elem)
				}
			}

			return NewArray(result)
		},
	},
	"flatten": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for flatten. got=%d, want=1 or 2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'flatten' must be ARRAY, got %s", args[0].Type())
			}

			depth := int64(1) // default: flatten one level only
			if len(args) == 2 {
				d, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'flatten' must be INT, got %s", args[1].Type())
				}
				depth = d.Value
			}

			result := flattenArray(arr.Elements, depth)
			return NewArray(result)
		},
	},
	"without": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for without. got=%d, want>=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'without' must be ARRAY, got %s", args[0].Type())
			}

			excludeSet := make(map[string]bool)
			for _, arg := range args[1:] {
				excludeSet[arg.Inspect()] = true
			}

			result := make([]Object, 0)
			for _, elem := range arr.Elements {
				if !excludeSet[elem.Inspect()] {
					result = append(result, elem)
				}
			}

			return NewArray(result)
		},
	},
	"take": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for take. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'take' must be ARRAY, got %s", args[0].Type())
			}

			n, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'take' must be INT, got %s", args[1].Type())
			}

			count := int(n.Value)
			if count < 0 {
				count = 0
			}
			if count > len(arr.Elements) {
				count = len(arr.Elements)
			}

			return NewArray(arr.Elements[:count])
		},
	},
	"drop": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for drop. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'drop' must be ARRAY, got %s", args[0].Type())
			}

			n, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'drop' must be INT, got %s", args[1].Type())
			}

			count := int(n.Value)
			if count < 0 {
				count = 0
			}
			if count > len(arr.Elements) {
				count = len(arr.Elements)
			}

			return NewArray(arr.Elements[count:])
		},
	},

	// ============================================================
	// Map Utility Functions
	// ============================================================
	"merge": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for merge. got=%d, want>=2", len(args))
			}

			result := make(map[HashKey]MapPair)

			for _, arg := range args {
				m, ok := arg.(*Map)
				if !ok {
					return newError("all arguments to 'merge' must be MAP, got %s", arg.Type())
				}

				for key, pair := range m.Pairs {
					result[key] = pair
				}
			}

			return NewMap(result)
		},
	},
	"entries": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for entries. got=%d, want=1", len(args))
			}

			m, ok := args[0].(*Map)
			if !ok {
				return newError("argument to 'entries' must be MAP, got %s", args[0].Type())
			}

			result := make([]Object, 0, len(m.Pairs))
			for _, pair := range m.Pairs {
				result = append(result, NewArray([]Object{pair.Key, pair.Value}))
			}

			return NewArray(result)
		},
	},

	// ============================================================
	// Utility Functions
	// ============================================================
	"format": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for format. got=%d, want>=1", len(args))
			}

			format, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'format' must be STRING, got %s", args[0].Type())
			}

			formatArgs := make([]interface{}, len(args)-1)
			for i, arg := range args[1:] {
				switch v := arg.(type) {
				case *Int:
					formatArgs[i] = v.Value
				case *Float:
					formatArgs[i] = v.Value
				case *String:
					formatArgs[i] = v.Value
				case *Bool:
					formatArgs[i] = v.Value
				default:
					formatArgs[i] = v.Inspect()
				}
			}

			return NewString(fmt.Sprintf(format.Value, formatArgs...))
		},
	},

	// ============================================================
	// Object Utility Functions
	// ============================================================
	"copy": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for copy. got=%d, want=1", len(args))
			}
			return deepCopyObject(args[0])
		},
	},
	"clone": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
			}
			return shallowCopyObject(args[0])
		},
	},
	"equals": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for equals. got=%d, want=2", len(args))
			}
			return &Bool{Value: deepEquals(args[0], args[1])}
		},
	},
	"defaults": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for defaults. got=%d, want=2", len(args))
			}

			target, ok := args[0].(*Map)
			if !ok {
				return newError("first argument to 'defaults' must be MAP, got %s", args[0].Type())
			}

			defaults, ok := args[1].(*Map)
			if !ok {
				return newError("second argument to 'defaults' must be MAP, got %s", args[1].Type())
			}

			result := make(map[HashKey]MapPair)
			for k, v := range target.Pairs {
				result[k] = v
			}
			for k, v := range defaults.Pairs {
				if _, exists := result[k]; !exists {
					result[k] = v
				}
			}

			return NewMap(result)
		},
	},

	// ============================================================
	// Encoding & Hash Functions
	// ============================================================
	"base64Encode": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for base64Encode. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'base64Encode' must be STRING, got %s", args[0].Type())
			}

			return NewString(base64.StdEncoding.EncodeToString([]byte(str.Value)))
		},
	},
	"base64Decode": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for base64Decode. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'base64Decode' must be STRING, got %s", args[0].Type())
			}

			decoded, err := base64.StdEncoding.DecodeString(str.Value)
			if err != nil {
				return newError("base64Decode failed: %s", err.Error())
			}

			return NewString(string(decoded))
		},
	},
	"hexEncode": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for hexEncode. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'hexEncode' must be STRING, got %s", args[0].Type())
			}

			return NewString(hex.EncodeToString([]byte(str.Value)))
		},
	},
	"hexDecode": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for hexDecode. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'hexDecode' must be STRING, got %s", args[0].Type())
			}

			decoded, err := hex.DecodeString(str.Value)
			if err != nil {
				return newError("hexDecode failed: %s", err.Error())
			}

			return NewString(string(decoded))
		},
	},
	"md5": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for md5. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'md5' must be STRING, got %s", args[0].Type())
			}

			hash := md5.Sum([]byte(str.Value))
			return NewString(hex.EncodeToString(hash[:]))
		},
	},
	"sha256": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sha256. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'sha256' must be STRING, got %s", args[0].Type())
			}

			hash := sha256.Sum256([]byte(str.Value))
			return NewString(hex.EncodeToString(hash[:]))
		},
	},

	// ============================================================
	// Time & UUID Functions
	// ============================================================
	"sleep": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for sleep. got=%d, want=1", len(args))
			}

			ms, ok := args[0].(*Int)
			if !ok {
				return newError("argument to 'sleep' must be INT (milliseconds), got %s", args[0].Type())
			}

			time.Sleep(time.Duration(ms.Value) * time.Millisecond)
			return NULL
		},
	},
	"now": {
		Fn: func(args ...Object) Object {
			return NewInt(time.Now().Unix())
		},
	},
	"nowMs": {
		Fn: func(args ...Object) Object {
			return NewInt(time.Now().UnixMilli())
		},
	},
	"uuid": {
		Fn: func(args ...Object) Object {
			initRand()
			randMu.Lock()
			defer randMu.Unlock()

			b := make([]byte, 16)
			randSrc.Read(b)
			// Version 4
			b[6] = (b[6] & 0x0f) | 0x40
			// Variant
			b[8] = (b[8] & 0x3f) | 0x80

			return NewString(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
		},
	},

	// ============================================================
	// String Enhancement Functions
	// ============================================================
	"trimPrefix": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for trimPrefix. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'trimPrefix' must be STRING, got %s", args[0].Type())
			}

			prefix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'trimPrefix' must be STRING, got %s", args[1].Type())
			}

			return NewString(strings.TrimPrefix(str.Value, prefix.Value))
		},
	},
	"trimSuffix": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for trimSuffix. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'trimSuffix' must be STRING, got %s", args[0].Type())
			}

			suffix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'trimSuffix' must be STRING, got %s", args[1].Type())
			}

			return NewString(strings.TrimSuffix(str.Value, suffix.Value))
		},
	},
	"count": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for count. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'count' must be STRING, got %s", args[0].Type())
			}

			substr, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'count' must be STRING, got %s", args[1].Type())
			}

			return NewInt(int64(strings.Count(str.Value, substr.Value)))
		},
	},
	"isDigit": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isDigit. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'isDigit' must be STRING, got %s", args[0].Type())
			}

			if len(str.Value) == 0 {
				return FALSE
			}

			for _, r := range str.Value {
				if r < '0' || r > '9' {
					return FALSE
				}
			}
			return TRUE
		},
	},
	"isAlpha": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isAlpha. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'isAlpha' must be STRING, got %s", args[0].Type())
			}

			if len(str.Value) == 0 {
				return FALSE
			}

			for _, r := range str.Value {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
					return FALSE
				}
			}
			return TRUE
		},
	},
	"isAlphaNum": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isAlphaNum. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'isAlphaNum' must be STRING, got %s", args[0].Type())
			}

			if len(str.Value) == 0 {
				return FALSE
			}

			for _, r := range str.Value {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
					return FALSE
				}
			}
			return TRUE
		},
	},

	// ============================================================
	// Array Enhancement Functions
	// ============================================================
	"find": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for find. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'find' must be ARRAY, got %s", args[0].Type())
			}

			for _, elem := range arr.Elements {
				if deepEquals(elem, args[1]) {
					return elem
				}
			}

			return NULL
		},
	},
	"findIndex": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for findIndex. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'findIndex' must be ARRAY, got %s", args[0].Type())
			}

			for i, elem := range arr.Elements {
				if deepEquals(elem, args[1]) {
					return NewInt(int64(i))
				}
			}

			return NewInt(-1)
		},
	},
	"includes": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for includes. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'includes' must be ARRAY, got %s", args[0].Type())
			}

			for _, elem := range arr.Elements {
				if deepEquals(elem, args[1]) {
					return TRUE
				}
			}

			return FALSE
		},
	},
	"shuffle": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for shuffle. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'shuffle' must be ARRAY, got %s", args[0].Type())
			}

			initRand()
			randMu.Lock()
			defer randMu.Unlock()

			result := make([]Object, len(arr.Elements))
			copy(result, arr.Elements)

			randSrc.Shuffle(len(result), func(i, j int) {
				result[i], result[j] = result[j], result[i]
			})

			return NewArray(result)
		},
	},
	"sample": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments for sample. got=%d, want=1 or 2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'sample' must be ARRAY, got %s", args[0].Type())
			}

			if len(arr.Elements) == 0 {
				return NULL
			}

			initRand()
			randMu.Lock()
			defer randMu.Unlock()

			count := int64(1)
			if len(args) == 2 {
				c, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to 'sample' must be INT, got %s", args[1].Type())
				}
				count = c.Value
			}

			if count == 1 {
				idx := randSrc.Intn(len(arr.Elements))
				return arr.Elements[idx]
			}

			result := make([]Object, 0, count)
			indices := randSrc.Perm(len(arr.Elements))
			for i := int64(0); i < count && i < int64(len(indices)); i++ {
				result = append(result, arr.Elements[indices[i]])
			}

			return NewArray(result)
		},
	},
	"chunk": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for chunk. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'chunk' must be ARRAY, got %s", args[0].Type())
			}

			size, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'chunk' must be INT, got %s", args[1].Type())
			}

			if size.Value <= 0 {
				return newError("chunk size must be positive")
			}

			chunks := make([]Object, 0)
			for i := 0; i < len(arr.Elements); i += int(size.Value) {
				end := i + int(size.Value)
				if end > len(arr.Elements) {
					end = len(arr.Elements)
				}
				chunks = append(chunks, NewArray(arr.Elements[i:end]))
			}

			return NewArray(chunks)
		},
	},

	// ============================================================
	// Command Line Argument Functions
	// ============================================================
	"getSwitch": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for getSwitch. got=%d, want=3", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'getSwitch' must be ARRAY, got %s", args[0].Type())
			}

			prefix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'getSwitch' must be STRING, got %s", args[1].Type())
			}

			// Third argument is the default value
			defaultValue := args[2]

			// Search for an element that starts with the prefix
			for _, elem := range arr.Elements {
				str, ok := elem.(*String)
				if !ok {
					continue
				}
				if strings.HasPrefix(str.Value, prefix.Value) {
					// Return the value after the prefix
					return NewString(strings.TrimPrefix(str.Value, prefix.Value))
				}
			}

			// Not found, return the default value
			return defaultValue
		},
	},
	"switchExists": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for switchExists. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'switchExists' must be ARRAY, got %s", args[0].Type())
			}

			switchName, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'switchExists' must be STRING, got %s", args[1].Type())
			}

			// Search for an exact match only
			for _, elem := range arr.Elements {
				str, ok := elem.(*String)
				if !ok {
					continue
				}
				if str.Value == switchName.Value {
					return TRUE
				}
			}

			return FALSE
		},
	},
	"toJson": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for toJson. got=%d, want>=1", len(args))
			}

			// Parse options
			indent := false
			sortKeys := false

			for i := 1; i < len(args); i++ {
				if str, ok := args[i].(*String); ok {
					switch str.Value {
					case "-indent":
						indent = true
					case "-sort":
						sortKeys = true
					}
				}
			}

			// Convert object to JSON using the exported function
			jsonBytes, err := ObjectToJSON(args[0], ObjectToJSONOptions{
				Indent:   indent,
				SortKeys: sortKeys,
			})
			if err != nil {
				return newError("toJson error: %s", err.Error())
			}

			return NewString(string(jsonBytes))
		},
	},
	"fromJson": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for fromJson. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'fromJson' must be STRING, got %s", args[0].Type())
			}

			// Parse JSON using the exported function
			obj, err := JSONToObject(str.Value)
			if err != nil {
				return newError("fromJson error: %s", err.Error())
			}

			return obj
		},
	},

	// ============================================================
	// Charlang-compatible Array Functions
	// ============================================================
	// append - append value(s) to array (returns new array)
	"append": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for append. got=%d, want>=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'append' must be ARRAY, got %s", args[0].Type())
			}

			// Create new array with appended elements
			newElements := make([]Object, len(arr.Elements)+len(args)-1)
			copy(newElements, arr.Elements)
			copy(newElements[len(arr.Elements):], args[1:])
			return NewArray(newElements)
		},
	},
	// appendArray - merge two arrays (alias for concat, returns new array)
	"appendArray": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for appendArray. got=%d, want=2", len(args))
			}

			arr1, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'appendArray' must be ARRAY, got %s", args[0].Type())
			}
			arr2, ok := args[1].(*Array)
			if !ok {
				return newError("second argument to 'appendArray' must be ARRAY, got %s", args[1].Type())
			}

			newElements := make([]Object, len(arr1.Elements)+len(arr2.Elements))
			copy(newElements, arr1.Elements)
			copy(newElements[len(arr1.Elements):], arr2.Elements)
			return NewArray(newElements)
		},
	},
	// appendList - alias for appendArray
	"appendList": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for appendList. got=%d, want=2", len(args))
			}

			arr1, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'appendList' must be ARRAY, got %s", args[0].Type())
			}
			arr2, ok := args[1].(*Array)
			if !ok {
				return newError("second argument to 'appendList' must be ARRAY, got %s", args[1].Type())
			}

			newElements := make([]Object, len(arr1.Elements)+len(arr2.Elements))
			copy(newElements, arr1.Elements)
			copy(newElements[len(arr1.Elements):], arr2.Elements)
			return NewArray(newElements)
		},
	},
	// appendSlice - alias for appendArray
	"appendSlice": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for appendSlice. got=%d, want=2", len(args))
			}

			arr1, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'appendSlice' must be ARRAY, got %s", args[0].Type())
			}
			arr2, ok := args[1].(*Array)
			if !ok {
				return newError("second argument to 'appendSlice' must be ARRAY, got %s", args[1].Type())
			}

			newElements := make([]Object, len(arr1.Elements)+len(arr2.Elements))
			copy(newElements, arr1.Elements)
			copy(newElements[len(arr1.Elements):], arr2.Elements)
			return NewArray(newElements)
		},
	},
	// arrayContains - check if array contains value
	"arrayContains": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for arrayContains. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'arrayContains' must be ARRAY, got %s", args[0].Type())
			}

			for _, elem := range arr.Elements {
				if deepEquals(elem, args[1]) {
					return TRUE
				}
			}
			return FALSE
		},
	},
	// removeItems - remove items from start to end index
	"removeItems": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for removeItems. got=%d, want=3", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to 'removeItems' must be ARRAY, got %s", args[0].Type())
			}

			start, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'removeItems' must be INT, got %s", args[1].Type())
			}

			end, ok := args[2].(*Int)
			if !ok {
				return newError("third argument to 'removeItems' must be INT, got %s", args[2].Type())
			}

			arrLen := len(arr.Elements)
			startIdx := int(start.Value)
			endIdx := int(end.Value)

			// Handle negative indices
			if startIdx < 0 {
				startIdx = arrLen + startIdx
			}
			if endIdx < 0 {
				endIdx = arrLen + endIdx
			}

			// Bounds check
			if startIdx < 0 {
				startIdx = 0
			}
			if endIdx > arrLen {
				endIdx = arrLen
			}
			if startIdx > endIdx {
				return arr
			}

			// Create new array without the removed items
			newElements := make([]Object, 0, arrLen-(endIdx-startIdx))
			newElements = append(newElements, arr.Elements[:startIdx]...)
			newElements = append(newElements, arr.Elements[endIdx:]...)
			return NewArray(newElements)
		},
	},
	// make - create array or map with specified length and capacity
	"make": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for make. got=%d, want>=2", len(args))
			}

			typeArg, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'make' must be STRING (type name), got %s", args[0].Type())
			}

			switch typeArg.Value {
			case "array", "Array":
				length := int64(0)
				capacity := int64(0)

				if len(args) >= 2 {
					l, ok := args[1].(*Int)
					if !ok {
						return newError("second argument to 'make' must be INT, got %s", args[1].Type())
					}
					length = l.Value
					capacity = length
				}
				if len(args) >= 3 {
					c, ok := args[2].(*Int)
					if !ok {
						return newError("third argument to 'make' must be INT, got %s", args[2].Type())
					}
					capacity = c.Value
				}

				if capacity < length {
					capacity = length
				}

				elements := make([]Object, length, capacity)
				for i := range elements {
					elements[i] = NULL
				}
				return NewArray(elements)

			case "map", "Map":
				capacity := int64(16)
				if len(args) >= 2 {
					c, ok := args[1].(*Int)
					if !ok {
						return newError("second argument to 'make' must be INT, got %s", args[1].Type())
					}
					capacity = c.Value
				}
				pairs := make(map[HashKey]MapPair, capacity)
				return NewMap(pairs)

			default:
				return newError("make: unsupported type '%s'. Use 'array' or 'map'", typeArg.Value)
			}
		},
	},
	// bytes - create byte array from integer arguments
	"bytes": {
		Fn: func(args ...Object) Object {
			elements := make([]Object, len(args))
			for i, arg := range args {
				switch v := arg.(type) {
				case *Int:
					if v.Value < 0 || v.Value > 255 {
						return newError("bytes: value at index %d out of range (0-255), got %d", i, v.Value)
					}
					elements[i] = v
				case *Float:
					val := int64(v.Value)
					if val < 0 || val > 255 {
						return newError("bytes: value at index %d out of range (0-255), got %d", i, val)
					}
					elements[i] = NewInt(val)
				default:
					return newError("bytes: argument at index %d must be INT (0-255), got %s", i, arg.Type())
				}
			}
			return NewArray(elements)
		},
	},
	// plt - pretty table print for debugging
	"plt": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				printPrettyTable(arg, 0)
			}
			return NULL
		},
	},
	// bigInt - create a BigInt from int, float, or string
	"bigInt": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for bigInt. got=%d, want=1", len(args))
			}
			switch v := args[0].(type) {
			case *BigInt:
				return v
			case *Int:
				return NewBigIntFromInt64(v.Value)
			case *Float:
				return NewBigFloatFromFloat64(v.Value).ToBigInt()
			case *BigFloat:
				return v.ToBigInt()
			case *String:
				bigInt, err := NewBigIntFromString(v.Value)
				if err != nil {
					return newError("bigInt: %v", err)
				}
				return bigInt
			default:
				return newError("bigInt: cannot convert %s to BigInt", args[0].Type())
			}
		},
	},
	// bigFloat - create a BigFloat from int, float, bigint, or string
	"bigFloat": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for bigFloat. got=%d, want=1", len(args))
			}
			switch v := args[0].(type) {
			case *BigFloat:
				return v
			case *BigInt:
				return v.ToBigFloat()
			case *Int:
				return NewBigFloatFromInt64(v.Value)
			case *Float:
				return NewBigFloatFromFloat64(v.Value)
			case *String:
				bigFloat, err := NewBigFloatFromString(v.Value)
				if err != nil {
					return newError("bigFloat: %v", err)
				}
				return bigFloat
			default:
				return newError("bigFloat: cannot convert %s to BigFloat", args[0].Type())
			}
		},
	},
	// isBigInt - check if value is a BigInt
	"isBigInt": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isBigInt. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*BigInt)
			if ok {
				return TRUE
			}
			return FALSE
		},
	},
	// isBigFloat - check if value is a BigFloat
	"isBigFloat": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isBigFloat. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*BigFloat)
			if ok {
				return TRUE
			}
			return FALSE
		},
	},

	// ============================================================
	// Concurrency Functions (Tube and Sync)
	// ============================================================

	// makeTube - create a new tube (channel)
	"makeTube": {
		Fn: func(args ...Object) Object {
			elemType := ObjectType("")
			buffer := 0

			argIdx := 0
			if len(args) > 0 {
				// Check if first argument is type string or number
				if str, ok := args[0].(*String); ok {
					elemType = ObjectType(str.Value)
					argIdx = 1
				} else if num, ok := args[0].(*Int); ok {
					buffer = int(num.Value)
					argIdx = 1
				}
			}

			// If there's a second argument, use as buffer
			if len(args) > argIdx {
				if num, ok := args[argIdx].(*Int); ok {
					buffer = int(num.Value)
				}
			}

			return NewTube(elemType, buffer)
		},
	},

	// closeTube - close a tube
	"closeTube": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("closeTube requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			t.Close()
			return NULL
		},
	},

	// tubeLen - get number of elements in tube buffer
	"tubeLen": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("tubeLen requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			return NewInt(int64(t.Len()))
		},
	},

	// tubeCap - get tube buffer capacity
	"tubeCap": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("tubeCap requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			return NewInt(int64(t.Cap()))
		},
	},

	// tubeClosed - check if tube is closed
	"tubeClosed": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("tubeClosed requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			if t.IsClosed() {
				return TRUE
			}
			return FALSE
		},
	},

	// tubeSend - send value to tube (blocking)
	"tubeSend": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("tubeSend requires 2 arguments (tube, value)")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("first argument must be a tube")
			}
			if !t.Send(args[1]) {
				return FALSE
			}
			return TRUE
		},
	},

	// tubeRecv - receive value from tube (blocking), returns [value, ok]
	"tubeRecv": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("tubeRecv requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			val, ok := t.Receive()
			if ok {
				return NewArray([]Object{val, TRUE})
			}
			return NewArray([]Object{val, FALSE})
		},
	},

	// tubeTrySend - try to send without blocking, returns [sent, ok]
	"tubeTrySend": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("tubeTrySend requires 2 arguments (tube, value)")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("first argument must be a tube")
			}
			sent, ok := t.TrySend(args[1])
			var sentBool, okBool *Bool
			if sent {
				sentBool = TRUE
			} else {
				sentBool = FALSE
			}
			if ok {
				okBool = TRUE
			} else {
				okBool = FALSE
			}
			return NewArray([]Object{sentBool, okBool})
		},
	},

	// tubeTryRecv - try to receive without blocking, returns [value, received, open]
	"tubeTryRecv": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("tubeTryRecv requires 1 argument")
			}
			t, ok := args[0].(*Tube)
			if !ok {
				return newError("argument must be a tube")
			}
			val, received, open := t.TryReceive()
			var recvBool, openBool *Bool
			if received {
				recvBool = TRUE
			} else {
				recvBool = FALSE
			}
			if open {
				openBool = TRUE
			} else {
				openBool = FALSE
			}
			return NewArray([]Object{val, recvBool, openBool})
		},
	},

	// newMutex - create a new mutex
	"newMutex": {
		Fn: func(args ...Object) Object {
			return NewMutex()
		},
	},

	// newRWMutex - create a new read-write mutex
	"newRWMutex": {
		Fn: func(args ...Object) Object {
			return NewRWMutex()
		},
	},

	// newWaitGroup - create a new wait group
	"newWaitGroup": {
		Fn: func(args ...Object) Object {
			return NewWaitGroup()
		},
	},

	// newOnce - create a new once
	"newOnce": {
		Fn: func(args ...Object) Object {
			return NewOnce()
		},
	},

	// newCond - create a new condition variable
	"newCond": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("newCond requires 1 argument (mutex)")
			}
			m, ok := args[0].(*Mutex)
			if !ok {
				return newError("argument must be a Mutex")
			}
			return NewCond(m)
		},
	},

	// newAtomic - create a new atomic integer
	"newAtomic": {
		Fn: func(args ...Object) Object {
			initial := int64(0)
			if len(args) > 0 {
				if num, ok := args[0].(*Int); ok {
					initial = num.Value
				}
			}
			return NewAtomicInt(initial)
		},
	},
	// Context builtins for timeout and cancellation
	"newContext": {
		Fn: func(args ...Object) Object {
			// Create a background context
			return NewBackgroundContext()
		},
	},
	"contextWithTimeout": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for contextWithTimeout. got=%d, want>=2", len(args))
			}

			// Get parent context (optional, can be null)
			var parent *Context
			if ctx, ok := args[0].(*Context); ok && ctx != nil {
				parent = ctx
			}

			// Get timeout duration in milliseconds
			timeoutMs, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to contextWithTimeout must be INT (milliseconds), got %s", args[1].Type())
			}

			return NewContextWithTimeout(parent, time.Duration(timeoutMs.Value)*time.Millisecond)
		},
	},
	"contextWithCancel": {
		Fn: func(args ...Object) Object {
			// Get parent context (optional, can be null)
			var parent *Context
			if len(args) > 0 {
				if ctx, ok := args[0].(*Context); ok && ctx != nil {
					parent = ctx
				}
			}

			return NewContextWithCancel(parent)
		},
	},
	"contextWithDeadline": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for contextWithDeadline. got=%d, want>=2", len(args))
			}

			// Get parent context (optional, can be null)
			var parent *Context
			if ctx, ok := args[0].(*Context); ok && ctx != nil {
				parent = ctx
			}

			// Get deadline as Unix timestamp (seconds or milliseconds)
			var deadline time.Time
			switch d := args[1].(type) {
			case *Int:
				// Assume milliseconds if large number, seconds otherwise
				if d.Value > 1e9 {
					deadline = time.Unix(0, d.Value*1e6) // milliseconds to nanoseconds
				} else {
					deadline = time.Unix(d.Value, 0)
				}
			case *Float:
				deadline = time.Unix(int64(d.Value), 0)
			default:
				return newError("second argument to contextWithDeadline must be INT or FLOAT, got %s", args[1].Type())
			}

			return NewContextWithDeadline(parent, deadline)
		},
	},
	"contextCancel": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for contextCancel. got=%d, want>=1", len(args))
			}

			ctx, ok := args[0].(*Context)
			if !ok {
				return newError("argument to contextCancel must be CONTEXT, got %s", args[0].Type())
			}

			ctx.Cancel()
			return NULL
		},
	},
	"contextDone": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for contextDone. got=%d, want>=1", len(args))
			}

			ctx, ok := args[0].(*Context)
			if !ok {
				return newError("argument to contextDone must be CONTEXT, got %s", args[0].Type())
			}

			return ctx.Done()
		},
	},
	"contextErr": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for contextErr. got=%d, want>=1", len(args))
			}

			ctx, ok := args[0].(*Context)
			if !ok {
				return newError("argument to contextErr must be CONTEXT, got %s", args[0].Type())
			}

			errStr := ctx.ErrString()
			if errStr == "" {
				return NULL
			}
			return NewString(errStr)
		},
	},
	"contextIsDone": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for contextIsDone. got=%d, want>=1", len(args))
			}

			ctx, ok := args[0].(*Context)
			if !ok {
				return newError("argument to contextIsDone must be CONTEXT, got %s", args[0].Type())
			}

			if ctx.IsDone() {
				return TRUE
			}
			return FALSE
		},
	},
	"contextDeadline": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for contextDeadline. got=%d, want>=1", len(args))
			}

			ctx, ok := args[0].(*Context)
			if !ok {
				return newError("argument to contextDeadline must be CONTEXT, got %s", args[0].Type())
			}

			dl, hasDeadline := ctx.Deadline()
			if !hasDeadline {
				return NULL
			}
			return NewInt(dl.UnixMilli())
		},
	},

	// ============================================================
	// HTTP Client Functions (getWeb family)
	// ============================================================

	// getWeb - fetch URL content and return as string
	// Usage: getWeb(url) or getWeb(url, "-object") for JSON parsing
	// Returns: string content, or parsed JSON object with "-object" flag
	"getWeb": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for getWeb. got=%d, want>=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'getWeb' must be STRING, got %s", args[0].Type())
			}

			// Parse options
			parseJSON := false
			timeout := 30 * time.Second

			for i := 1; i < len(args); i++ {
				switch opt := args[i].(type) {
				case *String:
					switch opt.Value {
					case "-object", "-json":
						parseJSON = true
					case "-bytes":
						// Handled by getWebBytes
						return newError("use getWebBytes() for byte output")
					}
				case *Int:
					timeout = time.Duration(opt.Value) * time.Second
				}
			}

			// Create HTTP client with timeout
			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(urlObj.Value)
			if err != nil {
				return newError("getWeb request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("getWeb read response failed: %v", err)
			}

			if resp.StatusCode >= 400 {
				return newError("getWeb HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			if parseJSON {
				// Parse JSON response
				var data interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					return newError("getWeb JSON parse failed: %v", err)
				}
				return GoValueToObject(data)
			}

			return NewString(string(body))
		},
	},

	// getWebBytes - fetch URL content and return as byte array
	// Usage: getWebBytes(url)
	// Returns: array of integers (bytes)
	"getWebBytes": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for getWebBytes. got=%d, want>=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'getWebBytes' must be STRING, got %s", args[0].Type())
			}

			// Parse timeout option
			timeout := 30 * time.Second
			if len(args) > 1 {
				if t, ok := args[1].(*Int); ok {
					timeout = time.Duration(t.Value) * time.Second
				}
			}

			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(urlObj.Value)
			if err != nil {
				return newError("getWebBytes request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("getWebBytes read response failed: %v", err)
			}

			if resp.StatusCode >= 400 {
				return newError("getWebBytes HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			// Return as array of integers (bytes)
			elements := make([]Object, len(body))
			for i, b := range body {
				elements[i] = NewInt(int64(b))
			}
			return NewArray(elements)
		},
	},

	// getWebObject - fetch URL content and parse as JSON object
	// Usage: getWebObject(url)
	// Returns: parsed JSON object (map or array)
	"getWebObject": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for getWebObject. got=%d, want>=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'getWebObject' must be STRING, got %s", args[0].Type())
			}

			// Parse timeout option
			timeout := 30 * time.Second
			if len(args) > 1 {
				if t, ok := args[1].(*Int); ok {
					timeout = time.Duration(t.Value) * time.Second
				}
			}

			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(urlObj.Value)
			if err != nil {
				return newError("getWebObject request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("getWebObject read response failed: %v", err)
			}

			if resp.StatusCode >= 400 {
				return newError("getWebObject HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			// Parse JSON response
			var data interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return newError("getWebObject JSON parse failed: %v", err)
			}
			return GoValueToObject(data)
		},
	},

	// postWeb - POST data to URL and return response
	// Usage: postWeb(url, body) or postWeb(url, body, contentType)
	// Returns: string content
	"postWeb": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for postWeb. got=%d, want>=2", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'postWeb' must be STRING, got %s", args[0].Type())
			}

			bodyObj, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'postWeb' must be STRING, got %s", args[1].Type())
			}

			contentType := "application/json"
			if len(args) > 2 {
				if ct, ok := args[2].(*String); ok {
					contentType = ct.Value
				}
			}

			// Parse timeout option
			timeout := 30 * time.Second
			if len(args) > 3 {
				if t, ok := args[3].(*Int); ok {
					timeout = time.Duration(t.Value) * time.Second
				}
			}

			client := &http.Client{Timeout: timeout}
			resp, err := client.Post(urlObj.Value, contentType, strings.NewReader(bodyObj.Value))
			if err != nil {
				return newError("postWeb request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("postWeb read response failed: %v", err)
			}

			return NewString(string(body))
		},
	},

	// postWebObject - POST data to URL and parse JSON response
	// Usage: postWebObject(url, body) or postWebObject(url, body, contentType)
	// Returns: parsed JSON object
	"postWebObject": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for postWebObject. got=%d, want>=2", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'postWebObject' must be STRING, got %s", args[0].Type())
			}

			bodyObj, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'postWebObject' must be STRING, got %s", args[1].Type())
			}

			contentType := "application/json"
			if len(args) > 2 {
				if ct, ok := args[2].(*String); ok {
					contentType = ct.Value
				}
			}

			// Parse timeout option
			timeout := 30 * time.Second
			if len(args) > 3 {
				if t, ok := args[3].(*Int); ok {
					timeout = time.Duration(t.Value) * time.Second
				}
			}

			client := &http.Client{Timeout: timeout}
			resp, err := client.Post(urlObj.Value, contentType, strings.NewReader(bodyObj.Value))
			if err != nil {
				return newError("postWebObject request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("postWebObject read response failed: %v", err)
			}

			// Parse JSON response
			var data interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return newError("postWebObject JSON parse failed: %v", err)
			}
			return GoValueToObject(data)
		},
	},

	// urlExists - check if URL exists (HTTP HEAD request)
	// Usage: urlExists(url)
	// Returns: boolean
	"urlExists": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for urlExists. got=%d, want=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'urlExists' must be STRING, got %s", args[0].Type())
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Head(urlObj.Value)
			if err != nil {
				return FALSE
			}
			defer resp.Body.Close()

			return &Bool{Value: resp.StatusCode < 400}
		},
	},

	// httpStatus - get HTTP status code and headers for a URL
	// Usage: httpStatus(url)
	// Returns: map with statusCode, status, headers
	"httpStatus": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for httpStatus. got=%d, want=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'httpStatus' must be STRING, got %s", args[0].Type())
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Head(urlObj.Value)
			if err != nil {
				return newError("httpStatus request failed: %v", err)
			}
			defer resp.Body.Close()

			// Build result map
			pairs := make(map[HashKey]MapPair)

			pairs[NewString("statusCode").HashKey()] = MapPair{
				Key:   NewString("statusCode"),
				Value: NewInt(int64(resp.StatusCode)),
			}

			pairs[NewString("status").HashKey()] = MapPair{
				Key:   NewString("status"),
				Value: NewString(resp.Status),
			}

			// Build headers map
			headerPairs := make(map[HashKey]MapPair)
			for k, v := range resp.Header {
				headerPairs[NewString(k).HashKey()] = MapPair{
					Key:   NewString(k),
					Value: NewString(strings.Join(v, ", ")),
				}
			}
			pairs[NewString("headers").HashKey()] = MapPair{
				Key:   NewString("headers"),
				Value: NewMap(headerPairs),
			}

			return NewMap(pairs)
		},
	},

	// downloadFile - download a file from URL to local path
	// Usage: downloadFile(url, localPath) or downloadFile(url, localPath, timeoutSeconds)
	// Returns: null on success, error on failure
	"downloadFile": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments for downloadFile. got=%d, want>=2", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'downloadFile' must be STRING (URL), got %s", args[0].Type())
			}

			pathObj, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'downloadFile' must be STRING (local path), got %s", args[1].Type())
			}

			// Parse timeout option
			timeout := 60 * time.Second
			if len(args) > 2 {
				if timeoutOpt, ok := args[2].(*Int); ok {
					timeout = time.Duration(timeoutOpt.Value) * time.Second
				}
			}

			// Create HTTP client with timeout
			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(urlObj.Value)
			if err != nil {
				return newError("downloadFile request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				return newError("downloadFile HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			// Create the local file
			out, err := os.Create(pathObj.Value)
			if err != nil {
				return newError("downloadFile create file failed: %v", err)
			}
			defer out.Close()

			// Copy the response body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return newError("downloadFile write failed: %v", err)
			}

			return NULL
		},
	},

	// ============================================================
	// Reader/Writer Functions
	// ============================================================

	// getWebReader - fetch URL and return a Reader for streaming
	// Usage: getWebReader(url)
	// Returns: Reader object
	"getWebReader": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments for getWebReader. got=%d, want>=1", len(args))
			}

			urlObj, ok := args[0].(*String)
			if !ok {
				return newError("first argument to 'getWebReader' must be STRING, got %s", args[0].Type())
			}

			// Parse timeout option
			timeout := 30 * time.Second
			if len(args) > 1 {
				if t, ok := args[1].(*Int); ok {
					timeout = time.Duration(t.Value) * time.Second
				}
			}

			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(urlObj.Value)
			if err != nil {
				return newError("getWebReader request failed: %v", err)
			}

			if resp.StatusCode >= 400 {
				resp.Body.Close()
				return newError("getWebReader HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			return NewReader(resp.Body)
		},
	},

	// ioCopy - copy data from reader to writer
	// Usage: ioCopy(writer, reader)
	// Returns: number of bytes copied
	"ioCopy": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for ioCopy. got=%d, want=2", len(args))
			}

			writer, ok := args[0].(*Writer)
			if !ok {
				return newError("first argument to 'ioCopy' must be WRITER, got %s", args[0].Type())
			}

			reader, ok := args[1].(*Reader)
			if !ok {
				return newError("second argument to 'ioCopy' must be READER, got %s", args[1].Type())
			}

			return IoCopy(writer, reader)
		},
	},

	// isReader - check if value is a Reader
	"isReader": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isReader. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Reader)
			return &Bool{Value: ok}
		},
	},

	// isWriter - check if value is a Writer
	"isWriter": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isWriter. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*Writer)
			return &Bool{Value: ok}
		},
	},

	// newBytesReader - create a Reader from byte array
	// Usage: newBytesReader(bytes)
	"newBytesReader": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for newBytesReader. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to 'newBytesReader' must be ARRAY, got %s", args[0].Type())
			}

			data := make([]byte, len(arr.Elements))
			for i, elem := range arr.Elements {
				b, ok := elem.(*Int)
				if !ok {
					return newError("array elements must be integers (0-255)")
				}
				if b.Value < 0 || b.Value > 255 {
					return newError("byte value out of range (0-255)")
				}
				data[i] = byte(b.Value)
			}

			return NewReader(strings.NewReader(string(data)))
		},
	},

	// newStringReader - create a Reader from string
	// Usage: newStringReader(str)
	"newStringReader": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for newStringReader. got=%d, want=1", len(args))
			}

			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to 'newStringReader' must be STRING, got %s", args[0].Type())
			}

			return NewReader(strings.NewReader(s.Value))
		},
	},

	// ============================================================
	// Encryption Functions (Charlang compatible)
	// ============================================================

	// encryptTextByTXTE - simple text encryption
	"encryptTextByTXTE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptTextByTXTE requires at least 1 argument")
			}
			text := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(encryptStringByTXTE(text, code))
		},
	},
	"decryptTextByTXTE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptTextByTXTE requires at least 1 argument")
			}
			hexStr := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(decryptStringByTXTE(hexStr, code))
		},
	},

	// encryptDataByTXDEE - enhanced data encryption
	"encryptDataByTXDEE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptDataByTXDEE requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := encryptDataByTXDEE(data, code)
			if result == nil {
				return newError("encryption failed")
			}
			return NewBytes(result)
		},
	},
	"decryptDataByTXDEE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptDataByTXDEE requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := decryptDataByTXDEE(data, code)
			if result == nil {
				return newError("decryption failed")
			}
			return NewBytes(result)
		},
	},
	"encryptTextByTXDEE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptTextByTXDEE requires at least 1 argument")
			}
			text := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(encryptStringByTXDEE(text, code))
		},
	},
	"decryptTextByTXDEE": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptTextByTXDEE requires at least 1 argument")
			}
			hexStr := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(decryptStringByTXDEE(hexStr, code))
		},
	},

	// encryptData/encryptBytes/decryptData/decryptBytes - TXDEF encryption
	"encryptData": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptData requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := encryptDataByTXDEF(data, code)
			if result == nil {
				return newError("encryption failed")
			}
			return NewBytes(result)
		},
	},
	"encryptBytes": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptBytes requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := encryptDataByTXDEF(data, code)
			if result == nil {
				return newError("encryption failed")
			}
			return NewBytes(result)
		},
	},
	"decryptData": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptData requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := decryptDataByTXDEF(data, code)
			if result == nil {
				return newError("decryption failed")
			}
			return NewBytes(result)
		},
	},
	"decryptBytes": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptBytes requires at least 1 argument")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			result := decryptDataByTXDEF(data, code)
			if result == nil {
				return newError("decryption failed")
			}
			return NewBytes(result)
		},
	},

	// encryptText/encryptStr/decryptText/decryptStr - TXDEF text encryption
	"encryptText": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptText requires at least 1 argument")
			}
			text := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(encryptStringByTXDEF(text, code))
		},
	},
	"encryptStr": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("encryptStr requires at least 1 argument")
			}
			text := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(encryptStringByTXDEF(text, code))
		},
	},
	"decryptText": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptText requires at least 1 argument")
			}
			hexStr := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(decryptStringByTXDEF(hexStr, code))
		},
	},
	"decryptStr": {
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("decryptStr requires at least 1 argument")
			}
			hexStr := args[0].Inspect()
			code := ""
			if len(args) > 1 {
				code = args[1].Inspect()
			}
			return NewString(decryptStringByTXDEF(hexStr, code))
		},
	},

	// encryptStream/decryptStream - stream encryption
	"encryptStream": {
		Fn: func(args ...Object) Object {
			if len(args) < 3 {
				return newError("encryptStream requires 3 arguments: reader, code, writer")
			}
			reader, ok := args[0].(io.Reader)
			if !ok {
				return newError("first argument must be a reader")
			}
			code := args[1].Inspect()
			writer, ok := args[2].(io.Writer)
			if !ok {
				return newError("third argument must be a writer")
			}
			err := encryptStreamByTXDEF(reader, code, writer)
			if err != nil {
				return newError("encryption failed: %v", err)
			}
			return NULL
		},
	},
	"decryptStream": {
		Fn: func(args ...Object) Object {
			if len(args) < 3 {
				return newError("decryptStream requires 3 arguments: reader, code, writer")
			}
			reader, ok := args[0].(io.Reader)
			if !ok {
				return newError("first argument must be a reader")
			}
			code := args[1].Inspect()
			writer, ok := args[2].(io.Writer)
			if !ok {
				return newError("third argument must be a writer")
			}
			err := decryptStreamByTXDEF(reader, code, writer)
			if err != nil {
				return newError("decryption failed: %v", err)
			}
			return NULL
		},
	},

	// aesEncrypt/aesDecrypt - AES encryption
	"aesEncrypt": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("aesEncrypt requires at least 2 arguments: data, key")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			key := []byte(args[1].Inspect())

			mode := ""
			if len(args) > 2 {
				mode = args[2].Inspect()
			}

			var result []byte
			var err error

			if strings.Contains(mode, "cbc") || strings.Contains(mode, "-cbc") {
				result, err = aesEncryptCBC(data, key)
			} else {
				result, err = aesEncryptECB(data, key)
			}

			if err != nil {
				return newError("AES encryption failed: %v", err)
			}
			return NewString(string(result))
		},
	},
	"aesDecrypt": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("aesDecrypt requires at least 2 arguments: data, key")
			}
			var data []byte
			switch v := args[0].(type) {
			case *Bytes:
				data = v.Value
			default:
				data = []byte(args[0].Inspect())
			}
			key := []byte(args[1].Inspect())

			mode := ""
			if len(args) > 2 {
				mode = args[2].Inspect()
			}

			var result []byte
			var err error

			if strings.Contains(mode, "cbc") || strings.Contains(mode, "-cbc") {
				result, err = aesDecryptCBC(data, key)
			} else {
				result, err = aesDecryptECB(data, key)
			}

			if err != nil {
				return newError("AES decryption failed: %v", err)
			}
			return NewString(string(result))
		},
	},

	// ============================================================
	// Database Builtin Functions
	// ============================================================
	// Two versions are provided:
	// 1. String-based (Charlang compatible): dbQuery, dbQueryRecs, dbQueryMap, etc.
	// 2. Typed (preserve native types): dbQueryTyped, dbQueryRowTyped, etc.

	// String-based functions (Charlang compatible)
	"formatSQLValue":  BuiltinFormatSQLValue,
	"dbConnect":       BuiltinDbConnect,
	"dbClose":         BuiltinDbClose,
	"dbQuery":         BuiltinDbQuery,
	"dbQueryOrdered":  BuiltinDbQueryOrdered,
	"dbQueryRecs":     BuiltinDbQueryRecs,
	"dbQueryMap":      BuiltinDbQueryMap,
	"dbQueryMapArray": BuiltinDbQueryMapArray,
	"dbQueryCount":    BuiltinDbQueryCount,
	"dbQueryFloat":    BuiltinDbQueryFloat,
	"dbQueryString":   BuiltinDbQueryString,
	"dbExec":          BuiltinDbExec,

	// Typed functions (preserve native data types)
	"dbQueryTyped":      BuiltinDbQueryTyped,
	"dbQueryRowTyped":   BuiltinDbQueryRowTyped,
	"dbQueryArrayTyped": BuiltinDbQueryArrayTyped,
	"dbQueryValueTyped": BuiltinDbQueryValueTyped,
}

// RunCodeImpl is the implementation function for runCode, set by the VM
var runCodeImpl func(code string, args *Map) (Object, error)

// SetRunCodeImpl registers the runCode implementation and returns the previous value
func SetRunCodeImpl(fn func(code string, args *Map) (Object, error)) func(code string, args *Map) (Object, error) {
	prev := runCodeImpl
	runCodeImpl = fn
	return prev
}

// LoadPluginImpl is the implementation function for loadPlugin, set by the VM
var loadPluginImpl func(path string) (Object, error)

// SetLoadPluginImpl registers the loadPlugin implementation and returns the previous value
func SetLoadPluginImpl(fn func(path string) (Object, error)) func(path string) (Object, error) {
	prev := loadPluginImpl
	loadPluginImpl = fn
	return prev
}

// CallUserFuncImpl is the implementation function for calling user-defined functions from builtins
// The callback receives the function object and arguments, and returns the result
var callUserFuncImpl func(fn Object, args ...Object) (Object, error)

// SetCallUserFuncImpl registers the callback for calling user functions and returns the previous value
func SetCallUserFuncImpl(fn func(fnObj Object, args ...Object) (Object, error)) func(Object, ...Object) (Object, error) {
	prev := callUserFuncImpl
	callUserFuncImpl = fn
	return prev
}

// CallUserFunc calls a user-defined function from within a builtin method
// Returns an error if the callback is not set or if the call fails
func CallUserFunc(fn Object, args ...Object) (Object, error) {
	if callUserFuncImpl == nil {
		return nil, fmt.Errorf("user function callback not available in this context")
	}
	return callUserFuncImpl(fn, args...)
}

// DelegateImpl is the implementation function for delegate, set by the VM
// It compiles source code and returns a callable closure
var delegateImpl func(source string) (Object, error)

// SetDelegateImpl registers the delegate implementation and returns the previous value
func SetDelegateImpl(fn func(source string) (Object, error)) func(source string) (Object, error) {
	prev := delegateImpl
	delegateImpl = fn
	return prev
}

// newError creates a new Error object with the given message
func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

// compareObjects compares two objects for equality
func compareObjects(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	return a.Inspect() == b.Inspect()
}

// Random number generator (thread-safe)
var (
	randMu   sync.Mutex
	randSrc  *rand.Rand
	randOnce sync.Once
)

// initRand initializes the random source once
func initRand() {
	randOnce.Do(func() {
		randSrc = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

// randInt63 returns a random int64 in a thread-safe manner
func randInt63() int64 {
	initRand()
	randMu.Lock()
	defer randMu.Unlock()
	return randSrc.Int63()
}

// flattenArray recursively flattens an array to the specified depth
func flattenArray(elements []Object, depth int64) []Object {
	result := make([]Object, 0)

	for _, elem := range elements {
		if arr, ok := elem.(*Array); ok && depth != 0 {
			newDepth := depth
			if depth > 0 {
				newDepth = depth - 1
			}
			result = append(result, flattenArray(arr.Elements, newDepth)...)
		} else {
			result = append(result, elem)
		}
	}

	return result
}

// deepCopyObject creates a deep copy of an object
func deepCopyObject(obj Object) Object {
	switch o := obj.(type) {
	case *Int:
		return NewInt(o.Value)
	case *Float:
		return NewFloat(o.Value)
	case *String:
		return NewString(o.Value)
	case *Bool:
		if o.Value {
			return TRUE
		}
		return FALSE
	case *Null:
		return NULL
	case *Array:
		elements := make([]Object, len(o.Elements))
		for i, elem := range o.Elements {
			elements[i] = deepCopyObject(elem)
		}
		return NewArray(elements)
	case *Map:
		pairs := make(map[HashKey]MapPair)
		for k, v := range o.Pairs {
			pairs[k] = MapPair{
				Key:   deepCopyObject(v.Key),
				Value: deepCopyObject(v.Value),
			}
		}
		return NewMap(pairs)
	default:
		// For other types, return as-is (functions, builtins, etc.)
		return obj
	}
}

// shallowCopyObject creates a shallow copy of an object
func shallowCopyObject(obj Object) Object {
	switch o := obj.(type) {
	case *Array:
		elements := make([]Object, len(o.Elements))
		copy(elements, o.Elements)
		return NewArray(elements)
	case *Map:
		pairs := make(map[HashKey]MapPair)
		for k, v := range o.Pairs {
			pairs[k] = v
		}
		return NewMap(pairs)
	default:
		// For primitive types, return as-is
		return obj
	}
}

// deepEquals compares two objects for deep equality
func deepEquals(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch aObj := a.(type) {
	case *Int:
		bObj, ok := b.(*Int)
		return ok && aObj.Value == bObj.Value
	case *Float:
		bObj, ok := b.(*Float)
		return ok && aObj.Value == bObj.Value
	case *String:
		bObj, ok := b.(*String)
		return ok && aObj.Value == bObj.Value
	case *Bool:
		bObj, ok := b.(*Bool)
		return ok && aObj.Value == bObj.Value
	case *Null:
		return true
	case *Array:
		bObj, ok := b.(*Array)
		if !ok || len(aObj.Elements) != len(bObj.Elements) {
			return false
		}
		for i := range aObj.Elements {
			if !deepEquals(aObj.Elements[i], bObj.Elements[i]) {
				return false
			}
		}
		return true
	case *Map:
		bObj, ok := b.(*Map)
		if !ok || len(aObj.Pairs) != len(bObj.Pairs) {
			return false
		}
		for k, v := range aObj.Pairs {
			bVal, exists := bObj.Pairs[k]
			if !exists || !deepEquals(v.Value, bVal.Value) {
				return false
			}
		}
		return true
	default:
		return a.Inspect() == b.Inspect()
	}
}

// printPrettyTable prints an object as a formatted table
func printPrettyTable(obj Object, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch v := obj.(type) {
	case *Array:
		if len(v.Elements) == 0 {
			fmt.Printf("%s[]\n", indentStr)
			return
		}
		fmt.Printf("%s[\n", indentStr)
		for i, elem := range v.Elements {
			fmt.Printf("%s  [%d]: ", indentStr, i)
			switch e := elem.(type) {
			case *Array:
				fmt.Println()
				printPrettyTable(e, indent+2)
			case *Map:
				fmt.Println()
				printPrettyTable(e, indent+2)
			default:
				fmt.Println(e.Inspect())
			}
		}
		fmt.Printf("%s]\n", indentStr)

	case *Map:
		if len(v.Pairs) == 0 {
			fmt.Printf("%s{}\n", indentStr)
			return
		}
		fmt.Printf("%s{\n", indentStr)
		// Sort keys for consistent output
		keys := make([]string, 0, len(v.Pairs))
		for _, pair := range v.Pairs {
			if ks, ok := pair.Key.(*String); ok {
				keys = append(keys, ks.Value)
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, pair := range v.Pairs {
				if ks, ok := pair.Key.(*String); ok && ks.Value == key {
					fmt.Printf("%s  %s: ", indentStr, key)
					switch val := pair.Value.(type) {
					case *Array:
						fmt.Println()
						printPrettyTable(val, indent+2)
					case *Map:
						fmt.Println()
						printPrettyTable(val, indent+2)
					default:
						fmt.Println(val.Inspect())
					}
					break
				}
			}
		}
		fmt.Printf("%s}\n", indentStr)

	default:
		fmt.Printf("%s%s\n", indentStr, obj.Inspect())
	}
}
