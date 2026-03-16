// pkg/stdlib/net_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callNetFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("net")
	if mod == nil {
		panic("net module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestNetIsOK(t *testing.T) {
	tests := []struct {
		code     int64
		expected bool
	}{
		{200, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}
	for _, tt := range tests {
		result := callNetFunc("isOK", Int(tt.code))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("isOK() should return Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isOK(%d) = %v, want %v", tt.code, b.Value, tt.expected)
		}
	}
}

func TestNetIsRedirect(t *testing.T) {
	tests := []struct {
		code     int64
		expected bool
	}{
		{200, false},
		{301, true},
		{302, true},
		{399, true},
		{400, false},
	}
	for _, tt := range tests {
		result := callNetFunc("isRedirect", Int(tt.code))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("isRedirect() should return Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isRedirect(%d) = %v, want %v", tt.code, b.Value, tt.expected)
		}
	}
}

func TestNetIsClientError(t *testing.T) {
	tests := []struct {
		code     int64
		expected bool
	}{
		{200, false},
		{400, true},
		{404, true},
		{499, true},
		{500, false},
	}
	for _, tt := range tests {
		result := callNetFunc("isClientError", Int(tt.code))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("isClientError() should return Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isClientError(%d) = %v, want %v", tt.code, b.Value, tt.expected)
		}
	}
}

func TestNetIsServerError(t *testing.T) {
	tests := []struct {
		code     int64
		expected bool
	}{
		{200, false},
		{400, false},
		{500, true},
		{503, true},
	}
	for _, tt := range tests {
		result := callNetFunc("isServerError", Int(tt.code))
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("isServerError() should return Bool, got %T", result)
		}
		if b.Value != tt.expected {
			t.Errorf("isServerError(%d) = %v, want %v", tt.code, b.Value, tt.expected)
		}
	}
}

func TestNetSetTimeout(t *testing.T) {
	result := callNetFunc("setTimeout", Int(60))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("setTimeout() should return Null, got %T", result)
	}
}

func TestNetGetErrors(t *testing.T) {
	// Wrong number of args
	result := callNetFunc("get")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("get() with no args should return Error")
	}

	// Wrong type
	result = callNetFunc("get", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("get() with non-string should return Error")
	}
}

func TestNetPostErrors(t *testing.T) {
	// Wrong number of args
	result := callNetFunc("post")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("post() with no args should return Error")
	}

	// Wrong type for URL
	result = callNetFunc("post", Int(42), String("body"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("post() with non-string URL should return Error")
	}
}

func TestNetHeadErrors(t *testing.T) {
	result := callNetFunc("head")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("head() with no args should return Error")
	}
}

func TestNetRequestErrors(t *testing.T) {
	result := callNetFunc("request")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("request() with no args should return Error")
	}
}

func TestNetGetJsonErrors(t *testing.T) {
	result := callNetFunc("getJson")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("getJson() with no args should return Error")
	}
}

func TestNetPostJsonErrors(t *testing.T) {
	result := callNetFunc("postJson")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("postJson() with no args should return Error")
	}
}

func TestNetDownloadErrors(t *testing.T) {
	result := callNetFunc("download")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download() with no args should return Error")
	}
}
