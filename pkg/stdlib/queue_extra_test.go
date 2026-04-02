// pkg/stdlib/queue_extra_test.go
// Additional tests for queue module to increase coverage.
package stdlib

import (
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// callQueueFunc calls a function from the queue module.
func callQueueFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("queue")
	if mod == nil {
		t := &testing.T{}
		t.Skip("queue module not found")
		return &objects.Error{Message: "queue module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestQueueCreate_ArgumentValidation tests create argument validation.
func TestQueueCreate_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, false},                 // no args ok
		{[]objects.Object{Int(10)}, false},          // valid capacity
		{[]objects.Object{Int(2)}, false},           // capacity < 4 treated as 0? Actually creates empty queue.
		{[]objects.Object{String("not int")}, true}, // non-int arg
		{[]objects.Object{Int(10), Int(20)}, true},  // too many args
	}
	for _, tt := range tests {
		result := callQueueFunc("create", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("create(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("create(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestQueueFromArray_ArgumentValidation tests fromArray argument validation.
func TestQueueFromArray_ArgumentValidation(t *testing.T) {
	tests := []struct {
		args    []objects.Object
		wantErr bool
	}{
		{[]objects.Object{}, true},
		{[]objects.Object{String("not array")}, true},
		{[]objects.Object{Array(Int(1), Int(2))}, false},
	}
	for _, tt := range tests {
		result := callQueueFunc("fromArray", tt.args...)
		if tt.wantErr {
			if _, ok := result.(*objects.Error); !ok {
				t.Errorf("fromArray(%v) expected error, got %T", tt.args, result)
			}
		} else {
			if _, ok := result.(*objects.Error); ok {
				msg := result.Inspect()
				if strings.Contains(msg, "must be a") || strings.Contains(msg, "takes exactly") {
					t.Errorf("fromArray(%v) got argument validation error: %s", tt.args, msg)
				}
			}
		}
	}
}

// TestQueueIsQueue_TypeCheck tests isQueue with various types.
func TestQueueIsQueue_TypeCheck(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	fn, ok := mod.Exports["isQueue"].(*objects.Builtin)
	if !ok {
		t.Fatal("isQueue not found or not builtin")
	}

	// Create a Queue object
	q := objects.NewQueue()
	res := fn.Fn(q)
	if b, ok := res.(*objects.Bool); !ok || !b.Value {
		t.Fatalf("isQueue should return true for Queue, got %T %v", res, res)
	}

	// Test with non-Queue objects
	nonQueueTypes := []objects.Object{
		String("not a queue"),
		Int(123),
		Bool(false),
		Null(),
		Array(Int(1)),
	}
	for _, obj := range nonQueueTypes {
		res2 := fn.Fn(obj)
		if b, ok := res2.(*objects.Bool); !ok || b.Value {
			t.Fatalf("isQueue should return false for %T, got %T %v", obj, res2, res2)
		}
	}
}
