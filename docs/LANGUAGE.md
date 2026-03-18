# Xxlang Language Reference

## Overview

Xxlang is a bytecode VM-based scripting language implemented in Go. It features:
- Bytecode compilation and virtual machine execution
- First-class functions with closures
- Object-oriented programming with classes and inheritance
- Module system with import/export
- Rich standard library
- Embeddable in Go applications

## File Extension

Xxlang source files use the `.xxl` extension. Compiled bytecode files use `.xxb`.

## CLI Usage

```bash
# Start interactive REPL
xxlang

# Run a local source file
xxlang run script.xxl
xxlang script.xxl                     # Shortcut

# Run a script from URL
xxlang https://example.com/script.xxl
xxlang run https://example.com/script.xxl

# Run compiled bytecode
xxlang run script.xxb

# Compile to standalone executable
xxlang compile -o program script.xxl

# Compile to bytecode only
xxlang compile --bytecode script.xxl

# Show help
xxlang help

# Show version
xxlang version
```

## REPL Commands

| Command | Description |
|---------|-------------|
| `exit`, `quit` | Exit the REPL |
| `help` | Show help message |
| `history` | Show command history |
| `clear` | Clear all variables and functions |

## Types

### Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `INT` | 64-bit integer | `42`, `-5`, `0` |
| `FLOAT` | 64-bit float | `3.14`, `-2.5`, `1.0e10` |
| `STRING` | UTF-8 string | `"hello"`, `` `raw string` `` |
| `BOOL` | Boolean | `true`, `false` |
| `NULL` | Null value | `null` |

### Composite Types

| Type | Description | Example |
|------|-------------|---------|
| `ARRAY` | Ordered collection | `[1, 2, 3]` |
| `MAP` | Key-value pairs | `{"a": 1, "b": 2}` |
| `FUNCTION` | Function object | `func(x) { return x * 2 }` |
| `CLOSURE` | Function with captured variables | See closures section |
| `CLASS` | Class definition | `class Point { ... }` |
| `INSTANCE` | Class instance | `new Point(1, 2)` |
| `MODULE` | Imported module | `import "math"` |
| `BUILTIN` | Built-in function | `len`, `pln`, `typeOf` |

### Type Checking

```xxl
typeOf(42)        // "INT"
typeOf(3.14)      // "FLOAT"
typeOf("hello")   // "STRING"
typeOf(true)      // "BOOL"
typeOf(null)      // "NULL"
typeOf([1, 2])    // "ARRAY"
typeOf({"a": 1})  // "MAP"
typeOf(func() {}) // "FUNCTION"
```

## Strings

Xxlang supports two types of string literals:

### Double-Quoted Strings

Standard strings with escape sequence support:

```xxl
var s1 = "hello"
var s2 = "line1\nline2"       // \n = newline
var s3 = "tab\there"          // \t = tab
var s4 = "quote: \"hello\""   // \" = quote
```

**Escape sequences:**
- `\n` - Newline
- `\t` - Tab
- `\r` - Carriage return
- `\\` - Backslash
- `\"` - Double quote
- `\0` - Null character

### Raw Strings (Backtick)

Raw strings do not process escape sequences, making them ideal for:
- Multi-line text
- File paths (Windows)
- Regular expressions
- Text containing backslashes or quotes

```xxl
// Multi-line string
var text = `line1
line2
line3`

// Windows path (no escape processing)
var path = `C:\Users\test\file.txt`

// Text with quotes
var msg = `He said "hello" to me`

// Regex pattern
var pattern = `\d+\.\d+`
```

## Variables

### var - Mutable Variables

```xxl
var x = 10
var name = "Alice"
var arr = [1, 2, 3]
```

### const - Constants

```xxl
const PI = 3.14159
const MAX_SIZE = 100
```

Constants cannot be reassigned after declaration.

## Variable Scope

Xxlang uses lexical scoping with the following rules:

### Scope Hierarchy

Variables are resolved from innermost to outermost scope:

1. **Local scope** - Variables declared inside a function
2. **Enclosing function scope** - Variables from outer functions (for closures)
3. **Global scope** - Variables declared at the top level

### Global Variables

Variables declared at the top level are global and accessible everywhere:

```xxl
var globalVar = "global"

func getGlobal() {
    return globalVar  // Can read global
}

func setGlobal(val) {
    globalVar = val   // Can modify global
}
```

### Local Variables

Variables declared inside a function are local to that function:

```xxl
func myFunction() {
    var localVar = "local"  // Only accessible inside myFunction
    return localVar
}
```

### Variable Shadowing

A local variable can shadow (hide) an outer variable with the same name:

```xxl
var name = "global"

func shadowExample() {
    var name = "local"     // Shadows global 'name'
    pln(name)          // Prints "local"
}

shadowExample()
pln(name)              // Prints "global" (unchanged)
```

Function parameters also shadow outer variables:

```xxl
var x = "global"

func paramShadow(x) {      // Parameter shadows global 'x'
    return x
}

pln(paramShadow("arg"))  // "arg"
pln(x)                   // "global" (unchanged)
```

### Nested Functions and Closures

Functions can be nested. Inner functions can capture variables from outer functions:

```xxl
func outer() {
    var message = "Hello"

    func inner() {
        return message  // Captures 'message' from outer
    }

    return inner()
}

pln(outer())  // "Hello"
```

Inner local variables shadow outer variables:

```xxl
func outer() {
    var x = "outer"

    func inner() {
        var x = "inner"  // Shadows outer's x
        return x
    }

    return inner() + " " + x
}

pln(outer())  // "inner outer"
```

### Closures

Closures capture variables by reference, allowing the captured variable to be modified:

```xxl
func makeCounter() {
    var count = 0

    func counter() {
        count = count + 1  // Modifies captured variable
        return count
    }

    return counter
}

var c1 = makeCounter()
pln(c1())  // 1
pln(c1())  // 2
pln(c1())  // 3

var c2 = makeCounter()  // New closure, new captured variable
pln(c2())  // 1 (independent from c1)
```

### Important: Multiple Closures Sharing Variables

> ⚠️ **Known Behavior**: When multiple closures are created in the same scope and capture the same variable, each closure currently gets its own **copy** of the variable, rather than sharing a single reference.

#### Affected Patterns

**Pattern 1: Multiple closures returned in a map/object**

```xxl
func createObject() {
    var value = "initial"

    return {
        "set": func(newVal) {
            value = newVal      // Modifies its own copy
        },
        "get": func() {
            return value         // Returns its own copy
        }
    }
}

var obj = createObject()
obj["set"]("updated")
pln(obj["get"]())  // "initial" (NOT "updated" as expected!)
```

**Pattern 2: Multiple closures in an array**

```xxl
func createCounters() {
    var count = 0

    return [
        func() { count = count + 1; return count },
        func() { return count }
    ]
}

var counters = createCounters()
pln(counters[0]())  // 1
pln(counters[1]())  // 0 (NOT 1 as expected!)
```

#### Workaround: Use a Map as Shared State

To share state between multiple closures, use a map (which is a reference type):

```xxl
func createObject() {
    var state = {"value": "initial"}  // Map is a reference type

    return {
        "set": func(newVal) {
            state["value"] = newVal   // Modifies the shared map
        },
        "get": func() {
            return state["value"]      // Reads from the shared map
        }
    }
}

var obj = createObject()
obj["set"]("updated")
pln(obj["get"]())  // "updated" ✓ Works correctly!
```

#### Pattern Comparison Table

| Pattern | Expected Behavior | Current Behavior | Status |
|---------|-------------------|------------------|--------|
| Single closure captures variable | Variable shared across calls | ✅ Works correctly | OK |
| Multiple closures in separate calls | Each closure has its own variable | ✅ Works correctly | OK |
| Multiple closures in same return | All share same variable reference | ❌ Each has its own copy | **Known Issue** |
| Multiple closures + map workaround | Map is shared by reference | ✅ Works correctly | Use This |

### Built-in Function Shadowing

Local variables can shadow built-in functions:

```xxl
func example() {
    var len = 100       // Shadows built-in len()
    pln(len)        // 100
}

example()
pln(len([1,2,3]))   // 3 (built-in still works outside)
```

### Scope Resolution Summary

```xxl
var a = "global"

func outer() {
    var a = "outer local"
    var b = "outer only"

    func inner() {
        var a = "inner local"  // Shadows outer's a
        pln(a)              // "inner local"
        pln(b)              // "outer only" (captured)
    }

    inner()
    pln(a)              // "outer local"
}
```

### Best Practices

1. **Avoid unnecessary variable shadowing**: Using distinct variable names improves code readability
2. **Use maps for shared state**: When multiple closures need to share variables, use a map as the shared state container
3. **Single closures work normally**: For single closure scenarios (like counters), directly capturing variables works correctly
4. **Be aware of scope boundaries**: Both function parameters and local variables shadow outer variables with the same name

## Operators

### Arithmetic

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `1 + 2` → `3` |
| `-` | Subtraction | `5 - 3` → `2` |
| `*` | Multiplication | `4 * 3` → `12` |
| `/` | Division | `10 / 2` → `5` |
| `%` | Modulo | `10 % 3` → `1` |

### Comparison

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `1 == 1` → `true` |
| `!=` | Not equal | `1 != 2` → `true` |
| `<` | Less than | `1 < 2` → `true` |
| `>` | Greater than | `2 > 1` → `true` |
| `<=` | Less or equal | `1 <= 1` → `true` |
| `>=` | Greater or equal | `2 >= 2` → `true` |

### Logical

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | AND | `true && false` → `false` |
| `\|\|` | OR | `true \|\| false` → `true` |
| `!` | NOT | `!true` → `false` |

### Assignment

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Assign | `x = 10` |
| `+=` | Add and assign | `x += 5` |
| `-=` | Subtract and assign | `x -= 3` |
| `*=` | Multiply and assign | `x *= 2` |
| `/=` | Divide and assign | `x /= 2` |
| `%=` | Modulo and assign | `x %= 3` |

### Increment/Decrement

```xxl
var i = 0
i++     // i is now 1
i--     // i is now 0
```

### String Concatenation

```xxl
var greeting = "Hello" + " " + "World"  // "Hello World"
```

## Control Flow

### If-Else

```xxl
if (condition) {
    // code
}

if (condition) {
    // code
} else {
    // code
}

if (condition1) {
    // code
} else if (condition2) {
    // code
} else {
    // code
}
```

### While Loop

```xxl
var i = 0
while (i < 5) {
    pln(i)
    i++
}
```

### For Loop

```xxl
// C-style for loop
for (var i = 0; i < 5; i++) {
    pln(i)
}

// For-in loop (arrays)
for (item in [1, 2, 3]) {
    pln(item)
}

// For-in loop (maps)
for (key, value in {"a": 1, "b": 2}) {
    pln(key + ": " + value)
}
```

### Break and Continue

```xxl
for (var i = 0; i < 10; i++) {
    if (i == 5) {
        break      // Exit loop
    }
    if (i % 2 == 0) {
        continue   // Skip to next iteration
    }
    pln(i)
}
```

### Switch Statement

```xxl
switch (value) {
    case 1:
        pln("one")
    case 2:
        pln("two")
    default:
        pln("other")
}
```

## Functions

### Basic Function

```xxl
func add(a, b) {
    return a + b
}

pln(add(3, 4))  // 7
```

### Return Statement

```xxl
func greet(name) {
    if (name == "") {
        return "Hello, stranger!"
    }
    return "Hello, " + name + "!"
}
```

### Closures

Functions can capture variables from their enclosing scope:

```xxl
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
pln(counter())  // 3
```

> 📘 **See [Variable Scope](#variable-scope) for detailed closure behavior, including important notes about multiple closures sharing variables.**

### Immediately Invoked Function Expression (IIFE)

```xxl
var result = func(x) {
    return x * x
}(5)  // result is 25
```

### Recursion

```xxl
func fib(n) {
    if (n <= 1) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

pln(fib(10))  // 55
```

### Tail Call Optimization

Tail Call Optimization (TCO) is applied **automatically** when a function call is in tail position. This means the function call must be the **last operation** before returning.

#### When TCO Works (Automatic)

```xxl
// ✅ Direct tail call
func sumTail(n, acc) {
    if (n <= 0) {
        return acc
    }
    return sumTail(n - 1, acc + n)  // TCO applies
}

// ✅ Tail call in conditional branches
func factorial(n, acc) {
    if (n <= 1) {
        return acc
    } else {
        return factorial(n - 1, acc * n)  // TCO applies
    }
}

// ✅ Multiple tail calls
func fibHelper(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibHelper(n - 1, b, a + b)  // TCO applies
}

pln(sumTail(10000, 0))     // Works without stack overflow!
pln(factorial(1000, 1))    // Works!
pln(fibHelper(10000, 0, 1)) // Works!
```

#### When TCO Does NOT Apply

```xxl
// ❌ Not a tail call: result is used in addition
func fib(n) {
    if (n <= 1) { return n }
    return fib(n - 1) + fib(n - 2)  // No TCO - needs addition after calls
}

// ❌ Not a tail call: result stored in variable first
func sumBad(n, acc) {
    if (n <= 0) { return acc }
    var result = sumBad(n - 1, acc + n)  // No TCO
    return result
}

// ❌ Not a tail call: operation after the call
func bad(n, acc) {
    if (n <= 0) { return acc }
    return bad(n - 1, acc + n) * 1  // No TCO - multiplication after call
}
```

#### TCO Rule Summary

| Pattern | TCO | Reason |
|---------|-----|--------|
| `return func(args)` | ✅ Yes | Call is the last operation |
| `return a + func(args)` | ❌ No | Addition needed after call |
| `return func(args) + func(args)` | ❌ No | Addition needed after calls |
| `var x = func(args); return x` | ❌ No | Assignment happens first |
| `return func(args) * 1` | ❌ No | Multiplication after call |

#### How to Write Tail-Recursive Functions

1. Pass accumulated results as parameters
2. The recursive call must be the **entire return value**
3. No operations after the recursive call

```xxl
// Convert naive recursion to tail recursion:
// Naive (slow, no TCO):
func fibNaive(n) {
    if (n <= 1) { return n }
    return fibNaive(n - 1) + fibNaive(n - 2)  // No TCO
}

// Tail-recursive (fast, TCO applies):
func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)  // TCO!
}
func fib(n) { return fibTail(n, 0, 1) }
```

**Performance impact**: TCO makes recursion ~420,000x faster for fib(35)!

## Arrays

### Creation

```xxl
var arr1 = [1, 2, 3, 4, 5]
var arr2 = []               // Empty array
var arr3 = ["a", "b", "c"]
```

### Access

```xxl
var arr = [10, 20, 30]
pln(arr[0])   // 10
pln(arr[1])   // 20
```

### Modification

```xxl
var arr = [1, 2, 3]
arr[1] = 20       // arr is now [1, 20, 3]
```

### Built-in Functions

```xxl
len([1, 2, 3])           // 3
first([1, 2, 3])         // 1
last([1, 2, 3])          // 3
push([1, 2], 3)          // [1, 2, 3]
pop([1, 2, 3])           // [1, 2] (returns modified array)
sort([3, 1, 2])          // [1, 2, 3]
reverse([1, 2, 3])       // [3, 2, 1]
sum([1, 2, 3, 4, 5])     // 15
indexOf([1, 2, 3], 2)    // 1
containsArr([1, 2, 3], 2) // true
concat([1, 2], [3, 4])   // [1, 2, 3, 4]
```

### Array Methods

```xxl
var arr = [1, 2, 3]

arr.len()           // 3
arr.push(4)         // [1, 2, 3, 4]
arr.pop()           // [1, 2, 3] (returns modified array)
arr.first()         // 1
arr.last()          // 3
arr.indexOf(2)      // 1
arr.contains(2)     // true
arr.reverse()       // [3, 2, 1]
```

## Maps

### Creation

```xxl
var map1 = {"name": "Alice", "age": 30}
var map2 = {}                        // Empty map
```

### Access

```xxl
var person = {"name": "Alice", "age": 30}
pln(person["name"])   // "Alice"
pln(person["age"])    // 30
```

### Modification

```xxl
var person = {"name": "Alice"}
person["age"] = 30        // Add new key
person["name"] = "Bob"    // Modify existing
```

### Built-in Functions

```xxl
len({"a": 1, "b": 2})    // 2
keys({"a": 1, "b": 2})    // ["a", "b"]
values({"a": 1, "b": 2})  // [1, 2]
hasKey({"a": 1}, "a")     // true
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

### Map Methods

```xxl
var m = {"a": 1, "b": 2}

m.len()              // 2
m.keys()             // ["a", "b"]
m.values()           // [1, 2]
m.hasKey("a")        // true
m.delete("a")        // {"b": 2}
```

## Classes

### Basic Class

```xxl
class Point {
    func init(x, y) {
        this.x = x
        this.y = y
    }

    func add(other) {
        return new Point(this.x + other.x, this.y + other.y)
    }

    func toString() {
        return "(" + this.x + ", " + this.y + ")"
    }
}

var p1 = new Point(1, 2)
var p2 = new Point(3, 4)
var p3 = p1.add(p2)
pln(p3.toString())  // "(4, 6)"
```

### Inheritance

```xxl
class Animal {
    func init(name) {
        this.name = name
    }

    func speak() {
        return "..."
    }
}

class Dog : Animal {
    func init(name) {
        super.init(name)
    }

    func speak() {
        return this.name + " says Woof!"
    }
}

var dog = new Dog("Rex")
pln(dog.speak())  // "Rex says Woof!"
```

### Super Calls

```xxl
class Parent {
    func greet() {
        return "Hello from Parent"
    }
}

class Child : Parent {
    func greet() {
        return super.greet() + " and Child"
    }
}
```

## Modules

### Import Syntax

```xxl
// Import standard library
import "math"
pln(math.sqrt(16))

// Import with alias
import "io" as io
io.println("Hello")

// Import specific functions
import "string" { upper, lower }
pln(upper("hello"))

// Import from relative path
import * as utils from "./utils"
pln(utils.add(5, 3))

// Import from absolute path
import * as math from "/home/user/project/math.xxl"

// Import WASM plugin by path
import * as fib from "./plugins/fib.wasm"
pln(fib.fast(50))
```

### Export Syntax

```xxl
// Export variable
export var PI = 3.14159

// Export function
export func add(a, b) {
    return a + b
}

// Export class
export class Point {
    func init(x, y) {
        this.x = x
        this.y = y
    }
}

// Export existing binding
var secret = "hidden"
export secret
```

### Module Resolution

Module paths are resolved based on prefix and file extension:

| Path Format | Description | Example |
|-------------|-------------|---------|
| `xxx` | Standard library module | `import "math"` |
| `plugin/xxx` | WASM plugin by name | `import "plugin/fib"` |
| `*.wasm` | WASM plugin by file path | `import * as fib from "./fib.wasm"` |
| `./xxx` or `../xxx` | Relative path module | `import * as utils from "./utils"` |
| `/xxx` | Absolute path module | `import * as math from "/home/user/math.xxl"` |

**Note:** For file paths, `.wasm` files are loaded as WASM plugins; other paths have `.xxl` extension auto-added.

## Error Handling

Xxlang provides comprehensive exception handling with `try`, `catch`, `finally`, and `throw`.

### Basic Try-Catch

```xxl
try {
    var result = riskyOperation()
    pln(result)
} catch (err) {
    pln("Caught error: ", err)
}
```

### Try-Finally (without catch)

Use `finally` alone for cleanup code that must run regardless of success or failure:

```xxl
var file = openFile("data.txt")
try {
    processFile(file)
} finally {
    file.close()  // Always executes
}
```

### Full Try-Catch-Finally

```xxl
try {
    var result = divide(10, 0)
    pln(result)
} catch (err) {
    pln("Error: ", err)
} finally {
    pln("Cleanup completed")
}
```

**Execution order:**
1. If no error: try block → finally block
2. If error occurs: try block (until error) → catch block → finally block

### Throw Statement

```xxl
func divide(a, b) {
    if (b == 0) {
        throw "division by zero"
    }
    return a / b
}

func validateAge(age) {
    if (age < 0) {
        throw "age cannot be negative"
    }
    if (age > 150) {
        throw "age seems invalid: " + age
    }
    return true
}
```

You can throw any value (string, number, object):

```xxl
throw "error message"      // String
throw 404                  // Number
throw {"code": 500}        // Map
```

### Nested Try-Catch

Exception handlers can be nested. Each `try` creates its own exception handler:

```xxl
try {
    try {
        throw "inner error"
    } catch (e) {
        pln("Inner caught: ", e)
    }
} catch (e) {
    pln("Outer caught: ", e)  // Not executed
}
// Output: Inner caught: inner error
```

### Re-throwing Exceptions

Catch and re-throw to add context:

```xxl
try {
    try {
        throw "file not found"
    } catch (e) {
        throw "config error: " + e  // Re-throw with context
    }
} catch (e) {
    pln(e)  // Output: config error: file not found
}
```

### Finally Overrides Exception

If `finally` throws an exception, it replaces the original:

```xxl
try {
    try {
        throw "first error"
    } finally {
        throw "second error"  // Replaces "first error"
    }
} catch (e) {
    pln(e)  // Output: second error
}
```

### Resource Cleanup Pattern

Use `finally` for guaranteed cleanup:

```xxl
func processFile(filename) {
    var file = openFile(filename)
    var success = false
    try {
        var data = file.read()
        processData(data)
        success = true
    } catch (err) {
        pln("Processing failed: ", err)
    } finally {
        file.close()
        if (!success) {
            deleteTempFiles()
        }
    }
}
```

### WithFile Pattern (Recommended for Resource Management)

Encapsulate resource management in a helper function:

```xxl
func withFile(filename, callback) {
    var file = openFile(filename)
    try {
        return callback(file)
    } finally {
        file.close()
    }
}

// Usage - automatic cleanup guaranteed
withFile("data.txt", func(f) {
    var content = f.read()
    pln(content)
    // file.close() is automatically called
})
```

### Multiple Resources

Chain `withFile` for multiple resources:

```xxl
func withFile(filename, callback) {
    var file = openFile(filename)
    try {
        return callback(file)
    } finally {
        file.close()
    }
}

// Multiple files with automatic cleanup
withFile("input.txt", func(input) {
    withFile("output.txt", func(output) {
        var data = input.read()
        output.write(process(data))
    })
})
```

### Exception Propagation

Uncaught exceptions propagate up the call stack:

```xxl
func level3() {
    throw "error at level 3"
}

func level2() {
    level3()  // No catch, exception propagates
}

func level1() {
    try {
        level2()
    } catch (e) {
        pln("Caught at level1: ", e)
    }
}

level1()  // Output: Caught at level1: error at level 3
```

### Summary Table

| Clause | Required? | Purpose |
|--------|-----------|---------|
| `try` | Yes | Contains code that might throw |
| `catch (var)` | Optional* | Handles the exception |
| `finally` | Optional* | Cleanup code (always runs) |

*At least one of `catch` or `finally` must be present.

### Best Practices

1. **Use specific error messages**: Include context in thrown errors
   ```xxl
   throw "file '" + filename + "' not found"
   ```

2. **Clean up in finally**: Always release resources in `finally`
   ```xxl
   try { ... } finally { resource.close() }
   ```

3. **Use withFile pattern**: Encapsulate resource management
   ```xxl
   withFile("data.txt", func(f) { /* use f */ })
   ```

4. **Don't catch everything silently**: Log or re-throw errors
   ```xxl
   catch (e) {
       log(e)
       throw e  // Re-throw if you can't handle it
   }
   ```

## Standard Library

Xxlang includes several standard library modules:

| Module | Description |
|--------|-------------|
| `io` | Input/output operations |
| `string` | String manipulation |
| `math` | Mathematical functions |
| `array` | Array utilities |
| `json` | JSON parsing and encoding |
| `regex` | Regular expressions |
| `crypto` | Cryptographic functions |

See [STDLIB.md](STDLIB.md) for complete documentation.

## Primitive Type Methods

All types have universal methods:

```xxl
42.typeOf()      // "INT"
3.14.typeOf()    // "FLOAT"
"hello".typeOf() // "STRING"

42.toStr()       // "42"
3.14.toStr()     // "3.14"
```

### Integer Methods

```xxl
42.toFloat()     // 42.0
(-5).abs()       // 5
```

### Float Methods

```xxl
3.14.toInt()     // 3
(-2.5).abs()     // 2.5
3.7.floor()      // 3
3.2.ceil()       // 4
3.5.round()      // 4
```

### String Methods

```xxl
"hello".len()           // 5
"hello".upper()         // "HELLO"
"HELLO".lower()         // "hello"
"  hello  ".trim()      // "hello"
"a,b,c".split(",")      // ["a", "b", "c"]
"hello".contains("ell") // true
"hello".indexOf("l")    // 2
"hello".startsWith("he") // true
"hello".endsWith("lo")  // true
"42".toInt()            // 42
"3.14".toFloat()        // 3.14
```

## Comments

```xxl
// Single line comment

/*
   Multi-line
   comment
*/
```

## Keywords

```
var const func return if else while for in break continue
class new this super true false null import export
try catch finally throw switch case default
```

## Performance Considerations

1. **Bytecode Compilation**: Code is compiled to bytecode before execution
2. **Tail Call Optimization**: Tail-recursive functions use constant stack space
3. **Closure Optimization**: Closures only capture variables they actually use
4. **Inline Caching**: Method lookups are cached where possible

## Embedding in Go

See [EMBEDDING.md](EMBEDDING.md) for guide on embedding Xxlang in Go applications.
