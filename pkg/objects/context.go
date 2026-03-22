// pkg/objects/context.go
// Context object for timeout and cancellation (similar to Go's context package)
package objects

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Context represents a cancellation context
type Context struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   *Tube // Tube that closes when context is done
	mu     sync.RWMutex
	id     uint64
}

// contextIDCounter generates unique context IDs
var contextIDCounter uint64

// NewContext creates a new context
func NewContext(ctx context.Context, cancel context.CancelFunc) *Context {
	contextIDCounter++
	return &Context{
		ctx:    ctx,
		cancel: cancel,
		id:     contextIDCounter,
	}
}

// NewBackgroundContext creates a background context (never cancels)
func NewBackgroundContext() *Context {
	ctx := context.Background()
	contextIDCounter++
	return &Context{
		ctx:    ctx,
		cancel: nil,
		id:     contextIDCounter,
	}
}

// NewContextWithTimeout creates a context that cancels after duration
func NewContextWithTimeout(parent *Context, timeout time.Duration) *Context {
	var parentCtx context.Context = context.Background()
	if parent != nil {
		parentCtx = parent.ctx
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	c := NewContext(ctx, cancel)
	return c
}

// NewContextWithCancel creates a context that can be cancelled manually
func NewContextWithCancel(parent *Context) *Context {
	var parentCtx context.Context = context.Background()
	if parent != nil {
		parentCtx = parent.ctx
	}

	ctx, cancel := context.WithCancel(parentCtx)
	return NewContext(ctx, cancel)
}

// NewContextWithDeadline creates a context that cancels at deadline
func NewContextWithDeadline(parent *Context, deadline time.Time) *Context {
	var parentCtx context.Context = context.Background()
	if parent != nil {
		parentCtx = parent.ctx
	}

	ctx, cancel := context.WithDeadline(parentCtx, deadline)
	return NewContext(ctx, cancel)
}

// Type returns the object type
func (c *Context) Type() ObjectType { return ContextType }

// TypeTag returns the type tag for fast type checking
func (c *Context) TypeTag() TypeTag { return TagContext }

// Inspect returns a string representation
func (c *Context) Inspect() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ctx.Err() != nil {
		return fmt.Sprintf("Context(id=%d, done, err=%v)", c.id, c.ctx.Err())
	}
	return fmt.Sprintf("Context(id=%d, active)", c.id)
}

// ToBool returns true if context is not done
func (c *Context) ToBool() *Bool {
	if c.ctx.Err() == nil {
		return TRUE
	}
	return FALSE
}

// HashKey returns a hash key for the context
func (c *Context) HashKey() HashKey {
	return HashKey{Type: ContextType, Value: c.id}
}

// Cancel cancels the context
func (c *Context) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}
}

// Done returns the done tube (creates it lazily)
// The tube is a zero-buffer tube that gets closed when context is done
func (c *Context) Done() *Tube {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.done != nil {
		return c.done
	}

	// Create a zero-buffer tube (unbuffered)
	// This tube will be closed when context is done
	c.done = NewTube("", 0)

	// Start a goroutine to close the tube when context is done
	go func() {
		<-c.ctx.Done()
		c.mu.Lock()
		if c.done != nil {
			c.done.Close()
		}
		c.mu.Unlock()
	}()

	return c.done
}

// Err returns the context error (nil if not done)
func (c *Context) Err() error {
	return c.ctx.Err()
}

// ErrString returns the error as a string
func (c *Context) ErrString() string {
	err := c.ctx.Err()
	if err == nil {
		return ""
	}
	return err.Error()
}

// IsDone returns true if context is done
func (c *Context) IsDone() bool {
	return c.ctx.Err() != nil
}

// Deadline returns the deadline time and whether a deadline is set
func (c *Context) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// DeadlineString returns the deadline as a string
func (c *Context) DeadlineString() string {
	dl, ok := c.ctx.Deadline()
	if !ok {
		return ""
	}
	return dl.Format(time.RFC3339)
}

// Value returns the value associated with key (for context values)
func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// GoContext returns the underlying Go context
func (c *Context) GoContext() context.Context {
	return c.ctx
}

// Context error types for Xxlang
var (
	ContextCanceled = &Error{Message: "context canceled"}
	ContextDeadlineExceeded = &Error{Message: "context deadline exceeded"}
)
