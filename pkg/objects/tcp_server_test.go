// pkg/objects/tcp_server_test.go
package objects

import (
	"testing"
	"time"
)

func TestTcpServerBasic(t *testing.T) {
	server := NewTcpServer()
	if server == nil {
		t.Fatal("NewTcpServer returned nil")
	}

	if server.Type() != TcpServerType {
		t.Errorf("expected type %s, got %s", TcpServerType, server.Type())
	}

	if server.TypeTag() != TagTcpServer {
		t.Errorf("expected type tag %d, got %d", TagTcpServer, server.TypeTag())
	}
}

func TestTcpServerListenError(t *testing.T) {
	server := NewTcpServer()

	// Invalid address should return error
	result := server.Listen("invalid:address:format")
	if _, ok := result.(*Error); !ok {
		t.Error("Listen on invalid address should return Error")
	}
}

func TestTcpServerIsClosed(t *testing.T) {
	server := NewTcpServer()

	if !server.IsClosed() {
		t.Error("new server should be closed (not listening)")
	}
}

func TestTcpServerClose(t *testing.T) {
	server := NewTcpServer()

	// Close on non-listening server should succeed
	result := server.Close()
	if _, ok := result.(*Error); ok {
		t.Errorf("Close on non-listening server should succeed: %v", result)
	}
}

func TestTcpServerListenAndClose(t *testing.T) {
	server := NewTcpServer()

	result := server.Listen(":0") // Random port
	if _, ok := result.(*Error); ok {
		t.Fatalf("Listen failed: %v", result)
	}

	if server.IsClosed() {
		t.Error("listening server should not be closed")
	}

	addr := server.Addr()
	if addr == nil {
		t.Fatal("Addr should return SocketAddr")
	}

	if addr.Port() == 0 {
		t.Error("Port should be assigned after listen")
	}

	result = server.Close()
	if _, ok := result.(*Error); ok {
		t.Fatalf("Close failed: %v", result)
	}

	if !server.IsClosed() {
		t.Error("closed server should be marked as closed")
	}
}

func TestTcpServerAccept(t *testing.T) {
	server := NewTcpServer()

	result := server.Listen(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Listen failed: %v", result)
	}
	defer server.Close()

	addr := server.Addr()

	// Connect in goroutine - keep connection open for a while
	connected := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		client := NewTcpClient()
		connResult := client.Connect(addr.ToStr())
		if _, ok := connResult.(*Error); !ok {
			close(connected)
			// Keep connection open for a while
			time.Sleep(500 * time.Millisecond)
		}
		client.Close()
	}()

	// Wait a bit for the connection attempt to start
	time.Sleep(100 * time.Millisecond)

	// Accept with timeout
	client := server.Accept(2000) // 2 second timeout
	// Check if client is NULL (the null object, not nil)
	if client == NULL {
		t.Fatal("Accept should return a client")
	}

	tc, ok := client.(*TcpClient)
	if !ok {
		t.Fatalf("Expected TcpClient, got %T", client)
	}
	tc.Close()
}

func TestTcpServerAcceptTimeout(t *testing.T) {
	server := NewTcpServer()

	result := server.Listen(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Listen failed: %v", result)
	}
	defer server.Close()

	// Accept with short timeout, no connection
	start := time.Now()
	client := server.Accept(100) // 100ms timeout
	elapsed := time.Since(start)

	// Check if client is NULL (the null object, not nil)
	if client != NULL {
		tc, ok := client.(*TcpClient)
		if ok {
			tc.Close()
		}
		t.Fatal("Accept should return null on timeout")
	}

	if elapsed < 80*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Accept should timeout around 100ms, took %v", elapsed)
	}
}

func TestTcpServerSetReuseAddr(t *testing.T) {
	server := NewTcpServer()

	// SetReuseAddr before listen should succeed
	result := server.SetReuseAddr(true)
	if _, ok := result.(*Error); ok {
		t.Errorf("SetReuseAddr should succeed: %v", result)
	}

	// Listen should work
	result = server.Listen(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Listen failed: %v", result)
	}
	defer server.Close()
}

func TestTcpServerSetTimeout(t *testing.T) {
	server := NewTcpServer()

	// SetTimeout should work
	result := server.SetTimeout(5000)
	if _, ok := result.(*Error); ok {
		t.Errorf("SetTimeout should succeed: %v", result)
	}
}

func TestTcpServerDoubleListen(t *testing.T) {
	server := NewTcpServer()

	result := server.Listen(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("First listen failed: %v", result)
	}
	defer server.Close()

	// Second listen should fail
	result = server.Listen(":0")
	if _, ok := result.(*Error); !ok {
		t.Error("Second listen should return Error")
	}
}

func TestTcpServerAddrBeforeListen(t *testing.T) {
	server := NewTcpServer()

	// Addr before listen should return nil
	addr := server.Addr()
	if addr != nil {
		t.Error("Addr before listen should return nil")
	}
}

func TestTcpServerAcceptBeforeListen(t *testing.T) {
	server := NewTcpServer()

	// Accept before listen should return error
	result := server.Accept(100)
	if _, ok := result.(*Error); !ok {
		t.Error("Accept before listen should return Error")
	}
}

func TestTcpServerOnAccept(t *testing.T) {
	server := NewTcpServer()

	// Set callback before listen
	result := server.OnAccept(func(client *TcpClient) {
		// Callback function - just close the client
		client.Close()
	})
	if _, ok := result.(*Error); ok {
		t.Errorf("OnAccept should succeed: %v", result)
	}
}

func TestTcpServerStartAndStop(t *testing.T) {
	server := NewTcpServer()

	result := server.Listen(":0")
	if _, ok := result.(*Error); ok {
		t.Fatalf("Listen failed: %v", result)
	}

	// Start should begin accept loop in a goroutine
	result = server.Start()
	if _, ok := result.(*Error); ok {
		t.Fatalf("Start failed: %v", result)
	}

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Close should stop the server
	result = server.Close()
	if _, ok := result.(*Error); ok {
		t.Fatalf("Close failed: %v", result)
	}
}

func TestTcpServerStartBeforeListen(t *testing.T) {
	server := NewTcpServer()

	// Start before listen should return error
	result := server.Start()
	if _, ok := result.(*Error); !ok {
		t.Error("Start before listen should return Error")
	}
}