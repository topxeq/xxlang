// pkg/objects/methods.go
package objects

import (
	"math"
	"path/filepath"
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
	BytesType:         bytesMethods,
	BytesBufferType:   bytesBufferMethods,
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
	// File upload types
	FileUploadType:      fileUploadMethods,
	FileUploadResultType: fileUploadResultMethods,
	// File types
	FileType:     fileMethods,
	FileInfoType: fileInfoMethods,
	// I/O types
	ReaderType:  readerMethods,
	WriterType:  writerMethods,
	ScannerType: scannerMethods,
	// Ordered map
	OrderedMapType: orderedMapMethods,
	// Queue and Set
	QueueType: queueMethods,
	SetType:   setMethods,
	// XLSX
	XLSXType: xlsxMethods,
	// XML
	XMLDocumentType: xmlDocumentMethods,
	XMLNodeType:     xmlNodeMethods,
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
	// getWriter returns a Writer for the StringBuilder.
	"getWriter": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getWriter. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for getWriter must be STRING_BUILDER, got %s", args[0].Type())
		}
		return NewWriter(self.GetIOWriter())
	}},
}

// ============================================================
// Bytes Methods
// ============================================================

var bytesMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for len must be BYTES, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"at": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for at. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for at must be BYTES, got %s", args[0].Type())
		}
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("argument for at must be INT, got %s", args[1].Type())
		}
		val, ok := self.At(int(idx.Value))
		if !ok {
			return newError("index out of range")
		}
		return NewInt(val)
	}},
	"slice": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for slice. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for slice must be BYTES, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("start index must be INT, got %s", args[1].Type())
		}
		end := len(self.Value)
		if len(args) == 3 {
			endVal, ok := args[2].(*Int)
			if !ok {
				return newError("end index must be INT, got %s", args[2].Type())
			}
			end = int(endVal.Value)
		}
		return self.Slice(int(start.Value), end)
	}},
	"toArray": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toArray. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for toArray must be BYTES, got %s", args[0].Type())
		}
		return self.ToArray()
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for toString must be BYTES, got %s", args[0].Type())
		}
		return NewString(self.String())
	}},
	"getReader": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getReader. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for getReader must be BYTES, got %s", args[0].Type())
		}
		return NewReader(self.GetIOReader())
	}},
	"hasPrefix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasPrefix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for hasPrefix must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for hasPrefix must be BYTES, got %s", args[1].Type())
		}
		return &Bool{Value: self.HasPrefix(other)}
	}},
	"hasSuffix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasSuffix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for hasSuffix must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for hasSuffix must be BYTES, got %s", args[1].Type())
		}
		return &Bool{Value: self.HasSuffix(other)}
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for contains must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for contains must be BYTES, got %s", args[1].Type())
		}
		return &Bool{Value: self.Contains(other)}
	}},
	"index": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for index. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for index must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for index must be BYTES, got %s", args[1].Type())
		}
		return NewInt(int64(self.Index(other)))
	}},
	"count": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for count. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for count must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for count must be BYTES, got %s", args[1].Type())
		}
		return NewInt(int64(self.Count(other)))
	}},
	"repeat": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for repeat. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for repeat must be BYTES, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for repeat must be INT, got %s", args[1].Type())
		}
		return self.Repeat(int(n.Value))
	}},
	"concat": {Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for concat. got=%d, want>=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for concat must be BYTES, got %s", args[0].Type())
		}
		others := make([]*Bytes, len(args)-1)
		for i, arg := range args[1:] {
			other, ok := arg.(*Bytes)
			if !ok {
				return newError("argument %d for concat must be BYTES, got %s", i+2, arg.Type())
			}
			others[i] = other
		}
		return self.Concat(others...)
	}},
	"equal": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for equal. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Bytes)
		if !ok {
			return newError("receiver for equal must be BYTES, got %s", args[0].Type())
		}
		other, ok := args[1].(*Bytes)
		if !ok {
			return newError("argument for equal must be BYTES, got %s", args[1].Type())
		}
		return &Bool{Value: self.Equal(other)}
	}},
}

// ============================================================
// BytesBuffer Methods
// ============================================================

var bytesBufferMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for len must be BYTES_BUFFER, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"cap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for cap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for cap must be BYTES_BUFFER, got %s", args[0].Type())
		}
		return NewInt(int64(self.Cap()))
	}},
	"write": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for write. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for write must be BYTES_BUFFER, got %s", args[0].Type())
		}
		switch v := args[1].(type) {
		case *String:
			n := self.WriteString(v.Value)
			return NewInt(int64(n))
		case *Array:
			// Convert array of ints to bytes
			data := make([]byte, len(v.Elements))
			for i, elem := range v.Elements {
				b, ok := elem.(*Int)
				if !ok {
					return newError("array elements must be INT for write, got %s", elem.Type())
				}
				if b.Value < 0 || b.Value > 255 {
					return newError("array element %d out of byte range: %d", i, b.Value)
				}
				data[i] = byte(b.Value)
			}
			n := self.Write(data)
			return NewInt(int64(n))
		default:
			return newError("argument for write must be STRING or ARRAY, got %s", args[1].Type())
		}
	}},
	"writeByte": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeByte. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeByte must be BYTES_BUFFER, got %s", args[0].Type())
		}
		b, ok := args[1].(*Int)
		if !ok {
			return newError("argument for writeByte must be INT, got %s", args[1].Type())
		}
		if b.Value < 0 || b.Value > 255 {
			return newError("byte value out of range: %d", b.Value)
		}
		err := self.WriteByte(byte(b.Value))
		if err != nil {
			return newError("writeByte error: %s", err.Error())
		}
		return NULL
	}},
	"writeInt16": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeInt16. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeInt16 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, ok := args[1].(*Int)
		if !ok {
			return newError("argument for writeInt16 must be INT, got %s", args[1].Type())
		}
		err := self.WriteInt16(int16(v.Value))
		if err != nil {
			return newError("writeInt16 error: %s", err.Error())
		}
		return NULL
	}},
	"writeInt32": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeInt32. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeInt32 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, ok := args[1].(*Int)
		if !ok {
			return newError("argument for writeInt32 must be INT, got %s", args[1].Type())
		}
		err := self.WriteInt32(int32(v.Value))
		if err != nil {
			return newError("writeInt32 error: %s", err.Error())
		}
		return NULL
	}},
	"writeInt64": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeInt64. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeInt64 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, ok := args[1].(*Int)
		if !ok {
			return newError("argument for writeInt64 must be INT, got %s", args[1].Type())
		}
		err := self.WriteInt64(v.Value)
		if err != nil {
			return newError("writeInt64 error: %s", err.Error())
		}
		return NULL
	}},
	"writeFloat32": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeFloat32. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeFloat32 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, ok := args[1].(*Float)
		if !ok {
			return newError("argument for writeFloat32 must be FLOAT, got %s", args[1].Type())
		}
		err := self.WriteFloat32(float32(v.Value))
		if err != nil {
			return newError("writeFloat32 error: %s", err.Error())
		}
		return NULL
	}},
	"writeFloat64": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeFloat64. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for writeFloat64 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, ok := args[1].(*Float)
		if !ok {
			return newError("argument for writeFloat64 must be FLOAT, got %s", args[1].Type())
		}
		err := self.WriteFloat64(v.Value)
		if err != nil {
			return newError("writeFloat64 error: %s", err.Error())
		}
		return NULL
	}},
	"bytes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for bytes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for bytes must be BYTES_BUFFER, got %s", args[0].Type())
		}
		data := self.Bytes()
		elements := make([]Object, len(data))
		for i, b := range data {
			elements[i] = NewInt(int64(b))
		}
		return &Array{Elements: elements}
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for toString must be BYTES_BUFFER, got %s", args[0].Type())
		}
		return NewString(self.String())
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for clear must be BYTES_BUFFER, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"reset": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reset. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for reset must be BYTES_BUFFER, got %s", args[0].Type())
		}
		self.Reset()
		return NULL
	}},
	"grow": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grow. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for grow must be BYTES_BUFFER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for grow must be INT, got %s", args[1].Type())
		}
		self.Grow(int(n.Value))
		return NULL
	}},
	"truncate": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for truncate. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for truncate must be BYTES_BUFFER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for truncate must be INT, got %s", args[1].Type())
		}
		self.Truncate(int(n.Value))
		return NULL
	}},
	"readByte": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readByte. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readByte must be BYTES_BUFFER, got %s", args[0].Type())
		}
		b, err := self.ReadByte()
		if err != nil {
			return NULL
		}
		return NewInt(int64(b))
	}},
	"readInt16": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readInt16. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readInt16 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, err := self.ReadInt16()
		if err != nil {
			return NULL
		}
		return NewInt(int64(v))
	}},
	"readInt32": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readInt32. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readInt32 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, err := self.ReadInt32()
		if err != nil {
			return NULL
		}
		return NewInt(int64(v))
	}},
	"readInt64": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readInt64. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readInt64 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, err := self.ReadInt64()
		if err != nil {
			return NULL
		}
		return NewInt(v)
	}},
	"readFloat32": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readFloat32. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readFloat32 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, err := self.ReadFloat32()
		if err != nil {
			return NULL
		}
		return NewFloat(float64(v))
	}},
	"readFloat64": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readFloat64. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for readFloat64 must be BYTES_BUFFER, got %s", args[0].Type())
		}
		v, err := self.ReadFloat64()
		if err != nil {
			return NULL
		}
		return NewFloat(v)
	}},
	"peek": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for peek. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for peek must be BYTES_BUFFER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for peek must be INT, got %s", args[1].Type())
		}
		data := self.Peek(int(n.Value))
		elements := make([]Object, len(data))
		for i, b := range data {
			elements[i] = NewInt(int64(b))
		}
		return &Array{Elements: elements}
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BytesBuffer)
		if !ok {
			return newError("receiver for isEmpty must be BYTES_BUFFER, got %s", args[0].Type())
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

// ============================================================
// FileUpload Methods
// ============================================================

var fileUploadMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"filename": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for filename. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for filename must be FILE_UPLOAD, got %s", args[0].Type())
		}
		if self.Header == nil {
			return NULL
		}
		return NewString(self.Header.Filename)
	}},
	"size": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for size. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for size must be FILE_UPLOAD, got %s", args[0].Type())
		}
		if self.Header == nil {
			return NewInt(0)
		}
		return NewInt(self.Header.Size)
	}},
	"extension": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for extension. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for extension must be FILE_UPLOAD, got %s", args[0].Type())
		}
		if self.Header == nil {
			return NULL
		}
		return NewString(strings.TrimPrefix(filepath.Ext(self.Header.Filename), "."))
	}},
	"contentType": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for contentType. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for contentType must be FILE_UPLOAD, got %s", args[0].Type())
		}
		if self.Header == nil {
			return NULL
		}
		return NewString(self.Header.Header.Get("Content-Type"))
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for save must be FILE_UPLOAD, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument to save must be STRING, got %s", args[1].Type())
		}
		savedPath, err := self.Save(path.Value)
		if err != nil {
			return newError("save failed: %v", err)
		}
		return NewString(savedPath)
	}},
	"saveToDir": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for saveToDir. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for saveToDir must be FILE_UPLOAD, got %s", args[0].Type())
		}
		dir, ok := args[1].(*String)
		if !ok {
			return newError("first argument to saveToDir must be STRING, got %s", args[1].Type())
		}
		autoRename := false
		if len(args) == 3 {
			ar, ok := args[2].(*Bool)
			if !ok {
				return newError("second argument to saveToDir must be BOOL, got %s", args[2].Type())
			}
			autoRename = ar.Value
		}
		savedPath, err := self.SaveToDir(dir.Value, autoRename)
		if err != nil {
			return newError("saveToDir failed: %v", err)
		}
		return NewString(savedPath)
	}},
	"read": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for read. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for read must be FILE_UPLOAD, got %s", args[0].Type())
		}
		content, err := self.ReadAsString()
		if err != nil {
			return newError("read failed: %v", err)
		}
		return NewString(content)
	}},
	"readBytes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readBytes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for readBytes must be FILE_UPLOAD, got %s", args[0].Type())
		}
		data, err := self.ReadAll()
		if err != nil {
			return newError("readBytes failed: %v", err)
		}
		return NewBytesBufferFromBytes(data)
	}},
	"hashSHA256": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hashSHA256. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for hashSHA256 must be FILE_UPLOAD, got %s", args[0].Type())
		}
		hash, err := self.HashSHA256()
		if err != nil {
			return newError("hashSHA256 failed: %v", err)
		}
		return NewString(hash)
	}},
	// getReader opens the uploaded file and returns a Reader for streaming access.
	// The returned Reader supports Close method.
	"getReader": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getReader. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for getReader must be FILE_UPLOAD, got %s", args[0].Type())
		}
		file, err := self.Open()
		if err != nil {
			return newError("getReader failed: %v", err)
		}
		return NewReader(file)
	}},
	// open opens the uploaded file and returns a Reader (alias for getReader).
	"open": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for open. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUpload)
		if !ok {
			return newError("receiver for open must be FILE_UPLOAD, got %s", args[0].Type())
		}
		file, err := self.Open()
		if err != nil {
			return newError("open failed: %v", err)
		}
		return NewReader(file)
	}},
}

// ============================================================
// FileUploadResult Methods
// ============================================================

var fileUploadResultMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"success": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for success. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUploadResult)
		if !ok {
			return newError("receiver for success must be FILE_UPLOAD_RESULT, got %s", args[0].Type())
		}
		return &Bool{Value: self.Success}
	}},
	"message": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for message. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUploadResult)
		if !ok {
			return newError("receiver for message must be FILE_UPLOAD_RESULT, got %s", args[0].Type())
		}
		return NewString(self.Message)
	}},
	"path": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for path. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUploadResult)
		if !ok {
			return newError("receiver for path must be FILE_UPLOAD_RESULT, got %s", args[0].Type())
		}
		return NewString(self.FilePath)
	}},
	"originalName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for originalName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUploadResult)
		if !ok {
			return newError("receiver for originalName must be FILE_UPLOAD_RESULT, got %s", args[0].Type())
		}
		return NewString(self.OriginalName)
	}},
	"size": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for size. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileUploadResult)
		if !ok {
			return newError("receiver for size must be FILE_UPLOAD_RESULT, got %s", args[0].Type())
		}
		return NewInt(self.Size)
	}},
}

// ============================================================
// File Methods
// ============================================================

var fileMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// close closes the file handle.
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for close must be FILE, got %s", args[0].Type())
		}
		err := self.Close()
		if err != nil {
			return newError("close failed: %s", err.Error())
		}
		return NULL
	}},
	// read reads up to n bytes from the file and returns as string.
	"read": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for read. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for read must be FILE, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for read must be INT, got %s", args[1].Type())
		}
		data, err := self.Read(int(n.Value))
		if err != nil {
			return newError("read failed: %s", err.Error())
		}
		return NewString(string(data))
	}},
	// readBytes reads up to n bytes from the file and returns as array of integers.
	"readBytes": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for readBytes. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for readBytes must be FILE, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for readBytes must be INT, got %s", args[1].Type())
		}
		data, err := self.Read(int(n.Value))
		if err != nil {
			return newError("readBytes failed: %s", err.Error())
		}
		elements := make([]Object, len(data))
		for i, b := range data {
			elements[i] = NewInt(int64(b))
		}
		return NewArray(elements)
	}},
	// readLine reads a single line from the file.
	"readLine": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readLine. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for readLine must be FILE, got %s", args[0].Type())
		}
		line, err := self.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				return NULL
			}
			return newError("readLine failed: %s", err.Error())
		}
		return NewString(line)
	}},
	// readAll reads all remaining content from the file.
	"readAll": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readAll. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for readAll must be FILE, got %s", args[0].Type())
		}
		data, err := self.ReadAll()
		if err != nil {
			return newError("readAll failed: %s", err.Error())
		}
		return NewString(string(data))
	}},
	// write writes a string to the file.
	"write": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for write. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for write must be FILE, got %s", args[0].Type())
		}
		s, ok := args[1].(*String)
		if !ok {
			return newError("argument for write must be STRING, got %s", args[1].Type())
		}
		n, err := self.WriteString(s.Value)
		if err != nil {
			return newError("write failed: %s", err.Error())
		}
		return NewInt(int64(n))
	}},
	// writeLine writes a string with newline to the file.
	"writeLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for writeLine must be FILE, got %s", args[0].Type())
		}
		s, ok := args[1].(*String)
		if !ok {
			return newError("argument for writeLine must be STRING, got %s", args[1].Type())
		}
		n, err := self.WriteString(s.Value + "\n")
		if err != nil {
			return newError("writeLine failed: %s", err.Error())
		}
		return NewInt(int64(n))
	}},
	// seek sets the file position.
	"seek": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for seek. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for seek must be FILE, got %s", args[0].Type())
		}
		offset, ok := args[1].(*Int)
		if !ok {
			return newError("offset for seek must be INT, got %s", args[1].Type())
		}
		whence := 0 // default: seek from start
		if len(args) == 3 {
			w, ok := args[2].(*Int)
			if !ok {
				return newError("whence for seek must be INT, got %s", args[2].Type())
			}
			whence = int(w.Value)
		}
		pos, err := self.Seek(offset.Value, whence)
		if err != nil {
			return newError("seek failed: %s", err.Error())
		}
		return NewInt(pos)
	}},
	// tell returns the current file position.
	"tell": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for tell. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for tell must be FILE, got %s", args[0].Type())
		}
		return NewInt(self.Tell())
	}},
	// flush flushes buffered data to disk.
	"flush": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for flush. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for flush must be FILE, got %s", args[0].Type())
		}
		err := self.Flush()
		if err != nil {
			return newError("flush failed: %s", err.Error())
		}
		return NULL
	}},
	// isOpen returns whether the file is open.
	"isOpen": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isOpen. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for isOpen must be FILE, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsOpen()}
	}},
	// name returns the file path.
	"name": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for name. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for name must be FILE, got %s", args[0].Type())
		}
		return NewString(self.GetName())
	}},
	// mode returns the file open mode.
	"mode": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for mode. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for mode must be FILE, got %s", args[0].Type())
		}
		return NewString(string(self.GetMode()))
	}},
	// lock places a lock on the file.
	"lock": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for lock. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for lock must be FILE, got %s", args[0].Type())
		}
		lockType, ok := args[1].(*Int)
		if !ok {
			return newError("lockType for lock must be INT, got %s", args[1].Type())
		}
		blocking := true
		if len(args) == 3 {
			b, ok := args[2].(*Bool)
			if !ok {
				return newError("blocking for lock must be BOOL, got %s", args[2].Type())
			}
			blocking = b.Value
		}
		err := self.Lock(FileLockType(lockType.Value), blocking)
		if err != nil {
			return newError("lock failed: %s", err.Error())
		}
		return NULL
	}},
	// unlock releases the file lock.
	"unlock": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for unlock. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for unlock must be FILE, got %s", args[0].Type())
		}
		err := self.Unlock()
		if err != nil {
			return newError("unlock failed: %s", err.Error())
		}
		return NULL
	}},
	// truncate truncates the file to the specified size.
	"truncate": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for truncate. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for truncate must be FILE, got %s", args[0].Type())
		}
		size, ok := args[1].(*Int)
		if !ok {
			return newError("size for truncate must be INT, got %s", args[1].Type())
		}
		err := self.Truncate(size.Value)
		if err != nil {
			return newError("truncate failed: %s", err.Error())
		}
		return NULL
	}},
	// stat returns file information.
	"stat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for stat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*File)
		if !ok {
			return newError("receiver for stat must be FILE, got %s", args[0].Type())
		}
		info, err := self.Stat()
		if err != nil {
			return newError("stat failed: %s", err.Error())
		}
		return NewFileInfo(info, self.Path)
	}},
}

// ============================================================
// FileInfo Methods
// ============================================================

var fileInfoMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// name returns the file name.
	"name": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for name. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for name must be FILE_INFO, got %s", args[0].Type())
		}
		return NewString(self.Name)
	}},
	// size returns the file size in bytes.
	"size": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for size. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for size must be FILE_INFO, got %s", args[0].Type())
		}
		return NewInt(self.Size)
	}},
	// mode returns the file mode as an integer.
	"mode": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for mode. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for mode must be FILE_INFO, got %s", args[0].Type())
		}
		return NewInt(int64(self.Mode))
	}},
	// modeStr returns the file mode as an octal string.
	"modeStr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for modeStr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for modeStr must be FILE_INFO, got %s", args[0].Type())
		}
		return NewString(self.GetModeString())
	}},
	// modTime returns the modification time as a formatted string.
	"modTime": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for modTime. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for modTime must be FILE_INFO, got %s", args[0].Type())
		}
		return NewString(self.GetModTimeString())
	}},
	// modTimeUnix returns the modification time as Unix timestamp in milliseconds.
	"modTimeUnix": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for modTimeUnix. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for modTimeUnix must be FILE_INFO, got %s", args[0].Type())
		}
		return NewInt(self.GetModTimeUnix())
	}},
	// isDir returns whether the file is a directory.
	"isDir": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isDir. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for isDir must be FILE_INFO, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsDir}
	}},
	// isFile returns whether this is a regular file.
	"isFile": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isFile. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for isFile must be FILE_INFO, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsRegular()}
	}},
	// isSymlink returns whether this is a symbolic link.
	"isSymlink": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isSymlink. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for isSymlink must be FILE_INFO, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsSymlink()}
	}},
	// path returns the full path to the file.
	"path": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for path. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FileInfo)
		if !ok {
			return newError("receiver for path must be FILE_INFO, got %s", args[0].Type())
		}
		return NewString(self.FullPath)
	}},
}

// ============================================================
// Reader Methods
// ============================================================

var readerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// read reads up to n bytes and returns as array of integers.
	"read": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for read. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for read must be READER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for read must be INT, got %s", args[1].Type())
		}
		return self.Read(int(n.Value))
	}},
	// readStr reads up to n bytes and returns as string.
	"readStr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for readStr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for readStr must be READER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for readStr must be INT, got %s", args[1].Type())
		}
		return self.ReadStr(int(n.Value))
	}},
	// readAllStr reads all remaining content as string.
	"readAllStr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readAllStr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for readAllStr must be READER, got %s", args[0].Type())
		}
		return self.ReadAllStr()
	}},
	// readAllBytes reads all remaining content as byte array.
	"readAllBytes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readAllBytes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for readAllBytes must be READER, got %s", args[0].Type())
		}
		return self.ReadAllBytes()
	}},
	// readLine reads a single line from the reader.
	"readLine": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for readLine. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for readLine must be READER, got %s", args[0].Type())
		}
		return self.ReadLine()
	}},
	// close closes the reader if it implements io.Closer.
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Reader)
		if !ok {
			return newError("receiver for close must be READER, got %s", args[0].Type())
		}
		return self.Close()
	}},
}

// ============================================================
// Writer Methods
// ============================================================

var writerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// write writes a byte array to the writer.
	"write": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for write. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Writer)
		if !ok {
			return newError("receiver for write must be WRITER, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("argument for write must be ARRAY, got %s", args[1].Type())
		}
		return self.WriteBytes(arr)
	}},
	// writeStr writes a string to the writer.
	"writeStr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeStr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Writer)
		if !ok {
			return newError("receiver for writeStr must be WRITER, got %s", args[0].Type())
		}
		s, ok := args[1].(*String)
		if !ok {
			return newError("argument for writeStr must be STRING, got %s", args[1].Type())
		}
		return self.WriteStr(s.Value)
	}},
	// writeBytes writes a byte array to the writer.
	"writeBytes": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeBytes. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Writer)
		if !ok {
			return newError("receiver for writeBytes must be WRITER, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("argument for writeBytes must be ARRAY, got %s", args[1].Type())
		}
		return self.WriteBytes(arr)
	}},
	// close closes the writer if it implements io.Closer.
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Writer)
		if !ok {
			return newError("receiver for close must be WRITER, got %s", args[0].Type())
		}
		return self.Close()
	}},
}

// ============================================================
// Scanner Methods
// ============================================================

var scannerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// next reads the next whitespace-delimited token.
	"next": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for next. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for next must be SCANNER, got %s", args[0].Type())
		}
		return self.next()
	}},
	// nextLine reads the next line.
	"nextLine": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for nextLine. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for nextLine must be SCANNER, got %s", args[0].Type())
		}
		return self.nextLine()
	}},
	// nextInt reads the next token as an integer.
	"nextInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for nextInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for nextInt must be SCANNER, got %s", args[0].Type())
		}
		return self.nextInt()
	}},
	// nextFloat reads the next token as a float.
	"nextFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for nextFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for nextFloat must be SCANNER, got %s", args[0].Type())
		}
		return self.nextFloat()
	}},
	// nextBool reads the next token as a boolean.
	"nextBool": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for nextBool. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for nextBool must be SCANNER, got %s", args[0].Type())
		}
		return self.nextBool()
	}},
	// hasNext checks if there is more input.
	"hasNext": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hasNext. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for hasNext must be SCANNER, got %s", args[0].Type())
		}
		return self.hasNext()
	}},
	// skipLine skips the current line.
	"skipLine": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for skipLine. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for skipLine must be SCANNER, got %s", args[0].Type())
		}
		return self.skipLine()
	}},
	// close closes the scanner.
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Scanner)
		if !ok {
			return newError("receiver for close must be SCANNER, got %s", args[0].Type())
		}
		return self.close()
	}},
}

// ============================================================
// OrderedMap Methods
// ============================================================

var orderedMapMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	// len returns the number of entries
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for len must be ORDERED_MAP, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},

	// keys returns keys in insertion order
	"keys": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for keys. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for keys must be ORDERED_MAP, got %s", args[0].Type())
		}
		return NewArray(self.GetOrderedKeys())
	}},

	// values returns values in insertion order
	"values": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for values. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for values must be ORDERED_MAP, got %s", args[0].Type())
		}
		return NewArray(self.GetOrderedValues())
	}},

	// hasKey checks if key exists
	"hasKey": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasKey. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for hasKey must be ORDERED_MAP, got %s", args[0].Type())
		}
		return &Bool{Value: self.HasKey(args[1])}
	}},

	// delete removes a key, maintaining order
	"delete": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for delete must be ORDERED_MAP, got %s", args[0].Type())
		}
		// Create a clone and delete from it (immutable operation)
		newMap := self.Clone()
		newMap.Delete(args[1])
		return newMap
	}},

	// entries returns [key, value] pairs in insertion order
	"entries": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for entries. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for entries must be ORDERED_MAP, got %s", args[0].Type())
		}
		return NewArray(self.GetOrderedPairs())
	}},

	// moveToFront moves key to position 0
	"moveToFront": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for moveToFront. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for moveToFront must be ORDERED_MAP, got %s", args[0].Type())
		}
		err := self.MoveToFront(args[1])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// moveToBack moves key to last position
	"moveToBack": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for moveToBack. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for moveToBack must be ORDERED_MAP, got %s", args[0].Type())
		}
		err := self.MoveToBack(args[1])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// moveBefore moves key1 before key2
	"moveBefore": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for moveBefore. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for moveBefore must be ORDERED_MAP, got %s", args[0].Type())
		}
		err := self.MoveBefore(args[1], args[2])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// moveAfter moves key1 after key2
	"moveAfter": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for moveAfter. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for moveAfter must be ORDERED_MAP, got %s", args[0].Type())
		}
		err := self.MoveAfter(args[1], args[2])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// swap swaps positions of two keys
	"swap": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for swap. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for swap must be ORDERED_MAP, got %s", args[0].Type())
		}
		err := self.Swap(args[1], args[2])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// insertAt inserts key-value pair at specific index
	"insertAt": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for insertAt. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for insertAt must be ORDERED_MAP, got %s", args[0].Type())
		}
		idx, ok := args[3].(*Int)
		if !ok {
			return newError("index must be INT, got %s", args[3].Type())
		}
		err := self.InsertAt(args[1], args[2], int(idx.Value))
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// indexOf returns index of key (-1 if not found)
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for indexOf must be ORDERED_MAP, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetIndex(args[1])))
	}},

	// getAt returns [key, value] at index
	"getAt": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getAt. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for getAt must be ORDERED_MAP, got %s", args[0].Type())
		}
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT, got %s", args[1].Type())
		}
		key, value, err := self.GetAt(int(idx.Value))
		if err != nil {
			return newError("%s", err.Error())
		}
		return NewArray([]Object{key, value})
	}},

	// setAt updates value at index
	"setAt": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setAt. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for setAt must be ORDERED_MAP, got %s", args[0].Type())
		}
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT, got %s", args[1].Type())
		}
		err := self.SetAt(int(idx.Value), args[2])
		if err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},

	// reverse reverses order of all elements
	"reverse": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for reverse must be ORDERED_MAP, got %s", args[0].Type())
		}
		self.Reverse()
		return NULL
	}},

	// sortByKey sorts by key alphabetically
	"sortByKey": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sortByKey. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for sortByKey must be ORDERED_MAP, got %s", args[0].Type())
		}
		self.SortByKey()
		return NULL
	}},

	// toMap converts to regular Map
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for toMap must be ORDERED_MAP, got %s", args[0].Type())
		}
		return self.ToMap()
	}},

	// clone creates a deep copy
	"clone": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*OrderedMap)
		if !ok {
			return newError("receiver for clone must be ORDERED_MAP, got %s", args[0].Type())
		}
		return self.Clone()
	}},
}

// ============================================================
// Queue Methods
// ============================================================

var queueMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for len must be QUEUE, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"push": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for push. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for push must be QUEUE, got %s", args[0].Type())
		}
		self.Push(args[1])
		return NULL
	}},
	"pop": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for pop. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for pop must be QUEUE, got %s", args[0].Type())
		}
		return self.Pop()
	}},
	"peek": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for peek. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for peek must be QUEUE, got %s", args[0].Type())
		}
		return self.Peek()
	}},
	"peekBack": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for peekBack. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for peekBack must be QUEUE, got %s", args[0].Type())
		}
		return self.PeekBack()
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for isEmpty must be QUEUE, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsEmpty()}
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for clear must be QUEUE, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"toArray": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toArray. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for toArray must be QUEUE, got %s", args[0].Type())
		}
		return self.ToArray()
	}},
	"clone": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Queue)
		if !ok {
			return newError("receiver for clone must be QUEUE, got %s", args[0].Type())
		}
		return self.Clone()
	}},
}

// ============================================================
// Set Methods
// ============================================================

var setMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for len must be SET, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"add": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for add. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for add must be SET, got %s", args[0].Type())
		}
		return &Bool{Value: self.Add(args[1])}
	}},
	"remove": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for remove. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for remove must be SET, got %s", args[0].Type())
		}
		return &Bool{Value: self.Remove(args[1])}
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for contains must be SET, got %s", args[0].Type())
		}
		return &Bool{Value: self.Contains(args[1])}
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for isEmpty must be SET, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsEmpty()}
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for clear must be SET, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"toArray": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toArray. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for toArray must be SET, got %s", args[0].Type())
		}
		return self.ToArray()
	}},
	"toSortedArray": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toSortedArray. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for toSortedArray must be SET, got %s", args[0].Type())
		}
		return self.ToSortedArray()
	}},
	"clone": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for clone must be SET, got %s", args[0].Type())
		}
		return self.Clone()
	}},
	"union": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for union. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for union must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for union must be SET, got %s", args[1].Type())
		}
		return self.Union(other)
	}},
	"intersect": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for intersect. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for intersect must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for intersect must be SET, got %s", args[1].Type())
		}
		return self.Intersect(other)
	}},
	"difference": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for difference. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for difference must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for difference must be SET, got %s", args[1].Type())
		}
		return self.Difference(other)
	}},
	"symmetricDiff": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for symmetricDiff. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for symmetricDiff must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for symmetricDiff must be SET, got %s", args[1].Type())
		}
		return self.SymmetricDifference(other)
	}},
	"isSubset": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isSubset. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for isSubset must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for isSubset must be SET, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsSubset(other)}
	}},
	"isSuperset": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isSuperset. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for isSuperset must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for isSuperset must be SET, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsSuperset(other)}
	}},
	"equals": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for equals. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Set)
		if !ok {
			return newError("receiver for equals must be SET, got %s", args[0].Type())
		}
		other, ok := args[1].(*Set)
		if !ok {
			return newError("argument for equals must be SET, got %s", args[1].Type())
		}
		return &Bool{Value: self.Equals(other)}
	}},
}

// ============================================================
// XLSX Methods
// ============================================================

var xlsxMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for close must be XLSX, got %s", args[0].Type())
		}
		self.Close()
		return NULL
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for save. got=%d, want>=1", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for save must be XLSX, got %s", args[0].Type())
		}
		path := ""
		if len(args) >= 2 {
			if p, ok := args[1].(*String); ok {
				path = p.Value
			}
		}
		if err := self.Save(path); err != nil {
			return newError("save failed: %s", err.Error())
		}
		return NULL
	}},
	"getSheetList": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSheetList. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getSheetList must be XLSX, got %s", args[0].Type())
		}
		list := self.GetSheetList()
		elements := make([]Object, len(list))
		for i, name := range list {
			elements[i] = &String{Value: name}
		}
		return &Array{Elements: elements}
	}},
	"getSheetCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSheetCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getSheetCount must be XLSX, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetSheetCount()))
	}},
	"getSheetName": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getSheetName. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getSheetName must be XLSX, got %s", args[0].Type())
		}
		idx, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT, got %s", args[1].Type())
		}
		name := self.GetSheetName(int(idx.Value))
		if name == "" {
			return newError("sheet index out of range: %d", idx.Value)
		}
		return &String{Value: name}
	}},
	"newSheet": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for newSheet. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for newSheet must be XLSX, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("argument for newSheet must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.NewSheet(name.Value)}
	}},
	"deleteSheet": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for deleteSheet. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for deleteSheet must be XLSX, got %s", args[0].Type())
		}
		// Support both string name and integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("argument for deleteSheet must be STRING or INT, got %s", args[1].Type())
		}
		return &Bool{Value: self.DeleteSheet(sheetName)}
	}},
	"getCell": {Fn: func(args ...Object) Object {
		if len(args) < 3 {
			return newError("wrong number of arguments for getCell. got=%d, want>=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getCell must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		// Args[2] can be string ref or row number
		if ref, ok := args[2].(*String); ok {
			return self.GetCell(sheetName, ref.Value)
		}
		// Row, col form
		if len(args) < 4 {
			return newError("wrong number of arguments for getCell with row/col")
		}
		row, ok1 := args[2].(*Int)
		col, ok2 := args[3].(*Int)
		if !ok1 || !ok2 {
			return newError("row and col must be INT")
		}
		return self.GetCellByIndex(sheetName, int(row.Value), int(col.Value))
	}},
	"setCell": {Fn: func(args ...Object) Object {
		if len(args) < 4 {
			return newError("wrong number of arguments for setCell. got=%d, want>=4", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for setCell must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		// Args[2] can be string ref or row number
		if ref, ok := args[2].(*String); ok {
			if len(args) < 4 {
				return newError("missing value argument")
			}
			if err := self.SetCell(sheetName, ref.Value, args[3]); err != nil {
				return newError("%s", err.Error())
			}
			return NULL
		}
		// Row, col form
		if len(args) < 5 {
			return newError("wrong number of arguments for setCell with row/col")
		}
		row, ok1 := args[2].(*Int)
		col, ok2 := args[3].(*Int)
		if !ok1 || !ok2 {
			return newError("row and col must be INT")
		}
		if err := self.SetCellByIndex(sheetName, int(row.Value), int(col.Value), args[4]); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"getRow": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getRow. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getRow must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		row, ok := args[2].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		return self.GetRow(sheetName, int(row.Value))
	}},
	"setRow": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for setRow. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for setRow must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		row, ok := args[2].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		values, ok := args[3].(*Array)
		if !ok {
			return newError("values must be ARRAY")
		}
		if err := self.SetRow(sheetName, int(row.Value), values); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"getCol": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getCol. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getCol must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		col, ok := args[2].(*Int)
		if !ok {
			return newError("col must be INT")
		}
		return self.GetCol(sheetName, int(col.Value))
	}},
	"getRange": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getRange. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getRange must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		rng, ok := args[2].(*String)
		if !ok {
			return newError("range must be STRING")
		}
		return self.GetRange(sheetName, rng.Value)
	}},
	"getRowCount": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getRowCount. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getRowCount must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		return NewInt(int64(self.GetRowCount(sheetName)))
	}},
	"getColCount": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getColCount. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getColCount must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		return NewInt(int64(self.GetColCount(sheetName)))
	}},
	"insertRow": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for insertRow. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for insertRow must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		row, ok := args[2].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		if err := self.InsertRow(sheetName, int(row.Value)); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"deleteRow": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for deleteRow. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for deleteRow must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		row, ok := args[2].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		if err := self.DeleteRow(sheetName, int(row.Value)); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"insertCol": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for insertCol. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for insertCol must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		col, ok := args[2].(*Int)
		if !ok {
			return newError("col must be INT")
		}
		if err := self.InsertCol(sheetName, int(col.Value)); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"deleteCol": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for deleteCol. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for deleteCol must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		col, ok := args[2].(*Int)
		if !ok {
			return newError("col must be INT")
		}
		if err := self.DeleteCol(sheetName, int(col.Value)); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"mergeCell": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for mergeCell. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for mergeCell must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		start, ok := args[2].(*String)
		if !ok {
			return newError("start ref must be STRING")
		}
		end, ok := args[3].(*String)
		if !ok {
			return newError("end ref must be STRING")
		}
		if err := self.MergeCell(sheetName, start.Value, end.Value); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"unmergeCell": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for unmergeCell. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for unmergeCell must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		ref, ok := args[2].(*String)
		if !ok {
			return newError("ref must be STRING")
		}
		if err := self.UnmergeCell(sheetName, ref.Value); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"getMerges": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getMerges. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getMerges must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		return self.GetMerges(sheetName)
	}},
	"getImages": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getImages. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getImages must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		return self.GetImages(sheetName)
	}},
	"extractImage": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for extractImage. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for extractImage must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		imageIdx, ok := args[2].(*Int)
		if !ok {
			return newError("image index must be INT")
		}
		outputPath, ok := args[3].(*String)
		if !ok {
			return newError("output path must be STRING")
		}
		if err := self.ExtractImage(sheetName, int(imageIdx.Value), outputPath.Value); err != nil {
			return newError("%s", err.Error())
		}
		return NULL
	}},
	"getImageData": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getImageData. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XLSX)
		if !ok {
			return newError("receiver for getImageData must be XLSX, got %s", args[0].Type())
		}
		// Resolve sheet name from string or integer index
		var sheetName string
		switch v := args[1].(type) {
		case *String:
			sheetName = v.Value
		case *Int:
			sheetName = self.GetSheetName(int(v.Value))
			if sheetName == "" {
				return newError("sheet index out of range: %d", v.Value)
			}
		default:
			return newError("sheet must be STRING or INT")
		}
		imageIdx, ok := args[2].(*Int)
		if !ok {
			return newError("image index must be INT")
		}
		data, err := self.GetImageData(sheetName, int(imageIdx.Value))
		if err != nil {
			return newError("%s", err.Error())
		}
		return &String{Value: data}
	}},
}

// ============================================================
// XML Document Methods
// ============================================================

var xmlDocumentMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"root": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for root. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for root must be XMLDocument, got %s", args[0].Type())
		}
		root := self.Root()
		if root == nil {
			return NULL
		}
		return root
	}},
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for find must be XMLDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		return self.Find(path.Value)
	}},
	"findFirst": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findFirst. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for findFirst must be XMLDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		node := self.FindFirst(path.Value)
		if node == nil {
			return NULL
		}
		return node
	}},
	"findElement": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findElement. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for findElement must be XMLDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		node := self.FindElement(path.Value)
		if node == nil {
			return NULL
		}
		return node
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for toString must be XMLDocument, got %s", args[0].Type())
		}
		return NewString(self.ToString())
	}},
	"toIndented": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toIndented. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for toIndented must be XMLDocument, got %s", args[0].Type())
		}
		return NewString(self.ToIndented())
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for save must be XMLDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		if err := self.Save(path.Value); err != nil {
			return newError("save failed: %s", err.Error())
		}
		return NULL
	}},
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for toMap must be XMLDocument, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
	"version": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for version. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for version must be XMLDocument, got %s", args[0].Type())
		}
		return NewString(self.Version())
	}},
	"encoding": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for encoding. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLDocument)
		if !ok {
			return newError("receiver for encoding must be XMLDocument, got %s", args[0].Type())
		}
		return NewString(self.Encoding())
	}},
}

// ============================================================
// XML Node Methods
// ============================================================

var xmlNodeMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"name": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for name. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for name must be XMLNode, got %s", args[0].Type())
		}
		return NewString(self.Name())
	}},
	"setName": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setName. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for setName must be XMLNode, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		self.SetName(name.Value)
		return NULL
	}},
	"text": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for text. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for text must be XMLNode, got %s", args[0].Type())
		}
		return NewString(self.Text())
	}},
	"setText": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setText. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for setText must be XMLNode, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		self.SetText(text.Value)
		return NULL
	}},
	"attr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for attr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for attr must be XMLNode, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		return NewString(self.Attr(name.Value))
	}},
	"setAttr": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setAttr. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for setAttr must be XMLNode, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		value, ok := args[2].(*String)
		if !ok {
			return newError("value must be STRING")
		}
		self.SetAttr(name.Value, value.Value)
		return NULL
	}},
	"delAttr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delAttr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for delAttr must be XMLNode, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		self.DelAttr(name.Value)
		return NULL
	}},
	"attrs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for attrs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for attrs must be XMLNode, got %s", args[0].Type())
		}
		return self.Attrs()
	}},
	"children": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for children. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for children must be XMLNode, got %s", args[0].Type())
		}
		return self.Children()
	}},
	"childCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for childCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for childCount must be XMLNode, got %s", args[0].Type())
		}
		return NewInt(int64(self.ChildCount()))
	}},
	"parent": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for parent. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for parent must be XMLNode, got %s", args[0].Type())
		}
		p := self.Parent()
		if p == nil {
			return NULL
		}
		return p
	}},
	"addChild": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addChild. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for addChild must be XMLNode, got %s", args[0].Type())
		}
		child, ok := args[1].(*XMLNode)
		if !ok {
			return newError("child must be XMLNode")
		}
		self.AddChild(child)
		return NULL
	}},
	"removeChild": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeChild. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for removeChild must be XMLNode, got %s", args[0].Type())
		}
		index, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT")
		}
		return &Bool{Value: self.RemoveChild(int(index.Value))}
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for clear must be XMLNode, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for find must be XMLNode, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		return self.Find(path.Value)
	}},
	"findFirst": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findFirst. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for findFirst must be XMLNode, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		node := self.FindFirst(path.Value)
		if node == nil {
			return NULL
		}
		return node
	}},
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for toMap must be XMLNode, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for toString must be XMLNode, got %s", args[0].Type())
		}
		return NewString(self.ToString())
	}},
	"toIndented": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toIndented. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*XMLNode)
		if !ok {
			return newError("receiver for toIndented must be XMLNode, got %s", args[0].Type())
		}
		return NewString(self.ToIndented())
	}},
}
