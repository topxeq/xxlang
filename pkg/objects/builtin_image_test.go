// pkg/objects/builtin_image_test.go
// Tests for image built-in functions
package objects

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestBuiltinCreateImage(t *testing.T) {
	tests := []struct {
		name        string
		args        []Object
		wantError   bool
		checkResult func(t *testing.T, result Object)
	}{
		{
			name:      "invalid width type",
			args:      []Object{NewString("100"), NewInt(100)},
			wantError: true,
		},
		{
			name:      "invalid height type",
			args:      []Object{NewInt(100), NewString("100")},
			wantError: true,
		},
		{
			name:      "missing required arguments",
			args:      []Object{NewInt(100)},
			wantError: true,
		},
		{
			name:      "too many arguments",
			args:      []Object{NewInt(100), NewInt(100), NewString("#FF0000"), NewString("extra")},
			wantError: true,
		},
		{
			name:      "valid image without color",
			args:      []Object{NewInt(10), NewInt(10)},
			wantError: false,
			checkResult: func(t *testing.T, result Object) {
				bytesObj, ok := result.(*Bytes)
				if !ok {
					t.Errorf("expected Bytes, got %T", result)
				}
				if len(bytesObj.Value) == 0 {
					t.Errorf("expected non-empty bytes")
				}
				// Should be valid PNG (starts with PNG signature)
				if !bytes.HasPrefix(bytesObj.Value, []byte{0x89, 0x50, 0x4E, 0x47}) {
					t.Errorf("expected PNG format")
				}
			},
		},
		{
			name:      "valid image with color",
			args:      []Object{NewInt(10), NewInt(10), NewString("#FF0000")},
			wantError: false,
			checkResult: func(t *testing.T, result Object) {
				bytesObj, ok := result.(*Bytes)
				if !ok {
					t.Errorf("expected Bytes, got %T", result)
				}
				if len(bytesObj.Value) == 0 {
					t.Errorf("expected non-empty bytes")
				}
			},
		},
		{
			name:      "valid image with short hex color",
			args:      []Object{NewInt(10), NewInt(10), NewString("#F00")},
			wantError: false,
		},
		{
			name:      "valid image with rgb functional notation",
			args:      []Object{NewInt(10), NewInt(10), NewString("rgb(255,0,0)")},
			wantError: false,
		},
		{
			name:      "valid image with named color",
			args:      []Object{NewInt(10), NewInt(10), NewString("red")},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinCreateImage(tt.args...)

			if tt.wantError {
				if _, ok := result.(*Error); !ok {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestImageObj(t *testing.T) {
	// Create a simple 2x2 red image
	img := &ImageObj{
		Value: createTestImage(2, 2, color.RGBA{255, 0, 0, 255}),
	}

	if img.Type() != "IMAGE" {
		t.Errorf("expected type IMAGE, got %s", img.Type())
	}

	if img.TypeTag() != TypeTag(100) {
		t.Errorf("expected TypeTag 100, got %d", img.TypeTag())
	}

	inspect := img.Inspect()
	if !strings.Contains(inspect, "Image") {
		t.Errorf("expected Inspect to contain 'Image', got %s", inspect)
	}

	if !img.ToBool().Value {
		t.Errorf("Image.ToBool() should return true")
	}

	key := img.HashKey()
	if key.Type != "IMAGE" {
		t.Errorf("expected HashKey.Type IMAGE, got %s", key.Type)
	}
}

func TestImageObj_Methods(t *testing.T) {
	img := &ImageObj{
		Value: createTestImage(10, 20, color.RGBA{0, 255, 0, 255}),
	}

	// Test width and height if available
	// Note: These methods are not implemented in the current code
	// but we test the type and basic properties
	if img.Type() != "IMAGE" {
		t.Errorf("wrong type")
	}
}

// createTestImage creates a simple solid color image for testing
func createTestImage(width, height int, col color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	return img
}
