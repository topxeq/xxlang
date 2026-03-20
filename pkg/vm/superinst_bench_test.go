// pkg/vm/superinst_bench_test.go
// Benchmarks for superinstruction performance
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// BenchmarkGetLocalConstSuperInstructions benchmarks the new GetLocal+Constant superinstructions
func BenchmarkGetLocalConstAdd(b *testing.B) {
	input := `
		var i = 10
		var result = 0
		for (var n = 0; n < 1000; n++) {
			result = i + 5
		}
		result
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

func BenchmarkGetLocalConstLess(b *testing.B) {
	input := `
		var count = 0
		for (var i = 0; i < 1000; i++) {
			count = count + 1
		}
		count
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

func BenchmarkGetLocalConstMul(b *testing.B) {
	input := `
		var i = 7
		var result = 0
		for (var n = 0; n < 1000; n++) {
			result = i * 3
		}
		result
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

// BenchmarkLoopWithSuperInstructions tests a realistic loop scenario
func BenchmarkLoopWithSuperInstructions(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 100; i++) {
			if (i < 50) {
				sum = sum + i
			}
			if (i > 25) {
				sum = sum + 1
			}
		}
		sum
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

// BenchmarkNestedLoopWithSuperInstructions tests nested loops
func BenchmarkNestedLoopWithSuperInstructions(b *testing.B) {
	input := `
		var total = 0
		for (var i = 0; i < 10; i++) {
			for (var j = 0; j < 10; j++) {
				if (i < 5) {
					if (j < 5) {
						total = total + 1
					}
				}
			}
		}
		total
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

// BenchmarkArithmeticLoop tests arithmetic-heavy loop
func BenchmarkArithmeticLoop(b *testing.B) {
	input := `
		var result = 0
		for (var i = 0; i < 100; i++) {
			result = result + i
			result = result * 2
			result = result - i
		}
		result
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runBenchVM(input)
		_ = vm
	}
}

// runBenchVM is a helper for benchmarks
func runBenchVM(input string) *VM {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.New()
	c.Compile(program)

	vm := New(c.Bytecode())
	vm.Run()
	return vm
}
