// pkg/objects/builtin.go
package objects

import (
	"fmt"
	"math"
	"math/rand"
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

    // ============================================================
    // String Functions
    // ============================================================
    "substr": {
        Fn: func(args ...Object) Object {
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

            return &String{Value: str.Value[start.Value:end]}
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
                elements[i] = &String{Value: part}
            }

            return &Array{Elements: elements}
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

            return &String{Value: strings.Join(parts, sep.Value)}
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

            return &String{Value: strings.TrimSpace(str.Value)}
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

            return &String{Value: strings.ToUpper(str.Value)}
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

            return &String{Value: strings.ToLower(str.Value)}
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

            return &String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
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
                    return &Int{Value: -arg.Value}
                }
                return arg
            case *Float:
                if arg.Value < 0 {
                    return &Float{Value: -arg.Value}
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

            return &Int{Value: int64(math.Floor(val))}
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

            return &Int{Value: int64(math.Ceil(val))}
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

            return &Float{Value: math.Sqrt(val)}
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

            return &Float{Value: math.Pow(base, exp)}
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
                return &Int{Value: int64(arg.Value)}
            case *String:
                val, err := strconv.ParseInt(arg.Value, 10, 64)
                if err != nil {
                    return newError("could not convert string to int: %s", arg.Value)
                }
                return &Int{Value: val}
            case *Bool:
                if arg.Value {
                    return &Int{Value: 1}
                }
                return &Int{Value: 0}
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
                return &Float{Value: float64(arg.Value)}
            case *Float:
                return arg
            case *String:
                val, err := strconv.ParseFloat(arg.Value, 64)
                if err != nil {
                    return newError("could not convert string to float: %s", arg.Value)
                }
                return &Float{Value: val}
            case *Bool:
                if arg.Value {
                    return &Float{Value: 1.0}
                }
                return &Float{Value: 0.0}
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
            return &String{Value: args[0].Inspect()}
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
            newElements := make([]Object, len(arr.Elements)+1)
            copy(newElements, arr.Elements)
            newElements[len(arr.Elements)] = args[1]
            return &Array{Elements: newElements}
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
            lastElem := arr.Elements[len(arr.Elements)-1]
            newElements := make([]Object, len(arr.Elements)-1)
            copy(newElements, arr.Elements[:len(arr.Elements)-1])
            return &Array{Elements: newElements, LastPopped: lastElem}
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
            return &Array{Elements: arr.Elements[start.Value:end]}
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
            return &Array{Elements: newElements}
        },
    },
    "indexOf": {
        Fn: func(args ...Object) Object {
            if len(args) != 2 {
                return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
            }

            arr, ok := args[0].(*Array)
            if !ok {
                return newError("first argument to 'indexOf' must be ARRAY, got %s", args[0].Type())
            }

            for i, elem := range arr.Elements {
                if compareObjects(elem, args[1]) {
                    return &Int{Value: int64(i)}
                }
            }

            return &Int{Value: -1}
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
            return &Array{Elements: keys}
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
            vals := make([]Object, len(m.Pairs))
            i := 0
            for _, pair := range m.Pairs {
                vals[i] = pair.Value
                i++
            }
            return &Array{Elements: vals}
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
            newPairs := make(map[HashKey]MapPair, len(m.Pairs)-1)
            for k, v := range m.Pairs {
                if k != args[1].HashKey() {
                    newPairs[k] = v
                }
            }
            return &Map{Pairs: newPairs}
        },
    },

    // ============================================================
    // Utility Functions
    // ============================================================
    "range": {
        Fn: func(args ...Object) Object {
            if len(args) < 1 || len(args) > 2 {
                return newError("wrong number of arguments for range. got=%d, want=1 or 2", len(args))
            }
            var start, end int64
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
            }
            elements := make([]Object, 0)
            if start <= end {
                elements = make([]Object, end-start+1)
                for i := start; i <= end; i++ {
                    elements[i-start] = &Int{Value: i}
                }
            } else {
                elements = make([]Object, start-end+1)
                for i := start; i >= end; i-- {
                    elements[start-i] = &Int{Value: i}
                }
            }
            return &Array{Elements: elements}
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
            return &Array{Elements: sorted}
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
                return &Int{Value: 0}
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
                return &Float{Value: float64(total)}
            }
            return &Int{Value: total}
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
                return &Float{Value: 0}
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
            return &Float{Value: total / float64(len(arr.Elements))}
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
            return &Array{Elements: reversed}
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

            return &String{Value: strings.Repeat(str.Value, int(count.Value))}
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
            return &String{Value: padding[:padLen] + str.Value}
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
            return &String{Value: str.Value + padding[:padLen]}
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

            return &String{Value: string(str.Value[idx])}
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

            return &String{Value: strings.TrimLeft(str.Value, cutset)}
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

            return &String{Value: strings.TrimRight(str.Value, cutset)}
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
                return &Int{Value: int64(math.Round(val))}
            }

            multiplier := math.Pow(10, float64(precision))
            result := math.Round(val*multiplier) / multiplier
            return &Float{Value: result}
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
                return &Int{Value: int64(result)}
            }
            return &Float{Value: result}
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
                    return &Int{Value: 1}
                } else if arg.Value < 0 {
                    return &Int{Value: -1}
                }
                return &Int{Value: 0}
            case *Float:
                if arg.Value > 0 {
                    return &Int{Value: 1}
                } else if arg.Value < 0 {
                    return &Int{Value: -1}
                }
                return &Int{Value: 0}
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
            return &Float{Value: float64(randInt63()) / float64(1<<63)}
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
            return &Int{Value: result}
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

            return &Array{Elements: result}
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

            depth := int64(-1) // -1 means infinite depth
            if len(args) == 2 {
                d, ok := args[1].(*Int)
                if !ok {
                    return newError("second argument to 'flatten' must be INT, got %s", args[1].Type())
                }
                depth = d.Value
            }

            result := flattenArray(arr.Elements, depth)
            return &Array{Elements: result}
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

            return &Array{Elements: result}
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

            return &Array{Elements: arr.Elements[:count]}
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

            return &Array{Elements: arr.Elements[count:]}
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

            return &Map{Pairs: result}
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
                result = append(result, &Array{Elements: []Object{pair.Key, pair.Value}})
            }

            return &Array{Elements: result}
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

            return &String{Value: fmt.Sprintf(format.Value, formatArgs...)}
        },
    },
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
