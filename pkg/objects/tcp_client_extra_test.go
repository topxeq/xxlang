package objects

import (
	"net"
	"testing"
	"time"
)

func TestTcpClient_New(t *testing.T) {
	client := NewTcpClient()
	if client == nil {
		t.Fatal("NewTcpClient() returned nil")
	}
	if client.closed {
		t.Error("New TCP client should not be closed")
	}
}

func TestTcpClient_NewFromConn(t *testing.T) {
	client := NewTcpClientFromConn(nil)
	if client == nil {
		t.Fatal("NewTcpClientFromConn() returned nil")
	}
}

func TestTcpClient_Type(t *testing.T) {
	client := NewTcpClient()
	if client.Type() != TcpClientType {
		t.Errorf("Type() = %v, want %v", client.Type(), TcpClientType)
	}
}

func TestTcpClient_TypeTag(t *testing.T) {
	client := NewTcpClient()
	if client.TypeTag() != TagTcpClient {
		t.Errorf("TypeTag() = %v, want %v", client.TypeTag(), TagTcpClient)
	}
}

func TestTcpClient_Inspect(t *testing.T) {
	client := NewTcpClient()

	if client.Inspect() != "<TCP_CLIENT unconnected>" {
		t.Errorf("Inspect() = %v, want <TCP_CLIENT unconnected>", client.Inspect())
	}

	client.closed = true
	if client.Inspect() != "<TCP_CLIENT closed>" {
		t.Errorf("Inspect() = %v, want <TCP_CLIENT closed>", client.Inspect())
	}
}

func TestTcpClient_ToBool(t *testing.T) {
	client := NewTcpClient()
	if client.ToBool().Value {
		t.Error("ToBool() should return false for unconnected client")
	}
}

func TestTcpClient_HashKey(t *testing.T) {
	client := NewTcpClient()
	hk := client.HashKey()
	if hk.Type != TcpClientType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, TcpClientType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}

func TestTcpClient_SetTimeout(t *testing.T) {
	client := NewTcpClient()

	result := client.SetTimeout(5000)
	if result != NULL {
		t.Errorf("SetTimeout() = %v, want NULL", result)
	}
}

func TestTcpClient_ConnectAlreadyConnected(t *testing.T) {
	client := NewTcpClient()
	client.conn = &mockConn{}

	result := client.Connect("127.0.0.1:9999")
	if _, ok := result.(*Error); !ok {
		t.Error("Connect() should return error for already connected client")
	}
}

func TestTcpClient_Close(t *testing.T) {
	client := NewTcpClient()

	result := client.Close()
	if result != NULL {
		t.Errorf("Close() = %v, want NULL", result)
	}
	if !client.closed {
		t.Error("Client should be closed after Close()")
	}
}

func TestTcpClient_SendStringClosed(t *testing.T) {
	client := NewTcpClient()
	client.Close()

	result := client.SendString("test")
	if _, ok := result.(*Error); !ok {
		t.Error("SendString() should return error for closed client")
	}
}

func TestTcpClient_SendBytesClosed(t *testing.T) {
	client := NewTcpClient()
	client.Close()

	result := client.SendBytes([]byte("test"))
	if _, ok := result.(*Error); !ok {
		t.Error("SendBytes() should return error for closed client")
	}
}

func TestTcpClient_ReceiveClosed(t *testing.T) {
	client := NewTcpClient()
	client.Close()

	result := client.Receive(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("Receive() should return error for closed client")
	}
}

func TestTcpClient_ReceiveLineClosed(t *testing.T) {
	client := NewTcpClient()
	client.Close()

	result := client.ReceiveLine()
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveLine() should return error for closed client")
	}
}

func TestTcpClient_ReceiveBytesClosed(t *testing.T) {
	client := NewTcpClient()
	client.Close()

	result := client.ReceiveBytes(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveBytes() should return error for closed client")
	}
}

type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
