# Plugin System

Xxlang supports two types of plugins for high-performance operations:

1. **Static Plugins** - Compile into your Go application (all platforms)
2. **WebAssembly Plugins** - Load `.wasm` files at runtime (all platforms, no CGO)

## Plugin Types Comparison

| Feature | Static Plugin | WASM Plugin | Native .so Plugin |
|---------|---------------|-------------|-------------------|
| Windows Support | ✅ Yes | ✅ Yes | ❌ No |
| Linux Support | ✅ Yes | ✅ Yes | ✅ Yes |
| macOS Support | ✅ Yes | ✅ Yes | ✅ Yes |
| Requires CGO | ❌ No | ❌ No | ✅ Yes |
| Runtime Loading | ❌ No | ✅ Yes | ✅ Yes |
| Performance | Fastest | Fast (~10-20% overhead) | Fastest |
| Security | Host process | Sandboxed | Host process |
| Distribution | Compiled in | Single .wasm file | .so file |

## Overview

Plugins provide:

- **Native Go performance** - Execute Go code directly from Xxlang
- **Complex algorithms** - Implement sophisticated algorithms in Go (e.g., matrix operations)
- **Extended functionality** - Add new capabilities not available in standard library
- **Batch processing** - Process multiple values efficiently in a single call

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Xxlang Code   │────▶│   Interpreter   │────▶│  Go Plugin      │
│                 │     │                 │     │  (fib.so)       │
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

## Dynamic Plugin Loading (.so files)

On Linux/macOS, you can load plugins from `.so` files at runtime:

### Build Plugin

```bash
go build -buildmode=plugin -o plugins/fib.so plugin/fibplugin.go
```

### Load in Program

```go
import "github.com/topxeq/xxlang/pkg/plugin"

loader := plugin.NewLoader()
loader.AddPath("./plugins")

p, err := loader.Load("fib")
if err != nil {
    panic(err)
}

// Plugin is now available as "plugin/fib" in Xxlang
```

**Note**: Dynamic loading requires:
- Linux, macOS, or FreeBSD (not Windows)
- Same Go version for plugin and main program
- CGO enabled

## Static vs Dynamic Plugins

| Aspect | Static (import) | WASM (.wasm) | Native (.so) |
|--------|-----------------|--------------|--------------|
| Platform | All platforms | **All platforms** | Linux/macOS only |
| Windows | ✅ Yes | ✅ **Yes** | ❌ No |
| Requires CGO | ❌ No | ❌ **No** | ✅ Yes |
| Distribution | Compiled in | Single .wasm file | .so file |
| Updates | Recompile required | Replace .wasm file | Replace .so file |
| Debugging | Easier | Medium | Harder |
| Performance | Fastest | ~10-20% overhead | Fastest |
| Recommended | Yes | **Yes for Windows** | No |

## WebAssembly Plugins

WASM plugins work on all platforms including Windows, without CGO.

### Creating a WASM Plugin

```go
// plugin/fib.go
package main

import (
	"unsafe"
)

// Plugin name - exported for the host
//export plugin_name
func pluginName() (ptr uint32, size uint32) {
	name := "fib"
	return stringToPtr(name)
}

// Plugin version
//export plugin_version
func pluginVersion() (ptr uint32, size uint32) {
	return stringToPtr("1.0.0-wasm")
}

// Fast Fibonacci - O(n) time complexity
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

// Memory allocation for WASM
var memory []byte

//export alloc
func alloc(size uint32) uint32 {
	offset := uint32(len(memory))
	memory = append(memory, make([]byte, size)...)
	return offset
}

func stringToPtr(s string) (uint32, uint32) {
	ptr := alloc(uint32(len(s)))
	copy(memory[ptr:], s)
	return ptr, uint32(len(s))
}

func main() {}
```

### Building WASM Plugins

```bash
# Install TinyGo (required for WASM plugins)
# See: https://tinygo.org/getting-started/

# Build the plugin
tinygo build -o fib.wasm -target=wasi plugin/fib.go
```

### Loading WASM Plugins

```go
import "github.com/topxeq/xxlang/pkg/plugin"

loader := plugin.NewLoader()
loader.AddPath("./plugins")

// Loads fib.wasm from ./plugins/
p, err := loader.Load("fib")
if err != nil {
    panic(err)
}

// Register for use in Xxlang
plugin.Register(p)
```

### WASM Plugin Conventions

| Export Name | Purpose |
|-------------|---------|
| `plugin_name` | Returns (ptr, size) for plugin name |
| `plugin_version` | Returns (ptr, size) for version string |
| `call_<name>` | Callable function from Xxlang (e.g., `call_fast` → `fib.fast`) |
| `alloc` | Memory allocation function |

### WASM Function Signatures

WASM plugins support these function types:

| Signature | Example |
|-----------|---------|
| `int64 → int64` | `call_fast(n int64) int64` |
| `int64 → int32` | `call_isFib(n int64) int32` (boolean result: 0 or 1) |
| `int64 → (ptr, count)` | `call_range_(n int64) (uint32, uint32)` (returns array) |

## Complete Example

See [examples/fib_plugin/](../examples/fib_plugin/) for a complete working example:

```
examples/fib_plugin/
├── main.go           # Main program
└── plugin/
    └── fibplugin.go  # Plugin implementation
```

Run the example:

```bash
cd examples/fib_plugin
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
