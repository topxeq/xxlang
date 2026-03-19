// pkg/objects/float.go
package objects

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"sync/atomic"
)

// Float represents a floating-point value
type Float struct {
	Value float64
}

// Pre-cached common float values
// Unlike integers, we can only cache a small set due to the infinite nature of floats
var (
	FLOAT_ZERO  = &Float{Value: 0.0}
	FLOAT_ONE   = &Float{Value: 1.0}
	FLOAT_NEG_ONE = &Float{Value: -1.0}
	FLOAT_TWO   = &Float{Value: 2.0}
	FLOAT_HALF  = &Float{Value: 0.5}
	FLOAT_TEN   = &Float{Value: 10.0}
	FLOAT_HUNDRED = &Float{Value: 100.0}
	FLOAT_THOUSAND = &Float{Value: 1000.0}
	FLOAT_PI    = &Float{Value: 3.14159265358979323846}
	FLOAT_E     = &Float{Value: 2.71828182845904523536}
)

// floatPool is a sync.Pool for float objects
// This helps reduce allocations for float values that are frequently created and discarded
var floatPool = sync.Pool{
	New: func() interface{} {
		atomic.AddInt64(&floatPoolStats.Created, 1)
		return &Float{}
	},
}

// FloatPoolStats tracks statistics about float pool usage
type FloatPoolStats struct {
	// CacheHits is the number of times a cached float was returned
	CacheHits int64
	// PoolHits is the number of times a pooled float was reused
	PoolHits int64
	// Created is the number of new floats allocated by the pool
	Created int64
	// Released is the number of floats returned to the pool
	Released int64
}

// floatPoolStats tracks pool statistics
var floatPoolStats struct {
	CacheHits int64
	PoolHits  int64
	Created   int64
	Released  int64
}

// NewFloat creates a new Float object, using cached values for common floats
func NewFloat(val float64) *Float {
	// Check for cached common values
	switch val {
	case 0.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_ZERO
	case 1.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_ONE
	case -1.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_NEG_ONE
	case 2.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_TWO
	case 0.5:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_HALF
	case 10.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_TEN
	case 100.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_HUNDRED
	case 1000.0:
		atomic.AddInt64(&floatPoolStats.CacheHits, 1)
		return FLOAT_THOUSAND
	}

	// For other values, use the pool
	obj := floatPool.Get().(*Float)
	obj.Value = val
	return obj
}

// ReleaseFloat returns a Float object to the pool if it's not a cached value
// This should be called when the object is no longer needed
func ReleaseFloat(obj *Float) {
	// Don't pool cached values
	switch obj.Value {
	case 0.0, 1.0, -1.0, 2.0, 0.5, 10.0, 100.0, 1000.0:
		return
	}
	atomic.AddInt64(&floatPoolStats.Released, 1)
	floatPool.Put(obj)
}

// ReleaseFloatSlice returns multiple Float objects to the pool
// This is more efficient than calling ReleaseFloat multiple times
func ReleaseFloatSlice(objs []*Float) {
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		switch obj.Value {
		case 0.0, 1.0, -1.0, 2.0, 0.5, 10.0, 100.0, 1000.0:
			continue
		}
		floatPool.Put(obj)
	}
	atomic.AddInt64(&floatPoolStats.Released, int64(len(objs)))
}

// GetFloatPoolStats returns current statistics about float pool usage
func GetFloatPoolStats() FloatPoolStats {
	return FloatPoolStats{
		CacheHits: atomic.LoadInt64(&floatPoolStats.CacheHits),
		PoolHits:  atomic.LoadInt64(&floatPoolStats.PoolHits),
		Created:   atomic.LoadInt64(&floatPoolStats.Created),
		Released:  atomic.LoadInt64(&floatPoolStats.Released),
	}
}

// ResetFloatPoolStats resets the pool statistics counters
func ResetFloatPoolStats() {
	atomic.StoreInt64(&floatPoolStats.CacheHits, 0)
	atomic.StoreInt64(&floatPoolStats.PoolHits, 0)
	atomic.StoreInt64(&floatPoolStats.Created, 0)
	atomic.StoreInt64(&floatPoolStats.Released, 0)
}

// NewFloatSlice creates multiple Float objects efficiently
// This is optimized for batch operations
func NewFloatSlice(values []float64) []*Float {
	result := make([]*Float, len(values))
	for i, v := range values {
		result[i] = NewFloat(v)
	}
	return result
}

// WarmFloatPool pre-allocates a number of Float objects into the pool
// This can improve performance by reducing allocations during hot paths
func WarmFloatPool(count int) {
	for i := 0; i < count; i++ {
		floatPool.Put(&Float{})
	}
}

// floatPoolBuffer is a thread-local buffer for temporary float objects
// This reduces contention on the global pool for short-lived operations
type floatPoolBuffer struct {
	buf []*Float
	mu  sync.Mutex
}

const floatPoolBufferSize = 64

// globalFloatBuffer is a shared buffer for temporary float allocations
var globalFloatBuffer = &floatPoolBuffer{
	buf: make([]*Float, 0, floatPoolBufferSize),
}

// GetTempFloat gets a temporary Float from the buffer or pool
func (b *floatPoolBuffer) GetTempFloat(val float64) *Float {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Try to get from buffer first
	if len(b.buf) > 0 {
		obj := b.buf[len(b.buf)-1]
		b.buf = b.buf[:len(b.buf)-1]
		obj.Value = val
		return obj
	}

	// Fall back to global pool
	obj := floatPool.Get().(*Float)
	obj.Value = val
	return obj
}

// PutTempFloat returns a temporary Float to the buffer or pool
func (b *floatPoolBuffer) PutTempFloat(obj *Float) {
	// Don't pool cached values
	switch obj.Value {
	case 0.0, 1.0, -1.0, 2.0, 0.5, 10.0, 100.0, 1000.0:
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) < floatPoolBufferSize {
		b.buf = append(b.buf, obj)
	} else {
		// Buffer full, return to global pool
		floatPool.Put(obj)
	}
}

// GetBufferedFloat gets a temporary Float using the global buffer
func GetBufferedFloat(val float64) *Float {
	return globalFloatBuffer.GetTempFloat(val)
}

// PutBufferedFloat returns a temporary Float to the global buffer
func PutBufferedFloat(obj *Float) {
	globalFloatBuffer.PutTempFloat(obj)
}

// Type returns the object type
func (f *Float) Type() ObjectType { return FloatType }

// TypeTag returns the type tag for fast type checking
func (f *Float) TypeTag() TypeTag { return TagFloat }

// Inspect returns the string representation
func (f *Float) Inspect() string { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

// ToBool converts the float to a boolean
func (f *Float) ToBool() *Bool {
	if f.Value == 0.0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%f", f.Value)))
	return HashKey{Type: FloatType, Value: h.Sum64()}
}
