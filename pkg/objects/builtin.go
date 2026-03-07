// pkg/objects/builtin.go
package objects

import "fmt"

// BuiltinFunction is the type for built-in functions
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinType }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) ToBool() *Bool    { return TRUE }
func (b *Builtin) HashKey() HashKey { return HashKey{Type: BuiltinType, Value: 0} }

// Builtins contains all built-in functions
var Builtins = map[string]*Builtin{
	"len": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for len. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return &Int{Value: int64(len(arg.Value))}
			case *Array:
				return &Int{Value: int64(len(arg.Elements))}
			case *Map:
				return &Int{Value: int64(len(arg.Pairs))}
			default:
				return newError("argument to 'len' not supported, got %s", args[0].Type())
			}
		},
	},
	"print": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},
	"println": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},
	"typeOf": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for typeOf. got=%d, want=1", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	},
}

// newError creates a new Error object with the given message
func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
