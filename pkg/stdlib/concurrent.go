// pkg/stdlib/concurrent.go
// Concurrency built-in functions for Xxlang
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "concurrent",
		Exports: map[string]objects.Object{
			// ============================================
			// Tube operations
			// ============================================

			// makeTube(type?, buffer?) - Create a new tube
			"makeTube": BuiltinFunc(func(args ...objects.Object) objects.Object {
				elemType := objects.ObjectType("")
				buffer := 0

				argIdx := 0
				if len(args) > 0 {
					// Check if first argument is type string or number
					if str, ok := args[0].(*objects.String); ok {
						elemType = objects.ObjectType(str.Value)
						argIdx = 1
					} else if num, ok := args[0].(*objects.Int); ok {
						buffer = int(num.Value)
						argIdx = 1
					}
				}

				// If there's a second argument, use as buffer
				if len(args) > argIdx {
					if num, ok := args[argIdx].(*objects.Int); ok {
						buffer = int(num.Value)
					}
				}

				return objects.NewTube(elemType, buffer)
			}),

			// closeTube(tube) - Close a tube
			"closeTube": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("closeTube requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				t.Close()
				return objects.NULL
			}),

			// tubeLen(tube) - Get number of elements in tube buffer
			"tubeLen": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("tubeLen requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				return Int(int64(t.Len()))
			}),

			// tubeCap(tube) - Get tube buffer capacity
			"tubeCap": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("tubeCap requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				return Int(int64(t.Cap()))
			}),

			// tubeClosed(tube) - Check if tube is closed
			"tubeClosed": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("tubeClosed requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				return Bool(t.IsClosed())
			}),

			// tubeSend(tube, value) - Send value to tube (blocking)
			"tubeSend": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("tubeSend requires 2 arguments (tube, value)")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("first argument must be a tube")
				}
				if !t.Send(args[1]) {
					return Bool(false)
				}
				return Bool(true)
			}),

			// tubeRecv(tube) - Receive value from tube (blocking), returns [value, ok]
			"tubeRecv": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("tubeRecv requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				val, ok := t.Receive()
				return Array(val, Bool(ok))
			}),

			// tubeTrySend(tube, value) - Try to send without blocking, returns [sent, ok]
			"tubeTrySend": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("tubeTrySend requires 2 arguments (tube, value)")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("first argument must be a tube")
				}
				sent, ok := t.TrySend(args[1])
				return Array(Bool(sent), Bool(ok))
			}),

			// tubeTryRecv(tube) - Try to receive without blocking, returns [value, received, open]
			"tubeTryRecv": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("tubeTryRecv requires 1 argument")
				}
				t, ok := args[0].(*objects.Tube)
				if !ok {
					return Error("argument must be a tube")
				}
				val, received, open := t.TryReceive()
				return Array(val, Bool(received), Bool(open))
			}),

			// ============================================
			// Synchronization primitives
			// ============================================

			// newMutex() - Create a new mutex
			"newMutex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewMutex()
			}),

			// newRWMutex() - Create a new read-write mutex
			"newRWMutex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewRWMutex()
			}),

			// newWaitGroup() - Create a new wait group
			"newWaitGroup": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewWaitGroup()
			}),

			// newOnce() - Create a new once
			"newOnce": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewOnce()
			}),

			// newCond(mutex) - Create a new condition variable
			"newCond": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("newCond requires 1 argument (mutex)")
				}
				m, ok := args[0].(*objects.Mutex)
				if !ok {
					return Error("argument must be a Mutex")
				}
				return objects.NewCond(m)
			}),

			// newAtomic(value?) - Create a new atomic integer
			"newAtomic": BuiltinFunc(func(args ...objects.Object) objects.Object {
				initial := int64(0)
				if len(args) > 0 {
					if num, ok := args[0].(*objects.Int); ok {
						initial = num.Value
					}
				}
				return objects.NewAtomicInt(initial)
			}),
		},
	})
}
