# Plugin System

Xxlang supports two types of plugins for high-performance operations:

1. **Static Plugins** - Compile into your Go application (all platforms)
2. **WebAssembly Plugins** - Load `.wasm` files at runtime (all platforms, no CGO)

## Plugin Types Comparison

| Feature | Static Plugin | WASM Plugin |
|---------|---------------|-------------|
| Windows Support | ✅ Yes | ✅ Yes |
| Linux Support | ✅ Yes | ✅ Yes |
| macOS Support | ✅ Yes | ✅ Yes |
| Requires CGO | ❌ No | ❌ No |
| Runtime Loading | ❌ No | ✅ Yes |
| Performance | Fastest | Fast (~10-20% overhead) |
| Security | Host process | Sandboxed |
| Distribution | Compiled in | Single .wasm file |

## Overview

Plugins provide:

- **Native Go performance** - Execute Go code directly from Xxlang
- **Complex algorithms** - Implement sophisticated algorithms in Go (e.g., matrix operations)
- **Extended functionality** - Add new capabilities not available in standard library
- **Batch processing** - Process multiple values efficiently in a single call

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Xxlang Code   │────▶│   Interpreter   │────▶│  WASM Plugin    │
│                 │     │                 │     │  (fib.wasm)     │
│ import "plugin/ │     │  Plugin Loader  │     │                 │
│        fib"     │     │                 │     │  - fib.fast()   │
│ fib.fast(50)    │◀────│  Registry       │◀────│  - fib.matrix() │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Creating a Plugin

### 1. Implement the Plugin Interface

```go
// plugin/fibplugin.go
package fibplugin

import (
    "github.com/topxeq/xxlang/pkg/objects"
    "github.com/topxeq/xxlang/pkg/plugin"
)

// FibPlugin implements plugin.Plugin interface
type FibPlugin struct{}

// Name returns the plugin name (used as "plugin/fib" in imports)
func (p *FibPlugin) Name() string {
    return "fib"
}

// Exports returns the functions and variables accessible from Xxlang
func (p *FibPlugin) Exports() map[string]objects.Object {
    return map[string]objects.Object{
        // High-performance Fibonacci - O(n) time complexity
        "fast": &objects.Builtin{
            Fn: func(args ...objects.Object) objects.Object {
                if len(args) != 1 {
                    return &objects.Error{Message: "fib.fast requires 1 argument"}
                }

                n, ok := args[0].(*objects.Int)
                if !ok {
                    return &objects.Error{Message: "argument must be integer"}
                }

                result := fibFast(n.Value)
                return &objects.Int{Value: result}
            },
        },

        // Plugin version
        "version": &objects.String{Value: "1.0.0"},
    }
}

// Go-native implementation
func fibFast(n int64) int64 {
    if n <= 1 {
        return n
    }
    a, b := int64(0), int64(1)
    for i := int64(2); i <= n; i++ {
        a, b = b, a+b
    }
    return b
}

// Register plugin automatically on import
func init() {
    plugin.Register(&FibPlugin{})
}
```

### 2. Use Plugin in Main Program

```go
// main.go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"

    // Import plugin to trigger init() registration
    _ "github.com/topxeq/xxlang/examples/fib_plugin/plugin"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Use plugin in Xxlang
    code := `
        import "plugin/fib"

        println("Version: " + fib.version)
        println("fib(50) = " + fib.fast(50).toStr())
    `

    interp.Eval(code)
}
```

## Plugin Interface

```go
type Plugin interface {
    // Name returns the plugin name
    // Used as "plugin/<name>" in Xxlang imports
    Name() string

    // Exports returns the module's exported symbols
    // Keys are names accessible from Xxlang code
    Exports() map[string]objects.Object
}
```

## Exporting Different Types

### Functions

```go
"myFunc": &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        // Process arguments
        // Return result
        return &objects.Int{Value: 42}
    },
},
```

### Variables

```go
// String
"version": &objects.String{Value: "1.0.0"},

// Integer
"maxSize": &objects.Int{Value: 1000},

// Float
"pi": &objects.Float{Value: 3.14159},

// Boolean
"enabled": objects.TRUE,

// Array
"defaults": &objects.Array{Elements: []objects.Object{
    &objects.Int{Value: 1},
    &objects.Int{Value: 2},
    &objects.Int{Value: 3},
}},
```

## Advanced Example: Matrix Fast Power

The Fibonacci example includes a matrix fast power implementation for O(log n) complexity:

```go
// Matrix fast power - O(log n) time complexity
"matrix": &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        n, _ := args[0].(*objects.Int)
        result := fibMatrix(n.Value)
        return &objects.Int{Value: result}
    },
},

func fibMatrix(n int64) int64 {
    if n <= 1 {
        return n
    }

    // Matrix multiplication
    mul := func(a, b [2][2]int64) [2][2]int64 {
        return [2][2]int64{
            {a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
            {a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
        }
    }

    // Fast power
    result := [2][2]int64{{1, 0}, {0, 1}}
    base := [2][2]int64{{1, 1}, {1, 0}}

    for n > 0 {
        if n&1 == 1 {
            result = mul(result, base)
        }
        base = mul(base, base)
        n >>= 1
    }

    return result[0][1]
}
```

## Performance Comparison

| Method | fib(35) Time | Complexity |
|--------|--------------|------------|
| Xxlang naive recursion | ~6.5 seconds | O(2^n) |
| Xxlang tail recursion (TCO) | ~136 µs | O(n) |
| Plugin fib.fast | ~37 µs | O(n) |
| Plugin fib.matrix | ~35 µs | O(log n) |

**Key insight**: Go plugins provide **180,000x** speedup over naive Xxlang recursion, and even **3-4x** faster than optimized Xxlang tail recursion.

## Using Plugins from Xxlang

```xxl
// Import the plugin
import "plugin/fib"

// Access exported variables
println("Plugin version: " + fib.version)

// Call exported functions
println("fib(10) = " + fib.fast(10).toStr())
println("fib(50) = " + fib.fast(50).toStr())

// Use matrix algorithm for large numbers
println("fib(92) = " + fib.matrix(92).toStr())

// Batch processing
var fibs = fib.range_(10)
println("First 11 Fibonacci numbers: " + fibs.toStr())

// Utility functions
println("Is 13 a Fibonacci number? " + fib.isFib(13).toStr())
```

## int64 Limits

The Fibonacci plugin uses `int64`, which has limits:

- Maximum value: `9,223,372,036,854,775,807`
- Largest Fibonacci in range: `fib(92) = 7,540,113,804,746,346,429`
- `fib(93)` overflows int64

For larger numbers, use `math/big.Int` in your plugin.

## Static vs WASM Plugins

| Aspect | Static (import) | WASM (.wasm) |
|--------|-----------------|--------------|
| Platform | All platforms | **All platforms** |
| Windows | ✅ Yes | ✅ **Yes** |
| Requires CGO | ❌ No | ❌ **No** |
| Distribution | Compiled in | Single .wasm file |
| Updates | Recompile required | Replace .wasm file |
| Debugging | Easier | Medium |
| Performance | Fastest | ~10-20% overhead |
| Recommended | Yes | **Yes** |

### WASM Plugin Language Comparison

| Language | Build Size | Performance | Difficulty | Notes |
|----------|------------|-------------|------------|-------|
| C | Smallest (~1.5KB) | Best | Medium | Most portable, requires clang |
| Zig | Small (~1.3KB) | Best | Easy | Modern language, excellent WASM support |
| TinyGo | Medium (~15KB) | Good | Easy | Go syntax, Go 1.19-1.24 only |
| Standard Go | ~1.8MB | N/A | N/A | ❌ **Does NOT work** - no function exports |

## WebAssembly Plugins

WASM plugins work on all platforms including Windows, without CGO.

**Important**: WASM plugins require **TinyGo**. Standard Go's `GOOS=wasip1` does NOT work for plugins.

**Why Standard Go doesn't work:**

| Issue | Standard Go (wasip1) | TinyGo |
|-------|----------------------|--------|
| Export custom functions | ❌ Not supported | ✅ Supported |
| Plugin mode | ❌ Application-only | ✅ Library mode |
| File size | ~1.8MB | ~15KB |

Standard Go's wasip1 target is designed for **standalone applications** (CLI tools), not libraries/plugins. The `//export` directive is ignored, and only `_start` and `memory` are exported. When `main()` returns, the Go runtime calls `proc_exit(0)`, terminating the module.

**TinyGo uses a different architecture** that supports both application and library modes.

### Installing TinyGo

```bash
# macOS
brew install tinygo

# Linux (Ubuntu/Debian)
wget https://github.com/tinygo-org/tinygo/releases/download/v0.36.0/tinygo_0.36.0_amd64.deb
sudo dpkg -i tinygo_0.36.0_amd64.deb

# Windows
# Download from https://tinygo.org/getting-started/install/windows/
```

**Note**: TinyGo version compatibility:
- TinyGo 0.36: Go 1.19-1.24
- Check [tinygo.org](https://tinygo.org) for latest compatibility

### Creating a WASM Plugin

```go
// plugin/fib.go
package main

import "unsafe"

var memory []byte

//export alloc
func alloc(size uint32) uint32 {
	offset := uint32(len(memory))
	memory = append(memory, make([]byte, size)...)
	return offset
}

//export plugin_name
func pluginName() (ptr uint32, size uint32) {
	name := "fib"
	ptr = alloc(uint32(len(name)))
	copy(memory[ptr:], name)
	return ptr, uint32(len(name))
}

//export call_fast
func fibFast(n int64) int64 {
	if n <= 1 {
		return n
	}
	a, b := int64(0), int64(1)
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {} // Required but not used
```

### Building WASM Plugins

#### Option 1: Using C (Recommended)

C is the most portable option for WASM plugins. You need `clang` with wasm32 target support.

```bash
# Install clang (Ubuntu/Debian)
apt install clang lld

# Build
clang -o fib.wasm --target=wasm32 -O2 fib.c \
    -nostdlib -nostartfiles \
    -Wl,--no-entry -Wl,--export-all
```

#### Option 2: Using Zig

Zig is a modern systems programming language with excellent WASM support.

```bash
# Install Zig from https://ziglang.org/learn/getting-started/

# Build
zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic
```

Complete Zig example:

```zig
// fib.zig - Fibonacci WASM plugin
var heap_ptr: usize = 65536; // Start at 64KB offset

// Memory allocation from WASM heap
export fn alloc(size: u32) u32 {
    if (heap_ptr % 8 != 0) {
        heap_ptr += 8 - (heap_ptr % 8);
    }
    const result: u32 = @intCast(heap_ptr);
    heap_ptr += size;
    return result;
}

// Plugin name - writes (ptr, size) to result_ptr
export fn plugin_name(result_ptr: u32) void {
    const offset = alloc(3);
    const mem = @as([*]u8, @ptrFromInt(offset));
    @memcpy(mem, "fib");
    const result = @as(*[2]u32, @ptrFromInt(result_ptr));
    result[0] = offset;
    result[1] = 3;
}

// Fibonacci function - accessible as "fast" in Xxlang
export fn call_fast(n: i64) i64 {
    if (n <= 1) return n;
    var a: i64 = 0;
    var b: i64 = 1;
    var i: i64 = 2;
    while (i <= n) : (i += 1) {
        const tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}
```

#### Option 3: Using TinyGo

**Note**: TinyGo 0.36 supports Go 1.19-1.24 only. For newer Go versions, use C or Zig instead.

**Do NOT use standard Go** (`GOOS=wasip1 GOARCH=wasm go build`) - it won't work. Standard Go exports only `_start` (entry point), not custom functions.

```bash
# Build with TinyGo
tinygo build -o fib.wasm -target=wasi fib.go
```

### WASM Plugin Specification

Any language that can compile to WebAssembly and export functions can be used to write plugins.

#### Supported Languages

| Language | Status | Build Size | Notes |
|----------|--------|------------|-------|
| **C** | ✅ Tested | ~1.5KB | Most portable, requires clang |
| **Zig** | ✅ Tested | ~1.3KB | Modern language, excellent WASM support |
| **Rust** | ✅ Tested | ~1.3KB | Use `wasm32-unknown-unknown` target |
| **TinyGo** | ✅ Tested | ~15KB | Go syntax, Go 1.19-1.24 only |
| **C++** | ✅ Compatible | ~2KB | Use clang with wasm32 target |
| **AssemblyScript** | ✅ Compatible | ~1KB | TypeScript-like syntax, WASM-native |
| **Standard Go** | ❌ Not supported | ~1.8MB | Only exports `_start`, not custom functions |

#### Plugin Requirements

A valid WASM plugin must:

1. **Export `alloc(size: u32) -> u32`** - Memory allocator for passing strings/arrays
2. **Export `plugin_name(result_ptr: u32)`** - Returns plugin name (optional but recommended)
3. **Export `plugin_version(result_ptr: u32)`** - Returns version string (optional)
4. **Export functions with `call_` prefix** - e.g., `call_fast` becomes `fast` in Xxlang

#### Data Type Mapping

| WASM Type | Xxlang Type | Notes |
|-----------|-------------|-------|
| `i64` | `Int` | Primary numeric type |
| `i32` | `Int` | For boolean returns (0/1) |
| `u32` | Memory pointer | For string/array operations |

#### String/Array Convention

For functions that return strings or arrays:

```c
// Write result to result_ptr as (pointer, size) pair
void call_something(uint32_t result_ptr) {
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = data_pointer;  // Pointer to data
    result[1] = data_size;      // Size in bytes
}
```

### WASM Plugin Structure

A WASM plugin must export these functions:

```c
// Required: Memory allocator
uint32_t alloc(uint32_t size);

// Optional: Plugin metadata
void plugin_name(uint32_t result_ptr);   // Writes (ptr, size) to result_ptr
void plugin_version(uint32_t result_ptr);

// Exported functions (prefix with "call_" for Xxlang)
int64_t call_fast(int64_t n);            // Accessible as "fast" in Xxlang
int64_t call_matrix(int64_t n);          // Accessible as "matrix" in Xxlang
```

**Zig equivalent:**

```zig
// Memory allocator
export fn alloc(size: u32) u32 { ... }

// Plugin metadata
export fn plugin_name(result_ptr: u32) void { ... }
export fn plugin_version(result_ptr: u32) void { ... }

// Exported functions
export fn call_fast(n: i64) i64 { ... }
export fn call_matrix(n: i64) i64 { ... }
```

**Rust equivalent:**

```rust
#![no_std]
#![no_main]

use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo) -> ! { loop {} }

static mut HEAP_PTR: usize = 65536;

#[no_mangle]
pub extern "C" fn alloc(size: u32) -> u32 {
    unsafe {
        if HEAP_PTR % 8 != 0 {
            HEAP_PTR += 8 - (HEAP_PTR % 8);
        }
        let result = HEAP_PTR as u32;
        HEAP_PTR += size as usize;
        result
    }
}

#[no_mangle]
pub extern "C" fn call_fast(n: i64) -> i64 {
    if n <= 1 { return n; }
    let (mut a, mut b) = (0i64, 1i64);
    for _ in 2..=n {
        let tmp = a + b;
        a = b;
        b = tmp;
    }
    b
}
```

Build: `rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o fib.wasm fib.rs`

**AssemblyScript equivalent:**

```typescript
var heapPtr: usize = 65536;

export function alloc(size: u32): u32 {
    if (heapPtr % 8 != 0) heapPtr += 8 - (heapPtr % 8);
    const result = heapPtr;
    heapPtr += size;
    return result as u32;
}

export function call_fast(n: i64): i64 {
    if (n <= 1) return n;
    let a: i64 = 0, b: i64 = 1;
    for (let i = 2; i <= n; i++) {
        const tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}
```

Build: `asc fib.ts -o fib.wasm --optimize`

### Complete C Example

```c
// fib.c - Fibonacci WASM plugin
#include <stdint.h>

extern unsigned char __heap_base;
static uintptr_t heap_ptr = 0;

// Memory allocator
uint32_t alloc(uint32_t size) {
    if (heap_ptr == 0) heap_ptr = (uintptr_t)&__heap_base;
    if (heap_ptr % 8 != 0) heap_ptr += 8 - (heap_ptr % 8);
    uintptr_t result = heap_ptr;
    heap_ptr += size;
    return (uint32_t)result;
}

// Plugin name
void plugin_name(uint32_t result_ptr) {
    const char* name = "fib";
    uint32_t offset = alloc(3);
    unsigned char* mem = (unsigned char*)offset;
    for (int i = 0; i < 3; i++) mem[i] = name[i];
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = offset;
    result[1] = 3;
}

// Fibonacci function - accessible as "fast" in Xxlang
int64_t call_fast(int64_t n) {
    if (n <= 1) return n;
    int64_t a = 0, b = 1;
    for (int64_t i = 2; i <= n; i++) {
        int64_t tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}

// Entry point (required but not used)
void _start(void) {}
```

Build and test:

```bash
# Build
clang -o fib.wasm --target=wasm32 -O2 fib.c \
    -nostdlib -nostartfiles \
    -Wl,--no-entry -Wl,--export-all

# Test in Xxlang
xxlang -e 'import "plugin/fib"; println(fib.fast(50))'
```

## Complete Example

See [examples/wasm_plugin/](../examples/wasm_plugin/) for a complete working WASM plugin example:

```
examples/wasm_plugin/
├── main.go           # Test program
└── plugin/
    ├── build.sh      # Build script (supports C, Go, Zig)
    ├── fib.c         # C implementation
    ├── fib.go        # TinyGo implementation
    ├── fib.zig       # Zig implementation
    └── fib.wasm      # Compiled WASM plugin
```

Run the example:

```bash
cd examples/wasm_plugin

# Build with C (recommended)
cd plugin && ./build.sh fib.c

# Or build with Zig
cd plugin && ./build.sh fib.zig

# Or build with TinyGo
cd plugin && ./build.sh fib.go

# Test
cd ..
go run main.go
```

## Best Practices

1. **Validate arguments** - Check argument count and types
2. **Return errors** - Use `&objects.Error{Message: "..."}` for invalid inputs
3. **Document exports** - Comment each exported function and variable
4. **Handle edge cases** - Test with zero, negative, and boundary values
5. **Use appropriate algorithms** - Choose O(log n) over O(n) when possible
6. **Batch operations** - Return arrays for multiple results

## Error Handling in Plugins

```go
"safeDiv": &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        // Check argument count
        if len(args) != 2 {
            return &objects.Error{Message: "safeDiv requires 2 arguments"}
        }

        // Check argument types
        a, ok1 := args[0].(*objects.Int)
        b, ok2 := args[1].(*objects.Int)

        if !ok1 || !ok2 {
            return &objects.Error{Message: "arguments must be integers"}
        }

        // Check for division by zero
        if b.Value == 0 {
            return &objects.Error{Message: "division by zero"}
        }

        return &objects.Int{Value: a.Value / b.Value}
    },
},
```
