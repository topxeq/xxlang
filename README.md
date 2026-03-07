# Xxlang

Xxlang (Chinese: 现象语言) is a line-by-line interpreted scripting language implemented in Go.

## Features

- Bytecode virtual machine for efficient execution
- Lightweight OOP with single inheritance
- Closures and first-class functions
- Module system for code organization
- Plugin system for extending functionality
- Comprehensive standard library

## Installation

```bash
go install github.com/topxeq/xxlang/cmd/xxlang@latest
```

## Quick Start

```bash
# Run a file
xxlang script.xxl

# Start REPL
xxlang -i

# Evaluate code
xxlang -e "println(1 + 2);"
```

## License

MIT
