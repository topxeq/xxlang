// pkg/stdlib/bytes.go
// Byte manipulation utilities for the Xxlang standard library.
package stdlib

import (
	"bytes"
	"encoding/binary"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/bytes",
		Exports: map[string]objects.Object{
			// Create byte array from string
			"fromString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromString() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("fromString() requires a string argument")
				}
				result := make([]objects.Object, len(s.Value))
				for i, c := range s.Value {
					result[i] = Int(int64(c))
				}
				return Array(result...)
			}),

			// Convert byte array to string
			"toString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toString() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("toString() requires an array argument")
				}
				result := make([]byte, len(arr.Elements))
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("toString() requires integer array elements")
					}
					if n.Value < 0 || n.Value > 255 {
						return Error("byte values must be 0-255")
					}
					result[i] = byte(n.Value)
				}
				return String(string(result))
			}),

			// Get byte at index
			"get": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("get() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("get() requires an array as first argument")
				}
				idx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("get() requires an integer index")
				}
				if idx.Value < 0 || int(idx.Value) >= len(arr.Elements) {
					return Error("index out of range")
				}
				return arr.Elements[idx.Value]
			}),

			// Set byte at index
			"set": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("set() takes exactly 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("set() requires an array as first argument")
				}
				idx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("set() requires an integer index")
				}
				if idx.Value < 0 || int(idx.Value) >= len(arr.Elements) {
					return Error("index out of range")
				}
				val, ok := args[2].(*objects.Int)
				if !ok {
					return Error("set() requires an integer value")
				}
				arr.Elements[idx.Value] = val
				return arr
			}),

			// Encode int to bytes (big endian)
			"encodeInt64BE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("encodeInt64BE() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("encodeInt64BE() requires an integer argument")
				}
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uint64(n.Value))
				result := make([]objects.Object, 8)
				for i, b := range buf {
					result[i] = Int(int64(b))
				}
				return Array(result...)
			}),

			// Decode int from bytes (big endian)
			"decodeInt64BE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("decodeInt64BE() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok || len(arr.Elements) != 8 {
					return Error("decodeInt64BE() requires an 8-element array")
				}
				buf := make([]byte, 8)
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("decodeInt64BE() requires integer array elements")
					}
					buf[i] = byte(n.Value)
				}
				return Int(int64(binary.BigEndian.Uint64(buf)))
			}),

			// Encode int to bytes (little endian)
			"encodeInt64LE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("encodeInt64LE() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("encodeInt64LE() requires an integer argument")
				}
				buf := make([]byte, 8)
				binary.LittleEndian.PutUint64(buf, uint64(n.Value))
				result := make([]objects.Object, 8)
				for i, b := range buf {
					result[i] = Int(int64(b))
				}
				return Array(result...)
			}),

			// Decode int from bytes (little endian)
			"decodeInt64LE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("decodeInt64LE() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok || len(arr.Elements) != 8 {
					return Error("decodeInt64LE() requires an 8-element array")
				}
				buf := make([]byte, 8)
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("decodeInt64LE() requires integer array elements")
					}
					buf[i] = byte(n.Value)
				}
				return Int(int64(binary.LittleEndian.Uint64(buf)))
			}),

			// Concat byte arrays
			"concat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("concat() takes at least 2 arguments")
				}
				result := []objects.Object{}
				for _, arg := range args {
					arr, ok := arg.(*objects.Array)
					if !ok {
						return Error("concat() requires array arguments")
					}
					result = append(result, arr.Elements...)
				}
				return Array(result...)
			}),

			// Slice byte array
			"slice": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("slice() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("slice() requires an array as first argument")
				}
				start, ok := args[1].(*objects.Int)
				if !ok {
					return Error("slice() requires integer indices")
				}
				end := int64(len(arr.Elements))
				if len(args) > 2 {
					e, ok := args[2].(*objects.Int)
					if ok {
						end = e.Value
					}
				}
				s := int(start.Value)
				e := int(end)
				if s < 0 {
					s = 0
				}
				if e > len(arr.Elements) {
					e = len(arr.Elements)
				}
				if s > e {
					return Array()
				}
				return Array(arr.Elements[s:e]...)
			}),

			// Compare byte arrays
			"compare": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("compare() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("compare() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("compare() requires array arguments")
				}
				// Convert to byte slices
				b1 := make([]byte, len(arr1.Elements))
				for i, elem := range arr1.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("compare() requires integer array elements")
					}
					b1[i] = byte(n.Value)
				}
				b2 := make([]byte, len(arr2.Elements))
				for i, elem := range arr2.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("compare() requires integer array elements")
					}
					b2[i] = byte(n.Value)
				}
				return Int(int64(bytes.Compare(b1, b2)))
			}),

			// Check if byte arrays are equal
			"equal": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("equal() takes exactly 2 arguments")
				}
				arr1, ok := args[0].(*objects.Array)
				if !ok {
					return Error("equal() requires array arguments")
				}
				arr2, ok := args[1].(*objects.Array)
				if !ok {
					return Error("equal() requires array arguments")
				}
				if len(arr1.Elements) != len(arr2.Elements) {
					return Bool(false)
				}
				for i := range arr1.Elements {
					if arr1.Elements[i].Inspect() != arr2.Elements[i].Inspect() {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Count bytes
			"count": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("count() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("count() requires an array as first argument")
				}
				target, ok := args[1].(*objects.Int)
				if !ok {
					return Error("count() requires an integer as second argument")
				}
				count := 0
				for _, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if ok && n.Value == target.Value {
						count++
					}
				}
				return Int(int64(count))
			}),

			// Find byte in array
			"indexOf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("indexOf() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("indexOf() requires an array as first argument")
				}
				target, ok := args[1].(*objects.Int)
				if !ok {
					return Error("indexOf() requires an integer as second argument")
				}
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if ok && n.Value == target.Value {
						return Int(int64(i))
					}
				}
				return Int(-1)
			}),
		},
	})
}
