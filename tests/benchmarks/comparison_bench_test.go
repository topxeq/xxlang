// tests/benchmarks/comparison_bench_test.go
// Comprehensive benchmarks comparing xxlang Register VM with Go and Python
package benchmarks

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// Helper to run xxlang code
func runXxlang(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil
	}
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		return err
	}
	v := vm.NewRegVM(c.Bytecode())
	return v.Run()
}

// ============================================
// Fibonacci Benchmarks
// ============================================

func BenchmarkXxlang_Fibonacci15(b *testing.B) {
	input := `
		func fib(n) {
			if (n < 2) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(15)
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_Fibonacci20(b *testing.B) {
	input := `
		func fib(n) {
			if (n < 2) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(20)
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_FibonacciIterative(b *testing.B) {
	input := `
		func fib(n) {
			var a = 0
			var b = 1
			for (var i = 0; i < n; i = i + 1) {
				var temp = a + b
				a = b
				b = temp
			}
			return a
		}
		fib(30)
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Factorial Benchmark
// ============================================

func BenchmarkXxlang_Factorial(b *testing.B) {
	input := `
		func fact(n) {
			if (n <= 1) { return 1 }
			return n * fact(n - 1)
		}
		fact(12)
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Loop Benchmarks
// ============================================

func BenchmarkXxlang_ForLoop100(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = sum + i
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_ForLoop1000(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + i
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_WhileLoop(b *testing.B) {
	input := `
		var i = 0
		var sum = 0
		while (i < 100) {
			sum = sum + i
			i = i + 1
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Function Call Benchmarks
// ============================================

func BenchmarkXxlang_FunctionCalls(b *testing.B) {
	input := `
		func add(a, b) { return a + b }
		func mul(a, b) { return a * b }
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = add(sum, i)
			sum = mul(sum, 1)
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Arithmetic Benchmarks
// ============================================

func BenchmarkXxlang_Arithmetic(b *testing.B) {
	input := `
		var a = 10
		var b = 20
		var c = a + b
		var d = c * 2
		var e = d - 5
		var f = e / 3
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_IntensiveArithmetic(b *testing.B) {
	input := `
		var x = 0
		for (var i = 0; i < 1000; i = i + 1) {
			x = x + i * 2 - i / 2
			x = x % 1000
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_NestedExpressions(b *testing.B) {
	input := `
		var a = 1 + 2 * 3 - 4 / 2
		var c = 10
		var d = (a + 5) * (c - a) / (5 + 1)
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Comparison Benchmarks
// ============================================

func BenchmarkXxlang_Comparisons(b *testing.B) {
	input := `
		var a = 10
		var b = 20
		var c = 0
		for (var i = 0; i < 100; i = i + 1) {
			if (a < b) { c = c + 1 }
			if (a > b) { c = c + 1 }
			if (a == b) { c = c + 1 }
			if (a != b) { c = c + 1 }
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Algorithm Benchmarks
// ============================================

func BenchmarkXxlang_PrimeCheck100(b *testing.B) {
	input := `
		func isPrime(n) {
			if (n < 2) { return false }
			var i = 2
			while (i * i <= n) {
				if (n % i == 0) { return false }
				i = i + 1
			}
			return true
		}
		for (var j = 0; j < 100; j = j + 1) {
			isPrime(j)
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_BubbleSort10(b *testing.B) {
	input := `
		func bubbleSort(arr) {
			var n = len(arr)
			for (var i = 0; i < n - 1; i = i + 1) {
				for (var j = 0; j < n - i - 1; j = j + 1) {
					if (arr[j] > arr[j + 1]) {
						var temp = arr[j]
						arr[j] = arr[j + 1]
						arr[j + 1] = temp
					}
				}
			}
		}
		bubbleSort([5, 3, 8, 4, 2, 7, 1, 10, 6, 9])
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// ============================================
// Additional Benchmarks for Real-world Scenarios
// ============================================

func BenchmarkXxlang_ArraySum1000(b *testing.B) {
	input := `
		var arr = []
		for (var i = 0; i < 1000; i = i + 1) {
			arr[i] = i
		}
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + arr[i]
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

func BenchmarkXxlang_StringConcat100(b *testing.B) {
	input := `
		var s = ""
		for (var i = 0; i < 100; i = i + 1) {
			s = s + "x"
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runXxlang(input)
	}
}

// Go versions of the additional benchmarks

func BenchmarkGo_ArraySum1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := make([]int, 1000)
		for j := 0; j < 1000; j++ {
			arr[j] = j
		}
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += arr[j]
		}
	}
}

func BenchmarkGo_StringConcat100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 100; j++ {
			s += "x"
		}
	}
}
