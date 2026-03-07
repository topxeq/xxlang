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
	data []objects.Object
	sp   int // Stack pointer (points to next free slot)
}

// NewStack creates a new stack with the default size
func NewStack() *Stack {
	return &Stack{
		data: make([]objects.Object, StackSize),
		sp:   0,
	}
}

// Push pushes an object onto the stack
func (s *Stack) Push(obj objects.Object) error {
	if s.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	s.data[s.sp] = obj
	s.sp++
	return nil
}

// Pop pops an object from the stack
func (s *Stack) Pop() objects.Object {
	if s.sp == 0 {
		return nil
	}
	s.sp--
	obj := s.data[s.sp]
	s.data[s.sp] = nil // Allow GC to collect the object
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
