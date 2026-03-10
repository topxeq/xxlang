// benchmarks/go/fib_test.go
// Go baseline benchmarks for comparison
package main

import (
	"fmt"
	"testing"
	"time"
)

// Fibonacci - recursive (intentionally naive for fair comparison)
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkFib10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(10)
	}
}

func BenchmarkFib20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

func BenchmarkFib30(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(30)
	}
}

func BenchmarkFib35(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(35)
	}
}

// Iterative Fibonacci (optimized version)
func fibIter(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func BenchmarkFibIter35(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibIter(35)
	}
}

// Loop performance
func BenchmarkLoopSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 10000; j++ {
			sum += j
		}
	}
}

// Array operations
func BenchmarkArraySum(b *testing.B) {
	arr := make([]int, 1000)
	for i := range arr {
		arr[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range arr {
			sum += v
		}
	}
}

// Function call overhead
func add(a, b int) int {
	return a + b
}

func BenchmarkFunctionCalls(b *testing.B) {
	sum := 0
	for i := 0; i < b.N; i++ {
		sum = add(sum, i)
	}
}

// Main function to run single-threaded timing
func main() {
	fmt.Println("=== Go Performance Benchmarks ===")
	fmt.Println()

	// Run each benchmark once to get timing
	benchmarks := []struct {
		name string
		fn   func()
	}{
		{"fib(10)", func() { fib(10) }},
		{"fib(20)", func() { fib(20) }},
		{"fib(30)", func() { fib(30) }},
		{"fib(35)", func() { fib(35) }},
		{"fibIter(35)", func() { fibIter(35) }},
	}

	for _, bm := range benchmarks {
		start := time.Now()
		const iterations = 3
		for i := 0; i < iterations; i++ {
			bm.fn()
		}
		elapsed := time.Since(start) / iterations
		fmt.Printf("%-15s: %v\n", bm.name, elapsed)
	}
}
