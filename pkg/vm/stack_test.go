// pkg/vm/stack_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestStackPushPop(t *testing.T) {
	s := NewStack()

	// Push some values
	val1 := &objects.Int{Value: 10}
	val2 := &objects.Int{Value: 20}
	val3 := &objects.String{Value: "hello"}

	if err := s.Push(val1); err != nil {
		t.Fatalf("Push failed: %v", err)
	}
	if err := s.Push(val2); err != nil {
		t.Fatalf("Push failed: %v", err)
	}
	if err := s.Push(val3); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if s.Len() != 3 {
		t.Errorf("expected stack length 3, got %d", s.Len())
	}

	// Pop in reverse order
	popped := s.Pop()
	if popped.Type() != objects.StringType {
		t.Errorf("expected String, got %s", popped.Type())
	}
	if popped.(*objects.String).Value != "hello" {
		t.Errorf("expected 'hello', got %s", popped.(*objects.String).Value)
	}

	popped = s.Pop()
	if popped.(*objects.Int).Value != 20 {
		t.Errorf("expected 20, got %d", popped.(*objects.Int).Value)
	}

	popped = s.Pop()
	if popped.(*objects.Int).Value != 10 {
		t.Errorf("expected 10, got %d", popped.(*objects.Int).Value)
	}

	if s.Len() != 0 {
		t.Errorf("expected empty stack, got length %d", s.Len())
	}
}

func TestStackTop(t *testing.T) {
	s := NewStack()

	// Top on empty stack
	top := s.Top()
	if top != nil {
		t.Errorf("expected nil on empty stack, got %v", top)
	}

	// Push and check top
	val1 := &objects.Int{Value: 42}
	s.Push(val1)

	top = s.Top()
	if top == nil {
		t.Fatal("expected value, got nil")
	}
	if top.(*objects.Int).Value != 42 {
		t.Errorf("expected 42, got %d", top.(*objects.Int).Value)
	}

	// Top should not remove the element
	if s.Len() != 1 {
		t.Errorf("Top should not remove element, expected length 1, got %d", s.Len())
	}

	// Push another and check top
	val2 := &objects.Int{Value: 100}
	s.Push(val2)

	top = s.Top()
	if top.(*objects.Int).Value != 100 {
		t.Errorf("expected 100, got %d", top.(*objects.Int).Value)
	}
}

func TestStackOverflow(t *testing.T) {
	s := NewStack()

	// Fill the stack to capacity
	for i := 0; i < StackSize; i++ {
		err := s.Push(&objects.Int{Value: int64(i)})
		if err != nil {
			t.Fatalf("Push at index %d failed: %v", i, err)
		}
	}

	// Next push should overflow
	err := s.Push(&objects.Int{Value: 999})
	if err == nil {
		t.Error("expected stack overflow error, got nil")
	}
}

func TestStackUnderflow(t *testing.T) {
	s := NewStack()

	// Pop on empty stack should return nil (not panic)
	popped := s.Pop()
	if popped != nil {
		t.Errorf("expected nil on empty stack pop, got %v", popped)
	}

	// Push one, pop one, then pop again
	s.Push(&objects.Int{Value: 1})
	s.Pop()

	popped = s.Pop()
	if popped != nil {
		t.Errorf("expected nil on empty stack pop, got %v", popped)
	}
}

func TestStackPeek(t *testing.T) {
	s := NewStack()

	// Peek on empty stack
	p := s.Peek(0)
	if p != nil {
		t.Errorf("expected nil on empty stack peek, got %v", p)
	}

	// Push values
	s.Push(&objects.Int{Value: 1})
	s.Push(&objects.Int{Value: 2})
	s.Push(&objects.Int{Value: 3})

	// Peek(0) should be same as Top
	p = s.Peek(0)
	if p.(*objects.Int).Value != 3 {
		t.Errorf("Peek(0): expected 3, got %d", p.(*objects.Int).Value)
	}

	// Peek(1) should be second from top
	p = s.Peek(1)
	if p.(*objects.Int).Value != 2 {
		t.Errorf("Peek(1): expected 2, got %d", p.(*objects.Int).Value)
	}

	// Peek(2) should be third from top
	p = s.Peek(2)
	if p.(*objects.Int).Value != 1 {
		t.Errorf("Peek(2): expected 1, got %d", p.(*objects.Int).Value)
	}

	// Peek beyond stack size
	p = s.Peek(10)
	if p != nil {
		t.Errorf("Peek beyond stack: expected nil, got %v", p)
	}

	// Peek should not modify stack
	if s.Len() != 3 {
		t.Errorf("Peek should not modify stack, expected length 3, got %d", s.Len())
	}
}

func TestStackLen(t *testing.T) {
	s := NewStack()

	if s.Len() != 0 {
		t.Errorf("expected empty stack, got length %d", s.Len())
	}

	s.Push(&objects.Int{Value: 1})
	if s.Len() != 1 {
		t.Errorf("expected length 1, got %d", s.Len())
	}

	s.Push(&objects.Int{Value: 2})
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}

	s.Pop()
	if s.Len() != 1 {
		t.Errorf("expected length 1 after pop, got %d", s.Len())
	}

	s.Pop()
	if s.Len() != 0 {
		t.Errorf("expected empty stack after pop, got length %d", s.Len())
	}
}
