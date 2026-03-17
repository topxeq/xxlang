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
    println(name)          // Prints "local"
}

shadowExample()
println(name)              // Prints "global" (unchanged)
```

Function parameters also shadow outer variables:

```xxl
var x = "global"

func paramShadow(x) {      // Parameter shadows global 'x'
    return x
}

println(paramShadow("arg"))  // "arg"
println(x)                   // "global" (unchanged)
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

println(outer())  // "Hello"
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

println(outer())  // "inner outer"
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
println(c1())  // 1
println(c1())  // 2
println(c1())  // 3

var c2 = makeCounter()  // New closure, new captured variable
println(c2())  // 1 (independent from c1)
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
println(obj["get"]())  // "initial" (NOT "updated" as expected!)
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
println(counters[0]())  // 1
println(counters[1]())  // 0 (NOT 1 as expected!)
```

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
println(obj["get"]())  // "updated" Works correctly!
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
println(counter())  // 1
println(counter())  // 2
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
println(account["deposit"](50))    // 150
println(account["withdraw"](30))   // 120
println(account["getBalance"]())   // 120
```

## Built-in Function Shadowing

Local variables can shadow built-in functions:

```xxl
func example() {
    var len = 100       // Shadows built-in len()
    println(len)        // 100
}

example()
println(len([1,2,3]))   // 3 (built-in still works outside)
```

## Scope Resolution Summary

```xxl
var a = "global"

func outer() {
    var a = "outer local"
    var b = "outer only"

    func inner() {
        var a = "inner local"  // Shadows outer's a
        println(a)              // "inner local"
        println(b)              // "outer only" (captured)
    }

    inner()
    println(a)              // "outer local"
}
```

## Best Practices

1. **Avoid unnecessary variable shadowing**: Using distinct variable names improves code readability
2. **Use maps for shared state**: When multiple closures need to share variables, use a map as the shared state container
3. **Single closures work normally**: For single closure scenarios (like counters), directly capturing variables works correctly
4. **Be aware of scope boundaries**: Both function parameters and local variables shadow outer variables with the same name
