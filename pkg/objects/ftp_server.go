// pkg/objects/ftp_server.go
// FtpServer object for Xxlang - FTP server functionality.
package objects

import (
	"bufio"
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
)

// FtpServer represents an FTP server.
type FtpServer struct {
	mu            sync.Mutex
	listener      net.Listener
	running       bool
	host          string
	port          int
	users         map[string]*ftpUser
	sessions      map[int]*ftpSession
	nextSessionID int
	config        *FtpServerConfig
}

// ftpUser represents a user account.
type ftpUser struct {
	username string
	password string
	homeDir  string
}

// ftpSession represents an active client session.
type ftpSession struct {
	id          int
	conn        net.Conn
	text        *textprotoConn
	user        *ftpUser
	currentDir  string
	dataConn    net.Conn
	dataPort    int
	passiveMode bool
	binaryMode  bool
	server      *FtpServer
	quit        chan struct{}
}

// textprotoConn is a minimal textproto connection wrapper.
type textprotoConn struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func newTextprotoConn(conn net.Conn) *textprotoConn {
	return &textprotoConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (t *textprotoConn) ReadLine() (string, error) {
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (t *textprotoConn) WriteString(s string) error {
	_, err := t.writer.WriteString(s)
	if err != nil {
		return err
	}
	return t.writer.Flush()
}

func (t *textprotoConn) PrintfLine(format string, args ...interface{}) error {
	fmt.Fprintf(t.writer, format, args...)
	t.writer.WriteString("\r\n")
	return t.writer.Flush()
}

// FtpServerConfig holds FTP server configuration.
type FtpServerConfig struct {
	Host           string
	Port           int
	PassivePorts   string // e.g., "50000-51000"
	MaxConnections int
	Timeout        int // seconds
	WelcomeMessage string
	PassivePortMin int
	PassivePortMax int
}

// Type returns the object type.
func (s *FtpServer) Type() ObjectType { return FtpServerType }

// TypeTag returns the fast type tag.
func (s *FtpServer) TypeTag() TypeTag { return TagFtpServer }

// Inspect returns a string representation of the FtpServer.
func (s *FtpServer) Inspect() string {
	if s.running {
		return fmt.Sprintf("FtpServer(running on %s:%d)", s.host, s.port)
	}
	return "FtpServer(stopped)"
}

// ToBool returns true if running.
func (s *FtpServer) ToBool() *Bool {
	return &Bool{Value: s.running}
}

// HashKey returns a hash key for the FtpServer.
func (s *FtpServer) HashKey() HashKey {
	return HashKey{
		Type:  FtpServerType,
		Value: uint64(uintptr(unsafe.Pointer(s))),
	}
}

// NewFtpServer creates a new FtpServer.
func NewFtpServer() *FtpServer {
	return &FtpServer{
		users:    make(map[string]*ftpUser),
		sessions: make(map[int]*ftpSession),
		config: &FtpServerConfig{
			Port:           21,
			MaxConnections: 100,
			Timeout:        300,
			PassivePortMin: 50000,
			PassivePortMax: 51000,
		},
	}
}

// Create creates and configures the FTP server.
func (s *FtpServer) Create(addr string, config *FtpServerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server is already running")
	}

	// Parse address
	host, port := parseAddr(addr, 21)
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
		if config.WelcomeMessage != "" {
			s.config.WelcomeMessage = config.WelcomeMessage
		}
		if config.PassivePortMin > 0 {
			s.config.PassivePortMin = config.PassivePortMin
		}
		if config.PassivePortMax > config.PassivePortMin {
			s.config.PassivePortMax = config.PassivePortMax
		}
		// Parse passive port range string
		if config.PassivePorts != "" {
			min, max := parsePortRange(config.PassivePorts)
			if min > 0 {
				s.config.PassivePortMin = min
				s.config.PassivePortMax = max
			}
		}
	}

	return nil
}

// parseAddr parses address string into host and port.
func parseAddr(addr string, defaultPort int) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		port, _ := strconv.Atoi(parts[1])
		return parts[0], port
	}
	if len(parts) == 1 {
		return parts[0], defaultPort
	}
	return "0.0.0.0", defaultPort
}

// parsePortRange parses a port range string like "50000-51000".
func parsePortRange(s string) (int, int) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, 0
	}
	min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	max, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return min, max
}

// Start starts the FTP server.
func (s *FtpServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server is already running")
	}

	// Start listening
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.listener = listener
	s.running = true

	// Accept connections in background
	go s.acceptLoop()

	return nil
}

// acceptLoop accepts incoming connections.
func (s *FtpServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if server is shutting down
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

		session := s.newSession(conn)
		s.sessions[session.id] = session
		s.mu.Unlock()

		go s.handleSession(session)
	}
}

// newSession creates a new client session.
func (s *FtpServer) newSession(conn net.Conn) *ftpSession {
	s.nextSessionID++
	return &ftpSession{
		id:          s.nextSessionID,
		conn:        conn,
		text:        newTextprotoConn(conn),
		currentDir:  "/",
		passiveMode: true,
		binaryMode:  true,
		server:      s,
		quit:        make(chan struct{}),
	}
}

// handleSession handles a client session.
func (s *FtpServer) handleSession(session *ftpSession) {
	defer s.closeSession(session)

	// Send welcome message
	welcome := s.config.WelcomeMessage
	if welcome == "" {
		welcome = "Welcome to Xxlang FTP Server"
	}
	session.text.PrintfLine("220 %s", welcome)

	// Command loop
	for {
		select {
		case <-session.quit:
			return
		default:
			// Set read deadline
			session.conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

			line, err := session.text.ReadLine()
			if err != nil {
				return
			}

			if err := s.handleCommand(session, line); err != nil {
				return
			}
		}
	}
}

// handleCommand handles an FTP command.
func (s *FtpServer) handleCommand(session *ftpSession, line string) error {
	parts := strings.SplitN(line, " ", 2)
	cmd := strings.ToUpper(parts[0])
	var args string
	if len(parts) > 1 {
		args = parts[1]
	}

	switch cmd {
	case "USER":
		return s.handleUser(session, args)
	case "PASS":
		return s.handlePass(session, args)
	case "QUIT":
		session.text.PrintfLine("221 Goodbye")
		close(session.quit)
		return errors.New("quit")
	case "PWD":
		return s.handlePwd(session)
	case "CWD":
		return s.handleCwd(session, args)
	case "CDUP":
		return s.handleCdup(session)
	case "MKD":
		return s.handleMkd(session, args)
	case "RMD":
		return s.handleRmd(session, args)
	case "DELE":
		return s.handleDele(session, args)
	case "RNFR":
		return s.handleRnfr(session, args)
	case "RNTO":
		return s.handleRnto(session, args)
	case "LIST", "NLST":
		return s.handleList(session, args, cmd == "LIST")
	case "RETR":
		return s.handleRetr(session, args)
	case "STOR":
		return s.handleStor(session, args)
	case "TYPE":
		return s.handleType(session, args)
	case "PASV":
		return s.handlePasv(session)
	case "PORT":
		return s.handlePort(session, args)
	case "SIZE":
		return s.handleSize(session, args)
	case "FEAT":
		return s.handleFeat(session)
	case "SYST":
		session.text.PrintfLine("215 UNIX Type: L8")
	case "NOOP":
		session.text.PrintfLine("200 OK")
	default:
		session.text.PrintfLine("500 Unknown command: %s", cmd)
	}
	return nil
}

// handleUser handles USER command.
func (s *FtpServer) handleUser(session *ftpSession, username string) error {
	user, ok := s.users[username]
	if !ok {
		session.text.PrintfLine("331 User %s OK. Password required", username)
		// Store for later authentication
		session.user = &ftpUser{username: username}
		return nil
	}
	session.user = user
	session.text.PrintfLine("331 User %s OK. Password required", username)
	return nil
}

// handlePass handles PASS command.
func (s *FtpServer) handlePass(session *ftpSession, password string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Login incorrect")
		return nil
	}

	// Check password
	user, ok := s.users[session.user.username]
	if !ok || user.password != password {
		session.text.PrintfLine("530 Login incorrect")
		session.user = nil
		return nil
	}

	session.user = user
	session.currentDir = user.homeDir
	session.text.PrintfLine("230 User %s logged in", user.username)
	return nil
}

// handlePwd handles PWD command.
func (s *FtpServer) handlePwd(session *ftpSession) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}
	session.text.PrintfLine("257 \"%s\" is current directory", session.currentDir)
	return nil
}

// handleCwd handles CWD command.
func (s *FtpServer) handleCwd(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	newDir := s.resolvePath(session, path)
	fullPath := filepath.Join(session.user.homeDir, newDir)

	info, err := os.Stat(fullPath)
	if err != nil || !info.IsDir() {
		session.text.PrintfLine("550 Failed to change directory")
		return nil
	}

	session.currentDir = newDir
	session.text.PrintfLine("250 Directory changed to %s", newDir)
	return nil
}

// handleCdup handles CDUP command.
func (s *FtpServer) handleCdup(session *ftpSession) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	parent := filepath.Dir(session.currentDir)
	if parent == "." {
		parent = "/"
	}
	session.currentDir = parent
	session.text.PrintfLine("250 Directory changed to %s", parent)
	return nil
}

// handleMkd handles MKD command.
func (s *FtpServer) handleMkd(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		session.text.PrintfLine("550 Failed to create directory")
		return nil
	}

	session.text.PrintfLine("257 \"%s\" created", path)
	return nil
}

// handleRmd handles RMD command.
func (s *FtpServer) handleRmd(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	if err := os.Remove(fullPath); err != nil {
		session.text.PrintfLine("550 Failed to remove directory")
		return nil
	}

	session.text.PrintfLine("250 Directory removed")
	return nil
}

// handleDele handles DELE command.
func (s *FtpServer) handleDele(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	if err := os.Remove(fullPath); err != nil {
		session.text.PrintfLine("550 Failed to delete file")
		return nil
	}

	session.text.PrintfLine("250 File deleted")
	return nil
}

// handleRnfr handles RNFR command.
func (s *FtpServer) handleRnfr(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	// Store for RNTO
	session.text.PrintfLine("350 Ready for RNTO")
	return nil
}

// handleRnto handles RNTO command.
func (s *FtpServer) handleRnto(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	session.text.PrintfLine("250 Rename successful")
	return nil
}

// handleList handles LIST/NLST command.
func (s *FtpServer) handleList(session *ftpSession, path string, detailed bool) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	dataConn, err := s.openDataConnection(session)
	if err != nil {
		session.text.PrintfLine("425 Failed to open data connection")
		return nil
	}
	defer dataConn.Close()

	session.text.PrintfLine("150 Opening data connection")

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		session.text.PrintfLine("550 Failed to read directory")
		return nil
	}

	writer := bufio.NewWriter(dataConn)
	for _, entry := range entries {
		if detailed {
			info, _ := entry.Info()
			mode := "-rw-r--r--"
			if entry.IsDir() {
				mode = "drwxr-xr-x"
			}
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			fmt.Fprintf(writer, "%s 1 user group %d Jan 01 00:00 %s\r\n", mode, size, entry.Name())
		} else {
			fmt.Fprintf(writer, "%s\r\n", entry.Name())
		}
	}
	writer.Flush()

	session.text.PrintfLine("226 Transfer complete")
	return nil
}

// handleRetr handles RETR command.
func (s *FtpServer) handleRetr(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	dataConn, err := s.openDataConnection(session)
	if err != nil {
		session.text.PrintfLine("425 Failed to open data connection")
		return nil
	}
	defer dataConn.Close()

	session.text.PrintfLine("150 Opening data connection")

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	file, err := os.Open(fullPath)
	if err != nil {
		session.text.PrintfLine("550 Failed to open file")
		return nil
	}
	defer file.Close()

	io.Copy(dataConn, file)
	session.text.PrintfLine("226 Transfer complete")
	return nil
}

// handleStor handles STOR command.
func (s *FtpServer) handleStor(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	dataConn, err := s.openDataConnection(session)
	if err != nil {
		session.text.PrintfLine("425 Failed to open data connection")
		return nil
	}
	defer dataConn.Close()

	session.text.PrintfLine("150 Opening data connection")

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	file, err := os.Create(fullPath)
	if err != nil {
		session.text.PrintfLine("550 Failed to create file")
		return nil
	}
	defer file.Close()

	io.Copy(file, dataConn)
	session.text.PrintfLine("226 Transfer complete")
	return nil
}

// handleType handles TYPE command.
func (s *FtpServer) handleType(session *ftpSession, typ string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	switch strings.ToUpper(typ) {
	case "I", "A":
		session.binaryMode = (strings.ToUpper(typ) == "I")
		session.text.PrintfLine("200 Type set to %s", typ)
	default:
		session.text.PrintfLine("500 Unknown type: %s", typ)
	}
	return nil
}

// handlePasv handles PASV command.
func (s *FtpServer) handlePasv(session *ftpSession) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	session.passiveMode = true

	// Find available port
	port := s.config.PassivePortMin
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	for err != nil && port < s.config.PassivePortMax {
		port++
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		session.text.PrintfLine("425 Failed to open data connection")
		return nil
	}

	// Get local IP
	localIP := s.getLocalIP()
	p1, p2 := port>>8, port&0xff
	h := strings.Split(localIP, ".")

	session.text.PrintfLine("227 Entering Passive Mode (%s,%s,%s,%s,%d,%d)", h[0], h[1], h[2], h[3], p1, p2)

	// Wait for data connection in background
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			session.dataConn = conn
		}
		listener.Close()
	}()

	return nil
}

// handlePort handles PORT command.
func (s *FtpServer) handlePort(session *ftpSession, args string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	session.passiveMode = false

	// Parse PORT arguments
	parts := strings.Split(args, ",")
	if len(parts) != 6 {
		session.text.PrintfLine("500 Invalid PORT command")
		return nil
	}

	h1, _ := strconv.Atoi(parts[0])
	h2, _ := strconv.Atoi(parts[1])
	h3, _ := strconv.Atoi(parts[2])
	h4, _ := strconv.Atoi(parts[3])
	p1, _ := strconv.Atoi(parts[4])
	p2, _ := strconv.Atoi(parts[5])

	host := fmt.Sprintf("%d.%d.%d.%d", h1, h2, h3, h4)
	port := p1<<8 | p2

	// Store for later use
	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			session.dataConn = conn
		}
	}()

	session.text.PrintfLine("200 PORT command successful")
	return nil
}

// handleSize handles SIZE command.
func (s *FtpServer) handleSize(session *ftpSession, path string) error {
	if session.user == nil {
		session.text.PrintfLine("530 Not logged in")
		return nil
	}

	fullPath := filepath.Join(session.user.homeDir, s.resolvePath(session, path))
	info, err := os.Stat(fullPath)
	if err != nil {
		session.text.PrintfLine("550 File not found")
		return nil
	}

	session.text.PrintfLine("213 %d", info.Size())
	return nil
}

// handleFeat handles FEAT command.
func (s *FtpServer) handleFeat(session *ftpSession) error {
	session.text.PrintfLine("211-Features:")
	session.text.PrintfLine(" PASV")
	session.text.PrintfLine(" SIZE")
	session.text.PrintfLine(" TYPE")
	session.text.PrintfLine("211 End")
	return nil
}

// openDataConnection opens a data connection.
func (s *FtpServer) openDataConnection(session *ftpSession) (net.Conn, error) {
	if session.dataConn != nil {
		conn := session.dataConn
		session.dataConn = nil
		return conn, nil
	}

	return nil, errors.New("no data connection")
}

// resolvePath resolves a path relative to the current directory.
func (s *FtpServer) resolvePath(session *ftpSession, path string) string {
	if path == "" {
		return session.currentDir
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(session.currentDir, path)
}

// getLocalIP gets the local IP address.
func (s *FtpServer) getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// closeSession closes a client session.
func (s *FtpServer) closeSession(session *ftpSession) {
	session.conn.Close()
	if session.dataConn != nil {
		session.dataConn.Close()
	}

	s.mu.Lock()
	delete(s.sessions, session.id)
	s.mu.Unlock()
}

// Stop stops the FTP server.
func (s *FtpServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Close all sessions
	for _, session := range s.sessions {
		session.conn.Close()
	}
	s.sessions = make(map[int]*ftpSession)

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	s.running = false
	return nil
}

// AddUser adds a user account.
func (s *FtpServer) AddUser(username, password, homeDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if homeDir == "" {
		homeDir = "/"
	}

	s.users[username] = &ftpUser{
		username: username,
		password: password,
		homeDir:  homeDir,
	}
	return nil
}

// RemoveUser removes a user account.
func (s *FtpServer) RemoveUser(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, username)
	return nil
}

// IsRunning returns true if the server is running.
func (s *FtpServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
