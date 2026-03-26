// pkg/objects/pdf.go
// PDF object types for PDF processing in Xxlang.
// Implemented using only Go standard library - no third-party dependencies.
package objects

import (
	"bytes"
	"compress/zlib"
	"encoding/ascii85"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
)

// ============================================================
// PDF Object Types
// ============================================================

// PDF represents an opened PDF file for reading and manipulation.
// It holds the parsed PDF structure and provides methods for
// extracting text, getting page info, and performing operations.
type PDF struct {
	mu sync.Mutex

	// Source information
	FilePath string    // Original file path (may be empty if from bytes)
	Source   []byte    // Raw PDF data
	Modified time.Time // Last modification time

	// Parsed structure
	Version    string            // PDF version (e.g., "1.4")
	Objects    map[int64]*PDFObj // Parsed PDF objects by object number
	XRefTable  *PDFXRefTable     // Cross-reference table
	Trailer    *PDFDict          // Trailer dictionary
	RootObj    int64             // Root object number (Catalog)
	InfoObj    int64             // Info object number (metadata)
	PageCount  int               // Number of pages
	Pages      []*PDFPage        // Cached page objects

	// State
	IsOpen     bool              // Whether the PDF is currently open
	IsModified bool              // Whether modifications have been made
}

// PDFDocument represents a new PDF document being created.
// Use this to create PDFs from scratch with text and pages.
type PDFDocument struct {
	mu sync.Mutex

	// Document properties
	Version    string            // PDF version (default "1.4")
	Title      string
	Author     string
	Subject    string
	Creator    string
	Producer   string

	// Pages
	Pages      []*PDFPageData    // Page data for new document

	// Object tracking
	NextObjNum int64             // Next object number to assign
	Objects    []interface{}     // Objects to write

	// Font settings
	DefaultFont string           // Default font name
	FontSize    float64          // Default font size
}

// PDFPageData represents a page in a PDFDocument being created.
type PDFPageData struct {
	Width      float64
	Height     float64
	Contents   *strings.Builder // Content stream
	Resources  map[string]interface{}
	Rotation   int
}

// PDFPage represents a page in an existing PDF document.
// It provides methods for extracting content and getting page properties.
type PDFPage struct {
	mu sync.Mutex

	// Parent reference
	PDF        *PDF              // Reference to parent PDF
	PageNum    int               // Page number (0-indexed)

	// Page properties from PDF
	Width      float64           // Page width in points
	Height     float64           // Page height in points
	Rotation   int               // Page rotation (0, 90, 180, 270)
	MediaBox   *PDFArray         // MediaBox array
	CropBox    *PDFArray         // CropBox array (optional)
	Contents   []int64           // Content stream object numbers
	Resources  *PDFDict          // Page resources dictionary

	// Cached content
	Text       string            // Cached extracted text
	Parsed     bool              // Whether content has been parsed
}

// PDFInfo contains metadata information about a PDF document.
type PDFInfo struct {
	Title        string
	Author       string
	Subject      string
	Keywords     string
	Creator      string
	Producer     string
	CreationDate string
	ModDate      string
	PageCount    int
	Version      string
	Encrypted    bool
	FileSize     int64
}

// ============================================================
// Internal PDF Parsing Types
// ============================================================

// PDFObj represents a parsed PDF object.
type PDFObj struct {
	Number   int64         // Object number
	Generation int64       // Generation number
	Value    interface{}   // The actual value (PDFDict, PDFArray, PDFStream, etc.)
	Offset   int64         // Byte offset in file (for reference)
}

// PDFDict represents a PDF dictionary object.
type PDFDict struct {
	Entries map[string]interface{}
}

// PDFArray represents a PDF array object.
type PDFArray struct {
	Elements []interface{}
}

// PDFStream represents a PDF stream object.
type PDFStream struct {
	Dict     *PDFDict
	Data     []byte
	RawData  []byte // Uncompressed data
}

// PDFRef represents an indirect object reference.
type PDFRef struct {
	ObjectNum  int64
	Generation int64
}

// PDFName represents a PDF name object.
type PDFName string

// PDFString represents a PDF string object.
type PDFString string

// PDFHexStr represents a PDF hexadecimal string.
type PDFHexStr string

// PDFXRefTable represents the cross-reference table.
type PDFXRefTable struct {
	Entries map[int64]*PDFXRefEntry
}

// PDFXRefEntry represents an entry in the cross-reference table.
type PDFXRefEntry struct {
	Type       int   // 0 = free, 1 = in-use, 2 = compressed
	Offset     int64 // Byte offset (type 1) or object stream number (type 2)
	Generation int64 // Generation number
	Index      int   // Index within object stream (type 2 only)
}

// ============================================================
// Object Interface Implementations
// ============================================================

// --- PDF Object ---

// Type returns the object type.
func (p *PDF) Type() ObjectType { return PDFType }

// TypeTag returns the type tag for fast type checking.
func (p *PDF) TypeTag() TypeTag { return TagPDF }

// Inspect returns a string representation of the PDF object.
func (p *PDF) Inspect() string {
	if p.IsOpen {
		return fmt.Sprintf("<PDF open path=%s pages=%d>", p.FilePath, p.PageCount)
	}
	return fmt.Sprintf("<PDF closed path=%s>", p.FilePath)
}

// ToBool returns true if the PDF is open.
func (p *PDF) ToBool() *Bool { return &Bool{Value: p.IsOpen} }

// HashKey returns a hash key for the PDF object.
func (p *PDF) HashKey() HashKey {
	return HashKey{Type: PDFType, Value: uint64(len(p.FilePath)) << 32}
}

// GetMember returns a member by name for script access.
func (p *PDF) GetMember(name string) Object {
	switch name {
	case "pageCount":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewInt(int64(p.PageCount))
		}}
	case "version":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewString(p.Version)
		}}
	case "getInfo":
		return &Builtin{Fn: func(args ...Object) Object {
			return p.GetInfo()
		}}
	case "getPage":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("getPage() takes exactly 1 argument")
			}
			idx, ok := args[0].(*Int)
			if !ok {
				return newError("getPage() requires an integer argument")
			}
			return p.GetPage(int(idx.Value))
		}}
	case "extractText":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("extractText() takes exactly 1 argument")
			}
			idx, ok := args[0].(*Int)
			if !ok {
				return newError("extractText() requires an integer argument")
			}
			return p.ExtractText(int(idx.Value))
		}}
	case "extractAllText":
		return &Builtin{Fn: func(args ...Object) Object {
			return p.ExtractAllText()
		}}
	case "rotatePage":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("rotatePage() takes exactly 2 arguments")
			}
			idx, ok := args[0].(*Int)
			if !ok {
				return newError("rotatePage() requires an integer as first argument")
			}
			angle, ok := args[1].(*Int)
			if !ok {
				return newError("rotatePage() requires an integer as second argument")
			}
			return p.RotatePage(int(idx.Value), int(angle.Value))
		}}
	case "close":
		return &Builtin{Fn: func(args ...Object) Object {
			return p.Close()
		}}
	case "saveAs":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("saveAs() takes exactly 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("saveAs() requires a string argument")
			}
			return p.SaveAs(path.Value)
		}}
	case "toBytes":
		return &Builtin{Fn: func(args ...Object) Object {
			return p.ToBytes()
		}}
	}
	return NULL
}

// --- PDFDocument Object ---

// Type returns the object type.
func (d *PDFDocument) Type() ObjectType { return PDFDocumentType }

// TypeTag returns the type tag for fast type checking.
func (d *PDFDocument) TypeTag() TypeTag { return TagPDFDocument }

// Inspect returns a string representation.
func (d *PDFDocument) Inspect() string {
	return fmt.Sprintf("<PDF_DOCUMENT pages=%d>", len(d.Pages))
}

// ToBool always returns true.
func (d *PDFDocument) ToBool() *Bool { return TRUE }

// HashKey returns a hash key.
func (d *PDFDocument) HashKey() HashKey {
	return HashKey{Type: PDFDocumentType, Value: uint64(len(d.Pages))}
}

// GetMember returns a member by name for script access.
func (d *PDFDocument) GetMember(name string) Object {
	switch name {
	case "addPage":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("addPage() takes at least 2 arguments")
			}
			width, ok := args[0].(*Float)
			if !ok {
				if w, ok := args[0].(*Int); ok {
					width = NewFloat(float64(w.Value))
				} else {
					return newError("addPage() requires numeric width")
				}
			}
			height, ok := args[1].(*Float)
			if !ok {
				if h, ok := args[1].(*Int); ok {
					height = NewFloat(float64(h.Value))
				} else {
					return newError("addPage() requires numeric height")
				}
			}
			return d.AddPage(width.Value, height.Value)
		}}
	case "writeText":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) < 4 {
				return newError("writeText() takes at least 4 arguments")
			}
			pageIdx, ok := args[0].(*Int)
			if !ok {
				return newError("writeText() requires integer page index")
			}
			text, ok := args[1].(*String)
			if !ok {
				return newError("writeText() requires string text")
			}
			x, ok := args[2].(*Float)
			if !ok {
				if xi, ok := args[2].(*Int); ok {
					x = NewFloat(float64(xi.Value))
				} else {
					return newError("writeText() requires numeric x")
				}
			}
			y, ok := args[3].(*Float)
			if !ok {
				if yi, ok := args[3].(*Int); ok {
					y = NewFloat(float64(yi.Value))
				} else {
					return newError("writeText() requires numeric y")
				}
			}
			opts := make(map[string]interface{})
			if len(args) > 4 {
				if m, ok := args[4].(*Map); ok {
					for _, pair := range m.Pairs {
						if key, ok := pair.Key.(*String); ok {
							opts[key.Value] = pair.Value
						}
					}
				}
			}
			return d.WriteText(int(pageIdx.Value), text.Value, x.Value, y.Value, opts)
		}}
	case "setFont":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("setFont() takes at least 2 arguments")
			}
			name, ok := args[0].(*String)
			if !ok {
				return newError("setFont() requires string font name")
			}
			size, ok := args[1].(*Float)
			if !ok {
				if s, ok := args[1].(*Int); ok {
					size = NewFloat(float64(s.Value))
				} else {
					return newError("setFont() requires numeric size")
				}
			}
			return d.SetFont(name.Value, size.Value)
		}}
	case "setTitle":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("setTitle() takes exactly 1 argument")
			}
			title, ok := args[0].(*String)
			if !ok {
				return newError("setTitle() requires a string argument")
			}
			d.Title = title.Value
			return NULL
		}}
	case "setAuthor":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("setAuthor() takes exactly 1 argument")
			}
			author, ok := args[0].(*String)
			if !ok {
				return newError("setAuthor() requires a string argument")
			}
			d.Author = author.Value
			return NULL
		}}
	case "save":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("save() takes exactly 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("save() requires a string argument")
			}
			return d.Save(path.Value)
		}}
	case "toBytes":
		return &Builtin{Fn: func(args ...Object) Object {
			return d.ToBytes()
		}}
	case "pageCount":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewInt(int64(len(d.Pages)))
		}}
	}
	return NULL
}

// --- PDFPage Object ---

// Type returns the object type.
func (pg *PDFPage) Type() ObjectType { return PDFPageType }

// TypeTag returns the type tag for fast type checking.
func (pg *PDFPage) TypeTag() TypeTag { return TagPDFPage }

// Inspect returns a string representation.
func (pg *PDFPage) Inspect() string {
	return fmt.Sprintf("<PDF_PAGE num=%d width=%.1f height=%.1f>",
		pg.PageNum, pg.Width, pg.Height)
}

// ToBool always returns true.
func (pg *PDFPage) ToBool() *Bool { return TRUE }

// HashKey returns a hash key.
func (pg *PDFPage) HashKey() HashKey {
	return HashKey{Type: PDFPageType, Value: uint64(pg.PageNum)}
}

// GetMember returns a member by name for script access.
func (pg *PDFPage) GetMember(name string) Object {
	switch name {
	case "width":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewFloat(pg.Width)
		}}
	case "height":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewFloat(pg.Height)
		}}
	case "pageNum":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewInt(int64(pg.PageNum))
		}}
	case "rotation":
		return &Builtin{Fn: func(args ...Object) Object {
			return NewInt(int64(pg.Rotation))
		}}
	case "extractText":
		return &Builtin{Fn: func(args ...Object) Object {
			return pg.ExtractText()
		}}
	}
	return NULL
}

// --- PDFInfo Object ---

// Type returns the object type.
func (i *PDFInfo) Type() ObjectType { return PDFInfoType }

// TypeTag returns the type tag for fast type checking.
func (i *PDFInfo) TypeTag() TypeTag { return TagPDFInfo }

// Inspect returns a string representation.
func (i *PDFInfo) Inspect() string {
	return fmt.Sprintf("<PDF_INFO title=%q author=%q pages=%d>",
		i.Title, i.Author, i.PageCount)
}

// ToBool always returns true.
func (i *PDFInfo) ToBool() *Bool { return TRUE }

// HashKey returns a hash key.
func (i *PDFInfo) HashKey() HashKey {
	return HashKey{Type: PDFInfoType, Value: uint64(i.PageCount)}
}

// GetMember returns a member by name for script access.
func (i *PDFInfo) GetMember(name string) Object {
	switch name {
	case "title":
		return NewString(i.Title)
	case "author":
		return NewString(i.Author)
	case "subject":
		return NewString(i.Subject)
	case "keywords":
		return NewString(i.Keywords)
	case "creator":
		return NewString(i.Creator)
	case "producer":
		return NewString(i.Producer)
	case "creationDate":
		return NewString(i.CreationDate)
	case "modDate":
		return NewString(i.ModDate)
	case "pageCount":
		return NewInt(int64(i.PageCount))
	case "version":
		return NewString(i.Version)
	case "encrypted":
		if i.Encrypted {
			return TRUE
		}
		return FALSE
	case "fileSize":
		return NewInt(i.FileSize)
	}
	return NULL
}

// ============================================================
// PDF Constructor and Core Methods
// ============================================================

// NewPDF creates a new PDF object from file data.
func NewPDF(data []byte, filePath string) *PDF {
	return &PDF{
		FilePath: filePath,
		Source:   data,
		Objects:  make(map[int64]*PDFObj),
		IsOpen:   true,
		Version:  "1.4",
	}
}

// NewPDFFromFile creates a PDF object from a file path.
func NewPDFFromFile(filePath string) (*PDF, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	pdf := NewPDF(data, filePath)
	if err := pdf.Parse(); err != nil {
		return nil, err
	}
	return pdf, nil
}

// NewPDFFromBytes creates a PDF object from byte data.
func NewPDFFromBytes(data []byte) (*PDF, error) {
	pdf := NewPDF(data, "")
	if err := pdf.Parse(); err != nil {
		return nil, err
	}
	return pdf, nil
}

// Close closes the PDF and releases resources.
func (p *PDF) Close() Object {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.IsOpen = false
	p.Source = nil
	p.Objects = nil
	p.XRefTable = nil
	return NULL
}

// Parse parses the PDF structure.
func (p *PDF) Parse() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.Source) == 0 {
		return fmt.Errorf("empty PDF data")
	}

	// Parse version
	p.parseVersion()

	// Parse xref table and trailer
	if err := p.parseXRefAndTrailer(); err != nil {
		return err
	}

	// Parse objects
	if err := p.parseObjects(); err != nil {
		return err
	}

	// Parse page tree
	if err := p.parsePageTree(); err != nil {
		return err
	}

	return nil
}

// parseVersion extracts the PDF version from the header.
func (p *PDF) parseVersion() {
	header := string(p.Source[:min(50, len(p.Source))])
	if strings.HasPrefix(header, "%PDF-") {
		// Extract version (e.g., "1.4", "1.7")
		re := regexp.MustCompile(`%PDF-(\d+\.\d+)`)
		matches := re.FindStringSubmatch(header)
		if len(matches) > 1 {
			p.Version = matches[1]
		}
	}
}

// parseXRefAndTrailer parses the cross-reference table and trailer.
func (p *PDF) parseXRefAndTrailer() error {
	// Find startxref
	startXRefOffset := p.findStartXRef()
	if startXRefOffset < 0 {
		return fmt.Errorf("could not find startxref")
	}

	// Parse xref table
	xref, err := p.parseXRefAt(startXRefOffset)
	if err != nil {
		// Try to find xref streams (PDF 1.5+)
		xref, err = p.parseXRefStream(startXRefOffset)
		if err != nil {
			return fmt.Errorf("failed to parse xref: %v", err)
		}
	}
	p.XRefTable = xref

	// Parse trailer
	trailer, err := p.parseTrailer(startXRefOffset)
	if err != nil {
		return err
	}
	p.Trailer = trailer

	// Get root object number
	if root, ok := trailer.Entries["Root"]; ok {
		if ref, ok := root.(PDFRef); ok {
			p.RootObj = ref.ObjectNum
		}
	}

	// Get info object number
	if info, ok := trailer.Entries["Info"]; ok {
		if ref, ok := info.(PDFRef); ok {
			p.InfoObj = ref.ObjectNum
		}
	}

	return nil
}

// findStartXRef locates the startxref offset.
func (p *PDF) findStartXRef() int64 {
	// Search backwards for "startxref"
	data := p.Source
	idx := bytes.LastIndex(data, []byte("startxref"))
	if idx < 0 {
		return -1
	}

	// Parse the offset after "startxref"
	line := string(data[idx:])
	lines := strings.Fields(line)
	if len(lines) < 2 {
		return -1
	}

	offset, err := strconv.ParseInt(lines[1], 10, 64)
	if err != nil {
		return -1
	}

	return offset
}

// parseXRefAt parses the cross-reference table at the given offset.
func (p *PDF) parseXRefAt(offset int64) (*PDFXRefTable, error) {
	xref := &PDFXRefTable{
		Entries: make(map[int64]*PDFXRefEntry),
	}

	// Find xref section - search for "xref" near the offset
	start := int(offset)
	if start >= len(p.Source) {
		return nil, fmt.Errorf("xref offset out of bounds")
	}

	// Search for "xref" within a reasonable range from the offset
	searchStart := start - 20
	if searchStart < 0 {
		searchStart = 0
	}
	searchEnd := start + 100
	if searchEnd > len(p.Source) {
		searchEnd = len(p.Source)
	}

	searchArea := p.Source[searchStart:searchEnd]
	xrefIdx := bytes.Index(searchArea, []byte("xref"))
	if xrefIdx < 0 {
		return nil, fmt.Errorf("xref keyword not found near offset")
	}

	section := p.Source[searchStart+xrefIdx:]

	// Find "xref" line
	lines := bytes.Split(section, []byte("\n"))
	lineIdx := 0

	// Skip to xref line
	for lineIdx < len(lines) && !bytes.HasPrefix(bytes.TrimSpace(lines[lineIdx]), []byte("xref")) {
		lineIdx++
	}
	if lineIdx >= len(lines) {
		return nil, fmt.Errorf("xref line not found")
	}
	lineIdx++ // Skip "xref" line

	// Parse subsections
	for lineIdx < len(lines) {
		line := bytes.TrimSpace(lines[lineIdx])
		if len(line) == 0 {
			lineIdx++
			continue
		}

		// Check for trailer
		if bytes.HasPrefix(line, []byte("trailer")) {
			break
		}

		// Parse subsection header "start count"
		parts := bytes.Fields(line)
		if len(parts) == 2 {
			startObj, err1 := strconv.ParseInt(string(parts[0]), 10, 64)
			count, err2 := strconv.ParseInt(string(parts[1]), 10, 64)
			if err1 == nil && err2 == nil && count > 0 {
				// Parse entries
				for j := int64(0); j < count && lineIdx+1 < len(lines); j++ {
					lineIdx++
					entryLine := bytes.TrimSpace(lines[lineIdx])
					entryParts := bytes.Fields(entryLine)
					if len(entryParts) >= 3 {
						entry := &PDFXRefEntry{}
						entry.Offset, _ = strconv.ParseInt(string(entryParts[0]), 10, 64)
						entry.Generation, _ = strconv.ParseInt(string(entryParts[1]), 10, 64)
						entry.Type = 0
						if bytes.Equal(entryParts[2], []byte("n")) {
							entry.Type = 1
						}
						xref.Entries[startObj+j] = entry
					}
				}
			}
		}
		lineIdx++
	}

	if len(xref.Entries) == 0 {
		return nil, fmt.Errorf("no xref entries found")
	}

	return xref, nil
}

// parseXRefStream parses an xref stream (PDF 1.5+).
func (p *PDF) parseXRefStream(offset int64) (*PDFXRefTable, error) {
	xref := &PDFXRefTable{
		Entries: make(map[int64]*PDFXRefEntry),
	}

	// Find object at offset
	start := int(offset)
	if start >= len(p.Source) {
		return nil, fmt.Errorf("xref stream offset out of bounds")
	}

	// Parse object header "N G obj"
	objHeader := string(p.Source[start:min(start+100, len(p.Source))])
	re := regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
	matches := re.FindStringSubmatch(objHeader)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid xref stream object")
	}

	objNum, _ := strconv.ParseInt(matches[1], 10, 64)

	// Get object
	obj, err := p.parseObjectAt(objNum, offset)
	if err != nil {
		return nil, err
	}

	stream, ok := obj.Value.(*PDFStream)
	if !ok {
		return nil, fmt.Errorf("xref object is not a stream")
	}

	// Parse stream data
	if err := p.parseXRefStreamData(stream, xref); err != nil {
		return nil, err
	}

	return xref, nil
}

// parseXRefStreamData parses xref stream data.
func (p *PDF) parseXRefStreamData(stream *PDFStream, xref *PDFXRefTable) error {
	// Decompress stream data
	data, err := p.decodeStream(stream)
	if err != nil {
		return err
	}

	// Get /W array (field widths)
	wArr, ok := stream.Dict.Entries["W"].(*PDFArray)
	if !ok {
		return fmt.Errorf("missing /W in xref stream")
	}

	if len(wArr.Elements) < 3 {
		return fmt.Errorf("invalid /W array")
	}

	w0 := int(wArr.Elements[0].(int64))
	w1 := int(wArr.Elements[1].(int64))
	w2 := int(wArr.Elements[2].(int64))
	entrySize := w0 + w1 + w2

	// Get /Index array (object ranges)
	indexArr, ok := stream.Dict.Entries["Index"].(*PDFArray)
	if !ok {
		// Default to [0 Size]
		size, _ := stream.Dict.Entries["Size"].(int64)
		indexArr = &PDFArray{Elements: []interface{}{int64(0), size}}
	}

	// Parse entries
	dataIdx := 0
	for i := 0; i < len(indexArr.Elements); i += 2 {
		startObj := indexArr.Elements[i].(int64)
		count := indexArr.Elements[i+1].(int64)

		for j := int64(0); j < count; j++ {
			if dataIdx+entrySize > len(data) {
				break
			}

			entry := &PDFXRefEntry{}

			// Read fields
			if w0 > 0 {
				entry.Type = int(readInt(data[dataIdx : dataIdx+w0]))
				dataIdx += w0
			} else {
				entry.Type = 1 // Default to in-use
			}

			if w1 > 0 {
				entry.Offset = readInt(data[dataIdx : dataIdx+w1])
				dataIdx += w1
			}

			if w2 > 0 {
				entry.Generation = readInt(data[dataIdx : dataIdx+w2])
				dataIdx += w2
			}

			xref.Entries[startObj+j] = entry
		}
	}

	return nil
}

// parseTrailer parses the PDF trailer dictionary.
func (p *PDF) parseTrailer(startXRefOffset int64) (*PDFDict, error) {
	// Find "trailer" in the file
	data := p.Source
	trailerIdx := bytes.Index(data, []byte("trailer"))
	if trailerIdx < 0 {
		// xref stream - trailer is in the stream dictionary
		return &PDFDict{Entries: make(map[string]interface{})}, nil
	}

	// Find << after trailer
	trailerData := data[trailerIdx:]
	dictStart := bytes.Index(trailerData, []byte("<<"))
	if dictStart < 0 {
		return &PDFDict{Entries: make(map[string]interface{})}, nil
	}

	// Parse dictionary
	dict, _, err := p.parseDictionary(trailerData[dictStart:])
	if err != nil {
		return nil, err
	}

	return dict, nil
}

// parseObjects parses all objects in the PDF.
func (p *PDF) parseObjects() error {
	for objNum, entry := range p.XRefTable.Entries {
		if entry.Type == 1 && entry.Offset > 0 {
			obj, err := p.parseObjectAt(objNum, entry.Offset)
			if err != nil {
				// Skip invalid objects
				continue
			}
			p.Objects[objNum] = obj
		}
	}
	return nil
}

// parseObjectAt parses an object at the given offset.
func (p *PDF) parseObjectAt(objNum, offset int64) (*PDFObj, error) {
	start := int(offset)
	if start >= len(p.Source) {
		return nil, fmt.Errorf("object offset out of bounds")
	}

	data := p.Source[start:]

	// Skip whitespace and find "obj"
	objIdx := bytes.Index(data, []byte("obj"))
	if objIdx < 0 {
		return nil, fmt.Errorf("object not found")
	}

	// Skip to after "obj"
	pos := objIdx + 3

	// Parse the object value
	value, endPos, err := p.parseValue(data[pos:])
	if err != nil {
		return nil, err
	}

	obj := &PDFObj{
		Number:   objNum,
		Value:    value,
		Offset:   offset,
	}

	// Check for stream - only if it's a dictionary and "stream" follows immediately
	if dict, ok := value.(*PDFDict); ok {
		remaining := data[pos+endPos:]
		// Skip whitespace to check if "stream" keyword follows
		tempPos := 0
		for tempPos < len(remaining) && pdfIsWhitespace(remaining[tempPos]) {
			tempPos++
		}

		// Check if the next non-whitespace content is "stream"
		if bytes.HasPrefix(remaining[tempPos:], []byte("stream")) {
			// Verify it's the "stream" keyword (followed by whitespace or newline)
			streamEnd := tempPos + 6
			if streamEnd < len(remaining) && (pdfIsWhitespace(remaining[streamEnd]) || remaining[streamEnd] == '\r' || remaining[streamEnd] == '\n') {
				// Find stream data start (after \n or \r\n)
				streamStart := streamEnd
				if streamStart < len(remaining) && remaining[streamStart] == '\r' {
					streamStart++
				}
				if streamStart < len(remaining) && remaining[streamStart] == '\n' {
					streamStart++
				}

				// Find stream end
				endStreamIdx := bytes.Index(remaining[streamStart:], []byte("endstream"))
				if endStreamIdx >= 0 {
					streamData := remaining[streamStart : streamStart+endStreamIdx]

					// Create PDFStream object
					pdfStream := &PDFStream{
						Dict:    dict,
						RawData: streamData,
					}
					obj.Value = pdfStream
				}
			}
		}
	}

	return obj, nil
}

// parseValue parses a PDF value (dictionary, array, string, number, etc.).
func (p *PDF) parseValue(data []byte) (interface{}, int, error) {
	// Skip whitespace
	pos := 0
	for pos < len(data) && pdfIsWhitespace(data[pos]) {
		pos++
	}

	if pos >= len(data) {
		return nil, 0, fmt.Errorf("unexpected end of data")
	}

	ch := data[pos]

	switch ch {
	case '<':
		if pos+1 < len(data) && data[pos+1] == '<' {
			// Dictionary
			dict, end, err := p.parseDictionary(data[pos:])
			return dict, pos + end, err
		}
		// Hex string
		str, end, err := p.parseHexString(data[pos:])
		return str, pos + end, err

	case '[':
		// Array
		arr, end, err := p.parseArray(data[pos:])
		return arr, pos + end, err

	case '(':
		// String
		str, end, err := p.parseString(data[pos:])
		return str, pos + end, err

	case '/':
		// Name
		name, end, err := p.parseName(data[pos:])
		return name, pos + end, err

	default:
		// Number or reference
		if pdfIsDigit(ch) || ch == '-' || ch == '+' || ch == '.' {
			return p.parseNumberOrRef(data[pos:])
		}

		// Keyword (true, false, null)
		kw, end, err := p.parseKeyword(data[pos:])
		return kw, pos + end, err
	}
}

// parseDictionary parses a PDF dictionary << ... >>.
func (p *PDF) parseDictionary(data []byte) (*PDFDict, int, error) {
	dict := &PDFDict{Entries: make(map[string]interface{})}

	if !bytes.HasPrefix(data, []byte("<<")) {
		return nil, 0, fmt.Errorf("expected <<")
	}

	pos := 2
	for pos < len(data) {
		// Skip whitespace
		for pos < len(data) && pdfIsWhitespace(data[pos]) {
			pos++
		}

		// Check for end
		if pos+1 < len(data) && data[pos] == '>' && data[pos+1] == '>' {
			return dict, pos + 2, nil
		}

		// Parse name (key)
		if data[pos] != '/' {
			return nil, 0, fmt.Errorf("expected name in dictionary")
		}
		name, end, err := p.parseName(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += end

		// Skip whitespace
		for pos < len(data) && pdfIsWhitespace(data[pos]) {
			pos++
		}

		// Parse value
		value, end, err := p.parseValue(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += end

		dict.Entries[string(name)] = value
	}

	return dict, pos, nil
}

// parseArray parses a PDF array [ ... ].
func (p *PDF) parseArray(data []byte) (*PDFArray, int, error) {
	arr := &PDFArray{}

	if !bytes.HasPrefix(data, []byte("[")) {
		return nil, 0, fmt.Errorf("expected [")
	}

	pos := 1
	for pos < len(data) {
		// Skip whitespace
		for pos < len(data) && pdfIsWhitespace(data[pos]) {
			pos++
		}

		// Check for end
		if data[pos] == ']' {
			return arr, pos + 1, nil
		}

		// Parse value
		value, end, err := p.parseValue(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += end

		arr.Elements = append(arr.Elements, value)
	}

	return arr, pos, nil
}

// parseString parses a PDF literal string ( ... ).
func (p *PDF) parseString(data []byte) (PDFString, int, error) {
	if len(data) == 0 || data[0] != '(' {
		return "", 0, fmt.Errorf("expected (")
	}

	var result []byte
	pos := 1
	parens := 1

	for pos < len(data) && parens > 0 {
		ch := data[pos]

		if ch == '\\' && pos+1 < len(data) {
			// Escape sequence
			pos++
			switch data[pos] {
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case '(':
				result = append(result, '(')
			case ')':
				result = append(result, ')')
			case '\\':
				result = append(result, '\\')
			case '\n':
				// Line continuation
			case '\r':
				// Line continuation
				if pos+1 < len(data) && data[pos+1] == '\n' {
					pos++
				}
			default:
				if isOctalDigit(data[pos]) {
					// Octal escape
					octal := data[pos] - '0'
					pos++
					if pos < len(data) && isOctalDigit(data[pos]) {
						octal = octal*8 + data[pos] - '0'
						pos++
						if pos < len(data) && isOctalDigit(data[pos]) {
							octal = octal*8 + data[pos] - '0'
						} else {
							pos--
						}
					} else {
						pos--
					}
					result = append(result, octal)
				} else {
					result = append(result, data[pos])
				}
			}
		} else if ch == '(' {
			parens++
			result = append(result, ch)
		} else if ch == ')' {
			parens--
			if parens > 0 {
				result = append(result, ch)
			}
		} else {
			result = append(result, ch)
		}
		pos++
	}

	return PDFString(result), pos, nil
}

// parseHexString parses a PDF hex string < ... >.
func (p *PDF) parseHexString(data []byte) (PDFHexStr, int, error) {
	if len(data) == 0 || data[0] != '<' {
		return "", 0, fmt.Errorf("expected <")
	}

	pos := 1
	end := bytes.IndexByte(data[pos:], '>')
	if end < 0 {
		return "", 0, fmt.Errorf("unterminated hex string")
	}

	return PDFHexStr(data[pos : pos+end]), pos + end + 1, nil
}

// parseName parses a PDF name /Name.
func (p *PDF) parseName(data []byte) (PDFName, int, error) {
	if len(data) == 0 || data[0] != '/' {
		return "", 0, fmt.Errorf("expected /")
	}

	pos := 1
	var result []byte

	for pos < len(data) {
		ch := data[pos]
		if pdfIsWhitespace(ch) || ch == '/' || ch == '[' || ch == ']' ||
			ch == '<' || ch == '>' || ch == '(' || ch == ')' {
			break
		}

		if ch == '#' && pos+2 < len(data) {
			// Hex escape
			hexStr := string(data[pos+1 : pos+3])
			if b, err := hex.DecodeString(hexStr); err == nil {
				result = append(result, b...)
			}
			pos += 3
		} else {
			result = append(result, ch)
			pos++
		}
	}

	return PDFName(result), pos, nil
}

// parseNumberOrRef parses a number or object reference.
func (p *PDF) parseNumberOrRef(data []byte) (interface{}, int, error) {
	// Read the number
	pos := 0
	neg := false
	if data[pos] == '-' {
		neg = true
		pos++
	} else if data[pos] == '+' {
		pos++
	}

	// Read integer part
	intPart := int64(0)
	for pos < len(data) && pdfIsDigit(data[pos]) {
		intPart = intPart*10 + int64(data[pos]-'0')
		pos++
	}

	// Check for decimal
	isFloat := false
	floatVal := float64(intPart)
	if pos < len(data) && data[pos] == '.' {
		isFloat = true
		pos++
		frac := float64(0)
		div := float64(10)
		for pos < len(data) && pdfIsDigit(data[pos]) {
			frac += float64(data[pos]-'0') / div
			div *= 10
			pos++
		}
		floatVal += frac
	}

	if neg {
		floatVal = -floatVal
		intPart = -intPart
	}

	// Skip whitespace
	startPos := pos
	for pos < len(data) && pdfIsWhitespace(data[pos]) {
		pos++
	}

	// Check for reference "R"
	if pos+1 < len(data) && data[pos] == 'R' && !isAlphaNum(data[pos+1]) {
		// This is a reference, but we only have one number
		// The first number should be the object number
		// We need to go back and parse properly
		return PDFRef{ObjectNum: intPart}, pos + 1, nil
	}

	// Check for "0 0 R" pattern
	if pos < len(data) && pdfIsDigit(data[pos]) {
		// Could be a reference "N G R"
		gen := int64(0)
		for pos < len(data) && pdfIsDigit(data[pos]) {
			gen = gen*10 + int64(data[pos]-'0')
			pos++
		}

		// Skip whitespace
		for pos < len(data) && pdfIsWhitespace(data[pos]) {
			pos++
		}

		if pos < len(data) && data[pos] == 'R' {
			return PDFRef{ObjectNum: intPart, Generation: gen}, pos + 1, nil
		}

		// Not a reference, rewind
		pos = startPos
	}

	if isFloat {
		return floatVal, pos, nil
	}
	return intPart, pos, nil
}

// parseKeyword parses a keyword (true, false, null).
func (p *PDF) parseKeyword(data []byte) (interface{}, int, error) {
	end := 0
	for end < len(data) && isAlphaNum(data[end]) {
		end++
	}

	kw := string(data[:end])
	switch kw {
	case "true":
		return true, end, nil
	case "false":
		return false, end, nil
	case "null":
		return nil, end, nil
	default:
		return kw, end, nil
	}
}

// parsePageTree parses the page tree to get all pages.
func (p *PDF) parsePageTree() error {
	if p.RootObj == 0 {
		return fmt.Errorf("no root object")
	}

	rootObj, ok := p.Objects[p.RootObj]
	if !ok {
		return fmt.Errorf("root object not found")
	}

	rootDict, ok := rootObj.Value.(*PDFDict)
	if !ok {
		return fmt.Errorf("root is not a dictionary")
	}

	pagesRef, ok := rootDict.Entries["Pages"].(PDFRef)
	if !ok {
		return fmt.Errorf("pages reference not found")
	}

	pages, err := p.collectPages(pagesRef)
	if err != nil {
		return err
	}

	p.Pages = pages
	p.PageCount = len(pages)

	return nil
}

// collectPages recursively collects all pages from the page tree.
func (p *PDF) collectPages(pagesRef PDFRef) ([]*PDFPage, error) {
	pagesObj, ok := p.Objects[pagesRef.ObjectNum]
	if !ok {
		return nil, fmt.Errorf("pages object not found: %d", pagesRef.ObjectNum)
	}

	pagesDict, ok := pagesObj.Value.(*PDFDict)
	if !ok {
		return nil, fmt.Errorf("pages is not a dictionary")
	}

	pageType, _ := pagesDict.Entries["Type"].(PDFName)
	if pageType == "Page" {
		// Single page
		pg, err := p.createPage(pagesRef.ObjectNum, pagesDict)
		if err != nil {
			return nil, err
		}
		return []*PDFPage{pg}, nil
	}

	// Page node
	var result []*PDFPage

	kids, ok := pagesDict.Entries["Kids"].(*PDFArray)
	if !ok {
		return nil, fmt.Errorf("no Kids array")
	}

	for _, kid := range kids.Elements {
		kidRef, ok := kid.(PDFRef)
		if !ok {
			continue
		}

		pages, err := p.collectPages(kidRef)
		if err != nil {
			continue
		}
		result = append(result, pages...)
	}

	return result, nil
}

// createPage creates a PDFPage from a page dictionary.
func (p *PDF) createPage(objNum int64, pageDict *PDFDict) (*PDFPage, error) {
	pg := &PDFPage{
		PDF:     p,
		PageNum: len(p.Pages),
	}

	// Get MediaBox
	if mediaBox, ok := pageDict.Entries["MediaBox"].(*PDFArray); ok {
		pg.MediaBox = mediaBox
		if len(mediaBox.Elements) >= 4 {
			// MediaBox: [x0 y0 x1 y1]
			x0 := getFloat(mediaBox.Elements[0])
			y0 := getFloat(mediaBox.Elements[1])
			x1 := getFloat(mediaBox.Elements[2])
			y1 := getFloat(mediaBox.Elements[3])
			pg.Width = x1 - x0
			pg.Height = y1 - y0
		}
	}

	// Get CropBox (optional)
	if cropBox, ok := pageDict.Entries["CropBox"].(*PDFArray); ok {
		pg.CropBox = cropBox
	}

	// Get rotation
	if rot, ok := pageDict.Entries["Rotate"].(int64); ok {
		pg.Rotation = int(rot)
	}

	// Get content streams
	if contents, ok := pageDict.Entries["Contents"]; ok {
		switch c := contents.(type) {
		case PDFRef:
			pg.Contents = []int64{c.ObjectNum}
		case *PDFArray:
			for _, elem := range c.Elements {
				if ref, ok := elem.(PDFRef); ok {
					pg.Contents = append(pg.Contents, ref.ObjectNum)
				}
			}
		}
	}

	// Get resources
	if resources, ok := pageDict.Entries["Resources"].(*PDFDict); ok {
		pg.Resources = resources
	} else if resRef, ok := pageDict.Entries["Resources"].(PDFRef); ok {
		if resObj, ok := p.Objects[resRef.ObjectNum]; ok {
			if resDict, ok := resObj.Value.(*PDFDict); ok {
				pg.Resources = resDict
			}
		}
	}

	return pg, nil
}

// ============================================================
// PDF Methods - Information and Text Extraction
// ============================================================

// GetInfo returns the document information.
func (p *PDF) GetInfo() *PDFInfo {
	info := &PDFInfo{
		PageCount: p.PageCount,
		Version:   p.Version,
		FileSize:  int64(len(p.Source)),
	}

	if p.InfoObj > 0 {
		if infoObj, ok := p.Objects[p.InfoObj]; ok {
			if infoDict, ok := infoObj.Value.(*PDFDict); ok {
				if title, ok := infoDict.Entries["Title"]; ok {
					info.Title = pdfStringToString(title)
				}
				if author, ok := infoDict.Entries["Author"]; ok {
					info.Author = pdfStringToString(author)
				}
				if subject, ok := infoDict.Entries["Subject"]; ok {
					info.Subject = pdfStringToString(subject)
				}
				if keywords, ok := infoDict.Entries["Keywords"]; ok {
					info.Keywords = pdfStringToString(keywords)
				}
				if creator, ok := infoDict.Entries["Creator"]; ok {
					info.Creator = pdfStringToString(creator)
				}
				if producer, ok := infoDict.Entries["Producer"]; ok {
					info.Producer = pdfStringToString(producer)
				}
				if creationDate, ok := infoDict.Entries["CreationDate"]; ok {
					info.CreationDate = pdfDateToString(creationDate)
				}
				if modDate, ok := infoDict.Entries["ModDate"]; ok {
					info.ModDate = pdfDateToString(modDate)
				}
			}
		}
	}

	return info
}

// GetPage returns a page by index.
func (p *PDF) GetPage(index int) Object {
	if index < 0 || index >= len(p.Pages) {
		return newError("page index out of range")
	}
	return p.Pages[index]
}

// ExtractText extracts text from a specific page.
func (p *PDF) ExtractText(pageIndex int) Object {
	if pageIndex < 0 || pageIndex >= len(p.Pages) {
		return newError("page index out of range")
	}
	return p.Pages[pageIndex].ExtractText()
}

// ExtractAllText extracts text from all pages.
func (p *PDF) ExtractAllText() Object {
	var result strings.Builder
	for i, pg := range p.Pages {
		if i > 0 {
			result.WriteString("\n\n")
		}
		text := pg.ExtractText()
		if s, ok := text.(*String); ok {
			result.WriteString(s.Value)
		}
	}
	return NewString(result.String())
}

// RotatePage rotates a page by the specified angle.
func (p *PDF) RotatePage(pageIndex int, angle int) Object {
	if pageIndex < 0 || pageIndex >= len(p.Pages) {
		return newError("page index out of range")
	}

	pg := p.Pages[pageIndex]
	newRotation := (pg.Rotation + angle) % 360
	if newRotation < 0 {
		newRotation += 360
	}
	pg.Rotation = newRotation
	p.IsModified = true

	return NULL
}

// SaveAs saves the PDF to a file.
func (p *PDF) SaveAs(filePath string) Object {
	data, err := p.renderPDF()
	if err != nil {
		return newError("failed to save PDF: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return newError("failed to write file: %v", err)
	}

	return NULL
}

// ToBytes returns the PDF as a byte array.
func (p *PDF) ToBytes() Object {
	data, err := p.renderPDF()
	if err != nil {
		return newError("failed to render PDF: %v", err)
	}

	elements := make([]Object, len(data))
	for i, b := range data {
		elements[i] = NewInt(int64(b))
	}
	return NewArray(elements)
}

// renderPDF renders the PDF to bytes.
func (p *PDF) renderPDF() ([]byte, error) {
	// If not modified, return original
	if !p.IsModified {
		return p.Source, nil
	}

	// TODO: Implement full PDF re-rendering
	// For now, just return the original
	return p.Source, nil
}

// ============================================================
// PDFPage Methods
// ============================================================

// ExtractText extracts text from the page.
func (pg *PDFPage) ExtractText() Object {
	if pg.Parsed && pg.Text != "" {
		return NewString(pg.Text)
	}

	var result strings.Builder

	for _, contentNum := range pg.Contents {
		obj, ok := pg.PDF.Objects[contentNum]
		if !ok {
			continue
		}

		stream, ok := obj.Value.(*PDFStream)
		if !ok {
			continue
		}

		data, err := pg.PDF.decodeStream(stream)
		if err != nil {
			continue
		}

		text := pg.extractTextFromStream(data)
		result.WriteString(text)
	}

	pg.Text = result.String()
	pg.Parsed = true

	return NewString(pg.Text)
}

// extractTextFromStream extracts text from a content stream.
func (pg *PDFPage) extractTextFromStream(data []byte) string {
	var result strings.Builder

	// Simple text extraction - parse Tj and TJ operators
	content := string(data)

	// Regular expressions for text operators
	tjRegex := regexp.MustCompile(`\(([^)]*)\)\s*Tj`)
	tjMatches := tjRegex.FindAllStringSubmatch(content, -1)
	for _, match := range tjMatches {
		if len(match) > 1 {
			result.WriteString(decodePDFString(match[1]))
		}
	}

	// Handle array text (TJ operator)
	tjArrayRegex := regexp.MustCompile(`\[([^\]]*)\]\s*TJ`)
	tjArrayMatches := tjArrayRegex.FindAllStringSubmatch(content, -1)
	for _, match := range tjArrayMatches {
		if len(match) > 1 {
			// Extract strings from array
			strRegex := regexp.MustCompile(`\(([^)]*)\)`)
			strMatches := strRegex.FindAllStringSubmatch(match[1], -1)
			for _, sm := range strMatches {
				if len(sm) > 1 {
					result.WriteString(decodePDFString(sm[1]))
				}
			}
		}
	}

	return result.String()
}

// ============================================================
// PDFDocument Methods
// ============================================================

// NewPDFDocument creates a new PDF document.
func NewPDFDocument() *PDFDocument {
	return &PDFDocument{
		Version:     "1.4",
		Pages:       make([]*PDFPageData, 0),
		NextObjNum:  1,
		Objects:     make([]interface{}, 0),
		DefaultFont: "Helvetica",
		FontSize:    12,
		Producer:    "Xxlang PDF Generator",
	}
}

// AddPage adds a new page to the document.
func (d *PDFDocument) AddPage(width, height float64) Object {
	d.mu.Lock()
	defer d.mu.Unlock()

	page := &PDFPageData{
		Width:     width,
		Height:    height,
		Contents:  &strings.Builder{},
		Resources: make(map[string]interface{}),
	}

	// Initialize content stream
	page.Contents.WriteString("BT\n")

	d.Pages = append(d.Pages, page)
	return NewInt(int64(len(d.Pages) - 1))
}

// WriteText writes text to a page.
func (d *PDFDocument) WriteText(pageIndex int, text string, x, y float64, opts map[string]interface{}) Object {
	d.mu.Lock()
	defer d.mu.Unlock()

	if pageIndex < 0 || pageIndex >= len(d.Pages) {
		return newError("page index out of range")
	}

	page := d.Pages[pageIndex]

	// Get options
	fontSize := d.FontSize
	if fs, ok := opts["fontSize"]; ok {
		switch v := fs.(type) {
		case float64:
			fontSize = v
		case int64:
			fontSize = float64(v)
		case *Float:
			fontSize = v.Value
		case *Int:
			fontSize = float64(v.Value)
		}
	}

	font := d.DefaultFont
	if f, ok := opts["font"]; ok {
		if s, ok := f.(*String); ok {
			font = s.Value
		} else if s, ok := f.(string); ok {
			font = s
		}
	}

	// Escape text for PDF
	escaped := escapePDFString(text)

	// Write text commands
	fmt.Fprintf(page.Contents, "/%s %.1f Tf\n", font, fontSize)
	fmt.Fprintf(page.Contents, "%.1f %.1f Td\n", x, y)
	fmt.Fprintf(page.Contents, "(%s) Tj\n", escaped)

	return NULL
}

// SetFont sets the default font and size.
func (d *PDFDocument) SetFont(name string, size float64) Object {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.DefaultFont = name
	d.FontSize = size
	return NULL
}

// Save saves the document to a file.
func (d *PDFDocument) Save(filePath string) Object {
	data, err := d.render()
	if err != nil {
		return newError("failed to save PDF: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return newError("failed to write file: %v", err)
	}

	return NULL
}

// ToBytes returns the document as a byte array.
func (d *PDFDocument) ToBytes() Object {
	data, err := d.render()
	if err != nil {
		return newError("failed to render PDF: %v", err)
	}

	elements := make([]Object, len(data))
	for i, b := range data {
		elements[i] = NewInt(int64(b))
	}
	return NewArray(elements)
}

// render renders the PDF document to bytes.
func (d *PDFDocument) render() ([]byte, error) {
	var buf bytes.Buffer

	// Header
	fmt.Fprintf(&buf, "%%PDF-%s\n", d.Version)
	buf.WriteString("%\xe2\xe3\xcf\xd3\n") // Binary marker

	// Track object positions
	positions := make(map[int64]int64)

	// Object 1: Catalog
	positions[1] = int64(buf.Len())
	fmt.Fprintf(&buf, "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n\n")

	// Object 2: Pages
	positions[2] = int64(buf.Len())
	fmt.Fprintf(&buf, "2 0 obj\n<< /Type /Pages /Kids [")
	for i := 0; i < len(d.Pages); i++ {
		fmt.Fprintf(&buf, "%d 0 R ", i+3)
	}
	fmt.Fprintf(&buf, "] /Count %d >>\nendobj\n\n", len(d.Pages))

	// Page objects (starting at object 3)
	for i, page := range d.Pages {
		// End text block if not done
		if !strings.HasSuffix(page.Contents.String(), "ET") {
			page.Contents.WriteString("\nET")
		}

		positions[int64(i+3)] = int64(buf.Len())

		// Page object
		fmt.Fprintf(&buf, "%d 0 obj\n", i+3)
		fmt.Fprintf(&buf, "<< /Type /Page /Parent 2 0 R ")
		fmt.Fprintf(&buf, "/MediaBox [0 0 %.1f %.1f] ", page.Width, page.Height)
		fmt.Fprintf(&buf, "/Contents %d 0 R ", len(d.Pages)+i+3)
		fmt.Fprintf(&buf, "/Resources << /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >> ")
		fmt.Fprintf(&buf, ">>\nendobj\n\n")
	}

	// Content stream objects
	for i, page := range d.Pages {
		content := page.Contents.String()
		positions[int64(len(d.Pages)+i+3)] = int64(buf.Len())

		fmt.Fprintf(&buf, "%d 0 obj\n", len(d.Pages)+i+3)
		fmt.Fprintf(&buf, "<< /Length %d >>\n", len(content))
		fmt.Fprintf(&buf, "stream\n%s\nendstream\n", content)
		fmt.Fprintf(&buf, "endobj\n\n")
	}

	// Info object (if metadata set)
	infoObjNum := int64(len(d.Pages)*2 + 3)
	if d.Title != "" || d.Author != "" || d.Producer != "" {
		positions[infoObjNum] = int64(buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n<< ", infoObjNum)
		if d.Title != "" {
			fmt.Fprintf(&buf, "/Title (%s) ", escapePDFString(d.Title))
		}
		if d.Author != "" {
			fmt.Fprintf(&buf, "/Author (%s) ", escapePDFString(d.Author))
		}
		if d.Subject != "" {
			fmt.Fprintf(&buf, "/Subject (%s) ", escapePDFString(d.Subject))
		}
		fmt.Fprintf(&buf, "/Creator (%s) ", escapePDFString(d.Creator))
		fmt.Fprintf(&buf, "/Producer (%s) ", escapePDFString(d.Producer))
		fmt.Fprintf(&buf, "/CreationDate (D:%s) ", time.Now().Format("20060102150405"))
		fmt.Fprintf(&buf, ">>\nendobj\n\n")
	}

	// Cross-reference table
	xrefPos := int64(buf.Len())
	buf.WriteString("xref\n")
	fmt.Fprintf(&buf, "0 %d\n", infoObjNum+1)
	buf.WriteString("0000000000 65535 f \n")

	for i := int64(1); i <= infoObjNum; i++ {
		if pos, ok := positions[i]; ok {
			fmt.Fprintf(&buf, "%010d 00000 n \n", pos)
		} else {
			buf.WriteString("0000000000 00000 f \n")
		}
	}

	// Trailer
	buf.WriteString("trailer\n")
	fmt.Fprintf(&buf, "<< /Size %d /Root 1 0 R ", infoObjNum+1)
	if d.Title != "" || d.Author != "" {
		fmt.Fprintf(&buf, "/Info %d 0 R ", infoObjNum)
	}
	buf.WriteString(">>\n")

	// Startxref
	fmt.Fprintf(&buf, "startxref\n%d\n%%%%EOF\n", xrefPos)

	return buf.Bytes(), nil
}

// ============================================================
// Helper Functions
// ============================================================

// decodeStream decodes a PDF stream based on its /Filter.
func (p *PDF) decodeStream(stream *PDFStream) ([]byte, error) {
	data := stream.RawData

	if stream.Dict == nil {
		return data, nil
	}

	filter, ok := stream.Dict.Entries["Filter"]
	if !ok {
		return data, nil
	}

	// Handle single filter
	switch f := filter.(type) {
	case PDFName:
		return p.applyFilter(data, string(f))
	case *PDFArray:
		// Apply multiple filters in order
		for _, elem := range f.Elements {
			if name, ok := elem.(PDFName); ok {
				var err error
				data, err = p.applyFilter(data, string(name))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return data, nil
}

// applyFilter applies a specific filter to data.
func (p *PDF) applyFilter(data []byte, filter string) ([]byte, error) {
	switch filter {
	case "FlateDecode":
		return p.flateDecode(data)
	case "ASCII85Decode":
		return p.ascii85Decode(data)
	case "ASCIIHexDecode":
		return p.asciiHexDecode(data)
	default:
		return data, nil
	}
}

// flateDecode decompresses zlib/deflate data.
func (p *PDF) flateDecode(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// ascii85Decode decodes ASCII85 encoded data.
func (p *PDF) ascii85Decode(data []byte) ([]byte, error) {
	decoder := ascii85.NewDecoder(bytes.NewReader(data))
	return io.ReadAll(decoder)
}

// asciiHexDecode decodes ASCII hex encoded data.
func (p *PDF) asciiHexDecode(data []byte) ([]byte, error) {
	// Remove whitespace
	clean := bytes.TrimSpace(data)
	return hex.DecodeString(string(clean))
}

// pdfIsWhitespace checks if a byte is PDF whitespace.
func pdfIsWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '\f'
}

// pdfIsDigit checks if a byte is a digit.
func pdfIsDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// isOctalDigit checks if a byte is an octal digit.
func isOctalDigit(b byte) bool {
	return b >= '0' && b <= '7'
}

// isAlphaNum checks if a byte is alphanumeric.
func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || pdfIsDigit(b)
}

// readInt reads a big-endian integer from bytes.
func readInt(data []byte) int64 {
	var result int64
	for _, b := range data {
		result = (result << 8) | int64(b)
	}
	return result
}

// getFloat gets a float value from an interface.
func getFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	default:
		return 0
	}
}

// pdfStringToString converts a PDF string to a Go string.
func pdfStringToString(v interface{}) string {
	switch val := v.(type) {
	case PDFString:
		return string(val)
	case PDFHexStr:
		decoded, _ := hex.DecodeString(string(val))
		return string(decoded)
	case string:
		return val
	default:
		return ""
	}
}

// pdfDateToString converts a PDF date string to a readable format.
func pdfDateToString(v interface{}) string {
	s := pdfStringToString(v)
	if strings.HasPrefix(s, "D:") {
		// PDF date format: D:YYYYMMDDHHmmSS
		if len(s) >= 15 {
			return s[2:6] + "-" + s[6:8] + "-" + s[8:10] + " " + s[10:12] + ":" + s[12:14]
		}
	}
	return s
}

// decodePDFString decodes escape sequences in a PDF string.
func decodePDFString(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			case '(':
				result.WriteByte('(')
			case ')':
				result.WriteByte(')')
			case '\\':
				result.WriteByte('\\')
			default:
				if isOctalDigit(s[i]) {
					// Octal escape
					octal := s[i] - '0'
					if i+1 < len(s) && isOctalDigit(s[i+1]) {
						i++
						octal = octal*8 + s[i] - '0'
						if i+1 < len(s) && isOctalDigit(s[i+1]) {
							i++
							octal = octal*8 + s[i] - '0'
						}
					}
					result.WriteByte(octal)
				} else {
					result.WriteByte(s[i])
				}
			}
		} else {
			result.WriteByte(s[i])
		}
		i++
	}

	// Handle UTF-16 BE BOM
	decoded := result.String()
	if strings.HasPrefix(decoded, "\xfe\xff") {
		// UTF-16 BE
		u16 := make([]uint16, 0, len(decoded)/2)
		for i := 2; i+1 < len(decoded); i += 2 {
			u16 = append(u16, uint16(decoded[i])<<8|uint16(decoded[i+1]))
		}
		return string(utf16.Decode(u16))
	}

	return decoded
}

// escapePDFString escapes special characters for a PDF string.
func escapePDFString(s string) string {
	var result strings.Builder
	for _, ch := range s {
		switch ch {
		case '(':
			result.WriteString("\\(")
		case ')':
			result.WriteString("\\)")
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}