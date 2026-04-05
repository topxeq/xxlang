// pkg/objects/object.go
package objects

import "fmt"

// ObjectType represents the type of an object
type ObjectType string

// Object types
const (
	NullType             ObjectType = "NULL"
	IntType              ObjectType = "INT"
	FloatType            ObjectType = "FLOAT"
	StringType           ObjectType = "STRING"
	CharsType            ObjectType = "CHARS"
	BoolType             ObjectType = "BOOL"
	ArrayType            ObjectType = "ARRAY"
	MapType              ObjectType = "MAP"
	FunctionType         ObjectType = "FUNCTION"
	BuiltinType          ObjectType = "BUILTIN"
	BytesType            ObjectType = "BYTES"
	ClassType            ObjectType = "CLASS"
	InstanceType         ObjectType = "INSTANCE"
	ErrorType            ObjectType = "ERROR"
	ReturnType           ObjectType = "RETURN"
	ClosureType          ObjectType = "CLOSURE"
	ModuleType           ObjectType = "MODULE"
	StringBuilderType    ObjectType = "STRING_BUILDER"
	BytesBufferType      ObjectType = "BYTES_BUFFER"
	BigIntType           ObjectType = "BIGINT"
	BigFloatType         ObjectType = "BIGFLOAT"
	HttpReqType          ObjectType = "HTTP_REQ"
	HttpRespType         ObjectType = "HTTP_RESP"
	HttpMuxType          ObjectType = "HTTP_MUX"
	WebSocketType        ObjectType = "WEBSOCKET"
	TubeType             ObjectType = "TUBE"
	MutexType            ObjectType = "MUTEX"
	RWMutexType          ObjectType = "RWMUTEX"
	WaitGroupType        ObjectType = "WAITGROUP"
	OnceType             ObjectType = "ONCE"
	CondType             ObjectType = "COND"
	AtomicIntType        ObjectType = "ATOMICINT"
	GoroutineType        ObjectType = "GOROUTINE"
	ContextType          ObjectType = "CONTEXT"
	FileUploadType       ObjectType = "FILE_UPLOAD"
	FileUploadResultType ObjectType = "FILE_UPLOAD_RESULT"
	FileType             ObjectType = "FILE"
	FileInfoType         ObjectType = "FILE_INFO"
	ReaderType           ObjectType = "READER"
	WriterType           ObjectType = "WRITER"
	ScannerType          ObjectType = "SCANNER"
	DBType               ObjectType = "DB"
	DBTxType             ObjectType = "DB_TX"
	DBRowsType           ObjectType = "DB_ROWS"
	DBStmtType           ObjectType = "DB_STMT"
	OrderedMapType       ObjectType = "ORDERED_MAP"
	QueueType            ObjectType = "QUEUE"
	SetType              ObjectType = "SET"
	XLSXType             ObjectType = "XLSX"
	XMLDocumentType      ObjectType = "XML_DOCUMENT"
	XMLNodeType          ObjectType = "XML_NODE"
	DocxDocumentType     ObjectType = "DOCX_DOCUMENT"
	DocxParagraphType    ObjectType = "DOCX_PARAGRAPH"
	DocxRunType          ObjectType = "DOCX_RUN"
	DocxTableType        ObjectType = "DOCX_TABLE"
	DocxTableRowType     ObjectType = "DOCX_TABLE_ROW"
	DocxTableCellType    ObjectType = "DOCX_TABLE_CELL"
	DocxImageType        ObjectType = "DOCX_IMAGE"
	DocxSectionType      ObjectType = "DOCX_SECTION"
	DocxHeaderType       ObjectType = "DOCX_HEADER"
	DocxFooterType       ObjectType = "DOCX_FOOTER"
	DocxStyleType        ObjectType = "DOCX_STYLE"
	DocxHyperlinkType    ObjectType = "DOCX_HYPERLINK"
	DocxBookmarkType     ObjectType = "DOCX_BOOKMARK"
	DocxTOCType          ObjectType = "DOCX_TOC"
	DocxTextBoxType      ObjectType = "DOCX_TEXT_BOX"
	DocxShapeType        ObjectType = "DOCX_SHAPE"
	DocxChartType        ObjectType = "DOCX_CHART"
	DocxCommentType      ObjectType = "DOCX_COMMENT"
	DocxRevisionType     ObjectType = "DOCX_REVISION"
	DocxFootnoteType     ObjectType = "DOCX_FOOTNOTE"
	DocxEndnoteType      ObjectType = "DOCX_ENDNOTE"
	// PPTX types
	PPTXDocumentType    ObjectType = "PPTX_DOCUMENT"
	PPTXSlideType       ObjectType = "PPTX_SLIDE"
	PPTXTextFrameType   ObjectType = "PPTX_TEXT_FRAME"
	PPTXParagraphType   ObjectType = "PPTX_PARAGRAPH"
	PPTXTextRunType     ObjectType = "PPTX_TEXT_RUN"
	PPTXShapeType       ObjectType = "PPTX_SHAPE"
	PPTXTableType       ObjectType = "PPTX_TABLE"
	PPTXTableCellType   ObjectType = "PPTX_TABLE_CELL"
	PPTXChartType       ObjectType = "PPTX_CHART"
	PPTXChartSeriesType ObjectType = "PPTX_CHART_SERIES"
	PPTXImageType       ObjectType = "PPTX_IMAGE"
	PPTXVideoType       ObjectType = "PPTX_VIDEO"
	PPTXAudioType       ObjectType = "PPTX_AUDIO"
	PPTXSlideLayoutType ObjectType = "PPTX_SLIDE_LAYOUT"
	PPTXSlideMasterType ObjectType = "PPTX_SLIDE_MASTER"
	PPTXThemeType       ObjectType = "PPTX_THEME"
	// PDF types
	PDFType         ObjectType = "PDF"
	PDFDocumentType ObjectType = "PDF_DOCUMENT"
	PDFPageType     ObjectType = "PDF_PAGE"
	PDFInfoType     ObjectType = "PDF_INFO"
	// Socket types
	SocketAddrType ObjectType = "SOCKET_ADDR"
	TcpServerType  ObjectType = "TCP_SERVER"
	TcpClientType  ObjectType = "TCP_CLIENT"
	UdpSocketType  ObjectType = "UDP_SOCKET"
	// LineEditor type
	LineEditorType ObjectType = "LINE_EDITOR"
	// SSH types
	SSHClientType ObjectType = "SSH_CLIENT"
	// FTP/SFTP types
	FtpClientType  ObjectType = "FTP_CLIENT"
	FtpServerType  ObjectType = "FTP_SERVER"
	SftpClientType ObjectType = "SFTP_CLIENT"
	SftpServerType ObjectType = "SFTP_SERVER"
	// HTML types
	HTMLDocumentType ObjectType = "HTML_DOCUMENT"
	HTMLElementType  ObjectType = "HTML_ELEMENT"
	// YAML types
	YAMLDocumentType ObjectType = "YAML_DOCUMENT"
	// TOML types
	TomlDocumentType ObjectType = "TOML_DOCUMENT"
	// Time type
	TimeType ObjectType = "TIME"
	// Rod browser types (for web scraping)
	RodBrowserType      ObjectType = "ROD_BROWSER"
	RodHTMLElementType  ObjectType = "ROD_HTML_ELEMENT"
)

// TypeTag is a fast integer type identifier for hot path checks
type TypeTag uint8

// Type tags for fast type checking (must match ObjectType order)
const (
	TagNull TypeTag = iota
	TagInt
	TagFloat
	TagString
	TagChars
	TagBool
	TagArray
	TagMap
	TagFunction
	TagBuiltin
	TagBytes
	TagClass
	TagInstance
	TagError
	TagReturn
	TagClosure
	TagModule
	TagStringBuilder
	TagBytesBuffer
	TagBigInt
	TagBigFloat
	TagHttpReq
	TagHttpResp
	TagHttpMux
	TagWebSocket
	TagTube
	TagMutex
	TagRWMutex
	TagWaitGroup
	TagOnce
	TagCond
	TagAtomicInt
	TagGoroutine
	TagContext
	TagFileUpload
	TagFileUploadResult
	TagFile
	TagFileInfo
	TagReader
	TagWriter
	TagScanner
	TagDB
	TagDBTx
	TagDBRows
	TagDBStmt
	TagOrderedMap
	TagQueue
	TagSet
	TagXLSX
	TagXMLDocument
	TagXMLNode
	TagDocxDocument
	TagDocxParagraph
	TagDocxRun
	TagDocxTable
	TagDocxTableRow
	TagDocxTableCell
	TagDocxImage
	TagDocxSection
	TagDocxHeader
	TagDocxFooter
	TagDocxStyle
	TagDocxHyperlink
	TagDocxBookmark
	TagDocxTOC
	TagDocxTextBox
	TagDocxShape
	TagDocxChart
	TagDocxComment
	TagDocxRevision
	TagDocxFootnote
	TagDocxEndnote
	// PPTX tags
	TagPPTXDocument
	TagPPTXSlide
	TagPPTXTextFrame
	TagPPTXParagraph
	TagPPTXTextRun
	TagPPTXShape
	TagPPTXTable
	TagPPTXTableCell
	TagPPTXChart
	TagPPTXChartSeries
	TagPPTXImage
	TagPPTXVideo
	TagPPTXAudio
	TagPPTXSlideLayout
	TagPPTXSlideMaster
	TagPPTXTheme
	// PDF tags
	TagPDF
	TagPDFDocument
	TagPDFPage
	TagPDFInfo
	// Socket tags
	TagSocketAddr
	TagTcpServer
	TagTcpClient
	TagUdpSocket
	// LineEditor tag
	TagLineEditor
	// SSH tag
	TagSSHClient
	// FTP/SFTP tags
	TagFtpClient
	TagFtpServer
	TagSftpClient
	TagSftpServer
	// HTML tags
	TagHTMLDocument
	TagHTMLElement
	// YAML tags
	TagYAMLDocument
	// Time tag
	TagTime
	// Rod browser tags
	TagRodBrowser
	TagRodHTMLElement
	TagUnknown
)

// Object is the base interface for all values in Xxlang
type Object interface {
	Type() ObjectType
	TypeTag() TypeTag // Fast type check without string comparison
	Inspect() string
	ToBool() *Bool
	HashKey() HashKey
}

// HashKey is used for map keys
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Null represents the null value
type Null struct{}

func (n *Null) Type() ObjectType { return NullType }
func (n *Null) TypeTag() TypeTag { return TagNull }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) ToBool() *Bool    { return FALSE }
func (n *Null) HashKey() HashKey { return HashKey{Type: NullType, Value: 0} }

// NULL is the singleton null value
var NULL = &Null{}

// Bool represents a boolean value
type Bool struct {
	Value bool
}

func (b *Bool) Type() ObjectType { return BoolType }
func (b *Bool) TypeTag() TypeTag { return TagBool }
func (b *Bool) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (b *Bool) ToBool() *Bool { return b }
func (b *Bool) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: BoolType, Value: value}
}

// TRUE and FALSE are singleton boolean values
var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
)

// Error represents a runtime error
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ErrorType }
func (e *Error) TypeTag() TypeTag { return TagError }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }
func (e *Error) ToBool() *Bool    { return FALSE }
func (e *Error) HashKey() HashKey { return HashKey{Type: ErrorType, Value: 0} }

// Return represents a return value (used internally)
type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return ReturnType }
func (r *Return) TypeTag() TypeTag { return TagReturn }
func (r *Return) Inspect() string  { return r.Value.Inspect() }
func (r *Return) ToBool() *Bool    { return r.Value.ToBool() }
func (r *Return) HashKey() HashKey { return HashKey{Type: ReturnType, Value: 0} }

// IsTruthy checks if an object is truthy
func IsTruthy(obj Object) bool {
	if obj == NULL {
		return false
	}
	return obj.ToBool().Value
}
