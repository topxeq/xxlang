package objects

import (
	"testing"
)

func TestTcpServer_New(t *testing.T) {
	server := NewTcpServer()
	if server == nil {
		t.Fatal("NewTcpServer() returned nil")
	}
	if !server.closed {
		t.Error("New TCP server should be closed initially")
	}
}

func TestTcpServer_Type(t *testing.T) {
	server := NewTcpServer()
	if server.Type() != TcpServerType {
		t.Errorf("Type() = %v, want %v", server.Type(), TcpServerType)
	}
}

func TestTcpServer_TypeTag(t *testing.T) {
	server := NewTcpServer()
	if server.TypeTag() != TagTcpServer {
		t.Errorf("TypeTag() = %v, want %v", server.TypeTag(), TagTcpServer)
	}
}

func TestTcpServer_Inspect(t *testing.T) {
	server := NewTcpServer()

	if server.Inspect() != "<TCP_SERVER closed>" {
		t.Errorf("Inspect() = %v, want <TCP_SERVER closed>", server.Inspect())
	}
}

func TestTcpServer_ToBool(t *testing.T) {
	server := NewTcpServer()
	if server.ToBool().Value {
		t.Error("ToBool() should return false for closed server")
	}
}

func TestTcpServer_HashKey(t *testing.T) {
	server := NewTcpServer()
	hk := server.HashKey()
	if hk.Type != TcpServerType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, TcpServerType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}

func TestTcpServer_SetTimeout(t *testing.T) {
	server := NewTcpServer()

	result := server.SetTimeout(5000)
	if result != NULL {
		t.Errorf("SetTimeout() = %v, want NULL", result)
	}
}

func TestTcpServer_SetReuseAddr(t *testing.T) {
	server := NewTcpServer()

	result := server.SetReuseAddr(true)
	if result != NULL {
		t.Errorf("SetReuseAddr() = %v, want NULL", result)
	}
}

func TestTcpServer_Listen(t *testing.T) {
	server := NewTcpServer()
	defer server.Close()

	result := server.Listen(":0")
	if result != NULL {
		t.Errorf("Listen() = %v, want NULL", result)
	}
	if server.closed {
		t.Error("Server should not be closed after Listen()")
	}
}

func TestTcpServer_ListenAlreadyListening(t *testing.T) {
	server := NewTcpServer()
	defer server.Close()

	server.Listen(":0")
	result := server.Listen(":0")
	if _, ok := result.(*Error); !ok {
		t.Error("Listen() should return error for already listening server")
	}
}

func TestTcpServer_Close(t *testing.T) {
	server := NewTcpServer()

	result := server.Close()
	if result != NULL {
		t.Errorf("Close() = %v, want NULL", result)
	}
	if !server.closed {
		t.Error("Server should be closed after Close()")
	}
}

func TestTcpServer_AcceptNotListening(t *testing.T) {
	server := NewTcpServer()

	result := server.Accept(100)
	if _, ok := result.(*Error); !ok {
		t.Error("Accept() should return error for non-listening server")
	}
}

func TestTcpServer_OnAccept(t *testing.T) {
	server := NewTcpServer()

	result := server.OnAccept(func(client *TcpClient) {})
	if result != NULL {
		t.Errorf("OnAccept() = %v, want NULL", result)
	}
}
