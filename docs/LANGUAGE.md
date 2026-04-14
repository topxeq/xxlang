# Xxlang Language Reference

## Overview

Xxlang is a bytecode VM-based scripting language implemented in Go. It features:
- Bytecode compilation and virtual machine execution
- First-class functions with closures
- Object-oriented programming with classes and inheritance
- Concurrency with goroutines, tubes (channels), and select
- Module system with import/export
- Rich standard library
- Embeddable in Go applications

## Table of Contents

- [CLI Usage](#cli-usage)
- [REPL Commands](#repl-commands)
- [Types](#types)
- [Strings](#strings)
- [Variables](#variables)
- [Variable Scope](#variable-scope)
- [Operators](#operators)
- [Control Flow](#control-flow)
- [Functions](#functions)
- [Arrays](#arrays)
- [Maps](#maps)
- [Classes](#classes)
- [Modules](#modules)
- [Error Handling](#error-handling)
- [Concurrency](#concurrency)
- [Standard Library](#standard-library)
- [Primitive Type Methods](#primitive-type-methods)
- [Comments](#comments)
- [Keywords](#keywords)
- [Performance Considerations](#performance-considerations)
- [Embedding in Go](#embedding-in-go)

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

# Run with JIT enabled (for compute-intensive workloads)
xxlang --jit script.xxl

# Run with debug output
xxlang --debug script.xxl
xxlang --debug --jit script.xxl

# Show help
xxlang help

# Show version
xxlang version
```

### CLI Options

| Option | Description |
|--------|-------------|
| `--jit` | Enable JIT compilation for hot paths |
| `--jit-threshold=N` | Set JIT hot path threshold (default: 100) |
| `--jit-debug` | Enable JIT-specific debug output |
| `--no-jit` | Explicitly disable JIT (default) |
| `--debug` | Show debug info (bytecode count, runtime, JIT usage) |
| `-o, --output path` | Output path for compiled file |
| `--target os/arch` | Cross-compile for target OS/architecture |
| `--bytecode` | Output as bytecode (.xxb) instead of executable |

### Debug Mode

The `--debug` flag provides comprehensive debug output:

```
[Debug] Source: script.xxl
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
| `BIGINT` | Arbitrary precision integer | `12345678901234567890n` |
| `FLOAT` | 64-bit float | `3.14`, `-2.5`, `1.0e10` |
| `BIGFLOAT` | Arbitrary precision float | `3.14159265358979323846m` |
| `STRING` | UTF-8 string | `"hello"`, `` `raw string` `` |
| `BOOL` | Boolean | `true`, `false` |
| `NULL` | Null value | `null` |

### Null Type

In Xxlang, `null` is the sole null value representing "no value" or "undefined". This design simplifies JavaScript's dual `null`/`undefined` concept.

**When null is returned:**
- Function without explicit return
- Accessing non-existent map key
- Array out-of-bounds access
- Function execution failure (when no error is needed)

```xxl
// Null value
var nothing = null
typeOf(null)      // "NULL"

// Null from various operations
var arr = [1, 2, 3]
first([])         // null (empty array)
arr[10]           // null (out of bounds)
{"a": 1}["b"]     // null (non-existent key)

// Checking for null
isNull(null)      // true
isUndefined(null) // true (alias for isNull)
isUndef(null)     // true (shorthand for isUndefined)
```

**Comparison with JavaScript:**

| Scenario | JavaScript | Xxlang |
|----------|------------|--------|
| Null value | `null` | `null` |
| Undefined | `undefined` | `null` |
| Uninitialized variable | `undefined` | `null` |
| Non-existent property | `undefined` | `null` |
| Function with no return | `undefined` | `null` |

This unified approach eliminates the confusion between `null` and `undefined` found in JavaScript.

**Note: `null` vs `nil`**

Xxlang only has `null`, not `nil`. This is different from Go:

| Aspect | Go | Xxlang |
|--------|-----|--------|
| Null keyword | `nil` | `null` |
| Usage | Pointers, interfaces, slices, maps, channels | Single null type |
| In Xxlang code | N/A | Use `null` only |

```xxl
// ✅ Correct in Xxlang
var x = null
if (x == null) { ... }

// ❌ Wrong - 'nil' does not exist in Xxlang
var y = nil       // Error!
```

The `nil` in Go is the zero value for pointer types, while Xxlang's `null` is a first-class value representing "no value".

### BigInt and BigFloat

Xxlang supports arbitrary precision numbers for calculations requiring high precision:

**BigInt** - Use the `n` suffix for arbitrary precision integers:

```xxl
// Regular int has limits (64-bit)
var small = 9223372036854775807  // Max int64

// BigInt has no practical limit
var big = 1234567890123456789012345678901234567890n
var huge = 10n ^ 100n  // 10^100

// BigInt arithmetic
var a = 1000000000000000000n
var b = 2000000000000000000n
pln(a + b)  // 3000000000000000000

// Type checking
typeOf(big)  // "BIGINT"
```

**BigFloat** - Use the `m` suffix for arbitrary precision decimals:

```xxl
// Regular float has precision limits
var approx = 0.1 + 0.2  // May have rounding errors

// BigFloat maintains exact precision
var precise = 0.1m + 0.2m  // Exactly 0.3

// High precision calculations
var pi = 3.141592653589793238462643383279m
var e = 2.718281828459045235360287471352m

// Financial calculations
var price = 19.99m
var tax = price * 0.08m
var total = price + tax

typeOf(pi)  // "BIGFLOAT"
```

**When to use:**
- **BigInt**: Cryptography, factorial calculations, IDs larger than int64
- **BigFloat**: Financial calculations, scientific computing, avoiding floating-point errors

### Composite Types

| Type | Description | Example |
|------|-------------|---------|
| `ARRAY` | Ordered collection | `[1, 2, 3]` |
| `MAP` | Key-value pairs | `{"a": 1, "b": 2}` |
| `CHARS` | Unicode character array | `toChars("中文")` |
| `FUNCTION` | Function object | `func(x) { return x * 2 }` |
| `CLOSURE` | Function with captured variables | See closures section |
| `CLASS` | Class definition | `class Point { ... }` |
| `INSTANCE` | Class instance | `new Point(1, 2)` |
| `MODULE` | Imported module | `import "math"` |
| `BUILTIN` | Built-in function | `len`, `pln`, `typeOf` |

### Chars Type

The `chars` type provides proper Unicode character handling. Strings in Xxlang are character-oriented (rune-based), and chars provides additional character-level operations like slicing and iteration.

```xxl
// String is character-oriented
var s = "中文"
pln(len(s))     // 2 (characters)
pln(s[0])       // "中" (character at index 0)

// Chars provides additional character-level operations
var c = toChars("中文")
pln(len(c))     // 2 (characters)
pln(c[0])       // "中" (full character)
pln(c[1])       // "文" (full character)

// Character slicing
var c2 = toChars("Hello世界🎉")
pln(c2.subStr(5, 7).toStr())  // "世界"
pln(c2[7])                     // "🎉"

// Chars methods
pln(c2.upper().toStr())        // "HELLO世界🎉"
pln(c2.reverse().toStr())      // "🎉界世olleH"
pln(c2.contains("世界"))        // true
```

**When to use chars:**
- Counting characters in Unicode text
- Extracting individual Unicode characters
- Character-based string manipulation
- Processing text with Chinese, Japanese, Korean, emoji, etc.

See [STDLIB.md](STDLIB.md#chars) for complete chars documentation.

### Type Checking

```xxl
typeOf(42)        // "INT"
typeOf(123n)      // "BIGINT"
typeOf(3.14)      // "FLOAT"
typeOf(3.14m)     // "BIGFLOAT"
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

### := Short Variable Declaration

Xxlang supports Go-style short variable declaration:

```xxl
x := 10
name := "Alice"
arr := [1, 2, 3]
```

The `:=` syntax is equivalent to `var` and creates a new variable in the current scope:

```xxl
// These are equivalent:
var a = 10
a := 10

// Works with any expression:
result := add(3, 4)
items := [1, 2, 3]
config := {"debug": true, "name": "test"}
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
| `:=` | Short variable declaration | `x := 10` |
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

### Ternary Operator

The ternary operator `?:` provides a concise way to write simple conditional expressions:

```xxl
var age = 20
var status = age >= 18 ? "adult" : "minor"
pln(status)  // "adult"

// Equivalent to:
var status2
if (age >= 18) {
    status2 = "adult"
} else {
    status2 = "minor"
}
```

**Syntax:**

```xxl
condition ? valueIfTrue : valueIfFalse
```

**Examples:**

```xxl
// Simple value selection
var max = a > b ? a : b

// String interpolation
var count = 1
var msg = count == 1 ? "1 item" : count + " items"

// Nested ternary (use sparingly for readability)
var score = 85
var grade = score >= 90 ? "A" : score >= 80 ? "B" : score >= 70 ? "C" : "F"

// With function calls
var result = isValid ? processData(data) : getDefault()

// In expressions
pln("Found " + (found ? "yes" : "no"))
```

**Best Practices:**

1. **Keep it simple**: Use ternary for simple value selection, not complex logic.
2. **Avoid deep nesting**: Nested ternaries are hard to read. Use `switch (true)` or if-else instead.
3. **Use parentheses for clarity**: When embedding in larger expressions, parentheses improve readability.

```xxl
// Good: Simple and clear
var discount = isMember ? 0.1 : 0

// Bad: Too complex for ternary
var result = a > b ? c > d ? e : f : g > h ? i : j  // Hard to read

// Better: Use switch or if-else
var result
switch (true) {
    case a > b && c > d:
        result = e
    case a > b:
        result = f
    case g > h:
        result = i
    default:
        result = j
}
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

The `switch` statement provides a clean way to dispatch based on a value. Unlike C/Java, Xxlang's switch **does not fall through** - execution stops after the matching case block.

#### Basic Syntax

```xxl
switch (expression) {
    case value1:
        // statements
    case value2:
        // statements
    default:
        // statements (optional)
}
```

#### Simple Example

```xxl
var day = 3
switch (day) {
    case 1:
        pln("Monday")
    case 2:
        pln("Tuesday")
    case 3:
        pln("Wednesday")
    case 4:
        pln("Thursday")
    case 5:
        pln("Friday")
    default:
        pln("Weekend")
}
// Output: Wednesday
```

#### No Fall-Through

Unlike C or Java, Xxlang's switch does **not fall through** to the next case. Each case block terminates automatically:

```xxl
var num = 1
switch (num) {
    case 1:
        pln("One")      // Executes this
    case 2:
        pln("Two")      // NOT executed (no fall-through)
    default:
        pln("Default")  // NOT executed
}
// Output: One
```

This means you don't need explicit `break` statements - the behavior is automatic.

#### Switch with Strings

```xxl
var fruit = "apple"
switch (fruit) {
    case "apple":
        pln("Red or Green")
    case "banana":
        pln("Yellow")
    case "orange":
        pln("Orange")
    default:
        pln("Unknown fruit")
}
// Output: Red or Green
```

#### Switch with Expressions

The switch expression can be any expression:

```xxl
var x = 10
var y = 20
switch (x + y) {
    case 10:
        pln("Sum is 10")
    case 20:
        pln("Sum is 20")
    case 30:
        pln("Sum is 30")
    default:
        pln("Sum is something else")
}
// Output: Sum is 30
```

#### Conditional Switch Pattern

A powerful pattern is using `switch (true)` with conditional expressions in case statements. This provides a cleaner alternative to if-else chains for range checks and complex conditions:

```xxl
var score = 85
switch (true) {
    case score >= 90:
        pln("Grade: A")
    case score >= 80:
        pln("Grade: B")
    case score >= 70:
        pln("Grade: C")
    case score >= 60:
        pln("Grade: D")
    default:
        pln("Grade: F")
}
// Output: Grade: B
```

Each case expression is evaluated in order until one matches `true`. This works because:

1. The switch expression is `true`
2. Each case evaluates a boolean expression
3. The first case that evaluates to `true` matches

This pattern is especially useful for:

```xxl
// Range comparisons
var age = 25
switch (true) {
    case age < 13:
        pln("Child")
    case age < 20:
        pln("Teenager")
    case age < 60:
        pln("Adult")
    default:
        pln("Senior")
}

// Multiple condition checks
var x = 10
var y = 20
switch (true) {
    case x < 0 || y < 0:
        pln("Negative value")
    case x > 100 && y > 100:
        pln("Both large")
    case x == y:
        pln("Equal")
    default:
        pln("Other case")
}
```

**Note**: Cases are evaluated in order. Place more specific conditions before general ones:

```xxl
var value = 15
switch (true) {
    case value > 20:    // Checked first
        pln("Large")
    case value > 10:    // Checked second - matches here
        pln("Medium")
    case value > 0:     // Never reached for value=15
        pln("Positive")
    default:
        pln("Non-positive")
}
// Output: Medium
```

#### Multiple Statements in Case

Each case can contain multiple statements:

```xxl
var val = 2
switch (val) {
    case 1:
        pln("Case 1")
        pln("Multiple lines work")
    case 2:
        pln("Case 2")
        var temp = 100
        pln("temp = " + temp.toStr())
    default:
        pln("Default case")
}
// Output:
// Case 2
// temp = 100
```

#### Default Clause

The `default` clause is optional and executes when no case matches:

```xxl
var n = 999
switch (n) {
    case 1:
        pln("One")
    default:
        pln("Not one, it's " + n.toStr())
}
// Output: Not one, it's 999
```

#### Switch vs If-Else Chain

The switch statement is cleaner than a long if-else chain:

```xxl
// Using switch (recommended for multiple discrete values)
switch (status) {
    case 200:
        pln("OK")
    case 404:
        pln("Not Found")
    case 500:
        pln("Server Error")
    default:
        pln("Unknown status")
}

// Equivalent if-else chain (more verbose)
if (status == 200) {
    pln("OK")
} else if (status == 404) {
    pln("Not Found")
} else if (status == 500) {
    pln("Server Error")
} else {
    pln("Unknown status")
}
```

#### Comparison with Other Languages

| Feature | Xxlang | C/Java | Go |
|---------|--------|--------|-----|
| Fall-through | No (auto break) | Yes (needs break) | No (auto break) |
| Default case | Optional | Optional | Optional |
| String matching | Yes | Yes (Java 7+) | Yes |
| Expression in switch | Yes | Yes | Yes |
| Multiple values per case | No | No | Yes |

#### Best Practices

1. **Use switch for discrete values**: When comparing against multiple specific values, switch is clearer than if-else chains.

2. **Always include default**: Handle unexpected values with a default clause.

   ```xxl
   switch (command) {
       case "start":
           startService()
       case "stop":
           stopService()
       default:
           pln("Unknown command: " + command)
   }
   ```

3. **Use conditional switch for range comparisons**: Use `switch (true)` pattern for cleaner range comparisons instead of if-else chains.

   ```xxl
   // Use conditional switch for ranges (cleaner than if-else)
   switch (true) {
       case score >= 90:
           grade = "A"
       case score >= 80:
           grade = "B"
       case score >= 70:
           grade = "C"
       default:
           grade = "F"
   }
   ```

4. **Order cases by frequency or logic**: Place commonly matched cases first for better readability and slight performance benefit.

   ```xxl
   switch (httpStatus) {
       case 200:   // Most common - check first
           handleSuccess()
       case 404:   // Common client error
           handleNotFound()
       case 500:   // Server error
           handleServerError()
       default:
           handleUnknown()
   }
   ```

5. **Each case must have its own body**: Unlike some languages, Xxlang does not support multiple case values falling through to shared code. Each case executes independently.

   ```xxl
   // NOT supported - multiple cases to same block
   // switch (x) {
   //     case 1:
   //     case 2:
   //     case 3:
   //         handleSmall()  // This does NOT work!
   // }

   // Correct approach - use conditional switch pattern instead
   switch (true) {
       case x == 1 || x == 2 || x == 3:
           handleSmall()
       case x >= 10:
           handleLarge()
       default:
           handleOther()
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

### Variadic Functions

Variadic functions can accept a variable number of arguments using the `...` syntax:

```xxl
// Basic variadic function
func sum(...args) {
    var total = 0
    for (x in args) {
        total = total + x
    }
    return total
}

pln(sum())           // 0
pln(sum(1))          // 1
pln(sum(1, 2, 3))    // 6
pln(sum(1, 2, 3, 4, 5))  // 15
```

**Mixed Parameters:**

Regular parameters must come before the variadic parameter:

```xxl
// Required parameters + variadic
func formatList(prefix, ...items) {
    var result = prefix + ": "
    for (i, item in items) {
        if (i > 0) {
            result = result + ", "
        }
        result = result + toStr(item)
    }
    return result
}

pln(formatList("Fruits", "apple", "banana", "orange"))
// Output: Fruits: apple, banana, orange
```

**Variadic Parameter is an Array:**

The variadic parameter becomes an array inside the function:

```xxl
func logAll(...messages) {
    // 'messages' is an array
    pln("Count:", len(messages))
    for (msg in messages) {
        pln("-", msg)
    }
}

logAll("Hello", "World", 42, true)
// Count: 4
// - Hello
// - World
// - 42
// - true
```

**Common Use Cases:**

```xxl
// Flexible string formatting
func sprintf(format, ...args) {
    var result = format
    for (arg in args) {
        result = replaceFirst(result, "{}", toStr(arg))
    }
    return result
}

pln(sprintf("Hello {}, you have {} messages", "Alice", 5))
// Output: Hello Alice, you have 5 messages

// Math operations on variable arguments
func max(...nums) {
    if (len(nums) == 0) {
        return null
    }
    var result = nums[0]
    for (num in nums) {
        if (num > result) {
            result = num
        }
    }
    return result
}

pln(max(3, 1, 4, 1, 5, 9, 2, 6))  // 9

// Array-like construction
func list(...items) {
    return items  // Simply return the variadic array
}

var myArray = list(1, 2, 3, 4, 5)
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

#### Method Tail Call Optimization

TCO also works for recursive method calls:

```xxl
class Counter {
    func countDown(n) {
        if (n <= 0) {
            return "done"
        }
        return this.countDown(n - 1)  // TCO applies!
    }
}

var c = new Counter()
pln(c.countDown(10000))  // Works without stack overflow!
```

**Supported method TCO patterns:**
- `return this.method(args)` - ✅ TCO applies
- `return self.method(args)` - ✅ TCO applies (when `self` is bound to `this`)
- `return obj.method(args)` - ❌ No TCO (different object)

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

## OrderedMaps

OrderedMap is a map that preserves the insertion order of key-value pairs. Unlike regular Maps, OrderedMaps maintain the order in which keys were added.

### Creation

```xxl
// Create empty OrderedMap
var om = newOrderedMap()

// Create from Map
var om2 = newOrderedMap({"a": 1, "b": 2})

// Create from Array of [key, value] pairs
var om3 = newOrderedMap([["x", 10], ["y", 20]])

// Create with capacity
var om4 = make("orderedMap", 100)
```

### Access and Modification

```xxl
var om = newOrderedMap()
om["name"] = "Alice"    // Add entry
om["age"] = 30
om["city"] = "Beijing"

pln(om["name"])         // "Alice"
pln(om.keys())          // ["name", "age", "city"] - order preserved
pln(om.values())        // ["Alice", 30, "Beijing"]
```

### OrderedMap Methods

```xxl
var om = newOrderedMap()
om["a"] = 1
om["b"] = 2
om["c"] = 3

// Basic operations
om.len()                // 3
om.hasKey("a")          // true
var newOm = om.delete("b")  // Returns new OrderedMap without "b"

// Ordered access
om.keys()               // ["a", "b", "c"]
om.values()             // [1, 2, 3]
om.entries()            // [["a", 1], ["b", 2], ["c", 3]]
om.indexOf("b")         // 1
om.getAt(0)             // ["a", 1]

// Reordering
om.moveToFront("c")     // ["c", "a", "b"]
om.moveToBack("a")      // ["c", "b", "a"]
om.swap("c", "a")       // ["a", "b", "c"]
om.reverse()            // ["c", "b", "a"]
om.sortByKey()          // ["a", "b", "c"]

// Conversion
om.toMap()              // Convert to regular Map
om.clone()              // Create a copy
```

### Use Cases

OrderedMaps are particularly useful for:

- Database query results where column order matters
- JSON serialization with predictable key order
- Configuration files where key order should be preserved
- Building ordered data structures

```xxl
// Database query with ordered columns
var rows = dbQueryOrdered(db, "SELECT id, name, email FROM users")
for (row in rows) {
    pln(row.keys())     // ["id", "name", "email"] - query column order
}
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
import "strings" { toUpper, toLower }
pln(toUpper("hello"))

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

### Name Conflict Resolution

When a module function has the same name as a built-in function, Xxlang resolves the conflict based on the import style:

#### Name Resolution Priority

Xxlang resolves names in the following order (highest to lowest priority):

1. **User-defined variables** (local, then enclosing, then global)
2. **Destructured imports** (`import { fn } from "module"`)
3. **Built-in functions**

This means:
- A local variable `abs` will shadow both built-in `abs` and imported `abs`
- A destructured import `import { abs } from "math"` will shadow built-in `abs`
- Namespace imports (`import * as m from "module"`) do NOT participate in shadowing

#### 1. Destructuring Import Overwrites Built-in

When using destructuring import (`import { func } from "module"`), the imported function **overwrites** the built-in:

```xxl
// Built-in abs function
pln(abs(-5))      // 5 (built-in)

// Destructuring import overwrites built-in
import { abs } from "math"
pln(abs(-10))     // 10 (module function, not built-in)
```

#### 2. Namespace Import Avoids Conflict (Recommended)

Using namespace import (`import * as name from "module"`) keeps both functions accessible:

```xxl
// Namespace import - no conflict
import * as math from "math"

// Built-in still works
pln(abs(-5))         // 5 (built-in)

// Module function via namespace
pln(math.abs(-10))   // 10 (module function)
```

#### 3. User Variables Override Built-ins

User-defined variables always take precedence over built-in functions:

```xxl
// Custom abs function
abs := func(x) { x * 2 }
pln(abs(5))        // 10 (custom function)
```

#### 4. Preserving Built-in Reference

To use both built-in and module functions with the same name, save the built-in first:

```xxl
// Save built-in reference
builtinAbs := abs

// Import module function
import { abs } from "math"

// Use both
pln(builtinAbs(-5))   // 5 (built-in)
pln(abs(-10))         // 10 (module)
```

#### Best Practices

| Import Style | Use Case |
|--------------|----------|
| `import * as m from "module"` | **Recommended**: Avoids naming conflicts, clear function origin |
| `import { fn } from "module"` | When no conflict exists, more concise syntax |
| `import m from "module"` | When module has a default export |

```xxl
// Best practice: use namespace imports
import * as crypto from "crypto"
import * as locale from "locale"
import * as math from "math"

// Clear and unambiguous
token := crypto.genJwtToken({"sub": "user"}, "secret")
pln(locale.toPinYin("中国"))
pln(math.sin(1.57))
```

#### Built-in to Module Migration

Some built-in functions have been moved to standard library modules for better organization:

| Module | Functions Moved from Built-ins |
|--------|-------------------------------|
| `math` | `sin, cos, tan, asin, acos, atan, atan2, exp, log, log10, log2, pi, e, degToRad, radToDeg, random, round` |
| `locale` | `toPinYin, kanaToRomaji, kanjiToKana, kanjiToRomaji` |
| `crypto` | `genJwtToken, parseJwtToken` |
| `task` | `isCronExprValid, isCronExprDue, runTicker, stopTicker` |
| `image` | `genQr, scanQr, getImageInfo, resizeImage` |
| `ftp` | `newFtpClient` (deleted - use `ftp.connect()` or `ftp.newClient()`) |
| `ssh` | `newSshClient` (deleted - use `ssh.connect()` or `ssh.newClient()`) |
| `xlsx` | `newExcel, openExcel` (deleted - use `xlsx.create()` and `xlsx.open()`) |
| `csv` | `readCsv, writeCsv` (deleted - use `csv.read()` and `csv.write()`) |
| `xml` | `parseXml, parseXmlFile, newXmlDoc` (deleted - use `xml.parse()`, `xml.parseFile()`, `xml.create()`) |
| `yaml` | `parseYaml, toYaml, yamlToJson, jsonToYaml` (deleted - use `yaml.parse()`, `yaml.stringify()`, `yaml.toJson()`, `yaml.fromJson()`) |

**Note:** `createImage` is kept as a built-in function (alias to `image.createImage`).

When migrating code that uses these functions, use namespace imports:

```xxl
// Old code (built-in)
pln(sin(1.57))
pln(toPinYin("中国"))

// New code (module-based)
import * as math from "math"
import * as locale from "locale"

pln(math.sin(1.57))
pln(locale.toPinYin("中国"))
```

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

## Concurrency

Xxlang provides Go-style concurrency primitives:

### Goroutines

Use the `run` keyword to start a new goroutine:

```xxl
// Run anonymous block
run {
    sleep(100)
    pln("Background task completed")
}

// Run function with arguments
run worker(1, 2, 3)
```

### Tubes (Channels)

Tubes are typed channels for communication between goroutines:

```xxl
var tube = makeTube(10)

// Send
tube <- 42

// Receive
var val = <- tube
```

### Select

Use `select` to wait on multiple tube operations:

```xxl
select {
    case val = <- tube1:
        pln("Received from tube1:", val)
    case tube2 <- data:
        pln("Sent to tube2")
    default:
        pln("No operation ready")
}
```

### Context

Context provides timeout and cancellation:

```xxl
var ctx = contextWithTimeout(5000)  // 5 second timeout

run {
    sleep(3000)
    contextCancel(ctx)
}

if (contextDone(ctx)) {
    pln("Context cancelled or timed out")
}
```

### Sync Primitives

```xxl
var mutex = newMutex()
var wg = newWaitGroup()
var atomic = newAtomic(0)

mutex.lock()
// critical section
mutex.unlock()

wg.add(1)
run {
    // do work
    wg.done()
}
wg.wait()  // Wait for all goroutines

atomic.add(1)
pln(atomic.load())
```

> 📘 **See [CONCURRENCY.md](CONCURRENCY.md) for complete concurrency documentation.**

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
| `time` | Time and date functions |
| `http` | HTTP client operations |
| `file` | File operations |

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
run select
```

## Performance Considerations

1. **Bytecode Compilation**: Code is compiled to bytecode before execution
2. **Tail Call Optimization**: Tail-recursive functions use constant stack space
3. **Closure Optimization**: Closures only capture variables they actually use
4. **Inline Caching**: Method lookups are cached where possible

## Embedding in Go

See [EMBEDDING.md](EMBEDDING.md) for guide on embedding Xxlang in Go applications.
