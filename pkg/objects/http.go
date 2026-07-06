// pkg/objects/http.go
// HTTP object types for server mode.
package objects

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

// WebSocketMessageType constants
const (
	WebSocketTextMessage   = 1
	WebSocketBinaryMessage = 2
	WebSocketCloseMessage  = 8
	WebSocketPingMessage   = 9
	WebSocketPongMessage   = 10
)

// HttpReq wraps http.Request for script access.
// It provides methods and member access for reading HTTP request data.
type HttpReq struct {
	// Value is the underlying Go http.Request
	Value *http.Request
	// Members holds dynamically accessed members for script use
	Members map[string]Object
}

// Type returns the object type.
func (r *HttpReq) Type() ObjectType { return HttpReqType }

// TypeTag returns the type tag for fast type checking.
func (r *HttpReq) TypeTag() TypeTag { return TagHttpReq }

// Inspect returns a string representation of the HTTP request.
func (r *HttpReq) Inspect() string {
	if r.Value == nil {
		return "[http_request nil]"
	}
	return fmt.Sprintf("[http_request %s %s]", r.Value.Method, r.Value.URL.Path)
}

// ToBool converts the HTTP request to a boolean (always true).
func (r *HttpReq) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the HTTP request.
// HTTP requests are not hashable in a meaningful way.
func (r *HttpReq) HashKey() HashKey {
	return HashKey{Type: HttpReqType, Value: 0}
}

// GetMember returns a member by name for script access.
func (r *HttpReq) GetMember(name string) Object {
	if r.Members != nil {
		if obj, ok := r.Members[name]; ok {
			return obj
		}
	}

	if r.Value == nil {
		return NULL
	}

	switch name {
	case "method":
		return NewString(r.Value.Method)
	case "url":
		return NewString(r.Value.URL.String())
	case "path":
		return NewString(r.Value.URL.Path)
	case "host":
		return NewString(r.Value.Host)
	case "remoteAddr":
		return NewString(r.Value.RemoteAddr)
	case "proto":
		return NewString(r.Value.Proto)
	case "contentLength":
		return NewInt(r.Value.ContentLength)
	case "header":
		return r.getHeaders()
	case "body":
		return r.getBody()
	case "files":
		return r.getFiles()
	case "getReader":
		// Returns a builtin function that creates a Reader from the request body
		return &Builtin{Fn: func(args ...Object) Object {
			if r.Value.Body == nil {
				return newError("request body is nil")
			}
			return NewReader(r.Value.Body)
		}}
	case "readBody":
		// readBody(maxBytes) reads the request body as a string, optionally
		// capped at maxBytes. Pass 0 or a negative value for no limit.
		// Unlike `body` (which always reads the whole thing), this lets the
		// script bound memory use on untrusted requests.
		return &Builtin{Fn: func(args ...Object) Object {
			var maxBytes int64 = -1
			if len(args) >= 1 {
				if n, ok := args[0].(*Int); ok {
					maxBytes = n.Value
				} else {
					return newError("argument to 'readBody' must be INT, got %s", args[0].Type())
				}
			}
			return r.readBody(maxBytes)
		}}
	}

	return NULL
}

// getHeaders converts request headers to a Map object.
func (r *HttpReq) getHeaders() Object {
	pairs := make(map[HashKey]MapPair)
	for key, values := range r.Value.Header {
		k := NewString(key)
		var v Object
		if len(values) == 1 {
			v = NewString(values[0])
		} else {
			elements := make([]Object, len(values))
			for i, val := range values {
				elements[i] = NewString(val)
			}
			v = NewArray(elements)
		}
		pairs[k.HashKey()] = MapPair{Key: k, Value: v}
	}
	return NewMap(pairs)
}

// getBody reads the request body as a string with no size limit.
// Note: This consumes the body and it cannot be read again.
// For untrusted requests, prefer readBody(maxBytes) to bound memory use.
func (r *HttpReq) getBody() Object {
	if r.Value.Body == nil {
		return NewString("")
	}
	body, err := io.ReadAll(r.Value.Body)
	if err != nil {
		return NewString("")
	}
	return NewString(string(body))
}

// readBody reads the request body as a string, optionally capped at maxBytes.
// maxBytes <= 0 means no limit. This is the script-facing, user-controllable
// counterpart to getBody — the language does not impose a hard cap so that
// scripts dealing with large legitimate bodies are not silently truncated;
// scripts handling untrusted input should call readBody with an explicit cap.
func (r *HttpReq) readBody(maxBytes int64) Object {
	if r.Value.Body == nil {
		return NewString("")
	}
	var reader io.Reader = r.Value.Body
	if maxBytes > 0 {
		reader = io.LimitReader(reader, maxBytes)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return NewString("")
	}
	return NewString(string(body))
}

// getFiles retrieves uploaded files from the request.
// Returns a map of field names to arrays of FileUpload objects.
func (r *HttpReq) getFiles() Object {
	if r.Value == nil {
		return NewMap(make(map[HashKey]MapPair))
	}

	// Parse multipart form (max 32MB in memory)
	if err := r.Value.ParseMultipartForm(32 << 20); err != nil {
		// Not a multipart form or error, return empty map
		return NewMap(make(map[HashKey]MapPair))
	}

	// Build result map
	pairs := make(map[HashKey]MapPair)
	for key, fileHeaders := range r.Value.MultipartForm.File {
		k := NewString(key)
		uploads := make([]Object, len(fileHeaders))
		for i, fh := range fileHeaders {
			uploads[i] = NewFileUpload(fh)
		}
		pairs[k.HashKey()] = MapPair{Key: k, Value: NewArray(uploads)}
	}

	return NewMap(pairs)
}

// HttpResp wraps http.ResponseWriter for script access.
// It provides methods for writing responses and setting headers.
type HttpResp struct {
	// Value is the underlying Go http.ResponseWriter
	Value http.ResponseWriter
	// Members holds dynamically accessed members for script use
	Members map[string]Object
	// written tracks if the response has been written to
	written bool
}

// Type returns the object type.
func (r *HttpResp) Type() ObjectType { return HttpRespType }

// TypeTag returns the type tag for fast type checking.
func (r *HttpResp) TypeTag() TypeTag { return TagHttpResp }

// Inspect returns a string representation of the HTTP response.
func (r *HttpResp) Inspect() string {
	return "[http_response]"
}

// ToBool converts the HTTP response to a boolean (always true).
func (r *HttpResp) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the HTTP response.
// HTTP responses are not hashable in a meaningful way.
func (r *HttpResp) HashKey() HashKey {
	return HashKey{Type: HttpRespType, Value: 0}
}

// Written returns whether the response has been written to.
func (r *HttpResp) Written() bool {
	return r.written
}

// SetWritten marks the response as written.
func (r *HttpResp) SetWritten() {
	r.written = true
}

// GetMember returns a member by name for script access.
func (r *HttpResp) GetMember(name string) Object {
	if r.Members != nil {
		if obj, ok := r.Members[name]; ok {
			return obj
		}
	}

	if r.Value == nil {
		return NULL
	}

	switch name {
	case "getWriter":
		// Returns a builtin function that creates a Writer from the response
		return &Builtin{Fn: func(args ...Object) Object {
			return NewWriter(r.Value)
		}}
	case "written":
		return &Bool{Value: r.written}
	}

	return NULL
}

// HttpMux wraps http.ServeMux for route registration.
// It allows scripts to register custom route handlers.
type HttpMux struct {
	// Value is the underlying Go http.ServeMux
	Value *http.ServeMux
}

// Type returns the object type.
func (m *HttpMux) Type() ObjectType { return HttpMuxType }

// TypeTag returns the type tag for fast type checking.
func (m *HttpMux) TypeTag() TypeTag { return TagHttpMux }

// Inspect returns a string representation of the HTTP mux.
func (m *HttpMux) Inspect() string {
	return "[http_mux]"
}

// ToBool converts the HTTP mux to a boolean (always true).
func (m *HttpMux) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the HTTP mux.
// HTTP muxes are not hashable in a meaningful way.
func (m *HttpMux) HashKey() HashKey {
	return HashKey{Type: HttpMuxType, Value: 0}
}

// NewHttpReq creates a new HttpReq object from an http.Request.
func NewHttpReq(req *http.Request) *HttpReq {
	return &HttpReq{
		Value:   req,
		Members: make(map[string]Object),
	}
}

// NewHttpResp creates a new HttpResp object from an http.ResponseWriter.
func NewHttpResp(res http.ResponseWriter) *HttpResp {
	return &HttpResp{
		Value:   res,
		Members: make(map[string]Object),
	}
}

// NewHttpMux creates a new HttpMux object.
func NewHttpMux() *HttpMux {
	return &HttpMux{
		Value: http.NewServeMux(),
	}
}

// WebSocket represents a WebSocket connection.
// It wraps the connection and provides methods for reading and writing messages.
type WebSocket struct {
	// Conn is the underlying WebSocket connection interface
	Conn WebSocketConn
	// mu protects concurrent access to the connection
	mu sync.Mutex
	// closed indicates if the connection has been closed
	closed bool
}

// WebSocketConn defines the interface for WebSocket connections.
// This allows different implementations (gorilla, standard library, etc.)
type WebSocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Type returns the object type.
func (ws *WebSocket) Type() ObjectType { return WebSocketType }

// TypeTag returns the type tag for fast type checking.
func (ws *WebSocket) TypeTag() TypeTag { return TagWebSocket }

// Inspect returns a string representation of the WebSocket.
func (ws *WebSocket) Inspect() string {
	return "[websocket]"
}

// ToBool converts the WebSocket to a boolean (always true).
func (ws *WebSocket) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the WebSocket.
// WebSockets are not hashable in a meaningful way.
func (ws *WebSocket) HashKey() HashKey {
	return HashKey{Type: WebSocketType, Value: 0}
}

// ReadMessage reads a message from the WebSocket connection.
// Returns an array: [messageType, data]
func (ws *WebSocket) ReadMessage() Object {
	if ws.Conn == nil || ws.closed {
		return &Error{Message: "websocket connection is closed"}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	messageType, data, err := ws.Conn.ReadMessage()
	if err != nil {
		return &Error{Message: fmt.Sprintf("read message failed: %v", err)}
	}

	// Convert data to string (for text) or array of bytes
	var dataObj Object
	if messageType == WebSocketTextMessage {
		dataObj = NewString(string(data))
	} else {
		// For binary/close messages, return as string (raw bytes)
		dataObj = NewString(string(data))
	}

	// Return array [messageType, data]
	elements := []Object{
		NewInt(int64(messageType)),
		dataObj,
	}
	return NewArray(elements)
}

// SendTextMessage sends a text message over the WebSocket.
func (ws *WebSocket) SendTextMessage(text string) Object {
	if ws.Conn == nil || ws.closed {
		return &Error{Message: "websocket connection is closed"}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	err := ws.Conn.WriteMessage(WebSocketTextMessage, []byte(text))
	if err != nil {
		return &Error{Message: fmt.Sprintf("send text message failed: %v", err)}
	}
	return NULL
}

// SendBinaryMessage sends a binary message over the WebSocket.
func (ws *WebSocket) SendBinaryMessage(data string) Object {
	if ws.Conn == nil || ws.closed {
		return &Error{Message: "websocket connection is closed"}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	err := ws.Conn.WriteMessage(WebSocketBinaryMessage, []byte(data))
	if err != nil {
		return &Error{Message: fmt.Sprintf("send binary message failed: %v", err)}
	}
	return NULL
}

// SendCloseMessage sends a close message over the WebSocket.
func (ws *WebSocket) SendCloseMessage() Object {
	if ws.Conn == nil || ws.closed {
		return &Error{Message: "websocket connection is closed"}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	err := ws.Conn.WriteMessage(WebSocketCloseMessage, []byte{})
	if err != nil {
		return &Error{Message: fmt.Sprintf("send close message failed: %v", err)}
	}
	return NULL
}

// Close closes the WebSocket connection.
func (ws *WebSocket) Close() Object {
	if ws.Conn == nil || ws.closed {
		return NULL
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.closed = true
	if ws.Conn != nil {
		ws.Conn.Close()
	}
	return NULL
}

// IsClosed returns whether the WebSocket is closed.
func (ws *WebSocket) IsClosed() bool {
	return ws.closed
}

// NewWebSocket creates a new WebSocket object.
func NewWebSocket(conn WebSocketConn) *WebSocket {
	return &WebSocket{
		Conn:   conn,
		closed: false,
	}
}
