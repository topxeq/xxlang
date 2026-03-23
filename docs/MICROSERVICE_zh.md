# Xxlang 微服务模式

Xxlang 内置 HTTP/HTTPS 服务器，用于构建 Web 应用和微服务。

## 概述

Xxlang 服务器模式提供：

- **HTTP/HTTPS 服务器** - 内置 Web 服务器，支持 SSL
- **微服务路由** - 通过 `/ms/` 和 `/api/` 路径访问 API 端点
- **动态页面** - `.xxl` 脚本作为网页
- **静态文件服务** - 自动处理静态文件
- **WebSocket 支持** - 实时双向通信
- **CORS 支持** - 默认启用跨域资源共享
- **Cookie 支持** - 内置 Cookie 处理

## 启动服务器

### 命令行

```bash
# 在默认端口启动服务器（HTTP: 80, HTTPS: 443）
xxlang serve

# 自定义端口
xxlang serve -http=8080 -https=8443

# 指定 Web 和微服务路径
xxlang serve -web=./www -ms=./api

# 使用 SSL 证书
xxlang serve -cert=./certs
```

### 服务器选项

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `-http=PORT` | HTTP 端口 | 80 |
| `-https=PORT` | HTTPS 端口 | 443 |
| `-web=PATH` | Web 根目录 | `.` |
| `-ms=PATH` | 微服务根目录 | `.` |
| `-cert=PATH` | SSL 证书目录 | `.` |
| `-config=FILE` | 配置文件（JSON） | - |

### 配置文件

```json
{
  "httpPort": 8080,
  "httpsPort": 8443,
  "webPath": "./www",
  "msPath": "./api",
  "certPath": "./certs"
}
```

## 请求处理

### 全局变量

服务器模式下可用的全局变量：

| 变量 | 描述 |
|------|------|
| `requestG` | HTTP 请求对象（HttpReq） |
| `responseG` | HTTP 响应写入器（HttpResp） |
| `paraMapG` | 请求参数（表单 + 查询组合） |
| `methodG` | HTTP 方法（GET、POST 等） |
| `reqUriG` | 请求 URI |
| `reqNameG` | 请求路径最后一段 |
| `webPathG` | Web 根路径 |
| `msPathG` | 微服务根路径 |
| `basePathG` | 基础路径（同 msPathG） |
| `runModeG` | 当前模式（"server"） |

### 请求对象属性

`requestG` 对象提供访问 HTTP 请求数据的属性：

```xxl
// 请求属性（通过成员访问）
var method = requestG.method         // HTTP 方法
var path = requestG.path             // 请求路径
var url = requestG.url               // 完整 URL
var host = requestG.host             // Host 头
var remoteAddr = requestG.remoteAddr // 客户端地址
var proto = requestG.proto           // 协议
var contentLength = requestG.contentLength
var header = requestG.header         // 头作为 Map
```

### 响应函数

```xxl
// 写入响应内容
writeResp(responseG, "Hello, World!")

// 设置状态码
status(responseG, 404)
writeResp(responseG, "Not Found")

// 设置响应头
setRespHeader(responseG, "Content-Type", "application/json")
addRespHeader(responseG, "X-Custom", "value")

// 重定向（默认 302）
redirect(responseG, "/login")
redirect(responseG, "/login", 301)  // 自定义状态码

// 提供静态文件
serveFile(responseG, requestG, "/static/image.png")

// 设置内容类型
setContentType(responseG, "application/json")

// 获取 MIME 类型
var mime = getMimeType("file.json")  // "application/json"

// 获取 HTTP 状态名称
var name = httpStatusName(200)  // "OK"
```

### 结束响应

返回特殊字符串 `"TX_END_RESPONSE_XT"` 表示响应完成：

```xxl
writeResp(responseG, "Hello")
return "TX_END_RESPONSE_XT"
```

## 微服务路由

微服务脚本放在微服务根目录中，通过 `/ms/` 或 `/api/` 路径访问。

### 路由前缀

| 前缀 | 描述 |
|------|------|
| `/ms/*` | 微服务路由 |
| `/api/*` | API 路由（/ms/ 的别名） |
| `/*` | Web 路由（动态 .xxl 页面和静态文件） |

### CORS 头

`/ms/` 和 `/api/` 路由自动启用 CORS：

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

### 目录结构

```
api/
├── user.xxl          # /ms/user 或 /api/user
├── user/
│   ├── profile.xxl   # /ms/user/profile
│   └── settings.xxl  # /ms/user/settings
└── product/
    └── list.xxl      # /ms/product/list
```

### 示例：用户 API

`api/user.xxl`:

```xxl
// GET /api/user
if (methodG == "GET") {
    var users = [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"}
    ]
    setContentType(responseG, "application/json")
    writeResp(responseG, toJson(users))
}

// POST /api/user
if (methodG == "POST") {
    var data = fromJson(getReqBody(requestG))
    // 处理用户创建...
    status(responseG, 201)
    writeResp(responseG, toJson({"success": true, "id": 3}))
}

return "TX_END_RESPONSE_XT"
```

### 示例：多动作路由器

`api/router.xxl`:

```xxl
var action = paraMapG["action"]

switch (action) {
case "add":
    var result = int(paraMapG["a"]) + int(paraMapG["b"])
    setRespHeader(responseG, "Content-Type", "application/json")
    writeResp(responseG, toJson({"result": result}))

case "multiply":
    var result = int(paraMapG["a"]) * int(paraMapG["b"])
    setRespHeader(responseG, "Content-Type", "application/json")
    writeResp(responseG, toJson({"result": result}))

default:
    status(responseG, 400)
    writeResp(responseG, toJson({"error": "未知操作"}))
}

return "TX_END_RESPONSE_XT"
```

## 动态网页

Web 根目录中的 Xxlang 脚本（`.xxl` 文件）作为动态页面执行。

### 示例：动态页面

`www/index.xxl`:

```xxl
var name = paraMapG["name"] || "访客"

setRespHeader(responseG, "Content-Type", "text/html; charset=utf-8")
writeResp(responseG, `
<!DOCTYPE html>
<html>
<head>
    <title>欢迎</title>
</head>
<body>
    <h1>你好，${name}！</h1>
    <p>当前时间：${now()}</p>
</body>
</html>
`)

return "TX_END_RESPONSE_XT"
```

## XHP 动态网页

XHP 文件（`.xhp`）是嵌入 Xxlang 代码的 HTML 文件，类似于 PHP。代码使用 `<?xhp ... ?>` 标签嵌入。

### 基本语法

```html
<html>
<head><title>XHP 页面</title></head>
<body>
<p>结果是：<?xhp return "1" + "2" ?></p>
</body>
</html>
```

### 输出方式

有两种方式从代码块输出内容：

#### 1. 使用 `return`

`return` 语句输出其值并退出当前代码块：

```html
<p><?xhp return "你好，世界！" ?></p>
```

#### 2. 使用 `echo()`

`echo()` 函数输出内容而不退出代码块：

```html
<?xhp echo("你好，") ?><?xhp echo("世界！") ?>
```

### 共享上下文（变量）

**同一 XHP 文件内的所有代码块共享相同的上下文。** 在一个代码块中定义的变量可以在后续代码块中使用：

```html
<?xhp var greeting = "你好" ?>
<?xhp var name = "世界" ?>
<p><?xhp echo(greeting) ?>，<?xhp echo(name) ?>！</p>
<!-- 输出：<p>你好，世界！</p> -->
```

变量可以在不同代码块之间修改：

```html
<?xhp var counter = 0 ?>
<?xhp counter = counter + 1 ?>
<?xhp counter = counter + 1 ?>
<p>计数器：<?xhp echo(counter) ?></p>
<!-- 输出：<p>计数器：2</p> -->
```

复杂数据也可以共享：

```html
<?xhp var user = {"name": "Alice", "age": 30} ?>
<p>姓名：<?xhp echo(user.name) ?></p>
<p>年龄：<?xhp echo(user.age) ?></p>
```

### 访问请求数据

所有标准全局变量都可用：

| 变量 | 描述 |
|------|------|
| `requestG` | HTTP 请求对象 |
| `responseG` | HTTP 响应写入器 |
| `paraMapG` | 请求参数（查询 + 表单） |
| `methodG` | HTTP 方法 |
| `reqUriG` | 请求 URI |

```html
<p>你好，<?xhp echo(paraMapG["name"] || "访客") ?>！</p>
```

### 完整示例

`www/demo.xhp`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>XHP 演示</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>XHP 动态网页</h1>

    <?xhp var title = "共享变量演示" ?>
    <h2><?xhp echo(title) ?></h2>

    <h2>简单表达式</h2>
    <p>1 + 2 = <?xhp return "1" + "2" ?></p>

    <h2>算术运算</h2>
    <p>10 * 5 = <?xhp return toStr(10 * 5) ?></p>

    <h2>当前时间</h2>
    <p>服务器时间：<?xhp return now() ?></p>

    <h2>参数访问</h2>
    <p>你好，<?xhp echo(paraMapG["name"] || "访客") ?>！</p>

    <h2>共享上下文</h2>
    <?xhp var a = 10 ?>
    <?xhp var b = 20 ?>
    <p>a = <?xhp echo(a) ?>, b = <?xhp echo(b) ?></p>
    <p>a + b = <?xhp echo(a + b) ?></p>

    <h2>使用 echo() 的循环</h2>
    <ul>
    <?xhp
        var items = ["苹果", "香蕉", "樱桃"]
        for (var i = 0; i < len(items); i = i + 1) {
            echo("<li>" + items[i] + "</li>")
        }
    ?>
    </ul>

    <h2>条件输出</h2>
    <?xhp
        if (paraMapG["show"] == "yes") {
            echo("<p style='color:green'>秘密内容！</p>")
        } else {
            echo("<p>添加 ?show=yes 查看内容</p>")
        }
    ?>
</body>
</html>
```

### XHP 与 XXL 动态页面对比

| 特性 | `.xhp`（嵌入式） | `.xxl`（脚本） |
|------|-----------------|----------------|
| 主要用途 | HTML 内嵌代码 | 完整脚本控制 |
| 代码风格 | 内联表达式 | 完整程序 |
| 输出方式 | `return` 或 `echo()` | `writeResp()` 输出 |
| 共享上下文 | 是（所有代码块共享变量） | 不适用 |
| 适用场景 | 模板、简单页面 | 复杂逻辑、API |

### 注意事项

- 同一文件内的所有 `<?xhp ... ?>` 代码块共享相同的上下文（变量）
- 使用 `return` 输出内容并退出代码块
- 使用 `echo()` 输出内容而不退出
- 复杂逻辑建议使用 `.xxl` 脚本

## 内置 HTTP 函数

### 请求函数

| 函数 | 描述 |
|------|------|
| `getReqHeader(request, name)` | 获取请求头值 |
| `getReqHeaders(request)` | 获取所有请求头（Map） |
| `getReqBody(request)` | 获取请求体（字符串） |
| `getReqBodyBytes(request)` | 获取请求体（字节数组） |
| `parseForm(request)` | 解析表单数据，返回 Map |
| `parseJSON(request)` | 解析 JSON 请求体，返回对象 |
| `queryParam(request, name)` | 获取 URL 查询参数 |
| `queryParams(request)` | 获取所有查询参数（Map） |
| `formValue(request, name)` | 获取表单值 |

### 响应函数

| 函数 | 描述 |
|------|------|
| `writeResp(response, content)` | 写入响应内容 |
| `status(response, code)` | 设置 HTTP 状态码 |
| `setRespHeader(response, name, value)` | 设置响应头 |
| `addRespHeader(response, name, value)` | 添加响应头 |
| `redirect(response, url, [code])` | 重定向到 URL（默认 302） |
| `serveFile(response, request, path)` | 提供静态文件 |
| `setContentType(response, mimeType)` | 设置 Content-Type 头 |
| `getMimeType(filename)` | 获取文件的 MIME 类型 |
| `httpStatusName(code)` | 获取状态名称（如 "OK"） |

### Cookie 函数

| 函数 | 描述 |
|------|------|
| `setCookie(response, name, value, [options])` | 设置 Cookie |
| `getCookie(request, name)` | 获取 Cookie 值 |
| `getCookies(request)` | 获取所有 Cookie（Map） |

#### Cookie 选项

`setCookie` 的 `options` 参数是一个 Map：

```xxl
setCookie(responseG, "session", "abc123", {
    "path": "/",
    "domain": "example.com",
    "maxAge": 3600,
    "secure": true,
    "httpOnly": true
})
```

### 工具函数

| 函数 | 描述 |
|------|------|
| `urlEncode(str)` | URL 编码 |
| `urlDecode(str)` | URL 解码 |
| `isHttpReq(obj)` | 检查是否为 HTTP 请求对象 |
| `isHttpResp(obj)` | 检查是否为 HTTP 响应对象 |

## WebSocket 支持

### 服务端 WebSocket

```xxl
// 升级 HTTP 连接到 WebSocket
var ws = webSocket(requestG, responseG)

if (ws == null) {
    status(responseG, 400)
    writeResp(responseG, "WebSocket 升级失败")
    return "TX_END_RESPONSE_XT"
}

// 读取消息：返回 [messageType, data]
var msgResult = ws.readMsg()
var msgType = msgResult[0]    // 1=文本, 2=二进制, 8=关闭
var msgData = msgResult[1]    // 消息内容

// 发送文本消息
ws.sendTextMsg("来自服务器的问候！")

// 发送二进制消息
ws.sendBinaryMsg(bytes([1, 2, 3, 4]))

// 发送关闭帧
ws.sendCloseMsg()

// 关闭连接
ws.close()

// 检查是否已关闭
var closed = ws.isClosed()
```

### WebSocket 消息类型

| 常量 | 值 | 描述 |
|------|-----|------|
| 文本消息 | 1 | 文本数据 |
| 二进制消息 | 2 | 二进制数据 |
| 关闭消息 | 8 | 关闭帧 |
| Ping 消息 | 9 | Ping 帧 |
| Pong 消息 | 10 | Pong 帧 |

### WebSocket 方法

| 方法 | 描述 |
|------|------|
| `ws.readMsg()` | 读取消息，返回 `[messageType, data]` |
| `ws.sendTextMsg(text)` | 发送文本消息 |
| `ws.sendBinaryMsg(bytes)` | 发送二进制消息 |
| `ws.sendCloseMsg()` | 发送关闭帧 |
| `ws.close()` | 关闭连接 |
| `ws.isClosed()` | 检查是否已关闭 |

### 内置 WebSocket 函数

| 函数 | 描述 |
|------|------|
| `webSocket(request, response)` | 升级到 WebSocket，返回 WebSocket 对象 |
| `wsReadMsg(ws)` | 从 WebSocket 读取消息 |
| `wsSendText(ws, text)` | 发送文本消息 |
| `wsSendBinary(ws, bytes)` | 发送二进制消息 |
| `wsSendClose(ws)` | 发送关闭帧 |
| `wsClose(ws)` | 关闭连接 |
| `isWebSocket(obj)` | 检查是否为 WebSocket 对象 |

### 示例：Echo 服务器

`api/ws/echo.xxl`:

```xxl
var ws = webSocket(requestG, responseG)
if (ws == null) {
    status(responseG, 400)
    writeResp(responseG, "WebSocket 升级失败")
    return "TX_END_RESPONSE_XT"
}

var running = true
while (running) {
    var result = ws.readMsg()
    var msgType = result[0]
    var msgData = result[1]

    if (msgType == 1) {
        // 文本消息
        ws.sendTextMsg("回显: " + msgData)
    } else if (msgType == 8) {
        // 关闭帧
        running = false
    }
}

ws.close()
```

## HTTP 客户端（net 模块）

使用 `net` 模块进行 HTTP 客户端操作：

```xxl
import "net"

// 简单 GET
var result = net.get("https://api.example.com/data")
var body = result[0]      // 响应体
var statusCode = result[1] // 状态码
var status = result[2]     // 状态文本

// POST 带请求体
var postResult = net.post("https://api.example.com/create", "data here")

// POST JSON
var jsonResult = net.postJson("https://api.example.com/api", {"key": "value"})

// GET JSON
var data = net.getJson("https://api.example.com/data.json")

// 自定义请求
var customResult = net.request("PUT", "https://api.example.com/item/1", "body", {
    "Authorization": "Bearer token"
})

// 设置超时（秒）
net.setTimeout(30)

// 状态码辅助函数
if (net.isOK(statusCode)) {
    pln("成功!")
}
if (net.isClientError(statusCode)) {
    pln("客户端错误!")
}
```

### net 模块函数

| 函数 | 描述 |
|------|------|
| `net.get(url)` | HTTP GET，返回 `[body, statusCode, status]` |
| `net.post(url, body, [contentType])` | HTTP POST |
| `net.head(url)` | HTTP HEAD 请求 |
| `net.request(method, url, [body], [headers])` | 自定义 HTTP 请求 |
| `net.getJson(url)` | GET 带 Accept: application/json |
| `net.postJson(url, jsonBody)` | POST JSON 内容 |
| `net.download(url)` | 下载内容为字节数组 |
| `net.setTimeout(seconds)` | 设置客户端超时 |
| `net.isOK(code)` | 检查是否为 2xx 状态 |
| `net.isRedirect(code)` | 检查是否为 3xx 状态 |
| `net.isClientError(code)` | 检查是否为 4xx 状态 |
| `net.isServerError(code)` | 检查是否为 5xx 状态 |

## JSON API 示例

### 完整 REST API

`api/todo.xxl`:

```xxl
setContentType(responseG, "application/json")

// 内存存储
var todos = [
    {"id": 1, "title": "学习 Xxlang", "done": false},
    {"id": 2, "title": "构建 API", "done": true}
]
var nextId = 3

// 按方法路由
if (methodG == "GET") {
    // 列出所有待办事项
    writeResp(responseG, toJson({"todos": todos}))
}

if (methodG == "POST") {
    // 创建待办事项
    var data = fromJson(getReqBody(requestG))
    var todo = {
        "id": nextId,
        "title": data.title,
        "done": false
    }
    todos = push(todos, todo)
    nextId = nextId + 1
    status(responseG, 201)
    writeResp(responseG, toJson(todo))
}

if (methodG == "PUT") {
    // 更新待办事项
    var data = fromJson(getReqBody(requestG))
    var id = toInt(paraMapG["id"])
    for (var i = 0; i < len(todos); i = i + 1) {
        if (todos[i].id == id) {
            todos[i].title = data.title || todos[i].title
            todos[i].done = data.done || todos[i].done
        }
    }
    writeResp(responseG, toJson({"success": true}))
}

if (methodG == "DELETE") {
    // 删除待办事项
    var id = toInt(paraMapG["id"])
    var newTodos = []
    for (var i = 0; i < len(todos); i = i + 1) {
        if (todos[i].id != id) {
            newTodos = push(newTodos, todos[i])
        }
    }
    todos = newTodos
    writeResp(responseG, toJson({"success": true}))
}

return "TX_END_RESPONSE_XT"
```

## 错误处理

```xxl
try {
    var data = fromJson(getReqBody(requestG))

    if (data.name == null) {
        status(responseG, 400)
        writeResp(responseG, toJson({"error": "name 是必需的"}))
        return "TX_END_RESPONSE_XT"
    }

    // 处理请求...
    writeResp(responseG, toJson({"success": true}))

} catch (e) {
    status(responseG, 500)
    writeResp(responseG, toJson({"error": e}))
}

return "TX_END_RESPONSE_XT"
```

## 服务器中的并发

使用 goroutine 处理后台任务：

```xxl
// 后台任务
run {
    sleep(1000)
    pln("后台清理完成")
}

// 立即响应
setContentType(responseG, "application/json")
writeResp(responseG, toJson({"status": "processing"}))

return "TX_END_RESPONSE_XT"
```

## 最佳实践

1. **始终设置 Content-Type** 用于 API 响应
2. **正确使用状态码**（200、201、400、404、500）
3. **处理前验证输入**
4. **优雅处理错误** 使用 try/catch
5. **使用 JSON** 作为 API 请求/响应体
6. **保持处理程序专注** - 每个端点一个文件
7. **使用并发** 处理后台任务
8. **记录错误** 用于调试
9. **返回 TX_END_RESPONSE_XT** 表示响应完成
10. **使用 net 模块** 进行 HTTP 客户端操作
