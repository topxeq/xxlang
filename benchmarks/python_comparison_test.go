package benchmarks

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// BenchmarkXxlangVsPython compares Xxlang performance with Python
// These benchmarks use equivalent Xxlang code to compare with pure Python

func BenchmarkXxlangFibRecursive10(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(10)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFibRecursive15(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(15)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFibRecursive20(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) { return n }
			return fib(n - 1) + fib(n - 2)
		}
		fib(20)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFibIterative10(b *testing.B) {
	input := `
		func fib(n) {
			var a = 0
			var b = 1
			for (var i = 0; i < n; i++) {
				var temp = a
				a = b
				b = temp + b
			}
			return a
		}
		fib(10)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFibIterative15(b *testing.B) {
	input := `
		func fib(n) {
			var a = 0
			var b = 1
			for (var i = 0; i < n; i++) {
				var temp = a
				a = b
				b = temp + b
			}
			return a
		}
		fib(15)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFibIterative1000(b *testing.B) {
	input := `
		func fib(n) {
			var a = 0
			var b = 1
			for (var i = 0; i < n; i++) {
				var temp = a
				a = b
				b = temp + b
			}
			return a
		}
		fib(1000)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFactorial10(b *testing.B) {
	input := `
		func fact(n) {
			var result = 1
			for (var i = 2; i <= n; i++) {
				result = result * i
			}
			return result
		}
		fact(10)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFactorial20(b *testing.B) {
	input := `
		func fact(n) {
			var result = 1
			for (var i = 2; i <= n; i++) {
				result = result * i
			}
			return result
		}
		fact(20)
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangForLoop100(b *testing.B) {
	input := `
		var total = 0
		for (var i = 0; i < 100; i++) {
			total = total + i
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangForLoop1000(b *testing.B) {
	input := `
		var total = 0
		for (var i = 0; i < 1000; i++) {
			total = total + i
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangWhileLoop100(b *testing.B) {
	input := `
		var total = 0
		var i = 0
		while (i < 100) {
			total = total + i
			i = i + 1
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangWhileLoop1000(b *testing.B) {
	input := `
		var total = 0
		var i = 0
		while (i < 1000) {
			total = total + i
			i = i + 1
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangNestedLoop10(b *testing.B) {
	input := `
		var total = 0
		for (var i = 0; i < 10; i++) {
			for (var j = 0; j < 10; j++) {
				total = total + i * j
			}
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangPrimeCheck100(b *testing.B) {
	input := `
		var count = 0
		for (var n = 2; n <= 100; n++) {
			var isPrime = true
			for (var i = 2; i * i <= n; i++) {
				if (n % i == 0) {
					isPrime = false
					break
				}
			}
			if (isPrime) {
				count = count + 1
			}
		}
		count
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangBubbleSort10(b *testing.B) {
	input := `
		var arr = [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]
		for (var i = 0; i < 10; i++) {
			for (var j = 0; j < 10 - i - 1; j++) {
				if (arr[j] > arr[j + 1]) {
					var temp = arr[j]
					arr[j] = arr[j + 1]
					arr[j + 1] = temp
				}
			}
		}
		arr
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangStringConcat100(b *testing.B) {
	input := `
		var result = ""
		for (var i = 0; i < 100; i++) {
			result = result + toStr(i)
		}
		result
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangArraySum1000(b *testing.B) {
	input := `
		var total = 0
		for (var i = 0; i < 1000; i++) {
			total = total + i
		}
		total
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangArithmeticOps100(b *testing.B) {
	input := `
		var result = 0
		for (var i = 0; i < 100; i++) {
			result = result + i
			result = result * 2
			result = result - i
		}
		result
	`
	benchmarkXxlang(b, input)
}

func BenchmarkXxlangFunctionCalls100(b *testing.B) {
	input := `
		func add(a, b) { return a + b }
		func mul(a, b) { return a * b }
		var result = 0
		for (var i = 0; i < 100; i++) {
			result = add(result, i)
			result = mul(result, 1)
		}
		result
	`
	benchmarkXxlang(b, input)
}

func benchmarkXxlang(b *testing.B, input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		if err := v.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// TestPythonComparison runs all benchmarks and prints a comparison
func TestPythonComparison(t *testing.T) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"FibRecursive10", `
			func fib(n) {
				if (n <= 1) { return n }
				return fib(n - 1) + fib(n - 2)
			}
			fib(10)
		`},
		{"FibRecursive15", `
			func fib(n) {
				if (n <= 1) { return n }
				return fib(n - 1) + fib(n - 2)
			}
			fib(15)
		`},
		{"FibIterative1000", `
			func fib(n) {
				var a = 0
				var b = 1
				for (var i = 0; i < n; i++) {
					var temp = a
					a = b
					b = temp + b
				}
				return a
			}
			fib(1000)
		`},
		{"ForLoop1000", `
			var total = 0
			for (var i = 0; i < 1000; i++) {
				total = total + i
			}
			total
		`},
		{"PrimeCheck100", `
			var count = 0
			for (var n = 2; n <= 100; n++) {
				var isPrime = true
				for (var i = 2; i * i <= n; i++) {
					if (n % i == 0) {
						isPrime = false
						break
					}
				}
				if (isPrime) {
					count = count + 1
				}
			}
			count
		`},
	}

	fmt.Println("\n========================================")
	fmt.Println("Xxlang Benchmark Results (for Python comparison)")
	fmt.Println("========================================")

	for _, bm := range benchmarks {
		l := lexer.New(bm.input)
		p := parser.New(l)
		program := p.ParseProgram()

		c := compiler.New()
		c.Compile(program)
		bytecode := c.Bytecode()

		// Run once to verify result
		v := vm.New(bytecode)
		v.Run()
		result := v.LastPopped()

		fmt.Printf("%-20s Result: %v\n", bm.name+":", result)
	}
}
