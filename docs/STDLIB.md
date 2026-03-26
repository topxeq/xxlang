# Xxlang Standard Library Reference

## Overview

Xxlang includes a comprehensive standard library organized into modules. All modules are imported using the `import` statement.

```xxl
import "io"
io.println("Hello, World!")
```

## Table of Contents

- [io](#io) - Input/output operations
- [file](#file) - Streaming file operations with File object
- [os](#os) - Operating system utilities and configuration
- [string](#string) - String utilities
- [chars](#chars) - Unicode character handling
- [stringbuilder](#stringbuilder) - Efficient string concatenation
- [bytesbuffer](#bytesbuffer) - Efficient byte buffer operations
- [math](#math) - Mathematical functions
- [array](#array) - Array utilities
- [json](#json) - JSON encoding/decoding
- [xml](#xml) - XML encoding/decoding
- [csv](#csv) - CSV file reading and writing
- [xlsx](#xlsx-module) - Excel file handling
- [regex](#regex) - Regular expressions
- [crypto](#crypto) - Cryptographic functions
- [time](#time) - Time and date functions
- [fmt](#fmt) - Formatting utilities
- [encoding](#encoding) - Base64 and hex encoding/decoding
- [uuid](#uuid) - UUID generation
- [debug](#debug) - Debugging utilities
- [Other Modules](#other-modules) - bytes, collections, env, fp, log, net, sort, strconv, text, validate

> **See also:** [File Operations Guide](FILE.md) for comprehensive file handling documentation.

---

## io

Input/output operations for reading, writing, and console interaction.

### Console Functions

#### print(args...)
Prints arguments without a trailing newline.

```xxl
io.print("Hello")
io.print(" ")
io.print("World")
// Output: Hello World
```

#### println(args...)
Prints arguments separated by spaces, followed by a newline.

```xxl
io.println("Hello", "World")  // "Hello World\n"
io.println(42, "is the answer")  // "42 is the answer\n"
```

#### printf(format, args...)
Formatted print using Go-style format specifiers.

```xxl
printf("Name: %s, Age: %d\n", "Alice", 30)
printf("Pi: %.2f\n", 3.14159)  // "Pi: 3.14"
```

#### readLine()
Reads a line from stdin.

```xxl
io.print("Enter name: ")
var name = io.readLine()
io.println("Hello, " + name)
```

#### readStdin()
Reads all content from stdin and returns as string. Useful for pipe processing.

```xxl
var content = io.readStdin()
```

#### readStdinBytes()
Reads all content from stdin and returns as byte array.

```xxl
var bytes = io.readStdinBytes()
```

#### writeStdout(content)
Writes a string to stdout without adding newline.

```xxl
io.writeStdout("Hello, World!")
```

#### writeStderr(content)
Writes a string to stderr without adding newline.

```xxl
io.writeStderr("Error message")
```

### File Functions

#### readFile(path)
Reads entire file content as string.

```xxl
var content = io.readFile("data.txt")
io.println(content)
```

#### writeFile(path, content)
Writes string content to file.

```xxl
io.writeFile("output.txt", "Hello, File!")
```

#### appendFile(path, content)
Appends string content to file.

```xxl
appendFile("log.txt", "New log entry\n")
```

#### exists(path)
Returns true if file or directory exists.

```xxl
if (io.exists("config.json")) {
    io.println("Config found")
}
```

#### remove(path)
Deletes a file.

```xxl
remove("temp.txt")
```

#### mkdir(path)
Creates directory (including parent directories).

```xxl
mkdir("path/to/directory")
```

### System Functions

#### cwd()
Returns current working directory.

```xxl
io.println(io.cwd())  // "/home/user/project"
```

#### exit(code)
Exits with status code.

```xxl
exit(0)   // Success
exit(1)   // Error
```

#### env(key)
Gets environment variable value.

```xxl
var home = env("HOME")
var path = env("PATH")
```

#### setEnv(key, value)
Sets environment variable.

```xxl
setEnv("DEBUG", "true")
```

#### args()
Returns command-line arguments as array.

```xxl
var args = io.args()
io.println(args[0])  // Program name
```

---

## file

Streaming file operations with File object for read/write/seek/lock operations.

> **See [File Operations Guide](FILE.md) for complete documentation.**

### Quick Start

```xxl
import * as file from "file"

// Simple read/write
var content = file.readAll("data.txt")
file.writeAll("output.txt", "Hello!")

// Streaming operations
var f = file.openWrite("log.txt")
f.writeLine("Log entry 1")
f.writeLine("Log entry 2")
f.close()

// File information
var info = file.stat("data.txt")
io.println("Size: " + info.size().toStr())
io.println("Modified: " + info.modTime())
```

### Key Functions

| Function | Description |
|----------|-------------|
| `open(path, mode)` | Open file with mode ("r", "w", "a", "rw") |
| `openRead(path)` | Open for reading |
| `openWrite(path)` | Open for writing (truncates) |
| `openAppend(path)` | Open for appending |
| `create(path)` | Create new file |
| `readAll(path)` | Read entire file as string |
| `writeAll(path, content)` | Write string to file |
| `readLines(path)` | Read lines into array |
| `copy(src, dst)` | Copy file |
| `move(src, dst)` | Move/rename file |
| `exists(path)` | Check if file exists |
| `stat(path)` | Get FileInfo object |
| `glob(pattern)` | Match files by pattern |

---

## os

Operating system utilities and configuration management.

### Configuration

#### getConfigObj()
Returns the Xxlang configuration as a map object. The configuration is read from a JSON file with the following search priority:

1. `~/.xxl/settings.json` (user home directory)
2. `/.xxl/settings.json` (Linux/Unix systems)
3. `C:\.xxl\settings.json` (Windows systems)

Returns an empty map if no configuration file is found.

```xxl
import "os"
var cfg = os.getConfigObj()
pln(cfg["cloudUrlBase"])

// Example config file (~/.xxl/settings.json):
// {
//   "cloudUrlBase": "https://example.com/",
//   "timeout": 30,
//   "debug": true
// }
```

#### getConfigStr(name)
Reads a config string from a `.cfg` file. The file is searched in the following order:

1. `~/.xxl/<name>.cfg` (user home directory)
2. `/.xxl/<name>.cfg` (Linux/Unix systems)
3. `C:\.xxl\<name>.cfg` (Windows systems)

Returns `null` if the file is not found.

```xxl
import "os"
var token = os.getConfigStr("api_token")
if (token != null) {
    pln("Token found: " + token)
}
```

#### setConfigStr(name, value)
Writes a config string to a `.cfg` file in the user's home directory (`~/.xxl/<name>.cfg`). Creates the `.xxl` directory if it doesn't exist.

```xxl
import "os"
os.setConfigStr("api_token", "my-secret-token")

// Later, read it back
var token = os.getConfigStr("api_token")
pln(token)  // "my-secret-token"
```

### System Information

#### platform()
Returns the current operating system name.

```xxl
pln(os.platform())  // "linux", "windows", "darwin"
```

#### arch()
Returns the CPU architecture.

```xxl
pln(os.arch())  // "amd64", "arm64"
```

#### hostname()
Returns the system hostname.

```xxl
pln(os.hostname())
```

#### home()
Returns the user's home directory.

```xxl
pln(os.home())  // "/home/user" or "C:\Users\user"
```

#### temp()
Returns the system temporary directory.

```xxl
pln(os.temp())  // "/tmp" on Unix
```

#### cpus()
Returns the number of CPU cores.

```xxl
pln(os.cpus())  // 8
```

### File System Operations

#### join(paths...)
Joins path components.

```xxl
os.join("a", "b", "c.txt")  // "a/b/c.txt" (Unix) or "a\b\c.txt" (Windows)
```

#### base(path)
Returns the last element of a path.

```xxl
os.base("/path/to/file.txt")  // "file.txt"
```

#### dir(path)
Returns the directory part of a path.

```xxl
os.dir("/path/to/file.txt")  // "/path/to"
```

#### ext(path)
Returns the file extension.

```xxl
os.ext("/path/to/file.txt")  // ".txt"
```

#### abs(path)
Returns the absolute path.

```xxl
pln(os.abs("./file.txt"))  // "/current/dir/file.txt"
```

#### isAbs(path)
Returns true if path is absolute.

```xxl
os.isAbs("/path/to/file")  // true
os.isAbs("./file")          // false
```

#### stat(path)
Returns file information as an array [name, size, isDir, modTime].

```xxl
var info = os.stat("file.txt")
pln(info[0])  // name
pln(info[1])  // size in bytes
pln(info[2])  // is directory?
pln(info[3])  // modification time
```

#### isDir(path)
Returns true if path is a directory.

```xxl
os.isDir("/path/to/dir")  // true
```

#### isFile(path)
Returns true if path is a file.

```xxl
os.isFile("/path/to/file.txt")  // true
```

#### listDir(path)
Returns array of directory entry names.

```xxl
var files = os.listDir(".")
for (f in files) {
    pln(f)
}
```

#### mkdir(path)
Creates a directory.

```xxl
os.mkdir("path/to/new/dir")
```

#### rename(oldPath, newPath)
Renames a file or directory.

```xxl
os.rename("old.txt", "new.txt")
```

#### copy(srcPath, dstPath)
Copies a file.

```xxl
os.copy("source.txt", "dest.txt")
```

#### chmod(path, mode)
Changes file permissions (Unix only).

```xxl
os.chmod("script.sh", 0o755)
```

### Process Execution

#### exec(command)
Executes a shell command and returns [output, exitCode, error].

```xxl
var result = os.exec("ls -la")
pln(result[0])  // output
pln(result[1])  // exit code
```

#### shell(command)
Executes a command through the system shell.

```xxl
var result = os.shell("echo hello")
pln(result[0])  // "hello\n"
```

### Temporary Files

#### tempFile(pattern)
Creates a temporary file and returns its path.

```xxl
var tmp = os.tempFile("myapp-*.txt")
// Use the file...
```

#### tempDir(pattern)
Creates a temporary directory and returns its path.

```xxl
var tmpDir = os.tempDir("myapp-*")
// Use the directory...
```

---

## string

String manipulation utilities.

#### len(s)
Returns string length.

```xxl
len("hello")  // 5
```

#### upper(s)
Converts to uppercase.

```xxl
upper("hello")  // "HELLO"
```

#### lower(s)
Converts to lowercase.

```xxl
lower("HELLO")  // "hello"
```

#### trim(s)
Removes leading and trailing whitespace.

```xxl
trim("  hello  ")  // "hello"
```

#### substr(s, start, end)
Extracts substring.

```xxl
substr("hello", 1, 4)  // "ell"
```

#### split(s, sep)
Splits string by separator.

```xxl
split("a,b,c", ",")  // ["a", "b", "c"]
```

#### join(arr, sep)
Joins array elements with separator.

```xxl
join(["a", "b", "c"], "-")  // "a-b-c"
```

#### containsStr(s, substr)
Returns true if string contains substring.

```xxl
containsStr("hello", "ell")  // true
containsStr("hello", "xyz")  // false
```

#### replace(s, old, new)
Replaces all occurrences.

```xxl
replace("hello world", "l", "L")  // "heLLo worLd"
```

#### startsWith(s, prefix)
Returns true if string starts with prefix.

```xxl
startsWith("hello", "he")  // true
```

#### endsWith(s, suffix)
Returns true if string ends with suffix.

```xxl
endsWith("hello", "lo")  // true
```

#### indexOf(s, substr)
Returns index of first occurrence, or -1.

```xxl
indexOf("hello", "l")   // 2
indexOf("hello", "x")   // -1
```

---

## stringbuilder

Efficient string concatenation using a mutable string builder. Unlike regular string concatenation which creates a new string each time, StringBuilder uses an internal buffer for better performance.

#### create(capacity?)
Creates a new StringBuilder instance. Optional capacity parameter pre-allocates buffer size.

```xxl
import "stringbuilder"

// Create empty builder
var sb = stringbuilder.create()

// Create with initial capacity
var sb2 = stringbuilder.create(1000)
```

#### isStringBuilder(obj)
Returns true if the object is a StringBuilder.

```xxl
var sb = stringbuilder.create()
stringbuilder.isStringBuilder(sb)   // true
stringbuilder.isStringBuilder(42)   // false
```

### StringBuilder Methods

#### write(str)
Appends a string to the builder. Returns the number of bytes written.

```xxl
var sb = stringbuilder.create()
sb.write("Hello")
sb.write(" ")
sb.write("World")
pln(sb.toString())  // "Hello World"
```

#### writeLine(str)
Appends a string followed by a newline.

```xxl
var sb = stringbuilder.create()
sb.writeLine("Line 1")
sb.writeLine("Line 2")
pln(sb.toString())
// Line 1
// Line 2
```

#### toString()
Returns the accumulated string.

```xxl
var sb = stringbuilder.create()
sb.write("test")
var result = sb.toString()  // "test"
```

#### len()
Returns the current length of the accumulated string.

```xxl
var sb = stringbuilder.create()
sb.write("Hello")
pln(sb.len())  // 5
```

#### isEmpty()
Returns true if the builder is empty.

```xxl
var sb = stringbuilder.create()
pln(sb.isEmpty())  // true
sb.write("test")
pln(sb.isEmpty())  // false
```

#### clear()
Clears all content from the builder.

```xxl
var sb = stringbuilder.create()
sb.write("Hello")
sb.clear()
pln(sb.len())  // 0
```

#### reset()
Alias for `clear()`. Resets the builder to empty state.

```xxl
var sb = stringbuilder.create()
sb.write("Hello")
sb.reset()
pln(sb.isEmpty())  // true
```

#### grow(n)
Pre-allocates buffer capacity for better performance when the final size is known.

```xxl
var sb = stringbuilder.create()
sb.grow(1000)  // Pre-allocate 1000 bytes
sb.write("Hello")
```

### Performance Example

```xxl
import "stringbuilder"

// Efficient string building
var sb = stringbuilder.create()
for (var i = 0; i < 1000; i = i + 1) {
    sb.writeLine("Line " + str(i))
}
var result = sb.toString()
pln("Total lines:", result.len())
```

---

## bytesbuffer

Efficient byte buffer operations for working with binary data. BytesBuffer provides a mutable buffer for reading and writing bytes, similar to Go's `bytes.Buffer`.

**Note:** BytesBuffer is NOT thread-safe. For concurrent access, use external synchronization with `Mutex` or `RWMutex` from the `sync` module.

### Module Functions

#### create(capacity?)
Creates a new BytesBuffer instance. Optional capacity parameter pre-allocates buffer size.

```xxl
import "bytesbuffer"

// Create empty buffer
var buf = bytesbuffer.create()

// Create with initial capacity
var buf2 = bytesbuffer.create(1024)
```

#### fromBytes(arr)
Creates a BytesBuffer from an array of integers (0-255).

```xxl
var buf = bytesbuffer.fromBytes([72, 101, 108, 108, 111])  // "Hello"
```

#### fromString(str)
Creates a BytesBuffer from a string.

```xxl
var buf = bytesbuffer.fromString("Hello World")
```

#### isBytesBuffer(obj)
Returns true if the object is a BytesBuffer.

```xxl
var buf = bytesbuffer.create()
bytesbuffer.isBytesBuffer(buf)   // true
bytesbuffer.isBytesBuffer(42)    // false
```

### BytesBuffer Methods

#### write(data)
Writes a string or byte array to the buffer. Returns the number of bytes written.

```xxl
var buf = bytesbuffer.create()
buf.write("Hello")
buf.write([32, 87, 111, 114, 108, 100])  // " World"
pln(buf.toString())  // "Hello World"
```

#### writeByte(b)
Writes a single byte (0-255).

```xxl
var buf = bytesbuffer.create()
buf.writeByte(72)   // 'H'
buf.writeByte(105)  // 'i'
pln(buf.toString())  // "Hi"
```

#### writeInt16(n), writeInt32(n), writeInt64(n)
Writes integers in little-endian format.

```xxl
var buf = bytesbuffer.create()
buf.writeInt32(123456)
buf.writeInt64(9876543210)
```

#### writeFloat32(n), writeFloat64(n)
Writes floats in little-endian format.

```xxl
var buf = bytesbuffer.create()
buf.writeFloat32(3.14)
buf.writeFloat64(2.718281828)
```

#### bytes()
Returns buffer contents as an array of integers (0-255).

```xxl
var buf = bytesbuffer.fromString("Hi")
var arr = buf.bytes()  // [72, 105]
```

#### toString()
Returns buffer contents as a string.

```xxl
var buf = bytesbuffer.create()
buf.write("Hello")
pln(buf.toString())  // "Hello"
```

#### len()
Returns the current length of the buffer.

```xxl
var buf = bytesbuffer.fromString("Hello")
pln(buf.len())  // 5
```

#### cap()
Returns the current capacity of the buffer.

```xxl
var buf = bytesbuffer.create(100)
pln(buf.cap())  // >= 100
```

#### readByte()
Reads and returns a single byte, or null if buffer is empty.

```xxl
var buf = bytesbuffer.fromString("Hi")
var b1 = buf.readByte()  // 72 ('H')
var b2 = buf.readByte()  // 105 ('i')
var b3 = buf.readByte()  // null (empty)
```

#### readInt16(), readInt32(), readInt64()
Reads integers in little-endian format. Returns null on error.

```xxl
var buf = bytesbuffer.create()
buf.writeInt32(12345)
var n = buf.readInt32()  // 12345
```

#### readFloat32(), readFloat64()
Reads floats in little-endian format. Returns null on error.

```xxl
var buf = bytesbuffer.create()
buf.writeFloat64(3.14159)
var f = buf.readFloat64()  // 3.14159
```

#### peek(n)
Returns the next n bytes without advancing the read position.

```xxl
var buf = bytesbuffer.fromString("Hello")
var preview = buf.peek(3)  // [72, 101, 108] ("Hel")
pln(buf.len())  // 5 (unchanged)
```

#### clear()
Clears all content from the buffer.

```xxl
var buf = bytesbuffer.fromString("Hello")
buf.clear()
pln(buf.len())  // 0
```

#### reset()
Alias for `clear()`. Resets the buffer to empty state.

#### grow(n)
Pre-allocates buffer capacity for better performance.

```xxl
var buf = bytesbuffer.create()
buf.grow(1024)  // Pre-allocate 1KB
```

#### truncate(n)
Discards all but the first n bytes.

```xxl
var buf = bytesbuffer.fromString("Hello World")
buf.truncate(5)
pln(buf.toString())  // "Hello"
```

#### isEmpty()
Returns true if the buffer is empty.

```xxl
var buf = bytesbuffer.create()
pln(buf.isEmpty())  // true
buf.write("test")
pln(buf.isEmpty())  // false
```

### Binary Protocol Example

```xxl
import "bytesbuffer"

// Build a simple binary message
var buf = bytesbuffer.create()

// Message header: magic number (4 bytes) + version (2 bytes) + length (4 bytes)
buf.writeInt32(0x4D455353)  // "MESS" magic number
buf.writeInt16(1)           // version 1
buf.writeInt32(12)          // payload length

// Payload
buf.write("Hello World!")

pln("Total bytes:", buf.len())  // 22 bytes
```

### Thread Safety Example

```xxl
import "bytesbuffer"
load "sync"

var buf = bytesbuffer.create()
var mu = sync.createMutex()

// Safe concurrent writes
spawn {
    mu.lock()
    buf.write("thread1")
    mu.unlock()
}

spawn {
    mu.lock()
    buf.write("thread2")
    mu.unlock()
}
```

---

## chars

The `chars` type provides proper Unicode character handling, where operations work on characters (code points) rather than bytes. This is essential for correctly processing text containing non-ASCII characters like Chinese, Japanese, Korean, emoji, and other Unicode characters.

### Why chars?

In Xxlang, the `string` type is byte-oriented (like Go), which means:
- `len("中文")` returns 6 (bytes), not 2 (characters)
- `"中文"[0]` returns a byte value, not a character
- `substr` uses byte indices

The `chars` type provides character-oriented operations:
- `len(toChars("中文"))` returns 2 (characters)
- `toChars("中文")[0]` returns "中" (full character)
- `subStr` uses character indices

### toChars(s)

Converts a string to a chars array for character-based operations.

```xxl
var s = "Hello世界🎉"
var c = toChars(s)

pln(len(s))      // 15 (bytes)
pln(len(c))      // 8 (characters)
```

### charLen(s)

Returns the number of Unicode characters in a string (without creating a chars object).

```xxl
charLen("Hello世界🎉")  // 8
charLen("中文测试")     // 4
charLen("hello")        // 5
```

### Chars Indexing

Access characters by their character index (not byte index):

```xxl
var c = toChars("Hello世界🎉")

pln(c[0])   // "H"
pln(c[5])   // "世"
pln(c[7])   // "🎉"
pln(c[-1])  // "🎉" (negative index from end)
```

### Chars Slicing

Extract substrings using character indices:

```xxl
var c = toChars("Hello世界🎉")

pln(c.subStr(0, 5).toStr())   // "Hello"
pln(c.subStr(5, 7).toStr())   // "世界"
```

### Chars Methods

#### toStr()

Converts chars back to a string.

```xxl
var c = toChars("Hello")
pln(c.toStr())  // "Hello"
```

#### upper()

Returns uppercase version (character-aware).

```xxl
var c = toChars("Hello World 你好")
pln(c.upper().toStr())  // "HELLO WORLD 你好"
```

#### lower()

Returns lowercase version (character-aware).

```xxl
var c = toChars("HELLO WORLD 你好")
pln(c.lower().toStr())  // "hello world 你好"
```

#### contains(substring)

Checks if chars contains a substring (character-aware).

```xxl
var c = toChars("Hello World 你好")
pln(c.contains("World"))  // true
pln(c.contains("你好"))    // true
pln(c.contains("xyz"))    // false
```

#### indexOf(substring)

Returns the character index of the first occurrence, or -1 if not found.

```xxl
var c = toChars("Hello World 你好")
pln(c.indexOf("World"))  // 6
pln(c.indexOf("你好"))    // 12
pln(c.indexOf("xyz"))    // -1
```

#### startsWith(prefix)

Checks if chars starts with the given prefix.

```xxl
var c = toChars("Hello World")
pln(c.startsWith("Hello"))  // true
pln(c.startsWith("World"))  // false
```

#### endsWith(suffix)

Checks if chars ends with the given suffix.

```xxl
var c = toChars("Hello World 你好")
pln(c.endsWith("你好"))    // true
pln(c.endsWith("World"))  // false
```

#### reverse()

Returns a reversed copy of the chars.

```xxl
var c = toChars("abc世")
pln(c.reverse().toStr())  // "世cba"
```

#### repeat(n)

Returns chars repeated n times.

```xxl
var c = toChars("abc世")
pln(c.repeat(3).toStr())  // "abc世abc世abc世"
```

### String vs Chars Comparison

| Operation | `string` (bytes) | `chars` (characters) |
|----------|------------------|---------------------|
| `len("中文")` | 6 (bytes) | N/A |
| `len(toChars("中文"))` | N/A | 2 (characters) |
| `"中文"[0]` | Byte value | N/A |
| `toChars("中文")[0]` | N/A | "中" (character) |
| `substr(s, 0, 1)` | Byte slice | N/A |
| `c.subStr(0, 1)` | N/A | Character slice |

### When to Use chars

Use `chars` when you need to:
- Count characters in Unicode text
- Extract or manipulate individual Unicode characters
- Perform character-based slicing
- Handle text containing Chinese, Japanese, Korean, emoji, etc.

Use `string` when you need to:
- Work with byte-oriented APIs
- Optimize for ASCII-only text
- Maintain compatibility with Go's string model

### Example: Processing Multilingual Text

```xxl
var text = "日本語English中文한국어"
var c = toChars(text)

pln("Text: ", text)
pln("Byte count: ", len(text))       // 20 bytes
pln("Character count: ", len(c))      // 13 characters

// Iterate over each character
for (var i = 0; i < len(c); i = i + 1) {
    pln("  [", i, "] = ", c[i])
}
```

---

## math

Mathematical functions.

#### abs(x)
Returns absolute value.

```xxl
abs(-42)    // 42
abs(-3.14)  // 3.14
```

#### floor(x)
Returns floor (round down).

```xxl
floor(3.7)   // 3
floor(-3.7)  // -4
```

#### ceil(x)
Returns ceiling (round up).

```xxl
ceil(3.2)    // 4
ceil(-3.2)   // -3
```

#### round(x)
Returns nearest integer.

```xxl
round(3.4)   // 3
round(3.6)   // 4
```

#### sqrt(x)
Returns square root.

```xxl
sqrt(16)   // 4
sqrt(2)    // 1.414...
```

#### pow(base, exp)
Returns base raised to power.

```xxl
pow(2, 8)   // 256
pow(10, 3)  // 1000
```

#### min(a, b)
Returns smaller value.

```xxl
min(3, 7)   // 3
```

#### max(a, b)
Returns larger value.

```xxl
max(3, 7)   // 7
```

#### sin(x), cos(x), tan(x)
Trigonometric functions (radians).

```xxl
sin(0)              // 0
cos(0)              // 1
tan(0)              // 0
```

#### log(x), log10(x)
Natural and base-10 logarithm.

```xxl
log(2.718281828)   // ~1
log10(100)         // 2
```

#### random()
Returns random float between 0 and 1.

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
Returns random integer in range.

```xxl
var die = randomInt(1, 6)
```

---

## array

Array manipulation utilities.

#### len(arr)
Returns array length.

```xxl
len([1, 2, 3])  // 3
```

#### first(arr)
Returns first element, or null if empty.

```xxl
first([1, 2, 3])  // 1
first([])         // null
```

#### last(arr)
Returns last element, or null if empty.

```xxl
last([1, 2, 3])  // 3
last([])         // null
```

#### push(arr, value)
Returns new array with value appended.

```xxl
push([1, 2], 3)  // [1, 2, 3]
```

#### pop(arr)
Returns new array without last element.

```xxl
pop([1, 2, 3])  // [1, 2]
```

#### sort(arr)
Returns sorted array (ascending).

```xxl
sort([3, 1, 4, 1, 5])  // [1, 1, 3, 4, 5]
```

#### reverse(arr)
Returns reversed array.

```xxl
reverse([1, 2, 3])  // [3, 2, 1]
```

#### sum(arr)
Returns sum of numeric elements.

```xxl
sum([1, 2, 3, 4, 5])  // 15
```

#### avg(arr)
Returns average of numeric elements.

```xxl
avg([1, 2, 3, 4, 5])  // 3
```

#### indexOf(arr, value)
Returns index of value, or -1.

```xxl
indexOf([1, 2, 3], 2)  // 1
indexOf([1, 2, 3], 5)  // -1
```

#### containsArr(arr, value)
Returns true if value is in array.

```xxl
containsArr([1, 2, 3], 2)  // true
```

#### concat(arr1, arr2)
Concatenates two arrays.

```xxl
concat([1, 2], [3, 4])  // [1, 2, 3, 4]
```

#### slice(arr, start, end)
Returns subarray.

```xxl
slice([1, 2, 3, 4, 5], 1, 4)  // [2, 3, 4]
```

#### isEmpty(arr)
Returns true if array is empty.

```xxl
isEmpty([])      // true
isEmpty([1])     // false
```

---

## json

JSON encoding, decoding, and JSONPath query operations.

### Basic Functions

#### parse(jsonString)
Parses JSON string to Xxlang value. Also available as builtin `fromJson`.

```xxl
var data = json.parse('{"name": "Alice", "age": 30}')
pln(data["name"])  // "Alice"

var arr = json.parse('[1, 2, 3]')
pln(arr[0])  // 1
```

#### stringify(value, indent)
Converts Xxlang value to JSON string. Also available as builtin `toJson`.

```xxl
var obj = {"name": "Alice", "age": 30}
pln(json.stringify(obj))        // {"name":"Alice","age":30}
pln(json.stringify(obj, "  "))  // Pretty printed with 2-space indent
pln(json.stringify(obj, 4))     // Pretty printed with 4-space indent
```

#### encode(value)
Alias for stringify without indent.

```xxl
json.encode({"a": 1})  // '{"a":1}'
```

#### decode(jsonString)
Alias for parse.

```xxl
json.decode('{"x": 10}')["x"]  // 10
```

#### toJson(value, options...)
Converts Xxlang value to JSON string with options. Matches builtin `toJson` behavior.

```xxl
json.toJson(obj, "-indent")  // Pretty print
json.toJson(obj, "-sort")    // Sort keys
```

#### fromJson(jsonString)
Alias for parse. Matches builtin `fromJson` behavior.

```xxl
json.fromJson('{"x": 10}')["x"]  // 10
```

### File Operations

#### readFile(path)
Reads and parses a JSON file.

```xxl
var config = json.readFile("config.json")
io.println(config["server"])
```

#### writeFile(path, obj, indent)
Writes object to JSON file.

```xxl
var data = {"name": "test", "values": [1, 2, 3]}
json.writeFile("output.json", data)
json.writeFile("output.json", data, "  ")  // With indent
```

#### writeFilePretty(path, obj, indent)
Writes formatted JSON to file.

```xxl
json.writeFilePretty("output.json", data, "  ")
```

#### updateFile(path, updates)
Updates a JSON file with new values.

```xxl
json.updateFile("config.json", {"debug": true, "timeout": 60})
```

#### appendToArrayFile(path, element)
Appends an element to a JSON array file.

```xxl
json.appendToArrayFile("logs.json", {"timestamp": time.unix(), "msg": "error"})
```

### Utility Functions

#### isValid(jsonString)
Returns true if the string is valid JSON.

```xxl
json.isValid('{"a": 1}')   // true
json.isValid('{invalid}')  // false
```

#### getType(jsonString)
Returns the type of JSON value: "object", "array", "string", "number", "boolean", "null", or "invalid".

```xxl
json.getType('{"a": 1}')  // "object"
json.getType('[1, 2, 3]') // "array"
json.getType('42')        // "number"
```

### JSONPath Operations

JSONPath is a query language for JSON, similar to XPath for XML. It allows you to select and extract data from JSON documents.

#### JSONPath Syntax

| Syntax | Description | Example |
|--------|-------------|---------|
| `$` | Root object | `$` |
| `.field` | Access field | `$.store.name` |
| `[n]` | Array index | `$.books[0]` |
| `[-n]` | Negative index (from end) | `$.books[-1]` |
| `[*]` | Wildcard (all elements) | `$.books[*]` |
| `..` | Recursive descent | `$..author` |
| `[start:end]` | Array slice | `$[1:5]` |
| `[a,b,c]` | Multiple indices | `$[0,2,4]` |
| `[?(expr)]` | Filter expression | `$[?(@.price < 10)]` |

#### Filter Expression Operators

Filter expressions support the following operators:

| Operator | Description | Example |
|----------|-------------|---------|
| `==`, `!=` | Equality comparison | `@.name == "Alice"` |
| `<`, `>`, `<=`, `>=` | Numeric comparison | `@.price < 10` |
| `&&` | Logical AND | `@.price > 5 && @.price < 20` |
| `\|\|` | Logical OR | `@.active \|\| @.pending` |
| `!` | Logical NOT | `!@.disabled` |
| `=~` | Regex match | `@.name =~ "^[A-Z]"` |
| `in` | Value in array | `@.category in ["fiction", "drama"]` |
| `nin` | Value not in array | `@.category nin ["fiction"]` |
| `contains` | String/array contains | `@.name contains "a"` |
| `startsWith` | String starts with | `@.name startsWith "The"` |
| `endsWith` | String ends with | `@.name endsWith "ing"` |
| `between` | Value in range (inclusive) | `@.price between [10, 100]` |
| `isNull` | Value is null | `@.email isNull` |
| `isNotNull` | Value is not null | `@.email isNotNull` |
| `isType` | Check value type | `@.age isType "number"` |
| `absent` | Field does not exist | `@.optional absent` |
| `empty()` | Check if empty | `empty(@.items)` |
| `length()` | Get length | `length(@.name) > 3` |

**Supported types for `isType`:** `number`, `int`, `float`, `string`, `boolean`, `array`, `object`, `null`

```xxl
// Filter examples
var books = json.getAll("$.store.book[?(@.price between [10, 20])]", data)
var withEmail = json.getAll("$.users[?(@.email isNotNull)]", data)
var numbers = json.getAll("$.items[?(@.value isType \"number\")]", data)
var missing = json.getAll("$.users[?(@.phone absent)]", data)
```

#### get(path, obj)
Gets the first value matching a JSONPath. Returns null if not found.

```xxl
var data = json.parse('{"store": {"book": [{"title": "Book1"}, {"title": "Book2"}]}}')

var title = json.get("$.store.book[0].title", data)  // "Book1"
var lastTitle = json.get("$.store.book[-1].title", data)  // "Book2"
```

#### getAll(path, obj)
Gets all values matching a JSONPath as an array.

```xxl
var data = json.parse('{"store": {"book": [{"title": "A"}, {"title": "B"}]}}')

var titles = json.getAll("$..title", data)  // ["A", "B"]
var books = json.getAll("$.store.book[*]", data)  // All books
```

#### getWithPath(path, obj)
Gets all values matching a JSONPath with their paths as a map.

```xxl
var data = json.parse('{"a": {"b": 1, "c": 2}}')
var result = json.getWithPath("$..*", data)
// {"$.a": {"b": 1, "c": 2}, "$.a.b": 1, "$.a.c": 2}
```

#### set(path, obj, value)
Sets a value at the specified JSONPath. Returns a new object (does not mutate original).

```xxl
var data = {"name": "Alice", "age": 30}
var newData = json.set("$.age", data, 31)
// newData = {"name": "Alice", "age": 31}

// Original unchanged
pln(data.age)  // 30
```

#### delete(path, obj)
Deletes values at the specified JSONPath. Returns a new object.

```xxl
var data = {"name": "Alice", "age": 30, "city": "NYC"}
var newData = json.delete("$.city", data)
// newData = {"name": "Alice", "age": 30}
```

#### has(path, obj)
Returns true if at least one value matches the JSONPath.

```xxl
var data = {"store": {"name": "MyStore"}}
json.has("$.store.name", data)  // true
json.has("$.store.owner", data)  // false
```

#### count(path, obj)
Returns the number of values matching the JSONPath.

```xxl
var data = json.parse('{"items": [1, 2, 3, 4, 5]}')
json.count("$.items[*]", data)  // 5
```

#### paths(obj)
Returns all JSONPath strings that exist in an object.

```xxl
var data = {"a": {"b": 1, "c": 2}}
var allPaths = json.paths(data)
// ["$.a", "$.a.b", "$.a.c"]
```

#### query(path, jsonString)
Combines parse and get in one operation.

```xxl
var title = json.query("$.store.book[0].title", '{"store": {"book": [{"title": "Test"}]}}')
// "Test"
```

#### queryAll(path, jsonString)
Combines parse and getAll in one operation.

```xxl
var titles = json.queryAll("$..title", '{"store": {"book": [{"title": "A"}, {"title": "B"}]}}')
// ["A", "B"]
```

### JSONPath Examples

```xxl
import "json"

var jsonStr = `
{
  "store": {
    "book": [
      {"category": "fiction", "title": "Book A", "price": 10},
      {"category": "fiction", "title": "Book B", "price": 15},
      {"category": "non-fiction", "title": "Book C", "price": 25}
    ],
    "name": "My Bookstore"
  }
}
`

var data = json.parse(jsonStr)

// Get store name
var storeName = json.get("$.store.name", data)  // "My Bookstore"

// Get first book title
var firstTitle = json.get("$.store.book[0].title", data)  // "Book A"

// Get last book
var lastBook = json.get("$.store.book[-1]", data)

// Get all book titles (recursive)
var titles = json.getAll("$..title", data)  // ["Book A", "Book B", "Book C"]

// Get all prices
var prices = json.getAll("$.store.book[*].price", data)  // [10, 15, 25]

// Filter books with price < 20
var cheapBooks = json.getAll("$.store.book[?(@.price < 20)]", data)

// Slice: first two books
var firstTwo = json.getAll("$.store.book[0:2]", data)

// Get multiple indices
var selected = json.getAll("$.store.book[0,2]", data)  // Books at index 0 and 2

// Check if path exists
json.has("$.store.book[0].title", data)  // true

// Count books
json.count("$.store.book[*]", data)  // 3

// Set a new price
var updated = json.set("$.store.book[0].price", data, 12.99)

// Delete a book
var fewer = json.delete("$.store.book[1]", data)
```

---

## xml

XML encoding, decoding, and manipulation utilities.

### Core Functions

#### parse(xmlStr)
Parses an XML string and returns an Xxlang object (map).

```xxl
import "xml"

var xmlStr = `<book id="123">
    <title>Learning Xxlang</title>
    <author>John Doe</author>
</book>`

var data = xml.parse(xmlStr)
// data is: {"book": {"@attributes": {"id": "123"}, "title": {"@text": "Learning Xxlang"}, ...}}
```

#### stringify(rootName, obj, indent?)
Converts an Xxlang object to an XML string.

```xxl
var obj = {
    "@attributes": {"version": "1.0"},
    "@text": "Some content",
    "name": {"@text": "Test"}
}
var xmlStr = xml.stringify("root", obj, 2)  // 2-space indent
```

#### encode(rootName, obj)
Converts an Xxlang object to a compact XML string (no indentation).

```xxl
var xmlStr = xml.encode("root", {"@text": "Hello"})
// <root>Hello</root>
```

#### decode(xmlStr)
Alias for `parse()`. Parses an XML string.

### File Operations

#### readFile(path)
Reads an XML file and parses it.

```xxl
var data = xml.readFile("config.xml")
pln(data["config"]["setting"]["@text"])
```

#### writeFile(path, rootName, obj, indent?)
Writes an Xxlang object to an XML file with XML declaration.

```xxl
xml.writeFile("output.xml", "config", data, 2)
```

### Element Extraction

#### getAttr(element, attrName)
Gets an attribute value from an XML element.

```xxl
var book = xml.getElement(data, "book")
var id = xml.getAttr(book, "id")  // "123"
```

#### getText(element)
Gets the text content from an XML element.

```xxl
var title = xml.getElement(book, "title")
var text = xml.getText(title)  // "Learning Xxlang"
```

#### getElement(element, name)
Gets a child element by name. Returns null if not found.

```xxl
var author = xml.getElement(book, "author")
```

#### getElements(element, name)
Gets all child elements with a given name (for repeated elements). Always returns an array.

```xxl
var chapters = xml.getElements(book, "chapter")
for (ch in chapters) {
    pln(xml.getText(ch))
}
```

### Element Modification

#### setAttr(element, name, value)
Sets an attribute on an XML element. Returns a new element with the attribute.

```xxl
var updated = xml.setAttr(book, "lang", "en")
```

#### setText(element, text)
Sets the text content of an XML element. Returns a new element.

```xxl
var updated = xml.setText(title, "New Title")
```

#### addElement(parent, name, child)
Adds a child element to an XML element. Returns a new element.

```xxl
var updated = xml.addElement(book, "publisher", {"@text": "Tech Books"})
```

#### newElement(name, text?, attributes?)
Creates a new XML element.

```xxl
var elem = xml.newElement("item", "Hello", {"type": "greeting"})
```

### Utilities

#### isValid(xmlStr)
Checks if a string is valid XML. Returns boolean.

```xxl
if (xml.isValid(xmlStr)) {
    var data = xml.parse(xmlStr)
}
```

#### escape(str)
Escapes special XML characters (<, >, &, ", ').

```xxl
var escaped = xml.escape("<tag>text & more</tag>")
// "&lt;tag&gt;text &amp; more&lt;/tag&gt;"
```

#### unescape(str)
Unescapes XML entities.

```xxl
var text = xml.unescape("&lt;tag&gt;")
// "<tag>"
```

### XML Structure

Parsed XML is represented as maps with special keys:

- `@attributes` - Map of element attributes
- `@text` - Text content of the element
- `@declaration` - XML declaration (if present)
- Child elements are stored by their tag names as keys

```xxl
// XML: <book id="123" lang="en"><title>Hello</title></book>
// Becomes:
{
    "book": {
        "@attributes": {"id": "123", "lang": "en"},
        "title": {"@text": "Hello"}
    }
}
```

---

## csv

CSV file reading and writing operations.

> **See [File Operations Guide](FILE.md) for complete CSV documentation.**

### Quick Start

```xxl
import * as csv from "csv"

// Read CSV file
var data = csv.read("data.csv")

// Read with header row (returns array of maps)
var records = csv.readWithHeader("users.csv")
io.println(records[0]["name"])

// Write CSV file
var rows = [
    ["name", "age"],
    ["Alice", "30"],
    ["Bob", "25"]
]
csv.write("output.csv", rows)
```

### Key Functions

| Function | Description |
|----------|-------------|
| `read(path)` | Read CSV as array of arrays |
| `readWithHeader(path)` | Read CSV with header as array of maps |
| `write(path, data)` | Write array of arrays to CSV |
| `writeWithHeader(path, data, headers)` | Write array of maps with header |
| `parse(str)` | Parse CSV string |
| `stringify(data)` | Convert to CSV string |

---

## regex

Regular expression operations.

#### match(pattern, str)
Returns true if string matches pattern.

```xxl
match("\\d+", "123")       // true
match("[a-z]+", "Hello")   // false
```

#### find(pattern, str)
Returns first match or null.

```xxl
find("\\d+", "abc123def")  // "123"
```

#### findAll(pattern, str)
Returns array of all matches.

```xxl
findAll("\\d+", "a1b22c333")  // ["1", "22", "333"]
```

#### replace(pattern, str, replacement)
Replaces all matches.

```xxl
replace("\\d+", "a1b2c3", "X")  // "aXbXcX"
```

#### split(pattern, str)
Splits string by pattern.

```xxl
split("\\s+", "a  b   c")  // ["a", "b", "c"]
```

---

## crypto

Cryptographic functions.

#### md5(s)
Returns MD5 hash as hex string.

```xxl
md5("hello")  // "5d41402abc4b2a76b9719d911017c592"
```

#### sha1(s)
Returns SHA1 hash as hex string.

```xxl
sha1("hello")  // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
```

#### sha256(s)
Returns SHA256 hash as hex string.

```xxl
sha256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
```

#### sha512(s)
Returns SHA512 hash as hex string.

```xxl
sha512("hello")  // "9b71d224bd62f3785d96d46ad3ea3d..."
```

#### hmacMd5(key, data)
Returns HMAC-MD5 hash.

```xxl
hmacMd5("secret", "message")  // Hex string
```

#### hmacSha1(key, data)
Returns HMAC-SHA1 hash.

```xxl
hmacSha1("secret", "message")  // Hex string
```

#### hmacSha256(key, data)
Returns HMAC-SHA256 hash.

```xxl
hmacSha256("secret", "message")  // Hex string
```

#### hmacSha512(key, data)
Returns HMAC-SHA512 hash.

```xxl
hmacSha512("secret", "message")  // Hex string
```

#### base64Encode(s)
Encodes string to base64.

```xxl
base64Encode("hello")  // "aGVsbG8="
```

#### base64Decode(s)
Decodes base64 string.

```xxl
base64Decode("aGVsbG8=")  // "hello"
```

#### base64UrlEncode(s)
URL-safe base64 encoding.

```xxl
base64UrlEncode("hello?world")
```

#### base64UrlDecode(s)
URL-safe base64 decoding.

```xxl
base64UrlDecode(base64UrlEncode("hello?world"))  // "hello?world"
```

#### hexEncode(bytes)
Encodes to hex string.

```xxl
hexEncode("hello")  // "68656c6c6f"
```

#### hexDecode(hex)
Decodes hex string.

```xxl
hexDecode("68656c6c6f")  // "hello"
```

#### randomBytes(n)
Returns n random bytes as string.

```xxl
var bytes = randomBytes(16)
```

#### randomHex(n)
Returns n random bytes as hex string.

```xxl
randomHex(16)  // "a1b2c3d4..." (32 hex chars)
```

#### randomBase64(n)
Returns n random bytes as base64 string.

```xxl
randomBase64(16)  // Random base64 encoded string
```

#### uuid()
Generates a random UUID (v4) string.

```xxl
uuid()  // "550e8400-e29b-41d4-a716-446655440000"
```

---

## time

Time and date functions for working with timestamps and durations.

### Timestamps

#### unix()
Returns current Unix timestamp in seconds.

```xxl
import "time"
var ts = time.unix()  // e.g., 1710422400
```

#### unixMs()
Returns current Unix timestamp in milliseconds.

```xxl
var ms = time.unixMs()  // e.g., 1710422400000
```

#### unixNano()
Returns current Unix timestamp in nanoseconds.

```xxl
var ns = time.unixNano()  // e.g., 1710422400000000000
```

### Date/Time Components

#### now()
Returns current time as a map with components.

```xxl
var t = time.now()
pln(t["year"])     // 2026
pln(t["month"])    // 3 (March)
pln(t["day"])      // 14
pln(t["hour"])     // 20
pln(t["minute"])   // 30
pln(t["second"])   // 45
pln(t["nanosecond"])  // 123456789
```

#### year(), month(), day(), hour(), minute(), second()
Returns current date/time component.

```xxl
pln(time.year())    // 2026
pln(time.month())   // 3
pln(time.day())     // 14
pln(time.hour())    // 20
pln(time.minute())  // 30
pln(time.second())  // 45
```

#### weekday()
Returns current weekday (0=Sunday, 1=Monday, ..., 6=Saturday).

```xxl
var wd = time.weekday()  // e.g., 4 (Thursday)
```

### Formatting

#### format(layout)
Formats current time using Go-style layout.

```xxl
pln(time.format("2006-01-02"))           // "2026-03-14"
pln(time.format("15:04:05"))             // "20:30:45"
pln(time.format("2006-01-02 15:04:05"))  // "2026-03-14 20:30:45"
```

#### formatUnix(timestamp, layout)
Formats a Unix timestamp.

```xxl
var ts = time.unix()
pln(time.formatUnix(ts, "2006-01-02"))
```

#### parse(layout, value)
Parses a time string and returns Unix timestamp.

```xxl
var ts = time.parse("2006-01-02", "2026-03-14")
```

### Sleep

#### sleep(ms)
Pauses execution for specified milliseconds.

```xxl
pln("Starting...")
time.sleep(1000)  // Sleep 1 second
pln("Done!")
```

#### sleepSec(seconds)
Pauses execution for specified seconds.

```xxl
time.sleepSec(2)  // Sleep 2 seconds
```

### Date Arithmetic

#### addDays(days)
Adds days to current time, returns Unix timestamp.

```xxl
var tomorrow = time.addDays(1)
var nextWeek = time.addDays(7)
```

#### addMonths(months)
Adds months to current time.

```xxl
var nextMonth = time.addMonths(1)
```

#### addYears(years)
Adds years to current time.

```xxl
var nextYear = time.addYears(1)
```

#### since(startMs)
Returns milliseconds elapsed since given timestamp.

```xxl
var start = time.unixMs()
// ... do some work ...
var elapsed = time.since(start)
pln("Took " + elapsed.toStr() + " ms")
```

### Calendar

#### isLeapYear(year)
Returns true if year is a leap year.

```xxl
time.isLeapYear(2024)  // true
time.isLeapYear(2023)  // false
```

#### daysInMonth(year, month)
Returns number of days in month.

```xxl
time.daysInMonth(2024, 2)  // 29 (leap year)
time.daysInMonth(2023, 2)  // 28
time.daysInMonth(2024, 1)  // 31
```

---

## Built-in Functions (Global)

These functions are available without importing any module.

### Type Functions

#### typeOf(value)
Returns type name as string.

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
```

#### len(value)
Returns length of string, array, or map.

```xxl
len("hello")      // 5
len([1, 2, 3])    // 3
len({"a": 1})     // 1
```

### Type Conversion

#### int(value)
Converts to integer.

```xxl
int(3.14)      // 3
int("42")      // 42
int(true)      // 1
```

#### float(value)
Converts to float.

```xxl
float(42)      // 42.0
float("3.14")  // 3.14
```

#### string(value)
Converts to string.

```xxl
string(42)     // "42"
string(3.14)   // "3.14"
string(true)   // "true"
```

### Map Operations

#### keys(map)
Returns array of map keys.

```xxl
keys({"a": 1, "b": 2})  // ["a", "b"]
```

#### values(map)
Returns array of map values.

```xxl
values({"a": 1, "b": 2})  // [1, 2]
```

#### hasKey(map, key)
Returns true if map has key.

```xxl
hasKey({"a": 1}, "a")   // true
hasKey({"a": 1}, "b")   // false
```

#### delete(map, key)
Returns map with key removed.

```xxl
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

### Range

#### range(start, stop)
Returns array of integers.

```xxl
range(0, 5)    // [0, 1, 2, 3, 4]
range(1, 4)    // [1, 2, 3]
```

### Assertions

#### assert(condition)
Throws error if condition is false.

```xxl
assert(1 + 1 == 2)
assert(x > 0, "x must be positive")
```

### String Utilities

#### repeat(s, n)
Returns string repeated n times.

```xxl
repeat("ab", 3)  // "ababab"
repeat("x", 5)   // "xxxxx"
```

#### lpad(s, length, padChar)
Pads string on the left to specified length.

```xxl
lpad("5", 4, "0")     // "0005"
lpad("hello", 10)      // "     hello"
```

#### rpad(s, length, padChar)
Pads string on the right to specified length.

```xxl
rpad("5", 4, "0")      // "5000"
rpad("hello", 10)      // "hello     "
```

#### charAt(s, index)
Returns character at index, or null if out of bounds.

```xxl
charAt("hello", 1)  // "e"
charAt("hello", 10) // null
```

#### trimLeft(s, cutset)
Removes leading characters.

```xxl
trimLeft("  hello")          // "hello"
trimLeft("xxhelloxx", "x")   // "helloxx"
```

#### trimRight(s, cutset)
Removes trailing characters.

```xxl
trimRight("hello  ")         // "hello"
trimRight("xxhelloxx", "x")  // "xxhello"
```

### Type Checking

#### isEmpty(value)
Returns true if value is empty string, array, map, or null.

```xxl
isEmpty("")          // true
isEmpty([])          // true
isEmpty({})          // true
isEmpty(null)        // true
isEmpty([1, 2])      // false
```

#### isString(value), isNumber(value), isInt(value), isFloat(value)
Returns true if value is of the specified type.

```xxl
isString("hello")    // true
isNumber(42)         // true
isNumber(3.14)       // true
isInt(42)            // true
isFloat(3.14)        // true
```

#### isArray(value), isMap(value), isBool(value), isNull(value), isFunction(value)
Returns true if value is of the specified type.

```xxl
isArray([1, 2, 3])     // true
isMap({"a": 1})         // true
isBool(true)            // true
isNull(null)            // true
isFunction(len)         // true
```

### Math Utilities

#### round(value, precision)
Rounds to nearest integer or specified decimal places.

```xxl
round(3.7)           // 4
round(3.14159, 2)    // 3.14
round(3.14159, 4)    // 3.1416
```

#### clamp(value, min, max)
Clamps value to range [min, max].

```xxl
clamp(15, 0, 10)     // 10
clamp(-5, 0, 10)     // 0
clamp(5, 0, 10)      // 5
```

#### sign(value)
Returns sign of number (-1, 0, or 1).

```xxl
sign(-42)    // -1
sign(0)      // 0
sign(42)     // 1
```

#### random()
Returns random float between 0 and 1.

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
Returns random integer in range [min, max].

```xxl
var die = randomInt(1, 6)   // 1, 2, 3, 4, 5, or 6
```

### Array Utilities

#### unique(arr)
Returns array with duplicate values removed.

```xxl
unique([1, 2, 2, 3, 3, 3])  // [1, 2, 3]
```

#### flatten(arr, depth)
Flattens nested arrays.

```xxl
flatten([[1, 2], [3, 4]])           // [1, 2, 3, 4]
flatten([[[1]], [2]], 1)             // [[1], 2]
```

#### without(arr, values...)
Returns array with specified values removed.

```xxl
without([1, 2, 3, 4], 2, 4)  // [1, 3]
```

#### take(arr, n)
Returns first n elements.

```xxl
take([1, 2, 3, 4, 5], 3)  // [1, 2, 3]
```

#### drop(arr, n)
Returns array without first n elements.

```xxl
drop([1, 2, 3, 4, 5], 2)  // [3, 4, 5]
```

### Map Utilities

#### merge(map1, map2, ...)
Merges multiple maps (later maps override earlier).

```xxl
merge({"a": 1}, {"b": 2})           // {"a": 1, "b": 2}
merge({"a": 1}, {"a": 2, "b": 3})    // {"a": 2, "b": 3}
```

#### entries(map)
Returns array of [key, value] pairs.

```xxl
entries({"x": 10, "y": 20})  // [["x", 10], ["y", 20]]
```

### Formatting

#### format(template, args...)
Formats string with Go-style format specifiers.

```xxl
format("Hello %s, you are %d", "Alice", 30)  // "Hello Alice, you are 30"
format("Pi is %.2f", 3.14159)                 // "Pi is 3.14"
```

---

## fmt

Formatting utilities.

#### sprintf(format, args...)
Returns formatted string using Go-style format specifiers.

```xxl
import "fmt"
var msg = fmt.sprintf("Name: %s, Age: %d", "Alice", 30)
pln(msg)  // Name: Alice, Age: 30
```

#### printf(format, args...)
Prints formatted string to stdout.

```xxl
fmt.printf("Value: %d\n", 42)
```

---

## encoding

Encoding/decoding utilities (Base64, Hex).

#### base64Encode(s)
Encodes string to base64.

```xxl
import "encoding"
encoding.base64Encode("hello")  // "aGVsbG8="
```

#### base64Decode(s)
Decodes base64 string.

```xxl
encoding.base64Decode("aGVsbG8=")  // "hello"
```

#### hexEncode(s)
Encodes string to hexadecimal.

```xxl
encoding.hexEncode("hello")  // "68656c6c6f"
```

#### hexDecode(s)
Decodes hexadecimal string.

```xxl
encoding.hexDecode("68656c6c6f")  // "hello"
```

---

## uuid

UUID generation.

#### uuid()
Generates a random UUID (v4) string.

```xxl
import "uuid"
var id = uuid.uuid()  // "550e8400-e29b-41d4-a716-446655440000"
```

---

## debug

Debugging utilities.

#### stacktrace()
Returns current call stack as string.

```xxl
import "debug"
pln(debug.stacktrace())
```

#### gc()
Triggers garbage collection.

```xxl
debug.gc()
```

---

## Other Modules

Additional standard library modules available:

| Module | Description |
|--------|-------------|
| `bytes` | Byte array operations |
| `collections` | Collection utilities (sets, stacks, queues) |
| `csv` | CSV file parsing and writing |
| `env` | Environment variable utilities |
| `fp` | Functional programming utilities (map, filter, reduce) |
| `log` | Logging utilities |
| `net` | Network utilities (HTTP client) |
| `sort` | Advanced sorting utilities |
| `strconv` | String conversion utilities |
| `text` | Text processing utilities |
| `validate` | Input validation utilities |

---

### bytes Module

Byte manipulation utilities for low-level operations.

```xxlang
load("bytes")

// Create byte array from string
var b = bytes.fromString("hello")  // [104, 101, 108, 108, 111]

// Convert back to string
var s = bytes.toString(b)  // "hello"

// Get/set byte at index
var byte = bytes.get(b, 0)  // 104
bytes.set(b, 0, 72)  // Change first byte

// Encode/decode integers (big/little endian)
var encoded = bytes.encodeInt64BE(12345)
var decoded = bytes.decodeInt64BE(encoded)

// Other operations
bytes.concat(b1, b2)     // Concatenate byte arrays
bytes.slice(b, 0, 3)     // Slice byte array
bytes.compare(b1, b2)    // Compare (-1, 0, 1)
bytes.equal(b1, b2)      // Check equality
bytes.count(b, 65)       // Count occurrences of byte
bytes.indexOf(b, 65)     // Find byte index (-1 if not found)
```

---

### collections Module

Collection utilities for working with arrays and sets.

```xxlang
load("collections")

// Set operations
var a = [1, 2, 3]
var b = [2, 3, 4]
collections.union(a, b)         // [1, 2, 3, 4]
collections.intersection(a, b)  // [2, 3]
collections.difference(a, b)    // [1]

// Chunk array
collections.chunk([1,2,3,4,5], 2)  // [[1,2], [3,4], [5]]

// Zip arrays
collections.zip([1,2], [3,4], [5,6])  // [[1,3,5], [2,4,6]]

// Flatten deeply nested arrays
collections.flattenDeep([[1,[2]],3])  // [1, 2, 3]

// Group and count
collections.countBy([1,1,2,3])  // [["1", 2], ["2", 1], ["3", 1]]
collections.groupBy([1,2,3,4], fn(x) { x % 2 })

// Partition by predicate
collections.partition([1,2,3,4,5], fn(x) { x > 2 })  // [[3,4,5], [1,2]]

// Take/drop
collections.take([1,2,3,4,5], 3)      // [1, 2, 3]
collections.drop([1,2,3,4,5], 2)      // [3, 4, 5]
collections.takeWhile([1,2,3,4], fn(x) { x < 3 })  // [1, 2]

// Find
collections.find([1,2,3,4], fn(x) { x > 2 })      // 3
collections.findIndex([1,2,3,4], fn(x) { x > 2 }) // 2

// Every/some
collections.every([1,2,3], fn(x) { x > 0 })  // true
collections.some([1,2,3], fn(x) { x > 2 })   // true

// Range with step
collections.rangeStep(0, 10, 2)  // [0, 2, 4, 6, 8]

// Repeat element
collections.repeat("x", 5)  // ["x", "x", "x", "x", "x"]

// Shuffle and sample
collections.shuffle([1,2,3,4,5])
collections.sample([1,2,3,4,5])      // Random element
collections.sample([1,2,3,4,5], 2)   // 2 random elements
```

---

### csv Module

CSV parsing and generation utilities.

```xxlang
load("csv")

// Parse CSV string
var data = csv.parse("a,b,c\n1,2,3\n4,5,6")
// [["a","b","c"], ["1","2","3"], ["4","5","6"]]

// Parse with header (returns array of maps)
var records = csv.parseWithHeader("name,age\nAlice,30\nBob,25")
// [{"name": "Alice", "age": "30"}, {"name": "Bob", "age": "25"}]

// Custom delimiter
csv.parse("a;b;c", ";")

// Generate CSV
csv.stringify([["a","b"], ["1","2"]])  // "a,b\n1,2"

// Generate from maps
csv.stringifyMaps([{"a": 1}], ["a"])  // "a\n1"

// Get column/row
csv.column(data, 0)  // Get first column
csv.row(data, 0)     // Get first row

// Transpose
csv.transpose([[1,2], [3,4]])  // [[1,3], [2,4]]

// Filter/map rows
csv.filterRows(data, fn(row) { row[0] == "1" })
csv.mapRows(data, fn(row) { row })

// Count
csv.rowCount(data)
csv.colCount(data)

// Skip/take
csv.skip(data, 1)   // Skip first row
csv.take(data, 2)   // Take first 2 rows

// Append/prepend row
csv.appendRow(data, ["x", "y"])
csv.prependRow(data, ["header"])
```

---

### env Module

Environment variable and system utilities.

```xxlang
load("env")

// Environment variables
env.get("HOME")            // Get env var
env.getOr("DEBUG", "0")    // Get with default
env.set("MY_VAR", "value") // Set env var
env.unset("MY_VAR")        // Unset env var
env.has("HOME")            // Check if exists
env.all()                  // Get all as array of pairs
env.map()                  // Get all as map
env.path()                 // Get PATH as array
env.expand("$HOME/test")   // Expand env vars in string

// Type-specific getters
env.getInt("PORT", 8080)   // Get as int with default
env.getBool("DEBUG", false) // Get as bool
env.lookup("HOME")         // [exists, value]

// Working directory
env.cwd()                  // Get current directory
env.cd("/tmp")             // Change directory

// Process info
env.pid()                  // Process ID
env.ppid()                 // Parent process ID
env.exe()                  // Executable path
env.exit(0)                // Exit program

// Arguments
env.args()                 // Command line args
env.scriptArgs()           // Args after --
env.mixArgs()              // Script args or all args

// User directories
env.cacheDir()             // User cache directory
env.configDir()            // User config directory

// Other
env.clear()                // Clear all env vars
env.streams()              // [stdin, stdout, stderr available]
```

---

### fp Module

Functional programming utilities.

```xxlang
load("fp")

// Function composition
var double = fn(x) { x * 2 }
var addOne = fn(x) { x + 1 }
var composed = fp.compose(double, addOne)
composed(5)  // (5 + 1) * 2 = 12

var piped = fp.pipe(double, addOne)
piped(5)     // (5 * 2) + 1 = 11

// Utility functions
fp.identity(5)         // 5
fp.constant(10)(x)     // Always returns 10
fp.alwaysTrue()        // true
fp.alwaysFalse()       // false

// Predicate combinators
fp.not(fn(x) { x > 0 })        // Negate predicate
fp.allPass(fn(x) { x > 0 }, fn(x) { x < 10 })
fp.anyPass(fn(x) { x < 0 }, fn(x) { x > 10 })

// Higher-order utilities
fp.tap(fn(x) { pln(x) })(5)   // Execute side effect, return value
fp.defaultTo(0)(null)          // Default if null

// Object utilities
fp.equals(5)(5)                // true
fp.prop("name")({"name": "A"}) // Get property
fp.pick(["a", "b"])({"a": 1, "b": 2, "c": 3})  // {"a": 1, "b": 2}
fp.omit(["c"])({"a": 1, "b": 2, "c": 3})       // {"a": 1, "b": 2}

// Array utilities
fp.concat([1,2], [3,4])  // [1,2,3,4]
fp.flatten([[1,2], [3]]) // [1,2,3]
fp.head([1,2,3])         // 1
fp.tail([1,2,3])         // [2,3]
fp.init([1,2,3])         // [1,2]
fp.last([1,2,3])         // 3
fp.length([1,2,3])       // 3
fp.isEmpty([])           // true

// Range
fp.range(5)              // [0,1,2,3,4]
fp.range(1, 5)           // [1,2,3,4]
fp.range(0, 10, 2)       // [0,2,4,6,8]

// Times
fp.times(3, fn(i) { i * 2 })  // [0, 2, 4]

// Memoize
var memoized = fp.memoize(fn(x) { /* expensive */ })

// Until (iterate until predicate true)
fp.until(fn(x) { x > 10 }, fn(x) { x * 2 }, 1)  // 16
```

---

### log Module

Logging utilities with levels.

```xxlang
load("log")

// Basic logging
log.debug("Debug message")
log.info("Info message")
log.warn("Warning message")
log.error("Error message")
log.fatal("Fatal error")  // Logs and exits

// Set log level
log.setLevel("debug")  // debug, info, warn, error
log.getLevel()         // Current level

// Format without printing
log.format("info", "message")  // "[timestamp] INFO: message"

// Log to file
log.toFile("app.log", "info", "message")

// Simple print
log.print("message")
log.printNoNL("no newline")
log.printf("Value: %d", 42)

// Log with prefix
log.withPrefix("APP", "message")

// JSON format
log.json("info", "message")
// {"timestamp":"...","level":"info","message":"..."}

// Stack trace
log.stack()

// Check if level enabled
log.isLevel("debug")  // true/false
```

---

### net Module

HTTP client utilities.

```xxlang
load("net")

// HTTP GET
var result = net.get("https://api.example.com/data")
// [body, statusCode, status]

// HTTP POST
var result = net.post("https://api.example.com/api", '{"key":"value"}')
var result = net.post(url, body, "application/json")

// Generic request
net.request("PUT", url, body, {"Authorization": "Bearer token"})

// HEAD request
var result = net.head(url)  // [statusCode, headers]

// Download file content
var content = net.download("https://example.com/file.txt")

// Set timeout (seconds)
net.setTimeout(60)

// Status code helpers
net.isOK(200)           // true (200-299)
net.isRedirect(301)     // true (300-399)
net.isClientError(404)  // true (400-499)
net.isServerError(500)  // true (500+)

// JSON helpers
net.getJson("https://api.example.com/data")
net.postJson("https://api.example.com/api", '{"key":"value"}')
```

---

### sort Module

Sorting utilities for arrays.

```xxlang
load("sort")

// Sort numbers
sort.numbers([3, 1, 4, 1, 5])       // [1, 1, 3, 4, 5]
sort.numbersDesc([3, 1, 4, 1, 5])   // [5, 4, 3, 1, 1]

// Sort strings
sort.strings(["c", "a", "b"])       // ["a", "b", "c"]
sort.stringsDesc(["c", "a", "b"])   // ["c", "b", "a"]

// Sort by key function
sort.by([{name: "Bob"}, {name: "Alice"}], fn(x) { x.name })

// Reverse array
sort.reverse([1, 2, 3, 4])  // [4, 3, 2, 1]

// Check if sorted
sort.isSorted([1, 2, 3])    // true

// Min/max
sort.min([3, 1, 2])         // 1
sort.max([3, 1, 2])         // 3
sort.minIndex([3, 1, 2])    // 1
sort.maxIndex([3, 1, 2])    // 0
```

---

### strconv Module

String conversion utilities.

```xxlang
load("strconv")

// Parse functions
strconv.parseInt("42")         // 42
strconv.parseInt("ff", 16)     // 255 (hex)
strconv.parseFloat("3.14")     // 3.14
strconv.parseBool("true")      // true

// Format functions
strconv.formatInt(42)          // "42"
strconv.formatInt(255, 16)     // "ff"
strconv.formatFloat(3.14)      // "3.14"
strconv.formatFloat(3.14, 2)   // "3.14" with precision
strconv.formatBool(true)       // "true"

// Quote/unquote
strconv.quote("hello\nworld")  // "\"hello\\nworld\""
strconv.unquote("\"hello\"")   // "hello"

// Type conversions
strconv.toString(42)           // "42"
strconv.toString(3.14)         // "3.14"
strconv.toInt("42")            // 42
strconv.toInt(3.14)            // 3
strconv.toFloat("3.14")        // 3.14
strconv.toFloat(42)            // 42.0
strconv.toBool("true")         // true
strconv.toBool(1)              // true

// JSON
strconv.toJSON(obj)
strconv.toJSONPretty(obj)

// Formatting helpers
strconv.formatNumber(1234.567, 2)  // "1234.57"
strconv.formatBytes(1536)          // "1.50 KB"
strconv.formatDuration(65000)      // "1m 5s"
```

---

### text Module

Text processing utilities.

```xxlang
load("text")

// Word wrap
text.wordWrap("Hello world this is a test", 10)
// "Hello\nworld this\nis a test"

// Truncate
text.truncate("Hello world", 8)        // "Hello..."
text.truncate("Hello world", 8, "…")   // "Hello w…"

// Count
text.wordCount("Hello world")   // 2
text.lineCount("Line1\nLine2")  // 2
text.charCount("中文")          // 2 (characters/runes)
text.byteCount("中文")          // 6 (bytes)

// Split/join
text.lines("a\nb\nc")           // ["a", "b", "c"]
text.joinLines(["a", "b"])      // "a\nb"
text.words("hello world")       // ["hello", "world"]
text.chars("abc")               // ["a", "b", "c"]

// Case conversion
text.title("hello world")       // "Hello World"
text.capitalize("hello")        // "Hello"
text.swapCase("Hello")          // "hELLO"

// Type checks
text.isAlphaNum("abc123")       // true
text.isAlpha("abc")             // true
text.isNumeric("123")           // true
text.isSpace("   ")             // true
text.isBlank("   ")             // true

// Whitespace
text.removeSpaces("a b c")      // "abc"
text.normalizeSpace("a   b")    // "a b"

// Padding
text.padLeft("5", 3, "0")       // "005"
text.padRight("5", 3, "0")      // "500"

// Indentation
text.indent("a\nb", "  ")       // "  a\n  b"
text.dedent("  a\n  b")         // "a\nb"
text.centerText("hi", 10)       // "    hi    "

// Repeat
text.repeat("ab", 3)            // "ababab"

// Character utilities
text.charAt("hello", 1)         // "e"
text.charCode("A", 0)           // 65
text.fromCode(65)               // "A"

// Escaping
text.shellEscape("hello'world") // "'hello'\"'\"'world'"
text.jsonEscape("hello\n")      // "hello\\n"
text.jsonUnescape("hello\\n")   // "hello\n"
```

---

### validate Module

Input validation utilities.

```xxlang
load("validate")

// Format validation
validate.isEmail("user@example.com")  // true
validate.isURL("https://example.com") // true

// Regex matching
validate.matches("hello123", "^[a-z]+[0-9]+$")  // true

// String validation
validate.lengthRange("hello", 1, 10)  // true
validate.required("  hello  ")        // true (not empty after trim)

// Array membership
validate.inArray("a", ["a", "b", "c"])    // true
validate.notInArray("x", ["a", "b", "c"]) // true

// Numeric range
validate.inRange(5, 1, 10)   // true

// Type checks
validate.isJSON('{"a":1}')        // true
validate.isAlphanumeric("abc123") // true
validate.isAlpha("abc")           // true
validate.isNumeric("123.45")      // true
validate.isInteger("123")         // true

// Format checks
validate.isHexColor("#ff0000")    // true
validate.isUUID("550e8400-e29b-41d4-a716-446655440000")  // true
validate.isIPv4("192.168.1.1")    // true
validate.isPhone("+1-555-123-4567")  // true
validate.isDate("2024-01-15")     // true
validate.isTime("14:30:00")       // true

// String operations
validate.startsWith("hello", "he")  // true
validate.endsWith("hello", "lo")    // true
validate.contains("hello", "ell")   // true

// Credit card (Luhn algorithm)
validate.isCreditCard("4111111111111111")  // true
```

---

### queue Module

A FIFO (First-In-First-Out) queue with O(1) push and pop operations.

```xxlang
load("queue")

// Create an empty queue
var q = queue.create()

// Create with initial capacity
var q2 = queue.create(100)

// Create from array
var q3 = queue.fromArray([1, 2, 3])

// Check if object is a queue
queue.isQueue(q)  // true

// Queue methods
q.push(1)           // Add to back
q.push(2)
q.push(3)
q.peek()            // 1 (front element, doesn't remove)
q.peekBack()        // 3 (back element, doesn't remove)
q.pop()             // 1 (removes from front)
q.len()             // 2
q.isEmpty()         // false
q.toArray()         // [2, 3]
q.clear()           // Remove all elements
q.clone()           // Create a copy
```

---

### set Module

An unordered collection of unique elements with set operations.

```xxlang
load("set")

// Create an empty set
var s = set.create()

// Create with initial elements
var s1 = set.create(1, 2, 3)

// Create with initial capacity
var s2 = set.create(100)

// Create from array
var s3 = set.fromArray([1, 2, 2, 3])  // {1, 2, 3}

// Check if object is a set
set.isSet(s1)  // true

// Basic operations
s1.add(4)           // true (added)
s1.add(3)           // false (already exists)
s1.contains(2)      // true
s1.remove(2)        // true
s1.remove(99)       // false (not found)
s1.len()            // 3
s1.isEmpty()        // false
s1.toArray()        // [1, 3, 4] (order not guaranteed)
s1.toSortedArray()  // [1, 3, 4] (sorted by Inspect())
s1.clear()          // Remove all elements
s1.clone()          // Create a copy

// Set operations
var a = set.create(1, 2, 3)
var b = set.create(2, 3, 4)

a.union(b)              // {1, 2, 3, 4}
a.intersect(b)          // {2, 3}
a.difference(b)         // {1}
a.symmetricDiff(b)      // {1, 4}

// Also available as module functions
set.union(a, b)
set.intersect(a, b)
set.difference(a, b)
set.symmetricDiff(a, b)

// Set comparisons
var c = set.create(1, 2)
var d = set.create(1, 2, 3)
var e = set.create(1, 2, 3)

c.isSubset(d)      // true
d.isSuperset(c)    // true
d.equals(e)        // true
```

---

### xlsx Module

Excel file handling for reading and writing xlsx files. Supports both sheet name and 1-based sheet index for all operations.

#### Module Functions

| Function | Description |
|----------|-------------|
| `xlsx.create()` | Create a new empty workbook |
| `xlsx.open(path)` | Open an existing xlsx file |
| `xlsx.isXLSX(obj)` | Check if object is an XLSX workbook |
| `xlsx.colToIndex(col)` | Convert column letter to index (A=1, AA=27) |
| `xlsx.indexToCol(idx)` | Convert index to column letter (1=A, 27=AA) |
| `xlsx.parseCellRef(ref)` | Parse cell reference, returns [col, row] |

#### Workbook Methods

| Method | Description |
|--------|-------------|
| `wb.getSheetList()` | Get array of sheet names |
| `wb.getSheetCount()` | Get number of sheets |
| `wb.getSheetName(idx)` | Get sheet name by 1-based index |
| `wb.newSheet(name)` | Create new sheet, returns true on success |
| `wb.deleteSheet(sheet)` | Delete sheet by name or index |
| `wb.save(path)` | Save workbook to file |
| `wb.close()` | Close workbook |

#### Cell Operations

All sheet parameters can be either sheet name (string) or sheet index (int, 1-based).

| Method | Description |
|--------|-------------|
| `wb.getCell(sheet, ref)` | Get cell value by reference (e.g., "A1") |
| `wb.getCell(sheet, row, col)` | Get cell value by row/col (1-based) |
| `wb.setCell(sheet, ref, value)` | Set cell value by reference |
| `wb.setCell(sheet, row, col, value)` | Set cell value by row/col |

#### Row/Column Operations

| Method | Description |
|--------|-------------|
| `wb.getRow(sheet, row)` | Get row as array |
| `wb.setRow(sheet, row, values)` | Set row from array |
| `wb.getCol(sheet, col)` | Get column as array |
| `wb.getRange(sheet, range)` | Get range as 2D array (e.g., "A1:C3") |
| `wb.getRowCount(sheet)` | Get number of rows |
| `wb.getColCount(sheet)` | Get number of columns |
| `wb.insertRow(sheet, row)` | Insert row at position |
| `wb.deleteRow(sheet, row)` | Delete row |
| `wb.insertCol(sheet, col)` | Insert column at position |
| `wb.deleteCol(sheet, col)` | Delete column |

#### Merge Operations

| Method | Description |
|--------|-------------|
| `wb.mergeCell(sheet, start, end)` | Merge cell range |
| `wb.unmergeCell(sheet, ref)` | Unmerge cells |
| `wb.getMerges(sheet)` | Get array of merge ranges |

#### Image Operations

| Method | Description |
|--------|-------------|
| `wb.getImages(sheet)` | Get array of image info maps |
| `wb.extractImage(sheet, idx, path)` | Extract image to file |
| `wb.getImageData(sheet, idx)` | Get image data as base64 string |

#### Example Usage

```xxlang
load("xlsx")

// Create a new workbook
var wb = xlsx.create()

// Open an existing file
var wb = xlsx.open("data.xlsx")

// Check if object is an XLSX workbook
xlsx.isXLSX(wb)  // true

// ========== Workbook Operations ==========

wb.getSheetList()                // ["Sheet1", "Sheet2"]
wb.getSheetCount()               // Number of sheets
wb.getSheetName(1)               // Get sheet name by 1-based index
wb.newSheet("Data")              // Create new sheet

// Delete sheet - supports both name and index
wb.deleteSheet("Sheet1")         // Delete by name
wb.deleteSheet(1)                // Delete by index

wb.save("output.xlsx")           // Save to file
wb.close()                       // Close workbook

// ========== Sheet Index Support ==========

// All methods that take a sheet name also accept a 1-based index:
wb.getCell(1, "A1")              // Get cell from sheet 1 by index
wb.setCell(2, "A1", "Hello")     // Set cell in sheet 2 by index
wb.getRow(1, 1)                  // Get row from sheet 1
wb.getRowCount(1)                // Get row count of sheet 1

// ========== Cell Operations ==========

// Read cells - sheet can be name or index
wb.getCell("Sheet1", "A1")       // By reference
wb.getCell("Sheet1", 1, 1)       // By row/col (1-based)
wb.getCell(1, "A1")              // Using sheet index

// Write cells
wb.setCell("Sheet1", "A1", "Hello")    // String
wb.setCell("Sheet1", "B1", 123)        // Number
wb.setCell("Sheet1", "C1", true)       // Boolean
wb.setCell("Sheet1", 2, 1, "World")    // By row/col

// ========== Row/Column Operations ==========

// Read row/column - sheet can be name or index
wb.getRow("Sheet1", 1)           // [A1, B1, C1, ...]
wb.getRow(1, 1)                  // From sheet 1 by index
wb.getCol("Sheet1", 1)           // [A1, A2, A3, ...]
wb.getRange("Sheet1", "A1:C3")   // 2D array

// Write row
wb.setRow("Sheet1", 1, ["Name", "Age", "City"])

// Dimensions
wb.getRowCount("Sheet1")         // Number of rows
wb.getColCount("Sheet1")         // Number of columns

// Insert/delete
wb.insertRow("Sheet1", 3)        // Insert row at position 3
wb.deleteRow("Sheet1", 5)        // Delete row 5
wb.insertCol("Sheet1", 2)        // Insert column at position 2
wb.deleteCol("Sheet1", 3)        // Delete column 3

// ========== Merge Cells ==========

wb.mergeCell("Sheet1", "A1", "C1")   // Merge A1:C1
wb.unmergeCell("Sheet1", "A1")       // Unmerge
wb.getMerges("Sheet1")               // ["A1:C1", ...]

// ========== Images ==========

// Get image information
var images = wb.getImages("Sheet1")
// Returns: [{col: 1, row: 1, colEnd: 3, rowEnd: 5, filename: "xl/media/image1.png"}, ...]

// Extract image to file
wb.extractImage("Sheet1", 0, "output.png")

// Get image data as base64
var data = wb.getImageData("Sheet1", 0)

// ========== Utility Functions ==========

xlsx.colToIndex("A")       // 1
xlsx.colToIndex("AA")      // 27
xlsx.indexToCol(1)         // "A"
xlsx.indexToCol(27)        // "AA"
xlsx.parseCellRef("A1")    // ["A", 1]
```

