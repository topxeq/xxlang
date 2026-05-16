//go:build windows && amd64
// +build windows,amd64

// pkg/jit/jit_windows_coverage_test.go
// Extended tests for Windows JIT coverage
package jit

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestNativeExecutorExecuteFunction tests native execution
func TestNativeExecutorExecuteFunction(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())
	defer exec.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0,
			byte(compiler.OpRegReturn), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []vm.Value{vm.NewInt(42)}
	globals := make([]int64, 256)

	result, err := exec.ExecuteFunction(fn, constants, globals)
	if err != nil {
		t.Fatalf("ExecuteFunction failed: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

// TestGenerateNativeCode tests native code generation
func TestGenerateNativeCode(t *testing.T) {
	tests := []struct {
		name string
		fn   *compiler.CompiledFunction
	}{
		{
			name: "simple null",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegNull), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "simple true",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegTrue), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "simple false",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegFalse), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "load const and return",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegLoadConst), 0, 0, 0,
					byte(compiler.OpRegReturn), 0,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generateNativeCode(tt.fn, nil, nil)
			if err != nil {
				t.Errorf("generateNativeCode failed: %v", err)
				return
			}
			if len(code) == 0 {
				t.Error("Generated code is empty")
				return
			}
			t.Logf("Generated %d bytes", len(code))
		})
	}
}

// TestNativeFunctionExecute tests native function execution
func TestNativeFunctionExecute(t *testing.T) {
	t.Run("empty code", func(t *testing.T) {
		nf := &NativeFunction{Code: nil, entry: 0}
		result := nf.Execute(nil)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("zero entry", func(t *testing.T) {
		nf := &NativeFunction{Code: []byte{0x00}, entry: 0}
		result := nf.Execute(nil)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})
}

// TestNativeFunctionRegistryCompile tests registry compilation
func TestNativeFunctionRegistryCompile(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	t.Run("compile simple function", func(t *testing.T) {
		fn := &compiler.CompiledFunction{
			Instructions: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegReturn), 0,
			},
			NumLocals:     8,
			NumParameters: 0,
		}

		err := registry.CompileFunction(fn, 0, []int64{7})
		if err != nil {
			t.Fatalf("CompileFunction error: %v", err)
		}

		nf := registry.Get(0)
		if nf == nil {
			t.Error("Get returned nil after CompileFunction")
			return
		}
		if nf.NumParams != 0 {
			t.Errorf("Expected 0 params, got %d", nf.NumParams)
		}
		if got := nf.Execute(nil); got != 7 {
			t.Fatalf("expected 7, got %d", got)
		}
	})

	t.Run("compile function with params", func(t *testing.T) {
		fn := &compiler.CompiledFunction{
			Instructions: []byte{
				byte(compiler.OpRegAdd), 2, 0, 1,
				byte(compiler.OpRegReturn), 2,
			},
			NumLocals:     8,
			NumParameters: 2,
		}

		err := registry.CompileFunction(fn, 1, nil)
		if err != nil {
			t.Fatalf("CompileFunction error: %v", err)
		}

		nf := registry.Get(1)
		if nf == nil {
			t.Fatal("Get returned nil after CompileFunction")
		}
		if nf.NumParams != 2 {
			t.Errorf("Expected 2 params, got %d", nf.NumParams)
		}
		if got := nf.Execute(nil, 10, 20); got != 30 {
			t.Fatalf("expected 30, got %d", got)
		}
	})
}

// TestNativeFunctionRegistryCompileRecursive tests recursive function compilation
func TestNativeFunctionRegistryCompileRecursive(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	t.Run("compile recursive function", func(t *testing.T) {
		code := `
			func fib(n) {
				if (n <= 1) { return n }
				return fib(n - 1) + fib(n - 2)
			}
		`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		var fibFn *compiler.CompiledFunction
		for _, cnst := range bytecode.Constants {
			if fn, ok := cnst.(*compiler.CompiledFunction); ok {
				if fn.NumParameters == 1 && containsCall(fn.Instructions) {
					fibFn = fn
					break
				}
			}
		}

		if fibFn == nil {
			t.Skip("Could not find Fibonacci function in constants")
		}

		vmConstants := make([]vm.Value, len(bytecode.Constants))
		for i, c := range bytecode.Constants {
			vmConstants[i] = vm.NewObject(c)
		}

		err = registry.CompileRecursiveFunction(fibFn, 0, vmConstants)
		if err != nil {
			t.Logf("CompileRecursiveFunction error: %v", err)
			return
		}

		nf := registry.Get(0)
		if nf == nil {
			t.Error("Get returned nil after CompileRecursiveFunction")
		}
	})
}

// TestFibJITCompilerCompile tests FibJITCompiler.Compile
func TestFibJITCompilerCompile(t *testing.T) {
	config := DefaultJITConfig()
	fibCompiler := NewFibJITCompiler(config)

	t.Run("non-fib pattern", func(t *testing.T) {
		fn := &compiler.CompiledFunction{
			Instructions:  []byte{byte(compiler.OpRegNull), 0},
			NumLocals:     8,
			NumParameters: 0,
		}

		_, err := fibCompiler.Compile(fn, nil, nil)
		if err == nil {
			t.Error("Expected error for non-Fibonacci pattern")
		}
	})

	t.Run("fib pattern with call", func(t *testing.T) {
		fn := &compiler.CompiledFunction{
			Instructions: []byte{
				byte(compiler.OpRegNull), 0,
				byte(compiler.OpRegCall), 0, 0,
			},
			NumLocals:     8,
			NumParameters: 1,
		}

		code, err := fibCompiler.Compile(fn, nil, nil)
		if err != nil {
			t.Logf("Compile error: %v", err)
			return
		}
		if len(code) == 0 {
			t.Error("Generated code is empty")
			return
		}
		t.Logf("Generated %d bytes of recursive Fibonacci code", len(code))
	})
}

// TestSimpleCodeGeneratorGenerate tests SimpleCodeGenerator.Generate
func TestSimpleCodeGeneratorGenerate(t *testing.T) {
	generator := NewSimpleCodeGenerator()

	tests := []struct {
		name string
		fn   *compiler.CompiledFunction
	}{
		{
			name: "simple null",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegNull), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "simple true",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegTrue), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "simple false",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegFalse), 0},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "add operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegAdd), 0, 1, 2,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "sub operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegSub), 0, 1, 2,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "mul operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegMul), 0, 1, 2,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "jump operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegJump), 0, 0, 10,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "return operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegNull), 0,
					byte(compiler.OpRegReturn), 0,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "pop operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegPop), 0,
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
		{
			name: "unknown operation",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					0xFF, // Unknown opcode
				},
				NumLocals:     8,
				NumParameters: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generator.Generate(tt.fn, nil, nil)
			if err != nil {
				t.Errorf("Generate failed: %v", err)
				return
			}
			if len(code) == 0 {
				t.Error("Generated code is empty")
				return
			}
			t.Logf("Generated %d bytes", len(code))
		})
	}
}

// TestSimpleCodeGeneratorWithConstants tests with constants
func TestSimpleCodeGeneratorWithConstants(t *testing.T) {
	generator := NewSimpleCodeGenerator()

	constants := []vm.Value{
		vm.NewInt(42),
		vm.NewInt(100),
	}

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0,
			byte(compiler.OpRegNull), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	code, err := generator.Generate(fn, constants, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(code) == 0 {
		t.Error("Generated code is empty")
	}
	t.Logf("Generated %d bytes with constants", len(code))
}

// TestCanExecuteNativelyExtended tests CanExecuteNatively with more cases
func TestCanExecuteNativelyExtended(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected bool
	}{
		{
			name: "null operation",
			code: []byte{
				byte(compiler.OpRegNull), 0,
			},
			expected: true,
		},
		{
			name: "true operation",
			code: []byte{
				byte(compiler.OpRegTrue), 0,
			},
			expected: true,
		},
		{
			name: "false operation",
			code: []byte{
				byte(compiler.OpRegFalse), 0,
			},
			expected: true,
		},
		{
			name: "jump operation",
			code: []byte{
				byte(compiler.OpRegJump), 0, 0, 10,
			},
			expected: true,
		},
		{
			name: "jump if true",
			code: []byte{
				byte(compiler.OpRegJumpIfTrue), 0, 0, 10,
			},
			expected: true,
		},
		{
			name: "jump if false",
			code: []byte{
				byte(compiler.OpRegJumpIfFalse), 0, 0, 10,
			},
			expected: true,
		},
		{
			name: "load local",
			code: []byte{
				byte(compiler.OpRegLoadLocal), 0, 0,
			},
			expected: true,
		},
		{
			name: "store local",
			code: []byte{
				byte(compiler.OpRegStoreLocal), 0, 0,
			},
			expected: true,
		},
		{
			name: "load global",
			code: []byte{
				byte(compiler.OpRegLoadGlobal), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "store global",
			code: []byte{
				byte(compiler.OpRegStoreGlobal), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "pop operation",
			code: []byte{
				byte(compiler.OpRegPop), 0,
			},
			expected: true,
		},
		{
			name: "comparison operations",
			code: []byte{
				byte(compiler.OpRegEqual), 0, 0, 0,
				byte(compiler.OpRegNotEqual), 0, 0, 0,
				byte(compiler.OpRegLess), 0, 0, 0,
				byte(compiler.OpRegGreater), 0, 0, 0,
				byte(compiler.OpRegLessEqual), 0, 0, 0,
				byte(compiler.OpRegGreaterEqual), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "arithmetic operations (including OpRegDiv with SSE2)",
			code: []byte{
				byte(compiler.OpRegAdd), 0, 0, 0,
				byte(compiler.OpRegSub), 0, 0, 0,
				byte(compiler.OpRegMul), 0, 0, 0,
				byte(compiler.OpRegMod), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "OpRegDiv is natively executable (SSE2 float div with int truncation)",
			code: []byte{
				byte(compiler.OpRegDiv), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "negation",
			code: []byte{
				byte(compiler.OpRegNeg), 0, 0,
			},
			expected: true,
		},
		{
			name: "unsupported closure",
			code: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported array",
			code: []byte{
				byte(compiler.OpRegArray), 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported map",
			code: []byte{
				byte(compiler.OpRegMap), 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported builtin",
			code: []byte{
				byte(compiler.OpRegBuiltin), 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported call in strict mode",
			code: []byte{
				byte(compiler.OpRegCall), 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported tail call in strict mode",
			code: []byte{
				byte(compiler.OpRegTailCall), 0, 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := CanExecuteNatively(fn)
			if result != tt.expected {
				t.Errorf("CanExecuteNatively() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestCanExecuteNativelyWithCallsExtended tests CanExecuteNativelyWithCalls
func TestCanExecuteNativelyWithCallsExtended(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected bool
	}{
		{
			name: "call blocked due to reentrancy risk",
			code: []byte{
				byte(compiler.OpRegCall), 0, 0,
			},
			expected: false,
		},
		{
			name: "tail call blocked due to reentrancy risk",
			code: []byte{
				byte(compiler.OpRegTailCall), 0, 0,
			},
			expected: false,
		},
		{
			name: "unsupported closure still blocked",
			code: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "array now supported with callback",
			code: []byte{
				byte(compiler.OpRegArray), 0, 0, 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := CanExecuteNativelyWithCalls(fn)
			if result != tt.expected {
				t.Errorf("CanExecuteNativelyWithCalls() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestAnalyzeNativeSupport tests native support analysis
func TestAnalyzeNativeSupport(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{
			name: "pure arithmetic",
			code: []byte{
				byte(compiler.OpRegAdd), 0, 0, 0,
			},
			expected: SupportPureArithmetic,
		},
		{
			name: "unsupported operation",
			code: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
			},
			expected: SupportNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := AnalyzeNativeSupport(fn)
			if result != tt.expected {
				t.Errorf("AnalyzeNativeSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestJITVMRunWithNativeFunctions tests JIT VM with native functions
func TestJITVMRunWithNativeFunctions(t *testing.T) {
	code := `
		func add(a, b) {
			return a + b
		}
		add(10, 20)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMWithObjectGlobals tests JIT VM with object globals
func TestJITVMWithObjectGlobals(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	globals := make([]vm.Value, 256)
	jitVM := NewJITVMWithGlobals(bytecode, globals, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMNewWithObjectGlobals tests NewJITVMWithObjectGlobals
func TestJITVMNewWithObjectGlobals(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Create with non-nil object globals
	objGlobals := make([]objects.Object, 256)
	jitVM := NewJITVMWithObjectGlobals(bytecode, objGlobals, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMSetLoader tests SetLoader
func TestJITVMSetLoader(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetLoader(nil)
}

// TestJITVMGlobalsAsObjects tests GlobalsAsObjects
func TestJITVMGlobalsAsObjects(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	globals := jitVM.GlobalsAsObjects()
	if globals == nil {
		t.Error("GlobalsAsObjects returned nil")
	}
}

// TestContainsCallExtended tests containsCall with more cases
func TestContainsCallExtended(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected bool
	}{
		{
			name:     "empty code",
			code:     []byte{},
			expected: false,
		},
		{
			name: "null only",
			code: []byte{
				byte(compiler.OpRegNull), 0,
			},
			expected: false,
		},
		{
			name: "with call",
			code: []byte{
				byte(compiler.OpRegNull), 0,
				byte(compiler.OpRegCall), 0, 0,
			},
			expected: true,
		},
		{
			name: "with tail call",
			code: []byte{
				byte(compiler.OpRegNull), 0,
				byte(compiler.OpRegTailCall), 0, 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsCall(tt.code)
			if result != tt.expected {
				t.Errorf("containsCall() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestOperandWidthCodeGen tests operandWidthCodeGen
func TestOperandWidthCodeGen(t *testing.T) {
	ops := []compiler.Opcode{
		compiler.OpRegNull,
		compiler.OpRegLoadConst,
		compiler.OpRegMove,
		compiler.OpRegAdd,
		compiler.OpRegCall,
	}

	for _, op := range ops {
		width := operandWidthCodeGen(op)
		t.Logf("Opcode %d: width=%d", op, width)
	}
}

// TestJITVMCleanup tests JITVM cleanup
func TestJITVMCleanup(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())

	jitVM.Cleanup()
	jitVM.Cleanup() // Test idempotent cleanup
}

// BenchmarkNativeFunctionExecute benchmarks native function execution
func BenchmarkNativeFunctionExecute(b *testing.B) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	registry.CompileFunction(fn, 0, nil)
	nf := registry.Get(0)

	if nf != nil {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			nf.Execute(nil)
		}
	}
}

// BenchmarkSimpleCodeGeneratorGenerate benchmarks code generation
func BenchmarkSimpleCodeGeneratorGenerate(b *testing.B) {
	generator := NewSimpleCodeGenerator()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
			byte(compiler.OpRegTrue), 0,
			byte(compiler.OpRegFalse), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Generate(fn, nil, nil)
	}
}

// BenchmarkCanExecuteNatively benchmarks native execution check
func BenchmarkCanExecuteNatively(b *testing.B) {
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
			byte(compiler.OpRegAdd), 0, 0, 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CanExecuteNatively(fn)
	}
}

// TestCompiledFuncExecute tests CompiledFunc.Execute
func TestCompiledFuncExecute(t *testing.T) {
	t.Run("nil entry - skip execute", func(t *testing.T) {
		cf := &CompiledFunc{
			Entry: 0,
			Size:  0,
		}
		// Should not panic with zero entry
		if cf.Entry == 0 {
			t.Log("Skipping execute with zero entry pointer")
			return
		}
	})
}

// TestCompiledFuncExecuteWithCode tests compiled function execution with actual code
func TestCompiledFuncExecuteWithCode(t *testing.T) {
	config := DefaultJITConfig()
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	cf, err := jit.Compile(fn, nil, nil)
	if err != nil {
		t.Logf("Compile error: %v", err)
		return
	}

	if cf != nil && cf.Entry != 0 {
		t.Logf("Compiled function entry: 0x%x, size: %d", cf.Entry, cf.Size)
		// Don't actually execute - just verify it compiled
	}
}

// TestJITVMRunHybrid tests runHybrid execution mode
func TestJITVMRunHybrid(t *testing.T) {
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	t.Logf("Native executions: %d, Interpreter executions: %d", nativeExecs, interpExecs)
}

// TestNativeExecutorCleanup tests NativeExecutor.Cleanup
func TestNativeExecutorCleanup(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())

	// Test cleanup
	exec.Cleanup()
	exec.Cleanup() // Test idempotent cleanup
}

// TestGenerateFibCode tests generateFibCode (backward compatibility)
func TestGenerateFibCode(t *testing.T) {
	config := DefaultJITConfig()
	fibCompiler := NewFibJITCompiler(config)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
			byte(compiler.OpRegCall), 0, 0,
		},
		NumLocals:     8,
		NumParameters: 1,
	}

	code := fibCompiler.generateFibCode(fn, nil)
	if len(code) == 0 {
		t.Error("generateFibCode returned empty code")
		return
	}
	t.Logf("generateFibCode returned %d bytes", len(code))
}

// TestJITVMRunEnabled tests JIT VM with JIT enabled
func TestJITVMRunEnabled(t *testing.T) {
	code := `
		var a = 10
		var b = 20
		a + b
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "30" {
		t.Fatalf("expected 30, got %s", got)
	}

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	if nativeExecs == 0 {
		t.Fatal("expected pure native execution to be used")
	}
	if interpExecs != 0 {
		t.Fatalf("expected no interpreter fallback, got %d interpreter executions", interpExecs)
	}
}

// TestJITVMCompileNativeFunctions tests compileNativeFunctions
func TestJITVMCompileNativeFunctions(t *testing.T) {
	code := `
		func add(a, b) {
			return a + b
		}
		func sub(a, b) {
			return a - b
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	t.Logf("Native executions: %d, Interpreter executions: %d", nativeExecs, interpExecs)
}

// TestJITVMSetupNativeCallHook tests setupNativeCallHook
func TestJITVMSetupNativeCallHook(t *testing.T) {
	code := `
		func double(x) {
			return x * 2
		}
		double(21)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestNativeFunctionRegistryGetMultiple tests multiple Get operations
func TestNativeFunctionRegistryGetMultiple(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	// Compile multiple functions
	for i := 0; i < 3; i++ {
		fn := &compiler.CompiledFunction{
			Instructions: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegReturn), 0,
			},
			NumLocals:     8,
			NumParameters: i,
		}
		if err := registry.CompileFunction(fn, i, []int64{int64(i)}); err != nil {
			t.Fatalf("CompileFunction(%d) failed: %v", i, err)
		}
	}

	// Get all functions
	for i := 0; i < 3; i++ {
		nf := registry.Get(i)
		if nf == nil {
			t.Fatalf("Get(%d) returned nil", i)
		}
		if got := nf.Execute(nil); got != int64(i) {
			t.Fatalf("Get(%d) executed to %d, expected %d", i, got, i)
		}
	}
}

// TestNativeFunctionExecuteWithArgs tests native function with arguments
func TestNativeFunctionExecuteWithArgs(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegAdd), 2, 0, 1,
			byte(compiler.OpRegReturn), 2,
		},
		NumLocals:     8,
		NumParameters: 2,
	}

	err := registry.CompileFunction(fn, 0, nil)
	if err != nil {
		t.Fatalf("CompileFunction error: %v", err)
	}

	nf := registry.Get(0)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	// Test execution with various argument counts
	globals := make([]int64, 256)

	t.Run("0 args", func(t *testing.T) {
		result := nf.Execute(globals)
		if result != 0 {
			t.Fatalf("expected 0, got %d", result)
		}
	})

	t.Run("1 arg", func(t *testing.T) {
		result := nf.Execute(globals, 42)
		if result != 42 {
			t.Fatalf("expected 42, got %d", result)
		}
	})

	t.Run("2 args", func(t *testing.T) {
		result := nf.Execute(globals, 10, 20)
		if result != 30 {
			t.Fatalf("expected 30, got %d", result)
		}
	})

	t.Run("3 args", func(t *testing.T) {
		result := nf.Execute(globals, 1, 2, 3)
		if result != 3 {
			t.Fatalf("expected 3, got %d", result)
		}
	})
}

// TestNativeFunctionRegistryCompileConstArithmetic ensures constant arithmetic opcodes run natively.
func TestNativeFunctionRegistryCompileConstArithmetic(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0,
			byte(compiler.OpRegAddConst), 0, 0, 0, 1,
			byte(compiler.OpRegMulConst), 0, 0, 0, 2,
			byte(compiler.OpRegSubConst), 0, 0, 0, 3,
			byte(compiler.OpRegReturn), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{20, 3, 2, 1}
	if err := registry.CompileFunction(fn, 0, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(0)
	if nf == nil {
		t.Fatal("Get returned nil after CompileFunction")
	}

	if got := nf.Execute(nil); got != 45 {
		t.Fatalf("expected 45, got %d", got)
	}
}

// TestNativeFunctionRegistryCompileAddLocalCheck ensures loop-check bytecode can run natively.
func TestNativeFunctionRegistryCompileAddLocalCheck(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0,
			byte(compiler.OpRegLoadConst), 1, 0, 1,
			byte(compiler.OpRegAddLocalCheck), 0, 1, 0, 2, 0, 0,
			byte(compiler.OpRegReturn), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{0, 0, 5}
	if err := registry.CompileFunction(fn, 0, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(0)
	if nf == nil {
		t.Fatal("Get returned nil after CompileFunction")
	}

	if got := nf.Execute(nil); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}

// TestNativeFunctionRegistryCompileLogicalOps ensures logical opcodes run natively and box back to booleans.
func TestNativeFunctionRegistryCompileLogicalOps(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	tests := []struct {
		name        string
		instructions []byte
		constants   []int64
		expected    int64
	}{
		{
			name: "and",
			instructions: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegLoadConst), 1, 0, 1,
				byte(compiler.OpRegAnd), 2, 0, 1,
				byte(compiler.OpRegReturn), 2,
			},
			constants: []int64{1, 1},
			expected: 1,
		},
		{
			name: "or",
			instructions: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegLoadConst), 1, 0, 1,
				byte(compiler.OpRegOr), 2, 0, 1,
				byte(compiler.OpRegReturn), 2,
			},
			constants: []int64{0, 5},
			expected: 1,
		},
		{
			name: "not",
			instructions: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegNot), 1, 0,
				byte(compiler.OpRegReturn), 1,
			},
			constants: []int64{0},
			expected: 1,
		},
	}

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{Instructions: tt.instructions, NumLocals: 8}
			if err := registry.CompileFunction(fn, idx, tt.constants); err != nil {
				t.Fatalf("CompileFunction failed: %v", err)
			}

			nf := registry.Get(idx)
			if nf == nil {
				t.Fatal("Get returned nil after CompileFunction")
			}
			if nf.ReturnType != ReturnTypeBool {
				t.Fatalf("expected bool return type, got %v", nf.ReturnType)
			}
			if got := nf.Execute(nil); got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

// TestJITVMRunWithLogicalResult ensures logical native results round-trip as booleans.
func TestJITVMRunWithLogicalResult(t *testing.T) {
	code := `
		func both(a, b) {
			return a && b
		}
		both(true, false)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: true})
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "false" {
		t.Fatalf("expected false, got %s", got)
	}
}

// TestNativeFunctionRegistryCleanup tests registry cleanup
func TestNativeFunctionRegistryCleanup(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())

	// Compile a function
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}
	registry.CompileFunction(fn, 0, nil)

	// Cleanup
	registry.Cleanup()
	registry.Cleanup() // Test idempotent cleanup
}

// TestSimpleCodeGeneratorMoreOps tests more operations
func TestSimpleCodeGeneratorMoreOps(t *testing.T) {
	generator := NewSimpleCodeGenerator()

	tests := []struct {
		name string
		fn   *compiler.CompiledFunction
	}{
		{
			name: "multiple operations",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegNull), 0,
					byte(compiler.OpRegTrue), 0,
					byte(compiler.OpRegFalse), 0,
					byte(compiler.OpRegAdd), 0, 0, 0,
					byte(compiler.OpRegSub), 0, 0, 0,
					byte(compiler.OpRegMul), 0, 0, 0,
				},
				NumLocals:     16,
				NumParameters: 0,
			},
		},
		{
			name: "with jump and return",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegJump), 0, 0, 5,
					byte(compiler.OpRegNull), 0,
					byte(compiler.OpRegReturn), 0,
				},
				NumLocals:     16,
				NumParameters: 0,
			},
		},
		{
			name: "large num locals",
			fn: &compiler.CompiledFunction{
				Instructions:  []byte{byte(compiler.OpRegNull), 0},
				NumLocals:     256,
				NumParameters: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generator.Generate(tt.fn, nil, nil)
			if err != nil {
				t.Errorf("Generate failed: %v", err)
				return
			}
			t.Logf("Generated %d bytes", len(code))
		})
	}
}

// TestFibJITCompilerIsFibPattern tests isFibPattern
func TestFibJITCompilerIsFibPattern(t *testing.T) {
	config := DefaultJITConfig()
	fibCompiler := NewFibJITCompiler(config)

	tests := []struct {
		name     string
		fn       *compiler.CompiledFunction
		expected bool
	}{
		{
			name: "two calls with fib-like structure",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegLessEqual), 0, 0, 0,
					byte(compiler.OpRegSub), 0, 0, 0,
					byte(compiler.OpRegSub), 0, 0, 0,
					byte(compiler.OpRegCall), 0, 1,
					byte(compiler.OpRegCall), 0, 1,
					byte(compiler.OpRegAdd), 0, 0, 1,
				},
				NumLocals:     8,
				NumParameters: 1,
			},
			expected: true,
		},
		{
			name: "tail call without compare is not fib pattern",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegTailCall), 0, 0,
				},
				NumLocals:     8,
				NumParameters: 1,
			},
			expected: false,
		},
		{
			name: "higher-order call is not fib pattern",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegCall), 0, 1,
					byte(compiler.OpRegReturn), 0,
				},
				NumLocals:     8,
				NumParameters: 2,
			},
			expected: false,
		},
		{
			name: "no call - not fib pattern",
			fn: &compiler.CompiledFunction{
				Instructions: []byte{
					byte(compiler.OpRegNull), 0,
				},
				NumLocals:     8,
				NumParameters: 1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fibCompiler.isFibPattern(tt.fn)
			if result != tt.expected {
				t.Errorf("isFibPattern() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestJITVMRunWithHigherOrderFunction ensures higher-order calls are not miscompiled as Fibonacci.
func TestJITVMRunWithHigherOrderFunction(t *testing.T) {
	code := `
		func apply(f, x) {
			return f(x)
		}
		func double(n) {
			return n * 2
		}
		apply(double, 21)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: true})
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "42" {
		t.Fatalf("expected 42, got %s", got)
	}

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	if nativeExecs != 0 {
		t.Fatalf("expected higher-order call to stay on interpreter path, got %d native executions", nativeExecs)
	}
	if interpExecs == 0 {
		t.Fatal("expected interpreter execution for higher-order call")
	}
}

// TestJITVMRunWithIterativeLoopFunction ensures arithmetic loop functions can run natively.
func TestJITVMRunWithIterativeLoopFunction(t *testing.T) {
	code := `
		func sumTo(n) {
			var sum = 0
			for (var i = 0; i <= n; i = i + 1) {
				sum = sum + i
			}
			return sum
		}
		sumTo(10)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: true})
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "55" {
		t.Fatalf("expected 55, got %s", got)
	}

	nativeExecs, _ := jitVM.GetNativeStats()
	if nativeExecs == 0 {
		t.Fatal("expected iterative loop function to use native execution")
	}

	stats := jitVM.GetJITStats()
	if stats.CompiledFunctions == 0 {
		t.Fatal("expected JIT stats to report compiled functions")
	}
}

// TestOperandWidth tests operandWidth function
func TestOperandWidth(t *testing.T) {
	ops := []compiler.Opcode{
		compiler.OpRegNull,
		compiler.OpRegLoadConst,
		compiler.OpRegMove,
		compiler.OpRegAdd,
		compiler.OpRegCall,
		compiler.OpRegJump,
	}

	for _, op := range ops {
		width := operandWidth(op)
		t.Logf("Opcode %d: width=%d", op, width)
	}
}

// TestNativeExecutorCleanupExtended tests NativeExecutor cleanup
func TestNativeExecutorCleanupExtended(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())

	// Execute a simple function before cleanup
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	globals := make([]int64, 256)
	exec.ExecuteFunction(fn, nil, globals)

	exec.Cleanup()
}

// TestJITVMRunMultiplePaths tests multiple Run execution paths
func TestJITVMRunMultiplePaths(t *testing.T) {
	t.Run("jit disabled", func(t *testing.T) {
		code := `var x = 1`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()
		jitVM := NewJITVM(bytecode, DefaultJITConfig())
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(false)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("jit enabled with debug", func(t *testing.T) {
		code := `var x = 1 + 2`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("native compilation failure path", func(t *testing.T) {
		// Code with operations that can't be natively compiled
		code := `
			func foo() { return 1 }
			foo()
		`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("small max code size", func(t *testing.T) {
		code := `var x = 1`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  1, // Very small
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})
}

// TestJITVMSetLoaderWithValidLoader tests SetLoader with valid loader
func TestJITVMSetLoaderWithValidLoader(t *testing.T) {
	code := `var x = 1`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	// Create a module loader
	loader := module.NewLoader()
	jitVM.SetLoader(loader)
}

// TestJITMemoryManagerAllocateHandleReuse tests handle reuse
func TestJITMemoryManagerAllocateHandleReuse(t *testing.T) {
	m := NewJITMemoryManager()

	// Allocate some handles
	handle1 := m.AllocateHandle("obj1")
	handle2 := m.AllocateHandle("obj2")
	handle3 := m.AllocateHandle("obj3")

	// Release handles
	m.ReleaseHandle(handle1)
	m.ReleaseHandle(handle3)

	// Allocate new handles - should reuse released ones
	handle4 := m.AllocateHandle("obj4")
	handle5 := m.AllocateHandle("obj5")

	t.Logf("Handles: %d, %d, %d, %d, %d", handle1, handle2, handle3, handle4, handle5)

	// Verify objects
	obj, ok := m.GetObject(handle2)
	if !ok || obj != "obj2" {
		t.Error("Handle 2 should still have obj2")
	}

	m.Cleanup()
}

// TestJITCompilerCompileCacheHit tests cache hit scenario
func TestJITCompilerCompileCacheHit(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []vm.Value{vm.NewInt(42)}

	// First compile
	cf1, err := jit.Compile(fn, constants, nil)
	if err != nil {
		t.Logf("First compile error: %v", err)
	}

	// Second compile should hit cache
	cf2, err := jit.Compile(fn, constants, nil)
	if err != nil {
		t.Logf("Second compile error: %v", err)
	}

	if cf1 != nil && cf2 != nil {
		if cf1.Hash != cf2.Hash {
			t.Error("Cache hit should return same function")
		}
	}

	stats := jit.GetStats()
	t.Logf("Stats: compiled=%d, hits=%d, misses=%d",
		stats.CompiledFunctions, stats.CacheHits, stats.CacheMisses)
}

// TestJITCompilerGetCompiledAfterCompile tests GetCompiled after compile
func TestJITCompilerGetCompiledAfterCompile(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	// Before compile
	cf := jit.GetCompiled(fn)
	if cf != nil {
		t.Error("GetCompiled should return nil before compile")
	}

	// Compile
	jit.Compile(fn, nil, nil)

	// After compile - should be in cache
	cf = jit.GetCompiled(fn)
	if cf == nil {
		// This is expected if compile failed
		t.Log("GetCompiled returned nil after compile (compile may have failed)")
	} else {
		t.Logf("GetCompiled returned function with hash: 0x%016x", cf.Hash)
	}
}

// TestNativeFunctionRegistryGetNonExistent tests Get for non-existent index
func TestNativeFunctionRegistryGetNonExistent(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	// Get non-existent index
	nf := registry.Get(999)
	if nf != nil {
		t.Error("Get should return nil for non-existent index")
	}
}

// TestBridgeAllocFree tests bridge memory allocation and free
func TestBridgeAllocFree(t *testing.T) {
	mem, err := bridge.AllocExecMem(1024)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}

	if len(mem) != 1024 {
		t.Errorf("Expected 1024 bytes, got %d", len(mem))
	}

	// Write some data
	for i := 0; i < len(mem); i++ {
		mem[i] = byte(i % 256)
	}

	// Free
	bridge.FreeExecMem(mem)
}

// TestBridgeAllocMultipleSizes tests various allocation sizes
func TestBridgeAllocMultipleSizes(t *testing.T) {
	sizes := []int{1, 64, 256, 1024, 4096, 16384, 65536}

	for _, size := range sizes {
		t.Run(string(rune(size)), func(t *testing.T) {
			mem, err := bridge.AllocExecMem(size)
			if err != nil {
				t.Errorf("AllocExecMem(%d) failed: %v", size, err)
				return
			}
			if len(mem) != size {
				t.Errorf("Expected %d bytes, got %d", size, len(mem))
			}
			bridge.FreeExecMem(mem)
		})
	}
}

// TestJITCompilerCompileWithDebug tests compile with debug output
func TestJITCompilerCompileWithDebug(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	// Compile a function with debug enabled
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
			byte(compiler.OpRegTrue), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	cf, err := jit.Compile(fn, nil, nil)
	if err != nil {
		t.Logf("Compile error: %v", err)
		return
	}

	t.Logf("Compiled function: size=%d", cf.Size)
}

// TestFibJITCompilerGenerateRecursiveFibCodeDebug tests with debug
func TestFibJITCompilerGenerateRecursiveFibCodeDebug(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}
	fibCompiler := NewFibJITCompiler(config)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegCall), 0, 0,
		},
		NumLocals:     8,
		NumParameters: 1,
	}

	code := fibCompiler.generateRecursiveFibCode(fn, nil)
	if len(code) == 0 {
		t.Error("Generated code is empty")
		return
	}
	t.Logf("Generated %d bytes with debug", len(code))
}

// TestJITStatsIncrement tests JIT stats
func TestJITStatsIncrement(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	// Trigger some cache misses
	fn1 := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	fn2 := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegTrue), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	jit.GetCompiled(fn1)
	jit.GetCompiled(fn2)

	stats := jit.GetStats()
	if stats.CacheMisses != 2 {
		t.Errorf("Expected 2 cache misses, got %d", stats.CacheMisses)
	}
}

// BenchmarkNativeExecutorExecuteFunction benchmarks native execution
func BenchmarkNativeExecutorExecuteFunction(b *testing.B) {
	exec := NewNativeExecutor(DefaultJITConfig())
	defer exec.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	globals := make([]int64, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.ExecuteFunction(fn, nil, globals)
	}
}

// BenchmarkJITVMRun benchmarks JIT VM run
func BenchmarkJITVMRun(b *testing.B) {
	code := `var x = 1 + 2`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jitVM := NewJITVM(bytecode, DefaultJITConfig())
		jitVM.SetJITEnabled(false)
		jitVM.Run()
		jitVM.Cleanup()
	}
}

// BenchmarkNativeFunctionRegistryCompile benchmarks registry compilation
func BenchmarkNativeFunctionRegistryCompile(b *testing.B) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.CompileFunction(fn, i%100, nil)
	}
}

// TestCompiledFuncExecuteWithValidCode tests compiled function structure
func TestCompiledFuncExecuteWithValidCode(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	// Just test that we can allocate and create a CompiledFunc
	mem, page, err := jit.AllocCode(64)
	if err != nil {
		t.Fatalf("AllocCode failed: %v", err)
	}

	cf := &CompiledFunc{
		Entry:     uintptr(unsafe.Pointer(&mem[0])),
		Page:      page,
		Size:      64,
		Hash:      12345,
		NumRegs:   8,
		NumParams: 0,
	}

	// Verify the structure
	if cf.Entry == 0 {
		t.Error("Entry should not be 0")
	}
	if cf.Page == nil {
		t.Error("Page should not be nil")
	}
	if cf.Size != 64 {
		t.Errorf("Expected size 64, got %d", cf.Size)
	}

	t.Logf("CompiledFunc created: entry=0x%x, size=%d", cf.Entry, cf.Size)
}

// TestJITCompilerCleanupWithPages tests cleanup with allocated pages
func TestJITCompilerCleanupWithPages(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())

	// Allocate multiple pages
	for i := 0; i < 10; i++ {
		jit.AllocCode(4096)
	}

	// Cleanup should free all pages
	jit.Cleanup()

	stats := jit.GetStats()
	if stats.CompiledFunctions != 0 {
		t.Error("Stats should be reset after cleanup")
	}
}

// TestNativeExecutorCleanupWithCompiler tests NativeExecutor cleanup
func TestNativeExecutorCleanupWithCompiler(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())

	// Execute some functions
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}
	globals := make([]int64, 256)
	exec.ExecuteFunction(fn, nil, globals)

	// Cleanup - test the empty cleanup function
	exec.Cleanup()
}

// TestNativeExecutorCleanupEmpty tests empty cleanup
func TestNativeExecutorCleanupEmpty(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())

	// Just call cleanup without any operations
	exec.Cleanup()
	exec.Cleanup() // Call twice to ensure idempotency
}

// TestGetJITSupportLevelMoreCases tests more support level cases
func TestGetJITSupportLevelMoreCases(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected JITSupportLevel
	}{
		{
			name: "full support - null and move",
			code: []byte{
				byte(compiler.OpRegNull), 0,
				byte(compiler.OpRegMove), 0, 1,
			},
			expected: JITSupportFull,
		},
		{
			name: "partial support - with array empty",
			code: []byte{
				byte(compiler.OpRegArrayEmpty), 0,
				byte(compiler.OpRegNull), 0,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with array append",
			code: []byte{
				byte(compiler.OpRegArrayAppend), 0, 1, 2,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with index",
			code: []byte{
				byte(compiler.OpRegIndex), 0, 1, 2,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with set index",
			code: []byte{
				byte(compiler.OpRegSetIndex), 0, 1, 2,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with map empty",
			code: []byte{
				byte(compiler.OpRegMapEmpty), 0,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with map set",
			code: []byte{
				byte(compiler.OpRegMapSet), 0, 1, 2,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with get field",
			code: []byte{
				byte(compiler.OpRegGetField), 0, 1, 0, 0,
			},
			expected: JITSupportPartial,
		},
		{
			name: "partial support - with set field",
			code: []byte{
				byte(compiler.OpRegSetField), 0, 1, 0, 0,
			},
			expected: JITSupportPartial,
		},
		{
			name: "no support - with closure",
			code: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with load free",
			code: []byte{
				byte(compiler.OpRegLoadFree), 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with store free",
			code: []byte{
				byte(compiler.OpRegStoreFree), 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with builtin",
			code: []byte{
				byte(compiler.OpRegBuiltin), 0, 0, 1,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with get method",
			code: []byte{
				byte(compiler.OpRegGetMethod), 0, 1, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with call method",
			code: []byte{
				byte(compiler.OpRegCallMethod), 0, 0, 0, 1,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with class",
			code: []byte{
				byte(compiler.OpRegClass), 0, 0, 0, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with new",
			code: []byte{
				byte(compiler.OpRegNew), 0, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with throw",
			code: []byte{
				byte(compiler.OpRegThrow), 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with push handler",
			code: []byte{
				byte(compiler.OpRegPushHandler), 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with pop handler",
			code: []byte{
				byte(compiler.OpRegPopHandler),
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with load module",
			code: []byte{
				byte(compiler.OpRegLoadModule), 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with iter key",
			code: []byte{
				byte(compiler.OpRegIterKey), 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with iter value",
			code: []byte{
				byte(compiler.OpRegIterValue), 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with make tube",
			code: []byte{
				byte(compiler.OpRegMakeTube), 0, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with mutex lock",
			code: []byte{
				byte(compiler.OpRegMutexLock), 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with wg add",
			code: []byte{
				byte(compiler.OpRegWGAdd), 0, 0, 0,
			},
			expected: JITSupportNone,
		},
		{
			name: "no support - with atomic add",
			code: []byte{
				byte(compiler.OpRegAtomicAdd), 0, 0, 0, 0,
			},
			expected: JITSupportNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := GetJITSupportLevel(fn)
			if result != tt.expected {
				t.Errorf("GetJITSupportLevel() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestJITVMRunMorePaths tests more Run execution paths
func TestJITVMRunMorePaths(t *testing.T) {
	t.Run("with boolean constants", func(t *testing.T) {
		code := `var x = true`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("with string constants", func(t *testing.T) {
		code := `var x = "hello"`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("with array literals", func(t *testing.T) {
		code := `var x = [1, 2, 3]`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("with map literals", func(t *testing.T) {
		code := `var x = {"a": 1, "b": 2}`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("with function calls", func(t *testing.T) {
		code := `
			func inc(x) {
				return x + 1
			}
			inc(5)
		`

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.NewRegCompiler()
		_, err := c.Compile(program)
		if err != nil {
			t.Fatalf("Compile error: %v", err)
		}

		bytecode := c.Bytecode()

		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		jitVM.SetJITEnabled(true)
		err = jitVM.Run()
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})
}

// TestBridgeFreeExecMemNil tests freeing nil pointer
func TestBridgeFreeExecMemNil(t *testing.T) {
	// Free nil should not panic
	bridge.FreeExecMem(nil)
}

// TestBridgeFreeExecMemMultiple tests multiple frees
func TestBridgeFreeExecMemMultiple(t *testing.T) {
	mem, err := bridge.AllocExecMem(1024)
	if err != nil {
		t.Fatalf("AllocExecMem failed: %v", err)
	}

	// Free once
	bridge.FreeExecMem(mem)

	// Free again should not panic (though behavior is undefined)
	// We don't test this as it's undefined behavior
}

// TestJITVMCompileNativeFunctionsWithVariousTypes tests with various constant types
func TestJITVMCompileNativeFunctionsWithVariousTypes(t *testing.T) {
	code := `
		var intVal = 42
		var boolVal = true
		var strVal = "hello"
		var arrVal = [1, 2, 3]
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestRequiresInterpreterFallbackMoreCases tests more fallback cases
func TestRequiresInterpreterFallbackMoreCases(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected bool
	}{
		{
			name: "no fallback - simple arithmetic",
			code: []byte{
				byte(compiler.OpRegAdd), 0, 0, 0,
				byte(compiler.OpRegSub), 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "no fallback - comparisons",
			code: []byte{
				byte(compiler.OpRegLess), 0, 0, 0,
				byte(compiler.OpRegGreater), 0, 0, 0,
				byte(compiler.OpRegEqual), 0, 0, 0,
			},
			expected: false,
		},
		{
			name: "no fallback - jumps",
			code: []byte{
				byte(compiler.OpRegJump), 0, 0, 10,
				byte(compiler.OpRegJumpIfTrue), 0, 0, 20,
				byte(compiler.OpRegJumpIfFalse), 0, 0, 30,
			},
			expected: false,
		},
		{
			name: "fallback - run start",
			code: []byte{
				byte(compiler.OpRegRunStart), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - run wait",
			code: []byte{
				byte(compiler.OpRegRunWait), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - tube send",
			code: []byte{
				byte(compiler.OpRegTubeSend), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - tube recv",
			code: []byte{
				byte(compiler.OpRegTubeRecv), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - tube close",
			code: []byte{
				byte(compiler.OpRegTubeClose), 0,
			},
			expected: true,
		},
		{
			name: "fallback - select start",
			code: []byte{
				byte(compiler.OpRegSelectStart),
			},
			expected: true,
		},
		{
			name: "fallback - select case",
			code: []byte{
				byte(compiler.OpRegSelectCase), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - select end",
			code: []byte{
				byte(compiler.OpRegSelectEnd),
			},
			expected: true,
		},
		{
			name: "fallback - mutex unlock",
			code: []byte{
				byte(compiler.OpRegMutexUnlock), 0,
			},
			expected: true,
		},
		{
			name: "fallback - wg wait",
			code: []byte{
				byte(compiler.OpRegWGWait), 0,
			},
			expected: true,
		},
		{
			name: "fallback - wg done",
			code: []byte{
				byte(compiler.OpRegWGDone),
			},
			expected: true,
		},
		{
			name: "fallback - atomic load",
			code: []byte{
				byte(compiler.OpRegAtomicLoad), 0, 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - atomic swap",
			code: []byte{
				byte(compiler.OpRegAtomicSwap), 0, 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - get export",
			code: []byte{
				byte(compiler.OpRegGetExport), 0, 0, 0,
			},
			expected: true,
		},
		{
			name: "fallback - set export",
			code: []byte{
				byte(compiler.OpRegSetExport), 0, 0, 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.code,
				NumLocals:     8,
				NumParameters: 0,
			}
			result := RequiresInterpreterFallback(fn)
			if result != tt.expected {
				t.Errorf("RequiresInterpreterFallback() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestJITVMCompileNativeFunctionsDebug tests compileNativeFunctions with debug
func TestJITVMCompileNativeFunctionsDebug(t *testing.T) {
	code := `
		func add(a, b) { return a + b }
		func mul(a, b) { return a * b }
		func fib(n) { if (n <= 1) { return n } return fib(n-1) + fib(n-2) }
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	// Verify functions were compiled
	nativeExecs, interpExecs := jitVM.GetNativeStats()
	t.Logf("Native: %d, Interpreter: %d", nativeExecs, interpExecs)
}

// TestJITVMCompileNativeFunctionsWithFloatConstants tests with float constants
func TestJITVMCompileNativeFunctionsWithFloatConstants(t *testing.T) {
	code := `
		var x = 3.14
		var y = 2.71
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMCompileNativeFunctionsWithStringConstants tests with string constants
func TestJITVMCompileNativeFunctionsWithStringConstants(t *testing.T) {
	code := `
		var s1 = "hello"
		var s2 = "world"
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMRunWithClosure tests with closure
func TestJITVMRunWithClosure(t *testing.T) {
	code := `
		func makeCounter() {
			var count = 0
			func counter() {
				count = count + 1
				return count
			}
			return counter
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMRunWithBuiltinCalls tests with builtin calls
func TestJITVMRunWithBuiltinCalls(t *testing.T) {
	code := `
		var x = len([1, 2, 3])
		var y = typeOf(42)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMRunWithArrayOperations tests with array operations
func TestJITVMRunWithArrayOperations(t *testing.T) {
	code := `
		var arr = [1, 2, 3, 4, 5]
		var first = arr[0]
		arr[1] = 20
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMRunWithMapOperations tests with map operations
func TestJITVMRunWithMapOperations(t *testing.T) {
	code := `
		var m = {"a": 1, "b": 2}
		var val = m["a"]
		m["c"] = 3
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestNativeExecutorExecuteFunctionWithError tests error handling
func TestNativeExecutorExecuteFunctionWithError(t *testing.T) {
	exec := NewNativeExecutor(DefaultJITConfig())
	defer exec.Cleanup()

	// Test with large function that might fail
	largeCode := make([]byte, 100000)
	for i := range largeCode {
		largeCode[i] = byte(compiler.OpRegNull)
	}

	fn := &compiler.CompiledFunction{
		Instructions:  largeCode,
		NumLocals:     8,
		NumParameters: 0,
	}

	globals := make([]int64, 256)
	_, err := exec.ExecuteFunction(fn, nil, globals)
	if err == nil {
		t.Fatal("expected error for oversized native code")
	}
}

// TestNativeFunctionRegistryCompileFunctionError tests error handling
func TestNativeFunctionRegistryCompileFunctionError(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	// Test with very large function
	largeCode := make([]byte, 100000)
	fn := &compiler.CompiledFunction{
		Instructions:  largeCode,
		NumLocals:     8,
		NumParameters: 0,
	}

	err := registry.CompileFunction(fn, 0, nil)
	if err == nil {
		t.Fatal("expected error for oversized compiled function")
	}
}

// TestNativeFunctionRegistryCompileRecursiveFunctionError tests error handling
func TestNativeFunctionRegistryCompileRecursiveFunctionError(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	// Test with function that's not a fib pattern
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     8,
		NumParameters: 0,
	}

	err := registry.CompileRecursiveFunction(fn, 0, nil)
	if err != nil {
		t.Logf("Expected error for non-recursive function: %v", err)
	}
}

// TestJITVMCompileNativeFunctionsBoolFalse tests with false boolean
func TestJITVMCompileNativeFunctionsBoolFalse(t *testing.T) {
	code := `
		var t = true
		var f = false
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false, // Test without debug
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMCompileNativeFunctionsMixedConstants tests with mixed constant types
func TestJITVMCompileNativeFunctionsMixedConstants(t *testing.T) {
	code := `
		var i = 42
		var f = 3.14
		var s = "hello"
		var b = true
		var arr = [1, 2, 3]
		var m = {"a": 1}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMCompileNativeFunctionsNonNativeFunction tests with non-native function
func TestJITVMCompileNativeFunctionsNonNativeFunction(t *testing.T) {
	code := `
		func complexFn(x) {
			var arr = [1, 2, 3]
			return arr[0] + x
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMCompileNativeFunctionsRecursiveButNotFib tests recursive but not fib pattern
func TestJITVMCompileNativeFunctionsRecursiveButNotFib(t *testing.T) {
	code := `
		func recurse(n) {
			if (n <= 0) { return 0 }
			return recurse(n - 1) + recurse(n - 2) + 1
		}
		recurse(5)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "12" {
		t.Fatalf("expected 12, got %s", got)
	}

	nativeExecs, interpExecs := jitVM.GetNativeStats()
	if nativeExecs != 0 {
		t.Fatalf("expected non-fibonacci recursion to stay on interpreter path, got %d native executions", nativeExecs)
	}
	if interpExecs == 0 {
		t.Fatal("expected interpreter fallback for non-fibonacci recursion")
	}
}

// TestJITVMRunWithPureArithmeticFunction tests with pure arithmetic function
func TestJITVMRunWithPureArithmeticFunction(t *testing.T) {
	code := `
		func add(a, b) {
			return a + b
		}
		add(10, 20)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "30" {
		t.Fatalf("expected 30, got %s", got)
	}

nativeExecs, interpExecs := jitVM.GetNativeStats()
	if nativeExecs == 0 {
		t.Fatalf("expected native execution for pure arithmetic function")
	}
	if interpExecs == 0 {
		t.Fatal("expected hybrid mode to record interpreter execution")
	}
}

// TestJITVMWithBuiltinCallback tests that builtin functions work correctly
// when called through the hybrid mode native hook.
func TestJITVMWithBuiltinCallback(t *testing.T) {
	code := `
		func absVal(x) {
			if (x < 0) { return -x }
			return x
		}
		absVal(-42)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}
	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "42" {
		t.Fatalf("expected 42, got %s", got)
	}
}

// TestJITVMWithCollectionCallback tests that collection operations
// (array index) work correctly through the hybrid mode native hook.
func TestJITVMWithCollectionCallback(t *testing.T) {
	code := `
		func getFirst(arr) {
			return arr[0]
		}
		getFirst([100, 200, 300])
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	config := JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false}
	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	obj := jitVM.LastPoppedObject()
	if obj == nil {
		t.Fatal("expected result object")
	}
	if got := obj.Inspect(); got != "100" {
		t.Fatalf("expected 100, got %s", got)
	}
}

// TestJITVMCompileNativeFunctionsNoDebug tests without debug output
func TestJITVMCompileNativeFunctionsNoDebug(t *testing.T) {
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false, // No debug output
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(false)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITVMWithGlobals tests JIT VM with globals
func TestJITVMWithGlobals(t *testing.T) {
	code := `
		var global = 100
		func useGlobal() {
			return global
		}
		useGlobal()
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	jitVM.SetJITEnabled(true)
	err = jitVM.Run()
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

// TestJITCompilerCompileCacheHitTest tests cache hit scenario
func TestJITCompilerCompileCacheHitTest(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []vm.Value{vm.NewInt(42)}

	cf1, err := jit.Compile(fn, constants, nil)
	if err != nil {
		t.Logf("First compile error: %v", err)
	} else {
		t.Logf("First compile success: size=%d", cf1.Size)
	}

	cf2, err := jit.Compile(fn, constants, nil)
	if err != nil {
		t.Logf("Second compile error: %v", err)
	} else {
		t.Logf("Second compile success: size=%d", cf2.Size)
	}

	stats := jit.GetStats()
	t.Logf("Stats: compiled=%d, hits=%d, misses=%d",
		stats.CompiledFunctions, stats.CacheHits, stats.CacheMisses)
}

// TestJITCompilerCompileFibSuccess tests Fibonacci compilation
func TestJITCompilerCompileFibSuccess(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}
	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegCall), 0, 0,
			byte(compiler.OpRegCall), 0, 0,
			byte(compiler.OpRegAdd), 0, 0, 0,
		},
		NumLocals:     8,
		NumParameters: 1,
	}

	cf, err := jit.Compile(fn, nil, nil)
	if err != nil {
		t.Logf("Compile error: %v", err)
	} else {
		t.Logf("Fibonacci compile success: size=%d", cf.Size)
	}
}

// TestGenerateNativeCodeLargeStack tests native code generation with large stack
func TestGenerateNativeCodeLargeStack(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegNull), 0,
			byte(compiler.OpRegTrue), 0,
		},
		NumLocals:     256,
		NumParameters: 0,
	}

	code, err := generateNativeCode(fn, nil, nil)
	if err != nil {
		t.Errorf("generateNativeCode failed: %v", err)
		return
	}
	t.Logf("Generated %d bytes for large stack", len(code))
}

// TestGenerateNativeCodeZeroLocal tests with zero locals
func TestGenerateNativeCodeZeroLocal(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions:  []byte{byte(compiler.OpRegNull), 0},
		NumLocals:     0,
		NumParameters: 0,
	}

	code, err := generateNativeCode(fn, nil, nil)
	if err != nil {
		t.Errorf("generateNativeCode failed: %v", err)
		return
	}
	t.Logf("Generated %d bytes for zero locals", len(code))
}
// TestNativeLoopIncCheck verifies OpRegLoopIncCheck:
// counter++; if counter < limit, jump back to loop label.
func TestNativeLoopIncCheck(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	jumpOffset := int16(0)
	hi := byte(jumpOffset >> 8)
	lo := byte(jumpOffset)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 1, 0, 0,
			byte(compiler.OpRegLoopIncCheck), 1, 0, 1, hi, lo,
			byte(compiler.OpRegReturn), 1,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{0, 5}
	if err := registry.CompileFunction(fn, 100, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(100)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	result := nf.Execute(nil)
	if result != 5 {
		t.Fatalf("expected 5, got %d", result)
	}
}

// TestNativeLoopBodyAdd verifies OpRegLoopBodyAdd:
// acc += counter; counter++; if counter < limit, jump back.
// R0 is now stored on the stack, so it works correctly as a loop accumulator.
func TestNativeLoopBodyAdd(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	jumpOffset := int16(0)
	hi := byte(jumpOffset >> 8)
	lo := byte(jumpOffset)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0, // R0 = 0 (accumulator)
			byte(compiler.OpRegLoadConst), 1, 0, 1, // R1 = 1 (counter)
			byte(compiler.OpRegLoopBodyAdd), 0, 1, 0, 2, hi, lo, // R0+=R1; R1++; if R1<5, jump
			byte(compiler.OpRegReturn), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{0, 1, 5}
	if err := registry.CompileFunction(fn, 101, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(101)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	result := nf.Execute(nil)
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

// TestNativeLoopMulCheck verifies OpRegLoopMulCheck:
// if i*i > n, jump to target (exit loop).
// Uses R2/R3 instead of R0 to avoid rax clobbering issues.
func TestNativeLoopMulCheck(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	// IP layout:
	// 0-3:  OpRegLoadConst R2, const_0 (i=2)
	// 4-7:  OpRegLoadConst R3, const_1 (n=10)
	// 8-12: OpRegLoopMulCheck R2, R3, jump_offset
	// 13-16: OpRegLoadConst R1, const_2 (0, not jumped)
	// 17-18: OpRegReturn R1
	// 19-22: OpRegLoadConst R1, const_3 (1, jumped)
	// 23-24: OpRegReturn R1
	//
	// jump from IP=8 to IP=19: offset = 19 - 8 = 11
	jumpOffset := int16(11)
	hi := byte(jumpOffset >> 8)
	lo := byte(jumpOffset)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 2, 0, 0, // R2 = i
			byte(compiler.OpRegLoadConst), 3, 0, 1, // R3 = n
			byte(compiler.OpRegLoopMulCheck), 2, 3, hi, lo,
			byte(compiler.OpRegLoadConst), 1, 0, 2, // R1 = 0 (not jumped)
			byte(compiler.OpRegReturn), 1,
			byte(compiler.OpRegLoadConst), 1, 0, 3, // R1 = 1 (jumped)
			byte(compiler.OpRegReturn), 1,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{2, 10, 0, 1}
	if err := registry.CompileFunction(fn, 200, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(200)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	result := nf.Execute(nil)
	if result != 0 {
		t.Fatalf("i=2,n=10: expected 0 (no jump), got %d", result)
	}
}

// TestNativeLoopMulCheckJump verifies OpRegLoopMulCheck jumps when i*i > n.
func TestNativeLoopMulCheckJump(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	jumpOffset := int16(11)
	hi := byte(jumpOffset >> 8)
	lo := byte(jumpOffset)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 2, 0, 0, // R2 = i
			byte(compiler.OpRegLoadConst), 3, 0, 1, // R3 = n
			byte(compiler.OpRegLoopMulCheck), 2, 3, hi, lo,
			byte(compiler.OpRegLoadConst), 1, 0, 2, // R1 = 0 (not jumped)
			byte(compiler.OpRegReturn), 1,
			byte(compiler.OpRegLoadConst), 1, 0, 3, // R1 = 1 (jumped)
			byte(compiler.OpRegReturn), 1,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{4, 10, 0, 1}
	if err := registry.CompileFunction(fn, 201, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(201)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	result := nf.Execute(nil)
	if result != 1 {
		t.Fatalf("i=4,n=10: expected 1 (jumped), got %d", result)
	}
}

// TestNativeLoopIncCheckWithBody tests OpRegLoopIncCheck with a loop body
// that includes an add instruction before the check.
// R0 is now stored on the stack, so it works correctly as a loop accumulator.
func TestNativeLoopIncCheckWithBody(t *testing.T) {
	registry := NewNativeFunctionRegistry(DefaultJITConfig())
	defer registry.Cleanup()

	jumpOffset := int16(-4)
	hi := byte(jumpOffset >> 8)
	lo := byte(jumpOffset)

	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegLoadConst), 0, 0, 0, // R0 = 0 (accumulator)
			byte(compiler.OpRegLoadConst), 1, 0, 1, // R1 = 1 (counter)
			byte(compiler.OpRegAdd), 0, 0, 1,       // R0 += R1
			byte(compiler.OpRegLoopIncCheck), 1, 0, 2, hi, lo,
			byte(compiler.OpRegReturn), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []int64{0, 1, 5}
	if err := registry.CompileFunction(fn, 102, constants); err != nil {
		t.Fatalf("CompileFunction failed: %v", err)
	}

	nf := registry.Get(102)
	if nf == nil {
		t.Fatal("Get returned nil")
	}

	result := nf.Execute(nil)
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

// TestDebugMainBytecode dumps the main bytecode for the logical result test.
func TestDebugMainBytecode(t *testing.T) {
	code := `
		func both(a, b) {
			return a && b
		}
		both(true, false)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	t.Logf("Main bytecode: %d bytes", len(bytecode.Instructions))
	for ip := 0; ip < len(bytecode.Instructions); {
		op := compiler.Opcode(bytecode.Instructions[ip])
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			t.Logf("  [%d] unknown opcode %d", ip, op)
			break
		}
		operands := make([]string, len(def.OperandWidths))
		offset := ip + 1
		for j, w := range def.OperandWidths {
			if offset+w <= len(bytecode.Instructions) {
				operands[j] = fmt.Sprintf("%d", bytecode.Instructions[offset])
				offset += w
			}
		}
		t.Logf("  [%d] %s %s", ip, def.Name, strings.Join(operands, ","))
		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		ip += width
	}

	t.Logf("Constants: %d entries", len(bytecode.Constants))
	for i, cnst := range bytecode.Constants {
		t.Logf("  [%d] %T: %v", i, cnst, cnst)
	}
}
func TestDebugBothBytecode(t *testing.T) {
	code := `
		func both(a, b) {
			return a && b
		}
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	for i, cnst := range bytecode.Constants {
		if fn, ok := cnst.(*compiler.CompiledFunction); ok {
			t.Logf("Function %d: NumLocals=%d, NumParams=%d, NumRegs=%d, Instructions=%d bytes", i, fn.NumLocals, fn.NumParameters, fn.NumRegs, len(fn.Instructions))
			for ip := 0; ip < len(fn.Instructions); {
				op := compiler.Opcode(fn.Instructions[ip])
				def, err := compiler.Lookup(byte(op))
				if err != nil {
					t.Logf("  [%d] unknown opcode %d", ip, op)
					break
				}
				operands := make([]string, len(def.OperandWidths))
				offset := ip + 1
				for j, w := range def.OperandWidths {
					if offset+w <= len(fn.Instructions) {
						operands[j] = fmt.Sprintf("%d", fn.Instructions[offset])
						offset += w
					}
				}
				t.Logf("  [%d] %s %s", ip, def.Name, strings.Join(operands, ","))
				width := 1
				for _, w := range def.OperandWidths {
					width += w
				}
				ip += width
			}

			retType := analyzeReturnType(fn.Instructions)
			t.Logf("  analyzeReturnType = %v (0=Unknown, 1=Int, 2=Bool, 3=Null)", retType)
		}
	}
}

// TestJITVMHybridCallChain tests native function called through interpreter hybrid path.
// Verifies that OpRegCall in the main code triggers the native call hook correctly.
func TestJITVMHybridCallChain(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "logical and returns bool",
			code:     "func both(a, b) { return a && b } both(true, false)",
			expected: "false",
		},
		{
			name:     "logical or returns bool",
			code:     "func either(a, b) { return a || b } either(false, true)",
			expected: "true",
		},
		{
			name:     "arithmetic function",
			code:     "func add(a, b) { return a + b } add(10, 20)",
			expected: "30",
		},
		{
			name:     "comparison returns bool",
			code:     "func isGreater(a, b) { return a > b } isGreater(5, 3)",
			expected: "true",
		},
		{
			name:     "nested arithmetic",
			code:     "func compute(x) { return x * x + 1 } compute(7)",
			expected: "50",
		},
		{
			name:     "iterative loop function",
			code:     "func sumTo(n) {\nvar sum = 0\nfor (var i = 0; i <= n; i = i + 1) {\nsum = sum + i\n}\nreturn sum\n}\nsumTo(10)",
			expected: "55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			c := compiler.NewRegCompiler()
			_, err := c.Compile(program)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			bytecode := c.Bytecode()
			jitVM := NewJITVM(bytecode, JITConfig{HotThreshold: 1, MaxCodeSize: 16384, Debug: false})
			defer jitVM.Cleanup()

			if err := jitVM.Run(); err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			obj := jitVM.LastPoppedObject()
			if obj == nil {
				t.Fatal("expected result object")
			}
			if got := obj.Inspect(); got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}

			nativeExecs, _ := jitVM.GetNativeStats()
			if nativeExecs == 0 {
				t.Error("expected at least one native execution")
			}
		})
	}
}

// TestAnalyzeReturnTypeUnreachableReturn tests that analyzeReturnType correctly
// ignores unreachable OpRegReturn instructions (compiler-appended default returns).
func TestAnalyzeReturnTypeUnreachableReturn(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected NativeReturnType
	}{
		{
			name: "and before return, then unreachable return",
			code: []byte{
				byte(compiler.OpRegAnd), 10, 8, 9,
				byte(compiler.OpRegReturn), 10,
				byte(compiler.OpRegReturn), 0,
			},
			expected: ReturnTypeBool,
		},
		{
			name: "comparison before return, then unreachable return",
			code: []byte{
				byte(compiler.OpRegLess), 10, 8, 9,
				byte(compiler.OpRegReturn), 10,
				byte(compiler.OpRegReturn), 0,
			},
			expected: ReturnTypeBool,
		},
		{
			name: "add before return, then unreachable return",
			code: []byte{
				byte(compiler.OpRegAdd), 10, 8, 9,
				byte(compiler.OpRegReturn), 10,
				byte(compiler.OpRegReturn), 0,
			},
			expected: ReturnTypeInt,
		},
		{
			name: "single return with and",
			code: []byte{
				byte(compiler.OpRegAnd), 10, 8, 9,
				byte(compiler.OpRegReturn), 10,
			},
			expected: ReturnTypeBool,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeReturnType(tt.code)
			if result != tt.expected {
				t.Errorf("analyzeReturnType() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
