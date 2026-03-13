// pkg/compiler/options.go
// Configuration options for compiler optimizations
package compiler

// OptimizationFlags controls which optimizations are enabled
type OptimizationFlags struct {
    BytecodeOptimizer bool
    InlineCache      bool
    ClosurePool      bool
    TypeSpecialization bool
}

// DefaultOptimizations returns flags with all optimizations enabled
func DefaultOptimizations() OptimizationFlags {
    return OptimizationFlags{
        BytecodeOptimizer:  true,
        InlineCache:       true,
        ClosurePool:       true,
        TypeSpecialization: true,
    }
}
