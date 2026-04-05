# SSH Module

The SSH module provides comprehensive SSH/SFTP functionality for remote server management, file transfer, and command execution.

## Overview

Xxlang's SSH implementation is built on `golang.org/x/crypto/ssh` and provides:

- **Password and key-based authentication** - Support for both password and private key authentication
- **Native SFTP protocol** - High-performance file transfer using the SFTP protocol (not shell-based)
- **Port forwarding** - SSH tunneling for secure access to internal services
- **Stream execution** - Real-time command output streaming
- **Directory recursion** - Upload/download entire directories
- **Cross-platform** - Works consistently on Windows, Linux, and macOS

## Quick Start

```xxl
import "ssh"

// Connect to server
var client = ssh.connect("server.com", 22, "username", "password")

// Execute command
var output = client.exec("uptime")
println(output)

// Upload file
client.upload("local.txt", "/remote/path/file.txt")

// Download file
client.download("/remote/config.json", "./config.json")

// Close connection
client.close()
```

## API Reference

### Module Functions

#### `ssh.connect(host, port, user, password)`

Create and connect to an SSH server with password authentication.

**Parameters:**
- `host` (string) - Server hostname or IP address
- `port` (int) - SSH port (usually 22)
- `user` (string) - Username
- `password` (string) - Password

**Returns:** `SSHClient` object

**Example:**
```xxl
var client = ssh.connect("192.168.1.100", 22, "root", "password123")
```

---

#### `ssh.connectWithKey(host, port, user, keyPath)`

Connect using a private key file.

**Parameters:**
- `host` (string) - Server hostname
- `port` (int) - SSH port
- `user` (string) - Username
- `keyPath` (string) - Path to private key file

**Returns:** `SSHClient` object

**Example:**
```xxl
var client = ssh.connectWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa")
```

---

#### `ssh.connectWithKeyStr(host, port, user, keyString)`

Connect using a private key string.

**Parameters:**
- `host` (string) - Server hostname
- `port` (int) - SSH port
- `user` (string) - Username
- `keyString` (string) - Private key content

**Example:**
```xxl
var keyStr = io.readFile("/home/user/.ssh/id_rsa")
var client = ssh.connectWithKeyStr("server.com", 22, "deploy", keyStr)
```

---

#### `ssh.connectWithConfig(config)`

Connect with full configuration options.

**Parameters:**
- `config` (map) - Configuration map with keys:
  - `host` (string) - Server hostname
  - `port` (int) - SSH port
  - `user` (string) - Username
  - `password` (string, optional) - Password
  - `keyPath` (string, optional) - Private key path
  - `keyString` (string, optional) - Private key content
  - `keyPassphrase` (string, optional) - Key passphrase
  - `timeout` (int, optional) - Connection timeout in seconds
  - `ignoreHostKey` (bool, optional) - Skip host key verification

**Returns:** `SSHClient` object

**Example:**
```xxl
var config = {
    "host": "server.com",
    "port": 22,
    "user": "deploy",
    "keyPath": "~/.ssh/id_rsa",
    "timeout": 60,
    "ignoreHostKey": true
}
var client = ssh.connectWithConfig(config)
```

---

#### `ssh.newClient()`

Create a new SSH client without connecting.

**Returns:** `SSHClient` object (call `connect()` to establish connection)

**Example:**
```xxl
var client = ssh.newClient()
client.connect("server.com", 22, "user", "pass")
```

---

#### `ssh.exec(host, port, user, password, command)`

Quick execute a single command without managing connection.

**Parameters:**
- `host`, `port`, `user`, `password` - Connection parameters
- `command` (string) - Command to execute

**Returns:** `string` - Command output

**Example:**
```xxl
var output = ssh.exec("192.168.1.100", 22, "root", "pass", "uptime")
```

---

#### `ssh.upload(host, port, user, password, localPath, remotePath)`

Quick upload a file without managing connection.

**Example:**
```xxl
ssh.upload("server.com", 22, "user", "pass", "file.txt", "/remote/file.txt")
```

---

#### `ssh.download(host, port, user, password, remotePath, localPath)`

Quick download a file without managing connection.

**Example:**
```xxl
ssh.download("server.com", 22, "user", "pass", "/remote/file.txt", "file.txt")
```

---

#### `ssh.uploadBytes(host, port, user, password, bytes, remotePath)`

Upload byte data to remote file.

**Example:**
```xxl
var data = io.readFileBytes("image.png")
ssh.uploadBytes("server.com", 22, "user", "pass", data, "/remote/image.png")
```

---

#### `ssh.downloadBytes(host, port, user, password, remotePath)`

Download remote file as bytes.

**Returns:** `bytes` - File content

**Example:**
```xxl
var data = ssh.downloadBytes("server.com", 22, "user", "pass", "/remote/image.png")
io.writeFileBytes("local.png", data)
```

---

#### `ssh.uploadWithKey(host, port, user, keyPath, localPath, remotePath)`

Upload file using key authentication.

**Example:**
```xxl
ssh.uploadWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa", "file.txt", "/remote/file.txt")
```

---

#### `ssh.downloadWithKey(host, port, user, keyPath, remotePath, localPath)`

Download file using key authentication.

**Example:**
```xxl
ssh.downloadWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa", "/remote/file.txt", "file.txt")
```

---

#### `ssh.testConnection(host, port, user, password)`

Test if server is reachable with given credentials.

**Returns:** `bool` - true if connection successful

**Example:**
```xxl
if (ssh.testConnection("server.com", 22, "user", "pass")) {
    println("Server is reachable")
} else {
    println("Cannot connect")
}
```

---

#### `ssh.testConnectionWithKey(host, port, user, keyPath)`

Test connection using key authentication.

**Example:**
```xxl
if (ssh.testConnectionWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa")) {
    println("Key authentication successful")
}
```

---

### SSHClient Object Methods

#### `client.connect(host, port, user, password)`

Establish SSH connection.

**Throws:** Error if connection fails

---

#### `client.connectWithKey(keyPath)`

Connect using private key file.

---

#### `client.connectWithKeyStr(keyString)`

Connect using private key string.

---

#### `client.connectWithConfig(config)`

Connect with configuration map.

---

#### `client.is_connected()`

Check if connected.

**Returns:** `bool`

---

#### `client.getHost()`

Get connected host.

**Returns:** `string`

---

#### `client.getUser()`

Get connected username.

**Returns:** `string`

---

#### `client.exec(command)`

Execute command and return output.

**Parameters:**
- `command` (string) - Command to execute

**Returns:** `string` - Combined stdout and stderr

**Example:**
```xxl
var hostname = client.exec("hostname")
var disk = client.exec("df -h /")
```

---

#### `client.execFull(command)`

Execute command with separate stdout/stderr.

**Returns:** `map` with keys: `stdout`, `stderr`, `exitCode`

**Example:**
```xxl
var result = client.execFull("some-command")
println("stdout:", result["stdout"])
println("stderr:", result["stderr"])
println("exit:", result["exitCode"])
```

---

#### `client.execStream(command, callback)`

Stream command output line by line.

**Parameters:**
- `command` (string) - Command to execute
- `callback` (function) - Called for each line

**Example:**
```xxl
client.execStream("tail -f /var/log/syslog", func(line) {
    println("[SYSLOG]", line)
})
```

---

#### `client.upload(localPath, remotePath)`

Upload file to server.

**Example:**
```xxl
client.upload("./script.xxl", "/tmp/script.xxl")
```

---

#### `client.download(remotePath, localPath)`

Download file from server.

**Example:**
```xxl
client.download("/var/log/app.log", "./logs/app.log")
```

---

#### `client.uploadBytes(bytes, remotePath)`

Upload byte data.

**Example:**
```xxl
var data = io.readFileBytes("image.png")
client.uploadBytes(data, "/remote/image.png")
```

---

#### `client.downloadBytes(remotePath)`

Download as bytes.

**Returns:** `bytes`

---

#### `client.readFile(remotePath)`

Read remote file content.

**Returns:** `string`

---

#### `client.readfile_bytes(remotePath)`

Read remote file as bytes.

**Returns:** `bytes`

---

#### `client.writeFile(remotePath, content)`

Write content to remote file.

**Example:**
```xxl
client.writeFile("/tmp/config.json", "{\"port\": 8080}")
```

---

#### `client.writefile_bytes(remotePath, bytes)`

Write bytes to remote file.

---

#### `client.mkdir(path)`

Create directory.

---

#### `client.mkdirAll(path)`

Create directories recursively.

---

#### `client.listDir(path)`

List directory contents.

**Returns:** `array` of maps with keys: `name`, `size`, `isDir`, `mode`, `modTime`

**Example:**
```xxl
var files = client.listDir("/tmp")
for (var i = 0; i < len(files); i = i + 1) {
    var f = files[i]
    println(f["name"], f["size"], f["isDir"])
}
```

---

#### `client.walkDir(path)`

Recursively walk directory tree.

**Returns:** `array` of file info maps

---

#### `client.uploadDir(localPath, remotePath)`

Upload entire directory recursively.

---

#### `client.downloadDir(remotePath, localPath)`

Download entire directory recursively.

---

#### `client.stat(path)`

Get file metadata.

**Returns:** `map` with keys: `size`, `isFile`, `isDir`, `mode`, `modTime`

---

#### `client.exists(path)`

Check if path exists.

**Returns:** `bool`

---

#### `client.rename(oldPath, newPath)`

Rename/move file.

---

#### `client.remove(path)`

Delete file.

---

#### `client.removeDir(path)`

Delete directory recursively.

---

#### `client.chmod(path, mode)`

Change file permissions.

**Example:**
```xxl
client.chmod("/tmp/script.sh", 755)
```

---

#### `client.localForward(localPort, remoteHost, remotePort)`

Create SSH tunnel (local port forwarding).

**Returns:** `Listener` object with `close()` method

**Example:**
```xxl
// Access local:8080 -> internal-service:80 through SSH tunnel
var listener = client.localForward(8080, "internal-service", 80)
// Keep tunnel open
sleep(3600)
listener.close()
```

---

#### `client.runScriptStr(scriptContent)`

Execute local script content on remote server.

**Example:**
```xxl
var script = io.readFile("./deploy.xxl")
var result = client.runScriptStr(script)
```

---

#### `client.close()`

Close SSH connection.

---

## Examples

See [examples/ssh_examples.xxl](../examples/ssh_examples.xxl) for 15 comprehensive examples covering:

1. Quick commands (one-liners)
2. File transfer (replacing sshrun)
3. SSH client object usage
4. SSH key authentication
5. Advanced file operations
6. Directory upload/download
7. Stream command output
8. Port forwarding
9. Remote script execution
10. Batch operations on multiple servers
11. Complete deployment workflow
12. Server monitoring
13. sshrun pattern replacements
14. HEX path support notes
15. Error handling and retry logic

## Comparison with sshrun

| Feature | sshrun | Xxlang |
|---------|--------|--------|
| Execute commands | ✅ | ✅ |
| Upload/download | ✅ | ✅ (native SFTP) |
| Password auth | ✅ | ✅ |
| Key auth | ✅ | ✅ |
| Directory transfer | ❌ | ✅ |
| Port forwarding | ❌ | ✅ |
| Stream output | ❌ | ✅ |
| Programmable | ❌ | ✅ |
| Cross-platform paths | HEX encoding needed | ✅ Native support |

## Notes

### Windows Git Bash Path Handling

Unlike sshrun which requires HEX path encoding to workaround MSYS2 path conversion issues, Xxlang's pure Go implementation handles paths correctly on all platforms:

```xxl
// This works directly on Windows Git Bash:
ssh.upload("server.com", 22, "user", "pass", "file.txt", "/root/file.txt")

// No need for HEX encoding!
```

### Security Considerations

1. **Host Key Verification**: By default, host keys are not verified. For production use, consider implementing host key checking.

2. **Credential Storage**: Avoid hardcoding credentials. Use environment variables or secure credential storage:
   ```xxl
   var password = os.getenv("SSH_PASSWORD")
   var client = ssh.connect(host, port, user, password)
   ```

3. **Key Passphrases**: When using encrypted private keys, provide the passphrase:
   ```xxl
   var config = {
       "host": "server.com",
       "user": "deploy",
       "keyPath": "~/.ssh/id_rsa",
       "keyPassphrase": "your-passphrase"
   }
   ```
