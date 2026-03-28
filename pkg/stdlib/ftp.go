// pkg/stdlib/ftp.go
// FTP module for Xxlang - FTP client and server functionality.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "ftp",
		Exports: map[string]objects.Object{
			// ============================================================
			// Client Creation Functions
			// ============================================================

			// newClient creates a new FTP client object (unconnected).
			// Use connect() for one-step connection.
			"newClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("newClient takes no arguments")
				}
				return objects.NewFtpClient()
			}),

			// connect establishes FTP connection with credentials.
			"connect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("connect takes exactly 4 arguments (host, port, user, password)")
				}
				host, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (host)")
				}
				port, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer (port)")
				}
				user, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (user)")
				}
				password, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (password)")
				}

				client := objects.NewFtpClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("FTP connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithConfig establishes FTP connection with a config map.
			"connectWithConfig": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("connectWithConfig takes exactly 1 argument (config)")
				}
				configMap, ok := args[0].(*objects.Map)
				if !ok {
					return Error("argument must be a map (config)")
				}

				config := &objects.FtpConfig{
					Timeout: 30,
					Passive: true,
					Binary:  true,
				}

				// Helper function to get string value from map
				getString := func(key string) (string, bool) {
					keyObj := objects.NewString(key)
					if pair, exists := configMap.Pairs[keyObj.HashKey()]; exists {
						if s, ok := pair.Value.(*objects.String); ok {
							return s.Value, true
						}
					}
					return "", false
				}

				// Helper function to get int value from map
				getInt := func(key string) (int, bool) {
					keyObj := objects.NewString(key)
					if pair, exists := configMap.Pairs[keyObj.HashKey()]; exists {
						if i, ok := pair.Value.(*objects.Int); ok {
							return int(i.Value), true
						}
					}
					return 0, false
				}

				// Helper function to get bool value from map
				getBool := func(key string) (bool, bool) {
					keyObj := objects.NewString(key)
					if pair, exists := configMap.Pairs[keyObj.HashKey()]; exists {
						if b, ok := pair.Value.(*objects.Bool); ok {
							return b.Value, true
						}
					}
					return false, false
				}

				// Extract host (required)
				if host, ok := getString("host"); ok {
					config.Host = host
				}
				if config.Host == "" {
					return Error("config must contain 'host' field")
				}

				// Extract port (optional, default 21)
				if port, ok := getInt("port"); ok {
					config.Port = port
				}
				if config.Port == 0 {
					config.Port = 21
				}

				// Extract user (required)
				if user, ok := getString("user"); ok {
					config.User = user
				}
				if config.User == "" {
					return Error("config must contain 'user' field")
				}

				// Extract password (required)
				if password, ok := getString("password"); ok {
					config.Password = password
				}

				// Extract optional timeout
				if timeout, ok := getInt("timeout"); ok {
					config.Timeout = timeout
				}

				// Extract optional passive mode
				if passive, ok := getBool("passive"); ok {
					config.Passive = passive
				}

				// Extract optional binary mode
				if binary, ok := getBool("binary"); ok {
					config.Binary = binary
				}

				client := objects.NewFtpClient()
				if err := client.ConnectWithConfig(config); err != nil {
					return Error("FTP connection failed: " + err.Error())
				}
				return client
			}),

			// ============================================================
			// Server Creation Functions
			// ============================================================

			// createServer creates an FTP server.
			"createServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("createServer takes at least 1 argument (addr)")
				}

				server := objects.NewFtpServer()

				var addr string
				var config *objects.FtpServerConfig

				if addrObj, ok := args[0].(*objects.String); ok {
					addr = addrObj.Value
				} else {
					return Error("first argument must be a string (addr)")
				}

				if len(args) > 1 {
					if configMap, ok := args[1].(*objects.Map); ok {
						config = &objects.FtpServerConfig{}

						// Helper function to get string value from map
						getString := func(key string) (string, bool) {
							keyObj := objects.NewString(key)
							if pair, exists := configMap.Pairs[keyObj.HashKey()]; exists {
								if s, ok := pair.Value.(*objects.String); ok {
									return s.Value, true
								}
							}
							return "", false
						}

						// Helper function to get int value from map
						getInt := func(key string) (int, bool) {
							keyObj := objects.NewString(key)
							if pair, exists := configMap.Pairs[keyObj.HashKey()]; exists {
								if i, ok := pair.Value.(*objects.Int); ok {
									return int(i.Value), true
								}
							}
							return 0, false
						}

						if host, ok := getString("host"); ok {
							config.Host = host
						}
						if port, ok := getInt("port"); ok {
							config.Port = port
						}
						if maxConn, ok := getInt("maxConnections"); ok {
							config.MaxConnections = maxConn
						}
						if timeout, ok := getInt("timeout"); ok {
							config.Timeout = timeout
						}
						if welcome, ok := getString("welcomeMessage"); ok {
							config.WelcomeMessage = welcome
						}
						if passivePorts, ok := getString("passivePorts"); ok {
							config.PassivePorts = passivePorts
						}
					}
				}

				if err := server.Create(addr, config); err != nil {
					return Error("failed to create FTP server: " + err.Error())
				}

				return server
			}),

			// ============================================================
			// Type Check Functions
			// ============================================================

			// isFtpClient checks if an object is an FtpClient.
			"isFtpClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isFtpClient takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.FtpClient)
				return Bool(ok)
			}),

			// isFtpServer checks if an object is an FtpServer.
			"isFtpServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isFtpServer takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.FtpServer)
				return Bool(ok)
			}),
		},
	})
}
