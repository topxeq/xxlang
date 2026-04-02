// pkg/objects/http_test.go
// Tests for HTTP object types (HttpReq, HttpResp)
package objects

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewHttpReq(t *testing.T) {
	req := &http.Request{}
	httpReq := NewHttpReq(req)

	if httpReq.Type() != HttpReqType {
		t.Errorf("expected type HTTP_REQ, got %s", httpReq.Type())
	}

	if httpReq.TypeTag() != TagHttpReq {
		t.Errorf("expected TypeTag TagHttpReq, got %d", httpReq.TypeTag())
	}

	if httpReq.Value != req {
		t.Errorf("expected Value to be the request")
	}
}

func TestHttpReq_Inspect(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	httpReq := NewHttpReq(req)

	inspect := httpReq.Inspect()
	if inspect == "" {
		t.Errorf("expected non-empty Inspect")
	}

	if !strings.Contains(inspect, "[http_request") {
		t.Errorf("expected Inspect to contain '[http_request', got %s", inspect)
	}

	if !strings.Contains(inspect, "GET") {
		t.Errorf("expected Inspect to contain method 'GET'")
	}
}

func TestHttpReq_ToBool(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	httpReq := NewHttpReq(req)

	if !httpReq.ToBool().Value {
		t.Errorf("HttpReq.ToBool() should return true")
	}
}

func TestHttpReq_HashKey(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	httpReq := NewHttpReq(req)

	key := httpReq.HashKey()
	if key.Type != HttpReqType {
		t.Errorf("expected HashKey.Type HTTP_REQ, got %s", key.Type)
	}
}

func TestHttpReq_GetMember(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/data?foo=bar", nil)
	req.Header.Set("Content-Type", "application/json")
	httpReq := NewHttpReq(req)

	// Test method member
	method := httpReq.GetMember("method")
	if method.Type() != StringType || method.(*String).Value != "POST" {
		t.Errorf("expected method 'POST', got %v", method)
	}

	// Test path member
	path := httpReq.GetMember("path")
	if path.Type() != StringType || path.(*String).Value != "/api/data" {
		t.Errorf("expected path '/api/data', got %v", path)
	}

	// Test host member
	req.Host = "example.com"
	host := httpReq.GetMember("host")
	if host.Type() != StringType || host.(*String).Value != "example.com" {
		t.Errorf("expected host 'example.com', got %v", host)
	}

	// Test header member
	header := httpReq.GetMember("header")
	if header.Type() != MapType {
		t.Errorf("expected header to be Map, got %s", header.Type())
	}

	// Test body member
	body := httpReq.GetMember("body")
	if body.Type() != StringType {
		t.Errorf("expected body to be String, got %s", body.Type())
	}

	// Test nonexistent member
	none := httpReq.GetMember("nonexistent")
	if none != NULL {
		t.Errorf("expected NULL for nonexistent member, got %v", none)
	}
}

func TestHttpResp_Basic(t *testing.T) {
	// Test basic HttpResp properties
	resp := &HttpResp{
		Value:   nil,
		Members: make(map[string]Object),
		written: false,
	}

	if resp.Type() != HttpRespType {
		t.Errorf("expected type HTTP_RESP, got %s", resp.Type())
	}

	if resp.TypeTag() != TagHttpResp {
		t.Errorf("expected TypeTag TagHttpResp, got %d", resp.TypeTag())
	}

	if resp.Inspect() != "[http_response]" {
		t.Errorf("expected Inspect '[http_response]', got %s", resp.Inspect())
	}

	if !resp.ToBool().Value {
		t.Errorf("ToBool() should return true")
	}

	key := resp.HashKey()
	if key.Type != HttpRespType {
		t.Errorf("expected HashKey.Type HTTP_RESP, got %s", key.Type)
	}
}
