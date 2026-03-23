# Xxlang Microservice Mode

Xxlang includes a built-in HTTP/HTTPS server for building web applications and microservices.

## Overview

Xxlang server mode provides:

- **HTTP/HTTPS Server** - Built-in web server with SSL support
- **Microservice Routes** - API endpoints via `/ms/` and `/api/` paths
- **Dynamic Pages** - `.xxl` scripts as web pages
- **Static File Serving** - Automatic static file handling
- **WebSocket Support** - Real-time bidirectional communication
- **CORS Support** - Cross-origin resource sharing enabled by default
- **Cookie Support** - Built-in cookie handling

## Starting the Server

### Command Line

```bash
# Start server on default ports (HTTP: 80, HTTPS: 443)
xxl serve

# Custom ports
xxl serve -http=8080 -https=8443

# Specify web and microservice paths
xxl serve -web=./www -ms=./api

# With SSL certificates
xxl serve -cert=./certs
```

### Server Options

| Option | Description | Default |
|--------|-------------|---------|
| `-http=PORT` | HTTP port | 80 |
| `-https=PORT` | HTTPS port | 443 |
| `-web=PATH` | Web root directory | `.` |
| `-ms=PATH` | Microservice root directory | `.` |
| `-cert=PATH` | SSL certificate directory | `.` |
| `-config=FILE` | Configuration file (JSON) | - |

### Configuration File

```json
{
  "httpPort": 8080,
  "httpsPort": 8443,
  "webPath": "./www",
  "msPath": "./api",
  "certPath": "./certs"
}
```

## Request Handling

### Global Variables

The following global variables are available in server mode:

| Variable | Description |
|----------|-------------|
| `requestG` | HTTP request object (HttpReq) |
| `responseG` | HTTP response writer (HttpResp) |
| `paraMapG` | Request parameters (form + query combined) |
| `methodG` | HTTP method (GET, POST, etc.) |
| `reqUriG` | Request URI |
| `reqNameG` | Last segment of request path |
| `webPathG` | Web root path |
| `msPathG` | Microservice root path |
| `basePathG` | Base path (same as msPathG) |
| `runModeG` | Current mode ("server") |

### Request Object Properties

The `requestG` object provides access to HTTP request data:

```xxl
// Request properties (via member access)
var method = requestG.method         // HTTP method
var path = requestG.path             // Request path
var url = requestG.url               // Full URL
var host = requestG.host             // Host header
var remoteAddr = requestG.remoteAddr // Client address
var proto = requestG.proto           // Protocol
var contentLength = requestG.contentLength
var header = requestG.header         // Headers as Map
```

### Response Functions

```xxl
// Write response content
writeResp(responseG, "Hello, World!")

// Set status code
status(responseG, 404)
writeResp(responseG, "Not Found")

// Set headers
setRespHeader(responseG, "Content-Type", "application/json")
addRespHeader(responseG, "X-Custom", "value")

// Redirect (default 302)
redirect(responseG, "/login")
redirect(responseG, "/login", 301)  // Custom status code

// Serve static file
serveFile(responseG, requestG, "/static/image.png")

// Set content type
setContentType(responseG, "application/json")

// Get MIME type
var mime = getMimeType("file.json")  // "application/json"

// Get HTTP status name
var name = httpStatusName(200)  // "OK"
```

### Ending Response

Return the special string `"TX_END_RESPONSE_XT"` to signal response completion:

```xxl
writeResp(responseG, "Hello")
return "TX_END_RESPONSE_XT"
```

## Microservice Routes

Microservice scripts are placed in the microservice root directory and accessed via `/ms/` or `/api/` paths.

### Route Prefixes

| Prefix | Description |
|--------|-------------|
| `/ms/*` | Microservice routes |
| `/api/*` | API routes (alias for /ms/) |
| `/*` | Web routes (dynamic .xxl pages and static files) |

### CORS Headers

CORS is automatically enabled for `/ms/` and `/api/` routes:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

### Directory Structure

```
api/
├── user.xxl          # /ms/user or /api/user
├── user/
│   ├── profile.xxl   # /ms/user/profile
│   └── settings.xxl  # /ms/user/settings
└── product/
    └── list.xxl      # /ms/product/list
```

### Example: User API

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
    // Process user creation...
    status(responseG, 201)
    writeResp(responseG, toJson({"success": true, "id": 3}))
}

return "TX_END_RESPONSE_XT"
```

### Example: Multi-action Router

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
    writeResp(responseG, toJson({"error": "unknown action"}))
}

return "TX_END_RESPONSE_XT"
```

## Dynamic Web Pages

Xxlang scripts (`.xxl` files) in the web root are executed as dynamic pages.

### Example: Dynamic Page

`www/index.xxl`:

```xxl
var name = paraMapG["name"] || "World"

setRespHeader(responseG, "Content-Type", "text/html; charset=utf-8")
writeResp(responseG, `
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
</head>
<body>
    <h1>Hello, ${name}!</h1>
    <p>The current time is: ${now()}</p>
</body>
</html>
`)

return "TX_END_RESPONSE_XT"
```

## Built-in HTTP Functions

### Request Functions

| Function | Description |
|----------|-------------|
| `getReqHeader(request, name)` | Get request header value |
| `getReqHeaders(request)` | Get all headers as Map |
| `getReqBody(request)` | Get request body as string |
| `getReqBodyBytes(request)` | Get request body as byte array |
| `parseForm(request)` | Parse form data, returns Map |
| `parseJSON(request)` | Parse JSON body, returns object |
| `queryParam(request, name)` | Get URL query parameter |
| `queryParams(request)` | Get all query parameters as Map |
| `formValue(request, name)` | Get form value |

### Response Functions

| Function | Description |
|----------|-------------|
| `writeResp(response, content)` | Write response content |
| `status(response, code)` | Set HTTP status code |
| `setRespHeader(response, name, value)` | Set response header |
| `addRespHeader(response, name, value)` | Add response header |
| `redirect(response, url, [code])` | Redirect to URL (default 302) |
| `serveFile(response, request, path)` | Serve static file |
| `setContentType(response, mimeType)` | Set Content-Type header |
| `getMimeType(filename)` | Get MIME type for file |
| `httpStatusName(code)` | Get status name (e.g., "OK") |

### Cookie Functions

| Function | Description |
|----------|-------------|
| `setCookie(response, name, value, [options])` | Set cookie |
| `getCookie(request, name)` | Get cookie value |
| `getCookies(request)` | Get all cookies as Map |

#### Cookie Options

The `options` parameter for `setCookie` is a Map:

```xxl
setCookie(responseG, "session", "abc123", {
    "path": "/",
    "domain": "example.com",
    "maxAge": 3600,
    "secure": true,
    "httpOnly": true
})
```

### Utility Functions

| Function | Description |
|----------|-------------|
| `urlEncode(str)` | URL encode string |
| `urlDecode(str)` | URL decode string |
| `isHttpReq(obj)` | Check if object is HTTP request |
| `isHttpResp(obj)` | Check if object is HTTP response |

## WebSocket Support

### Server-side WebSocket

```xxl
// Upgrade HTTP connection to WebSocket
var ws = webSocket(requestG, responseG)

if (ws == null) {
    status(responseG, 400)
    writeResp(responseG, "WebSocket upgrade failed")
    return "TX_END_RESPONSE_XT"
}

// Read message: returns [messageType, data]
var msgResult = ws.readMsg()
var msgType = msgResult[0]    // 1=text, 2=binary, 8=close
var msgData = msgResult[1]    // Message content

// Send text message
ws.sendTextMsg("Hello from server!")

// Send binary message
ws.sendBinaryMsg(bytes([1, 2, 3, 4]))

// Send close frame
ws.sendCloseMsg()

// Close connection
ws.close()

// Check if closed
var closed = ws.isClosed()
```

### WebSocket Message Types

| Constant | Value | Description |
|----------|-------|-------------|
| Text Message | 1 | Text data |
| Binary Message | 2 | Binary data |
| Close Message | 8 | Close frame |
| Ping Message | 9 | Ping frame |
| Pong Message | 10 | Pong frame |

### WebSocket Methods

| Method | Description |
|--------|-------------|
| `ws.readMsg()` | Read message, returns `[messageType, data]` |
| `ws.sendTextMsg(text)` | Send text message |
| `ws.sendBinaryMsg(bytes)` | Send binary message |
| `ws.sendCloseMsg()` | Send close frame |
| `ws.close()` | Close connection |
| `ws.isClosed()` | Check if closed |

### Built-in WebSocket Functions

| Function | Description |
|----------|-------------|
| `webSocket(request, response)` | Upgrade to WebSocket, returns WebSocket object |
| `wsReadMsg(ws)` | Read message from WebSocket |
| `wsSendText(ws, text)` | Send text message |
| `wsSendBinary(ws, bytes)` | Send binary message |
| `wsSendClose(ws)` | Send close frame |
| `wsClose(ws)` | Close connection |
| `isWebSocket(obj)` | Check if object is WebSocket |

### Example: Echo Server

`api/ws/echo.xxl`:

```xxl
var ws = webSocket(requestG, responseG)
if (ws == null) {
    status(responseG, 400)
    writeResp(responseG, "WebSocket upgrade failed")
    return "TX_END_RESPONSE_XT"
}

var running = true
while (running) {
    var result = ws.readMsg()
    var msgType = result[0]
    var msgData = result[1]

    if (msgType == 1) {
        // Text message
        ws.sendTextMsg("Echo: " + msgData)
    } else if (msgType == 8) {
        // Close frame
        running = false
    }
}

ws.close()
```

## HTTP Client (net Module)

Use the `net` module for HTTP client operations:

```xxl
import "net"

// Simple GET
var result = net.get("https://api.example.com/data")
var body = result[0]      // Response body
var statusCode = result[1] // Status code
var status = result[2]     // Status text

// POST with body
var postResult = net.post("https://api.example.com/create", "data here")

// POST JSON
var jsonResult = net.postJson("https://api.example.com/api", {"key": "value"})

// GET JSON
var data = net.getJson("https://api.example.com/data.json")

// Custom request
var customResult = net.request("PUT", "https://api.example.com/item/1", "body", {
    "Authorization": "Bearer token"
})

// Set timeout (seconds)
net.setTimeout(30)

// Status code helpers
if (net.isOK(statusCode)) {
    pln("Success!")
}
if (net.isClientError(statusCode)) {
    pln("Client error!")
}
```

### net Module Functions

| Function | Description |
|----------|-------------|
| `net.get(url)` | HTTP GET, returns `[body, statusCode, status]` |
| `net.post(url, body, [contentType])` | HTTP POST |
| `net.head(url)` | HTTP HEAD request |
| `net.request(method, url, [body], [headers])` | Custom HTTP request |
| `net.getJson(url)` | GET with Accept: application/json |
| `net.postJson(url, jsonBody)` | POST JSON content |
| `net.download(url)` | Download content as bytes |
| `net.setTimeout(seconds)` | Set client timeout |
| `net.isOK(code)` | Check if 2xx status |
| `net.isRedirect(code)` | Check if 3xx status |
| `net.isClientError(code)` | Check if 4xx status |
| `net.isServerError(code)` | Check if 5xx status |

## JSON API Example

### Complete REST API

`api/todo.xxl`:

```xxl
setContentType(responseG, "application/json")

// In-memory storage
var todos = [
    {"id": 1, "title": "Learn Xxlang", "done": false},
    {"id": 2, "title": "Build API", "done": true}
]
var nextId = 3

// Route by method
if (methodG == "GET") {
    // List all todos
    writeResp(responseG, toJson({"todos": todos}))
}

if (methodG == "POST") {
    // Create todo
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
    // Update todo
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
    // Delete todo
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

## Error Handling

```xxl
try {
    var data = fromJson(getReqBody(requestG))

    if (data.name == null) {
        status(responseG, 400)
        writeResp(responseG, toJson({"error": "name is required"}))
        return "TX_END_RESPONSE_XT"
    }

    // Process request...
    writeResp(responseG, toJson({"success": true}))

} catch (e) {
    status(responseG, 500)
    writeResp(responseG, toJson({"error": e}))
}

return "TX_END_RESPONSE_XT"
```

## Concurrency in Server

Use goroutines for background tasks:

```xxl
// Background task
run {
    sleep(1000)
    pln("Background cleanup completed")
}

// Respond immediately
setContentType(responseG, "application/json")
writeResp(responseG, toJson({"status": "processing"}))

return "TX_END_RESPONSE_XT"
```

## Best Practices

1. **Always set Content-Type** for API responses
2. **Use status codes** appropriately (200, 201, 400, 404, 500)
3. **Validate input** before processing
4. **Handle errors gracefully** with try/catch
5. **Use JSON** for API request/response bodies
6. **Keep handlers focused** - one file per endpoint
7. **Use concurrency** for background tasks
8. **Log errors** for debugging
9. **Return TX_END_RESPONSE_XT** to signal completion
10. **Use the net module** for HTTP client operations
