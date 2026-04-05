# Xxlang GUI Programming with WebView2

Xxlang provides GUI programming capabilities through WebView2, embedding a modern Chromium-based browser engine to render HTML/CSS/JavaScript interfaces while Xxlang handles backend logic.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Window Creation](#window-creation)
- [Two-Way Communication](#two-way-communication)
- [Message Handling](#message-handling)
- [Complete Example](#complete-example)
- [API Reference](#api-reference)

## Overview

Xxlang's GUI module (`gui`) provides:

- **WebView2 Integration** - Embedded Edge/Chromium browser
- **Two-Way Communication** - Xxlang ↔ JavaScript messaging
- **Non-Blocking Event Loop** - Poll-based message handling
- **Modern Web Frontend** - HTML/CSS/JavaScript for UI
- **Pure Go Implementation** - No CGO dependencies

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Xxlang Application                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐ │
│  │ Xxlang Code │◄──►│ gui Module  │◄──►│ WebView2        │ │
│  │ (Bytecode)  │    │ (stdlib)    │    │ (Edge Chromium) │ │
│  └─────────────┘    └─────────────┘    └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Window Creation

### Basic Window

```xxl
import "gui"

// Create a window: title, width, height
var window = gui.createWindow("My App", 800, 600)

// Load HTML content
var html = `
<!DOCTYPE html>
<html>
<head><title>My App</title></head>
<body><h1>Hello, Xxlang!</h1></body>
</html>
`
gui.setHTML(window, html)

// Run message loop (blocking)
gui.loop(window)
```

### Non-Blocking Event Loop

```xxl
import "gui"

var window = gui.createWindow("My App", 800, 600)
gui.setHTML(window, html)

// Non-blocking: poll messages while doing other work
var isRunning = true
while (isRunning && !gui.isClosed(window)) {
    // Process window events
    gui.poll(window)
    
    // Handle WebView messages
    while (gui.hasMessages(window)) {
        var msg = gui.popMessage(window)
        // Process message...
    }
    
    // Do other work
    doBackgroundTask()
    
    sleep(30)  // 30ms frame
}

gui.close(window)
```

## Two-Way Communication

### Direction 1: Xxlang → JavaScript

```xxl
// Execute JavaScript and get result (blocking)
var title = gui.evalJS(window, "document.title")
pln("Page title:", title)

// Execute JavaScript without waiting (non-blocking)
gui.evalJSAsync(window, "updateUI({count: 100})")

// Call JavaScript function with data
var data = json.encode({value: 42, label: "answer"})
var js = "window.xxlang.onData(" + data + ");"
gui.evalJSAsync(window, js)
```

### Direction 2: JavaScript → Xxlang

**JavaScript side:**
```javascript
// Send message to Xxlang
window.chrome.webview.postMessage({
    cmd: "buttonClick",
    data: {id: 1, value: "hello"}
});
```

**Xxlang side:**
```xxl
import "json"

func handleMessage(msg) {
    var data = json.fromJson(msg)
    
    if (data["cmd"] == "buttonClick") {
        pln("Button clicked:", data["value"])
    }
}

// In main loop
while (!gui.isClosed(window)) {
    gui.poll(window)
    
    while (gui.hasMessages(window)) {
        handleMessage(gui.popMessage(window))
    }
    
    sleep(30)
}
```

## Message Handling

### Message Queue

WebView2 maintains a thread-safe message queue:

| Function | Description |
|----------|-------------|
| `gui.hasMessages(handle)` | Check if messages are waiting |
| `gui.popMessage(handle)` | Get next message (string) |
| `gui.poll(handle)` | Process single window event |

### Bridge Pattern

For cleaner JavaScript → Xxlang communication, create a bridge:

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

// Usage
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

## Complete Example

### Monte Carlo Pi Calculator

This example demonstrates full two-way communication:

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
<h1>Monte Carlo Pi Calculator</h1>
<canvas id="canvas" width="320" height="320"></canvas>
<div id="stats">Pi = <span id="pi">0</span></div>
<div>
    <button onclick="cmd('start')">Start</button>
    <button class="stop" onclick="cmd('stop')">Stop</button>
    <button onclick="cmd('reset')">Reset</button>
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
    window = gui.createWindow("Pi Calculator", 600, 500)
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

## API Reference

### Window Creation

#### `gui.createWindow(title, width, height)`

Creates a new WebView2 window.

**Parameters:**
- `title` (string) - Window title
- `width` (int) - Window width in pixels
- `height` (int) - Window height in pixels

**Returns:** WebView handle or ERROR

```xxl
var w = gui.createWindow("My App", 800, 600)
```

#### `gui.setHTML(handle, html)`

Sets the HTML content of the WebView.

**Parameters:**
- `handle` (WebView) - Window handle
- `html` (string) - HTML content

```xxl
gui.setHTML(window, "<h1>Hello</h1>")
```

#### `gui.loadURL(handle, url)`

Navigates to a URL.

**Parameters:**
- `handle` (WebView) - Window handle
- `url` (string) - URL to load

```xxl
gui.loadURL(window, "https://example.com")
```

### JavaScript Execution

#### `gui.evalJS(handle, script)`

Executes JavaScript and returns the result (blocking).

**Parameters:**
- `handle` (WebView) - Window handle
- `script` (string) - JavaScript code

**Returns:** String result from JavaScript

```xxl
var title = gui.evalJS(window, "document.title")
```

#### `gui.evalJSAsync(handle, script)`

Executes JavaScript without waiting for result (non-blocking).

**Parameters:**
- `handle` (WebView) - Window handle
- `script` (string) - JavaScript code

```xxl
gui.evalJSAsync(window, "updateUI({count: 100})")
```

### Message Handling

#### `gui.poll(handle)`

Processes a single window message (non-blocking).

**Returns:** TRUE if message processed, FALSE otherwise

```xxl
gui.poll(window)
```

#### `gui.hasMessages(handle)`

Checks if WebView messages are waiting.

**Returns:** TRUE if messages available, FALSE otherwise

```xxl
if (gui.hasMessages(window)) {
    // Handle messages
}
```

#### `gui.popMessage(handle)`

Gets the next message from the queue.

**Returns:** Message string (JSON from JavaScript)

```xxl
var msg = gui.popMessage(window)
```

### Window Control

#### `gui.loop(handle)`

Runs the window message loop (blocking).

```xxl
gui.loop(window)  // Blocks until window closes
```

#### `gui.isClosed(handle)`

Checks if the window is closed.

**Returns:** TRUE if closed, FALSE otherwise

```xxl
while (!gui.isClosed(window)) {
    // Main loop
}
```

#### `gui.close(handle)`

Closes the window.

```xxl
gui.close(window)
```

### Utility

#### `gui.getVersion()`

Returns the installed WebView2 runtime version.

**Returns:** Version string

```xxl
var version = gui.getVersion()
pln("WebView2 version:", version)
```

## See Also

- [Standard Library](STDLIB.md) - Other standard library modules
- [Microservice Mode](MICROSERVICE.md) - HTTP/HTTPS server functionality
- [Service Mode](SERVICE.md) - Running as system service
