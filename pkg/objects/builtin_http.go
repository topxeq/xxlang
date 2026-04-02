// pkg/objects/builtin_http.go
// Built-in functions for HTTP server mode.
package objects

import (
	"net/http"
	"strconv"
	"time"
)

// HttpBuiltins contains all HTTP-related built-in functions.
// These are only available in server mode.
var HttpBuiltins = map[string]*Builtin{
	// writeResp writes content to the HTTP response.
	// Usage: writeResp(response, content)
	"writeResp": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for writeResp. got=%d, want=2", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'writeResp' must be HTTP_RESP, got %s", args[0].Type())
			}

			var content string
			switch c := args[1].(type) {
			case *String:
				content = c.Value
			default:
				content = c.Inspect()
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			_, err := resp.Value.Write([]byte(content))
			if err != nil {
				return newError("write response failed: %v", err)
			}
			resp.SetWritten()

			return NULL
		},
	},

	// setRespHeader sets a response header.
	// Usage: setRespHeader(response, key, value)
	"setRespHeader": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for setRespHeader. got=%d, want=3", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'setRespHeader' must be HTTP_RESP, got %s", args[0].Type())
			}

			key, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'setRespHeader' must be STRING, got %s", args[1].Type())
			}

			value, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'setRespHeader' must be STRING, got %s", args[2].Type())
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			resp.Value.Header().Set(key.Value, value.Value)
			return NULL
		},
	},

	// addRespHeader adds a response header (allows multiple values).
	// Usage: addRespHeader(response, key, value)
	"addRespHeader": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for addRespHeader. got=%d, want=3", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'addRespHeader' must be HTTP_RESP, got %s", args[0].Type())
			}

			key, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'addRespHeader' must be STRING, got %s", args[1].Type())
			}

			value, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'addRespHeader' must be STRING, got %s", args[2].Type())
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			resp.Value.Header().Add(key.Value, value.Value)
			return NULL
		},
	},

	// getReqHeader gets a request header value.
	// Usage: getReqHeader(request, key)
	"getReqHeader": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for getReqHeader. got=%d, want=2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'getReqHeader' must be HTTP_REQ, got %s", args[0].Type())
			}

			key, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'getReqHeader' must be STRING, got %s", args[1].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			value := req.Value.Header.Get(key.Value)
			return NewString(value)
		},
	},

	// getReqHeaders gets all request headers as a map.
	// Usage: getReqHeaders(request)
	"getReqHeaders": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getReqHeaders. got=%d, want=1", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("argument to 'getReqHeaders' must be HTTP_REQ, got %s", args[0].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			return req.getHeaders()
		},
	},

	// setCookie sets a cookie on the response.
	// Usage: setCookie(response, name, value, options?)
	// options is a map with optional keys: path, domain, maxAge, secure, httpOnly, sameSite
	"setCookie": {
		Fn: func(args ...Object) Object {
			if len(args) < 3 || len(args) > 4 {
				return newError("wrong number of arguments for setCookie. got=%d, want=3 or 4", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'setCookie' must be HTTP_RESP, got %s", args[0].Type())
			}

			name, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'setCookie' must be STRING, got %s", args[1].Type())
			}

			value, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'setCookie' must be STRING, got %s", args[2].Type())
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			cookie := &http.Cookie{
				Name:  name.Value,
				Value: value.Value,
			}

			// Parse options map
			if len(args) == 4 {
				opts, ok := args[3].(*Map)
				if !ok {
					return newError("fourth argument to 'setCookie' must be MAP, got %s", args[3].Type())
				}

				for _, pair := range opts.Pairs {
					keyStr, ok := pair.Key.(*String)
					if !ok {
						continue
					}

					switch keyStr.Value {
					case "path":
						if v, ok := pair.Value.(*String); ok {
							cookie.Path = v.Value
						}
					case "domain":
						if v, ok := pair.Value.(*String); ok {
							cookie.Domain = v.Value
						}
					case "maxAge":
						if v, ok := pair.Value.(*Int); ok {
							cookie.MaxAge = int(v.Value)
						}
					case "secure":
						if v, ok := pair.Value.(*Bool); ok {
							cookie.Secure = v.Value
						}
					case "httpOnly":
						if v, ok := pair.Value.(*Bool); ok {
							cookie.HttpOnly = v.Value
						}
					case "sameSite":
						if v, ok := pair.Value.(*Int); ok {
							cookie.SameSite = http.SameSite(v.Value)
						}
					case "expires":
						if v, ok := pair.Value.(*String); ok {
							t, err := time.Parse(time.RFC3339, v.Value)
							if err == nil {
								cookie.Expires = t
							}
						}
					}
				}
			}

			http.SetCookie(resp.Value, cookie)
			return NULL
		},
	},

	// getCookie gets a cookie value from the request.
	// Usage: getCookie(request, name)
	"getCookie": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for getCookie. got=%d, want=2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'getCookie' must be HTTP_REQ, got %s", args[0].Type())
			}

			name, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'getCookie' must be STRING, got %s", args[1].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			cookie, err := req.Value.Cookie(name.Value)
			if err != nil {
				return NULL
			}

			return NewString(cookie.Value)
		},
	},

	// getCookies gets all cookies from the request as a map.
	// Usage: getCookies(request)
	"getCookies": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for getCookies. got=%d, want=1", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("argument to 'getCookies' must be HTTP_REQ, got %s", args[0].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			pairs := make(map[HashKey]MapPair)
			for _, cookie := range req.Value.Cookies() {
				k := NewString(cookie.Name)
				pairs[k.HashKey()] = MapPair{Key: k, Value: NewString(cookie.Value)}
			}

			return NewMap(pairs)
		},
	},

	// parseForm parses form data from the request.
	// Returns a map of form values.
	// Usage: parseForm(request)
	"parseForm": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for parseForm. got=%d, want=1", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("argument to 'parseForm' must be HTTP_REQ, got %s", args[0].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			if err := req.Value.ParseForm(); err != nil {
				return newError("parse form failed: %v", err)
			}

			pairs := make(map[HashKey]MapPair)

			// Add URL query parameters
			for key, values := range req.Value.URL.Query() {
				k := NewString(key)
				var v Object
				if len(values) == 1 {
					v = NewString(values[0])
				} else {
					elements := make([]Object, len(values))
					for i, val := range values {
						elements[i] = NewString(val)
					}
					v = NewArray(elements)
				}
				pairs[k.HashKey()] = MapPair{Key: k, Value: v}
			}

			// Add form data (POST parameters)
			for key, values := range req.Value.PostForm {
				k := NewString(key)
				// If key already exists from URL query, combine values
				if existing, ok := pairs[k.HashKey()]; ok {
					if arr, ok := existing.Value.(*Array); ok {
						elements := arr.Elements
						for _, val := range values {
							elements = append(elements, NewString(val))
						}
						pairs[k.HashKey()] = MapPair{Key: k, Value: NewArray(elements)}
					} else {
						elements := []Object{existing.Value}
						for _, val := range values {
							elements = append(elements, NewString(val))
						}
						pairs[k.HashKey()] = MapPair{Key: k, Value: NewArray(elements)}
					}
				} else {
					var v Object
					if len(values) == 1 {
						v = NewString(values[0])
					} else {
						elements := make([]Object, len(values))
						for i, val := range values {
							elements[i] = NewString(val)
						}
						v = NewArray(elements)
					}
					pairs[k.HashKey()] = MapPair{Key: k, Value: v}
				}
			}

			return NewMap(pairs)
		},
	},

	// status sets the HTTP status code for the response.
	// Usage: status(response, code)
	"status": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for status. got=%d, want=2", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'status' must be HTTP_RESP, got %s", args[0].Type())
			}

			code, ok := args[1].(*Int)
			if !ok {
				return newError("second argument to 'status' must be INT, got %s", args[1].Type())
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			resp.Value.WriteHeader(int(code.Value))
			return NULL
		},
	},

	// redirect performs an HTTP redirect.
	// Usage: redirect(response, url, code?)
	// Default code is 302 (Found)
	"redirect": {
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments for redirect. got=%d, want=2 or 3", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'redirect' must be HTTP_RESP, got %s", args[0].Type())
			}

			url, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'redirect' must be STRING, got %s", args[1].Type())
			}

			code := 302 // Default: Found
			if len(args) == 3 {
				c, ok := args[2].(*Int)
				if !ok {
					return newError("third argument to 'redirect' must be INT, got %s", args[2].Type())
				}
				code = int(c.Value)
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			http.Redirect(resp.Value, &http.Request{}, url.Value, code)
			resp.SetWritten()
			return NULL
		},
	},

	// serveFile serves a static file.
	// Usage: serveFile(response, request, path)
	"serveFile": {
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments for serveFile. got=%d, want=3", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'serveFile' must be HTTP_RESP, got %s", args[0].Type())
			}

			req, ok := args[1].(*HttpReq)
			if !ok {
				return newError("second argument to 'serveFile' must be HTTP_REQ, got %s", args[1].Type())
			}

			path, ok := args[2].(*String)
			if !ok {
				return newError("third argument to 'serveFile' must be STRING, got %s", args[2].Type())
			}

			if resp.Value == nil || req.Value == nil {
				return newError("http response or request is nil")
			}

			http.ServeFile(resp.Value, req.Value, path.Value)
			resp.SetWritten()
			return NULL
		},
	},

	// setContentType sets the Content-Type header.
	// Usage: setContentType(response, mimeType)
	"setContentType": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for setContentType. got=%d, want=2", len(args))
			}

			resp, ok := args[0].(*HttpResp)
			if !ok {
				return newError("first argument to 'setContentType' must be HTTP_RESP, got %s", args[0].Type())
			}

			mimeType, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'setContentType' must be STRING, got %s", args[1].Type())
			}

			if resp.Value == nil {
				return newError("http response is nil")
			}

			resp.Value.Header().Set("Content-Type", mimeType.Value)
			return NULL
		},
	},

	// queryParam gets a URL query parameter by name.
	// Usage: queryParam(request, name)
	"queryParam": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for queryParam. got=%d, want=2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'queryParam' must be HTTP_REQ, got %s", args[0].Type())
			}

			name, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'queryParam' must be STRING, got %s", args[1].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			value := req.Value.URL.Query().Get(name.Value)
			return NewString(value)
		},
	},

	// queryParams gets all URL query parameters as a map.
	// Usage: queryParams(request)
	"queryParams": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for queryParams. got=%d, want=1", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("argument to 'queryParams' must be HTTP_REQ, got %s", args[0].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			pairs := make(map[HashKey]MapPair)
			for key, values := range req.Value.URL.Query() {
				k := NewString(key)
				var v Object
				if len(values) == 1 {
					v = NewString(values[0])
				} else {
					elements := make([]Object, len(values))
					for i, val := range values {
						elements[i] = NewString(val)
					}
					v = NewArray(elements)
				}
				pairs[k.HashKey()] = MapPair{Key: k, Value: v}
			}

			return NewMap(pairs)
		},
	},

	// formValue gets a form value by name.
	// Usage: formValue(request, name)
	"formValue": {
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments for formValue. got=%d, want=2", len(args))
			}

			req, ok := args[0].(*HttpReq)
			if !ok {
				return newError("first argument to 'formValue' must be HTTP_REQ, got %s", args[0].Type())
			}

			name, ok := args[1].(*String)
			if !ok {
				return newError("second argument to 'formValue' must be STRING, got %s", args[1].Type())
			}

			if req.Value == nil {
				return newError("http request is nil")
			}

			value := req.Value.FormValue(name.Value)
			return NewString(value)
		},
	},
}

// RegisterHttpBuiltins registers HTTP built-in functions into the main Builtins map.
func RegisterHttpBuiltins() {
	for name, builtin := range HttpBuiltins {
		Builtins[name] = builtin
	}
}

// httpStatusNames maps status codes to their names.
var httpStatusNames = map[int]string{
	200: "OK",
	201: "Created",
	202: "Accepted",
	204: "No Content",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	415: "Unsupported Media Type",
	422: "Unprocessable Entity",
	429: "Too Many Requests",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

// HttpStatusName returns the name for an HTTP status code.
// Usage: httpStatusName(code)
func init() {
	Builtins["httpStatusName"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for httpStatusName. got=%d, want=1", len(args))
			}

			code, ok := args[0].(*Int)
			if !ok {
				return newError("argument to 'httpStatusName' must be INT, got %s", args[0].Type())
			}

			name, ok := httpStatusNames[int(code.Value)]
			if !ok {
				return NewString(strconv.Itoa(int(code.Value)))
			}
			return NewString(name)
		},
	}

	// Additional helper functions

	// isHttpReq - check if value is an HTTP request
	Builtins["isHttpReq"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isHttpReq. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*HttpReq)
			return &Bool{Value: ok}
		},
	}

	// isHttpResp - check if value is an HTTP response
	Builtins["isHttpResp"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for isHttpResp. got=%d, want=1", len(args))
			}
			_, ok := args[0].(*HttpResp)
			return &Bool{Value: ok}
		},
	}

}
