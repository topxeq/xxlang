// pkg/objects/ssh_client.go
// SSHClient object for Xxlang - SSH client functionality.
package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"bufio"
	"golang.org/x/crypto/ssh"
)

// SSHClient represents an SSH client connection.
type SSHClient struct {
	mu          sync.Mutex
	client      *ssh.Client
	sftpChannel ssh.Channel  // SFTP subsystem channel for binary file transfer
	sftpNextID  uint32       // SFTP request ID counter
	host        string
	port        int
	user        string
	connected   bool
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

	// Close SFTP channel if open
	if c.sftpChannel != nil {
		c.sftpChannel.Close()
		c.sftpChannel = nil
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
// File Operations (via SFTP binary transfer)
// ============================================================

// ReadFile reads a remote file content via SFTP binary transfer.
func (c *SSHClient) ReadFile(remotePath string) (string, error) {
	data, err := c.ReadBytes(remotePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadBytes reads a remote file and returns raw bytes via SFTP binary transfer.
func (c *SSHClient) ReadBytes(remotePath string) ([]byte, error) {
	c.mu.Lock()

	if !c.connected {
		c.mu.Unlock()
		return nil, errors.New("not connected")
	}

	// Initialize SFTP channel if not yet open
	if err := c.initSftp(); err != nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to initialize SFTP: %w", err)
	}

	// Open remote file for reading
	handle, err := c.sftpOpenFile(remotePath, sshFxfRead)
	if err != nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to open remote file: %w", err)
	}

	// Read all data in 32KB chunks
	var data []byte
	offset := uint64(0)
	chunkSize := uint32(32768)

	for {
		chunk, err := c.sftpReadChunk(handle, offset, chunkSize)
		if err != nil {
			c.sftpCloseHandle(handle)
			c.mu.Unlock()
			return nil, fmt.Errorf("failed to read data: %w", err)
		}
		if len(chunk) == 0 {
			break
		}
		data = append(data, chunk...)
		offset += uint64(len(chunk))
		if uint32(len(chunk)) < chunkSize {
			break
		}
	}

	c.sftpCloseHandle(handle)
	c.mu.Unlock()
	return data, nil
}

// WriteFile writes content to a remote file.
func (c *SSHClient) WriteFile(remotePath, content string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.writeFileInternal(remotePath, content)
}

// writeFileInternal writes content to a remote file via SFTP (internal, assumes lock held).
func (c *SSHClient) writeFileInternal(remotePath, content string) error {
	// Initialize SFTP channel if not yet open
	if err := c.initSftp(); err != nil {
		return fmt.Errorf("failed to initialize SFTP: %w", err)
	}

	// Create parent directory if needed
	dir := path.Dir(remotePath)
	if dir != "." && dir != "/" && dir != "" {
		c.sftpMkdir(dir)
	}

	// Open remote file for writing (create or truncate)
	handle, err := c.sftpOpenFile(remotePath, sshFxfWrite|sshFxfCreat|sshFxfTrunc)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.sftpCloseHandle(handle)

	// Write data in 32KB chunks
	data := []byte(content)
	offset := uint64(0)
	chunkSize := uint32(32768)

	for offset < uint64(len(data)) {
		end := offset + uint64(chunkSize)
		if end > uint64(len(data)) {
			end = uint64(len(data))
		}
		if err := c.sftpWriteChunk(handle, offset, data[offset:end]); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
		offset = end
	}

	return nil
}

// execInternal executes a command without locking the mutex.
func (c *SSHClient) execInternal(cmd string) (string, error) {
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

// ============================================================
// SFTP Binary Transfer
// ============================================================

// SFTP packet types and flags (must match sftp_client.go constants).
const (
	sshFxpInit    = 1
	sshFxpVersion = 2
	sshFxpOpen    = 3
	sshFxpClose   = 4
	sshFxpRead    = 5
	sshFxpWrite   = 6
	sshFxpLstat   = 7
	sshFxpOpendir = 11
	sshFxpReaddir = 12
	sshFxpRemove  = 13
	sshFxpMkdir   = 14
	sshFxpRmdir   = 15
	sshFxpStat    = 17
	sshFxpRename  = 18
	sshFxpStatus  = 101
	sshFxpHandle  = 102
	sshFxpData    = 103
	sshFxpName    = 104
	sshFxpAttrs   = 105

	sshFxfWrite = 0x00000002
	sshFxfCreat = 0x00000008
	sshFxfTrunc = 0x00000010
	sshFxfRead  = 0x00000001

	// SFTP file attribute flags
	sftpAttrSize        = 0x00000001
	sftpAttrUidGid      = 0x00000002
	sftpAttrPermissions = 0x00000004
	sftpAttrAcmodtime   = 0x00000008
)

// initSftp lazily initializes the SFTP subsystem channel over the existing SSH connection.
// This must be called with c.mu held.
func (c *SSHClient) initSftp() error {
	if c.sftpChannel != nil {
		return nil
	}

	// Open an SSH channel for the SFTP subsystem
	channel, reqs, err := c.client.OpenChannel("session", nil)
	if err != nil {
		return fmt.Errorf("failed to open SFTP channel: %w", err)
	}
	go ssh.DiscardRequests(reqs)

	// Request the SFTP subsystem
	_, err = channel.SendRequest("subsystem", true, ssh.Marshal(struct {
		Subsystem string
	}{Subsystem: "sftp"}))
	if err != nil {
		channel.Close()
		return fmt.Errorf("failed to start SFTP subsystem: %w", err)
	}

	// Send SFTP INIT packet: [length=5][SSH_FXP_INIT][version=3]
	initPkt := make([]byte, 9)
	binary.BigEndian.PutUint32(initPkt[0:4], 5) // payload length
	initPkt[4] = byte(sshFxpInit)
	binary.BigEndian.PutUint32(initPkt[5:9], 3) // SFTP protocol version 3
	if _, err := channel.Write(initPkt); err != nil {
		channel.Close()
		return fmt.Errorf("failed to send SFTP INIT: %w", err)
	}

	// Read VERSION response
	resp, err := sftpReadPacket(channel)
	if err != nil {
		channel.Close()
		return fmt.Errorf("failed to read SFTP VERSION: %w", err)
	}
	if len(resp) == 0 || resp[0] != byte(sshFxpVersion) {
		channel.Close()
		return fmt.Errorf("unexpected SFTP response, expected VERSION")
	}

	c.sftpChannel = channel
	c.sftpNextID = 0
	return nil
}

// sftpNextRequestID returns the next SFTP request ID.
func (c *SSHClient) sftpNextRequestID() uint32 {
	c.sftpNextID++
	return c.sftpNextID
}

// sftpSendPacket sends an SFTP packet over the channel.
func (c *SSHClient) sftpSendPacket(pkt []byte) error {
	length := uint32(len(pkt))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], pkt)
	_, err := c.sftpChannel.Write(buf)
	return err
}

// sftpOpenFile opens a remote file via SFTP and returns the handle.
func (c *SSHClient) sftpOpenFile(path string, flags uint32) (string, error) {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path)+4+4)
	pkt[0] = byte(sshFxpOpen)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:9+len(path)], path)
	pos := 9 + len(path)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], flags)
	binary.BigEndian.PutUint32(pkt[pos+4:pos+8], 0) // attrs

	if err := c.sftpSendPacket(pkt); err != nil {
		return "", err
	}

	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return "", err
	}
	if len(resp) == 0 || resp[0] != byte(sshFxpHandle) {
		return "", fmt.Errorf("expected SFTP HANDLE response, got type %d", resp[0])
	}
	if len(resp) < 9 {
		return "", fmt.Errorf("SFTP HANDLE response too short")
	}
	handleLen := binary.BigEndian.Uint32(resp[5:9])
	if int(9+handleLen) > len(resp) {
		return "", fmt.Errorf("SFTP HANDLE response truncated")
	}
	return string(resp[9 : 9+handleLen]), nil
}

// sftpCloseHandle closes an SFTP file handle.
func (c *SSHClient) sftpCloseHandle(handle string) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(handle))
	pkt[0] = byte(sshFxpClose)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
	copy(pkt[9:], handle)

	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	return c.sftpExpectStatus(id)
}

// sftpWriteChunk writes a chunk of data at the given offset.
func (c *SSHClient) sftpWriteChunk(handle string, offset uint64, data []byte) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(handle)+8+4+len(data))
	pkt[0] = byte(sshFxpWrite)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
	copy(pkt[9:9+len(handle)], handle)
	pos := 9 + len(handle)
	binary.BigEndian.PutUint64(pkt[pos:pos+8], offset)
	binary.BigEndian.PutUint32(pkt[pos+8:pos+12], uint32(len(data)))
	copy(pkt[pos+12:], data)

	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	return c.sftpExpectStatus(id)
}

// sftpReadChunk reads a chunk of data at the given offset.
func (c *SSHClient) sftpReadChunk(handle string, offset uint64, length uint32) ([]byte, error) {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(handle)+8+4)
	pkt[0] = byte(sshFxpRead)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
	copy(pkt[9:9+len(handle)], handle)
	pos := 9 + len(handle)
	binary.BigEndian.PutUint64(pkt[pos:pos+8], offset)
	binary.BigEndian.PutUint32(pkt[pos+8:pos+12], length)

	if err := c.sftpSendPacket(pkt); err != nil {
		return nil, err
	}

	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("empty SFTP response")
	}
	if resp[0] == byte(sshFxpStatus) {
		return nil, nil // EOF or error, treat as EOF
	}
	if resp[0] != byte(sshFxpData) {
		return nil, fmt.Errorf("expected SFTP DATA response, got type %d", resp[0])
	}
	dataLen := binary.BigEndian.Uint32(resp[5:9])
	if int(9+dataLen) > len(resp) {
		return nil, fmt.Errorf("SFTP DATA response too short")
	}
	return resp[9 : 9+dataLen], nil
}

// sftpMkdir creates a remote directory via SFTP.
func (c *SSHClient) sftpMkdir(path string) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path)+4)
	pkt[0] = byte(sshFxpMkdir)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:9+len(path)], path)
	binary.BigEndian.PutUint32(pkt[9+len(path):], 0) // attrs

	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	// Ignore status response errors (directory may already exist)
	sftpReadPacket(c.sftpChannel)
	return nil
}

// sftpExpectStatus reads and validates a STATUS response.
func (c *SSHClient) sftpExpectStatus(expectedID uint32) error {
	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return err
	}
	if len(resp) < 5 || resp[0] != byte(sshFxpStatus) {
		return fmt.Errorf("expected SFTP STATUS response")
	}
	respID := binary.BigEndian.Uint32(resp[1:5])
	statusCode := binary.BigEndian.Uint32(resp[5:9])
	if respID != expectedID {
		return fmt.Errorf("SFTP response ID mismatch: expected %d, got %d", expectedID, respID)
	}
	if statusCode != 0 {
		return fmt.Errorf("SFTP error status: %d", statusCode)
	}
	return nil
}

// sftpExpectStatusOk reads a STATUS response and returns true if OK.
func (c *SSHClient) sftpExpectStatusOk() bool {
	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return false
	}
	if len(resp) < 9 || resp[0] != byte(sshFxpStatus) {
		return false
	}
	statusCode := binary.BigEndian.Uint32(resp[5:9])
	return statusCode == 0
}

// sftpStat returns file info via SFTP STAT (follows symlinks).
func (c *SSHClient) sftpStat(path string) (*SftpFileInfo, error) {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path))
	pkt[0] = byte(sshFxpStat)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:], path)

	if err := c.sftpSendPacket(pkt); err != nil {
		return nil, err
	}
	return c.sftpExpectAttrs(id)
}

// sftpLstat returns file info via SFTP LSTAT (does not follow symlinks).
func (c *SSHClient) sftpLstat(path string) (*SftpFileInfo, error) {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path))
	pkt[0] = byte(sshFxpLstat)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:], path)

	if err := c.sftpSendPacket(pkt); err != nil {
		return nil, err
	}
	return c.sftpExpectAttrs(id)
}

// sftpExpectAttrs reads and parses an ATTRS response.
func (c *SSHClient) sftpExpectAttrs(expectedID uint32) (*SftpFileInfo, error) {
	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 || resp[0] != byte(sshFxpAttrs) {
		return nil, fmt.Errorf("expected SFTP ATTRS response")
	}
	respID := binary.BigEndian.Uint32(resp[1:5])
	if respID != expectedID {
		return nil, fmt.Errorf("SFTP response ID mismatch")
	}

	info := &SftpFileInfo{}
	pos := 5
	if pos+4 > len(resp) {
		return info, nil
	}
	attrFlags := binary.BigEndian.Uint32(resp[pos : pos+4])
	pos += 4

	if attrFlags&sftpAttrSize != 0 {
		if pos+8 <= len(resp) {
			info.Size = int64(binary.BigEndian.Uint64(resp[pos : pos+8]))
		}
		pos += 8
	}
	if attrFlags&sftpAttrUidGid != 0 {
		pos += 8
	}
	if attrFlags&sftpAttrPermissions != 0 {
		if pos+4 <= len(resp) {
			info.Mode = binary.BigEndian.Uint32(resp[pos : pos+4])
			info.IsDir = (info.Mode & 0040000) != 0
		}
		pos += 4
	}
	if attrFlags&sftpAttrAcmodtime != 0 {
		if pos+8 <= len(resp) {
			pos += 4 // skip atime
			info.ModTime = int64(binary.BigEndian.Uint32(resp[pos : pos+4]))
			pos += 4
		} else {
			pos += 8
		}
	}
	return info, nil
}

// sftpRemove removes a remote file via SFTP.
func (c *SSHClient) sftpRemove(path string) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path))
	pkt[0] = byte(sshFxpRemove)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:], path)
	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	return c.sftpExpectStatus(id)
}

// sftpRmdir removes a remote directory via SFTP.
func (c *SSHClient) sftpRmdir(path string) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path))
	pkt[0] = byte(sshFxpRmdir)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:], path)
	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	return c.sftpExpectStatus(id)
}

// sftpRename renames a remote file or directory via SFTP.
func (c *SSHClient) sftpRename(oldPath, newPath string) error {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(oldPath)+4+len(newPath))
	pkt[0] = byte(sshFxpRename)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(oldPath)))
	copy(pkt[9:9+len(oldPath)], oldPath)
	pos := 9 + len(oldPath)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], uint32(len(newPath)))
	copy(pkt[pos+4:], newPath)
	if err := c.sftpSendPacket(pkt); err != nil {
		return err
	}
	return c.sftpExpectStatus(id)
}

// sftpListDir lists directory contents via SFTP OPENDIR/READDIR.
func (c *SSHClient) sftpListDir(path string) ([]SftpFileInfo, error) {
	id := c.sftpNextRequestID()
	pkt := make([]byte, 1+4+4+len(path))
	pkt[0] = byte(sshFxpOpendir)
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:], path)
	if err := c.sftpSendPacket(pkt); err != nil {
		return nil, err
	}

	resp, err := sftpReadPacket(c.sftpChannel)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 || resp[0] != byte(sshFxpHandle) {
		return nil, fmt.Errorf("expected SFTP HANDLE response for opendir")
	}
	if len(resp) < 9 {
			return nil, fmt.Errorf("SFTP HANDLE response too short for opendir")
		}
	handleLen := binary.BigEndian.Uint32(resp[5:9])
	handle := string(resp[9 : 9+handleLen])

	var result []SftpFileInfo
	for {
		id := c.sftpNextRequestID()
		pkt := make([]byte, 1+4+4+len(handle))
		pkt[0] = byte(sshFxpReaddir)
		binary.BigEndian.PutUint32(pkt[1:5], id)
		binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
		copy(pkt[9:], handle)
		if err := c.sftpSendPacket(pkt); err != nil {
			break
		}

		resp, err := sftpReadPacket(c.sftpChannel)
		if err != nil || len(resp) == 0 || resp[0] == byte(sshFxpStatus) || resp[0] != byte(sshFxpName) {
			break
		}

		count := binary.BigEndian.Uint32(resp[5:9])
		pos := 9
		for i := uint32(0); i < count; i++ {
			if pos+4 > len(resp) {
				break
			}
			nameLen := binary.BigEndian.Uint32(resp[pos : pos+4])
			pos += 4
			if pos+int(nameLen) > len(resp) {
				break
			}
			name := string(resp[pos : pos+int(nameLen)])
			pos += int(nameLen)

			// Skip long name
			if pos+4 > len(resp) {
				break
			}
			longNameLen := binary.BigEndian.Uint32(resp[pos : pos+4])
			pos += 4 + int(longNameLen)

			// Parse attrs
			if pos+4 > len(resp) {
				break
			}
			attrFlags := binary.BigEndian.Uint32(resp[pos : pos+4])
			pos += 4

			var size int64
			var mode uint32
			var modTime int64
			isDir := false

			if attrFlags&sftpAttrSize != 0 {
				if pos+8 <= len(resp) {
					size = int64(binary.BigEndian.Uint64(resp[pos : pos+8]))
				}
				pos += 8
			}
			if attrFlags&sftpAttrUidGid != 0 {
				pos += 8
			}
			if attrFlags&sftpAttrPermissions != 0 {
				if pos+4 <= len(resp) {
					mode = binary.BigEndian.Uint32(resp[pos : pos+4])
					isDir = (mode & 0040000) != 0
				}
				pos += 4
			}
			if attrFlags&sftpAttrAcmodtime != 0 {
				if pos+8 <= len(resp) {
					pos += 4
					modTime = int64(binary.BigEndian.Uint32(resp[pos : pos+4]))
					pos += 4
				} else {
					pos += 8
				}
			}

			if name == "." || name == ".." {
				continue
			}
			result = append(result, SftpFileInfo{
				Name:    name,
				Size:    size,
				Mode:    mode,
				ModTime: modTime,
				IsDir:   isDir,
			})
		}
	}

	c.sftpCloseHandle(handle)
	return result, nil
}

// sftpWalkDir recursively walks a remote directory via SFTP.
func (c *SSHClient) sftpWalkDir(path string) ([]SftpFileInfo, error) {
	var result []SftpFileInfo
	entries, err := c.sftpListDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fullPath := path + "/" + entry.Name
		if entry.IsDir {
			sub, err := c.sftpWalkDir(fullPath)
			if err != nil {
				continue
			}
			result = append(result, sub...)
		} else {
			result = append(result, SftpFileInfo{
				Name:    fullPath,
				Size:    entry.Size,
				Mode:    entry.Mode,
				ModTime: entry.ModTime,
				IsDir:   false,
			})
		}
	}
	return result, nil
}

// sftpMkdirAll creates a directory and all parents via SFTP.
func (c *SSHClient) sftpMkdirAll(dirPath string) error {
	info, err := c.sftpStat(dirPath)
	if err == nil && info.IsDir {
		return nil
	}
	parent := path.Dir(dirPath)
	if parent != dirPath && parent != "." && parent != "/" && parent != "" {
		if err := c.sftpMkdirAll(parent); err != nil {
			return err
		}
	}
	c.sftpMkdir(dirPath)
	return nil
}

// sftpReadPacket reads an SFTP packet from the channel.
func sftpReadPacket(r io.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	if length > 32*1024*1024 {
		return nil, errors.New("SFTP packet too large")
	}
	pkt := make([]byte, length)
	if _, err := io.ReadFull(r, pkt); err != nil {
		return nil, err
	}
	return pkt, nil
}

// Upload uploads a local file to remote server via SFTP.
func (c *SSHClient) Upload(localPath, remotePath string) error {
	// Read local file
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Initialize SFTP channel
	if err := c.initSftp(); err != nil {
		return fmt.Errorf("failed to initialize SFTP: %w", err)
	}

	// Create remote directory if needed
	remoteDir := path.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" && remoteDir != "" {
		c.sftpMkdir(remoteDir)
	}

	// Open remote file for writing (create or truncate)
	handle, err := c.sftpOpenFile(remotePath, sshFxfWrite|sshFxfCreat|sshFxfTrunc)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.sftpCloseHandle(handle)

	// Write data in 32KB chunks
	offset := uint64(0)
	chunkSize := uint32(32768)

	for offset < uint64(len(localContent)) {
		end := offset + uint64(chunkSize)
		if end > uint64(len(localContent)) {
			end = uint64(len(localContent))
		}
		if err := c.sftpWriteChunk(handle, offset, localContent[offset:end]); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
		offset = end
	}

	return nil
}

// Download downloads a remote file to local via SFTP.
func (c *SSHClient) Download(remotePath, localPath string) error {
	data, err := c.ReadBytes(remotePath)
	if err != nil {
		return err
	}

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if localDir != "." && localDir != "" {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}

	// Write to local file
	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

// Mkdir creates a remote directory via SFTP.
func (c *SSHClient) Mkdir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return err
	}
	return c.sftpMkdir(remotePath)
}

// MkdirAll creates a remote directory with parents via SFTP.
func (c *SSHClient) MkdirAll(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return err
	}
	return c.sftpMkdirAll(remotePath)
}

// Remove removes a remote file via SFTP.
func (c *SSHClient) Remove(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return err
	}
	return c.sftpRemove(remotePath)
}

// RemoveDir removes a remote directory recursively via SFTP.
func (c *SSHClient) RemoveDir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return err
	}
	return c.sftpRemoveDirRecursive(remotePath)
}

// sftpRemoveDirRecursive recursively removes a directory via SFTP.
func (c *SSHClient) sftpRemoveDirRecursive(path string) error {
	entries, err := c.sftpListDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullPath := path + "/" + entry.Name
		if entry.IsDir {
			if err := c.sftpRemoveDirRecursive(fullPath); err != nil {
				return err
			}
		} else {
			if err := c.sftpRemove(fullPath); err != nil {
				return err
			}
		}
	}
	return c.sftpRmdir(path)
}

// Rename renames a remote file or directory via SFTP.
func (c *SSHClient) Rename(oldPath, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return err
	}
	return c.sftpRename(oldPath, newPath)
}

// Stat returns file information via SFTP.
func (c *SSHClient) Stat(remotePath string) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return nil, errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return nil, err
	}

	info, err := c.sftpStat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return map[string]interface{}{
		"size":   info.Size,
		"path":   remotePath,
		"isDir":  info.IsDir,
		"isFile": !info.IsDir,
		"mode":    sftpModeToString(info.Mode, info.IsDir),
	"modTime": info.ModTime,
	}, nil
}

// sftpModeToString converts SFTP permission bits to a ls-style string.
func sftpModeToString(mode uint32, isDir bool) string {
	var buf [10]byte
	if isDir {
		buf[0] = 'd'
	} else {
		buf[0] = '-'
	}
	perms := []uint32{0400, 0200, 0100, 040, 020, 010, 04, 02, 01}
	ch := []byte{'r', 'w', 'x', 'r', 'w', 'x', 'r', 'w', 'x'}
	for i, p := range perms {
		if mode&p != 0 {
			buf[i+1] = ch[i]
		} else {
			buf[i+1] = '-'
		}
	}
	return string(buf[:])
}

// Exists checks if a path exists via SFTP.
func (c *SSHClient) Exists(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return false
	}
	if err := c.initSftp(); err != nil {
		return false
	}
	_, err := c.sftpStat(remotePath)
	return err == nil
}

// IsDir checks if path is a directory via SFTP.
func (c *SSHClient) IsDir(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return false
	}
	if err := c.initSftp(); err != nil {
		return false
	}
	info, err := c.sftpStat(remotePath)
	return err == nil && info.IsDir
}

// IsFile checks if path is a regular file via SFTP.
func (c *SSHClient) IsFile(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return false
	}
	if err := c.initSftp(); err != nil {
		return false
	}
	info, err := c.sftpStat(remotePath)
	return err == nil && !info.IsDir
}

// ListDir lists directory contents via SFTP.
func (c *SSHClient) ListDir(remotePath string) ([]map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return nil, errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return nil, err
	}

	entries, err := c.sftpListDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		result = append(result, map[string]interface{}{
			"name":  e.Name,
			"size":  e.Size,
			"isDir": e.IsDir,
			"mode":    sftpModeToString(e.Mode, e.IsDir),
		"modTime": e.ModTime,
		})
	}
	return result, nil
}

// WalkDir recursively lists all files in a directory via SFTP.
func (c *SSHClient) WalkDir(remotePath string) ([]map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return nil, errors.New("not connected")
	}
	if err := c.initSftp(); err != nil {
		return nil, err
	}

	entries, err := c.sftpWalkDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		result = append(result, map[string]interface{}{
			"path":    e.Name,
			"size":    e.Size,
			"isDir":   e.IsDir,
			"modTime": e.ModTime,
			"mode":    sftpModeToString(e.Mode, e.IsDir),
		})
	}
	return result, nil
}

// UploadDir uploads a local directory to remote via SFTP.
func (c *SSHClient) UploadDir(localDir, remoteDir string) error {
	if err := c.MkdirAll(remoteDir); err != nil {
		return err
	}
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

// DownloadDir downloads a remote directory to local via SFTP.
func (c *SSHClient) DownloadDir(remoteDir, localDir string) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

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

		if isDir, ok := file["isDir"].(bool); ok && isDir {
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
			// First connection: accept the key (no record in known_hosts yet)
			return nil
		}

		if !bytes.Equal(storedKey.Marshal(), key.Marshal()) {
			return fmt.Errorf("host key mismatch for %s", hostname)
		}

		return nil
	}
}