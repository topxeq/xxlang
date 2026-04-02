// pkg/objects/sftp_client_test.go
// Tests for SftpClient object.
package objects

import (
	"testing"
)

func TestSftpClient_New(t *testing.T) {
	client := NewSftpClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
}

func TestSftpClient_Type(t *testing.T) {
	client := NewSftpClient()
	if client.Type() != SftpClientType {
		t.Errorf("expected SftpClientType, got %v", client.Type())
	}
	if client.TypeTag() != TagSftpClient {
		t.Errorf("expected TagSftpClient, got %v", client.TypeTag())
	}
}

func TestSftpClient_Inspect(t *testing.T) {
	client := NewSftpClient()
	if client.Inspect() != "SftpClient(disconnected)" {
		t.Errorf("unexpected inspect: %s", client.Inspect())
	}
}

func TestSftpClient_HashKey(t *testing.T) {
	client := NewSftpClient()
	hk := client.HashKey()
	if hk.Type != SftpClientType {
		t.Errorf("expected SftpClientType, got %v", hk.Type)
	}
}

func TestSftpClient_ToBool(t *testing.T) {
	client := NewSftpClient()
	if client.ToBool().Value {
		t.Error("new client should be false")
	}
}

func TestSftpClient_ConnectNotRunning(t *testing.T) {
	client := NewSftpClient()
	err := client.Connect("127.0.0.1", 22, "user", "pass")
	if err == nil {
		client.Close()
		t.Error("expected error connecting to non-existent server")
	}
}

func TestSftpClient_OperationsNotConnected(t *testing.T) {
	client := NewSftpClient()

	if client.Exists("/tmp/file") {
		t.Error("Exists should return false when not connected")
	}
	if client.IsDir("/tmp/dir") {
		t.Error("IsDir should return false when not connected")
	}
	if client.IsFile("/tmp/file") {
		t.Error("IsFile should return false when not connected")
	}
	if err := client.Mkdir("/tmp/dir"); err == nil {
		t.Error("expected error for Mkdir when not connected")
	}
	if err := client.Rename("/tmp/a", "/tmp/b"); err == nil {
		t.Error("expected error for Rename when not connected")
	}
}

func TestSftpClient_CloseNotConnected(t *testing.T) {
	client := NewSftpClient()
	err := client.Close()
	if err != nil {
		t.Errorf("close on unconnected should not error: %v", err)
	}
}
