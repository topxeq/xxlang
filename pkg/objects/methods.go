// pkg/objects/methods.go
package objects

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
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
	MutexType:     mutexMethods,
	RWMutexType:   rwMutexMethods,
	WaitGroupType: waitGroupMethods,
	AtomicIntType: atomicIntMethods,
	TubeType:      tubeMethods,
	OnceType:      onceMethods,
	CondType:      condMethods,
	ContextType:   contextMethods,
	// File upload types
	FileUploadType:       fileUploadMethods,
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
	// PPTX
	PPTXDocumentType:  pptxDocumentMethods,
	PPTXSlideType:     pptxSlideMethods,
	PPTXTextFrameType: pptxTextFrameMethods,
	PPTXTextRunType:   pptxTextRunMethods,
	PPTXShapeType:     pptxShapeMethods,
	PPTXTableType:     pptxTableMethods,
	PPTXChartType:     pptxChartMethods,
	PPTXImageType:     pptxImageMethods,
	// XML
	XMLDocumentType: xmlDocumentMethods,
	XMLNodeType:     xmlNodeMethods,
	// Socket types
	SocketAddrType: socketAddrMethods,
	TcpServerType:  tcpServerMethods,
	TcpClientType:  tcpClientMethods,
	UdpSocketType:  udpSocketMethods,
	// LineEditor
	LineEditorType: lineEditorMethods,
	// SSH
	SSHClientType: sshClientMethods,
	// FTP/SFTP
	FtpClientType:  ftpClientMethods,
	FtpServerType:  ftpServerMethods,
	SftpClientType: sftpClientMethods,
	SftpServerType: sftpServerMethods,
	// HTML
	HTMLDocumentType: htmlDocumentMethods,
	HTMLElementType:  htmlElementMethods,
	// TOML
	TomlDocumentType: tomlDocumentMethods,
	// Time
	TimeType: timeMethods,
	// Rod browser (web scraping)
	RodBrowserType:     rodBrowserMethods,
	RodHTMLElementType: rodHTMLElementMethods,
	// Backup types
	BackupTaskType:   backupTaskMethods,
	BackupResultType: backupResultMethods,
	// HLBR (headless browser) types
	HlbrBrowserType: hlbrBrowserMethods,
	HlbrNodeType:    hlbrNodeMethods,
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
		return NewInt(int64(utf8.RuneCountInString(self.Value)))
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
		byteIdx := strings.Index(self.Value, substr.Value)
		if byteIdx < 0 {
			return NewInt(-1)
		}
		charIdx := utf8.RuneCountInString(self.Value[:byteIdx])
		return NewInt(int64(charIdx))
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
		runes := []rune(self.Value)
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
			return NewString(string(runes[startIdx:endIdx]))
		}
		return NewString(string(runes[startIdx:]))
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
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Chars)
		if !ok {
			return newError("receiver for toString must be CHARS, got %s", args[0].Type())
		}
		return NewString(string(self.Value))
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
// LineEditor Methods
// ============================================================

var lineEditorMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	// Basic Operations
	"lineCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lineCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for lineCount must be LINE_EDITOR, got %s", args[0].Type())
		}
		return NewInt(int64(self.LineCount()))
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for isEmpty must be LINE_EDITOR, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsEmpty()}
	}},
	"isModified": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isModified. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for isModified must be LINE_EDITOR, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsModified()}
	}},
	"getLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for getLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for getLine must be INT, got %s", args[1].Type())
		}
		line, ok := self.GetLine(int(n.Value))
		if !ok {
			return newError("line index out of range")
		}
		return NewString(line)
	}},
	"setLine": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setLine. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for setLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for setLine must be INT, got %s", args[1].Type())
		}
		text, ok := args[2].(*String)
		if !ok {
			return newError("second argument for setLine must be STRING, got %s", args[2].Type())
		}
		if !self.SetLine(int(n.Value), text.Value) {
			return newError("line index out of range")
		}
		return args[0] // Return self for chaining
	}},
	"addLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for addLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for addLine must be STRING, got %s", args[1].Type())
		}
		self.AddLine(text.Value)
		return args[0]
	}},
	"insertLine": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for insertLine. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for insertLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for insertLine must be INT, got %s", args[1].Type())
		}
		text, ok := args[2].(*String)
		if !ok {
			return newError("second argument for insertLine must be STRING, got %s", args[2].Type())
		}
		if !self.InsertLine(int(n.Value), text.Value) {
			return newError("line index out of range")
		}
		return args[0]
	}},
	"deleteLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for deleteLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for deleteLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for deleteLine must be INT, got %s", args[1].Type())
		}
		if !self.DeleteLine(int(n.Value)) {
			return newError("line index out of range")
		}
		return args[0]
	}},
	"deleteLines": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for deleteLines. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for deleteLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for deleteLines must be INT, got %s", args[1].Type())
		}
		end, ok := args[2].(*Int)
		if !ok {
			return newError("second argument for deleteLines must be INT, got %s", args[2].Type())
		}
		if !self.DeleteLines(int(start.Value), int(end.Value)) {
			return newError("invalid line range")
		}
		return args[0]
	}},

	// Line Range Operations
	"getLines": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getLines. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for getLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for getLines must be INT, got %s", args[1].Type())
		}
		end, ok := args[2].(*Int)
		if !ok {
			return newError("second argument for getLines must be INT, got %s", args[2].Type())
		}
		lines := self.GetLines(int(start.Value), int(end.Value))
		if lines == nil {
			return EMPTY_ARRAY
		}
		elements := make([]Object, len(lines))
		for i, line := range lines {
			elements[i] = NewString(line)
		}
		return NewArray(elements)
	}},
	"setLines": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setLines. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for setLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for setLines must be INT, got %s", args[1].Type())
		}
		arr, ok := args[2].(*Array)
		if !ok {
			return newError("second argument for setLines must be ARRAY, got %s", args[2].Type())
		}
		lines := make([]string, len(arr.Elements))
		for i, elem := range arr.Elements {
			if s, ok := elem.(*String); ok {
				lines[i] = s.Value
			} else {
				lines[i] = elem.Inspect()
			}
		}
		if !self.SetLines(int(start.Value), lines) {
			return newError("line index out of range")
		}
		return args[0]
	}},
	"appendLines": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for appendLines. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for appendLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("argument for appendLines must be ARRAY, got %s", args[1].Type())
		}
		lines := make([]string, len(arr.Elements))
		for i, elem := range arr.Elements {
			if s, ok := elem.(*String); ok {
				lines[i] = s.Value
			} else {
				lines[i] = elem.Inspect()
			}
		}
		self.AppendLines(lines)
		return args[0]
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for clear must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Clear()
		return args[0]
	}},

	// Search Operations
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for find must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for find must be STRING, got %s", args[1].Type())
		}
		lineNums := self.Find(text.Value)
		elements := make([]Object, len(lineNums))
		for i, n := range lineNums {
			elements[i] = NewInt(int64(n))
		}
		return NewArray(elements)
	}},
	"findRegex": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findRegex. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for findRegex must be LINE_EDITOR, got %s", args[0].Type())
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("argument for findRegex must be STRING, got %s", args[1].Type())
		}
		lineNums, err := self.FindRegex(pattern.Value)
		if err != nil {
			return newError("invalid regex pattern: %s", err.Error())
		}
		elements := make([]Object, len(lineNums))
		for i, n := range lineNums {
			elements[i] = NewInt(int64(n))
		}
		return NewArray(elements)
	}},
	"findAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for findAll must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for findAll must be STRING, got %s", args[1].Type())
		}
		lines := self.FindAll(text.Value)
		elements := make([]Object, len(lines))
		for i, line := range lines {
			elements[i] = NewString(line)
		}
		return NewArray(elements)
	}},
	"findFirst": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findFirst. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for findFirst must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for findFirst must be STRING, got %s", args[1].Type())
		}
		return NewInt(int64(self.FindFirst(text.Value)))
	}},
	"findLast": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findLast. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for findLast must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for findLast must be STRING, got %s", args[1].Type())
		}
		return NewInt(int64(self.FindLast(text.Value)))
	}},
	"grep": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grep. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for grep must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for grep must be STRING, got %s", args[1].Type())
		}
		return self.Grep(text.Value)
	}},
	"grepRegex": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grepRegex. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for grepRegex must be LINE_EDITOR, got %s", args[0].Type())
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("argument for grepRegex must be STRING, got %s", args[1].Type())
		}
		result, err := self.GrepRegex(pattern.Value)
		if err != nil {
			return newError("invalid regex pattern: %s", err.Error())
		}
		return result
	}},
	"grepNot": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grepNot. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for grepNot must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for grepNot must be STRING, got %s", args[1].Type())
		}
		return self.GrepNot(text.Value)
	}},
	"grepNotRegex": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grepNotRegex. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for grepNotRegex must be LINE_EDITOR, got %s", args[0].Type())
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("argument for grepNotRegex must be STRING, got %s", args[1].Type())
		}
		result, err := self.GrepNotRegex(pattern.Value)
		if err != nil {
			return newError("invalid regex pattern: %s", err.Error())
		}
		return result
	}},

	// Replace Operations
	"replace": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for replace. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replace must be LINE_EDITOR, got %s", args[0].Type())
		}
		old, ok := args[1].(*String)
		if !ok {
			return newError("first argument for replace must be STRING, got %s", args[1].Type())
		}
		newStr, ok := args[2].(*String)
		if !ok {
			return newError("second argument for replace must be STRING, got %s", args[2].Type())
		}
		count := self.Replace(old.Value, newStr.Value)
		return NewInt(int64(count))
	}},
	"replaceLine": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for replaceLine. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replaceLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for replaceLine must be INT, got %s", args[1].Type())
		}
		old, ok := args[2].(*String)
		if !ok {
			return newError("second argument for replaceLine must be STRING, got %s", args[2].Type())
		}
		newStr, ok := args[3].(*String)
		if !ok {
			return newError("third argument for replaceLine must be STRING, got %s", args[3].Type())
		}
		count := self.ReplaceLine(int(n.Value), old.Value, newStr.Value)
		return NewInt(int64(count))
	}},
	"replaceFirst": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for replaceFirst. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replaceFirst must be LINE_EDITOR, got %s", args[0].Type())
		}
		old, ok := args[1].(*String)
		if !ok {
			return newError("first argument for replaceFirst must be STRING, got %s", args[1].Type())
		}
		newStr, ok := args[2].(*String)
		if !ok {
			return newError("second argument for replaceFirst must be STRING, got %s", args[2].Type())
		}
		return &Bool{Value: self.ReplaceFirst(old.Value, newStr.Value)}
	}},
	"replaceLast": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for replaceLast. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replaceLast must be LINE_EDITOR, got %s", args[0].Type())
		}
		old, ok := args[1].(*String)
		if !ok {
			return newError("first argument for replaceLast must be STRING, got %s", args[1].Type())
		}
		newStr, ok := args[2].(*String)
		if !ok {
			return newError("second argument for replaceLast must be STRING, got %s", args[2].Type())
		}
		return &Bool{Value: self.ReplaceLast(old.Value, newStr.Value)}
	}},
	"replaceRegex": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for replaceRegex. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replaceRegex must be LINE_EDITOR, got %s", args[0].Type())
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("first argument for replaceRegex must be STRING, got %s", args[1].Type())
		}
		newStr, ok := args[2].(*String)
		if !ok {
			return newError("second argument for replaceRegex must be STRING, got %s", args[2].Type())
		}
		count, err := self.ReplaceRegex(pattern.Value, newStr.Value)
		if err != nil {
			return newError("invalid regex pattern: %s", err.Error())
		}
		return NewInt(int64(count))
	}},
	"replaceRange": {Fn: func(args ...Object) Object {
		if len(args) != 5 {
			return newError("wrong number of arguments for replaceRange. got=%d, want=5", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for replaceRange must be LINE_EDITOR, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for replaceRange must be INT, got %s", args[1].Type())
		}
		end, ok := args[2].(*Int)
		if !ok {
			return newError("second argument for replaceRange must be INT, got %s", args[2].Type())
		}
		old, ok := args[3].(*String)
		if !ok {
			return newError("third argument for replaceRange must be STRING, got %s", args[3].Type())
		}
		newStr, ok := args[4].(*String)
		if !ok {
			return newError("fourth argument for replaceRange must be STRING, got %s", args[4].Type())
		}
		count := self.ReplaceRange(int(start.Value), int(end.Value), old.Value, newStr.Value)
		return NewInt(int64(count))
	}},

	// Sort and Unique Operations
	"sort": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sort. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sort must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Sort()
		return args[0]
	}},
	"sortDesc": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sortDesc. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sortDesc must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.SortDesc()
		return args[0]
	}},
	"sortNum": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sortNum. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sortNum must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.SortNum()
		return args[0]
	}},
	"sortNumDesc": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sortNumDesc. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sortNumDesc must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.SortNumDesc()
		return args[0]
	}},
	"sortByCol": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for sortByCol. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sortByCol must be LINE_EDITOR, got %s", args[0].Type())
		}
		col, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for sortByCol must be INT, got %s", args[1].Type())
		}
		sep, ok := args[2].(*String)
		if !ok {
			return newError("second argument for sortByCol must be STRING, got %s", args[2].Type())
		}
		self.SortByCol(int(col.Value), sep.Value)
		return args[0]
	}},
	"sortByColNum": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for sortByColNum. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sortByColNum must be LINE_EDITOR, got %s", args[0].Type())
		}
		col, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for sortByColNum must be INT, got %s", args[1].Type())
		}
		sep, ok := args[2].(*String)
		if !ok {
			return newError("second argument for sortByColNum must be STRING, got %s", args[2].Type())
		}
		self.SortByColNum(int(col.Value), sep.Value)
		return args[0]
	}},
	"reverse": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for reverse must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Reverse()
		return args[0]
	}},
	"shuffle": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for shuffle. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for shuffle must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Shuffle()
		return args[0]
	}},
	"unique": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for unique. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for unique must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Unique()
		return args[0]
	}},
	"uniqueSorted": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for uniqueSorted. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for uniqueSorted must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.UniqueSorted()
		return args[0]
	}},
	"findDupes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for findDupes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for findDupes must be LINE_EDITOR, got %s", args[0].Type())
		}
		dupes := self.FindDupes()
		pairs := make([]Object, 0, len(dupes))
		for line, count := range dupes {
			pairs = append(pairs, &Array{Elements: []Object{
				NewString(line),
				NewInt(int64(count)),
			}})
		}
		return NewArray(pairs)
	}},
	"removeDupes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for removeDupes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for removeDupes must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.RemoveDupes()
		return args[0]
	}},
	"keepDupes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for keepDupes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for keepDupes must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.KeepDupes()
		return args[0]
	}},

	// Text Processing Operations
	"trim": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trim. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for trim must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Trim()
		return args[0]
	}},
	"trimLeft": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trimLeft. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for trimLeft must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.TrimLeft()
		return args[0]
	}},
	"trimRight": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trimRight. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for trimRight must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.TrimRight()
		return args[0]
	}},
	"removeEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for removeEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for removeEmpty must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.RemoveEmpty()
		return args[0]
	}},
	"removeBlank": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for removeBlank. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for removeBlank must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.RemoveBlank()
		return args[0]
	}},
	"dedent": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for dedent. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for dedent must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Dedent()
		return args[0]
	}},
	"indent": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indent. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for indent must be LINE_EDITOR, got %s", args[0].Type())
		}
		prefix, ok := args[1].(*String)
		if !ok {
			return newError("argument for indent must be STRING, got %s", args[1].Type())
		}
		self.Indent(prefix.Value)
		return args[0]
	}},
	"numberLines": {Fn: func(args ...Object) Object {
		if len(args) != 1 && len(args) != 2 {
			return newError("wrong number of arguments for numberLines. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for numberLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		start := 1
		if len(args) == 2 {
			if n, ok := args[1].(*Int); ok {
				start = int(n.Value)
			}
		}
		self.NumberLines(start)
		return args[0]
	}},
	"join": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for join. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for join must be LINE_EDITOR, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("argument for join must be STRING, got %s", args[1].Type())
		}
		return NewString(self.Join(sep.Value))
	}},
	"splitLines": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for splitLines. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for splitLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("argument for splitLines must be STRING, got %s", args[1].Type())
		}
		self.SplitLines(sep.Value)
		return args[0]
	}},
	"prefix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for prefix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for prefix must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for prefix must be STRING, got %s", args[1].Type())
		}
		self.Prefix(text.Value)
		return args[0]
	}},
	"suffix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for suffix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for suffix must be LINE_EDITOR, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("argument for suffix must be STRING, got %s", args[1].Type())
		}
		self.Suffix(text.Value)
		return args[0]
	}},
	"toUpperCase": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toUpperCase. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for toUpperCase must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.ToUpperCase()
		return args[0]
	}},
	"toLowerCase": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toLowerCase. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for toLowerCase must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.ToLowerCase()
		return args[0]
	}},

	// Export and Save Operations
	"toText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for toText must be LINE_EDITOR, got %s", args[0].Type())
		}
		return NewString(self.ToText())
	}},
	"toLines": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toLines. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for toLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		lines := self.ToLines()
		elements := make([]Object, len(lines))
		for i, line := range lines {
			elements[i] = NewString(line)
		}
		return NewArray(elements)
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 1 && len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for save must be LINE_EDITOR, got %s", args[0].Type())
		}
		path := ""
		if len(args) == 2 {
			if p, ok := args[1].(*String); ok {
				path = p.Value
			}
		}
		var err error
		if path != "" {
			err = self.SaveAs(path)
		} else {
			err = self.Save()
		}
		if err != nil {
			return newError("save failed: %s", err.Error())
		}
		return args[0]
	}},
	"saveAs": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for saveAs. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for saveAs must be LINE_EDITOR, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for saveAs must be STRING, got %s", args[1].Type())
		}
		if err := self.SaveAs(path.Value); err != nil {
			return newError("saveAs failed: %s", err.Error())
		}
		return args[0]
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for close must be LINE_EDITOR, got %s", args[0].Type())
		}
		self.Close()
		return NULL
	}},

	// Statistics Methods
	"charCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for charCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for charCount must be LINE_EDITOR, got %s", args[0].Type())
		}
		return NewInt(int64(self.CharCount()))
	}},
	"runeCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for runeCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for runeCount must be LINE_EDITOR, got %s", args[0].Type())
		}
		return NewInt(int64(self.RuneCount()))
	}},
	"wordCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for wordCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for wordCount must be LINE_EDITOR, got %s", args[0].Type())
		}
		return NewInt(int64(self.WordCount()))
	}},
	"info": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for info. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for info must be LINE_EDITOR, got %s", args[0].Type())
		}
		info := self.Info()
		pairs := make([]Object, 0, len(info))
		for k, v := range info {
			pairs = append(pairs, &Array{Elements: []Object{
				NewString(k),
				NewInt(int64(v)),
			}})
		}
		return NewArray(pairs)
	}},

	// File Operations
	"appendToFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for appendToFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for appendToFile must be LINE_EDITOR, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for appendToFile must be STRING, got %s", args[1].Type())
		}
		if err := self.AppendToFile(path.Value); err != nil {
			return newError("appendToFile failed: %s", err.Error())
		}
		return args[0]
	}},
	"appendFromFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for appendFromFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for appendFromFile must be LINE_EDITOR, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for appendFromFile must be STRING, got %s", args[1].Type())
		}
		if err := self.AppendFromFile(path.Value); err != nil {
			return newError("appendFromFile failed: %s", err.Error())
		}
		return args[0]
	}},

	// Line Operations (Additional)
	"swapLines": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for swapLines. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for swapLines must be LINE_EDITOR, got %s", args[0].Type())
		}
		n1, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for swapLines must be INT, got %s", args[1].Type())
		}
		n2, ok := args[2].(*Int)
		if !ok {
			return newError("second argument for swapLines must be INT, got %s", args[2].Type())
		}
		if !self.SwapLines(int(n1.Value), int(n2.Value)) {
			return newError("line index out of range")
		}
		return args[0]
	}},
	"moveLine": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for moveLine. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for moveLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		from, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for moveLine must be INT, got %s", args[1].Type())
		}
		to, ok := args[2].(*Int)
		if !ok {
			return newError("second argument for moveLine must be INT, got %s", args[2].Type())
		}
		if !self.MoveLine(int(from.Value), int(to.Value)) {
			return newError("line index out of range")
		}
		return args[0]
	}},
	"duplicateLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for duplicateLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for duplicateLine must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for duplicateLine must be INT, got %s", args[1].Type())
		}
		if !self.DuplicateLine(int(n.Value)) {
			return newError("line index out of range")
		}
		return args[0]
	}},

	// Text Processing (Additional)
	"truncate": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for truncate. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for truncate must be LINE_EDITOR, got %s", args[0].Type())
		}
		maxLen, ok := args[1].(*Int)
		if !ok {
			return newError("argument for truncate must be INT, got %s", args[1].Type())
		}
		self.Truncate(int(maxLen.Value))
		return args[0]
	}},
	"truncateWithEllipsis": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for truncateWithEllipsis. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for truncateWithEllipsis must be LINE_EDITOR, got %s", args[0].Type())
		}
		maxLen, ok := args[1].(*Int)
		if !ok {
			return newError("argument for truncateWithEllipsis must be INT, got %s", args[1].Type())
		}
		self.TruncateWithEllipsis(int(maxLen.Value))
		return args[0]
	}},
	"padLeft": {Fn: func(args ...Object) Object {
		if len(args) != 2 && len(args) != 3 {
			return newError("wrong number of arguments for padLeft. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for padLeft must be LINE_EDITOR, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for padLeft must be INT, got %s", args[1].Type())
		}
		padChar := " "
		if len(args) == 3 {
			if p, ok := args[2].(*String); ok {
				padChar = p.Value
			}
		}
		self.PadLeft(int(width.Value), padChar)
		return args[0]
	}},
	"padRight": {Fn: func(args ...Object) Object {
		if len(args) != 2 && len(args) != 3 {
			return newError("wrong number of arguments for padRight. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for padRight must be LINE_EDITOR, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("first argument for padRight must be INT, got %s", args[1].Type())
		}
		padChar := " "
		if len(args) == 3 {
			if p, ok := args[2].(*String); ok {
				padChar = p.Value
			}
		}
		self.PadRight(int(width.Value), padChar)
		return args[0]
	}},
	"center": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for center. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for center must be LINE_EDITOR, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("argument for center must be INT, got %s", args[1].Type())
		}
		self.Center(int(width.Value))
		return args[0]
	}},
	"stripPrefix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for stripPrefix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for stripPrefix must be LINE_EDITOR, got %s", args[0].Type())
		}
		prefix, ok := args[1].(*String)
		if !ok {
			return newError("argument for stripPrefix must be STRING, got %s", args[1].Type())
		}
		self.StripPrefix(prefix.Value)
		return args[0]
	}},
	"stripSuffix": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for stripSuffix. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for stripSuffix must be LINE_EDITOR, got %s", args[0].Type())
		}
		suffix, ok := args[1].(*String)
		if !ok {
			return newError("argument for stripSuffix must be STRING, got %s", args[1].Type())
		}
		self.StripSuffix(suffix.Value)
		return args[0]
	}},
	"comment": {Fn: func(args ...Object) Object {
		if len(args) != 1 && len(args) != 2 {
			return newError("wrong number of arguments for comment. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for comment must be LINE_EDITOR, got %s", args[0].Type())
		}
		prefix := "#"
		if len(args) == 2 {
			if p, ok := args[1].(*String); ok {
				prefix = p.Value
			}
		}
		self.Comment(prefix)
		return args[0]
	}},
	"uncomment": {Fn: func(args ...Object) Object {
		if len(args) != 1 && len(args) != 2 {
			return newError("wrong number of arguments for uncomment. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for uncomment must be LINE_EDITOR, got %s", args[0].Type())
		}
		prefix := "#"
		if len(args) == 2 {
			if p, ok := args[1].(*String); ok {
				prefix = p.Value
			}
		}
		self.Uncomment(prefix)
		return args[0]
	}},

	// Sample and Selection
	"sample": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sample. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*LineEditor)
		if !ok {
			return newError("receiver for sample must be LINE_EDITOR, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for sample must be INT, got %s", args[1].Type())
		}
		return self.Sample(int(n.Value))
	}},
}

// ============================================================
// SSHClient Methods
// ============================================================

var sshClientMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	// Connection Management
	"isConnected": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isConnected. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for isConnected must be SSH_CLIENT, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsConnected()}
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for close must be SSH_CLIENT, got %s", args[0].Type())
		}
		if err := self.Close(); err != nil {
			return newError("close failed: %s", err.Error())
		}
		return NULL
	}},
	"getHost": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getHost. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for getHost must be SSH_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetHost())
	}},
	"getPort": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPort. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for getPort must be SSH_CLIENT, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetPort()))
	}},
	"getUser": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getUser. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for getUser must be SSH_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetUser())
	}},

	// Command Execution
	"exec": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for exec. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for exec must be SSH_CLIENT, got %s", args[0].Type())
		}
		cmd, ok := args[1].(*String)
		if !ok {
			return newError("argument for exec must be STRING, got %s", args[1].Type())
		}
		output, err := self.Exec(cmd.Value)
		if err != nil {
			return newError("exec failed: %s", err.Error())
		}
		return NewString(output)
	}},
	"execFull": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for execFull. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for execFull must be SSH_CLIENT, got %s", args[0].Type())
		}
		cmd, ok := args[1].(*String)
		if !ok {
			return newError("argument for execFull must be STRING, got %s", args[1].Type())
		}
		result, err := self.ExecFull(cmd.Value)
		if err != nil {
			return newError("execFull failed: %s", err.Error())
		}
		// Convert result map to Xxlang map
		pairs := make([]Object, 0, len(result))
		for k, v := range result {
			var valObj Object
			switch vv := v.(type) {
			case string:
				valObj = NewString(vv)
			case int:
				valObj = NewInt(int64(vv))
			case int64:
				valObj = NewInt(vv)
			default:
				valObj = NewString(fmt.Sprintf("%v", vv))
			}
			pairs = append(pairs, &Array{Elements: []Object{NewString(k), valObj}})
		}
		return NewArray(pairs)
	}},
	"runScript": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for runScript. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for runScript must be SSH_CLIENT, got %s", args[0].Type())
		}
		scriptPath, ok := args[1].(*String)
		if !ok {
			return newError("argument for runScript must be STRING, got %s", args[1].Type())
		}
		output, err := self.RunScript(scriptPath.Value)
		if err != nil {
			return newError("runScript failed: %s", err.Error())
		}
		return NewString(output)
	}},
	"runScriptStr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for runScriptStr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for runScriptStr must be SSH_CLIENT, got %s", args[0].Type())
		}
		scriptStr, ok := args[1].(*String)
		if !ok {
			return newError("argument for runScriptStr must be STRING, got %s", args[1].Type())
		}
		output, err := self.RunScriptStr(scriptStr.Value)
		if err != nil {
			return newError("runScriptStr failed: %s", err.Error())
		}
		return NewString(output)
	}},

	// File Operations
	"readFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for readFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for readFile must be SSH_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for readFile must be STRING, got %s", args[1].Type())
		}
		content, err := self.ReadFile(remotePath.Value)
		if err != nil {
			return newError("readFile failed: %s", err.Error())
		}
		return NewString(content)
	}},
	"writeFile": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for writeFile. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for writeFile must be SSH_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for writeFile must be STRING, got %s", args[1].Type())
		}
		content, ok := args[2].(*String)
		if !ok {
			return newError("second argument for writeFile must be STRING, got %s", args[2].Type())
		}
		if err := self.WriteFile(remotePath.Value, content.Value); err != nil {
			return newError("writeFile failed: %s", err.Error())
		}
		return NULL
	}},
	"upload": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for upload. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for upload must be SSH_CLIENT, got %s", args[0].Type())
		}
		localPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for upload must be STRING, got %s", args[1].Type())
		}
		remotePath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for upload must be STRING, got %s", args[2].Type())
		}
		if err := self.Upload(localPath.Value, remotePath.Value); err != nil {
			return newError("upload failed: %s", err.Error())
		}
		return NULL
	}},
	"download": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for download. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for download must be SSH_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for download must be STRING, got %s", args[1].Type())
		}
		localPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for download must be STRING, got %s", args[2].Type())
		}
		if err := self.Download(remotePath.Value, localPath.Value); err != nil {
			return newError("download failed: %s", err.Error())
		}
		return NULL
	}},
	"uploadDir": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for uploadDir. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for uploadDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		localDir, ok := args[1].(*String)
		if !ok {
			return newError("first argument for uploadDir must be STRING, got %s", args[1].Type())
		}
		remoteDir, ok := args[2].(*String)
		if !ok {
			return newError("second argument for uploadDir must be STRING, got %s", args[2].Type())
		}
		if err := self.UploadDir(localDir.Value, remoteDir.Value); err != nil {
			return newError("uploadDir failed: %s", err.Error())
		}
		return NULL
	}},
	"downloadDir": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for downloadDir. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for downloadDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		remoteDir, ok := args[1].(*String)
		if !ok {
			return newError("first argument for downloadDir must be STRING, got %s", args[1].Type())
		}
		localDir, ok := args[2].(*String)
		if !ok {
			return newError("second argument for downloadDir must be STRING, got %s", args[2].Type())
		}
		if err := self.DownloadDir(remoteDir.Value, localDir.Value); err != nil {
			return newError("downloadDir failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for mkdir must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdir must be STRING, got %s", args[1].Type())
		}
		if err := self.Mkdir(path.Value); err != nil {
			return newError("mkdir failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdirAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdirAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for mkdirAll must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdirAll must be STRING, got %s", args[1].Type())
		}
		if err := self.MkdirAll(path.Value); err != nil {
			return newError("mkdirAll failed: %s", err.Error())
		}
		return NULL
	}},
	"remove": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for remove. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for remove must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for remove must be STRING, got %s", args[1].Type())
		}
		if err := self.Remove(path.Value); err != nil {
			return newError("remove failed: %s", err.Error())
		}
		return NULL
	}},
	"removeDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for removeDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for removeDir must be STRING, got %s", args[1].Type())
		}
		if err := self.RemoveDir(path.Value); err != nil {
			return newError("removeDir failed: %s", err.Error())
		}
		return NULL
	}},
	"rename": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for rename. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for rename must be SSH_CLIENT, got %s", args[0].Type())
		}
		oldPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for rename must be STRING, got %s", args[1].Type())
		}
		newPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for rename must be STRING, got %s", args[2].Type())
		}
		if err := self.Rename(oldPath.Value, newPath.Value); err != nil {
			return newError("rename failed: %s", err.Error())
		}
		return NULL
	}},
	"stat": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for stat. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for stat must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for stat must be STRING, got %s", args[1].Type())
		}
		info, err := self.Stat(path.Value)
		if err != nil {
			return newError("stat failed: %s", err.Error())
		}
		pairs := make([]Object, 0, len(info))
		for k, v := range info {
			var valObj Object
			switch vv := v.(type) {
			case string:
				valObj = NewString(vv)
			case int:
				valObj = NewInt(int64(vv))
			case int64:
				valObj = NewInt(vv)
			case bool:
				valObj = &Bool{Value: vv}
			default:
				valObj = NewString(fmt.Sprintf("%v", vv))
			}
			pairs = append(pairs, &Array{Elements: []Object{NewString(k), valObj}})
		}
		return NewArray(pairs)
	}},
	"exists": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for exists. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for exists must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for exists must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.Exists(path.Value)}
	}},
	"isDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for isDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for isDir must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsDir(path.Value)}
	}},
	"isFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for isFile must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for isFile must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsFile(path.Value)}
	}},
	"listDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for listDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for listDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for listDir must be STRING, got %s", args[1].Type())
		}
		files, err := self.ListDir(path.Value)
		if err != nil {
			return newError("listDir failed: %s", err.Error())
		}
		elements := make([]Object, len(files))
		for i, file := range files {
			pairs := make([]Object, 0, len(file))
			for k, v := range file {
				var valObj Object
				switch vv := v.(type) {
				case string:
					valObj = NewString(vv)
				case int:
					valObj = NewInt(int64(vv))
				case int64:
					valObj = NewInt(vv)
				case bool:
					valObj = &Bool{Value: vv}
				default:
					valObj = NewString(fmt.Sprintf("%v", vv))
				}
				pairs = append(pairs, &Array{Elements: []Object{NewString(k), valObj}})
			}
			elements[i] = NewArray(pairs)
		}
		return NewArray(elements)
	}},
	"walkDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for walkDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SSHClient)
		if !ok {
			return newError("receiver for walkDir must be SSH_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for walkDir must be STRING, got %s", args[1].Type())
		}
		files, err := self.WalkDir(path.Value)
		if err != nil {
			return newError("walkDir failed: %s", err.Error())
		}
		elements := make([]Object, len(files))
		for i, file := range files {
			pairs := make([]Object, 0, len(file))
			for k, v := range file {
				var valObj Object
				switch vv := v.(type) {
				case string:
					valObj = NewString(vv)
				case bool:
					valObj = &Bool{Value: vv}
				default:
					valObj = NewString(fmt.Sprintf("%v", vv))
				}
				pairs = append(pairs, &Array{Elements: []Object{NewString(k), valObj}})
			}
			elements[i] = NewArray(pairs)
		}
		return NewArray(elements)
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

// ============================================================
// SocketAddr Methods
// ============================================================

var socketAddrMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"host": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for host. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SocketAddr)
		if !ok {
			return newError("receiver for host must be SocketAddr, got %s", args[0].Type())
		}
		return NewString(self.Host())
	}},
	"port": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for port. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SocketAddr)
		if !ok {
			return newError("receiver for port must be SocketAddr, got %s", args[0].Type())
		}
		return NewInt(int64(self.Port()))
	}},
}

// ============================================================
// TcpServer Methods
// ============================================================

var tcpServerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"setReuseAddr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setReuseAddr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for setReuseAddr must be TcpServer, got %s", args[0].Type())
		}
		reuse, ok := args[1].(*Bool)
		if !ok {
			return newError("argument for setReuseAddr must be BOOL, got %s", args[1].Type())
		}
		return self.SetReuseAddr(reuse.Value)
	}},
	"setTimeout": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTimeout. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for setTimeout must be TcpServer, got %s", args[0].Type())
		}
		ms, ok := args[1].(*Int)
		if !ok {
			return newError("argument for setTimeout must be INT, got %s", args[1].Type())
		}
		return self.SetTimeout(int(ms.Value))
	}},
	"onAccept": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for onAccept. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for onAccept must be TcpServer, got %s", args[0].Type())
		}
		fn, ok := args[1].(*Function)
		if !ok {
			return newError("argument for onAccept must be FUNCTION, got %s", args[1].Type())
		}
		// Create a callback wrapper that calls the Xxlang function
		callback := func(client *TcpClient) {
			// Call the function with the client as argument
			// Note: This requires access to the evaluator, which is handled at a higher level
			// For now, we store the function and the server will need to be used with Start()
			_ = fn // Placeholder - actual invocation happens via evaluator
		}
		return self.OnAccept(callback)
	}},
	"listen": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for listen. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for listen must be TcpServer, got %s", args[0].Type())
		}
		addr, ok := args[1].(*String)
		if !ok {
			return newError("argument for listen must be STRING, got %s", args[1].Type())
		}
		return self.Listen(addr.Value)
	}},
	"accept": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for accept. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for accept must be TcpServer, got %s", args[0].Type())
		}
		timeoutMs := 0
		if len(args) == 2 {
			ms, ok := args[1].(*Int)
			if !ok {
				return newError("timeout argument for accept must be INT, got %s", args[1].Type())
			}
			timeoutMs = int(ms.Value)
		}
		return self.Accept(timeoutMs)
	}},
	"start": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for start. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for start must be TcpServer, got %s", args[0].Type())
		}
		return self.Start()
	}},
	"addr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for addr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for addr must be TcpServer, got %s", args[0].Type())
		}
		a := self.Addr()
		if a == nil {
			return NULL
		}
		return a
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for close must be TcpServer, got %s", args[0].Type())
		}
		return self.Close()
	}},
	"isClosed": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isClosed. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpServer)
		if !ok {
			return newError("receiver for isClosed must be TcpServer, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsClosed()}
	}},
}

// ============================================================
// TcpClient Methods
// ============================================================

var tcpClientMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"connect": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for connect. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for connect must be TcpClient, got %s", args[0].Type())
		}
		addr, ok := args[1].(*String)
		if !ok {
			return newError("argument for connect must be STRING, got %s", args[1].Type())
		}
		return self.Connect(addr.Value)
	}},
	"setTimeout": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTimeout. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for setTimeout must be TcpClient, got %s", args[0].Type())
		}
		ms, ok := args[1].(*Int)
		if !ok {
			return newError("argument for setTimeout must be INT, got %s", args[1].Type())
		}
		return self.SetTimeout(int(ms.Value))
	}},
	"receive": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for receive. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for receive must be TcpClient, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for receive must be INT, got %s", args[1].Type())
		}
		return self.Receive(int(n.Value))
	}},
	"receiveLine": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for receiveLine. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for receiveLine must be TcpClient, got %s", args[0].Type())
		}
		return self.ReceiveLine()
	}},
	"receiveBytes": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for receiveBytes. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for receiveBytes must be TcpClient, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for receiveBytes must be INT, got %s", args[1].Type())
		}
		return self.ReceiveBytes(int(n.Value))
	}},
	"send": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for send. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for send must be TcpClient, got %s", args[0].Type())
		}
		data, ok := args[1].(*String)
		if !ok {
			return newError("argument for send must be STRING, got %s", args[1].Type())
		}
		return self.SendString(data.Value)
	}},
	"sendBytes": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sendBytes. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for sendBytes must be TcpClient, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("argument for sendBytes must be ARRAY, got %s", args[1].Type())
		}
		// Convert array of ints to byte slice
		data := make([]byte, len(arr.Elements))
		for i, elem := range arr.Elements {
			b, ok := elem.(*Int)
			if !ok {
				return newError("array elements for sendBytes must be INT, got %s at index %d", elem.Type(), i)
			}
			if b.Value < 0 || b.Value > 255 {
				return newError("array element at index %d out of byte range: %d", i, b.Value)
			}
			data[i] = byte(b.Value)
		}
		return self.SendBytes(data)
	}},
	"localAddr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for localAddr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for localAddr must be TcpClient, got %s", args[0].Type())
		}
		return self.LocalAddr()
	}},
	"remoteAddr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for remoteAddr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for remoteAddr must be TcpClient, got %s", args[0].Type())
		}
		return self.RemoteAddr()
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for close must be TcpClient, got %s", args[0].Type())
		}
		return self.Close()
	}},
	"isClosed": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isClosed. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TcpClient)
		if !ok {
			return newError("receiver for isClosed must be TcpClient, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsClosed()}
	}},
}

// ============================================================
// UdpSocket Methods
// ============================================================

var udpSocketMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"bind": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for bind. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for bind must be UdpSocket, got %s", args[0].Type())
		}
		addr, ok := args[1].(*String)
		if !ok {
			return newError("argument for bind must be STRING, got %s", args[1].Type())
		}
		return self.Bind(addr.Value)
	}},
	"setTimeout": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTimeout. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for setTimeout must be UdpSocket, got %s", args[0].Type())
		}
		ms, ok := args[1].(*Int)
		if !ok {
			return newError("argument for setTimeout must be INT, got %s", args[1].Type())
		}
		return self.SetTimeout(int(ms.Value))
	}},
	"sendTo": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for sendTo. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for sendTo must be UdpSocket, got %s", args[0].Type())
		}
		data, ok := args[1].(*String)
		if !ok {
			return newError("data argument for sendTo must be STRING, got %s", args[1].Type())
		}
		addr, ok := args[2].(*String)
		if !ok {
			return newError("address argument for sendTo must be STRING, got %s", args[2].Type())
		}
		return self.SendTo(data.Value, addr.Value)
	}},
	"sendToBytes": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for sendToBytes. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for sendToBytes must be UdpSocket, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("data argument for sendToBytes must be ARRAY, got %s", args[1].Type())
		}
		addr, ok := args[2].(*String)
		if !ok {
			return newError("address argument for sendToBytes must be STRING, got %s", args[2].Type())
		}
		// Convert array of ints to byte slice
		data := make([]byte, len(arr.Elements))
		for i, elem := range arr.Elements {
			b, ok := elem.(*Int)
			if !ok {
				return newError("array elements for sendToBytes must be INT, got %s at index %d", elem.Type(), i)
			}
			if b.Value < 0 || b.Value > 255 {
				return newError("array element at index %d out of byte range: %d", i, b.Value)
			}
			data[i] = byte(b.Value)
		}
		return self.SendToBytes(data, addr.Value)
	}},
	"receiveFrom": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for receiveFrom. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for receiveFrom must be UdpSocket, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("buffer size argument for receiveFrom must be INT, got %s", args[1].Type())
		}
		// The method returns two values, but we only return the first one (data)
		// The second value (sender address) is ignored for simplicity
		// A more complete implementation would return an array with both values
		data, _ := self.ReceiveFrom(int(n.Value))
		return data
	}},
	"localAddr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for localAddr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for localAddr must be UdpSocket, got %s", args[0].Type())
		}
		return self.LocalAddr()
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for close must be UdpSocket, got %s", args[0].Type())
		}
		return self.Close()
	}},
	"isClosed": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isClosed. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*UdpSocket)
		if !ok {
			return newError("receiver for isClosed must be UdpSocket, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsClosed()}
	}},
}

// ============================================================
// FTP Client Methods
// ============================================================

var ftpClientMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"isConnected": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isConnected. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for isConnected must be FTP_CLIENT, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsConnected()}
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for close must be FTP_CLIENT, got %s", args[0].Type())
		}
		if err := self.Close(); err != nil {
			return newError("close failed: %s", err.Error())
		}
		return NULL
	}},
	"getHost": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getHost. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for getHost must be FTP_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetHost())
	}},
	"getPort": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPort. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for getPort must be FTP_CLIENT, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetPort()))
	}},
	"getUser": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getUser. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for getUser must be FTP_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetUser())
	}},
	"upload": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for upload. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for upload must be FTP_CLIENT, got %s", args[0].Type())
		}
		localPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for upload must be STRING, got %s", args[1].Type())
		}
		remotePath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for upload must be STRING, got %s", args[2].Type())
		}
		if err := self.Upload(localPath.Value, remotePath.Value); err != nil {
			return newError("upload failed: %s", err.Error())
		}
		return NULL
	}},
	"download": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for download. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for download must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for download must be STRING, got %s", args[1].Type())
		}
		localPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for download must be STRING, got %s", args[2].Type())
		}
		if err := self.Download(remotePath.Value, localPath.Value); err != nil {
			return newError("download failed: %s", err.Error())
		}
		return NULL
	}},
	"delete": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for delete must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for delete must be STRING, got %s", args[1].Type())
		}
		if err := self.Delete(remotePath.Value); err != nil {
			return newError("delete failed: %s", err.Error())
		}
		return NULL
	}},
	"rename": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for rename. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for rename must be FTP_CLIENT, got %s", args[0].Type())
		}
		oldPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for rename must be STRING, got %s", args[1].Type())
		}
		newPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for rename must be STRING, got %s", args[2].Type())
		}
		if err := self.Rename(oldPath.Value, newPath.Value); err != nil {
			return newError("rename failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for mkdir must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdir must be STRING, got %s", args[1].Type())
		}
		if err := self.Mkdir(remotePath.Value); err != nil {
			return newError("mkdir failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdirAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdirAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for mkdirAll must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdirAll must be STRING, got %s", args[1].Type())
		}
		if err := self.MkdirAll(remotePath.Value); err != nil {
			return newError("mkdirAll failed: %s", err.Error())
		}
		return NULL
	}},
	"rmdir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for rmdir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for rmdir must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for rmdir must be STRING, got %s", args[1].Type())
		}
		if err := self.Rmdir(remotePath.Value); err != nil {
			return newError("rmdir failed: %s", err.Error())
		}
		return NULL
	}},
	"rmdirAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for rmdirAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for rmdirAll must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for rmdirAll must be STRING, got %s", args[1].Type())
		}
		if err := self.RmdirAll(remotePath.Value); err != nil {
			return newError("rmdirAll failed: %s", err.Error())
		}
		return NULL
	}},
	"listDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for listDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for listDir must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for listDir must be STRING, got %s", args[1].Type())
		}
		files, err := self.ListDir(remotePath.Value)
		if err != nil {
			return newError("listDir failed: %s", err.Error())
		}
		elements := make([]Object, len(files))
		for i, file := range files {
			m := NewMapWithCapacity(3)
			nameKey := NewString("name")
			sizeKey := NewString("size")
			isDirKey := NewString("isDir")
			m.Pairs[nameKey.HashKey()] = MapPair{Key: nameKey, Value: NewString(file.Name)}
			m.Pairs[sizeKey.HashKey()] = MapPair{Key: sizeKey, Value: NewInt(file.Size)}
			m.Pairs[isDirKey.HashKey()] = MapPair{Key: isDirKey, Value: &Bool{Value: file.IsDir}}
			elements[i] = m
		}
		return NewArray(elements)
	}},
	"changeDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for changeDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for changeDir must be FTP_CLIENT, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for changeDir must be STRING, got %s", args[1].Type())
		}
		if err := self.ChangeDir(path.Value); err != nil {
			return newError("changeDir failed: %s", err.Error())
		}
		return NULL
	}},
	"currentDir": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for currentDir. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for currentDir must be FTP_CLIENT, got %s", args[0].Type())
		}
		dir, err := self.CurrentDir()
		if err != nil {
			return newError("currentDir failed: %s", err.Error())
		}
		return NewString(dir)
	}},
	"size": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for size. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for size must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for size must be STRING, got %s", args[1].Type())
		}
		size, err := self.Size(remotePath.Value)
		if err != nil {
			return newError("size failed: %s", err.Error())
		}
		return NewInt(size)
	}},
	"exists": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for exists. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for exists must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for exists must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.Exists(remotePath.Value)}
	}},
	"isDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for isDir must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for isDir must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsDir(remotePath.Value)}
	}},
	"isFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for isFile must be FTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for isFile must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsFile(remotePath.Value)}
	}},
	"setType": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setType. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for setType must be FTP_CLIENT, got %s", args[0].Type())
		}
		transferType, ok := args[1].(*String)
		if !ok {
			return newError("argument for setType must be STRING, got %s", args[1].Type())
		}
		if err := self.SetType(transferType.Value); err != nil {
			return newError("setType failed: %s", err.Error())
		}
		return NULL
	}},
	"setPassive": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setPassive. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpClient)
		if !ok {
			return newError("receiver for setPassive must be FTP_CLIENT, got %s", args[0].Type())
		}
		enabled, ok := args[1].(*Bool)
		if !ok {
			return newError("argument for setPassive must be BOOL, got %s", args[1].Type())
		}
		if err := self.SetPassive(enabled.Value); err != nil {
			return newError("setPassive failed: %s", err.Error())
		}
		return NULL
	}},
}

// ============================================================
// FTP Server Methods
// ============================================================

var ftpServerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"start": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for start. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpServer)
		if !ok {
			return newError("receiver for start must be FTP_SERVER, got %s", args[0].Type())
		}
		if err := self.Start(); err != nil {
			return newError("start failed: %s", err.Error())
		}
		return NULL
	}},
	"stop": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for stop. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpServer)
		if !ok {
			return newError("receiver for stop must be FTP_SERVER, got %s", args[0].Type())
		}
		if err := self.Stop(); err != nil {
			return newError("stop failed: %s", err.Error())
		}
		return NULL
	}},
	"addUser": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for addUser. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*FtpServer)
		if !ok {
			return newError("receiver for addUser must be FTP_SERVER, got %s", args[0].Type())
		}
		username, ok := args[1].(*String)
		if !ok {
			return newError("first argument for addUser must be STRING, got %s", args[1].Type())
		}
		password, ok := args[2].(*String)
		if !ok {
			return newError("second argument for addUser must be STRING, got %s", args[2].Type())
		}
		homeDir, ok := args[3].(*String)
		if !ok {
			return newError("third argument for addUser must be STRING, got %s", args[3].Type())
		}
		if err := self.AddUser(username.Value, password.Value, homeDir.Value); err != nil {
			return newError("addUser failed: %s", err.Error())
		}
		return NULL
	}},
	"removeUser": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeUser. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*FtpServer)
		if !ok {
			return newError("receiver for removeUser must be FTP_SERVER, got %s", args[0].Type())
		}
		username, ok := args[1].(*String)
		if !ok {
			return newError("argument for removeUser must be STRING, got %s", args[1].Type())
		}
		if err := self.RemoveUser(username.Value); err != nil {
			return newError("removeUser failed: %s", err.Error())
		}
		return NULL
	}},
	"isRunning": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isRunning. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*FtpServer)
		if !ok {
			return newError("receiver for isRunning must be FTP_SERVER, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsRunning()}
	}},
}

// ============================================================
// SFTP Client Methods
// ============================================================

var sftpClientMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"isConnected": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isConnected. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for isConnected must be SFTP_CLIENT, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsConnected()}
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for close must be SFTP_CLIENT, got %s", args[0].Type())
		}
		if err := self.Close(); err != nil {
			return newError("close failed: %s", err.Error())
		}
		return NULL
	}},
	"getHost": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getHost. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for getHost must be SFTP_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetHost())
	}},
	"getPort": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPort. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for getPort must be SFTP_CLIENT, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetPort()))
	}},
	"getUser": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getUser. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for getUser must be SFTP_CLIENT, got %s", args[0].Type())
		}
		return NewString(self.GetUser())
	}},
	"upload": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for upload. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for upload must be SFTP_CLIENT, got %s", args[0].Type())
		}
		localPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for upload must be STRING, got %s", args[1].Type())
		}
		remotePath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for upload must be STRING, got %s", args[2].Type())
		}
		if err := self.Upload(localPath.Value, remotePath.Value); err != nil {
			return newError("upload failed: %s", err.Error())
		}
		return NULL
	}},
	"download": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for download. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for download must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for download must be STRING, got %s", args[1].Type())
		}
		localPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for download must be STRING, got %s", args[2].Type())
		}
		if err := self.Download(remotePath.Value, localPath.Value); err != nil {
			return newError("download failed: %s", err.Error())
		}
		return NULL
	}},
	"delete": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for delete must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for delete must be STRING, got %s", args[1].Type())
		}
		if err := self.Delete(remotePath.Value); err != nil {
			return newError("delete failed: %s", err.Error())
		}
		return NULL
	}},
	"rename": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for rename. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for rename must be SFTP_CLIENT, got %s", args[0].Type())
		}
		oldPath, ok := args[1].(*String)
		if !ok {
			return newError("first argument for rename must be STRING, got %s", args[1].Type())
		}
		newPath, ok := args[2].(*String)
		if !ok {
			return newError("second argument for rename must be STRING, got %s", args[2].Type())
		}
		if err := self.Rename(oldPath.Value, newPath.Value); err != nil {
			return newError("rename failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for mkdir must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdir must be STRING, got %s", args[1].Type())
		}
		if err := self.Mkdir(remotePath.Value); err != nil {
			return newError("mkdir failed: %s", err.Error())
		}
		return NULL
	}},
	"mkdirAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for mkdirAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for mkdirAll must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for mkdirAll must be STRING, got %s", args[1].Type())
		}
		if err := self.MkdirAll(remotePath.Value); err != nil {
			return newError("mkdirAll failed: %s", err.Error())
		}
		return NULL
	}},
	"rmdir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for rmdir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for rmdir must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for rmdir must be STRING, got %s", args[1].Type())
		}
		if err := self.Rmdir(remotePath.Value); err != nil {
			return newError("rmdir failed: %s", err.Error())
		}
		return NULL
	}},
	"rmdirAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for rmdirAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for rmdirAll must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for rmdirAll must be STRING, got %s", args[1].Type())
		}
		if err := self.RmdirAll(remotePath.Value); err != nil {
			return newError("rmdirAll failed: %s", err.Error())
		}
		return NULL
	}},
	"listDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for listDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for listDir must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for listDir must be STRING, got %s", args[1].Type())
		}
		files, err := self.ListDir(remotePath.Value)
		if err != nil {
			return newError("listDir failed: %s", err.Error())
		}
		elements := make([]Object, len(files))
		for i, file := range files {
			m := NewMapWithCapacity(3)
			nameKey := NewString("name")
			sizeKey := NewString("size")
			isDirKey := NewString("isDir")
			m.Pairs[nameKey.HashKey()] = MapPair{Key: nameKey, Value: NewString(file.Name)}
			m.Pairs[sizeKey.HashKey()] = MapPair{Key: sizeKey, Value: NewInt(file.Size)}
			m.Pairs[isDirKey.HashKey()] = MapPair{Key: isDirKey, Value: &Bool{Value: file.IsDir}}
			elements[i] = m
		}
		return NewArray(elements)
	}},
	"stat": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for stat. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for stat must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for stat must be STRING, got %s", args[1].Type())
		}
		info, err := self.Stat(remotePath.Value)
		if err != nil {
			return newError("stat failed: %s", err.Error())
		}
		m := NewMapWithCapacity(3)
		nameKey := NewString("name")
		sizeKey := NewString("size")
		isDirKey := NewString("isDir")
		m.Pairs[nameKey.HashKey()] = MapPair{Key: nameKey, Value: NewString(remotePath.Value)}
		m.Pairs[sizeKey.HashKey()] = MapPair{Key: sizeKey, Value: NewInt(info.Size)}
		m.Pairs[isDirKey.HashKey()] = MapPair{Key: isDirKey, Value: &Bool{Value: info.IsDir}}
		return m
	}},
	"exists": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for exists. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for exists must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for exists must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.Exists(remotePath.Value)}
	}},
	"isDir": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isDir. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for isDir must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for isDir must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsDir(remotePath.Value)}
	}},
	"isFile": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for isFile. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpClient)
		if !ok {
			return newError("receiver for isFile must be SFTP_CLIENT, got %s", args[0].Type())
		}
		remotePath, ok := args[1].(*String)
		if !ok {
			return newError("argument for isFile must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.IsFile(remotePath.Value)}
	}},
}

// ============================================================
// SFTP Server Methods
// ============================================================

var sftpServerMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"start": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for start. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for start must be SFTP_SERVER, got %s", args[0].Type())
		}
		if err := self.Start(); err != nil {
			return newError("start failed: %s", err.Error())
		}
		return NULL
	}},
	"stop": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for stop. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for stop must be SFTP_SERVER, got %s", args[0].Type())
		}
		if err := self.Stop(); err != nil {
			return newError("stop failed: %s", err.Error())
		}
		return NULL
	}},
	"addUser": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for addUser. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for addUser must be SFTP_SERVER, got %s", args[0].Type())
		}
		username, ok := args[1].(*String)
		if !ok {
			return newError("first argument for addUser must be STRING, got %s", args[1].Type())
		}
		password, ok := args[2].(*String)
		if !ok {
			return newError("second argument for addUser must be STRING, got %s", args[2].Type())
		}
		homeDir, ok := args[3].(*String)
		if !ok {
			return newError("third argument for addUser must be STRING, got %s", args[3].Type())
		}
		if err := self.AddUser(username.Value, password.Value, homeDir.Value); err != nil {
			return newError("addUser failed: %s", err.Error())
		}
		return NULL
	}},
	"addUserWithKey": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for addUserWithKey. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for addUserWithKey must be SFTP_SERVER, got %s", args[0].Type())
		}
		username, ok := args[1].(*String)
		if !ok {
			return newError("first argument for addUserWithKey must be STRING, got %s", args[1].Type())
		}
		keyStr, ok := args[2].(*String)
		if !ok {
			return newError("second argument for addUserWithKey must be STRING, got %s", args[2].Type())
		}
		homeDir, ok := args[3].(*String)
		if !ok {
			return newError("third argument for addUserWithKey must be STRING, got %s", args[3].Type())
		}
		if err := self.AddUserWithKey(username.Value, keyStr.Value, homeDir.Value); err != nil {
			return newError("addUserWithKey failed: %s", err.Error())
		}
		return NULL
	}},
	"removeUser": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeUser. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for removeUser must be SFTP_SERVER, got %s", args[0].Type())
		}
		username, ok := args[1].(*String)
		if !ok {
			return newError("argument for removeUser must be STRING, got %s", args[1].Type())
		}
		if err := self.RemoveUser(username.Value); err != nil {
			return newError("removeUser failed: %s", err.Error())
		}
		return NULL
	}},
	"isRunning": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isRunning. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*SftpServer)
		if !ok {
			return newError("receiver for isRunning must be SFTP_SERVER, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsRunning()}
	}},
}

// ============================================================
// HTML Document Methods
// ============================================================

var htmlDocumentMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"root": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for root. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for root must be HTMLDocument, got %s", args[0].Type())
		}
		root := self.Root()
		if root == nil {
			return NULL
		}
		return root
	}},
	"head": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for head. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for head must be HTMLDocument, got %s", args[0].Type())
		}
		head := self.Head()
		if head == nil {
			return NULL
		}
		return head
	}},
	"body": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for body. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for body must be HTMLDocument, got %s", args[0].Type())
		}
		body := self.Body()
		if body == nil {
			return NULL
		}
		return body
	}},
	"title": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for title. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for title must be HTMLDocument, got %s", args[0].Type())
		}
		return NewString(self.Title())
	}},
	"setTitle": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTitle. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for setTitle must be HTMLDocument, got %s", args[0].Type())
		}
		title, ok := args[1].(*String)
		if !ok {
			return newError("title must be STRING")
		}
		self.SetTitle(title.Value)
		return NULL
	}},
	"docType": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for docType. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for docType must be HTMLDocument, got %s", args[0].Type())
		}
		return NewString(self.DocType())
	}},
	"getElementById": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getElementById. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for getElementById must be HTMLDocument, got %s", args[0].Type())
		}
		id, ok := args[1].(*String)
		if !ok {
			return newError("id must be STRING")
		}
		elem := self.GetElementById(id.Value)
		if elem == nil {
			return NULL
		}
		return elem
	}},
	"getElementsByTagName": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getElementsByTagName. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for getElementsByTagName must be HTMLDocument, got %s", args[0].Type())
		}
		tag, ok := args[1].(*String)
		if !ok {
			return newError("tag must be STRING")
		}
		return self.GetElementsByTagName(tag.Value)
	}},
	"getElementsByClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getElementsByClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for getElementsByClass must be HTMLDocument, got %s", args[0].Type())
		}
		className, ok := args[1].(*String)
		if !ok {
			return newError("className must be STRING")
		}
		return self.GetElementsByClassName(className.Value)
	}},
	"querySelector": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelector. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for querySelector must be HTMLDocument, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		elem := self.QuerySelector(selector.Value)
		if elem == nil {
			return NULL
		}
		return elem
	}},
	"querySelectorAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelectorAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for querySelectorAll must be HTMLDocument, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		return self.QuerySelectorAll(selector.Value)
	}},
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for find must be HTMLDocument, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		return self.Find(selector.Value)
	}},
	"findFirst": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findFirst. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for findFirst must be HTMLDocument, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		elem := self.FindFirst(selector.Value)
		if elem == nil {
			return NULL
		}
		return elem
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for toString must be HTMLDocument, got %s", args[0].Type())
		}
		return NewString(self.ToString())
	}},
	"toIndented": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toIndented. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for toIndented must be HTMLDocument, got %s", args[0].Type())
		}
		return NewString(self.ToIndented())
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for save must be HTMLDocument, got %s", args[0].Type())
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
	"setMeta": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setMeta. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for setMeta must be HTMLDocument, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		content, ok := args[2].(*String)
		if !ok {
			return newError("content must be STRING")
		}
		self.SetMeta(name.Value, content.Value)
		return NULL
	}},
	"addStyle": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addStyle. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for addStyle must be HTMLDocument, got %s", args[0].Type())
		}
		css, ok := args[1].(*String)
		if !ok {
			return newError("css must be STRING")
		}
		self.AddStyle(css.Value)
		return NULL
	}},
	"addScript": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for addScript. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for addScript must be HTMLDocument, got %s", args[0].Type())
		}
		js, ok := args[1].(*String)
		if !ok {
			return newError("js must be STRING")
		}
		src := ""
		if len(args) == 3 {
			srcStr, ok := args[2].(*String)
			if !ok {
				return newError("src must be STRING")
			}
			src = srcStr.Value
		}
		self.AddScript(js.Value, src)
		return NULL
	}},
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLDocument)
		if !ok {
			return newError("receiver for toMap must be HTMLDocument, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
}

// ============================================================
// HTML Element Methods
// ============================================================

var htmlElementMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"tagName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for tagName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for tagName must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.TagName())
	}},
	"setTagName": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTagName. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setTagName must be HTMLElement, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		self.SetTagName(name.Value)
		return NULL
	}},
	"text": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for text. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for text must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.TextContent())
	}},
	"setText": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setText. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setText must be HTMLElement, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		self.SetTextContent(text.Value)
		return NULL
	}},
	"html": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for html. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for html must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.InnerHTML())
	}},
	"setHtml": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setHtml. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setHtml must be HTMLElement, got %s", args[0].Type())
		}
		html, ok := args[1].(*String)
		if !ok {
			return newError("html must be STRING")
		}
		if err := self.SetInnerHTML(html.Value); err != nil {
			return newError("setHtml failed: %s", err.Error())
		}
		return NULL
	}},
	"outerHtml": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for outerHtml. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for outerHtml must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.OuterHTML())
	}},
	"attr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for attr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for attr must be HTMLElement, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		return NewString(self.Attribute(name.Value))
	}},
	"setAttr": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setAttr. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setAttr must be HTMLElement, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		value, ok := args[2].(*String)
		if !ok {
			return newError("value must be STRING")
		}
		self.SetAttribute(name.Value, value.Value)
		return NULL
	}},
	"hasAttr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasAttr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for hasAttr must be HTMLElement, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		return &Bool{Value: self.HasAttribute(name.Value)}
	}},
	"removeAttr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeAttr. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for removeAttr must be HTMLElement, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		self.RemoveAttribute(name.Value)
		return NULL
	}},
	"attrs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for attrs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for attrs must be HTMLElement, got %s", args[0].Type())
		}
		return self.Attributes()
	}},
	"id": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for id. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for id must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.ID())
	}},
	"setId": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setId. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setId must be HTMLElement, got %s", args[0].Type())
		}
		id, ok := args[1].(*String)
		if !ok {
			return newError("id must be STRING")
		}
		self.SetID(id.Value)
		return NULL
	}},
	"class": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for class. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for class must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.Class())
	}},
	"setClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for setClass must be HTMLElement, got %s", args[0].Type())
		}
		class, ok := args[1].(*String)
		if !ok {
			return newError("class must be STRING")
		}
		self.SetClass(class.Value)
		return NULL
	}},
	"addClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for addClass must be HTMLElement, got %s", args[0].Type())
		}
		className, ok := args[1].(*String)
		if !ok {
			return newError("className must be STRING")
		}
		self.AddClass(className.Value)
		return NULL
	}},
	"removeClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for removeClass must be HTMLElement, got %s", args[0].Type())
		}
		className, ok := args[1].(*String)
		if !ok {
			return newError("className must be STRING")
		}
		self.RemoveClass(className.Value)
		return NULL
	}},
	"hasClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for hasClass must be HTMLElement, got %s", args[0].Type())
		}
		className, ok := args[1].(*String)
		if !ok {
			return newError("className must be STRING")
		}
		return &Bool{Value: self.HasClass(className.Value)}
	}},
	"toggleClass": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for toggleClass. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for toggleClass must be HTMLElement, got %s", args[0].Type())
		}
		className, ok := args[1].(*String)
		if !ok {
			return newError("className must be STRING")
		}
		self.ToggleClass(className.Value)
		return NULL
	}},
	"children": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for children. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for children must be HTMLElement, got %s", args[0].Type())
		}
		return self.Children()
	}},
	"childCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for childCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for childCount must be HTMLElement, got %s", args[0].Type())
		}
		return NewInt(int64(self.ChildCount()))
	}},
	"firstChild": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for firstChild. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for firstChild must be HTMLElement, got %s", args[0].Type())
		}
		child := self.FirstChild()
		if child == nil {
			return NULL
		}
		return child
	}},
	"lastChild": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lastChild. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for lastChild must be HTMLElement, got %s", args[0].Type())
		}
		child := self.LastChild()
		if child == nil {
			return NULL
		}
		return child
	}},
	"parent": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for parent. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for parent must be HTMLElement, got %s", args[0].Type())
		}
		parent := self.Parent()
		if parent == nil {
			return NULL
		}
		return parent
	}},
	"appendChild": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for appendChild. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for appendChild must be HTMLElement, got %s", args[0].Type())
		}
		child, ok := args[1].(*HTMLElement)
		if !ok {
			return newError("child must be HTMLElement")
		}
		self.AppendChild(child)
		return NULL
	}},
	"removeChild": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for removeChild. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for removeChild must be HTMLElement, got %s", args[0].Type())
		}
		index, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT")
		}
		if !self.RemoveChild(int(index.Value)) {
			return newError("removeChild failed: invalid index")
		}
		return NULL
	}},
	"insertBefore": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for insertBefore. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for insertBefore must be HTMLElement, got %s", args[0].Type())
		}
		newElem, ok := args[1].(*HTMLElement)
		if !ok {
			return newError("newElem must be HTMLElement")
		}
		refElem, ok := args[2].(*HTMLElement)
		if !ok {
			return newError("refElem must be HTMLElement")
		}
		if !self.InsertBefore(newElem, refElem) {
			return newError("insertBefore failed: reference element not found")
		}
		return NULL
	}},
	"insertAfter": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for insertAfter. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for insertAfter must be HTMLElement, got %s", args[0].Type())
		}
		newElem, ok := args[1].(*HTMLElement)
		if !ok {
			return newError("newElem must be HTMLElement")
		}
		refElem, ok := args[2].(*HTMLElement)
		if !ok {
			return newError("refElem must be HTMLElement")
		}
		if !self.InsertAfter(newElem, refElem) {
			return newError("insertAfter failed: reference element not found")
		}
		return NULL
	}},
	"replaceChild": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for replaceChild. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for replaceChild must be HTMLElement, got %s", args[0].Type())
		}
		newElem, ok := args[1].(*HTMLElement)
		if !ok {
			return newError("newElem must be HTMLElement")
		}
		oldElem, ok := args[2].(*HTMLElement)
		if !ok {
			return newError("oldElem must be HTMLElement")
		}
		if !self.ReplaceChild(newElem, oldElem) {
			return newError("replaceChild failed: old element not found")
		}
		return NULL
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for clear must be HTMLElement, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"remove": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for remove. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for remove must be HTMLElement, got %s", args[0].Type())
		}
		self.Remove()
		return NULL
	}},
	"clone": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clone. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for clone must be HTMLElement, got %s", args[0].Type())
		}
		return self.Clone()
	}},
	"querySelector": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelector. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for querySelector must be HTMLElement, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		elem := self.QuerySelector(selector.Value)
		if elem == nil {
			return NULL
		}
		return elem
	}},
	"querySelectorAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelectorAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for querySelectorAll must be HTMLElement, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		return self.QuerySelectorAll(selector.Value)
	}},
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for find must be HTMLElement, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		return self.Find(selector.Value)
	}},
	"findFirst": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findFirst. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for findFirst must be HTMLElement, got %s", args[0].Type())
		}
		selector, ok := args[1].(*String)
		if !ok {
			return newError("selector must be STRING")
		}
		elem := self.FindFirst(selector.Value)
		if elem == nil {
			return NULL
		}
		return elem
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for toString must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.ToString())
	}},
	"toIndented": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toIndented. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for toIndented must be HTMLElement, got %s", args[0].Type())
		}
		return NewString(self.ToIndented())
	}},
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HTMLElement)
		if !ok {
			return newError("receiver for toMap must be HTMLElement, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
}

// ============================================================
// TOML Document Methods
// ============================================================

var tomlDocumentMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"get": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for get. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for get must be TomlDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING, got %s", args[1].Type())
		}
		val := self.Get(path.Value)
		if val == nil {
			return NULL
		}
		return val.ToXxlangObject()
	}},
	"set": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for set. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for set must be TomlDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING, got %s", args[1].Type())
		}
		value := FromXxlangObject(args[2])
		self.Set(path.Value, value)
		return NULL
	}},
	"remove": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for remove. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for remove must be TomlDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING, got %s", args[1].Type())
		}
		if self.Remove(path.Value) {
			return TRUE
		}
		return FALSE
	}},
	"has": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for has. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for has must be TomlDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING, got %s", args[1].Type())
		}
		if self.Has(path.Value) {
			return TRUE
		}
		return FALSE
	}},
	"keys": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for keys. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for keys must be TomlDocument, got %s", args[0].Type())
		}
		keys := self.Keys()
		elements := make([]Object, len(keys))
		for i, k := range keys {
			elements[i] = NewString(k)
		}
		return NewArray(elements)
	}},
	"sections": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for sections. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for sections must be TomlDocument, got %s", args[0].Type())
		}
		sections := self.Sections()
		elements := make([]Object, len(sections))
		for i, s := range sections {
			elements[i] = NewString(s)
		}
		return NewArray(elements)
	}},
	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for toMap must be TomlDocument, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for toString must be TomlDocument, got %s", args[0].Type())
		}
		return NewString(self.ToString())
	}},
	"toIndented": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toIndented. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for toIndented must be TomlDocument, got %s", args[0].Type())
		}
		return NewString(self.ToIndented())
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for save must be TomlDocument, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING, got %s", args[1].Type())
		}
		if err := self.Save(path.Value); err != nil {
			return newError("save failed: %v", err)
		}
		return NULL
	}},
	"merge": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for merge. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*TomlDocument)
		if !ok {
			return newError("receiver for merge must be TomlDocument, got %s", args[0].Type())
		}
		other, ok := args[1].(*TomlDocument)
		if !ok {
			return newError("other must be TomlDocument, got %s", args[1].Type())
		}
		if err := self.Merge(other); err != nil {
			return newError("merge failed: %v", err)
		}
		return NULL
	}},
}

// ============================================================
// Time Methods
// ============================================================

var timeMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toStr. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for toStr must be Time, got %s", args[0].Type())
		}
		return NewString(self.Format("2006-01-02 15:04:05"))
	}},

	// Getter methods
	"year": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for year. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for year must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetYear()))
	}},

	"month": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for month. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for month must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetMonth()))
	}},

	"day": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for day. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for day must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetDay()))
	}},

	"hour": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hour. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for hour must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetHour()))
	}},

	"minute": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for minute. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for minute must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetMinute()))
	}},

	"second": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for second. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for second must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetSecond()))
	}},

	"nanosecond": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for nanosecond. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for nanosecond must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetNanosecond()))
	}},

	"weekday": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for weekday. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for weekday must be Time, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetWeekday()))
	}},

	"timestamp": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for timestamp. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for timestamp must be Time, got %s", args[0].Type())
		}
		return NewInt(self.GetTimestamp())
	}},

	"timestampMs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for timestampMs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for timestampMs must be Time, got %s", args[0].Type())
		}
		return NewInt(self.GetTimestampMs())
	}},

	// Formatting
	"format": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for format. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for format must be Time, got %s", args[0].Type())
		}
		layout := "2006-01-02 15:04:05"
		if len(args) == 2 {
			l, ok := args[1].(*String)
			if !ok {
				return newError("layout must be STRING, got %s", args[1].Type())
			}
			layout = l.Value
		}
		return NewString(self.Format(layout))
	}},

	// Time operations
	"addSecs": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addSecs. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for addSecs must be Time, got %s", args[0].Type())
		}
		var secs float64
		switch arg := args[1].(type) {
		case *Int:
			secs = float64(arg.Value)
		case *Float:
			secs = arg.Value
		default:
			return newError("seconds must be INT or FLOAT, got %s", args[1].Type())
		}
		return self.AddSecs(secs)
	}},

	"addDate": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for addDate. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for addDate must be Time, got %s", args[0].Type())
		}
		years, ok := args[1].(*Int)
		if !ok {
			return newError("years must be INT, got %s", args[1].Type())
		}
		months, ok := args[2].(*Int)
		if !ok {
			return newError("months must be INT, got %s", args[2].Type())
		}
		days, ok := args[3].(*Int)
		if !ok {
			return newError("days must be INT, got %s", args[3].Type())
		}
		return self.AddDate(int(years.Value), int(months.Value), int(days.Value))
	}},

	"addDuration": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for addDuration. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for addDuration must be Time, got %s", args[0].Type())
		}
		durStr, ok := args[1].(*String)
		if !ok {
			return newError("duration must be STRING, got %s", args[1].Type())
		}
		result, err := self.AddDuration(durStr.Value)
		if err != nil {
			return newError("addDuration failed: %v", err)
		}
		return result
	}},

	// Comparison methods
	"before": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for before. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for before must be Time, got %s", args[0].Type())
		}
		other, ok := args[1].(*Time)
		if !ok {
			return newError("other must be Time, got %s", args[1].Type())
		}
		return &Bool{Value: self.Before(other)}
	}},

	"after": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for after. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for after must be Time, got %s", args[0].Type())
		}
		other, ok := args[1].(*Time)
		if !ok {
			return newError("other must be Time, got %s", args[1].Type())
		}
		return &Bool{Value: self.After(other)}
	}},

	"equal": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for equal. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for equal must be Time, got %s", args[0].Type())
		}
		other, ok := args[1].(*Time)
		if !ok {
			return newError("other must be Time, got %s", args[1].Type())
		}
		return &Bool{Value: self.Equal(other)}
	}},

	"diffSecs": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for diffSecs. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for diffSecs must be Time, got %s", args[0].Type())
		}
		other, ok := args[1].(*Time)
		if !ok {
			return newError("other must be Time, got %s", args[1].Type())
		}
		return NewFloat(self.DiffSecs(other))
	}},

	// Utility methods
	"isZero": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isZero. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for isZero must be Time, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsZero()}
	}},

	"toMap": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toMap. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Time)
		if !ok {
			return newError("receiver for toMap must be Time, got %s", args[0].Type())
		}
		return self.ToMap()
	}},
}

// ============================================================
// RodBrowser Methods (Web Scraping with Rod)
// ============================================================

var rodBrowserMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// Navigation
	"get": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for get. got=%d, want=2 (self + url)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for get must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Get(args[1:]...)
	}},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for close must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Close(args[1:]...)
	}},
	"refresh": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for refresh. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for refresh must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Refresh(args[1:]...)
	}},
	"back": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for back. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for back must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Back(args[1:]...)
	}},
	"forward": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for forward. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for forward must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Forward(args[1:]...)
	}},
	// Element operations
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for find must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Find(args[1:]...)
	}},
	"findAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findAll. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for findAll must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.FindAll(args[1:]...)
	}},
	"exists": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for exists. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for exists must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Exists(args[1:]...)
	}},
	"click": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for click. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for click must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Click(args[1:]...)
	}},
	"fill": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for fill. got=%d, want=3 (self + selector + value)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for fill must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Fill(args[1:]...)
	}},
	"wait": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for wait. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for wait must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Wait(args[1:]...)
	}},
	"waitLoad": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for waitLoad. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for waitLoad must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.WaitLoad(args[1:]...)
	}},
	"waitStable": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for waitStable. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for waitStable must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.WaitStable(args[1:]...)
	}},
	"fullscreen": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for fullscreen. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for fullscreen must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Fullscreen(args[1:]...)
	}},
	// JavaScript execution
	"eval": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for eval. got=%d, want=2 (self + js)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for eval must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Eval(args[1:]...)
	}},
	"inject": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for inject. got=%d, want=2 (self + js)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for inject must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Inject(args[1:]...)
	}},
	// Page content
	"html": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for html. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for html must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.HTML(args[1:]...)
	}},
	"text": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for text. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for text must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Text(args[1:]...)
	}},
	// Storage (localStorage, sessionStorage, cookies)
	"getLocalStorage": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getLocalStorage. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for getLocalStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.GetLocalStorage(args[1:]...)
	}},
	"getSessionStorage": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSessionStorage. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for getSessionStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.GetSessionStorage(args[1:]...)
	}},
	"setLocalStorage": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setLocalStorage. got=%d, want=3 (self + key + value)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for setLocalStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SetLocalStorage(args[1:]...)
	}},
	"setSessionStorage": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSessionStorage. got=%d, want=3 (self + key + value)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for setSessionStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SetSessionStorage(args[1:]...)
	}},
	"getCookies": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getCookies. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for getCookies must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.GetCookies(args[1:]...)
	}},
	"setCookies": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setCookies. got=%d, want=2 (self + cookies)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for setCookies must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SetCookies(args[1:]...)
	}},
	"clearCookies": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clearCookies. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for clearCookies must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.ClearCookies(args[1:]...)
	}},
	"saveStorage": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for saveStorage. got=%d, want=2 (self + path)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for saveStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SaveStorage(args[1:]...)
	}},
	"loadStorage": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for loadStorage. got=%d, want=2 (self + path)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for loadStorage must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.LoadStorage(args[1:]...)
	}},
	// Screenshot
	"screenshot": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for screenshot. got=%d, want=2 (self + path)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for screenshot must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.Screenshot(args[1:]...)
	}},
	// Viewport and User Agent
	"setViewport": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setViewport. got=%d, want=3 (self + width + height)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for setViewport must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SetViewport(args[1:]...)
	}},
	"setUserAgent": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setUserAgent. got=%d, want=2 (self + ua)", len(args))
		}
		self, ok := args[0].(*RodBrowser)
		if !ok {
			return newError("receiver for setUserAgent must be ROD_BROWSER, got %s", args[0].Type())
		}
		return self.SetUserAgent(args[1:]...)
	}},
}

// RodHTMLElementMethods - methods for RodHTMLElement
var rodHTMLElementMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	// Content extraction
	"getText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getText must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetText(args[1:]...)
	}},
	"getAttr": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getAttr. got=%d, want=2 (self + name)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getAttr must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetAttr(args[1:]...)
	}},
	"getProperty": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getProperty. got=%d, want=2 (self + name)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getProperty must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetProperty(args[1:]...)
	}},
	"getInnerHTML": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getInnerHTML. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getInnerHTML must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetInnerHTML(args[1:]...)
	}},
	"getOuterHTML": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getOuterHTML. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getOuterHTML must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetOuterHTML(args[1:]...)
	}},
	"getTagName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTagName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getTagName must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetTagName(args[1:]...)
	}},
	"getValue": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getValue. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getValue must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetValue(args[1:]...)
	}},
	// DOM traversal
	"find": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for find. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for find must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Find(args[1:]...)
	}},
	"findAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for findAll. got=%d, want=2 (self + selector)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for findAll must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.FindAll(args[1:]...)
	}},
	// Interaction
	"click": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for click. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for click must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Click(args[1:]...)
	}},
	"input": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for input. got=%d, want=2 (self + value)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for input must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Input(args[1:]...)
	}},
	"typeText": {Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for typeText. got=%d, want>=2 (self + text)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for typeText must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.TypeText(args[1:]...)
	}},
	"setValue": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setValue. got=%d, want=2 (self + value)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for setValue must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.SetValue(args[1:]...)
	}},
	"select": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for select. got=%d, want=2 (self + value)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for select must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Select(args[1:]...)
	}},
	"check": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for check. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for check must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Check(args[1:]...)
	}},
	"uncheck": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for uncheck. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for uncheck must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Uncheck(args[1:]...)
	}},
	"focus": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for focus. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for focus must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Focus(args[1:]...)
	}},
	"blur": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for blur. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for blur must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Blur(args[1:]...)
	}},
	"hover": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hover. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for hover must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Hover(args[1:]...)
	}},
	"press": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for press. got=%d, want=2 (self + key)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for press must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Press(args[1:]...)
	}},
	"selectAll": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for selectAll. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for selectAll must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.SelectAll(args[1:]...)
	}},
	"drag": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for drag. got=%d, want=3 (self + x + y)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for drag must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Drag(args[1:]...)
	}},
	// State checking
	"isVisible": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isVisible. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for isVisible must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.IsVisible(args[1:]...)
	}},
	"isEnabled": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEnabled. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for isEnabled must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.IsEnabled(args[1:]...)
	}},
	// Wait for state
	"waitFor": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for waitFor. got=%d, want=2 (self + state)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for waitFor must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.WaitFor(args[1:]...)
	}},
	// Screenshot
	"screenshot": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for screenshot. got=%d, want=2 (self + path)", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for screenshot must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.Screenshot(args[1:]...)
	}},
	// Position and size
	"getBoundingClientRect": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getBoundingClientRect. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*RodHTMLElement)
		if !ok {
			return newError("receiver for getBoundingClientRect must be ROD_HTML_ELEMENT, got %s", args[0].Type())
		}
		return self.GetBoundingClientRect(args[1:]...)
	}},
}

// ============================================================
// BackupTask Methods
// ============================================================

var backupTaskMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	// Configuration methods
	"setSourceLocal": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setSourceLocal. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setSourceLocal must be BACKUP_TASK, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for setSourceLocal must be STRING, got %s", args[1].Type())
		}
		self.SetSourceLocal(path.Value)
		return NULL
	}},
	"setTargetLocal": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTargetLocal. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setTargetLocal must be BACKUP_TASK, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("argument for setTargetLocal must be STRING, got %s", args[1].Type())
		}
		self.SetTargetLocal(path.Value)
		return NULL
	}},
	"setSourceRemote": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSourceRemote. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver must be BACKUP_TASK, got %s", args[0].Type())
		}
		client, ok := args[1].(*SSHClient)
		if !ok {
			return newError("second argument must be SSH_CLIENT, got %s", args[1].Type())
		}
		path, ok := args[2].(*String)
		if !ok {
			return newError("third argument must be STRING, got %s", args[2].Type())
		}
		self.SetSourceRemote(client, path.Value)
		return NULL
	}},
	"setTargetRemote": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setTargetRemote. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver must be BACKUP_TASK, got %s", args[0].Type())
		}
		client, ok := args[1].(*SSHClient)
		if !ok {
			return newError("second argument must be SSH_CLIENT, got %s", args[1].Type())
		}
		path, ok := args[2].(*String)
		if !ok {
			return newError("third argument must be STRING, got %s", args[2].Type())
		}
		self.SetTargetRemote(client, path.Value)
		return NULL
	}},
	"setMode": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setMode. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setMode must be BACKUP_TASK, got %s", args[0].Type())
		}
		mode, ok := args[1].(*String)
		if !ok {
			return newError("argument for setMode must be STRING, got %s", args[1].Type())
		}
		self.SetMode(mode.Value)
		return NULL
	}},
	"setCompareStrategy": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setCompareStrategy. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setCompareStrategy must be BACKUP_TASK, got %s", args[0].Type())
		}
		strategy, ok := args[1].(*String)
		if !ok {
			return newError("argument for setCompareStrategy must be STRING, got %s", args[1].Type())
		}
		self.SetCompareStrategy(strategy.Value)
		return NULL
	}},
	"setConflictPolicy": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setConflictPolicy. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setConflictPolicy must be BACKUP_TASK, got %s", args[0].Type())
		}
		policy, ok := args[1].(*String)
		if !ok {
			return newError("argument for setConflictPolicy must be STRING, got %s", args[1].Type())
		}
		self.SetConflictPolicy(policy.Value)
		return NULL
	}},
	"setExclude": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setExclude. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setExclude must be BACKUP_TASK, got %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("argument for setExclude must be ARRAY, got %s", args[1].Type())
		}
		// Convert array to string slice
		patterns := make([]string, 0, len(arr.Elements))
		for _, elem := range arr.Elements {
			if str, ok := elem.(*String); ok {
				patterns = append(patterns, str.Value)
			}
		}
		self.SetExcludePatterns(patterns)
		return NULL
	}},
	"setDryRun": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setDryRun. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for setDryRun must be BACKUP_TASK, got %s", args[0].Type())
		}
		_ = self // avoid unused variable error
		// Note: BackupTask currently does not have a dry-run field
		// This method is a placeholder for future implementation
		// The argument is accepted but not used yet
		return NULL
	}},
	// Execution methods
	"execute": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for execute. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for execute must be BACKUP_TASK, got %s", args[0].Type())
		}
		result := self.Run()
		return result
	}},
	"checkConflicts": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for checkConflicts. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupTask)
		if !ok {
			return newError("receiver for checkConflicts must be BACKUP_TASK, got %s", args[0].Type())
		}
		conflicts := self.CheckConflicts()
		// Convert string slice to array
		elements := make([]Object, 0, len(conflicts))
		for _, c := range conflicts {
			elements = append(elements, NewString(c))
		}
		return &Array{Elements: elements}
	}},
}

// ============================================================
// BackupResult Methods
// ============================================================

var backupResultMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	// Status methods
	"isSuccess": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isSuccess. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for isSuccess must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return &Bool{Value: self.Success}
	}},
	"hasConflicts": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hasConflicts. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for hasConflicts must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return &Bool{Value: self.HasConflicts()}
	}},
	"hasErrors": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for hasErrors. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for hasErrors must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return &Bool{Value: self.HasErrors()}
	}},
	"summary": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for summary. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for summary must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewString(self.Summary())
	}},

	// Data getter methods
	"getFilesCopied": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFilesCopied. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getFilesCopied must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewInt(int64(self.FilesCopied))
	}},
	"getFilesSkipped": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFilesSkipped. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getFilesSkipped must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewInt(int64(self.FilesSkipped))
	}},
	"getFilesDeleted": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFilesDeleted. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getFilesDeleted must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewInt(int64(self.FilesDeleted))
	}},
	"getBytesTransferred": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getBytesTransferred. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getBytesTransferred must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewInt(self.BytesTransferred)
	}},
	"getDuration": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getDuration. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getDuration must be BACKUP_RESULT, got %s", args[0].Type())
		}
		return NewFloat(self.Duration)
	}},
	"getErrors": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getErrors. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getErrors must be BACKUP_RESULT, got %s", args[0].Type())
		}
		elements := make([]Object, len(self.Errors))
		for i, e := range self.Errors {
			elements[i] = NewString(e)
		}
		return &Array{Elements: elements}
	}},
	"getConflicts": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getConflicts. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*BackupResult)
		if !ok {
			return newError("receiver for getConflicts must be BACKUP_RESULT, got %s", args[0].Type())
		}
		elements := make([]Object, len(self.Conflicts))
		for i, c := range self.Conflicts {
			elements[i] = NewString(c)
		}
		return &Array{Elements: elements}
	}},
}

// ============================================================
// HLBR Browser Methods (Lightweight Headless Browser)
// ============================================================

var hlbrBrowserMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"navigate": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for navigate. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for navigate must be HLBR_BROWSER, got %s", args[0].Type())
		}
		url, ok := args[1].(*String)
		if !ok {
			return newError("argument for navigate must be STRING, got %s", args[1].Type())
		}
		if err := self.browser.Navigate(url.Value); err != nil {
			return newError("navigate failed: %s", err.Error())
		}
		return self
	}},

	"getTitle": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTitle. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for getTitle must be HLBR_BROWSER, got %s", args[0].Type())
		}
		return NewString(self.browser.GetTitle())
	}},

	"getHTML": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getHTML. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for getHTML must be HLBR_BROWSER, got %s", args[0].Type())
		}
		return NewString(self.browser.GetHTML())
	}},

	"getText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for getText must be HLBR_BROWSER, got %s", args[0].Type())
		}
		return NewString(self.browser.GetText())
	}},

	"getURL": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getURL. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for getURL must be HLBR_BROWSER, got %s", args[0].Type())
		}
		return NewString(self.browser.GetURL())
	}},

	"querySelector": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelector. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for querySelector must be HLBR_BROWSER, got %s", args[0].Type())
		}
		sel, ok := args[1].(*String)
		if !ok {
			return newError("argument for querySelector must be STRING, got %s", args[1].Type())
		}
		node := self.browser.QuerySelector(sel.Value)
		if node == nil {
			return NULL
		}
		return NewHlbrNode(node)
	}},

	"querySelectorAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelectorAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for querySelectorAll must be HLBR_BROWSER, got %s", args[0].Type())
		}
		sel, ok := args[1].(*String)
		if !ok {
			return newError("argument for querySelectorAll must be STRING, got %s", args[1].Type())
		}
		nodes := self.browser.QuerySelectorAll(sel.Value)
		return hlbrNodesToXxArray(nodes)
	}},

	"evaluate": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for evaluate. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for evaluate must be HLBR_BROWSER, got %s", args[0].Type())
		}
		code, ok := args[1].(*String)
		if !ok {
			return newError("argument for evaluate must be STRING, got %s", args[1].Type())
		}
		result, err := self.browser.Evaluate(code.Value)
		if err != nil {
			return newError("evaluate failed: %s", err.Error())
		}
		return hlbrGoValueToObject(result)
	}},

	"screenshotText": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for screenshotText. got=%d, want=1-2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for screenshotText must be HLBR_BROWSER, got %s", args[0].Type())
		}
		width := 80
		if len(args) == 2 {
			if w, ok := args[1].(*Int); ok {
				width = int(w.Value)
			} else if w, ok := args[1].(*Float); ok {
				width = int(w.Value)
			}
		}
		return NewString(self.browser.ScreenshotText(width))
	}},

	"screenshotTextToFile": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for screenshotTextToFile. got=%d, want=2-3", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for screenshotTextToFile must be HLBR_BROWSER, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path argument for screenshotTextToFile must be STRING, got %s", args[1].Type())
		}
		width := 80
		if len(args) == 3 {
			if w, ok := args[2].(*Int); ok {
				width = int(w.Value)
			} else if w, ok := args[2].(*Float); ok {
				width = int(w.Value)
			}
		}
		if err := self.browser.ScreenshotTextToFile(path.Value, width); err != nil {
			return newError("screenshotTextToFile failed: %s", err.Error())
		}
		return NULL
	}},

	"setUserAgent": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setUserAgent. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for setUserAgent must be HLBR_BROWSER, got %s", args[0].Type())
		}
		ua, ok := args[1].(*String)
		if !ok {
			return newError("argument for setUserAgent must be STRING, got %s", args[1].Type())
		}
		self.browser.SetUserAgent(ua.Value)
		return NULL
	}},

	"setHeader": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setHeader. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for setHeader must be HLBR_BROWSER, got %s", args[0].Type())
		}
		key, ok := args[1].(*String)
		if !ok {
			return newError("key argument for setHeader must be STRING, got %s", args[1].Type())
		}
		val, ok := args[2].(*String)
		if !ok {
			return newError("value argument for setHeader must be STRING, got %s", args[2].Type())
		}
		self.browser.SetHeader(key.Value, val.Value)
		return NULL
	}},

	"getCookies": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getCookies. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for getCookies must be HLBR_BROWSER, got %s", args[0].Type())
		}
		return hlbrCookiesToXxArray(self.browser.GetCookies())
	}},

	"history": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for history. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for history must be HLBR_BROWSER, got %s", args[0].Type())
		}
		urls := self.browser.History()
		elems := make([]Object, len(urls))
		for i, u := range urls {
			elems[i] = NewString(u)
		}
		return &Array{Elements: elems}
	}},

	"back": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for back. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for back must be HLBR_BROWSER, got %s", args[0].Type())
		}
		if err := self.browser.Back(); err != nil {
			return newError("back failed: %s", err.Error())
		}
		return self
	}},

	"httpGet": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for httpGet. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for httpGet must be HLBR_BROWSER, got %s", args[0].Type())
		}
		url, ok := args[1].(*String)
		if !ok {
			return newError("argument for httpGet must be STRING, got %s", args[1].Type())
		}
		client := self.browser.Client()
		resp, err := client.Get(url.Value)
		if err != nil {
			return newError("httpGet failed: %s", err.Error())
		}
		pairs := make(map[HashKey]MapPair)
		scKey := NewString("statusCode")
		pairs[scKey.HashKey()] = MapPair{Key: scKey, Value: NewInt(int64(resp.StatusCode))}
		bodyKey := NewString("body")
		pairs[bodyKey.HashKey()] = MapPair{Key: bodyKey, Value: NewString(resp.Body)}
		urlKey := NewString("url")
		pairs[urlKey.HashKey()] = MapPair{Key: urlKey, Value: NewString(resp.URL)}
		return &Map{Pairs: pairs}
	}},

	"httpPost": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 4 {
			return newError("wrong number of arguments for httpPost. got=%d, want=2-4", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for httpPost must be HLBR_BROWSER, got %s", args[0].Type())
		}
		url, ok := args[1].(*String)
		if !ok {
			return newError("url argument for httpPost must be STRING, got %s", args[1].Type())
		}
		contentType := "application/x-www-form-urlencoded"
		body := ""
		if len(args) >= 3 {
			if b, ok := args[2].(*String); ok {
				body = b.Value
			}
		}
		if len(args) >= 4 {
			if ct, ok := args[3].(*String); ok {
				contentType = ct.Value
			}
		}
		client := self.browser.Client()
		resp, err := client.Post(url.Value, contentType, body)
		if err != nil {
			return newError("httpPost failed: %s", err.Error())
		}
		pairs := make(map[HashKey]MapPair)
		scKey := NewString("statusCode")
		pairs[scKey.HashKey()] = MapPair{Key: scKey, Value: NewInt(int64(resp.StatusCode))}
		bodyKey := NewString("body")
		pairs[bodyKey.HashKey()] = MapPair{Key: bodyKey, Value: NewString(resp.Body)}
		return &Map{Pairs: pairs}
	}},

	"httpPostForm": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for httpPostForm. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HlbrBrowser)
		if !ok {
			return newError("receiver for httpPostForm must be HLBR_BROWSER, got %s", args[0].Type())
		}
		url, ok := args[1].(*String)
		if !ok {
			return newError("url argument for httpPostForm must be STRING, got %s", args[1].Type())
		}
		dataMap, ok := args[2].(*Map)
		if !ok {
			return newError("data argument for httpPostForm must be MAP, got %s", args[2].Type())
		}
		data := make(map[string]string)
		for _, pair := range dataMap.Pairs {
			key := pair.Key.Inspect()
			val := ""
			if s, ok := pair.Value.(*String); ok {
				val = s.Value
			} else {
				val = pair.Value.Inspect()
			}
			data[key] = val
		}
		client := self.browser.Client()
		resp, err := client.PostForm(url.Value, data)
		if err != nil {
			return newError("httpPostForm failed: %s", err.Error())
		}
		pairs := make(map[HashKey]MapPair)
		scKey := NewString("statusCode")
		pairs[scKey.HashKey()] = MapPair{Key: scKey, Value: NewInt(int64(resp.StatusCode))}
		bodyKey := NewString("body")
		pairs[bodyKey.HashKey()] = MapPair{Key: bodyKey, Value: NewString(resp.Body)}
		return &Map{Pairs: pairs}
	}},
}

// ============================================================
// HLBR Node Methods (DOM Element)
// ============================================================

var hlbrNodeMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},

	"getText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getText must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.TextContent())
	}},

	"getHTML": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getHTML. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getHTML must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.InnerHTML())
	}},

	"getOuterHTML": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getOuterHTML. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getOuterHTML must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.OuterHTML())
	}},

	"getTagName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTagName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getTagName must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.TagName())
	}},

	"getAttribute": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getAttribute. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getAttribute must be HLBR_NODE, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name argument for getAttribute must be STRING, got %s", args[1].Type())
		}
		val := self.node.GetAttribute(name.Value)
		return NewString(val)
	}},

	"setAttribute": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setAttribute. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for setAttribute must be HLBR_NODE, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name argument for setAttribute must be STRING, got %s", args[1].Type())
		}
		val, ok := args[2].(*String)
		if !ok {
			return newError("value argument for setAttribute must be STRING, got %s", args[2].Type())
		}
		self.node.SetAttribute(name.Value, val.Value)
		return NULL
	}},

	"hasAttribute": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasAttribute. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for hasAttribute must be HLBR_NODE, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name argument for hasAttribute must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: self.node.HasAttribute(name.Value)}
	}},

	"getID": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getID. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getID must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.ID())
	}},

	"getClassName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getClassName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getClassName must be HLBR_NODE, got %s", args[0].Type())
		}
		return NewString(self.node.ClassName())
	}},

	"querySelector": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelector. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for querySelector must be HLBR_NODE, got %s", args[0].Type())
		}
		sel, ok := args[1].(*String)
		if !ok {
			return newError("argument for querySelector must be STRING, got %s", args[1].Type())
		}
		node := dom.QuerySelector(self.node, sel.Value)
		if node == nil {
			return NULL
		}
		return NewHlbrNode(node)
	}},

	"querySelectorAll": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for querySelectorAll. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for querySelectorAll must be HLBR_NODE, got %s", args[0].Type())
		}
		sel, ok := args[1].(*String)
		if !ok {
			return newError("argument for querySelectorAll must be STRING, got %s", args[1].Type())
		}
		nodes := dom.QuerySelectorAll(self.node, sel.Value)
		return hlbrNodesToXxArray(nodes)
	}},

	"getChildren": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getChildren. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getChildren must be HLBR_NODE, got %s", args[0].Type())
		}
		elems := make([]Object, len(self.node.Children))
		for i, child := range self.node.Children {
			elems[i] = NewHlbrNode(child)
		}
		return &Array{Elements: elems}
	}},

	"getParent": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getParent. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*HlbrNode)
		if !ok {
			return newError("receiver for getParent must be HLBR_NODE, got %s", args[0].Type())
		}
		if self.node.Parent == nil {
			return NULL
		}
		return NewHlbrNode(self.node.Parent)
	}},
}
