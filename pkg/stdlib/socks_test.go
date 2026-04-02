package stdlib

import (
	"net"
	"strconv"
	"testing"
)

// portFromListener is a small helper to extract the actual port from a listener
func portFromListener(t *testing.T, l net.Listener) int {
	t.Helper()
	addr := l.Addr().String()
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to parse listener address %q: %v", addr, err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		t.Fatalf("invalid port in listener address %q: %v", addr, err)
	}
	return port
}

func TestSocksServer_Lifecycle(t *testing.T) {
	t.Parallel()
	s := NewSocksServer()
	if s.IsRunning() {
		t.Fatalf("new server should not be running")
	}
	if err := s.Start(0); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	// ensure running
	if !s.IsRunning() {
		t.Fatalf("server should be running after Start")
	}
	// actual port is determined by OS when using 0
	port := portFromListener(t, s.listener)
	if port <= 0 {
		t.Fatalf("expected a valid ephemeral port, got %d", port)
	}
	// ensure idempotent Stop
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if s.IsRunning() {
		t.Fatalf("server should not be running after Stop")
	}
	// second Stop should be a no-op
	if err := s.Stop(); err != nil {
		t.Fatalf("second Stop should be no-op, got: %v", err)
	}
}

func TestSocksServer_AlreadyRunning(t *testing.T) {
	t.Parallel()
	s := NewSocksServer()
	if err := s.Start(0); err != nil {
		t.Fatalf("unexpected Start error: %v", err)
	}
	defer s.Stop()
	// starting again should fail
	if err := s.Start(0); err == nil {
		t.Fatalf("expected error when starting an already running server")
	}
}

func TestProxyServer_LifecycleAndErrors(t *testing.T) {
	t.Parallel()
	// error: empty listen address
	ps := NewProxyServer()
	if err := ps.Start("", "secret", false); err == nil {
		t.Fatalf("expected error for empty listen address")
	}
	// error: empty password
	ps2 := NewProxyServer()
	if err := ps2.Start("127.0.0.1:0", "", false); err == nil {
		t.Fatalf("expected error for empty password")
	}
	// valid start on ephemeral port
	if err := ps.Start("127.0.0.1:0", "secret", false); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ps.Stop()
	if !ps.IsRunning() {
		t.Fatalf("proxy server should be running")
	}
	port := portFromListener(t, ps.listener)
	if port <= 0 {
		t.Fatalf("expected valid port, got %d", port)
	}
	// idempotent Stop
	if err := ps.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if ps.IsRunning() {
		t.Fatalf("proxy server should be stopped")
	}
	// start again and create a client
	if err := ps.Start("127.0.0.1:0", "secret", false); err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
	defer ps.Stop()
	actualPort := ps.listener.Addr().(*net.TCPAddr).Port
	actualServerAddr := "127.0.0.1:" + strconv.Itoa(actualPort)
	// create a client and start
	pc := NewProxyClient()
	if err := pc.Start("127.0.0.1:0", actualServerAddr, "secret", false); err != nil {
		t.Fatalf("proxy client Start failed: %v", err)
	}
	defer pc.Stop()
	if !pc.IsRunning() {
		t.Fatalf("proxy client should be running")
	}
}

func TestSocksClient_BasicFailures(t *testing.T) {
	t.Parallel()
	// create a client and connect to an obviously non-listening proxy
	cc := NewSocksClient()
	if err := cc.Connect("127.0.0.1:9", "example.com:80", true); err == nil {
		t.Fatalf("expected connection error when proxy is not listening")
	}
	if err := cc.ConnectWithAuth("127.0.0.1:9", "example.com:80", "user", "pass"); err == nil {
		t.Fatalf("expected connection error for authenticated connect")
	}
	if cc.IsConnected() {
		t.Fatalf("client should not be connected")
	}
	if err := cc.Close(); err != nil {
		t.Fatalf("closing non-connected should be no-op: %v", err)
	}
}

func TestProxyClient_BasicLifecycle(t *testing.T) {
	t.Parallel()
	// startup a proxy server to connect to
	ps := NewProxyServer()
	if err := ps.Start("127.0.0.1:0", "secret", false); err != nil {
		t.Fatalf("proxy server start failed: %v", err)
	}
	defer ps.Stop()
	actualPort := ps.listener.Addr().(*net.TCPAddr).Port
	serverAddr := "127.0.0.1:" + strconv.Itoa(actualPort)
	pc := NewProxyClient()
	if err := pc.Start("127.0.0.1:0", serverAddr, "secret", false); err != nil {
		t.Fatalf("proxy client Start failed: %v", err)
	}
	defer pc.Stop()
	if !pc.IsRunning() {
		t.Fatalf("proxy client should be running")
	}
	if pc.GetLocalAddr() == "" || pc.GetServerAddr() == "" {
		t.Fatalf("client should have addresses set")
	}
}
