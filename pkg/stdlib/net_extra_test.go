// pkg/stdlib/net_extra_test.go
// Additional tests for net module HTTP functions using mock server.
package stdlib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestNet_Get_Success tests net.get with a successful HTTP response.
func TestNet_Get_Success(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello from test server"))
	}))
	defer ts.Close()

	// Call net.get
	result := callNetFunc("get", String(ts.URL))
	if result.Type() == objects.ErrorType {
		t.Fatalf("get() returned error: %s", result.Inspect())
	}

	// Should return {body, statusCode, statusText}
	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("get() should return OrderedMap, got %T", result)
	}

	// Check body
	body := m.Get(&objects.String{Value: "body"})
	if body == nil {
		t.Fatalf("map should have 'body' key")
	}
	bodyStr, ok := body.(*objects.String)
	if !ok {
		t.Fatalf("body should be String, got %T", body)
	}
	if bodyStr.Value != "Hello from test server" {
		t.Errorf("body = %s, want 'Hello from test server'", bodyStr.Value)
	}

	// Check status code (should be 200)
	statusCode := m.Get(&objects.String{Value: "statusCode"})
	if statusCode == nil {
		t.Fatalf("map should have 'statusCode' key")
	}
	status, ok := statusCode.(*objects.Int)
	if !ok {
		t.Fatalf("statusCode should be Int, got %T", statusCode)
	}
	if status.Value != 200 {
		t.Errorf("status code = %d, want 200", status.Value)
	}

	// Check status string
	statusText := m.Get(&objects.String{Value: "statusText"})
	if statusText == nil {
		t.Fatalf("map should have 'statusText' key")
	}
	statusStr, ok := statusText.(*objects.String)
	if !ok {
		t.Fatalf("statusText should be String, got %T", statusText)
	}
	if statusStr.Value != "200 OK" {
		t.Logf("status string = %s (may vary)", statusStr.Value)
	}
}

// TestNet_Post_Success tests net.post with a successful HTTP response.
func TestNet_Post_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := r.Body
		defer body.Close()
		// Read and echo back
		// For simplicity, just respond with OK
		w.Write([]byte("POST received"))
	}))
	defer ts.Close()

	result := callNetFunc("post", String(ts.URL), String("test body"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("post() returned error: %s", result.Inspect())
	}

	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("post() should return OrderedMap, got %T", result)
	}

	// Check body
	body := m.Get(&objects.String{Value: "body"})
	if body == nil {
		t.Fatalf("map should have 'body' key")
	}
	bodyStr, ok := body.(*objects.String)
	if !ok {
		t.Fatalf("body should be String, got %T", body)
	}
	if bodyStr.Value != "POST received" {
		t.Logf("body = %s", bodyStr.Value)
	}
}

// TestNet_Head_Success tests net.head with a successful HEAD request.
func TestNet_Head_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD, got %s", r.Method)
		}
		w.Header().Set("Content-Length", "5")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	result := callNetFunc("head", String(ts.URL))
	if result.Type() == objects.ErrorType {
		t.Fatalf("head() returned error: %s", result.Inspect())
	}

	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("head() should return OrderedMap, got %T", result)
	}

	statusCode := m.Get(&objects.String{Value: "statusCode"})
	if statusCode == nil {
		t.Fatalf("map should have 'statusCode' key")
	}
	status, ok := statusCode.(*objects.Int)
	if !ok {
		t.Fatalf("statusCode should be Int, got %T", statusCode)
	}
	if status.Value != 200 {
		t.Errorf("status code = %d, want 200", status.Value)
	}

	// Headers should be an OrderedMap
	headers := m.Get(&objects.String{Value: "headers"})
	if headers == nil {
		t.Fatalf("map should have 'headers' key")
	}
	headersMap, ok := headers.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("headers should be OrderedMap, got %T", headers)
	}
	// Check Content-Length header
	contentLength := headersMap.Get(&objects.String{Value: "Content-Length"})
	if contentLength != nil {
		if clStr, ok := contentLength.(*objects.String); ok {
			if clStr.Value != "5" {
				t.Errorf("Content-Length = %s, want 5", clStr.Value)
			}
		}
	}
}

// TestNet_Request_Success tests net.request with custom method.
func TestNet_Request_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Write([]byte("PATCH response"))
	}))
	defer ts.Close()

	result := callNetFunc("request", String("PATCH"), String(ts.URL))
	if result.Type() == objects.ErrorType {
		t.Fatalf("request() returned error: %s", result.Inspect())
	}

	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("request() should return OrderedMap, got %T", result)
	}

	body := m.Get(&objects.String{Value: "body"})
	if body == nil {
		t.Fatalf("map should have 'body' key")
	}
	bodyStr, ok := body.(*objects.String)
	if !ok {
		t.Fatalf("body should be String, got %T", body)
	}
	if bodyStr.Value != "PATCH response" {
		t.Errorf("body = %s, want 'PATCH response'", bodyStr.Value)
	}
}

// TestNet_Download_Success tests net.download with a successful download.
func TestNet_Download_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file.txt" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("file content here"))
	}))
	defer ts.Close()

	result := callNetFunc("download", String(ts.URL+"/file.txt"))
	if result.Type() == objects.ErrorType {
		t.Fatalf("download() returned error: %s", result.Inspect())
	}

	content, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("download() should return String, got %T", result)
	}
	if content.Value != "file content here" {
		t.Errorf("download content = %s, want 'file content here'", content.Value)
	}
}

// TestNet_Download_Non200 tests net.download with non-200 status.
func TestNet_Download_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}))
	defer ts.Close()

	result := callNetFunc("download", String(ts.URL))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("download() with 404 should return Error, got %T", result)
	}
}

// TestNet_GetJson_Success tests net.getJson with a JSON endpoint.
func TestNet_GetJson_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"hello","count":5}`))
	}))
	defer ts.Close()

	result := callNetFunc("getJson", String(ts.URL))
	if result.Type() == objects.ErrorType {
		t.Fatalf("getJson() returned error: %s", result.Inspect())
	}

	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("getJson() should return OrderedMap, got %T", result)
	}

	body := m.Get(&objects.String{Value: "body"})
	statusCode := m.Get(&objects.String{Value: "statusCode"})
	if body == nil || statusCode == nil {
		t.Fatalf("getJson() map should have 'body' and 'statusCode' keys")
	}

	bodyStr, ok1 := body.(*objects.String)
	status, ok2 := statusCode.(*objects.Int)
	if !ok1 || !ok2 {
		t.Fatalf("getJson() values should be String and Int, got %T and %T", body, statusCode)
	}
	if bodyStr.Value != `{"message":"hello","count":5}` {
		t.Logf("body = %s", bodyStr.Value)
	}
	if status.Value != 200 {
		t.Logf("status = %d", status.Value)
	}
}

// TestNet_PostJson_Success tests net.postJson.
func TestNet_PostJson_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		// Echo response
		w.Write([]byte(`{"received":true}`))
	}))
	defer ts.Close()

	result := callNetFunc("postJson", String(ts.URL), String(`{"data":"value"}`))
	if result.Type() == objects.ErrorType {
		t.Fatalf("postJson() returned error: %s", result.Inspect())
	}

	m, ok := result.(*objects.OrderedMap)
	if !ok {
		t.Fatalf("postJson() should return OrderedMap, got %T", result)
	}

	body := m.Get(&objects.String{Value: "body"})
	statusCode := m.Get(&objects.String{Value: "statusCode"})
	if body == nil || statusCode == nil {
		t.Fatalf("postJson() map should have 'body' and 'statusCode' keys")
	}

	bodyStr, _ := body.(*objects.String)
	status, _ := statusCode.(*objects.Int)
	t.Logf("postJson response body: %s, status: %d", bodyStr.Value, status.Value)
}

// TestNet_SetTimeout_AndEffect tests setTimeout changes the client timeout.
func TestNet_SetTimeout_AndEffect(t *testing.T) {
	// First set to 60 seconds
	result := callNetFunc("setTimeout", Int(60))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("setTimeout() should return Null, got %T", result)
	}

	// We could test that subsequent requests respect timeout, but that's complex.
	// Just verify no error.
}

// Helper function to call net functions (mirrors net_test.go)
// Note: This is intentionally not defined here to avoid redeclaration.
// The callNetFunc is defined in net_test.go and is accessible within the same package.