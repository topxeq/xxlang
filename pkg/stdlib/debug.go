// pkg/stdlib/debug.go
// Debug utilities for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "debug",
		Exports: map[string]objects.Object{
			// Type information
			"type": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("type() takes exactly 1 argument")
				}
				return String(string(args[0].Type()))
			}),

			"typeTag": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("typeTag() takes exactly 1 argument")
				}
				return Int(int64(args[0].TypeTag()))
			}),

			// Detailed inspect
			"inspect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("inspect() takes exactly 1 argument")
				}
				detail := fmt.Sprintf("Type: %s\nValue: %s", args[0].Type(), args[0].Inspect())
				return String(detail)
			}),

			// Memory info
			"memStats": BuiltinFunc(func(args ...objects.Object) objects.Object {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return Array(
					Int(int64(m.Alloc)),      // bytes allocated
					Int(int64(m.TotalAlloc)), // total bytes allocated
					Int(int64(m.Sys)),        // bytes from OS
					Int(int64(m.NumGC)),      // number of GCs
				)
			}),

			// GC
			"gc": BuiltinFunc(func(args ...objects.Object) objects.Object {
				runtime.GC()
				return Null()
			}),

			// Goroutines
			"goroutines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(runtime.NumGoroutine()))
			}),

			// Stack trace
			"stack": BuiltinFunc(func(args ...objects.Object) objects.Object {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				return String(string(buf[:n]))
			}),

			// Callers (simplified)
			"callers": BuiltinFunc(func(args ...objects.Object) objects.Object {
				pcs := make([]uintptr, 32)
				n := runtime.Callers(2, pcs)
				result := []objects.Object{}
				for i := 0; i < n; i++ {
					fn := runtime.FuncForPC(pcs[i])
					if fn != nil {
						result = append(result, String(fn.Name()))
					}
				}
				return Array(result...)
			}),

			// Assert
			"assert": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("assert() takes at least 1 argument")
				}
				cond, ok := args[0].(*objects.Bool)
				if !ok {
					return Error("assert() requires a boolean condition")
				}
				if !cond.Value {
					msg := "assertion failed"
					if len(args) > 1 {
						s, ok := args[1].(*objects.String)
						if ok {
							msg = s.Value
						}
					}
					return Error(msg)
				}
				return Null()
			}),

			// Assert equals
			"assertEquals": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("assertEquals() takes at least 2 arguments")
				}
				a := args[0].Inspect()
				b := args[1].Inspect()
				if a != b {
					msg := fmt.Sprintf("expected %s, got %s", b, a)
					if len(args) > 2 {
						s, ok := args[2].(*objects.String)
						if ok {
							msg = s.Value + ": " + msg
						}
					}
					return Error(msg)
				}
				return Null()
			}),

			// Dump value info
			"dump": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("dump() takes exactly 1 argument")
				}
				obj := args[0]
				info := make([]objects.Object, 0)

				info = append(info, Array(String("type"), String(string(obj.Type()))))
				info = append(info, Array(String("inspect"), String(obj.Inspect())))

				switch v := obj.(type) {
				case *objects.Int:
					info = append(info, Array(String("value"), String(fmt.Sprintf("%d", v.Value))))
				case *objects.Float:
					info = append(info, Array(String("value"), String(fmt.Sprintf("%f", v.Value))))
				case *objects.String:
					info = append(info, Array(String("length"), Int(int64(len(v.Value)))))
					info = append(info, Array(String("value"), String(v.Value)))
				case *objects.Bool:
					info = append(info, Array(String("value"), String(fmt.Sprintf("%v", v.Value))))
				case *objects.Array:
					info = append(info, Array(String("length"), Int(int64(len(v.Elements)))))
				case *objects.Map:
					info = append(info, Array(String("size"), Int(int64(len(v.Pairs)))))
				}

				return Array(info...)
			}),

			// Is nil/null
			"isNull": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isNull() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.Null)
				return Bool(ok)
			}),

			// Is truthy
			"isTruthy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isTruthy() takes exactly 1 argument")
				}
				return Bool(args[0].ToBool().Value)
			}),

			// Reflect (Go reflect info)
			"reflect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("reflect() takes exactly 1 argument")
				}
				v := reflect.ValueOf(args[0])
				return Array(
					String(v.Kind().String()),
					String(v.Type().String()),
				)
			}),

			// Benchmark helper - returns current time in nanoseconds
			"nanoTime": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(runtimeNano()))
			}),

			// Elapsed time helper
			"elapsed": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("elapsed() takes exactly 1 argument")
				}
				start, ok := args[0].(*objects.Int)
				if !ok {
					return Error("elapsed() requires an integer start time")
				}
				elapsed := runtimeNano() - start.Value
				return Int(elapsed)
			}),
		},
	})
}

// runtimeNano returns current nanoseconds
func runtimeNano() int64 {
	return time.Now().UnixNano()
}
