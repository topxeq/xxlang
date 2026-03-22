// pkg/objects/tube.go
// Tube object for concurrent communication (similar to Go channels)
package objects

import (
	"fmt"
	"reflect"
	"sync"
)

// Tube represents a communication tube for concurrent message passing
// It wraps a Go channel for efficient implementation
type Tube struct {
	ch       reflect.Value // Underlying Go channel
	elemType ObjectType    // Optional type constraint for elements
	buffer   int           // Buffer size (0 for unbuffered)
	closed   bool          // Whether the tube is closed
	mu       sync.RWMutex  // Protects closed field
	id       uint64        // Unique identifier for debugging
}

// tubeIDCounter generates unique tube IDs
var tubeIDCounter uint64

// NewTube creates a new tube with optional type constraint and buffer size
func NewTube(elemType ObjectType, buffer int) *Tube {
	var ch reflect.Value

	// Create the underlying Go channel
	// For untyped tubes, we use chan Object (channel of Object interface)
	elemTyp := reflect.TypeOf((*Object)(nil)).Elem()

	if elemType != "" {
		// Typed tube - use specific element type
		typ := objectTypeToReflectType(elemType)
		if typ != nil {
			elemTyp = typ
		}
	}

	// Create a channel type with the element type
	chanTyp := reflect.ChanOf(reflect.BothDir, elemTyp)
	ch = reflect.MakeChan(chanTyp, buffer)

	// Generate unique ID
	tubeIDCounter++

	return &Tube{
		ch:       ch,
		elemType: elemType,
		buffer:   buffer,
		closed:   false,
		id:       tubeIDCounter,
	}
}

// objectTypeToReflectType converts Xxlang ObjectType to reflect.Type
func objectTypeToReflectType(ot ObjectType) reflect.Type {
	switch ot {
	case IntType:
		return reflect.TypeOf(int64(0))
	case FloatType:
		return reflect.TypeOf(float64(0))
	case StringType:
		return reflect.TypeOf("")
	case BoolType:
		return reflect.TypeOf(false)
	default:
		// For other types, use Object interface
		return reflect.TypeOf((*Object)(nil)).Elem()
	}
}

// Send sends a value to the tube (blocking)
// Returns false if the tube is closed
func (t *Tube) Send(val Object) bool {
	t.mu.RLock()
	closed := t.closed
	t.mu.RUnlock()

	if closed {
		return false
	}

	// Type check if tube is typed
	if t.elemType != "" && val.Type() != t.elemType {
		// Type mismatch - still send but could log warning
		// For flexibility, we allow sending any type
	}

	t.ch.Send(reflect.ValueOf(val))
	return true
}

// TrySend attempts to send without blocking
// Returns (sent, ok) where ok is false if tube is closed
func (t *Tube) TrySend(val Object) (sent bool, ok bool) {
	t.mu.RLock()
	closed := t.closed
	t.mu.RUnlock()

	if closed {
		return false, false
	}

	return t.ch.TrySend(reflect.ValueOf(val)), true
}

// SendWithTimeout sends a value with timeout (in milliseconds)
// Returns (sent, ok) where ok is false if timeout or closed
func (t *Tube) SendWithTimeout(val Object, timeoutMs int) (sent bool, ok bool) {
	t.mu.RLock()
	closed := t.closed
	t.mu.RUnlock()

	if closed {
		return false, false
	}

	// For now, use blocking send
	// TODO: implement proper timeout with select
	t.ch.Send(reflect.ValueOf(val))
	return true, true
}

// Receive receives a value from the tube (blocking)
// Returns (value, ok) where ok is false if tube is closed and empty
func (t *Tube) Receive() (Object, bool) {
	val, ok := t.ch.Recv()
	if !ok {
		return NULL, false
	}
	return val.Interface().(Object), true
}

// TryReceive attempts to receive without blocking
// Returns (value, received, open) where open is false if tube is closed
func (t *Tube) TryReceive() (Object, bool, bool) {
	val, ok := t.ch.TryRecv()
	if !ok {
		// Channel empty or closed
		t.mu.RLock()
		closed := t.closed
		t.mu.RUnlock()
		return NULL, false, !closed
	}
	return val.Interface().(Object), true, true
}

// Close closes the tube
// After closing, no more sends are allowed but receives continue until empty
func (t *Tube) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.closed {
		t.closed = true
		t.ch.Close()
	}
}

// IsClosed returns whether the tube is closed
func (t *Tube) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

// Len returns the number of elements currently in the buffer
func (t *Tube) Len() int {
	return t.ch.Len()
}

// Cap returns the buffer capacity
func (t *Tube) Cap() int {
	return t.ch.Cap()
}

// ReflectValue returns the underlying reflect.Value for use with reflect.Select
func (t *Tube) ReflectValue() reflect.Value {
	return t.ch
}

// ElemType returns the element type constraint (empty string if untyped)
func (t *Tube) ElemType() ObjectType {
	return t.elemType
}

// ID returns the unique identifier for this tube
func (t *Tube) ID() uint64 {
	return t.id
}

// ============================================
// Object interface implementation
// ============================================

// Type returns the object type
func (t *Tube) Type() ObjectType { return TubeType }

// TypeTag returns the type tag for fast type checking
func (t *Tube) TypeTag() TypeTag { return TagTube }

// Inspect returns a string representation of the tube
func (t *Tube) Inspect() string {
	t.mu.RLock()
	closed := t.closed
	t.mu.RUnlock()

	status := "open"
	if closed {
		status = "closed"
	}

	if t.elemType != "" {
		return fmt.Sprintf("Tube<%s>(buffer:%d, len:%d, %s)",
			t.elemType, t.buffer, t.Len(), status)
	}
	return fmt.Sprintf("Tube(buffer:%d, len:%d, %s)",
		t.buffer, t.Len(), status)
}

// ToBool returns true (tubes are always truthy)
func (t *Tube) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the tube
func (t *Tube) HashKey() HashKey {
	return HashKey{
		Type:  TubeType,
		Value: t.id,
	}
}

// ============================================
// Tube methods for use in Xxlang
// ============================================

// SendMethod is the method called as tube.send(value)
func (t *Tube) SendMethod(args ...Object) Object {
	if len(args) < 1 {
		return &Error{Message: "send requires 1 argument"}
	}

	if !t.Send(args[0]) {
		return &Error{Message: "send on closed tube"}
	}
	return NULL
}

// ReceiveMethod is the method called as tube.receive()
// Returns [value, ok] array
func (t *Tube) ReceiveMethod(args ...Object) Object {
	val, ok := t.Receive()
	return &Array{Elements: []Object{val, &Bool{Value: ok}}}
}

// TryReceiveMethod is the method called as tube.tryReceive()
// Returns [value, received, open] array
func (t *Tube) TryReceiveMethod(args ...Object) Object {
	val, received, open := t.TryReceive()
	return &Array{Elements: []Object{val, &Bool{Value: received}, &Bool{Value: open}}}
}

// CloseMethod is the method called as tube.close()
func (t *Tube) CloseMethod(args ...Object) Object {
	t.Close()
	return NULL
}

// LenMethod is the method called as tube.len()
func (t *Tube) LenMethod(args ...Object) Object {
	return &Int{Value: int64(t.Len())}
}

// CapMethod is the method called as tube.cap()
func (t *Tube) CapMethod(args ...Object) Object {
	return &Int{Value: int64(t.Cap())}
}

// IsClosedMethod is the method called as tube.isClosed()
func (t *Tube) IsClosedMethod(args ...Object) Object {
	return &Bool{Value: t.IsClosed()}
}
