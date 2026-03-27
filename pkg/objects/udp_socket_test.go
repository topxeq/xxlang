// pkg/objects/udp_socket_test.go
package objects

import (
	"testing"
	"time"
)

func TestUdpSocketBasic(t *testing.T) {
	sock := NewUdpSocket()
	if sock == nil {
		t.Fatal("NewUdpSocket returned nil")
	}

	if sock.Type() != UdpSocketType {
		t.Errorf("expected type %s, got %s", UdpSocketType, sock.Type())
	}

	if sock.TypeTag() != TagUdpSocket {
		t.Errorf("expected type tag %d, got %d", TagUdpSocket, sock.TypeTag())
	}
}

func TestUdpSocketSendToError(t *testing.T) {
	sock := NewUdpSocket()

	// Send to invalid address should return error
	result := sock.SendTo("hello", "invalid:address:format")
	if _, ok := result.(*Error); !ok {
		t.Error("SendTo invalid address should return Error")
	}
}

func TestUdpSocketClose(t *testing.T) {
	sock := NewUdpSocket()

	// Close on non-bound socket should succeed
	result := sock.Close()
	if _, ok := result.(*Error); ok {
		t.Errorf("Close on non-bound socket should succeed: %v", result)
	}
}

func TestUdpSocketBindAndClose(t *testing.T) {
	sock := NewUdpSocket()

	result := sock.Bind(":0") // Random port
	if _, ok := result.(*Error); ok {
		t.Fatalf("Bind failed: %v", result)
	}

	if sock.IsClosed() {
		t.Error("bound socket should not be closed")
	}

	addr := sock.LocalAddr()
	if addr == nil {
		t.Fatal("LocalAddr should return SocketAddr")
	}

	result = sock.Close()
	if _, ok := result.(*Error); ok {
		t.Fatalf("Close failed: %v", result)
	}

	if !sock.IsClosed() {
		t.Error("closed socket should be marked as closed")
	}
}

func TestUdpSocketSendReceive(t *testing.T) {
	// Server socket
	server := NewUdpSocket()
	result := server.Bind(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Bind failed: %v", result)
	}
	defer server.Close()

	serverAddr := server.LocalAddr()

	// Client socket
	client := NewUdpSocket()
	defer client.Close()

	// Send message
	result = client.SendTo("hello", serverAddr.(*SocketAddr).ToStr())
	if _, ok := result.(*Error); ok {
		t.Fatalf("SendTo failed: %v", result)
	}

	// Receive message
	server.SetTimeout(1000) // 1 second
	data, addr := server.ReceiveFrom(1024)
	if data == nil {
		t.Fatal("ReceiveFrom should return data")
	}

	s, ok := data.(*String)
	if !ok {
		t.Fatalf("Expected String, got %T", data)
	}
	if s.Value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", s.Value)
	}

	if addr == nil {
		t.Fatal("ReceiveFrom should return address")
	}
}

func TestUdpSocketReceiveTimeout(t *testing.T) {
	sock := NewUdpSocket()
	result := sock.Bind(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Bind failed: %v", result)
	}
	defer sock.Close()

	sock.SetTimeout(100) // 100ms

	start := time.Now()
	data, _ := sock.ReceiveFrom(1024)
	elapsed := time.Since(start)

	// Check if data is NULL (the null object, not nil)
	if data != NULL {
		t.Error("ReceiveFrom should return null on timeout")
	}

	if elapsed < 80*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("ReceiveFrom should timeout around 100ms, took %v", elapsed)
	}
}

func TestUdpSocketSendReceiveBytes(t *testing.T) {
	// Server socket
	server := NewUdpSocket()
	result := server.Bind(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Bind failed: %v", result)
	}
	defer server.Close()

	serverAddr := server.LocalAddr()

	// Client socket
	client := NewUdpSocket()
	defer client.Close()

	// Send bytes
	result = client.SendToBytes([]byte{0x01, 0x02, 0x03}, serverAddr.(*SocketAddr).ToStr())
	if _, ok := result.(*Error); ok {
		t.Fatalf("SendToBytes failed: %v", result)
	}

	// Receive bytes
	server.SetTimeout(1000) // 1 second
	data, addr := server.ReceiveFromBytes(1024)
	if data == nil {
		t.Fatal("ReceiveFromBytes should return data")
	}

	arr, ok := data.(*Array)
	if !ok {
		t.Fatalf("Expected Array, got %T", data)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("Expected 3 bytes, got %d", len(arr.Elements))
	}

	if addr == nil {
		t.Fatal("ReceiveFromBytes should return address")
	}
}

func TestUdpSocketReceiveWithoutBind(t *testing.T) {
	sock := NewUdpSocket()
	defer sock.Close()

	sock.SetTimeout(100) // 100ms

	// ReceiveFrom without bind should return error
	data, _ := sock.ReceiveFrom(1024)
	if _, ok := data.(*Error); !ok {
		t.Error("ReceiveFrom without bind should return Error")
	}
}

func TestUdpSocketIsClosed(t *testing.T) {
	sock := NewUdpSocket()

	if sock.IsClosed() {
		t.Error("new socket should not be closed")
	}

	sock.Close()

	if !sock.IsClosed() {
		t.Error("closed socket should be marked as closed")
	}
}

func TestUdpSocketLocalAddrWithoutBind(t *testing.T) {
	sock := NewUdpSocket()

	addr := sock.LocalAddr()
	if _, ok := addr.(*Error); !ok {
		t.Error("LocalAddr without bind should return Error")
	}
}

func TestUdpSocketDoubleClose(t *testing.T) {
	sock := NewUdpSocket()
	sock.Bind(":0")

	// First close
	result := sock.Close()
	if _, ok := result.(*Error); ok {
		t.Errorf("First close should succeed: %v", result)
	}

	// Second close should also succeed (no-op)
	result = sock.Close()
	if _, ok := result.(*Error); ok {
		t.Errorf("Second close should succeed: %v", result)
	}
}