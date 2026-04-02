package stdlib

import (
	"os"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callScanQr calls the scanQr function from the image module.
func callScanQr(arg objects.Object) objects.Object {
	mod := Get("image")
	if mod == nil {
		t := &testing.T{}
		t.Skip("image module not found")
		return &objects.Error{Message: "image module not found"}
	}
	fn, ok := mod.Exports["scanQr"].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "scanQr not found"}
	}
	return fn.Fn(arg)
}

// getImageModuleExport retrieves a builtin export from the image module.
func getImageModuleExport(name string) *objects.Builtin {
	mod := Get("image")
	if mod == nil {
		return nil
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return nil
	}
	return fn
}

// TestScanQr_InvalidImage creates a temp file with non-image data and expects an error.
func TestScanQr_InvalidImage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_*.dat")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, err := tmpFile.Write([]byte{0x00, 0x01, 0x02}); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}
	scanResult := callScanQr(objects.NewString(tmpPath))
	if _, ok := scanResult.(*objects.Error); !ok {
		t.Fatalf("ScanQr expected error for invalid image, got %T", scanResult)
	}
}

// TestScanQr_NoQrInImage creates a minimal valid PNG without a QR code and expects an error.
func TestScanQr_NoQrInImage(t *testing.T) {
	// Minimal 1x1 white PNG
	pngData := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE,
		0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54,
		0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, 0x00,
		0x00, 0x04, 0x00, 0x01, 0x5C, 0x68, 0x9B, 0x51,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}
	tmpFile, err := os.CreateTemp("", "noqr_*.png")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, err := tmpFile.Write(pngData); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}
	scanResult := callScanQr(objects.NewString(tmpPath))
	if _, ok := scanResult.(*objects.Error); !ok {
		t.Fatalf("ScanQr expected error for image without QR, got %T", scanResult)
	}
}

// TestScanQr_InvalidPath calls ScanQr with a non-existent file path and expects an error.
func TestScanQr_InvalidPath(t *testing.T) {
	scanResult := callScanQr(objects.NewString("/nonexistent/path/file.png"))
	if _, ok := scanResult.(*objects.Error); !ok {
		t.Fatalf("ScanQr expected error for non-existent path, got %T", scanResult)
	}
}
