# Xxlang Terminology Glossary (术语词汇表)

This document provides standardized terminology and Chinese translations for Xxlang documentation.

---

## Core Concepts (核心概念)

| English | Chinese | Description |
|---------|---------|-------------|
| Bytecode | 字节码 | Compiled intermediate code executed by the VM |
| VM (Virtual Machine) | 虚拟机 | The runtime that executes bytecode |
| Register VM | 寄存器虚拟机 | VM using register-based architecture |
| Stack VM | 栈式虚拟机 | VM using stack-based architecture |
| JIT (Just-In-Time) | 即时编译 | Compilation to native code at runtime |
| Interpreter | 解释器 | Component that executes bytecode directly |
| Compiler | 编译器 | Component that converts source to bytecode |

## Types (类型)

| English | Chinese | Description |
|---------|---------|-------------|
| Type | 类型 | Data type in Xxlang |
| Primitive Type | 原始类型 | Basic types: INT, FLOAT, STRING, BOOL, NULL |
| Composite Type | 复合类型 | ARRAY, MAP, CHARS, FUNCTION, etc. |
| Integer | 整数 | 64-bit integer type |
| Float | 浮点数 | 64-bit floating-point type |
| String | 字符串 | UTF-8 encoded text |
| Boolean / Bool | 布尔值 | true or false |
| Array | 数组 | Ordered collection of values |
| Map | 映射 | Key-value pairs |
| BigInt | 大整数 | Arbitrary precision integer |
| BigFloat | 大浮点数 | Arbitrary precision decimal |
| Chars | 字符数组 | Unicode character array |
| Null | 空值 | Absence of value |
| Void | 无返回值 | No return value |

## Functions (函数)

| English | Chinese | Description |
|---------|---------|-------------|
| Function | 函数 | Reusable block of code |
| Built-in Function | 内置函数 | Predefined functions available globally |
| Closure | 闭包 | Function that captures variables from outer scope |
| Variadic Function | 可变参数函数 | Function accepting variable number of arguments |
| Anonymous Function | 匿名函数 | Function without a name |
| Callback | 回调函数 | Function passed as an argument |
| Return Value | 返回值 | Value returned by a function |
| Parameter | 参数 | Variable in function declaration |
| Argument | 实参 | Value passed to a function call |

## Variables (变量)

| English | Chinese | Description |
|---------|---------|-------------|
| Variable | 变量 | Named storage location |
| Scope | 作用域 | Region where a variable is accessible |
| Global Variable | 全局变量 | Variable accessible everywhere |
| Local Variable | 局部变量 | Variable accessible only in its function |
| Closure Variable | 闭包变量 | Variable captured by a closure |
| Constant | 常量 | Immutable named value |
| Assignment | 赋值 | Setting a variable's value |
| Declaration | 声明 | Introducing a new variable |
| Shadowing | 遮蔽 | Local variable hiding outer variable with same name |

## Control Flow (控制流)

| English | Chinese | Description |
|---------|---------|-------------|
| Control Flow | 控制流 | Order of statement execution |
| Conditional | 条件语句 | if/else statement |
| Loop | 循环 | Repeated execution (for, while) |
| Iteration | 迭代 | Single execution of a loop body |
| Break | 跳出 | Exit from a loop |
| Continue | 继续 | Skip to next iteration |
| Switch | 分支语句 | Multi-way branch statement |
| Case | 分支 | Individual option in switch |
| Default | 默认分支 | Fallback case in switch |
| Ternary Operator | 三元运算符 | condition ? true : false |

## Object-Oriented Programming (面向对象编程)

| English | Chinese | Description |
|---------|---------|-------------|
| Class | 类 | Blueprint for objects |
| Object | 对象 | Instance of a class |
| Instance | 实例 | Specific object of a class |
| Method | 方法 | Function belonging to a class |
| Property / Field | 属性 / 字段 | Variable belonging to a class |
| Constructor | 构造函数 | Method that initializes new instance |
| Inheritance | 继承 | Class deriving from another class |
| Super | 父类 | Parent class reference |
| This / Self | 当前对象 | Reference to current instance |

## Concurrency (并发)

| English | Chinese | Description |
|---------|---------|-------------|
| Concurrency | 并发 | Simultaneous execution of tasks |
| Goroutine | 协程 | Lightweight thread managed by runtime |
| Tube / Channel | 管道 | Communication channel between goroutines |
| Select | Select语句 | Wait on multiple channel operations |
| Context | 上下文 | Cancellation and timeout mechanism |
| Mutex | 互斥锁 | Mutual exclusion lock |
| WaitGroup | 等待组 | Synchronization primitive |
| Atomic | 原子操作 | Lock-free thread-safe operation |
| Deadlock | 死锁 | Circular waiting between goroutines |
| Race Condition | 竞态条件 | Unsafe concurrent access |

## Error Handling (错误处理)

| English | Chinese | Description |
|---------|---------|-------------|
| Error | 错误 | Exception or failure condition |
| Exception | 异常 | Runtime error that can be caught |
| Try-Catch | 异常捕获 | Error handling block |
| Throw | 抛出 | Raise an exception |
| Finally | 最终块 | Code that runs regardless of error |

## Modules (模块)

| English | Chinese | Description |
|---------|---------|-------------|
| Module | 模块 | Reusable code unit |
| Import | 导入 | Load a module |
| Export | 导出 | Make code available from module |
| Standard Library | 标准库 | Built-in modules |
| Package | 包 | Collection of related modules |

## File Operations (文件操作)

| English | Chinese | Description |
|---------|---------|-------------|
| File | 文件 | Named data storage |
| Directory / Folder | 目录 / 文件夹 | Container for files |
| Path | 路径 | Location of file or directory |
| Stream | 流 | Sequential data access |
| Read | 读取 | Get data from file |
| Write | 写入 | Put data to file |
| Append | 追加 | Add to end of file |
| Open | 打开 | Access a file |
| Close | 关闭 | Release file handle |

## Memory (内存)

| English | Chinese | Description |
|---------|---------|-------------|
| Memory | 内存 | Data storage area |
| Allocation | 分配 | Reserving memory |
| Garbage Collection | 垃圾回收 | Automatic memory reclamation |
| Heap | 堆 | Dynamic memory area |
| Stack | 栈 | LIFO memory area |

## Operators (运算符)

| English | Chinese | Description |
|---------|---------|-------------|
| Operator | 运算符 | Symbol for operation |
| Arithmetic | 算术运算 | +, -, *, /, % |
| Comparison | 比较运算 | ==, !=, <, >, <=, >= |
| Logical | 逻辑运算 | &&, \|\|, ! |
| Assignment | 赋值运算 | =, +=, -=, etc. |
| Increment | 自增 | ++ |
| Decrement | 自减 | -- |

## Performance (性能)

| English | Chinese | Description |
|---------|---------|-------------|
| Performance | 性能 | Execution speed and efficiency |
| Benchmark | 基准测试 | Standardized performance test |
| Optimization | 优化 | Improving performance |
| Tail Call Optimization (TCO) | 尾调用优化 | Converting recursion to iteration |
| Inline Caching | 内联缓存 | Caching method lookups |

---

## Naming Conventions (命名规范)

| Style | Name | Example |
|-------|------|---------|
| `camelCase` | Camel Case | `typeOf`, `getWeb`, `readFile` |
| `PascalCase` | Pascal Case | `Point`, `Array`, `String` |
| `UPPER_SNAKE` | Upper Snake | `MAX_SIZE`, `DEFAULT_PORT` |
| `lower_snake` | Lower Snake | (not used in Xxlang) |

**Note:** All built-in functions and variables use camelCase in Xxlang.

---

## Document Structure (文档结构)

| English | Chinese | Usage |
|---------|---------|-------|
| Table of Contents | 目录 | Document navigation |
| Overview | 概述 | High-level introduction |
| Quick Start | 快速开始 | Getting started guide |
| Reference | 参考 | Detailed API documentation |
| Example | 示例 | Code samples |
| Best Practice | 最佳实践 | Recommended approaches |
| Known Issue | 已知问题 | Documented limitations |
| Changelog | 更新日志 | Version history |

---

## Status

- Created: 2026-03-25
- Last Updated: 2026-03-25
