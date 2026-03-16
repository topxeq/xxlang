// pkg/stdlib/crypto.go
// Cryptography utilities for the Xxlang standard library.
// Pure Go implementation using standard library - no CGO required.
package stdlib

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "crypto",
		Exports: map[string]objects.Object{
			// Hash functions
			"md5": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("md5() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("md5() requires a string argument")
				}
				hash := md5.Sum([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha1": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha1() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha1() requires a string argument")
				}
				hash := sha1.Sum([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha256": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha256() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha256() requires a string argument")
				}
				hash := sha256.Sum256([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha512": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha512() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha512() requires a string argument")
				}
				hash := sha512.Sum512([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			// HMAC functions
			"hmacMd5": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, md5.New, "hmacMd5")
			}),

			"hmacSha1": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha1.New, "hmacSha1")
			}),

			"hmacSha256": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha256.New, "hmacSha256")
			}),

			"hmacSha512": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha512.New, "hmacSha512")
			}),

			// Encoding functions
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
					return Error(fmt.Sprintf("base64Decode() failed: %s", err.Error()))
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
					return Error(fmt.Sprintf("base64URLDecode() failed: %s", err.Error()))
				}
				return String(string(decoded))
			}),

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
					return Error(fmt.Sprintf("hexDecode() failed: %s", err.Error()))
				}
				return String(string(decoded))
			}),

			// Random functions
			"randomBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomBytes() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomBytes() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomBytes() requires a non-negative integer")
				}
				if n.Value > 1024*1024 { // 1MB limit
					return Error("randomBytes() size too large (max 1MB)")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomBytes() failed: %s", err.Error()))
				}

				// Return as hex-encoded string for convenience
				return String(hex.EncodeToString(bytes))
			}),

			"randomHex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomHex() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomHex() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomHex() requires a non-negative integer")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomHex() failed: %s", err.Error()))
				}

				return String(hex.EncodeToString(bytes))
			}),

			"randomBase64": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomBase64() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomBase64() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomBase64() requires a non-negative integer")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomBase64() failed: %s", err.Error()))
				}

				return String(base64.StdEncoding.EncodeToString(bytes))
			}),

			// UUID generation
			"uuid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				uuid := make([]byte, 16)
				_, err := rand.Read(uuid)
				if err != nil {
					return Error(fmt.Sprintf("uuid() failed: %s", err.Error()))
				}

				// Set version (4) and variant bits
				uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
				uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122

				return String(fmt.Sprintf("%x-%x-%x-%x-%x",
					uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]))
			}),
		},
	})
}

// hmacHash is a helper function for HMAC operations
func hmacHash(args []objects.Object, hashFunc func() hash.Hash, name string) objects.Object {
	if len(args) != 2 {
		return Error(fmt.Sprintf("%s() takes exactly 2 arguments", name))
	}

	key, ok := args[0].(*objects.String)
	if !ok {
		return Error(fmt.Sprintf("%s() requires a string key", name))
	}

	data, ok := args[1].(*objects.String)
	if !ok {
		return Error(fmt.Sprintf("%s() requires a string data", name))
	}

	h := hmac.New(hashFunc, []byte(key.Value))
	h.Write([]byte(data.Value))
	return String(hex.EncodeToString(h.Sum(nil)))
}
