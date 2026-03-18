// pkg/vm/value_stack.go
// Value-based stack implementation optimized for NaN-boxed values
package vm

// ValueStackSize is the maximum number of elements on the stack
const ValueStackSize = 2048

// ValueStack represents the VM operand stack using NaN-boxed values
type ValueStack struct {
	data       []Value
	sp         int   // Stack pointer (points to next free slot)
	lastPopped Value // Last popped element
}

// NewValueStack creates a new value stack
func NewValueStack() *ValueStack {
	return &ValueStack{
		data: make([]Value, ValueStackSize),
		sp:   0,
	}
}

// Push pushes a value onto the stack
// Inlined in hot paths for performance
func (s *ValueStack) Push(v Value) error {
	// Check bounds only when approaching limit
	if s.sp >= ValueStackSize-10 {
		if s.sp >= ValueStackSize {
			return errStackOverflow
		}
	}
	s.data[s.sp] = v
	s.sp++
	return nil
}

// MustPush pushes without error checking (for hot paths)
func (s *ValueStack) MustPush(v Value) {
	s.data[s.sp] = v
	s.sp++
}

// Pop pops a value from the stack
func (s *ValueStack) Pop() Value {
	if s.sp == 0 {
		return ValueNull
	}
	s.sp--
	v := s.data[s.sp]
	s.data[s.sp] = ValueNull // Clear for GC
	s.lastPopped = v
	return v
}

// PopNoClear pops without clearing (slightly faster, delays GC)
func (s *ValueStack) PopNoClear() Value {
	s.sp--
	v := s.data[s.sp]
	s.lastPopped = v
	return v
}

// Top returns the top element without removing it
func (s *ValueStack) Top() Value {
	if s.sp == 0 {
		return ValueNull
	}
	return s.data[s.sp-1]
}

// Peek returns the n-th element from the top (0 = top)
func (s *ValueStack) Peek(n int) Value {
	if s.sp <= n {
		return ValueNull
	}
	return s.data[s.sp-1-n]
}

// SetTop sets the top element
func (s *ValueStack) SetTop(v Value) {
	if s.sp > 0 {
		s.data[s.sp-1] = v
	}
}

// Len returns the number of elements on the stack
func (s *ValueStack) Len() int {
	return s.sp
}

// LastPopped returns the last popped element
func (s *ValueStack) LastPopped() Value {
	return s.lastPopped
}

// Reset clears the stack
func (s *ValueStack) Reset() {
	s.sp = 0
	for i := range s.data {
		s.data[i] = ValueNull
	}
}

// Get returns the value at index i
func (s *ValueStack) Get(i int) Value {
	if i < 0 || i >= s.sp {
		return ValueNull
	}
	return s.data[i]
}

// Set sets the value at index i
func (s *ValueStack) Set(i int, v Value) {
	if i >= 0 && i < ValueStackSize {
		s.data[i] = v
	}
}

// Drop removes n elements from the top
func (s *ValueStack) Drop(n int) {
	for i := 0; i < n && s.sp > 0; i++ {
		s.sp--
		s.data[s.sp] = ValueNull
	}
}

// Dup duplicates the top element
func (s *ValueStack) Dup() error {
	if s.sp == 0 {
		return errEmptyStack
	}
	return s.Push(s.data[s.sp-1])
}

// Swap swaps the top two elements
func (s *ValueStack) Swap() {
	if s.sp >= 2 {
		s.data[s.sp-1], s.data[s.sp-2] = s.data[s.sp-2], s.data[s.sp-1]
	}
}

// Rot3 rotates the top 3 elements: [a, b, c] -> [b, c, a]
func (s *ValueStack) Rot3() {
	if s.sp >= 3 {
		a := s.data[s.sp-3]
		s.data[s.sp-3] = s.data[s.sp-2]
		s.data[s.sp-2] = s.data[s.sp-1]
		s.data[s.sp-1] = a
	}
}

// Errors
var errStackOverflow = &stackError{"stack overflow"}
var errEmptyStack = &stackError{"stack underflow"}

type stackError struct {
	msg string
}

func (e *stackError) Error() string {
	return e.msg
}
