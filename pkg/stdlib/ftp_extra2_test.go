// pkg/stdlib/ftp_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ftpCall invokes a builtin from the ftp module.
func ftpCall(name string, args ...objects.Object) objects.Object {
	mod := Get("ftp")
	if mod == nil {
		return &objects.Error{Message: "ftp module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestFTP_Extra2_Init tests that ftp module registers all exports.
func TestFTP_Extra2_Init(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}
	expected := []string{
		"newClient", "connect", "connectWithConfig",
		"createServer", "isFtpClient", "isFtpServer",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in ftp module", name)
		}
	}
}

// TestFTP_Extra2_NewClient tests newClient.
func TestFTP_Extra2_NewClient(t *testing.T) {
	// newClient takes no args
	res := ftpCall("newClient")
	if _, ok := res.(*objects.FtpClient); !ok {
		t.Fatalf("newClient() should return FtpClient, got %s", res.Type())
	}

	// newClient with args should error
	res = ftpCall("newClient", String("extra"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with args should error")
	}
}

// TestFTP_Extra2_Connect_ArgumentValidation tests connect argument validation.
func TestFTP_Extra2_Connect_ArgumentValidation(t *testing.T) {
	// No args
	res := ftpCall("connect")
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with no args should error")
	}

	// Wrong number of args
	res = ftpCall("connect", String("host"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with 1 arg should error")
	}

	// Wrong host type
	res = ftpCall("connect", Int(123), Int(21), String("user"), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with int host should error")
	}

	// Wrong port type
	res = ftpCall("connect", String("host"), String("21"), String("user"), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with string port should error")
	}

	// Wrong user type
	res = ftpCall("connect", String("host"), Int(21), Int(123), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with int user should error")
	}

	// Wrong password type
	res = ftpCall("connect", String("host"), Int(21), String("user"), Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connect() with int password should error")
	}

	// Valid args (will fail to connect, but tests argument parsing)
	res = ftpCall("connect", String("localhost"), Int(21), String("user"), String("pass"))
	// This will return an error because connection fails, but that's expected
	_ = res
}

// TestFTP_Extra2_ConnectWithConfig_ArgumentValidation tests connectWithConfig.
func TestFTP_Extra2_ConnectWithConfig_ArgumentValidation(t *testing.T) {
	// No args
	res := ftpCall("connectWithConfig")
	if res.Type() != objects.ErrorType {
		t.Fatalf("connectWithConfig() with no args should error")
	}

	// Wrong type
	res = ftpCall("connectWithConfig", String("not a map"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("connectWithConfig() with string should error")
	}

	// With config map (will fail to connect, but tests argument parsing)
	config := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("host").HashKey():     {Key: String("host"), Value: String("localhost")},
		String("port").HashKey():     {Key: String("port"), Value: Int(21)},
		String("user").HashKey():     {Key: String("user"), Value: String("user")},
		String("password").HashKey(): {Key: String("password"), Value: String("pass")},
	}}
	res = ftpCall("connectWithConfig", config)
	// This will return an error because connection fails, but that's expected
	_ = res
}

// TestFTP_Extra2_CreateServer_ArgumentValidation tests createServer.
func TestFTP_Extra2_CreateServer_ArgumentValidation(t *testing.T) {
	// No args
	res := ftpCall("createServer")
	if res.Type() != objects.ErrorType {
		t.Fatalf("createServer() with no args should error")
	}

	// Wrong type for addr
	res = ftpCall("createServer", Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("createServer() with int addr should error")
	}

	// With string addr (may fail to create, but tests argument parsing)
	res = ftpCall("createServer", String(":2121"))
	// May return FtpServer or error
	_ = res
}

// TestFTP_Extra2_IsFtpClient tests isFtpClient.
func TestFTP_Extra2_IsFtpClient(t *testing.T) {
	// No args
	res := ftpCall("isFtpClient")
	if res.Type() != objects.ErrorType {
		t.Fatalf("isFtpClient() with no args should error")
	}

	// With FtpClient
	client := ftpCall("newClient")
	res = ftpCall("isFtpClient", client)
	if res.Type() != objects.BoolType {
		t.Fatalf("isFtpClient() should return bool, got %s", res.Type())
	}
	if !res.(*objects.Bool).Value {
		t.Fatalf("isFtpClient(FtpClient) should return true")
	}

	// With non-FtpClient
	res = ftpCall("isFtpClient", String("not a client"))
	if res.Type() != objects.BoolType {
		t.Fatalf("isFtpClient() should return bool, got %s", res.Type())
	}
	if res.(*objects.Bool).Value {
		t.Fatalf("isFtpClient(string) should return false")
	}
}

// TestFTP_Extra2_IsFtpServer tests isFtpServer.
func TestFTP_Extra2_IsFtpServer(t *testing.T) {
	// No args
	res := ftpCall("isFtpServer")
	if res.Type() != objects.ErrorType {
		t.Fatalf("isFtpServer() with no args should error")
	}

	// With non-FtpServer
	res = ftpCall("isFtpServer", String("not a server"))
	if res.Type() != objects.BoolType {
		t.Fatalf("isFtpServer() should return bool, got %s", res.Type())
	}
	if res.(*objects.Bool).Value {
		t.Fatalf("isFtpServer(string) should return false")
	}
}
