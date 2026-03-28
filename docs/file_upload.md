# File Upload Handling in Xxlang Server Mode

This document describes the file upload functionality available in Xxlang server mode.

## Overview

Xxlang provides comprehensive file upload handling capabilities for HTTP servers, including:

- **FileUpload object**: Represents an uploaded file with methods for accessing file info and saving
- **Built-in functions**: For parsing multipart forms, saving files, and validating uploads
- **Security features**: Path validation, file size limits, extension filtering

## FileUpload Object

The `FileUpload` object wraps an uploaded file from an HTTP request.

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `filename` | string | Original filename |
| `size` | int | File size in bytes |
| `contentType` | string | MIME type of the file |
| `extension` | string | File extension (without dot) |
| `header` | map | HTTP headers for the file |

### Methods

| Method | Return Type | Description |
|--------|-------------|-------------|
| `filename()` | string | Get original filename |
| `size()` | int | Get file size in bytes |
| `extension()` | string | Get file extension |
| `contentType()` | string | Get MIME type |
| `save(path)` | string | Save file to specified path |
| `saveToDir(dir, autoRename)` | string | Save to directory with optional auto-rename |
| `read()` | string | Read file content as string |
| `readBytes()` | BytesBuffer | Read file content as bytes |
| `hashSHA256()` | string | Calculate SHA256 hash |

## Built-in Functions

### getFileUploads(request)

Retrieves all uploaded files from an HTTP request.

```xxl
var files = getFileUploads(requestG)
// files is a map: { "fieldName": [FileUpload, ...], ... }
```

### getFileUpload(request, fieldName)

Gets a specific uploaded file by field name.

```xxl
var file = getFileUpload(requestG, "avatar")
if file != null {
    var path = file.saveToDir("./uploads", true)
}
```

### saveFile(fileUpload, path)

Saves an uploaded file to a specified path.

```xxl
var result = saveFile(file, "/path/to/save/file.txt")
if result.success {
    pln("File saved to: " + result.path)
}
```

### saveFileToDir(fileUpload, dir, autoRename)

Saves an uploaded file to a directory.

```xxl
var result = saveFileToDir(file, "./uploads", true)
```

### readFile(fileUpload)

Reads file content as a string.

```xxl
var content = readFile(file)
```

### readFileBytes(fileUpload)

Reads file content as a BytesBuffer.

```xxl
var bytes = readFileBytes(file)
```

### fileHashSHA256(fileUpload)

Calculates SHA256 hash of file content.

```xxl
var hash = fileHashSHA256(file)
```

### parseMultipartForm(request, maxMemory)

Parses a multipart form and returns both values and files.

```xxl
var result = parseMultipartForm(requestG, 32 * 1024 * 1024)
var values = result["values"]
var files = result["files"]
```

### safePath(baseDir, filename)

Validates and returns a safe file path (prevents directory traversal attacks).

```xxl
var path = safePath("./uploads", userFilename)
```

### validateFile(fileUpload, maxSize, allowedExtensions...)

Validates a file against size and extension constraints.

```xxl
if validateFile(file, 10*1024*1024, "jpg", "png", "gif") {
    // File is valid
}
```

## HttpReq Extensions

The `HttpReq` object has been extended with file-related members:

### Request Members

| Member | Type | Description |
|--------|------|-------------|
| `body` | string | Raw request body |
| `files` | map | Uploaded files (fieldName -> FileUpload array) |

### Accessing Files

```xxl
// Method 1: Using the files member
var files = requestG.files
for fieldName, fileArray in files {
    for file in fileArray {
        pln("File: " + file.filename())
    }
}

// Method 2: Using getFileUploads function
var files = getFileUploads(requestG)
```

## FileUploadResult Object

The `FileUploadResult` object represents the result of a file upload operation.

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `success` | bool | Whether operation succeeded |
| `message` | string | Result message or error |
| `filePath` | string | Path where file was saved |
| `originalName` | string | Original filename |
| `size` | int | File size in bytes |

### Methods

| Method | Return Type | Description |
|--------|-------------|-------------|
| `success()` | bool | Check if successful |
| `message()` | string | Get result message |
| `path()` | string | Get saved file path |
| `originalName()` | string | Get original filename |
| `size()` | int | Get file size |

## Complete Example

```xxl
// file_upload_handler.xxl
// Handle file upload in server mode

// Set response content type
setRespHeader(responseG, "Content-Type", "application/json")

// Get uploaded files
var files = getFileUploads(requestG)

if len(files) == 0 {
    writeResp(responseG, json::encode({
        "success": false,
        "message": "No files uploaded"
    }))
    return
}

// Configure upload settings
var config = {
    "uploadDir": "./uploads",
    "maxSize": 10 * 1024 * 1024,  // 10MB
    "allowedExts": ["txt", "pdf", "png", "jpg", "jpeg", "gif", "doc", "docx"]
}

// Process files
var results = []
for fieldName, fileArray in files {
    for file in fileArray {
        var filename = file.filename()
        var size = file.size()

        // Validate size
        if size > config["maxSize"] {
            results = append(results, {
                "fieldName": fieldName,
                "filename": filename,
                "success": false,
                "error": "File too large"
            })
            continue
        }

        // Validate extension
        var ext = file.extension()
        if !contains(config["allowedExts"], ext) {
            results = append(results, {
                "fieldName": fieldName,
                "filename": filename,
                "success": false,
                "error": "Extension not allowed"
            })
            continue
        }

        // Save file
        var result = saveUploadedFile(requestG, fieldName, config["uploadDir"], {
            "maxSize": config["maxSize"],
            "autoRename": true,
            "allowedExtensions": config["allowedExts"]
        })

        results = append(results, {
            "fieldName": fieldName,
            "filename": filename,
            "success": result.success,
            "path": result.path,
            "error": result.message
        })
    }
}

// Return results
writeResp(responseG, json::encode({
    "success": true,
    "count": len(results),
    "files": results
}))
```

## Security Considerations

1. **Path Traversal Prevention**: Always use `safePath()` when constructing file paths from user input.

2. **File Size Limits**: Always validate file size before processing large files.

3. **Extension Validation**: Restrict allowed file extensions to prevent malicious file uploads.

4. **MIME Type Validation**: Check the content type for additional security.

5. **Auto-Rename**: Use auto-rename to prevent file overwrites and path conflicts.

## Configuration

Default file upload configuration:

```xxl
var uploadConfig = {
    "maxSize": 10 * 1024 * 1024,  // 10MB
    "uploadDir": "./uploads",
    "autoRename": true,
    "allowedExtensions": [],
    "allowedMimeTypes": []
}
```
