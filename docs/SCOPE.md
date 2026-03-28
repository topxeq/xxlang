# Xxlang Variable Scope Reference

## Overview

Xxlang uses lexical scoping, where variables are resolved from innermost to outermost scope:

1. **Local scope** - Variables declared inside a function
2. **Enclosing function scope** - Variables from outer functions (for closures)
3. **Global scope** - Variables declared at the top level

## Global Variables

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

## Local Variables

Variables declared inside a function are local to that function:

```xxl
func myFunction() {
    var localVar = "local"  // Only accessible inside myFunction
    return localVar
}
```

## Variable Shadowing

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

## Nested Functions and Closures

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

## Closures

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

## Important: Multiple Closures Sharing Variables

> **Known Behavior**: When multiple closures are created in the same scope and capture the same variable, each closure currently gets its own **copy** of the variable, rather than sharing a single reference.

### Affected Patterns

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

**Reason**: The `set` and `get` closures each capture their own independent copy of `value`, they don't share the same reference.

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

**Reason**: Each closure captures its own independent copy of `count`, they don't share the same reference.

### Workaround: Use a Map as Shared State

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
pln(obj["get"]())  // "updated" Works correctly!
```

### Pattern Comparison Table

| Pattern | Expected Behavior | Current Behavior | Status |
|---------|-------------------|------------------|--------|
| Single closure captures variable | Variable shared across calls | Works correctly | OK |
| Multiple closures in separate calls | Each closure has its own variable | Works correctly | OK |
| Multiple closures in same return | All share same variable reference | Each has its own copy | **Known Issue** |
| Multiple closures + map workaround | Map is shared by reference | Works correctly | Use This |

### Practical Application Examples

#### Correct Pattern: Single Closure

```xxl
// Single closure works correctly
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
```

#### Pattern Requiring Map: Multiple Related Closures

```xxl
// Use map to share state
func createBankAccount(initialBalance) {
    var state = {"balance": initialBalance}

    return {
        "deposit": func(amount) {
            state["balance"] = state["balance"] + amount
            return state["balance"]
        },
        "withdraw": func(amount) {
            if (state["balance"] >= amount) {
                state["balance"] = state["balance"] - amount
                return state["balance"]
            }
            return null
        },
        "getBalance": func() {
            return state["balance"]
        }
    }
}

var account = createBankAccount(100)
pln(account["deposit"](50))    // 150
pln(account["withdraw"](30))   // 120
pln(account["getBalance"]())   // 120
```

## Built-in Function Shadowing

Local variables can shadow built-in functions:

```xxl
func example() {
    var len = 100       // Shadows built-in len()
    pln(len)        // 100
}

example()
pln(len([1,2,3]))   // 3 (built-in still works outside)
```

### Complete Name Resolution Priority

Xxlang resolves names in the following order (highest to lowest priority):

1. **Local variables** - Variables declared inside the current function
2. **Enclosing function variables** - Variables from outer functions (for closures)
3. **Global variables** - Variables declared at the top level
4. **Destructured imports** - Functions imported via `import { fn } from "module"`
5. **Built-in functions** - Predefined functions like `len`, `pln`, `typeOf`

### Module Imports and Scope

#### Destructuring Import

Destructured imports are added to the current scope and can shadow built-ins:

```xxl
// Built-in abs exists
pln(abs(-5))  // 5

// Destructuring import shadows built-in
import { abs } from "math"
pln(abs(-10))  // Uses module function, not built-in
```

#### Namespace Import

Namespace imports create a new namespace object and do NOT shadow built-ins:

```xxl
import * as math from "math"

// Both are accessible
pln(abs(-5))       // 5 (built-in)
pln(math.abs(-10)) // 10 (module function)
```

### Variable vs Import vs Built-in

```xxl
// Priority: variable > import > built-in

import { abs } from "math"  // Shadows built-in abs

abs := func(x) { x * 2 }    // Shadows module abs
pln(abs(5))                 // 10 (user variable, highest priority)
```

### Best Practices for Avoiding Conflicts

1. **Use namespace imports** (`import * as m from "module"`) to avoid shadowing built-ins
2. **Avoid naming variables after built-in functions** unless intentionally overriding
3. **Check for existing built-ins** before defining variables with common names like `len`, `type`, `abs`

## Scope Resolution Summary

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

## Best Practices

1. **Avoid unnecessary variable shadowing**: Using distinct variable names improves code readability
2. **Use maps for shared state**: When multiple closures need to share variables, use a map as the shared state container
3. **Single closures work normally**: For single closure scenarios (like counters), directly capturing variables works correctly
4. **Be aware of scope boundaries**: Both function parameters and local variables shadow outer variables with the same name
