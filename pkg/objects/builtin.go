// pkg/objects/builtin.go
package objects

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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
