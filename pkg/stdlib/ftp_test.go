// pkg/stdlib/ftp_test.go
// Tests for FTP module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestFtpModule_Exists(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Fatal("ftp module not found")
	}
	if mod.Name != "ftp" {
		t.Errorf("expected module name 'ftp', got %s", mod.Name)
	}
}

func TestFtpModule_Exports(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	expectedExports := []string{"newClient", "connect", "connectWithConfig"}
	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in ftp module", name)
		}
	}
}

func TestFtpModule_NewClient(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	result := fn.Fn()
	if result == nil {
		t.Fatal("newClient returned nil")
	}

	client, ok := result.(*objects.FtpClient)
	if !ok {
		t.Fatalf("expected FtpClient, got %T", result)
	}
	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
}

func TestFtpModule_NewClientArgCount(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	result := fn.Fn(objects.NewString("unexpected"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg count")
	}
}

func TestFtpModule_ConnectArgCount(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["connect"].(*objects.Builtin)

	// Wrong arg count
	result := fn.Fn(objects.NewString("host"), objects.NewInt(21))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg count")
	}
}

func TestFtpModule_ConnectArgTypes(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["connect"].(*objects.Builtin)

	// Wrong type for host
	result := fn.Fn(objects.NewInt(0), objects.NewInt(21), objects.NewString("user"), objects.NewString("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong host type")
	}
}

func TestFtpModule_ConnectWithConfigArgCount(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["connectWithConfig"].(*objects.Builtin)

	// No args
	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no args")
	}
}

func TestFtpModule_ConnectWithConfigArgType(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}

	fn := mod.Exports["connectWithConfig"].(*objects.Builtin)

	// Wrong type
	result := fn.Fn(objects.NewString("not a map"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg type")
	}
}
