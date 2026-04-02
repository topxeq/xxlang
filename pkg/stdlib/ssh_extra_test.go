// pkg/stdlib/ssh_extra_test.go
// Comprehensive tests for SSH module to increase coverage.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callSSHFunc calls a function from the ssh module.
func callSSHFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("ssh")
	if mod == nil {
		t := &testing.T{}
		t.Skip("ssh module not found")
		return &objects.Error{Message: "ssh module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestSSH_NewClient tests newClient with no arguments.
func TestSSH_NewClient(t *testing.T) {
	result := callSSHFunc("newClient")
	if _, ok := result.(*objects.Error); ok {
		t.Errorf("newClient() expected success, got error: %s", result.Inspect())
	}
}

// TestSSH_Connect_ArgumentValidation tests connect argument validation.
func TestSSH_Connect_ArgumentValidation(t *testing.T) {
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
		result := callSSHFunc("connect", tt.args...)
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

// TestSSH_ConnectWithKey_ArgumentValidation tests connectWithKey argument validation.
func TestSSH_ConnectWithKey_ArgumentValidation(t *testing.T) {
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
		result := callSSHFunc("connectWithKey", tt.args...)
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

// TestSSH_ConnectWithKeyStr_ArgumentValidation tests connectWithKeyStr argument validation.
func TestSSH_ConnectWithKeyStr_ArgumentValidation(t *testing.T) {
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
		result := callSSHFunc("connectWithKeyStr", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("connectWithKeyStr(%v) expected error, got %T", tt.args, result)
			}
		}
	}
}

// TestSSH_ConnectWithConfig_ArgumentValidation tests connectWithConfig argument validation.
func TestSSH_ConnectWithConfig_ArgumentValidation(t *testing.T) {
	// Missing arg
	result := callSSHFunc("connectWithConfig")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing arg should error")
	}

	// Wrong type - not a map
	result = callSSHFunc("connectWithConfig", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig wrong type should error")
	}

	// Map without host
	emptyMap := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	result = callSSHFunc("connectWithConfig", emptyMap)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig without host should error")
	}

	// Map with host but missing user
	partialMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: String("localhost")},
	}}
	result = callSSHFunc("connectWithConfig", partialMap)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("connectWithConfig missing user should error")
	}

	// Valid config with host and user (may fail connection but not arg validation)
	validMap := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		objects.NewString("host").HashKey(): {Key: objects.NewString("host"), Value: String("localhost")},
		objects.NewString("user").HashKey(): {Key: objects.NewString("user"), Value: String("testuser")},
	}}
	result = callSSHFunc("connectWithConfig", validMap)
	_ = result // we don't assert on result, just that no immediate arg error
}

// TestSSH_IsSSHClient_TypeCheck tests isSSHClient with various types.
func TestSSH_IsSSHClient_TypeCheck(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Skip("ssh module not found")
	}
	fn, ok := mod.Exports["isSSHClient"].(*objects.Builtin)
	if !ok {
		t.Fatal("isSSHClient not found or not builtin")
	}

	// Test with SSHClient object
	client := objects.NewSSHClient()
	res := fn.Fn(client)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isSSHClient should return true for SSHClient, got %T %v", res, res)
	}

	// Test with non-SSHClient objects
	nonClientTypes := []objects.Object{
		String("not a client"),
		Int(123),
		Bool(false),
		Null(),
	}
	for _, obj := range nonClientTypes {
		res2 := fn.Fn(obj)
		if b, ok := res2.(*objects.Bool); !ok || b.Value {
			t.Fatalf("isSSHClient should return false for %T, got %T %v", obj, res2, res2)
		}
	}
}

// TestSSH_Exec_ArgumentValidation tests exec argument validation.
func TestSSH_Exec_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("exec")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("exec", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with 3 args should return Error")
	}

	// Wrong types
	result = callSSHFunc("exec", Int(1), Int(22), String("user"), String("pass"), String("cmd"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with non-string host should error")
	}

	result = callSSHFunc("exec", String("host"), String("22"), String("user"), String("pass"), String("cmd"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with non-int port should error")
	}

	result = callSSHFunc("exec", String("host"), Int(22), Int(123), String("pass"), String("cmd"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with non-string user should error")
	}

	result = callSSHFunc("exec", String("host"), Int(22), String("user"), Int(456), String("cmd"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with non-string password should error")
	}

	result = callSSHFunc("exec", String("host"), Int(22), String("user"), String("pass"), Int(789))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("exec with non-string cmd should error")
	}
}

// TestSSH_Upload_ArgumentValidation tests upload argument validation.
func TestSSH_Upload_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("upload")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("upload", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with 3 args should return Error")
	}

	// Wrong types - host
	result = callSSHFunc("upload", Int(1), Int(22), String("user"), String("pass"), String("local"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-string host should error")
	}

	// Wrong type - port
	result = callSSHFunc("upload", String("host"), String("22"), String("user"), String("pass"), String("local"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-int port should error")
	}

	// Wrong type - user
	result = callSSHFunc("upload", String("host"), Int(22), Int(123), String("pass"), String("local"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-string user should error")
	}

	// Wrong type - password
	result = callSSHFunc("upload", String("host"), Int(22), String("user"), Int(456), String("local"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-string password should error")
	}

	// Wrong type - localPath
	result = callSSHFunc("upload", String("host"), Int(22), String("user"), String("pass"), Int(789), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-string localPath should error")
	}

	// Wrong type - remotePath
	result = callSSHFunc("upload", String("host"), Int(22), String("user"), String("pass"), String("local"), Int(101112))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("upload with non-string remotePath should error")
	}
}

// TestSSH_Download_ArgumentValidation tests download argument validation.
func TestSSH_Download_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("download")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("download", String("host"), Int(22), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with 4 args should return Error")
	}

	// Wrong types
	result = callSSHFunc("download", Int(1), Int(22), String("user"), String("pass"), String("remote"), String("local"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-string host should error")
	}

	result = callSSHFunc("download", String("host"), String("22"), String("user"), String("pass"), String("remote"), String("local"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-int port should error")
	}

	result = callSSHFunc("download", String("host"), Int(22), Int(123), String("pass"), String("remote"), String("local"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-string user should error")
	}

	result = callSSHFunc("download", String("host"), Int(22), String("user"), Int(456), String("remote"), String("local"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-string password should error")
	}

	result = callSSHFunc("download", String("host"), Int(22), String("user"), String("pass"), Int(789), String("local"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-string remotePath should error")
	}

	result = callSSHFunc("download", String("host"), Int(22), String("user"), String("pass"), String("remote"), Int(101112))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("download with non-string localPath should error")
	}
}

// TestSSH_UploadBytes_ArgumentValidation tests uploadBytes argument validation.
func TestSSH_UploadBytes_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("uploadBytes")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytes with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("uploadBytes", String("host"), Int(22), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytes with 4 args should return Error")
	}

	// Wrong type for bytes
	result = callSSHFunc("uploadBytes", String("host"), Int(22), String("user"), String("pass"), String("should be bytes"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytes with non-bytes data should error")
	}
}

// TestSSH_DownloadBytes_ArgumentValidation tests downloadBytes argument validation.
func TestSSH_DownloadBytes_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("downloadBytes")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("downloadBytes with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("downloadBytes", String("host"), Int(22), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("downloadBytes with 4 args should return Error")
	}

	// Wrong type for remotePath
	result = callSSHFunc("downloadBytes", Int(1), Int(22), String("user"), String("pass"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("downloadBytes with non-string host should error")
	}
}

// TestSSH_UploadBytesWithKey_ArgumentValidation tests uploadBytesWithKey argument validation.
func TestSSH_UploadBytesWithKey_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("uploadBytesWithKey")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytesWithKey with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("uploadBytesWithKey", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytesWithKey with 3 args should return Error")
	}

	// Wrong type for bytes
	result = callSSHFunc("uploadBytesWithKey", String("host"), Int(22), String("user"), String("key"), String("should be bytes"), String("remote"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("uploadBytesWithKey with non-bytes data should error")
	}
}

// TestSSH_DownloadBytesWithKey_ArgumentValidation tests downloadBytesWithKey argument validation.
func TestSSH_DownloadBytesWithKey_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("downloadBytesWithKey")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("downloadBytesWithKey with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("downloadBytesWithKey", String("host"), Int(22), String("user"), String("key"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("downloadBytesWithKey with 4 args should return Error")
	}
}

// TestSSH_TestConnection_ArgumentValidation tests testConnection argument validation.
func TestSSH_TestConnection_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("testConnection")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("testConnection", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with 3 args should return Error")
	}

	// Wrong types
	result = callSSHFunc("testConnection", Int(1), Int(22), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with non-string host should error")
	}

	result = callSSHFunc("testConnection", String("host"), String("22"), String("user"), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with non-int port should error")
	}

	result = callSSHFunc("testConnection", String("host"), Int(22), Int(123), String("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with non-string user should error")
	}

	result = callSSHFunc("testConnection", String("host"), Int(22), String("user"), Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnection with non-string password should error")
	}
}

// TestSSH_TestConnectionWithKey_ArgumentValidation tests testConnectionWithKey argument validation.
func TestSSH_TestConnectionWithKey_ArgumentValidation(t *testing.T) {
	// Missing args
	result := callSSHFunc("testConnectionWithKey")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with no args should return Error")
	}

	// Too few args
	result = callSSHFunc("testConnectionWithKey", String("host"), Int(22), String("user"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with 3 args should return Error")
	}

	// Wrong types
	result = callSSHFunc("testConnectionWithKey", Int(1), Int(22), String("user"), String("key"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with non-string host should error")
	}

	result = callSSHFunc("testConnectionWithKey", String("host"), String("22"), String("user"), String("key"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with non-int port should error")
	}

	result = callSSHFunc("testConnectionWithKey", String("host"), Int(22), Int(123), String("key"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with non-string user should error")
	}

	result = callSSHFunc("testConnectionWithKey", String("host"), Int(22), String("user"), Int(456))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("testConnectionWithKey with non-string keyPath should error")
	}
}

// TestSSH_ModuleExports tests that all expected exports exist.
func TestSSH_ModuleExports(t *testing.T) {
	mod := Get("ssh")
	if mod == nil {
		t.Skip("ssh module not found")
	}

	expectedExports := []string{
		"newClient",
		"connect",
		"connectWithKey",
		"connectWithKeyStr",
		"connectWithConfig",
		"isSSHClient",
		"exec",
		"upload",
		"download",
		"testConnection",
		"testConnectionWithKey",
		"uploadBytes",
		"downloadBytes",
		"uploadBytesWithKey",
		"downloadBytesWithKey",
	}

	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in ssh module", name)
		}
	}
}
