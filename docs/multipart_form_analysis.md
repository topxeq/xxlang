# Multipart Form 空间泄漏分析报告

## 1. 问题背景

用户报告部署网站之前出现的问题，怀疑与上传文件时分析 multipart form 后空间未释放有关。

## 2. 代码分析

### 2.1 清理机制现状

`pkg/server/server.go` 中已实现 `cleanupRequest` 函数：

```go
func cleanupRequest(req *http.Request) {
    if req == nil {
        return
    }
    if req.MultipartForm != nil {
        _ = req.MultipartForm.RemoveAll()
    }
    if req.Body != nil {
        _ = req.Body.Close()
    }
}
```

该函数在两个 HTTP handler 入口处通过 `defer` 调用：
- `handleMicroservice` (line 126)
- `handleWebRequest` (line 166)

### 2.2 文件上传 Builtins

`pkg/objects/builtin_file_upload.go` 中多个函数调用 `ParseMultipartForm`：

| 函数 | 调用方式 | maxMemory |
|------|---------|-----------|
| `getFileUploads` | `req.Value.ParseMultipartForm(32 << 20)` | 32MB |
| `getFileUpload` | `req.Value.ParseMultipartForm(32 << 20)` | 32MB |
| `parseMultipartForm` | `ParseMultipartForm(req.Value, maxMemory)` | 用户指定或32MB |
| `saveUploadedFile` | `req.Value.ParseMultipartForm(maxSize)` | 10MB默认 |

### 2.3 潜在泄漏路径

#### 路径 1: `executeScript` 中的 `req.ParseForm()`

```go
// server.go:209
func (s *Server) executeScript(scriptPath string, res http.ResponseWriter, req *http.Request) {
    // ...
    req.ParseForm()  // ← 可能触发 ParseMultipartForm
    // ...
}
```

**分析**：`req.ParseForm()` 对于 `multipart/form-data` 请求会调用 `ParseMultipartForm(32 << 20)`。如果上传文件超过 32MB，会创建临时文件。但 `cleanupRequest` 会在 handler 返回时调用 `RemoveAll()`，所以**不会泄漏**。

#### 路径 2: 脚本中调用文件上传 builtins

当脚本调用 `getFileUploads()` 等函数时：
1. `ParseMultipartForm` 被调用，可能创建临时文件
2. 脚本处理文件
3. 脚本返回
4. `cleanupRequest` 被调用，删除临时文件

**分析**：理论上不会泄漏。但存在以下风险：

#### 路径 3: Panic 导致清理跳过

如果脚本执行过程中发生 panic，`cleanupRequest` 仍然会被 `defer` 调用，所以**不会泄漏**。

#### 路径 4: 非 HTTP 上下文执行

如果通过 `xxl run` 或其他方式在非 HTTP 上下文中执行脚本，没有 `req` 对象，`cleanupRequest` 不会被调用。但如果脚本中没有调用 `ParseMultipartForm`，也不会创建临时文件，所以**不会泄漏**。

### 2.4 已识别的风险点

#### 风险 1: 重复解析

如果脚本多次调用文件上传函数：

```xxl
files1 := getFileUploads(requestG)
files2 := getFileUploads(requestG)  // 重复解析
```

每次调用都会重新解析 multipart form，但 `MultipartForm` 只会被设置一次。`RemoveAll` 会删除所有临时文件，所以**不会泄漏**，但有性能开销。

#### 风险 2: 脚本持有引用

如果脚本将 `FileUpload` 对象存储到全局变量或传递给其他 goroutine：

```xxl
globalFile := getFileUpload(requestG, "file")
// HTTP handler 返回后，globalFile 仍然持有临时文件引用
```

当 `cleanupRequest` 删除临时文件后，`globalFile` 的引用变为悬空。后续访问会出错，但**不会泄漏**。

#### 风险 3: 大文件占用磁盘

如果大量并发请求上传大文件，临时文件会占用大量磁盘空间，直到脚本执行完毕。这不是泄漏，但可能导致磁盘空间不足。

**示例计算**：
- 100 并发请求，每个上传 50MB 文件
- 临时文件占用：100 × 50MB = 5GB
- 如果脚本执行时间 10 秒，磁盘空间占用 5GB 持续 10 秒

#### 风险 4: `ParseForm` 与 `ParseMultipartForm` 的交互

`executeScript` 调用 `req.ParseForm()`，这会解析 multipart form。然后脚本可能再次调用 `getFileUploads()`，触发另一次 `ParseMultipartForm()`。

Go 的 `ParseMultipartForm` 实现：
- 如果 `req.MultipartForm` 已经设置，直接返回（不重新解析）
- 所以重复调用不会创建额外的临时文件

**结论**：不会泄漏。

## 3. 根本原因判断

### 结论：`cleanupRequest` 设计正确，不会导致空间泄漏

经过分析，`cleanupRequest` 在以下方面是正确的：

1. **调用时机**：在 `handleMicroservice` 和 `handleWebRequest` 的入口处通过 `defer` 注册，确保无论哪个代码路径都会执行
2. **清理范围**：调用 `MultipartForm.RemoveAll()` 删除所有临时文件
3. **异常安全**：`defer` 确保即使 panic 也会执行

### 可能的问题场景

如果确实存在空间泄漏，可能的原因是：

1. **长时间运行的脚本**：脚本执行时间过长，临时文件长时间占用磁盘空间
2. **高并发大文件上传**：大量并发请求同时上传大文件，临时文件累积
3. **临时目录空间不足**：`os.TempDir()` 返回的目录空间不足

### 建议改进

#### 改进 1: 添加临时文件清理日志

```go
func cleanupRequest(req *http.Request) {
    if req == nil {
        return
    }
    if req.MultipartForm != nil {
        if err := req.MultipartForm.RemoveAll(); err != nil {
            log.Printf("cleanupRequest: RemoveAll failed: %v", err)
        }
    }
    // ...
}
```

#### 改进 2: 监控临时文件数量

```go
func cleanupRequest(req *http.Request) {
    if req == nil {
        return
    }
    if req.MultipartForm != nil {
        fileCount := len(req.MultipartForm.File)
        if fileCount > 0 {
            log.Printf("cleanupRequest: removing %d file uploads", fileCount)
        }
        _ = req.MultipartForm.RemoveAll()
    }
    // ...
}
```

#### 改进 3: 使用自定义临时目录

```go
// 在 server 启动时设置
os.Setenv("TMPDIR", "/path/to/large/disk")
```

## 4. 总结

| 检查项 | 状态 | 说明 |
|--------|------|------|
| `cleanupRequest` 注册 | ✅ 正确 | 在两个 handler 入口都有 `defer` |
| `RemoveAll` 调用 | ✅ 正确 | 检查 `MultipartForm != nil` 后调用 |
| 异常安全 | ✅ 正确 | `defer` 确保 panic 时也会执行 |
| 重复解析 | ✅ 安全 | Go 内部去重，不会创建额外临时文件 |
| 磁盘空间 | ⚠️ 注意 | 高并发大文件上传时可能临时占用大量空间 |

**结论**：multipart form 清理逻辑正确，**不是**内存泄漏的根源。

## 5. 最终根因确认（2026 年 8 月）

经 pprof 排查，部署网站内存吃满的**真正根因**是 `pkg/vm/value.go` 的全局对象注册表 `globalRegistry`：

1. 每次 HTTP 请求创建 VM 执行脚本 → 字符串/数组/Map 等对象全部注册到全局注册表
2. 请求结束、VM 被 GC，但注册表**持有对象的强引用** → 对象永不回收
3. `nextIdx` 无限增长，`objects` slice 不断翻倍扩容（4096→8192→…→21 万槽位）
4. 数小时后内存被吃满（生产实测）

**修复**：`BeginExecution`/`EndExecution` 活跃计数 + 最后一个 VM 结束时 `ClearRegistry()`（详见 `optimization_report.md` §3.4）。修复后注册表存活对象稳定为 **0.00 MB**。

排查时间线回顾：
1. ~~怀疑并发/超时~~ → 已排除（用户要求移除硬性限制，实际无关）
2. ~~怀疑 multipart form 未释放~~ → 本报告排除（清理逻辑正确）
3. ✅ **确认 globalRegistry 永不清理** → 修复，内存收敛
