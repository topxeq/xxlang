// pkg/stdlib/encoding.go
// Encoding utilities for the Xxlang standard library.
package stdlib

import (
	"encoding/base64"
	"encoding/hex"
	"net/url"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/encoding",
		Exports: map[string]objects.Object{
			// Base64 encoding
			"base64Encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64Encode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64Encode() requires a string argument")
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"base64Decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64Decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64Decode() requires a string argument")
				}
				decoded, err := base64.StdEncoding.DecodeString(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(decoded))
			}),

			"base64URLEncode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64URLEncode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64URLEncode() requires a string argument")
				}
				encoded := base64.URLEncoding.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"base64URLDecode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64URLDecode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64URLDecode() requires a string argument")
				}
				decoded, err := base64.URLEncoding.DecodeString(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(decoded))
			}),

			// Hex encoding
			"hexEncode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("hexEncode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hexEncode() requires a string argument")
				}
				encoded := hex.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"hexDecode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("hexDecode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hexDecode() requires a string argument")
				}
				decoded, err := hex.DecodeString(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(decoded))
			}),

			// URL encoding
			"urlEncode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("urlEncode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("urlEncode() requires a string argument")
				}
				encoded := url.QueryEscape(s.Value)
				return String(encoded)
			}),

			"urlDecode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("urlDecode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("urlDecode() requires a string argument")
				}
				decoded, err := url.QueryUnescape(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(decoded)
			}),

			// URL parsing
			"parseURL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseURL() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseURL() requires a string argument")
				}
				u, err := url.Parse(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Array(
					String(u.Scheme),
					String(u.Host),
					String(u.Path),
					String(u.RawQuery),
					String(u.Fragment),
				)
			}),

			"buildURL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("buildURL() takes at least 2 arguments")
				}
				scheme, ok := args[0].(*objects.String)
				if !ok {
					return Error("buildURL() requires string arguments")
				}
				host, ok := args[1].(*objects.String)
				if !ok {
					return Error("buildURL() requires string arguments")
				}
				u := &url.URL{
					Scheme: scheme.Value,
					Host:   host.Value,
				}
				if len(args) > 2 {
					path, ok := args[2].(*objects.String)
					if ok {
						u.Path = path.Value
					}
				}
				return String(u.String())
			}),

			"setQuery": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("setQuery() takes exactly 3 arguments")
				}
				baseURL, ok := args[0].(*objects.String)
				if !ok {
					return Error("setQuery() requires string arguments")
				}
				key, ok := args[1].(*objects.String)
				if !ok {
					return Error("setQuery() requires string arguments")
				}
				value, ok := args[2].(*objects.String)
				if !ok {
					return Error("setQuery() requires string arguments")
				}
				u, err := url.Parse(baseURL.Value)
				if err != nil {
					return Error(err.Error())
				}
				q := u.Query()
				q.Set(key.Value, value.Value)
				u.RawQuery = q.Encode()
				return String(u.String())
			}),

			// Escape sequences
			"escapeHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("escapeHTML() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("escapeHTML() requires a string argument")
				}
				escaped := url.PathEscape(s.Value)
				return String(escaped)
			}),

			"unescapeHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("unescapeHTML() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("unescapeHTML() requires a string argument")
				}
				unescaped, err := url.PathUnescape(s.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(unescaped)
			}),
		},
	})
}
