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

# Run a source file
xxlang run script.xxl

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
| `STRING` | UTF-8 string | `"hello"`, `"world\n"` |
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
| `MODULE` | Imported module | `import "std/math"` |
| `BUILTIN` | Built-in function | `len`, `print`, `typeOf` |

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
    println(i)
    i++
}
```

### For Loop

```xxl
// C-style for loop
for (var i = 0; i < 5; i++) {
    println(i)
}

// For-in loop (arrays)
for (item in [1, 2, 3]) {
    println(item)
}

// For-in loop (maps)
for (key, value in {"a": 1, "b": 2}) {
    println(key + ": " + value)
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
    println(i)
}
```

### Switch Statement

```xxl
switch (value) {
    case 1:
        println("one")
    case 2:
        println("two")
    default:
        println("other")
}
```

## Functions

### Basic Function

```xxl
func add(a, b) {
    return a + b
}

println(add(3, 4))  // 7
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
println(counter())  // 1
println(counter())  // 2
println(counter())  // 3
```

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

println(fib(10))  // 55
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

println(sumTail(10000, 0))     // Works without stack overflow!
println(factorial(1000, 1))    // Works!
println(fibHelper(10000, 0, 1)) // Works!
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
println(arr[0])   // 10
println(arr[1])   // 20
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
println(person["name"])   // "Alice"
println(person["age"])    // 30
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
println(p3.toString())  // "(4, 6)"
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
println(dog.speak())  // "Rex says Woof!"
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
// Import entire module
import "std/math"
println(math.sqrt(16))

// Import with alias
import "std/io" as io
io.println("Hello")

// Import specific functions
import "std/string" { upper, lower }
println(upper("hello"))
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

1. `std/xxx` - Standard library modules
2. `plugin/xxx` - Native Go plugins (`.so` files)
3. `./xxx` or `../xxx` - Relative path modules
4. `/xxx` - Absolute path modules

## Error Handling

### Try-Catch-Finally

```xxl
try {
    var result = riskyOperation()
} catch (err) {
    println("Error: " + err)
} finally {
    println("Cleanup")
}
```

### Throw Statement

```xxl
func divide(a, b) {
    if (b == 0) {
        throw "division by zero"
    }
    return a / b
}
```

## Standard Library

Xxlang includes several standard library modules:

| Module | Description |
|--------|-------------|
| `std/io` | Input/output operations |
| `std/string` | String manipulation |
| `std/math` | Mathematical functions |
| `std/array` | Array utilities |
| `std/json` | JSON parsing and encoding |
| `std/regex` | Regular expressions |
| `std/crypto` | Cryptographic functions |

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
