// pkg/objects/docx.go
// DOCX object types for Xxlang - Microsoft Word document handling.
package objects

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

// DocxDocument represents a Microsoft Word document (.docx).
// A .docx file is a ZIP archive containing XML files.
type DocxDocument struct {
	// File path (empty for new documents)
	filePath string

	// Core content - references to XML structures
	xmlDoc *XMLDocument

	// ZIP reader for open documents
	zipReader *zip.ReadCloser

	// ZIP buffer for documents created from bytes
	zipData []byte

	// Relationships: ID -> target
	relationships map[string]string

	// Content types: extension/partName -> contentType
	contentTypes map[string]string

	// Media files: filename -> data
	mediaFiles map[string][]byte

	// Styles XML document
	stylesXML *XMLDocument

	// Numbering XML document
	numberingXML *XMLDocument

	// Document properties
	properties *DocxProperties

	// Headers and footers: ID -> XML content
	headers map[string]*XMLDocument
	footers map[string]*XMLDocument

	// Modified flag
	modified bool
}

// DocxProperties holds document metadata.
type DocxProperties struct {
	Title       string
	Subject     string
	Author      string
	Keywords    string
	Description string
	Created     string
	Modified    string
	LastModBy   string
	Revision    int
	Category    string
	ContentStat string
	Application string
	AppVersion  string
	Pages       int
	Words       int
	Characters  int
	Lines       int
	Paragraphs  int
}

// DocxParagraph represents a paragraph in a Word document.
type DocxParagraph struct {
	Document *DocxDocument
	XmlNode  *XMLNode

	// Cached properties
	alignment string
	style     string
}

// DocxRun represents a text run within a paragraph.
type DocxRun struct {
	Paragraph *DocxParagraph
	XmlNode   *XMLNode

	// Cached properties
	text      string
	bold      bool
	italic    bool
	underline string
	fontSize  int
	fontName  string
	color     string
}

// DocxTable represents a table in a Word document.
type DocxTable struct {
	Document *DocxDocument
	XmlNode  *XMLNode

	// Cached dimensions
	Rows int
	Cols int
}

// DocxTableRow represents a table row.
type DocxTableRow struct {
	Table   *DocxTable
	XmlNode *XMLNode
	Index   int
}

// DocxTableCell represents a table cell.
type DocxTableCell struct {
	Row     *DocxTableRow
	XmlNode *XMLNode
	ColIdx  int
}

// DocxImage represents an image in a Word document.
type DocxImage struct {
	document   *DocxDocument
	relationID string

	// Image data
	data   []byte
	format string

	// Dimensions in EMUs
	width  int
	height int

	// Positioning
	position string // "inline" or "floating"
	x        int
	y        int
}

// DocxSection represents a document section.
type DocxSection struct {
	document *DocxDocument
	xmlNode  *XMLNode
}

// DocxHeader represents a document header.
type DocxHeader struct {
	document   *DocxDocument
	relationID string
	xmlDoc     *XMLDocument
}

// DocxFooter represents a document footer.
type DocxFooter struct {
	document   *DocxDocument
	relationID string
	xmlDoc     *XMLDocument
}

// DocxStyle represents a document style.
type DocxStyle struct {
	document *DocxDocument
	xmlNode  *XMLNode

	id   string
	name string
}

// DocxHyperlink represents a hyperlink in the document.
type DocxHyperlink struct {
	document   *DocxDocument
	xmlNode    *XMLNode
	relationID string
	target     string
	text       string
}

// DocxBookmark represents a bookmark in the document.
type DocxBookmark struct {
	document *DocxDocument
	xmlNode  *XMLNode
	name     string
	id       int
}

// DocxTOC represents a table of contents.
type DocxTOC struct {
	document *DocxDocument
	xmlNode  *XMLNode
}

// DocxTextBox represents a text box.
type DocxTextBox struct {
	document *DocxDocument
	xmlNode  *XMLNode
}

// DocxShape represents a shape.
type DocxShape struct {
	document *DocxDocument
	xmlNode  *XMLNode
}

// DocxChart represents a chart.
type DocxChart struct {
	document   *DocxDocument
	relationID string
	xmlNode    *XMLNode
}

// DocxComment represents a comment.
type DocxComment struct {
	document *DocxDocument
	id       int
	author   string
	date     string
	content  []*DocxParagraph
}

// DocxRevision represents a revision (tracked change).
type DocxRevision struct {
	document *DocxDocument
	id       int
	revType  string // "insert" or "delete"
	author   string
	date     string
	content  string
}

// DocxFootnote represents a footnote.
type DocxFootnote struct {
	document   *DocxDocument
	id         int
	paragraphs []*DocxParagraph
}

// DocxEndnote represents an endnote.
type DocxEndnote struct {
	document   *DocxDocument
	id         int
	paragraphs []*DocxParagraph
}

// Type implementations for DocxDocument
func (d *DocxDocument) Type() ObjectType { return DocxDocumentType }
func (d *DocxDocument) TypeTag() TypeTag { return TagDocxDocument }
func (d *DocxDocument) ToBool() *Bool    { return TRUE }
func (d *DocxDocument) HashKey() HashKey {
	return HashKey{Type: DocxDocumentType, Value: uint64(uintptr(unsafe.Pointer(d)))}
}
func (d *DocxDocument) Inspect() string {
	if d.filePath != "" {
		return fmt.Sprintf("DocxDocument(path=%s)", d.filePath)
	}
	return "DocxDocument(new)"
}

// Type implementations for DocxParagraph
func (p *DocxParagraph) Type() ObjectType { return DocxParagraphType }
func (p *DocxParagraph) TypeTag() TypeTag { return TagDocxParagraph }
func (p *DocxParagraph) ToBool() *Bool    { return TRUE }
func (p *DocxParagraph) HashKey() HashKey {
	return HashKey{Type: DocxParagraphType, Value: uint64(uintptr(unsafe.Pointer(p)))}
}
func (p *DocxParagraph) Inspect() string {
	return "DocxParagraph()"
}

// Type implementations for DocxRun
func (r *DocxRun) Type() ObjectType { return DocxRunType }
func (r *DocxRun) TypeTag() TypeTag { return TagDocxRun }
func (r *DocxRun) ToBool() *Bool    { return TRUE }
func (r *DocxRun) HashKey() HashKey {
	return HashKey{Type: DocxRunType, Value: uint64(uintptr(unsafe.Pointer(r)))}
}
func (r *DocxRun) Inspect() string {
	return fmt.Sprintf("DocxRun(text=%q)", r.text)
}

// Type implementations for DocxTable
func (t *DocxTable) Type() ObjectType { return DocxTableType }
func (t *DocxTable) TypeTag() TypeTag { return TagDocxTable }
func (t *DocxTable) ToBool() *Bool    { return TRUE }
func (t *DocxTable) HashKey() HashKey {
	return HashKey{Type: DocxTableType, Value: uint64(uintptr(unsafe.Pointer(t)))}
}
func (t *DocxTable) Inspect() string {
	return fmt.Sprintf("DocxTable(rows=%d, cols=%d)", t.Rows, t.Cols)
}

// Type implementations for DocxTableRow
func (r *DocxTableRow) Type() ObjectType { return DocxTableRowType }
func (r *DocxTableRow) TypeTag() TypeTag { return TagDocxTableRow }
func (r *DocxTableRow) ToBool() *Bool    { return TRUE }
func (r *DocxTableRow) HashKey() HashKey {
	return HashKey{Type: DocxTableRowType, Value: uint64(uintptr(unsafe.Pointer(r)))}
}
func (r *DocxTableRow) Inspect() string {
	return fmt.Sprintf("DocxTableRow(index=%d)", r.Index)
}

// Type implementations for DocxTableCell
func (c *DocxTableCell) Type() ObjectType { return DocxTableCellType }
func (c *DocxTableCell) TypeTag() TypeTag { return TagDocxTableCell }
func (c *DocxTableCell) ToBool() *Bool    { return TRUE }
func (c *DocxTableCell) HashKey() HashKey {
	return HashKey{Type: DocxTableCellType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}
func (c *DocxTableCell) Inspect() string {
	return fmt.Sprintf("DocxTableCell(col=%d)", c.ColIdx)
}

// Type implementations for DocxImage
func (i *DocxImage) Type() ObjectType { return DocxImageType }
func (i *DocxImage) TypeTag() TypeTag { return TagDocxImage }
func (i *DocxImage) ToBool() *Bool    { return TRUE }
func (i *DocxImage) HashKey() HashKey {
	return HashKey{Type: DocxImageType, Value: uint64(uintptr(unsafe.Pointer(i)))}
}
func (i *DocxImage) Inspect() string {
	return fmt.Sprintf("DocxImage(format=%s, %dx%d)", i.format, i.width, i.height)
}

// Type implementations for DocxSection
func (s *DocxSection) Type() ObjectType { return DocxSectionType }
func (s *DocxSection) TypeTag() TypeTag { return TagDocxSection }
func (s *DocxSection) ToBool() *Bool    { return TRUE }
func (s *DocxSection) HashKey() HashKey {
	return HashKey{Type: DocxSectionType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *DocxSection) Inspect() string { return "DocxSection()" }

// Type implementations for DocxHeader
func (h *DocxHeader) Type() ObjectType { return DocxHeaderType }
func (h *DocxHeader) TypeTag() TypeTag { return TagDocxHeader }
func (h *DocxHeader) ToBool() *Bool    { return TRUE }
func (h *DocxHeader) HashKey() HashKey {
	return HashKey{Type: DocxHeaderType, Value: uint64(uintptr(unsafe.Pointer(h)))}
}
func (h *DocxHeader) Inspect() string { return "DocxHeader()" }

// Type implementations for DocxFooter
func (f *DocxFooter) Type() ObjectType { return DocxFooterType }
func (f *DocxFooter) TypeTag() TypeTag { return TagDocxFooter }
func (f *DocxFooter) ToBool() *Bool    { return TRUE }
func (f *DocxFooter) HashKey() HashKey {
	return HashKey{Type: DocxFooterType, Value: uint64(uintptr(unsafe.Pointer(f)))}
}
func (f *DocxFooter) Inspect() string { return "DocxFooter()" }

// Type implementations for DocxStyle
func (s *DocxStyle) Type() ObjectType { return DocxStyleType }
func (s *DocxStyle) TypeTag() TypeTag { return TagDocxStyle }
func (s *DocxStyle) ToBool() *Bool    { return TRUE }
func (s *DocxStyle) HashKey() HashKey {
	return HashKey{Type: DocxStyleType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *DocxStyle) Inspect() string {
	return fmt.Sprintf("DocxStyle(id=%s, name=%s)", s.id, s.name)
}

// Type implementations for DocxHyperlink
func (h *DocxHyperlink) Type() ObjectType { return DocxHyperlinkType }
func (h *DocxHyperlink) TypeTag() TypeTag { return TagDocxHyperlink }
func (h *DocxHyperlink) ToBool() *Bool    { return TRUE }
func (h *DocxHyperlink) HashKey() HashKey {
	return HashKey{Type: DocxHyperlinkType, Value: uint64(uintptr(unsafe.Pointer(h)))}
}
func (h *DocxHyperlink) Inspect() string {
	return fmt.Sprintf("DocxHyperlink(target=%s)", h.target)
}

// Type implementations for DocxBookmark
func (b *DocxBookmark) Type() ObjectType { return DocxBookmarkType }
func (b *DocxBookmark) TypeTag() TypeTag { return TagDocxBookmark }
func (b *DocxBookmark) ToBool() *Bool    { return TRUE }
func (b *DocxBookmark) HashKey() HashKey {
	return HashKey{Type: DocxBookmarkType, Value: uint64(uintptr(unsafe.Pointer(b)))}
}
func (b *DocxBookmark) Inspect() string {
	return fmt.Sprintf("DocxBookmark(name=%s)", b.name)
}

// Type implementations for DocxTOC
func (t *DocxTOC) Type() ObjectType { return DocxTOCType }
func (t *DocxTOC) TypeTag() TypeTag { return TagDocxTOC }
func (t *DocxTOC) ToBool() *Bool    { return TRUE }
func (t *DocxTOC) HashKey() HashKey {
	return HashKey{Type: DocxTOCType, Value: uint64(uintptr(unsafe.Pointer(t)))}
}
func (t *DocxTOC) Inspect() string { return "DocxTOC()" }

// Type implementations for DocxTextBox
func (t *DocxTextBox) Type() ObjectType { return DocxTextBoxType }
func (t *DocxTextBox) TypeTag() TypeTag { return TagDocxTextBox }
func (t *DocxTextBox) ToBool() *Bool    { return TRUE }
func (t *DocxTextBox) HashKey() HashKey {
	return HashKey{Type: DocxTextBoxType, Value: uint64(uintptr(unsafe.Pointer(t)))}
}
func (t *DocxTextBox) Inspect() string { return "DocxTextBox()" }

// Type implementations for DocxShape
func (s *DocxShape) Type() ObjectType { return DocxShapeType }
func (s *DocxShape) TypeTag() TypeTag { return TagDocxShape }
func (s *DocxShape) ToBool() *Bool    { return TRUE }
func (s *DocxShape) HashKey() HashKey {
	return HashKey{Type: DocxShapeType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *DocxShape) Inspect() string { return "DocxShape()" }

// Type implementations for DocxChart
func (c *DocxChart) Type() ObjectType { return DocxChartType }
func (c *DocxChart) TypeTag() TypeTag { return TagDocxChart }
func (c *DocxChart) ToBool() *Bool    { return TRUE }
func (c *DocxChart) HashKey() HashKey {
	return HashKey{Type: DocxChartType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}
func (c *DocxChart) Inspect() string { return "DocxChart()" }

// Type implementations for DocxComment
func (c *DocxComment) Type() ObjectType { return DocxCommentType }
func (c *DocxComment) TypeTag() TypeTag { return TagDocxComment }
func (c *DocxComment) ToBool() *Bool    { return TRUE }
func (c *DocxComment) HashKey() HashKey {
	return HashKey{Type: DocxCommentType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}
func (c *DocxComment) Inspect() string {
	return fmt.Sprintf("DocxComment(id=%d, author=%s)", c.id, c.author)
}

// Type implementations for DocxRevision
func (r *DocxRevision) Type() ObjectType { return DocxRevisionType }
func (r *DocxRevision) TypeTag() TypeTag { return TagDocxRevision }
func (r *DocxRevision) ToBool() *Bool    { return TRUE }
func (r *DocxRevision) HashKey() HashKey {
	return HashKey{Type: DocxRevisionType, Value: uint64(uintptr(unsafe.Pointer(r)))}
}
func (r *DocxRevision) Inspect() string {
	return fmt.Sprintf("DocxRevision(id=%d, type=%s)", r.id, r.revType)
}

// Type implementations for DocxFootnote
func (f *DocxFootnote) Type() ObjectType { return DocxFootnoteType }
func (f *DocxFootnote) TypeTag() TypeTag { return TagDocxFootnote }
func (f *DocxFootnote) ToBool() *Bool    { return TRUE }
func (f *DocxFootnote) HashKey() HashKey {
	return HashKey{Type: DocxFootnoteType, Value: uint64(uintptr(unsafe.Pointer(f)))}
}
func (f *DocxFootnote) Inspect() string {
	return fmt.Sprintf("DocxFootnote(id=%d)", f.id)
}

// Type implementations for DocxEndnote
func (e *DocxEndnote) Type() ObjectType { return DocxEndnoteType }
func (e *DocxEndnote) TypeTag() TypeTag { return TagDocxEndnote }
func (e *DocxEndnote) ToBool() *Bool    { return TRUE }
func (e *DocxEndnote) HashKey() HashKey {
	return HashKey{Type: DocxEndnoteType, Value: uint64(uintptr(unsafe.Pointer(e)))}
}
func (e *DocxEndnote) Inspect() string {
	return fmt.Sprintf("DocxEndnote(id=%d)", e.id)
}

// NewDocxDocument creates a new empty Word document.
func NewDocxDocument() *DocxDocument {
	doc := &DocxDocument{
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		headers:       make(map[string]*XMLDocument),
		footers:       make(map[string]*XMLDocument),
		properties:    &DocxProperties{},
	}

	// Create minimal document.xml structure
	doc.xmlDoc = NewXMLDocumentWithRoot("document")
	root := doc.xmlDoc.Root()
	root.SetName("w:document")

	// Add namespace attributes
	root.SetAttr("xmlns:w", NS_W)
	root.SetAttr("xmlns:r", NS_R)

	// Add body element
	body := NewXMLNode("w:body")
	root.AddChild(body)

	return doc
}

// OpenDocx opens an existing Word document from a file path.
func OpenDocx(path string) (*DocxDocument, error) {
	// Open ZIP file
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open docx file: %w", err)
	}

	doc := &DocxDocument{
		filePath:      path,
		zipReader:     reader,
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		headers:       make(map[string]*XMLDocument),
		footers:       make(map[string]*XMLDocument),
		properties:    &DocxProperties{},
	}

	// Parse ZIP contents
	if err := doc.parseZipContents(); err != nil {
		reader.Close()
		return nil, err
	}

	return doc, nil
}

// OpenDocxFromBytes opens a Word document from a byte slice.
func OpenDocxFromBytes(data []byte) (*DocxDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse docx data: %w", err)
	}

	doc := &DocxDocument{
		zipData:       data,
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		headers:       make(map[string]*XMLDocument),
		footers:       make(map[string]*XMLDocument),
		properties:    &DocxProperties{},
	}

	// Parse ZIP contents
	if err := doc.parseZipContentsFromReader(reader); err != nil {
		return nil, err
	}

	return doc, nil
}

// parseZipContents parses the ZIP file contents.
func (d *DocxDocument) parseZipContents() error {
	return d.parseZipContentsFromReader(&d.zipReader.Reader)
}

// parseZipContentsFromReader parses ZIP contents from a zip.Reader.
func (d *DocxDocument) parseZipContentsFromReader(reader *zip.Reader) error {
	for _, file := range reader.File {
		switch file.Name {
		case "word/document.xml":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			d.xmlDoc, err = ParseXML(string(data))
			if err != nil {
				return fmt.Errorf("failed to parse document.xml: %w", err)
			}

		case "word/styles.xml":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			d.stylesXML, err = ParseXML(string(data))
			if err != nil {
				return fmt.Errorf("failed to parse styles.xml: %w", err)
			}

		case "word/numbering.xml":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			d.numberingXML, err = ParseXML(string(data))
			if err != nil {
				return fmt.Errorf("failed to parse numbering.xml: %w", err)
			}

		case "word/_rels/document.xml.rels":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			if err := d.parseRelationships(data); err != nil {
				return err
			}

		case "[Content_Types].xml":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			if err := d.parseContentTypes(data); err != nil {
				return err
			}

		case "docProps/core.xml":
			data, err := d.readFileFromZip(file)
			if err != nil {
				return err
			}
			if err := d.parseCoreProperties(data); err != nil {
				return err
			}

		default:
			// Handle media files
			if strings.HasPrefix(file.Name, "word/media/") {
				data, err := d.readFileFromZip(file)
				if err != nil {
					return err
				}
				d.mediaFiles[filepath.Base(file.Name)] = data
			}
			// Handle headers
			if strings.HasPrefix(file.Name, "word/header") && strings.HasSuffix(file.Name, ".xml") {
				data, err := d.readFileFromZip(file)
				if err != nil {
					return err
				}
				xmlDoc, err := ParseXML(string(data))
				if err != nil {
					return err
				}
				relID := strings.TrimSuffix(strings.TrimPrefix(file.Name, "word/header"), ".xml")
				d.headers[relID] = xmlDoc
			}
			// Handle footers
			if strings.HasPrefix(file.Name, "word/footer") && strings.HasSuffix(file.Name, ".xml") {
				data, err := d.readFileFromZip(file)
				if err != nil {
					return err
				}
				xmlDoc, err := ParseXML(string(data))
				if err != nil {
					return err
				}
				relID := strings.TrimSuffix(strings.TrimPrefix(file.Name, "word/footer"), ".xml")
				d.footers[relID] = xmlDoc
			}
		}
	}

	return nil
}

// readFileFromZip reads a file from a ZIP archive.
func (d *DocxDocument) readFileFromZip(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// parseRelationships parses the document relationships.
func (d *DocxDocument) parseRelationships(data []byte) error {
	xmlDoc, err := ParseXML(string(data))
	if err != nil {
		return err
	}

	// Find all Relationship elements
	root := xmlDoc.Root()
	if root == nil {
		return nil
	}

	// Recursively find Relationship nodes
	var findRelationships func(node *XMLNode)
	findRelationships = func(node *XMLNode) {
		if strings.HasSuffix(node.Name(), "Relationship") {
			id := node.Attr("Id")
			target := node.Attr("Target")
			if id != "" && target != "" {
				d.relationships[id] = target
			}
		}
		for _, child := range node.children {
			findRelationships(child)
		}
	}
	findRelationships(root)

	return nil
}

// parseContentTypes parses the content types.
func (d *DocxDocument) parseContentTypes(data []byte) error {
	xmlDoc, err := ParseXML(string(data))
	if err != nil {
		return err
	}

	root := xmlDoc.Root()
	if root == nil {
		return nil
	}

	// Find Override elements
	var findOverrides func(node *XMLNode)
	findOverrides = func(node *XMLNode) {
		if strings.HasSuffix(node.Name(), "Override") {
			partName := node.Attr("PartName")
			contentType := node.Attr("ContentType")
			if partName != "" && contentType != "" {
				d.contentTypes[partName] = contentType
			}
		}
		if strings.HasSuffix(node.Name(), "Default") {
			ext := node.Attr("Extension")
			contentType := node.Attr("ContentType")
			if ext != "" && contentType != "" {
				d.contentTypes["."+ext] = contentType
			}
		}
		for _, child := range node.children {
			findOverrides(child)
		}
	}
	findOverrides(root)

	return nil
}

// parseCoreProperties parses document core properties.
func (d *DocxDocument) parseCoreProperties(data []byte) error {
	xmlDoc, err := ParseXML(string(data))
	if err != nil {
		return err
	}

	root := xmlDoc.Root()
	if root == nil {
		return nil
	}

	// Helper to find text content of a child element
	findText := func(parent *XMLNode, name string) string {
		for _, child := range parent.children {
			if strings.HasSuffix(child.Name(), name) {
				return child.Text()
			}
		}
		return ""
	}

	d.properties.Title = findText(root, "title")
	d.properties.Subject = findText(root, "subject")
	d.properties.Author = findText(root, "creator")
	d.properties.Keywords = findText(root, "keywords")
	d.properties.Description = findText(root, "description")
	d.properties.Created = findText(root, "created")
	d.properties.Modified = findText(root, "modified")
	d.properties.LastModBy = findText(root, "lastModifiedBy")

	return nil
}

// Close closes the document and releases resources.
func (d *DocxDocument) Close() error {
	if d.zipReader != nil {
		return d.zipReader.Close()
	}
	return nil
}

// SetModified sets the modified flag.
func (d *DocxDocument) SetModified(modified bool) {
	d.modified = modified
}

// IsModified returns whether the document has been modified.
func (d *DocxDocument) IsModified() bool {
	return d.modified
}

// GetProperties returns the document properties.
func (d *DocxDocument) GetProperties() *DocxProperties {
	return d.properties
}

// GetXMLDoc returns the underlying XML document.
func (d *DocxDocument) GetXMLDoc() *XMLDocument {
	return d.xmlDoc
}

// Save saves the document to a file.
func (d *DocxDocument) Save(path string) error {
	data, err := d.ToBytes()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ToBytes converts the document to a byte slice.
func (d *DocxDocument) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)

	// Write [Content_Types].xml
	if err := d.writeContentTypes(writer); err != nil {
		return nil, err
	}

	// Write _rels/.rels
	if err := d.writeRels(writer); err != nil {
		return nil, err
	}

	// Write word/document.xml
	if err := d.writeDocumentXML(writer); err != nil {
		return nil, err
	}

	// Write word/_rels/document.xml.rels
	if err := d.writeDocumentRels(writer); err != nil {
		return nil, err
	}

	// Write styles.xml
	if d.stylesXML != nil {
		if err := d.writeStylesXML(writer); err != nil {
			return nil, err
		}
	}

	// Write numbering.xml
	if d.numberingXML != nil {
		if err := d.writeNumberingXML(writer); err != nil {
			return nil, err
		}
	}

	// Write media files
	for name, data := range d.mediaFiles {
		w, err := writer.Create("word/media/" + name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
	}

	// Write docProps/core.xml
	if err := d.writeCoreProps(writer); err != nil {
		return nil, err
	}

	// Write headers
	for relID, xmlDoc := range d.headers {
		w, err := writer.Create("word/header" + relID + ".xml")
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(xmlDoc.ToString())); err != nil {
			return nil, err
		}
	}

	// Write footers
	for relID, xmlDoc := range d.footers {
		w, err := writer.Create("word/footer" + relID + ".xml")
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(xmlDoc.ToString())); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// writeContentTypes writes [Content_Types].xml.
func (d *DocxDocument) writeContentTypes(w *zip.Writer) error {
	f, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
</Types>`

	_, err = f.Write([]byte(content))
	return err
}

// writeRels writes _rels/.rels.
func (d *DocxDocument) writeRels(w *zip.Writer) error {
	f, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`

	_, err = f.Write([]byte(content))
	return err
}

// writeDocumentXML writes word/document.xml.
func (d *DocxDocument) writeDocumentXML(w *zip.Writer) error {
	f, err := w.Create("word/document.xml")
	if err != nil {
		return err
	}

	if d.xmlDoc == nil {
		return fmt.Errorf("document has no content")
	}

	_, err = f.Write([]byte(d.xmlDoc.ToString()))
	return err
}

// writeDocumentRels writes word/_rels/document.xml.rels.
func (d *DocxDocument) writeDocumentRels(w *zip.Writer) error {
	f, err := w.Create("word/_rels/document.xml.rels")
	if err != nil {
		return err
	}

	// Build relationships XML
	var rels strings.Builder
	rels.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
`)

	// Add style relationship
	rels.WriteString(`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
`)

	// Add media relationships
	mediaIdx := 2
	for id, target := range d.relationships {
		if strings.HasPrefix(target, "media/") {
			rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>
`, id, target))
		}
	}

	// Add header/footer relationships
	for relID := range d.headers {
		rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header%s.xml"/>
`, relID, relID))
		mediaIdx++
	}
	for relID := range d.footers {
		rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer%s.xml"/>
`, relID, relID))
		mediaIdx++
	}

	rels.WriteString("</Relationships>")

	_, err = f.Write([]byte(rels.String()))
	return err
}

// writeStylesXML writes word/styles.xml.
func (d *DocxDocument) writeStylesXML(w *zip.Writer) error {
	f, err := w.Create("word/styles.xml")
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(d.stylesXML.ToString()))
	return err
}

// writeNumberingXML writes word/numbering.xml.
func (d *DocxDocument) writeNumberingXML(w *zip.Writer) error {
	f, err := w.Create("word/numbering.xml")
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(d.numberingXML.ToString()))
	return err
}

// writeCoreProps writes docProps/core.xml.
func (d *DocxDocument) writeCoreProps(w *zip.Writer) error {
	f, err := w.Create("docProps/core.xml")
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
<dc:title>%s</dc:title>
<dc:subject>%s</dc:subject>
<dc:creator>%s</dc:creator>
<cp:keywords>%s</cp:keywords>
<dc:description>%s</dc:description>
<dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>
<dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>
<cp:lastModifiedBy>%s</cp:lastModifiedBy>
</cp:coreProperties>`,
		EscapeXMLText(d.properties.Title),
		EscapeXMLText(d.properties.Subject),
		EscapeXMLText(d.properties.Author),
		EscapeXMLText(d.properties.Keywords),
		EscapeXMLText(d.properties.Description),
		d.properties.Created,
		d.properties.Modified,
		EscapeXMLText(d.properties.LastModBy))

	_, err = f.Write([]byte(content))
	return err
}

// GetBody returns the body element of the document.
func (d *DocxDocument) GetBody() *XMLNode {
	if d.xmlDoc == nil {
		return nil
	}
	// Try to find body with namespace prefix
	body := d.xmlDoc.FindFirst("//w:body")
	if body != nil {
		return body
	}
	// Try without namespace prefix (for parsed documents)
	return d.xmlDoc.FindFirst("//body")
}

// FindElements finds elements by name, trying both with and without namespace prefix.
func (d *DocxDocument) FindElements(path string) *Array {
	if d.xmlDoc == nil {
		return &Array{}
	}
	// Try with namespace prefix
	result := d.xmlDoc.Find(path)
	if len(result.Elements) > 0 {
		return result
	}
	// Try without namespace prefix
	unprefixedPath := strings.ReplaceAll(path, "w:", "")
	return d.xmlDoc.Find(unprefixedPath)
}

// FindFirstElement finds the first element by name, trying both with and without namespace prefix.
func (d *DocxDocument) FindFirstElement(path string) *XMLNode {
	if d.xmlDoc == nil {
		return nil
	}
	// Try with namespace prefix
	result := d.xmlDoc.FindFirst(path)
	if result != nil {
		return result
	}
	// Try without namespace prefix
	unprefixedPath := strings.ReplaceAll(path, "w:", "")
	return d.xmlDoc.FindFirst(unprefixedPath)
}

// AddRelationship adds a relationship to the document.
func (d *DocxDocument) AddRelationship(relType, target string) string {
	// Generate a new relationship ID
	id := fmt.Sprintf("rId%d", len(d.relationships)+1)
	d.relationships[id] = target
	return id
}

// AddMediaFile adds a media file to the document.
func (d *DocxDocument) AddMediaFile(data []byte, ext string) string {
	name := fmt.Sprintf("image%d.%s", len(d.mediaFiles)+1, ext)
	d.mediaFiles[name] = data
	return name
}

// XML Namespaces for DOCX
const (
	NS_W    = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	NS_R    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	NS_WP   = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	NS_A    = "http://schemas.openxmlformats.org/drawingml/2006/main"
	NS_PIC  = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	NS_V    = "urn:schemas-microsoft-com:vml"
	NS_O    = "urn:schemas-microsoft-com:office:office"
	NS_W10  = "urn:schemas-microsoft-com:office:word"
	NS_W14  = "http://schemas.microsoft.com/office/word/2010/wordml"
	NS_W15  = "http://schemas.microsoft.com/office/word/2012/wordml"
	NS_MC   = "http://schemas.openxmlformats.org/markup-compatibility/2006"
	NS_CT   = "http://schemas.openxmlformats.org/package/2006/content-types"
	NS_RELS = "http://schemas.openxmlformats.org/package/2006/relationships"
)