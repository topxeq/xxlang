// +build amd64,!windows

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

	// Native function registry for compiled functions
	nativeRegistry *NativeFunctionRegistry

	// Map from register number to constant index for native functions
	// When OpRegClosure loads a native function into a register, we track it here
	registerToNative map[int]int

	// Statistics
	nativeExecs int64
	interpExecs int64
}

// NewJITVM creates a new JIT-enabled VM
func NewJITVM(bytecode *compiler.Bytecode, config JITConfig) *JITVM {
	j := &JITVM{
		RegVM:            vm.NewRegVM(bytecode),
		jit:              NewJITCompiler(config),
		nativeExec:       NewNativeExecutor(config),
		nativeRegistry:   NewNativeFunctionRegistry(config),
		registerToNative: make(map[int]int),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	// Pre-compile native-executable functions
	j.compileNativeFunctions()
	// Set up native call hook
	j.setupNativeCallHook()
	return j
}

// NewJITVMWithGlobals creates a JIT VM with custom globals
func NewJITVMWithGlobals(bytecode *compiler.Bytecode, globals []vm.Value, config JITConfig) *JITVM {
	j := &JITVM{
		RegVM:            vm.NewRegVMWithGlobals(bytecode, globals),
		jit:              NewJITCompiler(config),
		nativeExec:       NewNativeExecutor(config),
		nativeRegistry:   NewNativeFunctionRegistry(config),
		registerToNative: make(map[int]int),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	// Pre-compile native-executable functions
	j.compileNativeFunctions()
	// Set up native call hook
	j.setupNativeCallHook()
	return j
}

// NewJITVMWithObjectGlobals creates a JIT VM with globals as objects.Object
func NewJITVMWithObjectGlobals(bytecode *compiler.Bytecode, globals []objects.Object, config JITConfig) *JITVM {
	// Convert objects.Object to vm.Value
	valueGlobals := make([]vm.Value, len(globals))
	for i, obj := range globals {
		valueGlobals[i] = vm.NewObject(obj)
	}
	return NewJITVMWithGlobals(bytecode, valueGlobals, config)
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
	j.compileNativeFunctions()

	// Check if we have native functions compiled
	if j.config.Debug {
		fmt.Printf("[JIT] Compiled %d native functions\n", len(j.nativeRegistry.functions))
	}

	// Check if any native functions are available
	if len(j.nativeRegistry.functions) > 0 {
		// Use hybrid execution mode that intercepts calls to native functions
		return j.runHybrid()
	}

	// Fall back to interpreter
	j.interpExecs++
	if j.config.Debug {
		fmt.Printf("[JIT] Using interpreter (native=%v)\n", canNative)
	}
	return j.RegVM.Run()
}

// runHybrid executes bytecode with native function call interception
func (j *JITVM) runHybrid() error {
	if j.config.Debug {
		fmt.Printf("[JIT] Running in hybrid mode with %d native functions\n", len(j.nativeRegistry.functions))
	}

	// The native call hook is already set up in setupNativeCallHook
	// Just run the VM normally - the hook will intercept native function calls
	j.interpExecs++
	return j.RegVM.Run()
}

// setupNativeCallHook sets up the VM hook for native function execution
func (j *JITVM) setupNativeCallHook() {
	// Create a function index map for quick lookup
	// We need to map CompiledFunction pointers to their constant indices
	fnToIdx := make(map[*compiler.CompiledFunction]int)
	for i, c := range j.bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			fnToIdx[fn] = i
		}
	}

	// Set up the native call hook
	j.RegVM.SetNativeCallHook(func(fn *compiler.CompiledFunction, args []vm.Value, frame *vm.RegFrame) (vm.Value, bool) {
		// Check if this function has a native version
		idx, ok := fnToIdx[fn]
		if !ok {
			return vm.ValueNull, false
		}

		nativeFn := j.nativeRegistry.Get(idx)
		if nativeFn == nil {
			return vm.ValueNull, false
		}

		// Convert arguments to int64
		intArgs := make([]int64, len(args))
		for i, arg := range args {
			if arg.IsInt() {
				intArgs[i], _ = arg.ToInt()
			}
		}

		// Get globals
		vmGlobals := j.GetGlobals()
		globals := make([]int64, len(vmGlobals))
		for i, g := range vmGlobals {
			if g.IsInt() {
				globals[i], _ = g.ToInt()
			}
		}

		if j.config.Debug {
			fmt.Printf("[JIT] Native hook called: const[%d], args=%v, numParams=%d\n", idx, intArgs, nativeFn.NumParams)
		}

		// Execute native function
		result := nativeFn.Execute(globals, intArgs...)
		j.nativeExecs++

		if j.config.Debug {
			fmt.Printf("[JIT] Native hook executed function at const[%d], result=%d\n", idx, result)
		}

		return vm.NewInt(result), true
	})
}

// compileNativeFunctions pre-compiles all native-executable functions
// and attempts to compile recursive functions using the FibJIT compiler
func (j *JITVM) compileNativeFunctions() {
	// Extract integer constants
	intConstants := make([]int64, len(j.bytecode.Constants))
	for i, c := range j.bytecode.Constants {
		switch v := c.(type) {
		case *objects.Int:
			intConstants[i] = v.Value
		case *objects.Bool:
			if v.Value {
				intConstants[i] = 1
			}
		}
	}

	// Convert constants to vm.Value for FibJIT compiler
	vmConstants := make([]vm.Value, len(j.bytecode.Constants))
	for i, c := range j.bytecode.Constants {
		vmConstants[i] = vm.NewObject(c)
	}

	// Find all functions in constants and compile them
	for i, c := range j.bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			// First try: pure native execution (no function calls)
			if CanExecuteNatively(fn) {
				err := j.nativeRegistry.CompileFunction(fn, i, intConstants)
				if err != nil && j.config.Debug {
					fmt.Printf("[JIT] Failed to compile function at const[%d]: %v\n", i, err)
				}
				continue
			}

			// Second try: recursive function compilation (e.g., Fibonacci)
			// This handles functions with OpRegCall by transforming recursion
			// into iteration when the pattern is recognized
			if containsCall(fn.Instructions) {
				err := j.nativeRegistry.CompileRecursiveFunction(fn, i, vmConstants)
				if err != nil {
					if j.config.Debug {
						fmt.Printf("[JIT] Failed to compile recursive function at const[%d]: %v\n", i, err)
					}
				} else if j.config.Debug {
					fmt.Printf("[JIT] Successfully compiled recursive function at const[%d]\n", i)
				}
			}
		}
	}
}

// ExecuteNativeFunction executes a natively-compiled function with arguments
func (j *JITVM) ExecuteNativeFunction(constIdx int, args ...int64) (int64, bool) {
	nativeFn := j.nativeRegistry.Get(constIdx)
	if nativeFn == nil {
		return 0, false
	}

	// Convert globals to int64 array
	vmGlobals := j.GetGlobals()
	globals := make([]int64, len(vmGlobals))
	for i, g := range vmGlobals {
		if g.IsInt() {
			globals[i], _ = g.ToInt()
		}
	}

	return nativeFn.Execute(globals, args...), true
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
	j.nativeRegistry.Cleanup()
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
