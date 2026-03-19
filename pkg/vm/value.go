// pkg/vm/value.go
// NaN Boxing implementation for efficient value representation
//
// NaN boxing uses the IEEE 754 NaN representation to encode multiple types
// in a single 64-bit value without heap allocation.
//
// Memory layout:
// - Floats: stored as native IEEE 754 double (no NaN boxing needed)
// - Integers: stored in NaN payload (48 bits)
// - Booleans: stored in NaN payload
// - Null: special NaN value
// - Objects: stored as pointer in NaN payload
//
// IMPORTANT: For boxed objects, we maintain a global registry to ensure
// GC visibility. The Value stores an index into this registry.
package vm

import (
	"fmt"
	"math"
	"sync"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// Value is a NaN-boxed value that can represent multiple types efficiently
type Value uint64

// Tag values (stored in upper 16 bits when using NaN boxing)
const (
	tagInt    = 0x7FFC // Tagged integer
	tagBool   = 0x7FFD // Tagged boolean
	tagNull   = 0x7FFE // Tagged null
	tagObject = 0x7FFF // Tagged object pointer
)

// Payload mask (48 bits for data storage)
const payloadMask = 0x0000FFFFFFFFFFFF

// Tag values in position
const (
	tagIntValue    = uint64(tagInt) << 48
	tagBoolValue   = uint64(tagBool) << 48
	tagNullValue   = uint64(tagNull) << 48
	tagObjectValue = uint64(tagObject) << 48
)

// Special values - pre-computed for fast access
var (
	ValueNull  = Value(tagNullValue)
	ValueTrue  = Value(tagBoolValue | 1)
	ValueFalse = Value(tagBoolValue)
)

// NaN boxing boundary - values below this are normal floats
const nanBoundary = 0x7FF8000000000000

// NewInt creates a Value from an integer
// Stores up to 48 bits of integer data
func NewInt(n int64) Value {
	// Handle sign extension for negative numbers
	// Store as unsigned 48-bit value
	return Value(tagIntValue | (uint64(n) & payloadMask))
}

// NewFloat creates a Value from a float64
// Normal floats are stored directly as IEEE 754 doubles
func NewFloat(f float64) Value {
	bits := math.Float64bits(f)
	return Value(bits)
}

// NewBool creates a Value from a boolean
func NewBool(b bool) Value {
	if b {
		return ValueTrue
	}
	return ValueFalse
}

// objectRegistry is a GC-visible storage for boxed objects.
// This is critical: the registry itself is a GC root, so all stored
// objects remain reachable as long as the registry is alive.
// We use an index-based approach to store objects, which is safe
// because the GC can trace through the registry slice.
type objectRegistry struct {
	mu      sync.RWMutex
	objects []*objects.Object // Slice is GC-visible
	freeIdx []int             // Freed indices for reuse
	nextIdx int               // Next available index
}

// globalRegistry is the global object registry for the VM
var globalRegistry = &objectRegistry{
	objects: make([]*objects.Object, 1024), // Pre-allocate
	freeIdx: make([]int, 0),
}

// register stores an object and returns its index
func (r *objectRegistry) register(obj objects.Object) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Try to reuse a freed index
	if len(r.freeIdx) > 0 {
		idx := r.freeIdx[len(r.freeIdx)-1]
		r.freeIdx = r.freeIdx[:len(r.freeIdx)-1]
		r.objects[idx] = &obj
		return idx
	}

	// Allocate new index
	idx := r.nextIdx
	if idx >= len(r.objects) {
		// Grow the slice
		newObjects := make([]*objects.Object, len(r.objects)*2)
		copy(newObjects, r.objects)
		r.objects = newObjects
	}
	r.objects[idx] = &obj
	r.nextIdx++
	return idx
}

// get retrieves an object by index
func (r *objectRegistry) get(idx int) objects.Object {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if idx < 0 || idx >= len(r.objects) {
		return nil
	}
	obj := r.objects[idx]
	if obj == nil {
		return nil
	}
	return *obj
}

// release marks an index as free for reuse
func (r *objectRegistry) release(idx int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if idx >= 0 && idx < len(r.objects) {
		r.objects[idx] = nil
		r.freeIdx = append(r.freeIdx, idx)
	}
}

// Clear clears all objects from the registry
func (r *objectRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objects = make([]*objects.Object, 1024)
	r.freeIdx = r.freeIdx[:0]
	r.nextIdx = 0
}

// ClearRegistry clears the global object registry
// This should be called when starting a new execution context
func ClearRegistry() {
	globalRegistry.Clear()
}

// RegistryStats returns statistics about the object registry
func RegistryStats() (total, used int) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return len(globalRegistry.objects), globalRegistry.nextIdx - len(globalRegistry.freeIdx)
}

// NewObject creates a Value from an object pointer
func NewObject(obj objects.Object) Value {
	if obj == nil {
		return ValueNull
	}

	// Fast path for common types - unbox them
	switch o := obj.(type) {
	case *objects.Int:
		return NewInt(o.Value)
	case *objects.Float:
		return NewFloat(o.Value)
	case *objects.Bool:
		if o.Value {
			return ValueTrue
		}
		return ValueFalse
	case *objects.Null:
		return ValueNull
	}

	// For other types, store in the registry and use the index
	idx := globalRegistry.register(obj)
	return Value(tagObjectValue | uint64(idx))
}

// NewValue creates a Value from any object
func NewValue(obj objects.Object) Value {
	return NewObject(obj)
}

// ToObject converts a Value back to an objects.Object
func (v Value) ToObject() objects.Object {
	if v.IsNull() {
		return objects.NULL
	}
	if v.IsBool() {
		if v == ValueTrue {
			return objects.TRUE
		}
		return objects.FALSE
	}
	if v.IsInt() {
		return objects.NewInt(v.GetInt())
	}
	if v.IsFloat() {
		return &objects.Float{Value: v.GetFloat()}
	}
	if v.IsObject() {
		return v.GetObject()
	}
	// Treat as float by default
	return &objects.Float{Value: v.GetFloat()}
}

// Type checking methods

// IsInt returns true if the value is a tagged integer
func (v Value) IsInt() bool {
	return (uint64(v) >> 48) == tagInt
}

// IsFloat returns true if the value is a native float
func (v Value) IsFloat() bool {
	u := uint64(v)

	// Positive floats: 0x0000... to 0x7FF7... (below QNaN range)
	// Negative floats: 0x8000... to 0xFFFF... (have sign bit set)
	//
	// Our tagged values are in the range 0x7FFC... to 0x7FFF...
	// So a value is a float if it's:
	// - Below 0x7FFC000000000000 (positive floats, zero, etc.)
	// - Above 0x7FFFFFFFFFFFFFFF (negative floats)
	//
	// The only non-float values in the positive range are our tags:
	// 0x7FFC... (int), 0x7FFD... (bool), 0x7FFE... (null), 0x7FFF... (object)

	// Check if in our tag range (0x7FFC to 0x7FFF in high bits)
	highBits := u >> 48
	return highBits < tagInt || highBits > tagObject
}

// IsBool returns true if the value is a tagged boolean
func (v Value) IsBool() bool {
	return (uint64(v) >> 48) == tagBool
}

// IsNull returns true if the value is null
func (v Value) IsNull() bool {
	return v == ValueNull
}

// IsObject returns true if the value is a tagged object pointer
func (v Value) IsObject() bool {
	return (uint64(v) >> 48) == tagObject
}

// IsNumber returns true if the value is a number (int or float)
func (v Value) IsNumber() bool {
	return v.IsInt() || v.IsFloat()
}

// IsTruthy returns true if the value is truthy
func (v Value) IsTruthy() bool {
	if v.IsNull() {
		return false
	}
	if v == ValueFalse {
		return false
	}
	if v.IsInt() {
		return v.GetInt() != 0
	}
	if v.IsFloat() {
		return v.GetFloat() != 0
	}
	return true
}

// Value extraction methods

// GetInt extracts the integer value
func (v Value) GetInt() int64 {
	// Extract from payload and sign-extend from 48 bits
	payload := int64(uint64(v) & payloadMask)
	// Sign extend if the sign bit is set (bit 47)
	if payload >= (1 << 47) {
		payload -= (1 << 48)
	}
	return payload
}

// GetFloat extracts the float value
func (v Value) GetFloat() float64 {
	return math.Float64frombits(uint64(v))
}

// GetBool extracts the boolean value
func (v Value) GetBool() bool {
	return v == ValueTrue
}

// GetObject extracts the object from the registry
func (v Value) GetObject() objects.Object {
	idx := int(uint64(v) & payloadMask)
	obj := globalRegistry.get(idx)
	if obj == nil {
		return objects.NULL
	}
	return obj
}

// GetObjectNoLock extracts the object without mutex lock (for single-threaded use)
// This is faster but not safe for concurrent access
func (v Value) GetObjectNoLock() objects.Object {
	idx := int(uint64(v) & payloadMask)
	globalRegistry.mu.RLock()
	obj := globalRegistry.objects[idx]
	globalRegistry.mu.RUnlock()
	if obj == nil {
		return objects.NULL
	}
	return *obj
}

// IsClosure returns true if the value is a Closure
func (v Value) IsClosure() bool {
	if !v.IsObject() {
		return false
	}
	idx := int(uint64(v) & payloadMask)
	globalRegistry.mu.RLock()
	obj := globalRegistry.objects[idx]
	globalRegistry.mu.RUnlock()
	if obj == nil {
		return false
	}
	_, ok := (*obj).(*Closure)
	return ok
}

// IsCompiledFunction returns true if the value is a CompiledFunction
func (v Value) IsCompiledFunction() bool {
	if !v.IsObject() {
		return false
	}
	idx := int(uint64(v) & payloadMask)
	globalRegistry.mu.RLock()
	obj := globalRegistry.objects[idx]
	globalRegistry.mu.RUnlock()
	if obj == nil {
		return false
	}
	_, ok := (*obj).(*compiler.CompiledFunction)
	return ok
}

// GetClosure returns the Closure if this value is a Closure, or nil otherwise
func (v Value) GetClosure() *Closure {
	if !v.IsObject() {
		return nil
	}
	idx := int(uint64(v) & payloadMask)
	globalRegistry.mu.RLock()
	obj := globalRegistry.objects[idx]
	globalRegistry.mu.RUnlock()
	if obj == nil {
		return nil
	}
	if c, ok := (*obj).(*Closure); ok {
		return c
	}
	return nil
}

// GetCompiledFunction returns the CompiledFunction if this value is one, or nil otherwise
func (v Value) GetCompiledFunction() *compiler.CompiledFunction {
	if !v.IsObject() {
		return nil
	}
	idx := int(uint64(v) & payloadMask)
	globalRegistry.mu.RLock()
	obj := globalRegistry.objects[idx]
	globalRegistry.mu.RUnlock()
	if obj == nil {
		return nil
	}
	if f, ok := (*obj).(*compiler.CompiledFunction); ok {
		return f
	}
	return nil
}

// ToInt attempts to convert the value to an integer
func (v Value) ToInt() (int64, bool) {
	if v.IsInt() {
		return v.GetInt(), true
	}
	if v.IsFloat() {
		f := v.GetFloat()
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int64(f), true
		}
	}
	return 0, false
}

// ToFloat attempts to convert the value to a float
func (v Value) ToFloat() (float64, bool) {
	if v.IsFloat() {
		return v.GetFloat(), true
	}
	if v.IsInt() {
		return float64(v.GetInt()), true
	}
	return 0, false
}

// Arithmetic operations - hot path

// Add adds two values
func (v Value) Add(other Value) (Value, bool) {
	// Fast path: both integers
	if v.IsInt() && other.IsInt() {
		return NewInt(v.GetInt() + other.GetInt()), true
	}

	// String concatenation - check if either operand is a string object
	var str1, str2 string
	var hasStr1, hasStr2 bool

	if v.IsObject() {
		if s, ok := v.GetObject().(*objects.String); ok {
			str1 = s.Value
			hasStr1 = true
		}
	}
	if other.IsObject() {
		if s, ok := other.GetObject().(*objects.String); ok {
			str2 = s.Value
			hasStr2 = true
		}
	}

	// If either is a string, perform string concatenation
	if hasStr1 || hasStr2 {
		// Convert first operand to string if not already
		if !hasStr1 {
			if v.IsInt() {
				str1 = fmt.Sprintf("%d", v.GetInt())
			} else if v.IsFloat() {
				str1 = fmt.Sprintf("%g", v.GetFloat())
			} else if v.IsNull() {
				str1 = "null"
			} else if v.IsBool() {
				if v.GetBool() {
					str1 = "true"
				} else {
					str1 = "false"
				}
			} else {
				str1 = v.GetObject().Inspect()
			}
		}
		// Convert second operand to string if not already
		if !hasStr2 {
			if other.IsInt() {
				str2 = fmt.Sprintf("%d", other.GetInt())
			} else if other.IsFloat() {
				str2 = fmt.Sprintf("%g", other.GetFloat())
			} else if other.IsNull() {
				str2 = "null"
			} else if other.IsBool() {
				if other.GetBool() {
					str2 = "true"
				} else {
					str2 = "false"
				}
			} else {
				str2 = other.GetObject().Inspect()
			}
		}
		return NewObject(&objects.String{Value: str1 + str2}), true
	}

	// Mixed: convert to floats
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		return NewFloat(f1 + f2), true
	}
	return ValueNull, false
}

// Sub subtracts two values
func (v Value) Sub(other Value) (Value, bool) {
	if v.IsInt() && other.IsInt() {
		return NewInt(v.GetInt() - other.GetInt()), true
	}
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		return NewFloat(f1 - f2), true
	}
	return ValueNull, false
}

// Mul multiplies two values
func (v Value) Mul(other Value) (Value, bool) {
	if v.IsInt() && other.IsInt() {
		return NewInt(v.GetInt() * other.GetInt()), true
	}
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		return NewFloat(f1 * f2), true
	}
	return ValueNull, false
}

// Div divides two values
func (v Value) Div(other Value) (Value, bool) {
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		if f2 == 0 {
			return ValueNull, false
		}
		return NewFloat(f1 / f2), true
	}
	return ValueNull, false
}

// Mod computes modulo
func (v Value) Mod(other Value) (Value, bool) {
	if v.IsInt() && other.IsInt() {
		i2 := other.GetInt()
		if i2 == 0 {
			return ValueNull, false
		}
		return NewInt(v.GetInt() % i2), true
	}
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		if f2 == 0 {
			return ValueNull, false
		}
		return NewFloat(float64(int64(f1) % int64(f2))), true
	}
	return ValueNull, false
}

// Neg negates a value
func (v Value) Neg() (Value, bool) {
	if v.IsInt() {
		return NewInt(-v.GetInt()), true
	}
	if v.IsFloat() {
		return NewFloat(-v.GetFloat()), true
	}
	return ValueNull, false
}

// Comparison operations

// Less compares if v < other
func (v Value) Less(other Value) (bool, bool) {
	if v.IsInt() && other.IsInt() {
		return v.GetInt() < other.GetInt(), true
	}
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		return f1 < f2, true
	}
	return false, false
}

// Greater compares if v > other
func (v Value) Greater(other Value) (bool, bool) {
	if v.IsInt() && other.IsInt() {
		return v.GetInt() > other.GetInt(), true
	}
	f1, ok1 := v.ToFloat()
	f2, ok2 := other.ToFloat()
	if ok1 && ok2 {
		return f1 > f2, true
	}
	return false, false
}

// Equal compares if v == other
func (v Value) Equal(other Value) (bool, bool) {
	// Fast path: exact match (includes booleans, null, same integers)
	if v == other {
		return true, true
	}

	// Both integers - compare values
	if v.IsInt() && other.IsInt() {
		return v.GetInt() == other.GetInt(), true
	}

	// Both floats - compare values
	if v.IsFloat() && other.IsFloat() {
		return v.GetFloat() == other.GetFloat(), true
	}

	// Both booleans - already handled by == above, but explicit check
	if v.IsBool() && other.IsBool() {
		return v == other, true
	}

	// Both objects - compare using Object's Equal method
	if v.IsObject() && other.IsObject() {
		obj1 := v.GetObject()
		obj2 := other.GetObject()
		if obj1 == nil || obj2 == nil {
			return obj1 == obj2, true
		}
		// Use type-specific comparison
		if s1, ok1 := obj1.(*objects.String); ok1 {
			if s2, ok2 := obj2.(*objects.String); ok2 {
				return s1.Value == s2.Value, true
			}
		}
		// For other objects, use HashKey comparison for maps or direct comparison
		return obj1 == obj2, true
	}

	// Mixed int/float
	if v.IsNumber() && other.IsNumber() {
		f1, ok1 := v.ToFloat()
		f2, ok2 := other.ToFloat()
		if ok1 && ok2 {
			return f1 == f2, true
		}
	}

	return false, false
}

// NotEqual compares if v != other
func (v Value) NotEqual(other Value) (bool, bool) {
	eq, ok := v.Equal(other)
	return !eq, ok
}

// String returns a string representation for debugging
func (v Value) String() string {
	if v.IsNull() {
		return "null"
	}
	if v == ValueTrue {
		return "true"
	}
	if v == ValueFalse {
		return "false"
	}
	if v.IsInt() {
		return objects.NewInt(v.GetInt()).Inspect()
	}
	if v.IsFloat() {
		return (&objects.Float{Value: v.GetFloat()}).Inspect()
	}
	if v.IsObject() {
		obj := v.GetObject()
		if obj != nil {
			return obj.Inspect()
		}
		return "nil object"
	}
	return "unknown"
}

// ValueBool creates a Value from a boolean (alias for NewBool)
func ValueBool(b bool) Value {
	if b {
		return ValueTrue
	}
	return ValueFalse
}

// LessEqual compares if v <= other
func (v Value) LessEqual(other Value) Value {
	less, ok := v.Less(other)
	if !ok {
		return ValueFalse
	}
	if less {
		return ValueTrue
	}
	eq, ok := v.Equal(other)
	if !ok {
		return ValueFalse
	}
	return ValueBool(eq)
}

// GreaterEqual compares if v >= other
func (v Value) GreaterEqual(other Value) Value {
	greater, ok := v.Greater(other)
	if !ok {
		return ValueFalse
	}
	if greater {
		return ValueTrue
	}
	eq, ok := v.Equal(other)
	if !ok {
		return ValueFalse
	}
	return ValueBool(eq)
}

// EqualValue returns a Value representing equality comparison
func (v Value) EqualValue(other Value) Value {
	eq, _ := v.Equal(other)
	return ValueBool(eq)
}

// NotEqualValue returns a Value representing inequality comparison
func (v Value) NotEqualValue(other Value) Value {
	ne, _ := v.NotEqual(other)
	return ValueBool(ne)
}

// LessValue returns a Value representing less than comparison
func (v Value) LessValue(other Value) Value {
	less, _ := v.Less(other)
	return ValueBool(less)
}

// GreaterValue returns a Value representing greater than comparison
func (v Value) GreaterValue(other Value) Value {
	greater, _ := v.Greater(other)
	return ValueBool(greater)
}
