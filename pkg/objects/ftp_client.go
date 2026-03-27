// pkg/objects/ftp_client.go
// FtpClient object for Xxlang - FTP client functionality.
package objects

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// FtpClient represents an FTP client connection.
type FtpClient struct {
	mu        sync.Mutex
	conn      net.Conn
	text      *textproto.Conn
	host      string
	port      int
	user      string
	connected bool
	passive   bool
	binary    bool
	timeout   time.Duration
}

// Type returns the object type.
func (c *FtpClient) Type() ObjectType { return FtpClientType }

// TypeTag returns the fast type tag.
func (c *FtpClient) TypeTag() TypeTag { return TagFtpClient }

// Inspect returns a string representation of the FtpClient.
func (c *FtpClient) Inspect() string {
	if c.connected {
		return fmt.Sprintf("FtpClient(connected=%s@%s:%d)", c.user, c.host, c.port)
	}
	return "FtpClient(disconnected)"
}

// ToBool returns true if connected.
func (c *FtpClient) ToBool() *Bool {
	return &Bool{Value: c.connected}
}

// HashKey returns a hash key for the FtpClient.
func (c *FtpClient) HashKey() HashKey {
	return HashKey{
		Type:  FtpClientType,
		Value: uint64(uintptr(unsafe.Pointer(c))),
	}
}

// FtpConfig holds FTP connection configuration.
type FtpConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  int // seconds
	Passive  bool
	Binary   bool
}

// NewFtpClient creates a new FtpClient (unconnected).
func NewFtpClient() *FtpClient {
	return &FtpClient{
		connected: false,
		passive:   true, // Default to passive mode
		binary:    true, // Default to binary mode
		timeout:   30 * time.Second,
	}
}

// Connect establishes FTP connection with credentials.
func (c *FtpClient) Connect(host string, port int, user, password string) error {
	config := &FtpConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  30,
		Passive:  true,
		Binary:   true,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithConfig establishes FTP connection with full configuration.
func (c *FtpClient) ConnectWithConfig(config *FtpConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	// Set defaults
	if config.Port == 0 {
		config.Port = 21
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	c.timeout = time.Duration(config.Timeout) * time.Second
	c.passive = config.Passive
	c.binary = config.Binary

	// Connect to server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	dialer := net.Dialer{Timeout: c.timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.conn = conn
	c.text = textproto.NewConn(conn)

	// Read welcome message
	_, _, err = c.text.ReadResponse(220)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read welcome message: %w", err)
	}

	// Send USER
	err = c.sendCommandExpect("USER "+config.User, 331)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("USER command failed: %w", err)
	}

	// Send PASS
	err = c.sendCommandExpect("PASS "+config.Password, 230)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("PASS command failed: %w", err)
	}

	// Set transfer type
	if c.binary {
		err = c.sendCommandExpect("TYPE I", 200)
	} else {
		err = c.sendCommandExpect("TYPE A", 200)
	}
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("TYPE command failed: %w", err)
	}

	c.host = config.Host
	c.port = config.Port
	c.user = config.User
	c.connected = true

	return nil
}

// sendCommand sends a command and returns the response.
func (c *FtpClient) sendCommand(cmd string) (int, string, error) {
	err := c.text.PrintfLine("%s", cmd)
	if err != nil {
		return 0, "", err
	}
	return c.text.ReadResponse(0)
}

// sendCommandExpect sends a command and expects a specific response code.
func (c *FtpClient) sendCommandExpect(cmd string, expectCode int) error {
	code, _, err := c.sendCommand(cmd)
	if err != nil {
		return err
	}
	if code != expectCode {
		return fmt.Errorf("expected code %d, got %d", expectCode, code)
	}
	return nil
}

// ============================================================
// Connection Management
// ============================================================

// IsConnected returns the connection status.
func (c *FtpClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Close closes the FTP connection.
func (c *FtpClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Send QUIT command
	c.text.PrintfLine("QUIT")
	c.conn.Close()
	c.connected = false
	c.conn = nil
	c.text = nil
	return nil
}

// GetHost returns the connected host.
func (c *FtpClient) GetHost() string {
	return c.host
}

// GetPort returns the connected port.
func (c *FtpClient) GetPort() int {
	return c.port
}

// GetUser returns the username.
func (c *FtpClient) GetUser() string {
	return c.user
}

// ============================================================
// Transfer Mode Settings
// ============================================================

// SetPassive enables or disables passive mode.
func (c *FtpClient) SetPassive(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.passive = enabled
	return nil
}

// SetType sets the transfer type (binary or ascii).
func (c *FtpClient) SetType(transferType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	var cmd string
	if transferType == "binary" || transferType == "I" {
		cmd = "TYPE I"
		c.binary = true
	} else if transferType == "ascii" || transferType == "A" {
		cmd = "TYPE A"
		c.binary = false
	} else {
		return errors.New("invalid transfer type, use 'binary' or 'ascii'")
	}

	return c.sendCommandExpect(cmd, 200)
}

// ============================================================
// Data Connection
// ============================================================

// openDataConnection opens a data connection for file transfer.
func (c *FtpClient) openDataConnection() (net.Conn, error) {
	if c.passive {
		return c.openPassiveDataConnection()
	}
	return c.openActiveDataConnection()
}

// openPassiveDataConnection opens a passive mode data connection.
func (c *FtpClient) openPassiveDataConnection() (net.Conn, error) {
	// Send PASV command
	code, msg, err := c.sendCommand("PASV")
	if err != nil {
		return nil, err
	}
	if code != 227 {
		return nil, fmt.Errorf("PASV failed: %s", msg)
	}

	// Parse PASV response to get IP and port
	host, port, err := parsePasvResponse(msg)
	if err != nil {
		return nil, err
	}

	// Connect to data port
	dialer := net.Dialer{Timeout: c.timeout}
	return dialer.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
}

// openActiveDataConnection opens an active mode data connection.
func (c *FtpClient) openActiveDataConnection() (net.Conn, error) {
	// Listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	// Get local address
	localAddr := listener.Addr().(*net.TCPAddr)
	localIP := localAddr.IP

	// If connected to localhost, use 127.0.0.1
	if localIP.IsUnspecified() {
		localIP = net.ParseIP("127.0.0.1")
	}

	// Send PORT command
	h1, h2, h3, h4 := localIP[12], localIP[13], localIP[14], localIP[15]
	p1, p2 := localAddr.Port>>8, localAddr.Port&0xff
	portCmd := fmt.Sprintf("PORT %d,%d,%d,%d,%d,%d", h1, h2, h3, h4, p1, p2)

	err = c.sendCommandExpect(portCmd, 200)
	if err != nil {
		listener.Close()
		return nil, err
	}

	// Accept connection
	conn, err := listener.Accept()
	listener.Close()
	return conn, err
}

// parsePasvResponse parses PASV response to extract host and port.
func parsePasvResponse(msg string) (string, int, error) {
	// Look for pattern like (h1,h2,h3,h4,p1,p2)
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start == -1 || end == -1 || end <= start {
		return "", 0, errors.New("invalid PASV response format")
	}

	parts := strings.Split(msg[start+1:end], ",")
	if len(parts) != 6 {
		return "", 0, errors.New("invalid PASV response format")
	}

	h1, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	h2, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	h3, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	h4, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
	p1, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
	p2, _ := strconv.Atoi(strings.TrimSpace(parts[5]))

	host := fmt.Sprintf("%d.%d.%d.%d", h1, h2, h3, h4)
	port := p1<<8 | p2

	return host, port, nil
}

// ============================================================
// File Operations
// ============================================================

// Upload uploads a local file to the remote server.
func (c *FtpClient) Upload(localPath, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Open data connection
	dataConn, err := c.openDataConnection()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}
	defer dataConn.Close()

	// Send STOR command
	err = c.sendCommandExpect("STOR "+remotePath, 150)
	if err != nil {
		return fmt.Errorf("STOR command failed: %w", err)
	}

	// Transfer data
	_, err = io.Copy(dataConn, file)
	if err != nil {
		return fmt.Errorf("data transfer failed: %w", err)
	}

	dataConn.Close()

	// Read final response
	_, _, err = c.text.ReadResponse(226)
	return err
}

// Download downloads a remote file to the local filesystem.
func (c *FtpClient) Download(remotePath, localPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if localDir != "." && localDir != "" {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}

	// Open data connection
	dataConn, err := c.openDataConnection()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}
	defer dataConn.Close()

	// Send RETR command
	err = c.sendCommandExpect("RETR "+remotePath, 150)
	if err != nil {
		return fmt.Errorf("RETR command failed: %w", err)
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Transfer data
	_, err = io.Copy(file, dataConn)
	if err != nil {
		return fmt.Errorf("data transfer failed: %w", err)
	}

	// Read final response
	_, _, err = c.text.ReadResponse(226)
	return err
}

// Delete deletes a remote file.
func (c *FtpClient) Delete(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.sendCommandExpect("DELE "+remotePath, 250)
}

// Rename renames a remote file.
func (c *FtpClient) Rename(oldPath, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	err := c.sendCommandExpect("RNFR "+oldPath, 350)
	if err != nil {
		return err
	}

	return c.sendCommandExpect("RNTO "+newPath, 250)
}

// ============================================================
// Directory Operations
// ============================================================

// Mkdir creates a remote directory.
func (c *FtpClient) Mkdir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.sendCommandExpect("MKD "+remotePath, 257)
}

// MkdirAll creates a remote directory with parents.
func (c *FtpClient) MkdirAll(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Split path and create each component
	parts := strings.Split(remotePath, "/")
	path := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		if path == "" {
			path = part
		} else {
			path = path + "/" + part
		}

		// Try to create directory, ignore error if it already exists
		c.text.PrintfLine("MKD %s", path)
		c.text.ReadResponse(0)
	}

	return nil
}

// Rmdir removes an empty remote directory.
func (c *FtpClient) Rmdir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.sendCommandExpect("RMD "+remotePath, 250)
}

// RmdirAll removes a remote directory recursively.
func (c *FtpClient) RmdirAll(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Get directory listing
	files, err := c.listDirInternal(remotePath)
	if err != nil {
		return err
	}

	c.mu.Unlock()

	// Delete all files and subdirectories
	for _, file := range files {
		fullPath := remotePath + "/" + file.Name
		if file.IsDir {
			if err := c.RmdirAll(fullPath); err != nil {
				return err
			}
		} else {
			if err := c.Delete(fullPath); err != nil {
				return err
			}
		}
	}

	c.mu.Lock()

	// Remove the directory itself
	return c.sendCommandExpect("RMD "+remotePath, 250)
}

// ChangeDir changes the current working directory.
func (c *FtpClient) ChangeDir(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.sendCommandExpect("CWD "+path, 250)
}

// CurrentDir returns the current working directory.
func (c *FtpClient) CurrentDir() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	code, msg, err := c.sendCommand("PWD")
	if err != nil {
		return "", err
	}
	if code != 257 {
		return "", fmt.Errorf("PWD failed: %s", msg)
	}

	// Extract path from response (usually in quotes)
	start := strings.Index(msg, "\"")
	end := strings.LastIndex(msg, "\"")
	if start != -1 && end > start {
		return msg[start+1 : end], nil
	}
	return msg, nil
}

// FileInfo represents file information from FTP listing.
type FtpFileInfo struct {
	Name  string
	Size  int64
	IsDir bool
	Mode  string
}

// ListDir lists the contents of a remote directory.
func (c *FtpClient) ListDir(remotePath string) ([]FtpFileInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	return c.listDirInternal(remotePath)
}

// listDirInternal is the internal implementation of ListDir.
func (c *FtpClient) listDirInternal(remotePath string) ([]FtpFileInfo, error) {
	// Open data connection
	dataConn, err := c.openDataConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to open data connection: %w", err)
	}
	defer dataConn.Close()

	// Send LIST command
	err = c.sendCommandExpect("LIST "+remotePath, 150)
	if err != nil {
		return nil, fmt.Errorf("LIST command failed: %w", err)
	}

	// Read directory listing
	var lines []string
	scanner := bufio.NewScanner(dataConn)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Read final response
	c.text.ReadResponse(226)

	// Parse listing
	var result []FtpFileInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		info := parseListLine(line)
		if info.Name != "" && info.Name != "." && info.Name != ".." {
			result = append(result, info)
		}
	}

	return result, nil
}

// parseListLine parses a line from LIST output.
func parseListLine(line string) FtpFileInfo {
	// Common format: drwxr-xr-x  2 user group 4096 Jan 01 00:00 dirname
	// or: -rw-r--r--  1 user group 1234 Jan 01 00:00 filename
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return FtpFileInfo{}
	}

	info := FtpFileInfo{}
	info.Mode = fields[0]
	info.IsDir = strings.HasPrefix(fields[0], "d")

	// Size is usually at index 4 for most formats
	if len(fields) >= 5 {
		size, _ := strconv.ParseInt(fields[4], 10, 64)
		info.Size = size
	}

	// Name is everything after the date/time
	if len(fields) >= 9 {
		info.Name = strings.Join(fields[8:], " ")
	}

	return info
}

// ============================================================
// File Information
// ============================================================

// Size returns the size of a remote file.
func (c *FtpClient) Size(remotePath string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return 0, errors.New("not connected")
	}

	code, msg, err := c.sendCommand("SIZE " + remotePath)
	if err != nil {
		return 0, err
	}
	if code != 213 {
		return 0, fmt.Errorf("SIZE failed: %s", msg)
	}

	size, err := strconv.ParseInt(strings.TrimSpace(msg), 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// ModTime returns the modification time of a remote file.
func (c *FtpClient) ModTime(remotePath string) (time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return time.Time{}, errors.New("not connected")
	}

	code, msg, err := c.sendCommand("MDTM " + remotePath)
	if err != nil {
		return time.Time{}, err
	}
	if code != 213 {
		return time.Time{}, fmt.Errorf("MDTM failed: %s", msg)
	}

	// Parse MDTM format: YYYYMMDDHHMMSS
	return time.Parse("20060102150405", strings.TrimSpace(msg))
}

// Exists checks if a remote path exists.
func (c *FtpClient) Exists(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return false
	}

	// Try SIZE command (works for files)
	code, _, _ := c.sendCommand("SIZE " + remotePath)
	if code == 213 {
		return true
	}

	// Try CWD (works for directories)
	c.text.PrintfLine("CWD %s", remotePath)
	code, _, _ = c.text.ReadResponse(0)
	if code == 250 {
		// Go back to original directory
		c.text.PrintfLine("CDUP")
		c.text.ReadResponse(0)
		return true
	}

	return false
}

// IsDir checks if a remote path is a directory.
func (c *FtpClient) IsDir(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return false
	}

	c.text.PrintfLine("CWD %s", remotePath)
	code, _, _ := c.text.ReadResponse(0)
	if code == 250 {
		c.text.PrintfLine("CDUP")
		c.text.ReadResponse(0)
		return true
	}
	return false
}

// IsFile checks if a remote path is a file.
func (c *FtpClient) IsFile(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return false
	}

	code, _, _ := c.sendCommand("SIZE " + remotePath)
	return code == 213
}
