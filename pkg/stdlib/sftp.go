// pkg/stdlib/sftp.go
// SFTP module for Xxlang - SFTP client and server functionality.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "sftp",
		Exports: map[string]objects.Object{
			// ============================================================
			// Client Creation Functions
			// ============================================================

			// connect establishes SFTP connection with password.
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

				client := objects.NewSftpClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SFTP connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithKey establishes SFTP connection with private key file.
			"connectWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("connectWithKey takes exactly 4 arguments (host, port, user, keyPath)")
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
				keyPath, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (keyPath)")
				}

				client := objects.NewSftpClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SFTP connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithKeyStr establishes SFTP connection with private key string.
			"connectWithKeyStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("connectWithKeyStr takes exactly 4 arguments (host, port, user, keyStr)")
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
				keyStr, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (keyStr)")
				}

				client := objects.NewSftpClient()
				if err := client.ConnectWithKeyStr(host.Value, int(port.Value), user.Value, keyStr.Value); err != nil {
					return Error("SFTP connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithConfig establishes SFTP connection with a config map.
			"connectWithConfig": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("connectWithConfig takes exactly 1 argument (config)")
				}
				configMap, ok := args[0].(*objects.Map)
				if !ok {
					return Error("argument must be a map (config)")
				}

				config := &objects.SftpConfig{
					Timeout: 30,
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

				// Extract port (optional, default 22)
				if port, ok := getInt("port"); ok {
					config.Port = port
				}
				if config.Port == 0 {
					config.Port = 22
				}

				// Extract user (required)
				if user, ok := getString("user"); ok {
					config.User = user
				}
				if config.User == "" {
					return Error("config must contain 'user' field")
				}

				// Extract optional password
				if password, ok := getString("password"); ok {
					config.Password = password
				}

				// Extract optional keyPath
				if keyPath, ok := getString("keyPath"); ok {
					config.KeyPath = keyPath
				}

				// Extract optional keyStr
				if keyStr, ok := getString("keyStr"); ok {
					config.KeyStr = keyStr
				}

				// Extract optional keyPassphrase
				if keyPassphrase, ok := getString("keyPassphrase"); ok {
					config.KeyPassphrase = keyPassphrase
				}

				// Extract optional timeout
				if timeout, ok := getInt("timeout"); ok {
					config.Timeout = timeout
				}

				// Extract optional ignoreHostKey
				if ignoreHostKey, ok := getBool("ignoreHostKey"); ok {
					config.IgnoreHostKey = ignoreHostKey
				}

				client := objects.NewSftpClient()
				if err := client.ConnectWithConfig(config); err != nil {
					return Error("SFTP connection failed: " + err.Error())
				}
				return client
			}),

			// ============================================================
			// Server Creation Functions
			// ============================================================

			// createServer creates an SFTP server.
			"createServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("createServer takes at least 1 argument (addr)")
				}

				server := objects.NewSftpServer()

				var addr string
				var config *objects.SftpServerConfig

				if addrObj, ok := args[0].(*objects.String); ok {
					addr = addrObj.Value
				} else {
					return Error("first argument must be a string (addr)")
				}

				if len(args) > 1 {
					configMap, ok := args[1].(*objects.Map)
					if !ok {
						return Error("second argument must be a map (config)")
					}
					config = &objects.SftpServerConfig{}

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
					if hostKey, ok := getString("hostKey"); ok {
						config.HostKey = hostKey
					}
					if hostKeyPath, ok := getString("hostKeyPath"); ok {
						config.HostKeyPath = hostKeyPath
					}
				}

				if err := server.Create(addr, config); err != nil {
					return Error("failed to create SFTP server: " + err.Error())
				}

				return server
			}),

			// ============================================================
			// Type Check Functions
			// ============================================================

			// isSftpClient checks if an object is an SftpClient.
			"isSftpClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSftpClient takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.SftpClient)
				return Bool(ok)
			}),

			// isSftpServer checks if an object is an SftpServer.
			"isSftpServer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSftpServer takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.SftpServer)
				return Bool(ok)
			}),
		},
	})
}
