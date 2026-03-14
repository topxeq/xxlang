# Xxlang

[中文文档](README_zh.md)

Xxlang (Chinese: 现象语言) is a bytecode VM-based scripting language implemented in Go.

## Features

- **Bytecode VM** - Efficient execution with a stack-based virtual machine
- **Closures** - First-class functions with proper closure support
- **Classes & OOP** - Object-oriented programming with inheritance
- **Module System** - Import/export with standard library
- **Rich Built-ins** - 41+ built-in functions for string, math, array, and map operations
- **REPL** - Interactive REPL with multi-line support and persistent state
- **Embeddable** - Can be used as a library in other Go projects
- **Compilable** - Compile to standalone executable or bytecode

## Documentation

- [Language Reference](docs/LANGUAGE.md) - Complete language syntax and features
- [Standard Library](docs/STDLIB.md) - Built-in modules and functions
- [Embedding Guide](docs/EMBEDDING.md) - Using Xxlang in Go applications
- [Performance Benchmarks](benchmarks/RESULTS.md) - Performance analysis

## Installation

```bash
go install github.com/topxeq/xxlang/cmd/xxlang@latest
```

## Quick Start

```bash
# Run a file
xxlang run script.xxl

# Start interactive REPL
xxlang

# Compile to executable
xxlang compile -o program script.xxl
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
xxlang                        # Start REPL
xxlang run file.xxl           # Run source script
xxlang run file.xxb           # Run compiled bytecode
xxlang compile file.xxl       # Compile to executable wrapper
xxlang compile --bytecode file.xxl    # Compile to bytecode (.xxb)
xxlang compile -o out.xxb --bytecode file.xxl  # Compile with output path
xxlang version                # Show version
xxlang help                   # Show help
```

## Bytecode Compilation

Xxlang supports compiling source code to bytecode files for faster loading and distribution.

### Compile to Bytecode

```bash
# Compile script.xxl to script.xxb
xxlang compile --bytecode script.xxl

# Specify output path
xxlang compile --bytecode -o program.xxb script.xxl
```

### Run Bytecode

```bash
# Execute compiled bytecode
xxlang run script.xxb
```

### Benefits

| Feature | Source (.xxl) | Bytecode (.xxb) |
|---------|---------------|-----------------|
| Loading | Parse + Compile + Execute | Deserialize + Execute |
| Startup | ~5ms overhead | ~1ms overhead |
| Distribution | Source code visible | Obfuscated bytecode |
| Size | Smaller | ~5-10x larger |

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

## Building from Source

```bash
git clone https://github.com/topxeq/xxlang.git
cd xxlang
go build -o xxlang ./cmd/xxlang
```

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

See [benchmarks/FIB35_FINAL_REPORT.md](benchmarks/FIB35_FINAL_REPORT.md) for detailed analysis.

## License

MIT License
