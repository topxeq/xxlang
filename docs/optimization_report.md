# Xxlang 高并发脚本引擎优化与并发安全修复报告

## 1. Overview

本报告记录 Xxlang 脚本引擎为支持高并发 HTTP 场景所做的优化与并发安全修复。主要目标：高效处理大规模并行请求，同时保证线程安全、防止资源耗尽、确保资源正确释放。

**设计原则**：
- **无硬性限制**：不设并发数上限、不设执行超时（按 owner 要求，之前怀疑问题与并发/超时有关，实际无关）
- **单线程语义不变**：所有修复保证单线程行为与改动前完全一致
- **并发安全**：通过 `go test -race` + 并发压力测试验证

### 问题陈述

原始实现为每个请求创建新 `RegVM` 实例，存在以下瓶颈：

1. **内存泄漏**（严重）：全局对象注册表 `globalRegistry` 持有所有 VM 创建对象的强引用且**从不清理**，导致每次请求产生的对象永不回收（生产环境数小时内吃满内存）
2. **数据竞争**（严重）：全局共享帧池、全局单例回调在并发 VM 执行时互相污染（`go test -race` 证实）
3. **编译开销**：每次请求重复读文件 + 词法分析 + 语法分析 + 编译
4. **内存抖动**：每个 VM 分配 frames/registers/globals/stacks，高并发下 GC 压力大

### 解决方案总览

| 组件 | 文件 | 模式 |
|------|------|------|
| VM 实例池 | `pkg/vm/pool.go` | 对象池（无界，按需创建） |
| 脚本缓存 | `pkg/server/cache.go` | 读写锁缓存 |
| 并发执行器 | `pkg/server/concurrent.go` | errgroup（无超时） |
| 对象注册表修复 | `pkg/vm/value.go` | 加锁 + copy-on-write + 活跃计数清理 |
| 帧池修复 | `pkg/vm/reg_frame.go` | 全局池 → per-VM 私有池 |
| Builtin 回调修复 | `pkg/vm/builtins.go` | 全局单例 → per-VM 私有表 |
| 闭包回调修复 | `pkg/vm/closure.go` | 闭包绑定创建它的 VM |
| pprof 监控 | `pkg/server/server.go` | `-pprof=<port>` 参数 |

---

## 2. 架构

### 2.1 VM 实例池（无界）

M 个请求映射到 N 个 VM：池用小的缓冲 channel（容量 4）缓存空闲 VM，池空时**按需创建新 VM**，不阻塞、不设并发上限。

```
Request 1 ──┐
Request 2 ──┤
Request 3 ──┼──► VMPool (unbounded) ──► VM_1, VM_2, ..., VM_N (created on demand)
...         │
Request M ──┘
```

### 2.2 脚本缓存：编译一次，运行多次

脚本编译结果 `*compiler.Bytecode` 缓存共享（不可变，并发读安全），避免每请求重新编译。文件变更时通过 mod-time 检测热重载。

### 2.3 并发安全核心设计

并发安全的关键在于**消除 VM 之间的共享可变状态**：

| 原共享状态 | 问题 | 修复 |
|-----------|------|------|
| `objectRegistry` 全局单例 | 泄漏 + 并发竞争 | 加锁 + copy-on-write + `BeginExecution`/`EndExecution` 计数，最后一个执行者结束时清理 |
| `regFramePool` 全局 sync.Pool | 帧跨 VM 复用，寄存器/返回值错乱 | 每 VM 私有 `framePool`，帧带归属 |
| `runCodeImpl`/`loadPluginImpl`/`callUserFuncImpl`/`delegateImpl` 全局回调 | 并发 VM 互相覆盖，回调在错误 VM 执行 | 每 VM 私有 builtin 表 + 闭包绑定 VM |
| `vmGlobals` 在 `BeginExecution` 之前注册 | 并发下被另一请求提前 Clear | `BeginExecution` 必须位于任何 `NewObject` 之前 |

---

## 3. 实现细节

### 3.1 VM 实例池 (`pkg/vm/pool.go`)

```go
type VMPool struct {
    pool     chan *RegVM   // 小缓冲(4)缓存空闲 VM
    newFunc  func(*compiler.Bytecode) *RegVM
    bytecode *compiler.Bytecode
    closed   bool
}
```

- **`NewVMPoolWithBytecode(bytecode)`**: 无界池，池空时按需创建 VM
- **`Acquire()`**: 有缓存则复用，否则新建；**从不阻塞**
- **`Release(vm)`**: `resetVM()` 清状态后归还（缓冲满则丢弃，交给 GC）
- **`Close()`**: 关闭池

**`resetVM` 重置项**：globals 置 `ValueNull`、frameIndex 重置、主帧 IP/寄存器/局部变量清理、pendingException、tempStack、nextGlobalIndex。

### 3.2 脚本缓存 (`pkg/server/cache.go`)

```go
type CacheEntry struct {
    Bytecode   *compiler.Bytecode  // 不可变，共享安全
    ModTime    time.Time           // 热重载检测
    SourcePath string
}
```

- **`GetOrLoad(path)`**: 读锁快路径 + 写锁慢路径（double-check 防 thundering herd）
- **`UpdateScript(path)`**: mod-time 变更则重编译并原子替换
- **选择 `sync.RWMutex` 而非 `sync.Map`**：热重载需要原子替换整条记录；同 key 高竞争场景 RWMutex 语义更强

### 3.3 并发执行器 (`pkg/server/concurrent.go`)

每个脚本流程：

```
1. cache.GetOrLoad(path)   → *compiler.Bytecode
2. pool.Acquire()          → *RegVM（无阻塞）
3. vm.Run()                → 执行脚本（无超时）
4. pool.Release(vm)        → 重置 + 归还
5. 结果聚合                → errgroup
```

- **无超时**：脚本运行到完成（设计决策，避免硬性限制）
- **panic 恢复**：`defer/recover` 包裹，单个脚本 panic 不拖垮整个 errgroup
- **errgroup**：自动等待、错误传播、结果按索引写入无竞争

### 3.4 对象注册表修复 (`pkg/vm/value.go`) —— 内存泄漏根因

**根因**：`objectRegistry` 全局单例，所有 VM 的 `NewObject()` 注册对象都存其中；请求结束后 VM 可被 GC，但对象仍被注册表强引用 → 永不回收；`nextIdx` 无限增长，`objects` slice 不断翻倍（4096→8192→...）。

**修复**：
```go
type objectRegistry struct {
    mu      sync.Mutex
    objsPtr unsafe.Pointer // *[]unsafe.Pointer，扩容/清空时原子替换（copy-on-write）
    freeIdx []int
    nextIdx int32
}
```

- 写路径（register/release/Clear）加锁
- 读路径（get）无锁：原子加载 slice 指针 + 原子加载槽位（copy-on-write 保证并发读者看到稳定数组）
- **`BeginExecution()`/`EndExecution()` 活跃计数**：最后一个 VM 结束（计数归零）时 `ClearRegistry()`，对象全部可回收
- **关键约束**：`EndExecution` 必须在调用方**消费完结果之后**执行（如 `RunScriptOnHttp` 的 `defer`）；不能放进 `RegVM.Run()` 内部，否则调用方 `LastResult().ToObject()` 读到被清空的索引返回 null

**调用点**：
| 路径 | Begin/End 位置 |
|------|---------------|
| HTTP 请求 | `RunScriptOnHttp` 开头（任何 NewObject 之前）/ `defer` 结尾 |
| interpreter/REPL | `evalRegister` / `evalFileRegister` |
| 脚本 goroutine | `OpRegRunStart` 启动的 goroutine 内 |

### 3.5 帧池修复 (`pkg/vm/reg_frame.go`) —— 数据竞争

**根因**：`regFramePool` 全局 `sync.Pool`。并发 VM 执行时 A 释放的帧被 B 复用 → 寄存器/返回值错乱（race detector 报 `LastResult()` 读帧与另一 VM 写帧冲突）。

**修复**：per-VM 私有帧池
```go
type RegFrame struct {
    ...
    pool *sync.Pool  // 归属 VM 的私有池；nil 表示不回收
}
func (vm *RegVM) newFrame(fn *compiler.CompiledFunction) *RegFrame {
    f := vm.framePool.Get().(*RegFrame)
    f.pool = vm.framePool
    initFrame(f, fn)
    return f
}
```
所有 11 处 `NewRegFrame` 调用点改为 `vm.newFrame`。

### 3.6 Builtin 回调修复 (`pkg/vm/builtins.go`、`closure.go`) —— 数据竞争

**根因**：`objects` 包 4 个全局单例回调（`runCodeImpl`/`loadPluginImpl`/`callUserFuncImpl`/`delegateImpl`）由 `RegVM.Run()` 每次设置/恢复。并发 VM 互相覆盖 → `mapArray`/`filterArray` 的回调在错误 VM 上执行，或直接报 "callback not available"。

**修复（三层）**：
1. **per-VM builtin 表**：`buildVMBuiltinFuncs(vm)` 复制全局表，将 runCode/delegate/loadPlugin 替换为绑定本 VM 的闭包；`handleRegBuiltin`/`callBuiltin` 从 `vm.builtinFuncs` 取
2. **闭包绑定 VM**：`Closure` 实现 `objects.UserFuncCaller` 接口，`bindVMContext(vm)` 记录回调；`objects.CallUserFunc` 优先走闭包上下文
3. **裸函数包装**：编译器对无自由变量的 lambda 直接传 `*compiler.CompiledFunction`（不生成 Closure），`handleRegBuiltin` 收集参数时将其包装为 VM 绑定 Closure

### 3.7 pprof 监控 (`pkg/server/server.go`、`cmd/xxl/main.go`)

- `Config` 加 `PprofPort`；`-pprof=<port>` 启动独立 pprof HTTP 服务
- 部署命令：`xxl serve ... -pprof=6060`

---

## 4. 线程安全分析

| 组件 | 机制 |
|------|------|
| VM 池 | channel 操作本身线程安全 |
| 脚本缓存 | RWMutex：读并发 / 写独占 |
| 对象注册表 | 写锁 + 读无锁（copy-on-write 原子指针） |
| 帧池 | per-VM 私有，无跨 VM 共享 |
| builtin 表 | per-VM 私有副本 |
| 闭包回调 | 绑定创建 VM，接口分派 |
| errgroup | 线程安全；结果按索引写入无竞争 |

**验证**：`go test -race -run "TestConcurrent" ./pkg/vm/` 全部 PASS（含 24 worker × 40 轮 mapArray/filterArray 闭包回调并发测试）。

---

## 5. 资源清理

| 资源 | 清理机制 |
|------|---------|
| 注册表对象 | 最后一个 VM 结束时 `ClearRegistry()` |
| VM 实例 | `pool.Close()` 或 GC |
| 缓存字节码 | `cache.Clear()` 或进程退出 |
| 脚本 panic | `defer/recover` |
| multipart 临时文件 | `cleanupRequest`（defer 调用 `RemoveAll`） |

---

## 6. 性能特征

### 修复前后对比（生产实测，100 并发压测）

| 指标 | 修复前 | 修复后 |
|------|--------|--------|
| RSS（空闲） | 持续增长不回落（泄漏） | **完全回落** |
| 注册表存活对象 | 55MB（21 万槽位） | **0.00 MB** |
| 压测峰值 RSS | 峰值后不回落 | 峰值后回落（正常） |
| HeapReleased | 低（不归还 OS） | **大量归还 OS** |
| goroutine | 5 | 5（稳定） |

### 优化前后单请求成本

| 操作 | 优化前 | 优化后 |
|------|--------|--------|
| 脚本加载+编译 | ~10-50ms | 0（缓存命中） |
| VM 分配 | ~100μs | ~1μs（池复用） |
| 执行 | ~1-10ms | ~1-10ms |

---

## 7. 权衡与限制

1. **无界 VM 创建**：极端并发下内存随并发数增长（设计选择：不做硬性限制）
2. **无超时**：死循环脚本会一直运行（owner 明确要求移除超时）
3. **无自动池扩容**：VM 按需创建并缓存，不动态调整
4. **注册表读路径**：仍有一次原子指针加载（相对原始无锁实现有微小开销，换取并发安全）
5. **JIT 模式**：JIT 路径不使用 per-VM builtin 表，runCode/mapArray 等回调在 JIT 模式下原本就不可用（改动前亦如此），无回归

---

## 8. 文件清单

| 文件 | 变更 |
|------|------|
| `pkg/vm/pool.go` | 新增：无界 VM 实例池 |
| `pkg/server/cache.go` | 新增：脚本缓存 |
| `pkg/server/concurrent.go` | 新增：并发执行器 |
| `pkg/vm/value.go` | 修改：注册表加锁+copy-on-write+Begin/EndExecution |
| `pkg/vm/reg_frame.go` | 修改：全局帧池→per-VM 帧池 |
| `pkg/vm/reg_vm.go` | 修改：newFrame、私有 builtin、闭包包装、goroutine 保护 |
| `pkg/vm/builtins.go` | 修改：buildVMBuiltinFuncs |
| `pkg/vm/closure.go` | 修改：UserFuncCaller + bindVMContext |
| `pkg/objects/builtin.go` | 修改：CallUserFunc 优先闭包上下文 |
| `pkg/server/server.go` | 修改：BeginExecution + pprof |
| `pkg/interpreter/interpreter.go` | 修改：Begin/End 保护 |
| `cmd/xxl/main.go` | 修改：-pprof 参数 |

---

## 9. 回归验证

```bash
# 全量测试（12 个包）
go test ./pkg/... -count=1

# 并发安全（race detector）
go test -race -run "TestConcurrent" ./pkg/vm/ -count=1

# 生产压测 + 内存验证
seq 1 300 | xargs -P 100 -I {} curl -s -o /dev/null http://localhost:8080/
# 注册表存活 0.00MB，RSS 回落，goroutine 稳定
```

**结论**：单线程语义完全保留（12 包全量测试通过），并发安全与内存泄漏已修复（race 检测零冲突 + 压测内存收敛）。
