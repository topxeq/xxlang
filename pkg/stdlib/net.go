// pkg/stdlib/net.go
// Network utilities for the Xxlang standard library.
package stdlib

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// httpResponse creates a standard HTTP response map
func httpResponse(body string, statusCode int, statusText string) *objects.OrderedMap {
	result := objects.NewOrderedMap()
	result.Set(objects.NewString("body"), objects.NewString(body))
	result.Set(objects.NewString("statusCode"), objects.NewInt(int64(statusCode)))
	result.Set(objects.NewString("statusText"), objects.NewString(statusText))
	return result
}

// httpResponseWithHeaders creates an HTTP response map with headers
func httpResponseWithHeaders(body string, statusCode int, statusText string, headers *objects.OrderedMap) *objects.OrderedMap {
	result := objects.NewOrderedMap()
	result.Set(objects.NewString("body"), objects.NewString(body))
	result.Set(objects.NewString("statusCode"), objects.NewInt(int64(statusCode)))
	result.Set(objects.NewString("statusText"), objects.NewString(statusText))
	result.Set(objects.NewString("headers"), headers)
	return result
}

// httpHeadersToMap converts HTTP headers to OrderedMap
func httpHeadersToMap(header http.Header) *objects.OrderedMap {
	headers := objects.NewOrderedMap()
	for k, v := range header {
		headers.Set(objects.NewString(k), objects.NewString(strings.Join(v, ", ")))
	}
	return headers
}

func init() {
	Register(&Module{
		Name: "net",
		Exports: map[string]objects.Object{
			// HTTP GET - returns {body, statusCode, statusText}
			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("get() takes exactly 1 argument")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("get() requires a string URL")
				}
				resp, err := httpClient.Get(url.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				return httpResponse(string(body), resp.StatusCode, resp.Status)
			}),

			// HTTP POST - returns {body, statusCode, statusText}
			"post": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("post() takes at least 2 arguments")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("post() requires a string URL")
				}
				body, ok := args[1].(*objects.String)
				if !ok {
					return Error("post() requires a string body")
				}
				contentType := "application/json"
				if len(args) > 2 {
					ct, ok := args[2].(*objects.String)
					if ok {
						contentType = ct.Value
					}
				}
				resp, err := httpClient.Post(url.Value, contentType, strings.NewReader(body.Value))
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				return httpResponse(string(respBody), resp.StatusCode, resp.Status)
			}),

			// HTTP request with method - returns {body, statusCode, statusText, headers}
			"request": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("request() takes at least 2 arguments")
				}
				method, ok := args[0].(*objects.String)
				if !ok {
					return Error("request() requires a string method")
				}
				url, ok := args[1].(*objects.String)
				if !ok {
					return Error("request() requires a string URL")
				}
				var body io.Reader
				if len(args) > 2 {
					b, ok := args[2].(*objects.String)
					if ok {
						body = strings.NewReader(b.Value)
					}
				}
				req, err := http.NewRequest(method.Value, url.Value, body)
				if err != nil {
					return Error(err.Error())
				}
				// Add headers if provided
				if len(args) > 3 {
					headers, ok := args[3].(*objects.Map)
					if ok {
						for _, pair := range headers.Pairs {
							key, ok1 := pair.Key.(*objects.String)
							val, ok2 := pair.Value.(*objects.String)
							if ok1 && ok2 {
								req.Header.Set(key.Value, val.Value)
							}
						}
					}
				}
				resp, err := httpClient.Do(req)
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				return httpResponseWithHeaders(string(respBody), resp.StatusCode, resp.Status, httpHeadersToMap(resp.Header))
			}),

			// Head request - returns {statusCode, headers}
			"head": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("head() takes exactly 1 argument")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("head() requires a string URL")
				}
				resp, err := httpClient.Head(url.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				result := objects.NewOrderedMap()
				result.Set(objects.NewString("statusCode"), objects.NewInt(int64(resp.StatusCode)))
				result.Set(objects.NewString("headers"), httpHeadersToMap(resp.Header))
				return result
			}),

			// Download file - returns content (use io.writeFile to save)
			"download": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("download() takes exactly 1 argument")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("download() requires a string URL")
				}
				resp, err := httpClient.Get(url.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					return Error("download failed with status: " + resp.Status)
				}
				data, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(data))
			}),

			// Set timeout
			"setTimeout": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("setTimeout() takes exactly 1 argument")
				}
				timeout, ok := args[0].(*objects.Int)
				if !ok {
					return Error("setTimeout() requires an integer (seconds)")
				}
				httpClient.Timeout = time.Duration(timeout.Value) * time.Second
				return Null()
			}),

			// Status code helpers
			"isOK": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isOK() takes exactly 1 argument")
				}
				code, ok := args[0].(*objects.Int)
				if !ok {
					return Error("isOK() requires an integer")
				}
				return Bool(code.Value >= 200 && code.Value < 300)
			}),

			"isRedirect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isRedirect() takes exactly 1 argument")
				}
				code, ok := args[0].(*objects.Int)
				if !ok {
					return Error("isRedirect() requires an integer")
				}
				return Bool(code.Value >= 300 && code.Value < 400)
			}),

			"isClientError": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isClientError() takes exactly 1 argument")
				}
				code, ok := args[0].(*objects.Int)
				if !ok {
					return Error("isClientError() requires an integer")
				}
				return Bool(code.Value >= 400 && code.Value < 500)
			}),

			"isServerError": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isServerError() takes exactly 1 argument")
				}
				code, ok := args[0].(*objects.Int)
				if !ok {
					return Error("isServerError() requires an integer")
				}
				return Bool(code.Value >= 500)
			}),

			// JSON helpers - returns {body, statusCode}
			"getJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getJson() takes exactly 1 argument")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("getJson() requires a string URL")
				}
				req, err := http.NewRequest("GET", url.Value, nil)
				if err != nil {
					return Error(err.Error())
				}
				req.Header.Set("Accept", "application/json")
				resp, err := httpClient.Do(req)
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				result := objects.NewOrderedMap()
				result.Set(objects.NewString("body"), objects.NewString(string(body)))
				result.Set(objects.NewString("statusCode"), objects.NewInt(int64(resp.StatusCode)))
				return result
			}),

			"postJson": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("postJson() takes exactly 2 arguments")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("postJson() requires a string URL")
				}
				jsonBody, ok := args[1].(*objects.String)
				if !ok {
					return Error("postJson() requires a string JSON body")
				}
				resp, err := httpClient.Post(url.Value, "application/json", strings.NewReader(jsonBody.Value))
				if err != nil {
					return Error(err.Error())
				}
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return Error(err.Error())
				}
				result := objects.NewOrderedMap()
				result.Set(objects.NewString("body"), objects.NewString(string(body)))
				result.Set(objects.NewString("statusCode"), objects.NewInt(int64(resp.StatusCode)))
				return result
			}),
		},
	})
}