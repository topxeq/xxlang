// pkg/objects/builtin_http_test.go
// Tests for HTTP builtins
package objects

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHttpBuiltins(t *testing.T) {
	RegisterHttpBuiltins()

	if _, ok := Builtins["writeResp"]; !ok {
		t.Error("writeResp should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["setRespHeader"]; !ok {
		t.Error("setRespHeader should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["addRespHeader"]; !ok {
		t.Error("addRespHeader should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["getReqHeader"]; !ok {
		t.Error("getReqHeader should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["getReqHeaders"]; !ok {
		t.Error("getReqHeaders should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["setCookie"]; !ok {
		t.Error("setCookie should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["getCookie"]; !ok {
		t.Error("getCookie should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["getCookies"]; !ok {
		t.Error("getCookies should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["parseForm"]; !ok {
		t.Error("parseForm should be registered after RegisterHttpBuiltins")
	}
	if _, ok := Builtins["status"]; !ok {
		t.Error("status should be registered after RegisterHttpBuiltins")
	}
}

func TestHttpStatusName(t *testing.T) {
	tests := []struct {
		name     string
		code     int64
		expected string
	}{
		{"OK", 200, "OK"},
		{"Created", 201, "Created"},
		{"No Content", 204, "No Content"},
		{"Bad Request", 400, "Bad Request"},
		{"Unauthorized", 401, "Unauthorized"},
		{"Not Found", 404, "Not Found"},
		{"Internal Server Error", 500, "Internal Server Error"},
		{"Unknown", 999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Builtins["httpStatusName"].Fn(NewInt(tt.code))
			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, str.Value)
				}
			} else {
				t.Errorf("expected String, got %s", result.Type())
			}
		})
	}
}

func TestIsHttpReq(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	httpReq := NewHttpReq(req)

	result := Builtins["isHttpReq"].Fn(httpReq)
	if b, ok := result.(*Bool); ok {
		if !b.Value {
			t.Error("expected true for HttpReq")
		}
	} else {
		t.Errorf("expected Bool, got %s", result.Type())
	}

	result = Builtins["isHttpReq"].Fn(NewString("not a request"))
	if b, ok := result.(*Bool); ok {
		if b.Value {
			t.Error("expected false for non-HttpReq")
		}
	} else {
		t.Errorf("expected Bool, got %s", result.Type())
	}
}

func TestIsHttpResp(t *testing.T) {
	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := Builtins["isHttpResp"].Fn(httpResp)
	if b, ok := result.(*Bool); ok {
		if !b.Value {
			t.Error("expected true for HttpResp")
		}
	} else {
		t.Errorf("expected Bool, got %s", result.Type())
	}

	result = Builtins["isHttpResp"].Fn(NewString("not a response"))
	if b, ok := result.(*Bool); ok {
		if b.Value {
			t.Error("expected false for non-HttpResp")
		}
	} else {
		t.Errorf("expected Bool, got %s", result.Type())
	}
}

func TestUrlEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello world", "hello%20world"},
		{"empty", "", ""},
		{"no spaces", "hello", "hello"},
		{"special", "hello/world", "hello%2Fworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Builtins["urlEncode"].Fn(NewString(tt.input))
			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, str.Value)
				}
			} else {
				t.Errorf("expected String, got %s", result.Type())
			}
		})
	}
}

func TestUrlDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello%20world", "hello world"},
		{"empty", "", ""},
		{"no encoding", "hello", "hello"},
		{"slash", "hello%2Fworld", "hello/world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Builtins["urlDecode"].Fn(NewString(tt.input))
			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, str.Value)
				}
			} else {
				t.Errorf("expected String, got %s", result.Type())
			}
		})
	}
}

func TestWriteResp(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := Builtins["writeResp"].Fn(httpResp, NewString("Hello World"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if result != NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if w.Body.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got '%s'", w.Body.String())
	}
}

func TestSetRespHeader(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := Builtins["setRespHeader"].Fn(httpResp, NewString("Content-Type"), NewString("text/plain"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if result != NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestAddRespHeader(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := Builtins["addRespHeader"].Fn(httpResp, NewString("X-Custom"), NewString("value1"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if result != NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if w.Header().Get("X-Custom") != "value1" {
		t.Errorf("expected X-Custom 'value1', got '%s'", w.Header().Get("X-Custom"))
	}
}

func TestGetReqHeader(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	httpReq := NewHttpReq(req)

	result := Builtins["getReqHeader"].Fn(httpReq, NewString("Content-Type"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if str, ok := result.(*String); ok {
		if str.Value != "application/json" {
			t.Errorf("expected 'application/json', got '%s'", str.Value)
		}
	} else {
		t.Errorf("expected String, got %s", result.Type())
	}
}

func TestGetReqHeaders(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "value")
	httpReq := NewHttpReq(req)

	result := Builtins["getReqHeaders"].Fn(httpReq)
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if m, ok := result.(*Map); ok {
		foundContentType := false
		foundXCustom := false
		for _, pair := range m.Pairs {
			if str, ok := pair.Key.(*String); ok {
				if str.Value == "Content-Type" {
					foundContentType = true
				}
				if str.Value == "X-Custom" {
					foundXCustom = true
				}
			}
		}
		if !foundContentType {
			t.Error("expected Content-Type header in result")
		}
		if !foundXCustom {
			t.Error("expected X-Custom header in result")
		}
	} else {
		t.Errorf("expected Map, got %s", result.Type())
	}
}

func TestStatus(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	tests := []struct {
		name     string
		code     int64
		expected int
	}{
		{"OK", 200, 200},
		{"Created", 201, 201},
		{"Not Found", 404, 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Builtins["status"].Fn(httpResp, NewInt(tt.code))
			if err, ok := result.(*Error); ok {
				t.Errorf("unexpected error: %v", err.Message)
				return
			}
			if result != NULL {
				t.Errorf("expected NULL, got %s", result.Type())
			}
		})
	}
}

func TestSetCookie(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := Builtins["setCookie"].Fn(httpResp, NewString("session"), NewString("abc123"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if result != NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session" || cookies[0].Value != "abc123" {
		t.Errorf("expected cookie 'session=abc123', got '%s=%s'", cookies[0].Name, cookies[0].Value)
	}
}

func TestSetCookieWithOptions(t *testing.T) {
	RegisterHttpBuiltins()

	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	opts := NewMapWithCapacity(3)
	opts.Pairs[NewString("path").HashKey()] = MapPair{Key: NewString("path"), Value: NewString("/")}
	opts.Pairs[NewString("secure").HashKey()] = MapPair{Key: NewString("secure"), Value: &Bool{Value: false}}
	opts.Pairs[NewString("httpOnly").HashKey()] = MapPair{Key: NewString("httpOnly"), Value: &Bool{Value: true}}

	result := Builtins["setCookie"].Fn(httpResp, NewString("session"), NewString("abc123"), opts)
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if result != NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Path != "/" {
		t.Errorf("expected path '/', got '%s'", cookies[0].Path)
	}
	if !cookies[0].HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
}

func TestGetCookie(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	httpReq := NewHttpReq(req)

	result := Builtins["getCookie"].Fn(httpReq, NewString("session"))
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if str, ok := result.(*String); ok {
		if str.Value != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", str.Value)
		}
	} else {
		t.Errorf("expected String, got %s", result.Type())
	}
}

func TestGetCookieNotFound(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("GET", "/test", nil)
	httpReq := NewHttpReq(req)

	result := Builtins["getCookie"].Fn(httpReq, NewString("nonexistent"))
	if result != NULL {
		t.Errorf("expected NULL for missing cookie, got %s", result.Type())
	}
}

func TestGetCookies(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	req.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
	httpReq := NewHttpReq(req)

	result := Builtins["getCookies"].Fn(httpReq)
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if m, ok := result.(*Map); ok {
		foundSession := false
		foundTheme := false
		for _, pair := range m.Pairs {
			if str, ok := pair.Key.(*String); ok {
				if str.Value == "session" {
					foundSession = true
				}
				if str.Value == "theme" {
					foundTheme = true
				}
			}
		}
		if !foundSession {
			t.Error("expected session cookie in result")
		}
		if !foundTheme {
			t.Error("expected theme cookie in result")
		}
	} else {
		t.Errorf("expected Map, got %s", result.Type())
	}
}

func TestParseForm(t *testing.T) {
	RegisterHttpBuiltins()

	req := httptest.NewRequest("POST", "/test?foo=bar", strings.NewReader("name=john&age=30"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq := NewHttpReq(req)

	result := Builtins["parseForm"].Fn(httpReq)
	if err, ok := result.(*Error); ok {
		t.Errorf("unexpected error: %v", err.Message)
		return
	}
	if m, ok := result.(*Map); ok {
		foundFoo := false
		foundName := false
		for _, pair := range m.Pairs {
			if str, ok := pair.Key.(*String); ok {
				if str.Value == "foo" {
					foundFoo = true
				}
				if str.Value == "name" {
					foundName = true
				}
			}
		}
		if !foundFoo {
			t.Error("expected foo query param in result")
		}
		if !foundName {
			t.Error("expected name form field in result")
		}
	} else {
		t.Errorf("expected Map, got %s", result.Type())
	}
}

func TestHttpReqGetMember(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	req.Header.Set("Content-Type", "application/json")
	httpReq := NewHttpReq(req)

	tests := []struct {
		name     string
		member   string
		expected string
	}{
		{"method", "method", "GET"},
		{"path", "path", "/test"},
		{"host", "host", "example.com"},
		{"proto", "proto", "HTTP/1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := httpReq.GetMember(tt.member)
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("expected '%s', got '%s'", tt.expected, str.Value)
				}
			} else {
				t.Errorf("expected String, got %s", result.Type())
			}
		})
	}
}

func TestHttpRespGetMember(t *testing.T) {
	w := httptest.NewRecorder()
	httpResp := NewHttpResp(w)

	result := httpResp.GetMember("written")
	if b, ok := result.(*Bool); ok {
		if b.Value {
			t.Error("expected written to be false initially")
		}
	} else {
		t.Errorf("expected Bool, got %s", result.Type())
	}
}

func TestHttpReqNilValue(t *testing.T) {
	httpReq := &HttpReq{Value: nil, Members: make(map[string]Object)}

	result := httpReq.GetMember("method")
	if result != NULL {
		t.Errorf("expected NULL for nil request, got %s", result.Type())
	}
}

func TestHttpRespNilValue(t *testing.T) {
	httpResp := &HttpResp{Value: nil, Members: make(map[string]Object)}

	result := httpResp.GetMember("written")
	if result != NULL {
		t.Errorf("expected NULL for nil response, got %s", result.Type())
	}
}
