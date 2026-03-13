# Xxlang Language Design Document

**Date**: 2026-03-07
**Version**: 1.0
**Status**: Approved

## Overview

Xxlang (Chinese: 现象语言) is a line-by-line interpreted scripting language implemented in Go. It features a bytecode virtual machine for execution, lightweight OOP, and comprehensive standard library.

## Design Decisions Summary

| Aspect | Decision |
|--------|----------|
| Syntax | C-like (braces, familiar to most developers) |
| Data Types | Extended: int, float, string, bool, array, map, null, function, bytes |
| Control Structures | Extended: if/else, while, for, for-in, switch/case, try/catch, break, continue, return |
| Functions | Full closure support |
| Modules | Simple import statement |
| Plugins | Go shared libraries (.so/.dll) |
| Error Handling | Exceptions (try/catch/finally) |
| OOP | Single inheritance, all types inherit from Object |
| Standard Library | Full utilities (I/O, string, math, networking, JSON/XML, regex, database, crypto, compression) |
| Execution Model | Bytecode VM |

## Architecture

### Layered Architecture

```
┌─────────────────────────────────────────────────┐
│                    cmd/xxlang                    │
│              (CLI entry point)                   │
├─────────────────────────────────────────────────┤
│                   pkg/xxlang                     │
│  ┌─────────┬─────────┬──────────┬────────────┐ │
│  │  lexer  │ parser  │ compiler │     vm     │ │
│  │         │         │          │            │ │
│  │ tokens  │   AST   │ bytecode │ execution  │ │
│  └─────────┴─────────┴──────────┴────────────┘ │
│  ┌─────────┬─────────┬──────────┬────────────┐ │
│  │ objects │ runtime │  stdlib  │  plugins   │ │
│  │         │         │          │            │ │
│  │  types  │ context │ builtins │   loader   │ │
│  └─────────┴─────────┴──────────┴────────────┘ │
└─────────────────────────────────────────────────┘
```

## Package Structure

```
xxlang/
├── cmd/
│   └── xxlang/
│       └── main.go              # CLI entry point
├── pkg/
│   ├── lexer/
│   │   ├── lexer.go             # Lexer implementation
│   │   └── token.go             # Token types
│   ├── parser/
│   │   ├── parser.go            # Parser implementation
│   │   └── ast.go               # AST node definitions
│   ├── compiler/
│   │   ├── compiler.go          # Bytecode compiler
│   │   └── opcode.go            # Opcode definitions
│   ├── vm/
│   │   ├── vm.go                # Virtual machine
│   │   ├── frame.go             # Call frames
│   │   └── stack.go             # Operand stack
│   ├── objects/
│   │   ├── object.go            # Base Object interface
│   │   ├── int.go               # Integer type
│   │   ├── float.go             # Float type
│   │   ├── string.go            # String type
│   │   ├── bool.go              # Boolean type
│   │   ├── array.go             # Array type
│   │   ├── map.go               # Map type
│   │   ├── function.go          # Function type
│   │   ├── bytes.go             # Bytes type
│   │   ├── null.go              # Null type
│   │   └── class.go             # Class/Instance types
│   ├── runtime/
│   │   ├── runtime.go           # Runtime environment
│   │   ├── scope.go             # Variable scopes
│   │   └── module.go            # Module loader
│   ├── stdlib/
│   │   ├── stdlib.go            # Stdlib registry
│   │   ├── io.go                # I/O functions
│   │   ├── strings.go           # String functions
│   │   ├── math.go              # Math functions
│   │   ├── arrays.go            # Array functions
│   │   ├── maps.go              # Map functions
│   │   ├── files.go             # File operations
│   │   ├── net.go               # Networking
│   │   ├── json.go              # JSON parsing
│   │   ├── xml.go               # XML parsing
│   │   ├── regex.go             # Regular expressions
│   │   ├── db.go                # Database access
│   │   ├── crypto.go            # Cryptography
│   │   └── compress.go          # Compression
│   └── plugin/
│       └── loader.go            # Plugin loader
├── go.mod
├── go.sum
├── CLAUDE.md
└── README.md
```

## Object System

### Base Interface

```go
type Object interface {
    Type() ObjectType
    Inspect() string          // String representation
    ToBool() *Bool            // Truthiness for conditions

    // Universal methods (from base Object class)
    TypeOf() *String          // Returns type name
    ToStr() *String           // Convert to string
    Equals(other Object) *Bool // Equality check
    Hash() HashKey            // For use as map keys
}
```

### Type Hierarchy

| Type | Go Representation | Special Methods |
|------|-------------------|-----------------|
| int | `*Int { Value int64 }` | `parseStr()`, `toFloat()`, `abs()`, `min()`, `max()` |
| float | `*Float { Value float64 }` | `parseStr()`, `toInt()`, `round()`, `floor()`, `ceil()` |
| string | `*String { Value string }` | `len()`, `split()`, `join()`, `trim()`, `upper()`, `lower()`, `contains()`, `indexOf()`, `substring()` |
| bool | `*Bool { Value bool }` | `toStr()` |
| array | `*Array { Elements []Object }` | `len()`, `push()`, `pop()`, `shift()`, `slice()`, `indexOf()`, `join()`, `map()`, `filter()`, `reduce()` |
| map | `*Map { Pairs map[HashKey]*MapPair }` | `len()`, `keys()`, `values()`, `has()`, `remove()`, `merge()` |
| function | `*Function { Params, Body, Env }` | `call()`, `bind()` |
| bytes | `*Bytes { Value []byte }` | `len()`, `toStr()`, `fromStr()` |
| null | `*Null{}` | Singleton |
| class | `*Class { Name, Parent, Methods }` | `new()`, `extends()` |
| instance | `*Instance { Class, Fields }` | Field/method access |

## Lexer

### Token Types

- **Literals**: INT, FLOAT, STRING, IDENT
- **Operators**: +, -, *, /, %, =, ==, !=, <, >, <=, >=, &&, ||, !
- **Delimiters**: (, ), {, }, [, ], ,, :, ;, .
- **Keywords**: var, const, func, return, if, else, while, for, in, break, continue, switch, case, default, try, catch, finally, throw, class, extends, new, this, null, true, false, import, export

### Features
- Line and column tracking for error messages
- Single-line (`//`) and multi-line (`/* */`) comments
- String literals with escape sequences
- Number literals (integers and floats with scientific notation)

## Parser

### AST Node Types

**Statements**:
- VarStatement, ConstStatement, ReturnStatement
- ExpressionStatement, BlockStatement
- IfStatement, WhileStatement, ForStatement, ForClassicStatement
- SwitchStatement, CaseClause
- TryStatement, ThrowStatement
- ClassStatement, ImportStatement, ExportStatement

**Expressions**:
- Identifier, IntegerLiteral, FloatLiteral, StringLiteral, BooleanLiteral, NullLiteral
- ArrayLiteral, MapLiteral
- PrefixExpression, InfixExpression
- CallExpression, IndexExpression, DotExpression
- AssignmentExpression, FunctionLiteral
- NewExpression, ThisExpression

### Operator Precedence (highest to lowest)

1. `.` (member access), `()` (call), `[]` (index)
2. `!`, `-` (prefix)
3. `*`, `/`, `%`
4. `+`, `-`
5. `<`, `>`, `<=`, `>=`
6. `==`, `!=`
7. `&&`
8. `||`

## Bytecode Compiler

### Opcodes

- **Stack**: OpConstant, OpPop, OpDup
- **Arithmetic**: OpAdd, OpSub, OpMul, OpDiv, OpMod, OpNeg
- **Comparison**: OpEqual, OpNotEqual, OpLess, OpGreater, OpLessEqual, OpGreaterEqual
- **Logical**: OpAnd, OpOr, OpNot
- **Variables**: OpGetVar, OpSetVar, OpDefineVar, OpGetGlobal, OpSetGlobal
- **Scope**: OpPushScope, OpPopScope
- **Control Flow**: OpJump, OpJumpIfFalse, OpJumpIfTrue, OpCall, OpReturn
- **Collections**: OpArray, OpMap, OpIndex, OpSetIndex
- **Methods**: OpGetMethod, OpCallMethod
- **Classes**: OpNew, OpGetField, OpSetField
- **Modules**: OpImport, OpLoadPlugin
- **Exceptions**: OpThrow, OpPushHandler, OpPopHandler
- **Built-in**: OpBuiltin
- **Null**: OpNull

## Virtual Machine

### Components

- **Stack**: Operand stack for expression evaluation
- **Frames**: Call frames for function calls
- **Globals**: Global variable storage
- **Handlers**: Exception handler stack
- **Modules**: Loaded module cache

### Execution Model

Stack-based execution with call frames. Each function call creates a new frame with its own instruction pointer and local variables.

## Embedding API

```go
// Execute runs Xxlang code and returns the result
func Execute(code string, opts *Options) *ExecuteResult

// ExecuteWithArgs runs code with arguments passed as global variables
func ExecuteWithArgs(code string, args map[string]interface{}, opts *Options) *ExecuteResult

// NewVM creates a new VM instance for reuse
func NewVM(opts *Options) *VM

// VM methods
func (vm *VM) Run(code string) *ExecuteResult
func (vm *VM) RunWithArgs(code string, args map[string]interface{}) *ExecuteResult
func (vm *VM) SetGlobal(name string, value interface{}) error
func (vm *VM) GetGlobal(name string) (interface{}, error)
func (vm *VM) Reset()
```

## CLI

```
xxlang [options] [file] [args...]

Options:
  -c, --compile    Compile to standalone executable
  -o, --output     Output file name (for compilation)
  -e, --eval       Evaluate code from string
  -i, --interactive  Start REPL
  -v, --version    Show version
  -h, --help       Show help
  --module-path    Add module search path
  --debug          Enable debug output
  --timeout        Execution timeout
```

## Standard Library

### Core Built-ins
- Type operations: typeOf, toStr, toInt, toFloat, toBool
- I/O: print, println, input, readFile, writeFile, appendFile
- Collections: len, push, pop, shift, keys, values, has, delete, range
- Execution: runCode, exit

### String Methods
- len, upper, lower, trim, split, contains, indexOf, replace, substring, startsWith, endsWith, repeat, parseStr

### Math Module
- abs, ceil, floor, round, max, min, pow, sqrt, log, exp, sin, cos, tan, random, pi, e

### Array Methods
- len, push, pop, shift, unshift, indexOf, includes, slice, splice, concat, join, reverse, sort, map, filter, reduce, find, every, some, flat

### Map Methods
- len, keys, values, has, get, set, delete, clear, forEach, merge

### File System Module
- readFile, readBytes, writeFile, appendFile, delete, exists, rename, copy
- mkdir, rmdir, listDir, walk, isDir, isFile
- join, dir, base, ext, abs, size, modTime

### JSON Module
- parse, stringify

### Regex Module
- match, find, replace, split, compile

### HTTP Module
- Client: get, post, request
- Server: createServer, listen, close

### Database Module
- open, close, exec, query, queryOne, begin, commit, rollback

### Crypto Module
- md5, sha1, sha256, sha512, hmacSha256
- base64Encode, base64Decode, hexEncode, hexDecode
- randomBytes

### Compression Module
- gzip, gunzip, zip, unzip

## Module System

### Import Syntax

```xxl
import "utils/math.xxl";
import "utils/math.xxl" as m;
import { add, subtract } from "utils/math.xxl";
export func add(a, b) { return a + b; }
export const PI = 3.14159;
export default func main() { }
```

### Module Loading
- Search paths configurable
- Cache loaded modules
- Support .xxl files and index.xxl in directories

## Plugin System

### Plugin Interface (Go)

```go
var XxlangModule = plugin.ModuleInfo{
    Name:     "myplugin",
    Version:  "1.0.0",
    Functions: []string{"hello", "add"},
}

func Hello(args []objects.Object) (objects.Object, error) { }
func Add(args []objects.Object) (objects.Object, error) { }
```

### Plugin Usage (Xxlang)

```xxl
var myplugin = loadPlugin("myplugin");
myplugin.hello();
myplugin.add(1, 2);
```

## Executable Compilation

1. Compile source to bytecode
2. Serialize bytecode
3. Generate Go wrapper with embedded bytecode
4. Build with `go build` for target platform

### Cross-compilation Targets
- linux/amd64, linux/arm64
- darwin/amd64, darwin/arm64
- windows/amd64, windows/386

## Implementation Phases

### Phase 1: Core Language (MVP)
- Lexer, Parser (basic), Objects (basic), Compiler (basic), VM (basic), CLI

### Phase 2: Extended Types
- array, map, function, bytes; collection operations

### Phase 3: Control Flow & OOP
- switch/case, for-in, try/catch; classes, inheritance

### Phase 4: Modules & Plugins
- Module loader, plugin loader, caching

### Phase 5: Standard Library (Core)
- I/O, math, string functions

### Phase 6: Standard Library (Extended)
- File system, JSON, regex, HTTP, database, crypto, compression

### Phase 7: Compilation & Distribution
- Bytecode serialization, executable packaging, cross-compilation, REPL

### Phase 8: Optimization & Benchmarking
- Performance profiling, Fibonacci benchmark, optimization, documentation

## Requirements Checklist

- [x] Use Go standard library as much as possible
- [x] Documentation and code comments in English
- [x] Repository: github/topxeq/xxlang
- [x] Modern programming language features
- [x] Comprehensive documentation (to be written)
- [x] camelCase naming for built-in functions
- [x] Embedding support with parameter passing
- [x] Executable compilation (VM + code packaging)
- [x] runCode function for nested execution
- [x] Lightweight OOP with base class
- [x] Module loading during execution
- [x] Plugin system for extensions
- [ ] Fibonacci(35) benchmark vs Go/C/Java/Python (Phase 8)
