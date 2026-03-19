// tests/benchmark_test.go
// Performance benchmarks for the xxlang interpreter
package tests

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// ============================================
// Benchmark Helpers
// ============================================

func compileAndRun(input string) (objects.Object, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		return nil, err
	}

	bytecode := c.Bytecode()
	v := vm.New(bytecode)

	if err := v.Run(); err != nil {
		return nil, err
	}

	return v.LastPopped(), nil
}

// ============================================
// Pipeline Benchmarks (Individual Components)
// ============================================

func BenchmarkLexer(b *testing.B) {
	input := `
		var x = 10
		var y = 20
		func add(a, b) {
			return a + b
		}
		print(add(x, y))
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer.New(input)
	}
}

func BenchmarkParser(b *testing.B) {
	input := `
		var x = 10
		var y = 20
		func add(a, b) {
			return a + b
		}
		print(add(x, y))
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := parser.New(l)
		p.ParseProgram()
	}
}

func BenchmarkCompiler(b *testing.B) {
	input := `
		var x = 10
		var y = 20
		func add(a, b) {
			return a + b
		}
		print(add(x, y))
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := compiler.New()
		c.Compile(program)
	}
}

func BenchmarkVM(b *testing.B) {
	input := `
		var x = 10
		var y = 20
		func add(a, b) {
			return a + b
		}
		print(add(x, y))
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFullPipeline(b *testing.B) {
	input := `
		var x = 10
		var y = 20
		func add(a, b) {
			return a + b
		}
		print(add(x, y))
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileAndRun(input)
	}
}

// ============================================
// Arithmetic Benchmarks
// ============================================

func BenchmarkIntegerArithmetic(b *testing.B) {
	input := "1 + 2 * 3 - 4 / 2"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFloatArithmetic(b *testing.B) {
	input := "1.5 + 2.5 * 3.0 - 4.0 / 2.0"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Fibonacci Benchmarks
// ============================================

func BenchmarkFibonacci10(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(10)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFibonacci20(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(20)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFibonacci25(b *testing.B) {
	input := `
		func fib(n) {
			if (n <= 1) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(25)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Factorial Benchmarks
// ============================================

func BenchmarkFactorial10(b *testing.B) {
	input := `
		func factorial(n) {
			if (n <= 1) {
				return 1
			}
			return n * factorial(n - 1)
		}
		factorial(10)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFactorial20(b *testing.B) {
	input := `
		func factorial(n) {
			if (n <= 1) {
				return 1
			}
			return n * factorial(n - 1)
		}
		factorial(20)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Loop Benchmarks
// ============================================

func BenchmarkForLoop100(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 100; i = i + 1) {
			sum = sum + i
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkForLoop1000(b *testing.B) {
	input := `
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + i
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkWhileLoop100(b *testing.B) {
	input := `
		var i = 0
		while (i < 100) {
			i = i + 1
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Function Call Benchmarks
// ============================================

func BenchmarkFunctionCalls100(b *testing.B) {
	input := `
		func identity(x) {
			return x
		}
		var result = 0
		for (var i = 0; i < 100; i = i + 1) {
			result = identity(i)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkClosureCalls100(b *testing.B) {
	input := `
		func makeAdder(x) {
			func adder(y) {
				return x + y
			}
			return adder
		}
		var add5 = makeAdder(5)
		var result = 0
		for (var i = 0; i < 100; i = i + 1) {
			result = add5(i)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Array Benchmarks
// ============================================

func BenchmarkArrayCreation100(b *testing.B) {
	// Create array with 100 elements
	input := `
		var arr = [
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
			21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
			31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
			41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
			51, 52, 53, 54, 55, 56, 57, 58, 59, 60,
			61, 62, 63, 64, 65, 66, 67, 68, 69, 70,
			71, 72, 73, 74, 75, 76, 77, 78, 79, 80,
			81, 82, 83, 84, 85, 86, 87, 88, 89, 90,
			91, 92, 93, 94, 95, 96, 97, 98, 99, 100
		]
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkArrayIndexing1000(b *testing.B) {
	input := `
		var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
		var sum = 0
		for (var i = 0; i < 1000; i = i + 1) {
			sum = sum + arr[i % 10]
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// String Benchmarks
// ============================================

func BenchmarkStringConcatenation100(b *testing.B) {
	input := `
		var result = ""
		for (var i = 0; i < 100; i = i + 1) {
			result = result + "x"
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkBuiltinLen1000(b *testing.B) {
	input := `
		var s = "hello world, this is a test string"
		var result = 0
		for (var i = 0; i < 1000; i = i + 1) {
			result = len(s)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Complex Algorithm Benchmarks
// ============================================

func BenchmarkBubbleSort10(b *testing.B) {
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
			return arr
		}
		bubbleSort([5, 3, 8, 4, 2, 7, 1, 10, 6, 9])
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkPrimeCheck100(b *testing.B) {
	input := `
		func isPrime(n) {
			if (n <= 1) { return false }
			if (n <= 3) { return true }
			if (n % 2 == 0) { return false }
			var i = 3
			while (i * i <= n) {
				if (n % i == 0) { return false }
				i = i + 2
			}
			return true
		}
		var count = 0
		for (var i = 2; i <= 100; i = i + 1) {
			if (isPrime(i)) { count = count + 1 }
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Tail Call Optimization Benchmarks
// ============================================

func BenchmarkFibonacciTailRecursive35(b *testing.B) {
	input := `
		func fibTail(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibTail(n - 1, b, a + b)
		}
		func fib(n) {
			return fibTail(n, 0, 1)
		}
		fib(35)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkFibonacciTailRecursive10000(b *testing.B) {
	input := `
		func fibTail(n, a, b) {
			if (n == 0) { return a }
			if (n == 1) { return b }
			return fibTail(n - 1, b, a + b)
		}
		func fib(n) {
			return fibTail(n, 0, 1)
		}
		fib(10000)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

// ============================================
// Inline Cache Benchmarks (Method Calls)
// ============================================

func BenchmarkStringMethodCalls1000(b *testing.B) {
	// Tests inline cache for string methods
	input := `
		var s = "hello world"
		var result = ""
		for (var i = 0; i < 1000; i++) {
			result = s.toUpper()
			result = s.toLower()
			result = s.contains("hello")
			result = s.indexOf("world")
			result = len(s)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkArrayMethodCalls1000(b *testing.B) {
	// Tests inline cache for array methods
	input := `
		var arr = [1, 2, 3, 4, 5]
		var result = 0
		for (var i = 0; i < 1000; i++) {
			push(arr, i)
			result = len(arr)
			result = arr.indexOf(3)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}

func BenchmarkMapMethodCalls1000(b *testing.B) {
	// Tests inline cache for map methods
	input := `
		var m = {"a": 1, "b": 2, "c": 3}
		var result = 0
		for (var i = 0; i < 1000; i++) {
			result = len(m)
			result = m.containsKey("a")
			result = m.keys().len()
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	c.Compile(program)
	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := vm.New(bytecode)
		v.Run()
	}
}
