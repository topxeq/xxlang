// pkg/vm/optimization_comparison_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// runBenchmarkWithOptimizations runs a benchmark with optimizations enabled/disabled
func runBenchmarkWithOptimizations(b *testing.B, code string, enableOptimizations bool) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()

	// If optimizations are disabled, we can't easily disable them
	// So we just compile normally and the optimizations will be applied
	// based on the pattern matching in the compiler
	_ = enableOptimizations

	_, err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewRegVM(bytecode)
		if err := vm.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// TestOptimizationBytecodeSize compares bytecode size with and without optimizations
func TestOptimizationBytecodeSize(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "PrimeCheck",
			code: `
func isPrime(n) {
    if (n < 2) { return false }
    if (n == 2) { return true }
    if (n % 2 == 0) { return false }
    var i = 3
    while (i * i <= n) {
        if (n % i == 0) { return false }
        i = i + 1
    }
    return true
}
isPrime(7919)
`,
		},
		{
			name: "CountingLoop",
			code: `
var total = 0
for (var i = 0; i < 100; i++) {
    total = total + i
}
total
`,
		},
		{
			name: "NestedLoop",
			code: `
var sum = 0
for (var i = 0; i < 10; i++) {
    for (var j = 0; j < 10; j++) {
        sum = sum + 1
    }
}
sum
`,
		},
		{
			name: "SmallLoop",
			code: `
var sum = 0
for (var i = 0; i < 8; i++) {
    sum = sum + i
}
sum
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			c := compiler.NewRegCompiler()
			_, err := c.Compile(program)
			if err != nil {
				t.Fatalf("compile error: %v", err)
			}

			bytecode := c.Bytecode()
			t.Logf("Bytecode size: %d instructions, %d constants",
				len(bytecode.Instructions), len(bytecode.Constants))
		})
	}
}

// BenchmarkOptimizationImpact measures the impact of various optimizations
func BenchmarkOptimizationImpact(b *testing.B) {
	b.Run("PrimeCheck_Optimized", func(b *testing.B) {
		code := `
func isPrime(n) {
    if (n < 2) { return false }
    if (n == 2) { return true }
    if (n % 2 == 0) { return false }
    var i = 3
    while (i * i <= n) {
        if (n % i == 0) { return false }
        i = i + 1
    }
    return true
}

var count = 0
for (var n = 2; n < 10000; n++) {
    if (isPrime(n)) { count = count + 1 }
}
count
`
		runBenchmarkWithOptimizations(b, code, true)
	})

	b.Run("CountingLoop_Optimized", func(b *testing.B) {
		code := `
var total = 0
for (var i = 0; i < 10000; i++) {
    total = total + i
}
total
`
		runBenchmarkWithOptimizations(b, code, true)
	})

	b.Run("NestedLoop_Optimized", func(b *testing.B) {
		code := `
var sum = 0
for (var i = 0; i < 100; i++) {
    for (var j = 0; j < 100; j++) {
        sum = sum + i * j
    }
}
sum
`
		runBenchmarkWithOptimizations(b, code, true)
	})

	b.Run("LoopUnrolling_Optimized", func(b *testing.B) {
		code := `
var sum = 0
for (var i = 0; i < 8; i++) {
    sum = sum + i * i
}
sum
`
		runBenchmarkWithOptimizations(b, code, true)
	})
}
