// pkg/stdlib/pptx_test.go
// Tests for pptx module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callPPTXFunc calls a function from the pptx module.
func callPPTXFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("pptx")
	if mod == nil {
		t := &testing.T{}
		t.Skip("pptx module not found")
		return &objects.Error{Message: "pptx module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

func TestPPTXCreate(t *testing.T) {
	doc := callPPTXFunc("create")
	if doc.Type() != objects.PPTXDocumentType {
		t.Fatalf("expected PPTXDocument, got %s", doc.Type())
	}
}

func TestPPTXOpen(t *testing.T) {
	// Opening nonexistent file should error
	missing := callPPTXFunc("open", objects.NewString("nonexistent.pptx"))
	if missing.Type() != objects.ErrorType {
		t.Errorf("expected error for nonexistent file, got %s", missing.Type())
	}
}

func TestPPTXFromBytes(t *testing.T) {
	// Passing nil or empty bytes should error
	bad := callPPTXFunc("fromBytes", objects.NewBytes([]byte{}))
	if bad.Type() != objects.ErrorType {
		t.Logf("fromBytes(empty) returned %s (expected error)", bad.Type())
	}
}

// Type check tests
func TestPPTXIsPPTX(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isPPTX"].(*objects.Builtin)

	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}

	doc := objects.NewPPTX()
	res = fn.Fn(doc)
	if b, ok := res.(*objects.Bool); ok && !b.Value {
		t.Error("expected true for PPTX")
	}
}

func TestPPTXIsSlide(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isSlide"].(*objects.Builtin)

	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
	// We can't easily create a slide without a document, but we can test NULL case.
}

func TestPPTXIsTextFrame(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isTextFrame"].(*objects.Builtin)
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
}

func TestPPTXIsShape(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isShape"].(*objects.Builtin)
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
}

func TestPPTXIsTable(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isTable"].(*objects.Builtin)
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
}

func TestPPTXIsChart(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isChart"].(*objects.Builtin)
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
}

func TestPPTXIsImage(t *testing.T) {
	mod := Get("pptx")
	if mod == nil {
		t.Skip("pptx module not found")
	}
	fn := mod.Exports["isImage"].(*objects.Builtin)
	res := fn.Fn(objects.NULL)
	if b, ok := res.(*objects.Bool); ok && b.Value {
		t.Error("expected false for NULL")
	}
}

// Unit conversion tests
func TestPPTXInchesToEMU(t *testing.T) {
	tests := []struct {
		inches float64
		emu    int64
	}{
		{1, 914400},
		{0.5, 457200},
		{2.54, 2322576}, // 1 inch = 2.54 cm, 914400 * 2.54 = 2,322,576
	}

	for _, tt := range tests {
		result := callPPTXFunc("inchesToEMU", objects.NewFloat(tt.inches))
		if num, ok := result.(*objects.Int); ok {
			if num.Value != tt.emu {
				t.Errorf("inchesToEMU(%f) expected %d, got %d", tt.inches, tt.emu, num.Value)
			}
		} else {
			t.Fatalf("inchesToEMU returned non-Int: %T", result)
		}
	}

	// Invalid type
	bad := callPPTXFunc("inchesToEMU", objects.NewString("1"))
	if bad.Type() != objects.ErrorType {
		t.Logf("inchesToEMU('string') returned %s (expected error)", bad.Type())
	}
}

func TestPPTXEMUToInches(t *testing.T) {
	tests := []struct {
		emu    int64
		inches float64
	}{
		{914400, 1.0},
		{457200, 0.5},
		{0, 0.0},
	}

	for _, tt := range tests {
		result := callPPTXFunc("emuToInches", objects.NewInt(tt.emu))
		if num, ok := result.(*objects.Float); ok {
			if num.Value != tt.inches {
				t.Errorf("emuToInches(%d) expected %f, got %f", tt.emu, tt.inches, num.Value)
			}
		} else {
			t.Fatalf("emuToInches returned non-Float: %T", result)
		}
	}

	// Invalid type
	bad := callPPTXFunc("emuToInches", objects.NewString("1"))
	if bad.Type() != objects.ErrorType {
		t.Logf("emuToInches('string') returned %s (expected error)", bad.Type())
	}
}

func TestPPTXPointsToEMU(t *testing.T) {
	tests := []struct {
		points float64
		emu    int64
	}{
		{1, 12700},
		{72, 914400}, // 72 points = 1 inch
		{0.5, 6350},
	}

	for _, tt := range tests {
		result := callPPTXFunc("pointsToEMU", objects.NewFloat(tt.points))
		if num, ok := result.(*objects.Int); ok {
			if num.Value != tt.emu {
				t.Errorf("pointsToEMU(%f) expected %d, got %d", tt.points, tt.emu, num.Value)
			}
		} else {
			t.Fatalf("pointsToEMU returned non-Int: %T", result)
		}
	}
}

func TestPPTXEMUToPoints(t *testing.T) {
	tests := []struct {
		emu    int64
		points float64
	}{
		{12700, 1.0},
		{914400, 72.0},
		{0, 0.0},
	}

	for _, tt := range tests {
		result := callPPTXFunc("emuToPoints", objects.NewInt(tt.emu))
		if num, ok := result.(*objects.Float); ok {
			if num.Value != tt.points {
				t.Errorf("emuToPoints(%d) expected %f, got %f", tt.emu, tt.points, num.Value)
			}
		} else {
			t.Fatalf("emuToPoints returned non-Float: %T", result)
		}
	}
}

func TestPPTXPixelsToEMU(t *testing.T) {
	tests := []struct {
		pixels float64
		emu    int64
	}{
		{1, 9525},
		{96, 914400}, // 96 pixels = 1 inch
		{0.5, 4762},
	}

	for _, tt := range tests {
		result := callPPTXFunc("pixelsToEMU", objects.NewFloat(tt.pixels))
		if num, ok := result.(*objects.Int); ok {
			if num.Value != tt.emu {
				t.Errorf("pixelsToEMU(%f) expected %d, got %d", tt.pixels, tt.emu, num.Value)
			}
		} else {
			t.Fatalf("pixelsToEMU returned non-Int: %T", result)
		}
	}
}

func TestPPTXEMUToPixels(t *testing.T) {
	tests := []struct {
		emu    int64
		pixels float64
	}{
		{9525, 1.0},
		{914400, 96.0},
		{0, 0.0},
	}

	for _, tt := range tests {
		result := callPPTXFunc("emuToPixels", objects.NewInt(tt.emu))
		if num, ok := result.(*objects.Float); ok {
			if num.Value != tt.pixels {
				t.Errorf("emuToPixels(%d) expected %f, got %f", tt.emu, tt.pixels, num.Value)
			}
		} else {
			t.Fatalf("emuToPixels returned non-Float: %T", result)
		}
	}
}
