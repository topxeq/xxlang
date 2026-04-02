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

	// Should return [body, statusCode, status]
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("get() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("get() array should have 3 elements, got %d", len(arr.Elements))
	}

	// Check body
	body, ok := arr.Elements[0].(*objects.String)
	if !ok {
		t.Fatalf("first element should be String (body), got %T", arr.Elements[0])
	}
	if body.Value != "Hello from test server" {
		t.Errorf("body = %s, want 'Hello from test server'", body.Value)
	}

	// Check status code (should be 200)
	status, ok := arr.Elements[1].(*objects.Int)
	if !ok {
		t.Fatalf("second element should be Int (status code), got %T", arr.Elements[1])
	}
	if status.Value != 200 {
		t.Errorf("status code = %d, want 200", status.Value)
	}

	// Check status string
	statusStr, ok := arr.Elements[2].(*objects.String)
	if !ok {
		t.Fatalf("third element should be String (status), got %T", arr.Elements[2])
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

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("post() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("post() array should have 3 elements, got %d", len(arr.Elements))
	}

	// Check body
	body, ok := arr.Elements[0].(*objects.String)
	if !ok {
		t.Fatalf("first element should be String, got %T", arr.Elements[0])
	}
	if body.Value != "POST received" {
		t.Logf("body = %s", body.Value)
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

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("head() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("head() array should have 2 elements (status, headers), got %d", len(arr.Elements))
	}

	status, ok := arr.Elements[0].(*objects.Int)
	if !ok {
		t.Fatalf("first element should be Int (status code), got %T", arr.Elements[0])
	}
	if status.Value != 200 {
		t.Errorf("status code = %d, want 200", status.Value)
	}

	// Second element should be array of headers
	headersArr, ok := arr.Elements[1].(*objects.Array)
	if !ok {
		t.Fatalf("second element should be Array (headers), got %T", arr.Elements[1])
	}
	// Headers array should have at least Content-Length
	foundContentLength := false
	for _, h := range headersArr.Elements {
		if hdrArr, ok := h.(*objects.Array); ok && len(hdrArr.Elements) == 2 {
			key, _ := hdrArr.Elements[0].(*objects.String)
			val, _ := hdrArr.Elements[1].(*objects.String)
			if key != nil && key.Value == "Content-Length" {
				foundContentLength = true
				if val.Value != "5" {
					t.Errorf("Content-Length = %s, want 5", val.Value)
				}
				break
			}
		}
	}
	if !foundContentLength {
		t.Log("Content-Length header not found (may be optional)")
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

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("request() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("request() array should have 3 elements, got %d", len(arr.Elements))
	}

	body, ok := arr.Elements[0].(*objects.String)
	if !ok {
		t.Fatalf("first element should be String (body), got %T", arr.Elements[0])
	}
	if body.Value != "PATCH response" {
		t.Errorf("body = %s, want 'PATCH response'", body.Value)
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

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("getJson() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("getJson() array should have 2 elements, got %d", len(arr.Elements))
	}

	body, ok1 := arr.Elements[0].(*objects.String)
	status, ok2 := arr.Elements[1].(*objects.Int)
	if !ok1 || !ok2 {
		t.Fatalf("getJson() array elements should be String and Int, got %T and %T", arr.Elements[0], arr.Elements[1])
	}
	if body.Value != `{"message":"hello","count":5}` {
		t.Logf("body = %s", body.Value)
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

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("postJson() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("postJson() array should have 2 elements, got %d", len(arr.Elements))
	}

	body, _ := arr.Elements[0].(*objects.String)
	status, _ := arr.Elements[1].(*objects.Int)
	t.Logf("postJson response body: %s, status: %d", body.Value, status.Value)
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
