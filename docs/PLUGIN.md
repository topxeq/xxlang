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

---

## Static Plugins

Static plugins are Go packages that compile directly into your application.

### Creating a Static Plugin

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
        "version": &objects.String{Value: "1.0.0"},
    }
}

func fibFast(n int64) int64 {
    if n <= 1 { return n }
    a, b := int64(0), int64(1)
    for i := int64(2); i <= n; i++ {
        a, b = b, a+b
    }
    return b
}

func init() {
    plugin.Register(&FibPlugin{})
}
```

### Using Static Plugin

```go
// main.go
package main

import (
    "github.com/topxeq/xxlang/pkg/interpreter"
    _ "github.com/topxeq/xxlang/examples/fib_plugin/plugin"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())
    interp.Eval(`
        import "plugin/fib"
        println(fib.fast(50))
    `)
}
```

### Plugin Interface

```go
type Plugin interface {
    Name() string                              // Plugin name
    Exports() map[string]objects.Object        // Exported symbols
}
```

---

## WebAssembly Plugins

WASM plugins work on all platforms including Windows, without CGO. They can be written in any language that compiles to WebAssembly.

### Supported Languages

| Language | Status | Build Size | Notes |
|----------|--------|------------|-------|
| **C** | ✅ Tested | ~1.5KB | Most portable, requires clang |
| **Zig** | ✅ Tested | ~1.3KB | Modern language, excellent WASM support |
| **Rust** | ✅ Tested | ~1.3KB | Use `wasm32-unknown-unknown` target |
| **TinyGo** | ✅ Tested | ~15KB | Go syntax, Go 1.19-1.24 only |
| **C++** | ✅ Compatible | ~2KB | Use clang with wasm32 target |
| **AssemblyScript** | ✅ Compatible | ~1KB | TypeScript-like syntax |
| Standard Go | ❌ Not supported | ~1.8MB | Only exports `_start`, not custom functions |

**Why Standard Go doesn't work:** Go's wasip1 target is for standalone applications only. It ignores `//export` directives and only exports `_start` (entry point). Use TinyGo instead.

### Plugin Specification

A valid WASM plugin must export these functions:

| Function | Signature | Required | Description |
|----------|-----------|----------|-------------|
| `alloc` | `(size: u32) -> u32` | Yes | Memory allocator |
| `plugin_name` | `(result_ptr: u32) -> void` | Recommended | Returns plugin name |
| `plugin_version` | `(result_ptr: u32) -> void` | Optional | Returns version string |
| `call_xxx` | Various | As needed | Exported functions (prefix `call_` maps to `xxx` in Xxlang) |

**Data Type Mapping:**

| WASM Type | Xxlang Type | Notes |
|-----------|-------------|-------|
| `i64` | `Int` | Primary numeric type |
| `i32` | `Int` | For boolean returns (0/1) |
| `u32` | Pointer | Memory pointer for strings/arrays |

**String/Array Convention:**

Functions returning strings or arrays write `(pointer, size)` to `result_ptr`:

```c
void call_range_(int64_t n, uint32_t result_ptr) {
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = data_pointer;  // Pointer to data
    result[1] = data_size;      // Size in bytes
}
```

---

## Building WASM Plugins

### C

Most portable option. Requires `clang` with wasm32 target.

```bash
# Install (Ubuntu/Debian)
apt install clang lld

# Build
clang -o fib.wasm --target=wasm32 -O2 fib.c \
    -nostdlib -nostartfiles \
    -Wl,--no-entry -Wl,--export-all
```

**Complete Example:**

```c
// fib.c
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

// Plugin metadata
void plugin_name(uint32_t result_ptr) {
    const char* name = "fib";
    uint32_t offset = alloc(3);
    unsigned char* mem = (unsigned char*)offset;
    for (int i = 0; i < 3; i++) mem[i] = name[i];
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = offset;
    result[1] = 3;
}

void plugin_version(uint32_t result_ptr) {
    const char* version = "1.0.0-c";
    uint32_t offset = alloc(7);
    unsigned char* mem = (unsigned char*)offset;
    for (int i = 0; i < 7; i++) mem[i] = version[i];
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = offset;
    result[1] = 7;
}

// Fibonacci - O(n)
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

// Fibonacci - O(log n) matrix algorithm
int64_t call_matrix(int64_t n) {
    if (n <= 1) return n;
    int64_t result[2][2] = {{1, 0}, {0, 1}};
    int64_t base[2][2] = {{1, 1}, {1, 0}};
    int64_t temp[2][2];
    while (n > 0) {
        if (n & 1) {
            for (int i = 0; i < 2; i++)
                for (int j = 0; j < 2; j++)
                    temp[i][j] = result[i][0]*base[0][j] + result[i][1]*base[1][j];
            for (int i = 0; i < 2; i++)
                for (int j = 0; j < 2; j++)
                    result[i][j] = temp[i][j];
        }
        for (int i = 0; i < 2; i++)
            for (int j = 0; j < 2; j++)
                temp[i][j] = base[i][0]*base[0][j] + base[i][1]*base[1][j];
        for (int i = 0; i < 2; i++)
            for (int j = 0; j < 2; j++)
                base[i][j] = temp[i][j];
        n >>= 1;
    }
    return result[0][1];
}

void _start(void) {}
```

### Zig

Modern systems language with excellent WASM support.

```bash
# Install from https://ziglang.org/

# Build
zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic
```

**Complete Example:**

```zig
// fib.zig
var heap_ptr: usize = 65536;

export fn alloc(size: u32) u32 {
    if (heap_ptr % 8 != 0) heap_ptr += 8 - (heap_ptr % 8);
    const result: u32 = @intCast(heap_ptr);
    heap_ptr += size;
    return result;
}

fn writeString(s: []const u8, result_ptr: u32) void {
    const offset = alloc(@intCast(s.len));
    const mem = @as([*]u8, @ptrFromInt(offset));
    @memcpy(mem, s);
    const result = @as(*[2]u32, @ptrFromInt(result_ptr));
    result[0] = offset;
    result[1] = @intCast(s.len);
}

export fn plugin_name(result_ptr: u32) void {
    writeString("fib", result_ptr);
}

export fn plugin_version(result_ptr: u32) void {
    writeString("1.0.0-zig", result_ptr);
}

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

export fn call_matrix(n: i64) i64 {
    if (n <= 1) return n;
    var result = [2][2]i64{ .{ 1, 0 }, .{ 0, 1 } };
    var base = [2][2]i64{ .{ 1, 1 }, .{ 1, 0 } };
    var m = n;
    while (m > 0) {
        if (m & 1 == 1) {
            result = .{
                .{ result[0][0]*base[0][0] + result[0][1]*base[1][0], result[0][0]*base[0][1] + result[0][1]*base[1][1] },
                .{ result[1][0]*base[0][0] + result[1][1]*base[1][0], result[1][0]*base[0][1] + result[1][1]*base[1][1] },
            };
        }
        base = .{
            .{ base[0][0]*base[0][0] + base[0][1]*base[1][0], base[0][0]*base[0][1] + base[0][1]*base[1][1] },
            .{ base[1][0]*base[0][0] + base[1][1]*base[1][0], base[1][0]*base[0][1] + base[1][1]*base[1][1] },
        };
        m >>= 1;
    }
    return result[0][1];
}

export fn call_isFib(n: i64) i32 {
    if (n < 0) return 0;
    const n2 = n * n;
    var sqrt1: i64 = 0;
    while (sqrt1 * sqrt1 < 5 * n2 + 4) : (sqrt1 += 1) {}
    if (sqrt1 * sqrt1 == 5 * n2 + 4) return 1;
    var sqrt2: i64 = 0;
    while (sqrt2 * sqrt2 < 5 * n2 - 4) : (sqrt2 += 1) {}
    if (sqrt2 * sqrt2 == 5 * n2 - 4) return 1;
    return 0;
}

export fn call_range_(n: i64, result_ptr: u32) void {
    const result = @as(*[2]u32, @ptrFromInt(result_ptr));
    if (n < 0) {
        result[0] = 0;
        result[1] = 0;
        return;
    }
    const count: u32 = @intCast(n + 1);
    const ptr = alloc(count * 8);
    const arr = @as([*]i64, @ptrFromInt(ptr));
    var a: i64 = 0;
    var b: i64 = 1;
    for (0..@intCast(n + 1)) |i| {
        arr[i] = a;
        const tmp = a + b;
        a = b;
        b = tmp;
    }
    result[0] = ptr;
    result[1] = count;
}
```

### Rust

Requires `wasm32-unknown-unknown` target.

```bash
# Install target
rustup target add wasm32-unknown-unknown

# Build
rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o fib.wasm fib.rs
```

**Complete Example:**

```rust
// fib.rs
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

fn write_string(s: &str, result_ptr: u32) {
    unsafe {
        let bytes = s.as_bytes();
        let offset = alloc(bytes.len() as u32);
        let mem = offset as *mut u8;
        for (i, &byte) in bytes.iter().enumerate() {
            *mem.add(i) = byte;
        }
        let result = result_ptr as *mut u32;
        *result = offset;
        *result.add(1) = bytes.len() as u32;
    }
}

#[no_mangle]
pub extern "C" fn plugin_name(result_ptr: u32) {
    write_string("fib", result_ptr);
}

#[no_mangle]
pub extern "C" fn plugin_version(result_ptr: u32) {
    write_string("1.0.0-rust", result_ptr);
}

#[no_mangle]
pub extern "C" fn call_fast(n: i64) -> i64 {
    if n <= 1 { return n; }
    let mut a: i64 = 0;
    let mut b: i64 = 1;
    for _ in 2..=n {
        let tmp = a + b;
        a = b;
        b = tmp;
    }
    b
}

#[no_mangle]
pub extern "C" fn call_matrix(n: i64) -> i64 {
    if n <= 1 { return n; }
    let mut result = [[1i64, 0], [0, 1]];
    let mut base = [[1i64, 1], [1, 0]];
    let mut m = n;
    while m > 0 {
        if m & 1 == 1 {
            result = [
                [result[0][0]*base[0][0] + result[0][1]*base[1][0], result[0][0]*base[0][1] + result[0][1]*base[1][1]],
                [result[1][0]*base[0][0] + result[1][1]*base[1][0], result[1][0]*base[0][1] + result[1][1]*base[1][1]],
            ];
        }
        base = [
            [base[0][0]*base[0][0] + base[0][1]*base[1][0], base[0][0]*base[0][1] + base[0][1]*base[1][1]],
            [base[1][0]*base[0][0] + base[1][1]*base[1][0], base[1][0]*base[0][1] + base[1][1]*base[1][1]],
        ];
        m >>= 1;
    }
    result[0][1]
}

#[no_mangle]
pub extern "C" fn call_isFib(n: i64) -> i32 {
    if n < 0 { return 0; }
    let n2 = n * n;
    let check1 = 5 * n2 + 4;
    let check2 = 5 * n2 - 4;
    fn is_perfect_square(x: i64) -> bool {
        let mut sqrt: i64 = 0;
        while sqrt * sqrt < x { sqrt += 1; }
        sqrt * sqrt == x
    }
    if is_perfect_square(check1) || is_perfect_square(check2) { 1 } else { 0 }
}

#[no_mangle]
pub extern "C" fn call_range_(n: i64, result_ptr: u32) {
    unsafe {
        let result = result_ptr as *mut u32;
        if n < 0 {
            *result = 0;
            *result.add(1) = 0;
            return;
        }
        let count = (n + 1) as u32;
        let ptr = alloc(count * 8);
        let arr = ptr as *mut i64;
        let mut a: i64 = 0;
        let mut b: i64 = 1;
        for i in 0..=n {
            *arr.add(i as usize) = a;
            let tmp = a + b;
            a = b;
            b = tmp;
        }
        *result = ptr;
        *result.add(1) = count;
    }
}
```

### TinyGo

Go syntax with WASM support. **Note: TinyGo 0.36 only supports Go 1.19-1.24.**

```bash
# Install from https://tinygo.org/

# Build
tinygo build -o fib.wasm -target=wasi fib.go
```

**Complete Example:**

```go
// fib.go
package main

import "unsafe"

var memory []byte

//export alloc
func alloc(size uint32) uint32 {
    offset := uint32(len(memory))
    if offset%8 != 0 {
        offset += 8 - offset%8
    }
    newLen := offset + size
    if uint32(cap(memory)) < newLen {
        newMem := make([]byte, newLen*2)
        copy(newMem, memory)
        memory = newMem[:newLen]
    } else {
        memory = memory[:newLen]
    }
    return offset
}

//export plugin_name
func pluginName() (ptr uint32, size uint32) {
    name := "fib"
    ptr = alloc(uint32(len(name)))
    copy(memory[ptr:], name)
    return ptr, uint32(len(name))
}

//export plugin_version
func pluginVersion() (ptr uint32, size uint32) {
    version := "1.0.0-wasm"
    ptr = alloc(uint32(len(version)))
    copy(memory[ptr:], version)
    return ptr, uint32(len(version))
}

//export call_fast
func fibFast(n int64) int64 {
    if n <= 1 { return n }
    a, b := int64(0), int64(1)
    for i := int64(2); i <= n; i++ {
        a, b = b, a+b
    }
    return b
}

//export call_matrix
func fibMatrix(n int64) int64 {
    if n <= 1 { return n }
    mul := func(a, b [2][2]int64) [2][2]int64 {
        return [2][2]int64{
            {a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
            {a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
        }
    }
    result := [2][2]int64{{1, 0}, {0, 1}}
    base := [2][2]int64{{1, 1}, {1, 0}}
    for n > 0 {
        if n&1 == 1 { result = mul(result, base) }
        base = mul(base, base)
        n >>= 1
    }
    return result[0][1]
}

//export call_isFib
func isFib(n int64) int32 {
    if n < 0 { return 0 }
    n2 := n * n
    check := func(x int64) bool {
        sqrt := int64(0)
        for sqrt*sqrt < x { sqrt++ }
        return sqrt*sqrt == x
    }
    if check(5*n2+4) || check(5*n2-4) { return 1 }
    return 0
}

//export call_range_
func fibRange(n int64) (ptr uint32, count uint32) {
    if n < 0 { return 0, 0 }
    count = uint32(n + 1)
    ptr = alloc(count * 8)
    a, b := int64(0), int64(1)
    for i := int64(0); i <= n; i++ {
        *(*int64)(unsafe.Pointer(uintptr(ptr + uint32(i*8)))) = a
        a, b = b, a+b
    }
    return ptr, count
}

func main() {}
```

---

## Using Plugins from Xxlang

```xxl
import "plugin/fib"

// Version
println("Version: " + fib.version)

// Basic functions
println("fib(10) = " + fib.fast(10))
println("fib(50) = " + fib.fast(50))
println("fib(92) = " + fib.matrix(92))

// Utility
println("Is 13 Fibonacci? " + fib.isFib(13))

// Batch
var fibs = fib.range_(10)
println("First 11: " + fibs.toStr())
```

---

## Performance Comparison

| Method | fib(35) | Complexity |
|--------|---------|------------|
| Xxlang naive recursion | ~6.5s | O(2^n) |
| Xxlang tail recursion | ~70µs | O(n) |
| WASM fib.fast | ~50µs | O(n) |
| WASM fib.matrix | ~60µs | O(log n) |

**Key insight:** WASM plugins provide **100,000x** speedup over naive recursion.

---

## Complete Examples

See [examples/wasm_plugin/](../examples/wasm_plugin/):

```
examples/wasm_plugin/
├── main.go           # Test program
└── plugin/
    ├── build.sh      # Build script (supports .c, .zig, .rs, .go)
    ├── fib.c         # C implementation
    ├── fib.zig       # Zig implementation
    ├── fib.rs        # Rust implementation
    └── fib.go        # TinyGo implementation
```

Run tests:

```bash
cd examples/wasm_plugin

# Build with any language
./plugin/build.sh fib.c      # C
./plugin/build.sh fib.zig    # Zig
./plugin/build.sh fib.rs     # Rust
./plugin/build.sh fib.go     # TinyGo

# Test
go run main.go
```

---

## Best Practices

1. **Validate arguments** - Check count and types
2. **Return errors** - Use `&objects.Error{Message: "..."}`
3. **Handle edge cases** - Test with zero, negative, boundary values
4. **Use efficient algorithms** - Prefer O(log n) over O(n)
5. **Batch operations** - Return arrays for multiple results
