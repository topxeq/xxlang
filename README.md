# Xxlang
![Coverage](https://img.shields.io/badge/Coverage-46.5%25-yellow)

[中文文档](README_zh.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/topxeq/xxlang.svg)](https://pkg.go.dev/github.com/topxeq/xxlang)
[![Go Report Card](https://goreportcard.com/badge/github.com/topxeq/xxlang)](https://goreportcard.com/report/github.com/topxeq/xxlang)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Xxlang (Chinese: 现象语言) is a bytecode VM-based scripting language implemented in Go.

## Features

- **Bytecode VM** - Efficient execution with a stack-based virtual machine
- **Closures** - First-class functions with proper closure support
- **Classes & OOP** - Object-oriented programming with inheritance
- **Module System** - Import/export with standard library
- **Plugin System** - Write native Go plugins for high-performance operations
- **Rich Built-ins** - 41+ built-in functions for string, math, array, and map operations
- **REPL** - Interactive REPL with multi-line support and persistent state
- **Embeddable** - Can be used as a library in other Go projects
- **Compilable** - Compile to standalone executable or bytecode

## Documentation

- [Language Reference](docs/LANGUAGE.md) - Complete language syntax and features
- [Standard Library](docs/STDLIB.md) - Built-in modules and functions
- [Embedding Guide](docs/EMBEDDING.md) - Using Xxlang in Go applications
- [Plugin System](docs/PLUGIN.md) - Writing native Go plugins for high performance
- [Performance Benchmarks](benchmarks/RESULTS.md) - Performance analysis

## Installation

### Option 1: One-line Install (Recommended)

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
# or with wget
wget -qO- https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
```

### Option 2: Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/topxeq/xxlang/releases):

```bash
# Linux (amd64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-linux-amd64.tar.gz
tar -xzf xxlang-linux-amd64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# Linux (arm64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-linux-arm64.tar.gz
tar -xzf xxlang-linux-arm64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# Windows (PowerShell)
# Download from: https://github.com/topxeq/xxlang/releases/latest/download/xxlang-windows-amd64.zip
# Extract xxl.exe and add to PATH

# macOS (amd64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-darwin-amd64.tar.gz
tar -xzf xxlang-darwin-amd64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# macOS (arm64/Apple Silicon)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-darwin-arm64.tar.gz
tar -xzf xxlang-darwin-arm64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/
```

### Option 3: Install via Go

```bash
go install github.com/topxeq/xxlang/cmd/xxl@latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/topxeq/xxlang.git
cd xxlang
go build -o xxl ./cmd/xxl
```

## Updating

Xxlang supports self-updating from GitHub releases:

```bash
xxl update
```

This command will:
1. Check the latest release from GitHub
2. Compare with current version
3. Download the compressed archive for your OS and architecture
4. Extract the binary and replace the current executable

**Note**: On Windows, the old executable may remain as `xxl.exe.old` until next reboot.

## Quick Start

```bash
# Run a local file
xxl run script.xxl

# Run a script from URL
xxl https://raw.githubusercontent.com/user/repo/main/script.xxl

# Start interactive REPL
xxl

# Compile to executable
xxl compile -o program script.xxl
```

## Language Examples

### Variables and Types

```xxl
var x = 10
var y = 3.14
var name = "hello"
var arr = [1, 2, 3, 4, 5]
var map = {"a": 1, "b": 2}
```

### Functions and Closures

```xxl
func add(a, b) {
    return a + b
}

func makeCounter() {
    var count = 0
    func() {
        count = count + 1
        return count
    }
}

var counter = makeCounter()
println(counter())  // 1
println(counter())  // 2
```

### Classes and OOP

```xxl
class Point {
    func init(x, y) {
        this.x = x
        this.y = y
    }

    func add(other) {
        return new Point(this.x + other.x, this.y + other.y)
    }
}

var p1 = new Point(1, 2)
var p2 = new Point(3, 4)
var p3 = p1.add(p2)
```

### Control Flow

```xxl
// If-else
if (x > 0) {
    println("positive")
} else {
    println("non-positive")
}

// For loop
for (var i = 0; i < 5; i = i + 1) {
    println(i)
}

// For-in loop
for (item in [1, 2, 3]) {
    println(item)
}
```

### Modules

```xxl
// Import standard library
import "std/math"
println(math.sqrt(16))

// Import specific functions
import "std/io" { readFile, writeFile }
```

### Plugin System

Write native Go plugins for high-performance operations:

```xxl
// Import a Go plugin
import "plugin/fib"

// Call Go functions from Xxlang
println(fib.fast(50))      // 12586269025
println(fib.matrix(92))    // Largest Fibonacci in int64 range
```

**Two plugin types available:**

| Type | Windows | CGO | Runtime Loading |
|------|---------|-----|-----------------|
| Static Plugin | ✅ | ❌ | ❌ |
| WASM Plugin | ✅ | ❌ | ✅ |

| Method | fib(35) Time | Speedup |
|--------|--------------|---------|
| Xxlang naive recursion | 6.5 seconds | baseline |
| Xxlang tail recursion | 136 µs | 47,000x |
| Go plugin | **35 µs** | **180,000x** |

See [Plugin System](docs/PLUGIN.md) for details.

## Built-in Functions

### String Functions
```xxl
upper("hello")              // "HELLO"
lower("HELLO")              // "hello"
split("a,b,c", ",")         // ["a", "b", "c"]
containsStr("hello", "ell") // true
```

### Math Functions
```xxl
sqrt(16)    // 4
pow(2, 10)  // 1024
abs(-42)    // 42
floor(3.7)  // 3
ceil(3.2)   // 4
```

### Array Functions
```xxl
sort([3, 1, 2])    // [1, 2, 3]
sum([1, 2, 3])      // 6
reverse([1, 2, 3])  // [3, 2, 1]
push([1, 2], 3)     // [1, 2, 3]
```

## CLI Commands

```bash
xxl                           # Start REPL
xxl script.xxl                # Run local script (shortcut)
xxl run script.xxl            # Run source script
xxl run script.xxb            # Run compiled bytecode
xxl https://example.com/script.xxl  # Run script from URL
xxl compile script.xxl        # Compile to executable wrapper
xxl compile --bytecode script.xxl   # Compile to bytecode (.xxb)
xxl compile -o out.xxb --bytecode script.xxl   # Compile with output path
xxl update                    # Self-update to latest version
xxl version                   # Show version
xxl help                      # Show help
```

## Bytecode Compilation

Xxlang supports compiling source code to bytecode files for faster loading and distribution.

### Compile to Bytecode

```bash
# Compile script.xxl to script.xxb
xxl compile --bytecode script.xxl

# Specify output path
xxl compile --bytecode -o program.xxb script.xxl
```

### Run Bytecode

```bash
# Execute compiled bytecode
xxl run script.xxb
```

### Benefits

| Feature | Source (.xxl) | Bytecode (.xxb) |
|---------|---------------|-----------------|
| Loading | Parse + Compile + Execute | Deserialize + Execute |
| Startup | ~5ms overhead | ~1ms overhead |
| Distribution | Source code visible | Obfuscated bytecode |
| Size | Smaller | ~5-10x larger |
| Cross-platform | Yes | **Yes - identical bytecode runs everywhere** |

### Cross-Platform Compatibility

Xxlang bytecode files are **platform-independent**:

- Same `.xxb` file runs on Windows, Linux, macOS
- Supports different CPU architectures (amd64, arm64)
- No recompilation needed when moving between platforms

```bash
# Compile on Windows
xxl compile --bytecode script.xxl

# Copy script.xxb to Linux and run
xxl run script.xxb  # Works without modification!
```

This is possible because:
- Fixed byte order (Big Endian) for version number
- Go's gob encoding (platform-neutral serialization)
- IEEE 754 floating point (standard format)
- UTF-8 strings (universal encoding)
- No embedded file paths or platform-specific data

### Use Cases

- **Development**: Use source files for easy editing
- **Production**: Deploy bytecode for faster startup
- **Distribution**: Share bytecode to hide source code
- **Embedding**: Include bytecode in Go applications

## REPL Commands

| Command | Description |
|---------|-------------|
| `exit`, `quit` | Exit the REPL |
| `help` | Show help message |
| `history` | Show command history |
| `clear` | Clear all variables and functions |

## Running Tests

```bash
go test ./...
```

## Performance

Xxlang uses a bytecode VM with tail call optimization.

### Naive Recursion

| Language | fib(35) Time | Relative to C |
|----------|--------------|---------------|
| C (gcc -O2) | 25 ms | 1x |
| Go | 53 ms | 2.1x |
| Python | 2,714 ms | 107x |
| Xxlang | 6,324 ms | 250x |

### With Tail Call Optimization

| Language | fib(35) TCO | Relative to C |
|----------|-------------|---------------|
| C | ~0.001 ms | 1x |
| Xxlang | 0.015 ms | 15x |

**Key insight**: Algorithm choice matters more than language. Using TCO, Xxlang achieves **420,000x** speedup over naive recursion.

### TCO Rules

Tail Call Optimization applies **automatically** when the function call is the last operation:

| Pattern | TCO | Reason |
|---------|-----|--------|
| `return func(args)` | ✅ Yes | Call is the last operation |
| `return a + func(args)` | ❌ No | Addition needed after call |
| `return func(a) + func(b)` | ❌ No | Addition needed after calls |

```xxl
// ✅ TCO applies - instant execution
func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)  // Direct tail call
}

// ❌ No TCO - exponential time
func fibNaive(n) {
    if (n <= 1) { return n }
    return fibNaive(n - 1) + fibNaive(n - 2)  // Addition after calls
}
```

See [benchmarks/FIB35_FINAL_REPORT.md](benchmarks/FIB35_FINAL_REPORT.md) for detailed analysis.

### High Performance via Go Functions

When embedding Xxlang in Go, you can register Go functions for native performance:

```go
// Register a Go function
interp.SetGlobal("goFib", &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        n := args[0].(*objects.Int).Value
        return &objects.Int{Value: fibFast(n)}  // Go-native, instant!
    },
})

// Call from Xxlang
interp.Eval("goFib(100000)")  // Microseconds, not seconds!
```

| Method | fib(35) | Speedup |
|--------|---------|---------|
| Xxlang naive | 6.5 sec | baseline |
| Xxlang TCO | 200 µs | 32,000x |
| Go function | 25 µs | **260,000x** |

See [docs/EMBEDDING.md](docs/EMBEDDING.md) for complete examples.

## License

MIT License
