// pkg/stdlib/image.go
// Image processing module for Xxlang.
package stdlib

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/draw"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "image",
		Exports: map[string]objects.Object{
			"genQr":        BuiltinFunc(GenQr),
			"scanQr":       BuiltinFunc(ScanQr),
			"getImageInfo": BuiltinFunc(GetImageInfo),
			"resizeImage":  BuiltinFunc(ResizeImage),
			"createImage":  BuiltinFunc(CreateImage),
		},
	})
}

// GenQr generates a QR code image from text.
// Usage: image.genQr(text) -> bytes (PNG image)
//
//	image.genQr(text, size) -> bytes
//	image.genQr(text, size, level) -> bytes  (level: low, medium, high, highest)
//
// Example:
//
//	qrBytes := image.genQr("https://example.com")
//	saveBytes("qr.png", qrBytes)
//
//	qrBytes := image.genQr("Hello World", 256)
//	qrBytes := image.genQr("Data", 256, "high")
func GenQr(args ...objects.Object) objects.Object {
	if len(args) < 1 || len(args) > 3 {
		return Error("genQr() takes 1-3 arguments")
	}

	text, ok := args[0].(*objects.String)
	if !ok {
		return Error("first argument to 'genQr' must be STRING")
	}

	size := 256
	if len(args) >= 2 {
		if s, ok := args[1].(*objects.Int); ok {
			size = int(s.Value)
		}
	}

	level := objects.QRLevelMedium
	if len(args) >= 3 {
		if l, ok := args[2].(*objects.String); ok {
			switch strings.ToLower(l.Value) {
			case "low":
				level = objects.QRLevelLow
			case "medium":
				level = objects.QRLevelMedium
			case "high", "highest":
				level = objects.QRLevelHigh
			}
		}
	}

	pngBytes, err := objects.GenerateQRCode(text.Value, level, size)
	if err != nil {
		return Error("failed to generate QR code: " + err.Error())
	}

	return &objects.Bytes{Value: pngBytes}
}

// ScanQr scans a QR code from image bytes.
// Usage: image.scanQr(imageBytes) -> string or error
//
// Example:
//
//	imgBytes := loadBytes("qr.png")
//	result := image.scanQr(imgBytes)
func ScanQr(args ...objects.Object) objects.Object {
	if len(args) != 1 {
		return Error("scanQr() takes exactly 1 argument")
	}

	var imgBytes []byte
	switch b := args[0].(type) {
	case *objects.Bytes:
		imgBytes = b.Value
	case *objects.String:
		data, err := os.ReadFile(b.Value)
		if err != nil {
			return Error("failed to read image file: " + err.Error())
		}
		imgBytes = data
	default:
		return Error("argument to 'scanQr' must be BYTES or STRING")
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return Error("failed to decode image: " + err.Error())
	}

	result, err := objects.DecodeQRCode(img)
	if err != nil {
		return Error("no QR code found in image: " + err.Error())
	}

	return objects.NewString(result)
}

// GetImageInfo gets information about an image.
// Usage: image.getImageInfo(imageBytes) -> map
//
//	image.getImageInfo(imagePath) -> map
//
// Returns map with: width, height, format, hasAlpha, bounds
//
// Example:
//
//	info := image.getImageInfo(loadBytes("photo.jpg"))
//	pln("Width:", info["width"])
func GetImageInfo(args ...objects.Object) objects.Object {
	if len(args) != 1 {
		return Error("getImageInfo() takes exactly 1 argument")
	}

	var imgBytes []byte
	switch b := args[0].(type) {
	case *objects.Bytes:
		imgBytes = b.Value
	case *objects.String:
		data, err := os.ReadFile(b.Value)
		if err != nil {
			return Error("failed to read image file: " + err.Error())
		}
		imgBytes = data
	default:
		return Error("argument to 'getImageInfo' must be BYTES or STRING")
	}

	reader := bytes.NewReader(imgBytes)
	img, format, err := image.Decode(reader)
	if err != nil {
		return Error("failed to decode image: " + err.Error())
	}

	bounds := img.Bounds()

	hasAlpha := false
	if rgba, ok := img.(*image.RGBA); ok {
		for y := bounds.Min.Y; y < bounds.Max.Y && !hasAlpha; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := rgba.At(x, y).RGBA()
				if a != 0xffff {
					hasAlpha = true
					break
				}
			}
		}
	} else if _, ok := img.(*image.NRGBA); ok {
		hasAlpha = true
	}

	result := objects.NewMapWithCapacity(5)

	key := objects.NewString("width")
	result.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: objects.NewInt(int64(bounds.Dx()))}

	key = objects.NewString("height")
	result.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: objects.NewInt(int64(bounds.Dy()))}

	key = objects.NewString("format")
	result.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: objects.NewString(format)}

	key = objects.NewString("hasAlpha")
	result.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: &objects.Bool{Value: hasAlpha}}

	key = objects.NewString("bounds")
	boundsMap := objects.NewMapWithCapacity(4)
	bk := objects.NewString("minX")
	boundsMap.Pairs[bk.HashKey()] = objects.MapPair{Key: bk, Value: objects.NewInt(int64(bounds.Min.X))}
	bk = objects.NewString("minY")
	boundsMap.Pairs[bk.HashKey()] = objects.MapPair{Key: bk, Value: objects.NewInt(int64(bounds.Min.Y))}
	bk = objects.NewString("maxX")
	boundsMap.Pairs[bk.HashKey()] = objects.MapPair{Key: bk, Value: objects.NewInt(int64(bounds.Max.X))}
	bk = objects.NewString("maxY")
	boundsMap.Pairs[bk.HashKey()] = objects.MapPair{Key: bk, Value: objects.NewInt(int64(bounds.Max.Y))}
	result.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: boundsMap}

	return result
}

// ResizeImage resizes an image to specified dimensions.
// Usage: image.resizeImage(imageBytes, width, height) -> bytes (PNG)
//
//	image.resizeImage(imageBytes, width, height, format) -> bytes
//
// Example:
//
//	imgBytes := loadBytes("photo.jpg")
//	resized := image.resizeImage(imgBytes, 800, 600)
//	saveBytes("resized.png", resized)
func ResizeImage(args ...objects.Object) objects.Object {
	if len(args) < 3 || len(args) > 4 {
		return Error("resizeImage() takes 3-4 arguments")
	}

	var imgBytes []byte
	switch b := args[0].(type) {
	case *objects.Bytes:
		imgBytes = b.Value
	case *objects.String:
		data, err := os.ReadFile(b.Value)
		if err != nil {
			return Error("failed to read image file: " + err.Error())
		}
		imgBytes = data
	default:
		return Error("first argument to 'resizeImage' must be BYTES or STRING")
	}

	width, ok := args[1].(*objects.Int)
	if !ok {
		return Error("width must be INT")
	}

	height, ok := args[2].(*objects.Int)
	if !ok {
		return Error("height must be INT")
	}

	format := "png"
	if len(args) >= 4 {
		if f, ok := args[3].(*objects.String); ok {
			format = strings.ToLower(f.Value)
		}
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return Error("failed to decode image: " + err.Error())
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(width.Value), int(height.Value)))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	case "png":
		fallthrough
	default:
		err = png.Encode(&buf, dst)
	}

	if err != nil {
		return Error("failed to encode image: " + err.Error())
	}

	return &objects.Bytes{Value: buf.Bytes()}
}

// CreateImage creates a new solid color image.
// Usage: image.createImage(width, height) -> bytes (white PNG)
//
//	image.createImage(width, height, colorHex) -> bytes
//
// Example:
//
//	img := image.createImage(100, 100, "#FF0000")  // Red image
//	saveBytes("red.png", img)
func CreateImage(args ...objects.Object) objects.Object {
	if len(args) < 2 || len(args) > 3 {
		return Error("createImage() takes 2-3 arguments")
	}

	width, ok := args[0].(*objects.Int)
	if !ok {
		return Error("width must be INT")
	}

	height, ok := args[1].(*objects.Int)
	if !ok {
		return Error("height must be INT")
	}

	c := objects.ParseHexColor("#FFFFFF")
	if len(args) >= 3 {
		if colorStr, ok := args[2].(*objects.String); ok {
			c = objects.ParseHexColor(colorStr.Value)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width.Value), int(height.Value)))

	for y := 0; y < int(height.Value); y++ {
		for x := 0; x < int(width.Value); x++ {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return Error("failed to encode image: " + err.Error())
	}

	return &objects.Bytes{Value: buf.Bytes()}
}
