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
package vm

import (
	"math"
	"unsafe"

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

// boxedObject holds an object for NaN boxing
// We need this because we can't take the address of an interface directly
type boxedObject struct {
	obj objects.Object
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

	// For other types, store a pointer to a boxed object
	box := &boxedObject{obj: obj}
	ptr := uintptr(unsafe.Pointer(box))
	return Value(tagObjectValue | uint64(ptr))
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

// GetObject extracts the object pointer
func (v Value) GetObject() objects.Object {
	ptr := uintptr(uint64(v) & payloadMask)
	box := (*boxedObject)(unsafe.Pointer(ptr))
	if box == nil {
		return objects.NULL
	}
	return box.obj
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
