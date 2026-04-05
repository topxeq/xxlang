# SSH 模块

SSH 模块提供全面的 SSH/SFTP 功能，用于远程服务器管理、文件传输和命令执行。

## 概述

Xxlang 的 SSH 实现基于 `golang.org/x/crypto/ssh`，提供以下功能：

- **密码和密钥认证** - 支持密码和私钥两种认证方式
- **原生 SFTP 协议** - 使用 SFTP 协议进行高性能文件传输（非 shell 方式）
- **端口转发** - SSH 隧道，安全访问内部服务
- **流式执行** - 实时命令输出流
- **目录递归** - 上传/下载整个目录
- **跨平台** - 在 Windows、Linux 和 macOS 上一致工作

## 快速开始

```xxl
import "ssh"

// 连接到服务器
var client = ssh.connect("server.com", 22, "username", "password")

// 执行命令
var output = client.exec("uptime")
println(output)

// 上传文件
client.upload("local.txt", "/remote/path/file.txt")

// 下载文件
client.download("/remote/config.json", "./config.json")

// 关闭连接
client.close()
```

## API 参考

### 模块函数

#### `ssh.connect(host, port, user, password)`

使用密码认证创建并连接到 SSH 服务器。

**参数：**
- `host` (string) - 服务器主机名或 IP 地址
- `port` (int) - SSH 端口（通常为 22）
- `user` (string) - 用户名
- `password` (string) - 密码

**返回值：** `SSHClient` 对象

**示例：**
```xxl
var client = ssh.connect("192.168.1.100", 22, "root", "password123")
```

---

#### `ssh.connectWithKey(host, port, user, keyPath)`

使用私钥文件连接。

**参数：**
- `host` (string) - 服务器主机名
- `port` (int) - SSH 端口
- `user` (string) - 用户名
- `keyPath` (string) - 私钥文件路径

**返回值：** `SSHClient` 对象

**示例：**
```xxl
var client = ssh.connectWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa")
```

---

#### `ssh.connectWithKeyStr(host, port, user, keyString)`

使用私钥字符串连接。

**参数：**
- `host` (string) - 服务器主机名
- `port` (int) - SSH 端口
- `user` (string) - 用户名
- `keyString` (string) - 私钥内容

**示例：**
```xxl
var keyStr = io.readFile("/home/user/.ssh/id_rsa")
var client = ssh.connectWithKeyStr("server.com", 22, "deploy", keyStr)
```

---

#### `ssh.connectWithConfig(config)`

使用完整配置选项连接。

**参数：**
- `config` (map) - 配置映射，包含以下键：
  - `host` (string) - 服务器主机名
  - `port` (int) - SSH 端口
  - `user` (string) - 用户名
  - `password` (string, 可选) - 密码
  - `keyPath` (string, 可选) - 私钥路径
  - `keyString` (string, 可选) - 私钥内容
  - `keyPassphrase` (string, 可选) - 私钥密码短语
  - `timeout` (int, 可选) - 连接超时（秒）
  - `ignoreHostKey` (bool, 可选) - 跳过主机密钥验证

**返回值：** `SSHClient` 对象

**示例：**
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

创建新的 SSH 客户端但不连接。

**返回值：** `SSHClient` 对象（调用 `connect()` 建立连接）

**示例：**
```xxl
var client = ssh.newClient()
client.connect("server.com", 22, "user", "pass")
```

---

#### `ssh.exec(host, port, user, password, command)`

快速执行单个命令，无需管理连接。

**参数：**
- `host`, `port`, `user`, `password` - 连接参数
- `command` (string) - 要执行的命令

**返回值：** `string` - 命令输出

**示例：**
```xxl
var output = ssh.exec("192.168.1.100", 22, "root", "pass", "uptime")
```

---

#### `ssh.upload(host, port, user, password, localPath, remotePath)`

快速上传文件，无需管理连接。

**示例：**
```xxl
ssh.upload("server.com", 22, "user", "pass", "file.txt", "/remote/file.txt")
```

---

#### `ssh.download(host, port, user, password, remotePath, localPath)`

快速下载文件，无需管理连接。

**示例：**
```xxl
ssh.download("server.com", 22, "user", "pass", "/remote/file.txt", "file.txt")
```

---

#### `ssh.uploadBytes(host, port, user, password, bytes, remotePath)`

上传字节数据到远程文件。

**示例：**
```xxl
var data = io.readFileBytes("image.png")
ssh.uploadBytes("server.com", 22, "user", "pass", data, "/remote/image.png")
```

---

#### `ssh.downloadBytes(host, port, user, password, remotePath)`

下载远程文件为字节。

**返回值：** `bytes` - 文件内容

**示例：**
```xxl
var data = ssh.downloadBytes("server.com", 22, "user", "pass", "/remote/image.png")
io.writeFileBytes("local.png", data)
```

---

#### `ssh.uploadWithKey(host, port, user, keyPath, localPath, remotePath)`

使用密钥认证上传文件。

**示例：**
```xxl
ssh.uploadWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa", "file.txt", "/remote/file.txt")
```

---

#### `ssh.downloadWithKey(host, port, user, keyPath, remotePath, localPath)`

使用密钥认证下载文件。

**示例：**
```xxl
ssh.downloadWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa", "/remote/file.txt", "file.txt")
```

---

#### `ssh.testConnection(host, port, user, password)`

测试服务器是否可连接。

**返回值：** `bool` - 连接成功返回 true

**示例：**
```xxl
if (ssh.testConnection("server.com", 22, "user", "pass")) {
    println("服务器可连接")
} else {
    println("无法连接")
}
```

---

#### `ssh.testConnectionWithKey(host, port, user, keyPath)`

使用密钥认证测试连接。

**示例：**
```xxl
if (ssh.testConnectionWithKey("server.com", 22, "deploy", "~/.ssh/id_rsa")) {
    println("密钥认证成功")
}
```

---

### SSHClient 对象方法

#### `client.connect(host, port, user, password)`

建立 SSH 连接。

**抛出：** 连接失败时抛出错误

---

#### `client.connectWithKey(keyPath)`

使用私钥文件连接。

---

#### `client.connectWithKeyStr(keyString)`

使用私钥字符串连接。

---

#### `client.connectWithConfig(config)`

使用配置映射连接。

---

#### `client.is_connected()`

检查是否已连接。

**返回值：** `bool`

---

#### `client.getHost()`

获取连接的主机。

**返回值：** `string`

---

#### `client.getUser()`

获取连接的用户名。

**返回值：** `string`

---

#### `client.exec(command)`

执行命令并返回输出。

**参数：**
- `command` (string) - 要执行的命令

**返回值：** `string` - 合并的 stdout 和 stderr

**示例：**
```xxl
var hostname = client.exec("hostname")
var disk = client.exec("df -h /")
```

---

#### `client.execFull(command)`

执行命令，分离 stdout/stderr。

**返回值：** `map`，包含键：`stdout`、`stderr`、`exitCode`

**示例：**
```xxl
var result = client.execFull("some-command")
println("stdout:", result["stdout"])
println("stderr:", result["stderr"])
println("exit:", result["exitCode"])
```

---

#### `client.execStream(command, callback)`

逐行流式输出命令执行结果。

**参数：**
- `command` (string) - 要执行的命令
- `callback` (function) - 每行调用一次

**示例：**
```xxl
client.execStream("tail -f /var/log/syslog", func(line) {
    println("[SYSLOG]", line)
})
```

---

#### `client.upload(localPath, remotePath)`

上传文件到服务器。

**示例：**
```xxl
client.upload("./script.xxl", "/tmp/script.xxl")
```

---

#### `client.download(remotePath, localPath)`

从服务器下载文件。

**示例：**
```xxl
client.download("/var/log/app.log", "./logs/app.log")
```

---

#### `client.uploadBytes(bytes, remotePath)`

上传字节数据。

**示例：**
```xxl
var data = io.readFileBytes("image.png")
client.uploadBytes(data, "/remote/image.png")
```

---

#### `client.downloadBytes(remotePath)`

下载为字节。

**返回值：** `bytes`

---

#### `client.readFile(remotePath)`

读取远程文件内容。

**返回值：** `string`

---

#### `client.readfile_bytes(remotePath)`

读取远程文件为字节。

**返回值：** `bytes`

---

#### `client.writeFile(remotePath, content)`

写入内容到远程文件。

**示例：**
```xxl
client.writeFile("/tmp/config.json", "{\"port\": 8080}")
```

---

#### `client.writefile_bytes(remotePath, bytes)`

写入字节到远程文件。

---

#### `client.mkdir(path)`

创建目录。

---

#### `client.mkdirAll(path)`

递归创建目录。

---

#### `client.listDir(path)`

列出目录内容。

**返回值：** `map` 数组，包含键：`name`、`size`、`isDir`、`mode`、`modTime`

**示例：**
```xxl
var files = client.listDir("/tmp")
for (var i = 0; i < len(files); i = i + 1) {
    var f = files[i]
    println(f["name"], f["size"], f["isDir"])
}
```

---

#### `client.walkDir(path)`

递归遍历目录树。

**返回值：** 文件信息 map 数组

---

#### `client.uploadDir(localPath, remotePath)`

递归上传整个目录。

---

#### `client.downloadDir(remotePath, localPath)`

递归下载整个目录。

---

#### `client.stat(path)`

获取文件元数据。

**返回值：** `map`，包含键：`size`、`isFile`、`isDir`、`mode`、`modTime`

---

#### `client.exists(path)`

检查路径是否存在。

**返回值：** `bool`

---

#### `client.rename(oldPath, newPath)`

重命名/移动文件。

---

#### `client.remove(path)`

删除文件。

---

#### `client.removeDir(path)`

递归删除目录。

---

#### `client.chmod(path, mode)`

更改文件权限。

**示例：**
```xxl
client.chmod("/tmp/script.sh", 755)
```

---

#### `client.localForward(localPort, remoteHost, remotePort)`

创建 SSH 隧道（本地端口转发）。

**返回值：** `Listener` 对象，包含 `close()` 方法

**示例：**
```xxl
// 访问 local:8080 -> internal-service:80 通过 SSH 隧道
var listener = client.localForward(8080, "internal-service", 80)
// 保持隧道开启
sleep(3600)
listener.close()
```

---

#### `client.runScriptStr(scriptContent)`

在远程服务器上执行本地脚本内容。

**示例：**
```xxl
var script = io.readFile("./deploy.xxl")
var result = client.runScriptStr(script)
```

---

#### `client.close()`

关闭 SSH 连接。

---

## 示例

参见 [examples/ssh_examples.xxl](../examples/ssh_examples.xxl)，包含 15 个综合示例：

1. 快速命令（单行代码）
2. 文件传输（替代 sshrun）
3. SSH 客户端对象使用
4. SSH 密钥认证
5. 高级文件操作
6. 目录上传/下载
7. 流式命令输出
8. 端口转发
9. 远程脚本执行
10. 多台服务器批量操作
11. 完整部署工作流
12. 服务器监控
13. sshrun 模式替换
14. HEX 路径支持说明
15. 错误处理和重试逻辑

## 与 sshrun 的对比

| 功能 | sshrun | Xxlang |
|------|--------|--------|
| 执行命令 | ✅ | ✅ |
| 上传/下载 | ✅ | ✅（原生 SFTP） |
| 密码认证 | ✅ | ✅ |
| 密钥认证 | ✅ | ✅ |
| 目录传输 | ❌ | ✅ |
| 端口转发 | ❌ | ✅ |
| 流式输出 | ❌ | ✅ |
| 可编程 | ❌ | ✅ |
| 跨平台路径 | 需要 HEX 编码 | ✅ 原生支持 |

## 注意事项

### Windows Git Bash 路径处理

与 sshrun 不同（需要 HEX 路径编码来绕过 MSYS2 路径转换问题），Xxlang 的纯 Go 实现在所有平台上都能正确处理路径：

```xxl
// 在 Windows Git Bash 上直接工作：
ssh.upload("server.com", 22, "user", "pass", "file.txt", "/root/file.txt")

// 不需要 HEX 编码！
```

### 安全考虑

1. **主机密钥验证**：默认情况下不验证主机密钥。生产环境中建议实现主机密钥检查。

2. **凭据存储**：避免硬编码凭据。使用环境变量或安全的凭据存储：
   ```xxl
   var password = os.getenv("SSH_PASSWORD")
   var client = ssh.connect(host, port, user, password)
   ```

3. **密钥密码短语**：使用加密的私钥时，提供密码短语：
   ```xxl
   var config = {
       "host": "server.com",
       "user": "deploy",
       "keyPath": "~/.ssh/id_rsa",
       "keyPassphrase": "your-passphrase"
   }
   ```
