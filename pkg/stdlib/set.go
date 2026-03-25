// pkg/stdlib/set.go
// Set module for Xxlang - unordered collection of unique elements.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "set",
		Exports: map[string]objects.Object{
			// create creates a new empty Set with optional initial elements
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return objects.NewSet()
				}

				// If first arg is an int, use as capacity
				if cap, ok := args[0].(*objects.Int); ok && len(args) == 1 {
					if cap.Value < 4 {
						return objects.NewSet()
					}
					return objects.NewSetWithCapacity(int(cap.Value))
				}

				// Otherwise, add all arguments as elements
				s := objects.NewSetWithCapacity(len(args))
				for _, arg := range args {
					s.Add(arg)
				}
				return s
			}),

			// fromArray creates a Set from an array
			"fromArray": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromArray takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("argument must be an array")
				}
				return objects.NewSetFrom(arr)
			}),

			// isSet checks if an object is a Set
			"isSet": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSet takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.Set)
				return Bool(ok)
			}),

			// union returns a new set containing all elements from both sets
			"union": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("union takes exactly 2 arguments")
				}
				s1, ok := args[0].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				s2, ok := args[1].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				return s1.Union(s2)
			}),

			// intersect returns a new set containing elements present in both sets
			"intersect": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("intersect takes exactly 2 arguments")
				}
				s1, ok := args[0].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				s2, ok := args[1].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				return s1.Intersect(s2)
			}),

			// difference returns a new set containing elements in s1 but not in s2
			"difference": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("difference takes exactly 2 arguments")
				}
				s1, ok := args[0].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				s2, ok := args[1].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				return s1.Difference(s2)
			}),

			// symmetricDiff returns a new set containing elements in either set but not both
			"symmetricDiff": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("symmetricDiff takes exactly 2 arguments")
				}
				s1, ok := args[0].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				s2, ok := args[1].(*objects.Set)
				if !ok {
					return Error("arguments must be Sets")
				}
				return s1.SymmetricDifference(s2)
			}),
		},
	})
}