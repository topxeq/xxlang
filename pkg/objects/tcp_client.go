// pkg/objects/tcp_client.go
// TcpClient object type for TCP client connections.
package objects

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"
)

// TcpClient represents a TCP client connection.
// It provides methods for connecting to TCP servers, sending and receiving data.
// TcpClient is thread-safe and supports read timeouts.
type TcpClient struct {
	conn      net.Conn
	timeout   time.Duration
	closed    bool
	mu        sync.Mutex
	bufReader *bufio.Reader
}

// NewTcpClient creates a new unconnected TcpClient.
// The client must call Connect before sending or receiving data.
func NewTcpClient() *TcpClient {
	return &TcpClient{
		timeout: 0,
		closed:  false,
	}
}

// NewTcpClientFromConn creates a TcpClient from an existing net.Conn.
// This is useful for server-side accepted connections.
func NewTcpClientFromConn(conn net.Conn) *TcpClient {
	return &TcpClient{
		conn:   conn,
		closed: false,
	}
}

// Type returns the object type.
func (c *TcpClient) Type() ObjectType { return TcpClientType }

// TypeTag returns the type tag for fast type checking.
func (c *TcpClient) TypeTag() TypeTag { return TagTcpClient }

// Inspect returns a string representation of the TcpClient.
func (c *TcpClient) Inspect() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return "<TCP_CLIENT closed>"
	}
	if c.conn == nil {
		return "<TCP_CLIENT unconnected>"
	}
	return fmt.Sprintf("<TCP_CLIENT %s>", c.conn.RemoteAddr().String())
}

// ToBool returns true if connected and not closed.
func (c *TcpClient) ToBool() *Bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &Bool{Value: c.conn != nil && !c.closed}
}

// HashKey returns a hash key for the TcpClient.
func (c *TcpClient) HashKey() HashKey {
	c.mu.Lock()
	defer c.mu.Unlock()
	return HashKey{Type: TcpClientType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}

// Connect establishes a TCP connection to the specified address.
// The address should be in the format "host:port".
// Returns NULL on success, or an Error on failure.
func (c *TcpClient) Connect(addr string) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil && !c.closed {
		return newError("tcp client already connected")
	}
	var d net.Dialer
	if c.timeout > 0 {
		d.Timeout = c.timeout
	}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return newError("connect failed: %v", err)
	}
	c.conn = conn
	c.closed = false
	c.bufReader = bufio.NewReader(conn)
	return NULL
}

// SetTimeout sets the read timeout in milliseconds.
// A timeout of 0 means no timeout (blocking read).
// Returns NULL.
func (c *TcpClient) SetTimeout(ms int) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = time.Duration(ms) * time.Millisecond
	return NULL
}

// SendString sends a string over the connection.
// Returns the number of bytes sent as an Int, or an Error on failure.
func (c *TcpClient) SendString(data string) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.closed {
		return newError("tcp client not connected")
	}
	n, err := c.conn.Write([]byte(data))
	if err != nil {
		return newError("send failed: %v", err)
	}
	return NewInt(int64(n))
}

// SendBytes sends a byte slice over the connection.
// Returns the number of bytes sent as an Int, or an Error on failure.
func (c *TcpClient) SendBytes(data []byte) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.closed {
		return newError("tcp client not connected")
	}
	n, err := c.conn.Write(data)
	if err != nil {
		return newError("send failed: %v", err)
	}
	return NewInt(int64(n))
}

// Receive reads up to n bytes and returns the data as a String.
// Returns NULL on timeout or EOF, or an Error on failure.
func (c *TcpClient) Receive(n int) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.closed {
		return newError("tcp client not connected")
	}
	if c.timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	} else {
		c.conn.SetReadDeadline(time.Time{})
	}
	buf := make([]byte, n)
	numRead, err := c.conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL
		}
		if err.Error() == "EOF" {
			return NULL
		}
		return newError("receive failed: %v", err)
	}
	return NewString(string(buf[:numRead]))
}

// ReceiveBytes reads up to n bytes and returns the data as an Array of Ints.
// Each element in the array is a byte value (0-255).
// Returns NULL on timeout or EOF, or an Error on failure.
func (c *TcpClient) ReceiveBytes(n int) Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.closed {
		return newError("tcp client not connected")
	}
	if c.timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	} else {
		c.conn.SetReadDeadline(time.Time{})
	}
	buf := make([]byte, n)
	numRead, err := c.conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL
		}
		if err.Error() == "EOF" {
			return NULL
		}
		return newError("receive failed: %v", err)
	}
	elements := make([]Object, numRead)
	for i := 0; i < numRead; i++ {
		elements[i] = NewInt(int64(buf[i]))
	}
	return NewArray(elements)
}

// ReceiveLine reads until newline and returns the line as a String.
// The trailing newline and carriage return are removed.
// Returns NULL on timeout or EOF, or an Error on failure.
func (c *TcpClient) ReceiveLine() Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.closed {
		return newError("tcp client not connected")
	}
	if c.bufReader == nil {
		c.bufReader = bufio.NewReader(c.conn)
	}
	if c.timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	} else {
		c.conn.SetReadDeadline(time.Time{})
	}
	line, err := c.bufReader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return NULL
		}
		if err.Error() == "EOF" {
			if len(line) == 0 {
				return NULL
			}
			return NewString(line)
		}
		return newError("receiveLine failed: %v", err)
	}
	// Remove trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return NewString(line)
}

// LocalAddr returns the local address as a SocketAddr.
// Returns an Error if not connected.
func (c *TcpClient) LocalAddr() Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return newError("tcp client not connected")
	}
	addr := c.conn.LocalAddr().(*net.TCPAddr)
	return NewSocketAddr(addr.IP.String(), addr.Port)
}

// RemoteAddr returns the remote address as a SocketAddr.
// Returns an Error if not connected.
func (c *TcpClient) RemoteAddr() Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return newError("tcp client not connected")
	}
	addr := c.conn.RemoteAddr().(*net.TCPAddr)
	return NewSocketAddr(addr.IP.String(), addr.Port)
}

// Close closes the connection.
// Returns NULL on success, or an Error on failure.
func (c *TcpClient) Close() Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return NULL
	}
	c.closed = true
	if c.conn == nil {
		return NULL
	}
	err := c.conn.Close()
	c.conn = nil
	if err != nil {
		return newError("close failed: %v", err)
	}
	return NULL
}

// IsClosed returns true if the connection is closed.
func (c *TcpClient) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}