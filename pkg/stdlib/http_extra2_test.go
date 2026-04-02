// pkg/stdlib/http_extra2_test.go
// Additional tests for http module to increase coverage.
package stdlib

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// httpCall calls an http module builtin.
func httpCall(name string, args ...objects.Object) objects.Object {
	mod := Get("http")
	if mod == nil {
		return &objects.Error{Message: "http module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestHTTPExtra2_GetMimeType tests http.getMimeType with various extensions.
func TestHTTPExtra2_GetMimeType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/test.txt", "text/plain; charset=utf-8"},
		{"/test.html", "text/html; charset=utf-8"},
		{"/test.json", "application/json"},
		{"/test.xml", "text/xml; charset=utf-8"},
		{"/test.css", "text/css; charset=utf-8"},
		{"/test.js", "text/javascript; charset=utf-8"},
		{"/test.png", "image/png"},
		{"/test.jpg", "image/jpeg"},
		{"/test.gif", "image/gif"},
		{"/unknown.xyz", "application/octet-stream"},
		{"/noext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := httpCall("getMimeType", objects.NewString(tt.path))
			if _, ok := result.(*objects.Error); ok {
				t.Errorf("getMimeType error for %s: %s", tt.path, result.Inspect())
				return
			}
			str, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("Expected String, got %T", result)
			}
			if str.Value != tt.expected {
				t.Errorf("getMimeType(%s) = %s; want %s", tt.path, str.Value, tt.expected)
			}
		})
	}
}

// TestHTTPExtra2_ParseJSON tests http.parseJSON with various inputs.
func TestHTTPExtra2_ParseJSON(t *testing.T) {
	// Helper to create HttpReq with a body
	createReq := func(body string) *objects.HttpReq {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		return objects.NewHttpReq(req)
	}

	tests := []struct {
		name    string
		req     *objects.HttpReq
		json    string
		wantErr bool
	}{
		{"valid object", createReq(`{"name":"test","value":123}`), `{"name":"test","value":123}`, false},
		{"valid array", createReq(`[1,2,3]`), `[1,2,3]`, false},
		{"invalid JSON", createReq(`{invalid}`), `{invalid}`, true},
		{"empty body", createReq(``), ``, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqObj objects.Object
			if tt.req != nil {
				reqObj = tt.req
			} else {
				r := httptest.NewRequest("POST", "/test", strings.NewReader(tt.json))
				reqObj = objects.NewHttpReq(r)
			}
			result := httpCall("parseJSON", reqObj)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("parseJSON expected error, got %v", result)
				}
			} else {
				if _, ok := result.(*objects.Error); ok {
					t.Errorf("parseJSON unexpected error: %s", result.Inspect())
				}
				if result == objects.NULL {
					t.Error("parseJSON returned NULL")
				}
			}
		})
	}
}

// TestHTTPExtra2_GetReqBody tests http.getReqBody.
func TestHTTPExtra2_GetReqBody(t *testing.T) {
	createReq := func(body string) *objects.HttpReq {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		return objects.NewHttpReq(req)
	}

	tests := []struct {
		name string
		body string
		want string
	}{
		{"simple text", "Hello, World!", "Hello, World!"},
		{"JSON", `{"key":"value"}`, `{"key":"value"}`},
		{"empty", "", ""},
		{"multiline", "line1\nline2\nline3", "line1\nline2\nline3"},
		{"unicode", "你好世界", "你好世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createReq(tt.body)
			result := httpCall("getReqBody", req)
			if _, ok := result.(*objects.Error); ok {
				t.Fatalf("getReqBody error: %s", result.Inspect())
			}
			str, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("Expected String, got %T", result)
			}
			if str.Value != tt.want {
				t.Errorf("getReqBody = %q; want %q", str.Value, tt.want)
			}
		})
	}
}

// TestHTTPExtra2_GetReqBodyBytes tests http.getReqBodyBytes.
func TestHTTPExtra2_GetReqBodyBytes(t *testing.T) {
	createReq := func(body string) *objects.HttpReq {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		return objects.NewHttpReq(req)
	}

	tests := []struct {
		name string
		body string
		want []int64
	}{
		{"ASCII", "ABC", []int64{65, 66, 67}},
		{"empty", "", []int64{}},
		// UTF-8 for "你好": 你 (U+4F60) -> E4 BD A0, 好 (U+597D) -> E5 A5 BD
		{"unicode", "你好", []int64{0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createReq(tt.body)
			result := httpCall("getReqBodyBytes", req)
			if _, ok := result.(*objects.Error); ok {
				t.Fatalf("getReqBodyBytes error: %s", result.Inspect())
			}
			arr, ok := result.(*objects.Array)
			if !ok {
				t.Fatalf("Expected Array, got %T", result)
			}
			got := make([]int64, len(arr.Elements))
			for i, el := range arr.Elements {
				n, ok := el.(*objects.Int)
				if !ok {
					t.Fatalf("Element %d is not Int: %T", i, el)
				}
				got[i] = n.Value
			}
			if len(got) != len(tt.want) {
				t.Fatalf("Length mismatch: got %d, want %d", len(got), len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("Index %d: got %d, want %d", i, g, tt.want[i])
				}
			}
		})
	}
}

// TestHTTPExtra2_ErrorHandling tests error handling for all http functions.
func TestHTTPExtra2_ErrorHandling(t *testing.T) {
	// parseJSON with wrong argument type
	result := httpCall("parseJSON", objects.NewString(`{"a":1}`))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseJSON with string should error, got %v", result)
	}
	// parseJSON with nil request (inside HttpReq with nil Value)
	req := &objects.HttpReq{Value: nil}
	result = httpCall("parseJSON", req)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseJSON with nil request should error, got %v", result)
	}
	// parseJSON with no arguments
	result = httpCall("parseJSON")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseJSON with no args should error, got %v", result)
	}

	// getReqBody with wrong type
	result = httpCall("getReqBody", objects.NewString("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBody with string should error, got %v", result)
	}
	// getReqBody with nil request
	result = httpCall("getReqBody", req)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBody with nil request should error, got %v", result)
	}
	// getReqBody with no args
	result = httpCall("getReqBody")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBody with no args should error, got %v", result)
	}

	// getReqBodyBytes with wrong type
	result = httpCall("getReqBodyBytes", objects.NewString("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBodyBytes with string should error, got %v", result)
	}
	// getReqBodyBytes with nil request
	result = httpCall("getReqBodyBytes", req)
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBodyBytes with nil request should error, got %v", result)
	}
	// getReqBodyBytes with no args
	result = httpCall("getReqBodyBytes")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getReqBodyBytes with no args should error, got %v", result)
	}

	// getMimeType with wrong type
	result = httpCall("getMimeType", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getMimeType with int should error, got %v", result)
	}
	// getMimeType with no args
	result = httpCall("getMimeType")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("getMimeType with no args should error, got %v", result)
	}
}
