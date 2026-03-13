// pkg/vm/stack.go
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/objects"
)

// StackSize is the maximum number of elements on the stack
const StackSize = 2048

// Stack represents the VM operand stack
type Stack struct {
	data       []objects.Object
	sp         int            // Stack pointer (points to next free slot)
	lastPopped objects.Object // Last popped element
}

// NewStack creates a new stack with the default size
func NewStack() *Stack {
	return &Stack{
		data: make([]objects.Object, StackSize),
		sp:   0,
	}
}

// Push pushes an object onto the stack
// Optimized: bounds check only when near limit
func (s *Stack) Push(obj objects.Object) error {
	// Check bounds only when approaching limit
	if s.sp >= StackSize-10 {
		if s.sp >= StackSize {
			return fmt.Errorf("stack overflow")
		}
	}
	s.data[s.sp] = obj
	s.sp++
	return nil
}

// Pop pops an object from the stack
// Optimized: skip nil check when stack is known to be non-empty
func (s *Stack) Pop() objects.Object {
	if s.sp == 0 {
		return nil
	}
	s.sp--
	obj := s.data[s.sp]
	s.data[s.sp] = nil // Allow GC to collect the object
	s.lastPopped = obj
	return obj
}

// PopSkipGC pops an object without clearing the reference (faster, but may delay GC)
func (s *Stack) PopSkipGC() objects.Object {
	s.sp--
	obj := s.data[s.sp]
	s.lastPopped = obj
	return obj
}

// Top returns the top element without removing it
func (s *Stack) Top() objects.Object {
	if s.sp == 0 {
		return nil
	}
	return s.data[s.sp-1]
}

// Peek returns the n-th element from the top (0 = top, 1 = second from top, etc.)
func (s *Stack) Peek(n int) objects.Object {
	if s.sp <= n {
		return nil
	}
	return s.data[s.sp-1-n]
}

// Len returns the number of elements on the stack
func (s *Stack) Len() int {
	return s.sp
}

// LastPopped returns the last popped element
func (s *Stack) LastPopped() objects.Object {
	return s.lastPopped
}

// SetTop sets the top element without changing stack pointer
func (s *Stack) SetTop(obj objects.Object) {
	if s.sp > 0 {
		s.data[s.sp-1] = obj
	}
}
