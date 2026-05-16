//go:build windows && amd64
// +build windows,amd64

package jit

import (
	"sync"

	"github.com/topxeq/xxlang/pkg/objects"
)

// objectHandleTag is set in bit 62 of an int64 to indicate an object handle.
// Bit 62 is chosen because:
//   - Bit 63 (sign bit) is set for negative integers, so using bit 62 avoids
//     confusion with negative values while still leaving 62 bits for the handle ID.
//   - Values with bit 62 set but bit 63 clear (0x4000_0000_0000_0000 to
//     0x7FFF_FFFF_FFFF_FFFF, i.e., 2^62 to 2^63-1 ≈ 4.6×10^18) are extremely
//     rare as normal integer values in programs.
//   - If a plain integer happens to have bit 62 set, the registry lookup will
//     fail and it will be treated as an integer (safe fallback).
const objectHandleTag int64 = 1 << 62

// ObjectHandleRegistry maps int64 handles to objects for the JIT native code
// interface. Since native code can only pass int64 values, non-primitive objects
// (arrays, maps, instances, etc.) are registered and referenced by handle.
type ObjectHandleRegistry struct {
	mu     sync.RWMutex
	byID   map[int64]objects.Object
	nextID int64
}

// globalObjectRegistry is the shared object handle registry.
var globalObjectRegistry = &ObjectHandleRegistry{
	byID:   make(map[int64]objects.Object),
	nextID: 1,
}

// Register stores an object and returns a handle (int64) that native code can use
// to reference it. The handle has the objectHandleTag bit set to distinguish it
// from plain integer values.
func (r *ObjectHandleRegistry) Register(obj objects.Object) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	r.byID[id] = obj
	return id | objectHandleTag
}

// Lookup retrieves an object by its handle. Returns the object and true if found,
// or nil and false if the handle is not registered.
func (r *ObjectHandleRegistry) Lookup(handle int64) (objects.Object, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	obj, ok := r.byID[handle]
	return obj, ok
}

// Unregister removes a handle from the registry. This should be called when
// the native code no longer needs to reference the object, to prevent memory leaks.
func (r *ObjectHandleRegistry) Unregister(handle int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := handle &^ objectHandleTag
	delete(r.byID, id)
}

// Clear removes all registered objects. Used for testing or cleanup.
func (r *ObjectHandleRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID = make(map[int64]objects.Object)
	r.nextID = 1
}

// Size returns the number of currently registered objects (for diagnostics).
func (r *ObjectHandleRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byID)
}

// nativeIntToObject converts an int64 value from native code to an objects.Object.
// If the value has the objectHandleTag bit set and is found in the handle registry,
// the registered object is returned. Otherwise, the value is treated as a plain integer.
func nativeIntToObject(val int64) objects.Object {
	if val&objectHandleTag != 0 {
		id := val &^ objectHandleTag
		if obj, ok := globalObjectRegistry.Lookup(id); ok {
			return obj
		}
	}
	return &objects.Int{Value: val}
}

// objectToNativeInt converts an objects.Object to an int64 for native code.
// Primitive types (Int, Bool, Null) are converted directly.
// Non-primitive types (Array, Map, Instance, etc.) are registered in the
// handle registry and the handle is returned.
func objectToNativeInt(obj objects.Object) int64 {
	if obj == nil {
		return 0
	}
	switch v := obj.(type) {
	case *objects.Int:
		return v.Value
	case *objects.Bool:
		if v.Value {
			return 1
		}
		return 0
	case *objects.Null:
		return 0
	default:
		return globalObjectRegistry.Register(obj)
	}
}
