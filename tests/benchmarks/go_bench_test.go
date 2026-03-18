// tests/benchmarks/go_bench_test.go
// Go native performance benchmarks for comparison
package benchmarks

import (
	"testing"
)

// Fibonacci recursive
func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkGo_Fibonacci15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(15)
	}
}

func BenchmarkGo_Fibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

// Fibonacci iterative
func fibIter(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}

func BenchmarkGo_FibonacciIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibIter(30)
	}
}

// Factorial
func fact(n int) int {
	if n <= 1 {
		return 1
	}
	return n * fact(n-1)
}

func BenchmarkGo_Factorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fact(12)
	}
}

// For loop
func BenchmarkGo_ForLoop100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
			sum += j
		}
	}
}

func BenchmarkGo_ForLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += j
		}
	}
}

// Function calls
func add(a, b int) int {
	return a + b
}

func mul(a, b int) int {
	return a * b
}

func BenchmarkGo_FunctionCalls(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
			sum = add(sum, j)
			sum = mul(sum, 1)
		}
	}
}

// Arithmetic
func BenchmarkGo_Arithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 10
		b := 20
		c := a + b
		d := c * 2
		e := d - 5
		_ = e / 3
	}
}

// Intensive arithmetic
func BenchmarkGo_IntensiveArithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := 0
		for j := 0; j < 1000; j++ {
			x = x + j*2 - j/2
			x = x % 1000
		}
	}
}

// Comparisons
func BenchmarkGo_Comparisons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 10
		b := 20
		c := 0
		for j := 0; j < 100; j++ {
			if a < b {
				c++
			}
			if a > b {
				c++
			}
			if a == b {
				c++
			}
			if a != b {
				c++
			}
		}
	}
}

// While loop
func BenchmarkGo_WhileLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		i := 0
		sum := 0
		for i < 100 {
			sum += i
			i++
		}
	}
}

// Nested expressions
func BenchmarkGo_NestedExpressions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 1 + 2*3 - 4/2
		c := 10 // avoid division by zero
		d := (a + 5) * (c - a) / (5 + 1)
		_ = d
	}
}

// Prime check
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func BenchmarkGo_PrimeCheck100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			isPrime(j)
		}
	}
}

// Bubble sort
func bubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

func BenchmarkGo_BubbleSort10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := []int{5, 3, 8, 4, 2, 7, 1, 10, 6, 9}
		bubbleSort(arr)
	}
}
