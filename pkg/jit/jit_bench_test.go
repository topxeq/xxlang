// +build amd64,!windows

// pkg/jit/jit_bench_test.go
// Performance benchmarks for JIT compiler
package jit

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestJITBytecodeAnalysis analyzes bytecode to identify unsupported opcodes
func TestJITBytecodeAnalysis(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "SimpleArithmetic",
			code: `
				var a = 10
				var b = 20
				var c = a + b
				c
			`,
		},
		{
			name: "CountingLoop100",
			code: `
				var total = 0
				for (var i = 0; i < 100; i++) {
					total = total + i
				}
				total
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
			_, err := c.Compile(program)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			bytecode := c.Bytecode()

			fmt.Printf("\n=== %s ===\n", tt.name)
			fmt.Printf("Bytecode length: %d bytes\n", len(bytecode.Instructions))
			fmt.Printf("Constants: %d\n", len(bytecode.Constants))

			// Analyze opcodes
			ip := 0
			code := bytecode.Instructions
			unsupported := make(map[compiler.Opcode]int)

			for ip < len(code) {
				op := compiler.Opcode(code[ip])
				def, err := compiler.Lookup(byte(op))
				if err != nil {
					t.Fatalf("Unknown opcode %d at IP %d", op, ip)
				}

				// Check if supported
				supported := isOpcodeSupported(op)
				if !supported {
					unsupported[op]++
				}

				// Print instruction
				fmt.Printf("%4d: %s (op=%d)", ip, def.Name, op)
				if !supported {
					fmt.Printf(" [UNSUPPORTED]")
				}
				fmt.Println()

				// Skip operands
				ip++
				for _, w := range def.OperandWidths {
					if w == 1 {
						fmt.Printf("      operand: %d\n", code[ip])
					} else if w == 2 {
						fmt.Printf("      operand: %d\n", int(code[ip])<<8|int(code[ip+1]))
					}
					ip += w
				}
			}

			if len(unsupported) > 0 {
				t.Logf("Unsupported opcodes: %v", unsupported)
			}
		})
	}
}

func isOpcodeSupported(op compiler.Opcode) bool {
	supported := map[compiler.Opcode]bool{
		compiler.OpRegLoadConst:    true,
		compiler.OpRegLoadGlobal:   true,
		compiler.OpRegStoreGlobal:  true,
		compiler.OpRegMove:         true,
		compiler.OpRegAdd:          true,
		compiler.OpRegSub:          true,
		compiler.OpRegMul:          true,
		compiler.OpRegDiv:          true,
		compiler.OpRegMod:          true,
		compiler.OpRegNeg:          true,
		compiler.OpRegLess:         true,
		compiler.OpRegGreater:      true,
		compiler.OpRegEqual:        true,
		compiler.OpRegNotEqual:     true,
		compiler.OpRegLessEqual:    true,
		compiler.OpRegGreaterEqual: true,
		compiler.OpRegNot:          true,
		compiler.OpRegJump:         true,
		compiler.OpRegJumpIfTrue:   true,
		compiler.OpRegJumpIfFalse:  true,
		compiler.OpRegReturn:       true,
		compiler.OpRegNull:         true,
		compiler.OpRegTrue:         true,
		compiler.OpRegFalse:        true,
		compiler.OpRegLoadLocal:    true,
		compiler.OpRegStoreLocal:   true,
		compiler.OpRegIncLocal:     true,
		compiler.OpRegDecLocal:     true,
		compiler.OpRegLoopCountAdd: true,
		compiler.OpRegLoopBodyAdd:  true,
	}
	return supported[op]
}

// BenchmarkJITVsInterpreter compares JIT vs interpreter performance
func BenchmarkJITVsInterpreter(b *testing.B) {
	// Only test code that we know is fully supported
	tests := []struct {
		name string
		code string
	}{
		{
			name: "SimpleArithmetic",
			code: `
				var a = 10
				var b = 20
				var c = a + b
				c
			`,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Compile once
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				b.Fatalf("Parse errors: %v", p.Errors())
			}

			c := compiler.NewRegCompiler()
			_, err := c.Compile(program)
			if err != nil {
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
		})
	}
}

// BenchmarkJITCompiles measures JIT compilation speed
func BenchmarkJITCompiles(b *testing.B) {
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

	constants := make([]vm.Value, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		constants[i] = vm.NewObject(c)
	}

	mainFn := &compiler.CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     10,
		NumParameters: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit := NewJITCompiler(DefaultJITConfig())
		jit.Compile(mainFn, constants, nil)
		jit.Cleanup()
	}
}

// BenchmarkJITMemoryAllocation measures memory allocation speed
func BenchmarkJITMemoryAllocation(b *testing.B) {
	jit := NewJITCompiler(DefaultJITConfig())
	defer jit.Cleanup()

	sizes := []int{64, 256, 1024, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _, err := jit.AllocCode(size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
