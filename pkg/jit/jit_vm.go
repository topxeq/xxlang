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

	// Cached globals (int64 representation) - updated when globals change
	cachedGlobals     []int64
	cachedGlobalsLock sync.RWMutex

	// Inline cache: maps function pointer directly to native function
	// This avoids the map lookup overhead in the hot path
	nativeFuncCache map[*compiler.CompiledFunction]*NativeCompiledFunc

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
		nativeFuncCache:  make(map[*compiler.CompiledFunction]*NativeCompiledFunc),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	// Pre-compile native-executable functions
	j.compileNativeFunctions()
	// Pre-cache globals
	j.updateCachedGlobals()
	// Set up native call hook only if we have compiled functions
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
		nativeFuncCache:  make(map[*compiler.CompiledFunction]*NativeCompiledFunc),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	// Pre-compile native-executable functions
	j.compileNativeFunctions()
	// Pre-cache globals
	j.updateCachedGlobals()
	// Set up native call hook only if we have compiled functions
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
	// Only set up hook if we have compiled native functions
	if len(j.nativeFuncCache) == 0 {
		return
	}

	// Set up fast check for native functions
	j.RegVM.SetFastNativeCheck(func(fn *compiler.CompiledFunction) bool {
		_, ok := j.nativeFuncCache[fn]
		return ok
	})

	// Set up the native call hook with inline cache
	j.RegVM.SetNativeCallHook(func(fn *compiler.CompiledFunction, args []vm.Value, frame *vm.RegFrame) (vm.Value, bool) {
		// Fast path: check inline cache directly
		nativeFn, ok := j.nativeFuncCache[fn]
		if !ok {
			// Not a compiled function, let interpreter handle it
			return vm.ValueNull, false
		}

		// Convert arguments to int64 (only for compiled functions)
		intArgs := make([]int64, len(args))
		for i, arg := range args {
			if arg.IsInt() {
				intArgs[i], _ = arg.ToInt()
			}
		}

		// Use cached globals
		j.cachedGlobalsLock.RLock()
		globals := j.cachedGlobals
		j.cachedGlobalsLock.RUnlock()

		if j.config.Debug {
			fmt.Printf("[JIT] Native hook called: args=%v, numParams=%d\n", intArgs, nativeFn.NumParams)
		}

		// Execute native function
		result := nativeFn.Execute(globals, intArgs...)
		j.nativeExecs++

		if j.config.Debug {
			fmt.Printf("[JIT] Native hook executed, result=%d\n", result)
		}

		return vm.NewInt(result), true
	})
}

// updateCachedGlobals updates the cached int64 representation of globals
func (j *JITVM) updateCachedGlobals() {
	j.cachedGlobalsLock.Lock()
	defer j.cachedGlobalsLock.Unlock()

	vmGlobals := j.GetGlobals()
	// Only allocate if needed
	if j.cachedGlobals == nil || len(j.cachedGlobals) != len(vmGlobals) {
		j.cachedGlobals = make([]int64, len(vmGlobals))
	}
	for i, g := range vmGlobals {
		if g.IsInt() {
			j.cachedGlobals[i], _ = g.ToInt()
		} else {
			j.cachedGlobals[i] = 0
		}
	}
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
			// Skip simple functions (they are faster in interpreter)
			if !j.shouldCompile(fn) {
				continue
			}

			// First try: pure native execution (no function calls)
			if CanExecuteNatively(fn) {
				err := j.nativeRegistry.CompileFunction(fn, i, intConstants)
				if err != nil && j.config.Debug {
					fmt.Printf("[JIT] Failed to compile function at const[%d]: %v\n", i, err)
				}
				// Add to inline cache on success
				if nativeFn := j.nativeRegistry.Get(i); nativeFn != nil {
					j.nativeFuncCache[fn] = nativeFn
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
				} else {
					if j.config.Debug {
						fmt.Printf("[JIT] Successfully compiled recursive function at const[%d]\n", i)
					}
					// Add to inline cache on success
					if nativeFn := j.nativeRegistry.Get(i); nativeFn != nil {
						j.nativeFuncCache[fn] = nativeFn
					}
				}
			}
		}
	}
}

// shouldCompile determines if a function is worth JIT compiling
// Simple functions are often faster in the interpreter due to lower overhead
func (j *JITVM) shouldCompile(fn *compiler.CompiledFunction) bool {
	// Minimum bytecode size threshold (in bytes)
	// Functions smaller than this are faster in interpreter
	const minCodeSize = 20

	// Functions with very few instructions are not worth compiling
	if len(fn.Instructions) < minCodeSize {
		if j.config.Debug {
			fmt.Printf("[JIT] Skipping small function (%d bytes, threshold %d)\n",
				len(fn.Instructions), minCodeSize)
		}
		return false
	}

	// Count instruction complexity
	complexity := 0
	for i := 0; i < len(fn.Instructions); {
		op := compiler.Opcode(fn.Instructions[i])
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			break
		}

		// Loops and branches add complexity
		if op == compiler.OpRegJump || op == compiler.OpRegJumpIfFalse ||
			op == compiler.OpRegJumpIfTrue || op == compiler.OpRegLoopCountAdd {
			complexity += 5
		}
		// Function calls (recursion) add significant complexity
		if op == compiler.OpRegCall || op == compiler.OpRegTailCall {
			complexity += 10
		}
		// Basic operations
		complexity++

		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}

	// Minimum complexity threshold - lower for recursive functions
	const minComplexity = 5
	if complexity < minComplexity {
		if j.config.Debug {
			fmt.Printf("[JIT] Skipping simple function (complexity %d, threshold %d)\n",
				complexity, minComplexity)
		}
		return false
	}

	return true
}

// ExecuteNativeFunction executes a natively-compiled function with arguments
func (j *JITVM) ExecuteNativeFunction(constIdx int, args ...int64) (int64, bool) {
	nativeFn := j.nativeRegistry.Get(constIdx)
	if nativeFn == nil {
		return 0, false
	}

	// Use cached globals
	j.cachedGlobalsLock.RLock()
	globals := j.cachedGlobals
	j.cachedGlobalsLock.RUnlock()

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
