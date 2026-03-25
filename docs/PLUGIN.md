# Plugin System

Xxlang supports two types of plugins for high-performance operations:

| Feature | Static Plugin | WASM Plugin |
|---------|---------------|-------------|
| Platform | Windows, Linux, macOS | Windows, Linux, macOS |
| CGO Required | No | No |
| Runtime Loading | No (compile-time) | Yes |
| Performance | Fastest | Fast (~10-20% overhead) |
| Distribution | Compiled in | Single .wasm file |

## Table of Contents

- [WebAssembly Plugins (Recommended)](#webassembly-plugins-recommended)
- [Static Plugins](#static-plugins)
- [Performance](#performance)
- [Examples](#examples)
- [Full Code Examples](#full-code-examples)

---

## WebAssembly Plugins (Recommended)

### Quick Start

```bash
# 1. Build .wasm file (AssemblyScript - smallest)
asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2
```

**From Xxlang (recommended):**

```xxl
// Load and use directly in Xxlang
var fib = loadPlugin("./fib.wasm")
pln(fib.fast(50))
```

**From Go:**

```go
loader := plugin.NewLoader()
p, err := loader.LoadPath("./fib.wasm")
plugin.Register(p)

interp.Eval(`import "plugin/fib"; pln(fib.fast(50))`)
```

### Output Files

All languages produce a single `.wasm` file:

| Language | File Size | Build Command |
|----------|-----------|---------------|
| **AssemblyScript** | ~1KB | `asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2` |
| **C** | ~1.5KB | `clang -o fib.wasm --target=wasm32 -O2 fib.c -nostdlib -nostartfiles -Wl,--no-entry -Wl,--export-all` |
| **Zig** | ~1.3KB | `zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic` |
| **Rust** | ~1.3KB | `rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o fib.wasm fib.rs` |
| **TinyGo** | ~15KB | `tinygo build -o fib.wasm -target=wasi fib.go` |

### Loading Methods

**Method 1: Load from Xxlang code (recommended)**

```xxl
// Load WASM plugin directly in Xxlang
var fib = loadPlugin("./plugins/fib.wasm")

// Use the plugin
pln(fib.version)
pln(fib.fast(50))
```

**Method 2: Load from Go code**

```go
loader := plugin.NewLoader()
p, err := loader.LoadPath("./plugins/fib.wasm")
```

**Method 3: Load by name with search paths**

```go
loader := plugin.NewLoader()
loader.AddPath("./plugins")
p, err := loader.Load("fib")  // Searches for fib.wasm
```

### Supported Languages

| Language | Status | Notes |
|----------|--------|-------|
| AssemblyScript | Tested | Smallest output, TypeScript-like syntax |
| C | Tested | Most portable, requires clang |
| Zig | Tested | Modern language, excellent WASM support |
| Rust | Tested | Use `wasm32-unknown-unknown` target |
| TinyGo | Tested | Go syntax, requires Go 1.19-1.24 |
| C++ | Compatible | Use clang with wasm32 target |
| Standard Go | Not supported | Only exports `_start`, use TinyGo instead |

### Plugin Specification

Required exports:

| Function | Signature | Description |
|----------|-----------|-------------|
| `alloc` | `(size: u32) -> u32` | Memory allocator (required) |
| `plugin_name` | `(result_ptr: u32) -> void` | Returns plugin name |
| `plugin_version` | `(result_ptr: u32) -> void` | Returns version string |
| `call_xxx` | Various | Exported functions (`call_fast` → `fib.fast`) |

Data types:
- `i64` → `Int` (primary numeric type)
- `i32` → `Int` (boolean: 0/1)
- Strings/Arrays: write `(pointer, size)` to `result_ptr`

### Using from Xxlang

**Load plugin directly with loadPlugin():**

```xxl
// Load plugin by file path
var fib = loadPlugin("./plugins/fib.wasm")

pln(fib.version)           // "1.0.0-as"
pln(fib.fast(50))          // 12586269025
pln(fib.matrix(92))        // 7540113804746346429
pln(fib.isFib(13))         // true
pln(fib.range_(10))        // [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55]
```

**Or import with search paths (requires setup in Go):**

```xxl
import "plugin/fib"
pln(fib.fast(50))
```

---

## Static Plugins

Static plugins are Go packages compiled into your application.

### Creating a Static Plugin

```go
// plugin/fibplugin.go
package fibplugin

import (
    "github.com/topxeq/xxlang/pkg/objects"
    "github.com/topxeq/xxlang/pkg/plugin"
)

type FibPlugin struct{}

func (p *FibPlugin) Name() string { return "fib" }

func (p *FibPlugin) Exports() map[string]objects.Object {
    return map[string]objects.Object{
        "fast": &objects.Builtin{
            Fn: func(args ...objects.Object) objects.Object {
                n := args[0].(*objects.Int)
                return &objects.Int{Value: fibFast(n.Value)}
            },
        },
    }
}

func init() { plugin.Register(&FibPlugin{}) }
```

### Using Static Plugin

```go
import _ "github.com/topxeq/xxlang/examples/fib_plugin/plugin"

func main() {
    interp := interpreter.New(interpreter.WithStdlib())
    interp.Eval(`import "plugin/fib"; pln(fib.fast(50))`)
}
```

---

## Performance

| Method | fib(35) | Complexity |
|--------|---------|------------|
| Xxlang naive recursion | ~6.5s | O(2^n) |
| Xxlang tail recursion | ~70µs | O(n) |
| WASM fib.fast | ~50µs | O(n) |
| WASM fib.matrix | ~60µs | O(log n) |

**Key insight:** WASM plugins provide **100,000x** speedup over naive recursion.

---

## Examples

```
examples/wasm_plugin/
├── main.go                    # Comprehensive test
├── test_loadplugin_main.go    # loadPlugin() test from Xxlang
└── plugin/
    ├── build.sh               # Build script
    ├── fib.ts                 # AssemblyScript (978 bytes)
    ├── fib.c                  # C (~1.5KB)
    ├── fib.zig                # Zig (~1.3KB)
    ├── fib.rs                 # Rust (~1.3KB)
    └── fib.go                 # TinyGo (~15KB)
```

```bash
cd examples/wasm_plugin

# Build WASM plugin
./plugin/build.sh fib.ts

# Test from Go
go run main.go

# Test loadPlugin() from Xxlang
go run test_loadplugin_main.go
```

---

## Full Code Examples

### AssemblyScript Example (Recommended)

AssemblyScript produces the smallest WASM files (~1KB) and has TypeScript-like syntax:

```typescript
// fib.ts - AssemblyScript plugin for Fibonacci calculations

// Memory allocator (required)
var buffer: usize = 0

export function alloc(size: usize): usize {
    const ptr = buffer
    buffer += size
    return ptr
}

// Plugin metadata
export function plugin_name(ptr: usize): void {
    const name = "fib"
    for (let i = 0; i < name.length; i++) {
        store<u8>(ptr + i, name.charCodeAt(i))
    }
    store<u8>(ptr + name.length, 0)
}

export function plugin_version(ptr: usize): void {
    const version = "1.0.0-as"
    for (let i = 0; i < version.length; i++) {
        store<u8>(ptr + i, version.charCodeAt(i))
    }
    store<u8>(ptr + version.length, 0)
}

// Fast Fibonacci - O(n) iterative
export function call_fast(n: i64): i64 {
    if (n <= 1) return n
    let a: i64 = 0
    let b: i64 = 1
    for (let i = 2; i <= n; i++) {
        const temp = a + b
        a = b
        b = temp
    }
    return b
}

// Matrix exponentiation - O(log n)
export function call_matrix(n: i64): i64 {
    if (n <= 1) return n

    var a: i64 = 1, b: i64 = 1, c: i64 = 1, d: i64 = 0
    var ta: i64, tb: i64, tc: i64, td: i64
    var ra: i64 = 1, rb: i64 = 0, rc: i64 = 0, rd: i64 = 1

    while (n > 0) {
        if ((n & 1) == 1) {
            ta = ra * a + rb * c
            tb = ra * b + rb * d
            tc = rc * a + rd * c
            td = rc * b + rd * d
            ra = ta; rb = tb; rc = tc; rd = td
        }
        ta = a * a + b * c
        tb = a * b + b * d
        tc = c * a + d * c
        td = c * b + d * d
        a = ta; b = tb; c = tc; d = td
        n = n >> 1
    }
    return rb
}

// Check if number is Fibonacci
export function call_isFib(n: i64): i32 {
    // A number is Fibonacci if one or both of (5*n^2 + 4) or (5*n^2 - 4) is a perfect square
    const test1 = 5 * n * n + 4
    const test2 = 5 * n * n - 4
    return (isPerfectSquare(test1) || isPerfectSquare(test2)) ? 1 : 0
}

function isPerfectSquare(n: i64): bool {
    if (n < 0) return false
    const root = isqrt(n)
    return root * root == n
}

function isqrt(n: i64): i64 {
    if (n < 2) return n
    var x: i64 = n
    var y: i64 = (x + 1) >> 1
    while (y < x) {
        x = y
        y = (x + n / x) >> 1
    }
    return x
}

// Return array of Fibonacci numbers (writes to memory)
export function call_range_(n: i64, resultPtr: usize): usize {
    const startPtr = resultPtr
    // Write count
    store<i64>(resultPtr, n + 1)
    resultPtr += 8

    // Write Fibonacci sequence
    var a: i64 = 0, b: i64 = 1
    for (let i = 0; i <= n; i++) {
        store<i64>(resultPtr, a)
        resultPtr += 8
        const temp = a + b
        a = b
        b = temp
    }

    // Return total bytes written
    return resultPtr - startPtr
}
```

**Build command:**
```bash
asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2
```

### C Example

```c
// fib.c - C plugin for Fibonacci calculations
#include <stdint.h>

static uint8_t memory[65536];
static size_t buffer_ptr = 0;

// Required: Memory allocator
uint32_t alloc(uint32_t size) {
    uint32_t ptr = (uint32_t)buffer_ptr;
    buffer_ptr += size;
    return ptr;
}

// Plugin metadata
void plugin_name(uint32_t ptr) {
    const char* name = "fib";
    for (int i = 0; name[i]; i++) {
        ((uint8_t*)ptr)[i] = name[i];
    }
}

void plugin_version(uint32_t ptr) {
    const char* version = "1.0.0-c";
    for (int i = 0; version[i]; i++) {
        ((uint8_t*)ptr)[i] = version[i];
    }
}

// Fast Fibonacci - O(n) iterative
int64_t call_fast(int64_t n) {
    if (n <= 1) return n;
    int64_t a = 0, b = 1;
    for (int64_t i = 2; i <= n; i++) {
        int64_t temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

// Matrix exponentiation - O(log n)
int64_t call_matrix(int64_t n) {
    if (n <= 1) return n;

    int64_t a = 1, b = 1, c = 1, d = 0;
    int64_t ta, tb, tc, td;
    int64_t ra = 1, rb = 0, rc = 0, rd = 1;

    while (n > 0) {
        if (n & 1) {
            ta = ra * a + rb * c;
            tb = ra * b + rb * d;
            tc = rc * a + rd * c;
            td = rc * b + rd * d;
            ra = ta; rb = tb; rc = tc; rd = td;
        }
        ta = a * a + b * c;
        tb = a * b + b * d;
        tc = c * a + d * c;
        td = c * b + d * d;
        a = ta; b = tb; c = tc; d = td;
        n >>= 1;
    }
    return rb;
}

// Check if number is Fibonacci
int32_t call_isFib(int64_t n) {
    // Implementation similar to AssemblyScript version
    return 0; // Simplified
}

// Range function
uint32_t call_range_(int64_t n, uint32_t resultPtr) {
    int64_t* ptr = (int64_t*)resultPtr;
    *ptr++ = n + 1; // Count

    int64_t a = 0, b = 1;
    for (int64_t i = 0; i <= n; i++) {
        *ptr++ = a;
        int64_t temp = a + b;
        a = b;
        b = temp;
    }

    return (uint32_t)((uint8_t*)ptr - (uint8_t*)resultPtr);
}
```

**Build command:**
```bash
clang -o fib.wasm --target=wasm32 -O2 fib.c -nostdlib -nostartfiles \
    -Wl,--no-entry -Wl,--export-all
```

### Zig Example

```zig
// fib.zig - Zig plugin for Fibonacci calculations

var buffer: usize = 0;

export fn alloc(size: usize) usize {
    const ptr = buffer;
    buffer += size;
    return ptr;
}

export fn plugin_name(ptr: usize) void {
    const name = "fib";
    var i: usize = 0;
    while (i < name.len) : (i += 1) {
        @intToPtr(*u8, ptr + i).* = name[i];
    }
}

export fn plugin_version(ptr: usize) void {
    const version = "1.0.0-zig";
    var i: usize = 0;
    while (i < version.len) : (i += 1) {
        @intToPtr(*u8, ptr + i).* = version[i];
    }
}

export fn call_fast(n: i64) i64 {
    if (n <= 1) return n;
    var a: i64 = 0;
    var b: i64 = 1;
    var i: i64 = 2;
    while (i <= n) : (i += 1) {
        const temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

export fn call_matrix(n: i64) i64 {
    if (n <= 1) return n;
    // Matrix implementation...
    return call_fast(n); // Simplified
}
```

**Build command:**
```bash
zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic
```

---

See the `examples/wasm_plugin/plugin/` directory for complete implementations in each language.
