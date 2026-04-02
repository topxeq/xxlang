// pkg/stdlib/http_test.go
// Tests for http module.
package stdlib

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// mockHttpReq creates a mock *objects.HttpReq with given method, URL string, and body.
func mockHttpReq(method, urlStr, body string) *objects.HttpReq {
	req, _ := http.NewRequest(method, urlStr, strings.NewReader(body))
	// Parse URL to populate query parameters
	if req.URL == nil {
		req.URL = &url.URL{}
	}
	// We can also set query map if needed by parsing url.RawQuery
	return &objects.HttpReq{Value: req}
}

// Helper to build URL with query params
func mockHttpReqWithQuery(method, baseURL string, query map[string]string, body string) *objects.HttpReq {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(method, u.String(), strings.NewReader(body))
	return &objects.HttpReq{Value: req}
}

func TestHTTPParseJSON(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["parseJSON"].(*objects.Builtin)

	// Valid JSON object
	req := mockHttpReq("POST", "/", `{"name":"Alice","age":30}`)
	result := fn.Fn(req)
	if result.Type() != objects.MapType {
		t.Fatalf("expected Map, got %s", result.Type())
	}
	m := result.(*objects.Map)
	if nameVal, ok := m.Pairs[objects.NewString("name").HashKey()].Value.(*objects.String); ok {
		if nameVal.Value != "Alice" {
			t.Errorf("expected name Alice, got %s", nameVal.Value)
		}
	} else {
		t.Error("expected string for name")
	}
	if ageVal, ok := m.Pairs[objects.NewString("age").HashKey()].Value.(*objects.Int); ok {
		if ageVal.Value != 30 {
			t.Errorf("expected age 30, got %d", ageVal.Value)
		}
	} else {
		t.Error("expected int for age")
	}

	// Invalid JSON
	req = mockHttpReq("POST", "/", `{invalid}`)
	result = fn.Fn(req)
	if result.Type() != objects.ErrorType {
		t.Fatalf("expected Error for invalid JSON, got %s", result.Type())
	}

	// Nil request
	result = fn.Fn(objects.NULL)
	if result.Type() != objects.ErrorType {
		t.Fatalf("expected Error for nil request, got %s", result.Type())
	}

	// Empty body
	req = mockHttpReq("POST", "/", "")
	result = fn.Fn(req)
	if result.Type() != objects.ErrorType {
		t.Fatalf("expected Error for empty body, got %s", result.Type())
	}
}

func TestHTTPGetReqBody(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getReqBody"].(*objects.Builtin)

	req := mockHttpReq("GET", "/", "Hello, World!")
	result := fn.Fn(req)
	if str, ok := result.(*objects.String); ok {
		if str.Value != "Hello, World!" {
			t.Errorf("expected 'Hello, World!', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}

	// Nil request
	result = fn.Fn(objects.NULL)
	if result.Type() != objects.ErrorType {
		t.Fatalf("expected Error for nil request, got %s", result.Type())
	}
}

func TestHTTPGetReqBodyBytes(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getReqBodyBytes"].(*objects.Builtin)

	req := mockHttpReq("GET", "/", "ABC")
	result := fn.Fn(req)
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(arr.Elements))
	}
	expected := []byte{65, 66, 67} // 'A','B','C'
	for i, b := range expected {
		if elem, ok := arr.Elements[i].(*objects.Int); ok {
			if elem.Value != int64(b) {
				t.Errorf("byte %d: expected %d, got %d", i, b, elem.Value)
			}
		} else {
			t.Fatalf("expected Int at index %d, got %T", i, arr.Elements[i])
		}
	}
}

func TestHTTPGetMimeType(t *testing.T) {
	mod := Get("http")
	if mod == nil {
		t.Skip("http module not found")
	}
	fn := mod.Exports["getMimeType"].(*objects.Builtin)

	// Test .html file (may include charset on some platforms)
	result := fn.Fn(objects.NewString("index.html"))
	if str, ok := result.(*objects.String); ok {
		if str.Value != "text/html" && str.Value != "text/html; charset=utf-8" {
			t.Errorf("expected 'text/html' (or with charset), got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}

	// Test .json file
	result = fn.Fn(objects.NewString("data.json"))
	if str, ok := result.(*objects.String); ok {
		if str.Value != "application/json" {
			t.Errorf("expected 'application/json', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}

	// Test unknown extension
	result = fn.Fn(objects.NewString("file.unknown"))
	if str, ok := result.(*objects.String); ok {
		if str.Value != "application/octet-stream" {
			t.Errorf("expected 'application/octet-stream', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}

	// Test no extension
	result = fn.Fn(objects.NewString("noextension"))
	if str, ok := result.(*objects.String); ok {
		if str.Value != "application/octet-stream" {
			t.Errorf("expected 'application/octet-stream', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", result)
	}
}
