// pkg/vm/comparison_bench_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// BenchmarkComparisonPrimeCheck compares prime checking performance
// with and without the OpRegPrimeCheck optimization
func BenchmarkComparisonPrimeCheck(b *testing.B) {
	// Code that uses the standard loop-based prime check
	standardCode := `
func isPrime(n) {
    if (n < 2) {
        return false
    }
    if (n == 2) {
        return true
    }
    if (n % 2 == 0) {
        return false
    }
    var i = 3
    while (i * i <= n) {
        if (n % i == 0) {
            return false
        }
        i = i + 1
    }
    return true
}

var count = 0
for (var n = 2; n < 5000; n++) {
    if (isPrime(n)) {
        count = count + 1
    }
}
count
`

	b.Run("StandardLoop", func(b *testing.B) {
		l := lexer.New(standardCode)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}

// BenchmarkComparisonNestedLoop compares nested loop performance
func BenchmarkComparisonNestedLoop(b *testing.B) {
	// Nested loop with accumulator
	code := `
var sum = 0
for (var i = 0; i < 100; i++) {
    for (var j = 0; j < 100; j++) {
        sum = sum + i * j
    }
}
sum
`

	b.Run("NestedLoop100x100", func(b *testing.B) {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}

// BenchmarkComparisonLoopUnrolling compares loop unrolling
func BenchmarkComparisonLoopUnrolling(b *testing.B) {
	// Small loop that should be unrolled
	smallLoop := `
var sum = 0
for (var i = 0; i < 8; i++) {
    sum = sum + i * i
}
sum
`

	b.Run("SmallLoop8", func(b *testing.B) {
		l := lexer.New(smallLoop)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})

	// Larger loop that won't be unrolled
	largeLoop := `
var sum = 0
for (var i = 0; i < 100; i++) {
    sum = sum + i * i
}
sum
`

	b.Run("LargeLoop100", func(b *testing.B) {
		l := lexer.New(largeLoop)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}

// BenchmarkComparisonCountingLoop compares counting loop optimizations
func BenchmarkComparisonCountingLoop(b *testing.B) {
	// Simple counting loop that should use OpRegLoopBodyAdd
	countingLoop := `
var total = 0
for (var i = 0; i < 1000; i++) {
    total = total + i
}
total
`

	b.Run("CountingLoop1000", func(b *testing.B) {
		l := lexer.New(countingLoop)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}

// BenchmarkComparisonFibonacci compares recursive function performance
func BenchmarkComparisonFibonacci(b *testing.B) {
	fibCode := `
func fib(n) {
    if (n < 2) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

fib(30)
`

	b.Run("Fibonacci30", func(b *testing.B) {
		l := lexer.New(fibCode)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}

// BenchmarkComparisonArraySum compares array summation
func BenchmarkComparisonArraySum(b *testing.B) {
	arraySumCode := `
var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
           21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
           41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60,
           61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80,
           81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100]

var sum = 0
for (var i = 0; i < len(arr); i++) {
    sum = sum + arr[i]
}
sum
`

	b.Run("ArraySum100", func(b *testing.B) {
		l := lexer.New(arraySumCode)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			b.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.NewRegCompiler()
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
	})
}
