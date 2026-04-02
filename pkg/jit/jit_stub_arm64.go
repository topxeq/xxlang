//go:build !amd64 && !arm64
// +build !amd64,!arm64

// pkg/jit/jit_stub_arm64.go
// Stub implementation for platforms without JIT support
// JIT compilation is not supported on non-amd64 and non-arm64 platforms
// This allows the code to compile and run, but JIT features are disabled

package jit

import (
	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// CompiledFunc represents a JIT-compiled function (stub)
type CompiledFunc struct {
	Entry     uintptr
	Page      *CodePage
	Size      int
	Hash      uint64
	NumRegs   int
	NumParams int
}

// CodePage represents a page of executable memory (stub)
type CodePage struct {
	Data []byte
	Used int
}

// JITCompiler handles JIT compilation (stub)
type JITCompiler struct {
	config JITConfig
}

// NewJITCompiler creates a new JIT compiler stub
func NewJITCompiler(config JITConfig) *JITCompiler {
	return &JITCompiler{config: config}
}

// GetStats returns empty statistics
func (j *JITCompiler) GetStats() JITStats {
	return JITStats{}
}

// GetCompiled returns nil (no compiled functions)
func (j *JITCompiler) GetCompiled(fn *compiler.CompiledFunction) *CompiledFunc {
	return nil
}

// ShouldCompile always returns false
func (j *JITCompiler) ShouldCompile(fn *compiler.CompiledFunction) bool {
	return false
}

// Compile returns an error (not supported)
func (j *JITCompiler) Compile(fn *compiler.CompiledFunction, constants []vm.Value, globals []vm.Value) (*CompiledFunc, error) {
	return nil, nil
}

// Cleanup does nothing
func (j *JITCompiler) Cleanup() {}

// NativeExecutor executes JIT-compiled native code (stub)
type NativeExecutor struct {
	config JITConfig
}

// NewNativeExecutor creates a new native executor stub
func NewNativeExecutor(config JITConfig) *NativeExecutor {
	return &NativeExecutor{config: config}
}

// CanExecuteNatively always returns false on non-amd64
func CanExecuteNatively(fn *compiler.CompiledFunction) bool {
	return false
}

// ExecuteFunction returns an error (not supported)
func (ne *NativeExecutor) ExecuteFunction(fn *compiler.CompiledFunction, constants []vm.Value, globals []int64) (int64, error) {
	return 0, nil
}

// Cleanup does nothing
func (ne *NativeExecutor) Cleanup() {}

// NativeFunction represents a natively compiled function (stub)
type NativeFunction struct {
	Code      []byte
	NumParams int
}

// Execute returns 0 (stub)
func (nf *NativeFunction) Execute(globals []int64, args ...int64) int64 {
	return 0
}

// NativeFunctionRegistry manages compiled native functions (stub)
type NativeFunctionRegistry struct {
	config JITConfig
}

// NewNativeFunctionRegistry creates a new registry stub
func NewNativeFunctionRegistry(config JITConfig) *NativeFunctionRegistry {
	return &NativeFunctionRegistry{config: config}
}

// Get returns nil (no functions)
func (r *NativeFunctionRegistry) Get(idx int) *NativeFunction {
	return nil
}

// CompileFunction does nothing (stub)
func (r *NativeFunctionRegistry) CompileFunction(fn *compiler.CompiledFunction, idx int, constants []int64) error {
	return nil
}

// CompileRecursiveFunction does nothing (stub)
func (r *NativeFunctionRegistry) CompileRecursiveFunction(fn *compiler.CompiledFunction, idx int, constants []vm.Value) error {
	return nil
}

// Cleanup does nothing
func (r *NativeFunctionRegistry) Cleanup() {}

// AnalyzeNativeSupport returns SupportNone for non-amd64
func AnalyzeNativeSupport(fn *compiler.CompiledFunction) int {
	return 0 // SupportNone
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

// CallNativeCode stub
func CallNativeCode(codePtr uintptr, arg int64) int64 {
	return 0
}

// SimpleNativeCall stub
func SimpleNativeCall(code []byte, arg int64) int64 {
	return 0
}

// AllocCode stub
func (j *JITCompiler) AllocCode(size int) ([]byte, *CodePage, error) {
	return nil, nil, nil
}

// RecordExecution stub
func (j *JITCompiler) RecordExecution(fn *compiler.CompiledFunction) bool {
	return false
}

// hashBytecode stub
func hashBytecode(code []byte) uint64 {
	return 0
}
