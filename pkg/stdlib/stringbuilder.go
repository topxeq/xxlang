// pkg/stdlib/stringbuilder.go
// StringBuilder module for efficient string concatenation.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "stringbuilder",
		Exports: map[string]objects.Object{
			// create creates a new StringBuilder instance
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return Error("create takes at most 1 argument")
				}

				sb := objects.NewStringBuilder()

				// Optional initial capacity
				if len(args) == 1 {
					cap, ok := args[0].(*objects.Int)
					if !ok {
						return Error("capacity must be an integer")
					}
					if cap.Value > 0 {
						sb.Grow(int(cap.Value))
					}
				}

				return sb
			}),

			// isStringBuilder checks if an object is a StringBuilder
			"isStringBuilder": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isStringBuilder takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.StringBuilder)
				return Bool(ok)
			}),
		},
	})
}
