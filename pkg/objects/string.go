// pkg/objects/string.go
package objects

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
)

// String represents a string value
type String struct {
	Value string
}

// Pre-cached common string values
var (
	STRING_EMPTY = &String{Value: ""}
	STRING_TRUE  = &String{Value: "true"}
	STRING_FALSE = &String{Value: "false"}
	STRING_NULL  = &String{Value: "null"}
	STRING_INT   = &String{Value: "INT"}
	STRING_FLOAT = &String{Value: "FLOAT"}
	STRING_BOOL  = &String{Value: "BOOL"}
	STRING_STRING = &String{Value: "STRING"}
	STRING_ARRAY  = &String{Value: "ARRAY"}
	STRING_MAP    = &String{Value: "MAP"}
	STRING_FUNC   = &String{Value: "FUNCTION"}
	STRING_BUILTIN = &String{Value: "BUILTIN"}
	STRING_ERROR  = &String{Value: "ERROR"}
	STRING_NIL    = &String{Value: "nil"}
	STRING_ZERO   = &String{Value: "0"}
	STRING_ONE    = &String{Value: "1"}
	STRING_SPACE  = &String{Value: " "}
	STRING_NEWLINE = &String{Value: "\n"}
)

// stringPool is a sync.Pool for string objects
// This helps reduce allocations for strings that are frequently created and discarded
var stringPool = sync.Pool{
	New: func() interface{} {
		atomic.AddInt64(&stringPoolStats.Created, 1)
		return &String{}
	},
}

// StringPoolStats tracks statistics about string pool usage
type StringPoolStats struct {
	// CacheHits is the number of times a cached string was returned
	CacheHits int64
	// InternHits is the number of times an interned string was returned
	InternHits int64
	// PoolHits is the number of times a pooled string was reused
	PoolHits int64
	// Created is the number of new strings allocated by the pool
	Created int64
	// Released is the number of strings returned to the pool
	Released int64
	// Interned is the number of strings currently in the intern cache
	Interned int64
}

// stringPoolStats tracks pool statistics
var stringPoolStats struct {
	CacheHits  int64
	InternHits int64
	PoolHits   int64
	Created    int64
	Released   int64
}

// StringIntern is a string interning pool for frequently used strings
// This provides permanent caching for strings that are used often
var stringIntern sync.Map // map[string]*String

// internCount tracks the number of interned strings
var internCount int64

// NewString creates a new String object, using cached values for common strings
func NewString(val string) *String {
	// Check for pre-cached common values first
	switch val {
	case "":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_EMPTY
	case "true":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_TRUE
	case "false":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_FALSE
	case "null":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_NULL
	case "INT":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_INT
	case "FLOAT":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_FLOAT
	case "BOOL":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_BOOL
	case "STRING":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_STRING
	case "ARRAY":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_ARRAY
	case "MAP":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_MAP
	case "FUNCTION":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_FUNC
	case "BUILTIN":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_BUILTIN
	case "ERROR":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_ERROR
	case "nil":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_NIL
	case "0":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_ZERO
	case "1":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_ONE
	case " ":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_SPACE
	case "\n":
		atomic.AddInt64(&stringPoolStats.CacheHits, 1)
		return STRING_NEWLINE
	}

	// For other values, use the pool
	obj := stringPool.Get().(*String)
	obj.Value = val
	return obj
}

// ReleaseString returns a String object to the pool if it's not a cached value
// This should be called when the object is no longer needed
func ReleaseString(obj *String) {
	// Don't pool pre-cached values
	switch obj.Value {
	case "", "true", "false", "null", "INT", "FLOAT", "BOOL", "STRING",
		"ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n":
		return
	}
	atomic.AddInt64(&stringPoolStats.Released, 1)
	stringPool.Put(obj)
}

// ReleaseStringSlice returns multiple String objects to the pool
// This is more efficient than calling ReleaseString multiple times
func ReleaseStringSlice(objs []*String) {
	count := 0
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		switch obj.Value {
		case "", "true", "false", "null", "INT", "FLOAT", "BOOL", "STRING",
			"ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n":
			continue
		}
		stringPool.Put(obj)
		count++
	}
	if count > 0 {
		atomic.AddInt64(&stringPoolStats.Released, int64(count))
	}
}

// GetStringPoolStats returns current statistics about string pool usage
func GetStringPoolStats() StringPoolStats {
	return StringPoolStats{
		CacheHits:  atomic.LoadInt64(&stringPoolStats.CacheHits),
		InternHits: atomic.LoadInt64(&stringPoolStats.InternHits),
		PoolHits:   atomic.LoadInt64(&stringPoolStats.PoolHits),
		Created:    atomic.LoadInt64(&stringPoolStats.Created),
		Released:   atomic.LoadInt64(&stringPoolStats.Released),
		Interned:   atomic.LoadInt64(&internCount),
	}
}

// ResetStringPoolStats resets the pool statistics counters
func ResetStringPoolStats() {
	atomic.StoreInt64(&stringPoolStats.CacheHits, 0)
	atomic.StoreInt64(&stringPoolStats.InternHits, 0)
	atomic.StoreInt64(&stringPoolStats.PoolHits, 0)
	atomic.StoreInt64(&stringPoolStats.Created, 0)
	atomic.StoreInt64(&stringPoolStats.Released, 0)
}

// NewStringSlice creates multiple String objects efficiently
// This is optimized for batch operations
func NewStringSlice(values []string) []*String {
	result := make([]*String, len(values))
	for i, v := range values {
		result[i] = NewString(v)
	}
	return result
}

// WarmStringPool pre-allocates a number of String objects into the pool
// This can improve performance by reducing allocations during hot paths
func WarmStringPool(count int) {
	for i := 0; i < count; i++ {
		stringPool.Put(&String{})
	}
}

// stringPoolBuffer is a thread-local buffer for temporary string objects
// This reduces contention on the global pool for short-lived operations
type stringPoolBuffer struct {
	buf []*String
	mu  sync.Mutex
}

const stringPoolBufferSize = 64

// globalStringBuffer is a shared buffer for temporary string allocations
var globalStringBuffer = &stringPoolBuffer{
	buf: make([]*String, 0, stringPoolBufferSize),
}

// GetTempString gets a temporary String from the buffer or pool
func (b *stringPoolBuffer) GetTempString(val string) *String {
	// Check for pre-cached values
	switch val {
	case "", "true", "false", "null", "INT", "FLOAT", "BOOL", "STRING",
		"ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n":
		return NewString(val)
	}

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
	obj := stringPool.Get().(*String)
	obj.Value = val
	return obj
}

// PutTempString returns a temporary String to the buffer or pool
func (b *stringPoolBuffer) PutTempString(obj *String) {
	// Don't pool pre-cached values
	switch obj.Value {
	case "", "true", "false", "null", "INT", "FLOAT", "BOOL", "STRING",
		"ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n":
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) < stringPoolBufferSize {
		b.buf = append(b.buf, obj)
	} else {
		// Buffer full, return to global pool
		stringPool.Put(obj)
	}
}

// GetBufferedString gets a temporary String using the global buffer
func GetBufferedString(val string) *String {
	return globalStringBuffer.GetTempString(val)
}

// PutBufferedString returns a temporary String to the global buffer
func PutBufferedString(obj *String) {
	globalStringBuffer.PutTempString(obj)
}

// InternString returns a cached *String for the given value
// This provides permanent caching for strings that are used often
// Use this for strings that will be reused many times (e.g., identifiers, keywords)
func InternString(val string) *String {
	// Check cache first
	if cached, ok := stringIntern.Load(val); ok {
		atomic.AddInt64(&stringPoolStats.InternHits, 1)
		return cached.(*String)
	}

	// Check for pre-cached common values
	switch val {
	case "":
		return STRING_EMPTY
	case "true":
		return STRING_TRUE
	case "false":
		return STRING_FALSE
	case "null":
		return STRING_NULL
	}

	// Create new string object
	s := &String{Value: val}

	// Store in cache (loadOrStore to handle race condition)
	if actual, loaded := stringIntern.LoadOrStore(val, s); loaded {
		return actual.(*String)
	}
	atomic.AddInt64(&internCount, 1)
	return s
}

// InternBatch pre-interns a batch of strings
func InternBatch(strings []string) []*String {
	result := make([]*String, len(strings))
	for i, s := range strings {
		result[i] = InternString(s)
	}
	return result
}

// ClearInternCache clears the string intern cache
// Use with caution - this will cause all interned strings to be re-allocated
func ClearInternCache() {
	stringIntern = sync.Map{}
	atomic.StoreInt64(&internCount, 0)
}

// Type returns the object type
func (s *String) Type() ObjectType { return StringType }

// TypeTag returns the type tag for fast type checking
func (s *String) TypeTag() TypeTag { return TagString }

// Inspect returns the string representation
func (s *String) Inspect() string { return s.Value }

// ToBool converts the string to a boolean
func (s *String) ToBool() *Bool {
	if len(s.Value) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: StringType, Value: h.Sum64()}
}
