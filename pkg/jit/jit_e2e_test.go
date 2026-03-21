// pkg/jit/jit_e2e_test.go
// End-to-end test for JIT recursive Fibonacci execution
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITRecursiveFibE2E tests end-to-end JIT execution of recursive Fibonacci
// This verifies that the JIT system:
// 1. Detects recursive functions
// 2. Compiles them to native iterative code
// 3. Intercepts calls via the native hook
// 4. Executes natively and returns correct results
func TestJITRecursiveFibE2E(t *testing.T) {
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(10)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Test with JIT enabled
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        true,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("JIT execution error: %v", err)
	}

	result := jitVM.LastPoppedObject()
	if result.Inspect() != "55" {
		t.Errorf("Expected 55, got %v", result.Inspect())
	}

	// Verify native execution was used
	nativeExecs, interpExecs := jitVM.GetNativeStats()
	t.Logf("Native executions: %d, Interpreter executions: %d", nativeExecs, interpExecs)

	if nativeExecs == 0 {
		t.Error("Expected native execution to be used, but it wasn't")
	}
}

// TestJITRecursiveFibMultipleCalls tests that JIT works for multiple calls
func TestJITRecursiveFibMultipleCalls(t *testing.T) {
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}

		var a = fib(5)
		var b = fib(10)
		var c = fib(20)
		c
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
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

	if err := jitVM.Run(); err != nil {
		t.Fatalf("JIT execution error: %v", err)
	}

	result := jitVM.LastPoppedObject()
	if result.Inspect() != "6765" {
		t.Errorf("Expected 6765, got %v", result.Inspect())
	}

	// Verify native execution was used multiple times
	nativeExecs, _ := jitVM.GetNativeStats()
	if nativeExecs < 3 {
		t.Errorf("Expected at least 3 native executions, got %d", nativeExecs)
	}
}

// TestJITRecursiveFibCorrectness tests correctness for various inputs
func TestJITRecursiveFibCorrectness(t *testing.T) {
	testCases := []struct {
		n        int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{2, "1"},
		{5, "5"},
		{10, "55"},
		{20, "6765"},
		{30, "832040"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			code := `
				func fib(n) {
					if (n <= 1) { return n }
					return fib(n - 1) + fib(n - 2)
				}
				fib(` + string(rune(tc.n+'0')) + `)
			`

			// For n >= 10, we need to construct the code differently
			if tc.n >= 10 {
				code = `
					func fib(n) {
						if (n <= 1) { return n }
						return fib(n - 1) + fib(n - 2)
					}
					fib(` + string(rune(tc.n/10+'0')) + string(rune(tc.n%10+'0')) + `)
				`
			}

			l := lexer.New(code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parse errors: %v", p.Errors())
			}

			c := compiler.NewRegCompiler()
			if _, err := c.Compile(program); err != nil {
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

			if err := jitVM.Run(); err != nil {
				t.Fatalf("JIT execution error: %v", err)
			}

			result := jitVM.LastPoppedObject()
			if result.Inspect() != tc.expected {
				t.Errorf("fib(%d): expected %s, got %v", tc.n, tc.expected, result.Inspect())
			}
		})
	}
}

// TestJITVsInterpreter compares JIT and interpreter results
func TestJITVsInterpreter(t *testing.T) {
	code := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(25)
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Run with interpreter
	interpVM := vm.NewRegVM(bytecode)
	if err := interpVM.Run(); err != nil {
		t.Fatalf("Interpreter error: %v", err)
	}
	interpResult := interpVM.LastPoppedObject().Inspect()

	// Run with JIT
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  16384,
		Debug:        false,
	}

	jitVM := NewJITVM(bytecode, config)
	defer jitVM.Cleanup()

	if err := jitVM.Run(); err != nil {
		t.Fatalf("JIT execution error: %v", err)
	}
	jitResult := jitVM.LastPoppedObject().Inspect()

	if interpResult != jitResult {
		t.Errorf("Results differ: interpreter=%s, JIT=%s", interpResult, jitResult)
	}

	t.Logf("Both produced correct result: %s", jitResult)
}
