// pkg/objects/queue.go
// Queue type for Xxlang - FIFO data structure with O(1) push/pop operations.
package objects

import (
	"strconv"
	"unsafe"
)

// Queue represents a FIFO queue (not thread-safe).
type Queue struct {
	items []Object
	head  int // index of first element
	tail  int // index where next element will be inserted
	count int // number of elements
}

// NewQueue creates a new empty queue.
func NewQueue() *Queue {
	return &Queue{
		items: make([]Object, 16), // initial capacity
	}
}

// NewQueueWithCapacity creates a new queue with the specified initial capacity.
func NewQueueWithCapacity(capacity int) *Queue {
	if capacity < 4 {
		capacity = 4
	}
	return &Queue{
		items: make([]Object, capacity),
	}
}

// NewQueueFrom creates a new queue from an array.
func NewQueueFrom(arr *Array) *Queue {
	q := NewQueueWithCapacity(len(arr.Elements) + 16)
	for _, elem := range arr.Elements {
		q.Push(elem)
	}
	return q
}

// Type returns the object type.
func (q *Queue) Type() ObjectType { return QueueType }

// TypeTag returns the fast type tag.
func (q *Queue) TypeTag() TypeTag { return TagQueue }

// Inspect returns a string representation.
func (q *Queue) Inspect() string {
	return "Queue(len=" + strconv.Itoa(q.count) + ")"
}

// ToBool returns true (Queue is always truthy).
func (q *Queue) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the Queue.
func (q *Queue) HashKey() HashKey {
	return HashKey{
		Type:  QueueType,
		Value: uint64(uintptr(unsafe.Pointer(q))),
	}
}

// Push adds an element to the back of the queue.
func (q *Queue) Push(item Object) {
	// Grow the buffer if needed
	if q.count == len(q.items) {
		q.grow()
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % len(q.items)
	q.count++
}

// Pop removes and returns the front element of the queue.
// Returns NULL if the queue is empty.
func (q *Queue) Pop() Object {
	if q.count == 0 {
		return NULL
	}

	item := q.items[q.head]
	q.items[q.head] = nil // allow GC
	q.head = (q.head + 1) % len(q.items)
	q.count--
	return item
}

// Peek returns the front element without removing it.
// Returns NULL if the queue is empty.
func (q *Queue) Peek() Object {
	if q.count == 0 {
		return NULL
	}
	return q.items[q.head]
}

// PeekBack returns the back element without removing it.
// Returns NULL if the queue is empty.
func (q *Queue) PeekBack() Object {
	if q.count == 0 {
		return NULL
	}
	idx := (q.tail - 1 + len(q.items)) % len(q.items)
	return q.items[idx]
}

// Len returns the number of elements in the queue.
func (q *Queue) Len() int {
	return q.count
}

// IsEmpty returns true if the queue is empty.
func (q *Queue) IsEmpty() bool {
	return q.count == 0
}

// Clear removes all elements from the queue.
func (q *Queue) Clear() {
	// Clear all references for GC
	for i := range q.items {
		q.items[i] = nil
	}
	q.head = 0
	q.tail = 0
	q.count = 0
}

// ToArray returns all elements as an array (front to back).
func (q *Queue) ToArray() *Array {
	elements := make([]Object, q.count)
	for i := 0; i < q.count; i++ {
		idx := (q.head + i) % len(q.items)
		elements[i] = q.items[idx]
	}
	return &Array{Elements: elements}
}

// Clone returns a shallow copy of the queue.
func (q *Queue) Clone() *Queue {
	newQ := &Queue{
		items: make([]Object, len(q.items)),
		head:  q.head,
		tail:  q.tail,
		count: q.count,
	}
	copy(newQ.items, q.items)
	return newQ
}

// grow increases the capacity of the ring buffer.
func (q *Queue) grow() {
	newItems := make([]Object, len(q.items)*2)
	for i := 0; i < q.count; i++ {
		idx := (q.head + i) % len(q.items)
		newItems[i] = q.items[idx]
	}
	q.items = newItems
	q.head = 0
	q.tail = q.count
}