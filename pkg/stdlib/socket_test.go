// pkg/stdlib/socket_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callSocketFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("socket")
	if mod == nil {
		panic("socket module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestSocketModuleExists(t *testing.T) {
	mod := Get("socket")
	if mod == nil {
		t.Fatal("socket module not registered")
	}
}

func TestSocketParseAddr(t *testing.T) {
	result := callSocketFunc("parseAddr", String("127.0.0.1:8080"))

	addr, ok := result.(*objects.SocketAddr)
	if !ok {
		t.Fatalf("parseAddr should return SocketAddr, got %T", result)
	}

	if addr.Host() != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", addr.Host())
	}

	if addr.Port() != 8080 {
		t.Errorf("expected port 8080, got %d", addr.Port())
	}
}

func TestSocketParseAddrError(t *testing.T) {
	// Wrong number of args
	result := callSocketFunc("parseAddr")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parseAddr with no args should return Error")
	}

	// Wrong type
	result = callSocketFunc("parseAddr", Int(42))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("parseAddr with non-string should return Error")
	}
}

func TestSocketCreateTcpServer(t *testing.T) {
	result := callSocketFunc("createTcpServer", String(":0"))

	server, ok := result.(*objects.TcpServer)
	if !ok {
		t.Fatalf("createTcpServer should return TcpServer, got %T", result)
	}

	// Should be listening
	if server.IsClosed() {
		t.Error("server should be listening")
	}

	// Cleanup
	server.Close()
}

func TestSocketCreateTcpClient(t *testing.T) {
	result := callSocketFunc("createTcpClient")

	client, ok := result.(*objects.TcpClient)
	if !ok {
		t.Fatalf("createTcpClient should return TcpClient, got %T", result)
	}

	if client.IsClosed() {
		t.Error("new client should not be closed")
	}
}

func TestSocketCreateUdpSocket(t *testing.T) {
	result := callSocketFunc("createUdpSocket")

	sock, ok := result.(*objects.UdpSocket)
	if !ok {
		t.Fatalf("createUdpSocket should return UdpSocket, got %T", result)
	}
	_ = sock
}

func TestSocketResolveHost(t *testing.T) {
	result := callSocketFunc("resolveHost", String("localhost"))

	_, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("resolveHost should return String, got %T", result)
	}
}

func TestSocketLookupHost(t *testing.T) {
	result := callSocketFunc("lookupHost", String("localhost"))

	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("lookupHost should return Array, got %T", result)
	}

	if len(arr.Elements) == 0 {
		t.Error("lookupHost should return at least one address")
	}
}