# Xxlang WebView2 GUI 编程

Xxlang 通过 WebView2 提供 GUI 编程能力，嵌入基于 Chromium 的浏览器引擎渲染 HTML/CSS/JavaScript 界面，同时由 Xxlang 处理后端的业务逻辑。

## 目录

- [概述](#概述)
- [架构](#架构)
- [窗口创建](#窗口创建)
- [双向通信](#双向通信)
- [消息处理](#消息处理)
- [完整示例](#完整示例)
- [API 参考](#api-参考)

## 概述

Xxlang 的 `gui` 模块提供：

- **WebView2 集成** - 嵌入 Edge/Chromium 浏览器
- **双向通信** - Xxlang ↔ JavaScript 消息传递
- **非阻塞事件循环** - 基于轮询的消息处理
- **现代化 Web 前端** - 使用 HTML/CSS/JavaScript 构建 UI
- **纯 Go 实现** - 无 CGO 依赖

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Xxlang 应用程序                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐ │
│  │ Xxlang 代码  │◄──►│ gui 模块     │◄──►│ WebView2        │ │
│  │ (字节码)    │    │ (标准库)    │    │ (Edge Chromium) │ │
│  └─────────────┘    └─────────────┘    └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 窗口创建

### 基本窗口

```xxl
import "gui"

// 创建窗口：标题，宽度，高度
var window = gui.createWindow("我的应用", 800, 600)

// 加载 HTML 内容
var html = `
<!DOCTYPE html>
<html>
<head><title>我的应用</title></head>
<body><h1>你好，Xxlang!</h1></body>
</html>
`
gui.setHTML(window, html)

// 运行消息循环（阻塞）
gui.loop(window)
```

### 非阻塞事件循环

```xxl
import "gui"

var window = gui.createWindow("我的应用", 800, 600)
gui.setHTML(window, html)

// 非阻塞：轮询消息同时执行其他工作
var isRunning = true
while (isRunning && !gui.isClosed(window)) {
    // 处理窗口事件
    gui.poll(window)
    
    // 处理 WebView 消息
    while (gui.hasMessages(window)) {
        var msg = gui.popMessage(window)
        // 处理消息...
    }
    
    // 做其他工作
    doBackgroundTask()
    
    sleep(30)  // 30ms 帧
}

gui.close(window)
```

## 双向通信

### 方向 1: Xxlang → JavaScript

```xxl
// 执行 JavaScript 并获取结果（阻塞）
var title = gui.evalJS(window, "document.title")
pln("页面标题:", title)

// 执行 JavaScript 不等待（非阻塞）
gui.evalJSAsync(window, "updateUI({count: 100})")

// 调用 JavaScript 函数并传递数据
var data = json.encode({value: 42, label: "答案"})
var js = "window.xxlang.onData(" + data + ");"
gui.evalJSAsync(window, js)
```

### 方向 2: JavaScript → Xxlang

**JavaScript 端：**
```javascript
// 发送消息到 Xxlang
window.chrome.webview.postMessage({
    cmd: "buttonClick",
    data: {id: 1, value: "hello"}
});
```

**Xxlang 端：**
```xxl
import "json"

func handleMessage(msg) {
    var data = json.fromJson(msg)
    
    if (data["cmd"] == "buttonClick") {
        pln("按钮被点击:", data["value"])
    }
}

// 在主循环中
while (!gui.isClosed(window)) {
    gui.poll(window)
    
    while (gui.hasMessages(window)) {
        handleMessage(gui.popMessage(window))
    }
    
    sleep(30)
}
```

## 消息处理

### 消息队列

WebView2 维护一个线程安全的消息队列：

| 函数 | 说明 |
|------|------|
| `gui.hasMessages(handle)` | 检查是否有等待的消息 |
| `gui.popMessage(handle)` | 获取下一条消息（字符串） |
| `gui.poll(handle)` | 处理单个窗口事件 |

### 桥接模式

为了更清晰的 JavaScript → Xxlang 通信，可以创建桥接函数：

**JavaScript:**
```javascript
window.xxlang = {
    call: function(cmd, data) {
        window.chrome.webview.postMessage({
            cmd: cmd,
            data: data || {}
        });
    }
};

// 使用方式
xxlang.call("save", {name: "test", value: 123});
```

**Xxlang:**
```xxl
func handleMessage(msg) {
    var data = json.fromJson(msg)
    var cmd = data["cmd"]
    var payload = data["data"]
    
    if (cmd == "save") {
        saveData(payload["name"], payload["value"])
    }
}
```

## 完整示例

### 蒙特卡洛计算 Pi

这个示例展示了完整的双向通信：

```xxl
// monte_carlo_pi.xxl
import "gui"
import "math"
import "json"
import "array"

var totalPoints = 0
var insidePoints = 0
var isRunning = false
var window = null

func generatePoint() {
    var x = math.random()
    var y = math.random()
    var dist = math.sqrt((x - 0.5)^2 + (y - 0.5)^2)
    return [x, y, dist <= 0.5 ? 1 : 0]
}

func sendPoints(points) {
    var jsonStr = json.encode(points)
    var js = "window.xxlang.onPoints(" + jsonStr + ")"
    gui.evalJSAsync(window, js)
}

func sendStats() {
    var pi = totalPoints > 0 ? 4.0 * insidePoints / totalPoints : 0
    var js = "window.xxlang.onStats({pi:" + toStr(pi) + ",total:" + toStr(totalPoints) + "})"
    gui.evalJSAsync(window, js)
}

func handleMessage(msg) {
    var data = json.fromJson(msg)
    if (data["cmd"] == "start") {
        isRunning = true
    } else if (data["cmd"] == "stop") {
        isRunning = false
    } else if (data["cmd"] == "reset") {
        isRunning = false
        totalPoints = 0
        insidePoints = 0
        gui.evalJSAsync(window, "window.xxlang.onReset()")
    }
}

func createHTML() {
    return `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body { font-family: sans-serif; background: #1a1a2e; color: #eee; padding: 20px; }
#canvas { border: 2px solid #00d4ff; border-radius: 50%; background: #0f0f1a; }
button { background: #00d4ff; color: #1a1a2e; border: none; padding: 8px 16px; margin: 5px; cursor: pointer; }
.stop { background: #ff6b6b; }
</style>
</head>
<body>
<h1>蒙特卡洛 Pi 计算器</h1>
<canvas id="canvas" width="320" height="320"></canvas>
<div id="stats">Pi = <span id="pi">0</span></div>
<div>
    <button onclick="cmd('start')">开始</button>
    <button class="stop" onclick="cmd('stop')">停止</button>
    <button onclick="cmd('reset')">重置</button>
</div>
<script>
const canvas = document.getElementById('canvas');
const ctx = canvas.getContext('2d');

function drawPoint(x, y, inside) {
    ctx.beginPath();
    ctx.arc(x * 320, y * 320, 2, 0, Math.PI * 2);
    ctx.fillStyle = inside ? '#00ff88' : '#ff6b6b';
    ctx.fill();
}

function cmd(c) {
    window.chrome.webview.postMessage({cmd: c});
}

window.xxlang = {
    onPoints: function(pts) {
        for (var i = 0; i < pts.length; i++) {
            drawPoint(pts[i][0], pts[i][1], pts[i][2]);
        }
    },
    onStats: function(d) {
        document.getElementById('pi').textContent = d.pi.toFixed(6);
    },
    onReset: function() {
        ctx.clearRect(0, 0, 320, 320);
    }
};
</script>
</body>
</html>`
}

func main() {
    window = gui.createWindow("Pi 计算器", 600, 500)
    gui.setHTML(window, createHTML())
    
    while (!gui.isClosed(window)) {
        gui.poll(window)
        
        while (gui.hasMessages(window)) {
            handleMessage(gui.popMessage(window))
        }
        
        if (isRunning) {
            var points = []
            for (var i = 0; i < 100; i = i + 1) {
                var p = generatePoint()
                array.push(points, p)
                totalPoints = totalPoints + 1
                if (p[2] == 1) {
                    insidePoints = insidePoints + 1
                }
            }
            sendPoints(points)
            sendStats()
        }
        
        sleep(30)
    }
    
    gui.close(window)
}

main()
```

## API 参考

### 窗口创建

#### `gui.createWindow(title, width, height)`

创建新的 WebView2 窗口。

**参数:**
- `title` (string) - 窗口标题
- `width` (int) - 窗口宽度（像素）
- `height` (int) - 窗口高度（像素）

**返回值:** WebView 句柄或 ERROR

```xxl
var w = gui.createWindow("我的应用", 800, 600)
```

#### `gui.setHTML(handle, html)`

设置 WebView 的 HTML 内容。

**参数:**
- `handle` (WebView) - 窗口句柄
- `html` (string) - HTML 内容

```xxl
gui.setHTML(window, "<h1>你好</h1>")
```

#### `gui.loadURL(handle, url)`

导航到指定 URL。

**参数:**
- `handle` (WebView) - 窗口句柄
- `url` (string) - 要加载的 URL

```xxl
gui.loadURL(window, "https://example.com")
```

### JavaScript 执行

#### `gui.evalJS(handle, script)`

执行 JavaScript 并返回结果（阻塞）。

**参数:**
- `handle` (WebView) - 窗口句柄
- `script` (string) - JavaScript 代码

**返回值:** JavaScript 返回的字符串结果

```xxl
var title = gui.evalJS(window, "document.title")
```

#### `gui.evalJSAsync(handle, script)`

执行 JavaScript 不等待结果（非阻塞）。

**参数:**
- `handle` (WebView) - 窗口句柄
- `script` (string) - JavaScript 代码

```xxl
gui.evalJSAsync(window, "updateUI({count: 100})")
```

### 消息处理

#### `gui.poll(handle)`

处理单个窗口消息（非阻塞）。

**返回值:** TRUE 如果处理了消息，FALSE 否则

```xxl
gui.poll(window)
```

#### `gui.hasMessages(handle)`

检查是否有 WebView 消息等待。

**返回值:** TRUE 如果有可用消息，FALSE 否则

```xxl
if (gui.hasMessages(window)) {
    // 处理消息
}
```

#### `gui.popMessage(handle)`

从队列中获取下一条消息。

**返回值:** 消息字符串（来自 JavaScript 的 JSON）

```xxl
var msg = gui.popMessage(window)
```

### 窗口控制

#### `gui.loop(handle)`

运行窗口消息循环（阻塞）。

```xxl
gui.loop(window)  // 阻塞直到窗口关闭
```

#### `gui.isClosed(handle)`

检查窗口是否已关闭。

**返回值:** TRUE 如果已关闭，FALSE 否则

```xxl
while (!gui.isClosed(window)) {
    // 主循环
}
```

#### `gui.close(handle)`

关闭窗口。

```xxl
gui.close(window)
```

### 工具

#### `gui.getVersion()`

返回已安装的 WebView2 运行时版本。

**返回值:** 版本字符串

```xxl
var version = gui.getVersion()
pln("WebView2 版本:", version)
```

## 另请参阅

- [标准库](STDLIB_zh.md) - 其他标准库模块
- [微服务模式](MICROSERVICE_zh.md) - HTTP/HTTPS 服务器功能
- [服务模式](SERVICE.md) - 作为系统服务运行
