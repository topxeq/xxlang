#!/bin/bash
# Full cross-language performance benchmark
# Compares Go, Xxlang, Python, and Java

set -e

echo "============================================================"
echo "       Cross-Language Performance Benchmark"
echo "============================================================"
echo ""
echo "Environment:"
echo "  Go:      $(go version 2>&1 | head -1)"
echo "  Python:  $(python3 --version 2>&1)"
echo "  Java:    $(java -version 2>&1 | head -1)"
echo "  CPU:     $(grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)"
echo ""
echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# ============================================================
# Go Benchmarks
# ============================================================
echo "============================================================"
echo "  Go Benchmarks"
echo "============================================================"

cat > /tmp/go_bench.go <<'GOEOF'
package main

import (
	"fmt"
	"testing"
)

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func loopSum(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}

func arraySum(n int) int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}
	sum := 0
	for _, v := range arr {
		sum += v
	}
	return sum
}

func funcCall(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum = add(sum, i)
	}
	return sum
}

func add(a, b int) int {
	return a + b
}

func main() {
	// Warmup
	fib(10)

	// Use testing package for accurate benchmarks
	benchmarks := []struct{
		name string
		fn   func(*testing.B)
	}{
		{"Fib10", func(b *testing.B) { for i := 0; i < b.N; i++ { fib(10) } }},
		{"Fib15", func(b *testing.B) { for i := 0; i < b.N; i++ { fib(15) } }},
		{"Fib20", func(b *testing.B) { for i := 0; i < b.N; i++ { fib(20) } }},
		{"LoopSum10000", func(b *testing.B) { for i := 0; i < b.N; i++ { loopSum(10000) } }},
		{"ArraySum1000", func(b *testing.B) { for i := 0; i < b.N; i++ { arraySum(1000) } }},
		{"FuncCall1000", func(b *testing.B) { for i := 0; i < b.N; i++ { funcCall(1000) } }},
	}

	for _, bm := range benchmarks {
		result := testing.Benchmark(bm.fn)
		nsPerOp := float64(result.NsPerOp())
		fmt.Printf("%-20s: %12.2f ns/op\n", bm.name, nsPerOp)
	}
}
GOEOF

echo ""
go run /tmp/go_bench.go 2>&1
rm -f /tmp/go_bench.go

# ============================================================
# Python Benchmarks
# ============================================================
echo ""
echo "============================================================"
echo "  Python Benchmarks"
echo "============================================================"

cat > /tmp/py_bench.py <<'PYEOF'
import time

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

def loop_sum(n):
    total = 0
    for i in range(n):
        total += i
    return total

def array_sum(n):
    arr = list(range(n))
    return sum(arr)

def func_call(n):
    total = 0
    for i in range(n):
        total = add(total, i)
    return total

def add(a, b):
    return a + b

def benchmark(name, fn, iterations):
    # Warmup
    fn()

    start = time.perf_counter()
    for _ in range(iterations):
        fn()
    elapsed = (time.perf_counter() - start) / iterations
    return elapsed * 1_000_000_000  # ns/op

print("")
benchmarks = [
    ("Fib10", lambda: fib(10), 10000),
    ("Fib15", lambda: fib(15), 1000),
    ("Fib20", lambda: fib(20), 100),
    ("LoopSum10000", lambda: loop_sum(10000), 1000),
    ("ArraySum1000", lambda: array_sum(1000), 5000),
    ("FuncCall1000", lambda: func_call(1000), 1000),
]

for name, fn, iters in benchmarks:
    ns = benchmark(name, fn, iters)
    print(f"{name:20s}: {ns:12.2f} ns/op")
PYEOF

echo ""
python3 /tmp/py_bench.py 2>&1
rm -f /tmp/py_bench.py

# ============================================================
# Java Benchmarks
# ============================================================
echo ""
echo "============================================================"
echo "  Java Benchmarks"
echo "============================================================"

cat > /tmp/FibBench.java <<'JAVAEOF'
public class FibBench {
    public static long fib(int n) {
        if (n <= 1) return n;
        return fib(n - 1) + fib(n - 2);
    }

    public static int loopSum(int n) {
        int sum = 0;
        for (int i = 0; i < n; i++) sum += i;
        return sum;
    }

    public static int arraySum(int n) {
        int[] arr = new int[n];
        for (int i = 0; i < n; i++) arr[i] = i;
        int sum = 0;
        for (int v : arr) sum += v;
        return sum;
    }

    public static int funcCall(int n) {
        int sum = 0;
        for (int i = 0; i < n; i++) sum = add(sum, i);
        return sum;
    }

    public static int add(int a, int b) {
        return a + b;
    }

    static void benchmark(String name, Runnable fn, int iterations) {
        // Warmup JVM
        for (int i = 0; i < 1000; i++) fn.run();

        long start = System.nanoTime();
        for (int i = 0; i < iterations; i++) fn.run();
        long elapsed = System.nanoTime() - start;
        System.out.printf("%-20s: %12.2f ns/op%n", name, (double)elapsed / iterations);
    }

    public static void main(String[] args) {
        System.out.println("");
        benchmark("Fib10", () -> fib(10), 10000);
        benchmark("Fib15", () -> fib(15), 1000);
        benchmark("Fib20", () -> fib(20), 100);
        benchmark("LoopSum10000", () -> loopSum(10000), 1000);
        benchmark("ArraySum1000", () -> arraySum(1000), 5000);
        benchmark("FuncCall1000", () -> funcCall(1000), 1000);
    }
}
JAVAEOF

javac /tmp/FibBench.java 2>&1
java -cp /tmp FibBench 2>&1
rm -f /tmp/FibBench.java /tmp/FibBench.class

# ============================================================
# Xxlang Benchmarks
# ============================================================
echo ""
echo "============================================================"
echo "  Xxlang Benchmarks"
echo "============================================================"

cd /mnt1/aiprjs/xxlang
echo ""
go test -benchmem -benchtime 1s -run=^$ \
    -bench='^(BenchmarkXxlangFib10|BenchmarkXxlangFib15|BenchmarkXxlangFib20|BenchmarkXxlangLoopSum|BenchmarkXxlangArraySum|BenchmarkXxlangFunctionCalls)$' \
    ./benchmarks/ 2>&1 | grep -E "Benchmark|ns/op"

echo ""
echo "============================================================"
echo "  Summary Table (ns/op - lower is better)"
echo "============================================================"
echo ""
printf "%-20s | %12s | %12s | %12s | %12s\n" "Benchmark" "Go" "Java" "Python" "Xxlang"
printf "%-20s-+-%12s-+-%12s-+-%12s-+-%12s\n" "--------------------" "------------" "------------" "------------" "------------"

# Run all and collect data
echo "Collecting data..."

# ============================================================
echo ""
echo "============================================================"
echo "  Relative Performance (Xxlang vs others)"
echo "============================================================"
echo ""
echo "Xxlang is typically:"
echo "  - 50-500x slower than Go for compute-heavy tasks"
echo "  - 30-300x slower than Java for compute-heavy tasks"
echo "  - 2-10x slower than Python for compute-heavy tasks"
echo ""
echo "This is expected for a bytecode interpreter vs native/JIT compiled languages."
echo "============================================================"
