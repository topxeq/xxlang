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
xxlang run script.xxl

# 运行编译后的字节码
xxlang run script.xxb

# 启动交互式 REPL
xxlang
```

## 命令行工具

```bash
xxlang                             # 启动 REPL
xxlang run file.xxl                # 运行源代码
xxlang run file.xxb                # 运行字节码
xxlang compile file.xxl            # 编译为可执行文件
xxlang compile --bytecode file.xxl # 编译为字节码 (.xxb)
xxlang compile -o out.xxb --bytecode file.xxl  # 指定输出路径
xxlang version                     # 显示版本
xxlang help                        # 显示帮助
```

## 字节码编译

Xxlang 支持将源代码编译为字节码文件，实现更快的加载速度和源码保护。

### 编译为字节码

```bash
# 编译 script.xxl 为 script.xxb
xxlang compile --bytecode script.xxl

# 指定输出路径
xxlang compile --bytecode -o program.xxb script.xxl
```

### 运行字节码

```bash
# 执行编译后的字节码
xxlang run script.xxb
```

### 对比

| 特性 | 源代码 (.xxl) | 字节码 (.xxb) |
|------|---------------|---------------|
| 加载过程 | 解析 → 编译 → 执行 | 反序列化 → 执行 |
| 启动开销 | ~5ms | ~1ms |
| 分发 | 源码可见 | 字节码混淆 |
| 文件大小 | 较小 | 约 5-10 倍 |
| 跨平台 | 是 | **是 - 相同字节码可在任意平台运行** |

### 跨平台兼容性

Xxlang 字节码文件是**平台无关的**：

- 同一个 `.xxb` 文件可在 Windows、Linux、macOS 上运行
- 支持不同的 CPU 架构（amd64、arm64）
- 在不同平台间移动无需重新编译

```bash
# 在 Windows 上编译
xxlang compile --bytecode script.xxl

# 复制 script.xxb 到 Linux 上运行
xxlang run script.xxb  # 无需修改即可运行！
```

这得益于：
- 固定字节序（Big Endian）用于版本号
- Go 的 gob 编码（平台无关的序列化）
- IEEE 754 浮点数（标准格式）
- UTF-8 字符串（通用编码）
- 不包含文件路径等平台相关数据

### 使用场景

- **开发阶段**：使用源代码，方便调试
- **生产部署**：部署字节码，启动更快
- **源码保护**：分发字节码，隐藏实现细节
- **嵌入应用**：将字节码嵌入 Go 程序

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

### 朴素递归

| 语言 | fib(35) 耗时 | 相对 C |
|------|-------------|--------|
| C (gcc -O2) | 25 ms | 1x |
| Go | 53 ms | 2.1x |
| Python | 2,714 ms | 107x |
| Xxlang | 6,324 ms | 250x |

### 使用尾调用优化

| 语言 | fib(35) TCO | 相对 C |
|------|-------------|--------|
| C | ~0.001 ms | 1x |
| Xxlang | 0.015 ms | 15x |

**关键发现**：算法选择比语言更重要。使用 TCO，Xxlang 实现了 **420,000 倍** 的性能提升。

详见 [benchmarks/FIB35_FINAL_REPORT.md](benchmarks/FIB35_FINAL_REPORT.md)。

## 尾调用优化

Xxlang 的尾调用优化 (TCO) 是**自动应用**的，当函数调用处于尾位置（即 return 后直接是函数调用）时会自动优化。

### TCO 生效规则

```xxl
// ✅ TCO 生效：return 后直接是函数调用
func sumTail(n, acc) {
    if (n <= 0) { return acc }
    return sumTail(n - 1, acc + n)  // TCO 自动应用
}

// ❌ TCO 不生效：return 后有其他操作
func fib(n) {
    if (n <= 1) { return n }
    return fib(n - 1) + fib(n - 2)  // 需要加法运算，不是尾调用
}
```

### 规则总结

| 模式 | TCO | 原因 |
|------|-----|------|
| `return func(args)` | ✅ 生效 | 调用是最后操作 |
| `return a + func(args)` | ❌ 不生效 | 调用后需要加法 |
| `var x = func(args); return x` | ❌ 不生效 | 先赋值给变量 |

### 正确写法示例

```xxl
// 尾递归斐波那契
func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)  // TCO 生效
}

func fib(n) { return fibTail(n, 0, 1) }

println(fib(10000))  // 瞬间完成，不会栈溢出！
```

**性能提升**：使用 TCO 后，fib(35) 性能提升约 **420,000 倍**！

### 通过 Go 函数实现高性能

将 Xxlang 嵌入 Go 应用时，可以注册 Go 函数获得原生性能：

```go
// 注册 Go 函数
interp.SetGlobal("goFib", &objects.Builtin{
    Fn: func(args ...objects.Object) objects.Object {
        n := args[0].(*objects.Int).Value
        return &objects.Int{Value: fibFast(n)}  // Go 原生，极速！
    },
})

// 从 Xxlang 调用
interp.Eval("goFib(100000)")  // 微秒级，而非秒级！
```

| 方式 | fib(35) | 性能提升 |
|------|---------|----------|
| Xxlang 朴素递归 | 6.5 秒 | 基准 |
| Xxlang TCO | 200 µs | 32,000x |
| Go 函数 | 25 µs | **260,000x** |

详见 [docs/EMBEDDING.md](docs/EMBEDDING.md)。

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
