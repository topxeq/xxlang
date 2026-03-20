// pkg/jit/jit_vm.go
// JIT-enabled VM that combines interpreter with JIT compilation
package jit

import (
	"fmt"
	"sync"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITVM wraps a register VM with JIT compilation capability
type JITVM struct {
	*vm.RegVM
	jit        *JITCompiler
	nativeExec *NativeExecutor
	config     JITConfig
	enabled    bool
	compLock   sync.Mutex
	bytecode   *compiler.Bytecode

	// Statistics
	nativeExecs int64
	interpExecs int64
}

// NewJITVM creates a new JIT-enabled VM
func NewJITVM(bytecode *compiler.Bytecode, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:      vm.NewRegVM(bytecode),
		jit:        NewJITCompiler(config),
		nativeExec: NewNativeExecutor(config),
		config:     config,
		enabled:    true,
		bytecode:   bytecode,
	}
}

// NewJITVMWithGlobals creates a JIT VM with custom globals
func NewJITVMWithGlobals(bytecode *compiler.Bytecode, globals []vm.Value, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:      vm.NewRegVMWithGlobals(bytecode, globals),
		jit:        NewJITCompiler(config),
		nativeExec: NewNativeExecutor(config),
		config:     config,
		enabled:    true,
		bytecode:   bytecode,
	}
}

// Run executes the bytecode with JIT compilation for hot paths
func (j *JITVM) Run() error {
	if !j.enabled {
		return j.RegVM.Run()
	}

	// Check if we can execute the main code natively
	// This works for pure arithmetic/loop code without function calls
	mainFn := &compiler.CompiledFunction{
		Instructions:  j.bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	canNative := CanExecuteNatively(mainFn)
	if j.config.Debug {
		fmt.Printf("[JIT] CanExecuteNatively: %v, bytecode length: %d\n", canNative, len(j.bytecode.Instructions))
	}

	if canNative {
		// Try native execution
		start := time.Now()

		// Convert vm.Value globals to int64 array
		vmGlobals := j.GetGlobals()
		if j.config.Debug {
			fmt.Printf("[JIT] Globals array length: %d\n", len(vmGlobals))
		}

		globals := make([]int64, len(vmGlobals))
		for i, g := range vmGlobals {
			if g.IsInt() {
				globals[i], _ = g.ToInt()
			}
		}

		result, err := j.nativeExec.ExecuteFunction(mainFn, j.GetConstants(), globals)
		if err == nil {
			j.nativeExecs++
			if j.config.Debug {
				fmt.Printf("[JIT] Native execution succeeded in %v, result=%d\n", time.Since(start), result)
			}
			// Store the result so LastPoppedObject() can return it
			j.RegVM.SetLastResult(vm.NewInt(result))
			return nil
		}
		if j.config.Debug {
			fmt.Printf("[JIT] Native execution failed: %v, falling back to interpreter\n", err)
		}
	}

	// Try to find and compile recursive functions before falling back to interpreter
	j.tryCompileRecursiveFunctions()

	// Fall back to interpreter
	j.interpExecs++
	if j.config.Debug {
		fmt.Printf("[JIT] Using interpreter (native=%v)\n", canNative)
	}
	return j.RegVM.Run()
}

// tryCompileRecursiveFunctions attempts to find and compile recursive functions
// This enables future execution of compiled code when functions are called
func (j *JITVM) tryCompileRecursiveFunctions() {
	// Find all functions in constants and try to compile recursive ones
	for _, c := range j.bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			// Check if this is a recursive function that can be optimized
			fibCompiler := NewFibJITCompiler(j.config)
			analysis := fibCompiler.analyzeFunction(fn)

			if analysis.isSelfRecursive {
				// Try to compile this recursive function
				cf, err := j.jit.Compile(fn, j.GetConstants(), j.GetGlobals())
				if err == nil {
					if j.config.Debug {
						fmt.Printf("[JIT] Pre-compiled recursive function: hash=%016x, size=%d bytes\n",
							cf.Hash, cf.Size)
					}
				} else if j.config.Debug {
					fmt.Printf("[JIT] Failed to compile recursive function: %v\n", err)
				}
			}

			_ = fibCompiler // avoid unused variable error
		}
	}
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
	stats := j.jit.GetStats()
	// Add native execution stats
	return stats
}

// GetNativeStats returns native execution statistics
func (j *JITVM) GetNativeStats() (nativeExecs, interpExecs int64) {
	return j.nativeExecs, j.interpExecs
}

// Cleanup releases JIT resources
func (j *JITVM) Cleanup() {
	j.jit.Cleanup()
	j.nativeExec.Cleanup()
}

// CompileFunction compiles a specific function for JIT execution
func (j *JITVM) CompileFunction(fn *compiler.CompiledFunction) (*CompiledFunc, error) {
	return j.jit.Compile(fn, j.GetConstants(), j.GetGlobals())
}

// ExecuteCompiled executes a previously compiled function
func (j *JITVM) ExecuteCompiled(cf *CompiledFunc) int64 {
	return cf.Execute()
}

// SetSourcePath sets the source path for error messages
func (j *JITVM) SetSourcePath(path string) {
	j.RegVM.SetSourcePath(path)
}

// SetCurrentModule sets the current module context
func (j *JITVM) SetCurrentModule(module *objects.Module) {
	j.RegVM.SetCurrentModule(module)
}

// LastPoppedObject returns the last popped object
func (j *JITVM) LastPoppedObject() objects.Object {
	return j.RegVM.LastPoppedObject()
}
