// pkg/stdlib/socks_extra_test.go
// Additional tests for socks module builtin functions: createServer, startServer, isSocksServer, createClient, connect, connectWithAuth, isSocksClient, createProxyServer, startProxyServer, isProxyServer, createProxyClient, startProxyClient, isProxyClient
package stdlib

import (
	"net"
	"strconv"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callSocksFunc calls a function from the socks module.
func callSocksFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("socks")
	if mod == nil {
		return &objects.Error{Message: "socks module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// isSocksErr checks if an object is an error.
func isSocksErr(obj objects.Object) bool {
	_, ok := obj.(*objects.Error)
	return ok
}

// ============================================================
// SOCKS Server Builtin Tests
// ============================================================

// TestSocksCreateServer tests the "createServer" builtin.
func TestSocksCreateServer(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createServer")
	if isSocksErr(result) {
		t.Fatalf("createServer() failed: %v", result.Inspect())
	}
	_, ok := result.(*SocksServer)
	if !ok {
		t.Fatalf("createServer() should return SocksServer, got %T", result)
	}
}

// TestSocksCreateServer_Error tests createServer error handling.
func TestSocksCreateServer_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createServer", Int(123))
	if !isSocksErr(result) {
		t.Error("createServer with arguments should error")
	}
}

// TestSocksStartServer tests the "startServer" builtin.
func TestSocksStartServer(t *testing.T) {
	t.Parallel()
	server := callSocksFunc("createServer").(*SocksServer)

	result := callSocksFunc("startServer", Int(0), server)
	if isSocksErr(result) {
		t.Fatalf("startServer failed: %v", result.Inspect())
	}

	srv := result.(*SocksServer)
	if !srv.IsRunning() {
		t.Fatalf("server should be running after startServer")
	}

	// Verify actual bound port via listener
	if srv.listener == nil {
		t.Fatal("listener is nil")
	}
	_, p, err := net.SplitHostPort(srv.listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse listener address: %v", err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		t.Fatalf("invalid port: %v", err)
	}
	if port <= 0 {
		t.Fatalf("expected valid port, got %d", port)
	}

	srv.Stop()
}

// TestSocksStartServer_WithOptions tests startServer with SOCKS4 option.
func TestSocksStartServer_WithOptions(t *testing.T) {
	t.Parallel()
	server := callSocksFunc("createServer").(*SocksServer)

	result := callSocksFunc("startServer", Int(0), String("-socks4"), server)
	if isSocksErr(result) {
		t.Fatalf("startServer with -socks4 failed: %v", result.Inspect())
	}

	srv := result.(*SocksServer)
	if srv.IsRunning() {
		srv.Stop()
	}
}

// TestSocksStartServer_Error tests startServer error handling.
func TestSocksStartServer_Error(t *testing.T) {
	t.Parallel()

	// no arguments at all
	result := callSocksFunc("startServer")
	if !isSocksErr(result) {
		t.Error("startServer with no args should error")
	}

	// first arg not int
	result = callSocksFunc("startServer", String("not an int"))
	if !isSocksErr(result) {
		t.Error("startServer with non-int port should error")
	}

	// passing extra non-string options is allowed but ignored; no error expected
}

// TestSocksIsSocksServer tests the "isSocksServer" builtin.
func TestSocksIsSocksServer(t *testing.T) {
	t.Parallel()
	server := callSocksFunc("createServer").(*SocksServer)

	result := callSocksFunc("isSocksServer", server)
	if isSocksErr(result) {
		t.Fatalf("isSocksServer failed: %v", result.Inspect())
	}

	b, ok := result.(*objects.Bool)
	if !ok || !b.Value {
		t.Fatalf("isSocksServer(server) should return true, got %v", b)
	}

	// with non-SocksServer
	result = callSocksFunc("isSocksServer", String("not a server"))
	if isSocksErr(result) {
		t.Fatalf("isSocksServer with wrong type failed: %v", result.Inspect())
	}
	b, ok = result.(*objects.Bool)
	if !ok || b.Value {
		t.Fatalf("isSocksServer(non-server) should return false, got %v", b)
	}
}

// TestSocksIsSocksServer_Error tests isSocksServer error handling.
func TestSocksIsSocksServer_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("isSocksServer")
	if !isSocksErr(result) {
		t.Error("isSocksServer with no args should error")
	}
	result = callSocksFunc("isSocksServer", String("arg1"), String("arg2"))
	if !isSocksErr(result) {
		t.Error("isSocksServer with >1 args should error")
	}
}

// ============================================================
// SOCKS Client Builtin Tests
// ============================================================

// TestSocksCreateClient tests the "createClient" builtin.
func TestSocksCreateClient(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createClient")
	if isSocksErr(result) {
		t.Fatalf("createClient() failed: %v", result.Inspect())
	}
	_, ok := result.(*SocksClient)
	if !ok {
		t.Fatalf("createClient() should return SocksClient, got %T", result)
	}
}

// TestSocksCreateClient_Error tests createClient error handling.
func TestSocksCreateClient_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createClient", Int(123))
	if !isSocksErr(result) {
		t.Error("createClient with arguments should error")
	}
}

// TestSocksConnect tests the "connect" builtin.
func TestSocksConnect(t *testing.T) {
	t.Parallel()
	client := callSocksFunc("createClient").(*SocksClient)

	// Connecting to non-listening proxy should return an error
	result := callSocksFunc("connect", String("127.0.0.1:1"), String("example.com:80"), client)
	if !isSocksErr(result) {
		t.Fatal("connect to non-listening proxy should error")
	}
}

// TestSocksConnect_SOCKS4 tests connect with SOCKS4 option.
func TestSocksConnect_SOCKS4(t *testing.T) {
	t.Parallel()
	client := callSocksFunc("createClient").(*SocksClient)

	result := callSocksFunc("connect", String("127.0.0.1:1"), String("1.2.3.4:80"), String("-socks4"), client)
	if !isSocksErr(result) {
		t.Fatal("connect with -socks4 to non-listening proxy should error")
	}
}

// TestSocksConnect_Error tests connect error handling.
func TestSocksConnect_Error(t *testing.T) {
	t.Parallel()
	client := callSocksFunc("createClient").(*SocksClient)

	// missing target
	result := callSocksFunc("connect", String("proxy"), client)
	if !isSocksErr(result) {
		t.Error("connect with missing target should error")
	}

	// proxy not string
	result = callSocksFunc("connect", Int(123), String("target"), client)
	if !isSocksErr(result) {
		t.Error("connect with non-string proxy should error")
	}

	// target not string
	result = callSocksFunc("connect", String("proxy"), Int(456), client)
	if !isSocksErr(result) {
		t.Error("connect with non-string target should error")
	}

	// client not SocksClient
	result = callSocksFunc("connect", String("proxy"), String("target"), String("not client"))
	if !isSocksErr(result) {
		t.Error("connect with non-SocksClient arg should error")
	}
}

// TestSocksConnectWithAuth tests the "connectWithAuth" builtin.
func TestSocksConnectWithAuth(t *testing.T) {
	t.Parallel()

	// connectWithAuth returns a new client; connection will fail because proxy not listening, so expect error
	result := callSocksFunc("connectWithAuth", String("127.0.0.1:1"), String("example.com:80"), String("user"), String("pass"))
	if !isSocksErr(result) {
		t.Fatal("connectWithAuth to non-listening proxy should error")
	}
}

// TestSocksConnectWithAuth_Error tests connectWithAuth error handling.
func TestSocksConnectWithAuth_Error(t *testing.T) {
	t.Parallel()

	// missing args (only 4 args instead of 5? Actually need exactly 4: proxy, target, user, pass)
	// This test covers wrong number of args
	result := callSocksFunc("connectWithAuth", String("proxy"), String("target"), String("user"))
	if !isSocksErr(result) {
		t.Error("connectWithAuth with missing password should error")
	}

	// wrong types
	result = callSocksFunc("connectWithAuth", Int(1), String("target"), String("user"), String("pass"))
	if !isSocksErr(result) {
		t.Error("connectWithAuth with non-string proxy should error")
	}

	result = callSocksFunc("connectWithAuth", String("proxy"), Int(2), String("user"), String("pass"))
	if !isSocksErr(result) {
		t.Error("connectWithAuth with non-string target should error")
	}

	result = callSocksFunc("connectWithAuth", String("proxy"), String("target"), Int(3), String("pass"))
	if !isSocksErr(result) {
		t.Error("connectWithAuth with non-string username should error")
	}

	result = callSocksFunc("connectWithAuth", String("proxy"), String("target"), String("user"), Int(4))
	if !isSocksErr(result) {
		t.Error("connectWithAuth with non-string password should error")
	}
}

// TestSocksIsSocksClient tests the "isSocksClient" builtin.
func TestSocksIsSocksClient(t *testing.T) {
	t.Parallel()
	client := callSocksFunc("createClient").(*SocksClient)

	result := callSocksFunc("isSocksClient", client)
	if isSocksErr(result) {
		t.Fatalf("isSocksClient failed: %v", result.Inspect())
	}

	b, ok := result.(*objects.Bool)
	if !ok || !b.Value {
		t.Fatalf("isSocksClient(client) should return true, got %v", b)
	}

	// with non-SocksClient
	result = callSocksFunc("isSocksClient", String("not a client"))
	if isSocksErr(result) {
		t.Fatalf("isSocksClient with wrong type failed: %v", result.Inspect())
	}
	b, ok = result.(*objects.Bool)
	if !ok || b.Value {
		t.Fatalf("isSocksClient(non-client) should return false, got %v", b)
	}
}

// TestSocksIsSocksClient_Error tests isSocksClient error handling.
func TestSocksIsSocksClient_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("isSocksClient")
	if !isSocksErr(result) {
		t.Error("isSocksClient with no args should error")
	}
	result = callSocksFunc("isSocksClient", String("arg1"), String("arg2"))
	if !isSocksErr(result) {
		t.Error("isSocksClient with >1 args should error")
	}
}

// ============================================================
// Encrypted Proxy Server Builtin Tests
// ============================================================

// TestSocksCreateProxyServer tests the "createProxyServer" builtin.
func TestSocksCreateProxyServer(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createProxyServer")
	if isSocksErr(result) {
		t.Fatalf("createProxyServer() failed: %v", result.Inspect())
	}
	_, ok := result.(*ProxyServer)
	if !ok {
		t.Fatalf("createProxyServer() should return ProxyServer, got %T", result)
	}
}

// TestSocksCreateProxyServer_Error tests createProxyServer error handling.
func TestSocksCreateProxyServer_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createProxyServer", Int(123))
	if !isSocksErr(result) {
		t.Error("createProxyServer with arguments should error")
	}
}

// TestSocksStartProxyServer tests the "startProxyServer" builtin.
func TestSocksStartProxyServer(t *testing.T) {
	t.Parallel()

	result := callSocksFunc("startProxyServer", String("127.0.0.1:0"), String("secret"), Bool(false))
	if isSocksErr(result) {
		t.Fatalf("startProxyServer failed: %v", result.Inspect())
	}

	srv := result.(*ProxyServer)
	if !srv.IsRunning() {
		t.Fatalf("proxy server should be running after startProxyServer")
	}
	if srv.GetListenAddr() == "" {
		t.Fatalf("proxy server should have listen address")
	}

	srv.Stop()
}

// TestSocksStartProxyServer_Error tests startProxyServer error handling.
func TestSocksStartProxyServer_Error(t *testing.T) {
	t.Parallel()

	// not enough args
	result := callSocksFunc("startProxyServer", String("127.0.0.1:0"))
	if !isSocksErr(result) {
		t.Error("startProxyServer with missing password should error")
	}

	// empty listen address
	result = callSocksFunc("startProxyServer", String(""), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyServer with empty listen address should error")
	}

	// empty password
	result = callSocksFunc("startProxyServer", String("127.0.0.1:0"), String(""))
	if !isSocksErr(result) {
		t.Error("startProxyServer with empty password should error")
	}

	// wrong types
	result = callSocksFunc("startProxyServer", Int(123), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyServer with non-string listenAddr should error")
	}

	result = callSocksFunc("startProxyServer", String("127.0.0.1:0"), Int(456))
	if !isSocksErr(result) {
		t.Error("startProxyServer with non-string password should error")
	}
}

// TestSocksIsProxyServer tests the "isProxyServer" builtin.
func TestSocksIsProxyServer(t *testing.T) {
	t.Parallel()
	server := callSocksFunc("createProxyServer").(*ProxyServer)

	result := callSocksFunc("isProxyServer", server)
	if isSocksErr(result) {
		t.Fatalf("isProxyServer failed: %v", result.Inspect())
	}

	b, ok := result.(*objects.Bool)
	if !ok || !b.Value {
		t.Fatalf("isProxyServer(server) should return true, got %v", b)
	}

	// with non-ProxyServer
	result = callSocksFunc("isProxyServer", String("not a server"))
	if isSocksErr(result) {
		t.Fatalf("isProxyServer with wrong type failed: %v", result.Inspect())
	}
	b, ok = result.(*objects.Bool)
	if !ok || b.Value {
		t.Fatalf("isProxyServer(non-server) should return false, got %v", b)
	}
}

// TestSocksIsProxyServer_Error tests isProxyServer error handling.
func TestSocksIsProxyServer_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("isProxyServer")
	if !isSocksErr(result) {
		t.Error("isProxyServer with no args should error")
	}
	result = callSocksFunc("isProxyServer", String("arg1"), String("arg2"))
	if !isSocksErr(result) {
		t.Error("isProxyServer with >1 args should error")
	}
}

// ============================================================
// Encrypted Proxy Client Builtin Tests
// ============================================================

// TestSocksCreateProxyClient tests the "createProxyClient" builtin.
func TestSocksCreateProxyClient(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createProxyClient")
	if isSocksErr(result) {
		t.Fatalf("createProxyClient() failed: %v", result.Inspect())
	}
	_, ok := result.(*ProxyClient)
	if !ok {
		t.Fatalf("createProxyClient() should return ProxyClient, got %T", result)
	}
}

// TestSocksCreateProxyClient_Error tests createProxyClient error handling.
func TestSocksCreateProxyClient_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("createProxyClient", Int(123))
	if !isSocksErr(result) {
		t.Error("createProxyClient with arguments should error")
	}
}

// TestSocksStartProxyClient tests the "startProxyClient" builtin.
func TestSocksStartProxyClient(t *testing.T) {
	t.Parallel()
	// startProxyClient creates a new client and starts it; server is not reachable but that doesn't cause Start to fail.
	result := callSocksFunc("startProxyClient", String("127.0.0.1:0"), String("127.0.0.1:1"), String("secret"), Bool(false))
	if isSocksErr(result) {
		t.Fatalf("startProxyClient failed: %v", result.Inspect())
	}

	cl := result.(*ProxyClient)
	if !cl.IsRunning() {
		t.Fatalf("proxy client should be running after startProxyClient")
	}
	if cl.GetLocalAddr() == "" || cl.GetServerAddr() == "" {
		t.Fatalf("proxy client should have addresses set")
	}

	cl.Stop()
}

// TestSocksStartProxyClient_Error tests startProxyClient error handling.
func TestSocksStartProxyClient_Error(t *testing.T) {
	t.Parallel()

	// not enough args (missing password)
	result := callSocksFunc("startProxyClient", String("127.0.0.1:0"), String("127.0.0.1:1"))
	if !isSocksErr(result) {
		t.Error("startProxyClient with missing password should error")
	}

	// empty local address
	result = callSocksFunc("startProxyClient", String(""), String("127.0.0.1:1"), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyClient with empty local address should error")
	}

	// empty server address
	result = callSocksFunc("startProxyClient", String("127.0.0.1:0"), String(""), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyClient with empty server address should error")
	}

	// empty password
	result = callSocksFunc("startProxyClient", String("127.0.0.1:0"), String("127.0.0.1:1"), String(""))
	if !isSocksErr(result) {
		t.Error("startProxyClient with empty password should error")
	}

	// wrong types
	result = callSocksFunc("startProxyClient", Int(123), String("127.0.0.1:1"), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyClient with non-string localAddr should error")
	}

	result = callSocksFunc("startProxyClient", String("127.0.0.1:0"), Int(456), String("secret"))
	if !isSocksErr(result) {
		t.Error("startProxyClient with non-string serverAddr should error")
	}

	result = callSocksFunc("startProxyClient", String("127.0.0.1:0"), String("127.0.0.1:1"), Int(789))
	if !isSocksErr(result) {
		t.Error("startProxyClient with non-string password should error")
	}
}

// TestSocksIsProxyClient tests the "isProxyClient" builtin.
func TestSocksIsProxyClient(t *testing.T) {
	t.Parallel()
	client := callSocksFunc("createProxyClient").(*ProxyClient)

	result := callSocksFunc("isProxyClient", client)
	if isSocksErr(result) {
		t.Fatalf("isProxyClient failed: %v", result.Inspect())
	}

	b, ok := result.(*objects.Bool)
	if !ok || !b.Value {
		t.Fatalf("isProxyClient(client) should return true, got %v", b)
	}

	// with non-ProxyClient
	result = callSocksFunc("isProxyClient", String("not a client"))
	if isSocksErr(result) {
		t.Fatalf("isProxyClient with wrong type failed: %v", result.Inspect())
	}
	b, ok = result.(*objects.Bool)
	if !ok || b.Value {
		t.Fatalf("isProxyClient(non-client) should return false, got %v", b)
	}
}

// TestSocksIsProxyClient_Error tests isProxyClient error handling.
func TestSocksIsProxyClient_Error(t *testing.T) {
	t.Parallel()
	result := callSocksFunc("isProxyClient")
	if !isSocksErr(result) {
		t.Error("isProxyClient with no args should error")
	}
	result = callSocksFunc("isProxyClient", String("arg1"), String("arg2"))
	if !isSocksErr(result) {
		t.Error("isProxyClient with >1 args should error")
	}
}
