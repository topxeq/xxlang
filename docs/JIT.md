# JIT Compilation

Xxlang features a pure Go JIT (Just-In-Time) compiler that generates native x86-64 machine code for high-performance execution. The JIT compiler requires no CGO dependencies and achieves near-native performance for compute-intensive workloads.

## Platform Support

| Platform | JIT Support | Memory Allocation | Calling Convention |
|----------|-------------|-------------------|-------------------|
| Linux/amd64 | ✅ Full support | mmap | System V AMD64 ABI |
| Darwin/amd64 | ✅ Full support | mmap | System V AMD64 ABI |
| **Windows/amd64** | ✅ **Full support** | VirtualAlloc | Microsoft x64 ABI |
| Linux/arm64 | ⚠️ Interpreter only | - | - |
| Darwin/arm64 | ⚠️ Interpreter only | - | - |
| Windows/arm64 | ⚠️ Interpreter only | - | - |

**Note**: JIT is disabled by default. Enable it with `--jit` flag for compute-intensive workloads.

## CLI Usage

### Enable JIT

```bash
# Run with JIT enabled
xxl --jit script.xxl

# JIT with custom hot path threshold
xxl --jit --jit-threshold=10 script.xxl

# JIT with debug output
xxl --jit --jit-debug script.xxl

# Explicitly disable JIT (default)
xxl --no-jit script.xxl
```

### Debug Mode

The `--debug` flag provides comprehensive debug output:

```bash
xxl --debug script.xxl
xxl --debug --jit script.xxl
```

Output includes:
- Source file path and size
- Bytecode instruction count
- Number of constants
- Compile time
- JIT enabled status
- VM mode (Interpreter/JIT hybrid)
- Execution time
- JIT statistics (native vs interpreter executions)
- Total time

Example output:
```
[Debug] Source: /path/to/script.xxl
[Debug] Source size: 94 bytes
[Debug] Bytecode instructions: 31
[Debug] Constants: 5
[Debug] Compile time: 320.73µs
[Debug] JIT enabled: true
[Debug] VM mode: JIT (hybrid)
[Debug] Execution time: 245.767µs
[Debug] Native executions: 1
[Debug] Interpreter executions: 1
[Debug] Total time: 566.497µs
```

## Performance Highlights

### Cross-Language Fibonacci Benchmark (March 2026)

| Test | C | Java | Go | Python | Xxlang JIT | Xxlang VM |
|------|---|------|-----|--------|------------|-----------|
| fib(35) recursive | 45 ms | 47 ms | 52 ms | 2.7 s | **54 ms** | 5.02 s |
| fib(35) iterative | 19 ns | 21 ns | 21 ns | 837 ns | **23 ns** | 1.5 µs |
| fib(20) recursive | 315 ns | 330 ns | 315 ns | 17.8 µs | **334 ns** | 30 µs |
| fib(15) recursive | 28 ns | 30 ns | 28 ns | 1.5 µs | **30 ns** | 1.2 µs |

**Key insights:**
- **JIT matches native Go/Java**: Only 4-10% slower for recursive algorithms
- **JIT vs Python**: 50x faster for recursive Fibonacci
- **JIT vs VM**: 93x faster than the bytecode interpreter for recursive calls
- **Windows JIT**: Same performance as Linux/macOS

### Platform Comparison

The JIT performs identically across all supported platforms (Linux, macOS, Windows) on the same hardware:

```
fib(35) with JIT on Intel i7:
  Linux:   54 ms
  macOS:   54 ms
  Windows: 54 ms
```

## Embedded API

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    // Create interpreter with JIT enabled
    interp := interpreter.New(
        interpreter.WithStdlib(),
        interpreter.WithJIT(),
    )

    // Measure execution time
    start := time.Now()
    result, err := interp.Eval(`
        func fib(n) {
            if (n <= 1) { return n }
            return fib(n - 1) + fib(n - 2)
        }
        fib(35)
    `)
    elapsed := time.Since(start)

    if err != nil {
        panic(err)
    }

    fmt.Printf("Result: %s\n", result.Inspect())
    fmt.Printf("Execution time: %v\n", elapsed)
    fmt.Printf("JIT enabled: %v\n", interp.JITEnabled())
}
```

### JIT Configuration Options

```go
// Enable JIT with default settings
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithJIT(),
)

// Enable JIT with custom configuration
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithJITConfig(interpreter.JITConfig{
        Enabled:      true,
        HotThreshold: 10,   // Compile after 10 calls
        MaxCodeSize:  8192, // Max bytecode size for JIT
        Debug:        false,
    }),
)

// Use individual options
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithJIT(),
    interpreter.WithJITThreshold(10),
    interpreter.WithJITDebug(),
)
```

### Runtime Control

```go
// Check if JIT is enabled
if interp.JITEnabled() {
    fmt.Println("JIT is enabled")
}

// Enable/disable JIT at runtime
interp.SetJITEnabled(true)
interp.SetJITEnabled(false)

// Get/set full config
config := interp.GetJITConfig()
config.HotThreshold = 50
interp.SetJITConfig(config)
```

## Architecture

### Native Support Levels

The JIT compiler analyzes bytecode to determine the optimal execution strategy:

| Level | Operations | Callback Required |
|-------|------------|-------------------|
| **SupportPureArithmetic** | Arithmetic, comparisons, control flow, locals/globals | No |
| **SupportWithBuiltins** | Above + builtin function calls | Yes (builtin) |
| **SupportWithCalls** | Above + function calls (OpRegCall, OpRegTailCall) | Yes (function) |
| **SupportWithArrays** | Above + array/map operations | Yes (collection) |
| **SupportWithObjects** | Above + object field access and methods | Yes (object) |
| **SupportNone** | Closures, classes, exceptions | Falls back to interpreter |

### Callback Mechanism

The JIT implements a callback system allowing native x86-64 code to call back to Go functions:

```
┌─────────────────┐      Platform ABI      ┌──────────────────┐
│  Native Code    │ ──────────────────────►│ Assembly Bridge  │
│  (x86-64)       │   (System V or MS x64) │ (bridge_*.s)     │
└─────────────────┘                        └────────┬─────────┘
                                                    │
                                           Go calling convention
                                                    ▼
                                           ┌──────────────────┐
                                           │  Go Callback     │
                                           │  (native_executor.go) │
                                           └──────────────────┘
```

Four callback types are supported:

1. **Builtin Callback** - Execute builtin functions from native code
2. **Function Callback** - Dispatch function calls to other compiled or interpreted functions
3. **Collection Callback** - Array and map operations (create, index, append, set)
4. **Object Callback** - Field access and method invocation

### Function Argument Support

The JIT supports functions with up to 8 arguments:

```go
// Native function calling convention
// System V ABI (Linux/macOS): RDI, RSI, RDX, RCX, R8, R9
// Microsoft x64 ABI (Windows): RCX, RDX, R8, R9
func Execute(globals []int64, args ...int64) int64
```

For functions with more than 8 arguments, the system falls back to globals-only passing.

## Supported Operations

### Fully Supported (No Callbacks Needed)

| Category | Operations |
|----------|------------|
| **Data Movement** | LoadConst, Move, LoadLocal, StoreLocal, LoadGlobal, StoreGlobal |
| **Arithmetic** | Add, Sub, Mul, Div, Mod, Neg |
| **Comparison** | Less, Greater, Equal, NotEqual, LessEqual, GreaterEqual |
| **Logical** | And, Or, Not |
| **Control Flow** | Jump, JumpIfTrue, JumpIfFalse, Return |
| **Constants** | Null, True, False |
| **Loop Optimizations** | LoopCountAdd, LoopBodyAdd, LoopIncCheck |

### Supported with Callbacks

| Category | Operations | Callback Type |
|----------|------------|---------------|
| **Builtins** | OpRegBuiltin | Builtin callback |
| **Function Calls** | OpRegCall, OpRegTailCall | Function callback |
| **Arrays** | Array, ArrayEmpty, ArrayAppend, Index, SetIndex | Collection callback |
| **Maps** | Map, MapEmpty, MapSet | Collection callback |
| **Objects** | GetField, SetField, GetMethod, CallMethod | Object callback |

### Interpreter Fallback

Functions containing these operations fall back to the bytecode interpreter:

- Closures (OpRegClosure, OpRegLoadFree, OpRegStoreFree)
- Classes (OpRegClass, OpRegNew)
- Exception handling (OpRegThrow, OpRegPushHandler, OpRegPopHandler)
- Module loading (OpRegLoadModule)

## Tail Call Optimization

The JIT implements proper tail call optimization for recursive functions:

```javascript
// Tail-recursive Fibonacci - O(n) time, O(1) space
func fibHelper(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibHelper(n - 1, b, a + b)  // Tail call - reuses stack frame
}

func fib(n) {
    return fibHelper(n, 0, 1)
}

fib(1000)  // Executes instantly without stack overflow!
```

For native-compiled functions, tail calls compile to a direct jump to the function entry point, eliminating stack growth.

## Implementation Files

| File | Purpose |
|------|---------|
| `pkg/jit/jit.go` | Core JIT compiler infrastructure (Unix) |
| `pkg/jit/jit_windows_amd64.go` | Core JIT compiler infrastructure (Windows) |
| `pkg/jit/native_codegen.go` | Pure native x86-64 code generator |
| `pkg/jit/native_executor.go` | Native execution engine and callbacks |
| `pkg/jit/bridge/bridge_amd64.s` | Assembly bridge (System V ABI) |
| `pkg/jit/bridge/bridge_windows_amd64.s` | Assembly bridge (Microsoft x64 ABI) |
| `pkg/jit/bridge/mem_unix.go` | Memory allocation (mmap) |
| `pkg/jit/bridge/mem_windows.go` | Memory allocation (VirtualAlloc) |
| `pkg/jit/jit_vm.go` | JIT-enabled VM wrapper |
| `pkg/jit/jit_recursive.go` | Recursive function handling |

## Memory Management

### Executable Memory

The JIT allocates executable memory pages using platform-specific APIs:

```go
// Linux/macOS - mmap with PROT_READ|PROT_WRITE|PROT_EXEC
prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
flags := syscall.MAP_ANON | syscall.MAP_PRIVATE
data, err := syscall.Mmap(-1, 0, size, prot, flags)

// Windows - VirtualAlloc with PAGE_EXECUTE_READWRITE
mem, err := bridge.AllocExecMem(size)
```

### Cleanup

Always call `Cleanup()` to release allocated memory:

```go
executor := jit.NewNativeExecutor(config)
defer executor.Cleanup()  // Important!
```

## Debugging

Enable debug output to see generated code:

```go
config := jit.JITConfig{
    Debug: true,
}
```

Or use CLI:

```bash
xxl --jit --jit-debug script.xxl
xxl --debug --jit script.xxl
```

This prints:
- Generated code size
- First bytes of native code (hex dump)
- Function compilation details
- Callback pointer addresses
- Native vs interpreter execution counts

## Limitations

1. **Platform Support**: x86-64 only (Linux, macOS, Windows); ARM64 falls back to interpreter
2. **Closures**: Functions with closures fall back to interpreter
3. **Floating Point**: Limited support (integers are optimized)
4. **Strings**: String operations require callbacks

## Future Work

1. **ARM64 Support**: Extend code generation for Apple Silicon and ARM servers
2. **SIMD**: Use AVX/SSE for array operations
3. **Inline Caching**: Speed up property/method lookups
4. **Escape Analysis**: Reduce heap allocations
5. **Tiered Compilation**: Interpret first, compile hot paths

## Testing

Run the JIT tests:

```bash
# Run all JIT tests
go test ./pkg/jit/... -v

# Run with race detector
go test ./pkg/jit/... -race

# Run specific test
go test ./pkg/jit/... -run TestNativeExecutor
```

---

*Last updated: 2026-03-21*
