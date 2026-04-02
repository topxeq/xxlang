package objects

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		level   QRCodeLevel
		size    int
		wantErr bool
	}{
		{
			name:    "simple text",
			data:    "hello",
			level:   QRLevelLow,
			size:    200,
			wantErr: false,
		},
		{
			name:    "numeric data",
			data:    "1234567890",
			level:   QRLevelMedium,
			size:    200,
			wantErr: false,
		},
		{
			name:    "alphanumeric data",
			data:    "ABC123",
			level:   QRLevelQuartile,
			size:    200,
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    "",
			level:   QRLevelLow,
			size:    200,
			wantErr: true,
		},
		{
			name:    "high error correction",
			data:    "test data",
			level:   QRLevelHigh,
			size:    150,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateQRCode(tt.data, tt.level, tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQRCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) == 0 {
					t.Error("GenerateQRCode() returned empty data")
				}
			}
		})
	}
}

func TestGenerateQRCodeValidPNG(t *testing.T) {
	data := "test QR code data"
	pngData, err := GenerateQRCode(data, QRLevelLow, 200)
	if err != nil {
		t.Fatalf("GenerateQRCode() error = %v", err)
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Error("Invalid image dimensions")
	}
}

func TestDetermineMode(t *testing.T) {
	tests := []struct {
		data     string
		wantMode QRCodeMode
	}{
		{"12345", QRModeNumeric},
		{"ABC123", QRModeAlphanumeric},
		{"hello world", QRModeByte},
		{"", QRModeNumeric},
		{"0", QRModeNumeric},
		{"A", QRModeAlphanumeric},
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			got := determineMode(tt.data)
			if got != tt.wantMode {
				t.Errorf("determineMode(%q) = %v, want %v", tt.data, got, tt.wantMode)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		ch   rune
		want bool
	}{
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{' ', true},
		{'$', true},
		{'*', true},
		{'+', true},
		{'-', true},
		{'.', true},
		{'/', true},
		{':', true},
		{'a', false},
		{'z', false},
		{'@', false},
		{'#', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.ch), func(t *testing.T) {
			got := isAlphanumeric(tt.ch)
			if got != tt.want {
				t.Errorf("isAlphanumeric(%q) = %v, want %v", tt.ch, got, tt.want)
			}
		})
	}
}

func TestQRCodeEncodeData(t *testing.T) {
	qr := &QRCode{
		Version: 1,
		Level:   QRLevelLow,
		Mode:    QRModeByte,
		Size:    21,
		Data:    []byte("test"),
	}

	data, err := qr.encodeData()
	if err != nil {
		t.Errorf("encodeData() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("encodeData() returned empty data")
	}
}

func TestQRCodeGetModeIndicator(t *testing.T) {
	tests := []struct {
		mode QRCodeMode
		want int
	}{
		{QRModeNumeric, 0x1},
		{QRModeAlphanumeric, 0x2},
		{QRModeByte, 0x4},
		{QRModeKanji, 0x8},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			qr := &QRCode{Mode: tt.mode}
			got := qr.getModeIndicator()
			if got != tt.want {
				t.Errorf("getModeIndicator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQRCodeGetCharacterCountBits(t *testing.T) {
	tests := []struct {
		version int
		mode    QRCodeMode
		want    int
	}{
		{1, QRModeNumeric, 10},
		{1, QRModeAlphanumeric, 9},
		{1, QRModeByte, 8},
		{10, QRModeNumeric, 12},
		{10, QRModeByte, 16},
		{30, QRModeNumeric, 14},
		{30, QRModeByte, 16},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			qr := &QRCode{Version: tt.version, Mode: tt.mode}
			got := qr.getCharacterCountBits()
			if got != tt.want {
				t.Errorf("getCharacterCountBits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQRCodePlaceData(t *testing.T) {
	qr := &QRCode{
		Version: 1,
		Level:   QRLevelLow,
		Mode:    QRModeByte,
		Size:    21,
		Data:    []byte("test"),
		Modules: make([][]bool, 21),
	}
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, 21)
	}

	data := []byte{0x01, 0x02, 0x03, 0x04}
	ecData := []byte{0x05, 0x06}

	qr.placeData(data, ecData)
}

func TestQRCodeAddFinderPatterns(t *testing.T) {
	qr := &QRCode{
		Version: 1,
		Size:    21,
		Modules: make([][]bool, 21),
	}
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, 21)
	}

	qr.addFinderPatterns()

	if !qr.Modules[0][0] {
		t.Error("Top-left finder pattern not set correctly")
	}
	if !qr.Modules[0][14] {
		t.Error("Top-right finder pattern not set correctly")
	}
	if !qr.Modules[14][0] {
		t.Error("Bottom-left finder pattern not set correctly")
	}
}

func TestQRCodeAddTimingPatterns(t *testing.T) {
	qr := &QRCode{
		Version: 1,
		Size:    21,
		Modules: make([][]bool, 21),
	}
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, 21)
	}

	qr.addTimingPatterns()

	if qr.Modules[6][8] == qr.Modules[6][9] {
		t.Error("Timing pattern should alternate")
	}
}

func TestQRCodeRenderPNG(t *testing.T) {
	qr := &QRCode{
		Version: 1,
		Size:    21,
		Modules: make([][]bool, 21),
	}
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, 21)
	}

	for y := 0; y < 21; y++ {
		for x := 0; x < 21; x++ {
			qr.Modules[y][x] = (x+y)%2 == 0
		}
	}

	pngData, err := qr.renderPNG(200)
	if err != nil {
		t.Errorf("renderPNG() error = %v", err)
	}
	if len(pngData) == 0 {
		t.Error("renderPNG() returned empty data")
	}
}

func TestQRMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := qrMin(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("qrMin(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		x, want int
	}{
		{1, 1},
		{-1, 1},
		{0, 0},
		{-5, 5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := abs(tt.x)
			if got != tt.want {
				t.Errorf("abs(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}

func TestDecodeQRCode(t *testing.T) {
	pngData, err := GenerateQRCode("test", QRLevelLow, 100)
	if err != nil {
		t.Fatalf("GenerateQRCode() error = %v", err)
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	_, err = DecodeQRCode(img)
	_ = err
}

func TestDecodeQRCodeInvalidImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}

	_, err := DecodeQRCode(img)
	if err == nil {
		t.Error("DecodeQRCode() should return error for invalid image")
	}
}

func TestCalculateThreshold(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.Gray{Y: 128})
		}
	}

	threshold := calculateThreshold(img)
	if threshold == 0 {
		t.Error("calculateThreshold() returned 0")
	}
}

func TestFindFinderPatterns(t *testing.T) {
	matrix := make([][]bool, 50)
	for i := range matrix {
		matrix[i] = make([]bool, 50)
	}

	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			matrix[y][x] = true
		}
	}

	patterns := findFinderPatterns(matrix)
	if len(patterns) == 0 {
		t.Error("findFinderPatterns() should find at least one pattern")
	}
}

func TestGetDataCapacity(t *testing.T) {
	tests := []struct {
		version int
		level   QRCodeLevel
		mode    QRCodeMode
	}{
		{1, QRLevelLow, QRModeNumeric},
		{1, QRLevelMedium, QRModeByte},
		{5, QRLevelHigh, QRModeAlphanumeric},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			capacity := getDataCapacity(tt.version, tt.level, tt.mode)
			if capacity <= 0 {
				t.Errorf("getDataCapacity() = %d, want > 0", capacity)
			}
		})
	}
}
