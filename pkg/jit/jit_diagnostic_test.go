// +build amd64,!windows

// Diagnostic test to verify JIT execution
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// TestDiagnosticCanExecuteNatively tests why JIT is not executing
func TestDiagnosticCanExecuteNatively(t *testing.T) {
	// Tail-recursive Fibonacci
	tailRecCode := `
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

	l := lexer.New(tailRecCode)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	// Check the main code
	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	canNative := CanExecuteNatively(mainFn)
	t.Logf("Main code can execute natively: %v", canNative)
	t.Logf("Main code length: %d bytes", len(bytecode.Instructions))

	// Analyze opcodes in main code
	opCount := make(map[string]int)
	code := bytecode.Instructions
	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		def, err := compiler.Lookup(byte(op))
		if err != nil {
			t.Fatalf("Unknown opcode at %d", ip)
		}
		opCount[def.Name]++

		ip++
		for _, w := range def.OperandWidths {
			ip += w
		}
	}

	t.Logf("Opcode distribution in main code:")
	for op, count := range opCount {
		t.Logf("  %s: %d", op, count)
	}

	// Check each function
	for i, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			canFn := CanExecuteNatively(fn)
			t.Logf("Function %d (params=%d, locals=%d): canNative=%v, instructions=%d bytes",
				i, fn.NumParameters, fn.NumLocals, canFn, len(fn.Instructions))

			// Analyze opcodes in function
			fnOpCount := make(map[string]int)
			fnCode := fn.Instructions
			fnIP := 0
			for fnIP < len(fnCode) {
				op := compiler.Opcode(fnCode[fnIP])
				def, err := compiler.Lookup(byte(op))
				if err != nil {
					break
				}
				fnOpCount[def.Name]++

				fnIP++
				for _, w := range def.OperandWidths {
					fnIP += w
				}
			}

			for op, count := range fnOpCount {
				t.Logf("    %s: %d", op, count)
			}
		}
	}
}

// TestSimpleArithmeticJIT tests JIT on simple arithmetic that should work
func TestSimpleArithmeticJIT(t *testing.T) {
	// Simple arithmetic that should be natively executable
	code := `
		var a = 10
		var b = 20
		var c = a + b
		c
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.NewRegCompiler()
	c.Compile(program)
	bytecode := c.Bytecode()

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	canNative := CanExecuteNatively(mainFn)
	t.Logf("Simple arithmetic can execute natively: %v", canNative)

	if !canNative {
		// Find which opcode causes the issue
		code := bytecode.Instructions
		ip := 0
		for ip < len(code) {
			op := compiler.Opcode(code[ip])
			def, _ := compiler.Lookup(byte(op))
			t.Logf("Opcode: %s", def.Name)

			ip++
			for _, w := range def.OperandWidths {
				ip += w
			}
		}
	}
}
