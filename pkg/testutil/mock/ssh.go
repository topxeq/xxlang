// pkg/testutil/mock/ssh.go
// Mock SSH server for testing SSH client functionality.

package mock

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHMockServerConfig holds configuration for the mock SSH server.
type SSHMockServerConfig struct {
	// Address to bind to (default: 127.0.0.1:0 for random port)
	BindAddr string
	// Server version string
	Version string
	// Connection timeout
	Timeout time.Duration
}

// DefaultConfig returns a default configuration for the mock SSH server.
func DefaultConfig() *SSHMockServerConfig {
	return &SSHMockServerConfig{
		BindAddr: "127.0.0.1:0",
		Version:  "SSH-2.0-MockServer",
		Timeout:  30 * time.Second,
	}
}

// SSHMockServer is a mock SSH server for testing.
type SSHMockServer struct {
	config           *SSHMockServerConfig
	listener         net.Listener
	serverConfig     *ssh.ServerConfig
	userPasswords    map[string]string
	commandResponses map[string]string
	mu               sync.RWMutex
	running          bool
	port             int
	wg               sync.WaitGroup
}

// NewSSHMockServer creates a new mock SSH server.
func NewSSHMockServer(config *SSHMockServerConfig) *SSHMockServer {
	if config == nil {
		config = DefaultConfig()
	}

	// Generate a random RSA key for the server
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	server := &SSHMockServer{
		config:           config,
		userPasswords:    make(map[string]string),
		commandResponses: make(map[string]string),
	}

	// Create SSH signer from the private key
	signer, _ := ssh.NewSignerFromKey(privateKey)

	server.serverConfig = &ssh.ServerConfig{
		PasswordCallback: server.passwordCallback,
	}
	server.serverConfig.AddHostKey(signer)

	return server
}

// SetUserPassword sets a username/password combination for authentication.
func (s *SSHMockServer) SetUserPassword(user, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userPasswords[user] = password
}

// SetCommandResponse sets a response for a specific command.
func (s *SSHMockServer) SetCommandResponse(cmd, response string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandResponses[cmd] = response
}

// Start starts the mock SSH server.
func (s *SSHMockServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.mu.Unlock()

	var err error
	s.listener, err = net.Listen("tcp", s.config.BindAddr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Get the actual port
	s.port = s.listener.Addr().(*net.TCPAddr).Port
	s.running = true

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop stops the mock SSH server.
func (s *SSHMockServer) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

// Port returns the port the server is listening on.
func (s *SSHMockServer) Port() int {
	return s.port
}

func (s *SSHMockServer) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expectedPassword, ok := s.userPasswords[conn.User()]
	if !ok {
		return nil, fmt.Errorf("unknown user: %s", conn.User())
	}

	if string(password) != expectedPassword {
		return nil, fmt.Errorf("invalid password for user: %s", conn.User())
	}

	return &ssh.Permissions{
		Extensions: map[string]string{
			"user": conn.User(),
		},
	}, nil
}

func (s *SSHMockServer) acceptLoop() {
	defer s.wg.Done()

	for {
		s.mu.RLock()
		running := s.running
		s.mu.RUnlock()

		if !running {
			return
		}

		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.RLock()
			running := s.running
			s.mu.RUnlock()
			if running {
				// Only log error if we're still supposed to be running
			}
			return
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *SSHMockServer) handleConn(conn net.Conn) {
	defer s.wg.Add(-1)
	defer conn.Close()

	sshConn, newChannels, requests, err := ssh.NewServerConn(conn, s.serverConfig)
	if err != nil {
		return
	}
	defer sshConn.Close()

	// Handle global requests
	go ssh.DiscardRequests(requests)

	// Handle new channels
	for newChan := range newChannels {
		s.wg.Add(1)
		go s.handleChannel(newChan)
	}
}

func (s *SSHMockServer) handleChannel(newChan ssh.NewChannel) {
	defer s.wg.Add(-1)

	if newChan.ChannelType() != "session" {
		newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	channel, requests, err := newChan.Accept()
	if err != nil {
		return
	}
	defer channel.Close()

	// Handle requests
	for req := range requests {
		s.wg.Add(1)
		go s.handleRequest(channel, req)
	}
}

func (s *SSHMockServer) handleRequest(channel ssh.Channel, req *ssh.Request) {
	defer s.wg.Add(-1)

	switch req.Type {
	case "exec":
		// Parse the command
		var payload struct {
			Command string
		}
		if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
			channel.SendRequest("exit-status", false, ssh.Marshal(struct{ ExitStatus uint32 }{ExitStatus: 1}))
			return
		}

		// Get the response for this command
		s.mu.RLock()
		response, ok := s.commandResponses[payload.Command]
		s.mu.RUnlock()

		if !ok {
			// Default response
			response = ""
		}

		// Send stdout data
		channel.Write([]byte(response))

		// Send exit status
		channel.SendRequest("exit-status", false, ssh.Marshal(struct{ ExitStatus uint32 }{ExitStatus: 0}))

	default:
		// Unknown request type
	}

	if req.WantReply {
		channel.SendRequest(req.Type, false, nil)
	}
}

// encodeBase64 encodes bytes to base64 string.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// decodeBase64 decodes base64 string to bytes.
func decodeBase64(s string) ([]byte, error) {
	// Add padding if needed
	for len(s)%4 != 0 {
		s += "="
	}
	return base64.StdEncoding.DecodeString(s)
}

// escapeShellArg escapes a shell argument for safe use in shell commands.
func escapeShellArg(arg string) string {
	if arg == "" {
		return "''"
	}

	// Simple escaping: wrap in single quotes and escape any single quotes
	result := "'"
	for _, c := range arg {
		if c == '\'' {
			result += "'\\''"
		} else {
			result += string(c)
		}
	}
	result += "'"
	return result
}

// MockSSHConnection represents a mock SSH connection for testing.
type MockSSHConnection struct {
	Host       string
	Port       int
	User       string
	Connected  bool
	LastOutput string
	mu         sync.Mutex
}

// NewMockSSHConnection creates a new mock SSH connection.
func NewMockSSHConnection(host string, port int, user string) *MockSSHConnection {
	return &MockSSHConnection{
		Host:      host,
		Port:      port,
		User:      user,
		Connected: false,
	}
}

// Connect connects to the mock SSH server.
func (c *MockSSHConnection) Connect(password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Connected = true
	return nil
}

// Close closes the connection.
func (c *MockSSHConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Connected = false
	return nil
}

// Exec executes a command on the mock server.
func (c *MockSSHConnection) Exec(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Connected {
		return "", fmt.Errorf("not connected")
	}

	c.LastOutput = "mock output for: " + cmd
	return c.LastOutput, nil
}

// IsConnected returns whether the connection is active.
func (c *MockSSHConnection) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Connected
}

// Scanner is a simple scanner for reading lines from a reader.
type Scanner struct {
	reader *bufio.Reader
	line   string
	err    error
}

// NewScanner creates a new scanner for the given reader.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{reader: bufio.NewReader(r)}
}

// Scan reads the next line.
func (s *Scanner) Scan() bool {
	s.line, s.err = s.reader.ReadString('\n')
	return s.err == nil
}

// Text returns the current line.
func (s *Scanner) Text() string {
	return strings.TrimSpace(s.line)
}

// Err returns any error that occurred.
func (s *Scanner) Err() error {
	return s.err
}
