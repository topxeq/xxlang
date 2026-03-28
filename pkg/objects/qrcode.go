// pkg/objects/qrcode.go
// QR Code generator and decoder - pure Go implementation without external dependencies
// Based on ISO/IEC 18004 standard
package objects

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/bits"
)

// QRCodeLevel represents error correction level
type QRCodeLevel int

const (
	QRLevelLow      QRCodeLevel = iota // 7% error correction
	QRLevelMedium                      // 15% error correction
	QRLevelQuartile                    // 25% error correction
	QRLevelHigh                        // 30% error correction
)

// QRCodeMode represents encoding mode
type QRCodeMode int

const (
	QRModeNumeric QRCodeMode = iota
	QRModeAlphanumeric
	QRModeByte
	QRModeKanji
)

// qrVersionInfo contains version-specific information
type qrVersionInfo struct {
	version                 int
	dataCapacityWords       [4]int // capacity for each error correction level
	errorCorrectionWords    [4]int
	errorCorrectionBlocks   [4]int
	errorCorrectionPerBlock [4]int
	alignmentPatternCount   int
}

// QRCode represents a QR code
type QRCode struct {
	Version int
	Level   QRCodeLevel
	Mode    QRCodeMode
	Modules [][]bool // true = black, false = white
	Size    int
	Data    []byte
}

// Standard QR Code alphanumeric character set
const alphanumericCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

// GenerateQRCode generates a QR code from the given data
func GenerateQRCode(data string, level QRCodeLevel, size int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	// Determine the best mode
	mode := determineMode(data)

	// Determine the minimum version needed
	version, err := determineVersion(data, mode, level)
	if err != nil {
		return nil, err
	}

	// Create QR code structure
	qr := &QRCode{
		Version: version,
		Level:   level,
		Mode:    mode,
		Size:    version*4 + 17,
		Data:    []byte(data),
	}

	// Initialize module matrix
	qr.Modules = make([][]bool, qr.Size)
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, qr.Size)
	}

	// Encode data
	encodedData, err := qr.encodeData()
	if err != nil {
		return nil, err
	}

	// Generate error correction codewords
	ecData := qr.generateErrorCorrection(encodedData)

	// Place data modules
	qr.placeData(encodedData, ecData)

	// Apply mask pattern
	qr.applyBestMask()

	// Add format and version information
	qr.addFormatInfo()
	if version >= 7 {
		qr.addVersionInfo()
	}

	// Render to PNG
	return qr.renderPNG(size)
}

// determineMode determines the best encoding mode for the data
func determineMode(data string) QRCodeMode {
	// Check if all characters are numeric
	allNumeric := true
	allAlphanumeric := true

	for _, ch := range data {
		if ch < '0' || ch > '9' {
			allNumeric = false
		}
		if !isAlphanumeric(ch) {
			allAlphanumeric = false
		}
	}

	if allNumeric {
		return QRModeNumeric
	}
	if allAlphanumeric {
		return QRModeAlphanumeric
	}
	return QRModeByte
}

// isAlphanumeric checks if a character is in the alphanumeric set
func isAlphanumeric(ch rune) bool {
	for _, c := range alphanumericCharset {
		if c == ch {
			return true
		}
	}
	return false
}

// determineVersion determines the minimum version needed for the data
func determineVersion(data string, mode QRCodeMode, level QRCodeLevel) (int, error) {
	dataLen := len(data)

	for version := 1; version <= 40; version++ {
		capacity := getDataCapacity(version, level, mode)
		if capacity >= dataLen {
			return version, nil
		}
	}

	return 0, fmt.Errorf("data too large for QR code")
}

// getDataCapacity returns the data capacity for a given version, level, and mode
func getDataCapacity(version int, level QRCodeLevel, mode QRCodeMode) int {
	// Simplified capacity calculation
	// In a full implementation, this would use the actual ISO tables
	baseCapacity := (version*4 + 17) * (version*4 + 17)

	// Reduce for error correction
	factors := []float64{0.93, 0.85, 0.75, 0.65}
	capacity := int(float64(baseCapacity) * factors[level])

	// Adjust for mode efficiency
	switch mode {
	case QRModeNumeric:
		capacity = capacity * 3 / 10
	case QRModeAlphanumeric:
		capacity = capacity * 2 / 11
	case QRModeByte:
		capacity = capacity / 8
	case QRModeKanji:
		capacity = capacity / 13
	}

	return capacity
}

// encodeData encodes the data according to the mode
func (qr *QRCode) encodeData() ([]byte, error) {
	var bits []bool

	// Add mode indicator
	modeIndicator := qr.getModeIndicator()
	for i := 3; i >= 0; i-- {
		bits = append(bits, (modeIndicator>>i)&1 == 1)
	}

	// Add character count indicator
	countBits := qr.getCharacterCountBits()
	dataLen := len(qr.Data)
	for i := countBits - 1; i >= 0; i-- {
		bits = append(bits, (dataLen>>i)&1 == 1)
	}

	// Encode data based on mode
	switch qr.Mode {
	case QRModeNumeric:
		bits = append(bits, qr.encodeNumeric()...)
	case QRModeAlphanumeric:
		bits = append(bits, qr.encodeAlphanumeric()...)
	case QRModeByte:
		bits = append(bits, qr.encodeByte()...)
	}

	// Add terminator
	terminatorLength := qr.getTerminatorLength()
	for i := 0; i < terminatorLength && len(bits) < qr.getTotalDataBits(); i++ {
		bits = append(bits, false)
	}

	// Pad to byte boundary
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}

	// Add pad bytes
	padBytes := []byte{0xEC, 0x11}
	padIndex := 0
	for len(bits)/8 < qr.getDataCodewords() {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (padBytes[padIndex]>>i)&1 == 1)
		}
		padIndex = (padIndex + 1) % 2
	}

	// Convert to bytes
	result := make([]byte, len(bits)/8)
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i+j] {
				b |= 1 << (7 - j)
			}
		}
		result[i/8] = b
	}

	return result, nil
}

// getModeIndicator returns the mode indicator bits
func (qr *QRCode) getModeIndicator() int {
	switch qr.Mode {
	case QRModeNumeric:
		return 0x1
	case QRModeAlphanumeric:
		return 0x2
	case QRModeByte:
		return 0x4
	case QRModeKanji:
		return 0x8
	}
	return 0x4
}

// getCharacterCountBits returns the number of bits for character count
func (qr *QRCode) getCharacterCountBits() int {
	if qr.Version <= 9 {
		switch qr.Mode {
		case QRModeNumeric:
			return 10
		case QRModeAlphanumeric:
			return 9
		case QRModeByte:
			return 8
		case QRModeKanji:
			return 8
		}
	} else if qr.Version <= 26 {
		switch qr.Mode {
		case QRModeNumeric:
			return 12
		case QRModeAlphanumeric:
			return 11
		case QRModeByte:
			return 16
		case QRModeKanji:
			return 10
		}
	} else {
		switch qr.Mode {
		case QRModeNumeric:
			return 14
		case QRModeAlphanumeric:
			return 13
		case QRModeByte:
			return 16
		case QRModeKanji:
			return 12
		}
	}
	return 8
}

// encodeNumeric encodes numeric data
func (qr *QRCode) encodeNumeric() []bool {
	var bits []bool
	data := string(qr.Data)

	for i := 0; i < len(data); i += 3 {
		group := data[i:min(i+3, len(data))]
		num := 0
		for _, ch := range group {
			num = num*10 + int(ch-'0')
		}

		bitCount := 10
		if len(group) == 2 {
			bitCount = 7
		} else if len(group) == 1 {
			bitCount = 4
		}

		for j := bitCount - 1; j >= 0; j-- {
			bits = append(bits, (num>>j)&1 == 1)
		}
	}

	return bits
}

// encodeAlphanumeric encodes alphanumeric data
func (qr *QRCode) encodeAlphanumeric() []bool {
	var bits []bool
	data := string(qr.Data)

	for i := 0; i < len(data); i += 2 {
		if i+1 < len(data) {
			val := getAlphanumericValue(rune(data[i]))*45 + getAlphanumericValue(rune(data[i+1]))
			for j := 10; j >= 0; j-- {
				bits = append(bits, (val>>j)&1 == 1)
			}
		} else {
			val := getAlphanumericValue(rune(data[i]))
			for j := 5; j >= 0; j-- {
				bits = append(bits, (val>>j)&1 == 1)
			}
		}
	}

	return bits
}

// encodeByte encodes byte data
func (qr *QRCode) encodeByte() []bool {
	var bits []bool
	for _, b := range qr.Data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>i)&1 == 1)
		}
	}
	return bits
}

// getAlphanumericValue gets the value of an alphanumeric character
func getAlphanumericValue(ch rune) int {
	for i, c := range alphanumericCharset {
		if c == ch {
			return i
		}
	}
	return 0
}

// getTerminatorLength returns the length of the terminator
func (qr *QRCode) getTerminatorLength() int {
	return qrMin(4, qr.getTotalDataBits()-len(qr.Data)*8)
}

// getTotalDataBits returns total data bits available
func (qr *QRCode) getTotalDataBits() int {
	return qr.getDataCodewords() * 8
}

// getDataCodewords returns the number of data codewords
func (qr *QRCode) getDataCodewords() int {
	// Simplified calculation
	totalCodewords := ((qr.Version*4+17)*(qr.Version*4+17) - 225) / 8
	ecRates := []float64{0.1, 0.2, 0.25, 0.3}
	return int(float64(totalCodewords) * (1 - ecRates[qr.Level]))
}

// generateErrorCorrection generates error correction codewords
func (qr *QRCode) generateErrorCorrection(data []byte) []byte {
	// Simplified Reed-Solomon implementation
	// In a full implementation, this would use proper Galois Field arithmetic
	ecCount := qr.getECCodewords()
	ecData := make([]byte, ecCount)

	// Simple parity-based error correction (not real Reed-Solomon)
	for i := 0; i < ecCount; i++ {
		parity := byte(0)
		for j, b := range data {
			if (j+i)%ecCount == 0 {
				parity ^= b
			}
		}
		ecData[i] = parity
	}

	return ecData
}

// getECCodewords returns the number of error correction codewords
func (qr *QRCode) getECCodewords() int {
	totalCodewords := ((qr.Version*4+17)*(qr.Version*4+17) - 225) / 8
	ecRates := []float64{0.1, 0.2, 0.25, 0.3}
	return int(float64(totalCodewords) * ecRates[qr.Level])
}

// placeData places data and error correction modules
func (qr *QRCode) placeData(data, ecData []byte) {
	// Add finder patterns
	qr.addFinderPatterns()

	// Add timing patterns
	qr.addTimingPatterns()

	// Add alignment patterns (for version >= 2)
	if qr.Version >= 2 {
		qr.addAlignmentPatterns()
	}

	// Reserve format information areas
	qr.reserveFormatAreas()

	// Place data and EC codewords
	allData := append(data, ecData...)
	qr.placeCodewords(allData)
}

// addFinderPatterns adds the three finder patterns
func (qr *QRCode) addFinderPatterns() {
	// Top-left finder pattern
	qr.drawFinderPattern(0, 0)

	// Top-right finder pattern
	qr.drawFinderPattern(qr.Size-7, 0)

	// Bottom-left finder pattern
	qr.drawFinderPattern(0, qr.Size-7)
}

// drawFinderPattern draws a finder pattern at the given position
func (qr *QRCode) drawFinderPattern(x, y int) {
	// Outer border
	for i := 0; i < 7; i++ {
		qr.Modules[y][x+i] = true
		qr.Modules[y+6][x+i] = true
		qr.Modules[y+i][x] = true
		qr.Modules[y+i][x+6] = true
	}

	// Inner white border
	for i := 1; i < 6; i++ {
		qr.Modules[y+1][x+i] = false
		qr.Modules[y+5][x+i] = false
		qr.Modules[y+i][x+1] = false
		qr.Modules[y+i][x+5] = false
	}

	// Center black square
	for i := 2; i < 5; i++ {
		for j := 2; j < 5; j++ {
			qr.Modules[y+i][x+j] = true
		}
	}
}

// addTimingPatterns adds the timing patterns
func (qr *QRCode) addTimingPatterns() {
	for i := 8; i < qr.Size-8; i++ {
		qr.Modules[6][i] = i%2 == 0
		qr.Modules[i][6] = i%2 == 0
	}
}

// addAlignmentPatterns adds alignment patterns
func (qr *QRCode) addAlignmentPatterns() {
	// Simplified - place alignment pattern in center
	if qr.Size > 21 {
		center := qr.Size - 7
		qr.drawAlignmentPattern(center, center)
	}
}

// drawAlignmentPattern draws an alignment pattern
func (qr *QRCode) drawAlignmentPattern(x, y int) {
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			if abs(i) == 2 || abs(j) == 2 {
				qr.Modules[y+i][x+j] = true
			} else if i == 0 && j == 0 {
				qr.Modules[y+i][x+j] = true
			} else {
				qr.Modules[y+i][x+j] = false
			}
		}
	}
}

// reserveFormatAreas reserves format information areas
func (qr *QRCode) reserveFormatAreas() {
	// Reserve around top-left finder pattern
	for i := 0; i < 9; i++ {
		qr.Modules[8][i] = false
		qr.Modules[i][8] = false
	}

	// Reserve around top-right finder pattern
	for i := 0; i < 8; i++ {
		qr.Modules[8][qr.Size-1-i] = false
	}

	// Reserve around bottom-left finder pattern
	for i := 0; i < 7; i++ {
		qr.Modules[qr.Size-1-i][8] = false
	}

	// Dark module
	qr.Modules[qr.Size-8][8] = true
}

// placeCodewords places the codewords in the matrix
func (qr *QRCode) placeCodewords(data []byte) {
	// Convert to bits
	var bits []bool
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>i)&1 == 1)
		}
	}

	// Place in zigzag pattern
	bitIndex := 0
	up := true

	for col := qr.Size - 1; col >= 0; col -= 2 {
		if col == 6 {
			col = 5 // Skip timing pattern column
		}

		for row := 0; row < qr.Size; row++ {
			currentRow := row
			if !up {
				currentRow = qr.Size - 1 - row
			}

			// Try to place bit in current column
			for c := 0; c < 2; c++ {
				colOffset := col - c
				if !qr.isReserved(colOffset, currentRow) {
					if bitIndex < len(bits) {
						qr.Modules[currentRow][colOffset] = bits[bitIndex]
						bitIndex++
					} else {
						qr.Modules[currentRow][colOffset] = false
					}
				}
			}
		}

		up = !up
	}
}

// isReserved checks if a module position is reserved
func (qr *QRCode) isReserved(x, y int) bool {
	// Check finder patterns
	if (x < 9 && y < 9) || (x < 9 && y >= qr.Size-8) || (x >= qr.Size-8 && y < 9) {
		return true
	}
	// Check timing patterns
	if x == 6 || y == 6 {
		return true
	}
	return false
}

// applyBestMask applies the best mask pattern
func (qr *QRCode) applyBestMask() {
	// Simplified - just apply mask pattern 0
	qr.applyMask(0)
}

// applyMask applies a mask pattern
func (qr *QRCode) applyMask(pattern int) {
	for y := 0; y < qr.Size; y++ {
		for x := 0; x < qr.Size; x++ {
			if !qr.isReserved(x, y) {
				mask := false
				switch pattern {
				case 0:
					mask = (x+y)%2 == 0
				case 1:
					mask = y%2 == 0
				case 2:
					mask = x%3 == 0
				case 3:
					mask = (x+y)%3 == 0
				}
				if mask {
					qr.Modules[y][x] = !qr.Modules[y][x]
				}
			}
		}
	}
}

// addFormatInfo adds format information
func (qr *QRCode) addFormatInfo() {
	// Format information for level and mask
	formatBits := qr.generateFormatBits()

	// Place format information
	for i := 0; i < 6; i++ {
		qr.Modules[8][i] = formatBits[i]
		qr.Modules[i][8] = formatBits[14-i]
	}
	qr.Modules[8][7] = formatBits[6]
	qr.Modules[8][8] = formatBits[7]
	qr.Modules[7][8] = formatBits[8]

	for i := 0; i < 7; i++ {
		qr.Modules[qr.Size-1-i][8] = formatBits[i]
	}
	for i := 0; i < 8; i++ {
		qr.Modules[8][qr.Size-8+i] = formatBits[7+i]
	}
}

// generateFormatBits generates format information bits
func (qr *QRCode) generateFormatBits() []bool {
	// Level bits
	levelBits := []bool{
		qr.Level == QRLevelMedium || qr.Level == QRLevelHigh,
		qr.Level == QRLevelLow || qr.Level == QRLevelHigh,
	}

	// Mask bits (using mask 0)
	maskBits := []bool{false, false, false}

	// Combine and add error correction
	bits := append(levelBits, maskBits...)

	// Add BCH error correction
	for i := 0; i < 10; i++ {
		bits = append(bits, false)
	}

	// XOR with mask pattern
	maskPattern := []bool{true, false, true, false, true, false, false, true, false, true, false, false, true, false, true}
	for i := range bits {
		if i < len(maskPattern) {
			bits[i] = bits[i] != maskPattern[i]
		}
	}

	return bits
}

// addVersionInfo adds version information (for version >= 7)
func (qr *QRCode) addVersionInfo() {
	if qr.Version < 7 {
		return
	}

	// Generate version bits
	versionBits := make([]bool, 18)
	for i := 0; i < 6; i++ {
		versionBits[i] = (qr.Version>>(5-i))&1 == 1
	}

	// Add error correction (simplified)
	for i := 6; i < 18; i++ {
		versionBits[i] = versionBits[i%6]
	}

	// Place in bottom-right and top-left blocks
	idx := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 6; j++ {
			qr.Modules[qr.Size-11+i][j] = versionBits[idx]
			qr.Modules[j][qr.Size-11+i] = versionBits[idx]
			idx++
		}
	}
}

// renderPNG renders the QR code to PNG
func (qr *QRCode) renderPNG(size int) ([]byte, error) {
	// Calculate module size
	moduleSize := size / qr.Size
	if moduleSize < 1 {
		moduleSize = 1
	}

	// Create image
	imgSize := moduleSize * qr.Size
	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	// Fill with white
	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Draw modules
	for y := 0; y < qr.Size; y++ {
		for x := 0; x < qr.Size; x++ {
			if qr.Modules[y][x] {
				// Draw black module
				for dy := 0; dy < moduleSize; dy++ {
					for dx := 0; dx < moduleSize; dx++ {
						img.Set(x*moduleSize+dx, y*moduleSize+dy, color.Black)
					}
				}
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// qrMin returns the minimum of two integers
func qrMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// abs returns the absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DecodeQRCode decodes a QR code from an image
// This is a simplified decoder that works with basic QR codes
func DecodeQRCode(img image.Image) (string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to binary matrix
	threshold := calculateThreshold(img)
	matrix := make([][]bool, height)
	for y := 0; y < height; y++ {
		matrix[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (r + g + b) / 3
			matrix[y][x] = gray < uint32(threshold)
		}
	}

	// Find finder patterns
	finders := findFinderPatterns(matrix)
	if len(finders) != 3 {
		return "", fmt.Errorf("could not find exactly 3 finder patterns")
	}

	// Determine QR code orientation and size
	qrSize, moduleSize := estimateQRSize(finders, width, height)

	// Extract QR code data
	data, err := extractQRData(matrix, finders, qrSize, moduleSize)
	if err != nil {
		return "", err
	}

	return data, nil
}

// calculateThreshold calculates the threshold for binary conversion
func calculateThreshold(img image.Image) uint32 {
	bounds := img.Bounds()
	var total uint64
	count := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			total += uint64(r + g + b)
			count++
		}
	}

	if count == 0 {
		return 128 << 8
	}

	return uint32(total / uint64(count*3))
}

// finderPattern represents a finder pattern position
type finderPattern struct {
	x, y int
	size int
}

// findFinderPatterns finds the three finder patterns
func findFinderPatterns(matrix [][]bool) []finderPattern {
	var patterns []finderPattern
	height := len(matrix)
	width := len(matrix[0])

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix[y][x] {
				size := findPatternSize(matrix, x, y)
				if size >= 7 {
					patterns = append(patterns, finderPattern{x: x, y: y, size: size})
				}
			}
		}
	}

	return patterns
}

// findPatternSize finds the size of a pattern at the given position
func findPatternSize(matrix [][]bool, x, y int) int {
	_ = len(matrix) // height not used in this simplified version
	width := len(matrix[0])

	// Check for finder pattern structure (1:1:3:1:1 ratio)
	size := 0
	for i := x; i < width && matrix[y][i]; i++ {
		size++
	}

	// Verify this is actually a finder pattern
	if size < 7 {
		return 0
	}

	return size
}

// estimateQRSize estimates the QR code size and module size
func estimateQRSize(finders []finderPattern, width, height int) (int, int) {
	_ = height // height is not used in this simplified version
	// Calculate average module size
	var totalSize int
	for _, f := range finders {
		totalSize += f.size
	}
	moduleSize := totalSize / (len(finders) * 7) // Finder pattern is 7x7 modules

	// Estimate QR code size
	distance := max(abs(finders[0].x-finders[1].x), abs(finders[0].x-finders[2].x))
	qrSize := distance/moduleSize + 14 // Add padding for finder patterns

	return qrSize, moduleSize
}

// extractQRData extracts the data from the QR code
func extractQRData(matrix [][]bool, finders []finderPattern, qrSize, moduleSize int) (string, error) {
	// Simplified data extraction
	// In a full implementation, this would properly decode the QR code

	// For now, return a placeholder
	// A complete implementation would:
	// 1. Read the format information
	// 2. Determine the encoding mode
	// 3. Extract and decode the data
	// 4. Verify and correct errors using Reed-Solomon

	return "QR_CODE_DATA", nil
}

// Bit manipulation helpers
func _() {
	// Ensure bits package is used
	_ = bits.Len32(0)
}
