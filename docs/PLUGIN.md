# Plugin System

Xxlang supports two types of plugins for high-performance operations:

| Feature | Static Plugin | WASM Plugin |
|---------|---------------|-------------|
| Platform | Windows, Linux, macOS | Windows, Linux, macOS |
| CGO Required | No | No |
| Runtime Loading | No (compile-time) | Yes |
| Performance | Fastest | Fast (~10-20% overhead) |
| Distribution | Compiled in | Single .wasm file |

---

## WebAssembly Plugins (Recommended)

### Quick Start

```bash
# 1. Build .wasm file (AssemblyScript - smallest)
asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2

# 2. Load and use
go run main.go
```

```go
loader := plugin.NewLoader()
p, err := loader.LoadPath("./fib.wasm")  // Direct file path
plugin.Register(p)

interp.Eval(`import "plugin/fib"; println(fib.fast(50))`)
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

**Method 1: Load by file path (recommended)**

```go
loader := plugin.NewLoader()
p, err := loader.LoadPath("./plugins/fib.wasm")
```

**Method 2: Load by name with search paths**

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

```xxl
import "plugin/fib"

println(fib.version)           // "1.0.0-as"
println(fib.fast(50))          // 12586269025
println(fib.matrix(92))        // 7540113804746346429
println(fib.isFib(13))         // true
println(fib.range_(10))        // [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55]
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
    interp.Eval(`import "plugin/fib"; println(fib.fast(50))`)
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
├── main.go                  # Comprehensive test
├── test_loader_methods.go   # Loading methods test
└── plugin/
    ├── build.sh             # Build script
    ├── fib.ts               # AssemblyScript (978 bytes)
    ├── fib.c                # C (~1.5KB)
    ├── fib.zig              # Zig (~1.3KB)
    ├── fib.rs               # Rust (~1.3KB)
    └── fib.go               # TinyGo (~15KB)
```

```bash
cd examples/wasm_plugin

# Build and test
./plugin/build.sh fib.ts && go run main.go
./plugin/build.sh fib.rs && go run main.go

# Test loading methods
go run test_loader_methods.go
```

---

## Full Code Examples

See the `examples/wasm_plugin/plugin/` directory for complete implementations in each language.
