// pkg/jit/native_executor.go
// Native JIT executor that runs compiled x86-64 code directly
package jit

import (
	"fmt"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
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
// - No function calls (OpRegCall, OpRegTailCall)
// - No builtin calls
// - No array/map operations
// - Only: LoadConst, Move, Add, Sub, Mul, Div, Mod, Neg
//        Less, Greater, Equal, NotEqual, LessEqual, GreaterEqual
//        Jump, JumpIfTrue, JumpIfFalse, Return, Null, True, False
//        LoadGlobal, StoreGlobal (with globals pointer passed as argument)
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
			// These are fine for native execution (globals will be passed as first arg)

		// Unsupported - requires VM context
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
