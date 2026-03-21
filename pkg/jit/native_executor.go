// pkg/jit/native_executor.go
// Native JIT executor that runs compiled x86-64 code directly
package jit

import (
	"fmt"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

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
			compiler.OpRegNeg, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegEqual,
			compiler.OpRegNotEqual, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegReturn, compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegIncLocal, compiler.OpRegDecLocal,
			compiler.OpRegLoopCountAdd, compiler.OpRegLoopBodyAdd, compiler.OpRegLoopIncCheck,
			compiler.OpRegAddConst, compiler.OpRegSubConst, compiler.OpRegMulConst,
			compiler.OpRegAddLocalCheck, compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal:
			// These are fine for native execution

		// Unsupported - requires VM context or cross-function calls
		case compiler.OpRegCall, compiler.OpRegTailCall,
			compiler.OpRegArray, compiler.OpRegArrayEmpty, compiler.OpRegArrayAppend,
			compiler.OpRegMap, compiler.OpRegMapEmpty, compiler.OpRegMapSet,
			compiler.OpRegIndex, compiler.OpRegSetIndex,
			compiler.OpRegBuiltin,
			compiler.OpRegGetMethod, compiler.OpRegCallMethod,
			compiler.OpRegGetField, compiler.OpRegSetField,
			compiler.OpRegClosure, compiler.OpRegLoadFree, compiler.OpRegStoreFree,
			compiler.OpRegPush, compiler.OpRegPop,
			compiler.OpRegClass, compiler.OpRegNew,
			compiler.OpRegThrow, compiler.OpRegPushHandler, compiler.OpRegPopHandler,
			compiler.OpRegLoadModule, compiler.OpRegGetExport, compiler.OpRegSetExport,
			compiler.OpRegIterKey, compiler.OpRegIterValue,
			compiler.OpRegLoadFunc:
			return false

		default:
			// Unknown opcode - be conservative
			return false
		}

		// Skip operands
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return false
		}
		operandWidth := 0
		for _, w := range def.OperandWidths {
			operandWidth += w
		}
		ip += operandWidth + 1
	}

	return true
}

// ExecuteFunction compiles and executes a function natively with globals
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	// Check if can execute natively
	if !CanExecuteNatively(fn) {
		return 0, fmt.Errorf("function cannot be executed natively")
	}

	// Extract integer constants
	intConstants := make([]int64, len(constants))
	for i, c := range constants {
		if c.IsInt() {
			intConstants[i], _ = c.ToInt()
		}
	}

	// Compile using native code generator
	cg := NewNativeCodeGenerator()
	code, err := cg.Generate(fn, intConstants)
	if err != nil {
		return 0, fmt.Errorf("compilation failed: %w", err)
	}

	if ne.config.Debug {
		fmt.Printf("[JIT] Generated native code: %d bytes\n", len(code))
		// Print first 64 bytes of code in hex
		fmt.Printf("[JIT] Code: %x...\n", code[:min(64, len(code))])
	}

	// Allocate executable memory
	mem, _, err := ne.compiler.AllocCode(len(code))
	if err != nil {
		return 0, fmt.Errorf("memory allocation failed: %w", err)
	}

	copy(mem, code)

	// Execute with globals pointer
	// The native function takes globals pointer as first argument (in rdi)
	entry := uintptr(unsafe.Pointer(&mem[0]))
	result := callNativeWithGlobals(entry, globals)

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// callNativeWithGlobals calls a native function with globals pointer
// This function is implemented in assembly (bridge_amd64.s)
func callNative(entry uintptr, globals *int64) int64

// callNativeWithArgs calls a native function with globals and initial register values
// This function is implemented in assembly (bridge_amd64.s)
func callNativeWithArgs(entry uintptr, globals *int64, arg0, arg1, arg2 int64) int64

// callBuiltinCallback calls a Go callback for builtin functions from native code
// This function is implemented in assembly (bridge_amd64.s)
func callBuiltinCallback(callback uintptr, builtinIdx, numArgs int, argsPtr *int64) int64

// callFunctionCallback calls a Go callback for function dispatch from native code
// This function is implemented in assembly (bridge_amd64.s)
func callFunctionCallback(callback uintptr, funcReg, numArgs int, argsPtr *int64) int64

// callCollectionCallback calls a Go callback for collection operations from native code
// This function is implemented in assembly (bridge_amd64.s)
func callCollectionCallback(callback uintptr, opKind, numArgs int, argsPtr *int64) int64

// callNativeWithGlobals calls a native function with globals pointer
func callNativeWithGlobals(entry uintptr, globals []int64) int64 {
	if len(globals) == 0 {
		return callNative(entry, nil)
	}
	return callNative(entry, &globals[0])
}

// ExecuteBytecode executes a bytecode program natively
// Returns the result as an int64 (for numeric results)
func (ne *NativeExecutor) ExecuteBytecode(bytecode *compiler.Bytecode, globals []int64) (int64, error) {
	// Find main instructions (the top-level code)
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16, // reasonable default
		NumParameters: 0,
	}

	// Convert constants to VM values
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	return ne.ExecuteFunction(mainFn, constants, globals)
}

// Cleanup releases JIT resources
func (ne *NativeExecutor) Cleanup() {
	ne.compiler.Cleanup()
}

// ============================================================================
// High-performance native execution for simple programs
// ============================================================================

// NativeProgram represents a program compiled for native execution
type NativeProgram struct {
	entry     uintptr
	code      []byte
	page      *CodePage
	numRegs   int
	constants []int64 // Pre-extracted integer constants
}

// CompileProgram compiles a bytecode program to native code
func (ne *NativeExecutor) CompileProgram(bytecode *compiler.Bytecode) (*NativeProgram, error) {
	// Extract integer constants
	constants := make([]int64, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		switch v := c.(type) {
		case *objects.Int:
			constants[i] = v.Value
		case *objects.Bool:
			if v.Value {
				constants[i] = 1
			}
		}
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	// Use native code generator
	cg := NewNativeCodeGenerator()
	code, err := cg.Generate(mainFn, constants)
	if err != nil {
		return nil, err
	}

	// Allocate executable memory
	mem, page, err := ne.compiler.AllocCode(len(code))
	if err != nil {
		return nil, err
	}

	copy(mem, code)

	return &NativeProgram{
		entry:     uintptr(unsafe.Pointer(&mem[0])),
		code:      code,
		page:      page,
		numRegs:   16,
		constants: constants,
	}, nil
}

// Run executes the native program with globals
func (p *NativeProgram) Run(globals []int64) int64 {
	return callNativeWithGlobals(p.entry, globals)
}

// ============================================================================
// Native function registry for inter-function calls
// ============================================================================

// NativeCompiledFunc represents a natively-compiled function
type NativeCompiledFunc struct {
	Entry       uintptr    // Native code entry point
	Code        []byte     // Native code bytes
	Page        *CodePage  // Memory page (for cleanup)
	NumParams   int        // Number of parameters
	NumLocals   int        // Number of local variables
	Constants   []int64    // Integer constants
	IsRecursive bool       // True if function is self-recursive
	UseBridgeABI bool      // True if function uses System V ABI (bridge.Call1/2/3)
}

// NativeFunctionRegistry manages compiled native functions
type NativeFunctionRegistry struct {
	functions map[int]*NativeCompiledFunc // Map from constant index to compiled function
	compiler  *JITCompiler
	config    JITConfig
}

// NewNativeFunctionRegistry creates a new function registry
func NewNativeFunctionRegistry(config JITConfig) *NativeFunctionRegistry {
	return &NativeFunctionRegistry{
		functions: make(map[int]*NativeCompiledFunc),
		compiler:  NewJITCompiler(config),
		config:    config,
	}
}

// CompileFunction compiles a function and stores it in the registry
func (r *NativeFunctionRegistry) CompileFunction(fn *compiler.CompiledFunction, constIdx int, constants []int64) error {
	// Check if can execute natively
	if !CanExecuteNatively(fn) {
		return fmt.Errorf("function cannot be executed natively")
	}

	// Generate native code
	cg := NewNativeCodeGenerator()
	code, err := cg.Generate(fn, constants)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Allocate executable memory
	mem, page, err := r.compiler.AllocCode(len(code))
	if err != nil {
		return fmt.Errorf("memory allocation failed: %w", err)
	}

	copy(mem, code)

	// Check if recursive (has tail calls)
	isRecursive := containsTailCall(fn.Instructions)

	r.functions[constIdx] = &NativeCompiledFunc{
		Entry:       uintptr(unsafe.Pointer(&mem[0])),
		Code:        code,
		Page:        page,
		NumParams:   fn.NumParameters,
		NumLocals:   fn.NumLocals,
		Constants:   constants,
		IsRecursive: isRecursive,
	}

	if r.config.Debug {
		fmt.Printf("[JIT] Compiled function at const[%d]: %d bytes, %d params, recursive=%v\n",
			constIdx, len(code), fn.NumParameters, isRecursive)
	}

	return nil
}

// CompileRecursiveFunction compiles a recursive function using the FibJIT compiler.
// This handles functions that contain OpRegCall (self-recursive) by transforming
// recursion into iteration. The generated code uses System V AMD64 ABI
// (args in rdi, rsi, rdx) and is called via bridge.Call1/2/3.
func (r *NativeFunctionRegistry) CompileRecursiveFunction(fn *compiler.CompiledFunction, constIdx int, constants []vm.Value) error {
	// Use the FibJIT compiler which can detect and transform recursive patterns
	fibCompiler := NewFibJITCompiler(r.config)
	code, err := fibCompiler.Compile(fn, constants, nil)
	if err != nil {
		return fmt.Errorf("recursive compilation failed: %w", err)
	}

	// Allocate executable memory
	mem, page, err := r.compiler.AllocCode(len(code))
	if err != nil {
		return fmt.Errorf("memory allocation failed: %w", err)
	}

	copy(mem, code)

	r.functions[constIdx] = &NativeCompiledFunc{
		Entry:        uintptr(unsafe.Pointer(&mem[0])),
		Code:         code,
		Page:         page,
		NumParams:    fn.NumParameters,
		NumLocals:    fn.NumLocals,
		IsRecursive:  true,
		UseBridgeABI: true, // Uses System V ABI (bridge.Call1/2/3)
	}

	if r.config.Debug {
		fmt.Printf("[JIT] Compiled recursive function at const[%d]: %d bytes, %d params (bridge ABI)\n",
			constIdx, len(code), fn.NumParameters)
	}

	return nil
}

// containsTailCall checks if bytecode contains OpRegTailCall
func containsTailCall(code []byte) bool {
	for i := 0; i < len(code); {
		op := compiler.Opcode(code[i])
		if op == compiler.OpRegTailCall {
			return true
		}
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return false
		}
		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}
	return false
}

// containsCall checks if bytecode contains OpRegCall or OpRegTailCall
func containsCall(code []byte) bool {
	for i := 0; i < len(code); {
		op := compiler.Opcode(code[i])
		if op == compiler.OpRegCall || op == compiler.OpRegTailCall {
			return true
		}
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return false
		}
		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}
	return false
}

// Get returns a compiled function by its constant index
func (r *NativeFunctionRegistry) Get(constIdx int) *NativeCompiledFunc {
	return r.functions[constIdx]
}

// Has returns true if a function is compiled at the given index
func (r *NativeFunctionRegistry) Has(constIdx int) bool {
	_, ok := r.functions[constIdx]
	return ok
}

// Execute calls a native function with arguments
func (f *NativeCompiledFunc) Execute(globals []int64, args ...int64) int64 {
	// Debug: check if entry point is valid
	if f == nil {
		fmt.Println("[JIT] ERROR: NativeCompiledFunc is nil")
		return 0
	}
	if f.Entry == 0 {
		fmt.Println("[JIT] ERROR: Entry point is 0")
		return 0
	}

	// For functions compiled with System V ABI (e.g., recursive fib),
	// use bridge.Call1/2/3 which passes args in rdi, rsi, rdx
	if f.UseBridgeABI {
		fnPtr := (*byte)(unsafe.Pointer(f.Entry))
		switch len(args) {
		case 0:
			return bridge.Call0(fnPtr)
		case 1:
			return bridge.Call1(fnPtr, args[0])
		case 2:
			return bridge.Call2(fnPtr, args[0], args[1])
		case 3:
			return bridge.Call3(fnPtr, args[0], args[1], args[2])
		default:
			// Fall back to Call3 with first 3 args
			return bridge.Call3(fnPtr, args[0], args[1], args[2])
		}
	}

	// For functions with <= 3 args, use the optimized bridge
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

		return callNativeWithArgs(f.Entry, globalsPtr, arg0, arg1, arg2)
	}

	// For functions with more args, we'd need a different bridge
	// For now, just pass what we can
	return callNativeWithGlobals(f.Entry, globals)
}

// Cleanup releases all compiled functions
func (r *NativeFunctionRegistry) Cleanup() {
	r.compiler.Cleanup()
	r.functions = make(map[int]*NativeCompiledFunc)
}

// ============================================================================
// Native Support Level Analysis
// ============================================================================

// NativeSupportLevel indicates what level of native support a function has
type NativeSupportLevel int

const (
	// SupportNone indicates the function cannot be executed natively
	SupportNone NativeSupportLevel = iota
	// SupportPureArithmetic indicates pure arithmetic/control flow (no callbacks)
	SupportPureArithmetic
	// SupportWithBuiltins indicates arithmetic + builtin calls (requires callback)
	SupportWithBuiltins
	// SupportWithCalls indicates arithmetic + builtin + function calls
	SupportWithCalls
	// SupportWithArrays indicates arithmetic + builtin + calls + array operations
	SupportWithArrays
	// SupportWithObjects indicates full support including object field access
	SupportWithObjects
)

// AnalyzeNativeSupport analyzes the bytecode to determine native support level
func AnalyzeNativeSupport(fn *compiler.CompiledFunction) NativeSupportLevel {
	code := fn.Instructions
	ip := 0
	level := SupportPureArithmetic

	for ip < len(code) {
		op := compiler.Opcode(code[ip])

		switch op {
		// Supported operations - pure arithmetic and control flow
		case compiler.OpRegLoadConst, compiler.OpRegMove,
			compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
			compiler.OpRegNeg, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
			compiler.OpRegLess, compiler.OpRegGreater, compiler.OpRegEqual,
			compiler.OpRegNotEqual, compiler.OpRegLessEqual, compiler.OpRegGreaterEqual,
			compiler.OpRegJump, compiler.OpRegJumpIfTrue, compiler.OpRegJumpIfFalse,
			compiler.OpRegReturn, compiler.OpRegNull, compiler.OpRegTrue, compiler.OpRegFalse,
			compiler.OpRegIncLocal, compiler.OpRegDecLocal,
			compiler.OpRegLoopCountAdd, compiler.OpRegLoopBodyAdd, compiler.OpRegLoopIncCheck,
			compiler.OpRegAddConst, compiler.OpRegSubConst, compiler.OpRegMulConst,
			compiler.OpRegAddLocalCheck, compiler.OpRegLoadLocal, compiler.OpRegStoreLocal,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal:
			// These are fine for native execution

		// Builtin calls - supported with callback
		case compiler.OpRegBuiltin:
			if level < SupportWithBuiltins {
				level = SupportWithBuiltins
			}

		// Function calls - supported with dispatch
		case compiler.OpRegCall, compiler.OpRegTailCall:
			if level < SupportWithCalls {
				level = SupportWithCalls
			}

		// Array operations - supported with callback
		case compiler.OpRegArray, compiler.OpRegArrayEmpty, compiler.OpRegArrayAppend,
			compiler.OpRegIndex, compiler.OpRegSetIndex:
			if level < SupportWithArrays {
				level = SupportWithArrays
			}

		// Map operations - supported with callback
		case compiler.OpRegMap, compiler.OpRegMapEmpty, compiler.OpRegMapSet:
			if level < SupportWithArrays {
				level = SupportWithArrays
			}

		// Object operations - supported with inline caching
		case compiler.OpRegGetField, compiler.OpRegSetField,
			compiler.OpRegGetMethod, compiler.OpRegCallMethod:
			if level < SupportWithObjects {
				level = SupportWithObjects
			}

		// Unsupported - requires full VM context
		case compiler.OpRegClosure, compiler.OpRegLoadFree, compiler.OpRegStoreFree,
			compiler.OpRegPush, compiler.OpRegPop,
			compiler.OpRegClass, compiler.OpRegNew,
			compiler.OpRegThrow, compiler.OpRegPushHandler, compiler.OpRegPopHandler,
			compiler.OpRegLoadModule, compiler.OpRegGetExport, compiler.OpRegSetExport,
			compiler.OpRegIterKey, compiler.OpRegIterValue,
			compiler.OpRegLoadFunc:
			return SupportNone

		default:
			// Unknown opcode - be conservative
			return SupportNone
		}

		// Skip operands
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			return SupportNone
		}
		operandWidth := 0
		for _, w := range def.OperandWidths {
			operandWidth += w
		}
		ip += operandWidth + 1
	}

	return level
}

// CanExecuteNativelyWithBuiltins checks if a function can be executed natively with builtin callback support
func CanExecuteNativelyWithBuiltins(fn *compiler.CompiledFunction) bool {
	return AnalyzeNativeSupport(fn) >= SupportWithBuiltins
}

// ============================================================================
// Builtin Callback Support
// ============================================================================

// JITCallbackContext holds context for JIT callbacks to Go
type JITCallbackContext struct {
	globals      []vm.Value
	constants    []vm.Value
	builtins     []*objects.Builtin
	nativeFuncs  map[int]*NativeCompiledFunc // constIdx -> native function
	registers    []vm.Value                  // Current frame registers
	objects      []objects.Object            // Objects created by callbacks (for GC visibility)
}

// globalJITContext is the global callback context
var globalJITContext JITCallbackContext

// InitJITCallbackContext initializes the global JIT callback context
func InitJITCallbackContext(globals, constants []vm.Value) {
	globalJITContext.globals = globals
	globalJITContext.constants = constants
}

// SetNativeFunctionRegistry sets the native function registry for callbacks
func SetNativeFunctionRegistry(funcs map[int]*NativeCompiledFunc) {
	globalJITContext.nativeFuncs = funcs
}

// SetCurrentRegisters sets the current frame registers for callbacks
func SetCurrentRegisters(regs []vm.Value) {
	globalJITContext.registers = regs
}

// CallBuiltinFromNative is called from native code to execute a builtin function
// This is exported for use by the JIT callback mechanism
func CallBuiltinFromNative(builtinIdx, numArgs int, argsPtr *int64) int64 {
	if builtinIdx < 0 || builtinIdx >= len(globalJITContext.builtins) {
		// Lazy initialize builtins array
		if builtinIdx < 256 {
			if globalJITContext.builtins == nil {
				globalJITContext.builtins = make([]*objects.Builtin, 256)
			}
		}
	}

	builtin := globalJITContext.builtins[builtinIdx]
	if builtin == nil {
		// Lazy load builtin
		builtin = getBuiltin(builtinIdx)
		if builtin == nil {
			return 0
		}
		globalJITContext.builtins[builtinIdx] = builtin
	}

	// Convert args from int64 array to objects
	args := make([]objects.Object, numArgs)
	argsSlice := unsafe.Slice(argsPtr, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = &objects.Int{Value: argsSlice[i]}
	}

	// Call the builtin
	result := builtin.Fn(args...)

	// Convert result back to int64
	switch r := result.(type) {
	case *objects.Int:
		return r.Value
	case *objects.Bool:
		if r.Value {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// getBuiltin returns a builtin by index (implemented in vm/builtins.go)
func getBuiltin(idx int) *objects.Builtin {
	return vm.GetBuiltinByIndex(idx)
}

// GetBuiltinCallbackPtr returns the function pointer for builtin callbacks
// Note: This requires unsafe.Pointer to get the function address
func GetBuiltinCallbackPtr() uintptr {
	// Return a marker value - actual callback will be set up differently
	// For now, we'll use a different mechanism for callbacks
	return 0
}

// ============================================================================
// Function Call Callback Support (Phase 2)
// ============================================================================

// FunctionDispatchKind indicates what kind of function was found
type FunctionDispatchKind int

const (
	DispatchInvalid FunctionDispatchKind = iota
	DispatchNative  // Has native compiled code
	DispatchClosure // Closure (requires interpreter)
	DispatchFunc    // Regular compiled function (no native version)
	DispatchBuiltin // Builtin function
)

// FunctionDispatchInfo contains information about a function to dispatch to
type FunctionDispatchInfo struct {
	Kind        FunctionDispatchKind
	Entry       uintptr // For native functions
	ConstIdx    int     // Index in constants table
	NumParams   int     // Number of parameters
	NumFree     int     // Number of free variables (for closures)
}

// CallFunctionFromNative is called from native code to dispatch a function call
// funcReg: the register number containing the function value
// numArgs: number of arguments
// argsPtr: pointer to args array (int64 values)
// Returns the result of the function call
func CallFunctionFromNative(funcReg, numArgs int, argsPtr *int64) int64 {
	// Get the function value from registers
	if globalJITContext.registers == nil || funcReg >= len(globalJITContext.registers) {
		return 0
	}

	fnValue := globalJITContext.registers[funcReg]
	if fnValue.IsNull() {
		return 0
	}

	// Check if it's a native function (check constIdx first)
	// The function might be loaded from constants, so check if it has a native version
	dispatchInfo := resolveFunctionDispatch(fnValue, funcReg)
	if dispatchInfo.Kind == DispatchInvalid {
		return 0
	}

	// Convert args from int64 array to vm.Value
	argsSlice := unsafe.Slice(argsPtr, numArgs)
	args := make([]vm.Value, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = vm.NewInt(argsSlice[i])
	}

	switch dispatchInfo.Kind {
	case DispatchNative:
		// Call native function directly
		if dispatchInfo.Entry == 0 {
			return 0
		}

		// Get globals as int64 array
		globals := make([]int64, len(globalJITContext.globals))
		for i, g := range globalJITContext.globals {
			if g.IsInt() {
				globals[i], _ = g.ToInt()
			}
		}

		// Call via bridge
		nativeFunc := &NativeCompiledFunc{
			Entry:      dispatchInfo.Entry,
			NumParams:  dispatchInfo.NumParams,
			UseBridgeABI: true,
		}
		return nativeFunc.Execute(globals, int64SliceFromValueSlice(args)...)

	case DispatchBuiltin:
		// Delegate to builtin callback
		// Find builtin index - this is stored differently
		return 0 // Builtins are handled by CallBuiltinFromNative

	case DispatchClosure, DispatchFunc:
		// Fall back to interpreter - for now return 0
		// In a full implementation, we'd set up a frame and call the interpreter
		return 0

	default:
		return 0
	}
}

// resolveFunctionDispatch determines how to dispatch a function call
func resolveFunctionDispatch(fnValue vm.Value, funcReg int) FunctionDispatchInfo {
	info := FunctionDispatchInfo{Kind: DispatchInvalid}

	// Check if it's a compiled function
	if compiledFn := fnValue.GetCompiledFunction(); compiledFn != nil {
		info.NumParams = compiledFn.NumParameters

		// Check if we have a native version
		// Look up by checking if any constant matches this function
		for idx, c := range globalJITContext.constants {
			cf := c.GetCompiledFunction()
			if cf != nil && cf == compiledFn {
				info.ConstIdx = idx
				if nativeFn, ok := globalJITContext.nativeFuncs[idx]; ok && nativeFn != nil {
					info.Kind = DispatchNative
					info.Entry = nativeFn.Entry
					return info
				}
				break
			}
		}

		info.Kind = DispatchFunc
		return info
	}

	// Check if it's a closure
	if closure := fnValue.GetClosure(); closure != nil {
		info.Kind = DispatchClosure
		if closure.Fn != nil {
			info.NumParams = closure.Fn.NumParameters
			info.NumFree = len(closure.FreeVars)
		}
		return info
	}

	// Check if it's a builtin (shouldn't happen here, but handle it)
	if fnValue.IsObject() {
		obj := fnValue.ToObject()
		if _, ok := obj.(*objects.Builtin); ok {
			info.Kind = DispatchBuiltin
			return info
		}
	}

	return info
}

// int64SliceFromValueSlice converts vm.Value slice to int64 slice
func int64SliceFromValueSlice(values []vm.Value) []int64 {
	result := make([]int64, len(values))
	for i, v := range values {
		if v.IsInt() {
			result[i], _ = v.ToInt()
		}
	}
	return result
}

// CanExecuteNativelyWithCalls checks if a function can be executed natively with function call support
func CanExecuteNativelyWithCalls(fn *compiler.CompiledFunction) bool {
	return AnalyzeNativeSupport(fn) >= SupportWithCalls
}

// CanExecuteNativelyWithArrays checks if a function can be executed natively with array/map support
func CanExecuteNativelyWithArrays(fn *compiler.CompiledFunction) bool {
	return AnalyzeNativeSupport(fn) >= SupportWithArrays
}

// ============================================================================
// Collection Operations Callback Support (Phase 3)
// ============================================================================

// CallCollectionFromNative is called from native code to perform collection operations
// opKind: the operation kind (CollectionOpKind, defined in native_codegen.go)
// argsPtr: pointer to args array (operation-specific)
// Returns the result (for get operations) or 0
func CallCollectionFromNative(opKind, numArgs int, argsPtr *int64) int64 {
	argsSlice := unsafe.Slice(argsPtr, numArgs)

	switch CollectionOpKind(opKind) {
	case OpArrayCreate:
		// Create array from elements
		// args: element values
		elements := make([]objects.Object, numArgs)
		for i := 0; i < numArgs; i++ {
			elements[i] = &objects.Int{Value: argsSlice[i]}
		}
		arr := &objects.Array{Elements: elements}
		// Store in context for GC visibility
		globalJITContext.objects = append(globalJITContext.objects, arr)
		// Return a handle/index to the object
		return int64(len(globalJITContext.objects) - 1)

	case OpArrayEmpty:
		// Create empty array
		arr := &objects.Array{Elements: []objects.Object{}}
		globalJITContext.objects = append(globalJITContext.objects, arr)
		return int64(len(globalJITContext.objects) - 1)

	case OpArrayAppend:
		// Append element to array
		// args[0] = array handle, args[1] = element
		if numArgs < 2 {
			return 0
		}
		handle := int(argsSlice[0])
		if handle < 0 || handle >= len(globalJITContext.objects) {
			return 0
		}
		arr, ok := globalJITContext.objects[handle].(*objects.Array)
		if !ok {
			return 0
		}
		elem := &objects.Int{Value: argsSlice[1]}
		arr.Elements = append(arr.Elements, elem)
		return argsSlice[0] // Return same handle

	case OpArrayGet:
		// Get element from array
		// args[0] = array handle, args[1] = index
		if numArgs < 2 {
			return 0
		}
		handle := int(argsSlice[0])
		index := int(argsSlice[1])
		if handle < 0 || handle >= len(globalJITContext.objects) {
			return 0
		}
		arr, ok := globalJITContext.objects[handle].(*objects.Array)
		if !ok || index < 0 || index >= len(arr.Elements) {
			return 0
		}
		if elem, ok := arr.Elements[index].(*objects.Int); ok {
			return elem.Value
		}
		return 0

	case OpArraySet:
		// Set element in array
		// args[0] = array handle, args[1] = index, args[2] = value
		if numArgs < 3 {
			return 0
		}
		handle := int(argsSlice[0])
		index := int(argsSlice[1])
		if handle < 0 || handle >= len(globalJITContext.objects) {
			return 0
		}
		arr, ok := globalJITContext.objects[handle].(*objects.Array)
		if !ok || index < 0 || index >= len(arr.Elements) {
			return 0
		}
		arr.Elements[index] = &objects.Int{Value: argsSlice[2]}
		return argsSlice[0]

	case OpMapCreate:
		// Create map from key-value pairs
		// args: key0, val0, key1, val1, ...
		pairs := make(map[objects.HashKey]objects.MapPair)
		for i := 0; i+1 < numArgs; i += 2 {
			key := &objects.Int{Value: argsSlice[i]}
			val := &objects.Int{Value: argsSlice[i+1]}
			pairs[key.HashKey()] = objects.MapPair{Key: key, Value: val}
		}
		m := &objects.Map{Pairs: pairs}
		globalJITContext.objects = append(globalJITContext.objects, m)
		return int64(len(globalJITContext.objects) - 1)

	case OpMapEmpty:
		// Create empty map
		m := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
		globalJITContext.objects = append(globalJITContext.objects, m)
		return int64(len(globalJITContext.objects) - 1)

	case OpMapSet:
		// Set key-value in map
		// args[0] = map handle, args[1] = key, args[2] = value
		if numArgs < 3 {
			return 0
		}
		handle := int(argsSlice[0])
		if handle < 0 || handle >= len(globalJITContext.objects) {
			return 0
		}
		m, ok := globalJITContext.objects[handle].(*objects.Map)
		if !ok {
			return 0
		}
		key := &objects.Int{Value: argsSlice[1]}
		val := &objects.Int{Value: argsSlice[2]}
		m.Pairs[key.HashKey()] = objects.MapPair{Key: key, Value: val}
		return argsSlice[0]

	case OpMapGet:
		// Get value from map
		// args[0] = map handle, args[1] = key
		if numArgs < 2 {
			return 0
		}
		handle := int(argsSlice[0])
		if handle < 0 || handle >= len(globalJITContext.objects) {
			return 0
		}
		m, ok := globalJITContext.objects[handle].(*objects.Map)
		if !ok {
			return 0
		}
		key := &objects.Int{Value: argsSlice[1]}
		if pair, ok := m.Pairs[key.HashKey()]; ok {
			if val, ok := pair.Value.(*objects.Int); ok {
				return val.Value
			}
		}
		return 0

	default:
		return 0
	}
}
