# File Operations in Xxlang

## Overview

Xxlang provides comprehensive file handling capabilities through multiple modules:

- **file** - Streaming file operations with File object
- **io** - Simple one-shot file operations
- **os** - Path utilities and system file operations
- **csv** - CSV file reading and writing
- **json** - JSON file reading and writing

## Table of Contents

- [File Module](#file-module)
  - [File Object](#file-object)
  - [Opening Files](#opening-files)
  - [Reading and Writing](#reading-and-writing)
  - [File Information](#file-information)
  - [Directory Operations](#directory-operations)
  - [Path Utilities](#path-utilities)
- [File Object Reference](#file-object-reference)
- [FileInfo Object Reference](#fileinfo-object-reference)
- [CSV Module](#csv-module)
- [JSON File Operations](#json-file-operations)
- [OS Module Extensions](#os-module-extensions)
- [Error Handling](#error-handling)

---

## File Module

Import the file module to access streaming file operations:

```xxl
import * as file from "file"
```

### File Object

The `File` object provides streaming file I/O with methods for reading, writing, seeking, and locking.

#### Opening Files

```xxl
// Open for reading
var f = file.openRead("data.txt")

// Open for writing (creates or truncates)
var f = file.openWrite("output.txt")

// Open for appending
var f = file.openAppend("log.txt")

// Open with specific mode
var f = file.open("data.bin", file.MODE_RW)  // read-write
```

**Open Modes:**
| Constant | Value | Description |
|----------|-------|-------------|
| `MODE_READ` | `"r"` | Read-only |
| `MODE_WRITE` | `"w"` | Write-only (truncates) |
| `MODE_APPEND` | `"a"` | Append-only |
| `MODE_RW` | `"rw"` | Read and write |
| `MODE_RWPLUS` | `"rw+"` | Read and write (creates if not exists) |

#### Reading and Writing

```xxl
// Streaming read
var f = file.openRead("data.txt")
var content = f.readAll()      // Read entire file
var line = f.readLine()        // Read one line
var chunk = f.read(1024)       // Read 1024 bytes as string
var bytes = f.readBytes(100)   // Read 100 bytes as array
f.close()

// Streaming write
var f = file.openWrite("output.txt")
f.write("Hello, ")
f.writeLine("World!")
f.flush()  // Ensure data is written to disk
f.close()
```

#### File Positioning

```xxl
var f = file.open("data.bin", file.MODE_RW)

// Get current position
var pos = f.tell()

// Seek to position
f.seek(100)                    // Position 100 from start
f.seek(50, file.SEEK_CURRENT)  // 50 bytes forward
f.seek(-10, file.SEEK_END)     // 10 bytes before end

// Truncate file
f.truncate(1000)  // Set file size to 1000 bytes

f.close()
```

**Seek Constants:**
| Constant | Value | Description |
|----------|-------|-------------|
| `SEEK_START` | `0` | Seek from beginning |
| `SEEK_CURRENT` | `1` | Seek from current position |
| `SEEK_END` | `2` | Seek from end |

#### File Locking

```xxl
var f = file.openWrite("locked.txt")

// Exclusive lock (blocks until acquired)
f.lock(file.LOCK_EXCLUSIVE)

// Shared lock (for reading)
f.lock(file.LOCK_SHARED)

// Non-blocking lock attempt
var acquired = f.lock(file.LOCK_EXCLUSIVE, false)

// Release lock
f.unlock()

f.close()
```

**Lock Types:**
| Constant | Value | Description |
|----------|-------|-------------|
| `LOCK_SHARED` | `1` | Shared/read lock |
| `LOCK_EXCLUSIVE` | `2` | Exclusive/write lock |

### Quick File Operations

For simple use cases, use these one-shot operations:

```xxl
// Read entire file
var content = file.readAll("data.txt")

// Write entire file
file.writeAll("output.txt", "Hello, World!")

// Append to file
file.appendAll("log.txt", "New entry\n")

// Read lines into array
var lines = file.readLines("data.txt")
for (var i = 0; i < lines.len(); i = i + 1) {
    io.println("Line " + i.toStr() + ": " + lines[i])
}

// Write lines from array
var lines = ["Line 1", "Line 2", "Line 3"]
file.writeLines("output.txt", lines)
```

### File Information

```xxl
// Get detailed file info
var info = file.stat("data.txt")
io.println("Name: " + info.name())
io.println("Size: " + info.size().toStr() + " bytes")
io.println("IsDir: " + info.isDir().toStr())
io.println("ModTime: " + info.modTime())
io.println("Mode: " + info.modeStr())

// Quick checks
var exists = file.exists("data.txt")
var isFile = file.isFile("data.txt")
var isDir = file.isDir("data.txt")
var size = file.size("data.txt")
var modTime = file.modTime("data.txt")
```

### Directory Operations

```xxl
// Create directory
file.mkdir("path/to/directory")

// List directory contents
var entries = file.listDir("path/to/dir")
for (var i = 0; i < entries.len(); i = i + 1) {
    io.println(entries[i])
}

// List with details (returns array of [name, size, isDir, modTime])
var details = file.listDirFull("path/to/dir")

// Remove file
file.remove("temp.txt")

// Remove directory and contents
file.removeDir("temp_dir")

// Copy file
file.copy("source.txt", "dest.txt")

// Move/rename file
file.move("old.txt", "new.txt")
```

### Path Utilities

```xxl
// Get absolute path
var abs = file.abs("relative/path")

// Get base name
var name = file.base("/path/to/file.txt")  // "file.txt"

// Get directory
var dir = file.dir("/path/to/file.txt")    // "/path/to"

// Get extension
var ext = file.ext("/path/to/file.txt")    // ".txt"

// Join paths
var path = file.join("path", "to", "file.txt")  // "path/to/file.txt"

// Glob pattern matching
var goFiles = file.glob("*.go")
var allFiles = file.glob("**/*.txt")

// Current working directory
var cwd = file.cwd()

// Change directory
file.chdir("/new/path")

// Temporary files
var tempPath = file.tempFile()        // Create temp file
var tempDir = file.tempDir("prefix-") // Create temp directory

// Change permissions (Unix-style octal)
file.chmod("script.sh", 0o755)
```

---

## File Object Reference

The `File` object provides streaming file I/O.

### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `close()` | `f.close()` | Close the file handle |
| `read(n)` | `f.read(1024)` | Read up to n bytes as string |
| `readBytes(n)` | `f.readBytes(100)` | Read up to n bytes as integer array |
| `readLine()` | `f.readLine()` | Read one line (returns null at EOF) |
| `readAll()` | `f.readAll()` | Read all remaining content |
| `write(s)` | `f.write("text")` | Write string to file |
| `writeLine(s)` | `f.writeLine("text")` | Write string with newline |
| `seek(offset, whence)` | `f.seek(100, 0)` | Set file position |
| `tell()` | `f.tell()` | Get current position |
| `flush()` | `f.flush()` | Flush buffered data to disk |
| `isOpen()` | `f.isOpen()` | Check if file is open |
| `name()` | `f.name()` | Get file path |
| `mode()` | `f.mode()` | Get open mode |
| `lock(type, blocking)` | `f.lock(2, true)` | Place file lock |
| `unlock()` | `f.unlock()` | Release file lock |
| `truncate(size)` | `f.truncate(1000)` | Truncate file to size |
| `stat()` | `f.stat()` | Get FileInfo object |

### Example: Streaming Copy

```xxl
import * as file from "file"

func copyFile(src, dst) {
    var srcFile = file.openRead(src)
    var dstFile = file.openWrite(dst)

    var bufferSize = 8192
    while (true) {
        var chunk = srcFile.read(bufferSize)
        if (chunk.len() == 0) {
            break
        }
        dstFile.write(chunk)
    }

    srcFile.close()
    dstFile.close()
}

copyFile("large_file.bin", "copy.bin")
```

---

## FileInfo Object Reference

The `FileInfo` object contains file metadata.

### Methods

| Method | Return Type | Description |
|--------|-------------|-------------|
| `name()` | string | Base name of file |
| `size()` | int | Size in bytes |
| `mode()` | int | File mode bits |
| `modeStr()` | string | Mode as octal string (e.g., "755") |
| `modTime()` | string | Modification time formatted |
| `modTimeUnix()` | int | Modification time in milliseconds |
| `isDir()` | bool | True if directory |
| `isFile()` | bool | True if regular file |
| `isSymlink()` | bool | True if symbolic link |
| `path()` | string | Full path to file |

### Example

```xxl
var info = file.stat("document.pdf")
io.println("File: " + info.name())
io.println("Size: " + info.size().toStr() + " bytes")
io.println("Modified: " + info.modTime())
io.println("Permissions: " + info.modeStr())

if (info.isDir()) {
    io.println("This is a directory")
}
```

---

## CSV Module

Import the CSV module for spreadsheet-style file handling:

```xxl
import * as csv from "csv"
```

### Reading CSV Files

```xxl
// Read as array of arrays
var data = csv.read("data.csv")
io.println("Rows: " + data.len().toStr())
io.println("First cell: " + data[0][0])

// Read with header row (returns array of maps)
var records = csv.readWithHeader("users.csv")
for (var i = 0; i < records.len(); i = i + 1) {
    io.println("Name: " + records[i]["name"])
    io.println("Age: " + records[i]["age"])
}

// Custom delimiter
var tsvData = csv.read("data.tsv", "\t")
```

### Writing CSV Files

```xxl
// Write array of arrays
var data = [
    ["name", "age", "city"],
    ["Alice", "30", "New York"],
    ["Bob", "25", "Los Angeles"]
]
csv.write("output.csv", data)

// Write array of maps with header
var records = [
    {"name": "Alice", "age": "30"},
    {"name": "Bob", "age": "25"}
]
var headers = ["name", "age"]
csv.writeWithHeader("output.csv", records, headers)

// Append to existing file
var newRow = ["Charlie", "35", "Chicago"]
csv.append("output.csv", [newRow])
```

### String Operations

```xxl
// Parse CSV string
var csvStr = "a,b,c\n1,2,3"
var data = csv.parse(csvStr)

// Convert to CSV string
var csvStr = csv.stringify(data)
```

### CSV Utility Functions

```xxl
// Get column
var names = csv.column(data, 0)

// Get row
var row = csv.row(data, 0)

// Count
var rows = csv.rowCount(data)
var cols = csv.colCount(data)

// Filter and map
var filtered = csv.filterRows(data, func(row) {
    return row[1].toInt() > 25
})

// Skip and take
var page2 = csv.take(csv.skip(data, 10), 10)
```

---

## JSON File Operations

The json module provides file-based JSON operations:

```xxl
import * as json from "json"
```

### Reading JSON Files

```xxl
// Read and parse JSON file
var config = json.readFile("config.json")
io.println("Server: " + config["server"])
io.println("Port: " + config["port"].toStr())
```

### Writing JSON Files

```xxl
// Write object to JSON file
var config = {
    "server": "localhost",
    "port": 8080,
    "debug": true,
    "hosts": ["server1", "server2"]
}
json.writeFile("config.json", config)

// Write with pretty formatting
json.writeFilePretty("config.json", config, "  ")

// Update existing JSON file (merge)
json.updateFile("config.json", {
    "port": 9090,
    "newOption": "value"
})

// Append to JSON array file
json.appendToArrayFile("logs.json", {
    "timestamp": time.now(),
    "message": "Event logged"
})
```

---

## OS Module Extensions

The os module includes additional path and file utilities:

```xxl
import * as os from "os"
```

### Path Operations

```xxl
// Split path into directory and file
var parts = os.split("/path/to/file.txt")  // ["/path/to/", "file.txt"]

// Get relative path
var rel = os.relative("/home/user", "/home/user/docs/file.txt")
// Result: "docs/file.txt"

// Get volume name (Windows)
var vol = os.volumeName("C:\\path\\file.txt")  // "C:"

// Glob pattern matching
var matches = os.glob("src/**/*.go")
```

### Directory Walking

```xxl
// Walk directory tree
var allFiles = os.walkDir("src")
for (var i = 0; i < allFiles.len(); i = i + 1) {
    io.println(allFiles[i])
}
```

### Symlink Operations

```xxl
// Create symbolic link
os.symlink("target.txt", "link.txt")

// Read link target
var target = os.readlink("link.txt")

// Check if path is a symlink
var isLink = os.isLink("link.txt")

// Get info without following symlinks
var info = os.lstat("link.txt")
```

---

## Error Handling

When working with files, errors can occur for various reasons. Xxlang provides patterns to handle these errors gracefully.

### Checking File Existence

Always verify a file exists before operations:

```xxl
import * as file from "file"

func safeReadFile(path) {
    if (!file.exists(path)) {
        return error("File not found: " + path)
    }
    return file.readAll(path)
}

var result = safeReadFile("config.txt")
if (isError(result)) {
    pln("Error:", result)
} else {
    pln("Content:", result)
}
```

### Result Pattern

Use a result map for detailed error information:

```xxl
func readConfig(path) {
    // Check existence
    if (!file.exists(path)) {
        return {"ok": false, "error": "File not found", "path": path}
    }

    // Check if it's a file
    if (!file.isFile(path)) {
        return {"ok": false, "error": "Path is not a file", "path": path}
    }

    // Try to read
    var content = file.readAll(path)
    if (isError(content)) {
        return {"ok": false, "error": "Read failed: " + content.toStr(), "path": path}
    }

    return {"ok": true, "data": content}
}

var result = readConfig("config.json")
if (result["ok"]) {
    pln("Config loaded:", result["data"])
} else {
    pln("Failed:", result["error"])
}
```

### Safe File Operations

Create helper functions for common operations with built-in error handling:

```xxl
// Safe write with backup
func safeWrite(path, content) {
    // Create backup if file exists
    if (file.exists(path)) {
        var backup = path + ".bak"
        file.copy(path, backup)
    }

    // Write new content
    var f = file.openWrite(path)
    if (isError(f)) {
        return error("Cannot open file for writing: " + path)
    }

    f.write(content)
    f.close()
    return true
}

// Safe directory creation
func ensureDir(path) {
    if (file.exists(path)) {
        if (!file.isDir(path)) {
            return error("Path exists but is not a directory: " + path)
        }
        return true
    }
    return file.mkdir(path)
}

// Usage
if (isError(ensureDir("data/output"))) {
    pln("Failed to create directory")
}
```

### File Locking Errors

Handle file locking failures:

```xxl
func writeWithLock(path, content) {
    var f = file.openWrite(path)
    if (isError(f)) {
        return error("Cannot open: " + path)
    }

    // Try non-blocking lock
    var locked = f.lock(file.LOCK_EXCLUSIVE, false)
    if (!locked) {
        f.close()
        return error("File is locked by another process")
    }

    f.write(content)
    f.unlock()
    f.close()
    return true
}

var result = writeWithLock("shared.txt", "data")
if (isError(result)) {
    pln("Write failed:", result)
}
```

### Common Error Scenarios

| Operation | Possible Errors | Handling |
|-----------|-----------------|----------|
| `file.openRead` | File not found, permission denied | Check `exists()` and permissions first |
| `file.openWrite` | Permission denied, disk full | Create parent directories, check space |
| `file.readAll` | File too large, encoding error | Read in chunks, validate encoding |
| `file.mkdir` | Permission denied, parent not exists | Create parents with `mkdirAll` |
| `file.remove` | File not found, permission denied | Check existence, close open handles |
| `file.copy` | Source not found, dest not writable | Verify both paths, check disk space |

### Cleanup Pattern

Use defer-like patterns for cleanup (manual in Xxlang):

```xxl
func processFile(path) {
    var f = file.openRead(path)
    if (isError(f)) {
        return error("Cannot open: " + path)
    }

    // Process file...
    var content = f.readAll()

    // Always close, even on error
    f.close()

    if (isError(content)) {
        return error("Read failed")
    }

    return content
}
```

---

## Complete Example

```xxl
// Complete file handling example
import * as file from "file"
import * as csv from "csv"
import * as json from "json"
import * as os from "os"
import * as io from "io"

// Create data directory
file.mkdir("data")

// Write config file
var config = {
    "name": "MyApp",
    "version": "1.0.0",
    "settings": {
        "debug": true,
        "logLevel": "info"
    }
}
json.writeFilePretty("data/config.json", config, "  ")

// Write CSV data
var users = [
    ["id", "name", "email"],
    [1, "Alice", "alice@example.com"],
    [2, "Bob", "bob@example.com"]
]
csv.write("data/users.csv", users)

// Process CSV with streaming
var f = file.openRead("data/users.csv")
var header = f.readLine().split(",")
io.println("Columns: " + header.toStr())

while (true) {
    var line = f.readLine()
    if (line.typeOf() == "NULL") {
        break
    }
    var fields = line.split(",")
    io.println("User: " + fields[1] + " <" + fields[2] + ">")
}
f.close()

// Copy and move operations
file.copy("data/users.csv", "data/users_backup.csv")
file.move("data/users_backup.csv", "data/archive/users.csv")

// Clean up
file.removeDir("data")

io.println("File operations completed!")
```
