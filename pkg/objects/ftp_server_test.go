package objects

import (
	"testing"
)

func TestFtpServer_New(t *testing.T) {
	server := NewFtpServer()
	if server == nil {
		t.Fatal("NewFtpServer() returned nil")
	}
	if server.IsRunning() {
		t.Error("New FTP server should not be running")
	}
}

func TestFtpServer_Type(t *testing.T) {
	server := NewFtpServer()
	if server.Type() != FtpServerType {
		t.Errorf("Type() = %v, want %v", server.Type(), FtpServerType)
	}
}

func TestFtpServer_TypeTag(t *testing.T) {
	server := NewFtpServer()
	if server.TypeTag() != TagFtpServer {
		t.Errorf("TypeTag() = %v, want %v", server.TypeTag(), TagFtpServer)
	}
}

func TestFtpServer_Inspect(t *testing.T) {
	server := NewFtpServer()
	server.host = "localhost"
	server.port = 21

	if server.Inspect() != "FtpServer(stopped)" {
		t.Errorf("Inspect() = %v, want FtpServer(stopped)", server.Inspect())
	}

	server.running = true
	if server.Inspect() != "FtpServer(running on localhost:21)" {
		t.Errorf("Inspect() = %v, want FtpServer(running on localhost:21)", server.Inspect())
	}
}

func TestFtpServer_ToBool(t *testing.T) {
	server := NewFtpServer()
	if server.ToBool().Value {
		t.Error("ToBool() should return false for stopped server")
	}

	server.running = true
	if !server.ToBool().Value {
		t.Error("ToBool() should return true for running server")
	}
}

func TestFtpServer_Create(t *testing.T) {
	server := NewFtpServer()

	err := server.Create("localhost:2121", nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if server.host != "localhost" {
		t.Errorf("host = %v, want localhost", server.host)
	}
	if server.port != 2121 {
		t.Errorf("port = %v, want 2121", server.port)
	}
}

func TestFtpServer_CreateWithConfig(t *testing.T) {
	server := NewFtpServer()

	config := &FtpServerConfig{
		Host:           "0.0.0.0",
		Port:           8021,
		MaxConnections: 50,
		Timeout:        600,
		WelcomeMessage: "Welcome",
		PassivePortMin: 60000,
		PassivePortMax: 61000,
	}

	err := server.Create(":2121", config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if server.host != "0.0.0.0" {
		t.Errorf("host = %v, want 0.0.0.0", server.host)
	}
	if server.config.MaxConnections != 50 {
		t.Errorf("MaxConnections = %v, want 50", server.config.MaxConnections)
	}
}

func TestFtpServer_AddUser(t *testing.T) {
	server := NewFtpServer()

	err := server.AddUser("testuser", "testpass", "/home/test")
	if err != nil {
		t.Errorf("AddUser() error = %v", err)
	}

	if len(server.users) != 1 {
		t.Errorf("expected 1 user, got %d", len(server.users))
	}

	user, ok := server.users["testuser"]
	if !ok {
		t.Error("user not found")
		return
	}
	if user.password != "testpass" {
		t.Errorf("password = %v, want testpass", user.password)
	}
}

func TestFtpServer_RemoveUser(t *testing.T) {
	server := NewFtpServer()
	server.AddUser("testuser", "testpass", "/home/test")

	err := server.RemoveUser("testuser")
	if err != nil {
		t.Errorf("RemoveUser() error = %v", err)
	}

	if len(server.users) != 0 {
		t.Errorf("expected 0 users, got %d", len(server.users))
	}
}

func TestFtpServer_StopNotRunning(t *testing.T) {
	server := NewFtpServer()

	err := server.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running server should not error: %v", err)
	}
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		addr        string
		defaultPort int
		wantHost    string
		wantPort    int
	}{
		{"localhost:2121", 21, "localhost", 2121},
		{"localhost", 21, "localhost", 21},
		{"192.168.1.1:8021", 21, "192.168.1.1", 8021},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			host, port := parseAddr(tt.addr, tt.defaultPort)
			if host != tt.wantHost {
				t.Errorf("host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		s       string
		wantMin int
		wantMax int
	}{
		{"50000-51000", 50000, 51000},
		{"60000-61000", 60000, 61000},
		{"invalid", 0, 0},
		{"", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			min, max := parsePortRange(tt.s)
			if min != tt.wantMin {
				t.Errorf("min = %v, want %v", min, tt.wantMin)
			}
			if max != tt.wantMax {
				t.Errorf("max = %v, want %v", max, tt.wantMax)
			}
		})
	}
}

func TestFtpServer_HashKey(t *testing.T) {
	server := NewFtpServer()
	hk := server.HashKey()
	if hk.Type != FtpServerType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, FtpServerType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}
