// pkg/stdlib/socks.go
// SOCKS and encrypted proxy module for Xxlang.
// Provides SOCKS4/SOCKS5 proxy and encrypted tunnel proxy functionality.
package stdlib

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================================
// SOCKS Proxy Server/Client
// ============================================================

// SocksServer represents a SOCKS proxy server.
type SocksServer struct {
	mu       sync.Mutex
	listener net.Listener
	running  bool
	port     int
	socks5   bool
	auth     bool
	username string
	password string
}

// Type returns the object type.
func (s *SocksServer) Type() objects.ObjectType { return objects.ObjectType("SOCKS_SERVER") }

// TypeTag returns the fast type tag.
func (s *SocksServer) TypeTag() objects.TypeTag { return objects.TagUnknown }

// Inspect returns a string representation.
func (s *SocksServer) Inspect() string {
	if s.running {
		return fmt.Sprintf("SocksServer(running=true, port=%d, socks5=%v)", s.port, s.socks5)
	}
	return "SocksServer(running=false)"
}

// ToBool returns true if running.
func (s *SocksServer) ToBool() *objects.Bool { return &objects.Bool{Value: s.running} }

// HashKey returns a hash key.
func (s *SocksServer) HashKey() objects.HashKey {
	return objects.HashKey{
		Type:  objects.ObjectType("SOCKS_SERVER"),
		Value: uint64(s.port),
	}
}

// SocksClient represents a SOCKS proxy client connection.
type SocksClient struct {
	mu        sync.Mutex
	conn      net.Conn
	connected bool
	proxyAddr string
	target    string
	socks5    bool
}

// Type returns the object type.
func (c *SocksClient) Type() objects.ObjectType { return objects.ObjectType("SOCKS_CLIENT") }

// TypeTag returns the fast type tag.
func (c *SocksClient) TypeTag() objects.TypeTag { return objects.TagUnknown }

// Inspect returns a string representation.
func (c *SocksClient) Inspect() string {
	if c.connected {
		return fmt.Sprintf("SocksClient(connected=true, proxy=%s, target=%s)", c.proxyAddr, c.target)
	}
	return "SocksClient(connected=false)"
}

// ToBool returns true if connected.
func (c *SocksClient) ToBool() *objects.Bool { return &objects.Bool{Value: c.connected} }

// HashKey returns a hash key.
func (c *SocksClient) HashKey() objects.HashKey {
	return objects.HashKey{
		Type:  objects.ObjectType("SOCKS_CLIENT"),
		Value: 0,
	}
}

// NewSocksServer creates a new SOCKS server.
func NewSocksServer() *SocksServer {
	return &SocksServer{
		socks5: true,
	}
}

// NewSocksClient creates a new SOCKS client.
func NewSocksClient() *SocksClient {
	return &SocksClient{
		socks5: true,
	}
}

// Start starts the SOCKS server on the specified port.
func (s *SocksServer) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server already running")
	}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start SOCKS server: %w", err)
	}

	s.listener = listener
	s.port = port
	s.running = true

	go s.socksAcceptLoop()

	return nil
}

// Stop stops the SOCKS server.
func (s *SocksServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	err := s.listener.Close()
	s.running = false
	return err
}

// IsRunning returns whether the server is running.
func (s *SocksServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetPort returns the server port.
func (s *SocksServer) GetPort() int {
	return s.port
}

// SetSocks5 sets whether to use SOCKS5.
func (s *SocksServer) SetSocks5(socks5 bool) {
	s.socks5 = socks5
}

// SetAuth sets username/password authentication (SOCKS5 only).
func (s *SocksServer) SetAuth(username, password string) {
	s.auth = true
	s.username = username
	s.password = password
}

func (s *SocksServer) socksAcceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleSocksConnection(conn)
	}
}

func (s *SocksServer) handleSocksConnection(conn net.Conn) {
	defer conn.Close()

	if s.socks5 {
		s.handleSocks5(conn)
	} else {
		s.handleSocks4(conn)
	}
}

func (s *SocksServer) handleSocks5(conn net.Conn) {
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	if n < 2 || buf[0] != 0x05 {
		return
	}

	numMethods := int(buf[1])
	hasNoAuth := false
	hasUserPass := false

	for i := 0; i < numMethods && i+2 < n; i++ {
		switch buf[2+i] {
		case 0x00:
			hasNoAuth = true
		case 0x02:
			hasUserPass = true
		}
	}

	var method byte
	if s.auth && hasUserPass {
		method = 0x02
	} else if hasNoAuth {
		method = 0x00
	} else {
		conn.Write([]byte{0x05, 0xFF})
		return
	}

	conn.Write([]byte{0x05, method})

	if method == 0x02 {
		n, err = conn.Read(buf)
		if err != nil || n < 2 {
			return
		}

		ulen := int(buf[1])
		if n < 2+ulen+1 {
			return
		}
		username := string(buf[2 : 2+ulen])

		plen := int(buf[2+ulen])
		if n < 2+ulen+1+plen {
			return
		}
		password := string(buf[3+ulen : 3+ulen+plen])

		if username != s.username || password != s.password {
			conn.Write([]byte{0x01, 0x01})
			return
		}

		conn.Write([]byte{0x01, 0x00})
	}

	n, err = conn.Read(buf)
	if err != nil || n < 10 || buf[0] != 0x05 || buf[1] != 0x01 {
		return
	}

	var targetAddr string
	var targetPort int

	switch buf[3] {
	case 0x01:
		if n < 10 {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		targetPort = int(binary.BigEndian.Uint16(buf[8:10]))
	case 0x03:
		dlen := int(buf[4])
		if n < 5+dlen+2 {
			return
		}
		targetAddr = string(buf[5 : 5+dlen])
		targetPort = int(binary.BigEndian.Uint16(buf[5+dlen : 7+dlen]))
	case 0x04:
		if n < 22 {
			return
		}
		targetAddr = fmt.Sprintf("[%x:%x:%x:%x:%x:%x:%x:%x]",
			binary.BigEndian.Uint16(buf[4:6]),
			binary.BigEndian.Uint16(buf[6:8]),
			binary.BigEndian.Uint16(buf[8:10]),
			binary.BigEndian.Uint16(buf[10:12]),
			binary.BigEndian.Uint16(buf[12:14]),
			binary.BigEndian.Uint16(buf[14:16]),
			binary.BigEndian.Uint16(buf[16:18]),
			binary.BigEndian.Uint16(buf[18:20]))
		targetPort = int(binary.BigEndian.Uint16(buf[20:22]))
	default:
		return
	}

	target := fmt.Sprintf("%s:%d", targetAddr, targetPort)
	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}

func (s *SocksServer) handleSocks4(conn net.Conn) {
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil || n < 9 || buf[0] != 0x04 {
		return
	}

	if buf[1] != 0x01 {
		conn.Write([]byte{0x00, 0x5B, 0, 0, 0, 0, 0, 0})
		return
	}

	targetPort := int(binary.BigEndian.Uint16(buf[2:4]))
	targetAddr := fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])

	target := fmt.Sprintf("%s:%d", targetAddr, targetPort)
	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		conn.Write([]byte{0x00, 0x5B, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	conn.Write([]byte{0x00, 0x5A, 0, 0, 0, 0, 0, 0})

	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}

// Connect connects to a target through a SOCKS proxy.
func (c *SocksClient) Connect(proxyAddr, target string, socks5 bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to proxy: %w", err)
	}

	c.conn = conn
	c.proxyAddr = proxyAddr
	c.target = target
	c.socks5 = socks5

	if socks5 {
		err = c.socks5Handshake("", "")
	} else {
		err = c.socks4Handshake()
	}

	if err != nil {
		conn.Close()
		return err
	}

	c.connected = true
	return nil
}

// ConnectWithAuth connects with username/password authentication (SOCKS5).
func (c *SocksClient) ConnectWithAuth(proxyAddr, target, username, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to proxy: %w", err)
	}

	c.conn = conn
	c.proxyAddr = proxyAddr
	c.target = target
	c.socks5 = true

	err = c.socks5Handshake(username, password)
	if err != nil {
		conn.Close()
		return err
	}

	c.connected = true
	return nil
}

func (c *SocksClient) socks5Handshake(username, password string) error {
	methods := []byte{0x05, 0x01, 0x00}
	if username != "" {
		methods = []byte{0x05, 0x02, 0x00, 0x02}
	}

	if _, err := c.conn.Write(methods); err != nil {
		return err
	}

	buf := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		return err
	}

	if buf[0] != 0x05 {
		return errors.New("invalid SOCKS version")
	}

	if buf[1] == 0x02 {
		if username == "" {
			return errors.New("authentication required but not provided")
		}

		authBuf := make([]byte, 3+len(username)+len(password))
		authBuf[0] = 0x01
		authBuf[1] = byte(len(username))
		copy(authBuf[2:], username)
		authBuf[2+len(username)] = byte(len(password))
		copy(authBuf[3+len(username):], password)

		if _, err := c.conn.Write(authBuf); err != nil {
			return err
		}

		resp := make([]byte, 2)
		if _, err := io.ReadFull(c.conn, resp); err != nil {
			return err
		}

		if resp[1] != 0x00 {
			return errors.New("authentication failed")
		}
	} else if buf[1] != 0x00 {
		return errors.New("no acceptable authentication method")
	}

	host, portStr, err := net.SplitHostPort(c.target)
	if err != nil {
		return err
	}
	port, _ := strconv.Atoi(portStr)

	req := make([]byte, 0, 256)
	req = append(req, 0x05, 0x01, 0x00)

	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = append(req, 0x01)
			req = append(req, ip4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip...)
		}
	} else {
		req = append(req, 0x03, byte(len(host)))
		req = append(req, host...)
	}

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, uint16(port))
	req = append(req, portBuf...)

	if _, err := c.conn.Write(req); err != nil {
		return err
	}

	resp := make([]byte, 256)
	n, err := c.conn.Read(resp)
	if err != nil {
		return err
	}

	if n < 2 || resp[0] != 0x05 || resp[1] != 0x00 {
		return errors.New("connect failed")
	}

	return nil
}

func (c *SocksClient) socks4Handshake() error {
	host, portStr, err := net.SplitHostPort(c.target)
	if err != nil {
		return err
	}
	port, _ := strconv.Atoi(portStr)

	ip := net.ParseIP(host)
	if ip == nil {
		return errors.New("SOCKS4 requires IP address")
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return errors.New("SOCKS4 only supports IPv4")
	}

	req := make([]byte, 9)
	req[0] = 0x04
	req[1] = 0x01
	binary.BigEndian.PutUint16(req[2:4], uint16(port))
	copy(req[4:8], ip4)
	req[8] = 0x00

	if _, err := c.conn.Write(req); err != nil {
		return err
	}

	resp := make([]byte, 8)
	if _, err := io.ReadFull(c.conn, resp); err != nil {
		return err
	}

	if resp[1] != 0x5A {
		return errors.New("SOCKS4 connect failed")
	}

	return nil
}

// Close closes the connection.
func (c *SocksClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	return c.conn.Close()
}

// IsConnected returns whether the client is connected.
func (c *SocksClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Write writes data to the connection.
func (c *SocksClient) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return 0, errors.New("not connected")
	}

	return c.conn.Write(data)
}

// Read reads data from the connection.
func (c *SocksClient) Read(buf []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return 0, errors.New("not connected")
	}

	return c.conn.Read(buf)
}

// ============================================================
// Encrypted Proxy Server/Client (goconnectit style)
// ============================================================

// ProxyServer represents an encrypted proxy server.
type ProxyServer struct {
	mu          sync.Mutex
	listenAddr  string
	password    string
	verbose     bool
	listener    net.Listener
	connections map[net.Conn]struct{}
	connMutex   sync.RWMutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	startTime   time.Time
}

// Type returns the object type.
func (s *ProxyServer) Type() objects.ObjectType { return objects.ObjectType("PROXY_SERVER") }

// TypeTag returns the fast type tag.
func (s *ProxyServer) TypeTag() objects.TypeTag { return objects.TagUnknown }

// Inspect returns a string representation.
func (s *ProxyServer) Inspect() string {
	if s.running {
		return fmt.Sprintf("ProxyServer(running=true, addr=%s)", s.listenAddr)
	}
	return "ProxyServer(running=false)"
}

// ToBool returns true if running.
func (s *ProxyServer) ToBool() *objects.Bool { return &objects.Bool{Value: s.running} }

// HashKey returns a hash key.
func (s *ProxyServer) HashKey() objects.HashKey {
	return objects.HashKey{
		Type:  objects.ObjectType("PROXY_SERVER"),
		Value: uint64(s.startTime.Unix()),
	}
}

// ProxyClient represents an encrypted proxy client.
type ProxyClient struct {
	mu          sync.Mutex
	localAddr   string
	serverAddr  string
	password    string
	verbose     bool
	listener    net.Listener
	connections map[net.Conn]struct{}
	connMutex   sync.RWMutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	startTime   time.Time
}

// Type returns the object type.
func (c *ProxyClient) Type() objects.ObjectType { return objects.ObjectType("PROXY_CLIENT") }

// TypeTag returns the fast type tag.
func (c *ProxyClient) TypeTag() objects.TypeTag { return objects.TagUnknown }

// Inspect returns a string representation.
func (c *ProxyClient) Inspect() string {
	if c.running {
		return fmt.Sprintf("ProxyClient(running=true, local=%s, server=%s)", c.localAddr, c.serverAddr)
	}
	return "ProxyClient(running=false)"
}

// ToBool returns true if running.
func (c *ProxyClient) ToBool() *objects.Bool { return &objects.Bool{Value: c.running} }

// HashKey returns a hash key.
func (c *ProxyClient) HashKey() objects.HashKey {
	return objects.HashKey{
		Type:  objects.ObjectType("PROXY_CLIENT"),
		Value: uint64(c.startTime.Unix()),
	}
}

// EncryptedConn wraps a net.Conn with encryption.
type EncryptedConn struct {
	conn        net.Conn
	encryptor   cipher.Stream
	decryptor   cipher.Stream
	writeBuffer []byte
}

// deriveKey derives a 32-byte key from password using SHA-256.
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// NewEncryptedConn creates a new encrypted connection.
func NewEncryptedConn(conn net.Conn, password string, isServer bool) (*EncryptedConn, error) {
	key := deriveKey(password)

	encryptIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(encryptIV); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	var decryptIV []byte
	if isServer {
		decryptIV = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(conn, decryptIV); err != nil {
			return nil, fmt.Errorf("failed to receive IV: %w", err)
		}
		if _, err := conn.Write(encryptIV); err != nil {
			return nil, fmt.Errorf("failed to send IV: %w", err)
		}
	} else {
		if _, err := conn.Write(encryptIV); err != nil {
			return nil, fmt.Errorf("failed to send IV: %w", err)
		}
		decryptIV = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(conn, decryptIV); err != nil {
			return nil, fmt.Errorf("failed to receive IV: %w", err)
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	encryptor := cipher.NewCTR(block, encryptIV)
	decryptor := cipher.NewCTR(block, decryptIV)

	return &EncryptedConn{
		conn:        conn,
		encryptor:   encryptor,
		decryptor:   decryptor,
		writeBuffer: make([]byte, 4096),
	}, nil
}

// Read reads and decrypts data from the connection.
func (ec *EncryptedConn) Read(b []byte) (int, error) {
	n, err := ec.conn.Read(b)
	if err != nil {
		return n, err
	}
	ec.decryptor.XORKeyStream(b[:n], b[:n])
	return n, nil
}

// Write encrypts and writes data to the connection.
func (ec *EncryptedConn) Write(b []byte) (int, error) {
	if len(ec.writeBuffer) < len(b) {
		ec.writeBuffer = make([]byte, len(b))
	}
	ec.encryptor.XORKeyStream(ec.writeBuffer[:len(b)], b)
	return ec.conn.Write(ec.writeBuffer[:len(b)])
}

// Close closes the underlying connection.
func (ec *EncryptedConn) Close() error {
	return ec.conn.Close()
}

// NewProxyServer creates a new proxy server.
func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		connections: make(map[net.Conn]struct{}),
		stopChan:    make(chan struct{}),
	}
}

// NewProxyClient creates a new proxy client.
func NewProxyClient() *ProxyClient {
	return &ProxyClient{
		connections: make(map[net.Conn]struct{}),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the proxy server.
func (s *ProxyServer) Start(listenAddr, password string, verbose bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server already running")
	}

	if listenAddr == "" {
		return errors.New("listen address required")
	}
	if password == "" {
		return errors.New("password required")
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	s.listenAddr = listenAddr
	s.password = password
	s.verbose = verbose
	s.listener = listener
	s.connections = make(map[net.Conn]struct{})
	s.stopChan = make(chan struct{})
	s.startTime = time.Now()
	s.running = true

	s.wg.Add(1)
	go s.proxyServerAcceptLoop()

	return nil
}

// Stop stops the proxy server.
func (s *ProxyServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopChan)

	if err := s.listener.Close(); err != nil {
		// log error
	}

	s.connMutex.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[net.Conn]struct{})
	s.connMutex.Unlock()

	s.wg.Wait()
	s.running = false

	return nil
}

// IsRunning returns whether the server is running.
func (s *ProxyServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetListenAddr returns the listen address.
func (s *ProxyServer) GetListenAddr() string {
	return s.listenAddr
}

// Connections returns the number of active connections.
func (s *ProxyServer) Connections() int {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()
	return len(s.connections)
}

func (s *ProxyServer) proxyServerAcceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					continue
				}
			}

			s.wg.Add(1)
			go s.handleProxyServerConnection(conn)
		}
	}
}

func (s *ProxyServer) handleProxyServerConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.connMutex.Lock()
	s.connections[conn] = struct{}{}
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, conn)
		s.connMutex.Unlock()
	}()

	encConn, err := NewEncryptedConn(conn, s.password, true)
	if err != nil {
		return
	}

	var handshake [1]byte
	if _, err := io.ReadFull(encConn, handshake[:]); err != nil {
		return
	}

	expected := byte(len(s.password) % 256)
	if handshake[0] != expected {
		return
	}

	targetAddr, err := readTargetAddress(encConn)
	if err != nil {
		return
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return
	}
	defer targetConn.Close()

	encConn.Write([]byte{0x00})

	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(targetConn, encConn)
		targetConn.Close()
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(encConn, targetConn)
		encConn.Close()
	}()

	<-done
}

// Start starts the proxy client.
func (c *ProxyClient) Start(localAddr, serverAddr, password string, verbose bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return errors.New("client already running")
	}

	if localAddr == "" {
		return errors.New("local address required")
	}
	if serverAddr == "" {
		return errors.New("server address required")
	}
	if password == "" {
		return errors.New("password required")
	}

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}

	c.localAddr = localAddr
	c.serverAddr = serverAddr
	c.password = password
	c.verbose = verbose
	c.listener = listener
	c.connections = make(map[net.Conn]struct{})
	c.stopChan = make(chan struct{})
	c.startTime = time.Now()
	c.running = true

	c.wg.Add(1)
	go c.proxyClientAcceptLoop()

	return nil
}

// Stop stops the proxy client.
func (c *ProxyClient) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	close(c.stopChan)

	if err := c.listener.Close(); err != nil {
		// log error
	}

	c.connMutex.Lock()
	for conn := range c.connections {
		conn.Close()
	}
	c.connections = make(map[net.Conn]struct{})
	c.connMutex.Unlock()

	c.wg.Wait()
	c.running = false

	return nil
}

// IsRunning returns whether the client is running.
func (c *ProxyClient) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// GetLocalAddr returns the local address.
func (c *ProxyClient) GetLocalAddr() string {
	return c.localAddr
}

// GetServerAddr returns the server address.
func (c *ProxyClient) GetServerAddr() string {
	return c.serverAddr
}

// Connections returns the number of active connections.
func (c *ProxyClient) Connections() int {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return len(c.connections)
}

func (c *ProxyClient) proxyClientAcceptLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			conn, err := c.listener.Accept()
			if err != nil {
				select {
				case <-c.stopChan:
					return
				default:
					continue
				}
			}

			c.wg.Add(1)
			go c.handleProxyClientConnection(conn)
		}
	}
}

func (c *ProxyClient) handleProxyClientConnection(conn net.Conn) {
	defer c.wg.Done()
	defer conn.Close()

	c.connMutex.Lock()
	c.connections[conn] = struct{}{}
	c.connMutex.Unlock()

	defer func() {
		c.connMutex.Lock()
		delete(c.connections, conn)
		c.connMutex.Unlock()
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var firstByte [1]byte
	n, err := conn.Read(firstByte[:])
	if err != nil || n == 0 {
		return
	}
	conn.SetReadDeadline(time.Time{})

	if firstByte[0] == 0x05 {
		c.handleClientSOCKS5(conn, firstByte[0])
	} else if firstByte[0] == 'C' || firstByte[0] == 'G' || firstByte[0] == 'P' || firstByte[0] == 'D' || firstByte[0] == 'H' ||
		firstByte[0] == 'c' || firstByte[0] == 'g' || firstByte[0] == 'p' || firstByte[0] == 'd' || firstByte[0] == 'h' {
		c.handleClientHTTP(conn, firstByte[0])
	}
}

func (c *ProxyClient) handleClientHTTP(conn net.Conn, firstByte byte) {
	reader := bufio.NewReader(conn)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	firstLine = string(firstByte) + firstLine

	parts := strings.Fields(firstLine)
	if len(parts) < 3 {
		return
	}
	method := parts[0]
	target := parts[1]

	var targetAddr string
	var isConnect bool

	if method == "CONNECT" {
		isConnect = true
		targetAddr = target
		if !strings.Contains(targetAddr, ":") {
			targetAddr = targetAddr + ":443"
		}
	} else {
		isConnect = false
		parsedURL, err := url.Parse(target)
		if err != nil {
			return
		}
		targetAddr = parsedURL.Host
		if !strings.Contains(targetAddr, ":") {
			targetAddr = targetAddr + ":80"
		}
	}

	serverConn, err := net.DialTimeout("tcp", c.serverAddr, 10*time.Second)
	if err != nil {
		return
	}
	defer serverConn.Close()

	encConn, err := NewEncryptedConn(serverConn, c.password, false)
	if err != nil {
		return
	}

	handshake := byte(len(c.password) % 256)
	if _, err := encConn.Write([]byte{handshake}); err != nil {
		return
	}

	if err := writeTargetAddress(encConn, targetAddr); err != nil {
		return
	}

	var response [1]byte
	if _, err := io.ReadFull(encConn, response[:]); err != nil {
		return
	}
	if response[0] != 0x00 {
		return
	}

	if isConnect {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if line == "\r\n" || line == "\n" {
				break
			}
		}

		responseStr := "HTTP/1.1 200 Connection Established\r\n\r\n"
		if _, err := conn.Write([]byte(responseStr)); err != nil {
			return
		}
	} else {
		var headersBuilder strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			headersBuilder.WriteString(line)
			if line == "\r\n" || line == "\n" {
				break
			}
		}
		parsedURL, _ := url.Parse(target)
		newRequest := fmt.Sprintf("%s %s %s\r\n", method, parsedURL.Path, parts[2])
		newRequest += headersBuilder.String()
		if _, err := encConn.Write([]byte(newRequest)); err != nil {
			return
		}
	}

	buffered := reader.Buffered()
	var bufferedData []byte
	if buffered > 0 {
		bufferedData = make([]byte, buffered)
		reader.Read(bufferedData)
	}
	combinedReader := io.MultiReader(bytes.NewReader(bufferedData), conn)

	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(encConn, combinedReader)
		encConn.Close()
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(conn, encConn)
		if closer, ok := conn.(io.Closer); ok {
			closer.Close()
		}
	}()

	<-done
}

func (c *ProxyClient) handleClientSOCKS5(conn net.Conn, firstByte byte) {
	reader := bufio.NewReader(conn)

	var nmethods byte
	if err := binary.Read(reader, binary.BigEndian, &nmethods); err != nil {
		return
	}
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(reader, methods); err != nil {
		return
	}

	hasNoAuth := false
	for _, m := range methods {
		if m == 0x00 {
			hasNoAuth = true
			break
		}
	}

	if hasNoAuth {
		conn.Write([]byte{0x05, 0x00})
	} else {
		conn.Write([]byte{0x05, 0xFF})
		return
	}

	var req [10]byte
	if _, err := io.ReadFull(reader, req[:3]); err != nil {
		return
	}
	if req[0] != 0x05 || req[1] != 0x01 {
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var targetAddr string
	var atyp byte
	if err := binary.Read(reader, binary.BigEndian, &atyp); err != nil {
		return
	}

	switch atyp {
	case 0x01:
		var ip [4]byte
		if _, err := io.ReadFull(reader, ip[:]); err != nil {
			return
		}
		var port uint16
		if err := binary.Read(reader, binary.BigEndian, &port); err != nil {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)

	case 0x03:
		var len byte
		if err := binary.Read(reader, binary.BigEndian, &len); err != nil {
			return
		}
		domain := make([]byte, len)
		if _, err := io.ReadFull(reader, domain); err != nil {
			return
		}
		var port uint16
		if err := binary.Read(reader, binary.BigEndian, &port); err != nil {
			return
		}
		targetAddr = fmt.Sprintf("%s:%d", string(domain), port)

	case 0x04:
		var ip [16]byte
		if _, err := io.ReadFull(reader, ip[:]); err != nil {
			return
		}
		var port uint16
		if err := binary.Read(reader, binary.BigEndian, &port); err != nil {
			return
		}
		targetAddr = fmt.Sprintf("[%s]:%d", net.IP(ip[:]).String(), port)

	default:
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	serverConn, err := net.DialTimeout("tcp", c.serverAddr, 10*time.Second)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer serverConn.Close()

	encConn, err := NewEncryptedConn(serverConn, c.password, false)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	handshake := byte(len(c.password) % 256)
	if _, err := encConn.Write([]byte{handshake}); err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	if err := writeTargetAddress(encConn, targetAddr); err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var response [1]byte
	if _, err := io.ReadFull(encConn, response[:]); err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	if response[0] != 0x00 {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	resp := []byte{0x05, 0x00, 0x00, 0x01}
	resp = append(resp, []byte{0, 0, 0, 0}...)
	resp = append(resp, byte(localAddr.Port>>8), byte(localAddr.Port))
	conn.Write(resp)

	buffered := reader.Buffered()
	var bufferedData []byte
	if buffered > 0 {
		bufferedData = make([]byte, buffered)
		reader.Read(bufferedData)
	}
	combinedReader := io.MultiReader(bytes.NewReader(bufferedData), conn)

	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(encConn, combinedReader)
		encConn.Close()
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(conn, encConn)
		conn.Close()
	}()

	<-done
}

// readTargetAddress reads a target address from the encrypted connection.
func readTargetAddress(conn io.Reader) (string, error) {
	var lenBuf [1]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return "", err
	}
	addr := make([]byte, lenBuf[0])
	if _, err := io.ReadFull(conn, addr); err != nil {
		return "", err
	}
	return string(addr), nil
}

// writeTargetAddress writes a target address to the encrypted connection.
func writeTargetAddress(conn io.Writer, addr string) error {
	if len(addr) > 255 {
		return errors.New("address too long")
	}
	buf := make([]byte, 1+len(addr))
	buf[0] = byte(len(addr))
	copy(buf[1:], addr)
	_, err := conn.Write(buf)
	return err
}

// ============================================================
// Module Registration
// ============================================================

func init() {
	Register(&Module{
		Name: "socks",
		Exports: map[string]objects.Object{
			// ============================================================
			// SOCKS Server Functions
			// ============================================================

			"createServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createServer takes no arguments")
				}
				return NewSocksServer()
			}),

			"startServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("startServer takes at least 1 argument (port)")
				}
				port, ok := args[0].(*objects.Int)
				if !ok {
					return Error("first argument must be an integer (port)")
				}

				server := NewSocksServer()

				for i := 1; i < len(args); i++ {
					if opt, ok := args[i].(*objects.String); ok {
						switch opt.Value {
						case "-socks4":
							server.SetSocks5(false)
						case "-socks5":
							server.SetSocks5(true)
						default:
							if len(opt.Value) > 6 && opt.Value[:6] == "-auth=" {
								auth := opt.Value[6:]
								for j := 0; j < len(auth); j++ {
									if auth[j] == ':' {
										server.SetAuth(auth[:j], auth[j+1:])
										break
									}
								}
							}
						}
					}
				}

				if err := server.Start(int(port.Value)); err != nil {
					return Error("failed to start SOCKS server: " + err.Error())
				}

				return server
			}),

			"isSocksServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSocksServer takes exactly 1 argument")
				}
				_, ok := args[0].(*SocksServer)
				return Bool(ok)
			}),

			// ============================================================
			// SOCKS Client Functions
			// ============================================================

			"createClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createClient takes no arguments")
				}
				return NewSocksClient()
			}),

			"connect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("connect takes at least 2 arguments (proxyAddr, targetAddr)")
				}
				proxyAddr, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (proxyAddr)")
				}
				targetAddr, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (targetAddr)")
				}

				client := NewSocksClient()
				socks5 := true

				for i := 2; i < len(args); i++ {
					if opt, ok := args[i].(*objects.String); ok {
						switch opt.Value {
						case "-socks4":
							socks5 = false
						case "-socks5":
							socks5 = true
						}
					}
				}

				if err := client.Connect(proxyAddr.Value, targetAddr.Value, socks5); err != nil {
					return Error("SOCKS connect failed: " + err.Error())
				}

				return client
			}),

			"connectWithAuth": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("connectWithAuth takes exactly 4 arguments (proxyAddr, targetAddr, username, password)")
				}
				proxyAddr, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (proxyAddr)")
				}
				targetAddr, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (targetAddr)")
				}
				username, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (username)")
				}
				password, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (password)")
				}

				client := NewSocksClient()
				if err := client.ConnectWithAuth(proxyAddr.Value, targetAddr.Value, username.Value, password.Value); err != nil {
					return Error("SOCKS connect failed: " + err.Error())
				}

				return client
			}),

			"isSocksClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSocksClient takes exactly 1 argument")
				}
				_, ok := args[0].(*SocksClient)
				return Bool(ok)
			}),

			// ============================================================
			// Encrypted Proxy Server Functions (goconnectit style)
			// ============================================================

			"createProxyServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createProxyServer takes no arguments")
				}
				return NewProxyServer()
			}),

			// startProxyServer starts an encrypted proxy server.
			// Parameters: listenAddr, password, [verbose]
			// Returns: ProxyServer object or Error
			"startProxyServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("startProxyServer takes at least 2 arguments (listenAddr, password)")
				}
				listenAddr, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (listenAddr)")
				}
				password, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (password)")
				}

				verbose := false
				if len(args) > 2 {
					if v, ok := args[2].(*objects.Bool); ok {
						verbose = v.Value
					}
				}

				server := NewProxyServer()
				if err := server.Start(listenAddr.Value, password.Value, verbose); err != nil {
					return Error("failed to start proxy server: " + err.Error())
				}

				return server
			}),

			"isProxyServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isProxyServer takes exactly 1 argument")
				}
				_, ok := args[0].(*ProxyServer)
				return Bool(ok)
			}),

			// ============================================================
			// Encrypted Proxy Client Functions (goconnectit style)
			// ============================================================

			"createProxyClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createProxyClient takes no arguments")
				}
				return NewProxyClient()
			}),

			// startProxyClient starts an encrypted proxy client.
			// Parameters: localAddr, serverAddr, password, [verbose]
			// Returns: ProxyClient object or Error
			// The client supports HTTP, HTTPS (CONNECT), and SOCKS5 protocols.
			"startProxyClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("startProxyClient takes at least 3 arguments (localAddr, serverAddr, password)")
				}
				localAddr, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (localAddr)")
				}
				serverAddr, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (serverAddr)")
				}
				password, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (password)")
				}

				verbose := false
				if len(args) > 3 {
					if v, ok := args[3].(*objects.Bool); ok {
						verbose = v.Value
					}
				}

				client := NewProxyClient()
				if err := client.Start(localAddr.Value, serverAddr.Value, password.Value, verbose); err != nil {
					return Error("failed to start proxy client: " + err.Error())
				}

				return client
			}),

			"isProxyClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isProxyClient takes exactly 1 argument")
				}
				_, ok := args[0].(*ProxyClient)
				return Bool(ok)
			}),
		},
	})
}
