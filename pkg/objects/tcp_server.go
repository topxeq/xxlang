// pkg/objects/tcp_server.go
// TcpServer object type for TCP server operations.
package objects

import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"
)

// TcpServer represents a TCP server that can accept incoming connections.
// It provides methods for listening on a port, accepting connections, and
// handling connections via callbacks. TcpServer is thread-safe.
type TcpServer struct {
	mu          sync.Mutex
	listener    *net.TCPListener
	addr        *SocketAddr
	closed      bool
	timeout     time.Duration
	reuseAddr   bool
	onAccept    func(client *TcpClient)
	stopChan    chan struct{}
	running     bool
}

// NewTcpServer creates a new unbound TcpServer.
// The server must call Listen before accepting connections.
func NewTcpServer() *TcpServer {
	return &TcpServer{
		closed:    true,
		reuseAddr: true,
		stopChan:  make(chan struct{}),
	}
}

// Type returns the object type.
func (s *TcpServer) Type() ObjectType { return TcpServerType }

// TypeTag returns the type tag for fast type checking.
func (s *TcpServer) TypeTag() TypeTag { return TagTcpServer }

// Inspect returns a string representation of the TcpServer.
func (s *TcpServer) Inspect() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.listener == nil {
		return "<TCP_SERVER closed>"
	}
	return fmt.Sprintf("<TCP_SERVER %s>", s.addr.ToStr())
}

// ToBool returns true if the server is listening and not closed.
func (s *TcpServer) ToBool() *Bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &Bool{Value: s.listener != nil && !s.closed}
}

// HashKey returns a hash key for the TcpServer.
func (s *TcpServer) HashKey() HashKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	return HashKey{Type: TcpServerType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}

// Listen starts the TCP server on the specified address.
// The address should be in the format "host:port" or ":port".
// If port is 0, a random port will be assigned.
// Returns NULL on success, or an Error on failure.
func (s *TcpServer) Listen(addr string) Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already listening
	if s.listener != nil && !s.closed {
		return newError("tcp server already listening")
	}

	// Resolve TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return newError("resolve address failed: %v", err)
	}

	// Create listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return newError("listen failed: %v", err)
	}

	// Get the actual address (in case port was 0)
	actualAddr := listener.Addr().(*net.TCPAddr)
	s.listener = listener
	s.addr = NewSocketAddr(actualAddr.IP.String(), actualAddr.Port)
	s.closed = false
	s.stopChan = make(chan struct{})

	return NULL
}

// SetTimeout sets the accept timeout in milliseconds.
// This affects the Accept method when blocking.
// A timeout of 0 means no timeout (blocking accept).
// Returns NULL.
func (s *TcpServer) SetTimeout(ms int) Object {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = time.Duration(ms) * time.Millisecond
	return NULL
}

// SetReuseAddr sets whether to enable SO_REUSEADDR option.
// Must be called before Listen. Default is true.
// Returns NULL.
func (s *TcpServer) SetReuseAddr(reuse bool) Object {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reuseAddr = reuse
	return NULL
}

// OnAccept sets a callback function to be called for each accepted connection.
// When a callback is set, the server will automatically accept connections
// in the background when Start() is called.
// Returns NULL.
func (s *TcpServer) OnAccept(callback func(client *TcpClient)) Object {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onAccept = callback
	return NULL
}

// Accept waits for and accepts the next incoming connection.
// If timeout > 0, it will return NULL after the timeout expires.
// Returns a TcpClient on success, NULL on timeout, or an Error on failure.
// If the server is not listening, returns an Error.
func (s *TcpServer) Accept(timeoutMs int) Object {
	s.mu.Lock()
	if s.listener == nil || s.closed {
		s.mu.Unlock()
		return newError("tcp server not listening")
	}
	listener := s.listener
	s.mu.Unlock()

	// Set deadline if timeout specified
	if timeoutMs > 0 {
		deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
		if err := listener.SetDeadline(deadline); err != nil {
			return newError("set deadline failed: %v", err)
		}
	} else if s.timeout > 0 {
		deadline := time.Now().Add(s.timeout)
		if err := listener.SetDeadline(deadline); err != nil {
			return newError("set deadline failed: %v", err)
		}
	} else {
		// Clear deadline
		listener.SetDeadline(time.Time{})
	}

	conn, err := listener.Accept()
	if err != nil {
		// Check for timeout
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL
		}
		// Check if listener was closed
		s.mu.Lock()
		closed := s.closed
		s.mu.Unlock()
		if closed {
			return newError("tcp server closed")
		}
		return newError("accept failed: %v", err)
	}

	return NewTcpClientFromConn(conn)
}

// Start begins the accept loop in a background goroutine.
// Connections will be passed to the onAccept callback if set.
// Use Close() to stop the server.
// Returns NULL on success, or an Error if not listening or already running.
func (s *TcpServer) Start() Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil || s.closed {
		return newError("tcp server not listening")
	}

	if s.running {
		return newError("tcp server already running")
	}

	s.running = true
	stopChan := s.stopChan
	listener := s.listener

	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				// Set a short deadline for polling
				listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
				conn, err := listener.Accept()
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue // Poll again
					}
					// Real error or closed
					s.mu.Lock()
					running := s.running
					s.mu.Unlock()
					if !running {
						return
					}
					continue
				}

				// Handle connection
				client := NewTcpClientFromConn(conn)
				s.mu.Lock()
				callback := s.onAccept
				s.mu.Unlock()

				if callback != nil {
					go callback(client)
				}
			}
		}
	}()

	return NULL
}

// Addr returns the address the server is listening on.
// Returns nil if the server is not listening.
func (s *TcpServer) Addr() *SocketAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Close stops the server and releases resources.
// Returns NULL on success, or an Error on failure.
func (s *TcpServer) Close() Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return NULL
	}

	s.closed = true
	s.running = false

	// Signal stop to accept loop
	close(s.stopChan)

	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		if err != nil {
			return newError("close failed: %v", err)
		}
	}

	return NULL
}

// IsClosed returns true if the server is not listening.
func (s *TcpServer) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}