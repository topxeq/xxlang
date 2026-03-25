# Embedding Xxlang in Go Applications

## Overview

Xxlang can be embedded in Go applications to provide scripting capabilities. This guide covers:

- Basic evaluation
- Passing values between Go and Xxlang
- Module loading
- Error handling
- High-performance Go function integration

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration Options](#configuration-options)
- [Passing Values from Go to Xxlang](#passing-values-from-go-to-xxlang)
- [Getting Values from Xxlang](#getting-values-from-xxlang)
- [Type Conversion Helpers](#type-conversion-helpers)
- [Calling Xxlang Functions from Go](#calling-xxlang-functions-from-go)
- [Error Handling](#error-handling)
- [Module Loading](#module-loading)
- [Complete Example](#complete-example)
- [Best Practices](#best-practices)
- [High Performance via Go Functions](#high-performance-via-go-functions)
- [Thread Safety](#thread-safety)

## Installation

```bash
go get github.com/topxeq/xxlang/pkg/interpreter
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    // Create interpreter with stdlib enabled
    interp := interpreter.New(interpreter.WithStdlib())

    // Evaluate simple expression
    result, err := interp.Eval("2 + 2 * 3")
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Inspect()) // 8
}
```

## Configuration Options

### WithStdlib()

Enables all standard library modules (io, string, math, json, etc.).

```go
interp := interpreter.New(interpreter.WithStdlib())
```

### WithVMType(vmType) / WithStackVM()

Selects which VM to use. By default, Xxlang uses the **register-based VM** for best performance.

```go
// Default: Register VM (recommended for best performance)
interp := interpreter.New(interpreter.WithStdlib())

// Explicitly use Register VM
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithVMType(interpreter.RegisterVM),
)

// Use Stack VM (for compatibility or complex function usage)
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithStackVM(),
    // or: interpreter.WithVMType(interpreter.StackVM)
)
```

**Performance comparison:**

| Benchmark | Stack VM | Register VM | Speedup |
|-----------|----------|-------------|---------|
| Fibonacci(15) | 971 µs | 52 µs | **18.6x** |
| Factorial | 400 µs | 52 µs | **7.7x** |
| Function Calls | 534 µs | 78 µs | **6.8x** |
| Loop(1000) | 687 µs | 267 µs | **2.6x** |

**Note**: The register VM has limited support for user-defined functions. For complex function usage, use the stack VM instead.

### WithGlobals(globals)

Provides initial global variables.

```go
interp := interpreter.New(
    interpreter.WithStdlib(),
)
```

### WithJIT() / WithJITConfig()

Enables JIT (Just-In-Time) compilation for compute-intensive workloads.

**Note: JIT is disabled by default.** Enable it only when you need maximum performance for numerical algorithms.

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

// Or use individual options
interp := interpreter.New(
    interpreter.WithStdlib(),
    interpreter.WithJIT(),
    interpreter.WithJITThreshold(10),
    interpreter.WithJITDebug(),
)
```

**Runtime control:**

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

**JIT Performance:**

| Benchmark | Interpreter | JIT | Speedup |
|-----------|-------------|-----|---------|
| fib(35) recursive | ~5 seconds | 54 ms | **93x** |
| fib(35) iterative | 1.5 µs | 23 ns | **65x** |

**JIT Platform Support:**

| Platform | JIT Support |
|----------|-------------|
| Linux/amd64 | ✅ Full support |
| Darwin/amd64 | ✅ Full support |
| **Windows/amd64** | ✅ **Full support** |
| arm64 (all) | ⚠️ Interpreter only |

**When to use JIT:**
- Compute-intensive numerical algorithms
- Recursive functions (especially with TCO)
- Long-running scripts with hot loops

**When NOT to use JIT:**
- Simple scripts (overhead not worth it)
- Code with closures (falls back to interpreter)
- Code with classes (falls back to interpreter)
- I/O-bound scripts (JIT won't help)

### Debug and Statistics

For performance analysis and debugging, you can measure execution time and check JIT statistics:

```go
package main

import (
    "fmt"
    "time"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    interp := interpreter.New(
        interpreter.WithStdlib(),
        interpreter.WithJIT(),
        interpreter.WithJITDebug(), // Enable JIT debug output
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

**CLI Debug Mode:**

When using the `xxl` command-line tool, the `--debug` flag provides comprehensive debug output:

```bash
xxl --debug script.xxl
xxl --debug --jit script.xxl
```

Output includes:
- Source file path and size
- Bytecode instruction count
- Number of constants
- Compile time
- JIT status and VM mode
- Execution time
- JIT statistics (native vs interpreter executions)
- Total time

## Passing Values from Go to Xxlang

### SetGlobal(name, value)

Sets a global variable in the interpreter.

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Pass primitive types
    interp.SetGlobal("count", 42)
    interp.SetGlobal("price", 19.99)
    interp.SetGlobal("name", "Alice")
    interp.SetGlobal("active", true)

    // Use in Xxlang
    interp.Eval(`pln(name + " has " + count + " items")`)
    // Output: Alice has 42 items
}
```

### Type Conversions

Go types are automatically converted:

| Go Type | Xxlang Type |
|---------|-------------|
| int, int64 | INT |
| float64 | FLOAT |
| string | STRING |
| bool | BOOL |
| nil | NULL |
| []interface{} | ARRAY |
| map[string]interface{} | MAP |

## Getting Values from Xxlang

### GetGlobal(name)

Gets a global variable value.

```go
// Set a value in Xxlang
interp.Eval("var result = 42")

// Get it back in Go
val, ok := interp.GetGlobal("result")
if ok {
    if intVal, ok := val.(*objects.Int); ok {
        fmt.Println(intVal.Value) // 42
    }
}
```

### GetGlobalAs(name)

Gets a global variable as a Go value.

```go
interp.Eval("var data = [1, 2, 3, 4, 5]")

if val, ok := interp.GetGlobalAs("data"); ok {
    if arr, ok := val.([]interface{}); ok {
        fmt.Println(arr) // [1 2 3 4 5]
    }
}
```

### LastPopped()

Gets the last evaluated value.

```go
result, err := interp.Eval("2 + 2")
if err != nil {
    panic(err)
}
fmt.Println(result.Inspect()) // 4
```

## Type Conversion Helpers

### ToGo(obj)

Converts Xxlang object to Go value.

```go
import "github.com/topxeq/xxlang/pkg/interpreter"

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Create an Xxlang object
    interp.Eval(`var data = {"name": "Bob", "age": 25}`)

    // Get the object
    obj, _ := interp.GetGlobal("data")

    // Convert to Go
    goVal := interpreter.ToGo(obj)
    if m, ok := goVal.(map[string]interface{}); ok {
        fmt.Println(m["name"]) // Bob
    }
}
```

### FromGo(value)

Converts Go value to Xxlang object.

```go
// Create a Go slice
goSlice := []interface{}{1, 2, 3, 4, 5}

// Convert to Xxlang object
obj, err := interpreter.FromGo(goSlice)
if err != nil {
    panic(err)
}

// Use in interpreter
interp.SetGlobal("numbers", obj)
interp.Eval(`for (n in numbers) { pln(n) }`)
```

## Calling Xxlang Functions from Go

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Define a function in Xxlang
    interp.Eval(`
        func greet(name) {
            return "Hello, " + name + "!"
        }
    `)

    // Call the function
    result, err := interp.Eval(`greet("World")`)
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Inspect()) // "Hello, World!"
}
```

## Error Handling

### Compile Errors

```go
_, err := interp.Eval("var x = ")
if err != nil {
    fmt.Println("Compile error:", err)
    // Output: Compile error: parser errors:
    //   line 1:9: expected next token to be IDENT, got EOF instead
}
```

### Runtime Errors

```go
_, err := interp.Eval("var x = 1 / 0")
if err != nil {
    fmt.Println("Runtime error:", err)
    // Output: Runtime error: division by zero
}
```

### Custom Error Handling

```go
func safeEval(interp *interpreter.Interpreter, code string) (objects.Object, error) {
    result, err := interp.Eval(code)
    if err != nil {
        // Log and return nil
        log.Printf("Evaluation failed: %v", err)
        return nil, err
    }
    return result, nil
}
```

## Module Loading

Xxlang supports module imports. When embedding, you can:

### Using Standard Library

```go
interp := interpreter.New(interpreter.WithStdlib())

// Now Xxlang can import stdlib modules
interp.Eval(`
    import "math"
    pln(math.sqrt(16))
`)
```

### Custom Module Resolution

You can provide custom module loading by implementing a module resolver.

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    fmt.Println("=== Xxlang Embedding Example ===")
    fmt.Println()

    // Create interpreter with stdlib
    interp := interpreter.New(interpreter.WithStdlib())

    // 1. Simple evaluation
    fmt.Println("1. Simple evaluation:")
    result, _ := interp.Eval("2 + 2 * 3")
    fmt.Printf("   2 + 2 * 3 = %s\n", result.Inspect())
    fmt.Println()

    // 2. Pass values from Go
    fmt.Println("2. Pass values from Go:")
    interp.SetGlobal("x", 42)
    interp.Eval("pln(x)")
    fmt.Println()

    // 3. Get values back
    fmt.Println("3. Get values back:")
    interp.Eval("var result = x * 2")
    val, _ := interp.GetGlobal("result")
    fmt.Printf("   result = %s\n", val.Inspect())
    fmt.Println()

    // 4. Complex data structures
    fmt.Println("4. Complex data structures:")
    interp.Eval(`
        var person = {
            "name": "Alice",
            "age": 30,
            "skills": ["Go", "Python", "Xxlang"]
        }
    `)

    if obj, ok := interp.GetGlobalAs("person"); ok {
        fmt.Printf("   %v\n", obj)
    }
    fmt.Println()

    // 5. Functions and closures
    fmt.Println("5. Functions and closures:")
    interp.Eval(`
        func makeCounter() {
            var count = 0
            func() {
                count = count + 1
                return count
            }
        }

        var counter = makeCounter()
    `)

    for i := 0; i < 3; i++ {
        result, _ = interp.Eval("counter()")
        fmt.Printf("   counter() = %s\n", result.Inspect())
    }
    fmt.Println()

    // 6. Error handling
    fmt.Println("6. Error handling:")
    _, err := interp.Eval("var x = 1 / 0")
    if err != nil {
        fmt.Printf("   Caught error: %v\n", err)
    }
    fmt.Println()

    fmt.Println("=== Embedding Example Complete ===")
}
```

## Best Practices

### 1. Reuse Interpreter

Create the interpreter once and reuse it for multiple evaluations:

```go
// Good: Reuse interpreter
interp := interpreter.New(interpreter.WithStdlib())
for _, script := range scripts {
    interp.Eval(script)
}

// Bad: Create new interpreter each time
for _, script := range scripts {
    interp := interpreter.New(interpreter.WithStdlib())
    interp.Eval(script)
}
```

### 2. Error Handling

Always check for errors:

```go
result, err := interp.Eval(code)
if err != nil {
    // Handle error appropriately
    return err
}
```

### 3. Type Assertions

Use type assertions when working with Xxlang objects:

```go
if intVal, ok := result.(*objects.Int); ok {
    // Use intVal.Value
} else if strVal, ok := result.(*objects.String); ok {
    // Use strVal.Value
}
```

### 4. Context Management

For long-running scripts, consider implementing timeout handling:

```go
// Implement timeout at your application level
done := make(chan objects.Object, 1)
errCh := make(chan error, 1)

go func() {
    result, err := interp.Eval(longRunningScript)
    if err != nil {
        errCh <- err
    } else {
        done <- result
    }
}()

select {
case result := <-done:
    // Success
case err := <-errCh:
    // Error
case <-time.After(5 * time.Second):
    // Timeout
}
```

## High Performance via Go Functions

Xxlang can call Go functions directly, achieving native performance for compute-intensive tasks. This is the recommended approach when performance matters.

### Performance Comparison

| Execution Mode | fib(35) Time | Reason |
|----------------|--------------|--------|
| Xxlang VM naive recursion | ~5.0 seconds | Interpreter overhead |
| Xxlang JIT naive recursion | **54 ms** | Native x86-64 execution |
| Xxlang JIT tail recursion | ~200 µs | Native + TCO optimization |
| Go function call | **~25 µs** | Native execution, 260,000x faster |
| C native | 45 ms | Baseline |

### Using the JIT Compiler Programmatically

The JIT compiler can be used directly for compute-intensive workloads:

```go
import "github.com/topxeq/xxlang/pkg/jit"

func main() {
    // Create JIT executor
    config := jit.JITConfig{
        HotThreshold: 10,  // Compile after 10 calls
        MaxCodeSize:  65536,
        Debug:        false,
    }
    executor := jit.NewNativeExecutor(config)
    defer executor.Cleanup()

    // Check if a function can be compiled
    if jit.CanExecuteNatively(compiledFn) {
        result, err := executor.ExecuteFunction(compiledFn, constants, globals)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Result: %d\n", result)
    }
}
```

### Native Support Levels

The JIT compiler categorizes functions by their complexity:

```go
level := jit.AnalyzeNativeSupport(compiledFn)

switch level {
case jit.SupportPureArithmetic:
    // Fastest - pure arithmetic/control flow, no callbacks
case jit.SupportWithBuiltins:
    // Uses builtin callback for built-in functions
case jit.SupportWithCalls:
    // Uses function callback for inter-function calls
case jit.SupportWithArrays:
    // Uses collection callback for array/map operations
case jit.SupportWithObjects:
    // Uses object callback for field/method access
case jit.SupportNone:
    // Falls back to interpreter (closures, classes)
}
```

### Why This Works

| Execution Mode | fib(35) Time | Reason |
|----------------|--------------|--------|
| Xxlang naive recursion | ~6.5 seconds | Interpreter overhead |
| Xxlang tail recursion (TCO) | ~200 µs | Optimized, but still interpreted |
| Go function call | **~25 µs** | Native execution, 260,000x faster |

### Registering Go Functions

Use `SetGlobal` with a `*objects.Builtin` to register Go functions:

```go
package main

import (
    "github.com/topxeq/xxlang/pkg/interpreter"
    "github.com/topxeq/xxlang/pkg/objects"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Register a high-performance Go function
    interp.SetGlobal("goFib", &objects.Builtin{
        Fn: func(args ...objects.Object) objects.Object {
            if len(args) != 1 {
                return &objects.Error{Message: "goFib requires 1 argument"}
            }

            n, ok := args[0].(*objects.Int)
            if !ok {
                return &objects.Error{Message: "argument must be integer"}
            }

            // Go-native implementation - extremely fast
            result := fibFast(n.Value)
            return &objects.Int{Value: result}
        },
    })

    // Now Xxlang can call it
    result, _ := interp.Eval("goFib(100)")
    println(result.Inspect())  // Instant!
}

// Go-native Fibonacci - O(n) time, O(1) space
func fibFast(n int64) int64 {
    if n <= 1 {
        return n
    }
    var a, b int64 = 0, 1
    for i := int64(2); i <= n; i++ {
        a, b = b, a+b
    }
    return b
}
```

### Batch Processing

Return arrays from Go to process multiple values at once:

```go
// Register a batch function
interp.SetGlobal("goFibBatch", &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        n := args[0].(*objects.Int).Value

        // Compute all values in Go
        results := make([]objects.Object, n+1)
        for i := int64(0); i <= n; i++ {
            results[i] = &objects.Int{Value: fibFast(i)}
        }

        // Return array to Xxlang
        return &objects.Array{Elements: results}
    },
})

// Xxlang usage
interp.Eval(`
    var fibs = goFibBatch(1000)  // One call, 1000 results
    println(fibs[100])           // Access any result
`)
```

### Use Cases

| Scenario | Approach | Performance |
|----------|----------|-------------|
| Simple logic, configuration | Xxlang code | Fast enough |
| Recursive algorithms | Tail recursion (TCO) | Good |
| Numerical computation | Go function | Excellent |
| Image/matrix processing | Go function | Excellent |
| Batch data processing | Go batch function | Excellent |

### Type Conversions in Builtin Functions

```go
interp.SetGlobal("processArray", &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        // Get array argument
        arr, ok := args[0].(*objects.Array)
        if !ok {
            return &objects.Error{Message: "argument must be array"}
        }

        // Convert to Go slice for processing
        goSlice := make([]int64, len(arr.Elements))
        for i, elem := range arr.Elements {
            goSlice[i] = elem.(*objects.Int).Value
        }

        // Process in Go
        result := processInGo(goSlice)

        // Convert back to Xxlang
        return &objects.Int{Value: result}
    },
})
```

### Best Practices

1. **Keep Xxlang for glue logic** - configuration, orchestration, simple operations
2. **Use Go for compute** - numerical algorithms, data processing, heavy lifting
3. **Batch when possible** - one Go call with multiple results is faster than many calls
4. **Handle errors gracefully** - return `*objects.Error` for invalid inputs

## Thread Safety

The interpreter is NOT thread-safe. Create separate interpreters for concurrent use, or use synchronization:

```go
var mu sync.Mutex
var interp = interpreter.New(interpreter.WithStdlib())

func evalSafe(code string) (objects.Object, error) {
    mu.Lock()
    defer mu.Unlock()
    return interp.Eval(code)
}
```

## Common Embedding Scenarios

### 1. Configuration Script Engine

Use Xxlang as a configuration language with computed values:

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Load configuration script
    configScript := `
        var config = {
            "host": "localhost",
            "port": 8080,
            "maxConnections": 100,
            "timeout": 30,
            "debug": true
        }

        // Computed configuration
        var effectiveTimeout = config.timeout * 2
        var connectionString = config.host + ":" + config.port
    `

    _, err := interp.Eval(configScript)
    if err != nil {
        panic(err)
    }

    // Get configuration values
    if cfg, ok := interp.GetGlobalAs("config"); ok {
        if m, ok := cfg.(map[string]interface{}); ok {
            fmt.Printf("Host: %v\n", m["host"])
            fmt.Printf("Port: %v\n", m["port"])
            fmt.Printf("Debug: %v\n", m["debug"])
        }
    }
}
```

### 2. Plugin System for Go Applications

Allow users to extend your application with scripts:

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
    "github.com/topxeq/xxlang/pkg/objects"
)

// Application context exposed to scripts
type AppContext struct {
    Name    string
    Version string
}

func main() {
    interp := interpreter.New(interpreter.WithStdlib())

    // Expose application API to scripts
    interp.SetGlobal("app", &objects.Builtin{
        Fn: func(args ...objects.Object) objects.Object {
            return &objects.String{Value: "MyApp v1.0"}
        },
    })

    // Expose logging function
    interp.SetGlobal("log", &objects.Builtin{
        Fn: func(args ...objects.Object) objects.Object {
            if len(args) > 0 {
                fmt.Println("[PLUGIN]", args[0].Inspect())
            }
            return &objects.Null{}
        },
    })

    // User plugin script
    pluginScript := `
        func onInit() {
            log("Plugin initialized")
        }

        func processData(data) {
            log("Processing: " + data)
            return data * 2
        }

        onInit()
    `

    interp.Eval(pluginScript)

    // Call plugin function
    result, _ := interp.Eval(`processData(42)`)
    fmt.Println("Result:", result.Inspect())
}
```

### 3. Template Engine

Use Xxlang for dynamic content generation:

```go
package main

import (
    "fmt"
    "strings"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func renderTemplate(template string, data map[string]interface{}) (string, error) {
    interp := interpreter.New(interpreter.WithStdlib())

    // Pass data to interpreter
    for k, v := range data {
        interp.SetGlobal(k, v)
    }

    // Execute template
    result, err := interp.Eval(template)
    if err != nil {
        return "", err
    }

    return result.Inspect(), nil
}

func main() {
    template := `"Hello, " + name + "! You have " + count + " messages."`

    data := map[string]interface{}{
        "name":  "Alice",
        "count": 5,
    }

    output, err := renderTemplate(template, data)
    if err != nil {
        panic(err)
    }

    fmt.Println(output) // Hello, Alice! You have 5 messages.
}
```

### 4. Rule Engine

Implement business rules in scripts:

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
    "github.com/topxeq/xxlang/pkg/objects"
)

type Order struct {
    Amount   float64
    Items    int
    Customer string
    Region   string
}

func applyRules(order Order) (float64, []string) {
    interp := interpreter.New(interpreter.WithStdlib())

    // Pass order data
    interp.SetGlobal("amount", order.Amount)
    interp.SetGlobal("items", order.Items)
    interp.SetGlobal("customer", order.Customer)
    interp.SetGlobal("region", order.Region)

    // Define discount rules
    rulesScript := `
        var discounts = []
        var totalDiscount = 0

        // Bulk discount
        if (items >= 10) {
            totalDiscount = totalDiscount + amount * 0.1
            discounts = push(discounts, "Bulk discount: 10%")
        }

        // VIP customer
        if (customer == "VIP") {
            totalDiscount = totalDiscount + amount * 0.05
            discounts = push(discounts, "VIP discount: 5%")
        }

        // Regional promotion
        if (region == "US") {
            totalDiscount = totalDiscount + 10
            discounts = push(discounts, "Regional promo: $10")
        }

        [totalDiscount, discounts]
    `

    result, _ := interp.Eval(rulesScript)

    // Extract results
    if arr, ok := result.(*objects.Array); ok {
        discount := arr.Elements[0].(*objects.Float).Value
        var applied []string
        for _, d := range arr.Elements[1].(*objects.Array).Elements {
            applied = append(applied, d.(*objects.String).Value)
        }
        return discount, applied
    }

    return 0, nil
}

func main() {
    order := Order{
        Amount:   500,
        Items:    12,
        Customer: "VIP",
        Region:   "US",
    }

    discount, rules := applyRules(order)
    fmt.Printf("Order: $%.2f\n", order.Amount)
    fmt.Printf("Discount: $%.2f\n", discount)
    fmt.Println("Applied rules:")
    for _, r := range rules {
        fmt.Println("  -", r)
    }
}
```

### 5. Data Transformation Pipeline

Use Xxlang for ETL-like data processing:

```go
package main

import (
    "fmt"
    "github.com/topxeq/xxlang/pkg/interpreter"
)

func transformData(data []map[string]interface{}, transformScript string) ([]map[string]interface{}, error) {
    interp := interpreter.New(interpreter.WithStdlib())

    // Convert data to Xxlang array
    interp.Eval("var data = []")
    for _, item := range data {
        for k, v := range item {
            interp.SetGlobal("_temp", v)
            interp.Eval("var _item = " + k + " + _temp") // Simplified
        }
    }

    // Apply transformation
    result, err := interp.Eval(transformScript)
    if err != nil {
        return nil, err
    }

    // Convert back to Go
    if arr, ok := result.(*objects.Array); ok {
        var output []map[string]interface{}
        for _, elem := range arr.Elements {
            if m, ok := elem.(*objects.Map); ok {
                item := make(map[string]interface{})
                for k, v := range m.Value {
                    item[k] = interpreter.ToGo(v)
                }
                output = append(output, item)
            }
        }
        return output, nil
    }

    return nil, nil
}

func main() {
    data := []map[string]interface{}{
        {"name": "Alice", "score": 85},
        {"name": "Bob", "score": 92},
    }

    script := `
        var result = []
        for (item in data) {
            var transformed = {
                "name": item.name,
                "grade": item.score >= 90 ? "A" : "B"
            }
            result = push(result, transformed)
        }
        result
    `

    output, err := transformData(data, script)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Transformed: %v\n", output)
}
```

### 6. Web Server Script Handler

Embed Xxlang in a web server for dynamic request handling:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/topxeq/xxlang/pkg/interpreter"
    "github.com/topxeq/xxlang/pkg/objects"
)

func scriptHandler(script string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        interp := interpreter.New(interpreter.WithStdlib())

        // Pass HTTP context
        interp.SetGlobal("method", r.Method)
        interp.SetGlobal("path", r.URL.Path)
        interp.SetGlobal("query", r.URL.Query().Get("q"))

        // Provide response writer
        interp.SetGlobal("respond", &objects.Builtin{
            Fn: func(args ...objects.Object) objects.Object {
                if len(args) > 0 {
                    w.Write([]byte(args[0].Inspect()))
                }
                return &objects.Null{}
            },
        })

        // Execute script
        result, err := interp.Eval(script)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Use result if respond wasn't called
        if result != nil && result.Type() != "NULL" {
            w.Write([]byte(result.Inspect()))
        }
    }
}

func main() {
    // Dynamic endpoint
    http.HandleFunc("/api/hello", scriptHandler(`
        var name = query || "World"
        respond("Hello, " + name + "!")
    `))

    http.HandleFunc("/api/calc", scriptHandler(`
        var a = int(query)
        respond("Result: " + (a * 2))
    `))

    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```
