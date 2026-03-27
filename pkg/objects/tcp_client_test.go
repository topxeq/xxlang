// pkg/objects/tcp_client_test.go
package objects

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTcpClientBasic(t *testing.T) {
	client := NewTcpClient()
	if client == nil {
		t.Fatal("NewTcpClient returned nil")
	}

	if client.Type() != TcpClientType {
		t.Errorf("expected type %s, got %s", TcpClientType, client.Type())
	}

	if client.TypeTag() != TagTcpClient {
		t.Errorf("expected type tag %d, got %d", TagTcpClient, client.TypeTag())
	}

	if client.IsClosed() {
		t.Error("new client should not be closed")
	}
}

func TestTcpClientConnectError(t *testing.T) {
	client := NewTcpClient()

	// Connect to invalid address should return error
	err := client.Connect("invalid:address:12345")
	if _, ok := err.(*Error); !ok {
		t.Error("Connect to invalid address should return Error")
	}
}

func TestTcpClientSendError(t *testing.T) {
	client := NewTcpClient()

	// Send on unconnected client should return error
	err := client.SendString("hello")
	if _, ok := err.(*Error); !ok {
		t.Error("Send on unconnected client should return Error")
	}
}

func TestTcpClientReceiveError(t *testing.T) {
	client := NewTcpClient()

	// Receive on unconnected client should return error
	result := client.Receive(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("Receive on unconnected client should return Error")
	}
}

func TestTcpClientClose(t *testing.T) {
	client := NewTcpClient()

	// Close on unconnected client should succeed
	result := client.Close()
	if _, ok := result.(*Error); ok {
		t.Errorf("Close on unconnected client should succeed: %v", result)
	}

	if !client.IsClosed() {
		t.Error("client should be marked as closed")
	}
}

func TestTcpClientIntegration(t *testing.T) {
	// Start a simple echo server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	// Get the actual port
	addr := listener.Addr().(*net.TCPAddr)
	serverAddr := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	// Handle one connection in goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo back
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		conn.Write(buf[:n])
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Client connect and communicate
	client := NewTcpClient()
	result := client.Connect(serverAddr)
	if _, ok := result.(*Error); ok {
		t.Fatalf("Connect failed: %v", result)
	}
	defer client.Close()

	result = client.SendString("hello")
	if _, ok := result.(*Error); ok {
		t.Fatalf("SendString failed: %v", result)
	}

	result = client.Receive(1024)
	s, ok := result.(*String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", s.Value)
	}
}