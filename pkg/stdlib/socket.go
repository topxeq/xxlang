// Package stdlib provides standard library modules for the Xxlang language.
// This file implements the socket module for TCP/UDP network operations.
package stdlib

import (
	"net"
	"strconv"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "socket",
		Exports: map[string]objects.Object{
			// createTcpServer creates and starts a TCP server listening on the given address.
			// Parameter: addr (string) - address to listen on in "host:port" format
			// Returns: TcpServer object or Error
			"createTcpServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("createTcpServer() takes exactly 1 argument")
				}
				addr, ok := args[0].(*objects.String)
				if !ok {
					return Error("createTcpServer() requires a string address")
				}

				server := objects.NewTcpServer()
				result := server.Listen(addr.Value)
				if err, ok := result.(*objects.Error); ok {
					return err
				}
				return server
			}),

			// createTcpClient creates an unconnected TCP client.
			// Returns: TcpClient object
			"createTcpClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createTcpClient() takes no arguments")
				}
				return objects.NewTcpClient()
			}),

			// connectTcp creates a TCP client and connects to the specified address.
			// Parameter: addr (string) - address to connect to in "host:port" format
			// Returns: TcpClient object or Error
			"connectTcp": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("connectTcp() takes exactly 1 argument")
				}
				addr, ok := args[0].(*objects.String)
				if !ok {
					return Error("connectTcp() requires a string address")
				}

				client := objects.NewTcpClient()
				result := client.Connect(addr.Value)
				if err, ok := result.(*objects.Error); ok {
					return err
				}
				return client
			}),

			// createUdpSocket creates a UDP socket.
			// Returns: UdpSocket object
			"createUdpSocket": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("createUdpSocket() takes no arguments")
				}
				return objects.NewUdpSocket()
			}),

			// parseAddr parses a "host:port" string into a SocketAddr object.
			// Parameter: addrStr (string) - address in "host:port" format
			// Returns: SocketAddr object or Error
			"parseAddr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseAddr() takes exactly 1 argument")
				}
				addrStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseAddr() requires a string argument")
				}

				host, port, err := net.SplitHostPort(addrStr.Value)
				if err != nil {
					return Error("parseAddr() failed: " + err.Error())
				}

				portNum, _ := strconv.Atoi(port)
				return objects.NewSocketAddr(host, portNum)
			}),

			// resolveHost resolves a hostname to its first IP address.
			// Parameter: hostname (string) - hostname to resolve
			// Returns: String with IP address or Error
			"resolveHost": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("resolveHost() takes exactly 1 argument")
				}
				hostname, ok := args[0].(*objects.String)
				if !ok {
					return Error("resolveHost() requires a string argument")
				}

				addrs, err := net.LookupHost(hostname.Value)
				if err != nil {
					return Error("resolveHost() failed: " + err.Error())
				}

				if len(addrs) == 0 {
					return Error("resolveHost() found no addresses")
				}

				return String(addrs[0])
			}),

			// lookupHost resolves a hostname to all its IP addresses.
			// Parameter: hostname (string) - hostname to resolve
			// Returns: Array of String IP addresses or Error
			"lookupHost": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("lookupHost() takes exactly 1 argument")
				}
				hostname, ok := args[0].(*objects.String)
				if !ok {
					return Error("lookupHost() requires a string argument")
				}

				addrs, err := net.LookupHost(hostname.Value)
				if err != nil {
					return Error("lookupHost() failed: " + err.Error())
				}

				elements := make([]objects.Object, len(addrs))
				for i, addr := range addrs {
					elements[i] = String(addr)
				}

				return Array(elements...)
			}),
		},
	})
}