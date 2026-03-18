// tests/reg_vm_benchmark_test.go
// Performance benchmarks comparing stack-based VM vs register-based VM
package tests

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// ============================================
// Helper Functions
// ============================================

// compileAndRunStack compiles with stack-based compiler and runs on stack VM
func compileAndRunStack(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		return err
	}

	bytecode := c.Bytecode()
	v := vm.New(bytecode)

	return v.Run()
}

// compileAndRunRegister compiles with register-based compiler and runs on register VM
func compileAndRunRegister(input string) error {
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

	bytecode := c.Bytecode()
	v := vm.NewRegVM(bytecode)

	return v.Run()
}

// ============================================
// Arithmetic Benchmarks
// ============================================

func BenchmarkStackVM_Arithmetic(b *testing.B) {
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
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_Arithmetic(b *testing.B) {
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
		compileAndRunRegister(input)
	}
}

// ============================================
// Loop Benchmarks
// ============================================

func BenchmarkStackVM_ForLoop100(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = sum + i
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_ForLoop100(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = sum + i
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

func BenchmarkStackVM_ForLoop1000(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + i
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_ForLoop1000(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + i
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

// ============================================
// Fibonacci Benchmarks
// ============================================

func BenchmarkStackVM_Fibonacci15(b *testing.B) {
	input := `
		func fib(n) {
			if (n < 2) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(15)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_Fibonacci15(b *testing.B) {
	input := `
		func fib(n) {
			if (n < 2) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(15)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

func BenchmarkStackVM_FibonacciIterative(b *testing.B) {
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
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_FibonacciIterative(b *testing.B) {
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
		compileAndRunRegister(input)
	}
}

// ============================================
// Function Call Benchmarks
// ============================================

func BenchmarkStackVM_FunctionCalls(b *testing.B) {
	input := `
		func add(a, b) {
			return a + b
		}
		func mul(a, b) {
			return a * b
		}
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = add(sum, i)
			sum = mul(sum, 1)
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_FunctionCalls(b *testing.B) {
	input := `
		func add(a, b) {
			return a + b
		}
		func mul(a, b) {
			return a * b
		}
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = add(sum, i)
			sum = mul(sum, 1)
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

// ============================================
// Comparison Benchmarks
// ============================================

func BenchmarkStackVM_Comparisons(b *testing.B) {
	input := `
		var a = 10
		var b = 20
		var c = 0
		for (var i = 0; i < 100; i = i + 1) {
			if (a < b) {
				c = c + 1
			}
			if (a > b) {
				c = c + 1
			}
			if (a == b) {
				c = c + 1
			}
			if (a != b) {
				c = c + 1
			}
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_Comparisons(b *testing.B) {
	input := `
		var a = 10
		var b = 20
		var c = 0
		for (var i = 0; i < 100; i = i + 1) {
			if (a < b) {
				c = c + 1
			}
			if (a > b) {
				c = c + 1
			}
			if (a == b) {
				c = c + 1
			}
			if (a != b) {
				c = c + 1
			}
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

// ============================================
// Nested Expression Benchmarks
// ============================================

func BenchmarkStackVM_NestedExpressions(b *testing.B) {
	input := `
		var a = 1 + 2 * 3 - 4 / 2
		var b = (1 + 2) * (3 - 4) / 2
		var c = a + b * 2 - a / b
		var d = (a + b) * (c - a) / (b + 1)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_NestedExpressions(b *testing.B) {
	input := `
		var a = 1 + 2 * 3 - 4 / 2
		var b = (1 + 2) * (3 - 4) / 2
		var c = a + b * 2 - a / b
		var d = (a + b) * (c - a) / (b + 1)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

// ============================================
// Factorial Benchmarks
// ============================================

func BenchmarkStackVM_Factorial(b *testing.B) {
	input := `
		func fact(n) {
			if (n <= 1) {
				return 1
			}
			return n * fact(n - 1)
		}
		fact(12)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_Factorial(b *testing.B) {
	input := `
		func fact(n) {
			if (n <= 1) {
				return 1
			}
			return n * fact(n - 1)
		}
		fact(12)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}

// ============================================
// While Loop Benchmarks
// ============================================

func BenchmarkStackVM_WhileLoop(b *testing.B) {
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
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_WhileLoop(b *testing.B) {
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
		compileAndRunRegister(input)
	}
}

// ============================================
// Intensive Arithmetic Benchmark
// ============================================

func BenchmarkStackVM_IntensiveArithmetic(b *testing.B) {
	input := `
		var x = 0
		for (var i = 0; i < 1000; i = i + 1) {
			x = x + i * 2 - i / 2
			x = x % 1000
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunStack(input)
	}
}

func BenchmarkRegisterVM_IntensiveArithmetic(b *testing.B) {
	input := `
		var x = 0
		for (var i = 0; i < 1000; i = i + 1) {
			x = x + i * 2 - i / 2
			x = x % 1000
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRunRegister(input)
	}
}
