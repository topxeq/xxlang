// pkg/stdlib/uuid.go
// UUID generation utilities for the Xxlang standard library.
package stdlib

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/uuid",
		Exports: map[string]objects.Object{
			// Generate v4 UUID
			"v4": BuiltinFunc(func(args ...objects.Object) objects.Object {
				uuid := make([]byte, 16)
				_, err := rand.Read(uuid)
				if err != nil {
					return Error(err.Error())
				}
				// Set version (4) and variant bits
				uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
				uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122
				return String(fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]))
			}),

			// Generate v4 UUID without dashes
			"v4Short": BuiltinFunc(func(args ...objects.Object) objects.Object {
				uuid := make([]byte, 16)
				_, err := rand.Read(uuid)
				if err != nil {
					return Error(err.Error())
				}
				uuid[6] = (uuid[6] & 0x0f) | 0x40
				uuid[8] = (uuid[8] & 0x3f) | 0x80
				return String(hex.EncodeToString(uuid))
			}),

			// Generate a simple ID (not UUID compliant but fast)
			"simple": BuiltinFunc(func(args ...objects.Object) objects.Object {
				b := make([]byte, 8)
				_, err := rand.Read(b)
				if err != nil {
					return Error(err.Error())
				}
				return String(hex.EncodeToString(b))
			}),

			// Generate time-based ID (sortable)
			"timeID": BuiltinFunc(func(args ...objects.Object) objects.Object {
				now := time.Now().UnixNano()
				random := make([]byte, 4)
				_, err := rand.Read(random)
				if err != nil {
					return Error(err.Error())
				}
				return String(fmt.Sprintf("%016x%s", now, hex.EncodeToString(random)))
			}),

			// Generate a short random string
			"random": BuiltinFunc(func(args ...objects.Object) objects.Object {
				length := 16
				if len(args) > 0 {
					n, ok := args[0].(*objects.Int)
					if ok && n.Value > 0 && n.Value <= 64 {
						length = int(n.Value)
					}
				}
				const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
				b := make([]byte, length)
				r := make([]byte, length)
				_, err := rand.Read(r)
				if err != nil {
					return Error(err.Error())
				}
				for i := range b {
					b[i] = chars[int(r[i])%len(chars)]
				}
				return String(string(b))
			}),

			// Generate hex string
			"hex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				length := 16
				if len(args) > 0 {
					n, ok := args[0].(*objects.Int)
					if ok && n.Value > 0 && n.Value <= 64 {
						length = int(n.Value)
					}
				}
				b := make([]byte, length)
				_, err := rand.Read(b)
				if err != nil {
					return Error(err.Error())
				}
				return String(hex.EncodeToString(b))
			}),

			// Validate UUID format
			"isValid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isValid() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isValid() requires a string argument")
				}
				uuid := s.Value
				if len(uuid) != 36 {
					return Bool(false)
				}
				// Check dashes at correct positions
				if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
					return Bool(false)
				}
				// Check hex characters
				for i, c := range uuid {
					if i == 8 || i == 13 || i == 18 || i == 23 {
						continue
					}
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Parse UUID to bytes
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parse() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}
				// Remove dashes
				hexStr := ""
				for _, c := range s.Value {
					if c != '-' {
						hexStr += string(c)
					}
				}
				if len(hexStr) != 32 {
					return Error("invalid UUID format")
				}
				bytes, err := hex.DecodeString(hexStr)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(bytes))
			}),

			// Format bytes as UUID
			"format": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("format() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("format() requires a string argument")
				}
				hexStr := ""
				for _, c := range s.Value {
					if c != '-' {
						hexStr += string(c)
					}
				}
				if len(hexStr) != 32 {
					return Error("invalid UUID hex length")
				}
				return String(fmt.Sprintf("%s-%s-%s-%s-%s",
					hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32]))
			}),
		},
	})
}
