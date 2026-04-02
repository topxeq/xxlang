// pkg/objects/queue_test.go
package objects

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Fatal("expected queue instance")
	}
	if q.Len() != 0 {
		t.Errorf("expected empty queue, got %d elements", q.Len())
	}
}

func TestNewQueueWithCapacity(t *testing.T) {
	q := NewQueueWithCapacity(10)
	if q == nil {
		t.Fatal("expected queue instance")
	}
}

func TestNewQueueFrom(t *testing.T) {
	arr := &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}}
	q := NewQueueFrom(arr)
	if q == nil {
		t.Fatal("expected queue instance")
	}
	if q.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", q.Len())
	}
}

func TestQueuePushPop(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	q.Push(NewInt(2))
	q.Push(NewInt(3))

	if q.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", q.Len())
	}

	item := q.Pop()
	if item.(*Int).Value != 1 {
		t.Errorf("expected 1, got %v", item)
	}
	item = q.Pop()
	if item.(*Int).Value != 2 {
		t.Errorf("expected 2, got %v", item)
	}
}

func TestQueuePopEmpty(t *testing.T) {
	q := NewQueue()
	item := q.Pop()
	if item != NULL {
		t.Errorf("expected NULL for empty queue, got %v", item)
	}
}

func TestQueuePeek(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	q.Push(NewInt(2))

	item := q.Peek()
	if item.(*Int).Value != 1 {
		t.Errorf("expected 1, got %v", item)
	}
	if q.Len() != 2 {
		t.Error("Peek should not remove element")
	}
}

func TestQueuePeekEmpty(t *testing.T) {
	q := NewQueue()
	item := q.Peek()
	if item != NULL {
		t.Errorf("expected NULL for empty queue, got %v", item)
	}
}

func TestQueuePeekBack(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	q.Push(NewInt(2))

	item := q.PeekBack()
	if item.(*Int).Value != 2 {
		t.Errorf("expected 2, got %v", item)
	}
}

func TestQueueIsEmpty(t *testing.T) {
	q := NewQueue()
	if !q.IsEmpty() {
		t.Error("expected new queue to be empty")
	}
	q.Push(NewInt(1))
	if q.IsEmpty() {
		t.Error("expected queue with element to not be empty")
	}
}

func TestQueueClear(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	q.Push(NewInt(2))
	q.Clear()
	if q.Len() != 0 {
		t.Errorf("expected empty queue after Clear, got %d", q.Len())
	}
}

func TestQueueToArray(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	q.Push(NewInt(2))
	arr := q.ToArray()
	if arr == nil {
		t.Fatal("expected array")
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}
}

func TestQueueClone(t *testing.T) {
	q := NewQueue()
	q.Push(NewInt(1))
	cloned := q.Clone()
	if cloned.Len() != q.Len() {
		t.Error("expected cloned queue to have same length")
	}
}

func TestQueueGrow(t *testing.T) {
	q := NewQueueWithCapacity(4)
	for i := 0; i < 20; i++ {
		q.Push(NewInt(int64(i)))
	}
	if q.Len() != 20 {
		t.Errorf("expected 20 elements, got %d", q.Len())
	}
}
