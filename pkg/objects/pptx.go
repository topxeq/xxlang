// pkg/objects/pptx.go
// PPTX object types for Xxlang - PowerPoint file handling.
package objects

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

// XML Namespaces for PPTX
const (
	PPTX_NS_A       = "http://schemas.openxmlformats.org/drawingml/2006/main"
	PPTX_NS_R       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	PPTX_NS_P       = "http://schemas.openxmlformats.org/presentationml/2006/main"
	PPTX_NS_PIC     = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	PPTX_NS_C       = "http://schemas.openxmlformats.org/drawingml/2006/chart"
	PPTX_NS_C_R     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart"
	PPTX_NS_RELS    = "http://schemas.openxmlformats.org/package/2006/relationships"
	PPTX_NS_CONTENT = "http://schemas.openxmlformats.org/package/2006/content-types"
)

// EMU conversion constants (English Metric Units)
const (
	EMU_PER_INCH  = 914400
	EMU_PER_POINT = 12700
	EMU_PER_PIXEL = 9525 // at 96 DPI
)

// PPTXDocument represents a PowerPoint presentation.
// A .pptx file is a ZIP archive containing XML files.
type PPTXDocument struct {
	// File path (empty for new documents)
	filePath string

	// ZIP reader for open documents
	zipReader *zip.ReadCloser

	// ZIP data for documents created from bytes
	zipData []byte

	// Relationships: ID -> target
	relationships map[string]string

	// Content types: extension/partName -> contentType
	contentTypes map[string]string

	// Media files: filename -> data
	mediaFiles map[string][]byte

	// Slides in the presentation
	slides []*PPTXSlide

	// Slide layouts
	slideLayouts []*PPTXSlideLayout

	// Slide masters
	slideMasters []*PPTXSlideMaster

	// Themes
	themes []*PPTXTheme

	// Document properties
	properties *PPTXProperties

	// Modified flag
	modified bool
}

// PPTXProperties holds document metadata.
type PPTXProperties struct {
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
	Company     string
}

// PPTXSlide represents a single slide in a presentation.
type PPTXSlide struct {
	document   *PPTXDocument
	index      int
	layout     *PPTXSlideLayout
	shapes     []*PPTXShape
	textFrames []*PPTXTextFrame
	images     []*PPTXImage
	tables     []*PPTXTable
	charts     []*PPTXChart
	videos     []*PPTXVideo
	audios     []*PPTXAudio
	background *PPTXSlideBackground
	notes      string
	xmlData    []byte
}

// PPTXTextFrame represents a text box on a slide.
type PPTXTextFrame struct {
	slide      *PPTXSlide
	text       string
	paragraphs []*PPTXParagraph
	position   PPTXPosition
	size       PPTXSize
	rotation   float64
}

// PPTXParagraph represents a paragraph within a text frame.
type PPTXParagraph struct {
	frame     *PPTXTextFrame
	runs      []*PPTXTextRun
	bullet    *PPTXBulletStyle
	alignment string // "left", "center", "right", "justify"
	spacing   float64
	level     int
}

// PPTXTextRun represents a text run with uniform formatting.
type PPTXTextRun struct {
	paragraph *PPTXParagraph
	text      string
	fontName  string
	fontSize  int
	color     PPTXColor
	bold      bool
	italic    bool
	underline string // "none", "single", "double"
	strike    bool
}

// PPTXShapeKind defines shape types.
type PPTXShapeKind string

const (
	PPTXShapeRectangle  PPTXShapeKind = "rect"
	PPTXShapeOval       PPTXShapeKind = "ellipse"
	PPTXShapeRoundRect  PPTXShapeKind = "roundRect"
	PPTXShapeTriangle   PPTXShapeKind = "triangle"
	PPTXShapeLine       PPTXShapeKind = "line"
	PPTXShapeArrow      PPTXShapeKind = "arrow"
	PPTXShapeTextBox    PPTXShapeKind = "textBox"
	PPTXShapePicture    PPTXShapeKind = "picture"
)

// PPTXShape represents a shape on a slide.
type PPTXShape struct {
	slide       *PPTXSlide
	shapeKind   PPTXShapeKind
	position    PPTXPosition
	size        PPTXSize
	rotation    float64
	style       PPTXShapeStyle
	textFrame   *PPTXTextFrame
	shapeId     int
	placeholder bool
	phType      string // placeholder type
}

// PPTXChartKind defines chart types.
type PPTXChartKind string

const (
	PPTXChartBar        PPTXChartKind = "bar"
	PPTXChartBarStacked PPTXChartKind = "barStacked"
	PPTXChartColumn     PPTXChartKind = "column"
	PPTXChartLine       PPTXChartKind = "line"
	PPTXChartPie        PPTXChartKind = "pie"
	PPTXChartArea       PPTXChartKind = "area"
	PPTXChartScatter    PPTXChartKind = "scatter"
	PPTXChartRadar      PPTXChartKind = "radar"
	PPTXChartCombo      PPTXChartKind = "combo"
)

// PPTXTable represents a table on a slide.
type PPTXTable struct {
	slide    *PPTXSlide
	rows     int
	cols     int
	cells    [][]PPTXTableCell
	position PPTXPosition
	size     PPTXSize
	style    PPTXTableStyle
}

// PPTXTableCell represents a table cell.
type PPTXTableCell struct {
	table    *PPTXTable
	row      int
	col      int
	text     string
	spanRows int
	spanCols int
	style    PPTXCellStyle
}

// PPTXChart represents a chart on a slide.
type PPTXChart struct {
	slide     *PPTXSlide
	chartKind PPTXChartKind
	title     string
	data      PPTXChartData
	series    []*PPTXChartSeries
	position  PPTXPosition
	size      PPTXSize
	style     PPTXChartStyle
}

// PPTXChartData represents chart data.
type PPTXChartData struct {
	categories []string
	series     []PPTXChartSeriesData
}

// PPTXChartSeriesData represents a data series input.
type PPTXChartSeriesData struct {
	Name   string
	Values []float64
}

// PPTXChartSeries represents a data series in a chart.
type PPTXChartSeries struct {
	name   string
	values []float64
	color  PPTXColor
}

// PPTXImage represents an image on a slide.
type PPTXImage struct {
	slide      *PPTXSlide
	data       []byte
	format     string
	position   PPTXPosition
	size       PPTXSize
	rotation   float64
	relationID string
	filename   string
}

// PPTXVideo represents a video on a slide.
type PPTXVideo struct {
	slide      *PPTXSlide
	data       []byte
	format     string
	position   PPTXPosition
	size       PPTXSize
	poster     *PPTXImage
	relationID string
}

// PPTXAudio represents an audio on a slide.
type PPTXAudio struct {
	slide      *PPTXSlide
	data       []byte
	format     string
	position   PPTXPosition
	size       PPTXSize
	icon       *PPTXImage
	relationID string
}

// PPTXSlideLayout represents a slide layout template.
type PPTXSlideLayout struct {
	document     *PPTXDocument
	name         string
	index        int
	shapes       []*PPTXShape
	textFrames   []*PPTXTextFrame
	placeholders map[string]*PPTXShape
	xmlData      []byte
}

// PPTXSlideMaster represents a slide master.
type PPTXSlideMaster struct {
	document *PPTXDocument
	name     string
	layouts  []*PPTXSlideLayout
	theme    *PPTXTheme
	xmlData  []byte
}

// PPTXTheme represents a presentation theme.
type PPTXTheme struct {
	document     *PPTXDocument
	name         string
	colors       PPTXThemeColors
	fonts        PPTXThemeFonts
	formatScheme PPTXFormatScheme
	xmlData      []byte
}

// PPTXThemeColors defines theme color scheme.
type PPTXThemeColors struct {
	lt1, lt2 PPTXColor
	dk1, dk2 PPTXColor
	accent1  PPTXColor
	accent2  PPTXColor
	accent3  PPTXColor
	accent4  PPTXColor
	accent5  PPTXColor
	accent6  PPTXColor
	hlink    PPTXColor
	folHlink PPTXColor
}

// PPTXThemeFonts defines theme fonts.
type PPTXThemeFonts struct {
	majorFont string
	minorFont string
}

// PPTXFormatScheme defines format scheme.
type PPTXFormatScheme struct {
	name string
}

// PPTXPosition represents element position in EMUs (English Metric Units).
type PPTXPosition struct {
	X int64 // left position in EMUs
	Y int64 // top position in EMUs
}

// PPTXSize represents element dimensions in EMUs.
type PPTXSize struct {
	Width  int64
	Height int64
}

// PPTXColor represents RGB color.
type PPTXColor struct {
	R, G, B uint8
}

// PPTXBulletStyle represents bullet/numbering style.
type PPTXBulletStyle struct {
	Type    string // "bullet", "numbered", "none"
	Char    string // bullet character
	StartAt int    // starting number for numbered lists
	Font    PPTXFont
	Color   PPTXColor
}

// PPTXFont represents font properties.
type PPTXFont struct {
	Name   string
	Size   int  // in points
	Bold   bool
	Italic bool
}

// PPTXShapeStyle represents shape styling.
type PPTXShapeStyle struct {
	Fill             *PPTXColor
	FillTransparency float64
	BorderColor      *PPTXColor
	BorderWidth      int64
	BorderStyle      string // "solid", "dashed", "none"
	Shadow           *PPTXShadow
}

// PPTXShadow represents shadow effect.
type PPTXShadow struct {
	Type     string
	Color    PPTXColor
	Blur     int64
	Distance int64
	Angle    int
}

// PPTXTableStyle represents table styling.
type PPTXTableStyle struct {
	HeaderRow PPTXCellStyle
	TotalRow  PPTXCellStyle
	FirstCol  PPTXCellStyle
	LastCol   PPTXCellStyle
	BandRow   bool
	BandCol   bool
}

// PPTXCellStyle represents table cell styling.
type PPTXCellStyle struct {
	Fill        *PPTXColor
	BorderColor *PPTXColor
	BorderWidth int64
	Font        PPTXFont
	Alignment   string
	Vertical    string
}

// PPTXSlideBackground represents slide background.
type PPTXSlideBackground struct {
	fill     *PPTXColor
	image    *PPTXImage
	gradient *PPTXGradient
}

// PPTXGradient represents gradient fill.
type PPTXGradient struct {
	angle  float64
	colors []PPTXColor
	stops  []float64
}

// PPTXChartStyle represents chart styling.
type PPTXChartStyle struct {
	showLegend bool
	showLabels bool
}

// ========================================
// Type implementations for PPTXDocument
// ========================================

func (d *PPTXDocument) Type() ObjectType { return PPTXDocumentType }
func (d *PPTXDocument) TypeTag() TypeTag { return TagPPTXDocument }
func (d *PPTXDocument) ToBool() *Bool    { return TRUE }
func (d *PPTXDocument) HashKey() HashKey {
	return HashKey{Type: PPTXDocumentType, Value: uint64(uintptr(unsafe.Pointer(d)))}
}
func (d *PPTXDocument) Inspect() string {
	if d.filePath != "" {
		return fmt.Sprintf("PPTXDocument(path=%s, slides=%d)", d.filePath, len(d.slides))
	}
	return fmt.Sprintf("PPTXDocument(new, slides=%d)", len(d.slides))
}

// Type implementations for PPTXSlide
func (s *PPTXSlide) Type() ObjectType { return PPTXSlideType }
func (s *PPTXSlide) TypeTag() TypeTag { return TagPPTXSlide }
func (s *PPTXSlide) ToBool() *Bool    { return TRUE }
func (s *PPTXSlide) HashKey() HashKey {
	return HashKey{Type: PPTXSlideType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *PPTXSlide) Inspect() string {
	return fmt.Sprintf("PPTXSlide(index=%d)", s.index)
}

// Type implementations for PPTXTextFrame
func (tf *PPTXTextFrame) Type() ObjectType { return PPTXTextFrameType }
func (tf *PPTXTextFrame) TypeTag() TypeTag { return TagPPTXTextFrame }
func (tf *PPTXTextFrame) ToBool() *Bool    { return TRUE }
func (tf *PPTXTextFrame) HashKey() HashKey {
	return HashKey{Type: PPTXTextFrameType, Value: uint64(uintptr(unsafe.Pointer(tf)))}
}
func (tf *PPTXTextFrame) Inspect() string {
	return fmt.Sprintf("PPTXTextFrame(text=%q)", truncateText(tf.text, 30))
}

// Type implementations for PPTXParagraph
func (p *PPTXParagraph) Type() ObjectType { return PPTXParagraphType }
func (p *PPTXParagraph) TypeTag() TypeTag { return TagPPTXParagraph }
func (p *PPTXParagraph) ToBool() *Bool    { return TRUE }
func (p *PPTXParagraph) HashKey() HashKey {
	return HashKey{Type: PPTXParagraphType, Value: uint64(uintptr(unsafe.Pointer(p)))}
}
func (p *PPTXParagraph) Inspect() string {
	return "PPTXParagraph()"
}

// Type implementations for PPTXTextRun
func (r *PPTXTextRun) Type() ObjectType { return PPTXTextRunType }
func (r *PPTXTextRun) TypeTag() TypeTag { return TagPPTXTextRun }
func (r *PPTXTextRun) ToBool() *Bool    { return TRUE }
func (r *PPTXTextRun) HashKey() HashKey {
	return HashKey{Type: PPTXTextRunType, Value: uint64(uintptr(unsafe.Pointer(r)))}
}
func (r *PPTXTextRun) Inspect() string {
	return fmt.Sprintf("PPTXTextRun(text=%q)", truncateText(r.text, 20))
}

// Type implementations for PPTXShape
func (s *PPTXShape) Type() ObjectType { return PPTXShapeType }
func (s *PPTXShape) TypeTag() TypeTag { return TagPPTXShape }
func (s *PPTXShape) ToBool() *Bool    { return TRUE }
func (s *PPTXShape) HashKey() HashKey {
	return HashKey{Type: PPTXShapeType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *PPTXShape) Inspect() string {
	return fmt.Sprintf("PPTXShape(kind=%s)", s.shapeKind)
}

// Type implementations for PPTXTable
func (t *PPTXTable) Type() ObjectType { return PPTXTableType }
func (t *PPTXTable) TypeTag() TypeTag { return TagPPTXTable }
func (t *PPTXTable) ToBool() *Bool    { return TRUE }
func (t *PPTXTable) HashKey() HashKey {
	return HashKey{Type: PPTXTableType, Value: uint64(uintptr(unsafe.Pointer(t)))}
}
func (t *PPTXTable) Inspect() string {
	return fmt.Sprintf("PPTXTable(rows=%d, cols=%d)", t.rows, t.cols)
}

// Type implementations for PPTXTableCell
func (c *PPTXTableCell) Type() ObjectType { return PPTXTableCellType }
func (c *PPTXTableCell) TypeTag() TypeTag { return TagPPTXTableCell }
func (c *PPTXTableCell) ToBool() *Bool    { return TRUE }
func (c *PPTXTableCell) HashKey() HashKey {
	return HashKey{Type: PPTXTableCellType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}
func (c *PPTXTableCell) Inspect() string {
	return fmt.Sprintf("PPTXTableCell(row=%d, col=%d)", c.row, c.col)
}

// Type implementations for PPTXChart
func (c *PPTXChart) Type() ObjectType { return PPTXChartType }
func (c *PPTXChart) TypeTag() TypeTag { return TagPPTXChart }
func (c *PPTXChart) ToBool() *Bool    { return TRUE }
func (c *PPTXChart) HashKey() HashKey {
	return HashKey{Type: PPTXChartType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}
func (c *PPTXChart) Inspect() string {
	return fmt.Sprintf("PPTXChart(kind=%s, title=%q)", c.chartKind, c.title)
}

// Type implementations for PPTXChartSeries
func (s *PPTXChartSeries) Type() ObjectType { return PPTXChartSeriesType }
func (s *PPTXChartSeries) TypeTag() TypeTag { return TagPPTXChartSeries }
func (s *PPTXChartSeries) ToBool() *Bool    { return TRUE }
func (s *PPTXChartSeries) HashKey() HashKey {
	return HashKey{Type: PPTXChartSeriesType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}
func (s *PPTXChartSeries) Inspect() string {
	return fmt.Sprintf("PPTXChartSeries(name=%q)", s.name)
}

// Type implementations for PPTXImage
func (i *PPTXImage) Type() ObjectType { return PPTXImageType }
func (i *PPTXImage) TypeTag() TypeTag { return TagPPTXImage }
func (i *PPTXImage) ToBool() *Bool    { return TRUE }
func (i *PPTXImage) HashKey() HashKey {
	return HashKey{Type: PPTXImageType, Value: uint64(uintptr(unsafe.Pointer(i)))}
}
func (i *PPTXImage) Inspect() string {
	return fmt.Sprintf("PPTXImage(format=%s, size=%dx%d)", i.format, i.size.Width, i.size.Height)
}

// Type implementations for PPTXVideo
func (v *PPTXVideo) Type() ObjectType { return PPTXVideoType }
func (v *PPTXVideo) TypeTag() TypeTag { return TagPPTXVideo }
func (v *PPTXVideo) ToBool() *Bool    { return TRUE }
func (v *PPTXVideo) HashKey() HashKey {
	return HashKey{Type: PPTXVideoType, Value: uint64(uintptr(unsafe.Pointer(v)))}
}
func (v *PPTXVideo) Inspect() string {
	return fmt.Sprintf("PPTXVideo(format=%s)", v.format)
}

// Type implementations for PPTXAudio
func (a *PPTXAudio) Type() ObjectType { return PPTXAudioType }
func (a *PPTXAudio) TypeTag() TypeTag { return TagPPTXAudio }
func (a *PPTXAudio) ToBool() *Bool    { return TRUE }
func (a *PPTXAudio) HashKey() HashKey {
	return HashKey{Type: PPTXAudioType, Value: uint64(uintptr(unsafe.Pointer(a)))}
}
func (a *PPTXAudio) Inspect() string {
	return fmt.Sprintf("PPTXAudio(format=%s)", a.format)
}

// Type implementations for PPTXSlideLayout
func (l *PPTXSlideLayout) Type() ObjectType { return PPTXSlideLayoutType }
func (l *PPTXSlideLayout) TypeTag() TypeTag { return TagPPTXSlideLayout }
func (l *PPTXSlideLayout) ToBool() *Bool    { return TRUE }
func (l *PPTXSlideLayout) HashKey() HashKey {
	return HashKey{Type: PPTXSlideLayoutType, Value: uint64(uintptr(unsafe.Pointer(l)))}
}
func (l *PPTXSlideLayout) Inspect() string {
	return fmt.Sprintf("PPTXSlideLayout(name=%q)", l.name)
}

// Type implementations for PPTXSlideMaster
func (m *PPTXSlideMaster) Type() ObjectType { return PPTXSlideMasterType }
func (m *PPTXSlideMaster) TypeTag() TypeTag { return TagPPTXSlideMaster }
func (m *PPTXSlideMaster) ToBool() *Bool    { return TRUE }
func (m *PPTXSlideMaster) HashKey() HashKey {
	return HashKey{Type: PPTXSlideMasterType, Value: uint64(uintptr(unsafe.Pointer(m)))}
}
func (m *PPTXSlideMaster) Inspect() string {
	return fmt.Sprintf("PPTXSlideMaster(name=%q)", m.name)
}

// Type implementations for PPTXTheme
func (t *PPTXTheme) Type() ObjectType { return PPTXThemeType }
func (t *PPTXTheme) TypeTag() TypeTag { return TagPPTXTheme }
func (t *PPTXTheme) ToBool() *Bool    { return TRUE }
func (t *PPTXTheme) HashKey() HashKey {
	return HashKey{Type: PPTXThemeType, Value: uint64(uintptr(unsafe.Pointer(t)))}
}
func (t *PPTXTheme) Inspect() string {
	return fmt.Sprintf("PPTXTheme(name=%q)", t.name)
}

// ========================================
// Constructor functions
// ========================================

// NewPPTX creates a new empty presentation.
func NewPPTX() *PPTXDocument {
	return &PPTXDocument{
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		slides:        []*PPTXSlide{},
		slideLayouts:  []*PPTXSlideLayout{},
		slideMasters:  []*PPTXSlideMaster{},
		themes:        []*PPTXTheme{},
		properties:    &PPTXProperties{},
	}
}

// OpenPPTX opens an existing PPTX file from a file path.
func OpenPPTX(path string) (*PPTXDocument, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open pptx file: %w", err)
	}

	doc := &PPTXDocument{
		filePath:      path,
		zipReader:     reader,
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		slides:        []*PPTXSlide{},
		slideLayouts:  []*PPTXSlideLayout{},
		slideMasters:  []*PPTXSlideMaster{},
		themes:        []*PPTXTheme{},
		properties:    &PPTXProperties{},
	}

	// Parse ZIP contents
	if err := doc.parseZipContents(); err != nil {
		reader.Close()
		return nil, err
	}

	return doc, nil
}

// OpenPPTXFromBytes opens a PPTX presentation from a byte slice.
func OpenPPTXFromBytes(data []byte) (*PPTXDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pptx data: %w", err)
	}

	doc := &PPTXDocument{
		zipData:       data,
		relationships: make(map[string]string),
		contentTypes:  make(map[string]string),
		mediaFiles:    make(map[string][]byte),
		slides:        []*PPTXSlide{},
		slideLayouts:  []*PPTXSlideLayout{},
		slideMasters:  []*PPTXSlideMaster{},
		themes:        []*PPTXTheme{},
		properties:    &PPTXProperties{},
	}

	// Parse ZIP contents
	if err := doc.parseZipContentsFromReader(reader); err != nil {
		return nil, err
	}

	return doc, nil
}

// ========================================
// Parsing methods
// ========================================

// parseZipContents parses the ZIP file contents.
func (d *PPTXDocument) parseZipContents() error {
	return d.parseZipContentsFromReader(&d.zipReader.Reader)
}

// parseZipContentsFromReader parses ZIP contents from a zip.Reader.
func (d *PPTXDocument) parseZipContentsFromReader(reader *zip.Reader) error {
	// First pass: collect all files
	fileMap := make(map[string][]byte)
	for _, file := range reader.File {
		data, err := d.readFileFromZip(file)
		if err != nil {
			continue // Skip files that can't be read
		}
		fileMap[file.Name] = data
	}

	// Parse content types
	if data, ok := fileMap["[Content_Types].xml"]; ok {
		d.parseContentTypes(data)
	}

	// Parse presentation.xml
	if data, ok := fileMap["ppt/presentation.xml"]; ok {
		d.parsePresentation(data)
	}

	// Parse presentation relationships
	if data, ok := fileMap["ppt/_rels/presentation.xml.rels"]; ok {
		d.parseRelationships(data)
	}

	// Parse slides
	d.parseSlides(fileMap)

	// Parse slide layouts
	d.parseSlideLayouts(fileMap)

	// Parse slide masters
	d.parseSlideMasters(fileMap)

	// Parse themes
	d.parseThemes(fileMap)

	// Parse core properties
	if data, ok := fileMap["docProps/core.xml"]; ok {
		d.parseCoreProperties(data)
	}

	// Load media files
	for name, data := range fileMap {
		if strings.HasPrefix(name, "ppt/media/") {
			filename := strings.TrimPrefix(name, "ppt/media/")
			d.mediaFiles[filename] = data
		}
	}

	return nil
}

// readFileFromZip reads a file from a ZIP archive.
func (d *PPTXDocument) readFileFromZip(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// parseContentTypes parses the content types.
func (d *PPTXDocument) parseContentTypes(data []byte) {
	// Parse Override elements
	overrideRe := regexp.MustCompile(`<Override\s+PartName="([^"]+)"\s+ContentType="([^"]+)"`)
	matches := overrideRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		partName := string(m[1])
		contentType := string(m[2])
		d.contentTypes[partName] = contentType
	}

	// Parse Default elements
	defaultRe := regexp.MustCompile(`<Default\s+Extension="([^"]+)"\s+ContentType="([^"]+)"`)
	matches = defaultRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		ext := string(m[1])
		contentType := string(m[2])
		d.contentTypes["."+ext] = contentType
	}
}

// parsePresentation parses the main presentation.xml.
func (d *PPTXDocument) parsePresentation(data []byte) {
	// Extract slide IDs
	sldIdRe := regexp.MustCompile(`<p:sldId\s+[^>]*id="(\d+)"[^>]*r:id="([^"]+)"`)
	matches := sldIdRe.FindAllSubmatch(data, -1)

	for range matches {
		// Create placeholder slides (will be populated later)
		d.slides = append(d.slides, &PPTXSlide{
			document:   d,
			index:      len(d.slides) + 1,
			shapes:     []*PPTXShape{},
			textFrames: []*PPTXTextFrame{},
			images:     []*PPTXImage{},
			tables:     []*PPTXTable{},
			charts:     []*PPTXChart{},
			videos:     []*PPTXVideo{},
			audios:     []*PPTXAudio{},
		})
	}
}

// parseRelationships parses the presentation relationships.
func (d *PPTXDocument) parseRelationships(data []byte) {
	relRe := regexp.MustCompile(`<Relationship\s+Id="([^"]+)"\s+Type="([^"]+)"\s+Target="([^"]+)"`)
	matches := relRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		id := string(m[1])
		target := string(m[3])
		d.relationships[id] = target
	}
}

// parseSlides parses all slides in the presentation.
func (d *PPTXDocument) parseSlides(fileMap map[string][]byte) {
	for i := 1; i <= len(d.slides); i++ {
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", i)
		if data, ok := fileMap[slidePath]; ok {
			slide := d.slides[i-1]
			slide.xmlData = data
			d.parseSlideContent(slide, data)
		}
	}
}

// parseSlideContent parses a single slide's content.
func (d *PPTXDocument) parseSlideContent(slide *PPTXSlide, data []byte) {
	// Extract text content
	textRe := regexp.MustCompile(`<a:t>([^<]*)</a:t>`)
	matches := textRe.FindAllSubmatch(data, -1)
	var allText []string
	for _, m := range matches {
		allText = append(allText, string(m[1]))
	}

	// Create a text frame if there's text
	if len(allText) > 0 {
		tf := &PPTXTextFrame{
			slide:      slide,
			text:       strings.Join(allText, ""),
			paragraphs: []*PPTXParagraph{},
		}

		// Create paragraph with runs
		para := &PPTXParagraph{
			frame: tf,
			runs:  []*PPTXTextRun{},
		}
		for _, txt := range allText {
			para.runs = append(para.runs, &PPTXTextRun{
				paragraph: para,
				text:      txt,
			})
		}
		tf.paragraphs = append(tf.paragraphs, para)
		slide.textFrames = append(slide.textFrames, tf)
	}

	// Extract shapes
	d.parseShapes(slide, data)

	// Extract images
	d.parseSlideImages(slide, data)
}

// parseShapes extracts shapes from slide content.
func (d *PPTXDocument) parseShapes(slide *PPTXSlide, data []byte) {
	// Find all sp elements (shapes)
	spRe := regexp.MustCompile(`<p:sp[^>]*>(.*?)</p:sp>`)
	matches := spRe.FindAllSubmatch(data, -1)

	shapeId := 1
	for range matches {
		shape := &PPTXShape{
			slide:     slide,
			shapeId:   shapeId,
			shapeKind: PPTXShapeTextBox,
		}
		slide.shapes = append(slide.shapes, shape)
		shapeId++
	}
}

// parseSlideImages extracts images from slide content.
func (d *PPTXDocument) parseSlideImages(slide *PPTXSlide, data []byte) {
	// Find image references in the slide
	picRe := regexp.MustCompile(`<p:pic[^>]*>.*?<a:blip[^>]*r:embed="([^"]+)"`)
	matches := picRe.FindAllSubmatch(data, -1)

	for _, m := range matches {
		rId := string(m[1])
		image := &PPTXImage{
			slide:      slide,
			relationID: rId,
		}

		// Try to find the image data
		if target, ok := d.relationships[rId]; ok {
			image.filename = target
			mediaName := strings.TrimPrefix(target, "../media/")
			if imgData, ok := d.mediaFiles[mediaName]; ok {
				image.data = imgData
				image.format = detectImageFormat(imgData)
			}
		}

		slide.images = append(slide.images, image)
	}
}

// parseSlideLayouts parses slide layouts.
func (d *PPTXDocument) parseSlideLayouts(fileMap map[string][]byte) {
	for i := 1; ; i++ {
		layoutPath := fmt.Sprintf("ppt/slideLayouts/slideLayout%d.xml", i)
		if data, ok := fileMap[layoutPath]; ok {
			layout := &PPTXSlideLayout{
				document:     d,
				index:        i,
				xmlData:      data,
				placeholders: make(map[string]*PPTXShape),
			}

			// Extract layout name
			nameRe := regexp.MustCompile(`<p:cSld\s+name="([^"]+)"`)
			if m := nameRe.FindSubmatch(data); len(m) > 1 {
				layout.name = string(m[1])
			}

			d.slideLayouts = append(d.slideLayouts, layout)
		} else {
			break
		}
	}
}

// parseSlideMasters parses slide masters.
func (d *PPTXDocument) parseSlideMasters(fileMap map[string][]byte) {
	for i := 1; ; i++ {
		masterPath := fmt.Sprintf("ppt/slideMasters/slideMaster%d.xml", i)
		if data, ok := fileMap[masterPath]; ok {
			master := &PPTXSlideMaster{
				document: d,
				xmlData:  data,
			}

			// Extract master name
			nameRe := regexp.MustCompile(`<p:cSld\s+name="([^"]+)"`)
			if m := nameRe.FindSubmatch(data); len(m) > 1 {
				master.name = string(m[1])
			}

			d.slideMasters = append(d.slideMasters, master)
		} else {
			break
		}
	}
}

// parseThemes parses themes.
func (d *PPTXDocument) parseThemes(fileMap map[string][]byte) {
	for i := 1; ; i++ {
		themePath := fmt.Sprintf("ppt/theme/theme%d.xml", i)
		if data, ok := fileMap[themePath]; ok {
			theme := &PPTXTheme{
				document: d,
				xmlData:  data,
			}

			// Extract theme name
			nameRe := regexp.MustCompile(`<a:theme[^>]*name="([^"]+)"`)
			if m := nameRe.FindSubmatch(data); len(m) > 1 {
				theme.name = string(m[1])
			}

			d.themes = append(d.themes, theme)
		} else {
			break
		}
	}
}

// parseCoreProperties parses document core properties.
func (d *PPTXDocument) parseCoreProperties(data []byte) {
	extractProp := func(name string) string {
		re := regexp.MustCompile(`<` + name + `[^>]*>([^<]*)</` + name + `>`)
		if m := re.FindSubmatch(data); len(m) > 1 {
			return string(m[1])
		}
		return ""
	}

	d.properties.Title = extractProp("dc:title")
	d.properties.Subject = extractProp("dc:subject")
	d.properties.Author = extractProp("dc:creator")
	d.properties.Keywords = extractProp("cp:keywords")
	d.properties.Description = extractProp("dc:description")
	d.properties.Created = extractProp("dcterms:created")
	d.properties.Modified = extractProp("dcterms:modified")
	d.properties.LastModBy = extractProp("cp:lastModifiedBy")
}

// ========================================
// Document methods
// ========================================

// Close closes the document and releases resources.
func (d *PPTXDocument) Close() error {
	if d.zipReader != nil {
		return d.zipReader.Close()
	}
	return nil
}

// GetSlideCount returns the number of slides.
func (d *PPTXDocument) GetSlideCount() int {
	return len(d.slides)
}

// GetSlide returns a slide by index (1-based).
func (d *PPTXDocument) GetSlide(index int) *PPTXSlide {
	if index < 1 || index > len(d.slides) {
		return nil
	}
	return d.slides[index-1]
}

// GetSlides returns all slides.
func (d *PPTXDocument) GetSlides() []*PPTXSlide {
	return d.slides
}

// AddSlide adds a new slide to the presentation.
func (d *PPTXDocument) AddSlide() *PPTXSlide {
	slide := &PPTXSlide{
		document:   d,
		index:      len(d.slides) + 1,
		shapes:     []*PPTXShape{},
		textFrames: []*PPTXTextFrame{},
		images:     []*PPTXImage{},
		tables:     []*PPTXTable{},
		charts:     []*PPTXChart{},
		videos:     []*PPTXVideo{},
		audios:     []*PPTXAudio{},
	}
	d.slides = append(d.slides, slide)
	d.modified = true
	return slide
}

// DeleteSlide deletes a slide by index (1-based).
func (d *PPTXDocument) DeleteSlide(index int) bool {
	if index < 1 || index > len(d.slides) {
		return false
	}

	d.slides = append(d.slides[:index-1], d.slides[index:]...)

	// Re-index remaining slides
	for i, slide := range d.slides {
		slide.index = i + 1
	}

	d.modified = true
	return true
}

// MoveSlide moves a slide from one position to another.
func (d *PPTXDocument) MoveSlide(from, to int) bool {
	if from < 1 || from > len(d.slides) || to < 1 || to > len(d.slides) {
		return false
	}
	if from == to {
		return true
	}

	slide := d.slides[from-1]
	d.slides = append(d.slides[:from-1], d.slides[from:]...)

	// Insert at new position
	insertPos := to - 1
	if to > from {
		insertPos = to - 1
	}
	d.slides = append(d.slides[:insertPos], append([]*PPTXSlide{slide}, d.slides[insertPos:]...)...)

	// Re-index
	for i, s := range d.slides {
		s.index = i + 1
	}

	d.modified = true
	return true
}

// DuplicateSlide creates a copy of a slide.
func (d *PPTXDocument) DuplicateSlide(index int) *PPTXSlide {
	if index < 1 || index > len(d.slides) {
		return nil
	}

	src := d.slides[index-1]
	newSlide := &PPTXSlide{
		document:   d,
		index:      len(d.slides) + 1,
		shapes:     append([]*PPTXShape{}, src.shapes...),
		textFrames: append([]*PPTXTextFrame{}, src.textFrames...),
		images:     append([]*PPTXImage{}, src.images...),
		tables:     append([]*PPTXTable{}, src.tables...),
		charts:     append([]*PPTXChart{}, src.charts...),
		videos:     append([]*PPTXVideo{}, src.videos...),
		audios:     append([]*PPTXAudio{}, src.audios...),
		notes:      src.notes,
		xmlData:    src.xmlData,
	}

	d.slides = append(d.slides, newSlide)
	d.modified = true
	return newSlide
}

// GetProperties returns the document properties.
func (d *PPTXDocument) GetProperties() *PPTXProperties {
	return d.properties
}

// SetProperties sets the document properties.
func (d *PPTXDocument) SetProperties(props *PPTXProperties) {
	d.properties = props
	d.modified = true
}

// Save saves the document to a file.
func (d *PPTXDocument) Save(path string) error {
	data, err := d.ToBytes()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ToBytes converts the document to a byte slice.
func (d *PPTXDocument) ToBytes() ([]byte, error) {
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

	// Write ppt/presentation.xml
	if err := d.writePresentation(writer); err != nil {
		return nil, err
	}

	// Write ppt/_rels/presentation.xml.rels
	if err := d.writePresentationRels(writer); err != nil {
		return nil, err
	}

	// Write slides
	for i, slide := range d.slides {
		if err := d.writeSlide(writer, i+1, slide); err != nil {
			return nil, err
		}
	}

	// Write slide layouts
	for i, layout := range d.slideLayouts {
		if err := d.writeSlideLayout(writer, i+1, layout); err != nil {
			return nil, err
		}
	}

	// Write slide masters
	for i, master := range d.slideMasters {
		if err := d.writeSlideMaster(writer, i+1, master); err != nil {
			return nil, err
		}
	}

	// Write themes
	for i, theme := range d.themes {
		if err := d.writeTheme(writer, i+1, theme); err != nil {
			return nil, err
		}
	}

	// Write media files
	for name, data := range d.mediaFiles {
		w, err := writer.Create("ppt/media/" + name)
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

	if err := writer.Close(); err != nil {
		return nil, err
	}

	d.modified = false
	return buf.Bytes(), nil
}

// writeContentTypes writes [Content_Types].xml
func (d *PPTXDocument) writeContentTypes(w *zip.Writer) error {
	f, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
`)

	for i := range d.slides {
		buf.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
`, i+1))
	}

	for i := range d.slideLayouts {
		buf.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slideLayouts/slideLayout%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
`, i+1))
	}

	for i := range d.slideMasters {
		buf.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slideMasters/slideMaster%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
`, i+1))
	}

	for i := range d.themes {
		buf.WriteString(fmt.Sprintf(`<Override PartName="/ppt/theme/theme%d.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
`, i+1))
	}

	buf.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
</Types>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// writeRels writes _rels/.rels
func (d *PPTXDocument) writeRels(w *zip.Writer) error {
	f, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`

	_, err = f.Write([]byte(content))
	return err
}

// writePresentation writes ppt/presentation.xml
func (d *PPTXDocument) writePresentation(w *zip.Writer) error {
	f, err := w.Create("ppt/presentation.xml")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:sldIdLst>
`)

	for i := range d.slides {
		rId := fmt.Sprintf("rId%d", i+3) // rId1 and rId2 are used for other relationships
		buf.WriteString(fmt.Sprintf(`<p:sldId id="%d" r:id="%s"/>
`, 256+i, rId))
	}

	buf.WriteString(`</p:sldIdLst>
<p:sldSz cx="9144000" cy="6858000"/>
<p:notesSz cx="6858000" cy="9144000"/>
</p:presentation>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// writePresentationRels writes ppt/_rels/presentation.xml.rels
func (d *PPTXDocument) writePresentationRels(w *zip.Writer) error {
	f, err := w.Create("ppt/_rels/presentation.xml.rels")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
`)

	rid := 1

	// Add slide relationships
	for i := range d.slides {
		buf.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>
`, rid, i+1))
		rid++
	}

	// Add slide layout relationships
	for i := range d.slideLayouts {
		buf.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="slideLayouts/slideLayout%d.xml"/>
`, rid, i+1))
		rid++
	}

	// Add slide master relationships
	for i := range d.slideMasters {
		buf.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster%d.xml"/>
`, rid, i+1))
		rid++
	}

	// Add theme relationships
	for i := range d.themes {
		buf.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme%d.xml"/>
`, rid, i+1))
		rid++
	}

	buf.WriteString(`</Relationships>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// writeSlide writes a slide XML file
func (d *PPTXDocument) writeSlide(w *zip.Writer, index int, slide *PPTXSlide) error {
	f, err := w.Create(fmt.Sprintf("ppt/slides/slide%d.xml", index))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:nvGrpSpPr>
<p:cNvPr id="1" name=""/>
<p:cNvGrpSpPr/>
</p:nvGrpSpPr>
<p:grpSpPr/>
`)

	// Write shapes
	shapeId := 2
	for _, shape := range slide.shapes {
		d.writeShapeXML(&buf, shape, shapeId)
		shapeId++
	}

	// Write text frames
	for _, tf := range slide.textFrames {
		d.writeTextFrameXML(&buf, tf, shapeId)
		shapeId++
	}

	// Write images
	for _, img := range slide.images {
		d.writeImageXML(&buf, img, shapeId)
		shapeId++
	}

	// Write tables
	for _, table := range slide.tables {
		d.writeTableXML(&buf, table, shapeId)
		shapeId++
	}

	buf.WriteString(`</p:spTree>
</p:cSld>
<p:clrMapOvr>
<a:masterClrMapping/>
</p:clrMapOvr>
</p:sld>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// writeShapeXML writes shape XML
func (d *PPTXDocument) writeShapeXML(buf *bytes.Buffer, shape *PPTXShape, shapeId int) {
	buf.WriteString(fmt.Sprintf(`<p:sp>
<p:nvSpPr>
<p:cNvPr id="%d" name="Shape %d"/>
<p:cNvSpPr/>
<p:nvPr/>
</p:nvSpPr>
<p:spPr>
<a:xfrm>
<a:off x="%d" y="%d"/>
<a:ext cx="%d" cy="%d"/>
</a:xfrm>
<a:prstGeom prst="rect">
<a:avLst/>
</a:prstGeom>
`, shapeId, shapeId, shape.position.X, shape.position.Y, shape.size.Width, shape.size.Height))

	// Write fill
	if shape.style.Fill != nil {
		buf.WriteString(fmt.Sprintf(`<a:solidFill>
<a:srgbClr val="%02X%02X%02X"/>
</a:solidFill>
`, shape.style.Fill.R, shape.style.Fill.G, shape.style.Fill.B))
	} else {
		buf.WriteString(`<a:noFill/>`)
	}

	// Write border
	if shape.style.BorderColor != nil {
		buf.WriteString(fmt.Sprintf(`<a:ln>
<a:solidFill>
<a:srgbClr val="%02X%02X%02X"/>
</a:solidFill>
</a:ln>
`, shape.style.BorderColor.R, shape.style.BorderColor.G, shape.style.BorderColor.B))
	}

	buf.WriteString(`</p:spPr>
</p:sp>`)
}

// writeTextFrameXML writes text frame XML
func (d *PPTXDocument) writeTextFrameXML(buf *bytes.Buffer, tf *PPTXTextFrame, shapeId int) {
	buf.WriteString(fmt.Sprintf(`<p:sp>
<p:nvSpPr>
<p:cNvPr id="%d" name="TextBox %d"/>
<p:cNvSpPr txBox="1"/>
<p:nvPr/>
</p:nvSpPr>
<p:spPr>
<a:xfrm>
<a:off x="%d" y="%d"/>
<a:ext cx="%d" cy="%d"/>
</a:xfrm>
<a:prstGeom prst="rect">
<a:avLst/>
</a:prstGeom>
</p:spPr>
<p:txBody>
<a:bodyPr wrap="square" rtlCol="0">
<a:spAutoFit/>
</a:bodyPr>
<a:lstStyle/>
<a:p>
`, shapeId, shapeId, tf.position.X, tf.position.Y, tf.size.Width, tf.size.Height))

	// Write runs
	for _, para := range tf.paragraphs {
		for _, run := range para.runs {
			buf.WriteString(`<a:r>`)
			if run.fontName != "" || run.fontSize > 0 || run.bold || run.italic {
				buf.WriteString(`<a:rPr`)
				if run.fontName != "" {
					buf.WriteString(fmt.Sprintf(` typeface="%s"`, run.fontName))
				}
				if run.fontSize > 0 {
					buf.WriteString(fmt.Sprintf(` sz="%d"`, run.fontSize*100))
				}
				if run.bold {
					buf.WriteString(` b="1"`)
				}
				if run.italic {
					buf.WriteString(` i="1"`)
				}
				buf.WriteString(`/>`)
			}
			buf.WriteString(fmt.Sprintf(`<a:t>%s</a:t>`, escapePPTXXML(run.text)))
			buf.WriteString(`</a:r>`)
		}
	}

	buf.WriteString(`</a:p>
</p:txBody>
</p:sp>`)
}

// writeImageXML writes image XML
func (d *PPTXDocument) writeImageXML(buf *bytes.Buffer, img *PPTXImage, shapeId int) {
	// Generate relationship ID if not set
	rId := img.relationID
	if rId == "" {
		rId = fmt.Sprintf("rId%d", shapeId)
	}

	buf.WriteString(fmt.Sprintf(`<p:pic>
<p:nvPicPr>
<p:cNvPr id="%d" name="Image %d"/>
<p:cNvPicPr>
<a:picLocks noChangeAspect="1"/>
</p:cNvPicPr>
<p:nvPr/>
</p:nvPicPr>
<p:blipFill>
<a:blip r:embed="%s"/>
<a:stretch>
<a:fillRect/>
</a:stretch>
</p:blipFill>
<p:spPr>
<a:xfrm>
<a:off x="%d" y="%d"/>
<a:ext cx="%d" cy="%d"/>
</a:xfrm>
<a:prstGeom prst="rect">
<a:avLst/>
</a:prstGeom>
</p:spPr>
</p:pic>`, shapeId, shapeId, rId, img.position.X, img.position.Y, img.size.Width, img.size.Height))
}

// writeTableXML writes table XML
func (d *PPTXDocument) writeTableXML(buf *bytes.Buffer, table *PPTXTable, shapeId int) {
	buf.WriteString(fmt.Sprintf(`<p:graphicFrame>
<p:nvGraphicFramePr>
<p:cNvPr id="%d" name="Table %d"/>
<p:cNvGraphicFramePr/>
</p:nvGraphicFramePr>
<p:xfrm>
<a:off x="%d" y="%d"/>
<a:ext cx="%d" cy="%d"/>
</p:xfrm>
<a:graphic>
<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/table">
<a:tbl>
<a:tblPr/>
<a:tblGrid>
`, shapeId, shapeId, table.position.X, table.position.Y, table.size.Width, table.size.Height))

	// Write column widths
	colWidth := table.size.Width / int64(table.cols)
	for i := 0; i < table.cols; i++ {
		buf.WriteString(fmt.Sprintf(`<a:gridCol w="%d"/>`, colWidth))
	}

	buf.WriteString(`</a:tblGrid>`)

	// Write rows
	rowHeight := table.size.Height / int64(table.rows)
	for i := 0; i < table.rows; i++ {
		buf.WriteString(fmt.Sprintf(`<a:tr h="%d">`, rowHeight))
		for j := 0; j < table.cols; j++ {
			cell := table.cells[i][j]
			buf.WriteString(`<a:tc>`)
			buf.WriteString(fmt.Sprintf(`<a:txBody><a:p><a:r><a:t>%s</a:t></a:r></a:p></a:txBody>`, escapePPTXXML(cell.text)))
			buf.WriteString(`</a:tc>`)
		}
		buf.WriteString(`</a:tr>`)
	}

	buf.WriteString(`</a:tbl>
</a:graphicData>
</a:graphic>
</p:graphicFrame>`)
}

// writeSlideLayout writes a slide layout XML file
func (d *PPTXDocument) writeSlideLayout(w *zip.Writer, index int, layout *PPTXSlideLayout) error {
	f, err := w.Create(fmt.Sprintf("ppt/slideLayouts/slideLayout%d.xml", index))
	if err != nil {
		return err
	}

	// Use existing XML data if available
	if layout.xmlData != nil {
		_, err = f.Write(layout.xmlData)
		return err
	}

	// Create minimal layout
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld name="Blank">
<p:spTree>
<p:nvGrpSpPr>
<p:cNvPr id="1" name=""/>
<p:cNvGrpSpPr/>
</p:nvGrpSpPr>
<p:grpSpPr/>
</p:spTree>
</p:cSld>
<p:clrMapOvr>
<a:masterClrMapping/>
</p:clrMapOvr>
</p:sldLayout>`

	_, err = f.Write([]byte(content))
	return err
}

// writeSlideMaster writes a slide master XML file
func (d *PPTXDocument) writeSlideMaster(w *zip.Writer, index int, master *PPTXSlideMaster) error {
	f, err := w.Create(fmt.Sprintf("ppt/slideMasters/slideMaster%d.xml", index))
	if err != nil {
		return err
	}

	// Use existing XML data if available
	if master.xmlData != nil {
		_, err = f.Write(master.xmlData)
		return err
	}

	// Create minimal master
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:nvGrpSpPr>
<p:cNvPr id="1" name=""/>
<p:cNvGrpSpPr/>
</p:nvGrpSpPr>
<p:grpSpPr/>
</p:spTree>
</p:cSld>
<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
</p:sldMaster>`

	_, err = f.Write([]byte(content))
	return err
}

// writeTheme writes a theme XML file
func (d *PPTXDocument) writeTheme(w *zip.Writer, index int, theme *PPTXTheme) error {
	f, err := w.Create(fmt.Sprintf("ppt/theme/theme%d.xml", index))
	if err != nil {
		return err
	}

	// Use existing XML data if available
	if theme.xmlData != nil {
		_, err = f.Write(theme.xmlData)
		return err
	}

	// Create minimal theme
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="%s">
<a:themeElements>
<a:clrScheme name="Office">
<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
<a:dk2><a:srgbClr val="1F497D"/></a:dk2>
<a:lt2><a:srgbClr val="EEECE1"/></a:lt2>
<a:accent1><a:srgbClr val="4F81BD"/></a:accent1>
<a:accent2><a:srgbClr val="C0504D"/></a:accent2>
<a:accent3><a:srgbClr val="9BBB59"/></a:accent3>
<a:accent4><a:srgbClr val="8064A2"/></a:accent4>
<a:accent5><a:srgbClr val="4BACC6"/></a:accent5>
<a:accent6><a:srgbClr val="F79646"/></a:accent6>
<a:hlink><a:srgbClr val="0000FF"/></a:hlink>
<a:folHlink><a:srgbClr val="800080"/></a:folHlink>
</a:clrScheme>
<a:fontScheme name="Office">
<a:majorFont>
<a:latin typeface="Calibri"/>
<a:ea typeface=""/>
<a:cs typeface=""/>
</a:majorFont>
<a:minorFont>
<a:latin typeface="Calibri"/>
<a:ea typeface=""/>
<a:cs typeface=""/>
</a:minorFont>
</a:fontScheme>
<a:fmtScheme name="Office">
<a:fillStyleLst>
<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
<a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"/></a:gs></a:gsLst></a:gradFill>
</a:fillStyleLst>
<a:lnStyleLst>
<a:ln w="9525" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
</a:lnStyleLst>
<a:effectStyleLst>
<a:effectStyle><a:effectLst/></a:effectStyle>
</a:effectStyleLst>
<a:bgFillStyleLst>
<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
</a:bgFillStyleLst>
</a:fmtScheme>
</a:themeElements>
</a:theme>`, theme.name)

	_, err = f.Write([]byte(content))
	return err
}

// writeCoreProps writes docProps/core.xml
func (d *PPTXDocument) writeCoreProps(w *zip.Writer) error {
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
		escapePPTXXML(d.properties.Title),
		escapePPTXXML(d.properties.Subject),
		escapePPTXXML(d.properties.Author),
		escapePPTXXML(d.properties.Keywords),
		escapePPTXXML(d.properties.Description),
		d.properties.Created,
		d.properties.Modified,
		escapePPTXXML(d.properties.LastModBy))

	_, err = f.Write([]byte(content))
	return err
}

// escapePPTXXML escapes special characters for XML
func escapePPTXXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// ========================================
// Slide methods
// ========================================

// GetIndex returns the slide index (1-based).
func (s *PPTXSlide) GetIndex() int {
	return s.index
}

// GetTexts returns all text frames on the slide.
func (s *PPTXSlide) GetTexts() []*PPTXTextFrame {
	return s.textFrames
}

// GetShapes returns all shapes on the slide.
func (s *PPTXSlide) GetShapes() []*PPTXShape {
	return s.shapes
}

// GetImages returns all images on the slide.
func (s *PPTXSlide) GetImages() []*PPTXImage {
	return s.images
}

// GetTables returns all tables on the slide.
func (s *PPTXSlide) GetTables() []*PPTXTable {
	return s.tables
}

// GetCharts returns all charts on the slide.
func (s *PPTXSlide) GetCharts() []*PPTXChart {
	return s.charts
}

// GetAllText returns all text content combined.
func (s *PPTXSlide) GetAllText() string {
	var texts []string
	for _, tf := range s.textFrames {
		texts = append(texts, tf.text)
	}
	return strings.Join(texts, "\n")
}

// AddText adds a text frame to the slide.
func (s *PPTXSlide) AddText(text string, options map[string]interface{}) *PPTXTextFrame {
	tf := &PPTXTextFrame{
		slide:      s,
		text:       text,
		paragraphs: []*PPTXParagraph{},
	}

	// Parse options
	if x, ok := options["x"]; ok {
		tf.position.X = toInt64(x)
	}
	if y, ok := options["y"]; ok {
		tf.position.Y = toInt64(y)
	}
	if w, ok := options["width"]; ok {
		tf.size.Width = toInt64(w)
	}
	if h, ok := options["height"]; ok {
		tf.size.Height = toInt64(h)
	}

	// Create default paragraph
	para := &PPTXParagraph{
		frame: tf,
		runs:  []*PPTXTextRun{},
	}

	// Parse text style options
	run := &PPTXTextRun{
		paragraph: para,
		text:      text,
	}
	if fn, ok := options["fontName"]; ok {
		run.fontName = toString(fn)
	}
	if fs, ok := options["fontSize"]; ok {
		run.fontSize = toInt(fs)
	}
	if c, ok := options["color"]; ok {
		run.color = parsePPTXColor(c)
	}
	if b, ok := options["bold"]; ok {
		run.bold = toBool(b)
	}
	if i, ok := options["italic"]; ok {
		run.italic = toBool(i)
	}

	para.runs = append(para.runs, run)
	tf.paragraphs = append(tf.paragraphs, para)
	s.textFrames = append(s.textFrames, tf)
	s.document.modified = true

	return tf
}

// AddShape adds a shape to the slide.
func (s *PPTXSlide) AddShape(shapeKind PPTXShapeKind, options map[string]interface{}) *PPTXShape {
	shape := &PPTXShape{
		slide:     s,
		shapeKind: shapeKind,
		shapeId:   len(s.shapes) + 1,
	}

	// Parse options
	if x, ok := options["x"]; ok {
		shape.position.X = toInt64(x)
	}
	if y, ok := options["y"]; ok {
		shape.position.Y = toInt64(y)
	}
	if w, ok := options["width"]; ok {
		shape.size.Width = toInt64(w)
	}
	if h, ok := options["height"]; ok {
		shape.size.Height = toInt64(h)
	}
	if fill, ok := options["fill"]; ok {
		shape.style.Fill = parsePPTXColorPtr(fill)
	}
	if border, ok := options["borderColor"]; ok {
		shape.style.BorderColor = parsePPTXColorPtr(border)
	}

	s.shapes = append(s.shapes, shape)
	s.document.modified = true

	return shape
}

// AddImage adds an image to the slide from a file.
func (s *PPTXSlide) AddImage(path string, options map[string]interface{}) *PPTXImage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	return s.AddImageFromBytes(data, detectImageFormat(data), options)
}

// AddImageFromBytes adds an image to the slide from byte data.
func (s *PPTXSlide) AddImageFromBytes(data []byte, format string, options map[string]interface{}) *PPTXImage {
	image := &PPTXImage{
		slide:  s,
		data:   data,
		format: format,
	}

	// Parse options
	if x, ok := options["x"]; ok {
		image.position.X = toInt64(x)
	}
	if y, ok := options["y"]; ok {
		image.position.Y = toInt64(y)
	}
	if w, ok := options["width"]; ok {
		image.size.Width = toInt64(w)
	}
	if h, ok := options["height"]; ok {
		image.size.Height = toInt64(h)
	}

	// Add to media files
	filename := fmt.Sprintf("image%d.%s", len(s.document.mediaFiles)+1, format)
	s.document.mediaFiles[filename] = data
	image.filename = filename

	s.images = append(s.images, image)
	s.document.modified = true

	return image
}

// AddTable adds a table to the slide.
func (s *PPTXSlide) AddTable(rows, cols int, options map[string]interface{}) *PPTXTable {
	table := &PPTXTable{
		slide: s,
		rows:  rows,
		cols:  cols,
		cells: make([][]PPTXTableCell, rows),
	}

	// Initialize cells
	for i := range table.cells {
		table.cells[i] = make([]PPTXTableCell, cols)
		for j := range table.cells[i] {
			table.cells[i][j] = PPTXTableCell{
				table: table,
				row:   i + 1,
				col:   j + 1,
			}
		}
	}

	// Parse options
	if x, ok := options["x"]; ok {
		table.position.X = toInt64(x)
	}
	if y, ok := options["y"]; ok {
		table.position.Y = toInt64(y)
	}
	if w, ok := options["width"]; ok {
		table.size.Width = toInt64(w)
	}
	if h, ok := options["height"]; ok {
		table.size.Height = toInt64(h)
	}

	s.tables = append(s.tables, table)
	s.document.modified = true

	return table
}

// AddChart adds a chart to the slide.
func (s *PPTXSlide) AddChart(chartKind PPTXChartKind, data PPTXChartData, options map[string]interface{}) *PPTXChart {
	chart := &PPTXChart{
		slide:     s,
		chartKind: chartKind,
		data:      data,
		series:    []*PPTXChartSeries{},
	}

	// Parse options
	if title, ok := options["title"]; ok {
		chart.title = toString(title)
	}
	if x, ok := options["x"]; ok {
		chart.position.X = toInt64(x)
	}
	if y, ok := options["y"]; ok {
		chart.position.Y = toInt64(y)
	}
	if w, ok := options["width"]; ok {
		chart.size.Width = toInt64(w)
	}
	if h, ok := options["height"]; ok {
		chart.size.Height = toInt64(h)
	}

	// Create series from data
	for _, sd := range data.series {
		series := &PPTXChartSeries{
			name:   sd.Name,
			values: sd.Values,
		}
		chart.series = append(chart.series, series)
	}

	s.charts = append(s.charts, chart)
	s.document.modified = true

	return chart
}

// GetNotes returns the speaker notes.
func (s *PPTXSlide) GetNotes() string {
	return s.notes
}

// SetNotes sets the speaker notes.
func (s *PPTXSlide) SetNotes(text string) {
	s.notes = text
	s.document.modified = true
}

// ========================================
// TextFrame methods
// ========================================

// GetText returns the text content.
func (tf *PPTXTextFrame) GetText() string {
	return tf.text
}

// SetText sets the text content.
func (tf *PPTXTextFrame) SetText(text string) {
	tf.text = text
	tf.slide.document.modified = true
}

// GetParagraphs returns all paragraphs.
func (tf *PPTXTextFrame) GetParagraphs() []*PPTXParagraph {
	return tf.paragraphs
}

// GetPosition returns the position.
func (tf *PPTXTextFrame) GetPosition() PPTXPosition {
	return tf.position
}

// SetPosition sets the position.
func (tf *PPTXTextFrame) SetPosition(x, y int64) {
	tf.position.X = x
	tf.position.Y = y
	tf.slide.document.modified = true
}

// GetSize returns the size.
func (tf *PPTXTextFrame) GetSize() PPTXSize {
	return tf.size
}

// SetSize sets the size.
func (tf *PPTXTextFrame) SetSize(width, height int64) {
	tf.size.Width = width
	tf.size.Height = height
	tf.slide.document.modified = true
}

// ========================================
// Paragraph methods
// ========================================

// GetText returns the text content.
func (p *PPTXParagraph) GetText() string {
	var texts []string
	for _, run := range p.runs {
		texts = append(texts, run.text)
	}
	return strings.Join(texts, "")
}

// GetRuns returns all text runs.
func (p *PPTXParagraph) GetRuns() []*PPTXTextRun {
	return p.runs
}

// GetAlignment returns the alignment.
func (p *PPTXParagraph) GetAlignment() string {
	return p.alignment
}

// SetAlignment sets the alignment.
func (p *PPTXParagraph) SetAlignment(align string) {
	p.alignment = align
	p.frame.slide.document.modified = true
}

// ========================================
// TextRun methods
// ========================================

// GetText returns the text content.
func (r *PPTXTextRun) GetText() string {
	return r.text
}

// SetText sets the text content.
func (r *PPTXTextRun) SetText(text string) {
	r.text = text
	r.paragraph.frame.slide.document.modified = true
}

// GetFontName returns the font name.
func (r *PPTXTextRun) GetFontName() string {
	return r.fontName
}

// SetFontName sets the font name.
func (r *PPTXTextRun) SetFontName(name string) {
	r.fontName = name
	r.paragraph.frame.slide.document.modified = true
}

// GetFontSize returns the font size.
func (r *PPTXTextRun) GetFontSize() int {
	return r.fontSize
}

// SetFontSize sets the font size.
func (r *PPTXTextRun) SetFontSize(size int) {
	r.fontSize = size
	r.paragraph.frame.slide.document.modified = true
}

// IsBold returns whether the text is bold.
func (r *PPTXTextRun) IsBold() bool {
	return r.bold
}

// SetBold sets the bold flag.
func (r *PPTXTextRun) SetBold(bold bool) {
	r.bold = bold
	r.paragraph.frame.slide.document.modified = true
}

// IsItalic returns whether the text is italic.
func (r *PPTXTextRun) IsItalic() bool {
	return r.italic
}

// SetItalic sets the italic flag.
func (r *PPTXTextRun) SetItalic(italic bool) {
	r.italic = italic
	r.paragraph.frame.slide.document.modified = true
}

// GetColor returns the color as hex string.
func (r *PPTXTextRun) GetColor() string {
	return fmt.Sprintf("%02X%02X%02X", r.color.R, r.color.G, r.color.B)
}

// SetColor sets the color from hex string.
func (r *PPTXTextRun) SetColor(hex string) {
	r.color = parseHexColor(hex)
	r.paragraph.frame.slide.document.modified = true
}

// ========================================
// Shape methods
// ========================================

// GetKind returns the shape kind.
func (s *PPTXShape) GetKind() PPTXShapeKind {
	return s.shapeKind
}

// GetPosition returns the position.
func (s *PPTXShape) GetPosition() PPTXPosition {
	return s.position
}

// SetPosition sets the position.
func (s *PPTXShape) SetPosition(x, y int64) {
	s.position.X = x
	s.position.Y = y
	s.slide.document.modified = true
}

// GetSize returns the size.
func (s *PPTXShape) GetSize() PPTXSize {
	return s.size
}

// SetSize sets the size.
func (s *PPTXShape) SetSize(width, height int64) {
	s.size.Width = width
	s.size.Height = height
	s.slide.document.modified = true
}

// GetFill returns the fill color.
func (s *PPTXShape) GetFill() *PPTXColor {
	return s.style.Fill
}

// SetFill sets the fill color.
func (s *PPTXShape) SetFill(color string) {
	s.style.Fill = parsePPTXColorPtr(color)
	s.slide.document.modified = true
}

// GetTextFrame returns the text frame.
func (s *PPTXShape) GetTextFrame() *PPTXTextFrame {
	return s.textFrame
}

// AddTextFrame adds a text frame to the shape.
func (s *PPTXShape) AddTextFrame(text string) *PPTXTextFrame {
	tf := &PPTXTextFrame{
		slide: s.slide,
		text:  text,
	}
	s.textFrame = tf
	s.slide.document.modified = true
	return tf
}

// ========================================
// Table methods
// ========================================

// GetRowCount returns the number of rows.
func (t *PPTXTable) GetRowCount() int {
	return t.rows
}

// GetColCount returns the number of columns.
func (t *PPTXTable) GetColCount() int {
	return t.cols
}

// GetCell returns a cell by row and column (1-based).
func (t *PPTXTable) GetCell(row, col int) *PPTXTableCell {
	if row < 1 || row > t.rows || col < 1 || col > t.cols {
		return nil
	}
	return &t.cells[row-1][col-1]
}

// GetValue returns a cell value.
func (t *PPTXTable) GetValue(row, col int) string {
	cell := t.GetCell(row, col)
	if cell == nil {
		return ""
	}
	return cell.text
}

// SetValue sets a cell value.
func (t *PPTXTable) SetValue(row, col int, value string) {
	cell := t.GetCell(row, col)
	if cell != nil {
		cell.text = value
		t.slide.document.modified = true
	}
}

// GetPosition returns the position.
func (t *PPTXTable) GetPosition() PPTXPosition {
	return t.position
}

// SetPosition sets the position.
func (t *PPTXTable) SetPosition(x, y int64) {
	t.position.X = x
	t.position.Y = y
	t.slide.document.modified = true
}

// GetSize returns the size.
func (t *PPTXTable) GetSize() PPTXSize {
	return t.size
}

// SetSize sets the size.
func (t *PPTXTable) SetSize(width, height int64) {
	t.size.Width = width
	t.size.Height = height
	t.slide.document.modified = true
}

// ========================================
// Chart methods
// ========================================

// GetKind returns the chart kind.
func (c *PPTXChart) GetKind() PPTXChartKind {
	return c.chartKind
}

// GetTitle returns the chart title.
func (c *PPTXChart) GetTitle() string {
	return c.title
}

// SetTitle sets the chart title.
func (c *PPTXChart) SetTitle(title string) {
	c.title = title
	c.slide.document.modified = true
}

// GetSeriesCount returns the number of series.
func (c *PPTXChart) GetSeriesCount() int {
	return len(c.series)
}

// GetSeries returns all series.
func (c *PPTXChart) GetSeries() []*PPTXChartSeries {
	return c.series
}

// GetPosition returns the position.
func (c *PPTXChart) GetPosition() PPTXPosition {
	return c.position
}

// SetPosition sets the position.
func (c *PPTXChart) SetPosition(x, y int64) {
	c.position.X = x
	c.position.Y = y
	c.slide.document.modified = true
}

// GetSize returns the size.
func (c *PPTXChart) GetSize() PPTXSize {
	return c.size
}

// SetSize sets the size.
func (c *PPTXChart) SetSize(width, height int64) {
	c.size.Width = width
	c.size.Height = height
	c.slide.document.modified = true
}

// ========================================
// ChartSeries methods
// ========================================

// GetName returns the series name.
func (s *PPTXChartSeries) GetName() string {
	return s.name
}

// SetName sets the series name.
func (s *PPTXChartSeries) SetName(name string) {
	s.name = name
}

// GetValues returns the values.
func (s *PPTXChartSeries) GetValues() []float64 {
	return s.values
}

// SetValues sets the values.
func (s *PPTXChartSeries) SetValues(values []float64) {
	s.values = values
}

// ========================================
// Image methods
// ========================================

// GetData returns the image data.
func (i *PPTXImage) GetData() []byte {
	return i.data
}

// GetDataBase64 returns the image data as base64.
func (i *PPTXImage) GetDataBase64() string {
	return base64.StdEncoding.EncodeToString(i.data)
}

// Save saves the image to a file.
func (i *PPTXImage) Save(path string) error {
	return os.WriteFile(path, i.data, 0644)
}

// GetFormat returns the image format.
func (i *PPTXImage) GetFormat() string {
	return i.format
}

// GetPosition returns the position.
func (i *PPTXImage) GetPosition() PPTXPosition {
	return i.position
}

// SetPosition sets the position.
func (i *PPTXImage) SetPosition(x, y int64) {
	i.position.X = x
	i.position.Y = y
	i.slide.document.modified = true
}

// GetSize returns the size.
func (i *PPTXImage) GetSize() PPTXSize {
	return i.size
}

// SetSize sets the size.
func (i *PPTXImage) SetSize(width, height int64) {
	i.size.Width = width
	i.size.Height = height
	i.slide.document.modified = true
}

// ========================================
// Video methods
// ========================================

// GetData returns the video data.
func (v *PPTXVideo) GetData() []byte {
	return v.data
}

// Save saves the video to a file.
func (v *PPTXVideo) Save(path string) error {
	return os.WriteFile(path, v.data, 0644)
}

// GetFormat returns the video format.
func (v *PPTXVideo) GetFormat() string {
	return v.format
}

// GetPosition returns the position.
func (v *PPTXVideo) GetPosition() PPTXPosition {
	return v.position
}

// SetPosition sets the position.
func (v *PPTXVideo) SetPosition(x, y int64) {
	v.position.X = x
	v.position.Y = y
	v.slide.document.modified = true
}

// GetSize returns the size.
func (v *PPTXVideo) GetSize() PPTXSize {
	return v.size
}

// SetSize sets the size.
func (v *PPTXVideo) SetSize(width, height int64) {
	v.size.Width = width
	v.size.Height = height
	v.slide.document.modified = true
}

// ========================================
// Audio methods
// ========================================

// GetData returns the audio data.
func (a *PPTXAudio) GetData() []byte {
	return a.data
}

// Save saves the audio to a file.
func (a *PPTXAudio) Save(path string) error {
	return os.WriteFile(path, a.data, 0644)
}

// GetFormat returns the audio format.
func (a *PPTXAudio) GetFormat() string {
	return a.format
}

// GetPosition returns the position.
func (a *PPTXAudio) GetPosition() PPTXPosition {
	return a.position
}

// SetPosition sets the position.
func (a *PPTXAudio) SetPosition(x, y int64) {
	a.position.X = x
	a.position.Y = y
	a.slide.document.modified = true
}

// ========================================
// Helper functions
// ========================================

// truncateText truncates text to a maximum length.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// toInt64 converts various types to int64.
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case *Int:
		return val.Value
	case *Float:
		return int64(val.Value)
	default:
		return 0
	}
}

// toInt converts various types to int.
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case *Int:
		return int(val.Value)
	case *Float:
		return int(val.Value)
	default:
		return 0
	}
}

// toString converts various types to string.
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case *String:
		return val.Value
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toBool converts various types to bool.
func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case *Bool:
		return val.Value
	default:
		return false
	}
}

// parsePPTXColor parses a color from various types.
func parsePPTXColor(v interface{}) PPTXColor {
	switch val := v.(type) {
	case string:
		return parseHexColor(val)
	case *String:
		return parseHexColor(val.Value)
	default:
		return PPTXColor{R: 0, G: 0, B: 0}
	}
}

// parsePPTXColorPtr parses a color and returns a pointer.
func parsePPTXColorPtr(v interface{}) *PPTXColor {
	c := parsePPTXColor(v)
	return &c
}

// parseHexColor parses a hex color string.
func parseHexColor(hex string) PPTXColor {
	c := PPTXColor{}
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) >= 6 {
		r, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b, _ := strconv.ParseUint(hex[4:6], 16, 8)
		c.R = uint8(r)
		c.G = uint8(g)
		c.B = uint8(b)
	}
	return c
}

// detectImageFormat detects image format from magic bytes.
func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return "bin"
	}

	// Check magic bytes
	switch {
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "png"
	case data[0] == 0xFF && data[1] == 0xD8:
		return "jpg"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "gif"
	case data[0] == 0x42 && data[1] == 0x4D:
		return "bmp"
	default:
		return "bin"
	}
}
