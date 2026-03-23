# Xxlang (现象语言)
![Coverage](https://img.shields.io/badge/Coverage-81.8%25-brightgreen)

[English](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/topxeq/xxlang.svg)](https://pkg.go.dev/github.com/topxeq/xxlang)
[![Go Report Card](https://goreportcard.com/badge/github.com/topxeq/xxlang)](https://goreportcard.com/report/github.com/topxeq/xxlang)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Xxlang 是一个基于字节码虚拟机的脚本语言，使用 Go 语言实现。

## 特性

- **字节码虚拟机** - 基于寄存器的虚拟机，比栈式快 21%
- **JIT 编译** - x86-64 JIT 编译器，支持 Linux、macOS 和 Windows（递归算法快 93 倍）
- **方法尾调用优化** - 函数和方法的尾调用优化，支持高效递归
- **闭包支持** - 一等函数，完整的闭包支持
- **类与面向对象** - 支持类、继承和多级 super 调用
- **异常处理** - 完整的 try/catch/finally/throw 支持
- **模块系统** - 导入导出机制，丰富的标准库
- **插件系统** - WASM 插件实现高性能操作（兼容 Windows，无需 CGO）
- **丰富的内置函数** - 60+ 个内置函数，支持字符串、数学、数组和映射操作
- **交互式 REPL** - 支持多行输入和状态持久化
- **可嵌入** - 可作为库在其他 Go 项目中使用
- **可编译** - 编译为独立可执行文件或跨平台字节码
- **跨平台** - 支持 Linux、macOS、Windows 的 amd64 和 arm64 架构

## 文档

- [语言参考](docs/LANGUAGE.md) - 完整的语言语法和特性
- [内置函数](docs/BUILTINS_zh.md) - 全局内置函数
- [标准库](docs/STDLIB_zh.md) - 标准库模块（os、json、math 等）
- [并发编程](docs/CONCURRENCY_zh.md) - Goroutine、Tube（管道）、Select、Context、同步原语
- [微服务模式](docs/MICROSERVICE_zh.md) - HTTP/HTTPS 服务器、REST API、WebSocket
- [代码示例](docs/EXAMPLES_zh.md) - 常用场景代码示例
- [变量作用域](docs/SCOPE_zh.md) - 变量作用域和闭包行为
- [嵌入指南](docs/EMBEDDING.md) - 在 Go 应用中使用 Xxlang
- [JIT 编译](docs/JIT.md) - JIT 编译器文档
- [插件系统](docs/PLUGIN.md) - 编写原生 Go 插件实现高性能
- [性能测试](benchmarks/RESULTS.md) - 性能分析

## 安装

### 方式 1：一键安装（推荐）

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
```

或使用 wget：
```bash
wget -qO- https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
```

### 方式 2：下载预编译二进制文件

从 [GitHub Releases](https://github.com/topxeq/xxlang/releases) 下载适合你平台的最新版本：

```bash
# Linux (amd64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-linux-amd64.tar.gz
tar -xzf xxlang-linux-amd64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# Linux (arm64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-linux-arm64.tar.gz
tar -xzf xxlang-linux-arm64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# Windows (PowerShell)
# 下载: https://github.com/topxeq/xxlang/releases/latest/download/xxlang-windows-amd64.zip
# 解压得到 xxl.exe 并添加到 PATH

# macOS (amd64)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-darwin-amd64.tar.gz
tar -xzf xxlang-darwin-amd64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/

# macOS (arm64/Apple Silicon)
wget https://github.com/topxeq/xxlang/releases/latest/download/xxlang-darwin-arm64.tar.gz
tar -xzf xxlang-darwin-arm64.tar.gz
chmod +x xxl
sudo mv xxl /usr/local/bin/
```

### 方式 3：通过 Go 安装

```bash
go install github.com/topxeq/xxlang/cmd/xxl@latest
```

### 方式 4：从源码构建

```bash
git clone https://github.com/topxeq/xxlang.git
cd xxlang
go build -o xxl ./cmd/xxl
```

## 更新

Xxlang 支持从 GitHub Release 自动更新：

```bash
xxl update
```

此命令会：
1. 从 GitHub 获取最新版本信息
2. 与当前版本比较
3. 下载适合你操作系统和架构的压缩包
4. 解压并替换当前可执行文件

**注意**：在 Windows 上，旧的执行文件可能会保留为 `xxl.exe.old`，直到下次重启。

## 快速开始

```bash
# 运行本地脚本文件
xxl run script.xxl

# 从 URL 运行脚本
xxl https://raw.githubusercontent.com/user/repo/main/script.xxl

# 运行编译后的字节码
xxl run script.xxb

# 启动交互式 REPL
xxl
```

## 命令行工具

```bash
xxl                                # 启动 REPL（默认使用寄存器 VM）
xxl script.xxl                     # 运行本地脚本（快捷方式）
xxl run script.xxl                 # 运行源代码
xxl run script.xxb                 # 运行字节码
xxl https://example.com/script.xxl # 从 URL 运行脚本
xxl -cloud basic.xxl               # 从配置的云端 URL 运行脚本
xxl compile script.xxl             # 编译为可执行文件
xxl compile --bytecode script.xxl  # 编译为字节码 (.xxb)
xxl compile -o out.xxb --bytecode script.xxl   # 指定输出路径
xxl update                         # 自我更新到最新版本
xxl version                        # 显示版本
xxl help                           # 显示帮助
```

### VM 选择

Xxlang 支持两种虚拟机：

| VM | 说明 | 性能 |
|----|------|------|
| **寄存器 VM**（默认） | 现代、优化 | 平均快 21% |
| 栈式 VM | 传统、兼容 | 基准 |

```bash
xxl script.xxl                  # 寄存器 VM（默认，推荐）
xxl --vm=register script.xxl    # 明确使用寄存器 VM
xxl --vm=stack script.xxl       # 使用栈式 VM（兼容模式）
xxl --stack-vm script.xxl       # 同 --vm=stack
```

**注意**：寄存器 VM 对方法调用语法（如 `obj.method()`）的支持有限。对于使用此类特性的脚本，请使用 `--vm=stack`。

## 云端脚本执行

Xxlang 支持通过 `-cloud` 参数从配置的云端 URL 基地址执行脚本。

### 配置

创建 `~/.xxl/settings.json`（Linux 系统也可以是 `/.xxl/settings.json`，Windows 是 `C:\.xxl\settings.json`）：

```json
{
  "cloudUrlBase": "https://script.topget.org/"
}
```

### 使用

```bash
xxl -cloud basic.xxl
```

这将获取并执行 `https://script.topget.org/basic.xxl`。

你也可以在脚本中访问配置：

```xxl
import "os"
var cfg = os.getConfigObj()
pln(cfg["cloudUrlBase"])
```

## 字节码编译

Xxlang 支持将源代码编译为字节码文件，实现更快的加载速度和源码保护。

### 编译为字节码

```bash
# 编译 script.xxl 为 script.xxb
xxl compile --bytecode script.xxl

# 指定输出路径
xxl compile --bytecode -o program.xxb script.xxl
```

### 运行字节码

```bash
# 执行编译后的字节码
xxl run script.xxb
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
xxl compile --bytecode script.xxl

# 复制 script.xxb 到 Linux 上运行
xxl run script.xxb  # 无需修改即可运行！
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
pln(counter())  // 1
pln(counter())  // 2
```

### 类与面向对象

```xxl
class Point {
    func init(x, y) {
        this.x = x
        this.y = y
    }

    func add(other) {
        return new Point(this.x + other.x, this.y + other.y)
    }
}

var p1 = new Point(1, 2)
var p2 = new Point(3, 4)
var p3 = p1.add(p2)
```

### 控制流

```xxl
// 条件语句
if (x > 0) {
    pln("正数")
} else {
    pln("非正数")
}

// while 循环
var i = 0
while (i < 5) {
    pln(i)
    i = i + 1
}

// for 循环
for (var j = 0; j < 5; j = j + 1) {
    pln(j)
}

// for-in 循环
for (item in [1, 2, 3]) {
    pln(item)
}
```

### 模块

```xxl
// 导入标准库
import "math"
pln(math.sqrt(16))

// 导入特定函数
import "io" { readFile, writeFile }
```

### 插件系统

编写原生 Go 插件实现高性能操作：

```xxl
// 导入 Go 插件
import "plugin/fib"

// 从 Xxlang 调用 Go 函数
pln(fib.fast(50))      // 12586269025
pln(fib.matrix(92))    // int64 范围内最大的斐波那契数
```

**两种插件类型：**

| 类型 | Windows | 需要 CGO | 运行时加载 |
|------|---------|----------|------------|
| 静态插件 | ✅ | ❌ | ❌ |
| WASM 插件 | ✅ | ❌ | ✅ |

| 方式 | fib(35) 耗时 | 性能提升 |
|------|-------------|----------|
| Xxlang 朴素递归 | 6.5 秒 | 基准 |
| Xxlang 尾递归 | 136 µs | 47,000x |
| Go 插件 | **35 µs** | **180,000x** |

详见 [插件系统](docs/PLUGIN.md)。

### 内置函数

完整的内置函数参考请见 [内置函数参考手册](docs/BUILTINS_zh.md)。

标准库模块请见 [标准库参考手册](docs/STDLIB_zh.md)。

```xxl
// 字符串函数
pln(upper("hello"))           // HELLO
pln(lower("HELLO"))           // hello
pln(substr("hello", 1, 4))    // ell
pln(split("a,b,c", ","))      // [a, b, c]
pln(containsStr("hello", "ell"))  // true

// 数学函数
pln(sqrt(16))    // 4
pln(pow(2, 10))  // 1024
pln(abs(-42))    // 42
pln(floor(3.7))  // 3
pln(ceil(3.2))   // 4

// 数组函数
var arr = [3, 1, 4, 1, 5]
pln(sort(arr))      // [1, 1, 3, 4, 5]
pln(sum(arr))       // 14
pln(reverse(arr))   // [5, 1, 4, 1, 3]
pln(first(arr))     // 3
pln(last(arr))      // 5

// 映射函数
var m = {"a": 1, "b": 2}
pln(keys(m))        // [a, b]
pln(values(m))      // [1, 2]
pln(hasKey(m, "a")) // true
```

## REPL 命令

```
exit, quit  - 退出 REPL
help        - 显示帮助信息
history     - 显示命令历史
clear       - 清除所有变量和函数
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

Xxlang 使用基于寄存器的字节码虚拟机，支持尾调用优化，相比传统栈式虚拟机有显著的性能提升。

### 寄存器 VM vs 栈式 VM vs Go vs Python

| 基准测试 | 栈式 VM (µs) | 寄存器 VM (µs) | Go (µs) | Python (µs) | Reg vs Stack |
|----------|-------------:|---------------:|--------:|------------:|-------------:|
| Fibonacci(15) | 908 | 872 | 3.5 | 178 | 1.04x |
| Fibonacci(20) | 5,219 | 4,785 | 39 | 1,980 | 1.09x |
| Fibonacci 迭代(30) | 452 | 403 | ~0.01 | 1.6 | 1.12x |
| 阶乘(12) | 449 | 388 | ~0.01 | 1.4 | 1.16x |
| For 循环(100) | 426 | 382 | ~0.04 | 3.7 | 1.11x |
| **For 循环(1000)** | 813 | **474** | ~0.33 | 43.7 | **1.72x** |
| While 循环(100) | 432 | 348 | ~0.04 | 6.5 | 1.24x |
| 函数调用(200) | 556 | 526 | ~0.04 | 17.7 | 1.06x |
| 算术运算 | 397 | 338 | ~0.0003 | 0.15 | 1.17x |
| **密集算术(1000)** | 897 | **595** | ~0.65 | 139 | **1.51x** |
| 嵌套表达式 | 410 | 360 | ~0.0003 | 0.13 | 1.14x |
| 比较运算(400) | 529 | 472 | ~0.04 | 12.7 | 1.12x |
| **素数检测(100)** | 653 | **481** | 2.4 | 31 | **1.36x** |
| 冒泡排序(10) | 494 | 455 | ~0.04 | 6.7 | 1.09x |
| 数组求和(1000) | 383 | 353 | ~0.67 | 16.1 | 1.08x |
| **字符串拼接(100)** | 452 | **324** | 5.2 | 5.5 | **1.39x** |

**关键发现：**
- 寄存器 VM 比栈式 VM 平均快 **21%**
- 最高加速比：**1.72x**（For 循环 1000）
- 内存分配减少 **25.5%**

### 内存分配优化

| 基准测试 | 栈式 VM (分配次数) | 寄存器 VM (分配次数) | 减少比例 |
|----------|------------------:|---------------------:|---------:|
| Fibonacci(15) | 212 | 173 | 18.4% |
| Fibonacci 迭代 | 244 | 193 | 20.9% |
| For 循环(1000) | 716 | 137 | **80.9%** |
| 函数调用 | 289 | 230 | 20.4% |
| 密集算术 | 194 | 156 | 19.6% |

### fib(35) 性能对比 (2026年3月)

| 语言 | 耗时 | 相对 Go |
|------|------|---------|
| **C** | **45 ms** | 快 0.87 倍 |
| **Java** | **47 ms** | 快 0.90 倍 |
| **Go** | **52 ms** | 1x (基准) |
| **Xxlang JIT** | **54 ms** | 慢 1.04 倍 |
| Python | 2,706 ms | 慢 52 倍 |
| Xxlang VM | 5,020 ms | 慢 97 倍 |

**关键发现：**
- **JIT 接近原生性能**：与 C/Go/Java 差距在 4-10% 以内
- **JIT 比 Python 快 50 倍**：对于递归算法
- **无 CGO 依赖**：纯 Go 实现的 JIT 编译器

### JIT 编译器特性

| 特性 | 说明 |
|------|------|
| **回调机制** | 原生代码可回调 Go 执行内置函数、函数调用、数组和对象操作 |
| **对象句柄池** | 高效内存管理，句柄可复用 |
| **尾调用优化** | 递归函数编译为循环，O(1) 栈空间 |
| **8 参数支持** | 支持 0-8 个参数的函数 |
| **线程安全** | 使用 `sync.RWMutex` 保护所有回调操作 |

### 使用尾调用优化

| 语言 | fib(35) TCO | 相对 Go |
|------|-------------|---------|
| **Go (迭代)** | **~0.025 ms** | 1x |
| Xxlang TCO | 0.013 ms | ~0.5x |

**关键发现：**
- TCO 使递归比朴素方法快 **400,000 倍**
- Xxlang TCO 可以**匹配甚至超越** Go 的迭代实现

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

pln(fib(10000))  // 瞬间完成，不会栈溢出！
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

## 插件系统

编写原生 Go 插件实现高性能操作：

```xxl
// 导入 Go 插件
import "plugin/fib"

// 从 Xxlang 调用 Go 函数
pln(fib.fast(50))      // 12586269025
pln(fib.matrix(92))    // int64 范围内最大的斐波那契数
```

**两种插件类型：**

| 类型 | Windows | 需要 CGO | 运行时加载 |
|------|---------|----------|------------|
| 静态插件 | ✅ | ❌ | ❌ |
| WASM 插件 | ✅ | ❌ | ✅ |

| 方式 | fib(35) 耗时 | 性能提升 |
|------|-------------|----------|
| Xxlang 朴素递归 | 6.5 秒 | 基准 |
| Xxlang 尾递归 | 136 µs | 47,000x |
| Go 插件 | **35 µs** | **180,000x** |

详见 [docs/PLUGIN.md](docs/PLUGIN.md)。

## 内置函数列表

Xxlang 提供 60+ 个内置函数：

| 类别 | 函数 |
|------|------|
| 输出 | `pln`, `pr`, `pl`, `prf` |
| 通用 | `len`, `typeOf`, `toStr` |
| 字符串 | `substr`, `split`, `join`, `trim`, `upper`, `lower`, `containsStr`, `replace`, `startsWith`, `endsWith`, `repeat`, `charAt`, `padLeft`, `padRight`, `trimLeft`, `trimRight` |
| 数学 | `abs`, `floor`, `ceil`, `sqrt`, `pow`, `min`, `max`, `round`, `clamp`, `sign` |
| 类型转换 | `int`, `float`, `string` |
| 数组 | `push`, `pop`, `first`, `last`, `rest`, `concat`, `indexOf`, `containsArr`, `sort`, `sum`, `avg`, `reverse`, `unique`, `flatten`, `without`, `take`, `drop` |
| 映射 | `keys`, `values`, `hasKey`, `delete`, `merge`, `entries` |
| 类型检查 | `isEmpty`, `isString`, `isNumber`, `isInt`, `isFloat`, `isArray`, `isMap`, `isBool`, `isFunction`, `isNull` |
| 工具 | `range`, `runCode`, `loadPlugin`, `format`, `checkErr`, `checkEmpty`, `genOtpCode` |

## 许可证

MIT License
