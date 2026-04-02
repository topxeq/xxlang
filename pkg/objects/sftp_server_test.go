package objects

import (
	"testing"
)

func TestSftpServer_New(t *testing.T) {
	server := NewSftpServer()
	if server == nil {
		t.Fatal("NewSftpServer() returned nil")
	}
	if server.IsRunning() {
		t.Error("New SFTP server should not be running")
	}
}

func TestSftpServer_Type(t *testing.T) {
	server := NewSftpServer()
	if server.Type() != SftpServerType {
		t.Errorf("Type() = %v, want %v", server.Type(), SftpServerType)
	}
}

func TestSftpServer_TypeTag(t *testing.T) {
	server := NewSftpServer()
	if server.TypeTag() != TagSftpServer {
		t.Errorf("TypeTag() = %v, want %v", server.TypeTag(), TagSftpServer)
	}
}

func TestSftpServer_Inspect(t *testing.T) {
	server := NewSftpServer()
	server.host = "localhost"
	server.port = 22

	if server.Inspect() != "SftpServer(stopped)" {
		t.Errorf("Inspect() = %v, want SftpServer(stopped)", server.Inspect())
	}

	server.running = true
	if server.Inspect() != "SftpServer(running on localhost:22)" {
		t.Errorf("Inspect() = %v, want SftpServer(running on localhost:22)", server.Inspect())
	}
}

func TestSftpServer_ToBool(t *testing.T) {
	server := NewSftpServer()
	if server.ToBool().Value {
		t.Error("ToBool() should return false for stopped server")
	}

	server.running = true
	if !server.ToBool().Value {
		t.Error("ToBool() should return true for running server")
	}
}

func TestSftpServer_HashKey(t *testing.T) {
	server := NewSftpServer()
	hk := server.HashKey()
	if hk.Type != SftpServerType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, SftpServerType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}

func TestSftpServer_Create(t *testing.T) {
	server := NewSftpServer()

	err := server.Create("localhost:2222", nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if server.host != "localhost" {
		t.Errorf("host = %v, want localhost", server.host)
	}
	if server.port != 2222 {
		t.Errorf("port = %v, want 2222", server.port)
	}
}

func TestSftpServer_CreateWithConfig(t *testing.T) {
	server := NewSftpServer()

	config := &SftpServerConfig{
		Host:           "0.0.0.0",
		Port:           8022,
		MaxConnections: 50,
		Timeout:        600,
		HostKey:        "test-key",
		HostKeyPath:    "/path/to/key",
	}

	err := server.Create(":2222", config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if server.host != "0.0.0.0" {
		t.Errorf("host = %v, want 0.0.0.0", server.host)
	}
	if server.config.MaxConnections != 50 {
		t.Errorf("MaxConnections = %v, want 50", server.config.MaxConnections)
	}
	if server.config.HostKey != "test-key" {
		t.Errorf("HostKey = %v, want test-key", server.config.HostKey)
	}
}

func TestSftpServer_AddUser(t *testing.T) {
	server := NewSftpServer()

	err := server.AddUser("testuser", "testpass", "/home/test")
	if err != nil {
		t.Errorf("AddUser() error = %v", err)
	}

	if len(server.users) != 1 {
		t.Errorf("expected 1 user, got %d", len(server.users))
	}
}

func TestSftpServer_RemoveUser(t *testing.T) {
	server := NewSftpServer()
	server.AddUser("testuser", "testpass", "/home/test")

	err := server.RemoveUser("testuser")
	if err != nil {
		t.Errorf("RemoveUser() error = %v", err)
	}

	if len(server.users) != 0 {
		t.Errorf("expected 0 users, got %d", len(server.users))
	}
}

func TestSftpServer_StopNotRunning(t *testing.T) {
	server := NewSftpServer()

	err := server.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running server should not error: %v", err)
	}
}

func TestParseSftpAddr(t *testing.T) {
	tests := []struct {
		addr        string
		defaultPort int
		wantHost    string
		wantPort    int
	}{
		{"localhost:2222", 22, "localhost", 2222},
		{"localhost", 22, "localhost", 22},
		{"192.168.1.1:8022", 22, "192.168.1.1", 8022},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			host, port := parseSftpAddr(tt.addr, tt.defaultPort)
			if host != tt.wantHost {
				t.Errorf("host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}
