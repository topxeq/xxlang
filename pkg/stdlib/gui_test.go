// pkg/stdlib/gui_test.go
// Tests for GUI module to increase coverage.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callGuiFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("gui")
	if mod == nil {
		panic("gui module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// TestGui_ModuleExistence tests that the gui module is registered.
func TestGui_ModuleExistence(t *testing.T) {
	mod := Get("gui")
	if mod == nil {
		t.Fatal("gui module not found")
	}
	if mod.Name != "gui" {
		t.Errorf("expected module name 'gui', got %s", mod.Name)
	}
	expected := []string{"createWindow", "setHTML", "loadURL", "evalJS", "bind", "loop", "close", "getVersion", "isClosed"}
	for _, name := range expected {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in gui module", name)
		}
	}
}

// TestGui_CreateWindow_ArgumentValidation tests createWindow argument validation.
func TestGui_CreateWindow_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callGuiFunc("createWindow")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with no args should error")
	}
	// Only two args
	result = callGuiFunc("createWindow", String("title"), Int(800))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with 2 args should error")
	}
	// Wrong type for title
	result = callGuiFunc("createWindow", Int(1), Int(800), Int(600))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with non-string title should error")
	}
	// Wrong type for width
	result = callGuiFunc("createWindow", String("title"), String("800"), Int(600))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with non-int width should error")
	}
	// Wrong type for height
	result = callGuiFunc("createWindow", String("title"), Int(800), String("600"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with non-int height should error")
	}
	// Optional debug arg wrong type (provide userDataFolder first)
	result = callGuiFunc("createWindow", String("title"), Int(800), Int(600), String(""), String("true"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createWindow with non-bool debug should error")
	}
}

// TestGui_SetHTML_ArgumentValidation tests setHTML argument validation.
func TestGui_SetHTML_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callGuiFunc("setHTML")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("setHTML with no args should error")
	}
	// Only one arg
	result = callGuiFunc("setHTML", String("handle"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("setHTML with 1 arg should error")
	}
	// Wrong type for handle
	result = callGuiFunc("setHTML", Int(123), String("html"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("setHTML with non-WebView handle should error")
	}
	// Wrong type for html
	result = callGuiFunc("setHTML", objects.NewWebView(nil), Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("setHTML with non-string html should error")
	}
}

// TestGui_LoadURL_ArgumentValidation tests loadURL argument validation.
func TestGui_LoadURL_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("loadURL")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loadURL with no args should error")
	}
	result = callGuiFunc("loadURL", String("handle"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loadURL with 1 arg should error")
	}
	result = callGuiFunc("loadURL", Int(1), String("url"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loadURL with non-WebView handle should error")
	}
	result = callGuiFunc("loadURL", objects.NewWebView(nil), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loadURL with non-string url should error")
	}
}

// TestGui_EvalJS_ArgumentValidation tests evalJS argument validation.
func TestGui_EvalJS_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("evalJS")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("evalJS with no args should error")
	}
	result = callGuiFunc("evalJS", String("handle"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("evalJS with 1 arg should error")
	}
	result = callGuiFunc("evalJS", Int(1), String("script"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("evalJS with non-WebView handle should error")
	}
	result = callGuiFunc("evalJS", objects.NewWebView(nil), Int(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("evalJS with non-string script should error")
	}
}

// TestGui_Bind_ArgumentValidation tests bind argument validation.
func TestGui_Bind_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("bind")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with no args should error")
	}
	result = callGuiFunc("bind", String("handle"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with 1 arg should error")
	}
	result = callGuiFunc("bind", String("handle"), String("name"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with 2 args should error")
	}
	// Non-WebView handle
	result = callGuiFunc("bind", Int(1), String("name"), String("dummy"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with non-WebView handle should error")
	}
	// Name not string
	result = callGuiFunc("bind", objects.NewWebView(nil), Int(2), String("not a function"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with non-string name should error")
	}
	// Function not function
	result = callGuiFunc("bind", objects.NewWebView(nil), String("fn"), Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("bind with non-function third arg should error")
	}
}

// TestGui_Loop_ArgumentValidation tests loop argument validation.
func TestGui_Loop_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("loop")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loop with no args should error")
	}
	result = callGuiFunc("loop", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("loop with non-WebView arg should error")
	}
}

// TestGui_Close_ArgumentValidation tests close argument validation.
func TestGui_Close_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("close")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("close with no args should error")
	}
	result = callGuiFunc("close", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("close with non-WebView arg should error")
	}
}

// TestGui_GetVersion tests getVersion function (should return string or error).
func TestGui_GetVersion(t *testing.T) {
	result := callGuiFunc("getVersion")
	if result == nil {
		t.Fatal("getVersion returned nil")
	}
	// Could be string (version) or error if WebView2 not installed.
	// Accept either as long as not other type.
	switch result.(type) {
	case *objects.String, *objects.Error:
		// ok
	default:
		t.Fatalf("getVersion returned unexpected type %T", result)
	}
}

// TestGui_IsClosed_ArgumentValidation tests isClosed argument validation.
func TestGui_IsClosed_ArgumentValidation(t *testing.T) {
	result := callGuiFunc("isClosed")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isClosed with no args should error")
	}
	result = callGuiFunc("isClosed", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("isClosed with non-WebView arg should error")
	}
	// Note: testing with a valid WebView would require constructing one,
	// which depends on WebView2 runtime; skipping positive test.
}
