// pkg/objects/methods.go
package objects

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// TypeMethods maps ObjectType -> methodName -> *Builtin
var TypeMethods = map[ObjectType]map[string]*Builtin{
	IntType:           intMethods,
	FloatType:         floatMethods,
	BigIntType:        bigIntMethods,
	BigFloatType:      bigFloatMethods,
	StringType:        stringMethods,
	CharsType:         charsMethods,
	ArrayType:         arrayMethods,
	MapType:           mapMethods,
	BoolType:          boolMethods,
	NullType:          nullMethods,
	StringBuilderType: stringBuilderMethods,
	WebSocketType:     webSocketMethods,
	// Concurrency types
	MutexType:      mutexMethods,
	RWMutexType:    rwMutexMethods,
	WaitGroupType:  waitGroupMethods,
	AtomicIntType:  atomicIntMethods,
	TubeType:       tubeMethods,
	OnceType:       onceMethods,
	CondType:       condMethods,
	ContextType:    contextMethods,
}

// GetMethod returns the builtin method for the given object type and method name
func GetMethod(objType ObjectType, name string) (*Builtin, bool) {
	methods, ok := TypeMethods[objType]
	if !ok {
		return nil, false
	}
	method, ok := methods[name]
	return method, ok
}

// ============================================================
// Universal Methods (available on all types)
// ============================================================

// universalTypeOf returns the type of any object
func universalTypeOf(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for typeOf. got=%d, want=1", len(args))
	}
	return NewString(string(args[0].Type()))
}

// universalToStr returns the string representation of any object
func universalToStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toStr. got=%d, want=1", len(args))
	}
	return NewString(args[0].Inspect())
}

// ============================================================
// Int Methods
// ============================================================

var intMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Int)
		if !ok {
			return newError("receiver for toFloat must be INT, got %s", args[0].Type())
		}
		return NewFloat(float64(self.Value))
	}},
	"toBigInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toBigInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Int)
		if !ok {
			return newError("receiver for toBigInt must be INT, got %s", args[0].Type())
		}
		return NewBigIntFromInt64(self.Value)
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Int)
		if !ok {
			return newError("receiver for abs must be INT, got %s", args[0].Type())
		}
		if self.Value < 0 {
			return NewInt(-self.Value)
		}
		return self
	}},
}

// ============================================================
// Float Methods
// ============================================================

var floatMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for toInt must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(self.Value))
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for abs must be FLOAT, got %s", args[0].Type())
		}
		if self.Value < 0 {
			return NewFloat(-self.Value)
		}
		return self
	}},
	"floor": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for floor. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for floor must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Floor(self.Value)))
	}},
	"ceil": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for ceil. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for ceil must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Ceil(self.Value)))
	}},
	"round": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for round. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for round must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Round(self.Value)))
	}},
}

// ============================================================
// BigInt Methods
// ============================================================

var bigIntMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for toInt must be BIGINT, got %s", args[0].Type())
		}
		n, ok := self.ToInt64()
		if !ok {
			return newError("BigInt value overflow for int64")
		}
		return NewInt(n)
	}},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for toFloat must be BIGINT, got %s", args[0].Type())
		}
		return NewFloat(self.ToFloat64())
	}},
	"toBigFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toBigFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for toBigFloat must be BIGINT, got %s", args[0].Type())
		}
		return self.ToBigFloat()
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for abs must be BIGINT, got %s", args[0].Type())
		}
		return self.Abs()
	}},
	"neg": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for neg. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for neg must be BIGINT, got %s", args[0].Type())
		}
		return self.Neg()
	}},
	"bits": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for bits. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for bits must be BIGINT, got %s", args[0].Type())
		}
		return NewInt(int64(self.BitLen()))
	}},
	"sign": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sign. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigInt)
		if !ok {
			return newError("receiver for sign must be BIGINT, got %s", args[0].Type())
		}
		return NewInt(int64(self.Sign()))
	}},
}

// ============================================================
// BigFloat Methods
// ============================================================

var bigFloatMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for toInt must be BIGFLOAT, got %s", args[0].Type())
		}
		n, ok := self.ToInt64()
		if !ok {
			return newError("BigFloat value overflow for int64")
		}
		return NewInt(n)
	}},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for toFloat must be BIGFLOAT, got %s", args[0].Type())
		}
		f, _ := self.ToFloat64()
		return NewFloat(f)
	}},
	"toBigInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toBigInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for toBigInt must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.ToBigInt()
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for abs must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.Abs()
	}},
	"neg": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for neg. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for neg must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.Neg()
	}},
	"floor": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for floor. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for floor must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.Floor()
	}},
	"ceil": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for ceil. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for ceil must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.Ceil()
	}},
	"round": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for round. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for round must be BIGFLOAT, got %s", args[0].Type())
		}
		return self.Round()
	}},
	"precision": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for precision. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for precision must be BIGFLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(self.Precision()))
	}},
	"sign": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sign. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BigFloat)
		if !ok {
			return newError("receiver for sign must be BIGFLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(self.Sign()))
	}},
}

// ============================================================
// String Methods
// ============================================================

var stringMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for len must be STRING, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Value)))
	}},
	"upper": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for upper. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for upper must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.ToUpper(self.Value))
	}},
	"lower": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lower. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for lower must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.ToLower(self.Value))
	}},
	"trim": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trim. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trim must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.TrimSpace(self.Value))
	}},
	"trimLeft": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for trimLeft. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trimLeft must be STRING, got %s", args[0].Type())
		}
		cutset := " \t\n\r"
		if len(args) == 2 {
			cs, ok := args[1].(*String)
			if !ok {
				return newError("argument for trimLeft must be STRING, got %s", args[1].Type())
			}
			cutset = cs.Value
		}
		return NewString(strings.TrimLeft(self.Value, cutset))
	}},
	"trimRight": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for trimRight. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trimRight must be STRING, got %s", args[0].Type())
		}
		cutset := " \t\n\r"
		if len(args) == 2 {
			cs, ok := args[1].(*String)
			if !ok {
				return newError("argument for trimRight must be STRING, got %s", args[1].Type())
			}
			cutset = cs.Value
		}
		return NewString(strings.TrimRight(self.Value, cutset))
	}},
	"split": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for split. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for split must be STRING, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("separator for split must be STRING, got %s", args[1].Type())
		}
		parts := strings.Split(self.Value, sep.Value)
		elements := make([]Object, len(parts))
		for i, part := range parts {
			elements[i] = NewString(part)
		}
		return NewArray(elements)
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for contains must be STRING, got %s", args[0].Type())
		}
		substr, ok := args[1].(*String)
		if !ok {
			return newError("argument for contains must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.Contains(self.Value, substr.Value)}
	}},
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for indexOf must be STRING, got %s", args[0].Type())
		}
		substr, ok := args[1].(*String)
		if !ok {
			return newError("argument for indexOf must be STRING, got %s", args[1].Type())
		}
		return NewInt(int64(strings.Index(self.Value, substr.Value)))
	}},
	"startsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for startsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for startsWith must be STRING, got %s", args[0].Type())
		}
		prefix, ok := args[1].(*String)
		if !ok {
			return newError("argument for startsWith must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasPrefix(self.Value, prefix.Value)}
	}},
	"endsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for endsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for endsWith must be STRING, got %s", args[0].Type())
		}
		suffix, ok := args[1].(*String)
		if !ok {
			return newError("argument for endsWith must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasSuffix(self.Value, suffix.Value)}
	}},
	"subStr": {Fn: func(args ...Object) Object {
		// subStr uses BYTE indices (Go-compatible behavior)
		// For character-based slicing, use toChars().subStr() instead
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for subStr. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for subStr must be STRING, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("start index for subStr must be INT, got %s", args[1].Type())
		}
		byteLen := len(self.Value)
		startIdx := int(start.Value)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > byteLen {
			startIdx = byteLen
		}
		if len(args) == 3 {
			end, ok := args[2].(*Int)
			if !ok {
				return newError("end index for subStr must be INT, got %s", args[2].Type())
			}
			endIdx := int(end.Value)
			if endIdx < startIdx {
				endIdx = startIdx
			}
			if endIdx > byteLen {
				endIdx = byteLen
			}
			return NewString(self.Value[startIdx:endIdx])
		}
		return NewString(self.Value[startIdx:])
	}},
	"charLen": {Fn: func(args ...Object) Object {
		// charLen returns the number of Unicode characters (runes)
		// Use this instead of len() when working with Unicode text
		if len(args) != 1 {
			return newError("wrong number of arguments for charLen. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for charLen must be STRING, got %s", args[0].Type())
		}
		return NewInt(int64(len([]rune(self.Value))))
	}},
	"toChars": {Fn: func(args ...Object) Object {
		// toChars converts string to chars ([]rune) for character-based operations
		if len(args) != 1 {
			return newError("wrong number of arguments for toChars. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for toChars must be STRING, got %s", args[0].Type())
		}
		return NewCharsFromString(self.Value)
	}},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for toInt must be STRING, got %s", args[0].Type())
		}
		val, err := strconv.ParseInt(self.Value, 10, 64)
		if err != nil {
			return newError("could not convert string to int: %s", self.Value)
		}
		return NewInt(val)
	}},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for toFloat must be STRING, got %s", args[0].Type())
		}
		val, err := strconv.ParseFloat(self.Value, 64)
		if err != nil {
			return newError("could not convert string to float: %s", self.Value)
		}
		return NewFloat(val)
	}},
}

// ============================================================
// Chars Methods ([]rune-like operations for Unicode character handling)
// ============================================================

var charsMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for len must be CHARS, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Value)))
	}},
	"upper": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for upper. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for upper must be CHARS, got %s", args[0].Type())
		}
		return NewCharsFromString(strings.ToUpper(string(self.Value)))
	}},
	"lower": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lower. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for lower must be CHARS, got %s", args[0].Type())
		}
		return NewCharsFromString(strings.ToLower(string(self.Value)))
	}},
	"trim": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trim. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for trim must be CHARS, got %s", args[0].Type())
		}
		return NewCharsFromString(strings.TrimSpace(string(self.Value)))
	}},
	"split": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for split. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for split must be CHARS, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			sepChars, ok := args[1].(*Chars)
			if !ok {
				return newError("separator for split must be STRING or CHARS, got %s", args[1].Type())
			}
			sep = NewString(string(sepChars.Value))
		}
		parts := strings.Split(string(self.Value), sep.Value)
		elements := make([]Object, len(parts))
		for i, part := range parts {
			elements[i] = NewCharsFromString(part)
		}
		return NewArray(elements)
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for contains must be CHARS, got %s", args[0].Type())
		}
		var substr string
		switch s := args[1].(type) {
		case *String:
			substr = s.Value
		case *Chars:
			substr = string(s.Value)
		default:
			return newError("argument for contains must be STRING or CHARS, got %s", args[1].Type())
		}
		return &Bool{Value: strings.Contains(string(self.Value), substr)}
	}},
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for indexOf must be CHARS, got %s", args[0].Type())
		}
		var substr string
		switch s := args[1].(type) {
		case *String:
			substr = s.Value
		case *Chars:
			substr = string(s.Value)
		default:
			return newError("argument for indexOf must be STRING or CHARS, got %s", args[1].Type())
		}
		// Return character index, not byte index
		byteIdx := strings.Index(string(self.Value), substr)
		if byteIdx < 0 {
			return NewInt(-1)
		}
		// Convert byte index to character index
		charIdx := len([]rune(string(self.Value)[:byteIdx]))
		return NewInt(int64(charIdx))
	}},
	"startsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for startsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for startsWith must be CHARS, got %s", args[0].Type())
		}
		var prefix string
		switch s := args[1].(type) {
		case *String:
			prefix = s.Value
		case *Chars:
			prefix = string(s.Value)
		default:
			return newError("argument for startsWith must be STRING or CHARS, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasPrefix(string(self.Value), prefix)}
	}},
	"endsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for endsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for endsWith must be CHARS, got %s", args[0].Type())
		}
		var suffix string
		switch s := args[1].(type) {
		case *String:
			suffix = s.Value
		case *Chars:
			suffix = string(s.Value)
		default:
			return newError("argument for endsWith must be STRING or CHARS, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasSuffix(string(self.Value), suffix)}
	}},
	"subStr": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for subStr. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for subStr must be CHARS, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("start index for subStr must be INT, got %s", args[1].Type())
		}
		runes := self.Value
		runeLen := len(runes)
		startIdx := int(start.Value)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > runeLen {
			startIdx = runeLen
		}
		if len(args) == 3 {
			end, ok := args[2].(*Int)
			if !ok {
				return newError("end index for subStr must be INT, got %s", args[2].Type())
			}
			endIdx := int(end.Value)
			if endIdx < startIdx {
				endIdx = startIdx
			}
			if endIdx > runeLen {
				endIdx = runeLen
			}
			return NewChars(runes[startIdx:endIdx])
		}
		return NewChars(runes[startIdx:])
	}},
	"at": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for at. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for at must be CHARS, got %s", args[0].Type())
		}
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("index for at must be INT, got %s", args[1].Type())
		}
		runes := self.Value
		runeLen := len(runes)
		actualIdx := int(idx.Value)
		if actualIdx < 0 {
			actualIdx = runeLen + actualIdx
		}
		if actualIdx < 0 || actualIdx >= runeLen {
			return newError("chars index out of bounds: %d (length: %d)", idx.Value, runeLen)
		}
		return NewString(string(runes[actualIdx]))
	}},
	"reverse": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for reverse must be CHARS, got %s", args[0].Type())
		}
		if len(self.Value) == 0 {
			return self
		}
		reversed := make([]rune, len(self.Value))
		for i := 0; i < len(self.Value); i++ {
			reversed[i] = self.Value[len(self.Value)-1-i]
		}
		return NewChars(reversed)
	}},
	"repeat": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for repeat. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for repeat must be CHARS, got %s", args[0].Type())
		}
		count, ok := args[1].(*Int)
		if !ok {
			return newError("count for repeat must be INT, got %s", args[1].Type())
		}
		if count.Value < 0 {
			return newError("count for repeat must be non-negative")
		}
		if count.Value == 0 {
			return CHARS_EMPTY
		}
		result := make([]rune, 0, len(self.Value)*int(count.Value))
		for i := int64(0); i < count.Value; i++ {
			result = append(result, self.Value...)
		}
		return NewChars(result)
	}},
}

// ============================================================
// Array Methods
// ============================================================

var arrayMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for len must be ARRAY, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Elements)))
	}},
	"push": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for push. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for push must be ARRAY, got %s", args[0].Type())
		}
		newElements := make([]Object, len(self.Elements)+1)
		copy(newElements, self.Elements)
		newElements[len(self.Elements)] = args[1]
		return NewArray(newElements)
	}},
	"pop": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for pop. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for pop must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return newError("cannot pop from empty array")
		}
		lastElem := self.Elements[len(self.Elements)-1]
		newElements := make([]Object, len(self.Elements)-1)
		copy(newElements, self.Elements[:len(self.Elements)-1])
		result := NewArray(newElements)
		result.LastPopped = lastElem
		return result
	}},
	"first": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for first. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for first must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return NULL
		}
		return self.Elements[0]
	}},
	"last": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for last. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for last must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return NULL
		}
		return self.Elements[len(self.Elements)-1]
	}},
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for indexOf must be ARRAY, got %s", args[0].Type())
		}
		for i, elem := range self.Elements {
			if compareObjects(elem, args[1]) {
				return NewInt(int64(i))
			}
		}
		return NewInt(-1)
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for contains must be ARRAY, got %s", args[0].Type())
		}
		for _, elem := range self.Elements {
			if compareObjects(elem, args[1]) {
				return TRUE
			}
		}
		return FALSE
	}},
	"reverse": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for reverse must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return self
		}
		reversed := make([]Object, len(self.Elements))
		for i := 0; i < len(self.Elements); i++ {
			reversed[i] = self.Elements[len(self.Elements)-1-i]
		}
		return NewArray(reversed)
	}},
	"join": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for join. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for join must be ARRAY, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("separator for join must be STRING, got %s", args[1].Type())
		}
		parts := make([]string, len(self.Elements))
		for i, elem := range self.Elements {
			if s, ok := elem.(*String); ok {
				parts[i] = s.Value
			} else {
				parts[i] = elem.Inspect()
			}
		}
		return NewString(strings.Join(parts, sep.Value))
	}},
	// sortByFunc sorts the array in-place using a custom comparator function.
	// The comparator function receives two indices (idx1, idx2) and returns true
	// if the element at idx1 should come before the element at idx2.
	// Returns the array itself (sorted in-place).
	"sortByFunc": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sortByFunc. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for sortByFunc must be ARRAY, got %s", args[0].Type())
		}

		if len(self.Elements) <= 1 {
			return self
		}

		// The comparator can be a Function, Closure, or Builtin
		comparator := args[1]

		// Sort using the comparator
		sort.Slice(self.Elements, func(i, j int) bool {
			// Call the comparator with two indices
			result, err := CallUserFunc(comparator, NewInt(int64(i)), NewInt(int64(j)))
			if err != nil {
				// If there's an error, maintain original order
				return false
			}
			// Convert result to boolean
			if b, ok := result.(*Bool); ok {
				return b.Value
			}
			// Non-boolean result: treat truthy values as true
			if result.Type() == NullType {
				return false
			}
			return true
		})

		return self
	}},
}

// ============================================================
// Map Methods
// ============================================================

var mapMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for len must be MAP, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Pairs)))
	}},
	"keys": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for keys. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for keys must be MAP, got %s", args[0].Type())
		}
		keys := make([]Object, len(self.Pairs))
		i := 0
		for _, pair := range self.Pairs {
			keys[i] = pair.Key
			i++
		}
		// Sort keys for deterministic output
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Inspect() < keys[j].Inspect()
		})
		return NewArray(keys)
	}},
	"values": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for values. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for values must be MAP, got %s", args[0].Type())
		}
		// Get keys and sort them for deterministic order
		keys := make([]Object, len(self.Pairs))
		i := 0
		for _, pair := range self.Pairs {
			keys[i] = pair.Key
			i++
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Inspect() < keys[j].Inspect()
		})
		// Get values in the same order as sorted keys
		vals := make([]Object, len(keys))
		for i, key := range keys {
			vals[i] = self.Pairs[key.HashKey()].Value
		}
		return NewArray(vals)
	}},
	"hasKey": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasKey. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for hasKey must be MAP, got %s", args[0].Type())
		}
		_, exists := self.Pairs[args[1].HashKey()]
		return &Bool{Value: exists}
	}},
	"delete": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for delete must be MAP, got %s", args[0].Type())
		}
		newPairs := make(map[HashKey]MapPair, len(self.Pairs)-1)
		for k, v := range self.Pairs {
			if k != args[1].HashKey() {
				newPairs[k] = v
			}
		}
		return NewMap(newPairs)
	}},
}

// ============================================================
// Bool Methods
// ============================================================

var boolMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
}

// ============================================================
// Null Methods
// ============================================================

var nullMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
}

// ============================================================
// StringBuilder Methods
// ============================================================

var stringBuilderMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for len must be STRING_BUILDER, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"write": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for write. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for write must be STRING_BUILDER, got %s", args[0].Type())
		}
		str, ok := args[1].(*String)
		if !ok {
			return newError("argument for write must be STRING, got %s", args[1].Type())
		}
		n := self.Write(str.Value)
		return NewInt(int64(n))
	}},
	"writeLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for writeLine must be STRING_BUILDER, got %s", args[0].Type())
		}
		str, ok := args[1].(*String)
		if !ok {
			return newError("argument for writeLine must be STRING, got %s", args[1].Type())
		}
		n := self.WriteLine(str.Value)
		return NewInt(int64(n))
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for toString must be STRING_BUILDER, got %s", args[0].Type())
		}
		return NewString(self.String())
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for clear must be STRING_BUILDER, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"reset": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reset. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for reset must be STRING_BUILDER, got %s", args[0].Type())
		}
		self.Reset()
		return NULL
	}},
	"grow": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grow. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for grow must be STRING_BUILDER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for grow must be INT, got %s", args[1].Type())
		}
		self.Grow(int(n.Value))
		return NULL
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for isEmpty must be STRING_BUILDER, got %s", args[0].Type())
		}
		return &Bool{Value: self.Len() == 0}
	}},
}

// ============================================================
// WebSocket Methods
// ============================================================

var webSocketMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// readMsg reads a message from the WebSocket connection.
	// Returns an array [messageType, data] or an error.
	// messageType: 1=text, 2=binary, 8=close, 9=ping, 10=pong
	"readMsg": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readMsg. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for readMsg must be WEBSOCKET, got %s", args[0].Type())
		}
		return self.ReadMessage()
	}},
	// sendTextMsg sends a text message over the WebSocket.
	// Usage: conn.sendTextMsg(text)
	"sendTextMsg": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sendTextMsg. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for sendTextMsg must be WEBSOCKET, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for sendTextMsg must be STRING, got %s", args[1].Type())
		}
		return self.SendTextMessage(text.Value)
	}},
	// sendBinaryMsg sends a binary message over the WebSocket.
	// Usage: conn.sendBinaryMsg(data)
	"sendBinaryMsg": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sendBinaryMsg. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for sendBinaryMsg must be WEBSOCKET, got %s", args[0].Type())
		}
		data, ok := args[1].(*String)
		if !ok {
			return newError("argument for sendBinaryMsg must be STRING, got %s", args[1].Type())
		}
		return self.SendBinaryMessage(data.Value)
	}},
	// sendCloseMsg sends a close message over the WebSocket.
	// Usage: conn.sendCloseMsg()
	"sendCloseMsg": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sendCloseMsg. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for sendCloseMsg must be WEBSOCKET, got %s", args[0].Type())
		}
		return self.SendCloseMessage()
	}},
	// close closes the WebSocket connection.
	// Usage: conn.close()
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for close must be WEBSOCKET, got %s", args[0].Type())
		}
		return self.Close()
	}},
	// isClosed returns whether the WebSocket is closed.
	// Usage: conn.isClosed()
	"isClosed": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isClosed. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WebSocket)
		if !ok {
			return newError("receiver for isClosed must be WEBSOCKET, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsClosed()}
	}},
}

// ============================================================
// Mutex Methods
// ============================================================

var mutexMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"lock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Mutex)
		if !ok {
			return newError("receiver for lock must be MUTEX, got %s", args[0].Type())
		}
		self.Lock()
		return NULL
	}},
	"unlock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for unlock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Mutex)
		if !ok {
			return newError("receiver for unlock must be MUTEX, got %s", args[0].Type())
		}
		self.Unlock()
		return NULL
	}},
	"tryLock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for tryLock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Mutex)
		if !ok {
			return newError("receiver for tryLock must be MUTEX, got %s", args[0].Type())
		}
		if self.TryLock() {
			return TRUE
		}
		return FALSE
	}},
}

// ============================================================
// RWMutex Methods
// ============================================================

var rwMutexMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"lock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RWMutex)
		if !ok {
			return newError("receiver for lock must be RWMUTEX, got %s", args[0].Type())
		}
		self.Lock()
		return NULL
	}},
	"unlock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for unlock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RWMutex)
		if !ok {
			return newError("receiver for unlock must be RWMUTEX, got %s", args[0].Type())
		}
		self.Unlock()
		return NULL
	}},
	"rLock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for rLock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RWMutex)
		if !ok {
			return newError("receiver for rLock must be RWMUTEX, got %s", args[0].Type())
		}
		self.RLock()
		return NULL
	}},
	"rUnlock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for rUnlock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RWMutex)
		if !ok {
			return newError("receiver for rUnlock must be RWMUTEX, got %s", args[0].Type())
		}
		self.RUnlock()
		return NULL
	}},
}

// ============================================================
// WaitGroup Methods
// ============================================================

var waitGroupMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"add": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for add. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*WaitGroup)
		if !ok {
			return newError("receiver for add must be WAITGROUP, got %s", args[0].Type())
		}
		delta, ok := args[1].(*Int)
		if !ok {
			return newError("argument for add must be INT, got %s", args[1].Type())
		}
		self.Add(int(delta.Value))
		return NULL
	}},
	"done": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for done. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WaitGroup)
		if !ok {
			return newError("receiver for done must be WAITGROUP, got %s", args[0].Type())
		}
		self.Done()
		return NULL
	}},
	"wait": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for wait. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*WaitGroup)
		if !ok {
			return newError("receiver for wait must be WAITGROUP, got %s", args[0].Type())
		}
		self.Wait()
		return NULL
	}},
}

// ============================================================
// AtomicInt Methods
// ============================================================

var atomicIntMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"add": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for add. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*AtomicInt)
		if !ok {
			return newError("receiver for add must be ATOMICINT, got %s", args[0].Type())
		}
		delta, ok := args[1].(*Int)
		if !ok {
			return newError("argument for add must be INT, got %s", args[1].Type())
		}
		return NewInt(self.Add(delta.Value))
	}},
	"load": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for load. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*AtomicInt)
		if !ok {
			return newError("receiver for load must be ATOMICINT, got %s", args[0].Type())
		}
		return NewInt(self.Load())
	}},
	"store": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for store. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*AtomicInt)
		if !ok {
			return newError("receiver for store must be ATOMICINT, got %s", args[0].Type())
		}
		val, ok := args[1].(*Int)
		if !ok {
			return newError("argument for store must be INT, got %s", args[1].Type())
		}
		self.Store(val.Value)
		return NULL
	}},
	"swap": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for swap. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*AtomicInt)
		if !ok {
			return newError("receiver for swap must be ATOMICINT, got %s", args[0].Type())
		}
		newVal, ok := args[1].(*Int)
		if !ok {
			return newError("argument for swap must be INT, got %s", args[1].Type())
		}
		return NewInt(self.Swap(newVal.Value))
	}},
	"compareAndSwap": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for compareAndSwap. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*AtomicInt)
		if !ok {
			return newError("receiver for compareAndSwap must be ATOMICINT, got %s", args[0].Type())
		}
		oldVal, ok := args[1].(*Int)
		if !ok {
			return newError("old value for compareAndSwap must be INT, got %s", args[1].Type())
		}
		newVal, ok := args[2].(*Int)
		if !ok {
			return newError("new value for compareAndSwap must be INT, got %s", args[2].Type())
		}
		if self.CompareAndSwap(oldVal.Value, newVal.Value) {
			return TRUE
		}
		return FALSE
	}},
}

// ============================================================
// Tube Methods
// ============================================================

var tubeMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"send": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for send. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for send must be TUBE, got %s", args[0].Type())
		}
		if !self.Send(args[1]) {
			return FALSE
		}
		return TRUE
	}},
	"receive": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for receive. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for receive must be TUBE, got %s", args[0].Type())
		}
		val, ok := self.Receive()
		if ok {
			return NewArray([]Object{val, TRUE})
		}
		return NewArray([]Object{val, FALSE})
	}},
	"trySend": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for trySend. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for trySend must be TUBE, got %s", args[0].Type())
		}
		sent, ok := self.TrySend(args[1])
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
	}},
	"tryReceive": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for tryReceive. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for tryReceive must be TUBE, got %s", args[0].Type())
		}
		val, received, open := self.TryReceive()
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
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for close must be TUBE, got %s", args[0].Type())
		}
		self.Close()
		return NULL
	}},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for len must be TUBE, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"cap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for cap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for cap must be TUBE, got %s", args[0].Type())
		}
		return NewInt(int64(self.Cap()))
	}},
	"isClosed": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isClosed. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Tube)
		if !ok {
			return newError("receiver for isClosed must be TUBE, got %s", args[0].Type())
		}
		if self.IsClosed() {
			return TRUE
		}
		return FALSE
	}},
}

// ============================================================
// Once Methods
// ============================================================

var onceMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"do": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for do. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Once)
		if !ok {
			return newError("receiver for do must be ONCE, got %s", args[0].Type())
		}
		// Note: Once.do() has limited functionality from Go
		// The function argument needs special VM handling
		// For now, just mark as called
		_ = self
		_ = args[1]
		return NULL
	}},
}

// ============================================================
// Cond Methods
// ============================================================

var condMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"wait": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for wait. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Cond)
		if !ok {
			return newError("receiver for wait must be COND, got %s", args[0].Type())
		}
		self.Wait()
		return NULL
	}},
	"signal": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for signal. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Cond)
		if !ok {
			return newError("receiver for signal must be COND, got %s", args[0].Type())
		}
		self.Signal()
		return NULL
	}},
	"broadcast": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for broadcast. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Cond)
		if !ok {
			return newError("receiver for broadcast must be COND, got %s", args[0].Type())
		}
		self.Broadcast()
		return NULL
	}},
}

// ============================================================
// Context Methods
// ============================================================

var contextMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"done": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for done. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for done must be CONTEXT, got %s", args[0].Type())
		}
		return self.Done()
	}},
	"cancel": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for cancel. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for cancel must be CONTEXT, got %s", args[0].Type())
		}
		self.Cancel()
		return NULL
	}},
	"err": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for err. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for err must be CONTEXT, got %s", args[0].Type())
		}
		errStr := self.ErrString()
		if errStr == "" {
			return NULL
		}
		return NewString(errStr)
	}},
	"isDone": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isDone. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for isDone must be CONTEXT, got %s", args[0].Type())
		}
		if self.IsDone() {
			return TRUE
		}
		return FALSE
	}},
	"deadline": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for deadline. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for deadline must be CONTEXT, got %s", args[0].Type())
		}
		dl, hasDeadline := self.Deadline()
		if !hasDeadline {
			return NULL
		}
		return NewInt(dl.UnixMilli())
	}},
	"deadlineStr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for deadlineStr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Context)
		if !ok {
			return newError("receiver for deadlineStr must be CONTEXT, got %s", args[0].Type())
		}
		dlStr := self.DeadlineString()
		if dlStr == "" {
			return NULL
		}
		return NewString(dlStr)
	}},
}
