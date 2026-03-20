// pkg/jit/jit_vm.go
// JIT-enabled VM that combines interpreter with JIT compilation
package jit

import (
	"fmt"
	"sync"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITVM wraps a register VM with JIT compilation capability
type JITVM struct {
	*vm.RegVM
	jit      *JITCompiler
	config   JITConfig
	enabled  bool
	compLock sync.Mutex
}

// NewJITVM creates a new JIT-enabled VM
func NewJITVM(bytecode *compiler.Bytecode, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:  vm.NewRegVM(bytecode),
		jit:    NewJITCompiler(config),
		config: config,
		enabled: true,
	}
}

// NewJITVMWithGlobals creates a JIT VM with custom globals
func NewJITVMWithGlobals(bytecode *compiler.Bytecode, globals []vm.Value, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:  vm.NewRegVMWithGlobals(bytecode, globals),
		jit:    NewJITCompiler(config),
		config: config,
		enabled: true,
	}
}

// Run executes the bytecode with JIT compilation for hot paths
func (j *JITVM) Run() error {
	if !j.enabled {
		return j.RegVM.Run()
	}

	// For now, use interpreter with JIT compilation tracking
	// Full JIT integration requires more complex call handling
	return j.RegVM.Run()
}

// RunJIT attempts to run a function with JIT compilation
func (j *JITVM) RunJIT(fn *compiler.CompiledFunction) (int64, error) {
	j.compLock.Lock()
	defer j.compLock.Unlock()

	// Check if already compiled
	cf := j.jit.GetCompiled(fn)
	if cf != nil {
		// Execute compiled code
		return cf.Execute(), nil
	}

	// Check if should compile
	if !j.jit.ShouldCompile(fn) {
		return 0, fmt.Errorf("function not hot enough for JIT")
	}

	// Compile the function
	cf, err := j.jit.Compile(fn, j.GetConstants(), j.GetGlobals())
	if err != nil {
		return 0, fmt.Errorf("JIT compilation failed: %w", err)
	}

	// Execute compiled code
	return cf.Execute(), nil
}

// GetConstants returns the VM's constants
func (j *JITVM) GetConstants() []vm.Value {
	return j.RegVM.GetConstants()
}

// GetGlobals returns the VM's globals
func (j *JITVM) GetGlobals() []vm.Value {
	return j.RegVM.GetGlobals()
}

// SetJITEnabled enables or disables JIT compilation
func (j *JITVM) SetJITEnabled(enabled bool) {
	j.enabled = enabled
}

// GetJITStats returns JIT compilation statistics
func (j *JITVM) GetJITStats() JITStats {
	return j.jit.GetStats()
}

// Cleanup releases JIT resources
func (j *JITVM) Cleanup() {
	j.jit.Cleanup()
}

// CompileFunction compiles a specific function for JIT execution
func (j *JITVM) CompileFunction(fn *compiler.CompiledFunction) (*CompiledFunc, error) {
	return j.jit.Compile(fn, j.GetConstants(), j.GetGlobals())
}

// ExecuteCompiled executes a previously compiled function
func (j *JITVM) ExecuteCompiled(cf *CompiledFunc) int64 {
	return cf.Execute()
}
