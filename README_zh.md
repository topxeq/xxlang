# Xxlang (现象语言)

[English](README.md)

Xxlang 是一个基于字节码虚拟机的脚本语言，使用 Go 语言实现。

## 特性

- **字节码虚拟机** - 基于栈的虚拟机，高效执行
- **闭包支持** - 一等函数，完整的闭包支持
- **丰富的内置函数** - 41 个内置函数，支持字符串、数学、数组和映射操作
- **交互式 REPL** - 支持多行输入和状态持久化
- **可嵌入** - 可作为库在其他 Go 项目中使用

## 安装

```bash
go install github.com/topxeq/xxlang/cmd/xxlang@latest
```

## 快速开始

```bash
# 运行脚本文件
xxlang script.xxl

# 启动交互式 REPL
xxlang
```

## 语言示例

### 变量和类型

```xxl
var x = 10
var y = 3.14
var name = "hello"
var arr = [1, 2, 3, 4, 5]
var map = {"a": 1, "b": 2}
```

### 函数和闭包

```xxl
func add(a, b) {
    return a + b
}

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

### 控制流

```xxl
// 条件语句
if (x > 0) {
    println("正数")
} else {
    println("非正数")
}

// while 循环
var i = 0
while (i < 5) {
    println(i)
    i = i + 1
}

// for 循环
for (var j = 0; j < 5; j = j + 1) {
    println(j)
}

// for-in 循环
for (item in [1, 2, 3]) {
    println(item)
}
```

### 内置函数

```xxl
// 字符串函数
println(upper("hello"))           // HELLO
println(lower("HELLO"))           // hello
println(substr("hello", 1, 4))    // ell
println(split("a,b,c", ","))      // [a, b, c]
println(containsStr("hello", "ell"))  // true

// 数学函数
println(sqrt(16))    // 4
println(pow(2, 10))  // 1024
println(abs(-42))    // 42
println(floor(3.7))  // 3
println(ceil(3.2))   // 4

// 数组函数
var arr = [3, 1, 4, 1, 5]
println(sort(arr))      // [1, 1, 3, 4, 5]
println(sum(arr))       // 14
println(reverse(arr))   // [5, 1, 4, 1, 3]
println(first(arr))     // 3
println(last(arr))      // 5

// 映射函数
var m = {"a": 1, "b": 2}
println(keys(m))        // [a, b]
println(values(m))      // [1, 2]
println(hasKey(m, "a")) // true
```

## REPL 命令

```
exit, quit  - 退出 REPL
help        - 显示帮助信息
history     - 显示命令历史
clear       - 清除所有变量和函数
```

## 从源码构建

```bash
git clone https://github.com/topxeq/xxlang.git
cd xxlang
go build -o xxlang ./cmd/xxlang
```

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行集成测试
go test ./tests/... -v

# 运行性能测试
go test ./tests/... -bench=.
```

## 性能

Xxlang 使用字节码虚拟机，支持尾调用优化。

| 语言 | fib(35) 耗时 |
|------|-------------|
| Go | 0.056s |
| Python | 2.77s |
| Xxlang | 9.37s |

## 尾调用优化

Xxlang 支持尾调用优化，递归函数可以使用常量栈空间：

```xxl
func sumTail(n, acc) {
    if (n <= 0) {
        return acc
    }
    return sumTail(n - 1, acc + n)
}

println(sumTail(10000, 0))  // 50005000，不会栈溢出
```

## 内置函数列表

Xxlang 提供 41 个内置函数：

| 类别 | 函数 |
|------|------|
| 通用 | `len`, `typeOf`, `print`, `println` |
| 字符串 | `substr`, `split`, `join`, `trim`, `upper`, `lower`, `containsStr`, `replace`, `startsWith`, `endsWith` |
| 数学 | `abs`, `floor`, `ceil`, `sqrt`, `pow`, `min`, `max` |
| 类型转换 | `int`, `float`, `string` |
| 数组 | `push`, `pop`, `first`, `last`, `rest`, `concat`, `indexOf`, `containsArr`, `sort`, `sum`, `avg`, `reverse` |
| 映射 | `keys`, `values`, `hasKey`, `delete` |
| 迭代 | `range` |

## 许可证

MIT License
