// pkg/objects/socket_addr.go
// SocketAddr object type for network address representation.
package objects

import (
	"fmt"
)

// SocketAddr represents a network address with host and port.
// It is used for socket operations to specify endpoints.
type SocketAddr struct {
	host string
	port int
}

// NewSocketAddr creates a new SocketAddr with the given host and port.
func NewSocketAddr(host string, port int) *SocketAddr {
	return &SocketAddr{
		host: host,
		port: port,
	}
}

// Type returns the object type.
func (a *SocketAddr) Type() ObjectType { return SocketAddrType }

// TypeTag returns the type tag for fast type checking.
func (a *SocketAddr) TypeTag() TypeTag { return TagSocketAddr }

// Inspect returns a string representation of the SocketAddr.
func (a *SocketAddr) Inspect() string {
	return fmt.Sprintf("<SOCKET_ADDR %s:%d>", a.host, a.port)
}

// ToBool returns true (SocketAddr is always truthy).
func (a *SocketAddr) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the SocketAddr.
func (a *SocketAddr) HashKey() HashKey {
	return HashKey{
		Type:  SocketAddrType,
		Value: uint64(len(a.host))<<32 | uint64(a.port),
	}
}

// Host returns the host part of the address.
func (a *SocketAddr) Host() string {
	return a.host
}

// Port returns the port part of the address.
func (a *SocketAddr) Port() int {
	return a.port
}

// ToStr returns the address in "host:port" format.
// For IPv6 addresses, the host is wrapped in brackets.
func (a *SocketAddr) ToStr() string {
	// Check if the host contains a colon (IPv6 address)
	if len(a.host) > 0 && a.host[0] == ':' {
		// IPv6 address, wrap in brackets
		return fmt.Sprintf("[%s]:%d", a.host, a.port)
	}
	for i := 0; i < len(a.host); i++ {
		if a.host[i] == ':' {
			// IPv6 address, wrap in brackets
			return fmt.Sprintf("[%s]:%d", a.host, a.port)
		}
	}
	return fmt.Sprintf("%s:%d", a.host, a.port)
}