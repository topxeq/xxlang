package objects

import (
	"testing"
)

func TestUdpSocket_New(t *testing.T) {
	socket := NewUdpSocket()
	if socket == nil {
		t.Fatal("NewUdpSocket() returned nil")
	}
	if socket.closed {
		t.Error("New UDP socket should not be closed")
	}
}

func TestUdpSocket_Type(t *testing.T) {
	socket := NewUdpSocket()
	if socket.Type() != UdpSocketType {
		t.Errorf("Type() = %v, want %v", socket.Type(), UdpSocketType)
	}
}

func TestUdpSocket_TypeTag(t *testing.T) {
	socket := NewUdpSocket()
	if socket.TypeTag() != TagUdpSocket {
		t.Errorf("TypeTag() = %v, want %v", socket.TypeTag(), TagUdpSocket)
	}
}

func TestUdpSocket_Inspect(t *testing.T) {
	socket := NewUdpSocket()

	if socket.Inspect() != "<UDP_SOCKET unbound>" {
		t.Errorf("Inspect() = %v, want <UDP_SOCKET unbound>", socket.Inspect())
	}

	socket.closed = true
	if socket.Inspect() != "<UDP_SOCKET closed>" {
		t.Errorf("Inspect() = %v, want <UDP_SOCKET closed>", socket.Inspect())
	}
}

func TestUdpSocket_ToBool(t *testing.T) {
	socket := NewUdpSocket()
	if !socket.ToBool().Value {
		t.Error("ToBool() should return true for open socket")
	}

	socket.closed = true
	if socket.ToBool().Value {
		t.Error("ToBool() should return false for closed socket")
	}
}

func TestUdpSocket_HashKey(t *testing.T) {
	socket := NewUdpSocket()
	hk := socket.HashKey()
	if hk.Type != UdpSocketType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, UdpSocketType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}

func TestUdpSocket_SetTimeout(t *testing.T) {
	socket := NewUdpSocket()

	result := socket.SetTimeout(5000)
	if result != NULL {
		t.Errorf("SetTimeout() = %v, want NULL", result)
	}
}

func TestUdpSocket_Bind(t *testing.T) {
	socket := NewUdpSocket()
	defer socket.Close()

	result := socket.Bind(":0")
	if result != NULL {
		t.Errorf("Bind() = %v, want NULL", result)
	}
	if !socket.bound {
		t.Error("Socket should be bound after Bind()")
	}
}

func TestUdpSocket_BindAlreadyBound(t *testing.T) {
	socket := NewUdpSocket()
	defer socket.Close()

	socket.Bind(":0")
	result := socket.Bind(":0")
	if _, ok := result.(*Error); !ok {
		t.Error("Bind() should return error for already bound socket")
	}
}

func TestUdpSocket_BindClosed(t *testing.T) {
	socket := NewUdpSocket()
	socket.Close()

	result := socket.Bind(":0")
	if _, ok := result.(*Error); !ok {
		t.Error("Bind() should return error for closed socket")
	}
}

func TestUdpSocket_Close(t *testing.T) {
	socket := NewUdpSocket()

	result := socket.Close()
	if result != NULL {
		t.Errorf("Close() = %v, want NULL", result)
	}
	if !socket.closed {
		t.Error("Socket should be closed after Close()")
	}
}

func TestUdpSocket_SendToClosed(t *testing.T) {
	socket := NewUdpSocket()
	socket.Close()

	result := socket.SendTo("test", "127.0.0.1:12345")
	if _, ok := result.(*Error); !ok {
		t.Error("SendTo() should return error for closed socket")
	}
}

func TestUdpSocket_SendToBytesClosed(t *testing.T) {
	socket := NewUdpSocket()
	socket.Close()

	result := socket.SendToBytes([]byte("test"), "127.0.0.1:12345")
	if _, ok := result.(*Error); !ok {
		t.Error("SendToBytes() should return error for closed socket")
	}
}

func TestUdpSocket_ReceiveFromClosed(t *testing.T) {
	socket := NewUdpSocket()
	socket.Close()

	result, _ := socket.ReceiveFrom(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveFrom() should return error for closed socket")
	}
}

func TestUdpSocket_ReceiveFromUnbound(t *testing.T) {
	socket := NewUdpSocket()

	result, _ := socket.ReceiveFrom(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveFrom() should return error for unbound socket")
	}
}

func TestUdpSocket_ReceiveFromBytesClosed(t *testing.T) {
	socket := NewUdpSocket()
	socket.Close()

	result, _ := socket.ReceiveFromBytes(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveFromBytes() should return error for closed socket")
	}
}

func TestUdpSocket_ReceiveFromBytesUnbound(t *testing.T) {
	socket := NewUdpSocket()

	result, _ := socket.ReceiveFromBytes(1024)
	if _, ok := result.(*Error); !ok {
		t.Error("ReceiveFromBytes() should return error for unbound socket")
	}
}
