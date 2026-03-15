# Xxlang 变量作用域参考

## 概述

Xxlang 使用词法作用域（Lexical Scoping），变量的查找顺序从内到外：

1. **局部作用域** - 函数内部声明的变量
2. **外层函数作用域** - 外层函数的变量（用于闭包）
3. **全局作用域** - 顶层声明的变量

## 全局变量

在顶层声明的变量是全局变量，可以在任何地方访问：

```xxl
var globalVar = "global"

func getGlobal() {
    return globalVar  // 可以读取全局变量
}

func setGlobal(val) {
    globalVar = val   // 可以修改全局变量
}
```

## 局部变量

函数内部声明的变量是局部变量：

```xxl
func myFunction() {
    var localVar = "local"  // 只在 myFunction 内部可访问
    return localVar
}
```

## 变量遮蔽（Shadowing）

局部变量可以遮蔽（隐藏）外层的同名变量：

```xxl
var name = "global"

func shadowExample() {
    var name = "local"     // 遮蔽全局的 'name'
    println(name)          // 输出 "local"
}

shadowExample()
println(name)              // 输出 "global"（未改变）
```

函数参数也会遮蔽外层同名变量：

```xxl
var x = "global"

func paramShadow(x) {      // 参数遮蔽全局的 'x'
    return x
}

println(paramShadow("arg"))  // "arg"
println(x)                   // "global"（未改变）
```

## 嵌套函数与闭包

函数可以嵌套，内层函数可以捕获外层函数的变量：

```xxl
func outer() {
    var message = "Hello"

    func inner() {
        return message  // 捕获外层的 'message'
    }

    return inner()
}

println(outer())  // "Hello"
```

内层的局部变量会遮蔽外层变量：

```xxl
func outer() {
    var x = "outer"

    func inner() {
        var x = "inner"  // 遮蔽外层的 x
        return x
    }

    return inner() + " " + x
}

println(outer())  // "inner outer"
```

## 闭包

闭包通过引用捕获变量，允许修改捕获的变量：

```xxl
func makeCounter() {
    var count = 0

    func counter() {
        count = count + 1  // 修改捕获的变量
        return count
    }

    return counter
}

var c1 = makeCounter()
println(c1())  // 1
println(c1())  // 2
println(c1())  // 3

var c2 = makeCounter()  // 新闭包，新的捕获变量
println(c2())  // 1（与 c1 独立）
```

## ⚠️ 重要：多个闭包共享变量的情况

> **已知行为**：当多个闭包在同一作用域创建并捕获同一变量时，每个闭包会获得该变量的**独立副本**，而不是共享同一个引用。

### 受影响的模式

#### 模式 1：在 map/对象中返回多个闭包

```xxl
func createObject() {
    var value = "initial"

    return {
        "set": func(newVal) {
            value = newVal      // 修改自己的副本
        },
        "get": func() {
            return value         // 返回自己的副本
        }
    }
}

var obj = createObject()
obj["set"]("updated")
println(obj["get"]())  // "initial"（不是预期的 "updated"！）
```

**原因**：`set` 和 `get` 两个闭包各自捕获了 `value` 变量的独立副本，它们之间不共享。

#### 模式 2：在数组中返回多个闭包

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
println(counters[1]())  // 0（不是预期的 1！）
```

### 解决方案：使用 Map 作为共享状态

Map 是引用类型，可以被多个闭包正确共享：

```xxl
func createObject() {
    var state = {"value": "initial"}  // Map 是引用类型

    return {
        "set": func(newVal) {
            state["value"] = newVal   // 修改共享的 map
        },
        "get": func() {
            return state["value"]      // 读取共享的 map
        }
    }
}

var obj = createObject()
obj["set"]("updated")
println(obj["get"]())  // "updated" ✓ 正确！
```

### 模式对比表

| 模式 | 预期行为 | 当前行为 | 状态 |
|------|----------|----------|------|
| 单个闭包捕获变量 | 变量在多次调用间共享 | ✅ 正常工作 | OK |
| 多个闭包在不同调用中创建 | 每个闭包有独立的变量 | ✅ 正常工作 | OK |
| 多个闭包在同一返回值中 | 所有闭包共享同一变量引用 | ❌ 各自有独立副本 | **已知问题** |
| 使用 map 作为共享状态 | Map 通过引用被共享 | ✅ 正常工作 | 推荐使用 |

### 实际应用示例

#### 正确的模式：单个闭包

```xxl
// ✅ 单个闭包正常工作
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

#### 需要使用 map 的模式：多个相关闭包

```xxl
// ✅ 使用 map 共享状态
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

## 内置函数遮蔽

局部变量可以遮蔽内置函数：

```xxl
func example() {
    var len = 100       // 遮蔽内置的 len()
    println(len)        // 100
}

example()
println(len([1,2,3]))   // 3（内置函数在外部仍然正常）
```

## 作用域解析总结

```xxl
var a = "global"

func outer() {
    var a = "outer local"
    var b = "outer only"

    func inner() {
        var a = "inner local"  // 遮蔽外层的 a
        println(a)              // "inner local"
        println(b)              // "outer only"（捕获的外层变量）
    }

    inner()
    println(a)              // "outer local"
}
```

## 最佳实践

1. **避免不必要的变量遮蔽**：使用不同的变量名可以提高代码可读性
2. **需要共享状态时使用 map**：当多个闭包需要共享变量时，使用 map 作为共享状态容器
3. **单个闭包正常使用**：对于单个闭包的场景（如计数器），直接捕获变量即可
4. **注意作用域边界**：函数参数和局部变量都会遮蔽外层同名变量
