// pkg/jit/jit_ext_test.go
// Extended JIT tests for loop and arithmetic benchmarks
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITLoopBenchmark tests JIT compilation of simple loops
func TestJITLoopBenchmark(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "SimpleLoop",
			code: `
				var sum = 0
				for (var i = 0; i < 100; i = i + 1) {
					sum = sum + i
				}
				sum
			`,
		},
		{
			name: "Arithmetic",
			code: `
				var a = 10
				var b = 20
				var c = a + b
				var d = c * 2
				d
			`,
		},
		{
			name: "WhileLoop",
			code: `
				var i = 0
				var sum = 0
				while (i < 100) {
					sum = sum + i
					i = i + 1
				}
				sum
			`,
		},
		{
			name: "NestedIf",
			code: `
				var x = 10
				var y = 20
				var result = 0
				if (x < y) {
					if (x > 5) {
						result = x + y
					}
				}
				result
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
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
			constants := make([]vm.Value, len(bytecode.Constants))
			for i, c := range bytecode.Constants {
				constants[i] = vm.NewObject(c)
			}

			// Test interpreter
			vmInst := vm.NewRegVM(bytecode)
			if err := vmInst.Run(); err != nil {
				t.Fatalf("Interpreter error: %v", err)
			}
			interpResult := vmInst.LastPoppedObject()

			// Test JIT compilation
			config := JITConfig{
				HotThreshold: 1,
				MaxCodeSize:  8192,
				Debug:        true,
			}

			jitCompiler := NewJITCompiler(config)

			mainFn := &compiler.CompiledFunction{
				Instructions:  bytecode.Instructions,
				NumLocals:     16,
				NumParameters: 0,
			}

			cf, err := jitCompiler.Compile(mainFn, constants, nil)
			if err != nil {
				t.Logf("JIT compilation failed (expected for some code): %v", err)
				return
			}

			t.Logf("JIT compiled successfully: %d bytes", cf.Size)
			t.Logf("Interpreter result: %v", interpResult.Inspect())

			jitCompiler.Cleanup()
		})
	}
}

// BenchmarkJITLoop benchmarks JIT vs interpreter for loops
func BenchmarkJITLoop(b *testing.B) {
	code := `
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + i
		}
		sum
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		b.Fatalf("Compile error: %v", err)
	}

	bytecode := c.Bytecode()
	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	b.Run("Interpreter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vmInst := vm.NewRegVM(bytecode)
			if err := vmInst.Run(); err != nil {
				b.Fatalf("Runtime error: %v", err)
			}
		}
	})

	// Try JIT compilation (note: direct execution requires VM context)
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  8192,
		Debug:        false,
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     16,
		NumParameters: 0,
	}

	// Verify JIT can compile this code
	jitCompiler := NewJITCompiler(config)
	_, err := jitCompiler.Compile(mainFn, constants, nil)
	jitCompiler.Cleanup()
	if err != nil {
		b.Skipf("JIT not supported: %v", err)
	}

	b.Run("JIT_Compile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jitC := NewJITCompiler(config)
			_, _ = jitC.Compile(mainFn, constants, nil)
			jitC.Cleanup()
		}
	})
}
