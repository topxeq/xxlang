// pkg/objects/builtin_image.go
// Image processing built-in functions for Xxlang.
// Note: genQr, scanQr, getImageInfo, resizeImage have been moved to the 'image' stdlib module.
// createImage is kept as a builtin function (which is an alias to image.createImage).
package objects

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
)

func init() {
	Builtins["createImage"] = &Builtin{Fn: builtinCreateImage}
}

// ImageObj represents an image object in Xxlang
type ImageObj struct {
	Value image.Image
}

func (i *ImageObj) Type() ObjectType { return "IMAGE" }
func (i *ImageObj) TypeTag() TypeTag { return TypeTag(100) } // Custom type tag
func (i *ImageObj) Inspect() string {
	bounds := i.Value.Bounds()
	return fmt.Sprintf("Image(%dx%d)", bounds.Dx(), bounds.Dy())
}
func (i *ImageObj) ToBool() *Bool    { return TRUE }
func (i *ImageObj) HashKey() HashKey { return HashKey{Type: "IMAGE", Value: 0} }

// builtinCreateImage creates a new solid color image.
// This is kept as a builtin for convenience.
// Usage: createImage(width, height) -> bytes (white PNG)
//
//	createImage(width, height, colorHex) -> bytes
//
// Example:
//
//	img := createImage(100, 100, "#FF0000")  // Red image
//	saveBytes("red.png", img)
func builtinCreateImage(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong number of arguments for createImage. got=%d, want=2-3", len(args))
	}

	width, ok := args[0].(*Int)
	if !ok {
		return newError("width must be INT, got %s", args[0].Type())
	}

	height, ok := args[1].(*Int)
	if !ok {
		return newError("height must be INT, got %s", args[1].Type())
	}

	var c color.Color = color.White
	if len(args) >= 3 {
		if colorStr, ok := args[2].(*String); ok {
			c = ParseHexColor(colorStr.Value)
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
		return newError("failed to encode image: %v", err)
	}

	return &Bytes{Value: buf.Bytes()}
}

// ParseHexColor parses a hex color string.
// This is exported for use by the image stdlib module.
func ParseHexColor(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		r := HexToByte(s[0:2])
		g := HexToByte(s[2:4])
		b := HexToByte(s[4:6])
		return color.RGBA{R: r, G: g, B: b, A: 255}
	}
	if len(s) == 8 {
		r := HexToByte(s[0:2])
		g := HexToByte(s[2:4])
		b := HexToByte(s[4:6])
		a := HexToByte(s[6:8])
		return color.RGBA{R: r, G: g, B: b, A: a}
	}
	return color.White
}

// HexToByte converts a hex string to a byte.
func HexToByte(s string) uint8 {
	var b uint8
	for _, c := range s {
		b <<= 4
		switch {
		case c >= '0' && c <= '9':
			b |= uint8(c - '0')
		case c >= 'a' && c <= 'f':
			b |= uint8(c-'a') + 10
		case c >= 'A' && c <= 'F':
			b |= uint8(c-'A') + 10
		}
	}
	return b
}

// LoadImageFromReader loads an image from an io.Reader.
// This is exported for use by the image stdlib module.
func LoadImageFromReader(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}
