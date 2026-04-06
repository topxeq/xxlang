// +build windows,amd64

// pkg/jit/jit_windows_support.go
// Windows-specific JIT support files
// Includes cache, native executor, JIT VM, and code generators

package jit

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

// operandWidth returns the total operand width for an opcode
func operandWidth(op compiler.Opcode) int {
	def, err := compiler.Lookup(byte(op))
	if err != nil {
		return 0
	}
	total := 0
	for _, w := range def.OperandWidths {
		total += w
	}
	return total
}

// ============================================================
// Cache
// ============================================================

// hashBytecode generates a hash for bytecode
func hashBytecode(code []byte) uint64 {
	h := fnv.New64a()
	h.Write(code)
	return h.Sum64()
}

// ============================================================
// Native Executor
// ============================================================

// NativeExecutor executes JIT-compiled native code directly
type NativeExecutor struct {
	compiler *JITCompiler
	config   JITConfig
}

// NewNativeExecutor creates a new native JIT executor
func NewNativeExecutor(config JITConfig) *NativeExecutor {
	return &NativeExecutor{
		compiler: NewJITCompiler(config),
		config:   config,
	}
}

// CanExecuteNatively checks if a function can be executed natively
// Functions must meet these criteria:
// - No function calls (OpRegCall or OpRegTailCall)
// - No builtin calls
// - No array/map operations
// - Only: LoadConst, Move, Add, Sub, Mul, Div, Mod, Neg
//        Less, Greater, Equal, NotEqual, LessEqual, GreaterEqual
//        Jump, JumpIfTrue, JumpIfFalse, Return, Null, True, False
//        LoadGlobal, StoreGlobal (with globals pointer passed as argument)
//        LoadLocal, StoreLocal (local variables)
func CanExecuteNatively(fn *compiler.CompiledFunction) bool {
	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		// Supported operations - pure arithmetic and control flow
		case compiler.OpRegLoadConst, compiler.OpRegMove,
			compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
			compiler.OpRegNeg,
			compiler.OpRegEqual, compiler.OpRegNotEqual,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegReturn,
			compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal,
			compiler.OpRegPop:
			// Supported - continue

		default:
			// Unsupported operation
			return false
		}

		ip += 1 + operandWidth(op)
	}

	return true
}

// ExecuteFunction executes a compiled function natively
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	// Generate native code for Windows x64 ABI
	code, err := generateNativeCode(fn, constants, globals)
	if err != nil {
		return 0, err
	}

	// Allocate executable memory
	mem, err := bridge.AllocExecMem(len(code))
	if err != nil {
		return 0, err
	}
	defer bridge.FreeExecMem(mem)

	// Copy code
	copy(mem, code)

	// Execute using Windows calling convention
	// Safety: check if globals is not empty before accessing
	var globalsPtr *int64
	if len(globals) > 0 {
		globalsPtr = &globals[0]
	}
	result := callNativeWindows(uintptr(unsafe.Pointer(&mem[0])), globalsPtr)
	return result, nil
}

// Cleanup does nothing for now
func (ne *NativeExecutor) Cleanup() {}

// ============================================================
// Native Code Generator
// ============================================================

// generateNativeCode generates x86-64 native code for Windows
func generateNativeCode(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) ([]byte, error) {
	code := make([]byte, 0, 1024)

	// Prologue - save callee-saved registers (Windows x64 ABI)
	// rbx, rbp, rdi, rsi, r12-r15 are callee-saved
	code = append(code,
		0x53,             // push rbx
		0x41, 0x54,       // push r12
		0x41, 0x55,       // push r13
		0x41, 0x56,       // push r14
		0x41, 0x57,       // push r15
		0x48, 0x89, 0xCB, // mov rbx, rcx  (save globals pointer - first arg in rcx on Windows)
	)

	// Allocate local variables on stack (simulate registers)
	numRegs := fn.NumLocals
	if numRegs < 16 {
		numRegs = 16
	}
	stackSize := numRegs * 8
	if stackSize > 0 {
		// sub rsp, stackSize
		code = append(code, 0x48, 0x81, 0xEC)
		code = binary.LittleEndian.AppendUint32(code, uint32(stackSize))
	}

	// Generate code for each instruction
	// ... (simplified - just return 0 for now)

	// Epilogue
	if stackSize > 0 {
		// add rsp, stackSize
		code = append(code, 0x48, 0x81, 0xC4)
		code = binary.LittleEndian.AppendUint32(code, uint32(stackSize))
	}

	// Restore callee-saved registers
	code = append(code,
		0x41, 0x5F, // pop r15
		0x41, 0x5E, // pop r14
		0x41, 0x5D, // pop r13
		0x41, 0x5C, // pop r12
		0x5B,       // pop rbx
		0xC3,       // ret
	)

	return code, nil
}

// callNativeWindows calls native code with Windows x64 ABI
func callNativeWindows(entry uintptr, globals *int64) int64 {
	// Use bridge.Call1 for Windows
	return bridge.Call1((*byte)(unsafe.Pointer(entry)), *globals)
}

// ============================================================
// JIT VM
// ============================================================

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

	// Inline cache: maps function pointer directly to native function
	nativeFuncCache map[*compiler.CompiledFunction]*NativeFunction

	// Cached globals (int64 representation)
	cachedGlobals     []int64
	cachedGlobalsLock sync.RWMutex

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
		nativeFuncCache:  make(map[*compiler.CompiledFunction]*NativeFunction),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	j.compileNativeFunctions()
	j.updateCachedGlobals()
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
		nativeFuncCache:  make(map[*compiler.CompiledFunction]*NativeFunction),
		config:           config,
		enabled:          true,
		bytecode:         bytecode,
	}
	j.compileNativeFunctions()
	j.updateCachedGlobals()
	j.setupNativeCallHook()
	return j
}

// NewJITVMWithObjectGlobals creates a JIT VM with globals as objects.Object
func NewJITVMWithObjectGlobals(bytecode *compiler.Bytecode, globals []objects.Object, config JITConfig) *JITVM {
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
		vmGlobals := j.GetGlobals()
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
				fmt.Printf("[JIT] Native execution succeeded, result=%d\n", result)
			}
			j.RegVM.SetLastResult(vm.NewInt(result))
			return nil
		}
		if j.config.Debug {
			fmt.Printf("[JIT] Native execution failed: %v, falling back to interpreter\n", err)
		}
	}

	// Re-compile native functions (in case they weren't compiled yet)
	j.compileNativeFunctions()

	// Check if we have native functions compiled
	if j.config.Debug {
		count := atomic.LoadInt64(&j.nativeRegistry.count)
		fmt.Printf("[JIT] Compiled %d native functions\n", count)
	}

	// Check if any native functions are available
	if atomic.LoadInt64(&j.nativeRegistry.count) > 0 {
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
		count := atomic.LoadInt64(&j.nativeRegistry.count)
		fmt.Printf("[JIT] Running in hybrid mode with %d native functions\n", count)
	}

	// The native call hook is already set up in setupNativeCallHook
	// Just run the VM normally - the hook will intercept native function calls
	j.interpExecs++
	return j.RegVM.Run()
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

	// Convert constants to vm.Value for recursive function compilation
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
func (j *JITVM) shouldCompile(fn *compiler.CompiledFunction) bool {
	// Minimum bytecode size threshold (in bytes)
	const minCodeSize = 20

	if len(fn.Instructions) < minCodeSize {
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
		complexity++

		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}

	const minComplexity = 5
	return complexity >= minComplexity
}

// updateCachedGlobals updates the cached int64 representation of globals
func (j *JITVM) updateCachedGlobals() {
	j.cachedGlobalsLock.Lock()
	defer j.cachedGlobalsLock.Unlock()

	vmGlobals := j.GetGlobals()
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
			return vm.ValueNull, false
		}

		// Convert arguments to int64
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

// SetJITEnabled enables or disables JIT compilation
func (j *JITVM) SetJITEnabled(enabled bool) {
	j.enabled = enabled
}

// GetJITStats returns JIT compilation statistics
func (j *JITVM) GetJITStats() JITStats {
	return j.jit.GetStats()
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

// SetLoader sets the module loader
func (j *JITVM) SetLoader(loader interface{}) {
	// Type assertion for module loader
	if l, ok := loader.(*module.Loader); ok {
		j.RegVM.SetLoader(l)
	}
}

// GlobalsAsObjects returns globals as objects
func (j *JITVM) GlobalsAsObjects() []objects.Object {
	return j.RegVM.GlobalsAsObjects()
}

// ============================================================
// Native Function Registry
// ============================================================

// NativeFunction represents a natively compiled function
type NativeFunction struct {
	Code      []byte
	NumParams int
	entry     uintptr
}

// Execute runs the native function
func (nf *NativeFunction) Execute(globals []int64, args ...int64) int64 {
	if len(nf.Code) == 0 || nf.entry == 0 {
		return 0
	}

	// Call based on number of arguments (Windows x64 ABI uses rcx, rdx, r8, r9)
	switch len(args) {
	case 0:
		return bridge.Call0((*byte)(unsafe.Pointer(nf.entry)))
	case 1:
		return bridge.Call1((*byte)(unsafe.Pointer(nf.entry)), args[0])
	case 2:
		return bridge.Call2((*byte)(unsafe.Pointer(nf.entry)), args[0], args[1])
	default:
		return bridge.Call3((*byte)(unsafe.Pointer(nf.entry)), args[0], args[1], args[2])
	}
}

// NativeFunctionRegistry manages compiled native functions
type NativeFunctionRegistry struct {
	config    JITConfig
	functions sync.Map // map[int]*NativeFunction
	count     int64
}

// NewNativeFunctionRegistry creates a new registry
func NewNativeFunctionRegistry(config JITConfig) *NativeFunctionRegistry {
	return &NativeFunctionRegistry{config: config}
}

// Get returns a native function by constant index
func (r *NativeFunctionRegistry) Get(idx int) *NativeFunction {
	if v, ok := r.functions.Load(idx); ok {
		return v.(*NativeFunction)
	}
	return nil
}

// CompileFunction compiles a function to native code
func (r *NativeFunctionRegistry) CompileFunction(fn *compiler.CompiledFunction, idx int, constants []int64) error {
	// Generate native code
	code, err := generateNativeCode(fn, nil, nil)
	if err != nil {
		return err
	}

	// Allocate executable memory
	mem, err := bridge.AllocExecMem(len(code))
	if err != nil {
		return err
	}

	// Copy code
	copy(mem, code)

	// Store function
	nf := &NativeFunction{
		Code:      mem,
		NumParams: fn.NumParameters,
		entry:     uintptr(unsafe.Pointer(&mem[0])),
	}
	r.functions.Store(idx, nf)
	atomic.AddInt64(&r.count, 1)

	return nil
}

// CompileRecursiveFunction compiles a recursive function
func (r *NativeFunctionRegistry) CompileRecursiveFunction(fn *compiler.CompiledFunction, idx int, constants []vm.Value) error {
	// Use FibJIT compiler for recursive functions
	fibCompiler := NewFibJITCompiler(r.config)
	code, err := fibCompiler.Compile(fn, constants, nil)
	if err != nil {
		return err
	}

	// Allocate executable memory
	mem, err := bridge.AllocExecMem(len(code))
	if err != nil {
		return err
	}

	// Copy code
	copy(mem, code)

	// Store function
	nf := &NativeFunction{
		Code:      mem,
		NumParams: fn.NumParameters,
		entry:     uintptr(unsafe.Pointer(&mem[0])),
	}
	r.functions.Store(idx, nf)
	atomic.AddInt64(&r.count, 1)

	return nil
}

// Cleanup releases all resources
func (r *NativeFunctionRegistry) Cleanup() {
	r.functions.Range(func(key, value interface{}) bool {
		if nf, ok := value.(*NativeFunction); ok && len(nf.Code) > 0 {
			bridge.FreeExecMem(nf.Code)
		}
		return true
	})
	r.functions = sync.Map{}
}

// AnalyzeNativeSupport returns the support level
func AnalyzeNativeSupport(fn *compiler.CompiledFunction) int {
	if CanExecuteNatively(fn) {
		return SupportPureArithmetic
	}
	return SupportNone
}

// Support levels
const (
	SupportNone           = 0
	SupportPureArithmetic = 1
	SupportWithBuiltins   = 2
	SupportWithCalls      = 3
	SupportWithArrays     = 4
	SupportWithObjects    = 5
)
