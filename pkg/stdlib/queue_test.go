// pkg/stdlib/queue_test.go
// Tests for queue module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestQueueCreate(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	fn := mod.Exports["create"].(*objects.Builtin)

	// Create empty queue
	q := fn.Fn()
	if q.Type() != objects.QueueType {
		t.Fatalf("expected Queue, got %s", q.Type())
	}
	queue := q.(*objects.Queue)
	if queue.Len() != 0 {
		t.Errorf("expected empty queue size 0, got %d", queue.Len())
	}
}

func TestQueueFromArray(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	fn := mod.Exports["fromArray"].(*objects.Builtin)

	elements := []objects.Object{objects.NewString("a"), objects.NewString("b")}
	arr := objects.NewArray(elements)
	q := fn.Fn(arr)
	if q.Type() != objects.QueueType {
		t.Fatalf("expected Queue, got %s", q.Type())
	}
	queue := q.(*objects.Queue)
	if queue.Len() != 2 {
		t.Errorf("expected queue size 2, got %d", queue.Len())
	}
}

func TestQueueIsQueue(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	fn := mod.Exports["isQueue"].(*objects.Builtin)

	// Non-queue object
	result := fn.Fn(objects.NULL)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != false {
			t.Errorf("expected false for NULL, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}

	// Queue object
	q := objects.NewQueue()
	result = fn.Fn(q)
	if b, ok := result.(*objects.Bool); ok {
		if b.Value != true {
			t.Errorf("expected true for Queue, got %v", b.Value)
		}
	} else {
		t.Fatalf("expected Bool, got %T", result)
	}
}

func TestQueueEnqueueDequeue(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	// Use NewQueue directly for detailed method testing
	q := objects.NewQueue()

	// Enqueue items
	q.Push(objects.NewString("first"))
	q.Push(objects.NewString("second"))
	if q.Len() != 2 {
		t.Errorf("expected length 2, got %d", q.Len())
	}

	// Dequeue
	item := q.Pop()
	if str, ok := item.(*objects.String); ok {
		if str.Value != "first" {
			t.Errorf("expected 'first', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", item)
	}
	if q.Len() != 1 {
		t.Errorf("expected length 1 after dequeue, got %d", q.Len())
	}

	// Peek
	item = q.Peek()
	if str, ok := item.(*objects.String); ok {
		if str.Value != "second" {
			t.Errorf("expected 'second', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", item)
	}
	if q.Len() != 1 {
		t.Errorf("expected length still 1 after peek, got %d", q.Len())
	}

	// Dequeue remaining
	item = q.Pop()
	if str, ok := item.(*objects.String); ok {
		if str.Value != "second" {
			t.Errorf("expected 'second', got %s", str.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", item)
	}
	if q.Len() != 0 {
		t.Errorf("expected empty queue, got %d", q.Len())
	}

	// Dequeue from empty returns objects.NULL
	item = q.Pop()
	if item != objects.NULL {
		t.Errorf("expected objects.NULL from dequeue on empty, got %v", item)
	}

	// Peek on empty returns objects.NULL
	item = q.Peek()
	if item != objects.NULL {
		t.Errorf("expected objects.NULL from peek on empty, got %v", item)
	}
}

func TestQueueIsEmpty(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	q := objects.NewQueue()

	if !q.IsEmpty() {
		t.Error("expected empty queue to be empty")
	}
	q.Push(objects.NewInt(1))
	if q.IsEmpty() {
		t.Error("expected non-empty queue not to be empty")
	}
}

func TestQueueClear(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	q := objects.NewQueue()
	q.Push(objects.NewString("a"))
	q.Push(objects.NewString("b"))
	q.Clear()
	if q.Len() != 0 {
		t.Errorf("expected empty after clear, got %d", q.Len())
	}
	if !q.IsEmpty() {
		t.Error("expected empty queue to be empty after clear")
	}
}

func TestQueueToArray(t *testing.T) {
	mod := Get("queue")
	if mod == nil {
		t.Skip("queue module not found")
	}
	q := objects.NewQueue()
	q.Push(objects.NewString("x"))
	q.Push(objects.NewString("y"))
	q.Push(objects.NewString("z"))

	arr := q.ToArray()
	if len(arr.Elements) != 3 {
		t.Errorf("expected array length 3, got %d", len(arr.Elements))
	}
	// Check order preserved
	e0 := arr.Elements[0]
	e1 := arr.Elements[1]
	e2 := arr.Elements[2]
	if str0, ok := e0.(*objects.String); ok {
		if str0.Value != "x" {
			t.Errorf("expected 'x', got %s", str0.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", e0)
	}
	if str1, ok := e1.(*objects.String); ok {
		if str1.Value != "y" {
			t.Errorf("expected 'y', got %s", str1.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", e1)
	}
	if str2, ok := e2.(*objects.String); ok {
		if str2.Value != "z" {
			t.Errorf("expected 'z', got %s", str2.Value)
		}
	} else {
		t.Fatalf("expected String, got %T", e2)
	}
}
