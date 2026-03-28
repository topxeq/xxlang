// pkg/stdlib/image_test.go
// Tests for image module functions
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestGenQr(t *testing.T) {
	tests := []struct {
		name      string
		args      []objects.Object
		wantError bool
	}{
		{
			name:      "basic QR code",
			args:      []objects.Object{objects.NewString("https://example.com")},
			wantError: false,
		},
		{
			name:      "QR code with size",
			args:      []objects.Object{objects.NewString("Hello World"), objects.NewInt(128)},
			wantError: false,
		},
		{
			name:      "QR code with size and level",
			args:      []objects.Object{objects.NewString("Test"), objects.NewInt(256), objects.NewString("high")},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenQr(tt.args...)

			if tt.wantError {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
				return
			}

			if b, ok := result.(*objects.Bytes); ok {
				if len(b.Value) == 0 {
					t.Error("expected non-empty QR code bytes")
				}
			} else if err, ok := result.(*objects.Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
			} else {
				t.Errorf("expected Bytes, got %T", result)
			}
		})
	}
}

func TestScanQr(t *testing.T) {
	qrResult := GenQr(objects.NewString("TestQRContent"))
	if _, ok := qrResult.(*objects.Error); ok {
		t.Fatal("failed to generate QR code for test")
	}

	qrBytes, ok := qrResult.(*objects.Bytes)
	if !ok {
		t.Fatal("expected Bytes from genQr")
	}

	result := ScanQr(qrBytes)

	if str, ok := result.(*objects.String); ok {
		if str.Value != "TestQRContent" {
			t.Errorf("expected 'TestQRContent', got '%s'", str.Value)
		}
	} else if err, ok := result.(*objects.Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
	} else {
		t.Errorf("expected String, got %T", result)
	}
}

func TestGetImageInfo(t *testing.T) {
	createResult := CreateImage(objects.NewInt(100), objects.NewInt(200), objects.NewString("#FF0000"))
	if _, ok := createResult.(*objects.Error); ok {
		t.Fatal("failed to create image for test")
	}

	imgBytes, ok := createResult.(*objects.Bytes)
	if !ok {
		t.Fatal("expected Bytes from createImage")
	}

	result := GetImageInfo(imgBytes)

	if m, ok := result.(*objects.Map); ok {
		if pair, exists := m.Pairs[objects.NewString("width").HashKey()]; exists {
			if w, ok := pair.Value.(*objects.Int); ok {
				if w.Value != 100 {
					t.Errorf("expected width 100, got %d", w.Value)
				}
			}
		} else {
			t.Error("width not found in result")
		}

		if pair, exists := m.Pairs[objects.NewString("height").HashKey()]; exists {
			if h, ok := pair.Value.(*objects.Int); ok {
				if h.Value != 200 {
					t.Errorf("expected height 200, got %d", h.Value)
				}
			}
		} else {
			t.Error("height not found in result")
		}
	} else if err, ok := result.(*objects.Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
	} else {
		t.Errorf("expected Map, got %T", result)
	}
}

func TestResizeImage(t *testing.T) {
	createResult := CreateImage(objects.NewInt(200), objects.NewInt(200), objects.NewString("#00FF00"))
	if _, ok := createResult.(*objects.Error); ok {
		t.Fatal("failed to create image for test")
	}

	imgBytes, ok := createResult.(*objects.Bytes)
	if !ok {
		t.Fatal("expected Bytes from createImage")
	}

	result := ResizeImage(imgBytes, objects.NewInt(100), objects.NewInt(50))

	if b, ok := result.(*objects.Bytes); ok {
		if len(b.Value) == 0 {
			t.Error("expected non-empty resized image bytes")
		}
	} else if err, ok := result.(*objects.Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
	} else {
		t.Errorf("expected Bytes, got %T", result)
	}
}

func TestCreateImage(t *testing.T) {
	tests := []struct {
		name      string
		args      []objects.Object
		wantError bool
	}{
		{
			name:      "basic image",
			args:      []objects.Object{objects.NewInt(100), objects.NewInt(100)},
			wantError: false,
		},
		{
			name:      "image with color",
			args:      []objects.Object{objects.NewInt(50), objects.NewInt(50), objects.NewString("#FF0000")},
			wantError: false,
		},
		{
			name:      "image with RGBA color",
			args:      []objects.Object{objects.NewInt(50), objects.NewInt(50), objects.NewString("#FF000080")},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateImage(tt.args...)

			if tt.wantError {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
				return
			}

			if b, ok := result.(*objects.Bytes); ok {
				if len(b.Value) == 0 {
					t.Error("expected non-empty image bytes")
				}
			} else if err, ok := result.(*objects.Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
			} else {
				t.Errorf("expected Bytes, got %T", result)
			}
		})
	}
}
