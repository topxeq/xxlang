// pkg/stdlib/websocket_extra_test.go
// Additional tests for WebSocket module to increase coverage.
package stdlib

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callWebSocketFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("websocket")
	if mod == nil {
		panic("websocket module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// mockWebSocketConn is a minimal implementation of WebSocketConn for testing.
type mockWebSocketConn struct {
	readMsgCalled bool
	sendCalled    bool
	closeCalled   bool
	shouldErr     bool
	msgType       int
	data          []byte
}

func (m *mockWebSocketConn) ReadMessage() (int, []byte, error) {
	m.readMsgCalled = true
	if m.shouldErr {
		return 0, nil, fmt.Errorf("read error")
	}
	return m.msgType, m.data, nil
}

func (m *mockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	m.sendCalled = true
	if m.shouldErr {
		return fmt.Errorf("write error")
	}
	m.msgType = messageType
	m.data = data
	return nil
}

func (m *mockWebSocketConn) Close() error {
	m.closeCalled = true
	return nil
}

// TestWebSocket_ModuleExistence tests that the websocket module is registered.
func TestWebSocket_ModuleExistence(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Fatal("websocket module not found")
	}
	if mod.Name != "websocket" {
		t.Errorf("expected module name 'websocket', got %s", mod.Name)
	}
	expectedExports := []string{"upgrade", "readMsg", "sendText", "sendBinary", "sendClose", "close", "isWebSocket"}
	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in websocket module", name)
		}
	}
}

// TestWebSocket_IsWebSocket tests the isWebSocket function.
func TestWebSocket_IsWebSocket(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}
	fn, ok := mod.Exports["isWebSocket"].(*objects.Builtin)
	if !ok {
		t.Fatal("isWebSocket not found or not builtin")
	}
	// Create a WebSocket object (using a mock conn)
	conn := &mockWebSocketConn{}
	wsObj := objects.NewWebSocket(conn)
	res := fn.Fn(wsObj)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isWebSocket should return true for WebSocket, got %T %v", res, res)
	}
	// Test with non-WebSocket
	res2 := fn.Fn(String("not a websocket"))
	if b, ok := res2.(*objects.Bool); !ok || b.Value {
		t.Fatalf("isWebSocket should return false for non-WebSocket, got %T %v", res2, res2)
	}
}

// TestWebSocket_ReadMsg_ArgumentValidation tests readMsg argument validation.
func TestWebSocket_ReadMsg_ArgumentValidation(t *testing.T) {
	// Too many args
	result := callWebSocketFunc("readMsg", String("extra"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("readMsg with too many args should error")
	}
}

// TestWebSocket_SendText_ArgumentValidation tests sendText argument validation.
func TestWebSocket_SendText_ArgumentValidation(t *testing.T) {
	// Missing arg
	result := callWebSocketFunc("sendText")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendText missing args should error")
	}
	// Too few
	result = callWebSocketFunc("sendText", String("ws"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendText with 1 arg should error")
	}
	// First arg not WebSocket
	result = callWebSocketFunc("sendText", Int(123), String("text"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendText with non-WebSocket first arg should error")
	}
	// Second arg not string
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result = callWebSocketFunc("sendText", ws, Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendText with non-string second arg should error")
	}
}

// TestWebSocket_SendBinary_ArgumentValidation tests sendBinary argument validation.
func TestWebSocket_SendBinary_ArgumentValidation(t *testing.T) {
	result := callWebSocketFunc("sendBinary")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendBinary missing args should error")
	}
	result = callWebSocketFunc("sendBinary", String("ws"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendBinary with 1 arg should error")
	}
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result = callWebSocketFunc("sendBinary", ws, Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendBinary with non-string second arg should error")
	}
}

// TestWebSocket_SendClose_ArgumentValidation tests sendClose argument validation.
func TestWebSocket_SendClose_ArgumentValidation(t *testing.T) {
	result := callWebSocketFunc("sendClose")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendClose missing arg should error")
	}
	result = callWebSocketFunc("sendClose", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendClose with non-WebSocket arg should error")
	}
}

// TestWebSocket_Close_ArgumentValidation tests close argument validation.
func TestWebSocket_Close_ArgumentValidation(t *testing.T) {
	result := callWebSocketFunc("close")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("close missing arg should error")
	}
	result = callWebSocketFunc("close", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("close with non-WebSocket arg should error")
	}
}

// TestWebSocket_ReadMsg_WithNilConn tests readMsg when WebSocket has nil connection.
func TestWebSocket_ReadMsg_WithNilConn(t *testing.T) {
	ws := objects.NewWebSocket(nil)
	result := callWebSocketFunc("readMsg", ws)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("readMsg with nil conn should return error")
	}
}

// TestWebSocket_SendText_WithClosed tests sendText after WebSocket is closed.
func TestWebSocket_SendText_WithClosed(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	// Close the WebSocket
	_ = callWebSocketFunc("close", ws)
	// Now sendText should fail because closed
	result := callWebSocketFunc("sendText", ws, String("hello"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendText on closed WebSocket should error")
	}
}

// TestWebSocket_SendBinary_WithClosed tests sendBinary after close.
func TestWebSocket_SendBinary_WithClosed(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	_ = callWebSocketFunc("close", ws)
	result := callWebSocketFunc("sendBinary", ws, String("data"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendBinary on closed WebSocket should error")
	}
}

// TestWebSocket_SendClose_WithClosed tests sendClose after close.
func TestWebSocket_SendClose_WithClosed(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	_ = callWebSocketFunc("close", ws)
	result := callWebSocketFunc("sendClose", ws)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sendClose on closed WebSocket should error")
	}
}

// TestWebSocket_ReadMsg_Success tests readMsg success path with mock.
func TestWebSocket_ReadMsg_Success(t *testing.T) {
	conn := &mockWebSocketConn{
		msgType: 1, // text message
		data:    []byte("test message"),
	}
	ws := objects.NewWebSocket(conn)
	result := callWebSocketFunc("readMsg", ws)
	if result == nil {
		t.Fatal("readMsg returned nil")
	}
	// Should return an array [messageType, data]
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	// Check message type
	mt, ok := arr.Elements[0].(*objects.Int)
	if !ok || mt.Value != 1 {
		t.Errorf("expected message type 1, got %v", arr.Elements[0])
	}
	// Check data
	msg, ok := arr.Elements[1].(*objects.String)
	if !ok || msg.Value != "test message" {
		t.Errorf("expected message 'test message', got %v", arr.Elements[1])
	}
}

// TestWebSocket_SendText_Success tests sendText success.
func TestWebSocket_SendText_Success(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result := callWebSocketFunc("sendText", ws, String("hello"))
	if result != objects.NULL {
		t.Errorf("sendText should return NULL, got %v", result)
	}
	if !conn.sendCalled {
		t.Error("sendText should have called conn.WriteMessage")
	}
	// Check that it sent as text message (type 1)
	if conn.msgType != 1 {
		t.Errorf("expected message type 1, got %d", conn.msgType)
	}
	if string(conn.data) != "hello" {
		t.Errorf("expected data 'hello', got %s", string(conn.data))
	}
}

// TestWebSocket_SendBinary_Success tests sendBinary success.
func TestWebSocket_SendBinary_Success(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result := callWebSocketFunc("sendBinary", ws, String("binary"))
	if result != objects.NULL {
		t.Errorf("sendBinary should return NULL, got %v", result)
	}
	if !conn.sendCalled {
		t.Error("sendBinary should have called conn.WriteMessage")
	}
	// Binary message type is 2
	if conn.msgType != 2 {
		t.Errorf("expected message type 2, got %d", conn.msgType)
	}
	if string(conn.data) != "binary" {
		t.Errorf("expected data 'binary', got %s", string(conn.data))
	}
}

// TestWebSocket_SendClose_Success tests sendClose success.
func TestWebSocket_SendClose_Success(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result := callWebSocketFunc("sendClose", ws)
	if result != objects.NULL {
		t.Errorf("sendClose should return NULL, got %v", result)
	}
	if !conn.sendCalled {
		t.Error("sendClose should have called conn.WriteMessage")
	}
	// Close message type is 8
	if conn.msgType != 8 {
		t.Errorf("expected message type 8, got %d", conn.msgType)
	}
}

// TestWebSocket_Close_Success tests close success.
func TestWebSocket_Close_Success(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := objects.NewWebSocket(conn)
	result := callWebSocketFunc("close", ws)
	if result != objects.NULL {
		t.Errorf("close should return NULL, got %v", result)
	}
	if !conn.closeCalled {
		t.Error("close should have called conn.Close")
	}
	if !ws.IsClosed() {
		t.Error("WebSocket should be closed after close()")
	}
}
