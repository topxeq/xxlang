# Xxlang

[中文文档](README_zh.md)

Xxlang is a bytecode VM-based scripting language implemented in Go.

## Features

- **Bytecode VM** - Efficient execution with a stack-based virtual machine
- **Closures** - First-class functions with proper closure support
- **Rich Built-ins** - 41 built-in functions for string, math, array, and map operations
- **REPL** - Interactive REPL with multi-line support and persistent state
- **Embeddable** - Can be used as a library in other Go projects

## Installation

```bash
go install github.com/topxeq/xxlang/cmd/xxlang@latest
```

## Quick Start

```bash
# Run a file
xxlang script.xxl

# Start interactive REPL
xxlang
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

### Control Flow

```xxl
// If-else
if (x > 0) {
    println("positive")
} else {
    println("non-positive")
}

// While loop
var i = 0
while (i < 5) {
    println(i)
    i = i + 1
}

// For loop
for (var j = 0; j < 5; j = j + 1) {
    println(j)
}

// For-in loop
for (item in [1, 2, 3]) {
    println(item)
}
```

### Built-in Functions

```xxl
// String functions
println(upper("hello"))      // HELLO
println(lower("HELLO"))      // hello
println(substr("hello", 1, 4))  // ell
println(split("a,b,c", ","))    // [a, b, c]
println(containsStr("hello", "ell"))  // true

// Math functions
println(sqrt(16))    // 4
println(pow(2, 10))  // 1024
println(abs(-42))    // 42
println(floor(3.7))  // 3
println(ceil(3.2))   // 4

// Array functions
var arr = [3, 1, 4, 1, 5]
println(sort(arr))      // [1, 1, 3, 4, 5]
println(sum(arr))       // 14
println(reverse(arr))   // [5, 1, 4, 1, 3]
println(first(arr))     // 3
println(last(arr))      // 5

// Map functions
var m = {"a": 1, "b": 2}
println(keys(m))        // [a, b]
println(values(m))      // [1, 2]
println(hasKey(m, "a")) // true
```

## REPL Commands

```
exit, quit  - Exit the REPL
help        - Show help message
history     - Show command history
clear       - Clear all variables and functions
```

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

Xxlang uses a bytecode VM with tail call optimization (when possible).

| Language | fib(35) Time |
| Go | 0.056s |
| Python  | 2.77s |
| Xxlang | 9.37s |

## Tail Call Optimization

Xxlang supports tail call optimization for recursive functions, allowing constant stack space usage:

```xxl
func sumTail(n, acc) {
    if (n <= 0) {
        return acc
    }
    return sumTail(n - 1, acc + n)
}

```

## Built-in Functions

Xxlang provides 41 built-in functions for string, math, array, and map operations.

## REPL Commands

```
exit, quit  - Exit the REPL
help        - Show help message
history     - Show command history
clear       - Clear all variables and functions
```

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

## License

MIT License
