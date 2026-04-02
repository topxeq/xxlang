// pkg/objects/array.go
package objects

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// Array represents an array value
type Array struct {
	Elements   []Object
	LastPopped Object // Used by pop() to return the popped value
}

// Pre-cached empty array
var EMPTY_ARRAY = &Array{Elements: emptyElements}

var emptyElements = make([]Object, 0)

// arrayPool is a sync.Pool for array objects
var arrayPool = sync.Pool{
	New: func() interface{} {
		atomic.AddInt64(&arrayPoolStats.Created, 1)
		return &Array{}
	},
}

// ArrayPoolStats tracks statistics about array pool usage
type ArrayPoolStats struct {
	CacheHits int64
	PoolHits  int64
	Created   int64
	Released  int64
}

// arrayPoolStats tracks pool statistics
var arrayPoolStats struct {
	CacheHits int64
	PoolHits  int64
	Created   int64
	Released  int64
}

// NewArray creates a new Array object with the given elements
func NewArray(elements []Object) *Array {
	if len(elements) == 0 {
		atomic.AddInt64(&arrayPoolStats.CacheHits, 1)
		return EMPTY_ARRAY
	}
	obj := arrayPool.Get().(*Array)
	obj.Elements = elements
	obj.LastPopped = nil
	return obj
}

// NewArrayWithCapacity creates a new Array with pre-allocated capacity
func NewArrayWithCapacity(capacity int) *Array {
	if capacity == 0 {
		atomic.AddInt64(&arrayPoolStats.CacheHits, 1)
		return EMPTY_ARRAY
	}
	obj := arrayPool.Get().(*Array)
	if cap(obj.Elements) < capacity {
		obj.Elements = make([]Object, 0, capacity)
	} else {
		obj.Elements = obj.Elements[:0]
	}
	obj.LastPopped = nil
	return obj
}

// ReleaseArray returns an Array object to the pool
func ReleaseArray(obj *Array) {
	if obj == nil || obj == EMPTY_ARRAY {
		return
	}
	// Clear references to allow GC
	for i := range obj.Elements {
		obj.Elements[i] = nil
	}
	obj.Elements = obj.Elements[:0]
	obj.LastPopped = nil
	atomic.AddInt64(&arrayPoolStats.Released, 1)
	arrayPool.Put(obj)
}

// ReleaseArraySlice returns multiple Array objects to the pool
func ReleaseArraySlice(objs []*Array) {
	for _, obj := range objs {
		if obj == nil || obj == EMPTY_ARRAY {
			continue
		}
		for i := range obj.Elements {
			obj.Elements[i] = nil
		}
		obj.Elements = obj.Elements[:0]
		obj.LastPopped = nil
		arrayPool.Put(obj)
	}
	if len(objs) > 0 {
		atomic.AddInt64(&arrayPoolStats.Released, int64(len(objs)))
	}
}

// GetArrayPoolStats returns current statistics about array pool usage
func GetArrayPoolStats() ArrayPoolStats {
	return ArrayPoolStats{
		CacheHits: atomic.LoadInt64(&arrayPoolStats.CacheHits),
		PoolHits:  atomic.LoadInt64(&arrayPoolStats.PoolHits),
		Created:   atomic.LoadInt64(&arrayPoolStats.Created),
		Released:  atomic.LoadInt64(&arrayPoolStats.Released),
	}
}

// ResetArrayPoolStats resets the pool statistics counters
func ResetArrayPoolStats() {
	atomic.StoreInt64(&arrayPoolStats.CacheHits, 0)
	atomic.StoreInt64(&arrayPoolStats.PoolHits, 0)
	atomic.StoreInt64(&arrayPoolStats.Created, 0)
	atomic.StoreInt64(&arrayPoolStats.Released, 0)
}

// WarmArrayPool pre-allocates Array objects into the pool
func WarmArrayPool(count int) {
	for i := 0; i < count; i++ {
		arrayPool.Put(&Array{Elements: make([]Object, 0, 8)})
	}
}

// Type returns the object type
func (a *Array) Type() ObjectType { return ArrayType }

// TypeTag returns the type tag for fast type checking
func (a *Array) TypeTag() TypeTag { return TagArray }

// Inspect returns the string representation
func (a *Array) Inspect() string {
	visited := make(map[interface{}]struct{})
	return a.inspect(visited)
}

// inspect with cycle detection
func (a *Array) inspect(visited map[interface{}]struct{}) string {
	if _, ok := visited[a]; ok {
		return "[...]"
	}
	visited[a] = struct{}{}

	var out bytes.Buffer
	out.WriteString("[")
	for i, e := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		// Use cycle-aware inspection for elements
		switch v := e.(type) {
		case *Map:
			out.WriteString(v.inspect(visited))
		case *Array:
			out.WriteString(v.inspect(visited))
		case *OrderedMap:
			out.WriteString(v.inspect(visited))
		default:
			out.WriteString(e.Inspect())
		}
	}
	out.WriteString("]")
	return out.String()
}

// ToBool converts the array to a boolean
func (a *Array) ToBool() *Bool {
	if len(a.Elements) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (a *Array) HashKey() HashKey {
	// Arrays are not hashable
	return HashKey{Type: ArrayType, Value: 0}
}
