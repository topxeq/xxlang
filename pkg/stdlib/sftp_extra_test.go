// pkg/stdlib/sftp_extra_test.go
// Comprehensive tests for SFTP module to increase coverage.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestSftpConnect_ArgumentValidation tests connect argument validation.
func TestSftpConnect_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("host")}, true},
		{[]objects.Object{String("host"), Int(22)}, true},
		{[]objects.Object{String("host"), Int(22), String("user")}, true},
		{[]objects.Object{String("host"), Int(22), String("user"), Int(123)}, true}, // password not string
		{[]objects.Object{String("localhost"), Int(22), String("user"), String("pass")}, false},
	}
	for _, tt := range tests {
		result := callSftpFunc("connect", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("connect(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("connect(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSftpConnectWithKey_ArgumentValidation tests connectWithKey argument validation.
func TestSftpConnectWithKey_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("host")}, true},
		{[]objects.Object{String("host"), Int(22)}, true},
		{[]objects.Object{String("host"), Int(22), String("user")}, true},
		{[]objects.Object{String("host"), Int(22), String("user"), Int(123)}, true}, // keyPath not string
		// Valid types; connection may fail but not due to argument validation
		{[]objects.Object{String("localhost"), Int(22), String("testuser"), String("nonexistent_key.pem")}, false},
	}
	for _, tt := range tests {
		result := callSftpFunc("connectWithKey", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("connectWithKey(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("connectWithKey(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestSftpConnectWithKeyStr_ArgumentValidation tests connectWithKeyStr argument validation.
func TestSftpConnectWithKeyStr_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("host")}, true},
		{[]objects.Object{String("host"), Int(22)}, true},
		{[]objects.Object{String("host"), Int(22), String("user")}, true},
		{[]objects.Object{String("host"), Int(22), String("user"), Int(123)}, true}, // keyStr not string
		{[]objects.Object{String("host"), Int(22), String("user"), String("-----BEGIN RSA PRIVATE KEY-----")}, false},
	}
	for _, tt := range tests {
		result := callSftpFunc("connectWithKeyStr", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("connectWithKeyStr(%v) expected error, got %T", tt.args, result)
			}
		}
	}
}

// TestSftpConnectWithConfig_ArgumentValidation tests connectWithConfig argument validation.
func TestSftpConnectWithConfig_ArgumentValidation(t *testing.T) {
	// Missing arg
	result := callSftpFunc("connectWithConfig")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing arg should error")
	}

	// Wrong type - not a map
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

	// Map with host but missing user
	partialMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: String("localhost")},
	}}
	result = callSftpFunc("connectWithConfig", partialMap)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing user should error")
	}

	// Valid config with host and user (may fail connection but not arg validation)
	validMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: String("localhost")},
		objects.NewString("user").HashKey(): {Key: objects.NewString("user"), Value: String("testuser")},
	}}
	result = callSftpFunc("connectWithConfig", validMap)
	_ = result // we don't assert on result, just that no immediate arg error
}

// TestSftpCreateServer_ArgumentValidation tests createServer argument validation.
func TestSftpCreateServer_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSftpFunc("createServer")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer with no args should return Error")
	}

	// First arg not string
	result = callSftpFunc("createServer", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer with non-string addr should error")
	}

	// Second arg not map (if provided)
	result = callSftpFunc("createServer", String(":22"), String("not a map"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("createServer with non-map config should error")
	}

	// Valid call with just addr (may fail to start but arg validation ok)
	result = callSftpFunc("createServer", String(":0"))
	_ = result // don't assert, just ensure no arg validation error
}

// TestSftpIsSftpClient_TypeCheck tests isSftpClient with various types.
func TestSftpIsSftpClient_TypeCheck(t *testing.T) {
	mod := Get("sftp")
	if mod == nil {
		t.Skip("sftp module not found")
	}
	fn, ok := mod.Exports["isSftpClient"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSftpClient not found or not builtin")
	}

	// Test with SftpClient object
	client := objects.NewSftpClient()
	res := fn.Fn(client)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSftpClient should return true for SftpClient, got %T %v", res, res)
	}

	// Test with non-SftpClient objects
	nonClientTypes := []objects.Object{
		String("not a client"),
		Int(123),
		Bool(false),
		Null(),
	}
	for _, obj := range nonClientTypes {
		res2 := fn.Fn(obj)
		if b, ok := res2.(*objects.Bool); !ok || b.Value {
			t.Fatalf("isSftpClient should return false for %T, got %T %v", obj, res2, res2)
		}
	}
}

// TestSftpIsSftpServer_TypeCheck tests isSftpServer with various types.
func TestSftpIsSftpServer_TypeCheck(t *testing.T) {
	mod := Get("sftp")
	if mod == nil {
		t.Skip("sftp module not found")
	}
	fn, ok := mod.Exports["isSftpServer"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSftpServer not found or not builtin")
	}

	// Test with SftpServer object
	server := objects.NewSftpServer()
	res := fn.Fn(server)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSftpServer should return true for SftpServer, got %T %v", res, res)
	}

	// Test with non-SftpServer objects
	nonServerTypes := []objects.Object{
		String("not a server"),
		Int(123),
		Bool(false),
		Null(),
	}
	for _, obj := range nonServerTypes {
		res2 := fn.Fn(obj)
		if b, ok := res2.(*objects.Bool); !ok || b.Value {
			t.Fatalf("isSftpServer should return false for %T, got %T %v", obj, res2, res2)
		}
	}
}
