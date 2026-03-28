// pkg/stdlib/ssh.go
// SSH module for Xxlang - SSH client functionality.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "ssh",
		Exports: map[string]objects.Object{
			// ============================================================
			// Creation Functions
			// ============================================================

			// newClient creates a new SSH client object (unconnected).
			// Use connect() for one-step connection.
			"newClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("newClient takes no arguments")
				}
				return objects.NewSSHClient()
			}),

			// connect establishes SSH connection with password.
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

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithKey establishes SSH connection with private key file.
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

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithKeyStr establishes SSH connection with private key string.
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

				client := objects.NewSSHClient()
				if err := client.ConnectWithKeyStr(host.Value, int(port.Value), user.Value, keyStr.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				return client
			}),

			// connectWithConfig establishes SSH connection with a config map.
			"connectWithConfig": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("connectWithConfig takes exactly 1 argument (config)")
				}
				configMap, ok := args[0].(*objects.Map)
				if !ok {
					return Error("argument must be a map (config)")
				}

				config := &objects.SSHConfig{
					Timeout: 30, // default timeout
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
					config.Port = 22 // default SSH port
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

				// Extract optional knownHostsPath
				if knownHostsPath, ok := getString("knownHostsPath"); ok {
					config.KnownHostsPath = knownHostsPath
				}

				// Extract optional ignoreHostKey
				if ignoreHostKey, ok := getBool("ignoreHostKey"); ok {
					config.IgnoreHostKey = ignoreHostKey
				}

				client := objects.NewSSHClient()
				if err := client.ConnectWithConfig(config); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				return client
			}),

			// isSSHClient checks if an object is an SSHClient.
			"isSSHClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSSHClient takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.SSHClient)
				return Bool(ok)
			}),

			// ============================================================
			// Quick Functions - One-time operations
			// ============================================================

			// exec executes a single command on a remote server.
			"exec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 5 {
					return Error("exec takes exactly 5 arguments (host, port, user, password, cmd)")
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
				cmd, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (cmd)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				output, err := client.Exec(cmd.Value)
				if err != nil {
					return Error("exec failed: " + err.Error())
				}
				return objects.NewString(output)
			}),

			// upload uploads a file to a remote server.
			"upload": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("upload takes exactly 6 arguments (host, port, user, password, localPath, remotePath)")
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
				localPath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (localPath)")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				if err := client.Upload(localPath.Value, remotePath.Value); err != nil {
					return Error("upload failed: " + err.Error())
				}
				return objects.NULL
			}),

			// download downloads a file from a remote server.
			"download": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("download takes exactly 6 arguments (host, port, user, password, remotePath, localPath)")
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
				remotePath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (remotePath)")
				}
				localPath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (localPath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				if err := client.Download(remotePath.Value, localPath.Value); err != nil {
					return Error("download failed: " + err.Error())
				}
				return objects.NULL
			}),

			// testConnection tests if an SSH connection can be established.
			"testConnection": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("testConnection takes exactly 4 arguments (host, port, user, password)")
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

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return objects.FALSE
				}
				client.Close()
				return objects.TRUE
			}),

			// testConnectionWithKey tests if an SSH connection can be established with a key.
			"testConnectionWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("testConnectionWithKey takes exactly 4 arguments (host, port, user, keyPath)")
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

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return objects.FALSE
				}
				client.Close()
				return objects.TRUE
			}),

			// ============================================================
			// Byte Operations - Upload/Download bytes directly
			// ============================================================

			// uploadBytes uploads in-memory bytes to a remote path over SSH.
			// Parameters: host, port, user, password, bytes, remotePath
			// Returns: null on success, Error on failure
			"uploadBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("uploadBytes takes exactly 6 arguments (host, port, user, password, bytes, remotePath)")
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
				dataBytes, ok := args[4].(*objects.Bytes)
				if !ok {
					return Error("fifth argument must be bytes")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				// Write bytes to remote file using base64 encoding for binary safety
				if err := client.WriteFile(remotePath.Value, string(dataBytes.Value)); err != nil {
					return Error("uploadBytes failed: " + err.Error())
				}
				return objects.NULL
			}),

			// downloadBytes downloads a remote file and returns its content as bytes.
			// Parameters: host, port, user, password, remotePath
			// Returns: Bytes object on success, Error on failure
			"downloadBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 5 {
					return Error("downloadBytes takes exactly 5 arguments (host, port, user, password, remotePath)")
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
				remotePath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content, err := client.ReadFile(remotePath.Value)
				if err != nil {
					return Error("downloadBytes failed: " + err.Error())
				}
				return &objects.Bytes{Value: []byte(content)}
			}),

			// uploadBytesWithKey uploads in-memory bytes to a remote path using key authentication.
			// Parameters: host, port, user, keyPath, bytes, remotePath
			// Returns: null on success, Error on failure
			"uploadBytesWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("uploadBytesWithKey takes exactly 6 arguments (host, port, user, keyPath, bytes, remotePath)")
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
				dataBytes, ok := args[4].(*objects.Bytes)
				if !ok {
					return Error("fifth argument must be bytes")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				if err := client.WriteFile(remotePath.Value, string(dataBytes.Value)); err != nil {
					return Error("uploadBytesWithKey failed: " + err.Error())
				}
				return objects.NULL
			}),

			// downloadBytesWithKey downloads a remote file using key authentication.
			// Parameters: host, port, user, keyPath, remotePath
			// Returns: Bytes object on success, Error on failure
			"downloadBytesWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 5 {
					return Error("downloadBytesWithKey takes exactly 5 arguments (host, port, user, keyPath, remotePath)")
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
				remotePath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content, err := client.ReadFile(remotePath.Value)
				if err != nil {
					return Error("downloadBytesWithKey failed: " + err.Error())
				}
				return &objects.Bytes{Value: []byte(content)}
			}),
		},
	})
}
