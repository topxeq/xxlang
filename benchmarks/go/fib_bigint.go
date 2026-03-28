//go:build ignore

package main

import (
	"fmt"
	"math/big"
	"time"
)

// Iterative BigInt Fibonacci
func fibBigIntIter(n int) *big.Int {
	if n <= 1 {
		return big.NewInt(int64(n))
	}
	a := big.NewInt(0)
	b := big.NewInt(1)
	temp := big.NewInt(0)
	for i := 2; i <= n; i++ {
		temp.Set(a)
		temp.Add(temp, b)
		a.Set(b)
		b.Set(temp)
	}
	return b
}

// Recursive Fibonacci (int)
func fibInt(n int) int {
	if n <= 1 {
		return n
	}
	return fibInt(n-1) + fibInt(n-2)
}

// Iterative int Fibonacci
func fibIntIter(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	fmt.Println("=== Go Fibonacci Benchmark ===")
	fmt.Println()

	// Int recursive
	fmt.Println("--- Int (recursive) ---")
	start := time.Now()
	result := fibInt(35)
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("fib(35) = %d | time: %d ms\n", result, elapsed)

	start = time.Now()
	result = fibInt(40)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(40) = %d | time: %d ms\n", result, elapsed)

	// Int iterative
	fmt.Println("\n--- Int (iterative) ---")
	start = time.Now()
	result = fibIntIter(90)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(90) = %d | time: %d ms\n", result, elapsed)

	// BigInt iterative
	fmt.Println("\n--- BigInt (iterative) ---")
	start = time.Now()
	bigResult := fibBigIntIter(100)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(100) = %s | time: %d ms\n", bigResult.String(), elapsed)
	fmt.Printf("fib(100) bits: %d\n", bigResult.BitLen())

	start = time.Now()
	bigResult = fibBigIntIter(1000)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(1000) bits: %d | time: %d ms\n", bigResult.BitLen(), elapsed)

	start = time.Now()
	bigResult = fibBigIntIter(10000)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(10000) bits: %d | time: %d ms\n", bigResult.BitLen(), elapsed)

	start = time.Now()
	bigResult = fibBigIntIter(100000)
	elapsed = time.Since(start).Milliseconds()
	fmt.Printf("fib(100000) bits: %d | time: %d ms\n", bigResult.BitLen(), elapsed)

	// Show digits
	str := bigResult.String()
	fmt.Printf("fib(100000) length: %d digits\n", len(str))
	if len(str) > 50 {
		fmt.Printf("fib(100000) first 50: %s\n", str[:50])
		fmt.Printf("fib(100000) last 50: %s\n", str[len(str)-50:])
	}
}
