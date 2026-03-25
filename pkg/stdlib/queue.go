// pkg/stdlib/queue.go
// Queue module for Xxlang - FIFO data structure.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "queue",
		Exports: map[string]objects.Object{
			// create creates a new empty Queue
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return Error("create takes at most 1 argument")
				}

				if len(args) == 0 {
					return objects.NewQueue()
				}

				// Optional initial capacity
				cap, ok := args[0].(*objects.Int)
				if !ok {
					return Error("capacity must be an integer")
				}
				if cap.Value < 4 {
					return objects.NewQueue()
				}
				return objects.NewQueueWithCapacity(int(cap.Value))
			}),

			// fromArray creates a Queue from an array
			"fromArray": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromArray takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("argument must be an array")
				}
				return objects.NewQueueFrom(arr)
			}),

			// isQueue checks if an object is a Queue
			"isQueue": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isQueue takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.Queue)
				return Bool(ok)
			}),
		},
	})
}