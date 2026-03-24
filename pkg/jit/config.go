// pkg/jit/config.go
// JIT configuration types - available on all platforms

package jit

// JITConfig holds configuration for the JIT compiler
type JITConfig struct {
	// Minimum execution count before JIT compilation
	HotThreshold int
	// Maximum code size to JIT compile (in bytes)
	MaxCodeSize int
	// Enable debug output
	Debug bool
}

// JITStats tracks JIT compilation statistics
type JITStats struct {
	CompiledFunctions int64
	CacheHits         int64
	CacheMisses       int64
	TotalCodeSize     int64
}

// DefaultJITConfig returns default JIT configuration
func DefaultJITConfig() JITConfig {
	return JITConfig{
		HotThreshold: 100,
		MaxCodeSize:  4096,
		Debug:        false,
	}
}
