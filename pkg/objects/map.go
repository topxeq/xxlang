// pkg/objects/map.go
package objects

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"
)

// MapPair represents a key-value pair in a map
type MapPair struct {
	Key   Object
	Value Object
}

// Map represents a map value
type Map struct {
	Pairs       map[HashKey]MapPair
	sortedKeys  []Object // Cached sorted keys for iteration
	keysInvalid bool     // True if sortedKeys needs recomputation
}

// Pre-cached empty map no longer used; empty maps are created fresh.
// var EMPTY_MAP = &Map{Pairs: emptyPairs}

var emptyPairs = make(map[HashKey]MapPair)

// mapPool is a sync.Pool for map objects
var mapPool = sync.Pool{
	New: func() interface{} {
		atomic.AddInt64(&mapPoolStats.Created, 1)
		return &Map{}
	},
}

// MapPoolStats tracks statistics about map pool usage
type MapPoolStats struct {
	CacheHits int64
	PoolHits  int64
	Created   int64
	Released  int64
}

// mapPoolStats tracks pool statistics
var mapPoolStats struct {
	CacheHits int64
	PoolHits  int64
	Created   int64
	Released  int64
}

// NewMap creates a new Map object with the given pairs
func NewMap(pairs map[HashKey]MapPair) *Map {
	if len(pairs) == 0 {
		// Return a fresh empty map instead of shared singleton to avoid mutation issues.
		return &Map{Pairs: make(map[HashKey]MapPair)}
	}
	atomic.AddInt64(&mapPoolStats.CacheHits, 1)
	obj := mapPool.Get().(*Map)
	obj.Pairs = pairs
	obj.sortedKeys = nil
	obj.keysInvalid = true
	return obj
}

// NewMapWithCapacity creates a new Map with pre-allocated capacity
func NewMapWithCapacity(capacity int) *Map {
	if capacity == 0 {
		// Return a fresh empty map
		return &Map{Pairs: make(map[HashKey]MapPair)}
	}
	atomic.AddInt64(&mapPoolStats.CacheHits, 1)
	obj := mapPool.Get().(*Map)
	if obj.Pairs == nil {
		obj.Pairs = make(map[HashKey]MapPair, capacity)
	} else {
		// Clear existing pairs
		for k := range obj.Pairs {
			delete(obj.Pairs, k)
		}
	}
	obj.sortedKeys = nil
	obj.keysInvalid = true
	return obj
}

// ReleaseMap returns a Map object to the pool
func ReleaseMap(obj *Map) {
	if obj == nil {
		return
	}
	// Clear references to allow GC
	for k := range obj.Pairs {
		delete(obj.Pairs, k)
	}
	obj.sortedKeys = nil
	obj.keysInvalid = true
	atomic.AddInt64(&mapPoolStats.Released, 1)
	mapPool.Put(obj)
}

// ReleaseMapSlice returns multiple Map objects to the pool
func ReleaseMapSlice(objs []*Map) {
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		for k := range obj.Pairs {
			delete(obj.Pairs, k)
		}
		obj.sortedKeys = nil
		obj.keysInvalid = true
		mapPool.Put(obj)
	}
	if len(objs) > 0 {
		atomic.AddInt64(&mapPoolStats.Released, int64(len(objs)))
	}
}

// GetMapPoolStats returns current statistics about map pool usage
func GetMapPoolStats() MapPoolStats {
	return MapPoolStats{
		CacheHits: atomic.LoadInt64(&mapPoolStats.CacheHits),
		PoolHits:  atomic.LoadInt64(&mapPoolStats.PoolHits),
		Created:   atomic.LoadInt64(&mapPoolStats.Created),
		Released:  atomic.LoadInt64(&mapPoolStats.Released),
	}
}

// ResetMapPoolStats resets the pool statistics counters
func ResetMapPoolStats() {
	atomic.StoreInt64(&mapPoolStats.CacheHits, 0)
	atomic.StoreInt64(&mapPoolStats.PoolHits, 0)
	atomic.StoreInt64(&mapPoolStats.Created, 0)
	atomic.StoreInt64(&mapPoolStats.Released, 0)
}

// WarmMapPool pre-allocates Map objects into the pool
func WarmMapPool(count int) {
	for i := 0; i < count; i++ {
		mapPool.Put(&Map{Pairs: make(map[HashKey]MapPair, 8)})
	}
}

// Type returns the object type
func (m *Map) Type() ObjectType { return MapType }

// TypeTag returns the type tag for fast type checking
func (m *Map) TypeTag() TypeTag { return TagMap }

// Inspect returns the string representation
func (m *Map) Inspect() string {
	visited := make(map[interface{}]struct{})
	return m.inspect(visited)
}

// inspect with cycle detection
func (m *Map) inspect(visited map[interface{}]struct{}) string {
	// Detect cycle
	if _, ok := visited[m]; ok {
		return "{...}"
	}
	visited[m] = struct{}{}

	var out bytes.Buffer
	out.WriteString("{")

	// Collect and sort keys for deterministic output
	keys := make([]string, 0, len(m.Pairs))
	pairByKey := make(map[string]MapPair, len(m.Pairs))
	for _, pair := range m.Pairs {
		keyStr := pair.Key.Inspect()
		keys = append(keys, keyStr)
		pairByKey[keyStr] = pair
	}
	sort.Strings(keys)

	first := true
	for _, keyStr := range keys {
		if !first {
			out.WriteString(", ")
		}
		first = false
		pair := pairByKey[keyStr]
		out.WriteString(pair.Key.Inspect())
		out.WriteString(": ")
		// Use cycle-aware inspection for values
		switch v := pair.Value.(type) {
		case *Map:
			out.WriteString(v.inspect(visited))
		case *Array:
			out.WriteString(v.inspect(visited))
		case *OrderedMap:
			out.WriteString(v.inspect(visited))
		default:
			out.WriteString(pair.Value.Inspect())
		}
	}
	out.WriteString("}")
	return out.String()
}

// ToBool converts the map to a boolean
func (m *Map) ToBool() *Bool {
	if len(m.Pairs) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (m *Map) HashKey() HashKey {
	// Maps are not hashable
	return HashKey{Type: MapType, Value: 0}
}

// GetSortedKeys returns the keys in sorted order, caching the result
func (m *Map) GetSortedKeys() []Object {
	// Return cached keys if valid
	if !m.keysInvalid && m.sortedKeys != nil {
		return m.sortedKeys
	}

	// Build sorted keys
	keys := make([]Object, 0, len(m.Pairs))
	for _, pair := range m.Pairs {
		keys = append(keys, pair.Key)
	}

	// Sort keys by string representation for deterministic order
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Inspect() < keys[j].Inspect()
	})

	// Cache the result
	m.sortedKeys = keys
	m.keysInvalid = false
	return keys
}

// InvalidateKeysCache marks the sorted keys cache as invalid
// Call this when the map is modified
func (m *Map) InvalidateKeysCache() {
	m.keysInvalid = true
}
