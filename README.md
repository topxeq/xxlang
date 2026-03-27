# Xxlang
![Coverage](https://img.shields.io/badge/Coverage-60%25-yellow)

[中文文档](README_zh.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/topxeq/xxlang.svg)](https://pkg.go.dev/github.com/topxeq/xxlang)
[![Go Report Card](https://goreportcard.com/badge/github.com/topxeq/xxlang)](https://goreportcard.com/report/github.com/topxeq/xxlang)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Xxlang (Chinese: 现象语言) is a bytecode VM-based scripting language implemented in Go.

## Features

- **Bytecode VM** - Efficient execution with register-based virtual machine (21% faster than stack-based)
- **JIT Compilation** - Near-native performance with x86-64 JIT compiler supporting Linux, macOS, and Windows (93x faster for recursive algorithms)
- **Method TCO** - Tail call optimization for both functions and methods, enabling efficient recursion
- **Closures** - First-class functions with proper closure support
- **Classes & OOP** - Object-oriented programming with inheritance and multi-level super calls
- **Concurrency** - Go-style goroutines, tubes (channels), select, context for timeout/cancellation, sync primitives
- **Exception Handling** - Full try/catch/finally/throw support
- **Module System** - Import/export with standard library
- **Microservice Mode** - Built-in HTTP/HTTPS server, REST API support, WebSocket
- **Database Support** - SQLite, MySQL, PostgreSQL, Oracle, MSSQL Server (pure Go drivers, no CGO)
- **Plugin System** - WASM plugins for high-performance operations (Windows compatible, no CGO required)
- **Rich Built-ins** - 460+ built-in functions for string, math, array, map, HTTP, concurrency, crypto and more
- **REPL** - Interactive REPL with multi-line support and persistent state
- **Embeddable** - Can be used as a library in other Go projects
- **Compilable** - Compile to standalone executable or cross-platform bytecode
- **Cross-Platform** - Linux, macOS, Windows support for both amd64 and arm64

## Documentation

- [Language Reference](docs/LANGUAGE.md) - Complete language syntax and features
- [Built-in Functions](docs/BUILTINS.md) - Global built-in functions
- [Standard Library](docs/STDLIB.md) - Standard library modules (db, json, math, etc.)
- [Concurrency Programming](docs/CONCURRENCY.md) - Goroutines, tubes (channels), select, context, sync primitives
- [Microservice Mode](docs/MICROSERVICE.md) - HTTP/HTTPS server, REST API, WebSocket
- [Code Examples](docs/EXAMPLES.md) - Common scenarios code examples
- [Variable Scope](docs/SCOPE.md) - Variable scope and closure behavior
- [Embedding Guide](docs/EMBEDDING.md) - Using Xxlang in Go applications
- [JIT Compilation](docs/JIT.md) - JIT compiler documentation
- [Plugin System](docs/PLUGIN.md) - Writing native Go plugins for high performance
- [Performance Benchmarks](tests/benchmark_results.md) - Performance analysis

## Installation

### Option 1: One-line Install (Recommended)

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
```

or with wget:
```bash
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
# Extract xxl.exe (console) and xxlw.exe (GUI, no console window) and add to PATH
# Use xxlw.exe for GUI applications to avoid console window

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
pln(counter())  // 1
pln(counter())  // 2
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
    pln("positive")
} else {
    pln("non-positive")
}

// For loop
for (var i = 0; i < 5; i = i + 1) {
    pln(i)
}

// For-in loop
for (item in [1, 2, 3]) {
    pln(item)
}
```

### Modules

```xxl
// Import standard library
import "math"
pln(math.sqrt(16))

// Import specific functions
import "io" { readFile, writeFile }
```

### Database (db) Module

Xxlang provides built-in database support with pure Go drivers (no CGO required):

**Supported Databases:**
| Database | Driver Name | Notes |
|----------|-------------|-------|
| SQLite | `sqlite` | File-based, no server needed |
| MySQL | `mysql` | TCP connection |
| PostgreSQL | `postgres` | TCP connection |
| Oracle | `oracle` | Oracle DB |
| MSSQL Server | `mssql` | Microsoft SQL Server |

**Basic Usage:**
```xxl
db = import("db")

// Open SQLite database
conn = db.open("sqlite", "test.db")

// Create table
db.exec(conn, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")

// Insert data
db.exec(conn, "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 30)

// Query data
rows = db.query(conn, "SELECT * FROM users WHERE age > ?", 25)
for (row in rows) {
    pln(row["name"], row["age"])
}

// Query single row
row = db.queryRow(conn, "SELECT * FROM users WHERE name = ?", "Alice")
if (row != null) {
    pln(row)
}

// Close connection
db.close(conn)
```

**Transaction Support:**
```xxl
tx = db.begin(conn)
db.txExec(tx, "UPDATE users SET age = age + 1 WHERE id = ?", 1)
db.commit(tx)  // or db.rollback(tx)
```

**Prepared Statements:**
```xxl
stmt = db.prepare(conn, "INSERT INTO users (name, age) VALUES (?, ?)")
db.stmtExec(stmt, "Bob", 25)
db.stmtExec(stmt, "Charlie", 35)
db.close(stmt)
```

**Connection Pool Settings:**
```xxl
db.setMaxOpenConns(conn, 10)
db.setMaxIdleConns(conn, 5)
db.setConnMaxLifetime(conn, 3600)  // seconds
```

**Available Functions:**
| Function | Description |
|----------|-------------|
| `open(driver, dsn)` | Open database connection |
| `openWithoutPing(driver, dsn)` | Open without testing connection |
| `close(db)` | Close connection/rows/stmt |
| `ping(db)` | Test connection |
| `isConnected(db)` | Check connection status |
| `drivers()` | List registered drivers |
| `query(db, sql, args...)` | Query returning array of maps |
| `queryArrays(db, sql, args...)` | Query returning array of arrays |
| `queryRow(db, sql, args...)` | Query single row |
| `exec(db, sql, args...)` | Execute SQL, returns `[lastInsertId, rowsAffected]` |
| `begin(db)` | Start transaction |
| `commit(tx)` | Commit transaction |
| `rollback(tx)` | Rollback transaction |
| `txExec(tx, sql, args...)` | Execute in transaction |
| `txQuery(tx, sql, args...)` | Query in transaction |
| `prepare(db, sql)` | Create prepared statement |
| `stmtExec(stmt, args...)` | Execute prepared statement |
| `stmtQuery(stmt, args...)` | Query prepared statement |
| `stats(db)` | Get connection statistics |
| `columns(rows)` | Get column names |

### Database Built-in Functions

Xxlang provides database built-in functions (no module import needed) in two versions:

1. **String-based (Charlang compatible)**: All values converted to strings
2. **Typed (preserve native types)**: int, float, bool, string preserved

**Basic Usage:**
```xxl
// Connect to database (no import needed)
db = dbConnect("sqlite", "test.db")

// Create table
dbExec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, salary REAL)")

// Insert data with parameters
result = dbExec(db, "INSERT INTO users (name, age, salary) VALUES (?, ?, ?)", "Alice", 30, 50000.50)
id = result[0]         // last inserted ID
affected = result[1]   // rows affected

// String-based query (Charlang compatible - all values are strings)
rows = dbQuery(db, "SELECT * FROM users WHERE age > ?", 25)
pln(rows[0]["age"])      // "30" (string)

// Typed query (preserves native data types)
rowsTyped = dbQueryTyped(db, "SELECT * FROM users WHERE age > ?", 25)
pln(rowsTyped[0]["age"]) // 30 (int)

// Query with header row (first row is column names)
recs = dbQueryRecs(db, "SELECT name, age FROM users")
// recs[0] = ["name", "age"]       // header
// recs[1] = ["Alice", "30"]       // data row

// Group results by a column
byName = dbQueryMap(db, "SELECT * FROM users", "name")
// byName["Alice"] = {id: "1", name: "Alice", age: "30", salary: "50000.5"}

byAge = dbQueryMapArray(db, "SELECT * FROM users", "age")
// byAge["30"] = [{id: "1", name: "Alice", ...}]

// Query single values
count = dbQueryCount(db, "SELECT COUNT(*) FROM users")   // returns int
avgSalary = dbQueryFloat(db, "SELECT AVG(salary) FROM users")  // returns float
name = dbQueryString(db, "SELECT name FROM users WHERE id = 1") // returns string

// Typed versions preserve native types
val = dbQueryValueTyped(db, "SELECT salary FROM users WHERE id = 1") // returns float

// Close connection
dbClose(db)
```

**String-based Functions (Charlang Compatible):**
| Function | Description |
|----------|-------------|
| `formatSQLValue(str)` | Escape special characters for SQL strings |
| `dbConnect(driver, dsn)` | Connect to database, returns DB object |
| `dbClose(db)` | Close database connection |
| `dbQuery(db, sql, args...)` | Query returning array of maps (all values as strings) |
| `dbQueryOrdered(db, sql, args...)` | Query returning ordered array of arrays |
| `dbQueryRecs(db, sql, args...)` | Query returning 2D array with header row (all strings) |
| `dbQueryMap(db, sql, keyCol, args...)` | Query grouped by key column into map (all strings) |
| `dbQueryMapArray(db, sql, keyCol, args...)` | Query grouped by key column into map of arrays |
| `dbQueryCount(db, sql, args...)` | Query returning single integer (for COUNT) |
| `dbQueryFloat(db, sql, args...)` | Query returning single float (for SUM, AVG, etc.) |
| `dbQueryString(db, sql, args...)` | Query returning single string |
| `dbExec(db, sql, args...)` | Execute SQL, returns `[lastInsertId, rowsAffected]` |

**Typed Functions (Preserve Native Types):**
| Function | Description |
|----------|-------------|
| `dbQueryTyped(db, sql, args...)` | Query returning array of maps (types preserved) |
| `dbQueryRowTyped(db, sql, args...)` | Query single row, returns map or `null` |
| `dbQueryArrayTyped(db, sql, args...)` | Query returning 2D array (types preserved) |
| `dbQueryValueTyped(db, sql, args...)` | Query single value (type preserved) |

**Key Differences:**
- String-based functions: NULL → empty string, all values are strings
- Typed functions: NULL → `null`, native types preserved (int, float, bool, string, time)
- All functions work with DB object from db module
- No import required (built-in functions)

### Plugin System

Write native Go plugins for high-performance operations:

```xxl
// Import a Go plugin
import "plugin/fib"

// Call Go functions from Xxlang
pln(fib.fast(50))      // 12586269025
pln(fib.matrix(92))    // Largest Fibonacci in int64 range
```

**Two plugin types available:**

| Type | Windows | CGO | Runtime Loading |
|------|---------|-----|-----------------|
| Static Plugin | ✅ | ❌ | ❌ |
| WASM Plugin | ✅ | ❌ | ✅ |

| Method | fib(35) Time | vs Go |
|--------|--------------|-------|
| Go native recursive | 52 ms | baseline |
| **Xxlang JIT true recursive** | **54 ms** | **1.04x slower** |
| **Xxlang JIT iterative** | **23 ns** | **1.1x slower** |
| Python recursive | 2.7 s | 50x slower |
| Xxlang VM naive recursion | 5.02 s | 97x slower |
| Xxlang VM tail recursion | 739 µs | 14x slower |

See [Plugin System](docs/PLUGIN.md) for details.

## Built-in Functions

For a complete reference of all built-in functions, see [Built-in Functions Reference](docs/BUILTINS.md).

For standard library modules, see [Standard Library Reference](docs/STDLIB.md).

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
xxl                           # Start REPL (or run auto.xxl if exists in current directory)
xxl script.xxl                # Run local script (shortcut)
xxl run script.xxl            # Run source script
xxl run script.xxb            # Run compiled bytecode
xxl https://example.com/script.xxl  # Run script from URL
xxl -cloud basic.xxl          # Run script from configured cloud URL base
xxl compile script.xxl        # Compile to executable wrapper
xxl compile --bytecode script.xxl   # Compile to bytecode (.xxb)
xxl compile -o out.xxb --bytecode script.xxl   # Compile with output path
xxl update                    # Self-update to latest version
xxl version                   # Show version
xxl help                      # Show help
```

### Auto-execution

When `xxl` is run without arguments, it first checks for a file named `auto.xxl` in the current directory. If found, it executes that file instead of starting the REPL. This is useful for:

- **Project automation**: Place build/test scripts in `auto.xxl` and just run `xxl`
- **Configuration scripts**: Auto-configure environment settings
- **Quick tasks**: Common operations that run with a single `xxl` command

```bash
# Create auto.xxl in your project
echo 'pln("Building project...")' > auto.xxl

# Just run xxl to execute it
xxl
# Output: Building project...
```

### Direct Code Execution (-e flag)

Execute Xxlang code directly from the command line without creating a script file:

```bash
# Quick one-liners
xxl -e 'pln("Hello, World!")'
xxl -e '1 + 2 + 3'

# More complex expressions
xxl -e 'for (var i = 1; i <= 5; i++) { pln(i) }'

# Use --eval for clarity
xxl --eval 'var x = 10; var y = 20; pln(x + y)'
```

### Pipe Mode (-pipe flag)

Execute Xxlang code from stdin (standard input):

```bash
# Execute code from pipe
echo 'pln("Hello from pipe!")' | xxl -pipe

# Execute script file content
cat script.xxl | xxl -pipe

# Use with curl to run remote scripts
curl -s https://example.com/script.xxl | xxl -pipe
```

### Piping Data to Scripts

Without `-pipe`, stdin data is passed to the script for processing. Scripts can read stdin using `io.scan()`, `io.scanLine()`, etc.:

```bash
# Pipe JSON data to a formatter script
echo '{"name": "Alice", "age": 30}' | xxl format_json.xxl

# Pipe CSV data to a processor
cat data.csv | xxl process_csv.xxl

# Chain with other commands
ls -la | xxl filter_output.xxl
```

Example script (`format_json.xxl`):
```xxl
import { scan, parseJson, formatJson, pln } from "io"

var input = scan()
var data = parseJson(input)
pln(formatJson(data))
```

### View Mode (-view flag)

View script content without executing:

```bash
# View local script
xxl -view script.xxl

# View remote script
xxl -view https://example.com/script.xxl
```

### VM Selection

Xxlang uses a modern register-based virtual machine:

```bash
xxl script.xxl                  # Run with register VM (default)
```

## Cloud Script Execution

Xxlang supports executing scripts from a configured cloud URL base using the `-cloud` flag.

### Configuration

Create `~/.xxl/settings.json` (or `/.xxl/settings.json` on Linux, `C:\.xxl\settings.json` on Windows):

```json
{
  "cloudUrlBase": "https://script.topget.org/"
}
```

### Usage

```bash
xxl -cloud basic.xxl
```

This fetches and executes `https://script.topget.org/basic.xxl`.

You can also access the config from within scripts:

```xxl
import "os"
var cfg = os.getConfigObj()
pln(cfg["cloudUrlBase"])
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

Xxlang uses a register-based bytecode VM with JIT compilation, achieving near-native performance.

### JIT Performance (Cross-Language Comparison, March 2026)

| Test | C | Java | Go | Python | Xxlang JIT | Xxlang VM |
|------|---|------|-----|--------|------------|-----------|
| fib(35) recursive | 45 ms | 47 ms | 52 ms | 2.7 s | **54 ms** | 5.02 s |
| fib(35) iterative | 19 ns | 21 ns | 21 ns | 837 ns | **23 ns** | 1.5 µs |
| fib(20) recursive | 315 ns | 330 ns | 315 ns | 17.8 µs | **334 ns** | 30 µs |

**Key insights:**
- **JIT matches Go/Java**: Within 4-10% for recursive algorithms
- **JIT vs Python**: 50x faster for recursive Fibonacci
- **JIT vs VM**: 93x faster than bytecode interpreter for recursive calls
- **Pure Go JIT**: No CGO dependencies required

### Register VM Performance

| Benchmark | Register VM (µs) | Go (µs) | Python (µs) |
|-----------|-----------------:|--------:|------------:|
| Fibonacci(15) | 872 | 3.5 | 178 |
| Fibonacci(20) | 4,785 | 39 | 1,980 |
| For Loop(1000) | 474 | ~0.33 | 43.7 |
| Intensive Arithmetic(1000) | 595 | ~0.65 | 139 |
| Function Calls(200) | 526 | ~0.04 | 17.7 |

### Memory Efficiency

| Benchmark | Register VM (allocs) |
|-----------|---------------------:|
| Fibonacci(15) | 173 |
| Fibonacci Iterative | 193 |
| For Loop(1000) | 137 |
| Function Calls | 230 |
| Intensive Arithmetic | 156 |

### fib(35) Benchmark Comparison

| Language | Time | Relative to Go |
|----------|------|----------------|
| **C** | **45 ms** | 0.87x (faster) |
| **Java** | **47 ms** | 0.90x |
| **Go** | **52 ms** | 1x (baseline) |
| **Xxlang JIT** | **54 ms** | 1.04x slower |
| Python | 2,706 ms | 52x slower |
| Xxlang VM | 5,020 ms | 97x slower |

**Key insights:**
- **JIT matches native Go/Java**: Within 4-10% for compute-intensive recursive algorithms
- **JIT is 50x faster than Python**: For recursive algorithms
- **Algorithm choice dominates**: O(n) vs O(2^n) matters more than language

### JIT Features

The JIT compiler includes:

| Feature | Description |
|---------|-------------|
| **Callback Mechanism** | Native code can call back to Go for builtins, function calls, arrays, and objects |
| **Object Handle Pooling** | Efficient memory management with handle reuse |
| **Tail Call Optimization** | Recursive functions compile to loops with O(1) stack |
| **8-Argument Support** | Functions with up to 8 arguments supported |
| **Thread Safety** | `sync.RWMutex` protects all callback operations |

### With Tail Call Optimization

| Language | fib(35) TCO | Relative to Go |
|----------|-------------|----------------|
| **Go (iterative)** | **~0.025 ms** | 1x |
| Xxlang TCO | 0.013 ms | ~0.5x |

**Key insights:**
- TCO makes recursion **400,000x faster** than naive approach
- Xxlang TCO can **match or beat** Go's iterative implementation

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

See [docs/fib35_benchmark_report.md](docs/fib35_benchmark_report.md) for detailed analysis.

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

## Built-in Functions List

Xxlang provides 460+ built-in functions:

| Category | Functions |
|----------|-----------|
| Output | `pln`, `pr`, `pl`, `prf`, `plt` |
| General | `len`, `typeOf`, `toStr`, `copy`, `clone`, `equals`, `defaults` |
| String | `substr`, `split`, `join`, `trim`, `upper`, `lower`, `containsStr`, `replace`, `startsWith`, `endsWith`, `repeat`, `charAt`, `lpad`, `rpad`, `trimLeft`, `trimRight`, `trimPrefix`, `trimSuffix`, `count`, `isDigit`, `isAlpha`, `isAlphaNum` |
| Math | `abs`, `floor`, `ceil`, `sqrt`, `pow`, `min`, `max`, `round`, `clamp`, `sign`, `random`, `randomInt` |
| Type Conversion | `int`, `float`, `string`, `toJson`, `fromJson`, `bytes`, `bigInt`, `bigFloat` |
| Array | `push`, `pop`, `first`, `last`, `rest`, `concat`, `indexOf`, `containsArr`, `sort`, `sum`, `avg`, `reverse`, `unique`, `flatten`, `without`, `take`, `drop`, `find`, `findIndex`, `includes`, `shuffle`, `sample`, `chunk`, `append`, `appendArray`, `arrayContains`, `removeItems` |
| Map | `keys`, `values`, `hasKey`, `delete`, `merge`, `entries` |
| Type Checking | `isEmpty`, `isString`, `isNumber`, `isInt`, `isFloat`, `isArray`, `isMap`, `isBool`, `isFunction`, `isNull`, `isBigInt`, `isBigFloat` |
| Encoding | `base64Encode`, `base64Decode`, `hexEncode`, `hexDecode`, `urlEncode`, `urlDecode` |
| Crypto | `md5`, `sha256`, `encryptText`, `decryptText`, `aesEncrypt`, `aesDecrypt` |
| HTTP Client | `getWeb`, `getWebBytes`, `getWebObject`, `postWeb`, `postWebObject`, `urlExists`, `httpStatus`, `downloadFile` |
| HTTP Server | `writeResp`, `setRespHeader`, `getReqHeader`, `setCookie`, `getCookie`, `parseForm`, `parseJSON`, `getReqBody`, `status`, `redirect`, `serveFile`, `queryParam`, `formValue` |
| WebSocket | `webSocket`, `wsReadMsg`, `wsSendText`, `wsSendBinary`, `wsClose`, `isWebSocket` |
| Concurrency | `makeTube`, `closeTube`, `tubeSend`, `tubeRecv`, `newMutex`, `newRWMutex`, `newWaitGroup`, `newOnce`, `newCond`, `newAtomic` |
| Database (db module) | `open`, `openWithoutPing`, `close`, `ping`, `isConnected`, `query`, `queryArrays`, `queryRow`, `exec`, `begin`, `commit`, `rollback`, `txExec`, `txQuery`, `prepare`, `stmtExec`, `stmtQuery`, `drivers`, `stats`, `columns` |
| Database (built-in) | `formatSQLValue`, `dbConnect`, `dbClose`, `dbQuery`, `dbQueryOrdered`, `dbQueryRecs`, `dbQueryMap`, `dbQueryMapArray`, `dbQueryCount`, `dbQueryFloat`, `dbQueryString`, `dbExec`, `dbQueryTyped`, `dbQueryRowTyped`, `dbQueryArrayTyped`, `dbQueryValueTyped` |
| Context | `newContext`, `contextWithTimeout`, `contextWithCancel`, `contextWithDeadline`, `contextCancel`, `contextDone` |
| Time | `sleep`, `now`, `nowMs`, `uuid` |
| Utility | `range`, `runCode`, `loadPlugin`, `format`, `checkErr`, `checkEmpty`, `genOtpCode`, `make`, `delegate` |

## License

MIT License
