# Practical Examples

This directory contains **tested, production-ready** Xxlang script examples covering various real-world tasks.

All scripts have been verified to run correctly and produce the expected results shown in comments.

## Scripts

| File | Category | Description |
|------|----------|-------------|
| `weather_query.xxl` | HTTP / API | Query current weather for multiple cities via Open-Meteo public API (no key needed) |
| `ssh_key_auth_examples.xxl` | SSH | SSH key-based authentication (key file, key string, passphrase, config) |
| `backup_local_to_local.xxl` | Backup | Local-to-local incremental backup (full, incremental, mirror, hash, exclude) |
| `backup_local_to_remote.xxl` | Backup / SSH | Local-to-remote incremental backup via SFTP |
| `backup_remote_to_local.xxl` | Backup / SSH | Remote-to-local incremental backup via SFTP |
| `backup_remote_to_remote.xxl` | Backup / SSH | Server-to-server incremental sync via SFTP (two SSH connections) |

## Prerequisites

- Xxlang runtime (`xxl` command)
- For SSH/backup scripts: a reachable SSH server with SFTP subsystem
- For API scripts: internet access
- Replace placeholder values (`YOUR_HOST`, `YOUR_USER`, etc.) with real credentials

## Quick Start

### Weather Query

```xxl
// No setup needed - uses free public API
var d = getWebObject("https://api.open-meteo.com/v1/forecast?latitude=39.9&longitude=116.4&current=temperature_2m&timezone=auto")
pln("Beijing temp:", d["current"]["temperature_2m"], "C")
```

### Local Backup

```xxl
import "backup"
var result = backup.localToLocal("D:/myproject/src", "E:/backup/myproject")
pln(result.summary())
```

### SSH Key Authentication

```xxl
import "ssh"

// Login with private key file
var c = ssh.connectWithKey("YOUR_SERVER", 22, "YOUR_USER", "~/.ssh/id_rsa")
pln("connected:", c.isConnected())

// Or use config (supports keyPath, keyStr, keyPassphrase)
var c2 = ssh.connectWithConfig({
    "host": "YOUR_SERVER", "port": 22,
    "user": "YOUR_USER", "keyPath": "~/.ssh/id_rsa",
    "ignoreHostKey": true
})
```

### Remote Backup

```xxl
import "ssh"
import "backup"

var c = ssh.connectWithConfig({
    "host": "YOUR_HOST", "port": 22,
    "user": "YOUR_USER", "password": "YOUR_PASSWORD",
    "ignoreHostKey": true
})
pln(backup.localToRemote(c, "D:/myproject", "/var/backup/myproject").summary())
c.close()
```

## Features Demonstrated

### Backup Module
| Feature | Description |
|---------|-------------|
| Full backup | Copy all files from source to target |
| Incremental skip | Skip unchanged files (same size + mtime) |
| Detect modifications | Re-copy files whose size or mtime changed |
| Detect new files | Copy newly added files |
| Mirror mode | Delete target files not present in source |
| Hash comparison | Content-based comparison (MD5/SHA1), ignores timestamp |
| Exclude patterns | Skip files matching glob patterns (e.g. `*.log`, `tmp/`) |
| Non-mirror safety | Deleted source files kept in target unless mirror mode |
| Conflict detection | Preview which files would be overwritten |

### HTTP / API
| Feature | Description |
|---------|-------------|
| GET request | `getWeb(url)` returns raw string, `getWebObject(url)` returns parsed object |
| JSON parsing | `getWebObject` auto-parses JSON responses into Xxlang maps |
| Error handling | Check `typeOf(result) != "MAP"` for failed requests |

### SSH / SFTP
| Feature | Description |
|---------|-------------|
| Password auth | `ssh.connect(host, port, user, password)` |
| Key file auth | `ssh.connectWithKey(host, port, user, keyPath)` |
| Key string auth | `ssh.connectWithKeyStr(host, port, user, keyStr)` |
| Full config | `ssh.connectWithConfig({...})` supports password, keyPath, keyStr, keyPassphrase |
| File operations | `c.writeFile()`, `c.readFile()`, `c.remove()`, `c.removeDir()` |
| Directory ops | `c.mkdirAll()`, `c.exists()`, `c.isDir()` |
| Command exec | `c.exec()` for shell commands (use sparingly, prefer SFTP ops) |

## Notes

- All SSH file operations use SFTP binary transfer, ensuring binary-safe copy
- Paths use forward slashes (`/`) for cross-platform compatibility
- `sleep(1500)` before modifying files ensures mtime changes are detectable
- For production, prefer SSH key-based authentication over passwords
