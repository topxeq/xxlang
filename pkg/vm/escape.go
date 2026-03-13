// pkg/vm/escape.go
// Go escape analysis hints for performance optimization
// This file contains helper functions and directives to guide the compiler
// on memory allocation decisions

package vm

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

//go:noinline
func escapeToHeap(obj objects.Object) objects.Object {
	// This function forces escape analysis to understand
	// that the object must live on the heap
	return obj
}

// stackBool returns a boolean without forcing heap allocation
// Used for hot path boolean operations
func stackBool(value bool) *objects.Bool {
	if value {
		return objects.TRUE
	}
	return objects.FALSE
}

// nullIfZero returns NULL if the value is nil, otherwise returns the value
// This helps avoid unnecessary allocations for null checks
func nullIfZero(obj objects.Object) objects.Object {
	if obj == nil {
		return objects.NULL
	}
	return obj
}

// intResult creates an Int result, using cache for small values
// This is inlined for hot path arithmetic operations
//go:noinline
func intResult(val int64) objects.Object {
	return objects.NewInt(val)
}

// floatResult creates a Float result
//go:noinline
func floatResult(val float64) objects.Object {
	return &objects.Float{Value: val}
}
