// +build amd64,!windows

// pkg/jit/call.go
// Call trampoline for JIT <-> Go transitions
package jit

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

// CallTrampoline manages transitions between JIT code and Go interpreter
type CallTrampoline struct {
	// Function registry: function pointer -> function info
	funcRegistry sync.Map // map[uintptr]*FuncInfo

	// Compiled function cache
	compiledFuncs sync.Map // map[uint64]*CompiledFunc

	// Global JIT compiler instance
	jit *JITCompiler

	// Constants and globals for the current execution context
	constants []vm.Value
	globals   []vm.Value
}

// FuncInfo contains information about a callable function
type FuncInfo struct {
	Fn          *compiler.CompiledFunction
	IsJITed     bool
	Entry       uintptr // JIT entry point if compiled
	NumParams   int
	NumLocals   int
	Constants   []vm.Value
	FreeVars    []vm.Value
}

// CallFrame represents a call frame for JIT <-> Go transitions
type CallFrame struct {
	// Registers
	Registers [256]vm.Value

	// Return address (for JIT calls)
	ReturnAddr uintptr

	// Return value
	ReturnVal vm.Value

	// Function being executed
	Fn *FuncInfo
}

// Global trampoline instance
var globalTrampoline = &CallTrampoline{
	funcRegistry: sync.Map{},
	compiledFuncs: sync.Map{},
}

// GetTrampoline returns the global trampoline instance
func GetTrampoline() *CallTrampoline {
	return globalTrampoline
}

// SetJIT sets the JIT compiler instance
func (t *CallTrampoline) SetJIT(jit *JITCompiler) {
	t.jit = jit
}

// SetContext sets the constants and globals for the current execution
func (t *CallTrampoline) SetContext(constants, globals []vm.Value) {
	t.constants = constants
	t.globals = globals
}

// RegisterFunction registers a function for JIT calls
func (t *CallTrampoline) RegisterFunction(fn *compiler.CompiledFunction, constants []vm.Value) *FuncInfo {
	info := &FuncInfo{
		Fn:        fn,
		IsJITed:   false,
		NumParams: fn.NumParameters,
		NumLocals: fn.NumLocals,
		Constants: constants,
	}

	// Use the function pointer as key
	key := uintptr(unsafe.Pointer(fn))
	t.funcRegistry.Store(key, info)

	return info
}

// GetFunction retrieves a registered function
func (t *CallTrampoline) GetFunction(fn *compiler.CompiledFunction) *FuncInfo {
	key := uintptr(unsafe.Pointer(fn))
	if info, ok := t.funcRegistry.Load(key); ok {
		return info.(*FuncInfo)
	}
	return nil
}

// CompileFunction attempts to JIT compile a function
func (t *CallTrampoline) CompileFunction(fn *compiler.CompiledFunction, constants, globals []vm.Value) (*CompiledFunc, error) {
	if t.jit == nil {
		return nil, fmt.Errorf("JIT compiler not initialized")
	}

	// Check cache first
	hash := hashBytecode(fn.Instructions)
	if cf, ok := t.compiledFuncs.Load(hash); ok {
		return cf.(*CompiledFunc), nil
	}

	// Compile
	cf, err := t.jit.Compile(fn, constants, globals)
	if err != nil {
		return nil, err
	}

	// Cache the result
	t.compiledFuncs.Store(hash, cf)

	// Update function info
	if info := t.GetFunction(fn); info != nil {
		info.IsJITed = true
		info.Entry = cf.Entry
	}

	return cf, nil
}

// CallFromJIT is called from JIT code to execute a function
// This is the Go-side handler for JIT function calls
//
//export CallFromJIT
func CallFromJIT(fnPtr uintptr, args *vm.Value, numArgs int) int64 {
	// Look up function info
	infoIface, ok := globalTrampoline.funcRegistry.Load(fnPtr)
	if !ok {
		// Log error and return 0 instead of panicking
		// This prevents the entire program from crashing on unknown function
		fmt.Printf("[JIT] CallFromJIT: unknown function at %x, returning 0\n", fnPtr)
		return 0
	}
	info := infoIface.(*FuncInfo)

	// Create a temporary VM to execute the function
	// This is the interpreter fallback path
	result := executeFunctionInInterpreter(info, args, numArgs)

	// Return the result as int64 (for simple numeric results)
	if result.IsInt() {
		return result.GetInt()
	}
	if result.IsFloat() {
		return int64(result.GetFloat())
	}
	return 0
}

// executeFunctionInInterpreter runs a function in the interpreter
func executeFunctionInInterpreter(info *FuncInfo, argsPtr *vm.Value, numArgs int) vm.Value {
	// Create a new frame for the function
	frame := vm.NewRegFrame(info.Fn)
	frame.Constants = info.Constants

	// Convert pointer to slice for safe access
	args := (*[8]vm.Value)(unsafe.Pointer(argsPtr))

	// Copy arguments to registers
	for i := 0; i < numArgs && i < 8; i++ {
		frame.Registers[i] = args[i]
	}

	// Create a minimal VM context
	// For now, we use a simple interpreter loop
	result := interpretFunction(frame, info.Fn, info.Constants)

	return result
}

// interpretFunction interprets a single function
func interpretFunction(frame *vm.RegFrame, fn *compiler.CompiledFunction, constants []vm.Value) vm.Value {
	code := fn.Instructions
	ip := 0
	regs := frame.Registers[:]

	// Safety: iteration limit to prevent infinite loops
	maxIterations := len(code) * 2
	iterations := 0

	for ip < len(code) {
		iterations++
		if iterations > maxIterations {
			// Prevent infinite loop from corrupted bytecode
			return vm.ValueNull
		}

		op := compiler.Opcode(code[ip])

		// Safety: check instruction length before processing
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			ip++
			continue
		}
		instrLen := 1 + len(def.OperandWidths)
		if ip+instrLen > len(code) {
			// Truncated instruction, abort
			return vm.ValueNull
		}

		switch op {
		case compiler.OpRegLoadConst:
			dst := code[ip+1]
			idx := int(code[ip+2])<<8 | int(code[ip+3])
			if idx < len(constants) {
				regs[dst] = constants[idx]
			} else {
				regs[dst] = vm.ValueNull
			}
			ip += 4

		case compiler.OpRegMove:
			dst := code[ip+1]
			src := code[ip+2]
			regs[dst] = regs[src]
			ip += 3

		case compiler.OpRegAdd:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Add(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegSub:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Sub(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegMul:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Mul(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegDiv:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Div(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegLess:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			less, _ := regs[src1].Less(regs[src2])
			if less {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegLessEqual:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result := regs[src1].LessEqual(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegGreater:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			greater, _ := regs[src1].Greater(regs[src2])
			if greater {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegEqual:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			eq, _ := regs[src1].Equal(regs[src2])
			if eq {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegJumpIfFalse:
			condReg := code[ip+1]
			offset := int(int16(int(code[ip+2])<<8 | int(code[ip+3])))
			if !regs[condReg].IsTruthy() {
				ip += offset
			} else {
				ip += 4
			}

		case compiler.OpRegJump:
			offset := int(int16(int(code[ip+2])<<8 | int(code[ip+3])))
			ip += offset

		case compiler.OpRegReturn:
			retReg := code[ip+1]
			return regs[retReg]

		case compiler.OpRegNull:
			dst := code[ip+1]
			regs[dst] = vm.ValueNull
			ip += 2

		case compiler.OpRegTrue:
			dst := code[ip+1]
			regs[dst] = vm.ValueTrue
			ip += 2

		case compiler.OpRegFalse:
			dst := code[ip+1]
			regs[dst] = vm.ValueFalse
			ip += 2

		default:
			// Skip unknown opcodes
			def, err := compiler.Lookup(byte(op))
			if err != nil {
				ip++
				continue
			}
			ip++
			for _, w := range def.OperandWidths {
				ip += w
			}
		}
	}

	// Return last computed value
	return regs[compiler.ReturnRegister]
}

// TrampolineEntry is the assembly stub that bridges JIT to Go
// This is implemented in assembly for each platform
var TrampolineEntry uintptr

// CallJITFunction calls a JIT-compiled function from Go
func CallJITFunction(entry uintptr, args []vm.Value) vm.Value {
	if len(args) > 8 {
		// Too many arguments for register passing
		return vm.ValueNull
	}

	// For simple cases, we can call directly
	// The actual implementation depends on the platform
	result := callJITNative(entry, args)
	return result
}

// callJITNative is implemented in platform-specific assembly
func callJITNative(entry uintptr, args []vm.Value) vm.Value {
	// This is a Go implementation for platforms without assembly
	// In practice, this would be a direct call to the JIT code

	// For now, return null to indicate the function was called
	// Real implementation would save registers, call entry, restore
	return vm.ValueNull
}

// ============================================================================
// JIT Function Call Infrastructure
// ============================================================================

// CallContext holds the execution context for JIT function calls
type CallContext struct {
	// Registers for current frame
	Registers [256]vm.Value

	// Constants and globals
	Constants []vm.Value
	Globals   []vm.Value

	// Function being executed
	Function *compiler.CompiledFunction

	// Call depth for recursion limit
	Depth int
}

// JITCallFrame represents a saved frame for recursive JIT calls
type JITCallFrame struct {
	Registers   [256]vm.Value
	ReturnIP    int
	ReturnValue vm.Value
}

// Global call stack for JIT recursive calls
var jitCallStack = struct {
	sync.Mutex
	Frames []JITCallFrame
}{
	Frames: make([]JITCallFrame, 0, 1024),
}

// JITCallFunc executes a function call from JIT code
// This is the main entry point for JIT -> Interpreter transitions
func JITCallFunc(fnPtr uintptr, argsPtr *vm.Value, numArgs int, constantsPtr *vm.Value, numConstants int) int64 {
	// Look up function in registry
	infoIface, ok := globalTrampoline.funcRegistry.Load(fnPtr)
	if !ok {
		// Try to interpret from bytecode directly
		return 0
	}
	info := infoIface.(*FuncInfo)

	// Create a slice from the pointer
	args := unsafe.Slice(argsPtr, numArgs)
	constants := unsafe.Slice(constantsPtr, numConstants)

	// Execute the function
	result := executeFunctionInJITContext(info, args, constants)

	// Return int64 result
	if result.IsInt() {
		val, _ := result.ToInt()
		return val
	}
	return 0
}

// executeFunctionInJITContext executes a function with JIT calling convention
func executeFunctionInJITContext(info *FuncInfo, args []vm.Value, constants []vm.Value) vm.Value {
	// Create a new frame
	frame := vm.NewRegFrame(info.Fn)
	frame.Constants = constants

	// Copy arguments to registers (R0-R7)
	for i := 0; i < len(args) && i < 8; i++ {
		frame.Registers[i] = args[i]
	}

	// Execute using a minimal interpreter
	return interpretFunctionFast(frame, info.Fn, constants)
}

// interpretFunctionFast is an optimized interpreter for JIT-called functions
func interpretFunctionFast(frame *vm.RegFrame, fn *compiler.CompiledFunction, constants []vm.Value) vm.Value {
	code := fn.Instructions
	ip := 0
	regs := frame.Registers[:]

	// Safety: iteration limit to prevent infinite loops
	maxIterations := len(code) * 2
	iterations := 0

	for ip < len(code) {
		iterations++
		if iterations > maxIterations {
			return vm.ValueNull
		}

		op := compiler.Opcode(code[ip])

		// Safety: check instruction length before processing
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			ip++
			continue
		}
		instrLen := 1 + len(def.OperandWidths)
		if ip+instrLen > len(code) {
			return vm.ValueNull
		}

		switch op {
		case compiler.OpRegLoadConst:
			dst := code[ip+1]
			idx := int(code[ip+2])<<8 | int(code[ip+3])
			if idx < len(constants) {
				regs[dst] = constants[idx]
			} else {
				regs[dst] = vm.ValueNull
			}
			ip += 4

		case compiler.OpRegMove:
			dst := code[ip+1]
			src := code[ip+2]
			regs[dst] = regs[src]
			ip += 3

		case compiler.OpRegAdd:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Add(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegSub:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Sub(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegMul:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Mul(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegDiv:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			result, _ := regs[src1].Div(regs[src2])
			regs[dst] = result
			ip += 4

		case compiler.OpRegLess:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			less, _ := regs[src1].Less(regs[src2])
			if less {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegLessEqual:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			regs[dst] = regs[src1].LessEqual(regs[src2])
			ip += 4

		case compiler.OpRegGreater:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			greater, _ := regs[src1].Greater(regs[src2])
			if greater {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegEqual:
			dst := code[ip+1]
			src1 := code[ip+2]
			src2 := code[ip+3]
			eq, _ := regs[src1].Equal(regs[src2])
			if eq {
				regs[dst] = vm.ValueTrue
			} else {
				regs[dst] = vm.ValueFalse
			}
			ip += 4

		case compiler.OpRegJumpIfFalse:
			condReg := code[ip+1]
			offset := int(int16(int(code[ip+2])<<8 | int(code[ip+3])))
			if !regs[condReg].IsTruthy() {
				ip += offset
			} else {
				ip += 4
			}

		case compiler.OpRegJump:
			offset := int(int16(int(code[ip+2])<<8 | int(code[ip+3])))
			ip += offset

		case compiler.OpRegReturn:
			retReg := code[ip+1]
			return regs[retReg]

		case compiler.OpRegNull:
			dst := code[ip+1]
			regs[dst] = vm.ValueNull
			ip += 2

		case compiler.OpRegTrue:
			dst := code[ip+1]
			regs[dst] = vm.ValueTrue
			ip += 2

		case compiler.OpRegFalse:
			dst := code[ip+1]
			regs[dst] = vm.ValueFalse
			ip += 2

		case compiler.OpRegCall:
			// Handle recursive calls
			funcReg := code[ip+1]
			numArgs := code[ip+2]
			result := handleJITRecursiveCall(regs, int(funcReg), int(numArgs), constants)
			regs[compiler.ReturnRegister] = result
			ip += 3

		case compiler.OpRegTailCall:
			// Tail call: reuse frame
			funcReg := code[ip+1]
			numArgs := code[ip+2]
			// For self-recursive tail calls, update args and restart
			for i := 0; i < int(numArgs) && i < 8; i++ {
				regs[i] = regs[i+8] // Args are in R8-R15 for tail calls
			}
			ip = 0 // Restart function
			_ = funcReg

		default:
			// Skip unknown opcodes
			def, err := compiler.Lookup(byte(op))
			if err != nil {
				ip++
				continue
			}
			ip++
			for _, w := range def.OperandWidths {
				ip += w
			}
		}
	}

	// Return last computed value
	return regs[compiler.ReturnRegister]
}

// handleJITRecursiveCall handles recursive function calls from JIT
// MaxRecursionDepth is the maximum allowed recursion depth to prevent stack overflow
const MaxRecursionDepth = 1000

// recursionDepth tracks current recursion depth (thread-local via sync)
var recursionDepth struct {
	sync.Mutex
	depth int
}

func handleJITRecursiveCall(regs []vm.Value, funcReg, numArgs int, constants []vm.Value) vm.Value {
	// Check recursion depth
	recursionDepth.Lock()
	recursionDepth.depth++
	if recursionDepth.depth > MaxRecursionDepth {
		recursionDepth.depth--
		recursionDepth.Unlock()
		// Log warning about recursion limit exceeded
		fmt.Printf("[JIT WARNING] Recursion depth limit (%d) exceeded, returning error to prevent stack overflow\n", MaxRecursionDepth)
		// Return Error object so caller can detect and handle it
		return vm.NewObject(&objects.Error{
			Message: fmt.Sprintf("recursion depth limit (%d) exceeded", MaxRecursionDepth),
		})
	}
	recursionDepth.Unlock()

	// Ensure we decrement depth on return
	defer func() {
		recursionDepth.Lock()
		recursionDepth.depth--
		recursionDepth.Unlock()
	}()

	// Get function from register
	fn := regs[funcReg]

	// Check for compiled function or closure
	var cf *compiler.CompiledFunction
	if fn.IsCompiledFunction() {
		cf = fn.GetCompiledFunction()
	} else if fn.IsClosure() {
		closure := fn.GetClosure()
		if closure != nil {
			cf = closure.Fn
		}
	}

	if cf == nil {
		return vm.ValueNull
	}

	// Create new frame with arguments
	newFrame := vm.NewRegFrame(cf)
	newFrame.Constants = constants

	// Copy arguments from caller's registers
	for i := 0; i < numArgs && i < 8; i++ {
		newFrame.Registers[i] = regs[i]
	}

	// Recursively execute
	return interpretFunctionFast(newFrame, cf, constants)
}
