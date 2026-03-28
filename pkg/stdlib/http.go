// pkg/stdlib/http.go
// HTTP server utilities for the Xxlang standard library.
// These functions are designed for HTTP server mode.
package stdlib

import (
	"encoding/json"
	"io"
	"mime"
	"path/filepath"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "http",
		Exports: map[string]objects.Object{
			// parseJSON parses JSON body from the request.
			// Usage: http.parseJSON(request)
			"parseJSON": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseJSON() takes exactly 1 argument")
				}

				req, ok := args[0].(*objects.HttpReq)
				if !ok {
					return Error("argument to parseJSON must be HTTP_REQ")
				}

				if req.Value == nil {
					return Error("http request is nil")
				}

				body, err := io.ReadAll(req.Value.Body)
				if err != nil {
					return Error("read request body failed: " + err.Error())
				}

				var data interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					return Error("parse JSON failed: " + err.Error())
				}

				return objects.GoValueToObject(data)
			}),

			// getReqBody reads the raw request body as a string.
			// Usage: http.getReqBody(request)
			"getReqBody": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getReqBody() takes exactly 1 argument")
				}

				req, ok := args[0].(*objects.HttpReq)
				if !ok {
					return Error("argument to getReqBody must be HTTP_REQ")
				}

				if req.Value == nil {
					return Error("http request is nil")
				}

				body, err := io.ReadAll(req.Value.Body)
				if err != nil {
					return Error("read request body failed: " + err.Error())
				}

				return String(string(body))
			}),

			// getReqBodyBytes reads the raw request body and returns as integer array.
			// Usage: http.getReqBodyBytes(request)
			"getReqBodyBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getReqBodyBytes() takes exactly 1 argument")
				}

				req, ok := args[0].(*objects.HttpReq)
				if !ok {
					return Error("argument to getReqBodyBytes must be HTTP_REQ")
				}

				if req.Value == nil {
					return Error("http request is nil")
				}

				body, err := io.ReadAll(req.Value.Body)
				if err != nil {
					return Error("read request body failed: " + err.Error())
				}

				elements := make([]objects.Object, len(body))
				for i, b := range body {
					elements[i] = Int(int64(b))
				}
				return Array(elements...)
			}),

			// getMimeType returns the MIME type for a file path.
			// Usage: http.getMimeType(path)
			"getMimeType": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getMimeType() takes exactly 1 argument")
				}

				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument to getMimeType must be STRING")
				}

				ext := filepath.Ext(path.Value)
				mimeType := mime.TypeByExtension(ext)
				if mimeType == "" {
					mimeType = "application/octet-stream"
				}

				return String(mimeType)
			}),
		},
	})
}
