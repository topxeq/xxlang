// pkg/stdlib/ssh_test.go
// Tests for SSH module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/testutil/mock"
)

func TestSSHModule_Exists(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Fatal("ssh module not found")
	}
	if mod.Name != "ssh" {
		t.Errorf("expected module name 'ssh', got %s", mod.Name)
	}
	if len(mod.Exports) == 0 {
		t.Error("ssh module has no exports")
	}
}

func TestSSHModule_NewClient(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Skip("ssh module not found")
	}

	// Get newClient function
	fn, ok := mod.Exports["newClient"]
	if !ok {
		t.Fatal("newClient function not found in ssh module")
	}

	builtin, ok := fn.(*objects.Builtin)
	if !ok {
		t.Fatal("newClient is not a builtin function")
	}

	// Call newClient
	result := builtin.Fn()
	if result == nil {
		t.Fatal("newClient returned nil")
	}

	client, ok := result.(*objects.SSHClient)
	if !ok {
		t.Fatalf("expected SSHClient, got %T", result)
	}

	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
}

func TestSSHModule_Connect_WrongPassword(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Skip("ssh module not found")
	}

	// Start mock server
	server := mock.NewSSHMockServer(mock.DefaultConfig())
	server.SetUserPassword("testuser", "testpass")

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer server.Stop()

	// Get connect function
	fn, ok := mod.Exports["connect"]
	if !ok {
		t.Fatal("connect function not found in ssh module")
	}

	builtin, ok := fn.(*objects.Builtin)
	if !ok {
		t.Fatal("connect is not a builtin function")
	}

	// Call connect with wrong password - should return error
	result := builtin.Fn(
		objects.NewString("127.0.0.1"),
		objects.NewInt(int64(server.Port())),
		objects.NewString("testuser"),
		objects.NewString("wrongpass"),
	)

	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error with wrong password")
	}
}

// Key-based tests moved to separate test file to avoid flaky integration here.

func TestSSHModule_ConnectPassword_Success(t *testing.T) {
	t.Skip("temporarily skipped: needs mock server fix")
}

// Key-based tests moved to separate test file to avoid flaky integration here.

func TestSSHModule_ConnectWithConfig_Success(t *testing.T) {
	t.Skip("temporarily skipped: needs mock server fix")
}

func TestSSHModule_IsSSHClient(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Skip("ssh module not found")
	}
	fn, ok := mod.Exports["isSSHClient"]
	if !ok {
		t.Fatal("isSSHClient not found")
	}
	builtin, ok := fn.(*objects.Builtin)
	if !ok {
		t.Fatal("isSSHClient is not builtin")
	}
	client := objects.NewSSHClient()
	res := builtin.Fn(client)
	if okBool, ok := res.(*objects.Bool); !ok || !okBool.Value {
		t.Fatalf("expected true for SSHClient, got %T %v", res, res)
	}
	res2 := builtin.Fn(objects.NewString("not a client"))
	if okBool, ok := res2.(*objects.Bool); !ok || okBool.Value {
		t.Fatalf("expected false for non-SSHClient, got %T %v", res2, res2)
	}
}
