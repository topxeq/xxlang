# Plan: Refactor Multi-Value Returns to Map

## Overview
Change all functions that return arrays as multi-value containers to return maps instead.

## Target Version
v0.7.0

## Changes Summary

### 1. stdlib/net.go (6 functions) ✅ DONE
- `net.get(url)` → returns `{body, statusCode, statusText}`
- `net.post(url, body, [contentType])` → returns `{body, statusCode, statusText}`
- `net.request(method, url, [body], [headers])` → returns `{body, statusCode, statusText, headers}`
- `net.head(url)` → returns `{statusCode, headers}`
- `net.getJson(url)` → returns `{body, statusCode}`
- `net.postJson(url, body)` → returns `{body, statusCode}`

### 2. stdlib/db.go (1 function) ✅ DONE
- `db.exec(conn, sql, args...)` → returns `{lastInsertId, rowsAffected}`

### 3. stdlib/file.go (1 function) ✅ DONE
- `file.listDirFull(path)` → returns array of `{name, size, isDir, modTime}`

### 4. pkg/objects/builtin.go (3 tube functions) ✅ DONE
- `tubeRecv(tube)` → returns `{value, ok}`
- `tubeTrySend(tube, value)` → returns `{sent, ok}`
- `tubeTryRecv(tube)` → returns `{value, received, open}`

## Implementation Status

- [x] Modify stdlib/net.go
  - [x] net.get
  - [x] net.post
  - [x] net.request
  - [x] net.head
  - [x] net.getJson
  - [x] net.postJson

- [x] Modify stdlib/db.go
  - [x] db.exec

- [x] Modify stdlib/file.go
  - [x] file.listDirFull

- [x] Modify pkg/objects/builtin.go
  - [x] tubeRecv
  - [x] tubeTrySend
  - [x] tubeTryRecv

- [x] Verify all tests pass

- [ ] Update documentation (README.md)

## Files Modified
- `pkg/stdlib/net.go`
- `pkg/stdlib/db.go`
- `pkg/stdlib/file.go`
- `pkg/objects/builtin.go`

## Usage Examples

### net.get / net.post
```xxl
import { get, post } from "net"

var resp = get("https://example.com/api")
if (resp["statusCode"] == 200) {
    pln(resp["body"])
}

var result = post(url, jsonData)
pln("Status:", result["statusCode"])
```

### db.exec
```xxl
import { exec } from "db"

var result = exec(conn, "INSERT INTO users (name) VALUES (?)", "John")
pln("Last ID:", result["lastInsertId"])
pln("Rows affected:", result["rowsAffected"])
```

### file.listDirFull
```xxl
import { listDirFull } from "file"

var entries = listDirFull("/path")
for (var e in entries) {
    pln(e["name"], e["size"], e["isDir"])
}
```

### tube functions
```xxl
var t = makeTube(10)
tubeSend(t, 42)
var result = tubeRecv(t)
if (result["ok"]) {
    pln(result["value"])
}
```

## Breaking Changes
All existing code using array indexing will need to be updated:
- `result[0]` → `result["body"]`
- `result[1]` → `result["statusCode"]`
etc.