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
	"syscall"
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
	// Reject functions with too many registers to avoid excessive stack usage.
	if fn.NumRegs > 64 {
		return false
	}

	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		// Supported operations - pure arithmetic and control flow (no calls)
		// NOTE: OpRegDiv uses SSE2 float division with integer truncation.
		// For integer-divisible operands (e.g., 20/10=2), this matches VM behavior.
		// For non-integer results (e.g., 7/2=3.5), JIT returns 3 (truncated)
		// while VM returns 3.5 (float). Full NaN-boxing support is needed for exact semantics.
		case compiler.OpRegLoadConst, compiler.OpRegMove,
			compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
			compiler.OpRegAddConst, compiler.OpRegSubConst, compiler.OpRegMulConst,
			compiler.OpRegNeg, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
			compiler.OpRegEqual, compiler.OpRegNotEqual,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegIncLocal, compiler.OpRegDecLocal, compiler.OpRegLoopCountAdd, compiler.OpRegAddLocalCheck,
			compiler.OpRegLoopIncCheck, compiler.OpRegLoopBodyAdd, compiler.OpRegLoopMulCheck,
			compiler.OpRegReturn,
			compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal,
			compiler.OpRegPop:
			// Supported - continue

		default:
			// Unsupported: OpRegCall, OpRegTailCall, OpRegDiv (returns float), builtin, array, map, closure, etc.
			return false
		}

		ip += 1 + operandWidth(op)
	}

	return true
}

// CanExecuteNativelyWithCalls checks if a function can be executed natively
// with callback-based dispatch for builtins, arrays, maps, and objects.
// NOTE: OpRegCall/OpRegTailCall are excluded because re-entering native code
// from within a callback (nested execution) causes deadlocks with the bridge ABI.
// Functions with OpRegCall use hybrid mode (interpreter + native call hook) instead.
func CanExecuteNativelyWithCalls(fn *compiler.CompiledFunction) bool {
	if fn.NumRegs > 64 {
		return false
	}

	code := fn.Instructions
	ip := 0

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		case compiler.OpRegLoadConst, compiler.OpRegMove,
			compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
			compiler.OpRegAddConst, compiler.OpRegSubConst, compiler.OpRegMulConst,
			compiler.OpRegNeg, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
			compiler.OpRegEqual, compiler.OpRegNotEqual,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegIncLocal, compiler.OpRegDecLocal, compiler.OpRegLoopCountAdd, compiler.OpRegAddLocalCheck,
			compiler.OpRegLoopIncCheck, compiler.OpRegLoopBodyAdd, compiler.OpRegLoopMulCheck,
			compiler.OpRegReturn,
			compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal,
			compiler.OpRegBuiltin,
			compiler.OpRegArray, compiler.OpRegArrayEmpty, compiler.OpRegArrayAppend,
			compiler.OpRegIndex, compiler.OpRegSetIndex,
			compiler.OpRegMap, compiler.OpRegMapEmpty, compiler.OpRegMapSet,
			compiler.OpRegGetField, compiler.OpRegSetField, compiler.OpRegGetMethod,
			compiler.OpRegPush, compiler.OpRegPop:
			// Supported with callback dispatch - continue

		default:
			return false
		}

		ip += 1 + operandWidth(op)
	}

	return true
}

// ExecuteFunction executes a compiled function natively.
// Only supports functions without OpRegCall (checked by CanExecuteNatively).
// For functions with calls, use ExecuteFunctionWithCalls instead.
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	if !CanExecuteNatively(fn) {
		return 0, fmt.Errorf("function cannot be executed natively")
	}

	intConstants := make([]int64, len(constants))
	for i, c := range constants {
		if c.IsInt() {
			intConstants[i], _ = c.ToInt()
		}
	}

	cg := NewNativeCodeGenerator()

	code, err := cg.Generate(fn, intConstants)
	if err != nil {
		return 0, fmt.Errorf("compilation failed: %w", err)
	}

	mem, _, err := ne.compiler.AllocCode(len(code))
	if err != nil {
		return 0, fmt.Errorf("memory allocation failed: %w", err)
	}

	copy(mem, code)
	entry := uintptr(unsafe.Pointer(&mem[0]))

	return callNativeWithGlobals(entry, globals), nil
}

// ExecuteFunctionWithCalls executes a compiled function natively using syscall.SyscallN.
// ExecuteFunctionWithCalls executes a compiled function natively using syscall.SyscallN,
// which enables OpRegCall callbacks via syscall.NewCallback to work.
//
// NOTE: This method is no longer used by Run() for the main code path. The main code
// with OpRegCall now always uses hybrid mode (interpreter + native call hook) because
// compileCall cannot resolve all function references at code generation time (e.g.,
// higher-order function parameters). This method is retained for potential future use
// in specialized scenarios where all call targets are known at compile time.
//
func (ne *NativeExecutor) ExecuteFunctionWithCalls(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	if !CanExecuteNativelyWithCalls(fn) {
		return 0, fmt.Errorf("function cannot be executed natively (with calls)")
	}

	// Guard against nested syscall.SyscallN — this would deadlock because
	// the goroutine is already in _Gsyscall state from the outer call.
	if isInSyscallN() {
		return 0, fmt.Errorf("cannot execute with calls: already inside syscall.SyscallN (nested deadlock)")
	}

	intConstants := make([]int64, len(constants))
	for i, c := range constants {
		if c.IsInt() {
			intConstants[i], _ = c.ToInt()
		}
	}

	hasCall := containsCall(fn.Instructions)
	hasCallbacks := hasCall || hasBuiltinOps(fn.Instructions) || hasCollectionOps(fn.Instructions)

	cg := NewNativeCodeGenerator()
	// Always use syscall ABI for functions with callback dispatch.
	// syscall.NewCallback callbacks require the goroutine to be in _Gsyscall state,
	// which only happens during syscall.SyscallN execution.
	if hasCallbacks {
		cg.syscallABI = true
	}

	code, err := cg.Generate(fn, intConstants)
	if err != nil {
		return 0, fmt.Errorf("compilation failed: %w", err)
	}

	mem, _, err := ne.compiler.AllocCode(len(code))
	if err != nil {
		return 0, fmt.Errorf("memory allocation failed: %w", err)
	}

	copy(mem, code)
	entry := uintptr(unsafe.Pointer(&mem[0]))

	if hasCallbacks {
		var globalsPtr uintptr
		if len(globals) > 0 {
			globalsPtr = uintptr(unsafe.Pointer(&globals[0]))
		}
		setInSyscallN(true)
		r1, _, _ := syscall.SyscallN(entry, globalsPtr, 0, 0, 0)
		setInSyscallN(false)
		return int64(r1), nil
	}

	return callNativeWithGlobals(entry, globals), nil
}

// Cleanup releases native executor resources.
func (ne *NativeExecutor) Cleanup() { ne.compiler.Cleanup() }

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
	j.initWindowsCallback()
	j.InitBuiltinCallback()
	j.InitCollectionCallback()
	j.InitObjectCallback()
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
	j.initWindowsCallback()
	j.InitBuiltinCallback()
	j.InitCollectionCallback()
	j.InitObjectCallback()
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

	// Clean up object handles after execution to prevent memory leaks.
	// Handles are only needed during JIT execution; after Run returns,
	// all results have been converted back to VM values.
	defer globalObjectRegistry.Clear()

	// Always compile native functions and set up hooks before any execution path,
	// because the main bytecode may contain OpRegCall that calls into functions.
	j.compileNativeFunctions()
	j.setupNativeCallHook()

	// Check if we can execute the main code natively (without OpRegCall).
	// Pure arithmetic/loop code can run entirely natively via bridge ABI.
	// Code with OpRegCall always uses hybrid mode (interpreter + native hook)
	// because compileCall cannot resolve all function references at compile time
	// (e.g., higher-order function parameters).
	mainFn := &compiler.CompiledFunction{
		Instructions:  j.bytecode.Instructions,
		NumLocals:     j.bytecode.MainNumLocals,
		NumParameters: 0,
		NumRegs:       j.bytecode.MainNumRegs,
	}

	// If MainNumRegs is not set (old bytecode or zero), use safe defaults
	if mainFn.NumRegs == 0 {
		mainFn.NumRegs = 16
	}
	if mainFn.NumLocals == 0 {
		mainFn.NumLocals = 16
	}

	canNative := CanExecuteNatively(mainFn)
	canNativeWithCalls := CanExecuteNativelyWithCalls(mainFn)
	hasCall := containsCall(j.bytecode.Instructions)

	if j.config.Debug {
		fmt.Printf("[JIT] CanExecuteNatively: %v, CanExecuteNativelyWithCalls: %v, hasCall: %v, NumRegs: %d, NumLocals: %d, bytecode length: %d\n",
			canNative, canNativeWithCalls, hasCall, mainFn.NumRegs, mainFn.NumLocals, len(j.bytecode.Instructions))
	}

	// Execution strategy:
	// - Pure arithmetic main code (no OpRegCall): execute natively via bridge ABI (fastest)
	// - Main code with builtins/collections but no OpRegCall: execute via syscall.SyscallN
	//   so that syscall.NewCallback callbacks can properly re-enter Go.
	// - Main code with OpRegCall: always use hybrid mode (interpreter + native call hook).
	//   This is necessary because compileCall may not be able to resolve all function
	//   references at code generation time (e.g., higher-order functions passed as parameters).
	if canNative && !hasCall {
		j.updateCachedGlobals()

		vmGlobals := j.GetGlobals()
		globals := make([]int64, len(vmGlobals))
		for i, g := range vmGlobals {
			globals[i] = valueToNativeInt(g)
		}

		result, err := j.nativeExec.ExecuteFunction(mainFn, j.GetConstants(), globals)
		if err == nil {
			j.nativeExecs++
			if j.config.Debug {
				fmt.Printf("[JIT] Native execution succeeded, result=%d\n", result)
			}
			j.syncGlobalsToVM()
			returnType := analyzeReturnTypeWithConstants(mainFn.Instructions, j.bytecode.Constants)
			resultVal := nativeResultToValue(result, returnType)
			if j.config.Debug {
				fmt.Printf("[JIT] returnType=%v, resultVal=%v, resultIsBool=%v\n", returnType, resultVal, resultVal.IsBool())
			}
			j.RegVM.SetLastResult(resultVal)
			return nil
		}
		if j.config.Debug {
			fmt.Printf("[JIT] Native execution failed: %v, falling back to hybrid/interpreter\n", err)
		}
	}

	// Try native execution with callback dispatch for builtins/collections.
	// Uses syscall.SyscallN so that syscall.NewCallback callbacks can re-enter Go.
	if canNativeWithCalls && !hasCall {
		j.updateCachedGlobals()

		vmGlobals := j.GetGlobals()
		globals := make([]int64, len(vmGlobals))
		for i, g := range vmGlobals {
			globals[i] = valueToNativeInt(g)
		}

		result, err := j.nativeExec.ExecuteFunctionWithCalls(mainFn, j.GetConstants(), globals)
		if err == nil {
			j.nativeExecs++
			if j.config.Debug {
				fmt.Printf("[JIT] Native execution with callbacks succeeded, result=%d\n", result)
			}
			j.syncGlobalsToVM()
			returnType := analyzeReturnTypeWithConstants(mainFn.Instructions, j.bytecode.Constants)
			resultVal := nativeResultToValue(result, returnType)
			j.RegVM.SetLastResult(resultVal)
			return nil
		}
		if j.config.Debug {
			fmt.Printf("[JIT] Native execution with callbacks failed: %v, falling back to hybrid\n", err)
		}
	}

	// For main code with calls, or when strict native execution fails,
	// use hybrid mode (interpreter + native call hook) or pure interpreter.
	if j.config.Debug {
		count := atomic.LoadInt64(&j.nativeRegistry.count)
		fmt.Printf("[JIT] Compiled %d native functions\n", count)
	}

	if atomic.LoadInt64(&j.nativeRegistry.count) > 0 {
		return j.runHybrid()
	}

	// Fall back to interpreter
	j.interpExecs++
	if j.config.Debug {
		fmt.Printf("[JIT] Using interpreter (canNative=%v, hasCall=%v)\n", canNative, hasCall)
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
			if nativeFn := j.nativeRegistry.Get(i); nativeFn != nil {
				j.nativeFuncCache[fn] = nativeFn
				continue
			}

			// Skip simple functions (they are faster in interpreter)
			if !j.shouldCompile(fn) {
				continue
			}

			// First try: native execution (pure arithmetic, no calls)
			if CanExecuteNatively(fn) {
				err := j.nativeRegistry.CompileFunction(fn, i, intConstants)
				if err != nil && j.config.Debug {
					fmt.Printf("[JIT] Failed to compile function at const[%d]: %v\n", i, err)
				}
				if nativeFn := j.nativeRegistry.Get(i); nativeFn != nil {
					j.nativeFuncCache[fn] = nativeFn
				}
				continue
			}

			// Second try: native execution with callback-based dispatch.
			// DISABLED for non-main functions: these functions use syscall ABI
			// and can only be called via syscall.SyscallN, but the native hook
			// uses bridge ABI. Enabling this causes "bad g in cgocallback" crashes
			// when the native hook tries to call them.
			// The main code path in Run() handles CanExecuteNativelyWithCalls
			// separately via ExecuteFunctionWithCalls().
			if false && CanExecuteNativelyWithCalls(fn) {
				err := j.nativeRegistry.CompileFunctionWithCalls(fn, i, intConstants)
				if err != nil && j.config.Debug {
					fmt.Printf("[JIT] Failed to compile function-with-calls at const[%d]: %v\n", i, err)
				}
				// Do NOT add to nativeFuncCache — these functions use syscall ABI
				// and can only be called from Run() via ExecuteFunctionWithCalls,
				// not from the interpreter's native hook (bridge ABI).
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
	// Minimum bytecode size threshold (in bytes).
	// Lowered from 20 to 10 to allow simple arithmetic functions (e.g., add(a,b))
	// to benefit from JIT compilation when called repeatedly from hot loops.
	const minCodeSize = 10

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
		// Builtin calls add moderate complexity
		if op == compiler.OpRegBuiltin {
			complexity += 3
		}
		// Arithmetic ops add basic complexity (even simple functions benefit from JIT)
		if op == compiler.OpRegAdd || op == compiler.OpRegSub ||
			op == compiler.OpRegMul || op == compiler.OpRegDiv || op == compiler.OpRegMod ||
			op == compiler.OpRegAddConst || op == compiler.OpRegSubConst || op == compiler.OpRegMulConst {
			complexity += 2
		}
		complexity++

		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}

	const minComplexity = 3
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

// syncGlobalsToVM writes back modified globals from the native int64 cache
// to the VM's NaN-boxed Value array. This is necessary because native code
// writes StoreGlobal results to cachedGlobals (int64), but the interpreter
// reads from its own vm.Value globals, so without syncing, StoreGlobal
// in native functions would be invisible to the interpreter.
// Only globals that held integer values before native execution are synced,
// since native code can only produce int64 results.
func (j *JITVM) syncGlobalsToVM() {
	j.cachedGlobalsLock.RLock()
	cached := j.cachedGlobals
	j.cachedGlobalsLock.RUnlock()

	vmGlobals := j.GetGlobals()
	n := len(cached)
	if n > len(vmGlobals) {
		n = len(vmGlobals)
	}
	for i := 0; i < n; i++ {
		// Only write back if the VM global was an integer (native code
		// can't modify non-int globals) and the value differs
		if vmGlobals[i].IsInt() {
			oldVal, _ := vmGlobals[i].ToInt()
			if oldVal != cached[i] {
				vmGlobals[i] = vm.NewInt(cached[i])
			}
		}
	}
}

// setupNativeCallHook sets up the VM hook for native function execution
func (j *JITVM) setupNativeCallHook() {
	// Set up the native call hook regardless of whether nativeFuncCache is empty.
	// Even if no functions were compiled natively, the main code may contain
	// OpRegCall that dispatches through this hook when executed via syscall.SyscallN.
	// In that case, the hook falls back to interpreter execution.

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
			intArgs[i] = valueToNativeInt(arg)
		}

		// Refresh cached globals from the VM before native execution
		// so the native code sees the latest values written by the interpreter
		j.updateCachedGlobals()

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

		// Write back modified globals from native code to the VM
		// Native StoreGlobal writes to cachedGlobals but the interpreter
		// reads from its own vm.Value globals array, so we must sync back.
		j.syncGlobalsToVM()

		return nativeResultToValue(result, nativeFn.ReturnType), true
	})
}

// SetJITEnabled enables or disables JIT compilation
func (j *JITVM) SetJITEnabled(enabled bool) {
	j.enabled = enabled
}

// GetJITStats returns JIT compilation statistics
func (j *JITVM) GetJITStats() JITStats {
	stats := j.jit.GetStats()
	j.nativeRegistry.functions.Range(func(_, value interface{}) bool {
		nf, ok := value.(*NativeFunction)
		if !ok {
			return true
		}

		stats.CompiledFunctions++
		stats.TotalCodeSize += int64(len(nf.Code))
		return true
	})

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
	UseBridgeABI bool
	ReturnType NativeReturnType
	UseSyscallABI bool // When true, use syscall.Syscall6 for execution (enables OpRegCall callbacks)
	entryPtr  *byte
	entry     uintptr
}

func nativeResultToValue(result int64, returnType NativeReturnType) vm.Value {
	switch returnType {
	case ReturnTypeBool:
		return vm.NewBool(result != 0)
	case ReturnTypeNull:
		return vm.ValueNull
	default:
		return vm.NewInt(result)
	}
}

func valueToNativeInt(v vm.Value) int64 {
	if v.IsInt() {
		return v.GetInt()
	}
	if v.IsBool() {
		if v.GetBool() {
			return 1
		}
		return 0
	}
	return 0
}

// Execute runs the native function
func (nf *NativeFunction) Execute(globals []int64, args ...int64) int64 {
	if len(nf.Code) == 0 || nf.entry == 0 {
		return 0
	}

	// Syscall ABI: use syscall.Syscall6 for proper goroutine state transition.
	// The goroutine enters _Gsyscall state, enabling syscall.NewCallback callbacks
	// (needed for OpRegCall) to correctly re-enter Go via exitsyscall/entersyscall.
	if nf.UseSyscallABI {
		return nf.executeSyscall(globals, args)
	}

	if nf.UseBridgeABI {
		switch len(args) {
		case 0:
			return bridge.Call0(nf.entryPtr)
		case 1:
			return bridge.Call1(nf.entryPtr, args[0])
		case 2:
			return bridge.Call2(nf.entryPtr, args[0], args[1])
		default:
			return bridge.Call3(nf.entryPtr, args[0], args[1], args[2])
		}
	}

	if len(args) <= 3 {
		var arg0, arg1, arg2 int64
		if len(args) > 0 {
			arg0 = args[0]
		}
		if len(args) > 1 {
			arg1 = args[1]
		}
		if len(args) > 2 {
			arg2 = args[2]
		}

		var globalsPtr *int64
		if len(globals) > 0 {
			globalsPtr = &globals[0]
		}

		return callNativeWithArgs(nf.entry, globalsPtr, arg0, arg1, arg2)
	}

	if len(args) <= 8 {
		args8 := make([]int64, 8)
		copy(args8, args)

		var globalsPtr *int64
		if len(globals) > 0 {
			globalsPtr = &globals[0]
		}

		return callNativeWithArgs8(nf.entry, globalsPtr, &args8[0])
	}

	return callNativeWithGlobals(nf.entry, globals)
}

// executeSyscall calls the JIT function via syscall.SyscallN using Windows x64 ABI.
// This transitions the goroutine to _Gsyscall state, which is required for
// syscall.NewCallback to properly re-enter Go for OpRegCall dispatch.
func (nf *NativeFunction) executeSyscall(globals []int64, args []int64) int64 {
	var globalsPtr uintptr
	if len(globals) > 0 {
		globalsPtr = uintptr(unsafe.Pointer(&globals[0]))
	}

	callArgs := []uintptr{globalsPtr}
	for _, a := range args {
		if len(callArgs) >= 4 {
			break
		}
		callArgs = append(callArgs, uintptr(a))
	}
	// Pad to at least 4 args (globals + 3 args) for consistent JIT prologue behavior
	for len(callArgs) < 4 {
		callArgs = append(callArgs, 0)
	}

	r1, _, _ := syscall.SyscallN(nf.entry, callArgs...)
	return int64(r1)
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
	if !CanExecuteNatively(fn) {
		return fmt.Errorf("function cannot be executed natively")
	}

	// NOTE: Since CanExecuteNatively rejects OpRegCall, hasCall is always false.
	hasCall := containsCall(fn.Instructions)

	cg := NewNativeCodeGenerator()
	if hasCall {
		cg.syscallABI = true
	}

	code, err := cg.Generate(fn, constants)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	mem, err := bridge.AllocExecMem(len(code))
	if err != nil {
		return err
	}

	copy(mem, code)

	nf := &NativeFunction{
		Code:         mem,
		NumParams:    fn.NumParameters,
		ReturnType:   analyzeReturnType(fn.Instructions),
		UseBridgeABI: false,
		UseSyscallABI: hasCall,
		entryPtr:     &mem[0],
		entry:        uintptr(unsafe.Pointer(&mem[0])),
	}
	r.functions.Store(idx, nf)
	atomic.AddInt64(&r.count, 1)

	return nil
}

// CompileFunctionWithCalls compiles a function that contains OpRegBuiltin,
// array/map/object operations for native execution using callback dispatch.
// The compiled code uses syscall ABI (Windows x64 calling convention) so it
// can be executed via syscall.SyscallN, which puts the goroutine in _Gsyscall
// state. This is required for syscall.NewCallback callbacks to properly
// re-enter Go for builtin/collection/object dispatch.
// NOTE: Functions compiled this way must NOT be called from the native hook
// (bridge ABI); they can only be executed via ExecuteFunctionWithCalls().
func (r *NativeFunctionRegistry) CompileFunctionWithCalls(fn *compiler.CompiledFunction, idx int, constants []int64) error {
	if !CanExecuteNativelyWithCalls(fn) {
		return fmt.Errorf("function cannot be executed natively with calls")
	}

	// Ensure builtin callback is initialized before compiling
	if GetBuiltinCallbackPtr() == 0 {
		return fmt.Errorf("builtin callback not initialized")
	}

	cg := NewNativeCodeGenerator()
	cg.syscallABI = true // Windows x64 ABI for syscall.SyscallN entry

	code, err := cg.Generate(fn, constants)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	mem, err := bridge.AllocExecMem(len(code))
	if err != nil {
		return err
	}

	copy(mem, code)

	hasCall := containsCall(fn.Instructions)
	nf := &NativeFunction{
		Code:          mem,
		NumParams:     fn.NumParameters,
		ReturnType:    analyzeReturnType(fn.Instructions),
		UseBridgeABI:  false,
		UseSyscallABI: hasCall,
		entryPtr:      &mem[0],
		entry:         uintptr(unsafe.Pointer(&mem[0])),
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
		Code:         mem,
		NumParams:    fn.NumParameters,
		ReturnType:   ReturnTypeInt,
		UseBridgeABI: true,
		entryPtr:     &mem[0],
		entry:        uintptr(unsafe.Pointer(&mem[0])),
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
	atomic.StoreInt64(&r.count, 0)
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
