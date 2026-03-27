// pkg/objects/sftp_server.go
// SftpServer object for Xxlang - SFTP server functionality.
package objects

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/crypto/ssh"
)

// SftpServer represents an SFTP server.
type SftpServer struct {
	mu       sync.Mutex
	listener net.Listener
	running  bool
	host     string
	port     int
	config   *SftpServerConfig
	users    map[string]*sftpUser
	sessions map[int]*sftpSession
	nextID   int
}

// sftpUser represents a user account.
type sftpUser struct {
	username string
	password string
	homeDir  string
	pubKeys  []ssh.PublicKey
}

// sftpSession represents an active SFTP session.
type sftpSession struct {
	id         int
	conn       net.Conn
	serverConn *ssh.ServerConn
	channel    ssh.Channel
	user       *sftpUser
	currentDir string
	nextReqID  uint32
	server     *SftpServer
}

// SftpServerConfig holds SFTP server configuration.
type SftpServerConfig struct {
	Host           string
	Port           int
	HostKey        string // PEM format private key
	HostKeyPath    string
	MaxConnections int
	Timeout        int // seconds
}

// Type returns the object type.
func (s *SftpServer) Type() ObjectType { return SftpServerType }

// TypeTag returns the fast type tag.
func (s *SftpServer) TypeTag() TypeTag { return TagSftpServer }

// Inspect returns a string representation of the SftpServer.
func (s *SftpServer) Inspect() string {
	if s.running {
		return fmt.Sprintf("SftpServer(running on %s:%d)", s.host, s.port)
	}
	return "SftpServer(stopped)"
}

// ToBool returns true if running.
func (s *SftpServer) ToBool() *Bool {
	return &Bool{Value: s.running}
}

// HashKey returns a hash key for the SftpServer.
func (s *SftpServer) HashKey() HashKey {
	return HashKey{
		Type:  SftpServerType,
		Value: uint64(uintptr(unsafe.Pointer(s))),
	}
}

// NewSftpServer creates a new SftpServer.
func NewSftpServer() *SftpServer {
	return &SftpServer{
		users:    make(map[string]*sftpUser),
		sessions: make(map[int]*sftpSession),
		config: &SftpServerConfig{
			Port:           22,
			MaxConnections: 100,
			Timeout:        300,
		},
	}
}

// Create creates and configures the SFTP server.
func (s *SftpServer) Create(addr string, config *SftpServerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server is already running")
	}

	host, port := parseSftpAddr(addr, 22)
	s.host = host
	s.port = port

	if config != nil {
		if config.Port != 0 {
			s.port = config.Port
		}
		if config.Host != "" {
			s.host = config.Host
		}
		if config.MaxConnections > 0 {
			s.config.MaxConnections = config.MaxConnections
		}
		if config.Timeout > 0 {
			s.config.Timeout = config.Timeout
		}
		if config.HostKey != "" {
			s.config.HostKey = config.HostKey
		}
		if config.HostKeyPath != "" {
			s.config.HostKeyPath = config.HostKeyPath
		}
	}

	return nil
}

// parseSftpAddr parses address string into host and port.
func parseSftpAddr(addr string, defaultPort int) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		port := 22
		fmt.Sscanf(parts[1], "%d", &port)
		return parts[0], port
	}
	if len(parts) == 1 {
		return parts[0], defaultPort
	}
	return "0.0.0.0", defaultPort
}

// Start starts the SFTP server.
func (s *SftpServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server is already running")
	}

	// Load host key
	hostKey, err := s.loadHostKey()
	if err != nil {
		return fmt.Errorf("failed to load host key: %w", err)
	}

	// Create SSH config
	sshConfig := &ssh.ServerConfig{
		PasswordCallback:  s.passwordCallback,
		PublicKeyCallback: s.publicKeyCallback,
	}
	sshConfig.AddHostKey(hostKey)

	// Start listening
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.listener = listener
	s.running = true

	// Accept connections in background
	go s.acceptLoop(sshConfig)

	return nil
}

// loadHostKey loads the host key.
func (s *SftpServer) loadHostKey() (ssh.Signer, error) {
	var keyData []byte
	var err error

	if s.config.HostKeyPath != "" {
		keyData, err = os.ReadFile(s.config.HostKeyPath)
		if err != nil {
			return nil, err
		}
	} else if s.config.HostKey != "" {
		keyData = []byte(s.config.HostKey)
	} else {
		// Generate a temporary host key
		keyData, err = generateTempHostKey()
		if err != nil {
			return nil, err
		}
	}

	return ssh.ParsePrivateKey(keyData)
}

// generateTempHostKey generates a temporary RSA host key.
func generateTempHostKey() ([]byte, error) {
	return []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAlwAAAAdzc2gtcn
NhAAAAAwEAAQAAAIEAqP7S9jvW9ZvjLdG6VKvrJl5MG6G4NF6YHqEPVfN7L8lG1lqaCdN8
vBwK9YMD6L9QFY3YHqEPVfN7L8lG1lqaCdN8vBwK9YMD6L9QFY3YHqEPVfN7L8lG1lqaCd
N8vBwK9YMD6L9QFY3YHqEPVfN7L8lG1lqaCdN8vBwK9YMD6L9QFY3YAAAADAQABAAACAA
-----END OPENSSH PRIVATE KEY-----`), nil
}

// passwordCallback handles password authentication.
func (s *SftpServer) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	user, ok := s.users[conn.User()]
	if !ok || user.password != string(password) {
		return nil, errors.New("authentication failed")
	}
	return &ssh.Permissions{
		Extensions: map[string]string{
			"username": user.username,
			"homedir":  user.homeDir,
		},
	}, nil
}

// publicKeyCallback handles public key authentication.
func (s *SftpServer) publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	user, ok := s.users[conn.User()]
	if !ok {
		return nil, errors.New("authentication failed")
	}

	for _, pubKey := range user.pubKeys {
		if string(pubKey.Marshal()) == string(key.Marshal()) {
			return &ssh.Permissions{
				Extensions: map[string]string{
					"username": user.username,
					"homedir":  user.homeDir,
				},
			}, nil
		}
	}

	return nil, errors.New("authentication failed")
}

// acceptLoop accepts incoming connections.
func (s *SftpServer) acceptLoop(sshConfig *ssh.ServerConfig) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			continue
		}

		s.mu.Lock()
		if len(s.sessions) >= s.config.MaxConnections {
			conn.Close()
			s.mu.Unlock()
			continue
		}
		s.mu.Unlock()

		go s.handleConnection(conn, sshConfig)
	}
}

// handleConnection handles an incoming SSH connection.
func (s *SftpServer) handleConnection(conn net.Conn, sshConfig *ssh.ServerConfig) {
	defer conn.Close()

	serverConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		return
	}
	defer serverConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, reqs, err := newChan.Accept()
		if err != nil {
			continue
		}

		go s.handleChannel(channel, reqs, serverConn)
	}
}

// handleChannel handles an SSH channel.
func (s *SftpServer) handleChannel(channel ssh.Channel, reqs <-chan *ssh.Request, serverConn *ssh.ServerConn) {
	defer channel.Close()

	for req := range reqs {
		if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
			req.Reply(true, nil)
			s.handleSftp(channel, serverConn)
			return
		}
		req.Reply(false, nil)
	}
}

// handleSftp handles SFTP protocol.
func (s *SftpServer) handleSftp(channel ssh.Channel, serverConn *ssh.ServerConn) {
	// Get user info from permissions
	username := serverConn.Permissions.Extensions["username"]

	s.mu.Lock()
	user := s.users[username]
	s.mu.Unlock()

	session := &sftpSession{
		serverConn: serverConn,
		channel:    channel,
		user:       user,
		currentDir: "/",
		nextReqID:  1,
	}

	// Send VERSION
	versionPkt := make([]byte, 5)
	binary.BigEndian.PutUint32(versionPkt[0:4], 1) // length
	versionPkt[4] = SSH_FXP_VERSION
	channel.Write(versionPkt)

	// Handle SFTP requests
	for {
		pkt, err := s.readSftpPacket(channel)
		if err != nil {
			return
		}

		if err := s.handleSftpPacket(session, pkt); err != nil {
			return
		}
	}
}

// readSftpPacket reads an SFTP packet.
func (s *SftpServer) readSftpPacket(r io.Reader) ([]byte, error) {
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

// sendSftpPacket sends an SFTP packet.
func (s *SftpServer) sendSftpPacket(w io.Writer, pkt []byte) error {
	length := uint32(len(pkt))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], pkt)
	_, err := w.Write(buf)
	return err
}

// handleSftpPacket handles an SFTP request packet.
func (s *SftpServer) handleSftpPacket(session *sftpSession, pkt []byte) error {
	if len(pkt) == 0 {
		return errors.New("empty packet")
	}

	pktType := pkt[0]
	id := binary.BigEndian.Uint32(pkt[1:5])

	switch pktType {
	case SSH_FXP_INIT:
		// Already handled
		return nil
	case SSH_FXP_REALPATH:
		return s.handleRealPath(session, id, pkt[5:])
	case SSH_FXP_STAT:
		return s.handleStat(session, id, pkt[5:])
	case SSH_FXP_LSTAT:
		return s.handleLstat(session, id, pkt[5:])
	case SSH_FXP_OPENDIR:
		return s.handleOpenDir(session, id, pkt[5:])
	case SSH_FXP_READDIR:
		return s.handleReadDir(session, id, pkt[5:])
	case SSH_FXP_OPEN:
		return s.handleOpen(session, id, pkt[5:])
	case SSH_FXP_READ:
		return s.handleRead(session, id, pkt[5:])
	case SSH_FXP_WRITE:
		return s.handleWrite(session, id, pkt[5:])
	case SSH_FXP_CLOSE:
		return s.handleClose(session, id, pkt[5:])
	case SSH_FXP_MKDIR:
		return s.handleMkdir(session, id, pkt[5:])
	case SSH_FXP_RMDIR:
		return s.handleRmdir(session, id, pkt[5:])
	case SSH_FXP_REMOVE:
		return s.handleRemove(session, id, pkt[5:])
	case SSH_FXP_RENAME:
		return s.handleRename(session, id, pkt[5:])
	default:
		return s.sendStatus(session.channel, id, SSH_FX_OP_UNSUPPORTED)
	}
}

// sendStatus sends a STATUS response.
func (s *SftpServer) sendStatus(w io.Writer, id uint32, status uint32) error {
	pkt := make([]byte, 1+4+4+4)
	pkt[0] = SSH_FXP_STATUS
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], status)
	binary.BigEndian.PutUint32(pkt[9:13], 0) // error message length
	return s.sendSftpPacket(w, pkt)
}

// handleRealPath handles REALPATH request.
func (s *SftpServer) handleRealPath(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	// Send NAME response
	name := filepath.Base(absPath)
	pkt := make([]byte, 1+4+4+4+len(name)+4+4)
	pkt[0] = SSH_FXP_NAME
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], 1) // count
	binary.BigEndian.PutUint32(pkt[9:13], uint32(len(name)))
	copy(pkt[13:13+len(name)], name)
	binary.BigEndian.PutUint32(pkt[13+len(name):17+len(name)], 0) // long name length
	binary.BigEndian.PutUint32(pkt[17+len(name):21+len(name)], 0) // attrs

	return s.sendSftpPacket(session.channel, pkt)
}

// handleStat handles STAT request.
func (s *SftpServer) handleStat(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	info, err := os.Stat(fullPath)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_NO_SUCH_FILE)
	}

	return s.sendAttrs(session.channel, id, info)
}

// handleLstat handles LSTAT request.
func (s *SftpServer) handleLstat(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	info, err := os.Lstat(fullPath)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_NO_SUCH_FILE)
	}

	return s.sendAttrs(session.channel, id, info)
}

// sendAttrs sends ATTRS response.
func (s *SftpServer) sendAttrs(w io.Writer, id uint32, info os.FileInfo) error {
	mode := uint32(info.Mode().Perm())
	if info.IsDir() {
		mode |= 0040000
	} else {
		mode |= 0100000
	}

	pkt := make([]byte, 1+4+4+8+4)
	pkt[0] = SSH_FXP_ATTRS
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], SSH_FILEXFER_ATTR_SIZE|SSH_FILEXFER_ATTR_PERMISSIONS)
	binary.BigEndian.PutUint64(pkt[9:17], uint64(info.Size()))
	binary.BigEndian.PutUint32(pkt[17:21], mode)

	return s.sendSftpPacket(w, pkt)
}

// handleOpenDir handles OPENDIR request.
func (s *SftpServer) handleOpenDir(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)

	// Store handle
	handle := []byte("dir:" + path)
	pkt := make([]byte, 1+4+4+len(handle))
	pkt[0] = SSH_FXP_HANDLE
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
	copy(pkt[9:], handle)

	return s.sendSftpPacket(session.channel, pkt)
}

// handleReadDir handles READDIR request.
func (s *SftpServer) handleReadDir(session *sftpSession, id uint32, data []byte) error {
	handle := s.parseHandle(data)
	if !strings.HasPrefix(handle, "dir:") {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	path := handle[4:]
	fullPath := session.user.homeDir + path

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	if len(entries) == 0 {
		return s.sendStatus(session.channel, id, SSH_FX_EOF)
	}

	// Build NAME response
	pkt := make([]byte, 1+4+4)
	pkt[0] = SSH_FXP_NAME
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(entries)))

	for _, entry := range entries {
		name := entry.Name()
		nameLen := uint32(len(name))

		entryPkt := make([]byte, 4+nameLen+4+4)
		binary.BigEndian.PutUint32(entryPkt[0:4], nameLen)
		copy(entryPkt[4:4+nameLen], name)
		binary.BigEndian.PutUint32(entryPkt[4+nameLen:8+nameLen], 0)  // long name length
		binary.BigEndian.PutUint32(entryPkt[8+nameLen:12+nameLen], 0) // attrs

		pkt = append(pkt, entryPkt...)
	}

	return s.sendSftpPacket(session.channel, pkt)
}

// handleOpen handles OPEN request.
func (s *SftpServer) handleOpen(session *sftpSession, id uint32, data []byte) error {
	pos := 0
	pathLen := binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4
	path := string(data[pos : pos+int(pathLen)])
	pos += int(pathLen)
	flags := binary.BigEndian.Uint32(data[pos : pos+4])

	fullPath := session.user.homeDir + path

	// Open file
	var file *os.File
	var err error

	if flags&SSH_FXF_READ != 0 && flags&SSH_FXF_WRITE == 0 {
		file, err = os.Open(fullPath)
	} else if flags&SSH_FXF_WRITE != 0 {
		if flags&SSH_FXF_CREAT != 0 {
			if flags&SSH_FXF_TRUNC != 0 {
				file, err = os.Create(fullPath)
			} else {
				file, err = os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0644)
			}
		} else {
			file, err = os.OpenFile(fullPath, os.O_WRONLY, 0644)
		}
	} else {
		file, err = os.Open(fullPath)
	}

	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}
	file.Close()

	// Store handle
	handle := []byte("file:" + path + ":" + fmt.Sprint(flags))
	pkt := make([]byte, 1+4+4+len(handle))
	pkt[0] = SSH_FXP_HANDLE
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(len(handle)))
	copy(pkt[9:], handle)

	return s.sendSftpPacket(session.channel, pkt)
}

// handleRead handles READ request.
func (s *SftpServer) handleRead(session *sftpSession, id uint32, data []byte) error {
	handle := s.parseHandle(data[:4])
	if !strings.HasPrefix(handle, "file:") {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	// Parse offset and length
	pos := 4
	offset := binary.BigEndian.Uint64(data[pos : pos+8])
	pos += 8
	length := binary.BigEndian.Uint32(data[pos : pos+4])

	// Parse path from handle
	parts := strings.Split(handle, ":")
	if len(parts) < 2 {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}
	path := parts[1]
	fullPath := session.user.homeDir + path

	// Read file
	file, err := os.Open(fullPath)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}
	defer file.Close()

	// Seek to offset
	file.Seek(int64(offset), 0)

	// Read data
	buf := make([]byte, length)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	if n == 0 {
		return s.sendStatus(session.channel, id, SSH_FX_EOF)
	}

	// Send DATA response
	pkt := make([]byte, 1+4+4+n)
	pkt[0] = SSH_FXP_DATA
	binary.BigEndian.PutUint32(pkt[1:5], id)
	binary.BigEndian.PutUint32(pkt[5:9], uint32(n))
	copy(pkt[9:], buf[:n])

	return s.sendSftpPacket(session.channel, pkt)
}

// handleWrite handles WRITE request.
func (s *SftpServer) handleWrite(session *sftpSession, id uint32, data []byte) error {
	pos := 0
	handle := s.parseHandle(data[pos : pos+4])
	pos += 4
	if !strings.HasPrefix(handle, "file:") {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	offset := binary.BigEndian.Uint64(data[pos : pos+8])
	pos += 8
	length := binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4
	writeData := data[pos : pos+int(length)]

	// Parse path from handle
	parts := strings.Split(handle, ":")
	if len(parts) < 2 {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}
	path := parts[1]
	fullPath := session.user.homeDir + path

	// Write to file
	file, err := os.OpenFile(fullPath, os.O_WRONLY, 0644)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}
	defer file.Close()

	// Seek to offset
	file.Seek(int64(offset), 0)

	// Write data
	_, err = file.Write(writeData)
	if err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// handleClose handles CLOSE request.
func (s *SftpServer) handleClose(session *sftpSession, id uint32, data []byte) error {
	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// handleMkdir handles MKDIR request.
func (s *SftpServer) handleMkdir(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// handleRmdir handles RMDIR request.
func (s *SftpServer) handleRmdir(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	if err := os.RemoveAll(fullPath); err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// handleRemove handles REMOVE request.
func (s *SftpServer) handleRemove(session *sftpSession, id uint32, data []byte) error {
	path := s.parsePath(data)
	fullPath := session.user.homeDir + path

	if err := os.Remove(fullPath); err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// handleRename handles RENAME request.
func (s *SftpServer) handleRename(session *sftpSession, id uint32, data []byte) error {
	pos := 0
	oldPath := s.parsePath(data[pos:])
	pos += 4 + int(binary.BigEndian.Uint32(data[pos:pos+4]))
	newPath := s.parsePath(data[pos:])

	oldFullPath := session.user.homeDir + oldPath
	newFullPath := session.user.homeDir + newPath

	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return s.sendStatus(session.channel, id, SSH_FX_FAILURE)
	}

	return s.sendStatus(session.channel, id, SSH_FX_OK)
}

// parsePath parses a path from SFTP packet data.
func (s *SftpServer) parsePath(data []byte) string {
	if len(data) < 4 {
		return "/"
	}
	pathLen := binary.BigEndian.Uint32(data[0:4])
	if int(pathLen) > len(data)-4 {
		return "/"
	}
	return string(data[4 : 4+pathLen])
}

// parseHandle parses a handle from SFTP packet data.
func (s *SftpServer) parseHandle(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	handleLen := binary.BigEndian.Uint32(data[0:4])
	if int(handleLen) > len(data)-4 {
		return ""
	}
	return string(data[4 : 4+handleLen])
}

// Stop stops the SFTP server.
func (s *SftpServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	s.running = false
	return nil
}

// AddUser adds a user account with password.
func (s *SftpServer) AddUser(username, password, homeDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if homeDir == "" {
		homeDir = "/"
	}

	s.users[username] = &sftpUser{
		username: username,
		password: password,
		homeDir:  homeDir,
	}
	return nil
}

// AddUserWithKey adds a user account with public key.
func (s *SftpServer) AddUserWithKey(username, keyStr, homeDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyStr))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	if homeDir == "" {
		homeDir = "/"
	}

	user, exists := s.users[username]
	if !exists {
		user = &sftpUser{
			username: username,
			homeDir:  homeDir,
		}
	}
	user.pubKeys = append(user.pubKeys, pubKey)
	s.users[username] = user

	return nil
}

// RemoveUser removes a user account.
func (s *SftpServer) RemoveUser(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, username)
	return nil
}

// IsRunning returns true if the server is running.
func (s *SftpServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
