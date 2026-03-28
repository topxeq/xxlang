// pkg/objects/websocket_test.go
package objects

import (
	"testing"
)

// mockWebSocketConn is a mock implementation of WebSocketConn for testing
type mockWebSocketConn struct{}

func (m *mockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return 1, []byte("test message"), nil
}

func (m *mockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (m *mockWebSocketConn) Close() error {
	return nil
}

func TestWebSocketType(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := NewWebSocket(conn)

	if ws.Type() != WebSocketType {
		t.Errorf("expected WebSocketType, got %s", ws.Type())
	}

	if ws.TypeTag() != TagWebSocket {
		t.Errorf("expected TagWebSocket, got %d", ws.TypeTag())
	}

	if ws.Inspect() != "[websocket]" {
		t.Errorf("expected [websocket], got %s", ws.Inspect())
	}

	if !ws.ToBool().Value {
		t.Error("expected WebSocket to be truthy")
	}
}

func TestWebSocketMethods(t *testing.T) {
	// Test that WebSocket methods are registered
	methods, ok := TypeMethods[WebSocketType]
	if !ok {
		t.Fatal("WebSocket methods not registered")
	}

	expectedMethods := []string{
		"typeOf", "toStr", "readMsg", "sendTextMsg",
		"sendBinaryMsg", "sendCloseMsg", "close", "isClosed",
	}

	for _, method := range expectedMethods {
		if _, ok := methods[method]; !ok {
			t.Errorf("expected method '%s' to be registered", method)
		}
	}
}

func TestWebSocketGetMethod(t *testing.T) {
	// Test GetMethod for WebSocket
	method, ok := GetMethod(WebSocketType, "readMsg")
	if !ok {
		t.Error("expected readMsg method to be found")
	}
	if method == nil {
		t.Error("expected readMsg method to be non-nil")
	}

	// Test non-existent method
	_, ok = GetMethod(WebSocketType, "nonExistentMethod")
	if ok {
		t.Error("expected non-existent method to not be found")
	}
}

func TestWebSocketReadMessage(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := NewWebSocket(conn)

	result := ws.ReadMessage()
	if isErr(result) {
		t.Errorf("unexpected error: %s", result.Inspect())
	}

	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %s", result.Type())
	}

	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}

	msgType, ok := arr.Elements[0].(*Int)
	if !ok {
		t.Fatalf("expected Int, got %s", arr.Elements[0].Type())
	}
	if msgType.Value != 1 {
		t.Errorf("expected message type 1, got %d", msgType.Value)
	}

	msgData, ok := arr.Elements[1].(*String)
	if !ok {
		t.Fatalf("expected String, got %s", arr.Elements[1].Type())
	}
	if msgData.Value != "test message" {
		t.Errorf("expected 'test message', got '%s'", msgData.Value)
	}
}

func TestWebSocketSendTextMessage(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := NewWebSocket(conn)

	result := ws.SendTextMessage("hello")
	if isErr(result) {
		t.Errorf("unexpected error: %s", result.Inspect())
	}
}

func TestWebSocketClose(t *testing.T) {
	conn := &mockWebSocketConn{}
	ws := NewWebSocket(conn)

	if ws.IsClosed() {
		t.Error("expected WebSocket to not be closed initially")
	}

	result := ws.Close()
	if isErr(result) {
		t.Errorf("unexpected error: %s", result.Inspect())
	}

	if !ws.IsClosed() {
		t.Error("expected WebSocket to be closed after Close()")
	}
}

func isErr(obj Object) bool {
	_, ok := obj.(*Error)
	return ok
}
