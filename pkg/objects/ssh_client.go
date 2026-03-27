// pkg/objects/ssh_client.go
// SSHClient object for Xxlang - SSH client functionality.
package objects

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"bufio"
	"golang.org/x/crypto/ssh"
)

// SSHClient represents an SSH client connection.
type SSHClient struct {
	mu         sync.Mutex
	client     *ssh.Client
	sftpClient interface{} // *sftp.Client, using interface{} to avoid import if not needed
	host       string
	port       int
	user       string
	connected  bool
}

// Type returns the object type.
func (c *SSHClient) Type() ObjectType { return SSHClientType }

// TypeTag returns the fast type tag.
func (c *SSHClient) TypeTag() TypeTag { return TagSSHClient }

// Inspect returns a string representation of the SSHClient.
func (c *SSHClient) Inspect() string {
	if c.connected {
		return fmt.Sprintf("SSHClient(connected=%s@%s:%d)", c.user, c.host, c.port)
	}
	return "SSHClient(disconnected)"
}

// ToBool returns true if connected.
func (c *SSHClient) ToBool() *Bool {
	return &Bool{Value: c.connected}
}

// HashKey returns a hash key for the SSHClient.
func (c *SSHClient) HashKey() HashKey {
	return HashKey{
		Type:  SSHClientType,
		Value: uint64(uintptr(unsafe.Pointer(c))),
	}
}

// SSHConfig holds SSH connection configuration.
type SSHConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	KeyPath        string
	KeyStr         string
	KeyPassphrase  string
	Timeout        int // seconds
	KnownHostsPath string
	IgnoreHostKey  bool
}

// NewSSHClient creates a new SSHClient (unconnected).
func NewSSHClient() *SSHClient {
	return &SSHClient{
		connected: false,
	}
}

// Connect establishes SSH connection with password.
func (c *SSHClient) Connect(host string, port int, user, password string) error {
	config := &SSHConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithKey establishes SSH connection with private key file.
func (c *SSHClient) ConnectWithKey(host string, port int, user, keyPath string) error {
	config := &SSHConfig{
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
		Timeout: 30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithKeyStr establishes SSH connection with private key string.
func (c *SSHClient) ConnectWithKeyStr(host string, port int, user, keyStr string) error {
	config := &SSHConfig{
		Host:    host,
		Port:    port,
		User:    user,
		KeyStr:  keyStr,
		Timeout: 30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithConfig establishes SSH connection with full configuration.
func (c *SSHClient) ConnectWithConfig(config *SSHConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	// Set defaults
	if config.Port == 0 {
		config.Port = 22
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	// Build SSH config
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{},
	}

	// Add authentication methods
	if config.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.Password))
	}

	if config.KeyStr != "" {
		var signer ssh.Signer
		var err error
		if config.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(config.KeyStr), []byte(config.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(config.KeyStr))
		}
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	if config.KeyPath != "" {
		keyData, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}
		var signer ssh.Signer
		if config.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(config.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(keyData)
		}
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	// Host key verification
	if config.IgnoreHostKey {
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else if config.KnownHostsPath != "" {
		verifier, err := newHostKeyVerifier(config.KnownHostsPath)
		if err != nil {
			return fmt.Errorf("failed to load known_hosts: %w", err)
		}
		sshConfig.HostKeyCallback = verifier.callback()
	} else {
		// Default: try common known_hosts locations
		homeDir, err := os.UserHomeDir()
		if err == nil {
			knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
			if _, err := os.Stat(knownHostsPath); err == nil {
				verifier, err := newHostKeyVerifier(knownHostsPath)
				if err == nil {
					sshConfig.HostKeyCallback = verifier.callback()
				} else {
					sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
				}
			} else {
				sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
			}
		} else {
			sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		}
	}

	sshConfig.Timeout = time.Duration(config.Timeout) * time.Second

	// Connect
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.client = client
	c.host = config.Host
	c.port = config.Port
	c.user = config.User
	c.connected = true

	return nil
}

// ============================================================
// Connection Management
// ============================================================

// IsConnected returns the connection status.
func (c *SSHClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Close closes the SSH connection.
func (c *SSHClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Close SFTP client if open
	if c.sftpClient != nil {
		if closer, ok := c.sftpClient.(interface{ Close() error }); ok {
			closer.Close()
		}
		c.sftpClient = nil
	}

	err := c.client.Close()
	c.connected = false
	c.client = nil
	return err
}

// GetHost returns the connected host.
func (c *SSHClient) GetHost() string {
	return c.host
}

// GetPort returns the connected port.
func (c *SSHClient) GetPort() int {
	return c.port
}

// GetUser returns the username.
func (c *SSHClient) GetUser() string {
	return c.user
}

// ============================================================
// Command Execution
// ============================================================

// Exec executes a command and returns stdout.
func (c *SSHClient) Exec(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

// ExecFull executes a command and returns stdout, stderr, and exit code.
func (c *SSHClient) ExecFull(cmd string) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)

	result := map[string]interface{}{
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
		"exitCode": 0,
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result["exitCode"] = exitErr.ExitStatus()
		} else {
			result["exitCode"] = -1
		}
	}

	return result, nil
}

// ExecStream executes a command with streaming output.
// Returns a channel for output lines and an error channel.
func (c *SSHClient) ExecStream(cmd string, callback func(line string)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if callback != nil {
			callback(scanner.Text())
		}
	}

	return session.Wait()
}

// RunScript executes a local script file on the remote server.
func (c *SSHClient) RunScript(scriptPath string) (string, error) {
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script file: %w", err)
	}
	return c.Exec(string(scriptContent))
}

// RunScriptStr executes a script string on the remote server.
func (c *SSHClient) RunScriptStr(scriptStr string) (string, error) {
	return c.Exec(scriptStr)
}

// ============================================================
// File Operations (SFTP-like using SSH exec)
// ============================================================

// ReadFile reads a remote file content.
func (c *SSHClient) ReadFile(remotePath string) (string, error) {
	// Use cat command to read file
	output, err := c.Exec(fmt.Sprintf("cat %s", escapeShellArg(remotePath)))
	if err != nil {
		return "", fmt.Errorf("failed to read remote file: %w", err)
	}
	return output, nil
}

// WriteFile writes content to a remote file.
func (c *SSHClient) WriteFile(remotePath, content string) error {
	// Use cat with heredoc to write file
	cmd := fmt.Sprintf("cat > %s << 'XXLANG_EOF'\n%s\nXXLANG_EOF", remotePath, content)
	_, err := c.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to write remote file: %w", err)
	}
	return nil
}

// Upload uploads a local file to remote server.
func (c *SSHClient) Upload(localPath, remotePath string) error {
	// Read local file
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		// Use mkdir -p with error suppression
		c.Exec(fmt.Sprintf("mkdir -p %s 2>/dev/null || true", escapeShellArg(remoteDir)))
	}

	// Write to remote file using base64 encoding for binary safety
	encoded := encodeBase64(localContent)
	cmd := fmt.Sprintf("echo '%s' | base64 -d > %s", encoded, escapeShellArg(remotePath))
	_, err = c.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// Download downloads a remote file to local.
func (c *SSHClient) Download(remotePath, localPath string) error {
	// Use base64 encoding for binary safety
	output, err := c.Exec(fmt.Sprintf("base64 %s", escapeShellArg(remotePath)))
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Decode base64
	decoded, err := decodeBase64(strings.TrimSpace(output))
	if err != nil {
		return fmt.Errorf("failed to decode downloaded content: %w", err)
	}

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if localDir != "." {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}

	// Write to local file
	if err := os.WriteFile(localPath, decoded, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

// Mkdir creates a remote directory.
func (c *SSHClient) Mkdir(remotePath string) error {
	_, err := c.Exec(fmt.Sprintf("mkdir %s", escapeShellArg(remotePath)))
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// MkdirAll creates a remote directory with parents.
func (c *SSHClient) MkdirAll(remotePath string) error {
	_, err := c.Exec(fmt.Sprintf("mkdir -p %s", escapeShellArg(remotePath)))
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// Remove removes a remote file.
func (c *SSHClient) Remove(remotePath string) error {
	_, err := c.Exec(fmt.Sprintf("rm -f %s", escapeShellArg(remotePath)))
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

// RemoveDir removes a remote directory recursively.
func (c *SSHClient) RemoveDir(remotePath string) error {
	_, err := c.Exec(fmt.Sprintf("rm -rf %s", escapeShellArg(remotePath)))
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

// Rename renames a remote file or directory.
func (c *SSHClient) Rename(oldPath, newPath string) error {
	_, err := c.Exec(fmt.Sprintf("mv %s %s", escapeShellArg(oldPath), escapeShellArg(newPath)))
	if err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// Stat returns file information.
func (c *SSHClient) Stat(remotePath string) (map[string]interface{}, error) {
	output, err := c.Exec(fmt.Sprintf("stat -c '%%s|%%Y|%%F' %s 2>/dev/null", escapeShellArg(remotePath)))
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(output), "|")
	if len(parts) < 3 {
		return nil, errors.New("unexpected stat output format")
	}

	size, _ := strconv.ParseInt(parts[0], 10, 64)
	mtime, _ := strconv.ParseInt(parts[1], 10, 64)
	fileType := parts[2]

	return map[string]interface{}{
		"size":     size,
		"mtime":    mtime,
		"type":     strings.TrimSpace(fileType),
		"path":     remotePath,
		"isDir":    strings.TrimSpace(fileType) == "directory",
		"isFile":   strings.TrimSpace(fileType) == "regular file",
	}, nil
}

// Exists checks if a path exists.
func (c *SSHClient) Exists(remotePath string) bool {
	_, err := c.Exec(fmt.Sprintf("test -e %s", escapeShellArg(remotePath)))
	return err == nil
}

// IsDir checks if path is a directory.
func (c *SSHClient) IsDir(remotePath string) bool {
	_, err := c.Exec(fmt.Sprintf("test -d %s", escapeShellArg(remotePath)))
	return err == nil
}

// IsFile checks if path is a regular file.
func (c *SSHClient) IsFile(remotePath string) bool {
	_, err := c.Exec(fmt.Sprintf("test -f %s", escapeShellArg(remotePath)))
	return err == nil
}

// ListDir lists directory contents.
func (c *SSHClient) ListDir(remotePath string) ([]map[string]interface{}, error) {
	output, err := c.Exec(fmt.Sprintf("ls -la %s 2>/dev/null", escapeShellArg(remotePath)))
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []map[string]interface{}

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "total ") {
			continue
		}

		// Parse ls -la output
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// Skip . and ..
		name := strings.Join(fields[8:], " ")
		if name == "." || name == ".." {
			continue
		}

		mode := fields[0]
		size, _ := strconv.ParseInt(fields[4], 10, 64)

		result = append(result, map[string]interface{}{
			"name":  name,
			"mode":  mode,
			"size":  size,
			"isDir": strings.HasPrefix(mode, "d"),
		})
	}

	return result, nil
}

// WalkDir recursively lists directory contents.
func (c *SSHClient) WalkDir(remotePath string) ([]map[string]interface{}, error) {
	output, err := c.Exec(fmt.Sprintf("find %s -type f -o -type d 2>/dev/null", escapeShellArg(remotePath)))
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []map[string]interface{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		isDir, _ := c.Exec(fmt.Sprintf("test -d %s", escapeShellArg(line)))
		result = append(result, map[string]interface{}{
			"path":  line,
			"isDir": isDir == "",
		})
	}

	return result, nil
}

// UploadDir uploads a local directory to remote.
func (c *SSHClient) UploadDir(localDir, remoteDir string) error {
	// Create remote directory
	if err := c.MkdirAll(remoteDir); err != nil {
		return err
	}

	// Walk local directory
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		remotePath := filepath.Join(remoteDir, relPath)

		if info.IsDir() {
			return c.MkdirAll(remotePath)
		}

		return c.Upload(path, remotePath)
	})
}

// DownloadDir downloads a remote directory to local.
func (c *SSHClient) DownloadDir(remoteDir, localDir string) error {
	// Create local directory
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	// Get remote file list
	files, err := c.WalkDir(remoteDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		remotePath := file["path"].(string)
		relPath := strings.TrimPrefix(remotePath, remoteDir)
		if relPath == "" {
			continue
		}
		relPath = strings.TrimPrefix(relPath, "/")

		localPath := filepath.Join(localDir, relPath)

		if file["isDir"].(bool) {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return err
			}
		} else {
			if err := c.Download(remotePath, localPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================
// Port Forwarding
// ============================================================

// LocalForward creates local port forwarding.
// Returns a forwarding ID that can be used to stop it.
func (c *SSHClient) LocalForward(localPort int, remoteHost string, remotePort int) (net.Listener, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	localAddr := fmt.Sprintf("127.0.0.1:%d", localPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}

			go func() {
				defer localConn.Close()

				remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
				remoteConn, err := c.client.Dial("tcp", remoteAddr)
				if err != nil {
					return
				}
				defer remoteConn.Close()

				// Bidirectional copy
				go io.Copy(localConn, remoteConn)
				io.Copy(remoteConn, localConn)
			}()
		}
	}()

	return listener, nil
}

// GetClient returns the underlying ssh.Client (for advanced use).
func (c *SSHClient) GetClient() *ssh.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client
}

// ============================================================
// Helper Functions
// ============================================================

// escapeShellArg escapes a shell argument.
func escapeShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

// encodeBase64 encodes bytes to base64 string.
func encodeBase64(data []byte) string {
	return strings.ReplaceAll(string(encodeBase64Bytes(data)), "\n", "")
}

// decodeBase64 decodes base64 string to bytes.
func decodeBase64(s string) ([]byte, error) {
	return decodeBase64String(s)
}

// Simple base64 implementation to avoid importing encoding/base64
// These are implemented inline for simplicity

var base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func encodeBase64Bytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	var result []byte
	for i := 0; i < len(data); i += 3 {
		var n uint32
		remaining := len(data) - i

		n = uint32(data[i]) << 16
		if remaining > 1 {
			n |= uint32(data[i+1]) << 8
		}
		if remaining > 2 {
			n |= uint32(data[i+2])
		}

		result = append(result, base64Chars[(n>>18)&0x3F])
		result = append(result, base64Chars[(n>>12)&0x3F])

		if remaining > 1 {
			result = append(result, base64Chars[(n>>6)&0x3F])
		} else {
			result = append(result, '=')
		}

		if remaining > 2 {
			result = append(result, base64Chars[n&0x3F])
		} else {
			result = append(result, '=')
		}
	}

	return result
}

func decodeBase64String(s string) ([]byte, error) {
	// Remove any whitespace/newlines
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' {
			return -1
		}
		return r
	}, s)

	if len(s)%4 != 0 {
		return nil, errors.New("invalid base64 string length")
	}

	// Build decode table
	decodeTable := make(map[byte]int)
	for i := 0; i < 64; i++ {
		decodeTable[base64Chars[i]] = i
	}

	var result []byte
	for i := 0; i < len(s); i += 4 {
		var n uint32
		padCount := 0

		for j := 0; j < 4; j++ {
			if s[i+j] == '=' {
				padCount++
				continue
			}
			val, ok := decodeTable[s[i+j]]
			if !ok {
				return nil, fmt.Errorf("invalid base64 character: %c", s[i+j])
			}
			n |= uint32(val) << uint(18-j*6)
		}

		result = append(result, byte((n>>16)&0xFF))
		if padCount < 2 {
			result = append(result, byte((n>>8)&0xFF))
		}
		if padCount < 1 {
			result = append(result, byte(n&0xFF))
		}
	}

	return result, nil
}

// ============================================================
// knownhosts implementation (minimal, for host key verification)
// ============================================================

// hostKeyVerifier stores known host keys for verification
type hostKeyVerifier struct {
	hosts map[string]ssh.PublicKey
}

// newHostKeyVerifier creates a host key verifier from a known_hosts file.
func newHostKeyVerifier(filename string) (*hostKeyVerifier, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	verifier := &hostKeyVerifier{
		hosts: make(map[string]ssh.PublicKey),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Parse host key
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(fields[1] + " " + fields[2]))
		if err != nil {
			continue
		}

		// Store for each host (may have multiple hosts separated by comma)
		hosts := strings.Split(fields[0], ",")
		for _, host := range hosts {
			verifier.hosts[host] = key
		}
	}

	return verifier, nil
}

// callback returns a HostKeyCallback function.
func (v *hostKeyVerifier) callback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		storedKey, ok := v.hosts[hostname]
		if !ok {
			// Try with port
			if !strings.Contains(hostname, ":") {
				// Try with port 22
				storedKey, ok = v.hosts[hostname+":22"]
			}
		}

		if !ok {
			return fmt.Errorf("unknown host: %s", hostname)
		}

		if !bytes.Equal(storedKey.Marshal(), key.Marshal()) {
			return fmt.Errorf("host key mismatch for %s", hostname)
		}

		return nil
	}
}