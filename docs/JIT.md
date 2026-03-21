# JIT Compilation

Xxlang features a pure Go JIT (Just-In-Time) compiler that generates native x86-64 machine code for high-performance execution. The JIT compiler requires no CGO dependencies and achieves near-native performance for compute-intensive workloads.

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
┌─────────────────┐      System V ABI      ┌──────────────────┐
│  Native Code    │ ──────────────────────►│ Assembly Bridge  │
│  (x86-64)       │                        │ (bridge_amd64.s) │
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
// Native function calling convention (System V AMD64 ABI)
// Arguments passed in: RAX, RBX, RCX, RDX, R8, R9, R10, R11
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

## Object Handle Pooling

The JIT uses an object handle pooling system to manage objects created during native execution:

```go
type JITCallbackContext struct {
    objects     []objects.Object  // Objects created by callbacks
    freeHandles []int             // Reusable handle indices
}
```

When an object is created (array, map, etc.), a handle is allocated. Handles can be reused after release, preventing unbounded memory growth during long-running computations.

### Thread Safety

All callback operations use `sync.RWMutex` for thread-safe access:

```go
globalJITContext.mu.Lock()
defer globalJITContext.mu.Unlock()
```

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
| `pkg/jit/jit.go` | Core JIT compiler infrastructure |
| `pkg/jit/native_codegen.go` | Pure native x86-64 code generator |
| `pkg/jit/native_executor.go` | Native execution engine and callbacks |
| `pkg/jit/bridge_amd64.s` | Assembly bridge (System V ABI to Go) |
| `pkg/jit/jit_vm.go` | JIT-enabled VM wrapper |
| `pkg/jit/jit_recursive.go` | Recursive function handling |

## Usage

### CLI

```bash
# Run with JIT enabled
xxl --jit script.xxl

# For recursive algorithms, TCO is automatic
xxl fib_tco.xxl
```

### Programmatic API

```go
import "github.com/topxeq/xxlang/pkg/jit"

// Create native executor
config := jit.JITConfig{
    HotThreshold: 10,
    MaxCodeSize:  65536,
    Debug:        false,
}
executor := jit.NewNativeExecutor(config)
defer executor.Cleanup()

// Execute a function
result, err := executor.ExecuteFunction(fn, constants, globals)
```

### Checking Native Support

```go
// Check if a function can run natively
if jit.CanExecuteNatively(fn) {
    // Pure arithmetic - fastest execution
}

// Check support level
level := jit.AnalyzeNativeSupport(fn)
switch level {
case jit.SupportPureArithmetic:
    // No callbacks needed
case jit.SupportWithBuiltins:
    // Uses builtin callbacks
case jit.SupportWithCalls:
    // Uses function dispatch callbacks
}
```

## Memory Management

### Executable Memory

The JIT allocates executable memory pages using `mmap` (Linux/macOS) or `VirtualAlloc` (Windows):

```go
mem, page, err := compiler.AllocCode(codeSize)
copy(mem, code)
// mem is now executable
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

This prints:
- Generated code size
- First bytes of native code (hex dump)
- Function compilation details
- Callback pointer addresses

## Limitations

1. **Platform Support**: Currently x86-64 only (Linux, macOS, Windows)
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
