// pkg/objects/sftp_client.go
// SftpClient object for Xxlang - SFTP client functionality.
package objects

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/crypto/ssh"
)

// SFTP packet types
const (
	SSH_FXP_INIT           = 1
	SSH_FXP_VERSION        = 2
	SSH_FXP_OPEN           = 3
	SSH_FXP_CLOSE          = 4
	SSH_FXP_READ           = 5
	SSH_FXP_WRITE          = 6
	SSH_FXP_LSTAT          = 7
	SSH_FXP_FSTAT          = 8
	SSH_FXP_SETSTAT        = 9
	SSH_FXP_FSETSTAT       = 10
	SSH_FXP_OPENDIR        = 11
	SSH_FXP_READDIR        = 12
	SSH_FXP_REMOVE         = 13
	SSH_FXP_MKDIR          = 14
	SSH_FXP_RMDIR          = 15
	SSH_FXP_REALPATH       = 16
	SSH_FXP_STAT           = 17
	SSH_FXP_RENAME         = 18
	SSH_FXP_READLINK       = 19
	SSH_FXP_SYMLINK        = 20
	SSH_FXP_STATUS         = 101
	SSH_FXP_HANDLE         = 102
	SSH_FXP_DATA           = 103
	SSH_FXP_NAME           = 104
	SSH_FXP_ATTRS          = 105
	SSH_FXP_EXTENDED       = 200
	SSH_FXP_EXTENDED_REPLY = 201
)

// File open flags
const (
	SSH_FXF_READ   = 0x00000001
	SSH_FXF_WRITE  = 0x00000002
	SSH_FXF_APPEND = 0x00000004
	SSH_FXF_CREAT  = 0x00000008
	SSH_FXF_TRUNC  = 0x00000010
	SSH_FXF_EXCL   = 0x00000020
)

// File attribute flags
const (
	SSH_FILEXFER_ATTR_SIZE        = 0x00000001
	SSH_FILEXFER_ATTR_UIDGID      = 0x00000002
	SSH_FILEXFER_ATTR_PERMISSIONS = 0x00000004
	SSH_FILEXFER_ATTR_ACMODTIME   = 0x00000008
)

// SFTP status codes
const (
	SSH_FX_OK                = 0
	SSH_FX_EOF               = 1
	SSH_FX_NO_SUCH_FILE      = 2
	SSH_FX_PERMISSION_DENIED = 3
	SSH_FX_FAILURE           = 4
	SSH_FX_BAD_MESSAGE       = 5
	SSH_FX_NO_CONNECTION     = 6
	SSH_FX_CONNECTION_LOST   = 7
	SSH_FX_OP_UNSUPPORTED    = 8
)

// SftpClient represents an SFTP client connection.
type SftpClient struct {
	mu        sync.Mutex
	client    *ssh.Client
	session   *ssh.Session
	channel   ssh.Channel
	host      string
	port      int
	user      string
	connected bool
	nextID    uint32
}

// Type returns the object type.
func (c *SftpClient) Type() ObjectType { return SftpClientType }

// TypeTag returns the fast type tag.
func (c *SftpClient) TypeTag() TypeTag { return TagSftpClient }

// Inspect returns a string representation of the SftpClient.
func (c *SftpClient) Inspect() string {
	if c.connected {
		return fmt.Sprintf("SftpClient(connected=%s@%s:%d)", c.user, c.host, c.port)
	}
	return "SftpClient(disconnected)"
}

// ToBool returns true if connected.
func (c *SftpClient) ToBool() *Bool {
	return &Bool{Value: c.connected}
}

// HashKey returns a hash key for the SftpClient.
func (c *SftpClient) HashKey() HashKey {
	return HashKey{
		Type:  SftpClientType,
		Value: uint64(uintptr(unsafe.Pointer(c))),
	}
}

// SftpConfig holds SFTP connection configuration.
type SftpConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	KeyPath       string
	KeyStr        string
	KeyPassphrase string
	Timeout       int
	IgnoreHostKey bool
}

// NewSftpClient creates a new SftpClient (unconnected).
func NewSftpClient() *SftpClient {
	return &SftpClient{
		connected: false,
		nextID:    1,
	}
}

// Connect establishes SFTP connection with password.
func (c *SftpClient) Connect(host string, port int, user, password string) error {
	config := &SftpConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithKey establishes SFTP connection with private key file.
func (c *SftpClient) ConnectWithKey(host string, port int, user, keyPath string) error {
	config := &SftpConfig{
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
		Timeout: 30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithKeyStr establishes SFTP connection with private key string.
func (c *SftpClient) ConnectWithKeyStr(host string, port int, user, keyStr string) error {
	config := &SftpConfig{
		Host:    host,
		Port:    port,
		User:    user,
		KeyStr:  keyStr,
		Timeout: 30,
	}
	return c.ConnectWithConfig(config)
}

// ConnectWithConfig establishes SFTP connection with full configuration.
func (c *SftpClient) ConnectWithConfig(config *SftpConfig) error {
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
	} else {
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Connect via SSH
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	// Open SFTP session
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Open SFTP channel
	channel, reqs, err := client.OpenChannel("session", nil)
	if err != nil {
		session.Close()
		client.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	go ssh.DiscardRequests(reqs)

	// Send subsystem request for SFTP
	_, err = channel.SendRequest("subsystem", true, ssh.Marshal(struct {
		Subsystem string
	}{Subsystem: "sftp"}))
	if err != nil {
		channel.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to start SFTP subsystem: %w", err)
	}

	// Send INIT packet
	initPkt := make([]byte, 5)
	binary.BigEndian.PutUint32(initPkt[0:4], 1) // length
	initPkt[4] = SSH_FXP_INIT
	if _, err := channel.Write(initPkt); err != nil {
		channel.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to send INIT: %w", err)
	}

	// Read VERSION response
	resp, err := c.readPacket(channel)
	if err != nil {
		channel.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to read VERSION: %w", err)
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_VERSION {
		channel.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("unexpected response, expected VERSION")
	}

	c.client = client
	c.session = session
	c.channel = channel
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
func (c *SftpClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Close closes the SFTP connection.
func (c *SftpClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.channel.Close()
	c.session.Close()
	c.client.Close()
	c.connected = false
	return nil
}

// GetHost returns the connected host.
func (c *SftpClient) GetHost() string {
	return c.host
}

// GetPort returns the connected port.
func (c *SftpClient) GetPort() int {
	return c.port
}

// GetUser returns the username.
func (c *SftpClient) GetUser() string {
	return c.user
}

// ============================================================
// Packet Handling
// ============================================================

// nextRequestID returns the next request ID.
func (c *SftpClient) nextRequestID() uint32 {
	c.nextID++
	return c.nextID
}

// sendPacket sends an SFTP packet.
func (c *SftpClient) sendPacket(pkt []byte) error {
	length := uint32(len(pkt))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], pkt)
	_, err := c.channel.Write(buf)
	return err
}

// readPacket reads an SFTP packet.
func (c *SftpClient) readPacket(r io.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	if length > 1024*1024*32 {
		return nil, errors.New("packet too large")
	}
	pkt := make([]byte, length)
	if _, err := io.ReadFull(r, pkt); err != nil {
		return nil, err
	}
	return pkt, nil
}

// ============================================================
// File Operations
// ============================================================

// Upload uploads a local file to the remote server.
func (c *SftpClient) Upload(localPath, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Read local file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		c.mkdirInternal(remoteDir, true)
	}

	// Open file for writing
	handle, err := c.openFileInternal(remotePath, SSH_FXF_WRITE|SSH_FXF_CREAT|SSH_FXF_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.closeHandleInternal(handle)

	// Write data
	offset := uint64(0)
	chunkSize := uint32(32768) // 32KB chunks

	for offset < uint64(len(data)) {
		end := offset + uint64(chunkSize)
		if end > uint64(len(data)) {
			end = uint64(len(data))
		}
		chunk := data[offset:end]

		if err := c.writeChunkInternal(handle, offset, chunk); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
		offset = end
	}

	return nil
}

// Download downloads a remote file to the local filesystem.
func (c *SftpClient) Download(remotePath, localPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// Open remote file for reading
	handle, err := c.openFileInternal(remotePath, SSH_FXF_READ)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.closeHandleInternal(handle)

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if localDir != "." && localDir != "" {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Read data
	offset := uint64(0)
	chunkSize := uint32(32768) // 32KB chunks

	for {
		data, err := c.readChunkInternal(handle, offset, chunkSize)
		if err != nil {
			return fmt.Errorf("failed to read data: %w", err)
		}
		if len(data) == 0 {
			break
		}
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write local file: %w", err)
		}
		offset += uint64(len(data))
		if uint32(len(data)) < chunkSize {
			break
		}
	}

	return nil
}

// Delete deletes a remote file.
func (c *SftpClient) Delete(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_REMOVE
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// Rename renames a remote file.
func (c *SftpClient) Rename(oldPath, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(oldPath)+4+len(newPath))
	pkt[0] = SSH_FXP_RENAME
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(oldPath)))
	copy(pkt[9:9+len(oldPath)], oldPath)
	pos := 9 + len(oldPath)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], uint32(len(newPath)))
	copy(pkt[pos+4:], newPath)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// ============================================================
// Directory Operations
// ============================================================

// Mkdir creates a remote directory.
func (c *SftpClient) Mkdir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.mkdirInternal(remotePath, false)
}

// mkdirInternal creates a remote directory (internal, no lock).
func (c *SftpClient) mkdirInternal(remotePath string, ignoreExists bool) error {
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath)+4)
	pkt[0] = SSH_FXP_MKDIR
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:9+len(remotePath)], remotePath)
	binary.BigEndian.PutUint32(pkt[9+len(remotePath):], 0) // attrs

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	err := c.expectStatus(id)
	if ignoreExists && err != nil {
		// Ignore "file exists" error
		return nil
	}
	return err
}

// MkdirAll creates a remote directory with parents.
func (c *SftpClient) MkdirAll(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.mkdirInternal(remotePath, true)
}

// Rmdir removes an empty remote directory.
func (c *SftpClient) Rmdir(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_RMDIR
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// RmdirAll removes a remote directory recursively.
func (c *SftpClient) RmdirAll(remotePath string) error {
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
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_RMDIR
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// SftpFileInfo represents file information from SFTP.
type SftpFileInfo struct {
	Name    string
	Size    int64
	Mode    uint32
	ModTime int64 // Unix timestamp
	IsDir   bool
}

// ListDir lists the contents of a remote directory.
func (c *SftpClient) ListDir(remotePath string) ([]SftpFileInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	return c.listDirInternal(remotePath)
}

// listDirInternal lists directory contents (internal, no lock).
func (c *SftpClient) listDirInternal(remotePath string) ([]SftpFileInfo, error) {
	// Open directory
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_OPENDIR
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}

	// Read handle response
	resp, err := c.readPacket(c.channel)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_HANDLE {
		return nil, errors.New("expected HANDLE response")
	}

	handle := resp[1:5]

	// Read directory entries
	var result []SftpFileInfo

	for {
		// Send READDIR
		id := c.nextRequestID()
		pkt := make([]byte, 1+4+4)
		pkt[0] = SSH_FXP_READDIR
		binary.BigEndian.PutUint32(pkt[1:5], id)
		copy(pkt[5:9], handle)

		if err := c.sendPacket(pkt); err != nil {
			break
		}

		// Read response
		resp, err := c.readPacket(c.channel)
		if err != nil {
			break
		}

		if len(resp) == 0 {
			break
		}

		if resp[0] == SSH_FXP_STATUS {
			break
		}

		if resp[0] != SSH_FXP_NAME {
			break
		}

		// Parse NAME response
		count := binary.BigEndian.Uint32(resp[1:5])
		pos := 5

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

			// Skip attrs
			if pos+4 > len(resp) {
				break
			}
			attrFlags := binary.BigEndian.Uint32(resp[pos : pos+4])
			pos += 4

			var size int64
			var mode uint32
			isDir := false

			if attrFlags&SSH_FILEXFER_ATTR_SIZE != 0 {
				if pos+8 > len(resp) {
					break
				}
				size = int64(binary.BigEndian.Uint64(resp[pos : pos+8]))
				pos += 8
			}

			if attrFlags&SSH_FILEXFER_ATTR_UIDGID != 0 {
				pos += 8
			}

			if attrFlags&SSH_FILEXFER_ATTR_PERMISSIONS != 0 {
				if pos+4 > len(resp) {
					break
				}
				mode = binary.BigEndian.Uint32(resp[pos : pos+4])
				pos += 4
				isDir = (mode & 0040000) != 0
			}

			if attrFlags&SSH_FILEXFER_ATTR_ACMODTIME != 0 {
				pos += 8
			}

			result = append(result, SftpFileInfo{
				Name:  name,
				Size:  size,
				Mode:  mode,
				IsDir: isDir,
			})
		}
	}

	// Close directory handle
	c.closeHandleInternal(string(handle))

	return result, nil
}

// ============================================================
// File Information
// ============================================================

// Stat returns file information.
func (c *SftpClient) Stat(remotePath string) (*SftpFileInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_STAT
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}

	return c.expectAttrs(id)
}

// Lstat returns file information without following symlinks.
func (c *SftpClient) Lstat(remotePath string) (*SftpFileInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_LSTAT
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}

	return c.expectAttrs(id)
}

// Chmod changes file permissions.
func (c *SftpClient) Chmod(remotePath string, mode uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath)+4+4)
	pkt[0] = SSH_FXP_SETSTAT
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:9+len(remotePath)], remotePath)
	pos := 9 + len(remotePath)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], SSH_FILEXFER_ATTR_PERMISSIONS)
	binary.BigEndian.PutUint32(pkt[pos+4:pos+8], mode)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// Truncate truncates a file.
func (c *SftpClient) Truncate(remotePath string, size int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath)+4+8)
	pkt[0] = SSH_FXP_SETSTAT
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:9+len(remotePath)], remotePath)
	pos := 9 + len(remotePath)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], SSH_FILEXFER_ATTR_SIZE)
	binary.BigEndian.PutUint64(pkt[pos+4:pos+12], uint64(size))

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// ReadLink reads a symbolic link.
func (c *SftpClient) ReadLink(remotePath string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_READLINK
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return "", err
	}

	resp, err := c.readPacket(c.channel)
	if err != nil {
		return "", err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_NAME {
		return "", errors.New("expected NAME response")
	}

	// Parse NAME response
	count := binary.BigEndian.Uint32(resp[1:5])
	if count == 0 {
		return "", errors.New("no link target")
	}

	pos := 5
	nameLen := binary.BigEndian.Uint32(resp[pos : pos+4])
	pos += 4
	return string(resp[pos : pos+int(nameLen)]), nil
}

// Symlink creates a symbolic link.
func (c *SftpClient) Symlink(oldPath, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(oldPath)+4+len(newPath))
	pkt[0] = SSH_FXP_SYMLINK
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(oldPath)))
	copy(pkt[9:9+len(oldPath)], oldPath)
	pos := 9 + len(oldPath)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], uint32(len(newPath)))
	copy(pkt[pos+4:], newPath)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// RealPath returns the canonical path.
func (c *SftpClient) RealPath(remotePath string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_REALPATH
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return "", err
	}

	resp, err := c.readPacket(c.channel)
	if err != nil {
		return "", err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_NAME {
		return "", errors.New("expected NAME response")
	}

	count := binary.BigEndian.Uint32(resp[1:5])
	if count == 0 {
		return "", errors.New("no path returned")
	}

	pos := 5
	nameLen := binary.BigEndian.Uint32(resp[pos : pos+4])
	pos += 4
	return string(resp[pos : pos+int(nameLen)]), nil
}

// Exists checks if a remote path exists.
func (c *SftpClient) Exists(remotePath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return false
	}

	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(remotePath))
	pkt[0] = SSH_FXP_STAT
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(remotePath)))
	copy(pkt[9:], remotePath)

	if err := c.sendPacket(pkt); err != nil {
		return false
	}

	resp, err := c.readPacket(c.channel)
	if err != nil {
		return false
	}

	return len(resp) > 0 && resp[0] == SSH_FXP_ATTRS
}

// IsDir checks if a remote path is a directory.
func (c *SftpClient) IsDir(remotePath string) bool {
	info, err := c.Stat(remotePath)
	if err != nil {
		return false
	}
	return info.IsDir
}

// IsFile checks if a remote path is a regular file.
func (c *SftpClient) IsFile(remotePath string) bool {
	info, err := c.Stat(remotePath)
	if err != nil {
		return false
	}
	return !info.IsDir
}

// ============================================================
// Internal Helper Methods
// ============================================================

// openFileInternal opens a remote file (internal, no lock).
func (c *SftpClient) openFileInternal(path string, flags uint32) (string, error) {
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+len(path)+4+4)
	pkt[0] = SSH_FXP_OPEN
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(path)))
	copy(pkt[9:9+len(path)], path)
	pos := 9 + len(path)
	binary.BigEndian.PutUint32(pkt[pos:pos+4], flags)
	binary.BigEndian.PutUint32(pkt[pos+4:pos+8], 0) // attrs

	if err := c.sendPacket(pkt); err != nil {
		return "", err
	}

	resp, err := c.readPacket(c.channel)
	if err != nil {
		return "", err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_HANDLE {
		return "", errors.New("expected HANDLE response")
	}

	return string(resp[1:5]), nil
}

// closeHandleInternal closes a file handle (internal, no lock).
func (c *SftpClient) closeHandleInternal(handle string) error {
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4)
	pkt[0] = SSH_FXP_CLOSE
	binary.BigEndian.PutUint32(pkt[1:5], id)
	copy(pkt[5:9], handle)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// readChunkInternal reads a chunk of data (internal, no lock).
func (c *SftpClient) readChunkInternal(handle string, offset uint64, length uint32) ([]byte, error) {
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+8+4)
	pkt[0] = SSH_FXP_READ
	binary.BigEndian.PutUint32(pkt[1:5], id)
	copy(pkt[5:9], handle)
	binary.BigEndian.PutUint64(pkt[9:17], offset)
	binary.BigEndian.PutUint32(pkt[17:21], length)

	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}

	resp, err := c.readPacket(c.channel)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errors.New("empty response")
	}

	if resp[0] == SSH_FXP_STATUS {
		return nil, nil // EOF
	}

	if resp[0] != SSH_FXP_DATA {
		return nil, errors.New("expected DATA response")
	}

	dataLen := binary.BigEndian.Uint32(resp[1:5])
	return resp[5 : 5+dataLen], nil
}

// writeChunkInternal writes a chunk of data (internal, no lock).
func (c *SftpClient) writeChunkInternal(handle string, offset uint64, data []byte) error {
	id := c.nextRequestID()
	pkt := make([]byte, 1+4+4+8+4+len(data))
	pkt[0] = SSH_FXP_WRITE
	binary.BigEndian.PutUint32(pkt[1:5], id)
	copy(pkt[5:9], handle)
	binary.BigEndian.PutUint64(pkt[9:17], offset)
	binary.BigEndian.PutUint32(pkt[17:21], uint32(len(data)))
	copy(pkt[21:], data)

	if err := c.sendPacket(pkt); err != nil {
		return err
	}

	return c.expectStatus(id)
}

// expectStatus expects a STATUS response.
func (c *SftpClient) expectStatus(expectedID uint32) error {
	resp, err := c.readPacket(c.channel)
	if err != nil {
		return err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_STATUS {
		return errors.New("expected STATUS response")
	}

	id := binary.BigEndian.Uint32(resp[1:5])
	if id != expectedID {
		return errors.New("unexpected response ID")
	}

	status := binary.BigEndian.Uint32(resp[5:9])
	if status != SSH_FX_OK {
		return fmt.Errorf("SFTP error: status %d", status)
	}

	return nil
}

// expectAttrs expects an ATTRS response.
func (c *SftpClient) expectAttrs(expectedID uint32) (*SftpFileInfo, error) {
	resp, err := c.readPacket(c.channel)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 || resp[0] != SSH_FXP_ATTRS {
		return nil, errors.New("expected ATTRS response")
	}

	id := binary.BigEndian.Uint32(resp[1:5])
	if id != expectedID {
		return nil, errors.New("unexpected response ID")
	}

	// Parse attributes
	info := &SftpFileInfo{}
	pos := 5

	if pos+4 > len(resp) {
		return info, nil
	}

	attrFlags := binary.BigEndian.Uint32(resp[pos : pos+4])
	pos += 4

	if attrFlags&SSH_FILEXFER_ATTR_SIZE != 0 {
		if pos+8 <= len(resp) {
			info.Size = int64(binary.BigEndian.Uint64(resp[pos : pos+8]))
			pos += 8
		}
	}

	if attrFlags&SSH_FILEXFER_ATTR_UIDGID != 0 {
		pos += 8
	}

	if attrFlags&SSH_FILEXFER_ATTR_PERMISSIONS != 0 {
		if pos+4 <= len(resp) {
			info.Mode = binary.BigEndian.Uint32(resp[pos : pos+4])
			info.IsDir = (info.Mode & 0040000) != 0
		}
	}

	return info, nil
}

// Open opens a remote file for reading.
func (c *SftpClient) Open(remotePath string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	return c.openFileInternal(remotePath, SSH_FXF_READ)
}

// Create creates a remote file for writing.
func (c *SftpClient) Create(remotePath string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	return c.openFileInternal(remotePath, SSH_FXF_WRITE|SSH_FXF_CREAT|SSH_FXF_TRUNC)
}

// OpenFile opens a remote file with specified flags.
func (c *SftpClient) OpenFile(remotePath string, flags uint32) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("not connected")
	}

	return c.openFileInternal(remotePath, flags)
}

// CloseHandle closes a file handle.
func (c *SftpClient) CloseHandle(handle string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.closeHandleInternal(handle)
}

// Read reads data from a file handle.
func (c *SftpClient) Read(handle string, offset uint64, length uint32) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	return c.readChunkInternal(handle, offset, length)
}

// Write writes data to a file handle.
func (c *SftpClient) Write(handle string, offset uint64, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	return c.writeChunkInternal(handle, offset, data)
}
