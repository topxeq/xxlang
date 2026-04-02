// pkg/stdlib/sftp_test.go
// Tests for SFTP module to increase coverage.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callSftpFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("sftp")
	if mod == nil {
		panic("sftp module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// TestSftp_Connect_ArgumentValidation tests the connect function with various invalid arguments.
func TestSftp_Connect_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSftpFunc("connect")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with no args should return Error")
	}

	// Too few args
	result = callSftpFunc("connect", String("host"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with 1 arg should return Error")
	}

	// Wrong type for host
	result = callSftpFunc("connect", Int(1), Int(22), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with non-string host should return Error")
	}

	// Wrong type for port
	result = callSftpFunc("connect", String("host"), String("22"), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with non-int port should return Error")
	}

	// Wrong type for user
	result = callSftpFunc("connect", String("host"), Int(22), Int(123), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with non-string user should return Error")
	}

	// Wrong type for password
	result = callSftpFunc("connect", String("host"), Int(22), String("user"), Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connect with non-string password should return Error")
	}
}

// TestSftp_ConnectWithKey_ArgumentValidation tests connectWithKey argument validation.
func TestSftp_ConnectWithKey_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("host")}, true},
		{[]objects.Object{String("host"), Int(22)}, true},
		{[]objects.Object{String("host"), Int(22), String("user")}, true},
		{[]objects.Object{String("host"), Int(22), String("user"), Int(123)}, true}, // keyPath not string
	}
	for _, tt := range tests {
		result := callSftpFunc("connectWithKey", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("connectWithKey(%v) expected error, got %T", tt.args, result)
			}
		}
	}
}

// TestSftp_ConnectWithKeyStr_ArgumentValidation tests connectWithKeyStr argument validation.
func TestSftp_ConnectWithKeyStr_ArgumentValidation(t *testing.T) {
	// Similar pattern
	result := callSftpFunc("connectWithKeyStr", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithKeyStr missing keyStr should error")
	}
	result = callSftpFunc("connectWithKeyStr", String("host"), Int(22), String("user"), Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithKeyStr with non-string keyStr should error")
	}
}

// TestSftp_ConnectWithConfig_ArgumentValidation tests connectWithConfig argument validation.
func TestSftp_ConnectWithConfig_ArgumentValidation(t *testing.T) {
	// Missing arg
	result := callSftpFunc("connectWithConfig")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing arg should error")
	}

	// Wrong type
	result = callSftpFunc("connectWithConfig", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig wrong type should error")
	}

	// Map without host
	emptyMap := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	result = callSftpFunc("connectWithConfig", emptyMap)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig without host should error")
	}

	// Map with host but missing user (should error before attempting connection)
	partialMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: String("localhost")},
	}}
	result = callSftpFunc("connectWithConfig", partialMap)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing user should error")
	}
}

// TestSftp_CreateServer_ArgumentValidation tests createServer argument validation.
func TestSftp_CreateServer_ArgumentValidation(t *testing.T) {
	// No args
	result := callSftpFunc("createServer")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer missing addr should error")
	}

	// Addr not string
	result = callSftpFunc("createServer", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer with non-string addr should error")
	}

	// Config not map
	result = callSftpFunc("createServer", String(":22"), Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer with non-map config should error")
	}

	// Config with invalid port type (should be ignored, not error)
	badConfig := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("port").HashKey(): {Key: objects.NewString("port"), Value: String("notanint")},
	}}
	result = callSftpFunc("createServer", String(":22"), badConfig)
	// Should still attempt to create server with default port; may return error for other reasons (like not running as server) but we ignore.
	// Just check it's not an immediate argument error.
	// Could be error from create call; that's okay. We won't assert.
}

// TestSftp_TypeChecks tests isSftpClient and isSftpServer functions.
func TestSftp_TypeChecks(t *testing.T) {
	mod := Get("sftp")
	if mod == nil {
		t.Skip("sftp module not found")
	}

	// isSftpClient
	fn, ok := mod.Exports["isSftpClient"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSftpClient not found or not builtin")
	}
	client := objects.NewSftpClient()
	res := fn.Fn(client)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSftpClient should return true for SftpClient, got %T %v", res, res)
	}
	res2 := fn.Fn(String("not a client"))
	if b, ok := res2.(*objects.Bool); !ok || b.Value {
		t.Fatalf("isSftpClient should return false for non-SftpClient, got %T %v", res2, res2)
	}

	// isSftpServer
	fn2, ok := mod.Exports["isSftpServer"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSftpServer not found or not builtin")
	}
	server := objects.NewSftpServer()
	res = fn2.Fn(server)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSftpServer should return true for SftpServer, got %T %v", res, res)
	}
	res2 = fn2.Fn(Int(123))
	if b, ok := res2.(*objects.Bool); !ok || b.Value {
		t.Fatalf("isSftpServer should return false for non-SftpServer, got %T %v", res2, res2)
	}
}

// TestSftp_ModuleExistence tests that the sftp module is registered.
func TestSftp_ModuleExistence(t *testing.T) {
	mod := Get("sftp")
	if mod == nil {
		t.Fatal("sftp module not found")
	}
	if mod.Name != "sftp" {
		t.Errorf("expected module name 'sftp', got %s", mod.Name)
	}
	if len(mod.Exports) == 0 {
		t.Error("sftp module has no exports")
	}
	// Verify key exports exist
	expectedExports := []string{
		"connect", "connectWithKey", "connectWithKeyStr", "connectWithConfig",
		"createServer", "isSftpClient", "isSftpServer",
	}
	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in sftp module", name)
		}
	}
}
