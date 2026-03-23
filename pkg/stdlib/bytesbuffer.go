// pkg/stdlib/bytesbuffer.go
// BytesBuffer module for efficient byte buffer operations.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "bytesbuffer",
		Exports: map[string]objects.Object{
			// create creates a new BytesBuffer instance
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return Error("create takes at most 1 argument")
				}

				bb := objects.NewBytesBuffer()

				// Optional initial capacity
				if len(args) == 1 {
					cap, ok := args[0].(*objects.Int)
					if !ok {
						return Error("capacity must be an integer")
					}
					if cap.Value > 0 {
						bb.Grow(int(cap.Value))
					}
				}

				return bb
			}),

			// fromBytes creates a BytesBuffer from an array of bytes (ints 0-255)
			"fromBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromBytes takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("argument must be an array")
				}
				data := make([]byte, len(arr.Elements))
				for i, elem := range arr.Elements {
					b, ok := elem.(*objects.Int)
					if !ok {
						return Error("array elements must be integers")
					}
					if b.Value < 0 || b.Value > 255 {
						return Error("array element out of byte range")
					}
					data[i] = byte(b.Value)
				}
				return objects.NewBytesBufferFromBytes(data)
			}),

			// fromString creates a BytesBuffer from a string
			"fromString": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromString takes exactly 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				return objects.NewBytesBufferFromBytes([]byte(str.Value))
			}),

			// isBytesBuffer checks if an object is a BytesBuffer
			"isBytesBuffer": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isBytesBuffer takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.BytesBuffer)
				return Bool(ok)
			}),
		},
	})
}
