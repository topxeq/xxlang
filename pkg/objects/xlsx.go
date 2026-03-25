// pkg/objects/xlsx.go
// XLSX workbook type for Xxlang - Excel file handling.
package objects

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// XLSX represents an Excel workbook.
type XLSX struct {
	filePath string
	reader   *zip.ReadCloser
	sheets   map[string]*Sheet
	sheetOrder []string
	sharedStrings []string
	modified bool
}

// Sheet represents a worksheet.
type Sheet struct {
	Name     string
	Cells    map[string]*Cell // key: "A1", "B2", etc.
	Merges   []MergeRange
	RowCount int
	ColCount int
	Images   []ImageInfo
}

// Cell represents a cell.
type Cell struct {
	Value    string
	Type     string // "s"=shared string, "n"=number, "b"=bool, "str"=string, "d"=date
	Formula  string
	StyleIdx int
}

// MergeRange represents merged cells.
type MergeRange struct {
	StartCol, StartRow int
	EndCol, EndRow     int
}

// ImageInfo represents an image in a sheet.
type ImageInfo struct {
	Col, Row     int // Top-left position
	ColEnd, RowEnd int // Bottom-right position
	Filename     string // Original filename in xlsx
	Data         []byte // Image data
}

// NewXLSX creates a new empty workbook.
func NewXLSX() *XLSX {
	return &XLSX{
		sheets:       make(map[string]*Sheet),
		sheetOrder:   []string{},
		sharedStrings: []string{},
	}
}

// OpenXLSX opens an existing xlsx file.
func OpenXLSX(path string) (*XLSX, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}

	xlsx := &XLSX{
		filePath:      path,
		reader:        reader,
		sheets:        make(map[string]*Sheet),
		sheetOrder:    []string{},
		sharedStrings: []string{},
	}

	// Parse the xlsx structure
	if err := xlsx.parse(); err != nil {
		reader.Close()
		return nil, err
	}

	return xlsx, nil
}

// Type returns the object type.
func (x *XLSX) Type() ObjectType { return XLSXType }

// TypeTag returns the fast type tag.
func (x *XLSX) TypeTag() TypeTag { return TagXLSX }

// Inspect returns a string representation.
func (x *XLSX) Inspect() string {
	return fmt.Sprintf("XLSX(sheets=%d)", len(x.sheets))
}

// ToBool returns true (XLSX is always truthy).
func (x *XLSX) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the XLSX.
func (x *XLSX) HashKey() HashKey {
	return HashKey{
		Type:  XLSXType,
		Value: uint64(uintptr(unsafe.Pointer(x))),
	}
}

// parse reads the xlsx file structure.
func (x *XLSX) parse() error {
	// Parse shared strings
	if err := x.parseSharedStrings(); err != nil {
		// Shared strings file may not exist
	}

	// Parse workbook to get sheet names
	if err := x.parseWorkbook(); err != nil {
		return err
	}

	// Parse each sheet
	for _, sheetName := range x.sheetOrder {
		if err := x.parseSheet(sheetName); err != nil {
			return err
		}
	}

	// Parse images
	x.parseImages()

	return nil
}

// parseSharedStrings parses xl/sharedStrings.xml
func (x *XLSX) parseSharedStrings() error {
	file, err := x.findFile("xl/sharedStrings.xml")
	if err != nil {
		return nil // File may not exist
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Parse shared strings - simple extraction
	re := regexp.MustCompile(`<si[^>]*>(?:<t[^>]*>([^<]*)</t>|<r><t[^>]*>([^<]*)</t></r>)</si>`)
	matches := re.FindAllSubmatch(data, -1)
	x.sharedStrings = make([]string, len(matches))
	for i, m := range matches {
		if len(m[1]) > 0 {
			x.sharedStrings[i] = string(m[1])
		} else if len(m[2]) > 0 {
			x.sharedStrings[i] = string(m[2])
		}
	}
	return nil
}

// parseWorkbook parses xl/workbook.xml to get sheet names.
func (x *XLSX) parseWorkbook() error {
	file, err := x.findFile("xl/workbook.xml")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Extract sheet names
	re := regexp.MustCompile(`<sheet name="([^"]+)"[^>]*sheetId="(\d+)"[^>]*/>`)
	matches := re.FindAllSubmatch(data, -1)

	type sheetInfo struct {
		name string
		id   int
	}
	sheets := []sheetInfo{}
	for _, m := range matches {
		id, _ := strconv.Atoi(string(m[2]))
		sheets = append(sheets, sheetInfo{name: string(m[1]), id: id})
	}

	// Sort by sheetId
	sort.Slice(sheets, func(i, j int) bool { return sheets[i].id < sheets[j].id })

	x.sheetOrder = []string{}
	for _, s := range sheets {
		x.sheetOrder = append(x.sheetOrder, s.name)
		x.sheets[s.name] = &Sheet{
			Name:   s.name,
			Cells:  make(map[string]*Cell),
			Merges: []MergeRange{},
			Images: []ImageInfo{},
		}
	}
	return nil
}

// parseSheet parses a worksheet.
func (x *XLSX) parseSheet(name string) error {
	// Find sheet index
	idx := 0
	for i, n := range x.sheetOrder {
		if n == name {
			idx = i + 1
			break
		}
	}

	file, err := x.findFile(fmt.Sprintf("xl/worksheets/sheet%d.xml", idx))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	sheet := x.sheets[name]

	// Parse cells
	cellRe := regexp.MustCompile(`<c r="([A-Z]+)(\d+)"(?: t="([^"]*)")?(?: s="(\d*)")?>(?:<f>([^<]*)</f>)?(?:<v>([^<]*)</v>)?</c>`)
	matches := cellRe.FindAllSubmatch(data, -1)
	for _, m := range matches {
		col := string(m[1])
		row, _ := strconv.Atoi(string(m[2]))
		cellType := string(m[3])
		// styleIdx := string(m[4])
		formula := string(m[5])
		value := string(m[6])

		cell := &Cell{
			Value:   value,
			Type:    cellType,
			Formula: formula,
		}

		// Convert shared string index to actual value
		if cellType == "s" && value != "" {
			idx, err := strconv.Atoi(value)
			if err == nil && idx >= 0 && idx < len(x.sharedStrings) {
				cell.Value = x.sharedStrings[idx]
			}
		}

		ref := col + strconv.Itoa(row)
		sheet.Cells[ref] = cell

		// Update dimensions
		if row > sheet.RowCount {
			sheet.RowCount = row
		}
		colNum := colToNum(col)
		if colNum > sheet.ColCount {
			sheet.ColCount = colNum
		}
	}

	// Parse merges
	mergeRe := regexp.MustCompile(`<mergeCell ref="([A-Z]+)(\d+):([A-Z]+)(\d+)"`)
	mergeMatches := mergeRe.FindAllSubmatch(data, -1)
	for _, m := range mergeMatches {
		startCol := colToNum(string(m[1]))
		startRow, _ := strconv.Atoi(string(m[2]))
		endCol := colToNum(string(m[3]))
		endRow, _ := strconv.Atoi(string(m[4]))
		sheet.Merges = append(sheet.Merges, MergeRange{
			StartCol: startCol, StartRow: startRow,
			EndCol: endCol, EndRow: endRow,
		})
	}

	return nil
}

// parseImages extracts image information from the workbook.
func (x *XLSX) parseImages() {
	if x.reader == nil {
		return
	}

	// Find all drawing files
	drawingMap := make(map[int]string) // sheet index -> drawing file

	for _, f := range x.reader.File {
		if strings.HasPrefix(f.Name, "xl/drawings/drawing") &&
			strings.HasSuffix(f.Name, ".xml") {
			// Extract sheet number from filename
			re := regexp.MustCompile(`drawing(\d+)\.xml`)
			m := re.FindStringSubmatch(f.Name)
			if len(m) > 1 {
				idx, _ := strconv.Atoi(m[1])
				drawingMap[idx] = f.Name
			}
		}
	}

	// Parse each drawing file
	for sheetIdx, drawingFile := range drawingMap {
		if sheetIdx < 1 || sheetIdx > len(x.sheetOrder) {
			continue
		}
		sheetName := x.sheetOrder[sheetIdx-1]
		sheet := x.sheets[sheetName]

		file, err := x.findFile(drawingFile)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		// Parse image anchors - twoCellAnchor format
		anchorRe := regexp.MustCompile(`<xdr:twoCellAnchor[^>]*>.*?<xdr:from>\s*<xdr:col>(\d+)</xdr:col>\s*<xdr:row>(\d+)</xdr:row>.*?</xdr:from>\s*<xdr:to>\s*<xdr:col>(\d+)</xdr:col>\s*<xdr:row>(\d+)</xdr:row>.*?</xdr:to>\s*<xdr:pic>.*?<a:blip[^>]*r:embed="rId(\d+)"[^>]*/>`)
		matches := anchorRe.FindAllSubmatch(data, -1)
		for _, m := range matches {
			col, _ := strconv.Atoi(string(m[1]))
			row, _ := strconv.Atoi(string(m[2]))
			colEnd, _ := strconv.Atoi(string(m[3]))
			rowEnd, _ := strconv.Atoi(string(m[4]))
			rId := string(m[5])

			// Find image file from relationships
			imageFile := x.findImageFile(sheetIdx, rId)
			if imageFile != "" {
				imgData, _ := x.readMediaFile(imageFile)
				sheet.Images = append(sheet.Images, ImageInfo{
					Col: col, Row: row,
					ColEnd: colEnd, RowEnd: rowEnd,
					Filename: imageFile,
					Data:     imgData,
				})
			}
		}
	}
}

// findImageFile finds the image file for a given rId.
func (x *XLSX) findImageFile(sheetIdx int, rId string) string {
	if x.reader == nil {
		return ""
	}

	// Look for drawing relationships file
	relsFile := fmt.Sprintf("xl/drawings/_rels/drawing%d.xml.rels", sheetIdx)
	file, err := x.findFile(relsFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	// Find the target for the given rId
	re := regexp.MustCompile(`<Relationship[^>]*Id="` + regexp.QuoteMeta(rId) + `"[^>]*Target="([^"]+)"`)
	m := re.FindSubmatch(data)
	if len(m) > 1 {
		target := string(m[1])
		// Target is relative to drawings folder
		if strings.HasPrefix(target, "../media/") {
			return "xl/" + target[3:]
		}
		return "xl/drawings/" + target
	}
	return ""
}

// readMediaFile reads a media file from the xlsx.
func (x *XLSX) readMediaFile(path string) ([]byte, error) {
	file, err := x.findFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// findFile finds a file in the zip.
func (x *XLSX) findFile(name string) (io.ReadCloser, error) {
	if x.reader == nil {
		return nil, fmt.Errorf("file not open: %s", name)
	}
	for _, f := range x.reader.File {
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

// Close closes the workbook.
func (x *XLSX) Close() {
	if x.reader != nil {
		x.reader.Close()
		x.reader = nil
	}
}

// GetSheetList returns the list of sheet names.
func (x *XLSX) GetSheetList() []string {
	return x.sheetOrder
}

// GetSheet returns a sheet by name.
func (x *XLSX) GetSheet(name string) *Sheet {
	return x.sheets[name]
}

// NewSheet creates a new sheet.
func (x *XLSX) NewSheet(name string) bool {
	if _, exists := x.sheets[name]; exists {
		return false
	}
	x.sheets[name] = &Sheet{
		Name:     name,
		Cells:    make(map[string]*Cell),
		Merges:   []MergeRange{},
		Images:   []ImageInfo{},
	}
	x.sheetOrder = append(x.sheetOrder, name)
	x.modified = true
	return true
}

// DeleteSheet deletes a sheet.
func (x *XLSX) DeleteSheet(name string) bool {
	if _, exists := x.sheets[name]; !exists {
		return false
	}
	delete(x.sheets, name)
	for i, n := range x.sheetOrder {
		if n == name {
			x.sheetOrder = append(x.sheetOrder[:i], x.sheetOrder[i+1:]...)
			break
		}
	}
	x.modified = true
	return true
}

// GetCell returns a cell value.
func (x *XLSX) GetCell(sheetName, ref string) Object {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return newError("sheet not found: %s", sheetName)
	}

	cell := sheet.Cells[ref]
	if cell == nil {
		return NULL
	}

	return x.parseCellValue(cell)
}

// GetCellByIndex returns a cell value by row/col indices (1-based).
func (x *XLSX) GetCellByIndex(sheetName string, row, col int) Object {
	ref := numToCol(col) + strconv.Itoa(row)
	return x.GetCell(sheetName, ref)
}

// parseCellValue converts a cell value to an Object.
func (x *XLSX) parseCellValue(cell *Cell) Object {
	if cell == nil {
		return NULL
	}

	switch cell.Type {
	case "b": // Boolean
		return &Bool{Value: cell.Value == "1"}
	case "n": // Number
		if strings.Contains(cell.Value, ".") {
			f, err := strconv.ParseFloat(cell.Value, 64)
			if err != nil {
				return &String{Value: cell.Value}
			}
			return &Float{Value: f}
		}
		i, err := strconv.ParseInt(cell.Value, 10, 64)
		if err != nil {
			return &String{Value: cell.Value}
		}
		return &Int{Value: i}
	case "d": // Date
		return &String{Value: cell.Value}
	default: // String or shared string
		return &String{Value: cell.Value}
	}
}

// SetCell sets a cell value.
func (x *XLSX) SetCell(sheetName, ref string, value Object) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	cell := &Cell{}
	switch v := value.(type) {
	case *String:
		cell.Value = v.Value
		cell.Type = "str"
	case *Int:
		cell.Value = strconv.FormatInt(v.Value, 10)
		cell.Type = "n"
	case *Float:
		cell.Value = strconv.FormatFloat(v.Value, 'f', -1, 64)
		cell.Type = "n"
	case *Bool:
		if v.Value {
			cell.Value = "1"
		} else {
			cell.Value = "0"
		}
		cell.Type = "b"
	default:
		cell.Value = v.Inspect()
		cell.Type = "str"
	}

	sheet.Cells[ref] = cell
	x.modified = true

	// Update dimensions
	col, row := parseRef(ref)
	if row > sheet.RowCount {
		sheet.RowCount = row
	}
	colNum := colToNum(col)
	if colNum > sheet.ColCount {
		sheet.ColCount = colNum
	}

	return nil
}

// SetCellByIndex sets a cell value by row/col indices (1-based).
func (x *XLSX) SetCellByIndex(sheetName string, row, col int, value Object) error {
	ref := numToCol(col) + strconv.Itoa(row)
	return x.SetCell(sheetName, ref, value)
}

// GetRow returns a row as an array.
func (x *XLSX) GetRow(sheetName string, row int) *Array {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return &Array{}
	}

	elements := []Object{}
	for col := 1; col <= sheet.ColCount; col++ {
		ref := numToCol(col) + strconv.Itoa(row)
		cell := sheet.Cells[ref]
		if cell != nil {
			elements = append(elements, x.parseCellValue(cell))
		} else {
			elements = append(elements, NULL)
		}
	}
	return &Array{Elements: elements}
}

// SetRow sets a row from an array.
func (x *XLSX) SetRow(sheetName string, row int, values *Array) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	for col, val := range values.Elements {
		ref := numToCol(col + 1) + strconv.Itoa(row)
		x.SetCell(sheetName, ref, val)
	}
	return nil
}

// GetCol returns a column as an array.
func (x *XLSX) GetCol(sheetName string, col int) *Array {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return &Array{}
	}

	colStr := numToCol(col)
	elements := []Object{}
	for row := 1; row <= sheet.RowCount; row++ {
		ref := colStr + strconv.Itoa(row)
		cell := sheet.Cells[ref]
		if cell != nil {
			elements = append(elements, x.parseCellValue(cell))
		} else {
			elements = append(elements, NULL)
		}
	}
	return &Array{Elements: elements}
}

// GetRange returns a range of cells as a 2D array.
func (x *XLSX) GetRange(sheetName, rangeStr string) *Array {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return &Array{}
	}

	startCol, startRow, endCol, endRow := parseRange(rangeStr)
	rows := []Object{}
	for row := startRow; row <= endRow; row++ {
		rowArr := []Object{}
		for col := startCol; col <= endCol; col++ {
			ref := numToCol(col) + strconv.Itoa(row)
			cell := sheet.Cells[ref]
			if cell != nil {
				rowArr = append(rowArr, x.parseCellValue(cell))
			} else {
				rowArr = append(rowArr, NULL)
			}
		}
		rows = append(rows, &Array{Elements: rowArr})
	}
	return &Array{Elements: rows}
}

// GetRowCount returns the number of rows in a sheet.
func (x *XLSX) GetRowCount(sheetName string) int {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return 0
	}
	return sheet.RowCount
}

// GetColCount returns the number of columns in a sheet.
func (x *XLSX) GetColCount(sheetName string) int {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return 0
	}
	return sheet.ColCount
}

// InsertRow inserts a row at the specified position.
func (x *XLSX) InsertRow(sheetName string, row int) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	// Shift all cells below the insertion point
	newCells := make(map[string]*Cell)
	for ref, cell := range sheet.Cells {
		col, cellRow := parseRef(ref)
		if cellRow >= row {
			// Shift down
			newRef := col + strconv.Itoa(cellRow+1)
			newCells[newRef] = cell
		} else {
			newCells[ref] = cell
		}
	}
	sheet.Cells = newCells
	sheet.RowCount++
	x.modified = true
	return nil
}

// DeleteRow deletes a row.
func (x *XLSX) DeleteRow(sheetName string, row int) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	// Remove cells in the row and shift up
	newCells := make(map[string]*Cell)
	for ref, cell := range sheet.Cells {
		col, cellRow := parseRef(ref)
		if cellRow < row {
			newCells[ref] = cell
		} else if cellRow > row {
			// Shift up
			newRef := col + strconv.Itoa(cellRow-1)
			newCells[newRef] = cell
		}
		// cellRow == row: skip (deleted)
	}
	sheet.Cells = newCells
	if sheet.RowCount > 0 {
		sheet.RowCount--
	}
	x.modified = true
	return nil
}

// InsertCol inserts a column at the specified position.
func (x *XLSX) InsertCol(sheetName string, col int) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	// Shift all cells to the right
	newCells := make(map[string]*Cell)
	for ref, cell := range sheet.Cells {
		colStr, row := parseRef(ref)
		cellCol := colToNum(colStr)
		if cellCol >= col {
			// Shift right
			newRef := numToCol(cellCol+1) + strconv.Itoa(row)
			newCells[newRef] = cell
		} else {
			newCells[ref] = cell
		}
	}
	sheet.Cells = newCells
	sheet.ColCount++
	x.modified = true
	return nil
}

// DeleteCol deletes a column.
func (x *XLSX) DeleteCol(sheetName string, col int) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	// Remove cells in the column and shift left
	newCells := make(map[string]*Cell)
	for ref, cell := range sheet.Cells {
		colStr, row := parseRef(ref)
		cellCol := colToNum(colStr)
		if cellCol < col {
			newCells[ref] = cell
		} else if cellCol > col {
			// Shift left
			newRef := numToCol(cellCol-1) + strconv.Itoa(row)
			newCells[newRef] = cell
		}
		// cellCol == col: skip (deleted)
	}
	sheet.Cells = newCells
	if sheet.ColCount > 0 {
		sheet.ColCount--
	}
	x.modified = true
	return nil
}

// MergeCell merges a range of cells.
func (x *XLSX) MergeCell(sheetName string, startRef, endRef string) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	startCol, startRow := parseRef(startRef)
	endCol, endRow := parseRef(endRef)

	sheet.Merges = append(sheet.Merges, MergeRange{
		StartCol: colToNum(startCol),
		StartRow: startRow,
		EndCol:   colToNum(endCol),
		EndRow:   endRow,
	})
	x.modified = true
	return nil
}

// UnmergeCell removes a merge from cells.
func (x *XLSX) UnmergeCell(sheetName string, ref string) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	col, row := parseRef(ref)
	colNum := colToNum(col)

	// Find and remove the merge containing this cell
	for i, merge := range sheet.Merges {
		if colNum >= merge.StartCol && colNum <= merge.EndCol &&
			row >= merge.StartRow && row <= merge.EndRow {
			sheet.Merges = append(sheet.Merges[:i], sheet.Merges[i+1:]...)
			x.modified = true
			return nil
		}
	}
	return nil
}

// GetMerges returns merge ranges in a sheet.
func (x *XLSX) GetMerges(sheetName string) *Array {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return &Array{}
	}

	elements := []Object{}
	for _, merge := range sheet.Merges {
		startRef := numToCol(merge.StartCol) + strconv.Itoa(merge.StartRow)
		endRef := numToCol(merge.EndCol) + strconv.Itoa(merge.EndRow)
		elements = append(elements, &String{Value: startRef + ":" + endRef})
	}
	return &Array{Elements: elements}
}

// GetImages returns images in a sheet.
func (x *XLSX) GetImages(sheetName string) *Array {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return &Array{}
	}

	elements := []Object{}
	for _, img := range sheet.Images {
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		colKey := &String{Value: "col"}
		m.Pairs[colKey.HashKey()] = MapPair{Key: colKey, Value: &Int{Value: int64(img.Col + 1)}}

		rowKey := &String{Value: "row"}
		m.Pairs[rowKey.HashKey()] = MapPair{Key: rowKey, Value: &Int{Value: int64(img.Row + 1)}}

		colEndKey := &String{Value: "colEnd"}
		m.Pairs[colEndKey.HashKey()] = MapPair{Key: colEndKey, Value: &Int{Value: int64(img.ColEnd + 1)}}

		rowEndKey := &String{Value: "rowEnd"}
		m.Pairs[rowEndKey.HashKey()] = MapPair{Key: rowEndKey, Value: &Int{Value: int64(img.RowEnd + 1)}}

		filenameKey := &String{Value: "filename"}
		m.Pairs[filenameKey.HashKey()] = MapPair{Key: filenameKey, Value: &String{Value: img.Filename}}

		elements = append(elements, m)
	}
	return &Array{Elements: elements}
}

// ExtractImage extracts an image to a file.
func (x *XLSX) ExtractImage(sheetName string, imageIndex int, outputPath string) error {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return fmt.Errorf("sheet not found: %s", sheetName)
	}

	if imageIndex < 0 || imageIndex >= len(sheet.Images) {
		return fmt.Errorf("image index out of range")
	}

	img := sheet.Images[imageIndex]
	return os.WriteFile(outputPath, img.Data, 0644)
}

// GetImageData returns image data as base64.
func (x *XLSX) GetImageData(sheetName string, imageIndex int) (string, error) {
	sheet := x.sheets[sheetName]
	if sheet == nil {
		return "", fmt.Errorf("sheet not found: %s", sheetName)
	}

	if imageIndex < 0 || imageIndex >= len(sheet.Images) {
		return "", fmt.Errorf("image index out of range")
	}

	img := sheet.Images[imageIndex]
	return base64.StdEncoding.EncodeToString(img.Data), nil
}

// Save saves the workbook to a file.
func (x *XLSX) Save(path string) error {
	if path == "" {
		path = x.filePath
	}
	if path == "" {
		return fmt.Errorf("no file path specified")
	}

	// Create a new xlsx file
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	// Write content types
	if err := x.writeContentTypes(zipWriter); err != nil {
		return err
	}

	// Write workbook
	if err := x.writeWorkbook(zipWriter); err != nil {
		return err
	}

	// Write workbook relationships
	if err := x.writeWorkbookRels(zipWriter); err != nil {
		return err
	}

	// Write shared strings
	if err := x.writeSharedStrings(zipWriter); err != nil {
		return err
	}

	// Write sheets
	for i, name := range x.sheetOrder {
		if err := x.writeSheet(zipWriter, i+1, x.sheets[name]); err != nil {
			return err
		}
	}

	x.modified = false
	return nil
}

// writeContentTypes writes [Content_Types].xml
func (x *XLSX) writeContentTypes(w *zip.Writer) error {
	f, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
`
	for i := range x.sheetOrder {
		content += fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
`, i+1)
	}
	content += `<Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
</Types>`

	_, err = f.Write([]byte(content))
	return err
}

// writeWorkbook writes xl/workbook.xml
func (x *XLSX) writeWorkbook(w *zip.Writer) error {
	f, err := w.Create("xl/workbook.xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets>
`
	for i, name := range x.sheetOrder {
		content += fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>
`, escapeXML(name), i+1, i+1)
	}
	content += `</sheets>
</workbook>`

	_, err = f.Write([]byte(content))
	return err
}

// writeWorkbookRels writes xl/_rels/workbook.xml.rels
func (x *XLSX) writeWorkbookRels(w *zip.Writer) error {
	f, err := w.Create("xl/_rels/workbook.xml.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
`
	rid := 1
	for range x.sheetOrder {
		content += fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>
`, rid, rid)
		rid++
	}
	content += fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`, rid)

	_, err = f.Write([]byte(content))
	return err
}

// writeSharedStrings writes xl/sharedStrings.xml
func (x *XLSX) writeSharedStrings(w *zip.Writer) error {
	f, err := w.Create("xl/sharedStrings.xml")
	if err != nil {
		return err
	}

	// Collect all string values
	strings := make(map[string]int)
	strList := []string{}
	for _, sheet := range x.sheets {
		for _, cell := range sheet.Cells {
			if cell.Type == "str" || cell.Type == "" {
				if _, exists := strings[cell.Value]; !exists && cell.Value != "" {
					strings[cell.Value] = len(strList)
					strList = append(strList, cell.Value)
				}
			}
		}
	}

	// Update cell references to use shared strings
	for _, sheet := range x.sheets {
		for _, cell := range sheet.Cells {
			if (cell.Type == "str" || cell.Type == "") && cell.Value != "" {
				if idx, exists := strings[cell.Value]; exists {
					cell.Type = "s"
					cell.Value = strconv.Itoa(idx)
				}
			}
		}
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="%d" uniqueCount="%d">
`, len(strList), len(strList))

	for _, s := range strList {
		content += fmt.Sprintf(`<si><t>%s</t></si>
`, escapeXML(s))
	}
	content += `</sst>`

	_, err = f.Write([]byte(content))
	return err
}

// writeSheet writes a worksheet.
func (x *XLSX) writeSheet(w *zip.Writer, index int, sheet *Sheet) error {
	f, err := w.Create(fmt.Sprintf("xl/worksheets/sheet%d.xml", index))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<sheetData>
`)

	// Sort cells by row then column
	type cellRef struct {
		row int
		col int
		ref string
		cell *Cell
	}
	cells := make([]cellRef, 0, len(sheet.Cells))
	for ref, cell := range sheet.Cells {
		col, row := parseRef(ref)
		cells = append(cells, cellRef{row: row, col: colToNum(col), ref: ref, cell: cell})
	}
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].row != cells[j].row {
			return cells[i].row < cells[j].row
		}
		return cells[i].col < cells[j].col
	})

	// Write cells grouped by row
	currentRow := 0
	for _, c := range cells {
		if c.row != currentRow {
			if currentRow > 0 {
				buf.WriteString(`</row>
`)
			}
			buf.WriteString(fmt.Sprintf(`<row r="%d">
`, c.row))
			currentRow = c.row
		}

		typeAttr := ""
		if c.cell.Type != "" && c.cell.Type != "n" {
			typeAttr = fmt.Sprintf(` t="%s"`, c.cell.Type)
		}

		buf.WriteString(fmt.Sprintf(`<c r="%s"%s>`, c.ref, typeAttr))
		if c.cell.Formula != "" {
			buf.WriteString(fmt.Sprintf(`<f>%s</f>`, escapeXML(c.cell.Formula)))
		}
		if c.cell.Value != "" {
			buf.WriteString(fmt.Sprintf(`<v>%s</v>`, escapeXML(c.cell.Value)))
		}
		buf.WriteString(`</c>
`)
	}
	if currentRow > 0 {
		buf.WriteString(`</row>
`)
	}

	buf.WriteString(`</sheetData>
`)

	// Write merges
	if len(sheet.Merges) > 0 {
		buf.WriteString(`<mergeCells count="1">
`)
		for _, merge := range sheet.Merges {
			startRef := numToCol(merge.StartCol) + strconv.Itoa(merge.StartRow)
			endRef := numToCol(merge.EndCol) + strconv.Itoa(merge.EndRow)
			buf.WriteString(fmt.Sprintf(`<mergeCell ref="%s:%s"/>
`, startRef, endRef))
		}
		buf.WriteString(`</mergeCells>
`)
	}

	buf.WriteString(`</worksheet>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// Helper functions

// colToNum converts column letter to number (A=1, B=2, ..., Z=26, AA=27, ...)
func colToNum(col string) int {
	result := 0
	for _, c := range col {
		result = result*26 + int(c-'A') + 1
	}
	return result
}

// numToCol converts column number to letter (1=A, 2=B, ..., 26=Z, 27=AA, ...)
func numToCol(num int) string {
	result := ""
	for num > 0 {
		num--
		result = string(rune('A'+(num%26))) + result
		num /= 26
	}
	return result
}

// parseRef parses a cell reference (e.g., "A1" -> ("A", 1))
func parseRef(ref string) (string, int) {
	re := regexp.MustCompile(`^([A-Z]+)(\d+)$`)
	m := re.FindStringSubmatch(ref)
	if len(m) != 3 {
		return "", 0
	}
	row, _ := strconv.Atoi(m[2])
	return m[1], row
}

// parseRange parses a range string (e.g., "A1:C3" -> 1, 1, 3, 3)
func parseRange(rangeStr string) (startCol, startRow, endCol, endRow int) {
	parts := strings.Split(rangeStr, ":")
	if len(parts) != 2 {
		return 0, 0, 0, 0
	}
	startColStr, startRow := parseRef(parts[0])
	endColStr, endRow := parseRef(parts[1])
	return colToNum(startColStr), startRow, colToNum(endColStr), endRow
}

// escapeXML escapes special characters for XML.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// XLSX utility functions for module

// ColToIndex converts column letter to index (A=1).
func ColToIndex(col string) int {
	return colToNum(col)
}

// IndexToCol converts index to column letter.
func IndexToCol(index int) string {
	return numToCol(index)
}

// ParseCellRef parses a cell reference.
func ParseCellRef(ref string) (col string, row int) {
	return parseRef(ref)
}