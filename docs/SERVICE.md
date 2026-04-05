# Xxlang Service Mode

Xxlang supports running as a system service on Windows, Linux, and macOS. This allows Xxlang to run in the background and automatically execute scripts on startup.

## Table of Contents

- [Overview](#overview)
- [Service Commands](#service-commands)
- [Installation](#installation)
- [Directory Structure](#directory-structure)
- [Script Types](#script-types)
- [Configuration](#configuration)
- [Logging](#logging)
- [Use Cases](#use-cases)
- [Troubleshooting](#troubleshooting)

## Overview

Xxlang service mode provides:

- **Cross-Platform Support** - Windows, Linux, and macOS
- **Automatic Startup** - Service starts with the system
- **Script Execution** - Run scripts automatically on startup
- **Background Tasks** - Execute periodic tasks in goroutines
- **One-time Tasks** - Scripts that run once and auto-delete
- **Service Management** - Install, uninstall, start, stop, restart commands

## Service Commands

| Command | Description |
|---------|-------------|
| `-installService` | Install Xxlang as a system service |
| `-removeService` | Stop and remove the service (alias: `-uninstallService`) |
| `-startService` | Start the service |
| `-stopService` | Stop the service |
| `-restartService` | Restart the service |
| `-reinstallService` | Reinstall the service (stop, uninstall, install, start) |
| `-service` | Run in service mode (called by service manager) |

### Usage Examples

```bash
# Install service
xxl -installService

# Start service
xxl -startService

# Stop service
xxl -stopService

# Restart service
xxl -restartService

# Remove service
xxl -removeService

# Reinstall service
xxl -reinstallService
```

## Installation

### Windows

1. **Install as service** (requires Administrator privileges):
   ```cmd
   xxl.exe -installService
   ```

2. **Verify installation**:
   ```cmd
   sc query xxlang
   ```

3. **Start the service**:
   ```cmd
   xxl.exe -startService
   ```

4. **View logs**:
   ```cmd
   type C:\xxlang\xxlang.log
   ```

### Linux

1. **Install as service** (requires root privileges):
   ```bash
   sudo ./xxl -installService
   ```

2. **Verify installation**:
   ```bash
   systemctl status xxlang
   # or
   service xxlang status
   ```

3. **Start the service**:
   ```bash
   sudo ./xxl -startService
   ```

4. **View logs**:
   ```bash
   cat /xxlang/xxlang.log
   ```

### macOS

1. **Install as service** (requires sudo):
   ```bash
   sudo ./xxl -installService
   ```

2. **Load the LaunchDaemon**:
   ```bash
   sudo launchctl load /Library/LaunchDaemons/xxlang.plist
   ```

3. **View logs**:
   ```bash
   cat /xxlang/xxlang.log
   ```

## Directory Structure

### Windows
```
C:\xxlang\                    # Service base directory
├── xxlangwin.cfg            # Configuration file
├── xxlang.log               # Log file
├── task*.xxl                # Regular tasks (run once at startup)
├── threadTask*.xxl          # Thread tasks (run in goroutines)
└── autoRemoveTask*.xxl      # One-time tasks (deleted after execution)
```

### Linux/macOS
```
/xxlang/                     # Service base directory
├── xxlanglinux.cfg         # Configuration file
├── xxlang.log              # Log file
├── task*.xxl               # Regular tasks
├── threadTask*.xxl         # Thread tasks
└── autoRemoveTask*.xxl     # One-time tasks
```

## Script Types

The service supports three types of scripts based on filename patterns:

### 1. Regular Tasks (`task*.xxl`)

Executed **once at service startup** in sequence.

**Example: `task001.xxl`**
```xxl
// task001.xxl - Initialization task
println("=== Initialization Task ===")
println("Starting at:", time.dateTime())

// Initialize database connection
var db = dbConnect("sqlite", "/xxlang/data/app.db")

// Clean up expired data
dbExec(db, "DELETE FROM sessions WHERE expired < now()")

println("Initialization completed")
```

### 2. Thread Tasks (`threadTask*.xxl`)

Executed in **separate goroutines** (concurrent execution). Checked every 5 seconds.

**Example: `threadTask001.xxl`**
```xxl
// threadTask001.xxl - Background monitoring task
println("Starting background monitor...")

while (true) {
    // Check every minute
    sleep(60)
    
    var count = dbQueryCount(db, "SELECT COUNT(*) FROM pending_jobs")
    if (count > 0) {
        println("Pending jobs:", count)
        // Process jobs...
    }
}
```

### 3. Auto-Remove Tasks (`autoRemoveTask*.xxl`)

Executed **once and automatically deleted** after completion. Checked every 5 seconds.

**Example: `autoRemoveTask001.xxl`**
```xxl
// autoRemoveTask001.xxl - Cache cleanup after update
println("Executing cache cleanup...")

// Clean temporary files
io.removeDir("/xxlang/temp/cache")

// Rebuild indexes
println("Rebuilding indexes...")
// ... index rebuilding code ...

println("Cleanup completed, script will auto-delete")
```

## Configuration

### Configuration File

The service configuration file (`xxlangwin.cfg` or `xxlanglinux.cfg`) uses simple key=value format:

```properties
# Base path override
xxlangBasePath=/custom/path

# Custom settings
customSetting=value
```

### Loading Configuration

```xxl
// In your service script
var config = loadSimpleMap("/xxlang/xxlanglinux.cfg")
var basePath = config["xxlangBasePath"]
```

## Logging

### Log File Location

- **Windows**: `C:\xxlang\xxlang.log`
- **Linux/macOS**: `/xxlang/xxlang.log`

### Log Format

```
[2026-04-05 18:13:42] xxlang Vdev
[2026-04-05 18:13:42] os: linux, basePath: /xxlang, config: xxlanglinux.cfg
[2026-04-05 18:13:42] command-line args: [/root/xxl -service]
[2026-04-05 18:13:42] Service started.
[2026-04-05 18:13:42] running task: /xxlang/task001.xxl
[2026-04-05 18:13:42] task completed: /xxlang/task001.xxl
[2026-04-05 18:13:42] starting thread task: /xxlang/threadTask001.xxl
[2026-04-05 18:13:42] thread task completed
```

### Logging Functions

```xxl
// Use println/println for output (automatically logged)
println("Task completed")

// Write to log file directly
io.appendFile("/xxlang/xxlang.log", "Custom log entry\n")
```

## Use Cases

### 1. Scheduled Data Processing

```xxl
// threadTask_data_processor.xxl
import "io"
import "time"

println("Starting data processor...")

while (true) {
    // Process every hour
    sleep(3600)
    
    var files = io.readDir("/xxlang/input")
    for (var i = 0; i < len(files); i = i + 1) {
        if (strings.hasSuffix(files[i], ".dat")) {
            // Process file
            processFile("/xxlang/input/" + files[i])
            // Move to processed
            io.moveFile("/xxlang/input/" + files[i], "/xxlang/processed/" + files[i])
        }
    }
}
```

### 2. Health Check Monitor

```xxl
// threadTask_health_check.xxl
import "net"
import "db"

var endpoints = [
    "http://api1.example.com/health",
    "http://api2.example.com/health"
]

while (true) {
    sleep(300)  // Check every 5 minutes
    
    for (var i = 0; i < len(endpoints); i = i + 1) {
        var result = net.get(endpoints[i])
        var statusCode = result[1]
        
        if (statusCode != 200) {
            // Log failure
            io.appendFile("/xxlang/health.log", 
                time.dateTime() + " - " + endpoints[i] + " failed\n")
            
            // Send alert (email, webhook, etc.)
            sendAlert(endpoints[i] + " is down")
        }
    }
}
```

### 3. Automatic Backup

```xxl
// task_backup.xxl - Runs once at startup
import "io"
import "time"

var timestamp = time.dateTime()
var backupFile = "/xxlang/backups/backup_" + timestamp + ".zip"

println("Starting backup...")

// Create backup
io.zipDir("/xxlang/data", backupFile)

// Clean old backups (keep last 7 days)
var oldBackups = io.readDir("/xxlang/backups")
// ... cleanup logic ...

println("Backup completed:", backupFile)
```

## Troubleshooting

### Service Won't Start

1. **Check permissions**: Ensure running as Administrator (Windows) or root (Linux/macOS)
2. **Check logs**: Review `xxlang.log` for error messages
3. **Verify installation**: Re-run `-installService`

### Scripts Not Executing

1. **Check file naming**: Ensure files match `task*.xxl`, `threadTask*.xxl`, or `autoRemoveTask*.xxl`
2. **Check file location**: Scripts must be in the service base directory
3. **Check script syntax**: Run script manually to verify no syntax errors

### Service Crashes

1. **Review logs**: Check `xxlang.log` for error messages
2. **Test scripts**: Run scripts manually to identify issues
3. **Use try/catch**: Wrap code in try/catch to handle errors gracefully

### Linux/macOS: Permission Denied

```bash
# Set proper permissions
chmod 755 /xxlang
chmod 644 /xxlang/*.xxl
chmod +x /xxl
```

### Windows: Access Denied

- Run command prompt as Administrator
- Ensure the service user has read/write access to `C:\xxlang`

## Cross-Platform Considerations

1. **Path separators**: Use `io.pathJoin()` for cross-platform paths
2. **File permissions**: Linux/macOS require executable permissions
3. **Service names**: The service is named `xxlang` on all platforms
4. **Log file**: Same location relative to base directory

## API Reference

### Service-related Built-in Functions

No special built-in functions are needed. All standard Xxlang functions work in service mode.

### Accessing Service Context

```xxl
// Check if running in service mode
if (runModeG == "service") {
    println("Running as service")
}

// Access base path
println("Base path:", basePathG)
```

## Best Practices

1. **Use absolute paths** for file operations in service mode
2. **Log errors** to help with debugging
3. **Handle exceptions** with try/catch to prevent service crashes
4. **Use goroutines** for long-running background tasks
5. **Clean up resources** (database connections, files) properly
6. **Test scripts manually** before deploying as service tasks
7. **Monitor log file size** and implement rotation if needed
8. **Use meaningful script names** (e.g., `task001_init.xxl`, `threadTask002_monitor.xxl`)

## Security Considerations

1. **Run with minimal privileges** - Don't run as root/Administrator unless necessary
2. **Restrict directory access** - Limit service directory permissions
3. **Validate script content** - Only run trusted scripts
4. **Use secure configuration** - Don't store sensitive data in plain text config files
5. **Monitor service logs** - Regularly review for suspicious activity

## See Also

- [Microservice Mode](MICROSERVICE.md) - HTTP/HTTPS server functionality
- [Standard Library](STDLIB.md) - Available modules and functions
- [Language Reference](LANGUAGE.md) - Complete language syntax
