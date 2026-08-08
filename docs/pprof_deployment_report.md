# Xxlang 高并发优化、并发安全修复与 pprof 监控部署报告

## 1. 完成的工作

### 1.1 VM 实例池 (`pkg/vm/pool.go`)

**设计变更**：
- 从固定大小池改为**无界池**（按需创建 VM，不阻塞、无并发上限）
- 使用小缓冲（4）缓存空闲 VM，减少 GC 压力
- 无超时限制，脚本运行到完成

### 1.2 并发执行器 (`pkg/server/concurrent.go`)

- 无执行超时、无并发数限制（按 owner 要求移除一切硬性限制）
- 保留 panic 恢复和 errgroup 聚合

### 1.3 脚本缓存 (`pkg/server/cache.go`)

- `sync.RWMutex` 保护的线程安全缓存
- 热重载支持（文件变更检测）

### 1.4 并发安全修复（核心成果）

生产环境内存泄漏 + 数据竞争的**真正根因**（非并发/超时/multipart）：

| 根因 | 修复 |
|------|------|
| `globalRegistry` 对象注册表永不清理（**内存泄漏**） | `BeginExecution`/`EndExecution` 计数，最后一个 VM 结束时 `ClearRegistry()`；加锁 + copy-on-write |
| 全局帧池跨 VM 复用（**数据竞争**） | per-VM 私有帧池 |
| 全局单例 builtin 回调互相覆盖（**数据竞争**） | per-VM 私有 builtin 表 + 闭包绑定 VM |

详细设计见 `docs/optimization_report.md`。

### 1.5 pprof 监控端口

- `-pprof=<port>` 命令行参数，独立 pprof HTTP 服务器
- 支持 heap/goroutine/allocs/block/mutex 等 profile

---

## 2. Multipart Form 分析结论

**结论：multipart form 清理正确，不是内存泄漏根源。**

`cleanupRequest` 在两个 handler 入口 defer 调用，`MultipartForm.RemoveAll()` + `req.Body.Close()` 均正确执行，panic 也会触发（defer）。真正的泄漏根源是 `globalRegistry`（见 §1.4）。

---

## 3. pprof 监控部署

### 配置变更

**文件**: 启动包装脚本（示例：`/usr/local/bin/xxl-server-launcher`）
```bash
# 在 serve 命令后追加 -pprof=<port> 启用监控端口
# （端口仅监听 localhost，勿暴露公网）
exec "$BINARY" serve -web=/path/to/webroot -ms=/path/to/msroot -http=8080 -https=0 -pprof=6060
```

### 访问方式

```bash
# pprof 索引页面
curl http://localhost:6060/debug/pprof/

# Heap profile（文本格式）
curl http://localhost:6060/debug/pprof/heap?debug=1

# Goroutine 详情
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# 用 go tool pprof 图形化分析
go tool pprof http://localhost:6060/debug/pprof/heap
```

### ⚠️ heap profile 首行格式（易踩坑）

```
heap profile: N: M [N2: M2] @ heap/1048576
              ↑存活  ↑历史累计（一直增长，不能当内存用！）
```

- **M = 当前存活字节数**（inuse_space）——监控内存用这个
- **M2 = 历史累计分配**（total_alloc）——只增不减，误用会导致假告警

### 监控脚本 `check-xxl-server`（部署于监控主机）

```bash
check-xxl-server              # 基本检查（状态/内存/goroutine/HTTP）
check-xxl-server -j           # JSON 输出
check-xxl-server -q           # 静默模式（退出码 0/1/2/3）
check-xxl-server -w 500 -c 2000   # 自定义警告/严重阈值(MB)
check-xxl-server -p 6060 -h 8080  # 自定义端口
```

解析 heap 首行取**当前存活** M：
```bash
total_bytes=$(curl -s "http://localhost:6060/debug/pprof/heap?debug=1" \
    | head -1 | grep -oP 'heap profile: \d+:\s+\K\d+')
```

---

## 4. 修复后内存状态（实测）

### 空闲状态

```
HeapAlloc:     8 MB      ← 当前存活堆对象
HeapInuse:    10 MB      ← 实际占用堆
HeapReleased: 578 MB     ← 已归还操作系统（关键指标）
NumGC:        803 次     ← GC 正常工作
注册表存活:   0.00 MB    ← 无泄漏
Goroutines:   5 个       ← 无 goroutine 泄漏
RSS:          58 MB      ← 修复前持续增长（吃满内存）
```

### 压测前后（100 并发 × 300 请求）

```
压测前:   RSS 53MB
压测中:   RSS 峰值（并发编译+执行）
空闲30s:  RSS 完全回落
```

### 系统级验证

```
之前（泄漏时）:  used 持续增长，系统内存告急
现在（修复后）:  used 稳定低位，available 充足
注意: free 列显示偏低是正常的（buff/cache 占大头，可回收），看 available 才准确
```

---

## 5. 监控建议

### 5.1 定期检查（cron 或 watchdog）

```bash
check-xxl-server -q || echo "Xxlang server 异常，退出码 $?"
```

### 5.2 告警脚本（正确解析，勿用累计值）

```bash
# 当 heap 当前存活超过 1GB 时告警
HEAP=$(curl -s http://localhost:6060/debug/pprof/heap?debug=1 \
    | head -1 | grep -oP 'heap profile: \d+:\s+\K\d+')
if [ "${HEAP:-0}" -gt 1073741824 ]; then
    echo "Warning: heap inuse > 1GB" | mail -s "xxl memory alert" admin@example.com
fi
```

### 5.3 监控要点

| 指标 | 命令 | 关注点 |
|------|------|--------|
| 堆存活 | `heap?debug=1` 首行 M | 空闲后应回落 |
| 注册表 | pprof top 中 `objectRegistry` | 应 ≈ 0 |
| goroutine | `goroutine?debug=2` | 数量应稳定 |
| RSS | `ps aux` | 空闲后应回落 |
| HeapReleased | MemStats | 大 = 内存归还 OS 正常 |

> 安全提醒：pprof 端口只监听 localhost，勿暴露公网（heap/goroutine 含敏感信息）。

---

## 6. 文件清单

| 文件 | 变更 |
|------|------|
| `pkg/vm/pool.go` | 新增：无界 VM 池 |
| `pkg/server/cache.go` | 新增：脚本缓存 |
| `pkg/server/concurrent.go` | 新增：并发执行器 |
| `pkg/vm/value.go` | 修改：注册表修复 |
| `pkg/vm/reg_frame.go` | 修改：per-VM 帧池 |
| `pkg/vm/reg_vm.go` | 修改：并发安全 |
| `pkg/vm/builtins.go` | 修改：私有 builtin 表 |
| `pkg/vm/closure.go` | 修改：闭包绑定 VM |
| `pkg/objects/builtin.go` | 修改：CallUserFunc 接口分派 |
| `pkg/server/server.go` | 修改：pprof + BeginExecution |
| `pkg/interpreter/interpreter.go` | 修改：Begin/End 保护 |
| `cmd/xxl/main.go` | 修改：-pprof 参数 |
| `check-xxl-server` | 新增：监控脚本（部署在监控主机） |
| 启动包装脚本 | 修改：serve 命令追加 `-pprof=<port>` |

## 7. 验证命令

```bash
# 编译 + 全量测试（在仓库根目录）
go build ./... && go test ./pkg/... -count=1

# 并发安全
go test -race -run "TestConcurrent" ./pkg/vm/ -count=1

# 服务状态 + 内存
systemctl status xxl-server
check-xxl-server

# pprof 访问
curl http://localhost:6060/debug/pprof/
```
