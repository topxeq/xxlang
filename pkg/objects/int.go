// pkg/objects/int.go
package objects

import (
	"strconv"
	"sync"
	"sync/atomic"
)

// Int represents an integer value
type Int struct {
	Value int64
}

// IntCacheMin is the minimum cached integer value
const IntCacheMin = -1000

// IntCacheMax is the maximum cached integer value
// Extended to cover common loop indices and calculation results
const IntCacheMax = 100000

// intCacheSize is the number of pre-allocated integers
const intCacheSize = IntCacheMax - IntCacheMin + 1

// intCache holds pre-allocated integers for common values
var intCache [intCacheSize]*Int

// initIntCache pre-allocates common integer values
func initIntCache() {
	for i := IntCacheMin; i <= IntCacheMax; i++ {
		intCache[i-IntCacheMin] = &Int{Value: int64(i)}
	}
}

func init() {
	initIntCache()
}

// intPool is a sync.Pool for integers outside the cache range
// This helps reduce allocations for large integers that are frequently created and discarded
var intPool = sync.Pool{
	New: func() interface{} {
		atomic.AddInt64(&intPoolStats.Created, 1)
		return &Int{}
	},
}

// IntPoolStats tracks statistics about integer pool usage
type IntPoolStats struct {
	// CacheHits is the number of times a cached integer was returned
	CacheHits int64
	// PoolHits is the number of times a pooled integer was reused
	PoolHits int64
	// PoolMisses is the number of times a new integer had to be allocated
	PoolMisses int64
	// Created is the number of new integers allocated by the pool
	Created int64
	// Released is the number of integers returned to the pool
	Released int64
}

// intPoolStats tracks pool statistics
var intPoolStats struct {
	CacheHits  int64
	PoolHits   int64
	PoolMisses int64
	Created    int64
	Released   int64
}

// NewInt creates a new Int object, using cache for small values
func NewInt(val int64) *Int {
	if val >= IntCacheMin && val <= IntCacheMax {
		atomic.AddInt64(&intPoolStats.CacheHits, 1)
		return intCache[val-IntCacheMin]
	}
	// For values outside cache range, use the pool
	obj := intPool.Get().(*Int)
	obj.Value = val
	// Track if this was a pool hit (reused object) or miss (new allocation)
	// Note: We can't distinguish hits from misses directly, but Created counter
	// in the pool's New function tracks new allocations
	return obj
}

// ReleaseInt returns an Int object to the pool if it's outside the cache range
// This should be called when the object is no longer needed
func ReleaseInt(obj *Int) {
	// Only pool objects outside the cache range
	if obj.Value < IntCacheMin || obj.Value > IntCacheMax {
		atomic.AddInt64(&intPoolStats.Released, 1)
		intPool.Put(obj)
	}
}

// ReleaseIntSlice returns multiple Int objects to the pool
// This is more efficient than calling ReleaseInt multiple times
func ReleaseIntSlice(objs []*Int) {
	for _, obj := range objs {
		if obj != nil && (obj.Value < IntCacheMin || obj.Value > IntCacheMax) {
			intPool.Put(obj)
		}
	}
	atomic.AddInt64(&intPoolStats.Released, int64(len(objs)))
}

// GetIntPoolStats returns current statistics about integer pool usage
func GetIntPoolStats() IntPoolStats {
	return IntPoolStats{
		CacheHits:  atomic.LoadInt64(&intPoolStats.CacheHits),
		PoolHits:   atomic.LoadInt64(&intPoolStats.PoolHits),
		PoolMisses: atomic.LoadInt64(&intPoolStats.PoolMisses),
		Created:    atomic.LoadInt64(&intPoolStats.Created),
		Released:   atomic.LoadInt64(&intPoolStats.Released),
	}
}

// ResetIntPoolStats resets the pool statistics counters
func ResetIntPoolStats() {
	atomic.StoreInt64(&intPoolStats.CacheHits, 0)
	atomic.StoreInt64(&intPoolStats.PoolHits, 0)
	atomic.StoreInt64(&intPoolStats.PoolMisses, 0)
	atomic.StoreInt64(&intPoolStats.Created, 0)
	atomic.StoreInt64(&intPoolStats.Released, 0)
}

// IsCachedInt returns true if the given value is within the cached range
func IsCachedInt(val int64) bool {
	return val >= IntCacheMin && val <= IntCacheMax
}

// NewIntSlice creates multiple Int objects efficiently
// This is optimized for batch operations
func NewIntSlice(values []int64) []*Int {
	result := make([]*Int, len(values))
	for i, v := range values {
		if v >= IntCacheMin && v <= IntCacheMax {
			result[i] = intCache[v-IntCacheMin]
		} else {
			obj := intPool.Get().(*Int)
			obj.Value = v
			result[i] = obj
		}
	}
	return result
}

// WarmIntPool pre-allocates a number of Int objects into the pool
// This can improve performance by reducing allocations during hot paths
func WarmIntPool(count int) {
	for i := 0; i < count; i++ {
		intPool.Put(&Int{})
	}
}

// intPoolBuffer is a thread-local buffer for temporary integer objects
// This reduces contention on the global pool for short-lived operations
type intPoolBuffer struct {
	buf    []*Int
	idx    int
	mu     sync.Mutex
}

const intPoolBufferSize = 64

// globalIntBuffer is a shared buffer for temporary integer allocations
var globalIntBuffer = &intPoolBuffer{
	buf: make([]*Int, 0, intPoolBufferSize),
}

// GetTempInt gets a temporary Int from the buffer or pool
func (b *intPoolBuffer) GetTempInt(val int64) *Int {
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
	obj := intPool.Get().(*Int)
	obj.Value = val
	return obj
}

// PutTempInt returns a temporary Int to the buffer or pool
func (b *intPoolBuffer) PutTempInt(obj *Int) {
	if obj.Value >= IntCacheMin && obj.Value <= IntCacheMax {
		return // Don't pool cached values
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) < intPoolBufferSize {
		b.buf = append(b.buf, obj)
	} else {
		// Buffer full, return to global pool
		intPool.Put(obj)
	}
}

// GetBufferedInt gets a temporary Int using the global buffer
func GetBufferedInt(val int64) *Int {
	return globalIntBuffer.GetTempInt(val)
}

// PutBufferedInt returns a temporary Int to the global buffer
func PutBufferedInt(obj *Int) {
	globalIntBuffer.PutTempInt(obj)
}

// Type returns the object type
func (i *Int) Type() ObjectType { return IntType }

// TypeTag returns the type tag for fast type checking
func (i *Int) TypeTag() TypeTag { return TagInt }

// Inspect returns the string representation
func (i *Int) Inspect() string { return strconv.FormatInt(i.Value, 10) }

// ToBool converts the integer to a boolean
func (i *Int) ToBool() *Bool {
	if i.Value == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (i *Int) HashKey() HashKey {
	return HashKey{Type: IntType, Value: uint64(i.Value)}
}
