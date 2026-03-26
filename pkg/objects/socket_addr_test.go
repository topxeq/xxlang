// pkg/objects/socket_addr_test.go
package objects

import (
	"testing"
)

func TestSocketAddrBasic(t *testing.T) {
	addr := NewSocketAddr("127.0.0.1", 8080)
	if addr == nil {
		t.Fatal("NewSocketAddr returned nil")
	}

	if addr.Type() != SocketAddrType {
		t.Errorf("expected type %s, got %s", SocketAddrType, addr.Type())
	}

	if addr.TypeTag() != TagSocketAddr {
		t.Errorf("expected type tag %d, got %d", TagSocketAddr, addr.TypeTag())
	}

	if addr.ToBool() != TRUE {
		t.Error("SocketAddr should be truthy")
	}
}

func TestSocketAddrFields(t *testing.T) {
	addr := NewSocketAddr("192.168.1.1", 9000)

	if addr.Host() != "192.168.1.1" {
		t.Errorf("expected host '192.168.1.1', got '%s'", addr.Host())
	}

	if addr.Port() != 9000 {
		t.Errorf("expected port 9000, got %d", addr.Port())
	}

	if addr.ToStr() != "192.168.1.1:9000" {
		t.Errorf("expected '192.168.1.1:9000', got '%s'", addr.ToStr())
	}
}

func TestSocketAddrInspect(t *testing.T) {
	addr := NewSocketAddr("localhost", 80)
	inspect := addr.Inspect()
	if inspect == "" {
		t.Error("Inspect should not be empty")
	}
}

func TestSocketAddrHashKey(t *testing.T) {
	addr1 := NewSocketAddr("127.0.0.1", 8080)
	addr2 := NewSocketAddr("127.0.0.1", 8080)

	// Same address should have same hash key
	if addr1.HashKey() != addr2.HashKey() {
		t.Error("Same addresses should have same hash keys")
	}

	addr3 := NewSocketAddr("127.0.0.1", 9090)
	if addr1.HashKey() == addr3.HashKey() {
		t.Error("Different ports should have different hash keys")
	}
}