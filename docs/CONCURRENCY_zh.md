# Xxlang 并发编程

Xxlang 提供 Go 风格的并发原语，用于构建并发应用程序。

## 目录

- [概述](#概述)
- [Goroutine](#goroutine)
- [Tube（管道）](#tube管道)
- [Select 语句](#select-语句)
- [Context（上下文）](#context上下文)
- [同步原语](#同步原语)
- [内置函数](#内置函数)
- [标准库：concurrent 模块](#标准库concurrent-模块)
- [完整示例](#完整示例工作池)
- [最佳实践](#最佳实践)

## 概述

Xxlang 的并发模型灵感来自 Go 的 CSP（通信顺序进程）模型：

- **Goroutine** - 使用 `run` 关键字的轻量级并发执行
- **Tube（管道）** - 用于 goroutine 之间通信的类型化通道
- **Select** - 在多个 tube 上复用操作
- **Context（上下文）** - 超时和取消管理
- **同步原语** - Mutex、RWMutex、WaitGroup、Once、Cond、AtomicInt

## Goroutine

使用 `run` 关键字启动新的 goroutine：

```xxl
// 运行匿名代码块
run {
    sleep(100)
    pln("后台任务完成")
}

// 运行函数并传递参数
run worker(1, 2, 3)
```

### 示例：并行执行

```xxl
var counter = newAtomic(0)

// 启动多个 goroutine
for (var i = 0; i < 10; i = i + 1) {
    run {
        counter.add(1)
    }
}

sleep(50)
pln("计数器:", counter.load())  // 输出: 计数器: 10
```

### 注意事项

- `run` 语句返回 `null`
- Goroutine 并发且独立运行
- 使用同步原语（WaitGroup、Tube）来协调

## Tube（管道）

Tube 是用于 goroutine 之间通信的类型化通道。

### 创建 Tube

```xxl
// 无类型 tube，带缓冲
var tube = makeTube(10)

// 类型化 tube（元素必须匹配类型）
var intTube = makeTube("INT", 5)
var strTube = makeTube("STRING", 3)

// 非缓冲 tube（同步）
var syncTube = makeTube(0)
```

### 类型化 Tube 支持的类型

| 类型字符串 | 描述 |
|-----------|------|
| `"INT"` | 整数值 |
| `"FLOAT"` | 浮点值 |
| `"STRING"` | 字符串值 |
| `"BOOL"` | 布尔值 |
| `"ARRAY"` | 数组值 |
| `"MAP"` | 映射值 |

### 发送和接收

使用 `<-` 操作符进行 tube 操作：

```xxl
var tube = makeTube(1)

// 发送（阻塞）
tube <- 42

// 接收（阻塞）
var val = <- tube
pln(val)  // 输出: 42
```

### 缓冲与非缓冲

```xxl
// 缓冲 tube - 可以保存多个值
var buffered = makeTube(3)
buffered <- 1
buffered <- 2
buffered <- 3  // 所有发送立即成功

// 非缓冲 tube - 发送者阻塞直到有接收者
var unbuffered = makeTube(0)
run {
    unbuffered <- "hello"  // 阻塞直到被接收
}
var msg = <- unbuffered
```

### 非阻塞操作

```xxl
var tube = makeTube(1)

// TrySend - 非阻塞发送，返回布尔值
var sent = tubeTrySend(tube, 42)
if (sent) {
    pln("发送成功")
}

// TryReceive - 非阻塞接收
var result = tubeTryRecv(tube)
// 返回 [value, received, open]
var value = result[0]
var received = result[1]
var open = result[2]

// 检查 tube 状态
pln("长度:", tubeLen(tube))
pln("容量:", tubeCap(tube))
pln("已关闭:", tubeClosed(tube))
```

### 关闭 Tube

```xxl
var tube = makeTube(2)
tube <- 1
tube <- 2
closeTube(tube)

// 仍然可以接收缓冲的值
var a = <- tube  // 1
var b = <- tube  // 2
var c = <- tube  // null（tube 已关闭且为空）
```

### Tube 方法

| 方法 | 描述 |
|------|------|
| `tube.send(value)` | 发送值（阻塞） |
| `tube.receive()` | 接收值，返回 `[value, ok]` |
| `tube.trySend(value)` | 非阻塞发送，返回 `[sent, ok]` |
| `tube.tryReceive()` | 非阻塞接收，返回 `[value, received, open]` |
| `tube.close()` | 关闭 tube |
| `tube.len()` | 缓冲元素数量 |
| `tube.cap()` | 缓冲容量 |
| `tube.isClosed()` | 检查 tube 是否已关闭 |

## Select 语句

Select 允许同时等待多个 tube 操作。

### 基本 Select

```xxl
var tube1 = makeTube(1)
var tube2 = makeTube(1)

select {
case v = <- tube1:
    pln("来自 tube1:", v)
case v = <- tube2:
    pln("来自 tube2:", v)
}
```

### 带 Default 的 Select（非阻塞）

```xxl
var tube = makeTube(1)

select {
case v = <- tube:
    pln("接收到:", v)
default:
    pln("没有数据")
}
```

### 带 Send 的 Select

```xxl
var tube = makeTube(1)

select {
case tube <- 42:
    pln("发送了 42")
default:
    pln("Tube 已满，无法发送")
}
```

### 多个 Case 带 Default

```xxl
var tube1 = makeTube(1)
var tube2 = makeTube(1)

select {
case v = <- tube1:
    pln("tube1:", v)
case v = <- tube2:
    pln("tube2:", v)
default:
    pln("没有数据")
}
```

## Context（上下文）

Context 为 goroutine 提供超时和取消机制。

### 带超时的 Context

```xxl
var ctx = contextWithTimeout(null, 1000)  // 1 秒超时

run {
    select {
    case <- ctx.done():
        pln("Context 超时")
    }
}

sleep(1100)
pln("已完成:", ctx.isDone())  // true
```

### 可取消的 Context

```xxl
var ctx = contextWithCancel(null)

run {
    select {
    case <- ctx.done():
        pln("已取消!")
    }
}

sleep(100)
ctx.cancel()  // 发出取消信号
```

### 带截止时间的 Context

```xxl
var deadline = now(1) + 5000  // 5 秒后
var ctx = contextWithDeadline(null, deadline)

select {
case <- ctx.done():
    pln("到达截止时间:", ctx.err())
}
```

### 在 Select 中使用 Context

```xxl
var tube = makeTube(1)
var ctx = contextWithTimeout(null, 500)

run {
    sleep(1000)
    tube <- "延迟消息"
}

select {
case v = <- tube:
    pln("接收到:", v)
case <- ctx.done():
    pln("超时!")
}
```

### Context 方法

| 方法 | 描述 |
|------|------|
| `ctx.done()` | 返回当 context 完成时关闭的 tube |
| `ctx.cancel()` | 取消 context |
| `ctx.err()` | 如果已完成，返回错误字符串 |
| `ctx.isDone()` | 返回 context 是否已完成 |
| `ctx.deadline()` | 返回截止时间戳（毫秒）或 null |
| `ctx.deadlineStr()` | 返回格式化的截止时间字符串 |

## 同步原语

### Mutex（互斥锁）

```xxl
var mu = newMutex()
var counter = 0

run {
    mu.lock()
    counter = counter + 1
    mu.unlock()
}

run {
    mu.lock()
    counter = counter + 1
    mu.unlock()
}

sleep(100)
pln("计数器:", counter)  // 2
```

#### Mutex 方法

| 方法 | 描述 |
|------|------|
| `mu.lock()` | 获取锁（阻塞） |
| `mu.unlock()` | 释放锁 |
| `mu.tryLock()` | 尝试获取锁（非阻塞），返回布尔值 |

### RWMutex（读写锁）

```xxl
var rwmu = newRWMutex()
var data = {"name": "test"}

// 多个读者可以同时持有 RLock
rwmu.rLock()
pln(data.name)
rwmu.rUnlock()

// 单个写者 - 独占锁
rwmu.lock()
data.value = 42
rwmu.unlock()
```

#### RWMutex 方法

| 方法 | 描述 |
|------|------|
| `rwmu.lock()` | 获取写锁（独占） |
| `rwmu.unlock()` | 释放写锁 |
| `rwmu.tryLock()` | 尝试获取写锁（非阻塞） |
| `rwmu.rLock()` | 获取读锁（共享） |
| `rwmu.rUnlock()` | 释放读锁 |
| `rwmu.tryRLock()` | 尝试获取读锁（非阻塞） |

### WaitGroup（等待组）

```xxl
var wg = newWaitGroup()
var results = []

for (var i = 0; i < 5; i = i + 1) {
    wg.add(1)
    run {
        results = push(results, i * 2)
        wg.done()
    }
}

wg.wait()  // 等待所有 goroutine 完成
pln("结果:", results)
```

#### WaitGroup 方法

| 方法 | 描述 |
|------|------|
| `wg.add(delta)` | 将 delta 加到计数器 |
| `wg.done()` | 计数器减 1 |
| `wg.wait()` | 阻塞直到计数器为 0 |

### Once（单次执行）

```xxl
var once = newOnce()
var initialized = false

func initFunc() {
    initialized = true
    pln("初始化中...")
}

// 只会执行一次
once.do(initFunc)
once.do(initFunc)  // 无效果
```

#### Once 方法

| 方法 | 描述 |
|------|------|
| `once.do(func)` | 仅执行函数一次 |

### Cond（条件变量）

```xxl
var mu = newMutex()
var cond = newCond(mu)
var ready = false
var data = null

// 等待者
run {
    mu.lock()
    while (!ready) {
        cond.wait()
    }
    pln("获得数据:", data)
    mu.unlock()
}

// 发送信号者
sleep(100)
mu.lock()
data = "hello"
ready = true
cond.signal()
mu.unlock()
```

#### Cond 方法

| 方法 | 描述 |
|------|------|
| `cond.wait()` | 等待信号（必须持有锁） |
| `cond.signal()` | 唤醒一个等待者 |
| `cond.broadcast()` | 唤醒所有等待者 |
| `cond.lock()` | 获取关联的互斥锁 |
| `cond.unlock()` | 释放关联的互斥锁 |

### AtomicInt（原子整数）

```xxl
var counter = newAtomic(0)

// 原子加法返回新值
var v1 = counter.add(5)  // 5
var v2 = counter.add(3)  // 8

// 加载当前值
var current = counter.load()  // 8

// 存储新值
counter.store(100)

// 交换返回旧值
var old = counter.swap(200)  // old = 100, counter = 200

// 比较并交换
var swapped = counter.compareAndSwap(200, 300)  // true, counter = 300
var notSwapped = counter.compareAndSwap(100, 400)  // false, counter = 300
```

#### AtomicInt 方法

| 方法 | 描述 |
|------|------|
| `atomic.add(delta)` | 原子加法，返回新值 |
| `atomic.load()` | 加载当前值 |
| `atomic.store(value)` | 存储新值 |
| `atomic.swap(new)` | 与新值交换，返回旧值 |
| `atomic.compareAndSwap(old, new)` | CAS 操作，返回布尔值 |

## 内置函数

### Tube 函数

| 函数 | 描述 |
|------|------|
| `makeTube([type], buffer)` | 创建新 tube |
| `closeTube(tube)` | 关闭 tube |
| `tubeLen(tube)` | 获取缓冲元素数量 |
| `tubeCap(tube)` | 获取缓冲容量 |
| `tubeClosed(tube)` | 检查 tube 是否已关闭 |
| `tubeSend(tube, value)` | 发送值到 tube |
| `tubeRecv(tube)` | 从 tube 接收，返回 `[value, ok]` |
| `tubeTrySend(tube, value)` | 非阻塞发送，返回 `[sent, ok]` |
| `tubeTryRecv(tube)` | 非阻塞接收，返回 `[value, received, open]` |

### Context 函数

| 函数 | 描述 |
|------|------|
| `newContext()` | 创建后台 context |
| `contextWithTimeout(parent, ms)` | 带超时的 context（毫秒） |
| `contextWithCancel(parent)` | 可取消的 context |
| `contextWithDeadline(parent, timestamp)` | 带截止时间的 context（毫秒时间戳） |
| `contextCancel(ctx)` | 取消 context |
| `contextDone(ctx)` | 获取完成 tube |
| `contextErr(ctx)` | 获取错误字符串 |
| `contextIsDone(ctx)` | 检查是否已完成 |
| `contextDeadline(ctx)` | 获取截止时间戳 |

### 同步函数

| 函数 | 描述 |
|------|------|
| `newMutex()` | 创建新互斥锁 |
| `newRWMutex()` | 创建新读写锁 |
| `newWaitGroup()` | 创建新等待组 |
| `newOnce()` | 创建新单次执行 |
| `newCond([mutex])` | 创建新条件变量（可选 mutex） |
| `newAtomic([initial])` | 创建新原子整数（可选初始值） |

## 标准库：concurrent 模块

导入 `concurrent` 模块获取额外的并发工具：

```xxl
import "concurrent"

// 模块提供与内置函数相同的功能
var tube = concurrent.makeTube(10)
var mu = concurrent.newMutex()
```

## 完整示例：工作池

```xxl
// 使用 tube 和 WaitGroup 实现工作池
var jobs = makeTube(10)
var results = makeTube(10)
var wg = newWaitGroup()

// 工作函数
func worker(id) {
    defer wg.done()

    while (true) {
        var job = tubeTryRecv(jobs)
        if (!job[1]) {
            break  // 没有更多任务
        }
        var n = job[0]
        results <- n * n
    }
}

// 启动 3 个工作者
for (var i = 0; i < 3; i = i + 1) {
    wg.add(1)
    run worker(i)
}

// 发送 5 个任务
for (var i = 1; i <= 5; i = i + 1) {
    jobs <- i
}
closeTube(jobs)

// 等待工作者完成
wg.wait()
closeTube(results)

// 收集结果
var total = 0
while (true) {
    var r = tubeTryRecv(results)
    if (!r[1]) {
        break
    }
    pln("结果:", r[0])
    total = total + r[0]
}

pln("总计:", total)  // 1+4+9+16+25 = 55
```

## 完整示例：生产者-消费者

```xxl
var items = makeTube(5)
var done = makeTube(1)

// 生产者
run {
    for (var i = 1; i <= 10; i = i + 1) {
        items <- i
        pln("生产:", i)
    }
    closeTube(items)
}

// 消费者
run {
    while (true) {
        var item = <- items
        if (item == null) {
            break
        }
        pln("消费:", item)
    }
    done <- 1
}

// 等待消费者
<- done
pln("完成!")
```

## 完整示例：超时模式

```xxl
var tube = makeTube(1)
var ctx = contextWithTimeout(null, 500)

// 模拟慢操作
run {
    sleep(1000)
    tube <- "result"
}

select {
case v = <- tube:
    pln("获得结果:", v)
case <- ctx.done():
    pln("操作超时!")
}
```

## 最佳实践

1. **始终关闭 tube** 当生产者完成时，以防止 goroutine 泄漏
2. **使用 context** 进行超时和取消，而不是手动超时
3. **使用带 default 的 select** 进行非阻塞操作
4. **优先使用缓冲 tube** 当顺序不重要时可以提高吞吐量
5. **使用 WaitGroup** 等待多个 goroutine 完成
6. **使用原子操作** 代替互斥锁处理简单计数器
7. **使用 RWMutex** 当读操作比写操作更频繁时
8. **避免共享内存** - 优先使用 tube 进行通信
