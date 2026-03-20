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
// - No global variable access
// - No array/map operations
// - Only: LoadConst, Move, Add, Sub, Mul, Div, Mod, Neg
//        Less, Greater, Equal, NotEqual, LessEqual, GreaterEqual
//        Jump, JumpIfTrue, JumpIfFalse, Return, Null, True, False
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
			compiler.OpRegAddLocalCheck, compiler.OpRegLoadLocal, compiler.OpRegStoreLocal:
			// These are fine for native execution

		// Unsupported - requires VM context
		case compiler.OpRegCall, compiler.OpRegTailCall,
			compiler.OpRegLoadGlobal, compiler.OpRegStoreGlobal,
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

// ExecuteFunction compiles and executes a function natively
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value) (int64, error) {
	// Check if can execute natively
	if !CanExecuteNatively(fn) {
		return 0, fmt.Errorf("function cannot be executed natively")
	}

	// Check cache
	cf := ne.compiler.GetCompiled(fn)
	if cf != nil {
		return cf.Execute(), nil
	}

	// Compile
	cf, err := ne.compiler.Compile(fn, constants, nil)
	if err != nil {
		return 0, fmt.Errorf("compilation failed: %w", err)
	}

	// Execute
	return cf.Execute(), nil
}

// ExecuteBytecode executes a bytecode program natively
// Returns the result as an int64 (for numeric results)
func (ne *NativeExecutor) ExecuteBytecode(bytecode *compiler.Bytecode) (int64, error) {
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

	return ne.ExecuteFunction(mainFn, constants)
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

	// Use optimized code generator
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

// Run executes the native program
func (p *NativeProgram) Run() int64 {
	fn := *(*func() int64)(unsafe.Pointer(&p.entry))
	return fn()
}
