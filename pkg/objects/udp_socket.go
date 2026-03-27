// pkg/objects/udp_socket.go
// UdpSocket object type for UDP socket operations.
package objects

import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"
)

// UdpSocket represents a UDP socket for connectionless datagram communication.
// It supports two modes:
// 1. Send-only mode: Created with NewUdpSocket(), can send but cannot receive
// 2. Bidirectional mode: After Bind(), can both send and receive
// UdpSocket is thread-safe.
type UdpSocket struct {
	conn     *net.UDPConn
	timeout  time.Duration
	closed   bool
	mu       sync.Mutex
	bound    bool
	localAddr *net.UDPAddr
}

// NewUdpSocket creates a new unbound UDP socket.
// The socket can send datagrams immediately using SendTo/SendToBytes,
// but must call Bind() before it can receive datagrams.
func NewUdpSocket() *UdpSocket {
	return &UdpSocket{
		timeout: 0,
		closed:  false,
		bound:   false,
	}
}

// Type returns the object type.
func (s *UdpSocket) Type() ObjectType { return UdpSocketType }

// TypeTag returns the type tag for fast type checking.
func (s *UdpSocket) TypeTag() TypeTag { return TagUdpSocket }

// Inspect returns a string representation of the UdpSocket.
func (s *UdpSocket) Inspect() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "<UDP_SOCKET closed>"
	}
	if !s.bound {
		return "<UDP_SOCKET unbound>"
	}
	if s.localAddr != nil {
		return fmt.Sprintf("<UDP_SOCKET %s>", s.localAddr.String())
	}
	return "<UDP_SOCKET bound>"
}

// ToBool returns true if not closed.
func (s *UdpSocket) ToBool() *Bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &Bool{Value: !s.closed}
}

// HashKey returns a hash key for the UdpSocket.
func (s *UdpSocket) HashKey() HashKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	return HashKey{Type: UdpSocketType, Value: uint64(uintptr(unsafe.Pointer(s)))}
}

// Bind binds the socket to a local address.
// The address should be in the format "host:port" (e.g., ":8080" for all interfaces, port 8080).
// Use ":0" to bind to a random available port.
// Returns NULL on success, or an Error on failure.
func (s *UdpSocket) Bind(addr string) Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return newError("udp socket is closed")
	}

	if s.bound {
		return newError("udp socket already bound")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return newError("resolve address failed: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return newError("bind failed: %v", err)
	}

	s.conn = conn
	s.bound = true
	s.localAddr = conn.LocalAddr().(*net.UDPAddr)
	return NULL
}

// SetTimeout sets the read timeout in milliseconds.
// A timeout of 0 means no timeout (blocking read).
// Returns NULL.
func (s *UdpSocket) SetTimeout(ms int) Object {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = time.Duration(ms) * time.Millisecond
	return NULL
}

// SendTo sends a string to the specified address.
// The address should be in the format "host:port".
// Returns the number of bytes sent as an Int, or an Error on failure.
// This method can be used without calling Bind() first.
func (s *UdpSocket) SendTo(data string, addr string) Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return newError("udp socket is closed")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return newError("resolve address failed: %v", err)
	}

	var conn *net.UDPConn
	var n int

	if s.bound && s.conn != nil {
		// Use the bound connection
		n, err = s.conn.WriteToUDP([]byte(data), udpAddr)
	} else {
		// Create a temporary connection for send-only mode
		conn, err = net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return newError("dial failed: %v", err)
		}
		defer conn.Close()
		n, err = conn.Write([]byte(data))
	}

	if err != nil {
		return newError("send failed: %v", err)
	}

	return NewInt(int64(n))
}

// SendToBytes sends a byte slice to the specified address.
// The address should be in the format "host:port".
// Returns the number of bytes sent as an Int, or an Error on failure.
// This method can be used without calling Bind() first.
func (s *UdpSocket) SendToBytes(data []byte, addr string) Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return newError("udp socket is closed")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return newError("resolve address failed: %v", err)
	}

	var conn *net.UDPConn
	var n int

	if s.bound && s.conn != nil {
		// Use the bound connection
		n, err = s.conn.WriteToUDP(data, udpAddr)
	} else {
		// Create a temporary connection for send-only mode
		conn, err = net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return newError("dial failed: %v", err)
		}
		defer conn.Close()
		n, err = conn.Write(data)
	}

	if err != nil {
		return newError("send failed: %v", err)
	}

	return NewInt(int64(n))
}

// ReceiveFrom receives a datagram and returns the data as a String.
// Also returns the sender's address as a SocketAddr.
// Returns NULL for data on timeout, or an Error on failure.
// Requires Bind() to be called first.
func (s *UdpSocket) ReceiveFrom(bufSize int) (Object, Object) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return newError("udp socket is closed"), NULL
	}

	if !s.bound || s.conn == nil {
		return newError("udp socket not bound, cannot receive"), NULL
	}

	if s.timeout > 0 {
		s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	} else {
		s.conn.SetReadDeadline(time.Time{})
	}

	buf := make([]byte, bufSize)
	n, addr, err := s.conn.ReadFromUDP(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL, NULL
		}
		return newError("receive failed: %v", err), NULL
	}

	data := NewString(string(buf[:n]))
	senderAddr := NewSocketAddr(addr.IP.String(), addr.Port)
	return data, senderAddr
}

// ReceiveFromBytes receives a datagram and returns the data as an Array of Ints.
// Each element in the array is a byte value (0-255).
// Also returns the sender's address as a SocketAddr.
// Returns NULL for data on timeout, or an Error on failure.
// Requires Bind() to be called first.
func (s *UdpSocket) ReceiveFromBytes(bufSize int) (Object, Object) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return newError("udp socket is closed"), NULL
	}

	if !s.bound || s.conn == nil {
		return newError("udp socket not bound, cannot receive"), NULL
	}

	if s.timeout > 0 {
		s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	} else {
		s.conn.SetReadDeadline(time.Time{})
	}

	buf := make([]byte, bufSize)
	n, addr, err := s.conn.ReadFromUDP(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL, NULL
		}
		return newError("receive failed: %v", err), NULL
	}

	elements := make([]Object, n)
	for i := 0; i < n; i++ {
		elements[i] = NewInt(int64(buf[i]))
	}
	data := NewArray(elements)
	senderAddr := NewSocketAddr(addr.IP.String(), addr.Port)
	return data, senderAddr
}

// LocalAddr returns the local address as a SocketAddr.
// Returns an Error if not bound.
func (s *UdpSocket) LocalAddr() Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.bound || s.localAddr == nil {
		return newError("udp socket not bound")
	}

	return NewSocketAddr(s.localAddr.IP.String(), s.localAddr.Port)
}

// Close closes the UDP socket.
// Returns NULL on success, or an Error on failure.
func (s *UdpSocket) Close() Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return NULL
	}

	s.closed = true

	if s.conn == nil {
		return NULL
	}

	err := s.conn.Close()
	s.conn = nil
	if err != nil {
		return newError("close failed: %v", err)
	}

	return NULL
}

// IsClosed returns true if the socket is closed.
func (s *UdpSocket) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}