// Simple JIT test
package jit

import (
	"testing"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit/bridge"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestSimpleJIT tests the most basic JIT code
func TestSimpleJIT(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	// Generate simple code: return 42
	// Use a simple hand-crafted assembly
	// mov rax, 42; ret
	code := []byte{
		0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00, // mov rax, 42
		0xC3, // ret
	}

	t.Logf("Simple code: %x", code)

	jit := NewJITCompiler(config)
	defer jit.Cleanup()

	mem, _, err := jit.AllocCode(len(code))
	if err != nil {
		t.Fatalf("Memory allocation failed: %v", err)
	}

	copy(mem, code)
	fnPtr := (*byte)(unsafe.Pointer(&mem[0]))

	result := bridge.Call0(fnPtr)
	t.Logf("Result: %d (expected 42)", result)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestInterpreterOnly tests only the interpreter
func TestInterpreterOnly(t *testing.T) {
	// Tail-recursive Fibonacci code
	code := `
		func fibHelper(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibHelper(n - 1, b, a + b)
		}
		func fib(n) {
			return fibHelper(n, 0, 1)
		}
		fib(35)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Measure interpreter performance
	iterations := 5
	start := time.Now()
	for i := 0; i < iterations; i++ {
		vmInst := vm.NewRegVM(bytecode)
		vmInst.Run()
	}
	interpreterTime := time.Since(start) / time.Duration(iterations)

	result := vm.NewRegVM(bytecode)
	result.Run()
	t.Logf("Interpreter: %v per iteration, result=%v", interpreterTime, result.LastPoppedObject().Inspect())

	// Expected: 9227465
	if result.LastPoppedObject().Inspect() != "9227465" {
		t.Errorf("Wrong result: %v", result.LastPoppedObject().Inspect())
	}
}

// TestNativeCodeGeneratorSimple tests the native code generator with simple code
func TestNativeCodeGeneratorSimple(t *testing.T) {
	// Create a simple function that returns a constant
	// OpRegLoadConst dst=0, constIdx=0 (value 42)
	// OpRegReturn src=0
	// We need to manually create the bytecode

	// First, compile simple code
	code := `
		var x = 42
		x
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	t.Logf("Bytecode length: %d", len(bytecode.Instructions))
	t.Logf("Constants: %d", len(bytecode.Constants))

	// Check if this is simple enough
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	canNative := CanExecuteNatively(mainFn)
	t.Logf("Can execute natively: %v", canNative)

	if canNative {
		config := JITConfig{
			HotThreshold: 1,
			MaxCodeSize:  16384,
			Debug:        true,
		}

		jitVM := NewJITVM(bytecode, config)
		defer jitVM.Cleanup()

		err := jitVM.Run()
		if err != nil {
			t.Fatalf("JITVM run failed: %v", err)
		}

		result := jitVM.LastPoppedObject()
		t.Logf("JITVM result: %v", result.Inspect())

		nativeExecs, interpExecs := jitVM.GetNativeStats()
		t.Logf("Native executions: %d, Interpreter executions: %d", nativeExecs, interpExecs)
	}
}
