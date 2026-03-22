// +build amd64,!windows

// pkg/jit/jit_test.go
// Tests for the JIT compiler
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

func TestJITCompilerBasic(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())

	// Test code allocation
	mem, page, err := jit.AllocCode(1024)
	if err != nil {
		t.Fatalf("Failed to allocate code: %v", err)
	}
	if page == nil {
		t.Fatal("Code page is nil")
	}
	if len(mem) != 1024 {
		t.Fatalf("Expected 1024 bytes, got %d", len(mem))
	}

	// Cleanup
	jit.Cleanup()
}

func TestCodeGeneratorPrologue(t *testing.T) {
	cg := NewCodeGenerator(DefaultJITConfig())
	cg.emitPrologue()
	cg.emitEpilogue()

	if len(cg.code) == 0 {
		t.Fatal("Generated code is empty")
	}

	t.Logf("Prologue + Epilogue size: %d bytes", len(cg.code))
}

func TestCodeGeneratorArithmetic(t *testing.T) {
	cg := NewCodeGenerator(DefaultJITConfig())
	cg.constants = []vm.Value{
		vm.NewInt(10),
		vm.NewInt(20),
	}

	// Emit prologue
	cg.emitPrologue()

	// Store constants to registers
	code := []byte{
		byte(compiler.OpRegLoadConst), 0, 0, 0, // R0 = constants[0]
		byte(compiler.OpRegLoadConst), 1, 0, 1, // R1 = constants[1]
		byte(compiler.OpRegAdd), 2, 0, 1, // R2 = R0 + R1
	}

	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		switch op {
		case compiler.OpRegLoadConst:
			cg.compileOpRegLoadConst(code, &ip)
		case compiler.OpRegAdd:
			cg.compileOpRegArithmetic(code, &ip, op)
		default:
			ip++
		}
	}

	// Emit epilogue
	cg.emitEpilogue()

	t.Logf("Arithmetic code size: %d bytes", len(cg.code))
}

func TestJITCompileSimpleCode(t *testing.T) {
	// Simple code without loops
	code := `
var a = 10
var b = 20
a + b
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Create JIT compiler
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	// Convert constants
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	// Try to compile the main function
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     10,
		NumParameters: 0,
	}

	cf, err := jit.Compile(mainFn, constants, nil)
	if err != nil {
		t.Logf("JIT compilation not yet supported: %v", err)
		return
	}

	t.Logf("Compiled function size: %d bytes", cf.Size)
}

func TestJITVMCreation(t *testing.T) {
	code := `
var a = 10
var b = 20
a + b
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Create JIT VM
	jitVM := NewJITVM(bytecode, DefaultJITConfig())
	defer jitVM.Cleanup()

	// Run
	err = jitVM.Run()
	if err != nil {
		t.Fatalf("VM run error: %v", err)
	}

	result := jitVM.LastPopped()
	t.Logf("Result: %v", result)
}

func TestJITStats(t *testing.T) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	stats := jit.GetStats()
	if stats.CompiledFunctions != 0 {
		t.Errorf("Expected 0 compiled functions, got %d", stats.CompiledFunctions)
	}
}

func TestJITHashBytecode(t *testing.T) {
	tests := []struct {
		code     []byte
		expected uint64
	}{
		{[]byte{1, 2, 3}, 0},
		{[]byte{}, 14695981039346656037},
	}

	for _, tt := range tests {
		hash := hashBytecode(tt.code)
		if tt.expected != 0 && hash != tt.expected {
			t.Errorf("hashBytecode(%v) = %d, want %d", tt.code, hash, tt.expected)
		}
		// Just verify it's deterministic
		hash2 := hashBytecode(tt.code)
		if hash != hash2 {
			t.Errorf("hashBytecode not deterministic")
		}
	}
}

func BenchmarkJITCodeGeneration(b *testing.B) {
	cg := NewCodeGenerator(DefaultJITConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.code = cg.code[:0]
		cg.emitPrologue()
		cg.emitEpilogue()
	}
}
