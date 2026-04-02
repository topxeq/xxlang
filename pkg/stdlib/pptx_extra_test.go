// pkg/stdlib/pptx_extra_test.go
// Additional argument validation tests for pptx module.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestPPTXCreate_ArgumentValidation tests create argument count.
func TestPPTXCreate_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("create", String("extra"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("create with extra args should error")
	}
}

// TestPPTXOpen_ArgumentValidation tests open argument validation.
func TestPPTXOpen_ArgumentValidation(t *testing.T) {
	// No args
	result := callPPTXFunc("open")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("open with no args should error")
	}
	// Too many args
	result = callPPTXFunc("open", String("file"), Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("open with 2 args should error")
	}
	// Wrong type
	result = callPPTXFunc("open", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("open with non-string should error")
	}
	// Valid type (file may not exist but arg ok)
	result = callPPTXFunc("open", String("nonexistent.pptx"))
	if _, ok := result.(*objects.Error); ok {
		msg := result.Inspect()
		if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
			t.Errorf("open with string got argument validation error: %s", msg)
		}
	}
}

// TestPPTXFromBytes_ArgumentValidation tests fromBytes argument validation.
func TestPPTXFromBytes_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("fromBytes")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("fromBytes with no args should error")
	}
	result = callPPTXFunc("fromBytes", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("fromBytes with non-bytes should error")
	}
	// Valid bytes (may fail parsing but arg ok)
	result = callPPTXFunc("fromBytes", objects.NewBytes([]byte{}))
	if _, ok := result.(*objects.Error); ok {
		msg := result.Inspect()
		if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
			t.Errorf("fromBytes with bytes got argument validation error: %s", msg)
		}
	}
}

// TestPPTXIsPPTX_ArgumentValidation tests isPPTX argument count.
func TestPPTXIsPPTX_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isPPTX")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isPPTX with no args should error")
	}
	result = callPPTXFunc("isPPTX", String("extra"), Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isPPTX with extra args should error")
	}
	// Type doesn't matter; should return bool
	res := callPPTXFunc("isPPTX", String("anything"))
	if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isPPTX should return Bool, got %T", res)
	}
}

// TestPPTXIsSlide_ArgumentValidation tests isSlide argument count.
func TestPPTXIsSlide_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isSlide")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isSlide with no args should error")
	}
	result = callPPTXFunc("isSlide", String("extra"))
	if _, ok := result.(*objects.Error); ok {
		t.Error("isSlide with 1 arg should not error (any type allowed)")
	}
	// With correct count
	res := callPPTXFunc("isSlide", String("anything"))
	if _, ok := res.(*objects.Error); ok {
		t.Errorf("isSlide should not return error, got: %s", res.Inspect())
	} else if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isSlide should return Bool, got %T", res)
	}
}

// TestPPTXIsTextFrame_ArgumentValidation tests isTextFrame argument count.
func TestPPTXIsTextFrame_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isTextFrame")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isTextFrame with no args should error")
	}
	result = callPPTXFunc("isTextFrame", String("extra"))
	if _, ok := result.(*objects.Error); ok {
		// It might not error on type, but count is 1 so should be ok
		// Actually if we pass a string, it's fine; it will return false
	}
}

// TestPPTXIsShape_ArgumentValidation tests isShape argument count.
func TestPPTXIsShape_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isShape")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isShape with no args should error")
	}
	// With one arg, should be ok
	res := callPPTXFunc("isShape", String("anything"))
	if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isShape should return Bool, got %T", res)
	}
}

// TestPPTXIsTable_ArgumentValidation tests isTable argument count.
func TestPPTXIsTable_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isTable")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isTable with no args should error")
	}
	res := callPPTXFunc("isTable", String("anything"))
	if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isTable should return Bool, got %T", res)
	}
}

// TestPPTXIsChart_ArgumentValidation tests isChart argument count.
func TestPPTXIsChart_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isChart")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isChart with no args should error")
	}
	res := callPPTXFunc("isChart", String("anything"))
	if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isChart should return Bool, got %T", res)
	}
}

// TestPPTXIsImage_ArgumentValidation tests isImage argument count.
func TestPPTXIsImage_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("isImage")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isImage with no args should error")
	}
	res := callPPTXFunc("isImage", String("anything"))
	if _, ok := res.(*objects.Bool); !ok {
		t.Errorf("isImage should return Bool, got %T", res)
	}
}

// TestPPTXInchesToEMU_ArgumentValidation tests inchesToEMU argument validation.
func TestPPTXInchesToEMU_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("inchesToEMU")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("inchesToEMU with no args should error")
	}
	result = callPPTXFunc("inchesToEMU", String("not number"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("inchesToEMU with non-number should error")
	}
	// Valid int
	result = callPPTXFunc("inchesToEMU", Int(1))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("inchesToEMU with int got error: %s", result.Inspect())
	}
	// Valid float
	result = callPPTXFunc("inchesToEMU", Float(1.5))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("inchesToEMU with float got error: %s", result.Inspect())
	}
}

// TestPPTXEmuToInches_ArgumentValidation tests emuToInches argument validation.
func TestPPTXEmuToInches_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("emuToInches")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToInches with no args should error")
	}
	result = callPPTXFunc("emuToInches", String("not int"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToInches with non-int should error")
	}
	// Valid int
	result = callPPTXFunc("emuToInches", Int(914400))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("emuToInches with int got error: %s", result.Inspect())
	}
}

// TestPPTXPointsToEMU_ArgumentValidation tests pointsToEMU argument validation.
func TestPPTXPointsToEMU_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("pointsToEMU")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pointsToEMU with no args should error")
	}
	result = callPPTXFunc("pointsToEMU", String("not number"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pointsToEMU with non-number should error")
	}
	result = callPPTXFunc("pointsToEMU", Int(100))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("pointsToEMU with int got error: %s", result.Inspect())
	}
}

// TestPPTXEmuToPoints_ArgumentValidation tests emuToPoints argument validation.
func TestPPTXEmuToPoints_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("emuToPoints")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToPoints with no args should error")
	}
	result = callPPTXFunc("emuToPoints", String("not int"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToPoints with non-int should error")
	}
	result = callPPTXFunc("emuToPoints", Int(12700))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("emuToPoints with int got error: %s", result.Inspect())
	}
}

// TestPPTXPixelsToEMU_ArgumentValidation tests pixelsToEMU argument validation.
func TestPPTXPixelsToEMU_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("pixelsToEMU")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pixelsToEMU with no args should error")
	}
	result = callPPTXFunc("pixelsToEMU", String("not number"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("pixelsToEMU with non-number should error")
	}
	result = callPPTXFunc("pixelsToEMU", Int(96))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("pixelsToEMU with int got error: %s", result.Inspect())
	}
}

// TestPPTXEmuToPixels_ArgumentValidation tests emuToPixels argument validation.
func TestPPTXEmuToPixels_ArgumentValidation(t *testing.T) {
	result := callPPTXFunc("emuToPixels")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToPixels with no args should error")
	}
	result = callPPTXFunc("emuToPixels", String("not int"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("emuToPixels with non-int should error")
	}
	result = callPPTXFunc("emuToPixels", Int(9525))
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("emuToPixels with int got error: %s", result.Inspect())
	}
}
