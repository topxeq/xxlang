# Embedding Xxlang in Go Applications

## Overview

Xxlang can be embedded in Go applications to provide scripting capabilities. This guide covers:

- Basic evaluation
- Passing values between Go and Xxlang
- Module loading
- Error handling

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

### WithGlobals(globals)

Provides initial global variables.

```go
interp := interpreter.New(
    interpreter.WithStdlib(),
)
```

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
    interp.Eval(`println(name + " has " + count + " items")`)
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
interp.Eval(`for (n in numbers) { println(n) }`)
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
    import "std/math"
    println(math.sqrt(16))
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
    interp.Eval("println(x)")
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
