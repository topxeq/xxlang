// pkg/objects/ssh_client_test.go
// Tests for SSHClient object.
package objects

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/testutil/mock"
)

func TestSSHClient_New(t *testing.T) {
	client := NewSSHClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
	if client.Inspect() != "SSHClient(disconnected)" {
		t.Errorf("unexpected inspect: %s", client.Inspect())
	}
	if client.ToBool().Value {
		t.Error("new client should convert to false")
	}
}

func TestSSHClient_Type(t *testing.T) {
	client := NewSSHClient()
	if client.Type() != SSHClientType {
		t.Errorf("expected SSHClientType, got %v", client.Type())
	}
	if client.TypeTag() != TagSSHClient {
		t.Errorf("expected TagSSHClient, got %v", client.TypeTag())
	}
}

func TestSSHClient_ConnectPassword(t *testing.T) {
	// Skip SSH tests that require network mock server (can hang in some environments)
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}

	// Create and start mock server
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")
	server.SetCommandResponse("echo hello", "hello\n")
	server.SetCommandResponse("pwd", "/home/testuser\n")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	// Create client and connect with config to ignore host key
	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Verify connection state
	if !client.IsConnected() {
		t.Error("client should be connected")
	}
	if client.GetHost() != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", client.GetHost())
	}
	if client.GetUser() != "testuser" {
		t.Errorf("expected user testuser, got %s", client.GetUser())
	}
	if client.GetPort() != server.Port() {
		t.Errorf("port mismatch")
	}

	// Verify inspect
	inspect := client.Inspect()
	if inspect == "SSHClient(disconnected)" {
		t.Error("inspect should show connected state")
	}

	// Test command execution
	output, err := client.Exec("echo hello")
	if err != nil {
		t.Errorf("failed to execute command: %v", err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestSSHClient_ConnectWithConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("configuser", "configpass")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "configuser",
		Password:      "configpass",
		Timeout:       10,
		IgnoreHostKey: true,
	}

	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect with config: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Error("client should be connected")
	}
}

func TestSSHClient_WrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "correctpass")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	err := client.Connect("127.0.0.1", server.Port(), "testuser", "wrongpass")
	if err == nil {
		client.Close()
		t.Error("expected error with wrong password")
	}
}

func TestSSHClient_ExecNotConnected(t *testing.T) {
	client := NewSSHClient()
	_, err := client.Exec("echo test")
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestSSHClient_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	// Close connection
	err = client.Close()
	if err != nil {
		t.Errorf("failed to close: %v", err)
	}

	if client.IsConnected() {
		t.Error("client should be disconnected after close")
	}

	// Close again should be no-op
	err = client.Close()
	if err != nil {
		t.Errorf("second close should not error: %v", err)
	}
}

func TestSSHClient_HashKey(t *testing.T) {
	client := NewSSHClient()
	hk := client.HashKey()
	if hk.Type != SSHClientType {
		t.Errorf("expected SSHClientType, got %v", hk.Type)
	}
}

func TestSSHClient_DoubleConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Try to connect again while already connected
	err = client.ConnectWithConfig(config)
	if err == nil {
		t.Error("expected error when connecting twice")
	}
}

func TestSSHClient_CommandResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")
	server.SetCommandResponse("ls -la", "total 0\n")
	server.SetCommandResponse("cat /etc/os-release", "NAME=\"Test\"\n")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Test various commands
	tests := []struct {
		cmd      string
		expected string
	}{
		{"ls -la", "total 0\n"},
		{"cat /etc/os-release", "NAME=\"Test\"\n"},
	}

	for _, tt := range tests {
		output, err := client.Exec(tt.cmd)
		if err != nil {
			t.Errorf("command %q failed: %v", tt.cmd, err)
			continue
		}
		if output != tt.expected {
			t.Errorf("command %q: expected %q, got %q", tt.cmd, tt.expected, output)
		}
	}
}

func TestSSHClient_ExecFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")
	server.SetCommandResponse("test command", "output\n")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	result, err := client.ExecFull("test command")
	if err != nil {
		t.Fatalf("ExecFull failed: %v", err)
	}

	if result["stdout"] == nil {
		t.Error("expected stdout in result")
	}
	if result["stderr"] == nil {
		t.Error("expected stderr in result")
	}
	if result["exitCode"] == nil {
		t.Error("expected exitCode in result")
	}
}

func TestSSHClient_FileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH file operations test in short mode")
	}
	// File operations require a real SFTP subsystem which the mock server
	// does not provide. Test with a real server instead.
	t.Skip("file operations require real SFTP server; tested via backup_test1.xxl")
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")
	server.SetCommandResponse("cat '/tmp/test.txt'", "file content\n")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewSSHClient()
	config := &SSHConfig{
		Host:          "127.0.0.1",
		Port:          server.Port(),
		User:          "testuser",
		Password:      "testpass",
		IgnoreHostKey: true,
	}
	err := client.ConnectWithConfig(config)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Test ReadFile (uses cat command)
	// The mock server will respond based on prefix matching
	output, err := client.ReadFile("/tmp/test.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	_ = output // Mock server responds to configured commands
}

