// pkg/compiler/options.go
// Configuration options for compiler optimizations
package compiler

// OptimizationFlags controls which optimizations are enabled
type OptimizationFlags struct {
	BytecodeOptimizer   bool
	InlineFunctions     bool // Function inlining at call sites
	InlineCache         bool
	ClosurePool         bool
	TypeSpecialization  bool
	Superinstructions   bool // Combine common instruction sequences
	StrengthReduction   bool // Replace expensive ops with cheaper ones
	DeadCodeElimination bool // Remove unreachable code
}

// DefaultOptimizations returns flags with all optimizations enabled
func DefaultOptimizations() OptimizationFlags {
	return OptimizationFlags{
		BytecodeOptimizer:   true,
		InlineFunctions:     true,
		InlineCache:         true,
		ClosurePool:         true,
		TypeSpecialization:  true,
		Superinstructions:   true,
		StrengthReduction:   true,
		DeadCodeElimination: true,
	}
}

// NoInliningOptimizations returns flags with inlining disabled (for benchmarking)
func NoInliningOptimizations() OptimizationFlags {
	return OptimizationFlags{
		BytecodeOptimizer:   true,
		InlineFunctions:     false,
		InlineCache:         true,
		ClosurePool:         true,
		TypeSpecialization:  true,
		Superinstructions:   true,
		StrengthReduction:   true,
		DeadCodeElimination: true,
	}
}
