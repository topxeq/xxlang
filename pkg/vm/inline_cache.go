// pkg/vm/inline_cache.go
// Inline caching for property access and method calls
package vm

import (
	"unsafe"

	"github.com/topxeq/xxlang/pkg/objects"
)

// CacheResultType indicates what type of result is cached
type CacheResultType uint8

const (
	CacheResultNone         CacheResultType = iota // No valid cache entry
	CacheResultMethod                              // Cached method (builtin or function)
	CacheResultField                               // Cached field value
	CacheResultNull                                // Cached null (property not found)
	CacheResultPrimitiveMethod                     // Cached primitive type method
	CacheResultMapMethod                           // Cached map method
)

// InlineCache is a cache entry for property/method lookups
type InlineCache struct {
	// Key: type tag or class pointer for validation
	TypeTag     objects.TypeTag // For primitive types
	Class       *objects.Class  // For instances (pointer comparison)
	NameHash    uint32          // Hash of property/method name

	// Result
	ResultType CacheResultType // Type of cached result
	Method     objects.Object  // Cached method (if ResultType is Method or PrimitiveMethod)
	FieldIdx   int             // Cached field index (for instances)
}

// InlineCacheTable is a fixed-size inline cache
type InlineCacheTable struct {
	entries [CacheSize]InlineCache
	hits    int
	misses  int
}

// CacheSize is the number of cache entries
const CacheSize = 1024

// hashName creates a simple hash from a string
func hashName(s string) uint32 {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return h
}

// ComputeCacheIndex computes the cache index from type tag/class and name hash
func ComputeCacheIndex(typeTag objects.TypeTag, class *objects.Class, nameHash uint32) int {
	var key uint32
	if class != nil {
		// Use class pointer for instances
		key = uint32(uintptr(unsafe.Pointer(class))) ^ nameHash
	} else {
		// Use type tag for primitives
		key = uint32(typeTag) ^ nameHash
	}
	return int(key % CacheSize)
}

// Get looks up the cache entry
func (c *InlineCacheTable) Get(typeTag objects.TypeTag, class *objects.Class, nameHash uint32) *InlineCache {
	idx := ComputeCacheIndex(typeTag, class, nameHash)
	entry := &c.entries[idx]

	// Validate the entry matches
	if class != nil {
		if entry.Class != class || entry.NameHash != nameHash {
			c.misses++
			return nil
		}
	} else {
		if entry.TypeTag != typeTag || entry.NameHash != nameHash {
			c.misses++
			return nil
		}
	}

	// Check if entry is valid
	if entry.ResultType == CacheResultNone {
		c.misses++
		return nil
	}

	c.hits++
	return entry
}

// Set stores a cache entry
func (c *InlineCacheTable) Set(typeTag objects.TypeTag, class *objects.Class, nameHash uint32, resultType CacheResultType, method objects.Object, fieldIdx int) {
	idx := ComputeCacheIndex(typeTag, class, nameHash)
	entry := &c.entries[idx]

	entry.TypeTag = typeTag
	entry.Class = class
	entry.NameHash = nameHash
	entry.ResultType = resultType
	entry.Method = method
	entry.FieldIdx = fieldIdx
}

// Stats returns cache hit/miss statistics
func (c *InlineCacheTable) Stats() (hits, misses int) {
	return c.hits, c.misses
}

// Reset clears the cache
func (c *InlineCacheTable) Reset() {
	for i := range c.entries {
		c.entries[i] = InlineCache{}
	}
	c.hits = 0
	c.misses = 0
}
