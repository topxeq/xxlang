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

// DefaultJITConfig returns default JIT configuration
func DefaultJITConfig() JITConfig {
	return JITConfig{
		HotThreshold: 100,
		MaxCodeSize:  4096,
		Debug:        false,
	}
}
